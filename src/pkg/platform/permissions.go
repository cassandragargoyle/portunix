package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// IsRunningAsRoot checks if the current process is running with root/administrator privileges
func IsRunningAsRoot() bool {
	if runtime.GOOS == "windows" {
		// On Windows, check if running as Administrator
		// This is a simplified check - full implementation would use Windows API
		return false // Conservative default for now
	}

	// On Unix-like systems, check if UID is 0
	return os.Geteuid() == 0
}

// IsSudoAvailable checks if sudo command is available on the system
// Only applicable on Unix-like systems
func IsSudoAvailable() bool {
	if runtime.GOOS == "windows" {
		return false
	}

	// Check if sudo binary exists
	_, err := os.Stat("/usr/bin/sudo")
	return err == nil
}

// GetSudoPrefix returns the appropriate sudo prefix for commands
// Returns "sudo " if sudo is needed and available, empty string otherwise
func GetSudoPrefix() string {
	if runtime.GOOS == "windows" {
		return ""
	}

	if IsRunningAsRoot() {
		return ""
	}

	if IsSudoAvailable() {
		return "sudo "
	}

	return ""
}

// RequiresSudo checks if a given path requires sudo to write
// Returns true if the path is a system directory and we're not root
func RequiresSudo(path string) bool {
	// If running as root, sudo not needed
	if IsRunningAsRoot() {
		return false
	}

	// If sudo not available, return false (can't use it anyway)
	if !IsSudoAvailable() {
		return false
	}

	// Check if path is a system directory
	systemDirs := []string{
		"/usr/local",
		"/usr/bin",
		"/usr/sbin",
		"/opt",
		"/etc",
	}

	for _, sysDir := range systemDirs {
		if strings.HasPrefix(path, sysDir) {
			return !CanWriteToDirectory(path)
		}
	}

	return false
}

// CanWriteToDirectory checks if the current user can write to the specified directory
func CanWriteToDirectory(dir string) bool {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false
	}

	// Try to create a test file
	testFile := filepath.Join(dir, ".portunix_write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testFile)

	return true
}

// IsUserDirectory checks if a directory is within user's home directory
func IsUserDirectory(dir string) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	return strings.HasPrefix(dir, homeDir)
}
