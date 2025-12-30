package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"portunix.ai/portunix/src/helpers/ptx-prompting/internal/prompt"
	"portunix.ai/portunix/src/shared"
)

var (
	noCopy            bool    // --no-copy flag to disable clipboard
	quiet             bool    // --quiet flag to disable stdout
	outputFile        string
	templateVars      map[string]string
	interactiveMode   bool
	allowIncomplete   bool
	showPreview       bool
	verboseOutput     bool
	defaultValuesFile string
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build [template-file]",
	Short: "Build prompt from template",
	Long: `Build a customized prompt from a template file.

The template file can contain placeholders in {placeholder} format.
You can provide parameter values via command-line flags or interactive input.

By default, the output is displayed on stdout AND copied to clipboard.

Examples:
  ptx-prompting build prompts/translate.md                    # Output to stdout AND clipboard
  ptx-prompting build prompts/translate.md --no-copy          # Only stdout, no clipboard
  ptx-prompting build prompts/translate.md --quiet            # Only clipboard, no stdout
  ptx-prompting build prompts/translate.md -o result.txt      # Save to file + clipboard
  ptx-prompting build prompts/review.md --var programming_language=Go --var file_path=main.go
  ptx-prompting build prompts/translate.md --interactive      # Force interactive mode
  ptx-prompting build prompts/translate.md --preview          # Preview without building`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateFile := args[0]
		return buildPrompt(templateFile, cmd, args)
	},
}

func buildPrompt(templateFile string, cmd *cobra.Command, args []string) error {
	builder := prompt.NewPromptBuilder()
	clipboardMgr := shared.NewClipboardManager()

	// Parse additional CLI variables from unknown flags
	additionalVars, err := parseAdditionalVariables(cmd, args)
	if err != nil {
		return fmt.Errorf("failed to parse additional variables: %v", err)
	}

	// Merge template variables
	allVars := make(map[string]string)
	if templateVars != nil {
		for k, v := range templateVars {
			allVars[k] = v
		}
	}
	for k, v := range additionalVars {
		allVars[k] = v
	}

	// Load default values if specified
	defaultVals, err := loadDefaultValues(defaultValuesFile)
	if err != nil {
		return fmt.Errorf("failed to load default values: %v", err)
	}

	// Note: Interactive mode will be automatically enabled in builder if there are missing variables

	// Build options
	options := &prompt.BuildOptions{
		TemplateFile:      templateFile,
		Variables:         allVars,
		InteractiveMode:   interactiveMode,
		AllowIncomplete:   allowIncomplete,
		DefaultValues:     defaultVals,
	}

	// Show preview if requested
	if showPreview {
		return showBuildPreview(builder, options)
	}

	// Build the prompt
	result, err := builder.Build(options)
	if err != nil {
		return fmt.Errorf("failed to build prompt: %v", err)
	}

	// Show verbose output if requested
	if verboseOutput {
		showVerboseOutput(result)
	}

	// Handle output
	if outputFile != "" {
		err = writeToFile(result.Content, outputFile)
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}
		fmt.Printf("✅ Prompt saved to: %s\n", outputFile)
	}

	// Default behavior: copy to clipboard unless --no-copy is specified
	if !noCopy {
		err = clipboardMgr.WriteWithFeedback(result.Content)
		if err != nil {
			// If clipboard fails, just warn but continue
			fmt.Printf("⚠️  Could not copy to clipboard: %v\n", err)
		}
	}

	// Default behavior: output to stdout unless --quiet is specified
	if !quiet && outputFile == "" {
		fmt.Println(result.Content)
	}

	// Show warnings if any
	if len(result.UnusedVariables) > 0 {
		fmt.Printf("⚠️  Unused variables: %s\n", strings.Join(result.UnusedVariables, ", "))
	}

	return nil
}

// parseAdditionalVariables parses additional variables from CLI flags
func parseAdditionalVariables(cmd *cobra.Command, args []string) (map[string]string, error) {
	vars := make(map[string]string)

	// Parse flags that look like --key=value or --key value
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		// Skip known flags
		knownFlags := map[string]bool{
			"copy": true, "output": true, "var": true, "interactive": true,
			"allow-incomplete": true, "preview": true, "verbose": true,
			"defaults": true,
		}
		if !knownFlags[flag.Name] {
			vars[flag.Name] = flag.Value.String()
		}
	})

	return vars, nil
}

// loadDefaultValues loads default values from a file
func loadDefaultValues(filename string) (map[string]string, error) {
	if filename == "" {
		return nil, nil
	}

	// TODO: Implement loading default values from JSON/YAML file
	fmt.Printf("ℹ️  Default values file support not yet implemented: %s\n", filename)
	return make(map[string]string), nil
}

// showBuildPreview shows what the build would do without actually building
func showBuildPreview(builder *prompt.PromptBuilder, options *prompt.BuildOptions) error {
	preview, err := builder.PreviewBuild(options)
	if err != nil {
		return err
	}

	fmt.Printf("📋 Template: %s (%s format)\n", preview.Template.FilePath, preview.Template.Format)
	fmt.Printf("📊 Placeholders found: %d\n", len(preview.Template.Placeholders))

	if len(preview.Template.Placeholders) > 0 {
		fmt.Println("📝 Placeholders:")
		for _, ph := range preview.Template.Placeholders {
			status := "❌ missing"
			if value, exists := preview.ResolvedVariables[ph]; exists {
				status = fmt.Sprintf("✅ %s", value)
			}
			fmt.Printf("  • %s: %s\n", ph, status)
		}
	}

	if len(preview.MissingVariables) > 0 {
		fmt.Printf("⚠️  Missing variables: %s\n", strings.Join(preview.MissingVariables, ", "))
		if preview.WillPrompt {
			fmt.Println("ℹ️  Interactive mode will prompt for missing values")
		}
	}

	if len(preview.UnusedVariables) > 0 {
		fmt.Printf("⚠️  Unused variables: %s\n", strings.Join(preview.UnusedVariables, ", "))
	}

	return nil
}

// showVerboseOutput shows detailed information about the build result
func showVerboseOutput(result *prompt.BuildResult) {
	fmt.Printf("📋 Template: %s (%s format)\n", result.Template.FilePath, result.Template.Format)
	fmt.Printf("📊 Content length: %d characters\n", len(result.Content))
	fmt.Printf("🔧 Variables resolved: %d\n", len(result.ResolvedVariables))

	if len(result.ResolvedVariables) > 0 {
		fmt.Println("📝 Resolved variables:")
		for key, value := range result.ResolvedVariables {
			fmt.Printf("  • %s: %s\n", key, value)
		}
	}
}

// writeToFile writes content to a file
func writeToFile(content, filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", filename, err)
	}

	return nil
}

func init() {
	buildCmd.Flags().BoolVar(&noCopy, "no-copy", false, "Don't copy to clipboard (default: copy to clipboard)")
	buildCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Don't output to stdout (default: output to stdout)")
	buildCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write output to file")
	buildCmd.Flags().StringToStringVar(&templateVars, "var", nil, "Set template variable (--var key=value)")
	buildCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", false, "Interactive mode for missing variables")
	buildCmd.Flags().BoolVar(&allowIncomplete, "allow-incomplete", false, "Allow build with missing variables")
	buildCmd.Flags().BoolVar(&showPreview, "preview", false, "Preview what will be built without building")
	buildCmd.Flags().BoolVarP(&verboseOutput, "verbose", "v", false, "Show verbose output")
	buildCmd.Flags().StringVar(&defaultValuesFile, "defaults", "", "Load default values from file")
}