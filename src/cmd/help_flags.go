/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

// guardHelpLevelFlags makes every subcommand show its help instead of running
// when --help-expert or --help-ai is given. Both are root persistent flags
// that only the root help function handles, so without the guard e.g.
// `sandbox generate --help-expert` generated a .wsb file (issue #203).
// Commands with DisableFlagParsing get the flags as raw arguments, so those
// are checked instead of the parsed flag values
func guardHelpLevelFlags(parent *cobra.Command) {
	for _, sub := range parent.Commands() {
		guardHelpLevelFlags(sub)
		if run := sub.Run; run != nil {
			sub.Run = func(cmd *cobra.Command, args []string) {
				if showHelpForHelpLevelFlags(cmd, args) {
					return
				}
				run(cmd, args)
			}
		}
		if runE := sub.RunE; runE != nil {
			sub.RunE = func(cmd *cobra.Command, args []string) error {
				if showHelpForHelpLevelFlags(cmd, args) {
					return nil
				}
				return runE(cmd, args)
			}
		}
	}
}

// selfHandledHelpAIAnnotation marks a command whose Run handles --help-ai
// itself (plugin commands, issue #257); the guard then only covers --help-expert
const selfHandledHelpAIAnnotation = "portunix/self-handled-help-ai"

// showHelpForHelpLevelFlags prints the command help when a help level flag
// is set and reports whether it did
func showHelpForHelpLevelFlags(cmd *cobra.Command, args []string) bool {
	expert, ai := helpExpert, helpAI
	if cmd.DisableFlagParsing {
		expert, ai = hasRawHelpFlag(args, "--help-expert"), hasRawHelpFlag(args, "--help-ai")
	}
	if !expert && !ai {
		return false
	}
	if ai && cmd.Annotations[selfHandledHelpAIAnnotation] == "true" {
		return false
	}
	_ = cmd.Help()
	return true
}

// hasRawHelpFlag reports whether raw arguments contain the given boolean
// flag (not set to false) before a "--" separator
func hasRawHelpFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		name, value, hasValue := strings.Cut(arg, "=")
		if name == flag && (!hasValue || value != "false") {
			return true
		}
	}
	return false
}
