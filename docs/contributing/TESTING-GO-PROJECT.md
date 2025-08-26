# Testing Guidelines for Portunix Project

## Critical Testing Principles

### 🚨 GOLDEN RULE: Tests Must Test Portunix, Not Replicate It

**NEVER** write tests that implement the functionality they are supposed to test. Tests should verify that Portunix works correctly, not do the work for Portunix.

#### ❌ WRONG: Test Installing Dependencies Itself
```python
# BAD TEST - This test installs dependencies itself instead of letting Portunix do it
def test_powershell_installation():
    container = create_container()
    
    # ❌ WRONG: Test is installing dependencies
    run_command("apt-get install -y wget sudo lsb-release")
    
    # Then runs portunix
    run_command("portunix install powershell")
    
    # This test doesn't actually test if Portunix can handle a clean system!
```

#### ✅ CORRECT: Test Lets Portunix Handle Everything
```python
# GOOD TEST - This test gives Portunix a clean environment
def test_powershell_installation():
    container = create_clean_container()  # Just a base OS, nothing else
    
    # ✅ CORRECT: Give portunix full responsibility
    copy_portunix_binary(container)
    run_command("portunix install powershell")
    
    # Verify the result
    assert powershell_is_installed()
    
    # This test actually verifies that Portunix can handle everything!
```

## Integration Testing Architecture

### Container-Based Testing

Integration tests for Portunix use containers (Podman/Docker) to ensure:
1. **Clean environment** - Each test starts with a pristine OS
2. **Isolation** - Tests don't affect the host system
3. **Reproducibility** - Same test conditions every time
4. **Multi-distribution support** - Test across Ubuntu, Debian, Fedora, etc.

### Test Structure

```
test/
├── integration/           # Integration tests
│   └── issue_XXX_*.py    # Test files for specific issues
├── unit/                  # Unit tests (Go tests)
├── fixtures/              # Test data and configurations
├── results/               # Test outputs and logs
└── scripts/              # Test runner scripts
```

## Writing Integration Tests

### 1. Container Setup

Containers should be created in a **minimal state**:
- Base OS image only
- NO pre-installed tools
- NO package manager updates
- Let Portunix handle ALL dependencies

```python
def create_container(name: str, image: str) -> str:
    """Create a CLEAN container - no modifications!"""
    result = subprocess.run([
        "podman", "run", "-d",
        "--name", container_name,
        "--network", "host",  # Use host network for DNS
        image,
        "/bin/sh", "-c", "tail -f /dev/null"
    ])
    # DO NOT install anything here!
    return container_name
```

### 2. Testing Portunix Functionality

The test should:
1. Copy the portunix binary to the container
2. Run portunix commands
3. Verify the results

```python
def test_package_installation(container_name: str, package: str):
    # Step 1: Copy portunix
    copy_portunix_binary(container_name)
    
    # Step 2: Run portunix (it handles everything)
    result = run_command(f"portunix install {package}")
    
    # Step 3: Verify installation
    assert verify_package_installed(package)
```

### 3. Logging and Debugging

Tests should capture detailed logs for debugging:

```python
def install_package(container_name: str, package: str, log_file):
    log_file.write(f"Installing {package} via portunix\n")
    
    result = subprocess.run([
        "podman", "exec", container_name,
        "/usr/local/bin/portunix", "install", package
    ], capture_output=True, text=True)
    
    # Log both stdout and stderr
    log_file.write(result.stdout)
    if result.stderr:
        log_file.write(result.stderr)
    
    return result.returncode == 0
```

## Common Testing Mistakes to Avoid

### 1. Pre-installing Dependencies
**NEVER** install tools that Portunix should install itself:
- ❌ Don't install wget, curl, sudo, etc. in test setup
- ❌ Don't run apt-get update before testing
- ✅ Let Portunix detect and install what it needs

### 2. Assuming Environment State
**NEVER** assume tools are available:
- ❌ Don't assume sudo exists
- ❌ Don't assume package managers are configured
- ✅ Test that Portunix handles missing tools gracefully

### 3. Incomplete Verification
**ALWAYS** verify the actual result:
- ❌ Don't just check return codes
- ✅ Verify the software actually works (e.g., `pwsh --version`)
- ✅ Check that files are in expected locations

### 4. Network Issues in Containers
**ALWAYS** ensure containers have network access:
- Use `--network host` for Podman/Docker
- Consider IPv4 vs IPv6 issues
- Test DNS resolution works

## Running Tests

### Quick Test (Single Distribution)
```bash
cd test
source venv/bin/activate
python -m pytest integration/issue_012_powershell_installation_test.py::TestPowerShellInstallation::test_quick_ubuntu_22_installation -v
```

### Full Test Suite
```bash
cd test
source venv/bin/activate
python -m pytest integration/ -v
```

### With Detailed Output
```bash
python -m pytest integration/test_file.py -v -s --tb=short
```

## Test Environment Requirements

### Python Virtual Environment
```bash
cd test
python3 -m venv venv
source venv/bin/activate
pip install -r requirements-test.txt
```

### Container Runtime
- Podman (preferred) or Docker
- Must support rootless containers
- Network access for package downloads

## Debugging Failed Tests

### 1. Check Test Logs
Test logs are saved in `test/results/logs-TIMESTAMP/`:
```bash
cat test/results/logs-*/ubuntu-22.log
```

### 2. Keep Failed Containers
Modify test to keep containers for debugging:
```python
def teardown_method(self):
    if self._test_failed:
        print(f"Container {self.container_name} kept for debugging")
        return  # Don't remove container
    self._remove_container(self.container_name)
```

### 3. Manual Container Inspection
```bash
# List test containers
podman ps -a | grep portunix-test

# Enter container for debugging
podman exec -it portunix-test-ps-ubuntu-22 /bin/bash

# Check portunix logs
podman logs portunix-test-ps-ubuntu-22
```

## Test Coverage Guidelines

### What to Test
1. **Clean installations** - Software installs on fresh OS
2. **Dependency handling** - Portunix installs missing tools
3. **Error handling** - Graceful failures with helpful messages
4. **Multi-distribution support** - Works across different Linux distros
5. **Version compatibility** - Handles different OS versions

### Test Matrix
Each package should be tested on:
- Ubuntu 22.04, 24.04
- Debian 11, 12
- Fedora 39, 40
- Rocky Linux 9
- Linux Mint 21

## Writing Unit Tests (Go)

### Test File Naming
- Test files end with `_test.go`
- Integration tests end with `_integration_test.go`

### Example Unit Test
```go
func TestInstallPackage(t *testing.T) {
    // Test that InstallPackage function works correctly
    err := InstallPackage("wget", "")
    if err != nil {
        t.Errorf("InstallPackage failed: %v", err)
    }
}
```

### Running Go Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./app/install/...

# Verbose output
go test -v ./...
```

## CI/CD Integration

Tests should be automated in CI/CD pipeline:
1. Build portunix binary
2. Run unit tests
3. Run integration tests in containers
4. Generate test reports
5. Upload logs as artifacts

## Summary Checklist

Before committing a test, verify:
- [ ] Test uses clean containers (no pre-installed dependencies)
- [ ] Test lets Portunix handle all installations
- [ ] Test verifies actual functionality (not just return codes)
- [ ] Test captures detailed logs for debugging
- [ ] Test handles network issues gracefully
- [ ] Test works across multiple distributions
- [ ] Test doesn't duplicate Portunix logic

## Remember

**The purpose of testing is to verify that Portunix works correctly, not to implement its functionality in the test code. If your test is installing dependencies that Portunix should install, you're testing your test code, not Portunix!**