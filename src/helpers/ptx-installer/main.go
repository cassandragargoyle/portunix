/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"portunix.ai/portunix/src/helpers/ptx-installer/engine"
	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

func init() {
	// Pass embedded assets to registry package
	registry.SetEmbeddedAssets(embeddedAssets)
	// Pass embedded scripts to engine package
	engine.SetEmbeddedScripts(embeddedScripts)
}

var version = "dev"

// rootCmd represents the base command for ptx-installer
var rootCmd = &cobra.Command{
	Use:   "ptx-installer",
	Short: "Portunix Package Installation Helper",
	Long: `ptx-installer is a helper binary for Portunix that handles all package installation operations.
It provides package installation, management, and registry functionality.

This binary is typically invoked by the main portunix dispatcher and should not be used directly.`,
	Version:            version,
	DisableFlagParsing: true, // Allow passing all flags to subcommands
	Run: func(cmd *cobra.Command, args []string) {
		// Handle the dispatched command directly
		handleCommand(args)
	},
}

// handleCommand dispatches commands routed to this helper by the parent portunix
// binary (see src/dispatcher/dispatcher.go): "install" and "package". args arrive
// stripped of the binary name, so args[0] is the top-level command. Also handles
// the --version / -v meta-flag used by the dispatcher for version discovery.
func handleCommand(args []string) {
	// Handle dispatched commands: install, package
	if len(args) == 0 {
		fmt.Println("No command specified")
		fmt.Println("Usage: ptx-installer [command] [arguments]")
		fmt.Println("\nAvailable commands:")
		fmt.Println("  install  - Install software packages")
		fmt.Println("  package  - Package management operations")
		fmt.Println("  --help   - Show this help")
		return
	}

	command := args[0]

	// Handle version flag specially
	if command == "--version" || command == "-v" {
		fmt.Printf("ptx-installer version %s\n", version)
		return
	}

	subArgs := args[1:]

	switch command {
	case "install":
		handleInstall(subArgs)
	case "package":
		handlePackage(subArgs)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Use 'ptx-installer --help' for available commands")
	}
}

func handleInstall(args []string) {
	// Check for help flag first (before processing any arguments as package names)
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showInstallHelp()
			return
		}
	}

	// Issue #035: `portunix install --recommend-ai` lists the recommended AI
	// assistants for this system and offers to install the missing ones. The
	// flag is detected before args[0] is treated as a package name.
	for _, arg := range args {
		if arg == "--recommend-ai" {
			handleRecommendAI(args)
			return
		}
	}

	// Install command implementation
	if len(args) == 0 {
		showInstallHelp()
		return
	}

	// Parse arguments
	packageName := args[0]
	dryRun := false
	dataRoot := ""
	nonInteractive := false
	versionSel := "" // issue #035: --version selector resolved to a variant below

	// Issue #186d: admin subcommands previously living in
	// src/cmd/install_{apt,iso,chocolatey}.go were migrated into the helper.
	// They use their own argument-parsing routines in cmd_admin.go and must
	// be dispatched BEFORE the generic package-install logic below.
	switch strings.ToLower(packageName) {
	case "apt":
		handleInstallApt(args[1:])
		return
	case "iso":
		handleInstallIso(args[1:])
		return
	case "chocolatey", "choco":
		handleInstallChocolatey(args[1:])
		return
	}

	// Short-circuit: list-variants / list-methods don't need the full installer
	for _, arg := range args[1:] {
		if arg == "--list-variants" || arg == "--list-methods" {
			handleListVariants(packageName)
			return
		}
	}

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--dry-run":
			dryRun = true
		case arg == "--yes", arg == "-y":
			nonInteractive = true
		case strings.HasPrefix(arg, "--data-root="):
			dataRoot = strings.TrimPrefix(arg, "--data-root=")
		case arg == "--data-root" && i+1 < len(args):
			dataRoot = args[i+1]
			i++
		}
	}

	// Handle special container runtime packages
	switch strings.ToLower(packageName) {
	case "docker":
		dockerInstaller := engine.NewDockerInstaller(dryRun, dataRoot, nonInteractive)
		if err := dockerInstaller.Install(); err != nil {
			fmt.Printf("\n❌ Docker installation failed: %v\n", err)
			os.Exit(1)
		}
		return
	case "podman":
		podmanInstaller := engine.NewPodmanInstaller(dryRun)
		if err := podmanInstaller.Install(); err != nil {
			fmt.Printf("\n❌ Podman installation failed: %v\n", err)
			os.Exit(1)
		}
		return
	case "wsl":
		wslInstaller := engine.NewWSLInstaller(dryRun)
		if err := wslInstaller.Install(); err != nil {
			fmt.Printf("\n❌ WSL installation failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Standard package installation
	options := &engine.InstallOptions{
		PackageName: packageName,
		DryRun:      dryRun,
		Force:       false,
	}

	// Parse additional flags
	for i := 1; i < len(args); i++ {
		arg := args[i]

		if v, consumed, ok := parseVariantArg(args, i); ok {
			options.Variant = v
			i += consumed
		} else if strings.HasPrefix(arg, "--path=") {
			options.InstallPath = strings.TrimPrefix(arg, "--path=")
		} else if arg == "--path" && i+1 < len(args) {
			options.InstallPath = args[i+1]
			i++ // Skip next argument as it's the path value
		} else if arg == "--force" {
			options.Force = true
		} else if strings.HasPrefix(arg, "--version=") {
			versionSel = strings.TrimPrefix(arg, "--version=")
		} else if arg == "--version" && i+1 < len(args) {
			versionSel = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--db-host=") {
			options.DBHost = strings.TrimPrefix(arg, "--db-host=")
		} else if arg == "--db-host" && i+1 < len(args) {
			options.DBHost = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--db-port=") {
			options.DBPort = strings.TrimPrefix(arg, "--db-port=")
		} else if arg == "--db-port" && i+1 < len(args) {
			options.DBPort = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--db-user=") {
			options.DBUser = strings.TrimPrefix(arg, "--db-user=")
		} else if arg == "--db-user" && i+1 < len(args) {
			options.DBUser = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--db-password=") {
			options.DBPassword = strings.TrimPrefix(arg, "--db-password=")
		} else if arg == "--db-password" && i+1 < len(args) {
			options.DBPassword = args[i+1]
			i++
		}
	}

	// Create installer
	installer, err := engine.NewInstaller("./assets")
	if err != nil {
		fmt.Printf("❌ Error creating installer: %v\n", err)
		return
	}

	// Issue #035: resolve --version to a concrete variant. An explicit
	// --variant/--method always wins; --version is only consulted when no
	// variant was selected.
	if versionSel != "" && options.Variant == "" {
		variant, err := installer.ResolveVersion(packageName, versionSel)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			os.Exit(1)
		}
		options.Variant = variant
	}

	// Perform installation
	if err := installer.Install(options); err != nil {
		fmt.Printf("\n❌ Installation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Installation completed successfully!")
}

// handleRecommendAI implements `portunix install --recommend-ai` (issue #035).
// It detects the installable AI assistants for the current platform, reports
// their install state, and offers to install the missing ones. Recommendations
// are limited to assistants supported on this OS (those with a platform
// verification command). Honours --dry-run and -y/--yes.
func handleRecommendAI(args []string) {
	dryRun := false
	nonInteractive := false
	for _, arg := range args {
		switch arg {
		case "--dry-run":
			dryRun = true
		case "--yes", "-y":
			nonInteractive = true
		}
	}

	reg, err := registry.LoadPackageRegistry("./assets")
	if err != nil {
		fmt.Printf("❌ Error loading package registry: %v\n", err)
		os.Exit(1)
	}

	statuses := engine.DetectAIAssistants(reg, engine.GetOperatingSystem())

	fmt.Println("\n🤖 Recommended AI assistants for your system:")
	fmt.Println("═══════════════════════════════════════════════════════════")

	var missing []string
	for _, s := range statuses {
		// Assistants without a platform verification command are not
		// installable/detectable on this OS — skip them in recommendations.
		if s.VerifyCommand == "" {
			continue
		}
		marker := "⬜ available"
		if s.Installed {
			marker = "✅ installed"
			if s.Version != "" {
				marker += " (" + s.Version + ")"
			}
		} else {
			missing = append(missing, s.Name)
		}
		fmt.Printf("  %-16s %-24s %s\n", s.Name, s.DisplayName, marker)
	}

	if len(missing) == 0 {
		fmt.Println("\n✅ All recommended AI assistants are already installed.")
		return
	}

	if dryRun {
		fmt.Printf("\n🔍 DRY RUN: would install %s\n", strings.Join(missing, ", "))
		return
	}

	install := nonInteractive
	if !nonInteractive {
		fmt.Printf("\nInstall missing assistant(s) [%s]? [Y/n]: ", strings.Join(missing, ", "))
		reader := bufio.NewReader(os.Stdin)
		ans, _ := reader.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))
		install = ans == "" || ans == "y" || ans == "yes"
	}
	if !install {
		fmt.Println("Skipping installation.")
		return
	}

	installer, err := engine.NewInstaller("./assets")
	if err != nil {
		fmt.Printf("❌ Error creating installer: %v\n", err)
		os.Exit(1)
	}

	failed := 0
	for _, name := range missing {
		if err := installer.Install(&engine.InstallOptions{PackageName: name}); err != nil {
			fmt.Printf("\n❌ %s installation failed: %v\n", name, err)
			failed++
		}
	}
	if failed > 0 {
		fmt.Printf("\n⚠️  %d of %d assistant(s) failed to install.\n", failed, len(missing))
		os.Exit(1)
	}
	fmt.Println("\n✅ All recommended AI assistants are ready.")
}

func handlePackage(args []string) {
	// Check for help flag first
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showPackageHelp()
			return
		}
	}

	// Package management commands
	if len(args) == 0 {
		showPackageHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "list":
		handlePackageList(subArgs)
	case "search":
		handlePackageSearch(subArgs)
	case "info":
		handlePackageInfo(subArgs)
	case "detect":
		handlePackageDetect(subArgs)
	case "update":
		handlePackageUpdate(subArgs)
	default:
		fmt.Printf("Unknown package subcommand: %s\n", subcommand)
		fmt.Println("Use 'portunix package --help' for available subcommands")
	}
}

// handlePackageUpdate implements `portunix package update <package>` (issue
// #035). It reinstalls the package's auto-detected (preferred) variant with
// Force=true, which pulls the latest available version for packages whose
// install method tracks upstream (npm, official "latest" download URLs). This
// is intentionally distinct from `portunix update`, which self-updates the
// portunix binary. Honours --dry-run.
func handlePackageUpdate(args []string) {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Println("Usage: portunix package update <package> [--dry-run]")
			fmt.Println("\nReinstalls a package's latest available version (Force).")
			fmt.Println("Useful for AI assistants (claude-code, claude-desktop, gemini-cli)")
			fmt.Println("and other packages that track upstream releases.")
			fmt.Println("\nExamples:")
			fmt.Println("  portunix package update claude-desktop")
			fmt.Println("  portunix package update gemini-cli --dry-run")
			return
		}
	}

	if len(args) == 0 {
		fmt.Println("❌ No package specified")
		fmt.Println("Usage: portunix package update <package> [--dry-run]")
		os.Exit(1)
	}

	packageName := args[0]
	dryRun := false
	for _, arg := range args[1:] {
		if arg == "--dry-run" {
			dryRun = true
		}
	}

	installer, err := engine.NewInstaller("./assets")
	if err != nil {
		fmt.Printf("❌ Error creating installer: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔄 Updating %s to the latest available version...\n", packageName)
	if err := installer.Install(&engine.InstallOptions{
		PackageName: packageName,
		Force:       true,
		DryRun:      dryRun,
	}); err != nil {
		fmt.Printf("\n❌ Update failed: %v\n", err)
		os.Exit(1)
	}

	if !dryRun {
		fmt.Printf("\n✅ %s updated successfully!\n", packageName)
	}
}

func showPackageHelp() {
	fmt.Println("Package management and registry operations")
	fmt.Println()
	fmt.Println("Usage: portunix package <subcommand> [options]")
	fmt.Println()
	fmt.Println("Available subcommands:")
	fmt.Println("  list     List all available packages")
	fmt.Println("  search   Search for packages by name or description")
	fmt.Println("  info     Show detailed information about a package")
	fmt.Println("  detect   Detect installed AI assistants")
	fmt.Println("  update   Reinstall a package's latest available version")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help   Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix package list")
	fmt.Println("  portunix package list --category development/languages")
	fmt.Println("  portunix package search python")
	fmt.Println("  portunix package info nodejs")
	fmt.Println("  portunix package detect")
	fmt.Println("  portunix package detect --json")
}

func handlePackageList(args []string) {
	// Parse arguments
	var categoryFilter string
	var platformFilter string
	var formatJSON bool

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--category=") {
			categoryFilter = strings.TrimPrefix(arg, "--category=")
		} else if arg == "--category" && i+1 < len(args) {
			categoryFilter = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--platform=") {
			platformFilter = strings.TrimPrefix(arg, "--platform=")
		} else if arg == "--platform" && i+1 < len(args) {
			platformFilter = args[i+1]
			i++
		} else if arg == "--format=json" || arg == "--json" {
			formatJSON = true
		} else if arg == "--help" || arg == "-h" {
			fmt.Println("Usage: portunix package list [options]")
			fmt.Println("\nOptions:")
			fmt.Println("  --category <name>   Filter by category")
			fmt.Println("  --platform <name>   Filter by platform (windows, linux, darwin)")
			fmt.Println("  --format json       Output in JSON format")
			fmt.Println("  --help, -h          Show this help")
			fmt.Println("\nExamples:")
			fmt.Println("  portunix package list")
			fmt.Println("  portunix package list --category development/tools")
			fmt.Println("  portunix package list --platform linux")
			fmt.Println("  portunix package list --format json")
			return
		}
	}

	// Load package registry
	assetsPath := "./assets"
	reg, err := registry.LoadPackageRegistry(assetsPath)
	if err != nil {
		fmt.Printf("Error loading package registry from %s: %v\n", assetsPath, err)
		return
	}

	// Get packages
	allPackages := reg.GetAllPackages()
	if len(allPackages) == 0 {
		fmt.Println("No packages found in registry")
		return
	}

	// Filter packages
	packages := make(map[string]*registry.Package)
	for name, pkg := range allPackages {
		// Category filter
		if categoryFilter != "" && !strings.Contains(strings.ToLower(pkg.Metadata.Category), strings.ToLower(categoryFilter)) {
			continue
		}

		// Platform filter
		if platformFilter != "" {
			found := false
			for platform := range pkg.Spec.Platforms {
				if strings.EqualFold(platform, platformFilter) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		packages[name] = pkg
	}

	if len(packages) == 0 {
		fmt.Println("\n📦 No packages found matching the criteria")
		return
	}

	// Get sorted package names for consistent output order
	sortedNames := make([]string, 0, len(packages))
	for name := range packages {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	// Output in JSON format
	if formatJSON {
		type PackageInfo struct {
			Name        string   `json:"name"`
			DisplayName string   `json:"displayName"`
			Description string   `json:"description"`
			Category    string   `json:"category"`
			Platforms   []string `json:"platforms"`
		}

		packageList := make([]PackageInfo, 0)
		for _, name := range sortedNames {
			pkg := packages[name]
			platforms := make([]string, 0)
			for platform := range pkg.Spec.Platforms {
				platforms = append(platforms, platform)
			}
			sort.Strings(platforms)

			packageList = append(packageList, PackageInfo{
				Name:        pkg.Metadata.Name,
				DisplayName: pkg.Metadata.DisplayName,
				Description: pkg.Metadata.Description,
				Category:    pkg.Metadata.Category,
				Platforms:   platforms,
			})
		}

		jsonData, err := json.MarshalIndent(packageList, "", "  ")
		if err != nil {
			fmt.Printf("Error generating JSON: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
		return
	}

	// Standard output format
	fmt.Println("\n📦 Available Packages:")
	if categoryFilter != "" {
		fmt.Printf("   (Filtered by category: %s)\n", categoryFilter)
	}
	if platformFilter != "" {
		fmt.Printf("   (Filtered by platform: %s)\n", platformFilter)
	}
	fmt.Println("═══════════════════════════════════════════════════════════")

	for _, name := range sortedNames {
		pkg := packages[name]
		fmt.Printf("\n%-20s %s\n", name, pkg.Metadata.DisplayName)
		fmt.Printf("%-20s %s\n", "", pkg.Metadata.Description)
		fmt.Printf("%-20s Category: %s\n", "", pkg.Metadata.Category)
	}

	fmt.Printf("\nTotal packages: %d\n", len(packages))
}

func handlePackageSearch(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: portunix package search <query>")
		fmt.Println("\nSearches in package name, description, and category")
		fmt.Println("\nExamples:")
		fmt.Println("  portunix package search python")
		fmt.Println("  portunix package search \"web server\"")
		fmt.Println("  portunix package search ai")
		return
	}

	query := args[0]

	// Load package registry
	assetsPath := "./assets"
	reg, err := registry.LoadPackageRegistry(assetsPath)
	if err != nil {
		fmt.Printf("Error loading package registry: %v\n", err)
		return
	}

	// Search packages
	matches := reg.SearchPackages(query)

	if len(matches) == 0 {
		fmt.Printf("\n🔍 No packages found matching '%s'\n", query)
		return
	}

	fmt.Printf("\n🔍 Found %d package(s) matching '%s':\n", len(matches), query)
	fmt.Println("═══════════════════════════════════════════════════════════")

	for _, pkg := range matches {
		fmt.Printf("\n%-20s %s\n", pkg.Metadata.Name, pkg.Metadata.DisplayName)
		fmt.Printf("%-20s %s\n", "", pkg.Metadata.Description)
		fmt.Printf("%-20s Category: %s\n", "", pkg.Metadata.Category)

		// Show available platforms
		platforms := make([]string, 0)
		for platform := range pkg.Spec.Platforms {
			platforms = append(platforms, platform)
		}
		if len(platforms) > 0 {
			fmt.Printf("%-20s Platforms: %s\n", "", strings.Join(platforms, ", "))
		}
	}

	fmt.Printf("\nTotal matches: %d\n", len(matches))
}

func handlePackageInfo(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: portunix package info <package>")
		fmt.Println("\nShows detailed information about a package")
		fmt.Println("\nExamples:")
		fmt.Println("  portunix package info python")
		fmt.Println("  portunix package info hugo")
		fmt.Println("  portunix package info nodejs")
		return
	}

	packageName := args[0]

	// Load package registry
	assetsPath := "./assets"
	reg, err := registry.LoadPackageRegistry(assetsPath)
	if err != nil {
		fmt.Printf("Error loading package registry: %v\n", err)
		return
	}

	// Get package
	pkg, err := reg.GetPackage(packageName)
	if err != nil {
		fmt.Printf("\n❌ Package '%s' not found\n", packageName)
		fmt.Println("\nTry: portunix package search <query>")
		return
	}

	// Display package information
	fmt.Printf("\n📦 Package Information: %s\n", pkg.Metadata.Name)
	fmt.Println("═══════════════════════════════════════════════════════════")

	// Metadata
	fmt.Printf("\n🏷️  Metadata:\n")
	fmt.Printf("   Name:         %s\n", pkg.Metadata.Name)
	fmt.Printf("   Display Name: %s\n", pkg.Metadata.DisplayName)
	fmt.Printf("   Description:  %s\n", pkg.Metadata.Description)
	fmt.Printf("   Category:     %s\n", pkg.Metadata.Category)

	if pkg.Metadata.Homepage != "" {
		fmt.Printf("   Homepage:     %s\n", pkg.Metadata.Homepage)
	}
	if pkg.Metadata.Documentation != "" {
		fmt.Printf("   Documentation: %s\n", pkg.Metadata.Documentation)
	}
	if pkg.Metadata.License != "" {
		fmt.Printf("   License:      %s\n", pkg.Metadata.License)
	}
	if pkg.Metadata.Maintainer != "" {
		fmt.Printf("   Maintainer:   %s\n", pkg.Metadata.Maintainer)
	}

	// Dependencies
	if len(pkg.Spec.Dependencies) > 0 {
		fmt.Printf("\n📋 Dependencies:\n")
		for _, dep := range pkg.Spec.Dependencies {
			fmt.Printf("   - %s\n", dep)
		}
	}

	// Platforms and variants
	fmt.Printf("\n💻 Supported Platforms:\n")
	for platformName, platformSpec := range pkg.Spec.Platforms {
		fmt.Printf("\n   %s (type: %s)\n", platformName, platformSpec.Type)

		if len(platformSpec.Variants) > 0 {
			fmt.Printf("   Variants:\n")
			for variantName, variant := range platformSpec.Variants {
				fmt.Printf("     - %s (version: %s)\n", variantName, variant.Version)

				// Show URL if available
				if variant.URL != "" {
					fmt.Printf("       URL: %s\n", variant.URL)
				} else if len(variant.URLs) > 0 {
					fmt.Printf("       URLs:\n")
					for arch, url := range variant.URLs {
						fmt.Printf("         %s: %s\n", arch, url)
					}
				}

				// Show packages for package managers
				if len(variant.Packages) > 0 {
					fmt.Printf("       Packages: %v\n", variant.Packages)
				}
			}
		}
	}

	// AI Prompts (if available)
	if pkg.Spec.AIPrompts != nil && pkg.Spec.AIPrompts.VersionDiscovery != "" {
		fmt.Printf("\n🤖 AI Integration:\n")
		if pkg.Spec.AIPrompts.VersionDiscovery != "" {
			fmt.Printf("   Version Discovery: Available\n")
		}
		if pkg.Spec.AIPrompts.UrlResolution != "" {
			fmt.Printf("   URL Resolution: Available\n")
		}
	}

	fmt.Println()
}

// handlePackageDetect implements `portunix package detect`: it reports which
// installable AI assistants (claude-code, claude-desktop, gemini-cli, ...) are
// present on the current system. Detection runs each package's platform
// verification command via the engine. The --json output is the machine-
// readable basis for the remaining issue #035 features (--recommend-ai, the
// MCP serve init dependency hook).
func handlePackageDetect(args []string) {
	formatJSON := false
	for _, arg := range args {
		switch arg {
		case "--json", "--format=json":
			formatJSON = true
		case "--help", "-h":
			fmt.Println("Usage: portunix package detect [options]")
			fmt.Println("\nDetects which AI assistants are installed on this system")
			fmt.Println("(claude-code, claude-desktop, gemini-cli, ...).")
			fmt.Println("\nOptions:")
			fmt.Println("  --json        Output in JSON format")
			fmt.Println("  --help, -h    Show this help")
			fmt.Println("\nExamples:")
			fmt.Println("  portunix package detect")
			fmt.Println("  portunix package detect --json")
			return
		}
	}

	// Load package registry
	reg, err := registry.LoadPackageRegistry("./assets")
	if err != nil {
		fmt.Printf("Error loading package registry: %v\n", err)
		return
	}

	statuses := engine.DetectAIAssistants(reg, engine.GetOperatingSystem())

	// Output in JSON format
	if formatJSON {
		data, err := json.MarshalIndent(statuses, "", "  ")
		if err != nil {
			fmt.Printf("Error generating JSON: %v\n", err)
			return
		}
		fmt.Println(string(data))
		return
	}

	// Standard output format
	fmt.Println("\n🤖 AI Assistant Detection:")
	fmt.Println("═══════════════════════════════════════════════════════════")

	if len(statuses) == 0 {
		fmt.Println("No AI assistant packages found in registry")
		return
	}

	installed := 0
	for _, s := range statuses {
		marker := "⬜ not found"
		if s.Installed {
			installed++
			marker = "✅ installed"
			if s.Version != "" {
				marker += " (" + s.Version + ")"
			}
		}
		fmt.Printf("\n%-18s %s\n", s.Name, s.DisplayName)
		fmt.Printf("%-18s %s\n", "", marker)
	}

	fmt.Printf("\nDetected %d of %d AI assistant(s) installed.\n", installed, len(statuses))
}

// handleListVariants lists all installation variants (a.k.a. methods) available
// for a package on the current platform. Implements the discovery side of
// issue #079. Output marks the auto-detected variant with `*` and any variant
// that opts in via `"preferred": true` with `(preferred)`.
func handleListVariants(packageName string) {
	reg, err := registry.LoadPackageRegistry("./assets")
	if err != nil {
		fmt.Printf("❌ Error loading package registry: %v\n", err)
		os.Exit(1)
	}

	pkg, err := reg.GetPackage(packageName)
	if err != nil {
		fmt.Printf("❌ Package '%s' not found\n", packageName)
		fmt.Println("\nTry: portunix package search <query>")
		os.Exit(1)
	}

	currentOS := engine.GetOperatingSystem()
	platformSpec, exists := pkg.Spec.Platforms[currentOS]
	if !exists && currentOS == "windows_sandbox" {
		platformSpec, exists = pkg.Spec.Platforms["windows"]
	}
	if !exists {
		fmt.Printf("❌ Package '%s' is not available for platform '%s'\n", packageName, currentOS)
		// Show which platforms ARE supported, so the user has a path forward
		platforms := make([]string, 0, len(pkg.Spec.Platforms))
		for p := range pkg.Spec.Platforms {
			platforms = append(platforms, p)
		}
		sort.Strings(platforms)
		fmt.Printf("Supported platforms: %s\n", strings.Join(platforms, ", "))
		os.Exit(1)
	}

	if len(platformSpec.Variants) == 0 {
		fmt.Printf("Package '%s' has no variants defined for %s\n", packageName, currentOS)
		return
	}

	autoDetected := detectAutoVariant(&platformSpec)
	fmt.Print(formatVariantList(packageName, currentOS, autoDetected, &platformSpec))
}

// formatVariantList renders the human-readable variant list for a package.
// Pure function (no I/O) so the formatting can be unit-tested without spinning
// up the full installer.
func formatVariantList(packageName, currentOS, autoDetected string, platformSpec *registry.PlatformSpec) string {
	var b strings.Builder

	// Sort variant names for stable output
	names := make([]string, 0, len(platformSpec.Variants))
	for n := range platformSpec.Variants {
		names = append(names, n)
	}
	sort.Strings(names)

	fmt.Fprintf(&b, "\n📦 Available variants for '%s' on %s:\n", packageName, currentOS)
	b.WriteString("═══════════════════════════════════════════════════════════\n")
	b.WriteString("Legend:  *  = auto-detected default   (preferred) = marked preferred\n\n")

	for _, name := range names {
		variant := platformSpec.Variants[name]

		marker := " "
		if name == autoDetected {
			marker = "*"
		}

		// Effective installation type: variant override wins over platform default
		effectiveType := platformSpec.Type
		if variant.Type != "" {
			effectiveType = variant.Type
		}

		preferredTag := ""
		if variant.Preferred {
			preferredTag = " (preferred)"
		}

		fmt.Fprintf(&b, "  %s %-20s version: %s   type: %s%s\n", marker, name, variant.Version, effectiveType, preferredTag)
		if variant.Description != "" {
			fmt.Fprintf(&b, "    %s\n", variant.Description)
		}
	}

	b.WriteString("\nInstall with:\n")
	fmt.Fprintf(&b, "  portunix install %s --method=<variant>\n", packageName)
	fmt.Fprintf(&b, "  portunix install %s --variant=<variant>   (equivalent)\n", packageName)
	return b.String()
}

// parseVariantArg recognises the variant-selection flags --variant and its
// alias --method (issue #079). Returns the value, how many extra positional
// args were consumed (1 for the "--flag value" form, 0 for "--flag=value"),
// and whether the arg matched. ok=false means the caller should try other
// flag patterns.
func parseVariantArg(args []string, i int) (value string, consumed int, ok bool) {
	arg := args[i]
	switch {
	case strings.HasPrefix(arg, "--variant="):
		return strings.TrimPrefix(arg, "--variant="), 0, true
	case arg == "--variant" && i+1 < len(args):
		return args[i+1], 1, true
	case strings.HasPrefix(arg, "--method="):
		return strings.TrimPrefix(arg, "--method="), 0, true
	case arg == "--method" && i+1 < len(args):
		return args[i+1], 1, true
	}
	return "", 0, false
}

// detectAutoVariant mirrors engine.Installer.autoDetectVariant for display
// purposes only. Kept lightweight here to avoid plumbing a full Installer just
// to render the variant list — engine logic remains the source of truth at
// install time.
func detectAutoVariant(platformSpec *registry.PlatformSpec) string {
	pmToVariant := map[string]string{
		"apt-get": "apt",
		"apt":     "apt",
		"dnf":     "dnf",
		"yum":     "dnf",
		"pacman":  "pacman",
		"zypper":  "zypper",
	}
	if pm := engine.DetectPackageManager(); pm != "" {
		if v, ok := pmToVariant[pm]; ok {
			if _, exists := platformSpec.Variants[v]; exists {
				return v
			}
		}
	}
	if _, exists := platformSpec.Variants["default"]; exists {
		return "default"
	}
	if _, exists := platformSpec.Variants["standard"]; exists {
		return "standard"
	}
	for n := range platformSpec.Variants {
		return n
	}
	return ""
}

func showInstallHelp() {
	fmt.Println("Install software packages")
	fmt.Println("\nUsage: portunix install <package> [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  --variant=<variant>  Select package variant (e.g., --variant=21 for Java 21)")
	fmt.Println("  --method=<variant>   Alias for --variant (issue #079)")
	fmt.Println("  --version=<version>  Select a specific version; resolved to a variant by name or version")
	fmt.Println("  --recommend-ai       List recommended AI assistants and offer to install missing ones")
	fmt.Println("  --list-variants      List available variants for the package and exit")
	fmt.Println("  --list-methods       Alias for --list-variants")
	fmt.Println("  --path=<path>        Target installation path (for project generators like docusaurus)")
	fmt.Println("  --dry-run            Preview installation without executing")
	fmt.Println("  --force              Force reinstallation even if already installed")
	fmt.Println("  --data-root=<path>   Docker data-root directory (docker only; skips the prompt)")
	fmt.Println("  -y, --yes            Non-interactive mode: accept recommended defaults (docker only)")
	fmt.Println("  --db-host=<host>     Override container DB HOST env (container variants that read it)")
	fmt.Println("  --db-port=<port>     Override container DB PORT env")
	fmt.Println("  --db-user=<user>     Override container DB USER env")
	fmt.Println("  --db-password=<pwd>  Override container DB PASSWORD env")
	fmt.Println("  -h, --help           Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  portunix install python")
	fmt.Println("  portunix install java --variant=21")
	fmt.Println("  portunix install hugo --list-variants")
	fmt.Println("  portunix install hugo --method=snap")
	fmt.Println("  portunix install docusaurus --path ./my-docs")
	fmt.Println("  portunix install nodejs --dry-run")
	fmt.Println("  portunix install java --version 21")
	fmt.Println("  portunix install --recommend-ai")
	fmt.Println("  portunix install docker --data-root D:\\docker-data --yes")
	fmt.Println("  portunix install wsl                        (Windows only — prerequisite for Docker)")
	fmt.Println("  portunix install odoo --variant=container-external-db --db-host=my-pg")
	fmt.Println("\nUse 'portunix package list' to see available packages")
	fmt.Println("Use 'portunix package info <package>' for detailed package information")
}

func init() {
	// NOTE: Embedded assets support to be added later
	// For Phase 2, using external assets loading from filesystem

	// Add version information
	rootCmd.SetVersionTemplate("ptx-installer version {{.Version}}\n")
}

func main() {
	// Initialize embedded assets in registry package
	registry.SetEmbeddedAssets(embeddedAssets)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
