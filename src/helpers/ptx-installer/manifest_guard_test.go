/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/binary"
	"os"
	"strings"
	"testing"
)

// TestInstallerManifestRequestsAsInvoker is a regression guard for the
// recurring "ptx-installer.exe wants admin rights" problem. Because the binary
// filename contains "installer", Windows' installer-detection heuristic
// auto-elevates it (UAC) unless an explicit asInvoker manifest is embedded via
// ptx-installer_windows_*.syso. Privilege elevation is requested on-demand, per command,
// by the engine (issue #189) — the binary itself must run asInvoker.
//
// This test fails the build if the helper manifest is dropped or flipped to
// requireAdministrator / highestAvailable. The committed ptx-installer_windows_*.syso are
// regenerated from this manifest by build-with-version.sh on every build.
func TestInstallerManifestRequestsAsInvoker(t *testing.T) {
	const manifestPath = "ptx-installer.exe.manifest"

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("cannot read %s: %v", manifestPath, err)
	}
	manifest := string(data)

	if !strings.Contains(manifest, `level="asInvoker"`) {
		t.Errorf("%s must request asInvoker (requestedExecutionLevel level=\"asInvoker\"); "+
			"this is required so ptx-installer.exe never auto-elevates via Windows "+
			"installer-detection", manifestPath)
	}

	for _, forbidden := range []string{"requireAdministrator", "highestAvailable"} {
		if strings.Contains(manifest, forbidden) {
			t.Errorf("%s must NOT request %q — ptx-installer.exe must run asInvoker and "+
				"elevate on-demand per command (issue #189)", manifestPath, forbidden)
		}
	}
}

// TestSysoPerArchitecture is a regression guard for issue #250: every Windows
// resource object must be a COFF object of the architecture named in its file
// name. A single amd64 .syso linked into an arm64 image leaves the resource
// offsets unrelocated, which corrupts the manifest, icon and version info
func TestSysoPerArchitecture(t *testing.T) {
	const name = "ptx-installer"

	if _, err := os.Stat(name + ".syso"); err == nil {
		t.Errorf("%s.syso must not exist: without a _windows_<arch> suffix it is "+
			"linked into every GOARCH (issue #250)", name)
	}

	// COFF machine types: IMAGE_FILE_MACHINE_AMD64 and IMAGE_FILE_MACHINE_ARM64
	for arch, machine := range map[string]uint16{"amd64": 0x8664, "arm64": 0xaa64} {
		path := name + "_windows_" + arch + ".syso"
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("cannot read %s: %v", path, err)
			continue
		}
		if len(data) < 2 {
			t.Errorf("%s is too short to be a COFF object", path)
			continue
		}
		if got := binary.LittleEndian.Uint16(data); got != machine {
			t.Errorf("%s has COFF machine 0x%04x, want 0x%04x (%s); regenerate it "+
				"with make gen-resources or build-with-version.sh", path, got, machine, arch)
		}
	}
}
