package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var mcpReconfigureCmd = &cobra.Command{
	Use:   "reconfigure",
	Short: "Reconfigure Portunix MCP server with new settings (e.g., different port)",
	Long: `Reconfigure existing Portunix MCP server integration with new settings.
This command will remove the current configuration and set up a new one with
the specified parameters.

Examples:
  portunix mcp reconfigure --port 3002         # Change to port 3002
  portunix mcp reconfigure --permissions full  # Change permissions
  portunix mcp reconfigure --auto-port         # Auto-find available port`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		permissions, _ := cmd.Flags().GetString("permissions")
		autoPort, _ := cmd.Flags().GetBool("auto-port")

		if err := reconfigureMCPIntegration(port, permissions, autoPort); err != nil {
			log.Fatalf("Failed to reconfigure MCP integration: %v", err)
		}
	},
}

func reconfigureMCPIntegration(port int, permissions string, autoPort bool) error {
	fmt.Println("🔄 Reconfiguring Portunix MCP integration...")

	// Check current configuration
	fmt.Print("1. Checking current configuration... ")
	if !isMCPAlreadyConfigured() {
		fmt.Println("❌ NOT CONFIGURED")
		fmt.Println("No existing configuration found. Use 'portunix mcp configure' instead.")
		return fmt.Errorf("no existing MCP configuration found")
	}
	fmt.Println("✅ FOUND")

	// Auto-detect available port if requested
	if autoPort {
		fmt.Print("2. Finding available port... ")
		availablePorts := findAvailablePorts(1)
		if len(availablePorts) == 0 {
			fmt.Println("❌ NO AVAILABLE PORTS")
			return fmt.Errorf("no available ports found")
		}
		port = availablePorts[0]
		fmt.Printf("✅ PORT %d\n", port)
	} else if port == 0 {
		// Use default port if not specified
		port = 3001
	}

	// Check if new port is available
	if !autoPort {
		fmt.Printf("2. Checking port %d availability... ", port)
		if !isPortAvailable(port) {
			fmt.Println("❌ PORT OCCUPIED")
			availablePorts := findAvailablePorts(3)
			if len(availablePorts) > 0 {
				fmt.Printf("   Available ports: %v\n", availablePorts)
				fmt.Printf("   Suggestion: portunix mcp reconfigure --port %d\n", availablePorts[0])
			}
			return fmt.Errorf("port %d is not available", port)
		}
		fmt.Println("✅ AVAILABLE")
	}

	// Remove current configuration
	fmt.Print("3. Removing current configuration... ")
	if err := removeMCPIntegration(true); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		return fmt.Errorf("failed to remove current configuration: %w", err)
	}
	fmt.Println("✅ REMOVED")

	// Configure with new settings
	fmt.Printf("4. Configuring with new settings (port: %d, permissions: %s)... ", port, permissions)
	if err := configureMCPIntegration(port, permissions, true, false); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		return fmt.Errorf("failed to configure with new settings: %w", err)
	}
	fmt.Println("✅ CONFIGURED")

	// Final status
	fmt.Println("\n🎉 Reconfiguration completed successfully!")
	fmt.Printf("   MCP Server Port: %d\n", port)
	fmt.Printf("   Permissions: %s\n", permissions)
	fmt.Printf("   Start command: portunix mcp-server --port %d\n", port)

	return nil
}

func init() {
	mcpCmd.AddCommand(mcpReconfigureCmd)

	mcpReconfigureCmd.Flags().IntP("port", "p", 0, "Port for MCP server (0 = use default 3001)")
	mcpReconfigureCmd.Flags().StringP("permissions", "r", "limited", "Permission level: limited, standard, full")
	mcpReconfigureCmd.Flags().BoolP("auto-port", "a", false, "Automatically find available port")
}
