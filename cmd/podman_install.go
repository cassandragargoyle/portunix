package cmd

import (
	"fmt"

	"portunix.cz/app/podman"

	"github.com/spf13/cobra"
)

// podmanInstallCmd represents the podman install command
var podmanInstallCmd = &cobra.Command{
	Use:   "install [-y]",
	Short: "Install Podman with intelligent OS detection and storage optimization.",
	Long: `Install Podman with intelligent operating system detection and optimal storage configuration.

This command automatically detects your operating system and installs the appropriate Podman variant:
- Windows: Podman Desktop for Windows
- Ubuntu/Debian: Podman via apt package manager
- CentOS/RHEL/Rocky/Fedora: Podman via yum/dnf package manager
- Alpine: Podman via apk package manager
- Generic Linux: Podman binaries installation

Storage optimization features:
- Analyzes available disk space on all drives/partitions
- Identifies drive/partition with most available space
- Recommends optimal location for Podman data storage
- Prevents system drive from filling up with container images/containers

Key Podman advantages:
- Rootless containers by default (enhanced security)
- Daemonless operation (no background service required)
- OCI-compatible with Docker images
- Pod support for Kubernetes-style container grouping
- Better default security posture

Examples:
  portunix podman install              # Interactive installation with storage selection
  portunix podman install -y          # Automated installation (accepts recommended storage)

Flags:
  -y, --yes    Auto-accept recommended storage location without prompting

Windows Installation:
- Downloads and installs Podman Desktop for Windows
- Configures WSL2 backend if available
- Sets up rootless container configuration
- Configures Windows subsystem for Linux integration

Linux Installation:
- Ubuntu/Debian: Installs podman package via apt
- CentOS/RHEL/Rocky/Fedora: Installs podman package via yum/dnf
- Alpine: Installs podman package via apk
- Generic Linux: Downloads Podman binaries directly
- Configures rootless containers and user namespaces
- Sets up container storage in optimal location
- Configures podman system for current user

Post-Installation:
- Verifies Podman installation with 'podman --version'
- Tests rootless container capability
- Displays storage and security configuration summary`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		autoAccept, _ := cmd.Flags().GetBool("yes")
		
		err := podman.InstallPodman(autoAccept)
		if err != nil {
			fmt.Printf("Error installing Podman: %v\n", err)
			return
		}
		
		fmt.Println("\n🎉 Podman installation completed successfully!")
		fmt.Println("💡 Podman runs rootless by default for enhanced security.")
	},
}

func init() {
	podmanCmd.AddCommand(podmanInstallCmd)
	
	// Add flags
	podmanInstallCmd.Flags().BoolP("yes", "y", false, "Auto-accept recommended storage location without prompting")
}