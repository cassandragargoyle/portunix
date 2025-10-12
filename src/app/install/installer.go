package install

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"portunix.ai/app/install/ansible_galaxy"
	"portunix.ai/app/install/apt"
	"portunix.ai/app/install/pip"
)

// InstallPackage installs a package using the new configuration system
func InstallPackage(packageName, variant string) error {
	return InstallPackageWithDryRun(packageName, variant, false)
}

// InstallPackageWithOptions installs a package using the configuration system with advanced options
//
// This function is called when users use the enhanced installation parameters:
//
// Examples of CLI commands that trigger this function:
//   portunix install hugo --method=apt              # Override default installation method to use APT
//   portunix install hugo --method=snap             # Override default installation method to use Snap
//   portunix install hugo --version=latest          # Install latest stable release (e.g., v1.0.0, not beta/rc versions)
//   portunix install hugo --version=prerelease      # Install latest prerelease (beta, rc, or development versions)
//   portunix install hugo --version=v0.140.0        # Install specific version number
//   portunix install hugo --list-methods            # Show available methods (handled before this function)
//   portunix install hugo --dry-run                 # Preview installation without executing
//   portunix install hugo --method=apt --dry-run    # Combine method override with dry-run
//
// The function flow:
// 1. Load package configuration from assets/install-packages.json
// 2. Resolve the appropriate variant based on:
//    - Current OS platform (windows/linux/darwin) - each OS has different available methods
//    - Method override (if --method specified) - must exist for current OS
//    - Version requirements (if --version specified) - selects method capable of that version
//    - Default/preferred variant for the OS if no overrides
// 3. Delegate to InstallPackageWithDryRun with resolved variant
//
// Parameters are passed through InstallOptions struct which contains:
//   - PackageName: The package to install (e.g., "hugo")
//   - Method: Override method if specified (e.g., "apt", "snap", "deb")
//   - Version: Version request if specified (e.g., "latest", "prerelease", "v1.2.3")
//   - DryRun: Whether to only preview the installation
//   - Variant: Legacy variant parameter for backward compatibility
func InstallPackageWithOptions(options *InstallOptions) error {
	// First try the new registry system
	registry, err := LoadPackageRegistry("./assets")
	if err == nil {
		// Check if package exists in new registry
		pkg, err := registry.GetPackage(options.PackageName)
		if err == nil {
			// Package found in new registry, use it
			variant, err := resolveVariantFromRegistry(pkg, options)
			if err != nil {
				return err
			}
			// Use direct registry-based installation
			return installPackageFromRegistry(pkg, variant, options.DryRun)
		}
	}

	// Fall back to old config system
	config, err := LoadInstallConfig()
	if err != nil {
		return fmt.Errorf("failed to load install config: %w", err)
	}

	variant, err := resolveVariantWithMethodAndVersion(config, options)
	if err != nil {
		return err
	}

	return InstallPackageWithDryRun(options.PackageName, variant, options.DryRun)
}

// resolveVariantWithMethodAndVersion determines the effective variant based on method and version overrides
func resolveVariantWithMethodAndVersion(config *InstallConfig, options *InstallOptions) (string, error) {
	pkg, exists := config.Packages[options.PackageName]
	if !exists {
		return "", fmt.Errorf("package '%s' not found", options.PackageName)
	}

	// Get current OS platform
	currentOS := GetOperatingSystem()
	platform, exists := pkg.Platforms[currentOS]
	if !exists {
		if currentOS == "windows_sandbox" {
			platform, exists = pkg.Platforms["windows"]
		}
		if !exists {
			return "", fmt.Errorf("no platform configuration for %s", currentOS)
		}
	}

	// Handle version-specific logic first
	if options.Version != "" {
		return resolveVersionBasedVariant(&platform, options.Version)
	}

	// If method is specified, use it as variant (method override)
	if options.Method != "" {
		// Verify the method exists
		if _, exists := platform.Variants[options.Method]; !exists {
			return "", fmt.Errorf("method '%s' not found for package '%s' on platform '%s'", options.Method, options.PackageName, currentOS)
		}
		return options.Method, nil
	}

	// Otherwise use explicit variant
	if options.Variant != "" {
		// Verify the variant exists
		if _, exists := platform.Variants[options.Variant]; !exists {
			return "", fmt.Errorf("variant '%s' not found for package '%s' on platform '%s'", options.Variant, options.PackageName, currentOS)
		}
		return options.Variant, nil
	}

	// No override, use default
	return "", nil
}

// resolveVersionBasedVariant finds the best variant for the requested version
func resolveVersionBasedVariant(platform *PlatformConfig, versionRequest string) (string, error) {
	switch versionRequest {
	case "latest":
		// Find variant with latest version capability
		return findVariantWithLatestCapability(platform)
	case "prerelease":
		// Find variant with prerelease capability
		return findVariantWithPrereleaseCapability(platform)
	default:
		// Specific version - find variant that supports it
		return findVariantForSpecificVersion(platform, versionRequest)
	}
}

// findVariantWithLatestCapability finds a variant that can provide the latest version
func findVariantWithLatestCapability(platform *PlatformConfig) (string, error) {
	// Priority order: deb-latest, tar.gz direct downloads, snap, repository methods
	preferredOrder := []string{"deb-latest", "deb", "tar.gz", "zip", "snap", "apt", "repository"}

	for _, preferredType := range preferredOrder {
		for variantName, variant := range platform.Variants {
			methodType := platform.Type
			if variant.Type != "" {
				methodType = variant.Type
			}

			// Check if this variant can provide latest versions
			if methodType == preferredType || (preferredType == "deb-latest" && strings.Contains(variantName, "latest")) {
				return variantName, nil
			}
		}
	}

	// Fallback to first available variant
	for variantName := range platform.Variants {
		return variantName, nil
	}

	return "", fmt.Errorf("no variants available")
}

// findVariantWithPrereleaseCapability finds a variant that can provide prerelease versions
func findVariantWithPrereleaseCapability(platform *PlatformConfig) (string, error) {
	// Look for variants with prerelease support indicators
	for variantName, variant := range platform.Variants {
		methodType := platform.Type
		if variant.Type != "" {
			methodType = variant.Type
		}

		// Direct downloads are more likely to have prerelease access
		if methodType == "deb" || methodType == "tar.gz" || methodType == "zip" {
			return variantName, nil
		}

		// Check if variant name suggests prerelease capability
		if strings.Contains(strings.ToLower(variantName), "latest") ||
		   strings.Contains(strings.ToLower(variantName), "prerelease") {
			return variantName, nil
		}
	}

	// Fallback to latest capability
	return findVariantWithLatestCapability(platform)
}

// findVariantForSpecificVersion finds a variant that can install a specific version
func findVariantForSpecificVersion(platform *PlatformConfig, version string) (string, error) {
	// For specific versions, direct downloads are usually the best option
	preferredOrder := []string{"deb", "tar.gz", "zip", "msi", "exe"}

	for _, preferredType := range preferredOrder {
		for variantName, variant := range platform.Variants {
			methodType := platform.Type
			if variant.Type != "" {
				methodType = variant.Type
			}

			if methodType == preferredType {
				return variantName, nil
			}
		}
	}

	// Fallback to any available variant
	for variantName := range platform.Variants {
		return variantName, nil
	}

	return "", fmt.Errorf("no suitable variant found for version %s", version)
}

// InstallPackageWithDryRun installs a package using the new configuration system with dry-run support
func InstallPackageWithDryRun(packageName, variant string, dryRun bool) error {
	config, err := LoadInstallConfig()
	if err != nil {
		return fmt.Errorf("failed to load install config: %w", err)
	}

	// Auto-detect best variant for Linux distributions if no variant specified
	if variant == "" && runtime.GOOS == "linux" {
		if autoVariant, err := config.FindBestVariantForDistribution(packageName); err == nil {
			variant = autoVariant
			fmt.Printf("🔍 Auto-detected variant: %s\n", variant)
		}
	}

	pkg, platform, variantConfig, err := config.GetPackageInfo(packageName, variant)
	if err != nil {
		return err
	}

	// If no variant was specified, use the default variant for display
	if variant == "" {
		variant = pkg.DefaultVariant
	}

	// Handle prerequisites
	if err := handlePrerequisites(config, pkg, dryRun); err != nil {
		return fmt.Errorf("failed to handle prerequisites: %w", err)
	}

	// Determine actual install type
	installType := platform.Type
	if variantConfig.Type != "" {
		installType = variantConfig.Type
	}

	// Show detailed installation information
	fmt.Println("════════════════════════════════════════════════")
	fmt.Printf("📦 INSTALLING: %s\n", pkg.Name)
	fmt.Println("════════════════════════════════════════════════")
	fmt.Printf("📄 Description: %s\n", pkg.Description)
	fmt.Printf("🔧 Variant: %s (v%s)\n", variant, variantConfig.Version)
	fmt.Printf("💻 Platform: %s\n", GetOperatingSystem())
	fmt.Printf("🏗️  Installation type: %s\n", installType)

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

	// Check if package is already installed (skip in dry-run mode)
	if !dryRun && platform.Verification.Command != "" {
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

	if dryRun {
		fmt.Println("🔍 DRY-RUN MODE: Showing what would be installed")
		fmt.Printf("💡 To execute for real, remove the --dry-run flag\n")
		fmt.Println("════════════════════════════════════════════════")
		return nil
	}

	fmt.Println("🚀 Starting installation...")

	var installErr error

	// Use variant type if specified, otherwise use platform type (already determined above)

	switch installType {
	case "msi", "exe":
		installErr = installWindowsBinary(platform, variantConfig)
	case "zip", "tar.gz", "tar.xz":
		installErr = installArchive(platform, variantConfig)
	case "apt":
		installErr = installApt(platform, variantConfig)
	case "dnf", "yum":
		installErr = installDnf(platform, variantConfig)
	case "pacman":
		installErr = installPacman(platform, variantConfig)
	case "deb":
		installErr = installDeb(platform, variantConfig)
	case "snap":
		installErr = installSnap(platform, variantConfig)
	case "pip":
		installErr = installPip(platform, variantConfig)
	case "pipx":
		installErr = installPipx(platform, variantConfig)
	case "ansible-galaxy":
		installErr = installAnsibleGalaxy(platform, variantConfig)
	case "powershell":
		installErr = installPowerShell(platform, variantConfig)
	case "repository":
		installErr = installRepository(platform, variantConfig)
	case "direct_download":
		installErr = installDirectDownload(platform, variantConfig)
	case "redirect":
		installErr = installRedirect(platform, variantConfig)
	default:
		return fmt.Errorf("unsupported package type: %s", installType)
	}

	if installErr != nil {
		// Check if we should trigger fallback
		fallbackManager := NewFallbackManager()
		versionMatcher := NewVersionMatcher()

		// Get current OS version for fallback decision
		var currentVersion string
		if runtime.GOOS == "linux" {
			_, version, err := GetLinuxDistribution()
			if err == nil {
				currentVersion = version
			}
		}

		// Determine support level for better error handling
		var supportLevel SupportLevel = Supported
		if len(variantConfig.SupportedVersionRanges) > 0 && currentVersion != "" {
			supportLevel = versionMatcher.IsVersionSupported(currentVersion, variantConfig.SupportedVersionRanges, variantConfig.SupportedVersions)
		}

		// Try fallback if configured and appropriate
		if fallbackManager.ShouldTriggerFallback(installErr, supportLevel) && len(variantConfig.FallbackVariants) > 0 {
			fmt.Printf("\n🔄 Installation failed, trying fallback options...\n")
			return fallbackManager.ExecuteFallback(packageName, variant, config, installErr.Error(), variantConfig.FallbackVariants, variantConfig.FallbackStrategy)
		}

		// Enhanced error message
		fmt.Println("\n❌ Installation FAILED!")
		fmt.Printf("Package: %s\n", pkg.Name)
		fmt.Printf("Variant: %s\n", variant)
		fmt.Printf("Error: %v\n", installErr)

		// Provide suggestions based on the error
		if len(variantConfig.FallbackVariants) > 0 {
			fmt.Println("\n💡 Suggestions:")
			fmt.Printf("Try alternative variants: %s\n", strings.Join(variantConfig.FallbackVariants, ", "))
			fmt.Printf("Command: ./portunix install %s --variant %s\n", packageName, variantConfig.FallbackVariants[0])
		}

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

	// Post-install commands are handled by individual installer functions
	// (installArchive, installApt, etc.) where appropriate variables are available

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

	// Download to cache with proper filename resolution
	cacheDir := ".cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Try to get filename from URL path first (for cases without redirects)
	fallbackFilename, _ := variant.GetFileName()
	cachedFile := filepath.Join(cacheDir, fallbackFilename)

	// Check if already cached with fallback filename
	if _, err := os.Stat(cachedFile); os.IsNotExist(err) {
		// Use enhanced download with proper filename resolution
		filename, err := downloadFileWithProperFilename(url, cacheDir)
		if err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}
		cachedFile = filepath.Join(cacheDir, filename)
		fmt.Printf("Successfully downloaded as: %s\n", filename)
	} else {
		fmt.Printf("Using cached %s\n", fallbackFilename)
	}

	// Verify the downloaded file exists and is not empty
	if fileInfo, err := os.Stat(cachedFile); err != nil {
		return fmt.Errorf("downloaded file not found: %s", cachedFile)
	} else if fileInfo.Size() == 0 {
		return fmt.Errorf("downloaded file is empty: %s", cachedFile)
	}

	filename := filepath.Base(cachedFile)

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
	actualExtractTo := extractTo
	if strings.HasSuffix(filename, ".zip") {
		if err := extractZip(cachedFile, extractTo); err != nil {
			return err
		}
	} else if strings.HasSuffix(filename, ".tar.gz") {
		var err error
		actualExtractTo, err = extractTarGz(cachedFile, extractTo)
		if err != nil {
			return err
		}
	} else if strings.HasSuffix(filename, ".tar.xz") {
		if err := extractTarXz(cachedFile, extractTo); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported archive format: %s", filename)
	}

	// Run post-install commands
	extraVariables := map[string]string{
		"downloaded_file":    cachedFile,
		"extract_to":         extractTo,
		"actual_extract_to":  actualExtractTo,
	}
	if err := runPostInstallCommands(variant, extraVariables); err != nil {
		return fmt.Errorf("post-install commands failed: %w", err)
	}

	return nil
}

// installApt installs packages using apt package manager
func installApt(platform *PlatformConfig, variant *VariantConfig) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("apt installer only supported on Linux")
	}

	aptMgr := apt.NewAptManager()
	
	// Set dry-run mode if we're in dry-run mode
	// (This needs to be passed from the main function - for now defaulting to false)
	aptMgr.DryRun = false // TODO: Pass dry-run flag from main installer

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

	// Run post-install commands
	extraVariables := map[string]string{}
	if err := runPostInstallCommands(variant, extraVariables); err != nil {
		return fmt.Errorf("post-install commands failed: %w", err)
	}

	return nil
}

// installDnf installs packages using dnf package manager
func installDnf(platform *PlatformConfig, variant *VariantConfig) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("dnf installer only supported on Linux")
	}

	packages := variant.Packages
	if len(packages) == 0 {
		return fmt.Errorf("no packages specified for dnf installation")
	}

	// Check if dnf is available
	if _, err := exec.LookPath("dnf"); err != nil {
		return fmt.Errorf("DNF is not available on this system")
	}

	// Update package list first  
	fmt.Println("📋 Updating package metadata...")
	updateCmd := exec.Command("sudo", "dnf", "check-update")
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr
	// Note: dnf check-update returns exit code 100 when updates are available, which is normal
	if err := updateCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 100 {
			fmt.Println("✅ Package metadata updated (updates available)")
		} else {
			fmt.Printf("⚠️  Warning: dnf check-update failed: %v\n", err)
		}
	} else {
		fmt.Println("✅ Package metadata updated")
	}

	// Install packages
	fmt.Printf("📦 Installing packages: %s\n", strings.Join(packages, ", "))
	installCmd := exec.Command("sudo", "dnf", "install", "-y")
	installCmd.Args = append(installCmd.Args, packages...)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install packages via dnf: %w", err)
	}

	// Run post-install commands
	extraVariables := map[string]string{}
	if err := runPostInstallCommands(variant, extraVariables); err != nil {
		return fmt.Errorf("post-install commands failed: %w", err)
	}

	return nil
}

// installPacman installs packages using pacman package manager
func installPacman(platform *PlatformConfig, variant *VariantConfig) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("pacman installer only supported on Linux")
	}

	packages := variant.Packages
	if len(packages) == 0 {
		return fmt.Errorf("no packages specified for pacman installation")
	}

	// Check if pacman is available
	if _, err := exec.LookPath("pacman"); err != nil {
		return fmt.Errorf("pacman is not available on this system")
	}

	// Determine sudo prefix
	sudoPrefix := determineSudoPrefix()

	// Update package database first
	fmt.Println("📋 Updating package database...")
	var updateCmd *exec.Cmd
	if sudoPrefix == "" {
		updateCmd = exec.Command("pacman", "-Sy")
	} else {
		updateCmd = exec.Command("sudo", "pacman", "-Sy")
	}
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr
	if err := updateCmd.Run(); err != nil {
		fmt.Printf("⚠️  Warning: pacman -Sy failed: %v\n", err)
	} else {
		fmt.Println("✅ Package database updated")
	}

	// Install packages
	fmt.Printf("📦 Installing packages: %s\n", strings.Join(packages, ", "))
	var installCmd *exec.Cmd
	if sudoPrefix == "" {
		installCmd = exec.Command("pacman", "-S", "--noconfirm")
	} else {
		installCmd = exec.Command("sudo", "pacman", "-S", "--noconfirm")
	}
	installCmd.Args = append(installCmd.Args, packages...)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install packages via pacman: %w", err)
	}

	// Run post-install commands
	extraVariables := map[string]string{}
	if err := runPostInstallCommands(variant, extraVariables); err != nil {
		return fmt.Errorf("post-install commands failed: %w", err)
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

func extractTarGz(tarFile, destDir string) (string, error) {
	fmt.Printf("Extracting TAR.GZ file %s to %s...\n", tarFile, destDir)

	// Check if we need elevated permissions for the destination directory
	requiresSudo, fallbackDir := checkDirectoryPermissions(destDir)
	actualDestDir := destDir

	if requiresSudo {
		fmt.Printf("🔐 Destination directory %s requires elevated permissions\n", destDir)

		// Check if sudo is available
		if isSudoAvailable() {
			fmt.Println("✅ Using sudo for extraction...")
			cmd := exec.Command("sudo", "tar", "-xzf", tarFile, "-C", destDir)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				// Sudo failed (likely due to authentication), fall back to user directory
				fmt.Printf("⚠️  Sudo extraction failed (%v), falling back to user directory: %s\n", err, fallbackDir)
				actualDestDir = fallbackDir

				// Create fallback directory if it doesn't exist
				if err := os.MkdirAll(fallbackDir, 0755); err != nil {
					return fallbackDir, fmt.Errorf("failed to create fallback directory %s: %w", fallbackDir, err)
				}

				// Add user bin directory to PATH if not already present
				if err := addUserBinToPath(fallbackDir); err != nil {
					fmt.Printf("⚠️  Warning: Failed to add %s to PATH: %v\n", fallbackDir, err)
					fmt.Printf("💡 Please add '%s' to your PATH manually:\n", fallbackDir)
					fmt.Printf("   For bash: echo 'export PATH=\"%s:$PATH\"' >> ~/.bashrc\n", fallbackDir)
					fmt.Printf("   For zsh: echo 'export PATH=\"%s:$PATH\"' >> ~/.zshrc\n", fallbackDir)
				} else {
					fmt.Printf("✅ Added %s to PATH\n", fallbackDir)
				}
			} else {
				return destDir, nil
			}
		} else {
			// Fallback to user directory
			fmt.Printf("⚠️  Sudo not available, falling back to user directory: %s\n", fallbackDir)
			actualDestDir = fallbackDir

			// Create fallback directory if it doesn't exist
			if err := os.MkdirAll(fallbackDir, 0755); err != nil {
				return fallbackDir, fmt.Errorf("failed to create fallback directory %s: %w", fallbackDir, err)
			}

			// Add user bin directory to PATH if not already present
			if err := addUserBinToPath(fallbackDir); err != nil {
				fmt.Printf("⚠️  Warning: Failed to add %s to PATH: %v\n", fallbackDir, err)
				fmt.Printf("💡 Please add '%s' to your PATH manually:\n", fallbackDir)
				fmt.Printf("   For bash: echo 'export PATH=\"%s:$PATH\"' >> ~/.bashrc\n", fallbackDir)
				fmt.Printf("   For zsh: echo 'export PATH=\"%s:$PATH\"' >> ~/.zshrc\n", fallbackDir)
			} else {
				fmt.Printf("✅ Added %s to PATH\n", fallbackDir)
			}
		}
	}

	// Use tar command to extract to actual destination
	cmd := exec.Command("tar", "-xzf", tarFile, "-C", actualDestDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return actualDestDir, err
	}
	return actualDestDir, nil
}

func extractTarXz(tarFile, destDir string) error {
	fmt.Printf("Extracting TAR.XZ file %s to %s...\n", tarFile, destDir)

	// Check if xz-utils is available, install if needed
	if runtime.GOOS == "linux" {
		checkCmd := exec.Command("which", "xz")
		if err := checkCmd.Run(); err != nil {
			fmt.Println("🔧 Installing xz-utils for .tar.xz extraction...")
			
			// Try to install xz-utils based on available package manager
			installCmd := exec.Command("apt-get", "update", "&&", "apt-get", "install", "-y", "xz-utils")
			if isRunningAsRoot() {
				installCmd = exec.Command("sh", "-c", "apt-get update && apt-get install -y xz-utils")
			} else {
				installCmd = exec.Command("sh", "-c", "sudo apt-get update && sudo apt-get install -y xz-utils")
			}
			
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			if err := installCmd.Run(); err != nil {
				// If apt-get fails, try other package managers
				fmt.Println("⚠️  apt-get failed, trying alternative package managers...")
				
				// Try dnf/yum for RedHat-based systems
				if isRunningAsRoot() {
					installCmd = exec.Command("sh", "-c", "dnf install -y xz || yum install -y xz")
				} else {
					installCmd = exec.Command("sh", "-c", "sudo dnf install -y xz || sudo yum install -y xz")
				}
				installCmd.Stdout = os.Stdout
				installCmd.Stderr = os.Stderr
				installCmd.Run() // Try but don't fail if this doesn't work either
			}
		}
	}

	// Use tar command to extract with xz compression
	cmd := exec.Command("tar", "-xJf", tarFile, "-C", destDir)
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

	// Run post-install commands
	extraVariables := map[string]string{}
	if err := runPostInstallCommands(variant, extraVariables); err != nil {
		return fmt.Errorf("post-install commands failed: %w", err)
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

// ensureRequiredTools checks and installs required tools for repository operations
func ensureRequiredTools(distro string) error {
	fmt.Println("🔍 Checking for required tools...")

	// Define required tools based on distribution
	var requiredTools []string
	var packageManager string
	var installCmd []string

	switch distro {
	case "ubuntu", "kubuntu", "debian", "mint", "elementary":
		requiredTools = []string{"wget", "sudo", "lsb-release", "ca-certificates"}
		packageManager = "apt-get"

		// First, check if we can run commands without sudo (in containers often running as root)
		testCmd := exec.Command("id", "-u")
		output, _ := testCmd.Output()
		isRoot := strings.TrimSpace(string(output)) == "0"

		if isRoot {
			// Running as root, no sudo needed
			installCmd = []string{"apt-get", "install", "-y"}
		} else {
			// Not root, need sudo
			installCmd = []string{"sudo", "apt-get", "install", "-y"}
		}

	case "fedora", "rocky":
		requiredTools = []string{"wget", "sudo", "ca-certificates"}
		packageManager = "dnf"

		// Check if running as root
		testCmd := exec.Command("id", "-u")
		output, _ := testCmd.Output()
		isRoot := strings.TrimSpace(string(output)) == "0"

		if isRoot {
			installCmd = []string{"dnf", "install", "-y"}
		} else {
			installCmd = []string{"sudo", "dnf", "install", "-y"}
		}

	default:
		// Unknown distribution, skip tool check
		fmt.Println("⚠️  Unknown distribution, skipping tool check")
		return nil
	}

	// Check which tools are missing
	var missingTools []string
	for _, tool := range requiredTools {
		// Special handling for lsb-release (it's lsb_release command)
		checkTool := tool
		if tool == "lsb-release" {
			checkTool = "lsb_release"
		}

		// Check if tool exists
		checkCmd := exec.Command("which", checkTool)
		if err := checkCmd.Run(); err != nil {
			// Tool not found
			missingTools = append(missingTools, tool)
		}
	}

	// If no tools missing, we're done
	if len(missingTools) == 0 {
		fmt.Println("✅ All required tools are already installed")
		return nil
	}

	// Install missing tools
	fmt.Printf("📦 Installing missing tools: %s\n", strings.Join(missingTools, ", "))

	// Update package lists first (for apt-based systems)
	if packageManager == "apt-get" {
		fmt.Println("📋 Updating package lists...")
		updateCmd := exec.Command(installCmd[0], "update")
		updateCmd.Stdout = os.Stdout
		updateCmd.Stderr = os.Stderr
		if err := updateCmd.Run(); err != nil {
			// If running as non-root and apt-get fails, we might not have sudo installed
			// Try without sudo first
			if !strings.Contains(installCmd[0], "sudo") {
				fmt.Printf("⚠️  Warning: apt-get update failed: %v\n", err)
			}
		}
	}

	// Install the missing tools
	fullCmd := append(installCmd, missingTools...)
	fmt.Printf("   Command: %s\n", strings.Join(fullCmd, " "))

	cmd := exec.Command(fullCmd[0], fullCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install required tools: %w", err)
	}

	fmt.Println("✅ Required tools installed successfully")
	return nil
}

// installRepository installs packages using distribution-specific repositories
func installRepository(platform *PlatformConfig, variant *VariantConfig) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("repository installer only supported on Linux")
	}

	// Get current distribution
	distro, version, err := GetLinuxDistribution()
	if err != nil {
		return fmt.Errorf("failed to detect Linux distribution: %w", err)
	}

	fmt.Printf("Detected distribution: %s %s\n", distro, version)

	// Ensure required tools are installed before repository setup
	if err := ensureRequiredTools(distro); err != nil {
		return fmt.Errorf("failed to ensure required tools: %w", err)
	}

	// Check if this variant supports the current distribution
	distributionsList := variant.GetDistributionsList()
	if len(distributionsList) > 0 {
		supported := false
		for _, supportedDistro := range distributionsList {
			if supportedDistro == distro || supportedDistro == "universal" {
				supported = true
				break
			}
		}
		if !supported {
			return fmt.Errorf("this variant does not support %s distribution", distro)
		}
	}

	// Enhanced version compatibility checking using new version matcher
	versionMatcher := NewVersionMatcher()

	// Check version compatibility if specified (skip for universal distributions)
	universalVariant := false
	distributionsList2 := variant.GetDistributionsList()
	if len(distributionsList2) > 0 {
		for _, supportedDistro := range distributionsList2 {
			if supportedDistro == "universal" {
				universalVariant = true
				break
			}
		}
	}

	if !universalVariant {
		// Use new version range support if available
		var supportLevel SupportLevel
		if len(variant.SupportedVersionRanges) > 0 {
			supportLevel = versionMatcher.IsVersionSupported(version, variant.SupportedVersionRanges, variant.SupportedVersions)
		} else if len(variant.SupportedVersions) > 0 {
			// Fallback to legacy version checking
			supportLevel = versionMatcher.IsVersionSupported(version, []VersionRange{}, variant.SupportedVersions)
		} else {
			// No version restrictions
			supportLevel = Supported
		}

		// Show version support message
		fmt.Printf("🔍 Version support: %s\n", versionMatcher.GetVersionSupportMessage(version, supportLevel))

		// Handle unsupported versions
		if supportLevel == Unsupported {
			fallbackErr := fmt.Errorf("%s version %s not supported for this variant", distro, version)

			// Try fallback if configured
			if len(variant.FallbackVariants) > 0 {
				fallbackManager := NewFallbackManager()
				reason := fmt.Sprintf("%s %s not explicitly supported for this variant", distro, version)
				return fallbackManager.ExecuteFallback("powershell", "ubuntu", nil, reason, variant.FallbackVariants, variant.FallbackStrategy)
			}

			return fallbackErr
		}

		// Show warning for experimental versions
		if supportLevel == Experimental {
			fmt.Printf("⚠️  WARNING: %s %s is newer than tested versions. Proceeding anyway...\n", distro, version)
		}
	}

	// Run repository setup commands
	if len(variant.RepositorySetup) > 0 {
		fmt.Println("🔧 Setting up package repository...")
		fmt.Println("════════════════════════════════════════════════")

		// Define step descriptions for common repository setup operations
		stepDescriptions := map[int]string{
			0: "Download Microsoft repository configuration package",
			1: "Install Microsoft repository configuration",
			2: "Clean up temporary files",
			3: "Update package lists with new repository",
		}

		for i, cmdStr := range variant.RepositorySetup {
			// Get step description if available
			stepDesc := ""
			if desc, ok := stepDescriptions[i]; ok {
				stepDesc = desc
			}

			fmt.Printf("\n📌 Step %d/%d", i+1, len(variant.RepositorySetup))
			if stepDesc != "" {
				fmt.Printf(": %s", stepDesc)
			}
			fmt.Println()

			// Resolve variables in command
			resolvedCmd := ResolveVariables(cmdStr, map[string]string{
				"distribution": distro,
				"version":      version,
			})

			// Show the command template and resolved command
			fmt.Printf("   Template: %s\n", cmdStr)

			// Execute command and capture its actual output to show what it resolves to
			// Use sh -c to properly evaluate shell substitutions like $(lsb_release -rs)
			expandCmd := exec.Command("sh", "-c", fmt.Sprintf("echo %s", resolvedCmd))
			expandedOutput, _ := expandCmd.Output()
			if expandedOutput != nil && len(expandedOutput) > 0 {
				fmt.Printf("   Expanded: %s", string(expandedOutput))
			}

			fmt.Printf("   Executing: sh -c \"%s\"\n", resolvedCmd)
			fmt.Println("   Output:")

			cmd := exec.Command("sh", "-c", resolvedCmd)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("   ❌ Command failed with error: %v\n", err)
				return fmt.Errorf("repository setup step %d failed: %w", i+1, err)
			}
			fmt.Printf("   ✅ Step %d completed successfully\n", i+1)
		}
		fmt.Println("════════════════════════════════════════════════")
		fmt.Println("✅ Repository setup completed")
	}

	// Install packages
	if len(variant.Packages) == 0 {
		return fmt.Errorf("no packages specified for repository installation")
	}

	// Determine package manager based on distribution
	var installCmd []string
	fmt.Println("\n📦 Installing packages...")
	fmt.Println("════════════════════════════════════════════════")

	switch distro {
	case "ubuntu", "kubuntu", "debian", "mint", "elementary":
		// Update package list first
		fmt.Println("📋 Updating package lists...")
		fmt.Println("   Command: sudo apt-get update")
		updateCmd := exec.Command("sudo", "apt-get", "update")
		updateCmd.Stdout = os.Stdout
		updateCmd.Stderr = os.Stderr
		if err := updateCmd.Run(); err != nil {
			fmt.Printf("   ⚠️  Warning: apt update failed: %v\n", err)
		} else {
			fmt.Println("   ✅ Package lists updated")
		}

		installCmd = append([]string{"sudo", "apt-get", "install", "-y"}, variant.Packages...)
	case "fedora":
		installCmd = append([]string{"sudo", "dnf", "install", "-y"}, variant.Packages...)
	case "rocky":
		installCmd = append([]string{"sudo", "dnf", "install", "-y"}, variant.Packages...)
	case "arch":
		// Determine sudo prefix
		sudoPrefix := determineSudoPrefix()
		
		// Update package database first
		fmt.Println("📋 Updating package database...")
		if sudoPrefix == "" {
			fmt.Println("   Command: pacman -Sy")
			updateCmd := exec.Command("pacman", "-Sy")
			updateCmd.Stdout = os.Stdout
			updateCmd.Stderr = os.Stderr
			if err := updateCmd.Run(); err != nil {
				fmt.Printf("   ⚠️  Warning: pacman -Sy failed: %v\n", err)
			} else {
				fmt.Println("   ✅ Package database updated")
			}
			installCmd = append([]string{"pacman", "-S", "--noconfirm"}, variant.Packages...)
		} else {
			fmt.Println("   Command: sudo pacman -Sy")
			updateCmd := exec.Command("sudo", "pacman", "-Sy")
			updateCmd.Stdout = os.Stdout
			updateCmd.Stderr = os.Stderr
			if err := updateCmd.Run(); err != nil {
				fmt.Printf("   ⚠️  Warning: pacman -Sy failed: %v\n", err)
			} else {
				fmt.Println("   ✅ Package database updated")
			}
			installCmd = append([]string{"sudo", "pacman", "-S", "--noconfirm"}, variant.Packages...)
		}
	default:
		return fmt.Errorf("unsupported distribution for repository installation: %s", distro)
	}

	fmt.Printf("\n🎯 Installing packages: %s\n", strings.Join(variant.Packages, ", "))
	fmt.Printf("   Command: %s\n", strings.Join(installCmd, " "))
	fmt.Println("   Installation output:")
	fmt.Println("   " + strings.Repeat("-", 40))

	cmd := exec.Command(installCmd[0], installCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("   " + strings.Repeat("-", 40))
		fmt.Printf("   ❌ Package installation failed: %v\n", err)
		return fmt.Errorf("package installation failed: %w", err)
	}

	fmt.Println("   " + strings.Repeat("-", 40))
	fmt.Println("   ✅ Packages installed successfully")
	fmt.Println("════════════════════════════════════════════════")

	// Run post-install commands
	extraVariables := map[string]string{}
	if err := runPostInstallCommands(variant, extraVariables); err != nil {
		return fmt.Errorf("post-install commands failed: %w", err)
	}

	return nil
}


// installDirectDownload installs a package by directly downloading from URL
func installDirectDownload(platform *PlatformConfig, variant *VariantConfig) error {
	// Get URL from variant (should be in URL field for direct_download type)
	if variant.URL == "" {
		return fmt.Errorf("URL not specified for direct_download package")
	}

	// Get filename from URL
	filename := filepath.Base(variant.URL)
	if filename == "" || filename == "." {
		return fmt.Errorf("could not determine filename from URL: %s", variant.URL)
	}

	// Create cache directory
	cacheDir := ".cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cachedFile := filepath.Join(cacheDir, filename)

	// Download if not cached
	if _, err := os.Stat(cachedFile); os.IsNotExist(err) {
		fmt.Printf("Downloading %s...\n", filename)
		if err := downloadFile(cachedFile, variant.URL); err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}
		fmt.Printf("✅ Downloaded: %s\n", cachedFile)
	} else {
		fmt.Printf("Using cached file: %s\n", cachedFile)
	}

	// Handle extraction if needed
	if variant.Extract {
		extractDir := filepath.Join(cacheDir, "extracted")
		if err := os.MkdirAll(extractDir, 0755); err != nil {
			return fmt.Errorf("failed to create extraction directory: %w", err)
		}

		fmt.Printf("Extracting %s...\n", filename)
		if err := extractArchive(cachedFile, extractDir); err != nil {
			return fmt.Errorf("failed to extract: %w", err)
		}

		// Find the binary file
		binaryPath := filepath.Join(extractDir, variant.Binary)
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			return fmt.Errorf("binary '%s' not found in extracted files", variant.Binary)
		}

		// Install to target path
		targetPath := filepath.Join(variant.InstallPath, variant.Binary)
		
		// Create target directory
		if err := os.MkdirAll(variant.InstallPath, 0755); err != nil {
			return fmt.Errorf("failed to create install directory: %w", err)
		}

		// Copy binary to target (use sudo if required)
		if variant.RequiresSudo {
			fmt.Printf("Installing %s to %s (requires sudo)...\n", variant.Binary, targetPath)
			cmd := exec.Command("sudo", "cp", binaryPath, targetPath)
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to install binary: %w, output: %s", err, output)
			}
			
			// Make executable
			cmd = exec.Command("sudo", "chmod", "+x", targetPath)
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to make binary executable: %w, output: %s", err, output)
			}
		} else {
			fmt.Printf("Installing %s to %s...\n", variant.Binary, targetPath)
			if err := copyFile(binaryPath, targetPath); err != nil {
				return fmt.Errorf("failed to copy binary: %w", err)
			}
			
			// Make executable
			if err := os.Chmod(targetPath, 0755); err != nil {
				return fmt.Errorf("failed to make binary executable: %w", err)
			}
		}

		fmt.Printf("✅ Installed: %s\n", targetPath)
	}

	// Run post-install commands
	extraVariables := map[string]string{}
	if err := runPostInstallCommands(variant, extraVariables); err != nil {
		return fmt.Errorf("post-install commands failed: %w", err)
	}

	return nil
}


// extractArchive extracts an archive to target directory
func extractArchive(archivePath, targetDir string) error {
	var cmd *exec.Cmd
	
	if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
		cmd = exec.Command("tar", "-xzf", archivePath, "-C", targetDir)
	} else if strings.HasSuffix(archivePath, ".tar") {
		cmd = exec.Command("tar", "-xf", archivePath, "-C", targetDir)
	} else if strings.HasSuffix(archivePath, ".zip") {
		cmd = exec.Command("unzip", "-q", archivePath, "-d", targetDir)
	} else {
		return fmt.Errorf("unsupported archive format: %s", archivePath)
	}
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("extraction failed: %w, output: %s", err, output)
	}
	
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	cmd := exec.Command("cp", src, dst)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("copy failed: %w, output: %s", err, output)
	}
	return nil
}

// handlePrerequisites checks and installs prerequisite packages
func handlePrerequisites(config *InstallConfig, pkg *PackageConfig, dryRun bool) error {
	if len(pkg.Prerequisites) == 0 {
		return nil // No prerequisites
	}

	fmt.Printf("🔍 Checking prerequisites for %s...\n", pkg.Name)

	for _, prereq := range pkg.Prerequisites {
		fmt.Printf("   📋 Checking prerequisite: %s\n", prereq)

		// Check if prerequisite is already satisfied
		if isPackageInstalled(config, prereq) {
			fmt.Printf("   ✅ %s is already installed\n", prereq)
			continue
		}

		if dryRun {
			fmt.Printf("   🔄 [DRY-RUN] Would install prerequisite: %s\n", prereq)
			continue
		}

		// Install prerequisite
		fmt.Printf("   🔄 Installing prerequisite: %s\n", prereq)
		if err := InstallPackage(prereq, ""); err != nil {
			return fmt.Errorf("failed to install prerequisite '%s': %w", prereq, err)
		}
		fmt.Printf("   ✅ %s installed successfully\n", prereq)
	}

	if len(pkg.Prerequisites) > 0 {
		fmt.Println("🎉 All prerequisites satisfied!")
	}

	return nil
}

// isPackageInstalled checks if a package is already installed
func isPackageInstalled(config *InstallConfig, packageName string) bool {
	pkg, exists := config.Packages[packageName]
	if !exists {
		return false
	}

	// Get platform config for current OS
	currentOS := GetOperatingSystem()
	platform, exists := pkg.Platforms[currentOS]
	if !exists && currentOS == "windows_sandbox" {
		platform, exists = pkg.Platforms["windows"]
	}
	if !exists {
		return false
	}

	// Check verification command if available
	if platform.Verification.Command == "" {
		return false
	}

	cmd := exec.Command("cmd", "/C", platform.Verification.Command)
	if runtime.GOOS != "windows" {
		cmd = exec.Command("sh", "-c", platform.Verification.Command)
	}

	err := cmd.Run()
	expectedExitCode := platform.Verification.ExpectedExitCode
	if err == nil && expectedExitCode == 0 {
		return true
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			actualExitCode := exitErr.ExitCode()
			return actualExitCode == expectedExitCode
		}
	}

	return false
}

// downloadFile downloads a file from URL and saves it to the specified path
func downloadFile(filepath string, url string) error {
	fmt.Printf("Downloading %s...\n", url)
	
	// Create HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filepath, err)
	}
	defer out.Close()

	// Copy data from response to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file %s: %w", filepath, err)
	}

	fmt.Printf("✅ Downloaded: %s\n", filepath)
	return nil
}

// downloadFileWithProperFilename downloads a file and extracts proper filename from HTTP headers or redirects
func downloadFileWithProperFilename(url string, cacheDir string) (string, error) {
	fmt.Printf("Downloading %s...\n", url)

	// Create HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Extract filename from Content-Disposition header or final URL
	filename := extractFilenameFromResponse(resp)
	if filename == "" {
		// Fallback: generate filename based on content type
		filename = generateFilenameFromContentType(resp.Header.Get("Content-Type"))
	}

	filepath := filepath.Join(cacheDir, filename)

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", filepath, err)
	}
	defer out.Close()

	// Copy data from response to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file %s: %w", filepath, err)
	}

	fmt.Printf("✅ Downloaded: %s\n", filepath)
	return filename, nil
}

// extractFilenameFromResponse extracts filename from HTTP response headers or URL
func extractFilenameFromResponse(resp *http.Response) string {
	// Method 1: Content-Disposition header
	disposition := resp.Header.Get("Content-Disposition")
	if filename := parseContentDisposition(disposition); filename != "" {
		return filename
	}

	// Method 2: Final URL path after redirects
	if resp.Request != nil && resp.Request.URL != nil {
		path := resp.Request.URL.Path
		if filename := filepath.Base(path); filename != "" && filename != "." && filename != "/" {
			// Remove query parameters if any
			if idx := strings.Index(filename, "?"); idx >= 0 {
				filename = filename[:idx]
			}
			// Only return if it has a proper extension
			if strings.Contains(filename, ".") {
				return filename
			}
		}
	}

	return ""
}

// parseContentDisposition parses Content-Disposition header for filename
func parseContentDisposition(disposition string) string {
	if disposition == "" || !strings.Contains(disposition, "filename=") {
		return ""
	}

	parts := strings.Split(disposition, "filename=")
	if len(parts) < 2 {
		return ""
	}

	filename := strings.TrimSpace(parts[1])
	filename = strings.Trim(filename, `"`)
	filename = strings.Trim(filename, `'`)

	return filename
}

// generateFilenameFromContentType generates filename based on content type
func generateFilenameFromContentType(contentType string) string {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	switch {
	case strings.Contains(contentType, "application/octet-stream"),
		strings.Contains(contentType, "application/x-msdownload"),
		strings.Contains(contentType, "application/vnd.microsoft.portable-executable"):
		return fmt.Sprintf("download_%s.exe", timestamp)
	case strings.Contains(contentType, "application/zip"):
		return fmt.Sprintf("download_%s.zip", timestamp)
	case strings.Contains(contentType, "application/x-msi"),
		strings.Contains(contentType, "application/x-msdos-program"):
		return fmt.Sprintf("download_%s.msi", timestamp)
	default:
		return fmt.Sprintf("download_%s.exe", timestamp) // Default to .exe for Windows
	}
}

// determineSudoPrefix detects if sudo is needed and available
func determineSudoPrefix() string {
	// Check if running as root
	if isRunningAsRoot() {
		return "" // No sudo needed
	}
	
	// Check if sudo is available
	if isSudoAvailable() {
		return "sudo " // Use sudo with space
	}
	
	// No sudo available - commands may fail, but try without
	fmt.Printf("⚠️  Warning: Not running as root and sudo not available. Some operations may fail.\n")
	return ""
}

// isRunningAsRoot checks if the current process is running as root (UID 0)
func isRunningAsRoot() bool {
	if runtime.GOOS == "windows" {
		return false // Windows doesn't use UID 0 concept
	}
	
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	uid := strings.TrimSpace(string(output))
	return uid == "0"
}

// isSudoAvailable checks if sudo command is available in the system
func isSudoAvailable() bool {
	if runtime.GOOS == "windows" {
		return false // Windows doesn't use sudo
	}

	cmd := exec.Command("which", "sudo")
	return cmd.Run() == nil
}

// checkDirectoryPermissions checks if a directory requires elevated permissions and provides fallback path
func checkDirectoryPermissions(destDir string) (requiresSudo bool, fallbackDir string) {
	if runtime.GOOS == "windows" {
		return false, destDir // Windows doesn't use sudo concept
	}

	// Determine user home directory for fallback
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp" // Last resort fallback
	}
	fallbackDir = filepath.Join(homeDir, ".local", "bin")

	// System directories that typically require elevated permissions
	systemDirs := []string{
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
		"/usr/local",
		"/usr",
		"/opt",
	}

	// Check if the destination is a system directory
	for _, sysDir := range systemDirs {
		if strings.HasPrefix(destDir, sysDir) {
			// Check if we can write to this directory
			if !canWriteToDirectory(destDir) {
				return true, fallbackDir
			}
		}
	}

	// For non-system directories, still check write permissions
	if !canWriteToDirectory(destDir) {
		return true, fallbackDir
	}

	return false, destDir
}

// canWriteToDirectory checks if the current user can write to the specified directory
func canWriteToDirectory(dir string) bool {
	// Create the directory if it doesn't exist (without sudo)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false
	}

	// Try to create a test file
	testFile := filepath.Join(dir, ".portunix_write_test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()

	// Clean up test file
	os.Remove(testFile)
	return true
}

// addUserBinToPath adds the user bin directory to PATH in shell configuration files
func addUserBinToPath(binDir string) error {
	if runtime.GOOS == "windows" {
		return nil // Windows PATH handling is different and handled elsewhere
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Common shell configuration files to update
	shellConfigs := []string{
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".profile"),
	}

	pathExport := fmt.Sprintf("export PATH=\"%s:$PATH\"", binDir)
	pathExportComment := "# Added by Portunix for user-local binaries"

	for _, configFile := range shellConfigs {
		// Check if file exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			continue // Skip non-existent files
		}

		// Read current content
		content, err := os.ReadFile(configFile)
		if err != nil {
			continue // Skip files we can't read
		}

		// Check if PATH export already exists
		contentStr := string(content)
		if strings.Contains(contentStr, binDir) {
			continue // Already configured
		}

		// Append PATH export
		f, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			continue // Skip files we can't write to
		}

		_, err = fmt.Fprintf(f, "\n%s\n%s\n", pathExportComment, pathExport)
		f.Close()

		if err != nil {
			continue // Skip if write failed
		}

		fmt.Printf("📝 Updated %s with PATH export\n", configFile)
	}

	return nil
}

// isUserDirectory checks if a directory is in user space (not system-wide)
func isUserDirectory(dir string) bool {
	if runtime.GOOS == "windows" {
		// On Windows, check if it's in user profile
		userProfile := os.Getenv("USERPROFILE")
		return userProfile != "" && strings.HasPrefix(dir, userProfile)
	}

	// On Unix systems, check if it's in home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Common user directories
	userDirs := []string{
		homeDir,
		"/tmp",
		"/var/tmp",
	}

	for _, userDir := range userDirs {
		if strings.HasPrefix(dir, userDir) {
			return true
		}
	}

	return false
}

// runPostInstallCommands executes post-install commands with proper variable resolution
func runPostInstallCommands(variant *VariantConfig, extraVariables map[string]string) error {
	if len(variant.PostInstall) == 0 {
		return nil
	}

	fmt.Println("⚙️  Running post-install commands...")

	// Prepare variables for template resolution
	variables := make(map[string]string)
	
	// Add sudo prefix based on current execution context
	// If actual_extract_to is in user directory, don't use sudo
	actualExtractTo, hasActualExtractTo := extraVariables["actual_extract_to"]
	if hasActualExtractTo && isUserDirectory(actualExtractTo) {
		variables["sudo_prefix"] = "" // No sudo needed for user directory
	} else {
		variables["sudo_prefix"] = determineSudoPrefix()
	}
	
	// Add any extra variables provided by caller
	for key, value := range extraVariables {
		variables[key] = value
	}

	// Execute each post-install command
	for i, cmdTemplate := range variant.PostInstall {
		// Resolve template variables
		resolvedCmd := ResolveVariables(cmdTemplate, variables)
		
		fmt.Printf("   📌 Step %d/%d: %s\n", i+1, len(variant.PostInstall))
		fmt.Printf("   Template: %s\n", cmdTemplate)
		if resolvedCmd != cmdTemplate {
			fmt.Printf("   Resolved: %s\n", resolvedCmd)
		}

		// Execute command
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", resolvedCmd)
		} else {
			cmd = exec.Command("sh", "-c", resolvedCmd)
		}
		
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			fmt.Printf("   ❌ Command failed: %v\n", err)
			return fmt.Errorf("post-install command %d failed: %w", i+1, err)
		}
		
		fmt.Printf("   ✅ Step %d completed successfully\n", i+1)
	}

	fmt.Println("✅ All post-install commands completed successfully")
	return nil
}

// installRedirect handles redirect type installations by redirecting to another package
func installRedirect(platform *PlatformConfig, variant *VariantConfig) error {
	if variant.RedirectTo == "" {
		return fmt.Errorf("redirect_to field is required for redirect type")
	}

	fmt.Printf("🔀 Redirecting to package: %s\n", variant.RedirectTo)

	// Get the target variant to use
	targetVariant := variant.DefaultVariant
	if targetVariant == "" {
		targetVariant = "default"
	}

	fmt.Printf("📋 Using variant: %s\n", targetVariant)

	// Install the target package
	return InstallPackage(variant.RedirectTo, targetVariant)
}

// installPip installs packages using pip
func installPip(platform *PlatformConfig, variant *VariantConfig) error {
	pipMgr := pip.NewPipManager()

	// Set dry-run mode if we're in dry-run mode
	// (This needs to be passed from the main function - for now defaulting to false)
	pipMgr.DryRun = false // TODO: Pass dry-run flag from main installer

	// Check if pip is supported
	if !pipMgr.IsSupported() {
		return fmt.Errorf("pip is not available on this system. Please install Python and pip first")
	}

	packages := variant.Packages
	if len(packages) == 0 {
		return fmt.Errorf("no packages specified for pip installation")
	}

	fmt.Printf("🐍 Installing Python packages with pip...\n")

	// Install packages using pip module
	if err := pipMgr.Install(packages); err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	// Run post-install commands
	extraVariables := map[string]string{}
	if err := runPostInstallCommands(variant, extraVariables); err != nil {
		return fmt.Errorf("post-install commands failed: %w", err)
	}

	return nil
}

// installPipx installs packages using pipx (isolated Python environments)
func installPipx(platform *PlatformConfig, variant *VariantConfig) error {
	// Check if pipx is available
	if !isPipxAvailable() {
		return fmt.Errorf("pipx is not available on this system. Please install pipx first")
	}

	packages := variant.Packages
	if len(packages) == 0 {
		return fmt.Errorf("no packages specified for pipx installation")
	}

	fmt.Printf("📦 Installing Python applications with pipx in isolated environments...\n")

	// Install packages using pipx
	for _, pkg := range packages {
		if err := installPipxPackage(pkg); err != nil {
			return fmt.Errorf("failed to install package '%s' with pipx: %w", pkg, err)
		}
	}

	// Run post-install commands
	extraVariables := map[string]string{}
	if err := runPostInstallCommands(variant, extraVariables); err != nil {
		return fmt.Errorf("post-install commands failed: %w", err)
	}

	return nil
}

// isPipxAvailable checks if pipx is available on the system
func isPipxAvailable() bool {
	cmd := exec.Command("pipx", "--version")
	return cmd.Run() == nil
}

// installPipxPackage installs a single package using pipx
func installPipxPackage(pkg string) error {
	fmt.Printf("📦 Installing %s with pipx...\n", pkg)

	cmd := exec.Command("pipx", "install", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pipx install failed for '%s': %w", pkg, err)
	}

	fmt.Printf("✅ Successfully installed %s with pipx\n", pkg)
	return nil
}

// installAnsibleGalaxy installs Ansible collections using ansible-galaxy
func installAnsibleGalaxy(platform *PlatformConfig, variant *VariantConfig) error {
	galaxyMgr := ansible_galaxy.NewAnsibleGalaxyInstaller()

	// Set dry-run mode if we're in dry-run mode
	// (This needs to be passed from the main function - for now defaulting to false)
	galaxyMgr.DryRun = false // TODO: Pass dry-run flag from main installer

	// Check if ansible-galaxy is supported
	if !galaxyMgr.IsSupported() {
		return fmt.Errorf("ansible-galaxy is not available on this system. Please install Ansible first")
	}

	collections := variant.Packages
	if len(collections) == 0 {
		return fmt.Errorf("no collections specified for ansible-galaxy installation")
	}

	fmt.Printf("🎭 Installing Ansible collections with ansible-galaxy...\n")

	// Install collections using ansible-galaxy module
	if err := galaxyMgr.Install(collections); err != nil {
		return fmt.Errorf("failed to install collections: %w", err)
	}

	// Run post-install commands
	extraVariables := map[string]string{}
	if err := runPostInstallCommands(variant, extraVariables); err != nil {
		return fmt.Errorf("post-install commands failed: %w", err)
	}

	return nil
}

// resolveVariantFromRegistry determines the effective variant based on method and version overrides for new registry
func resolveVariantFromRegistry(pkg *Package, options *InstallOptions) (string, error) {
	// Get current OS platform
	currentOS := GetOperatingSystem()
	platform, exists := pkg.Spec.Platforms[currentOS]
	if !exists {
		if currentOS == "windows_sandbox" {
			platform, exists = pkg.Spec.Platforms["windows"]
		}
		if !exists {
			return "", fmt.Errorf("no platform configuration for %s", currentOS)
		}
	}

	// If method is specified, use it as variant (method override)
	if options.Method != "" {
		// Verify the method exists
		if _, exists := platform.Variants[options.Method]; !exists {
			return "", fmt.Errorf("method '%s' not found for package '%s' on platform '%s'", options.Method, options.PackageName, currentOS)
		}
		return options.Method, nil
	}

	// Otherwise use explicit variant
	if options.Variant != "" {
		// Verify the variant exists
		if _, exists := platform.Variants[options.Variant]; !exists {
			return "", fmt.Errorf("variant '%s' not found for package '%s' on platform '%s'", options.Variant, options.PackageName, currentOS)
		}
		return options.Variant, nil
	}

	// No override, use smart variant selection
	return findBestVariantForCurrentSystem(platform, currentOS)
}

// findBestVariantForCurrentSystem finds the best variant for the current system
func findBestVariantForCurrentSystem(platform PlatformSpec, currentOS string) (string, error) {
	// For Linux, try to detect distribution and find matching variant
	if currentOS == "linux" {
		return findBestVariantForLinux(platform)
	}

	// For Windows, prefer user installs over system installs for non-admin scenarios
	if currentOS == "windows" || currentOS == "windows_sandbox" {
		return findBestVariantForWindows(platform)
	}

	// For other platforms, just return first available variant
	for variantName := range platform.Variants {
		return variantName, nil
	}

	return "", fmt.Errorf("no variants available")
}

// findBestVariantForLinux finds the best variant for Linux based on distribution
func findBestVariantForLinux(platform PlatformSpec) (string, error) {
	// Check if this looks like a version-based package (Java, etc.)
	if variantName := findDefaultLTSVersion(platform); variantName != "" {
		return variantName, nil
	}

	// Try to detect Linux distribution
	distro, _, err := GetLinuxDistribution()
	if err != nil {
		// Fallback to snap if available, otherwise first variant
		if _, exists := platform.Variants["snap"]; exists {
			return "snap", nil
		}
		for variantName := range platform.Variants {
			return variantName, nil
		}
		return "", fmt.Errorf("no variants available")
	}

	// Priority mapping for different distributions
	distributionPreferences := map[string][]string{
		"ubuntu":      {"apt", "snap", "deb"},
		"debian":      {"apt", "deb", "snap"},
		"mint":        {"apt", "snap", "deb"},
		"elementary":  {"apt", "snap", "deb"},
		"fedora":      {"dnf", "rpm", "snap"},
		"rocky":       {"dnf", "rpm", "yum", "snap"},
		"almalinux":   {"dnf", "rpm", "yum", "snap"},
		"centos":      {"dnf", "rpm", "yum", "snap"},
		"arch":        {"pacman", "snap"},
	}

	// Get preferred variants for current distribution
	preferredVariants, exists := distributionPreferences[distro]
	if !exists {
		// Unknown distribution, fallback to universal variants
		preferredVariants = []string{"snap", "apt", "dnf", "pacman"}
	}

	// Try each preferred variant in order
	for _, variant := range preferredVariants {
		if _, exists := platform.Variants[variant]; exists {
			// Additional check: verify this variant supports current distribution
			if variantSupportsDistribution(platform.Variants[variant], distro) {
				return variant, nil
			}
		}
	}

	// No preferred variant found, try any available variant
	for variantName, variant := range platform.Variants {
		if variantSupportsDistribution(variant, distro) {
			return variantName, nil
		}
	}

	// Last resort: return any variant
	for variantName := range platform.Variants {
		return variantName, nil
	}

	return "", fmt.Errorf("no suitable variant found for distribution %s", distro)
}

// findBestVariantForWindows finds the best variant for Windows
func findBestVariantForWindows(platform PlatformSpec) (string, error) {
	// Check if this looks like a version-based package (Java, etc.)
	if variantName := findDefaultLTSVersion(platform); variantName != "" {
		return variantName, nil
	}

	// Prefer user variants over system variants (no admin required)
	preferredOrder := []string{"user", "stable", "latest", "system"}

	for _, variant := range preferredOrder {
		if _, exists := platform.Variants[variant]; exists {
			return variant, nil
		}
	}

	// Fallback to first available variant
	for variantName := range platform.Variants {
		return variantName, nil
	}

	return "", fmt.Errorf("no variants available for Windows")
}

// findDefaultLTSVersion finds the default LTS version for packages that use version numbers as variants
func findDefaultLTSVersion(platform PlatformSpec) string {
	// For Java-like packages, prefer LTS versions in order: 21, 17, 11, 8
	ltsVersions := []string{"21", "17", "11", "8"}

	for _, version := range ltsVersions {
		if _, exists := platform.Variants[version]; exists {
			return version
		}
	}

	// Check if all variants are numeric (indicating version-based package)
	hasOnlyNumericVariants := true
	for variantName := range platform.Variants {
		if !isNumericVersion(variantName) {
			hasOnlyNumericVariants = false
			break
		}
	}

	if hasOnlyNumericVariants {
		// Find the highest numeric version
		return findHighestNumericVersion(platform)
	}

	return "" // Not a version-based package
}

// isNumericVersion checks if a variant name is a numeric version
func isNumericVersion(variant string) bool {
	// Check if variant is a number (like "8", "11", "17", "21")
	for _, char := range variant {
		if char < '0' || char > '9' {
			return false
		}
	}
	return len(variant) > 0
}

// findHighestNumericVersion finds the highest numeric version variant
func findHighestNumericVersion(platform PlatformSpec) string {
	highestVersion := 0
	highestVariant := ""

	for variantName := range platform.Variants {
		if isNumericVersion(variantName) {
			if version := parseSimpleInt(variantName); version > highestVersion {
				highestVersion = version
				highestVariant = variantName
			}
		}
	}

	return highestVariant
}

// parseSimpleInt parses a simple integer string
func parseSimpleInt(s string) int {
	result := 0
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		}
	}
	return result
}

// variantSupportsDistribution checks if a variant supports the given distribution
func variantSupportsDistribution(variant VariantSpec, distro string) bool {
	if variant.Distributions == nil {
		return true // No restriction means it supports all distributions
	}

	// Handle the different distribution formats ([]string or map[string]interface{})
	switch distributions := variant.Distributions.(type) {
	case []interface{}:
		for _, dist := range distributions {
			if distStr, ok := dist.(string); ok {
				if distStr == distro || distStr == "universal" {
					return true
				}
			}
		}
	case map[string]interface{}:
		for dist := range distributions {
			if dist == distro || dist == "universal" {
				return true
			}
		}
	case []string:
		for _, dist := range distributions {
			if dist == distro || dist == "universal" {
				return true
			}
		}
	}

	return false
}

// InstallPackageWithDryRunLegacy installs a package using the legacy config with dry-run support
func InstallPackageWithDryRunLegacy(config *InstallConfig, packageName, variant string, dryRun bool) error {
	// This function delegates to the existing InstallPackageWithDryRun function
	// We maintain compatibility by using the converted legacy config
	return InstallPackageWithDryRun(packageName, variant, dryRun)
}

// installPackageFromRegistry installs a package directly from registry format
func installPackageFromRegistry(pkg *Package, variantName string, dryRun bool) error {
	// Get current OS platform
	currentOS := GetOperatingSystem()
	platform, exists := pkg.Spec.Platforms[currentOS]
	if !exists {
		if currentOS == "windows_sandbox" {
			platform, exists = pkg.Spec.Platforms["windows"]
		}
		if !exists {
			return fmt.Errorf("no platform configuration for %s", currentOS)
		}
	}

	// Get the specific variant
	variant, exists := platform.Variants[variantName]
	if !exists {
		return fmt.Errorf("variant '%s' not found for package '%s' on platform '%s'", variantName, pkg.Metadata.Name, currentOS)
	}

	// Convert to legacy format for the actual installation
	legacyPkg := &PackageConfig{
		Name:        pkg.Metadata.Name,
		Description: pkg.Metadata.Description,
		Platforms:   make(map[string]PlatformConfig),
	}

	// Convert the specific platform
	legacyPlatform := PlatformConfig{
		Type:        platform.Type,
		Variants:    make(map[string]VariantConfig),
		InstallArgs: platform.InstallArgs,
		Environment: platform.Environment,
	}

	// Convert verification
	if platform.Verification != nil {
		legacyPlatform.Verification = VerificationConfig{
			Command:          platform.Verification.Command,
			ExpectedExitCode: platform.Verification.ExpectedExitCode,
		}
	}

	// Convert the specific variant
	// Auto-detect type for package manager variants (apt, dnf, snap, etc.)
	variantType := variant.Type
	if variantType == "" && len(variant.Packages) > 0 {
		// If variant has packages but no type, use variant name as type
		// This handles package manager variants like apt, snap, dnf, etc.
		variantType = variantName
	}

	legacyVariant := VariantConfig{
		Version:       variant.Version,
		Type:          variantType,
		URL:           variant.URL,
		URLs:          variant.URLs,
		Packages:      variant.Packages,
		InstallScript: variant.InstallScript,
		InstallPath:   variant.InstallPath,
		ExtractTo:     variant.ExtractTo,
		Extract:       variant.Extract,
		Binary:        variant.Binary,
		RequiresSudo:  variant.RequiresSudo,
		PostInstall:   variant.PostInstall,
		InstallArgs:   variant.InstallArgs,
		Distributions: variant.Distributions,
	}

	legacyPlatform.Variants[variantName] = legacyVariant
	legacyPkg.Platforms[currentOS] = legacyPlatform

	// Create a minimal config with just this package
	config := &InstallConfig{
		Version:  "1.0",
		Packages: map[string]PackageConfig{pkg.Metadata.Name: *legacyPkg},
		Presets:  make(map[string]PresetConfig),
	}

	// Use the existing installation logic with the converted config
	return installPackageWithSpecificConfig(config, pkg.Metadata.Name, variantName, dryRun)
}

// installPackageWithSpecificConfig installs a package using a specific config
func installPackageWithSpecificConfig(config *InstallConfig, packageName, variant string, dryRun bool) error {
	// This mimics the logic from InstallPackageWithDryRun but with a specific config
	pkg, platform, variantConfig, err := config.GetPackageInfo(packageName, variant)
	if err != nil {
		return err
	}

	// Print installation header
	if dryRun {
		// Determine actual install type
		installType := platform.Type
		if variantConfig.Type != "" {
			installType = variantConfig.Type
		}

		fmt.Println("════════════════════════════════════════════════")
		fmt.Printf("📦 INSTALLING: %s\n", pkg.Name)
		fmt.Println("════════════════════════════════════════════════")
		fmt.Printf("📄 Description: %s\n", pkg.Description)
		fmt.Printf("🔧 Variant: %s (v%s)\n", variant, variantConfig.Version)
		fmt.Printf("💻 Platform: %s\n", GetOperatingSystem())
		fmt.Printf("🏗️  Installation type: %s\n", installType)

		// Show package-specific information
		if len(variantConfig.Packages) > 0 {
			fmt.Printf("📋 Packages: %s\n", strings.Join(variantConfig.Packages, ", "))
		}
		if variantConfig.URL != "" {
			fmt.Printf("🌐 Download URL: %s\n", variantConfig.URL)
		}
		if len(variantConfig.URLs) > 0 {
			url, _ := variantConfig.GetDownloadURL()
			fmt.Printf("🌐 Download URL: %s\n", url)
		}
		if variantConfig.ExtractTo != "" {
			fmt.Printf("📁 Extract to: %s\n", variantConfig.ExtractTo)
		}

		fmt.Println("════════════════════════════════════════════════")
		fmt.Println("🔍 DRY-RUN MODE: Showing what would be installed")
		fmt.Println("💡 To execute for real, remove the --dry-run flag")
		fmt.Println("════════════════════════════════════════════════")
		return nil
	}

	// For real installation, use the same logic as InstallPackageWithDryRun
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

	// Determine actual install type
	installType := platform.Type
	if variantConfig.Type != "" {
		installType = variantConfig.Type
	}

	var installErr error
	// Use variant type if specified, otherwise use platform type
	switch installType {
	case "msi", "exe":
		installErr = installWindowsBinary(platform, variantConfig)
	case "zip", "tar.gz", "tar.xz":
		installErr = installArchive(platform, variantConfig)
	case "apt":
		installErr = installApt(platform, variantConfig)
	case "dnf", "yum":
		installErr = installDnf(platform, variantConfig)
	case "pacman":
		installErr = installPacman(platform, variantConfig)
	case "deb":
		installErr = installDeb(platform, variantConfig)
	case "snap":
		installErr = installSnap(platform, variantConfig)
	case "pip":
		installErr = installPip(platform, variantConfig)
	case "pipx":
		installErr = installPipx(platform, variantConfig)
	case "ansible-galaxy":
		installErr = installAnsibleGalaxy(platform, variantConfig)
	case "powershell":
		installErr = installPowerShell(platform, variantConfig)
	case "repository":
		installErr = installRepository(platform, variantConfig)
	case "direct_download":
		installErr = installDirectDownload(platform, variantConfig)
	case "redirect":
		installErr = installRedirect(platform, variantConfig)
	default:
		return fmt.Errorf("unsupported package type: %s", installType)
	}

	if installErr != nil {
		return fmt.Errorf("failed to install %s: %w", packageName, installErr)
	}

	// Verify installation if verification command is provided
	if platform.Verification.Command != "" {
		fmt.Printf("✅ Verifying %s installation...\n", packageName)
		cmd := exec.Command("cmd", "/C", platform.Verification.Command)
		if runtime.GOOS != "windows" {
			cmd = exec.Command("sh", "-c", platform.Verification.Command)
		}

		err := cmd.Run()
		expectedExitCode := platform.Verification.ExpectedExitCode
		if err == nil && expectedExitCode == 0 {
			fmt.Printf("✅ %s installed and verified successfully!\n", pkg.Name)
		} else if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				actualExitCode := exitErr.ExitCode()
				if actualExitCode == expectedExitCode {
					fmt.Printf("✅ %s installed and verified successfully!\n", pkg.Name)
				} else {
					fmt.Printf("⚠️  %s installation completed but verification failed\n", pkg.Name)
				}
			}
		}
	} else {
		fmt.Printf("✅ %s installation completed!\n", pkg.Name)
	}

	fmt.Println("════════════════════════════════════════════════")
	return nil
}
