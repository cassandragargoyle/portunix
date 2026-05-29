/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"portunix.ai/portunix/src/helpers/ptx-github/internal/git"
)

var checkoutFlagPath string

func newCheckoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout <ref>",
		Short: "Checkout a branch, tag, or commit in a local clone",
		Long: `Switch the working tree of a previously-cloned repository to a different
branch, tag, or commit hash. Without --path, the current directory is used
(walking upward to find the .git root).

Examples:
  portunix github checkout main
  portunix github checkout v1.5.1
  portunix github checkout --path /opt/repo a1b2c3d`,
		Args: cobra.ExactArgs(1),
		RunE: runCheckout,
	}
	cmd.Flags().StringVar(&checkoutFlagPath, "path", "",
		"path to the repository (defaults to the current directory)")
	return cmd
}

func runCheckout(cmd *cobra.Command, args []string) error {
	ref := args[0]
	repoPath := checkoutFlagPath
	if repoPath == "" {
		root, err := git.FindRepoRoot(".")
		if err != nil {
			return err
		}
		repoPath = root
	}
	gm := git.NewManager("", nil)
	if err := gm.Checkout(repoPath, ref); err != nil {
		return fmt.Errorf("checkout %q: %w", ref, err)
	}
	fmt.Printf("Switched %s to %s\n", repoPath, ref)
	return nil
}
