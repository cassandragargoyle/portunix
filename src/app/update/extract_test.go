/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// zipEntries and tarEntries are file sets of flat release archives, the content of
// each entry is its own name so extracted files can be matched back
var zipEntries = []string{
	"LICENSE", "README.md", "install.bat", "install.ps1", "install.sh",
	"portunix.exe", "ptx-installer.exe", "ptx-container.exe", "ptx-future.exe",
	"ptx-stray",
}

var tarEntries = []string{
	"LICENSE", "README.md", "install.bat", "install.ps1", "install.sh",
	"portunix", "ptx-installer", "ptx-container", "ptx-future",
	"ptx-stray.exe",
}

func writeZip(t *testing.T, names []string) string {
	t.Helper()
	archivePath := filepath.Join(t.TempDir(), "release.zip")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	zw := zip.NewWriter(f)
	for _, name := range names {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip entry %s: %v", name, err)
		}
		if _, err := w.Write([]byte(name)); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	_ = f.Close()
	return archivePath
}

func writeTarGz(t *testing.T, names []string) string {
	t.Helper()
	archivePath := filepath.Join(t.TempDir(), "release.tar.gz")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create tar.gz: %v", err)
	}
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	for _, name := range names {
		hdr := &tar.Header{Name: name, Mode: 0755, Size: int64(len(name)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("tar header %s: %v", name, err)
		}
		if _, err := tw.Write([]byte(name)); err != nil {
			t.Fatalf("tar write %s: %v", name, err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	_ = f.Close()
	return archivePath
}

// checkExtracted verifies the extracted set by name and content and removes the temp files
func checkExtracted(t *testing.T, binaries []*BinaryInfo, suffix string, want []string) {
	t.Helper()
	var got []string
	for _, binary := range binaries {
		got = append(got, binary.Name)
		data, err := os.ReadFile(binary.Path)
		if err != nil {
			t.Errorf("read extracted %s: %v", binary.Name, err)
		} else if string(data) != binary.Name+suffix {
			t.Errorf("%s: extracted content %q, want %q", binary.Name, data, binary.Name+suffix)
		}
		_ = os.Remove(binary.Path)
	}
	sort.Strings(got)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("extracted %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("extracted %v, want %v", got, want)
		}
	}
}

func TestReleaseBinaryName(t *testing.T) {
	tests := []struct {
		entry   string
		windows bool
		want    string
		ok      bool
	}{
		{"portunix.exe", true, "portunix", true},
		{"PTX-Installer.EXE", true, "PTX-Installer", false},
		{"ptx-installer.exe", true, "ptx-installer", true},
		{"dir/ptx-mcp.exe", true, "ptx-mcp", true},
		{`dir\ptx-mcp.exe`, true, "ptx-mcp", true},
		{"ptx-.exe", true, "", false},
		{"ptx-mcp", true, "", false},
		{"install.ps1", true, "", false},
		{"portunix", false, "portunix", true},
		{"ptx-wizard", false, "ptx-wizard", true},
		{"ptx-wizard.exe", false, "", false},
		{"install.sh", false, "", false},
		{"LICENSE", false, "", false},
		{"portunix-backup", false, "", false},
	}
	for _, tt := range tests {
		got, ok := releaseBinaryName(tt.entry, tt.windows)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("releaseBinaryName(%q, %v) = %q, %v; want %q, %v", tt.entry, tt.windows, got, ok, tt.want, tt.ok)
		}
	}
}

// TestExtractAllFromZip verifies that update picks up every ptx-* helper of
// the archive, not a hardcoded subset (Issue #200)
func TestExtractAllFromZip(t *testing.T) {
	binaries, err := extractAllFromZip(writeZip(t, zipEntries))
	if err != nil {
		t.Fatalf("extractAllFromZip: %v", err)
	}
	checkExtracted(t, binaries, ".exe", []string{"portunix", "ptx-installer", "ptx-container", "ptx-future"})
}

// TestExtractAllFromTarGz verifies that update picks up every ptx-* helper of
// the archive, not a hardcoded subset (Issue #200)
func TestExtractAllFromTarGz(t *testing.T) {
	binaries, err := extractAllFromTarGz(writeTarGz(t, tarEntries))
	if err != nil {
		t.Fatalf("extractAllFromTarGz: %v", err)
	}
	checkExtracted(t, binaries, "", []string{"portunix", "ptx-installer", "ptx-container", "ptx-future"})
}

func TestExtractAllRequiresMainBinary(t *testing.T) {
	if _, err := extractAllFromZip(writeZip(t, []string{"ptx-installer.exe", "README.md"})); err == nil {
		t.Error("zip without portunix.exe: expected error")
	}
	if _, err := extractAllFromTarGz(writeTarGz(t, []string{"ptx-installer", "README.md"})); err == nil {
		t.Error("tar.gz without portunix: expected error")
	}
}
