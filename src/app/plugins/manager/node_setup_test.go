/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"portunix.ai/app/plugins"
	"portunix.ai/portunix/src/pkg/platform"
)

const packageWithDeps = `{"name": "demo", "dependencies": {"yaml": "^2.4.2"}}`

// writeNodePlugin writes package.json (when not empty) into a new plugin dir
func writeNodePlugin(t *testing.T, packageJSON string) string {
	t.Helper()
	dir := t.TempDir()
	if packageJSON != "" {
		if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644); err != nil { // #nosec G306 -- test fixture
			t.Fatal(err)
		}
	}
	return dir
}

// fakeNodeTools puts stub node and npm executables first in PATH (without npm
// PATH holds only the stub dir, so no system npm is found); the npm
// stub records its arguments and creates node_modules like a real install
func fakeNodeTools(t *testing.T, withNPM bool) string {
	t.Helper()
	if platform.IsWindows() {
		t.Skip("shell-script tool stubs are POSIX-specific")
	}
	binDir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte("#!/bin/sh\n"+body+"\n"), 0755); err != nil { // #nosec G306 -- test executable
			t.Fatal(err)
		}
	}
	write("node", "echo v20.11.1")
	if withNPM {
		write("npm", `echo "$@" > npm-args && mkdir -p node_modules`)
		// The stub needs mkdir; a real npm in the system PATH stays shadowed
		t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
		return binDir
	}
	t.Setenv("PATH", binDir)
	return binDir
}

func TestNodeDependenciesMissing(t *testing.T) {
	tests := []struct {
		name        string
		packageJSON string
		nodeModules bool
		want        bool
		wantErr     bool
	}{
		{name: "no package.json"},
		{name: "no dependencies", packageJSON: `{"name": "demo", "devDependencies": {"eslint": "^9"}}`},
		{name: "dependencies missing", packageJSON: packageWithDeps, want: true},
		{name: "dependencies shipped", packageJSON: packageWithDeps, nodeModules: true},
		{name: "invalid package.json", packageJSON: `{`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeNodePlugin(t, tt.packageJSON)
			if tt.nodeModules {
				if err := os.Mkdir(filepath.Join(dir, nodeModulesDir), 0755); err != nil {
					t.Fatal(err)
				}
			}

			got, err := nodeDependenciesMissing(dir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("missing = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlanNodeDependencies(t *testing.T) {
	t.Run("nothing to install", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		plan, err := planNodeDependencies(writeNodePlugin(t, `{"name": "demo"}`))
		if err != nil || plan != nil {
			t.Fatalf("plan = %+v, err = %v, want no plan and no error", plan, err)
		}
	})

	t.Run("node missing", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		_, err := planNodeDependencies(writeNodePlugin(t, packageWithDeps))
		if err == nil || !strings.Contains(err.Error(), plugins.NodeInstallFix) {
			t.Fatalf("error = %v, want install hint %q", err, plugins.NodeInstallFix)
		}
	})

	t.Run("npm missing", func(t *testing.T) {
		fakeNodeTools(t, false)
		_, err := planNodeDependencies(writeNodePlugin(t, packageWithDeps))
		if err == nil || !strings.Contains(err.Error(), "npm not found") || !strings.Contains(err.Error(), plugins.NodeInstallFix) {
			t.Fatalf("error = %v, want npm not found with install hint", err)
		}
	})

	t.Run("install without lockfile", func(t *testing.T) {
		fakeNodeTools(t, true)
		plan, err := planNodeDependencies(writeNodePlugin(t, packageWithDeps))
		if err != nil {
			t.Fatal(err)
		}
		if plan.args[0] != "install" {
			t.Errorf("npm command = %q, want install", plan.args[0])
		}
	})

	t.Run("ci with lockfile", func(t *testing.T) {
		fakeNodeTools(t, true)
		dir := writeNodePlugin(t, packageWithDeps)
		if err := os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte("{}"), 0644); err != nil { // #nosec G306 -- test fixture
			t.Fatal(err)
		}
		plan, err := planNodeDependencies(dir)
		if err != nil {
			t.Fatal(err)
		}
		if plan.args[0] != "ci" {
			t.Errorf("npm command = %q, want ci", plan.args[0])
		}
	})
}

func TestSetupNodeDependencies(t *testing.T) {
	if err := setupNodeDependencies(nil, t.TempDir()); err != nil {
		t.Fatalf("nil plan must be a no-op: %v", err)
	}

	fakeNodeTools(t, true)
	dir := writeNodePlugin(t, packageWithDeps)
	plan, err := planNodeDependencies(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := setupNodeDependencies(plan, dir); err != nil {
		t.Fatal(err)
	}

	args, err := os.ReadFile(filepath.Join(dir, "npm-args"))
	if err != nil {
		t.Fatalf("npm did not run in the plugin dir: %v", err)
	}
	if got := strings.TrimSpace(string(args)); got != "install --omit=dev --no-audit --no-fund" {
		t.Errorf("npm args = %q", got)
	}
	if err := checkNodeModules(dir); err != nil {
		t.Errorf("node_modules must be present after setup: %v", err)
	}
}

func TestSetupNodeDependencies_Failure(t *testing.T) {
	binDir := fakeNodeTools(t, false)
	if err := os.WriteFile(filepath.Join(binDir, "npm"), []byte("#!/bin/sh\necho 'E404 not found' >&2\nexit 1\n"), 0755); err != nil { // #nosec G306 -- test executable
		t.Fatal(err)
	}
	dir := writeNodePlugin(t, packageWithDeps)
	plan, err := planNodeDependencies(dir)
	if err != nil {
		t.Fatal(err)
	}
	err = setupNodeDependencies(plan, dir)
	if err == nil || !strings.Contains(err.Error(), "E404 not found") {
		t.Fatalf("error = %v, want npm output in the error", err)
	}
}

func TestCheckInstall_Node(t *testing.T) {
	dir := writeNodePlugin(t, packageWithDeps)
	rp := &RegistryPlugin{Name: "js", Runtime: "node", BinaryName: "src/main.js", InstallPath: dir}

	if state, _ := rp.CheckInstall(); state != InstallStateBroken {
		t.Fatalf("state without node_modules = %s, want broken", state)
	}
	if err := os.Mkdir(filepath.Join(dir, nodeModulesDir), 0755); err != nil {
		t.Fatal(err)
	}
	if state, reason := rp.CheckInstall(); state != InstallStateOK {
		t.Fatalf("state = %s (%s), want ok", state, reason)
	}
}

func TestBinaryPath_Node(t *testing.T) {
	rp := &RegistryPlugin{Runtime: "node", BinaryName: "src/main.js", InstallPath: "/opt/plugins/js"}
	want := filepath.Join("/opt/plugins/js", "src/main.js")
	if got := rp.BinaryPath(); got != want {
		t.Errorf("BinaryPath() = %q, want %q (no executable suffix)", got, want)
	}
}

// testNodeManifest returns a valid manifest of a Node.js helper plugin
func testNodeManifest(name string) *plugins.PluginManifest {
	return &plugins.PluginManifest{
		Name:        name,
		Version:     "1.0.0",
		Description: "Test plugin",
		Author:      "test",
		Plugin:      plugins.PluginBinaryConfig{Type: "helper", Binary: "src/main.js", Runtime: "node"},
		Dependencies: plugins.PluginDependencies{
			OSSupport: []string{"linux", "windows", "darwin"},
		},
	}
}

// writeNodePluginSource writes a Node.js plugin source with dependencies
func writeNodePluginSource(t *testing.T, manifest *plugins.PluginManifest) string {
	t.Helper()
	dir := writeNodePlugin(t, packageWithDeps)
	if err := plugins.SaveManifest(manifest, filepath.Join(dir, "plugin.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "main.js"), []byte("console.log('ok')\n"), 0644); err != nil { // #nosec G306 -- test fixture
		t.Fatal(err)
	}
	return dir
}

// captureStdout returns what fn prints to os.Stdout
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	done := make(chan []byte)
	go func() {
		data, _ := io.ReadAll(r)
		done <- data
	}()
	fn()
	_ = w.Close()
	os.Stdout = stdout
	return string(<-done)
}

// TestInstallPlugin_NodeMissing checks that a missing Node.js stops the
// install before anything is copied and without a "will be installed" warning
func TestInstallPlugin_NodeMissing(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("PATH isolation is POSIX-specific")
	}
	t.Setenv("PATH", t.TempDir())
	m := newTestManager(t)
	source := writeNodePluginSource(t, testNodeManifest("jsdemo"))

	var err error
	output := captureStdout(t, func() { err = m.InstallPlugin(filepath.Join(source, "plugin.json")) })

	if err == nil || !strings.Contains(err.Error(), "Node.js runtime not found") || !strings.Contains(err.Error(), plugins.NodeInstallFix) {
		t.Fatalf("error = %v, want Node.js runtime not found with install hint", err)
	}
	if strings.Contains(output, "will be installed") {
		t.Errorf("output contains the install warning before the hard error:\n%s", output)
	}
	if _, statErr := os.Stat(filepath.Join(m.config.PluginsDir, "jsdemo")); !os.IsNotExist(statErr) {
		t.Errorf("plugin directory left behind: %v", statErr)
	}
}

// TestForceInstallPlugin_NodeNPMFailureKeepsOld checks the atomic reinstall and
// that the "left unchanged" note comes before the npm output
func TestForceInstallPlugin_NodeNPMFailureKeepsOld(t *testing.T) {
	binDir := fakeNodeTools(t, false)
	if err := os.WriteFile(filepath.Join(binDir, "npm"), []byte("#!/bin/sh\necho 'npm ERR! simulated' >&2\nexit 1\n"), 0755); err != nil { // #nosec G306 -- test executable
		t.Fatal(err)
	}
	m := newTestManager(t)
	manifest := testNodeManifest("jsdemo")
	pluginDir := installOldVersion(t, m, manifest)
	source := writeNodePluginSource(t, manifest)

	var err error
	captureStdout(t, func() { err = m.ForceInstallPlugin(filepath.Join(source, "plugin.json")) })

	want := "failed to setup Node.js plugin (existing installation left unchanged): npm install failed"
	if err == nil || !strings.HasPrefix(err.Error(), want) || !strings.Contains(err.Error(), "npm ERR! simulated") {
		t.Fatalf("error = %v, want prefix %q and the npm output", err, want)
	}
	if _, statErr := os.Stat(filepath.Join(pluginDir, "old-marker")); statErr != nil {
		t.Errorf("old installation was changed: %v", statErr)
	}
}

func TestCheckHelperPluginHealth_NodeRuntime(t *testing.T) {
	dir := writeNodePlugin(t, "")
	if err := os.WriteFile(filepath.Join(dir, "main.js"), []byte("console.log('1.0.0')\n"), 0644); err != nil { // #nosec G306 -- test fixture
		t.Fatal(err)
	}
	plugin := &RegistryPlugin{Name: "jsdemo", Version: "1.0.0", Mode: "helper", Runtime: "node", BinaryName: "main.js", InstallPath: dir}
	m := newTestManager(t)

	t.Run("runtime missing", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		health := m.checkHelperPluginHealth(plugin)
		if health.Healthy || health.Status != "runtime_missing" || !strings.Contains(health.Message, plugins.NodeInstallFix) {
			t.Fatalf("health = %+v, want unhealthy runtime_missing with install hint", health)
		}
	})

	t.Run("runtime present", func(t *testing.T) {
		fakeNodeTools(t, false)
		health := m.checkHelperPluginHealth(plugin)
		if !health.Healthy {
			t.Fatalf("health = %+v, want healthy", health)
		}
	})
}
