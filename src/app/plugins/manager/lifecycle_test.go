/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"google.golang.org/grpc"
	"portunix.ai/app/plugins"
	pb "portunix.ai/app/plugins/proto"
	"portunix.ai/portunix/src/pkg/platform"
)

// testGRPCPluginEnv makes the test binary act as a gRPC plugin (helper process
// pattern): "1" serves the plugin protocol, "crash" exits at once with an error
const testGRPCPluginEnv = "PORTUNIX_TEST_GRPC_PLUGIN"

func TestMain(m *testing.M) {
	switch os.Getenv(testGRPCPluginEnv) {
	case "1":
		os.Exit(runTestGRPCPlugin(os.Args[1:]))
	case "crash":
		fmt.Fprintln(os.Stderr, "simulated crash on startup")
		os.Exit(3)
	}
	os.Exit(m.Run())
}

// testPluginServer implements the host plugin protocol for lifecycle tests
type testPluginServer struct {
	pb.UnimplementedPluginServiceServer
	server  *grpc.Server
	started time.Time
}

func (s *testPluginServer) Initialize(_ context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	return &pb.InitializeResponse{
		Success:    true,
		PluginInfo: &pb.PluginInfo{Name: req.PluginName, Version: req.Version},
	}, nil
}

func (s *testPluginServer) Execute(_ context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	return &pb.ExecuteResponse{Success: true, Output: "echo: " + req.Command}, nil
}

func (s *testPluginServer) Health(context.Context, *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Healthy:       true,
		Status:        "running",
		Message:       "test plugin healthy",
		UptimeSeconds: int64(time.Since(s.started).Seconds()),
	}, nil
}

func (s *testPluginServer) Shutdown(context.Context, *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	go s.server.GracefulStop()
	return &pb.ShutdownResponse{Success: true}, nil
}

// runTestGRPCPlugin serves the plugin protocol on the port given by --port
func runTestGRPCPlugin(args []string) int {
	port := ""
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--port" {
			port = args[i+1]
		}
	}
	listener, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v\n", err)
		return 1
	}

	server := grpc.NewServer()
	pb.RegisterPluginServiceServer(server, &testPluginServer{server: server, started: time.Now()})

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		server.GracefulStop()
	}()

	if err := server.Serve(listener); err != nil {
		fmt.Fprintf(os.Stderr, "serve: %v\n", err)
		return 1
	}
	return 0
}

// freePort returns a TCP port that is free at the time of the call
func freePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

// cliManager opens a manager over an existing plugins directory, like one CLI command does
func cliManager(t *testing.T, pluginsDir string, port int) *Manager {
	t.Helper()
	m, err := NewManager(ManagerConfig{
		PluginsDir:   pluginsDir,
		RegistryFile: filepath.Join(pluginsDir, "registry.json"),
		DefaultPort:  port,
		PortRange:    PortRange{Start: port, End: port + 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// runCLI runs fn on a fresh manager and shuts it down afterwards
func runCLI(t *testing.T, pluginsDir string, port int, fn func(m *Manager)) {
	t.Helper()
	m := cliManager(t, pluginsDir, port)
	defer func() { _ = m.Shutdown() }()
	fn(m)
}

// registerGRPCPlugin registers a gRPC plugin whose binary is a copy of the test binary
func registerGRPCPlugin(t *testing.T, pluginsDir, name string) {
	t.Helper()
	installDir := filepath.Join(pluginsDir, name)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		t.Fatal(err)
	}
	binary := platform.ExecutableName(name)
	if err := copyFile(os.Args[0], filepath.Join(installDir, binary), 0755); err != nil {
		t.Fatal(err)
	}

	registry, err := NewRegistry(filepath.Join(pluginsDir, "registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := &plugins.PluginManifest{
		Name:    name,
		Version: "1.0.0",
		Plugin:  plugins.PluginBinaryConfig{Type: "grpc", Binary: binary},
	}
	if err := registry.RegisterPlugin(manifest, installDir); err != nil {
		t.Fatal(err)
	}
}

// registryEntry reads the registry file as a new CLI process would
func registryEntry(t *testing.T, pluginsDir, name string) *RegistryPlugin {
	t.Helper()
	registry, err := NewRegistry(filepath.Join(pluginsDir, "registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	rp, err := registry.GetPluginRegistryData(name)
	if err != nil {
		t.Fatal(err)
	}
	return rp
}

func TestEnabledPluginReloadedFromRegistry(t *testing.T) {
	pluginsDir := t.TempDir()
	port := freePort(t)
	registerGRPCPlugin(t, pluginsDir, "reload")

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.EnablePlugin("reload"); err != nil {
			t.Fatalf("enable: %v", err)
		}
	})

	if rp := registryEntry(t, pluginsDir, "reload"); !rp.Enabled || rp.ActivePort != port {
		t.Fatalf("registry after enable: enabled=%v active_port=%d, want true/%d", rp.Enabled, rp.ActivePort, port)
	}

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.EnablePlugin("reload"); err == nil || !strings.Contains(err.Error(), "already enabled") {
			t.Errorf("second enable: got %v, want already enabled", err)
		}

		health, err := m.GetPluginHealth("reload")
		if err != nil {
			t.Fatalf("health of enabled plugin: %v", err)
		}
		if health.Healthy || health.Status != "not_running" {
			t.Errorf("health = %+v, want not_running", health)
		}

		if err := m.StopPlugin("reload"); err == nil || !strings.Contains(err.Error(), "is not running") {
			t.Errorf("stop of stopped plugin: got %v, want not running", err)
		}

		if err := m.DisablePlugin("reload"); err != nil {
			t.Fatalf("disable in new process: %v", err)
		}
	})

	if rp := registryEntry(t, pluginsDir, "reload"); rp.Enabled || rp.ActivePort != 0 {
		t.Fatalf("registry after disable: enabled=%v active_port=%d, want false/0", rp.Enabled, rp.ActivePort)
	}

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.StartPlugin("reload"); err == nil || !strings.Contains(err.Error(), "is not enabled") {
			t.Errorf("start of disabled plugin: got %v, want not enabled", err)
		}
	})
}

func TestStalePluginProcessCleared(t *testing.T) {
	pluginsDir := t.TempDir()
	port := freePort(t)
	registerGRPCPlugin(t, pluginsDir, "stale")

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.EnablePlugin("stale"); err != nil {
			t.Fatalf("enable: %v", err)
		}
	})

	// PID of a process that has already exited
	exited := exec.Command(os.Args[0], "-test.run=^$")
	if err := exited.Run(); err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistry(filepath.Join(pluginsDir, "registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.SetPluginProcess("stale", port, exited.Process.Pid); err != nil {
		t.Fatal(err)
	}
	if err := registry.UpdatePluginStatus("stale", plugins.PluginStatusRunning); err != nil {
		t.Fatal(err)
	}

	runCLI(t, pluginsDir, port, func(m *Manager) {
		// list and info read the registry only, the dead process must still show
		info, err := m.GetPlugin("stale")
		if err != nil {
			t.Fatal(err)
		}
		if info.Status != plugins.PluginStatusFailed {
			t.Errorf("info status of dead process = %s, want failed", info.Status)
		}

		health, err := m.GetPluginHealth("stale")
		if err != nil {
			t.Fatalf("health: %v", err)
		}
		if health.Status != "not_running" {
			t.Errorf("health status = %q, want not_running", health.Status)
		}
	})

	rp := registryEntry(t, pluginsDir, "stale")
	if rp.PID != 0 || rp.Status != plugins.PluginStatusFailed {
		t.Errorf("registry after stale PID: pid=%d status=%s, want 0/failed", rp.PID, rp.Status)
	}
}

func TestGRPCPluginLifecycleAcrossManagers(t *testing.T) {
	if testing.Short() {
		t.Skip("starts a plugin process")
	}

	pluginsDir := t.TempDir()
	port := freePort(t)
	registerGRPCPlugin(t, pluginsDir, "lifecycle")
	t.Setenv(testGRPCPluginEnv, "1")

	// Never leave a detached plugin process behind
	t.Cleanup(func() {
		registry, err := NewRegistry(filepath.Join(pluginsDir, "registry.json"))
		if err != nil {
			return
		}
		if rp, err := registry.GetPluginRegistryData("lifecycle"); err == nil && rp.PID > 0 {
			if proc, err := os.FindProcess(rp.PID); err == nil {
				_ = proc.Kill()
			}
		}
	})

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.EnablePlugin("lifecycle"); err != nil {
			t.Fatalf("enable: %v", err)
		}
	})

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.StartPlugin("lifecycle"); err != nil {
			t.Fatalf("start in new process: %v", err)
		}
	})

	if rp := registryEntry(t, pluginsDir, "lifecycle"); rp.PID == 0 || rp.Status != plugins.PluginStatusRunning {
		t.Fatalf("registry after start: pid=%d status=%s, want pid and running", rp.PID, rp.Status)
	}

	runCLI(t, pluginsDir, port, func(m *Manager) {
		health, err := m.GetPluginHealth("lifecycle")
		if err != nil {
			t.Fatalf("health: %v", err)
		}
		if !health.Healthy || health.Status != "running" {
			t.Fatalf("health = %+v, want healthy gRPC answer (plugin must survive manager shutdown)", health)
		}

		resp, err := m.ExecuteCommand("lifecycle", "ping", nil, nil)
		if err != nil {
			t.Fatalf("execute: %v", err)
		}
		if resp.Output != "echo: ping" {
			t.Errorf("execute output = %q, want %q", resp.Output, "echo: ping")
		}

		if err := m.StartPlugin("lifecycle"); err == nil || !strings.Contains(err.Error(), "already running") {
			t.Errorf("second start: got %v, want already running", err)
		}
	})

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.StopPlugin("lifecycle"); err != nil {
			t.Fatalf("stop in new process: %v", err)
		}
	})

	if rp := registryEntry(t, pluginsDir, "lifecycle"); rp.PID != 0 || rp.Status != plugins.PluginStatusStopped {
		t.Errorf("registry after stop: pid=%d status=%s, want 0/stopped", rp.PID, rp.Status)
	}
	if conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second); err == nil {
		conn.Close()
		t.Errorf("plugin still listens on port %d after stop", port)
	}

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.DisablePlugin("lifecycle"); err != nil {
			t.Fatalf("disable: %v", err)
		}
		if err := m.UninstallPlugin("lifecycle"); err != nil {
			t.Fatalf("uninstall: %v", err)
		}
	})
}

// killRecordedProcess kills the plugin process recorded in the registry, if any
func killRecordedProcess(pluginsDir, name string) {
	registry, err := NewRegistry(filepath.Join(pluginsDir, "registry.json"))
	if err != nil {
		return
	}
	if rp, err := registry.GetPluginRegistryData(name); err == nil && rp.PID > 0 {
		if proc, err := os.FindProcess(rp.PID); err == nil {
			_ = proc.Kill()
		}
	}
}

func TestConcurrentStartLeavesOneStoppableProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("starts plugin processes")
	}

	pluginsDir := t.TempDir()
	port := freePort(t)
	registerGRPCPlugin(t, pluginsDir, "concurrent")
	t.Setenv(testGRPCPluginEnv, "1")
	t.Cleanup(func() { killRecordedProcess(pluginsDir, "concurrent") })

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.EnablePlugin("concurrent"); err != nil {
			t.Fatalf("enable: %v", err)
		}
	})

	// Two CLI processes start the same plugin at the same time; the goroutines
	// report every error on the channel, t.Fatal there would hang the test
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			m, err := NewManager(ManagerConfig{
				PluginsDir:   pluginsDir,
				RegistryFile: filepath.Join(pluginsDir, "registry.json"),
				DefaultPort:  port,
				PortRange:    PortRange{Start: port, End: port + 10},
			})
			if err != nil {
				errs <- fmt.Errorf("new manager: %w", err)
				return
			}
			defer func() { _ = m.Shutdown() }()
			errs <- m.StartPlugin("concurrent")
		}()
	}
	var started int
	for i := 0; i < 2; i++ {
		var err error
		select {
		case err = <-errs:
		case <-time.After(60 * time.Second):
			t.Fatal("concurrent start did not finish within 60 s")
		}
		switch {
		case err == nil:
			started++
		case !strings.Contains(err.Error(), "already running"):
			t.Errorf("losing start: got %v, want already running", err)
		}
	}
	if started != 1 {
		t.Fatalf("successful starts = %d, want 1", started)
	}

	rp := registryEntry(t, pluginsDir, "concurrent")
	if !plugins.ProcessAlive(rp.PID) {
		t.Fatalf("registry pid %d is not a live process", rp.PID)
	}

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.StopPlugin("concurrent"); err != nil {
			t.Fatalf("stop: %v", err)
		}
	})
	if conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second); err == nil {
		conn.Close()
		t.Errorf("a plugin process still listens on port %d after stop", port)
	}
}

func TestStartFailsFastWhenPluginExits(t *testing.T) {
	pluginsDir := t.TempDir()
	port := freePort(t)
	registerGRPCPlugin(t, pluginsDir, "crashing")
	t.Setenv(testGRPCPluginEnv, "crash")

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.EnablePlugin("crashing"); err != nil {
			t.Fatalf("enable: %v", err)
		}

		begin := time.Now()
		err := m.StartPlugin("crashing")
		elapsed := time.Since(begin)

		if err == nil {
			t.Fatal("start of a crashing plugin succeeded")
		}
		for _, want := range []string{"exited during startup", "exit status 3", "simulated crash on startup"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("start error %q does not contain %q", err, want)
			}
		}
		if elapsed > 5*time.Second {
			t.Errorf("start took %s, want the exit to be detected before the connect timeout", elapsed)
		}
	})

	rp := registryEntry(t, pluginsDir, "crashing")
	if rp.PID != 0 || rp.Status != plugins.PluginStatusFailed {
		t.Errorf("registry after failed start: pid=%d status=%s, want 0/failed", rp.PID, rp.Status)
	}
}

func TestStartRefusesOccupiedPort(t *testing.T) {
	pluginsDir := t.TempDir()
	port := freePort(t)
	registerGRPCPlugin(t, pluginsDir, "occupied")
	t.Setenv(testGRPCPluginEnv, "1")
	t.Cleanup(func() { killRecordedProcess(pluginsDir, "occupied") })

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.EnablePlugin("occupied"); err != nil {
			t.Fatalf("enable: %v", err)
		}
	})

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	runCLI(t, pluginsDir, port, func(m *Manager) {
		err := m.StartPlugin("occupied")
		if err == nil || !strings.Contains(err.Error(), "already in use") {
			t.Fatalf("start on an occupied port: got %v, want already in use", err)
		}
	})

	if rp := registryEntry(t, pluginsDir, "occupied"); rp.PID != 0 {
		t.Errorf("registry pid = %d after refused start, want 0", rp.PID)
	}
}

func TestEnableAssignsFreePort(t *testing.T) {
	pluginsDir := t.TempDir()
	port := freePort(t)
	registerGRPCPlugin(t, pluginsDir, "freeport")

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.EnablePlugin("freeport"); err != nil {
			t.Fatalf("enable: %v", err)
		}
	})

	if rp := registryEntry(t, pluginsDir, "freeport"); rp.ActivePort == port || rp.ActivePort == 0 {
		t.Errorf("active port = %d, want a free port other than the occupied %d", rp.ActivePort, port)
	}
}

func TestHelperPluginHasNoService(t *testing.T) {
	pluginsDir := t.TempDir()
	port := freePort(t)

	registry, err := NewRegistry(filepath.Join(pluginsDir, "registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := &plugins.PluginManifest{
		Name:    "helper",
		Version: "1.0.0",
		Plugin:  plugins.PluginBinaryConfig{Type: "helper", Binary: "ptx-helper"},
	}
	if err := registry.RegisterPlugin(manifest, filepath.Join(pluginsDir, "helper")); err != nil {
		t.Fatal(err)
	}

	runCLI(t, pluginsDir, port, func(m *Manager) {
		if err := m.EnablePlugin("helper"); err != nil {
			t.Fatalf("enable: %v", err)
		}
	})

	if rp := registryEntry(t, pluginsDir, "helper"); !rp.Enabled || rp.ActivePort != 0 || rp.Status != plugins.PluginStatusReady {
		t.Errorf("helper after enable: enabled=%v active_port=%d status=%s, want true/0/ready", rp.Enabled, rp.ActivePort, rp.Status)
	}

	runCLI(t, pluginsDir, port, func(m *Manager) {
		begin := time.Now()
		err := m.StartPlugin("helper")
		if err == nil || !strings.Contains(err.Error(), "helper plugin") {
			t.Errorf("start of a helper plugin: got %v, want helper plugin refusal", err)
		}
		if time.Since(begin) > time.Second {
			t.Errorf("refusal took %s, want it at once", time.Since(begin))
		}

		if err := m.DisablePlugin("helper"); err != nil {
			t.Errorf("disable of an enabled helper plugin: %v", err)
		}
	})
}

func TestRegistryReadableDuringConcurrentWrites(t *testing.T) {
	pluginsDir := t.TempDir()
	registerGRPCPlugin(t, pluginsDir, "writes")
	registryFile := filepath.Join(pluginsDir, "registry.json")

	writer, err := NewRegistry(registryFile)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		for i := 0; i < 300; i++ {
			status := plugins.PluginStatusStopped
			if i%2 == 0 {
				status = plugins.PluginStatusRunning
			}
			if err := writer.UpdatePluginStatus("writes", status); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()

	// Every reader is a new CLI process loading the registry while it is saved
	for reads := 0; ; reads++ {
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("writer: %v", err)
			}
			if reads == 0 {
				t.Fatal("no read overlapped the writes")
			}
			return
		default:
		}
		if _, err := NewRegistry(registryFile); err != nil {
			t.Fatalf("read %d during writes: %v", reads, err)
		}
	}
}

func TestTryLockFileReportsHeldLock(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "registry.json.lock")
	open := func() *os.File {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })
		return f
	}
	holder, waiter := open(), open()

	if locked, err := tryLockFile(holder); err != nil || !locked {
		t.Fatalf("first lock: locked=%v err=%v, want true", locked, err)
	}
	if locked, err := tryLockFile(waiter); err != nil || locked {
		t.Fatalf("second lock while held: locked=%v err=%v, want false", locked, err)
	}
	if err := unlockFile(holder); err != nil {
		t.Fatal(err)
	}
	if locked, err := tryLockFile(waiter); err != nil || !locked {
		t.Fatalf("lock after release: locked=%v err=%v, want true", locked, err)
	}
	_ = unlockFile(waiter)
}

// symlinkOrSkip creates a symlink or skips where the OS does not allow it
func symlinkOrSkip(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks not available: %v", err)
	}
}

func TestRegistrySaveKeepsSymlink(t *testing.T) {
	pluginsDir := t.TempDir()
	dotfiles := t.TempDir()
	target := filepath.Join(dotfiles, "registry.json")
	link := filepath.Join(pluginsDir, "registry.json")

	if _, err := NewRegistry(target); err != nil {
		t.Fatal(err)
	}
	symlinkOrSkip(t, target, link)

	registry, err := NewRegistry(link)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterPlugin(&plugins.PluginManifest{
		Name: "linked", Version: "1.0.0", Plugin: plugins.PluginBinaryConfig{Binary: "linked"},
	}, filepath.Join(pluginsDir, "linked")); err != nil {
		t.Fatal(err)
	}

	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("registry.json symlink was replaced by a regular file")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"linked"`) {
		t.Error("symlink target was not updated")
	}
	if matches, _ := filepath.Glob(filepath.Join(pluginsDir, ".registry-*.tmp")); len(matches) > 0 {
		t.Errorf("temporary files left next to the symlink: %v", matches)
	}
}

func TestRegistrySaveCreatesDanglingSymlinkTarget(t *testing.T) {
	pluginsDir := t.TempDir()
	target := filepath.Join(t.TempDir(), "registry.json")
	link := filepath.Join(pluginsDir, "registry.json")
	symlinkOrSkip(t, target, link)

	if _, err := NewRegistry(link); err != nil {
		t.Fatal(err)
	}

	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("dangling registry symlink was replaced (err %v)", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Errorf("symlink target was not created: %v", err)
	}
}

func TestRegistryFileIsPrivate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes")
	}
	registryFile := filepath.Join(t.TempDir(), "registry.json")
	if _, err := NewRegistry(registryFile); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(registryFile)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0600 {
		t.Errorf("registry mode = %o, want 600", mode)
	}
}
