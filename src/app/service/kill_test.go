/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package service

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestKillProcess_ForceKillsDetachedProcessIgnoringInterrupt covers a plugin
// started by an earlier CLI (not our child) that does not exit on SIGINT
func TestKillProcess_ForceKillsDetachedProcessIgnoringInterrupt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell and signals required")
	}

	saved := killGracePeriod
	killGracePeriod = 300 * time.Millisecond
	t.Cleanup(func() { killGracePeriod = saved })

	// The inner shell is orphaned when the outer one exits, so it is not our child
	out, err := exec.Command("/bin/sh", "-c", `/bin/sh -c 'trap "" INT; while :; do sleep 1; done' >/dev/null 2>&1 & echo $!`).Output()
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		t.Fatalf("unexpected pid output %q: %v", out, err)
	}
	t.Cleanup(func() { _ = exec.Command("kill", "-9", strconv.Itoa(pid)).Run() })

	if !IsProcessAlive(pid) {
		t.Fatalf("test process %d not running", pid)
	}

	killProcess(pid)

	deadline := time.Now().Add(2 * time.Second)
	for IsProcessAlive(pid) && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if IsProcessAlive(pid) {
		t.Fatalf("process %d survived killProcess", pid)
	}
}
