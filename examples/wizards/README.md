# Wizard Examples

This directory contains example wizard definitions that demonstrate the capabilities of the Portunix Wizard Framework.

## Available Wizards

### 1. database-setup.yaml
A comprehensive database installation wizard that:
- Guides users through selecting database type (PostgreSQL, MySQL, MongoDB, Redis)
- Configures database name with validation
- Offers installation options (backups, monitoring, sample data)
- Provides conditional flow based on user choices
- Executes installation tasks with progress tracking

**Usage:**
```bash
portunix wizard run database-setup
```

### 2. dev-environment.yaml
A development environment setup wizard that:
- Helps choose programming language (Python, Java, Go, JavaScript)
- Selects IDE/editor (VS Code, IntelliJ, Vim)
- Sets up project structure with Git initialization
- Configures language-specific frameworks
- Provides tailored setup based on language choice

**Usage:**
```bash
portunix wizard run dev-environment
```

## Features Demonstrated

### UI Components
- **Info pages**: Welcome screens and information display
- **Select**: Single choice from options list
- **Multi-select**: Multiple choice selection with checkboxes
- **Input**: Text input with validation rules
- **Confirm**: Yes/No confirmation prompts
- **Progress**: Task execution with progress tracking
- **Success**: Completion screens with summary

### Advanced Features
- **Variable interpolation**: Use {{variable}} syntax in content
- **Conditional navigation**: Branch to different pages based on user input
- **Input validation**: Required fields, length limits, regex patterns
- **Task execution**: Run commands with progress tracking
- **Conditional tasks**: Execute tasks only when conditions are met

### Navigation Examples

**Simple navigation:**
```yaml
next:
  page: "next-page-id"
```

**Conditional navigation:**
```yaml
next:
  condition: "variable_name == value"
  true: "page-if-true"
  false: "page-if-false"
```

**Boolean condition:**
```yaml
next:
  condition: "enable_feature == true"
  true: "feature-config"
  false: "skip-feature"
```

### Validation Examples

**Required field with pattern:**
```yaml
validate:
  required: true
  min_length: 3
  pattern: "^[a-zA-Z][a-zA-Z0-9_]*$"
  message: "Name must start with letter and contain only letters, numbers, and underscores"
```

### Task Execution Examples

**Tasks with conditions and weights:**
```yaml
tasks:
  - id: "main-task"
    label: "Installing {{software_name}}"
    command: "portunix install {{software_name}}"
    weight: 50
    
  - id: "optional-task"
    label: "Loading sample data"
    command: "echo 'Loading data'"
    condition: "sample_data == true"
    weight: 30
```

## Creating Custom Wizards

1. **Create a new wizard:**
   ```bash
   portunix wizard create my-wizard
   ```

2. **Edit the generated YAML file:**
   - Define variables in the `variables` section
   - Create pages with appropriate types
   - Set up navigation rules
   - Add validation where needed

3. **Validate your wizard:**
   ```bash
   portunix wizard validate my-wizard.yaml
   ```

4. **Test your wizard:**
   ```bash
   portunix wizard run my-wizard.yaml
   ```

## Best Practices

1. **Use descriptive IDs**: Page and variable IDs should be clear and meaningful
2. **Provide helpful descriptions**: Add descriptions to options to guide users
3. **Validate input**: Always validate user input for critical fields
4. **Handle errors gracefully**: Provide clear error messages and recovery options
5. **Test thoroughly**: Test all navigation paths and edge cases
6. **Keep it simple**: Break complex workflows into smaller, manageable steps

## Wizard Structure

```yaml
wizard:
  id: "unique-wizard-id"
  name: "Human Readable Name"
  version: "1.0"
  description: "Brief description of what this wizard does"
  
  variables:
    # Define default values for variables
    
  pages:
    - id: "page-id"
      type: "page-type"
      title: "Page Title"
      # ... page-specific configuration
```

## Available Page Types

- `info`: Display information
- `select`: Single selection from list
- `multi_select`: Multiple selections from list  
- `input`: Text input
- `password`: Password input (masked)
- `confirm`: Yes/No confirmation
- `progress`: Task execution with progress
- `success`: Success message
- `error`: Error message

For more details on page types and their configuration options, see the main wizard framework documentation.