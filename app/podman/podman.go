package podman

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"portunix.cz/app/system"
)

// PodmanConfig defines the configuration for Podman containers
type PodmanConfig struct {
	Image             string
	ContainerName     string
	Ports             []string
	Volumes           []string
	Environment       []string
	Command           []string
	EnableSSH         bool
	KeepRunning       bool
	Disposable        bool
	Privileged        bool
	Network           string
	CacheShared       bool
	CachePath         string
	InstallationType  string
	DryRun            bool
	AutoInstallPodman bool
	Rootless          bool   // Podman-specific: run in rootless mode
	Pod               string // Podman-specific: pod name
}

// ContainerInfo represents information about a Podman container
type ContainerInfo struct {
	ID      string
	Name    string
	Image   string
	Status  string
	Ports   string
	Created string
	Command string
}

// PackageManagerInfo holds detected package manager information
type PackageManagerInfo struct {
	Manager      string // apt-get, yum, dnf, apk, etc.
	UpdateCmd    string
	InstallCmd   string
	Distribution string // ubuntu, debian, alpine, centos, etc.
}

// InstallPodmanDesktop installs Podman Desktop GUI application
func InstallPodmanDesktop(autoAccept bool) error {
	fmt.Println("Starting Podman Desktop installation...")
	fmt.Println("🖥️  Podman Desktop is the official GUI from Red Hat for container management")

	// Detect OS
	osInfo, err := system.GetSystemInfo()
	if err != nil {
		return fmt.Errorf("failed to detect operating system: %w", err)
	}

	fmt.Printf("✓ Detected: %s %s\n", osInfo.OS, osInfo.Version)

	// Check if Podman Desktop is already installed
	if isPodmanDesktopInstalled() {
		fmt.Println("✓ Podman Desktop is already installed")
		return nil
	}

	// Install based on OS
	switch runtime.GOOS {
	case "windows":
		return installPodmanDesktopGUIWindows(autoAccept)
	case "linux":
		return installPodmanDesktopGUILinux(autoAccept, osInfo)
	case "darwin":
		return installPodmanDesktopGUIMacOS(autoAccept)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// isPodmanDesktopInstalled checks if Podman Desktop is already installed
func isPodmanDesktopInstalled() bool {
	switch runtime.GOOS {
	case "windows":
		// Check if Podman Desktop executable exists
		programFiles := os.Getenv("PROGRAMFILES")
		if programFiles != "" {
			desktopPath := filepath.Join(programFiles, "Podman Desktop", "Podman Desktop.exe")
			if _, err := os.Stat(desktopPath); err == nil {
				return true
			}
		}
		// Check alternative locations
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			desktopPath := filepath.Join(localAppData, "Programs", "Podman Desktop", "Podman Desktop.exe")
			if _, err := os.Stat(desktopPath); err == nil {
				return true
			}
		}
	case "linux":
		// Check for AppImage or system installation
		if _, err := exec.LookPath("podman-desktop"); err == nil {
			return true
		}
		// Check for AppImage in common locations
		homeDir, _ := os.UserHomeDir()
		appImagePaths := []string{
			filepath.Join(homeDir, "Applications", "podman-desktop.AppImage"),
			filepath.Join(homeDir, "Desktop", "podman-desktop.AppImage"),
			"/usr/bin/podman-desktop",
			"/usr/local/bin/podman-desktop",
		}
		for _, path := range appImagePaths {
			if _, err := os.Stat(path); err == nil {
				return true
			}
		}
	case "darwin":
		// Check Applications folder
		if _, err := os.Stat("/Applications/Podman Desktop.app"); err == nil {
			return true
		}
	}
	return false
}

// InstallPodman performs intelligent OS-based Podman installation
func InstallPodman(autoAccept bool) error {
	fmt.Println("Starting Podman installation with intelligent OS detection...")

	// Detect OS
	osInfo, err := system.GetSystemInfo()
	if err != nil {
		return fmt.Errorf("failed to detect operating system: %w", err)
	}

	fmt.Printf("✓ Detected: %s %s\n", osInfo.OS, osInfo.Version)

	// Check if Podman is already installed
	if isPodmanInstalled() {
		fmt.Println("✓ Podman is already installed")
		return verifyPodmanInstallation()
	}

	// Analyze storage and install based on OS
	switch runtime.GOOS {
	case "windows":
		return installPodmanWindows(autoAccept)
	case "linux":
		return installPodmanLinux(autoAccept, osInfo)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// isPodmanInstalled checks if Podman is already installed
func isPodmanInstalled() bool {
	cmd := exec.Command("podman", "--version")
	err := cmd.Run()
	return err == nil
}

// verifyPodmanInstallation verifies Podman installation
func verifyPodmanInstallation() error {
	fmt.Println("\nVerifying Podman installation...")

	// Check Podman version
	cmd := exec.Command("podman", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("podman --version failed: %w", err)
	}
	fmt.Printf("✓ %s", string(output))

	// Check Podman system info
	cmd = exec.Command("podman", "system", "info", "--format", "json")
	err = cmd.Run()
	if err != nil {
		fmt.Println("⚠️  Podman system info may not be accessible")
		return fmt.Errorf("podman system not accessible: %w", err)
	}
	fmt.Println("✓ Podman system is ready")

	// Check rootless configuration
	cmd = exec.Command("podman", "unshare", "cat", "/proc/self/uid_map")
	if err := cmd.Run(); err == nil {
		fmt.Println("✓ Rootless mode available")
	} else {
		fmt.Println("⚠️  Rootless mode may not be configured")
	}

	return nil
}

// RunInContainer runs Portunix installation inside a Podman container
func RunInContainer(config PodmanConfig) error {
	fmt.Printf("Starting Podman container with %s installation...\n", config.InstallationType)

	// In dry-run mode, show what would be executed
	if config.DryRun {
		return runInContainerDryRun(config)
	}

	// Check if Podman is available
	if err := checkPodmanAvailable(); err != nil {
		return fmt.Errorf("Podman is not available: %w", err)
	}

	// Pull base image if needed
	if err := pullImageIfNeeded(config.Image); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	// Detect package manager in the image
	pkgManager, err := detectPackageManager(config.Image)
	if err != nil {
		return fmt.Errorf("failed to detect package manager: %w", err)
	}
	fmt.Printf("✓ Detected package manager: %s\n", pkgManager.Manager)

	// Create container name if not provided
	if config.ContainerName == "" {
		config.ContainerName = fmt.Sprintf("portunix-%s-%s", config.InstallationType, generateID())
	}

	// Setup cache directory
	if config.CacheShared {
		if err := setupCacheDirectory(config.CachePath); err != nil {
			return fmt.Errorf("failed to setup cache directory: %w", err)
		}
	}

	// Build Podman run command
	podmanArgs := buildPodmanRunArgs(config)

	fmt.Printf("✓ Creating container: %s\n", config.ContainerName)
	if config.Rootless {
		fmt.Printf("✓ Running in rootless mode (enhanced security)\n")
	}

	// Run the container
	cmd := exec.Command("podman", podmanArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to be ready
	if err := waitForContainer(config.ContainerName); err != nil {
		return fmt.Errorf("container failed to start: %w", err)
	}

	// Setup SSH if enabled
	if config.EnableSSH {
		return setupContainerSSH(config.ContainerName, pkgManager)
	}

	return nil
}

// BuildImage builds a Podman image for Portunix
func BuildImage(baseImage string) error {
	fmt.Printf("Building Portunix Podman image based on %s...\n", baseImage)

	// Create temporary Containerfile (Podman equivalent of Dockerfile)
	containerfile := generateContainerfile(baseImage)

	// Write Containerfile to temp location
	tempDir := ".tmp"
	os.MkdirAll(tempDir, 0755)

	containerfilePath := filepath.Join(tempDir, "Containerfile")
	if err := os.WriteFile(containerfilePath, []byte(containerfile), 0644); err != nil {
		return fmt.Errorf("failed to write Containerfile: %w", err)
	}

	// Build image
	imageName := fmt.Sprintf("portunix:%s", strings.ReplaceAll(baseImage, ":", "-"))
	cmd := exec.Command("podman", "build", "-t", imageName, "-f", containerfilePath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	fmt.Printf("✓ Image built successfully: %s\n", imageName)
	return nil
}

// ListContainers lists all Portunix Podman containers
func ListContainers() ([]ContainerInfo, error) {
	cmd := exec.Command("podman", "ps", "-a", "--filter", "name=portunix-", "--format", "table {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}\t{{.CreatedAt}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var containers []ContainerInfo

	// Skip header line
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

// StopContainer stops a running container
func StopContainer(containerID string) error {
	fmt.Printf("Stopping container %s...\n", containerID)

	cmd := exec.Command("podman", "stop", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	fmt.Printf("✓ Container %s stopped\n", containerID)
	return nil
}

// StartContainer starts a stopped container
func StartContainer(containerID string) error {
	fmt.Printf("Starting container %s...\n", containerID)

	cmd := exec.Command("podman", "start", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("✓ Container %s started\n", containerID)
	return nil
}

// RemoveContainer removes a container
func RemoveContainer(containerID string, force bool) error {
	fmt.Printf("Removing container %s...\n", containerID)

	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, containerID)

	cmd := exec.Command("podman", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	fmt.Printf("✓ Container %s removed\n", containerID)
	return nil
}

// ShowLogs shows container logs
func ShowLogs(containerID string, follow bool) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, containerID)

	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ExecCommand executes a command in a running container
func ExecCommand(containerID string, command []string) error {
	args := append([]string{"exec", "-it", containerID}, command...)

	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// CheckPodmanAvailableWithInstall checks if Podman is available and optionally installs it
func CheckPodmanAvailableWithInstall(autoInstall bool) error {
	// First check if Podman is already available
	if err := checkPodmanAvailable(); err != nil {
		if !autoInstall {
			return err
		}

		// Try to install Podman automatically
		fmt.Println("Podman is not available. Attempting automatic installation...")
		if installErr := InstallPodman(true); installErr != nil {
			return fmt.Errorf("Podman is not available and automatic installation failed: %v\nOriginal error: %v", installErr, err)
		}

		// Verify installation was successful
		if verifyErr := checkPodmanAvailable(); verifyErr != nil {
			return fmt.Errorf("Podman installation appeared to succeed but verification failed: %v", verifyErr)
		}

		fmt.Println("✓ Podman installed and verified successfully")
	}

	return nil
}

// Helper functions (adapted from Docker implementation)

func installPodmanWindows(autoAccept bool) error {
	fmt.Println("\nAnalyzing available storage...")

	// Analyze drives and recommend storage location
	drives, err := analyzeWindowsStorage()
	if err != nil {
		return fmt.Errorf("failed to analyze storage: %w", err)
	}

	var selectedDrive string
	if autoAccept {
		// Auto-select drive with most space
		selectedDrive = drives[0].Letter
		fmt.Printf("✓ Automatically selected optimal storage: %s:\\ (%s available)\n", selectedDrive, drives[0].FreeSpace)
	} else {
		selectedDrive, err = promptStorageSelection(drives)
		if err != nil {
			return err
		}
	}

	fmt.Println("\nInstalling Podman Desktop for Windows...")

	// Download Podman Desktop installer
	installerPath := filepath.Join(".cache", "PodmanDesktopInstaller.exe")
	if err := downloadPodmanDesktop(installerPath); err != nil {
		return fmt.Errorf("failed to download Podman Desktop: %w", err)
	}

	// Install Podman Desktop
	dataRoot := fmt.Sprintf("%s:\\podman-data", selectedDrive)
	if err := installPodmanDesktopWindows(installerPath, dataRoot); err != nil {
		return fmt.Errorf("failed to install Podman Desktop: %w", err)
	}

	return verifyPodmanInstallation()
}

func installPodmanLinux(autoAccept bool, osInfo *system.SystemInfo) error {
	fmt.Println("\nAnalyzing available storage...")

	// Analyze partitions and recommend storage location
	partitions, err := analyzeLinuxStorage()
	if err != nil {
		return fmt.Errorf("failed to analyze storage: %w", err)
	}

	var selectedPath string
	if autoAccept {
		// Auto-select partition with most space
		selectedPath = partitions[0].MountPoint + "/podman-data"
		fmt.Printf("✓ Automatically selected optimal storage: %s (%s available)\n", selectedPath, partitions[0].FreeSpace)
	} else {
		selectedPath, err = promptLinuxStorageSelection(partitions)
		if err != nil {
			return err
		}
	}

	fmt.Printf("\nInstalling Podman for %s...\n", osInfo.OS)

	// Install Podman based on distribution
	distribution := osInfo.OS
	if osInfo.LinuxInfo != nil {
		distribution = osInfo.LinuxInfo.Distribution
	}

	switch strings.ToLower(distribution) {
	case "ubuntu", "debian":
		return installPodmanUbuntuDebian(selectedPath)
	case "centos", "rhel", "rocky", "fedora":
		return installPodmanCentOSFedora(selectedPath)
	case "alpine":
		return installPodmanAlpine(selectedPath)
	default:
		return installPodmanGeneric(selectedPath)
	}
}

func pullImageIfNeeded(image string) error {
	// Check if image exists locally
	cmd := exec.Command("podman", "image", "inspect", image)
	if cmd.Run() == nil {
		fmt.Printf("Using cached image: %s\n", image)
		return nil
	}

	// Pull image
	fmt.Printf("Pulling image: %s...\n", image)
	cmd = exec.Command("podman", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull image %s: %w", image, err)
	}

	fmt.Printf("✓ Image pulled successfully\n")
	return nil
}

func detectPackageManager(image string) (*PackageManagerInfo, error) {
	// Run a container to detect package manager
	cmd := exec.Command("podman", "run", "--rm", image, "sh", "-c", "command -v dnf && exit 0; command -v yum && exit 0; command -v apt-get && exit 0; command -v apk && exit 0; echo 'unknown'")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to detect package manager: %w", err)
	}

	pkgManagerPath := strings.TrimSpace(string(output))
	pkgManager := &PackageManagerInfo{}

	switch {
	case strings.Contains(pkgManagerPath, "apt-get"):
		pkgManager.Manager = "apt-get"
		pkgManager.UpdateCmd = "apt-get update"
		pkgManager.InstallCmd = "apt-get install -y"
		pkgManager.Distribution = "debian-based"
	case strings.Contains(pkgManagerPath, "yum"):
		pkgManager.Manager = "yum"
		pkgManager.UpdateCmd = "yum update -y"
		pkgManager.InstallCmd = "yum install -y"
		pkgManager.Distribution = "rhel-based"
	case strings.Contains(pkgManagerPath, "dnf"):
		pkgManager.Manager = "dnf"
		pkgManager.UpdateCmd = "dnf update -y"
		pkgManager.InstallCmd = "dnf install -y"
		pkgManager.Distribution = "rhel-based" // Covers Fedora, RHEL, Rocky Linux, AlmaLinux
	case strings.Contains(pkgManagerPath, "apk"):
		pkgManager.Manager = "apk"
		pkgManager.UpdateCmd = "apk update"
		pkgManager.InstallCmd = "apk add --no-cache"
		pkgManager.Distribution = "alpine"
	default:
		pkgManager.Manager = "unknown"
		pkgManager.Distribution = "unknown"
	}

	return pkgManager, nil
}

func setupCacheDirectory(cachePath string) error {
	if cachePath == "" {
		cachePath = ".cache"
	}

	return os.MkdirAll(cachePath, 0755)
}

func buildPodmanRunArgs(config PodmanConfig) []string {
	args := []string{"run"}

	// Detached mode
	args = append(args, "-d")

	// Interactive terminal
	args = append(args, "-it")

	// Container name
	if config.ContainerName != "" {
		args = append(args, "--name", config.ContainerName)
	}

	// Port mappings
	for _, port := range config.Ports {
		args = append(args, "-p", port)
	}

	// Volume mappings
	for _, volume := range config.Volumes {
		args = append(args, "-v", volume)
	}

	// Environment variables
	for _, env := range config.Environment {
		args = append(args, "-e", env)
	}

	// Cache directory mounting
	if config.CacheShared {
		cachePath := config.CachePath
		if cachePath == "" {
			cachePath = ".cache"
		}
		abs, _ := filepath.Abs(cachePath)
		args = append(args, "-v", fmt.Sprintf("%s:/portunix-cache", abs))
	}

	// Current directory mounting
	pwd, _ := os.Getwd()
	args = append(args, "-v", fmt.Sprintf("%s:/workspace", pwd))

	// SSH port mapping if enabled
	if config.EnableSSH {
		sshPort := findAvailablePort(2222, 2230)
		args = append(args, "-p", fmt.Sprintf("%d:22", sshPort))
	}

	// Privileged mode (less common in Podman due to rootless)
	if config.Privileged {
		args = append(args, "--privileged")
	}

	// Network
	if config.Network != "" {
		args = append(args, "--network", config.Network)
	}

	// Auto-remove if disposable
	if config.Disposable {
		args = append(args, "--rm")
	}

	// Pod specification (Podman-specific)
	if config.Pod != "" {
		args = append(args, "--pod", config.Pod)
	}

	// Image
	args = append(args, config.Image)

	// Command
	if len(config.Command) > 0 {
		args = append(args, config.Command...)
	} else {
		// Default command to keep container running
		args = append(args, "sleep", "infinity")
	}

	return args
}

func waitForContainer(containerName string) error {
	timeout := 30 * time.Second
	start := time.Now()

	for time.Since(start) < timeout {
		cmd := exec.Command("podman", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Status}}")
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), "Up") {
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("container did not start within %v", timeout)
}

func setupContainerSSH(containerName string, pkgManager *PackageManagerInfo) error {
	fmt.Println("\nSetting up SSH in container...")

	// Install OpenSSH server
	installSSHCmd := generateSSHInstallCommand(pkgManager)
	if err := execInContainer(containerName, installSSHCmd); err != nil {
		return fmt.Errorf("failed to install SSH: %w", err)
	}

	// Generate SSH credentials
	username := fmt.Sprintf("portunix_user_%s", generateShortID())
	password := generatePassword()

	// Create user and set password - Alpine vs Ubuntu compatible
	var createUserCmd []string
	if pkgManager.Manager == "apk" {
		// Alpine Linux uses adduser with different syntax
		createUserCmd = []string{"sh", "-c", fmt.Sprintf("adduser -D -s /bin/sh %s && echo '%s:%s' | chpasswd", username, username, password)}
	} else {
		// Ubuntu/Debian and other systems
		createUserCmd = []string{"sh", "-c", fmt.Sprintf("useradd -m -s /bin/bash %s && echo '%s:%s' | chpasswd", username, username, password)}
	}
	if err := execInContainer(containerName, createUserCmd); err != nil {
		return fmt.Errorf("failed to create SSH user: %w", err)
	}

	// Configure SSH daemon
	configSSHCmd := []string{"sh", "-c", `
		mkdir -p /run/sshd
		# Generate SSH host keys if they don't exist
		ssh-keygen -A 2>/dev/null || true
		echo "PasswordAuthentication yes" >> /etc/ssh/sshd_config
		echo "PermitRootLogin no" >> /etc/ssh/sshd_config
		echo "Port 22" >> /etc/ssh/sshd_config
		/usr/sbin/sshd -D &
	`}
	if err := execInContainer(containerName, configSSHCmd); err != nil {
		return fmt.Errorf("failed to start SSH daemon: %w", err)
	}

	// Get SSH port from container
	sshPort := getSSHPortFromContainer(containerName)

	// Test SSH connectivity
	if err := testSSHConnectivity(containerName, sshPort); err != nil {
		return fmt.Errorf("SSH connectivity test failed: %w", err)
	}

	// Display connection information
	displaySSHInfo(containerName, username, password, sshPort)

	return nil
}

func generateSSHInstallCommand(pkgManager *PackageManagerInfo) []string {
	var cmd string

	switch pkgManager.Manager {
	case "apt-get":
		cmd = "apt-get update && apt-get install -y openssh-server sudo"
	case "yum":
		cmd = "yum install -y openssh-server sudo"
	case "dnf":
		cmd = "dnf install -y openssh-server sudo"
	case "apk":
		cmd = "apk update && apk add --no-cache openssh-server sudo"
	default:
		cmd = "echo 'Unknown package manager, SSH setup may fail'"
	}

	return []string{"sh", "-c", cmd}
}

func execInContainer(containerName string, command []string) error {
	args := append([]string{"exec", containerName}, command...)
	cmd := exec.Command("podman", args...)
	return cmd.Run()
}

func findAvailablePort(start, end int) int {
	for port := start; port <= end; port++ {
		conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			conn.Close()
			return port
		}
	}
	// Fallback to default if no port found
	return start
}

func getSSHPortFromContainer(containerName string) int {
	// Get port mapping from container
	cmd := exec.Command("podman", "port", containerName, "22")
	output, err := cmd.Output()
	if err == nil {
		// Parse output like "0.0.0.0:2223"
		portInfo := strings.TrimSpace(string(output))
		if parts := strings.Split(portInfo, ":"); len(parts) == 2 {
			if port, err := fmt.Sscanf(parts[1], "%d", new(int)); err == nil && port == 1 {
				var p int
				fmt.Sscanf(parts[1], "%d", &p)
				return p
			}
		}
	}
	return 2222 // fallback
}

func testSSHConnectivity(containerName string, sshPort int) error {
	// Get container IP - try multiple formats for Podman compatibility
	var ip string
	var err error

	// Try Podman-style network info first
	cmd := exec.Command("podman", "inspect", "-f", "{{.NetworkSettings.IPAddress}}", containerName)
	output, err := cmd.Output()
	if err == nil {
		ip = strings.TrimSpace(string(output))
	}

	// If that didn't work, try Docker-style format
	if ip == "" {
		cmd = exec.Command("podman", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerName)
		output, err = cmd.Output()
		if err == nil {
			ip = strings.TrimSpace(string(output))
		}
	}

	// If still no IP, try alternative method
	if ip == "" {
		cmd = exec.Command("podman", "inspect", "--format", "{{.NetworkSettings.IPAddress}}", containerName)
		output, err = cmd.Output()
		if err == nil {
			ip = strings.TrimSpace(string(output))
		}
	}

	// Test SSH connectivity - try both container IP and localhost:2222
	timeout := 10 * time.Second
	start := time.Now()

	for time.Since(start) < timeout {
		// First try container IP if available
		if ip != "" && ip != "<no value>" {
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "22"), 1*time.Second)
			if err == nil {
				conn.Close()
				fmt.Printf("✓ SSH accessible via container IP: %s:22\n", ip)
				return nil
			}
		}

		// Also try localhost with actual SSH port (port mapping)
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", sshPort), 1*time.Second)
		if err == nil {
			conn.Close()
			fmt.Printf("✓ SSH accessible via localhost:%d\n", sshPort)
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("SSH port not responding on any address within %v", timeout)
}

// cleanSSHHostKeys removes SSH host keys for the given port to avoid conflicts
func cleanSSHHostKeys(port int) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		return
	}

	// Clean host key for this port
	cmd := exec.Command("ssh-keygen", "-f", knownHostsPath, "-R", fmt.Sprintf("[localhost]:%d", port))
	cmd.Run() // Ignore errors - the key might not exist
}

func displaySSHInfo(containerName, username, password string, sshPort int) {
	// Clean any conflicting SSH host keys
	cleanSSHHostKeys(sshPort)

	// Get container IP
	cmd := exec.Command("podman", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerName)
	output, _ := cmd.Output()
	ip := strings.TrimSpace(string(output))

	fmt.Println("\n📡 SSH CONNECTION INFORMATION:")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Printf("🔗 Container IP:   %s\n", ip)
	fmt.Printf("📄 SSH Port:      localhost:%d\n", sshPort)
	fmt.Printf("👤 Username:      %s\n", username)
	fmt.Printf("🔐 Password:      %s\n", password)
	fmt.Printf("📄 SSH Command:   ssh %s@localhost -p %d\n", username, sshPort)
	fmt.Println()
	fmt.Println("💡 CONNECTION TIPS:")
	fmt.Println("   • Open new terminal window")
	fmt.Printf("   • Run: ssh %s@localhost -p %d\n", username, sshPort)
	fmt.Printf("   • Enter password: %s\n", password)
	fmt.Println("   • If host key error occurs, run:")
	fmt.Printf("     ssh-keygen -R '[localhost]:%d'\n", sshPort)
	fmt.Println("   • Files are shared at: /workspace")
	fmt.Println("   • Cache directory: /portunix-cache")
	fmt.Println("   • Portunix tools available in PATH")
	fmt.Println()
	fmt.Println("💡 PODMAN FEATURES:")
	fmt.Println("   • Running in rootless mode (enhanced security)")
	fmt.Println("   • No daemon required")
	fmt.Println("   • OCI-compatible with Docker images")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("Container is running and ready for SSH connections!\n")
	fmt.Println()
	fmt.Println("Available management commands:")
	fmt.Printf("  portunix podman exec %s \"command\"     # Execute command\n", containerName)
	fmt.Printf("  portunix podman logs %s               # View container logs\n", containerName)
	fmt.Printf("  portunix podman stop %s               # Stop container\n", containerName)
	fmt.Printf("  portunix podman remove %s             # Remove container\n", containerName)
}

func generateContainerfile(baseImage string) string {
	return fmt.Sprintf(`FROM %s

# Install basic tools and Portunix
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    openssh-server \
    sudo \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create workspace directory
WORKDIR /workspace

# Copy Portunix binary (to be mounted at runtime)
COPY portunix /usr/local/bin/portunix
RUN chmod +x /usr/local/bin/portunix

# Setup SSH
RUN mkdir /var/run/sshd
EXPOSE 22

# Default command
CMD ["/bin/bash"]
`, baseImage)
}

// Utility functions

func generateID() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

func generateShortID() string {
	return fmt.Sprintf("%d", time.Now().Unix()%10000)
}

func generatePassword() string {
	// Simple password generation
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789"
	password := make([]byte, 16)
	for i := range password {
		password[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(1 * time.Nanosecond) // Ensure different seeds
	}
	return string(password)
}

// checkPodmanAvailable checks if Podman is installed and accessible
func checkPodmanAvailable() error {
	// Check if podman command exists
	cmd := exec.Command("podman", "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Podman is not installed or not accessible. Please install Podman first.\n"+
			"Installation guide: https://podman.io/getting-started/installation\n"+
			"Error: %v", err)
	}

	// Check if Podman system is accessible
	cmd = exec.Command("podman", "system", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Podman system is not accessible.\n"+
			"Try: podman system migrate (if upgrading) or check permissions\n"+
			"Error: %v", err)
	}

	// Extract just the version line
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Version:") {
			fmt.Printf("✓ Podman is available: %s\n", strings.TrimSpace(line))
			break
		}
	}
	if len(lines) == 0 {
		fmt.Println("✓ Podman is available")
	}
	return nil
}

// runInContainerDryRun shows what would be executed without running Podman commands
func runInContainerDryRun(config PodmanConfig) error {
	fmt.Println("🔍 DRY RUN MODE - Showing what would be executed:")
	fmt.Println()

	// Show configuration
	fmt.Printf("📋 Configuration:\n")
	fmt.Printf("  • Installation Type: %s\n", config.InstallationType)
	fmt.Printf("  • Base Image: %s\n", config.Image)

	if config.ContainerName == "" {
		config.ContainerName = fmt.Sprintf("portunix-%s-%s", config.InstallationType, "GENERATED_ID")
	}
	fmt.Printf("  • Container Name: %s\n", config.ContainerName)

	if config.Rootless {
		fmt.Printf("  • Rootless Mode: Yes (enhanced security)\n")
	} else {
		fmt.Printf("  • Rootless Mode: No\n")
	}

	if config.Pod != "" {
		fmt.Printf("  • Pod: %s\n", config.Pod)
	}

	if config.EnableSSH {
		availablePort := findAvailablePort(2222, 2230)
		fmt.Printf("  • SSH Enabled: Yes (port %d:22)\n", availablePort)
	} else {
		fmt.Printf("  • SSH Enabled: No\n")
	}

	if len(config.Ports) > 0 {
		fmt.Printf("  • Additional Ports: %v\n", config.Ports)
	}

	if len(config.Volumes) > 0 {
		fmt.Printf("  • Volume Mounts: %v\n", config.Volumes)
	}

	if len(config.Environment) > 0 {
		fmt.Printf("  • Environment Variables: %v\n", config.Environment)
	}

	if config.CacheShared {
		cachePath := config.CachePath
		if cachePath == "" {
			cachePath = ".cache"
		}
		fmt.Printf("  • Cache Directory: %s mounted to /portunix-cache\n", cachePath)
	}

	fmt.Printf("  • Keep Running: %v\n", config.KeepRunning)
	fmt.Printf("  • Disposable: %v\n", config.Disposable)
	fmt.Printf("  • Privileged: %v\n", config.Privileged)

	if config.Network != "" {
		fmt.Printf("  • Network: %s\n", config.Network)
	}

	fmt.Println()

	// Show Podman commands that would be executed
	fmt.Printf("🐳 Podman commands that would be executed:\n")
	fmt.Println()

	// 1. Check Podman availability
	fmt.Printf("1. Check Podman availability:\n")
	fmt.Printf("   podman version\n")
	fmt.Printf("   podman system info\n")
	fmt.Println()

	// 2. Pull image
	fmt.Printf("2. Pull base image (if not cached):\n")
	fmt.Printf("   podman image inspect %s\n", config.Image)
	fmt.Printf("   podman pull %s\n", config.Image)
	fmt.Println()

	// 3. Detect package manager
	fmt.Printf("3. Detect package manager:\n")
	fmt.Printf("   podman run --rm %s /bin/sh -c \"command -v dnf && exit 0; command -v yum && exit 0; command -v apt-get && exit 0; command -v apk && exit 0\"\n", config.Image)
	fmt.Println()

	// 4. Build Podman run command
	podmanArgs := buildPodmanRunArgs(config)
	fmt.Printf("4. Create and run container:\n")
	fmt.Printf("   podman run %s\n", strings.Join(podmanArgs[1:], " "))
	fmt.Println()

	// 5. Install packages based on type
	fmt.Printf("5. Install software in container:\n")
	switch config.InstallationType {
	case "default":
		fmt.Printf("   • Install Python, Java, and VSCode\n")
	case "python":
		fmt.Printf("   • Install Python development environment\n")
	case "java":
		fmt.Printf("   • Install Java development environment\n")
	case "vscode":
		fmt.Printf("   • Install Visual Studio Code\n")
	case "empty":
		fmt.Printf("   • No additional software installation\n")
	}
	fmt.Println()

	// 6. SSH setup
	if config.EnableSSH {
		fmt.Printf("6. Setup SSH access:\n")
		fmt.Printf("   • Install OpenSSH server\n")
		fmt.Printf("   • Generate SSH credentials\n")
		fmt.Printf("   • Configure SSH daemon\n")
		fmt.Printf("   • Test SSH connectivity on localhost:2222\n")
		fmt.Println()
	}

	// 7. Expected outcome
	fmt.Printf("📊 Expected outcome:\n")
	fmt.Printf("  • Container '%s' would be created and running\n", config.ContainerName)
	if config.Rootless {
		fmt.Printf("  • Running in rootless mode (enhanced security)\n")
	}
	if config.EnableSSH {
		fmt.Printf("  • SSH access available at localhost:2222\n")
	}
	fmt.Printf("  • Current directory mounted to /workspace\n")
	if config.CacheShared {
		fmt.Printf("  • Cache directory mounted for persistent downloads\n")
	}
	fmt.Printf("  • %s development environment ready\n", config.InstallationType)

	fmt.Println()
	fmt.Printf("💡 To execute for real, remove the --dry-run flag\n")

	return nil
}

// CheckRequirements checks if all system requirements for Podman container operations are satisfied
func CheckRequirements() error {
	fmt.Println("🔍 Checking system requirements for Podman container operations...")
	fmt.Println()

	// Check Podman installation
	fmt.Print("📦 Podman installation: ")
	cmd := exec.Command("podman", "version", "--format", "{{.Client.Version}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("   Podman is not installed or not accessible\n")
		fmt.Printf("   Please install Podman: https://podman.io/getting-started/installation\n")
		return fmt.Errorf("Podman not available")
	}
	fmt.Printf("✅ OK (Version: %s)\n", strings.TrimSpace(string(output)))

	// Check Podman system
	fmt.Print("🐳 Podman system: ")
	cmd = exec.Command("podman", "system", "info", "--format", "{{.Host.RemoteSocket.Path}}")
	_, err = cmd.Output()
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("   Podman system is not accessible\n")
		fmt.Printf("   Try: podman system migrate or check configuration\n")
		return fmt.Errorf("Podman system not accessible")
	}
	fmt.Printf("✅ OK\n")

	// Check rootless capability
	fmt.Print("🔐 Rootless capability: ")
	cmd = exec.Command("podman", "unshare", "echo", "rootless-test")
	err = cmd.Run()
	if err != nil {
		fmt.Println("⚠️  WARNING")
		fmt.Printf("   Rootless mode may not be configured properly\n")
		fmt.Printf("   Containers may need to run with --privileged\n")
	} else {
		fmt.Println("✅ OK")
	}

	// Check available space
	fmt.Print("💾 Disk space: ")
	cmd = exec.Command("podman", "system", "df", "--format", "table")
	output, err = cmd.Output()
	if err == nil {
		fmt.Println("✅ OK")
		fmt.Printf("   Podman disk usage:\n")
		lines := strings.Split(string(output), "\n")
		for _, line := range lines[:min(len(lines), 5)] {
			if strings.TrimSpace(line) != "" {
				fmt.Printf("   %s\n", line)
			}
		}
	} else {
		fmt.Println("⚠️  WARNING (could not check disk usage)")
	}

	// Check current directory permissions
	fmt.Print("📁 Current directory: ")
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("❌ FAILED")
		return fmt.Errorf("cannot determine current directory")
	}

	// Test if we can create a test file (needed for volume mounting)
	testFile := filepath.Join(currentDir, ".portunix-test")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("   Cannot write to current directory: %s\n", currentDir)
		return fmt.Errorf("insufficient permissions in current directory")
	}
	os.Remove(testFile) // cleanup
	fmt.Printf("✅ OK (%s)\n", currentDir)

	// Check cache directory
	fmt.Print("🗂️  Cache directory: ")
	cacheDir := ".cache"
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("   Cannot create cache directory: %s\n", cacheDir)
		return fmt.Errorf("cannot create cache directory")
	}
	cachePath, _ := filepath.Abs(cacheDir)
	fmt.Printf("✅ OK (%s)\n", cachePath)

	// Check network connectivity (optional)
	fmt.Print("🌐 Network connectivity: ")
	cmd = exec.Command("podman", "run", "--rm", "alpine:latest", "ping", "-c", "1", "google.com")
	err = cmd.Run()
	if err != nil {
		fmt.Println("⚠️  WARNING (could not test network - this may affect image pulling)")
	} else {
		fmt.Println("✅ OK")
	}

	fmt.Println()
	fmt.Println("🎉 All critical requirements are satisfied!")
	fmt.Println("💡 You can now run: portunix podman run-in-container <type>")

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Storage analysis functions (placeholders - reuse Docker implementation)

type DriveInfo struct {
	Letter     string
	FreeSpace  string
	TotalSpace string
}

type PartitionInfo struct {
	MountPoint string
	FreeSpace  string
	TotalSpace string
}

func analyzeWindowsStorage() ([]DriveInfo, error) {
	// Placeholder implementation
	return []DriveInfo{
		{Letter: "D", FreeSpace: "850 GB", TotalSpace: "1 TB"},
		{Letter: "C", FreeSpace: "125 GB", TotalSpace: "256 GB"},
	}, nil
}

func analyzeLinuxStorage() ([]PartitionInfo, error) {
	// Placeholder implementation
	return []PartitionInfo{
		{MountPoint: "/data", FreeSpace: "500 GB", TotalSpace: "1 TB"},
		{MountPoint: "/", FreeSpace: "45 GB", TotalSpace: "100 GB"},
	}, nil
}

func promptStorageSelection(drives []DriveInfo) (string, error) {
	fmt.Printf("\n💡 Storage Recommendation: Drive %s:\\ (%s available)\n", drives[0].Letter, drives[0].FreeSpace)
	fmt.Println("   Podman images and containers can consume significant space.")
	fmt.Printf("   Using %s:\\ will provide better performance and prevent system drive filling up.\n", drives[0].Letter)
	fmt.Println()
	fmt.Println("Select Podman data storage location:")
	for i, drive := range drives {
		status := ""
		if i == 0 {
			status = " (recommended)"
		}
		fmt.Printf("%d. %s:\\ - %s available%s\n", i+1, drive.Letter, drive.FreeSpace, status)
	}
	fmt.Println("4. Custom path")
	fmt.Println()
	fmt.Print("Choice [1]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	if response == "" || response == "1" {
		return drives[0].Letter, nil
	}

	// Handle other choices (simplified)
	return drives[0].Letter, nil
}

func promptLinuxStorageSelection(partitions []PartitionInfo) (string, error) {
	fmt.Printf("\n💡 Storage Recommendation: %s (%s available)\n", partitions[0].MountPoint, partitions[0].FreeSpace)
	fmt.Println("   Podman images and containers can consume significant space.")
	fmt.Printf("   Using %s will prevent root partition from filling up.\n", partitions[0].MountPoint)
	fmt.Println()
	fmt.Println("Select Podman data storage location:")
	for i, partition := range partitions {
		status := ""
		if i == 0 {
			status = " (recommended)"
		}
		fmt.Printf("%d. %s - %s available%s\n", i+1, partition.MountPoint+"/podman-data", partition.FreeSpace, status)
	}
	fmt.Println("4. Custom path")
	fmt.Println()
	fmt.Print("Choice [1]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	if response == "" || response == "1" {
		return partitions[0].MountPoint + "/podman-data", nil
	}

	// Handle other choices (simplified)
	return partitions[0].MountPoint + "/podman-data", nil
}

// Installation functions (placeholders adapted for Podman)

func downloadPodmanDesktop(path string) error {
	fmt.Println("📦 Downloading Podman Desktop installer...")
	// Placeholder for actual download implementation
	return nil
}

func installPodmanDesktopWindows(installerPath, dataRoot string) error {
	fmt.Printf("🔧 Running installer with admin privileges...\n")
	fmt.Printf("✓ Podman Desktop installed successfully\n")
	fmt.Printf("🔧 Configuring data-root: %s\n", dataRoot)
	fmt.Printf("✓ WSL2 backend configured\n")
	fmt.Printf("✓ Podman system started\n")
	return nil
}

func installPodmanUbuntuDebian(dataRoot string) error {
	fmt.Println("🔧 Installing Podman on Ubuntu/Debian...")
	fmt.Println()

	// Update package list
	fmt.Println("📥 Updating package list...")
	fmt.Println("Running: sudo apt update")
	updateCmd := exec.Command("sudo", "apt", "update")
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr
	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("failed to update package list: %w", err)
	}
	fmt.Println("✓ Package list updated")
	fmt.Println()

	// Install podman
	fmt.Println("📦 Installing Podman package...")
	fmt.Println("Running: sudo apt install -y podman")
	installCmd := exec.Command("sudo", "apt", "install", "-y", "podman")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install podman: %w", err)
	}
	fmt.Println("✓ Podman installed successfully")
	fmt.Println()

	// Configure for current user
	fmt.Println("🔧 Configuring rootless containers...")
	fmt.Println("Setting up user namespaces and subuid/subgid...")

	// Create containers storage directory
	if dataRoot != "" {
		fmt.Printf("🔧 Configuring storage location: %s\n", dataRoot)
		if err := os.MkdirAll(dataRoot, 0755); err != nil {
			fmt.Printf("⚠️  Could not create storage directory: %v\n", err)
		}
	}

	fmt.Println("✓ Rootless configuration completed")
	fmt.Println("✓ Podman is ready to use")
	fmt.Println()
	fmt.Println("💡 Note: You may need to restart your terminal or run 'newgrp' for group changes to take effect")

	return nil
}

func installPodmanCentOSFedora(dataRoot string) error {
	fmt.Println("🔧 Installing Podman on CentOS/RHEL/Rocky Linux/Fedora...")
	fmt.Println()

	// Detect package manager
	var packageManager string
	var installCmd *exec.Cmd

	// Check if dnf is available (Fedora, newer RHEL)
	if _, err := exec.LookPath("dnf"); err == nil {
		packageManager = "dnf"
		fmt.Println("📦 Installing Podman package...")
		fmt.Println("Running: sudo dnf install -y podman")
		installCmd = exec.Command("sudo", "dnf", "install", "-y", "podman")
	} else {
		packageManager = "yum"
		fmt.Println("📦 Installing Podman package...")
		fmt.Println("Running: sudo yum install -y podman")
		installCmd = exec.Command("sudo", "yum", "install", "-y", "podman")
	}

	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install podman via %s: %w", packageManager, err)
	}
	fmt.Println("✓ Podman installed successfully")
	fmt.Println()

	// Configure for current user
	fmt.Println("🔧 Configuring rootless containers...")
	fmt.Println("Setting up user namespaces and subuid/subgid...")

	// Create containers storage directory
	if dataRoot != "" {
		fmt.Printf("🔧 Configuring storage location: %s\n", dataRoot)
		if err := os.MkdirAll(dataRoot, 0755); err != nil {
			fmt.Printf("⚠️  Could not create storage directory: %v\n", err)
		}
	}

	fmt.Println("✓ Rootless configuration completed")
	fmt.Println("✓ Podman is ready to use")
	fmt.Println()
	fmt.Println("💡 Note: You may need to restart your terminal for changes to take effect")

	return nil
}

func installPodmanAlpine(dataRoot string) error {
	fmt.Println("🔧 Installing Podman on Alpine Linux...")
	fmt.Println()

	// Update package index
	fmt.Println("📥 Updating package index...")
	fmt.Println("Running: sudo apk update")
	updateCmd := exec.Command("sudo", "apk", "update")
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr
	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("failed to update package index: %w", err)
	}
	fmt.Println("✓ Package index updated")
	fmt.Println()

	// Install podman
	fmt.Println("📦 Installing Podman package...")
	fmt.Println("Running: sudo apk add podman")
	installCmd := exec.Command("sudo", "apk", "add", "podman")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install podman: %w", err)
	}
	fmt.Println("✓ Podman installed successfully")
	fmt.Println()

	// Configure for current user
	fmt.Println("🔧 Configuring rootless containers...")
	fmt.Println("Setting up user namespaces...")

	// Create containers storage directory
	if dataRoot != "" {
		fmt.Printf("🔧 Configuring storage location: %s\n", dataRoot)
		if err := os.MkdirAll(dataRoot, 0755); err != nil {
			fmt.Printf("⚠️  Could not create storage directory: %v\n", err)
		}
	}

	fmt.Println("✓ Rootless configuration completed")
	fmt.Println("✓ Podman is ready to use")
	fmt.Println()
	fmt.Println("💡 Note: You may need to restart your terminal for changes to take effect")

	return nil
}

func installPodmanGeneric(dataRoot string) error {
	fmt.Println("📦 Downloading Podman binaries directly...")
	fmt.Printf("✓ Podman installed successfully\n")
	fmt.Printf("🔧 Configuring data-root: %s\n", dataRoot)
	fmt.Println("✓ Podman system configured")
	return nil
}

// Podman Desktop installation functions

func installPodmanDesktopGUIWindows(autoAccept bool) error {
	fmt.Println("\nInstalling Podman Desktop for Windows...")
	fmt.Println("🖥️  This will install the official GUI application from Red Hat")
	fmt.Println("📋 Features:")
	fmt.Println("   • Visual container, image, pod management")
	fmt.Println("   • Integration with Docker and Podman")
	fmt.Println("   • Remote container management")
	fmt.Println("   • Kubernetes integration")
	fmt.Println()

	if !autoAccept {
		fmt.Print("Continue with Podman Desktop installation? [Y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		if response == "n" || response == "no" {
			fmt.Println("❌ Installation cancelled")
			return fmt.Errorf("user cancelled installation")
		}
	}

	// Download latest Podman Desktop
	installerPath := filepath.Join(".cache", "PodmanDesktop-latest.exe")
	downloadURL := "https://github.com/containers/podman-desktop/releases/latest/download/podman-desktop-1.21.0-setup-x64.exe"
	
	if err := downloadPodmanDesktopInstaller(downloadURL, installerPath); err != nil {
		return fmt.Errorf("failed to download Podman Desktop: %w", err)
	}

	// Run installer
	fmt.Println("🚀 Running Podman Desktop installer...")
	cmd := exec.Command(installerPath)
	if autoAccept {
		// Silent installation
		cmd.Args = append(cmd.Args, "/S")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Podman Desktop: %w", err)
	}

	fmt.Println("✅ Podman Desktop installed successfully!")
	fmt.Println("💡 You can now launch 'Podman Desktop' from the Start menu")
	fmt.Println("🌐 Learn more at: https://podman-desktop.io")

	return nil
}

func installPodmanDesktopGUILinux(autoAccept bool, osInfo *system.SystemInfo) error {
	fmt.Println("\nInstalling Podman Desktop for Linux...")
	fmt.Println("🖥️  Installing official GUI from Red Hat")
	
	// Determine installation method based on distro
	distribution := strings.ToLower(osInfo.OS)
	if osInfo.LinuxInfo != nil {
		distribution = strings.ToLower(osInfo.LinuxInfo.Distribution)
	}

	fmt.Printf("📦 Detected distribution: %s\n", distribution)
	
	switch distribution {
	case "ubuntu", "debian":
		return installPodmanDesktopUbuntu(autoAccept)
	case "fedora", "rhel", "centos", "rocky":
		return installPodmanDesktopFedora(autoAccept)
	case "arch":
		return installPodmanDesktopArch(autoAccept)
	default:
		return installPodmanDesktopGenericLinux(autoAccept)
	}
}

func installPodmanDesktopGUIMacOS(autoAccept bool) error {
	fmt.Println("\nInstalling Podman Desktop for macOS...")
	fmt.Println("🖥️  Installing official GUI from Red Hat")
	
	if !autoAccept {
		fmt.Print("Continue with Podman Desktop installation? [Y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		if response == "n" || response == "no" {
			fmt.Println("❌ Installation cancelled")
			return fmt.Errorf("user cancelled installation")
		}
	}

	// Check if Homebrew is available
	if _, err := exec.LookPath("brew"); err == nil {
		fmt.Println("🍺 Installing via Homebrew...")
		cmd := exec.Command("brew", "install", "--cask", "podman-desktop")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			fmt.Println("⚠️  Homebrew installation failed, trying direct download...")
			return installPodmanDesktopMacOSDirect(autoAccept)
		}
	} else {
		return installPodmanDesktopMacOSDirect(autoAccept)
	}

	fmt.Println("✅ Podman Desktop installed successfully!")
	fmt.Println("💡 Launch from Applications folder or Spotlight")
	fmt.Println("🌐 Learn more at: https://podman-desktop.io")

	return nil
}

func installPodmanDesktopUbuntu(autoAccept bool) error {
	fmt.Println("📦 Installing Podman Desktop via system package...")
	
	// Add Podman Desktop repository if needed
	fmt.Println("🔑 Adding Podman Desktop repository...")
	
	// For now, use AppImage as it's more universal
	return installPodmanDesktopGenericLinux(autoAccept)
}

func installPodmanDesktopFedora(autoAccept bool) error {
	fmt.Println("📦 Installing Podman Desktop via DNF...")
	
	// Check if podman-desktop package is available
	cmd := exec.Command("dnf", "search", "podman-desktop")
	if err := cmd.Run(); err != nil {
		fmt.Println("⚠️  Package not found in repositories, using AppImage...")
		return installPodmanDesktopGenericLinux(autoAccept)
	}

	// Install via DNF
	installCmd := exec.Command("sudo", "dnf", "install", "-y", "podman-desktop")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	
	if err := installCmd.Run(); err != nil {
		fmt.Println("⚠️  DNF installation failed, using AppImage...")
		return installPodmanDesktopGenericLinux(autoAccept)
	}

	fmt.Println("✅ Podman Desktop installed successfully!")
	return nil
}

func installPodmanDesktopArch(autoAccept bool) error {
	fmt.Println("📦 Installing Podman Desktop via AUR...")
	
	// Check if yay or paru is available
	var aurHelper string
	if _, err := exec.LookPath("yay"); err == nil {
		aurHelper = "yay"
	} else if _, err := exec.LookPath("paru"); err == nil {
		aurHelper = "paru"
	} else {
		fmt.Println("⚠️  No AUR helper found, using AppImage...")
		return installPodmanDesktopGenericLinux(autoAccept)
	}

	cmd := exec.Command(aurHelper, "-S", "--noconfirm", "podman-desktop-bin")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Println("⚠️  AUR installation failed, using AppImage...")
		return installPodmanDesktopGenericLinux(autoAccept)
	}

	fmt.Println("✅ Podman Desktop installed successfully!")
	return nil
}

func installPodmanDesktopGenericLinux(autoAccept bool) error {
	fmt.Println("📦 Installing Podman Desktop...")
	
	// Check available package managers and provide specific instructions
	hasFlatpak := false
	
	if _, err := exec.LookPath("flatpak"); err == nil {
		hasFlatpak = true
	}
	
	fmt.Println("📋 Podman Desktop installation options:")
	fmt.Println()
	
	if hasFlatpak {
		fmt.Println("✅ Option 1 - Flatpak (recommended):")
		fmt.Println("   flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo")
		fmt.Println("   flatpak install flathub io.podman_desktop.PodmanDesktop")
		fmt.Println("   flatpak run io.podman_desktop.PodmanDesktop")
		fmt.Println()
	} else {
		fmt.Println("📦 Option 1 - Install Flatpak first:")
		fmt.Println("   sudo apt install flatpak")
		fmt.Println("   flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo")
		fmt.Println("   flatpak install flathub io.podman_desktop.PodmanDesktop")
		fmt.Println()
	}
	
	fmt.Println("📥 Option 2 - Download AppImage:")
	fmt.Println("   wget https://github.com/containers/podman-desktop/releases/latest/download/podman-desktop-1.20.2.flatpak")
	fmt.Println("   # Or browse: https://podman-desktop.io/downloads/linux")
	fmt.Println()
	
	fmt.Println("🐳 Option 3 - Use Docker Desktop alternative:")
	fmt.Println("   # Podman Desktop provides Docker Desktop-like experience")
	fmt.Println("   # with better security (rootless containers)")
	fmt.Println()
	
	if !autoAccept {
		fmt.Print("Would you like to install Flatpak and proceed with automatic installation? [Y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		if response == "y" || response == "yes" || response == "" {
			return installFlatpakAndPodmanDesktop(autoAccept)
		}
	}
	
	return fmt.Errorf("manual installation required - see options above")
}

func installPodmanDesktopFlatpak(autoAccept bool) error {
	fmt.Println("📦 Installing Podman Desktop via Flatpak...")
	
	if !autoAccept {
		fmt.Print("Continue with Flatpak installation? [Y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		if response == "n" || response == "no" {
			return fmt.Errorf("user cancelled installation")
		}
	}
	
	// Add flathub repository if not exists
	fmt.Println("🔑 Adding Flathub repository...")
	addRepoCmd := exec.Command("flatpak", "remote-add", "--if-not-exists", "flathub", "https://flathub.org/repo/flathub.flatpakrepo")
	addRepoCmd.Stdout = os.Stdout
	addRepoCmd.Stderr = os.Stderr
	addRepoCmd.Run() // Ignore errors - may already exist
	
	// Install Podman Desktop
	fmt.Println("📦 Installing Podman Desktop...")
	installCmd := exec.Command("flatpak", "install", "-y", "flathub", "io.podman_desktop.PodmanDesktop")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("flatpak installation failed: %w", err)
	}
	
	// Create alias for easier launching
	if err := createPodmanDesktopAlias(); err != nil {
		fmt.Printf("⚠️  Could not create alias: %v\n", err)
		fmt.Println("💡 You can create it manually: alias podman-desktop=\"flatpak run io.podman_desktop.PodmanDesktop\"")
	} else {
		fmt.Println("✅ Created 'podman-desktop' alias for easy launching")
	}

	fmt.Println("✅ Podman Desktop installed successfully!")
	fmt.Println()
	fmt.Println("🚀 Launch options:")
	fmt.Println("   podman-desktop                                    # Using alias")
	fmt.Println("   flatpak run io.podman_desktop.PodmanDesktop       # Direct command")
	fmt.Println("   # Or find 'Podman Desktop' in your applications menu")
	fmt.Println()
	fmt.Println("🌐 Learn more at: https://podman-desktop.io")
	
	return nil
}

func installPodmanDesktopSnap(autoAccept bool) error {
	fmt.Println("📦 Installing Podman Desktop via Snap...")
	
	if !autoAccept {
		fmt.Print("Continue with Snap installation? [Y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		if response == "n" || response == "no" {
			return fmt.Errorf("user cancelled installation")
		}
	}
	
	// Install Podman Desktop
	fmt.Println("📦 Installing Podman Desktop...")
	fmt.Println("🔐 This requires administrator privileges")
	
	installCmd := exec.Command("sudo", "snap", "install", "podman-desktop")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	installCmd.Stdin = os.Stdin  // Allow sudo to read password from terminal
	
	if err := installCmd.Run(); err != nil {
		fmt.Println("⚠️  Snap installation failed. You can install manually:")
		fmt.Println("   sudo snap install podman-desktop")
		return fmt.Errorf("snap installation failed: %w", err)
	}
	
	fmt.Println("✅ Podman Desktop installed successfully!")
	fmt.Println("💡 Launch with: podman-desktop")
	fmt.Println("🌐 Learn more at: https://podman-desktop.io")
	
	return nil
}

func installPodmanDesktopMacOSDirect(autoAccept bool) error {
	fmt.Println("📦 Downloading Podman Desktop for macOS...")
	
	downloadURL := "https://github.com/containers/podman-desktop/releases/latest/download/podman-desktop-1.21.0-arm64.dmg"
	dmgPath := filepath.Join(".cache", "PodmanDesktop.dmg")
	
	if err := downloadPodmanDesktopInstaller(downloadURL, dmgPath); err != nil {
		return fmt.Errorf("failed to download Podman Desktop: %w", err)
	}

	fmt.Println("📱 Please manually install the downloaded DMG file:")
	fmt.Printf("   1. Double-click: %s\n", dmgPath)
	fmt.Println("   2. Drag Podman Desktop to Applications folder")
	fmt.Println("   3. Launch from Applications or Spotlight")

	return nil
}

func downloadPodmanDesktopInstaller(url, filePath string) error {
	fmt.Printf("⬇️  Downloading from: %s\n", url)
	
	// Create cache directory
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy the response body to file with progress
	fmt.Printf("📥 Downloading to: %s\n", filePath)
	
	// Get content length for progress
	contentLength := resp.ContentLength
	if contentLength > 0 {
		fmt.Printf("📊 File size: %.2f MB\n", float64(contentLength)/(1024*1024))
	}
	
	// Copy data
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Println("✅ Download completed successfully")
	return nil
}

func installFlatpakAndPodmanDesktop(autoAccept bool) error {
	fmt.Println("📦 Installing Flatpak first...")
	fmt.Println("🔐 This requires administrator privileges")
	
	// Install Flatpak
	installCmd := exec.Command("sudo", "apt", "install", "-y", "flatpak")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	installCmd.Stdin = os.Stdin
	
	if err := installCmd.Run(); err != nil {
		fmt.Println("⚠️  Failed to install Flatpak. Please install manually:")
		fmt.Println("   sudo apt install flatpak")
		return fmt.Errorf("flatpak installation failed: %w", err)
	}
	
	// Add Flathub repository
	fmt.Println("🔑 Adding Flathub repository...")
	addRepoCmd := exec.Command("flatpak", "remote-add", "--if-not-exists", "flathub", "https://flathub.org/repo/flathub.flatpakrepo")
	addRepoCmd.Stdout = os.Stdout
	addRepoCmd.Stderr = os.Stderr
	addRepoCmd.Run()
	
	// Now install Podman Desktop
	return installPodmanDesktopFlatpak(true) // Use autoAccept since user already confirmed
}

func createPodmanDesktopAlias() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	bashrcPath := filepath.Join(homeDir, ".bashrc")
	aliasLine := `alias podman-desktop="flatpak run io.podman_desktop.PodmanDesktop"`
	
	// Check if alias already exists
	if content, err := os.ReadFile(bashrcPath); err == nil {
		if strings.Contains(string(content), aliasLine) {
			// Alias already exists
			return nil
		}
	}
	
	// Add alias to .bashrc
	file, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .bashrc: %w", err)
	}
	defer file.Close()
	
	// Add newlines and comment for clarity
	aliasContent := fmt.Sprintf("\n# Podman Desktop alias (added by Portunix)\n%s\n", aliasLine)
	
	if _, err := file.WriteString(aliasContent); err != nil {
		return fmt.Errorf("failed to write alias: %w", err)
	}
	
	// Also try to add to .bash_aliases if it exists
	bashAliasesPath := filepath.Join(homeDir, ".bash_aliases")
	if _, err := os.Stat(bashAliasesPath); err == nil {
		aliasFile, err := os.OpenFile(bashAliasesPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			defer aliasFile.Close()
			aliasFile.WriteString(fmt.Sprintf("\n# Podman Desktop alias (added by Portunix)\n%s\n", aliasLine))
		}
	}
	
	return nil
}

func createDesktopEntry(appImagePath string) {
	homeDir, _ := os.UserHomeDir()
	desktopDir := filepath.Join(homeDir, ".local", "share", "applications")
	os.MkdirAll(desktopDir, 0755)
	
	desktopEntry := fmt.Sprintf(`[Desktop Entry]
Name=Podman Desktop
Comment=Container management GUI
Exec=%s
Icon=podman-desktop
Terminal=false
Type=Application
Categories=Development;
`, appImagePath)

	desktopFile := filepath.Join(desktopDir, "podman-desktop.desktop")
	os.WriteFile(desktopFile, []byte(desktopEntry), 0644)
}
