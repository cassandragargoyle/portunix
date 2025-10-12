package virt

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"portunix.ai/app/virt/qemu"
	"portunix.ai/app/virt/virtualbox"
)

// selectBackend selects the appropriate virtualization backend
func selectBackend(config *Config) (Backend, error) {
	preferred := config.GetPreferredBackend()

	switch preferred {
	case "qemu":
		backend := qemu.NewBackend()
		if backend.IsAvailable() {
			return backend, nil
		}
		// Fallback to VirtualBox on Linux if configured as auto
		if config.VirtualizationBackend == "auto" && runtime.GOOS == "linux" {
			vboxBackend := virtualbox.NewBackend()
			if vboxBackend.IsAvailable() {
				return vboxBackend, nil
			}
		}
		return nil, fmt.Errorf("QEMU/KVM is not available. Install with: portunix install qemu")

	case "virtualbox":
		backend := virtualbox.NewBackend()
		if backend.IsAvailable() {
			return backend, nil
		}
		// Fallback to QEMU on Windows if using WSL2
		if config.VirtualizationBackend == "auto" && runtime.GOOS == "windows" {
			if isWSL2() {
				qemuBackend := qemu.NewBackend()
				if qemuBackend.IsAvailable() {
					return qemuBackend, nil
				}
			}
		}

		// Generate enhanced error message with diagnostic information
		errorMsg := generateVirtualBoxErrorMessage(backend)
		return nil, fmt.Errorf(errorMsg)

	default:
		return nil, fmt.Errorf("unsupported virtualization backend: %s", preferred)
	}
}

// GetAvailableBackends returns list of available backends
func GetAvailableBackends() []string {
	var backends []string

	qemuBackend := qemu.NewBackend()
	if qemuBackend.IsAvailable() {
		backends = append(backends, "qemu")
	}

	vboxBackend := virtualbox.NewBackend()
	if vboxBackend.IsAvailable() {
		backends = append(backends, "virtualbox")
	}

	return backends
}

// GetRecommendedBackend returns the recommended backend for current platform
func GetRecommendedBackend() string {
	switch runtime.GOOS {
	case "linux":
		// Check for KVM support
		if hasKVMSupport() {
			return "qemu"
		}
		return "virtualbox"
	case "windows":
		if isWSL2() {
			return "qemu" // WSL2 can run QEMU
		}
		return "virtualbox"
	case "darwin":
		return "virtualbox" // TODO: Add native macOS virtualization support
	default:
		return "qemu"
	}
}

// hasKVMSupport checks if KVM is available
func hasKVMSupport() bool {
	// Check if /dev/kvm exists
	if _, err := exec.LookPath("kvm-ok"); err == nil {
		cmd := exec.Command("kvm-ok")
		if err := cmd.Run(); err == nil {
			return true
		}
	}

	// Alternative check - look for KVM module
	cmd := exec.Command("lsmod")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check if kvm module is loaded
	return containsKVM(string(output))
}

// containsKVM checks if KVM modules are in lsmod output
func containsKVM(output string) bool {
	return containsString(output, "kvm_intel") || containsString(output, "kvm_amd") || containsString(output, "kvm")
}

// containsString checks if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsString(s[1:], substr))
}

// isWSL2 checks if running in WSL2
func isWSL2() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	// Check for WSL2 indicators
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return containsString(string(output), "microsoft") && containsString(string(output), "WSL2")
}

// CheckVirtualizationSupport performs comprehensive virtualization support check
func CheckVirtualizationSupport() (*VirtualizationStatus, error) {
	status := &VirtualizationStatus{
		Platform: runtime.GOOS,
	}

	// Check hardware virtualization
	status.HardwareVirtualization = checkHardwareVirtualization()

	// Check available backends
	status.AvailableBackends = GetAvailableBackends()
	status.RecommendedBackend = GetRecommendedBackend()

	// Platform-specific checks
	switch runtime.GOOS {
	case "linux":
		status.KVMSupport = hasKVMSupport()
		status.LibvirtInstalled = isLibvirtInstalled()
		status.QEMUInstalled = isQEMUInstalled()
		status.VirtualBoxInstalled = isVirtualBoxInstalled()
	case "windows":
		status.HyperVEnabled = isHyperVEnabled()
		status.WSL2Available = isWSL2()
		status.VirtualBoxInstalled = isVirtualBoxInstalled()
	}

	return status, nil
}

// VirtualizationStatus contains comprehensive virtualization status information
type VirtualizationStatus struct {
	Platform               string   `json:"platform"`
	HardwareVirtualization bool     `json:"hardware_virtualization"`
	AvailableBackends      []string `json:"available_backends"`
	RecommendedBackend     string   `json:"recommended_backend"`

	// Linux specific
	KVMSupport        bool `json:"kvm_support,omitempty"`
	LibvirtInstalled  bool `json:"libvirt_installed,omitempty"`
	QEMUInstalled     bool `json:"qemu_installed,omitempty"`

	// Windows specific
	HyperVEnabled bool `json:"hyperv_enabled,omitempty"`
	WSL2Available bool `json:"wsl2_available,omitempty"`

	// Cross-platform
	VirtualBoxInstalled bool `json:"virtualbox_installed,omitempty"`
}

// Helper functions for checking installed software
func checkHardwareVirtualization() bool {
	switch runtime.GOOS {
	case "linux":
		// Check CPU flags for VT-x/AMD-V
		cmd := exec.Command("grep", "-E", "(vmx|svm)", "/proc/cpuinfo")
		return cmd.Run() == nil
	case "windows":
		// Check Windows virtualization features
		cmd := exec.Command("powershell", "-Command", "Get-ComputerInfo | Select-Object -ExpandProperty HyperVisorPresent")
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		return containsString(string(output), "True")
	default:
		return false
	}
}

func isLibvirtInstalled() bool {
	_, err := exec.LookPath("virsh")
	return err == nil
}

func isQEMUInstalled() bool {
	_, err := exec.LookPath("qemu-system-x86_64")
	return err == nil
}

func isVirtualBoxInstalled() bool {
	_, err := exec.LookPath("VBoxManage")
	return err == nil
}

func isHyperVEnabled() bool {
	cmd := exec.Command("powershell", "-Command", "Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V-All | Select-Object -ExpandProperty State")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return containsString(string(output), "Enabled")
}

// generateVirtualBoxErrorMessage creates enhanced error message with diagnostic information
func generateVirtualBoxErrorMessage(backend interface{}) string {
	// Type assertion to get VirtualBox backend
	if vboxBackend, ok := backend.(*virtualbox.Backend); ok {
		diag := vboxBackend.GetDiagnosticInfo()

		errorMsg := "VirtualBox detection failed\n\n"
		errorMsg += "Diagnostic Information:\n"

		// PATH check
		if diag.PathEnvironment != "" {
			errorMsg += "✓ PATH: VBoxManage found at " + diag.PathEnvironment + "\n"
		} else {
			errorMsg += "✗ PATH: VBoxManage not found in system PATH\n"
		}

		// Platform-specific information
		switch diag.Platform {
		case "windows":
			// Registry information
			if len(diag.RegistryKeys) > 0 {
				foundRegistry := false
				for key, value := range diag.RegistryKeys {
					if value != "Not found" {
						errorMsg += "✓ Registry: " + key + " = " + value + "\n"
						foundRegistry = true
					}
				}
				if !foundRegistry {
					errorMsg += "✗ Registry: VirtualBox registry keys not found\n"
				}
			}

			// Services information
			if len(diag.RunningServices) > 0 {
				errorMsg += "? Services: " + fmt.Sprintf("%v", diag.RunningServices) + "\n"
			}

		case "linux":
			// Package manager information
			if diag.PackageManager != "" {
				if strings.Contains(diag.PackageManager, "✓") {
					errorMsg += "✓ Package Manager: " + diag.PackageManager + "\n"
				} else {
					errorMsg += "✗ Package Manager: " + diag.PackageManager + "\n"
				}
			}

		case "darwin":
			// Homebrew information
			if diag.BrewInstalled {
				errorMsg += "✓ Homebrew: VirtualBox installed via brew\n"
			} else {
				errorMsg += "✗ Homebrew: VirtualBox not found via brew\n"
			}
		}

		// Installation paths
		if len(diag.InstallationPaths) > 0 {
			errorMsg += "\nInstallation paths checked:\n"
			for _, path := range diag.InstallationPaths {
				errorMsg += "  " + path + "\n"
			}
		}

		// Suggestions
		if len(diag.Suggestions) > 0 {
			errorMsg += "\nSuggestions:\n"
			for _, suggestion := range diag.Suggestions {
				errorMsg += suggestion + "\n"
			}
		}

		return errorMsg
	}

	// Fallback error message
	return "VirtualBox is not available. Install with: portunix install virtualbox"
}