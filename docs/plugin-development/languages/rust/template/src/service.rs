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
