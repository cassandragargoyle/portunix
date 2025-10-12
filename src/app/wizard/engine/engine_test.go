package engine

import (
	"os"
	"path/filepath"
	"testing"

	"portunix.ai/app/wizard"
)

func TestWizardEngine_LoadWizard(t *testing.T) {
	engine := NewWizardEngine()

	// Create temporary wizard file
	tempDir := t.TempDir()
	wizardFile := filepath.Join(tempDir, "test-wizard.yaml")

	wizardContent := `wizard:
  id: "test-wizard"
  name: "Test Wizard"
  version: "1.0"
  description: "Test wizard for unit tests"
  
  variables:
    test_var: "default_value"
    
  pages:
    - id: "welcome"
      type: "info"
      title: "Welcome"
      content: "Welcome to test wizard"
      next:
        page: "complete"
    
    - id: "complete"
      type: "success"
      title: "Complete"
      content: "Test completed"
`

	err := os.WriteFile(wizardFile, []byte(wizardContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test wizard file: %v", err)
	}

	// Test loading wizard
	wizard, err := engine.LoadWizard(wizardFile)
	if err != nil {
		t.Fatalf("LoadWizard failed: %v", err)
	}

	// Verify wizard properties
	if wizard.ID != "test-wizard" {
		t.Errorf("Expected wizard ID 'test-wizard', got '%s'", wizard.ID)
	}

	if wizard.Name != "Test Wizard" {
		t.Errorf("Expected wizard name 'Test Wizard', got '%s'", wizard.Name)
	}

	if wizard.Version != "1.0" {
		t.Errorf("Expected wizard version '1.0', got '%s'", wizard.Version)
	}

	if len(wizard.Pages) != 2 {
		t.Errorf("Expected 2 pages, got %d", len(wizard.Pages))
	}

	// Test variables
	if wizard.Variables["test_var"] != "default_value" {
		t.Errorf("Expected test_var to be 'default_value', got '%v'", wizard.Variables["test_var"])
	}
}

func TestWizardEngine_LoadWizard_InvalidFile(t *testing.T) {
	engine := NewWizardEngine()

	// Test non-existent file
	_, err := engine.LoadWizard("non-existent-file.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	// Test invalid YAML
	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid.yaml")

	invalidContent := `invalid: yaml: content:
  - missing
    - indentation
`

	err = os.WriteFile(invalidFile, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid YAML file: %v", err)
	}

	_, err = engine.LoadWizard(invalidFile)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestWizardEngine_FindPage(t *testing.T) {
	engine := NewWizardEngine()

	wizard := &wizard.Wizard{
		Pages: []wizard.Page{
			{ID: "page1", Title: "Page 1"},
			{ID: "page2", Title: "Page 2"},
		},
	}

	// Test finding existing page
	page := engine.findPage(wizard, "page1")
	if page == nil {
		t.Error("Expected to find page1, got nil")
	} else if page.ID != "page1" {
		t.Errorf("Expected page ID 'page1', got '%s'", page.ID)
	}

	// Test finding non-existent page
	page = engine.findPage(wizard, "non-existent")
	if page != nil {
		t.Error("Expected nil for non-existent page, got page")
	}
}

func TestWizardEngine_EvaluateNavigation(t *testing.T) {
	engine := NewWizardEngine()

	ctx := &wizard.WizardContext{
		Variables: map[string]interface{}{
			"choice":         "option1",
			"enable_feature": true,
		},
	}

	tests := []struct {
		name     string
		nav      *wizard.NavigationRule
		expected string
	}{
		{
			name: "Simple navigation",
			nav: &wizard.NavigationRule{
				Page: "next-page",
			},
			expected: "next-page",
		},
		{
			name: "Conditional navigation - true",
			nav: &wizard.NavigationRule{
				Condition: "enable_feature == true",
				True:      "feature-page",
				False:     "skip-page",
			},
			expected: "feature-page",
		},
		{
			name: "Conditional navigation - false",
			nav: &wizard.NavigationRule{
				Condition: "choice == option2",
				True:      "option2-page",
				False:     "default-page",
			},
			expected: "default-page",
		},
		{
			name:     "Nil navigation",
			nav:      nil,
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := engine.evaluateNavigation(test.nav, ctx)
			if err != nil {
				t.Errorf("evaluateNavigation failed: %v", err)
			}
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestWizardEngine_Themes(t *testing.T) {
	engine := NewWizardEngine()

	// Test default themes exist
	themes := engine.ListThemes()
	expectedThemes := []string{"default", "colorful", "minimal"}

	for _, expected := range expectedThemes {
		found := false
		for _, theme := range themes {
			if theme == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected theme '%s' not found in %v", expected, themes)
		}
	}

	// Test setting valid theme
	err := engine.SetTheme("colorful")
	if err != nil {
		t.Errorf("SetTheme failed for valid theme: %v", err)
	}

	// Test setting invalid theme
	err = engine.SetTheme("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent theme, got nil")
	}
}
