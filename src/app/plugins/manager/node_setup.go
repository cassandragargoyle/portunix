/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"portunix.ai/app/plugins"
)

// nodeModulesDir is the directory npm installs plugin dependencies into
const nodeModulesDir = "node_modules"

// nodeDepsPlan describes how the npm dependencies of a Node.js plugin are installed
type nodeDepsPlan struct {
	npm  string   // npm executable
	args []string // npm arguments (ci or install)
}

// packageManifest is the part of package.json that plugin install reads
type packageManifest struct {
	Dependencies map[string]string `json:"dependencies"`
}

// nodeDependenciesMissing reports whether package.json in pluginDir declares
// runtime dependencies that are not present in node_modules; a plugin that
// ships node_modules is used as is
func nodeDependenciesMissing(pluginDir string) (bool, error) {
	data, err := os.ReadFile(filepath.Join(pluginDir, "package.json"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg packageManifest
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false, fmt.Errorf("invalid package.json: %w", err)
	}
	if len(pkg.Dependencies) == 0 {
		return false, nil
	}

	if info, err := os.Stat(filepath.Join(pluginDir, nodeModulesDir)); err == nil && info.IsDir() {
		return false, nil
	}
	return true, nil
}

// planNodeDependencies checks everything needed to install the npm
// dependencies of the plugin in pluginDir; nil plan means nothing to install
func planNodeDependencies(pluginDir string) (*nodeDepsPlan, error) {
	missing, err := nodeDependenciesMissing(pluginDir)
	if err != nil || !missing {
		return nil, err
	}

	if _, err := exec.LookPath("node"); err != nil {
		return nil, fmt.Errorf("Node.js runtime not found; it is needed to install the plugin dependencies. Install it with: %s", plugins.NodeInstallFix)
	}
	npm, err := exec.LookPath("npm")
	if err != nil {
		return nil, fmt.Errorf("npm not found; it is needed to install the plugin dependencies from package.json. Install it with: %s", plugins.NodeInstallFix)
	}

	// A lockfile gives a reproducible install; without one npm resolves the ranges
	command := "install"
	for _, lockfile := range []string{"package-lock.json", "npm-shrinkwrap.json"} {
		if _, err := os.Stat(filepath.Join(pluginDir, lockfile)); err == nil {
			command = "ci"
			break
		}
	}

	return &nodeDepsPlan{
		npm:  npm,
		args: []string{command, "--omit=dev", "--no-audit", "--no-fund"},
	}, nil
}

// setupNodeDependencies installs the npm dependencies of the plugin into
// pluginDir/node_modules; a nil plan is a no-op
func setupNodeDependencies(plan *nodeDepsPlan, pluginDir string) error {
	if plan == nil {
		return nil
	}

	fmt.Printf("Installing Node.js dependencies (npm %s)...\n", plan.args[0])
	cmd := exec.Command(plan.npm, plan.args...) // #nosec G204 -- npm from PATH, fixed arguments
	cmd.Dir = pluginDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npm %s failed: %w\n%s", plan.args[0], err, strings.TrimSpace(string(output)))
	}
	return nil
}

// checkNodeModules verifies that the npm dependencies of an installed
// Node.js plugin are present
func checkNodeModules(installPath string) error {
	missing, err := nodeDependenciesMissing(installPath)
	if err != nil {
		return err
	}
	if missing {
		return fmt.Errorf("Node.js dependencies not installed: %s not found in %s", nodeModulesDir, installPath)
	}
	return nil
}
