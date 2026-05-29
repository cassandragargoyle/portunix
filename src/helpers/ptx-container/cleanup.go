/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CleanupFilter expresses what `container cleanup` should match. Zero values
// mean "no filter for this dimension"; an all-zero filter without --all is
// treated as "expired containers only" so that scheduled service runs and
// accidental command invocations both default to the safest behaviour.
type CleanupFilter struct {
	OlderThan      time.Duration // matches managed containers whose CreatedAt is older
	Pattern        string        // glob matched against container name (filepath.Match syntax)
	Status         string        // "running", "exited", or "" for any
	ExcludeRunning bool
	ExpiredOnly    bool // default mode — only TTL-expired containers
	All            bool // override: every managed container matches
}

// CleanupResult is what the user sees and what callers (service loop, tests)
// inspect to count actions. Removed/Failed/Skipped are all populated even on
// dry-run so the report is complete.
type CleanupResult struct {
	Examined int
	Removed  []string // container "<runtime>:<name>" identifiers
	Skipped  []string
	Failed   []string
	DryRun   bool
}

// CleanupOptions bundles the side-effect-controlling flags so the function
// signature stays small.
type CleanupOptions struct {
	DryRun bool
	Force  bool // pass through to runtime rm -f (kills running containers)
}

// CleanupContainers executes the cleanup pass and returns a structured result.
// The function never panics on partial failures: a runtime rm error pushes the
// container into Failed and the loop continues, so service runs do not get
// stuck on a single misbehaving container.
func CleanupContainers(now time.Time, filter CleanupFilter, opts CleanupOptions) (CleanupResult, error) {
	managed, err := ListManaged("")
	if err != nil {
		return CleanupResult{}, err
	}

	result := CleanupResult{DryRun: opts.DryRun, Examined: len(managed)}

	for _, mc := range managed {
		if !matchesCleanupFilter(mc, filter, now) {
			result.Skipped = append(result.Skipped, identify(mc))
			continue
		}
		if opts.DryRun {
			result.Removed = append(result.Removed, identify(mc))
			continue
		}
		if err := removeManaged(mc, opts.Force); err != nil {
			result.Failed = append(result.Failed, fmt.Sprintf("%s (%v)", identify(mc), err))
			continue
		}
		result.Removed = append(result.Removed, identify(mc))
	}
	return result, nil
}

func identify(mc ManagedContainer) string {
	return fmt.Sprintf("%s:%s", mc.Runtime, mc.Name)
}

// matchesCleanupFilter applies the user-provided filter to a single container.
// Returns true when the container should be removed.
func matchesCleanupFilter(mc ManagedContainer, f CleanupFilter, now time.Time) bool {
	if f.ExcludeRunning && mc.Running {
		return false
	}
	if f.Status != "" && !strings.EqualFold(mc.Status, f.Status) {
		return false
	}
	if f.Pattern != "" {
		ok, _ := filepath.Match(f.Pattern, mc.Name)
		if !ok {
			return false
		}
	}
	if f.OlderThan > 0 {
		if mc.Metadata.CreatedAt.IsZero() ||
			now.Sub(mc.Metadata.CreatedAt) < f.OlderThan {
			return false
		}
	}

	// A filter that explicitly asks for --all matches everything that survived
	// the negative filters above. Otherwise default to expired-only — never
	// remove a healthy managed container by accident.
	if f.All {
		return true
	}
	if f.ExpiredOnly || (f.OlderThan == 0 && f.Pattern == "" && f.Status == "") {
		return mc.Metadata.IsExpired(now)
	}
	// Positive filters were provided (older-than/pattern/status). Match any
	// container that survived the negative filters.
	return true
}

func removeManaged(mc ManagedContainer, force bool) error {
	args := []string{"rm"}
	if force || mc.Running {
		args = append(args, "-f")
	}
	args = append(args, mc.ID)
	cmd := exec.Command(mc.Runtime, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s rm: %s", mc.Runtime, strings.TrimSpace(string(out)))
	}
	return nil
}

// handleContainerCleanup is the ptx-container subcommand entry point for
// `portunix container cleanup`. It parses flags, runs the cleanup pass, and
// renders the human-friendly report; the underlying CleanupContainers is the
// reusable engine.
func handleContainerCleanup(args []string) {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			showCleanupHelp()
			return
		}
	}

	filter := CleanupFilter{}
	opts := CleanupOptions{}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--dry-run":
			opts.DryRun = true
		case "--force", "-f":
			opts.Force = true
		case "--all":
			filter.All = true
		case "--expired":
			filter.ExpiredOnly = true
		case "--exclude-running":
			filter.ExcludeRunning = true
		case "--older-than":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "❌ --older-than requires a duration")
				os.Exit(2)
			}
			d, err := ParseDuration(args[i+1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ %v\n", err)
				os.Exit(2)
			}
			filter.OlderThan = d
			i++
		case "--pattern":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "❌ --pattern requires a glob")
				os.Exit(2)
			}
			filter.Pattern = args[i+1]
			i++
		case "--status":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "❌ --status requires a value")
				os.Exit(2)
			}
			filter.Status = args[i+1]
			i++
		default:
			fmt.Fprintf(os.Stderr, "❌ Unknown flag: %s\n", args[i])
			showCleanupHelp()
			os.Exit(2)
		}
		i++
	}

	res, err := CleanupContainers(time.Now(), filter, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Cleanup failed: %v\n", err)
		os.Exit(1)
	}

	printCleanupReport(res)
	if len(res.Failed) > 0 {
		os.Exit(1)
	}
}

func printCleanupReport(res CleanupResult) {
	mode := "removed"
	if res.DryRun {
		mode = "would remove"
	}
	fmt.Printf("📊 Cleanup report: %d container(s) examined\n", res.Examined)
	if len(res.Removed) > 0 {
		fmt.Printf("✅ %s %d:\n", mode, len(res.Removed))
		for _, name := range res.Removed {
			fmt.Printf("   - %s\n", name)
		}
	}
	if len(res.Skipped) > 0 && res.DryRun {
		// Only show skipped on dry-run; on real runs the noise isn't useful.
		fmt.Printf("⏭️  skipped %d (filtered out)\n", len(res.Skipped))
	}
	if len(res.Failed) > 0 {
		fmt.Printf("❌ failed %d:\n", len(res.Failed))
		for _, name := range res.Failed {
			fmt.Printf("   - %s\n", name)
		}
	}
	if len(res.Removed) == 0 && len(res.Failed) == 0 {
		fmt.Println("ℹ️  Nothing to clean up")
	}
}

func showCleanupHelp() {
	fmt.Println("Usage: portunix container cleanup [flags]")
	fmt.Println()
	fmt.Println("🧹 CLEANUP MANAGED CONTAINERS")
	fmt.Println()
	fmt.Println("Removes Portunix-managed containers (label portunix.managed=true) that")
	fmt.Println("match the supplied filter. With no filter, defaults to TTL-expired only.")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --all                 Remove every managed container")
	fmt.Println("  --expired             Remove only TTL-expired containers (default)")
	fmt.Println("  --older-than DURATION Match containers older than e.g. 1h, 7d, 30m")
	fmt.Println("  --pattern GLOB        Match container name against a glob (e.g. 'test-*')")
	fmt.Println("  --status STATE        Match container status (running, exited, ...)")
	fmt.Println("  --exclude-running     Skip currently running containers")
	fmt.Println("  --dry-run             Show what would be removed without acting")
	fmt.Println("  -f, --force           Force-remove running containers (rm -f)")
	fmt.Println("  -h, --help            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container cleanup --dry-run")
	fmt.Println("  portunix container cleanup --older-than 7d --pattern 'dev-*'")
	fmt.Println("  portunix container cleanup --all --force")
}
