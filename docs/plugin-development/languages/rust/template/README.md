# Rust Plugin Template

Template for Rust-based Portunix plugins using `tonic` (async gRPC) and `tokio`.

## Quick Start

```bash
cargo build --release
./target/release/plugin --port 50051 --config config.yaml

# Install into Portunix
portunix plugin install .
portunix plugin enable my-rust-plugin
portunix plugin start my-rust-plugin
```

## Structure

```text
.
├── plugin.json          # Plugin manifest consumed by Portunix
├── Cargo.toml           # Rust manifest
├── build.rs             # tonic-build proto codegen
├── config.yaml          # Default runtime configuration
├── proto/plugin.proto   # gRPC service definition
├── src/
│   ├── lib.rs           # Re-exports for tests (config, service, proto)
│   ├── main.rs          # Entry point — boots tonic server
│   ├── config.rs        # YAML config loader
│   ├── handlers.rs      # gRPC trait implementation
│   └── service.rs       # Business logic
└── tests/plugin.rs      # Integration tests
```

See `../getting-started.md` for the full walkthrough.
