/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// ptx-plugin-registry serves the PluginRegistryService gRPC API over a local
// transport (unix domain socket on Linux/macOS, TCP loopback on Windows).
// Hosting platforms (Synapse, Pack, Agent, ...) use it to discover which
// installed Portunix plugins target them and retrieve each plugin's
// platform-specific registration payload.
//
// Issue #175, DEC-2 / DEC-3 / DEC-4:
//   - Helper binary (not a top-level command) following the Helper Binary
//     Development checklist.
//   - Unix socket on Linux/macOS (default), TCP loopback on Windows.
//   - On-demand lifecycle — caller starts the process for the duration of the
//     query and stops it. No systemd integration in v1.
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"portunix.ai/app/plugins/manager"
	pb "portunix.ai/app/plugins/proto/pluginregistry"
	"portunix.ai/app/plugins/server"
)

var version = "dev"

// pluginRegistryCmd is the top-level command group for this helper.
var pluginRegistryCmd = &cobra.Command{
	Use:   "plugin-registry",
	Short: "Local gRPC registry for plugin platform-capability queries (issue #175)",
	Long: `The plugin-registry helper exposes the installed Portunix plugin registry
to hosting platforms (Synapse, Pack, Agent) via gRPC so platforms can discover
which plugins they should embed and retrieve each plugin's platform-specific
registration payload.

This helper is not meant to be invoked directly — use 'portunix plugin-registry'
through the main dispatcher.`,
}

// serveCmd starts the gRPC server.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the plugin-registry gRPC server",
	Long: `Start the PluginRegistryService gRPC server. Default transport is a unix
domain socket on Linux/macOS (filesystem permissions provide localhost
isolation) and TCP loopback on Windows.

Examples:
  portunix plugin-registry serve                                        # default transport for the platform
  portunix plugin-registry serve --mode unix --socket /tmp/ptx-reg.sock # explicit unix socket
  portunix plugin-registry serve --mode tcp --port 9500                 # explicit TCP loopback`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mode, _ := cmd.Flags().GetString("mode")
		port, _ := cmd.Flags().GetInt("port")
		bind, _ := cmd.Flags().GetString("bind")
		socket, _ := cmd.Flags().GetString("socket")
		pluginDir, _ := cmd.Flags().GetString("plugin-dir")

		if mode == "" {
			mode = defaultMode()
		}
		if mode == "unix" && socket == "" {
			socket = defaultSocketPath()
		}

		mgr, err := newRegistryManager(pluginDir)
		if err != nil {
			return fmt.Errorf("open plugin registry: %w", err)
		}
		defer mgr.Shutdown()

		lis, cleanup, err := openListener(mode, bind, port, socket)
		if err != nil {
			return err
		}
		defer cleanup()

		grpcServer := grpc.NewServer()
		pb.RegisterPluginRegistryServiceServer(grpcServer,
			server.NewPluginRegistryServer(mgr, version))

		// Graceful shutdown on SIGINT / SIGTERM.
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Fprintln(os.Stderr, "Shutting down plugin-registry gRPC server...")
			grpcServer.GracefulStop()
		}()

		fmt.Fprintf(os.Stderr, "plugin-registry gRPC server listening on %s\n", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil {
			return fmt.Errorf("grpc serve: %w", err)
		}
		return nil
	},
}

// rootCmd is the entry point when invoked via the dispatcher. The dispatcher
// passes the subcommand name as the first argument, so we register
// plugin-registry under rootCmd.
var rootCmd = &cobra.Command{
	Use:     "ptx-plugin-registry",
	Short:   "Portunix plugin-registry helper",
	Long:    `ptx-plugin-registry is a helper binary for Portunix that serves the plugin registry gRPC API.`,
	Version: version,
}

func init() {
	rootCmd.AddCommand(pluginRegistryCmd)
	pluginRegistryCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringP("mode", "m", "", "Transport: unix or tcp (default: unix on Linux/macOS, tcp on Windows)")
	serveCmd.Flags().IntP("port", "p", 9500, "TCP port (mode=tcp)")
	serveCmd.Flags().String("bind", "127.0.0.1", "TCP bind address (mode=tcp); loopback-only by default")
	serveCmd.Flags().StringP("socket", "s", "", "Unix socket path (mode=unix); defaults to $XDG_RUNTIME_DIR/portunix/plugin-registry.sock")
	serveCmd.Flags().String("plugin-dir", "", "Override plugin install directory (defaults to the system Portunix location)")

	rootCmd.SetVersionTemplate("ptx-plugin-registry version {{.Version}}\n")
}

// defaultMode picks the transport appropriate for the current platform.
func defaultMode() string {
	if runtime.GOOS == "windows" {
		return "tcp"
	}
	return "unix"
}

// defaultSocketPath returns the preferred unix socket path. $XDG_RUNTIME_DIR
// is the standard user-scoped runtime directory on Linux; on macOS the
// fallback under $HOME keeps the behaviour predictable when that variable is
// not set.
func defaultSocketPath() string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			dir = filepath.Join(home, ".portunix", "run")
		} else {
			dir = filepath.Join(os.TempDir(), "portunix")
		}
	} else {
		dir = filepath.Join(dir, "portunix")
	}
	return filepath.Join(dir, "plugin-registry.sock")
}

// openListener prepares the network listener according to the selected
// transport and returns a cleanup function (removes the unix socket on exit).
func openListener(mode, bind string, port int, socket string) (net.Listener, func(), error) {
	switch mode {
	case "tcp":
		addr := fmt.Sprintf("%s:%d", bind, port)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, nil, fmt.Errorf("listen tcp %s: %w", addr, err)
		}
		return lis, func() {}, nil
	case "unix":
		if err := os.MkdirAll(filepath.Dir(socket), 0o700); err != nil {
			return nil, nil, fmt.Errorf("create socket dir: %w", err)
		}
		// Stale socket from a previous unclean shutdown would prevent bind.
		_ = os.Remove(socket)
		lis, err := net.Listen("unix", socket)
		if err != nil {
			return nil, nil, fmt.Errorf("listen unix %s: %w", socket, err)
		}
		if err := os.Chmod(socket, 0o600); err != nil {
			lis.Close()
			return nil, nil, fmt.Errorf("chmod socket: %w", err)
		}
		return lis, func() { _ = os.Remove(socket) }, nil
	default:
		return nil, nil, fmt.Errorf("unknown mode %q (use 'unix' or 'tcp')", mode)
	}
}

// newRegistryManager opens the plugin manager against the Portunix plugin
// directory so the server can query it. Paths match src/cmd/plugin.go so the
// CLI and gRPC server see the same registry (DEC-5).
func newRegistryManager(pluginDir string) (*manager.Manager, error) {
	if pluginDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve home directory: %w", err)
		}
		pluginDir = filepath.Join(home, ".portunix", "plugins")
	}
	cfg := manager.ManagerConfig{
		PluginsDir:   pluginDir,
		RegistryFile: filepath.Join(pluginDir, "registry.json"),
		DefaultPort:  9001,
		PortRange:    manager.PortRange{Start: 9000, End: 9999},
	}
	return manager.NewManager(cfg)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.SetFlags(0)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
