package install

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// FallbackStrategy defines how fallback should be handled
type FallbackStrategy string

const (
	FallbackAuto            FallbackStrategy = "auto"                // Automatic fallback without confirmation
	FallbackAutoConfirm     FallbackStrategy = "auto_with_confirmation" // Automatic with user confirmation
	FallbackManual          FallbackStrategy = "manual"             // Manual fallback only
	FallbackDisabled        FallbackStrategy = "disabled"           // No fallback
)

// FallbackManager handles the fallback chain execution
type FallbackManager struct {
	versionMatcher *VersionMatcher
}

// NewFallbackManager creates a new fallback manager
func NewFallbackManager() *FallbackManager {
	return &FallbackManager{
		versionMatcher: NewVersionMatcher(),
	}
}

// ExecuteFallback implements cascading fallback with user communication
func (fb *FallbackManager) ExecuteFallback(packageName string, originalVariant string, config *InstallConfig, reason string, fallbackVariants []string, strategy FallbackStrategy) error {
	// Show the original failure
	fmt.Println("\n❌ Installation FAILED!")
	fmt.Printf("Package: %s\n", packageName)
	fmt.Printf("Variant: %s\n", originalVariant)
	fmt.Printf("Reason: %s\n", reason)
	
	// If no fallback variants or fallback disabled, show manual instructions
	if len(fallbackVariants) == 0 || strategy == FallbackDisabled {
		return fb.showManualInstructions(packageName, originalVariant)
	}

	fmt.Println("\n✅ Alternative options:")
	
	// Show available fallback options
	for i, variant := range fallbackVariants {
		fmt.Printf("%d. Install via %s variant:\n", i+1, variant)
		fmt.Printf("   ./portunix install %s --variant %s\n", packageName, variant)
		
		// Get variant description if available
		if description := fb.getVariantDescription(packageName, variant, config); description != "" {
			fmt.Printf("   %s\n", description)
		}
		fmt.Println()
	}

	// Show manual option
	fmt.Printf("%d. Use manual installation:\n", len(fallbackVariants)+1)
	if manualURL := fb.getManualInstallationURL(packageName); manualURL != "" {
		fmt.Printf("   See: %s\n", manualURL)
	} else {
		fmt.Printf("   Follow official documentation for manual installation\n")
	}
	fmt.Println()

	// Handle strategy
	switch strategy {
	case FallbackAuto:
		// Automatically try first fallback variant
		if len(fallbackVariants) > 0 {
			fmt.Printf("🔄 Automatically trying %s variant...\n", fallbackVariants[0])
			return fb.attemptFallbackInstall(packageName, fallbackVariants[0])
		}
		
	case FallbackAutoConfirm:
		// Ask user for confirmation
		if len(fallbackVariants) > 0 {
			if fb.promptUserConfirmation(fmt.Sprintf("Would you like to try %s variant automatically?", fallbackVariants[0])) {
				fmt.Printf("🔄 Installing via %s variant...\n", fallbackVariants[0])
				return fb.attemptFallbackInstall(packageName, fallbackVariants[0])
			}
		}
		
	case FallbackManual:
		// Just show options, no automatic execution
		fmt.Println("Please choose one of the above options and run the command manually.")
		
	case FallbackDisabled:
		// This case is handled above
		break
	}

	return fmt.Errorf("installation failed and no automatic fallback was executed")
}

// attemptFallbackInstall attempts to install using fallback variant
func (fb *FallbackManager) attemptFallbackInstall(packageName, fallbackVariant string) error {
	// Try to install the fallback variant
	err := InstallPackage(packageName, fallbackVariant)
	if err != nil {
		fmt.Printf("❌ Fallback installation also failed: %v\n", err)
		return err
	}
	
	fmt.Printf("✅ Successfully installed %s using %s variant!\n", packageName, fallbackVariant)
	return nil
}

// promptUserConfirmation prompts user for yes/no confirmation
func (fb *FallbackManager) promptUserConfirmation(message string) bool {
	fmt.Printf("%s [Y/n] ", message)
	
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	
	response = strings.TrimSpace(strings.ToLower(response))
	
	// Default to yes if empty response
	if response == "" || response == "y" || response == "yes" {
		return true
	}
	
	return false
}

// getVariantDescription returns description for a variant
func (fb *FallbackManager) getVariantDescription(packageName, variant string, config *InstallConfig) string {
	descriptions := map[string]map[string]string{
		"powershell": {
			"snap":       "(universal, recommended for newer distributions)",
			"ubuntu":     "(native APT repository for Ubuntu/Kubuntu)",
			"debian":     "(native APT repository for Debian)",
			"fedora":     "(native DNF repository for Fedora)",
			"rocky":      "(native DNF repository for Rocky Linux/CentOS)",
			"mint":       "(native APT repository for Linux Mint)",
			"elementary": "(native APT repository for Elementary OS)",
		},
	}
	
	if packageDescriptions, exists := descriptions[packageName]; exists {
		if description, exists := packageDescriptions[variant]; exists {
			return description
		}
	}
	
	// Try to get description from config
	if pkg, exists := config.Packages[packageName]; exists {
		if platform, exists := pkg.Platforms["linux"]; exists {
			if variantConfig, exists := platform.Variants[variant]; exists {
				if len(variantConfig.Distributions) > 0 {
					if variantConfig.Distributions[0] == "universal" {
						return "(universal compatibility)"
					}
					return fmt.Sprintf("(for %s)", strings.Join(variantConfig.Distributions, ", "))
				}
			}
		}
	}
	
	return ""
}

// getManualInstallationURL returns manual installation URL for package
func (fb *FallbackManager) getManualInstallationURL(packageName string) string {
	urls := map[string]string{
		"powershell": "https://docs.microsoft.com/powershell/scripting/install/installing-powershell-on-linux",
		"vscode":     "https://code.visualstudio.com/docs/setup/linux",
		"docker":     "https://docs.docker.com/engine/install/",
		"podman":     "https://podman.io/docs/installation",
		"java":       "https://adoptium.net/installation/linux/",
		"python":     "https://www.python.org/downloads/source/",
		"go":         "https://golang.org/doc/install",
	}
	
	return urls[packageName]
}

// showManualInstructions shows manual installation instructions when no fallback available
func (fb *FallbackManager) showManualInstructions(packageName, variant string) error {
	fmt.Println("\n📖 Manual Installation Required")
	fmt.Printf("Package: %s (variant: %s)\n", packageName, variant)
	fmt.Println()
	
	if url := fb.getManualInstallationURL(packageName); url != "" {
		fmt.Printf("Please follow the official documentation: %s\n", url)
	} else {
		fmt.Printf("Please consult the official documentation for %s installation.\n", packageName)
	}
	
	fmt.Println()
	fmt.Println("Common alternatives:")
	fmt.Printf("• Check if package is available via system package manager\n")
	fmt.Printf("• Download from official website\n")
	fmt.Printf("• Use container/snap version if available\n")
	
	return fmt.Errorf("manual installation required - no automatic fallback available")
}

// ShouldTriggerFallback determines if fallback should be triggered based on error and support level
func (fb *FallbackManager) ShouldTriggerFallback(err error, supportLevel SupportLevel) bool {
	if err == nil {
		return false
	}
	
	// Always trigger fallback for unsupported versions
	if supportLevel == Unsupported {
		return true
	}
	
	// Trigger fallback for common installation failures
	errorString := err.Error()
	triggerKeywords := []string{
		"repository setup failed",
		"package installation failed",
		"not supported",
		"not found",
		"failed to download",
		"command not found",
	}
	
	for _, keyword := range triggerKeywords {
		if strings.Contains(strings.ToLower(errorString), keyword) {
			return true
		}
	}
	
	return false
}

// CreateVersionCompatibilityError creates a detailed error for version compatibility issues
func (fb *FallbackManager) CreateVersionCompatibilityError(packageName, variant, version string, supportLevel SupportLevel) error {
	switch supportLevel {
	case Unsupported:
		return fmt.Errorf("%s not explicitly supported for %s %s variant", version, packageName, variant)
	case Experimental:
		return fmt.Errorf("%s is experimental for %s %s variant - proceed with caution", version, packageName, variant)
	case Compatible:
		// Compatible versions should not trigger errors, but we include this for completeness
		return fmt.Errorf("%s is compatible but untested for %s %s variant", version, packageName, variant)
	case Supported:
		// Supported versions should not create errors
		return nil
	default:
		return fmt.Errorf("unknown support level for %s %s on %s", packageName, variant, version)
	}
}