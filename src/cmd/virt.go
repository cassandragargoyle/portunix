package cmd

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"

    "github.com/spf13/cobra"
)

// virtCmd represents the virt command - now delegates to ptx-virt helper
var virtCmd = &cobra.Command{
    Use:   "virt",
    Short: "Virtual machine management",
    Long: `Manage virtual machines using VirtualBox, QEMU/KVM, VMware, or Hyper-V.

This command provides comprehensive VM management capabilities including:
- Create and manage VMs
- Install operating systems from ISOs
- Execute commands in VMs
- Manage snapshots
- Configure networking

The virtualization backend is automatically detected based on available providers.
Supported backends: VirtualBox, QEMU/KVM, VMware, Hyper-V

Use 'portunix virt check' to verify your system meets virtualization requirements.`,
    Run: func(cmd *cobra.Command, args []string) {
        // Try to dispatch to ptx-virt helper first
        if helperPath := findVirtHelper(); helperPath != "" {
            if err := dispatchToVirtHelper(helperPath, args); err != nil {
                fmt.Fprintf(os.Stderr, "Helper execution failed: %v\n", err)
                fmt.Fprintf(os.Stderr, "Falling back to built-in virt commands...\n")
                // Fall through to show help for built-in commands
            } else {
                return // Helper succeeded
            }
        }

        // Fallback: show help for built-in virt commands
        cmd.Help()
    },
}

// findVirtHelper finds the ptx-virt helper binary
func findVirtHelper() string {
    // Determine binary name with platform suffix
    helperName := "ptx-virt"
    if runtime.GOOS == "windows" {
        helperName += ".exe"
    }

    // Method 1: Check in same directory as main binary
    if execPath, err := os.Executable(); err == nil {
        execDir := filepath.Dir(execPath)
        helperPath := filepath.Join(execDir, helperName)
        if _, err := os.Stat(helperPath); err == nil {
            return helperPath
        }
    }

    // Method 2: Check in PATH
    if helperPath, err := exec.LookPath(helperName); err == nil {
        return helperPath
    }

    return "" // Helper not found
}

// dispatchToVirtHelper dispatches virt commands to the ptx-virt helper binary
func dispatchToVirtHelper(helperPath string, args []string) error {
    // Prepare arguments: ["virt"] + original args
    cmdArgs := append([]string{"virt"}, args...)

    // Execute helper binary
    cmd := exec.Command(helperPath, cmdArgs...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}

// virtWithHelperCheck creates a command wrapper that tries helper first, then fallback
func virtWithHelperCheck(fallbackCmd *cobra.Command) func(cmd *cobra.Command, args []string) {
    return func(cmd *cobra.Command, args []string) {
        // Try helper first
        if helperPath := findVirtHelper(); helperPath != "" {
            // Build full command path for helper by walking up parent commands
            var commandPath []string
            current := cmd
            for current != nil && current.Name() != "virt" {
                commandPath = append([]string{current.Name()}, commandPath...)
                current = current.Parent()
            }

            // Build command arguments for helper: ["virt", "subcommand", "subsubcommand", args...]
            helperArgs := commandPath
            helperArgs = append(helperArgs, args...)

            if err := dispatchToVirtHelper(helperPath, helperArgs); err == nil {
                return // Helper succeeded
            }
            // On helper failure, fall through to fallback
            fmt.Fprintf(os.Stderr, "Helper failed, using fallback implementation...\n")
        }

        // Use fallback implementation
        fallbackCmd.Run(cmd, args)
    }
}

func init() {
    rootCmd.AddCommand(virtCmd)

    // Subcommands will be added by individual virt_*.go files
    // They will use virtWithHelperCheck to try helper first, then fallback
}