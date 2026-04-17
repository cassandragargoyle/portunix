/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package alerts

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// AlertConfig represents the complete alerting configuration
type AlertConfig struct {
	Alerts AlertsConfig `yaml:"alerts"`
}

// AlertsConfig contains all alert settings
type AlertsConfig struct {
	Enabled  bool                     `yaml:"enabled"`
	Channels map[string]ChannelConfig `yaml:"channels"`
	Rules    []RuleConfig             `yaml:"rules"`
}

// ChannelConfig defines an alert channel
type ChannelConfig struct {
	Type       string            `yaml:"type"` // webhook, file, stdout, slack
	WebhookURL string            `yaml:"webhook_url,omitempty"`
	FilePath   string            `yaml:"file_path,omitempty"`
	Headers    map[string]string `yaml:"headers,omitempty"`
	Template   string            `yaml:"template,omitempty"`
}

// RuleConfig defines an alert rule
type RuleConfig struct {
	Name      string   `yaml:"name"`
	Condition string   `yaml:"condition"` // error_rate > 0.05, slow_ops > 10, etc.
	Window    string   `yaml:"window"`    // 5m, 1h, etc.
	Severity  string   `yaml:"severity"`  // critical, high, medium, low
	Channels  []string `yaml:"channels"`  // channel names to notify
	Cooldown  string   `yaml:"cooldown"`  // minimum time between alerts
	Message   string   `yaml:"message"`   // custom message template
}

// Rule represents a parsed and validated rule
type Rule struct {
	Name      string
	Condition Condition
	Window    time.Duration
	Severity  Severity
	Channels  []string
	Cooldown  time.Duration
	Message   string
	LastFired time.Time
	FireCount int
}

// Severity levels for alerts
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// ConditionType represents the type of condition
type ConditionType string

const (
	ConditionErrorRate    ConditionType = "error_rate"
	ConditionSlowOps      ConditionType = "slow_ops"
	ConditionErrorCount   ConditionType = "error_count"
	ConditionEventCount   ConditionType = "event_count"
	ConditionDuration     ConditionType = "duration"
	ConditionCustomMetric ConditionType = "custom"
)

// Condition represents a parsed condition
type Condition struct {
	Type      ConditionType
	Operator  string // >, <, >=, <=, ==, !=
	Threshold float64
	Field     string // for custom metrics
}

// DefaultAlertConfig returns default alerting configuration
func DefaultAlertConfig() *AlertConfig {
	return &AlertConfig{
		Alerts: AlertsConfig{
			Enabled: true,
			Channels: map[string]ChannelConfig{
				"stdout": {
					Type: "stdout",
				},
				"file": {
					Type:     "file",
					FilePath: "~/.portunix/trace/alerts.log",
				},
			},
			Rules: []RuleConfig{
				{
					Name:      "High Error Rate",
					Condition: "error_rate > 0.05",
					Window:    "5m",
					Severity:  "critical",
					Channels:  []string{"stdout"},
					Cooldown:  "10m",
				},
				{
					Name:      "Slow Operations",
					Condition: "slow_ops > 10",
					Window:    "5m",
					Severity:  "high",
					Channels:  []string{"stdout"},
					Cooldown:  "15m",
				},
			},
		},
	}
}

// LoadConfig loads alert configuration from file
func LoadConfig(configPath string) (*AlertConfig, error) {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return DefaultAlertConfig(), nil
		}
		configPath = filepath.Join(homeDir, ".portunix", "trace", "config", "alerts.yaml")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultAlertConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config AlertConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves alert configuration to file
func SaveConfig(config *AlertConfig, configPath string) error {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configPath = filepath.Join(homeDir, ".portunix", "trace", "config", "alerts.yaml")
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// ParseDuration parses duration string like "5m", "1h", "30s"
func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	return time.ParseDuration(s)
}

// ParseSeverity parses severity string
func ParseSeverity(s string) Severity {
	switch s {
	case "critical":
		return SeverityCritical
	case "high":
		return SeverityHigh
	case "medium":
		return SeverityMedium
	case "low":
		return SeverityLow
	default:
		return SeverityMedium
	}
}
