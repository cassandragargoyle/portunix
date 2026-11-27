package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"google.golang.org/grpc"

	"go-plugin-template/internal/config"
	"go-plugin-template/internal/handlers"
	"go-plugin-template/internal/manifest"
	"go-plugin-template/internal/services"
	pb "go-plugin-template/proto"
)

// version is set at build time (-ldflags "-X main.version=...")
var version = "dev"

// Without a command the plugin runs as a gRPC service (portunix plugin start);
// with a command it runs once as a CLI (portunix <plugin> <command> [args...])
func main() {
	var (
		port         = flag.Int("port", 50051, "gRPC server port")
		configPath   = flag.String("config", "config.yaml", "Configuration file path")
		manifestPath = flag.String("manifest", "plugin.json", "Plugin manifest path")
		showVersion  = flag.Bool("version", false, "Print the plugin version and exit")
	)
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fail("failed to load configuration: %v", err)
	}
	configureLogging(cfg.LogLevel)

	if *showVersion {
		fmt.Printf("%s %s (build %s)\n", cfg.Plugin.Name, cfg.Plugin.Version, version)
		return
	}

	if flag.NArg() > 0 {
		runCommand(cfg, flag.Arg(0), flag.Args()[1:])
		return
	}

	// Portunix runs the plugin with the install directory as working
	// directory, so the default manifest path resolves to its plugin.json
	m, err := manifest.Load(*manifestPath)
	if err != nil {
		fail("%v", err)
	}
	if err := serve(cfg, m, *port); err != nil {
		fail("%v", err)
	}
}

// runCommand executes one command and prints its output (CLI mode)
func runCommand(cfg *config.Config, command string, args []string) {
	output, err := services.NewPluginService(cfg).Execute(context.Background(), command, args)
	if err != nil {
		fail("%v", err)
	}
	fmt.Println(output)
}

// serve runs the gRPC server implementing portunix.plugins.PluginService
// until SIGINT or SIGTERM
func serve(cfg *config.Config, m *manifest.Manifest, port int) error {
	// Portunix connects over localhost; the plugin is not exposed to the network
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	server := grpc.NewServer()
	pb.RegisterPluginServiceServer(server, handlers.NewPluginHandler(cfg, m))

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		slog.Info("Shutting down gracefully...")
		server.GracefulStop()
	}()

	slog.Info("Plugin server listening", "plugin", m.Name, "version", m.Version, "address", listener.Addr().String())
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

// configureLogging logs to stderr, so stdout carries only command output
func configureLogging(level string) {
	var l slog.Level
	if err := l.UnmarshalText([]byte(strings.ToUpper(level))); err != nil {
		l = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l})))
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
