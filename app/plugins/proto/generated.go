// Package proto contains generated protobuf definitions for plugin communication
// This is a placeholder until we can generate actual protobuf files
package proto

import (
	"context"
)

// Placeholder types - these should be generated from .proto files
type PluginServiceClient interface {
	Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error)
	Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
	Health(ctx context.Context, req *HealthRequest) (*HealthResponse, error)
	Shutdown(ctx context.Context, req *ShutdownRequest) (*ShutdownResponse, error)
	GetInfo(ctx context.Context, req *GetInfoRequest) (*GetInfoResponse, error)
	ListCommands(ctx context.Context, req *ListCommandsRequest) (*ListCommandsResponse, error)
}

// DatastorePluginServiceClient interface for datastore plugins
type DatastorePluginServiceClient interface {
	// Base plugin functionality
	Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error)
	Health(ctx context.Context, req *HealthRequest) (*HealthResponse, error)
	Shutdown(ctx context.Context, req *ShutdownRequest) (*ShutdownResponse, error)
	
	// Datastore operations
	Store(ctx context.Context, req *StoreRequest) (*StoreResponse, error)
	Retrieve(ctx context.Context, req *RetrieveRequest) (*RetrieveResponse, error)
	Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error)
	Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error)
	List(ctx context.Context, req *ListKeysRequest) (*ListKeysResponse, error)
	
	// Management operations
	TestConnection(ctx context.Context, req *TestConnectionRequest) (*TestConnectionResponse, error)
	GetStats(ctx context.Context, req *GetStatsRequest) (*GetStatsResponse, error)
}

type InitializeRequest struct {
	PluginName  string
	Version     string
	Config      map[string]string
	Environment map[string]string
	Permissions *PluginPermissions
}

type InitializeResponse struct {
	Success    bool
	Message    string
	PluginInfo *PluginInfo
}

type ExecuteRequest struct {
	Command          string
	Args             []string
	Options          map[string]string
	Environment      map[string]string
	WorkingDirectory string
}

type ExecuteResponse struct {
	Success  bool
	Message  string
	Output   string
	Error    string
	ExitCode int32
	Metadata map[string]string
}

type HealthRequest struct{}

type HealthResponse struct {
	Healthy        bool
	Status         string
	Message        string
	UptimeSeconds  int64
	Metrics        map[string]string
}

type ShutdownRequest struct {
	Force          bool
	TimeoutSeconds int32
}

type ShutdownResponse struct {
	Success bool
	Message string
}

type GetInfoRequest struct{}

type GetInfoResponse struct {
	PluginInfo *PluginInfo
}

type ListCommandsRequest struct{}

type ListCommandsResponse struct {
	Commands     []*PluginCommand
	Capabilities *PluginCapabilities
}

type PluginInfo struct {
	Name                string
	Version             string
	Description         string
	Author              string
	License             string
	SupportedOs         []string
	Commands            []*PluginCommand
	Capabilities        *PluginCapabilities
	RequiredPermissions *PluginPermissions
}

type PluginCommand struct {
	Name        string
	Description string
	Subcommands []string
	Parameters  []*PluginParameter
	Examples    []string
}

type PluginParameter struct {
	Name         string
	Type         string
	Description  string
	Required     bool
	DefaultValue string
}

type PluginCapabilities struct {
	FilesystemAccess bool
	NetworkAccess    bool
	DatabaseAccess   bool
	ContainerAccess  bool
	SystemCommands   bool
	McpTools         []string
}

type PluginPermissions struct {
	Filesystem []string
	Network    []string
	Database   []string
	System     []string
	Level      string
}

// Datastore operation types
type StoreRequest struct {
	Key         string
	Value       []byte
	ContentType string
	Metadata    map[string]string
	Config      map[string]string
}

type StoreResponse struct {
	Success        bool
	Message        string
	ResultMetadata map[string]string
}

type RetrieveRequest struct {
	Key    string
	Filter map[string]string
	Config map[string]string
}

type RetrieveResponse struct {
	Success     bool
	Message     string
	Value       []byte
	ContentType string
	Metadata    map[string]string
}

type QueryRequest struct {
	Collection string
	Filter     map[string]string
	Sort       map[string]string
	Limit      int32
	Offset     int32
	Config     map[string]string
}

type QueryResponse struct {
	Success    bool
	Message    string
	Results    []*QueryResult
	TotalCount int32
}

type QueryResult struct {
	Key         string
	Value       []byte
	ContentType string
	Metadata    map[string]string
}

type DeleteRequest struct {
	Key    string
	Config map[string]string
}

type DeleteResponse struct {
	Success bool
	Message string
}

type ListKeysRequest struct {
	Pattern string
	Limit   int32
	Offset  int32
	Config  map[string]string
}

type ListKeysResponse struct {
	Success    bool
	Message    string
	Keys       []string
	TotalCount int32
}

type TestConnectionRequest struct {
	Config map[string]string
}

type TestConnectionResponse struct {
	Success        bool
	Message        string
	ConnectionInfo map[string]string
}

type GetStatsRequest struct{}

type GetStatsResponse struct {
	Success bool
	Message string
	Stats   *DatastoreStats
}

type DatastoreStats struct {
	TotalKeys              int64
	TotalSizeBytes         int64
	Collections            map[string]int64
	Performance            *PerformanceMetrics
	LastUpdatedTimestamp   int64
}

type PerformanceMetrics struct {
	AverageReadTimeMs  float64
	AverageWriteTimeMs float64
	OperationsPerSec   float64
	ErrorRate          float64
}

// NewPluginServiceClient creates a new plugin service client
// This is a placeholder - should be generated by protoc
func NewPluginServiceClient(conn interface{}) PluginServiceClient {
	// This would normally be generated by protoc
	return nil
}

// NewDatastorePluginServiceClient creates a new datastore plugin service client
// This is a placeholder - should be generated by protoc
func NewDatastorePluginServiceClient(conn interface{}) DatastorePluginServiceClient {
	// This would normally be generated by protoc
	return nil
}