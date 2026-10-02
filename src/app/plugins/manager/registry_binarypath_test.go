/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"path/filepath"
	"testing"

	"portunix.ai/portunix/src/pkg/platform"
)

func TestRegistryPluginBinaryPath(t *testing.T) {
	installPath := filepath.Join("plugins", "demo")
	venvBin := "bin"
	exe := ""
	if platform.IsWindows() {
		venvBin = "Scripts"
		exe = ".exe"
	}

	tests := []struct {
		name   string
		plugin RegistryPlugin
		want   string
	}{
		{
			name:   "native without extension",
			plugin: RegistryPlugin{Runtime: "native", InstallPath: installPath, BinaryName: "ptx-demo"},
			want:   filepath.Join(installPath, "ptx-demo"+exe),
		},
		{
			name:   "empty runtime treated as native",
			plugin: RegistryPlugin{InstallPath: installPath, BinaryName: "ptx-demo"},
			want:   filepath.Join(installPath, "ptx-demo"+exe),
		},
		{
			name:   "native with explicit exe is not doubled",
			plugin: RegistryPlugin{Runtime: "native", InstallPath: installPath, BinaryName: "ptx-demo.exe"},
			want:   filepath.Join(installPath, "ptx-demo.exe"),
		},
		{
			name:   "java jar unchanged",
			plugin: RegistryPlugin{Runtime: "java", InstallPath: installPath, BinaryName: "demo.jar"},
			want:   filepath.Join(installPath, "demo.jar"),
		},
		{
			name:   "python script unchanged",
			plugin: RegistryPlugin{Runtime: "python", InstallPath: installPath, BinaryName: "main.py"},
			want:   filepath.Join(installPath, "main.py"),
		},
		{
			name:   "python wheel resolves to venv entry point",
			plugin: RegistryPlugin{Runtime: "python", Wheel: "demo-1.0.0-py3-none-any.whl", InstallPath: installPath, BinaryName: "ptx-demo"},
			want:   filepath.Join(installPath, ".venv", venvBin, "ptx-demo"+exe),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.plugin.BinaryPath(); got != tt.want {
				t.Errorf("BinaryPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
