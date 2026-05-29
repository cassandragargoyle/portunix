//go:build !windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"os"
	"syscall"
)

// newDetachedSysProcAttr makes the daemon child a session leader so it
// outlives the foreground shell that started `service start --detach`.
func newDetachedSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

// processAlive returns true when the given PID corresponds to a live process.
// Sending signal 0 is the canonical POSIX existence check.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func signalProcess(pid int, sig syscall.Signal) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(sig)
}
