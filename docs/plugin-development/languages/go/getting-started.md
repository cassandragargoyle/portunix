# Go Plugin Development Guide

Go is the native language for Portunix plugins, offering the best performance and seamless integration with the core system. This guide will
help you create robust, efficient plugins using Go.

## Prerequisites

- Go 1.24 or later
- Protocol Buffers compiler (`protoc`) with the `protoc-gen-go` and `protoc-gen-go-grpc` plugins
- Portunix development environment

## Quick Start

Create a new Go plugin using the Portunix CLI:

```bash
portunix plugin create my-go-plugin --language=go
cd my-go-plugin
make build
```

This creates a complete Go project structure:

```text
my-go-plugin/
├── plugin.json                  # Plugin manifest
├── config.yaml                  # Plugin configuration (optional)
├── go.mod                       # Go module definition
├── go.sum                       # Go dependencies
├── Makefile                     # Build, proto, test and install targets
├── main.go                      # Plugin entry point
├── internal/
│   ├── config/config.go         # Configuration handling
│   ├── manifest/manifest.go     # plugin.json reader
│   ├── handlers/                # gRPC handlers
│   │   ├── plugin.go
│   │   └── plugin_test.go
│   └── services/                # Business logic
│       ├── plugin.go
│       └── plugin_test.go
└── proto/
    └── plugin.proto             # gRPC service definition
```

`make build` generates `proto/plugin.pb.go` and `proto/plugin_grpc.pb.go` and writes the
binary `./plugin` next to `plugin.json`, where its `binary` field points.

## Plugin Structure

### Service Mode and CLI Mode

The plugin binary runs in one of two modes:

- **Service mode** — `portunix plugin start` launches the binary without a command, with
  `--port N` and the plugin install directory as working directory. The plugin serves the
  gRPC service `portunix.plugins.PluginService` on `127.0.0.1:N`. Portunix then calls
  `Initialize`, `Health` and `Execute`; on `portunix plugin stop` it calls `Shutdown` and
  then sends SIGTERM
- **CLI mode** — `portunix <plugin> <command> [args...]` runs the binary directly with the
  command as positional arguments. The plugin executes the command once, prints the output
  to stdout and exits; no gRPC is involved

Both modes share the same business logic in `internal/services`.

### Main Entry Point (main.go)

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"google.golang.org/grpc"

	"go-plugin-template/internal/config"
	"go-plugin-template/internal/handlers"
	"go-plugin-template/internal/manifest"
	"go-plugin-template/internal/services"
	pb "go-plugin-template/proto"
)

// version is set at build time (-ldflags "-X main.version=...")
var version = "dev"

// Without a command the plugin runs as a gRPC service (portunix plugin start);
// with a command it runs once as a CLI (portunix <plugin> <command> [args...])
func main() {
	var (
		port         = flag.Int("port", 50051, "gRPC server port")
		configPath   = flag.String("config", "config.yaml", "Configuration file path")
		manifestPath = flag.String("manifest", "plugin.json", "Plugin manifest path")
		showVersion  = flag.Bool("version", false, "Print the plugin version and exit")
	)
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fail("failed to load configuration: %v", err)
	}
	configureLogging(cfg.LogLevel)

	if *showVersion {
		fmt.Printf("%s %s (build %s)\n", cfg.Plugin.Name, cfg.Plugin.Version, version)
		return
	}

	if flag.NArg() > 0 {
		runCommand(cfg, flag.Arg(0), flag.Args()[1:])
		return
	}

	// Portunix runs the plugin with the install directory as working
	// directory, so the default manifest path resolves to its plugin.json
	m, err := manifest.Load(*manifestPath)
	if err != nil {
		fail("%v", err)
	}
	if err := serve(cfg, m, *port); err != nil {
		fail("%v", err)
	}
}

// runCommand executes one command and prints its output (CLI mode)
func runCommand(cfg *config.Config, command string, args []string) {
	output, err := services.NewPluginService(cfg).Execute(context.Background(), command, args)
	if err != nil {
		fail("%v", err)
	}
	fmt.Println(output)
}

// serve runs the gRPC server implementing portunix.plugins.PluginService
// until SIGINT or SIGTERM
func serve(cfg *config.Config, m *manifest.Manifest, port int) error {
	// Portunix connects over localhost; the plugin is not exposed to the network
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	server := grpc.NewServer()
	pb.RegisterPluginServiceServer(server, handlers.NewPluginHandler(cfg, m))

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		slog.Info("Shutting down gracefully...")
		server.GracefulStop()
	}()

	slog.Info("Plugin server listening", "plugin", m.Name, "version", m.Version, "address", listener.Addr().String())
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

// configureLogging logs to stderr, so stdout carries only command output
func configureLogging(level string) {
	var l slog.Level
	if err := l.UnmarshalText([]byte(strings.ToUpper(level))); err != nil {
		l = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l})))
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
```

Command-line flags:

| Flag | Default | Description |
| ----- | ----- | ----- |
| `--port` | `50051` | gRPC server port (Portunix passes the port from `plugin.json`) |
| `--config` | `config.yaml` | Configuration file path |
| `--manifest` | `plugin.json` | Plugin manifest path |
| `--version` | `false` | Print the plugin version and exit |

Logging uses `log/slog` and goes to stderr, so stdout carries only command output in CLI mode.

### Configuration (internal/config/config.go)

`config.yaml` is optional; a missing file falls back to the defaults.

```yaml
plugin:
  name: my-go-plugin
  version: 1.0.0
  description: Example Go plugin for Portunix

log_level: info
timeout: 30
```

```go
// Config is the plugin configuration from config.yaml
type Config struct {
	Plugin   PluginConfig `yaml:"plugin"`
	LogLevel string       `yaml:"log_level"`
	Timeout  int          `yaml:"timeout"`
}

// PluginConfig is the plugin section of config.yaml
type PluginConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

// Load reads the configuration; a missing file falls back to the defaults
func Load(path string) (*Config, error) {
	var config Config
	data, err := os.ReadFile(path) // #nosec G304 -- path of the plugin's own config
	switch {
	case errors.Is(err, fs.ErrNotExist):
		config.Plugin = PluginConfig{Name: "my-go-plugin", Version: "1.0.0"}
	case err != nil:
		return nil, fmt.Errorf("failed to read config file: %w", err)
	default:
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Set defaults
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}
```

`Validate` requires `plugin.name` and `plugin.version` and a positive `timeout`. The
`timeout` (seconds) limits each gRPC `Execute` call.

### Plugin Manifest Reader (internal/manifest/manifest.go)

`plugin.json` is the single source of the plugin information and commands reported to
Portunix over gRPC. The `manifest` package parses it and maps it onto the contract
messages:

- `PluginInfo()` — name, version, description, author, license, `dependencies.os_support`,
  commands, capabilities and `permissions` (level defaults to `limited`)
- `PluginCommands()` — the `commands` array as `PluginCommand` messages (parameter type
  defaults to `string`)
- `Capabilities()` — access flags derived from the declared permissions, plus the names of
  the `ai_integration.mcp_tools`

```go
// Capabilities derives the plugin capabilities from the declared permissions
func (m *Manifest) Capabilities() *pb.PluginCapabilities {
	tools := make([]string, 0, len(m.AIIntegration.MCPTools))
	for _, t := range m.AIIntegration.MCPTools {
		tools = append(tools, t.Name)
	}
	return &pb.PluginCapabilities{
		FilesystemAccess: len(m.Permissions.Filesystem) > 0,
		NetworkAccess:    len(m.Permissions.Network) > 0,
		DatabaseAccess:   len(m.Permissions.Database) > 0,
		ContainerAccess:  false,
		SystemCommands:   len(m.Permissions.System) > 0,
		McpTools:         tools,
	}
}
```

### gRPC Handlers (internal/handlers/plugin.go)

The handlers implement `portunix.plugins.PluginService` — the service Portunix calls when
it starts the plugin as a service (`portunix plugin start`): `Initialize`, `GetInfo`,
`Health`, `Execute`, `ListCommands` and `Shutdown`. `proto/plugin.proto` is a copy of the
`PluginService` part of the Portunix contract (`src/app/plugins/proto/plugin.proto`); keep
its package, service and messages unchanged, otherwise Portunix fails with `Unimplemented`
on `Initialize`. Only the language options (`go_package`) may differ.

Two rules matter:

- **`Execute` reports command errors in the response** (`success=false`, `error`,
  `exit_code=1`), not as gRPC errors, so Portunix can show them to the user
- **`Shutdown` only cleans up and answers — it must not exit.** Portunix sends SIGTERM
  right after the call (handled in `main.go` with `GracefulStop`); a plugin that exits
  inside `Shutdown` drops the response and `portunix plugin stop` fails with `EOF`

```go
// PluginHandler implements portunix.plugins.PluginService; plugin info and
// commands come from plugin.json, so the manifest stays the single source
type PluginHandler struct {
	pb.UnimplementedPluginServiceServer
	config    *config.Config
	manifest  *manifest.Manifest
	service   *services.PluginService
	startedAt time.Time
}

// NewPluginHandler creates the gRPC handler
func NewPluginHandler(cfg *config.Config, m *manifest.Manifest) *PluginHandler {
	return &PluginHandler{
		config:    cfg,
		manifest:  m,
		service:   services.NewPluginService(cfg),
		startedAt: time.Now(),
	}
}

// Initialize is the first call Portunix makes after starting the plugin
func (h *PluginHandler) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	slog.Info("Initialize called", "plugin", req.GetPluginName(), "version", req.GetVersion())
	return &pb.InitializeResponse{
		Success:    true,
		Message:    "initialized",
		PluginInfo: h.manifest.PluginInfo(),
	}, nil
}

// GetInfo returns the plugin information from plugin.json
func (h *PluginHandler) GetInfo(ctx context.Context, req *pb.GetInfoRequest) (*pb.GetInfoResponse, error) {
	return &pb.GetInfoResponse{PluginInfo: h.manifest.PluginInfo()}, nil
}

// Health reports the plugin state for portunix plugin health
func (h *PluginHandler) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	healthy := h.service.CheckHealth()
	status, message := "running", "ok"
	if !healthy {
		status, message = "unhealthy", "health check failed"
	}
	return &pb.HealthResponse{
		Healthy:       healthy,
		Status:        status,
		Message:       message,
		UptimeSeconds: int64(time.Since(h.startedAt).Seconds()),
		Metrics:       map[string]string{},
	}, nil
}

// Execute runs a plugin command; command errors are reported in the
// response, not as gRPC errors, so Portunix can show them to the user
func (h *PluginHandler) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	slog.Info("Executing command", "command", req.GetCommand(), "args", req.GetArgs())

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.Timeout)*time.Second)
	defer cancel()

	output, err := h.service.Execute(timeoutCtx, req.GetCommand(), req.GetArgs())
	if err != nil {
		slog.Error("Command execution failed", "error", err)
		return &pb.ExecuteResponse{
			Success:  false,
			Message:  err.Error(),
			Error:    err.Error(),
			ExitCode: 1,
		}, nil
	}
	return &pb.ExecuteResponse{Success: true, Output: output, ExitCode: 0}, nil
}

// ListCommands returns the commands declared in plugin.json
func (h *PluginHandler) ListCommands(ctx context.Context, req *pb.ListCommandsRequest) (*pb.ListCommandsResponse, error) {
	return &pb.ListCommandsResponse{
		Commands:     h.manifest.PluginCommands(),
		Capabilities: h.manifest.Capabilities(),
	}, nil
}

// Shutdown only cleans up and answers; Portunix stops the process right after
// this call (SIGTERM, handled in main.go), exiting here would drop the response
func (h *PluginHandler) Shutdown(ctx context.Context, req *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	slog.Info("Shutdown requested")
	h.service.Cleanup()
	return &pb.ShutdownResponse{Success: true, Message: "cleanup done"}, nil
}
```

### Business Logic (internal/services/plugin.go)

`PluginService` holds the command implementations, shared by the gRPC handler and the CLI
mode. The template provides `hello`, `echo` and `uppercase`:

```go
// Execute runs a plugin command
func (s *PluginService) Execute(ctx context.Context, command string, args []string) (string, error) {
	slog.Debug("Executing command", "command", command, "args", args)

	switch command {
	case "hello":
		return s.handleHello(args)
	case "echo":
		return s.handleEcho(args)
	case "uppercase":
		return s.handleUppercase(args)
	default:
		return "", fmt.Errorf("unknown command: %s", command)
	}
}

func (s *PluginService) handleHello(args []string) (string, error) {
	name := "World"
	if len(args) > 0 && args[0] != "" {
		name = args[0]
	}
	return fmt.Sprintf("Hello, %s! This is %s v%s", name, s.config.Plugin.Name, s.config.Plugin.Version), nil
}

func (s *PluginService) handleEcho(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("echo command requires at least one argument")
	}
	return strings.Join(args, " "), nil
}

func (s *PluginService) handleUppercase(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("uppercase command requires at least one argument")
	}
	return strings.ToUpper(strings.Join(args, " ")), nil
}
```

`CheckHealth()` backs the `Health` RPC (always healthy in the template; check database
connections, external services or internal state here), and `Cleanup()` backs `Shutdown`
(close connections, remove temporary files, cancel background work). When you add a
command, declare it in the `commands` array of `plugin.json` as well, so Portunix lists it.

## Building and Testing

### Makefile Targets

| Target | Description |
| ----- | ----- |
| `make build` | Run `make proto`, then build `./plugin` with the version from `git describe` |
| `make proto` | Generate `proto/*.pb.go` with `protoc` (`paths=source_relative`) |
| `make test` | Run `go test -race` with a coverage report (`coverage.html`) |
| `make dev` | Run the plugin with `go run . --config=config.yaml` |
| `make package` | Build and pack `plugin`, `plugin.json` and `config.yaml` into `dist/` |
| `make install` | Build and run `portunix plugin install .` |
| `make validate` | Run `portunix plugin validate .` |
| `make clean` | Remove the binary, generated protobuf files and coverage output |

`make proto` runs:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	proto/*.proto
```

### Tests

The template has unit tests next to the code: `internal/services/plugin_test.go` covers the
commands, `internal/handlers/plugin_test.go` covers the gRPC handlers against the real
`plugin.json`.

```bash
make test
```

#### Test Example (internal/handlers/plugin_test.go)

```go
func newHandler(t *testing.T) *PluginHandler {
	t.Helper()
	m, err := manifest.Load(filepath.Join("..", "..", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Plugin: config.PluginConfig{Name: "my-go-plugin", Version: "1.0.0"}, Timeout: 30}
	return NewPluginHandler(cfg, m)
}

func TestExecute(t *testing.T) {
	h := newHandler(t)

	resp, err := h.Execute(context.Background(), &pb.ExecuteRequest{Command: "echo", Args: []string{"hi"}})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.GetSuccess() || resp.GetOutput() != "hi" || resp.GetExitCode() != 0 {
		t.Errorf("echo response = %v", resp)
	}

	// A failing command is reported in the response, not as a gRPC error
	resp, err = h.Execute(context.Background(), &pb.ExecuteRequest{Command: "unknown"})
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}
	if resp.GetSuccess() || resp.GetExitCode() == 0 || resp.GetError() == "" {
		t.Errorf("unknown command response = %v", resp)
	}
}
```

## Development Workflow

```bash
# Install dependencies
make deps

# Build (generates protobuf files first)
make build

# Run a command in CLI mode
./plugin hello Portunix

# Run as a gRPC service on 127.0.0.1:50051
./plugin --port 50051

# Install into Portunix
portunix plugin install .
portunix plugin enable go-plugin-template
portunix plugin start go-plugin-template
portunix plugin health go-plugin-template
portunix go-plugin-template hello Portunix
portunix plugin stop go-plugin-template
```

## Advanced Features

### MCP Tool Exposure

The Portunix contract has no separate tool RPCs. Declare MCP tools in `plugin.json` under
`ai_integration.mcp_tools`; the template reports their names in `capabilities.mcp_tools`
of `Initialize`, `GetInfo` and `ListCommands`, and a tool call reaches the plugin as an
`Execute` request with the tool name as `command`:

```json
{
  "ai_integration": {
    "mcp_tools": [
      { "name": "process_file", "description": "Process a file with custom logic" }
    ]
  }
}
```

```go
// internal/services/plugin.go — handle the tool like any other command
case "process_file":
	if len(args) == 0 {
		return "", fmt.Errorf("process_file requires a file path")
	}
	return fmt.Sprintf("Processed file: %s", args[0]), nil
```

### Error Handling

Return errors from `PluginService` with context (`fmt.Errorf("...: %w", err)`). The
handler turns every error into an `ExecuteResponse` with `success=false`, so do not return
gRPC status errors from `Execute`; reserve those for transport-level failures. In CLI mode
`main.go` prints the error to stderr and exits with code 1.

## Deployment

### Plugin Manifest

Go plugins use the `native` runtime. `binary` is the path of the executable relative to
the plugin directory:

```json
{
  "plugin": {
    "type": "grpc",
    "binary": "./plugin",
    "runtime": "native",
    "port": 9001,
    "health_check_interval": 30000000000
  }
}
```

- Portunix starts the plugin as `<install_path>/plugin --port <port>` with the install
  directory as working directory, so the default `--manifest plugin.json` and
  `--config config.yaml` resolve to the installed files
- The `commands` array is what `GetInfo` and `ListCommands` report; keep it in sync with
  `PluginService.Execute`
- `permissions` drive the reported `capabilities` and `required_permissions`

See `template/plugin.json` for the complete manifest.

### Containers

The plugin binds `127.0.0.1` only, because Portunix connects over localhost. Running it in
a container would require binding all interfaces, which the template deliberately does not
do; the template ships no Dockerfile. Distribute the plugin with `make package` instead.

## Troubleshooting

1. **`Unimplemented` on `Initialize`** — `proto/plugin.proto` diverged from the Portunix
   contract; restore the `portunix.plugins` package, service and messages and run `make proto`
2. **`portunix plugin stop` fails with `EOF`** — `Shutdown` exits the process; let it only
   clean up and answer, Portunix sends SIGTERM afterwards
3. **`failed to read manifest`** — the plugin was started outside its directory; pass
   `--manifest <path>/plugin.json`
4. **Build failures** — run `make deps`, check that `protoc-gen-go` and `protoc-gen-go-grpc`
   are on `PATH` and regenerate with `make proto`

## Next Steps

- Study the [template code](template/) for a complete example
- Learn about [MCP integration](../../mcp-integration/exposing-tools.md) for AI agents
- Review the [plugin architecture](../../architecture.md)

## Resources

- [Go gRPC Documentation](https://grpc.io/docs/languages/go/)
- [Protocol Buffers Go Tutorial](https://protobuf.dev/getting-started/gotutorial/)
- [log/slog package](https://pkg.go.dev/log/slog)
