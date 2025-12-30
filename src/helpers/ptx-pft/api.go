package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FiderClient is a client for Fider.io API
type FiderClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// FiderUser represents a user in Fider
type FiderUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// FiderTag represents a tag in Fider
type FiderTag struct {
	ID       int    `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	IsPublic bool   `json:"isPublic"`
}

// FiderPost represents a post/idea in Fider
type FiderPost struct {
	ID          int        `json:"id"`
	Number      int        `json:"number"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	User        FiderUser  `json:"user"`
	VotesCount  int        `json:"votesCount"`
	CreatedAt   time.Time  `json:"createdAt"`
	Tags        []FiderTag `json:"tags,omitempty"`
}

// FiderCreatePost represents the request body for creating a post
type FiderCreatePost struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// FiderError represents an error response from Fider
type FiderError struct {
	Errors []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"errors"`
}

// NewFiderClient creates a new Fider API client
func NewFiderClient(baseURL, apiKey string) *FiderClient {
	return &FiderClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *FiderClient) doRequest(method, path string, body interface{}) ([]byte, error) {
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

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

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
		var fiderErr FiderError
		if json.Unmarshal(respBody, &fiderErr) == nil && len(fiderErr.Errors) > 0 {
			return nil, fmt.Errorf("API error: %s", fiderErr.Errors[0].Message)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// CreatePost creates a new post/idea in Fider
func (c *FiderClient) CreatePost(title, description string) (*FiderPost, error) {
	reqBody := FiderCreatePost{
		Title:       title,
		Description: description,
	}

	respBody, err := c.doRequest("POST", "/api/v1/posts", reqBody)
	if err != nil {
		return nil, err
	}

	var post FiderPost
	if err := json.Unmarshal(respBody, &post); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &post, nil
}

// ListPosts returns all posts from Fider
func (c *FiderClient) ListPosts() ([]FiderPost, error) {
	respBody, err := c.doRequest("GET", "/api/v1/posts", nil)
	if err != nil {
		return nil, err
	}

	var posts []FiderPost
	if err := json.Unmarshal(respBody, &posts); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return posts, nil
}

// GetPost returns a specific post by number
func (c *FiderClient) GetPost(number int) (*FiderPost, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/api/v1/posts/%d", number), nil)
	if err != nil {
		return nil, err
	}

	var post FiderPost
	if err := json.Unmarshal(respBody, &post); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &post, nil
}

// TestConnection tests if the API connection works
func (c *FiderClient) TestConnection() error {
	_, err := c.doRequest("GET", "/api/v1/posts", nil)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	return nil
}

// ListUsers returns all users from Fider
func (c *FiderClient) ListUsers() ([]FiderUser, error) {
	respBody, err := c.doRequest("GET", "/api/v1/users", nil)
	if err != nil {
		return nil, err
	}

	var users []FiderUser
	if err := json.Unmarshal(respBody, &users); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return users, nil
}

// GetUser returns a specific user by ID
func (c *FiderClient) GetUser(id int) (*FiderUser, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/api/v1/users/%d", id), nil)
	if err != nil {
		return nil, err
	}

	var user FiderUser
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &user, nil
}
