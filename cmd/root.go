package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"portunix.cz/app/version"
)

var rootCmd = &cobra.Command{
	Use:     "portunix",
	Version: version.ProductVersion,
	Short:   "Portunix is a universal tool for managing development environments.",
	Long: `Portunix is a command-line interface (CLI) tool designed to simplify
the management of development environments. It allows you to install software,
configure settings, create virtual machines, and more.

Use the --help flag with any command to see more information about it.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	
	// Customize version template to match our version command output
	rootCmd.SetVersionTemplate("Portunix version {{.Version}}\n")
}
