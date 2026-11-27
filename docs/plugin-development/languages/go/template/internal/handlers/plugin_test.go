package handlers

import (
	"context"
	"path/filepath"
	"testing"

	"go-plugin-template/internal/config"
	"go-plugin-template/internal/manifest"
	pb "go-plugin-template/proto"
)

func newHandler(t *testing.T) *PluginHandler {
	t.Helper()
	m, err := manifest.Load(filepath.Join("..", "..", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Plugin: config.PluginConfig{Name: "my-go-plugin", Version: "1.0.0"}, Timeout: 30}
	return NewPluginHandler(cfg, m)
}

func TestInitializeReturnsManifestInfo(t *testing.T) {
	resp, err := newHandler(t).Initialize(context.Background(), &pb.InitializeRequest{PluginName: "go-plugin-template"})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.GetSuccess() {
		t.Fatal("Initialize must succeed")
	}
	info := resp.GetPluginInfo()
	if info.GetName() != "go-plugin-template" {
		t.Errorf("name = %q, want the plugin.json name", info.GetName())
	}
	if len(info.GetCommands()) == 0 || info.GetCommands()[0].GetName() != "hello" {
		t.Errorf("commands = %v, want the plugin.json commands", info.GetCommands())
	}
}

func TestHealthIsHealthy(t *testing.T) {
	resp, err := newHandler(t).Health(context.Background(), &pb.HealthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.GetHealthy() {
		t.Errorf("healthy = false, status %q", resp.GetStatus())
	}
}

func TestExecute(t *testing.T) {
	h := newHandler(t)

	resp, err := h.Execute(context.Background(), &pb.ExecuteRequest{Command: "echo", Args: []string{"hi"}})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.GetSuccess() || resp.GetOutput() != "hi" || resp.GetExitCode() != 0 {
		t.Errorf("echo response = %v", resp)
	}

	// A failing command is reported in the response, not as a gRPC error
	resp, err = h.Execute(context.Background(), &pb.ExecuteRequest{Command: "unknown"})
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}
	if resp.GetSuccess() || resp.GetExitCode() == 0 || resp.GetError() == "" {
		t.Errorf("unknown command response = %v", resp)
	}
}

func TestListCommandsAndShutdown(t *testing.T) {
	h := newHandler(t)

	list, err := h.ListCommands(context.Background(), &pb.ListCommandsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.GetCommands()) == 0 {
		t.Error("ListCommands must return the plugin.json commands")
	}

	resp, err := h.Shutdown(context.Background(), &pb.ShutdownRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.GetSuccess() {
		t.Error("Shutdown must answer success")
	}
}
