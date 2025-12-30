package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	createLang        string
	createDescription string
	createInteractive bool
	createDir         string
	createType        string
	createForce       bool
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [template-name]",
	Short: "Create new template",
	Long: `Create a new prompt template file.

The template will be created with basic structure and placeholder examples.
You can specify language, description, and template type.

Examples:
  ptx-prompting create review-template.md
  ptx-prompting create translate-template.md --lang cs --description "Czech translation template"
  ptx-prompting create custom-prompt.md --interactive
  ptx-prompting create debug-prompt.md --type debug --dir ./templates/en/`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]
		return createTemplate(templateName)
	},
}

func createTemplate(templateName string) error {
	// Interactive mode processing
	if createInteractive {
		return createTemplateInteractive(templateName)
	}

	// Determine full path
	fullPath, err := determineTemplatePath(templateName)
	if err != nil {
		return fmt.Errorf("failed to determine template path: %v", err)
	}

	// Check if file already exists
	if !createForce {
		if _, err := os.Stat(fullPath); err == nil {
			return fmt.Errorf("template already exists: %s (use --force to overwrite)", fullPath)
		}
	}

	// Generate template content
	content, err := generateTemplateContent()
	if err != nil {
		return fmt.Errorf("failed to generate template content: %v", err)
	}

	// Create directories if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Write template file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write template file: %v", err)
	}

	fmt.Printf("✅ Template created: %s\n", fullPath)
	fmt.Printf("📝 Language: %s\n", createLang)
	if createDescription != "" {
		fmt.Printf("📄 Description: %s\n", createDescription)
	}
	fmt.Printf("🔧 Type: %s\n", createType)

	return nil
}

func createTemplateInteractive(templateName string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("🎯 Creating template: %s\n\n", templateName)

	// Ask for language
	fmt.Printf("Language [%s]: ", createLang)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		createLang = strings.TrimSpace(input)
	}

	// Ask for description
	fmt.Print("Description: ")
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		createDescription = strings.TrimSpace(input)
	}

	// Ask for type
	fmt.Printf("Template type (translation, review, debug, custom) [%s]: ", createType)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		createType = strings.TrimSpace(input)
	}

	// Ask for directory
	defaultDir := filepath.Join("./templates", createLang)
	fmt.Printf("Directory [%s]: ", defaultDir)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		createDir = strings.TrimSpace(input)
	} else {
		createDir = defaultDir
	}

	// Confirm creation
	fmt.Printf("\n📋 Summary:\n")
	fmt.Printf("  Name: %s\n", templateName)
	fmt.Printf("  Language: %s\n", createLang)
	fmt.Printf("  Type: %s\n", createType)
	fmt.Printf("  Directory: %s\n", createDir)
	if createDescription != "" {
		fmt.Printf("  Description: %s\n", createDescription)
	}

	fmt.Print("\nCreate template? (y/N): ")
	if input, _ := reader.ReadString('\n'); strings.ToLower(strings.TrimSpace(input)) != "y" {
		fmt.Println("❌ Template creation cancelled")
		return nil
	}

	// Create the template
	return createTemplate(templateName)
}

func determineTemplatePath(templateName string) (string, error) {
	if createDir != "" {
		return filepath.Join(createDir, templateName), nil
	}

	// Default path based on language
	defaultDir := filepath.Join("./templates", createLang)
	return filepath.Join(defaultDir, templateName), nil
}

func generateTemplateContent() (string, error) {
	switch createType {
	case "translation":
		return generateTranslationTemplate(), nil
	case "review":
		return generateReviewTemplate(), nil
	case "debug":
		return generateDebugTemplate(), nil
	default:
		return generateCustomTemplate(), nil
	}
}

func generateTranslationTemplate() string {
	desc := createDescription
	if desc == "" {
		desc = "Translation request template"
	}

	return fmt.Sprintf(`# %s

Please translate the file {source_file} from {source_language} to {target_language}.
Save the result as {target_file}.

## Requirements:
- Preserve all formatting and structure
- Keep technical terms consistent
- Maintain the original tone and style
- Target audience: {audience}

## Additional context:
{context}

## Special instructions:
{special_instructions}
`, desc)
}

func generateReviewTemplate() string {
	desc := createDescription
	if desc == "" {
		desc = "Code review request template"
	}

	return fmt.Sprintf(`# %s

Please review the following code changes in {file_path}:

## Focus areas:
- {focus_area_1}
- {focus_area_2}
- {focus_area_3}

## Technical details:
- Programming language: {programming_language}
- Framework/Libraries: {framework}
- Context: {context_description}

## Review criteria:
- Code quality and readability
- Performance considerations
- Security implications
- Best practices adherence
- Testing coverage

## Additional notes:
{additional_notes}
`, desc)
}

func generateDebugTemplate() string {
	desc := createDescription
	if desc == "" {
		desc = "Debug assistance template"
	}

	return fmt.Sprintf(`# %s

I'm experiencing an issue with {problem_component} and need help debugging.

## Problem description:
{problem_description}

## Expected behavior:
{expected_behavior}

## Actual behavior:
{actual_behavior}

## Environment:
- Language/Framework: {language_framework}
- Version: {version}
- Operating System: {operating_system}

## Steps to reproduce:
{reproduction_steps}

## Error messages:
{error_messages}

## What I've tried:
{attempted_solutions}

## Additional context:
{additional_context}
`, desc)
}

func generateCustomTemplate() string {
	desc := createDescription
	if desc == "" {
		desc = "Custom prompt template"
	}

	return fmt.Sprintf(`# %s

{prompt_content}

## Variables:
- {variable_1}: {description_1}
- {variable_2}: {description_2}
- {variable_3}: {description_3}

## Context:
{context}

## Additional notes:
{notes}
`, desc)
}

func init() {
	createCmd.Flags().StringVar(&createLang, "lang", "en", "Template language")
	createCmd.Flags().StringVar(&createDescription, "description", "", "Template description")
	createCmd.Flags().BoolVarP(&createInteractive, "interactive", "i", false, "Interactive template creation")
	createCmd.Flags().StringVar(&createDir, "dir", "", "Target directory for template")
	createCmd.Flags().StringVar(&createType, "type", "custom", "Template type (translation, review, debug, custom)")
	createCmd.Flags().BoolVar(&createForce, "force", false, "Overwrite existing template")
}