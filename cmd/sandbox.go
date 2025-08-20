package cmd

import (
	"github.com/spf13/cobra"
)

// sandboxCmd represents the sandbox command
var sandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Manages Windows Sandbox instances.",
	Long: `The sandbox command provides functionalities to manage Windows Sandbox instances,
including starting a sandbox and generating .wsb configuration files.`,
}

func init() {
	rootCmd.AddCommand(sandboxCmd)
}
