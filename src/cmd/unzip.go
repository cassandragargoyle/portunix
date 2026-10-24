/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"fmt"
	"os"

	"portunix.ai/app"

	"github.com/spf13/cobra"
)

// unzipCmd represents the unzip command
var unzipCmd = &cobra.Command{
	Use:   "unzip",
	Short: "Extracts a ZIP file.",
	Long: `The unzip command extracts a ZIP file to a specified destination.

If no destination is provided, the file will be extracted to the same directory
as the ZIP file.

Example:
  portunix unzip --path /path/to/file.zip --destinationpath /path/to/destination`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		destinationPath, _ := cmd.Flags().GetString("destinationpath")

		if err := runUnzip(path, destinationPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// runUnzip extracts the archive at path, next to the archive when destinationPath is empty
func runUnzip(path, destinationPath string) error {
	if path == "" {
		return fmt.Errorf("please specify the path to the ZIP file with --path")
	}
	return app.UnzipFile(path, destinationPath)
}

func init() {
	rootCmd.AddCommand(unzipCmd)

	unzipCmd.Flags().String("path", "", "Path to the ZIP file to be extracted")
	unzipCmd.Flags().String("destinationpath", "", "Path to the target directory where the files should be extracted")
}
