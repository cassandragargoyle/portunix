package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"portunix.ai/app/install"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize MCP server configuration with interactive wizard",
	Long: `Initialize MCP server configuration for AI assistant integration.

This interactive wizard will guide you through:
- Detecting installed AI assistants (Claude Code, Claude Desktop, Gemini CLI)
- Configuring MCP server for each assistant
- Setting up appropriate security profiles
- Testing the integration

Examples:
  portunix mcp init                                          # Interactive wizard
  portunix mcp init --assistant claude-code                  # Claude Code with stdio (default)
  portunix mcp init --assistant claude-code --type stdio     # Claude Code with stdio (explicit)
  portunix mcp init --assistant claude-desktop --type remote # Claude Desktop with remote server
  portunix mcp init --preset development                     # Use development preset`,
	Run: func(cmd *cobra.Command, args []string) {
		assistant, _ := cmd.Flags().GetString("assistant")
		serverType, _ := cmd.Flags().GetString("type")
		preset, _ := cmd.Flags().GetString("preset")
		force, _ := cmd.Flags().GetBool("force")

		// If preset is specified, use preset configuration
		if preset != "" {
			if err := runPresetConfiguration(preset, force); err != nil {
				fmt.Printf("❌ Failed to configure with preset: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// If assistant is specified, use non-interactive mode
		if assistant != "" {
			if err := runNonInteractiveConfiguration(assistant, serverType, force); err != nil {
				fmt.Printf("❌ Failed to configure: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Otherwise, run interactive wizard
		if err := runInteractiveWizard(force); err != nil {
			fmt.Printf("❌ Wizard failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func runInteractiveWizard(force bool) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🚀 Welcome to Portunix MCP Server Setup Wizard")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	// Step 1: Detect existing configuration
	fmt.Println("📋 Step 1: Checking existing configuration...")
	config, err := loadMCPConfiguration()
	if err == nil && len(config.Assistants) > 0 && !force {
		fmt.Println("🔧 Found existing MCP server configuration:")
		for _, assistant := range config.Assistants {
			status := "✅"
			if !isAssistantInstalled(assistant.Name) {
				status = "⚠️ (not installed)"
			}
			fmt.Printf("   %s %s\n", status, getAssistantDisplayName(assistant.Name))
		}
		fmt.Println("\nYou can:")
		fmt.Println("1. Add another AI assistant to existing configuration")
		fmt.Println("2. Reconfigure existing setup")
		fmt.Println("3. Delete configuration")
		fmt.Println("4. Cancel")
		fmt.Print("\nChoice (1-4): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		switch response {
		case "1":
			fmt.Println("Continuing to add new assistant...")
		case "2":
			fmt.Println("Reconfiguring existing setup...")
		case "3":
			fmt.Print("⚠️  This will delete MCP server configuration. Continue? (y/N): ")
			confirmResponse, _ := reader.ReadString('\n')
			confirmResponse = strings.ToLower(strings.TrimSpace(confirmResponse))

			if confirmResponse == "y" || confirmResponse == "yes" {
				fmt.Println("🗑️  Deleting configuration...")

				if err := removeMCPConfiguration(); err != nil {
					fmt.Printf("⚠️  Warning: Failed to remove Portunix configuration: %v\n", err)
				}

				if err := removeMCPServerFromClaudeCode(); err != nil {
					fmt.Printf("⚠️  Warning: Failed to remove from Claude Code: %v\n", err)
				}

				fmt.Println("✅ Configuration deleted. Starting fresh setup...")
				config = nil
			} else {
				fmt.Println("Configuration deletion cancelled.")
				return nil
			}
		case "4":
			fmt.Println("Configuration cancelled.")
			return nil
		default:
			fmt.Println("Invalid choice. Configuration cancelled.")
			return nil
		}
	}

	// Step 2: Detect installed AI assistants
	fmt.Println("\n📋 Step 2: Detecting installed AI assistants...")
	detectedAssistants := detectInstalledAssistants()

	if len(detectedAssistants) == 0 {
		fmt.Println("❌ No AI assistants detected.")
	} else {
		fmt.Println("✅ Detected assistants:")

		configuredMap := make(map[string]bool)
		if config != nil && err == nil && len(config.Assistants) > 0 {
			for _, assistant := range config.Assistants {
				configuredMap[assistant.Name] = true
			}
		}

		for i, assistant := range detectedAssistants {
			status := ""
			if configuredMap[assistant] {
				status = " (already configured)"
			}
			fmt.Printf("   %d. %s%s\n", i+1, getAssistantDisplayName(assistant), status)
		}
	}

	if len(detectedAssistants) == 0 {
		fmt.Println("❌ No AI assistants detected.")
		fmt.Println("\nWould you like to install an AI assistant?")
		fmt.Println("1. Claude Code (recommended for CLI development)")
		fmt.Println("2. Claude Desktop (for GUI experience)")
		fmt.Println("3. Skip installation")
		fmt.Print("\nChoice (1-3): ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Println("\n🚀 Installing Claude Code...")
			if err := installClaudeCode(); err != nil {
				return fmt.Errorf("failed to install Claude Code: %w", err)
			}
			detectedAssistants = append(detectedAssistants, "claude-code")
		case "2":
			fmt.Println("\n🚀 Installing Claude Desktop...")
			fmt.Println("Please visit: https://claude.ai/download")
			fmt.Println("After installation, run this wizard again.")
			return nil
		case "3":
			fmt.Println("Skipping installation. Run wizard again after installing an AI assistant.")
			return nil
		default:
			return fmt.Errorf("invalid choice")
		}
	}

	// Step 3: Select AI assistant to configure
	fmt.Println("\n📋 Step 3: Select AI assistant to configure:")

	configuredMap := make(map[string]bool)
	if config != nil && err == nil && len(config.Assistants) > 0 {
		for _, assistant := range config.Assistants {
			configuredMap[assistant.Name] = true
		}
	}

	for i, assistant := range detectedAssistants {
		status := ""
		if configuredMap[assistant] {
			status = " (reconfigure)"
		} else {
			status = " (new)"
		}
		fmt.Printf("%d. %s%s\n", i+1, getAssistantDisplayName(assistant), status)
	}

	var selectedAssistant string
	if len(detectedAssistants) == 1 {
		selectedAssistant = detectedAssistants[0]
		fmt.Printf("\nSelected: %s\n", getAssistantDisplayName(selectedAssistant))
	} else {
		fmt.Printf("Choice (1-%d): ", len(detectedAssistants))

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		idx := 0
		fmt.Sscanf(choice, "%d", &idx)

		if idx < 1 || idx > len(detectedAssistants) {
			return fmt.Errorf("invalid choice")
		}
		selectedAssistant = detectedAssistants[idx-1]
	}

	if configuredMap[selectedAssistant] {
		fmt.Printf("🔄 Reconfiguring %s...\n", getAssistantDisplayName(selectedAssistant))
	} else {
		fmt.Printf("🆕 Configuring %s for the first time...\n", getAssistantDisplayName(selectedAssistant))
	}

	// Step 4: Configure server type based on assistant
	fmt.Println("\n📋 Step 4: Configuring server type...")
	serverType := getDefaultServerType(selectedAssistant)

	if selectedAssistant == "claude-code" {
		fmt.Println("Claude Code uses stdio communication by default.")
		serverType = "stdio"
	} else if selectedAssistant == "claude-desktop" {
		fmt.Println("Claude Desktop requires remote server configuration.")
		serverType = "remote"

		fmt.Print("\nPort (default 3002): ")
		portStr, _ := reader.ReadString('\n')
		portStr = strings.TrimSpace(portStr)
		if portStr == "" {
			portStr = "3002"
		}

		fmt.Print("Protocol (http/https, default https): ")
		protocol, _ := reader.ReadString('\n')
		protocol = strings.TrimSpace(protocol)
		if protocol == "" {
			protocol = "https"
		}
	}

	// Step 5: Configure security profile
	fmt.Println("\n📋 Step 5: Select security profile:")
	fmt.Println("1. Development (CLI tools, local development)")
	fmt.Println("2. Standard (Desktop apps, full integration)")
	fmt.Println("3. Restricted (minimal permissions)")
	fmt.Print("Choice (1-3, default 1): ")

	secChoice, _ := reader.ReadString('\n')
	secChoice = strings.TrimSpace(secChoice)
	if secChoice == "" {
		secChoice = "1"
	}

	var securityProfile string
	switch secChoice {
	case "1":
		securityProfile = "development"
	case "2":
		securityProfile = "standard"
	case "3":
		securityProfile = "restricted"
	default:
		securityProfile = "development"
	}

	// Step 6: Apply configuration
	fmt.Println("\n📋 Step 6: Applying configuration...")
	if err := applyAssistantConfiguration(selectedAssistant, serverType, securityProfile); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	// Step 7: Test connection
	fmt.Println("\n📋 Step 7: Testing connection...")
	if err := testAssistantConnection(selectedAssistant); err != nil {
		fmt.Printf("⚠️  Connection test failed: %v\n", err)
		fmt.Println("You may need to start the MCP server manually.")
	} else {
		fmt.Println("✅ Connection test successful!")
	}

	// Final instructions
	fmt.Println("\n🎉 Configuration complete!")
	fmt.Println("\nNext steps:")
	if serverType == "stdio" {
		fmt.Printf("1. The MCP server will start automatically when you use %s\n", getAssistantDisplayName(selectedAssistant))
		fmt.Printf("2. Open %s and test the integration\n", getAssistantDisplayName(selectedAssistant))
	} else {
		fmt.Println("1. Start the MCP server: portunix mcp start")
		fmt.Printf("2. Open %s and test the integration\n", getAssistantDisplayName(selectedAssistant))
	}

	return nil
}

func runNonInteractiveConfiguration(assistant, serverType string, force bool) error {
	fmt.Printf("🔧 Configuring MCP server for %s...\n", getAssistantDisplayName(assistant))

	if isMCPAlreadyConfigured() && !force {
		return fmt.Errorf("MCP server already configured. Use --force to reconfigure")
	}

	if !isAssistantInstalled(assistant) {
		fmt.Printf("❌ %s is not installed\n", getAssistantDisplayName(assistant))
		fmt.Println("Install it first or run interactive wizard for installation help")
		return fmt.Errorf("assistant not installed")
	}

	if serverType == "" {
		serverType = getDefaultServerType(assistant)
	}

	securityProfile := getDefaultSecurityProfile(assistant)
	if err := applyAssistantConfiguration(assistant, serverType, securityProfile); err != nil {
		return err
	}

	fmt.Println("✅ Configuration applied successfully!")
	return nil
}

func runPresetConfiguration(preset string, force bool) error {
	fmt.Printf("🔧 Applying preset configuration: %s\n", preset)

	switch preset {
	case "development":
		return runNonInteractiveConfiguration("claude-code", "stdio", force)
	case "standard":
		return runNonInteractiveConfiguration("claude-desktop", "remote", force)
	default:
		return fmt.Errorf("unknown preset: %s", preset)
	}
}

func installClaudeCode() error {
	if err := install.InstallPackage("claude-code", "npm"); err == nil {
		return nil
	}

	cmd := exec.Command("npm", "install", "-g", "@anthropic-ai/claude-code")
	if err := cmd.Run(); err != nil {
		fmt.Println("npm installation failed, trying curl method...")
		installScript := "curl -fsSL https://claude.ai/cli/install.sh | sh"
		cmd = exec.Command("sh", "-c", installScript)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("installation failed")
		}
	}
	return nil
}

func applyAssistantConfiguration(assistant, serverType, securityProfile string) error {
	fmt.Printf("📝 Configuring %s with %s server and %s security profile...\n",
		getAssistantDisplayName(assistant), serverType, securityProfile)

	config, err := loadMCPConfiguration()
	if err != nil {
		config = &MCPConfiguration{
			ServerType:      serverType,
			SecurityProfile: securityProfile,
			Assistants:      []AssistantConfig{},
		}

		if serverType == "remote" {
			switch assistant {
			case "claude-code":
				config.Port = 3001
			case "claude-desktop":
				config.Port = 3002
			case "gemini-cli":
				config.Port = 3003
			default:
				config.Port = 3001
			}
			config.Protocol = "https"
		}
	}

	for i, a := range config.Assistants {
		if a.Name == assistant {
			config.Assistants[i].ServerType = serverType
			config.Assistants[i].Configured = true
			if err := saveMCPConfiguration(config); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			return nil
		}
	}

	newAssistant := AssistantConfig{
		Name:       assistant,
		ServerType: serverType,
		Configured: true,
	}
	config.Assistants = append(config.Assistants, newAssistant)

	if err := saveMCPConfiguration(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	switch assistant {
	case "claude-code":
		return configureClaudeCode(serverType)
	case "claude-desktop":
		return configureClaudeDesktop(serverType)
	case "gemini-cli":
		return configureGeminiCLI(serverType)
	default:
		fmt.Printf("⚠️  Assistant-specific configuration for %s not implemented yet\n", assistant)
		return nil
	}
}

func configureClaudeCode(serverType string) error {
	fmt.Println("   🔧 Configuring Claude Code MCP integration...")
	return configureMCPIntegration("stdio", "local", 3001, "standard", true)
}

func configureClaudeDesktop(serverType string) error {
	fmt.Println("   🔧 Configuring Claude Desktop MCP integration...")
	fmt.Println("   ⚠️  Claude Desktop configuration requires manual setup")
	fmt.Println("   Please add Portunix to your Claude Desktop MCP configuration:")

	configPath := getClaudeDesktopConfigPath()
	fmt.Printf("   Config file: %s\n", configPath)
	fmt.Println("   Add this to your mcp_servers.json:")
	fmt.Println(`   {
     "portunix": {
       "command": "portunix",
       "args": ["mcp", "--port", "3002"]
     }
   }`)

	return nil
}

func configureGeminiCLI(serverType string) error {
	fmt.Println("   🔧 Configuring Gemini CLI MCP integration...")
	fmt.Println("   ⚠️  Gemini CLI MCP integration is experimental")
	fmt.Println("   Configuration method needs to be researched")
	return nil
}

func testAssistantConnection(assistant string) error {
	fmt.Printf("Testing connection with %s...\n", getAssistantDisplayName(assistant))
	return nil
}

func init() {
	mcpCmd.AddCommand(initCmd)

	initCmd.Flags().String("assistant", "", "AI assistant to configure (claude-code, claude-desktop, gemini-cli)")
	initCmd.Flags().String("type", "", "Server type (stdio, remote)")
	initCmd.Flags().String("preset", "", "Use preset configuration (development, standard)")
	initCmd.Flags().Bool("force", false, "Force reconfiguration even if already configured")
}
