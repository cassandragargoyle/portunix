package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"portunix.ai/app"
	"portunix.ai/app/github"
	"portunix.ai/app/plugins"
	"portunix.ai/app/plugins/manager"
)

var pluginManager *manager.Manager

// AvailablePluginInfo represents a plugin available for installation
type AvailablePluginInfo struct {
	Name     string
	Platform string
	Format   string
	Size     int64
	AssetName string
}

// pluginCmd represents the plugin command
var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Plugin management commands",
	Long: `Plugin management commands for Portunix. 

The plugin system allows you to extend Portunix functionality through
independent plugins using gRPC communication.

Examples:
  portunix plugin list                    # List all plugins
  portunix plugin install ./my-plugin/   # Install plugin from directory
  portunix plugin enable my-plugin       # Enable plugin
  portunix plugin start my-plugin        # Start plugin`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializePluginManager()
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if pluginManager != nil {
			return pluginManager.Shutdown()
		}
		return nil
	},
}

// Plugin list command
var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed plugins",
	Long:  `List all installed plugins with their status and information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		showAll, _ := cmd.Flags().GetBool("all")
		outputFormat, _ := cmd.Flags().GetString("output")

		return listPlugins(showAll, outputFormat)
	},
}

// Plugin install command
var pluginInstallCmd = &cobra.Command{
	Use:   "install <plugin-path>",
	Short: "Install a plugin",
	Long: `Install a plugin from a directory containing plugin.yaml manifest.

The plugin directory must contain:
- plugin.yaml (manifest file)
- Plugin binary
- Any additional plugin files

Example:
  portunix plugin install ./agile-plugin/`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginPath := args[0]
		return installPlugin(pluginPath)
	},
}

// Plugin uninstall command
var pluginUninstallCmd = &cobra.Command{
	Use:   "uninstall <plugin-name>",
	Short: "Uninstall a plugin",
	Long:  `Uninstall a plugin and remove all its files.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		force, _ := cmd.Flags().GetBool("force")
		return uninstallPlugin(pluginName, force)
	},
}

// Plugin enable command
var pluginEnableCmd = &cobra.Command{
	Use:   "enable <plugin-name>",
	Short: "Enable a plugin",
	Long:  `Enable a plugin to make it available for use.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		return enablePlugin(pluginName)
	},
}

// Plugin disable command
var pluginDisableCmd = &cobra.Command{
	Use:   "disable <plugin-name>",
	Short: "Disable a plugin",
	Long:  `Disable a plugin and stop it if running.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		return disablePlugin(pluginName)
	},
}

// Plugin start command
var pluginStartCmd = &cobra.Command{
	Use:   "start <plugin-name>",
	Short: "Start a plugin",
	Long:  `Start a plugin service.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		return startPlugin(pluginName)
	},
}

// Plugin stop command
var pluginStopCmd = &cobra.Command{
	Use:   "stop <plugin-name>",
	Short: "Stop a plugin",
	Long:  `Stop a plugin service.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		return stopPlugin(pluginName)
	},
}

// Plugin info command
var pluginInfoCmd = &cobra.Command{
	Use:   "info <plugin-name>",
	Short: "Show plugin information",
	Long:  `Show detailed information about a specific plugin.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		return showPluginInfo(pluginName)
	},
}

// Plugin health command
var pluginHealthCmd = &cobra.Command{
	Use:   "health <plugin-name>",
	Short: "Check plugin health",
	Long:  `Check the health status of a running plugin.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		return checkPluginHealth(pluginName)
	},
}

// Plugin create command
var pluginCreateCmd = &cobra.Command{
	Use:   "create <plugin-name>",
	Short: "Create a new plugin template",
	Long: `Create a new plugin template with basic structure and manifest.

This creates a directory with:
- plugin.yaml (manifest file)
- Basic plugin structure
- Example implementation

Example:
  portunix plugin create my-plugin`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		author, _ := cmd.Flags().GetString("author")
		description, _ := cmd.Flags().GetString("description")
		outputDir, _ := cmd.Flags().GetString("output")

		return createPluginTemplate(pluginName, author, description, outputDir)
	},
}

// Plugin validate command
var pluginValidateCmd = &cobra.Command{
	Use:   "validate <plugin-path>",
	Short: "Validate a plugin",
	Long:  `Validate a plugin manifest and structure.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginPath := args[0]
		return validatePlugin(pluginPath)
	},
}

// Plugin list available command
var pluginListAvailableCmd = &cobra.Command{
	Use:   "list-available",
	Short: "List available plugins from the official repository",
	Long: `List all available plugins from the official Portunix plugins repository.
	
This command fetches the latest release from cassandragargoyle/portunix-plugins
and shows all plugins available for installation.

Example:
  portunix plugin list-available`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listAvailablePlugins()
	},
}

// Plugin install from GitHub command
var pluginInstallGitHubCmd = &cobra.Command{
	Use:   "install-github <plugin-name>",
	Short: "Install a plugin from GitHub",
	Long: `Install a plugin from GitHub release assets.
	
This command downloads the latest release from the Portunix plugins repository
or a custom repository if specified with owner/repo format.

Examples:
  portunix plugin install-github agile-software-development
  portunix plugin install-github custom-org/my-plugin --version v1.2.0
  portunix plugin install-github agile-software-development --version v1.0.0`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		version, _ := cmd.Flags().GetString("version")
		force, _ := cmd.Flags().GetBool("force")
		return installPluginFromGitHub(pluginName, version, force)
	},
}

func init() {
	rootCmd.AddCommand(pluginCmd)

	// Add subcommands
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginListAvailableCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginCmd.AddCommand(pluginInstallGitHubCmd)
	pluginCmd.AddCommand(pluginUninstallCmd)
	pluginCmd.AddCommand(pluginEnableCmd)
	pluginCmd.AddCommand(pluginDisableCmd)
	pluginCmd.AddCommand(pluginStartCmd)
	pluginCmd.AddCommand(pluginStopCmd)
	pluginCmd.AddCommand(pluginInfoCmd)
	pluginCmd.AddCommand(pluginHealthCmd)
	pluginCmd.AddCommand(pluginCreateCmd)
	pluginCmd.AddCommand(pluginValidateCmd)

	// Flags for list command
	pluginListCmd.Flags().BoolP("all", "a", false, "Show all plugins (including disabled)")
	pluginListCmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")

	// Flags for uninstall command
	pluginUninstallCmd.Flags().BoolP("force", "f", false, "Force uninstall without confirmation")

	// Flags for create command
	pluginCreateCmd.Flags().StringP("author", "a", "", "Plugin author name")
	pluginCreateCmd.Flags().StringP("description", "d", "", "Plugin description")
	pluginCreateCmd.Flags().StringP("output", "o", ".", "Output directory for plugin template")

	// Flags for GitHub install command
	pluginInstallGitHubCmd.Flags().StringP("version", "v", "", "Specific version to install (default: latest)")
	pluginInstallGitHubCmd.Flags().BoolP("force", "f", false, "Force install even if plugin already exists")
}

// initializePluginManager initializes the plugin manager
func initializePluginManager() error {
	if pluginManager != nil {
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	config := manager.ManagerConfig{
		PluginsDir:          filepath.Join(homeDir, ".portunix", "plugins"),
		RegistryFile:        filepath.Join(homeDir, ".portunix", "plugins", "registry.json"),
		HealthCheckInterval: 30 * time.Second,
		DefaultPort:         9001,
		PortRange: manager.PortRange{
			Start: 9000,
			End:   9999,
		},
	}

	pluginManager, err = manager.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to initialize plugin manager: %w", err)
	}

	return nil
}

// listPlugins lists all installed plugins
func listPlugins(showAll bool, outputFormat string) error {
	plugins, err := pluginManager.ListPlugins()
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	if len(plugins) == 0 {
		fmt.Println("No plugins installed.")
		return nil
	}

	switch outputFormat {
	case "json":
		return outputJSON(plugins)
	case "yaml":
		return outputYAML(plugins)
	default:
		return outputPluginTable(plugins, showAll)
	}
}

// installPlugin installs a plugin
func installPlugin(pluginPath string) error {
	manifestPath := filepath.Join(pluginPath, "plugin.yaml")

	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin.yaml not found in %s", pluginPath)
	}

	fmt.Printf("Installing plugin from %s...\n", pluginPath)

	if err := pluginManager.InstallPlugin(manifestPath); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	fmt.Println("✅ Plugin installed successfully!")
	return nil
}

// uninstallPlugin uninstalls a plugin
func uninstallPlugin(pluginName string, force bool) error {
	if !force {
		fmt.Printf("Are you sure you want to uninstall plugin '%s'? (y/N): ", pluginName)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Uninstall cancelled.")
			return nil
		}
	}

	fmt.Printf("Uninstalling plugin '%s'...\n", pluginName)

	if err := pluginManager.UninstallPlugin(pluginName); err != nil {
		return fmt.Errorf("failed to uninstall plugin: %w", err)
	}

	fmt.Println("✅ Plugin uninstalled successfully!")
	return nil
}

// enablePlugin enables a plugin
func enablePlugin(pluginName string) error {
	fmt.Printf("Enabling plugin '%s'...\n", pluginName)

	if err := pluginManager.EnablePlugin(pluginName); err != nil {
		return fmt.Errorf("failed to enable plugin: %w", err)
	}

	fmt.Println("✅ Plugin enabled successfully!")
	return nil
}

// disablePlugin disables a plugin
func disablePlugin(pluginName string) error {
	fmt.Printf("Disabling plugin '%s'...\n", pluginName)

	if err := pluginManager.DisablePlugin(pluginName); err != nil {
		return fmt.Errorf("failed to disable plugin: %w", err)
	}

	fmt.Println("✅ Plugin disabled successfully!")
	return nil
}

// startPlugin starts a plugin
func startPlugin(pluginName string) error {
	fmt.Printf("Starting plugin '%s'...\n", pluginName)

	if err := pluginManager.StartPlugin(pluginName); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	fmt.Println("✅ Plugin started successfully!")
	return nil
}

// stopPlugin stops a plugin
func stopPlugin(pluginName string) error {
	fmt.Printf("Stopping plugin '%s'...\n", pluginName)

	if err := pluginManager.StopPlugin(pluginName); err != nil {
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	fmt.Println("✅ Plugin stopped successfully!")
	return nil
}

// showPluginInfo shows detailed information about a plugin
func showPluginInfo(pluginName string) error {
	pluginInfo, err := pluginManager.GetPlugin(pluginName)
	if err != nil {
		return fmt.Errorf("failed to get plugin info: %w", err)
	}

	fmt.Printf("Plugin Information: %s\n", pluginInfo.Name)
	fmt.Printf("==================%s\n", strings.Repeat("=", len(pluginInfo.Name)))
	fmt.Printf("Name:        %s\n", pluginInfo.Name)
	fmt.Printf("Version:     %s\n", pluginInfo.Version)
	fmt.Printf("Description: %s\n", pluginInfo.Description)
	fmt.Printf("Author:      %s\n", pluginInfo.Author)
	fmt.Printf("License:     %s\n", pluginInfo.License)
	fmt.Printf("Status:      %s\n", pluginInfo.Status.String())
	fmt.Printf("Last Seen:   %s\n", pluginInfo.LastSeen.Format(time.RFC3339))

	fmt.Printf("\nSupported OS: %s\n", strings.Join(pluginInfo.SupportedOS, ", "))

	fmt.Printf("\nCommands:\n")
	for _, cmd := range pluginInfo.Commands {
		fmt.Printf("  - %s: %s\n", cmd.Name, cmd.Description)
		if len(cmd.Subcommands) > 0 {
			fmt.Printf("    Subcommands: %s\n", strings.Join(cmd.Subcommands, ", "))
		}
	}

	fmt.Printf("\nCapabilities:\n")
	fmt.Printf("  Filesystem Access: %t\n", pluginInfo.Capabilities.FilesystemAccess)
	fmt.Printf("  Network Access:    %t\n", pluginInfo.Capabilities.NetworkAccess)
	fmt.Printf("  Database Access:   %t\n", pluginInfo.Capabilities.DatabaseAccess)
	fmt.Printf("  Container Access:  %t\n", pluginInfo.Capabilities.ContainerAccess)
	fmt.Printf("  System Commands:   %t\n", pluginInfo.Capabilities.SystemCommands)

	if len(pluginInfo.Capabilities.MCPTools) > 0 {
		fmt.Printf("  MCP Tools:         %s\n", strings.Join(pluginInfo.Capabilities.MCPTools, ", "))
	}

	return nil
}

// checkPluginHealth checks plugin health
func checkPluginHealth(pluginName string) error {
	health, err := pluginManager.GetPluginHealth(pluginName)
	if err != nil {
		return fmt.Errorf("failed to get plugin health: %w", err)
	}

	fmt.Printf("Plugin Health: %s\n", pluginName)
	fmt.Printf("==============%s\n", strings.Repeat("=", len(pluginName)))

	status := "❌ Unhealthy"
	if health.Healthy {
		status = "✅ Healthy"
	}
	fmt.Printf("Status:     %s\n", status)
	fmt.Printf("Message:    %s\n", health.Message)
	fmt.Printf("Uptime:     %d seconds\n", health.UptimeSeconds)
	fmt.Printf("Last Check: %s\n", health.LastCheckTime.Format(time.RFC3339))

	if len(health.Metrics) > 0 {
		fmt.Printf("\nMetrics:\n")
		for key, value := range health.Metrics {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	return nil
}

// createPluginTemplate creates a new plugin template
func createPluginTemplate(pluginName, author, description, outputDir string) error {
	// Default values
	if author == "" {
		author = "Unknown"
	}
	if description == "" {
		description = fmt.Sprintf("A Portunix plugin: %s", pluginName)
	}

	pluginDir := filepath.Join(outputDir, pluginName)

	// Create plugin directory
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Create plugin.yaml
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	manifestContent := plugins.GetManifestTemplate(pluginName, description, author)

	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		return fmt.Errorf("failed to create plugin.yaml: %w", err)
	}

	// Create basic plugin structure
	if err := createPluginStructure(pluginDir, pluginName); err != nil {
		return fmt.Errorf("failed to create plugin structure: %w", err)
	}

	fmt.Printf("✅ Plugin template created at: %s\n", pluginDir)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("1. cd %s\n", pluginDir)
	fmt.Printf("2. Edit plugin.yaml to configure your plugin\n")
	fmt.Printf("3. Implement your plugin in main.go\n")
	fmt.Printf("4. Build: go build -o %s\n", pluginName)
	fmt.Printf("5. Install: portunix plugin install .\n")

	return nil
}

// validatePlugin validates a plugin
func validatePlugin(pluginPath string) error {
	manifestPath := filepath.Join(pluginPath, "plugin.yaml")

	manifest, err := plugins.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Printf("✅ Plugin validation successful!\n")
	fmt.Printf("Plugin: %s v%s\n", manifest.Name, manifest.Version)
	fmt.Printf("Author: %s\n", manifest.Author)
	fmt.Printf("Description: %s\n", manifest.Description)

	return nil
}

// installPluginFromGitHub installs a plugin from GitHub releases
func installPluginFromGitHub(pluginName, version string, force bool) error {
	const defaultOwner = "cassandragargoyle"
	const defaultRepo = "portunix-plugins"
	
	var owner, repo string
	
	// Check if pluginName contains '/' (custom repository)
	if strings.Contains(pluginName, "/") {
		// Custom repository format: owner/repo
		parts := strings.SplitN(pluginName, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format, expected 'owner/repo', got '%s'", pluginName)
		}
		owner, repo = parts[0], parts[1]
	} else {
		// Default repository, pluginName is just the plugin name
		owner = defaultOwner
		repo = defaultRepo
		fmt.Printf("🏠 Using default repository: %s/%s for plugin '%s'\n", owner, repo, pluginName)
	}

	// Initialize GitHub service - use special service for default repository
	var githubService *github.Service
	if owner == defaultOwner && repo == defaultRepo {
		githubService = github.NewServiceForPortunixPlugins()
	} else {
		githubService = github.NewService()
	}

	fmt.Printf("🔍 Fetching plugin information from %s/%s...\n", owner, repo)

	// First check if repository exists and is accessible
	repoInfo, err := githubService.GetRepositoryInfo(owner, repo)
	if err != nil {
		return handleRepositoryError(owner, repo, err)
	}
	
	fmt.Printf("📂 Repository: %s (⭐ %d stars, 🍴 %d forks)\n", repoInfo.Description, repoInfo.Stars, repoInfo.Forks)

	// Get release information
	var release *github.Release
	if version == "" {
		// Get latest release
		release, err = githubService.GetLatestRelease(owner, repo)
		if err != nil {
			return handleReleaseError(owner, repo, err)
		}
		fmt.Printf("📦 Found latest release: %s\n", release.TagName)
	} else {
		// Get specific release
		release, err = githubService.GetClient().GetRelease(owner, repo, version)
		if err != nil {
			return handleSpecificReleaseError(owner, repo, version, err)
		}
		fmt.Printf("📦 Found release: %s\n", release.TagName)
	}

	// Detect current platform
	platform := detectPlatform()
	fmt.Printf("🖥️  Detected platform: %s\n", platform)

	// Find appropriate asset for current platform
	var asset *github.Asset
	if owner == defaultOwner && repo == defaultRepo {
		// For default repository, look for plugin-specific assets
		asset = findPluginAssetInRelease(release, pluginName, platform)
		if asset == nil {
			fmt.Printf("❌ No asset found for plugin '%s' on platform %s\n", pluginName, platform)
			fmt.Println("Available assets:")
			for _, a := range release.Assets {
				fmt.Printf("  - %s (%s)\n", a.Name, formatSize(a.Size))
			}
			return fmt.Errorf("no suitable asset found for plugin '%s' on platform %s", pluginName, platform)
		}
	} else {
		// For custom repositories, use existing platform detection
		var err error
		asset, err = githubService.FindPlatformAsset(owner, repo, release.TagName, platform)
		if err != nil {
			fmt.Printf("❌ No asset found for platform %s\n", platform)
			fmt.Println("Available assets:")
			for _, a := range release.Assets {
				fmt.Printf("  - %s (%s)\n", a.Name, formatSize(a.Size))
			}
			return fmt.Errorf("no suitable asset found for platform %s", platform)
		}
	}

	fmt.Printf("📥 Downloading asset: %s (%s)\n", asset.Name, formatSize(asset.Size))

	// Create temporary directory for download
	tempDir, err := os.MkdirTemp("", "portunix-plugin-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download the asset
	assetPath := filepath.Join(tempDir, asset.Name)
	progressCallback := func(downloaded, total int64, percentage int) {
		fmt.Printf("\r📥 Downloading... [%s] %d%% (%s/%s)", 
			progressBar(percentage, 30), percentage, 
			formatSize(downloaded), formatSize(total))
	}

	err = githubService.DownloadReleaseAssetWithProgress(owner, repo, release.TagName, asset.Name, assetPath, progressCallback)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	fmt.Println() // New line after progress

	fmt.Printf("✅ Downloaded successfully!\n")

	// Extract the plugin (assuming it's a compressed archive)
	extractDir := filepath.Join(tempDir, "extracted")
	if err := extractAsset(assetPath, extractDir); err != nil {
		return fmt.Errorf("failed to extract plugin: %w", err)
	}

	// Find plugin.yaml in extracted directory
	var manifestPath string
	if owner == defaultOwner && repo == defaultRepo {
		// For default repository, look for plugin manifest in plugin-specific directory
		manifestPath, err = findPluginManifestInDefaultRepo(extractDir, pluginName)
	} else {
		// For custom repositories, look in root directory
		manifestPath, err = findPluginManifest(extractDir)
	}
	if err != nil {
		return fmt.Errorf("failed to find plugin manifest: %w", err)
	}

	fmt.Printf("📄 Found plugin manifest: %s\n", manifestPath)

	// Install the plugin using existing plugin manager
	fmt.Printf("🔧 Installing plugin...\n")
	if err := pluginManager.InstallPlugin(manifestPath); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	fmt.Printf("✅ Plugin installed successfully from %s/%s!\n", owner, repo)
	return nil
}

// listAvailablePlugins lists available plugins from the default repository
func listAvailablePlugins() error {
	const defaultOwner = "cassandragargoyle"
	const defaultRepo = "portunix-plugins"

	// Initialize GitHub service with embedded token for default repository  
	githubService := github.NewServiceForPortunixPlugins()

	fmt.Printf("🔍 Fetching available plugins from %s/%s...\n", defaultOwner, defaultRepo)

	// First check if repository exists and is accessible
	_, err := githubService.GetRepositoryInfo(defaultOwner, defaultRepo)
	if err != nil {
		return handleRepositoryError(defaultOwner, defaultRepo, err)
	}

	// Get latest release
	release, err := githubService.GetLatestRelease(defaultOwner, defaultRepo)
	if err != nil {
		return handleReleaseError(defaultOwner, defaultRepo, err)
	}

	fmt.Printf("📦 Latest release: %s (published %s)\n", release.TagName, release.PublishedAt)
	
	if release.Body != "" {
		fmt.Printf("📝 Release notes: %s\n\n", release.Body)
	}

	// Parse available plugins from assets
	plugins := parsePluginsFromAssets(release.Assets)
	
	if len(plugins) == 0 {
		fmt.Println("❌ No plugins found in the latest release.")
		fmt.Println("\nAvailable assets:")
		for _, asset := range release.Assets {
			fmt.Printf("  - %s (%s)\n", asset.Name, formatSize(asset.Size))
		}
		return nil
	}

	fmt.Printf("📋 Available plugins (%d found):\n", len(plugins))
	fmt.Printf("%-30s %-15s %-20s %s\n", "PLUGIN NAME", "PLATFORM", "FORMAT", "SIZE")
	fmt.Printf("%-30s %-15s %-20s %s\n", strings.Repeat("-", 30), strings.Repeat("-", 15), strings.Repeat("-", 20), strings.Repeat("-", 10))

	for _, plugin := range plugins {
		fmt.Printf("%-30s %-15s %-20s %s\n", plugin.Name, plugin.Platform, plugin.Format, formatSize(plugin.Size))
	}

	fmt.Printf("\n💡 Install any plugin with: portunix plugin install-github <plugin-name>\n")
	fmt.Printf("💡 Example: portunix plugin install-github %s\n", plugins[0].Name)

	return nil
}

// outputJSON outputs plugins in JSON format
func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// outputYAML outputs plugins in YAML format
func outputYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(data)
}

// Helper functions for output formatting
func outputPluginTable(plugins []plugins.PluginInfo, showAll bool) error {
	// Implementation for table output
	fmt.Printf("%-20s %-10s %-15s %-30s\n", "NAME", "VERSION", "STATUS", "DESCRIPTION")
	fmt.Printf("%-20s %-10s %-15s %-30s\n", "----", "-------", "------", "-----------")

	for _, plugin := range plugins {
		fmt.Printf("%-20s %-10s %-15s %-30s\n",
			plugin.Name,
			plugin.Version,
			plugin.Status.String(),
			truncateString(plugin.Description, 30))
	}

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// createPluginStructure creates basic plugin file structure
func createPluginStructure(pluginDir, pluginName string) error {
	// Create directories
	dirs := []string{"src", "proto", "examples"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(pluginDir, dir), 0755); err != nil {
			return err
		}
	}

	// Create main.go template
	mainGoContent := fmt.Sprintf(`package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	// Import your plugin protocol definitions
)

func main() {
	// Parse command line arguments
	port := "9001"
	if len(os.Args) > 2 && os.Args[1] == "--port" {
		port = os.Args[2]
	}

	// Start gRPC server
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %%s: %%v", port, err)
	}

	server := grpc.NewServer()
	
	// Register your plugin service
	// pb.RegisterPluginServiceServer(server, &YourPluginService{})

	fmt.Printf("%%s plugin starting on port %%s\n", "%s", port)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("Shutting down...")
		server.GracefulStop()
	}()

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %%v", err)
	}
}
`, pluginName)

	mainGoPath := filepath.Join(pluginDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		return err
	}

	// Create README.md
	readmeContent := fmt.Sprintf("# %s Plugin\n\nA Portunix plugin for %s functionality.\n\n## Building\n\n```bash\ngo build -o %s\n```\n\n## Installation\n\n```bash\nportunix plugin install .\n```\n\n## Usage\n\n```bash\nportunix plugin enable %s\nportunix plugin start %s\n```\n", pluginName, pluginName, pluginName, pluginName, pluginName)

	readmePath := filepath.Join(pluginDir, "README.md")
	return os.WriteFile(readmePath, []byte(readmeContent), 0644)
}

// Helper functions for GitHub plugin installation

// detectPlatform detects the current platform for asset matching
func detectPlatform() string {
	// Use runtime detection or fall back to hardcoded values
	// For Phase 1 MVP, we'll use simple detection
	return "linux-amd64" // Default for development environment
}

// formatSize formats bytes into human readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// progressBar creates a simple progress bar string
func progressBar(percentage, width int) string {
	pos := percentage * width / 100
	bar := ""
	for i := 0; i < width; i++ {
		if i < pos {
			bar += "="
		} else if i == pos {
			bar += ">"
		} else {
			bar += " "
		}
	}
	return bar
}

// extractAsset extracts a downloaded asset to a directory
func extractAsset(assetPath, extractDir string) error {
	// Create extract directory
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extract directory: %w", err)
	}

	// For Phase 1 MVP, we'll handle tar.gz and zip files
	if strings.HasSuffix(assetPath, ".tar.gz") || strings.HasSuffix(assetPath, ".tgz") {
		return extractTarGz(assetPath, extractDir)
	} else if strings.HasSuffix(assetPath, ".zip") {
		return extractZip(assetPath, extractDir)
	} else {
		// If not a known archive, assume it's a single binary
		filename := filepath.Base(assetPath)
		destPath := filepath.Join(extractDir, filename)
		return copyFile(assetPath, destPath)
	}
}

// extractTarGz extracts a tar.gz file
func extractTarGz(tarPath, destDir string) error {
	// Use built-in archive functionality from app/archive.go
	return fmt.Errorf("tar.gz extraction not implemented yet - will use existing archive functionality")
}

// extractZip extracts a zip file
func extractZip(zipPath, destDir string) error {
	// Use built-in archive functionality from app/archive.go
	return app.UnzipFile(zipPath, destDir)
}

// copyFile copies a file from src to dst (used for single binary assets)
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

// findPluginManifest finds plugin.yaml in the extracted directory
func findPluginManifest(searchDir string) (string, error) {
	var manifestPath string

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "plugin.yaml" {
			manifestPath = path
			return filepath.SkipDir // Stop searching
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if manifestPath == "" {
		return "", fmt.Errorf("plugin.yaml not found in %s", searchDir)
	}

	return manifestPath, nil
}

// findPluginAssetInRelease finds the appropriate asset for a specific plugin in the default repository
func findPluginAssetInRelease(release *github.Release, pluginName, platform string) *github.Asset {
	// Expected asset naming convention for plugins: pluginName-platform.ext
	// Examples: agile-software-development-linux-amd64.tar.gz, agile-software-development-windows-amd64.zip
	
	expectedPatterns := []string{
		fmt.Sprintf("%s-%s.tar.gz", pluginName, platform),
		fmt.Sprintf("%s-%s.tgz", pluginName, platform),
		fmt.Sprintf("%s-%s.zip", pluginName, platform),
		// Fallback patterns without platform (universal binaries)
		fmt.Sprintf("%s.tar.gz", pluginName),
		fmt.Sprintf("%s.tgz", pluginName),
		fmt.Sprintf("%s.zip", pluginName),
	}
	
	// First try exact matches
	for _, pattern := range expectedPatterns {
		for _, asset := range release.Assets {
			if asset.Name == pattern {
				return asset
			}
		}
	}
	
	// Then try partial matches (in case of different naming conventions)
	for _, asset := range release.Assets {
		assetLower := strings.ToLower(asset.Name)
		pluginLower := strings.ToLower(pluginName)
		platformLower := strings.ToLower(platform)
		
		// Check if asset contains plugin name and platform
		if strings.Contains(assetLower, pluginLower) && strings.Contains(assetLower, platformLower) {
			return asset
		}
	}
	
	// Finally try just plugin name match (for universal assets)
	for _, asset := range release.Assets {
		assetLower := strings.ToLower(asset.Name)
		pluginLower := strings.ToLower(pluginName)
		
		if strings.Contains(assetLower, pluginLower) {
			return asset
		}
	}
	
	return nil
}

// findPluginManifestInDefaultRepo finds plugin.yaml in the default repository structure
func findPluginManifestInDefaultRepo(searchDir, pluginName string) (string, error) {
	// Expected directory structure: plugins/pluginName/plugin.yaml
	possiblePaths := []string{
		filepath.Join(searchDir, "plugins", pluginName, "plugin.yaml"),
		filepath.Join(searchDir, pluginName, "plugin.yaml"),
		filepath.Join(searchDir, "plugin.yaml"), // Fallback to root
	}
	
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	
	// If not found in expected locations, search recursively
	return findPluginManifest(searchDir)
}

// parsePluginsFromAssets extracts plugin information from GitHub release assets
func parsePluginsFromAssets(assets []*github.Asset) []AvailablePluginInfo {
	var plugins []AvailablePluginInfo
	pluginMap := make(map[string]bool) // To avoid duplicates
	
	for _, asset := range assets {
		pluginInfo := parsePluginFromAssetName(asset)
		if pluginInfo != nil {
			// Check for duplicates based on plugin name
			key := pluginInfo.Name
			if !pluginMap[key] {
				plugins = append(plugins, *pluginInfo)
				pluginMap[key] = true
			}
		}
	}
	
	return plugins
}

// parsePluginFromAssetName extracts plugin information from asset name
func parsePluginFromAssetName(asset *github.Asset) *AvailablePluginInfo {
	name := asset.Name
	
	// Expected patterns:
	// pluginname-platform.ext
	// pluginname.ext (universal)
	
	// Remove file extensions
	var pluginName, platform, format string
	
	if strings.HasSuffix(name, ".tar.gz") {
		format = "tar.gz"
		name = strings.TrimSuffix(name, ".tar.gz")
	} else if strings.HasSuffix(name, ".tgz") {
		format = "tgz" 
		name = strings.TrimSuffix(name, ".tgz")
	} else if strings.HasSuffix(name, ".zip") {
		format = "zip"
		name = strings.TrimSuffix(name, ".zip")
	} else {
		// Unknown format, skip
		return nil
	}
	
	// Try to parse platform
	platformPatterns := []string{
		"linux-amd64", "linux-386", "linux-arm64", "linux-arm",
		"windows-amd64", "windows-386",
		"darwin-amd64", "darwin-arm64",
	}
	
	platform = "universal"
	for _, p := range platformPatterns {
		if strings.HasSuffix(name, "-"+p) {
			platform = p
			pluginName = strings.TrimSuffix(name, "-"+p)
			break
		}
	}
	
	if platform == "universal" {
		pluginName = name
	}
	
	// Skip if plugin name is empty or looks like a system file
	if pluginName == "" || strings.HasPrefix(pluginName, ".") {
		return nil
	}
	
	return &AvailablePluginInfo{
		Name:      pluginName,
		Platform:  platform,
		Format:    format,
		Size:      asset.Size,
		AssetName: asset.Name,
	}
}

// Error handling functions for better user experience

// handleRepositoryError provides user-friendly error messages for repository access issues
func handleRepositoryError(owner, repo string, err error) error {
	errMsg := err.Error()
	
	if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "Not Found") {
		fmt.Printf("❌ Repository not found: %s/%s\n", owner, repo)
		fmt.Println("\nPossible issues:")
		fmt.Println("  • Repository doesn't exist")
		fmt.Println("  • Repository is private and you don't have access")
		fmt.Println("  • Repository name is misspelled")
		
		if owner == "cassandragargoyle" && repo == "portunix-plugins" {
			fmt.Println("\n💡 The official Portunix plugins repository is not yet created.")
			fmt.Println("💡 You can install plugins from custom repositories using: owner/repo format")
			fmt.Println("💡 Example: portunix plugin install-github microsoft/vscode")
		}
		
		return fmt.Errorf("repository %s/%s not accessible", owner, repo)
	}
	
	if strings.Contains(errMsg, "rate limit") {
		fmt.Println("❌ GitHub API rate limit exceeded")
		fmt.Println("\n💡 Set GITHUB_TOKEN environment variable to increase rate limits")
		fmt.Println("💡 Create token at: https://github.com/settings/tokens")
		return fmt.Errorf("GitHub API rate limit exceeded")
	}
	
	if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "connection") {
		fmt.Println("❌ Network connection error")
		fmt.Println("\n💡 Check your internet connection")
		fmt.Println("💡 GitHub might be temporarily unavailable")
		return fmt.Errorf("network error connecting to GitHub")
	}
	
	return fmt.Errorf("failed to access repository %s/%s: %w", owner, repo, err)
}

// handleReleaseError provides user-friendly error messages for release access issues
func handleReleaseError(owner, repo string, err error) error {
	errMsg := err.Error()
	
	if strings.Contains(errMsg, "no releases found") || strings.Contains(errMsg, "404") {
		fmt.Printf("📦 Repository %s/%s exists but has no releases\n", owner, repo)
		fmt.Println("\nThis repository doesn't have any published releases yet.")
		
		if owner == "cassandragargoyle" && repo == "portunix-plugins" {
			fmt.Println("\n💡 The official Portunix plugins repository is under development.")
			fmt.Println("💡 Check back later for available plugins.")
		} else {
			fmt.Println("\n💡 Contact the repository maintainer to publish releases.")
			fmt.Println("💡 Or try installing from a different repository.")
		}
		
		return fmt.Errorf("no releases available in %s/%s", owner, repo)
	}
	
	return handleRepositoryError(owner, repo, err)
}

// handleSpecificReleaseError provides user-friendly error messages for specific version issues
func handleSpecificReleaseError(owner, repo, version string, err error) error {
	errMsg := err.Error()
	
	if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "Not Found") {
		fmt.Printf("📦 Release version '%s' not found in %s/%s\n", version, owner, repo)
		
		// Try to show available releases
		githubService := github.NewService()
		releases, listErr := githubService.ListRepositoryReleases(owner, repo)
		if listErr == nil && len(releases) > 0 {
			fmt.Println("\nAvailable releases:")
			for i, rel := range releases {
				if i >= 5 { // Show only first 5
					fmt.Printf("  ... and %d more\n", len(releases)-5)
					break
				}
				fmt.Printf("  • %s (published %s)\n", rel.TagName, rel.PublishedAt)
			}
		} else {
			fmt.Println("\n💡 Use: portunix plugin install-github <plugin-name>  (without --version for latest)")
		}
		
		return fmt.Errorf("version %s not found in %s/%s", version, owner, repo)
	}
	
	return handleRepositoryError(owner, repo, err)
}
