/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"portunix.ai/portunix/src/helpers/ptx-specpm/internal/integration"
	"portunix.ai/portunix/src/helpers/ptx-specpm/internal/kit"
	"portunix.ai/portunix/src/helpers/ptx-specpm/internal/kitjson"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Show installed agents, wired integration, profile, and kit ref for the current project",
	RunE:  runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	manifestPath := filepath.Join(cwd, ".specpm", "kit.json")

	fmt.Printf("%s check  (helper version: %s)\n", displayName, version)
	fmt.Println("====================================")

	manifest, manifestErr := kitjson.Read(manifestPath)
	if manifestErr != nil {
		fmt.Printf("Project          : not initialised in this directory (%s)\n", manifestPath)
	} else {
		fmt.Printf("Project          : initialised\n")
		fmt.Printf("Kit source       : %s\n", manifest.Source)
		fmt.Printf("Kit ref          : %s\n", manifest.Ref)
		if manifest.Profile != "" {
			fmt.Printf("Active profile   : %s\n", manifest.Profile)
		}
		if manifest.Integration != "" {
			mode := "slash-commands"
			if manifest.Skills {
				mode = "skills"
			}
			fmt.Printf("Integration      : %s (%s)\n", manifest.Integration, mode)
		}
		if manifest.Ref != "" && manifest.Ref != kit.DefaultPinnedRef && manifest.Source == "git" {
			// ADR-041 D2: drift line is informational; exit code unchanged.
			// Text is part of the locked surface — do not paraphrase.
			fmt.Printf("Drift            : binary recommends %s (this Portunix release)\n", kit.DefaultPinnedRef)
			fmt.Printf("                   Run 'portunix specpm upgrade --to-default' to follow the recommendation.\n")
		}
	}

	fmt.Println()
	fmt.Printf("Default kit ref  : %s\n", kit.DefaultPinnedRef)
	fmt.Printf("Cache root       : %s\n", kit.CacheRoot())

	fmt.Println()
	fmt.Println("AI-agent CLI detection:")
	for _, drv := range integration.All() {
		installed, ver := drv.Detect()
		mark := "[ ]"
		ann := "not detected"
		if installed {
			mark = "[x]"
			ann = ver
			if ann == "" {
				ann = "installed"
			}
		}
		fmt.Printf("  %s %-10s  %s\n", mark, drv.Name(), ann)
	}

	return nil
}
