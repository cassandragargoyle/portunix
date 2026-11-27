/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package update

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReplaceBinaryNewTarget(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	target := filepath.Join(dir, "ptx-new")
	if err := os.WriteFile(src, []byte("new"), 0600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := replaceBinary(src, target); err != nil {
		t.Fatalf("replaceBinary: %v", err)
	}
	if got, _ := os.ReadFile(target); string(got) != "new" {
		t.Errorf("target content %q, want %q", got, "new")
	}
	for _, leftover := range []string{target + ".new", target + ".old"} {
		if _, err := os.Stat(leftover); err == nil {
			t.Errorf("%s left behind", filepath.Base(leftover))
		}
	}
}

// TestReplaceBinaryRunningTarget replaces a binary while a process runs from
// it, which an in-place overwrite cannot do (ETXTBSY on Linux, Issue #200)
func TestReplaceBinaryRunningTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a Unix sleep binary")
	}
	sleepPath, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("sleep not available")
	}

	dir := t.TempDir()
	target := filepath.Join(dir, "portunix")
	if err := copyFile(sleepPath, target); err != nil {
		t.Fatalf("copy sleep: %v", err)
	}
	if err := os.Chmod(target, 0755); err != nil { // #nosec G302 -- test binary must be executable
		t.Fatalf("chmod: %v", err)
	}

	cmd := exec.Command(target, "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start target: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	// Precondition: an in-place write to the running binary fails
	if f, err := os.OpenFile(target, os.O_WRONLY, 0); err == nil {
		_ = f.Close()
		t.Log("kernel allows writing a running binary, precondition not reproduced")
	}

	src := filepath.Join(dir, "src")
	if err := os.WriteFile(src, []byte("new"), 0600); err != nil {
		t.Fatalf("write src: %v", err)
	}
	if err := replaceBinary(src, target); err != nil {
		t.Fatalf("replaceBinary on a running target: %v", err)
	}
	if got, _ := os.ReadFile(target); string(got) != "new" {
		t.Errorf("target content %q, want %q", got, "new")
	}
}

func TestCheckWritePermission(t *testing.T) {
	dir := t.TempDir()
	if err := checkWritePermission(dir); err != nil {
		t.Fatalf("writable dir: %v", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("check left %d files behind", len(entries))
	}
	if err := checkWritePermission(filepath.Join(dir, "missing")); err == nil {
		t.Error("missing dir: expected error")
	}
}
