/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package skills

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Scope selects where skills are installed (ADR-043 D4)
type Scope int

const (
	ScopeProject Scope = iota
	ScopeUser
)

func (s Scope) String() string {
	if s == ScopeUser {
		return "user"
	}
	return "project"
}

// ResolvedSkill is a discovered skill ready to be installed
type ResolvedSkill struct {
	Name        string
	Description string
	// SourceDir is the absolute skill directory inside the cache or --source tree
	SourceDir string
	// Files are the skill files relative to SourceDir (slash separated)
	Files []string
}

// AgentTarget is one install destination: the universal directory or an
// agent-specific skills directory (ADR-043 D3)
type AgentTarget interface {
	ID() string
	DisplayName() string
	Detect() (installed bool, err error)
	SkillsDir(scope Scope) (string, error)
	Install(s ResolvedSkill, scope Scope, force bool) (written []string, err error)
}

// UniversalID is the target ID of the shared .agents/skills directory
const UniversalID = "universal"

// rootKind selects the directory a user-scope path is relative to
type rootKind int

const (
	rootHome   rootKind = iota // the user's home directory
	rootConfig                 // $XDG_CONFIG_HOME, falling back to ~/.config
)

// userPath is a path below the home or config directory
type userPath struct {
	root rootKind
	rel  string
}

func (p userPath) resolve() (string, error) {
	var base string
	switch p.root {
	case rootConfig:
		dir, err := configHome()
		if err != nil {
			return "", err
		}
		base = dir
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		base = home
	}
	return filepath.Join(base, filepath.FromSlash(p.rel)), nil
}

// dirTarget is an AgentTarget whose skills live in <projectDir>/<name>/
// below the working directory (project scope) or in <userDir>/<name>/
// (user scope). Layouts follow the `skills` CLI agent table (skills@1.5.9).
type dirTarget struct {
	id          string
	displayName string
	projectDir  string
	userDir     userPath
	// userEnv names an environment variable holding the agent's config
	// root; when set, user-scope skills go to $userEnv/skills
	userEnv string
	// detectPackage is the registry package used for Issue #035 detection
	detectPackage string
	// detectPaths mark the agent as present when any of them exists
	detectPaths []userPath
}

func (t *dirTarget) ID() string          { return t.id }
func (t *dirTarget) DisplayName() string { return t.displayName }

// Detect reports the agent as present when its registry package is
// installed or one of its configuration directories exists. A target
// without detection rules (universal) is always present.
func (t *dirTarget) Detect() (bool, error) {
	if t.detectPackage == "" && len(t.detectPaths) == 0 {
		return true, nil
	}
	if t.detectPackage != "" && detector != nil && detector(t.detectPackage) {
		return true, nil
	}
	for _, p := range t.detectPaths {
		dir, err := p.resolve()
		if err != nil {
			continue
		}
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return true, nil
		}
	}
	return false, nil
}

func (t *dirTarget) SkillsDir(scope Scope) (string, error) {
	if scope == ScopeUser {
		if t.userEnv != "" {
			if root := strings.TrimSpace(os.Getenv(t.userEnv)); root != "" {
				return filepath.Join(root, "skills"), nil
			}
		}
		return t.userDir.resolve()
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	return filepath.Join(wd, filepath.FromSlash(t.projectDir)), nil
}

func (t *dirTarget) Install(s ResolvedSkill, scope Scope, force bool) ([]string, error) {
	dir, err := t.SkillsDir(scope)
	if err != nil {
		return nil, err
	}
	dest := filepath.Join(dir, s.Name)
	if _, err := os.Stat(dest); err == nil && !force {
		return nil, fmt.Errorf("%s already exists (use --force to overwrite)", dest)
	}
	if err := writeSkillTree(s, dest); err != nil {
		return nil, err
	}
	return s.Files, nil
}

// detector reports whether the agent behind a registry package is installed.
// It is injected by the command layer so tests do not need the package registry.
var detector func(pkg string) bool

// SetDetector installs the agent detection callback (Issue #035 reuse)
func SetDetector(fn func(pkg string) bool) { detector = fn }

// configHome returns $XDG_CONFIG_HOME or ~/.config on every platform, the
// same convention the `skills` CLI uses for Amp, OpenCode and Zed
func configHome() (string, error) {
	if dir := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".config"), nil
}

// TargetDir is one skills directory together with every target resolving
// to it; agents sharing .agents/skills are written (and listed) once
type TargetDir struct {
	Dir     string
	Targets []AgentTarget
}

// IDs returns the IDs of the targets sharing the directory
func (d TargetDir) IDs() []string {
	ids := make([]string, len(d.Targets))
	for i, t := range d.Targets {
		ids[i] = t.ID()
	}
	return ids
}

// Label is the comma separated list of target IDs
func (d TargetDir) Label() string { return strings.Join(d.IDs(), ",") }

// GroupByDir resolves the targets in a scope and merges those sharing a
// directory, keeping the order of first appearance
func GroupByDir(targets []AgentTarget, scope Scope) ([]TargetDir, error) {
	var out []TargetDir
	index := map[string]int{}
	for _, t := range targets {
		dir, err := t.SkillsDir(scope)
		if err != nil {
			return nil, err
		}
		key := filepath.Clean(dir)
		if i, ok := index[key]; ok {
			out[i].Targets = append(out[i].Targets, t)
			continue
		}
		index[key] = len(out)
		out = append(out, TargetDir{Dir: dir, Targets: []AgentTarget{t}})
	}
	return out, nil
}

// writeSkillTree copies the skill files into dest atomically: the tree is
// staged in a sibling directory and swapped in with renames
func writeSkillTree(s ResolvedSkill, dest string) error {
	parent := filepath.Dir(dest)
	if err := os.MkdirAll(parent, 0o755); err != nil { // #nosec G301 -- skills directories are regular project files
		return fmt.Errorf("create %s: %w", parent, err)
	}
	staging := dest + ".portunix-new"
	backup := dest + ".portunix-old"
	_ = os.RemoveAll(staging)
	_ = os.RemoveAll(backup)
	for _, rel := range s.Files {
		if err := copyFile(filepath.Join(s.SourceDir, filepath.FromSlash(rel)), filepath.Join(staging, filepath.FromSlash(rel))); err != nil {
			_ = os.RemoveAll(staging)
			return err
		}
	}
	hadOld := false
	if _, err := os.Lstat(dest); err == nil {
		if err := os.Rename(dest, backup); err != nil {
			_ = os.RemoveAll(staging)
			return fmt.Errorf("replace %s: %w", dest, err)
		}
		hadOld = true
	}
	if err := os.Rename(staging, dest); err != nil {
		if hadOld {
			_ = os.Rename(backup, dest)
		}
		_ = os.RemoveAll(staging)
		return fmt.Errorf("install %s: %w", dest, err)
	}
	_ = os.RemoveAll(backup)
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { // #nosec G301 -- skills directories are regular project files
		return fmt.Errorf("create %s: %w", filepath.Dir(dst), err)
	}
	in, err := os.Open(src) // #nosec G304 -- file listed by skill discovery
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644) // #nosec G302 G304 -- skill files are regular readable project files
	if err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return fmt.Errorf("write %s: %w", dst, err)
	}
	return out.Close()
}
