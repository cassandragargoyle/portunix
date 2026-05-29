/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"portunix.ai/app/podman"
)

// podmanDesktopCmd represents the podman desktop command
var podmanDesktopCmd = &cobra.Command{
	Use:   "desktop",
	Short: "Install Podman Desktop - the official GUI from Red Hat",
	Long: `Install Podman Desktop - the official GUI application from Red Hat for container management.

Podman Desktop is a free, open-source GUI that provides:
• Visual container, image, pod, and volume management
• Integration with Podman and Docker engines
• Remote Podman server management
• Kubernetes workload management
• Multi-platform support (Windows, macOS, Linux)

Podman Desktop offers a Docker Desktop-like experience with the security
and flexibility of Podman.

System Requirements:
- Windows 10/11 with WSL2 (automatically configured)
- macOS 10.15+ (Intel or Apple Silicon)
- Linux with X11/Wayland desktop environment

Installation Methods:
- Windows: Native installer (.exe)
- macOS: Homebrew cask or DMG installer
- Linux: System packages, AUR, or AppImage

Examples:
  portunix podman desktop                    # Interactive installation
  portunix podman desktop -y                # Auto-accept installation
  
Learn more: https://podman-desktop.io`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		autoAccept, _ := cmd.Flags().GetBool("yes")

		err := podman.InstallPodmanDesktop(autoAccept)
		if err != nil {
			fmt.Printf("Error installing Podman Desktop: %v\n", err)
			return
		}

		fmt.Println("\n🎉 Podman Desktop installation completed successfully!")
		fmt.Println("🖥️  You now have a powerful GUI for container management")
		fmt.Println()
		fmt.Println("🚀 Next steps:")
		fmt.Println("   1. Launch Podman Desktop from your applications")
		fmt.Println("   2. Follow the onboarding wizard")
		fmt.Println("   3. Connect to local or remote Podman")
		fmt.Println()
		fmt.Println("📚 Resources:")
		fmt.Println("   • Documentation: https://podman-desktop.io/docs")
		fmt.Println("   • GitHub: https://github.com/containers/podman-desktop")
		fmt.Println("   • Community: https://github.com/containers/podman-desktop/discussions")
	},
}

func init() {
	podmanCmd.AddCommand(podmanDesktopCmd)

	// Add flags
	podmanDesktopCmd.Flags().BoolP("yes", "y", false, "Auto-accept installation without prompting")
}
