package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewFiderClient(t *testing.T) {
	client := NewFiderClient("http://localhost:3000", "test-api-key")

	if client.BaseURL != "http://localhost:3000" {
		t.Errorf("Expected BaseURL 'http://localhost:3000', got '%s'", client.BaseURL)
	}
	if client.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey 'test-api-key', got '%s'", client.APIKey)
	}
	if client.HTTPClient == nil {
		t.Error("HTTPClient should not be nil")
	}
}

func TestListPosts(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/api/v1/posts" {
			t.Errorf("Expected path '/api/v1/posts', got '%s'", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got '%s'", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected 'Bearer test-key', got '%s'", r.Header.Get("Authorization"))
		}

		// Return mock response
		posts := []FiderPost{
			{
				ID:          1,
				Number:      1,
				Title:       "Test Post 1",
				Slug:        "test-post-1",
				Description: "Description 1",
				Status:      "open",
				VotesCount:  5,
				CreatedAt:   time.Now(),
			},
			{
				ID:          2,
				Number:      2,
				Title:       "Test Post 2",
				Slug:        "test-post-2",
				Description: "Description 2",
				Status:      "planned",
				VotesCount:  10,
				CreatedAt:   time.Now(),
			},
		}
		json.NewEncoder(w).Encode(posts)
	}))
	defer server.Close()

	client := NewFiderClient(server.URL, "test-key")
	posts, err := client.ListPosts()

	if err != nil {
		t.Fatalf("ListPosts failed: %v", err)
	}
	if len(posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(posts))
	}
	if posts[0].Title != "Test Post 1" {
		t.Errorf("Expected 'Test Post 1', got '%s'", posts[0].Title)
	}
	if posts[1].VotesCount != 10 {
		t.Errorf("Expected 10 votes, got %d", posts[1].VotesCount)
	}
}

func TestGetPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/posts/42" {
			t.Errorf("Expected path '/api/v1/posts/42', got '%s'", r.URL.Path)
		}

		post := FiderPost{
			ID:          42,
			Number:      42,
			Title:       "Specific Post",
			Slug:        "specific-post",
			Description: "Detailed description",
			Status:      "completed",
			VotesCount:  25,
		}
		json.NewEncoder(w).Encode(post)
	}))
	defer server.Close()

	client := NewFiderClient(server.URL, "test-key")
	post, err := client.GetPost(42)

	if err != nil {
		t.Fatalf("GetPost failed: %v", err)
	}
	if post.Number != 42 {
		t.Errorf("Expected number 42, got %d", post.Number)
	}
	if post.Title != "Specific Post" {
		t.Errorf("Expected 'Specific Post', got '%s'", post.Title)
	}
}

func TestCreatePost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got '%s'", r.Method)
		}
		if r.URL.Path != "/api/v1/posts" {
			t.Errorf("Expected path '/api/v1/posts', got '%s'", r.URL.Path)
		}

		// Parse request body
		var reqBody FiderCreatePost
		json.NewDecoder(r.Body).Decode(&reqBody)

		if reqBody.Title != "New Feature" {
			t.Errorf("Expected title 'New Feature', got '%s'", reqBody.Title)
		}

		// Return created post
		post := FiderPost{
			ID:          100,
			Number:      100,
			Title:       reqBody.Title,
			Slug:        "new-feature",
			Description: reqBody.Description,
			Status:      "open",
			VotesCount:  0,
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(post)
	}))
	defer server.Close()

	client := NewFiderClient(server.URL, "test-key")
	post, err := client.CreatePost("New Feature", "Feature description")

	if err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}
	if post.Number != 100 {
		t.Errorf("Expected number 100, got %d", post.Number)
	}
	if post.Title != "New Feature" {
		t.Errorf("Expected 'New Feature', got '%s'", post.Title)
	}
}

func TestListUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/users" {
			t.Errorf("Expected path '/api/v1/users', got '%s'", r.URL.Path)
		}

		users := []FiderUser{
			{ID: 1, Name: "John Doe", Email: "john@example.com"},
			{ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
		}
		json.NewEncoder(w).Encode(users)
	}))
	defer server.Close()

	client := NewFiderClient(server.URL, "test-key")
	users, err := client.ListUsers()

	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
	if users[0].Name != "John Doe" {
		t.Errorf("Expected 'John Doe', got '%s'", users[0].Name)
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(FiderError{
			Errors: []struct {
				Field   string `json:"field"`
				Message string `json:"message"`
			}{
				{Field: "authorization", Message: "Invalid API key"},
			},
		})
	}))
	defer server.Close()

	client := NewFiderClient(server.URL, "invalid-key")
	_, err := client.ListPosts()

	if err == nil {
		t.Error("Expected error for unauthorized request")
	}
	if err.Error() != "API error: Invalid API key" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestTestConnection(t *testing.T) {
	// Test successful connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]FiderPost{})
	}))
	defer server.Close()

	client := NewFiderClient(server.URL, "test-key")
	err := client.TestConnection()

	if err != nil {
		t.Errorf("TestConnection should succeed: %v", err)
	}
}

func TestTestConnectionFailure(t *testing.T) {
	client := NewFiderClient("http://localhost:99999", "test-key")
	err := client.TestConnection()

	if err == nil {
		t.Error("Expected connection failure")
	}
}
