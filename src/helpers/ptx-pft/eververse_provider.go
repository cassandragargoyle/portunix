package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EververseProvider implements FeedbackProvider interface for Eververse/Supabase
type EververseProvider struct {
	client    *http.Client
	config    ProviderConfig
	supaURL   string
	anonKey   string
	serviceKey string
	projectID string
}

// NewEververseProvider creates a new Eververse provider
func NewEververseProvider() FeedbackProvider {
	return &EververseProvider{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Name returns the provider name
func (p *EververseProvider) Name() string {
	return "eververse"
}

// Connect establishes connection to Eververse via Supabase
func (p *EververseProvider) Connect(config ProviderConfig) error {
	p.config = config

	// Get Supabase configuration from options
	p.supaURL = config.Endpoint
	if p.supaURL == "" {
		p.supaURL = config.Options["supabase_url"]
	}
	if p.supaURL == "" {
		return fmt.Errorf("supabase_url is required for Eververse provider")
	}

	p.anonKey = config.APIToken
	if p.anonKey == "" {
		p.anonKey = config.Options["supabase_anon_key"]
	}
	if p.anonKey == "" {
		return fmt.Errorf("supabase_anon_key is required for Eververse provider")
	}

	p.serviceKey = config.Options["supabase_service_key"]
	p.projectID = config.Options["project_id"]

	// Test connection by checking health endpoint
	if err := p.testConnection(); err != nil {
		return fmt.Errorf("failed to connect to Eververse: %w", err)
	}

	return nil
}

// testConnection verifies connectivity to Supabase/Eververse
func (p *EververseProvider) testConnection() error {
	// Try to access Supabase REST API
	url := fmt.Sprintf("%s/rest/v1/", p.supaURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", p.anonKey)
	req.Header.Set("Authorization", "Bearer "+p.anonKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	// 200 or 404 (no tables yet) is acceptable for connection test
	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Close closes the connection
func (p *EververseProvider) Close() error {
	p.client = nil
	return nil
}

// EververseFeature represents a feature in Eververse
type EververseFeature struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Content     string                 `json:"content,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Priority    int                    `json:"priority,omitempty"`
	CreatedAt   string                 `json:"created_at,omitempty"`
	UpdatedAt   string                 `json:"updated_at,omitempty"`
	ProductID   string                 `json:"product_id,omitempty"`
	RoadmapID   string                 `json:"roadmap_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// EversverseFeedback represents feedback in Eververse
type EververseFeedback struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content,omitempty"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	FeatureID string `json:"feature_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

// List retrieves all feedback items from Eververse
func (p *EververseProvider) List() ([]FeedbackItem, error) {
	if p.client == nil {
		return nil, fmt.Errorf("provider not connected")
	}

	// Query features from Supabase
	features, err := p.listFeatures()
	if err != nil {
		// If features table doesn't exist, try feedback table
		feedback, fbErr := p.listFeedback()
		if fbErr != nil {
			return nil, fmt.Errorf("failed to list items: features: %v, feedback: %v", err, fbErr)
		}
		return feedback, nil
	}

	return features, nil
}

// listFeatures retrieves features from Eververse
func (p *EververseProvider) listFeatures() ([]FeedbackItem, error) {
	url := fmt.Sprintf("%s/rest/v1/features?select=*", p.supaURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	p.setHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list features: %d - %s", resp.StatusCode, string(body))
	}

	var features []EververseFeature
	if err := json.NewDecoder(resp.Body).Decode(&features); err != nil {
		return nil, fmt.Errorf("failed to decode features: %w", err)
	}

	items := make([]FeedbackItem, len(features))
	for i, f := range features {
		items[i] = p.featureToFeedbackItem(f)
	}

	return items, nil
}

// listFeedback retrieves feedback from Eververse
func (p *EververseProvider) listFeedback() ([]FeedbackItem, error) {
	url := fmt.Sprintf("%s/rest/v1/feedback?select=*", p.supaURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	p.setHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list feedback: %d - %s", resp.StatusCode, string(body))
	}

	var feedback []EververseFeedback
	if err := json.NewDecoder(resp.Body).Decode(&feedback); err != nil {
		return nil, fmt.Errorf("failed to decode feedback: %w", err)
	}

	items := make([]FeedbackItem, len(feedback))
	for i, f := range feedback {
		items[i] = p.feedbackToFeedbackItem(f)
	}

	return items, nil
}

// Get retrieves a specific feedback item by ID
func (p *EververseProvider) Get(id string) (*FeedbackItem, error) {
	if p.client == nil {
		return nil, fmt.Errorf("provider not connected")
	}

	// Try features first
	url := fmt.Sprintf("%s/rest/v1/features?id=eq.%s", p.supaURL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	p.setHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var features []EververseFeature
		if err := json.NewDecoder(resp.Body).Decode(&features); err != nil {
			return nil, err
		}
		if len(features) > 0 {
			item := p.featureToFeedbackItem(features[0])
			return &item, nil
		}
	}

	// Try feedback table
	url = fmt.Sprintf("%s/rest/v1/feedback?id=eq.%s", p.supaURL, id)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	p.setHeaders(req)

	resp, err = p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var feedback []EververseFeedback
		if err := json.NewDecoder(resp.Body).Decode(&feedback); err != nil {
			return nil, err
		}
		if len(feedback) > 0 {
			item := p.feedbackToFeedbackItem(feedback[0])
			return &item, nil
		}
	}

	return nil, fmt.Errorf("item not found: %s", id)
}

// Create creates a new feedback item in Eververse
func (p *EververseProvider) Create(item FeedbackItem) (*FeedbackItem, error) {
	if p.client == nil {
		return nil, fmt.Errorf("provider not connected")
	}

	feature := EververseFeature{
		Title:       item.Title,
		Description: item.Description,
		Status:      p.mapStatusToEververse(item.Status),
	}

	if item.Metadata != nil {
		if productID, ok := item.Metadata["product_id"]; ok {
			feature.ProductID = productID
		}
	}

	body, err := json.Marshal(feature)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/rest/v1/features", p.supaURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	p.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create feature: %d - %s", resp.StatusCode, string(respBody))
	}

	var created []EververseFeature
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, err
	}

	if len(created) == 0 {
		return nil, fmt.Errorf("no feature returned after creation")
	}

	result := p.featureToFeedbackItem(created[0])
	return &result, nil
}

// Update modifies an existing feedback item
func (p *EververseProvider) Update(item FeedbackItem) error {
	if p.client == nil {
		return fmt.Errorf("provider not connected")
	}

	update := map[string]interface{}{
		"title":       item.Title,
		"description": item.Description,
		"status":      p.mapStatusToEververse(item.Status),
	}

	body, err := json.Marshal(update)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/rest/v1/features?id=eq.%s", p.supaURL, item.ExternalID)
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	p.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update feature: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Delete removes a feedback item
func (p *EververseProvider) Delete(id string) error {
	if p.client == nil {
		return fmt.Errorf("provider not connected")
	}

	url := fmt.Sprintf("%s/rest/v1/features?id=eq.%s", p.supaURL, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	p.setHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete feature: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// setHeaders sets the required headers for Supabase API requests
func (p *EververseProvider) setHeaders(req *http.Request) {
	req.Header.Set("apikey", p.anonKey)
	// Use service key for privileged operations if available
	if p.serviceKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.serviceKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+p.anonKey)
	}
}

// featureToFeedbackItem converts an Eververse feature to FeedbackItem
func (p *EververseProvider) featureToFeedbackItem(f EververseFeature) FeedbackItem {
	metadata := map[string]string{
		"type":       "feature",
		"product_id": f.ProductID,
		"roadmap_id": f.RoadmapID,
	}

	// Extract categories from metadata if available
	var categories []string
	if f.Metadata != nil {
		for k, v := range f.Metadata {
			if str, ok := v.(string); ok {
				metadata[k] = str
				// Check for category-related fields
				if k == "category" || k == "categories" {
					categories = append(categories, str)
				}
			}
			// Handle categories as array
			if k == "categories" {
				if arr, ok := v.([]interface{}); ok {
					for _, item := range arr {
						if s, ok := item.(string); ok {
							categories = append(categories, s)
						}
					}
				}
			}
		}
	}

	// Use product_id as category if no explicit categories
	if len(categories) == 0 && f.ProductID != "" {
		categories = append(categories, f.ProductID)
	}

	return FeedbackItem{
		ID:          f.ID,
		ExternalID:  f.ID,
		Title:       f.Title,
		Description: f.Description,
		Status:      p.mapStatusToInternal(f.Status),
		Categories:  categories,
		Priority:    p.priorityToString(f.Priority),
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
		Metadata:    metadata,
	}
}

// feedbackToFeedbackItem converts Eververse feedback to FeedbackItem
func (p *EververseProvider) feedbackToFeedbackItem(f EververseFeedback) FeedbackItem {
	return FeedbackItem{
		ID:          f.ID,
		ExternalID:  f.ID,
		Title:       f.Title,
		Description: f.Content,
		Status:      p.mapStatusToInternal(f.Status),
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
		Metadata: map[string]string{
			"type":       "feedback",
			"feature_id": f.FeatureID,
			"user_id":    f.UserID,
		},
	}
}

// mapStatusToInternal maps Eververse status to internal pft status
func (p *EververseProvider) mapStatusToInternal(status string) string {
	// Eververse uses various status values
	switch status {
	case "idea", "draft", "new", "pending":
		return "open"
	case "under_review", "exploring", "researching":
		return "open"
	case "planned", "accepted", "prioritized":
		return "planned"
	case "in_progress", "building", "developing":
		return "started"
	case "completed", "done", "shipped", "released", "live":
		return "completed"
	case "declined", "rejected", "wont_do", "archived", "cancelled":
		return "declined"
	default:
		return status
	}
}

// mapStatusToEververse maps internal status to Eververse status
func (p *EververseProvider) mapStatusToEververse(internalStatus string) string {
	switch internalStatus {
	case "open":
		return "idea"
	case "planned":
		return "planned"
	case "started":
		return "in_progress"
	case "completed":
		return "completed"
	case "declined":
		return "declined"
	default:
		return internalStatus
	}
}

// priorityToString converts priority int to string
func (p *EververseProvider) priorityToString(priority int) string {
	switch priority {
	case 1:
		return "critical"
	case 2:
		return "high"
	case 3:
		return "medium"
	case 4:
		return "low"
	default:
		return "medium"
	}
}

// Register the Eververse provider
func init() {
	RegisterProvider("eververse", NewEververseProvider)
}
