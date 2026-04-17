/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var version = "dev"
var debugMode = false

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

// handleCommand dispatches commands routed to this helper by the parent portunix
// binary (see src/dispatcher/dispatcher.go): "container", "docker", and "podman".
// It strips the global --debug flag (used to log the underlying podman/docker
// argv) from args before routing. args arrive without the binary name prefix,
// so args[0] is the top-level command and the rest are subcommand + flags.
func handleCommand(args []string) {
	// Handle dispatched commands: container, docker, podman
	if len(args) == 0 {
		fmt.Println("No command specified")
		return
	}

	// Extract --debug flag from args
	var filteredArgs []string
	for _, arg := range args {
		if arg == "--debug" {
			debugMode = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	args = filteredArgs

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
			fmt.Println("  compose          Run docker-compose/podman-compose commands (universal runtime)")
			fmt.Println("  compose-preflight Check if compose is ready (daemon/socket running)")
			fmt.Println("  cp               Copy files/folders between container and host")
			fmt.Println("  exec             Execute command in container (universal runtime)")
			fmt.Println("  info             Show container runtime information and availability")
			fmt.Println("  inspect          Show low-level container details (universal runtime)")
			fmt.Println("  list             List containers from all available runtimes")
			fmt.Println("  logs             Show container logs (universal runtime)")
			fmt.Println("  network          Manage container networks (create/list/inspect/rm)")
			fmt.Println("  rm               Remove container (universal runtime)")
			fmt.Println("  run              Run new container (universal runtime)")
			fmt.Println("  run-in-container Run installation in container (RECOMMENDED for testing)")
			fmt.Println("  start            Start stopped container (universal runtime)")
			fmt.Println("  stop             Stop container (universal runtime)")
			fmt.Println("  volume           Manage container volumes (create/list/inspect/rm/prune)")
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
	case "compose":
		handleContainerCompose(cmdArgs)
	case "compose-preflight":
		handleComposePreflight(cmdArgs)
	case "network":
		handleContainerNetwork(cmdArgs)
	case "volume":
		handleContainerVolume(cmdArgs)
	case "inspect":
		handleContainerInspect(cmdArgs)
	default:
		fmt.Printf("Unknown %s subcommand: %s\n", command, subcommand)
		fmt.Printf("Available subcommands: run, run-in-container, exec, list, stop, start, rm, logs, cp, info, check, compose, compose-preflight, network, volume, inspect\n")
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
	fmt.Println("  --network: Connect container to a network")
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
	// Check for help flag first
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showExecHelp()
			return
		}
	}

	if len(args) < 2 {
		showExecHelp()
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
					os.Exit(1)
				}
			} else {
				fmt.Fprintf(os.Stderr, "❌ Error: Failed to execute command in container '%s': %v\n", containerName, err)
				os.Exit(1)
			}
		}
	} else if isDockerAvailable() {
		if err := execDockerCommand(containerName, command); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: Failed to execute command in container '%s': %v\n", containerName, err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintln(os.Stderr, "❌ Error: Neither Podman nor Docker is available")
		fmt.Fprintln(os.Stderr, "Please install Podman or Docker first")
		os.Exit(1)
	}
}

func handleContainerList(args []string) {
	// Check for help flag first
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showListHelp()
			return
		}
	}

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
	// Check for help flag first
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showCpHelp()
			return
		}
	}

	if len(args) != 2 {
		showCpHelp()
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
	// Check for help flag first
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showInfoHelp()
			return
		}
	}

	fmt.Println("🐳 Container Runtime Information")
	fmt.Println("===============================")

	// Podman status
	if isPodmanInstalled() {
		if isPodmanAvailable() {
			fmt.Println("✅ Podman: Available (running)")
			if out, err := exec.Command("podman", "version", "--format", "{{.Client.Version}}").Output(); err == nil {
				fmt.Printf("   Version: %s", string(out))
			}
		} else {
			fmt.Println("⚠️  Podman: Installed but not running")
			if out, err := exec.Command("podman", "--version").Output(); err == nil {
				fmt.Printf("   Version: %s", string(out))
			}
		}
	} else {
		fmt.Println("❌ Podman: Not installed")
	}

	// Docker status
	if isDockerInstalled() {
		if isDockerAvailable() {
			fmt.Println("✅ Docker: Available (running)")
			if out, err := exec.Command("docker", "version", "--format", "{{.Client.Version}}").Output(); err == nil {
				fmt.Printf("   Version: %s", string(out))
			}
		} else {
			fmt.Println("⚠️  Docker: Installed but daemon not running")
			if out, err := exec.Command("docker", "--version").Output(); err == nil {
				fmt.Printf("   Version: %s", string(out))
			}
		}
	} else {
		fmt.Println("❌ Docker: Not installed")
	}
}

// ComposeStatus represents the status of compose readiness
type ComposeStatus struct {
	Ready           bool
	Runtime         string
	Version         string
	DaemonRunning   bool
	ErrorMessage    string
	FixInstructions string
}

// CheckComposeReady checks if compose is ready to use and returns detailed status
func CheckComposeReady() ComposeStatus {
	status := ComposeStatus{}

	// Check Docker first
	dockerInstalled := isDockerCliInstalled()
	if dockerInstalled {
		// Check if Docker daemon is running
		if cmd := exec.Command("docker", "info"); cmd.Run() == nil {
			status.DaemonRunning = true
			// Check for Compose V2
			if cmd := exec.Command("docker", "compose", "version", "--short"); cmd.Run() == nil {
				output, _ := exec.Command("docker", "compose", "version", "--short").Output()
				status.Ready = true
				status.Runtime = "Docker Compose V2"
				status.Version = strings.TrimSpace(string(output))
				return status
			}
			// Check for Compose V1
			if cmd := exec.Command("docker-compose", "version", "--short"); cmd.Run() == nil {
				output, _ := exec.Command("docker-compose", "version", "--short").Output()
				status.Ready = true
				status.Runtime = "Docker Compose V1"
				status.Version = strings.TrimSpace(string(output))
				return status
			}
		} else {
			status.ErrorMessage = "Docker CLI installed but daemon is not running"
			status.FixInstructions = "Start Docker daemon with: sudo systemctl start docker"
			return status
		}
	}

	// Check Podman
	podmanInstalled := isPodmanCliInstalled()
	if podmanInstalled {
		// For Podman, we need to check if the socket file exists
		// because podman info can work without the socket, but compose needs it
		socketRunning := isPodmanSocketRunning()

		if !socketRunning {
			status.ErrorMessage = "Podman installed but socket is not running"
			status.FixInstructions = "systemctl --user enable --now podman.socket"
			return status
		}

		status.DaemonRunning = true

		// Check for built-in podman compose
		if cmd := exec.Command("podman", "compose", "version"); cmd.Run() == nil {
			output, _ := exec.Command("podman", "compose", "version").Output()
			status.Ready = true
			status.Runtime = "Podman Compose"
			status.Version = strings.TrimSpace(string(output))
			return status
		}
		// Check for standalone podman-compose
		if cmd := exec.Command("podman-compose", "version"); cmd.Run() == nil {
			output, _ := exec.Command("podman-compose", "version").Output()
			status.Ready = true
			status.Runtime = "podman-compose"
			status.Version = strings.TrimSpace(string(output))
			return status
		}
		// Podman works but no compose tool
		status.ErrorMessage = "Podman is running but no compose tool is available"
		status.FixInstructions = "Install podman-compose: pip install podman-compose"
		return status
	}

	// No container runtime found
	status.ErrorMessage = "No container runtime (Docker or Podman) is installed"
	status.FixInstructions = "Install Docker or Podman first"
	return status
}

// isPodmanSocketRunning checks if podman socket file exists and is accessible
func isPodmanSocketRunning() bool {
	// Check XDG_RUNTIME_DIR for user socket
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		// Fallback to /run/user/<uid>
		runtimeDir = fmt.Sprintf("/run/user/%d", os.Getuid())
	}

	socketPath := filepath.Join(runtimeDir, "podman", "podman.sock")

	// Check if socket file exists
	if _, err := os.Stat(socketPath); err == nil {
		return true
	}

	// Also try the systemctl status as fallback
	cmd := exec.Command("systemctl", "--user", "is-active", "podman.socket")
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) == "active" {
		return true
	}

	return false
}

// isDockerCliInstalled checks if Docker CLI is installed (not if daemon is running)
func isDockerCliInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// isPodmanCliInstalled checks if Podman CLI is installed (not if socket is running)
func isPodmanCliInstalled() bool {
	_, err := exec.LookPath("podman")
	return err == nil
}

// handleComposePreflight checks compose readiness and prints status
func handleComposePreflight(args []string) {
	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showComposePreflightHelp()
			return
		}
	}

	// Check for --json flag
	jsonOutput := false
	for _, arg := range args {
		if arg == "--json" {
			jsonOutput = true
			break
		}
	}

	status := CheckComposeReady()

	if jsonOutput {
		// JSON output for programmatic use
		fmt.Printf(`{"ready":%t,"runtime":"%s","version":"%s","daemon_running":%t,"error":"%s","fix":"%s"}`,
			status.Ready, status.Runtime, status.Version, status.DaemonRunning,
			status.ErrorMessage, status.FixInstructions)
		fmt.Println()
		if !status.Ready {
			os.Exit(1)
		}
		return
	}

	// Human-readable output
	if status.Ready {
		fmt.Printf("✅ Compose is ready: %s (%s)\n", status.Runtime, status.Version)
	} else {
		fmt.Printf("❌ Compose is NOT ready\n\n")
		fmt.Printf("Problem: %s\n\n", status.ErrorMessage)
		fmt.Printf("Solution: %s\n", status.FixInstructions)
		os.Exit(1)
	}
}

func showComposePreflightHelp() {
	fmt.Println("Usage: portunix container compose-preflight [OPTIONS]")
	fmt.Println()
	fmt.Println("🔍 CHECK COMPOSE READINESS")
	fmt.Println()
	fmt.Println("Verify that compose tools are ready to use. This checks:")
	fmt.Println("  • Docker/Podman installation")
	fmt.Println("  • Docker daemon or Podman socket status")
	fmt.Println("  • Compose tool availability")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --json       Output result as JSON for programmatic use")
	fmt.Println("  -h, --help   Show this help message")
	fmt.Println()
	fmt.Println("Exit codes:")
	fmt.Println("  0 - Compose is ready")
	fmt.Println("  1 - Compose is NOT ready (with instructions to fix)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container compose-preflight")
	fmt.Println("  portunix container compose-preflight --json")
}

// handleContainerCompose handles compose subcommand - passes through to docker-compose/podman-compose
func handleContainerCompose(args []string) {
	// Handle help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showComposeHelp()
			return
		}
	}

	// If no arguments, show compose runtime info
	if len(args) == 0 {
		runtime, version := detectComposeRuntime()
		if runtime == "" {
			fmt.Println("❌ " + getComposeInstallInstructions())
			return
		}
		fmt.Printf("🐳 Compose Runtime: %s (version %s)\n", runtime, version)
		fmt.Println("\nUse 'portunix container compose --help' for usage information.")
		return
	}

	// Detect and execute compose command
	runtime, _ := detectComposeRuntime()
	if runtime == "" {
		fmt.Println("❌ " + getComposeInstallInstructions())
		return
	}

	// Execute compose command
	var cmd *exec.Cmd
	switch runtime {
	case "Docker Compose V2":
		cmdArgs := append([]string{"compose"}, args...)
		cmd = exec.Command("docker", cmdArgs...)
	case "Docker Compose V1":
		cmd = exec.Command("docker-compose", args...)
	case "Podman Compose":
		// Built-in podman compose (Podman 3.0+)
		cmdArgs := append([]string{"compose"}, args...)
		cmd = exec.Command("podman", cmdArgs...)
	case "Podman Compose (standalone)":
		cmd = exec.Command("podman-compose", args...)
	default:
		fmt.Printf("❌ Unknown compose runtime: %s\n", runtime)
		return
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Compose command failed: %v\n", err)
		os.Exit(1)
	}
}

// detectComposeRuntime detects available compose tool and returns name and version
// It checks if the daemon is actually running, not just if the CLI binary exists
func detectComposeRuntime() (string, string) {
	// Try Docker - verify daemon is running with "docker info"
	if cmd := exec.Command("docker", "info"); cmd.Run() == nil {
		// Docker daemon is running, check for Compose V2
		if cmd := exec.Command("docker", "compose", "version", "--short"); cmd.Run() == nil {
			output, _ := exec.Command("docker", "compose", "version", "--short").Output()
			return "Docker Compose V2", strings.TrimSpace(string(output))
		}
		// Try Docker Compose V1
		if cmd := exec.Command("docker-compose", "version", "--short"); cmd.Run() == nil {
			output, _ := exec.Command("docker-compose", "version", "--short").Output()
			return "Docker Compose V1", strings.TrimSpace(string(output))
		}
	}

	// Try Podman - verify it's available with "podman info"
	if cmd := exec.Command("podman", "info"); cmd.Run() == nil {
		// Try built-in podman compose (Podman 3.0+)
		if cmd := exec.Command("podman", "compose", "version"); cmd.Run() == nil {
			output, _ := exec.Command("podman", "compose", "version").Output()
			return "Podman Compose", strings.TrimSpace(string(output))
		}
		// Try standalone podman-compose
		if cmd := exec.Command("podman-compose", "version"); cmd.Run() == nil {
			output, _ := exec.Command("podman-compose", "version").Output()
			return "Podman Compose (standalone)", strings.TrimSpace(string(output))
		}
	}

	return "", ""
}

// getComposeInstallInstructions returns help text for installing compose tools
func getComposeInstallInstructions() string {
	return `No compose tool detected. Install one of the following:

For Docker:
  Docker Compose V2 (recommended): Included with Docker Desktop
  Docker Compose V1: portunix install docker-compose

For Podman:
  podman-compose: portunix install podman-compose

Tip: Use 'portunix container check' to verify container runtime availability.`
}

// showComposeHelp displays help for compose subcommand
func showComposeHelp() {
	fmt.Println("Usage: portunix container compose [args...]")
	fmt.Println()
	fmt.Println("🐳 UNIVERSAL COMPOSE COMMAND")
	fmt.Println()
	fmt.Println("Execute docker-compose or podman-compose commands using automatic runtime detection.")
	fmt.Println("All arguments are passed directly to the detected compose tool.")
	fmt.Println()
	fmt.Println("🌟 AUTOMATIC RUNTIME DETECTION:")
	fmt.Println("  Priority order:")
	fmt.Println("  1. Docker Compose V2 (docker compose) - preferred")
	fmt.Println("  2. Docker Compose V1 (docker-compose) - fallback")
	fmt.Println("  3. Podman Compose (podman-compose) - alternative")
	fmt.Println()
	fmt.Println("💡 USAGE:")
	fmt.Println("  All standard compose commands and flags are supported:")
	fmt.Println()
	fmt.Println("  portunix container compose -f <file> up [service]")
	fmt.Println("  portunix container compose -f <file> down")
	fmt.Println("  portunix container compose -f <file> build [service]")
	fmt.Println("  portunix container compose -f <file> logs [service]")
	fmt.Println("  portunix container compose -f <file> ps")
	fmt.Println("  portunix container compose -f <file> exec <service> <command>")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container compose -f docker-compose.yml up -d")
	fmt.Println("  portunix container compose -f docker-compose.yml down")
	fmt.Println("  portunix container compose -f docker-compose.docs.yml up docs-server")
	fmt.Println("  portunix container compose -f docker-compose.yml logs -f web")
	fmt.Println("  portunix container compose -f docker-compose.yml ps")
	fmt.Println("  portunix container compose -f docker-compose.yml build --no-cache")
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

// isPodmanInstalled checks if Podman binary exists in PATH
func isPodmanInstalled() bool {
	_, err := exec.LookPath("podman")
	return err == nil
}

// isPodmanAvailable checks if Podman is installed AND functional
func isPodmanAvailable() bool {
	if !isPodmanInstalled() {
		return false
	}
	cmd := exec.Command("podman", "info")
	return cmd.Run() == nil
}

// isDockerInstalled checks if Docker binary exists in PATH
func isDockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// isDockerAvailable checks if Docker is installed AND daemon is running
func isDockerAvailable() bool {
	if !isDockerInstalled() {
		return false
	}
	cmd := exec.Command("docker", "info")
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

	// Build run arguments with TTY detection
	var runArgs []string
	if term.IsTerminal(int(os.Stdin.Fd())) {
		runArgs = []string{"run", "--name", containerName, "-it", "--rm"}
	} else {
		runArgs = []string{"run", "--name", containerName, "-i", "--rm"}
	}
	runArgs = append(runArgs, "-v", fmt.Sprintf("%s:/usr/local/bin/portunix", tempPath))
	runArgs = append(runArgs, imageName, "/bin/bash", "-c",
		fmt.Sprintf("apt-get update && apt-get install -y python3 python3-pip && chmod +x /usr/local/bin/portunix && portunix install %s", installationType))

	cmd := exec.Command("podman", runArgs...)
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

	// Build run arguments with TTY detection
	var runArgs []string
	if term.IsTerminal(int(os.Stdin.Fd())) {
		runArgs = []string{"run", "--name", containerName, "-it", "--rm"}
	} else {
		runArgs = []string{"run", "--name", containerName, "-i", "--rm"}
	}
	runArgs = append(runArgs, "-v", fmt.Sprintf("%s:/usr/local/bin/portunix", tempPath))
	runArgs = append(runArgs, imageName, "/bin/bash", "-c",
		fmt.Sprintf("apt-get update && apt-get install -y python3 python3-pip && chmod +x /usr/local/bin/portunix && portunix install %s", installationType))

	cmd := exec.Command("docker", runArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Container execution failed: %v\n", err)
	}
}

func runPodmanContainer(image string, command []string) {
	// Check if running in detached mode (service container)
	// Detached containers should NOT use --rm as they are persistent services
	detached := isDetachedMode(image, command)

	// Only use -t flag if stdin is a terminal
	// This prevents "the input device is not a TTY" error
	var args []string
	if detached {
		args = []string{"run", image}
	} else if term.IsTerminal(int(os.Stdin.Fd())) {
		args = []string{"run", "-it", "--rm", image}
	} else {
		args = []string{"run", "-i", "--rm", image}
	}
	args = append(args, command...)

	if debugMode {
		fmt.Fprintf(os.Stderr, "🔍 DEBUG podman args: %v\n", args)
		fmt.Fprintf(os.Stderr, "🔍 DEBUG detached: %v\n", detached)
	}

	cmd := exec.Command("podman", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Podman run failed: %v\n", err)
		// Surface the failure to the parent process (dispatcher / ptx-installer)
		// — otherwise callers see a 0 exit and treat a failed run as a success.
		os.Exit(1)
	}
}

func runDockerContainer(image string, command []string) {
	// Check if running in detached mode (service container)
	// Detached containers should NOT use --rm as they are persistent services
	detached := isDetachedMode(image, command)

	// Only use -t flag if stdin is a terminal
	// This prevents "the input device is not a TTY" error
	var args []string
	if detached {
		args = []string{"run", image}
	} else if term.IsTerminal(int(os.Stdin.Fd())) {
		args = []string{"run", "-it", "--rm", image}
	} else {
		args = []string{"run", "-i", "--rm", image}
	}
	args = append(args, command...)

	if debugMode {
		fmt.Fprintf(os.Stderr, "🔍 DEBUG docker args: %v\n", args)
		fmt.Fprintf(os.Stderr, "🔍 DEBUG detached: %v\n", detached)
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Docker run failed: %v\n", err)
		// See runPodmanContainer — surface failure so the parent sees non-zero exit.
		os.Exit(1)
	}
}

// isDetachedMode checks if -d or --detach flag is present in image name or command args
// When called from handleContainerRun, flags like -d end up as the "image" parameter
func isDetachedMode(image string, command []string) bool {
	if image == "-d" || image == "--detach" {
		return true
	}
	for _, arg := range command {
		if arg == "-d" || arg == "--detach" {
			return true
		}
	}
	return false
}

// execPodmanCommand executes a command inside an existing Podman container
func execPodmanCommand(containerName string, command []string) error {
	// Only use -t flag if stdin is a terminal (interactive mode)
	// This prevents "the input device is not a TTY" error on Windows
	var args []string
	if term.IsTerminal(int(os.Stdin.Fd())) {
		args = []string{"exec", "-it", containerName}
	} else {
		args = []string{"exec", "-i", containerName}
	}
	args = append(args, command...)

	cmd := exec.Command("podman", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// execDockerCommand executes a command inside an existing Docker container
func execDockerCommand(containerName string, command []string) error {
	// Only use -t flag if stdin is a terminal (interactive mode)
	// This prevents "the input device is not a TTY" error on Windows
	var args []string
	if term.IsTerminal(int(os.Stdin.Fd())) {
		args = []string{"exec", "-it", containerName}
	} else {
		args = []string{"exec", "-i", containerName}
	}
	args = append(args, command...)

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// selectRuntime returns "podman" if installed, otherwise "docker".
// Unlike isPodmanAvailable, this uses binary presence (LookPath) only — it
// does not call `podman info`, which can hang on misconfigured hosts. These
// passthrough subcommands (network, volume, inspect) do not need a daemon
// pre-check: the runtime surfaces any operational errors natively.
func selectRuntime() (string, error) {
	if isPodmanInstalled() {
		return "podman", nil
	}
	if isDockerInstalled() {
		return "docker", nil
	}
	return "", fmt.Errorf("neither Podman nor Docker is available")
}

// runPassthrough runs a runtime command with inherited stdio and returns its exit code.
// Used for subcommands that should surface the runtime's native output and exit status
// verbatim (list, inspect, volume prune, etc.).
func runPassthrough(runtime string, args ...string) int {
	if debugMode {
		fmt.Fprintf(os.Stderr, "🔍 DEBUG %s args: %v\n", runtime, args)
	}
	cmd := exec.Command(runtime, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "❌ %s execution failed: %v\n", runtime, err)
		return 1
	}
	return 0
}

// handleContainerInspect implements `container inspect <name> [-f '<tmpl>']`.
// Args are forwarded to the underlying runtime so flag compatibility is preserved.
func handleContainerInspect(args []string) {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showInspectHelp()
			return
		}
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "❌ Error: container name required")
		showInspectHelp()
		os.Exit(1)
	}
	runtime, err := selectRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	cmdArgs := append([]string{"inspect"}, args...)
	os.Exit(runPassthrough(runtime, cmdArgs...))
}

// handleContainerNetwork dispatches `container network <create|list|inspect|rm>`.
func handleContainerNetwork(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		showNetworkHelp()
		return
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "create":
		networkCreate(rest)
	case "list", "ls":
		networkList(rest)
	case "inspect":
		networkInspect(rest)
	case "rm", "remove":
		networkRm(rest)
	case "--help", "-h":
		showNetworkHelp()
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown network subcommand: %s\n", sub)
		showNetworkHelp()
		os.Exit(1)
	}
}

// networkExists returns true if the named network is present on the runtime.
func networkExists(runtime, name string) bool {
	cmd := exec.Command(runtime, "network", "inspect", name)
	return cmd.Run() == nil
}

// networkCreate creates a bridge network. Idempotent: returns success with an
// informational message if the network already exists.
func networkCreate(args []string) {
	var name, driver, subnet, gateway string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--driver":
			if i+1 < len(args) {
				driver = args[i+1]
				i++
			}
		case "--subnet":
			if i+1 < len(args) {
				subnet = args[i+1]
				i++
			}
		case "--gateway":
			if i+1 < len(args) {
				gateway = args[i+1]
				i++
			}
		case "--help", "-h":
			showNetworkHelp()
			return
		default:
			if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(os.Stderr, "❌ Unknown flag: %s\n", args[i])
				os.Exit(1)
			}
			if name == "" {
				name = args[i]
			}
		}
	}
	if name == "" {
		fmt.Fprintln(os.Stderr, "❌ Error: network name required")
		showNetworkHelp()
		os.Exit(1)
	}
	runtime, err := selectRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	if networkExists(runtime, name) {
		fmt.Printf("ℹ️  Network '%s' already exists (no action taken)\n", name)
		return
	}
	cmdArgs := []string{"network", "create"}
	if driver != "" {
		cmdArgs = append(cmdArgs, "--driver", driver)
	}
	if subnet != "" {
		cmdArgs = append(cmdArgs, "--subnet", subnet)
	}
	if gateway != "" {
		cmdArgs = append(cmdArgs, "--gateway", gateway)
	}
	cmdArgs = append(cmdArgs, name)
	os.Exit(runPassthrough(runtime, cmdArgs...))
}

// networkList lists networks from the available runtime.
func networkList(args []string) {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			showNetworkHelp()
			return
		}
	}
	runtime, err := selectRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	cmdArgs := append([]string{"network", "ls"}, args...)
	os.Exit(runPassthrough(runtime, cmdArgs...))
}

// networkInspect returns the runtime's inspect output, forwarding -f/--format.
func networkInspect(args []string) {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			showNetworkHelp()
			return
		}
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "❌ Error: network name required")
		os.Exit(1)
	}
	runtime, err := selectRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	cmdArgs := append([]string{"network", "inspect"}, args...)
	os.Exit(runPassthrough(runtime, cmdArgs...))
}

// networkRm removes one or more networks.
// Rootless Podman can emit a harmless "permission denied" warning on network rm
// even when the network is in fact removed; we post-verify with network inspect
// and treat successful removal as success regardless of the warning.
func networkRm(args []string) {
	var names []string
	for _, a := range args {
		if a == "--help" || a == "-h" {
			showNetworkHelp()
			return
		}
		if strings.HasPrefix(a, "-") {
			fmt.Fprintf(os.Stderr, "❌ Unknown flag: %s\n", a)
			os.Exit(1)
		}
		names = append(names, a)
	}
	if len(names) == 0 {
		fmt.Fprintln(os.Stderr, "❌ Error: at least one network name required")
		os.Exit(1)
	}
	runtime, err := selectRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	exitCode := 0
	for _, name := range names {
		existedBefore := networkExists(runtime, name)
		cmdArgs := []string{"network", "rm", name}
		if debugMode {
			fmt.Fprintf(os.Stderr, "🔍 DEBUG %s args: %v\n", runtime, cmdArgs)
		}
		cmd := exec.Command(runtime, cmdArgs...)
		output, runErr := cmd.CombinedOutput()
		if runErr == nil {
			fmt.Print(string(output))
			continue
		}
		// Rootless Podman caveat: `network rm` can emit a harmless warning
		// ("rootless netns: kill network process: permission denied") while
		// actually removing the network. Only treat the error as success
		// when the network existed before and is gone now.
		if existedBefore && !networkExists(runtime, name) {
			fmt.Fprintf(os.Stderr, "⚠️  %s", string(output))
			continue
		}
		// Real failure: surface runtime output and exit code.
		fmt.Fprint(os.Stderr, string(output))
		if ee, ok := runErr.(*exec.ExitError); ok && ee.ExitCode() != 0 {
			exitCode = ee.ExitCode()
		} else {
			exitCode = 1
		}
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

// handleContainerVolume dispatches `container volume <create|list|inspect|rm|prune>`.
func handleContainerVolume(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		showVolumeHelp()
		return
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "create":
		volumeCreate(rest)
	case "list", "ls":
		volumeList(rest)
	case "inspect":
		volumeInspect(rest)
	case "rm", "remove":
		volumeRm(rest)
	case "prune":
		volumePrune(rest)
	case "--help", "-h":
		showVolumeHelp()
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown volume subcommand: %s\n", sub)
		showVolumeHelp()
		os.Exit(1)
	}
}

// volumeExists returns true if the named volume is present on the runtime.
func volumeExists(runtime, name string) bool {
	cmd := exec.Command(runtime, "volume", "inspect", name)
	return cmd.Run() == nil
}

// volumeCreate creates a named volume. Idempotent on pre-existing volumes.
func volumeCreate(args []string) {
	var name, driver string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--driver":
			if i+1 < len(args) {
				driver = args[i+1]
				i++
			}
		case "--help", "-h":
			showVolumeHelp()
			return
		default:
			if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(os.Stderr, "❌ Unknown flag: %s\n", args[i])
				os.Exit(1)
			}
			if name == "" {
				name = args[i]
			}
		}
	}
	if name == "" {
		fmt.Fprintln(os.Stderr, "❌ Error: volume name required")
		showVolumeHelp()
		os.Exit(1)
	}
	runtime, err := selectRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	if volumeExists(runtime, name) {
		fmt.Printf("ℹ️  Volume '%s' already exists (no action taken)\n", name)
		return
	}
	cmdArgs := []string{"volume", "create"}
	if driver != "" {
		cmdArgs = append(cmdArgs, "--driver", driver)
	}
	cmdArgs = append(cmdArgs, name)
	os.Exit(runPassthrough(runtime, cmdArgs...))
}

func volumeList(args []string) {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			showVolumeHelp()
			return
		}
	}
	runtime, err := selectRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	cmdArgs := append([]string{"volume", "ls"}, args...)
	os.Exit(runPassthrough(runtime, cmdArgs...))
}

func volumeInspect(args []string) {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			showVolumeHelp()
			return
		}
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "❌ Error: volume name required")
		os.Exit(1)
	}
	runtime, err := selectRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	cmdArgs := append([]string{"volume", "inspect"}, args...)
	os.Exit(runPassthrough(runtime, cmdArgs...))
}

func volumeRm(args []string) {
	var names []string
	for _, a := range args {
		if a == "--help" || a == "-h" {
			showVolumeHelp()
			return
		}
		if strings.HasPrefix(a, "-") {
			fmt.Fprintf(os.Stderr, "❌ Unknown flag: %s\n", a)
			os.Exit(1)
		}
		names = append(names, a)
	}
	if len(names) == 0 {
		fmt.Fprintln(os.Stderr, "❌ Error: at least one volume name required")
		os.Exit(1)
	}
	runtime, err := selectRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	cmdArgs := append([]string{"volume", "rm"}, names...)
	os.Exit(runPassthrough(runtime, cmdArgs...))
}

func volumePrune(args []string) {
	force := false
	for _, a := range args {
		switch a {
		case "--force", "-f":
			force = true
		case "--help", "-h":
			showVolumeHelp()
			return
		default:
			fmt.Fprintf(os.Stderr, "❌ Unknown flag: %s\n", a)
			os.Exit(1)
		}
	}
	runtime, err := selectRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	cmdArgs := []string{"volume", "prune"}
	if force {
		cmdArgs = append(cmdArgs, "--force")
	}
	os.Exit(runPassthrough(runtime, cmdArgs...))
}

func showInspectHelp() {
	fmt.Println("Usage: portunix container inspect [OPTIONS] <container-name> [<container-name>...]")
	fmt.Println()
	fmt.Println("🔎 INSPECT CONTAINER")
	fmt.Println()
	fmt.Println("Return low-level information on the given container(s) from the")
	fmt.Println("automatically selected runtime (Podman first, Docker fallback).")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -f, --format <tmpl>   Go template for selective output (runtime semantics)")
	fmt.Println("  -h, --help            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container inspect my-container")
	fmt.Println("  portunix container inspect my-container -f '{{.NetworkSettings.Networks}}'")
	fmt.Println("  portunix container inspect my-container --format '{{.Config.Env}}'")
}

func showNetworkHelp() {
	fmt.Println("Usage: portunix container network <subcommand> [options]")
	fmt.Println()
	fmt.Println("🌐 MANAGE CONTAINER NETWORKS")
	fmt.Println()
	fmt.Println("Universal network management that auto-selects Podman or Docker.")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  create <name> [--driver <drv>] [--subnet <CIDR>] [--gateway <IP>]")
	fmt.Println("                  Create a network (idempotent — existing network is a no-op)")
	fmt.Println("  list            List available networks")
	fmt.Println("  inspect <name> [-f '<tmpl>']")
	fmt.Println("                  Show low-level network information")
	fmt.Println("  rm <name>...    Remove one or more networks")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container network create portunix-odoo-net")
	fmt.Println("  portunix container network create my-net --driver bridge --subnet 10.88.0.0/16")
	fmt.Println("  portunix container network list")
	fmt.Println("  portunix container network inspect portunix-odoo-net -f '{{.Subnets}}'")
	fmt.Println("  portunix container network rm portunix-odoo-net")
}

func showVolumeHelp() {
	fmt.Println("Usage: portunix container volume <subcommand> [options]")
	fmt.Println()
	fmt.Println("📦 MANAGE CONTAINER VOLUMES")
	fmt.Println()
	fmt.Println("Universal volume management that auto-selects Podman or Docker.")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  create <name> [--driver <drv>]")
	fmt.Println("                  Create a named volume (idempotent)")
	fmt.Println("  list            List available volumes")
	fmt.Println("  inspect <name> [-f '<tmpl>']")
	fmt.Println("                  Show low-level volume information")
	fmt.Println("  rm <name>...    Remove one or more volumes")
	fmt.Println("  prune [--force] Remove all unused volumes")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container volume create odoo-data")
	fmt.Println("  portunix container volume list")
	fmt.Println("  portunix container volume inspect odoo-data")
	fmt.Println("  portunix container volume rm odoo-data")
	fmt.Println("  portunix container volume prune --force")
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

func showCpHelp() {
	fmt.Println("Usage: portunix container cp <source> <destination>")
	fmt.Println()
	fmt.Println("📁 COPY FILES BETWEEN CONTAINER AND HOST")
	fmt.Println()
	fmt.Println("Copy files or directories between a container and the local filesystem.")
	fmt.Println()
	fmt.Println("🌟 UNIVERSAL OPERATION:")
	fmt.Println("  ✅ Works with both Docker and Podman containers")
	fmt.Println("  ✅ Automatic runtime detection")
	fmt.Println("  ✅ Supports copying in both directions")
	fmt.Println("  ✅ Preserves file permissions")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <source>        Source path (local file or container:path)")
	fmt.Println("  <destination>   Destination path (local file or container:path)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container cp ./config.json mycontainer:/app/config.json")
	fmt.Println("  portunix container cp mycontainer:/var/log/app.log ./logs/")
	fmt.Println("  portunix container cp ./scripts/ mycontainer:/opt/scripts/")
}

func showExecHelp() {
	fmt.Println("Usage: portunix container exec <container-name> <command> [args...]")
	fmt.Println()
	fmt.Println("🔧 EXECUTE COMMAND IN CONTAINER")
	fmt.Println()
	fmt.Println("Run a command inside a running container.")
	fmt.Println()
	fmt.Println("🌟 UNIVERSAL OPERATION:")
	fmt.Println("  ✅ Works with both Docker and Podman containers")
	fmt.Println("  ✅ Automatic runtime detection")
	fmt.Println("  ✅ Supports interactive commands")
	fmt.Println("  ✅ Pass-through of command arguments")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <container-name>   Name or ID of the container")
	fmt.Println("  <command>          Command to execute")
	fmt.Println("  [args...]          Optional arguments for the command")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help         Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container exec my-container bash")
	fmt.Println("  portunix container exec my-container ls -la /app")
	fmt.Println("  portunix container exec web-server cat /etc/nginx/nginx.conf")
	fmt.Println("  portunix container exec python-dev python --version")
}

func showListHelp() {
	fmt.Println("Usage: portunix container list [OPTIONS]")
	fmt.Println()
	fmt.Println("📋 LIST CONTAINERS")
	fmt.Println()
	fmt.Println("Display containers from all available runtimes.")
	fmt.Println()
	fmt.Println("🌟 UNIVERSAL OPERATION:")
	fmt.Println("  ✅ Shows containers from both Docker and Podman")
	fmt.Println("  ✅ Automatic runtime detection")
	fmt.Println("  ✅ Unified output format")
	fmt.Println("  ✅ Shows running and stopped containers")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container list")
}

func showInfoHelp() {
	fmt.Println("Usage: portunix container info")
	fmt.Println()
	fmt.Println("ℹ️ CONTAINER RUNTIME INFORMATION")
	fmt.Println()
	fmt.Println("Display information about available container runtimes.")
	fmt.Println()
	fmt.Println("🌟 DISPLAYS:")
	fmt.Println("  ✅ Docker availability and version")
	fmt.Println("  ✅ Podman availability and version")
	fmt.Println("  ✅ Runtime status and configuration")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix container info")
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
