# Issue #102: Compose Command Implementation

**Internal ID**:  #102
**Public ID**:  
**Title**: Compose Command Implementation
**Status**: 📋 Open
**Priority**: High
**Type**: Enhancement
**Labels**: enhancement, container, docker-compose, podman-compose, universal-interface
**Created**: 2025-12-01
**Requested by**: External team

## Summary

Implement a universal `compose` command (or `container compose`) that wraps docker-compose/podman-compose with automatic runtime detection, consistent with existing Portunix container architecture.

## Background

An external project requires Portunix to provide a unified compose interface for managing multi-container applications. This aligns with existing universal container command architecture that already abstracts Docker/Podman differences.

## Requirements

### Required Usage Patterns

```bash
# Start service from compose file
portunix compose -f docker-compose.docs.yml up docs-server

# Stop services
portunix compose -f docker-compose.docs.yml down

# Additional standard compose commands
portunix compose -f <file> up [service]
portunix compose -f <file> down
portunix compose -f <file> build [service]
portunix compose -f <file> logs [service]
portunix compose -f <file> ps
```

### Behavior Requirements

1. **Runtime Detection Priority**:
   - `docker compose` (Docker Compose V2 - preferred)
   - `docker-compose` (Docker Compose V1 - fallback)
   - `podman-compose` (Podman alternative)

2. **Argument Passthrough**: All arguments passed directly to detected compose tool

3. **Error Handling**: Clear error message with installation instructions when no compose tool available

### Integration Example (external project)

```python
# In generate_docs.py
portunix_cmd = [
    "portunix", "compose",
    "-f", str(compose_file),
    "up", service,
]
subprocess.run(portunix_cmd, cwd=str(PROJECT_ROOT))
```

## Technical Analysis

### Existing Architecture Patterns

Current container command structure (`src/cmd/container.go`):

```
containerCmd (root)
├── run
├── run-in-container
├── exec
├── cp
├── info
├── check
├── stop
├── start
├── rm/remove
├── logs
└── list/ls/ps
```

Runtime detection in `src/app/container/runtime.go`:
- `GetSelectedRuntime()` - returns configured runtime
- `IsDockerAvailable()` - checks Docker availability
- `IsPodmanAvailable()` - checks Podman availability

### Compose Tool Detection Requirements

Unlike single-container commands, compose requires different detection logic:

| Tool | Command | Check Method |
|------|---------|--------------|
| Docker Compose V2 | `docker compose` | `docker compose version` |
| Docker Compose V1 | `docker-compose` | `docker-compose version` |
| Podman Compose | `podman-compose` | `podman-compose version` |

## Architecture Decision Points

### Decision 1: Command Placement

**Decision**: Subcommand `portunix container compose`
- Pro: Consistent with existing architecture
- Pro: Groups all container-related commands
- Aligns with existing container command structure

### Decision 2: Runtime Selection Strategy

**Option A**: Independent detection (compose-specific)
- Detect compose tools separately from container runtime
- Pro: Compose tools can be installed independently
- Con: Could select different backend than container commands

**Option B**: Aligned with container runtime
- If `container_runtime: docker`, prefer docker-compose
- Pro: Consistent behavior across all container commands
- Con: Less flexible

**Recommendation**: Option A with preference order matching configured runtime

### Decision 3: Argument Handling

**Option A**: Full passthrough (minimal processing)
- Pass all arguments directly to compose tool
- Pro: Full compatibility with all compose features
- Con: Less Portunix-specific enhancements

**Option B**: Parsed arguments with Portunix enhancements
- Parse common flags, add Portunix-specific features
- Pro: Can add features like auto-file-detection
- Con: Risk of breaking edge cases

**Recommendation**: Option A for initial implementation

## Implementation Components

### New Files Required

1. `src/app/compose/runtime.go` - Compose runtime detection
2. `src/app/compose/executor.go` - Compose command execution
3. Update `src/cmd/container.go` - Add compose subcommand

### Runtime Detection Logic

```
func GetComposeRuntime() (ComposeRuntime, error):
    1. Load configuration for preferred runtime
    2. Detection order based on config:
       - If docker preferred: docker compose → docker-compose → podman-compose
       - If podman preferred: podman-compose → docker compose → docker-compose
    3. Return first available runtime
    4. If none available, return error with installation instructions
```

### Error Messages

When no compose tool available:

```
No compose tool detected. Install one of the following:

For Docker:
  Docker Compose V2 (recommended): Included with Docker Desktop
  Docker Compose V1: portunix install docker-compose

For Podman:
  podman-compose: portunix install podman-compose
```

## Acceptance Criteria

1. [ ] `portunix container compose -f <file> up [service]` works with Docker Compose V2
2. [ ] `portunix container compose -f <file> up [service]` works with Docker Compose V1
3. [ ] `portunix container compose -f <file> up [service]` works with podman-compose
4. [ ] `portunix container compose -f <file> down` works
5. [ ] `portunix container compose -f <file> build [service]` works
6. [ ] `portunix container compose -f <file> logs [service]` works
7. [ ] `portunix container compose -f <file> ps` works
8. [ ] Error message displayed when no compose tool available
9. [ ] `portunix container compose --help` shows comprehensive help
10. [ ] `portunix container compose` works as primary command
11. [ ] All arguments passed through correctly to underlying tool

## Testing Requirements

### Unit Tests

- Compose runtime detection logic
- Argument parsing and passthrough
- Error handling for missing tools

### Integration Tests

Container-based testing required per TESTING_METHODOLOGY.md:

```bash
# Test in Docker environment
portunix docker run ubuntu
# Inside container: test compose commands

# Test in Podman environment
portunix podman run ubuntu
# Inside container: test compose commands
```

### Test Scenarios

1. Docker Compose V2 available only
2. Docker Compose V1 available only
3. Podman-compose available only
4. Multiple compose tools available (priority order)
5. No compose tool available (error handling)
6. Complex compose files with multiple services

## Dependencies

- Existing container runtime detection (`src/app/container/runtime.go`)
- Cobra command framework (already in use)
- No new external dependencies required

## Future Enhancements (Out of Scope)

- Auto-detection of compose files in current directory
- Portunix-specific compose file extensions
- Integration with Portunix package installation for services
- Compose file validation

## References

- Existing container command: `src/cmd/container.go`
- Runtime detection: `src/app/container/runtime.go`
- Docker Compose V2 docs: https://docs.docker.com/compose/
- Podman Compose docs: https://github.com/containers/podman-compose

## Notes for Implementation

- Follow existing patterns in `container.go` for command structure
- Use `DisableFlagParsing: true` for full argument passthrough
- Consider adding compose tool installation to `assets/install-packages.json`
