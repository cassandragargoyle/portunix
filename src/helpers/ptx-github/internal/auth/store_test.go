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

func tempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	s, err := NewStore(path, "")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func TestStoreSaveAndGet(t *testing.T) {
	s := tempStore(t)
	if err := s.SaveToken("default", "ghp_abc", "alice", true, false); err != nil {
		t.Fatalf("save: %v", err)
	}
	tok, acc, err := s.GetToken("")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if tok != "ghp_abc" {
		t.Fatalf("token: got %q", tok)
	}
	if acc.Name != "default" || acc.Username != "alice" {
		t.Fatalf("account meta: %+v", acc)
	}

	if _, err := os.Stat(s.Path()); err != nil {
		t.Fatalf("store file missing: %v", err)
	}

	info, err := os.Stat(s.Path())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("permissions too permissive: %o", info.Mode().Perm())
	}
}

func TestStoreMultipleAccountsAndDefault(t *testing.T) {
	s := tempStore(t)
	if err := s.SaveToken("work", "tok1", "u1", false, false); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveToken("personal", "tok2", "u2", false, false); err != nil {
		t.Fatal(err)
	}
	if err := s.SetDefault("personal"); err != nil {
		t.Fatal(err)
	}
	tok, acc, err := s.GetToken("")
	if err != nil {
		t.Fatal(err)
	}
	if tok != "tok2" || acc.Name != "personal" {
		t.Fatalf("default not honoured: tok=%q acc=%+v", tok, acc)
	}

	tok, _, err = s.GetToken("work")
	if err != nil || tok != "tok1" {
		t.Fatalf("named lookup: tok=%q err=%v", tok, err)
	}

	if err := s.Delete("personal"); err != nil {
		t.Fatal(err)
	}
	_, def, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if def != "work" {
		t.Fatalf("default after delete: %q", def)
	}
}

func TestStoreUnknownAccount(t *testing.T) {
	s := tempStore(t)
	if _, _, err := s.GetToken("missing"); err == nil {
		t.Fatalf("expected error for missing account")
	}
}
