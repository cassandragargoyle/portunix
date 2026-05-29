//go:build windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"os"
	"syscall"
)

// newDetachedSysProcAttr asks Windows to detach the child into its own process
// group, which is the closest Windows-native equivalent of POSIX setsid for
// daemonisation. The DETACHED_PROCESS flag (0x00000008) prevents inheriting
// the parent console, so closing the launcher shell does not signal the child.
func newDetachedSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x00000008, // DETACHED_PROCESS
	}
}

// processAlive: on Windows os.FindProcess always succeeds (it's a no-op),
// so we additionally try OpenProcess via syscall.Handle to confirm the pid
// has a living process behind it.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	var exitCode uint32
	if err := syscall.GetExitCodeProcess(handle, &exitCode); err != nil {
		return false
	}
	const STILL_ACTIVE = 259
	return exitCode == STILL_ACTIVE
}

// signalProcess on Windows: SIGTERM/SIGINT are not natively supported, so we
// fall back to terminating the process. The lifecycle daemon's only signal
// consumer is the shutdown handler — losing graceful shutdown on Windows is
// an accepted trade-off and matches how most Go daemons behave there.
func signalProcess(pid int, _ syscall.Signal) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}
