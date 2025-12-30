package main

import (
	"fmt"
)

// EmailProvider implements FeedbackProvider interface for email-only mode
// In this mode, feedback is managed through email without external system
type EmailProvider struct {
	config ProviderConfig
}

// NewEmailProvider creates a new Email provider
func NewEmailProvider() FeedbackProvider {
	return &EmailProvider{}
}

// Name returns the provider name
func (p *EmailProvider) Name() string {
	return "email"
}

// Connect initializes the email provider (no external connection needed)
func (p *EmailProvider) Connect(config ProviderConfig) error {
	p.config = config
	return nil
}

// Close closes the provider
func (p *EmailProvider) Close() error {
	return nil
}

// List returns all feedback items (from local files only)
func (p *EmailProvider) List() ([]FeedbackItem, error) {
	// Email provider doesn't have external storage
	// Items are managed locally via markdown files
	return nil, fmt.Errorf("email provider: use local files for listing items")
}

// Get returns a specific feedback item
func (p *EmailProvider) Get(id string) (*FeedbackItem, error) {
	return nil, fmt.Errorf("email provider: use local files for getting items")
}

// Create creates a new feedback item
func (p *EmailProvider) Create(item FeedbackItem) (*FeedbackItem, error) {
	return nil, fmt.Errorf("email provider: use local files for creating items")
}

// Update updates an existing feedback item
func (p *EmailProvider) Update(item FeedbackItem) error {
	return fmt.Errorf("email provider: use local files for updating items")
}

// Delete removes a feedback item
func (p *EmailProvider) Delete(id string) error {
	return fmt.Errorf("email provider: use local files for deleting items")
}

// Register the Email provider
func init() {
	RegisterProvider("email", NewEmailProvider)
}
