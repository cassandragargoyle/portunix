package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"portunix.cz/app/podman"

	"github.com/spf13/cobra"
)

// podmanRunInContainerCmd represents the podman run-in-container command
var podmanRunInContainerCmd = &cobra.Command{
	Use:   "run-in-container [installation-type]",
	Short: "Run Portunix installation inside a Podman container with SSH enabled.",
	Long: `Run Portunix installation commands inside a Podman container. This command creates
a new Podman container with the specified base image, installs the requested software,
and sets up SSH access for interactive development.

Key Podman features:
- Rootless containers by default (enhanced security)
- Daemonless operation (no background daemon required)
- OCI-compatible with Docker images
- Pod support for Kubernetes-style container grouping
- Better default security posture
- No privileged daemon access required

Key features:
- Flexible base image selection (Ubuntu, Alpine, CentOS, Debian, etc.)
- Automatic package manager detection and adaptation
- SSH server setup with generated credentials
- Cache directory mounting for persistent downloads
- Volume mounting for file sharing
- Cross-platform support (Linux containers on Windows/Linux hosts)

Available installation types:
  default - Installs Python, Java, and VSCode (recommended, used if no type specified)
  empty   - Creates clean container without installing packages
  python  - Installs Python only
  java    - Installs Java only
  vscode  - Installs VSCode only

Examples:
  portunix podman run-in-container              # Uses default (Python + Java + VSCode)
  portunix podman run-in-container default
  portunix podman run-in-container python --image alpine:3.18
  portunix podman run-in-container java --image ubuntu:20.04 --rootless
  portunix podman run-in-container empty --image debian:bullseye
  portunix podman run-in-container default --image rockylinux:9
  portunix podman run-in-container default --name my-dev-env --keep-running

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
3. Creates rootless container with proper volume and port mappings
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
  --image string       Base container image (default: ubuntu:22.04)
  --name string        Custom container name
  --port strings       Additional port mappings (format: host:container)
  --volume strings     Additional volume mappings (format: host:container)
  --env strings        Environment variables (format: KEY=value)
  --cache-path string  Custom cache directory path (default: .cache)
  --no-cache          Disable cache directory mounting
  --ssh-enabled       Enable SSH server (default: true)
  --keep-running      Keep container running after installation
  --disposable        Auto-remove container when stopped
  --privileged        Run container with elevated privileges (disables rootless)
  --rootless          Run in rootless mode (default: true, enhanced security)
  --pod string        Run container in specified pod (Podman-specific)
  --network string    Network to use (default: bridge)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Default installation type if not provided
		installationType := "default"
		if len(args) > 0 {
			installationType = args[0]
		}
		
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
			fmt.Printf("Invalid installation type: %s\n", installationType)
			fmt.Printf("Valid types: %s\n", strings.Join(validTypes, ", "))
			return
		}
		
		// Parse flags
		image, _ := cmd.Flags().GetString("image")
		name, _ := cmd.Flags().GetString("name")
		ports, _ := cmd.Flags().GetStringSlice("port")
		volumes, _ := cmd.Flags().GetStringSlice("volume")
		envs, _ := cmd.Flags().GetStringSlice("env")
		cachePath, _ := cmd.Flags().GetString("cache-path")
		noCache, _ := cmd.Flags().GetBool("no-cache")
		sshEnabled, _ := cmd.Flags().GetBool("ssh-enabled")
		keepRunning, _ := cmd.Flags().GetBool("keep-running")
		disposable, _ := cmd.Flags().GetBool("disposable")
		privileged, _ := cmd.Flags().GetBool("privileged")
		rootless, _ := cmd.Flags().GetBool("rootless")
		pod, _ := cmd.Flags().GetString("pod")
		network, _ := cmd.Flags().GetString("network")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		autoInstall, _ := cmd.Flags().GetBool("auto-install")
		
		// Check if Podman is available (skip in dry-run mode)
		if !dryRun {
			if err := checkPodmanWithInstallPrompt(autoInstall); err != nil {
				return
			}
		}
		
		// Create Podman configuration
		config := podman.PodmanConfig{
			Image:             image,
			ContainerName:     name,
			Ports:             ports,
			Volumes:           volumes,
			Environment:       envs,
			EnableSSH:         sshEnabled,
			KeepRunning:       keepRunning,
			Disposable:        disposable,
			Privileged:        privileged,
			Network:           network,
			CacheShared:       !noCache,
			CachePath:         cachePath,
			InstallationType:  installationType,
			DryRun:            dryRun,
			AutoInstallPodman: autoInstall,
			Rootless:          rootless,
			Pod:               pod,
		}
		
		// Run container
		err := podman.RunInContainer(config)
		if err != nil {
			fmt.Printf("Error running container: %v\n", err)
			return
		}
		
		if !dryRun {
			fmt.Printf("\n🎉 Container setup completed successfully!\n")
			if sshEnabled {
				fmt.Printf("💡 You can now SSH into the container using the credentials shown above.\n")
			}
			if rootless {
				fmt.Printf("🔒 Running in rootless mode for enhanced security.\n")
			}
		}
	},
}

// checkPodmanWithInstallPrompt checks Podman availability and offers installation if not found
func checkPodmanWithInstallPrompt(autoInstall bool) error {
	if err := podman.CheckPodmanAvailableWithInstall(autoInstall); err != nil {
		// If auto-install flag was used, don't prompt - show helpful error
		if autoInstall {
			fmt.Printf("❌ Auto-installation failed: %v\n\n", err)
			
			// Check if it's a PATH issue
			if strings.Contains(err.Error(), "executable file not found in $PATH") {
				fmt.Println("⚠️  This might be a PATH refresh issue.")
				fmt.Println("📋 Try:")
				fmt.Println("   1. Close and reopen your terminal")
				fmt.Println("   2. Run: source ~/.bashrc (Linux)")
				fmt.Println("   3. Then try: portunix podman run-in-container")
				fmt.Println()
				fmt.Println("🔍 To verify: podman --version")
			} else {
				fmt.Println("💡 Alternative options:")
				fmt.Println("   • Manual install: portunix install podman -y")
				fmt.Println("   • Test mode: portunix podman run-in-container --dry-run")
				fmt.Println("   • Installation guide: https://podman.io/getting-started/installation")
			}
			return fmt.Errorf("podman not available")
		}
		
		// Show error and offer installation
		fmt.Printf("❌ Podman is not available: %v\n\n", err)
		
		fmt.Println("🔧 Would you like to install Podman now?")
		fmt.Println("   This will:")
		fmt.Println("   • Detect your operating system automatically")
		fmt.Println("   • Install the appropriate Podman package")
		fmt.Println("   • Configure rootless containers for enhanced security")
		fmt.Println("   • Optimize storage location")
		fmt.Println()
		
		fmt.Print("Install Podman? [Y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		if response == "" || response == "y" || response == "yes" {
			fmt.Println("\n🚀 Starting Podman installation...")
			if installErr := podman.InstallPodman(true); installErr != nil {
				fmt.Printf("❌ Installation failed: %v\n", installErr)
				fmt.Println("\n💡 You can try manual installation:")
				fmt.Println("   • Visit: https://podman.io/getting-started/installation")
				fmt.Printf("   • Or use: portunix install podman -y\n")
				return fmt.Errorf("podman installation failed")
			}
			
			// Verify installation
			if verifyErr := podman.CheckPodmanAvailableWithInstall(false); verifyErr != nil {
				fmt.Printf("⚠️  Installation completed but Podman is not yet accessible.\n")
				fmt.Println("   This is usually because system PATH needs to be refreshed.")
				fmt.Println()
				fmt.Println("📋 To resolve this:")
				fmt.Println("   1. Close and reopen your terminal window")
				fmt.Println("   2. Or run: source ~/.bashrc (Linux) / restart terminal (Windows)")
				fmt.Println("   3. Then try: portunix podman run-in-container")
				fmt.Println()
				fmt.Println("🔍 To verify Podman is working: podman --version")
				fmt.Println()
				fmt.Printf("💡 You can also test without installation: portunix podman run-in-container --dry-run\n")
				return fmt.Errorf("podman needs PATH refresh")
			}
			
			fmt.Println("✅ Podman installed successfully!")
			fmt.Println("🔒 Rootless containers are ready for enhanced security.")
			return nil
		} else {
			fmt.Println("\n❌ Installation cancelled.")
			fmt.Println("\n💡 To proceed, you can:")
			fmt.Printf("   • Install manually: portunix install podman -y\n")
			fmt.Printf("   • Use auto-install: portunix podman run-in-container --auto-install\n")
			fmt.Printf("   • Test without Podman: portunix podman run-in-container --dry-run\n")
			return fmt.Errorf("user declined installation")
		}
	}
	
	return nil
}

func init() {
	podmanCmd.AddCommand(podmanRunInContainerCmd)
	
	// Add flags
	podmanRunInContainerCmd.Flags().String("image", "ubuntu:22.04", "Base container image to use")
	podmanRunInContainerCmd.Flags().String("name", "", "Custom container name")
	podmanRunInContainerCmd.Flags().StringSlice("port", []string{}, "Additional port mappings (format: host:container)")
	podmanRunInContainerCmd.Flags().StringSlice("volume", []string{}, "Additional volume mappings (format: host:container)")
	podmanRunInContainerCmd.Flags().StringSlice("env", []string{}, "Environment variables (format: KEY=value)")
	podmanRunInContainerCmd.Flags().String("cache-path", "", "Custom cache directory path (default: .cache)")
	podmanRunInContainerCmd.Flags().Bool("no-cache", false, "Disable cache directory mounting")
	podmanRunInContainerCmd.Flags().Bool("ssh-enabled", true, "Enable SSH server setup")
	podmanRunInContainerCmd.Flags().Bool("keep-running", false, "Keep container running after installation")
	podmanRunInContainerCmd.Flags().Bool("disposable", false, "Auto-remove container when stopped")
	podmanRunInContainerCmd.Flags().Bool("privileged", false, "Run container with elevated privileges (disables rootless)")
	podmanRunInContainerCmd.Flags().Bool("rootless", true, "Run in rootless mode (enhanced security)")
	podmanRunInContainerCmd.Flags().String("pod", "", "Run container in specified pod (Podman-specific)")
	podmanRunInContainerCmd.Flags().String("network", "", "Network to use")
	podmanRunInContainerCmd.Flags().Bool("dry-run", false, "Show what would be executed without running commands")
	podmanRunInContainerCmd.Flags().Bool("auto-install", false, "Automatically install Podman if not available")
}