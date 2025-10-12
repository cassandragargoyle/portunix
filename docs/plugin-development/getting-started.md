# Getting Started with Portunix Plugin Development

This guide will help you create your first Portunix plugin in under 5 minutes, regardless of your chosen programming language.

## Prerequisites

Before creating a plugin, ensure you have:

1. **Portunix installed** and working on your system
2. **Development environment** for your chosen language
3. **Basic understanding** of gRPC concepts (helpful but not required)

## Quick Plugin Creation

Portunix provides a built-in command to scaffold new plugins:

```bash
portunix plugin create my-first-plugin
```

This command will:
1. Prompt you to choose a programming language
2. Create a directory structure with template files
3. Generate necessary configuration files
4. Provide next steps for development

## Choose Your Language

Select your preferred programming language and follow the specific guide:

- **[Go](languages/go/getting-started.md)** - Native language with best performance
- **[Python](languages/python/getting-started.md)** - Rapid development and rich ecosystem
- **[Java](languages/java/getting-started.md)** - Enterprise-grade with Spring Boot integration
- **[JavaScript/Node.js](languages/javascript/getting-started.md)** - Web ecosystem and npm packages
- **[Rust](languages/rust/getting-started.md)** - Memory safety and high performance

## Plugin Structure Overview

Every Portunix plugin follows this basic structure:

```
my-plugin/
├── plugin.yaml              # Plugin metadata and configuration
├── src/                     # Source code (language-specific structure)
├── proto/                   # gRPC protocol definitions (if custom)
├── tests/                   # Plugin tests
├── docs/                    # Plugin documentation
├── examples/                # Usage examples
└── scripts/                 # Build and deployment scripts
```

### Key Files

#### plugin.yaml
The plugin manifest file that defines:
- Plugin metadata (name, version, description)
- gRPC service configuration
- Dependencies and requirements
- Permissions and capabilities
- MCP tool definitions (if applicable)

#### src/
Contains your plugin's source code. The structure depends on your chosen language:
- Go: `main.go`, package structure
- Python: `main.py`, package structure, `requirements.txt`
- Java: Maven/Gradle project structure
- JavaScript: `index.js`, `package.json`
- Rust: `main.rs`, `Cargo.toml`

## Plugin Development Workflow

1. **Create Plugin Scaffold**
   ```bash
   portunix plugin create my-plugin
   cd my-plugin
   ```

2. **Implement Plugin Logic**
   - Define your plugin's functionality in the source files
   - Follow language-specific best practices
   - Implement required gRPC interfaces

3. **Configure Plugin Manifest**
   - Update `plugin.yaml` with correct metadata
   - Define required permissions
   - Configure MCP tools (if exposing to AI agents)

4. **Test Plugin Locally**
   ```bash
   portunix plugin install .
   portunix plugin enable my-plugin
   portunix plugin start my-plugin
   ```

5. **Validate Plugin**
   ```bash
   portunix plugin validate my-plugin
   portunix plugin health my-plugin
   ```

6. **Package for Distribution**
   ```bash
   portunix plugin package my-plugin
   ```

## Plugin Types and Examples

### Service Plugin Example
A background service that monitors system resources:

```yaml
# plugin.yaml
name: system-monitor
type: service
version: 1.0.0
description: Monitors system resources and alerts on thresholds
```

### Tool Plugin Example
A code generator tool:

```yaml
# plugin.yaml
name: code-generator
type: tool
version: 1.0.0
description: Generates boilerplate code for common patterns
```

### MCP Plugin Example
A plugin that exposes tools to AI agents:

```yaml
# plugin.yaml
name: ai-helper
type: mcp
version: 1.0.0
description: Provides AI-friendly tools for development tasks
mcp:
  tools:
    - name: generate_dockerfile
      description: Generate optimized Dockerfile for any project
```

## Common Plugin Patterns

### Configuration Management
```yaml
# plugin.yaml
configuration:
  required:
    - api_key
    - endpoint_url
  optional:
    - timeout: 30
    - retry_count: 3
```

### Dependency Declaration
```yaml
# plugin.yaml
dependencies:
  system:
    - docker
    - git
  plugins:
    - database-connector: ">=1.2.0"
```

### Permission Requirements
```yaml
# plugin.yaml
permissions:
  filesystem:
    - read: ["/tmp", "/var/log"]
    - write: ["/tmp/plugin-data"]
  network:
    - outbound: ["api.example.com:443"]
  system:
    - execute: ["docker", "git"]
```

## Next Steps

1. **Choose your language** and follow the specific guide
2. **Study examples** in the chosen language directory
3. **Read architecture overview** for deeper understanding
4. **Explore MCP integration** if building AI-friendly tools
5. **Join the community** for support and collaboration

## Troubleshooting

### Plugin Won't Start
- Check plugin logs: `portunix plugin logs my-plugin`
- Validate configuration: `portunix plugin validate my-plugin`
- Check dependencies: `portunix plugin info my-plugin`

### Permission Errors
- Review required permissions in `plugin.yaml`
- Check Portunix security settings
- Verify user permissions on the host system

### gRPC Connection Issues
- Ensure plugin implements required interfaces
- Check port conflicts with other plugins
- Verify network connectivity

## Resources

- [Plugin Architecture](architecture.md) - Deep dive into the plugin system
- [Testing Guide](testing/unit-testing.md) - How to test your plugins
- [Deployment Guide](deployment/packaging.md) - Package and distribute plugins
- [MCP Integration](mcp-integration/exposing-tools.md) - Expose tools to AI agents

---

Ready to build your first plugin? Choose your language and start coding!