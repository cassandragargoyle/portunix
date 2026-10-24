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
	"time"

	"portunix.ai/app/plugins"
	"portunix.ai/portunix/src/pkg/platform"
)

func TestMatureCPythonMinor(t *testing.T) {
	cases := []struct {
		now  string
		want string
	}{
		{"2026-10-24", "3.14"}, // 3.15 released this month, 3.14 a year ago
		{"2026-04-15", "3.14"}, // 3.14 (2025-10) six months old
		{"2026-03-15", "3.13"}, // 3.14 not yet six months old
		{"2025-04-01", "3.13"},
	}
	for _, c := range cases {
		now, err := time.Parse("2006-01-02", c.now)
		if err != nil {
			t.Fatal(err)
		}
		if got := matureCPythonMinor(now); got != c.want {
			t.Errorf("matureCPythonMinor(%s) = %s, want %s", c.now, got, c.want)
		}
	}
}

func TestUVPythonRequest(t *testing.T) {
	t.Setenv(pluginPythonVersionEnv, "3.13")

	manifest := testWheelManifest("demo")
	if got, defaulted := uvPythonRequest(manifest, "3.12"); got != "3.12" || defaulted {
		t.Errorf("pinned: got %q defaulted=%v, want 3.12 false", got, defaulted)
	}
	if got, defaulted := uvPythonRequest(manifest, ""); got != "3.13" || !defaulted {
		t.Errorf("unpinned: got %q defaulted=%v, want 3.13 true", got, defaulted)
	}

	manifest.Plugin.PythonMinVersion = "3.10"
	if got, defaulted := uvPythonRequest(manifest, ""); got != "3.13" || !defaulted {
		t.Errorf("min below default: got %q defaulted=%v, want 3.13 true", got, defaulted)
	}

	manifest.Plugin.PythonMinVersion = "3.14"
	if got, defaulted := uvPythonRequest(manifest, ""); got != ">=3.14" || defaulted {
		t.Errorf("min above default: got %q defaulted=%v, want >=3.14 false", got, defaulted)
	}
}

func TestDefaultPluginPythonVersionWithoutOverride(t *testing.T) {
	t.Setenv(pluginPythonVersionEnv, "")
	if got, want := defaultPluginPythonVersion(), matureCPythonMinor(time.Now()); got != want {
		t.Errorf("defaultPluginPythonVersion() = %s, want %s", got, want)
	}
}

// writeExecutable writes a shell script into dir
func writeExecutable(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0755); err != nil { // #nosec G306 -- test executable
		t.Fatal(err)
	}
	return path
}

func TestApplyInstalledRuntimeCheckUsesVenv(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("shell-script interpreter is POSIX-specific")
	}
	// No system python on PATH
	t.Setenv("PATH", t.TempDir())

	installPath := t.TempDir()
	binDir := filepath.Join(installPath, ".venv", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, binDir, "python", `echo "Python 3.14.1"`)

	rp := &RegistryPlugin{Name: "demo", Runtime: "python", Wheel: "demo-1.0.0-py3-none-any.whl", InstallPath: installPath}

	result := plugins.CheckPrerequisitesFromRegistry("demo", "1.0.0", "python", "", "", nil, nil)
	if result.Runtime.Status != plugins.CheckStatusError {
		t.Fatalf("precondition: host runtime check should fail without python3, got %+v", result.Runtime)
	}

	applyInstalledRuntimeCheck(result, rp)
	result.SetInstallCheck(installCheckResult(rp))
	if result.HasErrors() {
		t.Fatalf("healthy venv plugin reported errors: %+v / %+v", result.Runtime, result.Installation)
	}
	if result.Runtime.Found != "3.14.1" {
		t.Errorf("Found = %q, want venv version 3.14.1", result.Runtime.Found)
	}
}

func TestManifestPrerequisitesWithUV(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("shell-script uv is POSIX-specific")
	}
	binDir := t.TempDir()
	interpreter := writeExecutable(t, t.TempDir(), "python3.13", `echo "Python 3.13.5"`)
	writeExecutable(t, binDir, "uv", "echo "+interpreter)
	t.Setenv("PATH", binDir) // uv only, no system python
	t.Setenv(pluginPythonVersionEnv, "3.13")

	manifest := testWheelManifest("demo")
	result := manifestPrerequisites(manifest, t.TempDir())
	if result.Runtime.Status != plugins.CheckStatusOK || result.Runtime.Found != "3.13.5" {
		t.Fatalf("runtime = %+v, want ok from uv-managed 3.13.5", result.Runtime)
	}
	if result.HasErrors() {
		t.Fatalf("unexpected errors: %+v", result)
	}

	// Native plugin keeps the generic runtime check
	manifest.Plugin.Runtime = ""
	manifest.Plugin.Wheel = ""
	if result := manifestPrerequisites(manifest, t.TempDir()); result.Runtime.Required != "native (self-contained)" {
		t.Errorf("native runtime = %+v", result.Runtime)
	}
}
