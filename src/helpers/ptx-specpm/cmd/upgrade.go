/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"portunix.ai/portunix/src/helpers/ptx-specpm/internal/kit"
	"portunix.ai/portunix/src/helpers/ptx-specpm/internal/kitjson"
	"portunix.ai/portunix/src/helpers/ptx-specpm/internal/scaffold"
)

var (
	upgradeFlagRef       string
	upgradeFlagToDefault bool
	upgradeFlagSource    string
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Refresh .specpm/ kit cache; never touches specs/project/",
	Long: `upgrade refreshes the .specpm/ kit cache for the project in the current
directory. The project's currently-pinned ref (recorded in .specpm/kit.json) is
the source of truth; pin movement is always an explicit --ref or --to-default.

Forms:
  ptx-specpm upgrade                Refresh from project's pinned ref (idempotent on cache hit)
  ptx-specpm upgrade --ref <X>      Bump project pin to <X>; fetch and refresh
  ptx-specpm upgrade --to-default   Shortcut for --ref <DefaultPinnedRef>
  ptx-specpm upgrade --source <p>   Switch to local-path provider; refresh from <p>

Carve-outs preserved on every upgrade (per ADR-041 D3):
  .specpm/kit.json        manifest (rewritten in place)
  .specpm/extensions/     user-installed extensions (Phase 3+)
  .specpm/presets/        user-installed presets (Phase 3+)

Files under specs/project/<id>/ are NEVER touched.`,
	Args: cobra.NoArgs,
	RunE: runUpgrade,
}

func init() {
	f := upgradeCmd.Flags()
	f.StringVar(&upgradeFlagRef, "ref", "", "Bump project pin to this git ref (tag or commit SHA)")
	f.BoolVar(&upgradeFlagToDefault, "to-default", false, "Shortcut for --ref <DefaultPinnedRef> (track this Portunix release)")
	f.StringVar(&upgradeFlagSource, "source", "", "Switch to local-path provider (absolute path); incompatible with --ref")
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	manifestPath := filepath.Join(cwd, ".specpm", "kit.json")

	existing, err := kitjson.Read(manifestPath)
	if err != nil {
		return fmt.Errorf("no specpm project in %s (%s missing); run 'init' first", cwd, manifestPath)
	}

	provider, targetSource, targetRef, err := resolveUpgradeProvider(existing)
	if err != nil {
		return err
	}

	kitDir, fromCache, err := provider.Fetch()
	if err != nil {
		return fmt.Errorf("fetch kit: %w", err)
	}

	scaffoldResult, err := scaffold.Write(scaffold.Options{
		TargetDir: cwd,
		KitDir:    kitDir,
		Refresh:   true,
	})
	if err != nil {
		return fmt.Errorf("refresh scaffold: %w", err)
	}

	// Preserve profile/integration/skills across upgrade — they reflect the
	// project's wired state, not the kit version. Only source/ref/written_at
	// move. WrittenAt is reset so kitjson.Write stamps it.
	updated := existing
	updated.Source = targetSource
	updated.Ref = targetRef
	updated.WrittenAt = time.Time{}
	if err := kitjson.Write(manifestPath, updated); err != nil {
		return fmt.Errorf("write kit.json: %w", err)
	}

	printUpgradeSummary(cwd, existing, updated, fromCache, scaffoldResult)
	return nil
}

// resolveUpgradeProvider applies the --ref / --to-default / --source flags
// against the existing manifest and returns the provider, the target source
// label (as written into kit.json) and the target ref (as written into
// kit.json). The mutual-exclusion rules are:
//   - --ref and --source <non-git-path> are incompatible (--ref is a git ref).
//   - --to-default implies git mode and overrides the existing source.
//   - --source <path> sets ref="(local path)".
//   - --source git with no --ref keeps the existing ref.
//   - Plain upgrade keeps both source and ref from the manifest.
func resolveUpgradeProvider(existing kitjson.Manifest) (kit.Provider, string, string, error) {
	targetSource := existing.Source
	targetRef := existing.Ref

	if upgradeFlagToDefault && upgradeFlagRef != "" {
		return nil, "", "", errors.New("--to-default and --ref are mutually exclusive")
	}
	if upgradeFlagToDefault {
		targetSource = "git"
		targetRef = kit.DefaultPinnedRef
	}
	if upgradeFlagRef != "" {
		// --ref implies git mode unless --source git is explicitly co-passed.
		// --ref + --source <path> is rejected — see below.
		targetSource = "git"
		targetRef = upgradeFlagRef
	}
	if upgradeFlagSource != "" {
		if upgradeFlagSource == "git" {
			targetSource = "git"
			if targetRef == "" || targetRef == "(local path)" {
				targetRef = kit.DefaultPinnedRef
			}
		} else {
			if upgradeFlagRef != "" {
				return nil, "", "", errors.New("--ref is incompatible with --source <path>; --ref selects a git ref")
			}
			abs, err := filepath.Abs(upgradeFlagSource)
			if err != nil {
				return nil, "", "", fmt.Errorf("resolve --source path: %w", err)
			}
			targetSource = abs
			targetRef = "(local path)"
		}
	}

	if targetSource == "" {
		// Defensive: an init-with-empty-source would have written
		// source="git" already, but cope with hand-edited manifests.
		targetSource = "git"
		if targetRef == "" {
			targetRef = kit.DefaultPinnedRef
		}
	}

	if targetSource == "git" {
		if targetRef == "" || targetRef == "(local path)" {
			targetRef = kit.DefaultPinnedRef
		}
		return kit.NewGitProvider(kit.UpstreamURL, targetRef), targetSource, targetRef, nil
	}
	// targetSource is an absolute filesystem path.
	return kit.NewLocalPathProvider(targetSource), targetSource, "(local path)", nil
}

func printUpgradeSummary(target string, before, after kitjson.Manifest, fromCache bool, sc *scaffold.Result) {
	fmt.Printf("%s upgrade complete\n", displayName)
	fmt.Println("==============================")
	fmt.Printf("Target              : %s\n", target)
	fmt.Printf("Kit source          : %s\n", after.Source)
	if before.Ref != after.Ref {
		fmt.Printf("Kit ref             : %s -> %s\n", refLabel(before.Ref), refLabel(after.Ref))
	} else {
		fmt.Printf("Kit ref             : %s (unchanged)\n", refLabel(after.Ref))
	}
	if after.Source == "git" {
		if fromCache {
			fmt.Println("Cache               : hit (offline)")
		} else {
			fmt.Println("Cache               : miss (cloned upstream, populated cache)")
		}
	}
	if sc != nil && len(sc.WipedSubtrees) > 0 {
		fmt.Printf("Refreshed subtrees  : %v\n", sc.WipedSubtrees)
	}
	fmt.Println("Carve-outs          : kit.json / .specpm/extensions/ / .specpm/presets/ preserved")
	fmt.Println("specs/project/      : not touched")
	fmt.Println()
	if before.Ref != after.Ref {
		fmt.Println("Pin moved. Run 'portunix specpm check' to confirm the new state.")
	} else {
		fmt.Println("No pin movement; .specpm/ refreshed from the active source.")
	}
}

func refLabel(r string) string {
	if r == "" {
		return "(unset)"
	}
	return r
}
