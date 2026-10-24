/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"portunix.ai/app/plugins/manager"
)

// envValues returns all values of key found in env
func envValues(env []string, key string) []string {
	var values []string
	for _, entry := range env {
		k, v, _ := strings.Cut(entry, "=")
		if k == key {
			values = append(values, v)
		}
	}
	return values
}

// resolvedPath resolves symlinks so paths reported by a process can be compared
func resolvedPath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(strings.TrimSpace(path))
	if err != nil {
		t.Fatalf("failed to resolve %q: %v", path, err)
	}
	return resolved
}

func TestPluginEnvironment(t *testing.T) {
	base := []string{"HOME=/home/user", "PATH=/usr/bin", "EMPTY="}

	t.Run("adds caller directory and keeps other variables", func(t *testing.T) {
		got := pluginEnvironment(base, "/tmp/books")
		want := []string{"HOME=/home/user", "PATH=/usr/bin", "EMPTY=", "PORTUNIX_CWD=/tmp/books"}
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("pluginEnvironment() = %v, want %v", got, want)
		}
	})

	t.Run("replaces inherited value", func(t *testing.T) {
		env := append([]string{"PORTUNIX_CWD=/stale"}, base...)
		values := envValues(pluginEnvironment(env, "/tmp/books"), pluginCallerDirEnv)
		if len(values) != 1 || values[0] != "/tmp/books" {
			t.Errorf("PORTUNIX_CWD values = %v, want exactly [/tmp/books]", values)
		}
	})

	t.Run("unknown caller directory leaves variable unset", func(t *testing.T) {
		env := append([]string{"PORTUNIX_CWD=/stale"}, base...)
		got := pluginEnvironment(env, "")
		if values := envValues(got, pluginCallerDirEnv); len(values) != 0 {
			t.Errorf("PORTUNIX_CWD values = %v, want none", values)
		}
		if len(got) != len(base) {
			t.Errorf("got %d variables, want %d", len(got), len(base))
		}
	})

	t.Run("does not modify input", func(t *testing.T) {
		env := []string{"PORTUNIX_CWD=/stale", "HOME=/home/user"}
		pluginEnvironment(env, "/tmp/books")
		if env[0] != "PORTUNIX_CWD=/stale" || env[1] != "HOME=/home/user" {
			t.Errorf("input environment was modified: %v", env)
		}
	})
}

func TestBuildPluginCommand_AllRuntimes(t *testing.T) {
	installDir := t.TempDir()
	callerDir := t.TempDir()

	cases := []struct {
		name   string
		plugin manager.RegistryPlugin
	}{
		{"native", manager.RegistryPlugin{BinaryName: "stub"}},
		{"java", manager.RegistryPlugin{Runtime: "java", BinaryName: "stub.jar"}},
		{"python script", manager.RegistryPlugin{Runtime: "python", BinaryName: "stub.py"}},
		{"python wrapper", manager.RegistryPlugin{Runtime: "python", BinaryName: "stub.sh"}},
		{"python wheel", manager.RegistryPlugin{Runtime: "python", BinaryName: "stub", Wheel: "stub.whl"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plugin := tc.plugin
			plugin.Name = "stub"
			plugin.InstallPath = installDir

			cmd := buildPluginCommand(&plugin, []string{"arg"}, callerDir)

			if cmd.Dir != installDir {
				t.Errorf("cmd.Dir = %q, want install path %q", cmd.Dir, installDir)
			}
			values := envValues(cmd.Env, pluginCallerDirEnv)
			if len(values) != 1 || values[0] != callerDir {
				t.Errorf("PORTUNIX_CWD values = %v, want exactly [%s]", values, callerDir)
			}
		})
	}
}

// TestBuildPluginCommand_StubPluginReceivesCallerDir launches a real stub
// plugin and checks what the process itself sees
func TestBuildPluginCommand_StubPluginReceivesCallerDir(t *testing.T) {
	installDir := t.TempDir()
	callerDir := t.TempDir()

	// Stub prints only PORTUNIX_CWD and its own working directory
	scriptName := "stub.sh"
	script := "#!/bin/bash\nprintf '%s\\n' \"$PORTUNIX_CWD\"\npwd -P\n"
	if runtime.GOOS == "windows" {
		scriptName = "stub.ps1"
		script = "Write-Output $env:PORTUNIX_CWD\r\nWrite-Output (Get-Location).ProviderPath\r\n"
	}
	if err := os.WriteFile(filepath.Join(installDir, scriptName), []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write stub plugin: %v", err)
	}

	plugin := &manager.RegistryPlugin{
		Name:        "stub",
		Runtime:     "python",
		BinaryName:  scriptName,
		InstallPath: installDir,
	}

	// Stale inherited value must be replaced by the real caller directory
	t.Setenv(pluginCallerDirEnv, installDir)

	cmd := buildPluginCommand(plugin, nil, callerDir)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("stub plugin failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(strings.ReplaceAll(string(output), "\r\n", "\n")), "\n")
	if len(lines) != 2 {
		t.Fatalf("unexpected stub output: %q", string(output))
	}

	gotCallerDir := strings.TrimSpace(lines[0])
	if gotCallerDir != callerDir {
		t.Errorf("PORTUNIX_CWD = %q, want caller directory %q", gotCallerDir, callerDir)
	}
	if gotCallerDir == installDir {
		t.Errorf("PORTUNIX_CWD points to the plugin install path %q", installDir)
	}
	if !filepath.IsAbs(gotCallerDir) {
		t.Errorf("PORTUNIX_CWD = %q, want absolute path", gotCallerDir)
	}

	if got, want := resolvedPath(t, lines[1]), resolvedPath(t, installDir); got != want {
		t.Errorf("plugin working directory = %q, want install path %q", got, want)
	}
}

func TestPluginCommandSilencesUsage(t *testing.T) {
	cmd := createPluginCommand(&manager.RegistryPlugin{Name: "demo"})
	if !cmd.SilenceUsage {
		t.Error("plugin dispatcher command must not print usage on runtime errors")
	}
}

func TestPluginCheckArgs(t *testing.T) {
	if err := pluginCheckCmd.Flags().Set("all", "false"); err != nil {
		t.Fatal(err)
	}
	if err := pluginCheckCmd.Args(pluginCheckCmd, nil); err == nil {
		t.Error("check without plugin name must be an argument error (shown with usage)")
	}
	if err := pluginCheckCmd.Args(pluginCheckCmd, []string{"demo"}); err != nil {
		t.Errorf("check demo: unexpected error %v", err)
	}
	if err := pluginCheckCmd.Args(pluginCheckCmd, []string{"a", "b"}); err == nil {
		t.Error("check with two names must be an argument error")
	}
	if err := pluginCheckCmd.Flags().Set("all", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = pluginCheckCmd.Flags().Set("all", "false") }()
	if err := pluginCheckCmd.Args(pluginCheckCmd, nil); err != nil {
		t.Errorf("check --all: unexpected error %v", err)
	}
}
