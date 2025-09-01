package install

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"portunix.cz/app"
	"portunix.cz/app/docker"
	"portunix.cz/app/podman"
)

func ToArguments(what string) []string {
	arguments := make([]string, 1)
	arguments[0] = what
	return arguments
}

func Install(arguments []string) {
	if len(arguments) == 0 {
		return
	}

	// Special handling for Docker and Podman installation
	packageName := arguments[0]
	if packageName == "docker" {
		// Check for -y flag
		autoAccept := false
		for _, arg := range arguments[1:] {
			if arg == "-y" {
				autoAccept = true
				break
			}
		}

		if err := docker.InstallDocker(autoAccept); err != nil {
			return
		}
		return
	}

	if packageName == "podman" {
		// Check for -y flag
		autoAccept := false
		for _, arg := range arguments[1:] {
			if arg == "-y" {
				autoAccept = true
				break
			}
		}

		if err := podman.InstallPodman(autoAccept); err != nil {
			return
		}
		return
	}

	// Try new JSON-based installer first
	variant := ""

	// Check if variant is specified
	if len(arguments) > 1 {
		for _, arg := range arguments[1:] {
			if arg != "--gui" && arg != "--embeddable" && !strings.HasPrefix(arg, "--") {
				variant = arg
				break
			}
		}
	}

	// Check if it's a preset first
	if err := InstallPreset(packageName); err == nil {
		return // Success with preset
	}

	// Try to install using new system
	if err := InstallPackage(packageName, variant); err == nil {
		return // Success with new system
	} else {
		fmt.Printf("Error installing package '%s': %v\n", packageName, err)
		return // Show error and exit instead of falling back
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
	config, err := LoadInstallConfig()
	if err != nil {
		return fmt.Errorf("failed to load install config: %w", err)
	}

	preset, exists := config.Presets[presetName]
	if !exists {
		return fmt.Errorf("preset '%s' not found", presetName)
	}

	fmt.Printf("════════════════════════════════════════════════\n")
	fmt.Printf("📦 INSTALLING PRESET: %s\n", preset.Name)
	fmt.Printf("════════════════════════════════════════════════\n")
	fmt.Printf("📄 Description: %s\n", preset.Description)
	fmt.Printf("📦 Packages: %d\n", len(preset.Packages))
	fmt.Printf("════════════════════════════════════════════════\n")

	for i, pkg := range preset.Packages {
		fmt.Printf("Installing package %d/%d: %s (variant: %s)\n", i+1, len(preset.Packages), pkg.Name, pkg.Variant)
		if err := InstallPackage(pkg.Name, pkg.Variant); err != nil {
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
