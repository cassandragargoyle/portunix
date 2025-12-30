# TestFramework

Standardized testing framework for Portunix with verbose logging and structured test execution.

## Installation

The package is part of the Portunix project. Import it in your test files:

```go
import "portunix.ai/portunix/test/testframework"
```

## Quick Start

```go
func TestExample(t *testing.T) {
    tf := testframework.NewTestFramework("ExampleTest")
    tf.Start(t, "Test description")

    success := true
    defer tf.Finish(t, success)

    tf.Step(t, "Step description")
    tf.Success(t, "Step completed")
}
```

## Running Tests

```bash
# Standard run (minimal output)
go test ./test/integration/your_test.go

# Verbose mode (detailed logging with emojis)
go test ./test/integration/your_test.go -v

# With timeout for E2E tests
go test ./test/integration/your_test.go -v -timeout 10m
```

## API Reference

### Constructor

#### `NewTestFramework(testName string) *TestFramework`
Creates a new test framework instance. Automatically detects verbose mode via `testing.Verbose()`.

### Lifecycle Methods

#### `Start(t *testing.T, description string)`
Begins a test with header output in verbose mode.

#### `Finish(t *testing.T, success bool)`
Completes the test with summary (duration, step count, pass/fail status).

### Logging Methods

| Method | Emoji | Description |
|--------|-------|-------------|
| `Step(t, description, details...)` | 📋 | Numbered test step |
| `Success(t, message, details...)` | ✅ | Success message |
| `Error(t, message, details...)` | ❌ | Error (calls `t.Errorf`) |
| `Warning(t, message, details...)` | ⚠️ | Warning message |
| `Info(t, message, details...)` | ℹ️ | Informational message |

### Command Execution

#### `Command(t *testing.T, command string, args []string)`
Logs command execution with 🔧 emoji.

#### `Output(t *testing.T, output string, maxLength int)`
Logs command output with truncation for long outputs.

### Utility Methods

#### `Separator()`
Prints visual separator line in verbose mode.

#### `IsVerbose() bool`
Returns whether verbose mode is enabled.

### Binary Verification

#### `VerifyBinary(t *testing.T, relativePath string) (string, bool)`
Verifies binary exists at specified path.

#### `VerifyPortunixBinary(t *testing.T) (string, bool)`
Searches for Portunix binary in standard locations (`../../portunix`, `./portunix`, `../portunix`).

#### `MustVerifyPortunixBinary(t *testing.T) string`
Same as `VerifyPortunixBinary` but calls `t.FailNow()` if not found.

## Example: Complete Test

```go
package integration

import (
    "os/exec"
    "strings"
    "testing"

    "portunix.ai/portunix/test/testframework"
)

func TestMCPServe(t *testing.T) {
    tf := testframework.NewTestFramework("MCP_Serve")
    tf.Start(t, "Test MCP serve command")

    success := true
    defer tf.Finish(t, success)

    // Step 1: Verify binary
    binaryPath := tf.MustVerifyPortunixBinary(t)
    tf.Separator()

    // Step 2: Execute command
    tf.Step(t, "Execute MCP help command")
    tf.Command(t, binaryPath, []string{"mcp", "--help"})

    cmd := exec.Command(binaryPath, "mcp", "--help")
    output, err := cmd.CombinedOutput()

    if err != nil {
        tf.Error(t, "Command failed", err.Error())
        success = false
        return
    }

    tf.Output(t, string(output), 500)

    // Step 3: Verify output
    tf.Step(t, "Verify output contains expected text")
    if strings.Contains(string(output), "serve") {
        tf.Success(t, "Found 'serve' command in help")
    } else {
        tf.Error(t, "Missing 'serve' command in help")
        success = false
    }
}
```

## Output Examples

### Standard Mode
```
=== RUN   TestMCPServe
--- PASS: TestMCPServe (0.15s)
```

### Verbose Mode (`-v`)
```
================================================================================
🚀 STARTING: MCP_Serve
Description: Test MCP serve command
Time: 2025-01-12T15:30:45Z
================================================================================

📋 STEP 1: Verify Portunix binary exists
   ✅ Portunix binary found
      Size: 15234567 bytes
      Modified: 2025-01-12 15:28:33

────────────────────────────────────────────────────────────

📋 STEP 2: Execute MCP help command
   🔧 Executing: /path/to/portunix mcp --help
   📄 Output (567 chars):
      Manage MCP server integration...

📋 STEP 3: Verify output contains expected text
   ✅ Found 'serve' command in help

--------------------------------------------------------------------------------
🎉 COMPLETED: MCP_Serve
Duration: 150.234ms
Steps: 3
--------------------------------------------------------------------------------
```

## Best Practices

1. **Always use defer for Finish**: Ensures test summary is printed even on failures
2. **Use MustVerifyPortunixBinary**: Fails fast if binary is missing
3. **Add Separators between sections**: Improves readability in verbose mode
4. **Keep step descriptions concise**: Single line describing what's being tested
5. **Use appropriate log levels**: Success for positive outcomes, Error for failures
