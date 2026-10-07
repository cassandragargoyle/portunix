# Issue #159: Service Orchestration Command Group

**Status**: ✅ Implemented
**Priority**: High
**Type**: Feature
**Labels**: feature, grpc, service-management, process-orchestration

## Context

Portunix needs a new top-level `portunix service` command group to manage gRPC plugin
processes. Clients (VSCode extensions, CLI scripts) need to start, discover, and stop
plugin services without knowing internal details about how plugins are launched.

This is part of the cross-project Service Orchestration initiative.
See [portunix-architecture specification](../../../../portunix-architecture/docs/architecture/specifications/service-orchestration/README.md).

## Requirements

### 1. Command Group `portunix service`

Implement the following subcommands:

- `portunix service start <plugin> [--mode shared|exclusive|prefer-exclusive] [--output text|json]`
- `portunix service list [--output text|json]`
- `portunix service info <plugin> --port <port> [--output text|json]`
- `portunix service release --session <id>`
- `portunix service stop <plugin> --force`
- `portunix service stop --all --force`

### 2. State File Management

- State file: `~/.portunix/processes.json`
- Lock file: `~/.portunix/processes.lock` (flock-based)
- Session-based client tracking (unique session ID per `start` call)
- Stale entry cleanup via gRPC health check on every read

### 3. Port Allocator

- Configurable port range (default: 50100-50199)
- Scan state file for used ports, pick first available
- Pass port to plugin as `--grpc-port <port>`

### 4. Process Spawning

- Spawn plugin as detached process
- Wait for `grpc.health.v1.Health/Check` (SERVING) with configurable timeout (default: 10s)
- Query `grpc.reflection.v1.ServerReflection/ListServices` to discover services
- Store discovered services in state file

### 5. Allocation Modes

- `shared` (default): join existing shared instance or start new
- `exclusive`: always start dedicated instance, single session only
- `prefer-exclusive`: Portunix decides based on resource availability

### 6. Plugin List Enhancement

- Add `interface` column to `portunix plugin list` output
- Read from plugin manifest `interfaces` field (e.g. `["cli", "grpc"]`)
- Only plugins with `grpc` interface are eligible for `portunix service` commands

### 7. Integration with Plugin Lifecycle

- `portunix plugin disable <plugin>` must stop all running service instances
- `portunix plugin uninstall <plugin>` must stop all running service instances

## Acceptance Criteria

- [ ] `portunix service start reco` spawns ptx-reco, returns session + endpoint as JSON
- [ ] `portunix service start reco` (second call, shared) returns same endpoint, new session
- [ ] `portunix service start reco --mode exclusive` spawns new dedicated instance
- [ ] `portunix service list` shows running instances with health check validation
- [ ] `portunix service info reco --port 50101` shows gRPC services and methods
- [ ] `portunix service release --session <id>` removes session, stops if last
- [ ] `portunix service stop reco --force` kills regardless of active sessions
- [ ] Concurrent `service start` calls are safe (flock prevents race conditions)
- [ ] `portunix plugin list` shows INTERFACE column
- [ ] `portunix plugin disable reco` stops running reco services

## Technical Notes

- Cobra command registration in `src/cmd/`
- State file operations should be extracted into a reusable `servicestate` package
- Port allocator should verify port is actually free (not just absent from state file)
- Session IDs: short random alphanumeric strings (e.g. 6 chars)

## Related Issues

- portunix-reco INT-011 (gRPC service mode)
- portunix-vscode #019 (service integration)
- portunix-plugins #058 (interface field in manifests)
- portunix-architecture #003 (cross-project tracking)

---

**Created**: 2026-03-21
**Author**: Architect
