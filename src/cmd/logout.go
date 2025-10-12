package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logs out from a Portunix daemon.",
	Long:  `Logs out from the currently connected Portunix daemon.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Successfully logged out.")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
