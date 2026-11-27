/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package plugins

import "testing"

// TestCommandWarnings_Placeholder checks the placeholder command detection
// of `plugin validate` (issue #257)
func TestCommandWarnings_Placeholder(t *testing.T) {
	cases := []struct {
		plugin   string
		binary   string
		commands []string
		warn     bool
	}{
		{"scraper", "./ptx-scraper", []string{"ptx-scraper"}, true},
		{"scraper", "./scraper", []string{"scraper"}, true},
		{"scraper", "./scraper", []string{"Scraper"}, true},
		{"m365-extractor", "./m365", []string{"m365"}, true},
		{"text-extractor", "./tika-wrapper.jar", []string{"tika-wrapper"}, true},
		{"scraper", "./scraper", []string{"run"}, false},
		{"scraper", "./scraper", []string{"scraper", "run"}, false},
		{"scraper", "./scraper", nil, false},
		{"scraper", "./scraper", []string{"s"}, false},
		{"m365-extractor", "./m365", []string{"m365-export"}, false},
	}
	for _, c := range cases {
		manifest := &PluginManifest{Name: c.plugin, Plugin: PluginBinaryConfig{Binary: c.binary}}
		for _, name := range c.commands {
			manifest.Commands = append(manifest.Commands, PluginCommand{Name: name})
		}
		got := len(CommandWarnings(manifest)) > 0
		if got != c.warn {
			t.Errorf("plugin %q binary %q commands %v: warning = %v, want %v", c.plugin, c.binary, c.commands, got, c.warn)
		}
	}
}

// TestCreateDefaultManifest_NoPlaceholderCommand checks that the generated
// template follows the commands rule
func TestCreateDefaultManifest_NoPlaceholderCommand(t *testing.T) {
	manifest := CreateDefaultManifest("demo", "demo plugin", "tests")
	if warnings := CommandWarnings(manifest); len(warnings) > 0 {
		t.Errorf("default manifest has command warnings: %v", warnings)
	}
	for _, c := range manifest.Commands {
		if c.Name == "help" {
			t.Error("default manifest must not list a 'help' command")
		}
	}
}
