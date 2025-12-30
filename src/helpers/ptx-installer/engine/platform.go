package engine

import (
	"os"
	"path/filepath"

	"portunix.ai/portunix/src/pkg/platform"
)

// GetOperatingSystem returns the current operating system identifier
// This is a wrapper around platform.GetOS() for backward compatibility
func GetOperatingSystem() string {
	return platform.GetOS()
}

// GetArchitecture returns the current system architecture
// This is a wrapper around platform.GetArchitecture() for backward compatibility
func GetArchitecture() string {
	return platform.GetArchitecture()
}

// IsRunningAsRoot checks if the current process is running with root privileges
// This is a wrapper around platform.IsRunningAsRoot() for backward compatibility
func IsRunningAsRoot() bool {
	return platform.IsRunningAsRoot()
}

// IsSudoAvailable checks if sudo command is available
// This is a wrapper around platform.IsSudoAvailable() for backward compatibility
func IsSudoAvailable() bool {
	return platform.IsSudoAvailable()
}

// DetermineSudoPrefix returns the appropriate sudo prefix for commands
// This is a wrapper around platform.GetSudoPrefix() for backward compatibility
func DetermineSudoPrefix() string {
	return platform.GetSudoPrefix()
}

// IsUserDirectory checks if a directory is in user space
// This is a wrapper around platform.IsUserDirectory() for backward compatibility
func IsUserDirectory(dir string) bool {
	return platform.IsUserDirectory(dir)
}

// CanWriteToDirectory checks if current user can write to directory
// This is a wrapper around platform.CanWriteToDirectory() for backward compatibility
func CanWriteToDirectory(dir string) bool {
	return platform.CanWriteToDirectory(dir)
}

// CheckDirectoryPermissions checks if directory needs sudo and suggests fallback
func CheckDirectoryPermissions(destDir string) (requiresSudo bool, fallbackDir string) {
	// If running as root, no sudo needed
	if IsRunningAsRoot() {
		return false, ""
	}

	// Check if we can write to the directory
	if CanWriteToDirectory(destDir) {
		return false, ""
	}

	// If it's a system directory and we can't write, suggest user directory
	if !IsUserDirectory(destDir) && IsSudoAvailable() {
		homeDir, _ := os.UserHomeDir()
		fallbackDir = filepath.Join(homeDir, ".local", "bin")
		return true, fallbackDir
	}

	return false, ""
}
