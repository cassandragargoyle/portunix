# Issue #48: System Info Enhanced Container Detection

## Summary
The `portunix system info` command currently only shows Docker installation status. Users need comprehensive container runtime information including Podman detection and a unified container availability indicator.

## Problem Description
Current output of `portunix system info` only displays:
```
Docker: installed/not installed
```

This is insufficient for environments that use:
- Podman as the primary container runtime
- Both Docker and Podman simultaneously
- Neither but need to know container capability status

## Proposed Solution

### Enhanced Output Format
```
System Information:
...
Capabilities:
PowerShell:   true
Admin:        false
Container Available: true/false

Container Runtimes:
Docker:       installed/not installed
Podman:       installed/not installed

Certificates:
Available:    true/false
HTTPS:        true/false
Path:         /path/to/certificates
...
```

Where `Container Available` is `true` if either Docker OR Podman is installed.

## Implementation Details

### Detection Logic
1. Check for Docker installation (existing functionality)
2. Add Podman detection:
   - Check if `podman` command exists in PATH
   - Verify it's executable
   - Optional: Get version information

3. Set Container Available flag:
   ```go
   containerAvailable := dockerInstalled || podmanInstalled
   ```

### Code Location
- File: `cmd/system.go` (system info command)
- Function: Update the system info display logic
- Dependencies: May use existing detection from `app/docker/` and `app/podman/`

## Benefits
1. **Complete visibility**: Users see all available container runtimes
2. **Better decision making**: Users know if they can run containers at all
3. **Runtime choice**: Users can choose appropriate runtime based on availability
4. **Consistency**: Aligns with Portunix's universal container approach

## Testing Requirements
1. Test on system with only Docker installed
2. Test on system with only Podman installed
3. Test on system with both Docker and Podman
4. Test on system with neither installed
5. Verify correct `Container Available` status in all scenarios

## Acceptance Criteria
- [ ] `portunix system info` shows Docker status
- [ ] `portunix system info` shows Podman status
- [ ] `portunix system info` shows Container Available status
- [ ] Container Available is true when either runtime is present
- [ ] Container Available is false when no runtime is present
- [ ] Output format is clean and readable

## Priority
Medium - Quality of life improvement for container runtime visibility

## Type
Enhancement

## Labels
- enhancement
- container
- system-info
- user-experience
- docker
- podman

## Related Issues
- #029 - Universal Container Command Implementation
- #039 - Container Runtime Capability Detection

## Notes
This enhancement supports Portunix's philosophy of universal container management by providing clear visibility into available container runtimes.