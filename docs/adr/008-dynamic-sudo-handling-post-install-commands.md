# ADR 008: Dynamic Sudo Handling for Post-Install Commands

**Date**: 2025-09-12  
**Status**: Proposed  
**Author**: Claude Code Assistant  
**Related Issues**: #045 (Node.js Installation Critical Fixes)

## Context

During Node.js installation testing in containerized environments, we discovered a critical issue with post-install commands that assume the availability of `sudo`. The current implementation in `install-packages.json` hardcodes `sudo` commands which fail in:

1. **Container environments** running as root (UID 0)
2. **Systems without sudo** installed or available
3. **Embedded/minimal environments** where sudo is not present

### Current Problematic Implementation

```json
"post_install": [
  "sudo mkdir -p /usr/local/nodejs-20",
  "sudo tar -C /usr/local/nodejs-20 -xf ${downloaded_file} --strip-components=1",
  "sudo ln -sf /usr/local/nodejs-20/bin/node /usr/local/bin/node"
]
```

This approach fails because:
- **Containers often run as root**: `sudo` is unnecessary and may not be installed
- **Hardcoded assumptions**: No flexibility for different execution contexts
- **Poor error handling**: Commands fail without graceful fallback
- **Security concerns**: Unnecessary privilege escalation in some contexts

## Decision

We will implement a **Dynamic Sudo Detection and Template System** that:

1. **Runtime Detection**: Automatically detects if sudo is needed and available
2. **Template Variables**: Uses template placeholders instead of hardcoded commands
3. **Graceful Fallback**: Attempts commands without sudo first, then with sudo if needed
4. **Context Awareness**: Adapts behavior based on execution environment

## Proposed Solution

### 1. Template Variable System

Introduce new template variables in `install-packages.json`:

```json
"post_install": [
  "${sudo_prefix}mkdir -p /usr/local/nodejs-20",
  "${sudo_prefix}tar -C /usr/local/nodejs-20 -xf ${downloaded_file} --strip-components=1",
  "${sudo_prefix}ln -sf /usr/local/nodejs-20/bin/node /usr/local/bin/node",
  "${sudo_prefix}ln -sf /usr/local/nodejs-20/bin/npm /usr/local/bin/npm",
  "${sudo_prefix}ln -sf /usr/local/nodejs-20/bin/npx /usr/local/bin/npx"
]
```

### 2. Smart Sudo Detection

Implement `determineSudoPrefix()` function in installer:

```go
func determineSudoPrefix() string {
    // Check if running as root
    if isRunningAsRoot() {
        return "" // No sudo needed
    }
    
    // Check if sudo is available
    if isSudoAvailable() {
        return "sudo " // Use sudo
    }
    
    // No sudo available - commands may fail, but try without
    return ""
}

func isRunningAsRoot() bool {
    cmd := exec.Command("id", "-u")
    output, err := cmd.Output()
    if err != nil {
        return false
    }
    return strings.TrimSpace(string(output)) == "0"
}

func isSudoAvailable() bool {
    cmd := exec.Command("which", "sudo")
    return cmd.Run() == nil
}
```

### 3. Variable Resolution Enhancement

Extend existing `ResolveVariables()` function to support `${sudo_prefix}`:

```go
func ResolveVariables(template string, variables map[string]string) string {
    // Add sudo_prefix to standard variables
    if _, exists := variables["sudo_prefix"]; !exists {
        variables["sudo_prefix"] = determineSudoPrefix()
    }
    
    // Existing variable resolution logic...
    return resolveTemplate(template, variables)
}
```

### 4. Fallback Strategy (Alternative Approach)

For maximum compatibility, implement smart fallback commands:

```json
"post_install": [
  "mkdir -p /usr/local/nodejs-20 || ${sudo_prefix}mkdir -p /usr/local/nodejs-20",
  "tar -C /usr/local/nodejs-20 -xf ${downloaded_file} --strip-components=1 || ${sudo_prefix}tar -C /usr/local/nodejs-20 -xf ${downloaded_file} --strip-components=1"
]
```

## Implementation Plan

### Phase 1: Core Infrastructure
1. Implement `determineSudoPrefix()` and helper functions in installer.go
2. Enhance `ResolveVariables()` to support `${sudo_prefix}` template variable
3. Add comprehensive logging for sudo detection results

### Phase 2: Package Configuration Updates
1. Update Node.js package definitions to use `${sudo_prefix}` templates
2. Review and update other packages requiring privileged operations
3. Maintain backward compatibility with existing hardcoded commands

### Phase 3: Testing and Validation
1. Test in root containers (Docker/Podman as root)
2. Test in non-root containers with sudo available
3. Test in non-root containers without sudo
4. Test on regular Linux systems with sudo
5. Test edge cases and error conditions

## Benefits

### 1. **Universal Compatibility**
- Works in root containers without sudo
- Works on systems with sudo available
- Graceful handling of systems without sudo

### 2. **Enhanced Security**
- No unnecessary privilege escalation
- Minimal required permissions
- Clear security context awareness

### 3. **Better User Experience**
- Automatic adaptation to environment
- Clear error messages when operations fail
- Reduced manual intervention required

### 4. **Container-First Design**
- Optimized for modern containerized environments
- Supports both Docker and Podman execution contexts
- Compatible with CI/CD pipelines and automation

## Risks and Mitigation

### Risk 1: Command Execution Failures
**Mitigation**: Implement comprehensive error handling and clear error messages indicating permission issues.

### Risk 2: Security Implications
**Mitigation**: Always attempt operations without sudo first, only escalate privileges when necessary.

### Risk 3: Backward Compatibility
**Mitigation**: Maintain support for existing hardcoded sudo commands during transition period.

## Alternative Approaches Considered

### 1. **Hardcoded Fallback Commands**
```json
"post_install": [
  "mkdir -p /path || sudo mkdir -p /path"
]
```
**Rejected**: Verbose, error-prone, and still fails when sudo is not available.

### 2. **Environment-Specific Package Variants**
```json
"container": { "post_install": ["mkdir -p /path"] },
"system": { "post_install": ["sudo mkdir -p /path"] }
```
**Rejected**: Overly complex, requires environment detection at package selection level.

### 3. **Runtime Command Modification**
Modify commands at runtime without templates.
**Rejected**: Less transparent, harder to debug and maintain.

## Success Criteria

1. **Container Compatibility**: Node.js installation succeeds in root containers
2. **System Compatibility**: Node.js installation succeeds on regular Linux systems
3. **Error Handling**: Clear error messages when operations fail due to permissions
4. **Performance**: No significant overhead in sudo detection
5. **Maintainability**: Easy to apply to other packages requiring privileged operations

## Implementation Timeline

- **Week 1**: Implement core sudo detection infrastructure
- **Week 2**: Update Node.js package configuration and test
- **Week 3**: Apply to other affected packages and comprehensive testing
- **Week 4**: Documentation updates and deployment

## Future Considerations

1. **Cross-Platform Support**: Extend to Windows (UAC) and macOS privilege handling
2. **Capability Detection**: Detect specific system capabilities beyond just sudo
3. **User Prompts**: Interactive prompts for privilege escalation when appropriate
4. **Audit Logging**: Log all privilege escalation decisions for security auditing

---

**Decision**: Approved for implementation in Issue #045  
**Next Steps**: Begin implementation of Phase 1 - Core Infrastructure  
**Review Date**: 2025-09-19 (1 week post-implementation)