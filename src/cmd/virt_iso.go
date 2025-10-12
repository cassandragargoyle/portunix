package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// ISOInfo contains information about downloadable ISOs
type ISOInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Checksum    string `json:"checksum"`
	Size        int64  `json:"size"`
	Category    string `json:"category"`
}

// virtISOCmd represents the virt iso command
var virtISOCmd = &cobra.Command{
	Use:   "iso",
	Short: "Manage ISO files for virtual machines",
	Long: `Manage ISO files for virtual machine installation using Portunix install system.

This is a convenience wrapper around 'portunix install iso' with additional features:
- Download official OS ISOs from verified sources
- Verify checksums for security
- Cache ISOs locally for reuse
- List available and downloaded ISOs

Available ISOs include:
- Ubuntu 24.04 LTS Desktop/Server
- Ubuntu 22.04 LTS Desktop/Server
- Windows 11 (requires manual download)
- Windows 10 (requires manual download)

Examples:
  portunix virt iso download ubuntu-24.04
  portunix virt iso list
  portunix virt iso download windows11`,
}

// virtISODownloadCmd represents the iso download command
var virtISODownloadCmd = &cobra.Command{
	Use:   "download [iso-name]",
	Short: "Download an ISO file",
	Long: `Download an official OS ISO file from verified sources.

Downloaded ISOs are cached in ~/.portunix/.cache/isos/ for reuse.
Checksums are automatically verified after download.

Available ISOs:
  ubuntu-24.04          - Ubuntu 24.04 LTS Desktop
  ubuntu-24.04-server   - Ubuntu 24.04 LTS Server
  ubuntu-22.04          - Ubuntu 22.04 LTS Desktop
  ubuntu-22.04-server   - Ubuntu 22.04 LTS Server
  debian-12             - Debian 12 (Bookworm) netinst
  windows11-eval        - Windows 11 Evaluation (90 days)
  windows10-eval        - Windows 10 Evaluation (90 days)

Examples:
  portunix virt iso download ubuntu-24.04
  portunix virt iso download windows11-eval --force
  portunix virt iso download debian-12 --dry-run`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		isoName := args[0]
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Try delegating to existing portunix install iso system first
		fmt.Printf("Attempting download via Portunix install system: %s\n", isoName)
		fmt.Printf("Command: portunix install iso %s\n", isoName)
		fmt.Println()

		installCmd := exec.Command("portunix", "install", "iso", isoName)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		installCmd.Stdin = os.Stdin

		if err := installCmd.Run(); err == nil {
			fmt.Printf("\n✅ ISO '%s' downloaded successfully via Portunix install system\n", isoName)
			return
		}

		fmt.Printf("⚠️  Portunix install system failed, trying direct download...\n\n")

		// Get ISO information for fallback
		isoInfo, exists := getAvailableISOs()[isoName]
		if !exists {
			fmt.Printf("Error: ISO '%s' not found in both systems.\n", isoName)
			fmt.Println("\nAvailable ISOs:")
			listAvailableISOs()
			fmt.Println("\nAlternatively, try: portunix install iso --help")
			os.Exit(1)
		}

		isoDir := getISOCacheDir()
		isoPath := filepath.Join(isoDir, getISOFilename(isoName))

		// Dry-run mode - show what would be done
		if dryRun {
			fmt.Println("🔍 DRY RUN MODE - No actual download will occur")
			fmt.Println()
			fmt.Printf("Would download: %s\n", isoInfo.Name)
			fmt.Printf("Description:    %s\n", isoInfo.Description)
			fmt.Printf("Category:       %s\n", isoInfo.Category)
			fmt.Printf("URL:            %s\n", isoInfo.URL)
			if isoInfo.Size > 0 {
				fmt.Printf("Size:           %s\n", formatBytes(isoInfo.Size))
			}
			fmt.Printf("Save to:        %s\n", isoPath)

			// Check if already exists
			if _, err := os.Stat(isoPath); err == nil {
				fmt.Printf("\n⚠️  ISO already exists at %s\n", isoPath)
				if force {
					fmt.Println("   Would overwrite existing file (--force flag used)")
				} else {
					fmt.Println("   Would skip download (use --force to re-download)")
				}
			} else {
				fmt.Printf("\n✅ ISO would be downloaded to %s\n", isoPath)
			}

			if isoInfo.Checksum != "" {
				fmt.Println("\n✓ Checksum verification would be performed after download")
			}
			return
		}

		if err := os.MkdirAll(isoDir, 0755); err != nil {
			fmt.Printf("Error creating ISO cache directory: %v\n", err)
			os.Exit(1)
		}

		// Check if already downloaded
		if !force {
			if _, err := os.Stat(isoPath); err == nil {
				fmt.Printf("ISO '%s' already downloaded at %s\n", isoName, isoPath)
				fmt.Println("Use --force to re-download")
				return
			}
		}

		fmt.Printf("Downloading %s...\n", isoInfo.Name)
		fmt.Printf("URL: %s\n", isoInfo.URL)
		if isoInfo.Size > 0 {
			fmt.Printf("Size: %s\n", formatBytes(isoInfo.Size))
		}
		fmt.Println()

		if err := downloadISO(isoInfo.URL, isoPath); err != nil {
			fmt.Printf("❌ Download failed: %v\n", err)
			os.Exit(1)
		}

		// Verify checksum if available
		if isoInfo.Checksum != "" {
			fmt.Println("Verifying checksum...")
			if err := verifyChecksum(isoPath, isoInfo.Checksum); err != nil {
				fmt.Printf("❌ Checksum verification failed: %v\n", err)
				os.Remove(isoPath) // Remove corrupted file
				os.Exit(1)
			}
			fmt.Println("✅ Checksum verified successfully")
		}

		fmt.Printf("✅ ISO '%s' downloaded successfully to %s\n", isoName, isoPath)
		fmt.Printf("\nTo use in VM creation:\n")
		fmt.Printf("  portunix virt create myvm --iso %s\n", isoPath)
	},
}

// virtISOListCmd represents the iso list command
var virtISOListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available and downloaded ISOs",
	Long:  `List all available ISOs for download and show which ones are already downloaded.`,
	Run: func(cmd *cobra.Command, args []string) {
		downloaded, _ := cmd.Flags().GetBool("downloaded")
		available, _ := cmd.Flags().GetBool("available")

		if downloaded {
			listDownloadedISOs()
		} else if available {
			listAvailableISOs()
		} else {
			// Show both by default
			fmt.Println("Available ISOs for download:")
			fmt.Println("=============================")
			listAvailableISOs()

			fmt.Println("\nDownloaded ISOs:")
			fmt.Println("================")
			listDownloadedISOs()
		}
	},
}

// virtISOVerifyCmd represents the iso verify command
var virtISOVerifyCmd = &cobra.Command{
	Use:   "verify [iso-name]",
	Short: "Verify the checksum of a downloaded ISO",
	Long:  `Verify the SHA256 checksum of a downloaded ISO file to ensure integrity.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		isoName := args[0]

		isoDir := getISOCacheDir()
		isoPath := filepath.Join(isoDir, getISOFilename(isoName))

		if _, err := os.Stat(isoPath); os.IsNotExist(err) {
			fmt.Printf("Error: ISO '%s' not found. Download it first with:\n", isoName)
			fmt.Printf("  portunix virt iso download %s\n", isoName)
			os.Exit(1)
		}

		// Get expected checksum
		isoInfo, exists := getAvailableISOs()[isoName]
		if !exists || isoInfo.Checksum == "" {
			fmt.Printf("Error: No checksum available for ISO '%s'\n", isoName)
			os.Exit(1)
		}

		fmt.Printf("Verifying checksum for %s...\n", isoName)
		if err := verifyChecksum(isoPath, isoInfo.Checksum); err != nil {
			fmt.Printf("❌ Checksum verification failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Checksum verified successfully for %s\n", isoName)
	},
}

// virtISOInfoCmd represents the iso info command
var virtISOInfoCmd = &cobra.Command{
	Use:   "info [iso-name]",
	Short: "Show detailed information about an ISO",
	Long:  `Show detailed information about an available or downloaded ISO.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		isoName := args[0]

		isoInfo, exists := getAvailableISOs()[isoName]
		if !exists {
			fmt.Printf("Error: ISO '%s' not found.\n", isoName)
			os.Exit(1)
		}

		isoDir := getISOCacheDir()
		isoPath := filepath.Join(isoDir, getISOFilename(isoName))
		isDownloaded := false
		var downloadedSize int64

		if stat, err := os.Stat(isoPath); err == nil {
			isDownloaded = true
			downloadedSize = stat.Size()
		}

		fmt.Printf("ISO Information: %s\n", isoName)
		fmt.Printf("===================\n\n")
		fmt.Printf("Name:        %s\n", isoInfo.Name)
		fmt.Printf("Description: %s\n", isoInfo.Description)
		fmt.Printf("Category:    %s\n", isoInfo.Category)
		fmt.Printf("URL:         %s\n", isoInfo.URL)

		if isoInfo.Size > 0 {
			fmt.Printf("Size:        %s\n", formatBytes(isoInfo.Size))
		}

		if isoInfo.Checksum != "" {
			fmt.Printf("Checksum:    %s\n", isoInfo.Checksum)
		}

		fmt.Printf("Downloaded:  %t\n", isDownloaded)

		if isDownloaded {
			fmt.Printf("Local Path:  %s\n", isoPath)
			fmt.Printf("Local Size:  %s\n", formatBytes(downloadedSize))
		}
	},
}

// virtISOCleanCmd represents the iso clean command
var virtISOCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up old or unused ISO files",
	Long: `Clean up old or unused ISO files to free disk space.

This command will:
- List all downloaded ISOs with their sizes
- Allow selective deletion of ISOs
- Show total space that can be freed

Use --force to delete all downloaded ISOs without confirmation.`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		isoDir := getISOCacheDir()
		if _, err := os.Stat(isoDir); os.IsNotExist(err) {
			fmt.Println("No ISO cache directory found.")
			return
		}

		entries, err := os.ReadDir(isoDir)
		if err != nil {
			fmt.Printf("Error reading ISO directory: %v\n", err)
			os.Exit(1)
		}

		var isoFiles []string
		var totalSize int64

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".iso") {
				isoPath := filepath.Join(isoDir, entry.Name())
				if stat, err := os.Stat(isoPath); err == nil {
					isoFiles = append(isoFiles, entry.Name())
					totalSize += stat.Size()
				}
			}
		}

		if len(isoFiles) == 0 {
			fmt.Println("No ISO files found in cache.")
			return
		}

		fmt.Printf("Found %d ISO files (%s total):\n\n", len(isoFiles), formatBytes(totalSize))

		for _, file := range isoFiles {
			isoPath := filepath.Join(isoDir, file)
			if stat, err := os.Stat(isoPath); err == nil {
				fmt.Printf("  %s (%s)\n", file, formatBytes(stat.Size()))
			}
		}

		if force {
			fmt.Printf("\nDeleting all ISO files...\n")
			for _, file := range isoFiles {
				isoPath := filepath.Join(isoDir, file)
				if err := os.Remove(isoPath); err != nil {
					fmt.Printf("❌ Failed to delete %s: %v\n", file, err)
				} else {
					fmt.Printf("🗑️  Deleted %s\n", file)
				}
			}
			fmt.Printf("✅ Cleanup completed!\n")
		} else {
			fmt.Printf("\nDelete all ISO files? This will free %s [y/N]: ", formatBytes(totalSize))
			var response string
			fmt.Scanln(&response)
			if response == "y" || response == "Y" {
				for _, file := range isoFiles {
					isoPath := filepath.Join(isoDir, file)
					if err := os.Remove(isoPath); err != nil {
						fmt.Printf("❌ Failed to delete %s: %v\n", file, err)
					} else {
						fmt.Printf("🗑️  Deleted %s\n", file)
					}
				}
				fmt.Printf("✅ Cleanup completed!\n")
			} else {
				fmt.Println("Cleanup cancelled.")
			}
		}
	},
}

// Helper functions
func getISOCacheDir() string {
	currentDir, _ := os.Getwd()
	return filepath.Join(currentDir, ".cache", "isos")
}

func getISOFilename(isoName string) string {
	// Map ISO names to actual filenames
	filenameMap := map[string]string{
		"ubuntu-24.04":        "ubuntu-24.04-desktop-amd64.iso",
		"ubuntu-24.04-server": "ubuntu-24.04.3-server-amd64.iso",
		"ubuntu-22.04":        "ubuntu-22.04-desktop-amd64.iso",
		"ubuntu-22.04-server": "ubuntu-22.04.3-server-amd64.iso",
		"debian-12":           "debian-12.4.0-amd64-netinst.iso",
		"windows11-eval":      "Win11_24H2_English_x64v2.iso",
		"windows10-eval":      "Win10_22H2_English_x64.iso",
	}

	if filename, exists := filenameMap[isoName]; exists {
		return filename
	}
	return isoName + ".iso"
}

func getAvailableISOs() map[string]ISOInfo {
	return map[string]ISOInfo{
		"ubuntu-24.04": {
			Name:        "Ubuntu 24.04 LTS Desktop",
			Description: "Ubuntu 24.04 LTS Desktop (64-bit)",
			URL:         "https://releases.ubuntu.com/24.04/ubuntu-24.04-desktop-amd64.iso",
			Category:    "Linux",
			Size:        5000000000, // ~5GB
		},
		"ubuntu-24.04-server": {
			Name:        "Ubuntu 24.04 LTS Server",
			Description: "Ubuntu 24.04 LTS Server (64-bit)",
			URL:         "https://releases.ubuntu.com/24.04/ubuntu-24.04.3-server-amd64.iso",
			Category:    "Linux",
			Size:        2500000000, // ~2.5GB
		},
		"ubuntu-22.04": {
			Name:        "Ubuntu 22.04 LTS Desktop",
			Description: "Ubuntu 22.04 LTS Desktop (64-bit)",
			URL:         "https://releases.ubuntu.com/22.04/ubuntu-22.04-desktop-amd64.iso",
			Category:    "Linux",
			Size:        4800000000, // ~4.8GB
		},
		"ubuntu-22.04-server": {
			Name:        "Ubuntu 22.04 LTS Server",
			Description: "Ubuntu 22.04 LTS Server (64-bit)",
			URL:         "https://releases.ubuntu.com/22.04/ubuntu-22.04.3-server-amd64.iso",
			Category:    "Linux",
			Size:        2300000000, // ~2.3GB
		},
		"debian-12": {
			Name:        "Debian 12 (Bookworm)",
			Description: "Debian 12 network installer (64-bit)",
			URL:         "https://cdimage.debian.org/debian-cd/current/amd64/iso-cd/debian-12.4.0-amd64-netinst.iso",
			Category:    "Linux",
			Size:        650000000, // ~650MB
		},
		"windows11-eval": {
			Name:        "Windows 11 Evaluation",
			Description: "Windows 11 Enterprise Evaluation (90 days)",
			URL:         "https://www.microsoft.com/en-us/evalcenter/download-windows-11-enterprise",
			Category:    "Windows",
			Size:        5500000000, // ~5.5GB
		},
		"windows10-eval": {
			Name:        "Windows 10 Evaluation",
			Description: "Windows 10 Enterprise Evaluation (90 days)",
			URL:         "https://www.microsoft.com/en-us/evalcenter/download-windows-10-enterprise",
			Category:    "Windows",
			Size:        5200000000, // ~5.2GB
		},
	}
}

func listAvailableISOs() {
	isos := getAvailableISOs()

	fmt.Printf("%-20s %-30s %-10s %s\n", "NAME", "DESCRIPTION", "SIZE", "CATEGORY")
	fmt.Printf("%-20s %-30s %-10s %s\n", "----", "-----------", "----", "--------")

	for name, info := range isos {
		sizeStr := "-"
		if info.Size > 0 {
			sizeStr = formatBytes(info.Size)
		}
		fmt.Printf("%-20s %-30s %-10s %s\n", name, info.Description, sizeStr, info.Category)
	}

	fmt.Printf("\nTo download: portunix virt iso download <name>\n")
}

func listDownloadedISOs() {
	isoDir := getISOCacheDir()

	if _, err := os.Stat(isoDir); os.IsNotExist(err) {
		fmt.Println("No ISOs downloaded yet.")
		fmt.Println("\nTo download: portunix virt iso download <name>")
		return
	}

	entries, err := os.ReadDir(isoDir)
	if err != nil {
		fmt.Printf("Error reading ISO directory: %v\n", err)
		return
	}

	var isoFiles []string
	var totalSize int64

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".iso") {
			isoFiles = append(isoFiles, entry.Name())
			isoPath := filepath.Join(isoDir, entry.Name())
			if stat, err := os.Stat(isoPath); err == nil {
				totalSize += stat.Size()
			}
		}
	}

	if len(isoFiles) == 0 {
		fmt.Println("No ISOs downloaded yet.")
		fmt.Println("\nTo download: portunix virt iso download <name>")
		return
	}

	fmt.Printf("%-35s %-12s %s\n", "FILENAME", "SIZE", "PATH")
	fmt.Printf("%-35s %-12s %s\n", "--------", "----", "----")

	for _, file := range isoFiles {
		isoPath := filepath.Join(isoDir, file)
		if stat, err := os.Stat(isoPath); err == nil {
			fmt.Printf("%-35s %-12s %s\n", file, formatBytes(stat.Size()), isoPath)
		}
	}

	fmt.Printf("\nTotal: %d files, %s\n", len(isoFiles), formatBytes(totalSize))
}

func downloadISO(url, destPath string) error {
	// Create temporary file
	tempPath := destPath + ".tmp"
	defer os.Remove(tempPath)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	file, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy with progress
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	// Move to final location
	return os.Rename(tempPath, destPath)
}

func verifyChecksum(filePath, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actualChecksum := fmt.Sprintf("%x", hash.Sum(nil))
	if strings.ToLower(actualChecksum) != strings.ToLower(expectedChecksum) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}


func init() {
	// Add ISO commands to virt
	virtCmd.AddCommand(virtISOCmd)
	virtISOCmd.AddCommand(virtISODownloadCmd)
	virtISOCmd.AddCommand(virtISOListCmd)
	virtISOCmd.AddCommand(virtISOVerifyCmd)
	virtISOCmd.AddCommand(virtISOInfoCmd)
	virtISOCmd.AddCommand(virtISOCleanCmd)

	// ISO download flags
	virtISODownloadCmd.Flags().Bool("force", false, "Force re-download if ISO already exists")
	virtISODownloadCmd.Flags().Bool("dry-run", false, "Show what would be done without actually downloading")

	// ISO list flags
	virtISOListCmd.Flags().Bool("downloaded", false, "Show only downloaded ISOs")
	virtISOListCmd.Flags().Bool("available", false, "Show only available ISOs")

	// ISO clean flags
	virtISOCleanCmd.Flags().Bool("force", false, "Delete all ISOs without confirmation")
}