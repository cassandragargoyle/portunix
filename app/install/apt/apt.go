package apt

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// AptManager handles APT package management operations
type AptManager struct {
	Debug   bool
	DryRun  bool
	Timeout time.Duration
}

// NewAptManager creates a new APT manager instance
func NewAptManager() *AptManager {
	return &AptManager{
		Debug:   false,
		DryRun:  false,
		Timeout: 10 * time.Minute,
	}
}

// Repository represents an APT repository
type Repository struct {
	URI          string
	Distribution string
	Components   []string
	GPGKey       string
	GPGKeyURL    string
}

// PackageInfo represents information about an APT package
type PackageInfo struct {
	Name        string
	Version     string
	Description string
	Installed   bool
	Available   bool
}

// IsSupported checks if APT is supported on the current system
func (apt *AptManager) IsSupported() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	
	// Check if apt-get command exists
	_, err := exec.LookPath("apt-get")
	return err == nil
}

// Update updates the package list
func (apt *AptManager) Update() error {
	if !apt.IsSupported() {
		return fmt.Errorf("APT is not supported on this system")
	}
	
	fmt.Println("Updating package list...")
	
	if apt.DryRun {
		fmt.Println("[DRY RUN] Would run: sudo apt-get update")
		return nil
	}
	
	cmd := exec.Command("sudo", "apt-get", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// Install installs one or more packages
func (apt *AptManager) Install(packages []string) error {
	if !apt.IsSupported() {
		return fmt.Errorf("APT is not supported on this system")
	}
	
	if len(packages) == 0 {
		return fmt.Errorf("no packages specified")
	}
	
	fmt.Printf("Installing packages: %s\n", strings.Join(packages, ", "))
	
	if apt.DryRun {
		fmt.Printf("[DRY RUN] Would run: sudo apt-get install -y %s\n", strings.Join(packages, " "))
		return nil
	}
	
	args := []string{"apt-get", "install", "-y"}
	args = append(args, packages...)
	
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// Remove removes one or more packages
func (apt *AptManager) Remove(packages []string) error {
	if !apt.IsSupported() {
		return fmt.Errorf("APT is not supported on this system")
	}
	
	if len(packages) == 0 {
		return fmt.Errorf("no packages specified")
	}
	
	fmt.Printf("Removing packages: %s\n", strings.Join(packages, ", "))
	
	if apt.DryRun {
		fmt.Printf("[DRY RUN] Would run: sudo apt-get remove -y %s\n", strings.Join(packages, " "))
		return nil
	}
	
	args := []string{"apt-get", "remove", "-y"}
	args = append(args, packages...)
	
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// Purge completely removes packages including configuration files
func (apt *AptManager) Purge(packages []string) error {
	if !apt.IsSupported() {
		return fmt.Errorf("APT is not supported on this system")
	}
	
	if len(packages) == 0 {
		return fmt.Errorf("no packages specified")
	}
	
	fmt.Printf("Purging packages: %s\n", strings.Join(packages, ", "))
	
	if apt.DryRun {
		fmt.Printf("[DRY RUN] Would run: sudo apt-get purge -y %s\n", strings.Join(packages, " "))
		return nil
	}
	
	args := []string{"apt-get", "purge", "-y"}
	args = append(args, packages...)
	
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// Search searches for packages matching a pattern
func (apt *AptManager) Search(pattern string) ([]PackageInfo, error) {
	if !apt.IsSupported() {
		return nil, fmt.Errorf("APT is not supported on this system")
	}
	
	cmd := exec.Command("apt-cache", "search", pattern)
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
		
		parts := strings.SplitN(line, " - ", 2)
		if len(parts) != 2 {
			continue
		}
		
		pkg := PackageInfo{
			Name:        parts[0],
			Description: parts[1],
			Available:   true,
		}
		
		// Check if package is installed
		pkg.Installed = apt.IsInstalled(pkg.Name)
		
		packages = append(packages, pkg)
	}
	
	return packages, nil
}

// IsInstalled checks if a package is installed
func (apt *AptManager) IsInstalled(packageName string) bool {
	if !apt.IsSupported() {
		return false
	}
	
	cmd := exec.Command("dpkg", "-l", packageName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	// Check if package is installed (starts with "ii")
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ii") && strings.Contains(line, packageName) {
			return true
		}
	}
	
	return false
}

// GetPackageInfo gets detailed information about a package
func (apt *AptManager) GetPackageInfo(packageName string) (*PackageInfo, error) {
	if !apt.IsSupported() {
		return nil, fmt.Errorf("APT is not supported on this system")
	}
	
	cmd := exec.Command("apt-cache", "show", packageName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get package info: %w", err)
	}
	
	pkg := &PackageInfo{
		Name:      packageName,
		Available: true,
		Installed: apt.IsInstalled(packageName),
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Version:") {
			pkg.Version = strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
		} else if strings.HasPrefix(line, "Description:") {
			pkg.Description = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
		}
	}
	
	return pkg, nil
}

// AddRepository adds a new APT repository
func (apt *AptManager) AddRepository(repo Repository) error {
	if !apt.IsSupported() {
		return fmt.Errorf("APT is not supported on this system")
	}
	
	// Add GPG key if provided
	if repo.GPGKeyURL != "" {
		fmt.Printf("Adding GPG key from: %s\n", repo.GPGKeyURL)
		
		if apt.DryRun {
			fmt.Printf("[DRY RUN] Would add GPG key from: %s\n", repo.GPGKeyURL)
		} else {
			// Download and add GPG key
			cmd := exec.Command("wget", "-qO-", repo.GPGKeyURL)
			gpgCmd := exec.Command("sudo", "apt-key", "add", "-")
			
			pipe, err := cmd.StdoutPipe()
			if err != nil {
				return fmt.Errorf("failed to create pipe: %w", err)
			}
			
			gpgCmd.Stdin = pipe
			gpgCmd.Stdout = os.Stdout
			gpgCmd.Stderr = os.Stderr
			
			if err := cmd.Start(); err != nil {
				return fmt.Errorf("failed to start wget: %w", err)
			}
			
			if err := gpgCmd.Start(); err != nil {
				return fmt.Errorf("failed to start apt-key: %w", err)
			}
			
			if err := cmd.Wait(); err != nil {
				return fmt.Errorf("wget failed: %w", err)
			}
			
			if err := gpgCmd.Wait(); err != nil {
				return fmt.Errorf("apt-key failed: %w", err)
			}
		}
	} else if repo.GPGKey != "" {
		fmt.Println("Adding GPG key...")
		
		if apt.DryRun {
			fmt.Printf("[DRY RUN] Would add GPG key: %s\n", repo.GPGKey)
		} else {
			cmd := exec.Command("sudo", "apt-key", "adv", "--keyserver", "keyserver.ubuntu.com", "--recv-keys", repo.GPGKey)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to add GPG key: %w", err)
			}
		}
	}
	
	// Add repository to sources.list.d
	repoLine := fmt.Sprintf("deb %s %s %s", repo.URI, repo.Distribution, strings.Join(repo.Components, " "))
	sourcesFile := fmt.Sprintf("/etc/apt/sources.list.d/portunix-%s.list", strings.ReplaceAll(repo.Distribution, "/", "-"))
	
	fmt.Printf("Adding repository: %s\n", repoLine)
	
	if apt.DryRun {
		fmt.Printf("[DRY RUN] Would add to %s: %s\n", sourcesFile, repoLine)
		return nil
	}
	
	// Write repository to sources file
	cmd := exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' > %s", repoLine, sourcesFile))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add repository: %w", err)
	}
	
	fmt.Printf("Repository added to: %s\n", sourcesFile)
	
	// Update package list after adding repository
	return apt.Update()
}

// RemoveRepository removes an APT repository
func (apt *AptManager) RemoveRepository(distribution string) error {
	if !apt.IsSupported() {
		return fmt.Errorf("APT is not supported on this system")
	}
	
	sourcesFile := fmt.Sprintf("/etc/apt/sources.list.d/portunix-%s.list", strings.ReplaceAll(distribution, "/", "-"))
	
	fmt.Printf("Removing repository: %s\n", sourcesFile)
	
	if apt.DryRun {
		fmt.Printf("[DRY RUN] Would remove: %s\n", sourcesFile)
		return nil
	}
	
	cmd := exec.Command("sudo", "rm", "-f", sourcesFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}
	
	fmt.Printf("Repository removed: %s\n", sourcesFile)
	
	// Update package list after removing repository
	return apt.Update()
}

// ListInstalled lists all installed packages
func (apt *AptManager) ListInstalled() ([]PackageInfo, error) {
	if !apt.IsSupported() {
		return nil, fmt.Errorf("APT is not supported on this system")
	}
	
	cmd := exec.Command("dpkg", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	
	var packages []PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "ii") {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		
		pkg := PackageInfo{
			Name:      fields[1],
			Version:   fields[2],
			Installed: true,
			Available: true,
		}
		
		// Join remaining fields as description
		if len(fields) > 3 {
			pkg.Description = strings.Join(fields[3:], " ")
		}
		
		packages = append(packages, pkg)
	}
	
	return packages, scanner.Err()
}

// Upgrade upgrades all packages or specific packages
func (apt *AptManager) Upgrade(packages []string) error {
	if !apt.IsSupported() {
		return fmt.Errorf("APT is not supported on this system")
	}
	
	if len(packages) == 0 {
		fmt.Println("Upgrading all packages...")
		
		if apt.DryRun {
			fmt.Println("[DRY RUN] Would run: sudo apt-get upgrade -y")
			return nil
		}
		
		cmd := exec.Command("sudo", "apt-get", "upgrade", "-y")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	} else {
		fmt.Printf("Upgrading packages: %s\n", strings.Join(packages, ", "))
		
		if apt.DryRun {
			fmt.Printf("[DRY RUN] Would run: sudo apt-get install --only-upgrade -y %s\n", strings.Join(packages, " "))
			return nil
		}
		
		args := []string{"apt-get", "install", "--only-upgrade", "-y"}
		args = append(args, packages...)
		
		cmd := exec.Command("sudo", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

// Clean cleans the APT cache
func (apt *AptManager) Clean() error {
	if !apt.IsSupported() {
		return fmt.Errorf("APT is not supported on this system")
	}
	
	fmt.Println("Cleaning APT cache...")
	
	if apt.DryRun {
		fmt.Println("[DRY RUN] Would run: sudo apt-get clean && sudo apt-get autoremove -y")
		return nil
	}
	
	// Clean cache
	cleanCmd := exec.Command("sudo", "apt-get", "clean")
	cleanCmd.Stdout = os.Stdout
	cleanCmd.Stderr = os.Stderr
	if err := cleanCmd.Run(); err != nil {
		return fmt.Errorf("failed to clean cache: %w", err)
	}
	
	// Remove unnecessary packages
	autoremoveCmd := exec.Command("sudo", "apt-get", "autoremove", "-y")
	autoremoveCmd.Stdout = os.Stdout
	autoremoveCmd.Stderr = os.Stderr
	if err := autoremoveCmd.Run(); err != nil {
		return fmt.Errorf("failed to autoremove: %w", err)
	}
	
	return nil
}