package update

import (
	"bytes"
	"os"
	"path/filepath"
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
