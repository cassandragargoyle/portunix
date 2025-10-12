# Portunix MCP Command

## Quick Start

The `mcp` command manages the Model Context Protocol (MCP) server, enabling seamless integration between Portunix and AI assistants like Claude.

### Simplest Usage
```bash
# Configure MCP server
portunix mcp configure

# Start MCP server
portunix mcp serve

# Check MCP status
portunix mcp status
```

### Basic Syntax
```bash
portunix mcp [subcommand] [options]
```

### Common Subcommands
- `configure` - Set up MCP server configuration
- `serve` - Start the MCP server
- `status` - Check server status
- `stop` - Stop the server
- `reconfigure` - Update configuration
- `remove` - Remove MCP configuration

## Intermediate Usage

### MCP Configuration

Initial setup of the MCP server:

```bash
# Interactive configuration
portunix mcp configure

# Configuration with specific options
portunix mcp configure \
  --port 3000 \
  --host localhost \
  --name "Portunix MCP Server"

# Configuration for specific AI assistant
portunix mcp configure --client claude

# View current configuration
portunix mcp config show
```

Configuration prompts:
```
MCP Server Configuration
========================
Server Name: Portunix Development Server
Port: 3000
Host: localhost
Enable Authentication: No
AI Assistant: Claude
Auto-start: Yes
Log Level: info
```

### Starting the MCP Server

Run the MCP server for AI assistant integration:

```bash
# Start server (foreground)
portunix mcp serve

# Start server in background
portunix mcp serve --daemon

# Start with specific configuration
portunix mcp serve --config custom-mcp.json

# Start with custom port
portunix mcp serve --port 3001

# Start with verbose logging
portunix mcp serve --verbose
```

### Server Status and Management

Monitor and manage the MCP server:

```bash
# Check server status
portunix mcp status

# Detailed status information
portunix mcp status --detailed

# Stop the server
portunix mcp stop

# Restart the server
portunix mcp restart

# Reload configuration
portunix mcp reload
```

Status output example:
```
MCP Server Status
=================
Status: Running
PID: 12345
Port: 3000
Uptime: 2h 15m 33s
Connections: 1 active
Requests: 147 total
Errors: 0
Memory Usage: 45 MB
CPU Usage: 0.2%
```

### MCP Tools and Capabilities

The MCP server exposes various tools to AI assistants:

```bash
# List available tools
portunix mcp tools list

# Test specific tool
portunix mcp tools test install_package

# Add custom tool
portunix mcp tools add my-custom-tool

# Remove tool
portunix mcp tools remove my-custom-tool

# Tool documentation
portunix mcp tools docs install_package
```

Available MCP tools:
- `install_package` - Install software packages
- `system_info` - Get system information
- `run_command` - Execute shell commands
- `manage_containers` - Container operations
- `plugin_management` - Plugin lifecycle
- `file_operations` - File system operations

## Advanced Usage

### MCP Server Architecture

The MCP server implements the Model Context Protocol specification:

```
┌─────────────────┐    MCP Protocol    ┌─────────────────┐
│   AI Assistant  │ ◄─────────────────► │  Portunix MCP   │
│    (Claude)     │     (JSON-RPC)     │     Server      │
└─────────────────┘                     └─────────────────┘
                                                │
                                                │
                                                ▼
                                        ┌─────────────────┐
                                        │  Portunix Core  │
                                        │   Commands      │
                                        └─────────────────┘
```

### Configuration Management

Advanced configuration options:

```bash
# Export configuration
portunix mcp config export > mcp-config.json

# Import configuration
portunix mcp config import mcp-config.json

# Edit configuration
portunix mcp config edit

# Validate configuration
portunix mcp config validate

# Reset to defaults
portunix mcp config reset
```

Configuration file structure (`mcp-config.json`):
```json
{
  "server": {
    "name": "Portunix MCP Server",
    "version": "1.0.0",
    "port": 3000,
    "host": "localhost",
    "timeout": 30000,
    "maxConnections": 10
  },
  "security": {
    "enableAuth": false,
    "apiKey": "",
    "allowedOrigins": ["*"],
    "rateLimiting": {
      "enabled": true,
      "requests": 100,
      "window": 60000
    }
  },
  "tools": {
    "install_package": {
      "enabled": true,
      "permissions": ["install", "update"]
    },
    "system_info": {
      "enabled": true,
      "permissions": ["read"]
    },
    "run_command": {
      "enabled": false,
      "permissions": ["execute"],
      "restricted": true
    }
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "stdout",
    "file": "/var/log/portunix-mcp.log"
  },
  "integrations": {
    "claude": {
      "enabled": true,
      "endpoint": "stdio",
      "features": ["tools", "prompts", "resources"]
    }
  }
}
```

### Security Configuration

Security features for MCP server:

```bash
# Enable authentication
portunix mcp security auth enable

# Generate API key
portunix mcp security apikey generate

# Set allowed origins
portunix mcp security origins add "https://claude.ai"

# Enable rate limiting
portunix mcp security ratelimit enable --requests 100 --window 60s

# Enable SSL/TLS
portunix mcp security ssl enable --cert server.crt --key server.key
```

### Custom Tool Development

Create custom MCP tools:

```bash
# Create new tool template
portunix mcp tool create backup-database

# Tool definition structure
mkdir ~/.portunix/mcp/tools/backup-database
cat > ~/.portunix/mcp/tools/backup-database/tool.json << EOF
{
  "name": "backup_database",
  "description": "Backup database to specified location",
  "inputSchema": {
    "type": "object",
    "properties": {
      "database": {"type": "string", "description": "Database name"},
      "output": {"type": "string", "description": "Backup file path"}
    },
    "required": ["database", "output"]
  }
}
EOF

# Tool implementation
cat > ~/.portunix/mcp/tools/backup-database/handler.sh << EOF
#!/bin/bash
# Tool handler script
DATABASE=$1
OUTPUT=$2
mysqldump $DATABASE > $OUTPUT
echo "Backup completed: $OUTPUT"
EOF

# Register tool
portunix mcp tool register backup-database
```

### MCP Client Integration

Configure AI assistant clients:

```bash
# Claude Desktop configuration
portunix mcp client configure claude-desktop

# Generate client configuration
portunix mcp client config claude > claude-mcp-config.json

# Test client connection
portunix mcp client test claude
```

Claude Desktop configuration example:
```json
{
  "mcpServers": {
    "portunix": {
      "command": "portunix",
      "args": ["mcp", "serve"],
      "env": {
        "PORTUNIX_MCP_PORT": "3000"
      }
    }
  }
}
```

### Monitoring and Logging

Advanced monitoring capabilities:

```bash
# Real-time monitoring
portunix mcp monitor

# View server logs
portunix mcp logs

# Follow logs in real-time
portunix mcp logs --follow

# Export metrics
portunix mcp metrics --export prometheus

# Performance analysis
portunix mcp analyze --performance
```

### Load Balancing and Scaling

Scale MCP server for multiple connections:

```bash
# Start multiple server instances
portunix mcp cluster start --instances 3

# Load balancer configuration
portunix mcp loadbalancer configure \
  --algorithm round-robin \
  --health-check-interval 30s

# Scale cluster
portunix mcp cluster scale --instances 5

# Cluster status
portunix mcp cluster status
```

## Expert Tips & Tricks

### 1. Development Environment Setup

```bash
# Development mode with hot reload
portunix mcp serve --dev --watch

# Debug mode with detailed logging
portunix mcp serve --debug --trace

# Test environment
portunix mcp test-env setup
```

### 2. Integration Testing

```bash
# Test all MCP tools
portunix mcp test --all-tools

# Performance testing
portunix mcp benchmark --concurrent 10 --duration 60s

# Compatibility testing
portunix mcp compat-test --client claude
```

### 3. Backup and Recovery

```bash
# Backup MCP configuration
portunix mcp backup --output mcp-backup.tar.gz

# Restore configuration
portunix mcp restore --input mcp-backup.tar.gz

# Configuration migration
portunix mcp migrate --from 1.0 --to 2.0
```

### 4. Custom Protocol Extensions

```bash
# Add protocol extension
portunix mcp protocol add-extension streaming

# Custom message handlers
portunix mcp handler add custom-handler.js

# Protocol debugging
portunix mcp protocol debug --capture
```

### 5. Multi-tenant Setup

```bash
# Create tenant configuration
portunix mcp tenant create development

# Switch tenant context
portunix mcp tenant use development

# Tenant isolation
portunix mcp tenant isolate --resources --tools
```

## Integration Examples

### Claude Desktop Integration

```json
{
  "mcpServers": {
    "portunix": {
      "command": "portunix",
      "args": ["mcp", "serve"],
      "env": {
        "PORTUNIX_MCP_LOG_LEVEL": "debug"
      }
    }
  }
}
```

### VS Code Extension Integration

```json
{
  "portunix.mcp": {
    "enabled": true,
    "serverPort": 3000,
    "autoStart": true,
    "tools": ["install_package", "system_info"]
  }
}
```

### API Gateway Integration

```yaml
# Kong API Gateway
services:
  - name: portunix-mcp
    url: http://localhost:3000
    routes:
      - name: mcp-api
        paths: ["/mcp"]
        methods: ["POST"]
```

## Troubleshooting

### Common Issues

#### 1. Server Won't Start
```bash
# Check port availability
portunix mcp port-check 3000

# View startup logs
portunix mcp logs --startup

# Validate configuration
portunix mcp config validate

# Reset configuration
portunix mcp config reset
```

#### 2. Connection Issues
```bash
# Test server connectivity
portunix mcp ping

# Check firewall settings
portunix mcp network-check

# Verify client configuration
portunix mcp client verify claude
```

#### 3. Tool Execution Failures
```bash
# Test specific tool
portunix mcp tool test install_package

# Check tool permissions
portunix mcp tool permissions install_package

# Tool debugging
portunix mcp tool debug install_package --trace
```

#### 4. Performance Issues
```bash
# Analyze server performance
portunix mcp analyze --performance

# Resource monitoring
portunix mcp monitor --resources

# Optimize configuration
portunix mcp optimize
```

### Debug Mode

```bash
# Verbose server output
portunix mcp serve --verbose --debug

# Protocol message tracing
portunix mcp serve --trace-messages

# Tool execution tracing
portunix mcp serve --trace-tools
```

## Security Best Practices

### Authentication and Authorization

```bash
# Enable strong authentication
portunix mcp security auth enable --method jwt

# Configure API keys
portunix mcp security apikey create --name claude-client

# Set up role-based access
portunix mcp security role create --name readonly --permissions read
```

### Network Security

```bash
# Enable HTTPS
portunix mcp security ssl enable --auto-cert

# Configure firewall
portunix mcp security firewall enable --allow claude.ai

# VPN integration
portunix mcp security vpn configure
```

### Audit and Compliance

```bash
# Enable audit logging
portunix mcp audit enable

# Generate compliance report
portunix mcp audit report --format pdf

# Data privacy controls
portunix mcp privacy configure --gdpr
```

## Performance Optimization

### Server Tuning

```bash
# Optimize for high throughput
portunix mcp tune --profile high-throughput

# Memory optimization
portunix mcp tune --memory-limit 512M

# Connection pooling
portunix mcp tune --max-connections 100
```

### Caching

```bash
# Enable response caching
portunix mcp cache enable --ttl 300s

# Tool result caching
portunix mcp cache tools enable

# Clear cache
portunix mcp cache clear
```

## API Reference

### JSON-RPC Protocol

MCP uses JSON-RPC 2.0 for communication:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "install_package",
    "arguments": {
      "package": "nodejs",
      "variant": "latest"
    }
  }
}
```

### Tool Definition Schema

```json
{
  "name": "install_package",
  "description": "Install software packages",
  "inputSchema": {
    "type": "object",
    "properties": {
      "package": {
        "type": "string",
        "description": "Package name to install"
      },
      "variant": {
        "type": "string",
        "description": "Package variant",
        "default": "latest"
      }
    },
    "required": ["package"]
  }
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORTUNIX_MCP_PORT` | MCP server port | 3000 |
| `PORTUNIX_MCP_HOST` | MCP server host | localhost |
| `PORTUNIX_MCP_LOG_LEVEL` | Logging level | info |
| `PORTUNIX_MCP_CONFIG` | Configuration file | ~/.portunix/mcp.json |
| `PORTUNIX_MCP_AUTH` | Enable authentication | false |

## Related Commands

- [`install`](install.md) - Package installation (exposed as MCP tool)
- [`system`](system.md) - System information (exposed as MCP tool)
- [`plugin`](plugin.md) - Plugin management (exposed as MCP tool)
- [`config`](config.md) - Configuration management

## Command Reference

### Complete Parameter List

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--port` | int | `3000` | Server port |
| `--host` | string | `localhost` | Server host |
| `--config` | string | - | Configuration file |
| `--daemon` | boolean | `false` | Run as daemon |
| `--verbose` | boolean | `false` | Verbose output |
| `--debug` | boolean | `false` | Debug mode |
| `--trace` | boolean | `false` | Trace messages |
| `--auth` | boolean | `false` | Enable authentication |
| `--ssl` | boolean | `false` | Enable SSL/TLS |
| `--cert` | string | - | SSL certificate |
| `--key` | string | - | SSL private key |
| `--timeout` | duration | `30s` | Request timeout |
| `--max-connections` | int | `10` | Maximum connections |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Port already in use |
| 4 | Permission denied |
| 5 | SSL/TLS error |
| 6 | Authentication failed |
| 7 | Tool execution error |
| 8 | Client connection error |
| 9 | Resource limit exceeded |
| 10 | Protocol error |

## Version History

- **v1.5.0** - Added custom tool support
- **v1.4.0** - Implemented security features
- **v1.3.0** - Added clustering support
- **v1.2.0** - Enhanced monitoring
- **v1.1.0** - Added configuration management
- **v1.0.0** - Initial MCP server implementation