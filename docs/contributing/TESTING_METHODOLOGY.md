# Testing Methodology - Portunix

## Overview
This methodology defines standard procedures for testing in the Portunix project with emphasis on detailed logging and debugging.

## Test Framework

### Important Notes
- **TestFramework Package**: Tests using `testframework.NewTestFramework()` use the `testframework` package
- **Simple Usage**: `go test ./test/integration/test_file.go -v` (single file execution)
- **Package Import**: Tests must import `"portunix.ai/portunix/test/testframework"`

### Verbose Logging System
We have implemented `TestFramework` which provides:
- **Detailed logging** with emoji indicators
- **Verbose mode** activated by parameter
- **Structured steps** with numbering
- **Timing information** and summary
- **Color differentiation** for success/errors

### Verbose Mode Activation

#### Method 1: Single Test File (Recommended)
```bash
# Run single test with verbose output
go test ./test/integration/issue_037_mcp_serve_test.go -v

# Run specific test case
go test ./test/integration/issue_037_mcp_serve_test.go -v -run TC001
```

#### Method 2: All Integration Tests
```bash
# Run all tests in integration package
go test ./test/integration/ -v
```

#### Method 3: Specific Test Pattern
```bash
# Run tests matching pattern
go test ./test/integration/ -v -run "MCP"
```

#### Method 4: With Timeout for E2E Tests
```bash
# For longer E2E tests (container-based)
go test ./test/integration/claude_code_container_install_test.go -v -timeout 10m
```

## Container-Based Testing

### Portunix Container Integration
**MANDATORY**: All software installation testing MUST use Portunix native container commands instead of direct Docker/Podman calls.

#### Rationale
- **Universal compatibility**: Portunix automatically selects Docker or Podman based on availability
- **Integrated functionality**: Uses Portunix's container management system
- **Consistent environment**: Standardized container setup across all tests
- **Simplified commands**: No need to handle Docker vs Podman logic manually

#### Container Commands for Testing

##### Method 1: Portunix Container Run (Recommended)
```bash
# Use Portunix native container management
./portunix docker run-in-container default --image ubuntu:22.04
./portunix podman run alpine:latest
```

##### Method 2: Container Installation Testing
```go
// In Go tests - use Portunix container commands
tf.Command(t, binaryPath, []string{"docker", "run-in-container", "nodejs", "--image", "ubuntu:22.04"})
tf.Command(t, binaryPath, []string{"podman", "run", "alpine:latest"})
```

#### Test Implementation Pattern
```go
// ✅ CORRECT: Use Portunix container commands
func runContainerTest(t *testing.T, tf *testframework.TestFramework, binaryPath string) bool {
    // Create container using Portunix
    tf.Command(t, binaryPath, []string{"docker", "run-in-container", "nodejs", "--image", "ubuntu:22.04"})
    cmd := exec.Command(binaryPath, "docker", "run-in-container", "nodejs", "--image", "ubuntu:22.04")
    output, err := cmd.CombinedOutput()
    
    if err != nil {
        tf.Error(t, "Container test failed", err.Error())
        return false
    }
    
    tf.Success(t, "Container test completed")
    return true
}

// ❌ INCORRECT: Direct Docker/Podman calls
// Never use: exec.Command("docker", "run", ...)
// Never use: exec.Command("podman", "run", ...)
```

#### Container Testing Workflow

1. **Container Selection**
   - Let Portunix choose Docker or Podman automatically
   - Use standard images: ubuntu:22.04, debian:bookworm, alpine:latest
   - Specify container requirements via Portunix flags

2. **Installation Testing**
   ```go
   // Test package installation in container
   tf.Command(t, binaryPath, []string{"docker", "run-in-container", "python", "--image", "ubuntu:22.04"})
   tf.Command(t, binaryPath, []string{"docker", "run-in-container", "nodejs", "--image", "debian:bookworm"})
   ```

3. **Verification Testing**
   ```go
   // Verify installation worked inside container
   // Container automatically includes SSH access for verification
   ```

4. **Multi-Platform Testing**
   ```go
   // Test across different base images
   platforms := []string{"ubuntu:22.04", "debian:bookworm", "alpine:latest"}
   for _, platform := range platforms {
       tf.Step(t, fmt.Sprintf("Test on %s", platform))
       tf.Command(t, binaryPath, []string{"docker", "run-in-container", "nodejs", "--image", platform})
       // ... test logic
   }
   ```

### Output with Verbose Mode

**Without verbose:**
```
=== RUN   TestIssue037WithFramework
--- PASS: TestIssue037WithFramework (2.34s)
PASS
```

**With verbose mode:**
```
================================================================================
🚀 STARTING: Issue037_MCP_Serve
Description: Test MCP serve command implementation with detailed logging
Time: 2025-01-12T15:30:45Z
================================================================================

📋 STEP 1: Setup test binary
   ℹ️  Binary path: ../../portunix
   ✅ Binary found
      Size: 15234567 bytes
      Modified: 2025-01-12 15:28:33

────────────────────────────────────────────────────────────

📋 STEP 2: Test basic help display
   🔧 Executing: ../../portunix
   📄 Output (1234 chars):
      Portunix is a command-line interface (CLI) tool designed to simplify...
   ✅ Help content found
   ✅ No MCP server started (correct behavior)

────────────────────────────────────────────────────────────

📋 STEP 3: Test MCP command structure
   🔧 Executing: ../../portunix mcp --help
   📄 Output (567 chars):
      Manage MCP server integration with AI assistants...
   ✅ serve command found in MCP help

================================================================================
🎉 COMPLETED: Issue037_MCP_Serve
Duration: 2.345s
Steps: 5
================================================================================
```

## Framework API

### Basic Usage
```go
package integration

import (
    "testing"
    "portunix.ai/portunix/test/testframework"
)

func TestExample(t *testing.T) {
    // Initialize framework with package import
    tf := testframework.NewTestFramework("TestName")
    tf.Start(t, "Test description")
    
    success := true
    defer func() {
        tf.Finish(t, success)
    }()
    
    // Test steps
    tf.Step(t, "Step description")
    tf.Info(t, "Information message")
    tf.Success(t, "Success message")
    tf.Error(t, "Error message") // Sets success = false
}
```

### Available Methods

#### Structural
- `testframework.NewTestFramework(name string)` - Create new framework
- `Start(t, description)` - Start test with header
- `Finish(t, success)` - End test with summary
- `Step(t, description, details...)` - New test step

#### Logging
- `Success(t, message, details...)` - ✅ Success
- `Error(t, message, details...)` - ❌ Error
- `Warning(t, message, details...)` - ⚠️ Warning
- `Info(t, message, details...)` - ℹ️ Information

#### Special
- `Command(t, cmd, args)` - 🔧 Command logging
- `Output(t, output, maxLength)` - 📄 Command output
- `Separator()` - Visual separator
- `IsVerbose()` - Check verbose mode

## Recommended Practices

### 1. Structured Test
```go
package integration

import (
    "testing"
    "portunix.ai/portunix/test/testframework"
)

func TestFeature(t *testing.T) {
    tf := testframework.NewTestFramework("FeatureName")
    tf.Start(t, "Clear description of what we're testing")
    
    success := true
    defer tf.Finish(t, success)
    
    // Step 1: Setup
    tf.Step(t, "Setup environment")
    // ... setup logic
    
    tf.Separator()
    
    // Step 2: Main test
    tf.Step(t, "Test main functionality")
    // ... test logic
    
    if err != nil {
        tf.Error(t, "Test failed", err.Error())
        success = false
        return
    }
    tf.Success(t, "Test passed")
}
```

### 2. Command Testing Pattern
```go
tf.Step(t, "Execute command")
tf.Command(t, binaryPath, []string{"arg1", "arg2"})

cmd := exec.Command(binaryPath, "arg1", "arg2")
output, err := cmd.CombinedOutput()

tf.Output(t, string(output), 500) // Show max 500 chars

if err != nil {
    tf.Error(t, "Command failed", err.Error())
    success = false
    return
}
tf.Success(t, "Command successful")
```

### 3. Assertion Pattern
```go
if expected != actual {
    tf.Error(t, "Assertion failed",
        fmt.Sprintf("Expected: %v", expected),
        fmt.Sprintf("Actual: %v", actual))
    success = false
} else {
    tf.Success(t, "Assertion passed")
}
```

## Existing Tests

### Upgrading Existing Tests
1. **Keep original logic** - framework is additive
2. **Add framework wrapper** around existing code
3. **Gradually extend** with more detailed steps
4. **Test with and without verbose** mode

### Upgrade Examples

**Before:**
```go
func TestBasic(t *testing.T) {
    cmd := exec.Command("./portunix", "--help")
    output, err := cmd.Output()
    if err != nil {
        t.Fatalf("Failed: %v", err)
    }
    if !strings.Contains(string(output), "Usage:") {
        t.Error("Missing usage")
    }
}
```

**After:**
```go
package integration

import (
    "testing"
    "os/exec"
    "strings"
    "portunix.ai/portunix/test/testframework"
)

func TestBasic(t *testing.T) {
    tf := testframework.NewTestFramework("BasicHelp")
    tf.Start(t, "Test basic help functionality")
    
    success := true
    defer tf.Finish(t, success)
    
    tf.Step(t, "Execute help command")
    tf.Command(t, "./portunix", []string{"--help"})
    
    cmd := exec.Command("./portunix", "--help")
    output, err := cmd.Output()
    
    if err != nil {
        tf.Error(t, "Command failed", err.Error())
        success = false
        return
    }
    
    tf.Output(t, string(output), 300)
    
    if !strings.Contains(string(output), "Usage:") {
        tf.Error(t, "Missing usage text")
        success = false
    } else {
        tf.Success(t, "Usage text found")
    }
}
```

## Troubleshooting

### Common Errors

#### "undefined: testframework"
```bash
# ERROR - Missing import
package integration
// Missing: import "portunix.ai/portunix/test/testframework"
tf := NewTestFramework("TestName") // ERROR

# FIX - Add correct import
package integration
import "portunix.ai/portunix/test/testframework"
tf := testframework.NewTestFramework("TestName") // CORRECT
```

#### Module path issues
```bash
# ERROR - Wrong module path
import "portunix.ai/test/testframework"

# FIX - Correct module path  
import "portunix.ai/portunix/test/testframework"
```

#### Verbose output not showing
```bash
# Verbose output is controlled by Go's -v flag
go test ./test/integration/test_file.go -v

# TestFramework automatically detects verbose mode via testing.Verbose()
# No environment variables needed
```

## Debug Specific Problems

### For "invisible" tests
```bash
# If test seems to "not run"
go test ./path/to/test.go -v -timeout=30s

# For very detailed output with race detection
go test ./path/to/test.go -v -race -timeout=60s
```

### For Performance Problems
```bash
# With profiling
go test ./path/to/test.go -v -cpuprofile=cpu.prof -memprofile=mem.prof
```

### Testing New testframework Package
```bash
# Test single file with testframework
go test ./test/integration/issue_037_mcp_serve_test.go -v

# Test all integration tests
go test ./test/integration/ -v

# Test with timeout for E2E tests
go test ./test/integration/claude_code_container_install_test.go -v -timeout 10m
```

## CI/CD Integration

### GitHub Actions
```yaml
- name: Run integration tests with TestFramework
  run: |
    go test ./test/integration/... -v -timeout=10m
```

### Local Development
```bash
# Aliases for convenience
alias gotest-verbose='go test -v'
alias gotest-debug='go test -v -timeout=60s'
alias gotest-e2e='go test -v -timeout=10m'

# Usage examples
gotest-verbose ./test/integration/issue_037_mcp_serve_test.go
gotest-e2e ./test/integration/claude_code_container_install_test.go
```

## TestFramework Package

### Package Structure
```
test/
├── testframework/
│   └── framework.go          # Main TestFramework implementation
└── integration/
    ├── issue_037_mcp_serve_test.go    # Example usage
    └── claude_code_container_install_test.go  # E2E test example
```

### Key Features
- **Single File Execution**: No need to specify multiple files
- **Automatic Verbose Detection**: Uses Go's `testing.Verbose()`
- **Clean Package Structure**: Import once, use everywhere
- **Emoji Logging**: Visual indicators for test progress
- **Structured Steps**: Numbered steps with timing
- **Error Handling**: Proper success/failure tracking

### Migration from Old Framework
```go
// OLD (required two files)
// go test ./test.go ./test_framework.go
tf := NewTestFramework("TestName")

// NEW (single file execution)  
// go test ./test.go -v
import "portunix.ai/portunix/test/testframework"
tf := testframework.NewTestFramework("TestName")
```

### Usage Examples

#### Simple Test
```go
func TestSimple(t *testing.T) {
    tf := testframework.NewTestFramework("SimpleTest")
    tf.Start(t, "Basic functionality test")
    
    success := true
    defer tf.Finish(t, success)
    
    tf.Step(t, "Execute command")
    // ... test logic
    tf.Success(t, "Test completed")
}
```

#### E2E Container Test
```go
func TestE2E(t *testing.T) {
    tf := testframework.NewTestFramework("E2E_Container_Test")
    tf.Start(t, "End-to-end container testing with Claude Code setup")
    
    success := true
    defer tf.Finish(t, success)
    
    // Multiple test phases
    tf.Step(t, "Setup container environment")
    // ... setup logic
    tf.Separator()
    
    tf.Step(t, "Install Portunix")
    // ... installation logic
    tf.Separator()
    
    tf.Step(t, "Test MCP integration")  
    // ... integration testing
    
    tf.Success(t, "E2E test completed successfully")
}
```

---

**Created:** 2025-01-12  
**Updated:** 2025-09-12 (testframework package)  
**Version:** 2.0  
**Author:** Claude Code Assistant