/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRejectPlaintextPasswordFlag(t *testing.T) {
	cases := [][]string{
		{"ssh", "exec", "--password", "pw", "u@h", "cmd"},
		{"ssh", "exec", "--password=pw", "u@h", "cmd"},
	}
	for _, argv := range cases {
		err := RejectPlaintextPasswordFlag(argv)
		if err == nil {
			t.Fatalf("expected rejection for %v", argv)
		}
		if !strings.Contains(err.Error(), "--password") {
			t.Errorf("error should mention --password: %v", err)
		}
		if !strings.Contains(err.Error(), "--credential") {
			t.Errorf("error should list alternatives: %v", err)
		}
	}
}

func TestRejectPlaintextPasswordFlagAllowsOthers(t *testing.T) {
	argv := []string{"ssh", "exec", "--credential", "id", "u@h", "cmd"}
	if err := RejectPlaintextPasswordFlag(argv); err != nil {
		t.Fatalf("unexpected rejection: %v", err)
	}
}

func TestParseTarget(t *testing.T) {
	tests := []struct {
		in   string
		want Target
	}{
		{"alice@host", Target{User: "alice", Host: "host", Port: "22"}},
		{"alice@host:2222", Target{User: "alice", Host: "host", Port: "2222"}},
		{"host", Target{Host: "host", Port: "22"}},
		{"alice@[::1]:2022", Target{User: "alice", Host: "::1", Port: "2022"}},
	}
	for _, tt := range tests {
		got, err := ParseTarget(tt.in)
		if err != nil {
			t.Errorf("ParseTarget(%q) err: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseTarget(%q) = %+v, want %+v", tt.in, got, tt.want)
		}
	}
}

func TestParseTargetErrors(t *testing.T) {
	for _, in := range []string{"", "user@"} {
		if _, err := ParseTarget(in); err == nil {
			t.Errorf("expected error for %q", in)
		}
	}
}

func TestCredFlagsValidateMutualExclusion(t *testing.T) {
	f := CredFlags{Credential: "id", CredentialEnv: true}
	if err := f.Validate(); err == nil {
		t.Error("expected mutual exclusion error")
	}
	f = CredFlags{Credential: "id"}
	if err := f.Validate(); err != nil {
		t.Errorf("single source should validate: %v", err)
	}
	f = CredFlags{}
	if err := f.Validate(); err != nil {
		t.Errorf("empty should validate (falls back to default identity): %v", err)
	}
}

func TestReadFromFileRejectsLoosePerms(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission check is Unix-only")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "pass.txt")
	if err := os.WriteFile(path, []byte("hunter2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := readFromFile(path)
	if err == nil {
		t.Fatal("expected permission rejection for 0644 file")
	}
	if !strings.Contains(err.Error(), "permissions") {
		t.Errorf("expected permission error, got: %v", err)
	}
}

func TestReadFromFileAccepts0600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission check is Unix-only")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "pass.txt")
	if err := os.WriteFile(path, []byte("hunter2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := readFromFile(path)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if s.Reveal() != "hunter2" {
		t.Errorf("password = %q, want \"hunter2\"", s.Reveal())
	}
}

func TestReadFromFileRejectsEmpty(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "pass.txt")
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readFromFile(path); err == nil {
		t.Fatal("expected empty file rejection")
	}
}

func TestIsInteractiveUnderCI(t *testing.T) {
	t.Setenv("PORTUNIX_CI", "1")
	if IsInteractive() {
		t.Error("IsInteractive should be false under PORTUNIX_CI=1")
	}
}

func TestPromptPasswordUnderCIFails(t *testing.T) {
	t.Setenv("PORTUNIX_CI", "1")
	_, err := PromptPassword("x: ")
	if err == nil {
		t.Fatal("expected error when CI=1")
	}
	if !strings.Contains(err.Error(), "PORTUNIX_CI") {
		t.Errorf("expected PORTUNIX_CI mention: %v", err)
	}
}

func TestExpandPathTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	p, err := expandPath("~/foo")
	if err != nil {
		t.Fatal(err)
	}
	if p != filepath.Join(home, "foo") {
		t.Errorf("expandPath(~/foo) = %q, want %q", p, filepath.Join(home, "foo"))
	}
}
