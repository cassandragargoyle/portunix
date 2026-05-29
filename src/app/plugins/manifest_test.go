/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package plugins

import (
	"encoding/json"
	"strings"
	"testing"
)

// buildManifest returns a minimal valid manifest we can decorate with
// supported_platforms entries under test.
func buildManifest() *PluginManifest {
	return &PluginManifest{
		Name:        "demo",
		Version:     "1.0.0",
		Description: "demo plugin for tests",
		Author:      "tests",
		License:     "MIT",
		Plugin: PluginBinaryConfig{
			Type:    "grpc",
			Binary:  "./demo",
			Runtime: "native",
			Port:    9001,
		},
		Dependencies: PluginDependencies{
			OSSupport: []string{"linux"},
		},
	}
}

func TestValidateManifest_NoSupportedPlatforms_BackwardCompat(t *testing.T) {
	m := buildManifest()
	if err := ValidateManifest(m); err != nil {
		t.Fatalf("baseline manifest without supported_platforms must validate: %v", err)
	}
}

func TestValidateSupportedPlatforms_ValidEntries(t *testing.T) {
	m := buildManifest()
	m.SupportedPlatforms = []SupportedPlatform{
		{
			Name:            "synapse",
			MinVersion:      "0.3.0",
			MaxVersion:      "0.9.9",
			Features:        []string{"connector.command", "assistant.chat"},
			PlatformPayload: json.RawMessage(`{"surfaces":{"connector.command":true}}`),
		},
		{
			Name: "pack", // min/max/features/payload all optional
		},
	}
	if err := ValidateManifest(m); err != nil {
		t.Fatalf("valid supported_platforms rejected: %v", err)
	}
}

func TestValidateSupportedPlatforms_InvalidEntries(t *testing.T) {
	cases := []struct {
		label     string
		entries   []SupportedPlatform
		wantError string
	}{
		{
			"missing name",
			[]SupportedPlatform{{Name: ""}},
			"name is required",
		},
		{
			"invalid name pattern",
			[]SupportedPlatform{{Name: "Synapse"}},
			"invalid name",
		},
		{
			"duplicate platform",
			[]SupportedPlatform{{Name: "synapse"}, {Name: "synapse"}},
			"declared more than once",
		},
		{
			"invalid min_version",
			[]SupportedPlatform{{Name: "synapse", MinVersion: "abc"}},
			"min_version",
		},
		{
			"invalid max_version",
			[]SupportedPlatform{{Name: "synapse", MaxVersion: "1.x"}},
			"max_version",
		},
		{
			"min > max",
			[]SupportedPlatform{{Name: "synapse", MinVersion: "1.0.0", MaxVersion: "0.9.0"}},
			"must not exceed max_version",
		},
		{
			"invalid feature token",
			[]SupportedPlatform{{Name: "synapse", Features: []string{"Bad Feature"}}},
			"invalid feature",
		},
		{
			"duplicated feature",
			[]SupportedPlatform{{Name: "synapse", Features: []string{"a", "a"}}},
			"declared more than once",
		},
		{
			"platform_payload is array, not object",
			[]SupportedPlatform{{Name: "synapse", PlatformPayload: json.RawMessage(`[1,2]`)}},
			"must be a JSON object",
		},
		{
			"platform_payload is primitive",
			[]SupportedPlatform{{Name: "synapse", PlatformPayload: json.RawMessage(`"hello"`)}},
			"must be a JSON object",
		},
		{
			"platform_payload invalid JSON",
			[]SupportedPlatform{{Name: "synapse", PlatformPayload: json.RawMessage(`{not:valid}`)}},
			"invalid JSON",
		},
	}
	for _, c := range cases {
		m := buildManifest()
		m.SupportedPlatforms = c.entries
		err := ValidateManifest(m)
		if err == nil {
			t.Errorf("%s: expected error containing %q, got nil", c.label, c.wantError)
			continue
		}
		if !strings.Contains(err.Error(), c.wantError) {
			t.Errorf("%s: error = %q, want substring %q", c.label, err.Error(), c.wantError)
		}
	}
}

func TestLoadManifest_PreservesPlatformPayloadBytes(t *testing.T) {
	// Deliberately use an unusual key order and whitespace to verify that
	// json.RawMessage round-trips the bytes without re-serializing.
	rawPayload := `{"z":1,"a":{"nested":true},"m":"value"}`
	manifestJSON := `{
  "name": "demo",
  "version": "1.0.0",
  "description": "d",
  "author": "a",
  "plugin": {"type":"grpc","binary":"./d","runtime":"native","port":9001},
  "dependencies": {"os_support":["linux"]},
  "supported_platforms": [
    {
      "name": "synapse",
      "platform_payload": ` + rawPayload + `
    }
  ]
}`
	var m PluginManifest
	if err := json.Unmarshal([]byte(manifestJSON), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := ValidateManifest(&m); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(m.SupportedPlatforms) != 1 {
		t.Fatalf("expected 1 platform, got %d", len(m.SupportedPlatforms))
	}
	got := string(m.SupportedPlatforms[0].PlatformPayload)
	if got != rawPayload {
		t.Errorf("platform_payload bytes differ:\n got = %q\nwant = %q", got, rawPayload)
	}
}
