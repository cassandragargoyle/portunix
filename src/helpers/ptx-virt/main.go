package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var version = "dev"

//go:embed virtualbox-help.md
var virtualBoxHelpContent string

// rootCmd represents the base command for ptx-virt
var rootCmd = &cobra.Command{
	Use:   "ptx-virt",
	Short: "Portunix Virtualization Management Helper",
	Long: `ptx-virt is a helper binary for Portunix that handles all virtualization operations.
It provides unified interface for VirtualBox, QEMU/KVM, VMware, and Hyper-V management.

This binary is typically invoked by the main portunix dispatcher and should not be used directly.

Supported virtualization backends:
- VirtualBox (cross-platform)
- QEMU/KVM (Linux)
- VMware (cross-platform)
- Hyper-V (Windows)`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle the dispatched command directly
		handleCommand(args)
	},
}

func handleCommand(args []string) {
	// Handle dispatched commands: virt
	if len(args) == 0 {
		fmt.Println("No command specified")
		return
	}

	command := args[0]
	subArgs := args[1:]

	switch command {
	case "virt":
		if len(subArgs) == 0 {
			// Show virt help
			showVirtHelp()
		} else {
			handleVirtCommand(subArgs)
		}
	case "--version":
		fmt.Printf("ptx-virt version %s\n", version)
	case "--description":
		fmt.Println("Portunix Virtualization Management Helper")
	case "--list-commands":
		fmt.Println("virt")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Supported commands: virt")
	}
}

func showVirtHelp() {
	fmt.Println("Usage: portunix virt [subcommand]")
	fmt.Println()
	fmt.Println("Virtualization Management Commands:")
	fmt.Println("  check       - Check virtualization support and available backends")
	fmt.Println("  list        - List all virtual machines")
	fmt.Println("  create      - Create a new virtual machine")
	fmt.Println("  start       - Start a virtual machine")
	fmt.Println("  stop        - Stop a virtual machine")
	fmt.Println("  restart     - Restart a virtual machine")
	fmt.Println("  delete      - Delete a virtual machine")
	fmt.Println("  info        - Show detailed VM information")
	fmt.Println("  status      - Show VM status")
	fmt.Println("  ssh         - SSH into a virtual machine")
	fmt.Println("  snapshot    - Manage VM snapshots")
	fmt.Println("  install-qemu - Install QEMU/KVM")
	fmt.Println()
	fmt.Println("Backend Detection:")
	fmt.Println("  The helper automatically detects available virtualization backends")
	fmt.Println("  Priority: VirtualBox > QEMU/KVM > VMware > Hyper-V")
}

func handleVirtCommand(args []string) {
	if len(args) == 0 {
		showVirtHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "check":
		handleCheckCommand()
	case "list", "ls":
		handleListCommand(subArgs)
	case "create":
		handleCreateCommand(subArgs)
	case "start":
		handleStartCommand(subArgs)
	case "stop":
		handleStopCommand(subArgs)
	case "restart":
		handleRestartCommand(subArgs)
	case "delete", "rm":
		handleDeleteCommand(subArgs)
	case "info":
		handleInfoCommand(subArgs)
	case "status":
		handleStatusCommand(subArgs)
	case "ssh":
		handleSSHCommand(subArgs)
	case "snapshot":
		handleSnapshotCommand(subArgs)
	case "install-qemu":
		handleInstallQEMUCommand()
	case "--help", "-h":
		showVirtHelp()
	default:
		fmt.Printf("Unknown virt subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix virt --help' for available commands")
	}
}

// Command handlers - these use the VirtManager
func handleCheckCommand() {
	fmt.Println("Checking virtualization support...")

	manager, err := NewVirtManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	capabilities, err := manager.CheckCapabilities()
	if err != nil {
		fmt.Printf("Error checking capabilities: %v\n", err)
		return
	}

	// Display capabilities
	fmt.Printf("Platform: %s\n", capabilities.Platform)
	fmt.Printf("Hardware Virtualization: %v\n", capabilities.HardwareVirtualization)
	fmt.Printf("Recommended Provider: %s\n", capabilities.RecommendedProvider)
	fmt.Println("\nAvailable Providers:")
	for _, provider := range capabilities.AvailableProviders {
		status := "❌"
		if provider.Available {
			status = "✅"
		}
		fmt.Printf("  %s %s", status, provider.Name)
		if provider.Version != "" {
			fmt.Printf(" (v%s)", provider.Version)
		}
		if provider.InstallationPath != "" {
			fmt.Printf(" at %s", provider.InstallationPath)
		}
		fmt.Println()

		// Show features/details for the provider
		if len(provider.Features) > 0 {
			for _, feature := range provider.Features {
				fmt.Printf("    └─ %s\n", feature)
			}
		}

		// Show recommendations/warnings
		if len(provider.Recommendations) > 0 {
			for _, rec := range provider.Recommendations {
				fmt.Printf("    %s\n", rec)
			}
		}
	}
}

func handleListCommand(args []string) {
	fmt.Println("Listing virtual machines...")

	manager, err := NewVirtManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	vms, err := manager.List()
	if err != nil {
		fmt.Printf("\033[31mError listing VMs: %v\033[0m\n", err)

		// Check for common VirtualBox E_ACCESSDENIED error
		if strings.Contains(err.Error(), "E_ACCESSDENIED") || strings.Contains(err.Error(), "0x80070005") {
			fmt.Printf("\n\033[33m💡 This looks like a VirtualBox permission issue.\033[0m\n")
			fmt.Printf("Common solutions:\n")
			fmt.Printf("1. Reinstall VirtualBox (fixes 80%% of cases): Uninstall, restart, reinstall as admin\n")
			fmt.Printf("2. Restart VirtualBox service: Stop all VBox* processes in Task Manager\n")
			fmt.Printf("3. Run as administrator (temporary fix)\n")
			fmt.Printf("4. Fix COM permissions in DCOMCNFG\n")

			// Try to show detailed help from file
			showVirtualBoxHelp()
		}
		return
	}

	if len(vms) == 0 {
		fmt.Printf("No VMs found. Create one with: portunix virt create <name>\n")
		return
	}

	// Get provider version
	providerName := manager.GetProviderName()
	providerVersion := manager.GetProviderVersion()
	if providerVersion != "" {
		fmt.Printf("Provider: %s (%s)\n\n", providerName, providerVersion)
	} else {
		fmt.Printf("Provider: %s\n\n", providerName)
	}
	fmt.Printf("%-20s %-12s %-8s %-6s %-10s %-15s\n", "NAME", "STATE", "RAM", "CPUS", "DISK", "IP")
	fmt.Printf("%-20s %-12s %-8s %-6s %-10s %-15s\n", "----", "-----", "---", "----", "----", "--")

	errorCount := 0
	accessDeniedCount := 0
	for _, vm := range vms {
		ip := vm.IP
		if ip == "" {
			ip = "-"
		}

		// Color code the state
		var stateDisplay string
		switch string(vm.State) {
		case "running":
			stateDisplay = "\033[32m" + string(vm.State) + "\033[0m" // Green
		case "stopped":
			stateDisplay = "\033[33m" + string(vm.State) + "\033[0m" // Yellow
		case "error":
			stateDisplay = "\033[31m" + string(vm.State) + "\033[0m" // Red
			errorCount++
			// Check if this is specifically an access denied error
			if strings.Contains(vm.ErrorDetail, "Access denied") || strings.Contains(vm.ErrorDetail, "administrator") {
				accessDeniedCount++
			}
		case "not-found":
			stateDisplay = "\033[91m" + string(vm.State) + "\033[0m" // Bright red
		case "unknown":
			stateDisplay = "\033[90m" + string(vm.State) + "\033[0m" // Gray
		default:
			stateDisplay = string(vm.State)
		}

		fmt.Printf("%-20s %-12s %-8s %-6d %-10s %-15s\n",
			vm.Name, stateDisplay, vm.RAM, vm.CPUs, vm.DiskSize, ip)

		// Show error detail if available
		if vm.ErrorDetail != "" && string(vm.State) == "error" {
			fmt.Printf("    \033[90m└─ %s\033[0m\n", vm.ErrorDetail)
		}
	}

	// If most VMs are in error state, show help
	if errorCount > 0 {
		if accessDeniedCount > 0 {
			fmt.Printf("\n\033[33m💡 %d VM(s) have access permission issues.\033[0m\n", accessDeniedCount)
			fmt.Printf("This is usually caused by VirtualBox installation/permission problems.\n")
			fmt.Printf("Solutions:\n")
			fmt.Printf("1. Run as administrator: Right-click terminal → 'Run as administrator'\n")
			fmt.Printf("2. Reinstall VirtualBox: Uninstall → Restart → Reinstall as admin\n")
			fmt.Printf("3. See detailed help: portunix virt help virtualbox\n\n")
		} else if errorCount == len(vms) {
			fmt.Printf("\n\033[33m⚠️  All VMs are showing error state. This might indicate a VirtualBox problem.\033[0m\n")
			showVirtualBoxHelp()
		}
	}
}

func handleCreateCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("VM name required")
		fmt.Println("Usage: portunix virt create [vm-name] [options]")
		return
	}
	fmt.Printf("Creating VM: %s\n", args[0])
	// TODO: Implement using existing virt create logic
}

func handleStartCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("VM name required")
		fmt.Println("Usage: portunix virt start <vm-name>")
		return
	}

	vmName := args[0]
	fmt.Printf("Starting VM '%s'...\n", vmName)

	manager, err := NewVirtManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if err := manager.Start(vmName); err != nil {
		fmt.Printf("Error starting VM: %v\n", err)
		return
	}

	fmt.Printf("✅ VM '%s' started successfully!\n", vmName)
}

func handleStopCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("VM name required")
		return
	}
	fmt.Printf("Stopping VM: %s\n", args[0])
	// TODO: Implement using existing virt stop logic
}

func handleRestartCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("VM name required")
		return
	}
	fmt.Printf("Restarting VM: %s\n", args[0])
	// TODO: Implement using existing virt restart logic
}

func handleDeleteCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("VM name required")
		return
	}
	fmt.Printf("Deleting VM: %s\n", args[0])
	// TODO: Implement using existing virt delete logic
}

func handleInfoCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("VM name required")
		return
	}
	fmt.Printf("Getting info for VM: %s\n", args[0])
	// TODO: Implement using existing virt info logic
}

func handleStatusCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("VM name required")
		return
	}
	fmt.Printf("Getting status for VM: %s\n", args[0])
	// TODO: Implement using existing virt status logic
}

func handleSSHCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("VM name required")
		return
	}
	fmt.Printf("SSH into VM: %s\n", args[0])
	// TODO: Implement using existing virt SSH logic
}

func handleSnapshotCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Snapshot subcommand required")
		fmt.Println("Usage: portunix virt snapshot [create|list|restore|delete] [vm-name] [snapshot-name]")
		return
	}
	fmt.Printf("Snapshot operation: %s\n", args[0])
	// TODO: Implement using existing virt snapshot logic
}

func handleInstallQEMUCommand() {
	fmt.Println("Installing QEMU/KVM...")
	// TODO: Implement using existing QEMU installation logic
}

func init() {
	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version")
	rootCmd.Flags().Bool("description", false, "Show description")
	rootCmd.Flags().Bool("list-commands", false, "List available commands")
}

func showVirtualBoxHelp() {
	fmt.Printf("\n\033[36m──────────────────────────────────────────────────────\033[0m\n")
	fmt.Printf("\033[36mDetailed VirtualBox E_ACCESSDENIED Solutions:\033[0m\n")
	fmt.Printf("\033[36m──────────────────────────────────────────────────────\033[0m\n\n")

	// Use embedded content
	lines := strings.Split(virtualBoxHelpContent, "\n")
	maxLines := 50
	if len(lines) < maxLines {
		maxLines = len(lines)
	}

	for i := 0; i < maxLines; i++ {
		fmt.Println(lines[i])
	}

	if len(lines) > 50 {
		fmt.Printf("\n... (showing first 50 lines of embedded help)\n")
		fmt.Printf("To see full documentation, run: portunix virt help e-accessdenied\n")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}