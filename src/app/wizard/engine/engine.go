package engine

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"portunix.ai/app/wizard"
	"portunix.ai/app/wizard/components"
)

// WizardEngine implements the main wizard execution engine
type WizardEngine struct {
	themes         map[string]*wizard.Theme
	wizards        map[string]*wizard.Wizard
	nonInteractive bool
	preloadedVars  map[string]interface{}
}

// NewWizardEngine creates a new wizard engine
func NewWizardEngine() *WizardEngine {
	engine := &WizardEngine{
		themes:        make(map[string]*wizard.Theme),
		wizards:       make(map[string]*wizard.Wizard),
		preloadedVars: make(map[string]interface{}),
	}
	engine.loadDefaultThemes()
	return engine
}

// SetNonInteractive enables non-interactive mode. Components must source
// their values from preloaded variables instead of prompting the user.
func (e *WizardEngine) SetNonInteractive(enabled bool) {
	e.nonInteractive = enabled
}

// LoadVariablesFromConfig loads YAML/JSON config of variable values for
// non-interactive execution. The file must contain a top-level mapping
// of variable name to value.
func (e *WizardEngine) LoadVariablesFromConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	vars := make(map[string]interface{})
	if err := yaml.Unmarshal(data, &vars); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	for k, v := range vars {
		e.preloadedVars[k] = v
	}
	return nil
}

// LoadWizard loads a wizard from a YAML file
func (e *WizardEngine) LoadWizard(yamlPath string) (*wizard.Wizard, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wizard file: %w", err)
	}

	return e.RegisterWizardBytes(data)
}

// RegisterWizardBytes parses a wizard definition from raw YAML bytes and
// registers it in the engine. Used by plugin discovery and embedded wizards.
func (e *WizardEngine) RegisterWizardBytes(data []byte) (*wizard.Wizard, error) {
	var wizardDef struct {
		Wizard wizard.Wizard `yaml:"wizard"`
	}

	if err := yaml.Unmarshal(data, &wizardDef); err != nil {
		return nil, fmt.Errorf("failed to parse wizard YAML: %w", err)
	}

	if wizardDef.Wizard.ID == "" {
		return nil, fmt.Errorf("wizard YAML missing required 'id' field")
	}

	if wizardDef.Wizard.Variables == nil {
		wizardDef.Wizard.Variables = make(map[string]interface{})
	}

	e.wizards[wizardDef.Wizard.ID] = &wizardDef.Wizard
	return &wizardDef.Wizard, nil
}

// ListLoadedWizards returns all wizards currently registered with the engine
func (e *WizardEngine) ListLoadedWizards() map[string]*wizard.Wizard {
	out := make(map[string]*wizard.Wizard, len(e.wizards))
	for k, v := range e.wizards {
		out[k] = v
	}
	return out
}

// ExecuteWizard executes a wizard
func (e *WizardEngine) ExecuteWizard(wiz *wizard.Wizard) (*wizard.WizardResult, error) {
	ctx := &wizard.WizardContext{
		Wizard:         wiz,
		Variables:      make(map[string]interface{}),
		CurrentPage:    "",
		History:        []string{},
		StartTime:      time.Now(),
		Theme:          e.themes["default"],
		NonInteractive: e.nonInteractive,
	}

	// Copy initial variables, then overlay preloaded values from --config
	for k, v := range wiz.Variables {
		ctx.Variables[k] = v
	}
	for k, v := range e.preloadedVars {
		ctx.Variables[k] = v
	}

	result := &wizard.WizardResult{
		Completed: false,
		Variables: ctx.Variables,
		Duration:  0,
		Error:     nil,
	}

	// Find first page
	if len(wiz.Pages) == 0 {
		result.Error = fmt.Errorf("wizard has no pages")
		return result, result.Error
	}

	ctx.CurrentPage = wiz.Pages[0].ID

	// Execute wizard loop
	for ctx.CurrentPage != "" {
		page := e.findPage(wiz, ctx.CurrentPage)
		if page == nil {
			result.Error = fmt.Errorf("page not found: %s", ctx.CurrentPage)
			return result, result.Error
		}

		// Add to history
		ctx.History = append(ctx.History, ctx.CurrentPage)

		// Create and render component
		component, err := e.createComponent(page)
		if err != nil {
			result.Error = err
			return result, err
		}

		// In non-interactive mode, pre-fill the component from preloaded vars
		// so its Render() can validate & print instead of prompting.
		if ctx.NonInteractive && page.Variable != "" {
			if v, ok := ctx.Variables[page.Variable]; ok {
				component.SetValue(v)
			}
		}

		err = component.Render(ctx)
		if err != nil {
			result.Error = err
			return result, err
		}

		// Store component value in variables
		if page.Variable != "" {
			ctx.Variables[page.Variable] = component.GetValue()
		}

		// Handle multi-select options that set variables
		if page.Type == wizard.PageTypeMultiSelect {
			for _, option := range page.Options {
				if option.Variable != "" {
					// Check if option was selected
					if selected, ok := component.GetValue().([]string); ok {
						for _, sel := range selected {
							if sel == option.Value {
								ctx.Variables[option.Variable] = true
								break
							}
						}
						// Set to false if not selected
						if _, exists := ctx.Variables[option.Variable]; !exists {
							ctx.Variables[option.Variable] = false
						}
					}
				}
			}
		}

		// Determine next page
		nextPage, err := e.evaluateNavigation(page.Next, ctx)
		if err != nil {
			result.Error = err
			return result, err
		}

		ctx.CurrentPage = nextPage
	}

	result.Completed = true
	result.Duration = time.Since(ctx.StartTime)
	result.Variables = ctx.Variables
	return result, nil
}

// createComponent creates a component based on page type
func (e *WizardEngine) createComponent(page *wizard.Page) (wizard.Component, error) {
	switch page.Type {
	case wizard.PageTypeInfo:
		return components.NewInfoComponent(page.Title, page.Content), nil

	case wizard.PageTypeSelect:
		return components.NewSelectComponent(page.Title, page.Prompt, page.Options), nil

	case wizard.PageTypeMultiSelect:
		return components.NewMultiSelectComponent(page.Title, page.Prompt, page.Options), nil

	case wizard.PageTypeInput:
		comp := components.NewInputComponent(page.Title, page.Prompt)
		comp.Validation = page.Validate
		return comp, nil

	case wizard.PageTypePassword:
		comp := components.NewPasswordComponent(page.Title, page.Prompt)
		comp.Validation = page.Validate
		return comp, nil

	case wizard.PageTypeConfirm:
		return components.NewConfirmComponent(page.Title, page.Prompt), nil

	case wizard.PageTypeProgress:
		return components.NewProgressComponent(page.Title, page.Tasks), nil

	case wizard.PageTypeSuccess:
		return components.NewSuccessComponent(page.Title, page.Content), nil

	case wizard.PageTypeError:
		return components.NewErrorComponent(page.Title, page.Content, nil), nil

	default:
		return nil, fmt.Errorf("unknown page type: %s", page.Type)
	}
}

// findPage finds a page by ID
func (e *WizardEngine) findPage(wiz *wizard.Wizard, pageID string) *wizard.Page {
	for i := range wiz.Pages {
		if wiz.Pages[i].ID == pageID {
			return &wiz.Pages[i]
		}
	}
	return nil
}

// evaluateNavigation determines the next page based on navigation rules
func (e *WizardEngine) evaluateNavigation(nav *wizard.NavigationRule, ctx *wizard.WizardContext) (string, error) {
	if nav == nil {
		return "", nil // End of wizard
	}

	// Simple navigation
	if nav.Page != "" {
		return nav.Page, nil
	}

	// Conditional navigation
	if nav.Condition != "" {
		result := e.evaluateCondition(nav.Condition, ctx.Variables)
		if result {
			return nav.True, nil
		} else {
			return nav.False, nil
		}
	}

	return "", nil
}

// evaluateCondition evaluates a simple condition
func (e *WizardEngine) evaluateCondition(condition string, variables map[string]interface{}) bool {
	// This is a simplified condition evaluator
	// In a real implementation, you might want to use a proper expression parser
	return components.EvaluateCondition(condition, variables)
}

// loadDefaultThemes loads built-in themes
func (e *WizardEngine) loadDefaultThemes() {
	e.themes["default"] = &wizard.Theme{
		Name:           "default",
		PrimaryColor:   "cyan",
		SecondaryColor: "white",
		ErrorColor:     "red",
		SuccessColor:   "green",
		InfoColor:      "blue",
		ProgressStyle:  "bar",
		SpinnerStyle:   "dots",
	}

	e.themes["colorful"] = &wizard.Theme{
		Name:           "colorful",
		PrimaryColor:   "magenta",
		SecondaryColor: "yellow",
		ErrorColor:     "red",
		SuccessColor:   "green",
		InfoColor:      "blue",
		ProgressStyle:  "bar",
		SpinnerStyle:   "line",
	}

	e.themes["minimal"] = &wizard.Theme{
		Name:           "minimal",
		PrimaryColor:   "white",
		SecondaryColor: "white",
		ErrorColor:     "white",
		SuccessColor:   "white",
		InfoColor:      "white",
		ProgressStyle:  "simple",
		SpinnerStyle:   "dots",
	}
}

// SetTheme sets the current theme
func (e *WizardEngine) SetTheme(themeName string) error {
	if theme, exists := e.themes[themeName]; exists {
		// Apply theme to current context
		_ = theme
		return nil
	}
	return fmt.Errorf("theme not found: %s", themeName)
}

// ListThemes returns available themes
func (e *WizardEngine) ListThemes() []string {
	var themes []string
	for name := range e.themes {
		themes = append(themes, name)
	}
	return themes
}
