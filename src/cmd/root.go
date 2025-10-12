package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"portunix.ai/app/version"
)

var (
	rootCmd = &cobra.Command{
		Use:   "portunix",
		Short: "Portunix is a universal tool for managing environments.",
		Long: `Portunix is a command-line interface (CLI) tool designed to simplify
the management of environments. It allows you to install software,
configure settings, create virtual machines, and more.

Use the --help flag with any command to see more information about it.`,
	}

	// Help level flags
	helpExpert bool
	helpAI     bool
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// SetVersion sets the version for the root command
func SetVersion() {
	rootCmd.Version = version.ProductVersion
}

func init() {
	// Here you will define your flags and configuration settings.

	// Disable the default help command to avoid duplication
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// Add help level flags
	rootCmd.PersistentFlags().BoolVar(&helpExpert, "help-expert", false, "Show extended help with all options and examples")
	rootCmd.PersistentFlags().BoolVar(&helpAI, "help-ai", false, "Show machine-readable help in JSON format")

	// Override help function to support multiple levels
	rootCmd.SetHelpFunc(multiLevelHelp)

	// Customize version template to match our version command output
	rootCmd.SetVersionTemplate("Portunix version {{.Version}}\n")
}

// multiLevelHelp implements the multi-level help system
func multiLevelHelp(cmd *cobra.Command, args []string) {
	// Check if this is the root command
	if cmd == rootCmd {
		// Check which help level is requested
		if helpAI {
			// Generate AI help in JSON format
			aiHelp, err := GenerateAIHelp()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating AI help: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(aiHelp)
		} else if helpExpert {
			// Generate expert help with all details
			fmt.Println(GenerateExpertHelp())
		} else {
			// Generate basic help
			fmt.Println(GenerateBasicHelp())
		}
	} else {
		// For subcommands, use default cobra help
		cmd.Usage()
	}
}
