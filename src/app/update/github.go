package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	githubOwner = "cassandragargoyle"
	githubRepo  = "portunix"
	githubAPI   = "https://api.github.com"
)

// BinaryInfo represents information about an extracted binary
type BinaryInfo struct {
	Name string // Binary name (e.g., "portunix", "ptx-container")
	Path string // Temporary path to extracted binary
}

// CheckForUpdate checks if a newer version is available
func CheckForUpdate() (*ReleaseInfo, error) {
	latest, err := GetLatestRelease()
	if err != nil {
		return nil, err
	}

	// Compare versions
	currentVersion := Version
	if CompareVersions(currentVersion, latest.Version) >= 0 {
		return nil, nil // Already on latest or newer version
	}

	return latest, nil
}

// GetLatestRelease fetches the latest release information from GitHub
func GetLatestRelease() (*ReleaseInfo, error) {
	// First try to get the latest release (non-prerelease)
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPI, githubOwner, githubRepo)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("Portunix/%s", Version))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// If no latest release found, try to get the most recent release (including prereleases)
		return GetMostRecentRelease()
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	// Find the appropriate asset for this platform
	binaryName := GetBinaryName(release.TagName)
	checksumName := GetChecksumName(release.TagName)

	var binaryAsset, checksumAsset *GitHubAsset

	for i := range release.Assets {
		asset := &release.Assets[i]
		if asset.Name == binaryName {
			binaryAsset = asset
		} else if asset.Name == checksumName {
			checksumAsset = asset
		}
	}

	if binaryAsset == nil {
		return nil, fmt.Errorf("no binary found for %s/%s", GetOS(), GetArch())
	}

	info := &ReleaseInfo{
		Version:     release.TagName,
		DownloadURL: binaryAsset.BrowserDownloadURL,
		Size:        binaryAsset.Size,
		PublishedAt: release.PublishedAt,
	}

	if checksumAsset != nil {
		info.ChecksumURL = checksumAsset.BrowserDownloadURL
	}

	return info, nil
}

// GetMostRecentRelease fetches the most recent release (including prereleases)
func GetMostRecentRelease() (*ReleaseInfo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases", githubAPI, githubOwner, githubRepo)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("Portunix/%s", Version))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found")
	}

	// Use the first release (most recent)
	release := releases[0]

	// Find the appropriate asset for this platform
	binaryName := GetBinaryName(release.TagName)
	checksumName := GetChecksumName(release.TagName)

	var binaryAsset, checksumAsset *GitHubAsset

	for i := range release.Assets {
		asset := &release.Assets[i]
		if asset.Name == binaryName {
			binaryAsset = asset
		} else if asset.Name == checksumName {
			checksumAsset = asset
		}
	}

	if binaryAsset == nil {
		return nil, fmt.Errorf("no binary found for %s/%s", GetOS(), GetArch())
	}

	info := &ReleaseInfo{
		Version:     release.TagName,
		DownloadURL: binaryAsset.BrowserDownloadURL,
		Size:        binaryAsset.Size,
		PublishedAt: release.PublishedAt,
	}

	if checksumAsset != nil {
		info.ChecksumURL = checksumAsset.BrowserDownloadURL
	}

	return info, nil
}

// GetRelease fetches a specific release by version
func GetRelease(version string) (*ReleaseInfo, error) {
	// Ensure version has 'v' prefix
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	url := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", githubAPI, githubOwner, githubRepo, version)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("Portunix/%s", Version))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	// Find the appropriate asset for this platform
	binaryName := GetBinaryName(release.TagName)
	checksumName := GetChecksumName(release.TagName)

	var binaryAsset, checksumAsset *GitHubAsset

	for i := range release.Assets {
		asset := &release.Assets[i]
		if asset.Name == binaryName {
			binaryAsset = asset
		} else if asset.Name == checksumName {
			checksumAsset = asset
		}
	}

	if binaryAsset == nil {
		return nil, fmt.Errorf("no binary found for %s/%s", GetOS(), GetArch())
	}

	info := &ReleaseInfo{
		Version:     release.TagName,
		DownloadURL: binaryAsset.BrowserDownloadURL,
		Size:        binaryAsset.Size,
		PublishedAt: release.PublishedAt,
	}

	if checksumAsset != nil {
		info.ChecksumURL = checksumAsset.BrowserDownloadURL
	}

	return info, nil
}

// DownloadUpdate downloads the update archive and extracts the binary
func DownloadUpdate(release *ReleaseInfo) (string, error) {
	// Create temporary file for archive
	tmpArchive, err := os.CreateTemp("", "portunix-update-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpArchive.Name())
	tmpArchive.Close()

	// Download the archive
	client := &http.Client{
		Timeout: 5 * time.Minute, // Allow up to 5 minutes for download
	}

	resp, err := client.Get(release.DownloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Save archive to file
	out, err := os.OpenFile(tmpArchive.Name(), os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to open temp file: %w", err)
	}

	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return "", fmt.Errorf("failed to save download: %w", err)
	}

	// Verify archive checksum if available
	if release.ChecksumURL != "" {
		// Extract archive name from download URL
		urlParts := strings.Split(release.DownloadURL, "/")
		archiveName := urlParts[len(urlParts)-1]

		if err := VerifyArchiveChecksum(tmpArchive.Name(), release.ChecksumURL, archiveName); err != nil {
			return "", fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	// Return the archive path for multi-binary extraction
	return tmpArchive.Name(), nil
}

// extractBinary extracts the portunix binary from an archive
func extractBinary(archivePath string) (string, error) {
	// Create temporary file for extracted binary
	tmpBinary, err := os.CreateTemp("", "portunix-*.exe")
	if err != nil {
		return "", fmt.Errorf("failed to create temp binary: %w", err)
	}
	tmpBinary.Close()

	if runtime.GOOS == "windows" {
		// Extract from zip
		return extractFromZip(archivePath, tmpBinary.Name())
	} else {
		// Extract from tar.gz
		return extractFromTarGz(archivePath, tmpBinary.Name())
	}
}

// extractFromZip extracts portunix binary from a zip archive
func extractFromZip(zipPath, destPath string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	// Find portunix binary in the archive
	for _, file := range reader.File {
		if strings.Contains(file.Name, "portunix") && (strings.HasSuffix(file.Name, ".exe") || !strings.Contains(file.Name, ".")) {
			rc, err := file.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open file in zip: %w", err)
			}
			defer rc.Close()

			// Write to destination
			out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return "", fmt.Errorf("failed to create output file: %w", err)
			}
			defer out.Close()

			_, err = io.Copy(out, rc)
			if err != nil {
				return "", fmt.Errorf("failed to extract file: %w", err)
			}

			return destPath, nil
		}
	}

	return "", fmt.Errorf("portunix binary not found in zip")
}

// extractFromTarGz extracts portunix binary from a tar.gz archive
func extractFromTarGz(tarGzPath, destPath string) (string, error) {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return "", fmt.Errorf("failed to open tar.gz: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Find portunix binary in the archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar: %w", err)
		}

		// Check if this is the portunix binary
		if strings.Contains(header.Name, "portunix") && header.Typeflag == tar.TypeReg {
			// Write to destination
			out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return "", fmt.Errorf("failed to create output file: %w", err)
			}
			defer out.Close()

			_, err = io.Copy(out, tarReader)
			if err != nil {
				return "", fmt.Errorf("failed to extract file: %w", err)
			}

			// Set executable permissions
			if err := os.Chmod(destPath, 0755); err != nil {
				return "", fmt.Errorf("failed to set permissions: %w", err)
			}

			return destPath, nil
		}
	}

	return "", fmt.Errorf("portunix binary not found in tar.gz")
}

// DownloadFile downloads a file from a URL
func DownloadFile(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// extractAllFromZip extracts all portunix binaries from a zip archive
func extractAllFromZip(zipPath string) ([]*BinaryInfo, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	var binaries []*BinaryInfo
	binSuffix := ".exe"

	// Expected binary names
	expectedBinaries := []string{"portunix", "ptx-container", "ptx-mcp"}

	for _, expectedBinary := range expectedBinaries {
		found := false

		// Find the binary in the archive
		for _, file := range reader.File {
			fileName := strings.ToLower(file.Name)
			if strings.Contains(fileName, expectedBinary) && strings.HasSuffix(fileName, binSuffix) {
				// Create temporary file
				tmpFile, err := os.CreateTemp("", expectedBinary+"-*.exe")
				if err != nil {
					// Clean up any previously extracted files
					for _, binary := range binaries {
						os.Remove(binary.Path)
					}
					return nil, fmt.Errorf("failed to create temp file for %s: %w", expectedBinary, err)
				}
				tmpFile.Close()

				// Extract the binary
				rc, err := file.Open()
				if err != nil {
					os.Remove(tmpFile.Name())
					for _, binary := range binaries {
						os.Remove(binary.Path)
					}
					return nil, fmt.Errorf("failed to open %s in zip: %w", expectedBinary, err)
				}

				out, err := os.OpenFile(tmpFile.Name(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
				if err != nil {
					rc.Close()
					os.Remove(tmpFile.Name())
					for _, binary := range binaries {
						os.Remove(binary.Path)
					}
					return nil, fmt.Errorf("failed to create output file for %s: %w", expectedBinary, err)
				}

				_, err = io.Copy(out, rc)
				rc.Close()
				out.Close()

				if err != nil {
					os.Remove(tmpFile.Name())
					for _, binary := range binaries {
						os.Remove(binary.Path)
					}
					return nil, fmt.Errorf("failed to extract %s: %w", expectedBinary, err)
				}

				binaries = append(binaries, &BinaryInfo{
					Name: expectedBinary,
					Path: tmpFile.Name(),
				})

				found = true
				break
			}
		}

		// Main binary is required, helpers are optional
		if !found && expectedBinary == "portunix" {
			for _, binary := range binaries {
				os.Remove(binary.Path)
			}
			return nil, fmt.Errorf("main binary %s not found in zip", expectedBinary)
		}
	}

	if len(binaries) == 0 {
		return nil, fmt.Errorf("no portunix binaries found in zip")
	}

	return binaries, nil
}

// extractAllFromTarGz extracts all portunix binaries from a tar.gz archive
func extractAllFromTarGz(tarGzPath string) ([]*BinaryInfo, error) {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tar.gz: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	var binaries []*BinaryInfo

	// Expected binary names
	expectedBinaries := []string{"portunix", "ptx-container", "ptx-mcp"}
	foundBinaries := make(map[string]bool)

	// Extract all binaries from tar
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Clean up any extracted files
			for _, binary := range binaries {
				os.Remove(binary.Path)
			}
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		// Check if this is one of our expected binaries
		if header.Typeflag == tar.TypeReg {
			fileName := strings.ToLower(header.Name)

			for _, expectedBinary := range expectedBinaries {
				if strings.Contains(fileName, expectedBinary) && !strings.Contains(fileName, ".") {
					// Create temporary file
					tmpFile, err := os.CreateTemp("", expectedBinary+"-*")
					if err != nil {
						for _, binary := range binaries {
							os.Remove(binary.Path)
						}
						return nil, fmt.Errorf("failed to create temp file for %s: %w", expectedBinary, err)
					}
					tmpFile.Close()

					// Extract the binary
					out, err := os.OpenFile(tmpFile.Name(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
					if err != nil {
						os.Remove(tmpFile.Name())
						for _, binary := range binaries {
							os.Remove(binary.Path)
						}
						return nil, fmt.Errorf("failed to create output file for %s: %w", expectedBinary, err)
					}

					_, err = io.Copy(out, tarReader)
					out.Close()

					if err != nil {
						os.Remove(tmpFile.Name())
						for _, binary := range binaries {
							os.Remove(binary.Path)
						}
						return nil, fmt.Errorf("failed to extract %s: %w", expectedBinary, err)
					}

					// Set executable permissions
					if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
						os.Remove(tmpFile.Name())
						for _, binary := range binaries {
							os.Remove(binary.Path)
						}
						return nil, fmt.Errorf("failed to set permissions for %s: %w", expectedBinary, err)
					}

					if !foundBinaries[expectedBinary] {
						binaries = append(binaries, &BinaryInfo{
							Name: expectedBinary,
							Path: tmpFile.Name(),
						})
						foundBinaries[expectedBinary] = true
					}

					break
				}
			}
		}
	}

	// Main binary is required
	if !foundBinaries["portunix"] {
		for _, binary := range binaries {
			os.Remove(binary.Path)
		}
		return nil, fmt.Errorf("main binary portunix not found in tar.gz")
	}

	if len(binaries) == 0 {
		return nil, fmt.Errorf("no portunix binaries found in tar.gz")
	}

	return binaries, nil
}
