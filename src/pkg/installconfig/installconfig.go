/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package installconfig provides shared installation configuration types
// and loading helpers used by both the main portunix binary and helper
// binaries (e.g. docker/podman run-in-container commands need package and
// preset definitions to provision containers without pulling in the full
// install engine).
//
// Related: ADR-026 (Shared Platform Utilities), Issue #186c.
package installconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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
	Prerequisites  []string                  `json:"prerequisites,omitempty"`
	Platforms      map[string]PlatformConfig `json:"platforms"`
	DefaultVariant string                    `json:"default_variant"`
}

// PlatformConfig represents configuration for a specific platform (OS)
type PlatformConfig struct {
	Type         string                   `json:"type"` // msi, exe, zip, tar.gz, deb, apt, dnf, pacman, snap, repository, powershell
	Variants     map[string]VariantConfig `json:"variants"`
	InstallArgs  []string                 `json:"install_args,omitempty"`
	Verification VerificationConfig       `json:"verification,omitempty"`
	Environment  map[string]string        `json:"environment,omitempty"`
}

// VariantConfig represents a specific variant of a package
type VariantConfig struct {
	Version                string                `json:"version"`
	Type                   string                `json:"type,omitempty"`           // override platform type for this variant
	URL                    string                `json:"url,omitempty"`            // single URL for direct_download type
	URLs                   map[string]string     `json:"urls,omitempty"`           // arch -> url
	Packages               []string              `json:"packages,omitempty"`       // for apt/snap packages
	InstallScript          string                `json:"install_script,omitempty"` // for powershell/script installs
	InstallPath            string                `json:"install_path,omitempty"`
	ExtractTo              string                `json:"extract_to,omitempty"`
	Extract                bool                  `json:"extract,omitempty"`       // whether to extract archive for direct_download
	Binary                 string                `json:"binary,omitempty"`        // binary name to install for direct_download
	RequiresSudo           bool                  `json:"requires_sudo,omitempty"` // whether installation requires sudo
	PostInstall            []string              `json:"post_install,omitempty"`
	InstallArgs            []string              `json:"install_args,omitempty"`
	Distributions          interface{}           `json:"distributions,omitempty"`            // supported Linux distributions ([]string or map[string]interface{})
	SupportedVersions      []string              `json:"supported_versions,omitempty"`       // legacy explicit version support
	SupportedVersionRanges []VersionRange        `json:"supported_version_ranges,omitempty"` // new version range support
	FallbackVariants       []string              `json:"fallback_variants,omitempty"`        // fallback variants to try
	FallbackStrategy       FallbackStrategy      `json:"fallback_strategy,omitempty"`        // fallback strategy
	VersionSupportPolicy   *VersionSupportPolicy `json:"version_support_policy,omitempty"`   // version support policy
	RepositorySetup        []string              `json:"repository_setup,omitempty"`         // commands to setup repository
	RedirectTo             string                `json:"redirect_to,omitempty"`              // package to redirect to for redirect type
	DefaultVariant         string                `json:"default_variant,omitempty"`          // default variant for redirect target
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

// VersionSupportPolicy defines how version support is handled
type VersionSupportPolicy struct {
	ForwardCompatibility bool   `json:"forward_compatibility"` // Allow newer versions within range
	TestingRequirement   string `json:"testing_requirement"`   // none_for_interim, explicit, etc.
	MaintenanceSchedule  string `json:"maintenance_schedule"`  // quarterly, monthly, etc.
}

// VersionRange represents a range of supported versions
type VersionRange struct {
	Min        string `json:"min"`
	Max        string `json:"max"`
	Type       string `json:"type"`                 // lts_and_interim, lts_only, etc.
	Confidence string `json:"confidence,omitempty"` // high, medium, low
}

// FallbackStrategy defines how fallback should be handled
type FallbackStrategy string

const (
	FallbackAuto        FallbackStrategy = "auto"                   // Automatic fallback without confirmation
	FallbackAutoConfirm FallbackStrategy = "auto_with_confirmation" // Automatic with user confirmation
	FallbackManual      FallbackStrategy = "manual"                 // Manual fallback only
	FallbackDisabled    FallbackStrategy = "disabled"               // No fallback
)

// GetDistributionsList returns distributions as []string regardless of JSON format
func (v *VariantConfig) GetDistributionsList() []string {
	if v.Distributions == nil {
		return nil
	}

	// Handle []string format (old format)
	if distList, ok := v.Distributions.([]interface{}); ok {
		result := make([]string, len(distList))
		for i, dist := range distList {
			if distStr, ok := dist.(string); ok {
				result[i] = distStr
			}
		}
		return result
	}

	// Handle map[string]interface{} format (new format with apt/dnf keys)
	if distMap, ok := v.Distributions.(map[string]interface{}); ok {
		var result []string
		for key := range distMap {
			result = append(result, key)
		}
		return result
	}

	return nil
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
	// Since the embedded install-packages.json was removed, create a minimal config.
	// The registry-based system (LoadPackageRegistry) handles individual package files.
	config := &InstallConfig{
		Version:  "1.0",
		Packages: make(map[string]PackageConfig),
		Presets:  make(map[string]PresetConfig),
	}

	return config, nil
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
