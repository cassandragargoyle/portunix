/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"portunix.ai/portunix/src/helpers/ptx-github/internal/auth"
	"portunix.ai/portunix/src/helpers/ptx-github/internal/git"
)

var tagsFlagRemote bool

func newTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags [owner/repo|url]",
		Short: "List tags from a local repo or remote URL",
		Long: `List tags. With no argument, lists tags from the current repository.
With an argument, lists tags from the remote without cloning.

Examples:
  portunix github tags
  portunix github tags cassandragargoyle/portunix
  portunix github tags --remote https://github.com/cassandragargoyle/portunix`,
		Args: cobra.MaximumNArgs(1),
		RunE: runTags,
	}
	cmd.Flags().BoolVar(&tagsFlagRemote, "remote", false,
		"force remote listing even when arg is set (default: auto)")
	return cmd
}

func runTags(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		root, err := git.FindRepoRoot(".")
		if err != nil {
			return err
		}
		gm := git.NewManager("", nil)
		tags, err := gm.Tags(root)
		if err != nil {
			return err
		}
		return printTags(tags)
	}

	url, err := git.NormalizeRepoSpec(args[0])
	if err != nil {
		return err
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
	gm := git.NewManager(token, nil)
	tags, err := gm.RemoteTags(url)
	if err != nil {
		return err
	}
	return printTags(tags)
}

func printTags(tags []string) error {
	if flagJSON {
		return writeJSON(map[string]any{"tags": tags})
	}
	if len(tags) == 0 {
		fmt.Println("(no tags)")
		return nil
	}
	for _, t := range tags {
		fmt.Println(t)
	}
	return nil
}
