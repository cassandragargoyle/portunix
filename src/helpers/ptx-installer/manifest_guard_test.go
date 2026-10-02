/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"os"
	"strings"
	"testing"
)

// TestInstallerManifestRequestsAsInvoker is a regression guard for the
// recurring "ptx-installer.exe wants admin rights" problem. Because the binary
// filename contains "installer", Windows' installer-detection heuristic
// auto-elevates it (UAC) unless an explicit asInvoker manifest is embedded via
// ptx-installer.syso. Privilege elevation is requested on-demand, per command,
// by the engine (issue #189) — the binary itself must run asInvoker.
//
// This test fails the build if the helper manifest is dropped or flipped to
// requireAdministrator / highestAvailable. The committed ptx-installer.syso is
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
