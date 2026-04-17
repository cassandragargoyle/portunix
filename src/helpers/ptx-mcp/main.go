/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"portunix.ai/app/mcp"
)

var version = "dev"

// mcpCmd is the main mcp command group
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP (Model Context Protocol) integration with AI assistants",
	Long: `Manage MCP server integration with AI assistants like Claude Code.

This command provides utilities to configure, manage and monitor
the MCP server integration with various AI development tools.`,
}

// serveCmd starts the MCP server
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server for AI assistant integration",
	Long: `Start Model Context Protocol (MCP) server to enable AI assistants
to interact with Portunix functionality.

Communication Modes:
  stdio    - Standard input/output for direct AI integration (default)
  tcp      - TCP socket server for network-based connections
  unix     - Unix domain socket for local IPC

Examples:
  portunix mcp serve                           # Start in stdio mode (default)
  portunix mcp serve --mode stdio              # Explicit stdio mode
  portunix mcp serve --mode tcp --port 3001    # TCP mode on port 3001`,
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
			fmt.Printf("Starting MCP Server in TCP mode on port %d\n", port)
			fmt.Printf("Permission level: %s\n", permissions)
			if err := server.StartTCP(port); err != nil {
				log.Fatalf("Failed to start MCP server in TCP mode: %v", err)
			}

		case "unix":
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

// rootCmd represents the base command for ptx-mcp
var rootCmd = &cobra.Command{
	Use:   "portunix",
	Short: "Portunix MCP Server",
	Long: `Portunix MCP (Model Context Protocol) server for AI assistant integration.

This helper handles all MCP server operations and is invoked by the main
portunix dispatcher when running 'portunix mcp' commands.`,
	Version: version,
}

func init() {
	// Add mcp command to root
	rootCmd.AddCommand(mcpCmd)

	// Add serve command to mcp
	mcpCmd.AddCommand(serveCmd)

	// Serve command flags
	serveCmd.Flags().StringP("mode", "m", "stdio", "Communication mode: stdio, tcp, unix")
	serveCmd.Flags().IntP("port", "p", 3001, "Port for TCP mode")
	serveCmd.Flags().StringP("socket", "s", "/tmp/portunix.sock", "Socket path for Unix mode")
	serveCmd.Flags().StringP("permissions", "r", "limited", "Permission level: limited, standard, full")
	serveCmd.Flags().StringP("config", "c", "", "Path to configuration file")

	// Version template
	rootCmd.SetVersionTemplate("portunix mcp version {{.Version}}\n")
}

func showHelpAI() {
	type CommandInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	type AIHelp struct {
		Tool        string        `json:"tool"`
		Version     string        `json:"version"`
		Description string        `json:"description"`
		Commands    []CommandInfo `json:"commands"`
	}
	help := AIHelp{
		Tool:        "ptx-mcp",
		Version:     version,
		Description: "Model Context Protocol server for AI assistant integration",
		Commands: []CommandInfo{
			{Name: "mcp serve", Description: "Start MCP server for AI assistant integration"},
		},
	}
	data, _ := json.MarshalIndent(help, "", "  ")
	fmt.Println(string(data))
}

func showHelpExpert() {
	fmt.Printf("PTX-MCP v%s - Model Context Protocol Server\n", version)
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  MCP (Model Context Protocol) server enabling AI assistants like")
	fmt.Println("  Claude Code to interact with Portunix functionality.")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  mcp serve                Start MCP server")
	fmt.Println()
	fmt.Println("SERVE FLAGS:")
	fmt.Println("  -m, --mode <mode>        Communication mode: stdio, tcp, unix (default: stdio)")
	fmt.Println("  -p, --port <port>        Port for TCP mode (default: 3001)")
	fmt.Println("  -s, --socket <path>      Socket path for Unix mode (default: /tmp/portunix.sock)")
	fmt.Println("  -r, --permissions <level> Permission level: limited, standard, full (default: limited)")
	fmt.Println("  -c, --config <path>      Path to configuration file")
	fmt.Println()
	fmt.Println("COMMUNICATION MODES:")
	fmt.Println("  stdio    Standard input/output for direct AI integration (default)")
	fmt.Println("  tcp      TCP socket server for network-based connections")
	fmt.Println("  unix     Unix domain socket for local IPC")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  portunix mcp serve                           Start in stdio mode")
	fmt.Println("  portunix mcp serve --mode tcp --port 3001    TCP mode on port 3001")
	fmt.Println("  portunix mcp serve --mode unix               Unix socket mode")
	fmt.Println("  portunix mcp serve --permissions full        Full permissions")
}

func main() {
	// Handle --help-ai and --help-expert before cobra processing
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--help-ai":
			showHelpAI()
			return
		case "--help-expert":
			showHelpExpert()
			return
		}
	}

	// Handle dispatched commands - when called via dispatcher, args[0] is "mcp"
	if len(os.Args) > 1 && os.Args[0] != "mcp" {
		// Called directly as ptx-mcp, shift args to include mcp
		// e.g., ptx-mcp mcp serve -> process as mcp serve
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
