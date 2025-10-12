package logging

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestNewFactory(t *testing.T) {
	// Test with nil config
	factory := NewFactory(nil)
	if factory == nil {
		t.Fatal("NewFactory(nil) returned nil")
	}
	if factory.config == nil {
		t.Fatal("Factory config is nil")
	}

	// Test with custom config
	config := &Config{
		Level:  "debug",
		Format: "json",
		Output: []string{"console"},
	}
	factory = NewFactory(config)
	if factory.config != config {
		t.Error("Factory should use provided config")
	}
}

func TestCreateLogger(t *testing.T) {
	config := DefaultConfig()
	factory := NewFactory(config)

	logger := factory.CreateLogger("test-component")
	if logger == nil {
		t.Fatal("CreateLogger returned nil")
	}

	// Test that logger works
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test-component") {
		t.Errorf("Expected component name in output, got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected message in output, got: %s", output)
	}
}

func TestCreateMCPLogger(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	config := DefaultConfig()
	config.FilePath = filepath.Join(tempDir, "mcp.log")
	factory := NewFactory(config)

	logger := factory.CreateMCPLogger("mcp-server")
	if logger == nil {
		t.Fatal("CreateMCPLogger returned nil")
	}

	// Test that MCP logger logs to file
	logger.Warn("mcp server started")

	// Check if file was created
	if _, err := os.Stat(config.FilePath); os.IsNotExist(err) {
		t.Error("MCP log file was not created")
	}
}

func TestCreateTestLogger(t *testing.T) {
	config := DefaultConfig()
	factory := NewFactory(config)

	logger := factory.CreateTestLogger(t)
	if logger == nil {
		t.Fatal("CreateTestLogger returned nil")
	}

	// Test logger functionality
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.Error("test error")

	output := buf.String()
	if !strings.Contains(output, "test error") {
		t.Errorf("Expected error message in output, got: %s", output)
	}
}

func TestParseLevel(t *testing.T) {
	factory := NewFactory(DefaultConfig())

	testCases := []struct {
		input    string
		expected zerolog.Level
	}{
		{"trace", zerolog.TraceLevel},
		{"debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"warning", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"fatal", zerolog.FatalLevel},
		{"panic", zerolog.PanicLevel},
		{"invalid", zerolog.InfoLevel}, // Should default to info
		{"", zerolog.InfoLevel},        // Should default to info
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := factory.parseLevel(tc.input)
			if result != tc.expected {
				t.Errorf("parseLevel(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestCreateFileWriter(t *testing.T) {
	// Test with valid path
	tempDir := t.TempDir()
	config := &Config{
		FilePath: filepath.Join(tempDir, "test.log"),
	}
	factory := NewFactory(config)

	writer := factory.createFileWriter()
	if writer == nil {
		t.Error("createFileWriter returned nil for valid path")
	}

	// Test file creation
	if _, err := os.Stat(config.FilePath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// Test with empty path
	config.FilePath = ""
	writer = factory.createFileWriter()
	if writer != nil {
		t.Error("createFileWriter should return nil for empty path")
	}
}

func TestCreateConsoleWriter(t *testing.T) {
	// Test text format
	config := &Config{
		Format:     "text",
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
	}
	factory := NewFactory(config)

	var buf bytes.Buffer
	writer := factory.createConsoleWriter(&buf)
	if writer == nil {
		t.Error("createConsoleWriter returned nil")
	}

	// Test JSON format
	config.Format = "json"
	writer = factory.createConsoleWriter(&buf)
	if writer != &buf {
		t.Error("JSON format should return the original writer")
	}
}

func TestContainerDetection(t *testing.T) {
	// Test container detection (these are system-dependent)
	// We mainly test that functions don't panic

	inContainer := isInContainer()
	// Should not panic and return a boolean
	_ = inContainer

	containerID := getContainerID()
	// Should not panic and return a string (may be empty)
	_ = containerID

	// Test DetectEnvironment
	env := DetectEnvironment()
	if env == nil {
		t.Error("DetectEnvironment returned nil")
	}

	// Should contain basic environment info
	if env["os"] != runtime.GOOS {
		t.Errorf("Expected os=%s, got %s", runtime.GOOS, env["os"])
	}
	if env["arch"] != runtime.GOARCH {
		t.Errorf("Expected arch=%s, got %s", runtime.GOARCH, env["arch"])
	}
}

func TestFactoryWithDebugLevel(t *testing.T) {
	config := &Config{
		Level:      "debug",
		Format:     "text",
		Output:     []string{"console"},
		TimeFormat: "15:04:05",
	}
	factory := NewFactory(config)

	logger := factory.CreateLogger("debug-test")

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	// Set global level to debug to ensure debug messages are shown
	originalLevel := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	defer zerolog.SetGlobalLevel(originalLevel)

	logger.Debug("debug message")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("Debug message not found in output: %s", output)
	}
}

func TestFactoryWithMultipleOutputs(t *testing.T) {
	tempDir := t.TempDir()

	config := &Config{
		Level:    "info",
		Format:   "json",
		Output:   []string{"console", "file"},
		FilePath: filepath.Join(tempDir, "multi.log"),
	}
	factory := NewFactory(config)

	logger := factory.CreateLogger("multi-output")
	logger.Info("multi output test")

	// Check if file was created
	if _, err := os.Stat(config.FilePath); os.IsNotExist(err) {
		t.Error("Log file was not created for multi-output")
	}
}

func TestFactoryPlatformInfo(t *testing.T) {
	config := DefaultConfig()
	factory := NewFactory(config)

	logger := factory.CreateLogger("platform-test")

	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.Info("platform test")

	output := buf.String()
	// Platform info should be included (os, arch)
	if !strings.Contains(output, runtime.GOOS) {
		t.Errorf("Expected OS info in output: %s", output)
	}
	if !strings.Contains(output, runtime.GOARCH) {
		t.Errorf("Expected architecture info in output: %s", output)
	}
}

func BenchmarkCreateLogger(b *testing.B) {
	config := DefaultConfig()
	factory := NewFactory(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger := factory.CreateLogger("benchmark")
		_ = logger
	}
}

func BenchmarkLoggerCreationWithFile(b *testing.B) {
	tempDir := b.TempDir()

	config := &Config{
		Level:    "info",
		Format:   "json",
		Output:   []string{"file"},
		FilePath: filepath.Join(tempDir, "bench.log"),
	}
	factory := NewFactory(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger := factory.CreateLogger("benchmark")
		logger.Info("benchmark message")
	}
}

func TestMCPLoggerFileOutput(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "mcp-test.log")

	config := &Config{
		Level:    "debug",
		Format:   "json",
		Output:   []string{"console", "file"},
		FilePath: logPath,
	}
	factory := NewFactory(config)

	mcpLogger := factory.CreateMCPLogger("mcp-server")

	// Log a test message
	mcpLogger.Warn("MCP server initialization")
	mcpLogger.Error("Test error message")

	// Verify file exists and contains data
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Fatal("MCP log file was not created")
	}

	// Read file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "MCP server initialization") {
		t.Error("Expected message not found in MCP log file")
	}
	if !strings.Contains(contentStr, "mcp-server") {
		t.Error("Expected component name not found in MCP log file")
	}
}