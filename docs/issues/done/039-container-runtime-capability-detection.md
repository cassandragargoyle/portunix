# Issue #39: Container Runtime Capability Detection

## Summary
Implement a unified function in Portunix to detect and report container runtime capabilities, eliminating the need for tests and other components to manually check for Docker/Podman availability.

## Problem Statement
Currently, tests and potentially other parts of the codebase manually check for container runtime availability by executing commands like:
```go
hasDocker := exec.Command("docker", "version").Run() == nil
hasPodman := exec.Command("podman", "version").Run() == nil
```

This approach has several issues:
1. **Code duplication** - Same logic repeated in multiple places
2. **Inconsistency** - Different parts might check differently
3. **No central control** - Cannot easily enhance detection logic
4. **Limited information** - Only provides boolean availability, not detailed capabilities

## Proposed Solution
Create a centralized container runtime capability detection system that provides:

### 1. Core Functionality
```go
// Container runtime capabilities structure
type ContainerCapabilities struct {
    Available      bool
    Runtime        string // "docker", "podman", "both", "none"
    DockerVersion  string
    PodmanVersion  string
    Preferred      string // User-configured preference
    Features       map[string]bool // Specific feature support
}

// Main detection function
func GetContainerCapabilities() (*ContainerCapabilities, error)

// Simple check for tests
func HasContainerRuntime() bool
```

### 2. CLI Command
```bash
# Check container runtime capabilities
portunix container check

# Output example:
Container Runtime Status:
  Docker: ✓ Available (version 24.0.5)
  Podman: ✗ Not available
  Preferred: docker (auto-detected)
  
Capabilities:
  - Container run: ✓
  - Volume mounting: ✓
  - Network creation: ✓
  - Compose support: ✓
```

### 3. Integration Points
- **Tests**: Replace manual checks with `HasContainerRuntime()` call
- **Container commands**: Use capabilities to determine available operations
- **Installation**: Suggest container runtime installation if missing
- **MCP tools**: Report capabilities to AI assistants

## Implementation Plan

### Phase 1: Core Detection
1. Create `app/container/capabilities.go` with detection logic
2. Implement version detection for Docker and Podman
3. Add feature detection (compose, buildx, etc.)

### Phase 2: CLI Integration
1. Add `container check` command
2. Integrate with existing `docker` and `podman` commands
3. Update help messages based on available runtimes

### Phase 3: Test Migration
1. Replace manual checks in `test/e2e/container_run_e2e_test.go`
2. Update all other tests using container runtime
3. Create helper test utilities using the new API

### Phase 4: Configuration
1. Add user preference for preferred runtime
2. Store capability cache to avoid repeated checks
3. Add refresh mechanism for capability detection

## Benefits
1. **Single source of truth** for container runtime detection
2. **Better test reliability** - consistent detection across all tests
3. **Enhanced user experience** - clear feedback about available capabilities
4. **Easier maintenance** - central place to update detection logic
5. **AI-friendly** - structured information for AI assistants

## Technical Details

### Detection Logic
```go
func detectDocker() (*RuntimeInfo, error) {
    // Try docker version command
    // Parse version output
    // Check for specific features (compose, buildx)
    // Return structured info
}

func detectPodman() (*RuntimeInfo, error) {
    // Similar to Docker detection
    // Check for podman-compose
    // Check for specific Podman features
}
```

### Test Integration Example
```go
// Before
func TestContainerRun(t *testing.T) {
    hasDocker := exec.Command("docker", "version").Run() == nil
    hasPodman := exec.Command("podman", "version").Run() == nil
    
    if !hasDocker && !hasPodman {
        t.Skip("Skipping E2E tests: no container runtime available")
    }
}

// After
func TestContainerRun(t *testing.T) {
    if !container.HasContainerRuntime() {
        t.Skip("Skipping E2E tests: no container runtime available")
    }
    
    // Optional: Get more details
    caps, _ := container.GetContainerCapabilities()
    if !caps.Features["compose"] {
        t.Skip("Test requires compose support")
    }
}
```

## Acceptance Criteria
- [ ] Container capability detection implemented
- [ ] CLI command `container check` working
- [ ] All tests migrated to use new detection
- [ ] Documentation updated
- [ ] Configuration for runtime preference
- [ ] MCP tool for reporting capabilities

## Priority
High - Improves test reliability and user experience

## Labels
- enhancement
- container
- docker
- podman
- testing
- cli

## References
- Test file: `test/e2e/container_run_e2e_test.go`
- Related issues: #029 (Universal Container Command), #031 (Container Exec), #032 (Container Management)