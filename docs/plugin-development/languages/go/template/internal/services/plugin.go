package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"go-plugin-template/internal/config"
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
	logrus.WithFields(logrus.Fields{
		"command": command,
		"args":    args,
	}).Debug("Executing command")

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

func (s *PluginService) ProcessExample(ctx context.Context, input string) (string, error) {
	logrus.WithField("input", input).Debug("Processing example input")

	if input == "" {
		return "", fmt.Errorf("input cannot be empty")
	}

	// Example processing: reverse the string
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	result := string(runes)
	logrus.WithField("result", result).Debug("Processing completed")

	return fmt.Sprintf("Processed by %s: %s", s.config.Plugin.Name, result), nil
}

func (s *PluginService) CheckHealth() int32 {
	// Implement health check logic here
	// For this example, we'll always return healthy
	// In a real plugin, you might check:
	// - Database connections
	// - External service availability
	// - Resource usage
	// - Internal state consistency

	logrus.Debug("Health check performed")
	return 1 // SERVING (corresponds to grpc_health_v1.HealthCheckResponse_SERVING)
}

func (s *PluginService) Cleanup() {
	logrus.Info("Performing cleanup operations")

	// Implement cleanup logic here:
	// - Close database connections
	// - Clean up temporary files
	// - Cancel background operations
	// - Release resources

	logrus.Info("Cleanup completed")
}