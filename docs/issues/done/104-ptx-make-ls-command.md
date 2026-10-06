# Issue #104: PTX-Make LS Command Implementation

**Type**: Enhancement
**Priority**: Medium
**Status**: ✅ Implemented
**Closed**: 2026-05-10
**Labels**: enhancement, helper-binary, ptx-make, cross-platform, file-operations
**Related ADR**: ADR-027

---

## Summary

Implement `portunix make ls` command that provides cross-platform directory listing functionality. The command first attempts to use native `ls` command, and falls back to Go-native emulation when `ls` is unavailable or fails.

## Motivation

The PTX-Make helper currently lacks directory listing capability. While `portunix make exists` can check if a path exists, there is no way to list directory contents. This is essential for:

1. **Build scripts** - Listing files for processing in Makefiles
2. **Cross-platform compatibility** - Consistent output format across Windows/Linux
3. **Container environments** - Some minimal containers may lack `ls`
4. **CI/CD pipelines** - Predictable output for scripting

## Functional Requirements

### Basic Usage

```bash
# List current directory
portunix make ls

# List specific directory
portunix make ls /path/to/dir

# List with pattern (glob)
portunix make ls *.go
portunix make ls src/**/*.go
```

### Options (Following ls conventions)

```bash
# Long format with details
portunix make ls -l

# Include hidden files
portunix make ls -a

# Human-readable sizes
portunix make ls -h

# Recursive listing
portunix make ls -R

# Sort by time (newest first)
portunix make ls -t

# Sort by size (largest first)
portunix make ls -S

# Reverse sort order
portunix make ls -r

# Combine flags
portunix make ls -lah
```

### Output Format

**Default (names only):**

```text
file1.go
file2.go
subdir/
```

**Long format (-l):**

```text
-rw-r--r--  1024  2025-12-02 14:30  file1.go
-rw-r--r--  2048  2025-12-01 10:15  file2.go
drwxr-xr-x     -  2025-11-30 08:00  subdir/
```

**Long format with human-readable sizes (-lh):**

```text
-rw-r--r--  1.0K  2025-12-02 14:30  file1.go
-rw-r--r--  2.0K  2025-12-01 10:15  file2.go
drwxr-xr-x     -  2025-11-30 08:00  subdir/
```

## Technical Requirements

### Detection Strategy

1. **Native ls detection:**
   - Check for `ls` in PATH (works on Linux, macOS, Windows with Git Bash/WSL/MSYS2)
   - Use `exec.LookPath("ls")` for detection

2. **Execution flow:**
   ```go
   func lsCommand(args []string) error {
       if canUseNativeLS() {
           return executeNativeLS(args)
       }
       return emulateLS(args)
   }
   ```

3. **Native execution:**
   - Pass arguments directly to `ls`
   - Capture and return output unchanged

4. **Go emulation (fallback):**
   - Use `os.ReadDir()` for directory listing
   - Use `os.Stat()` for file metadata
   - Implement sorting and filtering in Go
   - Produce ls-compatible output format

### Cross-Platform Considerations

| Feature | Native ls | Go Emulation |
| ------- | --------- | ------------ |
| Basic list | ✅ Native | ✅ |
| Long format (-l) | ✅ Native | ✅ |
| Hidden files (-a) | ✅ Native | ✅ |
| Recursive (-R) | ✅ Native | ✅ |
| Glob patterns | ✅ Native | ✅ |
| Permissions | ✅ Native | ⚠️ Simplified on Windows |

### Performance

- Native command preferred for large directories (> 1000 files)
- Go emulation acceptable for small/medium directories
- Lazy loading for recursive listing

## Implementation Phases

### Phase 1: Basic Emulation

- [ ] Implement basic `ls` with Go-native `os.ReadDir()`
- [ ] Support path argument
- [ ] Default output format (names only)
- [ ] Directory indicator (`/` suffix)

### Phase 2: Options Support

- [ ] Implement `-l` (long format)
- [ ] Implement `-a` (hidden files)
- [ ] Implement `-h` (human-readable sizes)
- [ ] Implement combined flags (`-lah`)

### Phase 3: Native Detection

- [ ] Detect native `ls` availability via `exec.LookPath("ls")`
- [ ] Implement native execution path
- [ ] Pass-through arguments to native `ls`

### Phase 4: Advanced Features

- [ ] Implement `-R` (recursive)
- [ ] Implement `-t` (sort by time)
- [ ] Implement `-S` (sort by size)
- [ ] Implement `-r` (reverse sort)
- [ ] Glob pattern support

### Phase 5: Testing & Documentation

- [ ] Unit tests for Go emulation
- [ ] Integration tests with native commands
- [ ] Cross-platform testing (Linux, Windows)
- [ ] Documentation and examples

## Acceptance Criteria

- [ ] `portunix make ls` lists current directory
- [ ] `portunix make ls /path` lists specified directory
- [ ] `-l`, `-a`, `-h` flags work as expected
- [ ] Works on Linux without native `ls` (container scenarios)
- [ ] Works on Windows with consistent output format
- [ ] Exit code 0 on success, non-zero on error (path not found, permission denied)

## Examples

### Makefile Integration

```makefile
# List all Go files
list-sources:
	portunix make ls *.go

# Check build artifacts
list-dist:
	portunix make ls -lh dist/

# Find all test files recursively
list-tests:
	portunix make ls -R *_test.go
```

### Script Usage

```bash
# Check if directory has content
if portunix make ls dist/ > /dev/null 2>&1; then
    echo "dist/ has files"
fi

# Get file count
FILE_COUNT=$(portunix make ls src/*.go | wc -l)
```

## Related Issues

- Issue #103: PTX-Make Helper Implementation (parent)
- Issue #102: PTX-Make RM Command (similar pattern)

## Related ADRs

- ADR-027: PTX-Make Helper Architecture
- ADR-014: Git-like Dispatcher Pattern

---

## Revision History

| Date | Change | Author |
| ---- | ------ | ------ |
| 2025-12-02 | Initial creation | Kurc |
