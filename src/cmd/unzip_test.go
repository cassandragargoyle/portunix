/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestZip creates a ZIP archive with the given entry names and contents
func writeTestZip(t *testing.T, path string, entries map[string]string) {
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

// assertFileContent fails the test when the file does not hold the expected content
func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != want {
		t.Errorf("%s: expected %q, got %q", path, want, got)
	}
}

// assertNoFlagNamedEntries fails the test when an entry named after a flag exists in dir (#231)
func assertNoFlagNamedEntries(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "-") {
			t.Errorf("unexpected entry named after a flag: %s", e.Name())
		}
	}
}

// chdir switches the working directory for the duration of the test
func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

// TestRunUnzip_NextToArchive checks that an archive without destination extracts beside itself (#231)
func TestRunUnzip_NextToArchive(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	writeTestZip(t, filepath.Join(dir, "a.zip"), map[string]string{"a.txt": "alpha", "sub/b.txt": "beta"})

	if err := runUnzip("a.zip", ""); err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	assertFileContent(t, filepath.Join(dir, "a.txt"), "alpha")
	assertFileContent(t, filepath.Join(dir, "sub", "b.txt"), "beta")
	assertNoFlagNamedEntries(t, dir)
}

// TestRunUnzip_Destination checks extraction into a relative and an absolute destination (#231)
func TestRunUnzip_Destination(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	writeTestZip(t, filepath.Join(dir, "a.zip"), map[string]string{"a.txt": "alpha"})

	if err := runUnzip("a.zip", "out"); err != nil {
		t.Fatalf("unzip into relative destination failed: %v", err)
	}
	assertFileContent(t, filepath.Join(dir, "out", "a.txt"), "alpha")

	absDest := filepath.Join(t.TempDir(), "abs-out")
	if err := runUnzip(filepath.Join(dir, "a.zip"), absDest); err != nil {
		t.Fatalf("unzip into absolute destination failed: %v", err)
	}
	assertFileContent(t, filepath.Join(absDest, "a.txt"), "alpha")

	assertNoFlagNamedEntries(t, dir)
}

// TestRunUnzip_Errors checks that a missing path, a missing archive and a ZipSlip entry fail (#231)
func TestRunUnzip_Errors(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	writeTestZip(t, filepath.Join(dir, "evil.zip"), map[string]string{"../escaped.txt": "evil"})

	if err := runUnzip("", ""); err == nil {
		t.Error("expected an error for an empty path")
	}

	if err := runUnzip("missing.zip", "out-missing"); err == nil {
		t.Error("expected an error for a missing archive")
	}
	if _, err := os.Stat(filepath.Join(dir, "out-missing")); !os.IsNotExist(err) {
		t.Error("destination directory was created for a missing archive")
	}

	err := runUnzip("evil.zip", "out")
	if err == nil {
		t.Fatal("expected an error for an entry escaping the destination")
	}
	if !strings.Contains(err.Error(), "illegal file path") {
		t.Errorf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "escaped.txt")); !os.IsNotExist(statErr) {
		t.Error("entry was written outside the destination directory")
	}

	assertNoFlagNamedEntries(t, dir)
}
