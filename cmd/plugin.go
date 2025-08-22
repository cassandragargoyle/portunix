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
	"portunix.cz/app/plugins"
	"portunix.cz/app/plugins/manager"
)

var pluginManager *manager.Manager

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

func init() {
	rootCmd.AddCommand(pluginCmd)

	// Add subcommands
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
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