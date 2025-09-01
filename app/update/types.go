package update

import (
	"fmt"
	"runtime"
	"strings"
)

// Version is set at build time using ldflags
var Version = "v1.4.0"

// ReleaseInfo contains information about a GitHub release
type ReleaseInfo struct {
	Version     string
	DownloadURL string
	ChecksumURL string
	Size        int64
	PublishedAt string
}

// GitHubRelease represents a GitHub release API response
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	PublishedAt string        `json:"published_at"`
	Assets      []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a release asset
type GitHubAsset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GetOS returns the normalized OS name for release assets
func GetOS() string {
	return runtime.GOOS
}

// GetArch returns the normalized architecture name for release assets
func GetArch() string {
	return runtime.GOARCH
}

// GetBinaryName returns the expected binary name for the current platform
func GetBinaryName(version string) string {
	// Remove 'v' prefix from version for filename
	version = strings.TrimPrefix(version, "v")

	// Format: portunix_1.4.0_linux_amd64.tar.gz
	name := fmt.Sprintf("portunix_%s_%s_%s", version, GetOS(), GetArch())

	// Add appropriate archive extension
	if runtime.GOOS == "windows" {
		name += ".zip"
	} else {
		name += ".tar.gz"
	}

	return name
}

// GetChecksumName returns the expected checksum file name
func GetChecksumName(version string) string {
	// Remove 'v' prefix from version for filename
	version = strings.TrimPrefix(version, "v")

	// Format: checksums_1.4.0.txt
	return fmt.Sprintf("checksums_%s.txt", version)
}

// CompareVersions compares two semantic versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Split versions into parts
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	// Compare major, minor, patch
	for i := 0; i < 3; i++ {
		if parts1[i] < parts2[i] {
			return -1
		}
		if parts1[i] > parts2[i] {
			return 1
		}
	}

	return 0
}

// parseVersion parses a version string into [major, minor, patch]
func parseVersion(version string) [3]int {
	var parts [3]int
	versionParts := strings.Split(version, ".")

	for i := 0; i < len(versionParts) && i < 3; i++ {
		// Parse the number, ignoring any non-numeric suffix
		fmt.Sscanf(versionParts[i], "%d", &parts[i])
	}

	return parts
}
