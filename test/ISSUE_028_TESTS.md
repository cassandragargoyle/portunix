# Tests for Issue #028: Universal Container Parameters Support

This document describes the comprehensive test suite created for implementing universal container parameter support in Portunix, addressing requirements from the portunix-portunixcredits development team.

## Test Structure

### Unit Tests (`test/unit/`)

#### `container_params_test.go`
Tests container parameter parsing and validation logic.

**Key Features:**
- Volume mounting parameter parsing (`-v`, `--volume`)
- Port mapping parameter parsing (`-p`, `--port`) 
- Environment variable parsing (`-e`, `--env`)
- Working directory, user, privileged mode parameters
- Memory and CPU resource limit parameters
- Path validation and error handling
- Performance benchmarking

**Test Cases:**
```go
TestParseVolumeMount_ValidSpec_Success()
TestParseVolumeMount_ReadOnlySpec_Success() 
TestParseVolumeMount_MultipleVolumes_Success()
TestParsePortMapping_ValidSpec_Success()
TestParseEnvironment_ValidSpec_Success()
TestParseCombinedParameters_AllParams_Success()
TestValidateVolumePath_ValidPaths_Success()
```

#### `container_translation_test.go`
Tests Docker/Podman parameter translation and runtime compatibility.

**Key Features:**
- Docker parameter translation
- Podman parameter translation  
- Cross-runtime compatibility validation
- Runtime-specific parameter handling
- Parameter validation per runtime
- Translation performance benchmarks

**Test Cases:**
```go
TestDockerTranslator_VolumeTranslation_Success()
TestPodmanTranslator_NetworkTranslation_BridgeToSlirp()
TestCrossRuntimeCompatibility_SameParams_SimilarOutput()
TestNewRuntimeTranslator_Docker_ReturnsDockerTranslator()
```

### Integration Tests (`test/integration/`)

#### `container_volume_integration_test.go`
Real container testing with actual Docker/Podman using testcontainers-go.

**Key Features:**
- Live container volume mounting
- Read/write permission testing
- Symbolic link handling
- Performance testing with large projects
- Error handling with invalid configurations

**Test Cases:**
```go
TestVolumeMount_SingleVolume_Success()
TestVolumeMount_MultipleVolumes_Success()
TestVolumeMount_ReadOnly_Success()
TestVolumeMount_ContainerWritesToHost_Success()
TestVolumeMount_Performance_Benchmark()
```

### End-to-End Tests (`test/e2e/`)

#### `python_integration_test.py`
Tests the actual Python workflows requested by the development team.

**Key Features:**
- Original failing Python workflow simulation
- Complex multi-parameter container setup
- CI/CD automation workflow testing
- Error handling and fallback mechanisms
- Performance testing with realistic scenarios

**Test Cases:**
```python
test_volume_mounting_python_workflow()          # Original failing code
test_complex_python_container_setup()           # Multi-parameter setup
test_python_automation_workflow()               # CI/CD simulation
test_parameter_validation_edge_cases()          # Error handling
test_python_docker_fallback()                   # Fallback strategy
```

## Original Problem Being Solved

The portunix-portunixcredits team reported that this Python code doesn't work:

```python
cmd = [
    "portunix", "docker", "run-in-container", "go",
    "--name", self.container_name,
    "-v", f"{os.getcwd()}:/workspace",  # ❌ Not supported
    "--keep-running"
]
```

## Test Implementation Details

### 1. Parameter Parsing Tests
Tests cover all requested parameters:
- `-v, --volume`: Volume mounting with read-only support
- `-p, --port`: Port mapping with validation
- `-e, --env`: Environment variables
- `--workdir`: Working directory specification
- `--user`: User and group specification
- `--privileged`: Privileged mode flag
- `--network`: Network configuration
- `--memory`, `--cpus`: Resource limits

### 2. Cross-Runtime Translation
Tests ensure compatibility between Docker and Podman:
```go
// Docker translator
args = append(args, "-v", "/host:/container")

// Podman translator (same for basic cases)
args = append(args, "-v", "/host:/container")

// Network translation differs
if params.Network == "bridge" {
    // Docker: uses "bridge"
    // Podman: uses "slirp4netns" 
}
```

### 3. Real Container Testing
Integration tests use testcontainers-go for reliable testing:
```go
req := testcontainers.ContainerRequest{
    Image: "alpine:latest",
    BindMounts: map[string]string{
        suite.tempHostDir: "/mounted_data",
    },
    WaitingFor: wait.ForLog("expected output"),
}
```

### 4. Python Workflow Simulation
E2E tests implement the exact Python patterns requested:
```python
def create_development_container(project_path, container_name):
    cmd = [
        portunix_binary, "docker", "run-in-container", "go",
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

## Running Tests

### Prerequisites
```bash
go build -o portunix .
docker --version  # or podman --version
```

### Unit Tests
```bash
go test -tags unit ./test/unit/...
go test -tags unit -coverprofile=coverage.out ./test/unit/...
```

### Integration Tests
```bash
go test -tags integration ./test/integration/...
```

### End-to-End Python Tests
```bash
cd test/e2e/
python3 python_integration_test.py
```

### All Tests
```bash
# Run complete Issue #028 test suite
go test -tags unit ./test/unit/...
go test -tags integration ./test/integration/...
python3 test/e2e/python_integration_test.py
```

## Test Status and Expected Results

### Current Status (Before Implementation)
- **Unit Tests**: ✅ PASS (testing mock implementations)
- **Integration Tests**: ⚠️ SKIP (if Docker not available) / ❌ FAIL (parameter not implemented)
- **E2E Python Tests**: ❌ FAIL (expected - parameter not supported yet)

### After Implementation
- **Unit Tests**: ✅ PASS (all parsing and validation)
- **Integration Tests**: ✅ PASS (real container operations)  
- **E2E Python Tests**: ✅ PASS (full Python workflow support)

## Test Coverage

### Parameters Covered
- ✅ Volume mounting (`-v`)
- ✅ Port mapping (`-p`)
- ✅ Environment variables (`-e`)
- ✅ Working directory (`--workdir`)
- ✅ User specification (`--user`)
- ✅ Privileged mode (`--privileged`)
- ✅ Network configuration (`--network`)
- ✅ Resource limits (`--memory`, `--cpus`)

### Scenarios Covered
- ✅ Single parameter usage
- ✅ Multiple parameters combined
- ✅ Error handling and validation
- ✅ Cross-runtime compatibility (Docker/Podman)
- ✅ Real-world Python automation workflows
- ✅ Performance with large projects
- ✅ Fallback strategies

### Edge Cases Covered
- ✅ Invalid volume specifications
- ✅ Non-existent host paths
- ✅ Permission handling
- ✅ Symbolic links
- ✅ Resource limit validation
- ✅ Network configuration edge cases

## Implementation Guidance

The tests are designed to guide implementation:

1. **Start with unit tests** - implement parameter parsing
2. **Add integration tests** - implement real container operations
3. **Verify with E2E tests** - ensure Python workflows work
4. **Performance validation** - use benchmarks to ensure reasonable performance

Each test documents expected behavior and will pass when the corresponding functionality is implemented.

## Contributing

### Adding New Test Cases
1. **Unit Tests**: Add to appropriate test suite in `test/unit/`
2. **Integration Tests**: Add to `container_volume_integration_test.go`
3. **E2E Tests**: Add methods to `python_integration_test.py`

### Test Naming Convention
- **Unit**: `Test[Component]_[Scenario]_[Expected]`
- **Integration**: `Test[Feature]_[Scenario]_[Expected]` 
- **E2E**: `test_[workflow]_[scenario]`

---

**Status**: Tests created and ready for implementation validation.
**Issue**: #028 Universal Container Parameters Support  
**Requested by**: portunix-portunixcredits development team
**Implementation Target**: 4 weeks (1 senior developer)