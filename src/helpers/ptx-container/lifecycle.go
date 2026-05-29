/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Label keys used to persist lifecycle metadata on containers. Storing this on
// the container itself (instead of a separate registry file) survives ptx
// crashes and is the runtime-native way to filter — `docker/podman ps --filter
// label=portunix.managed=true` is one query without locking concerns.
const (
	LabelManaged     = "portunix.managed"
	LabelTTL         = "portunix.ttl"
	LabelCreatedAt   = "portunix.created_at"
	LabelExpiresAt   = "portunix.expires_at"
	LabelPolicy      = "portunix.policy"
	LabelHealthCheck = "portunix.health_check"
)

// Cleanup policies map directly to issue #027:
//   - on-exit:  cleanup when ptx-container's parent process exits (signal handler)
//   - ttl:      cleanup when expires_at passes (background service or `cleanup` cmd)
//   - manual:   no automatic cleanup; user must call `container rm` or `cleanup`
const (
	PolicyOnExit = "on-exit"
	PolicyTTL    = "ttl"
	PolicyManual = "manual"
)

// LifecycleConfig captures the lifecycle-related flags from `container run`.
// All fields are optional; an empty config produces no portunix.* labels and
// behaves like the original run.
type LifecycleConfig struct {
	TTL         time.Duration
	AutoCleanup bool
	Policy      string
	HealthCheck string
}

// HasAny returns true when at least one lifecycle attribute is set, in which
// case the container should be tagged as managed.
func (lc LifecycleConfig) HasAny() bool {
	return lc.TTL > 0 || lc.AutoCleanup || lc.Policy != "" || lc.HealthCheck != ""
}

// resolvePolicy fills in a sensible default policy from the other flags when
// the user did not pass --cleanup-policy explicitly.
func (lc LifecycleConfig) resolvePolicy() string {
	if lc.Policy != "" {
		return lc.Policy
	}
	if lc.TTL > 0 {
		return PolicyTTL
	}
	if lc.AutoCleanup {
		return PolicyOnExit
	}
	return PolicyManual
}

// ToLabels turns the config into runtime label arguments. The returned slice is
// in `--label key=value` form ready to splice into a `docker/podman run` argv.
// Returns nil when the config has nothing to record.
func (lc LifecycleConfig) ToLabels(now time.Time) []string {
	if !lc.HasAny() {
		return nil
	}

	policy := lc.resolvePolicy()
	created := now.UTC().Format(time.RFC3339)

	labels := []string{
		"--label", LabelManaged + "=true",
		"--label", LabelCreatedAt + "=" + created,
		"--label", LabelPolicy + "=" + policy,
	}
	if lc.TTL > 0 {
		expires := now.Add(lc.TTL).UTC().Format(time.RFC3339)
		labels = append(labels,
			"--label", LabelTTL+"="+lc.TTL.String(),
			"--label", LabelExpiresAt+"="+expires,
		)
	}
	if lc.HealthCheck != "" {
		// Stored as label only; runtime-native --health-cmd is appended separately
		// (see RuntimeHealthArgs) so it's actually evaluated by docker/podman.
		labels = append(labels, "--label", LabelHealthCheck+"="+lc.HealthCheck)
	}
	return labels
}

// RuntimeHealthArgs returns the `--health-cmd` flags so the underlying runtime
// actually performs the check. Returns nil when no health check is configured.
func (lc LifecycleConfig) RuntimeHealthArgs() []string {
	if lc.HealthCheck == "" {
		return nil
	}
	return []string{"--health-cmd", lc.HealthCheck}
}

// LifecycleMetadata is the parsed view of portunix.* labels read back from a
// running container. Fields are zero-valued when the corresponding label is
// missing or unparseable — callers should check `Managed` first.
type LifecycleMetadata struct {
	Managed     bool
	TTL         time.Duration
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Policy      string
	HealthCheck string
}

// IsExpired reports whether a TTL-policy container has passed its deadline.
// Returns false for non-TTL policies or when ExpiresAt is zero.
func (m LifecycleMetadata) IsExpired(now time.Time) bool {
	if m.Policy != PolicyTTL || m.ExpiresAt.IsZero() {
		return false
	}
	return now.After(m.ExpiresAt)
}

// ParseLabels turns the label map produced by `inspect --format '{{json .Config.Labels}}'`
// into a LifecycleMetadata. Unrecognised keys are ignored. Bad timestamps and
// durations are silently dropped (zero values), so the caller sees Managed=true
// but Expired=false instead of a misleading expiry.
func ParseLabels(labels map[string]string) LifecycleMetadata {
	md := LifecycleMetadata{}
	if labels == nil {
		return md
	}
	if labels[LabelManaged] == "true" {
		md.Managed = true
	}
	md.Policy = labels[LabelPolicy]
	md.HealthCheck = labels[LabelHealthCheck]

	if s := labels[LabelTTL]; s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			md.TTL = d
		}
	}
	if s := labels[LabelCreatedAt]; s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			md.CreatedAt = t.UTC()
		}
	}
	if s := labels[LabelExpiresAt]; s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			md.ExpiresAt = t.UTC()
		}
	}
	return md
}

// ParseDuration accepts the standard Go syntax (e.g. "30s", "5m", "2h") plus
// "<n>d" for days, which Go's time.ParseDuration rejects but is the natural
// way to express CI/test TTLs ("--ttl 7d").
func ParseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if strings.HasSuffix(s, "d") {
		days, err := strconv.ParseFloat(strings.TrimSuffix(s, "d"), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", s, err)
		}
		return time.Duration(days * float64(24*time.Hour)), nil
	}
	return time.ParseDuration(s)
}

// ValidatePolicy returns an error when policy is set to something other than
// the three known values. Empty string is valid (resolved later from other
// flags).
func ValidatePolicy(policy string) error {
	switch policy {
	case "", PolicyOnExit, PolicyTTL, PolicyManual:
		return nil
	default:
		return fmt.Errorf("invalid cleanup-policy %q (must be %s, %s, or %s)",
			policy, PolicyOnExit, PolicyTTL, PolicyManual)
	}
}
