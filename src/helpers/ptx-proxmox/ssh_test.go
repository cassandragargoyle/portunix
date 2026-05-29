/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"net/http"
	"testing"
)

func TestResolveIP_VMAgent(t *testing.T) {
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api2/json/nodes/pve1/qemu/100/agent/network-get-interfaces" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":{"result":[
			{"name":"lo","ip-addresses":[{"ip-address":"127.0.0.1","ip-address-type":"ipv4"}]},
			{"name":"eth0","ip-addresses":[
				{"ip-address":"fe80::1","ip-address-type":"ipv6"},
				{"ip-address":"10.0.0.5","ip-address-type":"ipv4"}
			]}
		]}}`))
	})
	ref := &ResourceRef{Node: "pve1", Type: ResourceTypeVM, VMID: 100}
	ip, err := c.ResolveIP(ref, true)
	if err != nil {
		t.Fatalf("ResolveIP: %v", err)
	}
	if ip != "10.0.0.5" {
		t.Errorf("expected 10.0.0.5 (first non-loopback IPv4), got %q", ip)
	}
}

func TestResolveIP_CTInterfaces(t *testing.T) {
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api2/json/nodes/pve1/lxc/200/interfaces" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":[
			{"name":"lo","inet":"127.0.0.1/8"},
			{"name":"eth0","inet":"10.0.0.6/24","inet6":"fe80::6/64"}
		]}`))
	})
	ref := &ResourceRef{Node: "pve1", Type: ResourceTypeCT, VMID: 200}
	ip, err := c.ResolveIP(ref, true)
	if err != nil {
		t.Fatalf("ResolveIP: %v", err)
	}
	if ip != "10.0.0.6" {
		t.Errorf("expected 10.0.0.6, got %q", ip)
	}
}

func TestResolveIP_AgentNotInstalled(t *testing.T) {
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "agent not found", http.StatusInternalServerError)
	})
	ref := &ResourceRef{Node: "pve1", Type: ResourceTypeVM, VMID: 100}
	_, err := c.ResolveIP(ref, true)
	if err == nil {
		t.Fatal("expected error when agent missing")
	}
}

func TestSplitSCPTarget(t *testing.T) {
	cases := []struct {
		in       string
		ident    string
		path     string
		isRemote bool
	}{
		{"./local.txt", "", "./local.txt", false},
		{"my-vm:/tmp/foo", "my-vm", "/tmp/foo", true},
		{"100:/tmp/foo", "100", "/tmp/foo", true},
		{"C:\\Users\\x\\foo.txt", "", "C:\\Users\\x\\foo.txt", false},
	}
	for _, c := range cases {
		ident, path, isRemote := splitSCPTarget(c.in)
		if ident != c.ident || path != c.path || isRemote != c.isRemote {
			t.Errorf("splitSCPTarget(%q) = (%q,%q,%v), want (%q,%q,%v)",
				c.in, ident, path, isRemote, c.ident, c.path, c.isRemote)
		}
	}
}

func TestBuildSCPArgs_PortAndIdentity(t *testing.T) {
	d := &sshDefaults{port: 2222, identityFile: "/tmp/key", extraOptions: []string{"StrictHostKeyChecking=no"}}
	got := buildSCPArgs(d, []string{"-r", "src", "dst"})
	want := []string{"-P", "2222", "-i", "/tmp/key", "-o", "StrictHostKeyChecking=no", "-r", "src", "dst"}
	if len(got) != len(want) {
		t.Fatalf("buildSCPArgs len mismatch: got %v want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("arg %d: got %q, want %q", i, got[i], want[i])
		}
	}
}
