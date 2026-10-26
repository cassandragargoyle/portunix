/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package skills

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Issue #190 O4: sources confirmed once are trusted user-wide, so installing
// the same source into another project does not ask again. Per-project
// manifests keep counting as trust for the projects that recorded them.

const (
	trustFileName    = "skills-trusted-sources.json"
	trustOverrideEnv = "PORTUNIX_SKILLS_TRUST"
	trustVersion     = 1
)

// TrustedSource is one source the user confirmed (or passed --yes for)
type TrustedSource struct {
	Source    string `json:"source"`
	TrustedAt string `json:"trustedAt"`
}

type trustStore struct {
	Version int             `json:"version"`
	Sources []TrustedSource `json:"sources"`
}

// TrustFilePath returns $UserConfigDir/portunix/skills-trusted-sources.json
// (XDG on Linux, AppData on Windows); PORTUNIX_SKILLS_TRUST overrides it
func TrustFilePath() (string, error) {
	if p := os.Getenv(trustOverrideEnv); p != "" {
		return p, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(base, "portunix", trustFileName), nil
}

func loadTrustStore() (*trustStore, string, error) {
	path, err := TrustFilePath()
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(path) // #nosec G304 -- Portunix user config file
	if errors.Is(err, os.ErrNotExist) {
		return &trustStore{Version: trustVersion}, path, nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", path, err)
	}
	var st trustStore
	if err := json.Unmarshal(data, &st); err != nil {
		// O5: tell the user how to get out of the fail-closed state
		return nil, "", fmt.Errorf("parse %s: %w\n"+
			"       The trusted skill sources file is corrupt. Fix the JSON, or delete the file to reset\n"+
			"       the record; new sources will then ask for confirmation again", path, err)
	}
	if st.Version > trustVersion {
		return nil, "", fmt.Errorf("%s has unsupported version %d (this Portunix supports %d)", path, st.Version, trustVersion)
	}
	st.Version = trustVersion
	return &st, path, nil
}

func (st *trustStore) save(path string) error {
	sort.Slice(st.Sources, func(i, j int) bool { return st.Sources[i].Source < st.Sources[j].Source })
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// TrustedSources lists the user-wide trusted sources
func TrustedSources() ([]TrustedSource, error) {
	st, _, err := loadTrustStore()
	if err != nil {
		return nil, err
	}
	return st.Sources, nil
}

// IsTrusted reports whether the source is in the user-wide trust record
func IsTrusted(source string) (bool, error) {
	st, _, err := loadTrustStore()
	if err != nil {
		return false, err
	}
	for _, s := range st.Sources {
		if s.Source == source {
			return true, nil
		}
	}
	return false, nil
}

// TrustSource records the source as trusted; recording it twice is a no-op
func TrustSource(source string) error {
	st, path, err := loadTrustStore()
	if err != nil {
		return err
	}
	for _, s := range st.Sources {
		if s.Source == source {
			return nil
		}
	}
	st.Sources = append(st.Sources, TrustedSource{Source: source, TrustedAt: time.Now().UTC().Format(time.RFC3339)})
	return st.save(path)
}

// UntrustSource removes the source from the trust record and reports
// whether it was present
func UntrustSource(source string) (bool, error) {
	st, path, err := loadTrustStore()
	if err != nil {
		return false, err
	}
	for i, s := range st.Sources {
		if s.Source == source {
			st.Sources = append(st.Sources[:i], st.Sources[i+1:]...)
			return true, st.save(path)
		}
	}
	return false, nil
}
