# ptx-plugin-registry — Plugin Platform-Capability gRPC Daemon

## Overview

`ptx-plugin-registry` is the Portunix-side gRPC daemon that answers
platform-capability queries about locally installed plugins. Hosting platforms
(Portunix **Synapse**, Pack, Agent, …) use it to discover which of the
installed Portunix plugins target them and retrieve each plugin's
platform-specific registration payload declared under
`supported_platforms[]` in the plugin manifest (schema v1.1.0+).

See Issue #175 and the design decisions captured in
`docs/issues/internal/done/175-plugin-platform-capability-query.md` for the
full rationale. Key points:

- The daemon serves `PluginRegistryService` defined in
  `src/app/plugins/proto/plugin_registry.proto` — a service **distinct** from
  `PluginService` (which is implemented by plugins).
- `ptx-plugin-registry` runs **on-demand**, not as a long-running system
  service. The consumer (e.g. Synapse) starts it for the duration of a
  discovery call and stops it afterwards. There is no systemd integration in
  v1.
- Matching is delegated to `manager.Registry.ListPluginsForPlatform` so the
  CLI (`portunix plugin list --platform=…`) and the gRPC RPC return the same
  dataset from the same source code path.

## Commands

| Command | Purpose |
| ------- | ------- |
| `portunix plugin-registry serve` | Start the gRPC server (blocking) |

### `serve` flags

| Flag | Default | Purpose |
| ---- | ------- | ------- |
| `--mode`, `-m` | `unix` on Linux/macOS, `tcp` on Windows | Transport selection |
| `--socket`, `-s` | `$XDG_RUNTIME_DIR/portunix/plugin-registry.sock` (fallback `$HOME/.portunix/run/...`) | Unix socket path |
| `--port`, `-p` | `9500` | TCP port (mode=tcp) |
| `--bind` | `127.0.0.1` | TCP bind address; loopback-only |
| `--plugin-dir` | `$HOME/.portunix/plugins` | Override plugin install directory |

## Transport and security

- **Unix domain socket** on Linux/macOS. The socket is created with
  permissions `0600` under `$XDG_RUNTIME_DIR/portunix/` (or a fallback under
  `$HOME/.portunix/run/`). Filesystem permissions provide localhost isolation
  — no auth tokens needed for same-user access.
- **TCP loopback** on Windows (no unix socket support pre-Win10). Bound to
  `127.0.0.1` only; remote access is explicitly out of scope.
- Stale sockets from a previous unclean shutdown are removed on startup.
- Graceful shutdown on SIGINT/SIGTERM (the server drains in-flight RPCs
  before closing).

## Example usage

```bash
# Start the daemon (unix socket, blocking — typically as a background subprocess)
portunix plugin-registry serve &
REG_PID=$!

# Query from a gRPC client (pseudocode — Synapse performs this programmatically)
#   client := pluginregistry.NewPluginRegistryServiceClient(conn)
#   resp, _ := client.ListPluginsForPlatform(ctx, &pb.ListPluginsForPlatformRequest{
#       Platform: "synapse",
#       PlatformVersion: "0.3.2",
#       Features: []string{"connector.command"},
#   })

# Tear down when done
kill $REG_PID
```

For the CLI-only path (which does not require starting the daemon), see
`portunix plugin list --platform=<name> -o json`.

## Related

- Plugin developer guide section:
  [Declaring Platform Support](../plugin-development/getting-started.md#declaring-platform-support-schema-v110-issue-175)
- Synapse extension manifest example:
  [Synapse Extension Plugin Manifest](../plugin-development/examples/synapse-extension-manifest.md)
- Proto contract: `src/app/plugins/proto/plugin_registry.proto`
- Issue: `docs/issues/internal/done/175-plugin-platform-capability-query.md`
