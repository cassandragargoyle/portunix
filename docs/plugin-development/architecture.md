# Portunix Plugin Architecture

This document provides a comprehensive overview of the Portunix plugin system architecture, designed for both human developers and AI agents who need to understand how plugins integrate with the core system.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Portunix Core                           │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   Plugin        │  │      MCP        │  │     CLI         │  │
│  │   Manager       │  │     Server      │  │   Interface     │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │     gRPC        │  │   Configuration │  │    Security     │  │
│  │    Gateway      │  │    Manager      │  │    Manager      │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                        ┌───────┴───────┐
                        │ gRPC Protocol │
                        └───────┬───────┘
                                │
     ┌──────────────────────────┼──────────────────────────┐
     │                          │                          │
┌────▼────┐              ┌──────▼──────┐              ┌────▼────┐
│ Plugin  │              │   Plugin    │              │ Plugin  │
│   Go    │              │   Python    │              │  Java   │
│         │              │             │              │         │
│ gRPC    │              │    gRPC     │              │  gRPC   │
│ Server  │              │   Server    │              │ Server  │
└─────────┘              └─────────────┘              └─────────┘
```

## Core Components

### Plugin Manager
The Plugin Manager is responsible for:
- **Lifecycle Management**: Starting, stopping, enabling, disabling plugins
- **Discovery**: Finding and registering plugins
- **Health Monitoring**: Checking plugin status and performance
- **Dependency Resolution**: Managing plugin dependencies
- **Security Enforcement**: Applying permission policies

#### Key Interfaces
```go
type PluginManager interface {
    Install(pluginPath string) error
    Enable(pluginName string) error
    Disable(pluginName string) error
    Start(pluginName string) error
    Stop(pluginName string) error
    Health(pluginName string) (*HealthStatus, error)
    List() ([]*PluginInfo, error)
}
```

### gRPC Gateway
The gRPC Gateway handles communication between the core system and plugins:
- **Protocol Translation**: Converts internal calls to gRPC
- **Load Balancing**: Distributes requests across plugin instances
- **Circuit Breaking**: Prevents cascading failures
- **Metrics Collection**: Tracks performance and usage

### Configuration Manager
Manages plugin configuration and settings:
- **Schema Validation**: Ensures configuration correctness
- **Environment Variable Injection**: Provides runtime configuration
- **Secret Management**: Securely handles sensitive data
- **Dynamic Reconfiguration**: Updates configuration without restart

### Security Manager
Enforces security policies for plugins:
- **Permission System**: Controls what plugins can access
- **Sandboxing**: Isolates plugin execution
- **Authentication**: Verifies plugin identity
- **Audit Logging**: Tracks security-relevant events

## Plugin Types

### Service Plugins
Long-running background services that provide continuous functionality.

**Characteristics:**
- Run continuously once started
- Maintain persistent state
- Handle multiple concurrent requests
- Provide health check endpoints

**Example Use Cases:**
- System monitoring
- Background data processing
- File watchers
- Network services

### Tool Plugins
On-demand utilities that perform specific tasks and exit.

**Characteristics:**
- Short-lived execution
- Stateless (or minimal state)
- Single-purpose functionality
- Fast startup time

**Example Use Cases:**
- Code generators
- File converters
- Deployment scripts
- Data validators

### MCP Plugins
Plugins that expose functionality to AI agents through the Model Context Protocol.

**Characteristics:**
- Implement MCP tool interfaces
- Provide structured schemas
- AI-friendly descriptions
- Error handling for agent consumption

**Example Use Cases:**
- Code analysis tools
- Project scaffolding
- Documentation generators
- Development workflow automation

## Plugin Lifecycle

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  CREATED    │───▶│  INSTALLED  │───▶│   ENABLED   │───▶│   STARTED   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                            │                   │                   │
                            │                   │                   ▼
                            │                   │            ┌─────────────┐
                            │                   │            │   RUNNING   │
                            │                   │            └─────────────┘
                            │                   │                   │
                            │                   ▼                   ▼
                            │            ┌─────────────┐    ┌─────────────┐
                            │            │  DISABLED   │    │   STOPPED   │
                            │            └─────────────┘    └─────────────┘
                            │                   │                   │
                            ▼                   ▼                   ▼
                     ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
                     │ UNINSTALLED │    │ UNINSTALLED │    │ UNINSTALLED │
                     └─────────────┘    └─────────────┘    └─────────────┘
```

### State Transitions

1. **CREATED → INSTALLED**
   - Plugin files copied to plugin directory
   - Dependencies resolved and installed
   - Configuration schema validated

2. **INSTALLED → ENABLED**
   - Plugin registered with Plugin Manager
   - Permissions granted
   - Health check endpoint configured

3. **ENABLED → STARTED**
   - Plugin process launched
   - gRPC server started
   - Initial health check performed

4. **STARTED → RUNNING**
   - Plugin reports ready status
   - Services exposed through gRPC Gateway
   - Monitoring begins

## Communication Protocols

### gRPC Service Definition
All plugins must implement the base plugin service:

```protobuf
syntax = "proto3";

package portunix.plugin;

service PluginService {
    // Required methods
    rpc GetInfo(GetInfoRequest) returns (GetInfoResponse);
    rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
    rpc Shutdown(ShutdownRequest) returns (ShutdownResponse);
    
    // Optional methods (plugin-specific)
    rpc Execute(ExecuteRequest) returns (ExecuteResponse);
    rpc Configure(ConfigureRequest) returns (ConfigureResponse);
}

message PluginInfo {
    string name = 1;
    string version = 2;
    string description = 3;
    repeated string capabilities = 4;
}
```

### MCP Integration
Plugins that expose MCP tools implement additional interfaces:

```protobuf
service MCPService {
    rpc ListTools(ListToolsRequest) returns (ListToolsResponse);
    rpc CallTool(CallToolRequest) returns (CallToolResponse);
    rpc GetToolSchema(GetToolSchemaRequest) returns (GetToolSchemaResponse);
}
```

## Configuration System

### Plugin Manifest (plugin.yaml)
```yaml
# Basic metadata
name: my-plugin
version: 1.0.0
description: Example plugin for Portunix
author: Developer Name
license: MIT

# Plugin type and capabilities
type: service  # service, tool, mcp
capabilities:
  - system-monitoring
  - file-processing

# Runtime configuration
runtime:
  language: go
  main: ./main
  port: 50051
  health_check_port: 50052

# Dependencies
dependencies:
  system:
    - docker: ">=20.0.0"
    - git: ">=2.0.0"
  plugins:
    - database-connector: ">=1.2.0"

# Permissions
permissions:
  filesystem:
    read: ["/tmp", "/var/log"]
    write: ["/tmp/plugin-data"]
  network:
    outbound: ["api.example.com:443"]
  system:
    execute: ["docker", "git"]

# Configuration schema
configuration:
  required:
    api_key:
      type: string
      description: API key for external service
      sensitive: true
    endpoint:
      type: string
      description: Service endpoint URL
  optional:
    timeout:
      type: integer
      default: 30
      description: Request timeout in seconds

# MCP tools (for MCP plugins)
mcp:
  tools:
    - name: process_file
      description: Process a file with custom logic
      parameters:
        file_path:
          type: string
          description: Path to the file to process
        options:
          type: object
          description: Processing options
```

## Security Model

### Permission System
Plugins operate under a strict permission model:

1. **Filesystem Permissions**
   - Read access to specific directories
   - Write access to designated areas
   - No access to sensitive system files

2. **Network Permissions**
   - Outbound connections to specified hosts/ports
   - No inbound connections by default
   - Controlled access to local services

3. **System Permissions**
   - Execution of approved binaries
   - Access to environment variables
   - Limited process management

### Sandboxing
Plugins run in isolated environments:
- Separate process space
- Limited resource allocation (CPU, memory)
- Network isolation
- Filesystem chroot (where supported)

## Plugin Registry

### Local Registry
Each Portunix installation maintains a local plugin registry:

```
~/.portunix/plugins/
├── registry.json           # Plugin metadata database
├── installed/              # Installed plugin binaries
│   ├── plugin-name-v1.0.0/
│   └── another-plugin-v2.1.0/
├── enabled/                # Symlinks to enabled plugins
├── configs/                # Plugin configurations
└── logs/                   # Plugin logs
```

### Remote Registry
Future support for remote plugin registries:
- Plugin discovery and installation
- Version management
- Security verification
- Community ratings and reviews

## Error Handling and Monitoring

### Health Checks
All plugins must implement health check endpoints:
- **Liveness**: Plugin process is running
- **Readiness**: Plugin is ready to handle requests
- **Startup**: Plugin has completed initialization

### Logging
Structured logging with standard levels:
- **ERROR**: Critical errors requiring attention
- **WARN**: Warnings that don't affect functionality
- **INFO**: General operational messages
- **DEBUG**: Detailed debugging information

### Metrics
Plugins can expose metrics for monitoring:
- Request counts and latencies
- Resource usage (CPU, memory)
- Custom business metrics
- Error rates and types

## Development Guidelines

### Best Practices
1. **Stateless Design**: Prefer stateless plugins when possible
2. **Graceful Shutdown**: Handle shutdown signals properly
3. **Resource Management**: Clean up resources on exit
4. **Error Propagation**: Return meaningful error messages
5. **Configuration Validation**: Validate configuration at startup

### Anti-Patterns
1. **Direct Core Access**: Don't access core system internals
2. **Shared State**: Avoid shared mutable state between plugins
3. **Blocking Operations**: Don't block on long-running operations
4. **Resource Leaks**: Always clean up resources
5. **Security Bypassing**: Don't attempt to bypass security measures

## Integration Points

### CLI Integration
Plugins can extend the Portunix CLI:
```bash
portunix my-plugin command --arg value
```

### MCP Integration
Expose tools to AI agents:
```json
{
  "name": "generate_dockerfile",
  "description": "Generate optimized Dockerfile for any project",
  "inputSchema": {
    "type": "object",
    "properties": {
      "project_path": {"type": "string"},
      "language": {"type": "string"}
    }
  }
}
```

### API Integration
Plugins can expose REST APIs through the core system:
```yaml
# plugin.yaml
api:
  endpoints:
    - path: /api/v1/process
      method: POST
      handler: processData
```

## Future Enhancements

### Planned Features
1. **Plugin Marketplace**: Central repository for plugin discovery
2. **Hot Reloading**: Update plugins without restart
3. **Plugin Composition**: Combine multiple plugins into workflows
4. **GraphQL API**: Alternative to gRPC for web-based plugins
5. **WebAssembly Support**: Run plugins in WASM runtime

### Extension Points
1. **Custom Protocols**: Support for non-gRPC protocols
2. **Event System**: Plugin-to-plugin communication
3. **Workflow Engine**: Orchestrate plugin interactions
4. **Resource Quotas**: Fine-grained resource control

---

This architecture provides a flexible, secure, and scalable foundation for building Portunix plugins while maintaining clear separation of concerns and enabling rich integration with both human developers and AI agents.