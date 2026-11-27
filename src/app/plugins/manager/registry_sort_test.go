/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"path/filepath"
	"reflect"
	"testing"

	"portunix.ai/app/plugins"
)

// newSortTestRegistry registers plugins with mixed-case names in a
// non-alphabetical order; all of them declare the "synapse" platform
func newSortTestRegistry(t *testing.T) *Registry {
	t.Helper()
	r, err := NewRegistry(filepath.Join(t.TempDir(), "registry.json"))
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	for _, name := range []string{"zeta", "Beta", "alpha", "foo", "Foo", "gamma", "Delta"} {
		r.data.Plugins[name] = &RegistryPlugin{
			Name:               name,
			Version:            "1.0.0",
			SupportedPlatforms: []plugins.SupportedPlatform{{Name: "synapse"}},
		}
	}
	return r
}

var wantSortedNames = []string{"alpha", "Beta", "Delta", "Foo", "foo", "gamma", "zeta"}

func TestListPlugins_SortedAlphabetically(t *testing.T) {
	r := newSortTestRegistry(t)

	// Repeat to catch random map iteration order leaking through
	for i := 0; i < 20; i++ {
		list, err := r.ListPlugins()
		if err != nil {
			t.Fatalf("ListPlugins: %v", err)
		}
		got := make([]string, 0, len(list))
		for _, p := range list {
			got = append(got, p.Name)
		}
		if !reflect.DeepEqual(got, wantSortedNames) {
			t.Fatalf("run %d: ListPlugins order = %v; want %v", i, got, wantSortedNames)
		}
	}
}

func TestListPluginsForPlatform_SortedAlphabetically(t *testing.T) {
	r := newSortTestRegistry(t)

	for i := 0; i < 20; i++ {
		matches, err := r.ListPluginsForPlatform("synapse", "", nil)
		if err != nil {
			t.Fatalf("ListPluginsForPlatform: %v", err)
		}
		got := pluginNames(matches)
		if !reflect.DeepEqual(got, wantSortedNames) {
			t.Fatalf("run %d: ListPluginsForPlatform order = %v; want %v", i, got, wantSortedNames)
		}
	}
}

func TestPluginNameLess(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"alpha", "beta", true},
		{"Beta", "alpha", false},
		{"alpha", "Beta", true},
		{"Foo", "foo", true},
		{"foo", "Foo", false},
		{"foo", "foo", false},
	}
	for _, tc := range cases {
		if got := pluginNameLess(tc.a, tc.b); got != tc.want {
			t.Errorf("pluginNameLess(%q, %q) = %v; want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
