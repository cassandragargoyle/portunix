package cmd

import (
	"fmt"

	"portunix.ai/app/sandbox"

	"github.com/spf13/cobra"
)

// sandboxDefaultCmd represents the default command
var sandboxDefaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Generates a default .wsb file and starts the sandbox.",
	Long: `Generates a .wsb configuration file with default settings (networking and clipboard enabled,
notepad.exe as logon command) and then starts the Windows Sandbox instance.

This command simplifies the process of launching a basic sandbox without manual .wsb file creation.

Example:
  portunix sandbox default`,
	Run: func(cmd *cobra.Command, args []string) {
		config := sandbox.SandboxConfig{
			EnableNetworking: true,
			EnableClipboard:  true,
			LogonCommand:     "notepad.exe",
		}

		wsbFilePath, _, err := sandbox.GenerateWsbFile(config)
		if err != nil {
			fmt.Printf("Error generating default .wsb file: %v\n", err)
			return
		}

		fmt.Printf("Generated default .wsb file: %s\n", wsbFilePath)

		err = sandbox.StartSandbox(wsbFilePath)
		if err != nil {
			fmt.Printf("Error starting default sandbox: %v\n", err)
		}

		// Optionally, remove the generated .wsb file after use
		// defer os.Remove(wsbFilePath)
	},
}

func init() {
	sandboxCmd.AddCommand(sandboxDefaultCmd)
}
