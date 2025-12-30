package main

import (
	"encoding/json"
	"fmt"
	"os"

	"portunix.ai/portunix/src/helpers/ptx-prompting/cmd"
)

var version = "dev"

// handleVersion outputs the version in the expected format
func handleVersion() {
	fmt.Printf("ptx-prompting version %s\n", version)
}

// handleListCommands outputs the commands this helper supports in JSON format
func handleListCommands() {
	commands := []string{"prompt"}
	output, err := json.Marshal(commands)
	if err != nil {
		fmt.Println(`["prompt"]`)
		return
	}
	fmt.Println(string(output))
}

// handleDescription outputs the description of this helper
func handleDescription() {
	fmt.Println("Template-based prompt generation for AI assistants")
}

// handleCommand processes dispatched commands from the main portunix binary
func handleCommand(args []string) {
	// Handle dispatched commands: prompt
	if len(args) == 0 {
		fmt.Println("No command specified")
		return
	}

	command := args[0]
	subArgs := args[1:]

	switch command {
	case "prompt":
		handlePromptCommand(subArgs)
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

// handlePromptCommand handles the 'prompt' command with its subcommands
func handlePromptCommand(args []string) {
	if len(args) == 0 {
		// Show prompt help
		fmt.Printf("Usage: portunix prompt [subcommand]\n")
		fmt.Println("\nAvailable subcommands:")
		fmt.Println("  build [template] [flags]  - Build prompt from template")
		fmt.Println("  list                      - List available templates")
		fmt.Println("  create [name]             - Create new template")
		fmt.Println("  --help                    - Show this help")
		return
	}

	// Simulate dispatched command by setting up fake cobra execution
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set up arguments as if they were called directly
	os.Args = []string{"ptx-prompting"}
	os.Args = append(os.Args, args...)

	// Execute the cobra command
	cmd.Execute()
}

func main() {
	// Set version
	cmd.SetVersion(version)

	// Handle special helper interface commands first
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
		}
	}

	// Handle dispatcher mode - when called with "prompt" as first argument
	if len(os.Args) > 1 && os.Args[1] == "prompt" {
		handlePromptCommand(os.Args[2:])
		return
	}

	// Default execution via Cobra
	cmd.Execute()
}