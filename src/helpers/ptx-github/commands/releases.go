/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
)

var (
	releasesFlagLimit  int
	releasesFlagAssets bool
)

func newReleasesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "releases <owner/repo|url>",
		Short: "List releases for a GitHub repository",
		Long: `List releases for a GitHub repository.

Examples:
  portunix github releases cassandragargoyle/portunix
  portunix github releases cassandragargoyle/portunix --limit 5 --assets
  portunix github releases cassandragargoyle/portunix --json`,
		Args: cobra.ExactArgs(1),
		RunE: runReleases,
	}
	cmd.Flags().IntVar(&releasesFlagLimit, "limit", 0, "show only the first N releases (0 = all)")
	cmd.Flags().BoolVar(&releasesFlagAssets, "assets", false, "include the asset list for each release")
	return cmd
}

func runReleases(cmd *cobra.Command, args []string) error {
	owner, repo, err := parseOwnerRepoFlexible(args[0])
	if err != nil {
		return err
	}
	c, err := newClient(context.Background())
	if err != nil {
		return err
	}
	releases, err := c.ListReleases(owner, repo)
	if err != nil {
		return err
	}
	if releasesFlagLimit > 0 && releasesFlagLimit < len(releases) {
		releases = releases[:releasesFlagLimit]
	}

	if flagJSON {
		out := make([]map[string]any, 0, len(releases))
		for _, r := range releases {
			out = append(out, releaseToMap(r, releasesFlagAssets))
		}
		return writeJSON(out)
	}

	if len(releases) == 0 {
		fmt.Printf("No releases for %s/%s\n", owner, repo)
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "TAG\tNAME\tPUBLISHED\tASSETS\tPRE\tDRAFT")
	for _, r := range releases {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\n",
			r.GetTagName(),
			truncate(r.GetName(), 40),
			r.GetPublishedAt().Format("2006-01-02"),
			len(r.Assets),
			yesNo(r.GetPrerelease()),
			yesNo(r.GetDraft()),
		)
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	if releasesFlagAssets {
		fmt.Println()
		for _, r := range releases {
			fmt.Printf("=== %s ===\n", r.GetTagName())
			for _, a := range r.Assets {
				fmt.Printf("  %s  (%s)\n", a.GetName(), formatBytes(int64(a.GetSize())))
			}
		}
	}
	return nil
}

func releaseToMap(r *github.RepositoryRelease, withAssets bool) map[string]any {
	m := map[string]any{
		"tag":          r.GetTagName(),
		"name":         r.GetName(),
		"published_at": r.GetPublishedAt().Format("2006-01-02T15:04:05Z"),
		"prerelease":   r.GetPrerelease(),
		"draft":        r.GetDraft(),
		"asset_count":  len(r.Assets),
		"url":          r.GetHTMLURL(),
	}
	if withAssets {
		assets := make([]map[string]any, 0, len(r.Assets))
		for _, a := range r.Assets {
			assets = append(assets, map[string]any{
				"name":         a.GetName(),
				"size":         a.GetSize(),
				"download_url": a.GetBrowserDownloadURL(),
				"content_type": a.GetContentType(),
			})
		}
		m["assets"] = assets
	}
	return m
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "-"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
