package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// vmStartCmd represents the vm start command
var vmStartCmd = &cobra.Command{
	Use:   "start [vm-name]",
	Short: "Start a virtual machine",
	Long:  `Start a virtual machine that has been previously created.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		console, _ := cmd.Flags().GetBool("console")
		
		fmt.Printf("Starting VM '%s'...\n", vmName)
		
		// Try virsh first
		if err := startLibvirtVM(vmName, console); err != nil {
			// Fallback to direct QEMU execution
			if err := startQemuVM(vmName); err != nil {
				fmt.Printf("Error: %v\n", fmt.Errorf("failed to start VM: %w", err)); os.Exit(1)
			}
		}
		
		fmt.Printf("\n✅ VM '%s' started successfully!\n", vmName)
		
		if !console {
			fmt.Println("\nTo connect to the VM console:")
			fmt.Printf("  portunix vm console %s\n", vmName)
		}
	},
}

// vmStopCmd represents the vm stop command
var vmStopCmd = &cobra.Command{
	Use:   "stop [vm-name]",
	Short: "Stop a virtual machine",
	Long:  `Stop a running virtual machine gracefully.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		force, _ := cmd.Flags().GetBool("force")
		
		fmt.Printf("Stopping VM '%s'...\n", vmName)
		
		// Try virsh shutdown
		if err := stopLibvirtVM(vmName, force); err != nil {
			fmt.Printf("Warning: Failed to stop VM via libvirt: %v\n", err)
			// For QEMU direct execution, we'd need to track PIDs
			fmt.Println("If the VM was started directly with QEMU, you may need to close the QEMU window manually.")
		} else {
			fmt.Printf("\n✅ VM '%s' stopped successfully!\n", vmName)
		}
	},
}

// vmRestartCmd represents the vm restart command
var vmRestartCmd = &cobra.Command{
	Use:   "restart [vm-name]",
	Short: "Restart a virtual machine",
	Long:  `Restart a running virtual machine.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		
		fmt.Printf("Restarting VM '%s'...\n", vmName)
		
		// Try virsh reboot
		restartCmd := exec.Command("virsh", "reboot", vmName)
		_, err := restartCmd.CombinedOutput()
		if err != nil {
			// Try stop and start
			fmt.Println("Attempting stop and start sequence...")
			stopLibvirtVM(vmName, false)
			// Wait a moment
			exec.Command("sleep", "2").Run()
			if err := startLibvirtVM(vmName, false); err != nil {
				fmt.Printf("Error: %v\n", fmt.Errorf("failed to restart VM: %w", err)); os.Exit(1)
			}
		}
		
		fmt.Printf("\n✅ VM '%s' restarted successfully!\n", vmName)
	},
}

// vmListCmd represents the vm list command
var vmListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all virtual machines",
	Long:  `List all virtual machines and their current status.`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		
		fmt.Println("Virtual Machines:")
		fmt.Println("=================")
		
		// Try virsh list
		listArgs := []string{"list"}
		if all {
			listArgs = append(listArgs, "--all")
		}
		
		listCmd := exec.Command("virsh", listArgs...)
		output, err := listCmd.Output()
		if err == nil {
			fmt.Print(string(output))
		}
		
		// Also check local VMs directory
		vmDir := filepath.Join(os.Getenv("HOME"), "VMs")
		if entries, err := os.ReadDir(vmDir); err == nil {
			fmt.Println("\nLocal VM directories:")
			fmt.Println("--------------------")
			for _, entry := range entries {
				if entry.IsDir() {
					diskPath := filepath.Join(vmDir, entry.Name(), entry.Name()+".qcow2")
					if _, err := os.Stat(diskPath); err == nil {
						fmt.Printf("  - %s (disk: %s)\n", entry.Name(), diskPath)
					}
				}
			}
		}
	},
}

// vmInfoCmd represents the vm info command
var vmInfoCmd = &cobra.Command{
	Use:   "info [vm-name]",
	Short: "Show detailed information about a VM",
	Long:  `Display detailed information about a virtual machine including configuration and resource usage.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		
		fmt.Printf("VM Information: %s\n", vmName)
		fmt.Println("==================")
		
		// Try virsh dominfo
		infoCmd := exec.Command("virsh", "dominfo", vmName)
		output, err := infoCmd.Output()
		if err == nil {
			fmt.Print(string(output))
			
			// Also show disk info
			fmt.Println("\nDisk Information:")
			fmt.Println("-----------------")
			blkCmd := exec.Command("virsh", "domblklist", vmName)
			if blkOutput, err := blkCmd.Output(); err == nil {
				fmt.Print(string(blkOutput))
			}
			
			// Show snapshot info
			fmt.Println("\nSnapshot Information:")
			fmt.Println("--------------------")
			snapCmd := exec.Command("virsh", "snapshot-list", vmName)
			if snapOutput, err := snapCmd.Output(); err == nil {
				fmt.Print(string(snapOutput))
			}
		} else {
			// Fallback to checking local directory
			vmDir := filepath.Join(os.Getenv("HOME"), "VMs", vmName)
			if stat, err := os.Stat(vmDir); err == nil {
				fmt.Printf("VM Directory: %s\n", vmDir)
				fmt.Printf("Created: %s\n", stat.ModTime().Format("2006-01-02 15:04:05"))
				
				// Check for disk
				diskPath := filepath.Join(vmDir, vmName+".qcow2")
				if diskStat, err := os.Stat(diskPath); err == nil {
					fmt.Printf("\nDisk Image: %s\n", diskPath)
					fmt.Printf("Size: %.2f GB\n", float64(diskStat.Size())/(1024*1024*1024))
					
					// Get qemu-img info
					imgCmd := exec.Command("qemu-img", "info", diskPath)
					if imgOutput, err := imgCmd.Output(); err == nil {
						fmt.Println("\nDisk Details:")
						fmt.Print(string(imgOutput))
					}
				}
				
				// Check for run script
				scriptPath := filepath.Join(vmDir, fmt.Sprintf("run-%s.sh", vmName))
				if _, err := os.Stat(scriptPath); err == nil {
					fmt.Printf("\nRun Script: %s\n", scriptPath)
				}
			} else {
				fmt.Printf("Error: %v\n", fmt.Errorf("VM '%s' not found", vmName)); os.Exit(1)
			}
		}
	},
}

// vmConsoleCmd represents the vm console command
var vmConsoleCmd = &cobra.Command{
	Use:   "console [vm-name]",
	Short: "Connect to VM console",
	Long:  `Connect to the console of a running virtual machine.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		
		fmt.Printf("Connecting to console of VM '%s'...\n", vmName)
		
		// First check if VM is managed by libvirt
		checkCmd := exec.Command("virsh", "domstate", vmName)
		if output, err := checkCmd.Output(); err == nil && strings.TrimSpace(string(output)) == "running" {
			// VM is managed by libvirt
			fmt.Println("(Press Ctrl+] to exit console)")
			consoleCmd := exec.Command("virsh", "console", vmName)
			consoleCmd.Stdin = os.Stdin
			consoleCmd.Stdout = os.Stdout
			consoleCmd.Stderr = os.Stderr
			
			if err := consoleCmd.Run(); err != nil {
				fmt.Printf("Error: %v\n", fmt.Errorf("failed to connect to console: %w", err)); os.Exit(1)
			}
			return
		}
		
		// Check if VM is running via QEMU directly (portunix managed)
		// Look for SPICE port in run script
		vmDir := filepath.Join(os.Getenv("HOME"), "VMs", vmName)
		scriptPath := filepath.Join(vmDir, fmt.Sprintf("run-%s.sh", vmName))
		
		if _, err := os.Stat(scriptPath); err == nil {
			// Check if VM is running
			psCmd := exec.Command("bash", "-c", fmt.Sprintf("ps aux | grep -E 'qemu.*-name.*%s' | grep -v grep", vmName))
			if output, err := psCmd.Output(); err == nil && len(output) > 0 {
				// VM is running, check what display protocol is used
				// Check for VNC first (more common)
				vncPort := "5900"
				
				// Try remote-viewer first (supports both VNC and SPICE)
				if _, err := exec.LookPath("remote-viewer"); err == nil {
					fmt.Println("Connecting via remote-viewer...")
					// Try VNC first
					viewerCmd := exec.Command("remote-viewer", fmt.Sprintf("vnc://localhost:%s", vncPort))
					if err := viewerCmd.Start(); err != nil {
						// Try SPICE as fallback
						viewerCmd = exec.Command("remote-viewer", fmt.Sprintf("spice://localhost:%s", vncPort))
						if err := viewerCmd.Start(); err != nil {
							fmt.Printf("Error: %v\n", fmt.Errorf("failed to launch remote-viewer: %w", err)); os.Exit(1)
						}
					}
					fmt.Printf("\n✅ Console viewer launched for VM '%s'\n", vmName)
					return
				}
				
				// Try vncviewer
				if _, err := exec.LookPath("vncviewer"); err == nil {
					fmt.Println("Connecting via VNC...")
					vncCmd := exec.Command("vncviewer", fmt.Sprintf("localhost:%s", vncPort))
					if err := vncCmd.Start(); err != nil {
						fmt.Printf("Error: %v\n", fmt.Errorf("failed to launch vncviewer: %w", err)); os.Exit(1)
					}
					fmt.Printf("\n✅ Console viewer launched for VM '%s'\n", vmName)
					return
				}
				
				// Try spicy as alternative (for SPICE)
				if _, err := exec.LookPath("spicy"); err == nil {
					fmt.Println("Connecting via SPICE (spicy)...")
					spicyCmd := exec.Command("spicy", "-h", "localhost", "-p", vncPort)
					if err := spicyCmd.Start(); err != nil {
						fmt.Printf("Error: %v\n", fmt.Errorf("failed to launch spicy: %w", err)); os.Exit(1)
					}
					fmt.Printf("\n✅ Console viewer launched for VM '%s'\n", vmName)
					return
				}
				
				// No viewer available
				fmt.Printf("Error: No display viewer found. Install one of:\n")
				fmt.Println("  - remote-viewer (virt-viewer package) - supports VNC and SPICE")
				fmt.Println("  - vncviewer (tigervnc-viewer or realvnc-vnc-viewer)")
				fmt.Println("  - spicy (spice-client-gtk package) - for SPICE only")
				fmt.Printf("\nManual connection:\n")
				fmt.Printf("  VNC: vnc://localhost:%s or localhost:%s\n", vncPort, vncPort)
				fmt.Printf("  SPICE: spice://localhost:%s\n", vncPort)
				os.Exit(1)
			} else {
				fmt.Printf("Error: VM '%s' is not running. Start it first with: portunix vm start %s\n", vmName, vmName)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Error: VM '%s' not found\n", vmName)
			os.Exit(1)
		}
	},
}

// vmDeleteCmd represents the vm delete command
var vmDeleteCmd = &cobra.Command{
	Use:   "delete [vm-name]",
	Short: "Delete a virtual machine",
	Long:  `Delete a virtual machine and optionally its disk images.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		keepDisk, _ := cmd.Flags().GetBool("keep-disk")
		
		fmt.Printf("⚠️  Warning: This will delete VM '%s'", vmName)
		if !keepDisk {
			fmt.Print(" and all its disk images")
		}
		fmt.Println(".")
		fmt.Print("Continue? [y/N]: ")
		
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Delete cancelled.")
			os.Exit(0)
		}
		
		fmt.Printf("\nDeleting VM '%s'...\n", vmName)
		
		// Try virsh undefine
		undefineArgs := []string{"undefine", vmName}
		if !keepDisk {
			undefineArgs = append(undefineArgs, "--remove-all-storage")
		}
		
		undefineCmd := exec.Command("virsh", undefineArgs...)
		output, err := undefineCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Warning: virsh undefine failed: %s\n", string(output))
		}
		
		// Also remove local directory if exists and not keeping disk
		if !keepDisk {
			vmDir := filepath.Join(os.Getenv("HOME"), "VMs", vmName)
			if _, err := os.Stat(vmDir); err == nil {
				fmt.Printf("Removing VM directory: %s\n", vmDir)
				if err := os.RemoveAll(vmDir); err != nil {
					fmt.Printf("Warning: Failed to remove directory: %v\n", err)
				}
			}
		}
		
		fmt.Printf("\n✅ VM '%s' deleted successfully!\n", vmName)
	},
}

// Helper functions
func startLibvirtVM(vmName string, console bool) error {
	startCmd := exec.Command("virsh", "start", vmName)
	output, err := startCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("virsh start failed: %s", string(output))
	}
	
	if console {
		// Connect to console
		consoleCmd := exec.Command("virsh", "console", vmName)
		consoleCmd.Stdin = os.Stdin
		consoleCmd.Stdout = os.Stdout
		consoleCmd.Stderr = os.Stderr
		consoleCmd.Run()
	}
	
	return nil
}

func startQemuVM(vmName string) error {
	// Look for run script
	vmDir := filepath.Join(os.Getenv("HOME"), "VMs", vmName)
	scriptPath := filepath.Join(vmDir, fmt.Sprintf("run-%s.sh", vmName))
	
	if _, err := os.Stat(scriptPath); err == nil {
		fmt.Printf("Starting VM using script: %s\n", scriptPath)
		cmd := exec.Command("bash", scriptPath)
		return cmd.Start()
	}
	
	return fmt.Errorf("no run script found for VM '%s'", vmName)
}

func stopLibvirtVM(vmName string, force bool) error {
	var cmd *exec.Cmd
	if force {
		cmd = exec.Command("virsh", "destroy", vmName)
	} else {
		cmd = exec.Command("virsh", "shutdown", vmName)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("virsh stop failed: %s", string(output))
	}
	
	return nil
}

func init() {
	vmStartCmd.Flags().BoolP("console", "c", false, "Connect to console after starting")
	vmStopCmd.Flags().BoolP("force", "f", false, "Force stop (destroy) the VM")
	vmListCmd.Flags().BoolP("all", "a", false, "Show all VMs including stopped ones")
	vmDeleteCmd.Flags().Bool("keep-disk", false, "Keep disk images when deleting VM")
	
	vmCmd.AddCommand(vmStartCmd)
	vmCmd.AddCommand(vmStopCmd)
	vmCmd.AddCommand(vmRestartCmd)
	vmCmd.AddCommand(vmListCmd)
	vmCmd.AddCommand(vmInfoCmd)
	vmCmd.AddCommand(vmConsoleCmd)
	vmCmd.AddCommand(vmDeleteCmd)
}