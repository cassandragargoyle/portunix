package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	appversion "portunix.cz/app/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  `Display current Portunix version and build information`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Portunix version %s\n", appversion.ProductVersion)
		fmt.Printf("Built with %s for %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}