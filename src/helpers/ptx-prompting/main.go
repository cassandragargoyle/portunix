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

// handleCommand dispatches the "prompt" command routed to this helper by the
// parent portunix binary (see src/dispatcher/dispatcher.go). args arrive
// without the binary name prefix; routing is delegated to handlePromptCommand
// which re-enters the helper's own cobra tree.
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
		Tool:        "ptx-prompting",
		Version:     version,
		Description: "Template-based prompt generation for AI assistants",
		Commands: []CommandInfo{
			{Name: "prompt build", Description: "Build prompt from template with variable substitution"},
			{Name: "prompt list", Description: "List available prompt templates"},
			{Name: "prompt create", Description: "Create new prompt template"},
		},
	}
	data, _ := json.MarshalIndent(help, "", "  ")
	fmt.Println(string(data))
}

func showHelpExpert() {
	fmt.Printf("PTX-PROMPTING v%s - Template-Based Prompt Generation\n", version)
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Generate structured prompts for AI assistants using templates")
	fmt.Println("  with variable substitution, role definitions, and context injection.")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  prompt build [template] [flags]   Build prompt from template")
	fmt.Println("    --var <key=value>                  Set template variable")
	fmt.Println("    --output <file>                    Write output to file")
	fmt.Println("    --format <text|json|markdown>      Output format")
	fmt.Println("  prompt list                        List available templates")
	fmt.Println("    --format <text|json>               Output format")
	fmt.Println("  prompt create [name]               Create new template")
	fmt.Println()
	fmt.Println("TEMPLATE LOCATIONS:")
	fmt.Println("  ./prompts/              Project-local templates")
	fmt.Println("  ~/.portunix/prompts/    User-global templates")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  portunix prompt list")
	fmt.Println("  portunix prompt build code-review --var file=main.go")
	fmt.Println("  portunix prompt create my-template")
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
		case "--help-ai":
			showHelpAI()
			return
		case "--help-expert":
			showHelpExpert()
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
