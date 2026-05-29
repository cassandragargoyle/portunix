/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"portunix.ai/portunix/src/helpers/ptx-wizard/cmd"
)

var version = "dev"

// handleVersion outputs the version in the format expected by the dispatcher
func handleVersion() {
	fmt.Printf("ptx-wizard version %s\n", version)
}

// handleListCommands outputs the commands this helper handles, in JSON
func handleListCommands() {
	commands := []string{"wizard"}
	output, err := json.Marshal(commands)
	if err != nil {
		fmt.Println(`["wizard"]`)
		return
	}
	fmt.Println(string(output))
}

// handleDescription outputs a one-line description of this helper
func handleDescription() {
	fmt.Println("Interactive CLI installation wizards (YAML-based)")
}

// runWizardSubcommand re-enters cobra to handle the wizard subcommand tree
func runWizardSubcommand(args []string) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"ptx-wizard"}
	os.Args = append(os.Args, args...)
	cmd.Execute()
}

func showHelpAI() {
	type CommandInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	type AIHelp struct {
		Tool        string        `json:"tool"`
		Version     string        `json:"version"`
		Description string        `json:"description"`
		Commands    []CommandInfo `json:"commands"`
	}
	help := AIHelp{
		Tool:        "ptx-wizard",
		Version:     version,
		Description: "Interactive CLI installation wizards (YAML-based)",
		Commands: []CommandInfo{
			{Name: "wizard run", Description: "Run a wizard by name or YAML file path"},
			{Name: "wizard list", Description: "List built-in and user wizards"},
			{Name: "wizard validate", Description: "Validate a wizard YAML definition"},
			{Name: "wizard create", Description: "Scaffold a new wizard YAML from template"},
		},
	}
	data, _ := json.MarshalIndent(help, "", "  ")
	fmt.Println(string(data))
}

func showHelpExpert() {
	fmt.Printf("PTX-WIZARD v%s - Interactive CLI Installation Wizards\n", version)
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Run YAML-defined interactive wizards for guided installation")
	fmt.Println("  and configuration. Supports conditional branching, themes,")
	fmt.Println("  validation, progress tracking, and non-interactive automation.")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  wizard run <name|file> [flags]   Run a wizard")
	fmt.Println("    --theme <name>                    UI theme: default | colorful | minimal")
	fmt.Println("    --non-interactive                 Disable prompts; require --config")
	fmt.Println("    --config <file>                   YAML/JSON variable values for non-interactive mode")
	fmt.Println("    --preset <name>                   Named preset (development, production)")
	fmt.Println("  wizard list                       List built-in and user wizards")
	fmt.Println("  wizard validate <file>            Validate a wizard YAML definition")
	fmt.Println("  wizard create <name>              Create new wizard YAML from template")
	fmt.Println()
	fmt.Println("WIZARD SEARCH PATHS:")
	fmt.Println("  ./examples/wizards/      Project-bundled wizards")
	fmt.Println("  ~/.portunix/wizards/     User wizards")
	fmt.Println("  $PORTUNIX_WIZARD_PATH    Extra path(s), OS-separator delimited")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  portunix wizard list")
	fmt.Println("  portunix wizard run database-setup")
	fmt.Println("  portunix wizard run ./my-wizard.yaml --theme colorful")
	fmt.Println("  portunix wizard run database-setup --non-interactive --config db.yaml")
}

func main() {
	cmd.SetVersion(version)

	// Helper interface (single-flag forms used by the dispatcher)
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "--version":
			handleVersion()
			return
		case "--list-commands":
			handleListCommands()
			return
		case "--description":
			handleDescription()
			return
		case "--help-ai":
			showHelpAI()
			return
		case "--help-expert":
			showHelpExpert()
			return
		}
	}

	// Dispatcher mode — when called with "wizard" as first argument
	if len(os.Args) > 1 && os.Args[1] == "wizard" {
		runWizardSubcommand(os.Args[2:])
		return
	}

	cmd.Execute()
}
