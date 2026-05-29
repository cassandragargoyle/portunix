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

var statusFlagPath string

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of the local Git repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := statusFlagPath
			if path == "" {
				root, err := git.FindRepoRoot(".")
				if err != nil {
					return err
				}
				path = root
			}
			gm := git.NewManager("", nil)
			st, err := gm.Status(path)
			if err != nil {
				return err
			}

			if flagJSON {
				return writeJSON(map[string]any{
					"path":       st.Path,
					"branch":     st.Branch,
					"head":       st.Head,
					"clean":      st.Clean,
					"modified":   st.Modified,
					"untracked":  st.Untracked,
					"added":      st.Added,
					"deleted":    st.Deleted,
					"remote_url": st.RemoteURL,
				})
			}

			fmt.Printf("Path:       %s\n", st.Path)
			fmt.Printf("Branch:     %s\n", emptyOr(st.Branch, "-"))
			fmt.Printf("HEAD:       %s\n", short(st.Head))
			if st.RemoteURL != "" {
				fmt.Printf("Remote:     %s\n", st.RemoteURL)
			}
			if st.Clean {
				fmt.Println("State:      clean")
			} else {
				fmt.Printf("State:      %d modified, %d added, %d deleted, %d untracked\n",
					st.Modified, st.Added, st.Deleted, st.Untracked)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&statusFlagPath, "path", "", "repository path (default: walk up from .)")
	return cmd
}

func short(hash string) string {
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}
