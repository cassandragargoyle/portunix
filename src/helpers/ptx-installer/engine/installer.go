/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"bufio"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
	"portunix.ai/portunix/src/pkg/archive"
)

// EmbeddedScriptsFS holds the embedded scripts filesystem (set from main package)
var EmbeddedScriptsFS embed.FS

// SetEmbeddedScripts sets the embedded scripts filesystem from the main package
func SetEmbeddedScripts(scriptsFS embed.FS) {
	EmbeddedScriptsFS = scriptsFS
}

// checkExistingDirectory checks if target directory exists and is not empty
// Returns true if installation should proceed, false if user cancelled
func checkExistingDirectory(targetDir string, force bool) (bool, error) {
	// Check if directory exists
	info, err := os.Stat(targetDir)
	if os.IsNotExist(err) {
		return true, nil // Directory doesn't exist, proceed
	}
	if err != nil {
		return false, fmt.Errorf("failed to check directory: %w", err)
	}

	if !info.IsDir() {
		return false, fmt.Errorf("target path exists but is not a directory: %s", targetDir)
	}

	// Check if directory is empty
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return false, fmt.Errorf("failed to read directory: %w", err)
	}

	if len(entries) == 0 {
		return true, nil // Directory is empty, proceed
	}

	// Directory exists and is not empty
	fmt.Printf("\n⚠️  Target directory already exists and is not empty:\n")
	fmt.Printf("   %s\n", targetDir)
	fmt.Printf("   Contains %d item(s)\n\n", len(entries))

	if force {
		fmt.Println("🔄 Force mode enabled, removing existing directory...")
		if err := os.RemoveAll(targetDir); err != nil {
			return false, fmt.Errorf("failed to remove existing directory: %w", err)
		}
		fmt.Println("✅ Existing directory removed")
		return true, nil
	}

	// Ask user for confirmation
	fmt.Print("Do you want to remove the existing directory and reinstall? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "y" || response == "yes" {
		fmt.Println("🔄 Removing existing directory...")
		if err := os.RemoveAll(targetDir); err != nil {
			return false, fmt.Errorf("failed to remove existing directory: %w", err)
		}
		fmt.Println("✅ Existing directory removed")
		return true, nil
	}

	fmt.Println("❌ Installation cancelled by user")
	return false, nil
}

// expandEnvVars expands environment variables in a string
// Supports both ${VAR} and %VAR% syntax for cross-platform compatibility
func expandEnvVars(s string) string {
	if s == "" {
		return s
	}

	// First expand ${VAR} syntax (Unix-style)
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	result := re.ReplaceAllStringFunc(s, func(match string) string {
		varName := match[2 : len(match)-1] // Extract VAR from ${VAR}
		if value := os.Getenv(varName); value != "" {
			return value
		}
		// Handle special Windows variables that might not be in env
		switch strings.ToUpper(varName) {
		case "PROGRAMFILES":
			if runtime.GOOS == "windows" {
				return os.Getenv("ProgramFiles")
			}
		case "LOCALAPPDATA":
			if runtime.GOOS == "windows" {
				return os.Getenv("LOCALAPPDATA")
			}
		case "APPDATA":
			if runtime.GOOS == "windows" {
				return os.Getenv("APPDATA")
			}
		case "USERPROFILE":
			if runtime.GOOS == "windows" {
				return os.Getenv("USERPROFILE")
			}
		}
		return match // Return original if not found
	})

	// Then expand %VAR% syntax (Windows-style)
	re = regexp.MustCompile(`%([^%]+)%`)
	result = re.ReplaceAllStringFunc(result, func(match string) string {
		varName := match[1 : len(match)-1] // Extract VAR from %VAR%
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match // Return original if not found
	})

	return result
}

// installResult is what each install method reports back so the engine can
// expand ${install_path} / ${extract_to} placeholders in PATH_APPEND
// (Issue #187 phase 2). Methods that don't produce a structured path
// (apt, dnf, snap, ...) leave both fields empty.
type installResult struct {
	// installPath is the final on-disk location of the installed package.
	// Archives: actualRoot after FindExtractedRoot. MSI/EXE: variant.InstallPath
	// (env-vars expanded). Download: target directory.
	installPath string
	// extractTo is the post-fallback pre-detection extract directory
	// (archives only). Lets packages like maven keep
	// `${extract_to}/apache-maven-3.9.9/bin` semantics.
	extractTo string
}

// InstallOptions contains options for package installation
type InstallOptions struct {
	PackageName string
	Variant     string
	InstallPath string // Target path for packages that require it (e.g., docusaurus)
	DryRun      bool
	Force       bool
	// Database connection overrides for container-type installs that read
	// PostgreSQL-style env keys (HOST, PORT, USER, PASSWORD). Empty values
	// leave the variant's JSON defaults untouched.
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
}

// Installer handles package installation operations
type Installer struct {
	registry   *registry.PackageRegistry
	cacheDir   string
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

// ResolveVersion maps a `--version` selector to a concrete variant name for the
// package on the current platform (issue #035 per-assistant version management).
// A selector matches either a variant whose name equals it (e.g. "21" for Java)
// or a variant whose Version field equals it (e.g. "17.0.16_8"). Returns an
// error listing the available versions when no variant matches.
func (i *Installer) ResolveVersion(packageName, version string) (string, error) {
	pkg, err := i.registry.GetPackage(packageName)
	if err != nil {
		return "", fmt.Errorf("package not found: %w", err)
	}

	currentOS := GetOperatingSystem()
	platformSpec, exists := pkg.Spec.Platforms[currentOS]
	if !exists && currentOS == "windows_sandbox" {
		platformSpec, exists = pkg.Spec.Platforms["windows"]
	}
	if !exists {
		return "", fmt.Errorf("package %s not available for platform %s", packageName, currentOS)
	}

	// Exact variant-name match takes precedence (e.g. Java "21").
	if _, ok := platformSpec.Variants[version]; ok {
		return version, nil
	}

	// Otherwise match against each variant's declared Version field.
	available := make([]string, 0, len(platformSpec.Variants))
	for name, spec := range platformSpec.Variants {
		if spec.Version == version {
			return name, nil
		}
		label := name
		if spec.Version != "" && spec.Version != name {
			label = fmt.Sprintf("%s (%s)", name, spec.Version)
		}
		available = append(available, label)
	}
	sort.Strings(available)

	return "", fmt.Errorf("version %q not available for %s on %s\n   Available versions: %s",
		version, packageName, currentOS, strings.Join(available, ", "))
}

// elevateInstall re-launches the current `portunix install …` command with
// elevated privileges to satisfy a requiresAdmin variant, then returns. The
// elevated child detects IsAdmin()==true (see decideElevation) and installs
// directly — no second prompt, no loop. On Windows it raises a single UAC
// prompt; on Linux/macOS it re-execs under sudo. When the user declines UAC
// (or sudo is unavailable) it surfaces an actionable fallback message.
func (i *Installer) elevateInstall(options *InstallOptions) error {
	if runtime.GOOS == "windows" {
		fmt.Println("🔐 This installation requires Administrator privileges.")
		fmt.Println("   A UAC prompt will appear — please click Yes to continue.")
		fmt.Println("   Note: installation output appears in the elevated window.")
	} else {
		fmt.Println("🔐 This installation requires root privileges — elevating via sudo...")
	}

	err := reExecElevated(os.Args[1:])
	if err == nil {
		return nil
	}
	if errors.Is(err, errUACDeclined) {
		return fmt.Errorf("❌ UAC elevation was declined.\n"+
			"   Installing '%s' requires Administrator privileges.\n"+
			"   Please re-run the command and approve the UAC prompt,\n"+
			"   or run Portunix from an elevated PowerShell or cmd.", options.PackageName)
	}
	if errors.Is(err, errSudoMissing) {
		return fmt.Errorf("❌ This installation requires root privileges, but sudo was not found.\n"+
			"   Please re-run as root:  sudo %s", strings.Join(os.Args, " "))
	}
	return fmt.Errorf("elevation failed: %w", err)
}

// Install installs a package with the given options
func (i *Installer) Install(options *InstallOptions) error {
	fmt.Printf("\n🔧 Installing package: %s\n", options.PackageName)

	// Get package from registry
	pkg, err := i.registry.GetPackage(options.PackageName)
	if err != nil {
		return fmt.Errorf("package not found: %w", err)
	}

	// Bundles aggregate several packages — install each member in order
	if pkg.Kind == "Bundle" {
		return i.installBundle(pkg, options)
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
		// Auto-detect variant based on system package manager
		variant = i.autoDetectVariant(&platformSpec)
	}

	// Check if variant exists
	if _, exists := platformSpec.Variants[variant]; !exists {
		return UnknownVariantError(options.PackageName, variant, GetOperatingSystem(), platformSpec.Variants)
	}

	// A package manager variant without its package manager falls back to
	// the package's direct installer (issue #222)
	if fallback, binary := packageManagerFallback(&platformSpec, variant, isCommandAvailable); fallback != "" {
		fmt.Printf("⚠️  %s not found, falling back to official installer (variant: %s)\n", binary, fallback)
		variant = fallback
	}
	variantSpec := platformSpec.Variants[variant]

	// Variants tracking the newest release resolve their version and URLs now
	variant, variantSpec, err = resolveOrFallback(&platformSpec, variant, variantSpec)
	if err != nil {
		return err
	}

	fmt.Printf("🎯 Variant: %s (version: %s)\n", variant, variantSpec.Version)

	// Check if admin/root privileges are required. Rather than failing when the
	// current process is not elevated, request elevation on-demand for this
	// exact command (issue #189): re-launch the same `portunix install …`
	// invocation under UAC (Windows) / sudo (Linux/macOS). dry-run always
	// proceeds in-process below so the plan can be previewed without a prompt.
	if variantSpec.RequiresAdmin {
		if decideElevation(options.DryRun, IsAdmin()) == elevationUAC {
			return i.elevateInstall(options)
		}
	}

	// Determine effective installation type:
	// Priority 1: Variant-specific type (e.g., pacman variant on Linux)
	// Priority 2: Platform type (fallback)
	effectiveType := platformSpec.Type
	if variantSpec.Type != "" {
		effectiveType = variantSpec.Type
	}

	// Handle dry-run
	if options.DryRun {
		fmt.Println("\n🔍 DRY RUN MODE - No actual installation will be performed")
		fmt.Printf("   Would install: %s\n", pkg.Metadata.Name)
		fmt.Printf("   Variant: %s\n", variant)
		fmt.Printf("   Version: %s\n", variantSpec.Version)
		fmt.Printf("   Type: %s\n", effectiveType)

		if downloadURL := selectArchURL(variantSpec); downloadURL != "" {
			fmt.Printf("   Download URL: %s\n", downloadURL)
		}
		if checksum := selectArchChecksum(variantSpec); checksum != "" {
			fmt.Printf("   Checksum: %s\n", checksum)
		}
		if len(variantSpec.AdditionalFiles) > 0 {
			fmt.Printf("   Additional files: %d\n", len(variantSpec.AdditionalFiles))
			for _, af := range variantSpec.AdditionalFiles {
				fmt.Printf("     - %s\n", af.URL)
			}
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

	// Perform installation based on effective type (variant type takes precedence)
	fmt.Printf("\n🚀 Starting installation (type: %s)...\n", effectiveType)

	var installErr error
	var result installResult
	switch effectiveType {
	case "tar.gz", "zip":
		result, installErr = i.installArchive(&platformSpec, &variantSpec, options)
	case "deb":
		installErr = i.installDeb(&platformSpec, &variantSpec, options)
	case "apt":
		installErr = i.installApt(&platformSpec, &variantSpec, options)
	case "dnf", "yum":
		installErr = i.installDnf(&platformSpec, &variantSpec, options)
	case "snap":
		installErr = i.installSnap(&platformSpec, &variantSpec, options)
	case "pacman":
		installErr = i.installPacman(&platformSpec, &variantSpec, options)
	case "msi", "exe":
		result, installErr = i.installWindowsBinary(&platformSpec, &variantSpec, options)
	case "chocolatey":
		installErr = i.installChocolatey(&platformSpec, &variantSpec, options)
	case "winget":
		installErr = i.installWinget(&platformSpec, &variantSpec, options)
	case "download":
		result, installErr = i.installDownload(&platformSpec, &variantSpec, options)
	case "script":
		installErr = i.installScript(&platformSpec, &variantSpec, options)
	case "container":
		installErr = i.installContainer(&platformSpec, &variantSpec, options)
	default:
		return fmt.Errorf("installation type %s not yet implemented in ptx-installer", effectiveType)
	}
	if installErr != nil {
		return installErr
	}

	// Apply environment.PATH_APPEND after a successful install. Env-var
	// placeholders (${LOCALAPPDATA}, $HOME, %USERPROFILE%, …) are expanded
	// up-front; ${install_path}/${extract_to} are then resolved from the
	// installResult returned by the install method (Issue #187 phase 2).
	// PATH failures are non-fatal — the package is installed even if PATH
	// wiring breaks.
	if pathAppend := platformSpec.Environment["PATH_APPEND"]; pathAppend != "" {
		expanded := expandEnvVars(pathAppend)
		if result.installPath != "" {
			expanded = strings.ReplaceAll(expanded, "${install_path}", result.installPath)
			expanded = strings.ReplaceAll(expanded, "%install_path%", result.installPath)
		}
		if result.extractTo != "" {
			expanded = strings.ReplaceAll(expanded, "${extract_to}", result.extractTo)
			expanded = strings.ReplaceAll(expanded, "%extract_to%", result.extractTo)
		}
		if strings.Contains(expanded, "${install_path}") || strings.Contains(expanded, "${extract_to}") {
			fmt.Printf("⚠️  PATH_APPEND contains unresolved placeholder, skipping: %s\n", expanded)
		} else {
			fmt.Printf("\n🔧 Adding to User PATH: %s\n", expanded)
			if err := AddToUserPath(expanded); err != nil {
				fmt.Printf("⚠️  PATH update failed (install succeeded): %v\n", err)
			}
		}
	}
	return nil
}

// installBundle installs every package listed in a Kind: "Bundle" entry, in
// declaration order. Member options inherit DryRun and Force; installation
// stops at the first member that fails.
func (i *Installer) installBundle(pkg *registry.Package, options *InstallOptions) error {
	fmt.Printf("📦 Bundle: %s\n", pkg.Metadata.DisplayName)
	fmt.Printf("📝 Description: %s\n", pkg.Metadata.Description)
	fmt.Printf("📋 Packages: %v\n", pkg.Spec.Bundle)

	for idx, member := range pkg.Spec.Bundle {
		fmt.Printf("\n────────────────────────────────────────\n")
		fmt.Printf("➡️  [%d/%d] %s\n", idx+1, len(pkg.Spec.Bundle), member)

		memberOptions := &InstallOptions{
			PackageName: member,
			DryRun:      options.DryRun,
			Force:       options.Force,
		}
		if err := i.Install(memberOptions); err != nil {
			return fmt.Errorf("bundle %s: failed to install %s: %w", pkg.Metadata.Name, member, err)
		}
	}

	fmt.Printf("\n✅ Bundle %s complete (%d packages)\n", pkg.Metadata.Name, len(pkg.Spec.Bundle))
	return nil
}

// installArchive installs from archive (tar.gz, zip). Returns an installResult
// so the engine can resolve ${install_path} / ${extract_to} in PATH_APPEND:
// installPath is the post-detection actual root, extractTo is the
// post-fallback pre-detection target.
func (i *Installer) installArchive(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) (installResult, error) {
	// Determine download URL (support both single URL and architecture-specific URLs)
	downloadURL := variant.URL
	if downloadURL == "" && len(variant.URLs) > 0 {
		// Select URL based on architecture
		arch := GetArchitecture()
		var ok bool
		downloadURL, ok = variant.URLs[arch]
		if !ok {
			return installResult{}, fmt.Errorf("no download URL found for architecture %s", arch)
		}
	}

	if downloadURL == "" {
		return installResult{}, fmt.Errorf("no download URL specified for archive installation")
	}

	// Download archive to cache
	fmt.Printf("📥 Downloading archive from: %s\n", downloadURL)
	archivePath, err := archive.DownloadFileWithProperFilename(downloadURL, i.cacheDir)
	if err != nil {
		return installResult{}, fmt.Errorf("download failed: %w", err)
	}

	// Determine extraction directory (expand environment variables)
	extractTo := expandEnvVars(variant.ExtractTo)
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
				return installResult{}, fmt.Errorf("failed to create extract directory: %w", err)
			}
			fmt.Printf("📁 Using fallback: %s\n", extractTo)
		} else {
			return installResult{}, fmt.Errorf("failed to create extract directory: %w", err)
		}
	}

	// Check if target directory already exists and is not empty (after determining final path)
	proceed, err := checkExistingDirectory(extractTo, options.Force)
	if err != nil {
		return installResult{}, fmt.Errorf("directory check failed: %w", err)
	}
	if !proceed {
		return installResult{}, nil // User cancelled installation
	}

	// Extract archive
	fmt.Printf("📦 Extracting to: %s\n", extractTo)
	if err := archive.ExtractArchive(archivePath, extractTo); err != nil {
		return installResult{}, fmt.Errorf("extraction failed: %w", err)
	}

	// Remember the post-fallback pre-detection path — packages like maven
	// reference it via ${extract_to} in PATH_APPEND.
	preDetectionExtractTo := extractTo

	// Find actual root directory (many archives have a single top-level directory)
	actualRoot, err := archive.FindExtractedRoot(extractTo)
	if err != nil {
		fmt.Printf("⚠️  Could not determine extracted root: %v\n", err)
	} else if actualRoot != extractTo {
		fmt.Printf("📁 Detected package root: %s\n", actualRoot)
		extractTo = actualRoot
	}

	// If binary name specified, find and link it
	if variant.Binary != "" {
		fmt.Printf("🔍 Looking for binary: %s\n", variant.Binary)
		binaryPath, err := archive.FindBinaryInExtracted(extractTo, variant.Binary)
		if err != nil {
			return installResult{}, fmt.Errorf("failed to find binary: %w", err)
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
			return installResult{}, fmt.Errorf("failed to create symlink: %w", err)
		}

		fmt.Printf("✅ Created symlink: %s -> %s\n", linkPath, binaryPath)
		fmt.Printf("💡 Make sure %s is in your PATH\n", binDir)
	}

	// Run install script if specified (embedded PowerShell/shell script)
	if len(variant.InstallScript) > 0 {
		firstScript := variant.InstallScript[0]
		if isEmbeddedScript(firstScript) {
			fmt.Println("📜 Running installation script...")
			if err := i.executeEmbeddedScript(firstScript, variant.InstallScriptArgs, extractTo, options.DryRun); err != nil {
				return installResult{}, fmt.Errorf("install script failed: %w", err)
			}
		}
	}

	// Run post-install commands if specified
	if len(variant.PostInstall) > 0 {
		fmt.Println("🔧 Running post-install commands...")
		var postInstallErrors []string
		for _, cmd := range variant.PostInstall {
			// Expand environment variables in command
			cmd = expandEnvVars(cmd)

			// Replace ${install_path} placeholder with actual extraction path
			cmd = strings.ReplaceAll(cmd, "${install_path}", extractTo)
			cmd = strings.ReplaceAll(cmd, "%install_path%", extractTo)

			// Replace original extractTo path with actual path if fallback was used
			originalExtractTo := expandEnvVars(variant.ExtractTo)
			if originalExtractTo != "" && originalExtractTo != extractTo {
				// Normalize paths for replacement (handle both / and \ on Windows)
				originalPath := filepath.ToSlash(originalExtractTo)
				actualPath := filepath.ToSlash(extractTo)
				cmd = strings.ReplaceAll(cmd, originalPath, actualPath)
				// Also try with backslashes for Windows
				originalPathWin := strings.ReplaceAll(originalExtractTo, "/", "\\")
				actualPathWin := strings.ReplaceAll(extractTo, "/", "\\")
				cmd = strings.ReplaceAll(cmd, originalPathWin, actualPathWin)
			}

			fmt.Printf("   Running: %s\n", cmd)

			// Execute the command
			execCmd := shellCommand(cmd)
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			if err := execCmd.Run(); err != nil {
				errMsg := fmt.Sprintf("Command failed: %s (error: %v)", cmd, err)
				fmt.Printf("   ❌ %s\n", errMsg)
				postInstallErrors = append(postInstallErrors, errMsg)
			}
		}

		if len(postInstallErrors) > 0 {
			return installResult{}, fmt.Errorf("❌ Installation failed: %d post-install command(s) failed:\n   - %s",
				len(postInstallErrors), strings.Join(postInstallErrors, "\n   - "))
		}
	}

	return installResult{installPath: extractTo, extractTo: preDetectionExtractTo}, nil
}

// installDownload downloads files directly to a target directory (no extraction).
// Returns installResult so the engine can resolve PATH_APPEND placeholders;
// both installPath and extractTo equal the target directory.
func (i *Installer) installDownload(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) (installResult, error) {
	// Determine download URL
	downloadURL := variant.URL
	if downloadURL == "" && len(variant.URLs) > 0 {
		arch := GetArchitecture()
		var ok bool
		downloadURL, ok = variant.URLs[arch]
		if !ok {
			return installResult{}, fmt.Errorf("no download URL found for architecture %s", arch)
		}
	}

	if downloadURL == "" && len(variant.AdditionalFiles) == 0 {
		return installResult{}, fmt.Errorf("no download URL specified for download installation")
	}

	// Determine target directory
	targetDir := expandEnvVars(variant.ExtractTo)
	homeDir, _ := os.UserHomeDir()

	if targetDir == "" {
		if runtime.GOOS == "windows" {
			targetDir = filepath.Join(homeDir, "AppData", "Local", "Programs", options.PackageName)
		} else {
			targetDir = filepath.Join(homeDir, ".local", "share", "portunix", "packages", options.PackageName)
		}
	}

	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return installResult{}, fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
	}

	// Collect all URLs to download
	type fileDownload struct {
		url      string
		filename string
	}
	var downloads []fileDownload

	// Main file
	if downloadURL != "" {
		// Prefer explicit binary name from variant spec (lets packages rename a
		// versioned upstream asset like tea-0.11.1-windows-amd64.exe to tea.exe
		// without shelling out to move/ren in postInstall).
		filename := variant.Binary
		if filename == "" {
			parts := strings.Split(downloadURL, "/")
			if len(parts) > 0 {
				filename = parts[len(parts)-1]
				if idx := strings.Index(filename, "?"); idx != -1 {
					filename = filename[:idx]
				}
			}
		}
		downloads = append(downloads, fileDownload{url: downloadURL, filename: filename})
	}

	// Additional files
	for _, af := range variant.AdditionalFiles {
		filename := af.Filename
		if filename == "" {
			parts := strings.Split(af.URL, "/")
			if len(parts) > 0 {
				filename = parts[len(parts)-1]
				if idx := strings.Index(filename, "?"); idx != -1 {
					filename = filename[:idx]
				}
			}
		}
		downloads = append(downloads, fileDownload{url: af.URL, filename: filename})
	}

	fmt.Printf("📁 Target directory: %s\n", targetDir)
	fmt.Printf("📥 Files to download: %d\n", len(downloads))

	// Download all files
	for idx, dl := range downloads {
		destPath := filepath.Join(targetDir, dl.filename)
		fmt.Printf("\n[%d/%d] %s\n", idx+1, len(downloads), dl.filename)
		if err := archive.DownloadFile(destPath, dl.url); err != nil {
			return installResult{}, fmt.Errorf("failed to download %s: %w", dl.filename, err)
		}
	}

	// Run post-install commands if specified
	if len(variant.PostInstall) > 0 {
		fmt.Println("\n🔧 Running post-install commands...")
		for _, cmd := range variant.PostInstall {
			cmd = expandEnvVars(cmd)
			cmd = strings.ReplaceAll(cmd, "${install_path}", targetDir)
			cmd = strings.ReplaceAll(cmd, "%install_path%", targetDir)

			fmt.Printf("   Running: %s\n", cmd)

			execCmd := shellCommand(cmd)
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			if err := execCmd.Run(); err != nil {
				return installResult{}, fmt.Errorf("post-install command failed: %s (error: %w)", cmd, err)
			}
		}
	}

	fmt.Printf("\n✅ Downloaded %d file(s) to %s\n", len(downloads), targetDir)
	return installResult{installPath: targetDir, extractTo: targetDir}, nil
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
	debPath, err := archive.DownloadFileWithProperFilename(downloadURL, i.cacheDir)
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

	// Add repository if specified (for packages not in standard repos)
	if variant.Repository != "" {
		if err := AddAptRepository(variant.Repository, variant.KeyUrl); err != nil {
			return fmt.Errorf("failed to add APT repository: %w", err)
		}
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

// installPacman installs via Pacman package manager (Arch Linux)
func (i *Installer) installPacman(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) error {
	if len(variant.Packages) == 0 {
		return fmt.Errorf("no packages specified for Pacman installation")
	}

	// Install packages via Pacman
	return InstallViaPacman(variant.Packages, variant.RequiresSudo)
}

// installChocolatey installs via Chocolatey package manager (Windows)
func (i *Installer) installChocolatey(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) error {
	if len(variant.Packages) == 0 {
		return fmt.Errorf("no packages specified for Chocolatey installation")
	}

	// Install packages via Chocolatey
	return InstallViaChocolatey(variant.Packages)
}

// installWinget installs via Windows Package Manager (winget)
func (i *Installer) installWinget(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) error {
	if len(variant.Packages) == 0 {
		return fmt.Errorf("no packages specified for Winget installation")
	}

	// Install packages via Winget
	return InstallViaWinget(variant.Packages)
}

// runCommandWithSpinner executes an external command with a brief startup
// spinner followed by direct terminal I/O. Stdin, stdout, and stderr are
// connected directly to the terminal so interactive prompts (e.g. sudo),
// progress bars using \r, and all terminal escape sequences work correctly.
func runCommandWithSpinner(cmd *exec.Cmd, label string) error {
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	// Show brief spinner animation before starting the command
	for i := 0; i < 10; i++ {
		fmt.Printf("\r   %s %s", spinnerChars[i], label)
		time.Sleep(80 * time.Millisecond)
	}
	// Clear spinner line before command takes over
	fmt.Print("\r\033[K")

	// Connect all streams directly to terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("   ❌ %s\n", label)
		return err
	}

	fmt.Printf("   ✅ %s\n", label)
	return nil
}

// truncateCommand shortens a command string for display in spinner label
func truncateCommand(cmd string, maxLen int) string {
	if len(cmd) <= maxLen {
		return cmd
	}
	return cmd[:maxLen-3] + "..."
}

// installScript installs via custom script (for npm-based tools like docusaurus)
// Supports two modes:
// 1. Embedded script: installScript contains a path like "windows/Install-Script.ps1"
// 2. Inline commands: installScript contains shell commands to execute
func (i *Installer) installScript(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) error {
	if len(variant.InstallScript) == 0 {
		return fmt.Errorf("no install script specified")
	}

	// Determine install path
	installPath := options.InstallPath
	if installPath == "" {
		// Use extractTo as install path if available
		if variant.ExtractTo != "" {
			installPath = expandEnvVars(variant.ExtractTo)
		} else {
			installPath = "./site" // Default path
		}
	}

	// Check if first entry is an embedded script path
	firstScript := variant.InstallScript[0]
	if isEmbeddedScript(firstScript) {
		return i.executeEmbeddedScript(firstScript, variant.InstallScriptArgs, installPath, options.DryRun)
	}

	// Fallback to inline command execution; the target only means something
	// for scripts that use ${INSTALL_PATH}
	if usesInstallPath(variant.InstallScript) {
		fmt.Printf("📝 Running install script (target: %s)...\n", installPath)
	} else {
		fmt.Println("📝 Running install script...")
	}

	// Execute each script line
	for _, script := range variant.InstallScript {
		// Replace ${INSTALL_PATH} placeholder
		expandedScript := strings.ReplaceAll(script, "${INSTALL_PATH}", installPath)

		if options.DryRun {
			fmt.Printf("   [DRY-RUN] Would run: %s\n", expandedScript)
			continue
		}

		// Build a short label for the spinner
		spinnerLabel := fmt.Sprintf("Running: %s", truncateCommand(expandedScript, 80))

		if err := runCommandWithSpinner(shellCommand(expandedScript), spinnerLabel); err != nil {
			return fmt.Errorf("script failed: %w", err)
		}
	}

	if options.DryRun {
		return nil
	}

	// A script can exit 0 without installing anything — confirm the result
	// with the package's verification command (issue #201)
	if platform.Verification != nil && strings.TrimSpace(platform.Verification.Command) != "" {
		if err := verifyInstallation(platform.Verification.Command, installPath); err != nil {
			return err
		}
	}

	fmt.Printf("✅ Script installation completed\n")
	return nil
}

// usesInstallPath reports whether any script line references ${INSTALL_PATH}
func usesInstallPath(scripts []string) bool {
	for _, s := range scripts {
		if strings.Contains(s, "${INSTALL_PATH}") {
			return true
		}
	}
	return false
}

// verifyInstallation runs the package's verification command after PATH has
// been refreshed, and fails when the command does not succeed
func verifyInstallation(command, installPath string) error {
	command = strings.ReplaceAll(command, "${INSTALL_PATH}", installPath)
	refreshPath()

	fmt.Printf("🔍 Verifying installation: %s\n", command)
	if installed, _ := runVerification(command); !installed {
		return fmt.Errorf("verification failed: '%s' did not succeed after the install script ran", command)
	}
	fmt.Println("✅ Verification passed")
	return nil
}

// isEmbeddedScript checks if the script path refers to an embedded script file
func isEmbeddedScript(scriptPath string) bool {
	// Embedded scripts are referenced by paths like "windows/Install-Script.ps1"
	if strings.HasSuffix(strings.ToLower(scriptPath), ".ps1") ||
		strings.HasSuffix(strings.ToLower(scriptPath), ".cmd") ||
		strings.HasSuffix(strings.ToLower(scriptPath), ".sh") {
		// Check if it starts with a directory prefix (not a command)
		return strings.Contains(scriptPath, "/") || strings.Contains(scriptPath, "\\")
	}
	return false
}

// executeEmbeddedScript extracts and executes an embedded script
func (i *Installer) executeEmbeddedScript(scriptPath, scriptArgs, installPath string, dryRun bool) error {
	// Normalize path separators
	normalizedPath := strings.ReplaceAll(scriptPath, "\\", "/")

	// Read embedded script
	embeddedPath := filepath.Join("assets", "scripts", normalizedPath)
	embeddedPath = strings.ReplaceAll(embeddedPath, "\\", "/")

	fmt.Printf("📜 Loading embedded script: %s\n", embeddedPath)

	scriptContent, err := EmbeddedScriptsFS.ReadFile(embeddedPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded script %s: %w", embeddedPath, err)
	}

	// Create temp directory for script execution
	tempDir, err := os.MkdirTemp("", "ptx-installer-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write script to temp file
	scriptFileName := filepath.Base(normalizedPath)
	tempScriptPath := filepath.Join(tempDir, scriptFileName)

	if err := os.WriteFile(tempScriptPath, scriptContent, 0755); err != nil {
		return fmt.Errorf("failed to write temp script: %w", err)
	}

	fmt.Printf("📝 Executing script with install path: %s\n", installPath)

	// Expand variables in script args
	expandedArgs := scriptArgs
	expandedArgs = strings.ReplaceAll(expandedArgs, "${install_path}", installPath)
	expandedArgs = strings.ReplaceAll(expandedArgs, "${INSTALL_PATH}", installPath)
	expandedArgs = expandEnvVars(expandedArgs)

	if dryRun {
		fmt.Printf("   [DRY-RUN] Would execute: %s %s\n", tempScriptPath, expandedArgs)
		return nil
	}

	// Execute based on script type
	var cmd *exec.Cmd
	if strings.HasSuffix(strings.ToLower(scriptFileName), ".ps1") {
		// PowerShell script
		psArgs := []string{
			"-ExecutionPolicy", "Bypass",
			"-File", tempScriptPath,
		}
		// Add script arguments if provided
		if expandedArgs != "" {
			// Parse arguments properly for PowerShell
			psArgs = append(psArgs, parseScriptArgs(expandedArgs)...)
		}
		cmd = exec.Command("powershell.exe", psArgs...)
	} else if strings.HasSuffix(strings.ToLower(scriptFileName), ".cmd") {
		// Windows batch script
		cmd = exec.Command("cmd", "/c", tempScriptPath, expandedArgs)
	} else {
		// Unix shell script
		cmd = exec.Command("sh", tempScriptPath, expandedArgs)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = installPath

	fmt.Printf("🔧 Running: %s\n", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("embedded script execution failed: %w", err)
	}

	fmt.Printf("✅ Embedded script completed successfully\n")
	return nil
}

// parseScriptArgs parses script arguments string into individual arguments
// Handles quoted strings and key=value pairs
func parseScriptArgs(args string) []string {
	var result []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range args {
		switch {
		case (r == '"' || r == '\'') && !inQuote:
			inQuote = true
			quoteChar = r
		case r == quoteChar && inQuote:
			inQuote = false
			quoteChar = 0
		case r == ' ' && !inQuote:
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// installWindowsBinary installs Windows binary (MSI/EXE). Returns installResult
// with installPath set to the variant's declared install location (used for
// ${install_path} expansion in PATH_APPEND). The MSI/EXE itself decides the
// real on-disk path; the JSON convention is to mirror it in `installPath`.
func (i *Installer) installWindowsBinary(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) (installResult, error) {
	// Determine download URL (support both single URL and architecture-specific URLs)
	downloadURL := variant.URL
	if downloadURL == "" && len(variant.URLs) > 0 {
		// Select URL based on architecture
		arch := GetArchitecture()
		// Map Go architecture names to common Windows architecture names
		archKey := arch
		if arch == "amd64" {
			archKey = "x64"
		} else if arch == "386" {
			archKey = "x86"
		}

		var ok bool
		downloadURL, ok = variant.URLs[archKey]
		if !ok {
			// Try original arch name
			downloadURL, ok = variant.URLs[arch]
		}
		if !ok {
			return installResult{}, fmt.Errorf("no download URL found for architecture %s", arch)
		}
	}

	if downloadURL == "" {
		return installResult{}, fmt.Errorf("no download URL specified for Windows installation")
	}

	// Download installer to cache
	fmt.Printf("📥 Downloading installer from: %s\n", downloadURL)
	installerPath, err := archive.DownloadFileWithProperFilename(downloadURL, i.cacheDir)
	if err != nil {
		return installResult{}, fmt.Errorf("download failed: %w", err)
	}
	fmt.Printf("✅ Downloaded to: %s\n", installerPath)

	// The installer is only needed for this run (issue #222)
	defer func() {
		if err := os.Remove(installerPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("⚠️  Failed to remove installer %s: %v\n", installerPath, err)
		}
	}()

	if checksum := selectArchChecksum(*variant); checksum != "" {
		if err := verifyFileChecksum(installerPath, checksum); err != nil {
			return installResult{}, err
		}
		fmt.Println("✅ Checksum verified")
	}

	// Run installer based on extension
	installArgs := effectiveInstallArgs(platform, variant)
	var runErr error
	if strings.HasSuffix(strings.ToLower(installerPath), ".msi") {
		runErr = i.runMsiInstaller(installerPath, installArgs)
	} else if strings.HasSuffix(strings.ToLower(installerPath), ".exe") {
		runErr = i.runExeInstaller(installerPath, installArgs)
	} else {
		return installResult{}, fmt.Errorf("unknown installer type: %s", installerPath)
	}
	if runErr != nil {
		return installResult{}, runErr
	}

	// MSI/EXE installers honor their own logic for the on-disk location.
	// Packages mirror that location in variant.installPath so PATH_APPEND can
	// reference it via ${install_path}. Both placeholder forms resolve here.
	resolved := expandEnvVars(variant.InstallPath)
	if resolved != "" {
		fmt.Printf("ℹ️  Open a new shell to pick up PATH changes (%s)\n", filepath.Clean(resolved))
	}
	return installResult{installPath: resolved, extractTo: resolved}, nil
}

// verifyFileChecksum compares the SHA-256 digest of a file with the expected
// value, given as hex with an optional "sha256:" prefix
func verifyFileChecksum(path, expected string) error {
	algorithm, digest, found := strings.Cut(expected, ":")
	if !found {
		algorithm, digest = "sha256", expected
	}
	if !strings.EqualFold(algorithm, "sha256") {
		return fmt.Errorf("unsupported checksum algorithm: %s", algorithm)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s for checksum: %w", path, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to read %s for checksum: %w", path, err)
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actual, strings.TrimSpace(digest)) {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", filepath.Base(path), digest, actual)
	}
	return nil
}

// effectiveInstallArgs returns the variant's installArgs when set and falls
// back to the platform's, so a variant can define its own installer switches
func effectiveInstallArgs(platform *registry.PlatformSpec, variant *registry.VariantSpec) []string {
	if len(variant.InstallArgs) > 0 {
		return variant.InstallArgs
	}
	return platform.InstallArgs
}

// runMsiInstaller runs an MSI installer using msiexec
func (i *Installer) runMsiInstaller(msiPath string, installArgs []string) error {
	fmt.Printf("🔧 Installing MSI package...\n")

	// Build msiexec command
	args := []string{"/i", msiPath}
	if len(installArgs) > 0 {
		args = append(args, installArgs...)
	} else {
		// Default silent install arguments
		args = append(args, "/quiet", "/norestart")
	}

	fmt.Printf("   Running: msiexec %s\n", strings.Join(args, " "))

	cmd := exec.Command("msiexec", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("msiexec failed: %w", err)
	}

	fmt.Println("✅ MSI installation completed")
	return nil
}

// runExeInstaller runs an EXE installer
func (i *Installer) runExeInstaller(exePath string, installArgs []string) error {
	fmt.Printf("🔧 Installing EXE package...\n")

	args := installArgs
	if len(args) == 0 {
		// Default silent install arguments (common patterns)
		args = []string{"/S", "/silent", "/quiet"}
	}

	fmt.Printf("   Running: %s %s\n", exePath, strings.Join(args, " "))

	cmd := exec.Command(exePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installer failed: %w", err)
	}

	fmt.Println("✅ EXE installation completed")
	return nil
}

// GetRegistry returns the package registry
func (i *Installer) GetRegistry() *registry.PackageRegistry {
	return i.registry
}

// GetCacheDir returns the cache directory
func (i *Installer) GetCacheDir() string {
	return i.cacheDir
}

// autoDetectVariant automatically detects the best variant based on system package manager
func (i *Installer) autoDetectVariant(platformSpec *registry.PlatformSpec) string {
	return SelectDefaultVariant(platformSpec)
}

// SelectDefaultVariant picks the variant installed on this system when none
// is given, based on the package managers available
func SelectDefaultVariant(platformSpec *registry.PlatformSpec) string {
	return selectDefaultVariant(platformSpec, DetectPackageManager(), DetectWindowsPackageManagers())
}

// selectDefaultVariant picks the variant installed when none is given:
// the variant of the detected package manager, then (for packages installed
// primarily through a Windows package manager) the variant of the first
// available Windows package manager or a direct installer, then a variant
// marked preferred, then "default" / "standard", otherwise the first variant
// declared in the manifest
func selectDefaultVariant(platformSpec *registry.PlatformSpec, detectedPM string, windowsPMs []string) string {
	// Map system package managers to variant names
	pmToVariant := map[string]string{
		"apt-get": "apt",
		"apt":     "apt",
		"dnf":     "dnf",
		"yum":     "dnf", // yum systems typically use dnf variant
		"pacman":  "pacman",
		"zypper":  "zypper",
	}

	if detectedPM != "" {
		if variantName, ok := pmToVariant[detectedPM]; ok {
			// Check if this variant exists for the package
			if _, exists := platformSpec.Variants[variantName]; exists {
				return variantName
			}
		}
	}

	names := orderedVariantNames(platformSpec)

	// Packages whose primary install method is a Windows package manager
	// follow the tooling present on the machine (issue #222); packages with a
	// direct default (exe/zip, ...) keep their manifest default
	if windowsPackageManagerBinary(platformSpec.Type) != "" {
		for _, pm := range windowsPMs {
			for _, variantName := range names {
				if variantType(platformSpec, variantName) == pm {
					return variantName
				}
			}
		}
		if direct := firstDirectVariant(platformSpec); direct != "" {
			return direct
		}
	}

	// A variant marked preferred in the manifest wins over the name heuristics
	for _, variantName := range names {
		if platformSpec.Variants[variantName].Preferred {
			return variantName
		}
	}

	if _, exists := platformSpec.Variants["default"]; exists {
		return "default"
	}
	if _, exists := platformSpec.Variants["standard"]; exists {
		return "standard"
	}

	// Last resort: the first variant the manifest declares
	if len(names) > 0 {
		return names[0]
	}

	return ""
}

// directInstallTypes are installation types that need no package manager
var directInstallTypes = map[string]bool{
	"exe":      true,
	"msi":      true,
	"zip":      true,
	"tar.gz":   true,
	"download": true,
}

// variantType returns the effective installation type of a variant
func variantType(platformSpec *registry.PlatformSpec, variantName string) string {
	if t := platformSpec.Variants[variantName].Type; t != "" {
		return t
	}
	return platformSpec.Type
}

// firstDirectVariant returns the first declared variant installed without a
// package manager, or "" when there is none
func firstDirectVariant(platformSpec *registry.PlatformSpec) string {
	for _, variantName := range orderedVariantNames(platformSpec) {
		if directInstallTypes[variantType(platformSpec, variantName)] {
			return variantName
		}
	}
	return ""
}

// packageManagerFallback returns the direct variant to install instead of
// variantName when that variant needs a Windows package manager whose binary
// is not available, together with the missing binary; it returns "" when no
// fallback is needed or none exists
func packageManagerFallback(platformSpec *registry.PlatformSpec, variantName string, available func(string) bool) (string, string) {
	binary := windowsPackageManagerBinary(variantType(platformSpec, variantName))
	if binary == "" || available(binary) {
		return "", ""
	}
	return firstDirectVariant(platformSpec), binary
}

// orderedVariantNames returns the variant names in manifest declaration
// order; specs built in code without that order fall back to sorted names
func orderedVariantNames(platformSpec *registry.PlatformSpec) []string {
	names := make([]string, 0, len(platformSpec.Variants))
	seen := make(map[string]bool, len(platformSpec.Variants))
	for _, variantName := range platformSpec.VariantOrder {
		if _, exists := platformSpec.Variants[variantName]; exists && !seen[variantName] {
			names = append(names, variantName)
			seen[variantName] = true
		}
	}

	// Variants missing from VariantOrder, in stable order
	var rest []string
	for variantName := range platformSpec.Variants {
		if !seen[variantName] {
			rest = append(rest, variantName)
		}
	}
	sort.Strings(rest)
	return append(names, rest...)
}

// UnknownVariantError builds the error returned when a user passes a
// --variant/--method that does not match any defined variant for the package
// on the current platform. Lists the available alternatives so the user can
// fix the command without re-running with --list-variants.
func UnknownVariantError(packageName, variant, currentOS string, variants map[string]registry.VariantSpec) error {
	available := make([]string, 0, len(variants))
	for n := range variants {
		available = append(available, n)
	}
	sort.Strings(available)
	return fmt.Errorf(
		"variant %q not found for package %s on %s\n  available variants: %s\n  hint: run 'portunix install %s --list-variants' for details",
		variant, packageName, currentOS,
		strings.Join(available, ", "), packageName,
	)
}
