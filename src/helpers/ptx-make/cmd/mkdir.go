package cmd

import (
	"fmt"
	"os"
)

// RunMkdir executes the mkdir command
func RunMkdir(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mkdir <path>")
	}
	path := args[0]
	return os.MkdirAll(path, 0755)
}
