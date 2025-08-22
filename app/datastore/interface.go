package datastore

import (
	"context"
	"time"
)

// DatastoreInterface defines the unified interface for all datastore operations
type DatastoreInterface interface {
	// Core operations
	Store(ctx context.Context, key string, value interface{}, metadata map[string]interface{}) error
	Retrieve(ctx context.Context, key string, filter map[string]interface{}) (interface{}, error)
	Query(ctx context.Context, criteria QueryCriteria) ([]QueryResult, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, pattern string) ([]string, error)
	
	// Management operations
	Health(ctx context.Context) (*HealthStatus, error)
	Stats(ctx context.Context) (*Stats, error)
	
	// Plugin lifecycle
	Initialize(ctx context.Context, config map[string]interface{}) error
	Close(ctx context.Context) error
}

// QueryCriteria defines search criteria for querying data
type QueryCriteria struct {
	Collection string                 `json:"collection,omitempty"`
	Filter     map[string]interface{} `json:"filter,omitempty"`
	Sort       map[string]int         `json:"sort,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
	Offset     int                    `json:"offset,omitempty"`
}

// QueryResult represents a single result from a query
type QueryResult struct {
	Key      string                 `json:"key"`
	Value    interface{}            `json:"value"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// HealthStatus represents the health status of a datastore
type HealthStatus struct {
	Healthy   bool              `json:"healthy"`
	Status    string            `json:"status"`
	Message   string            `json:"message,omitempty"`
	Uptime    time.Duration     `json:"uptime"`
	Metrics   map[string]string `json:"metrics,omitempty"`
	LastCheck time.Time         `json:"last_check"`
}

// Stats represents statistics about a datastore
type Stats struct {
	TotalKeys      int64             `json:"total_keys"`
	TotalSize      int64             `json:"total_size_bytes"`
	Collections    map[string]int64  `json:"collections,omitempty"`
	Performance    *PerformanceStats `json:"performance,omitempty"`
	LastUpdated    time.Time         `json:"last_updated"`
}

// PerformanceStats contains performance metrics
type PerformanceStats struct {
	AverageReadTime  time.Duration `json:"average_read_time"`
	AverageWriteTime time.Duration `json:"average_write_time"`
	OperationsPerSec float64       `json:"operations_per_sec"`
	ErrorRate        float64       `json:"error_rate"`
}

// DatastoreType represents the type of datastore
type DatastoreType string

const (
	DatastoreTypeFile        DatastoreType = "file"
	DatastoreTypeMongoDB     DatastoreType = "mongodb"
	DatastoreTypePostgreSQL  DatastoreType = "postgresql"
	DatastoreTypeRedis       DatastoreType = "redis"
	DatastoreTypeElastic     DatastoreType = "elasticsearch"
	DatastoreTypeSQLite      DatastoreType = "sqlite"
)

// DatastoreCapabilities defines what operations a datastore supports
type DatastoreCapabilities struct {
	SupportsTransactions bool     `json:"supports_transactions"`
	SupportsIndexing     bool     `json:"supports_indexing"`
	SupportsAggregation  bool     `json:"supports_aggregation"`
	SupportsFullText     bool     `json:"supports_full_text"`
	SupportedDataTypes   []string `json:"supported_data_types"`
	MaxKeySize           int      `json:"max_key_size"`
	MaxValueSize         int64    `json:"max_value_size"`
}

// DatastoreInfo contains information about a datastore implementation
type DatastoreInfo struct {
	Name         string                 `json:"name"`
	Type         DatastoreType          `json:"type"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Capabilities DatastoreCapabilities  `json:"capabilities"`
	ConfigSchema map[string]interface{} `json:"config_schema,omitempty"`
}