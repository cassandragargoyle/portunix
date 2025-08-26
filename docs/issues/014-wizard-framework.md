# Issue #014: Wizard Framework for Interactive CLI Installation

## Summary
Implement a comprehensive wizard framework in Portunix core that provides reusable CLI components for creating interactive, guided installation experiences. This framework will enable both core features and plugins to create user-friendly, step-by-step installation wizards with rich CLI components like progress bars, spinners, selection menus, and conditional flows.

## Problem Statement
Currently, Portunix installations and configurations require users to remember and type complex commands with multiple parameters. This creates a steep learning curve and increases the chance of errors. Users need a guided, interactive experience that walks them through installation processes step-by-step, similar to GUI installers but in the terminal environment.

## Proposed Solution

### Core Components

#### 1. CLI UI Components Library

**Input Components:**
- **Single Select**: Choose one option from a list
- **Multi Select**: Choose multiple options with checkboxes
- **Text Input**: Get text input with validation
- **Password Input**: Secure password entry with masking
- **Confirmation**: Yes/No prompts
- **Number Input**: Numeric input with range validation

**Display Components:**
- **Progress Bar**: Show installation progress
- **Spinner**: Animated spinners (dots, line, snake, etc.)
- **Status Messages**: Info, warning, error, success messages
- **Tables**: Display structured data
- **Panels**: Bordered content sections
- **Breadcrumbs**: Show current position in wizard

**Animation Components:**
- **Loading Animations**: Various loading indicators
- **Task List**: Show running tasks with status
- **Live Log Output**: Stream logs with formatting

#### 2. Wizard Definition Format (YAML)

```yaml
wizard:
  id: "database-setup"
  name: "Database Installation Wizard"
  version: "1.0"
  description: "Install and configure database systems"
  
  variables:
    db_type: ""
    db_name: ""
    enable_backup: false
    sample_data: false
  
  pages:
    - id: "welcome"
      type: "info"
      title: "Database Setup Wizard"
      content: |
        This wizard will help you install and configure a database.
        
        We'll guide you through:
        • Selecting database type
        • Configuration options
        • Initial setup
      next: "db_selection"
    
    - id: "db_selection"
      type: "select"
      title: "Choose Database Type"
      prompt: "Which database would you like to install?"
      options:
        - value: "postgresql"
          label: "PostgreSQL - Advanced open-source relational database"
          description: "Best for: Complex queries, ACID compliance"
        - value: "mysql"
          label: "MySQL - Popular open-source database"
          description: "Best for: Web applications, WordPress"
        - value: "mongodb"
          label: "MongoDB - Document-oriented NoSQL database"
          description: "Best for: Flexible schemas, JSON data"
      variable: "db_type"
      next: "db_options"
    
    - id: "db_options"
      type: "multi_select"
      title: "Installation Options"
      prompt: "Select additional options:"
      options:
        - value: "backup"
          label: "Enable automatic backups"
          variable: "enable_backup"
        - value: "monitoring"
          label: "Install monitoring tools"
        - value: "sample"
          label: "Load sample data"
          variable: "sample_data"
      next:
        condition: "sample_data == true"
        true: "sample_confirmation"
        false: "installation"
    
    - id: "sample_confirmation"
      type: "confirm"
      title: "Confirm Sample Data"
      prompt: "Sample data will use ~500MB. Continue?"
      variable: "confirm_sample"
      next:
        condition: "confirm_sample"
        true: "installation"
        false: "db_options"
    
    - id: "installation"
      type: "progress"
      title: "Installing {{db_type}}"
      tasks:
        - id: "download"
          label: "Downloading {{db_type}}"
          command: "portunix install {{db_type}}"
          weight: 40
        - id: "configure"
          label: "Configuring database"
          command: "portunix db configure {{db_type}}"
          weight: 30
        - id: "sample_data"
          label: "Loading sample data"
          command: "portunix db load-sample {{db_type}}"
          condition: "sample_data == true"
          weight: 30
      next: "complete"
    
    - id: "complete"
      type: "success"
      title: "Installation Complete!"
      content: |
        ✅ {{db_type}} has been successfully installed!
        
        Connection details:
        • Host: localhost
        • Port: {{db_port}}
        • Database: {{db_name}}
        
        To connect: portunix db connect {{db_type}}
```

#### 3. Wizard Engine API

```go
// Core wizard interfaces
type WizardEngine interface {
    LoadWizard(yamlPath string) (*Wizard, error)
    ExecuteWizard(wizard *Wizard) (*WizardResult, error)
    RegisterCustomComponent(name string, component Component)
}

type Component interface {
    Render(context *WizardContext) error
    GetValue() interface{}
    Validate() error
}

type WizardContext struct {
    Variables map[string]interface{}
    CurrentPage string
    History []string
    Theme *Theme
}

// Component examples
type SelectComponent struct {
    Title string
    Options []Option
    Selected int
}

type ProgressComponent struct {
    Tasks []Task
    CurrentTask int
    Progress float64
}
```

### Features

#### 1. Built-in Themes
- **Default**: Clean, professional appearance
- **Colorful**: Rich colors and icons
- **Minimal**: Simple, distraction-free
- **Accessible**: High contrast, screen-reader friendly

#### 2. Conditional Logic
- Variable-based branching
- Expression evaluation (e.g., `db_size > 100`)
- Dynamic page generation
- Skip conditions
- Loop support for repeated tasks

#### 3. Validation
- Input validation rules
- Custom validators
- Real-time validation feedback
- Dependency checking
- Pre-flight checks before installation

#### 4. State Management
- Save and resume wizard progress
- Rollback capability
- Variable persistence
- History navigation (back/forward)

#### 5. Integration Features
- Hook system for custom logic
- Plugin wizard registration
- Wizard composition (embed wizards)
- External command execution
- MCP tool integration

### CLI Usage Examples

```bash
# Run a wizard directly
portunix wizard run database-setup

# Run with preset values
portunix wizard run database-setup --preset production

# List available wizards
portunix wizard list

# Create wizard from template
portunix wizard create my-setup

# Validate wizard definition
portunix wizard validate setup.yaml

# Run in non-interactive mode
portunix wizard run database-setup --non-interactive --config setup.json
```

### Plugin Integration

Plugins can register their own wizards:

```go
// In plugin code
func RegisterWizards(engine WizardEngine) {
    engine.RegisterWizard("plugin-name", wizardYAML)
    engine.RegisterCustomComponent("special-input", &CustomInput{})
}
```

### Example Use Cases

#### Database Setup Wizard
- Select database type
- Choose version
- Configure connection settings
- Set up initial users
- Enable features (replication, backup)

#### Development Environment Wizard
- Choose programming language
- Select IDE/editor
- Install language tools
- Configure version managers
- Set up linters/formatters

#### Project Initialization Wizard
- Select project type
- Choose framework
- Configure dependencies
- Set up Git repository
- Initialize CI/CD

### Technical Implementation

#### Libraries to Consider
- **Survey**: Interactive prompts for Go
- **Bubble Tea**: Terminal UI framework
- **Cobra**: Enhanced command integration
- **YAML**: Configuration parsing
- **Expression evaluation**: For conditional logic

#### Performance Considerations
- Lazy loading of wizard pages
- Efficient terminal rendering
- Minimal CPU usage during idle
- Responsive to terminal resize

### Testing Requirements
- Unit tests for all components
- Integration tests for wizard flows
- Terminal emulation tests
- Cross-platform testing (Windows Terminal, iTerm2, etc.)
- Accessibility testing

### Documentation Requirements
- Wizard creation guide
- Component reference
- YAML schema documentation
- Best practices guide
- Example wizard library

## Acceptance Criteria
1. All core UI components implemented and working
2. YAML-based wizard definition fully functional
3. Conditional flow logic working correctly
4. Progress tracking and state management operational
5. At least 3 example wizards included
6. Plugin integration API documented and tested
7. Cross-platform compatibility verified
8. Theme system implemented with at least 3 themes
9. Non-interactive mode available for automation

## Dependencies
- Core Portunix CLI framework
- Terminal UI library selection
- YAML parser

## Priority
High - Wizards will significantly improve user experience across all Portunix features

## Estimated Effort
Medium-Large (2-3 weeks)

## Labels
- enhancement
- cli
- wizard
- framework
- user-experience
- core