# Plugin Commands

Complete plugin ecosystem with gRPC-based architecture, lifecycle management, and development tools.

## Commands in this Category

### [`plugin`](plugin.md) - Plugin Lifecycle Management
Comprehensive plugin system supporting creation, development, distribution, and runtime management.

**Quick Examples:**
```bash
portunix plugin list                      # List installed plugins
portunix plugin install agile             # Install plugin from registry
portunix plugin create my-plugin          # Create new plugin from template
portunix plugin health my-plugin          # Check plugin health
```

**Key Features:**
- Complete plugin lifecycle management (create, install, enable, start, stop, uninstall)
- gRPC-based communication for high performance
- Plugin templates for rapid development
- Health monitoring and diagnostics
- Permission system with fine-grained controls
- Plugin registry integration and distribution

**Common Use Cases:**
- Extending Portunix functionality
- Custom workflow automation
- Domain-specific tool integration
- Team-specific command creation
- Third-party service integration

## Plugin Architecture

### gRPC-Based Communication

Portunix uses a modern gRPC architecture for plugin communication:

```
┌─────────────────┐         gRPC          ┌─────────────────┐
│   Portunix      │ ◄─────────────────► │   Plugin        │
│   Core Engine   │   (High Performance) │   Process       │
└─────────────────┘                       └─────────────────┘
       │                                          │
       │                                          │
       ▼                                          ▼
┌─────────────────┐                       ┌─────────────────┐
│ Command Router  │                       │ Service Handler │
│ Plugin Registry │                       │ Tool Provider   │
│ Health Monitor  │                       │ Event Listener  │
└─────────────────┘                       └─────────────────┘
```

### Plugin Development Workflow

```bash
# 1. Create plugin from template
portunix plugin create weather-tracker \
  --language go \
  --template advanced \
  --author "Your Name"

# 2. Develop plugin functionality
cd weather-tracker
# Edit src/main.go, proto/plugin.proto

# 3. Build and test locally
make build
make test

# 4. Install for development
portunix plugin install . --dev

# 5. Test plugin functionality
portunix plugin health weather-tracker
portunix weather current --location "New York"

# 6. Package for distribution
portunix plugin package weather-tracker

# 7. Publish to registry
portunix plugin publish weather-tracker --registry official
```

## Plugin Categories

### Official Plugins

#### Agile Software Development Plugin
```bash
# Install agile development tools
portunix plugin install agile-software-development

# Use agile commands
portunix agile project create "WebApp Development"
portunix agile board show
portunix agile task add "Implement authentication"
```

**Features:**
- Kanban board management
- User story tracking
- Sprint planning tools
- Flow metrics analysis
- Team collaboration features

### Community Plugins

#### Development Tools
- **Code Quality Plugin** - Linting, formatting, security scanning
- **Database Plugin** - Database management and migration tools
- **Cloud Plugin** - AWS, Azure, GCP integration
- **Monitoring Plugin** - Application and infrastructure monitoring

#### Productivity Tools
- **Notes Plugin** - Technical documentation and note-taking
- **Time Tracking Plugin** - Development time tracking and reporting
- **Screenshot Plugin** - Automated screenshot and documentation tools

### Enterprise Plugins
- **Security Plugin** - Enterprise security scanning and compliance
- **Audit Plugin** - Change tracking and compliance reporting
- **Integration Plugin** - Enterprise system integration

## Plugin Development

### Creating Your First Plugin

```bash
# Generate plugin template
portunix plugin create hello-world --language go

# Generated structure:
hello-world/
├── plugin.json          # Plugin manifest
├── proto/
│   └── plugin.proto     # gRPC service definition
├── src/
│   └── main.go         # Main plugin implementation
├── cmd/
│   └── hello.go        # Command implementations
├── config/
│   └── defaults.json   # Default configuration
├── scripts/
│   ├── build.sh        # Build script
│   └── test.sh         # Test script
├── README.md
├── Makefile
└── .gitignore
```

### Plugin Manifest (plugin.json)

```json
{
  "name": "hello-world",
  "version": "1.0.0",
  "description": "A simple hello world plugin",
  "author": "Your Name <your.email@example.com>",
  "license": "MIT",
  "homepage": "https://github.com/username/hello-world-plugin",

  "runtime": {
    "language": "go",
    "version": ">=1.20"
  },

  "portunix": {
    "min_version": "1.5.0",
    "max_version": "2.0.0"
  },

  "commands": [
    {
      "name": "hello",
      "description": "Greet the user",
      "usage": "portunix hello [name]",
      "subcommands": ["world", "user", "team"]
    }
  ],

  "services": {
    "grpc": {
      "port": 50051,
      "proto": "proto/plugin.proto"
    },
    "health": {
      "endpoint": "/health",
      "interval": "30s"
    }
  },

  "permissions": [
    "filesystem:read",
    "network:http",
    "system:env"
  ],

  "dependencies": {
    "system": ["curl", "git"],
    "portunix": ["install", "system"]
  },

  "configuration": {
    "schema": "config/schema.json",
    "defaults": "config/defaults.json"
  },

  "hooks": {
    "postInstall": "scripts/post-install.sh",
    "preUninstall": "scripts/pre-uninstall.sh",
    "healthCheck": "scripts/health.sh"
  }
}
```

### gRPC Service Implementation

```go
// src/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "net"

    "google.golang.org/grpc"
    pb "github.com/username/hello-world-plugin/proto"
)

type pluginServer struct {
    pb.UnimplementedPluginServer
}

func (s *pluginServer) Initialize(ctx context.Context,
    req *pb.InitRequest) (*pb.InitResponse, error) {

    log.Printf("Initializing plugin: %s", req.GetConfig())

    return &pb.InitResponse{
        Status:  "ready",
        Version: "1.0.0",
        Message: "Hello World plugin initialized successfully",
    }, nil
}

func (s *pluginServer) Execute(ctx context.Context,
    req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {

    command := req.GetCommand()
    args := req.GetArgs()

    switch command {
    case "hello":
        return s.handleHello(args)
    case "world":
        return s.handleWorld(args)
    default:
        return &pb.ExecuteResponse{
            ExitCode: 1,
            Stderr:   []byte(fmt.Sprintf("Unknown command: %s", command)),
        }, nil
    }
}

func (s *pluginServer) handleHello(args []string) (*pb.ExecuteResponse, error) {
    name := "World"
    if len(args) > 0 {
        name = args[0]
    }

    message := fmt.Sprintf("Hello, %s! 👋\n", name)

    return &pb.ExecuteResponse{
        ExitCode: 0,
        Stdout:   []byte(message),
        Metadata: map[string]string{
            "greeting": name,
            "timestamp": fmt.Sprintf("%d", time.Now().Unix()),
        },
    }, nil
}

func (s *pluginServer) HealthCheck(ctx context.Context,
    req *pb.HealthRequest) (*pb.HealthResponse, error) {

    return &pb.HealthResponse{
        Status: "healthy",
        Uptime: time.Since(startTime).String(),
        Metrics: map[string]string{
            "requests": fmt.Sprintf("%d", requestCount),
            "memory": getMemoryUsage(),
        },
    }, nil
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

## Advanced Plugin Features

### Permission System

```bash
# View plugin permissions
portunix plugin permissions my-plugin

# Grant additional permissions
portunix plugin grant my-plugin \
  --permissions network:https,database:postgresql

# Revoke permissions
portunix plugin revoke my-plugin \
  --permissions filesystem:write

# Run with elevated permissions (requires confirmation)
portunix plugin run my-plugin --elevated
```

Permission categories:
- `filesystem:read/write` - File system access
- `network:http/https` - Network access
- `database:*` - Database connections
- `system:exec/env` - System command execution
- `docker:*` - Container operations
- `portunix:*` - Portunix core operations

### Inter-Plugin Communication

```bash
# Enable plugin messaging bus
portunix plugin message-bus enable

# Send message between plugins
portunix plugin send weather-tracker \
  --message '{"action": "refresh", "location": "NYC"}'

# Subscribe to plugin events
portunix plugin subscribe agile-plugin task-completed
```

### Plugin Configuration Management

```bash
# View plugin configuration
portunix plugin config my-plugin

# Set configuration values
portunix plugin config my-plugin set \
  --key api.endpoint \
  --value "https://api.example.com"

# Edit configuration interactively
portunix plugin config my-plugin edit

# Validate configuration
portunix plugin config my-plugin validate
```

### Plugin Testing and Debugging

```bash
# Run plugin test suite
portunix plugin test my-plugin

# Debug plugin execution
portunix plugin debug my-plugin \
  --attach-debugger \
  --log-level debug

# Profile plugin performance
portunix plugin profile my-plugin \
  --cpu-profile \
  --memory-profile \
  --duration 60s

# Trace plugin API calls
portunix plugin trace my-plugin --api-calls
```

## Plugin Distribution

### Plugin Registry

```bash
# Search for plugins
portunix plugin search weather

# Show plugin information
portunix plugin info agile-software-development

# Install from registry
portunix plugin install weather-tracker --version 2.1.0

# Update plugins
portunix plugin update --all

# List available updates
portunix plugin outdated
```

### Publishing Plugins

```bash
# Login to registry
portunix plugin registry login

# Validate before publishing
portunix plugin validate my-plugin

# Publish to registry
portunix plugin publish my-plugin \
  --registry official \
  --tag stable

# Create release
portunix plugin release my-plugin \
  --version 1.2.0 \
  --notes "Added new features and bug fixes"
```

### Private Plugin Distribution

```bash
# Setup private registry
portunix plugin registry add company \
  --url https://plugins.company.com \
  --auth-token $COMPANY_TOKEN

# Install from private registry
portunix plugin install company/internal-tools

# Share plugins via file
portunix plugin bundle my-plugin --output my-plugin.bundle
portunix plugin install my-plugin.bundle
```

## Integration with Portunix Ecosystem

### With Core Commands

```bash
# Plugin that extends install command
portunix plugin install package-manager-plugin

# Use enhanced installation
portunix install nodejs --plugin package-manager --optimization
```

### With Container Integration

```bash
# Plugin with container support
portunix plugin install container-dev-plugin

# Use plugin in containers
portunix docker run ubuntu --enable-plugins
portunix docker exec my-container portunix dev-tools setup
```

### With MCP Integration

```bash
# Plugin as MCP tool
portunix plugin install mcp-tools-plugin
portunix mcp serve --include-plugins

# AI assistant can now use plugin commands
```

### With Virtualization

```bash
# Install plugins in VMs
portunix virt ssh my-vm "portunix plugin install agile"

# VM-specific plugin operations
portunix plugin install vm-manager-plugin
portunix vm-tools snapshot-all --schedule daily
```

## Plugin Security

### Code Signing and Verification

```bash
# Sign plugin for distribution
portunix plugin sign my-plugin --key ~/.gnupg/plugin-signing-key

# Verify plugin signature
portunix plugin verify my-plugin

# Install only signed plugins
portunix config set plugin.requireSignature true
```

### Sandboxing and Isolation

```bash
# Run plugin in sandbox
portunix plugin sandbox my-plugin

# Configure sandbox restrictions
portunix plugin sandbox-config my-plugin \
  --no-network \
  --readonly-filesystem \
  --memory-limit 100M

# Audit plugin behavior
portunix plugin audit my-plugin --real-time
```

### Security Best Practices

- Always verify plugin signatures
- Review plugin permissions before installation
- Use least-privilege principle
- Regular security audits
- Monitor plugin behavior

## Performance Optimization

### Plugin Performance Tuning

```bash
# Optimize plugin startup
portunix plugin optimize my-plugin --startup-time

# Configure resource limits
portunix plugin limits my-plugin \
  --cpu 0.5 \
  --memory 256M \
  --disk-io 10M/s

# Enable plugin caching
portunix plugin cache my-plugin enable --size 50M
```

### Monitoring and Metrics

```bash
# Real-time plugin monitoring
portunix plugin monitor my-plugin

# Export plugin metrics
portunix plugin metrics my-plugin --export prometheus

# Performance analysis
portunix plugin analyze my-plugin --performance
```

## Troubleshooting

### Common Plugin Issues

```bash
# Plugin won't start
portunix plugin diagnose my-plugin --startup

# Communication errors
portunix plugin test-connection my-plugin

# Permission issues
portunix plugin permissions-check my-plugin

# Configuration problems
portunix plugin config my-plugin validate
```

### Debug Tools

```bash
# Plugin logs
portunix plugin logs my-plugin --follow

# System integration test
portunix plugin system-test my-plugin

# Dependency check
portunix plugin deps my-plugin --verify
```

## Best Practices

### Plugin Development
- Follow semantic versioning
- Comprehensive error handling
- Thorough testing and validation
- Clear documentation and examples
- Performance optimization

### Plugin Management
- Regular updates and security patches
- Monitor plugin health and performance
- Backup plugin configurations
- Review plugin permissions regularly

### Team Collaboration
- Use private registries for internal plugins
- Establish plugin development standards
- Code review processes
- Shared plugin templates

## Future Roadmap

### Planned Features
- **WebAssembly Support** - Run plugins in WASM for enhanced security
- **Plugin Marketplace** - Enhanced discovery and rating system
- **Visual Plugin Builder** - GUI-based plugin development
- **Advanced Analytics** - ML-based plugin performance optimization

### Integration Improvements
- **IDE Plugins** - VS Code, IntelliJ plugin development support
- **Cloud Integration** - Serverless plugin execution
- **API Gateway** - REST/GraphQL API exposure for plugins
- **Event Streaming** - Real-time plugin event processing

## Related Categories

- **[Core](../core/)** - Core system integration
- **[Integration](../integration/)** - MCP and external integrations
- **[Containers](../containers/)** - Container-aware plugins
- **[Utilities](../utilities/)** - Plugin development utilities

---

*The plugin system transforms Portunix from a tool into a platform, enabling unlimited extensibility and customization.*