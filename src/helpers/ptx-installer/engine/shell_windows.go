//go:build windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// shellCommand builds a command that runs script through cmd.exe verbatim.
// The command line is set directly because Go's argument escaping turns every
// inner `"` into `\"`, which cmd.exe does not understand (issue #201).
// With /s, cmd strips only the outer pair of quotes and keeps the rest intact
func shellCommand(script string) *exec.Cmd {
	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: `cmd /d /s /c "` + script + `"`,
	}
	return cmd
}

// refreshPath reloads PATH from the machine and user environment in the
// registry, so a tool installed into a directory that an installer has just
// added to the user PATH (e.g. %USERPROFILE%\.local\bin) is found
func refreshPath() {
	var parts []string
	if machine := readRegistryPath(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`); machine != "" {
		parts = append(parts, machine)
	}
	if user := readRegistryPath(registry.CURRENT_USER, `Environment`); user != "" {
		parts = append(parts, user)
	}
	if len(parts) == 0 {
		return
	}

	// Keep entries of the current process that the registry does not know
	// (e.g. directories added by the parent shell)
	merged := strings.Join(parts, string(os.PathListSeparator))
	known := make(map[string]bool)
	for _, p := range filepath.SplitList(merged) {
		known[strings.ToLower(filepath.Clean(p))] = true
	}
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if p != "" && !known[strings.ToLower(filepath.Clean(p))] {
			merged += string(os.PathListSeparator) + p
		}
	}
	os.Setenv("PATH", merged)
}

// readRegistryPath returns the expanded Path value of the given registry key
func readRegistryPath(root registry.Key, path string) string {
	key, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer key.Close()

	value, _, err := key.GetStringValue("Path")
	if err != nil {
		return ""
	}
	expanded, err := registry.ExpandString(value)
	if err != nil {
		return value
	}
	return expanded
}
