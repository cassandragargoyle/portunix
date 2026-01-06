package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
// NOTE: This command is only reached as a fallback when ptx-installer helper is not available.
// When ptx-installer is available, the dispatcher in main.go will route to it directly.
var installCmd = &cobra.Command{
	Use:   "install [package] [flags]",
	Short: "Installs specified software packages.",
	Long: `The install command allows you to install various software components.

This command requires the ptx-installer helper binary.

Examples:
  portunix install chrome
  portunix install java --variant 17
  portunix install python --embeddable
  portunix install docker
  portunix install hugo --dry-run

Use 'portunix package list' to see available packages.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		// This code is only reached when ptx-installer is not available
		// (dispatcher would have routed to ptx-installer if it existed)

		// Find expected helper location
		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)

		helperName := "ptx-installer"
		if runtime.GOOS == "windows" {
			helperName = "ptx-installer.exe"
		}

		expectedPath := filepath.Join(execDir, helperName)

		fmt.Println("❌ Error: ptx-installer helper not found")
		fmt.Println()
		fmt.Printf("Expected location: %s\n", expectedPath)
		fmt.Println()
		fmt.Println("The install command requires the ptx-installer helper binary.")
		fmt.Println("Please ensure ptx-installer is in the same directory as portunix.")
		fmt.Println()
		fmt.Println("If you built from source, run: make build")
		fmt.Println("This will build all helper binaries including ptx-installer.")

		os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
