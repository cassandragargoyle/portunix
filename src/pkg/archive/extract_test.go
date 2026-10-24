/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type testMember struct {
	name     string
	content  string
	linkname string
	dir      bool
}

// writeTarGz creates a .tar.gz archive with the given members
func writeTarGz(t *testing.T, path string, members []testMember) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for _, m := range members {
		hdr := &tar.Header{Name: m.name, Mode: 0o755}
		switch {
		case m.dir:
			hdr.Typeflag = tar.TypeDir
		case m.linkname != "":
			hdr.Typeflag = tar.TypeSymlink
			hdr.Linkname = m.linkname
		default:
			hdr.Typeflag = tar.TypeReg
			hdr.Size = int64(len(m.content))
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if hdr.Typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte(m.content)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestExtractTarGz_ReplacesLongerExistingFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "dest")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	// Previous version is longer than the new one
	if err := os.WriteFile(filepath.Join(dest, "tool"), []byte("old-binary-content-that-is-longer"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "other-tool"), []byte("keep"), 0o755); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(dir, "tool.tar.gz")
	writeTarGz(t, archivePath, []testMember{{name: "tool", content: "new"}})

	if err := ExtractArchive(archivePath, dest); err != nil {
		t.Fatalf("ExtractArchive: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "tool"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("tool content = %q, want %q (old tail must not survive)", got, "new")
	}
	if _, err := os.Stat(filepath.Join(dest, "other-tool")); err != nil {
		t.Errorf("unrelated file was removed: %v", err)
	}
}

func TestExtractTarGz_ReplacesExistingSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require privileges on Windows")
	}
	dir := t.TempDir()
	dest := filepath.Join(dir, "dest")
	archivePath := filepath.Join(dir, "pkg.tar.gz")
	writeTarGz(t, archivePath, []testMember{
		{name: "pkg/", dir: true},
		{name: "pkg/real", content: "x"},
		{name: "pkg/link", linkname: "real"},
	})

	// Second extraction is a reinstall over the existing symlink
	for i := 0; i < 2; i++ {
		if err := ExtractArchive(archivePath, dest); err != nil {
			t.Fatalf("ExtractArchive run %d: %v", i+1, err)
		}
	}

	target, err := os.Readlink(filepath.Join(dest, "pkg", "link"))
	if err != nil {
		t.Fatal(err)
	}
	if target != "real" {
		t.Errorf("symlink target = %q, want %q", target, "real")
	}
}

func TestListTopLevelEntries_TarGz(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "go.tar.gz")
	writeTarGz(t, archivePath, []testMember{
		{name: "go/bin/go", content: "x"},
		{name: "go/src/fmt/print.go", content: "x"},
		{name: "./LICENSE", content: "x"},
		{name: "README.md", content: "x"},
		{name: "empty/", dir: true},
	})

	got, err := ListTopLevelEntries(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	want := []TopLevelEntry{
		{Name: "LICENSE", IsDir: false},
		{Name: "README.md", IsDir: false},
		{Name: "empty", IsDir: true},
		{Name: "go", IsDir: true},
	}
	assertEntries(t, got, want)
}

func TestListTopLevelEntries_Zip(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "openssh.zip")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for _, name := range []string{"OpenSSH-Win64/", "OpenSSH-Win64/ssh.exe", "hugo.exe"} {
		if _, err := zw.Create(name); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()

	got, err := ListTopLevelEntries(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	want := []TopLevelEntry{
		{Name: "OpenSSH-Win64", IsDir: true},
		{Name: "hugo.exe", IsDir: false},
	}
	assertEntries(t, got, want)
}

func TestListTopLevelEntries_Unsupported(t *testing.T) {
	if _, err := ListTopLevelEntries("file.rar"); err == nil {
		t.Error("expected error for unsupported archive type")
	}
}

func assertEntries(t *testing.T, got, want []TopLevelEntry) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("entries = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestCopyLimited(t *testing.T) {
	var out bytes.Buffer
	if err := copyLimited(&out, strings.NewReader("12345"), 5); err != nil {
		t.Fatalf("content at the limit must pass: %v", err)
	}
	if out.String() != "12345" {
		t.Errorf("unexpected content: %q", out.String())
	}

	out.Reset()
	if err := copyLimited(&out, strings.NewReader("123456"), 5); err == nil {
		t.Error("content over the limit must fail")
	}
}
