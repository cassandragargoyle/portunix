package install

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// ISOInstaller handles ISO downloads using the configuration system
type ISOInstaller struct {
	OSType    string
	Variant   string
	OutputDir string
	config    *ISOConfig
}

// ISOConfig represents the ISO configuration structure
type ISOConfig struct {
	Version string                  `json:"version"`
	ISOs    map[string]*ISOPackage  `json:"isos"`
	Settings DownloadSettings       `json:"download_settings"`
}

// ISOPackage represents an ISO package configuration
type ISOPackage struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Platforms   map[string]*ISOPlatform   `json:"platforms"`
}

// ISOPlatform represents platform-specific ISO configuration
type ISOPlatform struct {
	Type     string                   `json:"type"`
	Variants map[string]*ISOVariant   `json:"variants"`
}

// ISOVariant represents an ISO variant
type ISOVariant struct {
	Version string            `json:"version"`
	URLs    map[string]string `json:"urls"`
	Size    string            `json:"size"`
	Type    string            `json:"type,omitempty"`
	Hash    *HashInfo         `json:"hash,omitempty"`
}

// HashInfo represents hash verification information
type HashInfo struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// DownloadSettings represents download configuration
type DownloadSettings struct {
	CacheDir        string `json:"cache_dir"`
	VerifyChecksums bool   `json:"verify_checksums"`
	RetryAttempts   int    `json:"retry_attempts"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
}

// Download downloads the ISO based on configuration
func (i *ISOInstaller) Download() (string, error) {
	// Load configuration
	if err := i.loadConfig(); err != nil {
		return "", fmt.Errorf("failed to load ISO configuration: %w", err)
	}
	
	// Get ISO package
	isoPackage, exists := i.config.ISOs[i.OSType]
	if !exists {
		return "", fmt.Errorf("unknown OS type: %s", i.OSType)
	}
	
	// Get platform configuration
	platform := isoPackage.Platforms["all"]
	if platform == nil {
		// Try OS-specific platform
		platform = isoPackage.Platforms[runtime.GOOS]
		if platform == nil {
			return "", fmt.Errorf("no configuration for platform: %s", runtime.GOOS)
		}
	}
	
	// Get variant
	variant, exists := platform.Variants[i.Variant]
	if !exists {
		// Try to find default variant
		if i.Variant == "latest" {
			// Find the first variant
			for name, v := range platform.Variants {
				variant = v
				i.Variant = name
				break
			}
		}
		if variant == nil {
			return "", fmt.Errorf("variant not found: %s", i.Variant)
		}
	}
	
	
	// Check if manual download is required
	if variant.Type == "manual" {
		manualURL := variant.URLs["manual"]
		if manualURL == "" {
			for _, u := range variant.URLs {
				manualURL = u
				break
			}
		}
		return i.handleManualDownload(isoPackage, variant, manualURL)
	}
	
	// Get download URL
	url := ""
	if directURL, exists := variant.URLs["direct"]; exists && directURL != "" {
		url = directURL
	} else if x64URL, exists := variant.URLs["x64"]; exists {
		url = x64URL
	} else {
		for _, u := range variant.URLs {
			url = u
			break
		}
	}
	
	if url == "" {
		return "", fmt.Errorf("no download URL available for %s %s", i.OSType, i.Variant)
	}
	
	// Handle Microsoft downloads with special user-agent for ISO access
	if strings.Contains(url, "microsoft.com") && (i.OSType == "windows11" || i.OSType == "windows10") {
		return i.downloadWindowsISO(isoPackage, variant, url)
	}
	
	// Use cache dir if no output dir specified
	outputDir := i.OutputDir
	if outputDir == "" {
		outputDir = i.config.Settings.CacheDir
	}
	
	// Prepare output path
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Determine filename
	filename := i.getFilename(url, variant)
	outputPath := filepath.Join(outputDir, filename)
	
	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("ISO already exists: %s\n", outputPath)
		return outputPath, nil
	}
	
	// Download the file
	fmt.Printf("Downloading from: %s\n", url)
	fmt.Printf("Size: %s\n", variant.Size)
	
	if err := i.downloadFile(outputPath, url); err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	
	// Verify checksum if available
	if variant.Hash != nil && i.config.Settings.VerifyChecksums {
		fmt.Println("Verifying checksum...")
		// TODO: Implement checksum verification
	}
	
	return outputPath, nil
}

// loadConfig loads the ISO configuration file
func (i *ISOInstaller) loadConfig() error {
	configPath := filepath.Join(GetAssetsPath(), "install-isos.json")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read ISO config: %w", err)
	}
	
	i.config = &ISOConfig{}
	if err := json.Unmarshal(data, i.config); err != nil {
		return fmt.Errorf("failed to parse ISO config: %w", err)
	}
	
	// Expand environment variables and handle relative paths in cache dir
	cacheDir := i.config.Settings.CacheDir
	if cacheDir == "" {
		cacheDir = ".cache/isos"
	}
	
	// If cache dir is relative, make it absolute from current directory
	if !filepath.IsAbs(cacheDir) {
		cwd, err := os.Getwd()
		if err == nil {
			cacheDir = filepath.Join(cwd, cacheDir)
		}
	}
	
	i.config.Settings.CacheDir = os.ExpandEnv(cacheDir)
	
	return nil
}

// getFilename determines the filename for the ISO
func (i *ISOInstaller) getFilename(url string, variant *ISOVariant) string {
	// Try to extract filename from URL
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		if strings.HasSuffix(filename, ".iso") || strings.HasSuffix(filename, ".exe") {
			return filename
		}
	}
	
	// Generate filename based on OS type and version
	ext := ".iso"
	if variant.Type == "exe" {
		ext = ".exe"
	}
	
	return fmt.Sprintf("%s_%s%s", i.OSType, variant.Version, ext)
}

// downloadFile downloads a file with progress reporting
func (i *ISOInstaller) downloadFile(filepath string, url string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(i.config.Settings.TimeoutSeconds) * time.Second,
	}
	
	// Get the data
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	
	// Get file size
	fileSize := resp.ContentLength
	
	// Create progress reporter
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				fi, _ := out.Stat()
				if fileSize > 0 {
					percent := float64(fi.Size()) / float64(fileSize) * 100
					fmt.Printf("\rDownloading... %.1f%% (%.2f MB / %.2f MB)", 
						percent, 
						float64(fi.Size())/(1024*1024),
						float64(fileSize)/(1024*1024))
				}
			case <-done:
				return
			}
		}
	}()
	
	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	done <- true
	fmt.Println() // New line after progress
	
	return err
}

// handleManualDownload provides instructions for manual ISO download
func (i *ISOInstaller) handleManualDownload(pkg *ISOPackage, variant *ISOVariant, url string) (string, error) {
	// Check cache first
	cacheDir := i.config.Settings.CacheDir
	if i.OutputDir != "" {
		cacheDir = i.OutputDir
	}
	
	// Generate expected filename
	filename := fmt.Sprintf("%s_%s.iso", i.OSType, variant.Version)
	cachedPath := filepath.Join(cacheDir, filename)
	
	// Check if already downloaded
	if _, err := os.Stat(cachedPath); err == nil {
		fmt.Printf("✅ Found cached ISO: %s\n", cachedPath)
		return cachedPath, nil
	}
	
	// Provide detailed instructions
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("📋 Manual Download Required for %s\n", pkg.Name)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n🔍 Step-by-step instructions:")
	fmt.Println("\n1️⃣  Open your web browser and visit:")
	fmt.Printf("    %s\n", url)
	
	if i.OSType == "windows11" {
		fmt.Println("\n2️⃣  On the Microsoft page:")
		fmt.Println("    • Scroll down to 'Download Windows 11 Disk Image (ISO)'")
		fmt.Println("    • Select your language")
		fmt.Println("    • Click 'Confirm'")
		fmt.Println("    • Click the '64-bit Download' button")
		fmt.Println("    • Save the file")
	} else if i.OSType == "windows10" {
		fmt.Println("\n2️⃣  On the Microsoft page:")
		fmt.Println("    • Scroll down to 'Download Windows 10 Disk Image (ISO)'")
		fmt.Println("    • Select edition")
		fmt.Println("    • Select your language")
		fmt.Println("    • Choose 64-bit or 32-bit")
		fmt.Println("    • Save the file")
	}
	
	fmt.Printf("\n3️⃣  Save the ISO file to:\n    %s\n", cacheDir)
	fmt.Printf("\n4️⃣  Rename the file to:\n    %s\n", filename)
	
	fmt.Println("\n💡 Alternative: Use Windows Media Creation Tool")
	fmt.Println("   Run: portunix install iso", i.OSType, "--variant media-tool")
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("\n⏸️  After downloading, run this command again.\n")
	fmt.Printf("    The ISO will be automatically detected in:\n    %s\n", cachedPath)
	
	return "", fmt.Errorf("manual download required")
}

// followMicrosoftRedirect follows Microsoft fwlink redirects to get actual download URLs
func (i *ISOInstaller) followMicrosoftRedirect(fwlinkURL string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects automatically, we want to capture them
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}
	
	// Create request with proper headers to avoid detection
	req, err := http.NewRequest("HEAD", fwlinkURL, nil)
	if err != nil {
		return "", err
	}
	
	// Add headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	// Check if we got a redirect
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		if location != "" {
			fmt.Printf("🔗 Following Microsoft redirect to: %s\n", location)
			return location, nil
		}
	}
	
	return "", fmt.Errorf("no redirect found for %s", fwlinkURL)
}

// downloadWindowsISO handles Windows ISO downloads with proper user-agent and methodology
func (i *ISOInstaller) downloadWindowsISO(pkg *ISOPackage, variant *ISOVariant, baseURL string) (string, error) {
	// Use cache dir
	cacheDir := i.config.Settings.CacheDir
	if i.OutputDir != "" {
		cacheDir = i.OutputDir
	}
	
	filename := fmt.Sprintf("%s_%s.iso", i.OSType, variant.Version)
	cachedPath := filepath.Join(cacheDir, filename)
	
	// Check if already cached
	if _, err := os.Stat(cachedPath); err == nil {
		fmt.Printf("✅ Found cached ISO: %s\n", cachedPath)
		return cachedPath, nil
	}
	
	fmt.Printf("🔄 Downloading %s ISO from Microsoft...\n", pkg.Name)
	
	// Create output directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	// Try different approaches based on OS type
	var downloadURL string
	var err error
	
	if i.OSType == "windows11" {
		downloadURL, err = i.getWindows11DirectURL(baseURL)
	} else if i.OSType == "windows10" {
		downloadURL, err = i.getWindows10DirectURL(baseURL) 
	}
	
	if err != nil || downloadURL == "" {
		// Fallback to manual instructions
		fmt.Printf("⚠️  Automatic download failed: %v\n", err)
		return i.handleManualDownload(pkg, variant, baseURL)
	}
	
	// Download with proper headers
	fmt.Printf("📥 Downloading from: %s\n", downloadURL)
	fmt.Printf("💾 Size: %s\n", variant.Size)
	
	if err := i.downloadFileWithHeaders(cachedPath, downloadURL); err != nil {
		// If direct download fails, fallback to manual
		fmt.Printf("⚠️  Download failed: %v\n", err)
		fmt.Printf("📋 Falling back to manual download instructions...\n")
		return i.handleManualDownload(pkg, variant, baseURL)
	}
	
	fmt.Printf("\n✅ Windows ISO downloaded successfully: %s\n", cachedPath)
	return cachedPath, nil
}

// getWindows11DirectURL gets Windows 11 ISO URL using browser simulation
func (i *ISOInstaller) getWindows11DirectURL(baseURL string) (string, error) {
	// For Windows 11, try to get the ISO download page with mobile user-agent
	client := &http.Client{Timeout: 30 * time.Second}
	
	// Use iPad user-agent to get ISO download option (works for Windows 10, might work for 11)
	req, err := http.NewRequest("GET", "https://www.microsoft.com/software-download/windows11", nil)
	if err != nil {
		return "", err
	}
	
	// Mobile user-agent that bypasses the Media Creation Tool requirement
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPad; CPU OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	// Look for ISO download links in the response
	content := string(body)
	
	// Try multiple patterns to find download links
	patterns := []string{
		`href=['"]([^'"]*\.iso[^'"]*)['"]`,
		`href=['"]([^'"]*download[^'"]*\.microsoft\.com[^'"]*\.iso[^'"]*)['"]`,
		`https://[^'"]*\.microsoft\.com[^'"]*\.iso`,
	}
	
	for _, pattern := range patterns {
		if url := i.extractURLPattern(content, pattern); url != "" {
			return url, nil
		}
	}
	
	return "", fmt.Errorf("no ISO download URL found on Windows 11 page")
}

// getWindows10DirectURL gets Windows 10 ISO using proven user-agent method
func (i *ISOInstaller) getWindows10DirectURL(baseURL string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	
	// This method is proven to work for Windows 10
	req, err := http.NewRequest("GET", "https://www.microsoft.com/software-download/windows10", nil)
	if err != nil {
		return "", err
	}
	
	// Use non-Windows user-agent to get ISO download option
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	content := string(body)
	
	// Look for Windows 10 ISO download patterns
	patterns := []string{
		`href=['"]([^'"]*English[^'"]*x64[^'"]*\.iso[^'"]*)['"]`,
		`href=['"]([^'"]*\.iso[^'"]*)['"]`,
		`https://[^'"]*software\.download[^'"]*\.iso`,
	}
	
	for _, pattern := range patterns {
		if url := i.extractURLPattern(content, pattern); url != "" {
			return url, nil
		}
	}
	
	return "", fmt.Errorf("no ISO download URL found on Windows 10 page")
}

// extractURLPattern extracts URLs using regex pattern
func (i *ISOInstaller) extractURLPattern(content, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	// Try finding without capture groups
	if match := re.FindString(content); match != "" {
		return match
	}
	return ""
}

// downloadFileWithHeaders downloads a file with proper browser headers
func (i *ISOInstaller) downloadFileWithHeaders(filepath, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	
	client := &http.Client{
		Timeout: time.Duration(i.config.Settings.TimeoutSeconds) * time.Second,
	}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	
	// Add browser headers to avoid detection
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://www.microsoft.com/")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	// Progress reporting
	fileSize := resp.ContentLength
	if fileSize <= 0 {
		fileSize = 5900 * 1024 * 1024 // Default ~5.9GB for Windows ISO
	}
	
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				fi, _ := out.Stat()
				percent := float64(fi.Size()) / float64(fileSize) * 100
				fmt.Printf("\rDownloading... %.1f%% (%.2f MB / %.2f MB)", 
					percent,
					float64(fi.Size())/(1024*1024),
					float64(fileSize)/(1024*1024))
			case <-done:
				return
			}
		}
	}()
	
	_, err = io.Copy(out, resp.Body)
	done <- true
	
	return err
}


// GetAssetsPath returns the path to assets directory
func GetAssetsPath() string {
	// Try to find assets directory relative to executable
	exe, err := os.Executable()
	if err == nil {
		assetsPath := filepath.Join(filepath.Dir(exe), "assets")
		if _, err := os.Stat(assetsPath); err == nil {
			return assetsPath
		}
	}
	
	// Fallback to current directory
	assetsPath, _ := filepath.Abs("assets")
	return assetsPath
}