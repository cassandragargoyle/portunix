/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestAssets creates a temporary assets/packages directory populated with
// the given filename->content JSON files and returns the assets root.
func writeTestAssets(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	pkgDir := filepath.Join(root, "packages")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("failed to create packages dir: %v", err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(pkgDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return root
}

// TestInstallBundle_MissingMemberFails verifies that installing a bundle whose
// member package does not exist stops with an error naming the missing member,
// even in dry-run mode.
func TestInstallBundle_MissingMemberFails(t *testing.T) {
	bundle := `{
  "apiVersion": "v1",
  "kind": "Bundle",
  "metadata": {
    "name": "test-bundle",
    "displayName": "Test Bundle",
    "description": "bundle with a missing member",
    "category": "development/ai-tools"
  },
  "spec": { "bundle": ["does-not-exist"] }
}`
	assets := writeTestAssets(t, map[string]string{"test-bundle.json": bundle})

	installer, err := NewInstaller(assets)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	err = installer.Install(&InstallOptions{PackageName: "test-bundle", DryRun: true})
	if err == nil {
		t.Fatal("expected error when bundle member is missing, got nil")
	}
	if !strings.Contains(err.Error(), "does-not-exist") {
		t.Errorf("error should name the missing member 'does-not-exist', got: %v", err)
	}
}
