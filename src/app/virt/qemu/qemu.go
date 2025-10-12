package qemu

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"portunix.ai/app/virt/types"
)


// Backend implements the QEMU/KVM backend
type Backend struct {
	vmDir string
}

// NewBackend creates a new QEMU backend
func NewBackend() *Backend {
	homeDir, _ := os.UserHomeDir()
	vmDir := filepath.Join(homeDir, ".portunix", "vms")
	os.MkdirAll(vmDir, 0755)

	return &Backend{
		vmDir: vmDir,
	}
}

// GetName returns the backend name
func (b *Backend) GetName() string {
	return "qemu"
}

// IsAvailable checks if QEMU is available
func (b *Backend) IsAvailable() bool {
	_, err := exec.LookPath("qemu-system-x86_64")
	return err == nil
}

// GetVersion returns QEMU version
func (b *Backend) GetVersion() (string, error) {
	cmd := exec.Command("qemu-system-x86_64", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse version from output
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		// Extract version number from first line
		re := regexp.MustCompile(`version\s+(\d+\.\d+\.\d+)`)
		matches := re.FindStringSubmatch(lines[0])
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return strings.TrimSpace(string(output)), nil
}

// Create creates a new VM
func (b *Backend) Create(config *types.VMConfig) error {
	vmPath := filepath.Join(b.vmDir, config.Name)
	if err := os.MkdirAll(vmPath, 0755); err != nil {
		return fmt.Errorf("failed to create VM directory: %w", err)
	}

	// Create disk image
	diskPath := filepath.Join(vmPath, fmt.Sprintf("%s.qcow2", config.Name))
	if err := b.createDisk(diskPath, config.DiskSize); err != nil {
		return fmt.Errorf("failed to create disk: %w", err)
	}

	// Create VM configuration
	vmConfig := &QEMUVMConfig{
		Name:     config.Name,
		DiskPath: diskPath,
		RAM:      config.RAM,
		CPUs:     config.CPUs,
		OSType:   config.OSType,
		ISO:      config.ISO,
		Network:  config.Network,
		Features: config.Features,
		Created:  time.Now(),
	}

	// Save configuration
	configPath := filepath.Join(vmPath, "config.json")
	configData, err := json.MarshalIndent(vmConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// Start starts a VM
func (b *Backend) Start(vmName string) error {
	config, err := b.loadVMConfig(vmName)
	if err != nil {
		return err
	}

	// Check if already running
	if b.isVMRunning(vmName) {
		return nil
	}

	args := b.buildQEMUArgs(config)

	cmd := exec.Command("qemu-system-x86_64", args...)
	cmd.Dir = filepath.Join(b.vmDir, vmName)

	// Start in background
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start VM: %w", err)
	}

	// Save PID for management
	pidPath := filepath.Join(b.vmDir, vmName, "qemu.pid")
	pidData := fmt.Sprintf("%d", cmd.Process.Pid)
	os.WriteFile(pidPath, []byte(pidData), 0644)

	return nil
}

// Stop stops a VM
func (b *Backend) Stop(vmName string, force bool) error {
	pid, err := b.getVMPID(vmName)
	if err != nil {
		return err
	}

	if pid == 0 {
		return nil // Already stopped
	}

	if force {
		// Force kill
		return b.killProcess(pid)
	}

	// Graceful shutdown via QEMU monitor
	return b.gracefulShutdown(vmName, pid)
}

// Restart restarts a VM
func (b *Backend) Restart(vmName string) error {
	if err := b.Stop(vmName, false); err != nil {
		return err
	}

	// Wait a bit for clean shutdown
	time.Sleep(2 * time.Second)

	return b.Start(vmName)
}

// Suspend suspends a VM
func (b *Backend) Suspend(vmName string) error {
	// Use QEMU monitor to pause
	return b.executeMonitorCommand(vmName, "stop")
}

// Resume resumes a VM
func (b *Backend) Resume(vmName string) error {
	// Use QEMU monitor to continue
	return b.executeMonitorCommand(vmName, "cont")
}

// Delete deletes a VM
func (b *Backend) Delete(vmName string, keepDisk bool) error {
	// Stop VM first
	b.Stop(vmName, true)

	vmPath := filepath.Join(b.vmDir, vmName)

	if !keepDisk {
		// Remove entire VM directory
		return os.RemoveAll(vmPath)
	}

	// Keep disk but remove other files
	files := []string{"config.json", "qemu.pid", "monitor.sock", "run.sh"}
	for _, file := range files {
		os.Remove(filepath.Join(vmPath, file))
	}

	return nil
}

// List lists all VMs
func (b *Backend) List() ([]*types.VMInfo, error) {
	var vms []*types.VMInfo

	entries, err := os.ReadDir(b.vmDir)
	if err != nil {
		return vms, nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		vmName := entry.Name()
		info, err := b.GetInfo(vmName)
		if err != nil {
			continue
		}

		vms = append(vms, info)
	}

	return vms, nil
}

// GetInfo gets VM information
func (b *Backend) GetInfo(vmName string) (*types.VMInfo, error) {
	config, err := b.loadVMConfig(vmName)
	if err != nil {
		return nil, err
	}

	state := b.GetState(vmName)

	info := &types.VMInfo{
		Name:      vmName,
		State:     state,
		Backend:   "qemu",
		RAM:       config.RAM,
		CPUs:      config.CPUs,
		DiskSize:  b.getDiskSize(config.DiskPath),
		OSType:    config.OSType,
		CreatedAt: config.Created,
	}

	if state == types.VMStateRunning {
		if ip, err := b.GetIP(vmName); err == nil {
			info.IP = ip
		}
		info.LastStarted = time.Now() // TODO: Track actual start time
	}

	return info, nil
}

// GetState gets VM state
func (b *Backend) GetState(vmName string) types.VMState {
	if b.isVMRunning(vmName) {
		return types.VMStateRunning
	}
	return types.VMStateStopped
}

// GetIP gets VM IP address
func (b *Backend) GetIP(vmName string) (string, error) {
	// Try to get IP from DHCP leases or ARP table
	// This is a simplified implementation
	return "127.0.0.1", nil // TODO: Implement proper IP detection
}

// IsSSHReady checks if SSH is ready
func (b *Backend) IsSSHReady(vmName string) bool {
	ip, err := b.GetIP(vmName)
	if err != nil {
		return false
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "22"), 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

// Connect connects to VM via SSH
func (b *Backend) Connect(vmName string, opts types.SSHOptions) error {
	ip, err := b.GetIP(vmName)
	if err != nil {
		return err
	}

	var args []string
	if opts.Command != "" {
		args = []string{"ssh", ip, opts.Command}
	} else {
		args = []string{"ssh", ip}
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Snapshot operations
func (b *Backend) CreateSnapshot(vmName, snapshotName, description string) error {
	config, err := b.loadVMConfig(vmName)
	if err != nil {
		return err
	}

	cmd := exec.Command("qemu-img", "snapshot", "-c", snapshotName, config.DiskPath)
	return cmd.Run()
}

func (b *Backend) ListSnapshots(vmName string) ([]*types.SnapshotInfo, error) {
	config, err := b.loadVMConfig(vmName)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("qemu-img", "snapshot", "-l", config.DiskPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return b.parseSnapshots(string(output), vmName), nil
}

func (b *Backend) RevertSnapshot(vmName, snapshotName string) error {
	config, err := b.loadVMConfig(vmName)
	if err != nil {
		return err
	}

	// VM must be stopped for snapshot revert
	if b.isVMRunning(vmName) {
		if err := b.Stop(vmName, true); err != nil {
			return err
		}
	}

	cmd := exec.Command("qemu-img", "snapshot", "-a", snapshotName, config.DiskPath)
	return cmd.Run()
}

func (b *Backend) DeleteSnapshot(vmName, snapshotName string) error {
	config, err := b.loadVMConfig(vmName)
	if err != nil {
		return err
	}

	cmd := exec.Command("qemu-img", "snapshot", "-d", snapshotName, config.DiskPath)
	return cmd.Run()
}

// File operations
func (b *Backend) CopyToVM(vmName, localPath, remotePath string) error {
	ip, err := b.GetIP(vmName)
	if err != nil {
		return err
	}

	cmd := exec.Command("scp", localPath, fmt.Sprintf("%s:%s", ip, remotePath))
	return cmd.Run()
}

func (b *Backend) CopyFromVM(vmName, remotePath, localPath string) error {
	ip, err := b.GetIP(vmName)
	if err != nil {
		return err
	}

	cmd := exec.Command("scp", fmt.Sprintf("%s:%s", ip, remotePath), localPath)
	return cmd.Run()
}

// Helper methods
func (b *Backend) createDisk(diskPath, size string) error {
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", diskPath, size)
	return cmd.Run()
}

func (b *Backend) loadVMConfig(vmName string) (*QEMUVMConfig, error) {
	configPath := filepath.Join(b.vmDir, vmName, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("VM not found: %s", vmName)
	}

	var config QEMUVMConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid VM config: %w", err)
	}

	return &config, nil
}

func (b *Backend) isVMRunning(vmName string) bool {
	pid, err := b.getVMPID(vmName)
	if err != nil || pid == 0 {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to test if process exists
	err = process.Signal(os.Signal(nil))
	return err == nil
}

func (b *Backend) getVMPID(vmName string) (int, error) {
	pidPath := filepath.Join(b.vmDir, vmName, "qemu.pid")
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, nil
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}

	return pid, nil
}

func (b *Backend) killProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Kill()
}

func (b *Backend) gracefulShutdown(vmName string, pid int) error {
	// Try ACPI shutdown first
	if err := b.executeMonitorCommand(vmName, "system_powerdown"); err == nil {
		// Wait for graceful shutdown
		for i := 0; i < 30; i++ {
			if !b.isVMRunning(vmName) {
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}

	// Force kill if graceful shutdown failed
	return b.killProcess(pid)
}

func (b *Backend) executeMonitorCommand(vmName, command string) error {
	// TODO: Implement QEMU monitor communication
	// For now, use kill signals
	pid, err := b.getVMPID(vmName)
	if err != nil {
		return err
	}

	if pid == 0 {
		return fmt.Errorf("VM not running")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	// Use SIGTERM for graceful shutdown
	return process.Signal(os.Interrupt)
}

func (b *Backend) getDiskSize(diskPath string) string {
	cmd := exec.Command("qemu-img", "info", "--output=json", diskPath)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	var info struct {
		VirtualSize int64 `json:"virtual-size"`
	}

	if err := json.Unmarshal(output, &info); err != nil {
		return "unknown"
	}

	return formatBytes(info.VirtualSize)
}

func (b *Backend) parseSnapshots(output, vmName string) []*types.SnapshotInfo {
	var snapshots []*types.SnapshotInfo

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "ID") && strings.Contains(line, "TAG") {
			continue // Skip header
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			snapshot := &types.SnapshotInfo{
				Name:      fields[1],
				VM:        vmName,
				CreatedAt: time.Now(), // TODO: Parse actual timestamp
			}
			snapshots = append(snapshots, snapshot)
		}
	}

	return snapshots
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (b *Backend) buildQEMUArgs(config *QEMUVMConfig) []string {
	var args []string

	// Basic arguments
	args = append(args, "-enable-kvm")
	args = append(args, "-name", config.Name)
	args = append(args, "-m", config.RAM)
	args = append(args, "-smp", fmt.Sprintf("%d", config.CPUs))
	args = append(args, "-cpu", "host")

	// Disk
	args = append(args, "-drive", fmt.Sprintf("file=%s,format=qcow2,if=virtio", config.DiskPath))

	// ISO/CDROM
	if config.ISO != "" {
		args = append(args, "-cdrom", config.ISO)
		args = append(args, "-boot", "d")
	}

	// Network
	args = append(args, "-netdev", "user,id=net0")
	args = append(args, "-device", "virtio-net,netdev=net0")

	// Graphics
	args = append(args, "-vga", "virtio")
	args = append(args, "-display", "gtk")

	// Features based on OS type
	if config.OSType == "windows11" || config.OSType == "windows10" {
		args = append(args, "-machine", "q35,smm=on")

		// UEFI support for Windows
		if ovmfCode := findOVMFCode(); ovmfCode != "" {
			ovmfVars := strings.Replace(config.DiskPath, ".qcow2", "_VARS.fd", 1)

			// Copy OVMF_VARS template if it doesn't exist
			if _, err := os.Stat(ovmfVars); os.IsNotExist(err) {
				exec.Command("cp", "/usr/share/OVMF/OVMF_VARS.fd", ovmfVars).Run()
			}

			args = append(args, "-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", ovmfCode))
			args = append(args, "-drive", fmt.Sprintf("if=pflash,format=raw,file=%s", ovmfVars))
		}

		// TPM for Windows 11
		if config.OSType == "windows11" {
			args = append(args, "-tpmdev", "emulator,id=tpm0,chardev=chrtpm")
			args = append(args, "-chardev", "socket,id=chrtpm,path=/tmp/swtpm-sock")
			args = append(args, "-device", "tpm-tis,tpmdev=tpm0")
		}
	}

	return args
}

func findOVMFCode() string {
	paths := []string{
		"/usr/share/OVMF/OVMF_CODE.fd",
		"/usr/share/edk2-ovmf/x64/OVMF_CODE.fd",
		"/usr/share/qemu/OVMF_CODE.fd",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// QEMUVMConfig represents QEMU VM configuration
type QEMUVMConfig struct {
	Name     string                 `json:"name"`
	DiskPath string                 `json:"disk_path"`
	RAM      string                 `json:"ram"`
	CPUs     int                    `json:"cpus"`
	OSType   string                 `json:"os_type"`
	ISO      string                 `json:"iso,omitempty"`
	Network  types.NetworkConfig     `json:"network"`
	Features map[string]string      `json:"features,omitempty"`
	Created  time.Time              `json:"created"`
}