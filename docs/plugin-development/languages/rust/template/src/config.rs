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

fn default_name() -> String {
    "my-rust-plugin".to_string()
}

fn default_version() -> String {
    "1.0.0".to_string()
}

fn default_log_level() -> String {
    "info".to_string()
}

fn default_port() -> u16 {
    50051
}

fn default_health_port() -> u16 {
    50052
}

impl PluginConfig {
    pub fn load<P: AsRef<Path>>(path: P) -> Result<Self> {
        let path = path.as_ref();
        if !path.exists() {
            return Ok(Self::default());
        }
        let raw = fs::read_to_string(path)
            .with_context(|| format!("reading config {}", path.display()))?;
        serde_yaml::from_str(&raw)
            .with_context(|| format!("parsing config {}", path.display()))
    }
}
