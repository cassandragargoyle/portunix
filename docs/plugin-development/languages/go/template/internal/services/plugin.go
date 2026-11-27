package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"go-plugin-template/internal/config"
)

// PluginService holds the plugin business logic, shared by the gRPC handler
// and the CLI mode
type PluginService struct {
	config *config.Config
}

// NewPluginService creates the business service
func NewPluginService(cfg *config.Config) *PluginService {
	return &PluginService{
		config: cfg,
	}
}

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

// CheckHealth reports whether the plugin can serve requests
func (s *PluginService) CheckHealth() bool {
	// Implement health check logic here
	// For this example, the plugin is always healthy
	// In a real plugin, you might check:
	// - Database connections
	// - External service availability
	// - Resource usage
	// - Internal state consistency

	slog.Debug("Health check performed")
	return true
}

// Cleanup releases resources before the plugin stops
func (s *PluginService) Cleanup() {
	slog.Info("Performing cleanup operations")

	// Implement cleanup logic here:
	// - Close database connections
	// - Clean up temporary files
	// - Cancel background operations
	// - Release resources

	slog.Info("Cleanup completed")
}
