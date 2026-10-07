# Issue #215: Universal Container Command Implementation

> **Renumbered:** formerly internal issue #029. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

**Status**: 📋 Open  
**Priority**: High  
**Type**: Enhancement  
**Labels**: container, docker, podman, configuration, universal-interface  
**Requested by**: portunix-portunixcredits development team  
**Created**: 2025-01-09  

## Problem Statement

Currently, Portunix requires users to choose explicitly between `portunix docker` and `portunix podman` commands, creating friction in development workflows. Users need to know which container runtime is available and make manual decisions about which command to use.

### Current Limitations
- Separate commands for Docker and Podman operations
- No unified interface for container operations
- No automatic container runtime selection
- No configuration-based container runtime preferences
- Users must manually determine which runtime to use

### Impact on Development Teams
- **Workflow complexity**: Different commands for different environments
- **Cross-platform issues**: Docker on Windows/macOS vs Podman on Linux
- **CI/CD complexity**: Different scripts for different container runtimes
- **Learning curve**: Users must learn both command sets

## Requirements

### Core Features

1. **Universal Container Command**
   - `portunix container` as unified interface
   - Automatic delegation to appropriate runtime (Docker/Podman)
   - All existing functionality from both runtimes available
   - Transparent parameter translation

2. **Configuration-Based Runtime Selection**
   - YAML configuration file support
   - Runtime preference configuration
   - Fallback logic for missing configuration
   - Default to Podman when no configuration present

3. **Configuration File Structure**
```yaml
# ~/.portunix/config.yaml
container_runtime: podman  # Options: docker, podman
    
# Other global settings
verbose: false
auto_update: true
```

### Configuration Logic

#### Runtime Selection Priority
1. **Explicit configuration**: Use `container_runtime` value if set
2. **Default fallback**: Use Podman if no configuration
3. **Availability check**: If configured runtime unavailable, show error

#### Configuration File Locations
- Primary: `./portunix-config.yaml` (project directory - for development)
- Secondary: `~/.portunix/config.yaml` (user directory)
- Tertiary: `~/.config/portunix/config.yaml` (Linux standard)
- Fallback: Built-in defaults (Podman preference)

## Technical Implementation

### 1. Configuration System
```go
// app/config/config.go
type Config struct {
    ContainerRuntime string `yaml:"container_runtime"` // docker, podman
    Verbose          bool   `yaml:"verbose"`
    AutoUpdate       bool   `yaml:"auto_update"`
}
```

### 2. Universal Container Command
```go
// cmd/container.go
var containerCmd = &cobra.Command{
    Use:   "container",
    Short: "Universal container management interface",
    Long: `Universal container management that automatically selects 
between Docker and Podman based on configuration and availability.`,
}

var containerRunCmd = &cobra.Command{
    Use:   "run-in-container [installation-type] [args...]",
    Short: "Run installation in container using configured runtime",
    Run: func(cmd *cobra.Command, args []string) {
        runtime, err := config.GetSelectedRuntime()
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }
        
        // Delegate to appropriate runtime
        switch runtime {
        case "docker":
            runDockerContainer(args)
        case "podman":
            runPodmanContainer(args)
        }
    },
}
```

### 3. Runtime Detection and Selection
```go
// app/container/runtime.go
func GetSelectedRuntime() (string, error) {
    config := LoadConfig()
    
    // Use configured runtime or default to podman
    runtime := config.ContainerRuntime
    if runtime == "" {
        runtime = "podman"
    }
    
    // Check if selected runtime is available
    switch runtime {
    case "docker":
        if IsDockerAvailable() {
            return "docker", nil
        }
        return "", errors.New("Docker not available")
    case "podman":
        if IsPodmanAvailable() {
            return "podman", nil
        }
        return "", errors.New("Podman not available")
    default:
        return "", fmt.Errorf("Unknown container runtime: %s", runtime)
    }
}
```

## Command Structure

### New Universal Commands
```bash
# Universal container interface
portunix container run-in-container [installation-type] [args...]
portunix container list
portunix container stop [name]
portunix container remove [name]
portunix container logs [name]
portunix container exec [name] [command]

# Configuration management
portunix config set container_runtime docker
portunix config get container_runtime
portunix config show
```

### Backward Compatibility
- All existing `portunix docker` and `portunix podman` commands remain functional
- No breaking changes to existing workflows
- New universal commands are additive

## Implementation Phases

### Phase 1: Configuration System (Week 1)
- [ ] Implement YAML configuration loading
- [ ] Create configuration structure and validation
- [ ] Add configuration file management
- [ ] Implement runtime detection logic

### Phase 2: Universal Container Command (Week 2)
- [ ] Create `portunix container` base command
- [ ] Implement `container run-in-container` with runtime delegation
- [ ] Add parameter passthrough to selected runtime
- [ ] Create runtime selection logic

### Phase 3: Additional Container Commands (Week 3)
- [ ] Implement `container list`, `container stop`, etc.
- [ ] Add configuration management commands
- [ ] Create comprehensive help and documentation
- [ ] Add error handling and user feedback

### Phase 4: Testing and Polish (Week 4)
- [ ] Comprehensive unit tests
- [ ] Integration tests for both runtimes
- [ ] Configuration validation tests
- [ ] Documentation and examples

## Acceptance Criteria

### Must Have
- [ ] `portunix container run-in-container go` works with runtime auto-selection
- [ ] Configuration file supports runtime preference
- [ ] Default to Podman when no configuration present
- [ ] All existing Docker/Podman parameters supported
- [ ] Backward compatibility maintained

### Should Have
- [ ] `portunix config` commands for configuration management
- [ ] Auto-detection of available container runtimes
- [ ] Clear error messages when runtime unavailable
- [ ] Runtime status information (`portunix container info`)

### Nice to Have
- [ ] Runtime performance comparison
- [ ] Migration helper for existing workflows
- [ ] IDE integration improvements
- [ ] Shell completion for universal commands

## Configuration Examples

### Default Configuration (Podman preference)
```yaml
# ~/.portunix/config.yaml
container_runtime: podman
verbose: false
auto_update: true
```

### Docker Preference
```yaml
# ~/.portunix/config.yaml
container_runtime: docker
verbose: false
auto_update: true
```

### No Configuration (uses Podman default)
```yaml
# ~/.portunix/config.yaml (container_runtime not specified)
verbose: false
auto_update: true
```

## Migration Path

### For Existing Users
1. **No immediate changes required** - existing commands work
2. **Gradual adoption** - users can migrate to universal commands
3. **Configuration optional** - defaults work without configuration
4. **Clear documentation** - migration guide and examples

### For New Users
1. **Start with universal commands** - `portunix container`
2. **Automatic runtime selection** - no manual choice needed
3. **Simple configuration** - single YAML file
4. **Consistent experience** - same commands across platforms

## Success Metrics

- [ ] Universal container commands handle 100% of Docker/Podman use cases
- [ ] Configuration system supports all runtime selection scenarios
- [ ] Zero breaking changes to existing workflows
- [ ] Improved developer experience with unified interface
- [ ] Comprehensive test coverage (>90%)

## Dependencies

- Issue #028: Universal Container Parameters Support (completed)
- YAML configuration library integration
- Container runtime detection utilities
- Enhanced error handling system

## Related Issues

- #002: Docker Management Command
- #003: Podman Management Command  
- #028: Universal Container Parameters Support
- Future: Container orchestration integration

---

**Priority**: High - This creates a unified interface that significantly improves developer experience and reduces cognitive load when working with containers across different environments.