/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package mcp

import (
	"reflect"
	"testing"
)

// Tests for PTX-PFT MCP tool argument builders (Issue #111).
// Builders translate MCP tool arguments into ptx-pft CLI arguments
// without executing the binary, so they are tested in isolation.

func TestBuildPftListArgs(t *testing.T) {
	tests := []struct {
		name string
		args map[string]interface{}
		want []string
	}{
		{
			name: "no arguments",
			args: map[string]interface{}{},
			want: []string{"pft", "list"},
		},
		{
			name: "voice voc with json format",
			args: map[string]interface{}{"voice": "voc", "format": "json"},
			want: []string{"pft", "list", "--voc", "--format", "json"},
		},
		{
			name: "voice all maps to no flag",
			args: map[string]interface{}{"voice": "all"},
			want: []string{"pft", "list"},
		},
		{
			name: "all filters combined",
			args: map[string]interface{}{
				"voice":         "vos",
				"all":           true,
				"category":      "ux",
				"uncategorized": true,
				"path":          "/docs",
			},
			want: []string{"pft", "list", "--vos", "--all", "--category", "ux", "--uncategorized", "--path", "/docs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPftListArgs(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildPftShowArgs(t *testing.T) {
	got, err := buildPftShowArgs(map[string]interface{}{"id": "UC001", "path": "/docs"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"pft", "show", "UC001", "--path", "/docs"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if _, err := buildPftShowArgs(map[string]interface{}{}); err == nil {
		t.Error("expected error for missing id")
	}
}

func TestBuildPftSyncArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		want    []string
		wantErr bool
	}{
		{
			name: "default bidirectional",
			args: map[string]interface{}{},
			want: []string{"pft", "sync"},
		},
		{
			name: "pull voc with dry run",
			args: map[string]interface{}{"direction": "pull", "voice": "voc", "dry_run": true},
			want: []string{"pft", "pull", "--voc", "--dry-run"},
		},
		{
			name: "push vos",
			args: map[string]interface{}{"direction": "push", "voice": "vos"},
			want: []string{"pft", "push", "--vos"},
		},
		{
			name:    "invalid direction",
			args:    map[string]interface{}{"direction": "sideways"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPftSyncArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildPftLinkArgs(t *testing.T) {
	got, err := buildPftLinkArgs(map[string]interface{}{"feedback_id": "UC001", "issue_id": "#107"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"pft", "link", "UC001", "#107"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if _, err := buildPftLinkArgs(map[string]interface{}{"feedback_id": "UC001"}); err == nil {
		t.Error("expected error for missing issue_id")
	}
	if _, err := buildPftLinkArgs(map[string]interface{}{"issue_id": "#107"}); err == nil {
		t.Error("expected error for missing feedback_id")
	}
}

func TestBuildPftReportArgs(t *testing.T) {
	got, err := buildPftReportArgs(map[string]interface{}{"type": "detailed", "output": "report.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"pft", "report", "--type", "detailed", "--output", "report.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBuildPftExportArgs(t *testing.T) {
	got, err := buildPftExportArgs(map[string]interface{}{"format": "md", "voice": "voc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"pft", "export", "--format", "md", "--voc"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if _, err := buildPftExportArgs(map[string]interface{}{}); err == nil {
		t.Error("expected error for missing format")
	}
}

func TestBuildPftNotifyArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		want    []string
		wantErr bool
	}{
		{
			name: "specific user",
			args: map[string]interface{}{
				"feedback_id":       "UC001",
				"notification_type": "vote",
				"recipients":        "specific",
				"email":             "user@example.com",
			},
			want: []string{"pft", "notify", "UC001", "--type", "vote", "--user", "user@example.com"},
		},
		{
			name: "all voc with dry run",
			args: map[string]interface{}{
				"feedback_id":       "UC001",
				"notification_type": "status_change",
				"recipients":        "all-voc",
				"dry_run":           true,
			},
			want: []string{"pft", "notify", "UC001", "--type", "status_change", "--all-voc", "--dry-run"},
		},
		{
			name:    "missing email for specific recipients",
			args:    map[string]interface{}{"feedback_id": "UC001", "notification_type": "vote"},
			wantErr: true,
		},
		{
			name:    "missing notification_type",
			args:    map[string]interface{}{"feedback_id": "UC001"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPftNotifyArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildPftUserListArgs(t *testing.T) {
	got, err := buildPftUserListArgs(map[string]interface{}{"category": "voc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"pft", "user", "list", "--voc"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	got, err = buildPftUserListArgs(map[string]interface{}{"category": "all"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want = []string{"pft", "user", "list"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if _, err := buildPftUserListArgs(map[string]interface{}{"category": "customer"}); err == nil {
		t.Error("expected error for invalid category")
	}
}

func TestBuildPftUserAddArgs(t *testing.T) {
	got, err := buildPftUserAddArgs(map[string]interface{}{
		"id":   "jane@acme.com",
		"name": "Jane Smith",
		"org":  "Acme Corp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"pft", "user", "add", "--id", "jane@acme.com", "--name", "Jane Smith", "--org", "Acme Corp"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if _, err := buildPftUserAddArgs(map[string]interface{}{"id": "jane@acme.com"}); err == nil {
		t.Error("expected error for missing name")
	}
	if _, err := buildPftUserAddArgs(map[string]interface{}{"name": "Jane Smith"}); err == nil {
		t.Error("expected error for missing id")
	}
}

func TestBuildPftUserUpdateArgs(t *testing.T) {
	got, err := buildPftUserUpdateArgs(map[string]interface{}{"id": "jane@acme.com", "name": "Jane Doe"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"pft", "user", "update", "jane@acme.com", "--name", "Jane Doe"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if _, err := buildPftUserUpdateArgs(map[string]interface{}{"id": "jane@acme.com"}); err == nil {
		t.Error("expected error when neither name nor org is provided")
	}
}

func TestBuildPftUserRoleArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		want    []string
		wantErr bool
	}{
		{
			name: "assign role",
			args: map[string]interface{}{"id": "john@example.com", "category": "vos", "role": "developer"},
			want: []string{"pft", "user", "role", "john@example.com", "--vos", "developer"},
		},
		{
			name: "assign proxy role",
			args: map[string]interface{}{"id": "john@example.com", "category": "vos", "role": "cio", "proxy": true},
			want: []string{"pft", "user", "role", "john@example.com", "--vos", "cio", "--proxy"},
		},
		{
			name: "remove role",
			args: map[string]interface{}{"id": "john@example.com", "category": "voc", "remove": true},
			want: []string{"pft", "user", "role", "john@example.com", "--voc", "--remove"},
		},
		{
			name:    "missing role without remove",
			args:    map[string]interface{}{"id": "john@example.com", "category": "voc"},
			wantErr: true,
		},
		{
			name:    "invalid category",
			args:    map[string]interface{}{"id": "john@example.com", "category": "admin", "role": "x"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPftUserRoleArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildPftUserLinkArgs(t *testing.T) {
	got, err := buildPftUserLinkArgs(map[string]interface{}{"id": "jane@acme.com", "external_id": "42"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"pft", "user", "link", "jane@acme.com", "--fider", "42"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	// Explicit fider provider is accepted
	if _, err := buildPftUserLinkArgs(map[string]interface{}{
		"id": "jane@acme.com", "external_id": "42", "provider": "fider",
	}); err != nil {
		t.Errorf("unexpected error for fider provider: %v", err)
	}

	// Unsupported provider is rejected with a clear message
	if _, err := buildPftUserLinkArgs(map[string]interface{}{
		"id": "jane@acme.com", "external_id": "42", "provider": "clearflask",
	}); err == nil {
		t.Error("expected error for unsupported provider")
	}
}

func TestBuildPftUserRemoveArgs(t *testing.T) {
	got, err := buildPftUserRemoveArgs(map[string]interface{}{"id": "jane@acme.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"pft", "user", "remove", "jane@acme.com"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if _, err := buildPftUserRemoveArgs(map[string]interface{}{}); err == nil {
		t.Error("expected error for missing id")
	}
}

func TestBuildPftRoleListArgs(t *testing.T) {
	got, err := buildPftRoleListArgs(map[string]interface{}{"category": "voe", "path": "/docs"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"pft", "role", "list", "--voe", "--path", "/docs"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPftToolBuildersCoverAllToolDefinitions(t *testing.T) {
	// Every advertised tool must have a registered builder and vice versa
	definitions := pftToolDefinitions()
	defined := make(map[string]bool, len(definitions))
	for _, tool := range definitions {
		name, ok := tool["name"].(string)
		if !ok || name == "" {
			t.Fatalf("tool definition without a name: %v", tool)
		}
		defined[name] = true
		if _, exists := pftToolBuilders[name]; !exists {
			t.Errorf("tool %q is advertised but has no builder", name)
		}
	}
	for name := range pftToolBuilders {
		if !defined[name] {
			t.Errorf("builder %q has no tool definition", name)
		}
	}
}

func TestPftResourceCommand(t *testing.T) {
	tests := []struct {
		uri      string
		want     []string
		wantMime string
		wantErr  bool
	}{
		{uri: "pft://config", want: []string{"pft", "configure", "--show"}, wantMime: "text/plain"},
		{uri: "pft://cache", want: []string{"pft", "cache", "status"}, wantMime: "text/plain"},
		{uri: "pft://users", want: []string{"pft", "user", "list"}, wantMime: "text/plain"},
		{uri: "pft://roles", want: []string{"pft", "role", "list"}, wantMime: "text/plain"},
		{uri: "pft://voc/UC001", want: []string{"pft", "show", "UC001"}, wantMime: "text/markdown"},
		{uri: "pft://vos/REQ002", want: []string{"pft", "show", "REQ002"}, wantMime: "text/markdown"},
		{uri: "pft://voc/", wantErr: true},
		{uri: "pft://unknown", wantErr: true},
		{uri: "http://example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			got, mime, err := pftResourceCommand(tt.uri)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
			if mime != tt.wantMime {
				t.Errorf("got mime %q, want %q", mime, tt.wantMime)
			}
		})
	}
}

func TestHandlePftToolCallUnknownTool(t *testing.T) {
	s := &Server{}
	if _, err := s.handlePftToolCall("pft_nonexistent", nil); err == nil {
		t.Error("expected error for unknown pft tool")
	}
}
