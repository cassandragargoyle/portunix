package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of Portunix MCP server integration with Claude Code",
	Long: `Check the current status of Portunix MCP server integration with Claude Code.

This command will:
- Check if Claude Code is installed and accessible
- Verify if Portunix MCP server is configured
- Display detailed configuration information

Examples:
  portunix mcp status           # Show status overview
  portunix mcp status --verbose # Show detailed information
  portunix mcp status --json    # Output in JSON format`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		if err := checkMCPStatus(verbose, jsonOutput); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check MCP status: %v\n", err)
			os.Exit(1)
		}
	},
}

type MCPStatus struct {
	ClaudeCodeInstalled bool   `json:"claude_code_installed"`
	MCPConfigured       bool   `json:"mcp_configured"`
	MCPServerRunning    bool   `json:"mcp_server_running"`
	PortunixPath        string `json:"portunix_path"`
	ConfiguredPort      int    `json:"configured_port,omitempty"`
	PortAvailable       bool   `json:"port_available"`
	SuggestedPorts      []int  `json:"suggested_ports,omitempty"`
}

func checkMCPStatus(verbose, jsonOutput bool) error {
	status := &MCPStatus{}

	// Check Claude Code installation
	status.ClaudeCodeInstalled = isClaudeCodeInstalled()

	// Get Portunix path
	if path, err := getPortunixExecutablePath(); err == nil {
		status.PortunixPath = path
	}

	// Check MCP configuration
	status.MCPConfigured = isMCPAlreadyConfigured()

	// Default port
	status.ConfiguredPort = 3001

	// Check port availability
	status.PortAvailable = isPortAvailable(status.ConfiguredPort)

	// Get suggested ports if current is not available
	if !status.PortAvailable {
		status.SuggestedPorts = findAvailablePorts(3)
	}

	// Check if MCP server is running
	status.MCPServerRunning = isMCPServerRunning()

	// Output results
	if jsonOutput {
		return outputStatusAsJSON(status)
	}
	return outputStatusAsText(status, verbose)
}

func isMCPServerRunning() bool {
	// Try to check with Claude if MCP server is responding
	claudePath, err := getClaudePath()
	if err != nil {
		return false
	}

	cmd := exec.Command(claudePath, "mcp", "get", "portunix")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	outputStr := string(output)
	return strings.Contains(outputStr, "✓") ||
		strings.Contains(outputStr, "✅") ||
		strings.Contains(outputStr, "Connected") ||
		strings.Contains(outputStr, "OK")
}

func outputStatusAsJSON(status *MCPStatus) error {
	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status to JSON: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func outputStatusAsText(status *MCPStatus, verbose bool) error {
	if !verbose {
		// Brief one-line status
		if status.ClaudeCodeInstalled && status.MCPConfigured && status.MCPServerRunning {
			fmt.Println("🎉 Portunix MCP: OPERATIONAL")
		} else if status.ClaudeCodeInstalled && status.MCPConfigured {
			fmt.Println("⚠️  Portunix MCP: CONFIGURED, NOT RUNNING")
		} else if status.ClaudeCodeInstalled {
			fmt.Println("⚠️  Portunix MCP: READY FOR CONFIGURATION")
		} else {
			fmt.Println("❌ Portunix MCP: NOT READY (install Claude Code)")
		}
		return nil
	}

	// Verbose output
	fmt.Println("📊 Portunix MCP Integration Status")
	fmt.Println("==================================")

	// Claude Code installation
	fmt.Print("Claude Code Installation: ")
	if status.ClaudeCodeInstalled {
		fmt.Println("✅ INSTALLED")
	} else {
		fmt.Println("❌ NOT FOUND")
		fmt.Println("  Install Claude Code: curl -fsSL https://claude.ai/cli/install.sh | sh")
	}

	// Portunix path
	fmt.Print("Portunix Executable: ")
	if status.PortunixPath != "" {
		fmt.Printf("✅ %s\n", status.PortunixPath)
	} else {
		fmt.Println("❌ NOT FOUND")
	}

	// MCP configuration
	fmt.Print("MCP Configuration: ")
	if status.MCPConfigured {
		fmt.Println("✅ CONFIGURED")
	} else {
		fmt.Println("❌ NOT CONFIGURED")
		fmt.Println("  Configure: portunix mcp configure")
	}

	// MCP server status
	fmt.Print("MCP Server Status: ")
	if status.MCPServerRunning {
		fmt.Printf("✅ RUNNING (port %d)\n", status.ConfiguredPort)
	} else {
		fmt.Println("❌ NOT RUNNING")
		if status.MCPConfigured {
			if !status.PortAvailable {
				fmt.Printf("  ⚠️  Port %d is occupied!\n", status.ConfiguredPort)
				if len(status.SuggestedPorts) > 0 {
					fmt.Printf("  Available ports: %v\n", status.SuggestedPorts)
				}
			} else {
				fmt.Println("  Start server: portunix mcp serve")
			}
		}
	}

	// Overall status
	fmt.Println("\nOverall Status:")
	if status.ClaudeCodeInstalled && status.MCPConfigured && status.MCPServerRunning {
		fmt.Println("🎉 FULLY OPERATIONAL")
	} else if status.ClaudeCodeInstalled && status.MCPConfigured {
		fmt.Println("⚠️  CONFIGURED BUT NOT RUNNING")
	} else if status.ClaudeCodeInstalled {
		fmt.Println("⚠️  READY FOR CONFIGURATION")
	} else {
		fmt.Println("❌ NOT READY")
	}

	// Quick commands
	fmt.Println("\nQuick Commands:")
	if !status.ClaudeCodeInstalled {
		fmt.Println("  curl -fsSL https://claude.ai/cli/install.sh | sh  # Install Claude Code")
	}
	if !status.MCPConfigured {
		fmt.Println("  portunix mcp configure                           # Configure integration")
	}
	if !status.MCPServerRunning && status.MCPConfigured {
		fmt.Println("  portunix mcp serve                               # Start MCP server")
	}
	if status.MCPConfigured {
		fmt.Println("  portunix mcp remove                              # Remove integration")
	}

	return nil
}

func init() {
	mcpCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolP("verbose", "v", false, "Show detailed information")
	statusCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
}
