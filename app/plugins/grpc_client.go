package plugins

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "portunix.cz/app/plugins/proto"
)

// GRPCPlugin implements the Plugin interface using gRPC communication
type GRPCPlugin struct {
	config     PluginConfig
	info       PluginInfo
	conn       *grpc.ClientConn
	client     pb.PluginServiceClient
	cmd        *exec.Cmd
	isRunning  bool
	lastHealth time.Time
}

// NewGRPCPlugin creates a new gRPC plugin instance
func NewGRPCPlugin(config PluginConfig) *GRPCPlugin {
	return &GRPCPlugin{
		config:    config,
		isRunning: false,
	}
}

// Initialize plugin with configuration
func (p *GRPCPlugin) Initialize(ctx context.Context, config PluginConfig) error {
	p.config = config
	return nil
}

// Start plugin service
func (p *GRPCPlugin) Start(ctx context.Context) error {
	if p.isRunning {
		return fmt.Errorf("plugin %s is already running", p.config.Name)
	}

	// Start plugin binary
	if err := p.startPluginBinary(ctx); err != nil {
		return fmt.Errorf("failed to start plugin binary: %w", err)
	}

	// Wait for plugin to start
	time.Sleep(2 * time.Second)

	// Connect to plugin via gRPC
	if err := p.connectGRPC(ctx); err != nil {
		p.stopPluginBinary()
		return fmt.Errorf("failed to connect to plugin: %w", err)
	}

	// Initialize plugin
	if err := p.initializePlugin(ctx); err != nil {
		p.Stop(ctx)
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	p.isRunning = true
	return nil
}

// Stop plugin service
func (p *GRPCPlugin) Stop(ctx context.Context) error {
	if !p.isRunning {
		return nil
	}

	var lastErr error

	// Try graceful shutdown via gRPC
	if p.client != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_, err := p.client.Shutdown(shutdownCtx, &pb.ShutdownRequest{
			Force:          false,
			TimeoutSeconds: 5,
		})
		if err != nil {
			lastErr = err
			log.Printf("Failed to shutdown plugin gracefully: %v", err)
		}
	}

	// Close gRPC connection
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			lastErr = err
			log.Printf("Failed to close gRPC connection: %v", err)
		}
		p.conn = nil
		p.client = nil
	}

	// Stop plugin binary
	if err := p.stopPluginBinary(); err != nil {
		lastErr = err
		log.Printf("Failed to stop plugin binary: %v", err)
	}

	p.isRunning = false
	return lastErr
}

// Execute a command
func (p *GRPCPlugin) Execute(ctx context.Context, request ExecuteRequest) (*ExecuteResponse, error) {
	if !p.isRunning || p.client == nil {
		return nil, fmt.Errorf("plugin %s is not running", p.config.Name)
	}

	grpcRequest := &pb.ExecuteRequest{
		Command:          request.Command,
		Args:             request.Args,
		Options:          request.Options,
		Environment:      request.Environment,
		WorkingDirectory: request.WorkingDirectory,
	}

	grpcResponse, err := p.client.Execute(ctx, grpcRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return &ExecuteResponse{
		Success:  grpcResponse.Success,
		Message:  grpcResponse.Message,
		Output:   grpcResponse.Output,
		Error:    grpcResponse.Error,
		ExitCode: int(grpcResponse.ExitCode),
		Metadata: grpcResponse.Metadata,
	}, nil
}

// GetInfo returns plugin information
func (p *GRPCPlugin) GetInfo() PluginInfo {
	return p.info
}

// Health check
func (p *GRPCPlugin) Health(ctx context.Context) PluginHealth {
	if !p.isRunning || p.client == nil {
		return PluginHealth{
			Healthy:       false,
			Status:        "not_running",
			Message:       "Plugin is not running",
			LastCheckTime: time.Now(),
		}
	}

	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := p.client.Health(healthCtx, &pb.HealthRequest{})
	if err != nil {
		return PluginHealth{
			Healthy:       false,
			Status:        "error",
			Message:       err.Error(),
			LastCheckTime: time.Now(),
		}
	}

	p.lastHealth = time.Now()
	return PluginHealth{
		Healthy:       response.Healthy,
		Status:        response.Status,
		Message:       response.Message,
		UptimeSeconds: response.UptimeSeconds,
		Metrics:       response.Metrics,
		LastCheckTime: p.lastHealth,
	}
}

// IsRunning checks if plugin is running
func (p *GRPCPlugin) IsRunning() bool {
	return p.isRunning
}

// startPluginBinary starts the plugin binary process
func (p *GRPCPlugin) startPluginBinary(ctx context.Context) error {
	args := []string{"--port", strconv.Itoa(p.config.Port)}

	p.cmd = exec.CommandContext(ctx, p.config.BinaryPath, args...)
	p.cmd.Dir = p.config.WorkingDir

	// Set environment variables
	if p.config.Environment != nil {
		for key, value := range p.config.Environment {
			p.cmd.Env = append(p.cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start plugin binary: %w", err)
	}

	return nil
}

// stopPluginBinary stops the plugin binary process
func (p *GRPCPlugin) stopPluginBinary() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	// Try graceful termination first
	if err := p.cmd.Process.Signal(os.Interrupt); err != nil {
		// Force kill if graceful termination fails
		return p.cmd.Process.Kill()
	}

	// Wait for process to exit
	done := make(chan error, 1)
	go func() {
		done <- p.cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		// Force kill after timeout
		return p.cmd.Process.Kill()
	}
}

// connectGRPC establishes gRPC connection to the plugin
func (p *GRPCPlugin) connectGRPC(ctx context.Context) error {
	address := fmt.Sprintf("localhost:%d", p.config.Port)

	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to plugin at %s: %w", address, err)
	}

	p.conn = conn
	p.client = pb.NewPluginServiceClient(conn)
	return nil
}

// initializePlugin sends initialization request to the plugin
func (p *GRPCPlugin) initializePlugin(ctx context.Context) error {
	request := &pb.InitializeRequest{
		PluginName:  p.config.Name,
		Version:     p.config.Version,
		Config:      make(map[string]string),
		Environment: p.config.Environment,
		Permissions: &pb.PluginPermissions{
			Filesystem: p.config.Permissions.Filesystem,
			Network:    p.config.Permissions.Network,
			Database:   p.config.Permissions.Database,
			System:     p.config.Permissions.System,
			Level:      p.config.Permissions.Level,
		},
	}

	response, err := p.client.Initialize(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("plugin initialization failed: %s", response.Message)
	}

	// Update plugin info from response
	if response.PluginInfo != nil {
		p.info = PluginInfo{
			Name:        response.PluginInfo.Name,
			Version:     response.PluginInfo.Version,
			Description: response.PluginInfo.Description,
			Author:      response.PluginInfo.Author,
			License:     response.PluginInfo.License,
			SupportedOS: response.PluginInfo.SupportedOs,
			Status:      PluginStatusRunning,
			LastSeen:    time.Now(),
		}

		// Convert commands
		for _, cmd := range response.PluginInfo.Commands {
			pluginCmd := PluginCommand{
				Name:        cmd.Name,
				Description: cmd.Description,
				Subcommands: cmd.Subcommands,
				Examples:    cmd.Examples,
			}

			// Convert parameters
			for _, param := range cmd.Parameters {
				pluginCmd.Parameters = append(pluginCmd.Parameters, PluginParameter{
					Name:         param.Name,
					Type:         param.Type,
					Description:  param.Description,
					Required:     param.Required,
					DefaultValue: param.DefaultValue,
				})
			}

			p.info.Commands = append(p.info.Commands, pluginCmd)
		}

		// Convert capabilities
		if response.PluginInfo.Capabilities != nil {
			p.info.Capabilities = PluginCapabilities{
				FilesystemAccess: response.PluginInfo.Capabilities.FilesystemAccess,
				NetworkAccess:    response.PluginInfo.Capabilities.NetworkAccess,
				DatabaseAccess:   response.PluginInfo.Capabilities.DatabaseAccess,
				ContainerAccess:  response.PluginInfo.Capabilities.ContainerAccess,
				SystemCommands:   response.PluginInfo.Capabilities.SystemCommands,
				MCPTools:         response.PluginInfo.Capabilities.McpTools,
			}
		}

		// Convert required permissions
		if response.PluginInfo.RequiredPermissions != nil {
			p.info.RequiredPermissions = PluginPermissions{
				Filesystem: response.PluginInfo.RequiredPermissions.Filesystem,
				Network:    response.PluginInfo.RequiredPermissions.Network,
				Database:   response.PluginInfo.RequiredPermissions.Database,
				System:     response.PluginInfo.RequiredPermissions.System,
				Level:      response.PluginInfo.RequiredPermissions.Level,
			}
		}
	}

	return nil
}
