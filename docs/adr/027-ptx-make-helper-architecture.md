# ADR-027: PTX-Make Helper Architecture

**Status**: Proposed
**Date**: 2025-12-02
**Architect**: Kurc
**Related Issues**: TBD (Issue to be created)
**Related ADRs**: ADR-014 (Git-like Dispatcher), ADR-026 (Shared Platform Utilities)

---

## Context

### Problem Statement

Makefile-based build systems suffer from cross-platform incompatibility issues. Common operations that work seamlessly on Unix-like systems fail or require complex workarounds on Windows:

```makefile
# Unix-only commands that fail on Windows
mkdir -p dist/bin
cp -r src/* dist/
rm -rf build/
chmod +x script.sh
git describe --tags --always --dirty
sha256sum dist/* > checksums.sha256
```

Current workarounds include:
- Conditional shell detection (`ifeq ($(OS),Windows_NT)`)
- Separate `.bat` and `.sh` scripts
- CMake or other meta-build systems
- GNU Make extensions (not available on Windows by default)

### Existing Platform Detection

Portunix already has robust platform detection through `src/pkg/platform/` (ADR-026):
- `GetOS()` - Returns "windows", "linux", "darwin"
- `GetArchitecture()` - Returns "x64", "x86", "arm64"
- `GetPlatform()` - Returns "linux-x64", "windows-x64", etc.
- `IsWindows()`, `IsLinux()`, `IsDarwin()` - Boolean helpers

This infrastructure can be leveraged for cross-platform Makefile utilities.

### Target Use Cases

1. **Portunix Internal**: Simplify `Makefile` for cross-platform builds
2. **External Projects**: Provide portable build utilities for any Makefile-based project
3. **CI/CD Pipelines**: Consistent behavior across Linux and Windows runners
4. **Developer Experience**: Same commands work on all platforms without modification

---

## Decision

Create a new helper binary `ptx-make` that provides cross-platform Makefile utility functions, following the established Portunix helper pattern (ADR-014).

### Command Structure

```
ptx-make <command> [arguments...]
```

Alternatively, via dispatcher:
```
portunix make <command> [arguments...]
```

---

## Functional Specification

### File Operations

#### `copy` - Cross-platform File Copy
```bash
ptx-make copy <source> <destination>
ptx-make copy src/*.go dist/
ptx-make copy config/ dist/config/
```

**Behavior:**
- Supports wildcards (`*`, `**`)
- Creates destination directory if it doesn't exist
- Copies directories recursively
- Preserves file permissions on Unix
- Returns exit code 0 on success, 1 on failure

**Implementation Notes:**
- Windows: Uses native Go `os.CopyFile` and `filepath.Walk`
- Unix: Uses native Go for consistency (not shell `cp`)

#### `mkdir` - Create Directory
```bash
ptx-make mkdir <path>
ptx-make mkdir dist/bin/release
```

**Behavior:**
- Creates all parent directories (equivalent to `mkdir -p`)
- No error if directory already exists
- Returns exit code 0 on success

#### `rm` - Remove File or Directory
```bash
ptx-make rm <path>
ptx-make rm dist/
ptx-make rm *.tmp
```

**Behavior:**
- Removes files and directories recursively
- No error if path doesn't exist
- Supports wildcards
- Returns exit code 0 on success

#### `exists` - Check Path Existence
```bash
ptx-make exists <path>
```

**Behavior:**
- Returns exit code 0 if path exists
- Returns exit code 1 if path doesn't exist
- Works for files and directories

**Makefile Usage:**
```makefile
check-config:
	@ptx-make exists config.yaml || echo "Config not found"
```

---

### Build Metadata

#### `version` - Git Version Tag
```bash
ptx-make version
# Output: v1.6.5
# Output: v1.6.4-dirty
# Output: abc1234 (if no tags)
```

**Behavior:**
- Returns current git tag if on tagged commit
- Appends `-dirty` if working directory has changes
- Falls back to short commit hash if no tags
- Equivalent to `git describe --tags --always --dirty`

#### `commit` - Git Commit Hash
```bash
ptx-make commit
# Output: abc1234
```

**Behavior:**
- Returns short (7 character) git commit hash
- Equivalent to `git rev-parse --short HEAD`

#### `timestamp` - UTC Timestamp
```bash
ptx-make timestamp
# Output: 2025-12-02T10:30:00Z
```

**Behavior:**
- Returns current UTC time in ISO 8601 format
- Always UTC (not local time) for reproducible builds
- Format: `YYYY-MM-DDTHH:MM:SSZ`

---

### Checksum

#### `checksum` - Generate SHA256 Checksums
```bash
ptx-make checksum <directory> [output-file]
ptx-make checksum dist/
ptx-make checksum dist/ checksums.sha256
```

**Behavior:**
- Generates SHA256 checksums for all files in directory
- Default output file: `checksums.sha256` in target directory
- Format compatible with `sha256sum -c`:
  ```
  abc123...  filename1.exe
  def456...  filename2.tar.gz
  ```
- Skips subdirectories (flat checksum)
- Returns exit code 0 on success

---

### Permissions

#### `chmod` - Set File Permissions
```bash
ptx-make chmod <mode> <file>
ptx-make chmod 755 script.sh
ptx-make chmod +x build.sh
```

**Behavior:**
- Sets file permissions on Unix systems
- **No-op on Windows** (returns exit code 0, does nothing)
- Supports octal notation (755, 644) and symbolic (+x, -w)
- Allows Makefile to be portable without conditionals

---

### JSON Generation

#### `json` - Generate JSON from Key-Value Pairs
```bash
ptx-make json <key>=<value> [key=value...]
ptx-make json version=1.0.0 platform=linux-amd64 commit=abc1234
```

**Output:**
```json
{
  "version": "1.0.0",
  "platform": "linux-amd64",
  "commit": "abc1234"
}
```

**Behavior:**
- Generates valid JSON object
- Properly escapes special characters
- Supports string values only (quotes handled automatically)
- Output to stdout (use shell redirection for file)

**Makefile Usage:**
```makefile
manifest:
	@ptx-make json version=$(VERSION) platform=$(PLATFORM) > manifest.json
```

---

### Environment Variables

#### `env` - Platform Environment Export
```bash
ptx-make env
```

**Output (Linux):**
```makefile
# Platform variables for Makefile
OS=linux
ARCH=x64
EXE=
SLASH=/
PATHSEP=:
```

**Output (Windows):**
```makefile
# Platform variables for Makefile
OS=windows
ARCH=x64
EXE=.exe
SLASH=\
PATHSEP=;
```

**Behavior:**
- Outputs Makefile-compatible variable assignments
- Can be included directly in Makefile

**Makefile Usage:**
```makefile
# Include platform-specific variables
include $(shell ptx-make env > .make-env && echo .make-env)

# Use variables
build:
	go build -o portunix$(EXE) .
```

**Alternative Usage:**
```makefile
# Inline evaluation
PLATFORM_VARS := $(shell ptx-make env)
$(eval $(PLATFORM_VARS))
```

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    portunix (Main Dispatcher)                    │
│                                                                 │
│  - Command parsing and routing                                  │
│  - Helper binary discovery                                      │
│  - Delegates 'make' command to ptx-make                         │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                │ Dispatcher delegates to:
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        ▼                       ▼                       ▼
┌──────────────┐       ┌──────────────┐        ┌──────────────┐
│ ptx-container│       │  ptx-make    │ ◄─────│ (other)      │
└──────────────┘       └──────────────┘        └──────────────┘
                              │
                              │ Uses
                              ▼
                    ┌────────────────────┐
                    │   src/pkg/platform │
                    │                    │
                    │  - GetOS()         │
                    │  - GetArchitecture()│
                    │  - IsWindows()     │
                    └────────────────────┘
```

### Binary Structure

```
src/helpers/ptx-make/
├── main.go              # Entry point, CLI setup
├── cmd/
│   ├── root.go          # Root command with help
│   ├── copy.go          # File copy operations
│   ├── mkdir.go         # Directory creation
│   ├── rm.go            # File/directory removal
│   ├── exists.go        # Existence check
│   ├── version.go       # Git version info
│   ├── commit.go        # Git commit hash
│   ├── timestamp.go     # UTC timestamp
│   ├── checksum.go      # SHA256 generation
│   ├── chmod.go         # Permission setting
│   ├── json.go          # JSON generation
│   └── env.go           # Environment export
├── internal/
│   ├── fileops/         # File operation utilities
│   ├── git/             # Git interaction utilities
│   └── output/          # Output formatting
└── go.mod
```

### Shared Package Usage

PTX-Make will import `src/pkg/platform` for OS detection:

```go
package cmd

import (
    "portunix.ai/portunix/src/pkg/platform"
)

func chmodCmd(mode string, file string) error {
    if platform.IsWindows() {
        // No-op on Windows
        return nil
    }
    // Unix implementation
    return os.Chmod(file, parseMode(mode))
}
```

---

## Trade-off Analysis

### Option A: Shell Scripts Per Platform
**Current State**

```makefile
ifeq ($(OS),Windows_NT)
    RM = del /Q
    MKDIR = mkdir
else
    RM = rm -rf
    MKDIR = mkdir -p
endif
```

**Pros:**
- ✅ No external dependency
- ✅ Native shell performance

**Cons:**
- ❌ Complex Makefile conditionals
- ❌ Different syntax for each platform
- ❌ Inconsistent behavior between shells
- ❌ Hard to maintain and debug

### Option B: PTX-Make Helper (Proposed)

```makefile
clean:
	ptx-make rm dist/
build:
	ptx-make mkdir dist/bin
	go build -o dist/bin/app$(shell ptx-make env | grep EXE | cut -d= -f2)
```

**Pros:**
- ✅ Single syntax for all platforms
- ✅ Consistent behavior guaranteed
- ✅ Leverages existing Portunix infrastructure
- ✅ Easy to test and maintain
- ✅ Extensible for future needs

**Cons:**
- ⚠️ Requires Portunix installation
- ⚠️ Slight overhead vs native commands (~5ms per call)
- ⚠️ Additional binary to distribute

### Option C: Pure Go Build Tool (Alternative)

Replace Makefile entirely with Go-based build tool (like Mage).

**Pros:**
- ✅ Pure Go, cross-platform by design
- ✅ Full programming language capabilities

**Cons:**
- ❌ Abandons Makefile ecosystem
- ❌ Learning curve for contributors
- ❌ Many existing projects use Makefile
- ❌ CI/CD systems expect Makefile

**Decision**: **Option B (PTX-Make Helper)** provides the best balance of portability, maintainability, and ecosystem compatibility.

---

## Implementation Phases

### Phase 1: Foundation (Week 1)
**Goal**: Create helper skeleton and core file operations

- [ ] Create `src/helpers/ptx-make/` directory structure
- [ ] Implement CLI with Cobra
- [ ] Implement `copy`, `mkdir`, `rm`, `exists` commands
- [ ] Add dispatcher routing in main binary
- [ ] Basic unit tests

**Deliverable**: `ptx-make copy`, `mkdir`, `rm`, `exists` working

### Phase 2: Build Metadata (Week 1-2)
**Goal**: Git and timestamp utilities

- [ ] Implement `version` command (git describe)
- [ ] Implement `commit` command (git rev-parse)
- [ ] Implement `timestamp` command (UTC time)
- [ ] Handle cases without git repository

**Deliverable**: Build metadata commands functional

### Phase 3: Advanced Features (Week 2)
**Goal**: Checksum, chmod, JSON, environment

- [ ] Implement `checksum` command
- [ ] Implement `chmod` command (with Windows no-op)
- [ ] Implement `json` command
- [ ] Implement `env` command

**Deliverable**: All commands implemented

### Phase 4: Integration & Testing (Week 2-3)
**Goal**: Integration testing and documentation

- [ ] Update Portunix `Makefile` to use ptx-make
- [ ] Cross-platform testing (Windows/Linux)
- [ ] Performance benchmarks
- [ ] Documentation and examples

**Deliverable**: Production-ready helper

---

## Makefile Migration Example

### Before (Platform-Specific)

```makefile
ifeq ($(OS),Windows_NT)
    RM = del /Q /S
    MKDIR = mkdir
    CP = copy
    EXE = .exe
    NULL = NUL
else
    RM = rm -rf
    MKDIR = mkdir -p
    CP = cp -r
    EXE =
    NULL = /dev/null
endif

VERSION := $(shell git describe --tags --always --dirty 2>$(NULL))
COMMIT := $(shell git rev-parse --short HEAD 2>$(NULL))

build:
	$(MKDIR) dist$(SEP)bin
	go build -ldflags "-X main.Version=$(VERSION)" -o dist/bin/app$(EXE) .

clean:
	$(RM) dist

checksum:
ifeq ($(OS),Windows_NT)
	certutil -hashfile dist\bin\app.exe SHA256
else
	sha256sum dist/bin/* > checksums.sha256
endif
```

### After (PTX-Make)

```makefile
VERSION := $(shell ptx-make version)
COMMIT := $(shell ptx-make commit)
include $(shell ptx-make env > .env.mk && echo .env.mk)

build:
	ptx-make mkdir dist/bin
	go build -ldflags "-X main.Version=$(VERSION)" -o dist/bin/app$(EXE) .

clean:
	ptx-make rm dist

checksum:
	ptx-make checksum dist/bin
```

---

## Consequences

### Positive Consequences

1. **Cross-Platform Compatibility**
   - Same Makefile works on Windows, Linux, and macOS
   - No platform-specific conditionals needed

2. **Simplified Makefiles**
   - Reduced complexity and line count
   - Easier to read and maintain

3. **Consistent Behavior**
   - Identical semantics across platforms
   - Predictable error handling

4. **Leverages Existing Infrastructure**
   - Uses `src/pkg/platform` for OS detection
   - Follows established helper pattern

5. **Extensible**
   - Easy to add new commands as needed
   - Can grow with project requirements

### Negative Consequences

1. **External Dependency**
   - Requires Portunix to be installed
   - Mitigation: PTX-Make can be distributed standalone

2. **Slight Performance Overhead**
   - Each command invokes a binary (~5ms)
   - Mitigation: Negligible for build operations

3. **Additional Binary**
   - One more helper to maintain and distribute
   - Mitigation: Already have 8 helpers, established pattern

### Risk Mitigation

**Risk**: Projects without Portunix can't use ptx-make
**Mitigation**: Provide standalone ptx-make binary download

**Risk**: Performance impact on large builds
**Mitigation**: Benchmark common operations, optimize hot paths

**Risk**: Missing edge cases in file operations
**Mitigation**: Comprehensive testing, gradual adoption

---

## Success Criteria

### Functional Requirements
- [ ] All 11 commands implemented and tested
- [ ] Windows and Linux platforms fully supported
- [ ] macOS compatibility (best effort)
- [ ] Proper error codes and messages

### Performance Requirements
- [ ] Command startup < 10ms
- [ ] File operations comparable to native commands
- [ ] No memory leaks in long operations

### Quality Requirements
- [ ] Unit test coverage > 80%
- [ ] Integration tests for all commands
- [ ] Documentation with examples
- [ ] Portunix Makefile migrated as proof-of-concept

---

## Related Decisions

- **ADR-014**: Git-like Dispatcher Pattern (helper architecture)
- **ADR-026**: Shared Platform Utilities (platform detection)
- **ADR-025**: PTX-Installer Helper (similar helper pattern)

## Versioning

**Target Release**: Version 1.8.0 or later
- New helper binary, minor version bump appropriate
- Follows established helper pattern

---

## Appendix A: Command Reference Summary

| Command | Description | Windows | Linux |
|---------|-------------|---------|-------|
| `copy <src> <dst>` | Copy files/directories | ✅ Native | ✅ Native |
| `mkdir <path>` | Create directory tree | ✅ Native | ✅ Native |
| `rm <path>` | Remove file/directory | ✅ Native | ✅ Native |
| `exists <path>` | Check existence | ✅ Native | ✅ Native |
| `version` | Git version tag | ✅ Git | ✅ Git |
| `commit` | Git commit hash | ✅ Git | ✅ Git |
| `timestamp` | UTC timestamp | ✅ Native | ✅ Native |
| `checksum <dir>` | SHA256 checksums | ✅ Native | ✅ Native |
| `chmod <mode> <file>` | Set permissions | ⚪ No-op | ✅ Native |
| `json <k=v>...` | Generate JSON | ✅ Native | ✅ Native |
| `env` | Platform variables | ✅ Native | ✅ Native |

Legend: ✅ Full support | ⚪ No-op (intentional)

---

## Approval

**Status**: Awaiting Review
**Architect**: Kurc
**Date**: 2025-12-02
