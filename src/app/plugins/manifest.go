package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"portunix.ai/app/version"
)

// LoadManifest loads a plugin manifest from a JSON file
func LoadManifest(manifestPath string) (*PluginManifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	// Validate manifest
	if err := ValidateManifest(&manifest); err != nil {
		return nil, fmt.Errorf("manifest validation failed: %w", err)
	}

	return &manifest, nil
}

// SaveManifest saves a plugin manifest to a JSON file
func SaveManifest(manifest *PluginManifest, manifestPath string) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode manifest: %w", err)
	}

	return os.WriteFile(manifestPath, data, 0644)
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

	// Validate plugin type (grpc = long-running service, helper = CLI executable)
	validTypes := map[string]bool{"grpc": true, "helper": true, "executable": true}
	if !validTypes[manifest.Plugin.Type] {
		return fmt.Errorf("unsupported plugin type: %s (supported: grpc, helper, executable)", manifest.Plugin.Type)
	}

	// Validate runtime (default to native if not specified)
	if manifest.Plugin.Runtime == "" {
		manifest.Plugin.Runtime = "native"
	}
	if manifest.Plugin.Runtime != "native" && manifest.Plugin.Runtime != "java" && manifest.Plugin.Runtime != "python" {
		return fmt.Errorf("unsupported runtime: %s (supported: native, java, python)", manifest.Plugin.Runtime)
	}

	// Validate port range (only required for gRPC service plugins)
	if manifest.Plugin.Type == "grpc" {
		if manifest.Plugin.Port < 9000 || manifest.Plugin.Port > 9999 {
			return fmt.Errorf("plugin port must be between 9000-9999, got: %d", manifest.Plugin.Port)
		}
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
			Runtime:             "native",
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

// GetManifestTemplate returns a template manifest as JSON string
func GetManifestTemplate(name, description, author string) string {
	manifest := CreateDefaultManifest(name, description, author)
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		// Fallback to hardcoded template if marshaling fails
		return fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "%s",
  "author": "%s",
  "license": "MIT",
  "plugin": {
    "type": "grpc",
    "binary": "./%s",
    "runtime": "native",
    "port": 9001,
    "health_check_interval": 30000000000
  },
  "dependencies": {
    "portunix_min_version": "%s",
    "os_support": ["linux", "windows", "darwin"]
  },
  "ai_integration": {
    "mcp_tools": []
  },
  "permissions": {
    "filesystem": ["read"],
    "network": ["outbound"],
    "database": [],
    "system": [],
    "level": "limited"
  },
  "commands": [
    {
      "name": "help",
      "description": "Show help for %s plugin",
      "subcommands": [],
      "parameters": [],
      "examples": ["portunix %s help"]
    }
  ]
}`, name, description, author, name, version.ProductVersion, name, name)
	}
	return string(data)
}
