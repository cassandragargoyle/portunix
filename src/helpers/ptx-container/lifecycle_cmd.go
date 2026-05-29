/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// handleContainerLifecycle implements `portunix container lifecycle` with the
// subcommands list/inspect/extend/policy. The structure mirrors the existing
// network/volume dispatch in main.go.
func handleContainerLifecycle(args []string) {
	if len(args) == 0 {
		showLifecycleHelp()
		return
	}
	for _, a := range args {
		if a == "--help" || a == "-h" {
			showLifecycleHelp()
			return
		}
	}

	sub := args[0]
	rest := args[1:]
	switch sub {
	case "list", "ls":
		lifecycleList(rest)
	case "inspect":
		lifecycleInspect(rest)
	case "extend":
		lifecycleExtend(rest)
	case "policy":
		lifecyclePolicy(rest)
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown lifecycle subcommand: %s\n", sub)
		showLifecycleHelp()
		os.Exit(2)
	}
}

func lifecycleList(_ []string) {
	managed, err := ListManaged("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
	if len(managed) == 0 {
		fmt.Println("ℹ️  No portunix-managed containers found")
		return
	}
	now := time.Now()
	fmt.Printf("%-12s %-24s %-10s %-10s %-12s %s\n",
		"RUNTIME", "NAME", "POLICY", "STATUS", "TTL-LEFT", "EXPIRES")
	fmt.Println(strings.Repeat("-", 90))
	for _, mc := range managed {
		ttlLeft := "-"
		expires := "-"
		if !mc.Metadata.ExpiresAt.IsZero() {
			expires = mc.Metadata.ExpiresAt.Format(time.RFC3339)
			remaining := time.Until(mc.Metadata.ExpiresAt)
			if remaining < 0 {
				ttlLeft = "EXPIRED"
			} else {
				ttlLeft = remaining.Round(time.Second).String()
			}
		}
		policy := mc.Metadata.Policy
		if policy == "" {
			policy = "-"
		}
		_ = now
		fmt.Printf("%-12s %-24s %-10s %-10s %-12s %s\n",
			mc.Runtime, truncate(mc.Name, 24), policy,
			truncate(mc.Status, 10), ttlLeft, expires)
	}
}

func lifecycleInspect(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: portunix container lifecycle inspect <name>")
		os.Exit(2)
	}
	mc, err := InspectManaged(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Name:        %s\n", mc.Name)
	fmt.Printf("Runtime:     %s\n", mc.Runtime)
	fmt.Printf("ID:          %s\n", mc.ID)
	fmt.Printf("Image:       %s\n", mc.Image)
	fmt.Printf("Status:      %s\n", mc.Status)
	fmt.Printf("Running:     %t\n", mc.Running)
	fmt.Printf("Policy:      %s\n", mc.Metadata.Policy)
	if mc.Metadata.TTL > 0 {
		fmt.Printf("TTL:         %s\n", mc.Metadata.TTL)
	}
	if !mc.Metadata.CreatedAt.IsZero() {
		fmt.Printf("CreatedAt:   %s\n", mc.Metadata.CreatedAt.Format(time.RFC3339))
	}
	if !mc.Metadata.ExpiresAt.IsZero() {
		fmt.Printf("ExpiresAt:   %s\n", mc.Metadata.ExpiresAt.Format(time.RFC3339))
		fmt.Printf("TTL-Left:    %s\n", time.Until(mc.Metadata.ExpiresAt).Round(time.Second))
	}
	if mc.Metadata.HealthCheck != "" {
		fmt.Printf("HealthCheck: %s\n", mc.Metadata.HealthCheck)
	}
}

// lifecycleExtend re-tags the container with a new expires_at = now + duration.
// We rewrite both expires_at and ttl labels so subsequent listings report the
// new horizon. Docker does not support label edit, so we read the existing
// labels, drop the changed ones, and call `runtime container update`/`label`.
//
// Both Docker and Podman accept `update --label` for restart-policy and
// resource limits but neither supports re-labeling a running container in a
// portable way. We work around this by writing into a side-car file under
// ~/.portunix/container/lifecycle/<id>.labels and merging it on read; that is
// over-engineering for the common case. Instead we use Docker/Podman 25+'s
// `container update --label-add` when present, and fall back to recreating
// just the label via `commit + tag` is too disruptive — so for older runtimes
// we accept that extend is best-effort and falls back to printing a warning
// with the recommended manual command.
func lifecycleExtend(args []string) {
	var name string
	var byDur time.Duration
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--by":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "❌ --by requires a duration")
				os.Exit(2)
			}
			d, err := ParseDuration(args[i+1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ %v\n", err)
				os.Exit(2)
			}
			byDur = d
			i++
		default:
			if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(os.Stderr, "❌ Unknown flag: %s\n", args[i])
				os.Exit(2)
			}
			name = args[i]
		}
	}
	if name == "" || byDur <= 0 {
		fmt.Fprintln(os.Stderr, "Usage: portunix container lifecycle extend <name> --by <duration>")
		os.Exit(2)
	}

	mc, err := InspectManaged(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}

	base := mc.Metadata.ExpiresAt
	if base.IsZero() || time.Now().After(base) {
		base = time.Now()
	}
	newExpires := base.Add(byDur).UTC().Format(time.RFC3339)
	newTTL := byDur.String()

	// `container update --label` is the only portable way to mutate labels in
	// Docker 25+ / Podman 5+. On older runtimes this fails — we surface the
	// error and tell the user the safe alternative (recreate the container).
	cmd := exec.Command(mc.Runtime, "container", "update",
		"--label-add", LabelExpiresAt+"="+newExpires,
		"--label-add", LabelTTL+"="+newTTL,
		mc.ID,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Could not extend TTL via `%s container update`: %s\n",
			mc.Runtime, strings.TrimSpace(string(out)))
		fmt.Fprintln(os.Stderr, "💡 Older runtimes do not support label updates. Recreate the container with the new --ttl.")
		os.Exit(1)
	}
	fmt.Printf("✅ Extended TTL of %s by %s — expires at %s\n", mc.Name, byDur, newExpires)
}

// lifecyclePolicy switches the cleanup policy via the same `container update`
// path as lifecycleExtend. See the comment there for runtime-version caveats.
func lifecyclePolicy(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: portunix container lifecycle policy <name> <on-exit|ttl|manual>")
		os.Exit(2)
	}
	name := args[0]
	policy := args[1]
	if err := ValidatePolicy(policy); err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(2)
	}
	if policy == "" {
		fmt.Fprintln(os.Stderr, "❌ Policy must be one of: on-exit, ttl, manual")
		os.Exit(2)
	}

	mc, err := InspectManaged(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
	cmd := exec.Command(mc.Runtime, "container", "update",
		"--label-add", LabelPolicy+"="+policy,
		mc.ID,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Could not change policy via `%s container update`: %s\n",
			mc.Runtime, strings.TrimSpace(string(out)))
		os.Exit(1)
	}
	fmt.Printf("✅ Policy of %s set to %s\n", mc.Name, policy)
}

func showLifecycleHelp() {
	fmt.Println("Usage: portunix container lifecycle <subcommand>")
	fmt.Println()
	fmt.Println("⏳ MANAGE CONTAINER LIFECYCLE POLICIES")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  list                List all portunix-managed containers with TTL info")
	fmt.Println("  inspect <name>      Show full lifecycle metadata for a container")
	fmt.Println("  extend <name> --by <duration>")
	fmt.Println("                      Extend TTL of a container (e.g. --by 1h)")
	fmt.Println("  policy <name> <policy>")
	fmt.Println("                      Change cleanup policy (on-exit | ttl | manual)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container lifecycle list")
	fmt.Println("  portunix container lifecycle inspect cosmos-testnet")
	fmt.Println("  portunix container lifecycle extend cosmos-testnet --by 30m")
	fmt.Println("  portunix container lifecycle policy cosmos-testnet manual")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
