/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"strings"
	"testing"
)

func TestCheckWindowsAdminRequired(t *testing.T) {
	tests := []struct {
		name        string
		dryRun      bool
		isAdmin     bool
		wantErr     bool
		wantMessage string
	}{
		{"admin + normal run: ok", false, true, false, ""},
		{"admin + dry-run: ok", true, true, false, ""},
		{"non-admin + dry-run: ok", true, false, false, ""},
		{"non-admin + normal run: error", false, false, true, "Administrator privileges"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := checkWindowsAdminRequired(tc.dryRun, tc.isAdmin)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tc.wantMessage) {
					t.Errorf("expected error containing %q, got: %v", tc.wantMessage, err)
				}
				return
			}
			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestDecidePrereqChoice(t *testing.T) {
	tests := []struct {
		name           string
		hasWSL         bool
		hasHyperV      bool
		dryRun         bool
		nonInteractive bool
		userInput      string
		want           prereqChoice
		wantErr        bool
	}{
		{"WSL present: nothing to do", true, false, false, false, "", prereqNone, false},
		{"Hyper-V present: nothing to do", false, true, false, false, "", prereqNone, false},
		{"both present: nothing to do", true, true, false, false, "", prereqNone, false},

		{"dry-run + none: auto WSL", false, false, true, false, "", prereqWSL, false},
		{"--yes + none: auto WSL", false, false, false, true, "", prereqWSL, false},

		{"interactive default (enter): WSL", false, false, false, false, "\n", prereqWSL, false},
		{"interactive 1: WSL", false, false, false, false, "1\n", prereqWSL, false},
		{"interactive 2: Hyper-V", false, false, false, false, "2\n", prereqHyperV, false},
		{"interactive 3: cancel", false, false, false, false, "3\n", prereqCancel, false},
		{"interactive invalid: error", false, false, false, false, "9\n", prereqNone, true},
		{"interactive garbage: error", false, false, false, false, "abc\n", prereqNone, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decidePrereqChoice(tc.hasWSL, tc.hasHyperV, tc.dryRun, tc.nonInteractive, tc.userInput)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (got=%v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

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
		{"non-admin + normal: UAC", false, false, elevationUAC},
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

func TestCheckWSLAdminRequired(t *testing.T) {
	tests := []struct {
		name    string
		dryRun  bool
		isAdmin bool
		wantErr bool
	}{
		{"admin + normal: ok", false, true, false},
		{"admin + dry-run: ok", true, true, false},
		{"non-admin + dry-run: ok", true, false, false},
		{"non-admin + normal: error", false, false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := checkWSLAdminRequired(tc.dryRun, tc.isAdmin)
			if (err != nil) != tc.wantErr {
				t.Errorf("got err=%v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "Administrator") {
				t.Errorf("error should mention Administrator: %v", err)
			}
		})
	}
}
