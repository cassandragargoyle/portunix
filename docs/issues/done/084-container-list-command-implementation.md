# Issue #84: Container List Command Implementation

**Type**: Internal
**Status**: ✅ Implemented
**Created**: 2025-09-28
**Priority**: High
**Category**: Container Management

## Description
The `portunix container list` command is not yet implemented. When users run this command, they receive an error message indicating that the functionality is missing. This is a core feature for container management that needs to be implemented.

## Current Behavior
```bash
$ portunix container list
Container list not yet implemented in helper
```

## Expected Behavior
The command should list all containers (both running and stopped) with relevant information such as:
- Container ID
- Container Name
- Status (Running/Stopped)
- Image
- Created time
- Ports (if applicable)

## Technical Requirements

### Command Structure
- Main command: `portunix container list`
- Alternative: `portunix container ls`
- Flags to implement:
  - `--all` or `-a`: Show all containers (default behavior)
  - `--running`: Show only running containers
  - `--format`: Custom output format (table/json/yaml)

### Implementation Details
1. **Detection**: Automatically detect whether Docker or Podman is available
2. **Unified Interface**: Provide consistent output regardless of underlying container runtime
3. **Error Handling**: Gracefully handle cases where neither Docker nor Podman is available

### Output Format (Default - Table)
```
CONTAINER ID    NAME                STATUS      IMAGE              CREATED
abc123def456    my-ubuntu-dev       Running     ubuntu:22.04       2 hours ago
def456ghi789    test-container      Stopped     alpine:latest      1 day ago
```

## Affected Files
- `app/docker/container_helper.go` - Main implementation location
- `app/docker/docker.go` - Docker-specific implementation
- `app/podman/podman.go` - Podman-specific implementation
- `cmd/container.go` - Command registration and CLI handling

## Testing Requirements
- Unit tests for parsing container list output
- Integration tests with actual Docker/Podman
- Test both Docker and Podman implementations
- Test error cases (no runtime available)

## Acceptance Criteria
- [ ] Command `portunix container list` works
- [ ] Command `portunix container ls` works (alias)
- [ ] Lists all containers by default
- [ ] Shows relevant container information
- [ ] Works with both Docker and Podman
- [ ] Handles missing container runtime gracefully
- [ ] Output is properly formatted and readable
- [ ] Tests pass on both Windows and Linux

## Notes
- This is a fundamental feature that many other container commands depend on
- Should maintain consistency with existing container command structure
- Consider future extensibility for additional filtering options