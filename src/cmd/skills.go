/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"portunix.ai/portunix/src/dispatcher"
)

// Issue #190: `portunix skills` is executed by the ptx-installer helper via
// the dispatcher. These cobra commands exist only for shell completion and
// for a clear error when the helper is missing; they never run otherwise.

// skillsAgentIDs mirrors the agent registry in
// src/helpers/ptx-installer/engine/skills/agents.go
var skillsAgentIDs = []string{"amp", "claude", "copilot", "cursor", "gemini", "opencode", "zed"}

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Install AI-agent skills from git repositories",
	Long:  "Install AI-agent skills (SKILL.md bundles) into .agents/skills/ and per-agent skills directories (handled by ptx-installer)",
	Run:   skillsHelperMissing,
}

func skillsHelperMissing(cmd *cobra.Command, args []string) {
	fmt.Fprintln(os.Stderr, "❌ 'portunix skills' requires the ptx-installer helper next to the portunix binary")
	os.Exit(1)
}

// skillsSubcommand builds a completion-only subcommand
func skillsSubcommand(use, short string) *cobra.Command {
	return &cobra.Command{Use: use, Short: short, Run: skillsHelperMissing}
}

// completeAgentIDs offers the agent IDs for --agent
func completeAgentIDs(withUniversal bool) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		ids := skillsAgentIDs
		if withUniversal {
			ids = append([]string{"universal"}, ids...)
		}
		return ids, cobra.ShellCompDirectiveNoFileComp
	}
}

// completeInstalledSkills asks the helper for the installed skill names
func completeInstalledSkills(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	helperPath, ok := dispatcher.NewDispatcher(rootCmd.Version).ShouldDispatch([]string{"skills"})
	if !ok {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	helperArgs := []string{"skills", "list", "--json"}
	if global, _ := cmd.Flags().GetBool("global"); global {
		helperArgs = append(helperArgs, "--global")
	}
	out, err := exec.Command(helperPath, helperArgs...).Output() // #nosec G204 -- helper path resolved by the dispatcher
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var entries []struct {
		Name    string `json:"name"`
		Managed bool   `json:"managed"`
	}
	if json.Unmarshal(out, &entries) != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	seen := map[string]bool{}
	var names []string
	for _, e := range entries {
		if e.Managed && !seen[e.Name] {
			seen[e.Name] = true
			names = append(names, e.Name)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	add := skillsSubcommand("add <repo-url|owner/repo|path>", "Install skills from a git repository or local path")
	add.Flags().StringSliceP("skill", "s", nil, "Skill to install (repeatable)")
	add.Flags().Bool("all", false, "Install every discovered skill")
	add.Flags().StringSliceP("agent", "a", nil, "Also install for this agent (repeatable)")
	add.Flags().Bool("all-agents", false, "Install for every supported agent")
	add.Flags().Bool("no-universal", false, "Skip the universal .agents/skills directory")
	add.Flags().String("ref", "", "Branch, tag or commit SHA")
	add.Flags().String("source", "", "Install from a local checkout")
	add.Flags().BoolP("global", "g", false, "Install under the home directory")
	add.Flags().Bool("user", false, "Alias of --global")
	add.Flags().BoolP("yes", "y", false, "Trust a new source without asking")
	add.Flags().BoolP("force", "f", false, "Overwrite skills that differ from the source")
	_ = add.RegisterFlagCompletionFunc("agent", completeAgentIDs(false))
	_ = add.MarkFlagDirname("source")

	list := skillsSubcommand("list", "List installed skills")
	list.Flags().BoolP("global", "g", false, "Use the user scope")
	list.Flags().Bool("json", false, "JSON output")

	remove := skillsSubcommand("remove <skill-name>", "Remove a Portunix-installed skill")
	remove.ValidArgsFunction = completeInstalledSkills
	remove.Flags().StringSliceP("agent", "a", nil, "Limit removal to this target")
	remove.Flags().BoolP("global", "g", false, "Use the user scope")
	_ = remove.RegisterFlagCompletionFunc("agent", completeAgentIDs(true))

	update := skillsSubcommand("update [<skill-name>...]", "Re-install managed skills from their recorded source")
	update.ValidArgsFunction = completeInstalledSkills
	update.Flags().BoolP("global", "g", false, "Use the user scope")
	update.Flags().String("ref", "", "Update to a different ref")
	update.Flags().BoolP("force", "f", false, "Overwrite locally modified skills")

	sources := skillsSubcommand("sources", "Show install targets and agent detection")
	sources.Flags().Bool("json", false, "JSON output")

	trust := skillsSubcommand("trust", "Show or revoke user-wide trusted sources")
	trustList := skillsSubcommand("list", "List trusted sources")
	trustList.Flags().Bool("json", false, "JSON output")
	trust.AddCommand(trustList, skillsSubcommand("remove <source>", "Revoke trust in a source"))

	skillsCmd.AddCommand(add, list, remove, update, sources, trust)
	rootCmd.AddCommand(skillsCmd)
}
