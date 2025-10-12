package pip

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// PipManager handles pip package management operations
type PipManager struct {
	Debug      bool
	DryRun     bool
	Timeout    time.Duration
	PipCommand string // pip, pip3, python -m pip, etc.
}

// NewPipManager creates a new pip manager instance
func NewPipManager() *PipManager {
	return &PipManager{
		Debug:      false,
		DryRun:     false,
		Timeout:    10 * time.Minute,
		PipCommand: "pip", // default, will be detected
	}
}

// PackageInfo represents information about a pip package
type PackageInfo struct {
	Name        string
	Version     string
	Description string
	Installed   bool
	Available   bool
}

// IsSupported checks if pip is supported on the current system
func (p *PipManager) IsSupported() bool {
	// Detect best pip command to use
	pipCommands := []string{"pip", "pip3", "python -m pip", "python3 -m pip"}

	for _, cmd := range pipCommands {
		if err := p.testPipCommand(cmd); err == nil {
			p.PipCommand = cmd
			return true
		}
	}

	return false
}

// testPipCommand tests if a pip command is available
func (p *PipManager) testPipCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("invalid command")
	}

	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.Command(parts[0], "--version")
	} else {
		cmd = exec.Command(parts[0], append(parts[1:], "--version")...)
	}

	return cmd.Run()
}

// Install installs packages using pip
func (p *PipManager) Install(packages []string) error {
	if len(packages) == 0 {
		return fmt.Errorf("no packages to install")
	}

	for _, pkg := range packages {
		if err := p.installPackage(pkg); err != nil {
			return fmt.Errorf("failed to install package '%s': %w", pkg, err)
		}
	}

	return nil
}

// installPackage installs a single package
func (p *PipManager) installPackage(pkg string) error {
	if p.Debug {
		fmt.Printf("📦 Installing pip package: %s\n", pkg)
	}

	if p.DryRun {
		fmt.Printf("🔄 [DRY-RUN] Would install pip package: %s\n", pkg)
		return nil
	}

	// Build pip install command
	parts := strings.Fields(p.PipCommand)
	args := append(parts[1:], "install", pkg)

	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.Command(parts[0], "install", pkg)
	} else {
		cmd = exec.Command(parts[0], args...)
	}

	if p.Debug {
		fmt.Printf("🔧 Executing: %s %s\n", cmd.Path, strings.Join(cmd.Args[1:], " "))
	}

	// Set up output handling
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start pip install: %w", err)
	}

	// Read output in real-time
	go p.readOutput(stdout, "STDOUT")
	go p.readOutput(stderr, "STDERR")

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("pip install failed: %w", err)
	}

	fmt.Printf("✅ Successfully installed pip package: %s\n", pkg)
	return nil
}

// readOutput reads and displays command output
func (p *PipManager) readOutput(pipe io.ReadCloser, prefix string) {
	defer pipe.Close()
	scanner := bufio.NewScanner(pipe)

	for scanner.Scan() {
		line := scanner.Text()
		if p.Debug {
			fmt.Printf("[%s] %s\n", prefix, line)
		} else {
			// Show important lines even in non-debug mode
			if strings.Contains(strings.ToLower(line), "error") ||
			   strings.Contains(strings.ToLower(line), "failed") ||
			   strings.Contains(strings.ToLower(line), "successfully installed") {
				fmt.Printf("   %s\n", line)
			}
		}
	}
}

// IsInstalled checks if a package is installed
func (p *PipManager) IsInstalled(packageName string) bool {
	parts := strings.Fields(p.PipCommand)
	args := append(parts[1:], "show", packageName)

	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.Command(parts[0], "show", packageName)
	} else {
		cmd = exec.Command(parts[0], args...)
	}

	return cmd.Run() == nil
}

// GetInstalledPackages returns list of installed packages
func (p *PipManager) GetInstalledPackages() ([]PackageInfo, error) {
	parts := strings.Fields(p.PipCommand)
	args := append(parts[1:], "list", "--format=freeze")

	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.Command(parts[0], "list", "--format=freeze")
	} else {
		cmd = exec.Command(parts[0], args...)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get pip package list: %w", err)
	}

	var packages []PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse package==version format
		parts := strings.Split(line, "==")
		if len(parts) >= 2 {
			packages = append(packages, PackageInfo{
				Name:      parts[0],
				Version:   parts[1],
				Installed: true,
				Available: true,
			})
		}
	}

	return packages, nil
}

// Upgrade upgrades a package or all packages
func (p *PipManager) Upgrade(packageName string) error {
	if p.DryRun {
		fmt.Printf("🔄 [DRY-RUN] Would upgrade pip package: %s\n", packageName)
		return nil
	}

	parts := strings.Fields(p.PipCommand)
	var args []string

	if packageName == "" {
		// Upgrade all packages
		args = append(parts[1:], "install", "--upgrade", "pip")
	} else {
		args = append(parts[1:], "install", "--upgrade", packageName)
	}

	var cmd *exec.Cmd
	if len(parts) == 1 {
		if packageName == "" {
			cmd = exec.Command(parts[0], "install", "--upgrade", "pip")
		} else {
			cmd = exec.Command(parts[0], "install", "--upgrade", packageName)
		}
	} else {
		cmd = exec.Command(parts[0], args...)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to upgrade pip package '%s': %w", packageName, err)
	}

	fmt.Printf("✅ Successfully upgraded pip package: %s\n", packageName)
	return nil
}