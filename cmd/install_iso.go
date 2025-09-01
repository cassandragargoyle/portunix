package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"portunix.cz/app/install"
)

// installISOCmd represents the install iso command
var installISOCmd = &cobra.Command{
	Use:   "iso [os-type]",
	Short: "Download OS installation ISO",
	Long: `Download official OS installation ISO files using the configured sources.
	
Available OS types:
- windows11 : Windows 11 latest version
- windows10 : Windows 10 latest version  
- ubuntu    : Ubuntu LTS versions
- debian    : Debian stable
- fedora    : Fedora Workstation`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		osType := "windows11"
		if len(args) > 0 {
			osType = args[0]
		}

		variant, _ := cmd.Flags().GetString("variant")
		if variant == "" {
			variant = "latest"
		}

		outputDir, _ := cmd.Flags().GetString("output")
		// If no output specified, use cache directory

		fmt.Printf("Downloading %s ISO (variant: %s)...\n", osType, variant)

		// Use the install system to download ISO
		installer := &install.ISOInstaller{
			OSType:    osType,
			Variant:   variant,
			OutputDir: outputDir,
		}

		isoPath, err := installer.Download()
		if err != nil {
			fmt.Printf("Error downloading ISO: %v\n", err)

			// Special handling for Windows ISOs
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
	},
}

func init() {
	installCmd.AddCommand(installISOCmd)

	installISOCmd.Flags().String("variant", "latest", "ISO variant (e.g., latest, 22.04, media-tool)")
	installISOCmd.Flags().StringP("output", "o", "", "Output directory for ISO file")
}
