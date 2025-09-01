package cmd

import (
	"github.com/spf13/cobra"
)

// vmCmd represents the vm command
var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Manage virtual machines with QEMU/KVM",
	Long: `Manage virtual machines using QEMU/KVM for running Windows and Linux guests.
	
This command provides functionality for:
- Creating and managing VMs
- Snapshot management for easy rollback
- Windows 11 support with TPM and Secure Boot
- Resource configuration (CPU, RAM, disk)`,
	Example: `  portunix vm install-qemu           # Install QEMU/KVM stack
  portunix vm create windows11       # Create a Windows 11 VM
  portunix vm snapshot create myvm   # Create a snapshot
  portunix vm start myvm              # Start a VM`,
}

func init() {
	rootCmd.AddCommand(vmCmd)
}
