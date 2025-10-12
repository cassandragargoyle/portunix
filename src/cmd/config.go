package cmd

import (
	"fmt"

	"portunix.ai/app/config"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Portunix configuration",
	Long: `Manage Portunix configuration settings.

Configuration is loaded from (in order of priority):
1. ./portunix-config.yaml (project directory)
2. ~/.portunix/config.yaml (user directory)
3. ~/.config/portunix/config.yaml (Linux standard)
4. Built-in defaults

Available configuration keys:
- container_runtime: Container runtime to use (docker, podman)
- verbose: Enable verbose output (true, false)
- auto_update: Enable automatic updates (true, false)`,
}

// configGetCmd gets a configuration value
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long:  `Get the value of a specific configuration key.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		
		value, err := config.GetConfigValue(key)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}
		
		fmt.Printf("%s\n", value)
	},
}

// configSetCmd sets a configuration value
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  `Set the value of a specific configuration key.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]
		
		err := config.SetConfigValue(key, value)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}
		
		fmt.Printf("✅ Configuration updated: %s = %s\n", key, value)
	},
}

// configShowCmd shows all configuration values
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all configuration values",
	Long:  `Display all current configuration values.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("❌ Error loading configuration: %v\n", err)
			return
		}
		
		fmt.Println("⚙️  Portunix Configuration")
		fmt.Println("========================")
		fmt.Printf("Container Runtime: %s\n", cfg.ContainerRuntime)
		fmt.Printf("Verbose: %v\n", cfg.Verbose)
		fmt.Printf("Auto Update: %v\n", cfg.AutoUpdate)
		
		fmt.Println("\n💡 Configuration file locations (in priority order):")
		fmt.Println("  1. ./portunix-config.yaml (project directory)")
		fmt.Println("  2. ~/.portunix/config.yaml (user directory)")
		fmt.Println("  3. ~/.config/portunix/config.yaml (Linux standard)")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)
}