package types

import (
	"time"
)

// VMState represents the current state of a VM
type VMState string

const (
	VMStateRunning   VMState = "running"
	VMStateStopped   VMState = "stopped"
	VMStateSuspended VMState = "suspended"
	VMStateError     VMState = "error"
	VMStateNotFound  VMState = "not-found"
	VMStateStarting  VMState = "starting"
	VMStateStopping  VMState = "stopping"
	VMStateUnknown   VMState = "unknown"
)

// VMInfo contains information about a VM
type VMInfo struct {
	Name        string    `json:"name"`
	State       VMState   `json:"state"`
	Backend     string    `json:"backend"`
	RAM         string    `json:"ram"`
	CPUs        int       `json:"cpus"`
	DiskSize    string    `json:"disk_size"`
	OSType      string    `json:"os_type"`
	IP          string    `json:"ip,omitempty"`
	VNCPort     int       `json:"vnc_port,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	LastStarted time.Time `json:"last_started,omitempty"`
	ErrorDetail string    `json:"error_detail,omitempty"`
}

// VMConfig represents VM configuration for creation
type VMConfig struct {
	Name       string            `json:"name"`
	Template   string            `json:"template,omitempty"`
	ISO        string            `json:"iso,omitempty"`
	RAM        string            `json:"ram"`
	CPUs       int               `json:"cpus"`
	DiskSize   string            `json:"disk_size"`
	OSType     string            `json:"os_type"`
	EnableSSH  bool              `json:"enable_ssh"`
	SSHKey     string            `json:"ssh_key,omitempty"`
	Network    NetworkConfig     `json:"network"`
	Features   map[string]string `json:"features,omitempty"`
	PostCreate []string          `json:"post_create,omitempty"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Mode     string            `json:"mode"` // nat, bridge, host
	Forwards []PortForward     `json:"forwards,omitempty"`
	IP       string            `json:"ip,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
}

// PortForward represents port forwarding configuration
type PortForward struct {
	Host  int    `json:"host"`
	Guest int    `json:"guest"`
	Proto string `json:"proto"` // tcp, udp
}

// SnapshotInfo contains information about a snapshot
type SnapshotInfo struct {
	Name        string    `json:"name"`
	VM          string    `json:"vm"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Size        int64     `json:"size"`
}

// SSHOptions represents options for SSH connections
type SSHOptions struct {
	Command     string        `json:"command,omitempty"`
	WaitTimeout time.Duration `json:"wait_timeout"`
	NoWait      bool          `json:"no_wait"`
	AutoStart   bool          `json:"auto_start"`
	CheckOnly   bool          `json:"check_only"`
}

// Backend represents a virtualization backend (QEMU, VirtualBox, etc.)
type Backend interface {
	// Basic lifecycle
	Create(config *VMConfig) error
	Start(vmName string) error
	Stop(vmName string, force bool) error
	Restart(vmName string) error
	Suspend(vmName string) error
	Resume(vmName string) error
	Delete(vmName string, keepDisk bool) error

	// Information
	List() ([]*VMInfo, error)
	GetInfo(vmName string) (*VMInfo, error)
	GetState(vmName string) VMState
	GetIP(vmName string) (string, error)

	// Snapshots
	CreateSnapshot(vmName, snapshotName, description string) error
	ListSnapshots(vmName string) ([]*SnapshotInfo, error)
	RevertSnapshot(vmName, snapshotName string) error
	DeleteSnapshot(vmName, snapshotName string) error

	// SSH and connectivity
	IsSSHReady(vmName string) bool
	Connect(vmName string, opts SSHOptions) error

	// File operations
	CopyToVM(vmName, localPath, remotePath string) error
	CopyFromVM(vmName, remotePath, localPath string) error

	// Configuration
	GetName() string
	IsAvailable() bool
	GetVersion() (string, error)
}