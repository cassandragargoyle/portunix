/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// powerShellSyntax matches constructs that only PowerShell understands:
// variable assignment, $env:, subexpressions, Verb-Noun cmdlets and the
// iwr/irm/iex aliases. cmd.exe stops at the first of them (issue #203)
var powerShellSyntax = regexp.MustCompile(
	`\$[A-Za-z_][A-Za-z0-9_]*\s*=|\$env:|\$\(|` +
		`\b(Write|Invoke|Start|Remove|Join|Add|Install|Get|Set|New|Test|Repair|Out|Expand)-[A-Z][A-Za-z]+|` +
		`(^|[\s;|])(iwr|irm|iex)(\s|$)`)

// usesPowerShellSyntax reports whether an inline script is PowerShell code.
// A cmd line that launches powershell itself (e.g. the uv installer) is not
func usesPowerShellSyntax(script string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(script))
	if strings.HasPrefix(trimmed, "powershell") || strings.HasPrefix(trimmed, "pwsh") {
		return false
	}
	return powerShellSyntax.MatchString(script)
}

// isEmbeddedScriptPath mirrors engine.isEmbeddedScript: a path to a bundled
// .ps1/.cmd/.sh file picks its interpreter from the extension
func isEmbeddedScriptPath(script string) bool {
	lower := strings.ToLower(script)
	if strings.HasSuffix(lower, ".ps1") || strings.HasSuffix(lower, ".cmd") || strings.HasSuffix(lower, ".sh") {
		return strings.Contains(script, "/") || strings.Contains(script, "\\")
	}
	return false
}

func TestUsesPowerShellSyntax(t *testing.T) {
	cases := map[string]bool{
		"$url = 'https://example.com'; Invoke-WebRequest -Uri $url":   true,
		"Write-Host 'Installing...'; npm install -g pkg":              true,
		"Add-AppxPackage -RegisterByFamilyName -MainPackage X":        true,
		"irm https://example.com/install.ps1 | iex":                   true,
		"$output = Join-Path $env:TEMP 'x.exe'":                       true,
		`powershell -ExecutionPolicy ByPass -c "irm https://x | iex"`: false,
		"npm install -g pkg": false,
		"curl -fL -o \"$TMP\" https://x && chmod +x \"$TMP\"":            false,
		"winget install --id Some.Package -e --accept-source-agreements": false,
		"echo %USERPROFILE%": false,
	}
	for script, want := range cases {
		if got := usesPowerShellSyntax(script); got != want {
			t.Errorf("usesPowerShellSyntax(%q) = %v, want %v", script, got, want)
		}
	}
}

// TestPackages_WindowsPowerShellScriptsDeclareShell fails when a Windows
// inline installScript is PowerShell code but would be handed to cmd.exe
// because the variant does not set "scriptShell": "powershell" (issue #203)
func TestPackages_WindowsPowerShellScriptsDeclareShell(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "assets", "packages", "*.json"))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no package manifests found in ../assets/packages")
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		var pkg Package
		if err := json.Unmarshal(data, &pkg); err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}

		windows, ok := pkg.Spec.Platforms["windows"]
		if !ok {
			continue
		}
		for name, variant := range windows.Variants {
			if variant.ScriptShell == ScriptShellPowerShell {
				continue
			}
			for _, script := range variant.InstallScript {
				if isEmbeddedScriptPath(script) || !usesPowerShellSyntax(script) {
					continue
				}
				t.Errorf("%s: windows variant %q has a PowerShell installScript without \"scriptShell\": \"powershell\"",
					filepath.Base(file), name)
				break
			}
		}
	}
}

// TestValidateVariant_ScriptShell verifies only known interpreters are accepted
func TestValidateVariant_ScriptShell(t *testing.T) {
	r := &PackageRegistry{}
	for _, shell := range ValidScriptShells {
		v := &VariantSpec{Version: "latest", InstallScript: StringOrSlice{"x"}, ScriptShell: shell}
		if err := r.validateVariant("v", v); err != nil {
			t.Errorf("scriptShell %q rejected: %v", shell, err)
		}
	}
	v := &VariantSpec{Version: "latest", InstallScript: StringOrSlice{"x"}, ScriptShell: "bash"}
	if err := r.validateVariant("v", v); err == nil {
		t.Error("scriptShell \"bash\" should be rejected")
	}
}
