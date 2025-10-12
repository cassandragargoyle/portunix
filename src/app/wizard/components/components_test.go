package components

import (
	"testing"

	"portunix.ai/app/wizard"
)

func TestSelectComponent_GetSetValue(t *testing.T) {
	options := []wizard.Option{
		{Value: "option1", Label: "Option 1"},
		{Value: "option2", Label: "Option 2"},
	}

	comp := NewSelectComponent("Test Select", "Choose option:", options)

	// Test initial state
	if comp.GetValue() != "" {
		t.Errorf("Expected empty initial value, got '%v'", comp.GetValue())
	}

	// Test setting value
	comp.SetValue("option1")
	if comp.GetValue() != "option1" {
		t.Errorf("Expected 'option1', got '%v'", comp.GetValue())
	}

	// Test validation
	err := comp.Validate()
	if err != nil {
		t.Errorf("Validation should pass with selected value: %v", err)
	}

	// Test validation with no selection
	comp.SetValue("")
	err = comp.Validate()
	if err == nil {
		t.Error("Expected validation error for empty selection")
	}
}

func TestInputComponent_Validation(t *testing.T) {
	comp := NewInputComponent("Test Input", "Enter value:")

	tests := []struct {
		name       string
		validation *wizard.ValidationRule
		value      string
		shouldFail bool
	}{
		{
			name:       "No validation",
			validation: nil,
			value:      "any value",
			shouldFail: false,
		},
		{
			name: "Required field - valid",
			validation: &wizard.ValidationRule{
				Required: true,
			},
			value:      "some value",
			shouldFail: false,
		},
		{
			name: "Required field - invalid",
			validation: &wizard.ValidationRule{
				Required: true,
			},
			value:      "",
			shouldFail: true,
		},
		{
			name: "Min length - valid",
			validation: &wizard.ValidationRule{
				MinLen: 5,
			},
			value:      "hello",
			shouldFail: false,
		},
		{
			name: "Min length - invalid",
			validation: &wizard.ValidationRule{
				MinLen: 5,
			},
			value:      "hi",
			shouldFail: true,
		},
		{
			name: "Max length - valid",
			validation: &wizard.ValidationRule{
				MaxLen: 10,
			},
			value:      "hello",
			shouldFail: false,
		},
		{
			name: "Max length - invalid",
			validation: &wizard.ValidationRule{
				MaxLen: 5,
			},
			value:      "hello world",
			shouldFail: true,
		},
		{
			name: "Pattern - valid",
			validation: &wizard.ValidationRule{
				Pattern: "^[a-z]+$",
			},
			value:      "hello",
			shouldFail: false,
		},
		{
			name: "Pattern - invalid",
			validation: &wizard.ValidationRule{
				Pattern: "^[a-z]+$",
			},
			value:      "Hello123",
			shouldFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			comp.Validation = test.validation
			comp.SetValue(test.value)

			err := comp.Validate()

			if test.shouldFail && err == nil {
				t.Error("Expected validation to fail, but it passed")
			}

			if !test.shouldFail && err != nil {
				t.Errorf("Expected validation to pass, but got error: %v", err)
			}
		})
	}
}

func TestConfirmComponent_GetSetValue(t *testing.T) {
	comp := NewConfirmComponent("Test Confirm", "Continue?")

	// Test initial state
	if comp.GetValue() != false {
		t.Errorf("Expected false initial value, got '%v'", comp.GetValue())
	}

	// Test setting value
	comp.SetValue(true)
	if comp.GetValue() != true {
		t.Errorf("Expected true, got '%v'", comp.GetValue())
	}

	// Test validation (should always pass)
	err := comp.Validate()
	if err != nil {
		t.Errorf("Confirm validation should always pass: %v", err)
	}
}

func TestMultiSelectComponent_GetSetValue(t *testing.T) {
	options := []wizard.Option{
		{Value: "option1", Label: "Option 1"},
		{Value: "option2", Label: "Option 2"},
		{Value: "option3", Label: "Option 3"},
	}

	comp := NewMultiSelectComponent("Test MultiSelect", "Choose options:", options)

	// Test initial state
	if comp.GetValue() == nil {
		t.Error("Expected non-nil initial value")
	}

	// Test setting values
	testValues := []string{"option1", "option3"}
	comp.SetValue(testValues)

	result := comp.GetValue().([]string)
	if len(result) != 2 {
		t.Errorf("Expected 2 selected values, got %d", len(result))
	}

	if result[0] != "option1" || result[1] != "option3" {
		t.Errorf("Expected ['option1', 'option3'], got %v", result)
	}

	// Test validation (should always pass for multi-select)
	err := comp.Validate()
	if err != nil {
		t.Errorf("MultiSelect validation should always pass: %v", err)
	}
}

func TestEvaluateCondition(t *testing.T) {
	variables := map[string]interface{}{
		"str_var":   "hello",
		"bool_var":  true,
		"false_var": false,
		"num_var":   42,
	}

	tests := []struct {
		condition string
		expected  bool
	}{
		{"str_var == hello", true},
		{"str_var == world", false},
		{"str_var != world", true},
		{"bool_var == true", true},
		{"bool_var == false", false},
		{"bool_var true", true},
		{"false_var false", true},
		{"false_var true", false},
		{"bool_var", true},
		{"false_var", false},
		{"nonexistent", false},
		{"", false},
		{"invalid condition format", false},
	}

	for _, test := range tests {
		t.Run(test.condition, func(t *testing.T) {
			result := EvaluateCondition(test.condition, variables)
			if result != test.expected {
				t.Errorf("Condition '%s': expected %v, got %v", test.condition, test.expected, result)
			}
		})
	}
}
