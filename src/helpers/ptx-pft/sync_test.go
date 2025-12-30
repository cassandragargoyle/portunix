package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMarkdownFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test markdown file
	content := `# UC001: User Login Feature

## Summary
Allow users to log in with email and password

## Priority
High

## Status
Open

## Description
Users should be able to log in to the application using their email address and password.
The login form should validate inputs and provide clear error messages.
`
	filePath := filepath.Join(tmpDir, "UC001-user-login.md")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	item, err := ParseMarkdownFile(filePath)
	if err != nil {
		t.Fatalf("ParseMarkdownFile failed: %v", err)
	}

	// Verify parsed values
	if item.Title != "UC001: User Login Feature" {
		t.Errorf("Expected title 'UC001: User Login Feature', got '%s'", item.Title)
	}
	if item.Summary != "Allow users to log in with email and password" {
		t.Errorf("Expected summary 'Allow users to log in with email and password', got '%s'", item.Summary)
	}
	if item.Priority != "High" {
		t.Errorf("Expected priority 'High', got '%s'", item.Priority)
	}
	if item.Status != "Open" {
		t.Errorf("Expected status 'Open', got '%s'", item.Status)
	}
	if item.ID != "UC001-user-login" {
		t.Errorf("Expected ID 'UC001-user-login', got '%s'", item.ID)
	}
}

func TestScanFeedbackDirectory(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	files := map[string]string{
		"UC001-login.md": `# UC001: Login
## Summary
User login
## Status
Open`,
		"UC002-signup.md": `# UC002: Signup
## Summary
User signup
## Status
Planned`,
		"README.md": `# Not a feedback file`,
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", name, err)
		}
	}

	// Scan directory
	items, err := ScanFeedbackDirectory(tmpDir, "voc")
	if err != nil {
		t.Fatalf("ScanFeedbackDirectory failed: %v", err)
	}

	// Should find 3 markdown files (including README.md)
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// All items should have type "voc"
	for _, item := range items {
		if item.Type != "voc" {
			t.Errorf("Expected type 'voc', got '%s'", item.Type)
		}
	}
}

func TestScanFeedbackDirectoryEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	items, err := ScanFeedbackDirectory(tmpDir, "voc")
	if err != nil {
		t.Fatalf("ScanFeedbackDirectory failed: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(items))
	}
}

func TestScanFeedbackDirectoryNotExists(t *testing.T) {
	items, err := ScanFeedbackDirectory("/nonexistent/path", "voc")
	if err != nil {
		t.Fatalf("Should not error on nonexistent directory: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected 0 items for nonexistent directory, got %d", len(items))
	}
}

func TestGenerateMarkdownFromPost(t *testing.T) {
	post := &FiderPost{
		Number:      42,
		Title:       "Add dark mode",
		Description: "Please add dark mode support for better UX.\n\nThis would help reduce eye strain.",
		Status:      "planned",
		VotesCount:  15,
		User:        FiderUser{Name: "John Doe"},
	}

	markdown := GenerateMarkdownFromPost(post, "voc")

	// Check title
	if !contains(markdown, "# UC042: Add dark mode") {
		t.Error("Generated markdown should contain title with UC prefix")
	}

	// Check status
	if !contains(markdown, "## Status\nplanned") {
		t.Error("Generated markdown should contain status section")
	}

	// Check priority based on votes (15 >= 10 = High)
	if !contains(markdown, "## Priority\nHigh") {
		t.Error("Generated markdown should have High priority for 15 votes")
	}

	// Check metadata
	if !contains(markdown, "- Fider ID: 42") {
		t.Error("Generated markdown should contain Fider ID")
	}
}

func TestCreateSlugFromTitle(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Simple Title", "simple-title"},
		{"Hello World!", "hello-world"},
		{"Přidej tmavý režim", "přidej-tmavý-režim"},
		{"Test 123 Numbers", "test-123-numbers"},
		{"  Extra   Spaces  ", "extra-spaces"},
		{"Very Long Title That Should Be Truncated To Maximum Forty Characters", "very-long-title-that-should-be-truncated"},
	}

	for _, tc := range tests {
		result := CreateSlugFromTitle(tc.title)
		if result != tc.expected {
			t.Errorf("CreateSlugFromTitle(%q) = %q, expected %q", tc.title, result, tc.expected)
		}
	}
}

func TestFindNextAvailableNumber(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some existing files
	files := []string{
		"UC001-first.md",
		"UC002-second.md",
		"UC005-fifth.md", // Gap at 3,4
	}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		os.WriteFile(path, []byte("test"), 0644)
	}

	next := FindNextAvailableNumber(tmpDir, "UC")
	if next != 6 {
		t.Errorf("Expected next number 6, got %d", next)
	}
}

func TestFindNextAvailableNumberEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	next := FindNextAvailableNumber(tmpDir, "UC")
	if next != 1 {
		t.Errorf("Expected next number 1 for empty directory, got %d", next)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"Short", 10, "Short"},
		{"Exactly ten", 11, "Exactly ten"},
		{"This is a long string", 10, "This is a ..."},
		{"With\nnewlines", 20, "With newlines"},
	}

	for _, tc := range tests {
		result := truncate(tc.input, tc.maxLen)
		if result != tc.expected {
			t.Errorf("truncate(%q, %d) = %q, expected %q", tc.input, tc.maxLen, result, tc.expected)
		}
	}
}

func TestExtractFiderID(t *testing.T) {
	tmpDir := t.TempDir()

	// File with Fider ID
	contentWithID := `# Test
## Metadata
- Fider ID: 42
- Author: Test
`
	withIDPath := filepath.Join(tmpDir, "with-id.md")
	os.WriteFile(withIDPath, []byte(contentWithID), 0644)

	// File without Fider ID
	contentWithoutID := `# Test
## Summary
Just a summary
`
	withoutIDPath := filepath.Join(tmpDir, "without-id.md")
	os.WriteFile(withoutIDPath, []byte(contentWithoutID), 0644)

	// Test extraction
	id, found := ExtractFiderID(withIDPath)
	if !found {
		t.Error("Should find Fider ID in file")
	}
	if id != 42 {
		t.Errorf("Expected ID 42, got %d", id)
	}

	_, found = ExtractFiderID(withoutIDPath)
	if found {
		t.Error("Should not find Fider ID in file without one")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
