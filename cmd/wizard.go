package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"portunix.cz/app"
	"portunix.cz/app/install"
)

// wizardCmd represents the wizard command
var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "Starts an interactive wizard.",
	Long: `Starts an interactive wizard to guide you through the installation and
configuration process. This is a user-friendly way to set up your environment
without having to remember all the commands and flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		startWizard()
	},
}

func init() {
	rootCmd.AddCommand(wizardCmd)
}

func startWizard() {
	app.PrintLogo()
	app.PrintLine("=")
	// Define a list of choices.
	choices := []string{"Install", "Config", "Check"}
	// Create a survey question for selecting a choice.
	var selectedChoice string
	prompt := &survey.Select{
		Message: "Select an option:",
		Options: choices,
		Default: choices[0], // The default choice (optional).
	}
	// Ask the user to select a choice.
	err := survey.AskOne(prompt, &selectedChoice)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	switch selectedChoice {
	case "Install":
		promptInstall()
	case "Config":
		promptConfig()

	}
	// Display a message.
	fmt.Println("Press Enter to continue...")
	// Wait for user input.
	var input string
	fmt.Scanf("%s", &input)
}

func promptInstall() {
	choices := []string{"sh-daemon", "Java Runtime Enviroment"}
	// Create a survey question for selecting a choice.
	var selectedChoice string
	prompt := &survey.Select{
		Message: "Select what to install:",
		Options: choices,
		Default: choices[0], // The default choice (optional).
	}
	err := survey.AskOne(prompt, &selectedChoice)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	switch selectedChoice {
	case "sh-daemon":
		install.Install(install.ToArguments("daemon"))
	case "Java Runtime Enviroment":
		install.Install(install.ToArguments("jre"))
	}
}

func promptConfig() {
	// TODO: Implement the config prompt
	fmt.Println("Configuration prompt is not implemented yet.")
}
