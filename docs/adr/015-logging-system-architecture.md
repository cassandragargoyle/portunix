# ADR-015: Logging System Architecture

**Status**: Proposed
**Date**: 2025-01-20
**Author**: Architect

## Context

Portunix currently lacks a comprehensive logging system. All output is handled through direct `fmt.Print*` statements to stdout/stderr, which presents several critical issues:

1. **No Debug Capability**: Cannot enable detailed debugging output without modifying code
2. **No Log Levels**: Cannot differentiate between errors, warnings, info, and debug messages
3. **No Structured Logging**: Cannot output logs in structured format (JSON) for log aggregation
4. **No File Logging**: All output goes to console only, no persistent logs for troubleshooting
5. **Poor Error Tracking**: Difficult to diagnose issues in production environments
6. **MCP Integration Issues**: MCP protocol requires clean STDIO, current print statements interfere
7. **Testing Difficulties**: Test output is cluttered with application output
8. **No Context**: Missing timestamps, source locations, correlation IDs
9. **Container/VM Debugging**: Hard to debug issues in containerized or virtualized environments

### Current State Analysis
- **fmt.Print usage**: Found in 100+ files across the codebase
- **Error handling**: Inconsistent error reporting directly to stdout
- **Debug output**: Hardcoded verbose flags in some modules (TestFramework)
- **MCP Server**: Requires clean STDIO but mixed with debug output

## Decision

Implement a centralized, structured logging system with the following architecture:

### 1. Logging Framework Selection
Use **zerolog** as the primary logging framework:
- High performance with zero allocations
- Structured logging with JSON output
- Context-aware logging
- Log level support
- Multiple output targets

### 2. Architecture Components

```
┌─────────────────────────────────────────────────┐
│             Application Layer                    │
├─────────────────────────────────────────────────┤
│                                                  │
│  ┌──────────────┐        ┌──────────────┐      │
│  │   Commands    │───────▶│   Logger     │      │
│  │              │        │   Instance   │      │
│  └──────────────┘        └──────────────┘      │
│                                 │               │
├─────────────────────────────────┼───────────────┤
│            Logging Layer        ▼               │
├──────────────────────────────────────────────────┤
│                                                  │
│  ┌──────────────┐        ┌──────────────┐      │
│  │   Config     │───────▶│   Factory    │      │
│  │              │        │              │      │
│  │ - Level      │        │ - Console    │      │
│  │ - Format     │        │ - File       │      │
│  │ - Outputs    │        │ - Syslog     │      │
│  └──────────────┘        └──────────────┘      │
│                                                  │
│  ┌──────────────────────────────────────┐      │
│  │          Output Handlers              │      │
│  ├──────────────────────────────────────┤      │
│  │ Console │ File │ Syslog │ JSON │     │      │
│  └──────────────────────────────────────┘      │
└──────────────────────────────────────────────────┘
```

### 3. Log Levels
- **TRACE**: Detailed execution flow
- **DEBUG**: Debugging information
- **INFO**: General information
- **WARN**: Warning conditions
- **ERROR**: Error conditions
- **FATAL**: Fatal errors causing exit
- **PANIC**: Panic conditions

### 4. Configuration
```go
type LogConfig struct {
    Level      string   `json:"level"`      // trace, debug, info, warn, error
    Format     string   `json:"format"`     // text, json
    Output     []string `json:"output"`     // console, file, syslog
    FilePath   string   `json:"file_path"`  // for file output
    TimeFormat string   `json:"time_format"`
    NoColor    bool     `json:"no_color"`   // disable color output
}
```

### 5. Context Fields
Every log entry should include:
- Timestamp
- Level
- Module/Component
- Function/Method
- Correlation ID (for request tracing)
- Platform info (OS, arch)
- Container/VM context if applicable

### 6. Special Handling

#### MCP Server Mode
When running as MCP server, redirect all logs to file or syslog:
```go
if runningAsMCP {
    logger.SetOutput(file)
    logger.SetLevel(zerolog.WarnLevel)
}
```

#### Container/VM Environments
Auto-detect and add context:
```go
if inContainer {
    logger = logger.With().Str("container", containerID).Logger()
}
```

#### Test Mode
Simplified output for tests:
```go
if testing.Testing() {
    logger.SetLevel(zerolog.ErrorLevel)
}
```

### 7. Migration Strategy

Phase 1: Infrastructure (Week 1)
- Implement logging package
- Create logger factory
- Add configuration support

Phase 2: Critical Paths (Week 2)
- Replace fmt.Print in error paths
- Add logging to MCP server
- Add logging to container operations

Phase 3: Full Migration (Week 3-4)
- Systematic replacement of all fmt.Print
- Add debug logging throughout
- Update tests

Phase 4: Enhancement (Week 5)
- Add correlation IDs
- Implement log rotation
- Add performance metrics

## Consequences

### Positive
- **Better Debugging**: Ability to enable verbose logging without code changes
- **Production Ready**: Proper logging for production environments
- **MCP Compatibility**: Clean separation of protocol and debug output
- **Structured Logs**: Machine-readable logs for analysis
- **Performance**: Zerolog has minimal performance impact
- **Testability**: Cleaner test output
- **Observability**: Better system observability and monitoring
- **Error Tracking**: Centralized error tracking and analysis

### Negative
- **Dependency**: Adds external dependency (zerolog)
- **Migration Effort**: Significant effort to replace all print statements
- **Learning Curve**: Team needs to learn new logging patterns
- **Configuration Complexity**: Additional configuration management
- **Backward Compatibility**: May affect scripts parsing current output

### Mitigation
- Provide compatibility mode for legacy output format
- Create migration guide and examples
- Implement gradually with feature flags
- Maintain simple API for common cases

## Implementation Details

### Package Structure
```
pkg/
└── logging/
    ├── logger.go       # Main logger interface
    ├── factory.go      # Logger factory
    ├── config.go       # Configuration
    ├── context.go      # Context management
    ├── handlers/       # Output handlers
    │   ├── console.go
    │   ├── file.go
    │   └── syslog.go
    └── middleware/     # HTTP/gRPC middleware
        └── correlation.go
```

### Usage Example
```go
// Initialize
log := logging.New("component-name")

// Simple logging
log.Info("Operation started")
log.Error("Operation failed", "error", err)

// Structured logging
log.With().
    Str("user", username).
    Int("port", 8080).
    Info("Server started")

// Context propagation
ctx = log.WithContext(ctx)
logFromCtx := logging.FromContext(ctx)
```

## Alternatives Considered

1. **Standard log package**: Too basic, no structure or levels
2. **logrus**: Good but slower than zerolog
3. **zap**: Fast but more complex API
4. **Custom solution**: Too much maintenance overhead
5. **No logging**: Current state, not sustainable

## References

- Issue #051: Git Dispatcher and Python Distribution (needs clean STDIO)
- Issue #037: MCP Server Implementation (STDIO conflicts)
- [Zerolog Documentation](https://github.com/rs/zerolog)
- [12-Factor App Logging](https://12factor.net/logs)
- [Structured Logging Best Practices](https://www.structlog.org/)

## Decision Outcome

**APPROVED** - Implementation should begin immediately due to critical need for proper logging in production environments and MCP integration requirements.

## Review Schedule

This ADR should be reviewed after Phase 2 implementation to assess if the chosen approach meets requirements and adjust if necessary.