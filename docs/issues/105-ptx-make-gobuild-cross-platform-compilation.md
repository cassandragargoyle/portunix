# Issue #105: PTX-Make GoBuild Cross-Platform Compilation

**Type**: Enhancement
**Priority**: High
**Status**: ✅ Implemented
**Labels**: enhancement, helper-binary, ptx-make, cross-platform, go-compilation, build-automation
**Related Issue**: #103 (PTX-Make Helper Implementation)

---

## Summary

Add `portunix make gobuild` command for cross-platform Go compilation that works on all platforms including Windows, where Unix-style inline environment variables don't work. The command uses **transparent Unix-style syntax** - same as original command.

## Problem Statement

On Windows, Unix-style inline environment variable syntax doesn't work:

```bash
# Works on Unix only:
GOOS=linux GOARCH=amd64 go build ...

# Windows requires different approach:
# - PowerShell: $env:GOOS="linux"; $env:GOARCH="amd64"; go build ...
# - CMD: set GOOS=linux && set GOARCH=amd64 && go build ...
```

This creates portability issues in Makefiles and build scripts that need to compile Go code for multiple platforms.

## Proposed Solution

New command `portunix make gobuild` with **transparent Unix-style syntax**:

```bash
portunix make gobuild [VAR=value]... go build [go-build-args]
```

The command parses `VAR=value` pairs at the beginning, sets them as environment variables, and executes the remaining command.

### Syntax

```bash
portunix make gobuild GOOS=linux GOARCH=amd64 go build -ldflags "..." -o output package
                      ^^^^^^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                      Environment variables    Actual go command (passed through)
```

### Usage Examples

```bash
# Cross-compile for Linux from Windows - SAME syntax as Unix!
portunix make gobuild GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=1.0.0" -o ../dist/linux-amd64/myapp .

# For macOS ARM64
portunix make gobuild GOOS=darwin GOARCH=arm64 go build -o ../dist/darwin-arm64/myapp .

# With CGO disabled
portunix make gobuild CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/myapp .

# Native build (no cross-compilation, just wrapper)
portunix make gobuild go build -o ./bin/myapp .
```

## Implementation Design

### Parsing Algorithm

1. Iterate through arguments from start
2. While argument matches pattern `KEY=value` (contains `=` and KEY is uppercase):
   - Extract KEY and value
   - Store in environment map
3. Remaining arguments are the actual command
4. Set environment variables
5. Execute command

### Pseudocode

```go
func gobuildCmd(args []string) error {
    env := make(map[string]string)
    cmdStart := 0

    // Parse VAR=value pairs
    for i, arg := range args {
        if isEnvVar(arg) { // matches ^[A-Z_][A-Z0-9_]*=
            parts := strings.SplitN(arg, "=", 2)
            env[parts[0]] = parts[1]
            cmdStart = i + 1
        } else {
            break
        }
    }

    if cmdStart >= len(args) {
        return fmt.Errorf("no command specified")
    }

    // Build command
    cmd := exec.Command(args[cmdStart], args[cmdStart+1:]...)
    cmd.Env = os.Environ()
    for k, v := range env {
        cmd.Env = append(cmd.Env, k+"="+v)
    }
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func isEnvVar(s string) bool {
    return regexp.MustCompile(`^[A-Z_][A-Z0-9_]*=`).MatchString(s)
}
```

## Makefile Integration

After implementation, Makefile lines need **minimal change**:

```makefile
# Before (platform-dependent, Unix only):
cd tagent && GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../$(DIST_LINUX)/tagent .

# After (cross-platform, same syntax!):
cd tagent && portunix make gobuild GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../$(DIST_LINUX)/tagent .
```

**Benefit**: Only prefix `portunix make gobuild ` - rest of command stays identical!

## Technical Requirements

### Platform Support
- **Windows**: Full support (primary motivation)
- **Linux**: Full support
- **macOS**: Full support

### Dependencies
- Go toolchain must be installed and accessible in PATH
- Uses existing ptx-make helper binary structure

### Architecture
- Extend existing `ptx-make` helper (Issue #103)
- Follow ADR-027 patterns

### Error Handling
- Clear error if no command specified after env vars
- Clear error if `go` command not found
- Forward all `go build` errors transparently
- Non-zero exit code on failure

## Acceptance Criteria

- [x] Command `portunix make gobuild` implemented
- [x] Works on Windows, Linux, and macOS
- [x] Parses `VAR=value` pairs correctly
- [x] Passes remaining args to executed command
- [x] Supports any environment variable (GOOS, GOARCH, CGO_ENABLED, etc.)
- [x] Error handling for missing command
- [x] Error handling for missing Go toolchain
- [x] Unit tests for env var parsing
- [ ] Integration test for actual compilation
- [x] Documentation with examples

## Out of Scope

- CGO cross-compilation (requires C toolchain) - but CGO_ENABLED=0 works
- Docker-based builds
- Automatic binary naming conventions

## Testing Plan

### Unit Tests
- Env var parsing (`GOOS=linux` → key=GOOS, value=linux)
- Multiple env vars parsing
- Command extraction after env vars
- Edge cases (value with `=` sign, empty value)

### Integration Tests
```bash
# Test in container
portunix container run ubuntu
cd /tmp
go mod init testmod
echo 'package main; func main() {}' > main.go

# Test cross-compilation
portunix make gobuild GOOS=linux GOARCH=amd64 go build -o ./test-binary .
./test-binary  # Should work

# Test with ldflags
portunix make gobuild GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=test" -o ./test2 .
```

### Cross-Platform Verification
- Windows → Linux compilation
- Linux → Windows compilation
- Linux → macOS compilation

## Related Issues

- #103: PTX-Make Helper Implementation
- #104: PTX-Make LS Command

## Related ADRs

- ADR-027: PTX-Make Helper Architecture

---

## Revision History

| Date | Change | Author |
|------|--------|--------|
| 2025-12-03 | Initial creation based on feature request | Kurc |
| 2025-12-03 | Changed to transparent Unix-style syntax per user request | Kurc |
| 2025-12-03 | Implemented and merged to main | Kurc |
