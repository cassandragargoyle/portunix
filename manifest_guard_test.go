package main

import (
	"os"
	"strings"
	"testing"
)

// TestManifestRequestsAsInvoker is a regression guard for issue #189: the
// embedded Windows manifest (portunix.exe.manifest → portunix.syso) MUST keep
// requesting asInvoker so portunix.exe never triggers UAC globally. Privilege
// elevation is requested on-demand, per command, by ptx-installer instead.
//
// This test fails the build (`make test`) if anyone flips the manifest to
// requireAdministrator / highestAvailable, or drops the explicit asInvoker that
// suppresses Windows' installer-detection auto-elevation heuristic.
func TestManifestRequestsAsInvoker(t *testing.T) {
	const manifestPath = "portunix.exe.manifest"

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("cannot read %s: %v", manifestPath, err)
	}
	manifest := string(data)

	if !strings.Contains(manifest, `level="asInvoker"`) {
		t.Errorf("%s must request asInvoker (requestedExecutionLevel level=\"asInvoker\"); "+
			"this is required so portunix.exe never auto-elevates", manifestPath)
	}

	for _, forbidden := range []string{"requireAdministrator", "highestAvailable"} {
		if strings.Contains(manifest, forbidden) {
			t.Errorf("%s must NOT request %q — portunix.exe must run asInvoker and "+
				"elevate on-demand per command (issue #189)", manifestPath, forbidden)
		}
	}
}
