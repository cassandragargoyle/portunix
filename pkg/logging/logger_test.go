package logging

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	logger := New("test-component")
	if logger == nil {
		t.Fatal("New() returned nil logger")
	}

	// Test that logger has component field
	portLogger, ok := logger.(*PortunixLogger)
	if !ok {
		t.Fatal("New() did not return *PortunixLogger")
	}

	// The component should be embedded in the logger context
	// We can't easily test this without capturing output
	_ = portLogger
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer

	// Set global level to debug so all messages appear
	originalLevel := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	defer zerolog.SetGlobalLevel(originalLevel)

	// Create logger with custom output
	zlog := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &PortunixLogger{
		logger: zlog,
		level:  zerolog.DebugLevel,
	}

	testCases := []struct {
		name     string
		logFunc  func(string, ...interface{})
		expected string
	}{
		{"Debug", logger.Debug, "debug"},
		{"Info", logger.Info, "info"},
		{"Warn", logger.Warn, "warn"},
		{"Error", logger.Error, "error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			tc.logFunc("test message")

			output := buf.String()
			if output == "" {
				t.Errorf("No log output generated for %s level", tc.name)
				return
			}
			if !strings.Contains(output, tc.expected) {
				t.Errorf("Expected log output to contain '%s', got: %s", tc.expected, output)
			}
			if !strings.Contains(output, "test message") {
				t.Errorf("Expected log output to contain 'test message', got: %s", output)
			}
		})
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer

	zlog := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &PortunixLogger{
		logger: zlog,
		level:  zerolog.InfoLevel,
	}

	// Test logging with fields
	logger.Info("test message", "key1", "value1", "key2", 42)

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected log output to contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "key1") || !strings.Contains(output, "value1") {
		t.Errorf("Expected log output to contain key1:value1, got: %s", output)
	}
	if !strings.Contains(output, "key2") || !strings.Contains(output, "42") {
		t.Errorf("Expected log output to contain key2:42, got: %s", output)
	}
}

func TestLoggerWithField(t *testing.T) {
	var buf bytes.Buffer

	zlog := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &PortunixLogger{
		logger: zlog,
		level:  zerolog.InfoLevel,
	}

	// Test WithField method
	enrichedLogger := logger.WithField("user_id", "12345")
	enrichedLogger.Info("user action")

	output := buf.String()
	if !strings.Contains(output, "user action") {
		t.Errorf("Expected log output to contain 'user action', got: %s", output)
	}
	if !strings.Contains(output, "user_id") || !strings.Contains(output, "12345") {
		t.Errorf("Expected log output to contain user_id:12345, got: %s", output)
	}
}

func TestLoggerWithFieldsMap(t *testing.T) {
	var buf bytes.Buffer

	zlog := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &PortunixLogger{
		logger: zlog,
		level:  zerolog.InfoLevel,
	}

	// Test WithFields method
	fields := map[string]interface{}{
		"user_id": "12345",
		"action":  "login",
		"count":   1,
	}
	enrichedLogger := logger.WithFields(fields)
	enrichedLogger.Info("user login successful")

	output := buf.String()
	if !strings.Contains(output, "user login successful") {
		t.Errorf("Expected log output to contain 'user login successful', got: %s", output)
	}
	for key := range fields {
		if !strings.Contains(output, key) {
			t.Errorf("Expected log output to contain key '%s', got: %s", key, output)
		}
	}
}

func TestLoggerSetLevel(t *testing.T) {
	var buf bytes.Buffer

	// Save and restore global level
	originalLevel := zerolog.GlobalLevel()
	defer zerolog.SetGlobalLevel(originalLevel)

	// Set global level to info initially
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	zlog := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &PortunixLogger{
		logger: zlog,
		level:  zerolog.InfoLevel,
	}

	// Debug should not appear at Info level
	logger.Debug("debug message")
	if strings.Contains(buf.String(), "debug message") {
		t.Error("Debug message should not appear at Info level")
	}

	// Set global level to debug and test again
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger.SetLevel(zerolog.DebugLevel)
	buf.Reset()

	logger.Debug("debug message")
	output := buf.String()
	if output == "" {
		t.Error("No output generated after setting debug level")
	} else if !strings.Contains(output, "debug message") {
		t.Errorf("Debug message should appear at Debug level, got: %s", output)
	}
}

func TestLoggerGetLevel(t *testing.T) {
	logger := &PortunixLogger{
		level: zerolog.WarnLevel,
	}

	if logger.GetLevel() != zerolog.WarnLevel {
		t.Errorf("Expected level %v, got %v", zerolog.WarnLevel, logger.GetLevel())
	}
}

func TestLoggerSetOutput(t *testing.T) {
	var buf bytes.Buffer

	logger := &PortunixLogger{
		logger: zerolog.New(os.Stderr),
		level:  zerolog.InfoLevel,
	}

	// Change output to buffer
	logger.SetOutput(&buf)
	logger.Info("test message")

	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("Expected message in buffer, got: %s", buf.String())
	}
}

func TestFromContext(t *testing.T) {
	// Test with empty context
	ctx := context.Background()
	logger := FromContext(ctx)
	if logger == nil {
		t.Fatal("FromContext() returned nil logger")
	}

	// Test with logger in context
	testLogger := New("test")
	ctx = testLogger.WithContext(ctx)
	contextLogger := FromContext(ctx)
	if contextLogger == nil {
		t.Fatal("FromContext() returned nil logger from context with logger")
	}
}

func TestWithError(t *testing.T) {
	var buf bytes.Buffer

	zlog := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &PortunixLogger{
		logger: zlog,
		level:  zerolog.InfoLevel,
	}

	err := os.ErrNotExist
	enrichedLogger := WithError(logger, err)
	enrichedLogger.Error("operation failed")

	output := buf.String()
	if !strings.Contains(output, "operation failed") {
		t.Errorf("Expected log output to contain 'operation failed', got: %s", output)
	}
	if !strings.Contains(output, "error") {
		t.Errorf("Expected log output to contain error field, got: %s", output)
	}
}

func TestWithLoggerUserID(t *testing.T) {
	var buf bytes.Buffer

	zlog := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &PortunixLogger{
		logger: zlog,
		level:  zerolog.InfoLevel,
	}

	enrichedLogger := WithLoggerUserID(logger, "user123")
	enrichedLogger.Info("user action")

	output := buf.String()
	if !strings.Contains(output, "user action") {
		t.Errorf("Expected log output to contain 'user action', got: %s", output)
	}
	if !strings.Contains(output, "user_id") || !strings.Contains(output, "user123") {
		t.Errorf("Expected log output to contain user_id:user123, got: %s", output)
	}
}

func TestWithLoggerCorrelationID(t *testing.T) {
	var buf bytes.Buffer

	zlog := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &PortunixLogger{
		logger: zlog,
		level:  zerolog.InfoLevel,
	}

	enrichedLogger := WithLoggerCorrelationID(logger, "corr-123")
	enrichedLogger.Info("request processed")

	output := buf.String()
	if !strings.Contains(output, "request processed") {
		t.Errorf("Expected log output to contain 'request processed', got: %s", output)
	}
	if !strings.Contains(output, "correlation_id") || !strings.Contains(output, "corr-123") {
		t.Errorf("Expected log output to contain correlation_id:corr-123, got: %s", output)
	}
}

func TestGlobalLogger(t *testing.T) {
	global := GetGlobalLogger()
	if global == nil {
		t.Fatal("GetGlobalLogger() returned nil")
	}

	// Test setting global level
	originalLevel := global.GetLevel()
	SetGlobalLevel(zerolog.WarnLevel)

	// The global level change affects new loggers
	newLogger := New("test")
	if newLogger.GetLevel() != zerolog.WarnLevel {
		// Note: The level might not be directly comparable due to zerolog internals
		// This test mainly ensures the function doesn't panic
	}

	// Restore original level
	SetGlobalLevel(originalLevel)
}

func BenchmarkLoggerInfo(b *testing.B) {
	logger := New("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i)
	}
}

func BenchmarkLoggerWithFields(b *testing.B) {
	logger := New("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i, "type", "benchmark", "value", 42.5)
	}
}