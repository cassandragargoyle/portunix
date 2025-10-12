# Go Plugin Development Guide

Go is the native language for Portunix plugins, offering the best performance and seamless integration with the core system. This guide will help you create robust, efficient plugins using Go.

## Prerequisites

- Go 1.19 or later
- Protocol Buffers compiler (`protoc`)
- Portunix development environment

## Quick Start

Create a new Go plugin using the Portunix CLI:

```bash
portunix plugin create my-go-plugin --language=go
cd my-go-plugin
```

This creates a complete Go project structure:

```
my-go-plugin/
├── plugin.yaml              # Plugin manifest
├── go.mod                   # Go module definition
├── go.sum                   # Go dependencies
├── main.go                  # Plugin entry point
├── internal/                # Internal packages
│   ├── config/             # Configuration handling
│   ├── handlers/           # gRPC handlers
│   └── services/           # Business logic
├── proto/                  # Protocol Buffer definitions
├── scripts/                # Build and deployment scripts
├── tests/                  # Test files
└── README.md               # Plugin documentation
```

## Plugin Structure

### Main Entry Point (main.go)
```go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"

    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
    
    pb "my-go-plugin/proto"
    "my-go-plugin/internal/config"
    "my-go-plugin/internal/handlers"
)

func main() {
    var (
        port       = flag.Int("port", 50051, "gRPC server port")
        configPath = flag.String("config", "config.yaml", "Configuration file path")
    )
    flag.Parse()

    // Load configuration
    cfg, err := config.Load(*configPath)
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Create gRPC server
    server := grpc.NewServer()
    
    // Register plugin service
    pluginHandler := handlers.NewPluginHandler(cfg)
    pb.RegisterPluginServiceServer(server, pluginHandler)
    
    // Register health service
    healthServer := health.NewServer()
    grpc_health_v1.RegisterHealthServer(server, healthServer)
    healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

    // Start server
    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    // Graceful shutdown
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
        <-sigChan
        
        log.Println("Shutting down gracefully...")
        server.GracefulStop()
    }()

    log.Printf("Plugin server listening on port %d", *port)
    if err := server.Serve(listener); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
```

### Configuration Package (internal/config/config.go)
```go
package config

import (
    "io/ioutil"
    "gopkg.in/yaml.v3"
)

type Config struct {
    Plugin   PluginConfig   `yaml:"plugin"`
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database,omitempty"`
    API      APIConfig      `yaml:"api,omitempty"`
}

type PluginConfig struct {
    Name        string `yaml:"name"`
    Version     string `yaml:"version"`
    Description string `yaml:"description"`
    LogLevel    string `yaml:"log_level" default:"info"`
}

type ServerConfig struct {
    Port        int    `yaml:"port" default:"50051"`
    HealthPort  int    `yaml:"health_port" default:"50052"`
    MetricsPort int    `yaml:"metrics_port" default:"8080"`
}

type DatabaseConfig struct {
    URL      string `yaml:"url"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

type APIConfig struct {
    Endpoints []Endpoint `yaml:"endpoints"`
}

type Endpoint struct {
    Path    string `yaml:"path"`
    Method  string `yaml:"method"`
    Handler string `yaml:"handler"`
}

func Load(path string) (*Config, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    return &config, nil
}

func (c *Config) Validate() error {
    if c.Plugin.Name == "" {
        return fmt.Errorf("plugin name is required")
    }
    if c.Plugin.Version == "" {
        return fmt.Errorf("plugin version is required")
    }
    return nil
}
```

### gRPC Handlers (internal/handlers/plugin.go)
```go
package handlers

import (
    "context"
    "fmt"

    pb "my-go-plugin/proto"
    "my-go-plugin/internal/config"
    "my-go-plugin/internal/services"
)

type PluginHandler struct {
    pb.UnimplementedPluginServiceServer
    config  *config.Config
    service *services.PluginService
}

func NewPluginHandler(cfg *config.Config) *PluginHandler {
    return &PluginHandler{
        config:  cfg,
        service: services.NewPluginService(cfg),
    }
}

func (h *PluginHandler) GetInfo(ctx context.Context, req *pb.GetInfoRequest) (*pb.GetInfoResponse, error) {
    return &pb.GetInfoResponse{
        Info: &pb.PluginInfo{
            Name:         h.config.Plugin.Name,
            Version:      h.config.Plugin.Version,
            Description:  h.config.Plugin.Description,
            Capabilities: []string{"example-capability"},
        },
    }, nil
}

func (h *PluginHandler) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
    status := h.service.CheckHealth()
    return &pb.HealthCheckResponse{
        Status: pb.HealthCheckResponse_Status(status),
    }, nil
}

func (h *PluginHandler) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
    result, err := h.service.Execute(ctx, req.Command, req.Args)
    if err != nil {
        return nil, fmt.Errorf("execution failed: %w", err)
    }

    return &pb.ExecuteResponse{
        Result: result,
        Status: pb.ExecuteResponse_SUCCESS,
    }, nil
}

func (h *PluginHandler) Shutdown(ctx context.Context, req *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
    // Perform cleanup operations
    h.service.Cleanup()
    
    return &pb.ShutdownResponse{
        Status: pb.ShutdownResponse_SUCCESS,
    }, nil
}
```

### Business Logic (internal/services/plugin.go)
```go
package services

import (
    "context"
    "fmt"
    "log"

    "my-go-plugin/internal/config"
)

type PluginService struct {
    config *config.Config
}

func NewPluginService(cfg *config.Config) *PluginService {
    return &PluginService{
        config: cfg,
    }
}

func (s *PluginService) Execute(ctx context.Context, command string, args []string) (string, error) {
    switch command {
    case "hello":
        return s.handleHello(args)
    case "process":
        return s.handleProcess(args)
    default:
        return "", fmt.Errorf("unknown command: %s", command)
    }
}

func (s *PluginService) handleHello(args []string) (string, error) {
    name := "World"
    if len(args) > 0 {
        name = args[0]
    }
    return fmt.Sprintf("Hello, %s!", name), nil
}

func (s *PluginService) handleProcess(args []string) (string, error) {
    if len(args) == 0 {
        return "", fmt.Errorf("process command requires at least one argument")
    }
    
    // Implement your processing logic here
    result := fmt.Sprintf("Processed: %v", args)
    return result, nil
}

func (s *PluginService) CheckHealth() int32 {
    // Implement health check logic
    // Return appropriate status code
    return 1 // SERVING
}

func (s *PluginService) Cleanup() {
    log.Println("Performing cleanup operations...")
    // Implement cleanup logic
}
```

## Building and Testing

### Build Script (scripts/build.sh)
```bash
#!/bin/bash

set -e

# Build plugin binary
echo "Building plugin..."
go build -o bin/plugin main.go

# Generate protocol buffers (if needed)
if [ -d "proto" ]; then
    echo "Generating protocol buffers..."
    protoc --go_out=. --go-grpc_out=. proto/*.proto
fi

# Run tests
echo "Running tests..."
go test ./...

# Create plugin package
echo "Creating plugin package..."
mkdir -p dist
tar -czf dist/plugin.tar.gz bin/ plugin.yaml

echo "Build completed successfully!"
```

### Test Example (tests/plugin_test.go)
```go
package tests

import (
    "context"
    "testing"
    
    "my-go-plugin/internal/config"
    "my-go-plugin/internal/handlers"
    pb "my-go-plugin/proto"
)

func TestPluginInfo(t *testing.T) {
    cfg := &config.Config{
        Plugin: config.PluginConfig{
            Name:        "test-plugin",
            Version:     "1.0.0",
            Description: "Test plugin",
        },
    }
    
    handler := handlers.NewPluginHandler(cfg)
    
    resp, err := handler.GetInfo(context.Background(), &pb.GetInfoRequest{})
    if err != nil {
        t.Fatalf("GetInfo failed: %v", err)
    }
    
    if resp.Info.Name != cfg.Plugin.Name {
        t.Errorf("Expected name %s, got %s", cfg.Plugin.Name, resp.Info.Name)
    }
}

func TestHealthCheck(t *testing.T) {
    cfg := &config.Config{}
    handler := handlers.NewPluginHandler(cfg)
    
    resp, err := handler.HealthCheck(context.Background(), &pb.HealthCheckRequest{})
    if err != nil {
        t.Fatalf("HealthCheck failed: %v", err)
    }
    
    if resp.Status != pb.HealthCheckResponse_SERVING {
        t.Errorf("Expected SERVING status, got %v", resp.Status)
    }
}
```

## Development Workflow

### 1. Development Setup
```bash
# Install dependencies
go mod download

# Generate protocol buffers
make proto

# Run in development mode
go run main.go -config=config.dev.yaml
```

### 2. Testing
```bash
# Run unit tests
go test ./...

# Run integration tests
go test -tags=integration ./tests/integration/

# Test with Portunix
portunix plugin install .
portunix plugin test my-go-plugin
```

### 3. Building
```bash
# Build binary
make build

# Build for multiple platforms
make build-all

# Create distribution package
make package
```

## Advanced Features

### MCP Integration
For plugins that expose tools to AI agents:

```go
// internal/handlers/mcp.go
package handlers

import (
    "context"
    "encoding/json"
    
    pb "my-go-plugin/proto"
)

func (h *PluginHandler) ListTools(ctx context.Context, req *pb.ListToolsRequest) (*pb.ListToolsResponse, error) {
    tools := []*pb.MCPTool{
        {
            Name:        "process_file",
            Description: "Process a file with custom logic",
            Schema: `{
                "type": "object",
                "properties": {
                    "file_path": {"type": "string"},
                    "options": {"type": "object"}
                },
                "required": ["file_path"]
            }`,
        },
    }
    
    return &pb.ListToolsResponse{Tools: tools}, nil
}

func (h *PluginHandler) CallTool(ctx context.Context, req *pb.CallToolRequest) (*pb.CallToolResponse, error) {
    switch req.ToolName {
    case "process_file":
        return h.handleProcessFile(ctx, req.Arguments)
    default:
        return nil, fmt.Errorf("unknown tool: %s", req.ToolName)
    }
}

func (h *PluginHandler) handleProcessFile(ctx context.Context, args string) (*pb.CallToolResponse, error) {
    var params struct {
        FilePath string                 `json:"file_path"`
        Options  map[string]interface{} `json:"options"`
    }
    
    if err := json.Unmarshal([]byte(args), &params); err != nil {
        return nil, fmt.Errorf("invalid arguments: %w", err)
    }
    
    // Process the file
    result := fmt.Sprintf("Processed file: %s", params.FilePath)
    
    return &pb.CallToolResponse{
        Result: result,
        Status: pb.CallToolResponse_SUCCESS,
    }, nil
}
```

### Configuration Validation
```go
func (c *Config) Validate() error {
    if c.Plugin.Name == "" {
        return fmt.Errorf("plugin name is required")
    }
    
    if c.Server.Port <= 0 || c.Server.Port > 65535 {
        return fmt.Errorf("invalid server port: %d", c.Server.Port)
    }
    
    // Validate API endpoints
    for _, endpoint := range c.API.Endpoints {
        if endpoint.Path == "" {
            return fmt.Errorf("endpoint path cannot be empty")
        }
        if endpoint.Method == "" {
            return fmt.Errorf("endpoint method cannot be empty")
        }
    }
    
    return nil
}
```

### Logging and Metrics
```go
import (
    "github.com/sirupsen/logrus"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    requestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "plugin_requests_total",
            Help: "Total number of requests processed",
        },
        []string{"method", "status"},
    )
    
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "plugin_request_duration_seconds",
            Help: "Duration of requests",
        },
        []string{"method"},
    )
)

func (h *PluginHandler) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
    timer := prometheus.NewTimer(requestDuration.WithLabelValues("execute"))
    defer timer.ObserveDuration()
    
    logrus.WithFields(logrus.Fields{
        "command": req.Command,
        "args":    req.Args,
    }).Info("Executing command")
    
    result, err := h.service.Execute(ctx, req.Command, req.Args)
    
    status := "success"
    if err != nil {
        status = "error"
        logrus.WithError(err).Error("Command execution failed")
    }
    
    requestsTotal.WithLabelValues("execute", status).Inc()
    
    if err != nil {
        return nil, err
    }
    
    return &pb.ExecuteResponse{
        Result: result,
        Status: pb.ExecuteResponse_SUCCESS,
    }, nil
}
```

## Best Practices

### Error Handling
```go
import (
    "errors"
    "fmt"
)

var (
    ErrInvalidConfiguration = errors.New("invalid configuration")
    ErrServiceUnavailable   = errors.New("service unavailable")
    ErrInvalidInput        = errors.New("invalid input")
)

func (s *PluginService) ProcessData(data []byte) error {
    if len(data) == 0 {
        return fmt.Errorf("%w: data cannot be empty", ErrInvalidInput)
    }
    
    // Process data...
    
    return nil
}
```

### Resource Management
```go
type PluginService struct {
    config   *config.Config
    db       *sql.DB
    clients  map[string]*http.Client
    shutdown chan struct{}
}

func (s *PluginService) Start() error {
    s.shutdown = make(chan struct{})
    
    // Initialize resources
    if err := s.initDatabase(); err != nil {
        return fmt.Errorf("failed to initialize database: %w", err)
    }
    
    return nil
}

func (s *PluginService) Stop() error {
    close(s.shutdown)
    
    // Cleanup resources
    if s.db != nil {
        s.db.Close()
    }
    
    for _, client := range s.clients {
        client.CloseIdleConnections()
    }
    
    return nil
}
```

### Testing Patterns
```go
func TestPluginServiceWithMocks(t *testing.T) {
    // Create mock dependencies
    mockDB := &MockDatabase{}
    mockClient := &MockHTTPClient{}
    
    service := &PluginService{
        db:     mockDB,
        client: mockClient,
    }
    
    // Set up expectations
    mockDB.On("Query", mock.Anything).Return([]byte("result"), nil)
    
    // Execute test
    result, err := service.ProcessData([]byte("input"))
    
    // Verify results
    assert.NoError(t, err)
    assert.Equal(t, "expected result", result)
    
    // Verify mock calls
    mockDB.AssertExpectations(t)
}
```

## Deployment

### Dockerfile
```dockerfile
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o plugin main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/plugin .
COPY --from=builder /app/plugin.yaml .

EXPOSE 50051 50052 8080

CMD ["./plugin"]
```

### Plugin Manifest
```yaml
name: my-go-plugin
version: 1.0.0
description: Example Go plugin for Portunix
author: Developer Name
license: MIT

type: service
capabilities:
  - data-processing
  - file-handling

runtime:
  language: go
  main: ./plugin
  port: 50051
  health_check_port: 50052
  metrics_port: 8080

dependencies:
  system:
    - docker: ">=20.0.0"

permissions:
  filesystem:
    read: ["/tmp", "/var/log"]
    write: ["/tmp/plugin-data"]
  network:
    outbound: ["api.example.com:443"]

configuration:
  required:
    database_url:
      type: string
      description: Database connection URL
  optional:
    timeout:
      type: integer
      default: 30
      description: Request timeout in seconds
```

## Troubleshooting

### Common Issues

1. **gRPC Connection Failed**
   - Verify port configuration
   - Check firewall settings
   - Ensure plugin implements required interfaces

2. **Configuration Errors**
   - Validate YAML syntax
   - Check required fields
   - Verify file permissions

3. **Build Failures**
   - Update Go dependencies: `go mod tidy`
   - Regenerate protobuf files: `make proto`
   - Check Go version compatibility

## Next Steps

- Study the [template code](template/) for a complete example
- Read [best practices](best-practices.md) for production deployment
- Explore [examples](examples/) for specific use cases
- Learn about [MCP integration](../../mcp-integration/exposing-tools.md) for AI agents

## Resources

- [Go gRPC Documentation](https://grpc.io/docs/languages/go/)
- [Protocol Buffers Go Tutorial](https://developers.google.com/protocol-buffers/docs/gotutorial)
- [Portunix Plugin API Reference](../../api-reference.md)