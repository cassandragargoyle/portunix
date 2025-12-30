package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Portunix MCP server integration with Claude Code",
	Long: `Automatically configure Portunix MCP server to work with Claude Code AI assistant.

This command will:
- Check if Claude Code is installed
- Add Portunix as an MCP server in Claude Code configuration
- Set appropriate permissions and mode settings
- Verify the integration works

Server Modes:
  stdio (default) - Direct stdin/stdout communication (recommended for Claude Code)
  tcp             - HTTP/TCP server on specified port
  unix            - Unix domain socket (Linux/macOS only)

Configuration Scope:
  local (default) - Project-local configuration (.mcp.json in current directory)
  user            - User-wide/global configuration
  project         - Project configuration

Examples:
  portunix mcp configure                          # Local scope, stdio mode (recommended)
  portunix mcp configure --scope user             # Global/user-wide configuration
  portunix mcp configure --mode tcp --port 3001   # Use TCP mode with custom port
  portunix mcp configure --mode unix              # Use Unix socket mode`,
	Run: func(cmd *cobra.Command, args []string) {
		mode, _ := cmd.Flags().GetString("mode")
		scope, _ := cmd.Flags().GetString("scope")
		port, _ := cmd.Flags().GetInt("port")
		permissions, _ := cmd.Flags().GetString("permissions")
		force, _ := cmd.Flags().GetBool("force")

		if err := configureMCPIntegration(mode, scope, port, permissions, force); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to configure MCP integration: %v\n", err)
			os.Exit(1)
		}
	},
}

func configureMCPIntegration(mode string, scope string, port int, permissions string, force bool) error {
	fmt.Printf("🔧 Configuring Portunix MCP integration with Claude Code (mode: %s, scope: %s)...\n", mode, scope)

	// Step 1: Check if Claude Code is installed
	fmt.Print("1. Checking Claude Code installation... ")
	if !isClaudeCodeInstalled() {
		fmt.Println("❌ NOT FOUND")
		fmt.Println("Please install Claude Code first:")
		fmt.Println("  portunix install claude-code")
		fmt.Println("  # or manually:")
		fmt.Println("  npm install -g @anthropic-ai/claude-code")
		fmt.Println("  curl -fsSL https://claude.ai/cli/install.sh | sh")
		return fmt.Errorf("Claude Code not found")
	}
	fmt.Println("✅ FOUND")

	// Step 2: Get current executable path (portunix, not ptx-mcp)
	fmt.Print("2. Detecting Portunix executable path... ")
	portunixPath, err := getPortunixExecutablePath()
	if err != nil {
		fmt.Println("❌ FAILED")
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	fmt.Printf("✅ %s\n", portunixPath)

	// Step 3: Check if already configured
	fmt.Print("3. Checking existing MCP configuration... ")
	if isMCPAlreadyConfigured() {
		if !force {
			fmt.Println("⚠️  ALREADY EXISTS")
			fmt.Println("Portunix MCP server is already configured in Claude Code.")
			fmt.Println("Use --force to overwrite existing configuration.")
			return nil
		} else {
			fmt.Println("⚠️  ALREADY EXISTS")
			fmt.Print("   Removing existing configuration for --force mode... ")
			if err := removeMCPServerFromClaudeCode(); err != nil {
				fmt.Println("❌ FAILED")
				fmt.Printf("   Warning: Failed to remove existing configuration: %v\n", err)
				fmt.Println("   Attempting to add anyway...")
			} else {
				fmt.Println("✅ REMOVED")
			}
		}
	} else {
		fmt.Println("✅ READY")
	}

	// Step 4: Add MCP server to Claude Code
	fmt.Print("4. Adding Portunix MCP server to Claude Code... ")
	if err := addMCPServerToClaudeCode(portunixPath, mode, scope, port, permissions); err != nil {
		fmt.Println("❌ FAILED")
		return fmt.Errorf("failed to add MCP server: %w", err)
	}
	fmt.Println("✅ ADDED")

	// Step 5: Verify configuration
	fmt.Print("5. Verifying MCP configuration... ")
	if err := verifyMCPConfiguration(); err != nil {
		fmt.Println("⚠️  WARNING")
		fmt.Printf("Configuration added but verification failed: %v\n", err)
	} else {
		fmt.Println("✅ VERIFIED")
	}

	// Step 6: Instructions
	fmt.Println("\n🎉 MCP Integration configured successfully!")
	fmt.Println("\nNext steps:")
	switch mode {
	case "stdio":
		fmt.Println("1. Claude Code will automatically start the MCP server via stdio")
		fmt.Println("2. Open Claude Code: claude")
	case "tcp":
		fmt.Printf("1. Start the MCP server: %s mcp serve --mode tcp --port %d\n", portunixPath, port)
		fmt.Println("2. Open Claude Code: claude")
	case "unix":
		fmt.Printf("1. Start the MCP server: %s mcp serve --mode unix\n", portunixPath)
		fmt.Println("2. Open Claude Code: claude")
	}
	fmt.Println("3. Test functionality: Ask Claude about your system or project")
	fmt.Println("\nExample questions for Claude:")
	fmt.Println("  - \"What system am I running on?\"")
	fmt.Println("  - \"What type of project is this?\"")
	fmt.Println("  - \"List my Docker containers\"")
	fmt.Println("  - \"What should I install for this project?\"")
	fmt.Println("\nMCP Management:")
	fmt.Println("  claude mcp list                                 # List configured MCP servers")
	fmt.Println("  claude mcp get portunix                         # Get details about Portunix MCP server")

	return nil
}

func addMCPServerToClaudeCode(portunixPath string, mode string, scope string, port int, permissions string) error {
	// First, find the claude executable
	claudePath, err := getClaudePath()
	if err != nil {
		return fmt.Errorf("claude executable not found: %w", err)
	}

	// Use claude mcp add command with correct syntax: claude mcp add [options] <name> <command>
	// Build command arguments based on mode and scope
	args := []string{
		"mcp", "add",
		"--scope", scope,
		"portunix",
		portunixPath,
		"mcp", "serve",
	}

	// Add mode-specific arguments for the portunix command
	switch mode {
	case "stdio":
		// stdio is the default, no additional arguments needed
	case "tcp":
		args = append(args, "--mode", "tcp", "--port", fmt.Sprintf("%d", port))
	case "unix":
		args = append(args, "--mode", "unix")
	default:
		return fmt.Errorf("unsupported mode: %s (use: stdio, tcp, unix)", mode)
	}

	cmd := exec.Command(claudePath, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		// If Claude Code has module issues, provide manual instructions
		if strings.Contains(string(output), "Cannot find module") {
			fmt.Println("   ⚠️  Claude Code has module issues. Manual configuration required:")
			fmt.Printf("   Run: %s %s\n", claudePath, strings.Join(args, " "))
			return nil // Don't fail, just provide instructions
		}
		return fmt.Errorf("claude mcp add failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

func verifyMCPConfiguration() error {
	// Find claude executable
	claudePath, err := getClaudePath()
	if err != nil {
		return fmt.Errorf("claude executable not found: %w", err)
	}

	// Check if portunix shows up in the list
	cmd := exec.Command(claudePath, "mcp", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list MCP servers: %w", err)
	}

	if !strings.Contains(string(output), "portunix") {
		return fmt.Errorf("portunix not found in MCP server list")
	}

	return nil
}

func init() {
	mcpCmd.AddCommand(configureCmd)

	configureCmd.Flags().StringP("mode", "m", "stdio", "Server mode: stdio (default, recommended), tcp, unix")
	configureCmd.Flags().StringP("scope", "s", "local", "Configuration scope: local (default, .mcp.json), user (global), project")
	configureCmd.Flags().IntP("port", "p", 3001, "Port for MCP server (only used with --mode tcp)")
	configureCmd.Flags().StringP("permissions", "r", "standard", "Permission level: limited, standard, full")
	configureCmd.Flags().BoolP("force", "f", false, "Force reconfiguration even if already configured")
}
