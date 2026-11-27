/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"portunix.ai/app/plugins"
	"portunix.ai/app/plugins/manager"
)

// newStubPlugin writes a stub plugin script (bash, or PowerShell on Windows)
// and returns an enabled registry entry for it
func newStubPlugin(t *testing.T, bashScript, ps1Script string) *manager.RegistryPlugin {
	t.Helper()
	installDir := t.TempDir()
	name, script := "stub.sh", "#!/bin/bash\n"+bashScript
	if runtime.GOOS == "windows" {
		name, script = "stub.ps1", ps1Script
	}
	if err := os.WriteFile(filepath.Join(installDir, name), []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write stub plugin: %v", err)
	}
	return &manager.RegistryPlugin{
		Name:        "stub",
		Version:     "1.2.3",
		Description: "stub plugin",
		Runtime:     "python",
		BinaryName:  name,
		InstallPath: installDir,
		Enabled:     true,
		Commands: []plugins.PluginCommand{
			{
				Name:        "run",
				Description: "Run the stub",
				Parameters:  []plugins.PluginParameter{{Name: "url", Type: "string", Required: true, Description: "Target URL"}},
				Examples:    []string{"portunix stub run https://example.com"},
			},
		},
	}
}

// decodeHelpAI parses the --help-ai output as a JSON object
func decodeHelpAI(t *testing.T, out []byte) map[string]any {
	t.Helper()
	var data map[string]any
	if err := json.Unmarshal(out, &data); err != nil {
		t.Fatalf("output is not a JSON object: %v\n%s", err, out)
	}
	return data
}

// TestGeneratePluginAIHelp checks the registry fallback JSON (issue #257)
func TestGeneratePluginAIHelp(t *testing.T) {
	plugin := &manager.RegistryPlugin{
		Name:        "scraper",
		Version:     "0.4.0",
		Description: "Web scraper",
		Mode:        "helper",
		Commands: []plugins.PluginCommand{
			{
				Name:        "fetch",
				Description: "Fetch a page",
				Parameters:  []plugins.PluginParameter{{Name: "url", Type: "string", Required: true, DefaultValue: "x"}},
			},
		},
	}
	out, err := GeneratePluginAIHelp(plugin)
	if err != nil {
		t.Fatalf("GeneratePluginAIHelp failed: %v", err)
	}

	var got pluginAIHelp
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if got.Tool != "scraper" || got.Version != "0.4.0" || got.Source != "registry" || got.Mode != "helper" || got.Runtime != "native" {
		t.Errorf("unexpected header: %+v", got)
	}
	if len(got.Commands) != 1 || got.Commands[0].Name != "fetch" {
		t.Fatalf("unexpected commands: %+v", got.Commands)
	}
	p := got.Commands[0].Parameters
	if len(p) != 1 || p[0].Name != "url" || !p[0].Required || p[0].Default != "x" {
		t.Errorf("unexpected parameters: %+v", p)
	}
	// Empty lists must be [] for clients, never null
	if strings.Contains(out, "null") {
		t.Errorf("output contains null:\n%s", out)
	}
	// Usage placeholders stay readable, not HTML-escaped as \u003c
	if !strings.Contains(out, `"usage": "portunix scraper <command> [args...]"`) || strings.Contains(out, `\u003c`) {
		t.Errorf("usage is escaped or missing:\n%s", out)
	}
	if strings.HasSuffix(out, "\n") {
		t.Error("output must not end with a newline, the caller adds one")
	}
}

// TestGeneratePluginAIHelp_NoCommands checks an empty commands list
func TestGeneratePluginAIHelp_NoCommands(t *testing.T) {
	out, err := GeneratePluginAIHelp(&manager.RegistryPlugin{Name: "bare"})
	if err != nil {
		t.Fatalf("GeneratePluginAIHelp failed: %v", err)
	}
	if !strings.Contains(out, `"commands": []`) {
		t.Errorf("expected empty commands list:\n%s", out)
	}
}

// TestRunPluginHelpAI_PluginAnswers checks that a plugin answering with JSON
// gets its own JSON printed, with the arguments and PORTUNIX_HELP_AI passed
func TestRunPluginHelpAI_PluginAnswers(t *testing.T) {
	plugin := newStubPlugin(t,
		`printf '{"source":"plugin","help_ai":"%s","args":"%s"}\n' "$PORTUNIX_HELP_AI" "$*"`+"\n",
		`Write-Output ('{"source":"plugin","help_ai":"' + $env:PORTUNIX_HELP_AI + '","args":"' + ($args -join ' ') + '"}')`+"\r\n")

	var out bytes.Buffer
	if err := runPluginHelpAI(plugin, []string{"run", "--help-ai"}, &out, 10*time.Second); err != nil {
		t.Fatalf("runPluginHelpAI failed: %v", err)
	}
	data := decodeHelpAI(t, out.Bytes())
	if data["source"] != "plugin" {
		t.Errorf("source = %v, want plugin answer", data["source"])
	}
	if data["help_ai"] != "1" {
		t.Errorf("%s = %v, want 1", pluginHelpAIEnv, data["help_ai"])
	}
	if data["args"] != "run --help-ai" {
		t.Errorf("args = %v, want %q", data["args"], "run --help-ai")
	}
}

// TestRunPluginHelpAI_Fallback checks the registry fallback for plugins that
// do not answer with a JSON object
func TestRunPluginHelpAI_Fallback(t *testing.T) {
	cases := []struct {
		name, bash, ps1 string
	}{
		{"unknown option", "echo 'No such option: --help-ai' >&2\nexit 2\n", "Write-Error 'No such option: --help-ai'\r\nexit 2\r\n"},
		{"text output", "echo 'Usage: stub [command]'\n", "Write-Output 'Usage: stub [command]'\r\n"},
		{"json with failure", "echo '{\"a\":1}'\nexit 1\n", "Write-Output '{\"a\":1}'\r\nexit 1\r\n"},
		{"json array", "echo '[1,2]'\n", "Write-Output '[1,2]'\r\n"},
		{"no output", "exit 0\n", "exit 0\r\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plugin := newStubPlugin(t, c.bash, c.ps1)
			var out bytes.Buffer
			if err := runPluginHelpAI(plugin, []string{"--help-ai"}, &out, 10*time.Second); err != nil {
				t.Fatalf("runPluginHelpAI failed: %v", err)
			}
			data := decodeHelpAI(t, out.Bytes())
			if data["source"] != "registry" || data["tool"] != "stub" {
				t.Errorf("expected registry fallback, got:\n%s", out.String())
			}
		})
	}
}

// TestRunPluginHelpAI_NotStarted checks that a disabled or missing plugin is
// not started and the registry answers
func TestRunPluginHelpAI_NotStarted(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran")
	plugin := newStubPlugin(t, "touch '"+marker+"'\necho '{}'\n", "New-Item -ItemType File '"+marker+"'\r\nWrite-Output '{}'\r\n")

	disabled := *plugin
	disabled.Enabled = false
	missing := *plugin
	missing.InstallPath = filepath.Join(t.TempDir(), "gone")

	for name, p := range map[string]*manager.RegistryPlugin{"disabled": &disabled, "missing": &missing} {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			if err := runPluginHelpAI(p, []string{"--help-ai"}, &out, 10*time.Second); err != nil {
				t.Fatalf("runPluginHelpAI failed: %v", err)
			}
			if data := decodeHelpAI(t, out.Bytes()); data["source"] != "registry" {
				t.Errorf("expected registry fallback, got:\n%s", out.String())
			}
			if _, err := os.Stat(marker); err == nil {
				t.Error("plugin was started")
			}
		})
	}
}

// TestRunPluginHelpAI_Timeout checks that a hanging plugin is killed and the
// timeout is an error without registry fallback
func TestRunPluginHelpAI_Timeout(t *testing.T) {
	plugin := newStubPlugin(t, "exec sleep 30\n", "Start-Sleep -Seconds 30\r\n")

	var out bytes.Buffer
	start := time.Now()
	err := runPluginHelpAI(plugin, []string{"--help-ai"}, &out, 300*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "did not answer --help-ai within") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("timeout must not print a fallback, got:\n%s", out.String())
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("timeout took %s, plugin was not killed", elapsed)
	}
}

// TestGuardHelpLevelFlags_SelfHandledHelpAI checks that a command marked as
// handling --help-ai itself runs for --help-ai and shows help for
// --help-expert (plugin commands, issue #257)
func TestGuardHelpLevelFlags_SelfHandledHelpAI(t *testing.T) {
	cases := []struct {
		args    []string
		wantRun bool
	}{
		{[]string{"sub", "--help-ai"}, true},
		{[]string{"sub", "run", "--help-ai"}, true},
		{[]string{"sub", "--help-expert"}, false},
		{[]string{"sub", "run"}, true},
	}
	for _, c := range cases {
		t.Run(strings.Join(c.args, " "), func(t *testing.T) {
			t.Cleanup(func() { helpExpert, helpAI = false, false })
			ran := false
			root := newHelpFlagTestTree(&ran, true)
			sub, _, _ := root.Find([]string{"sub"})
			sub.Annotations = map[string]string{selfHandledHelpAIAnnotation: "true"}
			guardHelpLevelFlags(root)
			root.SetArgs(c.args)
			if err := root.Execute(); err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if ran != c.wantRun {
				t.Errorf("ran = %v, want %v", ran, c.wantRun)
			}
		})
	}
}

// TestPluginCommandHandlesHelpAI checks that plugin commands are marked to
// handle --help-ai themselves
func TestPluginCommandHandlesHelpAI(t *testing.T) {
	cmd := createPluginCommand(&manager.RegistryPlugin{Name: "demo"})
	if cmd.Annotations[selfHandledHelpAIAnnotation] != "true" {
		t.Error("plugin command must handle --help-ai itself")
	}
}
