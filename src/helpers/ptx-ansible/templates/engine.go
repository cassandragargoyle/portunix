package templates

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"text/template"
)

// TemplateContext holds variables for template substitution
type TemplateContext struct {
	Name             string
	Engine           string
	Target           string
	OS               string
	Arch             string
	OSFamily         string
	ContainerRuntime string
}

// TemplateParameter represents a template parameter definition
type TemplateParameter struct {
	Name        string   `json:"name"`
	Required    bool     `json:"required"`
	Type        string   `json:"type"`
	Choices     []string `json:"choices,omitempty"`
	Default     string   `json:"default,omitempty"`
	Description string   `json:"description"`
}

// TemplateMetadata represents template metadata from metadata.json
type TemplateMetadata struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Version     string              `json:"version"`
	Parameters  []TemplateParameter `json:"parameters"`
	OSSupport   []string            `json:"os_support"`
	Tags        []string            `json:"tags"`
}

// TemplateInfo provides summary info for template listing
type TemplateInfo struct {
	Name        string
	Description string
	Engines     []string
}

// ListTemplates returns list of available templates
func ListTemplates() ([]TemplateInfo, error) {
	entries, err := TemplateFS.ReadDir("examples")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	var templates []TemplateInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Read metadata.json for this template
		metadataPath := path.Join("examples", entry.Name(), "metadata.json")
		metadataBytes, err := TemplateFS.ReadFile(metadataPath)
		if err != nil {
			// Skip templates without metadata
			continue
		}

		var metadata TemplateMetadata
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			continue
		}

		// Find available engines (files ending with .ptxbook.tmpl)
		var engines []string
		templateDir := path.Join("examples", entry.Name())
		files, _ := TemplateFS.ReadDir(templateDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".ptxbook.tmpl") {
				engineName := strings.TrimSuffix(f.Name(), ".ptxbook.tmpl")
				engines = append(engines, engineName)
			}
		}

		templates = append(templates, TemplateInfo{
			Name:        metadata.Name,
			Description: metadata.Description,
			Engines:     engines,
		})
	}

	return templates, nil
}

// GetTemplateMetadata returns detailed metadata for a template
func GetTemplateMetadata(templateName string) (*TemplateMetadata, error) {
	metadataPath := path.Join("examples", templateName, "metadata.json")
	metadataBytes, err := TemplateFS.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("template '%s' not found", templateName)
	}

	var metadata TemplateMetadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse template metadata: %w", err)
	}

	return &metadata, nil
}

// GenerateFromTemplate generates a .ptxbook file from a template
func GenerateFromTemplate(templateName, engine, projectName, target string) (string, error) {
	// Validate template exists
	metadata, err := GetTemplateMetadata(templateName)
	if err != nil {
		return "", err
	}

	// Validate engine parameter
	engineValid := false
	for _, param := range metadata.Parameters {
		if param.Name == "engine" {
			for _, choice := range param.Choices {
				if choice == engine {
					engineValid = true
					break
				}
			}
		}
	}
	if !engineValid {
		return "", fmt.Errorf("invalid engine '%s' for template '%s'", engine, templateName)
	}

	// Set default target if not provided
	if target == "" {
		target = "container"
	}

	// Validate target
	if target != "container" && target != "local" {
		return "", fmt.Errorf("invalid target '%s', must be 'container' or 'local'", target)
	}

	// Load template file
	templatePath := path.Join("examples", templateName, engine+".ptxbook.tmpl")
	templateBytes, err := TemplateFS.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("template file not found: %s", templatePath)
	}

	// Detect system info
	osName, arch, osFamily := detectSystemInfo()

	// Detect container runtime if target is container
	containerRuntime := ""
	if target == "container" {
		containerRuntime, _ = detectContainerRuntime()
		if containerRuntime == "" {
			containerRuntime = "docker" // Default assumption
		}
	}

	// Prepare context
	ctx := TemplateContext{
		Name:             projectName,
		Engine:           engine,
		Target:           target,
		OS:               osName,
		Arch:             arch,
		OSFamily:         osFamily,
		ContainerRuntime: containerRuntime,
	}

	// Parse and execute template
	tmpl, err := template.New("playbook").Parse(string(templateBytes))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var output strings.Builder
	if err := tmpl.Execute(&output, ctx); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return output.String(), nil
}

// detectSystemInfo detects OS information using portunix or fallback
func detectSystemInfo() (osName, arch, osFamily string) {
	// Try to use portunix system info
	cmd := exec.Command("portunix", "system", "info", "--format", "json")
	output, err := cmd.Output()
	if err == nil {
		var info struct {
			OS       string `json:"os"`
			Arch     string `json:"arch"`
			OSFamily string `json:"os_family"`
		}
		if json.Unmarshal(output, &info) == nil && info.OS != "" {
			return info.OS, info.Arch, info.OSFamily
		}
	}

	// Fallback to runtime detection
	osName = runtime.GOOS
	arch = runtime.GOARCH
	osFamily = getOSFamily(osName)
	return
}

// getOSFamily returns OS family from OS name
func getOSFamily(osName string) string {
	switch osName {
	case "windows":
		return "windows"
	case "darwin":
		return "darwin"
	default:
		return "linux"
	}
}

// detectContainerRuntime detects available container runtime
func detectContainerRuntime() (string, error) {
	// Try docker first
	if _, err := exec.LookPath("docker"); err == nil {
		return "docker", nil
	}

	// Try podman
	if _, err := exec.LookPath("podman"); err == nil {
		return "podman", nil
	}

	return "", fmt.Errorf("no container runtime found")
}

// WritePlaybook writes generated playbook content to a file
func WritePlaybook(content, outputPath string) error {
	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		return fmt.Errorf("file already exists: %s", outputPath)
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}
