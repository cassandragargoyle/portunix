# ADR-027: Compose Command Architecture

| Field | Value |
|-------|-------|
| **Status** | Proposed |
| **Date** | 2025-12-01 |
| **Author** | Architect |
| **Issue** | #102 |

## Context

External team (TovekToolsAgent) requires Portunix to provide a unified interface for docker-compose/podman-compose operations. Currently, users must manually determine which compose tool is available and invoke it directly.

Portunix already provides universal container commands (`portunix container run`, `portunix container exec`, etc.) that abstract Docker/Podman differences. A compose command would extend this abstraction to multi-container orchestration.

### Current State

```
portunix container
├── run
├── run-in-container
├── exec
├── cp
├── stop/start/rm
├── logs
└── list
```

### Required Interface

```bash
portunix compose -f docker-compose.yml up [service]
portunix compose -f docker-compose.yml down
portunix compose -f docker-compose.yml build [service]
portunix compose -f docker-compose.yml logs [service]
portunix compose -f docker-compose.yml ps
```

## Decision

### 1. Command Placement: Top-Level with Alias

Implement `portunix compose` as a top-level command with `portunix container compose` as an alias.

**Rationale**:
- Matches user expectation from `docker-compose` CLI pattern
- Shorter, more ergonomic command for frequent use
- Alias maintains consistency with container command hierarchy
- Top-level placement reflects that compose is a distinct tool, not a container lifecycle operation

### 2. Compose Runtime Detection: Independent Priority-Based

Implement separate compose runtime detection, independent of container runtime selection.

**Detection Order** (based on configured container runtime preference):

```
If container_runtime == "docker" or "auto":
    1. docker compose (V2, embedded in Docker CLI)
    2. docker-compose (V1, standalone binary)
    3. podman-compose (fallback)

If container_runtime == "podman":
    1. podman-compose
    2. docker compose (V2)
    3. docker-compose (V1)
```

**Rationale**:
- Docker Compose V2 is the modern, preferred tool
- V1 still widely used, provides backward compatibility
- Podman-compose supported for rootless environments
- Aligned with but not strictly bound to container runtime selection

### 3. Argument Handling: Full Passthrough

Use `DisableFlagParsing: true` and pass all arguments directly to detected compose tool.

**Rationale**:
- Maximum compatibility with all compose features
- Avoids maintenance burden of parsing ever-changing compose flags
- Users get full compose functionality without Portunix limitations
- Consistent with existing `container exec` and `container run-in-container` patterns

### 4. File Structure

```
src/
├── cmd/
│   └── compose.go          # Command definition, argument passthrough
└── app/
    └── compose/
        ├── runtime.go      # Compose tool detection
        └── executor.go     # Compose command execution
```

## Architecture

### Compose Runtime Detection

```
┌─────────────────────────────────────────────────────────────┐
│                    ComposeRuntime                           │
├─────────────────────────────────────────────────────────────┤
│ GetComposeRuntime() → (ComposeRuntime, error)               │
│ IsDockerComposeV2Available() → bool                         │
│ IsDockerComposeV1Available() → bool                         │
│ IsPodmanComposeAvailable() → bool                           │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    ComposeExecutor                          │
├─────────────────────────────────────────────────────────────┤
│ Execute(runtime, args) → error                              │
│ - Constructs proper command based on runtime                │
│ - Passes all args directly to compose tool                  │
│ - Streams stdout/stderr to terminal                         │
└─────────────────────────────────────────────────────────────┘
```

### Command Flow

```
portunix compose -f docker-compose.yml up web
         │
         ▼
┌─────────────────────┐
│  GetComposeRuntime  │
│  ─────────────────  │
│  1. Check V2        │
│  2. Check V1        │
│  3. Check Podman    │
└─────────────────────┘
         │
         ▼ (e.g., docker compose)
┌─────────────────────┐
│  ComposeExecutor    │
│  ─────────────────  │
│  docker compose     │
│    -f docker-       │
│    compose.yml      │
│    up web           │
└─────────────────────┘
         │
         ▼
     stdout/stderr
```

### Runtime Detection Commands

| Runtime | Detection Command | Success Indicator |
|---------|------------------|-------------------|
| Docker Compose V2 | `docker compose version` | Exit code 0 |
| Docker Compose V1 | `docker-compose version` | Exit code 0 |
| Podman Compose | `podman-compose version` | Exit code 0 |

### Error Messages

```
No compose tool available.

Install one of the following:

Docker Compose V2 (recommended):
  - Included with Docker Desktop
  - Linux: apt install docker-compose-plugin

Docker Compose V1:
  - portunix install docker-compose

Podman Compose:
  - portunix install podman-compose
```

## Consequences

### Positive

- **Unified Interface**: Single command works across Docker/Podman environments
- **Transparent Fallback**: Automatic tool selection without user configuration
- **Full Compatibility**: All compose features available via argument passthrough
- **Consistent UX**: Matches existing Portunix container command patterns
- **External Integration**: Enables tools like TovekToolsAgent to use Portunix abstraction

### Negative

- **Limited Enhancement Opportunity**: Passthrough mode prevents Portunix-specific features
- **Error Message Quality**: Compose tool errors shown verbatim, no Portunix context
- **Version Compatibility**: Different compose versions may have incompatible features

### Mitigations

- Document compose version requirements for specific features
- Consider future enhancement for file auto-detection (`-f` optional)
- Add `--verbose` flag to show which runtime was selected

## Alternatives Considered

### A: Subcommand Only (`portunix container compose`)

**Rejected**: Less ergonomic, user expectation is `compose` as primary command

### B: Aligned Runtime Selection

Force compose to use same backend as `container` commands.

**Rejected**: Compose tools are installed independently, would create false errors

### C: Parsed Arguments with Enhancements

Parse compose arguments, add Portunix features like auto-file-detection.

**Rejected**: High maintenance burden, risk of breaking edge cases, delays initial delivery

## Implementation Notes

- Use `cobra.Command` with `DisableFlagParsing: true`
- Execute compose command with `exec.Command().Run()` for terminal passthrough
- Consider adding installation definitions to `assets/install-packages.json`:
  - `docker-compose` (V1)
  - `podman-compose`

## References

- Issue #102: Compose Command Implementation
- Docker Compose V2: https://docs.docker.com/compose/
- Podman Compose: https://github.com/containers/podman-compose
- Existing pattern: `src/cmd/container.go`
