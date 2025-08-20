package cmd

import (
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new resource.",
	Long: `The create command is used to create new resources, such as virtual machines.

Use one of the available subcommands to create a specific resource.

Example:
  portunix create vm`,
}

func init() {
	rootCmd.AddCommand(createCmd)
}
