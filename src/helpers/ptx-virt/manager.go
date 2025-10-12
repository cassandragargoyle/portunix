package main

import (
	"fmt"
	"runtime"

	"portunix.ai/app/system"
	"portunix.ai/app/virt/types"
	"portunix.ai/app/virt/virtualbox"
)

// VirtManager manages virtualization operations using different providers
type VirtManager struct {
	provider VirtualizationProvider
	config   *VirtConfig
}

// VirtConfig contains configuration for the virtualization manager
type VirtConfig struct {
	PreferredProvider string `json:"preferred_provider,omitempty"`
	AutoDetect        bool   `json:"auto_detect"`
	AllowFallback     bool   `json:"allow_fallback"`
}

// NewVirtManager creates a new virtualization manager
func NewVirtManager() (*VirtManager, error) {
	config := &VirtConfig{
		AutoDetect:    true,
		AllowFallback: true,
	}

	manager := &VirtManager{
		config: config,
	}

	// Auto-detect and select provider
	provider, err := manager.selectProvider()
	if err != nil {
		return nil, fmt.Errorf("no virtualization provider available: %w", err)
	}

	manager.provider = provider
	return manager, nil
}

// selectProvider selects the best available virtualization provider
func (m *VirtManager) selectProvider() (VirtualizationProvider, error) {
	// Get system info for platform detection
	sysInfo, err := system.GetSystemInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get system info: %w", err)
	}

	var platform string
	if system.CheckCondition(sysInfo, "windows") {
		platform = "windows"
	} else if system.CheckCondition(sysInfo, "linux") {
		platform = "linux"
	} else if system.CheckCondition(sysInfo, "darwin") {
		platform = "darwin"
	} else {
		platform = runtime.GOOS
	}

	// Get provider priority for this platform
	priorities := GetProviderPriority(platform)

	// Try providers in priority order
	for _, providerName := range priorities {
		provider := m.createProvider(providerName)
		if provider != nil && provider.IsAvailable() {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("no virtualization provider available on %s", platform)
}

// createProvider creates a provider instance by name
func (m *VirtManager) createProvider(name string) VirtualizationProvider {
	switch name {
	case "virtualbox":
		return &VirtualBoxAdapter{
			backend: virtualbox.NewBackend(),
		}
	case "qemu", "kvm":
		return &QEMUAdapter{}
	case "vmware":
		// TODO: Implement VMware provider adapter
		return nil
	case "hyperv":
		// TODO: Implement Hyper-V provider adapter
		return nil
	default:
		return nil
	}
}

// Provider Management Methods

// GetProvider returns the current active provider
func (m *VirtManager) GetProvider() VirtualizationProvider {
	return m.provider
}

// GetProviderName returns the name of the current provider
func (m *VirtManager) GetProviderName() string {
	if m.provider == nil {
		return "none"
	}
	return m.provider.GetName()
}

// GetProviderVersion returns the version of the current provider
func (m *VirtManager) GetProviderVersion() string {
	if m.provider == nil {
		return ""
	}
	version, err := m.provider.GetVersion()
	if err != nil {
		return ""
	}
	return version
}

// VM Lifecycle Methods (delegate to provider)

// Create creates a new virtual machine
func (m *VirtManager) Create(config *types.VMConfig) error {
	if m.provider == nil {
		return fmt.Errorf("no virtualization provider available")
	}
	return m.provider.Create(config)
}

// Start starts a virtual machine
func (m *VirtManager) Start(vmName string) error {
	if m.provider == nil {
		return fmt.Errorf("no virtualization provider available")
	}
	return m.provider.Start(vmName)
}

// Stop stops a virtual machine
func (m *VirtManager) Stop(vmName string, force bool) error {
	if m.provider == nil {
		return fmt.Errorf("no virtualization provider available")
	}
	return m.provider.Stop(vmName, force)
}

// Restart restarts a virtual machine
func (m *VirtManager) Restart(vmName string) error {
	if m.provider == nil {
		return fmt.Errorf("no virtualization provider available")
	}
	return m.provider.Restart(vmName)
}

// Delete deletes a virtual machine
func (m *VirtManager) Delete(vmName string, keepDisk bool) error {
	if m.provider == nil {
		return fmt.Errorf("no virtualization provider available")
	}
	return m.provider.Delete(vmName, keepDisk)
}

// List lists all virtual machines
func (m *VirtManager) List() ([]*types.VMInfo, error) {
	if m.provider == nil {
		return nil, fmt.Errorf("no virtualization provider available")
	}
	return m.provider.List()
}

// GetInfo gets information about a virtual machine
func (m *VirtManager) GetInfo(vmName string) (*types.VMInfo, error) {
	if m.provider == nil {
		return nil, fmt.Errorf("no virtualization provider available")
	}
	return m.provider.GetInfo(vmName)
}

// GetState gets the state of a virtual machine
func (m *VirtManager) GetState(vmName string) types.VMState {
	if m.provider == nil {
		return types.VMStateError
	}
	return m.provider.GetState(vmName)
}

// Connect connects to a virtual machine via SSH
func (m *VirtManager) Connect(vmName string, opts SSHOptions) error {
	if m.provider == nil {
		return fmt.Errorf("no virtualization provider available")
	}
	return m.provider.Connect(vmName, opts)
}

// Check Methods

// CheckCapabilities checks virtualization capabilities
func (m *VirtManager) CheckCapabilities() (*VirtualizationCapabilities, error) {
	sysInfo, err := system.GetSystemInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get system info: %w", err)
	}

	var platform string
	if system.CheckCondition(sysInfo, "windows") {
		platform = "windows"
	} else if system.CheckCondition(sysInfo, "linux") {
		platform = "linux"
	} else if system.CheckCondition(sysInfo, "darwin") {
		platform = "darwin"
	} else {
		platform = runtime.GOOS
	}

	capabilities := &VirtualizationCapabilities{
		Platform:               platform,
		HardwareVirtualization: sysInfo.Capabilities.VirtualizationInfo.HardwareVirtualization,
		AvailableProviders:     []*ProviderStatus{},
	}

	// Check all possible providers
	allProviders := []string{"virtualbox", "qemu", "kvm", "vmware", "hyperv"}
	for _, providerName := range allProviders {
		provider := m.createProvider(providerName)
		if provider == nil {
			continue // Provider not implemented yet
		}

		status := &ProviderStatus{
			Name:                   providerName,
			Available:              provider.IsAvailable(),
			HardwareVirtualization: capabilities.HardwareVirtualization,
		}

		if status.Available {
			if version, err := provider.GetVersion(); err == nil {
				status.Version = version
			}

			// Get diagnostic info for additional details
			if diag := provider.GetDiagnosticInfo(); diag != nil {
				if diag.PathEnvironment != "" {
					status.InstallationPath = diag.PathEnvironment
				}
				if len(diag.Details) > 0 {
					status.Features = diag.Details
				}
				if len(diag.Suggestions) > 0 {
					status.Recommendations = diag.Suggestions
				}
			}
		}

		capabilities.AvailableProviders = append(capabilities.AvailableProviders, status)
	}

	// Set recommended provider
	priorities := GetProviderPriority(platform)
	for _, providerName := range priorities {
		for _, status := range capabilities.AvailableProviders {
			if status.Name == providerName && status.Available {
				capabilities.RecommendedProvider = providerName
				break
			}
		}
		if capabilities.RecommendedProvider != "" {
			break
		}
	}

	return capabilities, nil
}