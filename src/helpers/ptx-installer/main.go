package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"portunix.ai/portunix/src/helpers/ptx-installer/engine"
	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

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

	// Install command implementation
	if len(args) == 0 {
		showInstallHelp()
		return
	}

	// Parse arguments
	packageName := args[0]
	options := &engine.InstallOptions{
		PackageName: packageName,
		DryRun:      false,
		Force:       false,
	}

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "--variant=") {
			options.Variant = strings.TrimPrefix(arg, "--variant=")
		} else if arg == "--dry-run" {
			options.DryRun = true
		} else if arg == "--force" {
			options.Force = true
		}
	}

	// Create installer
	installer, err := engine.NewInstaller("./assets")
	if err != nil {
		fmt.Printf("❌ Error creating installer: %v\n", err)
		return
	}

	// Perform installation
	if err := installer.Install(options); err != nil {
		fmt.Printf("\n❌ Installation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Installation completed successfully!")
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
	default:
		fmt.Printf("Unknown package subcommand: %s\n", subcommand)
		fmt.Println("Use 'portunix package --help' for available subcommands")
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
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help   Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix package list")
	fmt.Println("  portunix package list --category development/languages")
	fmt.Println("  portunix package search python")
	fmt.Println("  portunix package info nodejs")
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

func showInstallHelp() {
	fmt.Println("Install software packages")
	fmt.Println("\nUsage: portunix install <package> [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  --variant=<variant>  Select package variant (e.g., --variant=21 for Java 21)")
	fmt.Println("  --dry-run            Preview installation without executing")
	fmt.Println("  --force              Force reinstallation even if already installed")
	fmt.Println("  -h, --help           Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  portunix install python")
	fmt.Println("  portunix install java --variant=21")
	fmt.Println("  portunix install nodejs --dry-run")
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
