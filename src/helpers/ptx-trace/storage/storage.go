/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"portunix.ai/portunix/src/helpers/ptx-trace/models"
)

const (
	// DefaultChunkSize is the default number of events per chunk file
	DefaultChunkSize = 10000
	// DefaultBaseDir is the default trace storage directory
	DefaultBaseDir = ".portunix/trace"
)

// Storage handles trace data persistence
type Storage struct {
	baseDir   string
	chunkSize int
	mu        sync.Mutex
}

// NewStorage creates a new storage instance
func NewStorage() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, DefaultBaseDir)
	return NewStorageWithPath(baseDir)
}

// NewStorageWithPath creates a storage instance with custom path
func NewStorageWithPath(baseDir string) (*Storage, error) {
	// Create base directories
	dirs := []string{
		baseDir,
		filepath.Join(baseDir, "sessions"),
		filepath.Join(baseDir, "index"),
		filepath.Join(baseDir, "exports"),
		filepath.Join(baseDir, "config"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return &Storage{
		baseDir:   baseDir,
		chunkSize: DefaultChunkSize,
	}, nil
}

// SessionDir returns the directory path for a session
func (s *Storage) SessionDir(sessionID string) string {
	return filepath.Join(s.baseDir, "sessions", sessionID)
}

// CreateSession creates a new session directory and manifest
func (s *Storage) CreateSession(session *models.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionDir := s.SessionDir(session.ID)

	// Create session directories
	dirs := []string{
		sessionDir,
		filepath.Join(sessionDir, "chunks"),
		filepath.Join(sessionDir, "index"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write manifest
	if err := s.writeJSON(filepath.Join(sessionDir, "manifest.json"), session); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// UpdateSession updates the session manifest
func (s *Storage) UpdateSession(session *models.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionDir := s.SessionDir(session.ID)
	return s.writeJSON(filepath.Join(sessionDir, "manifest.json"), session)
}

// LoadSession loads a session from disk
func (s *Storage) LoadSession(sessionID string) (*models.Session, error) {
	manifestPath := filepath.Join(s.SessionDir(sessionID), "manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var session models.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &session, nil
}

// ListSessions returns all sessions
func (s *Storage) ListSessions() ([]*models.Session, error) {
	sessionsDir := filepath.Join(s.baseDir, "sessions")

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.Session{}, nil
		}
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []*models.Session
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		session, err := s.LoadSession(entry.Name())
		if err != nil {
			// Skip invalid sessions
			continue
		}
		sessions = append(sessions, session)
	}

	// Sort by start time, newest first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	})

	return sessions, nil
}

// GetActiveSession returns the currently active session, if any
func (s *Storage) GetActiveSession() (*models.Session, error) {
	sessions, err := s.ListSessions()
	if err != nil {
		return nil, err
	}

	for _, session := range sessions {
		if session.Status == models.SessionStatusActive {
			return session, nil
		}
	}

	return nil, nil
}

// WriteEvent writes an event to the session's chunk file
func (s *Storage) WriteEvent(sessionID string, event *models.TraceEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chunkDir := filepath.Join(s.SessionDir(sessionID), "chunks")

	// Find current chunk or create new one
	chunkFile, err := s.getCurrentChunkFile(chunkDir)
	if err != nil {
		return fmt.Errorf("failed to get chunk file: %w", err)
	}

	// Append event as NDJSON
	f, err := os.OpenFile(chunkFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open chunk file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
}

// getCurrentChunkFile returns the current chunk file path, creating new one if needed
func (s *Storage) getCurrentChunkFile(chunkDir string) (string, error) {
	entries, err := os.ReadDir(chunkDir)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	// Find the latest chunk
	var latestChunk string
	var latestNum int

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ndjson") {
			continue
		}

		var num int
		if _, err := fmt.Sscanf(entry.Name(), "%06d.ndjson", &num); err == nil {
			if num > latestNum {
				latestNum = num
				latestChunk = filepath.Join(chunkDir, entry.Name())
			}
		}
	}

	// Check if we need a new chunk
	if latestChunk != "" {
		lineCount, err := s.countLines(latestChunk)
		if err != nil {
			return "", err
		}

		if lineCount < s.chunkSize {
			return latestChunk, nil
		}

		// Need a new chunk
		latestNum++
	} else {
		latestNum = 1
	}

	return filepath.Join(chunkDir, fmt.Sprintf("%06d.ndjson", latestNum)), nil
}

// countLines counts the number of lines in a file
func (s *Storage) countLines(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	count := 0
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	return count, nil
}

// ReadEvents reads events from a session with optional filtering
func (s *Storage) ReadEvents(sessionID string, filter *EventFilter) ([]*models.TraceEvent, error) {
	chunkDir := filepath.Join(s.SessionDir(sessionID), "chunks")

	entries, err := os.ReadDir(chunkDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.TraceEvent{}, nil
		}
		return nil, fmt.Errorf("failed to read chunks directory: %w", err)
	}

	var events []*models.TraceEvent

	// Sort entries by name to ensure chronological order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ndjson") {
			continue
		}

		chunkPath := filepath.Join(chunkDir, entry.Name())
		chunkEvents, err := s.readChunkFile(chunkPath, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk %s: %w", entry.Name(), err)
		}

		events = append(events, chunkEvents...)

		// Check limit
		if filter != nil && filter.Limit > 0 && len(events) >= filter.Limit {
			events = events[:filter.Limit]
			break
		}
	}

	return events, nil
}

// readChunkFile reads and filters events from a single chunk file
func (s *Storage) readChunkFile(path string, filter *EventFilter) ([]*models.TraceEvent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var events []*models.TraceEvent
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var event models.TraceEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Skip invalid lines
			continue
		}

		// Apply filter
		if filter != nil && !filter.Matches(&event) {
			continue
		}

		events = append(events, &event)
	}

	return events, nil
}

// writeJSON writes data as JSON to a file
func (s *Storage) writeJSON(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0644)
}

// DeleteSession deletes a session and all its data
func (s *Storage) DeleteSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionDir := s.SessionDir(sessionID)
	return os.RemoveAll(sessionDir)
}

// GetBaseDir returns the base storage directory
func (s *Storage) GetBaseDir() string {
	return s.baseDir
}

// EventFilter defines filtering criteria for events
type EventFilter struct {
	Operation string
	Status    string
	Level     models.Level
	Tag       string
	Since     *time.Time
	Until     *time.Time
	Limit     int
}

// Matches checks if an event matches the filter criteria
func (f *EventFilter) Matches(event *models.TraceEvent) bool {
	if f.Operation != "" && event.Operation.Name != f.Operation {
		return false
	}

	if f.Status != "" {
		if event.Output == nil || event.Output.Status != f.Status {
			return false
		}
	}

	if f.Level != "" && event.Level != f.Level {
		return false
	}

	if f.Tag != "" {
		found := false
		for _, t := range event.Tags {
			if t == f.Tag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if f.Since != nil && event.Timestamp.Before(*f.Since) {
		return false
	}

	if f.Until != nil && event.Timestamp.After(*f.Until) {
		return false
	}

	return true
}
