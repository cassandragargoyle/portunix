package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"

	"go-plugin-template/internal/config"
	"go-plugin-template/internal/services"
	pb "go-plugin-template/proto"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "plugin_requests_total",
			Help: "Total number of requests processed",
		},
		[]string{"method", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "plugin_request_duration_seconds",
			Help: "Duration of requests",
		},
		[]string{"method"},
	)
)

type PluginHandler struct {
	pb.UnimplementedPluginServiceServer
	config  *config.Config
	service *services.PluginService
}

func NewPluginHandler(cfg *config.Config) *PluginHandler {
	return &PluginHandler{
		config:  cfg,
		service: services.NewPluginService(cfg),
	}
}

func (h *PluginHandler) GetInfo(ctx context.Context, req *pb.GetInfoRequest) (*pb.GetInfoResponse, error) {
	timer := prometheus.NewTimer(requestDuration.WithLabelValues("get_info"))
	defer timer.ObserveDuration()

	logrus.Debug("GetInfo called")

	requestsTotal.WithLabelValues("get_info", "success").Inc()

	return &pb.GetInfoResponse{
		Info: &pb.PluginInfo{
			Name:         h.config.Plugin.Name,
			Version:      h.config.Plugin.Version,
			Description:  h.config.Plugin.Description,
			Capabilities: []string{"example-capability"},
		},
	}, nil
}

func (h *PluginHandler) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	timer := prometheus.NewTimer(requestDuration.WithLabelValues("health_check"))
	defer timer.ObserveDuration()

	status := h.service.CheckHealth()

	requestsTotal.WithLabelValues("health_check", "success").Inc()

	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_Status(status),
	}, nil
}

func (h *PluginHandler) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	timer := prometheus.NewTimer(requestDuration.WithLabelValues("execute"))
	defer timer.ObserveDuration()

	logrus.WithFields(logrus.Fields{
		"command": req.Command,
		"args":    req.Args,
	}).Info("Executing command")

	// Add timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.Timeout)*time.Second)
	defer cancel()

	result, err := h.service.Execute(timeoutCtx, req.Command, req.Args)
	if err != nil {
		logrus.WithError(err).Error("Command execution failed")
		requestsTotal.WithLabelValues("execute", "error").Inc()
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	requestsTotal.WithLabelValues("execute", "success").Inc()

	return &pb.ExecuteResponse{
		Result: result,
		Status: pb.ExecuteResponse_SUCCESS,
	}, nil
}

func (h *PluginHandler) ListTools(ctx context.Context, req *pb.ListToolsRequest) (*pb.ListToolsResponse, error) {
	timer := prometheus.NewTimer(requestDuration.WithLabelValues("list_tools"))
	defer timer.ObserveDuration()

	tools := []*pb.MCPTool{
		{
			Name:        "example_tool",
			Description: "Example tool for demonstration",
			Schema: `{
				"type": "object",
				"properties": {
					"input": {"type": "string", "description": "Input parameter"}
				},
				"required": ["input"]
			}`,
		},
	}

	requestsTotal.WithLabelValues("list_tools", "success").Inc()

	return &pb.ListToolsResponse{Tools: tools}, nil
}

func (h *PluginHandler) CallTool(ctx context.Context, req *pb.CallToolRequest) (*pb.CallToolResponse, error) {
	timer := prometheus.NewTimer(requestDuration.WithLabelValues("call_tool"))
	defer timer.ObserveDuration()

	logrus.WithFields(logrus.Fields{
		"tool": req.ToolName,
		"args": req.Arguments,
	}).Info("Calling MCP tool")

	switch req.ToolName {
	case "example_tool":
		return h.handleExampleTool(ctx, req.Arguments)
	default:
		requestsTotal.WithLabelValues("call_tool", "error").Inc()
		return nil, fmt.Errorf("unknown tool: %s", req.ToolName)
	}
}

func (h *PluginHandler) handleExampleTool(ctx context.Context, args string) (*pb.CallToolResponse, error) {
	var params struct {
		Input string `json:"input"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := h.service.ProcessExample(ctx, params.Input)
	if err != nil {
		requestsTotal.WithLabelValues("call_tool", "error").Inc()
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	requestsTotal.WithLabelValues("call_tool", "success").Inc()

	return &pb.CallToolResponse{
		Result: result,
		Status: pb.CallToolResponse_SUCCESS,
	}, nil
}

func (h *PluginHandler) Shutdown(ctx context.Context, req *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	logrus.Info("Shutdown requested")

	h.service.Cleanup()

	return &pb.ShutdownResponse{
		Status: pb.ShutdownResponse_SUCCESS,
	}, nil
}

func (h *PluginHandler) Shutdown() {
	h.service.Cleanup()
}