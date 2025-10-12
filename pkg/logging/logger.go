// Package logging provides structured logging for Portunix
package logging

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
)

// Logger is the main interface for logging operations
type Logger interface {
	// Core logging methods
	Trace(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	Panic(msg string, fields ...interface{})

	// Structured logging
	With() *zerolog.Event
	WithFields(fields map[string]interface{}) Logger
	WithField(key string, value interface{}) Logger

	// Context operations
	WithContext(ctx context.Context) context.Context

	// Level management
	SetLevel(level zerolog.Level)
	GetLevel() zerolog.Level

	// Output management
	SetOutput(w io.Writer)
}

// PortunixLogger implements the Logger interface using zerolog
type PortunixLogger struct {
	logger zerolog.Logger
	level  zerolog.Level
}

// Global logger instance
var (
	globalLogger *PortunixLogger
	defaultLevel = zerolog.InfoLevel
)

func init() {
	// Initialize global logger with sensible defaults
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
	}

	zlog := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	globalLogger = &PortunixLogger{
		logger: zlog,
		level:  defaultLevel,
	}

	zerolog.SetGlobalLevel(defaultLevel)
}

// New creates a new logger instance with the given component name
func New(component string) Logger {
	logger := globalLogger.logger.With().Str("component", component).Logger()
	return &PortunixLogger{
		logger: logger,
		level:  globalLogger.level,
	}
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	return globalLogger
}

// SetGlobalLevel sets the global log level
func SetGlobalLevel(level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
	globalLogger.level = level
}

// Implementation of Logger interface methods

func (l *PortunixLogger) Trace(msg string, fields ...interface{}) {
	event := l.logger.Trace()
	l.logWithFields(event, msg, fields...)
}

func (l *PortunixLogger) Debug(msg string, fields ...interface{}) {
	event := l.logger.Debug()
	l.logWithFields(event, msg, fields...)
}

func (l *PortunixLogger) Info(msg string, fields ...interface{}) {
	event := l.logger.Info()
	l.logWithFields(event, msg, fields...)
}

func (l *PortunixLogger) Warn(msg string, fields ...interface{}) {
	event := l.logger.Warn()
	l.logWithFields(event, msg, fields...)
}

func (l *PortunixLogger) Error(msg string, fields ...interface{}) {
	event := l.logger.Error()
	l.logWithFields(event, msg, fields...)
}

func (l *PortunixLogger) Fatal(msg string, fields ...interface{}) {
	event := l.logger.Fatal()
	l.logWithFields(event, msg, fields...)
}

func (l *PortunixLogger) Panic(msg string, fields ...interface{}) {
	event := l.logger.Panic()
	l.logWithFields(event, msg, fields...)
}

func (l *PortunixLogger) With() *zerolog.Event {
	return l.logger.Info()
}

func (l *PortunixLogger) WithFields(fields map[string]interface{}) Logger {
	newLogger := l.logger.With().Fields(fields).Logger()
	return &PortunixLogger{
		logger: newLogger,
		level:  l.level,
	}
}

func (l *PortunixLogger) WithField(key string, value interface{}) Logger {
	newLogger := l.logger.With().Interface(key, value).Logger()
	return &PortunixLogger{
		logger: newLogger,
		level:  l.level,
	}
}

func (l *PortunixLogger) WithContext(ctx context.Context) context.Context {
	return l.logger.WithContext(ctx)
}

func (l *PortunixLogger) SetLevel(level zerolog.Level) {
	l.level = level
	// Create a new logger with the specified level
	l.logger = l.logger.Level(level)
}

func (l *PortunixLogger) GetLevel() zerolog.Level {
	return l.level
}

func (l *PortunixLogger) SetOutput(w io.Writer) {
	l.logger = l.logger.Output(w)
}

// Helper function to handle field pairs
func (l *PortunixLogger) logWithFields(event *zerolog.Event, msg string, fields ...interface{}) {
	// Handle field pairs (key, value, key, value, ...)
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			event = event.Interface(key, fields[i+1])
		}
	}
	event.Msg(msg)
}

// FromContext extracts logger from context
func FromContext(ctx context.Context) Logger {
	zlog := zerolog.Ctx(ctx)
	if zlog != nil {
		return &PortunixLogger{
			logger: *zlog,
			level:  globalLogger.level,
		}
	}
	return globalLogger
}

// Helper functions for common field types

// WithError adds an error field to the logger
func WithError(l Logger, err error) Logger {
	return l.WithField("error", err.Error())
}