package cmd

import (
	"fmt"

	"portunix.cz/app/docker"

	"github.com/spf13/cobra"
)

// dockerInstallCmd represents the docker install command
var dockerInstallCmd = &cobra.Command{
	Use:   "install [-y]",
	Short: "Install Docker with intelligent OS detection and storage optimization.",
	Long: `Install Docker with intelligent operating system detection and optimal storage configuration.

This command automatically detects your operating system and installs the appropriate Docker variant:
- Windows: Docker Desktop for Windows with WSL2 backend
- Ubuntu/Debian: Docker Engine via apt package manager
- CentOS/RHEL/Rocky/Fedora: Docker Engine via yum/dnf package manager
- Alpine: Docker Engine via apk package manager
- Generic Linux: Docker binaries installation

Storage optimization features:
- Analyzes available disk space on all drives/partitions
- Identifies drive/partition with most available space
- Recommends optimal location for Docker data storage
- Prevents system drive from filling up with Docker images/containers

Examples:
  portunix docker install              # Interactive installation with storage selection
  portunix docker install -y          # Automated installation (accepts recommended storage)

Flags:
  -y, --yes    Auto-accept recommended storage location without prompting

Windows Installation:
- Downloads and installs Docker Desktop for Windows
- Configures custom data-root location if user selects non-default drive
- Handles Windows version requirements (Windows 10/11 Pro/Enterprise)
- Enables WSL2 backend if available
- Configures Windows features (Hyper-V or WSL2)

Linux Installation:
- Ubuntu/Debian: Installs docker.io or docker-ce via apt
- CentOS/RHEL/Rocky: Installs docker-ce via yum/dnf
- Alpine: Installs docker via apk
- Generic Linux: Downloads Docker binaries directly
- Configures custom docker data directory if user selects non-default location
- Configures docker daemon and adds user to docker group
- Enables and starts docker service

Post-Installation:
- Verifies Docker installation with 'docker --version'
- Tests Docker daemon accessibility
- Displays storage configuration summary`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		autoAccept, _ := cmd.Flags().GetBool("yes")

		err := docker.InstallDocker(autoAccept)
		if err != nil {
			fmt.Printf("Error installing Docker: %v\n", err)
			return
		}

		fmt.Println("\n🎉 Docker installation completed successfully!")
	},
}

func init() {
	dockerCmd.AddCommand(dockerInstallCmd)

	// Add flags
	dockerInstallCmd.Flags().BoolP("yes", "y", false, "Auto-accept recommended storage location without prompting")
}
