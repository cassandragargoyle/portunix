/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ManagedContainer is the in-memory view of a single Portunix-managed container
// queried back from a runtime. It pairs the runtime-side identity (ID/Name/
// Status/Image/runtime) with the parsed lifecycle metadata.
type ManagedContainer struct {
	Runtime  string // "docker" or "podman"
	ID       string
	Name     string
	Image    string
	Status   string // raw status string (e.g. "Up 5 minutes", "Exited (0) 1 hour ago")
	Running  bool
	Metadata LifecycleMetadata
}

// ListManaged returns all containers in the requested runtime that carry the
// `portunix.managed=true` label. Pass "" for both runtime values to query
// every available runtime.
//
// Containers without the label are intentionally excluded — issue #027 cleanup
// must never touch user-managed containers that ptx didn't create.
func ListManaged(runtime string) ([]ManagedContainer, error) {
	if runtime == "" {
		var all []ManagedContainer
		var firstErr error
		if isPodmanInstalled() {
			out, err := listManagedFor("podman")
			if err != nil && firstErr == nil {
				firstErr = err
			}
			all = append(all, out...)
		}
		if isDockerInstalled() {
			out, err := listManagedFor("docker")
			if err != nil && firstErr == nil {
				firstErr = err
			}
			all = append(all, out...)
		}
		return all, firstErr
	}
	return listManagedFor(runtime)
}

func listManagedFor(runtime string) ([]ManagedContainer, error) {
	// `--filter label=key=value` matches both running and stopped containers
	// when combined with `-a`. We use one inspect-per-container after the ps
	// stage to read the full label set — runtime ps output truncates labels.
	cmd := exec.Command(runtime, "ps", "-a",
		"--filter", "label="+LabelManaged+"=true",
		"--format", "{{.ID}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s ps failed: %w", runtime, err)
	}

	ids := splitNonEmptyLines(string(output))
	if len(ids) == 0 {
		return nil, nil
	}

	var managed []ManagedContainer
	for _, id := range ids {
		mc, err := inspectManagedFor(runtime, id)
		if err != nil {
			// Container may have been removed between ps and inspect — skip
			// instead of failing the whole listing.
			continue
		}
		managed = append(managed, mc)
	}
	return managed, nil
}

// InspectManaged looks up a single managed container by name or ID across the
// available runtimes. Returns ErrNotManaged when the container exists but is
// not tagged as portunix.managed=true.
func InspectManaged(nameOrID string) (ManagedContainer, error) {
	var lastErr error
	for _, rt := range availableRuntimes() {
		mc, err := inspectManagedFor(rt, nameOrID)
		if err == nil {
			return mc, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("container %q not found", nameOrID)
	}
	return ManagedContainer{}, lastErr
}

// ErrNotManaged is returned by InspectManaged when the container exists but
// has no portunix.managed label. Callers should treat this distinctly from
// "container not found".
var ErrNotManaged = fmt.Errorf("container is not portunix-managed")

// inspectShape mirrors the subset of `docker/podman inspect` JSON we need.
// We deliberately do not pull in any container library — a tiny shape is
// enough and keeps ptx-container's go.mod free of runtime SDKs.
type inspectShape struct {
	ID    string `json:"Id"`
	Name  string `json:"Name"`
	Image string `json:"Image"`
	State struct {
		Status  string `json:"Status"`
		Running bool   `json:"Running"`
	} `json:"State"`
	Config struct {
		Labels map[string]string `json:"Labels"`
		Image  string            `json:"Image"`
	} `json:"Config"`
}

func inspectManagedFor(runtime, nameOrID string) (ManagedContainer, error) {
	cmd := exec.Command(runtime, "inspect", nameOrID)
	output, err := cmd.Output()
	if err != nil {
		return ManagedContainer{}, fmt.Errorf("%s inspect %s: %w", runtime, nameOrID, err)
	}

	var raw []inspectShape
	if err := json.Unmarshal(output, &raw); err != nil {
		return ManagedContainer{}, fmt.Errorf("decode inspect output: %w", err)
	}
	if len(raw) == 0 {
		return ManagedContainer{}, fmt.Errorf("container %q not found", nameOrID)
	}

	r := raw[0]
	md := ParseLabels(r.Config.Labels)
	if !md.Managed {
		return ManagedContainer{}, ErrNotManaged
	}

	image := r.Config.Image
	if image == "" {
		image = r.Image
	}

	return ManagedContainer{
		Runtime:  runtime,
		ID:       r.ID,
		Name:     strings.TrimPrefix(r.Name, "/"),
		Image:    image,
		Status:   r.State.Status,
		Running:  r.State.Running,
		Metadata: md,
	}, nil
}

// FindExpired returns the subset of managed containers whose TTL has elapsed.
// It is a thin filter over ListManaged so the same results power the cleanup
// command and the background service.
func FindExpired(now time.Time) ([]ManagedContainer, error) {
	all, err := ListManaged("")
	if err != nil {
		return nil, err
	}
	var expired []ManagedContainer
	for _, mc := range all {
		if mc.Metadata.IsExpired(now) {
			expired = append(expired, mc)
		}
	}
	return expired, nil
}

// availableRuntimes lists the runtimes we can actually call inspect against.
// Order matters: podman first matches the rest of ptx-container's preference.
func availableRuntimes() []string {
	var rts []string
	if isPodmanInstalled() {
		rts = append(rts, "podman")
	}
	if isDockerInstalled() {
		rts = append(rts, "docker")
	}
	return rts
}

func splitNonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
