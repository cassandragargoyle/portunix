//go:build !windows

/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestRunPluginHelpAI_TimeoutKillsChildren checks that a timeout kills the
// children of a wrapper script too, not only the wrapper itself
func TestRunPluginHelpAI_TimeoutKillsChildren(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	// Wrapper keeps running while its child sleeps; no exec, so the child is a separate process
	plugin := newStubPlugin(t, "sleep 30 &\necho $! > '"+pidFile+"'\nwait\n", "")

	var out bytes.Buffer
	err := runPluginHelpAI(plugin, []string{"--help-ai"}, &out, 500*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "did not answer --help-ai within") {
		t.Fatalf("expected timeout error, got %v", err)
	}

	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("stub did not record its child pid: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatalf("invalid child pid %q: %v", data, err)
	}

	// The killed child may still be reaped by init, allow a short delay
	deadline := time.Now().Add(2 * time.Second)
	for {
		if err := syscall.Kill(pid, 0); errors.Is(err, syscall.ESRCH) {
			return
		}
		if time.Now().After(deadline) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
			t.Fatalf("child process %d survived the timeout", pid)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
