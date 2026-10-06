# Issue #185: Plugin Security Hardening

**Status:** 📋 Open
**Priority:** Medium
**Type:** Enhancement / Security
**Labels:** enhancement, plugin-system, security, grpc, tls, signing, audit
**Affects:** `src/app/plugins/`, `src/cmd/plugin.go`, `src/helpers/ptx-plugin-registry/`
**Parent:** [#007 — Plugin System with gRPC Architecture](done/007-plugin-system-grpc.md)
**Created:** 2026-05-08

## Summary

Plugin system core (issue #007) is fully implemented and shipping 19 plugins.
Several security items from the original #007 specification were deferred and
need to be addressed: encrypted plugin communication, plugin signature
verification, resource limits, audit logging, and a local-development
`plugin register` command.

## Motivation

- Harden the plugin runtime against compromised or malicious plugins
- Establish trust chain for the official `CassandraGargoyle/portunix-plugins`
  repository through digital signatures
- Provide observability of plugin behaviour (audit log)
- Prevent runaway plugins from exhausting host resources
- Smooth out plugin developer workflow with first-class `register` for local
  unpacked plugins (current workaround is `install` from a path)

## Scope

### 1. TLS for core ↔ plugin gRPC channel

- Generate per-plugin self-signed certificate at install time, pinned by the
  manager
- gRPC client (`src/app/plugins/grpc_client.go`) and registry server
  (`src/app/plugins/server/registry_server.go`) negotiate mTLS
- Configurable opt-out for local development (`--insecure-plugin-channel`)
- Document in `src/app/plugins/README.md`

### 2. Plugin signature verification

- Extend `plugin-index.json` entry with `signature` field (Ed25519 over the
  release tarball checksum)
- Embed CassandraGargoyle public key in core; verify on `plugin install`
- Reject install on signature mismatch unless `--allow-unsigned` is passed
- `verified: true` in registry must require a valid signature

### 3. Resource usage monitoring & limits

- Track CPU time, RSS, and FD count per running plugin process
- Surface metrics through `portunix plugin info <name>` and `plugin health`
- Configurable per-plugin limits in `plugin.yaml`
  (`limits: { cpu_percent, memory_mb, max_open_files }`)
- On limit breach: log + graceful shutdown attempt, then SIGKILL

### 4. Audit logging

- Append-only audit log under `~/.portunix/logs/plugin-audit.log`
- Records: install / uninstall / enable / disable / start / stop / RPC call
  summary (no payload bodies by default), with timestamp, plugin name,
  version, caller (CLI vs MCP vs API)
- `portunix plugin audit [--since duration] [--name <plugin>]` query command

### 5. `plugin register` command

- `portunix plugin register <path>` — register an unpacked plugin tree for
  local development without copying to the install dir
- Skips signature verification (development-only, must warn)
- Auto-detects `plugin.yaml` and `binary` path
- Survives across `portunix` invocations via registry entry flagged `dev: true`

## Acceptance Criteria

- [ ] mTLS handshake required between core and plugin gRPC server (verified
      via integration test in `test/integration/plugin_*_test.go`)
- [ ] `--insecure-plugin-channel` flag disables TLS for local development
- [ ] `plugin install` rejects tarball with mismatched signature; clear error
- [ ] `plugin install --allow-unsigned <path>` works for unsigned plugins
- [ ] Resource limits in `plugin.yaml` enforced; over-limit plugin is
      terminated and recorded in audit log
- [ ] `plugin info` shows live CPU / memory metrics
- [ ] Audit log records all lifecycle events; queryable via `plugin audit`
- [ ] `plugin register <path>` registers an unpacked plugin and persists
      across runs
- [ ] Documentation updated: `src/app/plugins/README.md`,
      `docs/plugin-development/`
- [ ] Integration tests pass on Linux and Windows
- [ ] Acceptance protocol `docs/testing/internal/acceptance-185.md` PASS

## Out of Scope

- Sandboxing plugins inside containers (separate effort, future issue)
- Cross-machine plugin distribution beyond GitHub releases
- Plugin marketplace UI

## Technical Notes

- Signing key management: keep the private key in the
  `CassandraGargoyle/portunix-plugins` release pipeline only; rotate via
  release of new core that bumps the embedded public key
- TLS cert lifecycle: regenerated on each `plugin install` /
  `plugin reinstall`; not rotated automatically afterwards
- Audit log rotation: simple size-based rotation (10 MB × 5 files)
- Resource monitoring on Windows: use `GetProcessTimes` /
  `GetProcessMemoryInfo`; on Linux: `/proc/<pid>/stat`, `/proc/<pid>/status`

## Implementation Plan

1. mTLS scaffolding (cert generation, manager + grpc_client wiring)
2. Signature verification path in `plugin install`
3. Resource monitoring goroutine in plugin manager + manifest schema bump
4. Audit log writer + `plugin audit` query command
5. `plugin register` command + dev-flag plumbing in registry
6. Integration tests + cross-platform validation
7. Acceptance protocol

## Priority

**Medium** — core plugin system works without these; this hardens it for
production / multi-tenant use.
