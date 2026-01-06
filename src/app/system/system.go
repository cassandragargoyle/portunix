package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Native implementation function pointers (set by platform-specific init())
var (
	nativeGetWindowsInfo           func(*SystemInfo) error
	nativeDetectWindowsEnvironment func(*SystemInfo)
	nativeIsAdmin                  func() bool
	nativeCheckHardwareVirt        func() bool
	nativeQueryRegistry            func(keyPath, valueName string) string
	nativeIsVirtualBoxAvailable    func() bool
)

// SystemInfo contains comprehensive system information
type SystemInfo struct {
	OS           string        `json:"os"`
	Version      string        `json:"version"`
	Build        string        `json:"build"`
	Architecture string        `json:"architecture"`
	Hostname     string        `json:"hostname"`
	Variant      string        `json:"variant"`
	Environment  []string      `json:"environment"`
	WindowsInfo  *WindowsInfo  `json:"windows_info,omitempty"`
	LinuxInfo    *LinuxInfo    `json:"linux_info,omitempty"`
	Capabilities *Capabilities `json:"capabilities"`
}

// WindowsInfo contains Windows-specific information
type WindowsInfo struct {
	Edition     string `json:"edition"`
	ProductName string `json:"product_name"`
	InstallDate string `json:"install_date"`
	BuildLab    string `json:"build_lab"`
}

// LinuxInfo contains Linux-specific information
type LinuxInfo struct {
	Distribution  string `json:"distribution"`
	Codename      string `json:"codename"`
	KernelVersion string `json:"kernel_version"`
}

// Capabilities shows what the system can do
type Capabilities struct {
	PowerShell            bool                   `json:"powershell"`
	Docker                bool                   `json:"docker"`
	DockerVersion         string                 `json:"docker_version,omitempty"`
	DockerDaemonRunning   bool                   `json:"docker_daemon_running,omitempty"`
	Podman                bool                   `json:"podman"`
	PodmanVersion         string                 `json:"podman_version,omitempty"`
	PodmanSocketRunning   bool                   `json:"podman_socket_running,omitempty"`
	ContainerAvailable    bool                   `json:"container_available"`
	ComposeInfo           *ComposeInfo           `json:"compose,omitempty"`
	Admin                 bool                   `json:"admin"`
	CertificateInfo       *CertificateInfo       `json:"certificate_bundle,omitempty"`
	VirtualizationInfo    *VirtualizationInfo    `json:"virtualization,omitempty"`
}

// ComposeInfo contains container compose tool information
type ComposeInfo struct {
	Available       bool   `json:"available"`
	Type            string `json:"type,omitempty"`            // "Docker Compose", "Podman Compose", "podman-compose"
	Version         string `json:"version,omitempty"`
	DaemonReady     bool   `json:"daemon_ready"`              // true if the underlying daemon/socket is running
	WarningMessage  string `json:"warning,omitempty"`         // warning if daemon not running
}

// VirtualizationInfo contains virtualization capabilities
type VirtualizationInfo struct {
	Backend                string   `json:"backend"`
	AvailableBackends      []string `json:"available_backends"`
	RecommendedBackend     string   `json:"recommended_backend"`
	HardwareVirtualization bool     `json:"hardware_virtualization"`
	QEMU                   bool     `json:"qemu"`
	QEMUVersion            string   `json:"qemu_version,omitempty"`
	VirtualBox             bool     `json:"virtualbox"`
	VirtualBoxVersion      string   `json:"virtualbox_version,omitempty"`
	KVMSupport             bool     `json:"kvm_support,omitempty"`
	LibvirtInstalled       bool     `json:"libvirt_installed,omitempty"`
	LibvirtVersion         string   `json:"libvirt_version,omitempty"`
}

// GetSystemInfo returns comprehensive system information
func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		Architecture: runtime.GOARCH,
		Capabilities: &Capabilities{},
		Environment:  []string{},
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	info.Hostname = hostname

	// Detect OS and get specific information
	switch runtime.GOOS {
	case "windows":
		err = getWindowsInfo(info)
	case "linux":
		err = getLinuxInfo(info)
	case "darwin":
		err = getMacOSInfo(info)
	default:
		info.OS = runtime.GOOS
		info.Version = "unknown"
		info.Variant = "unknown"
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get OS-specific info: %w", err)
	}

	// Detect environment variants
	detectEnvironment(info)

	// Check capabilities
	checkCapabilities(info)

	return info, nil
}

// CheckCondition checks if a specific condition is met
func CheckCondition(info *SystemInfo, condition string) bool {
	switch condition {
	case "windows":
		return info.OS == "Windows"
	case "linux":
		return info.OS == "Linux"
	case "macos", "darwin":
		return info.OS == "macOS"
	case "sandbox":
		return contains(info.Environment, "sandbox")
	case "docker":
		return contains(info.Environment, "docker")
	case "wsl":
		return contains(info.Environment, "wsl")
	case "vm":
		return contains(info.Environment, "vm")
	case "powershell":
		return info.Capabilities.PowerShell
	case "admin":
		return info.Capabilities.Admin
	default:
		return false
	}
}

// getWindowsInfo gets Windows-specific information
func getWindowsInfo(info *SystemInfo) error {
	// Try native API first if available (set by system_windows.go init())
	if nativeGetWindowsInfo != nil {
		if err := nativeGetWindowsInfo(info); err == nil {
			return nil
		}
	}

	// Fallback to external commands
	info.OS = "Windows"
	info.WindowsInfo = &WindowsInfo{}

	// Get Windows version using wmic
	if output, err := exec.Command("wmic", "os", "get", "Caption,Version,BuildNumber", "/format:csv").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Microsoft Windows") {
				parts := strings.Split(line, ",")
				if len(parts) >= 4 {
					info.WindowsInfo.ProductName = strings.TrimSpace(parts[1])
					info.Version = strings.TrimSpace(parts[3])
					info.Build = strings.TrimSpace(parts[2])

					// Determine Windows edition from product name
					if strings.Contains(info.WindowsInfo.ProductName, "Windows 11") {
						info.Version = "11"
					} else if strings.Contains(info.WindowsInfo.ProductName, "Windows 10") {
						info.Version = "10"
					}
				}
				break
			}
		}
	}

	// Fallback to ver command
	if info.Version == "" {
		if output, err := exec.Command("cmd", "/c", "ver").Output(); err == nil {
			verStr := string(output)
			if strings.Contains(verStr, "Version 10.0") {
				// Could be Windows 10 or 11, check build number
				if output, err := exec.Command("cmd", "/c", "wmic os get BuildNumber /value").Output(); err == nil {
					buildStr := string(output)
					if strings.Contains(buildStr, "BuildNumber=") {
						buildLine := strings.Split(buildStr, "\n")
						for _, line := range buildLine {
							if strings.HasPrefix(line, "BuildNumber=") {
								buildNum := strings.TrimPrefix(line, "BuildNumber=")
								buildNum = strings.TrimSpace(buildNum)
								info.Build = buildNum

								// Windows 11 starts from build 22000
								if buildNum >= "22000" {
									info.Version = "11"
								} else {
									info.Version = "10"
								}
								break
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// getLinuxInfo gets Linux-specific information
func getLinuxInfo(info *SystemInfo) error {
	info.OS = "Linux"
	info.LinuxInfo = &LinuxInfo{}

	// Read /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "NAME=") {
				info.LinuxInfo.Distribution = strings.Trim(strings.TrimPrefix(line, "NAME="), `"`)
			} else if strings.HasPrefix(line, "VERSION_ID=") {
				info.Version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
			} else if strings.HasPrefix(line, "VERSION_CODENAME=") {
				info.LinuxInfo.Codename = strings.Trim(strings.TrimPrefix(line, "VERSION_CODENAME="), `"`)
			} else if strings.HasPrefix(line, "BUILD_ID=") {
				info.Build = strings.Trim(strings.TrimPrefix(line, "BUILD_ID="), `"`)
			}
		}
	}

	// Get kernel version
	if output, err := exec.Command("uname", "-r").Output(); err == nil {
		info.LinuxInfo.KernelVersion = strings.TrimSpace(string(output))
		// If no BUILD_ID was found, use kernel build number as Build
		if info.Build == "" {
			// Extract kernel build number (e.g., from "6.14.0-29-generic" get "29")
			kernelParts := strings.Split(info.LinuxInfo.KernelVersion, "-")
			if len(kernelParts) >= 2 {
				info.Build = kernelParts[1]
			}
		}
	}

	return nil
}

// getMacOSInfo gets macOS-specific information
func getMacOSInfo(info *SystemInfo) error {
	info.OS = "macOS"

	// Get macOS version
	if output, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
		info.Version = strings.TrimSpace(string(output))
	}

	return nil
}

// detectEnvironment detects if running in special environments
func detectEnvironment(info *SystemInfo) {
	// Check for Windows Sandbox and VM
	if info.OS == "Windows" {
		// Try native detection first if available
		if nativeDetectWindowsEnvironment != nil {
			nativeDetectWindowsEnvironment(info)
			if info.Variant != "" {
				return
			}
		}

		// Fallback: Windows Sandbox typically has WDAGUtilityAccount user
		if username := os.Getenv("USERNAME"); username == "WDAGUtilityAccount" {
			info.Environment = append(info.Environment, "sandbox")
			info.Variant = "Sandbox"
		}

		// Fallback: Check for specific sandbox registry keys or processes
		if output, err := exec.Command("tasklist", "/FI", "IMAGENAME eq CExecSvc.exe", "/FO", "CSV", "/NH").Output(); err == nil {
			if strings.Contains(string(output), "CExecSvc.exe") {
				info.Environment = append(info.Environment, "sandbox")
				if info.Variant == "" {
					info.Variant = "Sandbox"
				}
			}
		}
	}

	// Check for Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		info.Environment = append(info.Environment, "docker")
		if info.Variant == "" {
			info.Variant = "Docker"
		}
	}

	// Check for WSL
	if info.OS == "Linux" {
		if _, err := os.Stat("/proc/version"); err == nil {
			if data, err := os.ReadFile("/proc/version"); err == nil {
				if strings.Contains(strings.ToLower(string(data)), "microsoft") {
					info.Environment = append(info.Environment, "wsl")
					if info.Variant == "" {
						info.Variant = "WSL"
					}
				}
			}
		}
	}

	// Check for VM (basic heuristics)
	if info.OS == "Windows" {
		// Check for common VM indicators with specific detection
		if output, err := exec.Command("wmic", "computersystem", "get", "manufacturer,model").Output(); err == nil {
			outputStr := strings.ToLower(string(output))

			// Check for specific VM types
			if strings.Contains(outputStr, "virtualbox") {
				info.Environment = append(info.Environment, "vm")
				if info.Variant == "" {
					info.Variant = "VirtualBox VM"
				}
			} else if strings.Contains(outputStr, "vmware") {
				info.Environment = append(info.Environment, "vm")
				if info.Variant == "" {
					info.Variant = "VMware VM"
				}
			} else if strings.Contains(outputStr, "hyper-v") {
				info.Environment = append(info.Environment, "vm")
				if info.Variant == "" {
					info.Variant = "Hyper-V VM"
				}
			} else if strings.Contains(outputStr, "qemu") {
				info.Environment = append(info.Environment, "vm")
				if info.Variant == "" {
					info.Variant = "QEMU VM"
				}
			} else if strings.Contains(outputStr, "xen") {
				info.Environment = append(info.Environment, "vm")
				if info.Variant == "" {
					info.Variant = "Xen VM"
				}
			}
		}
	}

	// Set default variant if none detected
	if info.Variant == "" {
		info.Variant = "Physical"
	}
}

// checkCapabilities checks system capabilities
func checkCapabilities(info *SystemInfo) {
	// Check PowerShell availability
	if _, err := exec.LookPath("powershell"); err == nil {
		info.Capabilities.PowerShell = true
	} else if _, err := exec.LookPath("pwsh"); err == nil {
		info.Capabilities.PowerShell = true
	}

	// Check Docker availability
	if _, err := exec.LookPath("docker"); err == nil {
		info.Capabilities.Docker = true
		// Get Docker version
		if version := GetDockerVersion(); version != "" {
			info.Capabilities.DockerVersion = version
		}
		// Check if daemon is running
		info.Capabilities.DockerDaemonRunning = IsDockerDaemonRunning()
	}

	// Check Podman availability
	if _, err := exec.LookPath("podman"); err == nil {
		info.Capabilities.Podman = true
		// Get Podman version
		if version := GetPodmanVersion(); version != "" {
			info.Capabilities.PodmanVersion = version
		}
		// Check if socket is running
		info.Capabilities.PodmanSocketRunning = IsPodmanSocketRunning()
	}

	// Set Container Available flag
	info.Capabilities.ContainerAvailable = info.Capabilities.Docker || info.Capabilities.Podman

	// Detect compose tool availability
	info.Capabilities.ComposeInfo = DetectComposeInfo(
		info.Capabilities.Docker,
		info.Capabilities.DockerDaemonRunning,
		info.Capabilities.Podman,
		info.Capabilities.PodmanSocketRunning,
	)

	// Check admin privileges (platform-specific)
	if info.OS == "Windows" {
		// Try native API first
		if nativeIsAdmin != nil {
			info.Capabilities.Admin = nativeIsAdmin()
		} else {
			// Fallback: Check if running as administrator
			if output, err := exec.Command("net", "session").Output(); err == nil {
				if len(output) > 0 {
					info.Capabilities.Admin = true
				}
			}
		}
	} else {
		// Unix-like systems - check if running as root
		if os.Geteuid() == 0 {
			info.Capabilities.Admin = true
		}
	}

	// Check certificate bundle availability
	if certInfo, err := DetectCertificateBundle(); err == nil {
		info.Capabilities.CertificateInfo = &certInfo
	}

	// Check virtualization capabilities
	checkVirtualizationCapabilities(info)
}

// checkVirtualizationCapabilities checks virtualization support
func checkVirtualizationCapabilities(info *SystemInfo) {
	virtInfo := &VirtualizationInfo{}

	// Check for QEMU
	if _, err := exec.LookPath("qemu-system-x86_64"); err == nil {
		virtInfo.QEMU = true
		virtInfo.AvailableBackends = append(virtInfo.AvailableBackends, "qemu")
		// Get QEMU version
		if version := GetQEMUVersion(); version != "" {
			virtInfo.QEMUVersion = version
		}
	}

	// Check for VirtualBox using enhanced detection
	virtInfo.VirtualBox = isVirtualBoxAvailable()
	if virtInfo.VirtualBox {
		virtInfo.AvailableBackends = append(virtInfo.AvailableBackends, "virtualbox")
		// Get VirtualBox version
		if version := GetVirtualBoxVersion(); version != "" {
			virtInfo.VirtualBoxVersion = version
		}
	}

	// Check for libvirt (Linux only)
	if info.OS == "Linux" {
		if _, err := exec.LookPath("virsh"); err == nil {
			virtInfo.LibvirtInstalled = true
			// Get Libvirt version
			if version := GetLibvirtVersion(); version != "" {
				virtInfo.LibvirtVersion = version
			}
		}

		// Check for KVM support
		virtInfo.KVMSupport = checkKVMSupport()
	}

	// Check hardware virtualization
	virtInfo.HardwareVirtualization = checkHardwareVirtualization()

	// Set recommended backend
	virtInfo.RecommendedBackend = getRecommendedBackend(info.OS, virtInfo)

	// Set current backend (first available)
	if len(virtInfo.AvailableBackends) > 0 {
		virtInfo.Backend = virtInfo.RecommendedBackend
		if virtInfo.Backend == "" {
			virtInfo.Backend = virtInfo.AvailableBackends[0]
		}
	} else {
		virtInfo.Backend = "none"
	}

	info.Capabilities.VirtualizationInfo = virtInfo
}

// checkKVMSupport checks if KVM is supported
func checkKVMSupport() bool {
	// Check if /dev/kvm exists
	if _, err := os.Stat("/dev/kvm"); err == nil {
		return true
	}

	// Check if KVM modules are loaded
	if output, err := exec.Command("lsmod").Output(); err == nil {
		return strings.Contains(string(output), "kvm")
	}

	return false
}

// checkHardwareVirtualization checks for hardware virtualization support
func checkHardwareVirtualization() bool {
	switch runtime.GOOS {
	case "linux":
		// Check CPU flags for VT-x/AMD-V
		if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
			return strings.Contains(string(data), "vmx") || strings.Contains(string(data), "svm")
		}
	case "windows":
		// Try native API first
		if nativeCheckHardwareVirt != nil {
			return nativeCheckHardwareVirt()
		}
		// Fallback: Check Windows virtualization features via PowerShell
		if output, err := exec.Command("powershell", "-Command", "Get-ComputerInfo | Select-Object -ExpandProperty HyperVisorPresent").Output(); err == nil {
			return strings.Contains(string(output), "True")
		}
	}
	return false
}

// getRecommendedBackend returns the recommended virtualization backend for the platform
func getRecommendedBackend(osType string, virtInfo *VirtualizationInfo) string {
	switch osType {
	case "Linux":
		if virtInfo.QEMU && virtInfo.KVMSupport {
			return "qemu"
		}
		if virtInfo.VirtualBox {
			return "virtualbox"
		}
	case "Windows":
		if virtInfo.VirtualBox {
			return "virtualbox"
		}
		if virtInfo.QEMU {
			return "qemu" // WSL2 scenario
		}
	case "macOS":
		if virtInfo.VirtualBox {
			return "virtualbox"
		}
	}
	return ""
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// isVirtualBoxAvailable performs enhanced VirtualBox detection
func isVirtualBoxAvailable() bool {
	// Method 1: Try PATH first
	if _, err := exec.LookPath("VBoxManage"); err == nil {
		return true
	}

	// Method 2: Platform-specific detection
	switch runtime.GOOS {
	case "windows":
		return isVirtualBoxAvailableWindows()
	case "linux":
		return isVirtualBoxAvailableLinux()
	case "darwin":
		return isVirtualBoxAvailableMacOS()
	default:
		return false
	}
}

// isVirtualBoxAvailableWindows checks for VirtualBox on Windows using registry
func isVirtualBoxAvailableWindows() bool {
	// Try native API first
	if nativeIsVirtualBoxAvailable != nil {
		return nativeIsVirtualBoxAvailable()
	}

	// Fallback: Check registry keys
	regKeys := []string{
		"HKEY_LOCAL_MACHINE\\SOFTWARE\\Oracle\\VirtualBox",
		"HKEY_LOCAL_MACHINE\\SOFTWARE\\WOW6432Node\\Oracle\\VirtualBox",
		"HKEY_CURRENT_USER\\SOFTWARE\\Oracle\\VirtualBox",
	}

	for _, regKey := range regKeys {
		if path := queryWindowsRegistry(regKey, "InstallDir"); path != "" {
			vboxManagePath := filepath.Join(path, "VBoxManage.exe")
			if _, err := os.Stat(vboxManagePath); err == nil {
				return true
			}
		}
	}

	// Check common installation paths
	commonPaths := []string{
		"C:\\Program Files\\Oracle\\VirtualBox\\VBoxManage.exe",
		"C:\\Program Files (x86)\\Oracle\\VirtualBox\\VBoxManage.exe",
		"C:\\VirtualBox\\VBoxManage.exe",
		"D:\\Program Files\\Oracle\\VirtualBox\\VBoxManage.exe",
		"D:\\Program Files (x86)\\Oracle\\VirtualBox\\VBoxManage.exe",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// isVirtualBoxAvailableLinux checks for VirtualBox on Linux
func isVirtualBoxAvailableLinux() bool {
	commonPaths := []string{
		"/usr/bin/VBoxManage",
		"/usr/local/bin/VBoxManage",
		"/opt/VirtualBox/VBoxManage",
		"/snap/bin/VBoxManage",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// isVirtualBoxAvailableMacOS checks for VirtualBox on macOS
func isVirtualBoxAvailableMacOS() bool {
	commonPaths := []string{
		"/usr/local/bin/VBoxManage",
		"/Applications/VirtualBox.app/Contents/MacOS/VBoxManage",
		"/opt/homebrew/bin/VBoxManage",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// queryWindowsRegistry queries Windows registry for a value
func queryWindowsRegistry(keyPath, valueName string) string {
	if runtime.GOOS != "windows" {
		return ""
	}

	// Try native API first
	if nativeQueryRegistry != nil {
		if val := nativeQueryRegistry(keyPath, valueName); val != "" {
			return val
		}
	}

	// Fallback to reg command
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
