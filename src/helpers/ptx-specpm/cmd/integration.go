/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"portunix.ai/portunix/src/helpers/ptx-specpm/internal/integration"
)

var integrationCmd = &cobra.Command{
	Use:   "integration",
	Short: "Inspect and manage AI-agent integrations",
}

var integrationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List supported AI-agent integration targets",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Supported integrations:")
		for _, d := range integration.All() {
			installed, ver := d.Detect()
			status := "supported"
			if installed {
				if ver != "" {
					status = "supported (installed: " + ver + ")"
				} else {
					status = "supported (installed)"
				}
			}
			fmt.Printf("  %-10s  %s  — %s\n", d.Name(), status, d.Description())
		}
		fmt.Println()
		fmt.Println("Phase 1 ships only the 'claude' driver end-to-end.")
		fmt.Println("Additional drivers (copilot, cursor, gemini, ...) arrive in Phase 4.")
		return nil
	},
}

func init() {
	integrationCmd.AddCommand(integrationListCmd)
}
