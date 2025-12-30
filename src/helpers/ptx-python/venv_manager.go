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
	Name          string
	Path          string
	PythonVersion string
	PackageCount  int
	Size          int64
	Created       string
	Exists        bool
}

// VenvManager handles virtual environment operations
type VenvManager struct {
	venvBaseDir string
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
	pipExe := vm.getPipExecutable(venvPath)
	cmd := exec.Command(pipExe, "list", "--format=freeze")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
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
func (vm *VenvManager) InstallPackage(venvName string, packageName string) error {
	venvPath := filepath.Join(vm.venvBaseDir, venvName)

	if !vm.VenvExists(venvName) {
		return fmt.Errorf("virtual environment '%s' does not exist", venvName)
	}

	pipExe := vm.getPipExecutable(venvPath)
	cmd := exec.Command(pipExe, "install", packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ListPackages lists installed packages in a virtual environment
func (vm *VenvManager) ListPackages(venvName string) error {
	venvPath := filepath.Join(vm.venvBaseDir, venvName)

	if !vm.VenvExists(venvName) {
		return fmt.Errorf("virtual environment '%s' does not exist", venvName)
	}

	pipExe := vm.getPipExecutable(venvPath)
	cmd := exec.Command(pipExe, "list")
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

	pipExe := vm.getPipExecutable(venvPath)
	cmd := exec.Command(pipExe, "install", "-r", requirementsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
