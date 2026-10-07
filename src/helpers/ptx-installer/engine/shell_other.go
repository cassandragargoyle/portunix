//go:build !windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"os"
	"os/exec"
	"path/filepath"
)

// shellCommand builds a command that runs script through sh
func shellCommand(script string) *exec.Cmd {
	return exec.Command("sh", "-c", script)
}

// refreshPath appends well-known user binary directories that install
// scripts commonly use (uv, rustup) to PATH when they exist and are missing
func refreshPath() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	for _, dir := range []string{
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, ".cargo", "bin"),
	} {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			continue
		}
		if !IsInUserPath(dir) {
			os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+dir)
		}
	}
}
