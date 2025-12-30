package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ClearFlaskClient is a client for ClearFlask API
type ClearFlaskClient struct {
	BaseURL    string
	APIKey     string
	ProjectID  string
	HTTPClient *http.Client
}

// ClearFlaskUser represents a user in ClearFlask
type ClearFlaskUser struct {
	UserID      string `json:"userId"`
	Name        string `json:"name"`
	Email       string `json:"email,omitempty"`
	IsMod       bool   `json:"isMod,omitempty"`
	Created     string `json:"created,omitempty"`
}

// ClearFlaskCategory represents a category/board in ClearFlask
type ClearFlaskCategory struct {
	CategoryID  string `json:"categoryId"`
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
}

// ClearFlaskStatus represents a status in ClearFlask
type ClearFlaskStatus struct {
	StatusID    string `json:"statusId"`
	Name        string `json:"name"`
	Color       string `json:"color,omitempty"`
	NextStatusIDs []string `json:"nextStatusIds,omitempty"`
}

// ClearFlaskIdea represents a post/idea in ClearFlask
type ClearFlaskIdea struct {
	IdeaID       string              `json:"ideaId"`
	Title        string              `json:"title"`
	Description  string              `json:"description,omitempty"`
	Slug         string              `json:"slug,omitempty"`
	CategoryID   string              `json:"categoryId,omitempty"`
	StatusID     string              `json:"statusId,omitempty"`
	TagIDs       []string            `json:"tagIds,omitempty"`
	AuthorUserID string              `json:"authorUserId,omitempty"`
	VoteValue    int                 `json:"voteValue,omitempty"`
	VotersCount  int                 `json:"votersCount,omitempty"`
	Created      string              `json:"created,omitempty"`
	Edited       string              `json:"edited,omitempty"`
	Author       *ClearFlaskUser     `json:"author,omitempty"`
	Category     *ClearFlaskCategory `json:"category,omitempty"`
	Status       *ClearFlaskStatus   `json:"status,omitempty"`
}

// ClearFlaskIdeaCreate represents the request body for creating an idea
type ClearFlaskIdeaCreate struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	CategoryID  string   `json:"categoryId,omitempty"`
	TagIDs      []string `json:"tagIds,omitempty"`
}

// ClearFlaskIdeaUpdate represents the request body for updating an idea
type ClearFlaskIdeaUpdate struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	StatusID    string   `json:"statusId,omitempty"`
	CategoryID  string   `json:"categoryId,omitempty"`
	TagIDs      []string `json:"tagIds,omitempty"`
}

// ClearFlaskSearchResult represents search/list results
type ClearFlaskSearchResult struct {
	Results []ClearFlaskIdea `json:"results"`
	Cursor  string           `json:"cursor,omitempty"`
	Hits    int              `json:"hits,omitempty"`
}

// ClearFlaskError represents an error response from ClearFlask
type ClearFlaskError struct {
	UserFacingMessage string `json:"userFacingMessage,omitempty"`
	Message           string `json:"message,omitempty"`
	Code              string `json:"code,omitempty"`
}

// NewClearFlaskClient creates a new ClearFlask API client
func NewClearFlaskClient(baseURL, apiKey, projectID string) *ClearFlaskClient {
	return &ClearFlaskClient{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		ProjectID: projectID,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *ClearFlaskClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// ClearFlask uses x-cf-token for API authentication
	req.Header.Set("x-cf-token", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var cfErr ClearFlaskError
		if json.Unmarshal(respBody, &cfErr) == nil {
			if cfErr.UserFacingMessage != "" {
				return nil, fmt.Errorf("API error: %s", cfErr.UserFacingMessage)
			}
			if cfErr.Message != "" {
				return nil, fmt.Errorf("API error: %s", cfErr.Message)
			}
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// TestConnection tests if the API connection works
func (c *ClearFlaskClient) TestConnection() error {
	// Try to list ideas to verify connection
	_, err := c.doRequest("GET", fmt.Sprintf("/api/v1/projects/%s/ideas", c.ProjectID), nil)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	return nil
}

// ListIdeas returns all ideas from a ClearFlask project
func (c *ClearFlaskClient) ListIdeas() ([]ClearFlaskIdea, error) {
	var allIdeas []ClearFlaskIdea
	cursor := ""

	for {
		path := fmt.Sprintf("/api/v1/projects/%s/ideas", c.ProjectID)
		if cursor != "" {
			path += "?cursor=" + cursor
		}

		respBody, err := c.doRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}

		var result ClearFlaskSearchResult
		if err := json.Unmarshal(respBody, &result); err != nil {
			// Try parsing as direct array (fallback)
			var ideas []ClearFlaskIdea
			if err2 := json.Unmarshal(respBody, &ideas); err2 != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}
			return ideas, nil
		}

		allIdeas = append(allIdeas, result.Results...)

		// Check for pagination
		if result.Cursor == "" {
			break
		}
		cursor = result.Cursor
	}

	return allIdeas, nil
}

// GetIdea returns a specific idea by ID
func (c *ClearFlaskClient) GetIdea(ideaID string) (*ClearFlaskIdea, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/api/v1/projects/%s/ideas/%s", c.ProjectID, ideaID), nil)
	if err != nil {
		return nil, err
	}

	var idea ClearFlaskIdea
	if err := json.Unmarshal(respBody, &idea); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &idea, nil
}

// CreateIdea creates a new idea in ClearFlask
func (c *ClearFlaskClient) CreateIdea(title, description, categoryID string, tagIDs []string) (*ClearFlaskIdea, error) {
	reqBody := ClearFlaskIdeaCreate{
		Title:       title,
		Description: description,
		CategoryID:  categoryID,
		TagIDs:      tagIDs,
	}

	respBody, err := c.doRequest("POST", fmt.Sprintf("/api/v1/projects/%s/ideas", c.ProjectID), reqBody)
	if err != nil {
		return nil, err
	}

	var idea ClearFlaskIdea
	if err := json.Unmarshal(respBody, &idea); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &idea, nil
}

// UpdateIdea updates an existing idea
func (c *ClearFlaskClient) UpdateIdea(ideaID string, update ClearFlaskIdeaUpdate) error {
	_, err := c.doRequest("PATCH", fmt.Sprintf("/api/v1/projects/%s/ideas/%s", c.ProjectID, ideaID), update)
	return err
}

// DeleteIdea deletes an idea
func (c *ClearFlaskClient) DeleteIdea(ideaID string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/api/v1/projects/%s/ideas/%s", c.ProjectID, ideaID), nil)
	return err
}

// ListCategories returns all categories in the project
func (c *ClearFlaskClient) ListCategories() ([]ClearFlaskCategory, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/api/v1/projects/%s/categories", c.ProjectID), nil)
	if err != nil {
		return nil, err
	}

	var categories []ClearFlaskCategory
	if err := json.Unmarshal(respBody, &categories); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return categories, nil
}

// ListStatuses returns all statuses in the project
func (c *ClearFlaskClient) ListStatuses() ([]ClearFlaskStatus, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/api/v1/projects/%s/statuses", c.ProjectID), nil)
	if err != nil {
		return nil, err
	}

	var statuses []ClearFlaskStatus
	if err := json.Unmarshal(respBody, &statuses); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return statuses, nil
}
