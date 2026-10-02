/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package selfinstall

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestNormalizeTargetPath verifies that --path is resolved to a full binary
// file path whether callers pass an install directory or a full file path
// (Issue #197).
func TestNormalizeTargetPath(t *testing.T) {
	bin := mainBinaryName()

	t.Run("empty stays empty", func(t *testing.T) {
		if got := NormalizeTargetPath(""); got != "" {
			t.Fatalf("expected empty, got %q", got)
		}
	})

	t.Run("directory gets binary name appended", func(t *testing.T) {
		dir := filepath.Join("tmp", "Portunix")
		want := filepath.Join(dir, bin)
		if got := NormalizeTargetPath(dir); got != want {
			t.Fatalf("dir input: want %q, got %q", want, got)
		}
	})

	t.Run("full binary file path is used verbatim", func(t *testing.T) {
		full := filepath.Join("tmp", "Portunix", bin)
		if got := NormalizeTargetPath(full); got != full {
			t.Fatalf("file input: want %q, got %q", full, got)
		}
	})

	t.Run("unix-style install dir", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("unix path semantics")
		}
		if got := NormalizeTargetPath("/usr/local/bin"); got != "/usr/local/bin/portunix" {
			t.Fatalf("want /usr/local/bin/portunix, got %q", got)
		}
	})

	t.Run("windows binary name is case-insensitive", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("windows-only casing rule")
		}
		full := filepath.Join(`C:\Portunix`, "PORTUNIX.EXE")
		if got := NormalizeTargetPath(full); got != full {
			t.Fatalf("case-insensitive file input: want %q, got %q", full, got)
		}
	})
}
