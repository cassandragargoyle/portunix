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
