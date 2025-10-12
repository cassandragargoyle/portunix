package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

// Factory creates configured logger instances
type Factory struct {
	config *Config
}

// NewFactory creates a new logger factory with the given configuration
func NewFactory(config *Config) *Factory {
	if config == nil {
		config = DefaultConfig()
	}
	return &Factory{config: config}
}

// CreateLogger creates a new logger with the specified component name
func (f *Factory) CreateLogger(component string) Logger {
	// Parse log level
	level := f.parseLevel(f.config.Level)

	// Create output writer based on configuration
	writer := f.createWriter()

	// Create zerolog logger
	zlog := zerolog.New(writer).
		With().
		Timestamp().
		Str("component", component)

	// Add caller information if not in production
	if f.config.Level == "debug" || f.config.Level == "trace" {
		zlog = zlog.Caller()
	}

	// Add platform information
	zlog = zlog.
		Str("os", runtime.GOOS).
		Str("arch", runtime.GOARCH)

	// Check for container environment
	if isInContainer() {
		zlog = zlog.Bool("container", true)
		if containerID := getContainerID(); containerID != "" {
			zlog = zlog.Str("container_id", containerID)
		}
	}

	logger := zlog.Logger().Level(level)

	return &PortunixLogger{
		logger: logger,
		level:  level,
	}
}

// CreateMCPLogger creates a special logger for MCP server mode
func (f *Factory) CreateMCPLogger(component string) Logger {
	config := f.config.Clone()

	// MCP mode modifications
	config.Output = []string{"file"}
	if config.FilePath == "" {
		config.FilePath = "/var/log/portunix/mcp.log"
	}
	if config.Level == "debug" || config.Level == "trace" {
		config.Level = "warn" // Reduce verbosity for MCP
	}

	// Create new factory with modified config
	mcpFactory := NewFactory(config)
	return mcpFactory.CreateLogger(component)
}

// CreateTestLogger creates a logger optimized for testing
func (f *Factory) CreateTestLogger(t *testing.T) Logger {
	// Check if we're in verbose mode
	verbose := testing.Verbose()

	config := f.config.Clone()
	if verbose {
		config.Level = "debug"
		config.Format = "text"
	} else {
		config.Level = "error"
	}

	// Create factory with test config
	testFactory := NewFactory(config)
	logger := testFactory.CreateLogger(t.Name())

	return logger
}

// parseLevel converts string level to zerolog.Level
func (f *Factory) parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// createWriter creates the appropriate output writer based on configuration
func (f *Factory) createWriter() io.Writer {
	var writers []io.Writer

	for _, output := range f.config.Output {
		switch strings.ToLower(output) {
		case "console", "stdout":
			writers = append(writers, f.createConsoleWriter(os.Stdout))
		case "stderr":
			writers = append(writers, f.createConsoleWriter(os.Stderr))
		case "file":
			if w := f.createFileWriter(); w != nil {
				writers = append(writers, w)
			}
		case "syslog":
			if w := f.createSyslogWriter(); w != nil {
				writers = append(writers, w)
			}
		}
	}

	if len(writers) == 0 {
		// Default to console if no valid writers
		writers = append(writers, f.createConsoleWriter(os.Stderr))
	}

	if len(writers) == 1 {
		return writers[0]
	}

	return zerolog.MultiLevelWriter(writers...)
}

// createConsoleWriter creates a console writer
func (f *Factory) createConsoleWriter(out io.Writer) io.Writer {
	if f.config.Format == "json" {
		return out
	}

	// Text format with pretty console output
	return zerolog.ConsoleWriter{
		Out:        out,
		TimeFormat: f.config.TimeFormat,
		NoColor:    f.config.NoColor || !isTerminal(out),
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			"component",
			zerolog.MessageFieldName,
		},
		FieldsExclude: []string{
			"os", "arch", // Exclude platform info from console
		},
	}
}

// createFileWriter creates a file writer
func (f *Factory) createFileWriter() io.Writer {
	if f.config.FilePath == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(f.config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory %s: %v\n", dir, err)
		return nil
	}

	// Open or create file
	file, err := os.OpenFile(f.config.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v\n", f.config.FilePath, err)
		return nil
	}

	// For file output, always use JSON format
	return file
}

// createSyslogWriter creates a syslog writer (placeholder for now)
func (f *Factory) createSyslogWriter() io.Writer {
	// TODO: Implement syslog support
	// This requires platform-specific implementation
	return nil
}

// isTerminal checks if the writer is a terminal
func isTerminal(w io.Writer) bool {
	// Simple check - can be enhanced with terminal detection libraries
	if f, ok := w.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return false
		}
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// isInContainer detects if running inside a container
func isInContainer() bool {
	// Check for Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check for containerd/k8s
	if _, err := os.Stat("/run/secrets/kubernetes.io"); err == nil {
		return true
	}

	// Check cgroup for container signatures
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(data)
		if strings.Contains(content, "docker") ||
		   strings.Contains(content, "containerd") ||
		   strings.Contains(content, "kubepods") {
			return true
		}
	}

	return false
}

// getContainerID attempts to get the container ID
func getContainerID() string {
	// Try to extract from /proc/self/cgroup
	if data, err := os.ReadFile("/proc/self/cgroup"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "/")
			if len(parts) > 0 {
				// Last part often contains container ID
				id := parts[len(parts)-1]
				if len(id) == 64 { // Docker container IDs are 64 chars
					return id[:12] // Return short form
				}
			}
		}
	}
	return ""
}

// DetectEnvironment returns information about the runtime environment
func DetectEnvironment() map[string]string {
	env := make(map[string]string)

	env["os"] = runtime.GOOS
	env["arch"] = runtime.GOARCH
	env["go_version"] = runtime.Version()

	if isInContainer() {
		env["container"] = "true"
		if id := getContainerID(); id != "" {
			env["container_id"] = id
		}
	}

	// Check for common CI environments
	if os.Getenv("CI") != "" {
		env["ci"] = "true"
	}
	if os.Getenv("GITHUB_ACTIONS") != "" {
		env["ci_platform"] = "github_actions"
	}

	return env
}