/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"portunix.ai/app/plugins"
	"portunix.ai/portunix/src/pkg/platform"
)

// newTestManager creates a manager over a temporary plugins directory
func newTestManager(t *testing.T) *Manager {
	t.Helper()
	pluginsDir := t.TempDir()
	m, err := NewManager(ManagerConfig{
		PluginsDir:   pluginsDir,
		RegistryFile: filepath.Join(pluginsDir, "registry.json"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Shutdown() })
	return m
}

// testWheelManifest returns a valid manifest of a Python wheel helper plugin
func testWheelManifest(name string) *plugins.PluginManifest {
	return &plugins.PluginManifest{
		Name:        name,
		Version:     "1.0.0",
		Description: "Test plugin",
		Author:      "test",
		Plugin: plugins.PluginBinaryConfig{
			Type:    "helper",
			Binary:  "ptx-" + name,
			Runtime: "python",
			Wheel:   "ptx_" + name + "-1.0.0-py3-none-any.whl",
		},
		Dependencies: plugins.PluginDependencies{OSSupport: []string{"linux", "windows", "darwin"}},
	}
}

// writePluginSource writes a plugin source directory with manifest, wheel and marker
func writePluginSource(t *testing.T, manifest *plugins.PluginManifest, marker string) string {
	t.Helper()
	dir := t.TempDir()
	if err := plugins.SaveManifest(manifest, filepath.Join(dir, "plugin.json")); err != nil {
		t.Fatal(err)
	}
	if manifest.Plugin.Wheel != "" {
		if err := os.WriteFile(filepath.Join(dir, manifest.Plugin.Wheel), []byte("wheel"), 0644); err != nil { // #nosec G306 -- test fixture
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, marker), []byte(marker), 0644); err != nil { // #nosec G306 -- test fixture
		t.Fatal(err)
	}
	return dir
}

// installOldVersion places a working "old" installation and registers it
func installOldVersion(t *testing.T, m *Manager, manifest *plugins.PluginManifest) string {
	t.Helper()
	pluginDir := filepath.Join(m.config.PluginsDir, manifest.Name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "old-marker"), []byte("old"), 0644); err != nil { // #nosec G306 -- test fixture
		t.Fatal(err)
	}
	if err := m.registry.RegisterPlugin(manifest, pluginDir); err != nil {
		t.Fatal(err)
	}
	return pluginDir
}

// fakePythonPath puts a shell-script python3 first on PATH (uv hidden); the
// script body decides which venv steps succeed
func fakePythonPath(t *testing.T, body string) {
	t.Helper()
	binDir := t.TempDir()
	script := "#!/bin/sh\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(binDir, "python3"), []byte(script), 0755); err != nil { // #nosec G306 -- test executable
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
}

// assertNoLeftovers checks that no staging or backup directories remain
func assertNoLeftovers(t *testing.T, pluginsDir string) {
	t.Helper()
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".staging-") || strings.HasPrefix(e.Name(), ".backup-") {
			t.Errorf("leftover directory: %s", e.Name())
		}
	}
}

// assertOldInstallIntact checks the previous files and registry entry survived
func assertOldInstallIntact(t *testing.T, m *Manager, pluginDir string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(pluginDir, "old-marker")); err != nil {
		t.Errorf("previous installation lost: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pluginDir, "new-marker")); err == nil {
		t.Error("new files leaked into the previous installation")
	}
	installPath, err := m.registry.GetPluginInstallPath("demo")
	if err != nil {
		t.Fatalf("registry entry lost: %v", err)
	}
	if installPath != pluginDir {
		t.Errorf("registry install path = %s, want %s", installPath, pluginDir)
	}
	assertNoLeftovers(t, m.config.PluginsDir)
}

func TestForceInstallPreCheckFailureKeepsOldInstall(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("fake python3 is a POSIX shell script")
	}

	m := newTestManager(t)
	manifest := testWheelManifest("demo")
	pluginDir := installOldVersion(t, m, manifest)
	source := writePluginSource(t, manifest, "new-marker")

	// venv module present but ensurepip missing (Debian python3 without python3-venv)
	fakePythonPath(t, `case "$*" in *ensurepip*) exit 1;; esac; exit 0`)

	err := m.ForceInstallPlugin(filepath.Join(source, "plugin.json"))
	if err == nil || !strings.Contains(err.Error(), "ensurepip") {
		t.Fatalf("expected ensurepip error, got %v", err)
	}
	if !strings.Contains(err.Error(), "portunix install uv") {
		t.Errorf("error does not suggest uv: %v", err)
	}
	assertOldInstallIntact(t, m, pluginDir)
}

func TestForceInstallSetupFailureRestoresOldInstall(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("fake python3 is a POSIX shell script")
	}

	m := newTestManager(t)
	manifest := testWheelManifest("demo")
	pluginDir := installOldVersion(t, m, manifest)
	source := writePluginSource(t, manifest, "new-marker")

	// All pre-checks pass, creating the venv itself fails
	fakePythonPath(t, `if [ "$1" = "-m" ] && [ "$2" = "venv" ] && [ "$3" != "--help" ]; then echo "venv failed" >&2; exit 1; fi; exit 0`)

	err := m.ForceInstallPlugin(filepath.Join(source, "plugin.json"))
	if err == nil || !strings.Contains(err.Error(), "previous installation restored") {
		t.Fatalf("expected restore error, got %v", err)
	}
	assertOldInstallIntact(t, m, pluginDir)
}

func TestForceInstallReplacesNativePlugin(t *testing.T) {
	m := newTestManager(t)
	manifest := &plugins.PluginManifest{
		Name:         "demo",
		Version:      "2.0.0",
		Description:  "Test plugin",
		Author:       "test",
		Plugin:       plugins.PluginBinaryConfig{Type: "helper", Binary: "ptx-demo"},
		Dependencies: plugins.PluginDependencies{OSSupport: []string{"linux", "windows", "darwin"}},
	}
	pluginDir := installOldVersion(t, m, manifest)
	source := writePluginSource(t, manifest, "new-marker")

	if err := m.ForceInstallPlugin(filepath.Join(source, "plugin.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(pluginDir, "new-marker")); err != nil {
		t.Errorf("new files not installed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pluginDir, "old-marker")); err == nil {
		t.Error("old files not replaced")
	}
	info, err := m.GetPlugin("demo")
	if err != nil || info.Version != "2.0.0" {
		t.Errorf("registry not updated: %+v, %v", info, err)
	}
	assertNoLeftovers(t, m.config.PluginsDir)
}

func TestRepairRecreatesBrokenVenv(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("fake python3 is a POSIX shell script")
	}

	m := newTestManager(t)
	manifest := testWheelManifest("demo")
	source := writePluginSource(t, manifest, "payload")

	// Installed copy with a venv bound to a removed interpreter
	pluginDir := filepath.Join(m.config.PluginsDir, "demo")
	if err := m.copyPluginFiles(filepath.Join(source, "plugin.json"), pluginDir); err != nil {
		t.Fatal(err)
	}
	_, interpreter := makeLinkedVenv(t, pluginDir)
	if err := m.registry.RegisterPlugin(manifest, pluginDir); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(interpreter); err != nil {
		t.Fatal(err)
	}
	if state, _, _, _ := m.CheckPluginInstall("demo"); state != InstallStateBroken {
		t.Fatalf("precondition: state = %s, want broken", state)
	}

	// Fake python3 creates a venv linked to a fresh interpreter and a no-op pip
	newInterpreter := filepath.Join(t.TempDir(), "python3.99")
	if err := os.WriteFile(newInterpreter, []byte("#!/bin/sh\n"), 0755); err != nil { // #nosec G306 -- test interpreter stub
		t.Fatal(err)
	}
	fakePythonPath(t, `if [ "$1" = "-m" ] && [ "$2" = "venv" ] && [ "$3" != "--help" ]; then
  /bin/mkdir -p "$3/bin"
  /bin/ln -s `+newInterpreter+` "$3/bin/python"
  printf '#!/bin/sh\nexit 0\n' > "$3/bin/pip"
  /bin/chmod +x "$3/bin/pip"
fi
exit 0`)

	if err := m.RepairPlugin("demo"); err != nil {
		t.Fatal(err)
	}
	if state, reason, _, _ := m.CheckPluginInstall("demo"); state != InstallStateOK {
		t.Fatalf("state after repair = %s (%s), want ok", state, reason)
	}
	if _, err := os.Stat(filepath.Join(pluginDir, "payload")); err != nil {
		t.Errorf("plugin files lost during repair: %v", err)
	}
	assertNoLeftovers(t, m.config.PluginsDir)
}

func TestRepairWithoutInstalledManifest(t *testing.T) {
	m := newTestManager(t)
	manifest := testWheelManifest("demo")
	pluginDir := filepath.Join(m.config.PluginsDir, "demo")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := m.registry.RegisterPlugin(manifest, pluginDir); err != nil {
		t.Fatal(err)
	}

	err := m.RepairPlugin("demo")
	if err == nil || !strings.Contains(err.Error(), "install --force") {
		t.Fatalf("expected hint to reinstall from source, got %v", err)
	}
}
