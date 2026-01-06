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
	hasKVM        bool
	libvirtStatus *virt.LibvirtStatus
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

// checkLibvirtReady checks if libvirt is available and running
// Returns nil if ready, error with user-friendly message otherwise
func (q *QEMUAdapter) checkLibvirtReady() error {
	status, err := virt.DetectLibvirtStatus()
	if err != nil {
		return fmt.Errorf("failed to detect libvirt status: %w", err)
	}
	q.libvirtStatus = status

	if !status.Installed {
		return fmt.Errorf("libvirt is not installed (use: portunix install libvirt)")
	}

	if !status.Running && !status.SocketActivated {
		return fmt.Errorf("libvirt daemon is not running (use: portunix virt check --fix-libvirt)")
	}

	return nil
}

// parseVirshListOutput parses output from 'virsh list --all'
// Example output:
//
//	Id   Name       State
//	--------------------------
//	1    myvm       running
//	-    testvm     shut off
func (q *QEMUAdapter) parseVirshListOutput(output []byte) []*types.VMInfo {
	vms := []*types.VMInfo{}
	lines := strings.Split(string(output), "\n")

	// Skip header lines (first 2 lines)
	for i, line := range lines {
		if i < 2 {
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// Fields: ID, Name, State (state can be two words like "shut off")
		var name string
		var stateStr string

		if fields[0] == "-" {
			// VM is not running, ID is "-"
			name = fields[1]
			if len(fields) >= 3 {
				stateStr = strings.Join(fields[2:], " ")
			}
		} else {
			// VM is running, ID is numeric
			name = fields[1]
			if len(fields) >= 3 {
				stateStr = strings.Join(fields[2:], " ")
			}
		}

		vm := &types.VMInfo{
			Name:    name,
			State:   q.parseVirshState(stateStr),
			Backend: "libvirt",
		}

		vms = append(vms, vm)
	}

	return vms
}

// parseVirshState converts virsh state string to VMState
func (q *QEMUAdapter) parseVirshState(state string) types.VMState {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "running":
		return types.VMStateRunning
	case "shut off", "shutoff":
		return types.VMStateStopped
	case "paused":
		return types.VMStateSuspended
	case "in shutdown":
		return types.VMStateStopping
	case "idle", "crashed", "dying":
		return types.VMStateError
	case "pmsuspended":
		return types.VMStateSuspended
	default:
		return types.VMStateUnknown
	}
}

// parseVirshDominfo parses output from 'virsh dominfo <vm>'
func (q *QEMUAdapter) parseVirshDominfo(output []byte, vmName string) *types.VMInfo {
	vm := &types.VMInfo{
		Name:    vmName,
		Backend: "libvirt",
		State:   types.VMStateUnknown,
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "State":
			vm.State = q.parseVirshState(value)
		case "Max memory":
			vm.RAM = value
		case "CPU(s)":
			if cpus, err := fmt.Sscanf(value, "%d", &vm.CPUs); err == nil && cpus > 0 {
				// CPUs parsed successfully
			}
		case "OS Type":
			vm.OSType = value
		}
	}

	return vm
}

// VM Lifecycle Management (stubs for basic adapter)

// Create creates a new virtual machine
func (q *QEMUAdapter) Create(config *types.VMConfig) error {
	return fmt.Errorf("VM creation requires libvirt (use: portunix virt install-qemu)")
}

// Start starts a virtual machine using virsh
func (q *QEMUAdapter) Start(vmName string) error {
	// Check if libvirt is ready
	if err := q.checkLibvirtReady(); err != nil {
		return err
	}

	// Check current state first
	state := q.GetState(vmName)
	if state == types.VMStateRunning {
		return nil // Already running
	}
	if state == types.VMStateNotFound {
		return fmt.Errorf("VM '%s' not found", vmName)
	}

	// Execute virsh start
	cmd := exec.Command("virsh", "start", vmName)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		// Parse error message for better user feedback
		if strings.Contains(outputStr, "Domain not found") {
			return fmt.Errorf("VM '%s' not found", vmName)
		}
		if strings.Contains(outputStr, "already active") {
			return nil // Already running, not an error
		}
		if strings.Contains(outputStr, "Permission denied") || strings.Contains(outputStr, "authentication") {
			return fmt.Errorf("permission denied - add user to libvirt group: sudo usermod -aG libvirt $USER")
		}
		// Return full output for debugging
		if outputStr != "" {
			return fmt.Errorf("%s", outputStr)
		}
		return fmt.Errorf("virsh start failed: %v", err)
	}

	return nil
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

// List returns list of VMs using virsh
func (q *QEMUAdapter) List() ([]*types.VMInfo, error) {
	// Check if libvirt is ready
	if err := q.checkLibvirtReady(); err != nil {
		return nil, err
	}

	// Execute virsh list --all
	cmd := exec.Command("virsh", "list", "--all")
	output, err := cmd.Output()
	if err != nil {
		// Check for permission error
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "Permission denied") ||
				strings.Contains(stderr, "authentication") ||
				strings.Contains(stderr, "polkit") {
				return nil, fmt.Errorf("permission denied - add user to libvirt group: sudo usermod -aG libvirt $USER")
			}
		}
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	return q.parseVirshListOutput(output), nil
}

// GetInfo gets information about a virtual machine using virsh dominfo
func (q *QEMUAdapter) GetInfo(vmName string) (*types.VMInfo, error) {
	// Check if libvirt is ready
	if err := q.checkLibvirtReady(); err != nil {
		return nil, err
	}

	// Execute virsh dominfo
	cmd := exec.Command("virsh", "dominfo", vmName)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "Domain not found") ||
				strings.Contains(stderr, "failed to get domain") {
				return nil, fmt.Errorf("VM '%s' not found", vmName)
			}
		}
		return nil, fmt.Errorf("failed to get VM info: %w", err)
	}

	vm := q.parseVirshDominfo(output, vmName)

	// Try to get IP address
	if ip, err := q.GetIP(vmName); err == nil {
		vm.IP = ip
	}

	return vm, nil
}

// GetState returns the state of a virtual machine using virsh domstate
func (q *QEMUAdapter) GetState(vmName string) types.VMState {
	// Check if libvirt is ready
	if err := q.checkLibvirtReady(); err != nil {
		return types.VMStateUnknown
	}

	// Execute virsh domstate
	cmd := exec.Command("virsh", "domstate", vmName)
	output, err := cmd.Output()
	if err != nil {
		return types.VMStateNotFound
	}

	return q.parseVirshState(string(output))
}

// GetIP gets the IP address of a virtual machine using virsh domifaddr
func (q *QEMUAdapter) GetIP(vmName string) (string, error) {
	// Check if libvirt is ready
	if err := q.checkLibvirtReady(); err != nil {
		return "", err
	}

	// Try virsh domifaddr first (requires qemu-guest-agent or DHCP lease)
	cmd := exec.Command("virsh", "domifaddr", vmName)
	output, err := cmd.Output()
	if err == nil {
		if ip := q.parseVirshDomifaddr(output); ip != "" {
			return ip, nil
		}
	}

	// Fallback: try with --source lease
	cmd = exec.Command("virsh", "domifaddr", vmName, "--source", "lease")
	output, err = cmd.Output()
	if err == nil {
		if ip := q.parseVirshDomifaddr(output); ip != "" {
			return ip, nil
		}
	}

	// Fallback: try with --source arp
	cmd = exec.Command("virsh", "domifaddr", vmName, "--source", "arp")
	output, err = cmd.Output()
	if err == nil {
		if ip := q.parseVirshDomifaddr(output); ip != "" {
			return ip, nil
		}
	}

	return "", fmt.Errorf("could not determine IP address for VM '%s'", vmName)
}

// parseVirshDomifaddr parses output from 'virsh domifaddr <vm>'
// Example output:
//
//	Name       MAC address          Protocol     Address
//	-------------------------------------------------------------------------------
//	vnet0      52:54:00:xx:xx:xx    ipv4         192.168.122.100/24
func (q *QEMUAdapter) parseVirshDomifaddr(output []byte) string {
	lines := strings.Split(string(output), "\n")

	// Skip header lines
	for i, line := range lines {
		if i < 2 {
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 4 {
			// Get the IP address (may have /24 suffix)
			addr := fields[3]
			// Remove CIDR notation if present
			if idx := strings.Index(addr, "/"); idx != -1 {
				addr = addr[:idx]
			}
			// Return first valid IPv4 address
			if strings.Contains(addr, ".") {
				return addr
			}
		}
	}

	return ""
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
