# Issue #214: Universal Container Parameters Support

> **Renumbered:** formerly internal issue #028. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

**Status**: 📋 Open  
**Priority**: High  
**Type**: Enhancement  
**Labels**: docker, podman, volume-mounting, cli, container-runtime  
**Requested by**: portunix-portunixcredits development team  
**Created**: 2025-01-09  

## Problem Statement

The current Portunix container commands (`docker run-in-container` and `podman run-in-container`) do not support direct Docker/Podman parameter passthrough, specifically volume mounting with `-v` flag and other common container parameters. This limits the flexibility of container operations and forces users to use runtime-specific commands instead of a unified Portunix interface.

### Current Limitations
- Volume mounting requires manual Docker/Podman command construction
- No unified parameter interface for container operations
- Different parameter syntax between Docker and Podman not handled
- Limited container configuration options through Portunix
- Developers forced to bypass Portunix for advanced container features

### Impact on Development Teams
- **Python integration**: Cannot use volume mounting in automated scripts
- **CI/CD pipelines**: Limited container configuration flexibility
- **Development workflows**: Manual workarounds required for volume mounting
- **Cross-platform compatibility**: Different behavior between Docker and Podman environments

## Current Behavior

### What doesn't work:
```python
# This fails - volume parameter not recognized
cmd = [
    "portunix", "docker", "run-in-container", "go",
    "--name", "test-container",
    "-v", f"{os.getcwd()}:/workspace",  # ❌ Not supported
    "--keep-running"
]
```

### Current workaround:
```python
# Manual Docker command construction
cmd = [
    "docker", "run", "-d", 
    "-v", f"{os.getcwd()}:/workspace",
    "--name", "test-container",
    "golang:latest"
]
```

## Requirements

### Core Features
1. **Universal Volume Mounting**
   - Support `-v` parameter for volume mounting
   - Automatic translation between Docker/Podman syntax differences
   - Path validation and normalization
   - Support for multiple volume mounts

2. **Container Parameter Passthrough**
   - Support common Docker/Podman parameters
   - Parameter validation and translation
   - Error handling for unsupported parameters
   - Documentation of supported parameters

3. **Cross-Runtime Compatibility**
   - Automatic detection of available container runtime
   - Parameter translation when syntax differs
   - Graceful fallback between Docker and Podman
   - Warning messages for runtime-specific features

### Supported Parameters Priority

#### High Priority (Must Have)
- `-v, --volume`: Volume mounting
- `-p, --port`: Port mapping  
- `-e, --env`: Environment variables
- `--name`: Container naming
- `-d, --detach`: Background execution
- `--rm`: Auto-remove container

#### Medium Priority (Should Have)
- `-w, --workdir`: Working directory
- `--privileged`: Privileged mode
- `--network`: Network configuration
- `-u, --user`: User specification
- `--restart`: Restart policy
- `--memory`: Memory limits

#### Low Priority (Nice to Have)
- `--cpus`: CPU limits
- `--hostname`: Hostname setting
- `--dns`: DNS configuration
- `--add-host`: Host entries
- `--cap-add/--cap-drop`: Capability management

## Technical Implementation

### 1. Enhanced Command Structure
```go
// cmd/docker_run_in_container.go enhancement
var dockerRunCmd = &cobra.Command{
    Use:   "run-in-container [installation-type] [container-args...]",
    Short: "Run installation in Docker container with Docker parameter support",
    Run: func(cmd *cobra.Command, args []string) {
        // Parse Docker/Podman parameters
        config, containerArgs := parseContainerArgs(args)
        
        // Validate and translate parameters
        translatedArgs, err := translateContainerParams(containerArgs)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }
        
        // Apply to container configuration
        applyContainerParams(&config, translatedArgs)
        
        // Run with enhanced configuration
        err = docker.RunInContainer(config)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }
    },
}
```

### 2. Parameter Parser and Translator
```go
// app/container/params.go
type ContainerParams struct {
    Volumes      []VolumeMount
    Ports        []PortMapping
    Environment  []EnvVar
    WorkingDir   string
    User         string
    Privileged   bool
    Network      string
    Memory       string
    CPUs         string
}

type VolumeMount struct {
    HostPath      string
    ContainerPath string
    ReadOnly      bool
}

func parseContainerArgs(args []string) (DockerConfig, []string) {
    // Separate installation type from container args
    // Parse Docker-style arguments
}

func translateContainerParams(args []string) (*ContainerParams, error) {
    // Parse Docker/Podman style arguments
    // Validate paths and values  
    // Return structured parameters
}
```

## Usage Examples

### Volume Mounting
```bash
# Universal volume mounting
portunix docker run-in-container go \
  --name dev-container \
  -v "$(pwd):/workspace" \
  -v "/tmp:/tmp" \
  --keep-running

# Multiple volumes with read-only
portunix podman run-in-container python \
  -v "/data:/app/data:ro" \
  -v "/logs:/app/logs" \
  --workdir /app
```

### Python Integration
```python
import subprocess
import os

def create_development_container(project_path, container_name):
    cmd = [
        "portunix", "docker", "run-in-container", "go",
        "--name", container_name,
        "-v", f"{project_path}:/workspace",
        "-p", "8080:8080",
        "-e", "GOPROXY=direct",
        "--workdir", "/workspace",
        "--keep-running"
    ]
    
    result = subprocess.run(cmd, capture_output=True, text=True)
    return result.returncode == 0
```

## Implementation Phases

### Phase 1: Volume Mounting Support (Week 1)
- [ ] Add `-v` parameter parsing to container commands
- [ ] Implement volume path validation and normalization
- [ ] Add volume mounting to DockerConfig and PodmanConfig
- [ ] Update container creation logic to handle volumes

### Phase 2: Common Parameters (Week 2)
- [ ] Add support for `-p` (ports), `-e` (environment), `--name`
- [ ] Implement parameter validation and sanitization
- [ ] Add `--workdir`, `--user`, `--memory` support
- [ ] Create parameter translation layer

### Phase 3: Advanced Parameters (Week 3)
- [ ] Add networking options (`--network`, `--dns`)
- [ ] Implement resource limits (`--cpus`, `--memory`)
- [ ] Add security options (`--privileged`, `--cap-add`)
- [ ] Create runtime-specific parameter handling

### Phase 4: Cross-Runtime Translation (Week 4)
- [ ] Implement Docker/Podman parameter translation
- [ ] Add automatic runtime detection and selection
- [ ] Create parameter compatibility warnings
- [ ] Add comprehensive documentation

## Acceptance Criteria

### Must Have
- [x] Volume mounting works with `-v` parameter
- [x] Port mapping works with `-p` parameter
- [x] Environment variables work with `-e` parameter
- [x] Parameters validated and sanitized
- [x] Docker and Podman compatibility
- [x] Python integration examples work

### Should Have
- [x] Working directory specification (`--workdir`)
- [x] User specification (`--user`)
- [x] Memory and CPU limits
- [x] Comprehensive parameter documentation
- [x] Error messages for invalid parameters

## Migration Plan

### Backward Compatibility
- All existing container commands continue to work unchanged
- New parameter support is additive and optional
- Default behavior remains the same
- Existing scripts and automation are not affected

## Success Metrics

### Quantitative Metrics
- 95% of common Docker/Podman parameters supported
- Zero breaking changes to existing container commands
- 90% reduction in need for direct Docker/Podman commands

### Qualitative Metrics
- Improved developer experience for container operations
- Simplified Python integration and automation
- Enhanced cross-platform container compatibility

---

**Assignee**: TBD  
**Epic**: Container Management Enhancement  
**Sprint**: TBD  
**Effort Estimate**: 4 weeks (1 senior developer)  
**Dependencies**: Issue #027 (Container Lifecycle Management)  
**Related Issues**: #002 (Docker Management Command), #003 (Podman Management Command)