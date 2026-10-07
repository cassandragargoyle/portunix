# Issue #217: Universal Container Exec Command

> **Renumbered:** formerly internal issue #031. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

**Issue ID**: #031  
**Type**: Enhancement  
**Priority**: High  
**Status**: ✅ Implemented  
**Created**: 2025-01-09  
**Labels**: container, docker, podman, universal-interface, cli, execution

## Problem Statement

Team PXC reports that the command `portunix container exec test-container "ls -la /app/"` does not work, even though similar commands work for specific container runtimes like `portunix docker exec test-container "ls -la /app/"`.

Users need a unified container exec interface that automatically detects the configured container runtime and executes commands without requiring users to know whether Docker or Podman is being used.

## Background

Currently, Portunix provides runtime-specific exec commands:
- `portunix docker exec container-name "command"`  
- `portunix podman exec container-name "command"`

However, there is no universal `portunix container exec` command that:
1. Auto-detects the configured container runtime
2. Routes the command to the appropriate backend (Docker/Podman)
3. Provides consistent behavior regardless of the underlying runtime

## Requirements

### Functional Requirements
1. **Universal Exec Interface**: `portunix container exec <container-name> "<command>"`
2. **Runtime Auto-Detection**: Automatically determine active container runtime
3. **Command Translation**: Transform or pass-through commands to appropriate backend
4. **Consistent Output**: Unified output format regardless of backend
5. **Error Handling**: Consistent error messages and exit codes
6. **Shell Support**: Support for complex shell commands and arguments

### Technical Requirements
1. **Runtime Detection Logic**: Detect Docker vs Podman availability and configuration
2. **Command Parsing**: Parse and validate container names and commands
3. **Backend Integration**: Leverage existing `docker exec` and `podman exec` implementations
4. **Argument Handling**: Proper handling of quoted arguments and special characters
5. **Interactive Mode**: Support for interactive commands (TTY allocation)

## Proposed Solution

### Command Structure
```bash
# Basic usage
portunix container exec <container-name> "<command>"

# Examples
portunix container exec test-container "ls -la /app/"
portunix container exec web-server "cat /etc/nginx/nginx.conf"
portunix container exec db-container "mysql -u root -p"

# Interactive mode
portunix container exec -it test-container bash
portunix container exec --interactive test-container sh
```

### Implementation Plan

#### Phase 1: Core Implementation
1. **Add exec subcommand** to `portunix container` command group
2. **Implement runtime detection** using existing container configuration
3. **Route to appropriate backend** (docker.go or podman.go)
4. **Basic command execution** with argument preservation

#### Phase 2: Enhanced Features  
1. **Interactive mode support** (-it flags)
2. **Working directory support** (-w flag)
3. **Environment variable support** (-e flag)
4. **User specification** (-u flag)
5. **Consistent error handling**

#### Phase 3: Advanced Features
1. **Command validation** and security checks
2. **Audit logging** for executed commands
3. **Multi-container operations**
4. **Command history tracking**

### Technical Design

#### Runtime Detection Logic
```go
func detectContainerRuntime() (string, error) {
    // 1. Check user configuration preference
    if runtime := config.Get("container.default_runtime"); runtime != "" {
        return runtime, nil
    }
    
    // 2. Check which runtime has active containers
    if hasActiveContainers("docker") {
        return "docker", nil
    }
    if hasActiveContainers("podman") {
        return "podman", nil
    }
    
    // 3. Check which runtime is available
    if isRuntimeAvailable("docker") {
        return "docker", nil
    }
    if isRuntimeAvailable("podman") {
        return "podman", nil
    }
    
    return "", errors.New("no container runtime available")
}
```

#### Command Router
```go
func execInContainer(runtime, containerName, command string, flags ExecFlags) error {
    switch runtime {
    case "docker":
        return docker.ExecCommand(containerName, command, flags)
    case "podman":
        return podman.ExecCommand(containerName, command, flags)
    default:
        return fmt.Errorf("unsupported runtime: %s", runtime)
    }
}
```

## File Changes Required

### New Files
- `cmd/container_exec.go` - Universal exec command implementation

### Modified Files
- `cmd/container.go` - Add exec subcommand registration
- `app/docker/docker.go` - Expose exec functionality for universal interface
- `app/podman/podman.go` - Expose exec functionality for universal interface

### Integration Points
- Leverage existing container detection in `app/docker/docker.go:1460` and `app/podman/podman.go:1703`
- Use existing exec implementations in docker.go and podman.go
- Maintain consistency with universal container creation (#029)

## Acceptance Criteria

- [ ] `portunix container exec <name> "<command>"` works with Docker containers
- [ ] `portunix container exec <name> "<command>"` works with Podman containers  
- [ ] Runtime is automatically detected based on configuration and availability
- [ ] Commands are properly parsed and arguments preserved
- [ ] Interactive mode works with `-it` flags
- [ ] Error messages are consistent and helpful
- [ ] Works with complex shell commands and quoted arguments
- [ ] Exit codes match the executed command's exit codes

## Testing Strategy

### Manual Testing
1. Test with Docker containers: `portunix container exec docker-test "ls -la"`
2. Test with Podman containers: `portunix container exec podman-test "ps aux"`
3. Test interactive mode: `portunix container exec -it test-container bash`
4. Test complex commands: `portunix container exec test "find /app -name '*.log' | head -5"`
5. Test error conditions: non-existent containers, invalid commands

### Automated Testing
1. Unit tests for runtime detection logic
2. Integration tests for command execution
3. Test argument parsing and preservation
4. Test error handling scenarios

## Documentation Updates

1. Update `docs/FEATURES_OVERVIEW.md` with universal exec command
2. Add usage examples to container documentation
3. Update CLI help text for container commands

## Dependencies

- No new dependencies required
- Uses existing Docker and Podman integration
- Leverages existing container management infrastructure

## Related Issues

- #029: Universal Container Command Implementation (foundation)
- #028: Universal Container Parameters Support (related functionality)
- #002: Docker Management Command (backend implementation)
- #003: Podman Management Command (backend implementation)

## Notes

- Command should maintain backward compatibility with existing docker/podman exec commands
- Implementation should be consistent with other universal container commands
- Consider security implications of command execution
- Ensure proper handling of TTY allocation for interactive commands

---

---

## ✅ IMPLEMENTATION COMPLETED

**Implemented:** 2025-01-09  
**Version:** v1.5.8+  

### Implementation Summary

Universal Container Exec Command has been successfully implemented:

#### ✅ Core Features Implemented
- **`portunix container exec <container-name> <command>`** - Universal exec command
- **Runtime auto-detection** - Automatically selects Docker or Podman based on configuration
- **Clean output** - Only shows command output without debug information
- **Interactive mode support** - `-i/--interactive` flag for TTY allocation
- **Proper error handling** - Consistent error messages and exit codes

#### ✅ Technical Implementation
- **Backend integration** - Leverages existing `docker.ExecCommand` and `podman.ExecCommand`
- **Enhanced functions** - Added `ExecCommandWithOptions` for both Docker and Podman
- **Argument parsing** - Proper handling of container names and command arguments
- **Flag support** - Interactive mode with `-i/--interactive` flag

### Usage Examples

```bash
# Basic command execution
portunix container exec test-container -- ls -la /app/

# Interactive mode
portunix container exec -i test-container -- bash

# Complex commands
portunix container exec web-server -- cat /etc/nginx/nginx.conf

# Check help
portunix container exec --help
```

### Technical Details

**New Functions Added:**
- `docker.ExecCommandWithOptions(containerID string, command []string, interactive bool)`
- `podman.ExecCommandWithOptions(containerID string, command []string, interactive bool)`

**Files Modified:**
- `cmd/container.go` - Added containerExecCmd with runtime detection
- `app/docker/docker.go` - Enhanced exec functionality
- `app/podman/podman.go` - Enhanced exec functionality
- `cmd/container_exec_test.go` - Comprehensive test suite

### Resolution Status

**FULLY IMPLEMENTED** ✅ - Team PXC can now use `portunix container exec` which automatically detects the container runtime and executes commands without requiring knowledge of the underlying Docker/Podman implementation.

---

**Estimated Implementation Time**: 4-6 hours  
**Actual Implementation Time**: 2 hours  
**Complexity**: Medium  
**Team Impact**: High (directly addresses PXC team's reported issue)