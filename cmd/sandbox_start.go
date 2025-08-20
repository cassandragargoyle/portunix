package cmd

import (
	"fmt"

	"portunix.cz/app/sandbox"

	"github.com/spf13/cobra"
)

// sandboxStartCmd represents the start command
var sandboxStartCmd = &cobra.Command{
	Use:   "start [wsb-file-path]",
	Short: "Starts a Windows Sandbox instance.",
	Long: `Starts a Windows Sandbox instance using a specified .wsb configuration file.

Example:
  portunix sandbox start C:\path\to\my_sandbox.wsb`,
	Args: cobra.ExactArgs(1), // Requires exactly one argument (the .wsb file path)
	Run: func(cmd *cobra.Command, args []string) {
		wsbFilePath := args[0]
		err := sandbox.StartSandbox(wsbFilePath)
		if err != nil {
			fmt.Printf("Error starting sandbox: %v\n", err)
		}
	},
}

func init() {
	sandboxCmd.AddCommand(sandboxStartCmd)
}
