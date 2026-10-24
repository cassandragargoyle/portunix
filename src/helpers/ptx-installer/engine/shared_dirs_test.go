/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"portunix.ai/portunix/src/pkg/archive"
)

// withSharedDirectories replaces the shared directory list for one test
func withSharedDirectories(t *testing.T, dirs ...string) {
	t.Helper()
	original := sharedDirectoriesFunc
	sharedDirectoriesFunc = func() []string { return dirs }
	t.Cleanup(func() { sharedDirectoriesFunc = original })
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected %s to exist: %v", path, err)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed (err=%v)", path, err)
	}
}

// writeTestTarGz creates a .tar.gz with regular files name -> content
func writeTestTarGz(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestIsSharedDirectory_ConfiguredList(t *testing.T) {
	shared := t.TempDir()
	withSharedDirectories(t, shared)

	if !isSharedDirectory(shared) {
		t.Errorf("%s should be shared", shared)
	}
	if !isSharedDirectory(shared + string(os.PathSeparator)) {
		t.Errorf("%s with trailing separator should be shared", shared)
	}
	if !isSharedDirectory(filepath.Join(shared, "sub", "..")) {
		t.Errorf("unclean path to %s should be shared", shared)
	}
	if isSharedDirectory(filepath.Join(shared, "hugo")) {
		t.Errorf("package subdirectory of %s must not be shared", shared)
	}
}

func TestIsSharedDirectory_FilesystemRoot(t *testing.T) {
	withSharedDirectories(t)
	root := string(os.PathSeparator)
	if runtime.GOOS == "windows" {
		root = `C:\`
	}
	if !isSharedDirectory(root) {
		t.Errorf("filesystem root %s should be shared", root)
	}
}

func TestIsSharedDirectory_Symlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require privileges on Windows")
	}
	shared := t.TempDir()
	withSharedDirectories(t, shared)

	link := filepath.Join(t.TempDir(), "link-to-shared")
	if err := os.Symlink(shared, link); err != nil {
		t.Fatal(err)
	}
	if !isSharedDirectory(link) {
		t.Errorf("symlink %s -> %s should be shared", link, shared)
	}
}

func TestIsSharedDirectory_DefaultUnixList(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix paths")
	}
	for _, dir := range []string{"/usr/local/bin", "/usr/local", "/usr/bin", "/opt", "/"} {
		if !isSharedDirectory(dir) {
			t.Errorf("%s should be shared", dir)
		}
	}
	for _, dir := range []string{"/usr/local/go", "/opt/maven", "/opt/java/jdk-21.0.8+9"} {
		if isSharedDirectory(dir) {
			t.Errorf("%s must not be shared", dir)
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		if !isSharedDirectory(filepath.Join(home, ".local", "bin")) {
			t.Errorf("~/.local/bin should be shared")
		}
		if isSharedDirectory(filepath.Join(home, ".local", "share", "portunix", "packages", "hugo")) {
			t.Errorf("per-package fallback directory must not be shared")
		}
	}
}

func TestCheckExistingDirectory_SharedNeverRemoved(t *testing.T) {
	shared := t.TempDir()
	withSharedDirectories(t, shared)
	writeFile(t, filepath.Join(shared, "other-tool"), "keep")

	for _, force := range []bool{false, true} {
		proceed, err := checkExistingDirectory(shared, force)
		if err != nil || !proceed {
			t.Fatalf("force=%v: proceed=%v err=%v, want proceed without error", force, proceed, err)
		}
		assertExists(t, filepath.Join(shared, "other-tool"))
	}
}

func TestCheckExistingDirectory_OwnedRemovedWithForce(t *testing.T) {
	withSharedDirectories(t)
	owned := filepath.Join(t.TempDir(), "hugo")
	writeFile(t, filepath.Join(owned, "old-file"), "old")

	proceed, err := checkExistingDirectory(owned, true)
	if err != nil || !proceed {
		t.Fatalf("proceed=%v err=%v", proceed, err)
	}
	assertNotExists(t, owned)
}

func TestPrepareSharedDirectory_RemovesOnlyArchiveDirectories(t *testing.T) {
	work := t.TempDir()
	shared := filepath.Join(work, "usr-local")
	withSharedDirectories(t, shared)

	// Previous Go installation plus unrelated content in the shared directory
	writeFile(t, filepath.Join(shared, "go", "stale-file"), "old")
	writeFile(t, filepath.Join(shared, "bin", "other-tool"), "keep")
	writeFile(t, filepath.Join(shared, "unrelated"), "keep")

	archivePath := filepath.Join(work, "go.tar.gz")
	writeTestTarGz(t, archivePath, map[string]string{"go/bin/go": "new-go", "LICENSE": "x"})

	proceed, err := prepareSharedDirectory(archivePath, shared, true)
	if err != nil || !proceed {
		t.Fatalf("proceed=%v err=%v", proceed, err)
	}
	assertNotExists(t, filepath.Join(shared, "go", "stale-file"))
	assertExists(t, filepath.Join(shared, "bin", "other-tool"))
	assertExists(t, filepath.Join(shared, "unrelated"))
}

func TestPrepareSharedDirectory_KeepsSharedSubdirectoryFromArchive(t *testing.T) {
	work := t.TempDir()
	shared := filepath.Join(work, "usr-local")
	sharedBin := filepath.Join(shared, "bin")
	withSharedDirectories(t, shared, sharedBin)
	writeFile(t, filepath.Join(sharedBin, "other-tool"), "keep")

	// Archive with a top-level bin/ extracted into /usr/local must not wipe /usr/local/bin
	archivePath := filepath.Join(work, "protoc.tar.gz")
	writeTestTarGz(t, archivePath, map[string]string{"bin/protoc": "x"})

	proceed, err := prepareSharedDirectory(archivePath, shared, true)
	if err != nil || !proceed {
		t.Fatalf("proceed=%v err=%v", proceed, err)
	}
	assertExists(t, filepath.Join(sharedBin, "other-tool"))
}

func TestSharedDirectory_ReinstallUpgradesInPlace(t *testing.T) {
	work := t.TempDir()
	shared := filepath.Join(work, "bin")
	withSharedDirectories(t, shared)
	writeFile(t, filepath.Join(shared, "hugo"), "hugo-0.150.1-longer-old-binary")
	writeFile(t, filepath.Join(shared, "act"), "keep")

	archivePath := filepath.Join(work, "hugo.tar.gz")
	writeTestTarGz(t, archivePath, map[string]string{"hugo": "hugo-0.167", "LICENSE": "x"})

	proceed, err := prepareSharedDirectory(archivePath, shared, true)
	if err != nil || !proceed {
		t.Fatalf("proceed=%v err=%v", proceed, err)
	}
	if err := archive.ExtractArchive(archivePath, shared); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(shared, "hugo"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hugo-0.167" {
		t.Errorf("hugo = %q, want %q", got, "hugo-0.167")
	}
	assertExists(t, filepath.Join(shared, "act"))
}
