/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package dispatcher

import (
	"strings"
	"testing"
)

// TestHelpLevelArgs verifies that --help-expert / --help-ai reach a helper
// as --help unless the helper handles them in the top level form (issue #203)
func TestHelpLevelArgs(t *testing.T) {
	d := &Dispatcher{helpers: make(map[string]*HelperConfig)}
	d.registerHelpers()

	cases := []struct {
		in, want string
	}{
		// docker install ran the installation before the rewrite
		{"docker install --help-expert", "docker install --help"},
		{"install --help-expert", "install --help"},
		{"install claude-code --help-ai", "install claude-code --help"},
		{"container --help-expert", "container --help"},
		{"wizard --help-ai", "wizard --help"},
		{"install --help-ai=true", "install --help"},
		// helpers that handle the flags themselves keep the top level form
		{"mcp --help-expert", "mcp --help-expert"},
		{"github --help-ai", "github --help-ai"},
		{"github clone --help-expert", "github clone --help"},
		// untouched arguments
		{"install --help-expert=false", "install --help-expert=false"},
		{"ssh exec host -- tool --help-ai", "ssh exec host -- tool --help-ai"},
		{"install nodejs --dry-run", "install nodejs --dry-run"},
		{"--help-expert", "--help-expert"},
	}
	for _, c := range cases {
		got := strings.Join(d.helpLevelArgs(strings.Fields(c.in)), " ")
		if got != c.want {
			t.Errorf("helpLevelArgs(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestHelpLevelArgs_DoesNotModifyInput verifies the caller's slice stays as is
func TestHelpLevelArgs_DoesNotModifyInput(t *testing.T) {
	d := &Dispatcher{helpers: make(map[string]*HelperConfig)}
	d.registerHelpers()
	args := []string{"install", "--help-expert"}
	d.helpLevelArgs(args)
	if args[1] != "--help-expert" {
		t.Errorf("input modified: %v", args)
	}
}
