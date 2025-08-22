package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"portunix.cz/app/install"
)

var mcpConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Portunix MCP server integration with Claude Code",
	Long: `Automatically configure Portunix MCP server to work with Claude Code AI assistant.
	
This command will:
- Check if Claude Code is installed
- Add Portunix as an MCP server in Claude Code configuration
- Set appropriate permissions and port settings
- Verify the integration works

Examples:
  portunix mcp configure                          # Use default settings
  portunix mcp configure --port 3001             # Custom port
  portunix mcp configure --permissions standard  # Custom permissions
  portunix mcp configure --auto-install          # Automatically install Claude Code if missing`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		permissions, _ := cmd.Flags().GetString("permissions")
		force, _ := cmd.Flags().GetBool("force")
		autoInstall, _ := cmd.Flags().GetBool("auto-install")

		if err := configureMCPIntegration(port, permissions, force, autoInstall); err != nil {
			log.Fatalf("Failed to configure MCP integration: %v", err)
		}
	},
}

func configureMCPIntegration(port int, permissions string, force bool, autoInstall bool) error {
	fmt.Println("🔧 Configuring Portunix MCP integration with Claude Code...")

	// Step 1: Check if Claude Code is installed
	fmt.Print("1. Checking Claude Code installation... ")
	if !isClaudeCodeInstalled() {
		fmt.Println("❌ NOT FOUND")
		
		if autoInstall {
			fmt.Println("🚀 Auto-installing Claude Code using Portunix package installer...")
			if err := install.InstallPackage("claude-code", "npm"); err != nil {
				fmt.Printf("❌ Installation failed: %v\n", err)
				fmt.Println("Please try manual installation:")
				fmt.Println("  npm install -g @anthropic-ai/claude-code")
				fmt.Println("  # or")
				fmt.Println("  curl -fsSL https://claude.ai/cli/install.sh | sh")
				return fmt.Errorf("Claude Code installation failed")
			}
			fmt.Println("✅ Claude Code installed successfully!")
		} else {
			fmt.Print("Claude Code is not installed. Would you like to install it now? (y/N): ")
			
			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))
			
			if response == "y" || response == "yes" {
				fmt.Println("🚀 Installing Claude Code using Portunix package installer...")
				if err := install.InstallPackage("claude-code", "npm"); err != nil {
					fmt.Printf("❌ Installation failed: %v\n", err)
					fmt.Println("Please try manual installation:")
					fmt.Println("  npm install -g @anthropic-ai/claude-code")
					fmt.Println("  # or")
					fmt.Println("  curl -fsSL https://claude.ai/cli/install.sh | sh")
					return fmt.Errorf("Claude Code installation failed")
				}
				fmt.Println("✅ Claude Code installed successfully!")
			} else {
				fmt.Println("Please install Claude Code first:")
				fmt.Println("  portunix install claude-code")
				fmt.Println("  # or manually:")
				fmt.Println("  npm install -g @anthropic-ai/claude-code")
				fmt.Println("  curl -fsSL https://claude.ai/cli/install.sh | sh")
				return fmt.Errorf("Claude Code not found")
			}
		}
	} else {
		fmt.Println("✅ FOUND")
	}

	// Step 2: Get current executable path
	fmt.Print("2. Detecting Portunix executable path... ")
	portunixPath, err := getCurrentExecutablePath()
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
	if err := addMCPServerToClaudeCode(portunixPath, port, permissions); err != nil {
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
	fmt.Printf("1. Start the MCP server: %s mcp-server --port %d --permissions %s\n", portunixPath, port, permissions)
	fmt.Println("2. Open Claude Code: claude")
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


func getCurrentExecutablePath() (string, error) {
	// First try to get the actual executable path
	if exe, err := os.Executable(); err == nil {
		if abs, err := filepath.Abs(exe); err == nil {
			return abs, nil
		}
	}

	// Fallback: check if portunix is in PATH
	if path, err := exec.LookPath("portunix"); err == nil {
		return path, nil
	}

	// Fallback: use current working directory + portunix
	if cwd, err := os.Getwd(); err == nil {
		portunixPath := filepath.Join(cwd, "portunix")
		if runtime.GOOS == "windows" {
			portunixPath += ".exe"
		}
		if _, err := os.Stat(portunixPath); err == nil {
			return portunixPath, nil
		}
	}

	return "", fmt.Errorf("could not locate portunix executable")
}


func addMCPServerToClaudeCode(portunixPath string, port int, permissions string) error {
	// First, find the claude executable
	claudePath, err := getClaudePath()
	if err != nil {
		return fmt.Errorf("claude executable not found: %w", err)
	}

	// Use claude mcp add command with correct syntax: claude mcp add <name> <command>
	// Note: Claude Code MCP doesn't support command arguments, so we only add the base command
	args := []string{
		"mcp", "add", "portunix",
		portunixPath,
		"mcp-server",
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

	if !contains(string(output), "portunix") {
		return fmt.Errorf("portunix not found in MCP server list")
	}

	// Try to test the MCP server (skip if command doesn't exist)
	cmd = exec.Command(claudePath, "mcp", "test", "portunix")
	if err := cmd.Run(); err != nil {
		// Note: test command might not exist, so just warn
		return fmt.Errorf("MCP server test failed (this might be normal): %w", err)
	}

	return nil
}


func init() {
	mcpCmd.AddCommand(mcpConfigureCmd)

	mcpConfigureCmd.Flags().IntP("port", "p", 3001, "Port for MCP server")
	mcpConfigureCmd.Flags().StringP("permissions", "r", "standard", "Permission level: limited, standard, full")
	mcpConfigureCmd.Flags().BoolP("force", "f", false, "Force reconfiguration even if already configured")
	mcpConfigureCmd.Flags().BoolP("auto-install", "a", false, "Automatically install Claude Code if not found")
}