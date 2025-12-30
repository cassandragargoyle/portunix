package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Template represents a parsed template with metadata
type Template struct {
	FilePath     string
	Content      string
	Placeholders []string
	Format       string // md, yaml, txt
}

// TemplateParser handles parsing of template files
type TemplateParser struct {
	placeholderRegex *regexp.Regexp
}

// NewTemplateParser creates a new template parser
func NewTemplateParser() *TemplateParser {
	// Regex to match {placeholder} format as specified in ADR-017
	regex := regexp.MustCompile(`\{([^}]+)\}`)
	return &TemplateParser{
		placeholderRegex: regex,
	}
}

// ParseTemplate parses a template file and extracts placeholders
func (tp *TemplateParser) ParseTemplate(filePath string) (*Template, error) {
	// Read template file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %v", filePath, err)
	}

	// Determine file format
	format := tp.getFileFormat(filePath)

	// Extract placeholders
	placeholders := tp.extractPlaceholders(string(content))

	template := &Template{
		FilePath:     filePath,
		Content:      string(content),
		Placeholders: placeholders,
		Format:       format,
	}

	return template, nil
}

// ParseTemplateContent parses template content directly (without file)
func (tp *TemplateParser) ParseTemplateContent(content string) *Template {
	placeholders := tp.extractPlaceholders(content)

	return &Template{
		FilePath:     "",
		Content:      content,
		Placeholders: placeholders,
		Format:       "unknown",
	}
}

// extractPlaceholders finds all unique placeholders in the content, preserving order
func (tp *TemplateParser) extractPlaceholders(content string) []string {
	matches := tp.placeholderRegex.FindAllStringSubmatch(content, -1)

	// Use map to track uniqueness while preserving order
	placeholderSet := make(map[string]bool)
	var placeholders []string

	for _, match := range matches {
		if len(match) > 1 {
			placeholder := strings.TrimSpace(match[1])
			if placeholder != "" && !placeholderSet[placeholder] {
				placeholderSet[placeholder] = true
				placeholders = append(placeholders, placeholder)
			}
		}
	}

	return placeholders
}

// getFileFormat determines the file format based on extension
func (tp *TemplateParser) getFileFormat(filePath string) string {
	if strings.HasSuffix(strings.ToLower(filePath), ".md") {
		return "md"
	}
	if strings.HasSuffix(strings.ToLower(filePath), ".yaml") || strings.HasSuffix(strings.ToLower(filePath), ".yml") {
		return "yaml"
	}
	if strings.HasSuffix(strings.ToLower(filePath), ".txt") {
		return "txt"
	}
	return "unknown"
}

// ValidateTemplate validates that a template is well-formed
func (tp *TemplateParser) ValidateTemplate(template *Template) []string {
	var errors []string

	// Check if template has content
	if strings.TrimSpace(template.Content) == "" {
		errors = append(errors, "template is empty")
	}

	// Check for malformed placeholders (unclosed braces)
	openBraces := strings.Count(template.Content, "{")
	closeBraces := strings.Count(template.Content, "}")
	if openBraces != closeBraces {
		errors = append(errors, fmt.Sprintf("mismatched braces: %d open, %d close", openBraces, closeBraces))
	}

	// Check for nested placeholders (not supported)
	nestedRegex := regexp.MustCompile(`\{[^}]*\{[^}]*\}[^}]*\}`)
	if nestedRegex.MatchString(template.Content) {
		errors = append(errors, "nested placeholders are not supported")
	}

	// Check for empty placeholders
	emptyRegex := regexp.MustCompile(`\{\s*\}`)
	if emptyRegex.MatchString(template.Content) {
		errors = append(errors, "empty placeholders found")
	}

	return errors
}

// ReplacePlaceholders replaces placeholders in template content with provided values
func (tp *TemplateParser) ReplacePlaceholders(template *Template, values map[string]string) string {
	result := template.Content

	for placeholder, value := range values {
		placeholderPattern := fmt.Sprintf("{%s}", placeholder)
		result = strings.ReplaceAll(result, placeholderPattern, value)
	}

	return result
}

// GetMissingPlaceholders returns placeholders that don't have values provided
func (tp *TemplateParser) GetMissingPlaceholders(template *Template, values map[string]string) []string {
	var missing []string

	for _, placeholder := range template.Placeholders {
		if _, exists := values[placeholder]; !exists {
			missing = append(missing, placeholder)
		}
	}

	return missing
}

// GetUnusedValues returns values that don't match any placeholder in the template
func (tp *TemplateParser) GetUnusedValues(template *Template, values map[string]string) []string {
	placeholderSet := make(map[string]bool)
	for _, placeholder := range template.Placeholders {
		placeholderSet[placeholder] = true
	}

	var unused []string
	for key := range values {
		if !placeholderSet[key] {
			unused = append(unused, key)
		}
	}

	return unused
}