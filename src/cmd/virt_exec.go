package cmd

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "strings"

    "github.com/spf13/cobra"
    "portunix.ai/app/virt"
)

// virtExecExtendedCmd extends the existing exec command with Windows support
// This is implemented as an enhancement to the existing command in virt_ssh.go

var virtCheckCmdFallback = &cobra.Command{
    Use:   "check",
    Short: "Check virtualization requirements",
    Long: `Check if the system meets the requirements for virtualization.

This command verifies:
- CPU virtualization support (VT-x/AMD-V)
- KVM availability on Linux
- QEMU installation
- UEFI firmware availability (for Windows VMs)
- TPM emulation support (for Windows 11)
- Hyper-V status on Windows`,
    Run: func(cmd *cobra.Command, args []string) {
        // Fallback implementation
        fmt.Println("Virtualization check: Feature not yet fully implemented")
        fmt.Println("Use 'portunix system info' for basic virtualization information")
        /*if err := virt.CheckVirtualizationRequirements(); err != nil {
            fmt.Fprintf(os.Stderr, "Virtualization check failed: %v\n", err)
            os.Exit(1)
        }*/
    },
}

var virtCheckCmd = &cobra.Command{
    Use:   "check",
    Short: "Check virtualization requirements",
    Long: `Check if the system meets the requirements for virtualization.

This command verifies:
- CPU virtualization support (VT-x/AMD-V)
- KVM availability on Linux
- QEMU installation
- UEFI firmware availability (for Windows VMs)
- TPM emulation support (for Windows 11)
- Hyper-V status on Windows
- Virtualization conflicts (VirtualBox/KVM)

Use --fix flag to interactively resolve detected conflicts.`,
    Run: handleVirtCheck,
}

var virtInstallQEMUCmd = &cobra.Command{
    Use:   "install-qemu",
    Short: "Install QEMU and required components",
    Long: `Install QEMU and virtualization components using Portunix package system.

This is a convenience wrapper around 'portunix install qemu' which installs:
- QEMU system emulator
- QEMU utilities
- OVMF (UEFI firmware)
- swtpm (TPM emulator for Windows 11)
- Bridge utilities for networking
- Libvirt for VM management

Supported distributions:
- Ubuntu/Debian (apt)
- Fedora/RHEL/CentOS (dnf/yum)
- Arch Linux (pacman)`,
    Run: func(cmd *cobra.Command, args []string) {
        // Use the existing Portunix install system
        fmt.Println("Installing QEMU using Portunix package system...")

        // This delegates to the existing 'portunix install qemu' command
        installCmd := exec.Command("portunix", "install", "qemu")
        installCmd.Stdout = os.Stdout
        installCmd.Stderr = os.Stderr
        installCmd.Stdin = os.Stdin

        if err := installCmd.Run(); err != nil {
            fmt.Fprintf(os.Stderr, "QEMU installation failed: %v\n", err)
            fmt.Fprintf(os.Stderr, "You can also try: portunix install qemu\n")
            os.Exit(1)
        }

        fmt.Println("\n✅ QEMU installation completed via Portunix package system")
    },
}

func init() {
    // Register check command that can delegate to helper
    virtCmd.AddCommand(virtCheckCmd)

    // Add flags for conflict resolution
    virtCheckCmd.Flags().Bool("fix", false, "Interactively fix detected conflicts")
    virtCheckCmd.Flags().Bool("unload-kvm", false, "Unload KVM modules (use with --fix)")
    virtCheckCmd.Flags().Bool("blacklist-kvm", false, "Blacklist KVM permanently (use with --fix)")
    virtCheckCmd.Flags().Bool("use-kvm", false, "Switch to KVM (use with --fix)")
    virtCheckCmd.Flags().Bool("fix-libvirt", false, "Fix libvirt daemon issues")
    virtCheckCmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
}

func handleVirtCheck(cmd *cobra.Command, args []string) {
    fix, _ := cmd.Flags().GetBool("fix")
    unloadKVM, _ := cmd.Flags().GetBool("unload-kvm")
    blacklistKVM, _ := cmd.Flags().GetBool("blacklist-kvm")
    useKVM, _ := cmd.Flags().GetBool("use-kvm")
    fixLibvirt, _ := cmd.Flags().GetBool("fix-libvirt")
    dryRun, _ := cmd.Flags().GetBool("dry-run")

    // If libvirt fix flag is set, handle it
    if fixLibvirt {
        handleLibvirtFix(cmd, dryRun)
        return
    }

    // If any conflict resolution flags are set, handle them
    if fix || unloadKVM || blacklistKVM || useKVM {
        handleConflictResolution(cmd, unloadKVM, blacklistKVM, useKVM, dryRun)
        return
    }

    // Otherwise, delegate to helper or fallback
    virtWithHelperCheck(virtCheckCmdFallback)(cmd, args)
}

func handleConflictResolution(cmd *cobra.Command, unloadKVM, blacklistKVM, useKVM, dryRun bool) {
    // Detect conflicts
    conflict, err := virt.DetectVirtualizationConflict()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error detecting conflicts: %v\n", err)
        os.Exit(1)
    }

    // Also check libvirt status
    libvirtStatus, libvirtErr := virt.DetectLibvirtStatus()

    // If no conflict, check libvirt and inform user
    if !conflict.Conflict {
        fmt.Println("✅ No virtualization conflicts detected")
        if conflict.VirtualBoxPresent {
            fmt.Println("   VirtualBox: Installed")
        }
        if conflict.KVMActive {
            fmt.Printf("   KVM: Active (%s)\n", strings.Join(conflict.LoadedModules, ", "))
        }

        // Check if libvirt needs fixing
        if libvirtErr == nil && libvirtStatus.Installed && len(libvirtStatus.Issues) > 0 {
            fmt.Println("\n⚠️  Libvirt Issues Detected:")
            for _, issue := range libvirtStatus.Issues {
                fmt.Printf("   • %s\n", issue)
            }
            fmt.Println("\n💡 Fix libvirt: sudo portunix virt check --fix-libvirt")
        }
        return
    }

    // Display conflict information
    fmt.Println("\n⚠️  VIRTUALIZATION CONFLICT DETECTED")
    fmt.Println("════════════════════════════════════════════════════════════")
    fmt.Printf("Conflict Type: %s\n", conflict.Type)
    fmt.Printf("Details: %s\n\n", conflict.Details)

    if conflict.VirtualBoxPresent {
        fmt.Println("✅ VirtualBox: Installed")
    }
    if conflict.KVMActive {
        fmt.Printf("✅ KVM: Active\n")
        fmt.Printf("   Loaded modules: %s\n", strings.Join(conflict.LoadedModules, ", "))
    }

    fmt.Println("\n💡 Recommendation:", conflict.Recommendation)
    fmt.Println("════════════════════════════════════════════════════════════\n")

    // If specific action flags are set, execute them
    if unloadKVM {
        executeUnloadKVM(dryRun)
        return
    }
    if blacklistKVM {
        executeBlacklistKVM(dryRun)
        return
    }
    if useKVM {
        executeSwitchToKVM(dryRun)
        return
    }

    // Interactive mode - show menu
    showConflictResolutionMenu(conflict, libvirtStatus, dryRun)
}

func showConflictResolutionMenu(conflict *virt.VirtualizationConflict, libvirtStatus *virt.LibvirtStatus, dryRun bool) {
    fmt.Println("Select resolution option:")
    fmt.Println()
    fmt.Println("1. Temporarily unload KVM (quick fix)")
    fmt.Println("   → Allows VirtualBox to run immediately")
    fmt.Println("   → KVM will reload on next boot")
    fmt.Println()
    fmt.Println("2. Permanently disable KVM (blacklist)")
    fmt.Println("   → Prevents KVM from loading on boot")
    fmt.Println("   → Requires reboot to take effect")
    fmt.Println()
    fmt.Println("3. Switch to KVM (unload VirtualBox)")
    fmt.Println("   → Unloads VirtualBox modules")
    fmt.Println("   → Loads KVM for QEMU usage")
    fmt.Println()

    // Add libvirt option if there are issues
    hasLibvirtIssues := libvirtStatus != nil && libvirtStatus.Installed && len(libvirtStatus.Issues) > 0
    if hasLibvirtIssues {
        fmt.Println("4. Fix libvirt daemon")
        fmt.Println("   → Unmask and start libvirt daemon")
        fmt.Println("   → Allows virt-manager to connect")
        fmt.Println()
        fmt.Println("5. Cancel (no changes)")
        fmt.Println()
        fmt.Print("Choose option [1-5]: ")
    } else {
        fmt.Println("4. Cancel (no changes)")
        fmt.Println()
        fmt.Print("Choose option [1-4]: ")
    }

    reader := bufio.NewReader(os.Stdin)
    choice, _ := reader.ReadString('\n')
    choice = strings.TrimSpace(choice)

    switch choice {
    case "1":
        executeUnloadKVM(dryRun)
    case "2":
        executeBlacklistKVM(dryRun)
    case "3":
        executeSwitchToKVM(dryRun)
    case "4":
        if hasLibvirtIssues {
            // Fix libvirt
            if libvirtStatus.Masked {
                executeUnmaskLibvirt(libvirtStatus.DaemonName, dryRun)
            } else if !libvirtStatus.Running {
                executeStartLibvirt(libvirtStatus.DaemonName, dryRun)
            }
            if !libvirtStatus.Enabled && !libvirtStatus.SocketActivated {
                executeEnableLibvirt(libvirtStatus.DaemonName, dryRun)
            }
        } else {
            fmt.Println("Cancelled. No changes made.")
        }
    case "5":
        if hasLibvirtIssues {
            fmt.Println("Cancelled. No changes made.")
        } else {
            fmt.Println("Invalid choice. No changes made.")
        }
    default:
        fmt.Println("Invalid choice. No changes made.")
    }
}

func executeUnloadKVM(dryRun bool) {
    fmt.Println("\n🔧 Unloading KVM modules...")

    if dryRun {
        fmt.Println("[DRY RUN] Would execute: sudo rmmod kvm_intel kvm_amd kvm")
        return
    }

    if err := virt.UnloadKVMModules(); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to unload KVM modules: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("✅ KVM modules unloaded successfully")
    fmt.Println("   VirtualBox should now work correctly")
    fmt.Println("\n💡 Note: KVM will reload on next boot")
    fmt.Println("   For permanent solution, use: sudo portunix virt check --fix --blacklist-kvm")
}

func executeBlacklistKVM(dryRun bool) {
    fmt.Println("\n🔧 Creating KVM blacklist configuration...")

    if dryRun {
        fmt.Println("[DRY RUN] Would create: /etc/modprobe.d/blacklist-kvm.conf")
        fmt.Println("[DRY RUN] Would update initramfs")
        return
    }

    if err := virt.BlacklistKVMModules(); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to blacklist KVM: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("✅ KVM blacklist created successfully")
    fmt.Println("   File: /etc/modprobe.d/blacklist-kvm.conf")
    fmt.Println("\n⚠️  Reboot required for changes to take effect")
    fmt.Println("   KVM will not load after reboot")
}

func executeSwitchToKVM(dryRun bool) {
    fmt.Println("\n🔧 Switching to KVM...")

    if dryRun {
        fmt.Println("[DRY RUN] Would execute:")
        fmt.Println("  1. sudo rmmod vboxnetadp vboxnetflt vboxpci vboxdrv")
        fmt.Println("  2. sudo modprobe kvm kvm_amd (or kvm_intel)")
        return
    }

    if err := virt.SwitchToKVM(); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to switch to KVM: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("✅ Switched to KVM successfully")
    fmt.Println("   VirtualBox modules unloaded")
    fmt.Println("   KVM modules loaded")
    fmt.Println("\n💡 QEMU/KVM VMs should now work correctly")
}

func handleLibvirtFix(cmd *cobra.Command, dryRun bool) {
    status, err := virt.DetectLibvirtStatus()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error detecting libvirt: %v\n", err)
        os.Exit(1)
    }

    if !status.Installed {
        fmt.Println("❌ Libvirt is not installed")
        fmt.Println("   Install: portunix install libvirt")
        return
    }

    fmt.Println("📊 Libvirt Status:")
    fmt.Printf("   Version: %s\n", status.Version)
    fmt.Printf("   Daemon type: %s\n", status.DaemonType)
    fmt.Printf("   Daemon name: %s\n", status.DaemonName)
    fmt.Printf("   Running: %v\n", status.Running)
    fmt.Printf("   Enabled: %v\n", status.Enabled)
    fmt.Printf("   Masked: %v\n", status.Masked)
    if status.SocketName != "" {
        fmt.Printf("   Socket: %s\n", status.SocketName)
        fmt.Printf("   Socket activated: %v\n", status.SocketActivated)
        if status.SocketFailed {
            fmt.Printf("   Socket state: ❌ FAILED (%s)\n", status.SocketState)
        }
    }

    // Show dependency status if any issues
    if len(status.MaskedDependencies) > 0 {
        fmt.Printf("   ❌ Masked dependencies: %s\n", strings.Join(status.MaskedDependencies, ", "))
    }
    if len(status.MissingDependencies) > 0 {
        fmt.Printf("   ❌ Missing dependencies: %s\n", strings.Join(status.MissingDependencies, ", "))
    }
    if len(status.FailedDependencies) > 0 {
        fmt.Printf("   ❌ Failed dependencies: %s\n", strings.Join(status.FailedDependencies, ", "))
    }

    fmt.Println()

    if len(status.Issues) == 0 {
        fmt.Println("✅ Libvirt daemon is running correctly")
        return
    }

    // Show issues
    fmt.Println("⚠️  Libvirt Issues Detected:")
    for _, issue := range status.Issues {
        fmt.Printf("   • %s\n", issue)
    }
    fmt.Println()

    // Fix based on issue type (priority order)
    // Dependencies FIRST - they cause other failures
    if len(status.MaskedDependencies) > 0 {
        executeUnmaskDependencies(status.MaskedDependencies, dryRun)
    } else if len(status.MissingDependencies) > 0 {
        executeInstallMissingDependencies(status.MissingDependencies, dryRun)
    } else if len(status.FailedDependencies) > 0 {
        executeResetDependencies(status.FailedDependencies, dryRun)
    } else if status.SocketFailed {
        // Socket is in failed state - offer reset or switch to direct daemon
        executeFixFailedSocket(status, dryRun)
    } else if status.Masked {
        executeUnmaskLibvirt(status.DaemonName, dryRun)
    } else if !status.Running && !status.SocketActivated {
        executeStartLibvirt(status.DaemonName, dryRun)
    } else if !status.Enabled && !status.SocketActivated {
        executeEnableLibvirt(status.DaemonName, dryRun)
    }
}

func executeUnmaskLibvirt(daemonName string, dryRun bool) {
    fmt.Println("🔧 Unmasking libvirt daemon...")

    if dryRun {
        fmt.Printf("[DRY RUN] Would execute:\n")
        fmt.Printf("  sudo systemctl unmask %s\n", daemonName)
        fmt.Printf("  sudo systemctl start %s\n", daemonName)
        return
    }

    if err := virt.UnmaskLibvirtDaemon(daemonName); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to unmask and start %s: %v\n", daemonName, err)
        os.Exit(1)
    }

    fmt.Println("✅ Libvirt daemon unmasked and started")
}

func executeStartLibvirt(daemonName string, dryRun bool) {
    fmt.Println("🔧 Starting libvirt daemon...")

    if dryRun {
        fmt.Printf("[DRY RUN] Would execute: sudo systemctl start %s\n", daemonName)
        return
    }

    if err := virt.StartLibvirtDaemon(daemonName); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to start %s: %v\n", daemonName, err)
        os.Exit(1)
    }

    fmt.Println("✅ Libvirt daemon started")
}

func executeEnableLibvirt(daemonName string, dryRun bool) {
    fmt.Println("🔧 Enabling libvirt daemon on boot...")

    if dryRun {
        fmt.Printf("[DRY RUN] Would execute: sudo systemctl enable %s\n", daemonName)
        return
    }

    if err := virt.EnableLibvirtDaemon(daemonName); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to enable %s: %v\n", daemonName, err)
        os.Exit(1)
    }

    fmt.Println("✅ Libvirt daemon enabled on boot")
}

func executeFixFailedSocket(status *virt.LibvirtStatus, dryRun bool) {
    fmt.Println("🔧 Fixing failed libvirt socket...")
    fmt.Println()
    fmt.Println("Socket is in failed state. Choose fix option:")
    fmt.Println()
    fmt.Println("1. Reset socket and retry")
    fmt.Println("   → Reset failed state and restart socket activation")
    fmt.Println("   → Recommended if socket failed temporarily")
    fmt.Println()
    fmt.Println("2. Switch to direct daemon")
    fmt.Println("   → Disable socket activation")
    fmt.Println("   → Enable and start daemon directly")
    fmt.Println("   → Recommended if socket keeps failing")
    fmt.Println()
    fmt.Println("3. Cancel (no changes)")
    fmt.Println()
    fmt.Print("Choose option [1-3]: ")

    reader := bufio.NewReader(os.Stdin)
    choice, _ := reader.ReadString('\n')
    choice = strings.TrimSpace(choice)

    switch choice {
    case "1":
        executeResetSocket(status.SocketName, dryRun)
    case "2":
        executeSwitchToDirectDaemon(status.DaemonName, status.SocketName, dryRun)
    case "3":
        fmt.Println("Cancelled. No changes made.")
    default:
        fmt.Println("Invalid choice. No changes made.")
    }
}

func executeResetSocket(socketName string, dryRun bool) {
    fmt.Println("\n🔧 Resetting failed socket...")

    if dryRun {
        fmt.Printf("[DRY RUN] Would execute:\n")
        fmt.Printf("  sudo systemctl reset-failed %s\n", socketName)
        daemonName := strings.Replace(socketName, ".socket", "", 1)
        fmt.Printf("  sudo systemctl reset-failed %s\n", daemonName)
        fmt.Printf("  sudo systemctl restart %s\n", socketName)
        return
    }

    if err := virt.ResetFailedSocket(socketName); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to reset socket: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("✅ Socket reset successfully")
    fmt.Println("   Try connecting with virt-manager now")
}

func executeSwitchToDirectDaemon(daemonName, socketName string, dryRun bool) {
    fmt.Println("\n🔧 Switching to direct daemon...")

    if dryRun {
        fmt.Printf("[DRY RUN] Would execute:\n")
        fmt.Printf("  sudo systemctl stop %s\n", socketName)
        fmt.Printf("  sudo systemctl disable %s\n", socketName)
        fmt.Printf("  sudo systemctl enable %s\n", daemonName)
        fmt.Printf("  sudo systemctl start %s\n", daemonName)
        return
    }

    if err := virt.SwitchToDirectDaemon(daemonName, socketName); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to switch to direct daemon: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("✅ Switched to direct daemon successfully")
    fmt.Println("   Socket activation disabled")
    fmt.Println("   Daemon is now running directly")
    fmt.Println("   Try connecting with virt-manager now")
}

func executeUnmaskDependencies(deps []string, dryRun bool) {
    fmt.Println("🔧 Unmasking dependencies...")

    for _, dep := range deps {
        fmt.Printf("   Unmasking %s...\n", dep)

        if dryRun {
            fmt.Printf("   [DRY RUN] Would execute: sudo systemctl unmask %s\n", dep)
            continue
        }

        if err := virt.UnmaskDependency(dep); err != nil {
            fmt.Fprintf(os.Stderr, "   ⚠️  Failed to unmask %s: %v\n", dep, err)
        } else {
            fmt.Printf("   ✅ Unmasked %s\n", dep)
        }
    }

    if !dryRun {
        fmt.Println("\n💡 Now try: sudo portunix virt check --fix-libvirt")
        fmt.Println("   This will restart the libvirt daemon")
    }
}

func executeResetDependencies(deps []string, dryRun bool) {
    fmt.Println("🔧 Resetting failed dependencies...")

    for _, dep := range deps {
        fmt.Printf("   Resetting %s...\n", dep)

        if dryRun {
            fmt.Printf("   [DRY RUN] Would execute:\n")
            fmt.Printf("     sudo systemctl reset-failed %s\n", dep)
            fmt.Printf("     sudo systemctl start %s\n", dep)
            continue
        }

        if err := virt.ResetFailedDependency(dep); err != nil {
            fmt.Fprintf(os.Stderr, "   ⚠️  Failed to reset %s: %v\n", dep, err)
        } else {
            fmt.Printf("   ✅ Reset %s\n", dep)
        }
    }

    if !dryRun {
        fmt.Println("\n💡 Dependencies reset. Libvirt should start now.")
    }
}

func executeInstallMissingDependencies(deps []string, dryRun bool) {
    fmt.Println("⚠️  Missing libvirt dependencies detected:")
    for _, dep := range deps {
        fmt.Printf("   • %s\n", dep)
    }
    fmt.Println()

    if dryRun {
        fmt.Println("[DRY RUN] Would execute: portunix install libvirt")
        return
    }

    // Ask user for confirmation
    fmt.Print("Install libvirt package? [y/N]: ")
    reader := bufio.NewReader(os.Stdin)
    response, _ := reader.ReadString('\n')
    response = strings.TrimSpace(strings.ToLower(response))

    if response != "y" && response != "yes" {
        fmt.Println("Installation cancelled.")
        fmt.Println("💡 To install manually: portunix install libvirt")
        return
    }

    // Use standard portunix install system
    fmt.Println("\n🔧 Installing libvirt...")
    installCmd := exec.Command("portunix", "install", "libvirt")
    installCmd.Stdout = os.Stdout
    installCmd.Stderr = os.Stderr
    installCmd.Stdin = os.Stdin

    if err := installCmd.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Installation failed: %v\n", err)
        fmt.Println("💡 Try manually: portunix install libvirt")
        return
    }

    fmt.Println("\n✅ Libvirt installed successfully")
    fmt.Println("💡 Now run: sudo portunix virt check --fix-libvirt")
}