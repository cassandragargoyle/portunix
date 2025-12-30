package main

import (
	"fmt"
	"os/exec"
	"strings"

	"portunix.ai/app/system"
	"portunix.ai/app/virt"
	"portunix.ai/app/virt/types"
)

// QEMUAdapter implements VirtualizationProvider interface for QEMU/KVM
type QEMUAdapter struct {
	hasKVM bool
}

// Provider Information

// GetName returns the provider name
func (q *QEMUAdapter) GetName() string {
	if q.hasKVM {
		return "kvm"
	}
	return "qemu"
}

// GetVersion returns QEMU version
func (q *QEMUAdapter) GetVersion() (string, error) {
	version := system.GetQEMUVersion()
	if version == "" {
		return "", fmt.Errorf("failed to get QEMU version")
	}
	// Remove 'v' prefix if present
	return strings.TrimPrefix(version, "v"), nil
}

// IsAvailable checks if QEMU is installed
func (q *QEMUAdapter) IsAvailable() bool {
	// Check if qemu-system-x86_64 exists
	_, err := exec.LookPath("qemu-system-x86_64")
	if err != nil {
		// Try alternative 'kvm' command
		_, err = exec.LookPath("kvm")
	}

	// Check KVM availability
	q.hasKVM = q.checkKVMModules()

	return err == nil
}

// GetDiagnosticInfo returns diagnostic information
func (q *QEMUAdapter) GetDiagnosticInfo() *DiagnosticInfo {
	diag := &DiagnosticInfo{
		Suggestions: []string{},
		Details:     []string{},
	}

	// Find QEMU path
	if path, err := exec.LookPath("qemu-system-x86_64"); err == nil {
		diag.PathEnvironment = path
	} else if path, err := exec.LookPath("kvm"); err == nil {
		diag.PathEnvironment = path
	}

	// Check KVM acceleration
	if q.hasKVM {
		diag.Details = append(diag.Details, "KVM acceleration: enabled")

		// Check which KVM module is loaded
		if kmods := q.getLoadedKVMModules(); len(kmods) > 0 {
			diag.Details = append(diag.Details, fmt.Sprintf("Loaded modules: %s", strings.Join(kmods, ", ")))
		}
	} else {
		diag.Details = append(diag.Details, "KVM acceleration: disabled (software emulation only)")
		diag.Suggestions = append(diag.Suggestions, "Load KVM kernel module: sudo modprobe kvm kvm_amd (or kvm_intel)")
	}

	// Check for VirtualBox conflict
	if q.hasKVM && q.isVirtualBoxConflict() {
		diag.Suggestions = append(diag.Suggestions, "⚠️  VirtualBox/KVM conflict detected. Run: portunix virt check --fix")
	}

	// Check libvirt status
	if libvirtStatus, err := virt.DetectLibvirtStatus(); err == nil {
		if libvirtStatus.Installed {
			if libvirtStatus.Running || libvirtStatus.SocketActivated {
				statusText := "running"
				if libvirtStatus.SocketActivated {
					statusText = "socket-activated"
				}
				diag.Details = append(diag.Details,
					fmt.Sprintf("Libvirt %s: %s (%s)", libvirtStatus.Version, statusText, libvirtStatus.DaemonType))
			} else {
				// Libvirt NOT running - this is critical for virt-manager
				diag.Details = append([]string{fmt.Sprintf("❌ Libvirt %s: NOT running (virt-manager won't connect!)", libvirtStatus.Version)}, diag.Details...)
				diag.Suggestions = append(diag.Suggestions, libvirtStatus.Recommendations...)
			}
		} else {
			diag.Suggestions = append(diag.Suggestions, "Install libvirt: portunix install libvirt")
		}
	}

	return diag
}

// checkKVMModules checks if KVM kernel modules are loaded
func (q *QEMUAdapter) checkKVMModules() bool {
	cmd := exec.Command("lsmod")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "kvm")
}

// getLoadedKVMModules returns list of loaded KVM modules
func (q *QEMUAdapter) getLoadedKVMModules() []string {
	cmd := exec.Command("lsmod")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	modules := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "kvm") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				modules = append(modules, parts[0])
			}
		}
	}
	return modules
}

// isVirtualBoxConflict checks if VirtualBox is installed (potential conflict)
func (q *QEMUAdapter) isVirtualBoxConflict() bool {
	_, err := exec.LookPath("VBoxManage")
	return err == nil
}

// VM Lifecycle Management (stubs for basic adapter)

// Create creates a new virtual machine
func (q *QEMUAdapter) Create(config *types.VMConfig) error {
	return fmt.Errorf("VM creation requires libvirt (use: portunix virt install-qemu)")
}

// Start starts a virtual machine
func (q *QEMUAdapter) Start(vmName string) error {
	return fmt.Errorf("VM management requires libvirt (use: portunix virt install-qemu)")
}

// Stop stops a virtual machine
func (q *QEMUAdapter) Stop(vmName string, force bool) error {
	return fmt.Errorf("VM management requires libvirt (use: portunix virt install-qemu)")
}

// Restart restarts a virtual machine
func (q *QEMUAdapter) Restart(vmName string) error {
	return fmt.Errorf("VM management requires libvirt (use: portunix virt install-qemu)")
}

// Suspend suspends a virtual machine
func (q *QEMUAdapter) Suspend(vmName string) error {
	return fmt.Errorf("VM management requires libvirt (use: portunix virt install-qemu)")
}

// Resume resumes a virtual machine
func (q *QEMUAdapter) Resume(vmName string) error {
	return fmt.Errorf("VM management requires libvirt (use: portunix virt install-qemu)")
}

// Delete deletes a virtual machine
func (q *QEMUAdapter) Delete(vmName string, keepDisk bool) error {
	return fmt.Errorf("VM management requires libvirt (use: portunix virt install-qemu)")
}

// VM Information

// List returns list of VMs (not implemented for basic adapter)
func (q *QEMUAdapter) List() ([]*types.VMInfo, error) {
	// Basic QEMU adapter doesn't track VMs
	// This would require libvirt integration
	return nil, fmt.Errorf("VM listing requires libvirt (use: portunix virt install-qemu)")
}

// GetInfo gets information about a virtual machine
func (q *QEMUAdapter) GetInfo(vmName string) (*types.VMInfo, error) {
	return nil, fmt.Errorf("VM info requires libvirt (use: portunix virt install-qemu)")
}

// GetState returns the state of a virtual machine
func (q *QEMUAdapter) GetState(vmName string) types.VMState {
	return types.VMStateUnknown
}

// GetIP gets the IP address of a virtual machine
func (q *QEMUAdapter) GetIP(vmName string) (string, error) {
	return "", fmt.Errorf("VM IP detection requires libvirt (use: portunix virt install-qemu)")
}

// VM Connection

// IsSSHReady checks if SSH is ready on the VM
func (q *QEMUAdapter) IsSSHReady(vmName string) bool {
	return false
}

// Connect connects to a virtual machine via SSH
func (q *QEMUAdapter) Connect(vmName string, opts SSHOptions) error {
	return fmt.Errorf("VM SSH connection requires libvirt (use: portunix virt install-qemu)")
}

// Snapshot Management

// CreateSnapshot creates a snapshot of a virtual machine
func (q *QEMUAdapter) CreateSnapshot(vmName, snapshotName, description string) error {
	return fmt.Errorf("Snapshot management requires libvirt (use: portunix virt install-qemu)")
}

// ListSnapshots lists all snapshots of a virtual machine
func (q *QEMUAdapter) ListSnapshots(vmName string) ([]*SnapshotInfo, error) {
	return nil, fmt.Errorf("Snapshot management requires libvirt (use: portunix virt install-qemu)")
}

// RevertSnapshot reverts to a snapshot
func (q *QEMUAdapter) RevertSnapshot(vmName, snapshotName string) error {
	return fmt.Errorf("Snapshot management requires libvirt (use: portunix virt install-qemu)")
}

// DeleteSnapshot deletes a snapshot
func (q *QEMUAdapter) DeleteSnapshot(vmName, snapshotName string) error {
	return fmt.Errorf("Snapshot management requires libvirt (use: portunix virt install-qemu)")
}

// File Operations

// CopyToVM copies a file to a virtual machine
func (q *QEMUAdapter) CopyToVM(vmName, localPath, remotePath string) error {
	return fmt.Errorf("File operations require libvirt (use: portunix virt install-qemu)")
}

// CopyFromVM copies a file from a virtual machine
func (q *QEMUAdapter) CopyFromVM(vmName, remotePath, localPath string) error {
	return fmt.Errorf("File operations require libvirt (use: portunix virt install-qemu)")
}
