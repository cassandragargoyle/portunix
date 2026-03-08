# Test Plan - Issue #052: Logging System Implementation

**Issue**: #052 - Logging System Implementation
**Feature Branch**: `feature/issue-052-logging-system`
**Tester**: QA/Test Engineer (Linux)
**Date**: 2025-09-20
**Testing OS**: Linux (host system validation required)

## Plan

### Testing Strategy
- **Unit Testing**: Core logging components (factory, config, handlers)
- **Integration Testing**: MCP STDIO separation and container context detection
- **E2E Testing**: Full command execution with logging enabled
- **Performance Testing**: Benchmarks for performance impact validation
- **Container Testing**: Using Portunix native container commands

### Test Coverage Targets
- **Unit Tests**: 95%+ coverage for `pkg/logging` package
- **Integration Tests**: Critical paths (MCP, containers, plugins)
- **E2E Tests**: Full user workflows with logging enabled
- **Performance**: <5% performance impact as per requirements

## Cases

### Unit Test Cases

#### TC-001: Logger Factory Creation
**Given**: Default logging configuration
**When**: Creating logger factory with NewFactory()
**Then**: Factory instance created with valid configuration

#### TC-002: Logger Creation with Component Name
**Given**: Logger factory with default config
**When**: Creating logger with component name "mcp-server"
**Then**: Logger instance includes component field in all log entries

#### TC-003: Log Level Configuration
**Given**: Config with different log levels
**When**: Setting levels (TRACE, DEBUG, INFO, WARN, ERROR, FATAL)
**Then**: Only messages at or above configured level are output

#### TC-004: Multiple Output Targets
**Given**: Config with console, file, and syslog outputs
**When**: Logging messages
**Then**: Messages appear in all configured outputs

#### TC-005: JSON vs Text Format
**Given**: Config with format set to "json" or "text"
**When**: Logging structured messages
**Then**: Output format matches configuration

#### TC-006: Environment Variable Override
**Given**: PORTUNIX_LOG_LEVEL=debug environment variable
**When**: Loading configuration
**Then**: Log level is set to debug regardless of default config

#### TC-007: Per-Module Log Levels
**Given**: Config with modules: {"mcp": "warn", "docker": "debug"}
**When**: Creating loggers for mcp and docker components
**Then**: Each logger uses its specific level

#### TC-008: Context Propagation
**Given**: Logger with correlation ID
**When**: Extracting logger from context
**Then**: Correlation ID is preserved in all subsequent log entries

#### TC-009: Container Environment Detection
**Given**: Running inside Docker container
**When**: Creating logger
**Then**: Container metadata (ID, container=true) added to log entries

#### TC-010: File Output and Directory Creation
**Given**: Config with file output to non-existent directory
**When**: Creating logger
**Then**: Directory is created and log file is written

### Integration Test Cases

#### TC-011: MCP Server STDIO Separation
**Given**: MCP server started with logging enabled
**When**: Server processes MCP protocol messages
**Then**:
- MCP protocol uses STDOUT/STDIN cleanly
- Log messages go to file/syslog only
- No interference with JSON-RPC communication

#### TC-012: Container Context Detection
**Given**: Portunix running in Ubuntu container via `portunix docker run-in-container`
**When**: Logging system initializes
**Then**:
- Container environment detected
- Container ID extracted and logged
- Container flag set to true in all log entries

#### TC-013: Plugin System Integration
**Given**: Plugin loaded with logging enabled
**When**: Plugin operations execute
**Then**: Plugin logs include component field and proper context

#### TC-014: Error Handling Paths
**Given**: Critical error scenarios (file not found, permission denied)
**When**: Error occurs with logging enabled
**Then**:
- Error details logged with appropriate level
- Stack trace included for debug level
- Operation context preserved

#### TC-015: Dynamic Log Level Changes
**Given**: Running system with INFO level
**When**: Changing log level to DEBUG at runtime
**Then**:
- New log entries use DEBUG level
- Existing loggers updated
- Configuration persisted

### E2E Test Cases

#### TC-016: Command Execution with Logging
**Given**: Portunix with logging configuration
**When**: Executing `portunix install nodejs --verbose`
**Then**:
- Installation steps logged with appropriate levels
- Debug information available when verbose
- No interference with normal output

#### TC-017: Multi-Container Environment
**Given**: Multiple containers created via Portunix
**When**: Operations executed in different containers
**Then**:
- Each container context properly identified
- Log entries distinguish between containers
- Container lifecycle events logged

#### TC-018: MCP Integration with AI Assistant
**Given**: MCP server running with logging
**When**: AI assistant connects and executes commands
**Then**:
- MCP protocol works without interference
- Server operations logged to file
- AI assistant receives clean JSON responses

#### TC-019: Performance Impact Validation
**Given**: Large installation operation (default profile)
**When**: Comparing execution with/without logging
**Then**: Performance impact is <5% as specified

#### TC-020: Log Rotation and File Management
**Given**: High-volume logging over extended period
**When**: Log files reach size limits
**Then**:
- Old logs rotated automatically
- Maximum file count respected
- No disk space exhaustion

### Failure Injection Cases

#### TC-021: File System Failures
**Given**: No write permissions to log directory
**When**: Attempting to initialize file logging
**Then**:
- Graceful fallback to console logging
- Error logged about permission issue
- System continues to operate

#### TC-022: Invalid Configuration
**Given**: Malformed JSON configuration file
**When**: Loading logging configuration
**Then**:
- Default configuration used
- Warning logged about invalid config
- System starts successfully

#### TC-023: Memory Pressure
**Given**: High-frequency logging under memory pressure
**When**: System resources constrained
**Then**:
- Logging remains stable
- No memory leaks detected
- Performance degrades gracefully

#### TC-024: Container Detection Failures
**Given**: Container environment with unusual setup
**When**: Container detection logic runs
**Then**:
- Falls back to host environment logging
- No crashes or errors
- Manual container flag can be set

## Coverage

### Unit Test Coverage Matrix

| Component | Function | Test Case | Expected Coverage |
|-----------|----------|-----------|-------------------|
| Logger | All public methods | TC-001 to TC-010 | 95%+ |
| Factory | CreateLogger, CreateMCPLogger | TC-001, TC-002 | 90%+ |
| Config | Load, Validate, Clone | TC-003, TC-006 | 95%+ |
| Handlers | Console, File, Syslog | TC-004, TC-010 | 85%+ |
| Context | Propagation, Extraction | TC-008 | 90%+ |

### Integration Test Coverage

| System Component | Integration Point | Test Case | Priority |
|------------------|-------------------|-----------|----------|
| MCP Server | STDIO separation | TC-011 | Critical |
| Container System | Context detection | TC-012 | High |
| Plugin System | Component logging | TC-013 | High |
| Error Handling | Critical paths | TC-014 | Critical |
| Runtime Config | Dynamic changes | TC-015 | Medium |

### E2E Test Coverage

| User Workflow | Scenario | Test Case | Business Impact |
|---------------|----------|-----------|-----------------|
| Package Installation | Verbose logging | TC-016 | High |
| Container Operations | Multi-container | TC-017 | High |
| AI Integration | MCP protocol | TC-018 | Critical |
| Performance | Long operations | TC-019 | Critical |
| Maintenance | Log rotation | TC-020 | Medium |

## CI Notes

### Automated Test Execution
```bash
# Unit tests with coverage
go test ./pkg/logging/... -v -coverprofile=coverage.out -covermode=atomic

# Integration tests using TestFramework
go test ./test/integration/issue_052_logging_*_test.go -v -timeout=10m

# E2E tests in containers
CONTAINER_TEST=true go test ./test/integration/issue_052_logging_e2e_test.go -v -timeout=15m

# Performance benchmarks
go test ./pkg/logging/... -bench=. -benchmem -timeout=30m
```

### Container Test Requirements
- **MANDATORY**: All integration tests MUST use Portunix container commands
- Use `portunix docker run-in-container` for installation testing
- Test across multiple container platforms (Ubuntu 22.04, Debian, Alpine)
- Document container OS in test results, not host OS

### Performance Benchmarks
```bash
# Baseline measurements
go test ./pkg/logging/... -bench=BenchmarkLogger -benchtime=10s -count=5

# Memory allocation tracking
go test ./pkg/logging/... -bench=. -benchmem -memprofile=mem.prof

# Performance comparison script
./scripts/benchmark-logging-impact.sh
```

### CI Pipeline Integration
- Tests must pass on Linux platform (tester requirement)
- Container tests validate cross-platform compatibility
- Performance regression gates (>5% impact fails build)
- Coverage gates (unit: >95%, integration: >80%)

## Acceptance Protocol

### Pre-Test Checklist
- [ ] Feature branch `feature/issue-052-logging-system` checked out
- [ ] Binary built successfully: `go build -o portunix`
- [ ] Testing environment validated: Linux host system
- [ ] Container runtime available (Docker/Podman via Portunix)
- [ ] Test data prepared and environment clean

### Unit Test Execution
- [ ] All unit tests pass: `go test ./pkg/logging/... -v`
- [ ] Code coverage ≥95%: Coverage report generated
- [ ] No memory leaks: Memory profile clean
- [ ] Benchmark performance: Baseline established

### Integration Test Execution
- [ ] MCP STDIO separation: TC-011 PASS
- [ ] Container context detection: TC-012 PASS
- [ ] Plugin system integration: TC-013 PASS
- [ ] Error handling paths: TC-014 PASS
- [ ] Dynamic configuration: TC-015 PASS

### E2E Test Execution
- [ ] Command execution logging: TC-016 PASS
- [ ] Multi-container operations: TC-017 PASS
- [ ] AI assistant integration: TC-018 PASS
- [ ] Performance impact <5%: TC-019 PASS
- [ ] Log rotation functional: TC-020 PASS

### Failure Injection Testing
- [ ] File system failures: TC-021 PASS
- [ ] Invalid configuration: TC-022 PASS
- [ ] Memory pressure: TC-023 PASS
- [ ] Container detection edge cases: TC-024 PASS

### Final Validation
- [ ] All acceptance criteria from Issue #052 met
- [ ] No regression in existing functionality
- [ ] Documentation updated and accurate
- [ ] Performance benchmarks within requirements
- [ ] Container testing completed successfully

### Security Validation
- [ ] No sensitive data in log outputs
- [ ] File permissions appropriate (644 for logs, 755 for directories)
- [ ] No log injection vulnerabilities
- [ ] Correlation IDs properly sanitized

### Approval Criteria
- **PASS**: All critical test cases pass, performance <5% impact
- **FAIL**: Any critical test case fails or performance >5% impact
- **CONDITIONAL**: Minor issues with documented workarounds

---

**Note**: This test plan follows the Issue Development Methodology requirement that testers must validate Linux host system for local tests and document container OS for containerized tests.