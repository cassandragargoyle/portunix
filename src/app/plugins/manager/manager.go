/*
 *  SPDX-License-Identifier: MIT
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
	"portunix.ai/portunix/src/pkg/platform"
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
	prereqResult := manifestPrerequisites(manifest, filepath.Dir(manifestPath))
	if prereqResult.HasErrors() {
		fmt.Printf("\n%s", prereqResult.FormatHuman())
		fmt.Println("\n\u26a0\ufe0f  Prerequisites check found issues. Plugin will be installed but may not work correctly.")
		fmt.Println("   Run 'portunix plugin check " + manifest.Name + "' for details.\n")
	} else if prereqResult.HasWarnings() {
		fmt.Printf("\n%s\n", prereqResult.FormatHuman())
	}

	// Copy plugin files to plugins directory
	pluginDir := filepath.Join(m.config.PluginsDir, manifest.Name)
	if err := m.copyPluginFiles(manifestPath, pluginDir); err != nil {
		return fmt.Errorf("failed to copy plugin files: %w", err)
	}

	// Setup Python wheel plugin (venv + pip install)
	if manifest.Plugin.Runtime == "python" && manifest.Plugin.Wheel != "" {
		if err := m.setupPythonWheelPlugin(manifest, pluginDir); err != nil {
			_ = os.RemoveAll(pluginDir)
			return fmt.Errorf("failed to setup Python wheel plugin: %w", err)
		}
	}

	// Register plugin (never leave files behind without a registry entry)
	if err := m.registry.RegisterPlugin(manifest, pluginDir); err != nil {
		_ = os.RemoveAll(pluginDir)
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	return nil
}

// ForceInstallPlugin reinstalls a plugin, replacing existing installation.
// The replacement is atomic from the user's point of view: new files are
// staged and checked first, the old installation is kept as a backup until
// the new one is set up and registered, and it is restored on any failure
func (m *Manager) ForceInstallPlugin(manifestPath string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.forceInstallPlugin(manifestPath)
}

// RepairPlugin reinstalls a plugin from its own installed manifest and files,
// recreating the Python venv; the original plugin source is not needed
func (m *Manager) RepairPlugin(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	registryData, err := m.registry.GetPluginRegistryData(name)
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(registryData.InstallPath, "plugin.json")
	if _, err := os.Stat(manifestPath); err != nil {
		return fmt.Errorf("installed manifest not found at %s; reinstall from source with: portunix plugin install --force <plugin-path>", manifestPath)
	}

	return m.forceInstallPlugin(manifestPath)
}

// forceInstallPlugin implements ForceInstallPlugin; caller holds m.mutex
func (m *Manager) forceInstallPlugin(manifestPath string) error {
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
	_, registryErr := m.registry.GetPlugin(manifest.Name)
	existsInRegistry := registryErr == nil
	pluginDir := filepath.Join(m.config.PluginsDir, manifest.Name)
	isWheel := manifest.Plugin.Runtime == "python" && manifest.Plugin.Wheel != ""

	// Check prerequisites
	prereqResult := manifestPrerequisites(manifest, filepath.Dir(manifestPath))
	if prereqResult.HasErrors() {
		fmt.Printf("\n%s", prereqResult.FormatHuman())
		fmt.Println("\n\u26a0\ufe0f  Prerequisites check found issues. Plugin will be installed but may not work correctly.")
		fmt.Println("   Run 'portunix plugin check " + manifest.Name + "' for details.\n")
	} else if prereqResult.HasWarnings() {
		fmt.Printf("\n%s\n", prereqResult.FormatHuman())
	}

	// Stage new files next to the target so the final rename stays on one filesystem
	if err := os.MkdirAll(m.config.PluginsDir, 0755); err != nil { // #nosec G301 -- installed software must be readable by all users
		return fmt.Errorf("failed to create plugins directory: %w", err)
	}
	stagingDir, err := os.MkdirTemp(m.config.PluginsDir, ".staging-"+manifest.Name+"-")
	if err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDir) // no-op once the staging directory was moved into place

	if err := m.copyPluginFiles(manifestPath, stagingDir); err != nil {
		return fmt.Errorf("failed to copy plugin files: %w", err)
	}

	// Pre-check the venv backend before anything touches the old installation
	if isWheel {
		if _, err := planPythonWheelVenv(manifest, stagingDir); err != nil {
			return fmt.Errorf("failed to setup Python wheel plugin: %w (existing installation left unchanged)", err)
		}
	}

	// Stop running plugin if it exists
	if existsInRegistry {
		if p, running := m.plugins[manifest.Name]; running {
			if err := p.Stop(m.ctx); err != nil {
				return fmt.Errorf("failed to stop plugin before reinstall: %w", err)
			}
			delete(m.plugins, manifest.Name)
		}
		m.stopServiceInstances(manifest.Name)
	}

	// Keep the old installation as a backup until the new one is complete
	backupDir := ""
	if _, err := os.Stat(pluginDir); err == nil {
		backupDir = filepath.Join(m.config.PluginsDir, fmt.Sprintf(".backup-%s-%d", manifest.Name, time.Now().UnixNano()))
		if err := os.Rename(pluginDir, backupDir); err != nil {
			return fmt.Errorf("failed to back up existing installation: %w", err)
		}
	}

	restore := func() {
		_ = os.RemoveAll(pluginDir)
		if backupDir != "" {
			if err := os.Rename(backupDir, pluginDir); err != nil {
				fmt.Printf("⚠️  Failed to restore previous installation from %s: %v\n", backupDir, err)
			}
		}
	}

	if err := os.Rename(stagingDir, pluginDir); err != nil {
		restore()
		return fmt.Errorf("failed to move new installation into place: %w", err)
	}

	// The venv is created in its final location: venvs embed absolute paths
	if isWheel {
		if err := m.setupPythonWheelPlugin(manifest, pluginDir); err != nil {
			restore()
			return fmt.Errorf("failed to setup Python wheel plugin: %w (previous installation restored)", err)
		}
	}

	// Register or re-register
	if existsInRegistry {
		err = m.registry.ReregisterPlugin(manifest, pluginDir)
	} else {
		err = m.registry.RegisterPlugin(manifest, pluginDir)
	}
	if err != nil {
		restore()
		return fmt.Errorf("failed to update plugin registry: %w (previous installation restored)", err)
	}

	if backupDir != "" {
		_ = os.RemoveAll(backupDir)
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
		return err
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
		return err
	}

	// Warn (do not block) if current OS is not in supported list
	if len(registryData.SupportedOS) > 0 {
		osCheck := plugins.CheckOSSupport(registryData.SupportedOS)
		if osCheck.Status == plugins.CheckStatusError {
			fmt.Printf("\n⚠️  Plugin %s does not declare support for current OS '%s' (declared: %s).\n"+
				"   Enabling anyway, but startup may fail. Run 'portunix plugin check %s' for details.\n\n",
				name, osCheck.Current, strings.Join(osCheck.Required, ", "), name)
		}
	}

	// Create plugin configuration
	config := plugins.PluginConfig{
		Name:           registryData.Name,
		Version:        registryData.Version,
		BinaryPath:     registryData.BinaryPath(),
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
		_ = m.registry.UpdatePluginStatus(name, plugins.PluginStatusFailed)
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
		_ = m.registry.UpdatePluginStatus(name, plugins.PluginStatusFailed)
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

// ListPlugins returns list of all registered plugins; the status reflects the
// installed files, so a missing directory or broken venv is never "ready"
func (m *Manager) ListPlugins() ([]plugins.PluginInfo, error) {
	pluginList, err := m.registry.ListPlugins()
	if err != nil {
		return nil, err
	}
	for i := range pluginList {
		m.applyInstallStatus(&pluginList[i])
	}
	return pluginList, nil
}

// applyInstallStatus overrides the stored status with the install state
func (m *Manager) applyInstallStatus(info *plugins.PluginInfo) {
	registryData, err := m.registry.GetPluginRegistryData(info.Name)
	if err != nil {
		return
	}
	state, _ := registryData.CheckInstall()
	info.Status = installStatus(info.Status, state)
}

// CheckPluginInstall returns the install state of a registered plugin, the
// reason and the command that repairs it
func (m *Manager) CheckPluginInstall(name string) (InstallState, string, string, error) {
	registryData, err := m.registry.GetPluginRegistryData(name)
	if err != nil {
		return "", "", "", err
	}
	state, reason := registryData.CheckInstall()
	return state, reason, registryData.RepairHint(state), nil
}

// ListPluginsForPlatform returns plugins whose supported_platforms[] declares
// the given platformName and match the optional version range and feature
// AND-filter. See Registry.ListPluginsForPlatform for filter semantics.
func (m *Manager) ListPluginsForPlatform(platformName, platformVersion string, requiredFeatures []string) ([]MatchedPlugin, error) {
	return m.registry.ListPluginsForPlatform(platformName, platformVersion, requiredFeatures)
}

// GetPlugin returns information about a specific plugin
func (m *Manager) GetPlugin(name string) (plugins.PluginInfo, error) {
	info, err := m.registry.GetPlugin(name)
	if err != nil {
		return info, err
	}
	m.applyInstallStatus(&info)
	return info, nil
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
	// Installed files first: a broken venv still leaves the entry script in place
	if state, reason := plugin.CheckInstall(); state != InstallStateOK {
		return plugins.PluginHealth{
			Healthy:       false,
			Status:        "install_" + string(state),
			Message:       fmt.Sprintf("%s. Fix: %s", reason, plugin.RepairHint(state)),
			LastCheckTime: time.Now(),
		}
	}

	binaryPath := plugin.BinaryPath()

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

	// Check binary is executable (Unix only — Windows uses extension-based execution)
	if goruntime.GOOS != "windows" && info.Mode()&0111 == 0 {
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

// pythonVenvPlan describes how the venv of a wheel plugin will be created
type pythonVenvPlan struct {
	wheelPath string
	uv        string // uv binary; empty means the system python3 fallback
	pythonCmd string // system python3 for the fallback
	pyVersion string // exact "major.minor" the wheel needs, empty if any
}

// planPythonWheelVenv runs every check needed to create the venv without
// side effects, so a reinstall can fail before touching the old installation
func planPythonWheelVenv(manifest *plugins.PluginManifest, pluginDir string) (*pythonVenvPlan, error) {
	// Check wheel file exists in plugin directory
	wheelPath := filepath.Join(pluginDir, manifest.Plugin.Wheel)
	if _, err := os.Stat(wheelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("wheel file not found: %s (expected at %s)", manifest.Plugin.Wheel, wheelPath)
	}

	// Exact minor for bytecode wheels, e.g. "3.13"; detected from the wheel
	// itself when the manifest does not pin it
	pyVersion := manifest.Plugin.PythonVersion
	if pyVersion == "" {
		pyVersion = detectWheelPythonVersion(wheelPath)
	}

	// Preferred path: uv provisions a uv-managed interpreter (downloads CPython
	// when needed) that survives OS upgrades of the system Python
	if uv := findUV(); uv != "" {
		return &pythonVenvPlan{wheelPath: wheelPath, uv: uv, pyVersion: pyVersion}, nil
	}

	// Fallback: system python3
	pythonCmd, err := systemPythonForVenv(manifest, pyVersion)
	if err != nil {
		return nil, err
	}
	return &pythonVenvPlan{wheelPath: wheelPath, pythonCmd: pythonCmd, pyVersion: pyVersion}, nil
}

// systemPythonForVenv returns the host python3 if it can create a working
// venv for the plugin, or an error suggesting uv
func systemPythonForVenv(manifest *plugins.PluginManifest, pyVersion string) (string, error) {
	pythonCmd := findPython()
	if pythonCmd == "" {
		return "", fmt.Errorf("Python not found. Install uv (portunix install uv) or Python (portunix install python)")
	}

	// Validate Python version if specified
	if manifest.Plugin.PythonMinVersion != "" {
		constraint := ">=" + manifest.Plugin.PythonMinVersion
		runtimeCheck := plugins.CheckPythonRuntime("python "+constraint, constraint)
		if runtimeCheck.Status == plugins.CheckStatusError {
			return "", fmt.Errorf("%s\n   Fix: %s", runtimeCheck.Message, runtimeCheck.Fix)
		}
	}

	// A bytecode-only wheel needs the exact minor; refuse to build a venv that
	// would only fail later at runtime with "bad magic number"
	if pyVersion != "" {
		hostMinor := pythonMinor(pythonCmd)
		if hostMinor != pyVersion {
			return "", fmt.Errorf(
				"plugin %s needs Python %s but host python is %s; install uv "+
					"(portunix install uv) so the exact interpreter can be provisioned",
				manifest.Name, pyVersion, hostMinor)
		}
	}

	// Check venv module is available
	checkCmd := exec.Command(pythonCmd, "-m", "venv", "--help")
	if err := checkCmd.Run(); err != nil {
		return "", fmt.Errorf("Python venv module not available. Install uv (portunix install uv), or install python3-venv (apt install python3-venv on Debian/Ubuntu, dnf install python3-venv on Fedora)")
	}

	// venv bootstraps pip through ensurepip, which distributions may ship separately
	ensurepipCmd := exec.Command(pythonCmd, "-c", "import ensurepip")
	if err := ensurepipCmd.Run(); err != nil {
		return "", fmt.Errorf("Python ensurepip module not available, so venv cannot install pip. Install uv (portunix install uv), or install python3-venv (apt install python3-venv on Debian/Ubuntu, dnf install python3-venv on Fedora)")
	}

	return pythonCmd, nil
}

// setupPythonWheelPlugin creates venv and installs the wheel package
func (m *Manager) setupPythonWheelPlugin(manifest *plugins.PluginManifest, pluginDir string) error {
	plan, err := planPythonWheelVenv(manifest, pluginDir)
	if err != nil {
		return err
	}

	venvPath := filepath.Join(pluginDir, ".venv")
	if plan.uv != "" {
		pyRequest, defaulted := uvPythonRequest(manifest, plan.pyVersion)
		if defaulted {
			fmt.Printf("  ⚠️  Plugin %s does not set plugin.python_version; using Python %s (override with %s)\n",
				manifest.Name, pyRequest, pluginPythonVersionEnv)
			fmt.Printf("     Recommended: pin the interpreter in plugin.json: \"python_version\": \"%s\"\n", pyRequest)
		}
		return m.setupPythonWheelPluginUV(manifest, pluginDir, plan.wheelPath, venvPath, plan.uv, pyRequest)
	}
	return m.setupPythonWheelPluginSystem(manifest, pluginDir, plan.wheelPath, venvPath, plan.pythonCmd)
}

// uvPythonRequest returns the --python request for uv: the exact minor when
// known, otherwise the default plugin minor (if it meets the manifest
// minimum), otherwise the manifest minimum; defaulted reports that the
// default minor was chosen because the plugin does not pin one
func uvPythonRequest(manifest *plugins.PluginManifest, pyVersion string) (request string, defaulted bool) {
	if pyVersion != "" {
		return pyVersion, false
	}
	defaultVersion := defaultPluginPythonVersion()
	if minVersion := manifest.Plugin.PythonMinVersion; minVersion != "" &&
		!plugins.SatisfiesVersionConstraint(defaultVersion, ">="+minVersion) {
		return ">=" + minVersion, false
	}
	return defaultVersion, true
}

// uvVenvArgs builds the "uv venv" arguments; managed restricts uv to its own
// interpreters so the venv does not bind to the system Python
func uvVenvArgs(venvPath, pyRequest string, managed bool) []string {
	args := []string{"venv"}
	if managed {
		args = append(args, "--managed-python")
	}
	if pyRequest != "" {
		args = append(args, "--python", pyRequest)
	}
	return append(args, venvPath)
}

// setupPythonWheelPluginUV creates the venv with a uv-managed interpreter
func (m *Manager) setupPythonWheelPluginUV(manifest *plugins.PluginManifest, pluginDir, wheelPath, venvPath, uv, pyRequest string) error {
	fmt.Printf("  Creating Python virtual environment with uv (Python %s)...\n", pyRequest)
	cmd := exec.Command(uv, uvVenvArgs(venvPath, pyRequest, true)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Offline or older uv: accept any interpreter uv can find
		fmt.Printf("  uv-managed Python unavailable, using any matching interpreter: %s\n", strings.TrimSpace(string(output)))
		_ = os.RemoveAll(venvPath)
		cmd = exec.Command(uv, uvVenvArgs(venvPath, pyRequest, false)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create virtual environment with uv: %s\n%s", err, string(output))
		}
	}

	install := func(whl string) error {
		cmd := exec.Command(uv, "pip", "install", "--python", venvPath, whl)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("uv pip install of %s failed: %s\n%s", filepath.Base(whl), err, string(output))
		}
		return nil
	}

	// Install extra wheels before the main wheel (dependency resolution)
	if err := installExtraWheels(manifest, pluginDir, install); err != nil {
		return err
	}

	// Install main wheel
	fmt.Printf("  Installing wheel: %s\n", manifest.Plugin.Wheel)
	if err := install(wheelPath); err != nil {
		return err
	}

	fmt.Printf("  Python wheel plugin setup complete\n")
	return nil
}

// setupPythonWheelPluginSystem creates the venv with the host's python3 (fallback)
func (m *Manager) setupPythonWheelPluginSystem(manifest *plugins.PluginManifest, pluginDir, wheelPath, venvPath, pythonCmd string) error {
	// Create virtual environment
	fmt.Printf("  Creating Python virtual environment...\n")
	cmd := exec.Command(pythonCmd, "-m", "venv", venvPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create virtual environment: %s\n%s", err, string(output))
	}

	pipPath := venvExecPath(venvPath, "pip")
	install := func(whl string) error {
		cmd := exec.Command(pipPath, "install", whl)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("pip install of %s failed: %s\n%s", filepath.Base(whl), err, string(output))
		}
		return nil
	}

	// Install extra wheels before the main wheel (dependency resolution)
	if err := installExtraWheels(manifest, pluginDir, install); err != nil {
		return err
	}

	// Install main wheel via pip
	fmt.Printf("  Installing wheel: %s\n", manifest.Plugin.Wheel)
	if err := install(wheelPath); err != nil {
		return err
	}

	fmt.Printf("  Python wheel plugin setup complete\n")
	return nil
}

// installExtraWheels resolves the extra_wheels glob patterns and installs each
// match via the provided installer (uv or pip)
func installExtraWheels(manifest *plugins.PluginManifest, pluginDir string, install func(whl string) error) error {
	if len(manifest.Plugin.ExtraWheels) == 0 {
		return nil
	}
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
			if err := install(whl); err != nil {
				return err
			}
		}
	}
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

// findUV returns the uv binary path or empty string
func findUV() string {
	if _, err := exec.LookPath("uv"); err == nil {
		return "uv"
	}
	return ""
}

// pythonMinor returns the "major.minor" version of the given python binary
// (e.g. "3.13"), or "unknown" if it cannot be determined
func pythonMinor(pythonCmd string) string {
	out, err := exec.Command(pythonCmd, "-c", "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// venvBinDir returns the venv binary directory name (platform-dependent)
func venvBinDir() string {
	if goruntime.GOOS == "windows" {
		return "Scripts"
	}
	return "bin"
}

// venvExecPath returns the absolute path to a venv executable, appending
// the .exe suffix on Windows when missing
func venvExecPath(venvDir, name string) string {
	return filepath.Join(venvDir, venvBinDir(), platform.ExecutableName(name))
}

// copyPluginFiles copies plugin files from source to destination
func (m *Manager) copyPluginFiles(manifestPath, destDir string) error {
	sourceDir := filepath.Dir(manifestPath)

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil { // #nosec G301 -- installed software must be readable by all users
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

		// Venvs embed absolute paths and are recreated, never copied
		if info.IsDir() && info.Name() == ".venv" && path != sourceDir {
			return filepath.SkipDir
		}

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
				_ = m.registry.UpdatePluginStatus(name, plugins.PluginStatusFailed)
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
		return nil, err
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
	applyInstalledRuntimeCheck(result, registryData)
	result.SetInstallCheck(installCheckResult(registryData))

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
		applyInstalledRuntimeCheck(result, registryData)
		result.SetInstallCheck(installCheckResult(registryData))

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

// installCheckResult converts the install state into a prerequisite check entry
func installCheckResult(rp *RegistryPlugin) *plugins.InstallCheckResult {
	state, reason := rp.CheckInstall()
	check := &plugins.InstallCheckResult{
		Status:  plugins.CheckStatusOK,
		State:   string(state),
		Message: reason,
		Fix:     rp.RepairHint(state),
	}
	if state != InstallStateOK {
		check.Status = plugins.CheckStatusError
	}
	return check
}

// applyInstalledRuntimeCheck replaces the host runtime check of an installed
// wheel plugin with a check of its venv interpreter, which is what it runs on
func applyInstalledRuntimeCheck(result *plugins.PrerequisiteCheckResult, rp *RegistryPlugin) {
	if rp.Runtime != "python" || rp.Wheel == "" || rp.InstallPath == "" {
		return
	}
	interpreter := venvExecPath(filepath.Join(rp.InstallPath, ".venv"), "python")
	result.SetRuntimeCheck(plugins.CheckVenvPythonRuntime(
		result.Runtime.Required, rp.RuntimeVersion, interpreter, rp.RepairHint(InstallStateBroken)))
}

// CheckManifestPrerequisites checks prerequisites for a manifest (used during install)
func (m *Manager) CheckManifestPrerequisites(manifest *plugins.PluginManifest) *plugins.PrerequisiteCheckResult {
	return plugins.CheckPrerequisites(manifest)
}

// manifestPrerequisites checks prerequisites before install or repair; for a
// wheel plugin with uv available the runtime is the interpreter uv will
// provide for the venv, not the host python3
func manifestPrerequisites(manifest *plugins.PluginManifest, sourceDir string) *plugins.PrerequisiteCheckResult {
	result := plugins.CheckPrerequisites(manifest)
	if manifest.Plugin.Runtime != "python" || manifest.Plugin.Wheel == "" {
		return result
	}
	uv := findUV()
	if uv == "" {
		return result
	}

	pyVersion := manifest.Plugin.PythonVersion
	if pyVersion == "" {
		pyVersion = detectWheelPythonVersion(filepath.Join(sourceDir, manifest.Plugin.Wheel))
	}
	pyRequest, _ := uvPythonRequest(manifest, pyVersion)
	result.SetRuntimeCheck(plugins.CheckUVPythonRuntime(result.Runtime.Required, uv, pyRequest))
	return result
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
	_, _ = orch.StopPlugin(name)
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
