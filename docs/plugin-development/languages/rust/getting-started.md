# Rust Plugin Development Guide

Rust offers memory safety, fearless concurrency, and excellent performance for Portunix plugins. This guide walks you through creating
Rust plugins using `tonic` (async gRPC for Rust) and `tokio`.

## Prerequisites

- Rust 1.85 or later — install with `portunix install rust --variant=stable` (rustup).
  On Ubuntu/Debian the plain `portunix install rust` installs the distribution
  package (Rust 1.75 on Ubuntu 24.04), which cannot build the template
- Protocol Buffers compiler (`protoc`) — required by `tonic-build` in `build.rs`;
  install with `portunix install protoc`
- Portunix development environment

## Quick Start

Start from a copy of the [template](template/):

```bash
cp -r docs/plugin-development/languages/rust/template my-rust-plugin
cd my-rust-plugin
cargo build --release
```

The template has this structure:

```text
my-rust-plugin/
├── plugin.json            # Plugin manifest
├── Cargo.toml             # Rust manifest (library + binary)
├── Cargo.lock
├── build.rs               # tonic codegen build script
├── config.yaml            # Default runtime configuration
├── proto/
│   └── plugin.proto       # gRPC contract (portunix.plugins.PluginService)
├── src/
│   ├── lib.rs             # Library root — exposes all modules and the generated proto
│   ├── main.rs            # Entry point — CLI command or gRPC server
│   ├── config.rs          # Configuration loading
│   ├── manifest.rs        # plugin.json loader, maps it onto proto messages
│   ├── handlers.rs        # gRPC handler implementation
│   └── service.rs         # Business logic
├── tests/
│   └── plugin.rs          # Integration tests (business logic and gRPC handlers)
└── README.md
```

## Plugin Structure

### Cargo Manifest (Cargo.toml)

The crate is both a library (`my_rust_plugin`) and a binary (`plugin`). All modules
live in the library, so the integration tests in `tests/` can use the gRPC handlers
directly; `main.rs` only wires them together.

```toml
[package]
name = "my-rust-plugin"
version = "1.0.0"
edition = "2021"
rust-version = "1.85"
description = "Example Rust plugin for Portunix"
license = "MIT"

[lib]
name = "my_rust_plugin"
path = "src/lib.rs"

[[bin]]
name = "plugin"
path = "src/main.rs"

[dependencies]
tokio = { version = "1.37", features = ["full"] }
tonic = "0.11"
tonic-health = "0.11"
prost = "0.12"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
serde_yaml = "0.9"
anyhow = "1.0"
thiserror = "1.0"
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["env-filter"] }
clap = { version = "4.5", features = ["derive"] }

[build-dependencies]
tonic-build = "0.11"

[dev-dependencies]
tokio = { version = "1.37", features = ["macros", "rt-multi-thread"] }
```

### Build Script (build.rs)

```rust
fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::configure()
        .build_server(true)
        .build_client(false)
        .compile(&["proto/plugin.proto"], &["proto"])?;
    Ok(())
}
```

### gRPC Contract (proto/plugin.proto)

`proto/plugin.proto` is a copy of the `PluginService` part of the Portunix contract
(`src/app/plugins/proto/plugin.proto`). Keep its package, service and messages
unchanged, otherwise Portunix cannot talk to the plugin (it fails with `Unimplemented`
on `Initialize`):

```protobuf
syntax = "proto3";

package portunix.plugins;

// PluginService defines the main interface for plugin communication
service PluginService {
    rpc Initialize(InitializeRequest) returns (InitializeResponse);
    rpc Execute(ExecuteRequest) returns (ExecuteResponse);
    rpc GetInfo(GetInfoRequest) returns (GetInfoResponse);
    rpc Health(HealthRequest) returns (HealthResponse);
    rpc Shutdown(ShutdownRequest) returns (ShutdownResponse);
    rpc ListCommands(ListCommandsRequest) returns (ListCommandsResponse);
}
```

See the template file for the full set of messages (`PluginInfo`, `PluginCommand`,
`PluginCapabilities`, `PluginPermissions`, ...).

### Library Root (src/lib.rs)

```rust
pub mod config;
pub mod handlers;
pub mod manifest;
pub mod service;

pub mod proto {
    tonic::include_proto!("portunix.plugins");
}
```

### Main Entry Point (src/main.rs)

The binary has two modes:

- **Service mode** — `portunix plugin start` launches the binary with `--port N` and
  the plugin install directory as working directory. The plugin serves
  `portunix.plugins.PluginService`; Portunix calls `Initialize`, `Health` and
  `Execute`, and on `portunix plugin stop` calls `Shutdown` and then sends SIGTERM
- **CLI mode** — `portunix my-rust-plugin hello Portunix` runs the binary directly
  with the command as positional arguments; the command runs once, without gRPC

Logs go to stderr, so stdout carries only command output, and ANSI colors are
switched off when stderr is not a terminal (Portunix redirects it into the plugin
log file). The server binds to `127.0.0.1`, because Portunix connects over
localhost, and stops on Ctrl+C or, on Unix, SIGTERM.

```rust
use std::io::IsTerminal;
use std::net::SocketAddr;

use anyhow::Result;
use clap::Parser;
use tonic::transport::Server;
use tracing_subscriber::EnvFilter;

use my_rust_plugin::config::PluginConfig;
use my_rust_plugin::handlers::PluginServiceImpl;
use my_rust_plugin::manifest::PluginManifest;
use my_rust_plugin::proto::plugin_service_server::PluginServiceServer;
use my_rust_plugin::service::PluginBusinessService;

/// Without a command the plugin runs as a gRPC service (portunix plugin start);
/// with a command it runs once as a CLI (portunix <plugin> <command> [args...])
#[derive(Parser, Debug)]
#[command(name = "plugin", version)]
struct Args {
    #[arg(long, default_value_t = 50051)]
    port: u16,

    #[arg(long, default_value = "config.yaml")]
    config: String,

    #[arg(long, default_value = "plugin.json")]
    manifest: String,

    /// Command and its arguments (CLI mode)
    #[arg(trailing_var_arg = true, allow_hyphen_values = true)]
    command: Vec<String>,
}

#[tokio::main]
async fn main() -> Result<()> {
    let args = Args::parse();

    // Logs go to stderr, so stdout carries only command output; no colors
    // when Portunix redirects them into the plugin log file
    tracing_subscriber::fmt()
        .with_writer(std::io::stderr)
        .with_ansi(std::io::stderr().is_terminal())
        .with_env_filter(
            EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info")),
        )
        .init();

    let config = PluginConfig::load(&args.config)?;

    if let Some((command, rest)) = args.command.split_first() {
        match PluginBusinessService::new(config).execute(command, rest) {
            Ok(output) => println!("{output}"),
            Err(err) => {
                eprintln!("Error: {err}");
                std::process::exit(1);
            }
        }
        return Ok(());
    }

    // Portunix runs the plugin with the install directory as working
    // directory, so the default manifest path resolves to its plugin.json
    let manifest = PluginManifest::load(&args.manifest)?;
    serve(config, manifest, args.port).await
}

/// Runs the gRPC server implementing portunix.plugins.PluginService until
/// SIGINT or SIGTERM
async fn serve(config: PluginConfig, manifest: PluginManifest, port: u16) -> Result<()> {
    // Portunix connects over localhost; the plugin is not exposed to the network
    let addr = SocketAddr::from(([127, 0, 0, 1], port));

    let (mut health_reporter, health_service) = tonic_health::server::health_reporter();
    health_reporter
        .set_serving::<PluginServiceServer<PluginServiceImpl>>()
        .await;

    tracing::info!(
        "Plugin server {} {} listening on {}",
        manifest.name,
        manifest.version,
        addr
    );

    Server::builder()
        .add_service(health_service)
        .add_service(PluginServiceServer::new(PluginServiceImpl::new(
            config, manifest,
        )))
        .serve_with_shutdown(addr, shutdown_signal())
        .await?;

    Ok(())
}

/// Resolves on Ctrl+C or, on Unix, SIGTERM (sent by portunix plugin stop)
async fn shutdown_signal() {
    #[cfg(unix)]
    {
        use tokio::signal::unix::{signal, SignalKind};
        let mut term = signal(SignalKind::terminate()).expect("install SIGTERM handler");
        tokio::select! {
            _ = tokio::signal::ctrl_c() => {}
            _ = term.recv() => {}
        }
    }
    #[cfg(not(unix))]
    {
        let _ = tokio::signal::ctrl_c().await;
    }
    tracing::info!("Shutting down gracefully...");
}
```

Command-line flags:

| Flag | Default | Description |
| ----- | ----- | ----- |
| `--port` | `50051` | gRPC port (Portunix passes the port from `plugin.json`) |
| `--config` | `config.yaml` | Runtime configuration file |
| `--manifest` | `plugin.json` | Plugin manifest read in service mode |
| `--version` | | Print the binary version and exit |
| `<command> [args...]` | | Run one command in CLI mode |

### Configuration (src/config.rs)

`config.yaml` is optional; a missing file yields the defaults.

```rust
use anyhow::{Context, Result};
use serde::Deserialize;
use std::fs;
use std::path::Path;

#[derive(Debug, Clone, Default, Deserialize)]
pub struct PluginConfig {
    #[serde(default)]
    pub plugin: PluginSection,
    #[serde(default)]
    pub server: ServerSection,
}

#[derive(Debug, Clone, Deserialize)]
pub struct PluginSection {
    #[serde(default = "default_name")]
    pub name: String,
    #[serde(default = "default_version")]
    pub version: String,
    #[serde(default)]
    pub description: String,
    #[serde(default = "default_log_level")]
    pub log_level: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct ServerSection {
    #[serde(default = "default_port")]
    pub port: u16,
    #[serde(default = "default_health_port")]
    pub health_port: u16,
}

// Default impls and default_* functions omitted, see template/src/config.rs

impl PluginConfig {
    pub fn load<P: AsRef<Path>>(path: P) -> Result<Self> {
        let path = path.as_ref();
        if !path.exists() {
            return Ok(Self::default());
        }
        let raw = fs::read_to_string(path)
            .with_context(|| format!("reading config {}", path.display()))?;
        serde_yaml::from_str(&raw).with_context(|| format!("parsing config {}", path.display()))
    }
}
```

### Plugin Manifest Loader (src/manifest.rs)

Plugin information and commands reported over gRPC come from `plugin.json`, so the
manifest stays the single source. `PluginManifest` is deserialized with `serde_json`
(all fields optional except `name`) and maps itself onto the proto messages:

- `plugin_info()` — `PluginInfo` with name, version, description, author, license,
  `supported_os` (from `dependencies.os_support`), commands, capabilities and
  `required_permissions` (level defaults to `limited`)
- `plugin_commands()` — `PluginCommand` list from `commands` (parameter type defaults
  to `string`)
- `capabilities()` — `PluginCapabilities`; access flags are derived from the
  non-empty `permissions` lists and `mcp_tools` holds the names from
  `ai_integration.mcp_tools`

```rust
impl PluginManifest {
    pub fn load<P: AsRef<Path>>(path: P) -> Result<Self> {
        let path = path.as_ref();
        let raw = fs::read_to_string(path)
            .with_context(|| format!("reading manifest {}", path.display()))?;
        let manifest: Self = serde_json::from_str(&raw)
            .with_context(|| format!("parsing manifest {}", path.display()))?;
        if manifest.name.is_empty() {
            bail!("manifest {} has no plugin name", path.display());
        }
        Ok(manifest)
    }

    pub fn capabilities(&self) -> PluginCapabilities {
        let p = &self.permissions;
        PluginCapabilities {
            filesystem_access: !p.filesystem.is_empty(),
            network_access: !p.network.is_empty(),
            database_access: !p.database.is_empty(),
            container_access: false,
            system_commands: !p.system.is_empty(),
            mcp_tools: self
                .ai_integration
                .mcp_tools
                .iter()
                .map(|t| t.name.clone())
                .collect(),
        }
    }

    // plugin_info() and plugin_commands() omitted, see template/src/manifest.rs
}
```

### gRPC Handler (src/handlers.rs)

The handler implements `portunix.plugins.PluginService`: `initialize`, `get_info`,
`health`, `execute`, `list_commands` and `shutdown`.

- `execute` reports command errors in the response (`success = false`, `error`,
  `exit_code = 1`), not as gRPC errors, so Portunix can show them to the user
- `shutdown` only cleans up and answers. Portunix sends SIGTERM right after it
  (handled by `shutdown_signal` in `main.rs`); a plugin that exits inside `shutdown`
  drops the response and `portunix plugin stop` fails with `EOF`

```rust
use std::collections::HashMap;
use std::time::Instant;

use tonic::{Request, Response, Status};

use crate::config::PluginConfig;
use crate::manifest::PluginManifest;
use crate::proto::plugin_service_server::PluginService;
use crate::proto::{
    ExecuteRequest, ExecuteResponse, GetInfoRequest, GetInfoResponse, HealthRequest,
    HealthResponse, InitializeRequest, InitializeResponse, ListCommandsRequest,
    ListCommandsResponse, ShutdownRequest, ShutdownResponse,
};
use crate::service::PluginBusinessService;

pub struct PluginServiceImpl {
    manifest: PluginManifest,
    service: PluginBusinessService,
    started_at: Instant,
}

impl PluginServiceImpl {
    pub fn new(config: PluginConfig, manifest: PluginManifest) -> Self {
        Self {
            manifest,
            service: PluginBusinessService::new(config),
            started_at: Instant::now(),
        }
    }
}

#[tonic::async_trait]
impl PluginService for PluginServiceImpl {
    async fn initialize(
        &self,
        request: Request<InitializeRequest>,
    ) -> Result<Response<InitializeResponse>, Status> {
        tracing::info!("Initialize called by {}", request.get_ref().plugin_name);
        Ok(Response::new(InitializeResponse {
            success: true,
            message: "initialized".to_string(),
            plugin_info: Some(self.manifest.plugin_info()),
        }))
    }

    async fn get_info(
        &self,
        _request: Request<GetInfoRequest>,
    ) -> Result<Response<GetInfoResponse>, Status> {
        Ok(Response::new(GetInfoResponse {
            plugin_info: Some(self.manifest.plugin_info()),
        }))
    }

    async fn health(
        &self,
        _request: Request<HealthRequest>,
    ) -> Result<Response<HealthResponse>, Status> {
        Ok(Response::new(HealthResponse {
            healthy: true,
            status: "running".to_string(),
            message: "ok".to_string(),
            uptime_seconds: self.started_at.elapsed().as_secs() as i64,
            metrics: HashMap::new(),
        }))
    }

    // Command errors are reported in the response, not as gRPC errors, so
    // Portunix can show them to the user
    async fn execute(
        &self,
        request: Request<ExecuteRequest>,
    ) -> Result<Response<ExecuteResponse>, Status> {
        let req = request.into_inner();
        let response = match self.service.execute(&req.command, &req.args) {
            Ok(output) => ExecuteResponse {
                success: true,
                output,
                exit_code: 0,
                ..Default::default()
            },
            Err(err) => ExecuteResponse {
                success: false,
                message: err.to_string(),
                error: err.to_string(),
                exit_code: 1,
                ..Default::default()
            },
        };
        Ok(Response::new(response))
    }

    async fn list_commands(
        &self,
        _request: Request<ListCommandsRequest>,
    ) -> Result<Response<ListCommandsResponse>, Status> {
        Ok(Response::new(ListCommandsResponse {
            commands: self.manifest.plugin_commands(),
            capabilities: Some(self.manifest.capabilities()),
        }))
    }

    // Portunix stops the process right after this call (SIGTERM, handled in
    // main.rs); exiting here would drop the response before it is sent
    async fn shutdown(
        &self,
        _request: Request<ShutdownRequest>,
    ) -> Result<Response<ShutdownResponse>, Status> {
        self.service.cleanup();
        Ok(Response::new(ShutdownResponse {
            success: true,
            message: "cleanup done".to_string(),
        }))
    }
}
```

### Business Logic (src/service.rs)

The same `PluginBusinessService` serves both modes: `main.rs` calls it directly in
CLI mode and `handlers.rs` calls it from `execute` in service mode.

```rust
use thiserror::Error;
use tracing::info;

use crate::config::PluginConfig;

#[derive(Debug, Error)]
pub enum ServiceError {
    #[error("unknown command: {0}")]
    UnknownCommand(String),
    #[error("process command requires at least one argument")]
    MissingArgument,
}

pub struct PluginBusinessService {
    #[allow(dead_code)]
    config: PluginConfig,
}

impl PluginBusinessService {
    pub fn new(config: PluginConfig) -> Self {
        Self { config }
    }

    pub fn execute(&self, command: &str, args: &[String]) -> Result<String, ServiceError> {
        match command {
            "hello" => Ok(self.handle_hello(args)),
            "process" => self.handle_process(args),
            other => Err(ServiceError::UnknownCommand(other.to_string())),
        }
    }

    fn handle_hello(&self, args: &[String]) -> String {
        let name = args.first().map(String::as_str).unwrap_or("World");
        format!("Hello, {name}!")
    }

    fn handle_process(&self, args: &[String]) -> Result<String, ServiceError> {
        if args.is_empty() {
            return Err(ServiceError::MissingArgument);
        }
        Ok(format!("Processed: {}", args.join(", ")))
    }

    pub fn cleanup(&self) {
        info!("Performing cleanup operations...");
    }
}
```

## Building and Testing

### Build

```bash
# Debug build
cargo build

# Release build (plugin.json points to target/release/plugin)
cargo build --release

# Run a command once (CLI mode)
cargo run -- hello Portunix

# Run as a gRPC service
cargo run -- --port 50051 --config config.yaml --manifest plugin.json
```

### Tests

```bash
cargo test
cargo test --release
```

#### Test Example (tests/plugin.rs)

The tests cover the business logic and, because the handlers live in the library,
the gRPC handlers too — called directly, without starting a server. Cargo runs the
tests from the template root, where `plugin.json` is:

```rust
use my_rust_plugin::config::PluginConfig;
use my_rust_plugin::service::PluginBusinessService;

#[test]
fn hello_returns_default_greeting() {
    let service = PluginBusinessService::new(PluginConfig::default());
    assert_eq!(service.execute("hello", &[]).unwrap(), "Hello, World!");
}

mod grpc {
    use my_rust_plugin::config::PluginConfig;
    use my_rust_plugin::handlers::PluginServiceImpl;
    use my_rust_plugin::manifest::PluginManifest;
    use my_rust_plugin::proto::plugin_service_server::PluginService;
    use my_rust_plugin::proto::ExecuteRequest;
    use tonic::Request;

    // Cargo runs tests from the template root, where plugin.json is
    fn handler() -> PluginServiceImpl {
        let manifest = PluginManifest::load("plugin.json").expect("load plugin.json");
        PluginServiceImpl::new(PluginConfig::default(), manifest)
    }

    #[tokio::test]
    async fn execute_reports_output_and_errors() {
        let h = handler();
        let ok = h
            .execute(Request::new(ExecuteRequest {
                command: "hello".to_string(),
                args: vec!["Portunix".to_string()],
                ..Default::default()
            }))
            .await
            .unwrap()
            .into_inner();
        assert!(ok.success);
        assert_eq!(ok.output, "Hello, Portunix!");

        let failed = h
            .execute(Request::new(ExecuteRequest {
                command: "unknown".to_string(),
                ..Default::default()
            }))
            .await
            .unwrap()
            .into_inner();
        assert!(!failed.success);
        assert_ne!(failed.exit_code, 0);
    }
}
```

The template file also tests `initialize`, `health`, `list_commands` and `shutdown`.

## Development Workflow

```bash
# Format + lint
cargo fmt
cargo clippy -- -D warnings

# Run with hot-reload (requires cargo-watch)
cargo install cargo-watch
cargo watch -x 'run -- --config config.yaml'

# Install into Portunix
cargo build --release
portunix plugin install .
portunix plugin enable my-rust-plugin

# CLI mode
portunix my-rust-plugin hello Portunix

# Service mode
portunix plugin start my-rust-plugin
portunix plugin stop my-rust-plugin
```

## Advanced Features

### MCP Tool Exposure

The Portunix contract has no separate tool RPCs. Declare MCP tools in `plugin.json`
under `ai_integration.mcp_tools`; the template reports their names in
`PluginCapabilities.mcp_tools` of `Initialize`, `GetInfo` and `ListCommands`, and a
tool call reaches the plugin as an `Execute` request with the tool name as `command`:

```json
{
  "ai_integration": {
    "mcp_tools": [
      { "name": "process_file", "description": "Process a file with custom logic" }
    ]
  }
}
```

```rust
// src/service.rs — handle the tool like any other command
"process_file" => Ok(format!("Processed file: {}", args.join(" "))),
```

### Logging

The template uses `tracing` + `tracing-subscriber` and writes logs to stderr. The
default level is `info`; set `RUST_LOG` for more detail:

```bash
RUST_LOG=debug cargo run
```

## Deployment

### Dockerfile

```dockerfile
FROM rust:1.85 AS builder
RUN apt-get update && apt-get install -y protobuf-compiler && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/plugin /usr/local/bin/plugin
COPY --from=builder /app/plugin.json /app/plugin.json
COPY --from=builder /app/config.yaml /app/config.yaml
WORKDIR /app
ENTRYPOINT ["plugin"]
```

The server binds to `127.0.0.1`, so in a container it is reachable only from inside
the container; the image is mainly useful for building the binary or running
commands in CLI mode.

### Plugin Manifest

Rust plugins use the `native` runtime. `binary` is the path of the compiled binary
relative to the plugin directory:

```json
{
  "plugin": {
    "type": "grpc",
    "binary": "./target/release/plugin",
    "runtime": "native",
    "port": 9001,
    "health_check_interval": 30000000000
  }
}
```

Portunix runs the binary with the plugin install directory as working directory, so
the default `--manifest plugin.json` and `--config config.yaml` resolve to the
installed files. See `template/plugin.json` for the complete manifest.

## Troubleshooting

1. **`lock file version 4 requires -Znext-lockfile-bump`** or **`feature edition2024
   is required`** — the toolchain is older than Rust 1.85; install a current one with
   `portunix install rust --variant=stable` and make sure `~/.cargo/bin` comes first
   in `PATH` (`source ~/.cargo/env`)
2. **`protoc` not found** — install it via the system package manager:
   `apt install protobuf-compiler` / `brew install protobuf` / `choco install protoc`.
3. **`tonic-build` codegen fails** — pin `tonic-build` and `tonic` to matching versions
   (both 0.11 in this template).
4. **`Unimplemented` on `Initialize`** — `proto/plugin.proto` no longer matches the
   Portunix contract; restore package `portunix.plugins` and service `PluginService`.
5. **`portunix plugin stop` fails with `EOF`** — the plugin exits inside `shutdown`;
   only clean up there and let the SIGTERM handler stop the server.
6. **`reading manifest plugin.json` at startup** — the binary was started in service
   mode outside the plugin directory; pass `--manifest <path>`.

## Next Steps

- Study the [template code](template/) for a complete example
- Learn about [MCP integration](../../mcp-integration/exposing-tools.md) for AI agents
- Review the [plugin architecture](../../architecture.md)

## Resources

- [tonic (gRPC for Rust)](https://github.com/hyperium/tonic)
- [tokio runtime](https://tokio.rs/)
- [Protocol Buffers Rust tutorial](https://protobuf.dev/getting-started/rustcrate/)
