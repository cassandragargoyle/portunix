package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var mcpServerConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage MCP server configuration",
	Long: `View and manage MCP server configuration.

This command allows you to:
- View current configuration
- Edit existing configuration
- Add new AI assistants
- Force reconfiguration

Examples:
  portunix mcp-server config                    # Show current configuration
  portunix mcp-server config --edit            # Interactive configuration editing
  portunix mcp-server config --force           # Force reconfiguration
  portunix mcp-server config --assistant claude-desktop --add  # Add new assistant`,
	Run: func(cmd *cobra.Command, args []string) {
		edit, _ := cmd.Flags().GetBool("edit")
		force, _ := cmd.Flags().GetBool("force")
		add, _ := cmd.Flags().GetBool("add")
		assistant, _ := cmd.Flags().GetString("assistant")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		
		// Handle add assistant
		if add && assistant != "" {
			if err := addAssistantConfiguration(assistant); err != nil {
				fmt.Printf("❌ Failed to add assistant: %v\n", err)
				os.Exit(1)
			}
			return
		}
		
		// Handle force reconfiguration
		if force {
			if err := forceReconfiguration(); err != nil {
				fmt.Printf("❌ Failed to reconfigure: %v\n", err)
				os.Exit(1)
			}
			return
		}
		
		// Handle edit mode
		if edit {
			if err := editConfiguration(); err != nil {
				fmt.Printf("❌ Failed to edit configuration: %v\n", err)
				os.Exit(1)
			}
			return
		}
		
		// Default: show configuration
		if jsonOutput {
			showJSONConfiguration()
		} else {
			showTextConfiguration()
		}
	},
}

func showTextConfiguration() {
	fmt.Println("📋 MCP Server Configuration")
	fmt.Println("=" + "========================")
	
	// Check if configuration exists
	if !isMCPConfigurationExists() {
		fmt.Println("\n❌ No configuration found")
		fmt.Println("\nRun 'portunix mcp-server init' to create initial configuration")
		return
	}
	
	// Load and display configuration
	config, err := loadMCPConfiguration()
	if err != nil {
		fmt.Printf("\n❌ Failed to load configuration: %v\n", err)
		return
	}
	
	fmt.Println("\n🔧 Server Settings:")
	fmt.Printf("   Type: %s\n", config.ServerType)
	if config.ServerType == "remote" {
		fmt.Printf("   Port: %d\n", config.Port)
		fmt.Printf("   Protocol: %s\n", config.Protocol)
	}
	fmt.Printf("   Security Profile: %s\n", config.SecurityProfile)
	
	fmt.Println("\n🤖 Configured Assistants:")
	if len(config.Assistants) == 0 {
		fmt.Println("   None configured")
	} else {
		for _, assistant := range config.Assistants {
			status := "✅"
			if !isAssistantInstalled(assistant.Name) {
				status = "⚠️"
			}
			fmt.Printf("   %s %s (%s)\n", status, getAssistantDisplayName(assistant.Name), assistant.ServerType)
		}
	}
	
	fmt.Println("\n🔒 Security Settings:")
	fmt.Printf("   Profile: %s\n", config.SecurityProfile)
	switch config.SecurityProfile {
	case "development":
		fmt.Println("   - Read/analyze operations allowed")
		fmt.Println("   - Package installation (with confirmation)")
		fmt.Println("   - Container management")
		fmt.Println("   - Project detection")
	case "standard":
		fmt.Println("   - All safe operations allowed")
		fmt.Println("   - Environment setup")
		fmt.Println("   - Project management")
	case "restricted":
		fmt.Println("   - Minimal permissions")
		fmt.Println("   - Read-only operations")
	}
	
	fmt.Println("\n📝 Management Commands:")
	fmt.Println("   portunix mcp-server config --edit     # Edit configuration")
	fmt.Println("   portunix mcp-server config --force    # Force reconfiguration")
	fmt.Println("   portunix mcp-server config --add      # Add new assistant")
}

func showJSONConfiguration() {
	config, err := loadMCPConfiguration()
	if err != nil {
		fmt.Printf(`{"error": "%v"}`, err)
		return
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf(`{"error": "%v"}`, err)
		return
	}
	
	fmt.Println(string(data))
}

func editConfiguration() error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("📝 Edit MCP Server Configuration")
	fmt.Println("================================")
	
	// Load existing configuration
	config, err := loadMCPConfiguration()
	if err != nil {
		// Create new configuration if none exists
		config = &MCPConfiguration{
			Assistants: []AssistantConfig{},
		}
	}
	
	// Edit server type
	fmt.Printf("\n1. Server Type (current: %s)\n", config.ServerType)
	fmt.Println("   Options: stdio, remote")
	fmt.Print("   New value (press Enter to keep current): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		config.ServerType = input
	}
	
	// Edit port if remote
	if config.ServerType == "remote" {
		fmt.Printf("\n2. Port (current: %d)\n", config.Port)
		fmt.Print("   New value (press Enter to keep current): ")
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			fmt.Sscanf(input, "%d", &config.Port)
		}
		
		fmt.Printf("\n3. Protocol (current: %s)\n", config.Protocol)
		fmt.Println("   Options: http, https, ws, wss")
		fmt.Print("   New value (press Enter to keep current): ")
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			config.Protocol = input
		}
	}
	
	// Edit security profile
	fmt.Printf("\n4. Security Profile (current: %s)\n", config.SecurityProfile)
	fmt.Println("   Options: development, standard, restricted")
	fmt.Print("   New value (press Enter to keep current): ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		config.SecurityProfile = input
	}
	
	// Save configuration
	if err := saveMCPConfiguration(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	
	fmt.Println("\n✅ Configuration updated successfully!")
	return nil
}

func forceReconfiguration() error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("⚠️  Warning: This will overwrite existing MCP server configuration")
	
	// Show current configuration
	if config, err := loadMCPConfiguration(); err == nil && len(config.Assistants) > 0 {
		fmt.Println("\nCurrent assistants:")
		for _, assistant := range config.Assistants {
			fmt.Printf("  - %s\n", getAssistantDisplayName(assistant.Name))
		}
	}
	
	fmt.Print("\nContinue? (y/N): ")
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	
	if response != "y" && response != "yes" {
		fmt.Println("Reconfiguration cancelled.")
		return nil
	}
	
	// Remove existing configuration
	if err := removeMCPConfiguration(); err != nil {
		fmt.Printf("⚠️  Warning: Failed to remove old configuration: %v\n", err)
	}
	
	// Run interactive wizard
	fmt.Println("\nStarting configuration wizard...")
	return runInteractiveWizard(true)
}

func addAssistantConfiguration(assistant string) error {
	fmt.Printf("➕ Adding %s to MCP configuration...\n", getAssistantDisplayName(assistant))
	
	// Check if assistant is installed
	if !isAssistantInstalled(assistant) {
		fmt.Printf("❌ %s is not installed\n", getAssistantDisplayName(assistant))
		fmt.Println("Install it first with 'portunix mcp-server init'")
		return fmt.Errorf("assistant not installed")
	}
	
	// Load existing configuration
	config, err := loadMCPConfiguration()
	if err != nil {
		config = &MCPConfiguration{
			Assistants: []AssistantConfig{},
		}
	}
	
	// Check if already configured
	for _, a := range config.Assistants {
		if a.Name == assistant {
			fmt.Printf("⚠️  %s is already configured\n", getAssistantDisplayName(assistant))
			return nil
		}
	}
	
	// Add new assistant
	newAssistant := AssistantConfig{
		Name:       assistant,
		ServerType: getDefaultServerType(assistant),
		Configured: true,
	}
	
	config.Assistants = append(config.Assistants, newAssistant)
	
	// Save configuration
	if err := saveMCPConfiguration(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	
	fmt.Printf("✅ %s added successfully!\n", getAssistantDisplayName(assistant))
	return nil
}

// Configuration structure
type MCPConfiguration struct {
	ServerType      string             `json:"server_type"`
	Port            int                `json:"port,omitempty"`
	Protocol        string             `json:"protocol,omitempty"`
	SecurityProfile string             `json:"security_profile"`
	Assistants      []AssistantConfig  `json:"assistants"`
}

type AssistantConfig struct {
	Name       string `json:"name"`
	ServerType string `json:"server_type"`
	Configured bool   `json:"configured"`
}

func getMCPConfigFile() string {
	// Use Portunix config directory
	configDir := filepath.Join(os.Getenv("HOME"), ".portunix")
	os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "mcp-server.json")
}

func isMCPConfigurationExists() bool {
	_, err := os.Stat(getMCPConfigFile())
	return err == nil
}

func loadMCPConfiguration() (*MCPConfiguration, error) {
	configFile := getMCPConfigFile()
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	
	var config MCPConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	// Set defaults if empty
	if config.ServerType == "" {
		config.ServerType = "stdio"
	}
	if config.SecurityProfile == "" {
		config.SecurityProfile = "development"
	}
	if config.Port == 0 && config.ServerType == "remote" {
		config.Port = 3001
	}
	
	return &config, nil
}

func saveMCPConfiguration(config *MCPConfiguration) error {
	configFile := getMCPConfigFile()
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(configFile, data, 0644)
}

func removeMCPConfiguration() error {
	configFile := getMCPConfigFile()
	return os.Remove(configFile)
}

func init() {
	mcpCmd.AddCommand(mcpServerConfigCmd)
	
	mcpServerConfigCmd.Flags().Bool("edit", false, "Interactive configuration editing")
	mcpServerConfigCmd.Flags().Bool("force", false, "Force reconfiguration")
	mcpServerConfigCmd.Flags().Bool("add", false, "Add new assistant to configuration")
	mcpServerConfigCmd.Flags().String("assistant", "", "Assistant to add (claude-code, claude-desktop, gemini-cli)")
	mcpServerConfigCmd.Flags().Bool("json", false, "Output in JSON format")
}