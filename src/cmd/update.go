/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"portunix.ai/app/update"
	appversion "portunix.ai/app/version"
)

var (
	checkOnly   bool
	forceUpdate bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Portunix to the latest version",
	Long:  `Check for updates and install the latest version of Portunix from GitHub releases`,
	Run: func(cmd *cobra.Command, args []string) {
		currentVersion := appversion.ProductVersion
		fmt.Printf("Current version: %s\n", currentVersion)

		if checkOnly {
			checkForUpdate()
			return
		}

		performUpdate(forceUpdate)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates without installing")
	updateCmd.Flags().BoolVar(&forceUpdate, "force", false, "Force update even if on latest version")
}

func checkForUpdate() {
	fmt.Println("Checking for updates...")

	release, err := update.CheckForUpdate()
	if err != nil {
		fmt.Printf("Error: Unable to check for updates\n  %v\n", err)
		fmt.Println("  Please check your internet connection and try again")
		os.Exit(1)
	}

	if release == nil {
		fmt.Println("✓ You are running the latest version!")
		return
	}

	fmt.Printf("Latest version: %s\n", release.Version)
	fmt.Printf("Update available! Run 'portunix update' to install.\n")
}

func performUpdate(force bool) {
	fmt.Println("Checking for updates...")

	release, err := update.CheckForUpdate()
	if err != nil {
		fmt.Printf("Error: Unable to check for updates\n  %v\n", err)
		fmt.Println("  Please check your internet connection and try again")
		os.Exit(1)
	}

	if release == nil && !force {
		fmt.Println("✓ You are running the latest version!")
		return
	}

	if release == nil && force {
		// Force reinstall current version
		currentVersion := appversion.ProductVersion
		fmt.Printf("⚠ Forcing reinstall of %s...\n", currentVersion)
		release, err = update.GetRelease(currentVersion)
		if err != nil {
			fmt.Printf("Error: Unable to get release information\n  %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("✓ New version available: %s\n", release.Version)
	}

	// Download update
	fmt.Printf("✓ Downloading portunix-%s-%s-%s...\n", release.Version, update.GetOS(), update.GetArch())
	archiveFile, err := update.DownloadUpdate(release)
	if err != nil {
		fmt.Printf("Error: Failed to download update\n  %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(archiveFile)
	// os.Exit skips deferred calls, so failures remove the archive first
	abort := func() {
		_ = os.Remove(archiveFile)
		os.Exit(1)
	}

	// Verify checksum
	fmt.Println("✓ Verifying checksum...")
	checksumURL := release.ChecksumURL
	if checksumURL == "" {
		fmt.Println("⚠ Warning: No checksum available for verification")
	} else {
		err = update.VerifyChecksum(archiveFile, checksumURL)
		if err != nil {
			fmt.Printf("Error: Checksum verification failed\n  %v\n", err)
			fmt.Println("  This could indicate a corrupted download or security issue")
			fmt.Println("  Update aborted for safety")
			abort()
		}
	}

	// Inform user about signature verification status. The actual signature
	// check happens inside DownloadUpdate (VerifyArchiveChecksumSigned); here
	// we only surface the user-visible state.
	if update.HasPublicKey() {
		if release.SignatureURL != "" {
			fmt.Println("✓ Signature verified")
		} else if update.RequireSignature {
			fmt.Println("Error: No signature available for this release")
			fmt.Println("  Signature is required and update was aborted for safety")
			abort()
		} else {
			fmt.Println("⚠ Warning: No signature available for this release (older release)")
		}
	}

	// Apply update
	fmt.Println("✓ Installing update...")
	err = update.ApplyUpdate(archiveFile)
	if err != nil {
		fmt.Printf("Error: Failed to apply update\n  %v\n", err)

		// Check if it's a permission error and provide better guidance
		if update.IsPermissionError(err) {
			fmt.Println("\n💡 Solutions:")
			fmt.Println("  1. Run as Administrator (Right-click cmd.exe -> Run as administrator)")
			fmt.Println("  2. Or download and reinstall manually from GitHub releases")
			fmt.Println("  3. Or move portunix to a user-writable location (like Documents)")
		}

		// ApplyUpdate has already rolled back every binary it replaced
		if errors.Is(err, update.ErrRestoreFailed) {
			fmt.Println("  Some binaries could not be restored, their *.backup files were kept")
			fmt.Println("  Manual intervention may be required")
		} else {
			fmt.Println("✓ Previous version kept")
		}
		abort()
	}

	fmt.Println("✓ Update completed successfully!")
	if !force {
		fmt.Printf("\nPortunix has been updated from %s to %s\n", appversion.ProductVersion, release.Version)
	} else {
		fmt.Printf("\nPortunix %s has been reinstalled successfully\n", release.Version)
	}
}
