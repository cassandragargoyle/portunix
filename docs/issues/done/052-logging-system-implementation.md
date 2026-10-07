# Issue #52: Logging System Implementation

**Status:** 📋 Open
**Priority:** Critical
**Type:** Enhancement
**Labels:** enhancement, logging, architecture, mcp, critical
**Created:** 2025-01-20
**Author:** Architect
**ADR:** [ADR-015: Logging System Architecture](../../adr/015-logging-system-architecture.md)

## Summary

Implement comprehensive logging system for Portunix to replace direct `fmt.Print*` statements with structured, leveled logging that supports proper debugging, production monitoring, and clean STDIO separation for MCP protocol.

## Background

Portunix currently lacks a proper logging system, which causes:
- No debug capability without code changes
- MCP protocol STDIO conflicts
- Poor error tracking in production
- Difficult container/VM debugging
- Cluttered test output

See [ADR-015](../../adr/015-logging-system-architecture.md) for detailed architecture decision.

## Requirements

### Functional Requirements

1. **Structured Logging**
   - JSON and text format support
   - Log levels: TRACE, DEBUG, INFO, WARN, ERROR, FATAL, PANIC
   - Contextual fields (timestamp, module, correlation ID)

2. **Multiple Output Targets**
   - Console output
   - File output with rotation
   - Syslog support
   - Clean STDIO for MCP mode

3. **Configuration**
   - Runtime log level changes
   - Per-module log level control
   - Environment variable overrides
   - Config file support

4. **Special Modes**
   - MCP server mode (file/syslog only)
   - Container detection and context
   - Test mode simplification

### Non-Functional Requirements

1. **Performance**
   - Minimal performance impact
   - Zero allocations in hot paths
   - Async logging for file output

2. **Compatibility**
   - Legacy output mode for scripts
   - Gradual migration support
   - Feature flags for rollout

## Implementation Plan

### Phase 1: Infrastructure (Week 1)
- [ ] Create `pkg/logging` package structure
- [ ] Implement logger factory with zerolog
- [ ] Add configuration management
- [ ] Create output handlers (console, file, syslog)
- [ ] Write unit tests for logging package

### Phase 2: Critical Paths (Week 2)
- [ ] Replace fmt.Print in error handling paths
- [ ] Add logging to MCP server (clean STDIO)
- [ ] Add logging to container operations
- [ ] Add logging to plugin system
- [ ] Validate MCP protocol compatibility

### Phase 3: Full Migration (Weeks 3-4)
- [ ] Systematic replacement of all fmt.Print statements
- [ ] Add debug logging to all major operations
- [ ] Update test framework for clean output
- [ ] Add correlation ID support
- [ ] Document logging conventions

### Phase 4: Enhancement (Week 5)
- [ ] Implement log rotation
- [ ] Add performance metrics logging
- [ ] Create log analysis tools
- [ ] Add distributed tracing support
- [ ] Complete migration guide

## Technical Details

### Package Structure
```
pkg/logging/
├── logger.go          # Main logger interface
├── factory.go         # Logger factory
├── config.go          # Configuration management
├── context.go         # Context propagation
├── handlers/
│   ├── console.go     # Console output
│   ├── file.go        # File output with rotation
│   └── syslog.go      # Syslog output
└── middleware/
    └── correlation.go # Correlation ID middleware
```

### Usage Examples

```go
// Basic usage
log := logging.New("component")
log.Info("Operation started")
log.Error("Operation failed", "error", err)

// Structured logging
log.With().
    Str("user", username).
    Int("port", 8080).
    Info("Server started")

// Context propagation
ctx = log.WithContext(ctx)
```

### Configuration Example

```yaml
logging:
  level: info
  format: json
  output:
    - console
    - file
  file_path: /var/log/portunix/app.log
  no_color: false
  modules:
    mcp: warn
    docker: debug
```

## Testing Requirements

1. **Unit Tests**
   - Logger factory creation
   - Configuration parsing
   - Output handler behavior
   - Context propagation

2. **Integration Tests**
   - MCP server STDIO separation
   - Container context detection
   - File rotation under load
   - Performance benchmarks

3. **E2E Tests**
   - Command execution with logging
   - Error scenarios with proper logging
   - Multi-level verbosity

## Migration Strategy

1. **Preparation**
   - Create logging package
   - Document conventions
   - Create migration tools

2. **Gradual Rollout**
   - Enable with feature flag
   - Start with new code only
   - Migrate critical paths first

3. **Full Migration**
   - Automated replacement script
   - Manual review of complex cases
   - Update all documentation

4. **Cleanup**
   - Remove old print statements
   - Remove feature flags
   - Archive migration tools

## Success Criteria

- [ ] All fmt.Print statements replaced
- [ ] MCP server runs with clean STDIO
- [ ] Debug mode available without code changes
- [ ] Log levels configurable at runtime
- [ ] Performance impact < 5% in benchmarks
- [ ] All tests pass with new logging
- [ ] Documentation complete

## Dependencies

- **External**: zerolog library
- **Internal**: All modules need updates
- **Breaking Changes**: Potential output format changes

## Risk Analysis

| Risk | Impact | Mitigation |
|------|--------|------------|
| Performance degradation | High | Benchmark before/after, use zero-allocation logger |
| Breaking script compatibility | Medium | Provide legacy output mode |
| MCP protocol issues | High | Extensive testing with AI assistants |
| Migration complexity | Medium | Automated tools, gradual rollout |

## References

- [ADR-015: Logging System Architecture](../../adr/015-logging-system-architecture.md)
- [Issue #037: MCP Server Implementation](037-revert-default-stdio-mcp-implement-mcp-serve.md)
- [Issue #051: Git Dispatcher Architecture](../051-git-dispatcher-python-distribution-architecture.md)
- [Zerolog Documentation](https://github.com/rs/zerolog)

## Acceptance Criteria

- [ ] Logging package implemented and tested
- [ ] Critical paths migrated (MCP, containers, plugins)
- [ ] All fmt.Print statements replaced
- [ ] Documentation and examples provided
- [ ] Performance benchmarks acceptable
- [ ] MCP protocol working correctly
- [ ] Tests updated and passing

## Notes

This is a critical infrastructure change that affects the entire codebase. The implementation should be done carefully with extensive testing at each phase. The MCP server compatibility is particularly critical as it requires clean STDIO separation.

Priority should be given to:
1. MCP server logging (Issue #037)
2. Container operations (Issues #029-032)
3. Plugin system (Issue #007)
4. Error handling paths