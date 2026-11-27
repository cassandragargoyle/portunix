/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

// requirePowerShell skips the test when the PowerShell executable is missing
func requirePowerShell(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath(powerShellExecutable()); err != nil {
		t.Skipf("%s not available", powerShellExecutable())
	}
}

// runScript runs script with the given shell and returns its trimmed output
func runScript(t *testing.T, shell, script string) (string, error) {
	t.Helper()
	cmd, cleanup, err := scriptCommand(shell, script)
	if err != nil {
		t.Fatalf("scriptCommand(%q) failed: %v", shell, err)
	}
	defer cleanup()
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// TestScriptCommand_PowerShellRunsVariablesAndQuotes verifies the
// claude-desktop pattern: variables, cmdlets and nested quotes reach
// PowerShell unchanged (issue #203)
func TestScriptCommand_PowerShellRunsVariablesAndQuotes(t *testing.T) {
	requirePowerShell(t)

	script := `$name = 'ptx'; $output = Join-Path 'a' 'b'; Write-Output ("$name-ok " + '"quoted"' + " $output")`
	out, err := runScript(t, registry.ScriptShellPowerShell, script)
	if err != nil {
		t.Fatalf("PowerShell script failed: %v\n%s", err, out)
	}
	sep := `\`
	if runtime.GOOS != "windows" {
		sep = "/"
	}
	want := `ptx-ok "quoted" a` + sep + "b"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

// TestScriptCommand_PowerShellCmdletErrorFails verifies that a failing cmdlet
// stops the script with a non-zero exit code instead of a false success
func TestScriptCommand_PowerShellCmdletErrorFails(t *testing.T) {
	requirePowerShell(t)

	out, err := runScript(t, registry.ScriptShellPowerShell,
		"Get-Item 'ptx-203-does-not-exist'; Write-Output ('should-' + 'not-run')")
	if err == nil {
		t.Fatalf("expected error for a failing cmdlet, got nil\n%s", out)
	}
	// The error message quotes the script line, so look for the evaluated text
	if strings.Contains(out, "should-not-run") {
		t.Errorf("script continued after the failing cmdlet: %s", out)
	}
}

// TestScriptCommand_PowerShellNativeExitCodePropagates verifies that a
// failing native command (e.g. npm) makes the script fail
func TestScriptCommand_PowerShellNativeExitCodePropagates(t *testing.T) {
	requirePowerShell(t)

	native := "cmd /c exit 4"
	if runtime.GOOS != "windows" {
		native = "sh -c 'exit 4'"
	}
	out, err := runScript(t, registry.ScriptShellPowerShell, native+"; Write-Output 'after'")
	if err == nil {
		t.Fatalf("expected error for a failing native command, got nil\n%s", out)
	}
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 4 {
		t.Errorf("expected exit code 4, got %v", err)
	}
}

// TestScriptCommand_DefaultShell verifies the platform default is unchanged
func TestScriptCommand_DefaultShell(t *testing.T) {
	out, err := runScript(t, registry.ScriptShellDefault, "echo ptx-default")
	if err != nil {
		t.Fatalf("default shell failed: %v\n%s", err, out)
	}
	if out != "ptx-default" {
		t.Errorf("output = %q, want %q", out, "ptx-default")
	}
}

// TestScriptCommand_CmdOnlyOnWindows verifies cmd is rejected elsewhere
func TestScriptCommand_CmdOnlyOnWindows(t *testing.T) {
	_, cleanup, err := scriptCommand(registry.ScriptShellCmd, "echo x")
	defer cleanup()
	if runtime.GOOS == "windows" && err != nil {
		t.Errorf("cmd should be accepted on Windows: %v", err)
	}
	if runtime.GOOS != "windows" && err == nil {
		t.Error("cmd should be rejected outside Windows")
	}
}

// TestScriptCommand_UnknownShellFails verifies an unknown interpreter is rejected
func TestScriptCommand_UnknownShellFails(t *testing.T) {
	_, cleanup, err := scriptCommand("bash", "echo x")
	defer cleanup()
	if err == nil {
		t.Error("expected error for unknown shell, got nil")
	}
}

// shellPackage returns a Windows-only type "script" package whose variant
// declares the given scriptShell
func shellPackage(name, shell, installScript string) string {
	shellField := ""
	if shell != "" {
		shellField = `"scriptShell": "` + shell + `", `
	}
	return `{
  "apiVersion": "v1",
  "kind": "Package",
  "metadata": { "name": "` + name + `", "displayName": "Test Shell", "description": "test", "category": "development/tools" },
  "spec": {
    "hasVariants": true,
    "platforms": { "windows": {
      "type": "script",
      "variants": { "latest": { "version": "latest", ` + shellField + `"installScript": "` + installScript + `" } },
      "verification": { "command": "exit 0", "expectedExitCode": 0 }
    } }
  }
}`
}

// TestInstallScript_PowerShellPackage verifies the claude-desktop case end to
// end: a PowerShell installScript runs when the variant declares
// scriptShell powershell, and fails under cmd.exe without it (issue #203)
func TestInstallScript_PowerShellPackage(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only package")
	}
	script := `$x = Join-Path 'a' 'b'; if (-not $x) { exit 1 }`

	assets := writeTestAssets(t, map[string]string{
		"ps-script.json":  shellPackage("ps-script", registry.ScriptShellPowerShell, script),
		"cmd-script.json": shellPackage("cmd-script", "", script),
	})
	installer, err := NewInstaller(assets)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	if err := installer.Install(&InstallOptions{PackageName: "ps-script"}); err != nil {
		t.Errorf("PowerShell package failed: %v", err)
	}
	if err := installer.Install(&InstallOptions{PackageName: "cmd-script"}); err == nil {
		t.Error("PowerShell script without scriptShell should fail under cmd.exe")
	}
}
