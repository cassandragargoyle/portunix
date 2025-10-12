package handlers

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// ConsoleHandler handles console output with pretty formatting
type ConsoleHandler struct {
	writer  io.Writer
	config  ConsoleConfig
	console zerolog.ConsoleWriter
}

// ConsoleConfig holds configuration for console output
type ConsoleConfig struct {
	Output      io.Writer // Output destination (stdout, stderr)
	TimeFormat  string    // Time format for timestamps
	NoColor     bool      // Disable color output
	ShowCaller  bool      // Show caller information
	ShowLevel   bool      // Show log level
	ShowTime    bool      // Show timestamp
	Compact     bool      // Use compact format
	JSONFormat  bool      // Use JSON format instead of pretty console
}

// DefaultConsoleConfig returns default console configuration
func DefaultConsoleConfig() ConsoleConfig {
	return ConsoleConfig{
		Output:     os.Stderr,
		TimeFormat: "15:04:05",
		NoColor:    false,
		ShowCaller: false,
		ShowLevel:  true,
		ShowTime:   true,
		Compact:    false,
		JSONFormat: false,
	}
}

// NewConsoleHandler creates a new console handler
func NewConsoleHandler(config ConsoleConfig) *ConsoleHandler {
	if config.Output == nil {
		config.Output = os.Stderr
	}

	handler := &ConsoleHandler{
		writer: config.Output,
		config: config,
	}

	// If JSON format is requested, return raw writer
	if config.JSONFormat {
		return handler
	}

	// Configure console writer for pretty output
	console := zerolog.ConsoleWriter{
		Out:        config.Output,
		TimeFormat: config.TimeFormat,
		NoColor:    config.NoColor || !isTerminal(config.Output),
	}

	// Configure parts order
	partsOrder := []string{}
	if config.ShowTime {
		partsOrder = append(partsOrder, zerolog.TimestampFieldName)
	}
	if config.ShowLevel {
		partsOrder = append(partsOrder, zerolog.LevelFieldName)
	}
	partsOrder = append(partsOrder, "component")
	if config.ShowCaller {
		partsOrder = append(partsOrder, zerolog.CallerFieldName)
	}
	partsOrder = append(partsOrder, zerolog.MessageFieldName)

	console.PartsOrder = partsOrder

	// Configure field exclusions for cleaner output
	if config.Compact {
		console.FieldsExclude = []string{
			"os", "arch", "go_version", "container", "ci",
		}
	}

	// Custom formatters for better readability
	console.FormatLevel = func(i interface{}) string {
		var l string
		var ll string
		if levelStr, ok := i.(string); ok {
			ll = levelStr
			switch levelStr {
			case "trace":
				l = "TRACE"
			case "debug":
				l = "DEBUG"
			case "info":
				l = " INFO"
			case "warn":
				l = " WARN"
			case "error":
				l = "ERROR"
			case "fatal":
				l = "FATAL"
			case "panic":
				l = "PANIC"
			default:
				l = levelStr
			}
		} else {
			if i == nil {
				l = "?????"
				ll = "unknown"
			} else {
				formatted := fmt.Sprintf("%s", i)
				if len(formatted) >= 3 {
					l = strings.ToUpper(formatted)[0:3]
				} else {
					l = strings.ToUpper(formatted)
				}
				ll = formatted
			}
		}
		return colorize(l, ll, !config.NoColor)
	}

	console.FormatCaller = func(i interface{}) string {
		var c string
		if cc, ok := i.(string); ok {
			c = cc
		}
		if len(c) > 0 {
			if cwd, err := os.Getwd(); err == nil {
				c = strings.TrimPrefix(c, cwd)
				c = strings.TrimPrefix(c, "/")
			}
		}
		return colorize(c, "caller", !config.NoColor)
	}

	handler.console = console
	return handler
}

// Write implements io.Writer interface
func (h *ConsoleHandler) Write(p []byte) (n int, err error) {
	if h.config.JSONFormat {
		return h.writer.Write(p)
	}
	return h.console.Write(p)
}

// Close closes the console handler (no-op for console)
func (h *ConsoleHandler) Close() error {
	return nil
}

// isTerminal checks if the writer is a terminal
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return false
		}
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// colorize adds ANSI color codes to text based on level
func colorize(text, level string, useColor bool) string {
	if !useColor {
		return text
	}

	var colorCode string
	switch level {
	case "trace":
		colorCode = "\033[90m" // Dark gray
	case "debug":
		colorCode = "\033[36m" // Cyan
	case "info":
		colorCode = "\033[32m" // Green
	case "warn":
		colorCode = "\033[33m" // Yellow
	case "error":
		colorCode = "\033[31m" // Red
	case "fatal":
		colorCode = "\033[35m" // Magenta
	case "panic":
		colorCode = "\033[41m" // Red background
	case "caller":
		colorCode = "\033[90m" // Dark gray
	default:
		return text
	}

	return colorCode + text + "\033[0m"
}

// GetWriter returns the underlying writer
func (h *ConsoleHandler) GetWriter() io.Writer {
	if h.config.JSONFormat {
		return h.writer
	}
	return h.console
}

// ConsoleHandlerBuilder provides a fluent interface for building console handlers
type ConsoleHandlerBuilder struct {
	config ConsoleConfig
}

// NewConsoleBuilder creates a new console handler builder
func NewConsoleBuilder() *ConsoleHandlerBuilder {
	return &ConsoleHandlerBuilder{
		config: DefaultConsoleConfig(),
	}
}

// WithOutput sets the output writer
func (b *ConsoleHandlerBuilder) WithOutput(w io.Writer) *ConsoleHandlerBuilder {
	b.config.Output = w
	return b
}

// WithStdout sets output to stdout
func (b *ConsoleHandlerBuilder) WithStdout() *ConsoleHandlerBuilder {
	b.config.Output = os.Stdout
	return b
}

// WithStderr sets output to stderr
func (b *ConsoleHandlerBuilder) WithStderr() *ConsoleHandlerBuilder {
	b.config.Output = os.Stderr
	return b
}

// WithTimeFormat sets the time format
func (b *ConsoleHandlerBuilder) WithTimeFormat(format string) *ConsoleHandlerBuilder {
	b.config.TimeFormat = format
	return b
}

// WithNoColor disables color output
func (b *ConsoleHandlerBuilder) WithNoColor(noColor bool) *ConsoleHandlerBuilder {
	b.config.NoColor = noColor
	return b
}

// WithCaller enables/disables caller information
func (b *ConsoleHandlerBuilder) WithCaller(show bool) *ConsoleHandlerBuilder {
	b.config.ShowCaller = show
	return b
}

// WithLevel enables/disables level display
func (b *ConsoleHandlerBuilder) WithLevel(show bool) *ConsoleHandlerBuilder {
	b.config.ShowLevel = show
	return b
}

// WithTime enables/disables timestamp display
func (b *ConsoleHandlerBuilder) WithTime(show bool) *ConsoleHandlerBuilder {
	b.config.ShowTime = show
	return b
}

// WithCompact enables compact mode
func (b *ConsoleHandlerBuilder) WithCompact(compact bool) *ConsoleHandlerBuilder {
	b.config.Compact = compact
	return b
}

// WithJSON enables JSON format output
func (b *ConsoleHandlerBuilder) WithJSON(json bool) *ConsoleHandlerBuilder {
	b.config.JSONFormat = json
	return b
}

// Build creates the console handler
func (b *ConsoleHandlerBuilder) Build() *ConsoleHandler {
	return NewConsoleHandler(b.config)
}

// Helper functions for common configurations

// NewStdoutHandler creates a console handler that writes to stdout
func NewStdoutHandler() *ConsoleHandler {
	return NewConsoleBuilder().WithStdout().Build()
}

// NewStderrHandler creates a console handler that writes to stderr
func NewStderrHandler() *ConsoleHandler {
	return NewConsoleBuilder().WithStderr().Build()
}

// NewJSONHandler creates a console handler that outputs JSON to stderr
func NewJSONHandler() *ConsoleHandler {
	return NewConsoleBuilder().WithJSON(true).Build()
}

// NewCompactHandler creates a compact console handler
func NewCompactHandler() *ConsoleHandler {
	return NewConsoleBuilder().WithCompact(true).Build()
}

// NewDebugHandler creates a debug-oriented console handler with caller info
func NewDebugHandler() *ConsoleHandler {
	return NewConsoleBuilder().
		WithCaller(true).
		WithTimeFormat("15:04:05.000").
		Build()
}