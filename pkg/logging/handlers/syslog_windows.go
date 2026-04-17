//go:build windows

package handlers

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

// SyslogPriority represents syslog priority on unsupported platforms
type SyslogPriority int

const (
	LOG_INFO     SyslogPriority = 6
	LOG_USER     SyslogPriority = 8
	LOG_KERN     SyslogPriority = 0
	LOG_MAIL     SyslogPriority = 16
	LOG_DAEMON   SyslogPriority = 24
	LOG_AUTH     SyslogPriority = 32
	LOG_SYSLOG   SyslogPriority = 40
	LOG_LPR      SyslogPriority = 48
	LOG_NEWS     SyslogPriority = 56
	LOG_UUCP     SyslogPriority = 64
	LOG_CRON     SyslogPriority = 72
	LOG_AUTHPRIV SyslogPriority = 80
	LOG_FTP      SyslogPriority = 88
	LOG_LOCAL0   SyslogPriority = 128
	LOG_LOCAL1   SyslogPriority = 136
	LOG_LOCAL2   SyslogPriority = 144
	LOG_LOCAL3   SyslogPriority = 152
	LOG_LOCAL4   SyslogPriority = 160
	LOG_LOCAL5   SyslogPriority = 168
	LOG_LOCAL6   SyslogPriority = 176
	LOG_LOCAL7   SyslogPriority = 184
)

// SyslogHandler handles syslog output (stub for Windows)
type SyslogHandler struct {
	config SyslogConfig
}

// SyslogConfig holds configuration for syslog output
type SyslogConfig struct {
	Network  string         // Network type: "tcp", "udp", or "" for local
	Address  string         // Address for remote syslog (host:port)
	Priority SyslogPriority // Syslog priority (facility + severity)
	Tag      string         // Syslog tag (application name)
	Facility SyslogPriority // Syslog facility
}

// DefaultSyslogConfig returns default syslog configuration
func DefaultSyslogConfig() SyslogConfig {
	return SyslogConfig{
		Network:  "",
		Address:  "",
		Priority: SyslogPriority(LOG_INFO | LOG_USER),
		Tag:      "portunix",
		Facility: LOG_USER,
	}
}

// NewSyslogHandler creates a new syslog handler (unsupported on Windows)
func NewSyslogHandler(config SyslogConfig) (*SyslogHandler, error) {
	return nil, fmt.Errorf("syslog is not supported on %s", runtime.GOOS)
}

// Write implements io.Writer interface
func (h *SyslogHandler) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("syslog is not supported on %s", runtime.GOOS)
}

// Close closes the syslog handler
func (h *SyslogHandler) Close() error {
	return nil
}

// GetWriter returns the underlying syslog writer
func (h *SyslogHandler) GetWriter() io.Writer {
	return h
}

// IsSyslogSupported checks if syslog is supported on the current platform
func IsSyslogSupported() bool {
	return false
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
func (b *SyslogHandlerBuilder) WithFacility(facility SyslogPriority) *SyslogHandlerBuilder {
	b.config.Facility = facility
	b.config.Priority = SyslogPriority((int(b.config.Priority) & 0x07) | int(facility))
	return b
}

// WithPriority sets the full syslog priority
func (b *SyslogHandlerBuilder) WithPriority(priority SyslogPriority) *SyslogHandlerBuilder {
	b.config.Priority = priority
	return b
}

// Build creates the syslog handler
func (b *SyslogHandlerBuilder) Build() (*SyslogHandler, error) {
	return NewSyslogHandler(b.config)
}

// NewLocalSyslogHandler creates a local syslog handler (unsupported on Windows)
func NewLocalSyslogHandler(tag string) (*SyslogHandler, error) {
	return nil, fmt.Errorf("syslog not supported on %s", runtime.GOOS)
}

// NewRemoteSyslogHandler creates a remote syslog handler (unsupported on Windows)
func NewRemoteSyslogHandler(network, address, tag string) (*SyslogHandler, error) {
	return nil, fmt.Errorf("syslog not supported on %s", runtime.GOOS)
}

// NewDaemonSyslogHandler creates a syslog handler for daemon applications (unsupported on Windows)
func NewDaemonSyslogHandler(tag string) (*SyslogHandler, error) {
	return nil, fmt.Errorf("syslog not supported on %s", runtime.GOOS)
}

// FacilityFromString converts string to syslog facility
func FacilityFromString(facility string) SyslogPriority {
	switch strings.ToLower(facility) {
	case "kern", "kernel":
		return LOG_KERN
	case "user":
		return LOG_USER
	case "mail":
		return LOG_MAIL
	case "daemon":
		return LOG_DAEMON
	case "auth":
		return LOG_AUTH
	case "syslog":
		return LOG_SYSLOG
	case "lpr":
		return LOG_LPR
	case "news":
		return LOG_NEWS
	case "uucp":
		return LOG_UUCP
	case "cron":
		return LOG_CRON
	case "authpriv":
		return LOG_AUTHPRIV
	case "ftp":
		return LOG_FTP
	case "local0":
		return LOG_LOCAL0
	case "local1":
		return LOG_LOCAL1
	case "local2":
		return LOG_LOCAL2
	case "local3":
		return LOG_LOCAL3
	case "local4":
		return LOG_LOCAL4
	case "local5":
		return LOG_LOCAL5
	case "local6":
		return LOG_LOCAL6
	case "local7":
		return LOG_LOCAL7
	default:
		return LOG_USER
	}
}
