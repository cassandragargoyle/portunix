/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package commands

import (
	"github.com/spf13/cobra"
)

// Persistent flags shared by all "github" subcommands.
var (
	flagAccount   string
	flagPassword  bool
	flagJSON      bool
	flagStorePath string
)

// NewGithubCmd returns the "github" subcommand tree mounted under ptx-github.
// The dispatcher invokes "ptx-github github ..." so the user-facing command
// is "portunix github ...".
func NewGithubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "github",
		Short: "GitHub repository, releases, and authentication operations",
		Long: `Work with GitHub from the command line.

Examples:
  portunix github clone cassandragargoyle/portunix
  portunix github releases cassandragargoyle/portunix
  portunix github download cassandragargoyle/portunix v1.5.1 --asset portunix-linux-amd64.tar.gz
  portunix github info cassandragargoyle/portunix
  portunix github status                # status of the current repo
  portunix github auth login            # store a personal access token

Authentication is read from (in order): --account, $PORTUNIX_GITHUB_TOKEN,
$GITHUB_TOKEN, $GH_TOKEN, the encrypted store at ~/.portunix/github/auth.json.`,
	}

	cmd.PersistentFlags().StringVar(&flagAccount, "account", "",
		"named account from the encrypted store (defaults to env / default account)")
	cmd.PersistentFlags().BoolVar(&flagPassword, "password", false,
		"prompt for password to unlock the encrypted token store")
	cmd.PersistentFlags().BoolVar(&flagJSON, "json", false,
		"output machine-readable JSON")
	cmd.PersistentFlags().StringVar(&flagStorePath, "store", "",
		"override path to the encrypted token store")

	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newCloneCmd())
	cmd.AddCommand(newCheckoutCmd())
	cmd.AddCommand(newReleasesCmd())
	cmd.AddCommand(newDownloadCmd())
	cmd.AddCommand(newInfoCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newTagsCmd())

	return cmd
}
