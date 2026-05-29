/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package server implements the ptx-side gRPC services that expose the plugin
// registry to hosting platforms (Synapse, Pack, Agent, ...). See issue #175
// and DEC-1 / DEC-2 in docs/issues/internal/175-plugin-platform-capability-query.md
// for the architectural rationale.
package server

import (
	"context"
	"fmt"

	"portunix.ai/app/plugins/manager"

	pb "portunix.ai/app/plugins/proto/pluginregistry"
)

// RegistryQuerier is the subset of manager.Manager (or manager.Registry) that
// the gRPC server needs. A narrow interface keeps the server testable with an
// in-memory fake and decouples transport from persistence.
type RegistryQuerier interface {
	ListPluginsForPlatform(platformName, platformVersion string, requiredFeatures []string) ([]manager.MatchedPlugin, error)
}

// PluginRegistryServer serves the PluginRegistryService defined in
// plugin_registry.proto. Matching is delegated to the registry so the CLI and
// gRPC paths yield the same dataset (DEC-5 — single source of truth).
type PluginRegistryServer struct {
	pb.UnimplementedPluginRegistryServiceServer
	reg     RegistryQuerier
	version string
}

// NewPluginRegistryServer constructs a server bound to the given registry and
// reports the given build version on HealthCheck.
func NewPluginRegistryServer(reg RegistryQuerier, version string) *PluginRegistryServer {
	return &PluginRegistryServer{reg: reg, version: version}
}

// ListPluginsForPlatform implements the core discovery RPC. The platform_payload
// is forwarded as raw JSON bytes to preserve byte-identical round-trip (DEC-7).
func (s *PluginRegistryServer) ListPluginsForPlatform(
	ctx context.Context,
	req *pb.ListPluginsForPlatformRequest,
) (*pb.ListPluginsForPlatformResponse, error) {
	if req == nil || req.GetPlatform() == "" {
		return nil, fmt.Errorf("platform is required")
	}

	matches, err := s.reg.ListPluginsForPlatform(req.GetPlatform(), req.GetPlatformVersion(), req.GetFeatures())
	if err != nil {
		return nil, err
	}

	resp := &pb.ListPluginsForPlatformResponse{
		Plugins: make([]*pb.MatchedPlugin, 0, len(matches)),
	}
	for _, m := range matches {
		resp.Plugins = append(resp.Plugins, &pb.MatchedPlugin{
			Name:                m.Plugin.Name,
			Version:             m.Plugin.Version,
			Description:         m.Plugin.Description,
			Author:              m.Plugin.Author,
			License:             m.Plugin.License,
			MinVersion:          m.Platform.MinVersion,
			MaxVersion:          m.Platform.MaxVersion,
			DeclaredFeatures:    append([]string(nil), m.Platform.Features...),
			MatchedFeatures:     append([]string(nil), m.MatchedFeatures...),
			PlatformPayloadJson: append([]byte(nil), m.Platform.PlatformPayload...),
		})
	}
	return resp, nil
}

// HealthCheck returns a minimal readiness signal used by Synapse (and tests)
// to confirm the daemon started before issuing the first query.
func (s *PluginRegistryServer) HealthCheck(
	ctx context.Context,
	req *pb.HealthCheckRequest,
) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Healthy: true,
		Status:  "ok",
		Version: s.version,
	}, nil
}
