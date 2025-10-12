package virt

import (
	"portunix.ai/app/virt/types"
)

// Re-export types for backwards compatibility
type VMState = types.VMState
type VMInfo = types.VMInfo
type VMConfig = types.VMConfig
type NetworkConfig = types.NetworkConfig
type PortForward = types.PortForward
type SnapshotInfo = types.SnapshotInfo
type SSHOptions = types.SSHOptions
type Backend = types.Backend

// Re-export constants
const (
	VMStateRunning   = types.VMStateRunning
	VMStateStopped   = types.VMStateStopped
	VMStateSuspended = types.VMStateSuspended
	VMStateError     = types.VMStateError
	VMStateNotFound  = types.VMStateNotFound
	VMStateStarting  = types.VMStateStarting
	VMStateStopping  = types.VMStateStopping
)

// Manager manages virtualization across different backends
type Manager struct {
	backend Backend
	config  *Config
}

// NewManager creates a new virtualization manager
func NewManager() (*Manager, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	backend, err := selectBackend(config)
	if err != nil {
		return nil, err
	}

	return &Manager{
		backend: backend,
		config:  config,
	}, nil
}

// GetBackend returns the current backend
func (m *Manager) GetBackend() Backend {
	return m.backend
}

// VM lifecycle methods
func (m *Manager) Create(config *VMConfig) error {
	return m.backend.Create(config)
}

func (m *Manager) Start(vmName string) error {
	return m.backend.Start(vmName)
}

func (m *Manager) Stop(vmName string, force bool) error {
	return m.backend.Stop(vmName, force)
}

func (m *Manager) Restart(vmName string) error {
	return m.backend.Restart(vmName)
}

func (m *Manager) Suspend(vmName string) error {
	return m.backend.Suspend(vmName)
}

func (m *Manager) Resume(vmName string) error {
	return m.backend.Resume(vmName)
}

func (m *Manager) Delete(vmName string, keepDisk bool) error {
	return m.backend.Delete(vmName, keepDisk)
}

// Information methods
func (m *Manager) List() ([]*VMInfo, error) {
	return m.backend.List()
}

func (m *Manager) GetInfo(vmName string) (*VMInfo, error) {
	return m.backend.GetInfo(vmName)
}

func (m *Manager) GetState(vmName string) VMState {
	return m.backend.GetState(vmName)
}

func (m *Manager) GetIP(vmName string) (string, error) {
	return m.backend.GetIP(vmName)
}

// Snapshot methods
func (m *Manager) CreateSnapshot(vmName, snapshotName, description string) error {
	return m.backend.CreateSnapshot(vmName, snapshotName, description)
}

func (m *Manager) ListSnapshots(vmName string) ([]*SnapshotInfo, error) {
	return m.backend.ListSnapshots(vmName)
}

func (m *Manager) RevertSnapshot(vmName, snapshotName string) error {
	return m.backend.RevertSnapshot(vmName, snapshotName)
}

func (m *Manager) DeleteSnapshot(vmName, snapshotName string) error {
	return m.backend.DeleteSnapshot(vmName, snapshotName)
}

// SSH and connectivity
func (m *Manager) IsSSHReady(vmName string) bool {
	return m.backend.IsSSHReady(vmName)
}

func (m *Manager) Connect(vmName string, opts SSHOptions) error {
	return m.backend.Connect(vmName, opts)
}

// File operations
func (m *Manager) CopyToVM(vmName, localPath, remotePath string) error {
	return m.backend.CopyToVM(vmName, localPath, remotePath)
}

func (m *Manager) CopyFromVM(vmName, remotePath, localPath string) error {
	return m.backend.CopyFromVM(vmName, remotePath, localPath)
}