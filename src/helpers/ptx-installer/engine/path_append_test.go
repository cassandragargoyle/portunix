/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"testing"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

// TestEffectivePathAppend_PerInstallType checks that the platform PATH_APPEND
// is applied for archive/script installs and skipped for package managers (issue #253)
func TestEffectivePathAppend_PerInstallType(t *testing.T) {
	platform := &registry.PlatformSpec{
		Environment: map[string]string{"PATH_APPEND": "/opt/tool/bin"},
	}
	variant := &registry.VariantSpec{}

	applied := []string{"tar.gz", "zip", "download", "script", "msi", "exe", "chocolatey", "winget"}
	for _, installType := range applied {
		if got := effectivePathAppend(platform, variant, installType); got != "/opt/tool/bin" {
			t.Errorf("type %s: expected platform PATH_APPEND, got %q", installType, got)
		}
	}

	skipped := []string{"deb", "apt", "dnf", "yum", "snap", "pacman"}
	for _, installType := range skipped {
		if got := effectivePathAppend(platform, variant, installType); got != "" {
			t.Errorf("type %s: expected PATH_APPEND to be skipped, got %q", installType, got)
		}
	}
}

// TestEffectivePathAppend_VariantOverride checks that a variant-level
// environment wins over the platform value for every install type
func TestEffectivePathAppend_VariantOverride(t *testing.T) {
	platform := &registry.PlatformSpec{
		Environment: map[string]string{"PATH_APPEND": "/opt/tool/bin"},
	}

	custom := &registry.VariantSpec{Environment: map[string]string{"PATH_APPEND": "/usr/lib/tool/bin"}}
	for _, installType := range []string{"apt", "tar.gz"} {
		if got := effectivePathAppend(platform, custom, installType); got != "/usr/lib/tool/bin" {
			t.Errorf("type %s: expected variant PATH_APPEND, got %q", installType, got)
		}
	}

	disabled := &registry.VariantSpec{Environment: map[string]string{"PATH_APPEND": ""}}
	if got := effectivePathAppend(platform, disabled, "tar.gz"); got != "" {
		t.Errorf("empty variant PATH_APPEND must disable PATH wiring, got %q", got)
	}

	otherKeys := &registry.VariantSpec{Environment: map[string]string{"JAVA_HOME": "/opt/java"}}
	if got := effectivePathAppend(platform, otherKeys, "tar.gz"); got != "/opt/tool/bin" {
		t.Errorf("variant without PATH_APPEND must fall back to platform, got %q", got)
	}
	if got := effectivePathAppend(platform, otherKeys, "apt"); got != "" {
		t.Errorf("variant without PATH_APPEND must not re-enable it for apt, got %q", got)
	}
}

// TestEffectivePathAppend_Manifests checks the affected packages from the
// shipped registry: package-manager variants get no PATH entry, archive and
// script variants keep theirs
func TestEffectivePathAppend_Manifests(t *testing.T) {
	reg, err := registry.LoadPackageRegistry("../assets")
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}

	cases := []struct {
		pkg, variant, want string
	}{
		{"maven", "apt", ""},
		{"maven", "latest", "/opt/maven/bin"},
		{"nodejs", "apt", ""},
		{"nodejs", "dnf", ""},
		{"nodejs", "pacman", ""},
		{"rust", "apt", ""},
		{"rust", "dnf", ""},
		{"rust", "pacman", ""},
		{"rust", "stable", "$HOME/.cargo/bin"},
	}

	for _, c := range cases {
		pkg, err := reg.GetPackage(c.pkg)
		if err != nil {
			t.Fatalf("package %s not found: %v", c.pkg, err)
		}
		platform, ok := pkg.Spec.Platforms["linux"]
		if !ok {
			t.Fatalf("package %s has no linux platform", c.pkg)
		}
		variant, ok := platform.Variants[c.variant]
		if !ok {
			t.Fatalf("package %s has no variant %s", c.pkg, c.variant)
		}
		installType := platform.Type
		if variant.Type != "" {
			installType = variant.Type
		}
		if got := effectivePathAppend(&platform, &variant, installType); got != c.want {
			t.Errorf("%s/%s (%s): expected %q, got %q", c.pkg, c.variant, installType, c.want, got)
		}
	}
}
