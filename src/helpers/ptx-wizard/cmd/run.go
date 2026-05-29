/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"portunix.ai/app/wizard/engine"
)

var (
	runTheme          string
	runNonInteractive bool
	runConfig         string
	runPreset         string
)

var runCmd = &cobra.Command{
	Use:   "run [wizard-name-or-file]",
	Short: "Run a wizard",
	Long: `Run an installation wizard by name or from a YAML file.

Examples:
  portunix wizard run database-setup
  portunix wizard run ./my-wizard.yaml
  portunix wizard run database-setup --non-interactive --config db.yaml`,
	Args: cobra.ExactArgs(1),
	Run:  runWizard,
}

func init() {
	runCmd.Flags().StringVar(&runTheme, "theme", "default", "UI theme (default, colorful, minimal)")
	runCmd.Flags().BoolVar(&runNonInteractive, "non-interactive", false, "Run in non-interactive mode (requires --config)")
	runCmd.Flags().StringVar(&runConfig, "config", "", "YAML/JSON file with variable values for non-interactive mode")
	runCmd.Flags().StringVar(&runPreset, "preset", "", "Apply named preset (e.g. development, production)")
}

func runWizard(_ *cobra.Command, args []string) {
	wizardInput := args[0]

	wizardEngine := engine.NewWizardEngine()

	if runTheme != "default" {
		if err := wizardEngine.SetTheme(runTheme); err != nil {
			fmt.Printf("Warning: %v, using default theme\n", err)
		}
	}

	wizardPath, err := resolveWizardPath(wizardInput)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	wiz, err := wizardEngine.LoadWizard(wizardPath)
	if err != nil {
		fmt.Printf("Error loading wizard: %v\n", err)
		os.Exit(1)
	}

	if runNonInteractive {
		if runConfig == "" {
			fmt.Println("Error: --config is required for --non-interactive")
			os.Exit(1)
		}
		if err := wizardEngine.LoadVariablesFromConfig(runConfig); err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
		wizardEngine.SetNonInteractive(true)
	}

	if runPreset != "" {
		if err := loadPreset(wizardEngine, wizardPath, runPreset); err != nil {
			fmt.Printf("Error loading preset '%s': %v\n", runPreset, err)
			os.Exit(1)
		}
	}

	fmt.Printf("Starting wizard: %s v%s\n", wiz.Name, wiz.Version)
	if wiz.Description != "" {
		fmt.Printf("Description: %s\n\n", wiz.Description)
	}

	result, err := wizardEngine.ExecuteWizard(wiz)
	if err != nil {
		fmt.Printf("Wizard failed: %v\n", err)
		os.Exit(1)
	}

	if result.Completed {
		fmt.Printf("\nWizard completed successfully in %v\n", result.Duration)
	} else {
		fmt.Printf("\nWizard was interrupted\n")
		os.Exit(1)
	}
}
