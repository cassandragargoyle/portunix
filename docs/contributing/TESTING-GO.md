# Go Testing Guidelines

> **IMPORTANT**: This document is distributed from BOOTSTRAP-SCRIPTS and MUST NOT be modified in individual projects as it will be automatically overwritten during updates. Each project should have a project-specific testing document (e.g., `TESTING-GO-PROJECT.md`) for project-specific differences or additions.

## Purpose
This document defines the testing strategy, structure, and best practices for Go projects in the CassandraGargoyle ecosystem.

## Testing Philosophy

### Testing Pyramid
1. **Unit Tests** (70%) - Fast, isolated tests for individual functions/methods (Go)
2. **Integration Tests** (20%) - Test component interactions (Python for complex scenarios)
3. **End-to-End Tests** (10%) - Full system workflow tests (Python)
4. **Smoke Tests** - Quick verification of basic functionality (Bash)

### Key Principles
- Write tests first (TDD approach when possible)
- Tests should be fast, reliable, and independent
- Each test should test one specific behavior
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- **Container Testing**: Use Podman for all container-based testing (preferred over Docker)
- **Language Selection**: Choose appropriate language for each test type:
  - **Go**: Unit tests, internal package testing
  - **Python**: Integration/E2E tests with complex scenarios (VM, SSH, containers)
  - **Bash**: Simple smoke tests and CLI verification

## Project Structure

### Directory Layout
```
project/
├── cmd/                    # Application entry points
├── pkg/                    # Public packages
├── internal/               # Private packages
├── test/                   # Test utilities and fixtures
│   ├── unit/               # Go unit tests (*_test.go)
│   ├── integration/        # Python integration tests
│   │   ├── conftest.py     # pytest configuration
│   │   ├── test_*.py       # Python test files
│   │   └── fixtures/       # Python test fixtures
│   ├── e2e/                # Python end-to-end tests
│   │   ├── test_*.py       # E2E test scenarios
│   │   └── helpers/        # Python test helpers
│   ├── smoke/              # Bash smoke tests
│   │   ├── test_*.sh       # Smoke test scripts
│   │   └── common.sh       # Shared bash functions
│   ├── fixtures/           # Shared test data files
│   ├── mocks/              # Generated mocks (Go)
│   ├── testdata/           # Test data for specific scenarios
│   └── scripts/            # Test execution scripts
│       ├── test.sh         # Main test runner (all languages)
│       ├── test-unit.sh    # Go unit tests only
│       ├── test-integration.py # Python integration tests
│       ├── test-e2e.py     # Python E2E tests
│       ├── test-smoke.sh   # Bash smoke tests
│       └── coverage.sh     # Coverage reporting
├── scripts/                # Build and deployment scripts
├── requirements-test.txt   # Python test dependencies
├── pytest.ini             # pytest configuration
└── go.mod
```

## Test Naming Conventions

### Test Files
- **Go unit tests**: `*_test.go` (in same package directory)
- **Go benchmark tests**: `*_benchmark_test.go`
- **Python integration tests**: `test_*.py` (in test/integration/)
- **Python E2E tests**: `test_*.py` (in test/e2e/)
- **Bash smoke tests**: `test_*.sh` (in test/smoke/)

### Test Functions
Use descriptive names that explain what is being tested.

*Code examples will be added after initial test implementation is completed.*

### Test Structure Pattern
Use **Arrange, Act, Assert** pattern.

*Code examples will be added after initial test implementation is completed.*

## Unit Testing

### Simple Function Tests
Use basic `testing` package for straightforward unit tests.

### Test Suite Pattern  
For complex integration tests or when you need setup/teardown, use testify/suite:

```go
type MyTestSuite struct {
    suite.Suite
    // shared resources
}

func (suite *MyTestSuite) SetupSuite() {
    // setup before all tests
}

func (suite *MyTestSuite) TearDownSuite() {
    // cleanup after all tests  
}

func (suite *MyTestSuite) SetupTest() {
    // setup before each test
}

func (suite *MyTestSuite) TearDownTest() {
    // cleanup after each test
}

func (suite *MyTestSuite) TestExample() {
    suite.Equal(expected, actual)
}

func TestMyTestSuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}
```

*Additional examples will be added after first test implementation.*

## Integration Testing

### Simple Table-Driven Approach
For straightforward integration tests, use basic Go testing patterns:

```go
func TestIssue012_PowerShellInstallation_AllDistros(t *testing.T) {
    supportedDistros := []struct {
        name    string
        image   string
        variant string
    }{
        {"Ubuntu 22.04", "ubuntu:22.04", "ubuntu"},
        {"Ubuntu 24.04", "ubuntu:24.04", "ubuntu"},
        {"Debian 11", "debian:bullseye", "debian"},
        {"Debian 12", "debian:bookworm", "debian"},
        {"Fedora 39", "fedora:39", "fedora"},
        {"Fedora 40", "fedora:40", "fedora"},
        {"Rocky Linux 9", "rockylinux:9", "rocky"},
        {"Linux Mint 21", "linuxmintd/mint21-amd64", "mint"},
    }

    for _, distro := range supportedDistros {
        t.Run(distro.name, func(t *testing.T) {
            t.Parallel()
            testPowerShellInstallationOnDistro(t, distro.image, distro.variant)
        })
    }
}
```

### Test Suite Approach for Complex Integration Tests
When you need shared setup/teardown, database connections, or complex resource management:

```go
type PowerShellIntegrationSuite struct {
    suite.Suite
    dockerClient *docker.Client
    sshClient    *ssh.Client
    containers   []string
}

func (suite *PowerShellIntegrationSuite) SetupSuite() {
    // Initialize shared resources
    suite.dockerClient = docker.NewClient()
    suite.containers = []string{}
}

func (suite *PowerShellIntegrationSuite) TearDownSuite() {
    // Cleanup all containers
    for _, containerID := range suite.containers {
        suite.dockerClient.RemoveContainer(containerID)
    }
}

func (suite *PowerShellIntegrationSuite) TestPowerShellInstallation() {
    supportedDistros := []struct {
        name    string
        image   string  
        variant string
    }{
        {"Ubuntu 22.04", "ubuntu:22.04", "ubuntu"},
        {"Ubuntu 24.04", "ubuntu:24.04", "ubuntu"},
        {"Debian 11", "debian:bullseye", "debian"},
        {"Debian 12", "debian:bookworm", "debian"},
        {"Fedora 39", "fedora:39", "fedora"},
        {"Fedora 40", "fedora:40", "fedora"},
        {"Rocky Linux 9", "rockylinux:9", "rocky"},
        {"Linux Mint 21", "linuxmintd/mint21-amd64", "mint"},
    }

    for _, distro := range supportedDistros {
        suite.Run(distro.name, func() {
            suite.testPowerShellOnDistro(distro.image, distro.variant)
        })
    }
}

func TestPowerShellIntegrationSuite(t *testing.T) {
    suite.Run(t, new(PowerShellIntegrationSuite))
}
```

### When to Use Each Approach

**Simple Table-Driven Tests:**
- Independent test cases
- No complex setup/teardown needed
- Fast, lightweight tests
- Parallel execution is straightforward

**Test Suite Pattern:**
- Shared expensive resources (DB connections, containers)
- Complex setup/teardown logic
- Need for test fixtures across multiple tests
- Sequential execution requirements

### Container-Based Testing

**Container Engine Preference:**
- **Always use Podman** for container-based testing
- Podman provides better security with rootless operation
- Native systemd integration and daemonless architecture
- Compatible with existing Docker-based test infrastructure

**Container Test Setup:**
```go
// Use podman commands instead of docker
func setupTestContainer(t *testing.T, image string) string {
    cmd := exec.Command("podman", "run", "-d", "--name", "test-container", image)
    output, err := cmd.Output()
    require.NoError(t, err)
    return strings.TrimSpace(string(output))
}

func cleanupTestContainer(t *testing.T, containerID string) {
    cmd := exec.Command("podman", "rm", "-f", containerID)
    cmd.Run() // Ignore errors during cleanup
}
```

*Additional examples will be added after first test implementation.*

## Python Integration & E2E Testing

### Setup Python Test Environment
```bash
# Install Python test dependencies
pip install -r requirements-test.txt

# Run Python tests
pytest test/integration/ -v
pytest test/e2e/ -v
```

### Python Test Structure
```python
# test/integration/test_powershell_installation.py
import pytest
import subprocess
import tempfile
from pathlib import Path

class TestPowerShellInstallation:
    """Integration tests for PowerShell installation across Linux distributions."""
    
    @pytest.fixture(autouse=True)
    def setup_container(self):
        """Setup and teardown test container."""
        self.container_id = None
        yield
        if self.container_id:
            subprocess.run(['podman', 'rm', '-f', self.container_id], 
                         capture_output=True)
    
    @pytest.mark.parametrize("distro,image", [
        ("Ubuntu 22.04", "ubuntu:22.04"),
        ("Ubuntu 24.04", "ubuntu:24.04"),
        ("Debian 11", "debian:bullseye"),
        ("Debian 12", "debian:bookworm"),
        ("Fedora 39", "fedora:39"),
        ("Fedora 40", "fedora:40"),
        ("Rocky Linux 9", "rockylinux:9"),
        ("Linux Mint 21", "linuxmintd/mint21-amd64"),
    ])
    def test_powershell_installation_on_distro(self, distro, image):
        """Test PowerShell installation on specific Linux distribution."""
        # Create container with Podman
        self.container_id = self._create_container(image)
        
        # Deploy Portunix via SSH
        self._deploy_portunix(self.container_id)
        
        # Install PowerShell
        result = self._install_powershell(self.container_id)
        assert result.returncode == 0, f"PowerShell installation failed on {distro}"
        
        # Verify PowerShell is working
        assert self._verify_powershell(self.container_id), f"PowerShell verification failed on {distro}"
    
    def _create_container(self, image: str) -> str:
        """Create and start test container."""
        result = subprocess.run(
            ['podman', 'run', '-d', '--name', f'test-{image}', image, 'sleep', 'infinity'],
            capture_output=True, text=True, check=True
        )
        return result.stdout.strip()
    
    def _deploy_portunix(self, container_id: str) -> None:
        """Deploy Portunix to container via SSH simulation."""
        # Implementation for SSH deployment
        pass
    
    def _install_powershell(self, container_id: str) -> subprocess.CompletedProcess:
        """Install PowerShell in container."""
        return subprocess.run(
            ['podman', 'exec', container_id, './portunix', 'install', 'powershell'],
            capture_output=True, text=True
        )
    
    def _verify_powershell(self, container_id: str) -> bool:
        """Verify PowerShell installation."""
        result = subprocess.run(
            ['podman', 'exec', container_id, 'pwsh', '--version'],
            capture_output=True, text=True
        )
        return result.returncode == 0 and 'PowerShell' in result.stdout
```

### Python Test Configuration
```ini
# pytest.ini
[tool:pytest]
testpaths = test/integration test/e2e
python_files = test_*.py
python_classes = Test*
python_functions = test_*
markers =
    slow: marks tests as slow (deselect with '-m "not slow"')
    integration: marks tests as integration tests
    e2e: marks tests as end-to-end tests
    podman: marks tests requiring Podman
addopts = -v --strict-markers
```

```txt
# requirements-test.txt
pytest>=7.0.0
pytest-xdist>=3.0.0  # Parallel test execution
pytest-mock>=3.6.0   # Mocking support
requests>=2.28.0     # HTTP requests
paramiko>=2.9.0      # SSH connections
pexpect>=4.8.0       # Interactive command testing
```

## Bash Smoke Testing

### Smoke Test Structure
```bash
#!/bin/bash
# test/smoke/test_basic_functionality.sh

# Import common functions
source "$(dirname "$0")/common.sh"

test_portunix_version() {
    info "Testing Portunix version command"
    
    local result
    result=$(./portunix version 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        success "Version command succeeded: $result"
        return 0
    else
        error "Version command failed with exit code $exit_code: $result"
        return 1
    fi
}

test_portunix_help() {
    info "Testing Portunix help command"
    
    local result
    result=$(./portunix --help 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && echo "$result" | grep -q "Usage:"; then
        success "Help command succeeded"
        return 0
    else
        error "Help command failed or missing usage information"
        return 1
    fi
}

test_powershell_help_parsing() {
    info "Testing PowerShell specific help"
    
    local result
    result=$(./portunix install powershell --help 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && echo "$result" | grep -q "PowerShell"; then
        success "PowerShell help command succeeded"
        return 0
    else
        error "PowerShell help command failed or missing PowerShell information"
        return 1
    fi
}

# Run all tests
main() {
    info "Starting Portunix smoke tests"
    
    local failed=0
    
    test_portunix_version || ((failed++))
    test_portunix_help || ((failed++))
    test_powershell_help_parsing || ((failed++))
    
    if [ $failed -eq 0 ]; then
        success "All smoke tests passed"
        exit 0
    else
        error "$failed smoke test(s) failed"
        exit 1
    fi
}

main "$@"
```

```bash
#!/bin/bash
# test/smoke/common.sh - Shared functions for smoke tests

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

# Check if required tools are available
check_dependencies() {
    local missing=()
    
    for tool in "$@"; do
        if ! command -v "$tool" > /dev/null 2>&1; then
            missing+=("$tool")
        fi
    done
    
    if [ ${#missing[@]} -ne 0 ]; then
        error "Missing required dependencies: ${missing[*]}"
        exit 1
    fi
}
```

## Test Execution Strategy

### Multi-Language Test Runner
```bash
#!/bin/bash
# test/scripts/test.sh - Main test runner

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Running CassandraGargoyle Multi-Language Test Suite${NC}"

# 1. Go Unit Tests
echo -e "${YELLOW}Running Go unit tests...${NC}"
go test -v ./...

# 2. Bash Smoke Tests
echo -e "${YELLOW}Running Bash smoke tests...${NC}"
for test_file in test/smoke/test_*.sh; do
    if [ -f "$test_file" ]; then
        echo "Running $test_file"
        bash "$test_file"
    fi
done

# 3. Python Integration Tests (if Python is available)
if command -v python3 > /dev/null 2>&1 && [ -f requirements-test.txt ]; then
    echo -e "${YELLOW}Running Python integration tests...${NC}"
    python3 -m pytest test/integration/ -v
    
    echo -e "${YELLOW}Running Python E2E tests...${NC}"
    python3 -m pytest test/e2e/ -v
else
    echo -e "${YELLOW}Skipping Python tests (Python not available or requirements missing)${NC}"
fi

# 4. Coverage Report
echo -e "${YELLOW}Generating Go coverage report...${NC}"
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

echo -e "${GREEN}All tests completed successfully!${NC}"
```

### Language-Specific Test Commands
```bash
# Run only Go unit tests
./test/scripts/test-unit.sh

# Run only Python integration tests
python3 -m pytest test/integration/ -v

# Run only Python E2E tests  
python3 -m pytest test/e2e/ -v

# Run only Bash smoke tests
./test/scripts/test-smoke.sh

# Run tests in parallel (Python)
python3 -m pytest test/integration/ -n auto  # Requires pytest-xdist
```

## Best Practices by Language

### Go Testing
- Keep unit tests in same package directory
- Use table-driven tests for multiple scenarios
- Mock external dependencies with interfaces
- Focus on business logic and algorithms

### Python Testing  
- Use pytest fixtures for resource management
- Parameterize tests for multiple test cases
- Use subprocess for CLI testing
- Handle container lifecycle properly
- Use pytest-mock for mocking system calls

### Bash Testing
- Keep tests simple and focused
- Use functions for reusable test logic
- Provide clear error messages
- Test CLI interfaces and basic functionality
- Exit with proper error codes

## Mocking and Test Doubles

*Mocking examples and interfaces will be completed after first test implementation.*

## Test Fixtures and Data

*Test data management examples will be completed after first test implementation.*

## Benchmarking

*Benchmarking examples will be completed after first test implementation.*

## Test Utilities

*Common test utilities will be completed after first test implementation.*

## Coverage Requirements

### Coverage Targets
- **Minimum coverage**: 80% for unit tests
- **Integration coverage**: 60% for integration tests
- **Critical packages**: 90% coverage required

### Generating Coverage Reports
```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Show coverage by function
go tool cover -func=coverage.out

# Check coverage threshold
go tool cover -func=coverage.out | grep total | grep -E \"[0-9]+\\.[0-9]+%\" | sed 's/.*\\t//' | sed 's/%//' | awk '{if ($1 < 80) exit 1}'
```

## Build Tags and Test Categories

### Using Build Tags
```go
// Unit tests (default)
//go:build !integration

// Integration tests
//go:build integration

// End-to-end tests
//go:build e2e
```

### Running Different Test Categories
```bash
# Unit tests only (default)
go test ./...

# Integration tests
go test -tags=integration ./...

# End-to-end tests
go test -tags=e2e ./...

# All tests
go test -tags=\"integration e2e\" ./...

# Short tests only
go test -short ./...
```

### Integration Tests
**IMPORTANT: ALWAYS use the launcher scripts in `test/scripts/`, never direct pytest commands!**

The launcher scripts ensure proper environment setup, Python virtual environment activation, dependency management, and container environment configuration.

#### Linux (Bash):
```bash
# General pattern for issue-specific tests
./test/scripts/issue-{number}-test-runner.sh [OPTIONS]

# PowerShell wrapper for cross-platform tests
./test/scripts/run-integration-tests.sh [OPTIONS]
```

#### Windows (PowerShell):
```powershell
# PowerShell wrapper for integration tests
.\test\scripts\run-integration-tests.ps1 [OPTIONS]
```

**Note:** PowerShell uses a single dash (-), not double dashes (--) for parameters!

**Prerequisites:** The `portunix` binary must be built and a working container engine (Podman/Docker) must be available.

**Results:** Tests generate HTML reports in `test/results/` and detailed logs.

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Test Suite

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run unit tests
        run: go test -v ./...
      
      - name: Generate coverage
        run: go test -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3

  integration-tests:
    runs-on: ubuntu-latest
    services:
      docker:
        image: docker:dind
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run integration tests
        run: go test -tags=integration -v ./...
```

## Testing Scripts

### Main Test Script
```bash
#!/bin/bash
# scripts/test.sh

set -e

# Colors
RED='\\033[0;31m'
GREEN='\\033[0;32m'
YELLOW='\\033[1;33m'
NC='\\033[0m'

echo -e \"${GREEN}Running Go Test Suite${NC}\"

# Unit tests
echo -e \"${YELLOW}Running unit tests...${NC}\"
go test -v ./...

# Integration tests
echo -e \"${YELLOW}Running integration tests...${NC}\"
go test -tags=integration -v ./...

# Coverage
echo -e \"${YELLOW}Generating coverage report...${NC}\"
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Check coverage threshold
COVERAGE=$(go tool cover -func=coverage.out | grep total | grep -E \"[0-9]+\\.[0-9]+%\" | sed 's/.*\\t//' | sed 's/%//')
THRESHOLD=80

if (( $(echo \"$COVERAGE < $THRESHOLD\" | bc -l) )); then
    echo -e \"${RED}Coverage $COVERAGE% is below threshold $THRESHOLD%${NC}\"
    exit 1
else
    echo -e \"${GREEN}Coverage $COVERAGE% meets threshold $THRESHOLD%${NC}\"
fi

echo -e \"${GREEN}All tests passed!${NC}\"
```

### Coverage Script
```bash
#!/bin/bash
# scripts/coverage.sh

set -e

echo \"Generating detailed coverage report...\"

# Generate coverage profile
go test -coverprofile=coverage.out ./...

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Show coverage by package
go tool cover -func=coverage.out

# Open HTML report (optional)
if command -v xdg-open > /dev/null; then
    xdg-open coverage.html
elif command -v open > /dev/null; then
    open coverage.html
fi

echo \"Coverage report generated: coverage.html\"
```

## Test Configuration

### go.mod for Testing
```go
// go.mod
module project.com

go 1.21

require (
    github.com/stretchr/testify v1.8.4
    github.com/testcontainers/testcontainers-go v0.24.1
)

// For test suites, also add:
// github.com/stretchr/testify/suite

require (
    // Test dependencies
    github.com/davecgh/go-spew v1.1.1 // indirect
    github.com/pmezard/go-difflib v1.0.0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

### Makefile Integration
```makefile
.PHONY: test test-unit test-integration test-coverage clean

# Run all tests
test:
\t@./scripts/test.sh

# Run unit tests only
test-unit:
\t@go test -v ./...

# Run integration tests only
test-integration:
\t@go test -tags=integration -v ./...

# Run tests with coverage
test-coverage:
\t@go test -coverprofile=coverage.out ./...
\t@go tool cover -html=coverage.out -o coverage.html

# Clean test artifacts
clean:
\t@rm -f coverage.out coverage.html
\t@go clean -testcache

# Install test dependencies
test-deps:
\t@go mod download
\t@go install github.com/golang/mock/mockgen@latest
```

## Best Practices Summary

### Test Organization
1. Use centralized `test/` directory for shared test infrastructure
2. Use descriptive test names with clear scenarios
3. Follow AAA pattern (Arrange, Act, Assert)
4. Use table-driven tests for multiple scenarios

### Test Quality
1. Tests should be fast, reliable, and independent
2. Mock external dependencies
3. Use temporary files/directories for file operations
4. Clean up resources with `t.Cleanup()`

### Coverage and Quality
1. Aim for 80%+ code coverage
2. Focus on testing critical paths and error conditions
3. Use integration tests for complex workflows
4. Benchmark performance-critical code

### Maintenance
1. Run tests in CI/CD pipeline
2. Update tests when code changes
3. Remove obsolete tests
4. Keep test dependencies up to date

## Real-World Test Examples

### Issue 12: PowerShell Installation in Linux Containers via SSH

**Scenario**: Test PowerShell installation across all supported Linux distributions. Portunix creates or connects to virtual machines, copies itself into the VM environment via SSH, and then executes installations within the isolated virtual environment. The created environments can be registered for Claude Code usage during development.

**Test Type**: Integration Test (involves multiple components: containers, SSH, package installation)

**Test Structure**:
```
test/
├── integration/
│   └── issue012_powershell_ssh_test.go
├── fixtures/
│   └── linux_distros/
│       ├── ubuntu-22.04-minimal.dockerfile
│       ├── ubuntu-24.04-minimal.dockerfile
│       ├── debian-bullseye-minimal.dockerfile
│       ├── debian-bookworm-minimal.dockerfile
│       ├── fedora-39-minimal.dockerfile
│       ├── fedora-40-minimal.dockerfile
│       ├── rocky-9-minimal.dockerfile
│       └── mint-21-minimal.dockerfile
└── testdata/
    └── test_configs.json
```

### Test Implementation

#### Test Structure Example

```go
// test/integration/issue012_powershell_ssh_test.go
package integration

// TestIssue012_PowerShellInstallation_AllDistros tests PowerShell installation
// across all supported Linux distributions using Portunix VM management
func TestIssue012_PowerShellInstallation_AllDistros(t *testing.T) {
    supportedDistros := []struct {
        name       string
        image      string
        variant    string
    }{
        {"Ubuntu 22.04", "ubuntu:22.04", "ubuntu"},
        {"Ubuntu 24.04", "ubuntu:24.04", "ubuntu"},
        {"Debian 11", "debian:bullseye", "debian"},
        {"Debian 12", "debian:bookworm", "debian"},
        {"Fedora 39", "fedora:39", "fedora"},
        {"Fedora 40", "fedora:40", "fedora"},
        {"Rocky Linux 9", "rockylinux:9", "rocky"},
        {"Linux Mint 21", "linuxmintd/mint21-amd64", "mint"},
    }

    for _, distro := range supportedDistros {
        t.Run(distro.name, func(t *testing.T) {
            t.Parallel()
            testPowerShellInstallationOnDistro(t, distro.image, distro.variant)
        })
    }
}

// Implementation details will be completed based on Portunix API integration
```

#### Test Implementation Details

*Implementation will be completed based on Portunix API integration.*

### Test Execution

**Running the test**:
```bash
# Run specific test
./test/scripts/test-integration.sh -run TestIssue012

# Run with verbose output  
go test -v ./test/integration -run TestIssue012

# Run in parallel for faster execution
go test -parallel 5 ./test/integration -run TestIssue012
```

**Expected Results**:
- All 8 Linux distributions should successfully install PowerShell
- PowerShell version verification should pass on all distros
- Test should complete within reasonable time (5-10 minutes per distro)
- Failed installations should provide clear error messages

---

**Note**: These guidelines should be adapted based on specific project requirements and complexity. Regular review ensures tests remain effective and maintainable.

*Created: 2025-08-23*
*Last updated: 2025-08-23*