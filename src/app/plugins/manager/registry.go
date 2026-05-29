/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package manager handles plugin lifecycle and registry persistence
// Manifest field mapping aligned with plugin-manifest.schema.json v1.0.0 (api/contract)
package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"portunix.ai/app/plugins"
)

// Registry manages the plugin registry file
type Registry struct {
	filePath string
	data     *RegistryData
	mutex    sync.RWMutex
}

// RegistryData represents the structure of the registry file
type RegistryData struct {
	Version    string                     `json:"version"`
	LastUpdate time.Time                  `json:"last_update"`
	Plugins    map[string]*RegistryPlugin `json:"plugins"`
}

// RegistryPlugin represents a plugin entry in the registry
type RegistryPlugin struct {
	Name                string                      `json:"name"`
	Version             string                      `json:"version"`
	Description         string                      `json:"description"`
	Author              string                      `json:"author"`
	License             string                      `json:"license"`
	InstallPath         string                      `json:"install_path"`
	BinaryName          string                      `json:"binary_name"`
	Mode                string                      `json:"mode"`                         // service or helper
	Runtime             string                      `json:"runtime"`                      // native, java, python
	RuntimeVersion      string                      `json:"runtime_version"`              // e.g., ">=21" for Java
	JVMArgs             []string                    `json:"jvm_args"`                     // JVM arguments for Java plugins
	Wheel               string                      `json:"wheel,omitempty"`              // Python wheel filename
	ExtraWheels         []string                    `json:"extra_wheels,omitempty"`       // Additional wheel files
	PythonMinVersion    string                      `json:"python_min_version,omitempty"` // Minimum Python version
	Port                int                         `json:"port"`
	Status              plugins.PluginStatus        `json:"status"`
	InstallTime         time.Time                   `json:"install_time"`
	LastSeen            time.Time                   `json:"last_seen"`
	SupportedOS         []string                    `json:"supported_os"`
	Commands            []plugins.PluginCommand     `json:"commands"`
	Capabilities        plugins.PluginCapabilities  `json:"capabilities"`
	RequiredPermissions plugins.PluginPermissions   `json:"required_permissions"`
	AIIntegration       plugins.AIIntegrationConfig `json:"ai_integration"`
	PortunixMinVersion  string                      `json:"portunix_min_version,omitempty"`
	OptionalTools       []plugins.OptionalTool      `json:"optional_tools,omitempty"`
	Enabled             bool                        `json:"enabled"`
	Interfaces          []string                    `json:"interfaces,omitempty"` // e.g. ["cli", "grpc"]
	SupportedPlatforms  []plugins.SupportedPlatform `json:"supported_platforms,omitempty"`
}

// NewRegistry creates a new plugin registry
func NewRegistry(filePath string) (*Registry, error) {
	registry := &Registry{
		filePath: filePath,
		data: &RegistryData{
			Version: "1.0.0",
			Plugins: make(map[string]*RegistryPlugin),
		},
	}

	// Load existing registry if it exists
	if err := registry.load(); err != nil {
		// If file doesn't exist, create new registry
		if os.IsNotExist(err) {
			if err := registry.save(); err != nil {
				return nil, fmt.Errorf("failed to create registry file: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load registry: %w", err)
		}
	}

	return registry, nil
}

// RegisterPlugin adds a plugin to the registry
func (r *Registry) RegisterPlugin(manifest *plugins.PluginManifest, installPath string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if plugin already exists
	if _, exists := r.data.Plugins[manifest.Name]; exists {
		return fmt.Errorf("plugin %s already registered", manifest.Name)
	}

	// Convert manifest commands to registry commands
	var commands []plugins.PluginCommand
	for _, cmd := range manifest.Commands {
		commands = append(commands, plugins.PluginCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
			Subcommands: cmd.Subcommands,
			Parameters:  cmd.Parameters,
			Examples:    cmd.Examples,
		})
	}

	// Determine mode and initial status
	// Mode can be explicitly set, or derived from Type
	mode := manifest.Plugin.Mode
	if mode == "" {
		// Derive mode from type: helper/executable types -> helper mode, grpc -> service mode
		if manifest.Plugin.Type == "helper" || manifest.Plugin.Type == "executable" {
			mode = "helper"
		} else {
			mode = "service" // default for grpc plugins
		}
	}
	initialStatus := plugins.PluginStatusStopped
	if mode == "helper" {
		initialStatus = plugins.PluginStatusReady
	}

	// Create registry entry
	registryPlugin := &RegistryPlugin{
		Name:                manifest.Name,
		Version:             manifest.Version,
		Description:         manifest.Description,
		Author:              manifest.Author,
		License:             manifest.License,
		InstallPath:         installPath,
		BinaryName:          manifest.Plugin.Binary,
		Mode:                mode,
		Runtime:             manifest.Plugin.Runtime,
		RuntimeVersion:      manifest.Plugin.RuntimeVersion,
		JVMArgs:             manifest.Plugin.JVMArgs,
		Wheel:               manifest.Plugin.Wheel,
		ExtraWheels:         manifest.Plugin.ExtraWheels,
		PythonMinVersion:    manifest.Plugin.PythonMinVersion,
		Port:                manifest.Plugin.Port,
		Status:              initialStatus,
		InstallTime:         time.Now(),
		LastSeen:            time.Now(),
		SupportedOS:         manifest.Dependencies.OSSupport,
		Commands:            commands,
		RequiredPermissions: manifest.Permissions,
		AIIntegration:       manifest.AIIntegration,
		PortunixMinVersion:  manifest.Dependencies.PortunixMinVersion,
		OptionalTools:       manifest.Dependencies.OptionalTools,
		Enabled:             false,
		Interfaces:          manifest.Plugin.Interfaces,
		SupportedPlatforms:  manifest.SupportedPlatforms,
	}

	// Add to registry
	r.data.Plugins[manifest.Name] = registryPlugin
	r.data.LastUpdate = time.Now()

	return r.save()
}

// ReregisterPlugin updates an existing plugin entry in the registry
func (r *Registry) ReregisterPlugin(manifest *plugins.PluginManifest, installPath string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	existing, exists := r.data.Plugins[manifest.Name]
	if !exists {
		return fmt.Errorf("plugin %s not found in registry", manifest.Name)
	}

	// Preserve enabled state and install time from previous registration
	wasEnabled := existing.Enabled
	installTime := existing.InstallTime

	// Convert manifest commands
	var commands []plugins.PluginCommand
	for _, cmd := range manifest.Commands {
		commands = append(commands, plugins.PluginCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
			Subcommands: cmd.Subcommands,
			Parameters:  cmd.Parameters,
			Examples:    cmd.Examples,
		})
	}

	mode := manifest.Plugin.Mode
	if mode == "" {
		if manifest.Plugin.Type == "helper" || manifest.Plugin.Type == "executable" {
			mode = "helper"
		} else {
			mode = "service"
		}
	}
	initialStatus := plugins.PluginStatusStopped
	if mode == "helper" {
		initialStatus = plugins.PluginStatusReady
	}

	r.data.Plugins[manifest.Name] = &RegistryPlugin{
		Name:                manifest.Name,
		Version:             manifest.Version,
		Description:         manifest.Description,
		Author:              manifest.Author,
		License:             manifest.License,
		InstallPath:         installPath,
		BinaryName:          manifest.Plugin.Binary,
		Mode:                mode,
		Runtime:             manifest.Plugin.Runtime,
		RuntimeVersion:      manifest.Plugin.RuntimeVersion,
		JVMArgs:             manifest.Plugin.JVMArgs,
		Wheel:               manifest.Plugin.Wheel,
		ExtraWheels:         manifest.Plugin.ExtraWheels,
		PythonMinVersion:    manifest.Plugin.PythonMinVersion,
		Port:                manifest.Plugin.Port,
		Status:              initialStatus,
		InstallTime:         installTime,
		LastSeen:            time.Now(),
		SupportedOS:         manifest.Dependencies.OSSupport,
		Commands:            commands,
		RequiredPermissions: manifest.Permissions,
		AIIntegration:       manifest.AIIntegration,
		PortunixMinVersion:  manifest.Dependencies.PortunixMinVersion,
		OptionalTools:       manifest.Dependencies.OptionalTools,
		Enabled:             wasEnabled,
		Interfaces:          manifest.Plugin.Interfaces,
		SupportedPlatforms:  manifest.SupportedPlatforms,
	}
	r.data.LastUpdate = time.Now()

	return r.save()
}

// UnregisterPlugin removes a plugin from the registry
func (r *Registry) UnregisterPlugin(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.data.Plugins[name]; !exists {
		return fmt.Errorf("plugin %s not found in registry", name)
	}

	delete(r.data.Plugins, name)
	r.data.LastUpdate = time.Now()

	return r.save()
}

// UpdatePluginStatus updates the status of a plugin
func (r *Registry) UpdatePluginStatus(name string, status plugins.PluginStatus) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	plugin, exists := r.data.Plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found in registry", name)
	}

	plugin.Status = status
	plugin.LastSeen = time.Now()
	r.data.LastUpdate = time.Now()

	return r.save()
}

// EnablePlugin marks a plugin as enabled
func (r *Registry) EnablePlugin(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	plugin, exists := r.data.Plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found in registry", name)
	}

	plugin.Enabled = true
	plugin.LastSeen = time.Now()
	r.data.LastUpdate = time.Now()

	return r.save()
}

// DisablePlugin marks a plugin as disabled
func (r *Registry) DisablePlugin(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	plugin, exists := r.data.Plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found in registry", name)
	}

	plugin.Enabled = false
	plugin.LastSeen = time.Now()
	r.data.LastUpdate = time.Now()

	return r.save()
}

// GetPlugin returns information about a specific plugin
func (r *Registry) GetPlugin(name string) (plugins.PluginInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	registryPlugin, exists := r.data.Plugins[name]
	if !exists {
		return plugins.PluginInfo{}, fmt.Errorf("plugin %s not found", name)
	}

	return plugins.PluginInfo{
		Name:                registryPlugin.Name,
		Version:             registryPlugin.Version,
		Description:         registryPlugin.Description,
		Author:              registryPlugin.Author,
		License:             registryPlugin.License,
		Mode:                registryPlugin.Mode,
		Interfaces:          registryPlugin.Interfaces,
		SupportedOS:         registryPlugin.SupportedOS,
		Commands:            registryPlugin.Commands,
		Capabilities:        registryPlugin.Capabilities,
		RequiredPermissions: registryPlugin.RequiredPermissions,
		Status:              registryPlugin.Status,
		LastSeen:            registryPlugin.LastSeen,
	}, nil
}

// ListPlugins returns list of all registered plugins
func (r *Registry) ListPlugins() ([]plugins.PluginInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var pluginList []plugins.PluginInfo
	for _, registryPlugin := range r.data.Plugins {
		pluginInfo := plugins.PluginInfo{
			Name:                registryPlugin.Name,
			Version:             registryPlugin.Version,
			Description:         registryPlugin.Description,
			Author:              registryPlugin.Author,
			License:             registryPlugin.License,
			Mode:                registryPlugin.Mode,
			Interfaces:          registryPlugin.Interfaces,
			SupportedOS:         registryPlugin.SupportedOS,
			Commands:            registryPlugin.Commands,
			Capabilities:        registryPlugin.Capabilities,
			RequiredPermissions: registryPlugin.RequiredPermissions,
			Status:              registryPlugin.Status,
			LastSeen:            registryPlugin.LastSeen,
		}
		pluginList = append(pluginList, pluginInfo)
	}

	return pluginList, nil
}

// ListEnabledPlugins returns list of enabled plugins
func (r *Registry) ListEnabledPlugins() ([]plugins.PluginInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var pluginList []plugins.PluginInfo
	for _, registryPlugin := range r.data.Plugins {
		if registryPlugin.Enabled {
			pluginInfo := plugins.PluginInfo{
				Name:                registryPlugin.Name,
				Version:             registryPlugin.Version,
				Description:         registryPlugin.Description,
				Author:              registryPlugin.Author,
				License:             registryPlugin.License,
				SupportedOS:         registryPlugin.SupportedOS,
				Commands:            registryPlugin.Commands,
				Capabilities:        registryPlugin.Capabilities,
				RequiredPermissions: registryPlugin.RequiredPermissions,
				Mode:                registryPlugin.Mode,
				Interfaces:          registryPlugin.Interfaces,
				Status:              registryPlugin.Status,
				LastSeen:            registryPlugin.LastSeen,
			}
			pluginList = append(pluginList, pluginInfo)
		}
	}

	return pluginList, nil
}

// GetPluginCommands returns list of commands for a specific plugin
func (r *Registry) GetPluginCommands(name string) ([]plugins.PluginCommand, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	registryPlugin, exists := r.data.Plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return registryPlugin.Commands, nil
}

// GetAllPluginCommands returns all commands from all enabled plugins
func (r *Registry) GetAllPluginCommands() (map[string][]plugins.PluginCommand, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	commands := make(map[string][]plugins.PluginCommand)
	for name, registryPlugin := range r.data.Plugins {
		if registryPlugin.Enabled {
			commands[name] = registryPlugin.Commands
		}
	}

	return commands, nil
}

// GetPluginMCPTools returns MCP tools for a specific plugin
func (r *Registry) GetPluginMCPTools(name string) ([]plugins.MCPTool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	registryPlugin, exists := r.data.Plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return registryPlugin.AIIntegration.MCPTools, nil
}

// GetAllPluginMCPTools returns all MCP tools from all enabled plugins
func (r *Registry) GetAllPluginMCPTools() (map[string][]plugins.MCPTool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tools := make(map[string][]plugins.MCPTool)
	for name, registryPlugin := range r.data.Plugins {
		if registryPlugin.Enabled {
			tools[name] = registryPlugin.AIIntegration.MCPTools
		}
	}

	return tools, nil
}

// load loads the registry from file
func (r *Registry) load() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(r.data)
}

// save saves the registry to file
func (r *Registry) save() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(r.filePath), 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	file, err := os.Create(r.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(r.data)
}

// GetRegistryStats returns statistics about the registry
func (r *Registry) GetRegistryStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_plugins"] = len(r.data.Plugins)

	var enabled, running, stopped, failed int
	for _, plugin := range r.data.Plugins {
		if plugin.Enabled {
			enabled++
		}
		switch plugin.Status {
		case plugins.PluginStatusRunning:
			running++
		case plugins.PluginStatusStopped:
			stopped++
		case plugins.PluginStatusFailed:
			failed++
		}
	}

	stats["enabled_plugins"] = enabled
	stats["running_plugins"] = running
	stats["stopped_plugins"] = stopped
	stats["failed_plugins"] = failed
	stats["last_update"] = r.data.LastUpdate
	stats["registry_version"] = r.data.Version

	return stats
}

// GetPluginInstallPath returns the install path for a plugin
func (r *Registry) GetPluginInstallPath(name string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	plugin, exists := r.data.Plugins[name]
	if !exists {
		return "", fmt.Errorf("plugin %s not found", name)
	}

	return plugin.InstallPath, nil
}

// GetPluginBinaryName returns the binary name for a plugin
func (r *Registry) GetPluginBinaryName(name string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	plugin, exists := r.data.Plugins[name]
	if !exists {
		return "", fmt.Errorf("plugin %s not found", name)
	}

	return plugin.BinaryName, nil
}

// GetPluginRegistryData returns the full registry data for a plugin
func (r *Registry) GetPluginRegistryData(name string) (*RegistryPlugin, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	plugin, exists := r.data.Plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

// BinaryPath returns the full path to the plugin binary
func (rp *RegistryPlugin) BinaryPath() string {
	if rp.Runtime == "python" && rp.Wheel != "" {
		return venvExecPath(filepath.Join(rp.InstallPath, ".venv"), rp.BinaryName)
	}
	return filepath.Join(rp.InstallPath, rp.BinaryName)
}

// MatchedPlugin is a registry entry that matched a platform query, together
// with the resolved SupportedPlatform record for the requested platform. Both
// CLI and gRPC query paths consume this shape (DEC-5: single source of truth).
type MatchedPlugin struct {
	Plugin          *RegistryPlugin           `json:"plugin"`
	Platform        plugins.SupportedPlatform `json:"platform"`
	MatchedFeatures []string                  `json:"matched_features,omitempty"`
}

// ListPluginsForPlatform returns plugins whose supported_platforms[] declares
// the given platformName. When platformVersion is non-empty, SemVer range
// matching against min_version/max_version is applied (inclusive on both
// ends, omitted bound means unbounded on that side). When requiredFeatures
// is non-empty, only plugins declaring ALL listed features are returned
// (AND-filter). Plugin registry lookup is a linear scan, sufficient for the
// expected plugin count; indexing by platform name can be added if it grows.
func (r *Registry) ListPluginsForPlatform(platformName, platformVersion string, requiredFeatures []string) ([]MatchedPlugin, error) {
	if platformName == "" {
		return nil, fmt.Errorf("platform name is required")
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	out := make([]MatchedPlugin, 0)
	for _, rp := range r.data.Plugins {
		sp, ok := findSupportedPlatform(rp.SupportedPlatforms, platformName)
		if !ok {
			continue
		}
		if platformVersion != "" {
			within, err := plugins.MatchesVersionRange(platformVersion, sp.MinVersion, sp.MaxVersion)
			if err != nil {
				// A malformed version in a stored manifest should not break the
				// whole query — skip this plugin but surface nothing to the
				// caller (install-time validation catches malformed bounds).
				continue
			}
			if !within {
				continue
			}
		}
		if !hasAllFeatures(sp.Features, requiredFeatures) {
			continue
		}
		matched := intersectFeatures(sp.Features, requiredFeatures)
		out = append(out, MatchedPlugin{
			Plugin:          rp,
			Platform:        sp,
			MatchedFeatures: matched,
		})
	}
	return out, nil
}

// findSupportedPlatform locates a SupportedPlatform entry by canonical name.
func findSupportedPlatform(list []plugins.SupportedPlatform, name string) (plugins.SupportedPlatform, bool) {
	for _, sp := range list {
		if sp.Name == name {
			return sp, true
		}
	}
	return plugins.SupportedPlatform{}, false
}

// hasAllFeatures reports whether declared contains every token in required.
// An empty required list matches any declaration.
func hasAllFeatures(declared, required []string) bool {
	if len(required) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(declared))
	for _, f := range declared {
		set[f] = struct{}{}
	}
	for _, f := range required {
		if _, ok := set[f]; !ok {
			return false
		}
	}
	return true
}

// intersectFeatures returns the subset of required that is also in declared,
// preserving the order of required. When required is empty, returns nil so the
// serialized output omits the field cleanly.
func intersectFeatures(declared, required []string) []string {
	if len(required) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(declared))
	for _, f := range declared {
		set[f] = struct{}{}
	}
	out := make([]string, 0, len(required))
	for _, f := range required {
		if _, ok := set[f]; ok {
			out = append(out, f)
		}
	}
	return out
}
