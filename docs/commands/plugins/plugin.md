# Portunix Plugin Command

## Quick Start

The `plugin` command manages the complete lifecycle of Portunix plugins, from creation and installation to execution and monitoring.

### Simplest Usage
```bash
# List all plugins
portunix plugin list

# Install a plugin
portunix plugin install agile-software-development

# Enable and start a plugin
portunix plugin enable my-plugin
portunix plugin start my-plugin
```

### Basic Syntax
```bash
portunix plugin [subcommand] [plugin-name] [options]
```

### Available Subcommands
- `list` - Show installed plugins
- `install` - Install a plugin
- `uninstall` - Remove a plugin
- `enable` - Enable a plugin
- `disable` - Disable a plugin
- `start` - Start plugin service
- `stop` - Stop plugin service
- `health` - Check plugin health
- `info` - Show plugin details
- `create` - Create new plugin
- `validate` - Validate plugin structure

## Intermediate Usage

### Plugin Lifecycle Management

The complete plugin lifecycle:

```bash
# 1. Create or install
portunix plugin create my-plugin    # Create from template
portunix plugin install ./my-plugin # Install from directory

# 2. Enable (makes plugin available)
portunix plugin enable my-plugin

# 3. Start (runs plugin service)
portunix plugin start my-plugin

# 4. Check health
portunix plugin health my-plugin

# 5. Stop when not needed
portunix plugin stop my-plugin

# 6. Disable or uninstall
portunix plugin disable my-plugin
portunix plugin uninstall my-plugin
```

### Creating a New Plugin

Generate a new plugin from template:

```bash
portunix plugin create weather-tracker

# With options
portunix plugin create weather-tracker \
  --language go \
  --template advanced \
  --output ~/plugins/
```

This creates:
```
weather-tracker/
├── plugin.json          # Plugin manifest
├── proto/
│   └── plugin.proto     # gRPC definitions
├── src/
│   └── main.go         # Main plugin code
├── README.md
├── Makefile
└── .gitignore
```

### Plugin Installation

Install from various sources:

```bash
# From local directory
portunix plugin install ./my-plugin

# From GitHub
portunix plugin install github:username/plugin-repo

# From plugin registry
portunix plugin install agile-software-development

# From URL
portunix plugin install https://example.com/plugin.tar.gz

# With specific version
portunix plugin install my-plugin --version 2.0.0
```

### Plugin Configuration

Each plugin can have its configuration:

```bash
# View plugin configuration
portunix plugin config my-plugin

# Set configuration value
portunix plugin config my-plugin set api.key "secret-key"

# Edit configuration file
portunix plugin config my-plugin edit
```

### Health Monitoring

Check plugin health and status:

```bash
# Basic health check
portunix plugin health my-plugin

# Detailed health report
portunix plugin health my-plugin --detailed

# Health check all plugins
portunix plugin health --all
```

Output example:
```
Plugin: my-plugin
Status: Healthy
Uptime: 2h 15m
Memory: 45 MB
CPU: 0.2%
Last Error: None
Requests: 1,234
Response Time: 15ms avg
```

## Advanced Usage

### Plugin Architecture

Portunix uses a gRPC-based plugin architecture:

```
┌──────────────┐         gRPC          ┌─────────────┐
│   Portunix   │ ◄──────────────────► │   Plugin    │
│     Core     │                       │   Process   │
└──────────────┘                       └─────────────┘
       │                                      │
       │                                      │
       ▼                                      ▼
┌──────────────┐                       ┌─────────────┐
│   Commands   │                       │   Services  │
│   Registry   │                       │   Handler   │
└──────────────┘                       └─────────────┘
```

### Plugin Manifest Structure

The `plugin.json` file defines plugin metadata:

```json
{
  "name": "agile-software-development",
  "version": "1.0.0",
  "description": "Agile project management tools",
  "author": "Portunix Team",
  "license": "MIT",
  "runtime": "go",
  "executable": "bin/agile-plugin",
  "dependencies": {
    "portunix": ">=1.5.0",
    "runtime": {
      "go": ">=1.20"
    }
  },
  "commands": [
    {
      "name": "agile",
      "description": "Agile project management",
      "subcommands": ["project", "board", "task"]
    }
  ],
  "services": {
    "grpc": {
      "port": 50051,
      "proto": "proto/plugin.proto"
    }
  },
  "permissions": [
    "filesystem:read",
    "filesystem:write",
    "network:http",
    "database:sqlite"
  ],
  "configuration": {
    "schema": "config/schema.json",
    "defaults": "config/defaults.json"
  },
  "hooks": {
    "postInstall": "scripts/post-install.sh",
    "preUninstall": "scripts/pre-uninstall.sh",
    "health": "scripts/health-check.sh"
  }
}
```

### gRPC Protocol Definition

Plugins communicate via Protocol Buffers:

```protobuf
syntax = "proto3";

package portunix.plugin;

service Plugin {
  // Lifecycle management
  rpc Initialize(InitRequest) returns (InitResponse);
  rpc Execute(ExecuteRequest) returns (ExecuteResponse);
  rpc Shutdown(ShutdownRequest) returns (ShutdownResponse);

  // Health monitoring
  rpc HealthCheck(HealthRequest) returns (HealthResponse);

  // Configuration
  rpc GetConfig(ConfigRequest) returns (ConfigResponse);
  rpc SetConfig(ConfigUpdateRequest) returns (ConfigUpdateResponse);
}

message ExecuteRequest {
  string command = 1;
  repeated string args = 2;
  map<string, string> env = 3;
  bytes stdin = 4;
}

message ExecuteResponse {
  int32 exit_code = 1;
  bytes stdout = 2;
  bytes stderr = 3;
  map<string, string> metadata = 4;
}
```

### Plugin Development

#### Go Plugin Template

```go
package main

import (
    "context"
    "log"
    "net"

    "google.golang.org/grpc"
    pb "github.com/portunix/plugin-sdk/proto"
)

type pluginServer struct {
    pb.UnimplementedPluginServer
}

func (s *pluginServer) Initialize(ctx context.Context,
    req *pb.InitRequest) (*pb.InitResponse, error) {
    // Initialize plugin
    return &pb.InitResponse{
        Status: "ready",
        Version: "1.0.0",
    }, nil
}

func (s *pluginServer) Execute(ctx context.Context,
    req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
    // Handle command execution
    switch req.Command {
    case "hello":
        return &pb.ExecuteResponse{
            ExitCode: 0,
            Stdout: []byte("Hello from plugin!"),
        }, nil
    default:
        return &pb.ExecuteResponse{
            ExitCode: 1,
            Stderr: []byte("Unknown command"),
        }, nil
    }
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    s := grpc.NewServer()
    pb.RegisterPluginServer(s, &pluginServer{})

    log.Printf("Plugin server starting on %v", lis.Addr())
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
```

### Plugin Registry

The centralized plugin registry structure:

```json
{
  "plugins": [
    {
      "name": "agile-software-development",
      "description": "Agile project management tools",
      "versions": [
        {
          "version": "1.0.0",
          "url": "https://github.com/portunix/plugins/releases/agile-1.0.0.tar.gz",
          "sha256": "abc123...",
          "compatibility": {
            "portunix": ">=1.5.0",
            "platforms": ["windows", "linux", "darwin"]
          }
        }
      ],
      "tags": ["productivity", "project-management", "agile"],
      "author": "Portunix Team",
      "homepage": "https://github.com/portunix/agile-plugin"
    }
  ]
}
```

### Plugin Permissions

Fine-grained permission system:

```bash
# View plugin permissions
portunix plugin permissions my-plugin

# Grant additional permissions
portunix plugin grant my-plugin network:https database:postgres

# Revoke permissions
portunix plugin revoke my-plugin filesystem:write

# Run with elevated permissions (requires confirmation)
portunix plugin run my-plugin --elevated
```

Permission categories:
- `filesystem:read` - Read files
- `filesystem:write` - Write files
- `network:http` - HTTP requests
- `network:https` - HTTPS requests
- `database:*` - Database access
- `system:exec` - Execute commands
- `system:env` - Environment variables
- `docker:*` - Docker operations

### Plugin Communication

Inter-plugin communication:

```bash
# Enable plugin messaging
portunix plugin message-bus enable

# Send message to plugin
portunix plugin send my-plugin "{'action': 'refresh'}"

# Subscribe to plugin events
portunix plugin subscribe weather-tracker temperature-update
```

### Plugin Debugging

Debug plugin issues:

```bash
# Run plugin in debug mode
portunix plugin debug my-plugin

# View plugin logs
portunix plugin logs my-plugin

# Tail logs in real-time
portunix plugin logs my-plugin -f

# Set log level
portunix plugin loglevel my-plugin debug

# Profile plugin performance
portunix plugin profile my-plugin --cpu --memory
```

### Plugin Testing

Test plugin functionality:

```bash
# Run plugin tests
portunix plugin test my-plugin

# Run specific test suite
portunix plugin test my-plugin --suite integration

# Generate test coverage
portunix plugin test my-plugin --coverage

# Benchmark plugin
portunix plugin benchmark my-plugin
```

## Expert Tips & Tricks

### 1. Plugin Development Workflow

```bash
# Development cycle
portunix plugin create my-plugin
cd my-plugin
make build
portunix plugin install . --dev
portunix plugin debug my-plugin
make test
portunix plugin validate .
```

### 2. Plugin Hot Reload

```bash
# Enable hot reload for development
portunix plugin develop my-plugin --hot-reload

# Watch for changes
portunix plugin watch my-plugin
```

### 3. Plugin Bundling

```bash
# Bundle plugin with dependencies
portunix plugin bundle my-plugin --output my-plugin.bundle

# Create standalone installer
portunix plugin package my-plugin --installer
```

### 4. Plugin Migration

```bash
# Migrate plugin data to new version
portunix plugin migrate my-plugin --from 1.0.0 --to 2.0.0

# Backup plugin data
portunix plugin backup my-plugin

# Restore plugin data
portunix plugin restore my-plugin --backup 2024-01-15.tar.gz
```

### 5. Plugin Distribution

```bash
# Publish to registry
portunix plugin publish my-plugin --registry official

# Sign plugin for distribution
portunix plugin sign my-plugin --key ~/.gnupg/plugin-key

# Create release
portunix plugin release my-plugin --version 1.0.0 --notes "Initial release"
```

## Troubleshooting

### Common Issues and Solutions

#### 1. Plugin Won't Start
```bash
# Check plugin status
portunix plugin status my-plugin

# View error logs
portunix plugin logs my-plugin --errors

# Verify dependencies
portunix plugin deps my-plugin --verify

# Reinstall plugin
portunix plugin reinstall my-plugin
```

#### 2. Permission Denied
```bash
# Check current permissions
portunix plugin permissions my-plugin

# Run with required permissions
portunix plugin run my-plugin --permissions filesystem:read,network:https
```

#### 3. Plugin Crashes
```bash
# Get crash report
portunix plugin crashlog my-plugin

# Run with memory limit
portunix plugin start my-plugin --memory-limit 100M

# Enable crash recovery
portunix plugin config my-plugin set recovery.enabled true
```

#### 4. Version Conflicts
```bash
# Check compatibility
portunix plugin compat my-plugin

# Force specific version
portunix plugin install my-plugin --version 1.0.0 --force

# Downgrade plugin
portunix plugin downgrade my-plugin 1.0.0
```

### Debug Mode

```bash
# Verbose output
portunix plugin -v install my-plugin

# Debug logging
portunix plugin --debug start my-plugin

# Trace execution
portunix plugin --trace execute my-plugin command
```

## Plugin Security

### Code Signing

```bash
# Verify plugin signature
portunix plugin verify my-plugin

# Import trusted keys
portunix plugin trust-key https://example.com/plugin.key

# Set signature requirement
portunix config set plugin.requireSignature true
```

### Sandboxing

```bash
# Run plugin in sandbox
portunix plugin sandbox my-plugin

# Configure sandbox restrictions
portunix plugin sandbox-config my-plugin \
  --no-network \
  --readonly-filesystem \
  --memory-limit 50M
```

### Audit Trail

```bash
# View plugin audit log
portunix plugin audit my-plugin

# Export audit report
portunix plugin audit --export report.json

# Monitor plugin activity
portunix plugin monitor my-plugin --real-time
```

## Performance Optimization

### Resource Management

```bash
# Set resource limits
portunix plugin limits my-plugin \
  --cpu 0.5 \
  --memory 100M \
  --disk-io 10M/s

# Auto-scaling configuration
portunix plugin autoscale my-plugin \
  --min-instances 1 \
  --max-instances 5 \
  --cpu-threshold 80
```

### Caching

```bash
# Enable plugin cache
portunix plugin cache my-plugin enable

# Clear plugin cache
portunix plugin cache my-plugin clear

# Configure cache size
portunix plugin cache my-plugin --size 100M
```

## Integration Examples

### CI/CD Integration

```yaml
# GitHub Actions
- name: Test Plugin
  run: |
    portunix plugin install ./my-plugin
    portunix plugin test my-plugin
    portunix plugin validate my-plugin
```

### Docker Integration

```dockerfile
# Dockerfile for plugin
FROM portunix:latest
COPY my-plugin /plugins/
RUN portunix plugin install /plugins/my-plugin
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: portunix-plugin
spec:
  template:
    spec:
      containers:
      - name: plugin
        image: portunix:latest
        command: ["portunix", "plugin", "serve", "my-plugin"]
```

## API Integration

### REST API

```bash
# Plugin management via API
curl -X POST http://localhost:8080/api/plugin/install \
  -H "Content-Type: application/json" \
  -d '{"name": "my-plugin", "version": "1.0.0"}'
```

### WebSocket Events

```javascript
// Subscribe to plugin events
const ws = new WebSocket('ws://localhost:8080/plugin-events');
ws.on('message', (data) => {
  const event = JSON.parse(data);
  console.log(`Plugin ${event.plugin}: ${event.status}`);
});
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORTUNIX_PLUGIN_DIR` | Plugin installation directory | ~/.portunix/plugins |
| `PORTUNIX_PLUGIN_REGISTRY` | Plugin registry URL | https://registry.portunix.ai |
| `PORTUNIX_PLUGIN_TIMEOUT` | Plugin operation timeout | 30s |
| `PORTUNIX_PLUGIN_DEBUG` | Enable debug mode | false |
| `PORTUNIX_PLUGIN_SANDBOX` | Enable sandboxing | true |

## Related Commands

- [`install`](install.md) - Install packages
- [`update`](update.md) - Update Portunix
- [`config`](config.md) - Configuration management
- [`mcp`](mcp.md) - MCP server integration

## Command Reference

### Complete Parameter List

| Subcommand | Parameters | Description |
|------------|------------|-------------|
| `list` | `--all`, `--enabled`, `--disabled` | List plugins |
| `install` | `--version`, `--force`, `--dev` | Install plugin |
| `uninstall` | `--keep-data`, `--force` | Remove plugin |
| `enable` | `--auto-start` | Enable plugin |
| `disable` | `--force` | Disable plugin |
| `start` | `--debug`, `--profile` | Start plugin |
| `stop` | `--force`, `--timeout` | Stop plugin |
| `health` | `--detailed`, `--all` | Health check |
| `info` | `--json`, `--yaml` | Plugin information |
| `create` | `--template`, `--language` | Create plugin |
| `validate` | `--strict`, `--fix` | Validate plugin |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Plugin not found |
| 3 | Plugin already exists |
| 4 | Permission denied |
| 5 | Dependency error |
| 6 | Version conflict |
| 7 | Validation failed |
| 8 | Communication error |
| 9 | Timeout |
| 10 | Resource limit exceeded |

## Version History

- **v1.5.0** - Added plugin sandboxing
- **v1.4.0** - Implemented plugin registry
- **v1.3.0** - Added inter-plugin communication
- **v1.2.0** - Introduced health monitoring
- **v1.1.0** - Added plugin templates
- **v1.0.0** - Initial plugin system