# Portunix Test Suite

Multi-language testing framework for Portunix following the CassandraGargoyle testing methodology.

## Directory Structure

```
test/
├── integration/           # Python integration tests
├── e2e/                  # Python end-to-end tests
├── smoke/                # Bash smoke tests
├── scripts/              # Test runner scripts
├── conftest.py          # pytest configuration
└── README.md            # This file
```

## Quick Start

### Python Integration Tests (Issue #012)

```bash
# Install Python dependencies
pip install -r requirements-test.txt

# Run quick test (Ubuntu 22.04 only)
python test/scripts/test-integration.py --quick

# Run full test suite (all 6 distributions)
python test/scripts/test-integration.py --full-suite

# Run specific distribution
python test/scripts/test-integration.py --distribution ubuntu-22

# Run with parallel execution
python test/scripts/test-integration.py --full-suite --parallel

# List available distributions
python test/scripts/test-integration.py --list-distributions

# Clean up test containers
python test/scripts/test-integration.py --cleanup
```

### Direct pytest Usage

```bash
# Run all integration tests
pytest test/integration/ -v

# Run quick tests only
pytest test/integration/ -m quick -v

# Run tests in parallel
pytest test/integration/ -n auto

# Generate HTML report
pytest test/integration/ --html=report.html --self-contained-html
```

### Available Test Markers

- `quick` - Quick tests (single distribution)
- `slow` - Slow tests (full suite)
- `integration` - Integration tests
- `e2e` - End-to-end tests
- `podman` - Tests requiring Podman

## Test Features

### Issue #012 PowerShell Installation Tests

- **8 Linux Distributions**: Ubuntu, Debian, Fedora, Rocky, Mint
- **Container-based**: Uses Podman rootless containers
- **Detailed Logging**: Step-by-step installation logs with emojis
- **Fallback Testing**: Tests snap installation when native fails
- **HTML Reports**: Professional test reports with statistics
- **Parallel Execution**: Run multiple distributions simultaneously

### Test Workflow

1. **Container Creation**: Creates isolated test containers
2. **Dependency Installation**: Installs basic dependencies (lsb-release, curl, etc.)
3. **Binary Deployment**: Copies portunix binary to container
4. **PowerShell Installation**: Tests native package installation
5. **Fallback Testing**: Tests snap fallback if native fails
6. **Verification**: Verifies PowerShell is working
7. **System Check**: Final system state verification
8. **Cleanup**: Removes containers (configurable)

### Error Handling

- Failed containers can be kept for debugging
- Detailed error logs with timestamps
- Automatic fallback testing
- Comprehensive verification steps

## Prerequisites

- Python 3.7+
- Podman (rootless mode recommended)
- Built portunix binary
- Python test dependencies: `pip install -r requirements-test.txt`

## Configuration

### pytest.ini
Controls pytest behavior, markers, and test discovery.

### requirements-test.txt
Python dependencies for testing framework.

### conftest.py
Shared fixtures and pytest configuration.

## Integration with CI/CD

The test suite is designed to work in CI/CD environments:

```yaml
# Example GitHub Actions usage
- name: Install Python dependencies
  run: pip install -r requirements-test.txt

- name: Run integration tests
  run: python test/scripts/test-integration.py --full-suite --parallel

- name: Upload test report
  uses: actions/upload-artifact@v3
  with:
    name: test-report
    path: test/results/
```

## Methodology

This follows the CassandraGargoyle testing pyramid:
- **Go**: Unit tests (70%)
- **Python**: Integration/E2E tests (30%) 
- **Bash**: Smoke tests (quick verification)

Each language is chosen for its strengths:
- **Python**: Better for complex scenarios, container management, SSH
- **Go**: Fast unit tests, internal package testing
- **Bash**: Simple CLI verification and smoke tests