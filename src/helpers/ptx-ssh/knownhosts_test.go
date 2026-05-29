/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"testing"
)

func TestParseHostKeyPolicy(t *testing.T) {
	cases := []struct {
		in      string
		want    HostKeyPolicy
		wantErr bool
	}{
		{"", PolicyAcceptNew, false},
		{"accept-new", PolicyAcceptNew, false},
		{"strict", PolicyStrict, false},
		{"off", PolicyOff, false},
		{"bogus", PolicyAcceptNew, true},
	}
	for _, tc := range cases {
		got, err := ParseHostKeyPolicy(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseHostKeyPolicy(%q) expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseHostKeyPolicy(%q) err: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseHostKeyPolicy(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestHostKeyOffRefusedUnderCI(t *testing.T) {
	t.Setenv("PORTUNIX_CI", "1")
	_, err := NewKnownHostsStore("", "", PolicyOff)
	if err == nil {
		t.Fatal("expected error when PORTUNIX_CI=1 and policy=off")
	}
}

func TestMatchesHost(t *testing.T) {
	if !matchesHost("example.com ssh-rsa AAAA...", "example.com") {
		t.Error("exact hostname should match")
	}
	if matchesHost("other.com ssh-rsa AAAA...", "example.com") {
		t.Error("different hostname should not match")
	}
	if !matchesHost("a.com,b.com ssh-rsa AAAA...", "b.com") {
		t.Error("comma-separated hosts should match")
	}
}

func TestMatchesHostBracketedForm(t *testing.T) {
	// Stored as "[127.0.0.1]:22522" (non-default port form), user queries
	// using the human-readable "127.0.0.1:22522".
	line := "[127.0.0.1]:22522 ecdsa-sha2-nistp256 AAAA..."
	if !matchesHost(line, "127.0.0.1:22522") {
		t.Error("host:port input should match [host]:port stored entry")
	}
	if !matchesHost(line, "127.0.0.1") {
		t.Error("bare host should match [host]:port stored entry")
	}
	if !matchesHost(line, "[127.0.0.1]:22522") {
		t.Error("exact bracketed form should match")
	}
	if matchesHost(line, "127.0.0.2:22522") {
		t.Error("different host should not match")
	}
	if matchesHost(line, "127.0.0.1:1234") {
		t.Error("different port should not match")
	}
}

func TestStripBrackets(t *testing.T) {
	h, p, ok := stripBrackets("[example.com]:2222")
	if !ok || h != "example.com" || p != "2222" {
		t.Errorf("stripBrackets = (%q,%q,%v)", h, p, ok)
	}
	_, _, ok = stripBrackets("plain.com")
	if ok {
		t.Error("plain hostname should report ok=false")
	}
	_, _, ok = stripBrackets("[]:22")
	if ok {
		t.Error("empty host should report ok=false")
	}
}
