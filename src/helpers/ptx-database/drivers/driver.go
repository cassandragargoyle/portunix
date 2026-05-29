/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package drivers defines the Driver interface that abstracts engine-specific
// database operations (lifecycle, backup, introspection) so that core
// commands in main.go remain engine-agnostic. New engines plug in by
// implementing Driver and registering in init().
package drivers

import (
	"fmt"
	"sort"
	"sync"
)

// Mode selects native (host service) or container-based execution.
type Mode string

const (
	ModeNative    Mode = "native"
	ModeContainer Mode = "container"
	ModeEmbedded  Mode = "embedded"
)

// Status reports the runtime state of an instance.
type Status struct {
	Engine   string `json:"engine"`
	Instance string `json:"instance,omitempty"`
	Mode     Mode   `json:"mode"`
	Running  bool   `json:"running"`
	Version  string `json:"version,omitempty"`
	Port     int    `json:"port,omitempty"`
	DataDir  string `json:"data_dir,omitempty"`
	Pid      int    `json:"pid,omitempty"`
	Uptime   string `json:"uptime,omitempty"`
	Message  string `json:"message,omitempty"`
}

// Database describes a single database/schema on an engine instance.
type Database struct {
	Name    string `json:"name"`
	Owner   string `json:"owner,omitempty"`
	Size    string `json:"size,omitempty"`
	Charset string `json:"charset,omitempty"`
}

// Table describes a single table/collection.
type Table struct {
	Schema string `json:"schema,omitempty"`
	Name   string `json:"name"`
	Rows   int64  `json:"rows,omitempty"`
	Size   string `json:"size,omitempty"`
}

// SchemaDef holds raw DDL or a structured schema dump.
type SchemaDef struct {
	Format string `json:"format"` // "sql" or "json"
	Body   string `json:"body"`
}

// BackupResult describes a finished backup.
type BackupResult struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	Engine    string `json:"engine"`
	Database  string `json:"database,omitempty"`
	Compress  bool   `json:"compressed"`
	Timestamp string `json:"timestamp"`
}

// BackupOpts controls backup behaviour.
type BackupOpts struct {
	Database    string
	Destination string // dir; helper picks file name
	Compress    bool
}

// RestoreOpts controls restore behaviour.
type RestoreOpts struct {
	File     string
	TargetDB string
}

// LifecycleOpts is passed to start/stop/restart/status.
type LifecycleOpts struct {
	Instance string
	Mode     Mode
}

// HealthReport summarises an engine ping result.
type HealthReport struct {
	Engine  string `json:"engine"`
	OK      bool   `json:"ok"`
	Latency string `json:"latency_ms,omitempty"`
	Message string `json:"message,omitempty"`
}

// Driver is the engine-agnostic interface that every database driver implements.
// Drivers shell out to engine CLI tools (psql, pg_dump, sqlite3, ...). They MUST
// not exec installation tools — installation is delegated to ptx-installer by
// the top-level command, not by the driver.
//
// All methods are expected to be cheap (< 5 s) for read paths so MCP tools
// remain responsive. Long-running operations (backup, restore) may block.
type Driver interface {
	// Engine returns the canonical engine name (e.g. "postgresql", "sqlite").
	Engine() string

	// SupportedModes returns the modes this driver understands.
	SupportedModes() []Mode

	// Status returns the runtime status of the named instance.
	Status(opts LifecycleOpts) (*Status, error)

	// Start starts the named instance (no-op for embedded engines).
	Start(opts LifecycleOpts) error

	// Stop stops the named instance.
	Stop(opts LifecycleOpts) error

	// Restart restarts the named instance.
	Restart(opts LifecycleOpts) error

	// Health pings the engine and reports liveness + version.
	Health(opts LifecycleOpts) (*HealthReport, error)

	// ListDatabases enumerates databases on the instance.
	ListDatabases(opts LifecycleOpts) ([]Database, error)

	// ListTables enumerates tables in the given database.
	ListTables(opts LifecycleOpts, db string) ([]Table, error)

	// DumpSchema returns DDL/JSON schema for the given database.
	DumpSchema(opts LifecycleOpts, db string, format string) (*SchemaDef, error)

	// Backup performs a backup of the given database. Implementations decide
	// the file name based on opts.Destination + timestamp.
	Backup(opts LifecycleOpts, b BackupOpts) (*BackupResult, error)

	// Restore restores a backup file into an instance.
	Restore(opts LifecycleOpts, r RestoreOpts) error
}

// registry holds drivers keyed by engine name. Registered via init() in
// driver-specific files.
var (
	regMu    sync.RWMutex
	registry = map[string]Driver{}
)

// Register adds a driver implementation. Panics on duplicate engine names so
// build-time mistakes surface immediately.
func Register(d Driver) {
	regMu.Lock()
	defer regMu.Unlock()
	name := d.Engine()
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("ptx-database: driver %q already registered", name))
	}
	registry[name] = d
}

// Get returns the driver for engine, or an error if unknown.
func Get(engine string) (Driver, error) {
	regMu.RLock()
	defer regMu.RUnlock()
	d, ok := registry[engine]
	if !ok {
		return nil, fmt.Errorf("unknown engine %q (supported: %v)", engine, listEngines())
	}
	return d, nil
}

// List returns all registered engine names sorted alphabetically.
func List() []string {
	regMu.RLock()
	defer regMu.RUnlock()
	return listEngines()
}

func listEngines() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
