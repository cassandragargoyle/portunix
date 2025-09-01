package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"portunix.cz/app/plugins"
)

// Manager handles plugin lifecycle and registry
type Manager struct {
	plugins      map[string]plugins.Plugin
	registry     *Registry
	config       ManagerConfig
	healthTicker *time.Ticker
	ctx          context.Context
	cancel       context.CancelFunc
	mutex        sync.RWMutex
}

// ManagerConfig holds configuration for the plugin manager
type ManagerConfig struct {
	PluginsDir          string        `json:"plugins_dir"`
	RegistryFile        string        `json:"registry_file"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	DefaultPort         int           `json:"default_port"`
	PortRange           PortRange     `json:"port_range"`
}

// PortRange defines the range of ports available for plugins
type PortRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// NewManager creates a new plugin manager
func NewManager(config ManagerConfig) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	registry, err := NewRegistry(config.RegistryFile)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create registry: %w", err)
	}

	manager := &Manager{
		plugins:  make(map[string]plugins.Plugin),
		registry: registry,
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start health check ticker
	if config.HealthCheckInterval > 0 {
		manager.healthTicker = time.NewTicker(config.HealthCheckInterval)
		go manager.healthCheckLoop()
	}

	return manager, nil
}

// InstallPlugin installs a plugin from a manifest file
func (m *Manager) InstallPlugin(manifestPath string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	manifest, err := plugins.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Validate plugin
	if err := m.validatePlugin(manifest); err != nil {
		return fmt.Errorf("plugin validation failed: %w", err)
	}

	// Copy plugin files to plugins directory
	pluginDir := filepath.Join(m.config.PluginsDir, manifest.Name)
	if err := m.copyPluginFiles(manifestPath, pluginDir); err != nil {
		return fmt.Errorf("failed to copy plugin files: %w", err)
	}

	// Register plugin
	if err := m.registry.RegisterPlugin(manifest, pluginDir); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	return nil
}

// UninstallPlugin removes a plugin
func (m *Manager) UninstallPlugin(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Stop plugin if running
	if plugin, exists := m.plugins[name]; exists {
		if err := plugin.Stop(m.ctx); err != nil {
			return fmt.Errorf("failed to stop plugin before uninstall: %w", err)
		}
		delete(m.plugins, name)
	}

	// Get install path and remove plugin directory
	installPath, err := m.registry.GetPluginInstallPath(name)
	if err != nil {
		return fmt.Errorf("plugin not found in registry: %w", err)
	}

	if err := os.RemoveAll(installPath); err != nil {
		return fmt.Errorf("failed to remove plugin directory: %w", err)
	}

	// Unregister plugin
	if err := m.registry.UnregisterPlugin(name); err != nil {
		return fmt.Errorf("failed to unregister plugin: %w", err)
	}

	return nil
}

// EnablePlugin enables a plugin and adds it to active plugins
func (m *Manager) EnablePlugin(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if already enabled
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %s is already enabled", name)
	}

	// Get plugin registry data
	registryData, err := m.registry.GetPluginRegistryData(name)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	// Create plugin configuration
	config := plugins.PluginConfig{
		Name:        registryData.Name,
		Version:     registryData.Version,
		BinaryPath:  filepath.Join(registryData.InstallPath, registryData.BinaryName),
		Port:        m.assignPort(),
		WorkingDir:  registryData.InstallPath,
		Environment: make(map[string]string),
		Permissions: registryData.RequiredPermissions,
	}

	// Create plugin instance
	plugin := plugins.NewGRPCPlugin(config)
	if err := plugin.Initialize(m.ctx, config); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// Add to active plugins
	m.plugins[name] = plugin

	// Update registry status
	if err := m.registry.UpdatePluginStatus(name, plugins.PluginStatusStopped); err != nil {
		return fmt.Errorf("failed to update plugin status: %w", err)
	}

	return nil
}

// DisablePlugin disables a plugin and removes it from active plugins
func (m *Manager) DisablePlugin(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s is not enabled", name)
	}

	// Stop plugin if running
	if plugin.IsRunning() {
		if err := plugin.Stop(m.ctx); err != nil {
			return fmt.Errorf("failed to stop plugin: %w", err)
		}
	}

	// Remove from active plugins
	delete(m.plugins, name)

	// Update registry status
	if err := m.registry.UpdatePluginStatus(name, plugins.PluginStatusStopped); err != nil {
		return fmt.Errorf("failed to update plugin status: %w", err)
	}

	return nil
}

// StartPlugin starts a specific plugin
func (m *Manager) StartPlugin(name string) error {
	m.mutex.RLock()
	plugin, exists := m.plugins[name]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s is not enabled", name)
	}

	if plugin.IsRunning() {
		return fmt.Errorf("plugin %s is already running", name)
	}

	// Update status to starting
	if err := m.registry.UpdatePluginStatus(name, plugins.PluginStatusStarting); err != nil {
		return fmt.Errorf("failed to update plugin status: %w", err)
	}

	// Start plugin
	if err := plugin.Start(m.ctx); err != nil {
		m.registry.UpdatePluginStatus(name, plugins.PluginStatusFailed)
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	// Update status to running
	if err := m.registry.UpdatePluginStatus(name, plugins.PluginStatusRunning); err != nil {
		return fmt.Errorf("failed to update plugin status: %w", err)
	}

	return nil
}

// StopPlugin stops a specific plugin
func (m *Manager) StopPlugin(name string) error {
	m.mutex.RLock()
	plugin, exists := m.plugins[name]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s is not enabled", name)
	}

	if !plugin.IsRunning() {
		return fmt.Errorf("plugin %s is not running", name)
	}

	// Update status to stopping
	if err := m.registry.UpdatePluginStatus(name, plugins.PluginStatusStopping); err != nil {
		return fmt.Errorf("failed to update plugin status: %w", err)
	}

	// Stop plugin
	if err := plugin.Stop(m.ctx); err != nil {
		m.registry.UpdatePluginStatus(name, plugins.PluginStatusFailed)
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	// Update status to stopped
	if err := m.registry.UpdatePluginStatus(name, plugins.PluginStatusStopped); err != nil {
		return fmt.Errorf("failed to update plugin status: %w", err)
	}

	return nil
}

// ExecuteCommand executes a command on a specific plugin
func (m *Manager) ExecuteCommand(pluginName, command string, args []string, options map[string]string) (*plugins.ExecuteResponse, error) {
	m.mutex.RLock()
	plugin, exists := m.plugins[pluginName]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %s is not enabled", pluginName)
	}

	if !plugin.IsRunning() {
		return nil, fmt.Errorf("plugin %s is not running", pluginName)
	}

	request := plugins.ExecuteRequest{
		Command: command,
		Args:    args,
		Options: options,
	}

	return plugin.Execute(m.ctx, request)
}

// ListPlugins returns list of all registered plugins
func (m *Manager) ListPlugins() ([]plugins.PluginInfo, error) {
	return m.registry.ListPlugins()
}

// GetPlugin returns information about a specific plugin
func (m *Manager) GetPlugin(name string) (plugins.PluginInfo, error) {
	return m.registry.GetPlugin(name)
}

// GetPluginHealth returns health status of a plugin
func (m *Manager) GetPluginHealth(name string) (plugins.PluginHealth, error) {
	m.mutex.RLock()
	plugin, exists := m.plugins[name]
	m.mutex.RUnlock()

	if !exists {
		return plugins.PluginHealth{}, fmt.Errorf("plugin %s is not enabled", name)
	}

	return plugin.Health(m.ctx), nil
}

// Shutdown stops all plugins and shuts down the manager
func (m *Manager) Shutdown() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Stop health check ticker
	if m.healthTicker != nil {
		m.healthTicker.Stop()
	}

	// Stop all plugins
	for name, plugin := range m.plugins {
		if plugin.IsRunning() {
			if err := plugin.Stop(m.ctx); err != nil {
				fmt.Printf("Error stopping plugin %s: %v\n", name, err)
			}
		}
	}

	// Cancel context
	m.cancel()

	return nil
}

// validatePlugin validates a plugin manifest
func (m *Manager) validatePlugin(manifest *plugins.PluginManifest) error {
	// Check required fields
	if manifest.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if manifest.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	if manifest.Plugin.Binary == "" {
		return fmt.Errorf("plugin binary is required")
	}

	// Check if plugin already exists
	if _, err := m.registry.GetPlugin(manifest.Name); err == nil {
		return fmt.Errorf("plugin %s already exists", manifest.Name)
	}

	return nil
}

// copyPluginFiles copies plugin files from source to destination
func (m *Manager) copyPluginFiles(manifestPath, destDir string) error {
	sourceDir := filepath.Dir(manifestPath)

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Copy all files from source to destination
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file
		return copyFile(path, destPath)
	})
}

// assignPort assigns an available port to a plugin
func (m *Manager) assignPort() int {
	// Start from default port and find next available
	port := m.config.DefaultPort
	if port == 0 {
		port = m.config.PortRange.Start
	}

	usedPorts := make(map[int]bool)
	for _, plugin := range m.plugins {
		if grpcPlugin, ok := plugin.(*plugins.GRPCPlugin); ok {
			// We'll need to add a method to get port from plugin
			// For now, skip this optimization
			_ = grpcPlugin
		}
	}

	for port <= m.config.PortRange.End {
		if !usedPorts[port] {
			return port
		}
		port++
	}

	// Return default if no port available
	return m.config.DefaultPort
}

// healthCheckLoop runs periodic health checks on all plugins
func (m *Manager) healthCheckLoop() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.healthTicker.C:
			m.performHealthChecks()
		}
	}
}

// performHealthChecks checks health of all running plugins
func (m *Manager) performHealthChecks() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for name, plugin := range m.plugins {
		if plugin.IsRunning() {
			health := plugin.Health(m.ctx)

			// Update registry with health status
			if !health.Healthy {
				m.registry.UpdatePluginStatus(name, plugins.PluginStatusFailed)
			}
		}
	}
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}
