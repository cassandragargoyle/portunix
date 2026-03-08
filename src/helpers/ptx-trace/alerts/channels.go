package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Channel represents an alert notification channel
type Channel interface {
	Send(alert *Alert) error
	Name() string
	Type() string
}

// ChannelFactory creates channels from configuration
type ChannelFactory struct{}

// NewChannelFactory creates a new channel factory
func NewChannelFactory() *ChannelFactory {
	return &ChannelFactory{}
}

// CreateChannel creates a channel from configuration
func (f *ChannelFactory) CreateChannel(name string, config ChannelConfig) (Channel, error) {
	switch config.Type {
	case "webhook":
		return NewWebhookChannel(name, config)
	case "file":
		return NewFileChannel(name, config)
	case "stdout":
		return NewStdoutChannel(name, config)
	case "slack":
		return NewSlackChannel(name, config)
	default:
		return nil, fmt.Errorf("unknown channel type: %s", config.Type)
	}
}

// CreateChannels creates all channels from configuration
func (f *ChannelFactory) CreateChannels(configs map[string]ChannelConfig) (map[string]Channel, error) {
	channels := make(map[string]Channel)
	for name, config := range configs {
		ch, err := f.CreateChannel(name, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create channel '%s': %w", name, err)
		}
		channels[name] = ch
	}
	return channels, nil
}

// WebhookChannel sends alerts via HTTP webhook
type WebhookChannel struct {
	name       string
	url        string
	headers    map[string]string
	template   *template.Template
	httpClient *http.Client
}

// NewWebhookChannel creates a new webhook channel
func NewWebhookChannel(name string, config ChannelConfig) (*WebhookChannel, error) {
	if config.WebhookURL == "" {
		return nil, fmt.Errorf("webhook_url is required for webhook channel")
	}

	ch := &WebhookChannel{
		name:    name,
		url:     config.WebhookURL,
		headers: config.Headers,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	if config.Template != "" {
		tmpl, err := template.New("alert").Parse(config.Template)
		if err != nil {
			return nil, fmt.Errorf("invalid template: %w", err)
		}
		ch.template = tmpl
	}

	return ch, nil
}

// Send sends an alert via webhook
func (c *WebhookChannel) Send(alert *Alert) error {
	var body []byte
	var err error

	if c.template != nil {
		var buf bytes.Buffer
		if err := c.template.Execute(&buf, alert); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
		body = buf.Bytes()
	} else {
		// Default JSON payload
		payload := map[string]interface{}{
			"alert":     alert.Rule.Name,
			"severity":  alert.Rule.Severity,
			"message":   alert.Message,
			"value":     alert.Value,
			"timestamp": alert.Timestamp.Format(time.RFC3339),
		}
		if alert.Context != nil && alert.Context.Session != nil {
			payload["session_id"] = alert.Context.Session.ID
			payload["session_name"] = alert.Context.Session.Name
		}
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *WebhookChannel) Name() string { return c.name }
func (c *WebhookChannel) Type() string { return "webhook" }

// FileChannel writes alerts to a file
type FileChannel struct {
	name     string
	filePath string
	template *template.Template
}

// NewFileChannel creates a new file channel
func NewFileChannel(name string, config ChannelConfig) (*FileChannel, error) {
	path := config.FilePath
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(homeDir, ".portunix", "trace", "alerts.log")
	}

	// Expand ~ in path
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(homeDir, path[1:])
	}

	ch := &FileChannel{
		name:     name,
		filePath: path,
	}

	if config.Template != "" {
		tmpl, err := template.New("alert").Parse(config.Template)
		if err != nil {
			return nil, fmt.Errorf("invalid template: %w", err)
		}
		ch.template = tmpl
	}

	return ch, nil
}

// Send writes an alert to file
func (c *FileChannel) Send(alert *Alert) error {
	// Ensure directory exists
	dir := filepath.Dir(c.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open file in append mode
	f, err := os.OpenFile(c.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var line string
	if c.template != nil {
		var buf bytes.Buffer
		if err := c.template.Execute(&buf, alert); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
		line = buf.String()
	} else {
		// Default format
		sessionInfo := ""
		if alert.Context != nil && alert.Context.Session != nil {
			sessionInfo = fmt.Sprintf(" session=%s", alert.Context.Session.ID)
		}
		line = fmt.Sprintf("[%s] %s [%s] %s (value=%.2f)%s\n",
			alert.Timestamp.Format("2006-01-02 15:04:05"),
			strings.ToUpper(string(alert.Rule.Severity)),
			alert.Rule.Name,
			alert.Message,
			alert.Value,
			sessionInfo,
		)
	}

	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func (c *FileChannel) Name() string { return c.name }
func (c *FileChannel) Type() string { return "file" }

// StdoutChannel prints alerts to stdout
type StdoutChannel struct {
	name     string
	template *template.Template
	colored  bool
}

// NewStdoutChannel creates a new stdout channel
func NewStdoutChannel(name string, config ChannelConfig) (*StdoutChannel, error) {
	ch := &StdoutChannel{
		name:    name,
		colored: true,
	}

	if config.Template != "" {
		tmpl, err := template.New("alert").Parse(config.Template)
		if err != nil {
			return nil, fmt.Errorf("invalid template: %w", err)
		}
		ch.template = tmpl
	}

	return ch, nil
}

// Send prints an alert to stdout
func (c *StdoutChannel) Send(alert *Alert) error {
	if c.template != nil {
		var buf bytes.Buffer
		if err := c.template.Execute(&buf, alert); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
		fmt.Println(buf.String())
		return nil
	}

	// Default colored output
	var severityColor string
	var severityIcon string
	switch alert.Rule.Severity {
	case SeverityCritical:
		severityColor = "\033[91m" // Red
		severityIcon = "🔴"
	case SeverityHigh:
		severityColor = "\033[93m" // Yellow
		severityIcon = "🟠"
	case SeverityMedium:
		severityColor = "\033[94m" // Blue
		severityIcon = "🟡"
	case SeverityLow:
		severityColor = "\033[90m" // Gray
		severityIcon = "⚪"
	default:
		severityColor = "\033[0m"
		severityIcon = "⚪"
	}
	resetColor := "\033[0m"

	sessionInfo := ""
	if alert.Context != nil && alert.Context.Session != nil {
		sessionInfo = fmt.Sprintf(" [session: %s]", alert.Context.Session.Name)
	}

	if c.colored {
		fmt.Printf("%s%s ALERT [%s]: %s%s (value: %.2f)%s\n",
			severityColor,
			severityIcon,
			alert.Rule.Name,
			alert.Message,
			resetColor,
			alert.Value,
			sessionInfo,
		)
	} else {
		fmt.Printf("%s ALERT [%s]: %s (value: %.2f)%s\n",
			severityIcon,
			alert.Rule.Name,
			alert.Message,
			alert.Value,
			sessionInfo,
		)
	}

	return nil
}

func (c *StdoutChannel) Name() string { return c.name }
func (c *StdoutChannel) Type() string { return "stdout" }

// SlackChannel sends alerts to Slack
type SlackChannel struct {
	name       string
	webhookURL string
	httpClient *http.Client
}

// NewSlackChannel creates a new Slack channel
func NewSlackChannel(name string, config ChannelConfig) (*SlackChannel, error) {
	if config.WebhookURL == "" {
		return nil, fmt.Errorf("webhook_url is required for slack channel")
	}

	return &SlackChannel{
		name:       name,
		webhookURL: config.WebhookURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Send sends an alert to Slack
func (c *SlackChannel) Send(alert *Alert) error {
	var emoji string
	var color string
	switch alert.Rule.Severity {
	case SeverityCritical:
		emoji = ":red_circle:"
		color = "#ff0000"
	case SeverityHigh:
		emoji = ":large_orange_circle:"
		color = "#ff8c00"
	case SeverityMedium:
		emoji = ":large_yellow_circle:"
		color = "#ffcc00"
	case SeverityLow:
		emoji = ":white_circle:"
		color = "#cccccc"
	default:
		emoji = ":white_circle:"
		color = "#cccccc"
	}

	sessionInfo := ""
	if alert.Context != nil && alert.Context.Session != nil {
		sessionInfo = fmt.Sprintf("\n*Session:* %s (%s)", alert.Context.Session.Name, alert.Context.Session.ID)
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"blocks": []map[string]interface{}{
					{
						"type": "section",
						"text": map[string]string{
							"type": "mrkdwn",
							"text": fmt.Sprintf("%s *%s Alert: %s*\n%s%s\n*Value:* %.2f",
								emoji,
								strings.Title(string(alert.Rule.Severity)),
								alert.Rule.Name,
								alert.Message,
								sessionInfo,
								alert.Value,
							),
						},
					},
					{
						"type": "context",
						"elements": []map[string]string{
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("Fired at %s", alert.Timestamp.Format(time.RFC3339)),
							},
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack returned error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *SlackChannel) Name() string { return c.name }
func (c *SlackChannel) Type() string { return "slack" }
