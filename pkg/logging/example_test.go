package logging_test

import (
	"context"
	"fmt"
	"os"

	"portunix.ai/portunix/pkg/logging"
	"portunix.ai/portunix/pkg/logging/middleware"
)

// Example demonstrates basic logging usage
func ExampleNew() {
	// Create a logger for a specific component
	logger := logging.New("example")

	// Basic logging
	logger.Info("Application started")
	logger.Debug("Debug information", "version", "1.0.0")
	logger.Error("Something went wrong", "error", "connection failed")

	// With structured fields
	logger.WithField("user_id", "12345").Info("User logged in")

	// Multiple fields
	logger.WithFields(map[string]interface{}{
		"user_id":    "12345",
		"session_id": "abc-def-123",
		"ip":         "192.168.1.1",
	}).Info("User action performed")
}

// Example demonstrates configuration usage
func ExampleConfig() {
	// Create custom configuration
	config := &logging.Config{
		Level:      "debug",
		Format:     "json",
		Output:     []string{"console", "file"},
		FilePath:   "/tmp/example.log",
		TimeFormat: "15:04:05",
		NoColor:    false,
		Modules: map[string]string{
			"database": "warn",
			"cache":    "error",
		},
	}

	// Update global configuration
	logging.UpdateGlobalConfig(config)

	// Get logger with custom config
	logger := logging.GetLogger("example")
	logger.Info("Using custom configuration")
}

// Example demonstrates context usage
func ExampleContext() {
	ctx := context.Background()

	// Add correlation ID to context
	ctx = logging.WithCorrelationID(ctx, "req-123")

	// Add user ID to context
	ctx = logging.WithUserID(ctx, "user-456")

	// Get logger from context
	logger := logging.LoggerFromContext(ctx)
	logger.Info("Processing request")

	// Use context logging helpers
	logging.InfoContext(ctx, "Context-aware logging")
}

// Example demonstrates correlation ID middleware
func ExampleCorrelationMiddleware() {
	// Create correlation middleware
	mw := middleware.NewCorrelationMiddleware(nil)

	// Use with HTTP handler (pseudo-code)
	fmt.Println("Middleware created for HTTP correlation tracking")
	_ = mw // Avoid unused variable warning

	// Create correlation context for non-HTTP scenarios
	ctx := middleware.NewContextWithCorrelationID(context.Background())

	// Use correlation ID
	if id, exists := logging.CorrelationIDFromContext(ctx); exists {
		fmt.Printf("Correlation ID: %s\n", id)
	}
}

// Example demonstrates file logging
func Example_fileLogging() {
	// Create configuration with file output
	config := &logging.Config{
		Level:    "info",
		Format:   "json",
		Output:   []string{"file"},
		FilePath: "/tmp/app.log",
	}

	// Update configuration
	logging.UpdateGlobalConfig(config)

	// Use logger
	logger := logging.GetLogger("file-example")
	logger.Info("This will be written to file")

	fmt.Println("Log written to /tmp/app.log")
}

// Example demonstrates different log levels
func Example_logLevels() {
	logger := logging.New("levels")

	// Set to debug level to see all messages
	logging.SetGlobalLogLevel("debug")

	logger.Trace("Very detailed trace information")
	logger.Debug("Debug information for developers")
	logger.Info("General information")
	logger.Warn("Warning: something might be wrong")
	logger.Error("Error: something is definitely wrong")

	// Note: Fatal and Panic would exit/panic the program
}

// Example demonstrates module-specific log levels
func Example_moduleLevels() {
	// Set different log levels for different modules
	logging.SetModuleLogLevel("database", "debug")
	logging.SetModuleLogLevel("cache", "warn")
	logging.SetGlobalLogLevel("info")

	// Database module will log debug messages
	dbLogger := logging.GetLogger("database")
	dbLogger.Debug("Database query executed")

	// Cache module will only log warnings and above
	cacheLogger := logging.GetLogger("cache")
	cacheLogger.Debug("This won't appear")
	cacheLogger.Warn("Cache miss detected")

	// Other modules use global level (info)
	apiLogger := logging.GetLogger("api")
	apiLogger.Debug("This won't appear")
	apiLogger.Info("API request received")
}

// Example demonstrates environment variable configuration
func Example_environmentConfig() {
	// Set environment variables (normally done outside the program)
	os.Setenv("PORTUNIX_LOG_LEVEL", "debug")
	os.Setenv("PORTUNIX_LOG_FORMAT", "json")
	os.Setenv("PORTUNIX_LOG_OUTPUT", "console,file")
	os.Setenv("PORTUNIX_LOG_FILE", "/tmp/env-config.log")
	os.Setenv("PORTUNIX_LOG_MODULE_AUTH", "warn")

	// Create config and load from environment
	config := logging.DefaultConfig()
	config.LoadFromEnv()

	// Use the environment configuration
	logging.UpdateGlobalConfig(config)

	logger := logging.GetLogger("env-example")
	logger.Info("Using environment configuration")

	// Clean up
	os.Unsetenv("PORTUNIX_LOG_LEVEL")
	os.Unsetenv("PORTUNIX_LOG_FORMAT")
	os.Unsetenv("PORTUNIX_LOG_OUTPUT")
	os.Unsetenv("PORTUNIX_LOG_FILE")
	os.Unsetenv("PORTUNIX_LOG_MODULE_AUTH")
}

// Example demonstrates error logging with structured data
func Example_errorLogging() {
	logger := logging.New("error-example")

	// Simulate an error
	err := fmt.Errorf("database connection failed")

	// Log error with context
	logger.WithField("database", "postgres").
		WithField("host", "localhost").
		WithField("port", 5432).
		Error("Database connection error", "error", err.Error())

	// Using helper function
	enrichedLogger := logging.WithError(logger, err)
	enrichedLogger.Error("Operation failed")
}

// Example demonstrates span/operation tracing
func Example_spanTracing() {
	ctx := context.Background()

	// Create a span for an operation
	spanCtx, finish := middleware.CreateSpan(ctx, "user_login")
	defer finish()

	// Use the span context for logging
	logging.InfoContext(spanCtx, "Starting user login process")

	// Create child span
	childCtx := middleware.CreateChildContext(spanCtx, "validate_credentials")
	logging.InfoContext(childCtx, "Validating user credentials")

	logging.InfoContext(spanCtx, "User login completed successfully")
}