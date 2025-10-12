package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"portunix.ai/app/virt"
)

// virtCreateCmd represents the virt create command
var virtCreateCmd = &cobra.Command{
	Use:   "create [vm-name]",
	Short: "Create a new virtual machine",
	Long: `Create a new virtual machine using the configured backend.

The command supports various options for customizing the VM:
- Template-based creation for common OS types
- Custom ISO mounting
- Resource allocation (RAM, CPU, disk)
- Network configuration

Examples:
  portunix virt create ubuntu-test --template ubuntu-24.04 --ram 4G --cpus 4
  portunix virt create win11-vm --template windows11 --ram 8G --disk 100G
  portunix virt create custom-vm --iso ~/Downloads/custom.iso --ram 2G`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Get flags
		template, _ := cmd.Flags().GetString("template")
		iso, _ := cmd.Flags().GetString("iso")
		ram, _ := cmd.Flags().GetString("ram")
		cpus, _ := cmd.Flags().GetInt("cpus")
		disk, _ := cmd.Flags().GetString("disk")
		osType, _ := cmd.Flags().GetString("os-type")
		enableSSH, _ := cmd.Flags().GetBool("enable-ssh")

		config := &virt.VMConfig{
			Name:      vmName,
			Template:  template,
			ISO:       iso,
			RAM:       ram,
			CPUs:      cpus,
			DiskSize:  disk,
			OSType:    osType,
			EnableSSH: enableSSH,
			Network: virt.NetworkConfig{
				Mode: "nat",
			},
		}

		// Apply template if specified
		if template != "" {
			tmpl, err := virt.GetTemplate(template)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			// Apply template defaults
			if config.RAM == "" {
				config.RAM = tmpl.RecommendedRAM
			}
			if config.DiskSize == "" {
				config.DiskSize = tmpl.RecommendedDisk
			}
			if config.OSType == "" {
				config.OSType = tmpl.OSVariant
			}
			if config.ISO == "" {
				config.ISO = tmpl.ISO
			}
		}

		// Apply defaults if not set
		if config.RAM == "" {
			config.RAM = "4G"
		}
		if config.CPUs == 0 {
			config.CPUs = 2
		}
		if config.DiskSize == "" {
			config.DiskSize = "40G"
		}
		if config.OSType == "" {
			config.OSType = "generic"
		}

		fmt.Printf("Creating VM '%s' using %s backend...\n", vmName, manager.GetBackend().GetName())
		if err := manager.Create(config); err != nil {
			fmt.Printf("Error creating VM: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ VM '%s' created successfully!\n", vmName)
		fmt.Println("\nNext steps:")
		fmt.Printf("  portunix virt start %s    # Start the VM\n", vmName)
		fmt.Printf("  portunix virt ssh %s      # SSH into VM (after installation)\n", vmName)
		fmt.Printf("  portunix virt info %s     # Show VM details\n", vmName)
	},
}

// virtStartCmd represents the virt start command
var virtStartCmd = &cobra.Command{
	Use:   "start [vm-name]",
	Short: "Start a virtual machine",
	Long:  `Start a virtual machine. If the VM is already running, this command has no effect.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Starting VM '%s'...\n", vmName)
		if err := manager.Start(vmName); err != nil {
			fmt.Printf("Error starting VM: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ VM '%s' started successfully!\n", vmName)
	},
}

// virtStopCmd represents the virt stop command
var virtStopCmd = &cobra.Command{
	Use:   "stop [vm-name]",
	Short: "Stop a virtual machine",
	Long:  `Stop a virtual machine gracefully. Use --force for immediate shutdown.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		force, _ := cmd.Flags().GetBool("force")

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if force {
			fmt.Printf("Force stopping VM '%s'...\n", vmName)
		} else {
			fmt.Printf("Stopping VM '%s' gracefully...\n", vmName)
		}

		if err := manager.Stop(vmName, force); err != nil {
			fmt.Printf("Error stopping VM: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ VM '%s' stopped successfully!\n", vmName)
	},
}

// virtRestartCmd represents the virt restart command
var virtRestartCmd = &cobra.Command{
	Use:   "restart [vm-name]",
	Short: "Restart a virtual machine",
	Long:  `Restart a virtual machine by stopping and starting it.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Restarting VM '%s'...\n", vmName)
		if err := manager.Restart(vmName); err != nil {
			fmt.Printf("Error restarting VM: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ VM '%s' restarted successfully!\n", vmName)
	},
}

// virtSuspendCmd represents the virt suspend command
var virtSuspendCmd = &cobra.Command{
	Use:   "suspend [vm-name]",
	Short: "Suspend a virtual machine",
	Long:  `Suspend a virtual machine to save its current state.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Suspending VM '%s'...\n", vmName)
		if err := manager.GetBackend().Suspend(vmName); err != nil {
			fmt.Printf("Error suspending VM: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ VM '%s' suspended successfully!\n", vmName)
	},
}

// virtResumeCmd represents the virt resume command
var virtResumeCmd = &cobra.Command{
	Use:   "resume [vm-name]",
	Short: "Resume a suspended virtual machine",
	Long:  `Resume a virtual machine from suspended state.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Resuming VM '%s'...\n", vmName)
		if err := manager.GetBackend().Resume(vmName); err != nil {
			fmt.Printf("Error resuming VM: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ VM '%s' resumed successfully!\n", vmName)
	},
}

// virtDeleteCmd represents the virt delete command
var virtDeleteCmd = &cobra.Command{
	Use:   "delete [vm-name]",
	Short: "Delete a virtual machine",
	Long:  `Delete a virtual machine and optionally keep its disk files.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		keepDisk, _ := cmd.Flags().GetBool("keep-disk")
		force, _ := cmd.Flags().GetBool("force")

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Confirm deletion unless forced
		if !force {
			fmt.Printf("Are you sure you want to delete VM '%s'? [y/N]: ", vmName)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Deletion cancelled.")
				return
			}
		}

		if keepDisk {
			fmt.Printf("Deleting VM '%s' (keeping disk)...\n", vmName)
		} else {
			fmt.Printf("Deleting VM '%s' and all its files...\n", vmName)
		}

		if err := manager.GetBackend().Delete(vmName, keepDisk); err != nil {
			fmt.Printf("Error deleting VM: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ VM '%s' deleted successfully!\n", vmName)
	},
}

// virtListCmdFallback represents the virt list command (fallback implementation)
var virtListCmdFallback = &cobra.Command{
	Use:   "list",
	Short: "List all virtual machines",
	Long:  `List all virtual machines managed by the current backend.`,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		vms, err := manager.List()
		if err != nil {
			fmt.Printf("Error listing VMs: %v\n", err)
			os.Exit(1)
		}

		if len(vms) == 0 {
			fmt.Printf("No VMs found. Create one with: portunix virt create <name>\n")
			return
		}

		fmt.Printf("Backend: %s\n\n", manager.GetBackend().GetName())
		fmt.Printf("%-20s %-12s %-8s %-6s %-10s %-15s\n", "NAME", "STATE", "RAM", "CPUS", "DISK", "IP")
		fmt.Printf("%-20s %-12s %-8s %-6s %-10s %-15s\n", "----", "-----", "---", "----", "----", "--")

		for _, vm := range vms {
			ip := vm.IP
			if ip == "" {
				ip = "-"
			}
			fmt.Printf("%-20s %-12s %-8s %-6d %-10s %-15s\n",
				vm.Name, vm.State, vm.RAM, vm.CPUs, vm.DiskSize, ip)
		}
	},
}

// virtListCmd represents the virt list command with helper delegation
var virtListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all virtual machines",
	Long:  `List all virtual machines managed by the current backend.`,
	Aliases: []string{"ls"},
	Run:   virtWithHelperCheck(virtListCmdFallback),
}

// virtInfoCmd represents the virt info command
var virtInfoCmd = &cobra.Command{
	Use:   "info [vm-name]",
	Short: "Show detailed information about a virtual machine",
	Long:  `Show detailed information about a virtual machine including configuration and status.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		info, err := manager.GetInfo(vmName)
		if err != nil {
			fmt.Printf("Error getting VM info: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("VM Information: %s\n", vmName)
		fmt.Printf("================\n\n")
		fmt.Printf("Name:       %s\n", info.Name)
		fmt.Printf("State:      %s\n", info.State)
		fmt.Printf("Backend:    %s\n", info.Backend)
		fmt.Printf("OS Type:    %s\n", info.OSType)
		fmt.Printf("RAM:        %s\n", info.RAM)
		fmt.Printf("CPUs:       %d\n", info.CPUs)
		fmt.Printf("Disk Size:  %s\n", info.DiskSize)
		fmt.Printf("Created:    %s\n", info.CreatedAt.Format("2006-01-02 15:04:05"))

		if info.IP != "" {
			fmt.Printf("IP Address: %s\n", info.IP)
		}

		if !info.LastStarted.IsZero() {
			fmt.Printf("Last Started: %s\n", info.LastStarted.Format("2006-01-02 15:04:05"))
		}

		if info.VNCPort > 0 {
			fmt.Printf("VNC Port:   %d\n", info.VNCPort)
		}
	},
}

// virtStatusCmd represents the virt status command
var virtStatusCmd = &cobra.Command{
	Use:   "status [vm-name]",
	Short: "Show the status of a virtual machine",
	Long:  `Show the current status of a virtual machine.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		state := manager.GetBackend().GetState(vmName)

		simple, _ := cmd.Flags().GetBool("simple")
		if simple {
			fmt.Println(state)
		} else {
			fmt.Printf("VM '%s' is %s\n", vmName, state)
		}
	},
}

func init() {
	// Add all lifecycle commands to virt
	virtCmd.AddCommand(virtCreateCmd)
	virtCmd.AddCommand(virtStartCmd)
	virtCmd.AddCommand(virtStopCmd)
	virtCmd.AddCommand(virtRestartCmd)
	virtCmd.AddCommand(virtSuspendCmd)
	virtCmd.AddCommand(virtResumeCmd)
	virtCmd.AddCommand(virtDeleteCmd)
	virtCmd.AddCommand(virtListCmd)
	virtCmd.AddCommand(virtInfoCmd)
	virtCmd.AddCommand(virtStatusCmd)

	// Create command flags
	virtCreateCmd.Flags().String("template", "", "VM template to use (ubuntu-24.04, windows11, etc.)")
	virtCreateCmd.Flags().String("iso", "", "ISO file to mount")
	virtCreateCmd.Flags().String("ram", "", "RAM allocation (e.g., 4G, 2048M)")
	virtCreateCmd.Flags().Int("cpus", 0, "Number of CPUs")
	virtCreateCmd.Flags().String("disk", "", "Disk size (e.g., 40G, 50000M)")
	virtCreateCmd.Flags().String("os-type", "", "OS type hint for optimization")
	virtCreateCmd.Flags().Bool("enable-ssh", false, "Enable SSH access")

	// Start command flags
	virtStartCmd.Flags().Bool("force", false, "Force restart if already running")

	// Stop command flags
	virtStopCmd.Flags().Bool("force", false, "Force immediate shutdown")

	// Delete command flags
	virtDeleteCmd.Flags().Bool("keep-disk", false, "Keep disk files when deleting")
	virtDeleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")

	// Status command flags
	virtStatusCmd.Flags().Bool("simple", false, "Output only the status value")
}