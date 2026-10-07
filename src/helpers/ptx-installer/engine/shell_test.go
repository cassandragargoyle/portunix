/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"runtime"
	"strings"
	"testing"
)

// TestShellCommand_QuotedArgumentsReachShellUnchanged verifies that an inline
// script with double-quoted arguments reaches the shell verbatim (issue #201).
// Go's default Windows escaping turned `"` into `\"`, which cmd.exe kept
func TestShellCommand_QuotedArgumentsReachShellUnchanged(t *testing.T) {
	script := `echo "hello world"`
	want := `"hello world"`
	if runtime.GOOS != "windows" {
		// sh removes the quotes as part of normal word parsing
		want = "hello world"
	}

	out, err := shellCommand(script).CombinedOutput()
	if err != nil {
		t.Fatalf("shellCommand(%q) failed: %v\n%s", script, err, out)
	}
	got := strings.TrimSpace(string(out))
	if got != want {
		t.Errorf("shellCommand(%q) output = %q, want %q", script, got, want)
	}
}

// TestShellCommand_NestedQuotesForPowerShell verifies the uv install pattern:
// a cmd line that hands a double-quoted script to PowerShell
func TestShellCommand_NestedQuotesForPowerShell(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell inline scripts are Windows-only")
	}

	script := `powershell -NoProfile -ExecutionPolicy ByPass -c "Write-Output ('ptx' + '-ok')"`
	out, err := shellCommand(script).CombinedOutput()
	if err != nil {
		t.Fatalf("shellCommand(%q) failed: %v\n%s", script, err, out)
	}
	if got := strings.TrimSpace(string(out)); got != "ptx-ok" {
		t.Errorf("PowerShell received the script as a literal instead of running it, output = %q", got)
	}
}

// TestShellCommand_ExitCodePropagates verifies a failing script reports an error
func TestShellCommand_ExitCodePropagates(t *testing.T) {
	if err := shellCommand("exit 3").Run(); err == nil {
		t.Error("expected error for a script exiting with code 3, got nil")
	}
}

// scriptPackage returns a type "script" package manifest for every platform
// with the given install script and verification command
func scriptPackage(name, installScript, verification string) string {
	platform := `{
        "type": "script",
        "variants": { "latest": { "version": "latest", "installScript": "` + installScript + `" } },
        "verification": { "command": "` + verification + `", "expectedExitCode": 0 }
      }`
	return `{
  "apiVersion": "v1",
  "kind": "Package",
  "metadata": { "name": "` + name + `", "displayName": "Test Script", "description": "test", "category": "development/tools" },
  "spec": {
    "hasVariants": true,
    "platforms": { "windows": ` + platform + `, "linux": ` + platform + `, "darwin": ` + platform + ` }
  }
}`
}

// TestInstallScript_VerificationFailureFails verifies that a script which
// exits 0 but installs nothing makes Install fail (issue #201)
func TestInstallScript_VerificationFailureFails(t *testing.T) {
	assets := writeTestAssets(t, map[string]string{
		"noop-script.json": scriptPackage("noop-script", "exit 0", "exit 1"),
	})
	installer, err := NewInstaller(assets)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	err = installer.Install(&InstallOptions{PackageName: "noop-script"})
	if err == nil {
		t.Fatal("expected verification error, got nil")
	}
	if !strings.Contains(err.Error(), "verification failed") {
		t.Errorf("error should say verification failed, got: %v", err)
	}
}

// TestInstallScript_VerificationSuccessPasses verifies the happy path
func TestInstallScript_VerificationSuccessPasses(t *testing.T) {
	assets := writeTestAssets(t, map[string]string{
		"ok-script.json": scriptPackage("ok-script", "exit 0", "exit 0"),
	})
	installer, err := NewInstaller(assets)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	if err := installer.Install(&InstallOptions{PackageName: "ok-script"}); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

// TestUsesInstallPath covers the target message decision
func TestUsesInstallPath(t *testing.T) {
	if usesInstallPath([]string{`powershell -c "irm x | iex"`}) {
		t.Error("script without ${INSTALL_PATH} reported as using it")
	}
	if !usesInstallPath([]string{"echo a", `mkdir "${INSTALL_PATH}"`}) {
		t.Error("script with ${INSTALL_PATH} not detected")
	}
}
