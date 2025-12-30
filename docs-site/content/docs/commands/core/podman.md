---
title: "podman"
description: "Podman container management"
---

# podman

Podman container management

## Usage

```bash
portunix podman [options] [arguments]
```

## Full Help

```
Usage: portunix podman [command]

Available Commands:
  check            Check container runtime capabilities and versions
  compose          Run docker-compose/podman-compose commands (universal runtime)
  cp               Copy files/folders between container and host
  exec             Execute command in container (universal runtime)
  info             Show container runtime information and availability
  list             List containers from all available runtimes
  logs             Show container logs (universal runtime)
  rm               Remove container (universal runtime)
  run              Run new container (universal runtime)
  run-in-container Run installation in container (RECOMMENDED for testing)
  start            Start stopped container (universal runtime)
  stop             Stop container (universal runtime)

Flags:
  -h, --help   help for podman

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples

Use "portunix podman [command] --help" for more information about a command.

```

## Subcommands

| Subcommand | Description |
|------------|-------------|
| [check](#check) | Check container runtime capabilities and versions |
| [compose](#compose) | Run docker-compose/podman-compose commands (universal runtime) |
| [cp](#cp) | Copy files/folders between container and host |
| [exec](#exec) | Execute command in container (universal runtime) |
| [info](#info) | Show container runtime information and availability |
| [list](#list) | List containers from all available runtimes |
| [logs](#logs) | Show container logs (universal runtime) |
| [rm](#rm) | Remove container (universal runtime) |
| [run](#run) | Run new container (universal runtime) |
| [run-in-container](#run-in-container) | Run installation in container (RECOMMENDED for testing) |
| [start](#start) | Start stopped container (universal runtime) |
| [stop](#stop) | Stop container (universal runtime) |

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

### list

List containers from all available runtimes

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

### rm

Remove container (universal runtime)

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
  -p, --port: Publish container ports to host
  -v, --volume: Bind mount volumes
  -e, --env: Set environment variables

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

