package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// RunChecksum executes the checksum command
func RunChecksum(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: checksum <directory> [output-file]")
	}

	dir := args[0]

	// Determine output file
	outputFile := filepath.Join(dir, "checksums.sha256")
	if len(args) > 1 {
		outputFile = args[1]
	}

	// Ensure directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("directory not found: %s", dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Build checksums
	var checksums []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		filePath := filepath.Join(dir, entry.Name())
		hash, err := hashFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to hash %s: %w", entry.Name(), err)
		}

		// Format: hash  filename (two spaces, sha256sum compatible)
		checksums = append(checksums, fmt.Sprintf("%s  %s", hash, entry.Name()))
	}

	// Write output file
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	for _, line := range checksums {
		fmt.Fprintln(file, line)
	}

	return nil
}

// hashFile computes SHA256 hash of a file
func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
