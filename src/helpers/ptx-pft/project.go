package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"portunix.ai/portunix/src/helpers/ptx-pft/templates"
)

// ProjectTemplateData holds data for template rendering
type ProjectTemplateData struct {
	ProjectName string
}

// handleProjectCommand handles the project subcommand
func handleProjectCommand(args []string) {
	if len(args) == 0 {
		showProjectHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		handleProjectCreateCommand(subArgs)
	case "--help", "-h":
		showProjectHelp()
	default:
		fmt.Printf("Unknown project subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix pft project --help' for available commands")
	}
}

// handleProjectCreateCommand creates a new project with the specified template
func handleProjectCreateCommand(args []string) {
	var projectName string
	var projectPath string
	var templateName string = "qfd" // Default template

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--template", "-t":
			if i+1 < len(args) {
				templateName = args[i+1]
				i++
			}
		case "--path", "-p":
			if i+1 < len(args) {
				projectPath = args[i+1]
				i++
			}
		case "--help", "-h":
			showProjectCreateHelp()
			return
		default:
			// First non-flag argument is the project name
			if projectName == "" && !isFlag(args[i]) {
				projectName = args[i]
			}
		}
	}

	if projectName == "" {
		fmt.Println("Error: project name is required")
		fmt.Println()
		showProjectCreateHelp()
		return
	}

	// Validate template
	if templateName != "qfd" && templateName != "basic" {
		fmt.Printf("Error: unknown template '%s'. Available: qfd, basic\n", templateName)
		return
	}

	// Determine project path
	if projectPath == "" {
		cwd, _ := os.Getwd()
		// Sanitize project name for directory
		safeName := sanitizeDirectoryName(projectName)
		projectPath = filepath.Join(cwd, safeName)
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		return
	}
	projectPath = absPath

	// Check if directory already exists
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		fmt.Printf("Error: directory already exists: %s\n", projectPath)
		return
	}

	fmt.Printf("Creating project '%s' with template '%s'\n", projectName, templateName)
	fmt.Printf("Location: %s\n", projectPath)
	fmt.Println()

	if templateName == "qfd" {
		err = createQFDProject(projectPath, projectName)
	} else {
		err = createBasicProject(projectPath, projectName)
	}

	if err != nil {
		fmt.Printf("Error creating project: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("Project created successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", projectPath)
	fmt.Println("  git init")
	fmt.Println("  portunix pft configure")
}

// createQFDProject creates a project with full ISO 16355 QFD structure
func createQFDProject(projectPath, projectName string) error {
	data := ProjectTemplateData{ProjectName: projectName}

	// Create main directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Voice directories with subdirectories
	voiceConfigs := []struct {
		name           string
		readmeTemplate string
		subdirs        []string
	}{
		{"VoC", "qfd/VoC-README.md.tmpl", []string{"verbatims", "needs"}},
		{"VoB", "qfd/VoB-README.md.tmpl", []string{"verbatims", "needs"}},
		{"VoE", "qfd/VoE-README.md.tmpl", []string{"verbatims", "needs", "constraints"}},
		{"VoS", "qfd/VoS-README.md.tmpl", []string{"verbatims", "needs"}},
	}

	for _, vc := range voiceConfigs {
		voicePath := filepath.Join(projectPath, vc.name)
		if err := os.MkdirAll(voicePath, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", vc.name, err)
		}

		// Create subdirectories with .gitkeep
		for _, subdir := range vc.subdirs {
			subdirPath := filepath.Join(voicePath, subdir)
			if err := os.MkdirAll(subdirPath, 0755); err != nil {
				return fmt.Errorf("failed to create %s/%s directory: %w", vc.name, subdir, err)
			}
			// Create .gitkeep
			gitkeepPath := filepath.Join(subdirPath, ".gitkeep")
			if err := os.WriteFile(gitkeepPath, []byte{}, 0644); err != nil {
				return fmt.Errorf("failed to create .gitkeep in %s/%s: %w", vc.name, subdir, err)
			}
		}

		// Create README from template
		readme, err := renderTemplate(vc.readmeTemplate, data)
		if err != nil {
			return fmt.Errorf("failed to render %s README: %w", vc.name, err)
		}
		readmePath := filepath.Join(voicePath, "README.md")
		if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
			return fmt.Errorf("failed to write %s README: %w", vc.name, err)
		}

		fmt.Printf("  Created: %s/\n", vc.name)
	}

	// Create requirements directory
	reqPath := filepath.Join(projectPath, "requirements")
	if err := os.MkdirAll(reqPath, 0755); err != nil {
		return fmt.Errorf("failed to create requirements directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(reqPath, ".gitkeep"), []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to create .gitkeep in requirements: %w", err)
	}
	reqReadme, err := renderTemplate("qfd/requirements-README.md.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render requirements README: %w", err)
	}
	if err := os.WriteFile(filepath.Join(reqPath, "README.md"), []byte(reqReadme), 0644); err != nil {
		return fmt.Errorf("failed to write requirements README: %w", err)
	}
	fmt.Println("  Created: requirements/")

	// Create matrices directory
	matPath := filepath.Join(projectPath, "matrices")
	if err := os.MkdirAll(matPath, 0755); err != nil {
		return fmt.Errorf("failed to create matrices directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(matPath, ".gitkeep"), []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to create .gitkeep in matrices: %w", err)
	}
	matReadme, err := renderTemplate("qfd/matrices-README.md.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render matrices README: %w", err)
	}
	if err := os.WriteFile(filepath.Join(matPath, "README.md"), []byte(matReadme), 0644); err != nil {
		return fmt.Errorf("failed to write matrices README: %w", err)
	}
	fmt.Println("  Created: matrices/")

	// Create main README
	mainReadme, err := renderTemplate("qfd/README.md.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render main README: %w", err)
	}
	if err := os.WriteFile(filepath.Join(projectPath, "README.md"), []byte(mainReadme), 0644); err != nil {
		return fmt.Errorf("failed to write main README: %w", err)
	}
	fmt.Println("  Created: README.md")

	return nil
}

// createBasicProject creates a minimal project with only Voice directories
func createBasicProject(projectPath, projectName string) error {
	// Create main directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create basic Voice directories (lowercase for backward compatibility)
	voices := []string{"voc", "vos", "vob", "voe"}
	for _, voice := range voices {
		voicePath := filepath.Join(projectPath, voice)
		if err := os.MkdirAll(voicePath, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", voice, err)
		}
		// Create .gitkeep
		gitkeepPath := filepath.Join(voicePath, ".gitkeep")
		if err := os.WriteFile(gitkeepPath, []byte{}, 0644); err != nil {
			return fmt.Errorf("failed to create .gitkeep in %s: %w", voice, err)
		}
		fmt.Printf("  Created: %s/\n", voice)
	}

	// Create simple README
	readme := fmt.Sprintf("# %s\n\nProduct Feedback Tool project.\n\n## Structure\n\n- `voc/` - Voice of Customer\n- `vos/` - Voice of Stakeholder\n- `vob/` - Voice of Business\n- `voe/` - Voice of Engineering\n", projectName)
	if err := os.WriteFile(filepath.Join(projectPath, "README.md"), []byte(readme), 0644); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}
	fmt.Println("  Created: README.md")

	return nil
}

// renderTemplate renders a template from embedded files
func renderTemplate(templatePath string, data ProjectTemplateData) (string, error) {
	content, err := templates.QFDTemplates.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	tmpl, err := template.New(templatePath).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	return buf.String(), nil
}

// sanitizeDirectoryName converts project name to safe directory name
func sanitizeDirectoryName(name string) string {
	// Replace spaces and special chars with hyphens
	result := make([]byte, 0, len(name))
	lastWasHyphen := false

	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result = append(result, c)
			lastWasHyphen = false
		} else if !lastWasHyphen {
			result = append(result, '-')
			lastWasHyphen = true
		}
	}

	// Trim trailing hyphen
	if len(result) > 0 && result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}

	return string(result)
}

// isFlag checks if argument is a flag
func isFlag(arg string) bool {
	return len(arg) > 0 && arg[0] == '-'
}

func showProjectHelp() {
	fmt.Println("Usage: portunix pft project [command]")
	fmt.Println()
	fmt.Println("Project Management Commands:")
	fmt.Println()
	fmt.Println("  create <name>        Create new PFT project")
	fmt.Println()
	fmt.Println("Run 'portunix pft project create --help' for more details")
}

func showProjectCreateHelp() {
	content, err := templates.QFDTemplates.ReadFile("qfd/project-create-help.txt")
	if err != nil {
		fmt.Println("Error loading help text")
		return
	}
	fmt.Print(string(content))
}
