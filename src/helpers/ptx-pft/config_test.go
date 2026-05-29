/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveProjectPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd failed: %v", err)
	}

	// Use OS-appropriate absolute paths so the test runs on both Linux and Windows
	absConfigDir := filepath.Join(cwd, "test-config-dir")
	absProjectDir := filepath.Join(cwd, "test-project-dir")
	absExplicit := filepath.Join(cwd, "test-explicit")

	configFile := filepath.Join(absConfigDir, ConfigFileName)

	tests := []struct {
		name           string
		configPath     string
		configFilePath string
		explicitPath   string
		want           string
	}{
		{
			name:           "explicit path overrides everything",
			configPath:     "/some/stored/path",
			configFilePath: configFile,
			explicitPath:   absExplicit,
			want:           absExplicit,
		},
		{
			name:           "explicit path overrides empty config path",
			configPath:     "",
			configFilePath: configFile,
			explicitPath:   absExplicit,
			want:           absExplicit,
		},
		{
			name:           "empty config path uses config file directory",
			configPath:     "",
			configFilePath: configFile,
			explicitPath:   "",
			want:           absConfigDir,
		},
		{
			name:           "relative config path resolved from config dir",
			configPath:     "subdir",
			configFilePath: configFile,
			explicitPath:   "",
			want:           filepath.Join(absConfigDir, "subdir"),
		},
		{
			name:           "relative parent config path resolved from config dir",
			configPath:     filepath.Join("..", "sibling"),
			configFilePath: configFile,
			explicitPath:   "",
			want:           filepath.Join(absConfigDir, "..", "sibling"),
		},
		{
			name:           "absolute config path used as-is",
			configPath:     absProjectDir,
			configFilePath: configFile,
			explicitPath:   "",
			want:           absProjectDir,
		},
		{
			name:           "empty config path with no config file falls back to cwd",
			configPath:     "",
			configFilePath: "",
			explicitPath:   "",
			want:           cwd,
		},
		{
			name:           "relative config path with no config file falls back to cwd",
			configPath:     "subdir",
			configFilePath: "",
			explicitPath:   "",
			want:           filepath.Join(cwd, "subdir"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Path: tt.configPath}
			got := ResolveProjectPath(cfg, tt.configFilePath, tt.explicitPath)
			if got != tt.want {
				t.Errorf("ResolveProjectPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestResolveProjectPath_ForeignAbsolutePath verifies that --path always wins
// over a stored absolute path. This is the core cross-platform scenario from
// issue #134: a config committed on Linux contains a Linux absolute path and
// is checked out on Windows (or vice versa). Without --path the stored path
// is returned as-is; with --path the explicit path takes precedence.
func TestResolveProjectPath_ForeignAbsolutePath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd failed: %v", err)
	}

	// Use an absolute path that is recognised as absolute on the host OS so
	// the test exercises the same code path on Linux and Windows. The path
	// itself is fictitious and not read from disk.
	var storedAbsPath string
	if runtime.GOOS == "windows" {
		storedAbsPath = `C:\stored\project`
	} else {
		storedAbsPath = "/stored/project"
	}

	cfg := &Config{Path: storedAbsPath}
	configFilePath := filepath.Join(cwd, "current-os-config", ConfigFileName)
	override := filepath.Join(cwd, "override-project")

	// Without --path: the stored absolute path is returned as-is.
	if got := ResolveProjectPath(cfg, configFilePath, ""); got != storedAbsPath {
		t.Errorf("without --path: ResolveProjectPath() = %q, want %q", got, storedAbsPath)
	}

	// With --path: the explicit path overrides the stored path. This is what
	// makes a config from a different OS usable.
	if got := ResolveProjectPath(cfg, configFilePath, override); got != override {
		t.Errorf("with --path: ResolveProjectPath() = %q, want %q", got, override)
	}
}

// TestResolveProjectPath_EmptyConfigCrossPlatform verifies that a config with
// empty path field works on any OS — the recommended cross-platform setup.
func TestResolveProjectPath_EmptyConfigCrossPlatform(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd failed: %v", err)
	}

	// Simulate config file checked into git, used on different OS
	configDir := filepath.Join(cwd, "shared-project")
	configFilePath := filepath.Join(configDir, ConfigFileName)

	cfg := &Config{Path: ""}
	got := ResolveProjectPath(cfg, configFilePath, "")
	if got != configDir {
		t.Errorf("ResolveProjectPath() = %q, want %q (config dir)", got, configDir)
	}
}

// TestResolveProjectPath_ExplicitPathConvertedToAbsolute verifies that a
// relative --path argument is converted to an absolute path so callers do not
// need to handle both forms.
func TestResolveProjectPath_ExplicitPathConvertedToAbsolute(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd failed: %v", err)
	}

	cfg := &Config{Path: "/should/be/ignored"}
	got := ResolveProjectPath(cfg, "/anywhere/.pft-config.json", "relative/path")

	want := filepath.Join(cwd, "relative", "path")
	if got != want {
		t.Errorf("ResolveProjectPath() = %q, want %q", got, want)
	}
}
