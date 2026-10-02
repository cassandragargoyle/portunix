/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"errors"
	"testing"
)

// TestDecideElevation exercises the shared elevation decision used by BOTH the
// Docker installer and the generic package installer (issue #189). The matrix
// is identical for both call sites — dry-run and already-elevated proceed
// in-process; only a real, non-elevated install re-launches under UAC/sudo.
func TestDecideElevation(t *testing.T) {
	tests := []struct {
		name    string
		dryRun  bool
		isAdmin bool
		want    elevationAction
	}{
		{"admin + normal: direct (no UAC)", false, true, elevationDirect},
		{"admin + dry-run: direct", true, true, elevationDirect},
		{"non-admin + dry-run: direct (skip launch)", true, false, elevationDirect},
		{"non-admin + normal: UAC/sudo re-launch", false, false, elevationUAC},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := decideElevation(tc.dryRun, tc.isAdmin)
			if got != tc.want {
				t.Errorf("decideElevation(dryRun=%v, isAdmin=%v) = %v, want %v",
					tc.dryRun, tc.isAdmin, got, tc.want)
			}
		})
	}
}

// TestElevationSentinelsDistinct guards the actionable-fallback branches in
// elevateInstall: errUACDeclined (Windows, user clicked No) and errSudoMissing
// (Linux/macOS, no sudo) must be non-nil and distinguishable so each maps to
// its own user message.
func TestElevationSentinelsDistinct(t *testing.T) {
	if errUACDeclined == nil || errSudoMissing == nil {
		t.Fatal("elevation sentinel errors must be non-nil")
	}
	if errors.Is(errUACDeclined, errSudoMissing) || errors.Is(errSudoMissing, errUACDeclined) {
		t.Error("errUACDeclined and errSudoMissing must be distinct sentinels")
	}
}
