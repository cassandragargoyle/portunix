/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"strings"
	"testing"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

func TestUnknownVariantError_ListsAvailableSorted(t *testing.T) {
	variants := map[string]registry.VariantSpec{
		"snap":     {Version: "latest"},
		"apt":      {Version: "latest"},
		"standard": {Version: "0.150.1"},
		"extended": {Version: "0.150.1"},
	}

	err := UnknownVariantError("hugo", "bogus", "linux", variants)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	msg := err.Error()

	// Variant name appears (quoted)
	if !strings.Contains(msg, `"bogus"`) {
		t.Errorf("error should mention requested variant 'bogus': %q", msg)
	}
	// Package and platform context
	if !strings.Contains(msg, "hugo") || !strings.Contains(msg, "linux") {
		t.Errorf("error should include package and OS context: %q", msg)
	}
	// All variants listed in alphabetical order
	wantList := "apt, extended, snap, standard"
	if !strings.Contains(msg, wantList) {
		t.Errorf("error should list available variants in sorted order %q\ngot: %q", wantList, msg)
	}
	// Hint pointing at --list-variants
	if !strings.Contains(msg, "--list-variants") {
		t.Errorf("error should hint at --list-variants: %q", msg)
	}
}

func TestUnknownVariantError_EmptyVariantsMap(t *testing.T) {
	err := UnknownVariantError("foo", "x", "linux", map[string]registry.VariantSpec{})
	if err == nil {
		t.Fatal("expected non-nil error even with empty variants")
	}
	msg := err.Error()
	if !strings.Contains(msg, "available variants:") {
		t.Errorf("error should still include 'available variants:' label: %q", msg)
	}
}

// TestEffectiveInstallArgs verifies that variant-level installArgs take
// precedence over the platform level (issue #202)
func TestEffectiveInstallArgs(t *testing.T) {
	platformArgs := []string{"/quiet", "/norestart"}
	variantArgs := []string{"/quiet", "InstallAllUsers=1", "PrependPath=1", "Include_test=0"}

	tests := []struct {
		name     string
		platform registry.PlatformSpec
		variant  registry.VariantSpec
		want     []string
	}{
		{"variant wins", registry.PlatformSpec{InstallArgs: platformArgs}, registry.VariantSpec{InstallArgs: variantArgs}, variantArgs},
		{"variant only", registry.PlatformSpec{}, registry.VariantSpec{InstallArgs: variantArgs}, variantArgs},
		{"platform fallback", registry.PlatformSpec{InstallArgs: platformArgs}, registry.VariantSpec{}, platformArgs},
		{"none", registry.PlatformSpec{}, registry.VariantSpec{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := effectiveInstallArgs(&tt.platform, &tt.variant)
			if strings.Join(got, " ") != strings.Join(tt.want, " ") {
				t.Errorf("effectiveInstallArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPythonManifest_WindowsFullVariant guards the shipped python.json:
// version 3.14.8, its own installer switches and admin elevation
func TestPythonManifest_WindowsFullVariant(t *testing.T) {
	reg, err := registry.LoadPackageRegistry("../assets")
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}
	pkg, err := reg.GetPackage("python")
	if err != nil {
		t.Fatalf("python package not found: %v", err)
	}
	platform := pkg.Spec.Platforms["windows"]
	full := platform.Variants["full"]

	if full.Version != "3.14.8" {
		t.Errorf("full variant version = %s, want 3.14.8", full.Version)
	}
	if got := strings.Join(effectiveInstallArgs(&platform, &full), " "); got != "/quiet InstallAllUsers=1 PrependPath=1 Include_test=0" {
		t.Errorf("full variant install args = %q", got)
	}
	if !full.RequiresAdmin {
		t.Error("full variant must require admin for InstallAllUsers=1")
	}
	for _, arch := range []string{"x64", "x86", "arm64"} {
		if !strings.Contains(full.URLs[arch], "/3.14.8/") {
			t.Errorf("full variant URL for %s = %q, want a 3.14.8 URL", arch, full.URLs[arch])
		}
	}
}

// TestPythonManifest_WindowsDefaultIsLatest guards that `portunix install
// python` without --variant installs the newest Python on Windows (issue #202)
func TestPythonManifest_WindowsDefaultIsLatest(t *testing.T) {
	reg, err := registry.LoadPackageRegistry("../assets")
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}
	pkg, err := reg.GetPackage("python")
	if err != nil {
		t.Fatalf("python package not found: %v", err)
	}
	platform := pkg.Spec.Platforms["windows"]

	if got := (&Installer{}).autoDetectVariant(&platform); got != "latest" {
		t.Errorf("default Windows variant = %s, want latest", got)
	}
	latest := platform.Variants["latest"]
	if latest.FallbackVariant != "full" {
		t.Errorf("latest variant must fall back to full, got %q", latest.FallbackVariant)
	}
	if latest.VersionResolver != "python.org" || !latest.RequiresAdmin {
		t.Errorf("latest variant must resolve from python.org and require admin: %+v", latest)
	}
}

// TestSelectDefaultVariant_DeclaredOrder verifies that the last resort is the
// first variant the manifest declares, not the alphabetically first one
func TestSelectDefaultVariant_DeclaredOrder(t *testing.T) {
	platform := registry.PlatformSpec{
		Variants: map[string]registry.VariantSpec{
			"stable": {Version: "1"},
			"beta":   {Version: "2"},
			"dev":    {Version: "3"},
		},
		VariantOrder: []string{"stable", "beta", "dev"},
	}
	for run := 0; run < 20; run++ {
		if got := selectDefaultVariant(&platform, "", nil); got != "stable" {
			t.Fatalf("selectDefaultVariant() = %s, want stable", got)
		}
	}
}

// TestSelectDefaultVariant_PackageManagerFirst verifies the package manager
// variant still wins over preferred and declaration order
func TestSelectDefaultVariant_PackageManagerFirst(t *testing.T) {
	platform := registry.PlatformSpec{
		Variants: map[string]registry.VariantSpec{
			"stable": {Version: "1", Preferred: true},
			"apt":    {Version: "latest"},
		},
		VariantOrder: []string{"stable", "apt"},
	}
	if got := selectDefaultVariant(&platform, "apt-get", nil); got != "apt" {
		t.Errorf("selectDefaultVariant() = %s, want apt", got)
	}
	if got := selectDefaultVariant(&platform, "", nil); got != "stable" {
		t.Errorf("selectDefaultVariant() without package manager = %s, want stable", got)
	}
}

// TestDefaultVariants_RealManifests pins the default variant of packages
// whose choice regressed when the fallback was alphabetical (issue #202, D1)
func TestDefaultVariants_RealManifests(t *testing.T) {
	reg, err := registry.LoadPackageRegistry("../assets")
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}

	tests := []struct {
		os, pm, pkg, want string
	}{
		{"windows", "", "chrome", "stable"},
		{"windows", "", "rust", "stable"},
		{"windows", "", "clang", "latest"},
		{"windows", "", "make", "latest"},
		{"windows", "", "ninja", "latest"},
		{"windows", "", "terraform", "latest"},
		{"windows", "", "vscode", "user"},
		{"windows", "", "claude-code", "npm"},
		{"windows", "", "java", "8"},
		{"windows", "", "python", "latest"},
		{"linux", "apt", "chrome", "ubuntu"},
		{"linux", "apt", "vscode", "stable"},
		{"linux", "apt", "terraform", "latest"},
		{"linux", "apt", "java", "8"},
		{"linux", "apt", "claude-code", "npm"},
		{"linux", "apt", "python", "apt"},
	}
	for _, tt := range tests {
		t.Run(tt.os+"/"+tt.pkg, func(t *testing.T) {
			pkg, err := reg.GetPackage(tt.pkg)
			if err != nil {
				t.Fatalf("package not found: %v", err)
			}
			platform, ok := pkg.Spec.Platforms[tt.os]
			if !ok {
				t.Fatalf("package has no %s platform", tt.os)
			}
			if got := selectDefaultVariant(&platform, tt.pm, nil); got != tt.want {
				t.Errorf("default variant = %s, want %s", got, tt.want)
			}
		})
	}
}

// gitLikePlatform mirrors the Windows platform of git.json: package manager
// variants first, then the official installer and its pinned fallback
func gitLikePlatform() registry.PlatformSpec {
	return registry.PlatformSpec{
		Type: "winget",
		Variants: map[string]registry.VariantSpec{
			"winget":           {Type: "winget", Version: "latest"},
			"chocolatey":       {Type: "chocolatey", Version: "latest"},
			"installer":        {Type: "exe", Version: "latest"},
			"installer-pinned": {Type: "exe", Version: "2.56.0.2"},
		},
		VariantOrder: []string{"winget", "chocolatey", "installer", "installer-pinned"},
	}
}

// TestSelectDefaultVariant_WindowsPackageManagers verifies that a package
// installed primarily through a Windows package manager follows the tooling
// available: winget, then chocolatey, then the direct installer (issue #222)
func TestSelectDefaultVariant_WindowsPackageManagers(t *testing.T) {
	tests := []struct {
		name string
		pms  []string
		want string
	}{
		{"winget and chocolatey", []string{"winget", "chocolatey"}, "winget"},
		{"winget only", []string{"winget"}, "winget"},
		{"chocolatey only", []string{"chocolatey"}, "chocolatey"},
		{"neither", nil, "installer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform := gitLikePlatform()
			if got := selectDefaultVariant(&platform, "", tt.pms); got != tt.want {
				t.Errorf("selectDefaultVariant() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestSelectDefaultVariant_WindowsPackageManagerMissingVariant verifies that
// an available package manager without a matching variant is skipped
func TestSelectDefaultVariant_WindowsPackageManagerMissingVariant(t *testing.T) {
	platform := registry.PlatformSpec{
		Type: "chocolatey",
		Variants: map[string]registry.VariantSpec{
			"chocolatey": {Version: "latest"},
			"zip":        {Type: "zip", Version: "1.0"},
		},
		VariantOrder: []string{"chocolatey", "zip"},
	}
	if got := selectDefaultVariant(&platform, "", []string{"winget", "chocolatey"}); got != "chocolatey" {
		t.Errorf("selectDefaultVariant() = %s, want chocolatey", got)
	}
	if got := selectDefaultVariant(&platform, "", []string{"winget"}); got != "zip" {
		t.Errorf("selectDefaultVariant() with winget only = %s, want zip", got)
	}
}

// TestSelectDefaultVariant_DirectDefaultUnaffected verifies that packages
// whose platform default is a direct installer keep it even when a package
// manager is available (issue #222 must not change clang, rust, make, ...)
func TestSelectDefaultVariant_DirectDefaultUnaffected(t *testing.T) {
	platform := registry.PlatformSpec{
		Type: "exe",
		Variants: map[string]registry.VariantSpec{
			"latest":     {Version: "20.1.0"},
			"chocolatey": {Type: "chocolatey", Version: "latest"},
			"winget":     {Type: "winget", Version: "latest"},
		},
		VariantOrder: []string{"latest", "chocolatey", "winget"},
	}
	if got := selectDefaultVariant(&platform, "", []string{"winget", "chocolatey"}); got != "latest" {
		t.Errorf("selectDefaultVariant() = %s, want latest", got)
	}
}

// TestPackageManagerFallback verifies the runtime fallback to the direct
// installer when the package manager of the chosen variant is missing,
// including an explicitly requested variant (issue #222)
func TestPackageManagerFallback(t *testing.T) {
	available := func(binaries ...string) func(string) bool {
		return func(binary string) bool {
			for _, b := range binaries {
				if b == binary {
					return true
				}
			}
			return false
		}
	}

	tests := []struct {
		name         string
		variant      string
		available    func(string) bool
		wantFallback string
		wantBinary   string
	}{
		{"winget missing", "winget", available("choco"), "installer", "winget"},
		{"chocolatey missing", "chocolatey", available(), "installer", "choco"},
		{"winget present", "winget", available("winget"), "", ""},
		{"explicit installer", "installer", available(), "", ""},
		{"explicit pinned installer", "installer-pinned", available(), "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform := gitLikePlatform()
			fallback, binary := packageManagerFallback(&platform, tt.variant, tt.available)
			if fallback != tt.wantFallback || binary != tt.wantBinary {
				t.Errorf("packageManagerFallback() = (%q, %q), want (%q, %q)", fallback, binary, tt.wantFallback, tt.wantBinary)
			}
		})
	}
}

// TestDefaultVariants_GitWindows pins the Windows default of the real git
// manifest for each combination of available package managers (issue #222)
func TestDefaultVariants_GitWindows(t *testing.T) {
	reg, err := registry.LoadPackageRegistry("../assets")
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}
	pkg, err := reg.GetPackage("git")
	if err != nil {
		t.Fatalf("package not found: %v", err)
	}
	platform := pkg.Spec.Platforms["windows"]

	tests := []struct {
		pms  []string
		want string
	}{
		{[]string{"winget", "chocolatey"}, "winget"},
		{[]string{"chocolatey"}, "chocolatey"},
		{nil, "installer"},
	}
	for _, tt := range tests {
		if got := selectDefaultVariant(&platform, "", tt.pms); got != tt.want {
			t.Errorf("git default with %v = %s, want %s", tt.pms, got, tt.want)
		}
	}

	installer := platform.Variants["installer"]
	if installer.VersionResolver != "git-for-windows" || installer.FallbackVariant != "installer-pinned" || !installer.RequiresAdmin {
		t.Errorf("installer variant must resolve from git-for-windows, fall back to installer-pinned and require admin: %+v", installer)
	}
	pinned := platform.Variants["installer-pinned"]
	for _, arch := range []string{"x64", "arm64"} {
		if pinned.URLs[arch] == "" || !strings.HasPrefix(pinned.Checksum[arch], "sha256:") {
			t.Errorf("installer-pinned must declare url and sha256 checksum for %s", arch)
		}
	}

	// Linux variants stay package manager based
	linux := pkg.Spec.Platforms["linux"]
	if got := selectDefaultVariant(&linux, "apt-get", nil); got != "apt" {
		t.Errorf("git linux default with apt = %s, want apt", got)
	}
	if got := selectDefaultVariant(&linux, "dnf", nil); got != "dnf" {
		t.Errorf("git linux default with dnf = %s, want dnf", got)
	}
}
