package cmd

import (
	"fmt"

	"portunix.ai/app/container"
	"portunix.ai/app/docker"
	"portunix.ai/app/podman"

	"github.com/spf13/cobra"
)

// containerCmd represents the universal container command
var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "Universal container management interface (RECOMMENDED)",
	Long: `🐳 PORTUNIX UNIVERSAL CONTAINER MANAGEMENT (Recommended)

The container command provides comprehensive container management capabilities
with automatic runtime selection and enhanced features for development.

🌟 WHY USE PORTUNIX CONTAINERS INSTEAD OF DIRECT DOCKER/PODMAN:
  ✅ Automatic Docker/Podman selection based on availability
  ✅ Integrated SSH server setup for easy container access
  ✅ Persistent cache directory mounting for faster installations
  ✅ Pre-configured development environments (Python, Java, Go, VS Code)
  ✅ Universal command interface across Windows/Linux platforms
  ✅ Simplified container lifecycle management
  ✅ Intelligent package manager detection and configuration

💡 RECOMMENDATION: Use 'portunix container' instead of direct 'docker' or 'podman'
   commands for development environments and package installation testing.

Key features:
  • Intelligent runtime selection (Docker/Podman based on availability)
  • Multi-platform container support (Ubuntu, Alpine, CentOS, Debian)
  • SSH server automatically configured in containers
  • Cache directory mounting for persistent downloads
  • Flexible base image selection
  • Rootless container support with Podman

Configuration priority:
  1. ./portunix-config.yaml (project directory)
  2. ~/.portunix/config.yaml (user directory) 
  3. ~/.config/portunix/config.yaml (Linux standard)
  4. Built-in defaults (automatic detection)

Example config file:
  container_runtime: auto  # auto, docker, or podman
  verbose: false
  auto_update: true

Quick Start Examples:
  portunix container run-in-container default        # Full dev environment
  portunix container run-in-container python         # Python development
  portunix container run-in-container nodejs         # Node.js environment
  portunix container exec my-container bash          # Access container
  portunix container list                            # Show containers`,
}

// containerRunInContainerCmd represents the universal run-in-container command
var containerRunInContainerCmd = &cobra.Command{
	Use:                "run-in-container [installation-type] [args...]",
	Short:              "Run installation in container (RECOMMENDED for testing)",
	DisableFlagParsing: true, // Allow container arguments to be passed through
	Args:               cobra.MinimumNArgs(1),
	Long: `🚀 RUN PORTUNIX INSTALLATION IN CONTAINER (Recommended for Testing)

This command creates an isolated container environment for testing software installations
without affecting your host system. It automatically selects Docker or Podman based on
availability and provides a complete development environment.

🌟 BENEFITS:
  ✅ Isolated testing environment - protects your host system
  ✅ Automatic runtime selection (Docker/Podman)
  ✅ Pre-configured SSH access for easy container management
  ✅ Persistent cache mounting for faster reinstalls
  ✅ Clean, reproducible environments for each test

💡 BEST PRACTICE: Always test package installations in containers before
   deploying to production or development machines.

Supported installation types:
  • empty: Create clean container without software installation
  • default: Install Python, Java 17, VS Code (recommended)
  • minimal: Install Python only (lightweight)
  • full: Install Python, Java 17, VS Code, Go (complete environment)
  • python: Install Python development environment
  • java: Install Java (default LTS version)
  • nodejs: Install Node.js and npm
  • vscode: Install Visual Studio Code
  • go: Install Go development environment
  • claude-code: Install Claude Code CLI

Examples:
  portunix container run-in-container default
  portunix container run-in-container nodejs --image ubuntu:22.04
  portunix container run-in-container python --image alpine:latest
  portunix container run-in-container claude-code --keep-running

Additional parameters (all Docker/Podman flags supported):
  --image: Specify base image (default: ubuntu:latest)
  --name: Custom container name
  --keep-running: Keep container running after installation
  -v, --volume: Mount additional volumes
  -p, --port: Map container ports to host
  -e, --env: Set environment variables`,
	Run: func(cmd *cobra.Command, args []string) {
		runtime, err := container.GetSelectedRuntime()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			fmt.Println("\n💡 Hint: Check available runtimes with 'portunix container info'")
			return
		}

		fmt.Printf("Using container runtime: %s\n", runtime)

		// Delegate to appropriate runtime implementation
		switch runtime {
		case "docker":
			runDockerContainer(args)
		case "podman":
			runPodmanContainer(args)
		default:
			fmt.Printf("❌ Error: Unsupported runtime: %s\n", runtime)
		}
	},
}

// containerExecCmd represents the universal exec command
var containerExecCmd = &cobra.Command{
	Use:                "exec [flags] <container-name> <command>",
	Short:              "Execute command in container (universal runtime)",
	DisableFlagParsing: true, // Allow command arguments with dashes to be passed through
	Long: `🔧 EXECUTE COMMAND IN CONTAINER

Execute a command in a running container using the automatically selected runtime.
This provides a universal interface that works with both Docker and Podman containers.

🌟 ADVANTAGES:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Consistent command interface
  ✅ Full support for interactive sessions

Examples:
  portunix container exec test-container "ls -la /app/"
  portunix container exec -it web-server bash
  portunix container exec db-container "cat /etc/hosts"
  portunix container exec --interactive test-container sh
  portunix container exec nodejs-dev "node --version"
  portunix container exec python-env "pip list"
  portunix container exec nodejs-dev sh -c "node --version"

The command preserves all arguments and supports interactive mode for shell access.
Use -i or --interactive for interactive sessions.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Manual flag parsing since DisableFlagParsing is true
		var interactive bool
		var containerName string
		var command []string

		// Parse arguments manually to extract flags
		filteredArgs := []string{}
		for _, arg := range args {
			if arg == "-i" || arg == "--interactive" {
				interactive = true
			} else if arg == "-it" {
				interactive = true
				filteredArgs = append(filteredArgs, "-t") // Keep -t for docker/podman
			} else {
				filteredArgs = append(filteredArgs, arg)
			}
		}

		if len(filteredArgs) < 2 {
			fmt.Printf("❌ Error: Usage: portunix container exec [flags] <container-name> <command>\n")
			return
		}

		containerName = filteredArgs[0]
		command = filteredArgs[1:]

		// Get configured runtime
		runtime, err := container.GetSelectedRuntime()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			fmt.Println("\n💡 Hint: Check available runtimes with 'portunix container info'")
			return
		}

		// Silent execution - only show command output

		// Delegate to appropriate runtime implementation
		switch runtime {
		case "docker":
			err = docker.ExecCommandWithOptions(containerName, command, interactive)
		case "podman":
			err = podman.ExecCommandWithOptions(containerName, command, interactive)
		default:
			fmt.Printf("❌ Error: Unsupported runtime: %s\n", runtime)
			return
		}

		if err != nil {
			fmt.Printf("❌ Error executing command: %v\n", err)
		}
	},
}

// containerInfoCmd shows information about container runtimes
var containerInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show container runtime information and availability",
	Long: `📊 CONTAINER RUNTIME INFORMATION

Display information about the selected container runtime and available runtimes.
This helps you understand which runtime Portunix will use for container operations.

🌟 KEY INFORMATION:
  • Shows currently selected runtime (Docker/Podman/auto)
  • Displays availability of each runtime on your system
  • Provides configuration instructions

Use this command to troubleshoot container runtime issues or verify installation.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🐳 Container Runtime Information")
		fmt.Println("===============================")

		// Show current configuration
		runtime, err := container.GetSelectedRuntime()
		if err != nil {
			fmt.Printf("❌ Selected runtime: %v\n", err)
		} else {
			fmt.Printf("✅ Selected runtime: %s\n", runtime)
		}

		fmt.Println()

		// Show availability of all runtimes
		info, err := container.GetRuntimeInfo()
		if err != nil {
			fmt.Printf("❌ Error getting runtime info: %v\n", err)
			return
		}

		fmt.Println("Runtime Availability:")
		for runtimeName, available := range info {
			status := "❌ Not available"
			if available {
				status = "✅ Available"
			}
			fmt.Printf("  %s: %s\n", runtimeName, status)
		}

		fmt.Println("\n💡 Configuration:")
		fmt.Println("  Set runtime: portunix config set container_runtime docker")
		fmt.Println("  Get runtime: portunix config get container_runtime")
	},
}

// containerCheckCmd shows detailed container runtime capabilities
var containerCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check container runtime capabilities and versions",
	Long: `🔍 CHECK CONTAINER RUNTIME CAPABILITIES

Detect and display detailed information about available container runtimes.

🌟 DETECTION INCLUDES:
  • Installed container runtimes (Docker/Podman)
  • Runtime versions and build information
  • Supported features and capabilities
  • System compatibility status
  • Recommendations for optimal setup

Use --refresh flag to force re-detection of capabilities.

This command helps diagnose container runtime issues and verify proper installation.`,
	Run: func(cmd *cobra.Command, args []string) {
		refresh, _ := cmd.Flags().GetBool("refresh")

		var caps *container.ContainerCapabilities
		var err error

		if refresh {
			caps, err = container.RefreshContainerCapabilities()
		} else {
			caps, err = container.GetContainerCapabilities()
		}

		if err != nil {
			fmt.Printf("❌ Error: Failed to detect container capabilities: %v\n", err)
			return
		}

		fmt.Print(caps.FormatCapabilities())

		if !caps.Available {
			fmt.Println("\nNo container runtime detected. You can install one using:")
			fmt.Println("  portunix install docker")
			fmt.Println("  portunix install podman")
		}
	},
}

// containerStopCmd represents the universal stop command
var containerStopCmd = &cobra.Command{
	Use:   "stop <container-name>",
	Short: "Stop container (universal runtime)",
	Long: `⏹️ STOP RUNNING CONTAINER

Stop a running container using the automatically selected runtime.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Graceful shutdown handling

Examples:
  portunix container stop test-container
  portunix container stop web-server
  portunix container stop nodejs-dev

The command finds and stops the container regardless of which runtime is hosting it.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]

		// Get configured runtime and delegate
		runtime, err := container.GetSelectedRuntime()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			fmt.Println("\n💡 Hint: Check available runtimes with 'portunix container info'")
			return
		}

		switch runtime {
		case "docker":
			err = docker.StopContainer(containerName)
		case "podman":
			err = podman.StopContainer(containerName)
		default:
			fmt.Printf("❌ Error: Unsupported runtime: %s\n", runtime)
			return
		}

		if err != nil {
			fmt.Printf("❌ Error stopping container: %v\n", err)
		} else {
			fmt.Printf("✅ Container '%s' stopped successfully\n", containerName)
		}
	},
}

// containerStartCmd represents the universal start command
var containerStartCmd = &cobra.Command{
	Use:   "start <container-name>",
	Short: "Start stopped container (universal runtime)",
	Long: `▶️ START STOPPED CONTAINER

Start a previously stopped container using the automatically selected runtime.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Preserves container state and data

Examples:
  portunix container start test-container
  portunix container start web-server
  portunix container start python-dev

The command finds and starts the container regardless of which runtime created it.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]

		// Get configured runtime and delegate
		runtime, err := container.GetSelectedRuntime()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			fmt.Println("\n💡 Hint: Check available runtimes with 'portunix container info'")
			return
		}

		switch runtime {
		case "docker":
			err = docker.StartContainer(containerName)
		case "podman":
			err = podman.StartContainer(containerName)
		default:
			fmt.Printf("❌ Error: Unsupported runtime: %s\n", runtime)
			return
		}

		if err != nil {
			fmt.Printf("❌ Error starting container: %v\n", err)
		} else {
			fmt.Printf("✅ Container '%s' started successfully\n", containerName)
		}
	},
}

// containerRemoveCmd represents the universal remove command
var containerRemoveCmd = &cobra.Command{
	Use:     "rm <container-name>",
	Short:   "Remove container (universal runtime)",
	Aliases: []string{"remove"},
	Long: `🗑️ REMOVE CONTAINER

Remove a container using the automatically selected runtime.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Supports force removal of running containers
  ✅ Docker/Podman compatible 'rm' command

Examples:
  portunix container rm test-container
  portunix container rm nodejs-dev --force
  portunix container rm web-server -f
  portunix container remove legacy-container  # Legacy alias still supported

Use --force flag to remove running containers forcefully.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]
		force, _ := cmd.Flags().GetBool("force")

		// Get configured runtime and delegate
		runtime, err := container.GetSelectedRuntime()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			fmt.Println("\n💡 Hint: Check available runtimes with 'portunix container info'")
			return
		}

		switch runtime {
		case "docker":
			err = docker.RemoveContainer(containerName, force)
		case "podman":
			err = podman.RemoveContainer(containerName, force)
		default:
			fmt.Printf("❌ Error: Unsupported runtime: %s\n", runtime)
			return
		}

		if err != nil {
			fmt.Printf("❌ Error removing container: %v\n", err)
		} else {
			fmt.Printf("✅ Container '%s' removed successfully\n", containerName)
		}
	},
}

// containerLogsCmd represents the universal logs command
var containerLogsCmd = &cobra.Command{
	Use:   "logs <container-name>",
	Short: "Show container logs (universal runtime)",
	Long: `📝 VIEW CONTAINER LOGS

Display logs from a container using the automatically selected runtime.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Real-time log streaming with --follow
  ✅ Consistent output format

Examples:
  portunix container logs test-container
  portunix container logs web-server --follow
  portunix container logs python-dev
  portunix container logs db-container -f

Use --follow flag to stream logs continuously.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]
		follow, _ := cmd.Flags().GetBool("follow")

		// Get configured runtime and delegate
		runtime, err := container.GetSelectedRuntime()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			fmt.Println("\n💡 Hint: Check available runtimes with 'portunix container info'")
			return
		}

		switch runtime {
		case "docker":
			err = docker.ShowLogs(containerName, follow)
		case "podman":
			err = podman.ShowLogs(containerName, follow)
		default:
			fmt.Printf("❌ Error: Unsupported runtime: %s\n", runtime)
			return
		}

		if err != nil {
			fmt.Printf("❌ Error showing logs: %v\n", err)
		}
	},
}

// containerRunCmd represents the universal run command
var containerRunCmd = &cobra.Command{
	Use:   "run [flags] <image> [command...]",
	Short: "Run new container (universal runtime)",
	Long: `🏃 RUN NEW CONTAINER

Create and start a new container using the automatically selected runtime.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman
  ✅ Automatic runtime selection
  ✅ Full compatibility with Docker/Podman flags
  ✅ Interactive and background modes supported

Examples:
  portunix container run ubuntu:22.04 echo "Hello World"
  portunix container run -d --name test-container ubuntu:22.04 bash
  portunix container run -it --name interactive-container ubuntu:22.04 bash
  portunix container run -d -p 8080:80 nginx:latest
  portunix container run -d --name test ubuntu:22.04 -- bash -c "echo test"

Supported flags:
  -d, --detach: Run container in background
  -i, --interactive: Keep STDIN open
  -t, --tty: Allocate pseudo-TTY
  --name: Assign a name to the container
  -p, --port: Publish container ports to host
  -v, --volume: Bind mount volumes
  -e, --env: Set environment variables

💡 TIP: For development environments, use 'run-in-container' instead.
Use -- to separate flags from command arguments when needed.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		detach, _ := cmd.Flags().GetBool("detach")
		interactive, _ := cmd.Flags().GetBool("interactive")
		tty, _ := cmd.Flags().GetBool("tty")
		name, _ := cmd.Flags().GetString("name")
		ports, _ := cmd.Flags().GetStringSlice("port")
		volumes, _ := cmd.Flags().GetStringSlice("volume")
		envVars, _ := cmd.Flags().GetStringSlice("env")

		image := args[0]
		command := args[1:]

		// Get configured runtime and delegate
		runtime, err := container.GetSelectedRuntime()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			fmt.Println("\n💡 Hint: Check available runtimes with 'portunix container info'")
			return
		}

		switch runtime {
		case "docker":
			err = docker.RunContainer(image, command, docker.ContainerRunOptions{
				Detach:      detach,
				Interactive: interactive,
				TTY:         tty,
				Name:        name,
				Ports:       ports,
				Volumes:     volumes,
				Environment: envVars,
			})
		case "podman":
			err = podman.RunContainer(image, command, podman.ContainerRunOptions{
				Detach:      detach,
				Interactive: interactive,
				TTY:         tty,
				Name:        name,
				Ports:       ports,
				Volumes:     volumes,
				Environment: envVars,
			})
		default:
			fmt.Printf("❌ Error: Unsupported runtime: %s\n", runtime)
			return
		}

		if err != nil {
			fmt.Printf("❌ Error running container: %v\n", err)
		}
	},
}

// containerListCmd represents the universal list command
var containerListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List containers from all available runtimes",
	Aliases: []string{"ls", "ps"},
	Long: `📋 LIST CONTAINERS

Display containers managed by Portunix across all available runtimes.

🌟 UNIVERSAL LISTING:
  ✅ Shows containers from both Docker and Podman
  ✅ Unified view of all development environments
  ✅ Automatic runtime detection
  ✅ Consistent output format
  ✅ Aliases 'ls' and 'ps' for convenience

Examples:
  portunix container list        # List all containers
  portunix container ls          # Same as list
  portunix container ps          # Docker-style alias

The listing includes container name, status, runtime, and creation time.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Get configured runtime and delegate
		runtime, err := container.GetSelectedRuntime()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			fmt.Println("\n💡 Hint: Check available runtimes with 'portunix container info'")
			return
		}

		var containers []interface{}

		switch runtime {
		case "docker":
			dockerContainers, err := docker.ListContainers()
			if err != nil {
				fmt.Printf("❌ Error listing containers: %v\n", err)
				return
			}
			for _, c := range dockerContainers {
				containers = append(containers, c)
			}
		case "podman":
			podmanContainers, err := podman.ListContainers()
			if err != nil {
				fmt.Printf("❌ Error listing containers: %v\n", err)
				return
			}
			for _, c := range podmanContainers {
				containers = append(containers, c)
			}
		default:
			fmt.Printf("❌ Error: Unsupported runtime: %s\n", runtime)
			return
		}

		if len(containers) == 0 {
			fmt.Printf("📋 No containers found using %s\n", runtime)
			return
		}

		fmt.Printf("Containers (%s):\n", runtime)
		fmt.Println("=================")

		// Print containers - this will work with both docker.ContainerInfo and podman.ContainerInfo
		for _, container := range containers {
			fmt.Printf("%v\n", container)
		}
	},
}

// containerCpCmd represents the universal copy command
var containerCpCmd = &cobra.Command{
	Use:   "cp <source> <destination>",
	Short: "Copy files/folders between container and host",
	Long: `📁 COPY FILES BETWEEN CONTAINER AND HOST

Copy files or directories to and from containers using the automatically selected runtime.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Bidirectional copy (host ↔ container)
  ✅ Preserves file permissions and ownership
  ✅ Supports recursive directory copying

Examples:
  portunix container cp file.txt container:/path/to/dest
  portunix container cp container:/path/to/file.txt ./local-file.txt
  portunix container cp ./local-dir container:/path/to/dest/
  portunix container cp container:/etc/config ./config-backup

Syntax:
  container:path  - File/directory inside container
  path            - File/directory on host

The command automatically detects direction based on container: prefix.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		destination := args[1]

		// Get configured runtime and delegate
		runtime, err := container.GetSelectedRuntime()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			fmt.Println("\n💡 Hint: Check available runtimes with 'portunix container info'")
			return
		}

		fmt.Printf("Copying using %s: %s → %s\n", runtime, source, destination)

		switch runtime {
		case "docker":
			err = docker.CopyFiles(source, destination)
		case "podman":
			err = podman.CopyFiles(source, destination)
		default:
			fmt.Printf("❌ Error: Unsupported runtime: %s\n", runtime)
			return
		}

		if err != nil {
			fmt.Printf("❌ Error copying files: %v\n", err)
		} else {
			fmt.Printf("✅ Files copied successfully\n")
		}
	},
}

// runDockerContainer delegates to docker run-in-container logic
func runDockerContainer(args []string) {
	// Use existing docker implementation
	// This reuses the logic from docker_run_in_container.go
	if len(args) == 0 {
		fmt.Println("❌ Error: Installation type required")
		return
	}

	installationType := args[0]
	containerArgs := args[1:]

	fmt.Printf("🐳 Running Docker container with installation type: %s\n", installationType)

	// Call docker.RunInContainer with the arguments
	err := docker.RunInContainerWithArgs(installationType, containerArgs)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
}

// runPodmanContainer delegates to podman run-in-container logic
func runPodmanContainer(args []string) {
	// Use existing podman implementation
	// This reuses the logic from podman_run_in_container.go
	if len(args) == 0 {
		fmt.Println("❌ Error: Installation type required")
		return
	}

	installationType := args[0]
	containerArgs := args[1:]

	fmt.Printf("Running Podman container with installation type: %s\n", installationType)

	// Call podman.RunInContainer with the arguments
	err := podman.RunInContainerWithArgs(installationType, containerArgs)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
}

func init() {
	rootCmd.AddCommand(containerCmd)
	containerCmd.AddCommand(containerRunCmd)
	containerCmd.AddCommand(containerRunInContainerCmd)
	containerCmd.AddCommand(containerExecCmd)
	containerCmd.AddCommand(containerCpCmd)
	containerCmd.AddCommand(containerInfoCmd)
	containerCmd.AddCommand(containerCheckCmd)
	containerCmd.AddCommand(containerStopCmd)
	containerCmd.AddCommand(containerStartCmd)
	containerCmd.AddCommand(containerRemoveCmd)
	containerCmd.AddCommand(containerLogsCmd)
	containerCmd.AddCommand(containerListCmd)

	// Add flags to run command
	containerRunCmd.Flags().BoolP("detach", "d", false, "Run container in background")
	containerRunCmd.Flags().BoolP("interactive", "i", false, "Keep STDIN open")
	containerRunCmd.Flags().BoolP("tty", "t", false, "Allocate pseudo-TTY")
	containerRunCmd.Flags().String("name", "", "Assign a name to the container")
	containerRunCmd.Flags().StringSliceP("port", "p", []string{}, "Publish container ports to host")
	containerRunCmd.Flags().StringSliceP("volume", "v", []string{}, "Bind mount volumes")
	containerRunCmd.Flags().StringSliceP("env", "e", []string{}, "Set environment variables")

	// Add interactive flag to exec command
	containerExecCmd.Flags().BoolP("interactive", "i", false, "Keep STDIN open and allocate pseudo-TTY")

	// Add force flag to remove command
	containerRemoveCmd.Flags().BoolP("force", "f", false, "Force removal of running container")

	// Add follow flag to logs command
	containerLogsCmd.Flags().BoolP("follow", "f", false, "Follow log output (stream logs continuously)")

	// Add refresh flag to check command
	containerCheckCmd.Flags().Bool("refresh", false, "Force refresh of capability detection")
}
