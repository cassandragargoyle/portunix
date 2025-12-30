package cmd

import (
	"fmt"
	"os"
)

// RunExists executes the exists command
func RunExists(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: exists <path>")
		os.Exit(1)
	}

	path := args[0]

	_, err := os.Stat(path)
	if err != nil {
		os.Exit(1)
	}
	// Path exists - exit code 0 (default)
}
