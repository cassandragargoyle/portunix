package main

import (
	"fmt"
	"strconv"
)

// FiderProvider implements FeedbackProvider interface for Fider.io
type FiderProvider struct {
	client *FiderClient
	config ProviderConfig
}

// NewFiderProvider creates a new Fider provider
func NewFiderProvider() FeedbackProvider {
	return &FiderProvider{}
}

// Name returns the provider name
func (p *FiderProvider) Name() string {
	return "fider"
}

// Connect establishes connection to Fider
func (p *FiderProvider) Connect(config ProviderConfig) error {
	p.config = config
	p.client = NewFiderClient(config.Endpoint, config.APIToken)
	return p.client.TestConnection()
}

// Close closes the connection
func (p *FiderProvider) Close() error {
	// HTTP client doesn't need explicit close
	p.client = nil
	return nil
}

// List returns all feedback items from Fider
func (p *FiderProvider) List() ([]FeedbackItem, error) {
	if p.client == nil {
		return nil, fmt.Errorf("provider not connected")
	}

	posts, err := p.client.ListPosts()
	if err != nil {
		return nil, err
	}

	items := make([]FeedbackItem, len(posts))
	for i, post := range posts {
		items[i] = fiderPostToFeedbackItem(post)
	}

	return items, nil
}

// Get returns a specific feedback item by ID
func (p *FiderProvider) Get(id string) (*FeedbackItem, error) {
	if p.client == nil {
		return nil, fmt.Errorf("provider not connected")
	}

	number, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID: %s", id)
	}

	post, err := p.client.GetPost(number)
	if err != nil {
		return nil, err
	}

	item := fiderPostToFeedbackItem(*post)
	return &item, nil
}

// Create creates a new feedback item in Fider
func (p *FiderProvider) Create(item FeedbackItem) (*FeedbackItem, error) {
	if p.client == nil {
		return nil, fmt.Errorf("provider not connected")
	}

	post, err := p.client.CreatePost(item.Title, item.Description)
	if err != nil {
		return nil, err
	}

	result := fiderPostToFeedbackItem(*post)
	return &result, nil
}

// Update updates an existing feedback item
func (p *FiderProvider) Update(item FeedbackItem) error {
	if p.client == nil {
		return fmt.Errorf("provider not connected")
	}

	// Fider API doesn't have direct update - would need to use different endpoint
	// For now, return not implemented
	return fmt.Errorf("update not implemented for Fider provider")
}

// Delete removes a feedback item
func (p *FiderProvider) Delete(id string) error {
	if p.client == nil {
		return fmt.Errorf("provider not connected")
	}

	// Fider API doesn't expose delete - admin only
	return fmt.Errorf("delete not implemented for Fider provider")
}

// fiderPostToFeedbackItem converts a FiderPost to FeedbackItem
func fiderPostToFeedbackItem(post FiderPost) FeedbackItem {
	// Map Fider tags to categories (use tag slug as category ID)
	var categories []string
	for _, tag := range post.Tags {
		categories = append(categories, tag.Slug)
	}

	return FeedbackItem{
		ID:          fmt.Sprintf("%d", post.Number),
		ExternalID:  fmt.Sprintf("%d", post.ID),
		Title:       post.Title,
		Description: post.Description,
		Status:      post.Status,
		Categories:  categories,
		Votes:       post.VotesCount,
		CreatedAt:   post.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Metadata: map[string]string{
			"slug":        post.Slug,
			"author_id":   fmt.Sprintf("%d", post.User.ID),
			"author_name": post.User.Name,
		},
	}
}

// feedbackItemToFiderPost converts a FeedbackItem to FiderCreatePost
func feedbackItemToFiderPost(item FeedbackItem) FiderCreatePost {
	return FiderCreatePost{
		Title:       item.Title,
		Description: item.Description,
	}
}

// Register the Fider provider
func init() {
	RegisterProvider("fider", NewFiderProvider)
}
