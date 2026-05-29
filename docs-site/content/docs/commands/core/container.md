---
title: "container"
description: "Universal container management"
---

# container

Universal container management

## Usage

```bash
portunix container [options] [arguments]
```

## Full Help

```
Usage: portunix container [command]

Available Commands:
  check            Check container runtime capabilities and versions
  compose          Run docker-compose/podman-compose commands (universal runtime)
  compose-preflight Check if compose is ready (daemon/socket running)
  cp               Copy files/folders between container and host
  exec             Execute command in container (universal runtime)
  info             Show container runtime information and availability
  inspect          Show low-level container details (universal runtime)
  list             List containers from all available runtimes (aliases: ls, ps)
  logs             Show container logs (universal runtime)
  network          Manage container networks (create/list/inspect/rm)
  cleanup          Remove portunix-managed containers (TTL/age/pattern filters)
  lifecycle        Manage container lifecycle policies (list/inspect/extend/policy)
  restart          Restart container (universal runtime)
  rm               Remove container (universal runtime, alias: remove)
  run              Run new container (universal runtime)
  run-in-container Run installation in container (RECOMMENDED for testing)
  service          Background lifecycle service (start/stop/status)
  start            Start stopped container (universal runtime)
  stop             Stop container (universal runtime)
  volume           Manage container volumes (create/list/inspect/rm/prune)

Flags:
  -h, --help   help for container

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples

Use "portunix container [command] --help" for more information about a command.

```

## Subcommands

| Subcommand | Description |
|------------|-------------|
| [check](#check) | Check container runtime capabilities and versions |
| [compose](#compose) | Run docker-compose/podman-compose commands (universal runtime) |
| [compose-preflight](#compose-preflight) | Check if compose is ready (daemon/socket running) |
| [cp](#cp) | Copy files/folders between container and host |
| [exec](#exec) | Execute command in container (universal runtime) |
| [info](#info) | Show container runtime information and availability |
| [inspect](#inspect) | Show low-level container details (universal runtime) |
| [list](#list) | List containers from all available runtimes (aliases: ls, ps) |
| [logs](#logs) | Show container logs (universal runtime) |
| [network](#network) | Manage container networks (create/list/inspect/rm) |
| [cleanup](#cleanup) | Remove portunix-managed containers (TTL/age/pattern filters) |
| [lifecycle](#lifecycle) | Manage container lifecycle policies (list/inspect/extend/policy) |
| [restart](#restart) | Restart container (universal runtime) |
| [rm](#rm) | Remove container (universal runtime, alias: remove) |
| [run](#run) | Run new container (universal runtime) |
| [run-in-container](#run-in-container) | Run installation in container (RECOMMENDED for testing) |
| [service](#service) | Background lifecycle service (start/stop/status) |
| [start](#start) | Start stopped container (universal runtime) |
| [stop](#stop) | Stop container (universal runtime) |
| [volume](#volume) | Manage container volumes (create/list/inspect/rm/prune) |

### check

Check container runtime capabilities and versions

```
Usage: portunix container check [OPTIONS]

🔍 CHECK CONTAINER RUNTIME CAPABILITIES

Detect and display detailed information about available container runtimes.

🌟 DETECTION INCLUDES:
  • Installed container runtimes (Docker/Podman)
  • Runtime versions and build information
  • Supported features and capabilities
  • System compatibility status
  • Recommendations for optimal setup

Options:
  --refresh      Force re-detection of capabilities
  -h, --help     Show this help message

Examples:
  portunix container check
  portunix container check --refresh

This command helps diagnose container runtime issues and verify proper installation.
```

### compose

Run docker-compose/podman-compose commands (universal runtime)

```
Usage: portunix container compose [args...]

🐳 UNIVERSAL COMPOSE COMMAND

Execute docker-compose or podman-compose commands using automatic runtime detection.
All arguments are passed directly to the detected compose tool.

🌟 AUTOMATIC RUNTIME DETECTION:
  Priority order:
  1. Docker Compose V2 (docker compose) - preferred
  2. Docker Compose V1 (docker-compose) - fallback
  3. Podman Compose (podman-compose) - alternative

💡 USAGE:
  All standard compose commands and flags are supported:

  portunix container compose -f <file> up [service]
  portunix container compose -f <file> down
  portunix container compose -f <file> build [service]
  portunix container compose -f <file> logs [service]
  portunix container compose -f <file> ps
  portunix container compose -f <file> exec <service> <command>

Examples:
  portunix container compose -f docker-compose.yml up -d
  portunix container compose -f docker-compose.yml down
  portunix container compose -f docker-compose.docs.yml up docs-server
  portunix container compose -f docker-compose.yml logs -f web
  portunix container compose -f docker-compose.yml ps
  portunix container compose -f docker-compose.yml build --no-cache
```

### compose-preflight

Check if compose is ready (daemon/socket running)

```
Usage: portunix container compose-preflight [OPTIONS]

🔍 CHECK COMPOSE READINESS

Verify that compose tools are ready to use. This checks:
  • Docker/Podman installation
  • Docker daemon or Podman socket status
  • Compose tool availability

Options:
  --json       Output result as JSON for programmatic use
  -h, --help   Show this help message

Exit codes:
  0 - Compose is ready
  1 - Compose is NOT ready (with instructions to fix)

Examples:
  portunix container compose-preflight
  portunix container compose-preflight --json
```

### cp

Copy files/folders between container and host

```
Usage: portunix container cp <source> <destination>

📁 COPY FILES BETWEEN CONTAINER AND HOST

Copy files or directories between a container and the local filesystem.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Supports copying in both directions
  ✅ Preserves file permissions

Arguments:
  <source>        Source path (local file or container:path)
  <destination>   Destination path (local file or container:path)

Options:
  -h, --help      Show this help message

Examples:
  portunix container cp ./config.json mycontainer:/app/config.json
  portunix container cp mycontainer:/var/log/app.log ./logs/
  portunix container cp ./scripts/ mycontainer:/opt/scripts/
```

### exec

Execute command in container (universal runtime)

```
Usage: portunix container exec <container-name> <command> [args...]

🔧 EXECUTE COMMAND IN CONTAINER

Run a command inside a running container.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Supports interactive commands
  ✅ Pass-through of command arguments

Arguments:
  <container-name>   Name or ID of the container
  <command>          Command to execute
  [args...]          Optional arguments for the command

Options:
  -h, --help         Show this help message

Examples:
  portunix container exec my-container bash
  portunix container exec my-container ls -la /app
  portunix container exec web-server cat /etc/nginx/nginx.conf
  portunix container exec python-dev python --version
```

### info

Show container runtime information and availability

```
Usage: portunix container info

ℹ️ CONTAINER RUNTIME INFORMATION

Display information about available container runtimes.

🌟 DISPLAYS:
  ✅ Docker availability and version
  ✅ Podman availability and version
  ✅ Runtime status and configuration

Options:
  -h, --help      Show this help message

Examples:
  portunix container info
```

### inspect

Show low-level container details (universal runtime)

```
Usage: portunix container inspect [OPTIONS] <container-name> [<container-name>...]

🔎 INSPECT CONTAINER

Return low-level information on the given container(s) from the
automatically selected runtime (Podman first, Docker fallback).

Options:
  -f, --format <tmpl>   Go template for selective output (runtime semantics)
  -h, --help            Show this help message

Examples:
  portunix container inspect my-container
  portunix container inspect my-container -f '{{.NetworkSettings.Networks}}'
  portunix container inspect my-container --format '{{.Config.Env}}'
```

### list

List containers from all available runtimes (aliases: ls, ps)

```
Usage: portunix container list [OPTIONS]

📋 LIST CONTAINERS

Display containers from all available runtimes.

🌟 UNIVERSAL OPERATION:
  ✅ Shows containers from both Docker and Podman
  ✅ Automatic runtime detection
  ✅ Unified output format
  ✅ Shows running and stopped containers

Options:
  -h, --help      Show this help message

Examples:
  portunix container list
```

### logs

Show container logs (universal runtime)

```
Usage: portunix container logs [OPTIONS] <container-name>

📝 VIEW CONTAINER LOGS

Display logs from a container using the automatically selected runtime.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Real-time log streaming with --follow
  ✅ Consistent output format

Options:
  -f, --follow    Follow log output (stream continuously)
  -h, --help      Show this help message

Examples:
  portunix container logs test-container
  portunix container logs web-server --follow
  portunix container logs python-dev
  portunix container logs db-container -f
```

### network

Manage container networks (create/list/inspect/rm)

```
Usage: portunix container network <subcommand> [options]

🌐 MANAGE CONTAINER NETWORKS

Universal network management that auto-selects Podman or Docker.

Subcommands:
  create <name> [--driver <drv>] [--subnet <CIDR>] [--gateway <IP>]
                  Create a network (idempotent — existing network is a no-op)
  list            List available networks
  inspect <name> [-f '<tmpl>']
                  Show low-level network information
  rm <name>...    Remove one or more networks

Options:
  -h, --help      Show this help message

Examples:
  portunix container network create portunix-odoo-net
  portunix container network create my-net --driver bridge --subnet 10.88.0.0/16
  portunix container network list
  portunix container network inspect portunix-odoo-net -f '{{.Subnets}}'
  portunix container network rm portunix-odoo-net
```

### cleanup

Remove portunix-managed containers (TTL/age/pattern filters)

```
Usage: portunix container cleanup [flags]

🧹 CLEANUP MANAGED CONTAINERS

Removes Portunix-managed containers (label portunix.managed=true) that
match the supplied filter. With no filter, defaults to TTL-expired only.

Flags:
  --all                 Remove every managed container
  --expired             Remove only TTL-expired containers (default)
  --older-than DURATION Match containers older than e.g. 1h, 7d, 30m
  --pattern GLOB        Match container name against a glob (e.g. 'test-*')
  --status STATE        Match container status (running, exited, ...)
  --exclude-running     Skip currently running containers
  --dry-run             Show what would be removed without acting
  -f, --force           Force-remove running containers (rm -f)
  -h, --help            Show this help message

Examples:
  portunix container cleanup --dry-run
  portunix container cleanup --older-than 7d --pattern 'dev-*'
  portunix container cleanup --all --force
```

### lifecycle

Manage container lifecycle policies (list/inspect/extend/policy)

```
Usage: portunix container lifecycle <subcommand>

⏳ MANAGE CONTAINER LIFECYCLE POLICIES

Subcommands:
  list                List all portunix-managed containers with TTL info
  inspect <name>      Show full lifecycle metadata for a container
  extend <name> --by <duration>
                      Extend TTL of a container (e.g. --by 1h)
  policy <name> <policy>
                      Change cleanup policy (on-exit | ttl | manual)

Examples:
  portunix container lifecycle list
  portunix container lifecycle inspect cosmos-testnet
  portunix container lifecycle extend cosmos-testnet --by 30m
  portunix container lifecycle policy cosmos-testnet manual
```

### restart

Restart container (universal runtime)

```
Usage: portunix container restart [OPTIONS] <container-name>

🔄 RESTART CONTAINER

Stop and start a container using the automatically selected runtime.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Preserves container state and data
  ✅ Consistent behavior across runtimes

Options:
  -h, --help      Show this help message

Examples:
  portunix container restart test-container
  portunix container restart web-server
  portunix container restart python-dev
```

### rm

Remove container (universal runtime, alias: remove)

```
Usage: portunix container rm [OPTIONS] <container-name> [<container-name>...]

🗑️ REMOVE CONTAINER

Remove one or more containers using the automatically selected runtime.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Supports force removal of running containers
  ✅ Docker/Podman compatible 'rm' command

Options:
  -f, --force    Force removal of running containers
  -h, --help     Show this help message

Examples:
  portunix container rm test-container
  portunix container rm nodejs-dev --force
  portunix container rm web-server -f
  portunix container rm container1 container2 container3
```

### run

Run new container (universal runtime)

```
Usage: portunix container run [flags] <image> [command...]

🏃 RUN NEW CONTAINER

Create and start a new container using the automatically selected runtime.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman
  ✅ Automatic runtime selection
  ✅ Full compatibility with Docker/Podman flags
  ✅ Interactive and background modes supported

Examples:
  portunix container run ubuntu:22.04 echo "Hello World"
  portunix container run -d --name test-container ubuntu:22.04 bash
  portunix container run -it --name interactive-container ubuntu:22.04 bash
  portunix container run -d -p 8080:80 nginx:latest
  portunix container run -d --name test ubuntu:22.04 -- bash -c "echo test"

Supported flags:
  -d, --detach: Run container in background
  -i, --interactive: Keep STDIN open
  -t, --tty: Allocate pseudo-TTY
  --name: Assign a name to the container
  --network: Connect container to a network
  -p, --port: Publish container ports to host
  -v, --volume: Bind mount volumes
  -e, --env: Set environment variables

Lifecycle flags (issue #027):
  --ttl DURATION         Auto-remove container after this duration (e.g. 2h, 7d)
  --auto-cleanup         Mark container for cleanup-on-exit
  --cleanup-policy P     Cleanup policy: on-exit | ttl | manual
  --health-check CMD     Health check command (recorded as label + --health-cmd)
  --max-memory SIZE      Memory limit alias (forwarded to --memory)
  --max-cpu COUNT        CPU limit alias (forwarded to --cpus)

💡 TIP: For development environments, use 'run-in-container' instead.
Use -- to separate flags from command arguments when needed.
```

### run-in-container

Run installation in container (RECOMMENDED for testing)

```
Usage: portunix container run-in-container [OPTIONS] <PACKAGE>

🐳 RUN PACKAGE INSTALLATION INSIDE CONTAINER

Run package installation inside a container environment for safe testing.

🌟 FEATURES:
  ✅ Isolated testing environment
  ✅ Automatic runtime selection (Podman/Docker)
  ✅ Clean container environment for each test
  ✅ Package installation validation
  ✅ Host system protection

Arguments:
  <PACKAGE>           Package to install (required)

Options:
  --image <IMAGE>     Container image to use (default: ubuntu:22.04)
  -h, --help          Show this help message

Examples:
  portunix container run-in-container nodejs
  portunix container run-in-container python --image debian:bookworm
  portunix container run-in-container ansible --image ubuntu:22.04
  portunix container run-in-container claude-code

💡 RECOMMENDATION: Use this command for testing package installations
   without affecting your host development environment.
```

### service

Background lifecycle service (start/stop/status)

```
Usage: portunix container service <subcommand>

⚙️  CONTAINER LIFECYCLE BACKGROUND SERVICE

Polls all portunix-managed containers and removes those whose TTL has
expired. Daemon state lives under ~/.portunix/container/.

Subcommands:
  start [--interval D] [--detach|--foreground]
                  Start the background lifecycle service (default: detach)
  stop            Stop the background lifecycle service
  status          Show whether the service is running
  run [--interval D]
                  Run the polling loop in the foreground (used by --detach)

Examples:
  portunix container service start --interval 1m
  portunix container service status
  portunix container service stop
```

### start

Start stopped container (universal runtime)

```
Usage: portunix container start [OPTIONS] <container-name>

▶️ START CONTAINER

Start a stopped container using the automatically selected runtime.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Restarts previously stopped containers
  ✅ Consistent behavior across runtimes

Options:
  -h, --help      Show this help message

Examples:
  portunix container start test-container
  portunix container start web-server
  portunix container start python-dev
```

### stop

Stop container (universal runtime)

```
Usage: portunix container stop [OPTIONS] <container-name>

🛑 STOP CONTAINER

Stop a running container using the automatically selected runtime.

🌟 UNIVERSAL OPERATION:
  ✅ Works with both Docker and Podman containers
  ✅ Automatic runtime detection
  ✅ Graceful shutdown of container processes
  ✅ Consistent behavior across runtimes

Options:
  -h, --help      Show this help message

Examples:
  portunix container stop test-container
  portunix container stop web-server
  portunix container stop python-dev
```

### volume

Manage container volumes (create/list/inspect/rm/prune)

```
Usage: portunix container volume <subcommand> [options]

📦 MANAGE CONTAINER VOLUMES

Universal volume management that auto-selects Podman or Docker.

Subcommands:
  create <name> [--driver <drv>]
                  Create a named volume (idempotent)
  list            List available volumes
  inspect <name> [-f '<tmpl>']
                  Show low-level volume information
  rm <name>...    Remove one or more volumes
  prune [--force] Remove all unused volumes

Options:
  -h, --help      Show this help message

Examples:
  portunix container volume create odoo-data
  portunix container volume list
  portunix container volume inspect odoo-data
  portunix container volume rm odoo-data
  portunix container volume prune --force
```

