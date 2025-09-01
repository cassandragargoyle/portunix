package datastore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"portunix.cz/app/plugins/proto"
)

// PluginDatastore wraps a gRPC datastore plugin client to implement DatastoreInterface
type PluginDatastore struct {
	client     proto.DatastorePluginServiceClient
	pluginName string
	pluginInfo *DatastoreInfo
	startTime  time.Time
}

// NewPluginDatastore creates a new plugin datastore wrapper
func NewPluginDatastore(client proto.DatastorePluginServiceClient, pluginName string) *PluginDatastore {
	return &PluginDatastore{
		client:     client,
		pluginName: pluginName,
		startTime:  time.Now(),
	}
}

// Initialize initializes the plugin datastore
func (p *PluginDatastore) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Convert config to string map for gRPC
	stringConfig := make(map[string]string)
	for k, v := range config {
		if str, ok := v.(string); ok {
			stringConfig[k] = str
		} else {
			// Convert complex types to JSON
			if jsonBytes, err := json.Marshal(v); err == nil {
				stringConfig[k] = string(jsonBytes)
			}
		}
	}

	req := &proto.InitializeRequest{
		PluginName: p.pluginName,
		Config:     stringConfig,
	}

	resp, err := p.client.Initialize(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("plugin initialization failed: %s", resp.Message)
	}

	// Extract plugin info if available
	if resp.PluginInfo != nil {
		p.pluginInfo = &DatastoreInfo{
			Name:        resp.PluginInfo.Name,
			Version:     resp.PluginInfo.Version,
			Description: resp.PluginInfo.Description,
			// TODO: Map other fields from PluginInfo to DatastoreInfo
		}
	}

	return nil
}

// Store stores data using the plugin
func (p *PluginDatastore) Store(ctx context.Context, key string, value interface{}, metadata map[string]interface{}) error {
	// Serialize value
	valueBytes, contentType, err := p.serializeValue(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	// Convert metadata to string map
	stringMetadata := p.convertMetadataToStrings(metadata)

	req := &proto.StoreRequest{
		Key:         key,
		Value:       valueBytes,
		ContentType: contentType,
		Metadata:    stringMetadata,
	}

	resp, err := p.client.Store(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to store data: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("store operation failed: %s", resp.Message)
	}

	return nil
}

// Retrieve retrieves data using the plugin
func (p *PluginDatastore) Retrieve(ctx context.Context, key string, filter map[string]interface{}) (interface{}, error) {
	// Convert filter to string map
	stringFilter := p.convertMetadataToStrings(filter)

	req := &proto.RetrieveRequest{
		Key:    key,
		Filter: stringFilter,
	}

	resp, err := p.client.Retrieve(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("retrieve operation failed: %s", resp.Message)
	}

	// Deserialize value
	value, err := p.deserializeValue(resp.Value, resp.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize value: %w", err)
	}

	return value, nil
}

// Query queries data using the plugin
func (p *PluginDatastore) Query(ctx context.Context, criteria QueryCriteria) ([]QueryResult, error) {
	// Convert criteria to gRPC request
	stringFilter := p.convertMetadataToStrings(criteria.Filter)
	stringSort := make(map[string]string)
	for k, v := range criteria.Sort {
		stringSort[k] = fmt.Sprintf("%d", v)
	}

	req := &proto.QueryRequest{
		Collection: criteria.Collection,
		Filter:     stringFilter,
		Sort:       stringSort,
		Limit:      int32(criteria.Limit),
		Offset:     int32(criteria.Offset),
	}

	resp, err := p.client.Query(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query data: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("query operation failed: %s", resp.Message)
	}

	// Convert results
	var results []QueryResult
	for _, grpcResult := range resp.Results {
		value, err := p.deserializeValue(grpcResult.Value, grpcResult.ContentType)
		if err != nil {
			// Skip results that can't be deserialized
			continue
		}

		result := QueryResult{
			Key:      grpcResult.Key,
			Value:    value,
			Metadata: p.convertStringsToMetadata(grpcResult.Metadata),
		}
		results = append(results, result)
	}

	return results, nil
}

// Delete deletes data using the plugin
func (p *PluginDatastore) Delete(ctx context.Context, key string) error {
	req := &proto.DeleteRequest{
		Key: key,
	}

	resp, err := p.client.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete data: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("delete operation failed: %s", resp.Message)
	}

	return nil
}

// List lists keys using the plugin
func (p *PluginDatastore) List(ctx context.Context, pattern string) ([]string, error) {
	req := &proto.ListKeysRequest{
		Pattern: pattern,
	}

	resp, err := p.client.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("list operation failed: %s", resp.Message)
	}

	return resp.Keys, nil
}

// Health returns health status from the plugin
func (p *PluginDatastore) Health(ctx context.Context) (*HealthStatus, error) {
	req := &proto.HealthRequest{}

	resp, err := p.client.Health(ctx, req)
	if err != nil {
		return &HealthStatus{
			Healthy:   false,
			Status:    "error",
			Message:   err.Error(),
			Uptime:    time.Since(p.startTime),
			LastCheck: time.Now(),
		}, nil
	}

	return &HealthStatus{
		Healthy:   resp.Healthy,
		Status:    resp.Status,
		Message:   resp.Message,
		Uptime:    time.Duration(resp.UptimeSeconds) * time.Second,
		Metrics:   resp.Metrics,
		LastCheck: time.Now(),
	}, nil
}

// Stats returns statistics from the plugin
func (p *PluginDatastore) Stats(ctx context.Context) (*Stats, error) {
	req := &proto.GetStatsRequest{}

	resp, err := p.client.GetStats(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("get stats operation failed: %s", resp.Message)
	}

	if resp.Stats == nil {
		return &Stats{}, nil
	}

	stats := &Stats{
		TotalKeys:   resp.Stats.TotalKeys,
		TotalSize:   resp.Stats.TotalSizeBytes,
		Collections: resp.Stats.Collections,
		LastUpdated: time.Unix(resp.Stats.LastUpdatedTimestamp, 0),
	}

	if resp.Stats.Performance != nil {
		stats.Performance = &PerformanceStats{
			AverageReadTime:  time.Duration(resp.Stats.Performance.AverageReadTimeMs) * time.Millisecond,
			AverageWriteTime: time.Duration(resp.Stats.Performance.AverageWriteTimeMs) * time.Millisecond,
			OperationsPerSec: resp.Stats.Performance.OperationsPerSec,
			ErrorRate:        resp.Stats.Performance.ErrorRate,
		}
	}

	return stats, nil
}

// Close closes the connection to the plugin
func (p *PluginDatastore) Close(ctx context.Context) error {
	req := &proto.ShutdownRequest{}

	resp, err := p.client.Shutdown(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to shutdown plugin: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("plugin shutdown failed: %s", resp.Message)
	}

	return nil
}

// Helper methods

func (p *PluginDatastore) serializeValue(value interface{}) ([]byte, string, error) {
	switch v := value.(type) {
	case string:
		return []byte(v), "text/plain", nil
	case []byte:
		return v, "application/octet-stream", nil
	default:
		// Serialize as JSON
		data, err := json.Marshal(value)
		if err != nil {
			return nil, "", fmt.Errorf("failed to JSON marshal value: %w", err)
		}
		return data, "application/json", nil
	}
}

func (p *PluginDatastore) deserializeValue(data []byte, contentType string) (interface{}, error) {
	switch contentType {
	case "text/plain":
		return string(data), nil
	case "application/octet-stream":
		return data, nil
	case "application/json":
		var value interface{}
		if err := json.Unmarshal(data, &value); err != nil {
			return nil, fmt.Errorf("failed to JSON unmarshal value: %w", err)
		}
		return value, nil
	default:
		// Default to raw bytes
		return data, nil
	}
}

func (p *PluginDatastore) convertMetadataToStrings(metadata map[string]interface{}) map[string]string {
	if metadata == nil {
		return nil
	}

	result := make(map[string]string)
	for k, v := range metadata {
		if str, ok := v.(string); ok {
			result[k] = str
		} else {
			// Convert to JSON string
			if jsonBytes, err := json.Marshal(v); err == nil {
				result[k] = string(jsonBytes)
			}
		}
	}
	return result
}

func (p *PluginDatastore) convertStringsToMetadata(stringMap map[string]string) map[string]interface{} {
	if stringMap == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range stringMap {
		// Try to parse as JSON first
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(v), &jsonValue); err == nil {
			result[k] = jsonValue
		} else {
			// Fall back to string
			result[k] = v
		}
	}
	return result
}
