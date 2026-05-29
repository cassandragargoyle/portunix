/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"strings"
	"testing"
)

func TestValidateWindowsPath(t *testing.T) {
	d := NewDockerInstaller(false, "", false)
	drives := []DriveInfo{
		{Letter: "C", FreeSpace: "200.0 GB", TotalSpace: "500.0 GB"},
		{Letter: "D", FreeSpace: "5.0 GB", TotalSpace: "100.0 GB"},
		{Letter: "E", FreeSpace: "50.0 GB", TotalSpace: "500.0 GB"},
	}

	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{"valid D drive", "D:\\docker-data", "insufficient free space"},
		{"valid E drive", "E:\\docker-data", ""},
		{"valid C drive", "C:\\my-docker", ""},
		{"missing drive letter", "\\docker-data", "must include drive letter"},
		{"relative path", "docker-data", "must include drive letter"},
		{"unknown drive", "Z:\\docker-data", "not found"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := d.validateWindowsPath(tc.path, drives)
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tc.wantErr, err)
			}
		})
	}
}

func TestResolveLinuxDataRoot_ExplicitFlag(t *testing.T) {
	d := NewDockerInstaller(false, "/mnt/ssd/docker", true)
	path, explicit, err := d.resolveLinuxDataRoot(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/mnt/ssd/docker" {
		t.Errorf("expected /mnt/ssd/docker, got %q", path)
	}
	if !explicit {
		t.Errorf("expected explicit=true for --data-root flag")
	}
}

func TestResolveWindowsDataRoot_ExplicitFlag(t *testing.T) {
	d := NewDockerInstaller(false, "D:\\docker-data", true)
	drives := []DriveInfo{
		{Letter: "C", FreeSpace: "200.0 GB", TotalSpace: "500.0 GB"},
		{Letter: "D", FreeSpace: "100.0 GB", TotalSpace: "500.0 GB"},
	}
	path, explicit, err := d.resolveWindowsDataRoot(drives)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "D:\\docker-data" {
		t.Errorf("expected D:\\docker-data, got %q", path)
	}
	if !explicit {
		t.Errorf("expected explicit=true for --data-root flag")
	}
}

func TestResolveWindowsDataRoot_ExplicitFlag_InsufficientSpace(t *testing.T) {
	d := NewDockerInstaller(false, "D:\\docker-data", true)
	drives := []DriveInfo{
		{Letter: "C", FreeSpace: "200.0 GB", TotalSpace: "500.0 GB"},
		{Letter: "D", FreeSpace: "3.0 GB", TotalSpace: "100.0 GB"},
	}
	_, _, err := d.resolveWindowsDataRoot(drives)
	if err == nil || !strings.Contains(err.Error(), "insufficient free space") {
		t.Errorf("expected insufficient-space error, got: %v", err)
	}
}
