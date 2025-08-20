package system

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SystemInfo contains comprehensive system information
type SystemInfo struct {
	OS           string            `json:"os"`
	Version      string            `json:"version"`
	Build        string            `json:"build"`
	Architecture string            `json:"architecture"`
	Hostname     string            `json:"hostname"`
	Variant      string            `json:"variant"`
	Environment  []string          `json:"environment"`
	WindowsInfo  *WindowsInfo      `json:"windows_info,omitempty"`
	LinuxInfo    *LinuxInfo        `json:"linux_info,omitempty"`
	Capabilities *Capabilities     `json:"capabilities"`
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
	PowerShell bool `json:"powershell"`
	Docker     bool `json:"docker"`
	Admin      bool `json:"admin"`
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
			}
		}
	}
	
	// Get kernel version
	if output, err := exec.Command("uname", "-r").Output(); err == nil {
		info.LinuxInfo.KernelVersion = strings.TrimSpace(string(output))
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
	// Check for Windows Sandbox
	if info.OS == "Windows" {
		// Windows Sandbox typically has WDAGUtilityAccount user
		if username := os.Getenv("USERNAME"); username == "WDAGUtilityAccount" {
			info.Environment = append(info.Environment, "sandbox")
			info.Variant = "Sandbox"
		}
		
		// Check for specific sandbox registry keys or processes
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
		// Check for common VM indicators
		vmIndicators := []string{"VMware", "VirtualBox", "Hyper-V", "QEMU", "Xen"}
		if output, err := exec.Command("wmic", "computersystem", "get", "manufacturer,model").Output(); err == nil {
			outputStr := strings.ToLower(string(output))
			for _, indicator := range vmIndicators {
				if strings.Contains(outputStr, strings.ToLower(indicator)) {
					info.Environment = append(info.Environment, "vm")
					if info.Variant == "" {
						info.Variant = "VM"
					}
					break
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
	}
	
	// Check admin privileges (platform-specific)
	if info.OS == "Windows" {
		// Check if running as administrator
		if output, err := exec.Command("net", "session").Output(); err == nil {
			if len(output) > 0 {
				info.Capabilities.Admin = true
			}
		}
	} else {
		// Unix-like systems - check if running as root
		if os.Geteuid() == 0 {
			info.Capabilities.Admin = true
		}
	}
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