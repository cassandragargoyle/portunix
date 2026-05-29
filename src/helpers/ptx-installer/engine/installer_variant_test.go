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
