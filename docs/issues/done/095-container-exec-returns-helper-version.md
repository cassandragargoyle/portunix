# Issue #95: Container exec Returns Helper Version Instead of Executing Command

## 🎯 Priority
**HIGH** - Breaks container command execution functionality

## 📋 Status
- **Created**: 2025-10-04
- **Status**: 🔴 Open
- **Assignee**: TBD
- **Branch**: TBD

## 📝 Problem Description

### Current Situation
The `./portunix container exec` command does not execute the specified command inside the container. Instead, it returns the ptx-container helper version information:

```bash
$ ./portunix container exec charming_kirch virsh --version
ptx-container version dev
```

**Expected Behavior:**
```bash
$ ./portunix container exec charming_kirch virsh --version
bash: virsh: command not found
# OR (if installed):
9.0.0
```

**Observed Issues:**
1. ❌ Command not executed inside container
2. ❌ Returns helper version instead of command output
3. ❌ No error message indicating failure
4. ❌ Makes container exec unusable for testing and debugging

### Impact
- Cannot execute commands inside running containers via Portunix CLI
- Breaks integration testing workflows
- Users must fall back to direct docker/podman exec commands
- Discovered during Issue #092 testing

### Context
This issue was discovered during acceptance testing of Issue #092 (libvirt package installation) when attempting to verify software installation inside containers.

## 🎯 Root Cause Analysis Required

### Investigate
1. **Command Routing**: Check how `exec` subcommand routes to ptx-container helper
2. **Argument Parsing**: Verify if command arguments are properly passed to helper
3. **Helper Implementation**: Check ptx-container helper's exec command handling
4. **Version Flag**: Why does exec trigger version output instead of command execution?

### Possible Causes
1. ptx-container helper intercepts `--version` flag from arguments
2. Incorrect argument forwarding from main binary to helper
3. exec subcommand not implemented in ptx-container helper
4. Version check happens before command execution logic

## 🎯 Required Solution

### Expected Flow
```
./portunix container exec <container-name> <command> [args...]
   ↓
Parse: container-name, command, args
   ↓
Route to: ptx-container helper
   ↓
Helper executes: podman exec <container-name> <command> [args...]
   ↓
Return: Command output from inside container
```

### Current Flow (Broken)
```
./portunix container exec charming_kirch virsh --version
   ↓
Route to: ptx-container helper
   ↓
Helper sees: --version flag somewhere
   ↓
Return: ptx-container version dev
```

### Implementation Fix

**Option 1: Fix Argument Parsing**
```go
// In ptx-container helper
func handleExec(args []string) error {
    // Don't check for --version if we're in exec subcommand
    if len(args) < 2 {
        return fmt.Errorf("exec requires container name and command")
    }

    containerName := args[0]
    command := args[1:]

    // Execute command in container
    return executeInContainer(containerName, command)
}
```

**Option 2: Proper Flag Handling**
```go
// Parse flags before version check
if subcommand == "exec" {
    // Skip version check, execute command
    return handleExecCommand(args)
}

// Version check only for main command
if hasVersionFlag(args) {
    showVersion()
    return
}
```

## 🎯 Acceptance Criteria

- [ ] `./portunix container exec <container> <command>` executes command inside container
- [ ] Command output/errors are displayed correctly
- [ ] Exit codes from container commands are preserved
- [ ] `--version` in executed command doesn't trigger helper version
- [ ] Works with complex commands including pipes and redirects
- [ ] Interactive commands work (stdin/stdout/stderr forwarding)

## 📊 Technical Details

### Files to Investigate

**1. Main Binary Container Command Handler**
- Check how exec subcommand is routed to helper
- Verify argument passing to ptx-container

**2. ptx-container Helper**
- Location: Likely in `helpers/ptx-container/`
- Check exec subcommand implementation
- Review version flag handling

**3. Command Routing**
- How main binary calls helpers
- Argument forwarding mechanism

### Standard Container exec Behavior
```bash
# Docker/Podman standard
docker exec container-name command arg1 arg2
podman exec container-name bash -c "echo hello"
podman exec -it container-name /bin/bash

# Should work similarly
./portunix container exec container-name command arg1 arg2
./portunix container exec container-name bash -c "echo hello"
./portunix container exec -it container-name /bin/bash
```

## 🧪 Testing Strategy

### Test Cases

**TC001: Simple Command Execution**
```bash
./portunix container run ubuntu:22.04
# Note container name
./portunix container exec <container-name> echo "Hello from container"
# Expected output: Hello from container
# NOT: ptx-container version dev
```

**TC002: Command with --version Flag**
```bash
./portunix container exec <container-name> virsh --version
# Expected: virsh version output OR command not found error
# NOT: ptx-container version dev
```

**TC003: Command with Multiple Arguments**
```bash
./portunix container exec <container-name> ls -la /etc
# Expected: Directory listing from inside container
```

**TC004: Interactive Command**
```bash
./portunix container exec -it <container-name> bash
# Expected: Interactive bash shell inside container
```

**TC005: Exit Code Preservation**
```bash
./portunix container exec <container-name> false
echo $?
# Expected: Exit code 1
```

**TC006: Stderr Output**
```bash
./portunix container exec <container-name> sh -c "echo error >&2"
# Expected: "error" on stderr
```

## 📚 References

### Related Issues
- **#092**: Libvirt Package Installation (where this was discovered)
- **#094**: Container rm Subcommand Not Recognized (similar CLI issue)

### Standard Implementations
- Docker exec: https://docs.docker.com/engine/reference/commandline/exec/
- Podman exec: https://docs.podman.io/en/latest/markdown/podman-exec.1.html

## 💡 Implementation Notes

### Command Forwarding Pattern
```go
// Correct pattern for exec
func executeExec(containerRuntime, containerName string, command []string) error {
    var cmd *exec.Command

    if containerRuntime == "docker" {
        args := append([]string{"exec", containerName}, command...)
        cmd = exec.Command("docker", args...)
    } else {
        args := append([]string{"exec", containerName}, command...)
        cmd = exec.Command("podman", args...)
    }

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin

    return cmd.Run()
}
```

### Version Flag Handling
```go
// Only show version for main command, not subcommands
func main() {
    args := os.Args[1:]

    // Check version BEFORE parsing subcommands
    if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
        fmt.Println("ptx-container version dev")
        return
    }

    // Now parse subcommand
    if len(args) < 1 {
        showUsage()
        return
    }

    subcommand := args[0]
    switch subcommand {
    case "exec":
        handleExec(args[1:]) // Pass remaining args
    // ... other subcommands
    }
}
```

## ⚠️ Related Bugs

This might be related to:
1. Issue #094 (container rm not working) - similar command routing problems
2. Possible systematic issue with helper argument passing
3. Consider full audit of all container subcommands

## 🎯 Success Metrics
- [ ] All test cases pass
- [ ] Command execution works like docker/podman exec
- [ ] Exit codes preserved
- [ ] Stderr/stdout properly forwarded
- [ ] Interactive commands work

## 📅 Estimated Effort
**Complexity**: Medium
**Time Estimate**: 1-2 hours
- Investigation: 30 minutes
- Fix implementation: 30-60 minutes
- Testing: 30 minutes

## 🔗 Related Files
```
helpers/ptx-container/         # Container helper implementation
cmd/container.go               # Main container command
# Check how exec subcommand is routed to helper
```

---

**Discovery Context**: Found during Issue #092 acceptance testing when attempting to verify libvirt installation inside Ubuntu container using `./portunix container exec charming_kirch virsh --version`.

**User Impact**: Critical - completely breaks container exec functionality, forcing users to bypass Portunix and use docker/podman directly.
