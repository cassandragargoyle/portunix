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
