/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestPluginList_PlatformFlags exercises the CLI half of issue #175:
// `ptx plugin list --platform=<name>` must accept the new flags without
// regressing the default (no-platform) behaviour, and --platform-version /
// --feature must only be accepted in conjunction with --platform.
func TestPluginList_PlatformFlags(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("CLI integration test runs on Linux/macOS only")
	}

	root := findProjectRoot()
	bin := filepath.Join(root, "portunix")
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		t.Skipf("portunix binary not built (run `make build`): %v", err)
	}

	// Isolate the test from the developer's registry by pointing $HOME at a
	// temp dir. The plugin manager resolves its registry under
	// $HOME/.portunix/plugins.
	tmpHome := t.TempDir()

	run := func(args ...string) (string, error) {
		cmd := exec.Command(bin, args...)
		cmd.Env = append(os.Environ(), "HOME="+tmpHome)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	t.Run("PlatformQuery_EmptyRegistryReturnsEmptyJSON", func(t *testing.T) {
		out, err := run("plugin", "list", "--platform=synapse", "-o", "json")
		if err != nil {
			t.Fatalf("CLI failed: %v\nOutput:\n%s", err, out)
		}
		trimmed := strings.TrimSpace(out)
		if trimmed != "[]" {
			t.Errorf("expected empty array, got: %q", trimmed)
		}
		var decoded []interface{}
		if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
			t.Errorf("output is not valid JSON: %v", err)
		}
	})

	t.Run("PlatformVersionRequiresPlatform", func(t *testing.T) {
		out, err := run("plugin", "list", "--platform-version=0.3.0")
		if err == nil {
			t.Fatalf("expected CLI to reject --platform-version without --platform, stdout:\n%s", out)
		}
		if !strings.Contains(out, "require --platform") {
			t.Errorf("expected guidance about --platform, got: %s", out)
		}
	})

	t.Run("FeatureRequiresPlatform", func(t *testing.T) {
		out, err := run("plugin", "list", "--feature=connector.command")
		if err == nil {
			t.Fatalf("expected CLI to reject --feature without --platform, stdout:\n%s", out)
		}
		if !strings.Contains(out, "require --platform") {
			t.Errorf("expected guidance about --platform, got: %s", out)
		}
	})

	t.Run("DefaultListStillWorks_BackwardCompat", func(t *testing.T) {
		// No --platform — plain list against empty registry should succeed.
		out, err := run("plugin", "list")
		if err != nil {
			t.Fatalf("plain plugin list regressed: %v\n%s", err, out)
		}
		if !strings.Contains(out, "No plugins installed") {
			t.Errorf("expected friendly empty-state message, got: %s", out)
		}
	})
}
