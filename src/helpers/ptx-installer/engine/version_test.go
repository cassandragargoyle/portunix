/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"strings"
	"testing"
)

// multiVersionPkg defines the same variants ("8"/"17"/"21") on every platform
// so the test is deterministic regardless of the host OS that ResolveVersion
// reads via GetOperatingSystem().
const multiVersionPkg = `{
  "apiVersion": "v1",
  "kind": "Package",
  "metadata": {
    "name": "java-test",
    "displayName": "Java (test)",
    "description": "multi-version package",
    "category": "development/languages"
  },
  "spec": {
    "platforms": {
      "windows": { "type": "msi", "variants": {
        "8":  { "version": "8u462b08",  "url": "https://example.com/jdk8.msi" },
        "17": { "version": "17.0.16_8", "url": "https://example.com/jdk17.msi" },
        "21": { "version": "21.0.8_9",  "url": "https://example.com/jdk21.msi" }
      } },
      "linux": { "type": "tar.gz", "variants": {
        "8":  { "version": "8u462b08",  "url": "https://example.com/jdk8.tar.gz" },
        "17": { "version": "17.0.16_8", "url": "https://example.com/jdk17.tar.gz" },
        "21": { "version": "21.0.8_9",  "url": "https://example.com/jdk21.tar.gz" }
      } },
      "darwin": { "type": "tar.gz", "variants": {
        "8":  { "version": "8u462b08",  "url": "https://example.com/jdk8.tar.gz" },
        "17": { "version": "17.0.16_8", "url": "https://example.com/jdk17.tar.gz" },
        "21": { "version": "21.0.8_9",  "url": "https://example.com/jdk21.tar.gz" }
      } }
    }
  }
}`

func newTestInstaller(t *testing.T, files map[string]string) *Installer {
	t.Helper()
	assets := writeTestAssets(t, files)
	installer, err := NewInstaller(assets)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}
	return installer
}

// TestResolveVersion_ByVariantName verifies a selector that matches a variant
// name (e.g. Java "21") resolves to that variant.
func TestResolveVersion_ByVariantName(t *testing.T) {
	installer := newTestInstaller(t, map[string]string{"java-test.json": multiVersionPkg})

	got, err := installer.ResolveVersion("java-test", "21")
	if err != nil {
		t.Fatalf("ResolveVersion returned error: %v", err)
	}
	if got != "21" {
		t.Errorf("ResolveVersion(\"21\") = %q, want \"21\"", got)
	}
}

// TestResolveVersion_ByVersionField verifies a selector that matches a
// variant's declared Version field (e.g. "17.0.16_8") resolves to that variant.
func TestResolveVersion_ByVersionField(t *testing.T) {
	installer := newTestInstaller(t, map[string]string{"java-test.json": multiVersionPkg})

	got, err := installer.ResolveVersion("java-test", "17.0.16_8")
	if err != nil {
		t.Fatalf("ResolveVersion returned error: %v", err)
	}
	if got != "17" {
		t.Errorf("ResolveVersion(\"17.0.16_8\") = %q, want \"17\"", got)
	}
}

// TestResolveVersion_Unknown verifies an unmatched selector returns an error
// that lists the available versions to guide the user.
func TestResolveVersion_Unknown(t *testing.T) {
	installer := newTestInstaller(t, map[string]string{"java-test.json": multiVersionPkg})

	_, err := installer.ResolveVersion("java-test", "99")
	if err == nil {
		t.Fatal("expected error for unknown version, got nil")
	}
	// The error should enumerate the available variants so the user can retry.
	for _, want := range []string{"8", "17", "21"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should list available version %q, got: %v", want, err)
		}
	}
}

// TestResolveVersion_UnknownPackage verifies a missing package is reported.
func TestResolveVersion_UnknownPackage(t *testing.T) {
	installer := newTestInstaller(t, map[string]string{"java-test.json": multiVersionPkg})

	_, err := installer.ResolveVersion("does-not-exist", "1.0")
	if err == nil {
		t.Fatal("expected error for unknown package, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}
