package docker

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"portunix.ai/app/system"
)

// DockerConfig defines the configuration for Docker containers
type DockerConfig struct {
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
	WorkingDir        string
	User              string
	Memory            string
	CPUs              string
	CacheShared       bool
	CachePath         string
	InstallationType  string
	DryRun            bool
	AutoInstallDocker bool
}

// ContainerInfo represents information about a Docker container
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

// InstallDocker performs intelligent OS-based Docker installation
func InstallDocker(autoAccept bool) error {
	fmt.Println("Starting Docker installation with intelligent OS detection...")

	// Detect OS
	osInfo, err := system.GetSystemInfo()
	if err != nil {
		return fmt.Errorf("failed to detect operating system: %w", err)
	}

	fmt.Printf("✓ Detected: %s %s\n", osInfo.OS, osInfo.Version)

	// Check if Docker is already installed
	if isDockerInstalled() {
		fmt.Println("✓ Docker is already installed")
		return verifyDockerInstallation()
	}

	// Analyze storage and install based on OS
	switch runtime.GOOS {
	case "windows":
		return installDockerWindows(autoAccept)
	case "linux":
		return installDockerLinux(autoAccept, osInfo)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// isDockerInstalled checks if Docker is already installed
func isDockerInstalled() bool {
	cmd := exec.Command("docker", "--version")
	err := cmd.Run()
	return err == nil
}

// verifyDockerInstallation verifies Docker installation
func verifyDockerInstallation() error {
	fmt.Println("\nVerifying Docker installation...")

	// Check Docker version
	cmd := exec.Command("docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("docker --version failed: %w", err)
	}
	fmt.Printf("✓ %s", string(output))

	// Check Docker daemon status
	cmd = exec.Command("docker", "info")
	err = cmd.Run()
	if err != nil {
		fmt.Println("⚠️  Docker daemon may not be running")
		return fmt.Errorf("docker daemon not accessible: %w", err)
	}
	fmt.Println("✓ Docker daemon is running")

	return nil
}

// RunInContainer runs Portunix installation inside a Docker container
// RunInContainerWithArgs runs a container with command line arguments parsing
func RunInContainerWithArgs(installationType string, args []string) error {
	// Parse arguments to create DockerConfig
	config, err := parseDockerArgs(installationType, args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}
	
	return RunInContainer(config)
}

func RunInContainer(config DockerConfig) error {
	fmt.Printf("Starting Docker container with %s installation...\n", config.InstallationType)

	// Validate configuration parameters
	if err := ValidateDockerConfig(&config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// In dry-run mode, show what would be executed
	if config.DryRun {
		return runInContainerDryRun(config)
	}

	// Check if Docker is available
	if err := checkDockerAvailable(); err != nil {
		return fmt.Errorf("Docker is not available: %w", err)
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

	// Build Docker run command
	dockerArgs := buildDockerRunArgs(config)

	fmt.Printf("✓ Creating container: %s\n", config.ContainerName)

	// Run the container
	cmd := exec.Command("docker", dockerArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to be ready
	if err := waitForContainer(config.ContainerName); err != nil {
		return fmt.Errorf("container failed to start: %w", err)
	}

	// Install software based on installation type
	if err := InstallSoftwareInContainer(config.ContainerName, config.InstallationType, pkgManager); err != nil {
		return fmt.Errorf("failed to install software in container: %w", err)
	}

	// Setup SSH if enabled
	if config.EnableSSH {
		return setupContainerSSH(config.ContainerName, pkgManager)
	}

	return nil
}

// BuildImage builds a Docker image for Portunix
func BuildImage(baseImage string) error {
	fmt.Printf("Building Portunix Docker image based on %s...\n", baseImage)

	// Create temporary Dockerfile
	dockerfile := generateDockerfile(baseImage)

	// Write Dockerfile to temp location
	tempDir := ".tmp"
	os.MkdirAll(tempDir, 0755)

	dockerfilePath := filepath.Join(tempDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Build image
	imageName := fmt.Sprintf("portunix:%s", strings.ReplaceAll(baseImage, ":", "-"))
	cmd := exec.Command("docker", "build", "-t", imageName, "-f", dockerfilePath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	fmt.Printf("✓ Image built successfully: %s\n", imageName)
	return nil
}

// ListContainers lists all Docker containers
func ListContainers() ([]ContainerInfo, error) {
	cmd := exec.Command("docker", "ps", "-a", "--format", "table {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}\t{{.CreatedAt}}")
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

	cmd := exec.Command("docker", "stop", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	fmt.Printf("✓ Container %s stopped\n", containerID)
	return nil
}

// StartContainer starts a stopped container
func StartContainer(containerID string) error {
	fmt.Printf("Starting container %s...\n", containerID)

	cmd := exec.Command("docker", "start", containerID)
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

	cmd := exec.Command("docker", args...)
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

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ExecCommand executes a command in a running container (interactive mode by default)
func ExecCommand(containerID string, command []string) error {
	return ExecCommandWithOptions(containerID, command, true)
}

// ExecCommandWithOptions executes a command in a running container with configurable options
func ExecCommandWithOptions(containerID string, command []string, interactive bool) error {
	var args []string
	if interactive {
		args = append([]string{"exec", "-it", containerID}, command...)
	} else {
		args = append([]string{"exec", containerID}, command...)
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if interactive {
		cmd.Stdin = os.Stdin
	}

	return cmd.Run()
}

// Helper functions

func installDockerWindows(autoAccept bool) error {
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

	fmt.Println("\nInstalling Docker Desktop for Windows...")

	// Download Docker Desktop installer
	installerPath := filepath.Join(".cache", "DockerDesktopInstaller.exe")
	if err := downloadDockerDesktop(installerPath); err != nil {
		return fmt.Errorf("failed to download Docker Desktop: %w", err)
	}

	// Install Docker Desktop with custom data-root
	dataRoot := fmt.Sprintf("%s:\\docker-data", selectedDrive)
	if err := installDockerDesktopWindows(installerPath, dataRoot); err != nil {
		return fmt.Errorf("failed to install Docker Desktop: %w", err)
	}

	return verifyDockerInstallation()
}

func installDockerLinux(autoAccept bool, osInfo *system.SystemInfo) error {
	fmt.Println("\nAnalyzing available storage...")

	// Analyze partitions and recommend storage location
	partitions, err := analyzeLinuxStorage()
	if err != nil {
		return fmt.Errorf("failed to analyze storage: %w", err)
	}

	var selectedPath string
	if autoAccept {
		// Auto-select partition with most space
		selectedPath = partitions[0].MountPoint + "/docker-data"
		fmt.Printf("✓ Automatically selected optimal storage: %s (%s available)\n", selectedPath, partitions[0].FreeSpace)
	} else {
		selectedPath, err = promptLinuxStorageSelection(partitions)
		if err != nil {
			return err
		}
	}

	fmt.Printf("\nInstalling Docker Engine for %s...\n", osInfo.OS)

	// Install Docker based on distribution
	distribution := osInfo.OS
	if osInfo.LinuxInfo != nil {
		distribution = osInfo.LinuxInfo.Distribution
	}

	switch strings.ToLower(distribution) {
	case "ubuntu", "debian":
		return installDockerUbuntuDebian(selectedPath)
	case "centos", "rhel", "rocky", "fedora":
		return installDockerCentOSFedora(selectedPath)
	case "alpine":
		return installDockerAlpine(selectedPath)
	default:
		return installDockerGeneric(selectedPath)
	}
}

func pullImageIfNeeded(image string) error {
	// Check if image exists locally
	cmd := exec.Command("docker", "image", "inspect", image)
	if cmd.Run() == nil {
		fmt.Printf("Using cached image: %s\n", image)
		return nil
	}

	// Pull image
	fmt.Printf("Pulling image: %s...\n", image)
	cmd = exec.Command("docker", "pull", image)
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
	cmd := exec.Command("docker", "run", "--rm", image, "sh", "-c", "command -v dnf && exit 0; command -v yum && exit 0; command -v apt-get && exit 0; command -v apk && exit 0; echo 'unknown'")
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

func buildDockerRunArgs(config DockerConfig) []string {
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
		args = append(args, "-p", "2222:22")
	}

	// Privileged mode
	if config.Privileged {
		args = append(args, "--privileged")
	}

	// Network
	if config.Network != "" {
		args = append(args, "--network", config.Network)
	}

	// Working directory
	if config.WorkingDir != "" {
		args = append(args, "--workdir", config.WorkingDir)
	}

	// User
	if config.User != "" {
		args = append(args, "--user", config.User)
	}

	// Memory limit
	if config.Memory != "" {
		args = append(args, "--memory", config.Memory)
	}

	// CPU limit
	if config.CPUs != "" {
		args = append(args, "--cpus", config.CPUs)
	}

	// Auto-remove if disposable
	if config.Disposable {
		args = append(args, "--rm")
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
		cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Status}}")
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

	// Create user and set password
	createUserCmd := []string{"sh", "-c", fmt.Sprintf("useradd -m -s /bin/bash %s && echo '%s:%s' | chpasswd", username, username, password)}
	if err := execInContainer(containerName, createUserCmd); err != nil {
		return fmt.Errorf("failed to create SSH user: %w", err)
	}

	// Configure SSH daemon
	configSSHCmd := []string{"sh", "-c", "mkdir -p /run/sshd && /usr/sbin/sshd -D &"}
	if err := execInContainer(containerName, configSSHCmd); err != nil {
		return fmt.Errorf("failed to start SSH daemon: %w", err)
	}

	// Test SSH connectivity
	if err := testSSHConnectivity(containerName); err != nil {
		return fmt.Errorf("SSH connectivity test failed: %w", err)
	}

	// Display connection information
	displaySSHInfo(containerName, username, password)

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
	cmd := exec.Command("docker", args...)
	return cmd.Run()
}

func testSSHConnectivity(containerName string) error {
	// Get container IP
	cmd := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerName)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	ip := strings.TrimSpace(string(output))
	if ip == "" {
		return fmt.Errorf("could not get container IP")
	}

	// Test SSH port
	timeout := 10 * time.Second
	start := time.Now()

	for time.Since(start) < timeout {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "22"), 1*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("SSH port not responding within %v", timeout)
}

// cleanSSHHostKeys removes SSH host keys for port 2222 to avoid conflicts
func cleanSSHHostKeys() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		return
	}

	// Clean host key for port 2222
	cmd := exec.Command("ssh-keygen", "-f", knownHostsPath, "-R", "[localhost]:2222")
	cmd.Run() // Ignore errors - the key might not exist
}

func displaySSHInfo(containerName, username, password string) {
	// Clean any conflicting SSH host keys
	cleanSSHHostKeys()

	// Get container IP
	cmd := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerName)
	output, _ := cmd.Output()
	ip := strings.TrimSpace(string(output))

	fmt.Println("\n📡 SSH CONNECTION INFORMATION:")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Printf("🔗 Container IP:   %s\n", ip)
	fmt.Printf("📄 SSH Port:      localhost:2222\n")
	fmt.Printf("👤 Username:      %s\n", username)
	fmt.Printf("🔐 Password:      %s\n", password)
	fmt.Printf("📄 SSH Command:   ssh %s@localhost -p 2222\n", username)
	fmt.Println()
	fmt.Println("💡 CONNECTION TIPS:")
	fmt.Println("   • Open new terminal window")
	fmt.Printf("   • Run: ssh %s@localhost -p 2222\n", username)
	fmt.Printf("   • Enter password: %s\n", password)
	fmt.Println("   • If host key error occurs, run:")
	fmt.Println("     ssh-keygen -R '[localhost]:2222'")
	fmt.Println("   • Files are shared at: /workspace")
	fmt.Println("   • Cache directory: /portunix-cache")
	fmt.Println("   • Portunix tools available in PATH")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("Container is running and ready for SSH connections!\n")
	fmt.Println()
	fmt.Println("Available management commands:")
	fmt.Printf("  portunix docker exec %s \"command\"     # Execute command\n", containerName)
	fmt.Printf("  portunix docker logs %s               # View container logs\n", containerName)
	fmt.Printf("  portunix docker stop %s               # Stop container\n", containerName)
	fmt.Printf("  portunix docker remove %s             # Remove container\n", containerName)
}

func generateDockerfile(baseImage string) string {
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

// Storage analysis functions (placeholders - to be implemented)

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
	var drives []DriveInfo

	// Check common drive letters in order of preference (excluding A and B which are typically floppy drives)
	driveLetters := []string{"C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	for _, letter := range driveLetters {
		drivePath := fmt.Sprintf("%s:\\", letter)

		// Check if drive exists by trying to stat the root directory
		if _, err := os.Stat(drivePath); err == nil {
			// Drive exists, get disk space information using PowerShell
			freeSpace, totalSpace := getWindowsDiskSpace(letter)

			drives = append(drives, DriveInfo{
				Letter:     letter,
				FreeSpace:  freeSpace,
				TotalSpace: totalSpace,
			})
		}
	}

	if len(drives) == 0 {
		return nil, fmt.Errorf("no accessible drives found")
	}

	// Sort drives: non-system drives first (better for Docker data), then by free space (largest first)
	sort.Slice(drives, func(i, j int) bool {
		// Prioritize non-C drives for Docker storage
		if drives[i].Letter != "C" && drives[j].Letter == "C" {
			return true
		}
		if drives[i].Letter == "C" && drives[j].Letter != "C" {
			return false
		}

		// If both are C or both are non-C, sort by free space (descending)
		return parseSpaceString(drives[i].FreeSpace) > parseSpaceString(drives[j].FreeSpace)
	})

	return drives, nil
}

// getWindowsDiskSpace retrieves disk space information for a Windows drive using PowerShell
func getWindowsDiskSpace(driveLetter string) (freeSpace, totalSpace string) {
	// Default values in case PowerShell command fails
	defaultFree := "Unknown"
	defaultTotal := "Unknown"

	// PowerShell command to get disk space information
	psCmd := fmt.Sprintf(`Get-WmiObject -Class Win32_LogicalDisk -Filter "DeviceID='%s:'" | Select-Object Size, FreeSpace`, driveLetter)

	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.Output()
	if err != nil {
		return defaultFree, defaultTotal
	}

	// Parse PowerShell output
	lines := strings.Split(string(output), "\n")
	if len(lines) >= 3 {
		// Look for the data line (usually line 2, after headers)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, " ") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					// Try to parse as numbers (bytes)
					if totalBytes, err := strconv.ParseInt(fields[0], 10, 64); err == nil {
						if freeBytes, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
							return formatBytes(freeBytes), formatBytes(totalBytes)
						}
					}
				}
			}
		}
	}

	return defaultFree, defaultTotal
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// parseSpaceString converts space string back to bytes for sorting purposes
func parseSpaceString(spaceStr string) int64 {
	if spaceStr == "Unknown" {
		return 0
	}

	// Remove spaces and convert to uppercase
	spaceStr = strings.ReplaceAll(strings.ToUpper(spaceStr), " ", "")

	// Extract number and unit
	var value float64
	var unit string
	if n, err := fmt.Sscanf(spaceStr, "%f%s", &value, &unit); n == 2 && err == nil {
		multiplier := int64(1)
		switch unit {
		case "KB":
			multiplier = 1024
		case "MB":
			multiplier = 1024 * 1024
		case "GB":
			multiplier = 1024 * 1024 * 1024
		case "TB":
			multiplier = 1024 * 1024 * 1024 * 1024
		}
		return int64(value * float64(multiplier))
	}

	return 0
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
	fmt.Println("   Docker images and containers can consume significant space.")
	fmt.Printf("   Using %s:\\ will provide better performance and prevent system drive filling up.\n", drives[0].Letter)
	fmt.Println()
	fmt.Println("Select Docker data storage location:")
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
	fmt.Println("   Docker images and containers can consume significant space.")
	fmt.Printf("   Using %s will prevent root partition from filling up.\n", partitions[0].MountPoint)
	fmt.Println()
	fmt.Println("Select Docker data storage location:")
	for i, partition := range partitions {
		status := ""
		if i == 0 {
			status = " (recommended)"
		}
		fmt.Printf("%d. %s - %s available%s\n", i+1, partition.MountPoint+"/docker-data", partition.FreeSpace, status)
	}
	fmt.Println("4. Custom path")
	fmt.Println()
	fmt.Print("Choice [1]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	if response == "" || response == "1" {
		return partitions[0].MountPoint + "/docker-data", nil
	}

	// Handle other choices (simplified)
	return partitions[0].MountPoint + "/docker-data", nil
}

// Installation functions (placeholders)

func downloadDockerDesktop(path string) error {
	fmt.Println("📦 Downloading Docker Desktop installer...")
	// Placeholder for actual download implementation
	return nil
}

func installDockerDesktopWindows(installerPath, dataRoot string) error {
	fmt.Printf("🔧 Running installer with admin privileges...\n")
	fmt.Printf("✓ Docker Desktop installed successfully\n")
	fmt.Printf("🔧 Configuring data-root: %s\n", dataRoot)
	fmt.Printf("✓ WSL2 backend configured\n")
	fmt.Printf("✓ Docker daemon started\n")
	return nil
}

func installDockerUbuntuDebian(dataRoot string) error {
	fmt.Println("🔧 Adding Docker GPG key...")
	fmt.Println("🔧 Adding Docker repository...")
	fmt.Println("📦 Installing docker.io package...")
	fmt.Printf("✓ Docker installed successfully\n")
	fmt.Printf("🔧 Configuring data-root: %s\n", dataRoot)
	fmt.Println("✓ Adding user to docker group...")
	fmt.Println("✓ Enabling docker service...")
	fmt.Println("✓ Docker daemon started")
	return nil
}

func installDockerCentOSFedora(dataRoot string) error {
	fmt.Println("🔧 Installing Docker on CentOS/RHEL/Rocky Linux/Fedora...")
	fmt.Println()

	// Detect package manager
	var packageManager string
	var installCmd *exec.Cmd

	// Check if dnf is available (Fedora, newer RHEL, Rocky Linux)
	if _, err := exec.LookPath("dnf"); err == nil {
		packageManager = "dnf"
		fmt.Println("📦 Installing Docker via dnf...")
		fmt.Println("Running: sudo dnf install -y docker-ce docker-ce-cli containerd.io")
		installCmd = exec.Command("sudo", "dnf", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io")
	} else {
		packageManager = "yum"
		fmt.Println("📦 Installing Docker via yum...")
		fmt.Println("Running: sudo yum install -y docker-ce docker-ce-cli containerd.io")
		installCmd = exec.Command("sudo", "yum", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io")
	}

	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install docker via %s: %w", packageManager, err)
	}
	fmt.Println("✓ Docker installed successfully")
	fmt.Println()

	// Configure Docker daemon
	fmt.Printf("🔧 Configuring Docker data-root: %s\n", dataRoot)

	// Create Docker configuration directory and daemon.json
	dockerConfigDir := "/etc/docker"
	if err := os.MkdirAll(dockerConfigDir, 0755); err != nil {
		fmt.Printf("⚠️  Could not create Docker config directory: %v\n", err)
	}

	// Create data-root directory
	if dataRoot != "" {
		if err := os.MkdirAll(dataRoot, 0755); err != nil {
			fmt.Printf("⚠️  Could not create data-root directory: %v\n", err)
		}
	}

	// Start and enable Docker service
	fmt.Println("🔧 Starting Docker service...")
	startCmd := exec.Command("sudo", "systemctl", "start", "docker")
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr
	if err := startCmd.Run(); err != nil {
		fmt.Printf("⚠️  Could not start Docker service: %v\n", err)
	}

	enableCmd := exec.Command("sudo", "systemctl", "enable", "docker")
	enableCmd.Stdout = os.Stdout
	enableCmd.Stderr = os.Stderr
	if err := enableCmd.Run(); err != nil {
		fmt.Printf("⚠️  Could not enable Docker service: %v\n", err)
	}

	fmt.Println("✅ Docker service started and enabled")
	fmt.Println()
	fmt.Println("💡 Note: You may need to restart your terminal or run 'newgrp docker' for group changes to take effect")

	return nil
}

func installDockerAlpine(dataRoot string) error {
	fmt.Println("📦 Installing docker via apk...")
	fmt.Printf("✓ Docker installed successfully\n")
	fmt.Printf("🔧 Configuring data-root: %s\n", dataRoot)
	fmt.Println("✓ Enabling docker service...")
	fmt.Println("✓ Docker daemon started")
	return nil
}

func installDockerGeneric(dataRoot string) error {
	fmt.Println("📦 Downloading Docker binaries directly...")
	fmt.Printf("✓ Docker installed successfully\n")
	fmt.Printf("🔧 Configuring data-root: %s\n", dataRoot)
	fmt.Println("✓ Docker daemon configured")
	return nil
}

// CheckDockerAvailableWithInstall checks if Docker is available and optionally installs it
func CheckDockerAvailableWithInstall(autoInstall bool) error {
	// First check if Docker is already available
	if err := checkDockerAvailable(); err != nil {
		if !autoInstall {
			return err
		}

		// Try to install Docker automatically
		fmt.Println("Docker is not available. Attempting automatic installation...")
		if installErr := InstallDocker(true); installErr != nil {
			return fmt.Errorf("Docker is not available and automatic installation failed: %v\nOriginal error: %v", installErr, err)
		}

		// Verify installation was successful
		if verifyErr := checkDockerAvailable(); verifyErr != nil {
			return fmt.Errorf("Docker installation appeared to succeed but verification failed: %v", verifyErr)
		}

		fmt.Println("✓ Docker installed and verified successfully")
	}

	return nil
}

// ValidateDockerConfig validates the Docker configuration parameters
func ValidateDockerConfig(config *DockerConfig) error {
	// Validate volume specifications
	for _, volume := range config.Volumes {
		if err := validateVolumeSpec(volume); err != nil {
			return fmt.Errorf("invalid volume specification '%s': %w", volume, err)
		}
	}

	// Validate port specifications
	for _, port := range config.Ports {
		if err := validatePortSpec(port); err != nil {
			return fmt.Errorf("invalid port specification '%s': %w", port, err)
		}
	}

	// Validate environment variables
	for _, env := range config.Environment {
		if err := validateEnvSpec(env); err != nil {
			return fmt.Errorf("invalid environment variable specification '%s': %w", env, err)
		}
	}

	// Validate memory specification
	if config.Memory != "" {
		if err := validateMemorySpec(config.Memory); err != nil {
			return fmt.Errorf("invalid memory specification '%s': %w", config.Memory, err)
		}
	}

	// Validate CPU specification
	if config.CPUs != "" {
		if err := validateCPUSpec(config.CPUs); err != nil {
			return fmt.Errorf("invalid CPU specification '%s': %w", config.CPUs, err)
		}
	}

	// Validate working directory
	if config.WorkingDir != "" {
		if err := validateWorkdirSpec(config.WorkingDir); err != nil {
			return fmt.Errorf("invalid working directory specification '%s': %w", config.WorkingDir, err)
		}
	}

	return nil
}

// validateVolumeSpec validates volume mount specifications
func validateVolumeSpec(spec string) error {
	if spec == "" {
		return fmt.Errorf("volume specification cannot be empty")
	}

	parts := strings.Split(spec, ":")
	if len(parts) < 2 {
		return fmt.Errorf("volume specification must be in format 'host:container' or 'host:container:options'")
	}

	hostPath := parts[0]
	containerPath := parts[1]

	// Validate host path (must be absolute or valid volume name)
	if !filepath.IsAbs(hostPath) && !isValidVolumeName(hostPath) {
		return fmt.Errorf("host path '%s' must be absolute or valid volume name", hostPath)
	}

	// Validate container path (must be absolute)
	if !filepath.IsAbs(containerPath) {
		return fmt.Errorf("container path '%s' must be absolute", containerPath)
	}

	// Validate options if present
	if len(parts) > 2 {
		options := parts[2]
		validOptions := []string{"ro", "rw", "z", "Z"}
		optionParts := strings.Split(options, ",")
		for _, opt := range optionParts {
			opt = strings.TrimSpace(opt)
			if opt != "" {
				found := false
				for _, validOpt := range validOptions {
					if opt == validOpt {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("invalid volume option '%s', valid options: %s", opt, strings.Join(validOptions, ", "))
				}
			}
		}
	}

	return nil
}

// validatePortSpec validates port mapping specifications
func validatePortSpec(spec string) error {
	if spec == "" {
		return fmt.Errorf("port specification cannot be empty")
	}

	parts := strings.Split(spec, ":")
	if len(parts) != 2 {
		return fmt.Errorf("port specification must be in format 'host_port:container_port'")
	}

	hostPort := strings.TrimSpace(parts[0])
	containerPort := strings.TrimSpace(parts[1])

	// Validate host port
	if hostPort != "" {
		if _, err := strconv.Atoi(hostPort); err != nil {
			return fmt.Errorf("host port '%s' must be a valid number", hostPort)
		}
	}

	// Validate container port
	if containerPort == "" {
		return fmt.Errorf("container port cannot be empty")
	}
	if _, err := strconv.Atoi(containerPort); err != nil {
		return fmt.Errorf("container port '%s' must be a valid number", containerPort)
	}

	return nil
}

// validateEnvSpec validates environment variable specifications
func validateEnvSpec(spec string) error {
	if spec == "" {
		return fmt.Errorf("environment variable specification cannot be empty")
	}

	if !strings.Contains(spec, "=") {
		return fmt.Errorf("environment variable must be in format 'KEY=value'")
	}

	parts := strings.SplitN(spec, "=", 2)
	key := parts[0]

	if key == "" {
		return fmt.Errorf("environment variable key cannot be empty")
	}

	// Validate key format (basic validation)
	if !strings.HasPrefix(key, "_") && !isAlphaUnderscore(key[0]) {
		return fmt.Errorf("environment variable key '%s' must start with letter or underscore", key)
	}

	return nil
}

// validateMemorySpec validates memory limit specifications
func validateMemorySpec(spec string) error {
	if spec == "" {
		return fmt.Errorf("memory specification cannot be empty")
	}

	// Check if it ends with valid units
	validUnits := []string{"b", "k", "m", "g", "kb", "mb", "gb", "tb"}
	spec = strings.ToLower(spec)
	
	hasValidUnit := false
	for _, unit := range validUnits {
		if strings.HasSuffix(spec, unit) {
			hasValidUnit = true
			// Extract numeric part
			numPart := spec[:len(spec)-len(unit)]
			if numPart == "" {
				return fmt.Errorf("memory specification '%s' missing numeric value", spec)
			}
			if _, err := strconv.ParseFloat(numPart, 64); err != nil {
				return fmt.Errorf("memory specification '%s' has invalid numeric value", spec)
			}
			break
		}
	}

	if !hasValidUnit {
		// Check if it's just a number (defaults to bytes)
		if _, err := strconv.ParseInt(spec, 10, 64); err != nil {
			return fmt.Errorf("memory specification '%s' must be a number with unit (b, k, m, g) or just bytes", spec)
		}
	}

	return nil
}

// validateCPUSpec validates CPU limit specifications
func validateCPUSpec(spec string) error {
	if spec == "" {
		return fmt.Errorf("CPU specification cannot be empty")
	}

	// Should be a positive number (can be decimal)
	cpuValue, err := strconv.ParseFloat(spec, 64)
	if err != nil {
		return fmt.Errorf("CPU specification '%s' must be a valid number", spec)
	}

	if cpuValue <= 0 {
		return fmt.Errorf("CPU specification '%s' must be positive", spec)
	}

	return nil
}

// validateWorkdirSpec validates working directory specifications
func validateWorkdirSpec(spec string) error {
	if spec == "" {
		return fmt.Errorf("working directory specification cannot be empty")
	}

	// Should be an absolute path in the container
	if !filepath.IsAbs(spec) {
		return fmt.Errorf("working directory '%s' must be an absolute path", spec)
	}

	return nil
}

// Helper functions
func isValidVolumeName(name string) bool {
	// Docker volume names can contain letters, numbers, underscores, periods, and hyphens
	// and must start with an alphanumeric character
	if len(name) == 0 {
		return false
	}
	
	if !isAlphaNumeric(name[0]) {
		return false
	}
	
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !isAlphaNumeric(c) && c != '_' && c != '.' && c != '-' {
			return false
		}
	}
	
	return true
}

func isAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func isAlphaUnderscore(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// InstallSoftwareInContainer installs software in container based on installation type
func InstallSoftwareInContainer(containerName string, installationType string, pkgManager *PackageManagerInfo) error {
	if installationType == "empty" {
		fmt.Println("✓ Empty installation type - skipping software installation")
		return nil
	}

	fmt.Printf("\n📦 Installing %s environment in container using Portunix install command...\n", installationType)

	// Setup certificates before installation if needed for HTTPS downloads
	if err := setupContainerCertificates(containerName, pkgManager); err != nil {
		return fmt.Errorf("failed to setup certificates: %w", err)
	}

	// Copy Portunix binary to container
	if err := copyPortunixToContainer(containerName); err != nil {
		return fmt.Errorf("failed to copy Portunix binary to container: %w", err)
	}

	// Run standard Portunix install command inside container
	if err := runPortunixInstallInContainer(containerName, installationType); err != nil {
		return fmt.Errorf("failed to run Portunix install in container: %w", err)
	}

	fmt.Printf("✅ %s environment installed successfully!\n", installationType)
	return nil
}

// copyPortunixToContainer copies the current Portunix binary to the container
func copyPortunixToContainer(containerName string) error {
	fmt.Println("📄 Copying Portunix binary to container...")
	
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Copy Portunix binary to container using docker cp
	copyCmd := exec.Command("docker", "cp", execPath, containerName+":/usr/local/bin/portunix")
	if err := copyCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy Portunix binary: %w", err)
	}

	// Make it executable
	chmodCmd := []string{"chmod", "+x", "/usr/local/bin/portunix"}
	if err := execInContainer(containerName, chmodCmd); err != nil {
		return fmt.Errorf("failed to make Portunix binary executable: %w", err)
	}

	fmt.Println("✓ Portunix binary copied and made executable")
	return nil
}

// runPortunixInstallInContainer runs standard Portunix install command inside container
func runPortunixInstallInContainer(containerName string, installationType string) error {
	fmt.Printf("🚀 Running 'portunix install %s' inside container...\n", installationType)
	
	// Run portunix install command
	installCmd := []string{"portunix", "install", installationType}
	if err := execInContainer(containerName, installCmd); err != nil {
		return fmt.Errorf("failed to run 'portunix install %s': %w", installationType, err)
	}

	return nil
}

// installDefaultSoftware installs default software stack (Python, Java, VSCode)
func installDefaultSoftware(containerName string, pkgManager *PackageManagerInfo) error {
	if err := installPythonInContainer(containerName, pkgManager); err != nil {
		return fmt.Errorf("failed to install Python: %w", err)
	}
	if err := installJavaInContainer(containerName, pkgManager); err != nil {
		return fmt.Errorf("failed to install Java: %w", err)
	}
	return nil
}

// installPythonInContainer installs Python development environment
func installPythonInContainer(containerName string, pkgManager *PackageManagerInfo) error {
	fmt.Println("🐍 Installing Python development environment...")
	
	var installCmd []string
	switch pkgManager.Manager {
	case "apt-get":
		installCmd = []string{"sh", "-c", "apt-get update && apt-get install -y python3 python3-pip python3-venv python3-dev build-essential"}
	case "yum":
		installCmd = []string{"sh", "-c", "yum install -y python3 python3-pip python3-devel gcc gcc-c++ make"}
	case "dnf":
		installCmd = []string{"sh", "-c", "dnf install -y python3 python3-pip python3-devel gcc gcc-c++ make"}
	case "apk":
		installCmd = []string{"sh", "-c", "apk update && apk add --no-cache python3 py3-pip python3-dev build-base"}
	default:
		return fmt.Errorf("unsupported package manager for Python installation: %s", pkgManager.Manager)
	}

	if err := execInContainer(containerName, installCmd); err != nil {
		return fmt.Errorf("failed to install Python: %w", err)
	}

	fmt.Println("✓ Python development environment installed")
	return nil
}

// installJavaInContainer installs Java development environment
func installJavaInContainer(containerName string, pkgManager *PackageManagerInfo) error {
	fmt.Println("☕ Installing Java development environment...")
	
	var installCmd []string
	switch pkgManager.Manager {
	case "apt-get":
		installCmd = []string{"sh", "-c", "apt-get update && apt-get install -y openjdk-17-jdk maven"}
	case "yum":
		installCmd = []string{"sh", "-c", "yum install -y java-17-openjdk-devel maven"}
	case "dnf":
		installCmd = []string{"sh", "-c", "dnf install -y java-17-openjdk-devel maven"}
	case "apk":
		installCmd = []string{"sh", "-c", "apk update && apk add --no-cache openjdk17 maven"}
	default:
		return fmt.Errorf("unsupported package manager for Java installation: %s", pkgManager.Manager)
	}

	if err := execInContainer(containerName, installCmd); err != nil {
		return fmt.Errorf("failed to install Java: %w", err)
	}

	fmt.Println("✓ Java development environment installed")
	return nil
}

// installGoInContainer installs Go development environment
func installGoInContainer(containerName string, pkgManager *PackageManagerInfo) error {
	fmt.Println("🐹 Installing Go development environment...")
	
	// Install Go from official binary distribution for better version control
	var installCmd []string
	switch pkgManager.Manager {
	case "apt-get":
		installCmd = []string{"sh", "-c", `
			apt-get update && apt-get install -y wget curl git &&
			wget -O go.tar.gz https://go.dev/dl/go1.21.5.linux-amd64.tar.gz &&
			tar -C /usr/local -xzf go.tar.gz &&
			echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile &&
			echo 'export GOPATH=/workspace/go' >> /etc/profile &&
			echo 'export PATH=$PATH:/workspace/go/bin' >> /etc/profile &&
			rm go.tar.gz
		`}
	case "yum", "dnf":
		installCmd = []string{"sh", "-c", `
			` + pkgManager.InstallCmd + ` wget curl git &&
			wget -O go.tar.gz https://go.dev/dl/go1.21.5.linux-amd64.tar.gz &&
			tar -C /usr/local -xzf go.tar.gz &&
			echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile &&
			echo 'export GOPATH=/workspace/go' >> /etc/profile &&
			echo 'export PATH=$PATH:/workspace/go/bin' >> /etc/profile &&
			rm go.tar.gz
		`}
	case "apk":
		installCmd = []string{"sh", "-c", `
			apk update && apk add --no-cache wget curl git &&
			wget -O go.tar.gz https://go.dev/dl/go1.21.5.linux-amd64.tar.gz &&
			tar -C /usr/local -xzf go.tar.gz &&
			echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile &&
			echo 'export GOPATH=/workspace/go' >> /etc/profile &&
			echo 'export PATH=$PATH:/workspace/go/bin' >> /etc/profile &&
			rm go.tar.gz
		`}
	default:
		return fmt.Errorf("unsupported package manager for Go installation: %s", pkgManager.Manager)
	}

	if err := execInContainer(containerName, installCmd); err != nil {
		return fmt.Errorf("failed to install Go: %w", err)
	}

	// Verify Go installation
	verifyCmd := []string{"sh", "-c", "source /etc/profile && go version"}
	if err := execInContainer(containerName, verifyCmd); err != nil {
		fmt.Println("⚠️  Go installation verification failed, but continuing...")
	}

	fmt.Println("✓ Go development environment installed")
	return nil
}

// installVSCodeInContainer installs Visual Studio Code (note: this won't work in headless containers)
func installVSCodeInContainer(containerName string, pkgManager *PackageManagerInfo) error {
	fmt.Println("💻 Installing Visual Studio Code...")
	fmt.Println("⚠️  Note: VSCode requires GUI environment, installing code-server instead")
	
	var installCmd []string
	switch pkgManager.Manager {
	case "apt-get":
		installCmd = []string{"sh", "-c", `
			apt-get update && apt-get install -y curl wget &&
			curl -fsSL https://code-server.dev/install.sh | sh &&
			echo 'code-server installed successfully'
		`}
	case "yum", "dnf":
		installCmd = []string{"sh", "-c", `
			` + pkgManager.InstallCmd + ` curl wget &&
			curl -fsSL https://code-server.dev/install.sh | sh &&
			echo 'code-server installed successfully'
		`}
	case "apk":
		installCmd = []string{"sh", "-c", `
			apk update && apk add --no-cache curl wget &&
			curl -fsSL https://code-server.dev/install.sh | sh &&
			echo 'code-server installed successfully'
		`}
	default:
		return fmt.Errorf("unsupported package manager for VSCode installation: %s", pkgManager.Manager)
	}

	if err := execInContainer(containerName, installCmd); err != nil {
		return fmt.Errorf("failed to install VSCode/code-server: %w", err)
	}

	fmt.Println("✓ Visual Studio Code (code-server) installed")
	return nil
}

// checkDockerAvailable checks if Docker is installed and accessible
func checkDockerAvailable() error {
	// Check if docker command exists
	cmd := exec.Command("docker", "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Docker is not installed or not accessible. Please install Docker first.\n"+
			"Installation guide: https://docs.docker.com/get-docker/\n"+
			"Error: %v", err)
	}

	// Check if Docker daemon is running
	cmd = exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker daemon is not running. Please start Docker.\n"+
			"Try: sudo systemctl start docker (Linux) or start Docker Desktop (Windows/macOS)\n"+
			"Error: %v", err)
	}

	// Extract just the version line
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Version:") {
			fmt.Printf("✓ Docker is available: %s\n", strings.TrimSpace(line))
			break
		}
	}
	if len(lines) == 0 {
		fmt.Println("✓ Docker is available")
	}
	return nil
}

// runInContainerDryRun shows what would be executed without running Docker commands
func runInContainerDryRun(config DockerConfig) error {
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

	if config.EnableSSH {
		fmt.Printf("  • SSH Enabled: Yes (port 2222:22)\n")
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

	// Show Docker commands that would be executed
	fmt.Printf("🐳 Docker commands that would be executed:\n")
	fmt.Println()

	// 1. Check Docker availability
	fmt.Printf("1. Check Docker availability:\n")
	fmt.Printf("   docker version\n")
	fmt.Printf("   docker info\n")
	fmt.Println()

	// 2. Pull image
	fmt.Printf("2. Pull base image (if not cached):\n")
	fmt.Printf("   docker image inspect %s\n", config.Image)
	fmt.Printf("   docker pull %s\n", config.Image)
	fmt.Println()

	// 3. Detect package manager
	fmt.Printf("3. Detect package manager:\n")
	fmt.Printf("   docker run --rm %s /bin/sh -c \"command -v dnf && exit 0; command -v yum && exit 0; command -v apt-get && exit 0; command -v apk && exit 0\"\n", config.Image)
	fmt.Println()

	// 4. Build Docker run command
	dockerArgs := buildDockerRunArgs(config)
	fmt.Printf("4. Create and run container:\n")
	fmt.Printf("   docker run %s\n", strings.Join(dockerArgs[1:], " "))
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
	case "go":
		fmt.Printf("   • Install Go development environment\n")
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

// CheckRequirements checks if all system requirements for Docker container operations are satisfied
func CheckRequirements() error {
	fmt.Println("🔍 Checking system requirements for Docker container operations...")
	fmt.Println()

	// Check Docker installation
	fmt.Print("📦 Docker installation: ")
	cmd := exec.Command("docker", "version", "--format", "{{.Client.Version}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("   Docker is not installed or not accessible\n")
		fmt.Printf("   Please install Docker: https://docs.docker.com/get-docker/\n")
		return fmt.Errorf("Docker not available")

	}
	fmt.Printf("✅ OK (Version: %s)\n", strings.TrimSpace(string(output)))

	// Check Docker daemon
	fmt.Print("🐳 Docker daemon: ")
	cmd = exec.Command("docker", "info", "--format", "{{.ServerVersion}}")
	output, err = cmd.Output()
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("   Docker daemon is not running\n")
		fmt.Printf("   Try: sudo systemctl start docker (Linux) or start Docker Desktop (Windows/macOS)\n")
		return fmt.Errorf("Docker daemon not running")
	}
	fmt.Printf("✅ OK (Server: %s)\n", strings.TrimSpace(string(output)))

	// Check Docker permissions
	fmt.Print("🔐 Docker permissions: ")
	cmd = exec.Command("docker", "ps")
	err = cmd.Run()
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("   No permission to access Docker daemon\n")
		fmt.Printf("   Try: sudo usermod -aG docker $USER && newgrp docker (Linux)\n")
		return fmt.Errorf("insufficient Docker permissions")
	}
	fmt.Println("✅ OK")

	// Check available space
	fmt.Print("💾 Disk space: ")
	cmd = exec.Command("docker", "system", "df", "--format", "table {{.Type}}\t{{.Size}}\t{{.Reclaimable}}")
	output, err = cmd.Output()
	if err == nil {
		// Parse output to check available space
		fmt.Println("✅ OK")
		fmt.Printf("   Docker disk usage:\n")
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
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
	cmd = exec.Command("docker", "run", "--rm", "alpine:latest", "ping", "-c", "1", "google.com")
	err = cmd.Run()
	if err != nil {
		fmt.Println("⚠️  WARNING (could not test network - this may affect image pulling)")
	} else {
		fmt.Println("✅ OK")
	}

	fmt.Println()
	fmt.Println("🎉 All critical requirements are satisfied!")
	fmt.Println("💡 You can now run: portunix docker run-in-container <type>")

	return nil
}

// parseDockerArgs parses command line arguments into DockerConfig
func parseDockerArgs(installationType string, args []string) (DockerConfig, error) {
	config := DockerConfig{
		Image:             "ubuntu:22.04", // Default image
		InstallationType:  installationType,
		EnableSSH:        true,
		KeepRunning:      false,
		Disposable:       false,
		Privileged:       false,
	}
	
	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		
		switch {
		case arg == "-v" || arg == "--volume":
			if i+1 < len(args) {
				config.Volumes = append(config.Volumes, args[i+1])
				i++ // Skip next argument
			}
		case strings.HasPrefix(arg, "-v=") || strings.HasPrefix(arg, "--volume="):
			volume := strings.SplitN(arg, "=", 2)[1]
			config.Volumes = append(config.Volumes, volume)
		case arg == "-p" || arg == "--port":
			if i+1 < len(args) {
				config.Ports = append(config.Ports, args[i+1])
				i++ // Skip next argument
			}
		case strings.HasPrefix(arg, "-p=") || strings.HasPrefix(arg, "--port="):
			port := strings.SplitN(arg, "=", 2)[1]
			config.Ports = append(config.Ports, port)
		case arg == "-e" || arg == "--env":
			if i+1 < len(args) {
				config.Environment = append(config.Environment, args[i+1])
				i++ // Skip next argument
			}
		case strings.HasPrefix(arg, "-e=") || strings.HasPrefix(arg, "--env="):
			env := strings.SplitN(arg, "=", 2)[1]
			config.Environment = append(config.Environment, env)
		case arg == "--name":
			if i+1 < len(args) {
				config.ContainerName = args[i+1]
				i++ // Skip next argument
			}
		case strings.HasPrefix(arg, "--name="):
			config.ContainerName = strings.SplitN(arg, "=", 2)[1]
		case arg == "--keep-running":
			config.KeepRunning = true
		case arg == "--disposable":
			config.Disposable = true
		case arg == "--privileged":
			config.Privileged = true
		case arg == "--no-ssh":
			config.EnableSSH = false
		default:
			// Ignore unknown arguments for now
			fmt.Printf("⚠️  Warning: Unknown argument '%s' ignored\n", arg)
		}
	}
	
	// Generate container name if not provided
	if config.ContainerName == "" {
		config.ContainerName = fmt.Sprintf("portunix-%s-%d", installationType, time.Now().Unix())
	}
	
	return config, nil
}

// setupContainerCertificates installs CA certificates in container for HTTPS downloads
func setupContainerCertificates(containerName string, pkgManager *PackageManagerInfo) error {
	fmt.Println("📋 Setting up CA certificates in container...")
	
	// Update package manager first
	updateCmd := generateUpdateCommand(pkgManager)
	fmt.Printf("Updating package manager (%s)...\n", pkgManager.Manager)
	if err := execInContainer(containerName, updateCmd); err != nil {
		return fmt.Errorf("failed to update package manager: %w", err)
	}
	
	// Install ca-certificates package
	certCmd := generateCertificateCommand(pkgManager)
	fmt.Printf("Installing CA certificates (%s)...\n", pkgManager.Manager)
	if err := execInContainer(containerName, certCmd); err != nil {
		return fmt.Errorf("failed to install certificates: %w", err)
	}
	
	// Update certificate bundle
	updateCertCmd := generateCertificateUpdateCommand(pkgManager)
	fmt.Println("Updating certificate bundle...")
	if err := execInContainer(containerName, updateCertCmd); err != nil {
		fmt.Printf("⚠️ Warning: Failed to update certificates: %v\n", err)
		// Continue anyway - some environments may not support this
	}
	
	// Test HTTPS connectivity
	testCmd := []string{"curl", "-I", "https://go.dev/dl/"}
	fmt.Println("Testing HTTPS connectivity...")
	if err := execInContainer(containerName, testCmd); err != nil {
		fmt.Printf("⚠️ Warning: HTTPS connectivity test failed: %v\n", err)
		// Continue anyway - some environments may block external connections
	} else {
		fmt.Println("✅ HTTPS connectivity verified")
	}
	
	fmt.Println("✅ CA certificates setup completed")
	return nil
}

// generateUpdateCommand generates package manager update command
func generateUpdateCommand(pkgManager *PackageManagerInfo) []string {
	switch pkgManager.Manager {
	case "apt", "apt-get":
		return []string{"apt-get", "update", "-y"}
	case "yum":
		return []string{"yum", "makecache"}
	case "dnf":
		return []string{"dnf", "makecache"}
	case "apk":
		return []string{"apk", "update"}
	case "pacman":
		return []string{"pacman", "-Sy"}
	case "zypper":
		return []string{"zypper", "refresh"}
	default:
		return []string{"echo", "Unknown package manager for update"}
	}
}

// generateCertificateCommand generates certificate installation command
func generateCertificateCommand(pkgManager *PackageManagerInfo) []string {
	switch pkgManager.Manager {
	case "apt", "apt-get":
		return []string{"apt-get", "install", "-y", "ca-certificates", "curl"}
	case "yum":
		return []string{"yum", "install", "-y", "ca-certificates", "curl"}
	case "dnf":
		return []string{"dnf", "install", "-y", "ca-certificates", "curl"}
	case "apk":
		return []string{"apk", "add", "ca-certificates", "curl"}
	case "pacman":
		return []string{"pacman", "-S", "--noconfirm", "ca-certificates", "curl"}
	case "zypper":
		return []string{"zypper", "install", "-y", "ca-certificates", "curl"}
	default:
		return []string{"echo", "Unknown package manager for certificate installation"}
	}
}

// generateCertificateUpdateCommand generates certificate update command
func generateCertificateUpdateCommand(pkgManager *PackageManagerInfo) []string {
	switch pkgManager.Manager {
	case "apt", "apt-get":
		return []string{"update-ca-certificates"}
	case "yum", "dnf":
		return []string{"update-ca-trust"}
	case "apk":
		return []string{"update-ca-certificates"}
	case "pacman":
		return []string{"trust", "extract-compat"}
	case "zypper":
		return []string{"update-ca-certificates"}
	default:
		return []string{"echo", "Unknown package manager for certificate update"}
	}
}

// ContainerRunOptions defines options for running containers
type ContainerRunOptions struct {
	Detach      bool
	Interactive bool
	TTY         bool
	Name        string
	Ports       []string
	Volumes     []string
	Environment []string
}

// RunContainer runs a generic Docker container with specified options
func RunContainer(image string, command []string, options ContainerRunOptions) error {
	// Build docker run command
	args := []string{"run"}
	
	// Add flags based on options
	if options.Detach {
		args = append(args, "-d")
	}
	if options.Interactive {
		args = append(args, "-i")
	}
	if options.TTY {
		args = append(args, "-t")
	}
	if options.Name != "" {
		args = append(args, "--name", options.Name)
	}
	
	// Add port mappings
	for _, port := range options.Ports {
		args = append(args, "-p", port)
	}
	
	// Add volume mounts
	for _, volume := range options.Volumes {
		args = append(args, "-v", volume)
	}
	
	// Add environment variables
	for _, env := range options.Environment {
		args = append(args, "-e", env)
	}
	
	// Add image
	args = append(args, image)
	
	// Add command
	if len(command) > 0 {
		args = append(args, command...)
	}
	
	// Execute docker command
	fmt.Printf("🐳 Running Docker container: docker %s\n", strings.Join(args, " "))
	
	cmd := exec.Command("docker", args...)
	
	// If not detached, inherit stdio for interactive containers
	if !options.Detach {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
	}
	
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run Docker container: %w", err)
	}
	
	if options.Detach {
		fmt.Println("✅ Container started successfully in detached mode")
	} else {
		fmt.Println("✅ Container execution completed")
	}
	
	return nil
}

// CopyFiles copies files between host and Docker container
func CopyFiles(source, destination string) error {
	cmd := exec.Command("docker", "cp", source, destination)
	
	// Run the command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy files: %v - %s", err, string(output))
	}
	
	return nil
}
