/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package container

import (
	"strings"
	"testing"
)

// fakeChecks returns runtime checks with fixed answers
func fakeChecks(dockerAvailable, dockerRunning, podmanAvailable, podmanRunning bool) runtimeChecks {
	answer := func(v bool) func() bool { return func() bool { return v } }
	return runtimeChecks{
		dockerAvailable: answer(dockerAvailable),
		dockerRunning:   answer(dockerRunning),
		podmanAvailable: answer(podmanAvailable),
		podmanRunning:   answer(podmanRunning),
	}
}

func TestChooseRuntime(t *testing.T) {
	tests := []struct {
		name       string
		configured string
		checks     runtimeChecks
		want       string
		wantErr    string
	}{
		{name: "TC-038-I008: Docker configured and available", configured: "docker", checks: fakeChecks(true, true, false, false), want: "docker"},
		{name: "TC-038-I009: Podman configured and available", configured: "podman", checks: fakeChecks(false, false, true, true), want: "podman"},
		{name: "TC-038-I010: Docker configured but not installed", configured: "docker", checks: fakeChecks(false, false, true, true), wantErr: "Docker not installed"},
		{name: "TC-038-I011: No runtime available", configured: "docker", checks: fakeChecks(false, false, false, false), wantErr: "Docker not installed"},
		{name: "TC-038-I012: Default to podman when no config", configured: "", checks: fakeChecks(true, true, true, true), want: "podman"},
		{name: "Docker installed but daemon not running", configured: "docker", checks: fakeChecks(true, false, true, true), wantErr: "daemon not running"},
		{name: "Podman installed but not functional", configured: "podman", checks: fakeChecks(true, true, true, false), wantErr: "not functional"},
		// No fallback to Docker: the default runtime must be usable itself
		{name: "Default podman missing, docker present", configured: "", checks: fakeChecks(true, true, false, false), wantErr: "Podman not installed"},
		{name: "Unknown runtime", configured: "lxc", checks: fakeChecks(true, true, true, true), wantErr: "unknown container runtime: lxc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := chooseRuntime(tt.configured, tt.checks)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("chooseRuntime() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("chooseRuntime() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("chooseRuntime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChooseRuntimeSkipsDaemonCheckWhenNotInstalled(t *testing.T) {
	checks := fakeChecks(false, false, false, false)
	checks.dockerRunning = func() bool {
		t.Error("daemon check ran although Docker is not installed")
		return false
	}
	if _, err := chooseRuntime("docker", checks); err == nil {
		t.Fatal("expected an error for a missing Docker")
	}
}
