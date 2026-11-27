/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"strings"
	"testing"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

// TestWingetInstallArgs verifies that every winget install names its source,
// so an unavailable msstore source cannot fail a community package (issue #203)
func TestWingetInstallArgs(t *testing.T) {
	cases := map[string]string{
		"Anthropic.Claude":                "--source winget",
		"SoftwareFreedomConservancy.QEMU": "--source winget",
		"Git.Git":                         "--source winget",
		"9NBLGGH4NNS1":                    "--source msstore",
		"9nblggh4nns1":                    "--source winget",
	}
	for pkg, wantSource := range cases {
		args := strings.Join(wingetInstallArgs(pkg), " ")
		if !strings.Contains(args, "--id "+pkg+" -e "+wantSource) {
			t.Errorf("wingetInstallArgs(%q) = %q, want %q", pkg, args, wantSource)
		}
		for _, want := range []string{"--accept-package-agreements", "--accept-source-agreements"} {
			if !strings.Contains(args, want) {
				t.Errorf("wingetInstallArgs(%q) misses %s", pkg, want)
			}
		}
	}
}

// TestDefaultVariant_WingetWindows pins the default variant of the winget
// manifest: the pinned msixbundle variant misses framework dependencies on a
// system without winget (issue #203, D3)
func TestDefaultVariant_WingetWindows(t *testing.T) {
	reg, err := registry.LoadPackageRegistry("../assets")
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}
	pkg, err := reg.GetPackage("winget")
	if err != nil {
		t.Fatalf("package not found: %v", err)
	}
	platform := pkg.Spec.Platforms["windows"]
	for _, pms := range [][]string{nil, {"chocolatey"}} {
		if got := selectDefaultVariant(&platform, "", pms); got != "powershell-module" {
			t.Errorf("winget default with %v = %s, want powershell-module", pms, got)
		}
	}
}
