/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
 
package manager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"portunix.ai/app/plugins"
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

	// Check prerequisites (warn but don't block installation)
	prereqResult := plugins.CheckPrerequisites(manifest)
	if prereqResult.HasErrors() {
		fmt.Printf("\n%s", prereqResult.FormatHuman())
		fmt.Println("\n\u26a0\ufe0f  Prerequisites check found issues. Plugin will be installed but may not work correctly.")
		fmt.Println("   Run 'portunix plugin check " + manifest.Name + "' for details.\n")
	} else if prereqResult.HasWarnings() {
		fmt.Printf("\n%s", prereqResult.FormatHuman())
	}

	// Copy plugin files to plugins directory
	pluginDir := filepath.Join(m.config.PluginsDir, manifest.Name)
	if err := m.copyPluginFiles(manifestPath, pluginDir); err != nil {
		return fmt.Errorf("failed to copy plugin files: %w", err)
	}

	// Setup Python wheel plugin (venv + pip install)
	if manifest.Plugin.Runtime == "python" && manifest.Plugin.Wheel != "" {
		if err := m.setupPythonWheelPlugin(manifest, pluginDir); err != nil {
			os.RemoveAll(pluginDir)
			return fmt.Errorf("failed to setup Python wheel plugin: %w", err)
		}
	}

	// Register plugin
	if err := m.registry.RegisterPlugin(manifest, pluginDir); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	return nil
}

// ForceInstallPlugin reinstalls a plugin, replacing existing installation
func (m *Manager) ForceInstallPlugin(manifestPath string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	manifest, err := plugins.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Basic validation (skip "already exists" check)
	if manifest.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if manifest.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	if manifest.Plugin.Binary == "" {
		return fmt.Errorf("plugin binary is required")
	}

	// Check if plugin exists (to decide register vs reregister)
	_, existsInRegistry := m.registry.GetPlugin(manifest.Name)
	pluginDir := filepath.Join(m.config.PluginsDir, manifest.Name)

	// Stop running plugin if it exists
	if existsInRegistry == nil {
		if p, running := m.plugins[manifest.Name]; running {
			if err := p.Stop(m.ctx); err != nil {
				return fmt.Errorf("failed to stop plugin before reinstall: %w", err)
			}
			delete(m.plugins, manifest.Name)
		}
		m.stopServiceInstances(manifest.Name)

		// Remove old venv for Python wheel plugins (will be recreated)
		venvPath := filepath.Join(pluginDir, ".venv")
		os.RemoveAll(venvPath)
	}

	// Check prerequisites
	prereqResult := plugins.CheckPrerequisites(manifest)
	if prereqResult.HasErrors() {
		fmt.Printf("\n%s", prereqResult.FormatHuman())
		fmt.Println("\n\u26a0\ufe0f  Prerequisites check found issues. Plugin will be installed but may not work correctly.")
		fmt.Println("   Run 'portunix plugin check " + manifest.Name + "' for details.\n")
	} else if prereqResult.HasWarnings() {
		fmt.Printf("\n%s", prereqResult.FormatHuman())
	}

	// Copy plugin files (overwrites existing)
	if err := m.copyPluginFiles(manifestPath, pluginDir); err != nil {
		return fmt.Errorf("failed to copy plugin files: %w", err)
	}

	// Setup Python wheel plugin
	if manifest.Plugin.Runtime == "python" && manifest.Plugin.Wheel != "" {
		if err := m.setupPythonWheelPlugin(manifest, pluginDir); err != nil {
			os.RemoveAll(pluginDir)
			return fmt.Errorf("failed to setup Python wheel plugin: %w", err)
		}
	}

	// Register or re-register
	if existsInRegistry == nil {
		if err := m.registry.ReregisterPlugin(manifest, pluginDir); err != nil {
			return fmt.Errorf("failed to update plugin registry: %w", err)
		}
	} else {
		if err := m.registry.RegisterPlugin(manifest, pluginDir); err != nil {
			return fmt.Errorf("failed to register plugin: %w", err)
		}
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

	// Stop any service orchestrator instances
	m.stopServiceInstances(name)

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
		Name:           registryData.Name,
		Version:        registryData.Version,
		BinaryPath:     filepath.Join(registryData.InstallPath, registryData.BinaryName),
		Runtime:        registryData.Runtime,
		RuntimeVersion: registryData.RuntimeVersion,
		JVMArgs:        registryData.JVMArgs,
		Port:           m.assignPort(),
		WorkingDir:     registryData.InstallPath,
		Environment:    make(map[string]string),
		Permissions:    registryData.RequiredPermissions,
	}

	// Create plugin instance
	plugin := plugins.NewGRPCPlugin(config)
	if err := plugin.Initialize(m.ctx, config); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// Add to active plugins
	m.plugins[name] = plugin

	// Mark plugin as enabled in registry
	if err := m.registry.EnablePlugin(name); err != nil {
		return fmt.Errorf("failed to enable plugin in registry: %w", err)
	}

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

	// Stop any service orchestrator instances
	m.stopServiceInstances(name)

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

	// Check runtime and OS prerequisites before starting (blocking)
	prereqResult, err := m.CheckPluginPrerequisites(name)
	if err == nil && prereqResult != nil {
		if prereqResult.Runtime != nil && prereqResult.Runtime.Status == plugins.CheckStatusError {
			return fmt.Errorf("cannot start plugin %s: %s\n   Fix: %s", name, prereqResult.Runtime.Message, prereqResult.Runtime.Fix)
		}
		if prereqResult.OSSupport != nil && prereqResult.OSSupport.Status == plugins.CheckStatusError {
			return fmt.Errorf("cannot start plugin %s: %s", name, prereqResult.OSSupport.Message)
		}
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

// GetPluginRegistryData returns the full registry data for a plugin
func (m *Manager) GetPluginRegistryData(name string) (*RegistryPlugin, error) {
	return m.registry.GetPluginRegistryData(name)
}

// GetPluginHealth returns health status of a plugin
func (m *Manager) GetPluginHealth(name string) (plugins.PluginHealth, error) {
	// Check active (gRPC/service) plugins first
	m.mutex.RLock()
	plugin, exists := m.plugins[name]
	m.mutex.RUnlock()

	if exists {
		return plugin.Health(m.ctx), nil
	}

	// Plugin not in active plugins map — check registry for helper plugins
	registryData, err := m.registry.GetPluginRegistryData(name)
	if err != nil {
		return plugins.PluginHealth{}, fmt.Errorf("plugin %s not found", name)
	}

	if registryData.Mode == "helper" {
		return m.checkHelperPluginHealth(registryData), nil
	}

	// Service plugin that is not enabled
	return plugins.PluginHealth{}, fmt.Errorf("plugin %s is not enabled. Enable it with: portunix plugin enable %s", name, name)
}

// checkHelperPluginHealth performs health check for helper-type plugins
func (m *Manager) checkHelperPluginHealth(plugin *RegistryPlugin) plugins.PluginHealth {
	// For Python wheel plugins, binary is in .venv/bin/
	var binaryPath string
	if plugin.Runtime == "python" && plugin.Wheel != "" {
		binaryPath = filepath.Join(plugin.InstallPath, ".venv", venvBinDir(), plugin.BinaryName)
	} else {
		binaryPath = filepath.Join(plugin.InstallPath, plugin.BinaryName)
	}

	// Check binary exists
	info, err := os.Stat(binaryPath)
	if err != nil {
		return plugins.PluginHealth{
			Healthy:       false,
			Status:        "binary_missing",
			Message:       fmt.Sprintf("Binary not found: %s", binaryPath),
			LastCheckTime: time.Now(),
		}
	}

	// Check binary is executable (on Unix)
	if info.Mode()&0111 == 0 {
		return plugins.PluginHealth{
			Healthy:       false,
			Status:        "not_executable",
			Message:       fmt.Sprintf("Binary is not executable: %s", binaryPath),
			LastCheckTime: time.Now(),
		}
	}

	// Try to get version output
	var cmd *exec.Cmd
	switch plugin.Runtime {
	case "java":
		cmd = exec.Command("java", "-jar", binaryPath, "--version")
	case "python":
		if plugin.Wheel != "" {
			// Wheel plugin: entry point is a standalone script in .venv/bin/
			cmd = exec.Command(binaryPath, "--version")
		} else {
			cmd = exec.Command("python3", binaryPath, "--version")
		}
	default:
		cmd = exec.Command(binaryPath, "--version")
	}

	output, err := cmd.CombinedOutput()
	versionInfo := strings.TrimSpace(string(output))

	if err != nil {
		// Binary exists and is executable but --version failed
		// This is still considered healthy for helper plugins (not all support --version)
		return plugins.PluginHealth{
			Healthy:       true,
			Status:        "ready",
			Message:       fmt.Sprintf("Plugin %s v%s — healthy (helper mode, --version not supported)", plugin.Name, plugin.Version),
			LastCheckTime: time.Now(),
		}
	}

	return plugins.PluginHealth{
		Healthy:       true,
		Status:        "ready",
		Message:       fmt.Sprintf("Plugin %s v%s — healthy (helper mode)", plugin.Name, plugin.Version),
		Metrics:       map[string]string{"version_output": versionInfo},
		LastCheckTime: time.Now(),
	}
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

// setupPythonWheelPlugin creates venv and installs the wheel package
func (m *Manager) setupPythonWheelPlugin(manifest *plugins.PluginManifest, pluginDir string) error {
	// Validate Python version if specified
	if manifest.Plugin.PythonMinVersion != "" {
		constraint := ">=" + manifest.Plugin.PythonMinVersion
		runtimeCheck := plugins.CheckPythonRuntime("python "+constraint, constraint)
		if runtimeCheck.Status == plugins.CheckStatusError {
			return fmt.Errorf("%s\n   Fix: %s", runtimeCheck.Message, runtimeCheck.Fix)
		}
	}

	// Check wheel file exists in plugin directory
	wheelPath := filepath.Join(pluginDir, manifest.Plugin.Wheel)
	if _, err := os.Stat(wheelPath); os.IsNotExist(err) {
		return fmt.Errorf("wheel file not found: %s (expected at %s)", manifest.Plugin.Wheel, wheelPath)
	}

	// Find python3 binary
	pythonCmd := findPython()
	if pythonCmd == "" {
		return fmt.Errorf("Python not found. Install with: portunix install python")
	}

	// Check venv module is available
	checkCmd := exec.Command(pythonCmd, "-m", "venv", "--help")
	if err := checkCmd.Run(); err != nil {
		return fmt.Errorf("Python venv module not available. Install it with: apt install python3-venv (Debian/Ubuntu) or dnf install python3-venv (Fedora)")
	}

	// Create virtual environment
	venvPath := filepath.Join(pluginDir, ".venv")
	fmt.Printf("  Creating Python virtual environment...\n")
	cmd := exec.Command(pythonCmd, "-m", "venv", venvPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create virtual environment: %s\n%s", err, string(output))
	}

	// Install extra wheels before the main wheel (dependency resolution)
	pipPath := filepath.Join(venvPath, venvBinDir(), "pip")
	if len(manifest.Plugin.ExtraWheels) > 0 {
		fmt.Printf("  Installing extra wheels...\n")
		for _, pattern := range manifest.Plugin.ExtraWheels {
			matches, err := filepath.Glob(filepath.Join(pluginDir, pattern))
			if err != nil {
				return fmt.Errorf("invalid extra_wheels glob pattern %q: %w", pattern, err)
			}
			if len(matches) == 0 {
				return fmt.Errorf("extra_wheels pattern %q matched no files in %s", pattern, pluginDir)
			}
			for _, whl := range matches {
				fmt.Printf("    Installing: %s\n", filepath.Base(whl))
				cmd = exec.Command(pipPath, "install", whl)
				output, err := cmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("pip install of %s failed: %s\n%s", filepath.Base(whl), err, string(output))
				}
			}
		}
	}

	// Install main wheel via pip
	fmt.Printf("  Installing wheel: %s\n", manifest.Plugin.Wheel)
	cmd = exec.Command(pipPath, "install", wheelPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pip install failed: %s\n%s", err, string(output))
	}

	fmt.Printf("  Python wheel plugin setup complete\n")
	return nil
}

// findPython returns the python3/python binary path or empty string
func findPython() string {
	candidates := []string{"python3", "python"}
	for _, cmd := range candidates {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
	}
	return ""
}

// venvBinDir returns the venv binary directory name (platform-dependent)
func venvBinDir() string {
	if goruntime.GOOS == "windows" {
		return "Scripts"
	}
	return "bin"
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

		// Copy file preserving permissions
		return copyFile(path, destPath, info.Mode())
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

// CheckPluginPrerequisites checks prerequisites for a single plugin
func (m *Manager) CheckPluginPrerequisites(name string) (*plugins.PrerequisiteCheckResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	registryData, err := m.registry.GetPluginRegistryData(name)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %w", err)
	}

	result := plugins.CheckPrerequisitesFromRegistry(
		registryData.Name,
		registryData.Version,
		registryData.Runtime,
		registryData.RuntimeVersion,
		registryData.PortunixMinVersion,
		registryData.SupportedOS,
		registryData.OptionalTools,
	)

	return result, nil
}

// CheckAllPrerequisites checks prerequisites for all installed plugins
func (m *Manager) CheckAllPrerequisites() (*plugins.PrerequisitesSummary, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	allPlugins, err := m.registry.ListPlugins()
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	summary := &plugins.PrerequisitesSummary{}
	summary.Summary.Total = len(allPlugins)

	for _, pluginInfo := range allPlugins {
		registryData, err := m.registry.GetPluginRegistryData(pluginInfo.Name)
		if err != nil {
			continue
		}

		result := plugins.CheckPrerequisitesFromRegistry(
			registryData.Name,
			registryData.Version,
			registryData.Runtime,
			registryData.RuntimeVersion,
			registryData.PortunixMinVersion,
			registryData.SupportedOS,
			registryData.OptionalTools,
		)

		summary.Results = append(summary.Results, *result)

		switch result.Overall {
		case plugins.CheckStatusOK:
			summary.Summary.OK++
		case plugins.CheckStatusWarning:
			summary.Summary.Warning++
		case plugins.CheckStatusError:
			summary.Summary.Error++
		}
	}

	return summary, nil
}

// CheckManifestPrerequisites checks prerequisites for a manifest (used during install)
func (m *Manager) CheckManifestPrerequisites(manifest *plugins.PluginManifest) *plugins.PrerequisiteCheckResult {
	return plugins.CheckPrerequisites(manifest)
}

// stopServiceInstances stops all service orchestrator instances for a plugin
func (m *Manager) stopServiceInstances(name string) {
	if serviceOrchestratorFactory == nil {
		return
	}
	orch, err := serviceOrchestratorFactory()
	if err != nil {
		return
	}
	orch.StopPlugin(name)
}

// serviceOrchestratorFactory creates a service orchestrator (set via RegisterServiceOrchestrator)
var serviceOrchestratorFactory func() (ServiceOrchestrator, error)

// ServiceOrchestrator interface for service lifecycle operations
type ServiceOrchestrator interface {
	StopPlugin(name string) (int, error)
}

// RegisterServiceOrchestrator sets the factory function for creating service orchestrators
func RegisterServiceOrchestrator(factory func() (ServiceOrchestrator, error)) {
	serviceOrchestratorFactory = factory
}

// copyFile copies a single file from src to dst preserving file mode (permissions)
func copyFile(src, dst string, mode os.FileMode) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}
