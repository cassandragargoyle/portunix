# Issue #175: Plugin Platform-Capability Query Interface

**Status:** ✅ Implemented
**Priority:** High
**Type:** Feature / Architecture
**Labels:** enhancement, plugin-system, architecture, grpc, cli, cross-team, api-contract, manifest
**Affects:** portunix plugin runtime, `CassandraGargoyle/api` contract schemas
**Related ask:** [Synapse Issue 012 — Plugin Platform-Capability Query Interface](../../../../portunix-architecture/docs/architecture/components/portunix-synapse/issues/012-portunix-plugin-platform-capability-query.md)
**Driver (cross-team):** Zdenek (Synapse architect)
**Created:** 2026-04-19
**Implemented:** 2026-04-19
**Acceptance:** [acceptance-175.md](../../testing/internal/acceptance-175.md) — PASS (36/36)

## Overview

Introduce a standardized mechanism in Portunix plugins through which a hosting platform
(Portunix Synapse, Pack, Agent, …) can discover which locally installed plugins are meant
to run under that platform and retrieve the platform-specific metadata needed to embed them.

This is a **cross-team ask** from the Synapse architecture team. Synapse is rolling out an
Extensions Platform and needs embedded extensions (connectors, assistants, audio/video
processors) to be delivered as Portunix plugins. Without this interface, Synapse is forced to
maintain a classpath fallback to discover its extensions.

The design is **platform-agnostic**: Synapse is the first consumer, but the same mechanism
will serve Pack, Agent and future hosting platforms.

## Problem Statement

### Current State

- `plugin-manifest.schema.json` (v1.0.0) describes a plugin's identity, runtime, commands,
  and AI integration, but has no way for a plugin to declare itself as an extension of a
  higher-level platform (e.g. Synapse).
- `ptx plugin list` and the `PluginService` gRPC API offer no way to filter plugins by
  target platform or to retrieve platform-specific ancillary metadata.
- Synapse currently needs to scan its own classpath to find extension jars — a temporary
  workaround that duplicates Portunix's packaging/installation/registry responsibilities.

### Goal

Provide a declarative, backward-compatible extension to the plugin manifest plus query APIs
on ptx so that a platform can answer:

> "Which of the installed Portunix plugins are meant to run under platform `<X>` at version
> `<Y>`, and what is the platform-specific payload needed to register them?"

## Requirements

The finalized interface MUST provide:

1. **Declarative platform-support list per plugin.** A plugin declares zero or more
   platforms it supports, each with:
   - `name` — canonical platform identifier (e.g. `synapse`, `pack`, `agent`).
   - `min_version` / `max_version` — SemVer range of the platform the plugin is compatible
     with.
   - `platform_payload` — opaque object, platform-specific, **not interpreted by Portunix**.
     Portunix validates shape (must be an object), not content. The owning platform
     validates content against its own schema (e.g. Synapse validates against
     `synapse-extension-manifest.schema.json`).
   - Optional `features[]` — short feature flags for capability negotiation without parsing
     `platform_payload`.

2. **Query API on the ptx plugin registry:**
   - CLI: `ptx plugin list --platform=<name> [--platform-version=<semver>] [--feature=<f>]`
   - Programmatic: new gRPC RPC on `PluginService` returning plugin manifest + matched
     `platform_payload`.
   - Response includes the matched `features[]` that the caller requested.

3. **Stable contract in the `api` repo.** The addition is expressed as a modification of
   `plugin-manifest.schema.json` (bumped to `$version: 1.1.0`) with backward compatibility:
   - Plugins without `supported_platforms` continue to work exactly as today.
   - The field is additive; existing tooling ignores unknown fields.

4. **Validation at `ptx plugin install` time.** Malformed `supported_platforms[]` entries
   fail install with a clear error — the platform never sees a broken declaration at
   runtime.

5. **Semantic-version matching.** The query MUST honour `min_version` / `max_version` vs.
   the caller's declared platform version so a plugin targeting Synapse `0.x` is not
   offered to Synapse `2.x`.

## Proposed Design (starting point)

The Synapse ask provides the initial design; final shape is owned by the Portunix team.
See the [source ask](../../../../portunix-architecture/docs/architecture/components/portunix-synapse/issues/012-portunix-plugin-platform-capability-query.md)
for the full proposal. Summary:

### Manifest extension

Add a top-level `supported_platforms[]` array to the plugin manifest:

```yaml
supported_platforms:
  - name: synapse
    min_version: "0.3.0"
    max_version: "0.9.x"
    features:
      - connector.command
    platform_payload:
      # Opaque to Portunix — owned by Synapse schema.
      surfaces:
        connector.command: true
      command:
        supportedCapabilities:
          - erp.customer.lookup
          - erp.invoice.create
      configSchema: config.schema.json
      secretsSchema: secrets.schema.json
```

### gRPC addition (on `PluginService` in `plugin.proto`)

```proto
rpc ListPluginsForPlatform(ListPluginsForPlatformRequest)
    returns (ListPluginsForPlatformResponse);

message ListPluginsForPlatformRequest {
  string platform         = 1;  // required, e.g. "synapse"
  string platform_version = 2;  // optional; semver string
  repeated string features = 3; // optional; AND-filter
}

message ListPluginsForPlatformResponse {
  repeated MatchedPlugin plugins = 1;
}

message MatchedPlugin {
  PluginInfo manifest = 1;              // existing GetInfo shape
  google.protobuf.Struct platform_payload = 2;
  repeated string matched_features = 3;
}
```

### CLI

```bash
ptx plugin list --platform=synapse --platform-version=0.3.2 --feature=connector.command
# -> JSON array of matching plugins (manifest + platform_payload)
```

### Runtime dynamic query (stretch / later)

A plugin-side gRPC `QueryPlatformSupport(...)` is listed as optional. **Out of scope for
the first delivery** — declarative manifest covers the common case.

## Implementation Decisions (2026-04-19)

During planning the proposed design was refined to match the current Portunix
architecture. The original ask places `ListPluginsForPlatform` on `PluginService`, but
that service is implemented **by plugins** (Portunix is the gRPC client). Plugins cannot
answer for the entire registry. The query must therefore be served by a new **ptx-side
gRPC service** backed by the existing registry.

### DEC-1: New gRPC service `PluginRegistryService`

A new proto file `src/app/plugins/proto/plugin_registry.proto` defines
`PluginRegistryService` with `ListPluginsForPlatform`. This is separate from
`PluginService` (plugin-side) so roles remain clear.

### DEC-2: Dedicated helper binary `ptx-plugin-registry`

Portunix does not have a long-running ptx-side gRPC daemon today. Precedent exists:
`ptx-mcp serve` already runs as TCP / Unix-socket / stdio daemon. Same pattern applied
here — new helper `ptx-plugin-registry` with subcommand `serve`:

```bash
ptx-plugin-registry serve --mode unix                       # Linux/macOS default
ptx-plugin-registry serve --mode tcp --port 9500 --bind 127.0.0.1   # Windows / explicit
```

Follows the full Helper Binary Development checklist
(`docs/contributing/HELPER-BINARY-DEVELOPMENT.md`).

### DEC-3: Transport

- **Unix domain socket** on Linux/macOS:
  `$XDG_RUNTIME_DIR/portunix/plugin-registry.sock` (fallback:
  `$HOME/.portunix/run/plugin-registry.sock`), permissions `0600`.
- **TCP loopback** on Windows (no unix sockets pre-Win10): `127.0.0.1:9500`.
- No TLS / tokens in v1 — filesystem permissions + loopback-only provide localhost
  isolation. Remote access is explicitly out of scope.

### DEC-4: Lifecycle — on-demand

The daemon runs **only while a consumer needs it**. Synapse starts the process as a
subprocess, calls `ListPluginsForPlatform`, then stops the daemon. No systemd, no
always-on service in v1. This keeps scope bounded; persistent-daemon + auth can come as a
follow-up issue when a concrete need arises.

### DEC-5: Core matching logic as a package function

`Registry.ListPluginsForPlatform(platform, version, features)` on
`src/app/plugins/manager/registry.go` is the **single source of truth**. Both the CLI
(`ptx plugin list --platform …`) and the gRPC server call the same function, so output
parity (acceptance criterion) is structural, not maintained by duplication.

### DEC-6: SemVer range syntax

Explicit `min_version` / `max_version` fields (inclusive). Open-ended ranges allowed by
omitting either bound. Rationale: matches the manifest example in this issue, consistent
with existing `portunix_min_version` field, no new parser required beyond
`golang.org/x/mod/semver` already used elsewhere in the codebase.

npm-style range strings (`>=0.3.0 <1.0.0`) can be added later as an additive feature if
demand arises.

### DEC-7: `platform_payload` storage — byte-identical round-trip

`platform_payload` stored as `json.RawMessage` end-to-end (manifest → registry → query
output). No unmarshalling on the Portunix side — guarantees byte-identical payload as
required by acceptance criterion.

### DEC-8: JSON Schema location

Additive fields inline in existing `plugin-manifest.schema.json`, schema version bumped
`1.0.0` → `1.1.0`. No sibling schema file — keeps the plugin manifest self-contained for
tooling that reads a single file.

## Implementation Plan

Proposed work streams (to be refined during planning):

### WS-1 — API contract (`CassandraGargoyle/api`)

- [ ] Design schema update for `supported_platforms[]` on `plugin-manifest.schema.json`.
- [ ] Bump schema version to `1.1.0` under the existing SemVer policy.
- [ ] Decide: inline inside `plugin-manifest.schema.json` vs. sibling
  `plugin-platform-support.schema.json` referenced from main manifest.
- [ ] Open schema PR referencing this issue and the Synapse ask as the driver.

### WS-2 — Go types and manifest parsing (`src/app/plugins/`)

- [ ] Add `SupportedPlatforms []SupportedPlatform` to `PluginManifest` in `types.go`.
- [ ] Define `SupportedPlatform` struct (name, min/max version, features,
  `platform_payload json.RawMessage`).
- [ ] Keep `platform_payload` as `json.RawMessage` — Portunix must **not** interpret
  content.
- [ ] Extend `ValidateManifest()` in `manifest.go`:
  - Shape validation (name non-empty, versions parse as SemVer, payload is an object).
  - Reject malformed entries at `ptx plugin install` time with actionable error messages.

### WS-3 — Registry storage (`src/app/plugins/manager/`)

- [ ] Persist parsed `supported_platforms` in the ptx plugin registry alongside the
  existing manifest data.
- [ ] Ensure storage round-trips `platform_payload` byte-identical (no re-serialization
  loss).
- [ ] Index by platform name for efficient lookup when query arrives (can start with
  linear scan; revisit if plugin count grows).

### WS-4 — gRPC API (new `src/app/plugins/proto/plugin_registry.proto`)

**Revised per DEC-1, DEC-2:** `PluginService` (plugin-side) is NOT modified. A new
`PluginRegistryService` is introduced in a dedicated proto file and served by a new
helper binary `ptx-plugin-registry`.

- [ ] Add new proto file `src/app/plugins/proto/plugin_registry.proto` with
  `PluginRegistryService.ListPluginsForPlatform` and its messages.
- [ ] Regenerate Go bindings.
- [ ] Implement server-side handler — delegates matching to
  `Registry.ListPluginsForPlatform(...)` (DEC-5).
- [ ] Create new helper binary `src/helpers/ptx-plugin-registry/` with subcommand `serve`
  following the Helper Binary Development checklist (11 phases).
- [ ] Unix socket transport on Linux/macOS, TCP loopback on Windows (DEC-3).
- [ ] On-demand lifecycle — no systemd / always-on daemon in v1 (DEC-4).

### WS-5 — CLI surface (`src/cmd/plugin.go`)

- [ ] Extend `ptx plugin list` with `--platform`, `--platform-version`, `--feature` flags.
- [ ] JSON output mode returns manifest + matched `platform_payload` per plugin.
- [ ] Human-readable output: compact (plugin name, matched features, payload omitted
  unless `--verbose`).

### WS-6 — Documentation + example

- [ ] Update Portunix plugin developer guide with a `supported_platforms` section.
- [ ] Ship a **Synapse example** manifest (reference-only; Synapse schema lives in Synapse
  repo).
- [ ] Document the SemVer range syntax accepted (consistent with existing Portunix
  conventions — npm-style `>=0.3.0 <1.0.0` or explicit `min/max`, decision deferred to
  WS-1).

### WS-7 — Tests

- [ ] Unit: manifest parse/validate for valid + malformed `supported_platforms[]`.
- [ ] Unit: SemVer range matching (in/out of range, edge cases).
- [ ] Unit: feature AND-filter (missing feature → plugin excluded).
- [ ] Integration: plugin without `supported_platforms` still installs and runs (backward
  compat).
- [ ] Integration: CLI `--platform` filter returns expected set.
- [ ] gRPC: `ListPluginsForPlatform` returns correct subset and round-trips
  `platform_payload`.

## Affected Components

- `CassandraGargoyle/api/contract/schemas/plugin-manifest.schema.json` — schema addition
  (v1.0.0 → v1.1.0)
- `src/app/plugins/types.go` — new `SupportedPlatform` struct + `SupportedPlatforms` field
- `src/app/plugins/manifest.go` — validation of `supported_platforms[]`
- `src/app/plugins/manager/registry.go` — storage, round-trip, `ListPluginsForPlatform`
  package function
- **NEW** `src/app/plugins/proto/plugin_registry.proto` — new `PluginRegistryService`
  (separate from `plugin.proto`, revised per DEC-1)
- **NEW** `src/helpers/ptx-plugin-registry/` — helper binary with `serve` subcommand
  (unix socket / TCP loopback)
- `src/cmd/plugin.go` — CLI flags `--platform`, `--platform-version`, `--feature`
- `src/dispatcher/dispatcher.go` — register `ptx-plugin-registry` helper
- `Makefile`, `build-with-version.sh`, `.goreleaser.yml` — helper binary build
- `src/app/selfinstall/install.go`, `src/app/update/github.go` — install helper binary
- `scripts/deploy-local.py`, `scripts/undeploy-local.py`,
  `scripts/create-platform-archives.py` — deploy helper binary
- `docs/` — plugin developer guide update with `supported_platforms` section

## Scope Boundaries

### In scope

- Schema addition in `api/contract/schemas/`.
- Parsing + validation at `ptx plugin install`.
- Registry storage of parsed `supported_platforms`.
- CLI + gRPC query APIs as described.
- Backward-compat tests for plugins without `supported_platforms`.
- Developer docs + Synapse example.

### Out of scope for this issue

- Synapse-side integration (owned by Synapse WS-2 under Synapse Issue 011).
- `platform_payload` **content** validation — each platform validates against its own
  schema.
- Runtime dynamic query (`QueryPlatformSupport`) — stretch only, not part of first
  delivery.
- Per-tenant isolation in the plugin registry — separate concern (Synapse UC-004 §4a,
  Phase 2).

## Constraints and Notes

- **One-version-per-plugin** is an existing Portunix constraint — this issue does **not**
  attempt to change it. The query API returns exactly one version per plugin, which is
  the de-facto current behaviour.
- **SemVer strings** for both plugin and platform versions. Range syntax to be finalised
  in WS-1 consistent with existing Portunix conventions.
- **Stability** — once merged, Synapse ingests on startup. Breaking changes to field names
  or semantics require coordinated cross-team release per `ARCHITECTURE-SYNC.md`.
- **Dispatcher pattern** — CLI changes flow through the main portunix dispatcher into the
  plugin subcommand; no new top-level binary is introduced.

## Acceptance Criteria

- [ ] `plugin-manifest.schema.json` v1.1.0 merged in `CassandraGargoyle/api` with
  `supported_platforms` additive field
- [ ] Plugins without `supported_platforms` install and run unchanged (regression test)
- [ ] Plugins with malformed `supported_platforms[]` are rejected at install time with a
  clear error
- [ ] `ptx plugin list --platform=synapse --platform-version=0.3.2` returns only plugins
  in SemVer range, with `platform_payload` in the output
- [ ] `--feature` flag AND-filters results correctly
- [ ] `ListPluginsForPlatform` gRPC RPC returns the same set as the CLI with
  byte-identical `platform_payload`
- [ ] Portunix plugin developer guide documents the new field with a Synapse example
- [ ] Synapse (driver team) can call `ListPluginsForPlatform("synapse", …)` against a
  local ptx, receive expected plugins, and drop their classpath fallback

## Hand-over / Coordination

- Upon acceptance, update the status line in the Synapse ask file
  ([012-portunix-plugin-platform-capability-query.md](../../../../portunix-architecture/docs/architecture/components/portunix-synapse/issues/012-portunix-plugin-platform-capability-query.md))
  to `accepted-by-portunix`.
- Schema PR in `CassandraGargoyle/api` must reference the Synapse ask (full URL).
- Notify Synapse architect (Zdenek) when the schema lands so Synapse Issue 011 WS-2 can
  track removal of the classpath fallback.

## Related Issues

- [#007](007-plugin-system-grpc.md) — Plugin System with gRPC Architecture
- [#024](210-plugin-registration-system.md) — Plugin Registration and Discovery System
- [#155](155-plugin-prerequisites-validation.md) — Plugin Prerequisites Validation
- [#162](162-plugin-interfaces-field-misplacement.md) — Plugin Interfaces Field
  Misplacement (schema/struct alignment precedent)

## References

- **Source ask:** Synapse Issue 012 —
  `portunix-architecture/docs/architecture/components/portunix-synapse/issues/012-portunix-plugin-platform-capability-query.md`
- Synapse UC-004 §4a — Phase 1 Delivery as Portunix Plugins
- Synapse Issue 011 — Extensions Platform Epic (WS-2 consumes this interface)
- Contract: `api/contract/schemas/plugin-manifest.schema.json` (v1.0.0 → target v1.1.0)
- Architecture sync policy: `docs/contributing/ARCHITECTURE-SYNC.md`

---

**Created:** 2026-04-19
**Assigned:** TBD
**Last Updated:** 2026-04-19
