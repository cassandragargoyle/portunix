package cmd

import (
	"fmt"
	"strings"

	"portunix.ai/app/docker"
	"portunix.ai/app/install"

	"github.com/spf13/cobra"
)

// dockerRunInContainerCmd represents the docker run-in-container command
var dockerRunInContainerCmd = &cobra.Command{
	Use:   "run-in-container [installation-type]",
	Short: "Run Portunix installation inside a Docker container with SSH enabled.",
	Long: `Run Portunix installation commands inside a Docker container. This command creates
a new Docker container with the specified base image, installs the requested software,
and sets up SSH access for interactive development.

Key features:
- Flexible base image selection (Ubuntu, Alpine, CentOS, Debian, etc.)
- Automatic package manager detection and adaptation
- SSH server setup with generated credentials
- Cache directory mounting for persistent downloads
- Volume mounting for file sharing
- Cross-platform support (Linux containers on Windows/Linux hosts)

Available installation types:
  default - Installs Python, Java, and VSCode (recommended)
  empty   - Creates clean container without installing packages
  python  - Installs Python only
  java    - Installs Java only
  go      - Installs Go development environment
  vscode  - Installs VSCode only

Examples:
  # Basic usage
  portunix docker run-in-container default
  
  # Python team use case (Issue #028) - now works!
  portunix docker run-in-container go --name mycontainer -v "$(pwd):/workspace" --keep-running
  
  # Complex setup with multiple parameters
  portunix docker run-in-container python \
    --name python-dev \
    -v "/host/project:/workspace" \
    -v "/host/cache:/cache" \
    -p "8080:80" \
    -p "3000:3000" \
    -e "NODE_ENV=production" \
    -e "DEBUG=true" \
    --workdir "/workspace" \
    --user "1000:1000" \
    --memory "2g" \
    --cpus "1.5" \
    --keep-running
  
  # Different base images
  portunix docker run-in-container java --image ubuntu:20.04
  portunix docker run-in-container empty --image alpine:3.18

Base image examples:
  --image ubuntu:22.04          # Ubuntu 22.04 LTS (default)
  --image ubuntu:20.04          # Ubuntu 20.04 LTS
  --image alpine:3.18           # Alpine Linux 3.18 (lightweight)
  --image alpine:latest         # Alpine Linux latest
  --image debian:bullseye       # Debian 11 (Bullseye)
  --image centos:8              # CentOS 8
  --image fedora:38             # Fedora 38
  --image rockylinux:9          # Rocky Linux 9 (RHEL-compatible)
  --image rockylinux:8          # Rocky Linux 8 (CentOS replacement)
  
Container workflow:
1. Pulls the specified base image (if not present locally)
2. Detects package manager (apt-get, yum, dnf, apk)
3. Creates container with proper volume and port mappings
4. Installs OpenSSH server and configures SSH access
5. Installs requested software using detected package manager
6. Generates SSH credentials and displays connection information
7. Container remains running for SSH access and development

SSH Setup:
- Automatic OpenSSH server installation and configuration
- Random username and password generation for security
- SSH port exposed on localhost:2222
- Connection information displayed after setup
- SSH connectivity test before completion

File sharing:
- Current directory mounted to /workspace in container
- Cache directory mounted to /portunix-cache (for package persistence)
- Shared downloads between host and container
- Persistent storage for Python packages, Java JDK, etc.

Universal Docker Parameters (same as docker run):
  -v, --volume strings     Volume mappings (format: host:container[:options])
  -p, --port strings       Port mappings (format: host:container)
  -e, --env strings        Environment variables (format: KEY=value)
  --workdir string         Working directory inside container
  --user string            Username or UID (format: <name|uid>[:<group|gid>])
  --privileged             Run container with elevated privileges
  --network string         Docker network to use (default: bridge)
  --memory string          Memory limit (e.g., 1g, 512m)
  --cpus string            Number of CPUs (e.g., 1.5)

Portunix-specific flags:
  --image string           Base Docker image (default: ubuntu:22.04)
  --name string            Custom container name
  --cache-path string      Custom cache directory path (default: .cache)
  --no-cache              Disable cache directory mounting
  --ssh-enabled           Enable SSH server (default: true)
  --keep-running          Keep container running after installation
  --disposable            Auto-remove container when stopped`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		installationType := args[0]

		// Load install config to get valid package names and presets
		installConfig, err := install.LoadInstallConfig()
		if err != nil {
			fmt.Printf("Error: Failed to load install config: %v\n", err)
			return
		}

		// Get valid types from packages and presets
		var validTypes []string

		// Add all package names
		for packageName := range installConfig.Packages {
			validTypes = append(validTypes, packageName)
		}

		// Add all preset names
		for presetName := range installConfig.Presets {
			validTypes = append(validTypes, presetName)
		}

		// Check if the installation type is valid
		isValid := false
		for _, validType := range validTypes {
			if installationType == validType {
				isValid = true
				break
			}
		}

		if !isValid {
			fmt.Printf("Error: Invalid installation type '%s'\n", installationType)
			fmt.Printf("Valid types: %s\n", strings.Join(validTypes, ", "))
			return
		}

		// Parse flags
		image, _ := cmd.Flags().GetString("image")
		if image == "" {
			image = "ubuntu:22.04" // Default base image
		}

		name, _ := cmd.Flags().GetString("name")
		ports, _ := cmd.Flags().GetStringSlice("port")
		volumes, _ := cmd.Flags().GetStringSlice("volume")
		envVars, _ := cmd.Flags().GetStringSlice("env")
		cachePath, _ := cmd.Flags().GetString("cache-path")
		noCache, _ := cmd.Flags().GetBool("no-cache")
		sshEnabled, _ := cmd.Flags().GetBool("ssh-enabled")
		keepRunning, _ := cmd.Flags().GetBool("keep-running")
		disposable, _ := cmd.Flags().GetBool("disposable")
		privileged, _ := cmd.Flags().GetBool("privileged")
		network, _ := cmd.Flags().GetString("network")
		workdir, _ := cmd.Flags().GetString("workdir")
		user, _ := cmd.Flags().GetString("user")
		memory, _ := cmd.Flags().GetString("memory")
		cpus, _ := cmd.Flags().GetString("cpus")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		checkRequirements, _ := cmd.Flags().GetBool("check-requirements")
		autoInstallDocker, _ := cmd.Flags().GetBool("auto-install-docker")

		// Handle check-requirements flag
		if checkRequirements {
			err := docker.CheckRequirements()
			if err != nil {
				fmt.Printf("❌ Requirements check failed: %v\n", err)
				return
			}
			fmt.Println("✅ All requirements are satisfied!")
			return
		}

		// Build Docker configuration
		config := docker.DockerConfig{
			Image:             image,
			ContainerName:     name,
			Ports:             ports,
			Volumes:           volumes,
			Environment:       envVars,
			EnableSSH:         sshEnabled,
			KeepRunning:       keepRunning,
			Disposable:        disposable,
			Privileged:        privileged,
			Network:           network,
			WorkingDir:        workdir,
			User:              user,
			Memory:            memory,
			CPUs:              cpus,
			CacheShared:       !noCache,
			CachePath:         cachePath,
			InstallationType:  installationType,
			DryRun:            dryRun,
			AutoInstallDocker: autoInstallDocker,
		}

		// Add SSH port if enabled
		if sshEnabled {
			config.Ports = append(config.Ports, "2222:22")
		}

		// Run container
		err = docker.RunInContainer(config)
		if err != nil {
			fmt.Printf("Error running container: %v\n", err)
			return
		}
	},
}

func init() {
	dockerCmd.AddCommand(dockerRunInContainerCmd)

	// Add flags
	dockerRunInContainerCmd.Flags().String("image", "", "Base Docker image (default: ubuntu:22.04)")
	dockerRunInContainerCmd.Flags().String("name", "", "Custom container name")
	dockerRunInContainerCmd.Flags().StringSliceP("port", "p", []string{}, "Additional port mappings (format: host:container)")
	dockerRunInContainerCmd.Flags().StringSliceP("volume", "v", []string{}, "Additional volume mappings (format: host:container)")
	dockerRunInContainerCmd.Flags().StringSliceP("env", "e", []string{}, "Environment variables (format: KEY=value)")
	dockerRunInContainerCmd.Flags().String("cache-path", "", "Custom cache directory path (default: .cache)")
	dockerRunInContainerCmd.Flags().Bool("no-cache", false, "Disable cache directory mounting")
	dockerRunInContainerCmd.Flags().Bool("ssh-enabled", true, "Enable SSH server")
	dockerRunInContainerCmd.Flags().Bool("keep-running", false, "Keep container running after installation")
	dockerRunInContainerCmd.Flags().Bool("disposable", false, "Auto-remove container when stopped")
	dockerRunInContainerCmd.Flags().Bool("privileged", false, "Run container with elevated privileges")
	dockerRunInContainerCmd.Flags().String("network", "", "Docker network to use (default: bridge)")
	dockerRunInContainerCmd.Flags().String("workdir", "", "Working directory inside container")
	dockerRunInContainerCmd.Flags().String("user", "", "Username or UID (format: <name|uid>[:<group|gid>])")
	dockerRunInContainerCmd.Flags().String("memory", "", "Memory limit (e.g., 1g, 512m)")
	dockerRunInContainerCmd.Flags().String("cpus", "", "Number of CPUs (e.g., 1.5)")
	dockerRunInContainerCmd.Flags().Bool("dry-run", false, "Show what would be executed without running Docker commands")
	dockerRunInContainerCmd.Flags().Bool("check-requirements", false, "Check system requirements without executing")
	dockerRunInContainerCmd.Flags().Bool("auto-install-docker", false, "Automatically install Docker if not available")
}
