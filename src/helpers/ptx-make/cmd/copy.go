package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// RunCopy executes the copy command
func RunCopy(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: copy <source> <destination>")
	}

	src := args[0]
	dst := args[1]

	// Handle wildcards
	if strings.Contains(src, "*") {
		matches, err := filepath.Glob(src)
		if err != nil {
			return fmt.Errorf("invalid pattern: %w", err)
		}
		if len(matches) == 0 {
			return fmt.Errorf("no files match pattern: %s", src)
		}

		// Ensure destination directory exists
		if err := os.MkdirAll(dst, 0755); err != nil {
			return fmt.Errorf("failed to create destination: %w", err)
		}

		for _, match := range matches {
			destPath := filepath.Join(dst, filepath.Base(match))
			if err := copyPath(match, destPath); err != nil {
				return fmt.Errorf("failed to copy %s: %w", match, err)
			}
		}
		return nil
	}

	// Single file/directory copy
	return copyPath(src, dst)
}

// copyPath copies a file or directory from src to dst
func copyPath(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source not found: %w", err)
	}

	if srcInfo.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	// If destination is a directory, use source filename
	dstInfo, err := os.Stat(dst)
	if err == nil && dstInfo.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	} else {
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
