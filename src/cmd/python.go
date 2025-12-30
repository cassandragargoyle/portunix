package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// pythonCmd represents the python command - delegates to ptx-python helper
var pythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Python development tools and utilities",
	Long: `Manage Python virtual environments, packages, and development tools.

This command provides comprehensive Python development capabilities including:
- Virtual environment (venv) management
- Package installation with pip
- Code building and distribution
- Code quality tools (linting, formatting)
- Multi-version Python management

The Python installation is managed through Portunix package system.

Examples:
  portunix python venv create myproject
  portunix python pip install requests
  portunix python venv list`,
	Run: func(cmd *cobra.Command, args []string) {
		// Try to dispatch to ptx-python helper first
		if helperPath := findPythonHelper(); helperPath != "" {
			if err := dispatchToPythonHelper(helperPath, args); err != nil {
				fmt.Fprintf(os.Stderr, "Helper execution failed: %v\n", err)
				fmt.Fprintf(os.Stderr, "Falling back to built-in python commands...\n")
				// Fall through to show help
			} else {
				return // Helper succeeded
			}
		}

		// Fallback: show help
		cmd.Help()
	},
}

// findPythonHelper finds the ptx-python helper binary
func findPythonHelper() string {
	// Determine binary name with platform suffix
	helperName := "ptx-python"
	if runtime.GOOS == "windows" {
		helperName += ".exe"
	}

	// Method 1: Check in same directory as main binary
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		helperPath := filepath.Join(execDir, helperName)
		if _, err := os.Stat(helperPath); err == nil {
			return helperPath
		}
	}

	// Method 2: Check in PATH
	if helperPath, err := exec.LookPath(helperName); err == nil {
		return helperPath
	}

	return "" // Helper not found
}

// dispatchToPythonHelper dispatches python commands to the ptx-python helper binary
func dispatchToPythonHelper(helperPath string, args []string) error {
	// Prepare arguments: ["python"] + original args
	cmdArgs := append([]string{"python"}, args...)

	// Execute helper binary
	cmd := exec.Command(helperPath, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(pythonCmd)
}
