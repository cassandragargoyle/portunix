package chocolatey

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ChocolateyManager handles Chocolatey package management operations
type ChocolateyManager struct {
	Debug   bool
	DryRun  bool
	Timeout time.Duration
}

// NewChocolateyManager creates a new Chocolatey manager instance
func NewChocolateyManager() *ChocolateyManager {
	return &ChocolateyManager{
		Debug:   false,
		DryRun:  false,
		Timeout: 10 * time.Minute,
	}
}

// PackageInfo represents information about a Chocolatey package
type PackageInfo struct {
	Name        string
	Version     string
	Description string
	Installed   bool
	Available   bool
}

// IsSupported checks if Chocolatey is supported on the current system
func (choco *ChocolateyManager) IsSupported() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// Check if choco command exists
	_, err := exec.LookPath("choco")
	return err == nil
}

// IsInstalled checks if Chocolatey itself is installed
func (choco *ChocolateyManager) IsInstalled() bool {
	return choco.IsSupported()
}

// InstallChocolatey installs Chocolatey package manager
func (choco *ChocolateyManager) InstallChocolatey() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("Chocolatey is only supported on Windows")
	}

	if choco.IsInstalled() {
		fmt.Println("Chocolatey is already installed")
		return nil
	}

	fmt.Println("Installing Chocolatey package manager...")

	if choco.DryRun {
		fmt.Println("[DRY RUN] Would install Chocolatey using PowerShell")
		return nil
	}

	// Install Chocolatey using the official installation script
	installScript := `Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))`

	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Chocolatey: %w", err)
	}

	fmt.Println("✓ Chocolatey installed successfully")
	fmt.Println("⚠️  You may need to restart your terminal or refresh environment variables")

	return nil
}

// Install installs one or more packages
func (choco *ChocolateyManager) Install(packages []string) error {
	if !choco.IsSupported() {
		return fmt.Errorf("Chocolatey is not installed on this system")
	}

	if len(packages) == 0 {
		return fmt.Errorf("no packages specified")
	}

	fmt.Printf("Installing packages: %s\n", strings.Join(packages, ", "))

	if choco.DryRun {
		fmt.Printf("[DRY RUN] Would run: choco install -y %s\n", strings.Join(packages, " "))
		return nil
	}

	args := []string{"install", "-y"}
	args = append(args, packages...)

	cmd := exec.Command("choco", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Uninstall removes one or more packages
func (choco *ChocolateyManager) Uninstall(packages []string) error {
	if !choco.IsSupported() {
		return fmt.Errorf("Chocolatey is not installed on this system")
	}

	if len(packages) == 0 {
		return fmt.Errorf("no packages specified")
	}

	fmt.Printf("Uninstalling packages: %s\n", strings.Join(packages, ", "))

	if choco.DryRun {
		fmt.Printf("[DRY RUN] Would run: choco uninstall -y %s\n", strings.Join(packages, " "))
		return nil
	}

	args := []string{"uninstall", "-y"}
	args = append(args, packages...)

	cmd := exec.Command("choco", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Search searches for packages matching a pattern
func (choco *ChocolateyManager) Search(pattern string) ([]PackageInfo, error) {
	if !choco.IsSupported() {
		return nil, fmt.Errorf("Chocolatey is not installed on this system")
	}

	cmd := exec.Command("choco", "search", pattern, "--limit-output")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to search packages: %w", err)
	}

	var packages []PackageInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}

		pkg := PackageInfo{
			Name:      parts[0],
			Version:   parts[1],
			Available: true,
		}

		// Check if package is installed
		pkg.Installed = choco.IsPackageInstalled(pkg.Name)

		packages = append(packages, pkg)
	}

	return packages, nil
}

// IsPackageInstalled checks if a specific package is installed
func (choco *ChocolateyManager) IsPackageInstalled(packageName string) bool {
	if !choco.IsSupported() {
		return false
	}

	cmd := exec.Command("choco", "list", "--local-only", "--limit-output", packageName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), packageName+"|") {
			return true
		}
	}

	return false
}

// GetPackageInfo gets detailed information about a package
func (choco *ChocolateyManager) GetPackageInfo(packageName string) (*PackageInfo, error) {
	if !choco.IsSupported() {
		return nil, fmt.Errorf("Chocolatey is not installed on this system")
	}

	// Get remote package info
	cmd := exec.Command("choco", "info", packageName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get package info: %w", err)
	}

	pkg := &PackageInfo{
		Name:      packageName,
		Available: true,
		Installed: choco.IsPackageInstalled(packageName),
	}

	// Parse output for version and description
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Version:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				pkg.Version = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Description:") || strings.Contains(line, "Summary:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				pkg.Description = strings.TrimSpace(parts[1])
			}
		}
	}

	return pkg, nil
}

// ListInstalled lists all installed packages
func (choco *ChocolateyManager) ListInstalled() ([]PackageInfo, error) {
	if !choco.IsSupported() {
		return nil, fmt.Errorf("Chocolatey is not installed on this system")
	}

	cmd := exec.Command("choco", "list", "--local-only", "--limit-output")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	var packages []PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}

		pkg := PackageInfo{
			Name:      parts[0],
			Version:   parts[1],
			Installed: true,
			Available: true,
		}

		packages = append(packages, pkg)
	}

	return packages, scanner.Err()
}

// Upgrade upgrades all packages or specific packages
func (choco *ChocolateyManager) Upgrade(packages []string) error {
	if !choco.IsSupported() {
		return fmt.Errorf("Chocolatey is not installed on this system")
	}

	if len(packages) == 0 {
		fmt.Println("Upgrading all packages...")

		if choco.DryRun {
			fmt.Println("[DRY RUN] Would run: choco upgrade all -y")
			return nil
		}

		cmd := exec.Command("choco", "upgrade", "all", "-y")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	} else {
		fmt.Printf("Upgrading packages: %s\n", strings.Join(packages, ", "))

		if choco.DryRun {
			fmt.Printf("[DRY RUN] Would run: choco upgrade -y %s\n", strings.Join(packages, " "))
			return nil
		}

		args := []string{"upgrade", "-y"}
		args = append(args, packages...)

		cmd := exec.Command("choco", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

// Pin pins a package to prevent upgrades
func (choco *ChocolateyManager) Pin(packageName string) error {
	if !choco.IsSupported() {
		return fmt.Errorf("Chocolatey is not installed on this system")
	}

	fmt.Printf("Pinning package: %s\n", packageName)

	if choco.DryRun {
		fmt.Printf("[DRY RUN] Would run: choco pin add -n %s\n", packageName)
		return nil
	}

	cmd := exec.Command("choco", "pin", "add", "-n", packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Unpin unpins a package to allow upgrades
func (choco *ChocolateyManager) Unpin(packageName string) error {
	if !choco.IsSupported() {
		return fmt.Errorf("Chocolatey is not installed on this system")
	}

	fmt.Printf("Unpinning package: %s\n", packageName)

	if choco.DryRun {
		fmt.Printf("[DRY RUN] Would run: choco pin remove -n %s\n", packageName)
		return nil
	}

	cmd := exec.Command("choco", "pin", "remove", "-n", packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ListSources lists configured Chocolatey sources
func (choco *ChocolateyManager) ListSources() error {
	if !choco.IsSupported() {
		return fmt.Errorf("Chocolatey is not installed on this system")
	}

	cmd := exec.Command("choco", "source", "list")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// AddSource adds a new Chocolatey source
func (choco *ChocolateyManager) AddSource(name, url string, username, password string) error {
	if !choco.IsSupported() {
		return fmt.Errorf("Chocolatey is not installed on this system")
	}

	fmt.Printf("Adding source: %s (%s)\n", name, url)

	if choco.DryRun {
		fmt.Printf("[DRY RUN] Would add source: %s -> %s\n", name, url)
		return nil
	}

	args := []string{"source", "add", "-n", name, "-s", url}

	if username != "" {
		args = append(args, "-u", username)
	}
	if password != "" {
		args = append(args, "-p", password)
	}

	cmd := exec.Command("choco", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RemoveSource removes a Chocolatey source
func (choco *ChocolateyManager) RemoveSource(name string) error {
	if !choco.IsSupported() {
		return fmt.Errorf("Chocolatey is not installed on this system")
	}

	fmt.Printf("Removing source: %s\n", name)

	if choco.DryRun {
		fmt.Printf("[DRY RUN] Would remove source: %s\n", name)
		return nil
	}

	cmd := exec.Command("choco", "source", "remove", "-n", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Clean cleans Chocolatey cache and temporary files
func (choco *ChocolateyManager) Clean() error {
	if !choco.IsSupported() {
		return fmt.Errorf("Chocolatey is not installed on this system")
	}

	fmt.Println("Cleaning Chocolatey cache...")

	if choco.DryRun {
		fmt.Println("[DRY RUN] Would run: choco cache clear")
		return nil
	}

	// Clear cache
	cmd := exec.Command("choco", "cache", "clear")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GetVersion gets the Chocolatey version
func (choco *ChocolateyManager) GetVersion() (string, error) {
	if !choco.IsSupported() {
		return "", fmt.Errorf("Chocolatey is not installed on this system")
	}

	cmd := exec.Command("choco", "version", "--limit-output")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Chocolatey version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
