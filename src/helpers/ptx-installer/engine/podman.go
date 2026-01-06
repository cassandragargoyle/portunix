package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// PodmanInstaller handles Podman installation on various platforms
type PodmanInstaller struct {
	storage *StorageAnalyzer
	dryRun  bool
}

// NewPodmanInstaller creates a new Podman installer instance
func NewPodmanInstaller(dryRun bool) *PodmanInstaller {
	return &PodmanInstaller{
		storage: NewStorageAnalyzer(5), // 5 GB minimum for Podman (less than Docker)
		dryRun:  dryRun,
	}
}

// Install performs Podman installation based on current platform
func (p *PodmanInstaller) Install() error {
	fmt.Println("🦭 Starting Podman installation...")

	// Check if Podman is already installed
	if p.isPodmanInstalled() {
		fmt.Println("✅ Podman is already installed")
		return p.verifyInstallation()
	}

	switch runtime.GOOS {
	case "windows":
		return p.installWindows()
	case "linux":
		return p.installLinux()
	case "darwin":
		return p.installMacOS()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func (p *PodmanInstaller) isPodmanInstalled() bool {
	cmd := exec.Command("podman", "--version")
	return cmd.Run() == nil
}

func (p *PodmanInstaller) verifyInstallation() error {
	fmt.Println("\n🔍 Verifying Podman installation...")

	// Check version
	cmd := exec.Command("podman", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("podman --version failed: %w", err)
	}
	fmt.Printf("✅ %s", string(output))

	// Check machine status on Windows/macOS
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		cmd = exec.Command("podman", "machine", "list")
		if output, err := cmd.Output(); err == nil {
			fmt.Printf("📋 Podman machines:\n%s", string(output))
		}
	}

	return nil
}

// installWindows installs Podman Desktop on Windows
func (p *PodmanInstaller) installWindows() error {
	fmt.Println("\n📊 Checking system requirements...")

	if p.dryRun {
		fmt.Println("\n🔍 DRY RUN - Would perform the following:")
		fmt.Println("   1. Download Podman Desktop installer")
		fmt.Println("   2. Install Podman Desktop")
		fmt.Println("   3. Initialize Podman machine")
		return nil
	}

	// Try WinGet first
	if _, err := exec.LookPath("winget"); err == nil {
		fmt.Println("📦 Installing Podman Desktop via WinGet...")
		cmd := exec.Command("winget", "install", "--id", "RedHat.Podman-Desktop", "-e", "--accept-package-agreements", "--accept-source-agreements")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			fmt.Println("\n✅ Podman Desktop installed successfully!")
			return p.initializePodmanMachine()
		}
		fmt.Println("⚠️  WinGet installation failed, trying direct download...")
	}

	// Download Podman Desktop installer
	fmt.Println("📥 Downloading Podman Desktop for Windows...")
	installerURL := "https://github.com/containers/podman-desktop/releases/latest/download/podman-desktop-setup.exe"
	installerPath := filepath.Join(os.TempDir(), "podman-desktop-setup.exe")

	if _, err := DownloadFileWithProperFilename(installerURL, os.TempDir()); err != nil {
		return fmt.Errorf("failed to download Podman Desktop: %w", err)
	}

	// Run installer
	fmt.Println("🔧 Installing Podman Desktop...")
	cmd := exec.Command(installerPath, "/S") // Silent install
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Podman Desktop installation failed: %w", err)
	}

	fmt.Println("\n✅ Podman Desktop installed successfully!")
	return p.initializePodmanMachine()
}

func (p *PodmanInstaller) initializePodmanMachine() error {
	fmt.Println("\n🔧 Initializing Podman machine...")

	// Check if machine already exists
	cmd := exec.Command("podman", "machine", "list", "--format", "{{.Name}}")
	output, _ := cmd.Output()
	if strings.TrimSpace(string(output)) != "" {
		fmt.Println("   Podman machine already exists")
		return nil
	}

	// Initialize machine
	cmd = exec.Command("podman", "machine", "init")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("⚠️  Machine initialization may require restart")
		return nil
	}

	// Start machine
	cmd = exec.Command("podman", "machine", "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	return nil
}

// installLinux installs Podman on Linux
func (p *PodmanInstaller) installLinux() error {
	fmt.Println("\n🐧 Installing Podman for Linux...")

	if p.dryRun {
		fmt.Println("\n🔍 DRY RUN - Would perform the following:")
		fmt.Println("   1. Detect Linux distribution")
		fmt.Println("   2. Install Podman via package manager")
		return nil
	}

	// Detect distribution
	distro := p.detectLinuxDistro()
	fmt.Printf("🐧 Detected distribution: %s\n", distro)

	switch distro {
	case "ubuntu", "debian":
		return p.installPodmanUbuntuDebian()
	case "fedora", "centos", "rhel", "rocky":
		return p.installPodmanFedoraCentOS()
	case "arch":
		return p.installPodmanArch()
	default:
		return p.installPodmanGeneric()
	}
}

func (p *PodmanInstaller) detectLinuxDistro() string {
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

func (p *PodmanInstaller) installPodmanUbuntuDebian() error {
	commands := [][]string{
		{"sudo", "apt-get", "update"},
		{"sudo", "apt-get", "install", "-y", "podman"},
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

	fmt.Println("\n✅ Podman installed successfully!")
	return nil
}

func (p *PodmanInstaller) installPodmanFedoraCentOS() error {
	commands := [][]string{
		{"sudo", "dnf", "install", "-y", "podman"},
	}

	for _, cmdArgs := range commands {
		fmt.Printf("🔧 Running: %s\n", strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// Try yum as fallback
			cmdArgs[1] = "yum"
			cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("command failed: %w", err)
			}
		}
	}

	fmt.Println("\n✅ Podman installed successfully!")
	return nil
}

func (p *PodmanInstaller) installPodmanArch() error {
	cmd := exec.Command("sudo", "pacman", "-Sy", "--noconfirm", "podman")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("podman installation failed: %w", err)
	}

	fmt.Println("\n✅ Podman installed successfully!")
	return nil
}

func (p *PodmanInstaller) installPodmanGeneric() error {
	// Try common package managers
	packageManagers := []struct {
		check   string
		install []string
	}{
		{"apt-get", []string{"sudo", "apt-get", "install", "-y", "podman"}},
		{"dnf", []string{"sudo", "dnf", "install", "-y", "podman"}},
		{"yum", []string{"sudo", "yum", "install", "-y", "podman"}},
		{"pacman", []string{"sudo", "pacman", "-Sy", "--noconfirm", "podman"}},
		{"zypper", []string{"sudo", "zypper", "install", "-y", "podman"}},
	}

	for _, pm := range packageManagers {
		if _, err := exec.LookPath(pm.check); err == nil {
			fmt.Printf("🔧 Using %s to install Podman...\n", pm.check)
			cmd := exec.Command(pm.install[0], pm.install[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err == nil {
				fmt.Println("\n✅ Podman installed successfully!")
				return nil
			}
		}
	}

	return fmt.Errorf("no supported package manager found. Please install Podman manually")
}

// installMacOS installs Podman on macOS
func (p *PodmanInstaller) installMacOS() error {
	fmt.Println("\n🍎 Installing Podman for macOS...")

	if p.dryRun {
		fmt.Println("\n🔍 DRY RUN - Would perform the following:")
		fmt.Println("   1. Check for Homebrew")
		fmt.Println("   2. Install Podman via brew")
		fmt.Println("   3. Initialize Podman machine")
		return nil
	}

	// Check for Homebrew
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew is required to install Podman on macOS. Install from https://brew.sh")
	}

	// Install Podman
	cmd := exec.Command("brew", "install", "podman")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Podman installation failed: %w", err)
	}

	fmt.Println("\n✅ Podman installed successfully!")
	return p.initializePodmanMachine()
}
