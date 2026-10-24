/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package plugins

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// writeScript writes an executable shell script and returns its path
func writeScript(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0755); err != nil { // #nosec G306 -- test executable
		t.Fatal(err)
	}
	return path
}

func skipOnWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell-script interpreters are POSIX-specific")
	}
}

func TestRecomputeOverall(t *testing.T) {
	r := &PrerequisiteCheckResult{
		Runtime:         &RuntimeCheckResult{Status: CheckStatusError},
		PortunixVersion: &VersionCheckResult{Status: CheckStatusOK},
		OSSupport:       &OSCheckResult{Status: CheckStatusOK},
	}
	r.recomputeOverall()
	if r.Overall != CheckStatusError {
		t.Fatalf("Overall = %s, want error", r.Overall)
	}

	// Replacing the failing runtime check clears the error
	r.SetRuntimeCheck(&RuntimeCheckResult{Status: CheckStatusOK})
	if r.Overall != CheckStatusOK {
		t.Fatalf("Overall = %s, want ok", r.Overall)
	}

	r.SetRuntimeCheck(&RuntimeCheckResult{Status: CheckStatusWarning})
	if r.Overall != CheckStatusWarning {
		t.Fatalf("Overall = %s, want warning", r.Overall)
	}

	r.SetInstallCheck(&InstallCheckResult{Status: CheckStatusError})
	if r.Overall != CheckStatusError {
		t.Fatalf("Overall = %s, want error after broken install", r.Overall)
	}
}

func TestCheckVenvPythonRuntime(t *testing.T) {
	skipOnWindows(t)
	dir := t.TempDir()
	interpreter := writeScript(t, dir, "python", `echo "Python 3.14.1"`)

	result := CheckVenvPythonRuntime("python", "", interpreter, "portunix plugin repair demo")
	if result.Status != CheckStatusOK || result.Found != "3.14.1" {
		t.Fatalf("got %+v, want ok with Found 3.14.1", result)
	}

	result = CheckVenvPythonRuntime("python >=3.15", ">=3.15", interpreter, "portunix plugin repair demo")
	if result.Status != CheckStatusError || result.Fix != "portunix plugin repair demo" {
		t.Fatalf("got %+v, want version error with repair fix", result)
	}

	result = CheckVenvPythonRuntime("python", "", filepath.Join(dir, "missing"), "portunix plugin repair demo")
	if result.Status != CheckStatusError || result.Fix != "portunix plugin repair demo" {
		t.Fatalf("got %+v, want error with repair fix", result)
	}
}

func TestCheckUVPythonRuntime(t *testing.T) {
	skipOnWindows(t)
	dir := t.TempDir()
	interpreter := writeScript(t, dir, "python3.14", `echo "Python 3.14.2"`)

	uvFound := writeScript(t, dir, "uv-found", "echo "+interpreter)
	result := CheckUVPythonRuntime("python", uvFound, "3.14")
	if result.Status != CheckStatusOK || result.Found != "3.14.2" {
		t.Fatalf("got %+v, want ok with Found 3.14.2", result)
	}

	uvMissing := writeScript(t, dir, "uv-missing", "exit 2")
	result = CheckUVPythonRuntime("python", uvMissing, "3.14")
	if result.Status != CheckStatusWarning {
		t.Fatalf("got %+v, want warning when uv has to download Python", result)
	}
	if result.Found != "" || strings.Count(result.Message, "uv") != 1 {
		t.Fatalf("got %+v, want empty Found and \"uv\" once in the message", result)
	}
}

func TestSatisfiesVersionConstraint(t *testing.T) {
	cases := []struct {
		version, constraint string
		want                bool
	}{
		{"3.14", ">=3.10", true},
		{"3.14", ">=3.14", true},
		{"3.14", ">=3.14.2", false},
		{"3.13", ">=3.15", false},
	}
	for _, c := range cases {
		if got := SatisfiesVersionConstraint(c.version, c.constraint); got != c.want {
			t.Errorf("SatisfiesVersionConstraint(%q, %q) = %v, want %v", c.version, c.constraint, got, c.want)
		}
	}
}
