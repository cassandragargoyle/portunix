use my_rust_plugin::config::PluginConfig;
use my_rust_plugin::service::PluginBusinessService;

#[test]
fn hello_returns_default_greeting() {
    let service = PluginBusinessService::new(PluginConfig::default());
    assert_eq!(service.execute("hello", &[]).unwrap(), "Hello, World!");
}

#[test]
fn hello_returns_custom_greeting() {
    let service = PluginBusinessService::new(PluginConfig::default());
    let args = vec!["Portunix".to_string()];
    assert_eq!(service.execute("hello", &args).unwrap(), "Hello, Portunix!");
}

#[test]
fn unknown_command_errors() {
    let service = PluginBusinessService::new(PluginConfig::default());
    assert!(service.execute("unknown", &[]).is_err());
}

#[test]
fn process_requires_arguments() {
    let service = PluginBusinessService::new(PluginConfig::default());
    assert!(service.execute("process", &[]).is_err());
}

mod grpc {
    use my_rust_plugin::config::PluginConfig;
    use my_rust_plugin::handlers::PluginServiceImpl;
    use my_rust_plugin::manifest::PluginManifest;
    use my_rust_plugin::proto::plugin_service_server::PluginService;
    use my_rust_plugin::proto::{
        ExecuteRequest, HealthRequest, InitializeRequest, ListCommandsRequest, ShutdownRequest,
    };
    use tonic::Request;

    // Cargo runs tests from the template root, where plugin.json is
    fn handler() -> PluginServiceImpl {
        let manifest = PluginManifest::load("plugin.json").expect("load plugin.json");
        PluginServiceImpl::new(PluginConfig::default(), manifest)
    }

    #[tokio::test]
    async fn initialize_returns_manifest_info() {
        let resp = handler()
            .initialize(Request::new(InitializeRequest::default()))
            .await
            .unwrap()
            .into_inner();
        assert!(resp.success);
        let info = resp.plugin_info.expect("plugin info");
        assert_eq!(info.name, "rust-plugin-template");
        assert_eq!(info.commands[0].name, "hello");
    }

    #[tokio::test]
    async fn health_is_healthy() {
        let resp = handler()
            .health(Request::new(HealthRequest {}))
            .await
            .unwrap()
            .into_inner();
        assert!(resp.healthy);
    }

    #[tokio::test]
    async fn execute_reports_output_and_errors() {
        let h = handler();
        let ok = h
            .execute(Request::new(ExecuteRequest {
                command: "hello".to_string(),
                args: vec!["Portunix".to_string()],
                ..Default::default()
            }))
            .await
            .unwrap()
            .into_inner();
        assert!(ok.success);
        assert_eq!(ok.output, "Hello, Portunix!");

        let failed = h
            .execute(Request::new(ExecuteRequest {
                command: "unknown".to_string(),
                ..Default::default()
            }))
            .await
            .unwrap()
            .into_inner();
        assert!(!failed.success);
        assert_ne!(failed.exit_code, 0);
    }

    #[tokio::test]
    async fn list_commands_and_shutdown() {
        let h = handler();
        let list = h
            .list_commands(Request::new(ListCommandsRequest {}))
            .await
            .unwrap()
            .into_inner();
        assert!(!list.commands.is_empty());

        let resp = h
            .shutdown(Request::new(ShutdownRequest::default()))
            .await
            .unwrap()
            .into_inner();
        assert!(resp.success);
    }
}
