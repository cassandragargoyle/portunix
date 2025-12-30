package cmd

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunVersion executes the version command for git tag info
func RunVersion(args []string) error {
	// Try git describe with tags
	gitCmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	output, err := gitCmd.Output()
	if err != nil {
		// Git not available or not a repo - try just commit hash
		return RunCommit(args)
	}

	version := strings.TrimSpace(string(output))
	fmt.Println(version)
	return nil
}
