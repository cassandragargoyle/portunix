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

var validateCmd = &cobra.Command{
	Use:   "validate [wizard-file]",
	Short: "Validate wizard YAML file",
	Long:  "Validate the syntax and structure of a wizard YAML file.",
	Args:  cobra.ExactArgs(1),
	Run:   validateWizard,
}

func validateWizard(_ *cobra.Command, args []string) {
	wizardPath := args[0]

	wizardEngine := engine.NewWizardEngine()
	wiz, err := wizardEngine.LoadWizard(wizardPath)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wizard '%s' is valid\n", wiz.Name)
	fmt.Printf("  ID: %s\n", wiz.ID)
	fmt.Printf("  Version: %s\n", wiz.Version)
	if wiz.Description != "" {
		fmt.Printf("  Description: %s\n", wiz.Description)
	}
	fmt.Printf("  Pages: %d\n", len(wiz.Pages))
}
