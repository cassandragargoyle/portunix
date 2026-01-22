package main

import (
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

// OSInfo contains OS-specific information for seed generation
// Must be compatible with Java's System.getProperty() values
type OSInfo struct {
	Hostname string
	Username string
	OSName   string // Must match Java: "Windows 10", "Windows 11", "Linux", "Mac OS X"
	HomeDir  string
}

// GetOSInfo retrieves OS information compatible with Java TokenStorage
func GetOSInfo() (*OSInfo, error) {
	info := &OSInfo{}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	info.Hostname = hostname

	// Get username
	currentUser, err := user.Current()
	if err != nil {
		info.Username = os.Getenv("USER")
		if info.Username == "" {
			info.Username = os.Getenv("USERNAME")
		}
	} else {
		info.Username = currentUser.Username
		// On Windows, username may include domain (DOMAIN\user), extract just the username
		if runtime.GOOS == "windows" {
			parts := strings.Split(info.Username, "\\")
			if len(parts) > 1 {
				info.Username = parts[len(parts)-1]
			}
		}
	}

	// Get home directory
	info.HomeDir, err = os.UserHomeDir()
	if err != nil {
		if currentUser != nil {
			info.HomeDir = currentUser.HomeDir
		} else {
			info.HomeDir = os.Getenv("HOME")
			if info.HomeDir == "" {
				info.HomeDir = os.Getenv("USERPROFILE")
			}
		}
	}

	// Get OS name - must match Java's System.getProperty("os.name")
	info.OSName = getJavaCompatibleOSName()

	return info, nil
}

// getJavaCompatibleOSName returns OS name compatible with Java's System.getProperty("os.name")
// Java returns:
// - Windows: "Windows 10", "Windows 11", "Windows Server 2019", etc.
// - Linux: "Linux"
// - macOS: "Mac OS X"
func getJavaCompatibleOSName() string {
	switch runtime.GOOS {
	case "windows":
		return getWindowsOSName()
	case "linux":
		return "Linux"
	case "darwin":
		return "Mac OS X"
	default:
		return runtime.GOOS
	}
}

// getWindowsOSName returns the Windows OS name matching Java's format
func getWindowsOSName() string {
	// Try to get Windows version using wmic
	output, err := exec.Command("wmic", "os", "get", "Caption", "/value").Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Caption=") {
				caption := strings.TrimPrefix(line, "Caption=")
				caption = strings.TrimSpace(caption)
				// Extract Windows version from caption
				// e.g., "Microsoft Windows 11 Pro" -> "Windows 11"
				// e.g., "Microsoft Windows 10 Enterprise" -> "Windows 10"
				if strings.Contains(caption, "Windows 11") {
					return "Windows 11"
				} else if strings.Contains(caption, "Windows 10") {
					return "Windows 10"
				} else if strings.Contains(caption, "Windows Server") {
					// Extract server version
					if strings.Contains(caption, "2022") {
						return "Windows Server 2022"
					} else if strings.Contains(caption, "2019") {
						return "Windows Server 2019"
					} else if strings.Contains(caption, "2016") {
						return "Windows Server 2016"
					}
					return "Windows Server"
				}
				// Fallback: try to extract "Windows X" pattern
				if idx := strings.Index(caption, "Windows"); idx >= 0 {
					rest := caption[idx:]
					parts := strings.Fields(rest)
					if len(parts) >= 2 {
						return parts[0] + " " + parts[1]
					}
				}
			}
		}
	}

	// Fallback: detect by build number
	return detectWindowsByBuildNumber()
}

// detectWindowsByBuildNumber detects Windows version by build number
func detectWindowsByBuildNumber() string {
	output, err := exec.Command("cmd", "/c", "ver").Output()
	if err != nil {
		return "Windows"
	}

	verStr := string(output)
	// Windows 11 starts from build 22000
	if strings.Contains(verStr, "Version 10.0") {
		// Extract build number
		buildOutput, err := exec.Command("cmd", "/c", "wmic os get BuildNumber /value").Output()
		if err == nil {
			lines := strings.Split(string(buildOutput), "\n")
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "BuildNumber=") {
					buildNum := strings.TrimPrefix(strings.TrimSpace(line), "BuildNumber=")
					buildNum = strings.TrimSpace(buildNum)
					if buildNum >= "22000" {
						return "Windows 11"
					}
					return "Windows 10"
				}
			}
		}
	}

	return "Windows"
}

// GenerateDefaultSeed generates the default seed for credential encryption
func GenerateDefaultSeed() (string, error) {
	info, err := GetOSInfo()
	if err != nil {
		return "", err
	}
	return GenerateSeed(info.Hostname, info.Username, info.OSName, info.HomeDir, "portunix-credential"), nil
}

// GenerateDefaultM365Seed generates the M365-compatible seed
func GenerateDefaultM365Seed() (string, error) {
	info, err := GetOSInfo()
	if err != nil {
		return "", err
	}
	return GenerateM365Seed(info.Hostname, info.Username, info.OSName, info.HomeDir), nil
}

// GeneratePasswordProtectedSeed generates a password-protected seed
func GeneratePasswordProtectedSeed(password string) (string, error) {
	info, err := GetOSInfo()
	if err != nil {
		return "", err
	}
	return GeneratePasswordSeed(info.Hostname, info.Username, info.OSName, info.HomeDir, password), nil
}
