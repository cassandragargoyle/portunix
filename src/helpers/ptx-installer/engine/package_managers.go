package engine

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// InstallViaAPT installs packages using APT package manager
func InstallViaAPT(packages []string, requiresSudo bool) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("APT is only available on Linux")
	}

	fmt.Printf("📦 Installing via APT: %v\n", packages)

	// Build command
	sudoPrefix := ""
	if requiresSudo && !IsRunningAsRoot() {
		if !IsSudoAvailable() {
			return fmt.Errorf("sudo is required but not available")
		}
		sudoPrefix = "sudo "
	}

	// Update package list first
	fmt.Println("🔄 Updating package list...")
	updateCmd := sudoPrefix + "apt-get update -qq"
	if err := runCommand(updateCmd); err != nil {
		fmt.Printf("⚠️  Warning: apt-get update failed: %v\n", err)
		// Continue anyway - packages might still install
	}

	// Install packages
	installCmd := sudoPrefix + "apt-get install -y " + strings.Join(packages, " ")
	fmt.Printf("🚀 Running: %s\n", installCmd)

	if err := runCommand(installCmd); err != nil {
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
