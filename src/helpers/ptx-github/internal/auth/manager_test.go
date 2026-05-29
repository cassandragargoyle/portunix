/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManagerEnvWinsOverStoreDefault(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "a.json"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveToken("default", "store-token", "", true, false); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PORTUNIX_GITHUB_TOKEN", "env-token")
	t.Setenv("GITHUB_TOKEN", "")

	mgr := NewManager(store)
	r, err := mgr.Resolve("")
	if err != nil {
		t.Fatal(err)
	}
	if r.Token != "env-token" || r.Source != SourceEnv {
		t.Fatalf("expected env, got %+v", r)
	}
}

func TestManagerStoreUsedWhenEnvEmpty(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "a.json"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveToken("default", "store-token", "", true, false); err != nil {
		t.Fatal(err)
	}
	for _, v := range EnvTokenVars {
		_ = os.Unsetenv(v)
	}
	mgr := NewManager(store)
	r, err := mgr.Resolve("")
	if err != nil {
		t.Fatal(err)
	}
	if r.Token != "store-token" || r.Source != SourceStore || r.AccountName != "default" {
		t.Fatalf("expected store/default, got %+v", r)
	}
}

func TestManagerNamedAccount(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "a.json"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveToken("default", "default-tok", "", true, false); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveToken("ci", "ci-tok", "", false, false); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PORTUNIX_GITHUB_TOKEN", "env-token")

	mgr := NewManager(store)
	r, err := mgr.Resolve("ci")
	if err != nil {
		t.Fatal(err)
	}
	if r.Token != "ci-tok" || r.Source != SourceStore {
		t.Fatalf("expected ci/store, got %+v", r)
	}
}

func TestManagerNoAuth(t *testing.T) {
	for _, v := range EnvTokenVars {
		_ = os.Unsetenv(v)
	}
	mgr := NewManager(nil)
	r, err := mgr.Resolve("")
	if err != nil {
		t.Fatal(err)
	}
	if r.Token != "" || r.Source != SourceNone {
		t.Fatalf("expected none, got %+v", r)
	}
}
