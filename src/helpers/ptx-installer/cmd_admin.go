/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"os"
	"strings"

	"portunix.ai/portunix/src/pkg/installapt"
	"portunix.ai/portunix/src/pkg/installchocolatey"
	"portunix.ai/portunix/src/pkg/installiso"
)

// handleInstallApt handles `ptx-installer install apt <subcommand> [args]`.
// Migrated from src/cmd/install_apt.go in Issue #186d. Uses manual arg
// parsing to stay consistent with the rest of ptx-installer's dispatcher
// surface.
func handleInstallApt(args []string) {
	if len(args) == 0 {
		showAptHelp()
		return
	}
	sub := args[0]
	rest := args[1:]

	switch sub {
	case "--help", "-h", "help":
		showAptHelp()
	case "install":
		aptInstall(rest)
	case "remove":
		aptRemove(rest, false)
	case "purge":
		aptRemove(rest, true)
	case "search":
		aptSearch(rest)
	case "list-installed":
		aptListInstalled()
	case "upgrade":
		aptUpgrade(rest)
	case "clean":
		aptClean(rest)
	case "repo":
		aptRepo(rest)
	default:
		fmt.Printf("Unknown apt subcommand: %s\n", sub)
		showAptHelp()
		os.Exit(1)
	}
}

func showAptHelp() {
	fmt.Println(`Usage: portunix install apt <subcommand> [args] [flags]

Subcommands:
  install <pkg...>        Install packages
  remove  <pkg...>        Remove packages
  purge   <pkg...>        Remove packages and their configuration
  search  <pattern>       Search for packages
  list-installed          List installed packages
  upgrade [pkg...]        Upgrade all or selected packages
  clean                   Clean APT cache
  repo add <uri> <dist> <components...> [--gpg-url URL] [--gpg-key KEY]
  repo remove <distribution>

Common flags:
  --dry-run               Show what would be done without executing`)
}

func newAptMgrOrExit() *installapt.AptManager {
	mgr := installapt.NewAptManager()
	if !mgr.IsSupported() {
		fmt.Println("Error: APT is not supported on this system")
		os.Exit(1)
	}
	return mgr
}

func consumeDryRun(args []string) (rest []string, dryRun bool) {
	rest = args[:0]
	for _, a := range args {
		if a == "--dry-run" {
			dryRun = true
			continue
		}
		rest = append(rest, a)
	}
	return rest, dryRun
}

func aptInstall(args []string) {
	pkgs, dryRun := consumeDryRun(args)
	if len(pkgs) == 0 {
		fmt.Println("Error: at least one package name is required")
		os.Exit(1)
	}
	mgr := newAptMgrOrExit()
	mgr.DryRun = dryRun
	if err := mgr.Update(); err != nil {
		fmt.Printf("Warning: Failed to update package list: %v\n", err)
	}
	if err := mgr.Install(pkgs); err != nil {
		fmt.Printf("Error: Failed to install packages: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Package installation completed successfully")
}

func aptRemove(args []string, purge bool) {
	pkgs, dryRun := consumeDryRun(args)
	// Parse --purge flag if user typed remove --purge instead of purge.
	cleaned := pkgs[:0]
	for _, a := range pkgs {
		if a == "--purge" {
			purge = true
			continue
		}
		cleaned = append(cleaned, a)
	}
	pkgs = cleaned
	if len(pkgs) == 0 {
		fmt.Println("Error: at least one package name is required")
		os.Exit(1)
	}
	mgr := newAptMgrOrExit()
	mgr.DryRun = dryRun
	var err error
	if purge {
		err = mgr.Purge(pkgs)
	} else {
		err = mgr.Remove(pkgs)
	}
	if err != nil {
		fmt.Printf("Error: Failed to remove packages: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Package removal completed successfully")
}

func aptSearch(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: search pattern is required")
		os.Exit(1)
	}
	mgr := newAptMgrOrExit()
	pkgs, err := mgr.Search(args[0])
	if err != nil {
		fmt.Printf("Error: Search failed: %v\n", err)
		os.Exit(1)
	}
	if len(pkgs) == 0 {
		fmt.Printf("No packages found matching '%s'\n", args[0])
		return
	}
	fmt.Printf("Found %d package(s) matching '%s':\n\n", len(pkgs), args[0])
	for _, p := range pkgs {
		status := "not installed"
		if p.Installed {
			status = "installed"
		}
		fmt.Printf("📦 %s (%s)\n", p.Name, status)
		if p.Description != "" {
			fmt.Printf("   %s\n", p.Description)
		}
		fmt.Println()
	}
}

func aptListInstalled() {
	mgr := newAptMgrOrExit()
	pkgs, err := mgr.ListInstalled()
	if err != nil {
		fmt.Printf("Error: Failed to list packages: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Installed packages (%d total):\n\n", len(pkgs))
	for _, p := range pkgs {
		fmt.Printf("📦 %s", p.Name)
		if p.Version != "" {
			fmt.Printf(" (v%s)", p.Version)
		}
		fmt.Println()
		if p.Description != "" {
			fmt.Printf("   %s\n", p.Description)
		}
	}
}

func aptUpgrade(args []string) {
	pkgs, dryRun := consumeDryRun(args)
	mgr := newAptMgrOrExit()
	mgr.DryRun = dryRun
	if err := mgr.Update(); err != nil {
		fmt.Printf("Warning: Failed to update package list: %v\n", err)
	}
	if err := mgr.Upgrade(pkgs); err != nil {
		fmt.Printf("Error: Failed to upgrade packages: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Package upgrade completed successfully")
}

func aptClean(args []string) {
	_, dryRun := consumeDryRun(args)
	mgr := newAptMgrOrExit()
	mgr.DryRun = dryRun
	if err := mgr.Clean(); err != nil {
		fmt.Printf("Error: Failed to clean: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ APT cache cleaned successfully")
}

func aptRepo(args []string) {
	if len(args) == 0 {
		fmt.Println(`Usage: portunix install apt repo <add|remove> ...`)
		return
	}
	switch args[0] {
	case "add":
		aptRepoAdd(args[1:])
	case "remove":
		aptRepoRemove(args[1:])
	default:
		fmt.Printf("Unknown apt repo subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func aptRepoAdd(args []string) {
	// Strip flags: --dry-run, --gpg-url <URL>, --gpg-key <KEY>
	var positional []string
	var dryRun bool
	var gpgURL, gpgKey string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--dry-run":
			dryRun = true
		case args[i] == "--gpg-url" && i+1 < len(args):
			gpgURL = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--gpg-url="):
			gpgURL = strings.TrimPrefix(args[i], "--gpg-url=")
		case args[i] == "--gpg-key" && i+1 < len(args):
			gpgKey = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--gpg-key="):
			gpgKey = strings.TrimPrefix(args[i], "--gpg-key=")
		default:
			positional = append(positional, args[i])
		}
	}
	if len(positional) < 3 {
		fmt.Println("Error: usage: install apt repo add <uri> <distribution> <component...>")
		os.Exit(1)
	}
	mgr := newAptMgrOrExit()
	mgr.DryRun = dryRun
	repo := installapt.Repository{
		URI:          positional[0],
		Distribution: positional[1],
		Components:   positional[2:],
		GPGKeyURL:    gpgURL,
		GPGKey:       gpgKey,
	}
	if err := mgr.AddRepository(repo); err != nil {
		fmt.Printf("Error: Failed to add repository: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Repository added successfully")
}

func aptRepoRemove(args []string) {
	rest, dryRun := consumeDryRun(args)
	if len(rest) == 0 {
		fmt.Println("Error: distribution name is required")
		os.Exit(1)
	}
	mgr := newAptMgrOrExit()
	mgr.DryRun = dryRun
	if err := mgr.RemoveRepository(rest[0]); err != nil {
		fmt.Printf("Error: Failed to remove repository: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Repository removed successfully")
}

// handleInstallIso handles `ptx-installer install iso <os-type> [--variant=...] [--output=...]`.
// Migrated from src/cmd/install_iso.go in Issue #186d.
func handleInstallIso(args []string) {
	osType := "windows11"
	variant := "latest"
	outputDir := ""

	positional := args[:0]
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--help" || args[i] == "-h":
			showIsoHelp()
			return
		case strings.HasPrefix(args[i], "--variant="):
			variant = strings.TrimPrefix(args[i], "--variant=")
		case args[i] == "--variant" && i+1 < len(args):
			variant = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--output="):
			outputDir = strings.TrimPrefix(args[i], "--output=")
		case args[i] == "--output" || args[i] == "-o":
			if i+1 < len(args) {
				outputDir = args[i+1]
				i++
			}
		default:
			positional = append(positional, args[i])
		}
	}
	if len(positional) > 0 {
		osType = positional[0]
	}

	fmt.Printf("Downloading %s ISO (variant: %s)...\n", osType, variant)
	installer := &installiso.ISOInstaller{
		OSType:    osType,
		Variant:   variant,
		OutputDir: outputDir,
	}
	isoPath, err := installer.Download()
	if err != nil {
		fmt.Printf("Error downloading ISO: %v\n", err)
		if osType == "windows11" || osType == "windows10" {
			fmt.Println("\n📋 Manual Download Instructions:")
			fmt.Printf("1. Visit: https://www.microsoft.com/software-download/%s\n", osType)
			fmt.Println("2. Download the Disk Image (ISO) for your language")
			fmt.Printf("3. Save it to: %s\n", outputDir)
			fmt.Println("\nAlternatively, use the Media Creation Tool:")
			fmt.Printf("  portunix install iso %s --variant media-tool\n", osType)
		}
		os.Exit(1)
	}
	fmt.Printf("\n✅ ISO downloaded successfully: %s\n", isoPath)
	fmt.Printf("\nTo create VM:\n  portunix vm create %s-vm --iso %s\n", osType, isoPath)
}

func showIsoHelp() {
	fmt.Println(`Usage: portunix install iso [os-type] [--variant=NAME] [--output=DIR]

Available OS types:
  windows11 (default)  Windows 11 latest version
  windows10            Windows 10 latest version
  ubuntu               Ubuntu LTS versions
  debian               Debian stable
  fedora               Fedora Workstation

Flags:
  --variant=NAME       ISO variant (e.g., latest, 22.04, media-tool)
  --output=DIR, -o DIR Output directory for ISO file`)
}

// handleInstallChocolatey handles `ptx-installer install chocolatey <subcommand>`.
// Migrated from src/cmd/install_chocolatey.go in Issue #186d.
func handleInstallChocolatey(args []string) {
	if len(args) == 0 {
		showChocolateyHelp()
		return
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "--help", "-h", "help":
		showChocolateyHelp()
	case "install":
		chocoInstall(rest)
	case "uninstall":
		chocoUninstall(rest)
	case "search":
		chocoSearch(rest)
	case "list", "list-installed":
		chocoList()
	case "upgrade":
		chocoUpgrade(rest)
	case "info":
		chocoInfo(rest)
	case "self-install":
		chocoSelfInstall()
	default:
		fmt.Printf("Unknown chocolatey subcommand: %s\n", sub)
		showChocolateyHelp()
		os.Exit(1)
	}
}

func showChocolateyHelp() {
	fmt.Println(`Usage: portunix install chocolatey <subcommand> [args] [flags]

Subcommands:
  install <pkg...>      Install packages
  uninstall <pkg...>    Uninstall packages
  search <pattern>      Search for packages
  list                  List installed packages
  upgrade [pkg...]      Upgrade all or selected packages
  info <pkg>            Show package information
  self-install          Install Chocolatey itself

Common flags:
  --dry-run             Show what would be done without executing`)
}

func newChocoMgrOrExit() *installchocolatey.ChocolateyManager {
	mgr := installchocolatey.NewChocolateyManager()
	if !mgr.IsSupported() {
		fmt.Println("Error: Chocolatey is not supported on this system (Windows + choco binary required)")
		os.Exit(1)
	}
	return mgr
}

func chocoInstall(args []string) {
	pkgs, dryRun := consumeDryRun(args)
	if len(pkgs) == 0 {
		fmt.Println("Error: at least one package name is required")
		os.Exit(1)
	}
	mgr := newChocoMgrOrExit()
	mgr.DryRun = dryRun
	if err := mgr.Install(pkgs); err != nil {
		fmt.Printf("Error: Failed to install packages: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Package installation completed successfully")
}

func chocoUninstall(args []string) {
	pkgs, dryRun := consumeDryRun(args)
	if len(pkgs) == 0 {
		fmt.Println("Error: at least one package name is required")
		os.Exit(1)
	}
	mgr := newChocoMgrOrExit()
	mgr.DryRun = dryRun
	if err := mgr.Uninstall(pkgs); err != nil {
		fmt.Printf("Error: Failed to uninstall packages: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Package uninstallation completed successfully")
}

func chocoSearch(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: search pattern is required")
		os.Exit(1)
	}
	mgr := newChocoMgrOrExit()
	pkgs, err := mgr.Search(args[0])
	if err != nil {
		fmt.Printf("Error: Search failed: %v\n", err)
		os.Exit(1)
	}
	if len(pkgs) == 0 {
		fmt.Printf("No packages found matching '%s'\n", args[0])
		return
	}
	fmt.Printf("Found %d package(s) matching '%s':\n\n", len(pkgs), args[0])
	for _, p := range pkgs {
		fmt.Printf("📦 %s", p.Name)
		if p.Version != "" {
			fmt.Printf(" (v%s)", p.Version)
		}
		fmt.Println()
		if p.Description != "" {
			fmt.Printf("   %s\n", p.Description)
		}
	}
}

func chocoList() {
	mgr := newChocoMgrOrExit()
	pkgs, err := mgr.ListInstalled()
	if err != nil {
		fmt.Printf("Error: Failed to list packages: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Installed packages (%d total):\n\n", len(pkgs))
	for _, p := range pkgs {
		fmt.Printf("📦 %s", p.Name)
		if p.Version != "" {
			fmt.Printf(" (v%s)", p.Version)
		}
		fmt.Println()
	}
}

func chocoUpgrade(args []string) {
	pkgs, dryRun := consumeDryRun(args)
	mgr := newChocoMgrOrExit()
	mgr.DryRun = dryRun
	if err := mgr.Upgrade(pkgs); err != nil {
		fmt.Printf("Error: Failed to upgrade packages: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Package upgrade completed successfully")
}

func chocoInfo(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: package name is required")
		os.Exit(1)
	}
	mgr := newChocoMgrOrExit()
	info, err := mgr.GetPackageInfo(args[0])
	if err != nil {
		fmt.Printf("Error: Failed to get package info: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Name:        %s\n", info.Name)
	fmt.Printf("Version:     %s\n", info.Version)
	if info.Description != "" {
		fmt.Printf("Description: %s\n", info.Description)
	}
}

func chocoSelfInstall() {
	mgr := installchocolatey.NewChocolateyManager()
	if err := mgr.InstallChocolatey(); err != nil {
		fmt.Printf("Error: Failed to install Chocolatey: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Chocolatey installed successfully")
}
