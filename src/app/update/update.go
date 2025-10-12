package update

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// CreateBackup creates a backup of the current binary
func CreateBackup() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	backupPath := execPath + ".backup"

	// Copy current binary to backup
	if err := copyFile(execPath, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupPath, nil
}

// RestoreBackup restores the binary from backup
func RestoreBackup(backupPath string) error {
	if backupPath == "" {
		return fmt.Errorf("no backup path provided")
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// On Windows, we need to rename the files instead of overwriting
	if runtime.GOOS == "windows" {
		// First, try to rename backup to a temp location
		tempPath := execPath + ".restore"
		if err := copyFile(backupPath, tempPath); err != nil {
			return fmt.Errorf("failed to prepare restore: %w", err)
		}

		// Try to remove the current (broken) binary
		if err := os.Remove(execPath); err != nil {
			// If we can't remove it, we don't have permission
			os.Remove(tempPath)
			return fmt.Errorf("failed to remove broken binary: %w", err)
		}

		// Rename temp to original location
		if err := os.Rename(tempPath, execPath); err != nil {
			os.Remove(tempPath)
			return fmt.Errorf("failed to restore backup: %w", err)
		}

		// Remove backup file
		os.Remove(backupPath)
	} else {
		// On Unix-like systems, we can overwrite
		if err := copyFile(backupPath, execPath); err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
	}

	return nil
}

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

	// Check if we have write permission
	if err := checkWritePermission(execPath); err != nil {
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

// applyUpdateUnix applies update on Unix-like systems
func applyUpdateUnix(execPath, newBinaryPath string) error {
	// Open new binary
	newBinary, err := os.Open(newBinaryPath)
	if err != nil {
		return fmt.Errorf("failed to open new binary: %w", err)
	}
	defer newBinary.Close()

	// Get file info for permissions
	info, err := newBinary.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat new binary: %w", err)
	}

	// Create temporary file in the same directory
	dir := filepath.Dir(execPath)
	tmpFile, err := os.CreateTemp(dir, ".portunix-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Copy new binary to temp file
	if _, err := io.Copy(tmpFile, newBinary); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to copy new binary: %w", err)
	}
	tmpFile.Close()

	// Set executable permissions
	if err := os.Chmod(tmpPath, info.Mode()|0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

// applyUpdateWindows applies update on Windows
func applyUpdateWindows(execPath, newBinaryPath string) error {
	// Try direct replacement first (works if we have admin rights)
	tempPath := execPath + ".new"

	// Copy new binary to temporary location
	if err := copyFile(newBinaryPath, tempPath); err != nil {
		return fmt.Errorf("failed to copy new binary: %w", err)
	}

	// Set executable permissions
	if err := os.Chmod(tempPath, 0755); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Try to rename current binary to backup
	backupPath := execPath + ".old"
	if err := os.Rename(execPath, backupPath); err != nil {
		// If we can't rename, we don't have permission - create PowerShell fallback
		os.Remove(tempPath)

		// Create PowerShell update script as fallback
		psScriptPath := execPath + ".update.ps1"
		psScript := fmt.Sprintf(`# Portunix Update Fallback Script
Write-Host "Finalizing Portunix update..." -ForegroundColor Cyan
Write-Host "This may take a few seconds..." -ForegroundColor Yellow
Start-Sleep -Seconds 2

try {
    # Remove old binary
    if (Test-Path "%s") {
        Remove-Item "%s" -Force
    }
    
    # Move new binary to final location
    Move-Item "%s" "%s" -Force
    
    Write-Host "✓ Update completed successfully!" -ForegroundColor Green
    Write-Host "You can now run: portunix --version" -ForegroundColor Green
    
} catch {
    Write-Error "Update failed: $_"
    Write-Host "Please run this script as Administrator" -ForegroundColor Red
}

# Clean up
Remove-Item $PSCommandPath -Force
Read-Host "Press Enter to close"
`, execPath, execPath, newBinaryPath, execPath)

		if err := os.WriteFile(psScriptPath, []byte(psScript), 0644); err == nil {
			return fmt.Errorf("permission denied\n  Cannot write to %s\n  \n  Alternative: Run this PowerShell script as Administrator:\n  %s\n  \n  Or try running as administrator: Right-click cmd.exe -> Run as administrator", execPath, psScriptPath)
		}

		return fmt.Errorf("permission denied\n  Cannot write to %s\n  Try running as administrator", execPath)
	}

	// Rename new binary to original location
	if err := os.Rename(tempPath, execPath); err != nil {
		// If this fails, try to restore the backup
		os.Rename(backupPath, execPath)
		os.Remove(tempPath)
		return fmt.Errorf("failed to install update: %w", err)
	}

	// Remove backup on success
	os.Remove(backupPath)
	return nil
}

// checkWritePermission checks if we have write permission to a file
func checkWritePermission(path string) error {
	// Try to open the file for writing (without truncating)
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	file.Close()
	return nil
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
					os.Remove(backup)
				}
				return fmt.Errorf("failed to backup %s: %w", binary.Name, err)
			}
			backups[binary.Name] = backupPath
		}
	}

	// Now apply updates for all binaries
	for _, binary := range binaries {
		targetPath := filepath.Join(targetDir, binary.Name+binSuffix)

		if err := copyFile(binary.Path, targetPath); err != nil {
			// Restore all backups on failure
			for name, backupPath := range backups {
				targetPath := filepath.Join(targetDir, name+binSuffix)
				copyFile(backupPath, targetPath)
			}
			return fmt.Errorf("failed to update %s: %w", binary.Name, err)
		}

		// Set executable permissions on Unix
		if runtime.GOOS != "windows" {
			if err := os.Chmod(targetPath, 0755); err != nil {
				// Restore all backups on failure
				for name, backupPath := range backups {
					targetPath := filepath.Join(targetDir, name+binSuffix)
					copyFile(backupPath, targetPath)
				}
				return fmt.Errorf("failed to set permissions for %s: %w", binary.Name, err)
			}
		}

		// Clean up temporary file
		os.Remove(binary.Path)
	}

	// Clean up backup files on success
	for _, backupPath := range backups {
		os.Remove(backupPath)
	}

	return nil
}
