package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"portunix.ai/app/mcp"
)

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server for AI assistant integration",
	Long: `Start Model Context Protocol (MCP) server to enable AI assistants 
to interact with Portunix functionality. The server provides a standardized 
interface for AI-driven development environment management.

Communication Modes:
  stdio    - Standard input/output for direct AI integration (default)
  tcp      - TCP socket server for network-based connections
  unix     - Unix domain socket for local IPC

Examples:
  portunix mcp serve                           # Start in stdio mode (default)
  portunix mcp serve --mode stdio              # Explicit stdio mode
  portunix mcp serve --mode tcp --port 3001    # TCP mode on port 3001
  portunix mcp serve --mode unix --socket /tmp/portunix.sock  # Unix socket mode`,
	Run: func(cmd *cobra.Command, args []string) {
		mode, _ := cmd.Flags().GetString("mode")
		port, _ := cmd.Flags().GetInt("port")
		socket, _ := cmd.Flags().GetString("socket")
		permissions, _ := cmd.Flags().GetString("permissions")
		config, _ := cmd.Flags().GetString("config")

		server := mcp.NewServer(port, permissions, config)

		switch mode {
		case "stdio":
			fmt.Fprintf(os.Stderr, "Starting MCP Server in stdio mode\n")
			fmt.Fprintf(os.Stderr, "Permission level: %s\n", permissions)
			if err := server.StartStdio(); err != nil {
				log.Fatalf("Failed to start MCP server in stdio mode: %v", err)
			}

		case "tcp":
			if port == 0 {
				port = 3001 // Default port
			}
			fmt.Printf("Starting MCP Server in TCP mode on port %d\n", port)
			fmt.Printf("Permission level: %s\n", permissions)
			if err := server.StartTCP(port); err != nil {
				log.Fatalf("Failed to start MCP server in TCP mode: %v", err)
			}

		case "unix":
			if socket == "" {
				socket = "/tmp/portunix.sock"
			}
			fmt.Printf("Starting MCP Server in Unix socket mode: %s\n", socket)
			fmt.Printf("Permission level: %s\n", permissions)
			if err := server.StartUnix(socket); err != nil {
				log.Fatalf("Failed to start MCP server in Unix socket mode: %v", err)
			}

		default:
			fmt.Fprintf(os.Stderr, "Unknown mode: %s. Supported modes: stdio, tcp, unix\n", mode)
			os.Exit(1)
		}
	},
}

func init() {
	mcpCmd.AddCommand(mcpServeCmd)

	mcpServeCmd.Flags().StringP("mode", "m", "stdio", "Communication mode: stdio, tcp, unix")
	mcpServeCmd.Flags().IntP("port", "p", 3001, "Port for TCP mode")
	mcpServeCmd.Flags().StringP("socket", "s", "/tmp/portunix.sock", "Socket path for Unix mode")
	mcpServeCmd.Flags().StringP("permissions", "r", "limited", "Permission level: limited, standard, full")
	mcpServeCmd.Flags().StringP("config", "c", "", "Path to configuration file")
}