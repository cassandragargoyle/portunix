/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package skills

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// ManifestFile is the provenance manifest stored in every skills root (ADR-043 D5)
const ManifestFile = ".portunix-skills.json"

// manifestVersion is the current manifest schema version
const manifestVersion = 1

// Manifest records the skills Portunix installed into one skills directory
type Manifest struct {
	Version int             `json:"version"`
	Skills  []ManifestEntry `json:"skills"`
}

// ManifestEntry is the provenance of one installed skill
type ManifestEntry struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"`
	Ref         string   `json:"ref"`
	Commit      string   `json:"commit"`
	InstalledAt string   `json:"installedAt"`
	Files       []string `json:"files"`
	Checksum    string   `json:"checksum"`
}

// LoadManifest reads the manifest of a skills directory; a missing file
// yields an empty manifest
func LoadManifest(skillsDir string) (*Manifest, error) {
	path := filepath.Join(skillsDir, ManifestFile)
	data, err := os.ReadFile(path) // #nosec G304 -- manifest inside a resolved skills directory
	if errors.Is(err, os.ErrNotExist) {
		return &Manifest{Version: manifestVersion}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if m.Version > manifestVersion {
		return nil, fmt.Errorf("%s has unsupported version %d (this Portunix supports %d)", path, m.Version, manifestVersion)
	}
	m.Version = manifestVersion
	return &m, nil
}

// Save writes the manifest atomically; an empty manifest removes the file
func (m *Manifest) Save(skillsDir string) error {
	path := filepath.Join(skillsDir, ManifestFile)
	if len(m.Skills) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", path, err)
		}
		return nil
	}
	sort.Slice(m.Skills, func(i, j int) bool { return m.Skills[i].Name < m.Skills[j].Name })
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(skillsDir, 0o755); err != nil { // #nosec G301 -- skills directories are regular project files
		return fmt.Errorf("create %s: %w", skillsDir, err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil { // #nosec G306 -- provenance manifest is not sensitive
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// Find returns the entry for a skill name
func (m *Manifest) Find(name string) *ManifestEntry {
	for i := range m.Skills {
		if m.Skills[i].Name == name {
			return &m.Skills[i]
		}
	}
	return nil
}

// Upsert adds or replaces the entry with the same name
func (m *Manifest) Upsert(e ManifestEntry) {
	if existing := m.Find(e.Name); existing != nil {
		*existing = e
		return
	}
	m.Skills = append(m.Skills, e)
}

// Remove drops the entry for a skill name and reports whether it existed
func (m *Manifest) Remove(name string) bool {
	for i := range m.Skills {
		if m.Skills[i].Name == name {
			m.Skills = append(m.Skills[:i], m.Skills[i+1:]...)
			return true
		}
	}
	return false
}

// HasSource reports whether any entry was installed from the given source
func (m *Manifest) HasSource(source string) bool {
	for _, e := range m.Skills {
		if e.Source == source {
			return true
		}
	}
	return false
}

// TreeChecksum hashes a set of files below dir: sha256 over the sorted
// "<path>\x00<sha256 of content>\n" lines
func TreeChecksum(dir string, files []string) (string, error) {
	sorted := append([]string(nil), files...)
	sort.Strings(sorted)
	h := sha256.New()
	for _, rel := range sorted {
		sum, err := fileSHA256(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s\x00%s\n", rel, sum)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// DirChecksum hashes every regular file currently present in dir
func DirChecksum(dir string) (string, error) {
	files, err := listSkillFiles(dir)
	if err != nil {
		return "", err
	}
	return TreeChecksum(dir, files)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path) // #nosec G304 -- file of a skill tree
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash %s: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
