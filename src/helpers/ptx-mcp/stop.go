package main

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop running MCP server",
	Long: `Stop the running MCP server process.

This command will gracefully shutdown the MCP server if it's running
as a daemon process.

Examples:
  portunix mcp stop        # Stop the server
  portunix mcp stop --force # Force stop if graceful shutdown fails`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		if err := stopMCPServer(force); err != nil {
			fmt.Printf("❌ Failed to stop MCP server: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ MCP server stopped successfully")
	},
}

func stopMCPServer(force bool) error {
	pidFile := getMCPPidFile()

	if _, err := os.Stat(pidFile); err != nil {
		return fmt.Errorf("MCP server is not running (no PID file found)")
	}

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	var pid int
	fmt.Sscanf(string(data), "%d", &pid)
	if pid == 0 {
		return fmt.Errorf("invalid PID in file")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	fmt.Printf("Stopping MCP server (PID: %d)...\n", pid)

	if runtime.GOOS == "windows" {
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to stop process: %w", err)
		}
	} else {
		if force {
			err = process.Signal(syscall.SIGKILL)
		} else {
			err = process.Signal(syscall.SIGTERM)
		}

		if err != nil {
			return fmt.Errorf("failed to send signal: %w", err)
		}
	}

	if err := os.Remove(pidFile); err != nil {
		fmt.Printf("⚠️  Warning: Failed to remove PID file: %v\n", err)
	}

	return nil
}

func init() {
	mcpCmd.AddCommand(stopCmd)

	stopCmd.Flags().BoolP("force", "f", false, "Force stop the server")
}
