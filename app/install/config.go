package install

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// DefaultInstallConfig holds the embedded default configuration
var DefaultInstallConfig string

// InstallConfig represents the complete installation configuration
type InstallConfig struct {
	Version  string                   `json:"version"`
	Packages map[string]PackageConfig `json:"packages"`
	Presets  map[string]PresetConfig  `json:"presets"`
}

// PackageConfig represents configuration for a single package
type PackageConfig struct {
	Name           string                    `json:"name"`
	Description    string                    `json:"description"`
	Platforms      map[string]PlatformConfig `json:"platforms"`
	DefaultVariant string                    `json:"default_variant"`
}

// PlatformConfig represents configuration for a specific platform (OS)
type PlatformConfig struct {
	Type         string                   `json:"type"` // msi, exe, zip, tar.gz, deb, apt, snap
	Variants     map[string]VariantConfig `json:"variants"`
	InstallArgs  []string                 `json:"install_args,omitempty"`
	Verification VerificationConfig       `json:"verification,omitempty"`
	Environment  map[string]string        `json:"environment,omitempty"`
}

// VariantConfig represents a specific variant of a package
type VariantConfig struct {
	Version       string            `json:"version"`
	URLs          map[string]string `json:"urls,omitempty"`           // arch -> url
	Packages      []string          `json:"packages,omitempty"`       // for apt/snap packages
	InstallScript string            `json:"install_script,omitempty"` // for powershell/script installs
	InstallPath   string            `json:"install_path,omitempty"`
	ExtractTo     string            `json:"extract_to,omitempty"`
	PostInstall   []string          `json:"post_install,omitempty"`
	InstallArgs   []string          `json:"install_args,omitempty"`
}

// VerificationConfig represents how to verify installation
type VerificationConfig struct {
	Command          string `json:"command"`
	ExpectedExitCode int    `json:"expected_exit_code"`
}

// PresetConfig represents a preset collection of packages
type PresetConfig struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Packages    []PresetPackageConfig `json:"packages"`
}

// PresetPackageConfig represents a package reference in a preset
type PresetPackageConfig struct {
	Name    string `json:"name"`
	Variant string `json:"variant"`
}

// LoadInstallConfig loads the installation configuration from embedded assets and user config
func LoadInstallConfig() (*InstallConfig, error) {
	// Load default config from embedded assets
	defaultConfig, err := loadDefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	// Try to load user config overlay
	userConfig, err := loadUserConfig()
	if err != nil {
		// User config is optional, just use default
		return defaultConfig, nil
	}

	// Merge user config with default
	mergedConfig := mergeConfigs(defaultConfig, userConfig)
	return mergedConfig, nil
}

// loadDefaultConfig loads the default configuration from embedded assets
func loadDefaultConfig() (*InstallConfig, error) {
	if DefaultInstallConfig == "" {
		return nil, fmt.Errorf("default install config not embedded")
	}

	var config InstallConfig
	if err := json.Unmarshal([]byte(DefaultInstallConfig), &config); err != nil {
		return nil, fmt.Errorf("failed to parse default config: %w", err)
	}

	return &config, nil
}

// loadUserConfig loads user configuration from ~/.portunix/install-config.json
func loadUserConfig() (*InstallConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".portunix", "install-config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("user config not found")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config InstallConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// mergeConfigs merges user configuration with default configuration
func mergeConfigs(defaultConfig, userConfig *InstallConfig) *InstallConfig {
	merged := *defaultConfig

	// Merge packages (user config overrides default)
	for name, pkg := range userConfig.Packages {
		merged.Packages[name] = pkg
	}

	// Merge presets (user config overrides default)
	for name, preset := range userConfig.Presets {
		merged.Presets[name] = preset
	}

	return &merged
}

// IsWindowsSandbox detects if the program is running in Windows Sandbox
func IsWindowsSandbox() bool {
	// Check multiple indicators that we're in Windows Sandbox

	// 1. Check for WDAGUtilityAccount user (primary indicator)
	if username := os.Getenv("USERNAME"); username == "WDAGUtilityAccount" {
		return true
	}

	// 2. Check for typical sandbox registry entries if possible
	// In sandbox, we often have limited registry access, but let's check
	cmd := exec.Command("reg", "query", "HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", "/v", "EditionID")
	output, err := cmd.Output()
	if err == nil && strings.Contains(strings.ToLower(string(output)), "core") {
		// Windows Sandbox often runs Windows Core edition
		return true
	}

	// 3. Check for sandbox-specific environment variables or paths
	if os.Getenv("WDAG_COMPUTERNAME") != "" {
		return true
	}

	// 4. Check if we're in typical sandbox temp directory structure
	if hostname, err := os.Hostname(); err == nil {
		if strings.HasPrefix(strings.ToLower(hostname), "sandbox-") ||
			strings.Contains(strings.ToLower(hostname), "wdag") {
			return true
		}
	}

	return false
}

// GetOperatingSystem returns the current operating system, with special handling for Windows Sandbox
func GetOperatingSystem() string {
	baseOS := runtime.GOOS

	// If we're on Windows, check if we're specifically in Windows Sandbox
	if baseOS == "windows" && IsWindowsSandbox() {
		return "windows_sandbox"
	}

	return baseOS
}

// GetArchitecture returns the current system architecture for URL selection
func GetArchitecture() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "386":
		return "x86"
	case "arm64":
		return "arm64"
	default:
		return "x64" // default fallback
	}
}

// ResolveVariables resolves template variables in strings
func ResolveVariables(input string, variables map[string]string) string {
	result := input
	for key, value := range variables {
		placeholder := "${" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// GetPackageInfo returns package configuration for given name and variant
func (config *InstallConfig) GetPackageInfo(packageName, variant string) (*PackageConfig, *PlatformConfig, *VariantConfig, error) {
	pkg, exists := config.Packages[packageName]
	if !exists {
		return nil, nil, nil, fmt.Errorf("package '%s' not found", packageName)
	}

	// Get current OS platform
	currentOS := GetOperatingSystem()
	platform, exists := pkg.Platforms[currentOS]

	// If sandbox-specific config doesn't exist, fallback to windows
	if !exists && currentOS == "windows_sandbox" {
		platform, exists = pkg.Platforms["windows"]
		if !exists {
			return nil, nil, nil, fmt.Errorf("package '%s' not supported on %s or windows", packageName, currentOS)
		}
	} else if !exists {
		return nil, nil, nil, fmt.Errorf("package '%s' not supported on %s", packageName, currentOS)
	}

	// Use default variant if not specified
	if variant == "" {
		variant = pkg.DefaultVariant
		// Override default variant for sandbox-specific packages
		if currentOS == "windows_sandbox" {
			// Check if sandbox-specific variant exists
			if _, exists := platform.Variants["sandbox"]; exists {
				variant = "sandbox"
			}
		}
	}

	variantConfig, exists := platform.Variants[variant]
	if !exists {
		return nil, nil, nil, fmt.Errorf("variant '%s' not found for package '%s' on %s", variant, packageName, currentOS)
	}

	return &pkg, &platform, &variantConfig, nil
}

// GetDownloadURL returns the download URL for current architecture
func (variant *VariantConfig) GetDownloadURL() (string, error) {
	arch := GetArchitecture()
	url, exists := variant.URLs[arch]
	if !exists {
		// Try x64 as fallback
		if arch != "x64" {
			url, exists = variant.URLs["x64"]
		}
		if !exists {
			return "", fmt.Errorf("no download URL found for architecture %s", arch)
		}
	}
	return url, nil
}

// GetFileName extracts filename from download URL
func (variant *VariantConfig) GetFileName() (string, error) {
	url, err := variant.GetDownloadURL()
	if err != nil {
		return "", err
	}

	// Extract filename from URL
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid URL: %s", url)
	}

	filename := parts[len(parts)-1]

	// Remove query parameters
	if idx := strings.Index(filename, "?"); idx >= 0 {
		filename = filename[:idx]
	}

	return filename, nil
}
