package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"portunix.cz/app/mcp"
)

var mcpServerCmd = &cobra.Command{
	Use:   "mcp-server",
	Short: "Start MCP server for AI assistant integration",
	Long: `Start Model Context Protocol (MCP) server to enable AI assistants 
to interact with Portunix functionality. The server provides a standardized 
interface for AI-driven development environment management.

Examples:
  portunix mcp-server                    # Start with default settings
  portunix mcp-server --port 3001        # Start on specific port
  portunix mcp-server --daemon           # Run as daemon`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		daemon, _ := cmd.Flags().GetBool("daemon")
		permissions, _ := cmd.Flags().GetString("permissions")
		config, _ := cmd.Flags().GetString("config")
		stdio, _ := cmd.Flags().GetBool("stdio")

		// Auto-detect stdio mode if running from Claude Code
		if !cmd.Flags().Changed("stdio") && mcp.IsRunningFromClaudeCode() {
			stdio = true
		}

		server := mcp.NewServer(port, permissions, config)
		
		if stdio {
			fmt.Fprintf(os.Stderr, "Starting MCP Server in stdio mode\n")
			fmt.Fprintf(os.Stderr, "Permission level: %s\n", permissions)
			if err := server.StartStdio(); err != nil {
				log.Fatalf("Failed to start MCP server in stdio mode: %v", err)
			}
		} else {
			fmt.Printf("Starting MCP Server on port %d\n", port)
			if daemon {
				fmt.Println("Running in daemon mode")
			}
			fmt.Printf("Permission level: %s\n", permissions)
			if err := server.Start(daemon); err != nil {
				log.Fatalf("Failed to start MCP server: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpServerCmd)

	mcpServerCmd.Flags().IntP("port", "p", 3001, "Port to run MCP server on")
	mcpServerCmd.Flags().BoolP("daemon", "d", false, "Run as daemon process")
	mcpServerCmd.Flags().BoolP("stdio", "s", false, "Run in stdio mode (for Claude Code integration)")
	mcpServerCmd.Flags().StringP("permissions", "r", "limited", "Permission level: limited, standard, full")
	mcpServerCmd.Flags().StringP("config", "c", "", "Path to configuration file")
}