package cmd

import (
	"fmt"
	"strings"

	"portunix.ai/app/install"

	"github.com/spf13/cobra"
)

// generateInstallHelp generates dynamic help content based on current configuration
func generateInstallHelp() string {
	config, err := install.LoadInstallConfig()
	if err != nil {
		// Fallback to basic help if config loading fails
		return `The install command allows you to install various software components.

You can specify one or more software packages to install.

Error loading package configuration: ` + err.Error() + `

Please check your installation.`
	}

	var helpText strings.Builder
	helpText.WriteString(`The install command allows you to install various software components.

You can specify one or more software packages to install.

`)
	
	// Add package list
	helpText.WriteString(config.GeneratePackageListDescription())
	
	// Add preset list
	helpText.WriteString(config.GeneratePresetListDescription())
	
	// Add variant list
	helpText.WriteString(config.GenerateVariantListDescription())
	
	return helpText.String()
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [software1] [software2] ...",
	Short: "Installs specified software.",
	Long:  "", // Will be set dynamically in init()
	DisableFlagParsing: true, // Allow passing flags to specific installers
	Run: func(cmd *cobra.Command, args []string) {
		// Check for package-specific help: "install packagename --help"
		if len(args) >= 2 {
			packageName := args[0]
			for _, arg := range args[1:] {
				if arg == "--help" || arg == "-h" {
					if err := install.ShowPackageHelp(packageName); err != nil {
						fmt.Printf("Help not available for package '%s': %v\n", packageName, err)
						fmt.Println("Use 'portunix install --help' for general installation help.")
					}
					return
				}
			}
		}

		// Check for general help flags
		for _, arg := range args {
			if arg == "--help" || arg == "-h" {
				cmd.Help()
				return
			}
		}

		// Check for AI assistant recommendation flag
		for _, arg := range args {
			if arg == "--recommend-ai" {
				install.RecommendAIAssistants()
				return
			}
		}

		// Check for invalid flags (starting with -- but not recognized)
		validFlags := []string{"--help", "--gui", "--embeddable", "--dry-run", "--recommend-ai", "--method", "--version", "--list-methods"}
		for _, arg := range args {
			if strings.HasPrefix(arg, "--") {
				isValid := false
				for _, validFlag := range validFlags {
					if arg == validFlag || strings.HasPrefix(arg, "--variant") || strings.HasPrefix(arg, "--method=") || strings.HasPrefix(arg, "--version=") {
						isValid = true
						break
					}
				}
				if !isValid {
					fmt.Printf("Error: unknown flag '%s'\n", arg)
					fmt.Println("Use 'portunix install --help' to see available options.")
					return
				}
			}
		}

		if len(args) == 0 {
			fmt.Println("Please specify the software to install.")
			return
		}
		install.Install(args)
	},
}

func init() {
	// Set dynamic help content
	installCmd.Long = generateInstallHelp() + `

Installation options:
  --method <method>     Override preferred installation method
  --version <version>   Specify version (latest, prerelease, or specific version)
  --list-methods        Show all available installation methods for package
  --dry-run             Preview installation without executing
  --variant <variant>   Specify package variant (legacy compatibility)

Python installation options:
  --gui         Use GUI installer instead of silent installation
  --embeddable  Install embeddable Python (portable ZIP version)

Examples:
  portunix install --help
  portunix install java
  portunix install java --variant 17
  portunix install nodejs
  portunix install nodejs --variant 20
  portunix install python --embeddable
  portunix install vscode chrome
  portunix install mvn
  portunix install mvn --variant 3.9.9
  portunix install chocolatey
  portunix install winget
  portunix install chrome
  portunix install chrome --variant fedora
  portunix install chrome --variant snap
  portunix install powershell
  portunix install powershell --variant ubuntu
  portunix install docker
  portunix install docker -y
  portunix install podman
  portunix install podman -y
  portunix install act
  portunix install gh
  portunix install actionlint
  portunix install github-actions
  portunix install claude-code
  portunix install claude-desktop
  portunix install gemini-cli
  portunix install ai-assistant-full
  portunix install hugo
  portunix install hugo --variant standard
  portunix install hugo --variant extended
  portunix install hugo-extended
  portunix install hugo --list-methods
  portunix install hugo --method=deb
  portunix install hugo --version=latest
  portunix install hugo --dry-run
  portunix install --recommend-ai`

	rootCmd.AddCommand(installCmd)
}
