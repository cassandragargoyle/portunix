# Issue #218: Universal Container Management Commands

> **Renumbered:** formerly internal issue #032. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

**Issue ID**: #032  
**Type**: Enhancement  
**Priority**: High  
**Status**: ✅ Implemented  
**Created**: 2025-01-09  
**Closed**: 2026-05-16  
**Labels**: container, docker, podman, universal-interface, cli, management

## Problem Statement

Team PXC reports that universal container management commands are missing from the `portunix container` interface. Currently, only `exec` and `run-in-container` commands exist under the universal container interface, but essential management commands like `stop`, `remove`, `start`, `logs`, etc. are only available under specific runtime commands (`portunix docker stop` or `portunix podman stop`).

Users need a complete set of universal container management commands that automatically detect the configured container runtime and execute the appropriate backend commands.

## Background

Currently available commands:
- ✅ `portunix container exec` - Universal exec command (Issue #031)
- ✅ `portunix container run-in-container` - Universal container creation
- ✅ `portunix container info` - Runtime information

**Missing universal commands:**
- ❌ `portunix container stop <container-name>` 
- ❌ `portunix container start <container-name>`
- ❌ `portunix container remove <container-name>`
- ❌ `portunix container logs <container-name>`
- ❌ `portunix container list` or `portunix container ls`
- ❌ `portunix container inspect <container-name>`

Users currently must use runtime-specific commands:
```bash
# Current (runtime-specific)
portunix docker stop test-container
portunix podman remove test-container

# Desired (universal)
portunix container stop test-container
portunix container remove test-container
```

## Requirements

### Functional Requirements
1. **Universal Stop**: `portunix container stop <container-name>` 
2. **Universal Start**: `portunix container start <container-name>`
3. **Universal Remove**: `portunix container remove <container-name> [--force]`
4. **Universal Logs**: `portunix container logs <container-name> [--follow]`
5. **Universal List**: `portunix container list` or `portunix container ls`
6. **Universal Inspect**: `portunix container inspect <container-name>`
7. **Runtime Auto-Detection**: Automatically determine active container runtime
8. **Consistent Output**: Unified output format regardless of backend
9. **Error Handling**: Consistent error messages and exit codes

### Technical Requirements
1. **Runtime Detection Logic**: Detect Docker vs Podman availability and configuration
2. **Command Delegation**: Route commands to appropriate backend implementations
3. **Argument Preservation**: Maintain all flags and options across runtime backends
4. **Container Discovery**: Find containers across different runtimes
5. **Force Operations**: Support `--force` flags where applicable

## Proposed Solution

### Command Structure
```bash
# Container lifecycle management
portunix container stop <container-name>
portunix container start <container-name>  
portunix container remove <container-name> [--force]
portunix container restart <container-name>

# Container inspection and monitoring
portunix container list [--all]
portunix container ls [--all]  # Alias for list
portunix container logs <container-name> [--follow] [--tail N]
portunix container inspect <container-name>

# Container status
portunix container ps      # Alias for list running containers
portunix container status <container-name>
```

### Implementation Plan

#### Phase 1: Core Management Commands
1. **Add stop subcommand** to `portunix container` command group
2. **Add start subcommand** for starting stopped containers
3. **Add remove subcommand** with --force flag support
4. **Implement runtime detection** using existing container configuration
5. **Route to appropriate backends** (docker.go or podman.go)

#### Phase 2: Inspection Commands  
1. **Add list/ls subcommand** for listing containers
2. **Add logs subcommand** with --follow support
3. **Add inspect subcommand** for detailed container information
4. **Add ps alias** for listing running containers

#### Phase 3: Advanced Features
1. **Add restart subcommand**
2. **Cross-runtime container discovery**
3. **Enhanced filtering options**
4. **Batch operations support**

### Technical Design

#### Runtime Detection and Container Discovery
```go
func findContainerRuntime(containerName string) (string, error) {
    // 1. Check user configuration preference
    preferredRuntime, _ := container.GetSelectedRuntime()
    
    // 2. Check if container exists in preferred runtime
    if containerExists(preferredRuntime, containerName) {
        return preferredRuntime, nil
    }
    
    // 3. Search across all available runtimes
    runtimes := []string{"docker", "podman"}
    for _, runtime := range runtimes {
        if isRuntimeAvailable(runtime) && containerExists(runtime, containerName) {
            return runtime, nil
        }
    }
    
    return "", fmt.Errorf("container '%s' not found in any runtime", containerName)
}

func containerExists(runtime, containerName string) bool {
    switch runtime {
    case "docker":
        return docker.ContainerExists(containerName)
    case "podman":
        return podman.ContainerExists(containerName)
    }
    return false
}
```

#### Universal Command Router
```go
func executeContainerCommand(command, containerName string, args []string) error {
    runtime, err := findContainerRuntime(containerName)
    if err != nil {
        return err
    }
    
    switch command {
    case "stop":
        return stopContainer(runtime, containerName)
    case "start":  
        return startContainer(runtime, containerName)
    case "remove":
        force := contains(args, "--force") || contains(args, "-f")
        return removeContainer(runtime, containerName, force)
    case "logs":
        follow := contains(args, "--follow") || contains(args, "-f")
        return showLogs(runtime, containerName, follow)
    }
    
    return fmt.Errorf("unsupported command: %s", command)
}
```

## File Changes Required

### New Files
- `cmd/container_stop.go` - Universal stop command
- `cmd/container_start.go` - Universal start command  
- `cmd/container_remove.go` - Universal remove command
- `cmd/container_logs.go` - Universal logs command
- `cmd/container_list.go` - Universal list command
- `cmd/container_inspect.go` - Universal inspect command

### Modified Files
- `cmd/container.go` - Add new subcommand registrations
- `app/docker/docker.go` - Expose container existence check functions
- `app/podman/podman.go` - Expose container existence check functions

### Integration Points
- Leverage existing container detection and management functions
- Use existing runtime selection logic from `app/container/`
- Maintain consistency with Issue #031 (universal exec)
- Follow patterns from Issue #029 (universal container creation)

## Acceptance Criteria

### Phase 1 - Core Management
- [ ] `portunix container stop <name>` works with Docker containers
- [ ] `portunix container stop <name>` works with Podman containers  
- [ ] `portunix container start <name>` works with both runtimes
- [ ] `portunix container remove <name>` works with both runtimes
- [ ] `portunix container remove <name> --force` works with both runtimes
- [ ] Runtime is automatically detected based on container location
- [ ] Error messages are consistent and helpful
- [ ] Commands work regardless of configured default runtime

### Phase 2 - Inspection Commands
- [ ] `portunix container list` shows containers from active runtime
- [ ] `portunix container logs <name>` works with both runtimes
- [ ] `portunix container logs <name> --follow` works with both runtimes
- [ ] `portunix container inspect <name>` works with both runtimes
- [ ] Commands find containers across runtimes when needed

### Phase 3 - Advanced Features
- [ ] `portunix container restart <name>` works with both runtimes
- [ ] Cross-runtime container discovery works correctly
- [ ] Batch operations are supported

## Testing Strategy

### Manual Testing
1. Test stop command with Docker containers: `portunix container stop docker-test`
2. Test stop command with Podman containers: `portunix container stop podman-test`
3. Test remove command: `portunix container remove test-container --force`
4. Test logs command: `portunix container logs test-container --follow`
5. Test list command: `portunix container list --all`
6. Test cross-runtime discovery: create container with Docker, try to stop with universal command
7. Test error conditions: non-existent containers, runtime not available

### Automated Testing
1. Unit tests for runtime detection logic
2. Integration tests for command delegation
3. Test container existence checking
4. Test error handling scenarios
5. Test flag preservation across backends

## Documentation Updates

1. Update `docs/FEATURES_OVERVIEW.md` with universal management commands
2. Add usage examples to container documentation
3. Update CLI help text for container commands
4. Create migration guide from runtime-specific to universal commands

## Dependencies

- No new dependencies required
- Uses existing Docker and Podman integration
- Leverages existing container management infrastructure
- Builds on Issue #031 (universal exec) patterns

## Related Issues

- #031: Universal Container Exec Command (foundation pattern)
- #029: Universal Container Command Implementation (related functionality)
- #028: Universal Container Parameters Support (related functionality)
- #002: Docker Management Command (backend implementation)
- #003: Podman Management Command (backend implementation)

## Notes

- Commands should maintain backward compatibility with existing docker/podman commands
- Implementation should be consistent with other universal container commands
- Consider performance implications of cross-runtime container discovery
- Ensure proper error handling when containers exist in multiple runtimes
- Consider adding aliases for common commands (ps → list, rm → remove)

## Migration Path

Users can gradually migrate from runtime-specific to universal commands:

**Current workflow:**
```bash
portunix docker ps
portunix docker stop my-container
portunix docker rm my-container
```

**New universal workflow:**
```bash
portunix container list
portunix container stop my-container  
portunix container remove my-container
```

Both workflows will continue to work, but universal commands provide better user experience and runtime abstraction.

---

**Estimated Implementation Time**: 6-8 hours  
**Complexity**: Medium-High  
**Team Impact**: High (directly addresses PXC team's workflow needs)