package cmd

import (
	"fmt"

	"portunix.cz/app/install"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [software1] [software2] ...",
	Short: "Installs specified software.",
	Long: `The install command allows you to install various software components.

You can specify one or more software packages to install.

Available software packages:
  java          - Java Development Kit (OpenJDK)
  python        - Python programming language
  vscode        - Visual Studio Code editor
  go            - Go programming language
  chrome        - Google Chrome browser
  mvn           - Apache Maven build tool
  chocolatey    - Chocolatey package manager for Windows
  winget        - Windows Package Manager
  docker        - Docker Engine/Desktop with intelligent OS detection
  podman        - Podman container engine with rootless support

Package variants (use with --variant):
  java: 8, 11, 17, 21 (default: 21)
  python: embeddable, full (default: embeddable)
  vscode: user, system (default: user)
  mvn: 3.9.9, latest, apt (Linux only) (default: latest)

Python installation options:
  --gui         Use GUI installer instead of silent installation
  --embeddable  Install embeddable Python (portable ZIP version)

Examples:
  portunix install --help
  portunix install java
  portunix install java --variant 17
  portunix install python --embeddable
  portunix install vscode chrome
  portunix install mvn
  portunix install mvn --variant 3.9.9
  portunix install chocolatey
  portunix install winget
  portunix install docker
  portunix install docker -y
  portunix install podman
  portunix install podman -y`,
	DisableFlagParsing: true, // Allow passing flags to specific installers
	Run: func(cmd *cobra.Command, args []string) {
		// Check for help flags first
		for _, arg := range args {
			if arg == "--help" || arg == "-h" {
				cmd.Help()
				return
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
	rootCmd.AddCommand(installCmd)
}
