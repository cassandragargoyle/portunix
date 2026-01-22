package engine

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
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

// InstallOptions contains options for package installation
type InstallOptions struct {
	PackageName string
	Variant     string
	InstallPath string // Target path for packages that require it (e.g., docusaurus)
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
		// Auto-detect variant based on system package manager
		variant = i.autoDetectVariant(&platformSpec)
	}

	// Check if variant exists
	variantSpec, exists := platformSpec.Variants[variant]
	if !exists {
		return fmt.Errorf("variant %s not found for package %s", variant, options.PackageName)
	}

	fmt.Printf("🎯 Variant: %s (version: %s)\n", variant, variantSpec.Version)

	// Check if admin/root privileges are required
	if variantSpec.RequiresAdmin && !IsAdmin() {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("❌ This installation requires Administrator privileges.\n   Please run PowerShell as Administrator and try again")
		}
		return fmt.Errorf("❌ This installation requires root privileges.\n   Please run with sudo and try again")
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

	// Perform installation based on effective type (variant type takes precedence)
	fmt.Printf("\n🚀 Starting installation (type: %s)...\n", effectiveType)

	switch effectiveType {
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
	case "pacman":
		return i.installPacman(&platformSpec, &variantSpec, options)
	case "msi", "exe":
		return i.installWindowsBinary(&platformSpec, &variantSpec, options)
	case "chocolatey":
		return i.installChocolatey(&platformSpec, &variantSpec, options)
	case "winget":
		return i.installWinget(&platformSpec, &variantSpec, options)
	case "script":
		return i.installScript(&platformSpec, &variantSpec, options)
	default:
		return fmt.Errorf("installation type %s not yet implemented in ptx-installer", effectiveType)
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
				return fmt.Errorf("failed to create extract directory: %w", err)
			}
			fmt.Printf("📁 Using fallback: %s\n", extractTo)
		} else {
			return fmt.Errorf("failed to create extract directory: %w", err)
		}
	}

	// Check if target directory already exists and is not empty (after determining final path)
	proceed, err := checkExistingDirectory(extractTo, options.Force)
	if err != nil {
		return fmt.Errorf("directory check failed: %w", err)
	}
	if !proceed {
		return nil // User cancelled installation
	}

	// Extract archive
	fmt.Printf("📦 Extracting to: %s\n", extractTo)
	if err := ExtractArchive(archivePath, extractTo); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Find actual root directory (many archives have a single top-level directory)
	actualRoot, err := FindExtractedRoot(extractTo)
	if err != nil {
		fmt.Printf("⚠️  Could not determine extracted root: %v\n", err)
	} else if actualRoot != extractTo {
		fmt.Printf("📁 Detected package root: %s\n", actualRoot)
		extractTo = actualRoot
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

	// Run install script if specified (embedded PowerShell/shell script)
	if len(variant.InstallScript) > 0 {
		firstScript := variant.InstallScript[0]
		if isEmbeddedScript(firstScript) {
			fmt.Println("📜 Running installation script...")
			if err := i.executeEmbeddedScript(firstScript, variant.InstallScriptArgs, extractTo, options.DryRun); err != nil {
				return fmt.Errorf("install script failed: %w", err)
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
			var execCmd *exec.Cmd
			if runtime.GOOS == "windows" {
				execCmd = exec.Command("cmd", "/C", cmd)
			} else {
				execCmd = exec.Command("sh", "-c", cmd)
			}
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			if err := execCmd.Run(); err != nil {
				errMsg := fmt.Sprintf("Command failed: %s (error: %v)", cmd, err)
				fmt.Printf("   ❌ %s\n", errMsg)
				postInstallErrors = append(postInstallErrors, errMsg)
			}
		}

		if len(postInstallErrors) > 0 {
			return fmt.Errorf("❌ Installation failed: %d post-install command(s) failed:\n   - %s",
				len(postInstallErrors), strings.Join(postInstallErrors, "\n   - "))
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

	// Fallback to inline command execution
	fmt.Printf("📝 Running install script (target: %s)...\n", installPath)

	// Execute each script line
	for _, script := range variant.InstallScript {
		// Replace ${INSTALL_PATH} placeholder
		expandedScript := strings.ReplaceAll(script, "${INSTALL_PATH}", installPath)

		if options.DryRun {
			fmt.Printf("   [DRY-RUN] Would run: %s\n", expandedScript)
			continue
		}

		fmt.Printf("   Running: %s\n", expandedScript)

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", expandedScript)
		} else {
			cmd = exec.Command("sh", "-c", expandedScript)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("script failed: %w", err)
		}
	}

	fmt.Printf("✅ Script installation completed\n")
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

// installWindowsBinary installs Windows binary (MSI/EXE)
func (i *Installer) installWindowsBinary(platform *registry.PlatformSpec, variant *registry.VariantSpec, options *InstallOptions) error {
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
			return fmt.Errorf("no download URL found for architecture %s", arch)
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no download URL specified for Windows installation")
	}

	// Download installer to cache
	fmt.Printf("📥 Downloading installer from: %s\n", downloadURL)
	installerPath, err := DownloadFileWithProperFilename(downloadURL, i.cacheDir)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	fmt.Printf("✅ Downloaded to: %s\n", installerPath)

	// Determine installation type and run installer
	if strings.HasSuffix(strings.ToLower(installerPath), ".msi") {
		return i.runMsiInstaller(installerPath, platform.InstallArgs)
	} else if strings.HasSuffix(strings.ToLower(installerPath), ".exe") {
		return i.runExeInstaller(installerPath, platform.InstallArgs)
	}

	return fmt.Errorf("unknown installer type: %s", installerPath)
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
	// Map system package managers to variant names
	pmToVariant := map[string]string{
		"apt-get": "apt",
		"apt":     "apt",
		"dnf":     "dnf",
		"yum":     "dnf", // yum systems typically use dnf variant
		"pacman":  "pacman",
		"zypper":  "zypper",
	}

	// Detect system package manager
	detectedPM := DetectPackageManager()
	if detectedPM != "" {
		if variantName, ok := pmToVariant[detectedPM]; ok {
			// Check if this variant exists for the package
			if _, exists := platformSpec.Variants[variantName]; exists {
				return variantName
			}
		}
	}

	// Fallback: prefer "default" or "standard", otherwise first available
	if _, exists := platformSpec.Variants["default"]; exists {
		return "default"
	}
	if _, exists := platformSpec.Variants["standard"]; exists {
		return "standard"
	}

	// Last resort: first available variant
	for variantName := range platformSpec.Variants {
		return variantName
	}

	return ""
}
