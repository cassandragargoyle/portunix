package cmd

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var mcpServerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show MCP server status and health information",
	Long: `Display current status of the MCP server including:
- Running state
- Process information
- Configuration details
- Connected assistants

Examples:
  portunix mcp-server status          # Show server status
  portunix mcp-server status --json   # Output in JSON format`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		
		if jsonOutput {
			showJSONStatus()
		} else {
			showTextStatus()
		}
	},
}

func showTextStatus() {
	fmt.Println("🔍 MCP Server Status")
	fmt.Println("=" + "==================")
	
	// Check if server is running
	pidFile := getMCPPidFile()
	if _, err := os.Stat(pidFile); err != nil {
		fmt.Println("\n📊 Server State: ❌ Not Running")
		fmt.Println("\nStart the server with:")
		fmt.Println("  portunix mcp-server start")
		return
	}
	
	// Read PID
	data, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Printf("\n📊 Server State: ⚠️  Unknown (failed to read PID: %v)\n", err)
		return
	}
	
	var pid int
	fmt.Sscanf(string(data), "%d", &pid)
	
	// Check if process is actually running
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Println("\n📊 Server State: ❌ Not Running (stale PID file)")
		os.Remove(pidFile)
		return
	}
	
	// On Unix, check if process is alive
	if runtime.GOOS != "windows" {
		if err := process.Signal(os.Signal(nil)); err != nil {
			fmt.Println("\n📊 Server State: ❌ Not Running (process died)")
			os.Remove(pidFile)
			return
		}
	}
	
	fmt.Println("\n📊 Server State: ✅ Running")
	fmt.Printf("   Process ID: %d\n", pid)
	
	// Get PID file modification time as start time approximation
	if stat, err := os.Stat(pidFile); err == nil {
		uptime := time.Since(stat.ModTime())
		fmt.Printf("   Uptime: %s\n", formatDuration(uptime))
	}
	
	// Show configuration (if available)
	fmt.Println("\n⚙️  Configuration:")
	if isMCPAlreadyConfigured() {
		fmt.Println("   Claude Code: ✅ Configured")
		
		// Check for other assistants
		if isClaudeDesktopInstalled() {
			fmt.Println("   Claude Desktop: ⚠️  Available but not configured")
		}
		if isGeminiCLIInstalled() {
			fmt.Println("   Gemini CLI: ⚠️  Available but not configured")
		}
	} else {
		fmt.Println("   No assistants configured")
		fmt.Println("\n   Run 'portunix mcp-server init' to configure")
	}
	
	// Show available tools
	fmt.Println("\n🔧 Available Tools:")
	fmt.Println("   - system_info: Get system information")
	fmt.Println("   - package_install: Install packages")
	fmt.Println("   - project_detect: Detect project type")
	fmt.Println("   - container_manage: Manage containers")
	fmt.Println("   - vm_manage: Manage virtual machines")
	
	// Show management commands
	fmt.Println("\n📝 Management Commands:")
	fmt.Println("   portunix mcp-server stop    # Stop the server")
	fmt.Println("   portunix mcp-server test    # Test connection")
	fmt.Println("   portunix mcp-server config  # View configuration")
}

func showJSONStatus() {
	// Simple JSON output for programmatic use
	pidFile := getMCPPidFile()
	running := false
	pid := 0
	
	if _, err := os.Stat(pidFile); err == nil {
		if data, err := os.ReadFile(pidFile); err == nil {
			fmt.Sscanf(string(data), "%d", &pid)
			if process, err := os.FindProcess(pid); err == nil {
				if runtime.GOOS == "windows" {
					running = true
				} else {
					running = process.Signal(os.Signal(nil)) == nil
				}
			}
		}
	}
	
	fmt.Printf(`{
  "running": %t,
  "pid": %d,
  "configured": %t,
  "assistants": {
    "claude-code": %t,
    "claude-desktop": %t,
    "gemini-cli": %t
  }
}
`, running, pid, isMCPAlreadyConfigured(), 
   isClaudeCodeInstalled(), 
   isClaudeDesktopInstalled(),
   isGeminiCLIInstalled())
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func init() {
	mcpCmd.AddCommand(mcpServerStatusCmd)
	
	mcpServerStatusCmd.Flags().Bool("json", false, "Output in JSON format")
}