/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateManifest_Runtimes(t *testing.T) {
	tests := []struct {
		runtime        string
		runtimeVersion string
		wantErr        string
	}{
		{runtime: ""},
		{runtime: "native"},
		{runtime: "java", runtimeVersion: ">=21"},
		{runtime: "python"},
		{runtime: "node"},
		{runtime: "node", runtimeVersion: ">=18"},
		{runtime: "jvm", wantErr: "unsupported runtime: jvm (supported: native, java, python, node)"},
		{runtime: "nodejs", wantErr: "unsupported runtime: nodejs"},
	}

	for _, tt := range tests {
		t.Run(tt.runtime+tt.runtimeVersion, func(t *testing.T) {
			m := buildManifest()
			m.Plugin.Runtime = tt.runtime
			m.Plugin.RuntimeVersion = tt.runtimeVersion

			err := ValidateManifest(m)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

// TestJavaScriptTemplateManifest keeps the JavaScript plugin template valid
func TestJavaScriptTemplateManifest(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "plugin-development", "languages", "javascript", "template", "plugin.json")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("load template manifest: %v", err)
	}
	if err := ValidateManifest(m); err != nil {
		t.Fatalf("template manifest must validate: %v", err)
	}
	if m.Plugin.Runtime != "node" {
		t.Errorf("runtime = %q, want node", m.Plugin.Runtime)
	}
	if strings.ContainsAny(m.Plugin.Binary, " \t") {
		t.Errorf("binary = %q, want the entry script path only", m.Plugin.Binary)
	}
}

// templatesDir is the directory of the language plugin templates
var templatesDir = filepath.Join("..", "..", "..", "docs", "plugin-development", "languages")

// TestLanguageTemplateManifests keeps the Go, Java and Rust plugin templates valid
func TestLanguageTemplateManifests(t *testing.T) {
	tests := []struct {
		language string
		runtime  string
	}{
		{language: "go", runtime: "native"},
		{language: "java", runtime: "java"},
		{language: "rust", runtime: "native"},
	}
	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			m, err := LoadManifest(filepath.Join(templatesDir, tt.language, "template", "plugin.json"))
			if err != nil {
				t.Fatalf("load template manifest: %v", err)
			}
			if err := ValidateManifest(m); err != nil {
				t.Fatalf("template manifest must validate: %v", err)
			}
			if m.Plugin.Runtime != tt.runtime {
				t.Errorf("runtime = %q, want %q", m.Plugin.Runtime, tt.runtime)
			}
			if len(m.Commands) == 0 {
				t.Error("template manifest must declare commands, the plugin serves them over gRPC")
			}
		})
	}
}

// protoDefinitions returns the proto lines without comments, blank lines,
// trailing whitespace and option lines, so only definitions are compared;
// options are language specific (go_package, java_package, ...) and may differ
func protoDefinitions(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path) // #nosec G304 -- test reads repository files
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "option ") {
			continue
		}
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// TestTemplateProtosMatchContract keeps every template proto equal to the
// PluginService part of the Portunix plugin contract
func TestTemplateProtosMatchContract(t *testing.T) {
	contract := protoDefinitions(t, filepath.Join("proto", "plugin.proto"))

	// Templates carry the base service only, not DatastorePluginService
	end := strings.Index(contract, "service DatastorePluginService")
	if end < 0 {
		t.Fatal("DatastorePluginService not found in the contract")
	}
	want := strings.TrimSpace(contract[:end])

	templates := map[string]string{
		"go":         filepath.Join("go", "template", "proto", "plugin.proto"),
		"java":       filepath.Join("java", "template", "src", "main", "proto", "plugin.proto"),
		"javascript": filepath.Join("javascript", "template", "proto", "plugin.proto"),
		"rust":       filepath.Join("rust", "template", "proto", "plugin.proto"),
	}
	for language, rel := range templates {
		t.Run(language, func(t *testing.T) {
			if got := protoDefinitions(t, filepath.Join(templatesDir, rel)); got != want {
				t.Errorf("%s template proto differs from the PluginService contract in src/app/plugins/proto/plugin.proto\n--- template\n%s\n--- contract\n%s", language, got, want)
			}
		})
	}
}
