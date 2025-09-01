package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"portunix.cz/app/selfinstall"
)

var (
	selfInstallSilent       bool
	selfInstallPath         string
	selfInstallCreateConfig bool
	selfInstallAddToPath    bool
)

var installSelfCmd = &cobra.Command{
	Use:   "install-self",
	Short: "Install Portunix binary to the system",
	Long: `Install Portunix binary to the system with interactive or silent mode.

The installation process will:
- Copy the binary to the installation directory
- Optionally add to system PATH
- Create configuration files if requested
- Verify the installation

Examples:
  # Interactive installation
  portunix install-self
  
  # Silent installation with defaults
  portunix install-self --silent
  
  # Install to specific location
  portunix install-self --path /usr/local/bin
  
  # Full silent installation with all options
  portunix install-self --silent --path /usr/local/bin --add-to-path --create-config`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the path of the current executable
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}

		// Check if already installed at target location
		if selfInstallPath != "" && execPath == selfInstallPath {
			fmt.Println("✓ Portunix is already installed at this location")
			return nil
		}

		if selfInstallSilent {
			// Silent installation
			options := selfinstall.Options{
				SourcePath:   execPath,
				TargetPath:   selfInstallPath,
				CreateConfig: selfInstallCreateConfig,
				AddToPath:    selfInstallAddToPath,
				Silent:       true,
			}

			if selfInstallPath == "" {
				options.TargetPath = selfinstall.GetDefaultInstallPath()
			}

			return selfinstall.InstallSilent(options)
		} else {
			// Interactive installation
			return selfinstall.InstallInteractive(execPath)
		}
	},
}

func init() {
	installSelfCmd.Flags().BoolVar(&selfInstallSilent, "silent", false, "Silent installation with defaults")
	installSelfCmd.Flags().StringVar(&selfInstallPath, "path", "", "Installation path")
	installSelfCmd.Flags().BoolVar(&selfInstallCreateConfig, "create-config", false, "Create configuration files")
	installSelfCmd.Flags().BoolVar(&selfInstallAddToPath, "add-to-path", false, "Add to system PATH")

	rootCmd.AddCommand(installSelfCmd)
}
