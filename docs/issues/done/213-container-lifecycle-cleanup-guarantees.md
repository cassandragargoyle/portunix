# Issue #213: Container Lifecycle Management with Cleanup Guarantees

> **Renumbered:** formerly internal issue #027. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

**Status**: ✅ Implemented  
**Priority**: High  
**Type**: Enhancement  
**Labels**: docker, lifecycle-management, cleanup, resource-management, testing  
**Requested by**: portunix-portunixcredits development team  
**Created**: 2025-01-09  
**Closed**: 2026-05-09  
**Acceptance**: [acceptance-027.md](../../testing/acceptance-027.md) — PASS (18/20)  

## Problem Statement

The current Portunix Docker container management lacks automatic cleanup guarantees and lifecycle management capabilities. This creates several challenges for development teams, particularly for testing scenarios:

### Current Limitations
- No automatic cleanup of containers after tests complete
- No TTL (Time-To-Live) mechanism for temporary containers
- No guaranteed cleanup on process termination or failure
- No resource monitoring and limits enforcement
- Manual cleanup required, leading to container accumulation
- No centralized lifecycle policies

### Impact on Development Teams
- **Testing environments**: Containers accumulate during CI/CD runs
- **Trial software testing**: VM snapshots and containers persist indefinitely
- **Resource exhaustion**: Disk space and memory consumed by orphaned containers
- **Development workflows**: Manual intervention required for cleanup
- **Multi-node testing**: No coordinated cleanup across distributed test environments

## Requirements

### Core Features
1. **Automatic Cleanup Guarantees**
   - Cleanup on process exit (SIGTERM/SIGINT handlers)
   - TTL-based automatic container removal
   - Policy-driven cleanup (on-exit, ttl, manual)
   - Graceful shutdown with cleanup verification

2. **Enhanced Container Configuration**
   - TTL specification for containers
   - Resource limits (memory, CPU)
   - Health checks for container monitoring
   - Auto-cleanup flags and policies

3. **Cleanup Commands**
   - Bulk cleanup with filters (age, pattern, status)
   - Orphaned container detection and removal
   - Force cleanup options
   - Cleanup verification and reporting

4. **Background Service**
   - Continuous monitoring of container lifecycle
   - Automatic enforcement of TTL policies
   - Resource usage monitoring and alerts
   - Cleanup registry maintenance

### Use Cases

#### 1. Cosmos SDK Test Networks
```bash
# Create test network with 2-hour TTL
portunix docker run-in-container go \
  --name cosmos-testnet \
  --ttl 2h \
  --auto-cleanup \
  --max-memory 4G \
  --health-check "curl localhost:26657/status"
```

#### 2. Trial Software Testing
```bash
# VM with automatic cleanup after trial period
portunix vm create trial-vm \
  --ttl 30d \
  --snapshot-before-cleanup \
  --cleanup-policy on-ttl-expire
```

#### 3. CI/CD Test Environments
```bash
# Parallel testing with guaranteed cleanup
portunix docker run-tests \
  --parallel 4 \
  --ttl 1h \
  --cleanup-on-failure \
  --resource-limits "2G,1cpu"
```

#### 4. Development Environment Management
```bash
# Cleanup old development containers
portunix docker cleanup \
  --older-than 7d \
  --pattern "dev-*" \
  --exclude-running \
  --dry-run
```

## Technical Implementation

### 1. Enhanced DockerConfig Structure
```go
type DockerConfig struct {
    // Existing fields...
    
    // Lifecycle management
    TTL               time.Duration // Time-to-live
    AutoCleanup       bool          // Cleanup on exit
    MaxMemory         string        // Memory limit
    MaxCPU            string        // CPU limit  
    HealthCheck       string        // Health check command
    CleanupPolicy     string        // "on-exit", "ttl", "manual"
    CleanupTimeout    time.Duration // Grace period for cleanup
}
```

### 2. New Commands
```go
// cmd/docker_cleanup.go
var dockerCleanupCmd = &cobra.Command{
    Use:   "cleanup",
    Short: "Clean up Docker containers with advanced filtering",
}

// cmd/docker_lifecycle.go  
var dockerLifecycleCmd = &cobra.Command{
    Use:   "lifecycle",
    Short: "Manage container lifecycle policies",
}
```

### 3. Lifecycle Manager Service
```go
// app/docker/lifecycle.go
type LifecycleManager struct {
    registry     ContainerRegistry
    cleanupChan  chan string
    stopChan     chan bool
    policies     map[string]CleanupPolicy
}

func (lm *LifecycleManager) Start() error
func (lm *LifecycleManager) RegisterContainer(config DockerConfig) error
func (lm *LifecycleManager) CleanupExpired() error
func (lm *LifecycleManager) GracefulShutdown() error
```

### 4. Container Registry
```go
// app/docker/registry.go
type ContainerRegistry struct {
    containers map[string]ContainerMetadata
    mutex      sync.RWMutex
}

type ContainerMetadata struct {
    ID           string
    Name         string
    CreatedAt    time.Time
    TTL          time.Duration
    Policy       string
    ResourceUsage ResourceMetrics
}
```

### 5. Signal Handlers
```go
// app/docker/signals.go
func SetupCleanupHandlers() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        log.Println("Received interrupt, cleaning up containers...")
        CleanupAllContainers()
        os.Exit(0)
    }()
}
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1)
- [ ] Extend DockerConfig with lifecycle fields
- [ ] Implement ContainerRegistry for metadata tracking
- [ ] Add basic TTL support to RunInContainer
- [ ] Create signal handlers for graceful cleanup

### Phase 2: Cleanup Commands (Week 2)
- [ ] Implement `portunix docker cleanup` command
- [ ] Add filtering capabilities (age, pattern, status)
- [ ] Create `portunix docker lifecycle` management command
- [ ] Add dry-run and verification options

### Phase 3: Background Service (Week 3)
- [ ] Implement LifecycleManager background service
- [ ] Add continuous TTL monitoring
- [ ] Implement resource usage tracking
- [ ] Add health check integration

### Phase 4: Advanced Features (Week 4)
- [ ] Add VM snapshot integration for cleanup
- [ ] Implement cleanup policies and templates
- [ ] Add metrics and reporting
- [ ] Create cleanup verification and rollback

## Testing Strategy

### Unit Tests
- ContainerRegistry operations
- TTL calculation and expiration logic
- Cleanup policy enforcement
- Signal handler behavior

### Integration Tests
- Container creation with TTL and cleanup
- Background service lifecycle management
- Multi-container cleanup scenarios
- Resource limit enforcement

### End-to-End Tests
- Cosmos SDK testnet with automatic cleanup
- CI/CD pipeline with guaranteed cleanup
- Trial software VM with snapshot cleanup
- Failure scenarios and cleanup verification

## Acceptance Criteria

### Must Have
- [x] Containers can be created with TTL specification
- [x] Automatic cleanup on process termination (SIGTERM/SIGINT)
- [x] Background service monitors and enforces TTL policies
- [x] Cleanup commands with filtering capabilities
- [x] Resource limits integration
- [x] Comprehensive test coverage

### Should Have
- [x] Health check integration for container monitoring
- [x] Cleanup verification and reporting
- [x] Integration with VM snapshot system
- [x] Metrics and usage statistics
- [x] Configuration templates for common use cases

### Nice to Have
- [x] Web dashboard for container lifecycle monitoring
- [x] Integration with external monitoring systems
- [x] Advanced resource scheduling and limits
- [x] Cleanup notifications and alerts

## Documentation Updates

### User Documentation
- Update Docker command documentation with lifecycle options
- Create lifecycle management tutorial
- Add troubleshooting guide for cleanup issues
- Document best practices for container resource management

### Developer Documentation
- API documentation for LifecycleManager
- Integration guide for custom cleanup policies
- Testing patterns for container lifecycle
- Performance considerations and tuning

## Migration Plan

### Backward Compatibility
- All new fields in DockerConfig are optional
- Existing containers continue to work without modification
- Default behavior unchanged for legacy configurations
- Gradual migration path for existing deployments

### Configuration Migration
- Automatic detection of legacy container configurations
- Migration utility for upgrading existing container metadata
- Configuration validation and warning system
- Rollback capability for failed migrations

## Risk Assessment

### Technical Risks
- **Performance impact**: Background service may consume resources
  - *Mitigation*: Configurable monitoring intervals, resource limits
- **Data loss**: Aggressive cleanup may remove important containers
  - *Mitigation*: Dry-run mode, confirmation prompts, backup options
- **Signal handling**: Complex signal handling may cause deadlocks
  - *Mitigation*: Timeout mechanisms, separate cleanup processes

### Operational Risks
- **Learning curve**: New lifecycle concepts may confuse users
  - *Mitigation*: Comprehensive documentation, gradual rollout
- **Configuration complexity**: Many new options may overwhelm users
  - *Mitigation*: Sensible defaults, preset configurations, templates

## Success Metrics

### Quantitative Metrics
- Reduction in orphaned containers by 90%
- Automated cleanup coverage of 95% for test scenarios
- Zero manual cleanup interventions for standard workflows
- Resource usage reduction by 60% in CI/CD environments

### Qualitative Metrics
- Developer satisfaction with cleanup reliability
- Reduced support requests for container management
- Improved CI/CD pipeline stability
- Enhanced development workflow efficiency

## Future Enhancements

### Phase 2 Features
- Integration with Kubernetes for orchestrated cleanup
- Advanced scheduling and resource allocation
- Multi-host container lifecycle coordination
- Integration with cloud provider cost management

### Integration Opportunities
- Jenkins/GitHub Actions plugin integration
- Terraform provider for infrastructure cleanup
- Monitoring system integrations (Prometheus, Grafana)
- Cost tracking and reporting systems

---

**Assignee**: TBD  
**Epic**: Container Management Enhancement  
**Sprint**: TBD  
**Effort Estimate**: 4 weeks (1 senior developer)  
**Dependencies**: None  
**Related Issues**: #002 (Docker Management Command)