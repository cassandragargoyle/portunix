package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var reconfigureCmd = &cobra.Command{
	Use:   "reconfigure",
	Short: "Reconfigure Portunix MCP server with new settings (e.g., different mode or port)",
	Long: `Reconfigure existing Portunix MCP server integration with new settings.
This command will remove the current configuration and set up a new one with
the specified parameters.

Examples:
  portunix mcp reconfigure                          # Reconfigure with stdio mode, local scope (default)
  portunix mcp reconfigure --scope user             # Reconfigure in global/user scope
  portunix mcp reconfigure --mode tcp --port 3002   # Change to TCP mode on port 3002
  portunix mcp reconfigure --permissions full       # Change permissions
  portunix mcp reconfigure --auto-port              # Auto-find available port (TCP mode)
  portunix mcp reconfigure --force                  # Force reconfigure even if not configured`,
	Run: func(cmd *cobra.Command, args []string) {
		mode, _ := cmd.Flags().GetString("mode")
		scope, _ := cmd.Flags().GetString("scope")
		port, _ := cmd.Flags().GetInt("port")
		permissions, _ := cmd.Flags().GetString("permissions")
		autoPort, _ := cmd.Flags().GetBool("auto-port")
		force, _ := cmd.Flags().GetBool("force")

		if err := reconfigureMCPIntegration(mode, scope, port, permissions, autoPort, force); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to reconfigure MCP integration: %v\n", err)
			os.Exit(1)
		}
	},
}

func reconfigureMCPIntegration(mode string, scope string, port int, permissions string, autoPort, force bool) error {
	fmt.Printf("🔄 Reconfiguring Portunix MCP integration (mode: %s, scope: %s)...\n", mode, scope)

	// Check current configuration
	fmt.Print("1. Checking current configuration... ")
	if !isMCPAlreadyConfigured() {
		if force {
			fmt.Println("⚠️  NOT CONFIGURED (force mode - will configure)")
		} else {
			fmt.Println("❌ NOT CONFIGURED")
			fmt.Println("No existing configuration found. Use 'portunix mcp configure' or --force.")
			return fmt.Errorf("no existing MCP configuration found")
		}
	} else {
		fmt.Println("✅ FOUND")
	}

	// Port handling only relevant for TCP mode
	if mode == "tcp" {
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
					fmt.Printf("   Suggestion: portunix mcp reconfigure --mode tcp --port %d\n", availablePorts[0])
				}
				return fmt.Errorf("port %d is not available", port)
			}
			fmt.Println("✅ AVAILABLE")
		}
	} else {
		fmt.Printf("2. Mode %s does not require port configuration... ✅ SKIPPED\n", mode)
	}

	// Remove current configuration (force mode to skip confirmation)
	fmt.Print("3. Removing current configuration... ")
	if isMCPAlreadyConfigured() {
		if err := removeMCPIntegration(true); err != nil {
			fmt.Printf("❌ FAILED: %v\n", err)
			return fmt.Errorf("failed to remove current configuration: %w", err)
		}
		fmt.Println("✅ REMOVED")
	} else {
		fmt.Println("⏭️  SKIPPED (no existing config)")
	}

	// Configure with new settings (force mode since we just removed)
	fmt.Printf("4. Configuring with new settings (mode: %s, scope: %s, permissions: %s)... ", mode, scope, permissions)
	if err := configureMCPIntegration(mode, scope, port, permissions, true); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		return fmt.Errorf("failed to configure with new settings: %w", err)
	}
	fmt.Println("✅ CONFIGURED")

	// Final status
	fmt.Println("\n🎉 Reconfiguration completed successfully!")
	fmt.Printf("   MCP Server Mode: %s\n", mode)
	fmt.Printf("   Configuration Scope: %s\n", scope)
	fmt.Printf("   Permissions: %s\n", permissions)
	switch mode {
	case "stdio":
		fmt.Println("   Claude Code will automatically start the MCP server via stdio")
	case "tcp":
		fmt.Printf("   MCP Server Port: %d\n", port)
		fmt.Printf("   Start command: portunix mcp serve --mode tcp --port %d\n", port)
	case "unix":
		fmt.Println("   Start command: portunix mcp serve --mode unix")
	}

	return nil
}

func init() {
	mcpCmd.AddCommand(reconfigureCmd)

	reconfigureCmd.Flags().StringP("mode", "m", "stdio", "Server mode: stdio (default, recommended), tcp, unix")
	reconfigureCmd.Flags().StringP("scope", "s", "local", "Configuration scope: local (default, .mcp.json), user (global), project")
	reconfigureCmd.Flags().IntP("port", "p", 0, "Port for MCP server (only used with --mode tcp, 0 = use default 3001)")
	reconfigureCmd.Flags().StringP("permissions", "r", "limited", "Permission level: limited, standard, full")
	reconfigureCmd.Flags().BoolP("auto-port", "a", false, "Automatically find available port (only for --mode tcp)")
	reconfigureCmd.Flags().BoolP("force", "f", false, "Force reconfigure even if not configured")
}
