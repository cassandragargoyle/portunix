package logging

import (
	"context"

	"github.com/rs/zerolog"
)

// contextKey is a type for context keys
type contextKey string

const (
	// loggerKey is the context key for storing logger
	loggerKey contextKey = "portunix_logger"

	// correlationIDKey is the context key for correlation ID
	correlationIDKey contextKey = "correlation_id"

	// userIDKey is the context key for user ID
	userIDKey contextKey = "user_id"
)

// WithLogger adds logger to context
func WithLogger(ctx context.Context, logger Logger) context.Context {
	if pl, ok := logger.(*PortunixLogger); ok {
		// Also add to zerolog context for compatibility
		ctx = pl.logger.WithContext(ctx)
	}
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext extracts logger from context
func LoggerFromContext(ctx context.Context) Logger {
	// First try our custom logger
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}

	// Then try zerolog logger
	if zlog := zerolog.Ctx(ctx); zlog != nil {
		return &PortunixLogger{
			logger: *zlog,
			level:  globalLogger.level,
		}
	}

	// Return global logger as fallback
	return globalLogger
}

// WithCorrelationID adds correlation ID to context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	// Add to context
	ctx = context.WithValue(ctx, correlationIDKey, correlationID)

	// Also add to logger if present
	if logger := LoggerFromContext(ctx); logger != nil {
		logger = logger.WithField("correlation_id", correlationID)
		ctx = WithLogger(ctx, logger)
	}

	return ctx
}

// CorrelationIDFromContext extracts correlation ID from context
func CorrelationIDFromContext(ctx context.Context) (string, bool) {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id, true
	}
	return "", false
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	// Add to context
	ctx = context.WithValue(ctx, userIDKey, userID)

	// Also add to logger if present
	if logger := LoggerFromContext(ctx); logger != nil {
		logger = logger.WithField("user_id", userID)
		ctx = WithLogger(ctx, logger)
	}

	return ctx
}

// UserIDFromContext extracts user ID from context
func UserIDFromContext(ctx context.Context) (string, bool) {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id, true
	}
	return "", false
}

// Helper functions for logger enrichment (non-context)

// WithLoggerError adds an error field to the logger
func WithLoggerError(l Logger, err error) Logger {
	return l.WithField("error", err.Error())
}

// WithLoggerUserID adds a user ID field to the logger
func WithLoggerUserID(l Logger, userID string) Logger {
	return l.WithField("user_id", userID)
}

// WithLoggerCorrelationID adds a correlation ID field to the logger
func WithLoggerCorrelationID(l Logger, correlationID string) Logger {
	return l.WithField("correlation_id", correlationID)
}

// EnrichContext adds common fields to context logger
func EnrichContext(ctx context.Context, fields map[string]interface{}) context.Context {
	logger := LoggerFromContext(ctx)
	if logger == nil {
		logger = globalLogger
	}

	// Add all fields to logger
	logger = logger.WithFields(fields)

	// Update context with enriched logger
	return WithLogger(ctx, logger)
}

// ContextualLogger provides context-aware logging
type ContextualLogger struct {
	ctx context.Context
}

// NewContextualLogger creates a new contextual logger
func NewContextualLogger(ctx context.Context) *ContextualLogger {
	return &ContextualLogger{ctx: ctx}
}

// Trace logs at trace level with context
func (cl *ContextualLogger) Trace(msg string, fields ...interface{}) {
	logger := LoggerFromContext(cl.ctx)
	logger.Trace(msg, fields...)
}

// Debug logs at debug level with context
func (cl *ContextualLogger) Debug(msg string, fields ...interface{}) {
	logger := LoggerFromContext(cl.ctx)
	logger.Debug(msg, fields...)
}

// Info logs at info level with context
func (cl *ContextualLogger) Info(msg string, fields ...interface{}) {
	logger := LoggerFromContext(cl.ctx)
	logger.Info(msg, fields...)
}

// Warn logs at warn level with context
func (cl *ContextualLogger) Warn(msg string, fields ...interface{}) {
	logger := LoggerFromContext(cl.ctx)
	logger.Warn(msg, fields...)
}

// Error logs at error level with context
func (cl *ContextualLogger) Error(msg string, fields ...interface{}) {
	logger := LoggerFromContext(cl.ctx)
	logger.Error(msg, fields...)
}

// Fatal logs at fatal level with context
func (cl *ContextualLogger) Fatal(msg string, fields ...interface{}) {
	logger := LoggerFromContext(cl.ctx)
	logger.Fatal(msg, fields...)
}

// Panic logs at panic level with context
func (cl *ContextualLogger) Panic(msg string, fields ...interface{}) {
	logger := LoggerFromContext(cl.ctx)
	logger.Panic(msg, fields...)
}

// Helper functions for quick context logging

// TraceContext logs at trace level using context logger
func TraceContext(ctx context.Context, msg string, fields ...interface{}) {
	LoggerFromContext(ctx).Trace(msg, fields...)
}

// DebugContext logs at debug level using context logger
func DebugContext(ctx context.Context, msg string, fields ...interface{}) {
	LoggerFromContext(ctx).Debug(msg, fields...)
}

// InfoContext logs at info level using context logger
func InfoContext(ctx context.Context, msg string, fields ...interface{}) {
	LoggerFromContext(ctx).Info(msg, fields...)
}

// WarnContext logs at warn level using context logger
func WarnContext(ctx context.Context, msg string, fields ...interface{}) {
	LoggerFromContext(ctx).Warn(msg, fields...)
}

// ErrorContext logs at error level using context logger
func ErrorContext(ctx context.Context, msg string, fields ...interface{}) {
	LoggerFromContext(ctx).Error(msg, fields...)
}

// FatalContext logs at fatal level using context logger
func FatalContext(ctx context.Context, msg string, fields ...interface{}) {
	LoggerFromContext(ctx).Fatal(msg, fields...)
}

// PanicContext logs at panic level using context logger
func PanicContext(ctx context.Context, msg string, fields ...interface{}) {
	LoggerFromContext(ctx).Panic(msg, fields...)
}