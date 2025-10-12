package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
)

var mcpServerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start MCP server for AI assistant integration",
	Long: `Start the MCP server with specified configuration.

The server can run in different modes:
- stdio: For direct process communication (Claude Code)
- remote: For network communication (Claude Desktop)

Examples:
  portunix mcp-server start                    # Start with default settings
  portunix mcp-server start --port 3001        # Start on specific port
  portunix mcp-server start --daemon           # Run as daemon process
  portunix mcp-server start --stdio            # Run in stdio mode`,
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
	// Check if server is already running
	if isServerRunning() {
		fmt.Println("⚠️  MCP server is already running")
		fmt.Println("Use 'portunix mcp-server stop' to stop it first")
		return nil
	}
	
	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Build command arguments
	args := []string{"mcp-server"}
	
	if stdio {
		args = append(args, "--stdio")
		fmt.Println("🚀 Starting MCP server in stdio mode...")
	} else {
		args = append(args, "--port", strconv.Itoa(port))
		fmt.Printf("🚀 Starting MCP server on port %d...\n", port)
	}
	
	args = append(args, "--permissions", permissions)
	
	if daemon {
		// Start as daemon process
		cmd := exec.Command(execPath, args...)
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
		
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}
		
		// Save PID for later management
		pidFile := getMCPPidFile()
		if err := savePID(pidFile, cmd.Process.Pid); err != nil {
			fmt.Printf("⚠️  Warning: Failed to save PID: %v\n", err)
		}
		
		fmt.Printf("✅ MCP server started as daemon (PID: %d)\n", cmd.Process.Pid)
		if !stdio {
			fmt.Printf("   Listening on port %d\n", port)
		}
		fmt.Println("\nManagement commands:")
		fmt.Println("  portunix mcp-server status  # Check server status")
		fmt.Println("  portunix mcp-server stop    # Stop the server")
		fmt.Println("  portunix mcp-server test    # Test connection")
	} else {
		// Run in foreground
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

func isServerRunning() bool {
	pidFile := getMCPPidFile()
	if _, err := os.Stat(pidFile); err != nil {
		return false
	}
	
	// Read PID and check if process exists
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}
	
	pid := 0
	fmt.Sscanf(string(data), "%d", &pid)
	if pid == 0 {
		return false
	}
	
	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	// On Unix, sending signal 0 checks if process exists
	if runtime.GOOS != "windows" {
		err = process.Signal(os.Signal(nil))
		return err == nil
	}
	
	// On Windows, we can't easily check, so assume it's running
	return true
}

func getMCPPidFile() string {
	// Use system temp directory for PID file
	return filepath.Join(os.TempDir(), "portunix-mcp-server.pid")
}

func savePID(pidFile string, pid int) error {
	return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}

func init() {
	mcpCmd.AddCommand(mcpServerStartCmd)
	
	mcpServerStartCmd.Flags().IntP("port", "p", 3001, "Port to run MCP server on")
	mcpServerStartCmd.Flags().BoolP("daemon", "d", false, "Run as daemon process")
	mcpServerStartCmd.Flags().BoolP("stdio", "s", false, "Run in stdio mode")
	mcpServerStartCmd.Flags().StringP("permissions", "r", "standard", "Permission level: limited, standard, full")
}