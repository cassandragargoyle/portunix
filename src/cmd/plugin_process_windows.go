//go:build windows

/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"os/exec"
	"strconv"
)

// preparePluginProcessTree needs no setup on Windows: taskkill /T walks the
// process tree by parent pid
func preparePluginProcessTree(cmd *exec.Cmd) {}

// killPluginProcessTree kills a started plugin together with its children
// (e.g. the interpreter started by a .ps1 wrapper)
func killPluginProcessTree(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	// #nosec G204 -- fixed system tool, pid is an integer
	if err := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid)).Run(); err != nil {
		_ = cmd.Process.Kill()
	}
}
