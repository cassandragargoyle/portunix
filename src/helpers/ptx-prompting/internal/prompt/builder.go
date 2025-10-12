package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"portunix.ai/portunix/src/helpers/ptx-prompting/internal/parser"
)

// PromptBuilder handles building prompts from templates with parameter resolution
type PromptBuilder struct {
	parser *parser.TemplateParser
}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		parser: parser.NewTemplateParser(),
	}
}

// BuildOptions contains options for building a prompt
type BuildOptions struct {
	TemplateFile      string
	Variables         map[string]string
	InteractiveMode   bool
	AllowIncomplete   bool
	DefaultValues     map[string]string
}

// BuildResult contains the result of a prompt build operation
type BuildResult struct {
	Content           string
	Template          *parser.Template
	ResolvedVariables map[string]string
	MissingVariables  []string
	UnusedVariables   []string
	Success           bool
	Errors            []string
}

// Build builds a prompt from a template with the given options
func (pb *PromptBuilder) Build(options *BuildOptions) (*BuildResult, error) {
	result := &BuildResult{
		ResolvedVariables: make(map[string]string),
		Success:           false,
	}

	// Parse template
	template, err := pb.parser.ParseTemplate(options.TemplateFile)
	if err != nil {
		return result, fmt.Errorf("failed to parse template: %v", err)
	}

	result.Template = template

	// Validate template
	validationErrors := pb.parser.ValidateTemplate(template)
	if len(validationErrors) > 0 {
		result.Errors = validationErrors
		return result, fmt.Errorf("template validation failed: %v", validationErrors)
	}

	// Resolve variables
	resolvedVars, err := pb.resolveVariables(template, options)
	if err != nil {
		return result, fmt.Errorf("failed to resolve variables: %v", err)
	}

	result.ResolvedVariables = resolvedVars

	// Check for missing variables
	missingVars := pb.parser.GetMissingPlaceholders(template, resolvedVars)
	result.MissingVariables = missingVars

	// Check for unused variables
	unusedVars := pb.parser.GetUnusedValues(template, resolvedVars)
	result.UnusedVariables = unusedVars

	// If we have missing variables and incomplete builds are not allowed, fail
	if len(missingVars) > 0 && !options.AllowIncomplete {
		return result, fmt.Errorf("missing required variables: %v", missingVars)
	}

	// Replace placeholders
	content := pb.parser.ReplacePlaceholders(template, resolvedVars)
	result.Content = content
	result.Success = true

	return result, nil
}

// BuildFromContent builds a prompt from template content directly
func (pb *PromptBuilder) BuildFromContent(content string, variables map[string]string) (*BuildResult, error) {
	result := &BuildResult{
		ResolvedVariables: make(map[string]string),
		Success:           false,
	}

	// Parse template content
	template := pb.parser.ParseTemplateContent(content)
	result.Template = template

	// Validate template
	validationErrors := pb.parser.ValidateTemplate(template)
	if len(validationErrors) > 0 {
		result.Errors = validationErrors
		return result, fmt.Errorf("template validation failed: %v", validationErrors)
	}

	// Use provided variables
	if variables != nil {
		for k, v := range variables {
			result.ResolvedVariables[k] = v
		}
	}

	// Check for missing variables
	missingVars := pb.parser.GetMissingPlaceholders(template, result.ResolvedVariables)
	result.MissingVariables = missingVars

	// Check for unused variables
	unusedVars := pb.parser.GetUnusedValues(template, result.ResolvedVariables)
	result.UnusedVariables = unusedVars

	// Replace placeholders
	resultContent := pb.parser.ReplacePlaceholders(template, result.ResolvedVariables)
	result.Content = resultContent
	result.Success = true

	return result, nil
}

// resolveVariables resolves template variables from multiple sources
func (pb *PromptBuilder) resolveVariables(template *parser.Template, options *BuildOptions) (map[string]string, error) {
	resolved := make(map[string]string)

	// Start with default values if provided
	if options.DefaultValues != nil {
		for k, v := range options.DefaultValues {
			resolved[k] = v
		}
	}

	// Override with CLI variables if provided
	if options.Variables != nil {
		for k, v := range options.Variables {
			resolved[k] = v
		}
	}

	// If interactive mode is enabled, prompt for missing variables
	// Or if we have missing variables and incomplete builds are not allowed, automatically enable interactive mode
	missingVars := pb.parser.GetMissingPlaceholders(template, resolved)
	shouldPrompt := options.InteractiveMode || (len(missingVars) > 0 && !options.AllowIncomplete)

	if shouldPrompt && len(missingVars) > 0 {
		fmt.Println("\n📝 Interactive mode - Please provide values for template placeholders:")
		fmt.Println(strings.Repeat("─", 60))

		for i, placeholder := range missingVars {
			fmt.Printf("\n[%d/%d] ", i+1, len(missingVars))
			value, err := pb.promptForVariable(placeholder, resolved)
			if err != nil {
				return resolved, fmt.Errorf("failed to get interactive input for %s: %v", placeholder, err)
			}
			resolved[placeholder] = value
		}
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println()
	}

	return resolved, nil
}

// promptForVariable prompts the user to enter a value for a variable
func (pb *PromptBuilder) promptForVariable(placeholder string, existingVars map[string]string) (string, error) {
	// Display friendly name (convert snake_case to Title Case)
	friendlyName := strings.ReplaceAll(placeholder, "_", " ")
	friendlyName = strings.Title(friendlyName)

	fmt.Printf("📌 %s\n", friendlyName)

	// Show context help for common placeholders
	if help := getPlaceholderHelp(placeholder); help != "" {
		fmt.Printf("   ℹ️  %s\n", help)
	}

	// Show default value if available from context
	defaultVal := getDefaultValue(placeholder, existingVars)
	if defaultVal != "" {
		fmt.Printf("   Default: %s\n", defaultVal)
		fmt.Printf("   Enter value (or press Enter for default): ")
	} else {
		fmt.Printf("   Enter value: ")
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Trim whitespace and newlines
	value := strings.TrimSpace(input)

	// Use default if empty and default exists
	if value == "" && defaultVal != "" {
		value = defaultVal
		fmt.Printf("   ✓ Using default: %s\n", defaultVal)
	} else if value != "" {
		fmt.Printf("   ✓ Set to: %s\n", value)
	}

	return value, nil
}

// getPlaceholderHelp returns context help for common placeholders
func getPlaceholderHelp(placeholder string) string {
	helpMap := map[string]string{
		"source_file":         "Path to the source file to be processed",
		"target_file":         "Path where the output will be saved",
		"source_language":     "Language of the source content (e.g., Czech, English)",
		"target_language":     "Language to translate/convert to (e.g., English, Czech)",
		"programming_language": "Programming language being used (e.g., Go, Python, JavaScript)",
		"file_path":           "Path to the file being analyzed",
		"context_description": "Brief description of the context or purpose",
		"audience":            "Target audience for the content (e.g., developers, users)",
		"focus_area_1":        "First area to focus on during review",
		"focus_area_2":        "Second area to focus on during review",
		"focus_area_3":        "Third area to focus on during review",
	}

	// Try exact match first
	if help, exists := helpMap[placeholder]; exists {
		return help
	}

	// Try to identify by suffix
	if strings.HasSuffix(placeholder, "_file") {
		return "Path to a file"
	}
	if strings.HasSuffix(placeholder, "_path") {
		return "Path to a file or directory"
	}
	if strings.HasSuffix(placeholder, "_language") {
		return "Language name"
	}
	if strings.HasSuffix(placeholder, "_description") {
		return "Brief description"
	}

	return ""
}

// getDefaultValue suggests a default value based on context
func getDefaultValue(placeholder string, existingVars map[string]string) string {
	// Common defaults
	defaults := map[string]string{
		"target_language": "English",
		"source_language": "Czech",
		"programming_language": "Go",
		"audience": "developers",
	}

	if val, exists := defaults[placeholder]; exists {
		return val
	}

	// Smart defaults based on other variables
	if placeholder == "target_file" {
		if sourceFile, exists := existingVars["source_file"]; exists {
			// Suggest output file based on input file
			if strings.Contains(sourceFile, ".") {
				ext := sourceFile[strings.LastIndex(sourceFile, "."):]
				base := sourceFile[:strings.LastIndex(sourceFile, ".")]
				return base + "_translated" + ext
			}
		}
	}

	return ""
}

// GetTemplateInfo returns information about a template without building it
func (pb *PromptBuilder) GetTemplateInfo(templateFile string) (*TemplateInfo, error) {
	template, err := pb.parser.ParseTemplate(templateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %v", err)
	}

	validationErrors := pb.parser.ValidateTemplate(template)

	info := &TemplateInfo{
		FilePath:         template.FilePath,
		Format:           template.Format,
		Placeholders:     template.Placeholders,
		PlaceholderCount: len(template.Placeholders),
		IsValid:          len(validationErrors) == 0,
		ValidationErrors: validationErrors,
		ContentLength:    len(template.Content),
	}

	return info, nil
}

// TemplateInfo contains information about a template
type TemplateInfo struct {
	FilePath         string
	Format           string
	Placeholders     []string
	PlaceholderCount int
	IsValid          bool
	ValidationErrors []string
	ContentLength    int
}

// PreviewBuild shows what variables would be resolved without actually building
func (pb *PromptBuilder) PreviewBuild(options *BuildOptions) (*BuildPreview, error) {
	template, err := pb.parser.ParseTemplate(options.TemplateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %v", err)
	}

	resolved := make(map[string]string)

	// Apply default values
	if options.DefaultValues != nil {
		for k, v := range options.DefaultValues {
			resolved[k] = v
		}
	}

	// Apply CLI variables
	if options.Variables != nil {
		for k, v := range options.Variables {
			resolved[k] = v
		}
	}

	preview := &BuildPreview{
		Template:          template,
		ResolvedVariables: resolved,
		MissingVariables:  pb.parser.GetMissingPlaceholders(template, resolved),
		UnusedVariables:   pb.parser.GetUnusedValues(template, resolved),
		WillPrompt:        options.InteractiveMode,
	}

	return preview, nil
}

// BuildPreview contains information about what a build would do
type BuildPreview struct {
	Template          *parser.Template
	ResolvedVariables map[string]string
	MissingVariables  []string
	UnusedVariables   []string
	WillPrompt        bool
}