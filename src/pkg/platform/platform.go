// Package platform provides cross-platform utilities for OS, architecture,
// and environment detection. This package is shared between the main portunix
// binary and all helper binaries to ensure consistent behavior.
//
// Related: ADR-026 (Shared Platform Utilities)
package platform

import (
	"os"
	"runtime"
)

// GetOS returns the normalized operating system identifier
// Returns: "windows", "linux", "darwin", or "windows_sandbox"
func GetOS() string {
	// Check for Windows Sandbox environment variable
	if os.Getenv("PORTUNIX_SANDBOX") == "windows" {
		return "windows_sandbox"
	}

	// Standard OS detection
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	case "darwin":
		return "darwin"
	default:
		return runtime.GOOS
	}
}

// GetArchitecture returns the normalized architecture identifier
// Normalizes Go's GOARCH values to package registry conventions:
// - amd64 → x64
// - 386 → x86
// - arm64 → arm64 (unchanged)
// - arm → arm (unchanged)
func GetArchitecture() string {
	arch := runtime.GOARCH

	// Normalize architecture names to match package registry convention
	switch arch {
	case "amd64":
		return "x64"
	case "386":
		return "x86"
	case "arm64":
		return "arm64"
	case "arm":
		return "arm"
	default:
		// For unknown architectures, return as-is for forward compatibility
		return arch
	}
}

// GetPlatform returns a combined platform identifier
// Format: "{os}-{arch}" (e.g., "linux-x64", "windows-x64", "darwin-arm64")
func GetPlatform() string {
	return GetOS() + "-" + GetArchitecture()
}

// IsWindows returns true if running on Windows
func IsWindows() bool {
	return runtime.GOOS == "windows" || os.Getenv("PORTUNIX_SANDBOX") == "windows"
}

// IsLinux returns true if running on Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsDarwin returns true if running on macOS
func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

// IsWindowsSandbox returns true if running in Windows Sandbox
func IsWindowsSandbox() bool {
	return os.Getenv("PORTUNIX_SANDBOX") == "windows"
}
