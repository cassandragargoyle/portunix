pub mod config;
pub mod service;

pub mod proto {
    tonic::include_proto!("rust_plugin_template");
}
