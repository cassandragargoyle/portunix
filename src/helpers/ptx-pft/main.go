package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var version = "dev"

// Voice directory names for QFD (PascalCase) and basic (lowercase) templates
var voiceNames = map[string][]string{
	"voc": {"VoC", "voc"},
	"vos": {"VoS", "vos"},
	"vob": {"VoB", "vob"},
	"voe": {"VoE", "voe"},
}

// getVoiceDir returns the path to a voice directory, checking both QFD (PascalCase)
// and basic (lowercase) variants. Returns the first existing path or lowercase fallback.
func getVoiceDir(projectDir, voice string) string {
	voice = strings.ToLower(voice)
	variants, ok := voiceNames[voice]
	if !ok {
		// Unknown voice, return as-is
		return filepath.Join(projectDir, voice)
	}

	// Try each variant in order (QFD first, then lowercase)
	for _, variant := range variants {
		path := filepath.Join(projectDir, variant)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Fallback to QFD (PascalCase) for new directories
	return filepath.Join(projectDir, variants[0])
}

// rootCmd represents the base command for ptx-pft
var rootCmd = &cobra.Command{
	Use:   "ptx-pft",
	Short: "Portunix Product Feedback Tool Helper",
	Long: `ptx-pft is a helper binary for Portunix that manages integration with
external Product Feedback Tools (Fider.io, Canny, ProductBoard, etc.).

It provides bidirectional synchronization between local project documentation
(markdown files) and external feedback systems.

This binary is typically invoked by the main portunix dispatcher and should not be used directly.

Supported features:
- Configure product feedback integration
- Deploy feedback tool infrastructure (Fider.io)
- Bidirectional synchronization of feedback items
- Link feedback to local documentation
- Generate feedback reports`,
	Version:            version,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		handleCommand(args)
	},
}

func handleCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("No command specified")
		return
	}

	command := args[0]
	subArgs := args[1:]

	switch command {
	case "pft":
		if len(subArgs) == 0 {
			showPFTHelp()
		} else {
			handlePFTCommand(subArgs)
		}
	case "--version":
		fmt.Printf("ptx-pft version %s\n", version)
	case "--description":
		fmt.Println("Portunix Product Feedback Tool Helper")
	case "--list-commands":
		fmt.Println("pft")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Supported commands: pft")
	}
}

func showPFTHelp() {
	fmt.Println("Usage: portunix pft [subcommand]")
	fmt.Println()
	fmt.Println("Product Feedback Tool Commands:")
	fmt.Println()
	fmt.Println("Project Management:")
	fmt.Println("  project create <name>    - Create new PFT project (default: qfd template)")
	fmt.Println("  project create <name> --template <tpl>")
	fmt.Println("                           - Create project with specific template (qfd, basic)")
	fmt.Println("  info                     - Show methodology documentation")
	fmt.Println("  info --json              - Output as JSON (for MCP integration)")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  configure                              - Interactive configuration wizard")
	fmt.Println("  configure --name <name> --path <path>  - Set global settings")
	fmt.Println("  configure --area <voc|vos|vob|voe> ... - Configure per-area provider")
	fmt.Println("  configure --smtp-host <host> ...       - Configure SMTP server")
	fmt.Println("  configure --show                       - Show current configuration")
	fmt.Println()
	fmt.Println("Infrastructure:")
	fmt.Println("  deploy                   - Deploy feedback tool to container")
	fmt.Println("  status                   - Check feedback tool status")
	fmt.Println("  destroy                  - Remove feedback tool instance")
	fmt.Println()
	fmt.Println("Synchronization:")
	fmt.Println("  sync                     - Full bidirectional sync")
	fmt.Println("  pull                     - Pull from external system")
	fmt.Println("  push                     - Push to external system")
	fmt.Println()
	fmt.Println("User/Customer Registry:")
	fmt.Println("  user list                - List all users")
	fmt.Println("  user add                 - Add new user")
	fmt.Println("  user role <id>           - Assign role to user")
	fmt.Println("  user link <id>           - Link user to external ID")
	fmt.Println("  user remove <id>         - Remove user")
	fmt.Println("  role list                - List available roles")
	fmt.Println("  role init                - Initialize default role files")
	fmt.Println()
	fmt.Println("Feedback Management:")
	fmt.Println("  list                     - List all feedback items")
	fmt.Println("  add                      - Add new feedback item")
	fmt.Println("  show <id>                - Show feedback details")
	fmt.Println("  link <id> <issue>        - Link feedback to local issue")
	fmt.Println()
	fmt.Println("Category Management:")
	fmt.Println("  category list            - List categories in area")
	fmt.Println("  category add <id>        - Create new category")
	fmt.Println("  category remove <id>     - Delete category")
	fmt.Println("  category rename <id>     - Rename category")
	fmt.Println("  category show <id>       - Show category details")
	fmt.Println()
	fmt.Println("Item Categorization:")
	fmt.Println("  assign <item-id> --category <cat-id>")
	fmt.Println("                           - Add category to item")
	fmt.Println("  unassign <item-id> --category <cat-id>")
	fmt.Println("                           - Remove category from item")
	fmt.Println("  unassign <item-id> --all - Remove all categories")
	fmt.Println()
	fmt.Println("Reporting:")
	fmt.Println("  report                   - Generate feedback report")
	fmt.Println("  export --format=md       - Export to markdown")
	fmt.Println()
	fmt.Println("Notifications:")
	fmt.Println("  notify <id> --user <email> --type <type>")
	fmt.Println("                           - Send notification to user")
	fmt.Println("  notify <id> --all-voc --type <type>")
	fmt.Println("                           - Notify all VoC users")
	fmt.Println("  notify <id> --all-vos --type <type>")
	fmt.Println("                           - Notify all VoS users")
	fmt.Println()
	fmt.Println("Available providers: " + strings.Join(ListProviders(), ", "))
	if len(ListProviders()) == 0 {
		fmt.Println("  (no providers registered yet - Phase 3)")
	}
}

func handlePFTCommand(args []string) {
	if len(args) == 0 {
		showPFTHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "project":
		handleProjectCommand(subArgs)
	case "info":
		handleInfoCommand(subArgs)
	case "example":
		handleExampleCommand(subArgs)
	case "configure":
		handleConfigureCommand(subArgs)
	case "deploy":
		handleDeployCommand(subArgs)
	case "status":
		handleStatusCommand(subArgs)
	case "destroy":
		handleDestroyCommand(subArgs)
	case "sync":
		handleSyncCommand(subArgs)
	case "pull":
		handlePullCommand(subArgs)
	case "push":
		handlePushCommand(subArgs)
	case "list":
		handleListCommand(subArgs)
	case "show":
		handleShowCommand(subArgs)
	case "add":
		handleAddCommand(subArgs)
	case "update":
		handleUpdateCommand(subArgs)
	case "link":
		handleLinkCommand(subArgs)
	case "report":
		handleReportCommand(subArgs)
	case "export":
		handleExportCommand(subArgs)
	case "cache":
		handleCacheCommand(subArgs)
	case "notify":
		handleNotifyCommand(subArgs)
	case "user":
		handleUserCommand(subArgs)
	case "role":
		handleRoleListCommand(subArgs)
	case "category":
		handleCategoryCommand(subArgs)
	case "assign":
		handleAssignCommand(subArgs)
	case "unassign":
		handleUnassignCommand(subArgs)
	case "--help", "-h":
		showPFTHelp()
	default:
		fmt.Printf("Unknown pft subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix pft --help' for available commands")
	}
}

// Configure command handlers
func handleConfigureCommand(args []string) {
	// Parse flags
	var name, path, area, provider, url, token, projectID string
	var smtpHost, smtpUser, smtpPass, smtpFrom string
	var smtpPort int
	var showConfig, fixPaths bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--fix-paths":
			fixPaths = true
		case "--name":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		case "--path":
			if i+1 < len(args) {
				path = args[i+1]
				i++
			}
		case "--area":
			if i+1 < len(args) {
				area = args[i+1]
				i++
			}
		case "--provider":
			if i+1 < len(args) {
				provider = args[i+1]
				i++
			}
		case "--url":
			if i+1 < len(args) {
				url = args[i+1]
				i++
			}
		case "--token":
			if i+1 < len(args) {
				token = args[i+1]
				i++
			}
		case "--project-id":
			if i+1 < len(args) {
				projectID = args[i+1]
				i++
			}
		case "--smtp-host":
			if i+1 < len(args) {
				smtpHost = args[i+1]
				i++
			}
		case "--smtp-port":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &smtpPort)
				i++
			}
		case "--smtp-user":
			if i+1 < len(args) {
				smtpUser = args[i+1]
				i++
			}
		case "--smtp-pass":
			if i+1 < len(args) {
				smtpPass = args[i+1]
				i++
			}
		case "--smtp-from":
			if i+1 < len(args) {
				smtpFrom = args[i+1]
				i++
			}
		case "--show":
			showConfig = true
		case "--help", "-h":
			showConfigureHelp()
			return
		}
	}

	// Show current configuration
	if showConfig {
		showCurrentConfig(path)
		return
	}

	// Fix absolute paths to relative for cross-platform compatibility
	if fixPaths {
		fixConfigPaths(path)
		return
	}

	// SMTP configuration
	if smtpHost != "" || smtpPort > 0 || smtpUser != "" || smtpFrom != "" {
		updateSMTPConfig(path, smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)
		return
	}

	// Per-area configuration
	if area != "" {
		updateAreaConfig(path, area, provider, url, token, projectID)
		return
	}

	// Global configuration (name, path)
	if name != "" || path != "" {
		updateGlobalConfig(name, path)
		return
	}

	// No parameters - run interactive wizard
	runConfigureWizard()
}

func showConfigureHelp() {
	fmt.Println("Usage: portunix pft configure [options]")
	fmt.Println()
	fmt.Println("Global options:")
	fmt.Println("  --name <name>         Set product name")
	fmt.Println("  --path <path>         Set path to local documents")
	fmt.Println("  --show                Show current configuration")
	fmt.Println("  --fix-paths           Convert absolute paths to relative for cross-platform use")
	fmt.Println()
	fmt.Println("Per-area options (requires --area):")
	fmt.Println("  --area <area>         Target area (voc, vos, vob, voe)")
	fmt.Println("  --provider <type>     Set provider (fider, clearflask, eververse, local)")
	fmt.Println("  --url <url>           Set provider endpoint URL")
	fmt.Println("  --token <token>       Set API token")
	fmt.Println("  --project-id <id>     Set project ID (for ClearFlask)")
	fmt.Println()
	fmt.Println("SMTP options:")
	fmt.Println("  --smtp-host <host>    SMTP server hostname")
	fmt.Println("  --smtp-port <port>    SMTP server port (default: 587)")
	fmt.Println("  --smtp-user <user>    SMTP username")
	fmt.Println("  --smtp-pass <pass>    SMTP password")
	fmt.Println("  --smtp-from <email>   Sender email address")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft configure --name 'MyProduct' --path /tmp/pft")
	fmt.Println("  portunix pft configure --area voc --provider fider --url http://localhost:3100")
	fmt.Println("  portunix pft configure --smtp-host smtp.example.com --smtp-port 587")
	fmt.Println()
	fmt.Println("Without options, runs an interactive configuration wizard.")
}

func showCurrentConfig(configPath string) {
	config, _, err := loadOrCreateConfig(configPath)
	if err != nil {
		fmt.Printf("No configuration found: %v\n", err)
		fmt.Println("Run 'portunix pft configure' to create one.")
		return
	}

	fmt.Println("Current configuration:")
	fmt.Println()
	fmt.Printf("  Product Name: %s\n", config.Name)
	fmt.Printf("  Document Path: %s\n", config.Path)
	fmt.Println()

	// Show per-area configuration
	areas := []struct {
		name string
		cfg  *AreaConfig
	}{
		{"VoC (Voice of Customer)", config.VoC},
		{"VoS (Voice of Stakeholder)", config.VoS},
		{"VoB (Voice of Business)", config.VoB},
		{"VoE (Voice of Engineer)", config.VoE},
	}

	for _, area := range areas {
		if area.cfg != nil && area.cfg.Provider != "" {
			fmt.Printf("  %s:\n", area.name)
			fmt.Printf("    Provider: %s\n", area.cfg.Provider)
			if area.cfg.URL != "" {
				fmt.Printf("    URL: %s\n", area.cfg.URL)
			}
			if area.cfg.APIToken != "" {
				fmt.Printf("    API Token: %s***\n", area.cfg.APIToken[:min(4, len(area.cfg.APIToken))])
			}
			if area.cfg.ProjectID != "" {
				fmt.Printf("    Project ID: %s\n", area.cfg.ProjectID)
			}
		} else {
			fmt.Printf("  %s: local (no external sync)\n", area.name)
		}
	}

	// Show SMTP configuration
	if config.SMTP != nil && config.SMTP.Host != "" {
		fmt.Println()
		fmt.Println("  SMTP:")
		fmt.Printf("    Host: %s:%d\n", config.SMTP.Host, config.SMTP.Port)
		fmt.Printf("    From: %s\n", config.SMTP.From)
		if config.SMTP.Username != "" {
			fmt.Printf("    Username: %s\n", config.SMTP.Username)
		}
	}

	fmt.Println()
	fmt.Println("Sync settings:")
	fmt.Printf("  Auto sync: %v\n", config.Sync.Auto)
	fmt.Printf("  Interval: %s\n", config.Sync.Interval)
	fmt.Printf("  Conflict resolution: %s\n", config.Sync.ConflictResolution)
}

// updateGlobalConfig updates global settings (name, path)
func updateGlobalConfig(name, path string) {
	config, _, err := loadOrCreateConfig(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if path != "" {
		absPath, _ := filepath.Abs(path)
		config.Path = absPath
		fmt.Printf("Document path set to: %s\n", absPath)
	}

	if name != "" {
		config.Name = name
		fmt.Printf("Product name set to: %s\n", name)
	}

	saveConfig(config)
}

// updateAreaConfig updates configuration for a specific area
func updateAreaConfig(configPath, area, provider, url, token, projectID string) {
	// Validate area
	if !IsValidArea(area) {
		fmt.Printf("Invalid area '%s'. Valid options: voc, vos, vob, voe\n", area)
		return
	}

	// Validate provider if specified
	if provider != "" {
		validProviders := []string{"fider", "clearflask", "eververse", "local"}
		isValid := false
		for _, p := range validProviders {
			if provider == p {
				isValid = true
				break
			}
		}
		if !isValid {
			fmt.Printf("Invalid provider '%s'. Valid options: fider, clearflask, eververse, local\n", provider)
			return
		}
	}

	config, _, err := loadOrCreateConfig(configPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Get or create area config
	areaCfg := config.GetAreaConfig(area)
	if areaCfg == nil {
		areaCfg = &AreaConfig{}
	}

	// Update fields
	if provider != "" {
		areaCfg.Provider = provider
		fmt.Printf("Area %s provider set to: %s\n", area, provider)
	}
	if url != "" {
		areaCfg.URL = url
		fmt.Printf("Area %s URL set to: %s\n", area, url)
	}
	if token != "" {
		areaCfg.APIToken = token
		fmt.Printf("Area %s API token updated\n", area)
	}
	if projectID != "" {
		areaCfg.ProjectID = projectID
		fmt.Printf("Area %s project ID set to: %s\n", area, projectID)
	}

	// Set the area config
	config.SetAreaConfig(area, areaCfg)

	saveConfig(config)
}

// updateSMTPConfig updates SMTP server configuration
func updateSMTPConfig(configPath, host string, port int, user, pass, from string) {
	config, _, err := loadOrCreateConfig(configPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if config.SMTP == nil {
		config.SMTP = &SMTPConfig{}
	}

	if host != "" {
		config.SMTP.Host = host
		fmt.Printf("SMTP host set to: %s\n", host)
	}
	if port > 0 {
		config.SMTP.Port = port
		fmt.Printf("SMTP port set to: %d\n", port)
	}
	if user != "" {
		config.SMTP.Username = user
		fmt.Printf("SMTP username set to: %s\n", user)
	}
	if pass != "" {
		config.SMTP.Password = pass
		fmt.Println("SMTP password updated")
	}
	if from != "" {
		config.SMTP.From = from
		fmt.Printf("SMTP from address set to: %s\n", from)
	}

	saveConfig(config)
}

// loadOrCreateConfig loads existing config or creates a new one
// Returns: config, configFilePath (path to .pft-config.json), error
func loadOrCreateConfig(path string) (*Config, string, error) {
	var config *Config
	var configFilePath string

	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, "", fmt.Errorf("error resolving path: %w", err)
		}
		configFilePath = filepath.Join(absPath, ConfigFileName)
		if _, statErr := os.Stat(configFilePath); statErr == nil {
			config, _ = LoadConfigFromPath(configFilePath)
		}
		if config == nil {
			config = NewDefaultConfig()
			config.Path = absPath
		}
	} else {
		// Try to find config in current or parent directories
		foundPath, err := findConfigFile()
		if err == nil {
			configFilePath = foundPath
			config, _ = LoadConfigFromPath(configFilePath)
		}
		if config == nil {
			config = NewDefaultConfig()
			// Set configFilePath to current directory for new configs
			cwd, _ := os.Getwd()
			configFilePath = filepath.Join(cwd, ConfigFileName)
		}
	}

	return config, configFilePath, nil
}

// saveConfig saves configuration to the appropriate path
func saveConfig(config *Config) {
	savePath := config.Path
	if savePath == "" {
		savePath, _ = os.Getwd()
	}
	if err := config.Save(savePath); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
		return
	}
	fmt.Printf("\nConfiguration saved to %s\n", GetConfigPath(savePath))
}

// fixConfigPaths converts absolute paths to empty/relative for cross-platform compatibility
func fixConfigPaths(configPath string) {
	config, configFilePath, err := loadOrCreateConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		return
	}

	if config.Path == "" {
		fmt.Println("Path is already empty (cross-platform compatible)")
		return
	}

	configDir := filepath.Dir(configFilePath)
	oldPath := config.Path

	// Check if path is absolute
	if !filepath.IsAbs(config.Path) {
		fmt.Printf("Path is already relative: %s\n", config.Path)
		return
	}

	// Check if the absolute path points to the config directory
	absConfigDir, _ := filepath.Abs(configDir)
	if config.Path == absConfigDir {
		// Path points to config directory - set to empty
		config.Path = ""
		fmt.Printf("Converting absolute path to empty (same as config directory):\n")
		fmt.Printf("  Old: %s\n", oldPath)
		fmt.Printf("  New: (empty - uses config file directory)\n")
	} else {
		// Try to make it relative to config directory
		relPath, err := filepath.Rel(configDir, config.Path)
		if err != nil {
			fmt.Printf("Cannot convert to relative path: %v\n", err)
			fmt.Println("The path may be on a different drive or volume.")
			return
		}

		// Check if relative path would go too many levels up
		if strings.HasPrefix(relPath, ".."+string(filepath.Separator)+".."+string(filepath.Separator)+"..") {
			fmt.Printf("Path is too far from config directory, keeping as empty:\n")
			fmt.Printf("  Old: %s\n", oldPath)
			config.Path = ""
			fmt.Printf("  New: (empty - uses config file directory)\n")
		} else {
			config.Path = relPath
			fmt.Printf("Converting absolute path to relative:\n")
			fmt.Printf("  Old: %s\n", oldPath)
			fmt.Printf("  New: %s\n", relPath)
		}
	}

	// Save the updated config
	if err := config.SaveToPath(configFilePath); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
		return
	}

	fmt.Println("\nConfiguration updated for cross-platform compatibility.")
	fmt.Println("The project will now work across different operating systems.")
}

func runConfigureWizard() {
	fmt.Println("Product Feedback Tool Configuration Wizard")
	fmt.Println("==========================================")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Load existing config if available
	config, err := LoadConfig()
	if err != nil {
		config = NewDefaultConfig()
	}

	// Product name
	fmt.Printf("Product name [%s]: ", config.Name)
	if name, _ := reader.ReadString('\n'); strings.TrimSpace(name) != "" {
		config.Name = strings.TrimSpace(name)
	}

	// Document path
	fmt.Printf("Document path [%s]: ", config.Path)
	if path, _ := reader.ReadString('\n'); strings.TrimSpace(path) != "" {
		absPath, err := filepath.Abs(strings.TrimSpace(path))
		if err == nil {
			config.Path = absPath
		}
	}

	fmt.Println()
	fmt.Println("Configure areas (VoC, VoS, VoB, VoE):")
	fmt.Println("  Available providers: fider, clearflask, eververse, local")
	fmt.Println()

	// Configure each area
	areas := []struct {
		name   string
		label  string
		config **AreaConfig
	}{
		{"voc", "VoC (Voice of Customer)", &config.VoC},
		{"vos", "VoS (Voice of Stakeholder)", &config.VoS},
		{"vob", "VoB (Voice of Business)", &config.VoB},
		{"voe", "VoE (Voice of Engineer)", &config.VoE},
	}

	for _, area := range areas {
		currentProvider := "local"
		currentURL := ""
		if *area.config != nil {
			if (*area.config).Provider != "" {
				currentProvider = (*area.config).Provider
			}
			currentURL = (*area.config).URL
		}

		fmt.Printf("%s provider [%s]: ", area.label, currentProvider)
		providerInput, _ := reader.ReadString('\n')
		provider := strings.TrimSpace(providerInput)
		if provider == "" {
			provider = currentProvider
		}

		if provider != "local" {
			fmt.Printf("  URL [%s]: ", currentURL)
			urlInput, _ := reader.ReadString('\n')
			url := strings.TrimSpace(urlInput)
			if url == "" {
				url = currentURL
			}

			fmt.Print("  API Token (leave empty to keep current): ")
			tokenInput, _ := reader.ReadString('\n')
			token := strings.TrimSpace(tokenInput)

			if *area.config == nil {
				*area.config = &AreaConfig{}
			}
			(*area.config).Provider = provider
			(*area.config).URL = url
			if token != "" {
				(*area.config).APIToken = token
			}
		} else {
			// Local provider - clear external config
			*area.config = nil
		}
	}

	fmt.Println()

	// Validate
	if err := config.Validate(); err != nil {
		fmt.Printf("Configuration validation failed: %v\n", err)
		return
	}

	// Save to the specified path if set, otherwise to current directory
	savePath := config.Path
	if savePath == "" {
		savePath, _ = os.Getwd()
	}
	if err := config.Save(savePath); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
		return
	}

	fmt.Printf("Configuration saved to %s\n", GetConfigPath(savePath))
}

// Infrastructure command handlers
func handleDeployCommand(args []string) {
	config, err := LoadConfig()
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	var result *DeployResult

	switch config.GetProvider() {
	case "fider":
		result, err = Deploy(config)
	case "clearflask":
		result, err = DeployClearFlask(config)
	case "eververse":
		result, err = DeployEververse(config)
	case "email":
		result, err = DeployEmailOnly(config)
	default:
		fmt.Printf("Provider '%s' deployment not yet implemented.\n", config.GetProvider())
		fmt.Println("Currently supported: fider, clearflask, eververse, email")
		return
	}

	if err != nil {
		fmt.Printf("Deployment failed: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println(result.Message)
}

func handleStatusCommand(args []string) {
	// First check if we have a config
	config, err := LoadConfig()
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	fmt.Println("Product Feedback Tool Status")
	fmt.Println("============================")
	fmt.Printf("Product: %s\n", config.Name)
	fmt.Printf("Provider: %s\n", config.GetProvider())
	if config.GetProvider() == "email" {
		fmt.Println("Mode: Email-only (sync disabled)")
	} else {
		fmt.Printf("Endpoint: %s\n", config.GetEndpoint())
	}
	fmt.Println()

	switch config.GetProvider() {
	case "fider":
		status, err := GetStatus()
		if err != nil {
			fmt.Printf("Error checking status: %v\n", err)
			return
		}

		switch status {
		case "not_deployed":
			fmt.Println("Infrastructure: Not deployed")
			fmt.Println("  Run 'portunix pft deploy' to deploy Fider")
		case "running":
			fmt.Println("Infrastructure: Running ✓")
			fmt.Println()
			info, _ := GetContainerInfo()
			fmt.Println(info)
		case "stopped":
			fmt.Println("Infrastructure: Stopped")
			fmt.Println("  Run 'portunix pft deploy' to start")
		case "partial":
			fmt.Println("Infrastructure: Partially running")
			fmt.Println()
			info, _ := GetContainerInfo()
			fmt.Println(info)
		default:
			fmt.Printf("Infrastructure: %s\n", status)
		}

	case "email":
		status, err := GetEmailOnlyStatus()
		if err != nil {
			fmt.Printf("Error checking status: %v\n", err)
			return
		}

		switch status {
		case "not_deployed":
			fmt.Println("Infrastructure: Not deployed")
			fmt.Println("  Run 'portunix pft deploy' to deploy Mailhog")
		case "running":
			fmt.Println("Infrastructure: Running ✓")
			fmt.Println("  Mailhog UI: http://localhost:3200")
			fmt.Println("  SMTP: localhost:1025")
		case "stopped":
			fmt.Println("Infrastructure: Stopped")
			fmt.Println("  Run 'portunix pft deploy' to start")
		default:
			fmt.Printf("Infrastructure: %s\n", status)
		}

	case "clearflask":
		status, err := GetClearFlaskStatus()
		if err != nil {
			fmt.Printf("Error checking status: %v\n", err)
			return
		}

		switch status {
		case "not_deployed":
			fmt.Println("Infrastructure: Not deployed")
			fmt.Println("  Run 'portunix pft deploy' to deploy ClearFlask")
		case "running":
			fmt.Println("Infrastructure: Running ✓")
			fmt.Println()
			info, _ := GetClearFlaskContainerInfo()
			fmt.Println(info)
		case "stopped":
			fmt.Println("Infrastructure: Stopped")
			fmt.Println("  Run 'portunix pft deploy' to start")
		case "partial":
			fmt.Println("Infrastructure: Partially running")
			fmt.Println()
			info, _ := GetClearFlaskContainerInfo()
			fmt.Println(info)
		default:
			fmt.Printf("Infrastructure: %s\n", status)
		}

	case "eververse":
		status, err := GetEververseStatus()
		if err != nil {
			fmt.Printf("Error checking status: %v\n", err)
			return
		}

		switch status {
		case "not_deployed":
			fmt.Println("Infrastructure: Not deployed")
			fmt.Println("  Run 'portunix pft deploy' to deploy Eververse")
			fmt.Println("  Note: Eververse requires ~6GB RAM and 12 containers")
		case "running":
			fmt.Println("Infrastructure: Running ✓")
			fmt.Println()
			info, _ := GetEververseContainerInfo()
			fmt.Println(info)
		case "stopped":
			fmt.Println("Infrastructure: Stopped")
			fmt.Println("  Run 'portunix pft deploy' to start")
		case "partial":
			fmt.Println("Infrastructure: Partially running")
			fmt.Println("  Some Supabase services may still be starting (allow 60-120s)")
			fmt.Println()
			info, _ := GetEververseContainerInfo()
			fmt.Println(info)
		default:
			fmt.Printf("Infrastructure: %s\n", status)
		}

	default:
		fmt.Printf("Infrastructure status for '%s': Not implemented\n", config.GetProvider())
	}
}

func handleDestroyCommand(args []string) {
	config, err := LoadConfig()
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Check for --volumes flag
	removeVolumes := false
	for _, arg := range args {
		if arg == "--volumes" || arg == "-v" {
			removeVolumes = true
		}
	}

	switch config.GetProvider() {
	case "fider":
		if removeVolumes {
			fmt.Println("WARNING: This will remove all Fider data including the database!")
			fmt.Print("Are you sure? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Aborted.")
				return
			}
		}

		if err := Destroy(removeVolumes); err != nil {
			fmt.Printf("Destroy failed: %v\n", err)
			return
		}

	case "email":
		if err := DestroyEmailOnly(removeVolumes); err != nil {
			fmt.Printf("Destroy failed: %v\n", err)
			return
		}

	case "clearflask":
		if removeVolumes {
			fmt.Println("WARNING: This will remove all ClearFlask data including MySQL database and Elasticsearch indices!")
			fmt.Print("Are you sure? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Aborted.")
				return
			}
		}

		if err := DestroyClearFlask(removeVolumes); err != nil {
			fmt.Printf("Destroy failed: %v\n", err)
			return
		}

	case "eververse":
		if removeVolumes {
			fmt.Println("WARNING: This will remove all Eververse data including PostgreSQL database and Supabase storage!")
			fmt.Println("This includes data from all 12 Supabase stack containers.")
			fmt.Print("Are you sure? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Aborted.")
				return
			}
		}

		if err := DestroyEververse(removeVolumes); err != nil {
			fmt.Printf("Destroy failed: %v\n", err)
			return
		}

	default:
		fmt.Printf("Provider '%s' destroy not yet implemented.\n", config.GetProvider())
		return
	}
}

// checkEmailOnlyMode checks if provider is email and shows error
func checkEmailOnlyMode() bool {
	config, err := LoadConfig()
	if err != nil {
		return false
	}
	if config.GetProvider() == "email" {
		fmt.Println("Sync commands are not available in email-only mode.")
		fmt.Println("Set provider to 'fider', 'clearflask', or 'eververse' to enable synchronization.")
		return true
	}
	return false
}

// Synchronization command handlers (Phase 4 - stubs)
func handleSyncCommand(args []string) {
	if checkEmailOnlyMode() {
		return
	}

	// Parse flags
	var syncVoC, syncVoS, dryRun bool
	var vocToken, vosToken string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--voc":
			syncVoC = true
		case "--vos":
			syncVoS = true
		case "--dry-run":
			dryRun = true
		case "--voc-token":
			if i+1 < len(args) {
				vocToken = args[i+1]
				i++
			}
		case "--vos-token":
			if i+1 < len(args) {
				vosToken = args[i+1]
				i++
			}
		case "--help", "-h":
			showSyncHelp()
			return
		}
	}

	// If neither specified, sync both
	if !syncVoC && !syncVoS {
		syncVoC = true
		syncVoS = true
	}

	config, configFilePath, err := LoadConfigWithFilePath()
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Use cross-platform path resolution
	basePath := ResolveProjectPath(config, configFilePath, "")

	// Update config with tokens if provided
	if vocToken != "" {
		config.VoC.APIToken = vocToken
	}
	if vosToken != "" {
		config.VoS.APIToken = vosToken
	}

	fmt.Printf("Synchronizing %s with Fider...\n", config.Name)
	if dryRun {
		fmt.Println("(dry-run mode - no changes will be made)")
	}
	fmt.Println()

	// Sync VoC
	if syncVoC {
		fmt.Println("🔄 VoC (Voice of Customer):")
		vocDir := getVoiceDir(basePath, "voc")

		vocURL := config.VoC.URL
		if vocURL == "" {
			vocURL = "http://localhost:3100"
		}
		vocAPIToken := config.VoC.APIToken
		if vocAPIToken == "" {
			vocAPIToken = config.GetAPIToken()
		}

		if vocAPIToken == "" {
			fmt.Println("   ✗ No API token configured for VoC")
			fmt.Println("   Run: portunix pft sync --voc --voc-token <your-token>")
		} else {
			client := NewFiderClient(vocURL, vocAPIToken)

			// Step 1: Pull new posts from Fider
			fmt.Println("   📥 Pulling new posts from Fider...")
			pulled, skippedPull, err := PullFromFider(client, vocDir, "voc", dryRun)
			if err != nil {
				fmt.Printf("   ✗ Pull failed: %v\n", err)
			} else {
				fmt.Printf("      Pulled: %d, Skipped: %d\n", pulled, skippedPull)
			}

			// Step 2: Push new local files to Fider
			fmt.Println("   📤 Pushing new local files to Fider...")
			items, err := ScanFeedbackDirectory(vocDir, "voc")
			if err != nil {
				fmt.Printf("   ✗ Failed to scan directory: %v\n", err)
			} else {
				pushed, skippedPush, err := PushNewToFider(client, items, dryRun, config.Name)
				if err != nil {
					fmt.Printf("   ✗ Push failed: %v\n", err)
				} else {
					fmt.Printf("      Pushed: %d, Skipped (already synced): %d\n", pushed, skippedPush)
				}
			}
		}
		fmt.Println()
	}

	// Sync VoS
	if syncVoS {
		fmt.Println("🔄 VoS (Voice of Stakeholder):")
		vosDir := getVoiceDir(basePath, "vos")

		vosURL := config.VoS.URL
		if vosURL == "" {
			vosURL = "http://localhost:3101"
		}
		vosAPIToken := config.VoS.APIToken
		if vosAPIToken == "" {
			vosAPIToken = config.GetAPIToken()
		}

		if vosAPIToken == "" {
			fmt.Println("   ✗ No API token configured for VoS")
			fmt.Println("   Run: portunix pft sync --vos --vos-token <your-token>")
		} else {
			client := NewFiderClient(vosURL, vosAPIToken)

			// Step 1: Pull new posts from Fider
			fmt.Println("   📥 Pulling new posts from Fider...")
			pulled, skippedPull, err := PullFromFider(client, vosDir, "vos", dryRun)
			if err != nil {
				fmt.Printf("   ✗ Pull failed: %v\n", err)
			} else {
				fmt.Printf("      Pulled: %d, Skipped: %d\n", pulled, skippedPull)
			}

			// Step 2: Push new local files to Fider
			fmt.Println("   📤 Pushing new local files to Fider...")
			items, err := ScanFeedbackDirectory(vosDir, "vos")
			if err != nil {
				fmt.Printf("   ✗ Failed to scan directory: %v\n", err)
			} else {
				pushed, skippedPush, err := PushNewToFider(client, items, dryRun, config.Name)
				if err != nil {
					fmt.Printf("   ✗ Push failed: %v\n", err)
				} else {
					fmt.Printf("      Pushed: %d, Skipped (already synced): %d\n", pushed, skippedPush)
				}
			}
		}
		fmt.Println()
	}

	// Save updated config with tokens if they were provided
	if vocToken != "" || vosToken != "" {
		configPath, _ := findConfigFile()
		if configPath != "" {
			config.SaveToPath(configPath)
			fmt.Println("Configuration updated with API tokens.")
		}
	}

	fmt.Println("Sync complete.")
}

func showSyncHelp() {
	fmt.Println("Usage: portunix pft sync [options]")
	fmt.Println()
	fmt.Println("Bidirectional synchronization between local files and Fider.")
	fmt.Println()
	fmt.Println("This command will:")
	fmt.Println("  1. Pull new posts from Fider (posts not yet in local files)")
	fmt.Println("  2. Push new local files to Fider (files without Fider ID)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --voc              Sync only VoC (Voice of Customer)")
	fmt.Println("  --vos              Sync only VoS (Voice of Stakeholder)")
	fmt.Println("  --voc-token <tok>  Set VoC Fider API token")
	fmt.Println("  --vos-token <tok>  Set VoS Fider API token")
	fmt.Println("  --dry-run          Show what would be synced without making changes")
	fmt.Println()
	fmt.Println("Note: Files with Fider ID in metadata are considered synced.")
	fmt.Println("      New local files will get Fider ID added after push.")
}

func handlePullCommand(args []string) {
	if checkEmailOnlyMode() {
		return
	}

	// Parse flags
	var pullVoC, pullVoS, dryRun bool
	var vocToken, vosToken string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--voc":
			pullVoC = true
		case "--vos":
			pullVoS = true
		case "--dry-run":
			dryRun = true
		case "--voc-token":
			if i+1 < len(args) {
				vocToken = args[i+1]
				i++
			}
		case "--vos-token":
			if i+1 < len(args) {
				vosToken = args[i+1]
				i++
			}
		case "--help", "-h":
			showPullHelp()
			return
		}
	}

	// If neither specified, pull both
	if !pullVoC && !pullVoS {
		pullVoC = true
		pullVoS = true
	}

	config, configFilePath, err := LoadConfigWithFilePath()
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Use cross-platform path resolution
	basePath := ResolveProjectPath(config, configFilePath, "")

	// Update config with tokens if provided
	if vocToken != "" {
		config.VoC.APIToken = vocToken
	}
	if vosToken != "" {
		config.VoS.APIToken = vosToken
	}

	fmt.Println("Pulling feedback from Fider...")
	if dryRun {
		fmt.Println("(dry-run mode - no files will be created)")
	}
	fmt.Println()

	// Pull VoC
	if pullVoC {
		fmt.Println("📥 VoC (Voice of Customer):")
		vocDir := getVoiceDir(basePath, "voc")

		vocURL := config.VoC.URL
		if vocURL == "" {
			vocURL = "http://localhost:3100"
		}
		vocAPIToken := config.VoC.APIToken
		if vocAPIToken == "" {
			vocAPIToken = config.GetAPIToken()
		}

		if vocAPIToken == "" {
			fmt.Println("   ✗ No API token configured for VoC")
			fmt.Println("   Run: portunix pft pull --voc --voc-token <your-token>")
		} else {
			client := NewFiderClient(vocURL, vocAPIToken)
			created, skipped, err := PullFromFider(client, vocDir, "voc", dryRun)
			if err != nil {
				fmt.Printf("   ✗ Pull failed: %v\n", err)
			} else {
				fmt.Printf("   Created: %d, Skipped: %d\n", created, skipped)
			}
		}
		fmt.Println()
	}

	// Pull VoS
	if pullVoS {
		fmt.Println("📥 VoS (Voice of Stakeholder):")
		vosDir := getVoiceDir(basePath, "vos")

		vosURL := config.VoS.URL
		if vosURL == "" {
			vosURL = "http://localhost:3101"
		}
		vosAPIToken := config.VoS.APIToken
		if vosAPIToken == "" {
			vosAPIToken = config.GetAPIToken()
		}

		if vosAPIToken == "" {
			fmt.Println("   ✗ No API token configured for VoS")
			fmt.Println("   Run: portunix pft pull --vos --vos-token <your-token>")
		} else {
			client := NewFiderClient(vosURL, vosAPIToken)
			created, skipped, err := PullFromFider(client, vosDir, "vos", dryRun)
			if err != nil {
				fmt.Printf("   ✗ Pull failed: %v\n", err)
			} else {
				fmt.Printf("   Created: %d, Skipped: %d\n", created, skipped)
			}
		}
		fmt.Println()
	}

	// Save updated config with tokens if they were provided
	if vocToken != "" || vosToken != "" {
		configPath, _ := findConfigFile()
		if configPath != "" {
			config.SaveToPath(configPath)
			fmt.Println("Configuration updated with API tokens.")
		}
	}
}

func showPullHelp() {
	fmt.Println("Usage: portunix pft pull [options]")
	fmt.Println()
	fmt.Println("Pull feedback from Fider and save as local markdown files.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --voc              Pull only VoC (Voice of Customer) posts")
	fmt.Println("  --vos              Pull only VoS (Voice of Stakeholder) posts")
	fmt.Println("  --voc-token <tok>  Set VoC Fider API token")
	fmt.Println("  --vos-token <tok>  Set VoS Fider API token")
	fmt.Println("  --dry-run          Show what would be pulled without creating files")
	fmt.Println()
	fmt.Println("Note: Existing files are skipped (not overwritten).")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft pull --voc")
	fmt.Println("  portunix pft pull --dry-run")
}

func handlePushCommand(args []string) {
	if checkEmailOnlyMode() {
		return
	}

	// Parse flags
	var pushVoC, pushVoS, dryRun bool
	var vocToken, vosToken string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--voc":
			pushVoC = true
		case "--vos":
			pushVoS = true
		case "--dry-run":
			dryRun = true
		case "--voc-token":
			if i+1 < len(args) {
				vocToken = args[i+1]
				i++
			}
		case "--vos-token":
			if i+1 < len(args) {
				vosToken = args[i+1]
				i++
			}
		case "--help", "-h":
			showPushHelp()
			return
		}
	}

	// If neither specified, push both
	if !pushVoC && !pushVoS {
		pushVoC = true
		pushVoS = true
	}

	config, configFilePath, err := LoadConfigWithFilePath()
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Use cross-platform path resolution
	basePath := ResolveProjectPath(config, configFilePath, "")

	// Update config with tokens if provided
	if vocToken != "" {
		config.VoC.APIToken = vocToken
	}
	if vosToken != "" {
		config.VoS.APIToken = vosToken
	}

	fmt.Println("Pushing feedback to Fider...")
	if dryRun {
		fmt.Println("(dry-run mode - no changes will be made)")
	}
	fmt.Println()

	// Push VoC
	if pushVoC {
		fmt.Println("📤 VoC (Voice of Customer):")
		vocDir := getVoiceDir(basePath, "voc")

		vocURL := config.VoC.URL
		if vocURL == "" {
			vocURL = "http://localhost:3100"
		}
		vocAPIToken := config.VoC.APIToken
		if vocAPIToken == "" {
			vocAPIToken = config.GetAPIToken() // Fallback to legacy
		}

		if vocAPIToken == "" {
			fmt.Println("   ✗ No API token configured for VoC")
			fmt.Println("   Run: portunix pft push --voc --voc-token <your-token>")
		} else {
			items, err := ScanFeedbackDirectory(vocDir, "voc")
			if err != nil {
				fmt.Printf("   ✗ Failed to scan VoC directory: %v\n", err)
			} else if len(items) == 0 {
				fmt.Println("   No VoC documents found")
			} else {
				fmt.Printf("   Found %d documents in %s\n", len(items), vocDir)
				client := NewFiderClient(vocURL, vocAPIToken)
				if err := PushToFider(client, items, dryRun); err != nil {
					fmt.Printf("   ✗ Push failed: %v\n", err)
				}
			}
		}
		fmt.Println()
	}

	// Push VoS
	if pushVoS {
		fmt.Println("📤 VoS (Voice of Stakeholder):")
		vosDir := getVoiceDir(basePath, "vos")

		vosURL := config.VoS.URL
		if vosURL == "" {
			vosURL = "http://localhost:3101"
		}
		vosAPIToken := config.VoS.APIToken
		if vosAPIToken == "" {
			vosAPIToken = config.GetAPIToken() // Fallback to legacy
		}

		if vosAPIToken == "" {
			fmt.Println("   ✗ No API token configured for VoS")
			fmt.Println("   Run: portunix pft push --vos --vos-token <your-token>")
		} else {
			items, err := ScanFeedbackDirectory(vosDir, "vos")
			if err != nil {
				fmt.Printf("   ✗ Failed to scan VoS directory: %v\n", err)
			} else if len(items) == 0 {
				fmt.Println("   No VoS documents found")
			} else {
				fmt.Printf("   Found %d documents in %s\n", len(items), vosDir)
				client := NewFiderClient(vosURL, vosAPIToken)
				if err := PushToFider(client, items, dryRun); err != nil {
					fmt.Printf("   ✗ Push failed: %v\n", err)
				}
			}
		}
		fmt.Println()
	}

	// Save updated config with tokens if they were provided
	if vocToken != "" || vosToken != "" {
		configPath, _ := findConfigFile()
		if configPath != "" {
			config.SaveToPath(configPath)
			fmt.Println("Configuration updated with API tokens.")
		}
	}
}

func showPushHelp() {
	fmt.Println("Usage: portunix pft push [options]")
	fmt.Println()
	fmt.Println("Push local feedback documents to Fider.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --voc              Push only VoC (Voice of Customer) documents")
	fmt.Println("  --vos              Push only VoS (Voice of Stakeholder) documents")
	fmt.Println("  --voc-token <tok>  Set VoC Fider API token")
	fmt.Println("  --vos-token <tok>  Set VoS Fider API token")
	fmt.Println("  --dry-run          Show what would be pushed without making changes")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft push --voc --voc-token abc123")
	fmt.Println("  portunix pft push --dry-run")
	fmt.Println("  portunix pft push")
}

// Feedback management handlers
func handleListCommand(args []string) {
	// Parse flags
	var listVoC, listVoS, showAll, uncategorizedOnly bool
	var format string = "table"
	var categoryFilter string
	var configPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--voc":
			listVoC = true
		case "--vos":
			listVoS = true
		case "--all", "-a":
			showAll = true
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--category":
			if i+1 < len(args) {
				categoryFilter = args[i+1]
				i++
			}
		case "--uncategorized":
			uncategorizedOnly = true
		case "--path":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case "--help", "-h":
			showListHelp()
			return
		}
	}

	// Default: list both
	if !listVoC && !listVoS {
		listVoC = true
		listVoS = true
	}

	config, configFilePath, err := loadOrCreateConfig(configPath)
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Use cross-platform path resolution
	projectDir := ResolveProjectPath(config, configFilePath, configPath)

	fmt.Printf("Feedback Items - %s\n", config.Name)
	if categoryFilter != "" {
		fmt.Printf("Filter: category = %s\n", categoryFilter)
	} else if uncategorizedOnly {
		fmt.Printf("Filter: uncategorized items only\n")
	}
	fmt.Println(strings.Repeat("=", 50))

	var allItems []FeedbackItem

	// List VoC items
	if listVoC {
		vocDir := getVoiceDir(projectDir, "voc")
		vocItems, err := scanLocalDirectory(vocDir, "voc")
		if err == nil && len(vocItems) > 0 {
			// Apply category filter
			filteredItems := filterItemsByCategory(vocItems, categoryFilter, uncategorizedOnly)
			if len(filteredItems) > 0 {
				fmt.Printf("\n📢 Voice of Customer (VoC) - %d items\n", len(filteredItems))
				fmt.Println(strings.Repeat("-", 40))
				for _, item := range filteredItems {
					printFeedbackItem(item, format, showAll)
				}
				allItems = append(allItems, filteredItems...)
			}
		} else if err != nil {
			fmt.Printf("\n📢 Voice of Customer (VoC)\n")
			fmt.Printf("   No items found (directory: %s)\n", vocDir)
		}
	}

	// List VoS items
	if listVoS {
		vosDir := getVoiceDir(projectDir, "vos")
		vosItems, err := scanLocalDirectory(vosDir, "vos")
		if err == nil && len(vosItems) > 0 {
			// Apply category filter
			filteredItems := filterItemsByCategory(vosItems, categoryFilter, uncategorizedOnly)
			if len(filteredItems) > 0 {
				fmt.Printf("\n🏢 Voice of Stakeholder (VoS) - %d items\n", len(filteredItems))
				fmt.Println(strings.Repeat("-", 40))
				for _, item := range filteredItems {
					printFeedbackItem(item, format, showAll)
				}
				allItems = append(allItems, filteredItems...)
			}
		} else if err != nil {
			fmt.Printf("\n🏢 Voice of Stakeholder (VoS)\n")
			fmt.Printf("   No items found (directory: %s)\n", vosDir)
		}
	}

	fmt.Printf("\nTotal: %d items\n", len(allItems))
}

// filterItemsByCategory filters items by category or uncategorized status
func filterItemsByCategory(items []FeedbackItem, categoryFilter string, uncategorizedOnly bool) []FeedbackItem {
	if categoryFilter == "" && !uncategorizedOnly {
		return items // no filter
	}

	filtered := make([]FeedbackItem, 0, len(items))
	for _, item := range items {
		if uncategorizedOnly {
			if len(item.Categories) == 0 {
				filtered = append(filtered, item)
			}
		} else if categoryFilter != "" {
			for _, cat := range item.Categories {
				if cat == categoryFilter {
					filtered = append(filtered, item)
					break
				}
			}
		}
	}
	return filtered
}

func printFeedbackItem(item FeedbackItem, format string, showAll bool) {
	switch format {
	case "json":
		data, _ := json.MarshalIndent(item, "   ", "  ")
		fmt.Printf("   %s\n", string(data))
	default: // table
		status := item.Status
		if status == "" {
			status = "open"
		}
		syncMark := ""
		if item.ExternalID != "" {
			syncMark = " [synced]"
		}
		categoryMark := ""
		if len(item.Categories) > 0 {
			categoryMark = " [" + strings.Join(item.Categories, ", ") + "]"
		}
		fmt.Printf("   %-10s %-40s (%s)%s%s\n", item.ID, truncateStr(item.Title, 40), status, categoryMark, syncMark)
		if showAll && item.Description != "" {
			desc := truncateStr(item.Description, 70)
			fmt.Printf("              %s\n", desc)
		}
	}
}

// scanLocalDirectory wraps ScanFeedbackDirectory and converts []*FeedbackItem to []FeedbackItem
func scanLocalDirectory(dir string, feedbackType string) ([]FeedbackItem, error) {
	ptrItems, err := ScanFeedbackDirectory(dir, feedbackType)
	if err != nil {
		return nil, err
	}
	items := make([]FeedbackItem, len(ptrItems))
	for i, ptr := range ptrItems {
		items[i] = *ptr
	}
	return items, nil
}

// truncateStr wraps truncateString from email.go for convenience
func truncateStr(s string, maxLen int) string {
	return truncateString(s, maxLen)
}

func showListHelp() {
	fmt.Println("Usage: portunix pft list [options]")
	fmt.Println()
	fmt.Println("List all feedback items from local directories")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --voc              List only VoC (Voice of Customer) items")
	fmt.Println("  --vos              List only VoS (Voice of Stakeholder) items")
	fmt.Println("  --all, -a          Show full descriptions")
	fmt.Println("  --format <fmt>     Output format (table, json)")
	fmt.Println("  --category <id>    Filter by category")
	fmt.Println("  --uncategorized    Show only uncategorized items")
	fmt.Println("  --help, -h         Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft list")
	fmt.Println("  portunix pft list --voc")
	fmt.Println("  portunix pft list --all")
	fmt.Println("  portunix pft list --format json")
	fmt.Println("  portunix pft list --category user-auth")
	fmt.Println("  portunix pft list --uncategorized")
}

func handleShowCommand(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		showShowHelp()
		return
	}

	// Parse arguments - first non-flag argument is itemID
	var itemID string
	var configPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--path":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case "--help", "-h":
			showShowHelp()
			return
		default:
			if !strings.HasPrefix(args[i], "-") && itemID == "" {
				itemID = args[i]
			}
		}
	}

	if itemID == "" {
		fmt.Println("Error: item ID is required")
		showShowHelp()
		return
	}

	config, configFilePath, err := loadOrCreateConfig(configPath)
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Use cross-platform path resolution
	projectDir := ResolveProjectPath(config, configFilePath, configPath)

	// Try to find item in VoC or VoS directories
	item, filePath, err := findFeedbackItem(projectDir, itemID)
	if err != nil {
		fmt.Printf("Item '%s' not found: %v\n", itemID, err)
		return
	}

	// Display item details
	fmt.Printf("Feedback Item: %s\n", item.ID)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Title:       %s\n", item.Title)
	fmt.Printf("Status:      %s\n", item.Status)
	fmt.Printf("Type:        %s\n", item.Type)
	fmt.Printf("File:        %s\n", filePath)

	if item.ExternalID != "" {
		fmt.Printf("Synced:      Yes (Fider ID: %s)\n", item.ExternalID)
	} else {
		fmt.Printf("Synced:      No\n")
	}

	if item.Votes > 0 {
		fmt.Printf("Votes:       %d\n", item.Votes)
	}

	if len(item.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(item.Tags, ", "))
	}

	if item.CreatedAt != "" {
		fmt.Printf("Created:     %s\n", item.CreatedAt)
	}

	if item.UpdatedAt != "" {
		fmt.Printf("Updated:     %s\n", item.UpdatedAt)
	}

	fmt.Println()
	fmt.Println("Description:")
	fmt.Println(strings.Repeat("-", 50))
	if item.Description != "" {
		fmt.Println(item.Description)
	} else {
		fmt.Println("(no description)")
	}
}

func findFeedbackItem(projectDir, itemID string) (*FeedbackItem, string, error) {
	// Try VoC directory (uses recursive ScanFeedbackDirectory)
	vocDir := getVoiceDir(projectDir, "voc")
	if item, path, err := findItemInDirectory(vocDir, itemID, "voc"); err == nil {
		return item, path, nil
	}

	// Try VoS directory
	vosDir := getVoiceDir(projectDir, "vos")
	if item, path, err := findItemInDirectory(vosDir, itemID, "vos"); err == nil {
		return item, path, nil
	}

	// Try VoB directory
	vobDir := getVoiceDir(projectDir, "vob")
	if item, path, err := findItemInDirectory(vobDir, itemID, "vob"); err == nil {
		return item, path, nil
	}

	// Try VoE directory
	voeDir := getVoiceDir(projectDir, "voe")
	if item, path, err := findItemInDirectory(voeDir, itemID, "voe"); err == nil {
		return item, path, nil
	}

	return nil, "", fmt.Errorf("item not found in voc/, vos/, vob/, or voe/ directories")
}

func findItemInDirectory(dir, itemID, feedbackType string) (*FeedbackItem, string, error) {
	// Use recursive ScanFeedbackDirectory to find items in subdirectories (needs/, verbatims/, etc.)
	items, err := ScanFeedbackDirectory(dir, feedbackType)
	if err != nil {
		return nil, "", err
	}

	for _, item := range items {
		// Match by ID from frontmatter or by filename prefix
		filename := filepath.Base(item.FilePath)
		if item.ID == itemID || strings.HasPrefix(filename, itemID+"-") || strings.HasPrefix(filename, itemID+".") {
			return item, item.FilePath, nil
		}
	}

	return nil, "", fmt.Errorf("item not found")
}

// handleAddCommand adds a new feedback item
func handleAddCommand(args []string) {
	var area, title, description, verbatim, category, author, source, status, configPath string
	var priority, legacyID string
	var products, targetUsers, related, tags []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--area":
			if i+1 < len(args) {
				area = args[i+1]
				i++
			}
		case "--title":
			if i+1 < len(args) {
				title = args[i+1]
				i++
			}
		case "--description":
			if i+1 < len(args) {
				description = args[i+1]
				i++
			}
		case "--verbatim":
			if i+1 < len(args) {
				verbatim = args[i+1]
				i++
			}
		case "--category":
			if i+1 < len(args) {
				category = strings.ToUpper(args[i+1])
				i++
			}
		case "--author":
			if i+1 < len(args) {
				author = args[i+1]
				i++
			}
		case "--source":
			if i+1 < len(args) {
				source = args[i+1]
				i++
			}
		case "--status":
			if i+1 < len(args) {
				status = args[i+1]
				i++
			}
		case "--priority":
			if i+1 < len(args) {
				priority = args[i+1]
				i++
			}
		case "--legacy-id":
			if i+1 < len(args) {
				legacyID = args[i+1]
				i++
			}
		case "--product":
			if i+1 < len(args) {
				products = append(products, args[i+1])
				i++
			}
		case "--target-user":
			if i+1 < len(args) {
				targetUsers = append(targetUsers, args[i+1])
				i++
			}
		case "--related":
			if i+1 < len(args) {
				related = append(related, args[i+1])
				i++
			}
		case "--tag":
			if i+1 < len(args) {
				tags = append(tags, args[i+1])
				i++
			}
		case "--path":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case "--help", "-h":
			showAddHelp()
			return
		}
	}

	// Validate required fields
	if area == "" {
		fmt.Println("Error: --area is required (voc, vos, vob, voe)")
		return
	}
	if !IsValidArea(area) {
		fmt.Printf("Error: invalid area '%s'. Valid options: voc, vos, vob, voe\n", area)
		return
	}
	if title == "" {
		fmt.Println("Error: --title is required")
		return
	}

	// Default status
	if status == "" {
		status = "pending"
	}

	// Load config
	config, configFilePath, err := loadOrCreateConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		return
	}

	// Use cross-platform path resolution
	projectDir := ResolveProjectPath(config, configFilePath, configPath)

	// Lookup author role from user registry
	var authorRole string
	if author != "" {
		registry, err := LoadUserRegistry(projectDir)
		if err == nil {
			user := registry.FindUserByName(author)
			if user != nil {
				authorRole = user.GetRoleForArea(area)
			}
		}
	}

	// Get the target directory (QFD-compatible: use needs/ subdirectory)
	areaDir := getVoiceDir(projectDir, area)
	targetDir := filepath.Join(areaDir, "needs")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	// Generate unique ID
	itemID := generateNextItemID(areaDir, area)

	// Create slug from title
	slug := createSlugFromTitle(title)
	if len(slug) > 40 {
		slug = slug[:40]
	}

	// Create filename
	filename := fmt.Sprintf("%s-%s.md", itemID, slug)
	filePath := filepath.Join(targetDir, filename)

	// Generate markdown content
	params := FeedbackItemParams{
		ID:          itemID,
		Title:       title,
		Area:        area,
		Description: description,
		Verbatim:    verbatim,
		Status:      status,
		Category:    category,
		Author:      author,
		AuthorRole:  authorRole,
		Source:      source,
		Priority:    priority,
		LegacyID:    legacyID,
		Products:    products,
		TargetUsers: targetUsers,
		Related:     related,
		Tags:        tags,
	}
	content := generateFeedbackMarkdown(params)

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("✓ Created feedback item '%s' in %s\n", itemID, area)
	fmt.Printf("  File: %s\n", filePath)
	if category != "" {
		fmt.Printf("  Category: %s\n", category)
	}
}

// handleUpdateCommand updates an existing feedback item
func handleUpdateCommand(args []string) {
	if len(args) == 0 {
		showUpdateHelp()
		return
	}

	// Check for help flag first
	if args[0] == "--help" || args[0] == "-h" {
		showUpdateHelp()
		return
	}

	// First argument is the item ID
	itemID := args[0]
	var title, description, verbatim, category, author, source, status, configPath string
	var priority string
	var products, targetUsers, related, tags []string
	var clearProducts, clearTargetUsers, clearRelated, clearTags bool

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--title":
			if i+1 < len(args) {
				title = args[i+1]
				i++
			}
		case "--description":
			if i+1 < len(args) {
				description = args[i+1]
				i++
			}
		case "--verbatim":
			if i+1 < len(args) {
				verbatim = args[i+1]
				i++
			}
		case "--category":
			if i+1 < len(args) {
				category = strings.ToUpper(args[i+1])
				i++
			}
		case "--author":
			if i+1 < len(args) {
				author = args[i+1]
				i++
			}
		case "--source":
			if i+1 < len(args) {
				source = args[i+1]
				i++
			}
		case "--status":
			if i+1 < len(args) {
				status = args[i+1]
				i++
			}
		case "--priority":
			if i+1 < len(args) {
				priority = args[i+1]
				i++
			}
		case "--product":
			if i+1 < len(args) {
				products = append(products, args[i+1])
				i++
			}
		case "--target-user":
			if i+1 < len(args) {
				targetUsers = append(targetUsers, args[i+1])
				i++
			}
		case "--related":
			if i+1 < len(args) {
				related = append(related, args[i+1])
				i++
			}
		case "--tag":
			if i+1 < len(args) {
				tags = append(tags, args[i+1])
				i++
			}
		case "--clear-products":
			clearProducts = true
		case "--clear-target-users":
			clearTargetUsers = true
		case "--clear-related":
			clearRelated = true
		case "--clear-tags":
			clearTags = true
		case "--path":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case "--help", "-h":
			showUpdateHelp()
			return
		}
	}

	// Load config
	config, configFilePath, err := loadOrCreateConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		return
	}

	// Use cross-platform path resolution
	projectDir := ResolveProjectPath(config, configFilePath, configPath)

	// Find the item file
	var itemPath string
	var itemArea string
	areas := []string{"voc", "vos", "vob", "voe"}

	for _, area := range areas {
		areaDir := getVoiceDir(projectDir, area)
		needsDir := filepath.Join(areaDir, "needs")

		// Search for file matching the ID
		filepath.WalkDir(needsDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.HasPrefix(d.Name(), itemID+"-") && strings.HasSuffix(d.Name(), ".md") {
				itemPath = path
				itemArea = area
				return filepath.SkipAll
			}
			return nil
		})
		if itemPath != "" {
			break
		}
	}

	if itemPath == "" {
		fmt.Printf("Error: item '%s' not found\n", itemID)
		return
	}

	// Read existing file
	content, err := os.ReadFile(itemPath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Parse existing YAML frontmatter
	existingParams := parseExistingItem(string(content))
	if existingParams == nil {
		fmt.Printf("Error: could not parse item file\n")
		return
	}

	// Update fields if provided
	if title != "" {
		existingParams.Title = title
	}
	if description != "" {
		existingParams.Description = description
	}
	if verbatim != "" {
		existingParams.Verbatim = verbatim
	}
	if category != "" {
		existingParams.Category = category
	}
	if author != "" {
		existingParams.Author = author
	}
	if source != "" {
		existingParams.Source = source
	}
	if status != "" {
		existingParams.Status = status
	}
	if priority != "" {
		existingParams.Priority = priority
	}

	// Handle array fields
	if clearProducts {
		existingParams.Products = nil
	}
	if len(products) > 0 {
		existingParams.Products = append(existingParams.Products, products...)
	}

	if clearTargetUsers {
		existingParams.TargetUsers = nil
	}
	if len(targetUsers) > 0 {
		existingParams.TargetUsers = append(existingParams.TargetUsers, targetUsers...)
	}

	if clearRelated {
		existingParams.Related = nil
	}
	if len(related) > 0 {
		existingParams.Related = append(existingParams.Related, related...)
	}

	if clearTags {
		existingParams.Tags = nil
	}
	if len(tags) > 0 {
		existingParams.Tags = append(existingParams.Tags, tags...)
	}

	// Set area from found location
	existingParams.Area = itemArea

	// Generate updated content
	newContent := generateFeedbackMarkdown(*existingParams)

	// Write file
	if err := os.WriteFile(itemPath, []byte(newContent), 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("✓ Updated feedback item '%s'\n", itemID)
	fmt.Printf("  File: %s\n", itemPath)
}

// parseExistingItem parses an existing markdown file and returns FeedbackItemParams
func parseExistingItem(content string) *FeedbackItemParams {
	params := &FeedbackItemParams{}

	// Check for YAML frontmatter
	if !strings.HasPrefix(content, "---") {
		return nil
	}

	// Find end of frontmatter
	endIndex := strings.Index(content[3:], "---")
	if endIndex == -1 {
		return nil
	}

	frontmatter := content[3 : endIndex+3]
	lines := strings.Split(frontmatter, "\n")

	var currentArrayField string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is an array item
		if strings.HasPrefix(line, "- ") {
			value := strings.TrimPrefix(line, "- ")
			switch currentArrayField {
			case "products":
				params.Products = append(params.Products, value)
			case "target_users":
				params.TargetUsers = append(params.TargetUsers, value)
			case "related":
				params.Related = append(params.Related, value)
			case "tags":
				params.Tags = append(params.Tags, value)
			}
			continue
		}

		// Parse key: value
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Check if this starts an array
		if value == "" {
			currentArrayField = key
			continue
		}
		currentArrayField = ""

		switch key {
		case "id":
			params.ID = value
		case "title":
			params.Title = value
		case "area":
			params.Area = value
		case "description":
			params.Description = value
		case "verbatim":
			params.Verbatim = value
		case "category":
			params.Category = value
		case "status":
			params.Status = value
		case "priority":
			params.Priority = value
		case "legacy_id":
			params.LegacyID = value
		case "author":
			params.Author = value
		case "source":
			params.Source = value
		}
	}

	// Extract description from markdown body if not in frontmatter
	bodyStart := endIndex + 6 // Skip past "---\n"
	if bodyStart < len(content) {
		body := content[bodyStart:]
		// Look for ## Popis section
		if idx := strings.Index(body, "## Popis"); idx != -1 {
			descStart := idx + len("## Popis")
			// Find next section or end
			descEnd := strings.Index(body[descStart:], "##")
			if descEnd == -1 {
				descEnd = len(body) - descStart
			}
			desc := strings.TrimSpace(body[descStart : descStart+descEnd])
			if desc != "" && params.Description == "" {
				params.Description = desc
			}
		}

		// Look for ## Verbatim section first (new format)
		if idx := strings.Index(body, "## Verbatim"); idx != -1 && params.Verbatim == "" {
			verbStart := idx + len("## Verbatim")
			// Find next section or end
			verbEnd := strings.Index(body[verbStart:], "##")
			if verbEnd == -1 {
				verbEnd = len(body) - verbStart
			}
			verbSection := strings.TrimSpace(body[verbStart : verbStart+verbEnd])
			// Extract content from blockquote
			if strings.HasPrefix(verbSection, "> ") {
				params.Verbatim = strings.TrimPrefix(verbSection, "> ")
			} else if verbSection != "" {
				params.Verbatim = verbSection
			}
		}

		// Fallback: Look for verbatim quote (blockquote after title) - old format
		if params.Verbatim == "" {
			if idx := strings.Index(body, "\n> "); idx != -1 {
				quoteEnd := strings.Index(body[idx+3:], "\n\n")
				if quoteEnd == -1 {
					quoteEnd = strings.Index(body[idx+3:], "\n#")
				}
				if quoteEnd != -1 {
					params.Verbatim = strings.TrimSpace(body[idx+3 : idx+3+quoteEnd])
				}
			}
		}
	}

	return params
}

func showUpdateHelp() {
	fmt.Println("Usage: portunix pft update <id> [options]")
	fmt.Println()
	fmt.Println("Update an existing feedback item/requirement.")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <id>                  Item ID (e.g., P01, P02)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --title <title>       Update title")
	fmt.Println("  --description <text>  Update description")
	fmt.Println("  --verbatim <quote>    Update verbatim quote")
	fmt.Println("  --category <id>       Update category")
	fmt.Println("  --author <name>       Update author")
	fmt.Println("  --source <text>       Update source")
	fmt.Println("  --status <status>     Update status")
	fmt.Println("  --priority <level>    Update priority")
	fmt.Println("  --product <name>      Add product (can be used multiple times)")
	fmt.Println("  --target-user <user>  Add target user (can be used multiple times)")
	fmt.Println("  --related <id>        Add related item (can be used multiple times)")
	fmt.Println("  --tag <tag>           Add tag (can be used multiple times)")
	fmt.Println("  --clear-products      Clear all products before adding new")
	fmt.Println("  --clear-target-users  Clear all target users before adding new")
	fmt.Println("  --clear-related       Clear all related items before adding new")
	fmt.Println("  --clear-tags          Clear all tags before adding new")
	fmt.Println("  --path <path>         Path to PFT project")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft update P01 --status implemented")
	fmt.Println("  portunix pft update P01 --title \"New title\" --priority high")
	fmt.Println("  portunix pft update P01 --clear-tags --tag newtag1 --tag newtag2")
}

// generateNextItemID generates the next sequential ID (P01, P02, ...)
func generateNextItemID(areaDir, area string) string {
	maxNum := 0

	// Scan all files in area directory and subdirectories
	filepath.WalkDir(areaDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		// Extract number from filename like P01-title.md or P05-title.md
		name := d.Name()
		if strings.HasPrefix(name, "P") {
			// Find the number after P
			var num int
			fmt.Sscanf(name[1:], "%d", &num)
			if num > maxNum {
				maxNum = num
			}
		}
		return nil
	})

	return fmt.Sprintf("P%02d", maxNum+1)
}

// FeedbackItemParams contains all parameters for generating feedback markdown
type FeedbackItemParams struct {
	ID          string
	Title       string
	Area        string
	Description string
	Verbatim    string
	Status      string
	Category    string
	Author      string
	AuthorRole  string
	Source      string
	Priority    string
	LegacyID    string
	Products    []string
	TargetUsers []string
	Related     []string
	Tags        []string
}

// generateFeedbackMarkdown generates markdown content with YAML frontmatter
func generateFeedbackMarkdown(params FeedbackItemParams) string {
	var sb strings.Builder
	now := time.Now().Format("2006-01-02")

	// YAML frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", params.ID))
	sb.WriteString(fmt.Sprintf("title: %s\n", params.Title))
	sb.WriteString(fmt.Sprintf("area: %s\n", params.Area))
	if params.Category != "" {
		sb.WriteString(fmt.Sprintf("category: %s\n", strings.ToUpper(params.Category)))
	}
	sb.WriteString(fmt.Sprintf("status: %s\n", params.Status))
	if params.Priority != "" {
		sb.WriteString(fmt.Sprintf("priority: %s\n", params.Priority))
	}
	if params.LegacyID != "" {
		sb.WriteString(fmt.Sprintf("legacy_id: %s\n", params.LegacyID))
	}
	if params.Author != "" {
		sb.WriteString(fmt.Sprintf("author: %s\n", params.Author))
	}
	if params.AuthorRole != "" {
		sb.WriteString(fmt.Sprintf("author_role: %s\n", params.AuthorRole))
	}
	if params.Source != "" {
		sb.WriteString(fmt.Sprintf("source: %s\n", params.Source))
	}
	sb.WriteString(fmt.Sprintf("created: %s\n", now))
	sb.WriteString(fmt.Sprintf("updated: %s\n", now))

	// Array fields
	if len(params.Products) > 0 {
		sb.WriteString("products:\n")
		for _, p := range params.Products {
			sb.WriteString(fmt.Sprintf("  - %s\n", p))
		}
	}
	if len(params.TargetUsers) > 0 {
		sb.WriteString("target_users:\n")
		for _, u := range params.TargetUsers {
			sb.WriteString(fmt.Sprintf("  - %s\n", u))
		}
	}
	if len(params.Related) > 0 {
		sb.WriteString("related:\n")
		for _, r := range params.Related {
			sb.WriteString(fmt.Sprintf("  - %s\n", r))
		}
	}
	if len(params.Tags) > 0 {
		sb.WriteString("tags:\n")
		for _, tag := range params.Tags {
			sb.WriteString(fmt.Sprintf("  - %s\n", tag))
		}
	}

	sb.WriteString("---\n\n")

	// Markdown content
	sb.WriteString(fmt.Sprintf("# %s\n\n", params.Title))

	if params.Verbatim != "" {
		sb.WriteString("## Verbatim\n\n")
		sb.WriteString(fmt.Sprintf("> %s\n\n", params.Verbatim))
	}

	if params.Description != "" {
		sb.WriteString("## Popis\n\n")
		sb.WriteString(params.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Stav implementace\n\n")
	sb.WriteString("| Fáze | Stav | Poznámka |\n")
	sb.WriteString("|------|------|----------|\n")
	sb.WriteString("| Analýza | ⏳ | - |\n")
	sb.WriteString("| Vývoj | ⏳ | - |\n")
	sb.WriteString("| Release | ⏳ | - |\n")

	return sb.String()
}

// createSlugFromTitle creates a URL-friendly slug from a title
func createSlugFromTitle(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters (keep only alphanumeric, hyphens, and some unicode letters)
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' ||
			(r >= 'á' && r <= 'ž') { // Keep Czech characters
			result.WriteRune(r)
		}
	}
	slug = result.String()

	// Remove multiple consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}

func showAddHelp() {
	fmt.Println("Usage: portunix pft add [options]")
	fmt.Println()
	fmt.Println("Add a new feedback item/requirement to the project.")
	fmt.Println()
	fmt.Println("Required Options:")
	fmt.Println("  --area <area>         Target area (voc, vos, vob, voe)")
	fmt.Println("  --title <title>       Item title")
	fmt.Println()
	fmt.Println("Optional:")
	fmt.Println("  --description <text>  Item description")
	fmt.Println("  --verbatim <quote>    Verbatim quote from customer/stakeholder")
	fmt.Println("  --category <id>       Category ID (e.g., A, B, USER-AUTH)")
	fmt.Println("  --author <name>       Author name")
	fmt.Println("  --source <text>       Source of requirement (e.g., 'Email from John')")
	fmt.Println("  --status <status>     Initial status (default: pending)")
	fmt.Println("  --priority <level>    Priority level (e.g., high, medium, low)")
	fmt.Println("  --legacy-id <id>      Legacy ID from previous system (e.g., UC001)")
	fmt.Println("  --product <name>      Product name (can be used multiple times)")
	fmt.Println("  --target-user <user>  Target user type (can be used multiple times)")
	fmt.Println("  --related <id>        Related item ID (can be used multiple times)")
	fmt.Println("  --tag <tag>           Tag for categorization (can be used multiple times)")
	fmt.Println("  --path <path>         Path to PFT project")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft add --area vos --title \"Search summarization\"")
	fmt.Println("  portunix pft add --area voc --title \"Dark mode\" --category A --author \"John\"")
	fmt.Println("  portunix pft add --area voc --title \"Chat\" --legacy-id UC001 --product \"Tovek AI\" --tag ai")
}

func showShowHelp() {
	fmt.Println("Usage: portunix pft show <id> [options]")
	fmt.Println()
	fmt.Println("Show details of a specific feedback item")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <id>    Item ID (e.g., UC001, P01) or full slug (e.g., P01-feature-name)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --path <dir>    Path to PFT project directory")
	fmt.Println("  --help, -h      Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft show UC001")
	fmt.Println("  portunix pft show P01 --path docs/pft-project")
	fmt.Println("  portunix pft show P01-feature-name --path /path/to/project")
}

func handleLinkCommand(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		showLinkHelp()
		return
	}

	if len(args) < 2 {
		fmt.Println("Usage: portunix pft link <feedback-id> <issue-id>")
		fmt.Println("Run 'portunix pft link --help' for more information.")
		return
	}

	feedbackID := args[0]
	issueID := args[1]

	config, configFilePath, err := LoadConfigWithFilePath()
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Use cross-platform path resolution
	projectDir := ResolveProjectPath(config, configFilePath, "")

	// Find the feedback item
	item, filePath, err := findFeedbackItem(projectDir, feedbackID)
	if err != nil {
		fmt.Printf("Feedback item '%s' not found: %v\n", feedbackID, err)
		return
	}

	// Read the markdown file
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	contentStr := string(content)

	// Check if already linked
	if strings.Contains(contentStr, "linked_issue:") {
		// Update existing link
		lines := strings.Split(contentStr, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "linked_issue:") {
				lines[i] = fmt.Sprintf("linked_issue: %s", issueID)
				break
			}
		}
		contentStr = strings.Join(lines, "\n")
	} else {
		// Add link to metadata section (after frontmatter or at top)
		if strings.HasPrefix(contentStr, "---") {
			// Find end of frontmatter
			endIdx := strings.Index(contentStr[3:], "---")
			if endIdx > 0 {
				// Insert before closing ---
				insertPos := 3 + endIdx
				contentStr = contentStr[:insertPos] + fmt.Sprintf("linked_issue: %s\n", issueID) + contentStr[insertPos:]
			}
		} else {
			// Add at the top as metadata comment
			contentStr = fmt.Sprintf("<!-- linked_issue: %s -->\n\n%s", issueID, contentStr)
		}
	}

	// Write updated content
	if err := os.WriteFile(filePath, []byte(contentStr), 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Printf("✓ Linked feedback '%s' to issue '%s'\n", feedbackID, issueID)
	fmt.Printf("  File: %s\n", filePath)
	fmt.Printf("  Item: %s\n", item.Title)
}

func showLinkHelp() {
	fmt.Println("Usage: portunix pft link <feedback-id> <issue-id>")
	fmt.Println()
	fmt.Println("Link a feedback item to a local issue")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <feedback-id>  Feedback item ID (e.g., UC001, REQ001)")
	fmt.Println("  <issue-id>     Local issue ID (e.g., #107, ISSUE-42)")
	fmt.Println()
	fmt.Println("The link is stored in the feedback item's markdown file as metadata.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft link UC001 #107")
	fmt.Println("  portunix pft link REQ002 ISSUE-42")
}

// Notification handlers
func handleNotifyCommand(args []string) {
	if len(args) == 0 {
		showNotifyHelp()
		return
	}

	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showNotifyHelp()
			return
		}
	}

	// First argument is item ID
	itemID := args[0]

	// Parse flags
	var userEmail, notifyTypeStr string
	var allVoC, allVoS, dryRun bool

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--user":
			if i+1 < len(args) {
				userEmail = args[i+1]
				i++
			}
		case "--type":
			if i+1 < len(args) {
				notifyTypeStr = args[i+1]
				i++
			}
		case "--all-voc":
			allVoC = true
		case "--all-vos":
			allVoS = true
		case "--dry-run":
			dryRun = true
		}
	}

	// Validate notification type
	if notifyTypeStr == "" {
		fmt.Println("Error: --type is required")
		fmt.Println("Valid types: vote, description, acceptance")
		return
	}

	notifyType, err := ParseNotificationType(notifyTypeStr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Validate recipient selection
	if userEmail == "" && !allVoC && !allVoS {
		fmt.Println("Error: recipient required (--user, --all-voc, or --all-vos)")
		return
	}

	// Load config
	config, err := LoadConfig()
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Get project directory
	projectDir := getProjectDir()

	// Load feedback item (try local files first)
	feedbackItem, fiderURL, postNumber, err := loadFeedbackItem(projectDir, itemID, config)
	if err != nil {
		fmt.Printf("Error loading feedback item '%s': %v\n", itemID, err)
		return
	}

	// Prepare email data
	emailData := EmailData{
		ProductName: config.Name,
		Title:       feedbackItem.Title,
		Description: feedbackItem.Description,
		FiderURL:    fiderURL,
		PostNumber:  postNumber,
		Provider:    config.GetProvider(),
		ItemID:      itemID,
	}

	// Collect recipients
	var recipients []struct {
		Email string
		Name  string
	}

	if userEmail != "" {
		// Single user specified
		registry, err := LoadUserRegistry(projectDir)
		if err == nil {
			user := registry.FindUserByEmail(userEmail)
			if user != nil {
				recipients = append(recipients, struct {
					Email string
					Name  string
				}{Email: userEmail, Name: user.Name})
			} else {
				recipients = append(recipients, struct {
					Email string
					Name  string
				}{Email: userEmail, Name: userEmail})
			}
		} else {
			recipients = append(recipients, struct {
				Email string
				Name  string
			}{Email: userEmail, Name: userEmail})
		}
	}

	if allVoC || allVoS {
		registry, err := LoadUserRegistry(projectDir)
		if err != nil {
			fmt.Printf("Error loading user registry: %v\n", err)
			return
		}

		if allVoC {
			vocUsers := registry.ListUsersByCategory("voc")
			for _, user := range vocUsers {
				if user.ID != "" && strings.Contains(user.ID, "@") {
					recipients = append(recipients, struct {
						Email string
						Name  string
					}{Email: user.ID, Name: user.Name})
				}
			}
		}

		if allVoS {
			vosUsers := registry.ListUsersByCategory("vos")
			for _, user := range vosUsers {
				if user.ID != "" && strings.Contains(user.ID, "@") {
					recipients = append(recipients, struct {
						Email string
						Name  string
					}{Email: user.ID, Name: user.Name})
				}
			}
		}
	}

	if len(recipients) == 0 {
		fmt.Println("No recipients found.")
		return
	}

	// Prepare SMTP client
	var smtpConfig SMTPConfig
	if config.SMTP != nil {
		smtpConfig = *config.SMTP
	}
	if smtpConfig.Host == "" {
		smtpConfig.Host = "localhost"
	}
	if smtpConfig.Port == 0 {
		smtpConfig.Port = 3200
	}
	if smtpConfig.From == "" {
		smtpConfig.From = "noreply@localhost"
	}

	client := NewSMTPClient(&smtpConfig)

	fmt.Printf("Sending %s notifications for: %s\n", notifyType, itemID)
	if dryRun {
		fmt.Println("(dry-run mode - no emails will be sent)")
	}
	fmt.Println()

	// Send notifications
	successCount := 0
	failCount := 0

	for _, recipient := range recipients {
		emailData.UserName = recipient.Name
		if emailData.UserName == "" {
			emailData.UserName = recipient.Email
		}

		subject, body, err := GenerateNotification(notifyType, emailData)
		if err != nil {
			fmt.Printf("   Error generating email for %s: %v\n", recipient.Email, err)
			failCount++
			continue
		}

		if dryRun {
			fmt.Printf("Would send to: %s\n", recipient.Email)
			fmt.Printf("Subject: %s\n", subject)
			fmt.Println("---")
			fmt.Println(body)
			fmt.Println("---")
			fmt.Println()
			successCount++
		} else {
			if err := client.SendEmail(recipient.Email, subject, body); err != nil {
				fmt.Printf("   Failed to send to %s: %v\n", recipient.Email, err)
				failCount++
			} else {
				fmt.Printf("   Sent to: %s\n", recipient.Email)
				successCount++
			}
		}
	}

	fmt.Println()
	if dryRun {
		fmt.Printf("Would send %d email(s)\n", successCount)
	} else {
		fmt.Printf("Sent: %d, Failed: %d\n", successCount, failCount)
	}
}

func showNotifyHelp() {
	fmt.Println("Usage: portunix pft notify <item-id> [options]")
	fmt.Println()
	fmt.Println("Send notification emails to users requesting action on feedback items.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --user <email>     Send to specific user")
	fmt.Println("  --all-voc          Send to all users with VoC role")
	fmt.Println("  --all-vos          Send to all users with VoS role")
	fmt.Println("  --type <type>      Notification type (required)")
	fmt.Println("  --dry-run          Show email without sending")
	fmt.Println()
	fmt.Println("Notification types:")
	fmt.Println("  vote        - Request user to vote for/against requirement")
	fmt.Println("  description - Request user to provide more details")
	fmt.Println("  acceptance  - Request user to define acceptance criteria")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft notify UC001 --user user@example.com --type vote")
	fmt.Println("  portunix pft notify REQ001 --all-voc --type description")
	fmt.Println("  portunix pft notify UC001 --user test@test.com --type vote --dry-run")
}

// loadFeedbackItem loads a feedback item from local files
func loadFeedbackItem(projectDir, itemID string, config *Config) (*FeedbackItem, string, int, error) {
	// Determine category from item ID prefix
	var searchDir, fiderURL string

	itemIDUpper := strings.ToUpper(itemID)
	if strings.HasPrefix(itemIDUpper, "UC") || strings.HasPrefix(itemIDUpper, "FR") {
		searchDir = getVoiceDir(projectDir, "voc")
		fiderURL = config.VoC.URL
		if fiderURL == "" {
			fiderURL = "http://localhost:3100"
		}
	} else if strings.HasPrefix(itemIDUpper, "REQ") || strings.HasPrefix(itemIDUpper, "NFR") {
		searchDir = getVoiceDir(projectDir, "vos")
		fiderURL = config.VoS.URL
		if fiderURL == "" {
			fiderURL = "http://localhost:3101"
		}
	} else {
		// Try both directories
		vocDir := getVoiceDir(projectDir, "voc")
		vosDir := getVoiceDir(projectDir, "vos")

		item, url, num, err := tryLoadFromDir(vocDir, itemID, config.VoC.URL, "http://localhost:3100")
		if err == nil {
			return item, url, num, nil
		}

		item, url, num, err = tryLoadFromDir(vosDir, itemID, config.VoS.URL, "http://localhost:3101")
		if err == nil {
			return item, url, num, nil
		}

		return nil, "", 0, fmt.Errorf("feedback item '%s' not found in voc/ or vos/", itemID)
	}

	return tryLoadFromDir(searchDir, itemID, fiderURL, fiderURL)
}

func tryLoadFromDir(dir, itemID, configURL, defaultURL string) (*FeedbackItem, string, int, error) {
	fiderURL := configURL
	if fiderURL == "" {
		fiderURL = defaultURL
	}

	// Search for file matching item ID
	pattern := filepath.Join(dir, fmt.Sprintf("%s*.md", itemID))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, "", 0, err
	}

	// Also try lowercase
	patternLower := filepath.Join(dir, fmt.Sprintf("%s*.md", strings.ToLower(itemID)))
	matchesLower, _ := filepath.Glob(patternLower)
	matches = append(matches, matchesLower...)

	if len(matches) == 0 {
		return nil, "", 0, fmt.Errorf("file not found for '%s'", itemID)
	}

	// Read first matching file
	content, err := os.ReadFile(matches[0])
	if err != nil {
		return nil, "", 0, err
	}

	// Parse markdown file
	item := &FeedbackItem{ID: itemID}
	lines := strings.Split(string(content), "\n")
	var fiderID int

	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			item.Title = strings.TrimPrefix(line, "# ")
			// Remove ID prefix from title if present
			if strings.Contains(item.Title, ":") {
				parts := strings.SplitN(item.Title, ":", 2)
				if len(parts) == 2 {
					item.Title = strings.TrimSpace(parts[1])
				}
			}
		}
		if strings.HasPrefix(line, "## Summary") || strings.HasPrefix(line, "## Description") {
			// Next non-empty lines are description
			continue
		}
		if strings.HasPrefix(line, "fider_id:") || strings.HasPrefix(line, "Fider ID:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &fiderID)
				item.ExternalID = strings.TrimSpace(parts[1])
			}
		}
	}

	// Extract description from Summary or Description section
	item.Description = extractSection(string(content), "Summary")
	if item.Description == "" {
		item.Description = extractSection(string(content), "Description")
	}

	return item, fiderURL, fiderID, nil
}

func extractSection(content, sectionName string) string {
	lines := strings.Split(content, "\n")
	inSection := false
	var result []string

	for _, line := range lines {
		if strings.HasPrefix(line, "## "+sectionName) {
			inSection = true
			continue
		}
		if inSection {
			if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "# ") {
				break
			}
			if strings.TrimSpace(line) != "" {
				result = append(result, line)
			}
		}
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}

// Reporting handlers
func handleReportCommand(args []string) {
	// Parse flags
	var reportType string = "summary"
	var outputFile string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--type":
			if i+1 < len(args) {
				reportType = args[i+1]
				i++
			}
		case "--output", "-o":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "--help", "-h":
			showReportHelp()
			return
		}
	}

	config, configFilePath, err := LoadConfigWithFilePath()
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Use cross-platform path resolution
	projectDir := ResolveProjectPath(config, configFilePath, "")

	// Collect all items
	var allItems []FeedbackItem
	vocDir := getVoiceDir(projectDir, "voc")
	vosDir := getVoiceDir(projectDir, "vos")

	vocItems, _ := scanLocalDirectory(vocDir, "voc")
	vosItems, _ := scanLocalDirectory(vosDir, "vos")
	allItems = append(allItems, vocItems...)
	allItems = append(allItems, vosItems...)

	// Generate report
	var report strings.Builder

	report.WriteString(fmt.Sprintf("# Feedback Report: %s\n\n", config.Name))
	report.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	switch reportType {
	case "summary":
		generateSummaryReport(&report, vocItems, vosItems)
	case "detailed":
		generateDetailedReport(&report, allItems)
	case "status":
		generateStatusReport(&report, allItems)
	default:
		generateSummaryReport(&report, vocItems, vosItems)
	}

	// Output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(report.String()), 0644); err != nil {
			fmt.Printf("Error writing report: %v\n", err)
			return
		}
		fmt.Printf("Report written to: %s\n", outputFile)
	} else {
		fmt.Println(report.String())
	}
}

func generateSummaryReport(report *strings.Builder, vocItems, vosItems []FeedbackItem) {
	report.WriteString("## Summary\n\n")
	report.WriteString(fmt.Sprintf("- **Voice of Customer (VoC)**: %d items\n", len(vocItems)))
	report.WriteString(fmt.Sprintf("- **Voice of Stakeholder (VoS)**: %d items\n", len(vosItems)))
	report.WriteString(fmt.Sprintf("- **Total**: %d items\n\n", len(vocItems)+len(vosItems)))

	// Count by status
	statusCounts := make(map[string]int)
	for _, item := range vocItems {
		status := item.Status
		if status == "" {
			status = "open"
		}
		statusCounts[status]++
	}
	for _, item := range vosItems {
		status := item.Status
		if status == "" {
			status = "open"
		}
		statusCounts[status]++
	}

	report.WriteString("## Status Distribution\n\n")
	for status, count := range statusCounts {
		report.WriteString(fmt.Sprintf("- %s: %d\n", status, count))
	}
	report.WriteString("\n")

	// Count by category
	allItems := append(vocItems, vosItems...)
	categoryCounts := make(map[string]int)
	uncategorizedCount := 0
	for _, item := range allItems {
		if len(item.Categories) == 0 {
			uncategorizedCount++
		} else {
			for _, cat := range item.Categories {
				categoryCounts[cat]++
			}
		}
	}

	report.WriteString("## Category Distribution\n\n")
	if len(categoryCounts) > 0 {
		for cat, count := range categoryCounts {
			report.WriteString(fmt.Sprintf("- %s: %d\n", cat, count))
		}
	}
	report.WriteString(fmt.Sprintf("- (uncategorized): %d\n", uncategorizedCount))
	report.WriteString("\n")

	// Count synced vs unsynced
	syncedCount := 0
	for _, item := range vocItems {
		if item.ExternalID != "" {
			syncedCount++
		}
	}
	for _, item := range vosItems {
		if item.ExternalID != "" {
			syncedCount++
		}
	}
	unsyncedCount := len(vocItems) + len(vosItems) - syncedCount

	report.WriteString("## Sync Status\n\n")
	report.WriteString(fmt.Sprintf("- Synced with Fider: %d\n", syncedCount))
	report.WriteString(fmt.Sprintf("- Local only: %d\n", unsyncedCount))
}

func generateDetailedReport(report *strings.Builder, items []FeedbackItem) {
	report.WriteString("## All Feedback Items\n\n")

	for _, item := range items {
		report.WriteString(fmt.Sprintf("### %s: %s\n\n", item.ID, item.Title))
		report.WriteString(fmt.Sprintf("- **Type**: %s\n", item.Type))
		report.WriteString(fmt.Sprintf("- **Status**: %s\n", item.Status))
		if len(item.Categories) > 0 {
			report.WriteString(fmt.Sprintf("- **Categories**: %s\n", strings.Join(item.Categories, ", ")))
		}
		if item.ExternalID != "" {
			report.WriteString(fmt.Sprintf("- **Fider ID**: %s\n", item.ExternalID))
		}
		if item.Votes > 0 {
			report.WriteString(fmt.Sprintf("- **Votes**: %d\n", item.Votes))
		}
		report.WriteString("\n")
		if item.Description != "" {
			report.WriteString(item.Description + "\n\n")
		}
		report.WriteString("---\n\n")
	}
}

func generateStatusReport(report *strings.Builder, items []FeedbackItem) {
	report.WriteString("## Status Report\n\n")
	report.WriteString("| ID | Title | Type | Status | Categories | Synced |\n")
	report.WriteString("|-----|-------|------|--------|------------|--------|\n")

	for _, item := range items {
		status := item.Status
		if status == "" {
			status = "open"
		}
		synced := "No"
		if item.ExternalID != "" {
			synced = "Yes"
		}
		categories := "-"
		if len(item.Categories) > 0 {
			categories = strings.Join(item.Categories, ", ")
		}
		report.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			item.ID, truncateStr(item.Title, 30), item.Type, status, categories, synced))
	}
}

func showReportHelp() {
	fmt.Println("Usage: portunix pft report [options]")
	fmt.Println()
	fmt.Println("Generate a feedback report")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --type <type>   Report type: summary, detailed, status (default: summary)")
	fmt.Println("  --output, -o    Output file (default: stdout)")
	fmt.Println("  --help, -h      Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft report")
	fmt.Println("  portunix pft report --type detailed")
	fmt.Println("  portunix pft report --type status -o report.md")
}

func handleExportCommand(args []string) {
	// Parse flags
	format := "md"
	var outputFile string
	var exportVoC, exportVoS bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--output", "-o":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "--voc":
			exportVoC = true
		case "--vos":
			exportVoS = true
		case "--help", "-h":
			showExportHelp()
			return
		}
		// Also support --format=xxx
		if strings.HasPrefix(args[i], "--format=") {
			format = strings.TrimPrefix(args[i], "--format=")
		}
	}

	// Default: export both
	if !exportVoC && !exportVoS {
		exportVoC = true
		exportVoS = true
	}

	config, configFilePath, err := LoadConfigWithFilePath()
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Use cross-platform path resolution
	projectDir := ResolveProjectPath(config, configFilePath, "")

	// Collect items
	var allItems []FeedbackItem
	if exportVoC {
		vocDir := getVoiceDir(projectDir, "voc")
		vocItems, _ := scanLocalDirectory(vocDir, "voc")
		allItems = append(allItems, vocItems...)
	}
	if exportVoS {
		vosDir := getVoiceDir(projectDir, "vos")
		vosItems, _ := scanLocalDirectory(vosDir, "vos")
		allItems = append(allItems, vosItems...)
	}

	// Export
	var output string
	switch format {
	case "json":
		data, err := json.MarshalIndent(allItems, "", "  ")
		if err != nil {
			fmt.Printf("Error creating JSON: %v\n", err)
			return
		}
		output = string(data)
	case "csv":
		var csv strings.Builder
		csv.WriteString("ID,Title,Type,Status,Categories,Votes,Synced\n")
		for _, item := range allItems {
			synced := "false"
			if item.ExternalID != "" {
				synced = "true"
			}
			categories := strings.Join(item.Categories, ";")
			csv.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",%d,%s\n",
				item.ID, item.Title, item.Type, item.Status, categories, item.Votes, synced))
		}
		output = csv.String()
	default: // md
		var md strings.Builder
		md.WriteString(fmt.Sprintf("# Feedback Export: %s\n\n", config.Name))
		md.WriteString(fmt.Sprintf("Exported: %s\n\n", time.Now().Format("2006-01-02")))
		for _, item := range allItems {
			md.WriteString(fmt.Sprintf("## %s: %s\n\n", item.ID, item.Title))
			catInfo := ""
			if len(item.Categories) > 0 {
				catInfo = fmt.Sprintf(" | **Categories:** %s", strings.Join(item.Categories, ", "))
			}
			md.WriteString(fmt.Sprintf("**Type:** %s | **Status:** %s%s\n\n", item.Type, item.Status, catInfo))
			if item.Description != "" {
				md.WriteString(item.Description + "\n\n")
			}
			md.WriteString("---\n\n")
		}
		output = md.String()
	}

	// Output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			fmt.Printf("Error writing export: %v\n", err)
			return
		}
		fmt.Printf("Exported %d items to: %s (format: %s)\n", len(allItems), outputFile, format)
	} else {
		fmt.Println(output)
	}
}

func showExportHelp() {
	fmt.Println("Usage: portunix pft export [options]")
	fmt.Println()
	fmt.Println("Export feedback items to various formats")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --format <fmt>  Export format: md, json, csv (default: md)")
	fmt.Println("  --output, -o    Output file (default: stdout)")
	fmt.Println("  --voc           Export only VoC items")
	fmt.Println("  --vos           Export only VoS items")
	fmt.Println("  --help, -h      Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft export")
	fmt.Println("  portunix pft export --format json -o items.json")
	fmt.Println("  portunix pft export --format csv --voc -o voc.csv")
}

func handleCacheCommand(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		showCacheHelp()
		return
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "status":
		handleCacheStatus(subArgs)
	case "clear":
		handleCacheClear(subArgs)
	case "cleanup":
		handleCacheCleanup(subArgs)
	default:
		fmt.Printf("Unknown cache subcommand: %s\n", subCmd)
		showCacheHelp()
	}
}

func handleCacheStatus(args []string) {
	projectDir := "."
	for i := 0; i < len(args); i++ {
		if args[i] == "--path" && i+1 < len(args) {
			projectDir = args[i+1]
			i++
		}
	}

	cache := NewSyncCache(projectDir)
	if err := cache.Load(); err != nil {
		fmt.Printf("Error loading cache: %v\n", err)
		return
	}

	cache.PrintCacheStatus()

	// Show recent entries
	entries := cache.GetAll()
	if len(entries) > 0 {
		fmt.Println()
		fmt.Println("Recent entries:")
		count := min(5, len(entries))
		for i := 0; i < count; i++ {
			e := entries[i]
			syncStatus := "not synced"
			if e.ExternalID != "" {
				syncStatus = fmt.Sprintf("synced (#%s)", e.ExternalID)
			}
			fmt.Printf("   %s: %s [%s]\n", e.ID, truncate(e.Title, 40), syncStatus)
		}
		if len(entries) > 5 {
			fmt.Printf("   ... and %d more\n", len(entries)-5)
		}
	}
}

func handleCacheClear(args []string) {
	projectDir := "."
	for i := 0; i < len(args); i++ {
		if args[i] == "--path" && i+1 < len(args) {
			projectDir = args[i+1]
			i++
		}
	}

	cache := NewSyncCache(projectDir)
	if err := cache.Load(); err != nil {
		fmt.Printf("Error loading cache: %v\n", err)
		return
	}

	entriesCount := len(cache.Entries)
	cache.Clear()

	if err := cache.Save(); err != nil {
		fmt.Printf("Error saving cache: %v\n", err)
		return
	}

	fmt.Printf("✓ Cache cleared (%d entries removed)\n", entriesCount)
}

func handleCacheCleanup(args []string) {
	projectDir := "."
	for i := 0; i < len(args); i++ {
		if args[i] == "--path" && i+1 < len(args) {
			projectDir = args[i+1]
			i++
		}
	}

	cache := NewSyncCache(projectDir)
	if err := cache.Load(); err != nil {
		fmt.Printf("Error loading cache: %v\n", err)
		return
	}

	removed := cache.CleanupOrphans()

	if err := cache.Save(); err != nil {
		fmt.Printf("Error saving cache: %v\n", err)
		return
	}

	if removed > 0 {
		fmt.Printf("✓ Cleaned up %d orphan entries\n", removed)
	} else {
		fmt.Println("✓ No orphan entries found")
	}
}

func showCacheHelp() {
	fmt.Println("Usage: portunix pft cache <subcommand> [options]")
	fmt.Println()
	fmt.Println("Manage local sync cache")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  status    Show cache status and statistics")
	fmt.Println("  clear     Clear all cache entries")
	fmt.Println("  cleanup   Remove orphan entries (files that no longer exist)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --path <dir>  Project directory (default: current)")
	fmt.Println("  --help, -h    Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft cache status")
	fmt.Println("  portunix pft cache clear")
	fmt.Println("  portunix pft cache cleanup --path ./my-project")
}

// Example command - creates demo with VoC/VoS structure and 2x Fider
func handleExampleCommand(args []string) {
	// Parse --path flag
	demoPath := ""
	noDeploy := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--path":
			if i+1 < len(args) {
				demoPath = args[i+1]
				i++
			}
		case "--no-deploy":
			noDeploy = true
		case "--help", "-h":
			showExampleHelp()
			return
		}
	}

	// Default demo path
	if demoPath == "" {
		cwd, _ := os.Getwd()
		demoPath = filepath.Join(cwd, "pft-demo")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(demoPath)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		return
	}
	demoPath = absPath

	fmt.Println("Product Feedback Tool - Quick Example (ISO 16355 QFD)")
	fmt.Println("=====================================================")
	fmt.Println()

	// Step 1: Create demo directory structure
	fmt.Printf("1. Creating demo directory structure: %s\n", demoPath)
	vocPath := filepath.Join(demoPath, "voc")
	vosPath := filepath.Join(demoPath, "vos")

	if err := os.MkdirAll(vocPath, 0755); err != nil {
		fmt.Printf("   Error creating voc/: %v\n", err)
		return
	}
	fmt.Println("   ✓ Created: voc/ (Voice of Customer - public)")

	if err := os.MkdirAll(vosPath, 0755); err != nil {
		fmt.Printf("   Error creating vos/: %v\n", err)
		return
	}
	fmt.Println("   ✓ Created: vos/ (Voice of Stakeholder - internal)")

	// Step 2: Create VoC samples (customer feedback)
	fmt.Println()
	fmt.Println("2. Creating VoC samples (customer feedback)...")
	vocSamples := getVoCSamples()
	for _, sample := range vocSamples {
		samplePath := filepath.Join(vocPath, sample.Filename)
		if err := os.WriteFile(samplePath, []byte(sample.Content), 0644); err != nil {
			fmt.Printf("   Error creating %s: %v\n", sample.Filename, err)
			continue
		}
		fmt.Printf("   ✓ voc/%s: %s\n", sample.Filename, sample.Title)
	}

	// Step 3: Create VoS samples (stakeholder requirements)
	fmt.Println()
	fmt.Println("3. Creating VoS samples (stakeholder requirements)...")
	vosSamples := getVoSSamples()
	for _, sample := range vosSamples {
		samplePath := filepath.Join(vosPath, sample.Filename)
		if err := os.WriteFile(samplePath, []byte(sample.Content), 0644); err != nil {
			fmt.Printf("   Error creating %s: %v\n", sample.Filename, err)
			continue
		}
		fmt.Printf("   ✓ vos/%s: %s\n", sample.Filename, sample.Title)
	}

	// Step 4: Configure ptx-pft
	fmt.Println()
	fmt.Println("4. Configuring ptx-pft...")
	config := NewDefaultConfig()
	config.Name = "Demo Product"
	config.Path = demoPath
	// Configure VoC with Fider
	config.VoC = &AreaConfig{
		Provider: "fider",
		URL:      "http://localhost:3000",
	}

	if err := config.Save(demoPath); err != nil {
		fmt.Printf("   Error saving config: %v\n", err)
		return
	}
	fmt.Printf("   ✓ Configuration saved to %s\n", GetConfigPath(demoPath))

	// Step 5: Deploy feedback tools (unless --no-deploy)
	if noDeploy {
		fmt.Println()
		fmt.Println("5. Skipping deployment (--no-deploy flag)")
		fmt.Println()
		showExampleSummary(demoPath, vocSamples, vosSamples, false)
		return
	}

	fmt.Println()
	fmt.Println("5. Checking compose readiness...")

	// Preflight check - verify compose is ready before attempting deployment
	preflight, err := CheckComposePreflight()
	if err != nil {
		fmt.Printf("   ⚠ Could not check compose readiness: %v\n", err)
		fmt.Println()
		showExampleSummary(demoPath, vocSamples, vosSamples, false)
		return
	}

	if !preflight.Ready {
		fmt.Println("   ❌ Compose is NOT ready")
		fmt.Println()
		fmt.Printf("   Problem: %s\n", preflight.ErrorMessage)
		fmt.Println()

		// Check if we can offer to fix it automatically (no sudo needed for user socket)
		if strings.Contains(preflight.FixInstructions, "systemctl --user") {
			fmt.Printf("   To fix this, run: %s\n", preflight.FixInstructions)
			fmt.Println()
			fmt.Print("   Do you want to run this command now? [Y/n]: ")

			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer == "" || answer == "y" || answer == "yes" || answer == "a" || answer == "ano" {
				fmt.Println()
				fmt.Printf("   Running: %s\n", preflight.FixInstructions)

				// Parse and execute the command
				parts := strings.Fields(preflight.FixInstructions)
				cmd := exec.Command(parts[0], parts[1:]...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				if err := cmd.Run(); err != nil {
					fmt.Printf("   ❌ Command failed: %v\n", err)
					fmt.Println()
					showExampleSummary(demoPath, vocSamples, vosSamples, false)
					return
				}

				fmt.Println("   ✓ Socket started successfully")
				fmt.Println()

				// Re-check preflight
				preflight, err = CheckComposePreflight()
				if err != nil || !preflight.Ready {
					fmt.Println("   ⚠ Socket started but compose still not ready")
					fmt.Println("   Please try running 'portunix pft example' again.")
					fmt.Println()
					showExampleSummary(demoPath, vocSamples, vosSamples, false)
					return
				}
				fmt.Println("   ✓ Compose is now ready")
			} else {
				fmt.Println()
				fmt.Println("   After fixing the issue, run 'portunix pft example' again.")
				fmt.Println()
				showExampleSummary(demoPath, vocSamples, vosSamples, false)
				return
			}
		} else {
			fmt.Printf("   Solution: %s\n", preflight.FixInstructions)
			fmt.Println()
			fmt.Println("   After fixing the issue, run 'portunix pft example' again.")
			fmt.Println()
			showExampleSummary(demoPath, vocSamples, vosSamples, false)
			return
		}
	} else {
		fmt.Println("   ✓ Compose is ready")
	}

	fmt.Println()
	fmt.Println("6. Deploying feedback tools (2x Fider)...")

	// Deploy VoC Fider (port 3100)
	fmt.Println()
	fmt.Println("   6a. VoC Fider (public, port 3100)...")
	vocConfig := NewDefaultConfig()
	vocConfig.Name = "Demo Product - VoC"
	vocConfig.Path = vocPath
	vocConfig.VoC = &AreaConfig{
		Provider: "fider",
		URL:      "http://localhost:3100",
	}

	vocResult, vocErr := DeployInstance("voc", 3100, vocConfig)
	if vocErr != nil {
		fmt.Printf("   ⚠ VoC deployment failed: %v\n", vocErr)
	} else {
		fmt.Println("   ✓ VoC Fider deployed on port 3100")
		_ = vocResult
	}

	// Deploy VoS Fider (port 3101)
	fmt.Println()
	fmt.Println("   6b. VoS Fider (internal, port 3101)...")
	vosConfig := NewDefaultConfig()
	vosConfig.Name = "Demo Product - VoS"
	vosConfig.Path = vosPath
	vosConfig.VoS = &AreaConfig{
		Provider: "fider",
		URL:      "http://localhost:3101",
	}

	vosResult, vosErr := DeployInstance("vos", 3101, vosConfig)
	if vosErr != nil {
		fmt.Printf("   ⚠ VoS deployment failed: %v\n", vosErr)
	} else {
		fmt.Println("   ✓ VoS Fider deployed on port 3101")
		_ = vosResult
	}

	fmt.Println()
	showExampleSummary(demoPath, vocSamples, vosSamples, vocErr == nil && vosErr == nil)
}

func showExampleSummary(demoPath string, vocSamples, vosSamples []SampleDocument, deployed bool) {
	fmt.Println("=====================================================")
	fmt.Println("Demo setup complete!")
	fmt.Println()
	fmt.Println("Directory structure:")
	fmt.Printf("  %s/\n", demoPath)
	fmt.Println("  ├── voc/                 (Voice of Customer - public)")
	for _, s := range vocSamples {
		fmt.Printf("  │   └── %s\n", s.Filename)
	}
	fmt.Println("  ├── vos/                 (Voice of Stakeholder - internal)")
	for _, s := range vosSamples {
		fmt.Printf("  │   └── %s\n", s.Filename)
	}
	fmt.Println("  └── .pft-config.json")
	fmt.Println()

	if deployed {
		fmt.Println("Fider instances:")
		fmt.Println("  VoC (public):   http://localhost:3100")
		fmt.Println("  VoS (internal): http://localhost:3101")
		fmt.Println()
		fmt.Println("Email capture (Mailhog):")
		fmt.Println("  VoC Mailhog:    http://localhost:3200")
		fmt.Println("  VoS Mailhog:    http://localhost:3201")
		fmt.Println()
		fmt.Println("Registration steps:")
		fmt.Println("  1. Open http://localhost:3100 (VoC Fider)")
		fmt.Println("  2. Fill in the signup form:")
		fmt.Println("       - Your name: e.g., 'Admin'")
		fmt.Println("       - Email: e.g., 'admin@local.test' (fake, captured by Mailhog)")
		fmt.Println("       - Site name: e.g., 'Customer Feedback'")
		fmt.Println("  3. Open http://localhost:3200 (Mailhog)")
		fmt.Println("  4. Click confirmation link in the email")
		fmt.Println("  5. Repeat for VoS (ports 3101/3201, site: 'Stakeholder Requirements')")
		fmt.Println()
		fmt.Println("To stop and remove:")
		fmt.Println("  portunix pft destroy           # keep data")
		fmt.Println("  portunix pft destroy --volumes # remove everything")
	} else {
		fmt.Println("To deploy manually:")
		fmt.Printf("  cd %s && portunix pft deploy\n", demoPath)
	}
}

func showExampleHelp() {
	fmt.Println("Usage: portunix pft example [options]")
	fmt.Println()
	fmt.Println("Creates a demo with VoC/VoS structure per ISO 16355 QFD.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --path <path>   Directory for demo files (default: ./pft-demo)")
	fmt.Println("  --no-deploy     Create files only, don't deploy containers")
	fmt.Println()
	fmt.Println("This command will:")
	fmt.Println("  1. Create voc/ directory with 3 customer feedback samples")
	fmt.Println("  2. Create vos/ directory with 3 stakeholder requirement samples")
	fmt.Println("  3. Deploy 2x Fider instances:")
	fmt.Println("     - VoC Fider on port 3100 (public, customer-facing)")
	fmt.Println("     - VoS Fider on port 3101 (internal, stakeholders)")
	fmt.Println()
	fmt.Println("VoC = Voice of Customer (public feedback)")
	fmt.Println("VoS = Voice of Stakeholder (internal requirements)")
}

// SampleDocument represents a sample document (VoC or VoS)
type SampleDocument struct {
	Filename string
	Title    string
	Content  string
}

// getVoCSamples returns Voice of Customer sample documents
func getVoCSamples() []SampleDocument {
	return []SampleDocument{
		{
			Filename: "UC001-user-login.md",
			Title:    "User Login",
			Content: `# UC001: User Login

## Summary
As a user, I want to log into the system securely so that I can access my personal dashboard.

## Priority
High

## Status
Open

## Description
Users need a secure authentication mechanism to access protected areas of the application.

### Acceptance Criteria
- [ ] User can enter username/email and password
- [ ] System validates credentials against database
- [ ] Failed attempts are logged and rate-limited
- [ ] Successful login redirects to dashboard
- [ ] Session is created with appropriate timeout
`,
		},
		{
			Filename: "UC002-data-export.md",
			Title:    "Data Export",
			Content: `# UC002: Data Export

## Summary
As a user, I want to export my data in various formats so that I can use it in other applications.

## Priority
Medium

## Status
Open

## Description
Users should be able to export their data for backup, analysis, or migration purposes.

### Acceptance Criteria
- [ ] Support CSV export format
- [ ] Support JSON export format
- [ ] Support PDF report generation
- [ ] Large exports are processed asynchronously
- [ ] User receives notification when export is ready
`,
		},
		{
			Filename: "UC003-notification-preferences.md",
			Title:    "Notification Preferences",
			Content: `# UC003: Notification Preferences

## Summary
As a user, I want to configure my notification preferences so that I only receive relevant alerts.

## Priority
Low

## Status
Open

## Description
Users need granular control over what notifications they receive and through which channels.

### Acceptance Criteria
- [ ] User can enable/disable email notifications
- [ ] User can enable/disable in-app notifications
- [ ] User can set notification frequency (immediate, daily digest, weekly)
- [ ] User can select notification categories
- [ ] Settings are saved and applied immediately
`,
		},
	}
}

// getVoSSamples returns Voice of Stakeholder sample documents
func getVoSSamples() []SampleDocument {
	return []SampleDocument{
		{
			Filename: "REQ001-gdpr-compliance.md",
			Title:    "GDPR Compliance",
			Content: `# REQ001: GDPR Compliance

## Type
Regulatory

## Stakeholder
Legal / Compliance

## Priority
Critical

## Status
Open

## Description
System must comply with GDPR (General Data Protection Regulation) requirements.

### Requirements
- [ ] User consent management for data processing
- [ ] Right to access personal data (data export)
- [ ] Right to erasure (data deletion on request)
- [ ] Data portability support
- [ ] Privacy by design principles
- [ ] Data breach notification procedures
- [ ] DPO (Data Protection Officer) contact information

## Compliance Deadline
Mandatory - must be implemented before launch
`,
		},
		{
			Filename: "REQ002-performance-sla.md",
			Title:    "Performance SLA",
			Content: `# REQ002: Performance SLA Requirements

## Type
Technical / Business

## Stakeholder
Operations / Business

## Priority
High

## Status
Open

## Description
System must meet defined Service Level Agreement performance targets.

### Requirements
- [ ] API response time < 200ms (95th percentile)
- [ ] Page load time < 2 seconds
- [ ] System uptime >= 99.9%
- [ ] Database query time < 100ms
- [ ] Support for 10,000 concurrent users
- [ ] Horizontal scaling capability

## Monitoring
- Real-time performance dashboards
- Automated alerting on SLA violations
- Monthly SLA compliance reports
`,
		},
		{
			Filename: "REQ003-security-audit.md",
			Title:    "Security Audit Requirements",
			Content: `# REQ003: Security Audit Requirements

## Type
Security / Compliance

## Stakeholder
Security Team / CISO

## Priority
Critical

## Status
Open

## Description
System must pass security audit and maintain security certifications.

### Requirements
- [ ] Annual penetration testing
- [ ] OWASP Top 10 vulnerability protection
- [ ] Encrypted data at rest (AES-256)
- [ ] Encrypted data in transit (TLS 1.3)
- [ ] Multi-factor authentication support
- [ ] Role-based access control (RBAC)
- [ ] Audit logging for all sensitive operations
- [ ] Secure key management (HSM or cloud KMS)

## Certifications
- SOC 2 Type II compliance
- ISO 27001 certification (planned)
`,
		},
	}
}

// User command handlers
func handleUserCommand(args []string) {
	if len(args) == 0 {
		showUserHelp()
		return
	}

	// Extract --path from args first
	var configPath string
	var filteredArgs []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--path" && i+1 < len(args) {
			configPath = args[i+1]
			i++ // skip next arg
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	if len(filteredArgs) == 0 {
		showUserHelp()
		return
	}

	subcommand := filteredArgs[0]
	subArgs := filteredArgs[1:]

	// Determine project directory
	var projectDir string
	if configPath != "" {
		projectDir = configPath
	} else {
		projectDir = getProjectDir()
	}

	switch subcommand {
	case "list":
		handleUserListCommand(subArgs, projectDir)
	case "add":
		handleUserAddCommand(subArgs, projectDir)
	case "update":
		handleUserUpdateCommand(subArgs, projectDir)
	case "role":
		handleUserRoleCommand(subArgs, projectDir)
	case "link":
		handleUserLinkCommand(subArgs, projectDir)
	case "remove":
		handleUserRemoveCommand(subArgs, projectDir)
	case "show":
		handleUserShowCommand(subArgs, projectDir)
	case "sync":
		handleUserSyncCommand(subArgs, projectDir)
	case "--help", "-h":
		showUserHelp()
	default:
		fmt.Printf("Unknown user subcommand: %s\n", subcommand)
		showUserHelp()
	}
}

func showUserHelp() {
	fmt.Println("Usage: portunix pft user <command> [options]")
	fmt.Println()
	fmt.Println("User Registry Commands:")
	fmt.Println()
	fmt.Println("  list [--voc|--vos|--vob|--voe]  List users (optionally filter by category)")
	fmt.Println("  add --id <email> --name <name>  Add a new user")
	fmt.Println("  update <id> [--name|--org]      Update user details")
	fmt.Println("  show <id>                       Show user details")
	fmt.Println("  role <id> --voc|--vos|--vob|--voe <role> [--proxy]")
	fmt.Println("                                  Assign role to user in category")
	fmt.Println("  role <id> --voc|--vos|--vob|--voe --remove")
	fmt.Println("                                  Remove role from category")
	fmt.Println("  link <id> --fider <fider-id>    Link user to Fider ID")
	fmt.Println("  remove <id>                     Remove user from registry")
	fmt.Println("  sync [--voc|--vos] [--dry-run]  Sync users from Fider")
	fmt.Println()
	fmt.Println("Options for 'add':")
	fmt.Println("  --id <email>      User ID (typically email)")
	fmt.Println("  --name <name>     User display name")
	fmt.Println("  --org <org>       Organization (optional)")
	fmt.Println()
	fmt.Println("Options for 'update':")
	fmt.Println("  --name <name>     New display name")
	fmt.Println("  --org <org>       New organization")
	fmt.Println()
	fmt.Println("Options for 'sync':")
	fmt.Println("  --voc             Sync only from VoC Fider instance")
	fmt.Println("  --vos             Sync only from VoS Fider instance")
	fmt.Println("  --dry-run         Show what would be synced without changes")
	fmt.Println("  --voc-token <tok> Set VoC Fider API token")
	fmt.Println("  --vos-token <tok> Set VoS Fider API token")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft user add --id user@example.com --name \"John Doe\"")
	fmt.Println("  portunix pft user update user@example.com --name \"Jane Doe\"")
	fmt.Println("  portunix pft user role user@example.com --vos developer")
	fmt.Println("  portunix pft user role user@example.com --vos cio --proxy")
	fmt.Println("  portunix pft user link user@example.com --fider 42")
	fmt.Println("  portunix pft user sync --voc")
}

func handleUserListCommand(args []string, projectDir string) {
	// Parse category filter
	var category string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--voc":
			category = "voc"
		case "--vos":
			category = "vos"
		case "--vob":
			category = "vob"
		case "--voe":
			category = "voe"
		}
	}

	registry, err := LoadUserRegistry(projectDir)
	if err != nil {
		fmt.Printf("Error loading users: %v\n", err)
		return
	}

	var users []User
	if category != "" {
		users = registry.ListUsersByCategory(category)
		fmt.Printf("Users with %s roles:\n\n", GetCategoryName(category))
	} else {
		users = registry.Users
		fmt.Println("All users:")
		fmt.Println()
	}

	PrintUserList(users, category)
}

func handleUserAddCommand(args []string, projectDir string) {
	var id, name, org string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--id":
			if i+1 < len(args) {
				id = args[i+1]
				i++
			}
		case "--name":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		case "--org":
			if i+1 < len(args) {
				org = args[i+1]
				i++
			}
		}
	}

	if id == "" {
		fmt.Println("Error: --id is required")
		fmt.Println("Usage: portunix pft user add --id <email> --name <name>")
		return
	}

	if name == "" {
		fmt.Println("Error: --name is required")
		fmt.Println("Usage: portunix pft user add --id <email> --name <name>")
		return
	}

	registry, err := LoadUserRegistry(projectDir)
	if err != nil {
		fmt.Printf("Error loading users: %v\n", err)
		return
	}

	user := User{
		ID:           id,
		Name:         name,
		Organization: org,
	}

	if err := registry.AddUser(user); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if err := SaveUserRegistry(projectDir, registry); err != nil {
		fmt.Printf("Error saving users: %v\n", err)
		return
	}

	fmt.Printf("✓ User '%s' added successfully\n", id)
}

func handleUserUpdateCommand(args []string, projectDir string) {
	if len(args) == 0 {
		fmt.Println("Usage: portunix pft user update <id> [--name <name>] [--org <org>]")
		return
	}

	id := args[0]
	var name, org string
	var clearOrg bool

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--name":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		case "--org":
			if i+1 < len(args) {
				org = args[i+1]
				i++
			}
		case "--org=":
			clearOrg = true
		}
	}

	if name == "" && org == "" && !clearOrg {
		fmt.Println("Error: at least one of --name or --org is required")
		fmt.Println("Usage: portunix pft user update <id> [--name <name>] [--org <org>]")
		return
	}

	registry, err := LoadUserRegistry(projectDir)
	if err != nil {
		fmt.Printf("Error loading users: %v\n", err)
		return
	}

	user := registry.FindUser(id)
	if user == nil {
		fmt.Printf("User '%s' not found\n", id)
		return
	}

	// Apply updates
	updated := false
	if name != "" {
		user.Name = name
		updated = true
		fmt.Printf("  Name updated to: %s\n", name)
	}
	if org != "" {
		user.Organization = org
		updated = true
		fmt.Printf("  Organization updated to: %s\n", org)
	}
	if clearOrg {
		user.Organization = ""
		updated = true
		fmt.Println("  Organization cleared")
	}

	if updated {
		user.UpdatedAt = time.Now()
		if err := SaveUserRegistry(projectDir, registry); err != nil {
			fmt.Printf("Error saving users: %v\n", err)
			return
		}
		fmt.Printf("✓ User '%s' updated successfully\n", id)
	}
}

func handleUserShowCommand(args []string, projectDir string) {
	if len(args) == 0 {
		fmt.Println("Usage: portunix pft user show <id>")
		return
	}

	id := args[0]

	registry, err := LoadUserRegistry(projectDir)
	if err != nil {
		fmt.Printf("Error loading users: %v\n", err)
		return
	}

	user := registry.FindUser(id)
	if user == nil {
		fmt.Printf("User '%s' not found\n", id)
		return
	}

	PrintUser(user)
}

func handleUserRoleCommand(args []string, projectDir string) {
	if len(args) == 0 {
		fmt.Println("Usage: portunix pft user role <id> --voc|--vos|--vob|--voe <role> [--proxy]")
		return
	}

	id := args[0]
	var category, role string
	var proxy, remove bool

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--voc":
			category = "voc"
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				role = args[i+1]
				i++
			}
		case "--vos":
			category = "vos"
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				role = args[i+1]
				i++
			}
		case "--vob":
			category = "vob"
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				role = args[i+1]
				i++
			}
		case "--voe":
			category = "voe"
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				role = args[i+1]
				i++
			}
		case "--proxy":
			proxy = true
		case "--remove":
			remove = true
		}
	}

	if category == "" {
		fmt.Println("Error: category required (--voc, --vos, --vob, or --voe)")
		return
	}

	registry, err := LoadUserRegistry(projectDir)
	if err != nil {
		fmt.Printf("Error loading users: %v\n", err)
		return
	}

	user := registry.FindUser(id)
	if user == nil {
		fmt.Printf("User '%s' not found\n", id)
		return
	}

	if remove {
		if err := user.RemoveRole(category); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("✓ Removed %s role from user '%s'\n", GetCategoryName(category), id)
	} else {
		if role == "" {
			fmt.Println("Error: role name required")
			fmt.Println("Usage: portunix pft user role <id> --vos <role> [--proxy]")
			return
		}

		// Validate role
		valid, err := ValidateRole(projectDir, category, role)
		if err != nil {
			fmt.Printf("Error validating role: %v\n", err)
			return
		}
		if !valid {
			fmt.Printf("Error: role '%s' is not valid for category '%s'\n", role, category)
			fmt.Printf("Run 'portunix pft role list --%s' to see available roles\n", category)
			return
		}

		if err := user.SetRole(category, role, proxy); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		proxyStr := ""
		if proxy {
			proxyStr = " (proxy)"
		}
		fmt.Printf("✓ Assigned %s role '%s'%s to user '%s'\n", GetCategoryName(category), role, proxyStr, id)
	}

	if err := SaveUserRegistry(projectDir, registry); err != nil {
		fmt.Printf("Error saving users: %v\n", err)
		return
	}
}

func handleUserLinkCommand(args []string, projectDir string) {
	if len(args) == 0 {
		fmt.Println("Usage: portunix pft user link <id> --fider <fider-id>")
		return
	}

	id := args[0]
	var fiderID int

	for i := 1; i < len(args); i++ {
		if args[i] == "--fider" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &fiderID)
			i++
		}
	}

	if fiderID == 0 {
		fmt.Println("Error: --fider <id> is required")
		return
	}

	registry, err := LoadUserRegistry(projectDir)
	if err != nil {
		fmt.Printf("Error loading users: %v\n", err)
		return
	}

	user := registry.FindUser(id)
	if user == nil {
		fmt.Printf("User '%s' not found\n", id)
		return
	}

	user.LinkFider(fiderID)

	if err := SaveUserRegistry(projectDir, registry); err != nil {
		fmt.Printf("Error saving users: %v\n", err)
		return
	}

	fmt.Printf("✓ Linked user '%s' to Fider ID %d\n", id, fiderID)
}

func handleUserRemoveCommand(args []string, projectDir string) {
	if len(args) == 0 {
		fmt.Println("Usage: portunix pft user remove <id>")
		return
	}

	id := args[0]

	registry, err := LoadUserRegistry(projectDir)
	if err != nil {
		fmt.Printf("Error loading users: %v\n", err)
		return
	}

	if err := registry.RemoveUser(id); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if err := SaveUserRegistry(projectDir, registry); err != nil {
		fmt.Printf("Error saving users: %v\n", err)
		return
	}

	fmt.Printf("✓ User '%s' removed\n", id)
}

func handleUserSyncCommand(args []string, projectDir string) {
	// Parse flags
	var syncVoC, syncVoS, dryRun bool
	var vocToken, vosToken string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--voc":
			syncVoC = true
		case "--vos":
			syncVoS = true
		case "--dry-run":
			dryRun = true
		case "--voc-token":
			if i+1 < len(args) {
				vocToken = args[i+1]
				i++
			}
		case "--vos-token":
			if i+1 < len(args) {
				vosToken = args[i+1]
				i++
			}
		case "--help", "-h":
			showUserSyncHelp()
			return
		}
	}

	// If neither specified, sync both
	if !syncVoC && !syncVoS {
		syncVoC = true
		syncVoS = true
	}

	config, err := LoadConfig()
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Update config with tokens if provided
	if vocToken != "" {
		config.VoC.APIToken = vocToken
	}
	if vosToken != "" {
		config.VoS.APIToken = vosToken
	}

	// Load user registry
	registry, err := LoadUserRegistry(projectDir)
	if err != nil {
		fmt.Printf("Error loading user registry: %v\n", err)
		return
	}

	fmt.Println("Synchronizing users from Fider...")
	if dryRun {
		fmt.Println("(dry-run mode - no changes will be made)")
	}
	fmt.Println()

	totalAdded := 0
	totalUpdated := 0
	totalSkipped := 0

	// Sync VoC users
	if syncVoC {
		fmt.Println("👥 VoC (Voice of Customer) users:")

		vocURL := config.VoC.URL
		if vocURL == "" {
			vocURL = "http://localhost:3100"
		}
		vocAPIToken := config.VoC.APIToken
		if vocAPIToken == "" {
			vocAPIToken = config.GetAPIToken()
		}

		if vocAPIToken == "" {
			fmt.Println("   ✗ No API token configured for VoC")
			fmt.Println("   Run: portunix pft user sync --voc --voc-token <your-token>")
		} else {
			client := NewFiderClient(vocURL, vocAPIToken)
			added, updated, skipped, err := syncUsersFromFider(client, registry, "voc", dryRun)
			if err != nil {
				fmt.Printf("   ✗ Sync failed: %v\n", err)
			} else {
				fmt.Printf("   Added: %d, Updated: %d, Skipped: %d\n", added, updated, skipped)
				totalAdded += added
				totalUpdated += updated
				totalSkipped += skipped
			}
		}
		fmt.Println()
	}

	// Sync VoS users
	if syncVoS {
		fmt.Println("👥 VoS (Voice of Stakeholder) users:")

		vosURL := config.VoS.URL
		if vosURL == "" {
			vosURL = "http://localhost:3101"
		}
		vosAPIToken := config.VoS.APIToken
		if vosAPIToken == "" {
			vosAPIToken = config.GetAPIToken()
		}

		if vosAPIToken == "" {
			fmt.Println("   ✗ No API token configured for VoS")
			fmt.Println("   Run: portunix pft user sync --vos --vos-token <your-token>")
		} else {
			client := NewFiderClient(vosURL, vosAPIToken)
			added, updated, skipped, err := syncUsersFromFider(client, registry, "vos", dryRun)
			if err != nil {
				fmt.Printf("   ✗ Sync failed: %v\n", err)
			} else {
				fmt.Printf("   Added: %d, Updated: %d, Skipped: %d\n", added, updated, skipped)
				totalAdded += added
				totalUpdated += updated
				totalSkipped += skipped
			}
		}
		fmt.Println()
	}

	// Save registry if not dry-run
	if !dryRun && (totalAdded > 0 || totalUpdated > 0) {
		if err := SaveUserRegistry(projectDir, registry); err != nil {
			fmt.Printf("Error saving user registry: %v\n", err)
			return
		}
	}

	// Save updated config with tokens if they were provided
	if vocToken != "" || vosToken != "" {
		configPath, _ := findConfigFile()
		if configPath != "" {
			config.SaveToPath(configPath)
			fmt.Println("Configuration updated with API tokens.")
		}
	}

	fmt.Printf("User sync complete. Added: %d, Updated: %d, Skipped: %d\n", totalAdded, totalUpdated, totalSkipped)
}

// syncUsersFromFider fetches users from Fider and syncs them to local registry
func syncUsersFromFider(client *FiderClient, registry *UserRegistry, category string, dryRun bool) (added, updated, skipped int, err error) {
	// Fetch users from Fider
	fiderUsers, err := client.ListUsers()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to fetch users from Fider: %w", err)
	}

	fmt.Printf("   Found %d users in Fider\n", len(fiderUsers))

	for _, fiderUser := range fiderUsers {
		// Skip users without email
		if fiderUser.Email == "" {
			fmt.Printf("   ⚠ Skipping user '%s' (no email)\n", fiderUser.Name)
			skipped++
			continue
		}

		// Try to find existing user by Fider ID first
		existingUser := registry.FindUserByFiderID(fiderUser.ID)
		if existingUser != nil {
			// User already linked by Fider ID
			fmt.Printf("   ⏭ %s (already linked via Fider ID %d)\n", fiderUser.Email, fiderUser.ID)
			skipped++
			continue
		}

		// Try to find by email
		existingUser = registry.FindUserByEmail(fiderUser.Email)
		if existingUser != nil {
			// User exists, update Fider ID link
			if dryRun {
				fmt.Printf("   🔗 Would link: %s → Fider ID %d\n", fiderUser.Email, fiderUser.ID)
			} else {
				existingUser.LinkFider(fiderUser.ID)
				fmt.Printf("   🔗 Linked: %s → Fider ID %d\n", fiderUser.Email, fiderUser.ID)
			}
			updated++
			continue
		}

		// New user - add to registry
		if dryRun {
			fmt.Printf("   ➕ Would add: %s (%s) with Fider ID %d\n", fiderUser.Email, fiderUser.Name, fiderUser.ID)
		} else {
			newUser := User{
				ID:   fiderUser.Email,
				Name: fiderUser.Name,
				ExternalIDs: &ExternalIDs{
					Fider: fiderUser.ID,
				},
			}

			// Assign default role based on category
			switch category {
			case "voc":
				newUser.Roles.VoC = &RoleAssignment{Role: "customer", Proxy: false}
			case "vos":
				newUser.Roles.VoS = &RoleAssignment{Role: "developer", Proxy: false}
			}

			if err := registry.AddUser(newUser); err != nil {
				fmt.Printf("   ✗ Failed to add %s: %v\n", fiderUser.Email, err)
				continue
			}
			fmt.Printf("   ➕ Added: %s (%s) with Fider ID %d\n", fiderUser.Email, fiderUser.Name, fiderUser.ID)
		}
		added++
	}

	return added, updated, skipped, nil
}

func showUserSyncHelp() {
	fmt.Println("Usage: portunix pft user sync [options]")
	fmt.Println()
	fmt.Println("Synchronize users from Fider instances to local registry.")
	fmt.Println()
	fmt.Println("This command will:")
	fmt.Println("  1. Fetch all users from Fider API")
	fmt.Println("  2. Add new users to local registry")
	fmt.Println("  3. Link existing users by email to their Fider IDs")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --voc              Sync only from VoC Fider instance")
	fmt.Println("  --vos              Sync only from VoS Fider instance")
	fmt.Println("  --voc-token <tok>  Set VoC Fider API token")
	fmt.Println("  --vos-token <tok>  Set VoS Fider API token")
	fmt.Println("  --dry-run          Show what would be synced without making changes")
	fmt.Println()
	fmt.Println("New users are assigned default roles:")
	fmt.Println("  VoC: customer")
	fmt.Println("  VoS: developer")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft user sync --voc")
	fmt.Println("  portunix pft user sync --dry-run")
	fmt.Println("  portunix pft user sync --voc --voc-token abc123")
}

// Role command handlers
func handleRoleListCommand(args []string) {
	if len(args) == 0 {
		showRoleHelp()
		return
	}

	// Extract --path from args first
	var configPath string
	var filteredArgs []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--path" && i+1 < len(args) {
			configPath = args[i+1]
			i++ // skip next arg
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	if len(filteredArgs) == 0 {
		showRoleHelp()
		return
	}

	subcommand := filteredArgs[0]
	subArgs := filteredArgs[1:]

	// Determine project directory
	var projectDir string
	if configPath != "" {
		projectDir = configPath
	} else {
		projectDir = getProjectDir()
	}

	switch subcommand {
	case "list":
		handleRoleListSubCommand(subArgs, projectDir)
	case "init":
		handleRoleInitCommand(projectDir)
	case "--help", "-h":
		showRoleHelp()
	default:
		// Treat as category flag for list
		handleRoleListSubCommand(filteredArgs, projectDir)
	}
}

func showRoleHelp() {
	fmt.Println("Usage: portunix pft role <command> [options]")
	fmt.Println()
	fmt.Println("Role Management Commands:")
	fmt.Println()
	fmt.Println("  list --voc|--vos|--vob|--voe  List roles for category")
	fmt.Println("  init                          Initialize default role files")
	fmt.Println()
	fmt.Println("Categories:")
	fmt.Println("  --voc    Voice of Customer (customer roles)")
	fmt.Println("  --vos    Voice of Stakeholder (internal stakeholder roles)")
	fmt.Println("  --vob    Voice of Business (business roles)")
	fmt.Println("  --voe    Voice of Engineer (engineering roles)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft role list --vos")
	fmt.Println("  portunix pft role init")
}

func handleRoleListSubCommand(args []string, projectDir string) {
	var category string
	for _, arg := range args {
		switch arg {
		case "--voc":
			category = "voc"
		case "--vos":
			category = "vos"
		case "--vob":
			category = "vob"
		case "--voe":
			category = "voe"
		}
	}

	if category == "" {
		fmt.Println("Error: category required (--voc, --vos, --vob, or --voe)")
		fmt.Println()
		showRoleHelp()
		return
	}

	roles, err := LoadRoles(projectDir, category)
	if err != nil {
		fmt.Printf("Error loading roles: %v\n", err)
		return
	}

	fmt.Printf("%s Roles:\n\n", GetCategoryName(category))
	PrintRoles(roles)
}

func handleRoleInitCommand(projectDir string) {
	fmt.Println("Initializing default role files...")
	fmt.Println()

	if err := InitializeRoles(projectDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("✓ Role files initialized")
}

func getProjectDir() string {
	config, configFilePath, err := LoadConfigWithFilePath()
	if err == nil {
		return ResolveProjectPath(config, configFilePath, "")
	}
	cwd, _ := os.Getwd()
	return cwd
}

// Category command handlers
func handleCategoryCommand(args []string) {
	if len(args) == 0 {
		showCategoryHelp()
		return
	}

	// Extract --path from args first
	var configPath string
	var filteredArgs []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--path" && i+1 < len(args) {
			configPath = args[i+1]
			i++ // skip next arg
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	if len(filteredArgs) == 0 {
		showCategoryHelp()
		return
	}

	subcommand := filteredArgs[0]
	subArgs := filteredArgs[1:]

	// Determine project directory
	var projectDir string
	if configPath != "" {
		projectDir = configPath
	} else {
		projectDir = getProjectDir()
	}

	switch subcommand {
	case "list":
		handleCategoryListCommand(subArgs, projectDir)
	case "add":
		handleCategoryAddCommand(subArgs, projectDir)
	case "remove":
		handleCategoryRemoveCommand(subArgs, projectDir)
	case "rename":
		handleCategoryRenameCommand(subArgs, projectDir)
	case "show":
		handleCategoryShowCommand(subArgs, projectDir)
	case "--help", "-h":
		showCategoryHelp()
	default:
		fmt.Printf("Unknown category subcommand: %s\n", subcommand)
		showCategoryHelp()
	}
}

func showCategoryHelp() {
	fmt.Println("Usage: portunix pft category <command> [options]")
	fmt.Println()
	fmt.Println("Category Management Commands:")
	fmt.Println()
	fmt.Println("  list [--area <area>]              List categories")
	fmt.Println("  add <id> --name <name> --area <area>")
	fmt.Println("                                    Create new category")
	fmt.Println("  remove <id> --area <area> [--force]")
	fmt.Println("                                    Delete category")
	fmt.Println("  rename <id> --name <name> --area <area>")
	fmt.Println("                                    Rename category")
	fmt.Println("  show <id> --area <area>           Show category details")
	fmt.Println()
	fmt.Println("Areas:")
	fmt.Println("  voc    Voice of Customer")
	fmt.Println("  vos    Voice of Stakeholder")
	fmt.Println("  vob    Voice of Business")
	fmt.Println("  voe    Voice of Engineer")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --name <name>         Category display name")
	fmt.Println("  --description <desc>  Category description")
	fmt.Println("  --color <hex>         Category color (e.g., #3B82F6)")
	fmt.Println("  --force               Force removal even if items assigned")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft category list --area voc")
	fmt.Println("  portunix pft category add user-auth --name \"User Authentication\" --area voc")
	fmt.Println("  portunix pft category remove user-auth --area voc")
}

func handleCategoryListCommand(args []string, projectDir string) {
	var area string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--area":
			if i+1 < len(args) {
				area = args[i+1]
				i++
			}
		}
	}

	// If no area specified, list all areas
	if area == "" {
		for _, a := range ValidAreaNames {
			printCategoriesForArea(projectDir, a)
		}
		return
	}

	if !IsValidArea(area) {
		fmt.Printf("Error: invalid area '%s' (valid: %s)\n", area, strings.Join(ValidAreaNames, ", "))
		return
	}

	printCategoriesForArea(projectDir, area)
}

func printCategoriesForArea(projectDir, area string) {
	cats, err := GetAllCategoriesWithCounts(projectDir, area)
	if err != nil {
		fmt.Printf("Error loading categories for %s: %v\n", area, err)
		return
	}

	areaNames := map[string]string{
		"voc": "Voice of Customer",
		"vos": "Voice of Stakeholder",
		"vob": "Voice of Business",
		"voe": "Voice of Engineer",
	}

	fmt.Printf("\n📁 %s (%s)\n", areaNames[area], area)
	fmt.Println(strings.Repeat("-", 50))

	if len(cats) == 0 {
		fmt.Println("   (no categories)")
		return
	}

	fmt.Printf("   %-20s %-25s %s\n", "ID", "NAME", "ITEMS")
	fmt.Println(strings.Repeat("-", 50))
	for _, cat := range cats {
		color := ""
		if cat.Color != "" {
			color = " " + cat.Color
		}
		fmt.Printf("   %-20s %-25s %d%s\n", cat.ID, truncateStr(cat.Name, 25), cat.Count, color)
	}
}

func handleCategoryAddCommand(args []string, projectDir string) {
	if len(args) == 0 {
		fmt.Println("Error: category ID required")
		fmt.Println("Usage: portunix pft category add <id> --name <name> --area <area>")
		return
	}

	categoryID := args[0]
	var name, description, color, area string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--name":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		case "--description":
			if i+1 < len(args) {
				description = args[i+1]
				i++
			}
		case "--color":
			if i+1 < len(args) {
				color = args[i+1]
				i++
			}
		case "--area":
			if i+1 < len(args) {
				area = args[i+1]
				i++
			}
		}
	}

	if area == "" {
		fmt.Println("Error: --area is required")
		return
	}
	if name == "" {
		fmt.Println("Error: --name is required")
		return
	}

	registry, err := LoadCategoryRegistry(projectDir, area)
	if err != nil {
		fmt.Printf("Error loading categories: %v\n", err)
		return
	}

	cat := Category{
		ID:          categoryID,
		Name:        name,
		Description: description,
		Color:       color,
	}

	if err := registry.AddCategory(cat); err != nil {
		fmt.Printf("Error adding category: %v\n", err)
		return
	}

	if err := SaveCategoryRegistry(projectDir, area, registry); err != nil {
		fmt.Printf("Error saving categories: %v\n", err)
		return
	}

	fmt.Printf("✓ Category '%s' added to %s\n", NormalizeCategoryID(categoryID), area)
}

func handleCategoryRemoveCommand(args []string, projectDir string) {
	if len(args) == 0 {
		fmt.Println("Error: category ID required")
		fmt.Println("Usage: portunix pft category remove <id> --area <area> [--force]")
		return
	}

	categoryID := args[0]
	var area string
	var force bool

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--area":
			if i+1 < len(args) {
				area = args[i+1]
				i++
			}
		case "--force":
			force = true
		}
	}

	if area == "" {
		fmt.Println("Error: --area is required")
		return
	}

	// Check if category has items assigned
	count, err := CountItemsInCategory(projectDir, area, categoryID)
	if err != nil {
		fmt.Printf("Error counting items: %v\n", err)
		return
	}

	if count > 0 && !force {
		fmt.Printf("Error: category '%s' has %d items assigned\n", categoryID, count)
		fmt.Println("Use --force to remove anyway (items will become uncategorized)")
		return
	}

	registry, err := LoadCategoryRegistry(projectDir, area)
	if err != nil {
		fmt.Printf("Error loading categories: %v\n", err)
		return
	}

	if err := registry.RemoveCategory(categoryID); err != nil {
		fmt.Printf("Error removing category: %v\n", err)
		return
	}

	if err := SaveCategoryRegistry(projectDir, area, registry); err != nil {
		fmt.Printf("Error saving categories: %v\n", err)
		return
	}

	fmt.Printf("✓ Category '%s' removed from %s\n", NormalizeCategoryID(categoryID), area)
	if count > 0 {
		fmt.Printf("  Note: %d items are now uncategorized\n", count)
	}
}

func handleCategoryRenameCommand(args []string, projectDir string) {
	if len(args) == 0 {
		fmt.Println("Error: category ID required")
		fmt.Println("Usage: portunix pft category rename <id> --name <name> --area <area>")
		return
	}

	categoryID := args[0]
	var name, description, color, area string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--name":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		case "--description":
			if i+1 < len(args) {
				description = args[i+1]
				i++
			}
		case "--color":
			if i+1 < len(args) {
				color = args[i+1]
				i++
			}
		case "--area":
			if i+1 < len(args) {
				area = args[i+1]
				i++
			}
		}
	}

	if area == "" {
		fmt.Println("Error: --area is required")
		return
	}

	registry, err := LoadCategoryRegistry(projectDir, area)
	if err != nil {
		fmt.Printf("Error loading categories: %v\n", err)
		return
	}

	updates := Category{
		Name:        name,
		Description: description,
		Color:       color,
	}

	if err := registry.UpdateCategory(categoryID, updates); err != nil {
		fmt.Printf("Error updating category: %v\n", err)
		return
	}

	if err := SaveCategoryRegistry(projectDir, area, registry); err != nil {
		fmt.Printf("Error saving categories: %v\n", err)
		return
	}

	fmt.Printf("✓ Category '%s' updated in %s\n", NormalizeCategoryID(categoryID), area)
}

func handleCategoryShowCommand(args []string, projectDir string) {
	if len(args) == 0 {
		fmt.Println("Error: category ID required")
		fmt.Println("Usage: portunix pft category show <id> --area <area>")
		return
	}

	categoryID := args[0]
	var area string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--area":
			if i+1 < len(args) {
				area = args[i+1]
				i++
			}
		}
	}

	if area == "" {
		fmt.Println("Error: --area is required")
		return
	}

	registry, err := LoadCategoryRegistry(projectDir, area)
	if err != nil {
		fmt.Printf("Error loading categories: %v\n", err)
		return
	}

	cat, err := registry.GetCategory(categoryID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	count, _ := CountItemsInCategory(projectDir, area, categoryID)

	fmt.Printf("Category: %s\n", cat.ID)
	fmt.Println(strings.Repeat("-", 30))
	fmt.Printf("  Name: %s\n", cat.Name)
	if cat.Description != "" {
		fmt.Printf("  Description: %s\n", cat.Description)
	}
	if cat.Color != "" {
		fmt.Printf("  Color: %s\n", cat.Color)
	}
	fmt.Printf("  Order: %d\n", cat.Order)
	fmt.Printf("  Items: %d\n", count)
	fmt.Printf("  Created: %s\n", cat.CreatedAt)
	fmt.Printf("  Updated: %s\n", cat.UpdatedAt)
}

// Assign/Unassign command handlers
func handleAssignCommand(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		showAssignHelp()
		return
	}

	// Parse arguments - first non-flag argument is itemID
	var itemID string
	var categoryID string
	var configPath string
	var setMode bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--category", "-c":
			if i+1 < len(args) {
				categoryID = args[i+1]
				i++
			}
		case "--path":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case "--set", "-s":
			setMode = true
		case "--help", "-h":
			showAssignHelp()
			return
		default:
			if !strings.HasPrefix(args[i], "-") && itemID == "" {
				itemID = args[i]
			}
		}
	}

	if itemID == "" {
		fmt.Println("Error: item ID is required")
		showAssignHelp()
		return
	}

	if categoryID == "" {
		fmt.Println("Error: --category is required")
		showAssignHelp()
		return
	}

	config, configFilePath, err := loadOrCreateConfig(configPath)
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Use cross-platform path resolution
	projectDir := ResolveProjectPath(config, configFilePath, configPath)

	// Find the item file
	filePath, feedbackType, err := findFeedbackItemFile(projectDir, itemID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Verify category exists in the area
	registry, err := LoadCategoryRegistry(projectDir, feedbackType)
	if err != nil {
		fmt.Printf("Error loading categories: %v\n", err)
		return
	}

	if !registry.HasCategory(categoryID) {
		fmt.Printf("Error: category '%s' not found in %s\n", categoryID, feedbackType)
		fmt.Println("Use 'portunix pft category list --area " + feedbackType + "' to see available categories")
		return
	}

	// Set or add category to file
	if setMode {
		// Replace all categories with the new one
		if err := SetCategoryToFile(filePath, categoryID); err != nil {
			fmt.Printf("Error setting category: %v\n", err)
			return
		}
		fmt.Printf("✓ Set category '%s' to %s (replaced all previous)\n", categoryID, itemID)
	} else {
		// Add category to existing ones
		if err := AddCategoryToFile(filePath, categoryID); err != nil {
			fmt.Printf("Error assigning category: %v\n", err)
			return
		}
		fmt.Printf("✓ Assigned category '%s' to %s\n", categoryID, itemID)
	}
}

func handleUnassignCommand(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		showUnassignHelp()
		return
	}

	// Parse arguments - first non-flag argument is itemID
	var itemID string
	var categoryID string
	var configPath string
	var removeAll bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--category", "-c":
			if i+1 < len(args) {
				categoryID = args[i+1]
				i++
			}
		case "--path":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case "--all":
			removeAll = true
		case "--help", "-h":
			showUnassignHelp()
			return
		default:
			if !strings.HasPrefix(args[i], "-") && itemID == "" {
				itemID = args[i]
			}
		}
	}

	if itemID == "" {
		fmt.Println("Error: item ID is required")
		showUnassignHelp()
		return
	}

	if categoryID == "" && !removeAll {
		fmt.Println("Error: --category or --all is required")
		showUnassignHelp()
		return
	}

	config, configFilePath, err := loadOrCreateConfig(configPath)
	if err != nil {
		fmt.Println("No configuration found. Run 'portunix pft configure' first.")
		return
	}

	// Use cross-platform path resolution
	projectDir := ResolveProjectPath(config, configFilePath, configPath)

	// Find the item file
	filePath, _, err := findFeedbackItemFile(projectDir, itemID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if removeAll {
		if err := ClearCategoriesFromFile(filePath); err != nil {
			fmt.Printf("Error removing categories: %v\n", err)
			return
		}
		fmt.Printf("✓ Removed all categories from %s\n", itemID)
	} else {
		if err := RemoveCategoryFromFile(filePath, categoryID); err != nil {
			fmt.Printf("Error removing category: %v\n", err)
			return
		}
		fmt.Printf("✓ Removed category '%s' from %s\n", categoryID, itemID)
	}
}

func showAssignHelp() {
	fmt.Println("Usage: portunix pft assign <item-id> --category <category-id> [options]")
	fmt.Println()
	fmt.Println("Assign a category to a feedback item.")
	fmt.Println("Items can have multiple categories (0..N).")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --category, -c <id>  Category ID to assign")
	fmt.Println("  --set, -s            Replace all categories (instead of adding)")
	fmt.Println("  --path <dir>         Path to PFT project directory")
	fmt.Println("  --help, -h           Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft assign UC001 --category user-auth        # Add category")
	fmt.Println("  portunix pft assign UC001 --category A --set          # Replace all with A")
	fmt.Println("  portunix pft assign P01 -c A -s --path docs/project   # Replace with path")
}

func showUnassignHelp() {
	fmt.Println("Usage: portunix pft unassign <item-id> --category <category-id> [options]")
	fmt.Println("       portunix pft unassign <item-id> --all [options]")
	fmt.Println()
	fmt.Println("Remove category(s) from a feedback item.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --category, -c <id>  Category ID to remove")
	fmt.Println("  --all                Remove all categories")
	fmt.Println("  --path <dir>         Path to PFT project directory")
	fmt.Println("  --help, -h           Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft unassign UC001 --category user-auth")
	fmt.Println("  portunix pft unassign UC001 --all")
	fmt.Println("  portunix pft unassign P01 --category A --path docs/pft-project")
}

// findFeedbackItemFile finds the file path for a feedback item by ID
func findFeedbackItemFile(projectDir, itemID string) (string, string, error) {
	// Search in all areas using getVoiceDir for proper case handling
	for _, area := range ValidAreaNames {
		areaDir := getVoiceDir(projectDir, area)
		// Use recursive ScanFeedbackDirectory to find items in subdirectories (needs/, verbatims/, etc.)
		items, err := ScanFeedbackDirectory(areaDir, area)
		if err != nil {
			continue
		}

		for _, item := range items {
			// Match by ID from frontmatter or by filename prefix
			filename := filepath.Base(item.FilePath)
			name := strings.TrimSuffix(filename, ".md")
			if item.ID == itemID || strings.HasPrefix(name, itemID+"-") || name == itemID {
				return item.FilePath, area, nil
			}
		}
	}

	return "", "", fmt.Errorf("feedback item '%s' not found", itemID)
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Show version")
	rootCmd.Flags().Bool("description", false, "Show description")
	rootCmd.Flags().Bool("list-commands", false, "List available commands")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
