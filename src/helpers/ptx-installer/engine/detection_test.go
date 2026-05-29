/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"testing"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

// pkg is a small helper to build a registry.Package for selection tests
func pkg(name, kind, category string) *registry.Package {
	return &registry.Package{
		Kind: kind,
		Metadata: registry.Metadata{
			Name:        name,
			DisplayName: name,
			Category:    category,
		},
	}
}

// TestSelectAIAssistants_FiltersBundlesAndCategory verifies that only
// Kind:"Package" entries in the AI tools category are selected, that bundles
// and unrelated packages are excluded, and that the result is sorted by name.
func TestSelectAIAssistants_FiltersBundlesAndCategory(t *testing.T) {
	pkgs := map[string]*registry.Package{
		"gemini-cli":         pkg("gemini-cli", "Package", aiAssistantCategory),
		"claude-code":        pkg("claude-code", "Package", aiAssistantCategory),
		"ai-assistant-basic": pkg("ai-assistant-basic", "Bundle", aiAssistantCategory), // excluded: bundle
		"python":             pkg("python", "Package", "development/languages"),        // excluded: category
	}

	got := selectAIAssistants(pkgs)

	if len(got) != 2 {
		t.Fatalf("expected 2 AI assistant packages, got %d", len(got))
	}
	// Sorted by name: claude-code before gemini-cli
	if got[0].Metadata.Name != "claude-code" || got[1].Metadata.Name != "gemini-cli" {
		t.Errorf("unexpected selection/order: %s, %s", got[0].Metadata.Name, got[1].Metadata.Name)
	}
}

// TestSelectAIAssistants_Empty verifies graceful handling of a registry with no
// matching packages.
func TestSelectAIAssistants_Empty(t *testing.T) {
	got := selectAIAssistants(map[string]*registry.Package{
		"python": pkg("python", "Package", "development/languages"),
	})
	if len(got) != 0 {
		t.Errorf("expected no AI assistants, got %d", len(got))
	}
}

// TestDetectAIAssistants_EndToEnd loads a registry from temp assets containing
// two AI assistant packages and one bundle, then verifies DetectAIAssistants
// returns one status per Package (bundle excluded), in name order, with the
// platform verification command populated. Install state is intentionally not
// asserted (it is environment-dependent).
func TestDetectAIAssistants_EndToEnd(t *testing.T) {
	claude := `{
  "apiVersion": "v1",
  "kind": "Package",
  "metadata": {
    "name": "claude-code",
    "displayName": "Claude Code",
    "description": "Anthropic CLI",
    "category": "development/ai-tools"
  },
  "spec": {
    "platforms": {
      "linux":   { "type": "script", "variants": { "npm": { "version": "latest", "installScript": "x" } }, "verification": { "command": "claude --version", "expectedExitCode": 0 } },
      "windows": { "type": "script", "variants": { "npm": { "version": "latest", "installScript": "x" } }, "verification": { "command": "claude --version", "expectedExitCode": 0 } }
    }
  }
}`
	gemini := `{
  "apiVersion": "v1",
  "kind": "Package",
  "metadata": {
    "name": "gemini-cli",
    "displayName": "Google Gemini CLI",
    "description": "Google CLI",
    "category": "development/ai-tools"
  },
  "spec": {
    "platforms": {
      "linux":   { "type": "npm", "variants": { "latest": { "version": "latest", "packages": ["@google/gemini-cli"] } }, "verification": { "command": "gemini --version", "expectedExitCode": 0 } },
      "windows": { "type": "npm", "variants": { "latest": { "version": "latest", "packages": ["@google/gemini-cli"] } }, "verification": { "command": "gemini --version", "expectedExitCode": 0 } }
    }
  }
}`
	bundle := `{
  "apiVersion": "v1",
  "kind": "Bundle",
  "metadata": {
    "name": "ai-assistant-basic",
    "displayName": "AI Assistant (basic)",
    "description": "bundle",
    "category": "development/ai-tools"
  },
  "spec": { "bundle": ["claude-code", "gemini-cli"] }
}`

	assets := writeTestAssets(t, map[string]string{
		"claude-code.json":        claude,
		"gemini-cli.json":         gemini,
		"ai-assistant-basic.json": bundle,
	})

	reg, err := registry.LoadPackageRegistry(assets)
	if err != nil {
		t.Fatalf("LoadPackageRegistry failed: %v", err)
	}

	for _, osName := range []string{"linux", "windows"} {
		got := DetectAIAssistants(reg, osName)
		if len(got) != 2 {
			t.Fatalf("os=%s: expected 2 statuses (bundle excluded), got %d", osName, len(got))
		}
		if got[0].Name != "claude-code" || got[1].Name != "gemini-cli" {
			t.Fatalf("os=%s: unexpected order: %s, %s", osName, got[0].Name, got[1].Name)
		}
		if got[0].VerifyCommand != "claude --version" {
			t.Errorf("os=%s: claude-code verify command = %q", osName, got[0].VerifyCommand)
		}
		if got[1].DisplayName != "Google Gemini CLI" {
			t.Errorf("os=%s: gemini display name = %q", osName, got[1].DisplayName)
		}
	}
}

func TestParseVersionOutput(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"semver in claude output", "1.2.3 (Claude Code)", "1.2.3"},
		{"two-part version", "gemini version 0.4", "0.4"},
		{"version with prefix text", "node v20.11.1 ok", "20.11.1"},
		{"no version (path lookup)", `C:\Users\me\claude-desktop.exe`, ""},
		{"empty", "", ""},
		{"whitespace only", "   \n  ", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseVersionOutput(tc.in); got != tc.want {
				t.Errorf("parseVersionOutput(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestIsPathLookup(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"where claude-desktop", true},
		{"which claude-desktop", true},
		{"  WHERE claude ", true},
		{"claude --version", false},
		{"gemini --version", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isPathLookup(tc.in); got != tc.want {
			t.Errorf("isPathLookup(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// TestVerifyCommandForOS covers platform selection, the windows_sandbox ->
// windows fallback, and the missing-platform / missing-verification cases.
func TestVerifyCommandForOS(t *testing.T) {
	p := &registry.Package{
		Kind: "Package",
		Spec: registry.PackageSpec{
			Platforms: map[string]registry.PlatformSpec{
				"windows": {
					Type:         "script",
					Verification: &registry.VerificationSpec{Command: "claude --version"},
				},
				"linux": {
					Type: "script",
					// no verification spec
				},
			},
		},
	}

	if got := verifyCommandForOS(p, "windows"); got != "claude --version" {
		t.Errorf("windows: got %q", got)
	}
	// windows_sandbox falls back to windows
	if got := verifyCommandForOS(p, "windows_sandbox"); got != "claude --version" {
		t.Errorf("windows_sandbox fallback: got %q", got)
	}
	// linux platform exists but has no verification command
	if got := verifyCommandForOS(p, "linux"); got != "" {
		t.Errorf("linux (no verification): got %q, want empty", got)
	}
	// unknown platform
	if got := verifyCommandForOS(p, "darwin"); got != "" {
		t.Errorf("darwin (missing): got %q, want empty", got)
	}
}
