/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package plugins

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "portunix.ai/app/plugins/proto"
)

// GRPCPlugin implements the Plugin interface using gRPC communication
type GRPCPlugin struct {
	config     PluginConfig
	info       PluginInfo
	conn       *grpc.ClientConn
	client     pb.PluginServiceClient
	cmd        *exec.Cmd
	exited     *processExit // set for a child process started here
	logPath    string       // plugin stdout and stderr
	pid        int          // plugin process, also set when attached to a process started elsewhere
	isRunning  bool
	lastHealth time.Time
}

const (
	// startConnectTimeout bounds the wait for a freshly started plugin to listen
	startConnectTimeout = 10 * time.Second
	// attachConnectTimeout bounds the connection to an already running plugin
	attachConnectTimeout = 3 * time.Second
	// stopTimeout bounds the wait for the plugin process to exit
	stopTimeout = 5 * time.Second
	// PluginLogFile receives the output of a plugin process in its working directory
	PluginLogFile = "plugin.log"
)

// processExit reports the end of a plugin child process
type processExit struct {
	done chan struct{} // closed once the process has been reaped
	err  error         // result of Wait, valid after done is closed
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

	// Another process on the port would answer in place of this plugin
	if err := PortAvailable(p.config.Port); err != nil {
		return err
	}

	// Start plugin binary
	if err := p.startPluginBinary(ctx); err != nil {
		return fmt.Errorf("failed to start plugin binary: %w", err)
	}

	// Connect to plugin via gRPC (blocks until the plugin listens or exits)
	if err := p.connectGRPC(ctx, startConnectTimeout); err != nil {
		if exitErr := p.startupExitError(); exitErr != nil {
			p.abortStart()
			return exitErr
		}
		p.abortStart()
		return fmt.Errorf("failed to connect to plugin: %w", err)
	}

	// Initialize plugin
	if err := p.initializePlugin(ctx); err != nil {
		p.abortStart()
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// The answer must come from the process started here, not from one that took the port
	if exitErr := p.startupExitError(); exitErr != nil {
		p.abortStart()
		return exitErr
	}

	p.isRunning = true
	return nil
}

// abortStart releases the connection and the process of a failed start
func (p *GRPCPlugin) abortStart() {
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
		p.client = nil
	}
	_ = p.stopPluginBinary()
	p.pid = 0
}

// startupExitError returns an error with the exit status and the last output
// lines when the plugin process has already exited, nil while it runs
func (p *GRPCPlugin) startupExitError() error {
	if p.exited == nil {
		return nil
	}
	select {
	case <-p.exited.done:
	default:
		return nil
	}

	msg := fmt.Sprintf("plugin %s exited during startup", p.config.Name)
	if p.exited.err != nil {
		msg += fmt.Sprintf(" (%v)", p.exited.err)
	}
	if output := lastLines(p.logPath, 10); output != "" {
		msg += "\n   Plugin output (" + p.logPath + "):\n" + output
	}
	return errors.New(msg)
}

// lastLines returns the last n non-empty lines of a file, indented
func lastLines(path string, n int) string {
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, "      "+strings.TrimRight(line, "\r"))
		}
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// PortAvailable fails when another process already listens on the local port
func PortAvailable(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("port %d is already in use by another process", port)
	}
	return listener.Close()
}

// Attach connects to a plugin process started by another Portunix process;
// the process is tracked even if its gRPC endpoint does not answer, so it
// can still be stopped
func (p *GRPCPlugin) Attach(ctx context.Context, pid int) error {
	if !ProcessAlive(pid) {
		return fmt.Errorf("plugin %s process %d is not running", p.config.Name, pid)
	}

	p.pid = pid
	p.isRunning = true

	if err := p.connectGRPC(ctx, attachConnectTimeout); err != nil {
		return fmt.Errorf("plugin %s process %d is running but not reachable: %w", p.config.Name, pid, err)
	}
	return nil
}

// Detach closes the gRPC connection and leaves the plugin process running
func (p *GRPCPlugin) Detach() {
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
		p.client = nil
	}
	p.cmd = nil
	p.exited = nil
	p.pid = 0
	p.isRunning = false
}

// PID returns the plugin process ID, 0 when not running
func (p *GRPCPlugin) PID() int {
	return p.pid
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

	p.pid = 0
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
	if p.isRunning && p.client == nil {
		return PluginHealth{
			Healthy:       false,
			Status:        "unreachable",
			Message:       fmt.Sprintf("Plugin process %d is running but its gRPC endpoint does not respond", p.pid),
			LastCheckTime: time.Now(),
		}
	}
	if !p.isRunning {
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

// startPluginBinary starts the plugin binary process detached from the
// caller, so the plugin keeps running after the CLI process exits
func (p *GRPCPlugin) startPluginBinary(ctx context.Context) error {
	var cmd string
	var args []string

	runtime := p.config.Runtime
	if runtime == "" {
		runtime = "native"
	}

	switch runtime {
	case "java":
		// For Java plugins, use java -jar to run the JAR file
		cmd = "java"
		// Add JVM arguments first
		if len(p.config.JVMArgs) > 0 {
			args = append(args, p.config.JVMArgs...)
		} else {
			// Default JVM arguments
			args = append(args, "-Xmx256m", "-Xms64m")
		}
		// Add JAR file
		args = append(args, "-jar", p.config.BinaryPath)
		// Add plugin arguments
		args = append(args, "--port", strconv.Itoa(p.config.Port))

	case "python":
		// For Python plugins, use python to run the script
		cmd = "python3"
		args = []string{p.config.BinaryPath, "--port", strconv.Itoa(p.config.Port)}

	default: // native
		// For native plugins, run the binary directly
		cmd = p.config.BinaryPath
		args = []string{"--port", strconv.Itoa(p.config.Port)}
	}

	p.cmd = exec.Command(cmd, args...) // #nosec G204 -- binary and runtime come from the installed plugin manifest
	p.cmd.Dir = p.config.WorkingDir
	p.cmd.SysProcAttr = detachedSysProcAttr()

	// Inherit current environment
	p.cmd.Env = os.Environ()

	// Set additional environment variables
	if p.config.Environment != nil {
		for key, value := range p.config.Environment {
			p.cmd.Env = append(p.cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Plugin output goes to a log file: a pipe would break once the CLI process exits
	p.logPath = ""
	if p.config.WorkingDir != "" {
		logPath := filepath.Join(p.config.WorkingDir, PluginLogFile)
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err == nil {
			defer logFile.Close() // the child keeps its own handle
			p.cmd.Stdout = logFile
			p.cmd.Stderr = logFile
			p.logPath = logPath
		}
	}

	// Log plugin startup for debugging
	log.Printf("Starting %s plugin: %s %v", runtime, cmd, args)

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start plugin binary (%s): %w", runtime, err)
	}
	p.pid = p.cmd.Process.Pid

	// Reap the child whenever it exits, also after Detach, so it never stays a zombie
	exited := &processExit{done: make(chan struct{})}
	go func(cmd *exec.Cmd) {
		exited.err = cmd.Wait()
		close(exited.done)
	}(p.cmd)
	p.exited = exited

	return nil
}

// stopPluginBinary stops the plugin binary process
func (p *GRPCPlugin) stopPluginBinary() error {
	if p.cmd == nil || p.cmd.Process == nil || p.exited == nil {
		return p.stopAttachedProcess()
	}

	// Already exited, e.g. after the gRPC Shutdown request
	select {
	case <-p.exited.done:
		return nil
	default:
	}

	// Try graceful termination first
	if err := p.cmd.Process.Signal(os.Interrupt); err != nil {
		// Force kill if graceful termination fails
		return p.cmd.Process.Kill()
	}

	// Wait for process to exit
	select {
	case <-p.exited.done:
		return nil
	case <-time.After(stopTimeout):
		// Force kill after timeout
		return p.cmd.Process.Kill()
	}
}

// stopAttachedProcess stops a plugin process that is not a child of this
// process and can therefore only be polled, not waited for
func (p *GRPCPlugin) stopAttachedProcess() error {
	if !ProcessAlive(p.pid) {
		return nil
	}

	if err := interruptProcess(p.pid); err != nil && ProcessAlive(p.pid) {
		return fmt.Errorf("failed to stop plugin process %d: %w", p.pid, err)
	}

	deadline := time.Now().Add(stopTimeout)
	for time.Now().Before(deadline) {
		if !ProcessAlive(p.pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Force kill after timeout
	proc, err := os.FindProcess(p.pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

// connectGRPC establishes gRPC connection to the plugin, waiting at most timeout
func (p *GRPCPlugin) connectGRPC(ctx context.Context, timeout time.Duration) error {
	address := fmt.Sprintf("localhost:%d", p.config.Port)

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// A start ends as soon as the plugin process exits
	if exited := p.exited; exited != nil {
		go func() {
			select {
			case <-exited.done:
				cancel()
			case <-dialCtx.Done():
			}
		}()
	}

	conn, err := grpc.DialContext(
		dialCtx,
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
