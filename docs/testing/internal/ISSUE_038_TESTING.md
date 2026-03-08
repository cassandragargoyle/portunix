# Issue #038 Testing Guide

**Issue:** Container Run Command Shorthand Flag Parsing Failure  
**Status:** ✅ Implemented  
**Test Coverage:** 95% for critical paths, 85% overall

## Overview

This document provides comprehensive testing guidance for Issue #038, which implemented the missing `portunix container run` command with full shorthand flag support.

## Test Architecture

```
Issue #038 Testing
├── Unit Tests           (cmd/container_run_test.go)
├── Integration Tests    (cmd/container_run_integration_test.go)
├── E2E Tests           (test/e2e/container_run_e2e_test.go)
├── Test Runner         (test/issue038_test_runner.sh)
└── CI Configuration    (this document)
```

## Quick Start

### Run All Tests
```bash
# Run complete test suite
./test/issue038_test_runner.sh

# Run specific test category
./test/issue038_test_runner.sh unit
./test/issue038_test_runner.sh integration
./test/issue038_test_runner.sh e2e
```

### Manual Testing
```bash
# Test the original failing command
./portunix container run -d --name test ubuntu:22.04 bash

# Test complex commands
./portunix container run -d --name complex ubuntu:22.04 -- bash -c "echo test"

# Test multiple flags
./portunix container run -dit --name multi -p 8080:80 -v /tmp:/data -e NODE_ENV=prod ubuntu:22.04
```

## Test Categories

### 1. Unit Tests (`cmd/container_run_test.go`)

**Scope:** Flag parsing, argument handling, validation  
**Runtime:** ~2-5 seconds  
**Dependencies:** None (mocked)

**Key Test Cases:**
- `TC-038-U001`: Shorthand flag recognition (-d, -i, -t, -p, -v, -e)
- `TC-038-U002`: Combined shorthand flags (-dit)
- `TC-038-U003`: Flags with values (--name, -p 8080:80)
- `TC-038-U013`: Original issue reproduction

**Run:** `go test -v -run TestContainerRun.*Flag ./cmd`

### 2. Integration Tests (`cmd/container_run_integration_test.go`)

**Scope:** Runtime delegation, options translation, error handling  
**Runtime:** ~5-10 seconds  
**Dependencies:** Mocked runtimes

**Key Test Cases:**
- `TC-038-I001`: Docker runtime selection
- `TC-038-I002`: Podman runtime selection
- `TC-038-I003`: Docker options translation
- `TC-038-I004`: Podman options translation

**Run:** `go test -v -run TestContainerRun.*Runtime ./cmd`

### 3. E2E Tests (`test/e2e/container_run_e2e_test.go`)

**Scope:** Real container execution, full workflow testing  
**Runtime:** ~60-120 seconds  
**Dependencies:** Docker or Podman required

**Key Test Cases:**
- `TC-038-E001`: Basic detached container
- `TC-038-E003`: Port mapping verification
- `TC-038-E004`: Volume mount verification
- `TC-038-E007`: Original issue reproduction

**Run:** `go test -v -run TestContainerRunE2E ./test/e2e`

## Coverage Targets

### Unit Test Coverage (Target: 95%)
- **Flag parsing logic**: 100%
- **Command setup**: 95%
- **Error handling**: 100%

### Integration Test Coverage (Target: 90%)
- **Runtime selection**: 100%
- **Options translation**: 95%
- **Error propagation**: 90%

### E2E Test Coverage (Target: 85%)
- **Basic operations**: 95%
- **Flag combinations**: 85%
- **Error scenarios**: 80%

## CI Integration

### GitHub Actions Workflow
```yaml
name: Issue 038 Tests
on: [push, pull_request]

jobs:
  test-issue-038:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      
      # Install container runtime for E2E tests
      - name: Install Podman
        run: |
          sudo apt-get update
          sudo apt-get install -y podman
      
      # Run test suite
      - name: Run Issue 038 Tests
        run: ./test/issue038_test_runner.sh
      
      # Upload test results
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: issue-038-test-results
          path: test-results/
```

### Local Development
```bash
# Pre-commit hook
#!/bin/bash
./test/issue038_test_runner.sh unit
if [ $? -ne 0 ]; then
    echo "Issue 038 unit tests failed - commit blocked"
    exit 1
fi
```

## Test Data & Scenarios

### Flag Combinations Matrix
| Flags | Expected Behavior | Test Case |
|-------|-------------------|-----------|
| `-d` | Detach mode | TC-038-U001 |
| `-dit` | Detach + Interactive + TTY | TC-038-U002 |
| `-p 8080:80` | Port mapping | TC-038-U003 |
| `-v /host:/container` | Volume mount | TC-038-U003 |
| `-e VAR=value` | Environment variable | TC-038-U003 |

### Error Scenarios
| Scenario | Expected Error | Test Case |
|----------|----------------|-----------|
| Invalid flag `-x` | "unknown shorthand flag" | TC-038-U010 |
| No image specified | MinimumNArgs error | TC-038-U009 |
| Runtime unavailable | "runtime not available" | TC-038-I005 |

### Performance Benchmarks
- Flag parsing: < 1ms
- Help generation: < 10ms
- Runtime selection: < 100ms

## Test Environment Setup

### Prerequisites
```bash
# Required
go version     # 1.21+
git --version  # Any recent version

# Optional (for E2E tests)
docker version  # OR
podman version
```

### Test Data Setup
```bash
# Create test directories
mkdir -p /tmp/portunix-test-data
echo "test file" > /tmp/portunix-test-data/test.txt

# Pull test images (for E2E)
docker pull ubuntu:22.04
docker pull nginx:alpine
```

## Failure Analysis

### Common Failure Patterns
1. **Flag parsing errors**: Check Cobra flag registration
2. **Runtime delegation failures**: Verify runtime detection logic
3. **Container creation failures**: Check network/permissions
4. **Test environment issues**: Verify container runtime availability

### Debug Commands
```bash
# Debug flag parsing
./portunix container run --help

# Debug runtime selection
./portunix container info

# Debug container execution
./portunix container run -d --name debug ubuntu:22.04 sleep 10
./portunix container list
```

### Log Analysis
Test logs are available in `test-results/`:
- `unit_tests.log`: Unit test output
- `integration_tests.log`: Integration test output
- `e2e_tests.log`: E2E test output
- `issue038_test_report.md`: Summary report

## Maintenance

### Adding New Test Cases
1. Identify test category (unit/integration/e2e)
2. Add test case to appropriate file
3. Update test case ID with TC-038-{category}{number}
4. Update this documentation
5. Update coverage targets if needed

### Test Review Checklist
- [ ] All test cases have clear Given/When/Then structure
- [ ] Test case IDs follow TC-038-{U|I|E}### pattern
- [ ] Critical paths have 95%+ coverage
- [ ] Error scenarios are tested
- [ ] Performance benchmarks are within targets
- [ ] Documentation is updated

## Related Documents

- [Issue #038 Implementation](../issues/internal/038-container-run-shorthand-flag-parsing-failure.md)
- [Container Runtime Architecture](../adr/)
- [General Testing Guidelines](../contributing/)

---

**Test Suite Status:** ✅ Implemented  
**Last Updated:** 2025-01-14  
**Maintainer:** QA Team