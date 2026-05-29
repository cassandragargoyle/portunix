/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
// Package profiles implements ADR-040 D4: profile resolution with
// precedence user > org > bundled. Phase 1 only verifies that the named
// profile resolves at one of the three tiers; the upstream PM-agent
// preflight reads the profile content at agent runtime.
package profiles

import (
	"fmt"
	"os"
	"path/filepath"
)

// Resolve checks whether a profile of the given name exists at any of the
// supported tiers and returns its absolute path. It searches:
//
//  1. user      ~/.config/portunix/specpm/profiles/<name>/
//  2. org       <targetDir>/.specpm/profiles/<name>/    (also covers bundled — fetched via kit)
//
// Returns the first match. The "bundled" tier is implemented inside the
// org tier today: bundled profiles arrive in .specpm/profiles/ as part of
// the kit fetch, so a single org-tier check covers both. (See ADR-040 D4
// Revisions / variant A for the reasoning.)
func Resolve(targetDir, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty profile name")
	}
	if userRoot, err := userProfilesRoot(); err == nil {
		candidate := filepath.Join(userRoot, name)
		if isDir(candidate) {
			return candidate, nil
		}
	}
	candidate := filepath.Join(targetDir, ".specpm", "profiles", name)
	if isDir(candidate) {
		return candidate, nil
	}
	return "", fmt.Errorf("profile %q not found at user (~/.config/portunix/specpm/profiles/%s) or in .specpm/profiles/%s (kit-bundled)", name, name, name)
}

func userProfilesRoot() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfg, "portunix", "specpm", "profiles"), nil
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
