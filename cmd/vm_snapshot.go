package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// vmSnapshotCmd represents the vm snapshot command
var vmSnapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage VM snapshots",
	Long:  `Create, list, revert, and delete VM snapshots for easy state management.`,
}

// vmSnapshotCreateCmd represents the vm snapshot create command
var vmSnapshotCreateCmd = &cobra.Command{
	Use:   "create [vm-name] [snapshot-name]",
	Short: "Create a new snapshot",
	Long: `Create a new snapshot of a virtual machine.
	
This allows you to save the current state of the VM and revert to it later.
Perfect for testing trial software or creating restore points.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		snapshotName := args[1]
		
		description, _ := cmd.Flags().GetString("description")
		if description == "" {
			description = fmt.Sprintf("Snapshot created on %s", time.Now().Format("2006-01-02 15:04:05"))
		}
		
		fmt.Printf("Creating snapshot '%s' for VM '%s'...\n", snapshotName, vmName)
		
		// Try using virsh first (if VM is managed by libvirt)
		if err := createLibvirtSnapshot(vmName, snapshotName, description); err != nil {
			// Fallback to qemu-img snapshot
			fmt.Println("Trying direct qemu-img snapshot...")
			if err := createQemuSnapshot(vmName, snapshotName); err != nil {
				fmt.Printf("Error: %v\n", fmt.Errorf("failed to create snapshot: %w", err)); os.Exit(1)
			}
		}
		
		fmt.Printf("\n✅ Snapshot '%s' created successfully!\n", snapshotName)
		fmt.Printf("\nTo revert to this snapshot:\n  portunix vm snapshot revert %s %s\n", vmName, snapshotName)
	},
}

// vmSnapshotListCmd represents the vm snapshot list command
var vmSnapshotListCmd = &cobra.Command{
	Use:   "list [vm-name]",
	Short: "List all snapshots for a VM",
	Long:  `List all available snapshots for a virtual machine.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		
		fmt.Printf("Snapshots for VM '%s':\n\n", vmName)
		
		// Try virsh first
		if err := listLibvirtSnapshots(vmName); err != nil {
			// Fallback to qemu-img
			if err := listQemuSnapshots(vmName); err != nil {
				fmt.Printf("No snapshots found or error listing snapshots: %v\n", err)
			}
		}
	},
}

// vmSnapshotRevertCmd represents the vm snapshot revert command
var vmSnapshotRevertCmd = &cobra.Command{
	Use:   "revert [vm-name] [snapshot-name]",
	Short: "Revert VM to a snapshot",
	Long: `Revert a virtual machine to a previously created snapshot.
	
Warning: This will discard any changes made since the snapshot was created.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		snapshotName := args[1]
		
		fmt.Printf("⚠️  Warning: Reverting to snapshot '%s' will discard all changes made since the snapshot was created.\n", snapshotName)
		fmt.Print("Continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Revert cancelled.")
			os.Exit(0)
		}
		
		fmt.Printf("\nReverting VM '%s' to snapshot '%s'...\n", vmName, snapshotName)
		
		// Try virsh first
		if err := revertLibvirtSnapshot(vmName, snapshotName); err != nil {
			// Fallback to qemu-img
			if err := revertQemuSnapshot(vmName, snapshotName); err != nil {
				fmt.Printf("Error: %v\n", fmt.Errorf("failed to revert snapshot: %w", err)); os.Exit(1)
			}
		}
		
		fmt.Printf("\n✅ Successfully reverted to snapshot '%s'!\n", snapshotName)
	},
}

// vmSnapshotDeleteCmd represents the vm snapshot delete command
var vmSnapshotDeleteCmd = &cobra.Command{
	Use:   "delete [vm-name] [snapshot-name]",
	Short: "Delete a snapshot",
	Long:  `Delete a snapshot from a virtual machine.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		snapshotName := args[1]
		
		fmt.Printf("⚠️  Warning: This will permanently delete snapshot '%s'.\n", snapshotName)
		fmt.Print("Continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Delete cancelled.")
			os.Exit(0)
		}
		
		fmt.Printf("\nDeleting snapshot '%s' from VM '%s'...\n", snapshotName, vmName)
		
		// Try virsh first
		if err := deleteLibvirtSnapshot(vmName, snapshotName); err != nil {
			// Fallback to qemu-img
			if err := deleteQemuSnapshot(vmName, snapshotName); err != nil {
				fmt.Printf("Error: %v\n", fmt.Errorf("failed to delete snapshot: %w", err)); os.Exit(1)
			}
		}
		
		fmt.Printf("\n✅ Snapshot '%s' deleted successfully!\n", snapshotName)
	},
}

// Libvirt-based snapshot functions
func createLibvirtSnapshot(vmName, snapshotName, description string) error {
	cmd := exec.Command("virsh", "snapshot-create-as", vmName, snapshotName, "--description", description)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("virsh snapshot failed: %s", string(output))
	}
	return nil
}

func listLibvirtSnapshots(vmName string) error {
	cmd := exec.Command("virsh", "snapshot-list", vmName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("virsh snapshot-list failed: %s", string(output))
	}
	fmt.Print(string(output))
	return nil
}

func revertLibvirtSnapshot(vmName, snapshotName string) error {
	// First, ensure the VM is shut down
	shutdownCmd := exec.Command("virsh", "destroy", vmName)
	shutdownCmd.Run() // Ignore error if VM is already shut down
	
	// Revert to snapshot
	cmd := exec.Command("virsh", "snapshot-revert", vmName, snapshotName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("virsh snapshot-revert failed: %s", string(output))
	}
	return nil
}

func deleteLibvirtSnapshot(vmName, snapshotName string) error {
	cmd := exec.Command("virsh", "snapshot-delete", vmName, snapshotName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("virsh snapshot-delete failed: %s", string(output))
	}
	return nil
}

// QEMU-img based snapshot functions (for standalone qcow2 files)
func getVMDiskPath(vmName string) (string, error) {
	// Try to find the disk in the standard location
	vmDir := filepath.Join(os.Getenv("HOME"), "VMs", vmName)
	diskPath := filepath.Join(vmDir, fmt.Sprintf("%s.qcow2", vmName))
	
	if _, err := os.Stat(diskPath); err == nil {
		return diskPath, nil
	}
	
	// Try to get from virsh if available
	cmd := exec.Command("virsh", "domblklist", vmName)
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, ".qcow2") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					return fields[1], nil
				}
			}
		}
	}
	
	return "", fmt.Errorf("could not find disk image for VM '%s'", vmName)
}

func createQemuSnapshot(vmName, snapshotName string) error {
	diskPath, err := getVMDiskPath(vmName)
	if err != nil {
		return err
	}
	
	cmd := exec.Command("qemu-img", "snapshot", "-c", snapshotName, diskPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qemu-img snapshot failed: %s", string(output))
	}
	return nil
}

func listQemuSnapshots(vmName string) error {
	diskPath, err := getVMDiskPath(vmName)
	if err != nil {
		return err
	}
	
	cmd := exec.Command("qemu-img", "snapshot", "-l", diskPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qemu-img snapshot list failed: %s", string(output))
	}
	fmt.Print(string(output))
	return nil
}

func revertQemuSnapshot(vmName, snapshotName string) error {
	diskPath, err := getVMDiskPath(vmName)
	if err != nil {
		return err
	}
	
	cmd := exec.Command("qemu-img", "snapshot", "-a", snapshotName, diskPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qemu-img snapshot revert failed: %s", string(output))
	}
	return nil
}

func deleteQemuSnapshot(vmName, snapshotName string) error {
	diskPath, err := getVMDiskPath(vmName)
	if err != nil {
		return err
	}
	
	cmd := exec.Command("qemu-img", "snapshot", "-d", snapshotName, diskPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qemu-img snapshot delete failed: %s", string(output))
	}
	return nil
}

func init() {
	vmSnapshotCmd.AddCommand(vmSnapshotCreateCmd)
	vmSnapshotCmd.AddCommand(vmSnapshotListCmd)
	vmSnapshotCmd.AddCommand(vmSnapshotRevertCmd)
	vmSnapshotCmd.AddCommand(vmSnapshotDeleteCmd)
	
	vmSnapshotCreateCmd.Flags().StringP("description", "d", "", "Description for the snapshot")
	
	vmCmd.AddCommand(vmSnapshotCmd)
}