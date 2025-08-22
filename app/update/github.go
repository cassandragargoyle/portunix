package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	githubOwner = "cassandragargoyle"
	githubRepo  = "Portunix"
	githubAPI   = "https://api.github.com"
)

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

// DownloadUpdate downloads the update binary to a temporary file
func DownloadUpdate(release *ReleaseInfo) (string, error) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "portunix-update-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpFile.Close()
	
	// Download the binary
	client := &http.Client{
		Timeout: 5 * time.Minute, // Allow up to 5 minutes for download
	}
	
	resp, err := client.Get(release.DownloadURL)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	
	// Open temp file for writing
	out, err := os.OpenFile(tmpFile.Name(), os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to open temp file: %w", err)
	}
	defer out.Close()
	
	// Copy download to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to save download: %w", err)
	}
	
	return tmpFile.Name(), nil
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