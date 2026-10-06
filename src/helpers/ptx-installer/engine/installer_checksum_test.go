/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helloSHA256 is the SHA-256 digest of "hello"
const helloSHA256 = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

func TestVerifyFileChecksum(t *testing.T) {
	path := filepath.Join(t.TempDir(), "installer.exe")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tests := []struct {
		name     string
		expected string
		wantErr  string
	}{
		{"prefixed", "sha256:" + helloSHA256, ""},
		{"plain hex", helloSHA256, ""},
		{"uppercase", "SHA256:" + strings.ToUpper(helloSHA256), ""},
		{"mismatch", "sha256:" + strings.Repeat("0", 64), "checksum mismatch"},
		{"unsupported algorithm", "md5:5d41402abc4b2a76b9719d911017c592", "unsupported checksum algorithm"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyFileChecksum(path, tt.expected)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("verifyFileChecksum() unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("verifyFileChecksum() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestVerifyFileChecksum_MissingFile(t *testing.T) {
	if err := verifyFileChecksum(filepath.Join(t.TempDir(), "missing.exe"), helloSHA256); err == nil {
		t.Fatal("expected error for a missing file, got nil")
	}
}
