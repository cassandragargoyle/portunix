package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "portunix prompt",
	Short: "Template-based prompt generation for AI assistants",
	Long: `Generate customized prompts for AI assistants from reusable templates.

The prompt command provides template-based prompt generation with automatic
placeholder detection and interactive parameter filling. Load templates with
placeholders, provide values via CLI or interactive mode, and get ready-to-use
prompts for Claude, ChatGPT, or other AI assistants.

Features:
• Load templates from .md, .yaml, .txt files
• Automatic placeholder detection ({placeholder} format)
• Interactive mode for missing parameters with smart defaults
• CLI arguments and interactive parameter input
• Output to stdout, clipboard, or file
• Multilingual template support

Examples:
  portunix prompt build template.md                    # Interactive mode
  portunix prompt build template.md --copy             # Build and copy to clipboard
  portunix prompt build template.md --var file=main.go # Provide parameters
  portunix prompt list                                 # List available templates
  portunix prompt create review.md                     # Create new template`,
	Version: "dev",
}

// SetVersion sets the version for the root command
func SetVersion(version string) {
	RootCmd.Version = version
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	RootCmd.AddCommand(buildCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(createCmd)
}