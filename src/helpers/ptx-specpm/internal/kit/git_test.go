/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package kit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeRefForFS(t *testing.T) {
	cases := map[string]string{
		"main":    "main",
		"v1.2.3":  "v1.2.3",
		"feat/x":  "feat_x",
		"":        "default",
		"a:b@c~d": "a_b_c_d",
	}
	for in, want := range cases {
		if got := sanitizeRefForFS(in); got != want {
			t.Errorf("sanitizeRefForFS(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsPopulatedKit(t *testing.T) {
	dir := t.TempDir()
	if isPopulatedKit(dir) {
		t.Fatalf("empty dir should not be a populated kit")
	}
	for _, sub := range []string{"templates", "agents", "workflows"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if !isPopulatedKit(dir) {
		t.Fatalf("dir with templates/agents/workflows should pass")
	}
}

func TestLocalPathProvider_Validation(t *testing.T) {
	miss := filepath.Join(t.TempDir(), "missing")
	if _, _, err := NewLocalPathProvider(miss).Fetch(); err == nil {
		t.Fatalf("missing path should error")
	}

	dir := t.TempDir()
	if _, _, err := NewLocalPathProvider(dir).Fetch(); err == nil {
		t.Fatalf("non-kit dir should error")
	}

	for _, sub := range []string{"templates", "agents", "workflows"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	got, fromCache, err := NewLocalPathProvider(dir).Fetch()
	if err != nil {
		t.Fatalf("populated kit should succeed: %v", err)
	}
	if got != dir {
		t.Errorf("expected kit dir %q, got %q", dir, got)
	}
	if !fromCache {
		t.Errorf("local path provider should always report fromCache=true")
	}
}
