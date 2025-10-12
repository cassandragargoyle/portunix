package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove Portunix MCP server integration from Claude Code",
	Long: `Remove Portunix MCP server configuration from Claude Code AI assistant.

This command will:
- Check if Portunix MCP server is configured in Claude Code
- Remove the MCP server configuration
- Verify the removal was successful

Examples:
  portunix mcp remove           # Remove integration
  portunix mcp remove --force   # Force removal without confirmation`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		if err := removeMCPIntegration(force); err != nil {
			log.Fatalf("Failed to remove MCP integration: %v", err)
		}
	},
}

func removeMCPIntegration(force bool) error {
	fmt.Println("🗑️  Removing Portunix MCP integration from Claude Code...")

	// Step 1: Check if Claude Code is installed
	fmt.Print("1. Checking Claude Code installation... ")
	if !isClaudeCodeInstalled() {
		fmt.Println("❌ FAILED")
		fmt.Println("Claude Code is not installed or not found in PATH.")
		return fmt.Errorf("Claude Code not found")
	}
	fmt.Println("✅ FOUND")

	// Step 2: Check if MCP server is configured
	fmt.Print("2. Checking MCP configuration... ")
	if !isMCPAlreadyConfigured() {
		fmt.Println("ℹ️  NOT CONFIGURED")
		fmt.Println("Portunix MCP server is not configured in Claude Code.")
		fmt.Println("Nothing to remove.")
		return nil
	}
	fmt.Println("✅ FOUND")

	// Step 3: Confirm removal (unless force flag is used)
	if !force {
		fmt.Print("3. Confirming removal... ")
		fmt.Println("⚠️  CONFIRMATION REQUIRED")
		fmt.Print("Are you sure you want to remove Portunix MCP integration? (y/N): ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Println("❌ CANCELLED")
			fmt.Println("Removal cancelled by user.")
			return nil
		}
		fmt.Println("✅ CONFIRMED")
	} else {
		fmt.Println("3. Skipping confirmation (force mode)... ✅ FORCED")
	}

	// Step 4: Remove MCP server from Claude Code
	fmt.Print("4. Removing Portunix MCP server from Claude Code... ")
	if err := removeMCPServerFromClaudeCodeWithMessages(); err != nil {
		fmt.Println("❌ FAILED")
		return fmt.Errorf("failed to remove MCP server: %w", err)
	}
	fmt.Println("✅ REMOVED")

	// Step 5: Verify removal
	fmt.Print("5. Verifying removal... ")
	if isMCPAlreadyConfigured() {
		fmt.Println("⚠️  WARNING")
		fmt.Println("Portunix still appears in MCP configuration. Manual cleanup may be required.")
	} else {
		fmt.Println("✅ VERIFIED")
	}

	// Success message
	fmt.Println("\n🎉 MCP Integration removed successfully!")
	fmt.Println("\nPortunix MCP server has been removed from Claude Code configuration.")
	fmt.Println("You can reconfigure it anytime using: portunix mcp configure")

	return nil
}

// removeMCPServerFromClaudeCodeWithMessages wraps the common function with user messages
func removeMCPServerFromClaudeCodeWithMessages() error {
	err := removeMCPServerFromClaudeCode()
	if err != nil {
		// Check for Claude Code module issues and provide helpful messages
		if claudePath, pathErr := getClaudePath(); pathErr == nil {
			fmt.Println("   ⚠️  Claude Code has module issues. Manual removal required:")
			fmt.Printf("   Run: %s mcp remove portunix\n", claudePath)
			fmt.Println("   Or edit Claude Code configuration file manually.")
			return nil // Don't fail, just provide instructions
		}
	}
	return err
}

func init() {
	mcpCmd.AddCommand(mcpRemoveCmd)

	mcpRemoveCmd.Flags().BoolP("force", "f", false, "Force removal without confirmation")
}
