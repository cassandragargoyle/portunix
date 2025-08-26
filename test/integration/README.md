# Plugin Installation Integration Tests

This directory contains integration tests for the Portunix plugin system.

## Test Coverage

### Plugin Installation Tests (`plugin_installation_test.go`)

1. **Plugin Build** - Building the test plugin from source
2. **Plugin Installation** - Installing a plugin from directory
3. **Plugin Listing** - Viewing installed plugins
4. **Plugin Information** - Getting detailed plugin metadata
5. **Plugin Lifecycle**:
   - Enable/Disable plugin
   - Start/Stop plugin service
   - Health check
6. **Plugin Uninstallation** - Removing installed plugins
7. **Plugin Validation** - Validating plugin configuration
8. **Plugin Creation** - Creating new plugin from template

### PowerShell Installation Tests (`issue_012_powershell_installation_test.go`)

1. **PowerShell Installation** - Installing PowerShell on Linux using quick variant
2. **Command Verification** - Testing PowerShell command availability and execution
3. **Version Detection** - Verifying PowerShell version output
4. **Preset Integration** - Testing PowerShell availability in installation system

## Running Tests

### Run all plugin tests
```bash
go test -v ./test/integration -run TestPlugin
```

### Run all PowerShell tests
```bash
go test -v ./test/integration -run TestPowerShell
```

### Run specific tests
```bash
# Test plugin installation workflow
go test -v ./test/integration -run TestPluginInstallation

# Test plugin validation
go test -v ./test/integration -run TestPluginValidation

# Test plugin creation
go test -v ./test/integration -run TestPluginCreate

# Test PowerShell installation
go test -v ./test/integration -run TestPowerShellInstallation

# Test PowerShell preset integration
go test -v ./test/integration -run TestPowerShellPresetsIntegration
```

### Run with timeout
```bash
go test -v ./test/integration -run TestPlugin -timeout 30s
```

## Test Plugin

The tests use a minimal test plugin located at `test/test-plugin/` which includes:
- `plugin.yaml` - Plugin configuration
- `main.go` - Basic gRPC server implementation  
- `README.md` - Plugin documentation

## Expected Behavior

Some tests may show warnings for:
- Plugin enable/disable state issues - This is expected for basic plugin implementation
- Health check failures - Test plugin doesn't implement health endpoint
- Start/stop issues - Plugin lifecycle management is still being developed

These are marked as warnings rather than failures since the core installation/uninstallation functionality works correctly.

## Prerequisites

- Go 1.21 or later
- Portunix binary built (`go build -o .` in project root)
- Linux or macOS (plugin tests not supported on Windows yet)

## Test Structure

```go
TestPluginInstallation
├── BuildTestPlugin      // Compile test plugin
├── InstallPlugin        // Install from directory
├── ListPlugins          // List installed plugins
├── PluginInfo           // Show plugin details
├── PluginLifecycle      // Test enable/start/stop/disable
│   ├── Enable
│   ├── Start
│   ├── Health
│   ├── Stop
│   └── Disable
└── UninstallPlugin      // Remove plugin

TestPluginValidation     // Validate plugin configuration
TestPluginCreate         // Create new plugin template
```

## Troubleshooting

If tests fail:

1. **Build failures**: Ensure Go is installed and project builds successfully
2. **Plugin not found**: Check that test-plugin directory exists
3. **Permission errors**: Some operations may require appropriate permissions
4. **State issues**: Plugin system may have persistent state - try cleaning plugin directory

## Future Improvements

- Add Windows support for plugin tests
- Test plugin communication via gRPC
- Test plugin with actual functionality
- Test plugin dependencies and version compatibility
- Test concurrent plugin operations
- Add benchmarks for plugin operations