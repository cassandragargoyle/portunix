package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var wingetCmd = &cobra.Command{
	Use:   "winget",
	Short: "Windows Package Manager operations and information",
	Long: `Information about Windows Package Manager (winget).
Microsoft's official package manager for Windows.

Examples:
  portunix winget --info`,
	Run: func(cmd *cobra.Command, args []string) {
		showInfo, _ := cmd.Flags().GetBool("info")
		if showInfo {
			fmt.Println("📦 Windows Package Manager (winget) Information")
			fmt.Println("═══════════════════════════════════════════════")
			fmt.Println()
			fmt.Println("winget is Microsoft's official package manager written in C++.")
			fmt.Println("It's an open-source project available on GitHub.")
			fmt.Println()
			fmt.Println("Works similarly to apt on Linux or choco on Windows.")
			fmt.Println()
			fmt.Println("Features:")
			fmt.Println("• Search for applications")
			fmt.Println("• Download from verified sources")
			fmt.Println("• Install and update in silent mode")
			fmt.Println("• Manage multiple applications at once")
			fmt.Println()
			fmt.Println("Official repository: https://github.com/microsoft/winget-cli")
			fmt.Println()
			return
		}
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(wingetCmd)

	// Add flags
	wingetCmd.Flags().Bool("info", false, "Show information about Windows Package Manager")
}
