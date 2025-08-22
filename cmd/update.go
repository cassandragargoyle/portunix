package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"portunix.cz/app/update"
	appversion "portunix.cz/app/version"
)

var (
	checkOnly bool
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
	tmpFile, err := update.DownloadUpdate(release)
	if err != nil {
		fmt.Printf("Error: Failed to download update\n  %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile)

	// Verify checksum
	fmt.Println("✓ Verifying checksum...")
	checksumURL := release.ChecksumURL
	if checksumURL == "" {
		fmt.Println("⚠ Warning: No checksum available for verification")
	} else {
		err = update.VerifyChecksum(tmpFile, checksumURL)
		if err != nil {
			fmt.Printf("Error: Checksum verification failed\n  %v\n", err)
			fmt.Println("  This could indicate a corrupted download or security issue")
			fmt.Println("  Update aborted for safety")
			os.Exit(1)
		}
	}

	// Create backup
	fmt.Println("✓ Creating backup...")
	backupPath, err := update.CreateBackup()
	if err != nil {
		fmt.Printf("Error: Failed to create backup\n  %v\n", err)
		os.Exit(1)
	}

	// Apply update
	fmt.Println("✓ Installing update...")
	err = update.ApplyUpdate(tmpFile)
	if err != nil {
		fmt.Printf("Error: Failed to apply update\n  %v\n", err)
		
		// Try to restore backup
		fmt.Println("Attempting to restore backup...")
		if restoreErr := update.RestoreBackup(backupPath); restoreErr != nil {
			fmt.Printf("Error: Failed to restore backup\n  %v\n", restoreErr)
			fmt.Println("  Manual intervention may be required")
		} else {
			fmt.Println("✓ Backup restored successfully")
		}
		os.Exit(1)
	}

	// Clean up backup on success
	os.Remove(backupPath)
	
	fmt.Println("✓ Update completed successfully!")
	if !force {
		fmt.Printf("\nPortunix has been updated from %s to %s\n", appversion.ProductVersion, release.Version)
	} else {
		fmt.Printf("\nPortunix %s has been reinstalled successfully\n", release.Version)
	}
}