/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"portunix.ai/app/plugins"
)

// newTestRegistry builds a Registry with a hand-crafted in-memory dataset so
// we can exercise ListPluginsForPlatform without hitting the filesystem-backed
// install path. Each entry declares a realistic supported_platforms mix.
func newTestRegistry(t *testing.T) *Registry {
	t.Helper()
	tmp := t.TempDir()
	r, err := NewRegistry(filepath.Join(tmp, "registry.json"))
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	r.data.Plugins["alpha"] = &RegistryPlugin{
		Name:    "alpha",
		Version: "1.0.0",
		SupportedPlatforms: []plugins.SupportedPlatform{
			{
				Name:            "synapse",
				MinVersion:      "0.3.0",
				MaxVersion:      "0.9.9",
				Features:        []string{"connector.command", "assistant.chat"},
				PlatformPayload: json.RawMessage(`{"surfaces":{"connector.command":true}}`),
			},
		},
	}
	r.data.Plugins["beta"] = &RegistryPlugin{
		Name:    "beta",
		Version: "2.5.0",
		SupportedPlatforms: []plugins.SupportedPlatform{
			{
				Name:       "synapse",
				MinVersion: "1.0.0",
				Features:   []string{"connector.command"},
			},
		},
	}
	r.data.Plugins["gamma"] = &RegistryPlugin{
		Name:    "gamma",
		Version: "0.1.0",
		SupportedPlatforms: []plugins.SupportedPlatform{
			{Name: "pack"},
		},
	}
	r.data.Plugins["delta"] = &RegistryPlugin{
		Name:    "delta",
		Version: "3.0.0",
		// No supported_platforms — pre-1.1.0 plugin, must never match.
	}
	return r
}

func TestListPluginsForPlatform_FiltersByPlatformName(t *testing.T) {
	r := newTestRegistry(t)

	got, err := r.ListPluginsForPlatform("synapse", "", nil)
	if err != nil {
		t.Fatalf("ListPluginsForPlatform: %v", err)
	}
	names := pluginNames(got)
	wantAny(t, names, []string{"alpha", "beta"})
	wantNone(t, names, []string{"gamma", "delta"})
}

func TestListPluginsForPlatform_SemVerRange(t *testing.T) {
	r := newTestRegistry(t)

	// 0.5.0 ∈ [0.3.0, 0.9.9]: only alpha matches.
	got, err := r.ListPluginsForPlatform("synapse", "0.5.0", nil)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	names := pluginNames(got)
	wantAny(t, names, []string{"alpha"})
	wantNone(t, names, []string{"beta"})

	// 1.5.0 ≥ 1.0.0 (beta) and > 0.9.9 (alpha max): only beta matches.
	got, err = r.ListPluginsForPlatform("synapse", "1.5.0", nil)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	names = pluginNames(got)
	wantAny(t, names, []string{"beta"})
	wantNone(t, names, []string{"alpha"})
}

func TestListPluginsForPlatform_FeatureANDFilter(t *testing.T) {
	r := newTestRegistry(t)

	// Asking for a feature only alpha declares should drop beta.
	got, err := r.ListPluginsForPlatform("synapse", "", []string{"assistant.chat"})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	names := pluginNames(got)
	wantAny(t, names, []string{"alpha"})
	wantNone(t, names, []string{"beta"})

	// Both features: still only alpha.
	got, err = r.ListPluginsForPlatform("synapse", "", []string{"connector.command", "assistant.chat"})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	names = pluginNames(got)
	if len(names) != 1 || names[0] != "alpha" {
		t.Errorf("got %v, want [alpha]", names)
	}
	if len(got[0].MatchedFeatures) != 2 {
		t.Errorf("matched_features = %v, want 2 entries", got[0].MatchedFeatures)
	}
}

func TestListPluginsForPlatform_UnknownPlatform(t *testing.T) {
	r := newTestRegistry(t)
	got, err := r.ListPluginsForPlatform("agent", "", nil)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result for unknown platform, got %v", pluginNames(got))
	}
}

func TestListPluginsForPlatform_PreservesPayloadBytes(t *testing.T) {
	r := newTestRegistry(t)
	got, err := r.ListPluginsForPlatform("synapse", "", nil)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	for _, m := range got {
		if m.Plugin.Name == "alpha" {
			want := `{"surfaces":{"connector.command":true}}`
			if string(m.Platform.PlatformPayload) != want {
				t.Errorf("payload bytes = %q, want %q", string(m.Platform.PlatformPayload), want)
			}
			return
		}
	}
	t.Fatal("alpha not found in result")
}

func TestListPluginsForPlatform_EmptyPlatformName(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.ListPluginsForPlatform("", "", nil); err == nil {
		t.Error("expected error for empty platform name")
	}
}

// --- helpers ---

func pluginNames(m []MatchedPlugin) []string {
	out := make([]string, len(m))
	for i, x := range m {
		out[i] = x.Plugin.Name
	}
	return out
}

func wantAny(t *testing.T, got, want []string) {
	t.Helper()
	set := make(map[string]struct{}, len(got))
	for _, n := range got {
		set[n] = struct{}{}
	}
	for _, w := range want {
		if _, ok := set[w]; !ok {
			t.Errorf("result missing %q (got %v)", w, got)
		}
	}
}

func wantNone(t *testing.T, got, excluded []string) {
	t.Helper()
	set := make(map[string]struct{}, len(got))
	for _, n := range got {
		set[n] = struct{}{}
	}
	for _, e := range excluded {
		if _, ok := set[e]; ok {
			t.Errorf("unexpected %q in result (got %v)", e, got)
		}
	}
}
