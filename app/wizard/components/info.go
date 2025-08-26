package components

import (
	"fmt"
	"strings"

	"github.com/cassandragargoyle/portunix/app/wizard"
)

// InfoComponent displays information without user input
type InfoComponent struct {
	Title   string
	Content string
}

// NewInfoComponent creates a new info component
func NewInfoComponent(title, content string) *InfoComponent {
	return &InfoComponent{
		Title:   title,
		Content: content,
	}
}

// Render displays the information
func (i *InfoComponent) Render(ctx *wizard.WizardContext) error {
	if i.Title != "" {
		title := expandVariables(i.Title, ctx.Variables)
		fmt.Printf("\n%s\n", formatTitle(title, ctx.Theme))
		fmt.Println(strings.Repeat("=", len(title)))
	}

	if i.Content != "" {
		content := expandVariables(i.Content, ctx.Variables)
		fmt.Printf("%s\n", content)
	}

	fmt.Print("\nPress Enter to continue...")
	fmt.Scanln()
	return nil
}

// GetValue returns nil as info doesn't collect values
func (i *InfoComponent) GetValue() interface{} {
	return nil
}

// SetValue does nothing for info component
func (i *InfoComponent) SetValue(value interface{}) {}

// Validate always returns nil for info
func (i *InfoComponent) Validate() error {
	return nil
}

// SuccessComponent displays success message
type SuccessComponent struct {
	InfoComponent
}

// NewSuccessComponent creates a new success component
func NewSuccessComponent(title, content string) *SuccessComponent {
	return &SuccessComponent{
		InfoComponent: InfoComponent{
			Title:   title,
			Content: content,
		},
	}
}

// Render displays the success message with styling
func (s *SuccessComponent) Render(ctx *wizard.WizardContext) error {
	if s.Title != "" {
		title := expandVariables(s.Title, ctx.Variables)
		fmt.Printf("\n🎉 %s\n", formatTitle(title, ctx.Theme))
		fmt.Println(strings.Repeat("=", len(title)+4))
	}

	if s.Content != "" {
		content := expandVariables(s.Content, ctx.Variables)
		fmt.Printf("✅ %s\n", content)
	}

	fmt.Print("\nPress Enter to finish...")
	fmt.Scanln()
	return nil
}

// ErrorComponent displays error message
type ErrorComponent struct {
	InfoComponent
	Error error
}

// NewErrorComponent creates a new error component
func NewErrorComponent(title, content string, err error) *ErrorComponent {
	return &ErrorComponent{
		InfoComponent: InfoComponent{
			Title:   title,
			Content: content,
		},
		Error: err,
	}
}

// Render displays the error message with styling
func (e *ErrorComponent) Render(ctx *wizard.WizardContext) error {
	if e.Title != "" {
		title := expandVariables(e.Title, ctx.Variables)
		fmt.Printf("\n❌ %s\n", formatTitle(title, ctx.Theme))
		fmt.Println(strings.Repeat("=", len(title)+4))
	}

	if e.Content != "" {
		content := expandVariables(e.Content, ctx.Variables)
		fmt.Printf("💥 %s\n", content)
	}

	if e.Error != nil {
		fmt.Printf("Error details: %v\n", e.Error)
	}

	fmt.Print("\nPress Enter to exit...")
	fmt.Scanln()
	return nil
}