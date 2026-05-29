/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package main

import (
	"strings"
	"testing"
)

func TestRootCmd_HasDatabaseSubcommand(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "database" {
			found = true
			// Verify alias "db" is registered too.
			hasDB := false
			for _, a := range c.Aliases {
				if a == "db" {
					hasDB = true
				}
			}
			if !hasDB {
				t.Error("database command missing 'db' alias")
			}
		}
	}
	if !found {
		t.Fatal("rootCmd does not have a 'database' subcommand")
	}
}

func TestDatabaseCmd_HasExpectedSubcommands(t *testing.T) {
	expected := []string{
		"install", "uninstall",
		"start", "stop", "restart", "status", "health",
		"list", "tables", "schema",
		"backup", "restore",
		"engines",
	}
	got := map[string]bool{}
	for _, c := range dbCmd.Commands() {
		got[c.Name()] = true
	}
	for _, name := range expected {
		if !got[name] {
			t.Errorf("missing subcommand %q", name)
		}
	}
}

func TestVersionString(t *testing.T) {
	if version == "" {
		t.Error("version variable should not be empty")
	}
}

func TestResolveDriver_RequiresEngine(t *testing.T) {
	// Reset persistent flags to a clean state before the test.
	flagEngine, flagInstance, flagMode, flagFormat = "", "", "", ""

	_, _, err := resolveDriver(nil)
	if err == nil {
		t.Fatal("expected error when engine is unspecified")
	}
	if !strings.Contains(err.Error(), "engine required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveDriver_PositionalOverridesFlag(t *testing.T) {
	flagEngine = "sqlite"
	defer func() { flagEngine = "" }()

	d, _, err := resolveDriver([]string{"postgresql"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Engine() != "postgresql" {
		t.Errorf("positional arg should override --engine: got %q", d.Engine())
	}
}

func TestResolveDriver_DefaultModeForEmbedded(t *testing.T) {
	flagEngine = "sqlite"
	defer func() { flagEngine = ""; flagMode = "" }()

	_, opts, err := resolveDriver(nil)
	if err != nil {
		t.Fatal(err)
	}
	// SQLite supports only embedded — resolveDriver should pick it automatically.
	if string(opts.Mode) != "embedded" {
		t.Errorf("expected embedded mode for sqlite, got %q", opts.Mode)
	}
}
