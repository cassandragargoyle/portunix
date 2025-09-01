package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"portunix.cz/app/datastore"
	"portunix.cz/app/plugins/manager"
)

var datastoreManager *datastore.Manager

// datastoreCmd represents the datastore command
var datastoreCmd = &cobra.Command{
	Use:   "datastore",
	Short: "Datastore management commands",
	Long: `Datastore management commands for Portunix.

The datastore system provides unified data storage with routing to different backends.
Built-in file-based storage is available by default, with optional plugins for 
external databases (MongoDB, PostgreSQL, Redis, etc.).

Examples:
  portunix datastore config                    # Show current configuration
  portunix datastore store docs/readme "content"  # Store data
  portunix datastore get docs/readme           # Retrieve data
  portunix datastore query "docs/*"            # Query data by pattern`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeDatastoreManager()
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if datastoreManager != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			return datastoreManager.Close(ctx)
		}
		return nil
	},
}

// Datastore config commands
var datastoreConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Datastore configuration management",
	Long:  `Manage datastore configuration including routing rules and plugin settings.`,
}

var datastoreConfigShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current datastore configuration",
	Long:  `Display the current datastore configuration including routes and plugins.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")
		return showDatastoreConfig(outputFormat)
	},
}

var datastoreConfigEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit datastore configuration",
	Long:  `Open the datastore configuration file in the default editor.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return editDatastoreConfig()
	},
}

var datastoreConfigValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate datastore configuration",
	Long:  `Validate the current datastore configuration for errors.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return validateDatastoreConfig()
	},
}

// Data operation commands
var datastoreStoreCmd = &cobra.Command{
	Use:   "store <key> <value>",
	Short: "Store data",
	Long: `Store data with automatic routing based on configuration.

Examples:
  portunix datastore store "docs/readme" "# My Project"
  echo "content" | portunix datastore store "docs/readme" -`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		var value string

		if len(args) == 2 {
			if args[1] == "-" {
				// Read from stdin
				data, err := os.ReadFile("/dev/stdin")
				if err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
				value = string(data)
			} else {
				value = args[1]
			}
		} else {
			return fmt.Errorf("value is required")
		}

		metadata := make(map[string]interface{})
		metadataFlag, _ := cmd.Flags().GetString("metadata")
		if metadataFlag != "" {
			if err := json.Unmarshal([]byte(metadataFlag), &metadata); err != nil {
				return fmt.Errorf("invalid metadata JSON: %w", err)
			}
		}

		return storeData(key, value, metadata)
	},
}

var datastoreGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Retrieve data",
	Long:  `Retrieve data by key.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")
		return getData(key, outputFormat)
	},
}

var datastoreQueryCmd = &cobra.Command{
	Use:   "query <pattern>",
	Short: "Query data by pattern",
	Long: `Query data using pattern matching.

Examples:
  portunix datastore query "docs/*"
  portunix datastore query "*" --filter '{"type":"document"}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")
		filterFlag, _ := cmd.Flags().GetString("filter")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		var filter map[string]interface{}
		if filterFlag != "" {
			if err := json.Unmarshal([]byte(filterFlag), &filter); err != nil {
				return fmt.Errorf("invalid filter JSON: %w", err)
			}
		}

		return queryData(pattern, filter, limit, offset, outputFormat)
	},
}

var datastoreDeleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete data",
	Long:  `Delete data by key.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		force, _ := cmd.Flags().GetBool("force")
		return deleteData(key, force)
	},
}

var datastoreListCmd = &cobra.Command{
	Use:   "list [pattern]",
	Short: "List keys",
	Long: `List all keys or keys matching a pattern.

Examples:
  portunix datastore list
  portunix datastore list "docs/*"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := "*"
		if len(args) > 0 {
			pattern = args[0]
		}
		outputFormat, _ := cmd.Flags().GetString("output")
		return listKeys(pattern, outputFormat)
	},
}

// Status and management commands
var datastoreStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show datastore status",
	Long:  `Show health and statistics for all configured datastores.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showDatastoreStatus()
	},
}

var datastorePluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Datastore plugin management",
	Long:  `Manage datastore plugins.`,
}

var datastorePluginsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available datastore plugins",
	Long:  `List all available datastore plugins and their information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listDatastorePlugins()
	},
}

func init() {
	rootCmd.AddCommand(datastoreCmd)

	// Config subcommands
	datastoreCmd.AddCommand(datastoreConfigCmd)
	datastoreConfigCmd.AddCommand(datastoreConfigShowCmd)
	datastoreConfigCmd.AddCommand(datastoreConfigEditCmd)
	datastoreConfigCmd.AddCommand(datastoreConfigValidateCmd)

	// Data operation commands
	datastoreCmd.AddCommand(datastoreStoreCmd)
	datastoreCmd.AddCommand(datastoreGetCmd)
	datastoreCmd.AddCommand(datastoreQueryCmd)
	datastoreCmd.AddCommand(datastoreDeleteCmd)
	datastoreCmd.AddCommand(datastoreListCmd)

	// Status and management commands
	datastoreCmd.AddCommand(datastoreStatusCmd)
	datastoreCmd.AddCommand(datastorePluginsCmd)
	datastorePluginsCmd.AddCommand(datastorePluginsListCmd)

	// Flags for various commands
	datastoreConfigShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")

	datastoreStoreCmd.Flags().StringP("metadata", "m", "", "Metadata as JSON string")

	datastoreGetCmd.Flags().StringP("output", "o", "raw", "Output format: raw, json, yaml")

	datastoreQueryCmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")
	datastoreQueryCmd.Flags().StringP("filter", "f", "", "Filter criteria as JSON string")
	datastoreQueryCmd.Flags().IntP("limit", "l", 0, "Limit number of results")
	datastoreQueryCmd.Flags().IntP("offset", "s", 0, "Offset for pagination")

	datastoreDeleteCmd.Flags().BoolP("force", "f", false, "Force delete without confirmation")

	datastoreListCmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")
}

// initializeDatastoreManager initializes the datastore manager
func initializeDatastoreManager() error {
	if datastoreManager != nil {
		return nil
	}

	// Load configuration
	config, err := datastore.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load datastore configuration: %w", err)
	}

	// Initialize plugin manager if needed
	var pluginMgr *manager.Manager
	if pluginManager != nil {
		pluginMgr = pluginManager
	}

	// Create datastore manager
	datastoreManager, err = datastore.NewManager(config, pluginMgr)
	if err != nil {
		return fmt.Errorf("failed to create datastore manager: %w", err)
	}

	// Initialize datastores
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := datastoreManager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize datastores: %w", err)
	}

	return nil
}

// showDatastoreConfig shows the current datastore configuration
func showDatastoreConfig(outputFormat string) error {
	config, err := datastore.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(config)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		defer encoder.Close()
		return encoder.Encode(config)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// editDatastoreConfig opens the configuration file in an editor
func editDatastoreConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".portunix", "datastore.yaml")

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano" // Default editor
	}

	fmt.Printf("Opening %s with %s...\n", configPath, editor)
	return fmt.Errorf("editor functionality not implemented yet")
}

// validateDatastoreConfig validates the configuration
func validateDatastoreConfig() error {
	config, err := datastore.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := datastore.ValidateConfig(config); err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		return err
	}

	fmt.Println("✅ Configuration is valid!")
	return nil
}

// storeData stores data using the datastore manager
func storeData(key, value string, metadata map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := datastoreManager.Store(ctx, key, value, metadata); err != nil {
		return fmt.Errorf("failed to store data: %w", err)
	}

	fmt.Printf("✅ Data stored successfully: %s\n", key)
	return nil
}

// getData retrieves data using the datastore manager
func getData(key, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	value, err := datastoreManager.Retrieve(ctx, key, nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve data: %w", err)
	}

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		defer encoder.Close()
		return encoder.Encode(value)
	case "raw":
		fmt.Print(value)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// queryData queries data using the datastore manager
func queryData(pattern string, filter map[string]interface{}, limit, offset int, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	criteria := datastore.QueryCriteria{
		Collection: pattern,
		Filter:     filter,
		Limit:      limit,
		Offset:     offset,
	}

	results, err := datastoreManager.Query(ctx, criteria)
	if err != nil {
		return fmt.Errorf("failed to query data: %w", err)
	}

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(results)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		defer encoder.Close()
		return encoder.Encode(results)
	case "table":
		return outputQueryTable(results)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// deleteData deletes data using the datastore manager
func deleteData(key string, force bool) error {
	if !force {
		fmt.Printf("Are you sure you want to delete '%s'? (y/N): ", key)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Delete cancelled.")
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := datastoreManager.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete data: %w", err)
	}

	fmt.Printf("✅ Data deleted successfully: %s\n", key)
	return nil
}

// listKeys lists keys using the datastore manager
func listKeys(pattern, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	keys, err := datastoreManager.List(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(keys)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		defer encoder.Close()
		return encoder.Encode(keys)
	case "table":
		for _, key := range keys {
			fmt.Println(key)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// showDatastoreStatus shows the status of all datastores
func showDatastoreStatus() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health, err := datastoreManager.Health(ctx)
	if err != nil {
		return fmt.Errorf("failed to get health status: %w", err)
	}

	stats, err := datastoreManager.Stats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get statistics: %w", err)
	}

	fmt.Println("Datastore Status")
	fmt.Println("================")

	for pluginName, healthStatus := range health {
		statusIcon := "❌"
		if healthStatus.Healthy {
			statusIcon = "✅"
		}

		fmt.Printf("\n%s %s\n", statusIcon, pluginName)
		fmt.Printf("  Status: %s\n", healthStatus.Status)
		if healthStatus.Message != "" {
			fmt.Printf("  Message: %s\n", healthStatus.Message)
		}
		fmt.Printf("  Uptime: %v\n", healthStatus.Uptime)

		if pluginStats, exists := stats[pluginName]; exists {
			fmt.Printf("  Keys: %d\n", pluginStats.TotalKeys)
			fmt.Printf("  Size: %d bytes\n", pluginStats.TotalSize)
		}
	}

	return nil
}

// listDatastorePlugins lists available datastore plugins
func listDatastorePlugins() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info, err := datastoreManager.GetDatastoreInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get datastore info: %w", err)
	}

	fmt.Printf("%-20s %-10s %-15s %-30s\n", "NAME", "TYPE", "VERSION", "DESCRIPTION")
	fmt.Printf("%-20s %-10s %-15s %-30s\n", "----", "----", "-------", "-----------")

	for name, datastoreInfo := range info {
		fmt.Printf("%-20s %-10s %-15s %-30s\n",
			name,
			string(datastoreInfo.Type),
			datastoreInfo.Version,
			truncateString(datastoreInfo.Description, 30))
	}

	return nil
}

// outputQueryTable outputs query results in table format
func outputQueryTable(results []datastore.QueryResult) error {
	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	fmt.Printf("%-30s %-20s %-30s\n", "KEY", "TYPE", "VALUE (preview)")
	fmt.Printf("%-30s %-20s %-30s\n", "---", "----", "-------------")

	for _, result := range results {
		valuePreview := fmt.Sprintf("%v", result.Value)
		valuePreview = strings.ReplaceAll(valuePreview, "\n", " ")

		fmt.Printf("%-30s %-20s %-30s\n",
			truncateString(result.Key, 30),
			"data",
			truncateString(valuePreview, 30))
	}

	fmt.Printf("\nTotal: %d results\n", len(results))
	return nil
}
