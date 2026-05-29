/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import "github.com/spf13/cobra"

var (
	version     = "dev"
	displayName = "ptx-specpm" // overridden by main when invoked via dispatcher
)

// SetVersion is called from main to inject the build-time version string.
func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

// SetDisplayName is called from main to set the user-facing brand string
// according to invocation context (direct vs. via portunix dispatcher).
func SetDisplayName(name string) {
	displayName = name
	rootCmd.SetVersionTemplate(name + " version {{.Version}}\n")
}

var rootCmd = &cobra.Command{
	Use:   "ptx-specpm",
	Short: "Initialize project-management specifications using spec-kit-pm",
	Long: `ptx-specpm is the Portunix helper for the spec-kit-pm framework. It is
registered with the dispatcher as 'portunix specpm' and bootstraps the
spec-kit-pm scaffold inside a working directory so AI agents can drive
project-management workflows (Charter, plan, risks, decisions, KPIs,
reports) via /specpm.* slash-commands or skills.

The kit content lives upstream at CassandraGargoyle/spec-kit-pm and is
fetched at runtime via shallow clone (default) or read from a local path
(forks / air-gapped). Subsequent runs of the same kit ref are served from
a per-ref cache and never hit the network.

This binary is typically invoked by the main portunix dispatcher and
should not be used directly.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command. Returns the underlying error so main can
// decide on the exit code.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd, checkCmd, versionCmd, integrationCmd, upgradeCmd)
}
