//go:build !windows

/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"os/exec"
	"syscall"
)

// preparePluginProcessTree starts the plugin as leader of its own process
// group, so killPluginProcessTree reaches children of wrapper scripts
func preparePluginProcessTree(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// killPluginProcessTree kills the whole process group of a started plugin
func killPluginProcessTree(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	// Negative pid addresses the process group led by the plugin
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
		_ = cmd.Process.Kill()
	}
}
