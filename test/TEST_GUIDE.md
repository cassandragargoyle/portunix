# Portunix Testing Guide

Comprehensive testing guide for developers working on Portunix.

## 🚀 Quick Start

### Prerequisites
```bash
# Install required tools
make deps

# Setup test environment
make dev-setup

# Verify setup
make status
```

### Running Tests
```bash
# All tests
make test

# Unit tests only (fast)
make test-unit

# Integration tests (requires Docker)
make test-integration

# With coverage
make test-coverage
```

## 📁 Test Structure

```
Portunix/
├── pkg/                    # Testable packages
│   └── docker/
│       ├── docker.go
│       ├── docker_test.go           # Unit tests
│       └── docker_integration_test.go # Integration tests
├── test/                   # Test infrastructure
│   ├── fixtures/           # Test data
│   ├── mocks/             # Generated mocks
│   └── integration/       # Integration test suites
├── internal/
│   └── testutils/         # Test utilities
└── .github/workflows/     # CI/CD pipelines
```

## 🧪 Test Categories

### 1. Unit Tests (`//go:build unit`)
Fast, isolated tests with mocked dependencies:

```go
//go:build unit
// +build unit

func TestDockerInstall_ValidConfig_Success(t *testing.T) {
    // Arrange
    mockDetector := new(MockSystemDetector)
    mockDetector.On("IsDockerInstalled").Return(false)
    
    // Act
    err := InstallDockerWithDetector(mockDetector, true)
    
    // Assert
    assert.NoError(t, err)
    mockDetector.AssertExpectations(t)
}
```

**Run unit tests:**
```bash
go test -tags=unit ./...
# or
make test-unit
```

### 2. Integration Tests (`//go:build integration`)
Tests with real Docker containers and external dependencies:

```go
//go:build integration
// +build integration

func TestDockerContainer_RealFlow_Success(t *testing.T) {
    testutils.SkipIfNoDocker(t)
    
    // Use testcontainers for real Docker testing
    container, err := testcontainers.GenericContainer(ctx, req)
    require.NoError(t, err)
    defer container.Terminate(ctx)
    
    // Test real container operations
}
```

**Run integration tests:**
```bash
go test -tags=integration ./...
# or  
make test-integration
```

### 3. CLI Tests
Test command-line interface functionality:

```bash
# Test CLI commands
make test-cli

# Manual CLI testing
./portunix docker --help
./portunix docker run-in-container empty --image alpine:latest
```

## 🎭 Mocking and Interfaces

### Creating Mockable Interfaces
```go
//go:generate mockgen -source=interfaces.go -destination=../../test/mocks/docker_mock.go

type DockerClient interface {
    PullImage(ctx context.Context, image string) error
    CreateContainer(ctx context.Context, config ContainerConfig) (string, error)
}
```

### Using Mocks in Tests
```go
func TestWithMock(t *testing.T) {
    // Create mock
    mockClient := new(MockDockerClient)
    
    // Set expectations
    mockClient.On("PullImage", mock.Anything, "alpine:latest").Return(nil)
    
    // Use mock in test
    err := mockClient.PullImage(context.Background(), "alpine:latest")
    
    // Verify expectations
    assert.NoError(t, err)
    mockClient.AssertExpectations(t)
}
```

### Generate Mocks
```bash
go generate ./...
# or
make mocks
```

## 🔧 Test Utilities

### Helper Functions
```go
import "portunix.cz/internal/testutils"

func TestWithHelpers(t *testing.T) {
    // Create temp directory
    tempDir := testutils.CreateTempDir(t, "test-prefix")
    
    // Skip if Docker unavailable
    testutils.SkipIfNoDocker(t)
    
    // Wait for condition
    testutils.WaitForCondition(t, func() bool {
        return containerIsReady()
    }, 30*time.Second, "container ready")
    
    // Create test cache
    cacheDir := testutils.SetupTestCache(t)
}
```

### Test Fixtures
```go
// Create test data
testutils.CreateTestFixture(t, 
    "test/fixtures/docker/valid_dockerfile", 
    testutils.MockDockerfile("ubuntu:22.04"))

testutils.CreateTestFixture(t, 
    "test/fixtures/install/config.json", 
    testutils.MockPackageJSON())
```

## 📊 Coverage and Quality

### Coverage Reports
```bash
# Generate coverage
make test-coverage

# View HTML report
open coverage.html

# Coverage summary
go tool cover -func=coverage.out
```

### Quality Checks
```bash
# Linting
make lint

# Security scan
make security

# All quality checks
make ci-test
```

### Coverage Requirements
- **Unit tests**: ≥ 80%
- **Integration tests**: ≥ 60%
- **Critical paths**: ≥ 95%
- **Error handling**: 100%

## 🐳 Docker Testing

### Testing Docker Functionality

**Without Docker (mocked):**
```bash
go test -tags=unit ./pkg/docker/
```

**With Docker (real containers):**
```bash
go test -tags=integration ./pkg/docker/
```

### Docker Test Utilities
```go
// Check Docker availability
if !testutils.IsDockerAvailable() {
    t.Skip("Docker not available")
}

// Use testcontainers
container := testcontainers.GenericContainer(ctx, 
    testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "alpine:latest",
            Cmd:   []string{"sleep", "30"},
        },
        Started: true,
    })
```

## 🚀 Performance Testing

### Benchmarks
```go
func BenchmarkDockerOperation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        performOperation()
    }
}
```

**Run benchmarks:**
```bash
make benchmark

# Docker specific benchmarks
make benchmark-docker

# Compare benchmarks
go test -bench=. -count=3 ./... | tee benchmark.txt
```

### Performance Targets
- **Container creation**: < 30s
- **Image pull**: < 2min (depends on size)
- **Unit test suite**: < 30s
- **Integration test suite**: < 10min

## 🔄 CI/CD Integration

### GitHub Actions Pipeline
Tests run automatically on:
- Push to main/master/develop
- Pull requests
- Feature branch pushes

### Pipeline Stages
1. **Lint** - Code quality and formatting
2. **Unit Tests** - Fast isolated tests
3. **Integration Tests** - Real Docker testing
4. **Security** - Vulnerability scanning
5. **Cross-Platform** - Windows/Linux/macOS
6. **Coverage** - Coverage reporting
7. **Quality Gate** - Final verification

### Local CI Simulation
```bash
# Run full CI pipeline locally
make ci-test

# Setup CI environment
make ci-setup
```

## 🛠️ Development Workflow

### TDD Workflow
1. **Write failing test**
   ```bash
   go test -tags=unit ./pkg/docker/ -run TestNewFeature
   ```

2. **Implement minimum code**
   ```go
   func NewFeature() error {
       return nil // Make test pass
   }
   ```

3. **Refactor and improve**
   ```bash
   make test-unit
   make lint
   ```

### Adding New Tests

**1. Unit Test:**
```go
//go:build unit
// +build unit

func TestNewFunction_ValidInput_Success(t *testing.T) {
    // Arrange
    input := "test"
    expected := "expected result"
    
    // Act
    result := NewFunction(input)
    
    // Assert
    assert.Equal(t, expected, result)
}
```

**2. Integration Test:**
```go
//go:build integration
// +build integration

func TestNewFeature_Integration_Success(t *testing.T) {
    testutils.SkipIfNoDocker(t)
    
    // Test with real dependencies
}
```

### Test Naming Convention
```
TestFunction_Scenario_ExpectedResult
TestDockerInstall_ValidLinuxSystem_Success
TestPackageManager_Ubuntu_ReturnsApt
TestContainerCreate_InvalidConfig_ReturnsError
```

## 🐛 Debugging Tests

### Verbose Output
```bash
go test -v ./...
go test -v -tags=integration ./pkg/docker/
```

### Debug Specific Test
```bash
go test -v -run TestSpecificFunction ./pkg/docker/
```

### Test with Debugging
```go
func TestDebug(t *testing.T) {
    if testing.Verbose() {
        log.SetOutput(os.Stdout)
    }
    
    // Your test code
    t.Logf("Debug info: %s", debugInfo)
}
```

### Common Issues

**1. Docker not available:**
```bash
# Check Docker status
docker version
make docker-test-env
```

**2. Permission issues:**
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
```

**3. Test flakiness:**
```bash
# Run multiple times
go test -count=10 ./...
```

## 📈 Monitoring and Metrics

### Test Reports
```bash
# Generate comprehensive test report
make test-report

# View results
cat test_report.txt
```

### Coverage Badge
Coverage badge is automatically updated in README.md via CI/CD.

### Quality Dashboard
- **Codecov**: Coverage tracking
- **SonarCloud**: Code quality
- **GitHub Actions**: Build status

## 🔍 Best Practices

### DO ✅
- Write tests for all public functions
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Test error conditions
- Use descriptive test names
- Keep tests fast and focused
- Use testify for assertions
- Clean up resources in tests

### DON'T ❌
- Test implementation details
- Write flaky tests
- Ignore test failures
- Skip test cleanup
- Use real external services in unit tests
- Write tests that depend on each other
- Commit sensitive data in test fixtures

### Test Organization
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "result",
            wantErr:  false,
        },
        {
            name:     "invalid input", 
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## 📚 References

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Framework](https://github.com/stretchr/testify)
- [Testcontainers Go](https://golang.testcontainers.org/)
- [GoMock](https://github.com/golang/mock)
- [Go Test Best Practices](https://github.com/golang/go/wiki/TestComments)

---

**Happy Testing! 🧪✨**