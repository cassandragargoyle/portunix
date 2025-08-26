package components

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cassandragargoyle/portunix/app/wizard"
)

// SelectComponent implements single selection from a list
type SelectComponent struct {
	Title       string
	Prompt      string
	Options     []wizard.Option
	Selected    string
	Help        string
	PageBreak   bool
}

// NewSelectComponent creates a new select component
func NewSelectComponent(title, prompt string, options []wizard.Option) *SelectComponent {
	return &SelectComponent{
		Title:   title,
		Prompt:  prompt,
		Options: options,
	}
}

// Render displays the select component and captures user input
func (s *SelectComponent) Render(ctx *wizard.WizardContext) error {
	if s.Title != "" {
		fmt.Printf("\n%s\n", formatTitle(s.Title, ctx.Theme))
		fmt.Println(strings.Repeat("-", len(s.Title)))
	}

	// Prepare options for survey
	var optionLabels []string
	optionMap := make(map[string]string)
	
	for _, opt := range s.Options {
		label := opt.Label
		if opt.Description != "" {
			label = fmt.Sprintf("%s - %s", opt.Label, opt.Description)
		}
		optionLabels = append(optionLabels, label)
		optionMap[label] = opt.Value
	}

	prompt := &survey.Select{
		Message: s.Prompt,
		Options: optionLabels,
		Help:    s.Help,
	}

	var answer string
	err := survey.AskOne(prompt, &answer)
	if err != nil {
		return err
	}

	// Map back to value
	s.Selected = optionMap[answer]
	return nil
}

// GetValue returns the selected value
func (s *SelectComponent) GetValue() interface{} {
	return s.Selected
}

// SetValue sets the selected value
func (s *SelectComponent) SetValue(value interface{}) {
	if str, ok := value.(string); ok {
		s.Selected = str
	}
}

// Validate checks if a selection has been made
func (s *SelectComponent) Validate() error {
	if s.Selected == "" {
		return fmt.Errorf("no selection made")
	}
	return nil
}

// formatTitle formats the title with theme colors
func formatTitle(title string, theme *wizard.Theme) string {
	if theme == nil {
		return title
	}
	// In a real implementation, apply color formatting based on theme
	return title
}