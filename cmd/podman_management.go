package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"portunix.cz/app/podman"

	"github.com/spf13/cobra"
)

// podmanBuildCmd represents the podman build command
var podmanBuildCmd = &cobra.Command{
	Use:   "build [base-image]",
	Short: "Build optimized Portunix Podman images.",
	Long: `Build optimized Portunix Podman images with common tools and configurations.

This command creates Podman images that include Portunix binary and commonly used
development tools. The images are optimized for faster container startup and
include proper PATH configuration.

Supported base images:
  ubuntu:22.04    - Ubuntu 22.04 LTS (default, full-featured)
  ubuntu:20.04    - Ubuntu 20.04 LTS (stable)
  alpine:3.18     - Alpine Linux 3.18 (lightweight)
  alpine:latest   - Alpine Linux latest (minimal footprint)
  debian:bullseye - Debian 11 (stable)
  centos:8        - CentOS 8 (enterprise)
  fedora:38       - Fedora 38 (cutting-edge)
  rockylinux:9    - Rocky Linux 9 (RHEL-compatible, CentOS successor)

Podman-specific features:
- Rootless image building by default
- Daemonless build process
- OCI-compatible images (work with Docker too)
- Multi-stage builds for smaller images
- Automatic tagging with version
- Package caching for repeated builds
- Common development tools pre-installed
- Portunix binary embedded in image

Examples:
  portunix podman build                    # Build with ubuntu:22.04
  portunix podman build ubuntu:20.04      # Build with Ubuntu 20.04
  portunix podman build alpine:3.18       # Build lightweight Alpine image
  portunix podman build debian:bullseye   # Build Debian-based image
  portunix podman build rockylinux:9      # Build Rocky Linux-based image`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Podman is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := podman.CheckPodmanAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Podman is not available: %v\n", err)
			return
		}

		baseImage := "ubuntu:22.04"
		if len(args) > 0 {
			baseImage = args[0]
		}

		err := podman.BuildImage(baseImage)
		if err != nil {
			fmt.Printf("Error building image: %v\n", err)
			return
		}
	},
}

// podmanListCmd represents the podman list command
var podmanListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Portunix Podman containers.",
	Long: `List all Podman containers created by Portunix with their current status,
ports, and other relevant information.

The output includes:
- Container ID (short format)
- Container name
- Base image used
- Current status (running, stopped, exited)
- Port mappings
- Creation timestamp

Podman-specific information:
- Shows rootless vs privileged containers
- Displays pod associations if any
- Shows security context information

Examples:
  portunix podman list                     # List all Portunix containers
  
Output format:
  ID       NAME                IMAGE           STATUS      PORTS           CREATED
  abc123   portunix-python     ubuntu:22.04    Running     22:2222         2 hours ago
  def456   portunix-java       alpine:3.18     Stopped     -               1 hour ago`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Podman is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := podman.CheckPodmanAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Podman is not available: %v\n", err)
			return
		}

		containers, err := podman.ListContainers()
		if err != nil {
			fmt.Printf("Error listing containers: %v\n", err)
			return
		}

		if len(containers) == 0 {
			fmt.Println("No Portunix Podman containers found.")
			fmt.Println("Create a container with: portunix podman run-in-container <type>")
			return
		}

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tIMAGE\tSTATUS\tPORTS\tCREATED")
		fmt.Fprintln(w, "------\t----\t-----\t------\t-----\t-------")

		for _, container := range containers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				container.ID[:8], // Short ID
				container.Name,
				container.Image,
				container.Status,
				container.Ports,
				container.Created,
			)
		}

		w.Flush()

		fmt.Printf("\nFound %d Portunix container(s).\n", len(containers))
		fmt.Println("\nManagement commands:")
		fmt.Println("  portunix podman start <container-id>    # Start stopped container")
		fmt.Println("  portunix podman stop <container-id>     # Stop running container")
		fmt.Println("  portunix podman remove <container-id>   # Remove container")
		fmt.Println("  portunix podman logs <container-id>     # View container logs")
		fmt.Println("  portunix podman exec <container-id>     # Execute command in container")
	},
}

// podmanStartCmd represents the podman start command
var podmanStartCmd = &cobra.Command{
	Use:   "start <container-id>",
	Short: "Start a stopped Podman container.",
	Long: `Start a previously stopped Portunix Podman container.

This command will start a container that was created but is currently stopped.
The container will resume with its previous configuration including volume
mounts, port mappings, and environment variables.

Examples:
  portunix podman start abc123              # Start container with ID abc123
  portunix podman start portunix-python     # Start container by name`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Podman is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := podman.CheckPodmanAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Podman is not available: %v\n", err)
			return
		}

		containerID := args[0]
		err := podman.StartContainer(containerID)
		if err != nil {
			fmt.Printf("Error starting container: %v\n", err)
			return
		}
	},
}

// podmanStopCmd represents the podman stop command
var podmanStopCmd = &cobra.Command{
	Use:   "stop <container-id>",
	Short: "Stop a running Podman container.",
	Long: `Stop a currently running Portunix Podman container.

This command gracefully stops a running container. The container can be
started again later with the 'start' command, preserving its state and
configuration.

Examples:
  portunix podman stop abc123              # Stop container with ID abc123
  portunix podman stop portunix-python     # Stop container by name`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Podman is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := podman.CheckPodmanAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Podman is not available: %v\n", err)
			return
		}

		containerID := args[0]
		err := podman.StopContainer(containerID)
		if err != nil {
			fmt.Printf("Error stopping container: %v\n", err)
			return
		}
	},
}

// podmanRemoveCmd represents the podman remove command
var podmanRemoveCmd = &cobra.Command{
	Use:   "remove <container-id>",
	Short: "Remove a Podman container.",
	Long: `Remove a Portunix Podman container permanently.

This command will delete the container and free up disk space. If the
container is currently running, use --force flag to stop and remove it.

WARNING: This action cannot be undone. All data inside the container
(not in mounted volumes) will be lost.

Examples:
  portunix podman remove abc123            # Remove stopped container
  portunix podman remove abc123 --force    # Force remove running container`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Podman is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := podman.CheckPodmanAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Podman is not available: %v\n", err)
			return
		}

		containerID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		err := podman.RemoveContainer(containerID, force)
		if err != nil {
			fmt.Printf("Error removing container: %v\n", err)
			return
		}
	},
}

// podmanLogsCmd represents the podman logs command
var podmanLogsCmd = &cobra.Command{
	Use:   "logs <container-id>",
	Short: "View logs from a Podman container.",
	Long: `Display logs from a Portunix Podman container.

This command shows the console output from the container, which includes
any output from processes running inside the container.

Examples:
  portunix podman logs abc123              # Show logs from container
  portunix podman logs abc123 --follow     # Follow logs in real-time`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Podman is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := podman.CheckPodmanAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Podman is not available: %v\n", err)
			return
		}

		containerID := args[0]
		follow, _ := cmd.Flags().GetBool("follow")

		err := podman.ShowLogs(containerID, follow)
		if err != nil {
			fmt.Printf("Error showing logs: %v\n", err)
			return
		}
	},
}

// podmanExecCmd represents the podman exec command
var podmanExecCmd = &cobra.Command{
	Use:   "exec <container-id> <command> [args...]",
	Short: "Execute a command in a running Podman container.",
	Long: `Execute a command inside a running Portunix Podman container.

This command allows you to run arbitrary commands inside the container
environment, similar to SSH access but without network setup.

Examples:
  portunix podman exec abc123 bash         # Open bash shell
  portunix podman exec abc123 ls -la       # List files
  portunix podman exec abc123 python --version  # Check Python version`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Podman is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := podman.CheckPodmanAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Podman is not available: %v\n", err)
			return
		}

		containerID := args[0]
		command := args[1:]

		err := podman.ExecCommand(containerID, command)
		if err != nil {
			fmt.Printf("Error executing command: %v\n", err)
			return
		}
	},
}

// podmanCheckRequirementsCmd represents the podman check-requirements command
var podmanCheckRequirementsCmd = &cobra.Command{
	Use:   "check-requirements",
	Short: "Check system requirements for Podman container operations.",
	Long: `Check if all system requirements for Podman container operations are satisfied.

This command verifies:
- Podman installation and version
- Podman system accessibility
- Rootless container capability
- Disk space availability
- Directory permissions for volume mounting
- Cache directory setup
- Network connectivity for image pulling

Examples:
  portunix podman check-requirements       # Check all requirements`,
	Run: func(cmd *cobra.Command, args []string) {
		err := podman.CheckRequirements()
		if err != nil {
			fmt.Printf("Requirements check failed: %v\n", err)
			return
		}
	},
}

func init() {
	// Add subcommands to podman
	podmanCmd.AddCommand(podmanBuildCmd)
	podmanCmd.AddCommand(podmanListCmd)
	podmanCmd.AddCommand(podmanStartCmd)
	podmanCmd.AddCommand(podmanStopCmd)
	podmanCmd.AddCommand(podmanRemoveCmd)
	podmanCmd.AddCommand(podmanLogsCmd)
	podmanCmd.AddCommand(podmanExecCmd)
	podmanCmd.AddCommand(podmanCheckRequirementsCmd)

	// Add flags
	podmanBuildCmd.Flags().Bool("auto-install", false, "Automatically install Podman if not available")
	podmanListCmd.Flags().Bool("auto-install", false, "Automatically install Podman if not available")
	podmanStartCmd.Flags().Bool("auto-install", false, "Automatically install Podman if not available")
	podmanStopCmd.Flags().Bool("auto-install", false, "Automatically install Podman if not available")
	podmanRemoveCmd.Flags().Bool("auto-install", false, "Automatically install Podman if not available")
	podmanRemoveCmd.Flags().BoolP("force", "f", false, "Force remove running container")
	podmanLogsCmd.Flags().Bool("auto-install", false, "Automatically install Podman if not available")
	podmanLogsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	podmanExecCmd.Flags().Bool("auto-install", false, "Automatically install Podman if not available")
}
