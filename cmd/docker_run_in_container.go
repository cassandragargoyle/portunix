package cmd

import (
	"fmt"
	"strings"

	"portunix.cz/app/docker"

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
  vscode  - Installs VSCode only

Examples:
  portunix docker run-in-container default
  portunix docker run-in-container python --image alpine:3.18
  portunix docker run-in-container java --image ubuntu:20.04
  portunix docker run-in-container empty --image debian:bullseye
  portunix docker run-in-container default --image rockylinux:9
  portunix docker run-in-container default --name my-dev-env --keep-running

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

Flags:
  --image string       Base Docker image (default: ubuntu:22.04)
  --name string        Custom container name
  --port strings       Additional port mappings (format: host:container)
  --volume strings     Additional volume mappings (format: host:container)
  --env strings        Environment variables (format: KEY=value)
  --cache-path string  Custom cache directory path (default: .cache)
  --no-cache          Disable cache directory mounting
  --ssh-enabled       Enable SSH server (default: true)
  --keep-running      Keep container running after installation
  --disposable        Auto-remove container when stopped
  --privileged        Run container with elevated privileges
  --network string    Docker network to use (default: bridge)`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		installationType := args[0]
		
		// Validate installation type
		validTypes := []string{"default", "empty", "python", "java", "vscode"}
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
		err := docker.RunInContainer(config)
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
	dockerRunInContainerCmd.Flags().StringSlice("port", []string{}, "Additional port mappings (format: host:container)")
	dockerRunInContainerCmd.Flags().StringSlice("volume", []string{}, "Additional volume mappings (format: host:container)")
	dockerRunInContainerCmd.Flags().StringSlice("env", []string{}, "Environment variables (format: KEY=value)")
	dockerRunInContainerCmd.Flags().String("cache-path", "", "Custom cache directory path (default: .cache)")
	dockerRunInContainerCmd.Flags().Bool("no-cache", false, "Disable cache directory mounting")
	dockerRunInContainerCmd.Flags().Bool("ssh-enabled", true, "Enable SSH server")
	dockerRunInContainerCmd.Flags().Bool("keep-running", false, "Keep container running after installation")
	dockerRunInContainerCmd.Flags().Bool("disposable", false, "Auto-remove container when stopped")
	dockerRunInContainerCmd.Flags().Bool("privileged", false, "Run container with elevated privileges")
	dockerRunInContainerCmd.Flags().String("network", "", "Docker network to use (default: bridge)")
	dockerRunInContainerCmd.Flags().Bool("dry-run", false, "Show what would be executed without running Docker commands")
	dockerRunInContainerCmd.Flags().Bool("check-requirements", false, "Check system requirements without executing")
	dockerRunInContainerCmd.Flags().Bool("auto-install-docker", false, "Automatically install Docker if not available")
}