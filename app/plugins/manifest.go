package plugins

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"portunix.cz/app/version"
)

// LoadManifest loads a plugin manifest from a YAML file
func LoadManifest(manifestPath string) (*PluginManifest, error) {
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest file: %w", err)
	}
	defer file.Close()

	var manifest PluginManifest
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	// Validate manifest
	if err := ValidateManifest(&manifest); err != nil {
		return nil, fmt.Errorf("manifest validation failed: %w", err)
	}

	return &manifest, nil
}

// SaveManifest saves a plugin manifest to a YAML file
func SaveManifest(manifest *PluginManifest, manifestPath string) error {
	file, err := os.Create(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to create manifest file: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	return encoder.Encode(manifest)
}

// ValidateManifest validates a plugin manifest
func ValidateManifest(manifest *PluginManifest) error {
	// Required fields
	if manifest.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if manifest.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	if manifest.Description == "" {
		return fmt.Errorf("plugin description is required")
	}
	if manifest.Author == "" {
		return fmt.Errorf("plugin author is required")
	}
	if manifest.Plugin.Binary == "" {
		return fmt.Errorf("plugin binary is required")
	}

	// Validate plugin type
	if manifest.Plugin.Type != "grpc" {
		return fmt.Errorf("unsupported plugin type: %s (only 'grpc' is supported)", manifest.Plugin.Type)
	}

	// Validate port range
	if manifest.Plugin.Port < 9000 || manifest.Plugin.Port > 9999 {
		return fmt.Errorf("plugin port must be between 9000-9999, got: %d", manifest.Plugin.Port)
	}

	// Validate health check interval
	if manifest.Plugin.HealthCheckInterval == 0 {
		manifest.Plugin.HealthCheckInterval = 30 * time.Second // Default
	}

	// Validate OS support
	if len(manifest.Dependencies.OSSupport) == 0 {
		return fmt.Errorf("at least one supported OS is required")
	}
	for _, os := range manifest.Dependencies.OSSupport {
		if os != "linux" && os != "windows" && os != "darwin" {
			return fmt.Errorf("unsupported OS: %s (supported: linux, windows, darwin)", os)
		}
	}

	// Validate permission level
	if manifest.Permissions.Level == "" {
		manifest.Permissions.Level = "limited" // Default
	}
	if manifest.Permissions.Level != "limited" &&
		manifest.Permissions.Level != "standard" &&
		manifest.Permissions.Level != "full" {
		return fmt.Errorf("invalid permission level: %s (valid: limited, standard, full)", manifest.Permissions.Level)
	}

	// Validate commands
	for i, cmd := range manifest.Commands {
		if cmd.Name == "" {
			return fmt.Errorf("command %d: name is required", i)
		}
		if cmd.Description == "" {
			return fmt.Errorf("command %s: description is required", cmd.Name)
		}

		// Validate parameters
		for j, param := range cmd.Parameters {
			if param.Name == "" {
				return fmt.Errorf("command %s, parameter %d: name is required", cmd.Name, j)
			}
			if param.Type == "" {
				param.Type = "string" // Default
			}
			if param.Type != "string" && param.Type != "int" && param.Type != "bool" && param.Type != "array" {
				return fmt.Errorf("command %s, parameter %s: invalid type %s", cmd.Name, param.Name, param.Type)
			}
		}
	}

	// Validate MCP tools
	for i, tool := range manifest.AIIntegration.MCPTools {
		if tool.Name == "" {
			return fmt.Errorf("MCP tool %d: name is required", i)
		}
		if tool.Description == "" {
			return fmt.Errorf("MCP tool %s: description is required", tool.Name)
		}
	}

	return nil
}

// CreateDefaultManifest creates a default plugin manifest
func CreateDefaultManifest(name, description, author string) *PluginManifest {
	return &PluginManifest{
		Name:        name,
		Version:     "1.0.0",
		Description: description,
		Author:      author,
		License:     "MIT",
		Plugin: PluginBinaryConfig{
			Type:                "grpc",
			Binary:              fmt.Sprintf("./%s", name),
			Port:                9001,
			HealthCheckInterval: 30 * time.Second,
		},
		Dependencies: PluginDependencies{
			PortunixMinVersion: version.ProductVersion,
			OSSupport:          []string{"linux", "windows", "darwin"},
		},
		AIIntegration: AIIntegrationConfig{
			MCPTools: []MCPTool{},
		},
		Permissions: PluginPermissions{
			Filesystem: []string{"read"},
			Network:    []string{"outbound"},
			Database:   []string{},
			System:     []string{},
			Level:      "limited",
		},
		Commands: []PluginCommand{
			{
				Name:        "help",
				Description: fmt.Sprintf("Show help for %s plugin", name),
				Subcommands: []string{},
				Parameters:  []PluginParameter{},
				Examples:    []string{fmt.Sprintf("portunix %s help", name)},
			},
		},
	}
}

// GetManifestTemplate returns a template manifest as YAML string
func GetManifestTemplate(name, description, author string) string {
	_ = CreateDefaultManifest(name, description, author) // Future use

	return fmt.Sprintf(`# Plugin identification
name: "%s"
version: "1.0.0"
description: "%s"
author: "%s"
license: "MIT"

# Plugin configuration
plugin:
  type: "grpc"
  binary: "./%s"
  port: 9001
  health_check_interval: 30s
  
# Dependencies
dependencies:
  portunix_min_version: "%s"
  os_support: ["linux", "windows", "darwin"]
  
# AI Integration
ai_integration:
  mcp_tools:
    - name: "example_tool"
      description: "Example MCP tool"
  
# Permissions
permissions:
  filesystem: ["read"]
  network: ["outbound"]
  database: []
  system: []
  level: "limited"
  
# Commands exposed to Portunix CLI
commands:
  - name: "help"
    description: "Show help for %s plugin"
    subcommands: []
    parameters: []
    examples:
      - "portunix %s help"
`, name, description, author, name, version.ProductVersion, name, name)
}
