/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunCommit executes the commit command
func RunCommit(args []string) error {
	gitCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := gitCmd.Output()
	if err != nil {
		// Not a git repo or git not available
		fmt.Println("unknown")
		return nil
	}

	commit := strings.TrimSpace(string(output))
	fmt.Println(commit)
	return nil
}
