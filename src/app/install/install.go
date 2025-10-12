package install

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"portunix.ai/app"
	"portunix.ai/app/docker"
	"portunix.ai/app/podman"
)

// InstallOptions holds parsed install command options
type InstallOptions struct {
	PackageName  string
	Method       string
	Version      string
	DryRun       bool
	ListMethods  bool
	Variant      string
	AutoAccept   bool
	GUI          bool
	Embeddable   bool
	OtherArgs    []string
}

func ToArguments(what string) []string {
	arguments := make([]string, 1)
	arguments[0] = what
	return arguments
}

func Install(arguments []string) {
	if len(arguments) == 0 {
		return
	}

	// Parse new flags
	installOptions := parseInstallFlags(arguments)

	// Handle special commands
	if installOptions.ListMethods {
		if installOptions.PackageName == "" {
			fmt.Println("Error: --list-methods requires a package name")
			fmt.Println("Usage: portunix install <package> --list-methods")
			return
		}
		showPackageMethods(installOptions.PackageName)
		return
	}

	// Special handling for Docker and Podman installation
	if installOptions.PackageName == "docker" {
		if err := docker.InstallDocker(installOptions.AutoAccept); err != nil {
			return
		}
		return
	}

	if installOptions.PackageName == "podman" {
		if err := podman.InstallPodman(installOptions.AutoAccept); err != nil {
			return
		}
		return
	}

	// Check if it's a preset first
	if err := InstallPresetWithDryRun(installOptions.PackageName, installOptions.DryRun); err == nil {
		return // Success with preset
	}

	// Try to install using new system with method override
	if err := InstallPackageWithOptions(installOptions); err == nil {
		return // Success with new system
	} else {
		fmt.Printf("Error installing package '%s': %v\n", installOptions.PackageName, err)
		os.Exit(1) // Exit with error code
	}

	// Fall back to old system
	os := runtime.GOOS
	if os == "linux" {
		InstallLnx(arguments)
	} else if os == "windows" {
		InstallWin(arguments)
	} else {
		//TODO:
	}
}

func ProcessArgumentsInstall(arguments []string) (map[string]string, []string) {
	//TODO: use list
	enabledArguments := []string{"version", "variant"}
	return app.ProcessArguments(arguments, enabledArguments)
}

func ProcessArgumentsInstallJava(arguments []string) map[string]string {
	argsMap, other := ProcessArgumentsInstall(arguments)
	// Check if the first 'other' argument is a version
	if len(other) > 0 {
		if _, err := strconv.Atoi(other[0]); err == nil {
			argsMap["version"] = other[0]
			other = other[1:] // Remove the version from 'other'
		}
	}
	for _, str := range other {
		if str == "openjdk" {
			argsMap["variant"] = str
		}
	}
	return argsMap
}

// ShowPackageHelp displays specific help information for a package
func ShowPackageHelp(packageName string) error {
	switch packageName {
	case "powershell":
		return showPowerShellHelp()
	case "java":
		return showJavaHelp()
	case "python":
		return showPythonHelp()
	case "vscode":
		return showVSCodeHelp()
	case "docker":
		return showDockerHelp()
	case "podman":
		return showPodmanHelp()
	case "chrome":
		return showChromeHelp()
	default:
		return fmt.Errorf("package '%s' does not have specific help available", packageName)
	}
}

// InstallPreset installs a preset collection of packages
func InstallPreset(presetName string) error {
	return InstallPresetWithDryRun(presetName, false)
}

// InstallPresetWithDryRun installs a preset collection of packages with dry-run support
func InstallPresetWithDryRun(presetName string, dryRun bool) error {
	config, err := LoadInstallConfig()
	if err != nil {
		return fmt.Errorf("failed to load install config: %w", err)
	}

	preset, exists := config.Presets[presetName]
	if !exists {
		return fmt.Errorf("preset '%s' not found", presetName)
	}

	fmt.Printf("════════════════════════════════════════════════\n")
	if dryRun {
		fmt.Printf("🔍 DRY-RUN: PRESET WOULD BE INSTALLED: %s\n", preset.Name)
	} else {
		fmt.Printf("📦 INSTALLING PRESET: %s\n", preset.Name)
	}
	fmt.Printf("════════════════════════════════════════════════\n")
	fmt.Printf("📄 Description: %s\n", preset.Description)
	fmt.Printf("📦 Packages: %d\n", len(preset.Packages))
	fmt.Printf("════════════════════════════════════════════════\n")

	if dryRun {
		fmt.Println("🔍 Packages that would be installed:")
		for i, pkg := range preset.Packages {
			fmt.Printf("  %d. %s (variant: %s)\n", i+1, pkg.Name, pkg.Variant)
		}
		fmt.Printf("💡 To execute for real, remove the --dry-run flag\n")
		return nil
	}

	for i, pkg := range preset.Packages {
		fmt.Printf("Installing package %d/%d: %s (variant: %s)\n", i+1, len(preset.Packages), pkg.Name, pkg.Variant)
		if err := InstallPackageWithDryRun(pkg.Name, pkg.Variant, dryRun); err != nil {
			fmt.Printf("❌ Failed to install %s: %v\n", pkg.Name, err)
			return fmt.Errorf("preset installation failed at package %s: %w", pkg.Name, err)
		}
		fmt.Printf("✅ Successfully installed %s\n", pkg.Name)
	}

	fmt.Printf("════════════════════════════════════════════════\n")
	fmt.Printf("✅ PRESET INSTALLATION COMPLETED: %s\n", preset.Name)
	fmt.Printf("════════════════════════════════════════════════\n")

	return nil
}

// showPowerShellHelp displays PowerShell-specific installation help
func showPowerShellHelp() error {
	fmt.Println("PowerShell Installation Help")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  portunix install powershell [flags]")
	fmt.Println()
	fmt.Println("Description:")
	fmt.Println("  Install Microsoft PowerShell on Linux distributions")
	fmt.Println()

	if runtime.GOOS == "linux" {
		fmt.Println("Available variants for Linux:")
		fmt.Println("  ubuntu      - Install via Microsoft APT repository (Ubuntu/Kubuntu)")
		fmt.Println("  debian      - Install via Microsoft APT repository (Debian)")
		fmt.Println("  fedora      - Install via Microsoft DNF repository (Fedora)")
		fmt.Println("  rocky       - Install via Microsoft DNF repository (Rocky Linux)")
		fmt.Println("  mint        - Install via Microsoft APT repository (Linux Mint)")
		fmt.Println("  elementary  - Install via Microsoft APT repository (Elementary OS)")
		fmt.Println("  snap        - Install via Snap package (universal fallback)")
		fmt.Println()
		fmt.Println("Auto-detection:")
		fmt.Println("  If no variant is specified, Portunix will automatically detect")
		fmt.Println("  your Linux distribution and choose the best installation method.")
	} else {
		fmt.Println("Available variants for Windows:")
		fmt.Println("  latest      - Install latest PowerShell version (default)")
	}

	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix install powershell")
	fmt.Println("  portunix install powershell --variant ubuntu")
	fmt.Println("  portunix install powershell --variant snap")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --variant string   Specify installation variant")
	fmt.Println("  -h, --help         help for powershell")

	return nil
}

// showJavaHelp displays Java-specific installation help
func showJavaHelp() error {
	fmt.Println("Java Installation Help")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  portunix install java [version] [flags]")
	fmt.Println()
	fmt.Println("Available versions:")
	fmt.Println("  8, 11, 17, 21 (default: 21)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix install java")
	fmt.Println("  portunix install java --variant 17")
	fmt.Println("  portunix install java 11")

	return nil
}

// showPythonHelp displays Python-specific installation help
func showPythonHelp() error {
	fmt.Println("Python Installation Help")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  portunix install python [flags]")
	fmt.Println()
	fmt.Println("Available variants:")
	fmt.Println("  embeddable  - Install embeddable Python (portable ZIP version)")
	fmt.Println("  full        - Install full Python with all components")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --gui         Use GUI installer instead of silent installation")
	fmt.Println("  --embeddable  Install embeddable Python (portable ZIP version)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix install python")
	fmt.Println("  portunix install python --embeddable")
	fmt.Println("  portunix install python --gui")

	return nil
}

// showVSCodeHelp displays VS Code-specific installation help
func showVSCodeHelp() error {
	fmt.Println("Visual Studio Code Installation Help")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  portunix install vscode [flags]")
	fmt.Println()
	fmt.Println("Available variants:")
	fmt.Println("  user    - Install for current user (default)")
	fmt.Println("  system  - Install system-wide")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix install vscode")
	fmt.Println("  portunix install vscode --variant system")

	return nil
}

// showDockerHelp displays Docker-specific installation help
func showDockerHelp() error {
	fmt.Println("Docker Installation Help")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  portunix install docker [flags]")
	fmt.Println()
	fmt.Println("Description:")
	fmt.Println("  Install Docker Engine/Desktop with intelligent OS detection")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -y    Auto-accept installation prompts")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix install docker")
	fmt.Println("  portunix install docker -y")

	return nil
}

// showPodmanHelp displays Podman-specific installation help
func showPodmanHelp() error {
	fmt.Println("Podman Installation Help")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  portunix install podman [flags]")
	fmt.Println()
	fmt.Println("Description:")
	fmt.Println("  Install Podman container engine with rootless support")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -y    Auto-accept installation prompts")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix install podman")
	fmt.Println("  portunix install podman -y")

	return nil
}

// showChromeHelp displays Chrome-specific installation help
func showChromeHelp() error {
	fmt.Println("Google Chrome Installation Help")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  portunix install chrome [flags]")
	fmt.Println()
	fmt.Println("Description:")
	fmt.Println("  Install Google Chrome web browser with automatic repository configuration")
	fmt.Println()
	
	if runtime.GOOS == "linux" {
		fmt.Println("Available variants for Linux:")
		fmt.Println("  ubuntu      - Official APT repository (Ubuntu/Kubuntu/Lubuntu/Xubuntu)")
		fmt.Println("  debian      - Official APT repository (Debian 11/12)")
		fmt.Println("  mint        - Official APT repository (Linux Mint 21.x)")
		fmt.Println("  elementary  - Official APT repository (Elementary OS 7.x)")
		fmt.Println("  fedora      - Official DNF repository (Fedora 38/39/40)")
		fmt.Println("  rocky       - Official DNF repository (Rocky/AlmaLinux/CentOS 8/9)")
		fmt.Println("  deb-direct  - Direct .deb download (universal fallback)")
		fmt.Println("  rpm-direct  - Direct .rpm download (universal fallback)")
		fmt.Println("  snap        - Chromium via Snap (open-source alternative)")
		fmt.Println()
		fmt.Println("Auto-detection:")
		fmt.Println("  If no variant is specified, Portunix will automatically detect")
		fmt.Println("  your Linux distribution and choose the best installation method.")
	} else if runtime.GOOS == "windows" {
		fmt.Println("Available variants for Windows:")
		fmt.Println("  stable      - Stable release channel (default)")
		fmt.Println("  beta        - Beta testing channel")
		fmt.Println("  dev         - Developer channel")
	} else if runtime.GOOS == "darwin" {
		fmt.Println("Available variants for macOS:")
		fmt.Println("  stable      - Stable release channel (default)")
	}
	
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix install chrome")
	fmt.Println("  portunix install chrome --variant ubuntu")
	fmt.Println("  portunix install chrome --variant snap")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --variant string   Specify installation variant")
	fmt.Println("  -h, --help         help for chrome")
	fmt.Println()
	fmt.Println("Note: Chrome requires acceptance of Google's Terms of Service.")
	fmt.Println("      Consider 'chromium' via snap for an open-source alternative.")

	return nil
}

// parseInstallFlags parses command line arguments for install command
func parseInstallFlags(arguments []string) *InstallOptions {
	options := &InstallOptions{}

	// Get package name (first non-flag argument)
	for _, arg := range arguments {
		if !strings.HasPrefix(arg, "--") && !strings.HasPrefix(arg, "-") {
			options.PackageName = arg
			break
		}
	}

	// Parse flags
	for i := 0; i < len(arguments); i++ {
		arg := arguments[i]

		switch {
		case arg == "--dry-run":
			options.DryRun = true
		case arg == "--list-methods":
			options.ListMethods = true
		case arg == "--gui":
			options.GUI = true
		case arg == "--embeddable":
			options.Embeddable = true
		case arg == "-y":
			options.AutoAccept = true
		case strings.HasPrefix(arg, "--method="):
			options.Method = strings.TrimPrefix(arg, "--method=")
		case arg == "--method" && i+1 < len(arguments):
			i++
			options.Method = arguments[i]
		case strings.HasPrefix(arg, "--version="):
			options.Version = strings.TrimPrefix(arg, "--version=")
		case arg == "--version" && i+1 < len(arguments):
			i++
			options.Version = arguments[i]
		case strings.HasPrefix(arg, "--variant="):
			options.Variant = strings.TrimPrefix(arg, "--variant=")
		case arg == "--variant" && i+1 < len(arguments):
			i++
			options.Variant = arguments[i]
		case !strings.HasPrefix(arg, "-") && arg != options.PackageName:
			// Handle positional variant (legacy compatibility)
			if options.Variant == "" {
				options.Variant = arg
			} else {
				options.OtherArgs = append(options.OtherArgs, arg)
			}
		}
	}

	return options
}

// showPackageMethods displays available installation methods for a package
func showPackageMethods(packageName string) {
	config, err := LoadInstallConfig()
	if err != nil {
		fmt.Printf("Error loading package configuration: %v\n", err)
		return
	}

	pkg, exists := config.Packages[packageName]
	if !exists {
		fmt.Printf("Package '%s' not found\n", packageName)
		return
	}

	fmt.Printf("Available installation methods for '%s':\n", packageName)
	fmt.Printf("  Package: %s\n", pkg.Name)
	fmt.Printf("  Description: %s\n", pkg.Description)
	fmt.Println()

	// Get current OS platform
	currentOS := GetOperatingSystem()
	platform, exists := pkg.Platforms[currentOS]
	if !exists {
		if currentOS == "windows_sandbox" {
			platform, exists = pkg.Platforms["windows"]
		}
		if !exists {
			fmt.Printf("No installation methods available for platform: %s\n", currentOS)
			return
		}
	}

	fmt.Printf("Platform: %s\n", currentOS)
	fmt.Println("Methods:")

	// Show variants as methods
	for variantName, variant := range platform.Variants {
		methodType := platform.Type
		if variant.Type != "" {
			methodType = variant.Type
		}

		// Determine if this is the preferred (default) method
		isPreferred := ""
		if variantName == pkg.DefaultVariant {
			isPreferred = " (preferred)"
		}

		fmt.Printf("  %s%s - %s [%s]\n", variantName, isPreferred, getMethodDescription(methodType, &variant), methodType)

		// Show version support if available
		if variant.Version != "" {
			fmt.Printf("    Version: %s\n", variant.Version)
		}

		// Show supported distributions if available
		if methodType == "repository" || methodType == "apt" {
			distributions := variant.GetDistributionsList()
			if len(distributions) > 0 {
				fmt.Printf("    Distributions: %s\n", strings.Join(distributions, ", "))
			}
		}
	}

	fmt.Println()
	fmt.Println("Usage examples:")
	fmt.Printf("  portunix install %s                    # Use preferred method\n", packageName)
	fmt.Printf("  portunix install %s --method=<method>  # Use specific method\n", packageName)
	fmt.Printf("  portunix install %s --dry-run          # Preview installation\n", packageName)
}

// getMethodDescription returns a description for the installation method
func getMethodDescription(methodType string, variant *VariantConfig) string {
	switch methodType {
	case "apt":
		return "Install via APT package manager"
	case "deb":
		return "Install .deb package directly"
	case "snap":
		return "Install via Snap package manager"
	case "repository":
		return "Install via distribution repository"
	case "msi":
		return "Install MSI package"
	case "exe":
		return "Install executable package"
	case "zip":
		return "Extract ZIP archive"
	case "tar.gz":
		return "Extract TAR.GZ archive"
	case "chocolatey":
		return "Install via Chocolatey package manager"
	case "winget":
		return "Install via Windows Package Manager"
	case "pip":
		return "Install via Python pip"
	case "pipx":
		return "Install via pipx (isolated environment)"
	case "direct_download":
		return "Direct download and install"
	case "redirect":
		if variant.RedirectTo != "" {
			return fmt.Sprintf("Redirect to %s package", variant.RedirectTo)
		}
		return "Redirect to another package"
	default:
		return fmt.Sprintf("Install via %s", methodType)
	}
}
