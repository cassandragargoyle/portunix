use tonic::{Request, Response, Status};

use my_rust_plugin::config::PluginConfig;
use my_rust_plugin::proto::plugin_service_server::PluginService;
use my_rust_plugin::proto::{
    execute_response, health_check_response, shutdown_response, CallToolRequest, CallToolResponse,
    ExecuteRequest, ExecuteResponse, GetInfoRequest, GetInfoResponse, HealthCheckRequest,
    HealthCheckResponse, ListToolsRequest, ListToolsResponse, PluginInfo, ShutdownRequest,
    ShutdownResponse,
};
use my_rust_plugin::service::PluginBusinessService;

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
