/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"portunix.ai/portunix/src/helpers/ptx-specpm/internal/kit"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print helper version and bundled spec-kit-pm pinned ref",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s version %s\n", displayName, version)
		fmt.Printf("kit (default pinned ref) %s\n", kit.DefaultPinnedRef)
		return nil
	},
}
