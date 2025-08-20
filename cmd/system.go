package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"portunix.cz/app/system"
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

func printSystemInfo(info *system.SystemInfo) {
	fmt.Printf("System Information:\n")
	fmt.Printf("==================\n")
	fmt.Printf("OS:           %s\n", info.OS)
	fmt.Printf("Version:      %s\n", info.Version)
	fmt.Printf("Build:        %s\n", info.Build)
	fmt.Printf("Architecture: %s\n", info.Architecture)
	fmt.Printf("Hostname:     %s\n", info.Hostname)
	fmt.Printf("Variant:      %s\n", info.Variant)
	fmt.Printf("Environment:  %s\n", strings.Join(info.Environment, ", "))
	
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
	fmt.Printf("Docker:       %t\n", info.Capabilities.Docker)
	fmt.Printf("Admin:        %t\n", info.Capabilities.Admin)
}

func init() {
	rootCmd.AddCommand(systemCmd)
	systemCmd.AddCommand(systemInfoCmd)
	systemCmd.AddCommand(systemCheckCmd)
	
	// Add flags for system info command
	systemInfoCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	systemInfoCmd.Flags().BoolP("short", "s", false, "Short output (OS Version Variant)")
}