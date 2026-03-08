package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"portunix.ai/portunix/src/helpers/ptx-trace/models"
)

// Index provides SQLite-based indexing for trace sessions and events
type Index struct {
	db      *sql.DB
	baseDir string
}

// NewIndex creates a new SQLite index
func NewIndex(baseDir string) (*Index, error) {
	indexDir := filepath.Join(baseDir, "index")
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create index directory: %w", err)
	}

	dbPath := filepath.Join(indexDir, "sessions.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	idx := &Index{
		db:      db,
		baseDir: baseDir,
	}

	if err := idx.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return idx, nil
}

// initSchema creates the database schema
func (idx *Index) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		status TEXT NOT NULL,
		started_at DATETIME NOT NULL,
		ended_at DATETIME,
		total_events INTEGER DEFAULT 0,
		error_count INTEGER DEFAULT 0,
		warning_count INTEGER DEFAULT 0,
		success_count INTEGER DEFAULT 0,
		tags TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		trace_id TEXT,
		parent_id TEXT,
		timestamp DATETIME NOT NULL,
		duration_us INTEGER,
		operation_type TEXT,
		operation_name TEXT NOT NULL,
		operation_category TEXT,
		level TEXT NOT NULL,
		status TEXT,
		has_error BOOLEAN DEFAULT FALSE,
		error_code TEXT,
		error_message TEXT,
		error_severity TEXT,
		tags TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE INDEX IF NOT EXISTS idx_events_session ON events(session_id);
	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_events_operation ON events(operation_name);
	CREATE INDEX IF NOT EXISTS idx_events_level ON events(level);
	CREATE INDEX IF NOT EXISTS idx_events_error ON events(has_error);
	CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
	CREATE INDEX IF NOT EXISTS idx_sessions_started ON sessions(started_at);
	`

	_, err := idx.db.Exec(schema)
	return err
}

// Close closes the database connection
func (idx *Index) Close() error {
	return idx.db.Close()
}

// IndexSession adds or updates a session in the index
func (idx *Index) IndexSession(session *models.Session) error {
	tags := strings.Join(session.Tags, ",")

	var errorCount, warningCount, successCount int64
	if session.Stats != nil {
		errorCount = session.Stats.ByStatus["error"]
		warningCount = session.Stats.ByStatus["warning"]
		successCount = session.Stats.ByStatus["success"]
	}

	totalEvents := int64(0)
	if session.Stats != nil {
		totalEvents = session.Stats.TotalEvents
	}

	_, err := idx.db.Exec(`
		INSERT INTO sessions (id, name, status, started_at, ended_at, total_events, error_count, warning_count, success_count, tags, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			status = excluded.status,
			ended_at = excluded.ended_at,
			total_events = excluded.total_events,
			error_count = excluded.error_count,
			warning_count = excluded.warning_count,
			success_count = excluded.success_count,
			tags = excluded.tags,
			updated_at = CURRENT_TIMESTAMP
	`, session.ID, session.Name, session.Status, session.StartedAt, session.EndedAt,
		totalEvents, errorCount, warningCount, successCount, tags)

	return err
}

// IndexEvent adds an event to the index
func (idx *Index) IndexEvent(event *models.TraceEvent) error {
	tags := strings.Join(event.Tags, ",")

	hasError := event.Error != nil
	var errorCode, errorMessage, errorSeverity string
	if event.Error != nil {
		errorCode = event.Error.Code
		errorMessage = event.Error.Message
		errorSeverity = string(event.Error.Severity)
	}

	status := ""
	if event.Output != nil {
		status = event.Output.Status
	}

	_, err := idx.db.Exec(`
		INSERT INTO events (id, session_id, trace_id, parent_id, timestamp, duration_us,
			operation_type, operation_name, operation_category, level, status,
			has_error, error_code, error_message, error_severity, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO NOTHING
	`, event.ID, event.SessionID, event.TraceID, event.ParentID, event.Timestamp, event.DurationUS,
		event.Operation.Type, event.Operation.Name, event.Operation.Category, event.Level, status,
		hasError, errorCode, errorMessage, errorSeverity, tags)

	return err
}

// QueryResult represents a query result row
type QueryResult struct {
	ID            string
	SessionID     string
	Timestamp     time.Time
	DurationUS    int64
	OperationType string
	OperationName string
	Level         string
	Status        string
	HasError      bool
	ErrorCode     string
	ErrorMessage  string
	Tags          string
}

// QueryEvents executes a query on events
func (idx *Index) QueryEvents(sessionID string, filter *QueryFilter) ([]*QueryResult, error) {
	query := `SELECT id, session_id, timestamp, duration_us, operation_type, operation_name,
		level, status, has_error, error_code, error_message, tags FROM events WHERE 1=1`
	args := []interface{}{}

	if sessionID != "" {
		query += " AND session_id = ?"
		args = append(args, sessionID)
	}

	if filter != nil {
		if filter.Operation != "" {
			query += " AND operation_name = ?"
			args = append(args, filter.Operation)
		}
		if filter.Level != "" {
			query += " AND level = ?"
			args = append(args, filter.Level)
		}
		if filter.Status != "" {
			query += " AND status = ?"
			args = append(args, filter.Status)
		}
		if filter.HasError {
			query += " AND has_error = TRUE"
		}
		if filter.ErrorCode != "" {
			query += " AND error_code = ?"
			args = append(args, filter.ErrorCode)
		}
		if filter.Since != nil {
			query += " AND timestamp >= ?"
			args = append(args, *filter.Since)
		}
		if filter.Until != nil {
			query += " AND timestamp <= ?"
			args = append(args, *filter.Until)
		}
		if filter.Tag != "" {
			query += " AND tags LIKE ?"
			args = append(args, "%"+filter.Tag+"%")
		}
		if filter.MinDuration > 0 {
			query += " AND duration_us >= ?"
			args = append(args, filter.MinDuration)
		}
	}

	query += " ORDER BY timestamp DESC"

	if filter != nil && filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := idx.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*QueryResult
	for rows.Next() {
		r := &QueryResult{}
		var status, errorCode, errorMessage, tags sql.NullString
		var durationUS sql.NullInt64

		err := rows.Scan(&r.ID, &r.SessionID, &r.Timestamp, &durationUS,
			&r.OperationType, &r.OperationName, &r.Level, &status,
			&r.HasError, &errorCode, &errorMessage, &tags)
		if err != nil {
			return nil, err
		}

		if durationUS.Valid {
			r.DurationUS = durationUS.Int64
		}
		if status.Valid {
			r.Status = status.String
		}
		if errorCode.Valid {
			r.ErrorCode = errorCode.String
		}
		if errorMessage.Valid {
			r.ErrorMessage = errorMessage.String
		}
		if tags.Valid {
			r.Tags = tags.String
		}

		results = append(results, r)
	}

	return results, nil
}

// QueryFilter defines filtering criteria for queries
type QueryFilter struct {
	Operation   string
	Level       string
	Status      string
	HasError    bool
	ErrorCode   string
	Tag         string
	Since       *time.Time
	Until       *time.Time
	MinDuration int64
	Limit       int
}

// ErrorGroup represents grouped errors
type ErrorGroup struct {
	ErrorCode    string
	ErrorMessage string
	Count        int
	FirstSeen    time.Time
	LastSeen     time.Time
	Operations   []string
	Severity     string
}

// GetErrorGroups returns grouped errors for a session
func (idx *Index) GetErrorGroups(sessionID string, limit int) ([]*ErrorGroup, error) {
	query := `
		SELECT error_code, error_message, error_severity, COUNT(*) as count,
			MIN(timestamp) as first_seen, MAX(timestamp) as last_seen,
			GROUP_CONCAT(DISTINCT operation_name) as operations
		FROM events
		WHERE has_error = TRUE
	`
	args := []interface{}{}

	if sessionID != "" {
		query += " AND session_id = ?"
		args = append(args, sessionID)
	}

	query += " GROUP BY error_code, error_message ORDER BY count DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := idx.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*ErrorGroup
	for rows.Next() {
		g := &ErrorGroup{}
		var operations string
		var severity sql.NullString
		var firstSeenStr, lastSeenStr string

		err := rows.Scan(&g.ErrorCode, &g.ErrorMessage, &severity, &g.Count,
			&firstSeenStr, &lastSeenStr, &operations)
		if err != nil {
			return nil, err
		}

		// Parse timestamps from SQLite string format
		g.FirstSeen, _ = time.Parse("2006-01-02T15:04:05.999999999Z07:00", firstSeenStr)
		if g.FirstSeen.IsZero() {
			g.FirstSeen, _ = time.Parse("2006-01-02 15:04:05.999999999-07:00", firstSeenStr)
		}
		g.LastSeen, _ = time.Parse("2006-01-02T15:04:05.999999999Z07:00", lastSeenStr)
		if g.LastSeen.IsZero() {
			g.LastSeen, _ = time.Parse("2006-01-02 15:04:05.999999999-07:00", lastSeenStr)
		}

		if severity.Valid {
			g.Severity = severity.String
		}
		if operations != "" {
			g.Operations = strings.Split(operations, ",")
		}

		groups = append(groups, g)
	}

	return groups, nil
}

// GetSessionStats returns statistics for a session from the index
func (idx *Index) GetSessionStats(sessionID string) (*SessionIndexStats, error) {
	var stats SessionIndexStats

	row := idx.db.QueryRow(`
		SELECT total_events, error_count, warning_count, success_count
		FROM sessions WHERE id = ?
	`, sessionID)

	err := row.Scan(&stats.TotalEvents, &stats.ErrorCount, &stats.WarningCount, &stats.SuccessCount)
	if err != nil {
		return nil, err
	}

	// Get operation stats
	rows, err := idx.db.Query(`
		SELECT operation_name, COUNT(*) as count, AVG(duration_us) as avg_duration
		FROM events WHERE session_id = ?
		GROUP BY operation_name
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats.ByOperation = make(map[string]*OperationIndexStats)
	for rows.Next() {
		var opName string
		var opStats OperationIndexStats
		if err := rows.Scan(&opName, &opStats.Count, &opStats.AvgDuration); err != nil {
			return nil, err
		}
		stats.ByOperation[opName] = &opStats
	}

	return &stats, nil
}

// SessionIndexStats contains indexed session statistics
type SessionIndexStats struct {
	TotalEvents  int64
	ErrorCount   int64
	WarningCount int64
	SuccessCount int64
	ByOperation  map[string]*OperationIndexStats
}

// OperationIndexStats contains indexed operation statistics
type OperationIndexStats struct {
	Count       int64
	AvgDuration float64
}

// RebuildSessionIndex rebuilds the index for a session from NDJSON files
func (idx *Index) RebuildSessionIndex(sessionID string, events []*models.TraceEvent, session *models.Session) error {
	tx, err := idx.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing events for this session
	_, err = tx.Exec("DELETE FROM events WHERE session_id = ?", sessionID)
	if err != nil {
		return err
	}

	// Insert events
	stmt, err := tx.Prepare(`
		INSERT INTO events (id, session_id, trace_id, parent_id, timestamp, duration_us,
			operation_type, operation_name, operation_category, level, status,
			has_error, error_code, error_message, error_severity, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, event := range events {
		tags := strings.Join(event.Tags, ",")
		hasError := event.Error != nil
		var errorCode, errorMessage, errorSeverity string
		if event.Error != nil {
			errorCode = event.Error.Code
			errorMessage = event.Error.Message
			errorSeverity = string(event.Error.Severity)
		}
		status := ""
		if event.Output != nil {
			status = event.Output.Status
		}

		_, err = stmt.Exec(event.ID, event.SessionID, event.TraceID, event.ParentID,
			event.Timestamp, event.DurationUS, event.Operation.Type, event.Operation.Name,
			event.Operation.Category, event.Level, status, hasError, errorCode,
			errorMessage, errorSeverity, tags)
		if err != nil {
			return err
		}
	}

	// Update session
	if session != nil {
		if err := idx.indexSessionTx(tx, session); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (idx *Index) indexSessionTx(tx *sql.Tx, session *models.Session) error {
	tags := strings.Join(session.Tags, ",")

	var errorCount, warningCount, successCount int64
	if session.Stats != nil {
		errorCount = session.Stats.ByStatus["error"]
		warningCount = session.Stats.ByStatus["warning"]
		successCount = session.Stats.ByStatus["success"]
	}

	totalEvents := int64(0)
	if session.Stats != nil {
		totalEvents = session.Stats.TotalEvents
	}

	_, err := tx.Exec(`
		INSERT INTO sessions (id, name, status, started_at, ended_at, total_events, error_count, warning_count, success_count, tags, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			status = excluded.status,
			ended_at = excluded.ended_at,
			total_events = excluded.total_events,
			error_count = excluded.error_count,
			warning_count = excluded.warning_count,
			success_count = excluded.success_count,
			tags = excluded.tags,
			updated_at = CURRENT_TIMESTAMP
	`, session.ID, session.Name, session.Status, session.StartedAt, session.EndedAt,
		totalEvents, errorCount, warningCount, successCount, tags)

	return err
}
