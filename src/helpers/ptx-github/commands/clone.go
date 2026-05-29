/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"portunix.ai/portunix/src/helpers/ptx-github/internal/auth"
	"portunix.ai/portunix/src/helpers/ptx-github/internal/git"
)

var (
	cloneFlagBranch     string
	cloneFlagDepth      int
	cloneFlagBare       bool
	cloneFlagRecurse    bool
	cloneFlagNoCheckout bool
	cloneFlagQuiet      bool
)

func newCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone <owner/repo|url> [path]",
		Short: "Clone a GitHub repository (pure Go, no external git binary)",
		Long: `Clone a repository using go-git. Authentication is taken from the
configured account or environment variables — the same as for the API
commands. SSH URLs are passed through, but only HTTPS clones use the
stored token.

Examples:
  portunix github clone cassandragargoyle/portunix
  portunix github clone cassandragargoyle/portunix ./my-portunix
  portunix github clone cassandragargoyle/portunix --branch dev --depth 1
  portunix github clone https://github.com/cassandragargoyle/portunix.git`,
		Args: cobra.RangeArgs(1, 2),
		RunE: runClone,
	}
	cmd.Flags().StringVarP(&cloneFlagBranch, "branch", "b", "", "branch or tag to checkout")
	cmd.Flags().IntVar(&cloneFlagDepth, "depth", 0, "create a shallow clone with this many commits")
	cmd.Flags().BoolVar(&cloneFlagBare, "bare", false, "create a bare repository")
	cmd.Flags().BoolVar(&cloneFlagRecurse, "recurse-submodules", false, "clone submodules recursively")
	cmd.Flags().BoolVar(&cloneFlagNoCheckout, "no-checkout", false, "do not checkout HEAD after clone")
	cmd.Flags().BoolVarP(&cloneFlagQuiet, "quiet", "q", false, "suppress progress output")
	return cmd
}

func runClone(cmd *cobra.Command, args []string) error {
	url, err := git.NormalizeRepoSpec(args[0])
	if err != nil {
		return err
	}

	dest := ""
	if len(args) >= 2 {
		dest = args[1]
	} else {
		_, repo, err := git.SplitOwnerRepo(url)
		if err != nil {
			return err
		}
		dest = filepath.Clean(repo)
	}

	mgr, err := authManager()
	if err != nil {
		return err
	}
	resolved, _ := mgr.Resolve(flagAccount)
	token := ""
	if resolved != nil && resolved.Source != auth.SourceNone {
		token = resolved.Token
	}

	var progress io.Writer
	if !cloneFlagQuiet {
		progress = os.Stderr
	}

	gm := git.NewManager(token, progress)
	if err := gm.Clone(git.CloneOptions{
		URL:        url,
		Dest:       dest,
		Branch:     cloneFlagBranch,
		Depth:      cloneFlagDepth,
		Bare:       cloneFlagBare,
		Recurse:    cloneFlagRecurse,
		NoCheckout: cloneFlagNoCheckout,
	}); err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	fmt.Printf("Cloned %s into %s\n", url, dest)
	return nil
}
