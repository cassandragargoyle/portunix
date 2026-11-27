# Rust Plugin Template

Template for Rust-based Portunix plugins using `tonic` (async gRPC) and `tokio`. The
plugin implements the Portunix contract `portunix.plugins.PluginService`.

## Prerequisites

- Rust 1.85 or later: `portunix install rust --variant=stable` (on Ubuntu/Debian
  the plain `portunix install rust` installs the older distribution package)
- `protoc`: `portunix install protoc`

## Quick Start

```bash
cargo build --release
cargo test

# Run a command once (CLI mode)
./target/release/plugin hello Portunix

# Run as a gRPC service on 127.0.0.1
./target/release/plugin --port 50051 --config config.yaml --manifest plugin.json

# Install into Portunix
portunix plugin install .
portunix plugin enable my-rust-plugin
portunix my-rust-plugin hello Portunix
portunix plugin start my-rust-plugin
portunix plugin stop my-rust-plugin
```

`portunix plugin start` runs the binary with `--port N` in the install directory and
calls `Initialize`, `Health` and `Execute`; `portunix plugin stop` calls `Shutdown`
and then sends SIGTERM. `shutdown` must only clean up and answer — the process ends on
the SIGTERM. Plugin info and commands are read from `plugin.json`.

## Structure

```text
.
├── plugin.json          # Plugin manifest consumed by Portunix
├── Cargo.toml           # Rust manifest (library my_rust_plugin + binary plugin)
├── build.rs             # tonic-build proto codegen
├── config.yaml          # Default runtime configuration
├── proto/plugin.proto   # Copy of the Portunix PluginService contract
├── src/
│   ├── lib.rs           # Library root (config, handlers, manifest, service, proto)
│   ├── main.rs          # Entry point — CLI command or gRPC server
│   ├── config.rs        # YAML config loader
│   ├── manifest.rs      # plugin.json loader, maps it onto proto messages
│   ├── handlers.rs      # PluginService trait implementation
│   └── service.rs       # Business logic
└── tests/plugin.rs      # Tests for business logic and gRPC handlers
```

See `../getting-started.md` for the full walkthrough.
