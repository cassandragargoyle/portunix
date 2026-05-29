/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
		err  bool
	}{
		{"30s", 30 * time.Second, false},
		{"5m", 5 * time.Minute, false},
		{"2h", 2 * time.Hour, false},
		{"7d", 7 * 24 * time.Hour, false},
		{"1.5d", time.Duration(1.5 * float64(24*time.Hour)), false},
		{"", 0, true},
		{"abc", 0, true},
		{"3x", 0, true},
		{"-5m", -5 * time.Minute, false},
	}
	for _, c := range cases {
		got, err := ParseDuration(c.in)
		if c.err {
			if err == nil {
				t.Errorf("ParseDuration(%q) expected error, got %v", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseDuration(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseDuration(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestValidatePolicy(t *testing.T) {
	for _, p := range []string{"", PolicyOnExit, PolicyTTL, PolicyManual} {
		if err := ValidatePolicy(p); err != nil {
			t.Errorf("ValidatePolicy(%q) unexpected error: %v", p, err)
		}
	}
	if err := ValidatePolicy("forever"); err == nil {
		t.Error("ValidatePolicy(\"forever\") expected error")
	}
}

func TestLifecycleConfig_HasAny(t *testing.T) {
	if (LifecycleConfig{}).HasAny() {
		t.Error("zero LifecycleConfig should not be HasAny")
	}
	if !(LifecycleConfig{TTL: time.Hour}).HasAny() {
		t.Error("TTL alone should trigger HasAny")
	}
	if !(LifecycleConfig{AutoCleanup: true}).HasAny() {
		t.Error("AutoCleanup alone should trigger HasAny")
	}
	if !(LifecycleConfig{Policy: PolicyManual}).HasAny() {
		t.Error("Policy alone should trigger HasAny")
	}
	if !(LifecycleConfig{HealthCheck: "echo ok"}).HasAny() {
		t.Error("HealthCheck alone should trigger HasAny")
	}
}

func TestLifecycleConfig_resolvePolicy(t *testing.T) {
	if got := (LifecycleConfig{}).resolvePolicy(); got != PolicyManual {
		t.Errorf("empty -> %s, want %s", got, PolicyManual)
	}
	if got := (LifecycleConfig{TTL: time.Minute}).resolvePolicy(); got != PolicyTTL {
		t.Errorf("ttl-only -> %s, want %s", got, PolicyTTL)
	}
	if got := (LifecycleConfig{AutoCleanup: true}).resolvePolicy(); got != PolicyOnExit {
		t.Errorf("auto-cleanup -> %s, want %s", got, PolicyOnExit)
	}
	if got := (LifecycleConfig{TTL: time.Minute, Policy: PolicyManual}).resolvePolicy(); got != PolicyManual {
		t.Errorf("explicit policy should win: got %s", got)
	}
}

func TestLifecycleConfig_ToLabels(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	lc := LifecycleConfig{TTL: 2 * time.Hour, HealthCheck: "curl localhost"}

	labels := lc.ToLabels(now)
	// Each label is 2 args (--label key=value)
	if len(labels)%2 != 0 {
		t.Fatalf("ToLabels produced odd count: %d", len(labels))
	}

	// Build a map for assertion convenience.
	m := labelsToMap(labels)
	if m[LabelManaged] != "true" {
		t.Errorf("managed label missing: %v", m)
	}
	if m[LabelPolicy] != PolicyTTL {
		t.Errorf("policy label = %q, want %q", m[LabelPolicy], PolicyTTL)
	}
	if m[LabelTTL] != "2h0m0s" {
		t.Errorf("ttl label = %q", m[LabelTTL])
	}
	wantExpires := now.Add(2 * time.Hour).Format(time.RFC3339)
	if m[LabelExpiresAt] != wantExpires {
		t.Errorf("expires_at = %q, want %q", m[LabelExpiresAt], wantExpires)
	}
	if m[LabelHealthCheck] != "curl localhost" {
		t.Errorf("health_check = %q", m[LabelHealthCheck])
	}
}

func TestLifecycleConfig_ToLabels_Empty(t *testing.T) {
	if labels := (LifecycleConfig{}).ToLabels(time.Now()); labels != nil {
		t.Errorf("empty config should produce nil labels, got %v", labels)
	}
}

func TestParseLabels_RoundTrip(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	lc := LifecycleConfig{TTL: 30 * time.Minute, AutoCleanup: true, HealthCheck: "true"}

	args := lc.ToLabels(now)
	md := ParseLabels(labelsToMap(args))

	if !md.Managed {
		t.Error("expected Managed=true")
	}
	if md.TTL != 30*time.Minute {
		t.Errorf("TTL=%s, want 30m", md.TTL)
	}
	if md.Policy != PolicyTTL {
		t.Errorf("Policy=%q, want %q", md.Policy, PolicyTTL)
	}
	if !md.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt=%v, want %v", md.CreatedAt, now)
	}
	if !md.ExpiresAt.Equal(now.Add(30 * time.Minute)) {
		t.Errorf("ExpiresAt=%v", md.ExpiresAt)
	}
	if md.HealthCheck != "true" {
		t.Errorf("HealthCheck=%q", md.HealthCheck)
	}
}

func TestLifecycleMetadata_IsExpired(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		md   LifecycleMetadata
		want bool
	}{
		{"empty", LifecycleMetadata{}, false},
		{"manual policy ignored", LifecycleMetadata{Policy: PolicyManual, ExpiresAt: now.Add(-time.Hour)}, false},
		{"on-exit policy ignored", LifecycleMetadata{Policy: PolicyOnExit, ExpiresAt: now.Add(-time.Hour)}, false},
		{"ttl future", LifecycleMetadata{Policy: PolicyTTL, ExpiresAt: now.Add(time.Hour)}, false},
		{"ttl past", LifecycleMetadata{Policy: PolicyTTL, ExpiresAt: now.Add(-time.Second)}, true},
		{"ttl no expires_at", LifecycleMetadata{Policy: PolicyTTL}, false},
	}
	for _, c := range cases {
		if got := c.md.IsExpired(now); got != c.want {
			t.Errorf("%s: IsExpired = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestExtractLifecycleFlags(t *testing.T) {
	args := []string{
		"--name", "foo",
		"--ttl", "1h",
		"-p", "8080:80",
		"--auto-cleanup",
		"--cleanup-policy", "ttl",
		"--health-check", "curl localhost",
		"--max-memory", "512m",
		"--max-cpu", "1.5",
		"ubuntu:22.04",
		"bash",
	}
	rest, lc, err := extractLifecycleFlags(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lc.TTL != time.Hour {
		t.Errorf("TTL=%s", lc.TTL)
	}
	if !lc.AutoCleanup {
		t.Error("AutoCleanup not set")
	}
	if lc.Policy != PolicyTTL {
		t.Errorf("Policy=%q", lc.Policy)
	}
	if lc.HealthCheck != "curl localhost" {
		t.Errorf("HealthCheck=%q", lc.HealthCheck)
	}
	want := []string{
		"--name", "foo",
		"-p", "8080:80",
		"--memory", "512m",
		"--cpus", "1.5",
		"ubuntu:22.04",
		"bash",
	}
	if !equalStringSlice(rest, want) {
		t.Errorf("rest = %v\nwant %v", rest, want)
	}
}

func TestExtractLifecycleFlags_Errors(t *testing.T) {
	cases := [][]string{
		{"--ttl"},
		{"--ttl", "junk"},
		{"--cleanup-policy", "forever"},
		{"--max-memory"},
		{"--max-cpu"},
	}
	for _, args := range cases {
		if _, _, err := extractLifecycleFlags(args); err == nil {
			t.Errorf("expected error for %v", args)
		}
	}
}

func labelsToMap(args []string) map[string]string {
	m := map[string]string{}
	for i := 0; i+1 < len(args); i += 2 {
		if args[i] != "--label" {
			continue
		}
		kv := args[i+1]
		for j := 0; j < len(kv); j++ {
			if kv[j] == '=' {
				m[kv[:j]] = kv[j+1:]
				break
			}
		}
	}
	return m
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
