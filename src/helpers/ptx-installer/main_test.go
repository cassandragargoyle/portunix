/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"strings"
	"testing"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

func TestParseVariantArg_VariantAndMethodEquivalent(t *testing.T) {
	// All four spellings must yield the same Variant value. This locks in
	// the issue #079 contract: --method is a true alias of --variant.
	cases := []struct {
		name         string
		args         []string
		startIndex   int
		wantValue    string
		wantConsumed int
		wantOK       bool
	}{
		{"variant equals", []string{"hugo", "--variant=snap"}, 1, "snap", 0, true},
		{"variant space", []string{"hugo", "--variant", "snap"}, 1, "snap", 1, true},
		{"method equals", []string{"hugo", "--method=snap"}, 1, "snap", 0, true},
		{"method space", []string{"hugo", "--method", "snap"}, 1, "snap", 1, true},
		{"unrelated flag", []string{"hugo", "--dry-run"}, 1, "", 0, false},
		{"variant missing value at end", []string{"hugo", "--variant"}, 1, "", 0, false},
		{"method missing value at end", []string{"hugo", "--method"}, 1, "", 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			value, consumed, ok := parseVariantArg(tc.args, tc.startIndex)
			if ok != tc.wantOK {
				t.Errorf("ok = %v, want %v", ok, tc.wantOK)
			}
			if value != tc.wantValue {
				t.Errorf("value = %q, want %q", value, tc.wantValue)
			}
			if consumed != tc.wantConsumed {
				t.Errorf("consumed = %d, want %d", consumed, tc.wantConsumed)
			}
		})
	}
}

func TestFormatVariantList_RendersAllFields(t *testing.T) {
	platform := &registry.PlatformSpec{
		Type: "tar.gz",
		Variants: map[string]registry.VariantSpec{
			"standard": {Version: "0.150.1", Description: "Plain Hugo"},
			"extended": {Version: "0.150.1", Description: "Hugo with Sass/SCSS", Preferred: true},
			"apt":      {Version: "latest", Type: "apt"},
		},
	}

	out := formatVariantList("hugo", "linux", "apt", platform)

	// Header includes package and platform
	if !strings.Contains(out, "hugo") || !strings.Contains(out, "linux") {
		t.Errorf("output should include package + platform: %s", out)
	}
	// All three variants must appear
	for _, v := range []string{"standard", "extended", "apt"} {
		if !strings.Contains(out, v) {
			t.Errorf("variant %q missing from output:\n%s", v, out)
		}
	}
	// Auto-detected marker on apt — line containing "apt" must start with the "*"
	if !strings.Contains(out, "* apt") {
		t.Errorf("auto-detected variant 'apt' should be marked with '*':\n%s", out)
	}
	// Preferred tag on extended
	if !strings.Contains(out, "extended") || !strings.Contains(out, "(preferred)") {
		t.Errorf("preferred tag missing for 'extended':\n%s", out)
	}
	// Description rendered when present
	if !strings.Contains(out, "Hugo with Sass/SCSS") {
		t.Errorf("description for 'extended' missing:\n%s", out)
	}
	// Variant-level type override (apt) wins over platform default (tar.gz)
	// Find the line containing "apt" with apt type, not tar.gz
	aptLine := ""
	for _, ln := range strings.Split(out, "\n") {
		if strings.Contains(ln, " apt ") || strings.Contains(ln, " apt  ") {
			aptLine = ln
			break
		}
	}
	if aptLine != "" && !strings.Contains(aptLine, "type: apt") {
		t.Errorf("apt variant should report effective type 'apt' (variant override), got line: %q", aptLine)
	}
	// Install hint references both --method and --variant aliases
	if !strings.Contains(out, "--method=") || !strings.Contains(out, "--variant=") {
		t.Errorf("install hint should mention both --method and --variant:\n%s", out)
	}
}

func TestFormatVariantList_StableOrdering(t *testing.T) {
	// Map iteration is randomized in Go; confirm we sort variant names.
	platform := &registry.PlatformSpec{
		Type: "script",
		Variants: map[string]registry.VariantSpec{
			"zulu":  {Version: "1"},
			"alpha": {Version: "1"},
			"mango": {Version: "1"},
		},
	}

	out := formatVariantList("pkg", "linux", "", platform)
	idxAlpha := strings.Index(out, "alpha")
	idxMango := strings.Index(out, "mango")
	idxZulu := strings.Index(out, "zulu")

	if !(idxAlpha < idxMango && idxMango < idxZulu) {
		t.Errorf("variants not in alphabetical order: alpha=%d mango=%d zulu=%d\n%s", idxAlpha, idxMango, idxZulu, out)
	}
}
