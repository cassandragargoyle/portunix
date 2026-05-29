/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package installconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestLoadDefaultConfig(t *testing.T) {
	cfg, err := loadDefaultConfig()
	if err != nil {
		t.Fatalf("loadDefaultConfig() unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("loadDefaultConfig() returned nil config")
	}
	if cfg.Version != "1.0" {
		t.Errorf("default Version = %q, want %q", cfg.Version, "1.0")
	}
	if cfg.Packages == nil {
		t.Error("default Packages map is nil; expected initialized empty map")
	}
	if cfg.Presets == nil {
		t.Error("default Presets map is nil; expected initialized empty map")
	}
	if len(cfg.Packages) != 0 {
		t.Errorf("default Packages should be empty, got %d entries", len(cfg.Packages))
	}
	if len(cfg.Presets) != 0 {
		t.Errorf("default Presets should be empty, got %d entries", len(cfg.Presets))
	}
}

func TestMergeConfigs_UserOverridesDefault(t *testing.T) {
	defaults := &InstallConfig{
		Version:  "1.0",
		Packages: map[string]PackageConfig{"python": {Name: "python", Description: "default"}},
		Presets:  map[string]PresetConfig{"default": {Name: "default", Description: "default preset"}},
	}
	user := &InstallConfig{
		Packages: map[string]PackageConfig{
			"python": {Name: "python", Description: "user override"},
			"node":   {Name: "node", Description: "added by user"},
		},
		Presets: map[string]PresetConfig{
			"custom": {Name: "custom", Description: "user preset"},
		},
	}

	merged := mergeConfigs(defaults, user)

	if merged.Packages["python"].Description != "user override" {
		t.Errorf("user package did not override default; got %q",
			merged.Packages["python"].Description)
	}
	if _, ok := merged.Packages["node"]; !ok {
		t.Error("user-added package 'node' missing after merge")
	}
	if _, ok := merged.Presets["default"]; !ok {
		t.Error("default preset 'default' missing after merge")
	}
	if _, ok := merged.Presets["custom"]; !ok {
		t.Error("user-added preset 'custom' missing after merge")
	}
}

func TestLoadInstallConfig_NoUserOverlay(t *testing.T) {
	withTempHome(t, "")

	cfg, err := LoadInstallConfig()
	if err != nil {
		t.Fatalf("LoadInstallConfig() unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadInstallConfig() returned nil")
	}
	if cfg.Version != "1.0" {
		t.Errorf("Version = %q, want %q", cfg.Version, "1.0")
	}
	if len(cfg.Packages) != 0 {
		t.Errorf("Packages should be empty without user overlay, got %d", len(cfg.Packages))
	}
}

func TestLoadInstallConfig_WithUserOverlay(t *testing.T) {
	home := t.TempDir()
	withTempHome(t, home)

	overlay := &InstallConfig{
		Version: "1.0",
		Packages: map[string]PackageConfig{
			"go": {Name: "go", Description: "user-defined go"},
		},
		Presets: map[string]PresetConfig{
			"mine": {Name: "mine", Description: "personal preset"},
		},
	}
	writeUserConfig(t, home, overlay)

	cfg, err := LoadInstallConfig()
	if err != nil {
		t.Fatalf("LoadInstallConfig() unexpected error: %v", err)
	}

	if pkg, ok := cfg.Packages["go"]; !ok {
		t.Error("user-defined package 'go' missing")
	} else if pkg.Description != "user-defined go" {
		t.Errorf("Description = %q, want %q", pkg.Description, "user-defined go")
	}
	if _, ok := cfg.Presets["mine"]; !ok {
		t.Error("user-defined preset 'mine' missing")
	}
}

func TestLoadInstallConfig_MalformedUserConfigFallsBackToDefault(t *testing.T) {
	home := t.TempDir()
	withTempHome(t, home)

	configDir := filepath.Join(home, ".portunix")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "install-config.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Malformed user config is treated as missing; defaults should be returned.
	cfg, err := LoadInstallConfig()
	if err != nil {
		t.Fatalf("LoadInstallConfig() unexpected error: %v", err)
	}
	if len(cfg.Packages) != 0 {
		t.Errorf("expected empty Packages on malformed overlay, got %d", len(cfg.Packages))
	}
}

func TestVariantConfig_GetDistributionsList(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{
			name:     "nil distributions",
			input:    nil,
			expected: nil,
		},
		{
			name:     "list of strings (legacy)",
			input:    []interface{}{"ubuntu", "debian"},
			expected: []string{"ubuntu", "debian"},
		},
		{
			name:     "map keyed by manager (new)",
			input:    map[string]interface{}{"apt": nil, "dnf": nil},
			expected: []string{"apt", "dnf"},
		},
		{
			name:     "list with non-string entry — non-string becomes empty",
			input:    []interface{}{"ubuntu", 42},
			expected: []string{"ubuntu", ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := &VariantConfig{Distributions: tc.input}
			got := v.GetDistributionsList()

			// Order is undefined for map case; sort both sides for comparison.
			sort.Strings(got)
			expected := append([]string(nil), tc.expected...)
			sort.Strings(expected)

			if !reflect.DeepEqual(got, expected) {
				t.Errorf("got %v, want %v", got, expected)
			}
		})
	}
}

func TestInstallConfig_JSONRoundTrip(t *testing.T) {
	cfg := &InstallConfig{
		Version: "1.0",
		Packages: map[string]PackageConfig{
			"python": {
				Name:           "python",
				Description:    "Python interpreter",
				DefaultVariant: "default",
				Platforms: map[string]PlatformConfig{
					"linux": {
						Type: "apt",
						Variants: map[string]VariantConfig{
							"default": {
								Version:  "3.13",
								Packages: []string{"python3"},
							},
						},
					},
				},
			},
		},
		Presets: map[string]PresetConfig{
			"default": {
				Name:        "default",
				Description: "default preset",
				Packages:    []PresetPackageConfig{{Name: "python", Variant: "default"}},
			},
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var back InstallConfig
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if back.Version != cfg.Version {
		t.Errorf("Version mismatch after round-trip: %q vs %q", back.Version, cfg.Version)
	}
	if back.Packages["python"].Platforms["linux"].Variants["default"].Version != "3.13" {
		t.Errorf("Variant version lost in round-trip")
	}
	if back.Presets["default"].Packages[0].Name != "python" {
		t.Errorf("Preset package name lost in round-trip")
	}
}

// withTempHome redirects os.UserHomeDir() to home for the duration of the
// test by setting HOME on Unix and USERPROFILE on Windows. Restoration is
// registered via t.Cleanup.
func withTempHome(t *testing.T, home string) {
	t.Helper()

	prevHome, hadHome := os.LookupEnv("HOME")
	prevUserProfile, hadUserProfile := os.LookupEnv("USERPROFILE")

	t.Cleanup(func() {
		if hadHome {
			_ = os.Setenv("HOME", prevHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
		if hadUserProfile {
			_ = os.Setenv("USERPROFILE", prevUserProfile)
		} else {
			_ = os.Unsetenv("USERPROFILE")
		}
	})

	if home == "" {
		home = t.TempDir()
	}
	_ = os.Setenv("HOME", home)
	_ = os.Setenv("USERPROFILE", home)
}

func writeUserConfig(t *testing.T, home string, cfg *InstallConfig) {
	t.Helper()

	configDir := filepath.Join(home, ".portunix")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal user config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "install-config.json"), data, 0o600); err != nil {
		t.Fatalf("WriteFile user config: %v", err)
	}
}
