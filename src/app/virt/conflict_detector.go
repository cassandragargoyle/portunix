package virt

import (
	"fmt"
	"os/exec"
	"strings"
)

// ConflictType represents the type of virtualization conflict
type ConflictType string

const (
	ConflictNone         ConflictType = "none"
	ConflictVBoxKVM      ConflictType = "virtualbox-kvm"
	ConflictHyperVVBox   ConflictType = "hyperv-virtualbox"
	ConflictHyperVKVM    ConflictType = "hyperv-kvm"
)

// VirtualizationConflict represents a detected virtualization conflict
type VirtualizationConflict struct {
	Type              ConflictType
	Conflict          bool
	VirtualBoxPresent bool
	KVMActive         bool
	HyperVActive      bool
	LoadedModules     []string
	Recommendation    string
	Details           string
}

// DetectVirtualizationConflict detects conflicts between virtualization technologies
func DetectVirtualizationConflict() (*VirtualizationConflict, error) {
	conflict := &VirtualizationConflict{
		Type:          ConflictNone,
		Conflict:      false,
		LoadedModules: []string{},
	}

	// Check VirtualBox installation
	conflict.VirtualBoxPresent = isVirtualBoxInstalled()

	// Check KVM modules
	kvmModules, kvmActive := getKVMModules()
	conflict.KVMActive = kvmActive
	conflict.LoadedModules = kvmModules

	// Check Hyper-V (Windows only)
	conflict.HyperVActive = isHyperVActive()

	// Determine conflict type
	if conflict.VirtualBoxPresent && conflict.KVMActive {
		conflict.Conflict = true
		conflict.Type = ConflictVBoxKVM
		conflict.Details = "VirtualBox and KVM both require exclusive access to hardware virtualization (AMD-V/Intel VT-x)"
		conflict.Recommendation = "Unload KVM modules to use VirtualBox, or switch to KVM/QEMU"
	} else if conflict.VirtualBoxPresent && conflict.HyperVActive {
		conflict.Conflict = true
		conflict.Type = ConflictHyperVVBox
		conflict.Details = "VirtualBox and Hyper-V cannot run simultaneously on Windows"
		conflict.Recommendation = "Disable Hyper-V to use VirtualBox"
	} else if conflict.KVMActive && conflict.HyperVActive {
		conflict.Conflict = true
		conflict.Type = ConflictHyperVKVM
		conflict.Details = "Hyper-V and KVM conflict detected (WSL2 scenario)"
		conflict.Recommendation = "Disable Hyper-V or use WSL2 with Hyper-V backend"
	}

	return conflict, nil
}

// getKVMModules returns loaded KVM modules and whether KVM is active
func getKVMModules() ([]string, bool) {
	cmd := exec.Command("lsmod")
	output, err := cmd.Output()
	if err != nil {
		return nil, false
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

	return modules, len(modules) > 0
}

// isHyperVActive checks if Hyper-V is active (Windows only)
func isHyperVActive() bool {
	// Check via PowerShell on Windows
	cmd := exec.Command("powershell", "-Command",
		"Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V-All | Select-Object -ExpandProperty State")

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == "Enabled"
}

// UnloadKVMModules unloads KVM kernel modules
func UnloadKVMModules() error {
	// Unload in reverse order (dependent modules first)
	modules := []string{"kvm_intel", "kvm_amd", "kvm"}

	for _, module := range modules {
		cmd := exec.Command("sudo", "rmmod", module)
		// Ignore errors - module might not be loaded
		cmd.Run()
	}

	// Verify KVM is unloaded
	_, active := getKVMModules()
	if active {
		return fmt.Errorf("failed to unload KVM modules")
	}

	return nil
}

// BlacklistKVMModules creates blacklist configuration to prevent KVM from loading
func BlacklistKVMModules() error {
	blacklistContent := `# Blacklist KVM modules (VirtualBox compatibility)
# Created by Portunix virt check --fix
blacklist kvm
blacklist kvm_amd
blacklist kvm_intel
`

	// Write blacklist file
	cmd := exec.Command("sudo", "tee", "/etc/modprobe.d/blacklist-kvm.conf")
	cmd.Stdin = strings.NewReader(blacklistContent)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create blacklist file: %w", err)
	}

	// Update initramfs
	updateCmd := exec.Command("sudo", "update-initramfs", "-u")
	if err := updateCmd.Run(); err != nil {
		// Try dracut for Fedora/RHEL
		dracutCmd := exec.Command("sudo", "dracut", "--force")
		if err := dracutCmd.Run(); err != nil {
			return fmt.Errorf("failed to update initramfs: %w", err)
		}
	}

	return nil
}

// SwitchToKVM unloads VirtualBox modules and loads KVM
func SwitchToKVM() error {
	// Unload VirtualBox modules
	vboxModules := []string{"vboxnetadp", "vboxnetflt", "vboxpci", "vboxdrv"}

	for _, module := range vboxModules {
		cmd := exec.Command("sudo", "rmmod", module)
		// Ignore errors
		cmd.Run()
	}

	// Load KVM modules
	cpuInfo, err := exec.Command("cat", "/proc/cpuinfo").Output()
	if err != nil {
		return fmt.Errorf("failed to read CPU info: %w", err)
	}

	var kvmModule string
	if strings.Contains(string(cpuInfo), "vmx") {
		kvmModule = "kvm_intel"
	} else if strings.Contains(string(cpuInfo), "svm") {
		kvmModule = "kvm_amd"
	} else {
		return fmt.Errorf("CPU does not support hardware virtualization")
	}

	// Load KVM
	if err := exec.Command("sudo", "modprobe", "kvm").Run(); err != nil {
		return fmt.Errorf("failed to load kvm module: %w", err)
	}

	if err := exec.Command("sudo", "modprobe", kvmModule).Run(); err != nil {
		return fmt.Errorf("failed to load %s module: %w", kvmModule, err)
	}

	return nil
}
