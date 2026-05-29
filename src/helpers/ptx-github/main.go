/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"portunix.ai/portunix/src/helpers/ptx-github/commands"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "ptx-github",
	Short: "Portunix GitHub integration helper",
	Long: `ptx-github is a helper binary that provides GitHub integration for Portunix.

It is invoked by the main portunix dispatcher when the user runs
"portunix github ...". It supports:

  - Cloning repositories (pure-Go via go-git)
  - Listing and downloading GitHub releases and assets
  - Repository metadata lookup
  - Authentication management with encrypted token storage

This binary is normally not used directly.`,
	Version: version,
}

func init() {
	rootCmd.SetVersionTemplate("ptx-github version {{.Version}}\n")

	rootCmd.AddCommand(commands.NewGithubCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
