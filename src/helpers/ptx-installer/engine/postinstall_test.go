/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// withSudoPrefix replaces the sudo prefix decision for the duration of a test
func withSudoPrefix(t *testing.T, prefix string) {
	t.Helper()
	orig := postInstallSudoPrefix
	postInstallSudoPrefix = func() string { return prefix }
	t.Cleanup(func() { postInstallSudoPrefix = orig })
}

// TestExpandPostInstallCommand covers the placeholders manifests rely on
// (issue #227)
func TestExpandPostInstallCommand(t *testing.T) {
	t.Setenv("PTX_TEST_227_SET", "from-env")

	tests := []struct {
		name        string
		cmd         string
		installPath string
		sudoPrefix  string
		want        string
	}{
		{"hugo as root", "${sudo_prefix}chmod +x ${actual_extract_to}/hugo", "/usr/local/bin", "",
			"chmod +x /usr/local/bin/hugo"},
		{"hugo as non-root", "${sudo_prefix}chmod +x ${actual_extract_to}/hugo", "/usr/local/bin", "sudo ",
			"sudo chmod +x /usr/local/bin/hugo"},
		{"fallback directory", "chmod +x ${actual_extract_to}/hugo", "/home/u/.local/share/portunix/packages/hugo", "",
			"chmod +x /home/u/.local/share/portunix/packages/hugo/hugo"},
		{"install_path both forms", "ln -s ${install_path}/bin %install_path%/b", "/opt/x", "",
			"ln -s /opt/x/bin /opt/x/b"},
		{"sudo in a pipeline", "echo x | ${sudo_prefix}tee /etc/profile.d/go.sh", "/usr/local", "sudo ",
			"echo x | sudo tee /etc/profile.d/go.sh"},
		{"sudo without install path", "${sudo_prefix}ln -sf /usr/bin/node /usr/local/bin/node || true", "", "sudo ",
			"sudo ln -sf /usr/bin/node /usr/local/bin/node || true"},
		{"set environment variable", "echo ${PTX_TEST_227_SET}", "", "", "echo from-env"},
		{"plain shell variable untouched", "echo 'export PATH=$PATH:/usr/local/go/bin'", "", "",
			"echo 'export PATH=$PATH:/usr/local/go/bin'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withSudoPrefix(t, tt.sudoPrefix)
			got, err := expandPostInstallCommand(tt.cmd, tt.installPath)
			if err != nil {
				t.Fatalf("expandPostInstallCommand(%q) unexpected error: %v", tt.cmd, err)
			}
			if got != tt.want {
				t.Errorf("expandPostInstallCommand(%q) = %q, want %q", tt.cmd, got, tt.want)
			}
		})
	}
}

// TestExpandPostInstallCommand_UnknownPlaceholder verifies that a placeholder
// the engine cannot resolve is an error naming it, never an empty string
func TestExpandPostInstallCommand_UnknownPlaceholder(t *testing.T) {
	withSudoPrefix(t, "")

	tests := []struct {
		name        string
		cmd         string
		installPath string
		wantNamed   []string
	}{
		{"unsupported placeholder", "tar -C /usr/local -xzf ${downloaded_file}", "/usr/local",
			[]string{"${downloaded_file}"}},
		{"path placeholder without install path", "chmod +x ${actual_extract_to}/hugo", "",
			[]string{"${actual_extract_to}"}},
		{"several placeholders", "cp ${foo} ${bar} ${foo}", "/opt/x", []string{"${bar}, ${foo}"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandPostInstallCommand(tt.cmd, tt.installPath)
			if err == nil {
				t.Fatalf("expandPostInstallCommand(%q) = %q, want error", tt.cmd, got)
			}
			for _, name := range tt.wantNamed {
				if !strings.Contains(err.Error(), name) {
					t.Errorf("error %q does not name %s", err, name)
				}
			}
		})
	}
}

// TestRunPostInstallCommands_UnknownPlaceholderNotExecuted verifies that a
// command with an unknown placeholder is reported and never reaches the shell
func TestRunPostInstallCommands_UnknownPlaceholderNotExecuted(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX shell commands")
	}
	withSudoPrefix(t, "")

	dir := t.TempDir()
	failures := runPostInstallCommands([]string{
		"touch ${install_path}/bad${nope}",
		"touch ${actual_extract_to}/good",
	}, dir, "")

	if len(failures) != 1 || !strings.Contains(failures[0], "${nope}") {
		t.Fatalf("failures = %q, want one failure naming ${nope}", failures)
	}
	if _, err := os.Stat(filepath.Join(dir, "bad")); err == nil {
		t.Error("command with unknown placeholder was executed")
	}
	if _, err := os.Stat(filepath.Join(dir, "good")); err != nil {
		t.Errorf("valid command after a failed one did not run: %v", err)
	}
}

// TestRunPostInstallCommands_RewritesOriginalPath verifies that a literal
// extractTo path in a command follows the package to its real location
func TestRunPostInstallCommands_RewritesOriginalPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX shell commands")
	}
	withSudoPrefix(t, "")

	dir := t.TempDir()
	failures := runPostInstallCommands([]string{"touch /opt/ptx-test-227/marker"}, dir, "/opt/ptx-test-227")
	if len(failures) != 0 {
		t.Fatalf("unexpected failures: %q", failures)
	}
	if _, err := os.Stat(filepath.Join(dir, "marker")); err != nil {
		t.Errorf("command was not redirected to the real install path: %v", err)
	}
}

// collectPostInstall gathers every postInstall command in a manifest together
// with the JSON path it was found at
func collectPostInstall(node interface{}, path string, out map[string][]string) {
	switch v := node.(type) {
	case map[string]interface{}:
		for key, child := range v {
			if key == "postInstall" {
				if cmds, ok := child.([]interface{}); ok {
					for _, c := range cmds {
						if s, ok := c.(string); ok {
							out[path+"/postInstall"] = append(out[path+"/postInstall"], s)
						}
					}
				}
				continue
			}
			collectPostInstall(child, path+"/"+key, out)
		}
	case []interface{}:
		for _, child := range v {
			collectPostInstall(child, path, out)
		}
	}
}

// TestPackageManifests_PostInstallPlaceholders is a guard over every package
// manifest: postInstall may only use ${…} placeholders the engine substitutes,
// and must use ${sudo_prefix} instead of a hard-coded sudo (issue #227)
func TestPackageManifests_PostInstallPlaceholders(t *testing.T) {
	manifests, err := filepath.Glob(filepath.Join("..", "assets", "packages", "*.json"))
	if err != nil || len(manifests) == 0 {
		t.Fatalf("no package manifests found: %v", err)
	}

	supported := make(map[string]bool)
	for _, p := range PostInstallPlaceholders {
		supported[p] = true
	}

	for _, manifest := range manifests {
		data, err := os.ReadFile(manifest)
		if err != nil {
			t.Fatalf("cannot read %s: %v", manifest, err)
		}
		var root interface{}
		if err := json.Unmarshal(data, &root); err != nil {
			t.Fatalf("cannot parse %s: %v", manifest, err)
		}

		commands := make(map[string][]string)
		collectPostInstall(root, "", commands)

		name := filepath.Base(manifest)
		for path, cmds := range commands {
			for _, cmd := range cmds {
				for _, placeholder := range unresolvedPlaceholderRe.FindAllString(cmd, -1) {
					if !supported[placeholder] {
						t.Errorf("%s %s: unsupported placeholder %s in %q (supported: %s)",
							name, path, placeholder, cmd, strings.Join(PostInstallPlaceholders, ", "))
					}
				}
				for _, word := range strings.Fields(cmd) {
					if word == "sudo" {
						t.Errorf("%s %s: hard-coded sudo in %q — use ${sudo_prefix}", name, path, cmd)
					}
				}
			}
		}
	}
}
