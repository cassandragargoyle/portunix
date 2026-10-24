# MCP Server Core (`server.go`)

## Overview

`server.go` is the **transport and dispatch core** of the Portunix MCP (Model Context Protocol)
server. It receives JSON-RPC 2.0 requests from an AI assistant (Claude Code and similar),
looks up a handler by method name, and sends back a JSON-RPC response.

It contains **no business logic**. What the server can actually do (tools, resources,
package install, VMs, PFT, edge, ...) lives in the handler files next to it:

| File | Content |
| ----- | ----- |
| `server.go` | Server struct, JSON-RPC types, transports, request dispatch |
| `handlers.go` | MCP protocol methods (`initialize`, `tools/list`, `tools/call`, ...) and tool logic |
| `vm_handlers.go`, `pft_handlers.go`, `edge_handlers.go` | Domain-specific tool handlers |
| `stdio.go` | Alternative stdio entry points (`RunStdioMode`, `StdioHandler`) |

The server is started by the `ptx-mcp` helper binary (`src/helpers/ptx-mcp/main.go`),
which the main `portunix` dispatcher invokes for `portunix mcp serve`.

## Request Flow

```text
AI assistant
    |  one JSON-RPC message (line / WebSocket frame)
    v
transport (stdio | TCP | Unix socket | WebSocket)
    |  json.Unmarshal -> MCPRequest
    v
processRequest()
    |  s.handlers[request.Method]
    v
MethodHandler(params) -> (result, error)
    |  wrapped into MCPResponse
    v
transport writes JSON back to the client
```

All transports share the same `processRequest()`, so a handler behaves identically
regardless of how the client is connected.

## Key Types

| Type | Purpose |
| ---- | ------- |
| `Server` | Holds `Port`, `Permissions`, `Config`, WebSocket upgrader and the handler map |
| `MethodHandler` | `func(params json.RawMessage) (interface{}, error)` — signature of every handler |
| `MCPRequest` | Incoming JSON-RPC request (`jsonrpc`, `method`, `params`, `id`) |
| `MCPResponse` | Outgoing JSON-RPC response (`result` or `error`, same `id`) |
| `MCPError` | JSON-RPC error object (`code`, `message`, `data`) |

`NewServer(port, permissions, config)` builds the struct and calls `registerHandlers()`.

## Transports

| Method | Transport | Framing | Used by |
| ------ | --------- | ------- | ------- |
| `StartStdio()` | stdin/stdout | line-based (`bufio.Scanner`) | `ptx-mcp serve --mode stdio` (default) |
| `StartStdioNewline()` | stdin/stdout | newline-delimited, buffered, CRLF-tolerant | `RunStdioMode()` in `stdio.go` |
| `StartTCP(port)` | TCP listener | newline-delimited | `ptx-mcp serve --mode tcp` |
| `StartUnix(path)` | Unix domain socket | newline-delimited | `ptx-mcp serve --mode unix` |
| `Start(daemon)` | HTTP + WebSocket (`/mcp`, `/health`) | one JSON per WS message | not wired to any command |

Common behavior:

- **stdout is reserved for JSON-RPC** in stdio modes; all logs go to stderr
  (`log.SetOutput(os.Stderr)`), otherwise the client would fail to parse output
- Every transport installs a SIGINT/SIGTERM handler for graceful shutdown
- Malformed JSON lines are logged to stderr and skipped — no error response is sent
- TCP and Unix connections are served concurrently (one goroutine per connection) via
  `handleGenericConnection()`, which uses a 30 s read deadline and on timeout sends a
  keep-alive message `{"jsonrpc":"2.0","result":{"type":"ping"}}` instead of closing
- `StartUnix` removes a stale socket file before listening and deletes it on shutdown
- `/health` returns `status`, `port`, `permissions` and a UTC timestamp

## Dispatch and Errors

`processRequest()` returns:

| Situation | JSON-RPC error code | `message` | `data` |
| --------- | ------------------- | --------- | ------ |
| Method not registered | `-32601` | `Method not found` | method name |
| Handler returned an error | `-32603` | `Internal error` | `err.Error()` |
| Success | — | — | handler result in `result` |

## Registered Methods

`registerHandlers()` maps method names to handlers. Two groups:

- **Standard MCP protocol** — what real MCP clients call:
  `initialize`, `ping`, `tools/list`, `tools/call`, `resources/list`, `resources/read`.
  Individual tools exposed to the AI (`get_system_info`, `install_package`, `vm_*`,
  `pft_*`, ...) are listed in `handleToolsList` and routed inside `handleToolsCall`
  in `handlers.go`, **not** here
- **Legacy direct methods** with `mcp_` prefix (`mcp_get_system_info`,
  `mcp_install_package`, `mcp_list_containers`, `mcp_create_project`, ...) — callable as
  top-level JSON-RPC methods, grouped as system info, dev environment, packages,
  containers, security and workflow automation

To add a new AI tool, extend `handleToolsList` / `handleToolsCall` in `handlers.go`;
touch `registerHandlers()` only for a new top-level JSON-RPC method.

## Notes and Caveats

- `Permissions` is only reported (startup log, `/health`); `server.go` does not enforce it
- `Config` is stored but not read in this file
- WebSocket `CheckOrigin` accepts all origins
- `IsRunningFromClaudeCode()` is a heuristic (piped stdin, `$_` contains `claude`,
  missing `/proc/self/fd/0`), prints `DEBUG:` lines to stderr and is currently not called
  anywhere; on Windows the `/proc` check always reports true
- `StartStdio` and `StartStdioNewline` are near-duplicates; `ptx-mcp` uses the former,
  `stdio.go` the latter
- `StartStdio` uses `bufio.Scanner` with the default 64 KiB line limit — larger requests
  end the loop with an error
- `TCP`/`Unix` keep-alive ping is sent as a response without `id`, which strict JSON-RPC
  clients may reject
