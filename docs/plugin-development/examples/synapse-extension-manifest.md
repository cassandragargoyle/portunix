---
title: Synapse Extension Plugin Manifest Example
category: Plugin Development / Examples
related_issue: 175
---

# Synapse Extension Plugin — Manifest Example

This is a **reference-only** manifest showing how a Portunix plugin declares
itself as a Synapse extension under the `supported_platforms[]` field added in
plugin-manifest schema v1.1.0 (issue #175).

The `platform_payload` block is **opaque to Portunix** — Synapse owns its
schema and validates the content on its side. This example reflects the shape
Synapse currently expects for connector-command extensions; for the
authoritative definition consult the Synapse repository.

## Example `plugin.json`

```json
{
  "name": "erp-connector",
  "version": "1.2.0",
  "description": "ERP connector extension for Portunix Synapse",
  "author": "acme-integration",
  "license": "MIT",
  "plugin": {
    "type": "grpc",
    "binary": "./erp-connector",
    "runtime": "native",
    "port": 9042,
    "health_check_interval": 30000000000
  },
  "dependencies": {
    "portunix_min_version": "2.2.4",
    "os_support": ["linux", "darwin"]
  },
  "permissions": {
    "filesystem": ["read"],
    "network": ["outbound"],
    "level": "standard"
  },
  "commands": [
    {
      "name": "help",
      "description": "Show help for the ERP connector plugin",
      "examples": ["portunix erp-connector help"]
    }
  ],
  "supported_platforms": [
    {
      "name": "synapse",
      "min_version": "0.3.0",
      "max_version": "0.9.9",
      "features": [
        "connector.command",
        "connector.config-schema"
      ],
      "platform_payload": {
        "surfaces": {
          "connector.command": true
        },
        "command": {
          "supportedCapabilities": [
            "erp.customer.lookup",
            "erp.invoice.create",
            "erp.invoice.status"
          ]
        },
        "configSchema": "config.schema.json",
        "secretsSchema": "secrets.schema.json"
      }
    }
  ]
}
```

## Discovery from Synapse

At startup, Synapse queries the local Portunix registry for plugins targeting
it. Two entry points produce the same dataset — the Go matcher is the single
source of truth (see `Registry.ListPluginsForPlatform`).

### CLI path (recommended for one-shot discovery)

```bash
portunix plugin list --platform=synapse --platform-version=0.3.5 \
                     --feature=connector.command -o json
```

### gRPC path (recommended for long-lived callers)

Synapse starts `ptx-plugin-registry serve` as a subprocess for the duration of
the query window, then shuts it down.

```bash
portunix plugin-registry serve --mode unix                            # Linux/macOS
portunix plugin-registry serve --mode tcp --port 9500 --bind 127.0.0.1  # Windows
```

The gRPC contract is defined in
`src/app/plugins/proto/plugin_registry.proto` (service
`PluginRegistryService`, RPC `ListPluginsForPlatform`). The resulting
`MatchedPlugin.platform_payload_json` is transported as raw JSON bytes to
preserve byte-identical round-trip from the manifest.

## Backward compatibility

A plugin that omits `supported_platforms` installs and runs exactly as before;
older Portunix releases ignore the field entirely. Synapse simply does not see
such plugins in its discovery query, which matches the intent — the plugin
does not target Synapse.
