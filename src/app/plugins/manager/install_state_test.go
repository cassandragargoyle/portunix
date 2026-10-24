/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"os"
	"path/filepath"
	"testing"

	"portunix.ai/app/plugins"
	"portunix.ai/portunix/src/pkg/platform"
)

// makeLinkedVenv builds a venv layout whose python symlinks to a temporary
// interpreter copy and returns the venv dir and the interpreter path
func makeLinkedVenv(t *testing.T, installPath string) (string, string) {
	t.Helper()
	baseDir := t.TempDir()
	interpreter := filepath.Join(baseDir, "python3.99")
	if err := os.WriteFile(interpreter, []byte("#!/bin/sh\n"), 0755); err != nil { // #nosec G306 -- test interpreter stub
		t.Fatal(err)
	}

	venvDir := filepath.Join(installPath, ".venv")
	binDir := filepath.Join(venvDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(interpreter, filepath.Join(binDir, "python3.99")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("python3.99", filepath.Join(binDir, "python")); err != nil {
		t.Fatal(err)
	}
	cfg := "home = " + baseDir + "\nversion = 3.99.0\n"
	if err := os.WriteFile(filepath.Join(venvDir, "pyvenv.cfg"), []byte(cfg), 0644); err != nil { // #nosec G306 -- test fixture
		t.Fatal(err)
	}
	return venvDir, interpreter
}

func TestCheckInstall(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("venv symlink layout is POSIX-specific")
	}

	t.Run("healthy venv", func(t *testing.T) {
		installPath := t.TempDir()
		makeLinkedVenv(t, installPath)
		rp := &RegistryPlugin{Name: "demo", Runtime: "python", Wheel: "demo.whl", InstallPath: installPath}
		if state, reason := rp.CheckInstall(); state != InstallStateOK {
			t.Fatalf("state = %s (%s), want ok", state, reason)
		}
	})

	t.Run("removed base interpreter", func(t *testing.T) {
		installPath := t.TempDir()
		_, interpreter := makeLinkedVenv(t, installPath)
		if err := os.Remove(interpreter); err != nil {
			t.Fatal(err)
		}
		rp := &RegistryPlugin{Name: "demo", Runtime: "python", Wheel: "demo.whl", InstallPath: installPath}
		state, reason := rp.CheckInstall()
		if state != InstallStateBroken {
			t.Fatalf("state = %s, want broken", state)
		}
		if reason == "" {
			t.Error("expected a reason for the broken state")
		}
		if hint := rp.RepairHint(state); hint != "portunix plugin repair demo" {
			t.Errorf("hint = %q", hint)
		}
	})

	t.Run("venv missing", func(t *testing.T) {
		rp := &RegistryPlugin{Name: "demo", Runtime: "python", Wheel: "demo.whl", InstallPath: t.TempDir()}
		if state, _ := rp.CheckInstall(); state != InstallStateBroken {
			t.Fatalf("state = %s, want broken", state)
		}
	})

	t.Run("install directory missing", func(t *testing.T) {
		rp := &RegistryPlugin{Name: "demo", Runtime: "native", InstallPath: filepath.Join(t.TempDir(), "gone")}
		if state, _ := rp.CheckInstall(); state != InstallStateMissing {
			t.Fatalf("state = %s, want missing", state)
		}
	})

	t.Run("native plugin ignores venv", func(t *testing.T) {
		rp := &RegistryPlugin{Name: "demo", Runtime: "native", InstallPath: t.TempDir()}
		if state, _ := rp.CheckInstall(); state != InstallStateOK {
			t.Fatalf("state = %s, want ok", state)
		}
	})
}

func TestInstallStatus(t *testing.T) {
	tests := []struct {
		stored plugins.PluginStatus
		state  InstallState
		want   plugins.PluginStatus
	}{
		{plugins.PluginStatusReady, InstallStateOK, plugins.PluginStatusReady},
		{plugins.PluginStatusReady, InstallStateBroken, plugins.PluginStatusBroken},
		{plugins.PluginStatusReady, InstallStateMissing, plugins.PluginStatusFailed},
		{plugins.PluginStatusRunning, InstallStateOK, plugins.PluginStatusRunning},
	}
	for _, tt := range tests {
		if got := installStatus(tt.stored, tt.state); got != tt.want {
			t.Errorf("installStatus(%s, %s) = %s, want %s", tt.stored, tt.state, got, tt.want)
		}
	}
}

func TestListPluginsReportsBrokenVenv(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("venv symlink layout is POSIX-specific")
	}

	m := newTestManager(t)
	pluginDir := filepath.Join(m.config.PluginsDir, "demo")
	_, interpreter := makeLinkedVenv(t, pluginDir)
	if err := m.registry.RegisterPlugin(testWheelManifest("demo"), pluginDir); err != nil {
		t.Fatal(err)
	}

	assertStatus := func(want plugins.PluginStatus) {
		t.Helper()
		list, err := m.ListPlugins()
		if err != nil {
			t.Fatal(err)
		}
		if len(list) != 1 || list[0].Status != want {
			t.Fatalf("ListPlugins status = %v, want %s", list, want)
		}
		info, err := m.GetPlugin("demo")
		if err != nil {
			t.Fatal(err)
		}
		if info.Status != want {
			t.Fatalf("GetPlugin status = %s, want %s", info.Status, want)
		}
	}

	assertStatus(plugins.PluginStatusReady)

	if err := os.Remove(interpreter); err != nil {
		t.Fatal(err)
	}
	assertStatus(plugins.PluginStatusBroken)

	health, err := m.GetPluginHealth("demo")
	if err != nil {
		t.Fatal(err)
	}
	if health.Healthy {
		t.Errorf("health reported healthy for broken venv: %+v", health)
	}

	result, err := m.CheckPluginPrerequisites("demo")
	if err != nil {
		t.Fatal(err)
	}
	if result.Installation == nil || result.Installation.Status != plugins.CheckStatusError || !result.HasErrors() {
		t.Errorf("plugin check did not report broken install: %+v", result.Installation)
	}

	if err := os.RemoveAll(pluginDir); err != nil {
		t.Fatal(err)
	}
	assertStatus(plugins.PluginStatusFailed)
}
