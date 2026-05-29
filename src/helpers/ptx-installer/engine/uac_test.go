/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"strings"
	"testing"
)

func TestPSSingleQuote(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", `C:\Users\me\installer.exe`, `'C:\Users\me\installer.exe'`},
		{"with space", `C:\Program Files\Docker\installer.exe`, `'C:\Program Files\Docker\installer.exe'`},
		{"single quote escaped by doubling", `it's-a-trap`, `'it''s-a-trap'`},
		{"empty", ``, `''`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := psSingleQuote(tc.in)
			if got != tc.want {
				t.Errorf("psSingleQuote(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestBuildUACScript(t *testing.T) {
	t.Run("no args: omits ArgumentList", func(t *testing.T) {
		s := buildUACScript(`C:\foo.exe`, nil)
		if strings.Contains(s, "-ArgumentList") {
			t.Errorf("expected no -ArgumentList for empty args, got: %s", s)
		}
		if !strings.Contains(s, `-FilePath 'C:\foo.exe'`) {
			t.Errorf("expected quoted FilePath, got: %s", s)
		}
	})

	t.Run("with args: builds quoted ArgumentList", func(t *testing.T) {
		s := buildUACScript(`C:\foo.exe`, []string{"install", "--accept-license"})
		if !strings.Contains(s, `-ArgumentList @('install','--accept-license')`) {
			t.Errorf("expected quoted ArgumentList, got: %s", s)
		}
	})

	t.Run("UAC decline mapped to exit 1223", func(t *testing.T) {
		s := buildUACScript(`C:\foo.exe`, nil)
		if !strings.Contains(s, "exit 1223") {
			t.Errorf("expected UAC-decline exit code 1223 in script, got: %s", s)
		}
		if !strings.Contains(s, "canceled by the user") {
			t.Errorf("expected cancellation match clause, got: %s", s)
		}
	})

	t.Run("forwards child exit code", func(t *testing.T) {
		s := buildUACScript(`C:\foo.exe`, nil)
		if !strings.Contains(s, "exit $p.ExitCode") {
			t.Errorf("expected child exit-code forwarding, got: %s", s)
		}
	})

	t.Run("escapes embedded single quote in path", func(t *testing.T) {
		s := buildUACScript(`C:\it's\foo.exe`, nil)
		if !strings.Contains(s, `'C:\it''s\foo.exe'`) {
			t.Errorf("expected doubled single quote in FilePath, got: %s", s)
		}
	})
}
