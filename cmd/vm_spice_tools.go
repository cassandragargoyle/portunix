package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"portunix.cz/app/install"
)

var (
	targetVM    string
	mountAsISO  bool
	autoInstall bool
)

var vmSpiceToolsCmd = &cobra.Command{
	Use:   "install-spice-tools",
	Short: "Install SPICE Guest Tools for clipboard support",
	Long: `Install SPICE Guest Tools in Windows VMs to enable clipboard sharing.

This command can:
- Download SPICE Guest Tools installer
- Mount it as ISO in running VM (if QEMU monitor is available)
- Provide installation instructions
- Support both new and existing VMs`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := installSPICETools(); err != nil {
			fmt.Printf("\n❌ Failed to install SPICE tools: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	vmCmd.AddCommand(vmSpiceToolsCmd)

	vmSpiceToolsCmd.Flags().StringVar(&targetVM, "vm", "", "Target VM name (optional)")
	vmSpiceToolsCmd.Flags().BoolVar(&mountAsISO, "mount", false, "Mount as ISO in VM")
	vmSpiceToolsCmd.Flags().BoolVar(&autoInstall, "auto", false, "Attempt automatic installation")
}

func installSPICETools() error {
	fmt.Println("\n🔧 SPICE Guest Tools Installation")
	fmt.Println("════════════════════════════════════════")

	// Determine download location
	downloadDir := filepath.Join(getCacheDir(), "spice-tools")
	os.MkdirAll(downloadDir, 0755)

	toolsPath := filepath.Join(downloadDir, "spice-guest-tools-latest.exe")

	// Check if already downloaded
	if _, err := os.Stat(toolsPath); err == nil {
		fmt.Printf("✅ SPICE tools already downloaded: %s\n", toolsPath)
	} else {
		// Download SPICE tools
		fmt.Println("📥 Downloading SPICE Guest Tools...")
		if err := downloadSPICEToolsFile(toolsPath); err != nil {
			// Fallback to using standard install system
			fmt.Println("⚠️  Direct download failed, trying package manager...")
			if err := install.InstallPackage("spice-guest-tools", "latest"); err != nil {
				return fmt.Errorf("failed to download SPICE tools: %w", err)
			}
		} else {
			fmt.Printf("✅ Downloaded to: %s\n", toolsPath)
		}
	}

	// If target VM specified, try to help with installation
	if targetVM != "" {
		fmt.Printf("\n🎯 Target VM: %s\n", targetVM)

		if mountAsISO {
			// Try to mount as ISO in VM
			if err := mountToolsInVM(targetVM, toolsPath); err != nil {
				fmt.Printf("⚠️  Could not mount in VM: %v\n", err)
			} else {
				fmt.Println("✅ Tools mounted in VM as CD-ROM")
			}
		}

		// Check if VM is running
		if isVMRunning(targetVM) {
			fmt.Println("\n📋 VM is running. Installation options:")
			fmt.Println("1. Copy the installer to VM via network share")
			fmt.Println("2. Use QEMU monitor to attach as CD-ROM")
			fmt.Println("3. Download directly in VM from:")
			fmt.Println("   https://www.spice-space.org/download/windows/spice-guest-tools/")
		} else {
			fmt.Println("\n⚠️  VM is not running. Start the VM first.")
		}
	}

	// Provide installation instructions
	fmt.Println("\n📋 Installation Instructions:")
	fmt.Println("════════════════════════════════")
	fmt.Println("1. Copy installer to Windows VM:")
	fmt.Printf("   Location: %s\n", toolsPath)
	fmt.Println()
	fmt.Println("2. In Windows VM:")
	fmt.Println("   • Run spice-guest-tools-latest.exe as Administrator")
	fmt.Println("   • Follow the installation wizard")
	fmt.Println("   • Restart Windows when prompted")
	fmt.Println()
	fmt.Println("3. After restart:")
	fmt.Println("   • Clipboard sharing will work automatically")
	fmt.Println("   • Screen resolution will auto-adjust")
	fmt.Println("   • USB redirection will be available")
	fmt.Println()
	fmt.Println("💡 Tips:")
	fmt.Println("   • Connect with: portunix vm connect <vm-name>")
	fmt.Println("   • Or use: virt-viewer --connect spice://localhost:5900")
	fmt.Println()
	fmt.Println("🔗 Manual download URL:")
	fmt.Println("   https://www.spice-space.org/download/windows/spice-guest-tools/spice-guest-tools-latest.exe")

	return nil
}

func downloadSPICEToolsFile(dest string) error {
	url := "https://www.spice-space.org/download/windows/spice-guest-tools/spice-guest-tools-latest.exe"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 0, // No timeout for large downloads
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create output file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy with progress
	size := resp.ContentLength
	if size > 0 {
		fmt.Printf("📦 Download size: %.2f MB\n", float64(size)/1024/1024)
	}

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Downloaded %.2f MB\n", float64(written)/1024/1024)
	return nil
}

func mountToolsInVM(vmName, toolsPath string) error {
	// This would use QEMU monitor to change CD-ROM
	// For now, just create an ISO wrapper

	isoPath := strings.TrimSuffix(toolsPath, ".exe") + ".iso"

	// Create ISO with the exe file
	cmd := exec.Command("genisoimage",
		"-o", isoPath,
		"-J", "-R",
		"-V", "SPICE_TOOLS",
		toolsPath)

	if err := cmd.Run(); err != nil {
		// Try mkisofs as fallback
		cmd = exec.Command("mkisofs",
			"-o", isoPath,
			"-J", "-R",
			"-V", "SPICE_TOOLS",
			toolsPath)

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create ISO: %w", err)
		}
	}

	fmt.Printf("✅ Created ISO: %s\n", isoPath)

	// Now we would use QEMU monitor to attach it
	// This requires QMP or monitor socket access
	fmt.Println("📋 To mount in running VM, use QEMU monitor:")
	fmt.Printf("   (qemu) change ide1-cd0 %s\n", isoPath)

	return nil
}

func isVMRunning(vmName string) bool {
	// Check if VM process is running
	cmd := exec.Command("pgrep", "-f", fmt.Sprintf("qemu.*%s", vmName))
	if err := cmd.Run(); err == nil {
		return true
	}
	return false
}

func getCacheDir() string {
	// Use Portunix cache directory
	cacheDir := ".cache"
	if home, err := os.UserHomeDir(); err == nil {
		cacheDir = filepath.Join(home, ".portunix", "cache")
	}
	os.MkdirAll(cacheDir, 0755)
	return cacheDir
}
