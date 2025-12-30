package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// AuditLevel represents the severity level of audit events
type AuditLevel string

const (
	AuditLevelInfo     AuditLevel = "INFO"
	AuditLevelWarning  AuditLevel = "WARNING"
	AuditLevelError    AuditLevel = "ERROR"
	AuditLevelCritical AuditLevel = "CRITICAL"
)

// AuditEvent represents a single audit log entry
type AuditEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       AuditLevel             `json:"level"`
	Action      string                 `json:"action"`
	User        string                 `json:"user"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target,omitempty"`
	Environment string                 `json:"environment"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
}

// AuditFilter represents filtering criteria for audit queries
type AuditFilter struct {
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Level       AuditLevel `json:"level,omitempty"`
	Action      string     `json:"action,omitempty"`
	User        string     `json:"user,omitempty"`
	Environment string     `json:"environment,omitempty"`
	Success     *bool      `json:"success,omitempty"`
	Limit       int        `json:"limit,omitempty"`
}

// AuditLogger interface defines audit logging capabilities
type AuditLogger interface {
	LogEvent(event *AuditEvent) error
	QueryEvents(filter *AuditFilter) ([]*AuditEvent, error)
	GetStats() (*AuditStats, error)
	Cleanup(olderThan time.Duration) error
}

// FileAuditLogger implements file-based audit logging
type FileAuditLogger struct {
	logDir      string
	maxFileSize int64
	retention   time.Duration
}

// AuditStats represents audit statistics
type AuditStats struct {
	TotalEvents    int64                    `json:"total_events"`
	EventsByLevel  map[AuditLevel]int64     `json:"events_by_level"`
	EventsByAction map[string]int64         `json:"events_by_action"`
	SuccessRate    float64                  `json:"success_rate"`
	LastEvent      *time.Time               `json:"last_event,omitempty"`
	OldestEvent    *time.Time               `json:"oldest_event,omitempty"`
	FileSizes      map[string]int64         `json:"file_sizes"`
}

// AuditManager manages audit logging with multiple backends
type AuditManager struct {
	loggers []AuditLogger
	config  *AuditConfig
}

// AuditConfig represents audit system configuration
type AuditConfig struct {
	Enabled     bool          `json:"enabled"`
	LogDir      string        `json:"log_dir"`
	MaxFileSize int64         `json:"max_file_size"`
	Retention   time.Duration `json:"retention"`
	Formats     []string      `json:"formats"`
	Levels      []AuditLevel  `json:"levels"`
}

// NewFileAuditLogger creates a new file-based audit logger
func NewFileAuditLogger(logDir string, maxFileSize int64, retention time.Duration) (*FileAuditLogger, error) {
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	return &FileAuditLogger{
		logDir:      logDir,
		maxFileSize: maxFileSize,
		retention:   retention,
	}, nil
}

// LogEvent writes an audit event to the log file
func (f *FileAuditLogger) LogEvent(event *AuditEvent) error {
	if event.ID == "" {
		event.ID = generateAuditID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	logFile := f.getLogFileName(event.Timestamp)
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer file.Close()

	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	if _, err := file.WriteString(string(eventData) + "\n"); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	// Check if file rotation is needed
	if stat, err := file.Stat(); err == nil && stat.Size() > f.maxFileSize {
		f.rotateLogFile(logFile)
	}

	return nil
}

// QueryEvents retrieves audit events based on filter criteria
func (f *FileAuditLogger) QueryEvents(filter *AuditFilter) ([]*AuditEvent, error) {
	var events []*AuditEvent

	files, err := f.getLogFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get log files: %w", err)
	}

	for _, file := range files {
		fileEvents, err := f.readEventsFromFile(file, filter)
		if err != nil {
			continue // Skip corrupted files
		}
		events = append(events, fileEvents...)
	}

	// Sort events by timestamp (newest first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	// Apply limit if specified
	if filter.Limit > 0 && len(events) > filter.Limit {
		events = events[:filter.Limit]
	}

	return events, nil
}

// GetStats calculates audit statistics
func (f *FileAuditLogger) GetStats() (*AuditStats, error) {
	stats := &AuditStats{
		EventsByLevel:  make(map[AuditLevel]int64),
		EventsByAction: make(map[string]int64),
		FileSizes:      make(map[string]int64),
	}

	files, err := f.getLogFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get log files: %w", err)
	}

	var successCount int64
	var totalCount int64

	for _, file := range files {
		// Get file size
		if stat, err := os.Stat(file); err == nil {
			stats.FileSizes[filepath.Base(file)] = stat.Size()
		}

		// Read events from file
		events, err := f.readEventsFromFile(file, &AuditFilter{})
		if err != nil {
			continue
		}

		for _, event := range events {
			totalCount++
			stats.EventsByLevel[event.Level]++
			stats.EventsByAction[event.Action]++

			if event.Success {
				successCount++
			}

			// Track oldest and newest events
			if stats.OldestEvent == nil || event.Timestamp.Before(*stats.OldestEvent) {
				stats.OldestEvent = &event.Timestamp
			}
			if stats.LastEvent == nil || event.Timestamp.After(*stats.LastEvent) {
				stats.LastEvent = &event.Timestamp
			}
		}
	}

	stats.TotalEvents = totalCount
	if totalCount > 0 {
		stats.SuccessRate = float64(successCount) / float64(totalCount) * 100
	}

	return stats, nil
}

// Cleanup removes old audit log files based on retention policy
func (f *FileAuditLogger) Cleanup(olderThan time.Duration) error {
	files, err := f.getLogFiles()
	if err != nil {
		return fmt.Errorf("failed to get log files: %w", err)
	}

	cutoff := time.Now().Add(-olderThan)
	var deletedCount int

	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			continue
		}

		if stat.ModTime().Before(cutoff) {
			if err := os.Remove(file); err == nil {
				deletedCount++
			}
		}
	}

	return nil
}

// NewAuditManager creates a new audit manager
func NewAuditManager(config *AuditConfig) (*AuditManager, error) {
	if !config.Enabled {
		return &AuditManager{config: config}, nil
	}

	var loggers []AuditLogger

	// Add file logger
	fileLogger, err := NewFileAuditLogger(config.LogDir, config.MaxFileSize, config.Retention)
	if err != nil {
		return nil, fmt.Errorf("failed to create file audit logger: %w", err)
	}
	loggers = append(loggers, fileLogger)

	return &AuditManager{
		loggers: loggers,
		config:  config,
	}, nil
}

// LogPlaybookExecution logs playbook execution events
func (am *AuditManager) LogPlaybookExecution(user, environment, playbookPath string, success bool, duration time.Duration, err error) error {
	if !am.config.Enabled {
		return nil
	}

	event := &AuditEvent{
		Level:       AuditLevelInfo,
		Action:      "playbook.execute",
		User:        user,
		Source:      "ptx-ansible",
		Target:      playbookPath,
		Environment: environment,
		Success:     success,
		Duration:    duration,
		Details: map[string]interface{}{
			"playbook_path": playbookPath,
		},
	}

	if err != nil {
		event.Level = AuditLevelError
		event.Error = err.Error()
	}

	return am.logToAll(event)
}

// LogSecretAccess logs secret access events
func (am *AuditManager) LogSecretAccess(user, environment, secretStore, secretKey string, success bool) error {
	if !am.config.Enabled {
		return nil
	}

	event := &AuditEvent{
		Level:       AuditLevelInfo,
		Action:      "secret.access",
		User:        user,
		Source:      "ptx-ansible",
		Target:      fmt.Sprintf("%s:%s", secretStore, secretKey),
		Environment: environment,
		Success:     success,
		Details: map[string]interface{}{
			"secret_store": secretStore,
			"secret_key":   secretKey,
		},
	}

	if !success {
		event.Level = AuditLevelWarning
	}

	return am.logToAll(event)
}

// LogRoleAccess logs role-based access control events
func (am *AuditManager) LogRoleAccess(user, environment, requiredRole, action string, success bool) error {
	if !am.config.Enabled {
		return nil
	}

	level := AuditLevelInfo
	if !success {
		level = AuditLevelCritical
	}

	event := &AuditEvent{
		Level:       level,
		Action:      "rbac.access",
		User:        user,
		Source:      "ptx-ansible",
		Target:      action,
		Environment: environment,
		Success:     success,
		Details: map[string]interface{}{
			"required_role": requiredRole,
			"attempted_action": action,
		},
	}

	return am.logToAll(event)
}

// LogSystemEvent logs general system events
func (am *AuditManager) LogSystemEvent(level AuditLevel, action, user, environment string, details map[string]interface{}) error {
	if !am.config.Enabled {
		return nil
	}

	event := &AuditEvent{
		Level:       level,
		Action:      action,
		User:        user,
		Source:      "ptx-ansible",
		Environment: environment,
		Success:     level != AuditLevelError && level != AuditLevelCritical,
		Details:     details,
	}

	return am.logToAll(event)
}

// QueryEvents queries events from all loggers
func (am *AuditManager) QueryEvents(filter *AuditFilter) ([]*AuditEvent, error) {
	if !am.config.Enabled || len(am.loggers) == 0 {
		return []*AuditEvent{}, nil
	}

	// Use the first logger for queries
	return am.loggers[0].QueryEvents(filter)
}

// GetStats gets statistics from all loggers
func (am *AuditManager) GetStats() (*AuditStats, error) {
	if !am.config.Enabled || len(am.loggers) == 0 {
		return &AuditStats{}, nil
	}

	// Use the first logger for stats
	return am.loggers[0].GetStats()
}

// Cleanup performs cleanup on all loggers
func (am *AuditManager) Cleanup() error {
	if !am.config.Enabled {
		return nil
	}

	for _, logger := range am.loggers {
		if err := logger.Cleanup(am.config.Retention); err != nil {
			// Log error but continue with other loggers
			continue
		}
	}

	return nil
}

// Helper functions

func (am *AuditManager) logToAll(event *AuditEvent) error {
	for _, logger := range am.loggers {
		if err := logger.LogEvent(event); err != nil {
			// Log to other loggers even if one fails
			continue
		}
	}
	return nil
}

func (f *FileAuditLogger) getLogFileName(timestamp time.Time) string {
	dateStr := timestamp.Format("2006-01-02")
	return filepath.Join(f.logDir, fmt.Sprintf("audit-%s.log", dateStr))
}

func (f *FileAuditLogger) getLogFiles() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(f.logDir, "audit-*.log"))
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (f *FileAuditLogger) readEventsFromFile(filename string, filter *AuditFilter) ([]*AuditEvent, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []*AuditEvent
	decoder := json.NewDecoder(file)

	for decoder.More() {
		var event AuditEvent
		if err := decoder.Decode(&event); err != nil {
			continue // Skip malformed entries
		}

		if f.matchesFilter(&event, filter) {
			events = append(events, &event)
		}
	}

	return events, nil
}

func (f *FileAuditLogger) matchesFilter(event *AuditEvent, filter *AuditFilter) bool {
	if filter.StartTime != nil && event.Timestamp.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && event.Timestamp.After(*filter.EndTime) {
		return false
	}
	if filter.Level != "" && event.Level != filter.Level {
		return false
	}
	if filter.Action != "" && event.Action != filter.Action {
		return false
	}
	if filter.User != "" && event.User != filter.User {
		return false
	}
	if filter.Environment != "" && event.Environment != filter.Environment {
		return false
	}
	if filter.Success != nil && event.Success != *filter.Success {
		return false
	}
	return true
}

func (f *FileAuditLogger) rotateLogFile(filename string) error {
	rotatedName := fmt.Sprintf("%s.%d", filename, time.Now().Unix())
	return os.Rename(filename, rotatedName)
}

func generateAuditID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
}

// GetDefaultAuditConfig returns default audit configuration
func GetDefaultAuditConfig() *AuditConfig {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".portunix", "audit")

	return &AuditConfig{
		Enabled:     true,
		LogDir:      logDir,
		MaxFileSize: 10 * 1024 * 1024, // 10MB
		Retention:   90 * 24 * time.Hour, // 90 days
		Formats:     []string{"json"},
		Levels:      []AuditLevel{AuditLevelInfo, AuditLevelWarning, AuditLevelError, AuditLevelCritical},
	}
}