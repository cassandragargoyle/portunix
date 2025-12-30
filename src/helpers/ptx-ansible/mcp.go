package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MCPTools provides MCP integration for AI-assisted .ptxbook management
type MCPTools struct {
	// Configuration for MCP integration
	ToolsEnabled bool
	OutputDir    string
}

// MCPToolResult represents the result of an MCP tool operation
type MCPToolResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewMCPTools creates a new MCP tools instance
func NewMCPTools() *MCPTools {
	return &MCPTools{
		ToolsEnabled: true,
		OutputDir:    "./generated-playbooks",
	}
}

// GeneratePlaybookFromPrompt creates a .ptxbook file from natural language description
func (mcp *MCPTools) GeneratePlaybookFromPrompt(prompt string, metadata map[string]interface{}) (*MCPToolResult, error) {
	if !mcp.ToolsEnabled {
		return &MCPToolResult{
			Success: false,
			Error:   "MCP tools are not enabled",
		}, nil
	}

	// Create output directory if it doesn't exist
	os.MkdirAll(mcp.OutputDir, 0755)

	// Parse metadata
	name := "generated-playbook"
	description := "AI-generated playbook"
	if metadata != nil {
		if n, ok := metadata["name"].(string); ok {
			name = n
		}
		if d, ok := metadata["description"].(string); ok {
			description = d
		}
	}

	// Generate .ptxbook based on prompt analysis
	ptxbook, err := mcp.analyzePromptAndGeneratePlaybook(prompt, name, description)
	if err != nil {
		return &MCPToolResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to generate playbook: %v", err),
		}, err
	}

	// Save to file
	filename := fmt.Sprintf("%s.ptxbook", sanitizeFilename(name))
	filepath := filepath.Join(mcp.OutputDir, filename)

	if err := mcp.savePlaybookToFile(ptxbook, filepath); err != nil {
		return &MCPToolResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to save playbook: %v", err),
		}, err
	}

	return &MCPToolResult{
		Success: true,
		Message: fmt.Sprintf("Generated playbook: %s", filename),
		Data: map[string]interface{}{
			"filename": filename,
			"path":     filepath,
			"playbook": ptxbook,
		},
	}, nil
}

// analyzePromptAndGeneratePlaybook analyzes a natural language prompt and generates a .ptxbook
func (mcp *MCPTools) analyzePromptAndGeneratePlaybook(prompt, name, description string) (*PtxbookFile, error) {
	prompt = strings.ToLower(prompt)

	// Initialize playbook structure
	ptxbook := &PtxbookFile{
		APIVersion: "portunix.ai/v1",
		Kind:       "Playbook",
		Metadata: PtxbookMetadata{
			Name:        name,
			Description: description,
		},
		Spec: PtxbookSpec{
			Variables: make(map[string]interface{}),
		},
	}

	// Analyze prompt for package requirements
	packages := mcp.detectPackagesFromPrompt(prompt)
	if len(packages) > 0 {
		ptxbook.Spec.Portunix = &PtxbookPortunix{
			Packages: packages,
		}
	}

	// Analyze prompt for conditional logic
	conditions := mcp.detectConditionsFromPrompt(prompt)
	if len(conditions) > 0 {
		// Add conditional packages
		for condition, condPackages := range conditions {
			for i, pkg := range condPackages {
				pkg.When = condition
				if ptxbook.Spec.Portunix == nil {
					ptxbook.Spec.Portunix = &PtxbookPortunix{}
				}
				ptxbook.Spec.Portunix.Packages = append(ptxbook.Spec.Portunix.Packages, pkg)
				_ = i // Prevent unused variable error
			}
		}
	}

	// Analyze prompt for environment-specific requirements
	envVars := mcp.detectEnvironmentVariablesFromPrompt(prompt)
	if len(envVars) > 0 {
		ptxbook.Spec.Environment = envVars
	}

	// Detect if Ansible is needed
	if mcp.detectAnsibleRequirement(prompt) {
		ptxbook.Spec.Requirements = &PtxbookRequirements{
			Ansible: &AnsibleRequirements{
				MinVersion: "2.15.0",
			},
		}
		// Add placeholder Ansible section
		ptxbook.Spec.Ansible = &PtxbookAnsible{
			Playbooks: []AnsiblePlaybook{
				{
					Path: "./ansible/generated-playbook.yml",
				},
			},
		}
	}

	// Add rollback configuration for complex setups
	if mcp.shouldEnableRollback(prompt) {
		ptxbook.Spec.Rollback = &PtxbookRollback{
			Enabled:      true,
			PreserveLogs: true,
			Timeout:      "5m",
			OnFailure: []RollbackAction{
				{
					Type:        "command",
					Command:     "echo 'Rollback: Cleaning up generated configuration'",
					Description: "Clean up any generated configuration files",
				},
			},
		}
	}

	return ptxbook, nil
}

// detectPackagesFromPrompt analyzes prompt for package requirements
func (mcp *MCPTools) detectPackagesFromPrompt(prompt string) []PtxbookPackage {
	packages := make([]PtxbookPackage, 0)

	// Package detection patterns
	packagePatterns := map[string]PtxbookPackage{
		"java":         {Name: "java", Variant: "17"},
		"node":         {Name: "nodejs", Variant: "20"},
		"python":       {Name: "python", Variant: "3.13"},
		"docker":       {Name: "docker", Variant: "latest"},
		"vscode":       {Name: "vscode", Variant: "stable"},
		"go":           {Name: "go", Variant: "latest"},
		"git":          {Name: "git", Variant: "latest"},
		"ansible":      {Name: "ansible", Variant: "latest"},
		"powershell":   {Name: "powershell", Variant: "latest"},
		"chrome":       {Name: "chrome", Variant: "stable"},
		"claude-code":  {Name: "claude-code", Variant: "latest"},
	}

	for keyword, pkg := range packagePatterns {
		if strings.Contains(prompt, keyword) {
			packages = append(packages, pkg)
		}
	}

	// Detect development environment requests
	if strings.Contains(prompt, "development environment") || strings.Contains(prompt, "dev setup") {
		packages = append(packages, PtxbookPackage{Name: "java", Variant: "17"})
		packages = append(packages, PtxbookPackage{Name: "nodejs", Variant: "20"})
		packages = append(packages, PtxbookPackage{Name: "vscode", Variant: "stable"})
	}

	// Detect web development
	if strings.Contains(prompt, "web development") || strings.Contains(prompt, "frontend") {
		packages = append(packages, PtxbookPackage{Name: "nodejs", Variant: "20"})
		packages = append(packages, PtxbookPackage{Name: "chrome", Variant: "stable"})
	}

	return packages
}

// detectConditionsFromPrompt analyzes prompt for conditional requirements
func (mcp *MCPTools) detectConditionsFromPrompt(prompt string) map[string][]PtxbookPackage {
	conditions := make(map[string][]PtxbookPackage)

	// OS-specific conditions
	if strings.Contains(prompt, "on linux") || strings.Contains(prompt, "linux only") {
		conditions["os == 'linux'"] = []PtxbookPackage{
			{Name: "powershell", Variant: "latest", When: "os == 'linux'"},
		}
	}

	if strings.Contains(prompt, "on windows") || strings.Contains(prompt, "windows only") {
		conditions["os == 'windows'"] = []PtxbookPackage{
			{Name: "chocolatey", Variant: "latest", When: "os == 'windows'"},
		}
	}

	// Container-specific conditions
	if strings.Contains(prompt, "in container") || strings.Contains(prompt, "containerized") {
		conditions["is_container"] = []PtxbookPackage{
			{Name: "docker", Variant: "latest", When: "is_container"},
		}
	}

	return conditions
}

// detectEnvironmentVariablesFromPrompt extracts environment variables from prompt
func (mcp *MCPTools) detectEnvironmentVariablesFromPrompt(prompt string) map[string]interface{} {
	envVars := make(map[string]interface{})

	// Common environment detection
	if strings.Contains(prompt, "production") {
		envVars["environment"] = "production"
	} else if strings.Contains(prompt, "staging") {
		envVars["environment"] = "staging"
	} else {
		envVars["environment"] = "development"
	}

	// Architecture detection
	if strings.Contains(prompt, "arm64") || strings.Contains(prompt, "apple silicon") {
		envVars["arch"] = "arm64"
	}

	return envVars
}

// detectAnsibleRequirement determines if Ansible is needed
func (mcp *MCPTools) detectAnsibleRequirement(prompt string) bool {
	ansibleKeywords := []string{
		"ansible", "playbook", "orchestration", "configuration management",
		"deploy", "infrastructure", "automation", "provision",
	}

	for _, keyword := range ansibleKeywords {
		if strings.Contains(prompt, keyword) {
			return true
		}
	}

	return false
}

// shouldEnableRollback determines if rollback should be enabled
func (mcp *MCPTools) shouldEnableRollback(prompt string) bool {
	rollbackKeywords := []string{
		"production", "critical", "rollback", "safe", "backup",
		"enterprise", "mission critical", "high availability",
	}

	for _, keyword := range rollbackKeywords {
		if strings.Contains(prompt, keyword) {
			return true
		}
	}

	return false
}

// savePlaybookToFile saves a .ptxbook file to disk
func (mcp *MCPTools) savePlaybookToFile(ptxbook *PtxbookFile, filepath string) error {
	// Convert to YAML (simplified - in real implementation would use yaml package)
	yamlContent, err := mcp.convertToYAML(ptxbook)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, []byte(yamlContent), 0644)
}

// convertToYAML converts a PtxbookFile to YAML format
func (mcp *MCPTools) convertToYAML(ptxbook *PtxbookFile) (string, error) {
	// Simplified YAML generation (in production, use gopkg.in/yaml.v3)
	yaml := fmt.Sprintf(`# Generated by Portunix MCP Tools
# Created: %s

apiVersion: %s
kind: %s
metadata:
  name: "%s"
  description: "%s"

spec:
`, time.Now().Format("2006-01-02 15:04:05"), ptxbook.APIVersion, ptxbook.Kind, ptxbook.Metadata.Name, ptxbook.Metadata.Description)

	// Add variables
	if len(ptxbook.Spec.Variables) > 0 {
		yaml += "  variables:\n"
		for k, v := range ptxbook.Spec.Variables {
			yaml += fmt.Sprintf("    %s: \"%v\"\n", k, v)
		}
		yaml += "\n"
	}

	// Add environment
	if len(ptxbook.Spec.Environment) > 0 {
		yaml += "  environment:\n"
		for k, v := range ptxbook.Spec.Environment {
			yaml += fmt.Sprintf("    %s: \"%v\"\n", k, v)
		}
		yaml += "\n"
	}

	// Add requirements
	if ptxbook.Spec.Requirements != nil && ptxbook.Spec.Requirements.Ansible != nil {
		yaml += "  requirements:\n"
		yaml += "    ansible:\n"
		yaml += fmt.Sprintf("      min_version: \"%s\"\n\n", ptxbook.Spec.Requirements.Ansible.MinVersion)
	}

	// Add Portunix packages
	if ptxbook.Spec.Portunix != nil && len(ptxbook.Spec.Portunix.Packages) > 0 {
		yaml += "  portunix:\n"
		yaml += "    packages:\n"
		for _, pkg := range ptxbook.Spec.Portunix.Packages {
			yaml += fmt.Sprintf("      - name: \"%s\"\n", pkg.Name)
			if pkg.Variant != "" {
				yaml += fmt.Sprintf("        variant: \"%s\"\n", pkg.Variant)
			}
			if pkg.When != "" {
				yaml += fmt.Sprintf("        when: \"%s\"\n", pkg.When)
			}
		}
		yaml += "\n"
	}

	// Add Ansible section
	if ptxbook.Spec.Ansible != nil && len(ptxbook.Spec.Ansible.Playbooks) > 0 {
		yaml += "  ansible:\n"
		yaml += "    playbooks:\n"
		for _, playbook := range ptxbook.Spec.Ansible.Playbooks {
			yaml += fmt.Sprintf("      - path: \"%s\"\n", playbook.Path)
			if playbook.When != "" {
				yaml += fmt.Sprintf("        when: \"%s\"\n", playbook.When)
			}
		}
		yaml += "\n"
	}

	// Add rollback section
	if ptxbook.Spec.Rollback != nil && ptxbook.Spec.Rollback.Enabled {
		yaml += "  rollback:\n"
		yaml += "    enabled: true\n"
		yaml += fmt.Sprintf("    preserve_logs: %t\n", ptxbook.Spec.Rollback.PreserveLogs)
		if ptxbook.Spec.Rollback.Timeout != "" {
			yaml += fmt.Sprintf("    timeout: \"%s\"\n", ptxbook.Spec.Rollback.Timeout)
		}
		if len(ptxbook.Spec.Rollback.OnFailure) > 0 {
			yaml += "    on_failure:\n"
			for _, action := range ptxbook.Spec.Rollback.OnFailure {
				yaml += fmt.Sprintf("      - type: \"%s\"\n", action.Type)
				if action.Command != "" {
					yaml += fmt.Sprintf("        command: \"%s\"\n", action.Command)
				}
				if action.Description != "" {
					yaml += fmt.Sprintf("        description: \"%s\"\n", action.Description)
				}
			}
		}
	}

	return yaml, nil
}

// ValidatePlaybook validates a .ptxbook file and provides suggestions
func (mcp *MCPTools) ValidatePlaybook(filepath string) (*MCPToolResult, error) {
	ptxbook, err := ParsePtxbookFile(filepath)
	if err != nil {
		return &MCPToolResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse playbook: %v", err),
		}, err
	}

	suggestions := make([]string, 0)

	// Check for best practices
	if ptxbook.Spec.Rollback == nil {
		suggestions = append(suggestions, "Consider enabling rollback protection for safer execution")
	}

	if len(ptxbook.Spec.Variables) == 0 && len(ptxbook.Spec.Environment) == 0 {
		suggestions = append(suggestions, "Consider adding variables for better templating flexibility")
	}

	if ptxbook.Spec.Portunix != nil {
		for _, pkg := range ptxbook.Spec.Portunix.Packages {
			if pkg.Variant == "" {
				suggestions = append(suggestions, fmt.Sprintf("Package '%s' could benefit from explicit variant specification", pkg.Name))
			}
		}
	}

	return &MCPToolResult{
		Success: true,
		Message: "Playbook validation completed",
		Data: map[string]interface{}{
			"valid":       true,
			"suggestions": suggestions,
			"metadata":    ptxbook.Metadata,
		},
	}, nil
}

// ListPlaybooks lists available .ptxbook files
func (mcp *MCPTools) ListPlaybooks(directory string) (*MCPToolResult, error) {
	if directory == "" {
		directory = "."
	}

	playbooks := make([]map[string]interface{}, 0)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".ptxbook") {
			ptxbook, parseErr := ParsePtxbookFile(path)
			if parseErr == nil {
				playbooks = append(playbooks, map[string]interface{}{
					"path":        path,
					"name":        ptxbook.Metadata.Name,
					"description": ptxbook.Metadata.Description,
					"has_ansible": ptxbook.Spec.Ansible != nil,
					"has_rollback": ptxbook.Spec.Rollback != nil,
					"package_count": func() int {
						if ptxbook.Spec.Portunix != nil {
							return len(ptxbook.Spec.Portunix.Packages)
						}
						return 0
					}(),
				})
			}
		}

		return nil
	})

	if err != nil {
		return &MCPToolResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to scan directory: %v", err),
		}, err
	}

	return &MCPToolResult{
		Success: true,
		Message: fmt.Sprintf("Found %d playbooks", len(playbooks)),
		Data:    playbooks,
	}, nil
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	// Replace spaces and invalid characters
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ToLower(name)

	// Remove non-alphanumeric characters except hyphens and underscores
	result := ""
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			result += string(char)
		}
	}

	return result
}

// ExportMCPToolsManifest exports MCP tools manifest for integration with AI assistants
func (mcp *MCPTools) ExportMCPToolsManifest() (*MCPToolResult, error) {
	manifest := map[string]interface{}{
		"name":        "ptx-ansible-mcp-tools",
		"version":     "1.0.0",
		"description": "MCP tools for AI-assisted Portunix playbook management",
		"tools": []map[string]interface{}{
			{
				"name":        "generate_playbook",
				"description": "Generate a .ptxbook file from natural language description",
				"parameters": map[string]interface{}{
					"prompt": map[string]interface{}{
						"type":        "string",
						"description": "Natural language description of the desired infrastructure setup",
					},
					"metadata": map[string]interface{}{
						"type":        "object",
						"description": "Optional metadata (name, description)",
						"optional":    true,
					},
				},
			},
			{
				"name":        "validate_playbook",
				"description": "Validate a .ptxbook file and provide suggestions",
				"parameters": map[string]interface{}{
					"filepath": map[string]interface{}{
						"type":        "string",
						"description": "Path to the .ptxbook file to validate",
					},
				},
			},
			{
				"name":        "list_playbooks",
				"description": "List available .ptxbook files in a directory",
				"parameters": map[string]interface{}{
					"directory": map[string]interface{}{
						"type":        "string",
						"description": "Directory to search for playbooks (default: current directory)",
						"optional":    true,
					},
				},
			},
		},
		"capabilities": []string{
			"playbook_generation",
			"validation",
			"template_processing",
			"conditional_execution",
			"rollback_configuration",
		},
	}

	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to generate manifest: %v", err),
		}, err
	}

	// Save manifest file
	manifestPath := filepath.Join(mcp.OutputDir, "mcp-tools-manifest.json")
	os.MkdirAll(mcp.OutputDir, 0755)

	if err := os.WriteFile(manifestPath, manifestJSON, 0644); err != nil {
		return &MCPToolResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to save manifest: %v", err),
		}, err
	}

	return &MCPToolResult{
		Success: true,
		Message: "MCP tools manifest exported successfully",
		Data: map[string]interface{}{
			"manifest_path": manifestPath,
			"manifest":      manifest,
		},
	}, nil
}