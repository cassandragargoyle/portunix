package models

import (
	"time"
)

// SessionStatus represents the status of a trace session
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusFailed    SessionStatus = "failed"
	SessionStatusCancelled SessionStatus = "cancelled"
)

// Session represents a trace session
type Session struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	StartedAt time.Time     `json:"started_at"`
	EndedAt   *time.Time    `json:"ended_at,omitempty"`
	Status    SessionStatus `json:"status"`

	Source      *SessionSource      `json:"source,omitempty"`
	Destination *SessionDestination `json:"destination,omitempty"`

	Stats  *SessionStats  `json:"stats,omitempty"`
	Config *SessionConfig `json:"config,omitempty"`

	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SessionSource describes the data source for a session
type SessionSource struct {
	Type      string   `json:"type"`
	Files     []string `json:"files,omitempty"`
	URL       string   `json:"url,omitempty"`
	Table     string   `json:"table,omitempty"`
	TotalRows int64    `json:"total_rows,omitempty"`
}

// SessionDestination describes the data destination
type SessionDestination struct {
	Type     string `json:"type"`
	URL      string `json:"url,omitempty"`
	Table    string `json:"table,omitempty"`
	Inserted int64  `json:"inserted,omitempty"`
	Updated  int64  `json:"updated,omitempty"`
	Skipped  int64  `json:"skipped,omitempty"`
}

// SessionStats contains session statistics
type SessionStats struct {
	TotalEvents int64            `json:"total_events"`
	ByStatus    map[string]int64 `json:"by_status"`
	ByOperation map[string]*OperationStats `json:"by_operation,omitempty"`
	ByLevel     map[string]int64 `json:"by_level,omitempty"`
}

// OperationStats contains per-operation statistics
type OperationStats struct {
	Count       int64   `json:"count"`
	AvgDuration float64 `json:"avg_us"`
	MinDuration int64   `json:"min_us,omitempty"`
	MaxDuration int64   `json:"max_us,omitempty"`
	TotalErrors int64   `json:"errors,omitempty"`
}

// SessionConfig contains session configuration
type SessionConfig struct {
	SamplingRate  float64 `json:"sampling_rate"`
	PIIMasking    bool    `json:"pii_masking"`
	RetentionDays int     `json:"retention_days,omitempty"`
	ChunkSize     int     `json:"chunk_size,omitempty"`
}

// NewSession creates a new session with defaults
func NewSession(name string) *Session {
	now := time.Now().UTC()
	return &Session{
		ID:        GenerateSessionID(name),
		Name:      name,
		StartedAt: now,
		Status:    SessionStatusActive,
		Stats: &SessionStats{
			TotalEvents: 0,
			ByStatus:    make(map[string]int64),
			ByOperation: make(map[string]*OperationStats),
			ByLevel:     make(map[string]int64),
		},
		Config: &SessionConfig{
			SamplingRate:  1.0,
			PIIMasking:    false,
			RetentionDays: 30,
			ChunkSize:     10000,
		},
		Tags:     []string{},
		Metadata: make(map[string]string),
	}
}

// End marks the session as completed
func (s *Session) End(status SessionStatus) {
	now := time.Now().UTC()
	s.EndedAt = &now
	s.Status = status
}

// AddTag adds a tag to the session
func (s *Session) AddTag(tag string) *Session {
	s.Tags = append(s.Tags, tag)
	return s
}

// SetSource sets the data source
func (s *Session) SetSource(sourceType string, files []string) *Session {
	s.Source = &SessionSource{
		Type:  sourceType,
		Files: files,
	}
	return s
}

// SetDestination sets the data destination
func (s *Session) SetDestination(destType, url, table string) *Session {
	s.Destination = &SessionDestination{
		Type:  destType,
		URL:   url,
		Table: table,
	}
	return s
}

// IncrementEventCount increments the event counters
func (s *Session) IncrementEventCount(status string, level Level, operation string, durationUS int64) {
	if s.Stats == nil {
		s.Stats = &SessionStats{}
	}

	// Initialize maps if nil (can happen after JSON deserialization)
	if s.Stats.ByStatus == nil {
		s.Stats.ByStatus = make(map[string]int64)
	}
	if s.Stats.ByOperation == nil {
		s.Stats.ByOperation = make(map[string]*OperationStats)
	}
	if s.Stats.ByLevel == nil {
		s.Stats.ByLevel = make(map[string]int64)
	}

	s.Stats.TotalEvents++
	s.Stats.ByStatus[status]++
	s.Stats.ByLevel[string(level)]++

	if operation != "" {
		if s.Stats.ByOperation[operation] == nil {
			s.Stats.ByOperation[operation] = &OperationStats{}
		}
		opStats := s.Stats.ByOperation[operation]
		opStats.Count++

		// Update average duration
		if durationUS > 0 {
			opStats.AvgDuration = ((opStats.AvgDuration * float64(opStats.Count-1)) + float64(durationUS)) / float64(opStats.Count)
			if opStats.MinDuration == 0 || durationUS < opStats.MinDuration {
				opStats.MinDuration = durationUS
			}
			if durationUS > opStats.MaxDuration {
				opStats.MaxDuration = durationUS
			}
		}

		if level == LevelError {
			opStats.TotalErrors++
		}
	}
}

// Duration returns the session duration
func (s *Session) Duration() time.Duration {
	if s.EndedAt != nil {
		return s.EndedAt.Sub(s.StartedAt)
	}
	return time.Since(s.StartedAt)
}
