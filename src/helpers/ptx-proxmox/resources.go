/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"net/url"
	"strconv"
)

// ListResources returns all VMs and CTs visible to the user. typeFilter may be
// empty (both), "vm" (qemu only), or "ct" (lxc only). nodeFilter restricts to
// a single node (empty = all nodes).
func (c *Client) ListResources(typeFilter, nodeFilter string) ([]ResourceRef, error) {
	var raw []clusterResource
	// /cluster/resources?type=vm returns both QEMU and LXC — a quirk of the
	// API that saves us a second call.
	if err := c.get("/cluster/resources?type=vm", &raw); err != nil {
		return nil, err
	}
	out := make([]ResourceRef, 0, len(raw))
	for _, r := range raw {
		if r.Type != string(ResourceTypeVM) && r.Type != string(ResourceTypeCT) {
			continue
		}
		if typeFilter == "vm" && r.Type != string(ResourceTypeVM) {
			continue
		}
		if typeFilter == "ct" && r.Type != string(ResourceTypeCT) {
			continue
		}
		if nodeFilter != "" && r.Node != nodeFilter {
			continue
		}
		out = append(out, ResourceRef{
			Node:    r.Node,
			Type:    ResourceType(r.Type),
			VMID:    r.VMID,
			Name:    r.Name,
			Status:  r.Status,
			MaxMem:  r.MaxMem,
			MaxDisk: r.MaxDisk,
			CPUs:    r.CPUs,
		})
	}
	return out, nil
}

// ResolveResource looks up a VM or CT by either numeric vmid or name. Names
// must be unique across the cluster (matches Proxmox UI behaviour — duplicate
// names cause ambiguity errors here).
func (c *Client) ResolveResource(identifier string) (*ResourceRef, error) {
	if identifier == "" {
		return nil, fmt.Errorf("empty identifier")
	}
	all, err := c.ListResources("", "")
	if err != nil {
		return nil, err
	}
	// Numeric identifier → direct vmid match.
	if id, err := strconv.Atoi(identifier); err == nil {
		for i := range all {
			if all[i].VMID == id {
				return &all[i], nil
			}
		}
		return nil, fmt.Errorf("no VM/CT with vmid %d", id)
	}
	// Name-based match.
	var matches []*ResourceRef
	for i := range all {
		if all[i].Name == identifier {
			matches = append(matches, &all[i])
		}
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no VM/CT named %q", identifier)
	case 1:
		return matches[0], nil
	default:
		nodes := make([]string, 0, len(matches))
		for _, m := range matches {
			nodes = append(nodes, fmt.Sprintf("%s:%d", m.Node, m.VMID))
		}
		return nil, fmt.Errorf("ambiguous name %q — also found on %v — use numeric vmid instead",
			identifier, nodes)
	}
}

// StatusCurrent returns the "current" status object for a VM or CT. Proxmox
// returns a free-form map that differs between qemu and lxc, so we expose it
// as-is for the CLI to pretty-print.
func (c *Client) StatusCurrent(r *ResourceRef) (map[string]interface{}, error) {
	path := fmt.Sprintf("/nodes/%s/%s/%d/status/current",
		url.PathEscape(r.Node), r.Type.PathSegment(), r.VMID)
	var out map[string]interface{}
	if err := c.get(path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Config returns the VM or CT configuration (cores, memory, net0, etc.).
func (c *Client) Config(r *ResourceRef) (map[string]interface{}, error) {
	path := fmt.Sprintf("/nodes/%s/%s/%d/config",
		url.PathEscape(r.Node), r.Type.PathSegment(), r.VMID)
	var out map[string]interface{}
	if err := c.get(path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// lifecyclePost issues a POST to /status/<action> and returns the task UPID.
// Proxmox write endpoints always return a UPID as the `data` field.
func (c *Client) lifecyclePost(r *ResourceRef, action string, form url.Values) (string, error) {
	path := fmt.Sprintf("/nodes/%s/%s/%d/status/%s",
		url.PathEscape(r.Node), r.Type.PathSegment(), r.VMID, action)
	var upid string
	if err := c.postForm(path, form, &upid); err != nil {
		return "", err
	}
	return upid, nil
}

// Start boots a VM or CT. Returns the task UPID.
func (c *Client) Start(r *ResourceRef) (string, error) { return c.lifecyclePost(r, "start", nil) }

// Stop forces a power-off (hard stop). Use Shutdown for graceful.
func (c *Client) Stop(r *ResourceRef) (string, error) { return c.lifecyclePost(r, "stop", nil) }

// Shutdown sends the guest OS a shutdown signal.
func (c *Client) Shutdown(r *ResourceRef) (string, error) { return c.lifecyclePost(r, "shutdown", nil) }

// Reboot restarts a VM or CT.
func (c *Client) Reboot(r *ResourceRef) (string, error) {
	// LXC uses "reboot"; QEMU accepts "reboot" as well since 7.x. Older
	// servers may require "reset" for qemu — not our target here.
	return c.lifecyclePost(r, "reboot", nil)
}

// Destroy deletes the VM or CT. Returns the task UPID.
func (c *Client) Destroy(r *ResourceRef, purge bool) (string, error) {
	path := fmt.Sprintf("/nodes/%s/%s/%d",
		url.PathEscape(r.Node), r.Type.PathSegment(), r.VMID)
	if purge {
		path += "?purge=1"
	}
	var upid string
	if err := c.delete(path, &upid); err != nil {
		return "", err
	}
	return upid, nil
}
