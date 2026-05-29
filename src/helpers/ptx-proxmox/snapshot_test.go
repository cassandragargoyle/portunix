/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestSnapshots_List(t *testing.T) {
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api2/json/nodes/pve1/qemu/100/snapshot" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":[
			{"name":"current"},
			{"name":"pre-upgrade","parent":"current","description":"before 8.2","snaptime":1713000000}
		]}`))
	})
	ref := &ResourceRef{Node: "pve1", Type: ResourceTypeVM, VMID: 100}
	snaps, err := c.Snapshots(ref)
	if err != nil {
		t.Fatalf("Snapshots: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snaps))
	}
	if snaps[1].Name != "pre-upgrade" || snaps[1].Description != "before 8.2" {
		t.Errorf("unexpected snapshot data: %+v", snaps[1])
	}
}

func TestCreateSnapshot_Form(t *testing.T) {
	var form string
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		form = r.PostForm.Encode()
		_, _ = w.Write([]byte(`{"data":"UPID:snapcreate"}`))
	})
	ref := &ResourceRef{Node: "pve1", Type: ResourceTypeVM, VMID: 100}
	if _, err := c.CreateSnapshot(ref, "before-upgrade", "disk + ram", true); err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}
	if !strings.Contains(form, "snapname=before-upgrade") {
		t.Errorf("missing snapname in form: %s", form)
	}
	if !strings.Contains(form, "vmstate=1") {
		t.Errorf("missing vmstate=1 for --include-ram: %s", form)
	}
}

func TestDeleteSnapshot_Force(t *testing.T) {
	var query string
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"data":"UPID:snapdel"}`))
	})
	ref := &ResourceRef{Node: "pve1", Type: ResourceTypeVM, VMID: 100}
	if _, err := c.DeleteSnapshot(ref, "old", true); err != nil {
		t.Fatalf("DeleteSnapshot: %v", err)
	}
	if query != "force=1" {
		t.Errorf("expected force=1, got %q", query)
	}
}
