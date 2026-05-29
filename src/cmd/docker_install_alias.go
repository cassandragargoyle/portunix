/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// dockerInstallCmd is a thin alias that forwards `portunix docker install`
// to the canonical `portunix install docker` form. The root dispatcher then
// routes the request to the ptx-installer helper, which contains the actual
// installation logic (storage analysis, UAC elevation, WSL/Hyper-V prereqs).
//
// Both invocations therefore share a single code path (Issue #019,
// "Inconsistent Commands"). Flag parsing is disabled so user-supplied flags
// (--dry-run, --data-root, --yes, ...) pass through unchanged.
var dockerInstallCmd = &cobra.Command{
	Use:   "install [flags]",
	Short: "Install Docker (alias for 'portunix install docker').",
	Long: `Install Docker on the current system.

This command is an alias for the canonical form 'portunix install docker'.
Both forms invoke the same ptx-installer helper and behave identically.

Supported flags are forwarded as-is:
  --dry-run               Preview the installation without making changes
  --data-root <path>      Explicit Docker data-root directory
  --yes                   Non-interactive mode (use recommended defaults)

Examples:
  portunix docker install
  portunix docker install --dry-run
  portunix docker install --data-root D:\docker-data
  portunix docker install --yes`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDockerInstallAlias(args)
	},
}

// runDockerInstallAlias re-executes the current portunix binary with
// `install docker <args>` so the root dispatcher can route to ptx-installer.
func runDockerInstallAlias(args []string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve portunix executable: %w", err)
	}

	forwarded := append([]string{"install", "docker"}, args...)
	c := exec.Command(exePath, forwarded...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			os.Exit(ee.ExitCode())
		}
		return err
	}
	return nil
}

func init() {
	dockerCmd.AddCommand(dockerInstallCmd)
}
