/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// PythonInstall describes a Python interpreter discovered on the system.
type PythonInstall struct {
	Executable string `json:"executable"`
	Version    string `json:"version"`
	Source     string `json:"source"` // "PATH" or "portunix"
}

// VersionManager handles Python version discovery and selection.
type VersionManager struct{}

// NewVersionManager creates a new version manager.
func NewVersionManager() *VersionManager {
	return &VersionManager{}
}

// ListInstalls discovers Python interpreters available to the user. It checks
// the most common executable names (python3, python, python3.x for x=8..13)
// in PATH and the Portunix-managed Python install dir
// (~/.portunix/python/installs/). Duplicates (same executable path) are
// removed. Order is stable: PATH-sourced first, then portunix.
func (vm *VersionManager) ListInstalls() ([]*PythonInstall, error) {
	var installs []*PythonInstall
	seen := make(map[string]bool)

	// 1) Common names in PATH
	candidates := []string{"python3", "python"}
	for minor := 8; minor <= 13; minor++ {
		candidates = append(candidates, fmt.Sprintf("python3.%d", minor))
		if runtime.GOOS == "windows" {
			candidates = append(candidates, fmt.Sprintf("python3.%d.exe", minor))
		}
	}
	for _, name := range candidates {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		// Resolve symlinks so we deduplicate properly.
		resolved, _ := filepath.EvalSymlinks(path)
		if resolved == "" {
			resolved = path
		}
		if seen[resolved] {
			continue
		}
		seen[resolved] = true
		ver, err := getPythonVersionForExe(path)
		// Skip stubs (e.g. Windows Store python3.exe) that fail to report a
		// version — they are not usable Python interpreters.
		if err != nil || ver == "" {
			continue
		}
		installs = append(installs, &PythonInstall{
			Executable: path,
			Version:    ver,
			Source:     "PATH",
		})
	}

	// 2) Portunix-managed Python installs (~/.portunix/python/installs/*)
	if home, err := os.UserHomeDir(); err == nil {
		installsDir := filepath.Join(home, ".portunix", "python", "installs")
		entries, err := os.ReadDir(installsDir)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				exe := filepath.Join(installsDir, e.Name(), "bin", "python3")
				if runtime.GOOS == "windows" {
					exe = filepath.Join(installsDir, e.Name(), "python.exe")
				}
				if _, err := os.Stat(exe); err != nil {
					continue
				}
				resolved, _ := filepath.EvalSymlinks(exe)
				if resolved == "" {
					resolved = exe
				}
				if seen[resolved] {
					continue
				}
				seen[resolved] = true
				ver, verErr := getPythonVersionForExe(exe)
				if verErr != nil || ver == "" {
					continue
				}
				installs = append(installs, &PythonInstall{
					Executable: exe,
					Version:    ver,
					Source:     "portunix",
				})
			}
		}
	}

	return installs, nil
}

// DetectProjectVersion looks for a project Python version requirement.
// Search order:
//  1. .python-version (pyenv-compatible, single line: "3.11" or "3.11.5")
//  2. pyproject.toml under [project] requires-python or [tool.poetry] python
//
// Returns ("", nil) if no project version is configured (not an error).
func (vm *VersionManager) DetectProjectVersion() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	pvFile := filepath.Join(cwd, ".python-version")
	if data, err := os.ReadFile(pvFile); err == nil {
		ver := strings.TrimSpace(string(data))
		if ver != "" {
			return ver, nil
		}
	}

	pyproject := filepath.Join(cwd, "pyproject.toml")
	f, err := os.Open(pyproject)
	if err != nil {
		return "", nil
	}
	defer f.Close()

	requiresRe := regexp.MustCompile(`(?i)^\s*requires-python\s*=\s*["']([^"']+)["']`)
	poetryRe := regexp.MustCompile(`(?i)^\s*python\s*=\s*["']([^"']+)["']`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if m := requiresRe.FindStringSubmatch(line); len(m) == 2 {
			return strings.TrimSpace(m[1]), nil
		}
		if m := poetryRe.FindStringSubmatch(line); len(m) == 2 {
			return strings.TrimSpace(m[1]), nil
		}
	}
	return "", nil
}

// WriteProjectVersion writes a .python-version file in the current directory.
// This is the safest cross-platform way to "switch" Python without mutating
// the global PATH; pyenv, uv, and a number of other tools honor this file.
func (vm *VersionManager) WriteProjectVersion(version string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	target := filepath.Join(cwd, ".python-version")
	if err := os.WriteFile(target, []byte(version+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write .python-version: %v", err)
	}
	return nil
}

// getPythonVersionForExe returns "3.11.5" given a python executable path.
// Returns ("", err) when the executable cannot be run.
func getPythonVersionForExe(exe string) (string, error) {
	cmd := exec.Command(exe, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(string(out))
	v = strings.TrimPrefix(v, "Python ")
	return v, nil
}
