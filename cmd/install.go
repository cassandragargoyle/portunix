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
  claude-code   - Anthropic's official CLI for Claude AI assistant
  powershell    - Cross-platform PowerShell scripting environment
  docker        - Docker Engine/Desktop with intelligent OS detection
  podman        - Podman container engine with rootless support
  spice-guest-tools - SPICE guest tools for Windows VMs (clipboard sharing)
  act           - Run GitHub Actions locally using Docker containers
  actionlint    - Static checker for GitHub Actions workflow files  
  gh            - GitHub's official command line tool

Package variants (use with --variant):
  java: 8, 11, 17, 21 (default: 21)
  python: embeddable, full (default: embeddable)
  vscode: user, system (default: user)
  mvn: 3.9.9, latest, apt (Linux only) (default: latest)
  claude-code: npm, curl (default: npm)
  powershell: latest (Windows), ubuntu, debian, fedora, rocky, mint, elementary, snap (Linux) (default: auto-detect)
  chrome: stable, beta, dev (Windows), ubuntu, debian, fedora, rocky, mint, deb-direct, rpm-direct, snap (Linux) (default: auto-detect)

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
  portunix install github-actions`,
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
