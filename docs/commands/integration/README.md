# Integration Commands

AI assistant integration and external system connectivity through the Model Context Protocol (MCP) and other integration mechanisms.

## Commands in this Category

### [`mcp`](mcp.md) - Model Context Protocol Server
MCP server for seamless integration with AI assistants like Claude, providing structured tool access and real-time communication.

**Quick Examples:**
```bash
portunix mcp configure                    # Setup MCP server
portunix mcp serve                        # Start MCP server
portunix mcp status                       # Check server status
portunix mcp tools list                   # Show available MCP tools
```

**Key Features:**
- Claude and other AI assistant integration
- Structured tool exposure to AI systems
- Real-time bidirectional communication
- Custom tool development framework
- Security controls and authentication
- Performance monitoring and analytics

**Common Use Cases:**
- AI-assisted development workflows
- Automated task execution via AI
- Natural language system administration
- Intelligent development environment setup
- AI-powered troubleshooting and diagnostics

---

### `api` - REST API Server *(Coming Soon)*
RESTful API server for external system integration and programmatic access to Portunix functionality.

**Planned Features:**
- Full REST API for all Portunix commands
- WebSocket support for real-time events
- Authentication and authorization
- Rate limiting and request validation
- API documentation and testing tools

## MCP Integration Architecture

### AI Assistant Communication

The MCP server enables sophisticated AI assistant integration:

```
┌─────────────────┐    JSON-RPC/MCP    ┌─────────────────┐
│   AI Assistant  │ ◄─────────────────► │  Portunix MCP   │
│    (Claude)     │     Protocol        │     Server      │
└─────────────────┘                     └─────────────────┘
       │                                         │
       │ Natural Language                        │ Structured
       │ Instructions                            │ Tool Calls
       ▼                                         ▼
┌─────────────────┐                     ┌─────────────────┐
│   User Intent   │                     │  Portunix Core  │
│   Interpretation│                     │   Operations    │
└─────────────────┘                     └─────────────────┘
```

### Available MCP Tools

The MCP server exposes these tools to AI assistants:

#### Core System Tools
- `install_package` - Install software packages
- `system_info` - Get comprehensive system information
- `update_portunix` - Update Portunix to latest version
- `check_requirements` - Verify system requirements

#### Container Management Tools
- `create_container` - Create and configure containers
- `manage_containers` - Start, stop, and manage containers
- `container_ssh` - Access containers via SSH
- `container_logs` - Retrieve container logs

#### Virtualization Tools
- `create_vm` - Create virtual machines
- `manage_vms` - VM lifecycle management
- `vm_snapshots` - Snapshot creation and management
- `vm_provisioning` - Automated VM setup

#### Plugin Management Tools
- `plugin_lifecycle` - Install, enable, and manage plugins
- `plugin_development` - Create and develop plugins
- `plugin_health` - Monitor plugin status

#### File System Tools
- `file_operations` - Read, write, and manage files
- `directory_operations` - Directory management
- `search_files` - File content and name searching

## AI-Assisted Workflows

### Development Environment Setup

```
User: "Set up a full-stack development environment with Node.js, Python,
       PostgreSQL, and Redis in a container"

AI Assistant via MCP:
1. create_container(image: "ubuntu:22.04", name: "fullstack-dev")
2. install_package(packages: ["nodejs", "python", "postgresql", "redis"])
3. container_ssh(container: "fullstack-dev", setup_services: true)
4. system_info(verify_installation: ["nodejs", "python", "postgresql", "redis"])

Result: Complete development environment ready in minutes
```

### Automated Troubleshooting

```
User: "My container won't start and I'm getting network errors"

AI Assistant via MCP:
1. system_info(focus: "containers", "network")
2. container_logs(container: "user-container", errors_only: true)
3. check_requirements(component: "docker", network: true)
4. suggest_fixes(based_on: "analysis_results")

Result: Automated diagnosis and fix suggestions
```

### Cross-Platform Testing

```
User: "Test my application across different operating systems"

AI Assistant via MCP:
1. create_vm(os: "ubuntu-22.04", name: "test-ubuntu")
2. create_vm(os: "centos-8", name: "test-centos")
3. create_vm(os: "windows-server-2022", name: "test-windows")
4. For each VM:
   - vm_provisioning(install_app: true)
   - run_tests()
   - collect_results()

Result: Comprehensive cross-platform test results
```

## Custom MCP Tool Development

### Creating Custom Tools

```bash
# Create custom MCP tool
portunix mcp tool create backup-database

# Tool definition (backup-database/tool.json)
{
  "name": "backup_database",
  "description": "Backup database with compression and verification",
  "inputSchema": {
    "type": "object",
    "properties": {
      "database": {
        "type": "string",
        "description": "Database name or connection string"
      },
      "output_path": {
        "type": "string",
        "description": "Backup file destination"
      },
      "compression": {
        "type": "string",
        "enum": ["gzip", "bzip2", "none"],
        "default": "gzip"
      },
      "verify": {
        "type": "boolean",
        "default": true,
        "description": "Verify backup integrity"
      }
    },
    "required": ["database", "output_path"]
  }
}

# Tool implementation (backup-database/handler.sh)
#!/bin/bash
DATABASE=$1
OUTPUT_PATH=$2
COMPRESSION=${3:-gzip}
VERIFY=${4:-true}

# Backup logic
case $COMPRESSION in
  gzip)
    mysqldump $DATABASE | gzip > $OUTPUT_PATH
    ;;
  bzip2)
    mysqldump $DATABASE | bzip2 > $OUTPUT_PATH
    ;;
  none)
    mysqldump $DATABASE > $OUTPUT_PATH
    ;;
esac

# Verification
if [ "$VERIFY" = "true" ]; then
  echo "Verifying backup integrity..."
  # Verification logic
fi

echo "Backup completed: $OUTPUT_PATH"
```

### Advanced Tool Features

```javascript
// JavaScript-based custom tool
// tools/deployment-manager/handler.js

const { MCPTool } = require('@portunix/mcp-sdk');

class DeploymentManager extends MCPTool {
  async deploy(params) {
    const { environment, application, version } = params;

    try {
      // Pre-deployment checks
      await this.validateEnvironment(environment);
      await this.checkResources(application);

      // Deploy application
      const deploymentId = await this.performDeployment({
        environment,
        application,
        version
      });

      // Post-deployment verification
      await this.verifyDeployment(deploymentId);

      return {
        success: true,
        deploymentId,
        message: `Successfully deployed ${application} v${version} to ${environment}`
      };

    } catch (error) {
      return {
        success: false,
        error: error.message,
        rollbackId: await this.rollback()
      };
    }
  }

  async validateEnvironment(environment) {
    // Environment validation logic
    const systemInfo = await this.portunix.system.info();
    const requirements = await this.checkRequirements(environment);

    if (!requirements.met) {
      throw new Error(`Environment requirements not met: ${requirements.missing}`);
    }
  }
}

module.exports = DeploymentManager;
```

## Integration Security

### Authentication and Authorization

```bash
# Enable MCP authentication
portunix mcp security auth enable --method jwt

# Generate API keys for AI assistants
portunix mcp security apikey create \
  --name claude-desktop \
  --permissions "install,system,containers" \
  --expires 90d

# Configure client certificates
portunix mcp security cert generate \
  --client claude-desktop \
  --validity 1y
```

### Rate Limiting and Controls

```bash
# Configure rate limiting
portunix mcp security ratelimit configure \
  --requests-per-minute 100 \
  --burst 20 \
  --client claude-desktop

# Set operation timeouts
portunix mcp security timeout set \
  --tool install_package --timeout 300s \
  --tool create_vm --timeout 600s

# Enable audit logging
portunix mcp security audit enable \
  --log-level detailed \
  --retention 30d
```

### Network Security

```bash
# Enable TLS/SSL
portunix mcp security ssl enable \
  --cert server.crt \
  --key server.key \
  --auto-cert

# Restrict access by IP
portunix mcp security firewall enable \
  --allow 192.168.1.0/24 \
  --allow 10.0.0.0/8 \
  --deny all

# VPN integration
portunix mcp security vpn configure \
  --provider wireguard \
  --config vpn.conf
```

## Advanced Integration Patterns

### Multi-Tenant MCP Setup

```bash
# Create tenant configurations
portunix mcp tenant create development \
  --tools "install,system,containers" \
  --resource-limits "cpu:2,memory:4G"

portunix mcp tenant create production \
  --tools "system,monitoring" \
  --resource-limits "cpu:1,memory:2G" \
  --read-only

# Switch tenant context
portunix mcp tenant use development
```

### Workflow Automation

```yaml
# workflow-config.yaml
workflows:
  - name: "setup-dev-environment"
    description: "Complete development environment setup"
    steps:
      - tool: system_info
        params:
          components: ["hardware", "os", "network"]

      - tool: install_package
        params:
          packages: ["nodejs", "python", "docker", "git"]
          profile: "development"
        conditions:
          - system.ram >= "8GB"
          - system.cpu_cores >= 4

      - tool: create_container
        params:
          name: "dev-environment"
          image: "ubuntu:22.04"
          packages: ["vscode-server", "development-tools"]
        depends_on: ["install_package"]

      - tool: plugin_lifecycle
        params:
          action: "install"
          plugins: ["agile-software-development"]
        depends_on: ["create_container"]

  - name: "backup-and-update"
    description: "System backup and update workflow"
    schedule: "daily"
    steps:
      - tool: backup_system
        params:
          include: ["configurations", "user-data"]

      - tool: update_portunix
        params:
          channel: "stable"
          verify: true
        depends_on: ["backup_system"]
```

### Event-Driven Integration

```bash
# Configure event webhooks
portunix mcp events webhook add \
  --url https://hooks.slack.com/services/xxx \
  --events "install.completed,vm.created,error.occurred"

# Custom event handlers
portunix mcp events handler create deployment-success \
  --trigger "install.completed" \
  --action "send-notification" \
  --template "Package {{package}} installed successfully"

# Event filtering
portunix mcp events filter add \
  --event "*.error" \
  --condition "severity >= warning" \
  --action "alert-ops-team"
```

## Performance and Scalability

### Load Balancing

```bash
# Multi-instance MCP setup
portunix mcp cluster create \
  --instances 3 \
  --load-balancer round-robin \
  --health-check-interval 30s

# Auto-scaling configuration
portunix mcp autoscale configure \
  --min-instances 2 \
  --max-instances 10 \
  --cpu-threshold 80% \
  --memory-threshold 85%
```

### Caching and Optimization

```bash
# Enable response caching
portunix mcp cache enable \
  --ttl 300s \
  --max-size 1GB \
  --strategies "lru,lfu"

# Tool result caching
portunix mcp cache tools configure \
  --tool system_info --ttl 60s \
  --tool install_package --ttl 0 \
  --tool file_operations --ttl 30s
```

### Monitoring and Analytics

```bash
# Real-time monitoring
portunix mcp monitor --dashboard

# Export metrics
portunix mcp metrics export \
  --format prometheus \
  --endpoint http://prometheus:9090

# Performance analysis
portunix mcp analyze \
  --period 24h \
  --focus "response-times,error-rates,usage-patterns"
```

## Integration Examples

### Claude Desktop Integration

```json
// claude_desktop_config.json
{
  "mcpServers": {
    "portunix": {
      "command": "portunix",
      "args": ["mcp", "serve"],
      "env": {
        "PORTUNIX_MCP_LOG_LEVEL": "info",
        "PORTUNIX_MCP_AUTH": "true"
      }
    }
  }
}
```

### VS Code Extension Integration

```typescript
// VS Code extension integration
import { MCPClient } from '@portunix/mcp-client';

export class PortunixExtension {
  private mcpClient: MCPClient;

  async activate() {
    this.mcpClient = new MCPClient('http://localhost:3000');

    // Register VS Code commands
    vscode.commands.registerCommand('portunix.install', async (packageName) => {
      const result = await this.mcpClient.call('install_package', {
        package: packageName,
        variant: 'latest'
      });

      vscode.window.showInformationMessage(
        `Package ${packageName} installed successfully`
      );
    });
  }
}
```

### CI/CD Pipeline Integration

```yaml
# GitHub Actions
name: Deploy with Portunix MCP
on: [push]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Setup Portunix
        run: |
          curl -sSL https://install.portunix.ai | bash
          portunix mcp serve --daemon

      - name: Deploy via MCP
        run: |
          curl -X POST http://localhost:3000/tools/deploy \
            -H "Content-Type: application/json" \
            -d '{
              "environment": "${{ github.ref_name }}",
              "application": "${{ github.repository }}",
              "version": "${{ github.sha }}"
            }'
```

## Troubleshooting Integration Issues

### MCP Connection Problems

```bash
# Test MCP server connectivity
portunix mcp ping

# Validate MCP configuration
portunix mcp config validate

# Debug MCP communication
portunix mcp serve --debug --trace-messages

# Check client compatibility
portunix mcp client test claude-desktop
```

### Tool Execution Issues

```bash
# Test specific MCP tool
portunix mcp tools test install_package \
  --params '{"package": "nodejs"}'

# Debug tool execution
portunix mcp tools debug install_package \
  --trace --verbose

# Validate tool definitions
portunix mcp tools validate --all
```

### Performance Issues

```bash
# Analyze MCP performance
portunix mcp analyze --performance \
  --period 1h \
  --focus response-times

# Monitor resource usage
portunix mcp monitor --resources \
  --alerts --thresholds "cpu:80,memory:85"

# Optimize configuration
portunix mcp optimize --auto-tune
```

## Best Practices

### Security
- Always enable authentication for production
- Use least-privilege principle for tool access
- Regular security audits and updates
- Monitor and log all MCP interactions

### Performance
- Implement appropriate caching strategies
- Monitor resource usage and scale accordingly
- Optimize tool execution times
- Use async operations where possible

### Development
- Thorough testing of custom tools
- Clear documentation and examples
- Error handling and graceful degradation
- Version compatibility management

## Future Roadmap

### Planned Features
- **Multi-Protocol Support** - GraphQL, WebSocket, gRPC integrations
- **Advanced AI Features** - Context awareness, learning capabilities
- **Workflow Engine** - Visual workflow builder and automation
- **Marketplace Integration** - Third-party tool marketplace

### Integration Improvements
- **IDE Plugins** - IntelliJ, Eclipse, Vim integration
- **Mobile Apps** - Mobile AI assistant integration
- **Voice Interfaces** - Voice command processing
- **AR/VR Support** - Spatial computing integration

## Related Categories

- **[Core](../core/)** - Core tools exposed via MCP
- **[Plugins](../plugins/)** - Plugin tools in MCP
- **[Containers](../containers/)** - Container management via MCP
- **[Virtualization](../virtualization/)** - VM operations via MCP

---

*Integration commands bridge the gap between human intent and system operations, enabling natural language system administration and AI-assisted development.*