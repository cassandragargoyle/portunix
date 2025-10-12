package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"portunix.ai/app/system"
	"portunix.ai/portunix/src/dispatcher"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "System information and OS detection commands",
	Long: `Provides system information and OS detection capabilities.
Useful for scripts that need to adapt behavior based on the operating system,
version, and environment (like Windows Sandbox, Docker, WSL, etc.)`,
}

var systemInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display detailed system information",
	Long: `Displays comprehensive system information including:
- Operating system (Windows, Linux, macOS)
- OS version and build
- Environment variant (Sandbox, Docker, WSL, VM)
- Architecture
- Hostname
- PowerShell/Shell availability

Output can be formatted as JSON for programmatic use.`,
	Run: func(cmd *cobra.Command, args []string) {
		formatJSON, _ := cmd.Flags().GetBool("json")
		formatShort, _ := cmd.Flags().GetBool("short")

		sysInfo, err := system.GetSystemInfo()
		if err != nil {
			fmt.Printf("Error getting system information: %v\n", err)
			os.Exit(1)
		}

		if formatJSON {
			jsonData, err := json.MarshalIndent(sysInfo, "", "  ")
			if err != nil {
				fmt.Printf("Error formatting JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonData))
		} else if formatShort {
			fmt.Printf("%s %s %s\n", sysInfo.OS, sysInfo.Version, sysInfo.Variant)
		} else {
			printSystemInfo(sysInfo)
		}
	},
}

var systemCheckCmd = &cobra.Command{
	Use:   "check [condition]",
	Short: "Check specific system conditions",
	Long: `Check specific system conditions and return appropriate exit codes.
Useful for conditional execution in scripts.

Available conditions:
  windows         - Check if running on Windows
  linux          - Check if running on Linux  
  macos          - Check if running on macOS
  sandbox        - Check if running in Windows Sandbox
  docker         - Check if running in Docker
  wsl            - Check if running in WSL
  vm             - Check if running in a VM
  powershell     - Check if PowerShell is available
  admin          - Check if running with administrator privileges

Examples:
  portunix system check windows     # Exit 0 if Windows, 1 otherwise
  portunix system check sandbox     # Exit 0 if Windows Sandbox, 1 otherwise`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		condition := strings.ToLower(args[0])

		sysInfo, err := system.GetSystemInfo()
		if err != nil {
			fmt.Printf("Error getting system information: %v\n", err)
			os.Exit(1)
		}

		result := system.CheckCondition(sysInfo, condition)
		if !result {
			os.Exit(1)
		}
	},
}

var systemDispatcherCmd = &cobra.Command{
	Use:   "dispatcher",
	Short: "Display dispatcher and helper binary information",
	Long: `Displays information about the Git-like dispatcher architecture including:
- Dispatcher version and configuration
- Available helper binaries (ptx-*)
- Helper binary versions and compatibility
- Commands handled by each helper

This command is useful for debugging dispatcher issues and verifying
the Git-like architecture implementation.`,
	Run: func(cmd *cobra.Command, args []string) {
		formatJSON, _ := cmd.Flags().GetBool("json")

		// Get version from global variable or environment
		version := "dev" // This should match the version used in main.go

		disp := dispatcher.NewDispatcher(version)

		if formatJSON {
			// Discover all helpers and format as JSON
			helpers, err := disp.DiscoverHelpers()
			if err != nil {
				fmt.Printf("Error discovering helpers: %v\n", err)
				os.Exit(1)
			}

			output := map[string]interface{}{
				"dispatcher_version": version,
				"executable_dir":     disp.GetExecutableDir(),
				"helpers":           helpers,
			}

			jsonData, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				fmt.Printf("Error formatting JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonData))
		} else {
			printDispatcherInfo(disp, version)
		}
	},
}

func printDispatcherInfo(disp *dispatcher.Dispatcher, version string) {
	fmt.Printf("Dispatcher Information:\n")
	fmt.Printf("======================\n")
	fmt.Printf("Dispatcher Version: %s\n", version)
	fmt.Printf("Executable Dir:     %s\n", disp.GetExecutableDir())
	fmt.Printf("\n")

	// Discover helper binaries
	helpers, err := disp.DiscoverHelpers()
	if err != nil {
		fmt.Printf("Error discovering helpers: %v\n", err)
		return
	}

	if len(helpers) == 0 {
		fmt.Printf("Helper Binaries: None found\n")
		fmt.Printf("\nNote: In Phase 1, no helper binaries are expected.\n")
		fmt.Printf("Helper binaries will be implemented in Phase 2:\n")
		fmt.Printf("  - ptx-container (for docker/podman/container commands)\n")
		fmt.Printf("  - ptx-mcp (for mcp commands)\n")
	} else {
		fmt.Printf("Helper Binaries:\n")
		fmt.Printf("================\n")
		for _, helper := range helpers {
			fmt.Printf("\nHelper: %s\n", helper.Name)
			fmt.Printf("  Binary:      %s\n", helper.Binary)
			fmt.Printf("  Available:   %t\n", helper.Available)
			if helper.Available {
				fmt.Printf("  Version:     %s\n", helper.Version)
				if len(helper.Commands) > 0 {
					fmt.Printf("  Commands:    %s\n", strings.Join(helper.Commands, ", "))
				}
				fmt.Printf("  Description: %s\n", helper.Description)
			}
		}
	}

	// Show legacy command handling only if helpers are not available
	hasContainerHelper := false
	hasMCPHelper := false

	for _, helper := range helpers {
		if helper.Available {
			switch helper.Name {
			case "ptx-container":
				hasContainerHelper = true
			case "ptx-mcp":
				hasMCPHelper = true
			}
		}
	}

	if !hasContainerHelper || !hasMCPHelper {
		fmt.Printf("\nLegacy Commands:\n")
		fmt.Printf("===============\n")
		fmt.Printf("The following commands are handled by the main binary\n")
		fmt.Printf("(helper binaries not available):\n")

		if !hasContainerHelper {
			fmt.Printf("  - container, docker, podman (fallback to main binary)\n")
		}
		if !hasMCPHelper {
			fmt.Printf("  - mcp (fallback to main binary)\n")
		}
	} else {
		fmt.Printf("\nPhase 2 Status:\n")
		fmt.Printf("==============\n")
		fmt.Printf("✅ Helper binary architecture active\n")
		fmt.Printf("✅ All commands properly delegated to helper binaries\n")
		fmt.Printf("✅ Fallback mechanism available if helpers are unavailable\n")
	}
}

func formatInstalled(installed bool) string {
	if installed {
		return "installed"
	}
	return "not installed"
}

func formatInstalledWithVersion(installed bool, backend string) string {
	if !installed {
		return "not installed"
	}

	var version string
	switch backend {
	case "docker":
		version = system.GetDockerVersion()
	case "podman":
		version = system.GetPodmanVersion()
	case "qemu":
		version = system.GetQEMUVersion()
	case "virtualbox":
		version = system.GetVirtualBoxVersion()
	case "libvirt":
		version = system.GetLibvirtVersion()
	}

	if version != "" {
		return version
	}
	return "installed"
}

func printSystemInfo(info *system.SystemInfo) {
	fmt.Printf("System Information:\n")
	fmt.Printf("==================\n")
	fmt.Printf("OS:           %s\n", info.OS)
	fmt.Printf("Version:      %s\n", info.Version)
	fmt.Printf("Build:        %s\n", info.Build)
	fmt.Printf("Architecture: %s\n", info.Architecture)
	fmt.Printf("Hostname:     %s\n", info.Hostname)
	fmt.Printf("Variant:      %s\n", info.Variant)
	if len(info.Environment) > 0 {
		fmt.Printf("Environment:  %s\n", strings.Join(info.Environment, ", "))
	}

	if info.WindowsInfo != nil {
		fmt.Printf("\nWindows Details:\n")
		fmt.Printf("Edition:      %s\n", info.WindowsInfo.Edition)
		fmt.Printf("Product:      %s\n", info.WindowsInfo.ProductName)
		fmt.Printf("Install Date: %s\n", info.WindowsInfo.InstallDate)
	}

	if info.LinuxInfo != nil {
		fmt.Printf("\nLinux Details:\n")
		fmt.Printf("Distribution: %s\n", info.LinuxInfo.Distribution)
		fmt.Printf("Codename:     %s\n", info.LinuxInfo.Codename)
		fmt.Printf("Kernel:       %s\n", info.LinuxInfo.KernelVersion)
	}

	fmt.Printf("\nCapabilities:\n")
	fmt.Printf("PowerShell:   %t\n", info.Capabilities.PowerShell)
	fmt.Printf("Admin:        %t\n", info.Capabilities.Admin)
	fmt.Printf("Container Available: %t\n", info.Capabilities.ContainerAvailable)

	// Add hardware virtualization capability
	if info.Capabilities.VirtualizationInfo != nil {
		fmt.Printf("Virtualization: %t\n", info.Capabilities.VirtualizationInfo.HardwareVirtualization)
	}

	fmt.Printf("\nContainer Runtimes:\n")
	fmt.Printf("Docker:       %s\n", formatInstalledWithVersion(info.Capabilities.Docker, "docker"))
	fmt.Printf("Podman:       %s\n", formatInstalledWithVersion(info.Capabilities.Podman, "podman"))

	// Virtualization backends
	if info.Capabilities.VirtualizationInfo != nil {
		fmt.Printf("\nVirtualization Backends:\n")
		virtInfo := info.Capabilities.VirtualizationInfo
		fmt.Printf("QEMU/KVM:     %s\n", formatInstalledWithVersion(virtInfo.QEMU, "qemu"))
		fmt.Printf("VirtualBox:   %s\n", formatInstalledWithVersion(virtInfo.VirtualBox, "virtualbox"))
		if virtInfo.LibvirtInstalled && info.OS == "Linux" {
			fmt.Printf("Libvirt:      %s\n", formatInstalledWithVersion(virtInfo.LibvirtInstalled, "libvirt"))
		}
	}

	// Certificate information
	fmt.Printf("\nCertificates:\n")
	if info.Capabilities.CertificateInfo != nil {
		certInfo := info.Capabilities.CertificateInfo
		fmt.Printf("Available:    %t", certInfo.Available)
		if certInfo.Available {
			fmt.Printf("\nHTTPS:        %t", certInfo.HTTPSWorking)
			if certInfo.Path != "" {
				fmt.Printf("\nPath:         %s", certInfo.Path)
			}
		}
		fmt.Printf("\n")
	} else {
		fmt.Printf("Available:    unknown\n")
	}
}

func init() {
	rootCmd.AddCommand(systemCmd)
	systemCmd.AddCommand(systemInfoCmd)
	systemCmd.AddCommand(systemCheckCmd)
	systemCmd.AddCommand(systemDispatcherCmd)

	// Add flags for system info command
	systemInfoCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	systemInfoCmd.Flags().BoolP("short", "s", false, "Short output (OS Version Variant)")

	// Add flags for dispatcher command
	systemDispatcherCmd.Flags().BoolP("json", "j", false, "Output as JSON")
}
