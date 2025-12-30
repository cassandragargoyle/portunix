package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RunRm executes the rm command
func RunRm(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: rm <path>")
	}

	path := args[0]

	// Handle wildcards
	if strings.Contains(path, "*") {
		matches, err := filepath.Glob(path)
		if err != nil {
			// Invalid pattern - not an error, just nothing to do
			return nil
		}

		for _, match := range matches {
			if err := os.RemoveAll(match); err != nil {
				// Ignore errors for individual files
				continue
			}
		}
		return nil
	}

	// Single path removal - ignore if doesn't exist
	err := os.RemoveAll(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
