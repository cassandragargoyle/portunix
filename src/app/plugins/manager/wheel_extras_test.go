/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"portunix.ai/app/plugins"
	"portunix.ai/portunix/src/pkg/platform"
)

// extrasManifest returns a wheel manifest that declares the extras "train" and "gpu"
func extrasManifest() *plugins.PluginManifest {
	manifest := testWheelManifest("demo")
	manifest.Plugin.OptionalExtras = []plugins.WheelExtra{
		{Name: "train", Description: "Model training"},
		{Name: "gpu"},
	}
	return manifest
}

// fakePipLog installs a fake python3 whose venv pip appends its arguments to
// the returned log file, one invocation per line
func fakePipLog(t *testing.T) string {
	t.Helper()
	logFile := filepath.Join(t.TempDir(), "pip.log")
	fakePythonPath(t, `if [ "$1" = "-m" ] && [ "$2" = "venv" ] && [ "$3" != "--help" ]; then
  /bin/mkdir -p "$3/bin"
  printf '#!/bin/sh\necho "$@" >> `+logFile+`\n' > "$3/bin/pip"
  /bin/chmod +x "$3/bin/pip"
fi
exit 0`)
	return logFile
}

// fakeUVLog puts a fake uv first on PATH that appends its arguments to the
// returned log file, one invocation per line
func fakeUVLog(t *testing.T) string {
	t.Helper()
	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "uv.log")
	script := "#!/bin/sh\necho \"$@\" >> " + logFile + "\nexit 0\n"
	if err := os.WriteFile(filepath.Join(binDir, "uv"), []byte(script), 0755); err != nil { // #nosec G306 -- test executable
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	return logFile
}

// readLog returns the logged invocations
func readLog(t *testing.T, logFile string) []string {
	t.Helper()
	data, err := os.ReadFile(logFile) // #nosec G304 -- test log in temp dir
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

func TestWheelRequirement(t *testing.T) {
	if got := wheelRequirement("/p/x.whl", nil); got != "/p/x.whl" {
		t.Errorf("without extras = %q", got)
	}
	if got := wheelRequirement("/p/x.whl", []string{"gpu", "train"}); got != "/p/x.whl[gpu,train]" {
		t.Errorf("with extras = %q", got)
	}
}

func TestDroppedExtras(t *testing.T) {
	got := droppedExtras([]string{"train", "gpu"}, []string{"Train"})
	if !reflect.DeepEqual(got, []string{"gpu"}) {
		t.Errorf("droppedExtras = %v, want [gpu]", got)
	}
	if got := droppedExtras(nil, []string{"train"}); len(got) != 0 {
		t.Errorf("droppedExtras with nothing installed = %v", got)
	}
}

func TestInstallWithExtrasPip(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("fake python3 is a POSIX shell script")
	}
	m := newTestManager(t)
	source := writePluginSource(t, extrasManifest(), "payload")
	logFile := fakePipLog(t)

	opts := InstallOptions{Extras: []string{"train,gpu", "train"}}
	if err := m.InstallPluginWithOptions(filepath.Join(source, "plugin.json"), opts); err != nil {
		t.Fatal(err)
	}

	calls := readLog(t, logFile)
	last := calls[len(calls)-1]
	if !strings.HasSuffix(last, "ptx_demo-1.0.0-py3-none-any.whl[gpu,train]") {
		t.Errorf("pip install argument = %q, want wheel with [gpu,train]", last)
	}
	data, err := m.GetPluginRegistryData("demo")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data.InstalledExtras, []string{"gpu", "train"}) {
		t.Errorf("registry installed extras = %v, want [gpu train]", data.InstalledExtras)
	}
	if len(data.OptionalExtras) != 2 {
		t.Errorf("registry optional extras = %v, want 2 entries", data.OptionalExtras)
	}
}

func TestInstallWithoutExtrasPip(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("fake python3 is a POSIX shell script")
	}
	m := newTestManager(t)
	source := writePluginSource(t, extrasManifest(), "payload")
	logFile := fakePipLog(t)

	if err := m.InstallPlugin(filepath.Join(source, "plugin.json")); err != nil {
		t.Fatal(err)
	}

	for _, call := range readLog(t, logFile) {
		if strings.Contains(call, "[") {
			t.Errorf("default install requested extras: %q", call)
		}
	}
	data, err := m.GetPluginRegistryData("demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(data.InstalledExtras) != 0 {
		t.Errorf("registry installed extras = %v, want none", data.InstalledExtras)
	}
}

func TestInstallWithExtrasUV(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("fake uv is a POSIX shell script")
	}
	m := newTestManager(t)
	source := writePluginSource(t, extrasManifest(), "payload")
	logFile := fakeUVLog(t)

	if err := m.InstallPluginWithOptions(filepath.Join(source, "plugin.json"), InstallOptions{Extras: []string{"train"}}); err != nil {
		t.Fatal(err)
	}

	calls := readLog(t, logFile)
	last := calls[len(calls)-1]
	if !strings.HasPrefix(last, "pip install --python ") || !strings.HasSuffix(last, ".whl[train]") {
		t.Errorf("uv install call = %q, want uv pip install of the wheel with [train]", last)
	}
}

func TestInstallUnknownExtraInstallsNothing(t *testing.T) {
	m := newTestManager(t)
	source := writePluginSource(t, extrasManifest(), "payload")

	err := m.InstallPluginWithOptions(filepath.Join(source, "plugin.json"), InstallOptions{Extras: []string{"cuda"}})
	if err == nil {
		t.Fatal("expected error for undeclared extra")
	}
	if !strings.Contains(err.Error(), "train") || !strings.Contains(err.Error(), "gpu") {
		t.Errorf("error does not list declared extras: %v", err)
	}
	if _, err := os.Stat(filepath.Join(m.config.PluginsDir, "demo")); err == nil {
		t.Error("plugin directory created despite invalid extra")
	}
	if _, err := m.registry.GetPlugin("demo"); err == nil {
		t.Error("plugin registered despite invalid extra")
	}
}

func TestForceInstallUnknownExtraKeepsOldInstall(t *testing.T) {
	m := newTestManager(t)
	manifest := extrasManifest()
	pluginDir := installOldVersion(t, m, manifest)
	source := writePluginSource(t, manifest, "new-marker")

	if err := m.ForceInstallPluginWithOptions(filepath.Join(source, "plugin.json"), InstallOptions{Extras: []string{"cuda"}}); err == nil {
		t.Fatal("expected error for undeclared extra")
	}
	assertOldInstallIntact(t, m, pluginDir)
}

func TestRepairKeepsInstalledExtras(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("fake python3 is a POSIX shell script")
	}
	m := newTestManager(t)
	source := writePluginSource(t, extrasManifest(), "payload")
	logFile := fakePipLog(t)

	if err := m.InstallPluginWithOptions(filepath.Join(source, "plugin.json"), InstallOptions{Extras: []string{"train"}}); err != nil {
		t.Fatal(err)
	}
	if err := m.RepairPlugin("demo"); err != nil {
		t.Fatal(err)
	}

	calls := readLog(t, logFile)
	if last := calls[len(calls)-1]; !strings.HasSuffix(last, ".whl[train]") {
		t.Errorf("repair pip argument = %q, want wheel with [train]", last)
	}
	data, err := m.GetPluginRegistryData("demo")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data.InstalledExtras, []string{"train"}) {
		t.Errorf("installed extras after repair = %v, want [train]", data.InstalledExtras)
	}
}
