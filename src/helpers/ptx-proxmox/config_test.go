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

// withTempConfig redirects the config file to a temp path for the duration of
// the test. Returns the target path so the test can inspect it.
func withTempConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "proxmox.json")
	t.Setenv(configOverrideEnv, path)
	return path
}

func TestLoadConfig_Missing(t *testing.T) {
	withTempConfig(t)
	cfg, _, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig on missing file: %v", err)
	}
	if cfg == nil || cfg.Profiles == nil {
		t.Fatal("expected empty but non-nil Profiles map")
	}
	if len(cfg.Profiles) != 0 {
		t.Fatalf("expected 0 profiles, got %d", len(cfg.Profiles))
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := withTempConfig(t)

	cfg := &Config{
		Profiles: map[string]*Profile{
			"prod": {
				Host:        "pve.example.com",
				Port:        8006,
				AuthType:    AuthTypeToken,
				TokenID:     "root@pam!ci",
				TokenSecret: "secret-value",
				VerifyTLS:   true,
			},
		},
		CurrentProfile: "prod",
	}
	saved, err := saveConfig(cfg)
	if err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	if saved != path {
		t.Fatalf("saveConfig path mismatch: %s != %s", saved, path)
	}

	// File permissions should be 0600 (credential file).
	// Skip on Windows where Go translates permissions loosely.
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat saved file: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Fatalf("expected file mode 0600, got %o", perm)
		}
	}

	loaded, _, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if loaded.CurrentProfile != "prod" {
		t.Fatalf("CurrentProfile: got %q want %q", loaded.CurrentProfile, "prod")
	}
	p, ok := loaded.Profiles["prod"]
	if !ok {
		t.Fatal("profile 'prod' missing after reload")
	}
	if p.TokenSecret != "secret-value" {
		t.Fatalf("TokenSecret: got %q want %q", p.TokenSecret, "secret-value")
	}
}

func TestResolveProfile_SinglePromoted(t *testing.T) {
	cfg := &Config{Profiles: map[string]*Profile{
		"only": {Host: "h"},
	}}
	name, p, err := resolveProfile(cfg, "")
	if err != nil {
		t.Fatalf("resolveProfile: %v", err)
	}
	if name != "only" || p.Host != "h" {
		t.Fatalf("resolveProfile mismatch: %q %+v", name, p)
	}
}

func TestResolveProfile_NoneConfigured(t *testing.T) {
	cfg := &Config{Profiles: map[string]*Profile{}}
	if _, _, err := resolveProfile(cfg, ""); err == nil {
		t.Fatal("expected error for empty config")
	}
}

func TestResolveProfile_MultipleNoCurrent(t *testing.T) {
	cfg := &Config{Profiles: map[string]*Profile{
		"a": {Host: "a"},
		"b": {Host: "b"},
	}}
	if _, _, err := resolveProfile(cfg, ""); err == nil {
		t.Fatal("expected error for ambiguous profile resolution")
	}
}

func TestResolveProfile_Named(t *testing.T) {
	cfg := &Config{Profiles: map[string]*Profile{
		"a": {Host: "a"},
		"b": {Host: "b"},
	}}
	name, p, err := resolveProfile(cfg, "b")
	if err != nil {
		t.Fatalf("resolveProfile: %v", err)
	}
	if name != "b" || p.Host != "b" {
		t.Fatalf("resolveProfile picked wrong: %q %+v", name, p)
	}
}

func TestResolveProfile_Unknown(t *testing.T) {
	cfg := &Config{Profiles: map[string]*Profile{"a": {}}}
	if _, _, err := resolveProfile(cfg, "missing"); err == nil {
		t.Fatal("expected error for unknown profile name")
	}
}

func TestMaskSecret(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "(empty)"},
		{"abcd", "****"},
		{"abcdef", "**cdef"},
		{"very-long-secret", "************cret"},
	}
	for _, c := range cases {
		if got := maskSecret(c.in); got != c.want {
			t.Errorf("maskSecret(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
