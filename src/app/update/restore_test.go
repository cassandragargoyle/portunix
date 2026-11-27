package update

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRestoreBackups(t *testing.T) {
	dir := t.TempDir()
	backupPath := filepath.Join(dir, "portunix.backup")
	if err := os.WriteFile(backupPath, []byte("old"), 0600); err != nil {
		t.Fatalf("write backup: %v", err)
	}
	targetPath := filepath.Join(dir, "portunix")
	if err := os.WriteFile(targetPath, []byte("broken"), 0600); err != nil {
		t.Fatalf("write target: %v", err)
	}

	if err := restoreBackups(map[string]string{"portunix": backupPath}, dir, ""); err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	got, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "old" {
		t.Errorf("expected restored content %q, got %q", "old", got)
	}
}

func TestRestoreBackupsReportsFailures(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "ptx-container.backup")

	err := restoreBackups(map[string]string{"ptx-container": missing}, dir, "")
	if err == nil {
		t.Fatal("expected an error for a missing backup file")
	}
	if !strings.Contains(err.Error(), "ptx-container") {
		t.Errorf("error does not name the binary: %v", err)
	}
}

func TestCopyLimited(t *testing.T) {
	var dst bytes.Buffer
	if err := copyLimited(&dst, strings.NewReader("12345"), 5); err != nil {
		t.Fatalf("copy within the limit failed: %v", err)
	}
	if dst.String() != "12345" {
		t.Errorf("expected full content, got %q", dst.String())
	}

	dst.Reset()
	if err := copyLimited(&dst, strings.NewReader("123456"), 5); err == nil {
		t.Error("expected an error when the source exceeds the limit")
	}
}

// TestApplyUpdateAllBinariesRollback injects a failure for the second binary:
// the first is restored, the untouched ones keep their content, no backup or
// temp file is left (Issue #200)
func TestApplyUpdateAllBinariesRollback(t *testing.T) {
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	dir := t.TempDir()
	tmp := t.TempDir()

	names := []string{"portunix", "ptx-blocked", "ptx-later"}
	var binaries []*BinaryInfo
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name+suffix), []byte("old-"+name), 0600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		src := filepath.Join(tmp, name)
		if err := os.WriteFile(src, []byte("new-"+name), 0600); err != nil {
			t.Fatalf("write new %s: %v", name, err)
		}
		binaries = append(binaries, &BinaryInfo{Name: name, Path: src})
	}
	// A non-empty directory at the .new path blocks replacing ptx-blocked
	blocker := filepath.Join(dir, "ptx-blocked"+suffix+".new")
	if err := os.MkdirAll(filepath.Join(blocker, "x"), 0750); err != nil {
		t.Fatalf("create blocker: %v", err)
	}

	err := applyUpdateAllBinaries(dir, binaries)
	if err == nil || !strings.Contains(err.Error(), "ptx-blocked") {
		t.Fatalf("expected failure for ptx-blocked, got %v", err)
	}
	if errors.Is(err, ErrRestoreFailed) {
		t.Errorf("rollback reported as failed: %v", err)
	}

	for _, name := range names {
		got, _ := os.ReadFile(filepath.Join(dir, name+suffix))
		if string(got) != "old-"+name {
			t.Errorf("%s content %q after rollback, want %q", name, got, "old-"+name)
		}
	}
	if err := os.RemoveAll(blocker); err != nil {
		t.Fatalf("remove blocker: %v", err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != len(names) {
		var left []string
		for _, e := range entries {
			left = append(left, e.Name())
		}
		t.Errorf("leftovers in install dir: %v", left)
	}
	if entries, _ := os.ReadDir(tmp); len(entries) != 0 {
		t.Errorf("%d extracted temp files left", len(entries))
	}
}

func TestRestoreBackupsRemovesRestoredBackups(t *testing.T) {
	dir := t.TempDir()
	okBackup := filepath.Join(dir, "portunix.backup")
	if err := os.WriteFile(okBackup, []byte("old"), 0600); err != nil {
		t.Fatalf("write backup: %v", err)
	}
	missing := filepath.Join(dir, "ptx-container.backup")

	err := restoreBackups(map[string]string{"portunix": okBackup, "ptx-container": missing}, dir, "")
	if !errors.Is(err, ErrRestoreFailed) {
		t.Fatalf("expected ErrRestoreFailed, got %v", err)
	}
	if _, err := os.Stat(okBackup); err == nil {
		t.Error("backup of the restored binary was kept")
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "portunix")); string(got) != "old" {
		t.Errorf("portunix content %q, want %q", got, "old")
	}
}
