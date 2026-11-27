/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"portunix.ai/app/version"
)

// platformNamePattern is the canonical identifier shape for a hosting platform
// declared in supported_platforms[] (e.g. "synapse", "pack", "agent"). Must match
// the regex in plugin-manifest.schema.json v1.1.0.
var platformNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// featureTokenPattern is the shape for capability tokens in
// supported_platforms[].features. Must match the regex in plugin-manifest.schema.json v1.1.0.
var featureTokenPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]*$`)

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

	return os.WriteFile(manifestPath, data, 0644) // #nosec G306 -- plugin manifest is a public asset
}

// supportedRuntimes lists the plugin runtimes the dispatcher can start
var supportedRuntimes = map[string]bool{"native": true, "java": true, "python": true, "node": true}

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
	if !supportedRuntimes[manifest.Plugin.Runtime] {
		return fmt.Errorf("unsupported runtime: %s (supported: native, java, python, node)", manifest.Plugin.Runtime)
	}

	// Validate Python wheel plugin configuration
	if manifest.Plugin.Wheel != "" {
		if manifest.Plugin.Runtime != "python" {
			return fmt.Errorf("wheel field requires runtime: python")
		}
		if !strings.HasSuffix(manifest.Plugin.Wheel, ".whl") {
			return fmt.Errorf("wheel file must have .whl extension: %s", manifest.Plugin.Wheel)
		}
	}

	// Validate extra_wheels requires wheel field
	if len(manifest.Plugin.ExtraWheels) > 0 && manifest.Plugin.Wheel == "" {
		return fmt.Errorf("extra_wheels requires wheel field to be set")
	}

	if err := ValidateOptionalExtras(manifest); err != nil {
		return err
	}

	if manifest.Plugin.PythonMinVersion != "" {
		if manifest.Plugin.Runtime != "python" {
			return fmt.Errorf("python_min_version can only be used with runtime: python")
		}
		// Bridge to runtime_version for prerequisite checking
		if manifest.Plugin.RuntimeVersion == "" {
			manifest.Plugin.RuntimeVersion = ">=" + manifest.Plugin.PythonMinVersion
		}
	}

	if manifest.Plugin.PythonVersion != "" {
		if manifest.Plugin.Runtime != "python" {
			return fmt.Errorf("python_version can only be used with runtime: python")
		}
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

	// Validate supported_platforms[] (schema v1.1.0, additive). Portunix checks
	// only the shape — platform_payload content is opaque and owned by the
	// target platform.
	if err := validateSupportedPlatforms(manifest.SupportedPlatforms); err != nil {
		return err
	}

	return nil
}

// validateSupportedPlatforms validates the shape of each supported_platforms[]
// entry. Any malformed entry fails plugin install so the target platform
// never sees a broken declaration at runtime.
func validateSupportedPlatforms(entries []SupportedPlatform) error {
	seen := make(map[string]struct{}, len(entries))
	for i, sp := range entries {
		if sp.Name == "" {
			return fmt.Errorf("supported_platforms[%d]: name is required", i)
		}
		if !platformNamePattern.MatchString(sp.Name) {
			return fmt.Errorf("supported_platforms[%d]: invalid name %q (must match %s)",
				i, sp.Name, platformNamePattern.String())
		}
		if _, dup := seen[sp.Name]; dup {
			return fmt.Errorf("supported_platforms: platform %q declared more than once", sp.Name)
		}
		seen[sp.Name] = struct{}{}

		var minVer, maxVer *semver
		if sp.MinVersion != "" {
			v, err := parseSemver(sp.MinVersion)
			if err != nil {
				return fmt.Errorf("supported_platforms[%s].min_version: %w", sp.Name, err)
			}
			minVer = &v
		}
		if sp.MaxVersion != "" {
			v, err := parseSemver(sp.MaxVersion)
			if err != nil {
				return fmt.Errorf("supported_platforms[%s].max_version: %w", sp.Name, err)
			}
			maxVer = &v
		}
		if minVer != nil && maxVer != nil && compareSemver(*minVer, *maxVer) > 0 {
			return fmt.Errorf("supported_platforms[%s]: min_version %q must not exceed max_version %q",
				sp.Name, sp.MinVersion, sp.MaxVersion)
		}

		featSeen := make(map[string]struct{}, len(sp.Features))
		for j, f := range sp.Features {
			if !featureTokenPattern.MatchString(f) {
				return fmt.Errorf("supported_platforms[%s].features[%d]: invalid feature %q (must match %s)",
					sp.Name, j, f, featureTokenPattern.String())
			}
			if _, dup := featSeen[f]; dup {
				return fmt.Errorf("supported_platforms[%s].features: feature %q declared more than once", sp.Name, f)
			}
			featSeen[f] = struct{}{}
		}

		// platform_payload must be a JSON object when present. Portunix does
		// not interpret its content — only that it is shaped as an object so
		// platforms receive a structured payload.
		if len(sp.PlatformPayload) > 0 {
			trimmed := bytes.TrimSpace(sp.PlatformPayload)
			if len(trimmed) == 0 || trimmed[0] != '{' {
				return fmt.Errorf("supported_platforms[%s].platform_payload: must be a JSON object", sp.Name)
			}
			var probe map[string]json.RawMessage
			if err := json.Unmarshal(sp.PlatformPayload, &probe); err != nil {
				return fmt.Errorf("supported_platforms[%s].platform_payload: invalid JSON object: %w", sp.Name, err)
			}
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
		// One entry per subcommand called as `portunix <plugin> <name>` (issue #257)
		Commands: []PluginCommand{
			{
				Name:        "hello",
				Description: "Print a greeting",
				Subcommands: []string{},
				Parameters: []PluginParameter{
					{Name: "name", Type: "string", Description: "Name to greet", DefaultValue: "World"},
				},
				Examples: []string{
					fmt.Sprintf("portunix %s hello", name),
					fmt.Sprintf("portunix %s hello Portunix", name),
				},
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
      "name": "hello",
      "description": "Print a greeting",
      "subcommands": [],
      "parameters": [
        {
          "name": "name",
          "type": "string",
          "description": "Name to greet",
          "required": false,
          "default_value": "World"
        }
      ],
      "examples": ["portunix %s hello", "portunix %s hello Portunix"]
    }
  ]
}`, name, description, author, name, version.ProductVersion, name, name)
	}
	return string(data)
}

// wheelExtraNamePattern restricts extra names to characters that are safe
// inside a pip requirement specifier ("<wheel>[name]")
var wheelExtraNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// extraSeparatorRun matches the separator runs that PEP 685 folds into "-"
var extraSeparatorRun = regexp.MustCompile(`[-_.]+`)

// ValidateOptionalExtras checks the optional_extras declaration of a wheel plugin
func ValidateOptionalExtras(manifest *PluginManifest) error {
	extras := manifest.Plugin.OptionalExtras
	if len(extras) == 0 {
		return nil
	}
	if manifest.Plugin.Wheel == "" {
		return fmt.Errorf("optional_extras requires wheel field to be set")
	}
	seen := make(map[string]bool, len(extras))
	for _, extra := range extras {
		if !wheelExtraNamePattern.MatchString(extra.Name) {
			return fmt.Errorf("invalid optional_extras name %q: must match %s", extra.Name, wheelExtraNamePattern.String())
		}
		key := NormalizeExtraName(extra.Name)
		if seen[key] {
			return fmt.Errorf("duplicate optional_extras name: %s", extra.Name)
		}
		seen[key] = true
	}
	return nil
}

// NormalizeExtraName returns the PEP 685 normalized form of an extra name
func NormalizeExtraName(name string) string {
	return strings.ToLower(extraSeparatorRun.ReplaceAllString(name, "-"))
}

// ResolveWheelExtras checks the requested extras (entries may be comma
// separated) against the manifest declaration and returns them as declared,
// de-duplicated and sorted
func ResolveWheelExtras(manifest *PluginManifest, requested []string) ([]string, error) {
	if len(requested) == 0 {
		return nil, nil
	}
	declared := make(map[string]string, len(manifest.Plugin.OptionalExtras))
	var names []string
	for _, extra := range manifest.Plugin.OptionalExtras {
		declared[NormalizeExtraName(extra.Name)] = extra.Name
		names = append(names, extra.Name)
	}

	// Accept comma separated values as well as separate entries
	var items []string
	for _, raw := range requested {
		items = append(items, strings.Split(raw, ",")...)
	}

	chosen := make(map[string]bool)
	var resolved []string
	for _, raw := range items {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		canonical, ok := declared[NormalizeExtraName(name)]
		if !ok {
			if len(names) == 0 {
				return nil, fmt.Errorf("plugin %s declares no optional extras, so --extra %s cannot be installed", manifest.Name, name)
			}
			return nil, fmt.Errorf("plugin %s does not declare extra %q (declared extras: %s)", manifest.Name, name, strings.Join(names, ", "))
		}
		if !chosen[canonical] {
			chosen[canonical] = true
			resolved = append(resolved, canonical)
		}
	}
	sort.Strings(resolved)
	return resolved, nil
}
