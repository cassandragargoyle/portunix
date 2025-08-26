package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cassandragargoyle/portunix/app/wizard/engine"
	"github.com/spf13/cobra"
)

var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "Interactive installation wizards",
	Long: `Run interactive installation wizards that guide you through
complex setup processes step by step.

Wizards provide an easy way to install and configure software
with rich CLI components like progress bars, selection menus,
and conditional flows.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var wizardRunCmd = &cobra.Command{
	Use:   "run [wizard-name-or-file]",
	Short: "Run a wizard",
	Long: `Run an installation wizard by name or from a YAML file.

Examples:
  portunix wizard run database-setup
  portunix wizard run ./my-wizard.yaml`,
	Args: cobra.ExactArgs(1),
	Run:  runWizard,
}

var wizardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available wizards",
	Long:  "List all built-in and user-defined wizards.",
	Run:   listWizards,
}

var wizardValidateCmd = &cobra.Command{
	Use:   "validate [wizard-file]",
	Short: "Validate wizard YAML file",
	Long:  "Validate the syntax and structure of a wizard YAML file.",
	Args:  cobra.ExactArgs(1),
	Run:   validateWizard,
}

var wizardCreateCmd = &cobra.Command{
	Use:   "create [wizard-name]",
	Short: "Create a new wizard from template",
	Long:  "Create a new wizard YAML file from a template.",
	Args:  cobra.ExactArgs(1),
	Run:   createWizard,
}

// Command flags
var (
	wizardTheme          string
	wizardNonInteractive bool
	wizardConfig         string
	wizardPreset         string
)

func init() {
	rootCmd.AddCommand(wizardCmd)
	wizardCmd.AddCommand(wizardRunCmd)
	wizardCmd.AddCommand(wizardListCmd)
	wizardCmd.AddCommand(wizardValidateCmd)
	wizardCmd.AddCommand(wizardCreateCmd)

	// Flags for wizard run
	wizardRunCmd.Flags().StringVar(&wizardTheme, "theme", "default", "UI theme (default, colorful, minimal)")
	wizardRunCmd.Flags().BoolVar(&wizardNonInteractive, "non-interactive", false, "Run in non-interactive mode")
	wizardRunCmd.Flags().StringVar(&wizardConfig, "config", "", "Configuration file for non-interactive mode")
	wizardRunCmd.Flags().StringVar(&wizardPreset, "preset", "", "Use preset values (development, production)")
}

func runWizard(cmd *cobra.Command, args []string) {
	wizardInput := args[0]

	wizardEngine := engine.NewWizardEngine()

	// Set theme if specified
	if wizardTheme != "default" {
		err := wizardEngine.SetTheme(wizardTheme)
		if err != nil {
			fmt.Printf("Warning: %v, using default theme\n", err)
		}
	}

	// Determine wizard file path
	wizardPath, err := resolveWizardPath(wizardInput)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Load wizard
	wizard, err := wizardEngine.LoadWizard(wizardPath)
	if err != nil {
		fmt.Printf("Error loading wizard: %v\n", err)
		os.Exit(1)
	}

	// Handle non-interactive mode
	if wizardNonInteractive {
		if wizardConfig == "" {
			fmt.Println("Error: --config required for non-interactive mode")
			os.Exit(1)
		}
		fmt.Println("Non-interactive mode not yet implemented")
		os.Exit(1)
	}

	// Run wizard
	fmt.Printf("Starting wizard: %s v%s\n", wizard.Name, wizard.Version)
	fmt.Printf("Description: %s\n\n", wizard.Description)

	result, err := wizardEngine.ExecuteWizard(wizard)
	if err != nil {
		fmt.Printf("Wizard failed: %v\n", err)
		os.Exit(1)
	}

	if result.Completed {
		fmt.Printf("\n🎉 Wizard completed successfully in %v\n", result.Duration)
	} else {
		fmt.Printf("\n⚠️  Wizard was interrupted\n")
		os.Exit(1)
	}
}

func listWizards(cmd *cobra.Command, args []string) {
	fmt.Println("Available Wizards:")
	fmt.Println("==================")

	// Built-in wizards
	builtinWizards := []struct {
		name        string
		description string
		file        string
	}{
		{"database-setup", "Install and configure database systems", "examples/wizards/database-setup.yaml"},
		{"dev-environment", "Set up development environment", "examples/wizards/dev-environment.yaml"},
	}

	fmt.Println("\n📦 Built-in Wizards:")
	for _, w := range builtinWizards {
		fmt.Printf("  %-20s %s\n", w.name, w.description)
	}

	// Check for user wizards directory
	userWizardsDir := filepath.Join(os.Getenv("HOME"), ".portunix", "wizards")
	if _, err := os.Stat(userWizardsDir); err == nil {
		fmt.Println("\n👤 User Wizards:")
		files, err := os.ReadDir(userWizardsDir)
		if err == nil {
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".yaml") {
					name := strings.TrimSuffix(file.Name(), ".yaml")
					fmt.Printf("  %-20s %s\n", name, file.Name())
				}
			}
		}
	} else {
		fmt.Println("\n💡 Tip: Create ~/.portunix/wizards/ directory to add custom wizards")
	}

	fmt.Println("\nUsage:")
	fmt.Println("  portunix wizard run <wizard-name>")
	fmt.Println("  portunix wizard run <path-to-yaml-file>")
}

func validateWizard(cmd *cobra.Command, args []string) {
	wizardPath := args[0]

	wizardEngine := engine.NewWizardEngine()
	wizard, err := wizardEngine.LoadWizard(wizardPath)
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Wizard '%s' is valid\n", wizard.Name)
	fmt.Printf("   ID: %s\n", wizard.ID)
	fmt.Printf("   Version: %s\n", wizard.Version)
	fmt.Printf("   Description: %s\n", wizard.Description)
	fmt.Printf("   Pages: %d\n", len(wizard.Pages))
}

func createWizard(cmd *cobra.Command, args []string) {
	wizardName := args[0]
	fileName := wizardName + ".yaml"

	if _, err := os.Stat(fileName); err == nil {
		fmt.Printf("Error: File '%s' already exists\n", fileName)
		os.Exit(1)
	}

	template := `wizard:
  id: "` + wizardName + `"
  name: "` + strings.Title(wizardName) + ` Setup Wizard"
  version: "1.0"
  description: "Description of your wizard"
  
  variables:
    # Define variables here
    
  pages:
    - id: "welcome"
      type: "info"
      title: "Welcome"
      content: |
        Welcome to the ` + wizardName + ` setup wizard!
        
        This wizard will guide you through the setup process.
      next:
        page: "complete"
    
    - id: "complete"
      type: "success"
      title: "Setup Complete!"
      content: |
        ✅ Setup completed successfully!
`

	err := os.WriteFile(fileName, []byte(template), 0644)
	if err != nil {
		fmt.Printf("Error creating wizard file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Created wizard template: %s\n", fileName)
	fmt.Printf("💡 Edit the file and run: portunix wizard validate %s\n", fileName)
}

func resolveWizardPath(input string) (string, error) {
	// If it's a file path (contains / or \\ or ends with .yaml)
	if strings.Contains(input, "/") || strings.Contains(input, "\\") || strings.HasSuffix(input, ".yaml") {
		if _, err := os.Stat(input); err != nil {
			return "", fmt.Errorf("wizard file not found: %s", input)
		}
		return input, nil
	}

	// Check built-in wizards
	builtinPath := filepath.Join("examples", "wizards", input+".yaml")
	if _, err := os.Stat(builtinPath); err == nil {
		return builtinPath, nil
	}

	// Check user wizards directory
	userWizardsDir := filepath.Join(os.Getenv("HOME"), ".portunix", "wizards")
	userPath := filepath.Join(userWizardsDir, input+".yaml")
	if _, err := os.Stat(userPath); err == nil {
		return userPath, nil
	}

	return "", fmt.Errorf("wizard not found: %s. Use 'portunix wizard list' to see available wizards", input)
}
