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
	Port                int                         `json:"port"`
	Status              plugins.PluginStatus        `json:"status"`
	InstallTime         time.Time                   `json:"install_time"`
	LastSeen            time.Time                   `json:"last_seen"`
	SupportedOS         []string                    `json:"supported_os"`
	Commands            []plugins.PluginCommand     `json:"commands"`
	Capabilities        plugins.PluginCapabilities  `json:"capabilities"`
	RequiredPermissions plugins.PluginPermissions   `json:"required_permissions"`
	AIIntegration       plugins.AIIntegrationConfig `json:"ai_integration"`
	Enabled             bool                        `json:"enabled"`
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

	// Create registry entry
	registryPlugin := &RegistryPlugin{
		Name:                manifest.Name,
		Version:             manifest.Version,
		Description:         manifest.Description,
		Author:              manifest.Author,
		License:             manifest.License,
		InstallPath:         installPath,
		BinaryName:          manifest.Plugin.Binary,
		Port:                manifest.Plugin.Port,
		Status:              plugins.PluginStatusStopped,
		InstallTime:         time.Now(),
		LastSeen:            time.Now(),
		SupportedOS:         manifest.Dependencies.OSSupport,
		Commands:            commands,
		RequiredPermissions: manifest.Permissions,
		AIIntegration:       manifest.AIIntegration,
		Enabled:             false,
	}

	// Add to registry
	r.data.Plugins[manifest.Name] = registryPlugin
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
