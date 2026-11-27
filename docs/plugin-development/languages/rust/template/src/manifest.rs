//! plugin.json, the single source of the plugin information and commands
//! reported to Portunix over gRPC

use anyhow::{bail, Context, Result};
use serde::Deserialize;
use std::fs;
use std::path::Path;

use crate::proto::{
    PluginCapabilities, PluginCommand, PluginInfo, PluginParameter, PluginPermissions,
};

#[derive(Debug, Clone, Default, Deserialize)]
#[serde(default)]
pub struct PluginManifest {
    pub name: String,
    pub version: String,
    pub description: String,
    pub author: String,
    pub license: String,
    pub commands: Vec<Command>,
    pub dependencies: Dependencies,
    pub ai_integration: AiIntegration,
    pub permissions: Permissions,
}

#[derive(Debug, Clone, Default, Deserialize)]
#[serde(default)]
pub struct Command {
    pub name: String,
    pub description: String,
    pub subcommands: Vec<String>,
    pub parameters: Vec<Parameter>,
    pub examples: Vec<String>,
}

#[derive(Debug, Clone, Default, Deserialize)]
#[serde(default)]
pub struct Parameter {
    pub name: String,
    #[serde(rename = "type")]
    pub kind: String,
    pub description: String,
    pub required: bool,
    pub default_value: String,
}

#[derive(Debug, Clone, Default, Deserialize)]
#[serde(default)]
pub struct Dependencies {
    pub os_support: Vec<String>,
}

#[derive(Debug, Clone, Default, Deserialize)]
#[serde(default)]
pub struct AiIntegration {
    pub mcp_tools: Vec<McpTool>,
}

#[derive(Debug, Clone, Default, Deserialize)]
#[serde(default)]
pub struct McpTool {
    pub name: String,
}

#[derive(Debug, Clone, Default, Deserialize)]
#[serde(default)]
pub struct Permissions {
    pub filesystem: Vec<String>,
    pub network: Vec<String>,
    pub database: Vec<String>,
    pub system: Vec<String>,
    pub level: String,
}

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

    pub fn plugin_info(&self) -> PluginInfo {
        let p = &self.permissions;
        PluginInfo {
            name: self.name.clone(),
            version: self.version.clone(),
            description: self.description.clone(),
            author: self.author.clone(),
            license: self.license.clone(),
            supported_os: self.dependencies.os_support.clone(),
            commands: self.plugin_commands(),
            capabilities: Some(self.capabilities()),
            required_permissions: Some(PluginPermissions {
                filesystem: p.filesystem.clone(),
                network: p.network.clone(),
                database: p.database.clone(),
                system: p.system.clone(),
                level: if p.level.is_empty() {
                    "limited".to_string()
                } else {
                    p.level.clone()
                },
            }),
        }
    }

    pub fn plugin_commands(&self) -> Vec<PluginCommand> {
        self.commands
            .iter()
            .map(|c| PluginCommand {
                name: c.name.clone(),
                description: c.description.clone(),
                subcommands: c.subcommands.clone(),
                parameters: c
                    .parameters
                    .iter()
                    .map(|p| PluginParameter {
                        name: p.name.clone(),
                        r#type: if p.kind.is_empty() {
                            "string".to_string()
                        } else {
                            p.kind.clone()
                        },
                        description: p.description.clone(),
                        required: p.required,
                        default_value: p.default_value.clone(),
                    })
                    .collect(),
                examples: c.examples.clone(),
            })
            .collect()
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
}
