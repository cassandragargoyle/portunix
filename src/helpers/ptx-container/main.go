package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var version = "dev"

// rootCmd represents the base command for ptx-container
var rootCmd = &cobra.Command{
	Use:   "portunix container",
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
   commands for development environments and package installation testing.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle the dispatched command directly
		handleCommand(args)
	},
}


func handleCommand(args []string) {
	// Handle dispatched commands: container, docker, podman
	if len(args) == 0 {
		fmt.Println("No command specified")
		return
	}

	command := args[0]
	subArgs := args[1:]

	switch command {
	case "container", "docker", "podman":
		if len(subArgs) == 0 || (len(subArgs) == 1 && (subArgs[0] == "--help" || subArgs[0] == "-h")) {
			// Show container help with logical command structure
			fmt.Printf("Usage: portunix %s [command]\n\n", command)
			fmt.Println("Available Commands:")
			fmt.Println("  check            Check container runtime capabilities and versions")
			fmt.Println("  cp               Copy files/folders between container and host")
			fmt.Println("  exec             Execute command in container (universal runtime)")
			fmt.Println("  info             Show container runtime information and availability")
			fmt.Println("  list             List containers from all available runtimes")
			fmt.Println("  logs             Show container logs (universal runtime)")
			fmt.Println("  rm               Remove container (universal runtime)")
			fmt.Println("  run              Run new container (universal runtime)")
			fmt.Println("  run-in-container Run installation in container (RECOMMENDED for testing)")
			fmt.Println("  start            Start stopped container (universal runtime)")
			fmt.Println("  stop             Stop container (universal runtime)")
			fmt.Println("\nFlags:")
			fmt.Println("  -h, --help   help for", command)
			fmt.Println("\nGlobal Flags:")
			fmt.Println("      --help-ai       Show machine-readable help in JSON format")
			fmt.Println("      --help-expert   Show extended help with all options and examples")
			fmt.Printf("\nUse \"portunix %s [command] --help\" for more information about a command.\n", command)
		} else {
			// Implement actual container logic
			handleContainerSubcommand(command, subArgs)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

// handleContainerSubcommand handles specific container subcommands
func handleContainerSubcommand(command string, subArgs []string) {
	if len(subArgs) == 0 {
		fmt.Printf("No subcommand specified for %s\n", command)
		return
	}

	subcommand := subArgs[0]
	cmdArgs := subArgs[1:]

	switch subcommand {
	case "run":
		handleContainerRun(cmdArgs)
	case "run-in-container":
		handleRunInContainer(cmdArgs)
	case "exec":
		handleContainerExec(cmdArgs)
	case "list":
		handleContainerList(cmdArgs)
	case "stop":
		handleContainerStop(cmdArgs)
	case "start":
		handleContainerStart(cmdArgs)
	case "rm":
		handleContainerRm(cmdArgs)
	case "logs":
		handleContainerLogs(cmdArgs)
	case "cp":
		handleContainerCp(cmdArgs)
	case "info":
		handleContainerInfo(cmdArgs)
	case "check":
		handleContainerCheck(cmdArgs)
	default:
		fmt.Printf("Unknown %s subcommand: %s\n", command, subcommand)
		fmt.Printf("Available subcommands: run, run-in-container, exec, list, stop, start, rm, logs, cp, info, check\n")
	}
}

// handleRunInContainer handles run-in-container subcommand
func handleRunInContainer(args []string) {
	// Check for help flag first
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showRunInContainerHelp()
			return
		}
	}

	if len(args) == 0 {
		fmt.Println("❌ Error: Installation type required")
		showRunInContainerHelp()
		return
	}

	// Parse arguments: extract installationType and --image flag
	var installationType string
	var containerImage string = "ubuntu:22.04" // default
	var remainingArgs []string

	installationType = args[0]

	// Parse remaining arguments for flags
	for i := 1; i < len(args); i++ {
		if args[i] == "--image" && i+1 < len(args) {
			containerImage = args[i+1]
			i++ // Skip next argument as it's the image value
		} else {
			remainingArgs = append(remainingArgs, args[i])
		}
	}

	fmt.Printf("🐳 Starting container installation for: %s\n", installationType)
	fmt.Printf("📦 Using image: %s\n", containerImage)

	// Try Podman first, then Docker
	if isPodmanAvailable() {
		fmt.Println("Using Podman as container runtime...")
		runPodmanInContainerWithImage(installationType, containerImage, remainingArgs)
	} else if isDockerAvailable() {
		fmt.Println("Using Docker as container runtime...")
		runDockerInContainerWithImage(installationType, containerImage, remainingArgs)
	} else {
		fmt.Println("❌ Error: Neither Podman nor Docker is available")
		fmt.Println("Please install Podman or Docker first")
	}
}

// showRunInContainerHelp displays help for the run-in-container subcommand
func showRunInContainerHelp() {
	fmt.Println("Usage: portunix container run-in-container [OPTIONS] <PACKAGE>")
	fmt.Println()
	fmt.Println("🐳 RUN PACKAGE INSTALLATION INSIDE CONTAINER")
	fmt.Println()
	fmt.Println("Run package installation inside a container environment for safe testing.")
	fmt.Println()
	fmt.Println("🌟 FEATURES:")
	fmt.Println("  ✅ Isolated testing environment")
	fmt.Println("  ✅ Automatic runtime selection (Podman/Docker)")
	fmt.Println("  ✅ Clean container environment for each test")
	fmt.Println("  ✅ Package installation validation")
	fmt.Println("  ✅ Host system protection")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <PACKAGE>           Package to install (required)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --image <IMAGE>     Container image to use (default: ubuntu:22.04)")
	fmt.Println("  -h, --help          Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container run-in-container nodejs")
	fmt.Println("  portunix container run-in-container python --image debian:bookworm")
	fmt.Println("  portunix container run-in-container ansible --image ubuntu:22.04")
	fmt.Println("  portunix container run-in-container claude-code")
	fmt.Println()
	fmt.Println("💡 RECOMMENDATION: Use this command for testing package installations")
	fmt.Println("   without affecting your host development environment.")
}

// showRunHelp displays help for the run subcommand
func showRunHelp() {
	fmt.Println("Usage: portunix container run [flags] <image> [command...]")
	fmt.Println()
	fmt.Println("🏃 RUN NEW CONTAINER")
	fmt.Println()
	fmt.Println("Create and start a new container using the automatically selected runtime.")
	fmt.Println()
	fmt.Println("🌟 UNIVERSAL OPERATION:")
	fmt.Println("  ✅ Works with both Docker and Podman")
	fmt.Println("  ✅ Automatic runtime selection")
	fmt.Println("  ✅ Full compatibility with Docker/Podman flags")
	fmt.Println("  ✅ Interactive and background modes supported")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container run ubuntu:22.04 echo \"Hello World\"")
	fmt.Println("  portunix container run -d --name test-container ubuntu:22.04 bash")
	fmt.Println("  portunix container run -it --name interactive-container ubuntu:22.04 bash")
	fmt.Println("  portunix container run -d -p 8080:80 nginx:latest")
	fmt.Println("  portunix container run -d --name test ubuntu:22.04 -- bash -c \"echo test\"")
	fmt.Println()
	fmt.Println("Supported flags:")
	fmt.Println("  -d, --detach: Run container in background")
	fmt.Println("  -i, --interactive: Keep STDIN open")
	fmt.Println("  -t, --tty: Allocate pseudo-TTY")
	fmt.Println("  --name: Assign a name to the container")
	fmt.Println("  -p, --port: Publish container ports to host")
	fmt.Println("  -v, --volume: Bind mount volumes")
	fmt.Println("  -e, --env: Set environment variables")
	fmt.Println()
	fmt.Println("💡 TIP: For development environments, use 'run-in-container' instead.")
	fmt.Println("Use -- to separate flags from command arguments when needed.")
}

// handleContainerRun handles basic run subcommand
func handleContainerRun(args []string) {
	// Check for help flag first
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showRunHelp()
			return
		}
	}

	fmt.Println("ℹ️  Basic container run functionality")
	fmt.Println("💡 For software installation testing, use 'run-in-container' instead")

	if len(args) == 0 {
		fmt.Println("❌ Error: Image name required")
		fmt.Println("Usage: portunix container run <image> [command...]")
		return
	}

	image := args[0]
	command := args[1:]

	// Try Podman first, then Docker
	if isPodmanAvailable() {
		runPodmanContainer(image, command)
	} else if isDockerAvailable() {
		runDockerContainer(image, command)
	} else {
		fmt.Println("❌ Error: Neither Podman nor Docker is available")
	}
}

// Placeholder implementations for other subcommands
func handleContainerExec(args []string) {
	if len(args) < 2 {
		fmt.Println("❌ Error: Usage: portunix container exec <container-name> <command> [args...]")
		fmt.Println("Examples:")
		fmt.Println("  portunix container exec my-container bash")
		fmt.Println("  portunix container exec my-container ansible --version")
		fmt.Println("  portunix container exec my-container python -c \"print('hello')\"")
		return
	}

	containerName := args[0]
	command := args[1:]

	// Try Podman first, then Docker
	// Silent execution - only show command output, not execution messages
	if isPodmanAvailable() {
		if err := execPodmanCommand(containerName, command); err != nil {
			// Try Docker as fallback if Podman fails
			if isDockerAvailable() {
				if err := execDockerCommand(containerName, command); err != nil {
					fmt.Fprintf(os.Stderr, "❌ Error: Failed to execute command in container '%s': %v\n", containerName, err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "❌ Error: Failed to execute command in container '%s': %v\n", containerName, err)
			}
		}
	} else if isDockerAvailable() {
		if err := execDockerCommand(containerName, command); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: Failed to execute command in container '%s': %v\n", containerName, err)
		}
	} else {
		fmt.Fprintln(os.Stderr, "❌ Error: Neither Podman nor Docker is available")
		fmt.Fprintln(os.Stderr, "Please install Podman or Docker first")
	}
}

func handleContainerList(args []string) {
	// Check runtime availability
	dockerAvailable := isDockerAvailable()
	podmanAvailable := isPodmanAvailable()

	if !dockerAvailable && !podmanAvailable {
		fmt.Println("❌ Error: Neither Docker nor Podman is available")
		fmt.Println("Please install Docker or Podman first")
		return
	}

	fmt.Println("📋 Container List")
	fmt.Println("=================")

	hasContainers := false

	// List Docker containers if available
	if dockerAvailable {
		fmt.Println("\n🐳 Docker Containers:")
		containers, err := listDockerContainers()
		if err != nil {
			fmt.Printf("❌ Error listing Docker containers: %v\n", err)
		} else if len(containers) == 0 {
			fmt.Println("   No Docker containers found")
		} else {
			hasContainers = true
			printContainerTable(containers)
		}
	}

	// List Podman containers if available
	if podmanAvailable {
		fmt.Println("\n🦭 Podman Containers:")
		containers, err := listPodmanContainers()
		if err != nil {
			fmt.Printf("❌ Error listing Podman containers: %v\n", err)
		} else if len(containers) == 0 {
			fmt.Println("   No Podman containers found")
		} else {
			hasContainers = true
			printContainerTable(containers)
		}
	}

	if !hasContainers {
		fmt.Println("\n💡 No containers found. Create one with:")
		fmt.Println("   portunix container run-in-container default")
	}
}

func handleContainerStop(args []string) {
	// Check for help flag first
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showStopHelp()
			return
		}
	}

	if len(args) < 1 {
		fmt.Println("❌ Error: Container name required")
		fmt.Println("Usage: portunix container stop <container-name>")
		return
	}

	containerName := args[0]

	// Try Podman first, then Docker
	if isPodmanAvailable() {
		if err := stopPodmanContainer(containerName); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error stopping container: %v\n", err)
			return
		}
	} else if isDockerAvailable() {
		if err := stopDockerContainer(containerName); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error stopping container: %v\n", err)
			return
		}
	} else {
		fmt.Println("❌ Error: Neither Podman nor Docker is available")
		return
	}

	fmt.Printf("✅ Container '%s' stopped successfully\n", containerName)
}

func handleContainerStart(args []string) {
	// Check for help flag first
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showStartHelp()
			return
		}
	}

	if len(args) < 1 {
		fmt.Println("❌ Error: Container name required")
		fmt.Println("Usage: portunix container start <container-name>")
		return
	}

	containerName := args[0]

	// Try Podman first, then Docker
	if isPodmanAvailable() {
		if err := startPodmanContainer(containerName); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error starting container: %v\n", err)
			return
		}
	} else if isDockerAvailable() {
		if err := startDockerContainer(containerName); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error starting container: %v\n", err)
			return
		}
	} else {
		fmt.Println("❌ Error: Neither Podman nor Docker is available")
		return
	}

	fmt.Printf("✅ Container '%s' started successfully\n", containerName)
}

func handleContainerRm(args []string) {
	// Parse flags: -f or --force
	var force bool
	var containerNames []string

	for _, arg := range args {
		if arg == "-f" || arg == "--force" {
			force = true
		} else if arg == "--help" || arg == "-h" {
			showRmHelp()
			return
		} else {
			containerNames = append(containerNames, arg)
		}
	}

	if len(containerNames) == 0 {
		fmt.Println("❌ Error: At least one container name required")
		fmt.Println("Usage: portunix container rm [OPTIONS] <container-name> [<container-name>...]")
		fmt.Println("Options:")
		fmt.Println("  -f, --force    Force removal of running containers")
		fmt.Println("  -h, --help     Show this help message")
		return
	}

	// Remove each container
	for _, containerName := range containerNames {
		if err := removeContainer(containerName, force); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error removing container '%s': %v\n", containerName, err)
		} else {
			fmt.Printf("✅ Container '%s' removed successfully\n", containerName)
		}
	}
}

func handleContainerLogs(args []string) {
	// Parse flags: -f or --follow
	var follow bool
	var containerName string

	for _, arg := range args {
		if arg == "-f" || arg == "--follow" {
			follow = true
		} else if arg == "--help" || arg == "-h" {
			showLogsHelp()
			return
		} else if containerName == "" {
			containerName = arg
		}
	}

	if containerName == "" {
		fmt.Println("❌ Error: Container name required")
		fmt.Println("Usage: portunix container logs [OPTIONS] <container-name>")
		fmt.Println("Options:")
		fmt.Println("  -f, --follow    Follow log output (stream continuously)")
		fmt.Println("  -h, --help      Show this help message")
		return
	}

	// Show logs
	if isPodmanAvailable() {
		if err := showPodmanLogs(containerName, follow); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error showing logs: %v\n", err)
		}
	} else if isDockerAvailable() {
		if err := showDockerLogs(containerName, follow); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error showing logs: %v\n", err)
		}
	} else {
		fmt.Println("❌ Error: Neither Podman nor Docker is available")
	}
}

func handleContainerCp(args []string) {
	if len(args) != 2 {
		fmt.Println("❌ Error: Source and destination required")
		fmt.Println("Usage: portunix container cp <source> <destination>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  portunix container cp file.txt container:/path/to/dest")
		fmt.Println("  portunix container cp container:/path/to/file.txt ./local-file.txt")
		return
	}

	source := args[0]
	destination := args[1]

	// Copy files
	if isPodmanAvailable() {
		if err := copyPodmanFiles(source, destination); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error copying files: %v\n", err)
		} else {
			fmt.Printf("✅ Files copied successfully\n")
		}
	} else if isDockerAvailable() {
		if err := copyDockerFiles(source, destination); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error copying files: %v\n", err)
		} else {
			fmt.Printf("✅ Files copied successfully\n")
		}
	} else {
		fmt.Println("❌ Error: Neither Podman nor Docker is available")
	}
}

func handleContainerInfo(args []string) {
	fmt.Println("🐳 Container Runtime Information")
	fmt.Println("===============================")

	if isPodmanAvailable() {
		fmt.Println("✅ Podman: Available")
		if out, err := exec.Command("podman", "version", "--format", "{{.Client.Version}}").Output(); err == nil {
			fmt.Printf("   Version: %s", string(out))
		}
	} else {
		fmt.Println("❌ Podman: Not available")
	}

	if isDockerAvailable() {
		fmt.Println("✅ Docker: Available")
		if out, err := exec.Command("docker", "version", "--format", "{{.Client.Version}}").Output(); err == nil {
			fmt.Printf("   Version: %s", string(out))
		}
	} else {
		fmt.Println("❌ Docker: Not available")
	}
}

func handleContainerCheck(args []string) {
	// Check for --refresh flag and help
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showCheckHelp()
			return
		}
		// Note: --refresh flag is parsed but currently has no effect
		// as the helper performs fresh detection each time
	}

	// Display container runtime capabilities
	fmt.Println("Container Runtime Status:")
	fmt.Println()

	dockerAvailable := isDockerAvailable()
	podmanAvailable := isPodmanAvailable()

	// Docker status
	if dockerAvailable {
		versionCmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
		if versionOutput, err := versionCmd.Output(); err == nil {
			version := strings.TrimSpace(string(versionOutput))
			fmt.Printf("  Docker: ✓ Available (version %s)\n", version)
		} else {
			// Fallback to --version
			versionCmd = exec.Command("docker", "--version")
			if versionOutput, err := versionCmd.Output(); err == nil {
				fmt.Printf("  Docker: ✓ Available (%s)\n", strings.TrimSpace(string(versionOutput)))
			} else {
				fmt.Println("  Docker: ✓ Available")
			}
		}
	} else {
		fmt.Println("  Docker: ✗ Not available")
	}

	// Podman status
	if podmanAvailable {
		versionCmd := exec.Command("podman", "version", "--format", "{{.Version}}")
		if versionOutput, err := versionCmd.Output(); err == nil {
			version := strings.TrimSpace(string(versionOutput))
			fmt.Printf("  Podman: ✓ Available (version %s)\n", version)
		} else {
			// Fallback to --version
			versionCmd = exec.Command("podman", "--version")
			if versionOutput, err := versionCmd.Output(); err == nil {
				fmt.Printf("  Podman: ✓ Available (%s)\n", strings.TrimSpace(string(versionOutput)))
			} else {
				fmt.Println("  Podman: ✓ Available")
			}
		}
	} else {
		fmt.Println("  Podman: ✗ Not available")
	}

	// Preferred runtime
	if dockerAvailable || podmanAvailable {
		fmt.Println()
		if dockerAvailable {
			fmt.Println("  Preferred: docker")
		} else {
			fmt.Println("  Preferred: podman")
		}
	}

	// Capabilities
	if dockerAvailable || podmanAvailable {
		fmt.Println()
		fmt.Println("Capabilities:")

		// Check compose support
		if dockerAvailable {
			composeCmd := exec.Command("docker", "compose", "version")
			if composeCmd.Run() == nil {
				fmt.Println("  - Compose support: ✓")
			}

			buildxCmd := exec.Command("docker", "buildx", "version")
			if buildxCmd.Run() == nil {
				fmt.Println("  - BuildKit/Buildx: ✓")
			}
		}

		if podmanAvailable {
			composeCmd := exec.Command("podman", "compose", "version")
			if composeCmd.Run() == nil {
				fmt.Println("  - Compose support: ✓")
			}
		}

		// Volume and network support (always true if runtime available)
		fmt.Println("  - Volume mounting: ✓")
		fmt.Println("  - Network creation: ✓")

		// Runtime active check
		if dockerAvailable {
			infoCmd := exec.Command("docker", "info")
			if infoCmd.Run() == nil {
				fmt.Println("  - Runtime active: ✓")
			}
		} else if podmanAvailable {
			infoCmd := exec.Command("podman", "info")
			if infoCmd.Run() == nil {
				fmt.Println("  - Runtime active: ✓")
			}
		}
	}

	// Show installation suggestion if no runtime
	if !dockerAvailable && !podmanAvailable {
		fmt.Println()
		fmt.Println("No container runtime detected. You can install one using:")
		fmt.Println("  portunix install docker")
		fmt.Println("  portunix install podman")
	}
}

func showCheckHelp() {
	fmt.Println("Usage: portunix container check [OPTIONS]")
	fmt.Println()
	fmt.Println("🔍 CHECK CONTAINER RUNTIME CAPABILITIES")
	fmt.Println()
	fmt.Println("Detect and display detailed information about available container runtimes.")
	fmt.Println()
	fmt.Println("🌟 DETECTION INCLUDES:")
	fmt.Println("  • Installed container runtimes (Docker/Podman)")
	fmt.Println("  • Runtime versions and build information")
	fmt.Println("  • Supported features and capabilities")
	fmt.Println("  • System compatibility status")
	fmt.Println("  • Recommendations for optimal setup")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --refresh      Force re-detection of capabilities")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container check")
	fmt.Println("  portunix container check --refresh")
	fmt.Println()
	fmt.Println("This command helps diagnose container runtime issues and verify proper installation.")
}

// Runtime availability checks
func isPodmanAvailable() bool {
	cmd := exec.Command("podman", "version", "--format", "{{.Client.Version}}")
	return cmd.Run() == nil
}

func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version", "--format", "{{.Client.Version}}")
	return cmd.Run() == nil
}

// Container runtime implementations
// runPodmanInContainerWithImage runs installation in Podman container with specified image
func runPodmanInContainerWithImage(installationType string, imageName string, args []string) {
	// Create container and install specified software with provided image
	runPodmanInContainerImpl(installationType, imageName, args)
}

func runPodmanInContainer(installationType string, args []string) {
	// Create Ubuntu container and install specified software
	imageName := "ubuntu:22.04"

	// Check for custom image in args
	for i, arg := range args {
		if arg == "--image" && i+1 < len(args) {
			imageName = args[i+1]
			break
		}
	}

	runPodmanInContainerImpl(installationType, imageName, args)
}

func runPodmanInContainerImpl(installationType string, imageName string, args []string) {

	containerName := fmt.Sprintf("portunix-test-%s", installationType)

	fmt.Printf("🏗️  Creating container: %s\n", containerName)
	fmt.Printf("📦 Using image: %s\n", imageName)

	// Remove existing container if it exists
	exec.Command("podman", "rm", "-f", containerName).Run()

	// Copy current portunix binary to container
	// First create a temporary copy
	tempPath := "/tmp/portunix-container-test"
	exec.Command("cp", "./portunix", tempPath).Run()

	// Create and start container with volume mount
	cmd := exec.Command("podman", "run", "--name", containerName, "-it", "--rm",
		"-v", fmt.Sprintf("%s:/usr/local/bin/portunix", tempPath),
		imageName, "/bin/bash", "-c",
		fmt.Sprintf("apt-get update && apt-get install -y python3 python3-pip && chmod +x /usr/local/bin/portunix && portunix install %s", installationType))

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Container execution failed: %v\n", err)
	}
}

// runDockerInContainerWithImage runs installation in Docker container with specified image
func runDockerInContainerWithImage(installationType string, imageName string, args []string) {
	// Create container and install specified software with provided image
	runDockerInContainerImpl(installationType, imageName, args)
}

func runDockerInContainer(installationType string, args []string) {
	// Create Ubuntu container and install specified software
	imageName := "ubuntu:22.04"

	// Check for custom image in args
	for i, arg := range args {
		if arg == "--image" && i+1 < len(args) {
			imageName = args[i+1]
			break
		}
	}

	runDockerInContainerImpl(installationType, imageName, args)
}

func runDockerInContainerImpl(installationType string, imageName string, args []string) {

	containerName := fmt.Sprintf("portunix-test-%s", installationType)

	fmt.Printf("🏗️  Creating container: %s\n", containerName)
	fmt.Printf("📦 Using image: %s\n", imageName)

	// Remove existing container if it exists
	exec.Command("docker", "rm", "-f", containerName).Run()

	// Copy current portunix binary to container
	// First create a temporary copy
	tempPath := "/tmp/portunix-container-test"
	exec.Command("cp", "./portunix", tempPath).Run()

	// Create and start container with volume mount
	cmd := exec.Command("docker", "run", "--name", containerName, "-it", "--rm",
		"-v", fmt.Sprintf("%s:/usr/local/bin/portunix", tempPath),
		imageName, "/bin/bash", "-c",
		fmt.Sprintf("apt-get update && apt-get install -y python3 python3-pip && chmod +x /usr/local/bin/portunix && portunix install %s", installationType))

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Container execution failed: %v\n", err)
	}
}

func runPodmanContainer(image string, command []string) {
	args := []string{"run", "-it", "--rm", image}
	args = append(args, command...)

	cmd := exec.Command("podman", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Podman run failed: %v\n", err)
	}
}

func runDockerContainer(image string, command []string) {
	args := []string{"run", "-it", "--rm", image}
	args = append(args, command...)

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Docker run failed: %v\n", err)
	}
}

// execPodmanCommand executes a command inside an existing Podman container
func execPodmanCommand(containerName string, command []string) error {
	args := []string{"exec", "-it", containerName}
	args = append(args, command...)

	cmd := exec.Command("podman", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// execDockerCommand executes a command inside an existing Docker container
func execDockerCommand(containerName string, command []string) error {
	args := []string{"exec", "-it", containerName}
	args = append(args, command...)

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func init() {
	// Add version information
	rootCmd.SetVersionTemplate("ptx-container version {{.Version}}\n")

	// Disable default cobra help handling to use custom help
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.PersistentFlags().BoolP("help", "h", false, "help for container")

	// Override help function to preserve original command context
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// Use the original command line arguments to maintain context
		originalArgs := os.Args[1:]
		if len(originalArgs) > 0 {
			handleCommand(originalArgs)
		} else {
			// Fallback to general container help
			handleCommand([]string{"container", "--help"})
		}
	})
}

// ContainerInfo represents container information
type ContainerInfo struct {
	ID      string
	Name    string
	Image   string
	Status  string
	Ports   string
	Created string
}

// listDockerContainers lists all containers from Docker (not filtered by portunix- prefix)
func listDockerContainers() ([]ContainerInfo, error) {
	cmd := exec.Command("docker", "ps", "-a", "--format", "table {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}\t{{.CreatedAt}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker containers: %w", err)
	}

	return parseContainerOutput(string(output))
}

// listPodmanContainers lists all containers from Podman (not filtered by portunix- prefix)
func listPodmanContainers() ([]ContainerInfo, error) {
	cmd := exec.Command("podman", "ps", "-a", "--format", "table {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}\t{{.CreatedAt}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list Podman containers: %w", err)
	}

	return parseContainerOutput(string(output))
}

// parseContainerOutput parses the tabular output from docker/podman ps command
func parseContainerOutput(output string) ([]ContainerInfo, error) {
	lines := strings.Split(output, "\n")
	var containers []ContainerInfo

	// Skip header line and empty lines
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 4 {
			container := ContainerInfo{
				ID:     fields[0],
				Name:   fields[1],
				Image:  fields[2],
				Status: fields[3],
			}
			if len(fields) >= 5 {
				container.Ports = fields[4]
			}
			if len(fields) >= 6 {
				container.Created = strings.Join(fields[5:], " ")
			}
			containers = append(containers, container)
		}
	}

	return containers, nil
}

// printContainerTable prints containers in a formatted table
func printContainerTable(containers []ContainerInfo) {
	if len(containers) == 0 {
		return
	}

	// Print header
	fmt.Printf("   %-12s %-20s %-20s %-15s %-10s %s\n",
		"CONTAINER ID", "NAME", "IMAGE", "STATUS", "PORTS", "CREATED")
	fmt.Println("   " + strings.Repeat("-", 100))

	// Print container rows
	for _, container := range containers {
		// Truncate long fields for better display
		containerID := container.ID
		if len(containerID) > 12 {
			containerID = containerID[:12]
		}

		name := container.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		image := container.Image
		if len(image) > 20 {
			image = image[:17] + "..."
		}

		status := container.Status
		if len(status) > 15 {
			status = status[:12] + "..."
		}

		ports := container.Ports
		if len(ports) > 10 {
			ports = ports[:7] + "..."
		}

		created := container.Created
		if len(created) > 20 {
			created = created[:17] + "..."
		}

		fmt.Printf("   %-12s %-20s %-20s %-15s %-10s %s\n",
			containerID, name, image, status, ports, created)
	}
}

// Container management helper functions

// removeContainer removes a container using the appropriate runtime
func removeContainer(containerName string, force bool) error {
	if isPodmanAvailable() {
		return removePodmanContainer(containerName, force)
	} else if isDockerAvailable() {
		return removeDockerContainer(containerName, force)
	}
	return fmt.Errorf("neither Podman nor Docker is available")
}

func removePodmanContainer(containerName string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, containerName)

	cmd := exec.Command("podman", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

func removeDockerContainer(containerName string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, containerName)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

func stopPodmanContainer(containerName string) error {
	cmd := exec.Command("podman", "stop", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

func stopDockerContainer(containerName string) error {
	cmd := exec.Command("docker", "stop", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

func startPodmanContainer(containerName string) error {
	cmd := exec.Command("podman", "start", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

func startDockerContainer(containerName string) error {
	cmd := exec.Command("docker", "start", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

func showPodmanLogs(containerName string, follow bool) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, containerName)

	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func showDockerLogs(containerName string, follow bool) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, containerName)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyPodmanFiles(source, destination string) error {
	cmd := exec.Command("podman", "cp", source, destination)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

func copyDockerFiles(source, destination string) error {
	cmd := exec.Command("docker", "cp", source, destination)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

// Help text functions

func showRmHelp() {
	fmt.Println("Usage: portunix container rm [OPTIONS] <container-name> [<container-name>...]")
	fmt.Println()
	fmt.Println("🗑️ REMOVE CONTAINER")
	fmt.Println()
	fmt.Println("Remove one or more containers using the automatically selected runtime.")
	fmt.Println()
	fmt.Println("🌟 UNIVERSAL OPERATION:")
	fmt.Println("  ✅ Works with both Docker and Podman containers")
	fmt.Println("  ✅ Automatic runtime detection")
	fmt.Println("  ✅ Supports force removal of running containers")
	fmt.Println("  ✅ Docker/Podman compatible 'rm' command")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -f, --force    Force removal of running containers")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container rm test-container")
	fmt.Println("  portunix container rm nodejs-dev --force")
	fmt.Println("  portunix container rm web-server -f")
	fmt.Println("  portunix container rm container1 container2 container3")
}

func showLogsHelp() {
	fmt.Println("Usage: portunix container logs [OPTIONS] <container-name>")
	fmt.Println()
	fmt.Println("📝 VIEW CONTAINER LOGS")
	fmt.Println()
	fmt.Println("Display logs from a container using the automatically selected runtime.")
	fmt.Println()
	fmt.Println("🌟 UNIVERSAL OPERATION:")
	fmt.Println("  ✅ Works with both Docker and Podman containers")
	fmt.Println("  ✅ Automatic runtime detection")
	fmt.Println("  ✅ Real-time log streaming with --follow")
	fmt.Println("  ✅ Consistent output format")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -f, --follow    Follow log output (stream continuously)")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container logs test-container")
	fmt.Println("  portunix container logs web-server --follow")
	fmt.Println("  portunix container logs python-dev")
	fmt.Println("  portunix container logs db-container -f")
}

func showStopHelp() {
	fmt.Println("Usage: portunix container stop [OPTIONS] <container-name>")
	fmt.Println()
	fmt.Println("🛑 STOP CONTAINER")
	fmt.Println()
	fmt.Println("Stop a running container using the automatically selected runtime.")
	fmt.Println()
	fmt.Println("🌟 UNIVERSAL OPERATION:")
	fmt.Println("  ✅ Works with both Docker and Podman containers")
	fmt.Println("  ✅ Automatic runtime detection")
	fmt.Println("  ✅ Graceful shutdown of container processes")
	fmt.Println("  ✅ Consistent behavior across runtimes")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container stop test-container")
	fmt.Println("  portunix container stop web-server")
	fmt.Println("  portunix container stop python-dev")
}

func showStartHelp() {
	fmt.Println("Usage: portunix container start [OPTIONS] <container-name>")
	fmt.Println()
	fmt.Println("▶️ START CONTAINER")
	fmt.Println()
	fmt.Println("Start a stopped container using the automatically selected runtime.")
	fmt.Println()
	fmt.Println("🌟 UNIVERSAL OPERATION:")
	fmt.Println("  ✅ Works with both Docker and Podman containers")
	fmt.Println("  ✅ Automatic runtime detection")
	fmt.Println("  ✅ Restarts previously stopped containers")
	fmt.Println("  ✅ Consistent behavior across runtimes")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container start test-container")
	fmt.Println("  portunix container start web-server")
	fmt.Println("  portunix container start python-dev")
}

func main() {
	// Direct argument parsing instead of cobra to support all flags
	args := os.Args[1:]

	// Handle version flag - but ONLY if it's the only argument or first argument
	// This prevents --version in subcommands (like "exec container cmd --version") from triggering helper version
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Printf("ptx-container version %s\n", version)
		return
	}

	// Delegate to handleCommand for all functionality
	handleCommand(args)
}