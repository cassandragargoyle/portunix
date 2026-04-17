/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package alerts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"portunix.ai/portunix/src/helpers/ptx-trace/models"
)

// Manager manages alert evaluation and notification
type Manager struct {
	config    *AlertConfig
	evaluator *Evaluator
	channels  map[string]Channel
	history   *AlertHistory
	mu        sync.RWMutex
	enabled   bool
}

// NewManager creates a new alert manager
func NewManager(configPath string) (*Manager, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	evaluator, err := NewEvaluator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create evaluator: %w", err)
	}

	factory := NewChannelFactory()
	channels, err := factory.CreateChannels(config.Alerts.Channels)
	if err != nil {
		return nil, fmt.Errorf("failed to create channels: %w", err)
	}

	history, err := NewAlertHistory("")
	if err != nil {
		return nil, fmt.Errorf("failed to create history: %w", err)
	}

	return &Manager{
		config:    config,
		evaluator: evaluator,
		channels:  channels,
		history:   history,
		enabled:   config.Alerts.Enabled,
	}, nil
}

// NewManagerWithConfig creates a manager with provided config
func NewManagerWithConfig(config *AlertConfig) (*Manager, error) {
	evaluator, err := NewEvaluator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create evaluator: %w", err)
	}

	factory := NewChannelFactory()
	channels, err := factory.CreateChannels(config.Alerts.Channels)
	if err != nil {
		return nil, fmt.Errorf("failed to create channels: %w", err)
	}

	history, err := NewAlertHistory("")
	if err != nil {
		return nil, fmt.Errorf("failed to create history: %w", err)
	}

	return &Manager{
		config:    config,
		evaluator: evaluator,
		channels:  channels,
		history:   history,
		enabled:   config.Alerts.Enabled,
	}, nil
}

// Enable enables alert processing
func (m *Manager) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables alert processing
func (m *Manager) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// IsEnabled returns whether alerting is enabled
func (m *Manager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// ProcessSession evaluates alerts for a session and sends notifications
func (m *Manager) ProcessSession(session *models.Session, events []*models.TraceEvent) ([]*Alert, error) {
	m.mu.RLock()
	enabled := m.enabled
	m.mu.RUnlock()

	if !enabled {
		return nil, nil
	}

	alerts := m.evaluator.EvaluateSession(session, events)
	if len(alerts) == 0 {
		return nil, nil
	}

	var errors []error
	for _, alert := range alerts {
		// Record in history
		if err := m.history.Add(alert); err != nil {
			errors = append(errors, fmt.Errorf("failed to record alert: %w", err))
		}

		// Send to channels
		for _, channelName := range alert.Rule.Channels {
			ch, ok := m.channels[channelName]
			if !ok {
				errors = append(errors, fmt.Errorf("channel not found: %s", channelName))
				continue
			}

			if err := ch.Send(alert); err != nil {
				errors = append(errors, fmt.Errorf("failed to send to %s: %w", channelName, err))
			}
		}
	}

	if len(errors) > 0 {
		return alerts, fmt.Errorf("alert processing errors: %v", errors)
	}

	return alerts, nil
}

// ProcessContext evaluates alerts for a pre-built context
func (m *Manager) ProcessContext(ctx *EvaluationContext) ([]*Alert, error) {
	m.mu.RLock()
	enabled := m.enabled
	m.mu.RUnlock()

	if !enabled {
		return nil, nil
	}

	alerts := m.evaluator.Evaluate(ctx)
	if len(alerts) == 0 {
		return nil, nil
	}

	var errors []error
	for _, alert := range alerts {
		if err := m.history.Add(alert); err != nil {
			errors = append(errors, fmt.Errorf("failed to record alert: %w", err))
		}

		for _, channelName := range alert.Rule.Channels {
			ch, ok := m.channels[channelName]
			if !ok {
				errors = append(errors, fmt.Errorf("channel not found: %s", channelName))
				continue
			}

			if err := ch.Send(alert); err != nil {
				errors = append(errors, fmt.Errorf("failed to send to %s: %w", channelName, err))
			}
		}
	}

	if len(errors) > 0 {
		return alerts, fmt.Errorf("alert processing errors: %v", errors)
	}

	return alerts, nil
}

// GetRules returns all configured rules
func (m *Manager) GetRules() []*Rule {
	return m.evaluator.GetRules()
}

// GetChannels returns all configured channels
func (m *Manager) GetChannels() map[string]Channel {
	return m.channels
}

// GetHistory returns the alert history
func (m *Manager) GetHistory() *AlertHistory {
	return m.history
}

// ResetCooldowns resets all rule cooldowns
func (m *Manager) ResetCooldowns() {
	m.evaluator.ResetCooldowns()
}

// AlertHistory stores alert history
type AlertHistory struct {
	filePath string
	alerts   []*AlertRecord
	mu       sync.RWMutex
}

// AlertRecord represents a stored alert
type AlertRecord struct {
	ID          string    `json:"id"`
	RuleName    string    `json:"rule_name"`
	Severity    Severity  `json:"severity"`
	Message     string    `json:"message"`
	Value       float64   `json:"value"`
	Timestamp   time.Time `json:"timestamp"`
	SessionID   string    `json:"session_id,omitempty"`
	SessionName string    `json:"session_name,omitempty"`
	Channels    []string  `json:"channels"`
}

// NewAlertHistory creates a new alert history
func NewAlertHistory(filePath string) (*AlertHistory, error) {
	if filePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		filePath = filepath.Join(homeDir, ".portunix", "trace", "alert-history.json")
	}

	h := &AlertHistory{
		filePath: filePath,
		alerts:   make([]*AlertRecord, 0),
	}

	// Load existing history
	if err := h.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return h, nil
}

// Add adds an alert to history
func (h *AlertHistory) Add(alert *Alert) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	record := &AlertRecord{
		ID:        fmt.Sprintf("alert-%d", time.Now().UnixNano()),
		RuleName:  alert.Rule.Name,
		Severity:  alert.Rule.Severity,
		Message:   alert.Message,
		Value:     alert.Value,
		Timestamp: alert.Timestamp,
		Channels:  alert.Rule.Channels,
	}

	if alert.Context != nil && alert.Context.Session != nil {
		record.SessionID = alert.Context.Session.ID
		record.SessionName = alert.Context.Session.Name
	}

	h.alerts = append(h.alerts, record)

	// Keep only last 1000 alerts
	if len(h.alerts) > 1000 {
		h.alerts = h.alerts[len(h.alerts)-1000:]
	}

	return h.save()
}

// GetRecent returns recent alerts
func (h *AlertHistory) GetRecent(limit int) []*AlertRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if limit <= 0 || limit > len(h.alerts) {
		limit = len(h.alerts)
	}

	result := make([]*AlertRecord, limit)
	for i := 0; i < limit; i++ {
		result[i] = h.alerts[len(h.alerts)-limit+i]
	}
	return result
}

// GetBySeverity returns alerts filtered by severity
func (h *AlertHistory) GetBySeverity(severity Severity, limit int) []*AlertRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []*AlertRecord
	for i := len(h.alerts) - 1; i >= 0 && len(result) < limit; i-- {
		if h.alerts[i].Severity == severity {
			result = append(result, h.alerts[i])
		}
	}
	return result
}

// GetBySession returns alerts for a specific session
func (h *AlertHistory) GetBySession(sessionID string) []*AlertRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []*AlertRecord
	for _, a := range h.alerts {
		if a.SessionID == sessionID {
			result = append(result, a)
		}
	}
	return result
}

// GetByTimeRange returns alerts within a time range
func (h *AlertHistory) GetByTimeRange(start, end time.Time) []*AlertRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []*AlertRecord
	for _, a := range h.alerts {
		if (a.Timestamp.Equal(start) || a.Timestamp.After(start)) &&
			(a.Timestamp.Equal(end) || a.Timestamp.Before(end)) {
			result = append(result, a)
		}
	}
	return result
}

// Count returns total number of alerts
func (h *AlertHistory) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.alerts)
}

// Clear removes all alert history
func (h *AlertHistory) Clear() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.alerts = make([]*AlertRecord, 0)
	return h.save()
}

func (h *AlertHistory) load() error {
	data, err := os.ReadFile(h.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &h.alerts)
}

func (h *AlertHistory) save() error {
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(h.alerts, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(h.filePath, data, 0644)
}

// Stats returns alert statistics
type AlertStats struct {
	Total          int            `json:"total"`
	BySeverity     map[string]int `json:"by_severity"`
	ByRule         map[string]int `json:"by_rule"`
	LastAlert      *time.Time     `json:"last_alert,omitempty"`
	AlertsToday    int            `json:"alerts_today"`
	AlertsThisWeek int            `json:"alerts_this_week"`
}

// GetStats returns alert statistics
func (h *AlertHistory) GetStats() *AlertStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := &AlertStats{
		Total:      len(h.alerts),
		BySeverity: make(map[string]int),
		ByRule:     make(map[string]int),
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -7)

	for _, a := range h.alerts {
		stats.BySeverity[string(a.Severity)]++
		stats.ByRule[a.RuleName]++

		if a.Timestamp.After(todayStart) {
			stats.AlertsToday++
		}
		if a.Timestamp.After(weekStart) {
			stats.AlertsThisWeek++
		}
	}

	if len(h.alerts) > 0 {
		lastTime := h.alerts[len(h.alerts)-1].Timestamp
		stats.LastAlert = &lastTime
	}

	return stats
}
