//! Implementation of portunix.plugins.PluginService; plugin info and commands
//! come from plugin.json, so the manifest stays the single source

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
