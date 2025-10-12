package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"portunix.ai/app/virt"
)

// virtSSHCmd represents the virt ssh command
var virtSSHCmd = &cobra.Command{
	Use:   "ssh [vm-name] [command]",
	Short: "SSH into a virtual machine with smart boot waiting",
	Long: `SSH into a virtual machine with automatic boot waiting and state management.

Features:
- Automatically waits for VM boot if it's starting
- Can auto-start stopped/suspended VMs with --start flag
- Configurable wait timeout for SSH availability
- Supports running specific commands remotely

The command intelligently handles VM states:
- Running: Connects immediately if SSH is ready, otherwise waits
- Stopped/Suspended: Requires --start flag or fails
- Starting: Automatically waits for boot completion

Examples:
  portunix virt ssh ubuntu-test                    # SSH into VM (wait if booting)
  portunix virt ssh ubuntu-test --start            # Start VM if needed, then SSH
  portunix virt ssh ubuntu-test --wait-timeout 60s # Wait up to 60 seconds
  portunix virt ssh ubuntu-test "uname -a"         # Run specific command
  portunix virt ssh ubuntu-test --check            # Just check SSH availability`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		var command string
		if len(args) > 1 {
			command = args[1]
		}

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Parse flags
		autoStart, _ := cmd.Flags().GetBool("start")
		noWait, _ := cmd.Flags().GetBool("no-wait")
		checkOnly, _ := cmd.Flags().GetBool("check")
		waitTimeoutStr, _ := cmd.Flags().GetString("wait-timeout")

		waitTimeout := 30 * time.Second
		if waitTimeoutStr != "" {
			if duration, err := time.ParseDuration(waitTimeoutStr); err == nil {
				waitTimeout = duration
			}
		}

		opts := virt.SSHOptions{
			Command:     command,
			WaitTimeout: waitTimeout,
			NoWait:      noWait,
			AutoStart:   autoStart,
			CheckOnly:   checkOnly,
		}

		// Check current VM state and provide user feedback
		state := manager.GetBackend().GetState(vmName)
		if state == virt.VMStateError || state == virt.VMStateNotFound {
			fmt.Printf("Error: VM '%s' not found\n", vmName)
			os.Exit(1)
		}

		if checkOnly {
			if manager.GetBackend().IsSSHReady(vmName) {
				fmt.Printf("✅ SSH is ready on VM '%s'\n", vmName)
				return
			} else {
				fmt.Printf("❌ SSH is not ready on VM '%s'\n", vmName)
				os.Exit(1)
			}
		}

		// Handle different VM states with user feedback
		switch state {
		case virt.VMStateRunning:
			if manager.GetBackend().IsSSHReady(vmName) {
				fmt.Printf("🔗 Connecting to VM '%s'...\n", vmName)
			} else if !noWait {
				fmt.Printf("⏳ VM '%s' is running, waiting for SSH availability...\n", vmName)
			}

		case virt.VMStateStopped:
			if autoStart {
				fmt.Printf("🚀 VM '%s' is stopped, starting...\n", vmName)
			} else {
				fmt.Printf("❌ VM '%s' is not running (use --start to auto-start)\n", vmName)
				os.Exit(1)
			}

		case virt.VMStateSuspended:
			if autoStart {
				fmt.Printf("▶️ VM '%s' is suspended, resuming...\n", vmName)
			} else {
				fmt.Printf("❌ VM '%s' is suspended (use --start to auto-resume)\n", vmName)
				os.Exit(1)
			}

		case virt.VMStateStarting:
			fmt.Printf("🔄 VM '%s' is starting, waiting for boot completion...\n", vmName)

		case virt.VMStateNotFound:
			fmt.Printf("❌ VM '%s' not found\n", vmName)
			os.Exit(1)

		default:
			fmt.Printf("❌ VM '%s' is in an invalid state: %s\n", vmName, state)
			os.Exit(1)
		}

		// Attempt SSH connection
		if err := manager.Connect(vmName, opts); err != nil {
			fmt.Printf("❌ SSH connection failed: %v\n", err)
			os.Exit(1)
		}
	},
}

// virtCopyCmd represents the virt copy command
var virtCopyCmd = &cobra.Command{
	Use:   "copy [source] [destination]",
	Short: "Copy files between host and virtual machine",
	Long: `Copy files between the host system and a virtual machine via SSH/SCP.

The copy command uses SSH file transfer, so the VM must be running and have SSH access configured.

Syntax:
  host-to-vm:  portunix virt copy ./local-file vm-name:/remote/path
  vm-to-host:  portunix virt copy vm-name:/remote/path ./local-file

Examples:
  portunix virt copy ./app.tar.gz ubuntu-test:/tmp/
  portunix virt copy ubuntu-test:/etc/hosts ./hosts.backup
  portunix virt copy ./portunix ubuntu-test:/usr/local/bin/`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		destination := args[1]

		manager, err := virt.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Determine copy direction and VM name
		var vmName, localPath, remotePath string
		var isHostToVM bool

		if strings.Contains(source, ":") {
			// VM to host
			parts := strings.SplitN(source, ":", 2)
			vmName = parts[0]
			remotePath = parts[1]
			localPath = destination
			isHostToVM = false
		} else if strings.Contains(destination, ":") {
			// Host to VM
			parts := strings.SplitN(destination, ":", 2)
			vmName = parts[0]
			remotePath = parts[1]
			localPath = source
			isHostToVM = true
		} else {
			fmt.Printf("Error: Invalid copy syntax. Use vm-name:/path for VM files\n")
			os.Exit(1)
		}

		// Check VM state
		state := manager.GetBackend().GetState(vmName)
		if state == virt.VMStateError || state == virt.VMStateNotFound {
			fmt.Printf("Error: VM '%s' not found\n", vmName)
			os.Exit(1)
		}

		if state != virt.VMStateRunning {
			fmt.Printf("Error: VM '%s' is not running (state: %s)\n", vmName, state)
			os.Exit(1)
		}

		if !manager.GetBackend().IsSSHReady(vmName) {
			fmt.Printf("Error: SSH is not available on VM '%s'\n", vmName)
			os.Exit(1)
		}

		if isHostToVM {
			fmt.Printf("📤 Copying %s to %s:%s...\n", localPath, vmName, remotePath)
			if err := manager.CopyToVM(vmName, localPath, remotePath); err != nil {
				fmt.Printf("❌ Copy failed: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("📥 Copying %s:%s to %s...\n", vmName, remotePath, localPath)
			if err := manager.CopyFromVM(vmName, remotePath, localPath); err != nil {
				fmt.Printf("❌ Copy failed: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("✅ Copy completed successfully!\n")
	},
}

// virtExecCmd represents the virt exec command
var virtExecCmd = &cobra.Command{
	Use:   "exec [vm-name] [command]",
	Short: "Execute a command in a virtual machine",
	Long: `Execute a command in a virtual machine via SSH.

This is a convenience wrapper around 'virt ssh vm-name command' with additional features:
- Automatic quoting of complex commands
- Better error handling for command execution
- Optional output capture

Examples:
  portunix virt exec ubuntu-test "ls -la /home"
  portunix virt exec ubuntu-test "sudo apt update && sudo apt upgrade -y"
  portunix virt exec win-vm "dir C:\\"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		command := args[1]

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

		if state != virt.VMStateRunning {
			fmt.Printf("Error: VM '%s' is not running (state: %s)\n", vmName, state)
			os.Exit(1)
		}

		if !manager.GetBackend().IsSSHReady(vmName) {
			fmt.Printf("Error: SSH is not available on VM '%s'\n", vmName)
			os.Exit(1)
		}

		fmt.Printf("🔧 Executing command on VM '%s': %s\n", vmName, command)

		opts := virt.SSHOptions{
			Command:     command,
			WaitTimeout: 5 * time.Second,
			NoWait:      true, // Don't wait since we already checked
		}

		if err := manager.Connect(vmName, opts); err != nil {
			fmt.Printf("❌ Command execution failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Add SSH commands to virt
	virtCmd.AddCommand(virtSSHCmd)
	virtCmd.AddCommand(virtCopyCmd)
	virtCmd.AddCommand(virtExecCmd)

	// SSH command flags
	virtSSHCmd.Flags().Bool("start", false, "Automatically start/resume VM if needed")
	virtSSHCmd.Flags().Bool("no-wait", false, "Don't wait for SSH availability")
	virtSSHCmd.Flags().Bool("check", false, "Just check if SSH is ready (don't connect)")
	virtSSHCmd.Flags().String("wait-timeout", "30s", "Maximum time to wait for SSH (e.g., 60s, 2m)")
}