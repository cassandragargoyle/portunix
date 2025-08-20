package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"portunix.cz/app/install/apt"
)

// InstallPackage installs a package using the new configuration system
func InstallPackage(packageName, variant string) error {
	config, err := LoadInstallConfig()
	if err != nil {
		return fmt.Errorf("failed to load install config: %w", err)
	}

	pkg, platform, variantConfig, err := config.GetPackageInfo(packageName, variant)
	if err != nil {
		return err
	}

	// Show detailed installation information
	fmt.Println("════════════════════════════════════════════════")
	fmt.Printf("📦 INSTALLING: %s\n", pkg.Name)
	fmt.Println("════════════════════════════════════════════════")
	fmt.Printf("📄 Description: %s\n", pkg.Description)
	fmt.Printf("🔧 Variant: %s (v%s)\n", variant, variantConfig.Version)
	fmt.Printf("💻 Platform: %s\n", GetOperatingSystem())
	fmt.Printf("🏗️  Installation type: %s\n", platform.Type)

	// Show download URL if available
	if url, err := variantConfig.GetDownloadURL(); err == nil {
		fmt.Printf("🌐 Download URL: %s\n", url)
	}

	// Show install path or extract location
	if variantConfig.InstallPath != "" {
		fmt.Printf("📁 Install path: %s\n", variantConfig.InstallPath)
	} else if variantConfig.ExtractTo != "" {
		fmt.Printf("📁 Extract to: %s\n", variantConfig.ExtractTo)
	}

	// Show packages if it's a package manager install
	if len(variantConfig.Packages) > 0 {
		fmt.Printf("📋 Packages: %s\n", strings.Join(variantConfig.Packages, ", "))
	}

	fmt.Println("════════════════════════════════════════════════")

	// Check if package is already installed
	if platform.Verification.Command != "" {
		fmt.Printf("🔍 Checking if %s is already installed...\n", packageName)
		cmd := exec.Command("cmd", "/C", platform.Verification.Command)
		if runtime.GOOS != "windows" {
			cmd = exec.Command("sh", "-c", platform.Verification.Command)
		}

		err := cmd.Run()
		expectedExitCode := platform.Verification.ExpectedExitCode
		if err == nil && expectedExitCode == 0 {
			fmt.Printf("✅ %s is already installed and working!\n", pkg.Name)
			fmt.Println("No installation needed.")
			return nil
		} else if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				actualExitCode := exitErr.ExitCode()
				if actualExitCode == expectedExitCode {
					fmt.Printf("✅ %s is already installed and working!\n", pkg.Name)
					fmt.Println("No installation needed.")
					return nil
				}
			}
			fmt.Printf("📋 %s is not installed, proceeding with installation...\n", packageName)
		}
	}

	fmt.Println("🚀 Starting installation...")

	var installErr error
	switch platform.Type {
	case "msi", "exe":
		installErr = installWindowsBinary(platform, variantConfig)
	case "zip", "tar.gz":
		installErr = installArchive(platform, variantConfig)
	case "apt":
		installErr = installApt(platform, variantConfig)
	case "deb":
		installErr = installDeb(platform, variantConfig)
	case "snap":
		installErr = installSnap(platform, variantConfig)
	case "powershell":
		installErr = installPowerShell(platform, variantConfig)
	default:
		return fmt.Errorf("unsupported package type: %s", platform.Type)
	}

	if installErr != nil {
		fmt.Println("\n❌ Installation FAILED!")
		fmt.Printf("Error: %v\n", installErr)
		return installErr
	}

	fmt.Println("\n✅ Installation COMPLETED successfully!")

	// Show verification if available
	if platform.Verification.Command != "" {
		fmt.Printf("🔍 Verifying installation with: %s\n", platform.Verification.Command)
	}

	// Set environment variables automatically
	if len(platform.Environment) > 0 {
		fmt.Println("🌍 Setting environment variables...")
		envVarsSet := false
		for key, value := range platform.Environment {
			resolvedValue := ResolveVariables(value, map[string]string{
				"install_path": variantConfig.InstallPath,
				"extract_to":   variantConfig.ExtractTo,
			})

			if err := setEnvironmentVariable(key, resolvedValue); err != nil {
				fmt.Printf("⚠️  Warning: Failed to set %s=%s: %v\n", key, resolvedValue, err)
				fmt.Printf("   Please set manually: %s=%s\n", key, resolvedValue)
			} else {
				fmt.Printf("✅ Set %s=%s\n", key, resolvedValue)
				envVarsSet = true
			}
		}

		// If running in PowerShell and variables were set, reload them
		if envVarsSet && isRunningInPowerShell() {
			fmt.Println("🔄 Refreshing environment variables in current PowerShell session...")
			refreshPowerShellEnvironment()
		}
	}

	// Show post-install commands
	if len(variantConfig.PostInstall) > 0 {
		fmt.Println("⚙️  Post-install commands:")
		for _, cmd := range variantConfig.PostInstall {
			fmt.Printf("   %s\n", cmd)
		}
	}

	fmt.Println("════════════════════════════════════════════════")

	return nil
}

// installWindowsBinary installs MSI or EXE packages on Windows
func installWindowsBinary(platform *PlatformConfig, variant *VariantConfig) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("binary installer only supported on Windows")
	}

	// Get download URL for current architecture
	url, err := variant.GetDownloadURL()
	if err != nil {
		return err
	}

	// Get filename
	filename, err := variant.GetFileName()
	if err != nil {
		return err
	}

	// Download to cache
	cacheDir := ".cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cachedFile := filepath.Join(cacheDir, filename)

	// Check if already cached
	if _, err := os.Stat(cachedFile); os.IsNotExist(err) {
		fmt.Printf("Downloading %s...\n", filename)
		if err := downloadFile(cachedFile, url); err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}
	} else {
		fmt.Printf("Using cached %s\n", filename)
	}

	// Prepare install arguments
	args := platform.InstallArgs
	if len(variant.InstallArgs) > 0 {
		args = variant.InstallArgs
	}

	// Install
	fmt.Printf("Installing %s...\n", filename)
	if platform.Type == "msi" {
		cmd := exec.Command("msiexec", append([]string{"/i", cachedFile}, args...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	} else {
		cmd := exec.Command(cachedFile, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

// installArchive installs ZIP or TAR.GZ packages
func installArchive(platform *PlatformConfig, variant *VariantConfig) error {
	// Get download URL
	url, err := variant.GetDownloadURL()
	if err != nil {
		return err
	}

	// Get filename
	filename, err := variant.GetFileName()
	if err != nil {
		return err
	}

	// Download to cache
	cacheDir := ".cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cachedFile := filepath.Join(cacheDir, filename)

	// Download if not cached
	if _, err := os.Stat(cachedFile); os.IsNotExist(err) {
		fmt.Printf("Downloading %s...\n", filename)
		if err := downloadFile(cachedFile, url); err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}
	}

	// Extract
	extractTo := variant.ExtractTo
	if extractTo == "" {
		extractTo = "./extracted"
	}

	fmt.Printf("Extracting to %s...\n", extractTo)
	if err := os.MkdirAll(extractTo, 0755); err != nil {
		return fmt.Errorf("failed to create extract directory: %w", err)
	}

	// Extract based on file type
	if strings.HasSuffix(filename, ".zip") {
		return extractZip(cachedFile, extractTo)
	} else if strings.HasSuffix(filename, ".tar.gz") {
		return extractTarGz(cachedFile, extractTo)
	}

	return fmt.Errorf("unsupported archive format: %s", filename)
}

// installApt installs packages using apt package manager
func installApt(platform *PlatformConfig, variant *VariantConfig) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("apt installer only supported on Linux")
	}

	aptMgr := apt.NewAptManager()

	// Check if APT is supported
	if !aptMgr.IsSupported() {
		return fmt.Errorf("APT is not available on this system")
	}

	packages := variant.Packages
	if len(packages) == 0 {
		return fmt.Errorf("no packages specified for apt installation")
	}

	// Update package list first
	if err := aptMgr.Update(); err != nil {
		fmt.Printf("Warning: apt update failed: %v\n", err)
	}

	// Install packages using APT module
	if err := aptMgr.Install(packages); err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	// Run post-install commands if specified
	if len(variant.PostInstall) > 0 {
		fmt.Println("Running post-install commands...")
		for _, cmdStr := range variant.PostInstall {
			fmt.Printf("Running: %s\n", cmdStr)
			cmd := exec.Command("sh", "-c", cmdStr)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Warning: post-install command failed: %v\n", err)
			}
		}
	}

	return nil
}

// installDeb installs .deb packages
func installDeb(platform *PlatformConfig, variant *VariantConfig) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("deb installer only supported on Linux")
	}

	// Get download URL
	url, err := variant.GetDownloadURL()
	if err != nil {
		return err
	}

	// Get filename
	filename, err := variant.GetFileName()
	if err != nil {
		return err
	}

	// Download to cache
	cacheDir := ".cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cachedFile := filepath.Join(cacheDir, filename)

	// Download if not cached
	if _, err := os.Stat(cachedFile); os.IsNotExist(err) {
		fmt.Printf("Downloading %s...\n", filename)
		if err := downloadFile(cachedFile, url); err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}
	}

	// Install using dpkg
	fmt.Printf("Installing %s...\n", filename)
	cmd := exec.Command("sudo", "dpkg", "-i", cachedFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Try to fix broken dependencies
		fixCmd := exec.Command("sudo", "apt-get", "install", "-f")
		fixCmd.Stdout = os.Stdout
		fixCmd.Stderr = os.Stderr
		fixCmd.Run()
	}

	return nil
}

// installSnap installs packages using snap
func installSnap(platform *PlatformConfig, variant *VariantConfig) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("snap installer only supported on Linux")
	}

	packages := variant.Packages
	if len(packages) == 0 {
		return fmt.Errorf("no packages specified for snap installation")
	}

	for _, pkg := range packages {
		fmt.Printf("Installing snap package: %s\n", pkg)

		args := []string{"snap", "install"}
		args = append(args, variant.InstallArgs...)
		args = append(args, pkg)

		cmd := exec.Command("sudo", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install snap package %s: %w", pkg, err)
		}
	}

	return nil
}

func extractZip(zipFile, destDir string) error {
	fmt.Printf("Extracting ZIP file %s to %s...\n", zipFile, destDir)

	// Use PowerShell to extract ZIP file on Windows
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s' -Force", zipFile, destDir))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Use unzip command on Linux/Unix
	cmd := exec.Command("unzip", "-o", zipFile, "-d", destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extractTarGz(tarFile, destDir string) error {
	fmt.Printf("Extracting TAR.GZ file %s to %s...\n", tarFile, destDir)

	// Use tar command to extract
	cmd := exec.Command("tar", "-xzf", tarFile, "-C", destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// installPowerShell installs packages using PowerShell scripts
func installPowerShell(platform *PlatformConfig, variant *VariantConfig) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("PowerShell installer only supported on Windows")
	}

	if variant.InstallScript == "" {
		return fmt.Errorf("no install script specified for PowerShell installation")
	}

	fmt.Println("Executing PowerShell installation script...")

	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", variant.InstallScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("PowerShell installation failed: %w", err)
	}

	// Run post-install commands if specified
	if len(variant.PostInstall) > 0 {
		fmt.Println("Running post-install commands...")
		for _, cmdStr := range variant.PostInstall {
			fmt.Printf("Running: %s\n", cmdStr)
			postCmd := exec.Command("cmd", "/c", cmdStr)
			postCmd.Stdout = os.Stdout
			postCmd.Stderr = os.Stderr
			if err := postCmd.Run(); err != nil {
				fmt.Printf("Warning: post-install command failed: %v\n", err)
			}
		}
	}

	return nil
}

// setEnvironmentVariable sets system environment variable on Windows
func setEnvironmentVariable(key, value string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("environment variable setting only supported on Windows")
	}

	// Handle PATH_APPEND specially
	if key == "PATH_APPEND" {
		// Get current PATH
		cmd := exec.Command("powershell", "-Command", "[Environment]::GetEnvironmentVariable('PATH', 'Machine')")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get current PATH: %w", err)
		}

		currentPath := strings.TrimSpace(string(output))

		// Check if path is already in PATH
		if strings.Contains(currentPath, value) {
			fmt.Printf("   PATH already contains %s\n", value)
			return nil
		}

		// Append to PATH
		newPath := currentPath + ";" + value
		cmd = exec.Command("powershell", "-Command",
			fmt.Sprintf("[Environment]::SetEnvironmentVariable('PATH', '%s', 'Machine')", newPath))
		return cmd.Run()
	}

	// Set regular environment variable
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("[Environment]::SetEnvironmentVariable('%s', '%s', 'Machine')", key, value))
	return cmd.Run()
}

// isRunningInPowerShell detects if the application is running in PowerShell
func isRunningInPowerShell() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// Check for PowerShell-specific environment variables
	psModulePath := os.Getenv("PSModulePath")
	psDistribution := os.Getenv("POWERSHELL_DISTRIBUTION_CHANNEL")

	return psModulePath != "" || psDistribution != ""
}

// refreshPowerShellEnvironment refreshes environment variables in current PowerShell session
func refreshPowerShellEnvironment() {
	if runtime.GOOS != "windows" {
		return
	}

	// This script reloads machine environment variables into current PowerShell session
	script := `
	foreach($level in "Machine","User") {
		[Environment]::GetEnvironmentVariables($level).GetEnumerator() | % {
			if($_.Name -match "^(Path|PSModulePath)$") {
				$combined = (Get-ChildItem Env:$($_.Name)).Value + ";" + $_.Value
				Set-Item -Path "Env:$($_.Name)" -Value $combined
			} else {
				Set-Item -Path "Env:$($_.Name)" -Value $_.Value
			}
		}
	}
	Write-Host "Environment variables refreshed!"
	`

	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
