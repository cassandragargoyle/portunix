package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"portunix.cz/app/edge"
)

var edgeCmd = &cobra.Command{
	Use:   "edge",
	Short: "Manage VPS edge/bastion infrastructure",
	Long: `Edge infrastructure management for deploying and managing VPS edge/bastion hosts.
	
This command allows you to:
- Deploy edge infrastructure with reverse proxy and VPN tunneling
- Configure domains and TLS certificates  
- Manage WireGuard VPN connections
- Set up security hardening and monitoring`,
}

var edgeInitCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize edge infrastructure configuration",
	Long:  `Initialize a new edge infrastructure configuration with templates and default settings.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "default-edge"
		if len(args) > 0 {
			name = args[0]
		}

		configDir := filepath.Join(".", "edge-config", name)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		manager := edge.NewManager()
		return manager.InitializeConfiguration(name, configDir)
	},
}

var edgeDeployCmd = &cobra.Command{
	Use:   "deploy [config-path]",
	Short: "Deploy edge infrastructure",
	Long:  `Deploy edge infrastructure based on configuration file.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := "./edge-config"
		if len(args) > 0 {
			configPath = args[0]
		}

		manager := edge.NewManager()
		return manager.Deploy(configPath)
	},
}

var edgeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show edge infrastructure status",
	Long:  `Display status of edge infrastructure components including services, VPN, and security.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := edge.NewManager()
		return manager.ShowStatus()
	},
}

var edgeStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop edge infrastructure services",
	Long:  `Stop all edge infrastructure services including reverse proxy, VPN, and security services.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := edge.NewManager()
		return manager.Stop()
	},
}

var edgeStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start edge infrastructure services",
	Long:  `Start all edge infrastructure services including reverse proxy, VPN, and security services.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := edge.NewManager()
		return manager.Start()
	},
}

var edgeLogsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "Show logs for edge services",
	Long:  `Display logs for edge infrastructure services. Specify service name or leave empty for all services.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := ""
		if len(args) > 0 {
			service = args[0]
		}

		follow, _ := cmd.Flags().GetBool("follow")
		tail, _ := cmd.Flags().GetInt("tail")

		manager := edge.NewManager()
		return manager.ShowLogs(service, follow, tail)
	},
}

var edgeConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage edge configuration",
	Long:  `Manage edge infrastructure configuration including domains, VPN clients, and security settings.`,
}

var edgeConfigAddDomainCmd = &cobra.Command{
	Use:   "add-domain <domain> <upstream-host> <upstream-port>",
	Short: "Add domain to edge configuration",
	Long:  `Add a new domain to the edge infrastructure configuration.`,
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		upstreamHost := args[1]
		upstreamPort := args[2]

		manager := edge.NewManager()
		return manager.AddDomain(domain, upstreamHost, upstreamPort)
	},
}

var edgeConfigAddClientCmd = &cobra.Command{
	Use:   "add-client <name> <public-key>",
	Short: "Add VPN client to edge configuration",
	Long:  `Add a new VPN client to the WireGuard configuration.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		clientName := args[0]
		publicKey := args[1]

		manager := edge.NewManager()
		return manager.AddVPNClient(clientName, publicKey)
	},
}

var edgeInstallCmd = &cobra.Command{
	Use:   "install [preset]",
	Short: "Install edge infrastructure packages",
	Long: `Install packages required for edge infrastructure.
	
Available presets:
- minimal: Caddy and WireGuard only
- secure: Full security stack with fail2ban and firewall`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		preset := "edge-minimal"
		if len(args) > 0 {
			switch args[0] {
			case "minimal":
				preset = "edge-minimal"
			case "secure":
				preset = "edge-secure"
			default:
				return fmt.Errorf("unknown preset: %s. Available: minimal, secure", args[0])
			}
		}

		fmt.Printf("Installing edge infrastructure packages (%s preset)...\n", preset)

		// Use existing install command
		installArgs := []string{"install", preset}
		return installCmd.RunE(cmd, installArgs)
	},
}

func init() {
	// Add edge command to root
	rootCmd.AddCommand(edgeCmd)

	// Add subcommands
	edgeCmd.AddCommand(edgeInitCmd)
	edgeCmd.AddCommand(edgeDeployCmd)
	edgeCmd.AddCommand(edgeStatusCmd)
	edgeCmd.AddCommand(edgeStartCmd)
	edgeCmd.AddCommand(edgeStopCmd)
	edgeCmd.AddCommand(edgeLogsCmd)
	edgeCmd.AddCommand(edgeConfigCmd)
	edgeCmd.AddCommand(edgeInstallCmd)

	// Add config subcommands
	edgeConfigCmd.AddCommand(edgeConfigAddDomainCmd)
	edgeConfigCmd.AddCommand(edgeConfigAddClientCmd)

	// Add flags
	edgeLogsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	edgeLogsCmd.Flags().IntP("tail", "t", 100, "Number of lines to show from end of logs")
}
