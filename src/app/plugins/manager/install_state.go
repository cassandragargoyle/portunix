/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"portunix.ai/app/plugins"
)

// InstallState describes whether the installed files of a plugin are usable
type InstallState string

const (
	InstallStateOK      InstallState = "ok"
	InstallStateMissing InstallState = "missing" // install directory does not exist
	InstallStateBroken  InstallState = "broken"  // files present but unusable
)

// CheckInstall inspects the installed files of the plugin without running it
// and returns the state together with a human-readable reason
func (rp *RegistryPlugin) CheckInstall() (InstallState, string) {
	if rp.InstallPath == "" {
		return InstallStateOK, ""
	}
	if _, err := os.Stat(rp.InstallPath); err != nil {
		return InstallStateMissing, fmt.Sprintf("install directory not found: %s", rp.InstallPath)
	}
	if rp.Runtime == "python" && rp.Wheel != "" {
		if err := checkPythonVenv(filepath.Join(rp.InstallPath, ".venv")); err != nil {
			return InstallStateBroken, err.Error()
		}
	}
	return InstallStateOK, ""
}

// RepairHint returns the command that fixes the given install state
func (rp *RegistryPlugin) RepairHint(state InstallState) string {
	switch state {
	case InstallStateBroken:
		return "portunix plugin repair " + rp.Name
	case InstallStateMissing:
		return "portunix plugin install --force <plugin-path>  (or: portunix plugin uninstall " + rp.Name + ")"
	default:
		return ""
	}
}

// installStatus maps an install state onto the plugin status shown to users;
// a stored status is kept only when the installation is usable
func installStatus(stored plugins.PluginStatus, state InstallState) plugins.PluginStatus {
	switch state {
	case InstallStateBroken:
		return plugins.PluginStatusBroken
	case InstallStateMissing:
		return plugins.PluginStatusFailed
	default:
		return stored
	}
}

// checkPythonVenv verifies that the venv interpreter resolves to an existing
// file; os.Stat follows the symlink chain, so a removed base interpreter
// (e.g. /usr/bin/python3.13 after an OS upgrade) is reported as an error
func checkPythonVenv(venvDir string) error {
	if _, err := os.Stat(venvDir); err != nil {
		return fmt.Errorf("virtual environment not found: %s", venvDir)
	}

	interpreter := venvExecPath(venvDir, "python")
	if _, err := os.Stat(interpreter); err != nil {
		target := interpreter
		if resolved, linkErr := os.Readlink(interpreter); linkErr == nil {
			target = resolved
		}
		return fmt.Errorf("venv interpreter is missing: %s -> %s", interpreter, target)
	}

	// Windows venvs copy a launcher that still needs the base installation
	if home := pyvenvHome(venvDir); home != "" {
		if _, err := os.Stat(home); err != nil {
			return fmt.Errorf("base Python installation not found: %s", home)
		}
	}

	return nil
}

// pyvenvHome returns the "home" value of pyvenv.cfg (the base interpreter
// directory), or empty string when it cannot be read
func pyvenvHome(venvDir string) string {
	f, err := os.Open(filepath.Join(venvDir, "pyvenv.cfg")) // #nosec G304 -- path inside plugin install dir
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, value, found := strings.Cut(scanner.Text(), "=")
		if found && strings.TrimSpace(key) == "home" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
