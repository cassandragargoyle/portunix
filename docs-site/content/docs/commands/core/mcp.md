---
title: "mcp"
description: "MCP server for AI assistants"
---

# mcp

MCP server for AI assistants

## Usage

```bash
portunix mcp [options] [arguments]
```

## Full Help

```
Manage MCP server integration with AI assistants like Claude Code.

This command provides utilities to configure, manage and monitor
the MCP server integration with various AI development tools.

Usage:
  portunix mcp [command]

Available Commands:
  config      Manage MCP server configuration
  configure   Configure Portunix MCP server integration with Claude Code
  init        Initialize MCP server configuration with interactive wizard
  reconfigure Reconfigure Portunix MCP server with new settings (e.g., different mode or port)
  remove      Remove Portunix MCP server integration from Claude Code
  serve       Start MCP server for AI assistant integration
  start       Start MCP server for AI assistant integration
  status      Check status of Portunix MCP server integration with Claude Code
  stop        Stop running MCP server
  test        Test MCP server connection with AI assistants

Flags:
  -h, --help   help for mcp

Use "portunix mcp [command] --help" for more information about a command.

```

## Subcommands

| Subcommand | Description |
|------------|-------------|
| [config](#config) | Manage MCP server configuration |
| [configure](#configure) | Configure Portunix MCP server integration with Claude Code |
| [init](#init) | Initialize MCP server configuration with interactive wizard |
| [reconfigure](#reconfigure) | Reconfigure Portunix MCP server with new settings (e.g., different mode or port) |
| [remove](#remove) | Remove Portunix MCP server integration from Claude Code |
| [serve](#serve) | Start MCP server for AI assistant integration |
| [start](#start) | Start MCP server for AI assistant integration |
| [status](#status) | Check status of Portunix MCP server integration with Claude Code |
| [stop](#stop) | Stop running MCP server |
| [test](#test) | Test MCP server connection with AI assistants |

### config

Manage MCP server configuration

```
View and manage MCP server configuration.

This command allows you to:
- View current configuration
- Edit existing configuration
- Add new AI assistants
- Force reconfiguration

Examples:
  portunix mcp config                    # Show current configuration
  portunix mcp config --edit            # Interactive configuration editing
  portunix mcp config --force           # Force reconfiguration
  portunix mcp config --assistant claude-desktop --add  # Add new assistant

Usage:
  portunix mcp config [flags]

Flags:
      --add                Add new assistant to configuration
      --assistant string   Assistant to add (claude-code, claude-desktop, gemini-cli)
      --edit               Interactive configuration editing
      --force              Force reconfiguration
  -h, --help               help for config
      --json               Output in JSON format
```

### configure

Configure Portunix MCP server integration with Claude Code

```
Automatically configure Portunix MCP server to work with Claude Code AI assistant.

This command will:
- Check if Claude Code is installed
- Add Portunix as an MCP server in Claude Code configuration
- Set appropriate permissions and mode settings
- Verify the integration works

Server Modes:
  stdio (default) - Direct stdin/stdout communication (recommended for Claude Code)
  tcp             - HTTP/TCP server on specified port
  unix            - Unix domain socket (Linux/macOS only)

Configuration Scope:
  local (default) - Project-local configuration (.mcp.json in current directory)
  user            - User-wide/global configuration
  project         - Project configuration

Examples:
  portunix mcp configure                          # Local scope, stdio mode (recommended)
  portunix mcp configure --scope user             # Global/user-wide configuration
  portunix mcp configure --mode tcp --port 3001   # Use TCP mode with custom port
  portunix mcp configure --mode unix              # Use Unix socket mode

Usage:
  portunix mcp configure [flags]

Flags:
  -f, --force                Force reconfiguration even if already configured
  -h, --help                 help for configure
  -m, --mode string          Server mode: stdio (default, recommended), tcp, unix (default "stdio")
  -r, --permissions string   Permission level: limited, standard, full (default "standard")
  -p, --port int             Port for MCP server (only used with --mode tcp) (default 3001)
  -s, --scope string         Configuration scope: local (default, .mcp.json), user (global), project (default "local")
```

### init

Initialize MCP server configuration with interactive wizard

```
Initialize MCP server configuration for AI assistant integration.

This interactive wizard will guide you through:
- Detecting installed AI assistants (Claude Code, Claude Desktop, Gemini CLI)
- Configuring MCP server for each assistant
- Setting up appropriate security profiles
- Testing the integration

Examples:
  portunix mcp init                                          # Interactive wizard
  portunix mcp init --assistant claude-code                  # Claude Code with stdio (default)
  portunix mcp init --assistant claude-code --type stdio     # Claude Code with stdio (explicit)
  portunix mcp init --assistant claude-desktop --type remote # Claude Desktop with remote server
  portunix mcp init --preset development                     # Use development preset

Usage:
  portunix mcp init [flags]

Flags:
      --assistant string   AI assistant to configure (claude-code, claude-desktop, gemini-cli)
      --force              Force reconfiguration even if already configured
  -h, --help               help for init
      --preset string      Use preset configuration (development, standard)
      --type string        Server type (stdio, remote)
```

### reconfigure

Reconfigure Portunix MCP server with new settings (e.g., different mode or port)

```
Reconfigure existing Portunix MCP server integration with new settings.
This command will remove the current configuration and set up a new one with
the specified parameters.

Examples:
  portunix mcp reconfigure                          # Reconfigure with stdio mode, local scope (default)
  portunix mcp reconfigure --scope user             # Reconfigure in global/user scope
  portunix mcp reconfigure --mode tcp --port 3002   # Change to TCP mode on port 3002
  portunix mcp reconfigure --permissions full       # Change permissions
  portunix mcp reconfigure --auto-port              # Auto-find available port (TCP mode)
  portunix mcp reconfigure --force                  # Force reconfigure even if not configured

Usage:
  portunix mcp reconfigure [flags]

Flags:
  -a, --auto-port            Automatically find available port (only for --mode tcp)
  -f, --force                Force reconfigure even if not configured
  -h, --help                 help for reconfigure
  -m, --mode string          Server mode: stdio (default, recommended), tcp, unix (default "stdio")
  -r, --permissions string   Permission level: limited, standard, full (default "limited")
  -p, --port int             Port for MCP server (only used with --mode tcp, 0 = use default 3001)
  -s, --scope string         Configuration scope: local (default, .mcp.json), user (global), project (default "local")
```

### remove

Remove Portunix MCP server integration from Claude Code

```
Remove Portunix MCP server configuration from Claude Code AI assistant.

This command will:
- Check if Portunix MCP server is configured in Claude Code
- Remove the MCP server configuration
- Verify the removal was successful

Examples:
  portunix mcp remove           # Remove integration
  portunix mcp remove --force   # Force removal without confirmation

Usage:
  portunix mcp remove [flags]

Flags:
  -f, --force   Force removal without confirmation
  -h, --help    help for remove
```

### serve

Start MCP server for AI assistant integration

```
Start Model Context Protocol (MCP) server to enable AI assistants
to interact with Portunix functionality.

Communication Modes:
  stdio    - Standard input/output for direct AI integration (default)
  tcp      - TCP socket server for network-based connections
  unix     - Unix domain socket for local IPC

Examples:
  portunix mcp serve                           # Start in stdio mode (default)
  portunix mcp serve --mode stdio              # Explicit stdio mode
  portunix mcp serve --mode tcp --port 3001    # TCP mode on port 3001

Usage:
  portunix mcp serve [flags]
  portunix mcp serve [command]

Available Commands:
  status      Show MCP server status and health information

Flags:
  -c, --config string        Path to configuration file
  -h, --help                 help for serve
  -m, --mode string          Communication mode: stdio, tcp, unix (default "stdio")
  -r, --permissions string   Permission level: limited, standard, full (default "limited")
  -p, --port int             Port for TCP mode (default 3001)
  -s, --socket string        Socket path for Unix mode (default "/tmp/portunix.sock")

Use "portunix mcp serve [command] --help" for more information about a command.
```

### start

Start MCP server for AI assistant integration

```
Start the MCP server with specified configuration.

The server can run in different modes:
- stdio: For direct process communication (Claude Code)
- remote: For network communication (Claude Desktop)

Examples:
  portunix mcp start                    # Start with default settings
  portunix mcp start --port 3001        # Start on specific port
  portunix mcp start --daemon           # Run as daemon process
  portunix mcp start --stdio            # Run in stdio mode

Usage:
  portunix mcp start [flags]

Flags:
  -d, --daemon               Run as daemon process
  -h, --help                 help for start
  -r, --permissions string   Permission level: limited, standard, full (default "standard")
  -p, --port int             Port to run MCP server on (default 3001)
  -s, --stdio                Run in stdio mode
```

### status

Check status of Portunix MCP server integration with Claude Code

```
Check the current status of Portunix MCP server integration with Claude Code.

This command will:
- Check if Claude Code is installed and accessible
- Verify if Portunix MCP server is configured
- Display detailed configuration information

Examples:
  portunix mcp status           # Show status overview
  portunix mcp status --verbose # Show detailed information
  portunix mcp status --json    # Output in JSON format

Usage:
  portunix mcp status [flags]

Flags:
  -h, --help      help for status
  -j, --json      Output in JSON format
  -v, --verbose   Show detailed information
```

### stop

Stop running MCP server

```
Stop the running MCP server process.

This command will gracefully shutdown the MCP server if it's running
as a daemon process.

Examples:
  portunix mcp stop        # Stop the server
  portunix mcp stop --force # Force stop if graceful shutdown fails

Usage:
  portunix mcp stop [flags]

Flags:
  -f, --force   Force stop the server
  -h, --help    help for stop
```

### test

Test MCP server connection with AI assistants

```
Test the MCP server connection and verify integration with AI assistants.

This command will:
- Check if MCP server is running
- Test connection to the server
- Verify assistant integration
- Run basic functionality tests

Examples:
  portunix mcp test                         # Test all configured assistants
  portunix mcp test --assistant claude-code # Test specific assistant
  portunix mcp test --verbose              # Show detailed test output

Usage:
  portunix mcp test [flags]

Flags:
      --assistant string   Test specific assistant (claude-code, claude-desktop, gemini-cli)
  -h, --help               help for test
  -v, --verbose            Show detailed test output
```

