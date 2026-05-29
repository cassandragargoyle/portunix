use anyhow::Result;
use clap::Parser;
use tonic::transport::Server;
use tracing_subscriber::EnvFilter;

use my_rust_plugin::config::PluginConfig;
use my_rust_plugin::proto::plugin_service_server::PluginServiceServer;

mod handlers;

use handlers::PluginServiceImpl;

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
        .with_env_filter(EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info")))
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
    let _ = tokio::signal::ctrl_c().await;
    tracing::info!("Shutting down gracefully...");
}
