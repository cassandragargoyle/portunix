package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// VenvInfo holds information about a virtual environment
type VenvInfo struct {
	Name          string            `json:"name"`
	Path          string            `json:"path"`
	PythonVersion string            `json:"python_version"`
	PackageCount  int               `json:"package_count"`
	Size          int64             `json:"size_bytes"`
	SizeHuman     string            `json:"size_human"`
	Created       string            `json:"created,omitempty"`
	Exists        bool              `json:"exists"`
	IsLocal       bool              `json:"is_local"`
	Components    map[string]string `json:"components,omitempty"` // pip, setuptools, wheel versions
}

// VenvManager handles virtual environment operations
type VenvManager struct {
	venvBaseDir string
}

// VenvTarget represents the resolved target for venv operations
type VenvTarget struct {
	Path    string // Full path to the venv
	IsLocal bool   // True if project-local (./.venv or --path)
	Name    string // Name (for centralized) or empty for local
}

// NewVenvManager creates a new virtual environment manager
func NewVenvManager() (*VenvManager, error) {
	// Default venv location: ~/.portunix/python/venvs/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	venvBaseDir := filepath.Join(homeDir, ".portunix", "python", "venvs")

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(venvBaseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create venv base directory: %v", err)
	}

	return &VenvManager{
		venvBaseDir: venvBaseDir,
	}, nil
}

// CreateVenv creates a new virtual environment
func (vm *VenvManager) CreateVenv(name string, pythonVersion string) error {
	venvPath := filepath.Join(vm.venvBaseDir, name)

	// Check if venv already exists
	if _, err := os.Stat(venvPath); err == nil {
		return fmt.Errorf("virtual environment '%s' already exists", name)
	}

	// Determine Python executable
	pythonCmd := "python3"
	if runtime.GOOS == "windows" {
		pythonCmd = "python"
	}

	// If specific version requested, try to use it
	if pythonVersion != "" {
		pythonCmd = "python" + pythonVersion
	}

	// Create venv using python -m venv
	cmd := exec.Command(pythonCmd, "-m", "venv", venvPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create venv: %v\nOutput: %s", err, string(output))
	}

	fmt.Printf("✅ Virtual environment '%s' created at: %s\n", name, venvPath)
	return nil
}

// ListVenvs lists all virtual environments
func (vm *VenvManager) ListVenvs() ([]*VenvInfo, error) {
	entries, err := os.ReadDir(vm.venvBaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*VenvInfo{}, nil // No venvs directory yet
		}
		return nil, err
	}

	var venvs []*VenvInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		info, err := vm.GetVenvInfo(entry.Name())
		if err != nil {
			// Skip if not a valid venv
			continue
		}

		venvs = append(venvs, info)
	}

	return venvs, nil
}

// VenvExists checks if a virtual environment exists
func (vm *VenvManager) VenvExists(name string) bool {
	venvPath := filepath.Join(vm.venvBaseDir, name)

	// Check if directory exists
	info, err := os.Stat(venvPath)
	if err != nil {
		return false
	}

	// Check if it's a directory
	if !info.IsDir() {
		return false
	}

	// Check for Python executable in venv
	pythonExe := vm.getPythonExecutable(venvPath)
	if _, err := os.Stat(pythonExe); err != nil {
		return false
	}

	return true
}

// GetVenvInfo gets detailed information about a virtual environment
func (vm *VenvManager) GetVenvInfo(name string) (*VenvInfo, error) {
	// Check if venv exists
	if !vm.VenvExists(name) {
		return nil, fmt.Errorf("virtual environment '%s' does not exist", name)
	}

	venvPath := filepath.Join(vm.venvBaseDir, name)
	info := &VenvInfo{
		Name:   name,
		Path:   venvPath,
		Exists: true,
	}

	// Get Python version
	pythonVersion, err := vm.getPythonVersion(venvPath)
	if err == nil {
		info.PythonVersion = pythonVersion
	}

	// Get package count
	packageCount, err := vm.getPackageCount(venvPath)
	if err == nil {
		info.PackageCount = packageCount
	}

	// Get directory size
	size, err := vm.getDirSize(venvPath)
	if err == nil {
		info.Size = size
	}

	return info, nil
}

// DeleteVenv removes a virtual environment
func (vm *VenvManager) DeleteVenv(name string) error {
	venvPath := filepath.Join(vm.venvBaseDir, name)

	// Check if venv exists
	if !vm.VenvExists(name) {
		return fmt.Errorf("virtual environment '%s' does not exist", name)
	}

	// Remove directory
	if err := os.RemoveAll(venvPath); err != nil {
		return fmt.Errorf("failed to delete venv: %v", err)
	}

	fmt.Printf("✅ Virtual environment '%s' deleted\n", name)
	return nil
}

// Helper functions
func (vm *VenvManager) getPythonExecutable(venvPath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvPath, "Scripts", "python.exe")
	}
	return filepath.Join(venvPath, "bin", "python")
}

func (vm *VenvManager) getPipExecutable(venvPath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvPath, "Scripts", "pip.exe")
	}
	return filepath.Join(venvPath, "bin", "pip")
}

func (vm *VenvManager) getPythonVersion(venvPath string) (string, error) {
	pythonExe := vm.getPythonExecutable(venvPath)
	cmd := exec.Command(pythonExe, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse "Python 3.11.5" -> "3.11.5"
	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "Python ")
	return version, nil
}

func (vm *VenvManager) getPackageCount(venvPath string) (int, error) {
	// Use python -m pip pattern for reliable pip operations
	pythonExe := vm.getPythonExecutable(venvPath)
	cmd := exec.Command(pythonExe, "-m", "pip", "list", "--format=freeze")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}
	return len(lines), nil
}

func (vm *VenvManager) getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// InstallPackage installs a package in a virtual environment
// Uses python -m pip pattern for reliable pip operations including self-upgrade
func (vm *VenvManager) InstallPackage(venvName string, packageName string) error {
	venvPath := filepath.Join(vm.venvBaseDir, venvName)

	if !vm.VenvExists(venvName) {
		return fmt.Errorf("virtual environment '%s' does not exist", venvName)
	}

	pythonExe := vm.getPythonExecutable(venvPath)
	cmd := exec.Command(pythonExe, "-m", "pip", "install", packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ListPackages lists installed packages in a virtual environment
// Uses python -m pip pattern for reliable pip operations
func (vm *VenvManager) ListPackages(venvName string) error {
	venvPath := filepath.Join(vm.venvBaseDir, venvName)

	if !vm.VenvExists(venvName) {
		return fmt.Errorf("virtual environment '%s' does not exist", venvName)
	}

	pythonExe := vm.getPythonExecutable(venvPath)
	cmd := exec.Command(pythonExe, "-m", "pip", "list")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// InstallRequirements installs packages from requirements.txt file
func (vm *VenvManager) InstallRequirements(venvName string, requirementsPath string) error {
	venvPath := filepath.Join(vm.venvBaseDir, venvName)

	if !vm.VenvExists(venvName) {
		return fmt.Errorf("virtual environment '%s' does not exist", venvName)
	}

	// Check if requirements file exists
	if _, err := os.Stat(requirementsPath); os.IsNotExist(err) {
		return fmt.Errorf("requirements file not found: %s", requirementsPath)
	}

	// Use python -m pip pattern for reliable pip operations
	pythonExe := vm.getPythonExecutable(venvPath)
	cmd := exec.Command(pythonExe, "-m", "pip", "install", "-r", requirementsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ResolveVenvPath resolves the venv path based on flags precedence:
// 1. --path <explicit> → Use explicit path
// 2. --local → Use ./.venv
// 3. --venv <name> → Use ~/.portunix/python/venvs/<name>
// 4. Auto-detect → If ./.venv exists, use it (for pip commands only)
// 5. No target → Error
func (vm *VenvManager) ResolveVenvPath(localFlag bool, pathFlag string, venvName string, autoDetect bool) (*VenvTarget, error) {
	// Priority 1: Explicit path
	if pathFlag != "" {
		absPath, err := filepath.Abs(pathFlag)
		if err != nil {
			return nil, fmt.Errorf("invalid path: %v", err)
		}
		return &VenvTarget{
			Path:    absPath,
			IsLocal: true,
			Name:    "",
		}, nil
	}

	// Priority 2: --local flag
	if localFlag {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %v", err)
		}
		return &VenvTarget{
			Path:    filepath.Join(cwd, ".venv"),
			IsLocal: true,
			Name:    "",
		}, nil
	}

	// Priority 3: --venv <name>
	if venvName != "" {
		return &VenvTarget{
			Path:    filepath.Join(vm.venvBaseDir, venvName),
			IsLocal: false,
			Name:    venvName,
		}, nil
	}

	// Priority 4: Auto-detect ./.venv (only for pip commands)
	if autoDetect {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %v", err)
		}
		localVenvPath := filepath.Join(cwd, ".venv")
		if vm.VenvExistsAtPath(localVenvPath) {
			return &VenvTarget{
				Path:    localVenvPath,
				IsLocal: true,
				Name:    "",
			}, nil
		}
	}

	// Priority 5: No target specified
	return nil, fmt.Errorf("no virtual environment specified. Use --local, --path, or --venv flag")
}

// VenvExistsAtPath checks if a valid venv exists at the given path
func (vm *VenvManager) VenvExistsAtPath(venvPath string) bool {
	// Check if directory exists
	info, err := os.Stat(venvPath)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	// Check for Python executable in venv
	pythonExe := vm.getPythonExecutable(venvPath)
	if _, err := os.Stat(pythonExe); err != nil {
		return false
	}

	return true
}

// DetectRequirementsFile looks for requirements.txt or pyproject.toml
func (vm *VenvManager) DetectRequirementsFile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %v", err)
	}

	// Check for requirements.txt first
	reqPath := filepath.Join(cwd, "requirements.txt")
	if _, err := os.Stat(reqPath); err == nil {
		return reqPath, nil
	}

	// Check for pyproject.toml
	pyprojectPath := filepath.Join(cwd, "pyproject.toml")
	if _, err := os.Stat(pyprojectPath); err == nil {
		return pyprojectPath, nil
	}

	return "", fmt.Errorf("no requirements.txt or pyproject.toml found in current directory")
}

// CreateLocalVenv creates a project-local virtual environment at ./.venv or custom path
func (vm *VenvManager) CreateLocalVenv(venvPath string, force bool, pythonVersion string) error {
	// Check if venv already exists
	if _, err := os.Stat(venvPath); err == nil {
		if !force {
			return fmt.Errorf("virtual environment already exists at '%s'. Use --force to recreate", venvPath)
		}
		// Remove existing venv
		fmt.Printf("Removing existing venv at %s...\n", venvPath)
		if err := os.RemoveAll(venvPath); err != nil {
			return fmt.Errorf("failed to remove existing venv: %v", err)
		}
	}

	// Determine Python executable
	pythonCmd := vm.findPythonExecutable(pythonVersion)

	// Create venv using python -m venv
	fmt.Printf("Creating virtual environment at %s...\n", venvPath)
	cmd := exec.Command(pythonCmd, "-m", "venv", venvPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create venv: %v\nOutput: %s", err, string(output))
	}

	fmt.Printf("✅ Virtual environment created at: %s\n", venvPath)
	return nil
}

// UpgradePip upgrades pip in the specified venv using python -m pip
func (vm *VenvManager) UpgradePip(venvPath string) error {
	pythonExe := vm.getPythonExecutable(venvPath)

	fmt.Println("Upgrading pip...")
	cmd := exec.Command(pythonExe, "-m", "pip", "install", "--upgrade", "pip")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to upgrade pip: %v", err)
	}

	fmt.Println("✅ pip upgraded successfully")
	return nil
}

// InstallRequirementsAtPath installs packages from requirements file to venv at path
func (vm *VenvManager) InstallRequirementsAtPath(venvPath string, requirementsPath string) error {
	// Check if venv exists
	if !vm.VenvExistsAtPath(venvPath) {
		return fmt.Errorf("virtual environment does not exist at '%s'", venvPath)
	}

	// Check if requirements file exists
	if _, err := os.Stat(requirementsPath); os.IsNotExist(err) {
		return fmt.Errorf("requirements file not found: %s", requirementsPath)
	}

	// Determine install method based on file type
	pythonExe := vm.getPythonExecutable(venvPath)

	if strings.HasSuffix(requirementsPath, "pyproject.toml") {
		// Install using pip install . for pyproject.toml
		fmt.Println("Installing from pyproject.toml...")
		projectDir := filepath.Dir(requirementsPath)
		cmd := exec.Command(pythonExe, "-m", "pip", "install", "-e", projectDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Install using pip install -r for requirements.txt
	fmt.Printf("Installing from %s...\n", filepath.Base(requirementsPath))
	cmd := exec.Command(pythonExe, "-m", "pip", "install", "-r", requirementsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// InstallPackageAtPath installs a package to venv at specified path
func (vm *VenvManager) InstallPackageAtPath(venvPath string, packageName string) error {
	if !vm.VenvExistsAtPath(venvPath) {
		return fmt.Errorf("virtual environment does not exist at '%s'", venvPath)
	}

	pythonExe := vm.getPythonExecutable(venvPath)
	cmd := exec.Command(pythonExe, "-m", "pip", "install", packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ListPackagesAtPath lists installed packages in venv at specified path
func (vm *VenvManager) ListPackagesAtPath(venvPath string) error {
	if !vm.VenvExistsAtPath(venvPath) {
		return fmt.Errorf("virtual environment does not exist at '%s'", venvPath)
	}

	pythonExe := vm.getPythonExecutable(venvPath)
	cmd := exec.Command(pythonExe, "-m", "pip", "list")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// FreezePackagesAtPath outputs pip freeze for venv at specified path
func (vm *VenvManager) FreezePackagesAtPath(venvPath string) error {
	if !vm.VenvExistsAtPath(venvPath) {
		return fmt.Errorf("virtual environment does not exist at '%s'", venvPath)
	}

	pythonExe := vm.getPythonExecutable(venvPath)
	cmd := exec.Command(pythonExe, "-m", "pip", "freeze")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetVenvInfoAtPath gets detailed information about a venv at specified path
func (vm *VenvManager) GetVenvInfoAtPath(venvPath string) (*VenvInfo, error) {
	if !vm.VenvExistsAtPath(venvPath) {
		return nil, fmt.Errorf("virtual environment does not exist at '%s'", venvPath)
	}

	info := &VenvInfo{
		Name:    filepath.Base(venvPath),
		Path:    venvPath,
		Exists:  true,
		IsLocal: true,
	}

	// Get Python version
	pythonVersion, err := vm.getPythonVersionAtPath(venvPath)
	if err == nil {
		info.PythonVersion = pythonVersion
	}

	// Get package count
	packageCount, err := vm.getPackageCountAtPath(venvPath)
	if err == nil {
		info.PackageCount = packageCount
	}

	// Get directory size
	size, err := vm.getDirSize(venvPath)
	if err == nil {
		info.Size = size
		info.SizeHuman = formatSizeBytes(size)
	}

	return info, nil
}

// GetVenvInfoAtPathVerbose gets detailed info including component versions
func (vm *VenvManager) GetVenvInfoAtPathVerbose(venvPath string) (*VenvInfo, error) {
	info, err := vm.GetVenvInfoAtPath(venvPath)
	if err != nil {
		return nil, err
	}

	// Get component versions
	info.Components = vm.getComponentVersions(venvPath)

	return info, nil
}

// getComponentVersions gets versions of key Python components (pip, setuptools, wheel)
func (vm *VenvManager) getComponentVersions(venvPath string) map[string]string {
	components := make(map[string]string)
	pythonExe := vm.getPythonExecutable(venvPath)

	// Components to check
	pkgs := []string{"pip", "setuptools", "wheel"}

	for _, pkg := range pkgs {
		cmd := exec.Command(pythonExe, "-m", "pip", "show", pkg)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		// Parse version from output
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "Version:") {
				version := strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
				components[pkg] = version
				break
			}
		}
	}

	return components
}

// formatSizeBytes formats bytes to human readable string
func formatSizeBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// DeleteVenvAtPath removes a venv at specified path
func (vm *VenvManager) DeleteVenvAtPath(venvPath string) error {
	if !vm.VenvExistsAtPath(venvPath) {
		return fmt.Errorf("virtual environment does not exist at '%s'", venvPath)
	}

	if err := os.RemoveAll(venvPath); err != nil {
		return fmt.Errorf("failed to delete venv: %v", err)
	}

	fmt.Printf("✅ Virtual environment deleted at '%s'\n", venvPath)
	return nil
}

// findPythonExecutable finds the appropriate Python executable
func (vm *VenvManager) findPythonExecutable(pythonVersion string) string {
	// If specific version requested
	if pythonVersion != "" {
		versionedCmd := "python" + pythonVersion
		if _, err := exec.LookPath(versionedCmd); err == nil {
			return versionedCmd
		}
		// Try with dot notation (e.g., python3.11)
		if !strings.Contains(pythonVersion, ".") {
			versionedCmd = "python" + pythonVersion[0:1] + "." + pythonVersion[1:]
			if _, err := exec.LookPath(versionedCmd); err == nil {
				return versionedCmd
			}
		}
	}

	// Default: python3 on Unix, python on Windows
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "python3"
}

// getPythonVersionAtPath gets Python version from venv at specified path
func (vm *VenvManager) getPythonVersionAtPath(venvPath string) (string, error) {
	pythonExe := vm.getPythonExecutable(venvPath)
	cmd := exec.Command(pythonExe, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "Python ")
	return version, nil
}

// getPackageCountAtPath gets package count using python -m pip
func (vm *VenvManager) getPackageCountAtPath(venvPath string) (int, error) {
	pythonExe := vm.getPythonExecutable(venvPath)
	cmd := exec.Command(pythonExe, "-m", "pip", "list", "--format=freeze")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}
	return len(lines), nil
}

// GetActivationCommand returns the shell command to activate the venv
func (vm *VenvManager) GetActivationCommand(venvPath string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\Scripts\\activate", venvPath)
	}
	return fmt.Sprintf("source %s/bin/activate", venvPath)
}
