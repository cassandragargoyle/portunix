/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"strings"
	"testing"
)

// TestResolvePluginBinary_UnknownPluginMessage checks that the reason is printed once (#237)
func TestResolvePluginBinary_UnknownPluginMessage(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	pluginManager = nil
	t.Cleanup(func() { pluginManager = nil })

	_, _, _, err := resolvePluginBinary("nonexistent-plugin-xyz")
	if err == nil {
		t.Fatal("expected an error for an unknown plugin")
	}

	msg := err.Error()
	if got := strings.Count(msg, "not found"); got != 1 {
		t.Errorf("expected \"not found\" once, got %d times: %q", got, msg)
	}
	if !strings.Contains(msg, "nonexistent-plugin-xyz") {
		t.Errorf("expected plugin name in message, got %q", msg)
	}
}
