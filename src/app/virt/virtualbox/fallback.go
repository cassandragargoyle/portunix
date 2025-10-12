package virtualbox

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"portunix.ai/app/virt/types"
)

// VBoxMachine represents the root element of a .vbox file
type VBoxMachine struct {
	XMLName xml.Name `xml:"VirtualBox"`
	Machine Machine  `xml:"Machine"`
}

// Machine represents the Machine element in .vbox file
type Machine struct {
	Name     string   `xml:"name,attr"`
	UUID     string   `xml:"uuid,attr"`
	OSType   string   `xml:"OSType,attr"`
	Hardware Hardware `xml:"Hardware"`
}

// Hardware represents hardware configuration
type Hardware struct {
	Memory Memory `xml:"Memory"`
	CPU    CPU    `xml:"CPU"`
}

// Memory represents memory configuration
type Memory struct {
	RAMSize int `xml:"RAMSize,attr"`
}

// CPU represents CPU configuration
type CPU struct {
	Count int `xml:"count,attr"`
}

// getVMsDirectory returns the VirtualBox VMs directory path
func (b *Backend) getVMsDirectory() string {
	if runtime.GOOS == "windows" {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "VirtualBox VMs")
	}
	// Linux/Mac
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "VirtualBox VMs")
}

// parseVBoxFile parses a .vbox XML file to extract VM configuration
func (b *Backend) parseVBoxFile(vmPath string) (*types.VMInfo, error) {
	vboxFile := filepath.Join(vmPath, filepath.Base(vmPath)+".vbox")

	// Check if .vbox file exists
	if _, err := os.Stat(vboxFile); os.IsNotExist(err) {
		// Try alternative naming pattern
		entries, err := os.ReadDir(vmPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read VM directory: %w", err)
		}

		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".vbox") {
				vboxFile = filepath.Join(vmPath, entry.Name())
				break
			}
		}
	}

	// Read and parse XML file
	data, err := os.ReadFile(vboxFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read .vbox file: %w", err)
	}

	var vboxMachine VBoxMachine
	if err := xml.Unmarshal(data, &vboxMachine); err != nil {
		return nil, fmt.Errorf("failed to parse .vbox XML: %w", err)
	}

	// Convert to VMInfo
	info := &types.VMInfo{
		Name:     vboxMachine.Machine.Name,
		Backend:  "virtualbox",
		OSType:   vboxMachine.Machine.OSType,
		CPUs:     vboxMachine.Machine.Hardware.CPU.Count,
		State:    types.VMStateUnknown, // Will be determined separately
	}

	// Convert RAM from MB to human-readable format
	ramMB := vboxMachine.Machine.Hardware.Memory.RAMSize
	if ramMB >= 1024 {
		info.RAM = fmt.Sprintf("%dG", ramMB/1024)
	} else {
		info.RAM = fmt.Sprintf("%dM", ramMB)
	}

	// Try to get disk size from VM directory
	info.DiskSize = b.getVMDiskSize(vmPath)

	return info, nil
}

// getVMDiskSize tries to determine VM disk size from .vdi files
func (b *Backend) getVMDiskSize(vmPath string) string {
	entries, err := os.ReadDir(vmPath)
	if err != nil {
		return "unknown"
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".vdi") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			// Get file size in bytes
			sizeBytes := info.Size()
			// Convert to GB
			sizeGB := sizeBytes / (1024 * 1024 * 1024)
			if sizeGB > 0 {
				return fmt.Sprintf("%dG", sizeGB)
			}
			// If less than 1GB, show in MB
			sizeMB := sizeBytes / (1024 * 1024)
			return fmt.Sprintf("%dM", sizeMB)
		}
	}

	return "unknown"
}

// getRunningVMs returns a list of running VM names
func (b *Backend) getRunningVMs() ([]string, error) {
	cmd := exec.Command(b.getVBoxManageCommand(), "list", "runningvms")
	output, err := cmd.Output()
	if err != nil {
		// Even if command fails, we might have partial output
		if output == nil || len(output) == 0 {
			return []string{}, nil
		}
	}

	var runningVMs []string
	re := regexp.MustCompile(`"([^"]+)"\s+\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(string(output), -1)

	for _, match := range matches {
		if len(match) >= 2 {
			runningVMs = append(runningVMs, match[1])
		}
	}

	return runningVMs, nil
}

// isVMRunning checks if a VM is in the running list
func (b *Backend) isVMRunning(vmName string, runningVMs []string) bool {
	for _, running := range runningVMs {
		if running == vmName {
			return true
		}
	}
	return false
}

// GetInfoWithFallback attempts to get VM info with fallback methods
func (b *Backend) GetInfoWithFallback(vmName string) (*types.VMInfo, error) {
	// First try the standard method
	cmd := exec.Command(b.getVBoxManageCommand(), "showvminfo", vmName, "--machinereadable")
	output, err := cmd.CombinedOutput()

	// If no error, proceed with normal parsing
	if err == nil {
		info := b.parseVMInfo(string(output), vmName)
		return info, nil
	}

	// Check if we got E_ACCESSDENIED error
	if strings.Contains(string(output), "E_ACCESSDENIED") {
		// Try fallback method - parse .vbox file
		vmsDir := b.getVMsDirectory()
		vmPath := filepath.Join(vmsDir, vmName)

		info, parseErr := b.parseVBoxFile(vmPath)
		if parseErr != nil {
			// If parsing also fails, return basic info with error state
			return &types.VMInfo{
				Name:        vmName,
				State:       types.VMStateError,
				Backend:     "virtualbox",
				RAM:         "unknown",
				CPUs:        0,
				DiskSize:    "unknown",
				OSType:      "unknown",
				ErrorDetail: "Access denied - try running as administrator",
			}, nil
		}

		// We got info from .vbox file, now determine state
		runningVMs, _ := b.getRunningVMs()
		if b.isVMRunning(vmName, runningVMs) {
			info.State = types.VMStateRunning
		} else {
			info.State = types.VMStateStopped
		}

		return info, nil
	}

	// If no E_ACCESSDENIED error, proceed with normal parsing
	if err != nil {
		return nil, fmt.Errorf("VBoxManage showvminfo failed for VM '%s': %v\nOutput: %s", vmName, err, string(output))
	}

	info := b.parseVMInfo(string(output), vmName)
	return info, nil
}

// detectAccessDeniedError checks if error output contains E_ACCESSDENIED
func (b *Backend) detectAccessDeniedError(output string) bool {
	return strings.Contains(output, "E_ACCESSDENIED") ||
	       strings.Contains(output, "0x80070005") ||
	       strings.Contains(output, "The object functionality is limited")
}