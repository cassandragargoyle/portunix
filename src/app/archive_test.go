package app

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeZip creates a ZIP archive with the given entry names and contents
func writeZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, content := range entries {
		entry, err := w.Create(name)
		if err != nil {
			t.Fatalf("create entry %s: %v", name, err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatalf("write entry %s: %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
}

func TestUnzipExtractsFiles(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "ok.zip")
	writeZip(t, zipPath, map[string]string{"a.txt": "alpha", "sub/b.txt": "beta"})

	dest := filepath.Join(dir, "out")
	if err := unzip(zipPath, dest); err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	for name, want := range map[string]string{"a.txt": "alpha", "sub/b.txt": "beta"} {
		got, err := os.ReadFile(filepath.Join(dest, filepath.FromSlash(name)))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if string(got) != want {
			t.Errorf("%s: expected %q, got %q", name, want, got)
		}
	}
}

func TestUnzipRejectsZipSlip(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "evil.zip")
	writeZip(t, zipPath, map[string]string{"../escaped.txt": "evil"})

	dest := filepath.Join(dir, "out")
	err := unzip(zipPath, dest)
	if err == nil {
		t.Fatal("expected an error for an entry escaping the destination")
	}
	if !strings.Contains(err.Error(), "illegal file path") {
		t.Errorf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "escaped.txt")); !os.IsNotExist(statErr) {
		t.Error("entry was written outside the destination directory")
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
