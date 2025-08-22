package update

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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
		// Remove the current (broken) binary
		if err := os.Remove(execPath); err != nil {
			return fmt.Errorf("failed to remove broken binary: %w", err)
		}
		// Rename backup to original
		if err := os.Rename(backupPath, execPath); err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
	} else {
		// On Unix-like systems, we can overwrite
		if err := copyFile(backupPath, execPath); err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
	}
	
	return nil
}

// ApplyUpdate replaces the current binary with the new one
func ApplyUpdate(newBinaryPath string) error {
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
	
	// On Windows, we need special handling because we can't replace a running executable
	if runtime.GOOS == "windows" {
		return applyUpdateWindows(execPath, newBinaryPath)
	}
	
	// On Unix-like systems, we can replace the file directly
	return applyUpdateUnix(execPath, newBinaryPath)
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
	// On Windows, we create a batch script to replace the binary after the program exits
	batchPath := execPath + ".update.bat"
	
	// Create batch script
	script := fmt.Sprintf(`@echo off
echo Finalizing update...
ping 127.0.0.1 -n 2 > nul
move /Y "%s" "%s"
if %%errorlevel%% neq 0 (
    echo Update failed!
    pause
    exit /b 1
)
echo Update completed successfully!
del "%%~f0"
`, newBinaryPath, execPath)
	
	if err := os.WriteFile(batchPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("failed to create update script: %w", err)
	}
	
	fmt.Println("\nUpdate prepared. Please run the following command to complete the update:")
	fmt.Printf("  %s\n", batchPath)
	fmt.Println("\nAlternatively, close this program and run the update script manually.")
	
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