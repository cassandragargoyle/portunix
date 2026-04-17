/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package export

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/lib/pq"              // PostgreSQL driver
	"portunix.ai/portunix/src/helpers/ptx-trace/models"
	"portunix.ai/portunix/src/helpers/ptx-trace/storage"
)

// DatabaseExportOptions configures database export
type DatabaseExportOptions struct {
	ConnectionString string // Database connection string
	Driver           string // "postgres" or "mysql"
	Table            string // Target table name
	Mode             string // "insert", "upsert", or "replace"
	BatchSize        int    // Number of records per batch
	CreateTable      bool   // Auto-create table if not exists
	IncludeSession   bool   // Also export session metadata
}

// DefaultDatabaseExportOptions returns default options
func DefaultDatabaseExportOptions() *DatabaseExportOptions {
	return &DatabaseExportOptions{
		Driver:         "postgres",
		Table:          "trace_events",
		Mode:           "insert",
		BatchSize:      100,
		CreateTable:    true,
		IncludeSession: true,
	}
}

// DatabaseExporter exports trace data to SQL databases
type DatabaseExporter struct {
	storage *storage.Storage
}

// NewDatabaseExporter creates a new database exporter
func NewDatabaseExporter(store *storage.Storage) *DatabaseExporter {
	return &DatabaseExporter{
		storage: store,
	}
}

// DatabaseExportResult contains export statistics
type DatabaseExportResult struct {
	SessionID      string
	EventsExported int
	Duration       time.Duration
	Table          string
	Driver         string
}

// Export exports a session to a database
func (e *DatabaseExporter) Export(sessionID string, opts *DatabaseExportOptions) (*DatabaseExportResult, error) {
	if opts == nil {
		opts = DefaultDatabaseExportOptions()
	}

	startTime := time.Now()

	// Detect driver from connection string if not specified
	if opts.Driver == "" {
		opts.Driver = detectDriver(opts.ConnectionString)
	}

	// Connect to database
	db, err := sql.Open(opts.Driver, opts.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables if requested
	if opts.CreateTable {
		if err := e.createTables(db, opts); err != nil {
			return nil, fmt.Errorf("failed to create tables: %w", err)
		}
	}

	// Load session if needed
	if opts.IncludeSession {
		session, err := e.storage.LoadSession(sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to load session: %w", err)
		}
		if err := e.exportSession(db, session, opts); err != nil {
			return nil, fmt.Errorf("failed to export session: %w", err)
		}
	}

	// Load and export events
	events, err := e.storage.ReadEvents(sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read events: %w", err)
	}

	exported, err := e.exportEvents(db, sessionID, events, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to export events: %w", err)
	}

	return &DatabaseExportResult{
		SessionID:      sessionID,
		EventsExported: exported,
		Duration:       time.Since(startTime),
		Table:          opts.Table,
		Driver:         opts.Driver,
	}, nil
}

func (e *DatabaseExporter) createTables(db *sql.DB, opts *DatabaseExportOptions) error {
	var sessionsSQL, eventsSQL string

	if opts.Driver == "postgres" {
		sessionsSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s_sessions (
				id VARCHAR(255) PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				status VARCHAR(50) NOT NULL,
				started_at TIMESTAMP WITH TIME ZONE NOT NULL,
				ended_at TIMESTAMP WITH TIME ZONE,
				duration_ms BIGINT,
				tags TEXT[],
				source_type VARCHAR(50),
				source_path TEXT,
				destination_type VARCHAR(50),
				destination_path TEXT,
				total_events BIGINT DEFAULT 0,
				error_count BIGINT DEFAULT 0,
				config JSONB,
				stats JSONB,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			)`, opts.Table)

		eventsSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(255) PRIMARY KEY,
				session_id VARCHAR(255) NOT NULL REFERENCES %s_sessions(id),
				trace_id VARCHAR(255),
				parent_id VARCHAR(255),
				timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
				duration_us BIGINT,
				operation_type VARCHAR(100),
				operation_name VARCHAR(255) NOT NULL,
				operation_category VARCHAR(100),
				level VARCHAR(20) NOT NULL,
				input_fields JSONB,
				input_source JSONB,
				output_fields JSONB,
				output_status VARCHAR(50),
				error_code VARCHAR(100),
				error_message TEXT,
				error_severity VARCHAR(20),
				tags TEXT[],
				context JSONB,
				performance JSONB,
				metadata JSONB,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			)`, opts.Table, opts.Table)
	} else {
		// MySQL syntax
		sessionsSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s_sessions (
				id VARCHAR(255) PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				status VARCHAR(50) NOT NULL,
				started_at DATETIME NOT NULL,
				ended_at DATETIME,
				duration_ms BIGINT,
				tags JSON,
				source_type VARCHAR(50),
				source_path TEXT,
				destination_type VARCHAR(50),
				destination_path TEXT,
				total_events BIGINT DEFAULT 0,
				error_count BIGINT DEFAULT 0,
				config JSON,
				stats JSON,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`, opts.Table)

		eventsSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(255) PRIMARY KEY,
				session_id VARCHAR(255) NOT NULL,
				trace_id VARCHAR(255),
				parent_id VARCHAR(255),
				timestamp DATETIME NOT NULL,
				duration_us BIGINT,
				operation_type VARCHAR(100),
				operation_name VARCHAR(255) NOT NULL,
				operation_category VARCHAR(100),
				level VARCHAR(20) NOT NULL,
				input_fields JSON,
				input_source JSON,
				output_fields JSON,
				output_status VARCHAR(50),
				error_code VARCHAR(100),
				error_message TEXT,
				error_severity VARCHAR(20),
				tags JSON,
				context JSON,
				performance JSON,
				metadata JSON,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (session_id) REFERENCES %s_sessions(id)
			)`, opts.Table, opts.Table)
	}

	if _, err := db.Exec(sessionsSQL); err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	if _, err := db.Exec(eventsSQL); err != nil {
		return fmt.Errorf("failed to create events table: %w", err)
	}

	// Create indexes
	indexes := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_session_id ON %s(session_id)", opts.Table, opts.Table),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_timestamp ON %s(timestamp)", opts.Table, opts.Table),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_operation ON %s(operation_name)", opts.Table, opts.Table),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_level ON %s(level)", opts.Table, opts.Table),
	}

	for _, idx := range indexes {
		db.Exec(idx) // Ignore errors for index creation
	}

	return nil
}

func (e *DatabaseExporter) exportSession(db *sql.DB, session *models.Session, opts *DatabaseExportOptions) error {
	var durationMs *int64
	if session.EndedAt != nil {
		d := session.EndedAt.Sub(session.StartedAt).Milliseconds()
		durationMs = &d
	}

	tagsJSON, _ := json.Marshal(session.Tags)
	configJSON, _ := json.Marshal(session.Config)
	statsJSON, _ := json.Marshal(session.Stats)

	var sourceType, sourcePath, destType, destPath string
	if session.Source != nil {
		sourceType = session.Source.Type
		if len(session.Source.Files) > 0 {
			sourcePath = strings.Join(session.Source.Files, ",")
		}
	}
	if session.Destination != nil {
		destType = session.Destination.Type
		destPath = session.Destination.Table
	}

	var totalEvents, errorCount int64
	if session.Stats != nil {
		totalEvents = session.Stats.TotalEvents
		errorCount = int64(session.Stats.ByStatus["error"])
	}

	var query string
	if opts.Driver == "postgres" {
		if opts.Mode == "upsert" {
			query = fmt.Sprintf(`
				INSERT INTO %s_sessions (id, name, status, started_at, ended_at, duration_ms, tags,
					source_type, source_path, destination_type, destination_path,
					total_events, error_count, config, stats)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
				ON CONFLICT (id) DO UPDATE SET
					name = EXCLUDED.name,
					status = EXCLUDED.status,
					ended_at = EXCLUDED.ended_at,
					duration_ms = EXCLUDED.duration_ms,
					total_events = EXCLUDED.total_events,
					error_count = EXCLUDED.error_count,
					stats = EXCLUDED.stats`,
				opts.Table)
		} else {
			query = fmt.Sprintf(`
				INSERT INTO %s_sessions (id, name, status, started_at, ended_at, duration_ms, tags,
					source_type, source_path, destination_type, destination_path,
					total_events, error_count, config, stats)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
				ON CONFLICT (id) DO NOTHING`,
				opts.Table)
		}

		_, err := db.Exec(query,
			session.ID, session.Name, session.Status,
			session.StartedAt, session.EndedAt, durationMs,
			string(tagsJSON), sourceType, sourcePath, destType, destPath,
			totalEvents, errorCount, string(configJSON), string(statsJSON))
		return err
	}

	// MySQL
	if opts.Mode == "upsert" || opts.Mode == "replace" {
		query = fmt.Sprintf(`
			REPLACE INTO %s_sessions (id, name, status, started_at, ended_at, duration_ms, tags,
				source_type, source_path, destination_type, destination_path,
				total_events, error_count, config, stats)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			opts.Table)
	} else {
		query = fmt.Sprintf(`
			INSERT IGNORE INTO %s_sessions (id, name, status, started_at, ended_at, duration_ms, tags,
				source_type, source_path, destination_type, destination_path,
				total_events, error_count, config, stats)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			opts.Table)
	}

	_, err := db.Exec(query,
		session.ID, session.Name, session.Status,
		session.StartedAt, session.EndedAt, durationMs,
		string(tagsJSON), sourceType, sourcePath, destType, destPath,
		totalEvents, errorCount, string(configJSON), string(statsJSON))
	return err
}

func (e *DatabaseExporter) exportEvents(db *sql.DB, sessionID string, events []*models.TraceEvent, opts *DatabaseExportOptions) (int, error) {
	if len(events) == 0 {
		return 0, nil
	}

	exported := 0

	// Process in batches
	for i := 0; i < len(events); i += opts.BatchSize {
		end := i + opts.BatchSize
		if end > len(events) {
			end = len(events)
		}

		batch := events[i:end]
		if err := e.exportBatch(db, sessionID, batch, opts); err != nil {
			return exported, err
		}

		exported += len(batch)
	}

	return exported, nil
}

func (e *DatabaseExporter) exportBatch(db *sql.DB, sessionID string, events []*models.TraceEvent, opts *DatabaseExportOptions) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, event := range events {
		if err := e.insertEvent(tx, sessionID, event, opts); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (e *DatabaseExporter) insertEvent(tx *sql.Tx, sessionID string, event *models.TraceEvent, opts *DatabaseExportOptions) error {
	// Prepare JSON fields
	var inputFields, inputSource, outputFields, context, performance, metadata []byte
	var outputStatus, errorCode, errorMessage, errorSeverity string
	var tags []byte

	if event.Input != nil {
		inputFields, _ = json.Marshal(event.Input.Fields)
		inputSource, _ = json.Marshal(event.Input.Source)
	}

	if event.Output != nil {
		outputFields, _ = json.Marshal(event.Output.Fields)
		outputStatus = event.Output.Status
	}

	if event.Error != nil {
		errorCode = event.Error.Code
		errorMessage = event.Error.Message
		errorSeverity = string(event.Error.Severity)
	}

	tags, _ = json.Marshal(event.Tags)
	context, _ = json.Marshal(event.Context)
	performance, _ = json.Marshal(event.Performance)
	metadata, _ = json.Marshal(event.Metadata)

	var query string
	if opts.Driver == "postgres" {
		conflictAction := "DO NOTHING"
		if opts.Mode == "upsert" {
			conflictAction = `DO UPDATE SET
				duration_us = EXCLUDED.duration_us,
				output_fields = EXCLUDED.output_fields,
				output_status = EXCLUDED.output_status,
				error_code = EXCLUDED.error_code,
				error_message = EXCLUDED.error_message`
		}

		query = fmt.Sprintf(`
			INSERT INTO %s (id, session_id, trace_id, parent_id, timestamp, duration_us,
				operation_type, operation_name, operation_category, level,
				input_fields, input_source, output_fields, output_status,
				error_code, error_message, error_severity, tags, context, performance, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
			ON CONFLICT (id) %s`,
			opts.Table, conflictAction)

		_, err := tx.Exec(query,
			event.ID, sessionID, event.TraceID, event.ParentID, event.Timestamp, event.DurationUS,
			event.Operation.Type, event.Operation.Name, event.Operation.Category, event.Level,
			nullableJSON(inputFields), nullableJSON(inputSource), nullableJSON(outputFields), outputStatus,
			nullableString(errorCode), nullableString(errorMessage), nullableString(errorSeverity),
			nullableJSON(tags), nullableJSON(context), nullableJSON(performance), nullableJSON(metadata))
		return err
	}

	// MySQL
	verb := "INSERT IGNORE INTO"
	if opts.Mode == "replace" || opts.Mode == "upsert" {
		verb = "REPLACE INTO"
	}

	query = fmt.Sprintf(`
		%s %s (id, session_id, trace_id, parent_id, timestamp, duration_us,
			operation_type, operation_name, operation_category, level,
			input_fields, input_source, output_fields, output_status,
			error_code, error_message, error_severity, tags, context, performance, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		verb, opts.Table)

	_, err := tx.Exec(query,
		event.ID, sessionID, event.TraceID, event.ParentID, event.Timestamp, event.DurationUS,
		event.Operation.Type, event.Operation.Name, event.Operation.Category, event.Level,
		nullableJSON(inputFields), nullableJSON(inputSource), nullableJSON(outputFields), outputStatus,
		nullableString(errorCode), nullableString(errorMessage), nullableString(errorSeverity),
		nullableJSON(tags), nullableJSON(context), nullableJSON(performance), nullableJSON(metadata))
	return err
}

// Helper functions

func detectDriver(connStr string) string {
	lower := strings.ToLower(connStr)
	if strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://") {
		return "postgres"
	}
	if strings.Contains(lower, "mysql") || strings.Contains(lower, ":3306") {
		return "mysql"
	}
	return "postgres" // default
}

func nullableJSON(data []byte) interface{} {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	return string(data)
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
