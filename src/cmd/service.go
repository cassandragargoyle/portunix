/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"portunix.ai/app/plugins/manager"
	"portunix.ai/app/service"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service orchestration commands",
	Long: `Manage gRPC plugin service instances.

The service command group provides lifecycle management for gRPC plugin
processes. Clients (VSCode extensions, CLI scripts) can start, discover,
and stop plugin services without knowing internal details.

Examples:
  portunix service start reco                          # Start shared instance
  portunix service start reco --mode exclusive         # Start dedicated instance
  portunix service list                                # Show running instances
  portunix service info reco --port 50101              # Show gRPC services
  portunix service release --session abc123            # Release a session
  portunix service stop reco --force                   # Force stop all instances
  portunix service stop --all --force                  # Stop everything`,
}

var serviceStartCmd = &cobra.Command{
	Use:   "start <plugin>",
	Short: "Start a plugin service instance",
	Long: `Start a gRPC plugin service instance or join an existing one.

Allocation modes:
  shared (default)    Join existing shared instance or start new one
  exclusive           Always start a dedicated instance
  prefer-exclusive    Portunix decides based on resource availability`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		modeStr, _ := cmd.Flags().GetString("mode")
		outputFormat, _ := cmd.Flags().GetString("output")

		mode := service.AllocationMode(modeStr)
		return serviceStart(pluginName, mode, outputFormat)
	},
}

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List running service instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")
		return serviceList(outputFormat)
	},
}

var serviceInfoCmd = &cobra.Command{
	Use:   "info <plugin>",
	Short: "Show gRPC service details for a running instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		port, _ := cmd.Flags().GetInt("port")
		outputFormat, _ := cmd.Flags().GetString("output")
		return serviceInfo(pluginName, port, outputFormat)
	},
}

var serviceReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release a client session",
	Long:  `Remove a session from a service instance. If it's the last session, the instance is stopped.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID, _ := cmd.Flags().GetString("session")
		if sessionID == "" {
			return fmt.Errorf("--session flag is required")
		}
		return serviceRelease(sessionID)
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop [plugin]",
	Short: "Force stop service instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		all, _ := cmd.Flags().GetBool("all")

		if !force {
			return fmt.Errorf("--force flag is required for stop command")
		}

		if all {
			return serviceStopAll()
		}

		if len(args) < 1 {
			return fmt.Errorf("plugin name required, or use --all to stop everything")
		}

		return serviceStopPlugin(args[0])
	},
}

func init() {
	// Register service orchestrator factory for plugin lifecycle integration
	manager.RegisterServiceOrchestrator(func() (manager.ServiceOrchestrator, error) {
		orch, err := service.NewOrchestrator()
		if err != nil {
			return nil, err
		}
		return orch, nil
	})

	rootCmd.AddCommand(serviceCmd)

	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceListCmd)
	serviceCmd.AddCommand(serviceInfoCmd)
	serviceCmd.AddCommand(serviceReleaseCmd)
	serviceCmd.AddCommand(serviceStopCmd)

	// Flags
	serviceStartCmd.Flags().StringP("mode", "m", "shared", "Allocation mode: shared, exclusive, prefer-exclusive")
	serviceStartCmd.Flags().StringP("output", "o", "text", "Output format: text, json")

	serviceListCmd.Flags().StringP("output", "o", "text", "Output format: text, json")

	serviceInfoCmd.Flags().IntP("port", "p", 0, "Port of the service instance")
	serviceInfoCmd.MarkFlagRequired("port")
	serviceInfoCmd.Flags().StringP("output", "o", "text", "Output format: text, json")

	serviceReleaseCmd.Flags().StringP("session", "s", "", "Session ID to release")

	serviceStopCmd.Flags().BoolP("force", "f", false, "Force stop (required)")
	serviceStopCmd.Flags().BoolP("all", "a", false, "Stop all instances")
}

func serviceStart(pluginName string, mode service.AllocationMode, outputFormat string) error {
	// Validate mode
	switch mode {
	case service.ModeShared, service.ModeExclusive, service.ModePreferExclusive:
	default:
		return fmt.Errorf("invalid mode: %s (expected shared, exclusive, or prefer-exclusive)", mode)
	}

	// Resolve plugin binary path from registry
	binaryPath, runtime, jvmArgs, err := resolvePluginBinary(pluginName)
	if err != nil {
		return err
	}

	orch, err := service.NewOrchestrator()
	if err != nil {
		return err
	}

	result, err := orch.StartService(pluginName, binaryPath, runtime, jvmArgs, mode)
	if err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	if outputFormat == "json" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	action := "Started new"
	if !result.NewProcess {
		action = "Joined existing"
	}
	fmt.Printf("✅ %s %s service instance\n", action, pluginName)
	fmt.Printf("   Session:  %s\n", result.SessionID)
	fmt.Printf("   Endpoint: %s\n", result.Endpoint)
	fmt.Printf("   PID:      %d\n", result.PID)
	fmt.Printf("   Mode:     %s\n", result.Mode)
	if len(result.Services) > 0 {
		fmt.Printf("   Services: %s\n", strings.Join(result.Services, ", "))
	}

	return nil
}

func serviceList(outputFormat string) error {
	orch, err := service.NewOrchestrator()
	if err != nil {
		return err
	}

	instances, err := orch.ListInstances()
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		fmt.Println("No running service instances.")
		return nil
	}

	if outputFormat == "json" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(instances)
	}

	fmt.Printf("%-20s %-8s %-10s %-12s %-10s %s\n",
		"PLUGIN", "PORT", "PID", "MODE", "SESSIONS", "UPTIME")
	fmt.Printf("%-20s %-8s %-10s %-12s %-10s %s\n",
		strings.Repeat("-", 20), strings.Repeat("-", 8), strings.Repeat("-", 10),
		strings.Repeat("-", 12), strings.Repeat("-", 10), strings.Repeat("-", 10))

	for _, inst := range instances {
		uptime := time.Since(inst.StartedAt).Truncate(time.Second)
		fmt.Printf("%-20s %-8d %-10d %-12s %-10d %s\n",
			inst.PluginName,
			inst.Port,
			inst.PID,
			inst.Mode,
			len(inst.Sessions),
			uptime.String(),
		)
	}

	return nil
}

func serviceInfo(pluginName string, port int, outputFormat string) error {
	orch, err := service.NewOrchestrator()
	if err != nil {
		return err
	}

	inst, methods, err := orch.GetInstanceInfo(pluginName, port)
	if err != nil {
		return err
	}

	if outputFormat == "json" {
		data := map[string]interface{}{
			"instance": inst,
			"methods":  methods,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	}

	fmt.Printf("Service Instance: %s (port %d)\n", pluginName, port)
	fmt.Printf("  PID:      %d\n", inst.PID)
	fmt.Printf("  Mode:     %s\n", inst.Mode)
	fmt.Printf("  Sessions: %d\n", len(inst.Sessions))
	fmt.Printf("  Started:  %s\n", inst.StartedAt.Format(time.RFC3339))

	if len(inst.Services) > 0 {
		fmt.Printf("\n  Registered services:\n")
		for _, svc := range inst.Services {
			fmt.Printf("    - %s\n", svc)
		}
	}

	if methods != nil {
		fmt.Printf("\n  gRPC methods:\n")
		for _, m := range methods {
			fmt.Printf("    %s/%s\n", m.Service, m.Method)
		}
	}

	if len(inst.Sessions) > 0 {
		fmt.Printf("\n  Active sessions:\n")
		for _, sess := range inst.Sessions {
			fmt.Printf("    %s (created %s)\n", sess.ID, sess.CreatedAt.Format(time.RFC3339))
		}
	}

	return nil
}

func serviceRelease(sessionID string) error {
	orch, err := service.NewOrchestrator()
	if err != nil {
		return err
	}

	stopped, err := orch.ReleaseSession(sessionID)
	if err != nil {
		return err
	}

	if stopped {
		fmt.Printf("✅ Session %s released, instance stopped (was last session)\n", sessionID)
	} else {
		fmt.Printf("✅ Session %s released\n", sessionID)
	}

	return nil
}

func serviceStopPlugin(pluginName string) error {
	orch, err := service.NewOrchestrator()
	if err != nil {
		return err
	}

	count, err := orch.StopPlugin(pluginName)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Stopped %d instance(s) of %s\n", count, pluginName)
	return nil
}

func serviceStopAll() error {
	orch, err := service.NewOrchestrator()
	if err != nil {
		return err
	}

	count, err := orch.StopAll()
	if err != nil {
		return err
	}

	if count == 0 {
		fmt.Println("No running instances to stop.")
	} else {
		fmt.Printf("✅ Stopped %d instance(s)\n", count)
	}

	return nil
}

// resolvePluginBinary resolves the binary path for a plugin from the registry
func resolvePluginBinary(pluginName string) (binaryPath, runtime string, jvmArgs []string, err error) {
	if err := initializePluginManager(); err != nil {
		return "", "", nil, fmt.Errorf("failed to initialize plugin manager: %w", err)
	}
	defer func() {
		if pluginManager != nil {
			pluginManager.Shutdown()
		}
	}()

	registryData, regErr := pluginManager.GetPluginRegistryData(pluginName)
	if regErr != nil {
		return "", "", nil, fmt.Errorf("plugin %s not found: %w", pluginName, regErr)
	}

	// Only gRPC/service plugins can be started as services
	if registryData.Mode == "helper" {
		hint := ""
		if registryData.Runtime == "python" && registryData.Wheel != "" {
			hint = fmt.Sprintf(
				"\nHint: Python plugins can run as gRPC services. Set \"type\": \"grpc\" in plugin.json,"+
					" then reinstall:\n  portunix plugin uninstall %s && portunix plugin install <path>",
				pluginName)
		}
		return "", "", nil, fmt.Errorf("plugin %s is a helper plugin, not a gRPC service%s", pluginName, hint)
	}

	if !registryData.Enabled {
		return "", "", nil, fmt.Errorf("plugin %s is not enabled. Enable it with: portunix plugin enable %s", pluginName, pluginName)
	}

	binaryPath = registryData.BinaryPath()
	runtime = registryData.Runtime
	jvmArgs = registryData.JVMArgs

	return binaryPath, runtime, jvmArgs, nil
}
