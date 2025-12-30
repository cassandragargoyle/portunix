package main

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotifyVote        NotificationType = "vote"
	NotifyDescription NotificationType = "description"
	NotifyAcceptance  NotificationType = "acceptance"
)

// SMTPClient handles email sending
type SMTPClient struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// NewSMTPClient creates a new SMTP client from config
func NewSMTPClient(config *SMTPConfig) *SMTPClient {
	return &SMTPClient{
		Host:     config.Host,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
		From:     config.From,
	}
}

// EmailData contains data for email templates
type EmailData struct {
	ProductName string
	UserName    string
	Title       string
	Description string
	FiderURL    string
	PostNumber  int
	Provider    string // Provider name (email, fider, etc.)
	ItemID      string // Local item ID (e.g., UC001)
}

// SendEmail sends an email via SMTP
func (c *SMTPClient) SendEmail(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)

	// Build message
	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", c.From, to, subject, body)

	// Use authentication if credentials provided
	var auth smtp.Auth
	if c.Username != "" && c.Password != "" {
		auth = smtp.PlainAuth("", c.Username, c.Password, c.Host)
	}

	err := smtp.SendMail(addr, auth, c.From, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// GenerateNotification generates email subject and body for the given notification type
func GenerateNotification(notifyType NotificationType, data EmailData) (subject, body string, err error) {
	// Load template from file
	templateContent, err := loadTemplate(data.Provider, string(notifyType))
	if err != nil {
		return "", "", err
	}

	// Parse template (first line = subject, after --- = body)
	subjectTmpl, bodyTmpl, err := parseTemplateFile(templateContent)
	if err != nil {
		return "", "", err
	}

	return executeTemplates(subjectTmpl, bodyTmpl, data)
}

// loadTemplate loads a template file from assets/templates/<provider>/<type>.md
func loadTemplate(provider, notifyType string) (string, error) {
	// Find template file - check multiple locations
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)

	locations := []string{
		filepath.Join(execDir, "assets", "templates", provider, notifyType+".md"),
		filepath.Join(execDir, "..", "assets", "templates", provider, notifyType+".md"),
		filepath.Join("assets", "templates", provider, notifyType+".md"),
	}

	var data []byte
	var err error
	for _, loc := range locations {
		if data, err = os.ReadFile(loc); err == nil {
			return string(data), nil
		}
	}

	return "", fmt.Errorf("template not found: %s/%s.md (searched: %v)", provider, notifyType, locations)
}

// parseTemplateFile parses template content into subject and body
// Format: first line = subject, --- = separator, rest = body
func parseTemplateFile(content string) (subject, body string, err error) {
	lines := strings.SplitN(content, "\n", 2)
	if len(lines) < 2 {
		return "", "", fmt.Errorf("invalid template format: missing separator")
	}

	subject = strings.TrimSpace(lines[0])

	// Find separator and extract body
	rest := lines[1]
	if idx := strings.Index(rest, "---"); idx != -1 {
		body = strings.TrimSpace(rest[idx+3:])
	} else {
		body = strings.TrimSpace(rest)
	}

	return subject, body, nil
}

func executeTemplates(subjectTmpl, bodyTmpl string, data EmailData) (string, string, error) {
	funcMap := template.FuncMap{
		"truncate": truncateString,
	}

	// Execute subject template
	subjT, err := template.New("subject").Funcs(funcMap).Parse(subjectTmpl)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse subject template: %w", err)
	}

	var subjBuf bytes.Buffer
	if err := subjT.Execute(&subjBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute subject template: %w", err)
	}

	// Execute body template
	bodyT, err := template.New("body").Funcs(funcMap).Parse(bodyTmpl)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse body template: %w", err)
	}

	var bodyBuf bytes.Buffer
	if err := bodyT.Execute(&bodyBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute body template: %w", err)
	}

	return subjBuf.String(), bodyBuf.String(), nil
}

func truncateString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ParseNotificationType parses a string to NotificationType
func ParseNotificationType(s string) (NotificationType, error) {
	switch strings.ToLower(s) {
	case "vote":
		return NotifyVote, nil
	case "description":
		return NotifyDescription, nil
	case "acceptance":
		return NotifyAcceptance, nil
	default:
		return "", fmt.Errorf("unknown notification type: %s (valid: vote, description, acceptance)", s)
	}
}
