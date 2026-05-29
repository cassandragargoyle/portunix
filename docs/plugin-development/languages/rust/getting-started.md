# Rust Plugin Development Guide

Rust offers memory safety, fearless concurrency, and excellent performance for Portunix plugins. This guide walks you through creating
Rust plugins using `tonic` (async gRPC for Rust) and `tokio`.

## Prerequisites

- Rust 1.70 or later (1.65 minimum) — install via [rustup](https://rustup.rs/)
- Protocol Buffers compiler (`protoc`)
- Portunix development environment

## Quick Start

Create a new Rust plugin using the Portunix CLI:

```bash
portunix plugin create my-rust-plugin --language=rust
cd my-rust-plugin
cargo build
```

This creates a complete Rust project structure:

```text
my-rust-plugin/
├── plugin.json            # Plugin manifest
├── Cargo.toml             # Rust manifest
├── Cargo.lock
├── build.rs               # tonic codegen build script
├── proto/
│   └── plugin.proto       # gRPC service definition
├── src/
│   ├── main.rs            # Entry point
│   ├── config.rs          # Configuration loading
│   ├── handlers.rs        # gRPC handler implementation
│   └── service.rs         # Business logic
├── tests/
│   └── plugin.rs          # Integration tests
└── README.md
```

## Plugin Structure

### Cargo Manifest (Cargo.toml)

```toml
[package]
name = "my-rust-plugin"
version = "1.0.0"
edition = "2021"
description = "Example Rust plugin for Portunix"
license = "MIT"

[[bin]]
name = "plugin"
path = "src/main.rs"

[dependencies]
tokio = { version = "1.37", features = ["full"] }
tonic = "0.11"
tonic-health = "0.11"
prost = "0.12"
serde = { version = "1.0", features = ["derive"] }
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

### Main Entry Point (src/main.rs)

```rust
use anyhow::Result;
use clap::Parser;
use tonic::transport::Server;
use tracing_subscriber::EnvFilter;

mod config;
mod handlers;
mod service;

pub mod proto {
    tonic::include_proto!("rust_plugin_template");
}

use config::PluginConfig;
use handlers::PluginServiceImpl;
use proto::plugin_service_server::PluginServiceServer;

#[derive(Parser, Debug)]
#[command(name = "plugin")]
struct Args {
    #[arg(long, default_value_t = 50051)]
    port: u16,

    #[arg(long, default_value = "config.yaml")]
    config: String,
}

#[tokio::main]
async fn main() -> Result<()> {
    let args = Args::parse();

    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .init();

    let config = PluginConfig::load(&args.config)?;
    let addr = format!("0.0.0.0:{}", args.port).parse()?;

    let (mut health_reporter, health_service) = tonic_health::server::health_reporter();
    health_reporter
        .set_serving::<PluginServiceServer<PluginServiceImpl>>()
        .await;

    let plugin_service = PluginServiceImpl::new(config.clone());

    tracing::info!("Plugin server listening on {}", addr);

    Server::builder()
        .add_service(health_service)
        .add_service(PluginServiceServer::new(plugin_service))
        .serve_with_shutdown(addr, shutdown_signal())
        .await?;

    Ok(())
}

async fn shutdown_signal() {
    tokio::signal::ctrl_c().await.ok();
    tracing::info!("Shutting down gracefully...");
}
```

### Configuration (src/config.rs)

```rust
use anyhow::{Context, Result};
use serde::Deserialize;
use std::fs;
use std::path::Path;

#[derive(Debug, Clone, Deserialize)]
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

impl Default for PluginSection {
    fn default() -> Self {
        Self {
            name: default_name(),
            version: default_version(),
            description: String::new(),
            log_level: default_log_level(),
        }
    }
}

impl Default for ServerSection {
    fn default() -> Self {
        Self {
            port: default_port(),
            health_port: default_health_port(),
        }
    }
}

fn default_name() -> String { "my-rust-plugin".to_string() }
fn default_version() -> String { "1.0.0".to_string() }
fn default_log_level() -> String { "info".to_string() }
fn default_port() -> u16 { 50051 }
fn default_health_port() -> u16 { 50052 }

impl PluginConfig {
    pub fn load<P: AsRef<Path>>(path: P) -> Result<Self> {
        let path = path.as_ref();
        if !path.exists() {
            return Ok(Self {
                plugin: PluginSection::default(),
                server: ServerSection::default(),
            });
        }
        let raw = fs::read_to_string(path)
            .with_context(|| format!("reading config {}", path.display()))?;
        serde_yaml::from_str(&raw)
            .with_context(|| format!("parsing config {}", path.display()))
    }
}
```

### gRPC Handler (src/handlers.rs)

```rust
use tonic::{Request, Response, Status};

use crate::config::PluginConfig;
use crate::proto::plugin_service_server::PluginService;
use crate::proto::{
    execute_response, health_check_response, shutdown_response, CallToolRequest, CallToolResponse,
    ExecuteRequest, ExecuteResponse, GetInfoRequest, GetInfoResponse, HealthCheckRequest,
    HealthCheckResponse, ListToolsRequest, ListToolsResponse, PluginInfo, ShutdownRequest,
    ShutdownResponse,
};
use crate::service::PluginBusinessService;

pub struct PluginServiceImpl {
    config: PluginConfig,
    service: PluginBusinessService,
}

impl PluginServiceImpl {
    pub fn new(config: PluginConfig) -> Self {
        let service = PluginBusinessService::new(config.clone());
        Self { config, service }
    }
}

#[tonic::async_trait]
impl PluginService for PluginServiceImpl {
    async fn get_info(
        &self,
        _request: Request<GetInfoRequest>,
    ) -> Result<Response<GetInfoResponse>, Status> {
        let info = PluginInfo {
            name: self.config.plugin.name.clone(),
            version: self.config.plugin.version.clone(),
            description: self.config.plugin.description.clone(),
            capabilities: vec!["example-capability".to_string()],
        };
        Ok(Response::new(GetInfoResponse { info: Some(info) }))
    }

    async fn health_check(
        &self,
        _request: Request<HealthCheckRequest>,
    ) -> Result<Response<HealthCheckResponse>, Status> {
        Ok(Response::new(HealthCheckResponse {
            status: health_check_response::Status::Serving as i32,
        }))
    }

    async fn execute(
        &self,
        request: Request<ExecuteRequest>,
    ) -> Result<Response<ExecuteResponse>, Status> {
        let req = request.into_inner();
        match self.service.execute(&req.command, &req.args) {
            Ok(result) => Ok(Response::new(ExecuteResponse {
                result,
                status: execute_response::Status::Success as i32,
                error_message: String::new(),
            })),
            Err(err) => Ok(Response::new(ExecuteResponse {
                result: String::new(),
                status: execute_response::Status::Error as i32,
                error_message: err.to_string(),
            })),
        }
    }

    async fn shutdown(
        &self,
        _request: Request<ShutdownRequest>,
    ) -> Result<Response<ShutdownResponse>, Status> {
        self.service.cleanup();
        Ok(Response::new(ShutdownResponse {
            status: shutdown_response::Status::Success as i32,
            message: String::new(),
        }))
    }

    async fn list_tools(
        &self,
        _request: Request<ListToolsRequest>,
    ) -> Result<Response<ListToolsResponse>, Status> {
        Ok(Response::new(ListToolsResponse { tools: vec![] }))
    }

    async fn call_tool(
        &self,
        _request: Request<CallToolRequest>,
    ) -> Result<Response<CallToolResponse>, Status> {
        Err(Status::not_found("No tools registered"))
    }
}
```

### Business Logic (src/service.rs)

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

# Release build
cargo build --release

# Run locally
cargo run -- --port 50051 --config config.yaml
```

### Tests

```bash
cargo test
cargo test --release
```

#### Unit Test Example (tests/plugin.rs)

```rust
use my_rust_plugin::config::PluginConfig;
use my_rust_plugin::service::PluginBusinessService;

#[test]
fn hello_returns_default_greeting() {
    let service = PluginBusinessService::new(PluginConfig::default());
    assert_eq!(service.execute("hello", &[]).unwrap(), "Hello, World!");
}

#[test]
fn unknown_command_errors() {
    let service = PluginBusinessService::new(PluginConfig::default());
    assert!(service.execute("unknown", &[]).is_err());
}
```

> **Note:** to expose `config` / `service` from the integration test, the template
> defines a `lib.rs` re-exporting both modules (see `template/src/lib.rs`).

## Development Workflow

```bash
# Format + lint
cargo fmt
cargo clippy -- -D warnings

# Run with hot-reload (requires cargo-watch)
cargo install cargo-watch
cargo watch -x 'run -- --config config.yaml'

# Install into Portunix
portunix plugin install .
portunix plugin enable my-rust-plugin
portunix plugin start my-rust-plugin
```

## Advanced Features

### MCP Tool Exposure

```rust
async fn list_tools(
    &self,
    _request: Request<ListToolsRequest>,
) -> Result<Response<ListToolsResponse>, Status> {
    let tool = crate::proto::McpTool {
        name: "process_file".to_string(),
        description: "Process a file with custom logic".to_string(),
        schema: r#"{"type":"object","properties":{"file_path":{"type":"string"}},"required":["file_path"]}"#
            .to_string(),
    };
    Ok(Response::new(ListToolsResponse { tools: vec![tool] }))
}
```

### Logging

The template uses `tracing` + `tracing-subscriber`. Set `RUST_LOG=debug` for verbose output:

```bash
RUST_LOG=debug cargo run
```

## Deployment

### Dockerfile

```dockerfile
FROM rust:1.78 AS builder
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/plugin /usr/local/bin/plugin
COPY --from=builder /app/plugin.json /app/plugin.json
WORKDIR /app
EXPOSE 50051
ENTRYPOINT ["plugin"]
```

### Plugin Manifest

See `template/plugin.json` for the complete manifest.

## Troubleshooting

1. **`protoc` not found** — install it via the system package manager:
   `apt install protobuf-compiler` / `brew install protobuf` / `choco install protoc`.
2. **`tonic-build` codegen fails** — pin `tonic-build` and `tonic` to matching versions
   (both 0.11 in this template).
3. **TLS issues at runtime** — for local dev use plain HTTP via `Server::builder()`;
   for production enable `tls` feature on `tonic`.

## Next Steps

- Study the [template code](template/) for a complete example
- Learn about [MCP integration](../../mcp-integration/exposing-tools.md) for AI agents
- Review the [plugin architecture](../../architecture.md)

## Resources

- [tonic (gRPC for Rust)](https://github.com/hyperium/tonic)
- [tokio runtime](https://tokio.rs/)
- [Protocol Buffers Rust tutorial](https://protobuf.dev/getting-started/rustcrate/)
