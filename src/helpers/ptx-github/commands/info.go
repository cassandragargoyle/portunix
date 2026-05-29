/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <owner/repo|url>",
		Short: "Show metadata for a GitHub repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, repo, err := parseOwnerRepoFlexible(args[0])
			if err != nil {
				return err
			}
			c, err := newClient(context.Background())
			if err != nil {
				return err
			}
			r, err := c.Repository(owner, repo)
			if err != nil {
				return err
			}

			latestTag := ""
			if rel, err := c.LatestRelease(owner, repo); err == nil {
				latestTag = rel.GetTagName()
			}

			if flagJSON {
				return writeJSON(map[string]any{
					"full_name":      r.GetFullName(),
					"description":    r.GetDescription(),
					"default_branch": r.GetDefaultBranch(),
					"language":       r.GetLanguage(),
					"stars":          r.GetStargazersCount(),
					"forks":          r.GetForksCount(),
					"open_issues":    r.GetOpenIssuesCount(),
					"updated_at":     r.GetUpdatedAt().Format("2006-01-02T15:04:05Z"),
					"clone_url":      r.GetCloneURL(),
					"ssh_url":        r.GetSSHURL(),
					"latest_tag":     latestTag,
				})
			}

			fmt.Printf("Repository:    %s\n", r.GetFullName())
			if d := r.GetDescription(); d != "" {
				fmt.Printf("Description:   %s\n", d)
			}
			fmt.Printf("Default:       %s\n", r.GetDefaultBranch())
			fmt.Printf("Language:      %s\n", emptyOr(r.GetLanguage(), "-"))
			fmt.Printf("Stars:         %d\n", r.GetStargazersCount())
			fmt.Printf("Forks:         %d\n", r.GetForksCount())
			fmt.Printf("Open issues:   %d\n", r.GetOpenIssuesCount())
			fmt.Printf("Updated:       %s\n", r.GetUpdatedAt().Format("2006-01-02"))
			fmt.Printf("Clone URL:     %s\n", r.GetCloneURL())
			fmt.Printf("SSH URL:       %s\n", r.GetSSHURL())
			if latestTag != "" {
				fmt.Printf("Latest tag:    %s\n", latestTag)
			}
			return nil
		},
	}
}
