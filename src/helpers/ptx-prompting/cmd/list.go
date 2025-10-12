package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"portunix.ai/portunix/src/helpers/ptx-prompting/internal/prompt"
)

var (
	listLang       string
	listPath       string
	listDetailed   bool
	listBuiltinOnly bool
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Long: `List all available prompt templates.

Templates are searched in the following locations:
- Built-in templates (embedded in binary)
- User templates in current directory (./templates/)
- Custom template paths

Examples:
  ptx-prompting list
  ptx-prompting list --lang cs
  ptx-prompting list --path /custom/templates
  ptx-prompting list --detailed`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listTemplates()
	},
}

type TemplateEntry struct {
	Name         string
	Path         string
	Language     string
	Format       string
	IsBuiltin    bool
	Placeholders int
	Size         int64
	Error        string
}

func listTemplates() error {
	templates, err := findAllTemplates()
	if err != nil {
		return fmt.Errorf("failed to find templates: %v", err)
	}

	// Filter by language if specified
	if listLang != "" {
		filtered := make([]TemplateEntry, 0)
		for _, tmpl := range templates {
			if tmpl.Language == listLang {
				filtered = append(filtered, tmpl)
			}
		}
		templates = filtered
	}

	// Filter builtin only if specified
	if listBuiltinOnly {
		filtered := make([]TemplateEntry, 0)
		for _, tmpl := range templates {
			if tmpl.IsBuiltin {
				filtered = append(filtered, tmpl)
			}
		}
		templates = filtered
	}

	if len(templates) == 0 {
		fmt.Println("No templates found")
		if listLang != "" {
			fmt.Printf("Try without --lang filter or check available languages\n")
		}
		return nil
	}

	// Sort templates by name
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})

	if listDetailed {
		showDetailedList(templates)
	} else {
		showSimpleList(templates)
	}

	return nil
}

func findAllTemplates() ([]TemplateEntry, error) {
	var templates []TemplateEntry

	// Find built-in templates (embedded in binary)
	builtinTemplates := findBuiltinTemplates()
	templates = append(templates, builtinTemplates...)

	// Find user templates in current directory
	userTemplates, err := findUserTemplates("./templates")
	if err == nil {
		templates = append(templates, userTemplates...)
	}

	// Find templates in custom path if specified
	if listPath != "" {
		customTemplates, err := findUserTemplates(listPath)
		if err != nil {
			return templates, fmt.Errorf("failed to search custom path %s: %v", listPath, err)
		}
		templates = append(templates, customTemplates...)
	}

	return templates, nil
}

func findBuiltinTemplates() []TemplateEntry {
	// TODO: In production, these would be embedded in the binary
	// For now, return example built-in templates
	return []TemplateEntry{
		{
			Name:         "translate",
			Path:         "builtin://templates/en/translate.md",
			Language:     "en",
			Format:       "md",
			IsBuiltin:    true,
			Placeholders: 4, // source_file, target_file, target_language, audience
		},
		{
			Name:         "review",
			Path:         "builtin://templates/en/review.md",
			Language:     "en",
			Format:       "md",
			IsBuiltin:    true,
			Placeholders: 5, // file_path, focus_area_1, focus_area_2, focus_area_3, programming_language, context_description
		},
		{
			Name:         "translate",
			Path:         "builtin://templates/cs/translate.md",
			Language:     "cs",
			Format:       "md",
			IsBuiltin:    true,
			Placeholders: 4,
		},
	}
}

func findUserTemplates(basePath string) ([]TemplateEntry, error) {
	var templates []TemplateEntry

	// Check if path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return templates, nil // No user templates, not an error
	}

	builder := prompt.NewPromptBuilder()

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		if info.IsDir() {
			return nil
		}

		// Check if it's a template file
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" && ext != ".yaml" && ext != ".yml" {
			return nil
		}

		// Get relative path for name
		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			relPath = filepath.Base(path)
		}

		// Determine language from path
		lang := "unknown"
		pathParts := strings.Split(relPath, string(filepath.Separator))
		if len(pathParts) > 1 {
			possibleLang := pathParts[0]
			if len(possibleLang) == 2 { // Assume 2-letter language code
				lang = possibleLang
			}
		}

		// Get template info
		templateInfo, err := builder.GetTemplateInfo(path)

		template := TemplateEntry{
			Name:         strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
			Path:         path,
			Language:     lang,
			Format:       strings.TrimPrefix(ext, "."),
			IsBuiltin:    false,
			Size:         info.Size(),
		}

		if err != nil {
			template.Error = err.Error()
		} else {
			template.Placeholders = templateInfo.PlaceholderCount
		}

		templates = append(templates, template)
		return nil
	})

	return templates, err
}

func showSimpleList(templates []TemplateEntry) {
	fmt.Printf("Found %d template(s):\n\n", len(templates))

	// Group by language
	byLang := make(map[string][]TemplateEntry)
	for _, tmpl := range templates {
		byLang[tmpl.Language] = append(byLang[tmpl.Language], tmpl)
	}

	// Show grouped by language
	languages := make([]string, 0, len(byLang))
	for lang := range byLang {
		languages = append(languages, lang)
	}
	sort.Strings(languages)

	for _, lang := range languages {
		fmt.Printf("📁 %s/\n", lang)
		for _, tmpl := range byLang[lang] {
			source := "user"
			if tmpl.IsBuiltin {
				source = "builtin"
			}

			status := "✅"
			if tmpl.Error != "" {
				status = "❌"
			}

			fmt.Printf("  %s %-20s (%s, %d placeholders, %s)\n",
				status, tmpl.Name, tmpl.Format, tmpl.Placeholders, source)
		}
		fmt.Println()
	}
}

func showDetailedList(templates []TemplateEntry) {
	fmt.Printf("Found %d template(s) with details:\n\n", len(templates))

	for i, tmpl := range templates {
		if i > 0 {
			fmt.Println("─────────────────────────────────────────")
		}

		status := "✅ Valid"
		if tmpl.Error != "" {
			status = fmt.Sprintf("❌ Error: %s", tmpl.Error)
		}

		source := "👤 User"
		if tmpl.IsBuiltin {
			source = "📦 Built-in"
		}

		fmt.Printf("📝 %s (%s)\n", tmpl.Name, tmpl.Language)
		fmt.Printf("   Path: %s\n", tmpl.Path)
		fmt.Printf("   Format: %s | Placeholders: %d | %s\n", tmpl.Format, tmpl.Placeholders, source)
		if tmpl.Size > 0 {
			fmt.Printf("   Size: %d bytes\n", tmpl.Size)
		}
		fmt.Printf("   Status: %s\n", status)
	}
}

func init() {
	listCmd.Flags().StringVar(&listLang, "lang", "", "Filter templates by language (en, cs, etc.)")
	listCmd.Flags().StringVar(&listPath, "path", "", "Search templates in custom path")
	listCmd.Flags().BoolVarP(&listDetailed, "detailed", "d", false, "Show detailed information")
	listCmd.Flags().BoolVar(&listBuiltinOnly, "builtin", false, "Show only built-in templates")
}