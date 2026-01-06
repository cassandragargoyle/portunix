package engine

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// AddAptRepository adds a third-party APT repository with optional GPG key
func AddAptRepository(repository string, keyUrl string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("APT repositories are only available on Linux")
	}

	needsSudo := !IsRunningAsRoot()
	if needsSudo && !IsSudoAvailable() {
		return fmt.Errorf("sudo is required to add APT repository")
	}

	// Download and install GPG key if provided
	if keyUrl != "" {
		fmt.Printf("🔑 Adding GPG key from: %s\n", keyUrl)

		// Extract keyring filename from URL or use default
		keyringPath := "/usr/share/keyrings/portunix-added-keyring.gpg"
		if strings.Contains(keyUrl, "githubcli") {
			keyringPath = "/usr/share/keyrings/githubcli-archive-keyring.gpg"
		}

		// Download key to temp file first
		tmpFile := "/tmp/portunix-key.gpg"
		fmt.Printf("🚀 Downloading key...\n")
		curlCmd := exec.Command("curl", "-fsSL", "-o", tmpFile, keyUrl)
		curlCmd.Stdout = os.Stdout
		curlCmd.Stderr = os.Stderr
		if err := curlCmd.Run(); err != nil {
			return fmt.Errorf("failed to download GPG key: %w", err)
		}

		// Dearmor and install key with sudo
		fmt.Printf("🚀 Installing GPG key to %s\n", keyringPath)
		var gpgCmd *exec.Cmd
		if needsSudo {
			gpgCmd = exec.Command("sudo", "bash", "-c", fmt.Sprintf("gpg --dearmor -o %s < %s", keyringPath, tmpFile))
		} else {
			gpgCmd = exec.Command("bash", "-c", fmt.Sprintf("gpg --dearmor -o %s < %s", keyringPath, tmpFile))
		}
		gpgCmd.Stdin = os.Stdin
		gpgCmd.Stdout = os.Stdout
		gpgCmd.Stderr = os.Stderr
		if err := gpgCmd.Run(); err != nil {
			return fmt.Errorf("failed to install GPG key: %w", err)
		}

		// Cleanup temp file
		os.Remove(tmpFile)
	}

	// Add repository to sources.list.d
	fmt.Printf("📦 Adding APT repository: %s\n", repository)

	// Create sources list file
	repoFile := "/etc/apt/sources.list.d/portunix-added.list"
	if strings.Contains(repository, "github") {
		repoFile = "/etc/apt/sources.list.d/github-cli.list"
	}

	// Write repository file
	var teeCmd *exec.Cmd
	if needsSudo {
		teeCmd = exec.Command("sudo", "tee", repoFile)
	} else {
		teeCmd = exec.Command("tee", repoFile)
	}
	teeCmd.Stdin = strings.NewReader(repository + "\n")
	teeCmd.Stdout = nil // Suppress tee output
	teeCmd.Stderr = os.Stderr
	if err := teeCmd.Run(); err != nil {
		return fmt.Errorf("failed to add repository: %w", err)
	}

	fmt.Println("✅ Repository added successfully")
	return nil
}

// InstallViaAPT installs packages using APT package manager
func InstallViaAPT(packages []string, requiresSudo bool) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("APT is only available on Linux")
	}

	fmt.Printf("📦 Installing via APT: %v\n", packages)

	needsSudo := requiresSudo && !IsRunningAsRoot()
	if needsSudo && !IsSudoAvailable() {
		return fmt.Errorf("sudo is required but not available")
	}

	// Update package list first
	fmt.Println("🔄 Updating package list...")
	var updateCmd *exec.Cmd
	if needsSudo {
		updateCmd = exec.Command("sudo", "apt-get", "update", "-qq")
	} else {
		updateCmd = exec.Command("apt-get", "update", "-qq")
	}
	updateCmd.Stdin = os.Stdin
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr
	if err := updateCmd.Run(); err != nil {
		fmt.Printf("⚠️  Warning: apt-get update failed: %v\n", err)
		// Continue anyway - packages might still install
	}

	// Install packages
	args := []string{"apt-get", "install", "-y"}
	args = append(args, packages...)
	if needsSudo {
		args = append([]string{"sudo"}, args...)
	}
	fmt.Printf("🚀 Running: %s\n", strings.Join(args, " "))

	var installCmd *exec.Cmd
	if needsSudo {
		installArgs := append([]string{"apt-get", "install", "-y"}, packages...)
		installCmd = exec.Command("sudo", installArgs...)
	} else {
		installArgs := append([]string{"install", "-y"}, packages...)
		installCmd = exec.Command("apt-get", installArgs...)
	}
	installCmd.Stdin = os.Stdin
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("apt-get install failed: %w", err)
	}

	fmt.Println("✅ APT installation completed")
	return nil
}

// InstallViaDNF installs packages using DNF/YUM package manager
func InstallViaDNF(packages []string, requiresSudo bool) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("DNF/YUM is only available on Linux")
	}

	fmt.Printf("📦 Installing via DNF: %v\n", packages)

	// Determine which package manager to use
	packageManager := "dnf"
	if !isCommandAvailable("dnf") {
		if isCommandAvailable("yum") {
			packageManager = "yum"
		} else {
			return fmt.Errorf("neither DNF nor YUM package manager found")
		}
	}

	// Build command
	sudoPrefix := ""
	if requiresSudo && !IsRunningAsRoot() {
		if !IsSudoAvailable() {
			return fmt.Errorf("sudo is required but not available")
		}
		sudoPrefix = "sudo "
	}

	// Install packages
	installCmd := sudoPrefix + packageManager + " install -y " + strings.Join(packages, " ")
	fmt.Printf("🚀 Running: %s\n", installCmd)

	if err := runCommand(installCmd); err != nil {
		return fmt.Errorf("%s install failed: %w", packageManager, err)
	}

	fmt.Printf("✅ %s installation completed\n", strings.ToUpper(packageManager))
	return nil
}

// InstallViaSnap installs packages using Snap package manager
func InstallViaSnap(packages []string, classic bool) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("Snap is only available on Linux")
	}

	if !isCommandAvailable("snap") {
		return fmt.Errorf("snap command not found - please install snapd")
	}

	fmt.Printf("📦 Installing via Snap: %v\n", packages)

	// Build command
	sudoPrefix := ""
	if !IsRunningAsRoot() {
		if !IsSudoAvailable() {
			return fmt.Errorf("sudo is required for snap installation")
		}
		sudoPrefix = "sudo "
	}

	// Install each package (snap doesn't support multiple packages in one command)
	for _, pkg := range packages {
		classicFlag := ""
		if classic {
			classicFlag = " --classic"
		}

		installCmd := sudoPrefix + "snap install " + pkg + classicFlag
		fmt.Printf("🚀 Running: %s\n", installCmd)

		if err := runCommand(installCmd); err != nil {
			return fmt.Errorf("snap install failed for %s: %w", pkg, err)
		}
	}

	fmt.Println("✅ Snap installation completed")
	return nil
}

// InstallViaPacman installs packages using Pacman package manager (Arch Linux)
func InstallViaPacman(packages []string, requiresSudo bool) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("Pacman is only available on Linux")
	}

	if !isCommandAvailable("pacman") {
		return fmt.Errorf("pacman command not found")
	}

	fmt.Printf("📦 Installing via Pacman: %v\n", packages)

	// Build command
	sudoPrefix := ""
	if requiresSudo && !IsRunningAsRoot() {
		if !IsSudoAvailable() {
			return fmt.Errorf("sudo is required but not available")
		}
		sudoPrefix = "sudo "
	}

	// Sync package databases first
	fmt.Println("🔄 Syncing package databases...")
	syncCmd := sudoPrefix + "pacman -Sy --noconfirm"
	if err := runCommand(syncCmd); err != nil {
		fmt.Printf("⚠️  Warning: pacman -Sy failed: %v\n", err)
	}

	// Install packages
	installCmd := sudoPrefix + "pacman -S --noconfirm " + strings.Join(packages, " ")
	fmt.Printf("🚀 Running: %s\n", installCmd)

	if err := runCommand(installCmd); err != nil {
		return fmt.Errorf("pacman install failed: %w", err)
	}

	fmt.Println("✅ Pacman installation completed")
	return nil
}

// InstallDebPackage installs a .deb package file
func InstallDebPackage(debFile string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("DEB packages are only supported on Linux")
	}

	if !isCommandAvailable("dpkg") {
		return fmt.Errorf("dpkg command not found")
	}

	fmt.Printf("📦 Installing DEB package: %s\n", debFile)

	// Build command
	sudoPrefix := ""
	if !IsRunningAsRoot() {
		if !IsSudoAvailable() {
			return fmt.Errorf("sudo is required for deb installation")
		}
		sudoPrefix = "sudo "
	}

	// Install package
	installCmd := sudoPrefix + "dpkg -i " + debFile
	fmt.Printf("🚀 Running: %s\n", installCmd)

	if err := runCommand(installCmd); err != nil {
		// Try to fix dependencies
		fmt.Println("⚠️  Fixing dependencies...")
		fixCmd := sudoPrefix + "apt-get install -f -y"
		if fixErr := runCommand(fixCmd); fixErr != nil {
			return fmt.Errorf("dpkg install failed and dependency fix failed: %w", err)
		}
	}

	fmt.Println("✅ DEB package installation completed")
	return nil
}

// runCommand executes a shell command
func runCommand(cmdStr string) error {
	// Split command for exec
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = nil // Could be set to os.Stdout for verbose output
	cmd.Stderr = nil // Could be set to os.Stderr for verbose output

	return cmd.Run()
}

// isCommandAvailable checks if a command is available in PATH
func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// DetectPackageManager detects the available package manager on the system
func DetectPackageManager() string {
	if runtime.GOOS != "linux" {
		return ""
	}

	// Check in order of preference
	managers := []string{"apt-get", "dnf", "yum", "pacman", "zypper"}
	for _, mgr := range managers {
		if isCommandAvailable(mgr) {
			return mgr
		}
	}

	return ""
}

// InstallViaChocolatey installs packages using Chocolatey package manager (Windows)
func InstallViaChocolatey(packages []string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("Chocolatey is only available on Windows")
	}

	if !isCommandAvailable("choco") {
		return fmt.Errorf("Chocolatey is not installed. Install it first with: portunix install chocolatey")
	}

	fmt.Printf("📦 Installing via Chocolatey: %v\n", packages)

	args := []string{"install", "-y"}
	args = append(args, packages...)

	cmd := exec.Command("choco", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("choco install failed: %w", err)
	}

	fmt.Println("✅ Chocolatey installation completed")
	return nil
}

// InstallViaWinget installs packages using Windows Package Manager (winget)
func InstallViaWinget(packages []string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("Winget is only available on Windows")
	}

	if !isCommandAvailable("winget") {
		return fmt.Errorf("Winget is not installed")
	}

	fmt.Printf("📦 Installing via Winget: %v\n", packages)

	for _, pkg := range packages {
		cmd := exec.Command("winget", "install", "--id", pkg, "-e", "--accept-package-agreements", "--accept-source-agreements")
		cmd.Stdout = nil
		cmd.Stderr = nil

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("winget install failed for %s: %w", pkg, err)
		}
	}

	fmt.Println("✅ Winget installation completed")
	return nil
}
