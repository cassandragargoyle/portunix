package components

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cassandragargoyle/portunix/app/wizard"
)

// MultiSelectComponent implements multiple selection from a list
type MultiSelectComponent struct {
	Title     string
	Prompt    string
	Options   []wizard.Option
	Selected  []string
	Help      string
	PageBreak bool
}

// NewMultiSelectComponent creates a new multi-select component
func NewMultiSelectComponent(title, prompt string, options []wizard.Option) *MultiSelectComponent {
	return &MultiSelectComponent{
		Title:   title,
		Prompt:  prompt,
		Options: options,
	}
}

// Render displays the multi-select component and captures user input
func (ms *MultiSelectComponent) Render(ctx *wizard.WizardContext) error {
	if ms.Title != "" {
		fmt.Printf("\n%s\n", formatTitle(ms.Title, ctx.Theme))
		fmt.Println(strings.Repeat("-", len(ms.Title)))
	}

	// Prepare options for survey
	var optionLabels []string
	optionMap := make(map[string]string)

	for _, opt := range ms.Options {
		label := opt.Label
		if opt.Description != "" {
			label = fmt.Sprintf("%s - %s", opt.Label, opt.Description)
		}
		optionLabels = append(optionLabels, label)
		optionMap[label] = opt.Value
	}

	prompt := &survey.MultiSelect{
		Message: ms.Prompt,
		Options: optionLabels,
		Help:    ms.Help,
	}

	var answers []string
	err := survey.AskOne(prompt, &answers)
	if err != nil {
		return err
	}

	// Map back to values
	ms.Selected = make([]string, len(answers))
	for i, answer := range answers {
		ms.Selected[i] = optionMap[answer]
	}

	return nil
}

// GetValue returns the selected values
func (ms *MultiSelectComponent) GetValue() interface{} {
	return ms.Selected
}

// SetValue sets the selected values
func (ms *MultiSelectComponent) SetValue(value interface{}) {
	if arr, ok := value.([]string); ok {
		ms.Selected = arr
	}
}

// Validate checks if selections have been made
func (ms *MultiSelectComponent) Validate() error {
	// Multi-select can be empty, so no validation needed by default
	return nil
}

// EvaluateCondition is a helper function for evaluating conditions
// This is exported so it can be used by the engine
func EvaluateCondition(condition string, variables map[string]interface{}) bool {
	parts := strings.Fields(condition)
	if len(parts) < 2 {
		return false
	}

	varName := parts[0]
	operator := parts[1]

	value, exists := variables[varName]
	if !exists {
		return false
	}

	switch operator {
	case "==":
		if len(parts) >= 3 {
			compareValue := parts[2]
			return fmt.Sprintf("%v", value) == compareValue
		}
	case "!=":
		if len(parts) >= 3 {
			compareValue := parts[2]
			return fmt.Sprintf("%v", value) != compareValue
		}
	case "true":
		if b, ok := value.(bool); ok {
			return b
		}
		return fmt.Sprintf("%v", value) == "true"
	case "false":
		if b, ok := value.(bool); ok {
			return !b
		}
		return fmt.Sprintf("%v", value) == "false"
	}

	// Handle simple boolean variable names
	if len(parts) == 1 {
		if b, ok := value.(bool); ok {
			return b
		}
		return fmt.Sprintf("%v", value) == "true"
	}

	return false
}
