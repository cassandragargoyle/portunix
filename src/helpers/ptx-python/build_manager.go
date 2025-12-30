package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// BuildManager handles Python build and distribution operations
type BuildManager struct {
	venvManager *VenvManager
}

// NewBuildManager creates a new build manager
func NewBuildManager() (*BuildManager, error) {
	vm, err := NewVenvManager()
	if err != nil {
		return nil, err
	}

	return &BuildManager{
		venvManager: vm,
	}, nil
}

// BuildExeOptions holds options for building executables with PyInstaller
type BuildExeOptions struct {
	Script     string
	Name       string
	Icon       string
	OneFile    bool
	Console    bool
	Windowed   bool
	VenvName   string
	OutputDir  string
	ExtraArgs  []string
}

// BuildFreezeOptions holds options for building with cx_Freeze
type BuildFreezeOptions struct {
	Script        string
	Name          string
	Icon          string
	VenvName      string
	TargetVersion string
	OutputDir     string
	ExtraArgs     []string
}

// BuildExe builds a Python script into a standalone executable using PyInstaller
func (bm *BuildManager) BuildExe(opts BuildExeOptions) error {
	// Validate script exists
	if _, err := os.Stat(opts.Script); os.IsNotExist(err) {
		return fmt.Errorf("script file not found: %s", opts.Script)
	}

	// Ensure PyInstaller is installed
	if err := bm.ensurePyInstallerInstalled(opts.VenvName); err != nil {
		return fmt.Errorf("failed to ensure PyInstaller is installed: %v", err)
	}

	// Build PyInstaller command
	args := bm.buildPyInstallerArgs(opts)

	// Get pip executable path to determine venv
	var pyinstallerCmd string
	if opts.VenvName != "" {
		venvPath := filepath.Join(bm.venvManager.venvBaseDir, opts.VenvName)
		if runtime.GOOS == "windows" {
			pyinstallerCmd = filepath.Join(venvPath, "Scripts", "pyinstaller.exe")
		} else {
			pyinstallerCmd = filepath.Join(venvPath, "bin", "pyinstaller")
		}
	} else {
		pyinstallerCmd = "pyinstaller"
	}

	// Execute PyInstaller
	fmt.Printf("Building executable from %s...\n", opts.Script)
	cmd := exec.Command(pyinstallerCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(opts.Script)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("PyInstaller build failed: %v", err)
	}

	// Display success message
	outputPath := "dist"
	if opts.OutputDir != "" {
		outputPath = opts.OutputDir
	}

	exeName := opts.Name
	if exeName == "" {
		exeName = strings.TrimSuffix(filepath.Base(opts.Script), filepath.Ext(opts.Script))
	}
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}

	fmt.Printf("\n✅ Executable built successfully!\n")
	fmt.Printf("Output: %s/%s\n", outputPath, exeName)

	return nil
}

// buildPyInstallerArgs constructs PyInstaller command arguments
func (bm *BuildManager) buildPyInstallerArgs(opts BuildExeOptions) []string {
	args := []string{}

	// Basic script argument
	args = append(args, opts.Script)

	// Name option
	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}

	// One file mode
	if opts.OneFile {
		args = append(args, "--onefile")
	}

	// Console/Windowed mode
	if opts.Windowed {
		args = append(args, "--windowed", "--noconsole")
	} else if opts.Console {
		args = append(args, "--console")
	}

	// Icon
	if opts.Icon != "" {
		args = append(args, "--icon", opts.Icon)
	}

	// Output directory
	if opts.OutputDir != "" {
		args = append(args, "--distpath", opts.OutputDir)
	}

	// Clean build
	args = append(args, "--clean")

	// Extra arguments
	args = append(args, opts.ExtraArgs...)

	return args
}

// ensurePyInstallerInstalled checks if PyInstaller is installed and installs it if not
func (bm *BuildManager) ensurePyInstallerInstalled(venvName string) error {
	var checkCmd *exec.Cmd

	if venvName != "" {
		// Check in specific venv
		venvPath := filepath.Join(bm.venvManager.venvBaseDir, venvName)
		pipExe := bm.venvManager.getPipExecutable(venvPath)
		checkCmd = exec.Command(pipExe, "show", "pyinstaller")
	} else {
		// Check global Python
		checkCmd = exec.Command("pip", "show", "pyinstaller")
	}

	output, err := checkCmd.CombinedOutput()
	if err == nil && strings.Contains(string(output), "Name: pyinstaller") {
		// PyInstaller is already installed
		return nil
	}

	// Install PyInstaller
	fmt.Println("PyInstaller not found. Installing PyInstaller...")

	if venvName != "" {
		return bm.venvManager.InstallPackage(venvName, "pyinstaller")
	} else {
		installCmd := exec.Command("pip", "install", "pyinstaller")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		return installCmd.Run()
	}
}

// BuildFreeze builds a Python script using cx_Freeze
func (bm *BuildManager) BuildFreeze(opts BuildFreezeOptions) error {
	// Validate script exists
	if _, err := os.Stat(opts.Script); os.IsNotExist(err) {
		return fmt.Errorf("script file not found: %s", opts.Script)
	}

	// Ensure cx_Freeze is installed
	if err := bm.ensureCxFreezeInstalled(opts.VenvName); err != nil {
		return fmt.Errorf("failed to ensure cx_Freeze is installed: %v", err)
	}

	// Build cx_Freeze command
	args := []string{opts.Script, "build"}

	if opts.OutputDir != "" {
		args = append(args, "--build-exe", opts.OutputDir)
	}

	// Get cxfreeze executable
	var cxfreezeCmd string
	if opts.VenvName != "" {
		venvPath := filepath.Join(bm.venvManager.venvBaseDir, opts.VenvName)
		if runtime.GOOS == "windows" {
			cxfreezeCmd = filepath.Join(venvPath, "Scripts", "cxfreeze.exe")
		} else {
			cxfreezeCmd = filepath.Join(venvPath, "bin", "cxfreeze")
		}
	} else {
		cxfreezeCmd = "cxfreeze"
	}

	// Execute cx_Freeze
	fmt.Printf("Building executable with cx_Freeze from %s...\n", opts.Script)
	cmd := exec.Command(cxfreezeCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(opts.Script)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cx_Freeze build failed: %v", err)
	}

	fmt.Printf("\n✅ Executable built successfully with cx_Freeze!\n")
	return nil
}

// ensureCxFreezeInstalled checks if cx_Freeze is installed and installs it if not
func (bm *BuildManager) ensureCxFreezeInstalled(venvName string) error {
	var checkCmd *exec.Cmd

	if venvName != "" {
		venvPath := filepath.Join(bm.venvManager.venvBaseDir, venvName)
		pipExe := bm.venvManager.getPipExecutable(venvPath)
		checkCmd = exec.Command(pipExe, "show", "cx_Freeze")
	} else {
		checkCmd = exec.Command("pip", "show", "cx_Freeze")
	}

	output, err := checkCmd.CombinedOutput()
	if err == nil && strings.Contains(string(output), "Name: cx-Freeze") {
		return nil
	}

	// Install cx_Freeze
	fmt.Println("cx_Freeze not found. Installing cx_Freeze...")

	if venvName != "" {
		return bm.venvManager.InstallPackage(venvName, "cx_Freeze")
	} else {
		installCmd := exec.Command("pip", "install", "cx_Freeze")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		return installCmd.Run()
	}
}

// BuildWheel builds a Python wheel distribution
func (bm *BuildManager) BuildWheel(venvName string, projectPath string) error {
	// Ensure build tools are installed
	if err := bm.ensureBuildToolsInstalled(venvName); err != nil {
		return fmt.Errorf("failed to ensure build tools installed: %v", err)
	}

	var pythonExe string
	if venvName != "" {
		venvPath := filepath.Join(bm.venvManager.venvBaseDir, venvName)
		pythonExe = bm.venvManager.getPythonExecutable(venvPath)
	} else {
		if runtime.GOOS == "windows" {
			pythonExe = "python"
		} else {
			pythonExe = "python3"
		}
	}

	// Build wheel
	fmt.Println("Building wheel distribution...")
	cmd := exec.Command(pythonExe, "-m", "build", "--wheel")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if projectPath != "" {
		cmd.Dir = projectPath
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wheel build failed: %v", err)
	}

	fmt.Printf("\n✅ Wheel distribution built successfully!\n")
	fmt.Println("Output: dist/*.whl")
	return nil
}

// BuildSdist builds a Python source distribution
func (bm *BuildManager) BuildSdist(venvName string, projectPath string) error {
	// Ensure build tools are installed
	if err := bm.ensureBuildToolsInstalled(venvName); err != nil {
		return fmt.Errorf("failed to ensure build tools installed: %v", err)
	}

	var pythonExe string
	if venvName != "" {
		venvPath := filepath.Join(bm.venvManager.venvBaseDir, venvName)
		pythonExe = bm.venvManager.getPythonExecutable(venvPath)
	} else {
		if runtime.GOOS == "windows" {
			pythonExe = "python"
		} else {
			pythonExe = "python3"
		}
	}

	// Build sdist
	fmt.Println("Building source distribution...")
	cmd := exec.Command(pythonExe, "-m", "build", "--sdist")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if projectPath != "" {
		cmd.Dir = projectPath
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sdist build failed: %v", err)
	}

	fmt.Printf("\n✅ Source distribution built successfully!\n")
	fmt.Println("Output: dist/*.tar.gz")
	return nil
}

// ensureBuildToolsInstalled ensures build, wheel, and setuptools are installed
func (bm *BuildManager) ensureBuildToolsInstalled(venvName string) error {
	packages := []string{"build", "wheel", "setuptools"}

	for _, pkg := range packages {
		var checkCmd *exec.Cmd

		if venvName != "" {
			venvPath := filepath.Join(bm.venvManager.venvBaseDir, venvName)
			pipExe := bm.venvManager.getPipExecutable(venvPath)
			checkCmd = exec.Command(pipExe, "show", pkg)
		} else {
			checkCmd = exec.Command("pip", "show", pkg)
		}

		output, err := checkCmd.CombinedOutput()
		if err == nil && strings.Contains(string(output), fmt.Sprintf("Name: %s", pkg)) {
			continue
		}

		// Install package
		fmt.Printf("Installing %s...\n", pkg)
		if venvName != "" {
			if err := bm.venvManager.InstallPackage(venvName, pkg); err != nil {
				return err
			}
		} else {
			installCmd := exec.Command("pip", "install", pkg)
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			if err := installCmd.Run(); err != nil {
				return err
			}
		}
	}

	return nil
}
