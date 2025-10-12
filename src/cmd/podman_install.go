package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"portunix.ai/app/podman"

	"github.com/spf13/cobra"
)

// podmanInstallCmd represents the podman install command
var podmanInstallCmd = &cobra.Command{
	Use:   "install [-y] [--desktop-only] [--with-desktop]",
	Short: "Install Podman CLI and/or Desktop with intelligent OS detection.",
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

Installation Options:
  CLI Only:     Install just the command-line tools
  Desktop Only: Install just the GUI application (Podman Desktop)
  Both:         Install CLI + Desktop for complete experience

Examples:
  portunix podman install              # Interactive CLI installation with Desktop option
  portunix podman install -y          # Automated CLI installation (accepts defaults)
  portunix podman install --desktop-only  # Install only Podman Desktop GUI
  portunix podman install --with-desktop  # Install both CLI and Desktop
  portunix podman desktop             # Alternative way to install just Desktop

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
		desktopOnly, _ := cmd.Flags().GetBool("desktop-only")
		withDesktop, _ := cmd.Flags().GetBool("with-desktop")

		// If desktop-only flag is set, install only Podman Desktop
		if desktopOnly {
			err := podman.InstallPodmanDesktop(autoAccept)
			if err != nil {
				fmt.Printf("Error installing Podman Desktop: %v\n", err)
				return
			}
			fmt.Println("\n🎉 Podman Desktop installation completed successfully!")
			fmt.Println("🖥️  Launch Podman Desktop to get started with container management")
			return
		}

		// Install Podman CLI
		err := podman.InstallPodman(autoAccept)
		if err != nil {
			fmt.Printf("Error installing Podman: %v\n", err)
			return
		}

		fmt.Println("\n🎉 Podman CLI installation completed successfully!")
		fmt.Println("💡 Podman runs rootless by default for enhanced security.")

		// Ask about Desktop installation if not auto-accepting and not explicitly requested
		if !withDesktop && !autoAccept {
			fmt.Println()
			fmt.Println("🖥️  Would you also like to install Podman Desktop?")
			fmt.Println("   Podman Desktop is the official GUI from Red Hat that provides:")
			fmt.Println("   • Visual container and image management")
			fmt.Println("   • Integration with Docker and Podman")
			fmt.Println("   • Remote container management")
			fmt.Println("   • Kubernetes integration")
			fmt.Println("   Learn more: https://podman-desktop.io")
			fmt.Println()

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Install Podman Desktop? [y/N]: ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "y" || response == "yes" {
				withDesktop = true
			}
		}

		// Install Desktop if requested
		if withDesktop {
			fmt.Println("\n📦 Installing Podman Desktop...")
			err := podman.InstallPodmanDesktop(autoAccept)
			if err != nil {
				fmt.Printf("⚠️  Podman Desktop installation failed: %v\n", err)
				fmt.Println("💡 You can install it later with: portunix podman desktop")
				return
			}
			fmt.Println("\n🎉 Complete installation finished!")
			fmt.Println("✅ Podman CLI - ready for command-line container management")
			fmt.Println("✅ Podman Desktop - ready for GUI container management")
		}
	},
}

func init() {
	podmanCmd.AddCommand(podmanInstallCmd)

	// Add flags
	podmanInstallCmd.Flags().BoolP("yes", "y", false, "Auto-accept recommended storage location without prompting")
	podmanInstallCmd.Flags().Bool("desktop-only", false, "Install only Podman Desktop (GUI)")
	podmanInstallCmd.Flags().Bool("with-desktop", false, "Install both Podman CLI and Desktop")
}
