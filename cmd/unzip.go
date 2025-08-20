package cmd

import (
	"fmt"

	"portunix.cz/app"

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

		if path == "" {
			fmt.Println("Please specify the path to the ZIP file.")
			return
		}

		var arguments []string
		arguments = append(arguments, fmt.Sprintf("--path=%s", path))
		if destinationPath != "" {
			arguments = append(arguments, fmt.Sprintf("--destinationpath=%s", destinationPath))
		}

		app.Unzip(arguments)
	},
}

func init() {
	rootCmd.AddCommand(unzipCmd)

	unzipCmd.Flags().String("path", "", "Path to the ZIP file to be extracted")
	unzipCmd.Flags().String("destinationpath", "", "Path to the target directory where the files should be extracted")
}