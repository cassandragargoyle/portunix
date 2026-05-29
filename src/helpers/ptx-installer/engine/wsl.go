/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// WSLInstaller handles Windows Subsystem for Linux installation. WSL is
// Windows-only; other platforms refuse the install up front.
type WSLInstaller struct {
	dryRun bool
}

// NewWSLInstaller creates a new WSL installer instance.
func NewWSLInstaller(dryRun bool) *WSLInstaller {
	return &WSLInstaller{dryRun: dryRun}
}

// wslRequiredFeatures are the Windows optional features that must be enabled
// for WSL2 (and Docker Desktop's WSL2 backend) to work.
var wslRequiredFeatures = []string{
	"VirtualMachinePlatform",
	"Microsoft-Windows-Subsystem-Linux",
}

// Install performs WSL2 installation on Windows. Uses PowerShell's
// Enable-WindowsOptionalFeature instead of `wsl --install` because the
// latter is a stub on some Windows 10/early 11 builds that cannot bootstrap
// itself, while the feature API is available everywhere.
func (w *WSLInstaller) Install() error {
	fmt.Println("🐧 Starting WSL2 installation...")

	if runtime.GOOS != "windows" {
		return fmt.Errorf("❌ WSL is Windows-only (current platform: %s)", runtime.GOOS)
	}

	if w.isWSLInstalled() {
		fmt.Println("✅ WSL is already installed")
		return nil
	}

	if err := checkWSLAdminRequired(w.dryRun, IsAdmin()); err != nil {
		return err
	}

	if w.dryRun {
		fmt.Println("\n🔍 DRY RUN - Would perform the following:")
		for i, f := range wslRequiredFeatures {
			fmt.Printf("   %d. Enable Windows feature: %s\n", i+1, f)
		}
		fmt.Println("   N. Prompt user to restart Windows before WSL is usable")
		return nil
	}

	fmt.Println("\n🔧 Enabling Windows features for WSL2...")
	for _, feature := range wslRequiredFeatures {
		fmt.Printf("   Enabling %s ...\n", feature)
		if err := enableWindowsOptionalFeature(feature); err != nil {
			return fmt.Errorf("failed to enable %s: %w", feature, err)
		}
	}

	fmt.Println("\n✅ WSL2 Windows features enabled successfully!")
	fmt.Println("⚠️  A system restart is required to finalize WSL2 setup.")
	fmt.Println("   After restart, run one of:")
	fmt.Println("     portunix install docker      (to install Docker Desktop)")
	fmt.Println("     wsl --set-default-version 2  (to ensure WSL2 is the default)")
	return nil
}

func (w *WSLInstaller) isWSLInstalled() bool {
	return hasWindowsWSL()
}

// enableWindowsOptionalFeature enables a single Windows optional feature via
// PowerShell. Requires an elevated process. `-NoRestart` avoids an auto
// reboot prompt; the caller handles restart messaging.
func enableWindowsOptionalFeature(feature string) error {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		fmt.Sprintf("Enable-WindowsOptionalFeature -Online -FeatureName %s -All -NoRestart -ErrorAction Stop", feature))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// checkWSLAdminRequired mirrors the Docker installer's admin gate: the
// underlying `wsl --install` call requires elevated privileges, but --dry-run
// should still work from a non-elevated shell for plan preview.
func checkWSLAdminRequired(dryRun, isAdmin bool) error {
	if dryRun || isAdmin {
		return nil
	}
	return fmt.Errorf("❌ WSL installation requires Administrator privileges.\n" +
		"   Please run Portunix from an elevated PowerShell or cmd and try again.\n" +
		"   Tip: --dry-run works from a non-elevated shell.")
}
