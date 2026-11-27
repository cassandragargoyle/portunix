/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

// depPackage returns a type "script" package for every platform with the
// given install script, verification command and dependencies
func depPackage(name, installScript, verification string, deps ...string) string {
	quoted := make([]string, len(deps))
	for i, d := range deps {
		quoted[i] = `"` + d + `"`
	}
	platform := `{
        "type": "script",
        "variants": { "latest": { "version": "latest", "installScript": "` + installScript + `" } },
        "verification": { "command": "` + verification + `", "expectedExitCode": 0 }
      }`
	return `{
  "apiVersion": "v1",
  "kind": "Package",
  "metadata": { "name": "` + name + `", "displayName": "Test Dep", "description": "test", "category": "development/tools" },
  "spec": {
    "hasVariants": true,
    "platforms": { "windows": ` + platform + `, "linux": ` + platform + `, "darwin": ` + platform + ` },
    "dependencies": [` + strings.Join(quoted, ", ") + `]
  }
}`
}

// markerCommands returns shell commands that create a marker file and test
// for it, in the default shell of the platform
func markerCommands(t *testing.T) (marker, create, check string) {
	t.Helper()
	marker = filepath.Join(t.TempDir(), "dep-installed")
	if strings.ContainsAny(marker, " \"") {
		t.Skipf("temp path %q needs quoting", marker)
	}
	if runtime.GOOS == "windows" {
		// JSON-escaped backslashes are added by the caller
		return marker, "type nul > " + marker, "if exist " + marker + " (exit 0) else (exit 1)"
	}
	return marker, "touch " + marker, "test -f " + marker
}

// jsonEscape escapes backslashes for embedding into a JSON string literal
func jsonEscape(s string) string {
	return strings.ReplaceAll(s, `\`, `\\`)
}

// TestInstall_MissingDependencyIsInstalledFirst verifies that a dependency
// whose verification fails is installed before the package (issue #203)
func TestInstall_MissingDependencyIsInstalledFirst(t *testing.T) {
	marker, create, check := markerCommands(t)
	assets := writeTestAssets(t, map[string]string{
		"dep-b.json": depPackage("dep-b", jsonEscape(create), jsonEscape(check)),
		// The package itself only succeeds when the dependency ran before it
		"pkg-a.json": depPackage("pkg-a", jsonEscape(check), "exit 0", "dep-b"),
	})
	installer, err := NewInstaller(assets)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	if err := installer.Install(&InstallOptions{PackageName: "pkg-a"}); err != nil {
		t.Fatalf("install with missing dependency failed: %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("dependency was not installed: %v", err)
	}
}

// TestInstall_InstalledDependencyIsSkipped verifies that a dependency whose
// verification succeeds is not installed again
func TestInstall_InstalledDependencyIsSkipped(t *testing.T) {
	assets := writeTestAssets(t, map[string]string{
		// Installing dep-b would fail; it must be skipped
		"dep-b.json": depPackage("dep-b", "exit 1", "exit 0"),
		"pkg-a.json": depPackage("pkg-a", "exit 0", "exit 0", "dep-b"),
	})
	installer, err := NewInstaller(assets)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	if err := installer.Install(&InstallOptions{PackageName: "pkg-a"}); err != nil {
		t.Fatalf("installed dependency should be skipped, got: %v", err)
	}
}

// TestInstall_FailedDependencyStopsInstall verifies that the package is not
// installed when its dependency fails
func TestInstall_FailedDependencyStopsInstall(t *testing.T) {
	marker, create, _ := markerCommands(t)
	assets := writeTestAssets(t, map[string]string{
		"dep-b.json": depPackage("dep-b", "exit 1", "exit 1"),
		"pkg-a.json": depPackage("pkg-a", jsonEscape(create), "exit 0", "dep-b"),
	})
	installer, err := NewInstaller(assets)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	err = installer.Install(&InstallOptions{PackageName: "pkg-a"})
	if err == nil || !strings.Contains(err.Error(), "dependency dep-b") {
		t.Fatalf("expected dependency error, got: %v", err)
	}
	if _, statErr := os.Stat(marker); statErr == nil {
		t.Error("package was installed although its dependency failed")
	}
}

// TestInstallScript_VerificationSeesPathAppend verifies the claude-code curl
// case: a tool installed into the PATH_APPEND directory, which is not yet in
// PATH, passes the verification (issue #203). installScript is called
// directly so the user PATH in the registry is not modified
func TestInstallScript_VerificationSeesPathAppend(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("uses a .cmd probe")
	}
	binDir := filepath.Join(t.TempDir(), "bin")
	if strings.ContainsAny(binDir, " \"") {
		t.Skipf("temp path %q needs quoting", binDir)
	}
	t.Setenv("PTX_TEST_BIN", binDir)
	t.Setenv("PATH", os.Getenv("PATH"))

	platform := &registry.PlatformSpec{
		Type:         "script",
		Verification: &registry.VerificationSpec{Command: "ptxprobe203", ExpectedExitCode: 0},
		Environment:  map[string]string{"PATH_APPEND": "${PTX_TEST_BIN}"},
	}
	variant := &registry.VariantSpec{
		Version: "latest",
		// The closing parenthesis keeps the line from looking like an embedded .cmd path
		InstallScript: registry.StringOrSlice{"mkdir " + binDir + " && (echo @exit 0> " + binDir + `\ptxprobe203.cmd)`},
	}

	i := &Installer{}
	if err := i.installScript(platform, variant, &InstallOptions{}); err != nil {
		t.Fatalf("verification should find the tool in PATH_APPEND: %v", err)
	}
	// The directory is on the process PATH only; Install must still persist it
	if !needsPersisting(binDir) {
		t.Error("PATH_APPEND directory added for the verification must still be written to the user PATH")
	}
}

// TestNeedsPersisting verifies that only directories this process put on PATH
// itself are persisted again; directories the user PATH already has are not
// (issue #203, D1)
func TestNeedsPersisting(t *testing.T) {
	existing := t.TempDir()
	added := t.TempDir()
	missing := t.TempDir()
	t.Setenv("PATH", existing)

	prependProcessPath(added)
	if !IsInUserPath(added) {
		t.Fatal("prependProcessPath did not update the process PATH")
	}

	cases := map[string]bool{existing: false, added: true, missing: true}
	for dir, want := range cases {
		if got := needsPersisting(dir); got != want {
			t.Errorf("needsPersisting(%q) = %v, want %v", dir, got, want)
		}
	}
}

func TestResolvePathAppend(t *testing.T) {
	t.Setenv("PTX_TEST_DIR", "/opt/x")
	cases := map[string]string{
		"":                         "",
		"${PTX_TEST_DIR}/bin":      "/opt/x/bin",
		"${install_path}/bin":      "/inst/bin",
		"${PTX_UNSET_VAR_203}/bin": "",
	}
	for in, want := range cases {
		if got := resolvePathAppend(in, "/inst"); got != want {
			t.Errorf("resolvePathAppend(%q) = %q, want %q", in, got, want)
		}
	}
}
