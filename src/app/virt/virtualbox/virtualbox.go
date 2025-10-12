package virtualbox

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"portunix.ai/app/system"
	"portunix.ai/app/virt/types"
)

// Backend implements the VirtualBox backend
type Backend struct {
	vmDir        string
	vboxManagePath string
}

// NewBackend creates a new VirtualBox backend
func NewBackend() *Backend {
	homeDir, _ := os.UserHomeDir()
	vmDir := filepath.Join(homeDir, ".portunix", "vms")
	os.MkdirAll(vmDir, 0755)

	backend := &Backend{
		vmDir: vmDir,
	}

	// Detect and cache VBoxManage path
	backend.vboxManagePath = backend.detectVirtualBox()

	return backend
}

// GetName returns the backend name
func (b *Backend) GetName() string {
	return "virtualbox"
}

// IsAvailable checks if VirtualBox is available
func (b *Backend) IsAvailable() bool {
	// Enhanced detection for Windows and cross-platform compatibility
	return b.detectVirtualBox() != ""
}

// GetVersion returns VirtualBox version
func (b *Backend) GetVersion() (string, error) {
	cmd := exec.Command(b.getVBoxManageCommand(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(output))
	// Remove revision info if present (e.g., "7.0.12r159484" -> "7.0.12")
	if idx := strings.Index(version, "r"); idx != -1 {
		version = version[:idx]
	}

	return version, nil
}

// Create creates a new VM
func (b *Backend) Create(config *types.VMConfig) error {
	vmPath := filepath.Join(b.vmDir, config.Name)
	if err := os.MkdirAll(vmPath, 0755); err != nil {
		return fmt.Errorf("failed to create VM directory: %w", err)
	}

	// Convert RAM to MB
	ramMB, err := b.parseRAMToMB(config.RAM)
	if err != nil {
		return fmt.Errorf("invalid RAM specification: %w", err)
	}

	// Convert disk size to MB
	diskMB, err := b.parseDiskToMB(config.DiskSize)
	if err != nil {
		return fmt.Errorf("invalid disk size specification: %w", err)
	}

	// Create VM
	if err := b.vboxManage("createvm", "--name", config.Name, "--ostype", b.getOSType(config.OSType), "--register"); err != nil {
		return fmt.Errorf("failed to create VM: %w", err)
	}

	// Configure VM
	configCmds := [][]string{
		{"modifyvm", config.Name, "--memory", fmt.Sprintf("%d", ramMB)},
		{"modifyvm", config.Name, "--cpus", fmt.Sprintf("%d", config.CPUs)},
		{"modifyvm", config.Name, "--vram", "128"},
		{"modifyvm", config.Name, "--graphicscontroller", "vmsvga"},
		{"modifyvm", config.Name, "--nic1", "nat"},
		{"modifyvm", config.Name, "--audio", "pulse"},
		{"modifyvm", config.Name, "--clipboard", "bidirectional"},
		{"modifyvm", config.Name, "--draganddrop", "bidirectional"},
	}

	// Add OS-specific configurations
	if config.OSType == "windows11" {
		configCmds = append(configCmds, [][]string{
			{"modifyvm", config.Name, "--firmware", "efi"},
			{"modifyvm", config.Name, "--tpm-type", "2.0"},
			{"modifyvm", config.Name, "--secure-boot", "on"},
		}...)
	} else if config.OSType == "windows10" {
		configCmds = append(configCmds, [][]string{
			{"modifyvm", config.Name, "--firmware", "efi"},
		}...)
	}

	for _, cmd := range configCmds {
		if err := b.vboxManage(cmd...); err != nil {
			return fmt.Errorf("failed to configure VM: %w", err)
		}
	}

	// Create storage controller
	if err := b.vboxManage("storagectl", config.Name, "--name", "SATA", "--add", "sata", "--controller", "IntelAhci"); err != nil {
		return fmt.Errorf("failed to create storage controller: %w", err)
	}

	// Create and attach hard disk
	vdiPath := filepath.Join(vmPath, fmt.Sprintf("%s.vdi", config.Name))
	if err := b.vboxManage("createmedium", "disk", "--filename", vdiPath, "--size", fmt.Sprintf("%d", diskMB)); err != nil {
		return fmt.Errorf("failed to create disk: %w", err)
	}

	if err := b.vboxManage("storageattach", config.Name, "--storagectl", "SATA", "--port", "0", "--device", "0", "--type", "hdd", "--medium", vdiPath); err != nil {
		return fmt.Errorf("failed to attach disk: %w", err)
	}

	// Attach ISO if provided
	if config.ISO != "" {
		// Create DVD controller
		if err := b.vboxManage("storagectl", config.Name, "--name", "IDE", "--add", "ide"); err != nil {
			return fmt.Errorf("failed to create IDE controller: %w", err)
		}

		if err := b.vboxManage("storageattach", config.Name, "--storagectl", "IDE", "--port", "0", "--device", "0", "--type", "dvddrive", "--medium", config.ISO); err != nil {
			return fmt.Errorf("failed to attach ISO: %w", err)
		}

		// Set boot order to DVD first
		if err := b.vboxManage("modifyvm", config.Name, "--boot1", "dvd", "--boot2", "disk"); err != nil {
			return fmt.Errorf("failed to set boot order: %w", err)
		}
	}

	// Save VM configuration
	vmConfig := &VBoxVMConfig{
		Name:      config.Name,
		DiskPath:  vdiPath,
		RAM:       config.RAM,
		CPUs:      config.CPUs,
		OSType:    config.OSType,
		ISO:       config.ISO,
		Network:   config.Network,
		Features:  config.Features,
		Created:   time.Now(),
	}

	configPath := filepath.Join(vmPath, "config.json")
	configData, err := json.MarshalIndent(vmConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, configData, 0644)
}

// Start starts a VM
func (b *Backend) Start(vmName string) error {
	state := b.GetState(vmName)
	if state == types.VMStateRunning {
		return nil
	}

	return b.vboxManage("startvm", vmName, "--type", "gui")
}

// Stop stops a VM
func (b *Backend) Stop(vmName string, force bool) error {
	state := b.GetState(vmName)
	if state == types.VMStateStopped {
		return nil
	}

	if force {
		return b.vboxManage("controlvm", vmName, "poweroff")
	}

	return b.vboxManage("controlvm", vmName, "acpipowerbutton")
}

// Restart restarts a VM
func (b *Backend) Restart(vmName string) error {
	if err := b.Stop(vmName, false); err != nil {
		return err
	}

	// Wait for shutdown
	for i := 0; i < 30; i++ {
		if b.GetState(vmName) == types.VMStateStopped {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return b.Start(vmName)
}

// Suspend suspends a VM
func (b *Backend) Suspend(vmName string) error {
	return b.vboxManage("controlvm", vmName, "savestate")
}

// Resume resumes a VM
func (b *Backend) Resume(vmName string) error {
	return b.Start(vmName) // VirtualBox doesn't have separate resume command
}

// Delete deletes a VM
func (b *Backend) Delete(vmName string, keepDisk bool) error {
	// Stop VM first
	b.Stop(vmName, true)

	if keepDisk {
		// Unregister VM but keep disk
		return b.vboxManage("unregistervm", vmName)
	}

	// Remove VM and all files
	return b.vboxManage("unregistervm", vmName, "--delete")
}

// List lists all VMs
func (b *Backend) List() ([]*types.VMInfo, error) {
	var vms []*types.VMInfo

	cmd := exec.Command(b.getVBoxManageCommand(), "list", "vms")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// VBoxManage might return error but still provide usable output
		// Check if we have VM data in the output despite the error
		outputStr := string(output)
		re := regexp.MustCompile(`"([^"]+)"\s+\{([^}]+)\}`)
		if !re.MatchString(outputStr) {
			return vms, fmt.Errorf("VBoxManage command failed: %v\nOutput: %s", err, outputStr)
		}
		// Continue parsing despite error - we have VM data
	}

	// Parse VM list
	re := regexp.MustCompile(`"([^"]+)"\s+\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(string(output), -1)

	for _, match := range matches {
		if len(match) >= 2 {
			vmName := match[1]

			// Skip inaccessible VMs - they can't be managed
			if vmName == "<inaccessible>" {
				vms = append(vms, &types.VMInfo{
					Name:        vmName,
					State:       types.VMStateNotFound,
					Backend:     "virtualbox",
					RAM:         "unknown",
					CPUs:        0,
					DiskSize:    "unknown",
					OSType:      "unknown",
					CreatedAt:   time.Time{},
					ErrorDetail: "VM files not accessible - VM may have been moved or deleted",
				})
				continue
			}

			info, err := b.GetInfoWithFallback(vmName)
			if err != nil {
				// If both primary and fallback methods fail, create basic VM info
				vms = append(vms, &types.VMInfo{
					Name:        vmName,
					State:       types.VMStateError,
					Backend:     "virtualbox",
					RAM:         "unknown",
					CPUs:        0,
					DiskSize:    "unknown",
					OSType:      "unknown",
					CreatedAt:   time.Time{}, // Use zero time for error state
					ErrorDetail: err.Error(),
				})
				continue
			}
			vms = append(vms, info)
		}
	}

	return vms, nil
}

// GetInfo gets VM information
func (b *Backend) GetInfo(vmName string) (*types.VMInfo, error) {
	return b.GetInfoWithFallback(vmName)
}

// GetState gets VM state
func (b *Backend) GetState(vmName string) types.VMState {
	cmd := exec.Command(b.getVBoxManageCommand(), "showvminfo", vmName, "--machinereadable")
	output, err := cmd.Output()
	if err != nil {
		return types.VMStateNotFound
	}

	// Parse state from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VMState=") {
			state := strings.TrimPrefix(line, "VMState=")
			state = strings.Trim(state, "\"")
			return b.convertVBoxState(state)
		}
	}

	return types.VMStateNotFound
}

// GetIP gets VM IP address
func (b *Backend) GetIP(vmName string) (string, error) {
	cmd := exec.Command(b.getVBoxManageCommand(), "guestproperty", "get", vmName, "/VirtualBox/GuestInfo/Net/0/V4/IP")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse IP from output
	if strings.Contains(string(output), "No value set") {
		return "", fmt.Errorf("IP not available")
	}

	re := regexp.MustCompile(`Value: (.+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1]), nil
	}

	return "", fmt.Errorf("IP not found")
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
	args := []string{"snapshot", vmName, "take", snapshotName}
	if description != "" {
		args = append(args, "--description", description)
	}
	return b.vboxManage(args...)
}

func (b *Backend) ListSnapshots(vmName string) ([]*types.SnapshotInfo, error) {
	cmd := exec.Command(b.getVBoxManageCommand(), "snapshot", vmName, "list", "--machinereadable")
	output, err := cmd.CombinedOutput()

	outputStr := string(output)

	// Check if error is due to no snapshots vs VM not found
	if strings.Contains(outputStr, "does not have any snapshots") {
		// VM exists but has no snapshots - return empty slice
		return []*types.SnapshotInfo{}, nil
	}
	if strings.Contains(outputStr, "Could not find a registered machine") {
		return nil, fmt.Errorf("VM not found: %s", vmName)
	}

	// If we have snapshot data in output, process it (even if err != nil)
	// VBoxManage sometimes returns non-zero exit code even on success
	if strings.Contains(outputStr, "SnapshotName=") {
		return b.parseSnapshots(outputStr, vmName), nil
	}

	// If we got an error and no recognizable output, return the error
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %v", err)
	}

	return b.parseSnapshots(outputStr, vmName), nil
}

func (b *Backend) RevertSnapshot(vmName, snapshotName string) error {
	return b.vboxManage("snapshot", vmName, "restore", snapshotName)
}

func (b *Backend) DeleteSnapshot(vmName, snapshotName string) error {
	return b.vboxManage("snapshot", vmName, "delete", snapshotName)
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
func (b *Backend) vboxManage(args ...string) error {
	vboxCmd := b.getVBoxManageCommand()
	cmd := exec.Command(vboxCmd, args...)
	return cmd.Run()
}

// getVBoxManageCommand returns the VBoxManage command path
func (b *Backend) getVBoxManageCommand() string {
	if b.vboxManagePath != "" {
		return b.vboxManagePath
	}
	return "VBoxManage" // fallback to PATH
}

func (b *Backend) parseRAMToMB(ramStr string) (int, error) {
	ramStr = strings.ToUpper(strings.TrimSpace(ramStr))

	// Parse number and unit
	re := regexp.MustCompile(`^(\d+)\s*([KMGT]?)B?$`)
	matches := re.FindStringSubmatch(ramStr)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid RAM format: %s", ramStr)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	unit := ""
	if len(matches) > 2 {
		unit = matches[2]
	}

	switch unit {
	case "", "M":
		return value, nil
	case "G":
		return value * 1024, nil
	case "T":
		return value * 1024 * 1024, nil
	case "K":
		return value / 1024, nil
	default:
		return 0, fmt.Errorf("unsupported RAM unit: %s", unit)
	}
}

func (b *Backend) parseDiskToMB(diskStr string) (int, error) {
	diskStr = strings.ToUpper(strings.TrimSpace(diskStr))

	// Parse number and unit
	re := regexp.MustCompile(`^(\d+)\s*([KMGT]?)B?$`)
	matches := re.FindStringSubmatch(diskStr)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid disk format: %s", diskStr)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	unit := ""
	if len(matches) > 2 {
		unit = matches[2]
	}

	switch unit {
	case "", "M":
		return value, nil
	case "G":
		return value * 1024, nil
	case "T":
		return value * 1024 * 1024, nil
	case "K":
		return value / 1024, nil
	default:
		return 0, fmt.Errorf("unsupported disk unit: %s", unit)
	}
}

func (b *Backend) getOSType(osType string) string {
	osTypeMap := map[string]string{
		"ubuntu":    "Ubuntu_64",
		"debian":    "Debian_64",
		"centos":    "RedHat_64",
		"fedora":    "Fedora_64",
		"windows10": "Windows10_64",
		"windows11": "Windows11_64",
		"linux":     "Linux_64",
	}

	if vboxType, exists := osTypeMap[osType]; exists {
		return vboxType
	}

	return "Other_64"
}

func (b *Backend) convertVBoxState(vboxState string) types.VMState {
	switch vboxState {
	case "running":
		return types.VMStateRunning
	case "poweroff", "aborted":
		return types.VMStateStopped
	case "saved", "suspended":
		return types.VMStateSuspended
	case "starting":
		return types.VMStateStarting
	case "stopping":
		return types.VMStateStopping
	default:
		return types.VMStateError
	}
}

func (b *Backend) parseVMInfo(output, vmName string) *types.VMInfo {
	info := &types.VMInfo{
		Name:    vmName,
		Backend: "virtualbox",
	}

	var diskPath string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Remove any carriage return characters
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "VMState=") {
			state := strings.TrimPrefix(line, "VMState=")
			state = strings.Trim(state, "\"")
			info.State = b.convertVBoxState(state)
		} else if strings.HasPrefix(line, "memory=") {
			ramMB := strings.TrimPrefix(line, "memory=")
			ramMB = strings.Trim(ramMB, "\"")
			if mb, err := strconv.Atoi(ramMB); err == nil {
				if mb >= 1024 {
					info.RAM = fmt.Sprintf("%dG", mb/1024)
				} else {
					info.RAM = fmt.Sprintf("%dM", mb)
				}
			}
		} else if strings.HasPrefix(line, "cpus=") {
			cpusStr := strings.TrimPrefix(line, "cpus=")
			cpusStr = strings.Trim(cpusStr, "\"")
			if cpus, err := strconv.Atoi(cpusStr); err == nil {
				info.CPUs = cpus
			}
		} else if strings.HasPrefix(line, "ostype=") {
			ostype := strings.TrimPrefix(line, "ostype=")
			ostype = strings.Trim(ostype, "\"")
			info.OSType = ostype
		} else if (strings.Contains(line, "-0-0") || strings.Contains(line, "-1-0")) &&
			   (strings.Contains(line, ".vdi") || strings.Contains(line, ".vmdk") || strings.Contains(line, ".vhd")) &&
			   !strings.Contains(line, "ImageUUID") && !strings.Contains(line, "IsEjected") {
			// Extract disk path from any controller-0-0="path/to/disk.vdi"
			if idx := strings.Index(line, "="); idx != -1 {
				path := strings.Trim(line[idx+1:], "\"")
				if path != "none" && path != "emptydrive" && diskPath == "" {
					diskPath = path
				}
			}
		}
	}

	// If we didn't find a state, set it to unknown
	if info.State == "" {
		info.State = types.VMStateUnknown
	}

	// Get disk size if we found a disk path
	if diskPath != "" {
		info.DiskSize = b.getDiskSize(diskPath)
	} else {
		info.DiskSize = "unknown"
	}

	return info
}

// getDiskSize gets disk size from VBoxManage showmediuminfo
func (b *Backend) getDiskSize(diskPath string) string {
	cmd := exec.Command(b.getVBoxManageCommand(), "showmediuminfo", diskPath)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Parse output to find Capacity line
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Capacity:") {
			// Extract size from "Capacity:       51200 MBytes"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				sizeStr := parts[1]
				if size, err := strconv.Atoi(sizeStr); err == nil {
					// Convert MB to GB if >= 1024 MB
					if size >= 1024 {
						return fmt.Sprintf("%dG", size/1024)
					}
					return fmt.Sprintf("%dM", size)
				}
			}
		}
	}

	return "unknown"
}

func (b *Backend) parseSnapshots(output, vmName string) []*types.SnapshotInfo {
	var snapshots []*types.SnapshotInfo

	// Parse snapshots using regex to match all variants (SnapshotName, SnapshotName-1, SnapshotName-1-1, etc.)
	// VirtualBox uses suffixes for nested snapshots
	snapshotNamePattern := regexp.MustCompile(`^SnapshotName(-[\d-]+)?="([^"]+)"$`)
	snapshotUUIDPattern := regexp.MustCompile(`^SnapshotUUID(-[\d-]+)?="([^"]+)"$`)
	snapshotDescPattern := regexp.MustCompile(`^SnapshotDescription(-[\d-]+)?="([^"]*)"$`)

	// Create map to group snapshot data by suffix
	snapshotMap := make(map[string]*types.SnapshotInfo)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Match snapshot name (with any suffix)
		if matches := snapshotNamePattern.FindStringSubmatch(line); len(matches) >= 3 {
			suffix := matches[1] // e.g., "", "-1", "-1-1"
			name := matches[2]
			if name != "" {
				snapshotMap[suffix] = &types.SnapshotInfo{
					Name: name,
					VM:   vmName,
				}
			}
		} else if matches := snapshotDescPattern.FindStringSubmatch(line); len(matches) >= 3 {
			// Match snapshot description
			suffix := matches[1]
			description := matches[2]
			if snapshot, exists := snapshotMap[suffix]; exists {
				snapshot.Description = description
			}
		} else if matches := snapshotUUIDPattern.FindStringSubmatch(line); len(matches) >= 3 {
			// Match snapshot UUID
			suffix := matches[1]
			uuid := matches[2]
			if snapshot, exists := snapshotMap[suffix]; exists {
				b.enrichSnapshotInfo(snapshot, vmName, uuid)
			}
		}
	}

	// Convert map to slice
	for _, snapshot := range snapshotMap {
		snapshots = append(snapshots, snapshot)
	}

	return snapshots
}

// enrichSnapshotInfo enriches snapshot info with size and timestamp
func (b *Backend) enrichSnapshotInfo(snapshot *types.SnapshotInfo, vmName, snapshotUUID string) {
	// Get VM info to find snapshot folder
	cmd := exec.Command(b.getVBoxManageCommand(), "showvminfo", vmName, "--machinereadable")
	output, err := cmd.Output()
	if err != nil {
		// Set default values on error
		snapshot.CreatedAt = time.Time{}
		snapshot.Size = 0
		return
	}

	var snapshotDisk string

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Snapshots/") && strings.Contains(line, ".vdi") {
			// Find snapshot disk path (like "SATA-0-0"="/path/to/snapshot.vdi")
			if idx := strings.Index(line, "="); idx != -1 {
				path := strings.Trim(line[idx+1:], "\"")
				if path != "none" && strings.Contains(path, "Snapshots/") {
					snapshotDisk = path
					break
				}
			}
		}
	}

	// Try to find snapshot creation time from disk file
	if snapshotDisk != "" {
		if stat, err := os.Stat(snapshotDisk); err == nil {
			snapshot.CreatedAt = stat.ModTime()
		}

		// Get disk size using VBoxManage showmediuminfo
		cmd := exec.Command(b.getVBoxManageCommand(), "showmediuminfo", snapshotDisk)
		if output, err := cmd.Output(); err == nil {
			snapshot.Size = b.parseSnapshotSize(string(output))
		}
	}

	// Fallback: If we didn't get timestamp from file, use zero time
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = time.Time{}
	}
}

// parseSnapshotSize parses snapshot size from VBoxManage showmediuminfo output
func (b *Backend) parseSnapshotSize(output string) int64 {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Size on disk:") {
			// Extract size from "Size on disk:   2 MBytes"
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				sizeStr := parts[3] // "2"
				unit := parts[4]    // "MBytes"

				if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
					switch strings.ToLower(unit) {
					case "bytes", "byte":
						return size
					case "kbytes", "kbyte", "kb":
						return size * 1024
					case "mbytes", "mbyte", "mb":
						return size * 1024 * 1024
					case "gbytes", "gbyte", "gb":
						return size * 1024 * 1024 * 1024
					default:
						return size // assume bytes
					}
				}
			}
		}
	}
	return 0
}

// VBoxVMConfig represents VirtualBox VM configuration
type VBoxVMConfig struct {
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

// detectVirtualBox performs comprehensive VirtualBox detection
func (b *Backend) detectVirtualBox() string {
	// Method 1: PATH environment check
	if path, err := exec.LookPath("VBoxManage"); err == nil {
		return path
	}

	// Method 2: Platform-specific detection using system info
	sysInfo, err := system.GetSystemInfo()
	if err != nil {
		// Fallback to runtime.GOOS
		switch runtime.GOOS {
		case "windows":
			return b.detectVirtualBoxWindows()
		case "linux":
			return b.detectVirtualBoxLinux()
		case "darwin":
			return b.detectVirtualBoxMacOS()
		default:
			return ""
		}
	}

	if system.CheckCondition(sysInfo, "windows") {
		return b.detectVirtualBoxWindows()
	} else if system.CheckCondition(sysInfo, "linux") {
		return b.detectVirtualBoxLinux()
	} else if system.CheckCondition(sysInfo, "darwin") {
		return b.detectVirtualBoxMacOS()
	}

	return ""
}

// detectVirtualBoxWindows implements Windows-specific VirtualBox detection
func (b *Backend) detectVirtualBoxWindows() string {
	// Method 1: Registry detection
	if path := b.checkVirtualBoxRegistry(); path != "" {
		return path
	}

	// Method 2: Common installation paths
	commonPaths := []string{
		"C:\\Program Files\\Oracle\\VirtualBox\\VBoxManage.exe",
		"C:\\Program Files (x86)\\Oracle\\VirtualBox\\VBoxManage.exe",
		"C:\\VirtualBox\\VBoxManage.exe",
		"D:\\Program Files\\Oracle\\VirtualBox\\VBoxManage.exe",
		"D:\\Program Files (x86)\\Oracle\\VirtualBox\\VBoxManage.exe",
	}

	for _, path := range commonPaths {
		if b.fileExists(path) && b.isExecutable(path) {
			return path
		}
	}

	return ""
}

// detectVirtualBoxLinux implements Linux-specific VirtualBox detection
func (b *Backend) detectVirtualBoxLinux() string {
	// Check common installation paths on Linux
	commonPaths := []string{
		"/usr/bin/VBoxManage",
		"/usr/local/bin/VBoxManage",
		"/opt/VirtualBox/VBoxManage",
		"/snap/bin/VBoxManage",
	}

	for _, path := range commonPaths {
		if b.fileExists(path) && b.isExecutable(path) {
			return path
		}
	}

	return ""
}

// detectVirtualBoxMacOS implements macOS-specific VirtualBox detection
func (b *Backend) detectVirtualBoxMacOS() string {
	// Check common installation paths on macOS
	commonPaths := []string{
		"/usr/local/bin/VBoxManage",
		"/Applications/VirtualBox.app/Contents/MacOS/VBoxManage",
		"/opt/homebrew/bin/VBoxManage",
	}

	for _, path := range commonPaths {
		if b.fileExists(path) && b.isExecutable(path) {
			return path
		}
	}

	return ""
}

// checkVirtualBoxRegistry checks Windows registry for VirtualBox installation
func (b *Backend) checkVirtualBoxRegistry() string {
	if runtime.GOOS != "windows" {
		return ""
	}

	// Registry keys to check
	regKeys := []string{
		"HKEY_LOCAL_MACHINE\\SOFTWARE\\Oracle\\VirtualBox",
		"HKEY_LOCAL_MACHINE\\SOFTWARE\\WOW6432Node\\Oracle\\VirtualBox",
		"HKEY_CURRENT_USER\\SOFTWARE\\Oracle\\VirtualBox",
	}

	for _, regKey := range regKeys {
		if path := b.queryRegistry(regKey, "InstallDir"); path != "" {
			vboxManagePath := filepath.Join(path, "VBoxManage.exe")
			if b.fileExists(vboxManagePath) {
				return vboxManagePath
			}
		}
	}

	return ""
}

// queryRegistry queries Windows registry for a value
func (b *Backend) queryRegistry(keyPath, valueName string) string {
	if runtime.GOOS != "windows" {
		return ""
	}

	cmd := exec.Command("reg", "query", keyPath, "/v", valueName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse registry output
	// Format: "    InstallDir    REG_SZ    C:\Program Files\Oracle\VirtualBox\"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, valueName) && strings.Contains(line, "REG_SZ") {
			// Split on whitespace and find the value after REG_SZ
			parts := strings.Fields(line)
			regSzIndex := -1
			for i, part := range parts {
				if part == "REG_SZ" {
					regSzIndex = i
					break
				}
			}
			if regSzIndex >= 0 && regSzIndex+1 < len(parts) {
				// Join all parts after REG_SZ (in case path contains spaces)
				value := strings.Join(parts[regSzIndex+1:], " ")
				// Remove trailing backslash and quotes if present
				value = strings.TrimRight(value, "\\")
				value = strings.Trim(value, "\"")
				return value
			}
		}
	}

	return ""
}

// fileExists checks if a file exists
func (b *Backend) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isExecutable checks if a file is executable
func (b *Backend) isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// On Windows, check file extension
	if runtime.GOOS == "windows" {
		return strings.HasSuffix(strings.ToLower(path), ".exe")
	}

	// On Unix-like systems, check execute permission
	return info.Mode()&0111 != 0
}

// GetDiagnosticInfo returns detailed diagnostic information for troubleshooting
func (b *Backend) GetDiagnosticInfo() *VirtualBoxDiagnosticInfo {
	diag := &VirtualBoxDiagnosticInfo{
		Platform: runtime.GOOS,
	}

	// Check PATH environment
	if path, err := exec.LookPath("VBoxManage"); err == nil {
		diag.PathEnvironment = path
	}

	// Platform-specific checks
	switch runtime.GOOS {
	case "windows":
		diag.RegistryKeys = b.checkAllRegistryKeys()
		diag.InstallationPaths = b.checkAllInstallationPaths()
		diag.RunningServices = b.checkVBoxServices()
	case "linux":
		diag.InstallationPaths = b.checkAllInstallationPaths()
		diag.PackageManager = b.checkPackageManager()
	case "darwin":
		diag.InstallationPaths = b.checkAllInstallationPaths()
		diag.BrewInstalled = b.checkBrewInstallation()
	}

	// Generate suggestions
	diag.Suggestions = b.generateSuggestions(diag)

	return diag
}

// VirtualBoxDiagnosticInfo contains diagnostic information
type VirtualBoxDiagnosticInfo struct {
	Platform          string            `json:"platform"`
	PathEnvironment   string            `json:"path_environment,omitempty"`
	RegistryKeys      map[string]string `json:"registry_keys,omitempty"`
	InstallationPaths []string          `json:"installation_paths,omitempty"`
	RunningServices   []string          `json:"running_services,omitempty"`
	PackageManager    string            `json:"package_manager,omitempty"`
	BrewInstalled     bool              `json:"brew_installed,omitempty"`
	Suggestions       []string          `json:"suggestions"`
}

// checkAllRegistryKeys checks all registry locations
func (b *Backend) checkAllRegistryKeys() map[string]string {
	if runtime.GOOS != "windows" {
		return nil
	}

	results := make(map[string]string)
	regKeys := []string{
		"HKEY_LOCAL_MACHINE\\SOFTWARE\\Oracle\\VirtualBox",
		"HKEY_LOCAL_MACHINE\\SOFTWARE\\WOW6432Node\\Oracle\\VirtualBox",
		"HKEY_CURRENT_USER\\SOFTWARE\\Oracle\\VirtualBox",
	}

	for _, regKey := range regKeys {
		if value := b.queryRegistry(regKey, "InstallDir"); value != "" {
			results[regKey] = value
		} else {
			results[regKey] = "Not found"
		}
	}

	return results
}

// checkAllInstallationPaths checks all common installation paths
func (b *Backend) checkAllInstallationPaths() []string {
	var foundPaths []string
	var commonPaths []string

	switch runtime.GOOS {
	case "windows":
		commonPaths = []string{
			"C:\\Program Files\\Oracle\\VirtualBox\\VBoxManage.exe",
			"C:\\Program Files (x86)\\Oracle\\VirtualBox\\VBoxManage.exe",
			"C:\\VirtualBox\\VBoxManage.exe",
			"D:\\Program Files\\Oracle\\VirtualBox\\VBoxManage.exe",
			"D:\\Program Files (x86)\\Oracle\\VirtualBox\\VBoxManage.exe",
		}
	case "linux":
		commonPaths = []string{
			"/usr/bin/VBoxManage",
			"/usr/local/bin/VBoxManage",
			"/opt/VirtualBox/VBoxManage",
			"/snap/bin/VBoxManage",
		}
	case "darwin":
		commonPaths = []string{
			"/usr/local/bin/VBoxManage",
			"/Applications/VirtualBox.app/Contents/MacOS/VBoxManage",
			"/opt/homebrew/bin/VBoxManage",
		}
	}

	for _, path := range commonPaths {
		if b.fileExists(path) {
			foundPaths = append(foundPaths, path+" ✓")
		} else {
			foundPaths = append(foundPaths, path+" ✗")
		}
	}

	return foundPaths
}

// checkVBoxServices checks VirtualBox Windows services
func (b *Backend) checkVBoxServices() []string {
	if runtime.GOOS != "windows" {
		return nil
	}

	services := []string{"VBoxSDS", "VBoxSVC"}
	var results []string

	for _, service := range services {
		cmd := exec.Command("sc", "query", service)
		if err := cmd.Run(); err == nil {
			results = append(results, service+" ✓")
		} else {
			results = append(results, service+" ✗")
		}
	}

	return results
}

// checkPackageManager checks Linux package manager installation
func (b *Backend) checkPackageManager() string {
	if runtime.GOOS != "linux" {
		return ""
	}

	// Check common package managers
	managers := []struct {
		cmd  string
		args []string
		name string
	}{
		{"dpkg", []string{"-l", "virtualbox*"}, "apt/dpkg"},
		{"rpm", []string{"-qa", "VirtualBox*"}, "yum/rpm"},
		{"snap", []string{"list", "virtualbox"}, "snap"},
		{"flatpak", []string{"list", "--app", "|", "grep", "virtualbox"}, "flatpak"},
	}

	for _, mgr := range managers {
		if _, err := exec.LookPath(mgr.cmd); err == nil {
			cmd := exec.Command(mgr.cmd, mgr.args...)
			if output, err := cmd.Output(); err == nil && len(output) > 0 {
				return mgr.name + " ✓"
			}
		}
	}

	return "Not found via package managers"
}

// checkBrewInstallation checks macOS Homebrew installation
func (b *Backend) checkBrewInstallation() bool {
	if runtime.GOOS != "darwin" {
		return false
	}

	cmd := exec.Command("brew", "list", "virtualbox")
	return cmd.Run() == nil
}

// generateSuggestions generates troubleshooting suggestions
func (b *Backend) generateSuggestions(diag *VirtualBoxDiagnosticInfo) []string {
	var suggestions []string

	// Universal suggestions
	suggestions = append(suggestions, "1. Verify VirtualBox is installed: Try running 'VBoxManage --version' manually")

	// Platform-specific suggestions
	switch diag.Platform {
	case "windows":
		if diag.PathEnvironment == "" {
			suggestions = append(suggestions, "2. Add VirtualBox to PATH: Add installation directory to system PATH environment variable")
		}
		if len(diag.RegistryKeys) == 0 || allRegistryKeysNotFound(diag.RegistryKeys) {
			suggestions = append(suggestions, "3. Reinstall VirtualBox: Download and reinstall from https://www.virtualbox.org/")
		}
		if len(diag.RunningServices) == 0 {
			suggestions = append(suggestions, "4. Start VirtualBox services: Run 'services.msc' and start VirtualBox services")
		}
		suggestions = append(suggestions, "5. Run as Administrator: Try running portunix with administrator privileges")

	case "linux":
		if diag.PackageManager == "" || strings.Contains(diag.PackageManager, "Not found") {
			suggestions = append(suggestions, "2. Install via package manager:")
			suggestions = append(suggestions, "   - Ubuntu/Debian: sudo apt install virtualbox")
			suggestions = append(suggestions, "   - CentOS/RHEL: sudo yum install VirtualBox")
			suggestions = append(suggestions, "   - Snap: sudo snap install virtualbox")
		}
		suggestions = append(suggestions, "3. Add user to vboxusers group: sudo usermod -a -G vboxusers $USER")

	case "darwin":
		if !diag.BrewInstalled {
			suggestions = append(suggestions, "2. Install via Homebrew: brew install --cask virtualbox")
		}
		suggestions = append(suggestions, "3. Check Security & Privacy settings: Allow Oracle kernel extensions")
	}

	suggestions = append(suggestions, "Manual installation: portunix install virtualbox")

	return suggestions
}

// allRegistryKeysNotFound checks if all registry keys returned "Not found"
func allRegistryKeysNotFound(regKeys map[string]string) bool {
	for _, value := range regKeys {
		if value != "Not found" {
			return false
		}
	}
	return true
}