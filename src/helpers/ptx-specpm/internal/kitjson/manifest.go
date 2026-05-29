/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
// Package kitjson reads and writes .specpm/kit.json — the per-project
// manifest that records what 'init' actually did, so 'check' and a future
// 'upgrade' can reason about the project's state.
package kitjson

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manifest is the on-disk form of .specpm/kit.json.
type Manifest struct {
	Source      string    `json:"source"`            // "git" | "<absolute-path>"
	Ref         string    `json:"ref"`               // pinned-ref string or "(local path)"
	Profile     string    `json:"profile,omitempty"` // resolved profile name, if any
	Integration string    `json:"integration,omitempty"`
	Skills      bool      `json:"skills,omitempty"`
	WrittenAt   time.Time `json:"written_at"`
}

// Write persists m to path, creating parent directories as needed.
func Write(path string, m Manifest) error {
	if m.WrittenAt.IsZero() {
		m.WrittenAt = time.Now().UTC()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

// Read loads a manifest previously written by Write.
func Read(path string) (Manifest, error) {
	var m Manifest
	data, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, fmt.Errorf("parse %s: %w", path, err)
	}
	return m, nil
}
