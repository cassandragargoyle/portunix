/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package server

import (
	"context"
	"encoding/json"
	"testing"

	"portunix.ai/app/plugins"
	"portunix.ai/app/plugins/manager"
	pb "portunix.ai/app/plugins/proto/pluginregistry"
)

// stubRegistry is a minimal RegistryQuerier that returns a canned matcher
// output. It lets us test the proto mapping without standing up a filesystem
// registry.
type stubRegistry struct {
	lastReq struct {
		platform        string
		platformVersion string
		features        []string
	}
	result []manager.MatchedPlugin
	err    error
}

func (s *stubRegistry) ListPluginsForPlatform(platform, version string, features []string) ([]manager.MatchedPlugin, error) {
	s.lastReq.platform = platform
	s.lastReq.platformVersion = version
	s.lastReq.features = features
	return s.result, s.err
}

func TestListPluginsForPlatform_HappyPath(t *testing.T) {
	payload := `{"surfaces":{"connector.command":true}}`
	stub := &stubRegistry{
		result: []manager.MatchedPlugin{
			{
				Plugin: &manager.RegistryPlugin{
					Name:        "erp-connector",
					Version:     "1.2.0",
					Description: "ERP bridge",
					Author:      "acme",
					License:     "MIT",
				},
				Platform: plugins.SupportedPlatform{
					Name:            "synapse",
					MinVersion:      "0.3.0",
					MaxVersion:      "0.9.9",
					Features:        []string{"connector.command", "assistant.chat"},
					PlatformPayload: json.RawMessage(payload),
				},
				MatchedFeatures: []string{"connector.command"},
			},
		},
	}

	srv := NewPluginRegistryServer(stub, "test-1.0.0")
	resp, err := srv.ListPluginsForPlatform(context.Background(), &pb.ListPluginsForPlatformRequest{
		Platform:        "synapse",
		PlatformVersion: "0.5.0",
		Features:        []string{"connector.command"},
	})
	if err != nil {
		t.Fatalf("ListPluginsForPlatform: %v", err)
	}
	if len(resp.Plugins) != 1 {
		t.Fatalf("got %d plugins, want 1", len(resp.Plugins))
	}
	m := resp.Plugins[0]
	if m.Name != "erp-connector" || m.Version != "1.2.0" {
		t.Errorf("identity mismatch: name=%q version=%q", m.Name, m.Version)
	}
	if m.MinVersion != "0.3.0" || m.MaxVersion != "0.9.9" {
		t.Errorf("range mismatch: min=%q max=%q", m.MinVersion, m.MaxVersion)
	}
	if len(m.DeclaredFeatures) != 2 {
		t.Errorf("declared_features = %v, want 2 entries", m.DeclaredFeatures)
	}
	if len(m.MatchedFeatures) != 1 || m.MatchedFeatures[0] != "connector.command" {
		t.Errorf("matched_features = %v, want [connector.command]", m.MatchedFeatures)
	}
	if string(m.PlatformPayloadJson) != payload {
		t.Errorf("platform_payload not byte-identical:\n got = %q\nwant = %q", string(m.PlatformPayloadJson), payload)
	}

	if stub.lastReq.platform != "synapse" || stub.lastReq.platformVersion != "0.5.0" {
		t.Errorf("request not forwarded: %+v", stub.lastReq)
	}
}

func TestListPluginsForPlatform_EmptyPlatformRejected(t *testing.T) {
	srv := NewPluginRegistryServer(&stubRegistry{}, "dev")
	if _, err := srv.ListPluginsForPlatform(context.Background(), &pb.ListPluginsForPlatformRequest{}); err == nil {
		t.Error("expected error for empty platform")
	}
}

func TestHealthCheck(t *testing.T) {
	srv := NewPluginRegistryServer(&stubRegistry{}, "v1.2.3")
	resp, err := srv.HealthCheck(context.Background(), &pb.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	if !resp.Healthy || resp.Version != "v1.2.3" {
		t.Errorf("unexpected HealthCheck response: %+v", resp)
	}
}

func TestListPluginsForPlatform_NilPayloadOmitted(t *testing.T) {
	stub := &stubRegistry{
		result: []manager.MatchedPlugin{
			{
				Plugin:   &manager.RegistryPlugin{Name: "plain", Version: "0.1.0"},
				Platform: plugins.SupportedPlatform{Name: "pack"},
			},
		},
	}
	srv := NewPluginRegistryServer(stub, "dev")
	resp, err := srv.ListPluginsForPlatform(context.Background(), &pb.ListPluginsForPlatformRequest{Platform: "pack"})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(resp.Plugins) != 1 {
		t.Fatalf("got %d, want 1", len(resp.Plugins))
	}
	if len(resp.Plugins[0].PlatformPayloadJson) != 0 {
		t.Errorf("expected empty payload bytes, got %q", string(resp.Plugins[0].PlatformPayloadJson))
	}
}
