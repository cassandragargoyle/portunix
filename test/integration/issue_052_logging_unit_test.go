package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/pkg/logging"
	"portunix.ai/portunix/test/testframework"
	"github.com/rs/zerolog"
)

func TestIssue052LoggingUnitValidation(t *testing.T) {
	tf := testframework.NewTestFramework("Issue052_Logging_Unit_Validation")
	tf.Start(t, "Comprehensive unit test validation for logging system components")

	success := true
	defer tf.Finish(t, success)

	// Step 1: Test logger factory creation
	tf.Step(t, "Test logger factory creation")
	factory := logging.NewFactory(nil)
	if factory == nil {
		tf.Error(t, "Factory creation failed", "NewFactory(nil) returned nil")
		success = false
		return
	}
	tf.Success(t, "Factory created successfully")

	tf.Separator()

	// Step 2: Test component logger creation
	tf.Step(t, "Test component logger creation")
	logger := factory.CreateLogger("test-component")
	if logger == nil {
		tf.Error(t, "Logger creation failed", "CreateLogger returned nil")
		success = false
		return
	}

	// Test logger output
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.Info("Unit test message", "test_id", "052-001")

	output := buf.String()
	if !strings.Contains(output, "test-component") {
		tf.Error(t, "Component name missing", "Expected 'test-component' in output", output)
		success = false
		return
	}
	if !strings.Contains(output, "Unit test message") {
		tf.Error(t, "Message missing", "Expected 'Unit test message' in output", output)
		success = false
		return
	}
	tf.Success(t, "Component logger working correctly")

	tf.Separator()

	// Step 3: Test log levels
	tf.Step(t, "Test log level functionality")
	testLevels := []struct {
		level    zerolog.Level
		logFunc  func(string, ...interface{})
		levelStr string
	}{
		{zerolog.DebugLevel, logger.Debug, "debug"},
		{zerolog.InfoLevel, logger.Info, "info"},
		{zerolog.WarnLevel, logger.Warn, "warn"},
		{zerolog.ErrorLevel, logger.Error, "error"},
	}

	for _, test := range testLevels {
		buf.Reset()
		logger.SetLevel(test.level)
		test.logFunc("Test message for " + test.levelStr)

		output := buf.String()
		if !strings.Contains(output, test.levelStr) {
			tf.Error(t, "Log level test failed", "Expected level '"+test.levelStr+"' in output", output)
			success = false
			return
		}
	}
	tf.Success(t, "All log levels working correctly")

	tf.Separator()

	// Step 4: Test structured logging with fields
	tf.Step(t, "Test structured logging with fields")
	buf.Reset()
	logger.Info("Structured log test",
		"user_id", "test-user-123",
		"operation", "unit-test",
		"count", 42)

	output = buf.String()
	expectedFields := []string{"user_id", "test-user-123", "operation", "unit-test", "42"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			tf.Error(t, "Structured field missing", "Expected field '"+field+"' in output", output)
			success = false
			return
		}
	}
	tf.Success(t, "Structured logging working correctly")

	tf.Separator()

	// Step 5: Test MCP logger creation
	tf.Step(t, "Test MCP logger creation")
	tempDir := t.TempDir()
	config := &logging.Config{
		Level:    "warn",
		Format:   "json",
		Output:   []string{"file"},
		FilePath: filepath.Join(tempDir, "mcp-test.log"),
	}
	mcpFactory := logging.NewFactory(config)
	mcpLogger := mcpFactory.CreateMCPLogger("mcp-server")

	if mcpLogger == nil {
		tf.Error(t, "MCP logger creation failed", "CreateMCPLogger returned nil")
		success = false
		return
	}

	mcpLogger.Warn("MCP server test message")

	// Verify file was created
	if _, err := os.Stat(config.FilePath); os.IsNotExist(err) {
		tf.Error(t, "MCP log file not created", err.Error())
		success = false
		return
	}
	tf.Success(t, "MCP logger created and working")

	tf.Separator()

	// Step 6: Test configuration validation
	tf.Step(t, "Test configuration validation")
	invalidConfig := &logging.Config{
		Level:  "invalid-level",
		Format: "invalid-format",
		Output: []string{"invalid-output"},
	}

	err := invalidConfig.Validate()
	if err != nil {
		tf.Error(t, "Configuration validation error", err.Error())
		success = false
		return
	}

	// Check if invalid values were corrected
	if invalidConfig.Level != "info" {
		tf.Error(t, "Invalid level not corrected", "Expected 'info', got '"+invalidConfig.Level+"'")
		success = false
		return
	}
	if invalidConfig.Format != "text" {
		tf.Error(t, "Invalid format not corrected", "Expected 'text', got '"+invalidConfig.Format+"'")
		success = false
		return
	}
	if len(invalidConfig.Output) != 1 || invalidConfig.Output[0] != "console" {
		tf.Error(t, "Invalid output not corrected", "Expected ['console'], got", invalidConfig.Output)
		success = false
		return
	}
	tf.Success(t, "Configuration validation working correctly")

	tf.Separator()

	// Step 7: Test environment variable configuration
	tf.Step(t, "Test environment variable configuration")

	// Set test environment variables
	os.Setenv("PORTUNIX_LOG_LEVEL", "debug")
	os.Setenv("PORTUNIX_LOG_FORMAT", "json")
	os.Setenv("PORTUNIX_LOG_MODULE_MCP", "error")

	defer func() {
		os.Unsetenv("PORTUNIX_LOG_LEVEL")
		os.Unsetenv("PORTUNIX_LOG_FORMAT")
		os.Unsetenv("PORTUNIX_LOG_MODULE_MCP")
	}()

	envConfig := logging.DefaultConfig()
	envConfig.LoadFromEnv()

	if envConfig.Level != "debug" {
		tf.Error(t, "Environment level not loaded", "Expected 'debug', got '"+envConfig.Level+"'")
		success = false
		return
	}
	if envConfig.Format != "json" {
		tf.Error(t, "Environment format not loaded", "Expected 'json', got '"+envConfig.Format+"'")
		success = false
		return
	}
	if mcpLevel := envConfig.GetModuleLevel("mcp"); mcpLevel != "error" {
		tf.Error(t, "Module level not loaded", "Expected 'error', got '"+mcpLevel+"'")
		success = false
		return
	}
	tf.Success(t, "Environment variable configuration working")

	tf.Separator()

	// Step 8: Test context propagation
	tf.Step(t, "Test context propagation")
	contextLogger := logger.WithField("correlation_id", "test-correlation-123")
	if contextLogger == nil {
		tf.Error(t, "Context logger creation failed", "WithField returned nil")
		success = false
		return
	}

	buf.Reset()
	contextLogger.SetOutput(&buf)
	contextLogger.Info("Context test message")

	output = buf.String()
	if !strings.Contains(output, "correlation_id") || !strings.Contains(output, "test-correlation-123") {
		tf.Error(t, "Context propagation failed", "Expected correlation_id in output", output)
		success = false
		return
	}
	tf.Success(t, "Context propagation working correctly")

	tf.Info(t, "All unit tests completed successfully")
}

func TestIssue052LoggingPerformanceBenchmark(t *testing.T) {
	tf := testframework.NewTestFramework("Issue052_Logging_Performance_Benchmark")
	tf.Start(t, "Performance benchmark testing for logging system")

	success := true
	defer tf.Finish(t, success)

	// Step 1: Baseline performance test
	tf.Step(t, "Baseline performance measurement")
	logger := logging.New("benchmark")

	// Capture baseline
	tf.Info(t, "Running performance baseline measurement...")

	// Simple benchmark - measure time for 1000 log calls
	start := time.Now()
	for i := 0; i < 1000; i++ {
		logger.Info("Benchmark message", "iteration", i, "type", "performance-test")
	}
	duration := time.Since(start)

	tf.Info(t, "Baseline performance",
		"messages", "1000",
		"duration", duration.String(),
		"msgs_per_sec", int(1000.0/duration.Seconds()))

	// Performance should be reasonable (>10k messages/sec on modern hardware)
	msgsPerSec := 1000.0 / duration.Seconds()
	if msgsPerSec < 1000 {
		tf.Warning(t, "Performance may be suboptimal", "messages_per_second", int(msgsPerSec))
	} else {
		tf.Success(t, "Performance is acceptable", "messages_per_second", int(msgsPerSec))
	}

	tf.Separator()

	// Step 2: Structured logging performance
	tf.Step(t, "Structured logging performance")
	start = time.Now()
	for i := 0; i < 1000; i++ {
		logger.Info("Structured benchmark",
			"iteration", i,
			"user_id", "user-123",
			"operation", "test",
			"timestamp", time.Now(),
			"success", true)
	}
	structuredDuration := time.Since(start)

	tf.Info(t, "Structured logging performance",
		"messages", "1000",
		"duration", structuredDuration.String(),
		"msgs_per_sec", int(1000.0/structuredDuration.Seconds()))

	tf.Success(t, "Structured logging performance measured")

	tf.Separator()

	// Step 3: File output performance
	tf.Step(t, "File output performance test")
	tempDir := t.TempDir()
	config := &logging.Config{
		Level:    "info",
		Format:   "json",
		Output:   []string{"file"},
		FilePath: filepath.Join(tempDir, "perf-test.log"),
	}

	factory := logging.NewFactory(config)
	fileLogger := factory.CreateLogger("file-benchmark")

	start = time.Now()
	for i := 0; i < 1000; i++ {
		fileLogger.Info("File benchmark message", "iteration", i)
	}
	fileDuration := time.Since(start)

	tf.Info(t, "File output performance",
		"messages", "1000",
		"duration", fileDuration.String(),
		"msgs_per_sec", int(1000.0/fileDuration.Seconds()))

	// Verify file was written
	if stat, err := os.Stat(config.FilePath); err != nil {
		tf.Error(t, "Log file not created", err.Error())
		success = false
		return
	} else if stat.Size() == 0 {
		tf.Error(t, "Log file is empty", "File size is 0")
		success = false
		return
	}

	tf.Success(t, "File output performance acceptable")

	tf.Info(t, "Performance benchmark completed")
}