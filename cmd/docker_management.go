package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"portunix.cz/app/docker"

	"github.com/spf13/cobra"
)

// dockerBuildCmd represents the docker build command
var dockerBuildCmd = &cobra.Command{
	Use:   "build [base-image]",
	Short: "Build optimized Portunix Docker images.",
	Long: `Build optimized Portunix Docker images with common tools and configurations.

This command creates Docker images that include Portunix binary and commonly used
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

Features:
- Multi-stage builds for smaller images
- Automatic tagging with version
- Package caching for repeated builds
- Common development tools pre-installed
- Portunix binary embedded in image

Examples:
  portunix docker build                    # Build with ubuntu:22.04
  portunix docker build ubuntu:20.04      # Build with Ubuntu 20.04
  portunix docker build alpine:3.18       # Build lightweight Alpine image
  portunix docker build debian:bullseye   # Build Debian-based image
  portunix docker build rockylinux:9      # Build Rocky Linux-based image`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Docker is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := docker.CheckDockerAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Docker is not available: %v\n", err)
			return
		}

		baseImage := "ubuntu:22.04"
		if len(args) > 0 {
			baseImage = args[0]
		}

		err := docker.BuildImage(baseImage)
		if err != nil {
			fmt.Printf("Error building image: %v\n", err)
			return
		}
	},
}

// dockerListCmd represents the docker list command
var dockerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Portunix Docker containers.",
	Long: `List all Docker containers created by Portunix with their current status,
ports, and other relevant information.

The output includes:
- Container ID (short format)
- Container name
- Base image used
- Current status (running, stopped, exited)
- Port mappings
- Creation timestamp

Examples:
  portunix docker list                     # List all Portunix containers
  
Output format:
  ID       NAME                IMAGE           STATUS      PORTS           CREATED
  abc123   portunix-python     ubuntu:22.04    Running     22:2222         2 hours ago
  def456   portunix-java       alpine:3.18     Stopped     -               1 hour ago`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Docker is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := docker.CheckDockerAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Docker is not available: %v\n", err)
			return
		}

		containers, err := docker.ListContainers()
		if err != nil {
			fmt.Printf("Error listing containers: %v\n", err)
			return
		}

		if len(containers) == 0 {
			fmt.Println("No Portunix containers found.")
			return
		}

		// Create tabwriter for aligned output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tIMAGE\tSTATUS\tPORTS\tCREATED")

		for _, container := range containers {
			// Truncate ID to 8 characters
			shortID := container.ID
			if len(shortID) > 8 {
				shortID = shortID[:8]
			}

			// Truncate name to remove portunix- prefix if present
			displayName := container.Name
			if strings.HasPrefix(displayName, "portunix-") {
				displayName = strings.TrimPrefix(displayName, "portunix-")
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				shortID,
				displayName,
				container.Image,
				container.Status,
				container.Ports,
				container.Created)
		}

		w.Flush()
	},
}

// dockerStartCmd represents the docker start command
var dockerStartCmd = &cobra.Command{
	Use:   "start <container-id>",
	Short: "Start a stopped Portunix container.",
	Long: `Start a previously stopped Portunix container by its ID or name.

The container will resume from its previous state with all data and
configurations preserved.

Examples:
  portunix docker start abc123            # Start by container ID
  portunix docker start portunix-python   # Start by container name`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Docker is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := docker.CheckDockerAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Docker is not available: %v\n", err)
			return
		}

		containerID := args[0]

		err := docker.StartContainer(containerID)
		if err != nil {
			fmt.Printf("Error starting container: %v\n", err)
			return
		}
	},
}

// dockerStopCmd represents the docker stop command
var dockerStopCmd = &cobra.Command{
	Use:   "stop <container-id>",
	Short: "Stop a running Portunix container.",
	Long: `Stop a running Portunix container gracefully. The container state
and data will be preserved and can be restarted later.

Examples:
  portunix docker stop abc123             # Stop by container ID
  portunix docker stop portunix-python    # Stop by container name`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Docker is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := docker.CheckDockerAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Docker is not available: %v\n", err)
			return
		}

		containerID := args[0]

		err := docker.StopContainer(containerID)
		if err != nil {
			fmt.Printf("Error stopping container: %v\n", err)
			return
		}
	},
}

// dockerRemoveCmd represents the docker remove command
var dockerRemoveCmd = &cobra.Command{
	Use:   "remove <container-id>",
	Short: "Remove a Portunix container.",
	Long: `Remove a Portunix container permanently. This will delete the container
and all its data. The operation cannot be undone.

Use the --force flag to remove running containers.

Examples:
  portunix docker remove abc123           # Remove stopped container
  portunix docker remove abc123 --force   # Force remove running container`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Docker is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := docker.CheckDockerAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Docker is not available: %v\n", err)
			return
		}

		containerID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		err := docker.RemoveContainer(containerID, force)
		if err != nil {
			fmt.Printf("Error removing container: %v\n", err)
			return
		}
	},
}

// dockerLogsCmd represents the docker logs command
var dockerLogsCmd = &cobra.Command{
	Use:   "logs <container-id>",
	Short: "View logs from a Portunix container.",
	Long: `Display logs from a Portunix container. This shows the output from
processes running inside the container.

Use the --follow flag to continuously stream new log entries.

Examples:
  portunix docker logs abc123             # Show container logs
  portunix docker logs abc123 --follow    # Follow logs in real-time`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Docker is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := docker.CheckDockerAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Docker is not available: %v\n", err)
			return
		}

		containerID := args[0]
		follow, _ := cmd.Flags().GetBool("follow")

		err := docker.ShowLogs(containerID, follow)
		if err != nil {
			fmt.Printf("Error showing logs: %v\n", err)
			return
		}
	},
}

// dockerExecCmd represents the docker exec command
var dockerExecCmd = &cobra.Command{
	Use:   "exec <container-id> <command>",
	Short: "Execute a command in a running Portunix container.",
	Long: `Execute a command inside a running Portunix container. This opens
an interactive terminal session in the container.

The command runs with the same environment and permissions as processes
inside the container.

Examples:
  portunix docker exec abc123 bash        # Open bash shell
  portunix docker exec abc123 "ls -la"    # Run specific command
  portunix docker exec abc123 python3     # Start Python interpreter`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Docker is available
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		if err := docker.CheckDockerAvailableWithInstall(autoInstall); err != nil {
			fmt.Printf("Docker is not available: %v\n", err)
			return
		}

		containerID := args[0]
		command := args[1:]

		err := docker.ExecCommand(containerID, command)
		if err != nil {
			fmt.Printf("Error executing command: %v\n", err)
			return
		}
	},
}

func init() {
	dockerCmd.AddCommand(dockerBuildCmd)
	dockerCmd.AddCommand(dockerListCmd)
	dockerCmd.AddCommand(dockerStartCmd)
	dockerCmd.AddCommand(dockerStopCmd)
	dockerCmd.AddCommand(dockerRemoveCmd)
	dockerCmd.AddCommand(dockerLogsCmd)
	dockerCmd.AddCommand(dockerExecCmd)

	// Add flags
	dockerBuildCmd.Flags().BoolP("auto-install", "y", false, "Automatically install Docker if not available")
	dockerListCmd.Flags().BoolP("auto-install", "y", false, "Automatically install Docker if not available")
	dockerStartCmd.Flags().BoolP("auto-install", "y", false, "Automatically install Docker if not available")
	dockerStopCmd.Flags().BoolP("auto-install", "y", false, "Automatically install Docker if not available")
	dockerRemoveCmd.Flags().BoolP("force", "f", false, "Force remove running container")
	dockerRemoveCmd.Flags().BoolP("auto-install", "y", false, "Automatically install Docker if not available")
	dockerLogsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	dockerLogsCmd.Flags().BoolP("auto-install", "y", false, "Automatically install Docker if not available")
	dockerExecCmd.Flags().BoolP("auto-install", "y", false, "Automatically install Docker if not available")
}
