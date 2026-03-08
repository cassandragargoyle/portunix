---
title: "aiops"
description: "AI operations and GPU management"
---

# aiops

AI operations and GPU management

## Usage

```bash
portunix aiops [options] [arguments]
```

## Full Help

```
Usage: portunix aiops [subcommand]

AI Operations Commands:

GPU Operations:
  gpu status               - Show GPU status and driver info
  gpu status --watch       - Real-time GPU monitoring (default: 5s refresh)
  gpu status --watch --interval 2  - Custom refresh interval
  gpu usage                - Show GPU utilization summary
  gpu processes            - List processes using GPU
  gpu check                - Verify GPU and container toolkit readiness

Ollama Container Operations:
  ollama container create  - Create Ollama container (with GPU if available)
  ollama container create --cpu  - Force CPU-only mode
  ollama container status  - Show Ollama container status
  ollama container start   - Start stopped Ollama container
  ollama container stop    - Stop running Ollama container
  ollama container remove  - Remove Ollama container

Model Operations:
  model list               - List installed models in container
  model list --available   - List available models from Ollama registry
  model install <name>     - Install model to container
  model info <name>        - Show model details
  model remove <name>      - Remove model from container
  model run <name>         - Interactive chat with model

Open WebUI Operations:
  webui container create   - Create Open WebUI container
  webui container status   - Show WebUI container status
  webui container start    - Start stopped WebUI container
  webui container stop     - Stop running WebUI container
  webui container remove   - Remove WebUI container
  webui open               - Open WebUI in browser

Stack Operations:
  stack create             - Create full stack (Ollama + WebUI)
  stack status             - Show all containers status
  stack start              - Start all containers
  stack stop               - Stop all containers
  stack remove             - Remove all containers

```

