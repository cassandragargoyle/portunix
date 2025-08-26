package wizard

import (
	"time"
)

// PageType defines the type of wizard page
type PageType string

const (
	PageTypeInfo        PageType = "info"
	PageTypeSelect      PageType = "select"
	PageTypeMultiSelect PageType = "multi_select"
	PageTypeInput       PageType = "input"
	PageTypePassword    PageType = "password"
	PageTypeConfirm     PageType = "confirm"
	PageTypeProgress    PageType = "progress"
	PageTypeSuccess     PageType = "success"
	PageTypeError       PageType = "error"
)

// Wizard represents a complete wizard definition
type Wizard struct {
	ID          string                 `yaml:"id"`
	Name        string                 `yaml:"name"`
	Version     string                 `yaml:"version"`
	Description string                 `yaml:"description"`
	Variables   map[string]interface{} `yaml:"variables"`
	Pages       []Page                 `yaml:"pages"`
}

// Page represents a single page in the wizard
type Page struct {
	ID       string                 `yaml:"id"`
	Type     PageType               `yaml:"type"`
	Title    string                 `yaml:"title"`
	Content  string                 `yaml:"content,omitempty"`
	Prompt   string                 `yaml:"prompt,omitempty"`
	Options  []Option               `yaml:"options,omitempty"`
	Variable string                 `yaml:"variable,omitempty"`
	Next     *NavigationRule        `yaml:"next,omitempty"`
	Tasks    []Task                 `yaml:"tasks,omitempty"`
	Validate *ValidationRule        `yaml:"validate,omitempty"`
}

// Option represents a selectable option
type Option struct {
	Value       string `yaml:"value"`
	Label       string `yaml:"label"`
	Description string `yaml:"description,omitempty"`
	Variable    string `yaml:"variable,omitempty"`
}

// NavigationRule defines conditional navigation
type NavigationRule struct {
	// Simple navigation - just a page ID
	Page string `yaml:"page,omitempty"`
	
	// Conditional navigation
	Condition string `yaml:"condition,omitempty"`
	True      string `yaml:"true,omitempty"`
	False     string `yaml:"false,omitempty"`
}

// Task represents a task to be executed
type Task struct {
	ID        string  `yaml:"id"`
	Label     string  `yaml:"label"`
	Command   string  `yaml:"command"`
	Condition string  `yaml:"condition,omitempty"`
	Weight    float64 `yaml:"weight,omitempty"`
}

// ValidationRule defines input validation
type ValidationRule struct {
	Required bool   `yaml:"required,omitempty"`
	MinLen   int    `yaml:"min_length,omitempty"`
	MaxLen   int    `yaml:"max_length,omitempty"`
	Pattern  string `yaml:"pattern,omitempty"`
	Message  string `yaml:"message,omitempty"`
}

// WizardContext maintains the current state of wizard execution
type WizardContext struct {
	Wizard      *Wizard
	Variables   map[string]interface{}
	CurrentPage string
	History     []string
	StartTime   time.Time
	Theme       *Theme
}

// Theme defines visual styling for the wizard
type Theme struct {
	Name           string
	PrimaryColor   string
	SecondaryColor string
	ErrorColor     string
	SuccessColor   string
	InfoColor      string
	ProgressStyle  string
	SpinnerStyle   string
}

// WizardResult contains the result of wizard execution
type WizardResult struct {
	Completed bool
	Variables map[string]interface{}
	Duration  time.Duration
	Error     error
}

// Component interface for all UI components
type Component interface {
	Render(ctx *WizardContext) error
	GetValue() interface{}
	SetValue(interface{})
	Validate() error
}