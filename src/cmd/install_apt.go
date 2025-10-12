package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"portunix.ai/app/install/apt"
)

var aptCmd = &cobra.Command{
	Use:   "apt",
	Short: "APT package manager operations (Linux only)",
	Long: `Manage packages using APT package manager on Linux systems.
This command provides a convenient interface to APT operations including
installation, removal, search, and repository management.

Examples:
  portunix install apt install python3 git
  portunix install apt search nodejs
  portunix install apt remove vim
  portunix install apt list-installed
  portunix install apt clean`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var aptInstallCmd = &cobra.Command{
	Use:   "install [packages...]",
	Short: "Install packages using APT",
	Long: `Install one or more packages using APT package manager.
This will automatically update the package list before installation.

Examples:
  portunix install apt install python3
  portunix install apt install git vim curl`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		aptMgr := apt.NewAptManager()

		if !aptMgr.IsSupported() {
			fmt.Println("Error: APT is not supported on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		aptMgr.DryRun = dryRun

		// Update package list first
		if err := aptMgr.Update(); err != nil {
			fmt.Printf("Warning: Failed to update package list: %v\n", err)
		}

		// Install packages
		if err := aptMgr.Install(args); err != nil {
			fmt.Printf("Error: Failed to install packages: %v\n", err)
			return
		}

		fmt.Println("✓ Package installation completed successfully")
	},
}

var aptRemoveCmd = &cobra.Command{
	Use:   "remove [packages...]",
	Short: "Remove packages using APT",
	Long: `Remove one or more packages using APT package manager.
Use --purge flag to also remove configuration files.

Examples:
  portunix install apt remove vim
  portunix install apt remove --purge apache2`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		aptMgr := apt.NewAptManager()

		if !aptMgr.IsSupported() {
			fmt.Println("Error: APT is not supported on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		purge, _ := cmd.Flags().GetBool("purge")
		aptMgr.DryRun = dryRun

		var err error
		if purge {
			err = aptMgr.Purge(args)
		} else {
			err = aptMgr.Remove(args)
		}

		if err != nil {
			fmt.Printf("Error: Failed to remove packages: %v\n", err)
			return
		}

		fmt.Println("✓ Package removal completed successfully")
	},
}

var aptSearchCmd = &cobra.Command{
	Use:   "search [pattern]",
	Short: "Search for packages using APT",
	Long: `Search for packages matching the given pattern.
Shows package name, installation status, and description.

Examples:
  portunix install apt search python
  portunix install apt search web-server`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		aptMgr := apt.NewAptManager()

		if !aptMgr.IsSupported() {
			fmt.Println("Error: APT is not supported on this system")
			return
		}

		packages, err := aptMgr.Search(args[0])
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

			fmt.Printf("📦 %s (%s)\n", pkg.Name, status)
			if pkg.Description != "" {
				fmt.Printf("   %s\n", pkg.Description)
			}
			fmt.Println()
		}
	},
}

var aptListCmd = &cobra.Command{
	Use:   "list-installed",
	Short: "List all installed packages",
	Long: `List all packages installed on the system using APT.
Shows package name, version, and description.

Examples:
  portunix install apt list-installed
  portunix install apt list-installed | grep python`,
	Run: func(cmd *cobra.Command, args []string) {
		aptMgr := apt.NewAptManager()

		if !aptMgr.IsSupported() {
			fmt.Println("Error: APT is not supported on this system")
			return
		}

		packages, err := aptMgr.ListInstalled()
		if err != nil {
			fmt.Printf("Error: Failed to list packages: %v\n", err)
			return
		}

		fmt.Printf("Installed packages (%d total):\n\n", len(packages))

		for _, pkg := range packages {
			fmt.Printf("📦 %s", pkg.Name)
			if pkg.Version != "" {
				fmt.Printf(" (v%s)", pkg.Version)
			}
			fmt.Println()

			if pkg.Description != "" {
				fmt.Printf("   %s\n", pkg.Description)
			}
		}
	},
}

var aptUpgradeCmd = &cobra.Command{
	Use:   "upgrade [packages...]",
	Short: "Upgrade packages using APT",
	Long: `Upgrade all packages or specific packages to their latest versions.
If no packages are specified, all upgradable packages will be upgraded.

Examples:
  portunix install apt upgrade
  portunix install apt upgrade git vim`,
	Run: func(cmd *cobra.Command, args []string) {
		aptMgr := apt.NewAptManager()

		if !aptMgr.IsSupported() {
			fmt.Println("Error: APT is not supported on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		aptMgr.DryRun = dryRun

		// Update package list first
		if err := aptMgr.Update(); err != nil {
			fmt.Printf("Warning: Failed to update package list: %v\n", err)
		}

		// Upgrade packages
		if err := aptMgr.Upgrade(args); err != nil {
			fmt.Printf("Error: Failed to upgrade packages: %v\n", err)
			return
		}

		fmt.Println("✓ Package upgrade completed successfully")
	},
}

var aptCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean APT cache and remove unnecessary packages",
	Long: `Clean the APT cache and remove packages that are no longer needed.
This helps free up disk space by removing cached package files and
automatically removing packages that were installed as dependencies
but are no longer required.

Examples:
  portunix install apt clean`,
	Run: func(cmd *cobra.Command, args []string) {
		aptMgr := apt.NewAptManager()

		if !aptMgr.IsSupported() {
			fmt.Println("Error: APT is not supported on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		aptMgr.DryRun = dryRun

		if err := aptMgr.Clean(); err != nil {
			fmt.Printf("Error: Failed to clean: %v\n", err)
			return
		}

		fmt.Println("✓ APT cache cleaned successfully")
	},
}

var aptRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage APT repositories",
	Long: `Manage APT repositories including adding and removing custom repositories.

Examples:
  portunix install apt repo add
  portunix install apt repo remove`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var aptRepoAddCmd = &cobra.Command{
	Use:   "add [uri] [distribution] [components...]",
	Short: "Add APT repository",
	Long: `Add a new APT repository to the system.
Optionally specify GPG key URL for repository verification.

Examples:
  portunix install apt repo add https://deb.nodesource.com/node_18.x focal main --gpg-url https://deb.nodesource.com/gpgkey/nodesource.gpg.key
  portunix install apt repo add ppa:deadsnakes/ppa`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		aptMgr := apt.NewAptManager()

		if !aptMgr.IsSupported() {
			fmt.Println("Error: APT is not supported on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		gpgURL, _ := cmd.Flags().GetString("gpg-url")
		gpgKey, _ := cmd.Flags().GetString("gpg-key")

		aptMgr.DryRun = dryRun

		repo := apt.Repository{
			URI:          args[0],
			Distribution: args[1],
			Components:   args[2:],
			GPGKeyURL:    gpgURL,
			GPGKey:       gpgKey,
		}

		if err := aptMgr.AddRepository(repo); err != nil {
			fmt.Printf("Error: Failed to add repository: %v\n", err)
			return
		}

		fmt.Println("✓ Repository added successfully")
	},
}

var aptRepoRemoveCmd = &cobra.Command{
	Use:   "remove [distribution]",
	Short: "Remove APT repository",
	Long: `Remove an APT repository from the system.

Examples:
  portunix install apt repo remove focal`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		aptMgr := apt.NewAptManager()

		if !aptMgr.IsSupported() {
			fmt.Println("Error: APT is not supported on this system")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		aptMgr.DryRun = dryRun

		if err := aptMgr.RemoveRepository(args[0]); err != nil {
			fmt.Printf("Error: Failed to remove repository: %v\n", err)
			return
		}

		fmt.Println("✓ Repository removed successfully")
	},
}

func init() {
	installCmd.AddCommand(aptCmd)

	// Add subcommands
	aptCmd.AddCommand(aptInstallCmd)
	aptCmd.AddCommand(aptRemoveCmd)
	aptCmd.AddCommand(aptSearchCmd)
	aptCmd.AddCommand(aptListCmd)
	aptCmd.AddCommand(aptUpgradeCmd)
	aptCmd.AddCommand(aptCleanCmd)
	aptCmd.AddCommand(aptRepoCmd)

	// Repository commands
	aptRepoCmd.AddCommand(aptRepoAddCmd)
	aptRepoCmd.AddCommand(aptRepoRemoveCmd)

	// Add flags
	aptInstallCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	aptRemoveCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	aptRemoveCmd.Flags().Bool("purge", false, "Remove configuration files as well")
	aptUpgradeCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	aptCleanCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	aptRepoAddCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	aptRepoAddCmd.Flags().String("gpg-url", "", "URL to download GPG key from")
	aptRepoAddCmd.Flags().String("gpg-key", "", "GPG key ID to add from keyserver")
	aptRepoRemoveCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
}
