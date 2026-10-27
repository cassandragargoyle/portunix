/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// pluginTemplateProto holds the host plugin protocol and its generated Go code,
// copied from src/app/plugins/proto; the .tmpl suffix keeps the Go files out
// of this module's build. TestPluginTemplateProtoInSync guards the copies
//
//go:embed plugintemplate/plugin.proto plugintemplate/plugin.pb.go.tmpl plugintemplate/plugin_grpc.pb.go.tmpl
var pluginTemplateProto embed.FS

// writePluginTemplateProto writes the host plugin protocol into proto/ of a new plugin
func writePluginTemplateProto(pluginDir string) error {
	files := map[string]string{
		"plugintemplate/plugin.proto":           "plugin.proto",
		"plugintemplate/plugin.pb.go.tmpl":      "plugin.pb.go",
		"plugintemplate/plugin_grpc.pb.go.tmpl": "plugin_grpc.pb.go",
	}
	for src, dst := range files {
		data, err := pluginTemplateProto.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read embedded %s: %w", src, err)
		}
		if err := os.WriteFile(filepath.Join(pluginDir, "proto", dst), data, 0644); err != nil { // #nosec G306 -- scaffolded project files use regular source tree permissions
			return err
		}
	}
	return nil
}

// pluginTemplateMain returns main.go of a new plugin: a gRPC server that
// implements the Portunix plugin protocol with example commands
func pluginTemplateMain(pluginName string) string {
	return strings.ReplaceAll(pluginTemplateMainSource, "__PLUGIN_NAME__", pluginName)
}

const pluginTemplateMainSource = `package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"

	pb "__PLUGIN_NAME__/proto"
)

const (
	pluginName    = "__PLUGIN_NAME__"
	pluginVersion = "1.0.0"
)

// pluginServer implements the Portunix plugin protocol (proto/plugin.proto);
// Portunix starts the binary with --port, calls Initialize, then Execute,
// Health and finally Shutdown
type pluginServer struct {
	pb.UnimplementedPluginServiceServer
	server  *grpc.Server
	started time.Time
}

// commands lists the commands this plugin provides; keep plugin.json in sync
func commands() []*pb.PluginCommand {
	return []*pb.PluginCommand{
		{Name: "help", Description: "Show help for " + pluginName + " plugin"},
		{Name: "hello", Description: "Print a greeting", Examples: []string{"hello World"}},
	}
}

func info() *pb.PluginInfo {
	return &pb.PluginInfo{
		Name:        pluginName,
		Version:     pluginVersion,
		SupportedOs: []string{"linux", "windows", "darwin"},
		Commands:    commands(),
	}
}

// Initialize is called by Portunix right after the plugin process starts
func (s *pluginServer) Initialize(_ context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	log.Printf("initialized by %s", req.GetPluginName())
	return &pb.InitializeResponse{Success: true, Message: "initialized", PluginInfo: info()}, nil
}

// Execute runs one plugin command
func (s *pluginServer) Execute(_ context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	switch req.GetCommand() {
	case "help":
		var lines []string
		for _, c := range commands() {
			lines = append(lines, fmt.Sprintf("%-10s %s", c.GetName(), c.GetDescription()))
		}
		return &pb.ExecuteResponse{Success: true, Output: strings.Join(lines, "\n")}, nil
	case "hello":
		name := "World"
		if len(req.GetArgs()) > 0 {
			name = req.GetArgs()[0]
		}
		return &pb.ExecuteResponse{Success: true, Output: fmt.Sprintf("Hello, %s!", name)}, nil
	default:
		return &pb.ExecuteResponse{Success: false, Error: "unknown command: " + req.GetCommand(), ExitCode: 1}, nil
	}
}

func (s *pluginServer) GetInfo(context.Context, *pb.GetInfoRequest) (*pb.GetInfoResponse, error) {
	return &pb.GetInfoResponse{PluginInfo: info()}, nil
}

func (s *pluginServer) ListCommands(context.Context, *pb.ListCommandsRequest) (*pb.ListCommandsResponse, error) {
	return &pb.ListCommandsResponse{Commands: commands()}, nil
}

// Health is polled by "portunix plugin health"
func (s *pluginServer) Health(context.Context, *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Healthy:       true,
		Status:        "running",
		Message:       pluginName + " is running",
		UptimeSeconds: int64(time.Since(s.started).Seconds()),
	}, nil
}

// Shutdown is called by "portunix plugin stop" before the process is ended
func (s *pluginServer) Shutdown(context.Context, *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	go s.server.GracefulStop()
	return &pb.ShutdownResponse{Success: true, Message: "shutting down"}, nil
}

func main() {
	port := flag.Int("port", 9001, "gRPC port assigned by Portunix")
	flag.Parse()

	// Listen on loopback only: the plugin serves the local Portunix process
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", *port, err)
	}

	server := grpc.NewServer()
	pb.RegisterPluginServiceServer(server, &pluginServer{server: server, started: time.Now()})

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		server.GracefulStop()
	}()

	log.Printf("%s plugin listening on port %d", pluginName, *port)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
`
