/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"testing"
	"time"
)

func mc(name string, md LifecycleMetadata, running bool, status string) ManagedContainer {
	return ManagedContainer{
		Runtime: "docker", Name: name, ID: name + "-id",
		Status: status, Running: running, Metadata: md,
	}
}

func TestMatchesCleanupFilter_DefaultIsExpiredOnly(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	expired := mc("a", LifecycleMetadata{
		Policy: PolicyTTL, ExpiresAt: now.Add(-time.Hour),
	}, false, "exited")
	live := mc("b", LifecycleMetadata{
		Policy: PolicyTTL, ExpiresAt: now.Add(time.Hour),
	}, true, "running")

	f := CleanupFilter{}
	if !matchesCleanupFilter(expired, f, now) {
		t.Error("default filter should match expired containers")
	}
	if matchesCleanupFilter(live, f, now) {
		t.Error("default filter should NOT match live TTL containers")
	}
}

func TestMatchesCleanupFilter_All(t *testing.T) {
	now := time.Now()
	live := mc("b", LifecycleMetadata{
		Policy: PolicyTTL, ExpiresAt: now.Add(time.Hour),
	}, true, "running")
	if !matchesCleanupFilter(live, CleanupFilter{All: true}, now) {
		t.Error("--all should match running containers")
	}
}

func TestMatchesCleanupFilter_ExcludeRunning(t *testing.T) {
	now := time.Now()
	running := mc("r", LifecycleMetadata{Policy: PolicyManual}, true, "running")
	if matchesCleanupFilter(running, CleanupFilter{All: true, ExcludeRunning: true}, now) {
		t.Error("--exclude-running should exclude running containers")
	}
}

func TestMatchesCleanupFilter_Pattern(t *testing.T) {
	now := time.Now()
	dev := mc("dev-foo", LifecycleMetadata{Policy: PolicyManual}, false, "exited")
	prod := mc("prod-bar", LifecycleMetadata{Policy: PolicyManual}, false, "exited")

	f := CleanupFilter{Pattern: "dev-*"}
	if !matchesCleanupFilter(dev, f, now) {
		t.Error("pattern dev-* should match dev-foo")
	}
	if matchesCleanupFilter(prod, f, now) {
		t.Error("pattern dev-* should NOT match prod-bar")
	}
}

func TestMatchesCleanupFilter_OlderThan(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	old := mc("old", LifecycleMetadata{
		Policy: PolicyManual, CreatedAt: now.Add(-48 * time.Hour),
	}, false, "exited")
	fresh := mc("fresh", LifecycleMetadata{
		Policy: PolicyManual, CreatedAt: now.Add(-30 * time.Minute),
	}, false, "exited")

	f := CleanupFilter{OlderThan: 24 * time.Hour}
	if !matchesCleanupFilter(old, f, now) {
		t.Error("older-than 24h should match a 48h-old container")
	}
	if matchesCleanupFilter(fresh, f, now) {
		t.Error("older-than 24h should NOT match a 30m-old container")
	}
}

func TestMatchesCleanupFilter_StatusInsensitive(t *testing.T) {
	now := time.Now()
	ex := mc("e", LifecycleMetadata{Policy: PolicyManual}, false, "Exited")
	if !matchesCleanupFilter(ex, CleanupFilter{Status: "exited"}, now) {
		t.Error("status filter should be case-insensitive")
	}
}

func TestMatchesCleanupFilter_PositiveFilterIgnoresExpiredOnlyDefault(t *testing.T) {
	now := time.Now()
	// Manual-policy containers are never "expired" — but with --pattern they
	// should still match because the user supplied a positive filter.
	c := mc("dev-1", LifecycleMetadata{Policy: PolicyManual, CreatedAt: now}, false, "exited")
	if !matchesCleanupFilter(c, CleanupFilter{Pattern: "dev-*"}, now) {
		t.Error("positive --pattern should match even non-expired manual containers")
	}
}
