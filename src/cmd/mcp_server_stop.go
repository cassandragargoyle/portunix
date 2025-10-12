package cmd

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
)

var mcpServerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop running MCP server",
	Long: `Stop the running MCP server process.

This command will gracefully shutdown the MCP server if it's running
as a daemon process.

Examples:
  portunix mcp-server stop        # Stop the server
  portunix mcp-server stop --force # Force stop if graceful shutdown fails`,
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
	
	// Check if PID file exists
	if _, err := os.Stat(pidFile); err != nil {
		return fmt.Errorf("MCP server is not running (no PID file found)")
	}
	
	// Read PID
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}
	
	var pid int
	fmt.Sscanf(string(data), "%d", &pid)
	if pid == 0 {
		return fmt.Errorf("invalid PID in file")
	}
	
	// Find process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}
	
	// Send termination signal
	fmt.Printf("Stopping MCP server (PID: %d)...\n", pid)
	
	if runtime.GOOS == "windows" {
		// On Windows, use Kill directly
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to stop process: %w", err)
		}
	} else {
		// On Unix, try graceful shutdown first
		if force {
			err = process.Signal(syscall.SIGKILL)
		} else {
			err = process.Signal(syscall.SIGTERM)
		}
		
		if err != nil {
			return fmt.Errorf("failed to send signal: %w", err)
		}
	}
	
	// Remove PID file
	if err := os.Remove(pidFile); err != nil {
		fmt.Printf("⚠️  Warning: Failed to remove PID file: %v\n", err)
	}
	
	return nil
}

func init() {
	mcpCmd.AddCommand(mcpServerStopCmd)
	
	mcpServerStopCmd.Flags().BoolP("force", "f", false, "Force stop the server")
}