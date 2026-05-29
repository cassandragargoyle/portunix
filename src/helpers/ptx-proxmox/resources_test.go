/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// clusterResourcesResponse is the JSON body TestListResources serves — matches
// what a real Proxmox cluster returns, pared down to fields we actually read.
const clusterResourcesResponse = `{"data":[
  {"id":"qemu/100","type":"qemu","node":"pve1","vmid":100,"name":"web","status":"running","maxmem":4294967296,"maxdisk":32212254720,"cpus":2},
  {"id":"qemu/101","type":"qemu","node":"pve2","vmid":101,"name":"db","status":"stopped","maxmem":8589934592,"maxdisk":53687091200,"cpus":4},
  {"id":"lxc/200","type":"lxc","node":"pve1","vmid":200,"name":"cache","status":"running","maxmem":1073741824,"maxdisk":8589934592,"cpus":1},
  {"id":"node/pve1","type":"node","node":"pve1"}
]}`

func TestListResources_NoFilter(t *testing.T) {
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api2/json/cluster/resources") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(clusterResourcesResponse))
	})
	items, err := c.ListResources("", "")
	if err != nil {
		t.Fatalf("ListResources: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 VM/CT items, got %d", len(items))
	}
}

func TestListResources_TypeAndNodeFilter(t *testing.T) {
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(clusterResourcesResponse))
	})
	// CTs only
	items, err := c.ListResources("ct", "")
	if err != nil {
		t.Fatalf("ListResources ct: %v", err)
	}
	if len(items) != 1 || items[0].Type != ResourceTypeCT {
		t.Fatalf("expected 1 CT, got %#v", items)
	}
	// Only pve1 VMs
	items, err = c.ListResources("vm", "pve1")
	if err != nil {
		t.Fatalf("ListResources vm pve1: %v", err)
	}
	if len(items) != 1 || items[0].VMID != 100 {
		t.Fatalf("expected vmid 100, got %#v", items)
	}
}

func TestResolveResource_ByVMID(t *testing.T) {
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(clusterResourcesResponse))
	})
	r, err := c.ResolveResource("101")
	if err != nil {
		t.Fatalf("resolve 101: %v", err)
	}
	if r.Node != "pve2" || r.Name != "db" {
		t.Fatalf("wrong resolution: %+v", r)
	}
}

func TestResolveResource_ByName(t *testing.T) {
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(clusterResourcesResponse))
	})
	r, err := c.ResolveResource("cache")
	if err != nil {
		t.Fatalf("resolve by name: %v", err)
	}
	if r.Type != ResourceTypeCT || r.VMID != 200 {
		t.Fatalf("wrong resolution: %+v", r)
	}
}

func TestResolveResource_NotFound(t *testing.T) {
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(clusterResourcesResponse))
	})
	if _, err := c.ResolveResource("nope"); err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestResolveResource_AmbiguousName(t *testing.T) {
	// Build a response with duplicate names.
	body, _ := json.Marshal(map[string]interface{}{"data": []map[string]interface{}{
		{"id": "qemu/100", "type": "qemu", "node": "pve1", "vmid": 100, "name": "dup"},
		{"id": "qemu/101", "type": "qemu", "node": "pve2", "vmid": 101, "name": "dup"},
	}})
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	})
	_, err := c.ResolveResource("dup")
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected ambiguous error, got %v", err)
	}
}

func TestLifecycle_PostsToCorrectPath(t *testing.T) {
	var seenPath string
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		// Proxmox wraps UPIDs in a string.
		_, _ = w.Write([]byte(`{"data":"UPID:pve1:00001234:ABCD:0:qmstart:100:root@pam:"}`))
	})
	ref := &ResourceRef{Node: "pve1", Type: ResourceTypeVM, VMID: 100}
	upid, err := c.Start(ref)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !strings.HasPrefix(upid, "UPID:") {
		t.Errorf("expected UPID prefix, got %q", upid)
	}
	if seenPath != "/api2/json/nodes/pve1/qemu/100/status/start" {
		t.Errorf("unexpected start path: %s", seenPath)
	}
}

func TestDestroy_WithPurge(t *testing.T) {
	var seenPath string
	var seenQuery string
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"data":"UPID:destroy"}`))
	})
	ref := &ResourceRef{Node: "pve1", Type: ResourceTypeCT, VMID: 200}
	if _, err := c.Destroy(ref, true); err != nil {
		t.Fatalf("Destroy: %v", err)
	}
	if seenPath != "/api2/json/nodes/pve1/lxc/200" {
		t.Errorf("destroy path: %s", seenPath)
	}
	if seenQuery != "purge=1" {
		t.Errorf("expected purge=1 query, got %q", seenQuery)
	}
}

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		in   uint64
		want string
	}{
		{0, "0B"},
		{500, "500B"},
		{1024, "1.0K"},
		{1024 * 1024, "1.0M"},
		{4 * 1024 * 1024 * 1024, "4.0G"},
	}
	for _, c := range cases {
		if got := humanBytes(c.in); got != c.want {
			t.Errorf("humanBytes(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}
