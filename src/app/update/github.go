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
	"path"
	"runtime"
	"strings"
	"time"
)

const (
	githubOwner = "cassandragargoyle"
	githubRepo  = "portunix"
	githubAPI   = "https://api.github.com"
)

// maxExtractedBinarySize caps a single binary extracted from a release archive
// to stop decompression bombs
const maxExtractedBinarySize int64 = 1 << 30

// copyLimited copies src to dst and fails once more than limit bytes were written
func copyLimited(dst io.Writer, src io.Reader, limit int64) error {
	written, err := io.CopyN(dst, src, limit+1)
	if err != nil && err != io.EOF {
		return err
	}
	if written > limit {
		return fmt.Errorf("extracted file exceeds the %d byte limit", limit)
	}
	return nil
}

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

	return buildReleaseInfo(&release)
}

// buildReleaseInfo locates the platform-specific binary, checksum and
// signature assets in a GitHub release and assembles a ReleaseInfo.
func buildReleaseInfo(release *GitHubRelease) (*ReleaseInfo, error) {
	binaryName := GetBinaryName(release.TagName)
	checksumName := GetChecksumName(release.TagName)
	signatureName := GetSignatureName(release.TagName)

	var binaryAsset, checksumAsset, signatureAsset *GitHubAsset

	for i := range release.Assets {
		asset := &release.Assets[i]
		switch asset.Name {
		case binaryName:
			binaryAsset = asset
		case checksumName:
			checksumAsset = asset
		case signatureName:
			signatureAsset = asset
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

	if signatureAsset != nil {
		info.SignatureURL = signatureAsset.BrowserDownloadURL
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
	return buildReleaseInfo(&release)
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

	return buildReleaseInfo(&release)
}

// DownloadUpdate downloads and verifies the update archive, the caller
// extracts it and removes the returned file
func DownloadUpdate(release *ReleaseInfo) (archivePath string, err error) {
	// Create temporary file for archive
	tmpArchive, err := os.CreateTemp("", "portunix-update-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	// The archive is kept on success, the caller still needs it
	defer func() {
		if err != nil {
			_ = os.Remove(tmpArchive.Name())
		}
	}()
	_ = tmpArchive.Close()

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
	out, err := os.OpenFile(tmpArchive.Name(), os.O_WRONLY|os.O_TRUNC, 0755) // #nosec G302 -- binaries must be executable by all users
	if err != nil {
		return "", fmt.Errorf("failed to open temp file: %w", err)
	}

	_, err = io.Copy(out, resp.Body)
	_ = out.Close()
	if err != nil {
		return "", fmt.Errorf("failed to save download: %w", err)
	}

	// Verify archive checksum (and signature, if present) if checksum is available
	if release.ChecksumURL != "" {
		// Extract archive name from download URL
		urlParts := strings.Split(release.DownloadURL, "/")
		archiveName := urlParts[len(urlParts)-1]

		if err := VerifyArchiveChecksumSigned(tmpArchive.Name(), release.ChecksumURL, release.SignatureURL, archiveName); err != nil {
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
	_ = tmpBinary.Close()

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
			out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755) // #nosec G302 -- binaries must be executable by all users
			if err != nil {
				return "", fmt.Errorf("failed to create output file: %w", err)
			}
			defer out.Close()

			err = copyLimited(out, rc, maxExtractedBinarySize)
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
			out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755) // #nosec G302 -- binaries must be executable by all users
			if err != nil {
				return "", fmt.Errorf("failed to create output file: %w", err)
			}
			defer out.Close()

			err = copyLimited(out, tarReader, maxExtractedBinarySize)
			if err != nil {
				return "", fmt.Errorf("failed to extract file: %w", err)
			}

			// Set executable permissions
			if err := os.Chmod(destPath, 0755); err != nil { // #nosec G302 -- binaries must be executable by all users
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

// releaseBinaryName returns the binary name (without .exe) of an archive entry
// when the entry is the main binary or a ptx-* helper, so helpers added under
// src/helpers/ are picked up without a change here
func releaseBinaryName(entryName string, windows bool) (string, bool) {
	name := path.Base(strings.ReplaceAll(entryName, "\\", "/"))
	if windows {
		if !strings.HasSuffix(strings.ToLower(name), ".exe") {
			return "", false
		}
		name = name[:len(name)-len(".exe")]
	} else if strings.Contains(name, ".") {
		return "", false
	}

	if name == "portunix" || (strings.HasPrefix(name, "ptx-") && len(name) > len("ptx-")) {
		return name, true
	}
	return "", false
}

// extractToTemp writes one archive entry to a temporary executable file
func extractToTemp(name string, src io.Reader) (string, error) {
	pattern := name + "-*"
	if runtime.GOOS == "windows" {
		pattern += ".exe"
	}
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file for %s: %w", name, err)
	}

	err = copyLimited(tmpFile, src, maxExtractedBinarySize)
	_ = tmpFile.Close()
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to extract %s: %w", name, err)
	}

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil { // #nosec G302 -- binaries must be executable by all users
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to set permissions for %s: %w", name, err)
	}

	return tmpFile.Name(), nil
}

// removeExtracted deletes temporary files of already extracted binaries
func removeExtracted(binaries []*BinaryInfo) {
	for _, binary := range binaries {
		_ = os.Remove(binary.Path)
	}
}

// requireMainBinary fails when the extracted set lacks the main binary
func requireMainBinary(binaries []*BinaryInfo, archiveKind string) error {
	for _, binary := range binaries {
		if binary.Name == "portunix" {
			return nil
		}
	}
	removeExtracted(binaries)
	return fmt.Errorf("main binary portunix not found in %s", archiveKind)
}

// extractAllFromZip extracts the main binary and every ptx-* helper from a zip archive
func extractAllFromZip(zipPath string) ([]*BinaryInfo, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	var binaries []*BinaryInfo
	seen := make(map[string]bool)

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		name, ok := releaseBinaryName(file.Name, true)
		if !ok || seen[name] {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			removeExtracted(binaries)
			return nil, fmt.Errorf("failed to open %s in zip: %w", name, err)
		}
		tmpPath, err := extractToTemp(name, rc)
		_ = rc.Close()
		if err != nil {
			removeExtracted(binaries)
			return nil, err
		}

		binaries = append(binaries, &BinaryInfo{Name: name, Path: tmpPath})
		seen[name] = true
	}

	if err := requireMainBinary(binaries, "zip"); err != nil {
		return nil, err
	}
	return binaries, nil
}

// extractAllFromTarGz extracts the main binary and every ptx-* helper from a tar.gz archive
func extractAllFromTarGz(tarGzPath string) ([]*BinaryInfo, error) {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tar.gz: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var binaries []*BinaryInfo
	seen := make(map[string]bool)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			removeExtracted(binaries)
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		name, ok := releaseBinaryName(header.Name, false)
		if !ok || seen[name] {
			continue
		}

		tmpPath, err := extractToTemp(name, tarReader)
		if err != nil {
			removeExtracted(binaries)
			return nil, err
		}

		binaries = append(binaries, &BinaryInfo{Name: name, Path: tmpPath})
		seen[name] = true
	}

	if err := requireMainBinary(binaries, "tar.gz"); err != nil {
		return nil, err
	}
	return binaries, nil
}
