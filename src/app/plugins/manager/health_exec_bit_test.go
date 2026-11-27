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

	"portunix.ai/portunix/src/pkg/platform"
)

func TestRequiresExecBit(t *testing.T) {
	tests := []struct {
		name   string
		plugin RegistryPlugin
		want   bool
	}{
		{name: "native (empty runtime)", plugin: RegistryPlugin{}, want: true},
		{name: "native", plugin: RegistryPlugin{Runtime: "native"}, want: true},
		{name: "python wheel", plugin: RegistryPlugin{Runtime: "python", Wheel: "demo-1.0.0-py3-none-any.whl"}, want: true},
		{name: "python script", plugin: RegistryPlugin{Runtime: "python"}, want: false},
		{name: "java", plugin: RegistryPlugin{Runtime: "java"}, want: false},
		{name: "node", plugin: RegistryPlugin{Runtime: "node"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requiresExecBit(&tt.plugin); got != tt.want {
				t.Errorf("requiresExecBit() = %v, want %v", got, tt.want)
			}
		})
	}
}

// writeFile0644 writes a non-executable file, creating parent directories
func writeFile0644(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0644); err != nil { // #nosec G306 -- test fixture
		t.Fatal(err)
	}
}

// stubInterpreters puts stub java, python3 and node executables alone in PATH
func stubInterpreters(t *testing.T) {
	t.Helper()
	binDir := t.TempDir()
	for _, name := range []string{"java", "python3", "node"} {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte("#!/bin/sh\necho 1.0.0\n"), 0755); err != nil { // #nosec G306 -- test executable
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir)
}

func TestCheckHelperPluginHealth_ExecBitPerRuntime(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("executable bit is POSIX-specific")
	}
	stubInterpreters(t)
	m := newTestManager(t)

	tests := []struct {
		name       string
		plugin     RegistryPlugin
		setup      func(t *testing.T, dir string)
		wantStatus string
	}{
		{
			name:       "native binary without exec bit",
			plugin:     RegistryPlugin{Name: "fx-native", BinaryName: "fx-native"},
			setup:      func(t *testing.T, dir string) { writeFile0644(t, filepath.Join(dir, "fx-native"), "#!/bin/sh\n") },
			wantStatus: "not_executable",
		},
		{
			name:   "python wheel entry point without exec bit",
			plugin: RegistryPlugin{Name: "fx-wheel", Runtime: "python", Wheel: "fx_wheel-1.0.0-py3-none-any.whl", BinaryName: "fx-wheel"},
			setup: func(t *testing.T, dir string) {
				venvDir, _ := makeLinkedVenv(t, dir)
				writeFile0644(t, filepath.Join(venvDir, "bin", "fx-wheel"), "#!/bin/sh\n")
			},
			wantStatus: "not_executable",
		},
		{
			name:       "python script without exec bit",
			plugin:     RegistryPlugin{Name: "fx-python", Runtime: "python", BinaryName: "main.py"},
			setup:      func(t *testing.T, dir string) { writeFile0644(t, filepath.Join(dir, "main.py"), "print('1.0.0')\n") },
			wantStatus: "ready",
		},
		{
			name:       "java jar without exec bit",
			plugin:     RegistryPlugin{Name: "fx-java", Runtime: "java", BinaryName: "fx-java.jar"},
			setup:      func(t *testing.T, dir string) { writeFile0644(t, filepath.Join(dir, "fx-java.jar"), "PK") },
			wantStatus: "ready",
		},
		{
			name:   "node script without exec bit",
			plugin: RegistryPlugin{Name: "fx-node", Runtime: "node", BinaryName: "main.js"},
			setup: func(t *testing.T, dir string) {
				writeFile0644(t, filepath.Join(dir, "main.js"), "console.log('1.0.0')\n")
			},
			wantStatus: "ready",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)
			plugin := tt.plugin
			plugin.Version = "1.0.0"
			plugin.Mode = "helper"
			plugin.InstallPath = dir

			health := m.checkHelperPluginHealth(&plugin)
			if health.Status != tt.wantStatus || health.Healthy != (tt.wantStatus == "ready") {
				t.Fatalf("health = %+v, want status %q", health, tt.wantStatus)
			}
		})
	}
}
