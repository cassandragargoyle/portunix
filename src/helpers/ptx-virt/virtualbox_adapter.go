package main

import (
	"time"

	"portunix.ai/app/virt/types"
	"portunix.ai/app/virt/virtualbox"
)

// VirtualBoxAdapter adapts the existing VirtualBox backend to the VirtualizationProvider interface
type VirtualBoxAdapter struct {
	backend *virtualbox.Backend
}

// Provider Information

// GetName returns the provider name
func (v *VirtualBoxAdapter) GetName() string {
	return v.backend.GetName()
}

// GetVersion returns the provider version
func (v *VirtualBoxAdapter) GetVersion() (string, error) {
	return v.backend.GetVersion()
}

// IsAvailable checks if the provider is available
func (v *VirtualBoxAdapter) IsAvailable() bool {
	return v.backend.IsAvailable()
}

// GetDiagnosticInfo returns diagnostic information
func (v *VirtualBoxAdapter) GetDiagnosticInfo() *DiagnosticInfo {
	vboxDiag := v.backend.GetDiagnosticInfo()
	if vboxDiag == nil {
		return nil
	}

	// Convert VirtualBox diagnostic info to our interface
	return &DiagnosticInfo{
		Platform:          vboxDiag.Platform,
		PathEnvironment:   vboxDiag.PathEnvironment,
		RegistryKeys:      vboxDiag.RegistryKeys,
		InstallationPaths: vboxDiag.InstallationPaths,
		RunningServices:   vboxDiag.RunningServices,
		PackageManager:    vboxDiag.PackageManager,
		BrewInstalled:     vboxDiag.BrewInstalled,
		Suggestions:       vboxDiag.Suggestions,
	}
}

// VM Lifecycle Management

// Create creates a new virtual machine
func (v *VirtualBoxAdapter) Create(config *types.VMConfig) error {
	return v.backend.Create(config)
}

// Start starts a virtual machine
func (v *VirtualBoxAdapter) Start(vmName string) error {
	return v.backend.Start(vmName)
}

// Stop stops a virtual machine
func (v *VirtualBoxAdapter) Stop(vmName string, force bool) error {
	return v.backend.Stop(vmName, force)
}

// Restart restarts a virtual machine
func (v *VirtualBoxAdapter) Restart(vmName string) error {
	return v.backend.Restart(vmName)
}

// Suspend suspends a virtual machine
func (v *VirtualBoxAdapter) Suspend(vmName string) error {
	return v.backend.Suspend(vmName)
}

// Resume resumes a virtual machine
func (v *VirtualBoxAdapter) Resume(vmName string) error {
	return v.backend.Resume(vmName)
}

// Delete deletes a virtual machine
func (v *VirtualBoxAdapter) Delete(vmName string, keepDisk bool) error {
	return v.backend.Delete(vmName, keepDisk)
}

// VM Information

// List lists all virtual machines
func (v *VirtualBoxAdapter) List() ([]*types.VMInfo, error) {
	return v.backend.List()
}

// GetInfo gets information about a virtual machine
func (v *VirtualBoxAdapter) GetInfo(vmName string) (*types.VMInfo, error) {
	return v.backend.GetInfo(vmName)
}

// GetState gets the state of a virtual machine
func (v *VirtualBoxAdapter) GetState(vmName string) types.VMState {
	return v.backend.GetState(vmName)
}

// GetIP gets the IP address of a virtual machine
func (v *VirtualBoxAdapter) GetIP(vmName string) (string, error) {
	return v.backend.GetIP(vmName)
}

// VM Connection

// IsSSHReady checks if SSH is ready on the virtual machine
func (v *VirtualBoxAdapter) IsSSHReady(vmName string) bool {
	return v.backend.IsSSHReady(vmName)
}

// Connect connects to a virtual machine via SSH
func (v *VirtualBoxAdapter) Connect(vmName string, opts SSHOptions) error {
	// Convert our SSH options to VirtualBox SSH options
	// Note: VirtualBox types.SSHOptions has different fields, so we use Command only
	vboxOpts := types.SSHOptions{
		Command:     opts.Command,
		WaitTimeout: 30 * time.Second, // Use default timeout
		NoWait:      false,
		AutoStart:   true,
		CheckOnly:   false,
	}
	return v.backend.Connect(vmName, vboxOpts)
}

// Snapshot Management

// CreateSnapshot creates a snapshot of a virtual machine
func (v *VirtualBoxAdapter) CreateSnapshot(vmName, snapshotName, description string) error {
	return v.backend.CreateSnapshot(vmName, snapshotName, description)
}

// ListSnapshots lists snapshots of a virtual machine
func (v *VirtualBoxAdapter) ListSnapshots(vmName string) ([]*SnapshotInfo, error) {
	vboxSnapshots, err := v.backend.ListSnapshots(vmName)
	if err != nil {
		return nil, err
	}

	// Convert VirtualBox snapshots to our interface
	snapshots := make([]*SnapshotInfo, len(vboxSnapshots))
	for i, vboxSnapshot := range vboxSnapshots {
		snapshots[i] = &SnapshotInfo{
			Name:        vboxSnapshot.Name,
			VM:          vboxSnapshot.VM,
			Description: vboxSnapshot.Description,
			CreatedAt:   vboxSnapshot.CreatedAt,
			Size:        vboxSnapshot.Size,
			Parent:      "", // VirtualBox types.SnapshotInfo doesn't have Parent field
		}
	}

	return snapshots, nil
}

// RevertSnapshot reverts a virtual machine to a snapshot
func (v *VirtualBoxAdapter) RevertSnapshot(vmName, snapshotName string) error {
	return v.backend.RevertSnapshot(vmName, snapshotName)
}

// DeleteSnapshot deletes a snapshot
func (v *VirtualBoxAdapter) DeleteSnapshot(vmName, snapshotName string) error {
	return v.backend.DeleteSnapshot(vmName, snapshotName)
}

// File Operations

// CopyToVM copies a file to a virtual machine
func (v *VirtualBoxAdapter) CopyToVM(vmName, localPath, remotePath string) error {
	return v.backend.CopyToVM(vmName, localPath, remotePath)
}

// CopyFromVM copies a file from a virtual machine
func (v *VirtualBoxAdapter) CopyFromVM(vmName, remotePath, localPath string) error {
	return v.backend.CopyFromVM(vmName, remotePath, localPath)
}