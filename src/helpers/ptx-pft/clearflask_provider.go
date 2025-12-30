package main

import (
	"fmt"
)

// ClearFlaskProvider implements FeedbackProvider interface for ClearFlask
type ClearFlaskProvider struct {
	client *ClearFlaskClient
	config ProviderConfig
	// Cache for status mapping
	statuses   []ClearFlaskStatus
	categories []ClearFlaskCategory
}

// NewClearFlaskProvider creates a new ClearFlask provider
func NewClearFlaskProvider() FeedbackProvider {
	return &ClearFlaskProvider{}
}

// Name returns the provider name
func (p *ClearFlaskProvider) Name() string {
	return "clearflask"
}

// Connect establishes connection to ClearFlask
func (p *ClearFlaskProvider) Connect(config ProviderConfig) error {
	p.config = config

	// Get project_id from options (required for ClearFlask)
	projectID := config.Options["project_id"]
	if projectID == "" {
		return fmt.Errorf("project_id is required for ClearFlask provider")
	}

	p.client = NewClearFlaskClient(config.Endpoint, config.APIToken, projectID)

	// Test connection
	if err := p.client.TestConnection(); err != nil {
		return err
	}

	// Cache statuses and categories for mapping
	var err error
	p.statuses, err = p.client.ListStatuses()
	if err != nil {
		// Non-fatal: we can work without status cache
		p.statuses = nil
	}

	p.categories, err = p.client.ListCategories()
	if err != nil {
		// Non-fatal: we can work without category cache
		p.categories = nil
	}

	return nil
}

// Close closes the connection
func (p *ClearFlaskProvider) Close() error {
	p.client = nil
	p.statuses = nil
	p.categories = nil
	return nil
}

// List returns all feedback items from ClearFlask
func (p *ClearFlaskProvider) List() ([]FeedbackItem, error) {
	if p.client == nil {
		return nil, fmt.Errorf("provider not connected")
	}

	ideas, err := p.client.ListIdeas()
	if err != nil {
		return nil, err
	}

	items := make([]FeedbackItem, len(ideas))
	for i, idea := range ideas {
		items[i] = p.clearflaskIdeaToFeedbackItem(idea)
	}

	return items, nil
}

// Get returns a specific feedback item by ID
func (p *ClearFlaskProvider) Get(id string) (*FeedbackItem, error) {
	if p.client == nil {
		return nil, fmt.Errorf("provider not connected")
	}

	idea, err := p.client.GetIdea(id)
	if err != nil {
		return nil, err
	}

	item := p.clearflaskIdeaToFeedbackItem(*idea)
	return &item, nil
}

// Create creates a new feedback item in ClearFlask
func (p *ClearFlaskProvider) Create(item FeedbackItem) (*FeedbackItem, error) {
	if p.client == nil {
		return nil, fmt.Errorf("provider not connected")
	}

	// Get category from metadata if available
	categoryID := ""
	if item.Metadata != nil {
		categoryID = item.Metadata["category_id"]
	}

	idea, err := p.client.CreateIdea(item.Title, item.Description, categoryID, item.Tags)
	if err != nil {
		return nil, err
	}

	result := p.clearflaskIdeaToFeedbackItem(*idea)
	return &result, nil
}

// Update updates an existing feedback item
func (p *ClearFlaskProvider) Update(item FeedbackItem) error {
	if p.client == nil {
		return fmt.Errorf("provider not connected")
	}

	update := ClearFlaskIdeaUpdate{
		Title:       item.Title,
		Description: item.Description,
		TagIDs:      item.Tags,
	}

	// Map status if provided
	if item.Status != "" {
		statusID := p.mapStatusToID(item.Status)
		if statusID != "" {
			update.StatusID = statusID
		}
	}

	// Get category from metadata if available
	if item.Metadata != nil && item.Metadata["category_id"] != "" {
		update.CategoryID = item.Metadata["category_id"]
	}

	return p.client.UpdateIdea(item.ExternalID, update)
}

// Delete removes a feedback item
func (p *ClearFlaskProvider) Delete(id string) error {
	if p.client == nil {
		return fmt.Errorf("provider not connected")
	}

	return p.client.DeleteIdea(id)
}

// clearflaskIdeaToFeedbackItem converts a ClearFlaskIdea to FeedbackItem
func (p *ClearFlaskProvider) clearflaskIdeaToFeedbackItem(idea ClearFlaskIdea) FeedbackItem {
	// Get status name from cached statuses
	statusName := idea.StatusID
	if idea.Status != nil {
		statusName = idea.Status.Name
	} else if p.statuses != nil {
		for _, s := range p.statuses {
			if s.StatusID == idea.StatusID {
				statusName = s.Name
				break
			}
		}
	}

	// Get category name from cached categories
	categoryName := idea.CategoryID
	if idea.Category != nil {
		categoryName = idea.Category.Name
	} else if p.categories != nil {
		for _, c := range p.categories {
			if c.CategoryID == idea.CategoryID {
				categoryName = c.Name
				break
			}
		}
	}

	// Build metadata
	metadata := map[string]string{
		"slug":        idea.Slug,
		"category_id": idea.CategoryID,
		"category":    categoryName,
		"status_id":   idea.StatusID,
	}

	// Add author info if available
	if idea.Author != nil {
		metadata["author_id"] = idea.Author.UserID
		metadata["author_name"] = idea.Author.Name
	} else if idea.AuthorUserID != "" {
		metadata["author_id"] = idea.AuthorUserID
	}

	// Map ClearFlask category to Categories (if category exists)
	var categories []string
	if categoryName != "" && categoryName != idea.CategoryID {
		categories = append(categories, categoryName)
	} else if idea.CategoryID != "" {
		categories = append(categories, idea.CategoryID)
	}

	return FeedbackItem{
		ID:          idea.IdeaID,
		ExternalID:  idea.IdeaID,
		Title:       idea.Title,
		Description: idea.Description,
		Status:      p.mapStatusToInternal(statusName),
		Categories:  categories,
		Tags:        idea.TagIDs,
		Votes:       idea.VotersCount,
		CreatedAt:   idea.Created,
		UpdatedAt:   idea.Edited,
		Metadata:    metadata,
	}
}

// mapStatusToInternal maps ClearFlask status to internal pft status
func (p *ClearFlaskProvider) mapStatusToInternal(status string) string {
	// ClearFlask uses customizable statuses, but common ones include:
	// "Under Review", "Planned", "In Progress", "Completed", "Closed"
	statusLower := status

	switch statusLower {
	case "under review", "new", "open", "pending":
		return "open"
	case "planned", "accepted":
		return "planned"
	case "in progress", "started", "working":
		return "started"
	case "completed", "done", "implemented", "released":
		return "completed"
	case "closed", "declined", "rejected", "wont do", "duplicate":
		return "declined"
	default:
		// Return as-is for unknown statuses
		return status
	}
}

// mapStatusToID maps internal status name to ClearFlask status ID
func (p *ClearFlaskProvider) mapStatusToID(internalStatus string) string {
	if p.statuses == nil {
		return ""
	}

	// Try to find matching status by internal name mapping
	for _, s := range p.statuses {
		mapped := p.mapStatusToInternal(s.Name)
		if mapped == internalStatus {
			return s.StatusID
		}
	}

	// Try exact match
	for _, s := range p.statuses {
		if s.Name == internalStatus || s.StatusID == internalStatus {
			return s.StatusID
		}
	}

	return ""
}

// Register the ClearFlask provider
func init() {
	RegisterProvider("clearflask", NewClearFlaskProvider)
}
