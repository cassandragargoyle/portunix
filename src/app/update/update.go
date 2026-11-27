/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package update

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// ErrRestoreFailed marks a failed update whose rollback left some binaries
// unrestored, their *.backup files are kept
var ErrRestoreFailed = errors.New("restoring backups failed")

// ApplyUpdate replaces the current binary and helper binaries with new ones
func ApplyUpdate(archivePath string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Check write permission on the directory, the running binary itself
	// cannot be opened for writing (ETXTBSY on Linux)
	if err := checkWritePermission(filepath.Dir(execPath)); err != nil {
		if runtime.GOOS != "windows" {
			return fmt.Errorf("permission denied\n  Cannot write to %s\n  Try running with sudo: sudo portunix update", execPath)
		}
		return fmt.Errorf("permission denied\n  Cannot write to %s\n  Try running as administrator", execPath)
	}

	// Extract all binaries from archive
	binaries, err := extractAllBinaries(archivePath)
	if err != nil {
		return fmt.Errorf("failed to extract binaries: %w", err)
	}

	// Apply update for all binaries
	targetDir := filepath.Dir(execPath)
	if err := applyUpdateAllBinaries(targetDir, binaries); err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	return nil
}

// replaceBinary installs src at target by renaming instead of overwriting, so
// a running target can be replaced (ETXTBSY on Linux, sharing violation on
// Windows), the previous file is moved aside to target.old
func replaceBinary(src, target string) error {
	newPath := target + ".new"
	oldPath := target + ".old"

	if err := copyFile(src, newPath); err != nil {
		_ = os.Remove(newPath)
		return err
	}
	if err := os.Chmod(newPath, 0755); err != nil { // #nosec G302 -- binaries must be executable by all users
		_ = os.Remove(newPath)
		return err
	}

	// Leftover from a previous update, gone unless that process still runs
	_ = os.Remove(oldPath)

	movedAside := false
	if _, err := os.Stat(target); err == nil {
		if err := os.Rename(target, oldPath); err != nil {
			_ = os.Remove(newPath)
			return fmt.Errorf("failed to move %s aside: %w", filepath.Base(target), err)
		}
		movedAside = true
	}

	if err := os.Rename(newPath, target); err != nil {
		_ = os.Remove(newPath)
		if movedAside {
			if restoreErr := os.Rename(oldPath, target); restoreErr != nil {
				return fmt.Errorf("%w (moving %s back failed: %v)", err, filepath.Base(oldPath), restoreErr)
			}
		}
		return err
	}

	// A running binary cannot be removed on Windows, it is cleaned up by the next update
	_ = os.Remove(oldPath)
	return nil
}

// checkWritePermission checks if new files can be created in dir
func checkWritePermission(dir string) error {
	file, err := os.CreateTemp(dir, ".portunix-write-check-*")
	if err != nil {
		return err
	}
	_ = file.Close()
	return os.Remove(file.Name())
}

// IsPermissionError checks if an error is related to permissions
func IsPermissionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "přístup byl odepřen") ||
		strings.Contains(errStr, "cannot write") ||
		strings.Contains(errStr, "administrator")
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// Get source file info
	info, err := source.Stat()
	if err != nil {
		return err
	}

	destination, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// extractAllBinaries extracts all binaries from the update archive
func extractAllBinaries(archivePath string) ([]*BinaryInfo, error) {
	if runtime.GOOS == "windows" {
		return extractAllFromZip(archivePath)
	} else {
		return extractAllFromTarGz(archivePath)
	}
}

// applyUpdateAllBinaries applies update for all binaries
func applyUpdateAllBinaries(targetDir string, binaries []*BinaryInfo) error {
	// First, create backups of all existing binaries
	backups := make(map[string]string)

	binSuffix := ""
	if runtime.GOOS == "windows" {
		binSuffix = ".exe"
	}

	for _, binary := range binaries {
		targetPath := filepath.Join(targetDir, binary.Name+binSuffix)

		// Check if target exists
		if _, err := os.Stat(targetPath); err == nil {
			backupPath := targetPath + ".backup"
			if err := copyFile(targetPath, backupPath); err != nil {
				// Clean up any backups we've already created
				for _, backup := range backups {
					_ = os.Remove(backup)
				}
				return fmt.Errorf("failed to backup %s: %w", binary.Name, err)
			}
			backups[binary.Name] = backupPath
		}
	}

	// Now apply updates for all binaries
	for i, binary := range binaries {
		targetPath := filepath.Join(targetDir, binary.Name+binSuffix)

		if err := replaceBinary(binary.Path, targetPath); err != nil {
			removeExtracted(binaries[i:])

			// Roll back only the binaries already replaced, the others are untouched
			applied := make(map[string]string)
			for _, done := range binaries[:i] {
				if backupPath, ok := backups[done.Name]; ok {
					applied[done.Name] = backupPath
					delete(backups, done.Name)
				}
			}
			for _, backupPath := range backups {
				_ = os.Remove(backupPath)
			}

			if restoreErr := restoreBackups(applied, targetDir, binSuffix); restoreErr != nil {
				return fmt.Errorf("failed to update %s: %w (%w)", binary.Name, err, restoreErr)
			}
			return fmt.Errorf("failed to update %s: %w", binary.Name, err)
		}

		// Clean up temporary file
		_ = os.Remove(binary.Path)
	}

	// Clean up backup files on success
	for _, backupPath := range backups {
		_ = os.Remove(backupPath)
	}

	return nil
}

// restoreBackups puts every backup back in place of its binary and removes
// it, the backups of binaries that could not be restored are left in place
// and reported
func restoreBackups(backups map[string]string, targetDir, binSuffix string) error {
	var failed []string
	for name, backupPath := range backups {
		targetPath := filepath.Join(targetDir, name+binSuffix)
		if err := replaceBinary(backupPath, targetPath); err != nil {
			failed = append(failed, fmt.Sprintf("%s from %s: %v", name, backupPath, err))
			continue
		}
		_ = os.Remove(backupPath)
	}
	sort.Strings(failed)
	if len(failed) > 0 {
		return fmt.Errorf("%w: %s", ErrRestoreFailed, strings.Join(failed, "; "))
	}
	return nil
}
