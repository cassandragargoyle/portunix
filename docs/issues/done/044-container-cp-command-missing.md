# Issue #44: Container CP Command Missing from Portunix Container System

## Issue Overview
**ID**: #044  
**Title**: Container CP Command Missing from Portunix Container System  
**Type**: Bug/Enhancement  
**Priority**: High  
**Status**: 📋 Open  
**Created**: 2025-09-12  
**Discovered During**: Issue #041 Node.js/npm Installation Testing

## Problem Description

During container-based testing for Issue #041, it was discovered that the `portunix container cp` command is missing from the available container commands, despite being a critical functionality for container operations.

### Current Situation

```bash
$ ./portunix container --help

Available Commands:
  check            Check container runtime capabilities
  exec             Execute command in container
  info             Show container runtime information
  list             List containers
  logs             Show container logs
  remove           Remove container
  run              Run container
  run-in-container Run installation in container
  start            Start container
  stop             Stop container
  
# MISSING: cp - Copy files to/from container
```

### Impact

This missing functionality forces developers to:
1. Use direct `podman cp` or `docker cp` commands (breaking abstraction)
2. Cannot copy files to/from containers using Portunix unified interface
3. Test automation requires workarounds
4. Violates the principle of complete container abstraction

## Expected Behavior

```bash
# Should work:
portunix container cp local-file.txt container-name:/path/in/container
portunix container cp container-name:/path/in/container local-file.txt

# Similar to:
docker cp local-file.txt container-name:/path/in/container
podman cp local-file.txt container-name:/path/in/container
```

## Proposed Solution

### Implementation Requirements

1. **Add CP Command**: Implement `container cp` subcommand
2. **Bidirectional Support**: Support both directions (to/from container)
3. **Runtime Agnostic**: Work with both Docker and Podman
4. **Consistent Interface**: Match Docker/Podman cp syntax

### Command Structure

```go
// cmd/container_cp.go
var containerCpCmd = &cobra.Command{
    Use:   "cp <source> <destination>",
    Short: "Copy files/folders between container and host",
    Long:  `Copy files to and from containers using configured runtime`,
    Example: `
  # Copy file to container
  portunix container cp file.txt container:/path/to/dest
  
  # Copy file from container
  portunix container cp container:/path/to/file.txt ./local-file.txt
  
  # Copy directory
  portunix container cp ./local-dir container:/path/to/dest`,
    RunE: runContainerCp,
}
```

### Implementation Details

```go
func runContainerCp(cmd *cobra.Command, args []string) error {
    if len(args) != 2 {
        return fmt.Errorf("exactly 2 arguments required: source and destination")
    }
    
    runtime := getConfiguredRuntime() // docker or podman
    
    // Pass through to underlying runtime
    cpCmd := exec.Command(runtime, "cp", args[0], args[1])
    return cpCmd.Run()
}
```

## Acceptance Criteria

- [ ] `portunix container cp` command exists
- [ ] Can copy files from host to container
- [ ] Can copy files from container to host  
- [ ] Can copy directories (recursive)
- [ ] Works with both Docker and Podman runtimes
- [ ] Preserves file permissions and ownership
- [ ] Help text includes usage examples
- [ ] Error handling for non-existent containers/files
- [ ] Command appears in `container --help` output

## Testing Requirements

### Test Cases

1. **Copy file to container**
   ```bash
   echo "test" > test.txt
   portunix container cp test.txt container:/tmp/test.txt
   portunix container exec container cat /tmp/test.txt
   # Should output: test
   ```

2. **Copy file from container**
   ```bash
   portunix container exec container sh -c "echo 'container data' > /tmp/data.txt"
   portunix container cp container:/tmp/data.txt ./retrieved.txt
   cat retrieved.txt
   # Should output: container data
   ```

3. **Copy directory**
   ```bash
   mkdir test-dir && echo "file1" > test-dir/file1.txt
   portunix container cp test-dir container:/tmp/
   portunix container exec container ls /tmp/test-dir/
   # Should list: file1.txt
   ```

4. **Error handling**
   ```bash
   portunix container cp nonexistent.txt container:/tmp/
   # Should error: file not found
   
   portunix container cp test.txt nonexistent-container:/tmp/
   # Should error: container not found
   ```

## Workaround (Temporary)

Until this is implemented, tests must use direct runtime commands:

```go
// Current workaround in tests
runtime := "podman" // or detect dynamically
cmd := exec.Command(runtime, "cp", source, destination)

// Should be:
cmd := exec.Command("portunix", "container", "cp", source, destination)
```

## Related Issues

- **Issue #029**: Universal Container Command Implementation - foundational work
- **Issue #041**: Node.js/npm Installation Support - discovered this issue during testing
- **Issue #043**: Container RM Command Alias - related container command enhancement

## Priority Justification

**High Priority** because:
- **Core Functionality**: File copying is essential for container operations
- **Testing Blocked**: Container-based tests require this functionality
- **Abstraction Broken**: Forces direct runtime usage, breaking Portunix abstraction
- **User Experience**: Missing basic expected container operation

## Implementation Notes

### Quick Implementation Path
1. Add `container_cp.go` in `cmd/` directory
2. Register command in container command group
3. Pass through to underlying runtime (Docker/Podman)
4. Add basic error handling
5. Update help documentation
6. Add integration tests

### Considerations
- Syntax should match Docker/Podman for familiarity
- Support both `container:path` and `path` formats
- Handle Windows path separators if needed
- Consider progress indicators for large files (future enhancement)

## Expected Outcome

After implementation:
- Complete container abstraction maintained
- No need for direct Docker/Podman calls
- Simplified test automation
- Consistent user experience across all container operations
- Full parity with Docker/Podman basic operations

---

**Created**: 2025-09-12  
**Author**: QA/Test Engineer  
**Severity**: High (Core functionality missing)  
**Component**: Container Management System  
**Discovered During**: Container-based test development for Issue #041