/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// sharedDirectoriesFunc lists directories shared by many programs (replaceable in tests)
var sharedDirectoriesFunc = sharedDirectories

// sharedDirectories returns well-known directories that never belong to a single
// package, so the installer must never remove them (issue #226)
func sharedDirectories() []string {
	home, _ := os.UserHomeDir()

	if runtime.GOOS == "windows" {
		var dirs []string
		for _, env := range []string{"SystemRoot", "ProgramFiles", "ProgramFiles(x86)", "ProgramW6432",
			"ProgramData", "LOCALAPPDATA", "APPDATA", "USERPROFILE", "PUBLIC"} {
			if v := os.Getenv(env); v != "" {
				dirs = append(dirs, v)
			}
		}
		if root := os.Getenv("SystemRoot"); root != "" {
			dirs = append(dirs, filepath.Join(root, "System32"))
		}
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			dirs = append(dirs, filepath.Join(local, "Programs"))
		}
		dirs = append(dirs, `C:\Program Files`, `C:\Program Files (x86)`, `C:\ProgramData`, `C:\Windows`,
			`C:\Windows\System32`, `C:\Users`, `C:\Portunix`, `C:\Portunix\bin`)
		if home != "" {
			dirs = append(dirs, home)
		}
		return dirs
	}

	dirs := []string{
		"/", "/bin", "/sbin", "/lib", "/lib64", "/etc", "/var", "/tmp", "/opt", "/home", "/root",
		"/usr", "/usr/bin", "/usr/sbin", "/usr/lib", "/usr/lib64", "/usr/share",
		"/usr/local", "/usr/local/bin", "/usr/local/sbin", "/usr/local/lib", "/usr/local/share",
		"/usr/local/include", "/usr/local/opt",
		"/opt/homebrew", "/opt/homebrew/bin", "/Applications",
	}
	if home != "" {
		dirs = append(dirs, home,
			filepath.Join(home, "bin"),
			filepath.Join(home, ".local"),
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, ".local", "share"),
			filepath.Join(home, ".config"),
			filepath.Join(home, "Applications"))
	}
	return dirs
}

// isSharedDirectory reports whether dir is a filesystem root or one of the well-known
// shared directories; symlinks are resolved on both sides before comparing
func isSharedDirectory(dir string) bool {
	candidates := normalizedPaths(dir)
	for _, c := range candidates {
		if isFilesystemRoot(c) {
			return true
		}
	}

	for _, shared := range sharedDirectoriesFunc() {
		for _, s := range normalizedPaths(shared) {
			for _, c := range candidates {
				if c == s {
					return true
				}
			}
		}
	}
	return false
}

// normalizedPaths returns the cleaned absolute path and, if different, its symlink-resolved form
func normalizedPaths(path string) []string {
	if path == "" {
		return nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = filepath.Clean(path)
	}
	paths := []string{normalizeCase(abs)}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil && resolved != abs {
		paths = append(paths, normalizeCase(resolved))
	}
	return paths
}

// normalizeCase makes Windows paths compare case-insensitively
func normalizeCase(path string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(filepath.Clean(path))
	}
	return filepath.Clean(path)
}

// isFilesystemRoot reports "/" or a Windows volume root such as "C:\"
func isFilesystemRoot(path string) bool {
	clean := filepath.Clean(path)
	return filepath.Dir(clean) == clean
}
