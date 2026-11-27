// Package handlers implements portunix.plugins.PluginService
package handlers

import (
	"context"
	"log/slog"
	"time"

	"go-plugin-template/internal/config"
	"go-plugin-template/internal/manifest"
	"go-plugin-template/internal/services"
	pb "go-plugin-template/proto"
)

// PluginHandler implements portunix.plugins.PluginService; plugin info and
// commands come from plugin.json, so the manifest stays the single source
type PluginHandler struct {
	pb.UnimplementedPluginServiceServer
	config    *config.Config
	manifest  *manifest.Manifest
	service   *services.PluginService
	startedAt time.Time
}

// NewPluginHandler creates the gRPC handler
func NewPluginHandler(cfg *config.Config, m *manifest.Manifest) *PluginHandler {
	return &PluginHandler{
		config:    cfg,
		manifest:  m,
		service:   services.NewPluginService(cfg),
		startedAt: time.Now(),
	}
}

// Initialize is the first call Portunix makes after starting the plugin
func (h *PluginHandler) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	slog.Info("Initialize called", "plugin", req.GetPluginName(), "version", req.GetVersion())
	return &pb.InitializeResponse{
		Success:    true,
		Message:    "initialized",
		PluginInfo: h.manifest.PluginInfo(),
	}, nil
}

// GetInfo returns the plugin information from plugin.json
func (h *PluginHandler) GetInfo(ctx context.Context, req *pb.GetInfoRequest) (*pb.GetInfoResponse, error) {
	return &pb.GetInfoResponse{PluginInfo: h.manifest.PluginInfo()}, nil
}

// Health reports the plugin state for portunix plugin health
func (h *PluginHandler) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	healthy := h.service.CheckHealth()
	status, message := "running", "ok"
	if !healthy {
		status, message = "unhealthy", "health check failed"
	}
	return &pb.HealthResponse{
		Healthy:       healthy,
		Status:        status,
		Message:       message,
		UptimeSeconds: int64(time.Since(h.startedAt).Seconds()),
		Metrics:       map[string]string{},
	}, nil
}

// Execute runs a plugin command; command errors are reported in the
// response, not as gRPC errors, so Portunix can show them to the user
func (h *PluginHandler) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	slog.Info("Executing command", "command", req.GetCommand(), "args", req.GetArgs())

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.Timeout)*time.Second)
	defer cancel()

	output, err := h.service.Execute(timeoutCtx, req.GetCommand(), req.GetArgs())
	if err != nil {
		slog.Error("Command execution failed", "error", err)
		return &pb.ExecuteResponse{
			Success:  false,
			Message:  err.Error(),
			Error:    err.Error(),
			ExitCode: 1,
		}, nil
	}
	return &pb.ExecuteResponse{Success: true, Output: output, ExitCode: 0}, nil
}

// ListCommands returns the commands declared in plugin.json
func (h *PluginHandler) ListCommands(ctx context.Context, req *pb.ListCommandsRequest) (*pb.ListCommandsResponse, error) {
	return &pb.ListCommandsResponse{
		Commands:     h.manifest.PluginCommands(),
		Capabilities: h.manifest.Capabilities(),
	}, nil
}

// Shutdown only cleans up and answers; Portunix stops the process right after
// this call (SIGTERM, handled in main.go), exiting here would drop the response
func (h *PluginHandler) Shutdown(ctx context.Context, req *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	slog.Info("Shutdown requested")
	h.service.Cleanup()
	return &pb.ShutdownResponse{Success: true, Message: "cleanup done"}, nil
}
