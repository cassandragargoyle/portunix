package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DockerInstaller handles Docker installation on various platforms
type DockerInstaller struct {
	storage *StorageAnalyzer
	dryRun  bool
}

// NewDockerInstaller creates a new Docker installer instance
func NewDockerInstaller(dryRun bool) *DockerInstaller {
	return &DockerInstaller{
		storage: NewStorageAnalyzer(10), // 10 GB minimum for Docker
		dryRun:  dryRun,
	}
}

// Install performs Docker installation based on current platform
func (d *DockerInstaller) Install() error {
	fmt.Println("🐳 Starting Docker installation with intelligent storage detection...")

	// Check if Docker is already installed
	if d.isDockerInstalled() {
		fmt.Println("✅ Docker is already installed")
		return d.verifyInstallation()
	}

	switch runtime.GOOS {
	case "windows":
		return d.installWindows()
	case "linux":
		return d.installLinux()
	case "darwin":
		return d.installMacOS()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func (d *DockerInstaller) isDockerInstalled() bool {
	cmd := exec.Command("docker", "--version")
	return cmd.Run() == nil
}

func (d *DockerInstaller) verifyInstallation() error {
	fmt.Println("\n🔍 Verifying Docker installation...")

	// Check version
	cmd := exec.Command("docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("docker --version failed: %w", err)
	}
	fmt.Printf("✅ %s", string(output))

	// Check daemon
	cmd = exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		fmt.Println("⚠️  Docker daemon may not be running")
		fmt.Println("   Start Docker Desktop or run: sudo systemctl start docker")
	} else {
		fmt.Println("✅ Docker daemon is running")
	}

	return nil
}

// installWindows installs Docker Desktop on Windows
func (d *DockerInstaller) installWindows() error {
	fmt.Println("\n📊 Analyzing available storage...")

	drives, err := d.storage.GetWindowsDrives()
	if err != nil {
		return fmt.Errorf("failed to analyze storage: %w", err)
	}

	// Display storage options
	fmt.Println("\n💾 Available drives:")
	for _, drive := range drives {
		status := ""
		spaceBytes := parseSpaceString(drive.FreeSpace)
		if spaceBytes < d.storage.minSpace {
			status = " ⚠️ (insufficient space)"
		}
		fmt.Printf("   %s:\\ - %s free / %s total%s\n", drive.Letter, drive.FreeSpace, drive.TotalSpace, status)
	}

	// Get recommended drive
	selectedDrive, err := d.storage.AnalyzeStorage()
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Selected storage: %s:\\ (optimal choice)\n", selectedDrive)

	if d.dryRun {
		fmt.Println("\n🔍 DRY RUN - Would perform the following:")
		fmt.Printf("   1. Download Docker Desktop installer\n")
		fmt.Printf("   2. Install Docker Desktop with data-root: %s:\\docker-data\n", selectedDrive)
		fmt.Printf("   3. Configure Docker settings\n")
		fmt.Printf("   4. Verify installation\n")
		return nil
	}

	// Download Docker Desktop
	fmt.Println("\n📥 Downloading Docker Desktop for Windows...")
	dockerURL := "https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe"

	installerPath, err := DownloadFileWithProperFilename(dockerURL, os.TempDir())
	if err != nil {
		return fmt.Errorf("failed to download Docker Desktop: %w", err)
	}

	// Verify the downloaded file exists
	if _, err := os.Stat(installerPath); os.IsNotExist(err) {
		return fmt.Errorf("downloaded installer not found at: %s", installerPath)
	}

	// Run installer
	fmt.Println("🔧 Installing Docker Desktop...")
	dataRoot := fmt.Sprintf("%s:\\docker-data", selectedDrive)

	// Docker Desktop installer arguments
	args := []string{
		"install",
		"--accept-license",
		"--quiet",
	}

	cmd := exec.Command(installerPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker Desktop installation failed: %w", err)
	}

	// Configure data-root if not on C:
	if selectedDrive != "C" {
		fmt.Printf("⚙️  Configuring Docker data-root to: %s\n", dataRoot)
		if err := d.configureWindowsDataRoot(dataRoot); err != nil {
			fmt.Printf("⚠️  Could not configure data-root: %v\n", err)
			fmt.Printf("   You can manually set data-root in Docker Desktop settings\n")
		}
	}

	fmt.Println("\n✅ Docker Desktop installed successfully!")

	// Try to start Docker Desktop automatically
	fmt.Println("🚀 Starting Docker Desktop...")
	startCmd := exec.Command("cmd", "/C", "start", "", "Docker Desktop")
	startCmd.Run() // Ignore errors

	// Wait for Docker daemon to become available (up to 120 seconds)
	fmt.Println("⏳ Waiting for Docker daemon to start...")
	for i := 0; i < 24; i++ {
		time.Sleep(5 * time.Second)

		checkCmd := exec.Command("docker", "version")
		if checkCmd.Run() == nil {
			fmt.Println("✅ Docker daemon is running and ready!")
			return nil
		}
		fmt.Printf("   Waiting... (%d/120s)\n", (i+1)*5)
	}

	// Docker didn't start within timeout
	fmt.Println("\n⚠️  Docker Desktop is starting but daemon is not ready yet.")
	fmt.Println("   You may need to wait a bit longer or restart your computer.")

	return nil
}

func (d *DockerInstaller) configureWindowsDataRoot(dataRoot string) error {
	// Create data directory
	if err := os.MkdirAll(dataRoot, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Docker Desktop config is in %USERPROFILE%\.docker\daemon.json
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".docker")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "daemon.json")

	// Write daemon.json with data-root
	config := fmt.Sprintf(`{
  "data-root": "%s"
}`, strings.ReplaceAll(dataRoot, "\\", "\\\\"))

	return os.WriteFile(configPath, []byte(config), 0644)
}

// installLinux installs Docker Engine on Linux
func (d *DockerInstaller) installLinux() error {
	fmt.Println("\n📊 Analyzing available storage...")

	partitions, err := d.storage.GetLinuxPartitions()
	if err != nil {
		return fmt.Errorf("failed to analyze storage: %w", err)
	}

	// Display storage options
	fmt.Println("\n💾 Available partitions:")
	for _, part := range partitions {
		status := ""
		spaceBytes := parseSpaceString(part.FreeSpace)
		if spaceBytes < d.storage.minSpace {
			status = " ⚠️ (insufficient space)"
		}
		fmt.Printf("   %s - %s free / %s total%s\n", part.MountPoint, part.FreeSpace, part.TotalSpace, status)
	}

	// Get recommended partition
	selectedPath, err := d.storage.AnalyzeStorage()
	if err != nil {
		return err
	}

	dataRoot := filepath.Join(selectedPath, "docker-data")
	fmt.Printf("\n✅ Selected storage: %s (optimal choice)\n", dataRoot)

	if d.dryRun {
		fmt.Println("\n🔍 DRY RUN - Would perform the following:")
		fmt.Println("   1. Install Docker prerequisites")
		fmt.Println("   2. Add Docker repository")
		fmt.Println("   3. Install Docker Engine")
		fmt.Printf("   4. Configure data-root: %s\n", dataRoot)
		fmt.Println("   5. Start Docker service")
		fmt.Println("   6. Add current user to docker group")
		return nil
	}

	// Detect distribution
	distro := d.detectLinuxDistro()
	fmt.Printf("🐧 Detected distribution: %s\n", distro)

	switch distro {
	case "ubuntu", "debian":
		return d.installDockerUbuntuDebian(dataRoot)
	case "fedora", "centos", "rhel", "rocky":
		return d.installDockerFedoraCentOS(dataRoot)
	case "arch":
		return d.installDockerArch(dataRoot)
	default:
		return d.installDockerGeneric(dataRoot)
	}
}

func (d *DockerInstaller) detectLinuxDistro() string {
	// Try /etc/os-release first
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		content := strings.ToLower(string(data))
		if strings.Contains(content, "ubuntu") {
			return "ubuntu"
		}
		if strings.Contains(content, "debian") {
			return "debian"
		}
		if strings.Contains(content, "fedora") {
			return "fedora"
		}
		if strings.Contains(content, "centos") {
			return "centos"
		}
		if strings.Contains(content, "rhel") || strings.Contains(content, "red hat") {
			return "rhel"
		}
		if strings.Contains(content, "rocky") {
			return "rocky"
		}
		if strings.Contains(content, "arch") {
			return "arch"
		}
	}

	return "unknown"
}

func (d *DockerInstaller) installDockerUbuntuDebian(dataRoot string) error {
	commands := [][]string{
		{"sudo", "apt-get", "update"},
		{"sudo", "apt-get", "install", "-y", "ca-certificates", "curl", "gnupg"},
		{"sudo", "install", "-m", "0755", "-d", "/etc/apt/keyrings"},
	}

	// Add Docker GPG key
	fmt.Println("🔑 Adding Docker GPG key...")
	gpgCmd := exec.Command("bash", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg")
	if err := gpgCmd.Run(); err != nil {
		fmt.Printf("⚠️  GPG key setup may have issues: %v\n", err)
	}

	// Add repository
	fmt.Println("📦 Adding Docker repository...")
	repoCmd := exec.Command("bash", "-c", `echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null`)
	if err := repoCmd.Run(); err != nil {
		fmt.Printf("⚠️  Repository setup may have issues: %v\n", err)
	}

	commands = append(commands,
		[]string{"sudo", "apt-get", "update"},
		[]string{"sudo", "apt-get", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"},
	)

	for _, cmdArgs := range commands {
		fmt.Printf("🔧 Running: %s\n", strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
	}

	// Configure data-root if not default location
	if dataRoot != "/var/lib/docker" {
		if err := d.configureLinuxDataRoot(dataRoot); err != nil {
			fmt.Printf("⚠️  Could not configure data-root: %v\n", err)
		}
	}

	// Add user to docker group
	d.addUserToDockerGroup()

	// Start and enable Docker
	exec.Command("sudo", "systemctl", "start", "docker").Run()
	exec.Command("sudo", "systemctl", "enable", "docker").Run()

	fmt.Println("\n✅ Docker Engine installed successfully!")
	return nil
}

func (d *DockerInstaller) installDockerFedoraCentOS(dataRoot string) error {
	commands := [][]string{
		{"sudo", "dnf", "-y", "install", "dnf-plugins-core"},
		{"sudo", "dnf", "config-manager", "--add-repo", "https://download.docker.com/linux/fedora/docker-ce.repo"},
		{"sudo", "dnf", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"},
	}

	for _, cmdArgs := range commands {
		fmt.Printf("🔧 Running: %s\n", strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// Try yum as fallback
			if cmdArgs[1] == "dnf" {
				cmdArgs[1] = "yum"
				cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("command failed: %w", err)
				}
			} else {
				return fmt.Errorf("command failed: %w", err)
			}
		}
	}

	if dataRoot != "/var/lib/docker" {
		d.configureLinuxDataRoot(dataRoot)
	}

	d.addUserToDockerGroup()
	exec.Command("sudo", "systemctl", "start", "docker").Run()
	exec.Command("sudo", "systemctl", "enable", "docker").Run()

	fmt.Println("\n✅ Docker Engine installed successfully!")
	return nil
}

func (d *DockerInstaller) installDockerArch(dataRoot string) error {
	commands := [][]string{
		{"sudo", "pacman", "-Sy", "--noconfirm", "docker"},
	}

	for _, cmdArgs := range commands {
		fmt.Printf("🔧 Running: %s\n", strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
	}

	if dataRoot != "/var/lib/docker" {
		d.configureLinuxDataRoot(dataRoot)
	}

	d.addUserToDockerGroup()
	exec.Command("sudo", "systemctl", "start", "docker").Run()
	exec.Command("sudo", "systemctl", "enable", "docker").Run()

	fmt.Println("\n✅ Docker Engine installed successfully!")
	return nil
}

func (d *DockerInstaller) installDockerGeneric(dataRoot string) error {
	// Use official Docker installation script
	fmt.Println("🔧 Using Docker's official installation script...")

	cmd := exec.Command("bash", "-c", "curl -fsSL https://get.docker.com | sudo sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker installation script failed: %w", err)
	}

	if dataRoot != "/var/lib/docker" {
		d.configureLinuxDataRoot(dataRoot)
	}

	d.addUserToDockerGroup()
	exec.Command("sudo", "systemctl", "start", "docker").Run()
	exec.Command("sudo", "systemctl", "enable", "docker").Run()

	fmt.Println("\n✅ Docker Engine installed successfully!")
	return nil
}

func (d *DockerInstaller) configureLinuxDataRoot(dataRoot string) error {
	// Create data directory
	if err := exec.Command("sudo", "mkdir", "-p", dataRoot).Run(); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create daemon.json
	config := fmt.Sprintf(`{
  "data-root": "%s"
}`, dataRoot)

	tmpFile := filepath.Join(os.TempDir(), "docker-daemon.json")
	if err := os.WriteFile(tmpFile, []byte(config), 0644); err != nil {
		return err
	}

	// Move to /etc/docker/daemon.json
	exec.Command("sudo", "mkdir", "-p", "/etc/docker").Run()
	return exec.Command("sudo", "mv", tmpFile, "/etc/docker/daemon.json").Run()
}

func (d *DockerInstaller) addUserToDockerGroup() {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("LOGNAME")
	}

	if user != "" && user != "root" {
		fmt.Printf("👤 Adding user '%s' to docker group...\n", user)
		exec.Command("sudo", "usermod", "-aG", "docker", user).Run()
		fmt.Println("   Log out and back in for group changes to take effect")
	}
}

// installMacOS installs Docker Desktop on macOS
func (d *DockerInstaller) installMacOS() error {
	fmt.Println("\n🍎 Installing Docker Desktop for macOS...")

	if d.dryRun {
		fmt.Println("\n🔍 DRY RUN - Would perform the following:")
		fmt.Println("   1. Check for Homebrew")
		fmt.Println("   2. Install Docker Desktop via brew cask")
		return nil
	}

	// Check for Homebrew
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew is required to install Docker Desktop on macOS. Install from https://brew.sh")
	}

	cmd := exec.Command("brew", "install", "--cask", "docker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker Desktop installation failed: %w", err)
	}

	fmt.Println("\n✅ Docker Desktop installed successfully!")
	fmt.Println("   Open Docker Desktop from Applications to complete setup")

	return nil
}
