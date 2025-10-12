package handlers

import (
	"fmt"
	"io"
	"log/syslog"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
)

// SyslogHandler handles syslog output
type SyslogHandler struct {
	writer *syslog.Writer
	config SyslogConfig
}

// SyslogConfig holds configuration for syslog output
type SyslogConfig struct {
	Network  string            // Network type: "tcp", "udp", or "" for local
	Address  string            // Address for remote syslog (host:port)
	Priority syslog.Priority   // Syslog priority (facility + severity)
	Tag      string            // Syslog tag (application name)
	Facility syslog.Priority   // Syslog facility
}

// DefaultSyslogConfig returns default syslog configuration
func DefaultSyslogConfig() SyslogConfig {
	return SyslogConfig{
		Network:  "", // Local syslog
		Address:  "",
		Priority: syslog.LOG_INFO | syslog.LOG_USER,
		Tag:      "portunix",
		Facility: syslog.LOG_USER,
	}
}

// NewSyslogHandler creates a new syslog handler
func NewSyslogHandler(config SyslogConfig) (*SyslogHandler, error) {
	// Check if syslog is supported on this platform
	if !IsSyslogSupported() {
		return nil, fmt.Errorf("syslog is not supported on %s", runtime.GOOS)
	}

	var writer *syslog.Writer
	var err error

	if config.Network != "" && config.Address != "" {
		// Remote syslog
		writer, err = syslog.Dial(config.Network, config.Address, config.Priority, config.Tag)
	} else {
		// Local syslog
		writer, err = syslog.New(config.Priority, config.Tag)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to syslog: %w", err)
	}

	return &SyslogHandler{
		writer: writer,
		config: config,
	}, nil
}

// Write implements io.Writer interface
func (h *SyslogHandler) Write(p []byte) (n int, err error) {
	// Parse the log entry to extract level and message
	level, message := h.parseLogEntry(string(p))

	// Write to appropriate syslog level
	switch level {
	case zerolog.TraceLevel, zerolog.DebugLevel:
		err = h.writer.Debug(message)
	case zerolog.InfoLevel:
		err = h.writer.Info(message)
	case zerolog.WarnLevel:
		err = h.writer.Warning(message)
	case zerolog.ErrorLevel:
		err = h.writer.Err(message)
	case zerolog.FatalLevel, zerolog.PanicLevel:
		err = h.writer.Crit(message)
	default:
		err = h.writer.Info(message)
	}

	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// parseLogEntry extracts log level and message from JSON log entry
func (h *SyslogHandler) parseLogEntry(entry string) (zerolog.Level, string) {
	// Simple JSON parsing for level and message
	// This is a basic implementation - could be enhanced with proper JSON parsing

	entry = strings.TrimSpace(entry)

	// Extract level
	level := zerolog.InfoLevel
	if strings.Contains(entry, `"level":"trace"`) {
		level = zerolog.TraceLevel
	} else if strings.Contains(entry, `"level":"debug"`) {
		level = zerolog.DebugLevel
	} else if strings.Contains(entry, `"level":"info"`) {
		level = zerolog.InfoLevel
	} else if strings.Contains(entry, `"level":"warn"`) {
		level = zerolog.WarnLevel
	} else if strings.Contains(entry, `"level":"error"`) {
		level = zerolog.ErrorLevel
	} else if strings.Contains(entry, `"level":"fatal"`) {
		level = zerolog.FatalLevel
	} else if strings.Contains(entry, `"level":"panic"`) {
		level = zerolog.PanicLevel
	}

	// For syslog, send the entire JSON entry as the message
	// This preserves all structured data
	return level, entry
}

// Close closes the syslog handler
func (h *SyslogHandler) Close() error {
	if h.writer != nil {
		err := h.writer.Close()
		h.writer = nil
		return err
	}
	return nil
}

// GetWriter returns the underlying syslog writer
func (h *SyslogHandler) GetWriter() io.Writer {
	return h
}

// IsSyslogSupported checks if syslog is supported on the current platform
func IsSyslogSupported() bool {
	// Syslog is supported on Unix-like systems
	switch runtime.GOOS {
	case "linux", "darwin", "freebsd", "openbsd", "netbsd", "dragonfly", "solaris":
		return true
	case "windows":
		// Windows doesn't have native syslog, but could support remote syslog
		return false
	default:
		return false
	}
}

// SyslogHandlerBuilder provides a fluent interface for building syslog handlers
type SyslogHandlerBuilder struct {
	config SyslogConfig
}

// NewSyslogBuilder creates a new syslog handler builder
func NewSyslogBuilder() *SyslogHandlerBuilder {
	return &SyslogHandlerBuilder{
		config: DefaultSyslogConfig(),
	}
}

// WithNetwork sets the network type for remote syslog
func (b *SyslogHandlerBuilder) WithNetwork(network string) *SyslogHandlerBuilder {
	b.config.Network = network
	return b
}

// WithAddress sets the address for remote syslog
func (b *SyslogHandlerBuilder) WithAddress(address string) *SyslogHandlerBuilder {
	b.config.Address = address
	return b
}

// WithTCP configures for TCP remote syslog
func (b *SyslogHandlerBuilder) WithTCP(address string) *SyslogHandlerBuilder {
	b.config.Network = "tcp"
	b.config.Address = address
	return b
}

// WithUDP configures for UDP remote syslog
func (b *SyslogHandlerBuilder) WithUDP(address string) *SyslogHandlerBuilder {
	b.config.Network = "udp"
	b.config.Address = address
	return b
}

// WithLocal configures for local syslog
func (b *SyslogHandlerBuilder) WithLocal() *SyslogHandlerBuilder {
	b.config.Network = ""
	b.config.Address = ""
	return b
}

// WithTag sets the syslog tag
func (b *SyslogHandlerBuilder) WithTag(tag string) *SyslogHandlerBuilder {
	b.config.Tag = tag
	return b
}

// WithFacility sets the syslog facility
func (b *SyslogHandlerBuilder) WithFacility(facility syslog.Priority) *SyslogHandlerBuilder {
	b.config.Facility = facility
	b.config.Priority = (b.config.Priority & 0x07) | facility
	return b
}

// WithPriority sets the full syslog priority
func (b *SyslogHandlerBuilder) WithPriority(priority syslog.Priority) *SyslogHandlerBuilder {
	b.config.Priority = priority
	return b
}

// Build creates the syslog handler
func (b *SyslogHandlerBuilder) Build() (*SyslogHandler, error) {
	return NewSyslogHandler(b.config)
}

// Helper functions for common syslog configurations

// NewLocalSyslogHandler creates a local syslog handler
func NewLocalSyslogHandler(tag string) (*SyslogHandler, error) {
	if !IsSyslogSupported() {
		return nil, fmt.Errorf("syslog not supported on %s", runtime.GOOS)
	}

	return NewSyslogBuilder().
		WithLocal().
		WithTag(tag).
		WithFacility(syslog.LOG_USER).
		Build()
}

// NewRemoteSyslogHandler creates a remote syslog handler
func NewRemoteSyslogHandler(network, address, tag string) (*SyslogHandler, error) {
	return NewSyslogBuilder().
		WithNetwork(network).
		WithAddress(address).
		WithTag(tag).
		WithFacility(syslog.LOG_USER).
		Build()
}

// NewDaemonSyslogHandler creates a syslog handler for daemon applications
func NewDaemonSyslogHandler(tag string) (*SyslogHandler, error) {
	if !IsSyslogSupported() {
		return nil, fmt.Errorf("syslog not supported on %s", runtime.GOOS)
	}

	return NewSyslogBuilder().
		WithLocal().
		WithTag(tag).
		WithFacility(syslog.LOG_DAEMON).
		Build()
}

// FacilityFromString converts string to syslog facility
func FacilityFromString(facility string) syslog.Priority {
	switch strings.ToLower(facility) {
	case "kern", "kernel":
		return syslog.LOG_KERN
	case "user":
		return syslog.LOG_USER
	case "mail":
		return syslog.LOG_MAIL
	case "daemon":
		return syslog.LOG_DAEMON
	case "auth":
		return syslog.LOG_AUTH
	case "syslog":
		return syslog.LOG_SYSLOG
	case "lpr":
		return syslog.LOG_LPR
	case "news":
		return syslog.LOG_NEWS
	case "uucp":
		return syslog.LOG_UUCP
	case "cron":
		return syslog.LOG_CRON
	case "authpriv":
		return syslog.LOG_AUTHPRIV
	case "ftp":
		return syslog.LOG_FTP
	case "local0":
		return syslog.LOG_LOCAL0
	case "local1":
		return syslog.LOG_LOCAL1
	case "local2":
		return syslog.LOG_LOCAL2
	case "local3":
		return syslog.LOG_LOCAL3
	case "local4":
		return syslog.LOG_LOCAL4
	case "local5":
		return syslog.LOG_LOCAL5
	case "local6":
		return syslog.LOG_LOCAL6
	case "local7":
		return syslog.LOG_LOCAL7
	default:
		return syslog.LOG_USER
	}
}