package cmd

import (
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP (Model Context Protocol) integration with AI assistants",
	Long: `Manage MCP server integration with AI assistants like Claude Code.
	
This command provides utilities to configure, manage and monitor
the MCP server integration with various AI development tools.`,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
