# plugins

Package `plugins` is the core of Portunix's plugin system. It defines the
plugin contract, loads and validates plugin manifests, and speaks gRPC to
out-of-process plugin binaries. It also exposes a registry service that
hosting platforms (Synapse, Pack, Agent, …) query to discover which installed
plugins target them.

CLI entry point: `portunix plugin …` (see `src/cmd/plugin.go`).

## Architecture

Plugins are **separate processes** that Portunix launches and communicates
with over **gRPC on localhost**. This keeps plugins isolated (own memory,
own runtime — native, Java, or Python) while giving Portunix a uniform API.

Two gRPC services are defined:

- **`PluginService`** (`proto/plugin.proto`) — implemented **by each plugin**;
  Portunix is the client. Methods: `Initialize`, `Execute`, `GetInfo`,
  `Health`, `Shutdown`, `ListCommands`.
- **`PluginRegistryService`** (`proto/plugin_registry.proto`) — implemented
  **by Portunix** (served by the `ptx-plugin-registry` helper); hosting
  platforms are clients. Methods: `ListPluginsForPlatform`, `HealthCheck`.
  See issue #175 (DEC-1/DEC-2) for rationale.

```text
           ┌──────────────┐       ┌──────────────┐
platforms  │   Synapse    │──▶│ PluginRegistry │──▶ registry.json
 (clients) │   Pack / …   │       │   Service    │
           └──────────────┘       └──────────────┘
                                         ▲
                                         │ served by ptx-plugin-registry
                                   ┌─────┴─────┐
                                   │ portunix  │
                                   │  plugin   │
                                   │  manager  │
                                   └─────┬─────┘
                                         │ PluginService (gRPC client)
                     ┌───────────────────┼───────────────────┐
                     ▼                   ▼                   ▼
               ┌──────────┐       ┌──────────┐        ┌──────────┐
               │ plugin A │       │ plugin B │        │ plugin C │
               │ (native) │       │  (java)  │        │ (python) │
               └──────────┘       └──────────┘        └──────────┘
```

## Layout

```text
src/app/plugins/
├── types.go            # Plugin interface, PluginConfig, PluginInfo, PluginManifest, …
├── manifest.go         # LoadManifest / SaveManifest / ValidateManifest
├── grpc_client.go      # GRPCPlugin — client-side impl of Plugin over gRPC
├── prerequisites.go    # Runtime (Java/Python/native) availability checks
├── semver.go           # SemVer parsing + range matching for supported_platforms
├── manager/            # Plugin lifecycle & registry persistence
│   ├── manager.go      # Manager: install/enable/start/stop/uninstall, health loop
│   └── registry.go     # Registry: JSON-backed registry.json
├── proto/              # gRPC definitions and generated code
│   ├── plugin.proto           # PluginService (plugin → ptx)
│   ├── plugin_registry.proto  # PluginRegistryService (ptx → platforms)
│   ├── generated.go
│   └── pluginregistry/        # generated code for the registry service
└── server/
    └── registry_server.go     # gRPC server for PluginRegistryService
```

## Key types

- `Plugin` (interface) — contract every plugin fulfils:
  `Initialize`, `Start`, `Stop`, `Execute`, `GetInfo`, `Health`, `IsRunning`.
- `PluginConfig` — runtime configuration passed to a plugin (binary path,
  runtime kind & version, port, permissions, env, working dir).
- `PluginInfo` — metadata returned by `GetInfo` (commands, capabilities,
  required permissions, supported OS, mode — `service` or `helper`).
- `PluginManifest` — on-disk manifest shape
  (`plugin-manifest.schema.json v1.1.0` in the `api/contract` repo).
- `GRPCPlugin` — default `Plugin` implementation: spawns the plugin binary,
  dials the gRPC port, and forwards interface calls.
- `manager.Manager` — orchestrates the full lifecycle and drives the
  periodic health-check loop.
- `manager.Registry` — persists the set of installed plugins in
  `registry.json`.

## Manifest

Plugins ship a JSON manifest describing their identity, commands, supported
platforms, runtime requirements, and required permissions. `LoadManifest`
reads the file and `ValidateManifest` enforces:

- SemVer-shaped `version`
- `supported_platforms[].name` matching `^[a-z][a-z0-9_-]*$`
- `supported_platforms[].features[]` matching `^[a-z][a-z0-9._-]*$`

Schemas live in the shared contract repo
(`../api/contract/schemas/`).

## Runtime prerequisites

`prerequisites.go` probes the host for the runtime the manifest declares
(`native`, `java`, `python`) and checks version ranges against
`runtime_version` (e.g. `">=21"`). Results are surfaced through `ptx plugin
check` before install/start, preventing opaque failures at launch time.

## Related

- CLI wiring: `src/cmd/plugin.go`
- Platform-query helper: `ptx-plugin-registry` (see issue #175)
- Manifest schema: `api/contract/schemas/plugin-manifest.schema.json`
- Feature overview: `docs/FEATURES_OVERVIEW.md` (§ Plugin System)
