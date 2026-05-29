/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// configFileName is the name of the Proxmox config file inside the Portunix
// config directory. Full path on Linux: ~/.config/portunix/proxmox.json
const configFileName = "proxmox.json"

// configOverrideEnv lets tests (and users) redirect the config file elsewhere.
const configOverrideEnv = "PORTUNIX_PROXMOX_CONFIG"

// configPath returns the absolute path to the Proxmox config file. It honours
// the PORTUNIX_PROXMOX_CONFIG override, falling back to $UserConfigDir/portunix
// (cross-platform: XDG on Linux, AppData on Windows).
func configPath() (string, error) {
	if p := os.Getenv(configOverrideEnv); p != "" {
		return p, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(base, "portunix", configFileName), nil
}

// loadConfig reads the Proxmox config file. A missing file is not an error —
// it returns an empty Config so the first `auth login` can populate it.
func loadConfig() (*Config, string, error) {
	path, err := configPath()
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Profiles: map[string]*Profile{}}, path, nil
		}
		return nil, path, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, path, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]*Profile{}
	}
	return &cfg, path, nil
}

// saveConfig writes the config atomically (via temp file + rename) with 0600
// permissions. Token secrets live in this file, so loose perms would be a
// credential leak.
func saveConfig(cfg *Config) (string, error) {
	path, err := configPath()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return path, fmt.Errorf("create config dir %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return path, fmt.Errorf("encode config: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".proxmox-*.json.tmp")
	if err != nil {
		return path, fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	// Best-effort cleanup of the temp file if something fails between here and
	// the rename.
	defer func() {
		if _, statErr := os.Stat(tmpName); statErr == nil {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return path, fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return path, fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return path, fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return path, fmt.Errorf("rename %s -> %s: %w", tmpName, path, err)
	}
	return path, nil
}

// resolveProfile picks the profile to operate on. Empty name -> current; if
// current is also empty, the single profile is chosen if there's exactly one.
func resolveProfile(cfg *Config, name string) (string, *Profile, error) {
	if name == "" {
		name = cfg.CurrentProfile
	}
	if name == "" {
		// Single-profile convenience: if only one profile is stored, use it.
		if len(cfg.Profiles) == 1 {
			for n, p := range cfg.Profiles {
				return n, p, nil
			}
		}
		if len(cfg.Profiles) == 0 {
			return "", nil, fmt.Errorf("no profiles configured — run `portunix proxmox auth login` first")
		}
		return "", nil, fmt.Errorf("multiple profiles configured and no current selected — pass --profile NAME")
	}
	p, ok := cfg.Profiles[name]
	if !ok {
		return "", nil, fmt.Errorf("profile %q not found", name)
	}
	return name, p, nil
}
