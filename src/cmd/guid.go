package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"portunix.ai/app/guid"
)

var guidCmd = &cobra.Command{
	Use:    "guid",
	Short:  "Generate and validate GUIDs/UUIDs",
	Long:   `Generate random or deterministic GUIDs/UUIDs and validate UUID format`,
	Hidden: true, // Hidden from standard help - only visible in expert and AI help
	Run: func(cmd *cobra.Command, args []string) {
		// Show help when no subcommand is provided
		cmd.Help()
	},
}

var guidRandomCmd = &cobra.Command{
	Use:   "random",
	Short: "Generate a random UUID v4",
	Long:  `Generate a cryptographically secure random UUID v4 following RFC 4122 standard`,
	Run: func(cmd *cobra.Command, args []string) {
		uuid, err := guid.GenerateRandom()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating random GUID: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(uuid)
	},
}

var guidFromCmd = &cobra.Command{
	Use:   "from <string1> <string2>",
	Short: "Generate deterministic UUID from two strings",
	Long:  `Generate a deterministic UUID v5 based on two input strings. Same input strings will always produce the same UUID`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		str1 := args[0]
		str2 := args[1]

		uuid, err := guid.GenerateFromStrings(str1, str2)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating deterministic GUID: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(uuid)
	},
}

var guidValidateCmd = &cobra.Command{
	Use:   "validate <uuid>",
	Short: "Validate UUID format",
	Long:  `Check if the provided string is a valid UUID format according to RFC 4122`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uuidStr := args[0]

		if guid.Validate(uuidStr) {
			fmt.Println("Valid UUID")
		} else {
			fmt.Println("Invalid UUID format")
			os.Exit(1)
		}
	},
}

func init() {
	// Add subcommands to guid command
	guidCmd.AddCommand(guidRandomCmd)
	guidCmd.AddCommand(guidFromCmd)
	guidCmd.AddCommand(guidValidateCmd)

	// Add guid command to root
	rootCmd.AddCommand(guidCmd)
}