package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

const (
	DefaultHealthTimeout = 10 * time.Second
)

// AllocationMode defines how service instances are allocated
type AllocationMode string

const (
	ModeShared          AllocationMode = "shared"
	ModeExclusive       AllocationMode = "exclusive"
	ModePreferExclusive AllocationMode = "prefer-exclusive"
)

// StartResult contains the result of a service start operation
type StartResult struct {
	SessionID  string   `json:"session_id"`
	Endpoint   string   `json:"endpoint"`
	Port       int      `json:"port"`
	PID        int      `json:"pid"`
	Mode       string   `json:"mode"`
	Services   []string `json:"services"`
	NewProcess bool     `json:"new_process"`
}

// Orchestrator manages service lifecycle
type Orchestrator struct {
	stateManager  *StateManager
	portAllocator *PortAllocator
	healthTimeout time.Duration
}

// NewOrchestrator creates a new service orchestrator
func NewOrchestrator() (*Orchestrator, error) {
	sm, err := NewStateManager()
	if err != nil {
		return nil, err
	}

	return &Orchestrator{
		stateManager:  sm,
		portAllocator: NewPortAllocator(DefaultPortStart, DefaultPortEnd),
		healthTimeout: DefaultHealthTimeout,
	}, nil
}

// StartService starts or joins a plugin service instance
func (o *Orchestrator) StartService(pluginName, binaryPath, runtime string, jvmArgs []string, mode AllocationMode) (*StartResult, error) {
	var result *StartResult

	err := o.stateManager.WithLock(func(state *ProcessState) error {
		// Clean stale entries
		o.cleanStaleEntries(state)

		// In shared mode, try to join existing instance
		if mode == ModeShared || mode == ModePreferExclusive {
			for key, inst := range state.Instances {
				if inst.PluginName == pluginName && inst.Mode == string(ModeShared) {
					// Verify it's actually alive
					if !IsProcessAlive(inst.PID) {
						delete(state.Instances, key)
						continue
					}
					// Verify via gRPC health check
					if !o.checkHealth(inst.Port) {
						// Process alive but not serving - remove
						killProcess(inst.PID)
						delete(state.Instances, key)
						continue
					}
					// Join existing instance
					sessionID := GenerateSessionID()
					inst.Sessions = append(inst.Sessions, Session{
						ID:        sessionID,
						CreatedAt: time.Now(),
					})
					state.Instances[key] = inst
					result = &StartResult{
						SessionID:  sessionID,
						Endpoint:   fmt.Sprintf("localhost:%d", inst.Port),
						Port:       inst.Port,
						PID:        inst.PID,
						Mode:       inst.Mode,
						Services:   inst.Services,
						NewProcess: false,
					}
					return nil
				}
			}
		}

		// Allocate port
		port, err := o.portAllocator.AllocatePort(state)
		if err != nil {
			return err
		}

		// Spawn process
		pid, err := o.spawnPlugin(binaryPath, runtime, jvmArgs, port)
		if err != nil {
			return fmt.Errorf("failed to spawn plugin: %w", err)
		}

		// Wait for health check
		services, err := o.waitForHealthy(port)
		if err != nil {
			killProcess(pid)
			return fmt.Errorf("plugin failed to become healthy: %w", err)
		}

		// Determine effective mode
		effectiveMode := string(mode)
		if mode == ModePreferExclusive {
			effectiveMode = string(ModeExclusive)
		}
		if mode == ModeShared {
			effectiveMode = string(ModeShared)
		}

		sessionID := GenerateSessionID()
		key := InstanceKey(pluginName, port)
		state.Instances[key] = Instance{
			PluginName: pluginName,
			PID:        pid,
			Port:       port,
			Mode:       effectiveMode,
			Sessions: []Session{{
				ID:        sessionID,
				CreatedAt: time.Now(),
			}},
			Services:  services,
			StartedAt: time.Now(),
		}

		result = &StartResult{
			SessionID:  sessionID,
			Endpoint:   fmt.Sprintf("localhost:%d", port),
			Port:       port,
			PID:        pid,
			Mode:       effectiveMode,
			Services:   services,
			NewProcess: true,
		}
		return nil
	})

	return result, err
}

// ListInstances returns all running instances with health validation
func (o *Orchestrator) ListInstances() (map[string]Instance, error) {
	var instances map[string]Instance

	err := o.stateManager.WithLock(func(state *ProcessState) error {
		o.cleanStaleEntries(state)
		instances = state.Instances
		return nil
	})

	return instances, err
}

// GetInstanceInfo returns gRPC service details for a specific instance
func (o *Orchestrator) GetInstanceInfo(pluginName string, port int) (*Instance, []ServiceMethod, error) {
	state, err := o.stateManager.ReadState()
	if err != nil {
		return nil, nil, err
	}

	key := InstanceKey(pluginName, port)
	inst, exists := state.Instances[key]
	if !exists {
		return nil, nil, fmt.Errorf("no instance found for %s on port %d", pluginName, port)
	}

	methods, err := o.discoverMethods(port)
	if err != nil {
		return &inst, nil, fmt.Errorf("failed to discover methods: %w", err)
	}

	return &inst, methods, nil
}

// ServiceMethod represents a gRPC service method
type ServiceMethod struct {
	Service string `json:"service"`
	Method  string `json:"method"`
}

// ReleaseSession removes a session and stops the instance if it was the last
func (o *Orchestrator) ReleaseSession(sessionID string) (bool, error) {
	var stopped bool

	err := o.stateManager.WithLock(func(state *ProcessState) error {
		for key, inst := range state.Instances {
			for i, sess := range inst.Sessions {
				if sess.ID == sessionID {
					// Remove session
					inst.Sessions = append(inst.Sessions[:i], inst.Sessions[i+1:]...)
					if len(inst.Sessions) == 0 {
						// Last session - stop instance
						killProcess(inst.PID)
						delete(state.Instances, key)
						stopped = true
					} else {
						state.Instances[key] = inst
					}
					return nil
				}
			}
		}
		return fmt.Errorf("session %s not found", sessionID)
	})

	return stopped, err
}

// StopPlugin forcefully stops all instances of a plugin
func (o *Orchestrator) StopPlugin(pluginName string) (int, error) {
	var count int

	err := o.stateManager.WithLock(func(state *ProcessState) error {
		for key, inst := range state.Instances {
			if inst.PluginName == pluginName {
				killProcess(inst.PID)
				delete(state.Instances, key)
				count++
			}
		}
		if count == 0 {
			return fmt.Errorf("no running instances of %s", pluginName)
		}
		return nil
	})

	return count, err
}

// StopAll forcefully stops all running service instances
func (o *Orchestrator) StopAll() (int, error) {
	var count int

	err := o.stateManager.WithLock(func(state *ProcessState) error {
		for key, inst := range state.Instances {
			killProcess(inst.PID)
			delete(state.Instances, key)
			count++
		}
		return nil
	})

	return count, err
}

// cleanStaleEntries removes entries for dead processes
func (o *Orchestrator) cleanStaleEntries(state *ProcessState) {
	for key, inst := range state.Instances {
		if !IsProcessAlive(inst.PID) {
			delete(state.Instances, key)
			continue
		}
		if !o.checkHealth(inst.Port) {
			killProcess(inst.PID)
			delete(state.Instances, key)
		}
	}
}

// spawnPlugin starts a plugin binary as a detached process
func (o *Orchestrator) spawnPlugin(binaryPath, runtime string, jvmArgs []string, port int) (int, error) {
	var cmd *exec.Cmd

	if runtime == "" {
		runtime = "native"
	}

	switch runtime {
	case "java":
		args := []string{}
		if len(jvmArgs) > 0 {
			args = append(args, jvmArgs...)
		} else {
			args = append(args, "-Xmx256m", "-Xms64m")
		}
		args = append(args, "-jar", binaryPath, "--grpc-port", strconv.Itoa(port))
		cmd = exec.Command("java", args...)

	case "python":
		cmd = exec.Command(binaryPath, "--grpc-port", strconv.Itoa(port))

	default: // native
		cmd = exec.Command(binaryPath, "--grpc-port", strconv.Itoa(port))
	}

	// Detach from parent process group
	cmd.SysProcAttr = platformSysProcAttr()
	cmd.Dir = filepath.Dir(binaryPath)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start process: %w", err)
	}

	// Release the process so it survives parent exit
	pid := cmd.Process.Pid
	cmd.Process.Release()

	return pid, nil
}

// waitForHealthy polls the gRPC health endpoint until SERVING
func (o *Orchestrator) waitForHealthy(port int) ([]string, error) {
	deadline := time.Now().Add(o.healthTimeout)

	for time.Now().Before(deadline) {
		if o.checkHealth(port) {
			services, _ := o.discoverServices(port)
			return services, nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return nil, fmt.Errorf("health check timeout after %s", o.healthTimeout)
}

// checkHealth performs a gRPC health check
func (o *Orchestrator) checkHealth(port int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return false
	}
	defer conn.Close()

	client := healthpb.NewHealthClient(conn)
	resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{})
	if err != nil {
		return false
	}

	return resp.Status == healthpb.HealthCheckResponse_SERVING
}

// discoverServices queries gRPC reflection to list available services
func (o *Orchestrator) discoverServices(port int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := rpb.NewServerReflectionClient(conn)
	stream, err := client.ServerReflectionInfo(ctx)
	if err != nil {
		return nil, err
	}

	if err := stream.Send(&rpb.ServerReflectionRequest{
		MessageRequest: &rpb.ServerReflectionRequest_ListServices{
			ListServices: "",
		},
	}); err != nil {
		return nil, err
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	listResp := resp.GetListServicesResponse()
	if listResp == nil {
		return nil, nil
	}

	var services []string
	for _, svc := range listResp.Service {
		// Filter out internal services
		if svc.Name == "grpc.reflection.v1alpha.ServerReflection" ||
			svc.Name == "grpc.reflection.v1.ServerReflection" ||
			svc.Name == "grpc.health.v1.Health" {
			continue
		}
		services = append(services, svc.Name)
	}

	return services, nil
}

// discoverMethods returns all methods for services on a port
func (o *Orchestrator) discoverMethods(port int) ([]ServiceMethod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := rpb.NewServerReflectionClient(conn)
	stream, err := client.ServerReflectionInfo(ctx)
	if err != nil {
		return nil, err
	}

	// List services
	if err := stream.Send(&rpb.ServerReflectionRequest{
		MessageRequest: &rpb.ServerReflectionRequest_ListServices{
			ListServices: "",
		},
	}); err != nil {
		return nil, err
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	listResp := resp.GetListServicesResponse()
	if listResp == nil {
		return nil, nil
	}

	var methods []ServiceMethod
	for _, svc := range listResp.Service {
		// Skip internal services
		if strings.HasPrefix(svc.Name, "grpc.") {
			continue
		}

		// Get file descriptor for this service
		if err := stream.Send(&rpb.ServerReflectionRequest{
			MessageRequest: &rpb.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: svc.Name,
			},
		}); err != nil {
			continue
		}

		fdResp, err := stream.Recv()
		if err != nil {
			continue
		}

		fdProto := fdResp.GetFileDescriptorResponse()
		if fdProto == nil {
			// Fallback: just list the service without methods
			methods = append(methods, ServiceMethod{
				Service: svc.Name,
				Method:  "(methods unavailable)",
			})
			continue
		}

		// Parse file descriptors for method names
		// For simplicity, just add the service entry
		methods = append(methods, ServiceMethod{
			Service: svc.Name,
			Method:  "*",
		})
	}

	return methods, nil
}

func killProcess(pid int) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	// Try graceful first
	if err := proc.Signal(os.Interrupt); err != nil {
		proc.Kill()
		return
	}
	// Wait briefly, then force kill
	done := make(chan struct{})
	go func() {
		proc.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		proc.Kill()
	}
}
