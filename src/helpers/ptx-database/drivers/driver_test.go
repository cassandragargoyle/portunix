/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package drivers

import (
	"strings"
	"testing"
)

func TestRegistry_KnownEngines(t *testing.T) {
	engines := List()
	if len(engines) < 2 {
		t.Fatalf("expected at least postgresql + sqlite, got %v", engines)
	}
	want := map[string]bool{"postgresql": false, "sqlite": false}
	for _, e := range engines {
		if _, ok := want[e]; ok {
			want[e] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("driver %q not registered", name)
		}
	}
}

func TestGet_UnknownEngine(t *testing.T) {
	_, err := Get("does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown engine")
	}
	if !strings.Contains(err.Error(), "unknown engine") {
		t.Errorf("error did not mention 'unknown engine': %v", err)
	}
}

func TestPostgresql_SupportedModes(t *testing.T) {
	d, err := Get("postgresql")
	if err != nil {
		t.Fatal(err)
	}
	modes := d.SupportedModes()
	if len(modes) == 0 {
		t.Fatal("postgresql driver returned no supported modes")
	}
	hasNative, hasContainer := false, false
	for _, m := range modes {
		if m == ModeNative {
			hasNative = true
		}
		if m == ModeContainer {
			hasContainer = true
		}
	}
	if !hasNative || !hasContainer {
		t.Errorf("expected postgresql to support native+container, got %v", modes)
	}
}

func TestSqlite_EmbeddedOnly(t *testing.T) {
	d, err := Get("sqlite")
	if err != nil {
		t.Fatal(err)
	}
	modes := d.SupportedModes()
	if len(modes) != 1 || modes[0] != ModeEmbedded {
		t.Errorf("sqlite driver should be embedded-only, got %v", modes)
	}
	// Lifecycle ops on embedded engines must be no-ops returning a clear error.
	if err := d.Start(LifecycleOpts{}); err == nil {
		t.Error("expected sqlite Start() to return error (embedded)")
	}
}

func TestUtil_NonEmptyLines(t *testing.T) {
	got := nonEmptyLines("a\n\nb\r\n  \nc\n")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("nonEmptyLines: got %v, want %v", got, want)
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("nonEmptyLines[%d] = %q, want %q", i, got[i], v)
		}
	}
}

func TestUtil_SplitPipe(t *testing.T) {
	got := splitPipe("a | b|c")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("splitPipe: got %v, want %v", got, want)
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("splitPipe[%d] = %q, want %q", i, got[i], v)
		}
	}
}

func TestUtil_ShellQuote(t *testing.T) {
	cases := map[string]string{
		"hello":     "'hello'",
		"a b":       "'a b'",
		"it's mine": `'it'\''s mine'`,
		"":          "''",
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Errorf("shellQuote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUtil_HumanBytes(t *testing.T) {
	// Exact values are tied to humanBytes' formatting; just sanity-check
	// the unit suffix bumps as size grows.
	if !strings.HasSuffix(humanBytes(500), " B") {
		t.Errorf("500 should be in bytes")
	}
	if !strings.Contains(humanBytes(2048), "KiB") {
		t.Errorf("2048 should be KiB")
	}
}

func TestUtil_ContainerName(t *testing.T) {
	if got := containerName("", "postgres"); got != "portunix-postgres" {
		t.Errorf("default container name = %q, want portunix-postgres", got)
	}
	if got := containerName("custom-pg", "postgres"); got != "custom-pg" {
		t.Errorf("explicit container name = %q, want custom-pg", got)
	}
}
