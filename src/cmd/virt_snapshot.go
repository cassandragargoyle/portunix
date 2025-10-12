package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"portunix.ai/app/virt"
)

// virtSnapshotCmd represents the virt snapshot command
var virtSnapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage virtual machine snapshots",
	Long: `Manage virtual machine snapshots for easy rollback and system state preservation.

Snapshots allow you to:
- Save the current state of a VM before making changes
- Quickly revert to a known good state
- Create multiple restore points during development
- Test software installation and configuration safely

Available subcommands:
  create  - Create a new snapshot
  list    - List all snapshots for a VM
  revert  - Revert VM to a snapshot
  delete  - Delete a snapshot

Examples:
  portunix virt snapshot create ubuntu-test clean-install
  portunix virt snapshot list ubuntu-test
  portunix virt snapshot revert ubuntu-test clean-install
  portunix virt snapshot delete ubuntu-test old-snapshot`,
}

// virtSnapshotCreateCmd represents the snapshot create command
var virtSnapshotCreateCmd = &cobra.Command{
	Use:   "create [vm-name] [snapshot-name]",
	Short: "Create a new snapshot of a virtual machine",
	Long: `Create a new snapshot of a virtual machine's current state.

The snapshot captures the entire VM state including:
- Disk contents
- Memory state (if VM is running)
- Configuration settings

Snapshots are useful for:
- Creating restore points before system changes
- Testing software installation
- Development environment checkpoints
- Backup before risky operations

Examples:
  portunix virt snapshot create ubuntu-test before-update
  portunix virt snapshot create web-server "Clean LAMP setup" --description "Fresh Ubuntu with Apache, MySQL, PHP"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		snapshotName := args[1]
		description, _ := cmd.Flags().GetString("description")

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Creating snapshot '%s' for VM '%s'...\n", snapshotName, vmName)
		if description != "" {
			fmt.Printf("Description: %s\n", description)
		}

		if err := manager.CreateSnapshot(vmName, snapshotName, description); err != nil {
			fmt.Printf("❌ Failed to create snapshot: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Snapshot '%s' created successfully!\n", snapshotName)
		fmt.Printf("\nTo revert to this snapshot later:\n")
		fmt.Printf("  portunix virt snapshot revert %s %s\n", vmName, snapshotName)
	},
}

// virtSnapshotListCmdFallback represents the snapshot list command fallback implementation
var virtSnapshotListCmdFallback = &cobra.Command{
	Use:   "list [vm-name]",
	Short: "List all snapshots for a virtual machine",
	Long:  `List all snapshots for a virtual machine with creation times and descriptions.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		snapshots, err := manager.ListSnapshots(vmName)
		if err != nil {
			fmt.Printf("❌ Failed to list snapshots: %v\n", err)
			os.Exit(1)
		}

		if len(snapshots) == 0 {
			fmt.Printf("No snapshots found for VM '%s'\n", vmName)
			fmt.Printf("\nCreate a snapshot with:\n")
			fmt.Printf("  portunix virt snapshot create %s <snapshot-name>\n", vmName)
			return
		}

		fmt.Printf("Snapshots for VM '%s':\n\n", vmName)
		fmt.Printf("%-20s %-20s %-10s %s\n", "NAME", "CREATED", "SIZE", "DESCRIPTION")
		fmt.Printf("%-20s %-20s %-10s %s\n", "----", "-------", "----", "-----------")

		for _, snapshot := range snapshots {
			// Handle empty/zero timestamp
			createdStr := "-"
			if !snapshot.CreatedAt.IsZero() {
				createdStr = snapshot.CreatedAt.Format("2006-01-02 15:04")
			}

			sizeStr := formatBytes(snapshot.Size)
			description := snapshot.Description
			if description == "" {
				description = "-"
			}

			fmt.Printf("%-20s %-20s %-10s %s\n",
				snapshot.Name, createdStr, sizeStr, description)
		}

		fmt.Printf("\nTo revert to a snapshot:\n")
		fmt.Printf("  portunix virt snapshot revert %s <snapshot-name>\n", vmName)
	},
}

// virtSnapshotListCmd represents the snapshot list command (fallback only until helper is implemented)
var virtSnapshotListCmd = &cobra.Command{
	Use:   "list [vm-name]",
	Short: "List all snapshots for a virtual machine",
	Long:  `List all snapshots for a virtual machine with creation times and descriptions.`,
	Args:  cobra.ExactArgs(1),
	Run:   virtSnapshotListCmdFallback.Run,
}

// virtSnapshotRevertCmd represents the snapshot revert command
var virtSnapshotRevertCmd = &cobra.Command{
	Use:   "revert [vm-name] [snapshot-name]",
	Short: "Revert a virtual machine to a snapshot",
	Long: `Revert a virtual machine to a previously created snapshot.

WARNING: This operation will:
- Restore the VM to the exact state when the snapshot was created
- Lose any changes made since the snapshot was created
- Stop the VM if it's currently running (required for revert)

The revert process:
1. Stops the VM if running
2. Restores disk state from snapshot
3. VM remains stopped after revert

Examples:
  portunix virt snapshot revert ubuntu-test clean-install
  portunix virt snapshot revert web-server before-update --force`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		snapshotName := args[1]
		force, _ := cmd.Flags().GetBool("force")

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Check VM state
		state := manager.GetBackend().GetState(vmName)
		if state == virt.VMStateError || state == virt.VMStateNotFound {
			fmt.Printf("Error: VM '%s' not found\n", vmName)
			os.Exit(1)
		}

		// Warn about running VM
		if state == virt.VMStateRunning && !force {
			fmt.Printf("⚠️  WARNING: VM '%s' is currently running.\n", vmName)
			fmt.Printf("Reverting to snapshot will stop the VM and lose any unsaved changes.\n\n")
			fmt.Printf("Are you sure you want to continue? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Revert cancelled.")
				return
			}
		}

		fmt.Printf("🔄 Reverting VM '%s' to snapshot '%s'...\n", vmName, snapshotName)

		if err := manager.RevertSnapshot(vmName, snapshotName); err != nil {
			fmt.Printf("❌ Failed to revert snapshot: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ VM '%s' reverted to snapshot '%s' successfully!\n", vmName, snapshotName)
		fmt.Printf("\nVM is now stopped. To start it:\n")
		fmt.Printf("  portunix virt start %s\n", vmName)
	},
}

// virtSnapshotDeleteCmd represents the snapshot delete command
var virtSnapshotDeleteCmd = &cobra.Command{
	Use:   "delete [vm-name] [snapshot-name]",
	Short: "Delete a snapshot",
	Long: `Delete a snapshot to free up disk space.

WARNING: Once deleted, a snapshot cannot be recovered.
Make sure you no longer need the snapshot before deleting it.

Examples:
  portunix virt snapshot delete ubuntu-test old-snapshot
  portunix virt snapshot delete web-server backup-2023-01 --force`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		snapshotName := args[1]
		force, _ := cmd.Flags().GetBool("force")

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Confirm deletion unless forced
		if !force {
			fmt.Printf("⚠️  Are you sure you want to delete snapshot '%s' from VM '%s'?\n", snapshotName, vmName)
			fmt.Printf("This action cannot be undone. [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Deletion cancelled.")
				return
			}
		}

		fmt.Printf("🗑️  Deleting snapshot '%s' from VM '%s'...\n", snapshotName, vmName)

		if err := manager.DeleteSnapshot(vmName, snapshotName); err != nil {
			fmt.Printf("❌ Failed to delete snapshot: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Snapshot '%s' deleted successfully!\n", snapshotName)
	},
}

// Helper function to format bytes
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "-"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	// Add snapshot commands to virt
	virtCmd.AddCommand(virtSnapshotCmd)
	virtSnapshotCmd.AddCommand(virtSnapshotCreateCmd)
	virtSnapshotCmd.AddCommand(virtSnapshotListCmd)
	virtSnapshotCmd.AddCommand(virtSnapshotRevertCmd)
	virtSnapshotCmd.AddCommand(virtSnapshotDeleteCmd)

	// Snapshot create flags
	virtSnapshotCreateCmd.Flags().String("description", "", "Description for the snapshot")

	// Snapshot revert flags
	virtSnapshotRevertCmd.Flags().Bool("force", false, "Skip confirmation prompts")

	// Snapshot delete flags
	virtSnapshotDeleteCmd.Flags().Bool("force", false, "Skip confirmation prompts")
}