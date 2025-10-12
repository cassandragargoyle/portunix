package main

import (
	"time"

	"portunix.ai/app/virt/types"
)

// VirtualizationProvider defines the interface that all virtualization backends must implement
type VirtualizationProvider interface {
	// Provider Information
	GetName() string
	GetVersion() (string, error)
	IsAvailable() bool
	GetDiagnosticInfo() *DiagnosticInfo

	// VM Lifecycle Management
	Create(config *types.VMConfig) error
	Start(vmName string) error
	Stop(vmName string, force bool) error
	Restart(vmName string) error
	Suspend(vmName string) error
	Resume(vmName string) error
	Delete(vmName string, keepDisk bool) error

	// VM Information
	List() ([]*types.VMInfo, error)
	GetInfo(vmName string) (*types.VMInfo, error)
	GetState(vmName string) types.VMState
	GetIP(vmName string) (string, error)

	// VM Connection
	IsSSHReady(vmName string) bool
	Connect(vmName string, opts SSHOptions) error

	// Snapshot Management
	CreateSnapshot(vmName, snapshotName, description string) error
	ListSnapshots(vmName string) ([]*SnapshotInfo, error)
	RevertSnapshot(vmName, snapshotName string) error
	DeleteSnapshot(vmName, snapshotName string) error

	// File Operations
	CopyToVM(vmName, localPath, remotePath string) error
	CopyFromVM(vmName, remotePath, localPath string) error
}

// DiagnosticInfo contains diagnostic information for troubleshooting
type DiagnosticInfo struct {
	Platform          string            `json:"platform"`
	PathEnvironment   string            `json:"path_environment,omitempty"`
	RegistryKeys      map[string]string `json:"registry_keys,omitempty"`
	InstallationPaths []string          `json:"installation_paths,omitempty"`
	RunningServices   []string          `json:"running_services,omitempty"`
	PackageManager    string            `json:"package_manager,omitempty"`
	BrewInstalled     bool              `json:"brew_installed,omitempty"`
	Details           []string          `json:"details,omitempty"`
	Suggestions       []string          `json:"suggestions"`
}

// SSHOptions contains SSH connection options
type SSHOptions struct {
	User        string `json:"user,omitempty"`
	Port        int    `json:"port,omitempty"`
	PrivateKey  string `json:"private_key,omitempty"`
	Command     string `json:"command,omitempty"`
	Interactive bool   `json:"interactive"`
}

// SnapshotInfo contains snapshot information
type SnapshotInfo struct {
	Name        string    `json:"name"`
	VM          string    `json:"vm"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Size        int64     `json:"size,omitempty"`
	Parent      string    `json:"parent,omitempty"`
}

// ProviderStatus represents the status of a virtualization provider
type ProviderStatus struct {
	Name                   string   `json:"name"`
	Available              bool     `json:"available"`
	Version                string   `json:"version,omitempty"`
	InstallationPath       string   `json:"installation_path,omitempty"`
	HardwareVirtualization bool     `json:"hardware_virtualization"`
	Features               []string `json:"features,omitempty"`
	Issues                 []string `json:"issues,omitempty"`
	Recommendations        []string `json:"recommendations,omitempty"`
}

// VirtualizationCapabilities represents system virtualization capabilities
type VirtualizationCapabilities struct {
	Platform               string            `json:"platform"`
	HardwareVirtualization bool              `json:"hardware_virtualization"`
	NestedVirtualization   bool              `json:"nested_virtualization"`
	AvailableProviders     []*ProviderStatus `json:"available_providers"`
	RecommendedProvider    string            `json:"recommended_provider"`
	RequiredFeatures       []string          `json:"required_features,omitempty"`
}

// ProviderPriority defines the priority order for provider selection
var ProviderPriority = []string{
	"virtualbox",
	"qemu",
	"kvm",
	"vmware",
	"hyperv",
}

// GetProviderPriority returns the priority order for providers on the current platform
func GetProviderPriority(platform string) []string {
	switch platform {
	case "windows":
		return []string{"virtualbox", "hyperv", "vmware", "qemu"}
	case "linux":
		return []string{"kvm", "qemu", "virtualbox", "vmware"}
	case "darwin":
		return []string{"virtualbox", "vmware", "qemu"}
	default:
		return ProviderPriority
	}
}