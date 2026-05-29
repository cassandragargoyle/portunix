/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package integration

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "portunix.ai/app/plugins/proto/pluginregistry"
)

// TestPtxPluginRegistry_Lifecycle is an end-to-end smoke test for the
// ptx-plugin-registry helper introduced in issue #175. It starts the helper
// over a unix socket, waits for the socket to appear, dials gRPC, and verifies
// HealthCheck + ListPluginsForPlatform against a fresh (empty) registry.
//
// Skipped on Windows — the helper listens on TCP loopback there, which is
// still covered by the unit tests in src/app/plugins/server/. This
// integration test focuses on the unix-socket path because that is the
// default transport and the one with the most environment-specific
// behaviour.
func TestPtxPluginRegistry_Lifecycle(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket transport not tested on Windows")
	}

	root := findProjectRoot()
	helperBin := filepath.Join(root, "ptx-plugin-registry")
	if _, err := os.Stat(helperBin); os.IsNotExist(err) {
		t.Skipf("ptx-plugin-registry binary not built (run `make build`): %v", err)
	}

	// Point the daemon at an empty plugin-dir so the test is hermetic — no
	// interference with the developer's ~/.portunix/plugins registry.
	tmpPluginDir := t.TempDir()
	socketPath := filepath.Join(t.TempDir(), "plugin-registry.sock")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, helperBin,
		"plugin-registry", "serve",
		"--mode", "unix",
		"--socket", socketPath,
		"--plugin-dir", tmpPluginDir,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper: %v", err)
	}
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Signal(os.Interrupt)
			_, _ = cmd.Process.Wait()
		}
	})

	// Wait for the socket to appear and be connectable. The helper reports a
	// startup line to stderr, but polling the socket directly is more robust.
	if err := waitForSocket(socketPath, 5*time.Second); err != nil {
		t.Fatalf("socket never became ready: %v", err)
	}

	conn, err := grpc.NewClient("unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc NewClient: %v", err)
	}
	defer conn.Close()

	client := pb.NewPluginRegistryServiceClient(conn)

	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := client.HealthCheck(ctx, &pb.HealthCheckRequest{})
		if err != nil {
			t.Fatalf("HealthCheck: %v", err)
		}
		if !resp.Healthy {
			t.Errorf("expected healthy, got %+v", resp)
		}
		if resp.Version == "" {
			t.Error("expected non-empty version in response")
		}
	})

	t.Run("ListPluginsForPlatform_EmptyRegistry", func(t *testing.T) {
		resp, err := client.ListPluginsForPlatform(ctx, &pb.ListPluginsForPlatformRequest{
			Platform: "synapse",
		})
		if err != nil {
			t.Fatalf("ListPluginsForPlatform: %v", err)
		}
		if len(resp.Plugins) != 0 {
			t.Errorf("expected 0 plugins in empty registry, got %d", len(resp.Plugins))
		}
	})

	t.Run("ListPluginsForPlatform_EmptyPlatformRejected", func(t *testing.T) {
		if _, err := client.ListPluginsForPlatform(ctx, &pb.ListPluginsForPlatformRequest{}); err == nil {
			t.Error("expected error when platform is empty")
		}
	})
}

// waitForSocket polls a unix socket path until a connection succeeds or the
// timeout expires. Useful because the daemon reports readiness on stderr but
// we cannot reliably parse that without introducing brittle assumptions
// about log output.
func waitForSocket(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			conn, err := net.Dial("unix", path)
			if err == nil {
				conn.Close()
				return nil
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for socket %s", path)
}
