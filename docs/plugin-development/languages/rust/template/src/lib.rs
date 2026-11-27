pub mod config;
pub mod handlers;
pub mod manifest;
pub mod service;

pub mod proto {
    tonic::include_proto!("portunix.plugins");
}
