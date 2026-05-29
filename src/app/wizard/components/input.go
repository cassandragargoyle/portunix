package components

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"portunix.ai/app/wizard"
)

// InputComponent handles text input
type InputComponent struct {
	Title      string
	Prompt     string
	Default    string
	Value      string
	Help       string
	Validation *wizard.ValidationRule
}

// NewInputComponent creates a new input component
func NewInputComponent(title, prompt string) *InputComponent {
	return &InputComponent{
		Title:  title,
		Prompt: prompt,
	}
}

// Render displays the input component
func (i *InputComponent) Render(ctx *wizard.WizardContext) error {
	if i.Title != "" {
		fmt.Printf("\n%s\n", formatTitle(i.Title, ctx.Theme))
		fmt.Println(strings.Repeat("-", len(i.Title)))
	}

	if ctx.NonInteractive {
		if i.Value == "" {
			i.Value = i.Default
		}
		if err := i.Validate(); err != nil {
			return fmt.Errorf("non-interactive: %w", err)
		}
		fmt.Printf("%s %s\n", i.Prompt, i.Value)
		return nil
	}

	prompt := &survey.Input{
		Message: i.Prompt,
		Default: i.Default,
		Help:    i.Help,
	}

	var answer string
	err := survey.AskOne(prompt, &answer, survey.WithValidator(i.createValidator()))
	if err != nil {
		return err
	}

	i.Value = answer
	return nil
}

// GetValue returns the input value
func (i *InputComponent) GetValue() interface{} {
	return i.Value
}

// SetValue sets the input value
func (i *InputComponent) SetValue(value interface{}) {
	if str, ok := value.(string); ok {
		i.Value = str
	}
}

// Validate validates the input
func (i *InputComponent) Validate() error {
	if i.Validation == nil {
		return nil
	}

	if i.Validation.Required && i.Value == "" {
		if i.Validation.Message != "" {
			return fmt.Errorf(i.Validation.Message)
		}
		return fmt.Errorf("this field is required")
	}

	if i.Validation.MinLen > 0 && len(i.Value) < i.Validation.MinLen {
		return fmt.Errorf("minimum length is %d characters", i.Validation.MinLen)
	}

	if i.Validation.MaxLen > 0 && len(i.Value) > i.Validation.MaxLen {
		return fmt.Errorf("maximum length is %d characters", i.Validation.MaxLen)
	}

	if i.Validation.Pattern != "" {
		matched, err := regexp.MatchString(i.Validation.Pattern, i.Value)
		if err != nil {
			return err
		}
		if !matched {
			if i.Validation.Message != "" {
				return fmt.Errorf(i.Validation.Message)
			}
			return fmt.Errorf("invalid format")
		}
	}

	return nil
}

// createValidator creates a survey validator from validation rules
func (i *InputComponent) createValidator() survey.Validator {
	return func(val interface{}) error {
		str := fmt.Sprintf("%v", val)
		i.Value = str
		return i.Validate()
	}
}

// PasswordComponent handles password input
type PasswordComponent struct {
	InputComponent
}

// NewPasswordComponent creates a new password component
func NewPasswordComponent(title, prompt string) *PasswordComponent {
	return &PasswordComponent{
		InputComponent: InputComponent{
			Title:  title,
			Prompt: prompt,
		},
	}
}

// Render displays the password component
func (p *PasswordComponent) Render(ctx *wizard.WizardContext) error {
	if p.Title != "" {
		fmt.Printf("\n%s\n", formatTitle(p.Title, ctx.Theme))
		fmt.Println(strings.Repeat("-", len(p.Title)))
	}

	if ctx.NonInteractive {
		if p.Value == "" {
			return fmt.Errorf("non-interactive: missing value for password '%s'", p.Prompt)
		}
		if err := p.Validate(); err != nil {
			return fmt.Errorf("non-interactive: %w", err)
		}
		fmt.Printf("%s ********\n", p.Prompt)
		return nil
	}

	prompt := &survey.Password{
		Message: p.Prompt,
		Help:    p.Help,
	}

	var answer string
	err := survey.AskOne(prompt, &answer, survey.WithValidator(p.createValidator()))
	if err != nil {
		return err
	}

	p.Value = answer
	return nil
}

// ConfirmComponent handles yes/no confirmation
type ConfirmComponent struct {
	Title   string
	Prompt  string
	Default bool
	Value   bool
	Help    string
}

// NewConfirmComponent creates a new confirm component
func NewConfirmComponent(title, prompt string) *ConfirmComponent {
	return &ConfirmComponent{
		Title:  title,
		Prompt: prompt,
	}
}

// Render displays the confirm component
func (c *ConfirmComponent) Render(ctx *wizard.WizardContext) error {
	if c.Title != "" {
		fmt.Printf("\n%s\n", formatTitle(c.Title, ctx.Theme))
		fmt.Println(strings.Repeat("-", len(c.Title)))
	}

	if ctx.NonInteractive {
		// In non-interactive mode, c.Value was set by the engine from
		// preloaded variables; if not, fall back to Default.
		fmt.Printf("%s %v\n", c.Prompt, c.Value)
		return nil
	}

	prompt := &survey.Confirm{
		Message: c.Prompt,
		Default: c.Default,
		Help:    c.Help,
	}

	var answer bool
	err := survey.AskOne(prompt, &answer)
	if err != nil {
		return err
	}

	c.Value = answer
	return nil
}

// GetValue returns the confirmation value
func (c *ConfirmComponent) GetValue() interface{} {
	return c.Value
}

// SetValue sets the confirmation value
func (c *ConfirmComponent) SetValue(value interface{}) {
	if b, ok := value.(bool); ok {
		c.Value = b
	}
}

// Validate always returns nil for confirm
func (c *ConfirmComponent) Validate() error {
	return nil
}
