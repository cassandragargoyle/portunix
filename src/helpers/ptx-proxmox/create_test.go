/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestNextVMID(t *testing.T) {
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api2/json/cluster/nextid" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":"102"}`))
	})
	id, err := c.NextVMID()
	if err != nil {
		t.Fatalf("NextVMID: %v", err)
	}
	if id != 102 {
		t.Errorf("expected 102, got %d", id)
	}
}

func TestCreateVM_Form(t *testing.T) {
	var (
		seenPath string
		seenForm url.Values
	)
	p := &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"}
	c, _ := newTestClient(t, p, func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		seenForm = r.PostForm
		_, _ = w.Write([]byte(`{"data":"UPID:create-vm"}`))
	})

	form := url.Values{}
	form.Set("vmid", "101")
	form.Set("name", "test-vm")
	form.Set("cores", "2")
	form.Set("memory", "4096")
	upid, err := c.CreateVM("pve1", form)
	if err != nil {
		t.Fatalf("CreateVM: %v", err)
	}
	if !strings.HasPrefix(upid, "UPID:") {
		t.Errorf("unexpected UPID: %s", upid)
	}
	if seenPath != "/api2/json/nodes/pve1/qemu" {
		t.Errorf("unexpected path: %s", seenPath)
	}
	if seenForm.Get("vmid") != "101" || seenForm.Get("name") != "test-vm" {
		t.Errorf("form values wrong: %v", seenForm)
	}
}

func TestUpdateVMConfig_OnlyVM(t *testing.T) {
	c, _ := newTestClient(t, &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"},
		func(w http.ResponseWriter, r *http.Request) {
			// Should not be called for CT
			t.Fatalf("unexpected request for CT: %s", r.URL.Path)
		})
	ref := &ResourceRef{Node: "pve1", Type: ResourceTypeCT, VMID: 200}
	if err := c.UpdateVMConfig(ref, url.Values{"ciuser": {"deploy"}}); err == nil {
		t.Fatal("expected error for CT target")
	}
}

func TestUpdateVMConfig_PutsToConfigEndpoint(t *testing.T) {
	var (
		method   string
		seenPath string
		seenForm url.Values
	)
	c, _ := newTestClient(t, &Profile{AuthType: AuthTypeToken, TokenID: "a", TokenSecret: "b"},
		func(w http.ResponseWriter, r *http.Request) {
			method = r.Method
			seenPath = r.URL.Path
			_ = r.ParseForm()
			seenForm = r.PostForm
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":null}`))
		})
	ref := &ResourceRef{Node: "pve1", Type: ResourceTypeVM, VMID: 100}
	form := url.Values{"ciuser": {"deploy"}, "ipconfig0": {"ip=dhcp"}}
	if err := c.UpdateVMConfig(ref, form); err != nil {
		t.Fatalf("UpdateVMConfig: %v", err)
	}
	if method != http.MethodPut {
		t.Errorf("expected PUT, got %s", method)
	}
	if seenPath != "/api2/json/nodes/pve1/qemu/100/config" {
		t.Errorf("unexpected path: %s", seenPath)
	}
	if seenForm.Get("ciuser") != "deploy" || seenForm.Get("ipconfig0") != "ip=dhcp" {
		t.Errorf("form values wrong: %v", seenForm)
	}
}
