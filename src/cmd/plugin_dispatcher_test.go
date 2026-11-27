/*
 *  SPDX-License-Identifier: MIT
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

// envNameCountFold counts entries in env whose name equals key case-insensitively
func envNameCountFold(env []string, key string) int {
	n := 0
	for _, entry := range env {
		k, _, _ := strings.Cut(entry, "=")
		if strings.EqualFold(k, key) {
			n++
		}
	}
	return n
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

// TestPluginEnvironment checks how pluginEnvironment builds the plugin process
// environment: PORTUNIX_CWD is set exactly once, stale inherited values are
// dropped and all other variables pass through unchanged
func TestPluginEnvironment(t *testing.T) {
	// Inherited environment without PORTUNIX_CWD; EMPTY covers a variable with no value
	base := []string{"HOME=/home/user", "PATH=/usr/bin", "EMPTY="}

	// Caller directory is appended, other variables keep their order and values
	t.Run("adds caller directory and keeps other variables", func(t *testing.T) {
		got := pluginEnvironment(base, "/tmp/books")
		want := []string{"HOME=/home/user", "PATH=/usr/bin", "EMPTY=", "PORTUNIX_CWD=/tmp/books"}
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("pluginEnvironment() = %v, want %v", got, want)
		}
	})

	// PORTUNIX_CWD already set in the calling shell is replaced, not duplicated
	t.Run("replaces inherited value", func(t *testing.T) {
		env := append([]string{"PORTUNIX_CWD=/stale"}, base...)
		values := envValues(pluginEnvironment(env, "/tmp/books"), pluginCallerDirEnv)
		if len(values) != 1 || values[0] != "/tmp/books" {
			t.Errorf("PORTUNIX_CWD values = %v, want exactly [/tmp/books]", values)
		}
	})

	// Windows-only branch of pluginEnvironment: name match is case-insensitive there
	t.Run("differently cased inherited name", func(t *testing.T) {
		env := append([]string{"Portunix_Cwd=/stale"}, base...)
		got := pluginEnvironment(env, "/tmp/books")
		stale := envValues(got, "Portunix_Cwd")
		if runtime.GOOS == "windows" {
			// Windows environment names are case-insensitive, so the stale entry must go
			if len(stale) != 0 {
				t.Errorf("Portunix_Cwd values = %v, want none on Windows", stale)
			}
		} else if len(stale) != 1 || stale[0] != "/stale" {
			// Elsewhere it is an unrelated variable and must be kept
			t.Errorf("Portunix_Cwd values = %v, want [/stale]", stale)
		}
		if values := envValues(got, pluginCallerDirEnv); len(values) != 1 || values[0] != "/tmp/books" {
			t.Errorf("PORTUNIX_CWD values = %v, want exactly [/tmp/books]", values)
		}
	})

	// Value is stored verbatim, without quoting, escaping or re-encoding
	t.Run("caller directory with space and non-ASCII characters", func(t *testing.T) {
		dir := filepath.Join("/tmp", "knihy ř", "Přílohy")
		values := envValues(pluginEnvironment(base, dir), pluginCallerDirEnv)
		if len(values) != 1 || values[0] != dir {
			t.Errorf("PORTUNIX_CWD values = %v, want exactly [%s]", values, dir)
		}
	})

	// Empty caller directory (os.Getwd failed) removes the stale value and sets nothing
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

	// Function works on a copy, the os.Environ() slice passed in stays intact
	t.Run("does not modify input", func(t *testing.T) {
		env := []string{"PORTUNIX_CWD=/stale", "HOME=/home/user"}
		pluginEnvironment(env, "/tmp/books")
		if env[0] != "PORTUNIX_CWD=/stale" || env[1] != "HOME=/home/user" {
			t.Errorf("input environment was modified: %v", env)
		}
	})
}

// TestBuildPluginCommand_AllRuntimes checks that every plugin runtime gets the
// install path as working directory and the caller directory in PORTUNIX_CWD;
// commands are only built, not started, so no runtime needs to be installed
func TestBuildPluginCommand_AllRuntimes(t *testing.T) {
	installDir := t.TempDir()
	callerDir := t.TempDir()

	// One case per branch of the runtime switch in buildPluginCommand
	cases := []struct {
		name   string
		plugin manager.RegistryPlugin
	}{
		{"native", manager.RegistryPlugin{BinaryName: "stub"}},
		{"java", manager.RegistryPlugin{Runtime: "java", BinaryName: "stub.jar"}},
		{"python script", manager.RegistryPlugin{Runtime: "python", BinaryName: "stub.py"}},
		{"python wrapper", manager.RegistryPlugin{Runtime: "python", BinaryName: "stub.sh"}},
		{"python wheel", manager.RegistryPlugin{Runtime: "python", BinaryName: "stub", Wheel: "stub.whl"}},
		{"node", manager.RegistryPlugin{Runtime: "node", BinaryName: "src/main.js"}},
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

// TestBuildPluginCommand_Node checks that a Node.js plugin runs its entry
// script through node with the plugin args after the script path
func TestBuildPluginCommand_Node(t *testing.T) {
	installDir := t.TempDir()
	plugin := &manager.RegistryPlugin{Name: "js", Runtime: "node", BinaryName: "src/main.js", InstallPath: installDir}

	cmd := buildPluginCommand(plugin, []string{"hello", "Portunix"}, "")

	if filepath.Base(cmd.Path) != "node" && !strings.HasPrefix(strings.ToLower(filepath.Base(cmd.Path)), "node.") {
		t.Errorf("cmd.Path = %q, want node", cmd.Path)
	}
	wantArgs := []string{"node", filepath.Join(installDir, "src/main.js"), "hello", "Portunix"}
	if strings.Join(cmd.Args, "|") != strings.Join(wantArgs, "|") {
		t.Errorf("cmd.Args = %q, want %q", cmd.Args, wantArgs)
	}
}

// TestBuildPluginCommand_StubPluginReceivesCallerDir launches a real stub
// plugin and checks what the process itself sees: PORTUNIX_CWD is the absolute
// caller directory, while the working directory stays the plugin install path;
// on Windows the stub runs as .ps1 through powershell.exe, elsewhere as .sh through bash
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

	// Normalize Windows line endings; expected lines are PORTUNIX_CWD and working directory
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

	// Resolve symlinks (e.g. /tmp on macOS, 8.3 short names on Windows) before comparing
	if got, want := resolvedPath(t, lines[1]), resolvedPath(t, installDir); got != want {
		t.Errorf("plugin working directory = %q, want install path %q", got, want)
	}
}

// TestBuildPluginCommand_StubPluginReceivesSpecialCallerDir launches a real stub
// plugin from a caller directory containing a space and non-ASCII characters,
// with a differently cased stale variable inherited from the calling shell
func TestBuildPluginCommand_StubPluginReceivesSpecialCallerDir(t *testing.T) {
	installDir := t.TempDir()
	callerDir := filepath.Join(t.TempDir(), "knihy ř", "Přílohy")
	if err := os.MkdirAll(callerDir, 0o755); err != nil {
		t.Fatalf("failed to create caller directory: %v", err)
	}
	outFile := filepath.Join(installDir, "cwd.txt")

	// Stub writes PORTUNIX_CWD as UTF-8 into the file given as first argument,
	// so the check does not depend on the console code page
	scriptName := "stub.sh"
	script := "#!/bin/bash\nprintf '%s' \"$PORTUNIX_CWD\" > \"$1\"\n"
	if runtime.GOOS == "windows" {
		scriptName = "stub.ps1"
		script = "[System.IO.File]::WriteAllText($args[0], $env:PORTUNIX_CWD, (New-Object System.Text.UTF8Encoding $false))\r\n"
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

	// On Windows this name refers to PORTUNIX_CWD itself and must not leak through
	t.Setenv("Portunix_Cwd", installDir)

	cmd := buildPluginCommand(plugin, []string{outFile}, callerDir)

	// Check cmd.Env itself: on Windows os/exec deduplicates names case-insensitively
	// before starting the process, which would hide a leaked Portunix_Cwd from the stub
	want := 1 // the exported PORTUNIX_CWD
	if runtime.GOOS != "windows" {
		want = 2 // Portunix_Cwd is an unrelated variable there and is kept
	}
	if got := envNameCountFold(cmd.Env, pluginCallerDirEnv); got != want {
		t.Errorf("cmd.Env has %d entries named %s case-insensitively, want %d", got, pluginCallerDirEnv, want)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("stub plugin failed: %v\n%s", err, output)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read stub output: %v", err)
	}
	// Byte-exact comparison: any re-encoding or quoting would break it
	if got := string(data); got != callerDir {
		t.Errorf("PORTUNIX_CWD = %q, want caller directory %q", got, callerDir)
	}
}

// TestPluginCommandSilencesUsage checks that a failing plugin does not make
// cobra print the dispatcher command usage after the plugin error output
func TestPluginCommandSilencesUsage(t *testing.T) {
	cmd := createPluginCommand(&manager.RegistryPlugin{Name: "demo"})
	if !cmd.SilenceUsage {
		t.Error("plugin dispatcher command must not print usage on runtime errors")
	}
}

// TestPluginCheckArgs checks argument validation of `plugin check`: exactly one
// plugin name without --all, no name required with --all
func TestPluginCheckArgs(t *testing.T) {
	// Without --all
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
	// With --all; flag is package-level state, so reset it for other tests
	if err := pluginCheckCmd.Flags().Set("all", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = pluginCheckCmd.Flags().Set("all", "false") }()
	if err := pluginCheckCmd.Args(pluginCheckCmd, nil); err != nil {
		t.Errorf("check --all: unexpected error %v", err)
	}
}
