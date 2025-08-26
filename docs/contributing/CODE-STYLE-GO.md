# Go Code Style Guidelines

## Purpose
This document defines the Go coding standards for CassandraGargoyle projects, based on established patterns from the Portunix project and Go community best practices.

## General Principles

### 1. Follow Go Conventions
- Use `go fmt` for all code formatting
- Follow standard Go naming conventions
- Use `gofmt`, `go vet`, and `golangci-lint` for code quality
- Embrace Go idioms: "Don't communicate by sharing memory; share memory by communicating"

### 2. Code Organization
- Use Go modules for dependency management
- Organize code into logical packages
- Keep packages focused and cohesive
- Avoid circular dependencies

## File and Package Naming

### File Naming
Use **snake_case** with descriptive prefixes:
```
✅ install_linux.go
✅ install_windows.go  
✅ docker_test.go
✅ config_manager.go
❌ installLinux.go
❌ Docker_Test.go
```

### Platform-Specific Files
Use build tags and descriptive suffixes:
```go
// +build linux
// install_linux.go

// +build windows  
// install_windows.go
```

### Package Names
- Use lowercase, single-word names when possible
- Use descriptive names that reflect the package purpose
- Examples: `install`, `docker`, `datastore`, `plugins`

## Package Structure

### Recommended Hierarchy
```
project/
├── main.go                 # Application entry point
├── cmd/                    # Command implementations (cobra commands)
│   ├── root.go
│   ├── install.go
│   └── version.go
├── app/                    # Core application logic
│   ├── install/           # Installation functionality
│   │   ├── install.go
│   │   ├── apt/
│   │   └── chocolatey/
│   ├── docker/            # Docker management
│   └── datastore/         # Data persistence
├── pkg/                    # Public packages for external use
├── internal/              # Private packages
└── test/                  # Test fixtures and utilities
```

### Module Organization
Use local module replacements for large projects:
```go
// go.mod
replace project.com/app => ./app
replace project.com/cmd => ./cmd
```

## Naming Conventions

### Functions and Methods
- **Exported**: PascalCase (`Install`, `GetOSName`, `LoadConfig`)
- **Unexported**: camelCase (`loadDefaultConfig`, `parseVersion`)
- Use descriptive verbs: `Get`, `Set`, `Load`, `Save`, `Create`, `Delete`

### Variables
- **Constants**: PascalCase or ALL_CAPS for package constants
- **Package variables**: PascalCase for exported, camelCase for unexported
- **Local variables**: Concise camelCase

```go
// Constants
const DefaultTimeout = 30 * time.Second
const MAX_RETRY_COUNT = 3

// Package variables
var DefaultInstallConfig *Config
var version = \"v1.0.0\"

// Local variables
func processRequest() {
    execDir := getExecutableDir()
    configPath := filepath.Join(execDir, \"config.yaml\")
}
```

### Interfaces
- Use descriptive names with meaningful suffixes
- Prefer small, focused interfaces
- Use `Interface` suffix for comprehensive interfaces

```go
type DatastoreInterface interface {
    Store(ctx context.Context, key string, value interface{}) error
    Retrieve(ctx context.Context, key string) (interface{}, error)
}

type Writer interface {
    Write([]byte) (int, error)
}
```

## Import Organization

### Standard Import Grouping
```go
import (
    // Standard library
    \"context\"
    \"fmt\"
    \"os\"
    \"path/filepath\"
    
    // Third-party packages
    \"github.com/spf13/cobra\"
    \"gopkg.in/yaml.v3\"
    
    // Local packages
    \"project.com/app/install\"
    \"project.com/app/datastore\"
)
```

### Import Aliases
Use aliases for clarity and conflict resolution:
```go
import (
    \"context\"
    appConfig \"project.com/app/config\"
    installConfig \"project.com/app/install/config\"
)
```

## Error Handling

### Standard Patterns
1. **Early Return**: Handle errors immediately and return early
2. **Error Wrapping**: Use `fmt.Errorf` with `%w` verb
3. **Nil Checks**: Always check for nil before dereferencing

```go
// Early return with error wrapping
func loadConfiguration(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf(\"failed to read config file %s: %w\", path, err)
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf(\"failed to parse config: %w\", err)
    }
    
    return &config, nil
}

// Optional operations (don't fail on missing optional configs)
func loadOptionalConfig() *Config {
    config, err := loadUserConfig()
    if err != nil {
        // User config is optional, use defaults
        return getDefaultConfig()
    }
    return config
}
```

### Error Types
Create custom error types for domain-specific errors:
```go
type ConfigError struct {
    Field string
    Value string
    Err   error
}

func (e ConfigError) Error() string {
    return fmt.Sprintf(\"invalid config field '%s' with value '%s': %v\", e.Field, e.Value, e.Err)
}

func (e ConfigError) Unwrap() error {
    return e.Err
}
```

## Struct and Interface Design

### Struct Definition
```go
type PluginConfig struct {
    Name            string            `json:\"name\" yaml:\"name\"`
    Version         string            `json:\"version\" yaml:\"version\"`
    BinaryPath      string            `json:\"binary_path\" yaml:\"binary_path\"`
    Environment     map[string]string `json:\"environment,omitempty\" yaml:\"environment,omitempty\"`
    Permissions     PluginPermissions `json:\"permissions\" yaml:\"permissions\"`
    
    // Unexported fields for internal state
    isLoaded bool
    loadTime time.Time
}
```

### Interface Design Best Practices
- Keep interfaces small and focused
- Use context.Context as first parameter for long-running operations
- Return meaningful errors
- Use consistent return patterns

```go
type PackageManager interface {
    Install(ctx context.Context, pkg string) error
    Update(ctx context.Context, pkg string) error
    Remove(ctx context.Context, pkg string) error
    List(ctx context.Context) ([]Package, error)
}
```

## Comments and Documentation

### Package Documentation
```go
// Package install provides cross-platform software installation capabilities.
// 
// This package supports multiple package managers including APT, Chocolatey,
// and Homebrew, with automatic OS detection and fallback mechanisms.
package install
```

### Function Documentation
```go
// InstallPackage installs the specified package using the appropriate package manager.
// It returns an error if the installation fails or if no suitable package manager is found.
//
// The function automatically detects the operating system and chooses the best
// available package manager for the platform.
func InstallPackage(ctx context.Context, packageName string) error {
    // implementation
}
```

### Inline Comments
- Explain complex logic, not obvious code
- Use comments to explain \"why\", not \"what\"
- Keep comments up to date with code changes

```go
// Check if we're running in Windows Sandbox environment
// This affects installation paths and permissions
if runtime.GOOS == \"windows\" && isWindowsSandbox() {
    installPath = getSandboxSafePath(installPath)
}
```

## Code Patterns and Idioms

### Configuration Pattern
```go
type Config struct {
    // Embed default configuration
    *DefaultConfig
    
    // Override specific fields
    Timeout    time.Duration `yaml:\"timeout\"`
    RetryCount int           `yaml:\"retry_count\"`
}

func LoadConfig(path string) (*Config, error) {
    config := &Config{
        DefaultConfig: getDefaultConfig(),
    }
    
    if data, err := os.ReadFile(path); err == nil {
        _ = yaml.Unmarshal(data, config)
    }
    
    return config, nil
}
```

### String Enumeration Pattern
```go
type Status int

const (
    StatusUnknown Status = iota
    StatusRunning
    StatusStopped
    StatusError
)

func (s Status) String() string {
    switch s {
    case StatusRunning:
        return \"running\"
    case StatusStopped:
        return \"stopped\"
    case StatusError:
        return \"error\"
    default:
        return \"unknown\"
    }
}
```

### Factory Pattern
```go
type PackageManager interface {
    Install(pkg string) error
}

func NewPackageManager() PackageManager {
    switch runtime.GOOS {
    case \"linux\":
        return NewAptManager()
    case \"windows\":
        return NewChocolateyManager()
    case \"darwin\":
        return NewHomebrewManager()
    default:
        return NewGenericManager()
    }
}
```

## Testing Guidelines

### Test File Organization
- Use `*_test.go` suffix
- Keep tests in the same package (prefer whitebox testing)
- Use `internal/testutils` for shared test utilities

### Test Function Naming
```go
func TestInstallPackage(t *testing.T) {
    // Basic functionality test
}

func TestInstallPackage_WithInvalidPackage_ReturnsError(t *testing.T) {
    // Error case test
}

func BenchmarkInstallPackage(b *testing.B) {
    // Performance test
}
```

### Test Structure
```go
func TestFunction(t *testing.T) {
    // Arrange
    manager := NewPackageManager()
    packageName := \"test-package\"
    
    // Act
    err := manager.Install(packageName)
    
    // Assert
    if err != nil {
        t.Errorf(\"expected no error, got %v\", err)
    }
}
```

## Build and Deployment

### Build Tags
Use build tags for platform-specific code:
```go
// +build linux darwin

package unix

// +build windows

package windows
```

### Embedded Assets
Use `//go:embed` for static assets:
```go
//go:embed assets/*
var assets embed.FS

func getAsset(name string) ([]byte, error) {
    return assets.ReadFile(\"assets/\" + name)
}
```

## Security Best Practices

### Input Validation
- Validate all external input
- Use path.Clean() for file paths
- Sanitize command arguments

```go
func InstallFromPath(userPath string) error {
    // Clean and validate the path
    cleanPath := filepath.Clean(userPath)
    if !strings.HasPrefix(cleanPath, \"/safe/install/\") {
        return errors.New(\"invalid installation path\")
    }
    
    // Proceed with installation
    return install(cleanPath)
}
```

### Secrets Management
- Never log sensitive information
- Use environment variables for secrets
- Clear sensitive data from memory when possible

## Performance Guidelines

### Memory Management
- Use sync.Pool for frequently allocated objects
- Prefer slices over arrays for flexibility
- Use streaming for large data processing

### Goroutine Management
- Always provide context for cancellation
- Use worker pools for bounded concurrency
- Close channels to signal completion

```go
func processFiles(ctx context.Context, files []string) error {
    const maxWorkers = 10
    sem := make(chan struct{}, maxWorkers)
    
    var wg sync.WaitGroup
    for _, file := range files {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case sem <- struct{}{}:
            wg.Add(1)
            go func(f string) {
                defer wg.Done()
                defer func() { <-sem }()
                processFile(f)
            }(file)
        }
    }
    
    wg.Wait()
    return nil
}
```

## Tools and Linting

### Required Tools
- `gofmt` - Code formatting
- `go vet` - Static analysis
- `golangci-lint` - Comprehensive linting
- `go mod tidy` - Dependency management

### Recommended golangci-lint Configuration
```yaml
# .golangci.yml
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - gocyclo
    - gofmt
    - misspell
```

### Pre-commit Hooks
```bash
#!/bin/sh
# .git/hooks/pre-commit
go fmt ./...
go vet ./...
golangci-lint run
go test -short ./...
```

## Migration Guide

### Updating Existing Code
When updating existing code to follow these guidelines:

1. Run `go fmt` to fix formatting
2. Update naming conventions gradually
3. Add missing error handling
4. Improve documentation
5. Add tests for critical functions

### Legacy Code
- Document deviations from guidelines
- Create issues for technical debt
- Refactor incrementally during feature work

---

**Note**: These guidelines are based on the established patterns in the Portunix project and Go community standards. They should be adapted as the team's practices evolve.

*Created: 2025-08-23*
*Last updated: 2025-08-23*