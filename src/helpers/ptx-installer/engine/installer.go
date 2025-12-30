package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

// InstallOptions contains options for package installation
type InstallOptions struct {
	PackageName string
	Variant     string
	DryRun      bool
	Force       bool
}

// Installer handles package installation operations
type Installer struct {
	registry  *registry.PackageRegistry
	cacheDir  string
	assetsPath string
}

// NewInstaller creates a new installer instance
func NewInstaller(assetsPath string) (*Installer, error) {
	// Load package registry
	reg, err := registry.LoadPackageRegistry(assetsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load package registry: %w", err)
	}

	// Determine cache directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".portunix", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Installer{
		registry:   reg,
		cacheDir:   cacheDir,
		assetsPath: assetsPath,
	}, nil
}

// Install installs a package with the given options
func (i *Installer) Install(options *InstallOptions) error {
	fmt.Printf("\n🔧 Installing package: %s\n", options.PackageName)

	// Get package from registry
	pkg, err := i.registry.GetPackage(options.PackageName)
	if err != nil {
		return fmt.Errorf("package not found: %w", err)
	}

	fmt.Printf("📦 Package: %s\n", pkg.Metadata.DisplayName)
	fmt.Printf("📝 Description: %s\n", pkg.Metadata.Description)

	// Get current platform
	currentOS := GetOperatingSystem()
	fmt.Printf("💻 Platform: %s\n", currentOS)

	// Check if package supports current platform
	platformSpec, exists := pkg.Spec.Platforms[currentOS]
	if !exists {
		// Try fallback for windows_sandbox
		if currentOS == "windows_sandbox" {
			platformSpec, exists = pkg.Spec.Platforms["windows"]
		}
		if !exists {
			return fmt.Errorf("package %s not available for platform %s", options.PackageName, currentOS)
		}
	}

	// Determine variant to install
	variant := options.Variant
	if variant == "" {
		// Use default variant - prefer "default" or "standard", otherwise first available
		if _, exists := platformSpec.Variants["default"]; exists {
			variant = "default"
		} else if _, exists := platformSpec.Variants["standard"]; exists {
			variant = "standard"
		} else {
			// Fallback to first available variant
			for variantName := range platformSpec.Variants {
				variant = variantName
				break
			}
		}
	}

	// Check if variant exists
	variantSpec, exists := platformSpec.Variants[variant]
	if !exists {
		return fmt.Errorf("variant %s not found for package %s", variant, options.PackageName)
	}

	fmt.Printf("🎯 Variant: %s (version: %s)\n", variant, variantSpec.Version)

	// Handle dry-run
	if options.DryRun {
		fmt.Println("\n🔍 DRY RUN MODE - No actual installation will be performed")
		fmt.Printf("   Would install: %s\n", pkg.Metadata.Name)
		fmt.Printf("   Variant: %s\n", variant)
		fmt.Printf("   Version: %s\n", variantSpec.Version)
		fmt.Printf("   Type: %s\n", platformSpec.Type)

		if variantSpec.URL != "" {
			fmt.Printf("   Download URL: %s\n", variantSpec.URL)
		}

		return nil
	}

	// Resolve dependencies first
	if len(pkg.Spec.Dependencies) > 0 {
		fmt.Printf("\n📋 Checking dependencies: %v\n", pkg.Spec.Dependencies)

		deps, err := i.registry.ResolveDependencies(options.PackageName)
		if err != nil {
			return fmt.Errorf("dependency resolution failed: %w", err)
		}

		fmt.Printf("✅ Dependency resolution: %d packages in order\n", len(deps))

		// Install dependencies (simplified - in full implementation would check if already installed)
		for _, dep := range deps {
			if dep != options.PackageName {
				fmt.Printf("⚠️  Dependency %s should be installed first\n", dep)
			}
		}
	}

	// Perform installation based on platform type
	fmt.Printf("\n🚀 Starting installation (type: %s)...\n", platformSpec.Type)

	switch platformSpec.Type {
	case "tar.gz", "zip":
		return i.installArchive(&platformSpec, &variantSpec, options)
	case "deb":
		return i.installDeb(&platformSpec, &variantSpec, options)
	case "apt":
		return i.installApt(&platformSpec, &variantSpec, options)
	case "dnf", "yum":
		return i.installDnf(&platformSpec, &variantSpec, options)
	case "snap":
		return i.installSnap(&platformSpec, &variantSpec, options)
	case "msi", "exe":
		return i.installWindowsBinary(&platformSpec, &variantSpec, options)
	default:
		return fmt.Errorf("installation type %s not yet implemented in ptx-installer", platformSpec.Type)
	}
}

// installArchive installs from archive (tar.gz, zip)
func (i *Installer) installArchive(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) error {
	// Determine download URL (support both single URL and architecture-specific URLs)
	downloadURL := variant.URL
	if downloadURL == "" && len(variant.URLs) > 0 {
		// Select URL based on architecture
		arch := GetArchitecture()
		var ok bool
		downloadURL, ok = variant.URLs[arch]
		if !ok {
			return fmt.Errorf("no download URL found for architecture %s", arch)
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no download URL specified for archive installation")
	}

	// Download archive to cache
	fmt.Printf("📥 Downloading archive from: %s\n", downloadURL)
	archivePath, err := DownloadFileWithProperFilename(downloadURL, i.cacheDir)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Determine extraction directory
	extractTo := variant.ExtractTo
	homeDir, _ := os.UserHomeDir()

	// Determine fallback directory based on OS
	var fallbackDir string
	if runtime.GOOS == "windows" {
		// Windows: Use AppData\Local\Programs (standard user app location)
		fallbackDir = filepath.Join(homeDir, "AppData", "Local", "Programs", options.PackageName)
	} else {
		// Linux/macOS: Use ~/.local/share/portunix/packages
		fallbackDir = filepath.Join(homeDir, ".local", "share", "portunix", "packages", options.PackageName)
	}

	if extractTo == "" {
		extractTo = fallbackDir
	}

	// Ensure extract directory exists, with fallback to user directory
	if err := os.MkdirAll(extractTo, 0755); err != nil {
		// Primary path failed (likely permission denied), try fallback
		if extractTo != fallbackDir {
			fmt.Printf("⚠️  Cannot create %s (permission denied), using user directory\n", extractTo)
			extractTo = fallbackDir
			if err := os.MkdirAll(extractTo, 0755); err != nil {
				return fmt.Errorf("failed to create extract directory: %w", err)
			}
			fmt.Printf("📁 Using fallback: %s\n", extractTo)
		} else {
			return fmt.Errorf("failed to create extract directory: %w", err)
		}
	}

	// Extract archive
	fmt.Printf("📦 Extracting to: %s\n", extractTo)
	if err := ExtractArchive(archivePath, extractTo); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// If binary name specified, find and link it
	if variant.Binary != "" {
		fmt.Printf("🔍 Looking for binary: %s\n", variant.Binary)
		binaryPath, err := FindBinaryInExtracted(extractTo, variant.Binary)
		if err != nil {
			return fmt.Errorf("failed to find binary: %w", err)
		}

		// Create symlink in ~/.local/bin
		homeDir, _ := os.UserHomeDir()
		binDir := filepath.Join(homeDir, ".local", "bin")
		os.MkdirAll(binDir, 0755)

		linkPath := filepath.Join(binDir, variant.Binary)

		// Remove existing symlink if exists
		os.Remove(linkPath)

		// Create symlink
		if err := os.Symlink(binaryPath, linkPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}

		fmt.Printf("✅ Created symlink: %s -> %s\n", linkPath, binaryPath)
		fmt.Printf("💡 Make sure %s is in your PATH\n", binDir)
	}

	// Run post-install commands if specified
	if len(variant.PostInstall) > 0 {
		fmt.Println("🔧 Running post-install commands...")
		for _, cmd := range variant.PostInstall {
			// Replace original extractTo path with actual path if fallback was used
			if variant.ExtractTo != "" && variant.ExtractTo != extractTo {
				// Normalize paths for replacement (handle both / and \ on Windows)
				originalPath := filepath.ToSlash(variant.ExtractTo)
				actualPath := filepath.ToSlash(extractTo)
				cmd = strings.ReplaceAll(cmd, originalPath, actualPath)
				// Also try with backslashes for Windows
				originalPathWin := strings.ReplaceAll(variant.ExtractTo, "/", "\\")
				actualPathWin := strings.ReplaceAll(extractTo, "/", "\\")
				cmd = strings.ReplaceAll(cmd, originalPathWin, actualPathWin)
			}
			fmt.Printf("   Running: %s\n", cmd)
			// TODO: Implement post-install command execution
		}
	}

	return nil
}

// installDeb installs a .deb package
func (i *Installer) installDeb(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) error {
	// Determine download URL (support both single URL and architecture-specific URLs)
	downloadURL := variant.URL
	if downloadURL == "" && len(variant.URLs) > 0 {
		// Select URL based on architecture
		arch := GetArchitecture()
		var ok bool
		downloadURL, ok = variant.URLs[arch]
		if !ok {
			return fmt.Errorf("no download URL found for architecture %s", arch)
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no download URL specified for deb installation")
	}

	// Download .deb file to cache
	fmt.Printf("📥 Downloading .deb package from: %s\n", downloadURL)
	debPath, err := DownloadFileWithProperFilename(downloadURL, i.cacheDir)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Install using dpkg
	return InstallDebPackage(debPath)
}

// installApt installs via APT package manager
func (i *Installer) installApt(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) error {
	if len(variant.Packages) == 0 {
		return fmt.Errorf("no packages specified for APT installation")
	}

	// Install packages via APT
	return InstallViaAPT(variant.Packages, variant.RequiresSudo)
}

// installDnf installs via DNF/YUM package manager
func (i *Installer) installDnf(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) error {
	if len(variant.Packages) == 0 {
		return fmt.Errorf("no packages specified for DNF/YUM installation")
	}

	// Install packages via DNF/YUM
	return InstallViaDNF(variant.Packages, variant.RequiresSudo)
}

// installSnap installs via Snap package manager
func (i *Installer) installSnap(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) error {
	if len(variant.Packages) == 0 {
		return fmt.Errorf("no packages specified for Snap installation")
	}

	// Check if classic confinement needed (usually for development tools)
	classic := false
	if variant.InstallArgs != nil {
		for _, arg := range variant.InstallArgs {
			if arg == "--classic" {
				classic = true
				break
			}
		}
	}

	// Install packages via Snap
	return InstallViaSnap(variant.Packages, classic)
}

// installWindowsBinary installs Windows binary (MSI/EXE)
func (i *Installer) installWindowsBinary(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) error {
	fmt.Println("📦 Windows binary installation not yet fully implemented")
	return fmt.Errorf("windows binary installation coming in Phase 2 completion")
}

// GetRegistry returns the package registry
func (i *Installer) GetRegistry() *registry.PackageRegistry {
	return i.registry
}

// GetCacheDir returns the cache directory
func (i *Installer) GetCacheDir() string {
	return i.cacheDir
}
