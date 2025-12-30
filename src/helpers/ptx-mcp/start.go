package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start MCP server for AI assistant integration",
	Long: `Start the MCP server with specified configuration.

The server can run in different modes:
- stdio: For direct process communication (Claude Code)
- remote: For network communication (Claude Desktop)

Examples:
  portunix mcp start                    # Start with default settings
  portunix mcp start --port 3001        # Start on specific port
  portunix mcp start --daemon           # Run as daemon process
  portunix mcp start --stdio            # Run in stdio mode`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		daemon, _ := cmd.Flags().GetBool("daemon")
		stdio, _ := cmd.Flags().GetBool("stdio")
		permissions, _ := cmd.Flags().GetString("permissions")

		if err := startMCPServer(port, daemon, stdio, permissions); err != nil {
			fmt.Printf("❌ Failed to start MCP server: %v\n", err)
			os.Exit(1)
		}
	},
}

func startMCPServer(port int, daemon, stdio bool, permissions string) error {
	if isServerRunning() {
		fmt.Println("⚠️  MCP server is already running")
		fmt.Println("Use 'portunix mcp stop' to stop it first")
		return nil
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	args := []string{"mcp", "serve"}

	if stdio {
		args = append(args, "--mode", "stdio")
		fmt.Println("🚀 Starting MCP server in stdio mode...")
	} else {
		args = append(args, "--mode", "tcp", "--port", strconv.Itoa(port))
		fmt.Printf("🚀 Starting MCP server on port %d...\n", port)
	}

	args = append(args, "--permissions", permissions)

	if daemon {
		cmd := exec.Command(execPath, args...)
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		pidFile := getMCPPidFile()
		if err := savePID(pidFile, cmd.Process.Pid); err != nil {
			fmt.Printf("⚠️  Warning: Failed to save PID: %v\n", err)
		}

		fmt.Printf("✅ MCP server started as daemon (PID: %d)\n", cmd.Process.Pid)
		if !stdio {
			fmt.Printf("   Listening on port %d\n", port)
		}
		fmt.Println("\nManagement commands:")
		fmt.Println("  portunix mcp status   # Check server status")
		fmt.Println("  portunix mcp stop     # Stop the server")
		fmt.Println("  portunix mcp test     # Test connection")
	} else {
		cmd := exec.Command(execPath, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		fmt.Println("✅ Starting MCP server in foreground...")
		fmt.Println("Press Ctrl+C to stop")

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("server exited with error: %w", err)
		}
	}

	return nil
}

func init() {
	mcpCmd.AddCommand(startCmd)

	startCmd.Flags().IntP("port", "p", 3001, "Port to run MCP server on")
	startCmd.Flags().BoolP("daemon", "d", false, "Run as daemon process")
	startCmd.Flags().BoolP("stdio", "s", false, "Run in stdio mode")
	startCmd.Flags().StringP("permissions", "r", "standard", "Permission level: limited, standard, full")
}
