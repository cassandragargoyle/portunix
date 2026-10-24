/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

// Placeholders the engine substitutes in postInstall commands. Anything else
// written as ${name} must be an environment variable that is set at install
// time, otherwise the command is rejected (issue #227)
const (
	placeholderInstallPath     = "${install_path}"
	placeholderActualExtractTo = "${actual_extract_to}"
	placeholderSudoPrefix      = "${sudo_prefix}"
)

// PostInstallPlaceholders lists the engine placeholders valid in postInstall
var PostInstallPlaceholders = []string{
	placeholderInstallPath,
	placeholderActualExtractTo,
	placeholderSudoPrefix,
}

// unresolvedPlaceholderRe matches any ${name} left after substitution
var unresolvedPlaceholderRe = regexp.MustCompile(`\$\{[^}]+\}`)

// postInstallSudoPrefix returns "sudo " when commands need elevation through
// sudo: the process is not root and sudo exists. Root shells, systems without
// sudo and Windows get an empty prefix. Variable so tests can replace it
var postInstallSudoPrefix = func() string {
	if runtime.GOOS == "windows" || IsRunningAsRoot() || !IsSudoAvailable() {
		return ""
	}
	return "sudo "
}

// expandPostInstallCommand substitutes engine placeholders and environment
// variables in a postInstall command. installPath is the directory the package
// really landed in (after the user-directory fallback); it is empty for
// install types without one, which makes the path placeholders invalid there.
// A ${name} that survives substitution is an error naming the placeholder —
// the shell would otherwise expand it to an empty string
func expandPostInstallCommand(cmd, installPath string) (string, error) {
	cmd = strings.ReplaceAll(cmd, placeholderSudoPrefix, postInstallSudoPrefix())
	if installPath != "" {
		cmd = strings.ReplaceAll(cmd, placeholderInstallPath, installPath)
		cmd = strings.ReplaceAll(cmd, "%install_path%", installPath)
		cmd = strings.ReplaceAll(cmd, placeholderActualExtractTo, installPath)
	}
	cmd = expandEnvVars(cmd)

	if unresolved := unresolvedPlaceholderRe.FindAllString(cmd, -1); len(unresolved) > 0 {
		sort.Strings(unresolved)
		return "", fmt.Errorf("unknown placeholder %s in post-install command: %s",
			strings.Join(uniqueStrings(unresolved), ", "), cmd)
	}
	return cmd, nil
}

// uniqueStrings removes adjacent duplicates from a sorted slice
func uniqueStrings(sorted []string) []string {
	var out []string
	for _, s := range sorted {
		if len(out) == 0 || out[len(out)-1] != s {
			out = append(out, s)
		}
	}
	return out
}

// runPostInstallCommands expands and runs postInstall commands and returns one
// message per failed command; remaining commands still run after a failure.
// When the package landed somewhere else than the manifest's extractTo
// (originalPath), literal occurrences of that path are rewritten to installPath
func runPostInstallCommands(commands []string, installPath, originalPath string) []string {
	var failures []string
	for _, raw := range commands {
		cmd, err := expandPostInstallCommand(raw, installPath)
		if err != nil {
			fmt.Printf("   ❌ %v\n", err)
			failures = append(failures, err.Error())
			continue
		}

		if originalPath != "" && installPath != "" && originalPath != installPath {
			// Normalize paths for replacement (handle both / and \ on Windows)
			cmd = strings.ReplaceAll(cmd, filepath.ToSlash(originalPath), filepath.ToSlash(installPath))
			cmd = strings.ReplaceAll(cmd,
				strings.ReplaceAll(originalPath, "/", "\\"),
				strings.ReplaceAll(installPath, "/", "\\"))
		}

		fmt.Printf("   Running: %s\n", cmd)

		execCmd := shellCommand(cmd)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		if err := execCmd.Run(); err != nil {
			errMsg := fmt.Sprintf("Command failed: %s (error: %v)", cmd, err)
			fmt.Printf("   ❌ %s\n", errMsg)
			failures = append(failures, errMsg)
		}
	}
	return failures
}

// runPackageManagerPostInstall runs postInstall after a package-manager
// install (apt, dnf, pacman, snap, deb). The package itself is already
// installed by the package manager, so failures here are reported as warnings
func runPackageManagerPostInstall(commands []string) {
	if len(commands) == 0 {
		return
	}
	fmt.Println("🔧 Running post-install commands...")
	if failures := runPostInstallCommands(commands, "", ""); len(failures) > 0 {
		fmt.Printf("⚠️  %d post-install command(s) failed (package is installed):\n   - %s\n",
			len(failures), strings.Join(failures, "\n   - "))
	}
}
