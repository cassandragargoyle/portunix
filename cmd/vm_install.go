package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"portunix.cz/app/install"
)

// vmInstallCmd represents the vm install-qemu command
var vmInstallCmd = &cobra.Command{
	Use:   "install-qemu",
	Short: "Install QEMU/KVM virtualization stack",
	Long: `Install QEMU/KVM and related tools for virtual machine management.
	
This will install:
- QEMU/KVM for virtualization
- libvirt for VM management
- virt-manager for GUI management (optional)
- Required networking tools`,
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS == "windows" {
			fmt.Println("Error: QEMU/KVM installation is only supported on Linux. For Windows, use WSL2")
			os.Exit(1)
		}

		fmt.Println("Installing QEMU/KVM virtualization stack...")

		// Check if running on Linux
		if runtime.GOOS != "linux" {
			fmt.Println("Error: QEMU/KVM is only supported on Linux systems")
			os.Exit(1)
		}

		// Check for virtualization support
		if err := checkVirtualizationSupport(); err != nil {
			fmt.Printf("Warning: %v\n", err)
			fmt.Println("You may still install QEMU, but performance will be limited without hardware virtualization.")
			fmt.Print("Continue? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				os.Exit(0)
			}
		}

		// Install QEMU using configuration-based installation system
		if err := install.InstallPackage("qemu", "default"); err != nil {
			fmt.Printf("Error: failed to install QEMU/KVM packages: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n✅ QEMU/KVM installation completed successfully!")
		fmt.Println("\n📋 Next steps:")
		fmt.Println("1. Activate group permissions (choose one option):")
		fmt.Println("   a) Log out and log back in (permanent solution)")
		fmt.Println("   b) Or run: newgrp libvirt (temporary for this session)")
		fmt.Println("2. Run 'portunix vm check' to verify installation")
		fmt.Println("3. Start virt-manager GUI or create VMs with 'portunix vm create'")
		fmt.Println("\n💡 If virt-manager shows connection errors, use option 1a or 1b above")
	},
}

// vmCheckCmd represents the vm check command
var vmCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check QEMU/KVM installation and virtualization support",
	Long:  `Check if QEMU/KVM is properly installed and hardware virtualization is supported.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking QEMU/KVM installation...")

		// Check virtualization support
		if err := checkVirtualizationSupport(); err != nil {
			fmt.Printf("❌ Hardware virtualization: %v\n", err)
		} else {
			fmt.Println("✅ Hardware virtualization: Supported")
		}

		// Check QEMU installation
		if _, err := exec.LookPath("qemu-system-x86_64"); err != nil {
			fmt.Println("❌ QEMU: Not installed")
		} else {
			fmt.Println("✅ QEMU: Installed")
		}

		// Check libvirt
		if _, err := exec.LookPath("virsh"); err != nil {
			fmt.Println("❌ libvirt: Not installed")
		} else {
			fmt.Println("✅ libvirt: Installed")
			// Check if libvirt service is running
			cmd := exec.Command("systemctl", "is-active", "libvirtd")
			if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) == "active" {
				fmt.Println("✅ libvirtd service: Running")
			} else {
				fmt.Println("❌ libvirtd service: Not running")
			}
		}

		// Check user groups
		groupsCmd := exec.Command("groups")
		if output, err := groupsCmd.Output(); err == nil {
			groups := string(output)
			hasLibvirt := strings.Contains(groups, "libvirt")
			hasKvm := strings.Contains(groups, "kvm")

			if hasLibvirt {
				fmt.Println("✅ User in libvirt group: Yes")
			} else {
				fmt.Println("❌ User in libvirt group: No")
			}
			if hasKvm {
				fmt.Println("✅ User in kvm group: Yes")
			} else {
				fmt.Println("❌ User in kvm group: No")
			}

			// Show help if user is not in groups
			if !hasLibvirt || !hasKvm {
				fmt.Println("\n💡 To fix group membership:")
				fmt.Println("   sudo usermod -aG libvirt,kvm $USER")
				fmt.Println("   Then either:")
				fmt.Println("   - Log out and back in (permanent)")
				fmt.Println("   - Or run: newgrp libvirt (temporary)")
			}
		}
	},
}

func checkVirtualizationSupport() error {
	// Check for Intel VT-x or AMD-V
	cpuinfo, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return fmt.Errorf("cannot read CPU information")
	}

	cpuinfoStr := string(cpuinfo)
	if !strings.Contains(cpuinfoStr, "vmx") && !strings.Contains(cpuinfoStr, "svm") {
		return fmt.Errorf("CPU does not support hardware virtualization (Intel VT-x or AMD-V)")
	}

	// Check if KVM module is loaded
	if _, err := os.Stat("/dev/kvm"); os.IsNotExist(err) {
		return fmt.Errorf("KVM module is not loaded. Enable virtualization in BIOS/UEFI")
	}

	return nil
}

func init() {
	vmCmd.AddCommand(vmInstallCmd)
	vmCmd.AddCommand(vmCheckCmd)
}
