/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package plugins

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CommandWarnings returns non-fatal problems of the manifest `commands` list.
// The rule (issue #257): one entry per subcommand a user calls as
// `portunix <plugin> <name>`, no entry for the binary itself, an empty list
// for a plugin without subcommands
func CommandWarnings(manifest *PluginManifest) []string {
	var warnings []string
	if len(manifest.Commands) == 1 && isPlaceholderCommand(manifest.Commands[0].Name, manifest) {
		warnings = append(warnings, fmt.Sprintf(
			"command '%s' looks like a placeholder for the plugin itself; list the subcommands called as 'portunix %s <command>', or use an empty list for a plugin without subcommands",
			manifest.Commands[0].Name, manifest.Name))
	}
	return warnings
}

// isPlaceholderCommand reports whether a command name refers to the plugin
// or its binary instead of a real subcommand: the plugin name, ptx-<plugin>,
// the binary base name, or a dash-separated prefix of the plugin name
func isPlaceholderCommand(command string, manifest *PluginManifest) bool {
	name := strings.ToLower(strings.TrimSpace(command))
	plugin := strings.ToLower(manifest.Name)
	if name == "" || plugin == "" {
		return false
	}
	trimmed := strings.TrimPrefix(name, "ptx-")
	if trimmed == plugin || strings.HasPrefix(plugin, trimmed+"-") {
		return true
	}
	if binary := binaryBaseName(manifest.Plugin.Binary); binary != "" && name == binary {
		return true
	}
	return false
}

// binaryBaseName returns the lower-case file name of the plugin binary
// without directory and extension
func binaryBaseName(binary string) string {
	base := filepath.Base(filepath.ToSlash(strings.TrimSpace(binary)))
	if base == "." || base == "/" {
		return ""
	}
	return strings.ToLower(strings.TrimSuffix(base, filepath.Ext(base)))
}
