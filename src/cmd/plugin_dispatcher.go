package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"portunix.ai/app/plugins/manager"
)

// initPluginDispatcher registers enabled plugins as dynamic subcommands
func initPluginDispatcher() {
	// Get plugins directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return // silently fail - plugins won't be available
	}

	registryPath := filepath.Join(homeDir, ".portunix", "plugins", "registry.json")

	// Check if registry exists
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		return // no plugins installed
	}

	// Load registry
	registry, err := manager.NewRegistry(registryPath)
	if err != nil {
		return // silently fail
	}

	// Get all plugins (we want to show all installed, not just enabled)
	plugins, err := registry.ListPlugins()
	if err != nil {
		return
	}

	// Register each plugin as a subcommand
	for _, plugin := range plugins {
		// Get full registry data for runtime info
		registryData, err := registry.GetPluginRegistryData(plugin.Name)
		if err != nil {
			continue
		}

		// Create command for this plugin
		pluginCmd := createPluginCommand(registryData)
		rootCmd.AddCommand(pluginCmd)
	}
}

// createPluginCommand creates a Cobra command for a plugin
func createPluginCommand(plugin *manager.RegistryPlugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   plugin.Name + " [command] [args...]",
		Short: plugin.Description,
		Long: fmt.Sprintf(`%s

Version: %s
Author: %s
Mode: %s
Runtime: %s`,
			plugin.Description,
			plugin.Version,
			plugin.Author,
			getPluginMode(plugin.Mode),
			getPluginRuntime(plugin.Runtime)),
		DisableFlagParsing: true, // pass all flags to plugin
		RunE: func(cmd *cobra.Command, args []string) error {
			return executePlugin(plugin, args)
		},
	}

	return cmd
}

// getPluginMode returns human-readable mode
func getPluginMode(mode string) string {
	if mode == "" {
		return "service"
	}
	return mode
}

// getPluginRuntime returns human-readable runtime
func getPluginRuntime(runtime string) string {
	if runtime == "" {
		return "native"
	}
	return runtime
}

// executePlugin executes the plugin binary with given arguments
func executePlugin(plugin *manager.RegistryPlugin, args []string) error {
	// Re-read registry to get current enabled status
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	registryPath := filepath.Join(homeDir, ".portunix", "plugins", "registry.json")
	registry, err := manager.NewRegistry(registryPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin registry: %w", err)
	}

	// Get fresh plugin data
	freshPlugin, err := registry.GetPluginRegistryData(plugin.Name)
	if err != nil {
		return fmt.Errorf("plugin '%s' not found: %w", plugin.Name, err)
	}

	// Check if plugin is enabled
	if !freshPlugin.Enabled {
		return fmt.Errorf("plugin '%s' is not enabled. Enable it with: portunix plugin enable %s", plugin.Name, plugin.Name)
	}

	// Build binary path
	binaryPath := filepath.Join(freshPlugin.InstallPath, freshPlugin.BinaryName)

	// Determine runtime and build command
	runtime := freshPlugin.Runtime
	if runtime == "" {
		runtime = "native"
	}

	var cmd *exec.Cmd

	switch runtime {
	case "java":
		cmd = buildJavaCommand(binaryPath, freshPlugin.JVMArgs, args)
	case "python":
		cmd = buildPythonCommand(binaryPath, args)
	default: // native
		cmd = buildNativeCommand(binaryPath, args)
	}

	// Set working directory
	cmd.Dir = freshPlugin.InstallPath

	// Connect stdio
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute and return exit code
	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		return fmt.Errorf("failed to execute plugin: %w", err)
	}

	return nil
}

// buildJavaCommand builds command for Java plugin
func buildJavaCommand(jarPath string, jvmArgs []string, pluginArgs []string) *exec.Cmd {
	args := []string{}

	// Add JVM args
	if len(jvmArgs) > 0 {
		args = append(args, jvmArgs...)
	} else {
		// Default JVM args
		args = append(args, "-Xmx256m", "-Xms64m")
	}

	// Add UTF-8 encoding if not already specified in jvmArgs
	jvmArgsStr := strings.Join(jvmArgs, " ")
	if !strings.Contains(jvmArgsStr, "-Dfile.encoding") {
		args = append(args, "-Dfile.encoding=UTF-8")
	}
	if !strings.Contains(jvmArgsStr, "-Dstdout.encoding") {
		args = append(args, "-Dstdout.encoding=UTF-8")
	}
	if !strings.Contains(jvmArgsStr, "-Dstderr.encoding") {
		args = append(args, "-Dstderr.encoding=UTF-8")
	}

	// Add -jar and path
	args = append(args, "-jar", jarPath)

	// Add plugin args
	args = append(args, pluginArgs...)

	return exec.Command("java", args...)
}

// buildPythonCommand builds command for Python plugin
func buildPythonCommand(scriptPath string, pluginArgs []string) *exec.Cmd {
	args := []string{scriptPath}
	args = append(args, pluginArgs...)
	return exec.Command("python3", args...)
}

// buildNativeCommand builds command for native plugin
func buildNativeCommand(binaryPath string, pluginArgs []string) *exec.Cmd {
	return exec.Command(binaryPath, pluginArgs...)
}

// GetPluginNames returns list of registered plugin names (for help/completion)
func GetPluginNames() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	registryPath := filepath.Join(homeDir, ".portunix", "plugins", "registry.json")
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		return nil
	}

	registry, err := manager.NewRegistry(registryPath)
	if err != nil {
		return nil
	}

	plugins, err := registry.ListPlugins()
	if err != nil {
		return nil
	}

	var names []string
	for _, p := range plugins {
		names = append(names, p.Name)
	}

	return names
}

// isPluginCommand checks if a command name is a registered plugin
func isPluginCommand(name string) bool {
	names := GetPluginNames()
	for _, n := range names {
		if strings.EqualFold(n, name) {
			return true
		}
	}
	return false
}
