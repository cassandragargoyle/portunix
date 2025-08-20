package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"portunix.cz/app/install/chocolatey"
)

var chocoCmd = &cobra.Command{
	Use:   "choco",
	Short: "Chocolatey package manager operations (Windows only)",
	Long: `Manage packages using Chocolatey package manager on Windows systems.
This command provides a convenient interface to Chocolatey operations including
installation, removal, search, and source management.

Examples:
  portunix choco install git nodejs
  portunix choco search python
  portunix choco uninstall vim
  portunix choco list-installed
  portunix choco upgrade all
  portunix choco --info
  portunix choco --info`,
	Run: func(cmd *cobra.Command, args []string) {
		showInfo, _ := cmd.Flags().GetBool("info")
		if showInfo {
			fmt.Println("🍫 Chocolatey Package Manager Information")
			fmt.Println("════════════════════════════════════════")
			fmt.Println()
			fmt.Println("Chocolatey is a console application written in C# that uses .NET Framework 4.x.")
			fmt.Println()
			fmt.Println("It is distributed as a NuGet package (chocolatey.nupkg). The initial installation")
			fmt.Println("is performed via PowerShell script that downloads and runs this package.")
			fmt.Println()
			fmt.Println("choco.exe is essentially a wrapper over NuGet that extends its capabilities")
			fmt.Println("(application installation, version management, scripting).")
			fmt.Println()
			fmt.Println("It can use PowerShell scripts to install applications in the background.")
			fmt.Println("Most packages consist of instructions on how to download MSI/EXE files")
			fmt.Println("and run the installer in silent mode.")
			fmt.Println()
			fmt.Println("Official website: https://community.chocolatey.org/")
			fmt.Println()
			return
		}
		cmd.Help()
	},
}

var chocoInstallSelfCmd = &cobra.Command{
	Use:   "install-chocolatey",
	Short: "Install Chocolatey package manager",
	Long: `Install Chocolatey package manager on Windows.
This will download and install Chocolatey using the official installation script.

Examples:
  portunix choco install-chocolatey`,
	Run: func(cmd *cobra.Command, args []string) {
		chocoMgr := chocolatey.NewChocolateyManager()

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		chocoMgr.DryRun = dryRun

		if err := chocoMgr.InstallChocolatey(); err != nil {
			fmt.Printf("Error: Failed to install Chocolatey: %v\n", err)
			return
		}

		fmt.Println("✓ Chocolatey installation completed successfully")
	},
}

var chocoInstallCmd = &cobra.Command{
	Use:   "install [packages...]",
	Short: "Install packages using Chocolatey",
	Long: `Install one or more packages using Chocolatey package manager.

Examples:
  portunix choco install git
  portunix choco install nodejs python`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		chocoMgr := chocolatey.NewChocolateyManager()

		if !chocoMgr.IsSupported() {
			fmt.Println("Error: Chocolatey is not installed on this system")
			fmt.Println("Run 'portunix choco install-chocolatey' to install it first")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		chocoMgr.DryRun = dryRun

		if err := chocoMgr.Install(args); err != nil {
			fmt.Printf("Error: Failed to install packages: %v\n", err)
			return
		}

		fmt.Println("✓ Package installation completed successfully")
	},
}

var chocoUninstallCmd = &cobra.Command{
	Use:   "uninstall [packages...]",
	Short: "Uninstall packages using Chocolatey",
	Long: `Uninstall one or more packages using Chocolatey package manager.

Examples:
  portunix choco uninstall git
  portunix choco uninstall nodejs python`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		chocoMgr := chocolatey.NewChocolateyManager()

		if !chocoMgr.IsSupported() {
			fmt.Println("Error: Chocolatey is not installed on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		chocoMgr.DryRun = dryRun

		if err := chocoMgr.Uninstall(args); err != nil {
			fmt.Printf("Error: Failed to uninstall packages: %v\n", err)
			return
		}

		fmt.Println("✓ Package uninstallation completed successfully")
	},
}

var chocoSearchCmd = &cobra.Command{
	Use:   "search [pattern]",
	Short: "Search for packages using Chocolatey",
	Long: `Search for packages matching the given pattern.
Shows package name, version, and installation status.

Examples:
  portunix choco search python
  portunix choco search git`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		chocoMgr := chocolatey.NewChocolateyManager()

		if !chocoMgr.IsSupported() {
			fmt.Println("Error: Chocolatey is not installed on this system")
			return
		}

		packages, err := chocoMgr.Search(args[0])
		if err != nil {
			fmt.Printf("Error: Search failed: %v\n", err)
			return
		}

		if len(packages) == 0 {
			fmt.Printf("No packages found matching '%s'\n", args[0])
			return
		}

		fmt.Printf("Found %d package(s) matching '%s':\n\n", len(packages), args[0])

		for _, pkg := range packages {
			status := "not installed"
			if pkg.Installed {
				status = "installed"
			}

			fmt.Printf("📦 %s (v%s) - %s\n", pkg.Name, pkg.Version, status)
		}
	},
}

var chocoListCmd = &cobra.Command{
	Use:     "list-installed",
	Aliases: []string{"list"},
	Short:   "List all installed packages",
	Long: `List all packages installed via Chocolatey.
Shows package name and version.

Examples:
  portunix choco list-installed
  portunix choco list`,
	Run: func(cmd *cobra.Command, args []string) {
		chocoMgr := chocolatey.NewChocolateyManager()

		if !chocoMgr.IsSupported() {
			fmt.Println("Error: Chocolatey is not installed on this system")
			return
		}

		packages, err := chocoMgr.ListInstalled()
		if err != nil {
			fmt.Printf("Error: Failed to list packages: %v\n", err)
			return
		}

		fmt.Printf("Installed packages (%d total):\n\n", len(packages))

		for _, pkg := range packages {
			fmt.Printf("📦 %s (v%s)\n", pkg.Name, pkg.Version)
		}
	},
}

var chocoUpgradeCmd = &cobra.Command{
	Use:   "upgrade [packages...]",
	Short: "Upgrade packages using Chocolatey",
	Long: `Upgrade all packages or specific packages to their latest versions.
If no packages are specified, all upgradable packages will be upgraded.

Examples:
  portunix choco upgrade
  portunix choco upgrade git nodejs`,
	Run: func(cmd *cobra.Command, args []string) {
		chocoMgr := chocolatey.NewChocolateyManager()

		if !chocoMgr.IsSupported() {
			fmt.Println("Error: Chocolatey is not installed on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		chocoMgr.DryRun = dryRun

		if err := chocoMgr.Upgrade(args); err != nil {
			fmt.Printf("Error: Failed to upgrade packages: %v\n", err)
			return
		}

		fmt.Println("✓ Package upgrade completed successfully")
	},
}

var chocoPinCmd = &cobra.Command{
	Use:   "pin [package]",
	Short: "Pin a package to prevent upgrades",
	Long: `Pin a package to prevent it from being upgraded.

Examples:
  portunix choco pin nodejs`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		chocoMgr := chocolatey.NewChocolateyManager()

		if !chocoMgr.IsSupported() {
			fmt.Println("Error: Chocolatey is not installed on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		chocoMgr.DryRun = dryRun

		if err := chocoMgr.Pin(args[0]); err != nil {
			fmt.Printf("Error: Failed to pin package: %v\n", err)
			return
		}

		fmt.Println("✓ Package pinned successfully")
	},
}

var chocoUnpinCmd = &cobra.Command{
	Use:   "unpin [package]",
	Short: "Unpin a package to allow upgrades",
	Long: `Unpin a package to allow it to be upgraded.

Examples:
  portunix choco unpin nodejs`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		chocoMgr := chocolatey.NewChocolateyManager()

		if !chocoMgr.IsSupported() {
			fmt.Println("Error: Chocolatey is not installed on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		chocoMgr.DryRun = dryRun

		if err := chocoMgr.Unpin(args[0]); err != nil {
			fmt.Printf("Error: Failed to unpin package: %v\n", err)
			return
		}

		fmt.Println("✓ Package unpinned successfully")
	},
}

var chocoSourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Manage Chocolatey sources",
	Long: `Manage Chocolatey package sources.

Examples:
  portunix choco source list
  portunix choco source add mysource https://myrepo.com/api/v2`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var chocoSourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured sources",
	Long: `List all configured Chocolatey sources.

Examples:
  portunix choco source list`,
	Run: func(cmd *cobra.Command, args []string) {
		chocoMgr := chocolatey.NewChocolateyManager()

		if !chocoMgr.IsSupported() {
			fmt.Println("Error: Chocolatey is not installed on this system")
			return
		}

		if err := chocoMgr.ListSources(); err != nil {
			fmt.Printf("Error: Failed to list sources: %v\n", err)
			return
		}
	},
}

var chocoSourceAddCmd = &cobra.Command{
	Use:   "add [name] [url]",
	Short: "Add a new source",
	Long: `Add a new Chocolatey package source.

Examples:
  portunix choco source add mysource https://myrepo.com/api/v2
  portunix choco source add private https://private.repo.com --username user --password pass`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		chocoMgr := chocolatey.NewChocolateyManager()

		if !chocoMgr.IsSupported() {
			fmt.Println("Error: Chocolatey is not installed on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		chocoMgr.DryRun = dryRun

		if err := chocoMgr.AddSource(args[0], args[1], username, password); err != nil {
			fmt.Printf("Error: Failed to add source: %v\n", err)
			return
		}

		fmt.Println("✓ Source added successfully")
	},
}

var chocoSourceRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a source",
	Long: `Remove a Chocolatey package source.

Examples:
  portunix choco source remove mysource`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		chocoMgr := chocolatey.NewChocolateyManager()

		if !chocoMgr.IsSupported() {
			fmt.Println("Error: Chocolatey is not installed on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		chocoMgr.DryRun = dryRun

		if err := chocoMgr.RemoveSource(args[0]); err != nil {
			fmt.Printf("Error: Failed to remove source: %v\n", err)
			return
		}

		fmt.Println("✓ Source removed successfully")
	},
}

var chocoCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean Chocolatey cache",
	Long: `Clean the Chocolatey cache to free up disk space.

Examples:
  portunix choco clean`,
	Run: func(cmd *cobra.Command, args []string) {
		chocoMgr := chocolatey.NewChocolateyManager()

		if !chocoMgr.IsSupported() {
			fmt.Println("Error: Chocolatey is not installed on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		chocoMgr.DryRun = dryRun

		if err := chocoMgr.Clean(); err != nil {
			fmt.Printf("Error: Failed to clean cache: %v\n", err)
			return
		}

		fmt.Println("✓ Cache cleaned successfully")
	},
}

func init() {
	rootCmd.AddCommand(chocoCmd)

	// Add flags
	chocoCmd.Flags().Bool("info", false, "Show information about Chocolatey")

	// Add subcommands
	chocoCmd.AddCommand(chocoInstallSelfCmd)
	chocoCmd.AddCommand(chocoInstallCmd)
	chocoCmd.AddCommand(chocoUninstallCmd)
	chocoCmd.AddCommand(chocoSearchCmd)
	chocoCmd.AddCommand(chocoListCmd)
	chocoCmd.AddCommand(chocoUpgradeCmd)
	chocoCmd.AddCommand(chocoPinCmd)
	chocoCmd.AddCommand(chocoUnpinCmd)
	chocoCmd.AddCommand(chocoSourceCmd)
	chocoCmd.AddCommand(chocoCleanCmd)

	// Source commands
	chocoSourceCmd.AddCommand(chocoSourceListCmd)
	chocoSourceCmd.AddCommand(chocoSourceAddCmd)
	chocoSourceCmd.AddCommand(chocoSourceRemoveCmd)

	// Add flags
	chocoInstallSelfCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	chocoInstallCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	chocoUninstallCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	chocoUpgradeCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	chocoPinCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	chocoUnpinCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	chocoSourceAddCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	chocoSourceAddCmd.Flags().String("username", "", "Username for authenticated source")
	chocoSourceAddCmd.Flags().String("password", "", "Password for authenticated source")
	chocoSourceRemoveCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	chocoCleanCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
}
