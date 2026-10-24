/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"archive/zip"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestWheelTagPythonVersion(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"ptx_demo-1.0.0-cp313-cp313-linux_x86_64.whl", "3.13"},
		{"ptx_demo-1.0.0-1-cp39-cp39-win_amd64.whl", "3.9"},
		{"ptx_demo-1.0.0-cp314-none-any.whl", "3.14"},
		{"ptx_demo-1.0.0-py3-none-any.whl", ""},
		{"ptx_demo-1.0.0-cp312.cp313-abi3-any.whl", ""},
		{"ptx_demo-1.0.0.whl", ""},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := wheelTagPythonVersion(tt.filename); got != tt.want {
				t.Errorf("wheelTagPythonVersion(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestPycMagicVersion(t *testing.T) {
	tests := []struct {
		magic uint16
		want  string
	}{
		{3413, "3.8"},
		{3439, "3.10"},
		{3495, "3.11"},
		{3531, "3.12"},
		{3571, "3.13"},
		{3627, "3.14"},
		{20121, ""}, // Python 2.7
	}
	for _, tt := range tests {
		if got := pycMagicVersion(tt.magic); got != tt.want {
			t.Errorf("pycMagicVersion(%d) = %q, want %q", tt.magic, got, tt.want)
		}
	}
}

// writeTestWheel creates a wheel zip with the given files (name -> content)
func writeTestWheel(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

// pycHeader returns a minimal .pyc header for the magic number
func pycHeader(magic uint16) []byte {
	header := make([]byte, 16)
	binary.LittleEndian.PutUint16(header, magic)
	header[2], header[3] = '\r', '\n'
	return header
}

func TestDetectWheelPythonVersion(t *testing.T) {
	dir := t.TempDir()

	t.Run("sourceless bytecode", func(t *testing.T) {
		path := filepath.Join(dir, "ptx_bc-1.0.0-py3-none-any.whl")
		writeTestWheel(t, path, map[string][]byte{
			"ptx_bc/__init__.pyc": pycHeader(3571),
		})
		if got := detectWheelPythonVersion(path); got != "3.13" {
			t.Errorf("got %q, want 3.13", got)
		}
	})

	t.Run("cached bytecode ignored", func(t *testing.T) {
		path := filepath.Join(dir, "ptx_src-1.0.0-py3-none-any.whl")
		writeTestWheel(t, path, map[string][]byte{
			"ptx_src/__init__.py":                          []byte("x = 1\n"),
			"ptx_src/__pycache__/__init__.cpython-313.pyc": pycHeader(3571),
		})
		if got := detectWheelPythonVersion(path); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("filename tag wins", func(t *testing.T) {
		path := filepath.Join(dir, "ptx_tag-1.0.0-cp312-cp312-linux_x86_64.whl")
		writeTestWheel(t, path, map[string][]byte{
			"ptx_tag/__init__.pyc": pycHeader(3571),
		})
		if got := detectWheelPythonVersion(path); got != "3.12" {
			t.Errorf("got %q, want 3.12", got)
		}
	})

	t.Run("not a zip", func(t *testing.T) {
		path := filepath.Join(dir, "ptx_bad-1.0.0-py3-none-any.whl")
		if err := os.WriteFile(path, []byte("not a zip"), 0644); err != nil {
			t.Fatal(err)
		}
		if got := detectWheelPythonVersion(path); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}
