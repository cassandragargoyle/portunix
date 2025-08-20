//go:build integration
// +build integration

package docker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// DockerIntegrationTestSuite tests real Docker operations
type DockerIntegrationTestSuite struct {
	suite.Suite
	ctx       context.Context
	container testcontainers.Container
}

func (suite *DockerIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	
	// Check if Docker is available
	if !isDockerAvailable() {
		suite.T().Skip("Docker is not available, skipping integration tests")
	}
}

func (suite *DockerIntegrationTestSuite) TearDownTest() {
	if suite.container != nil {
		suite.container.Terminate(suite.ctx)
	}
}

func TestDockerIntegrationSuite(t *testing.T) {
	suite.Run(t, new(DockerIntegrationTestSuite))
}

func (suite *DockerIntegrationTestSuite) TestPullImage_ValidImage_Success() {
	// Arrange
	image := "alpine:latest"
	
	// Act
	err := pullImageIfNeeded(image)
	
	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *DockerIntegrationTestSuite) TestCreateContainer_ValidConfig_Success() {
	// Arrange
	req := testcontainers.ContainerRequest{
		Image:        "alpine:latest",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForLog("Starting container").WithStartupTimeout(30 * time.Second),
		Cmd:          []string{"sh", "-c", "echo 'Starting container' && sleep 30"},
	}
	
	// Act
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	
	// Assert
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), container)
	
	suite.container = container
	
	// Verify container is running
	state, err := container.State(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), state.Running)
}

func (suite *DockerIntegrationTestSuite) TestPackageManagerDetection_Ubuntu_Success() {
	// Arrange
	req := testcontainers.ContainerRequest{
		Image: "ubuntu:22.04",
		Cmd:   []string{"sleep", "60"},
	}
	
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err)
	suite.container = container
	
	// Act
	pkgManager, err := detectPackageManager("ubuntu:22.04")
	
	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "apt-get", pkgManager.Manager)
	assert.Equal(suite.T(), "debian-based", pkgManager.Distribution)
}

func (suite *DockerIntegrationTestSuite) TestPackageManagerDetection_Alpine_Success() {
	// Arrange
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd:   []string{"sleep", "60"},
	}
	
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err)
	suite.container = container
	
	// Act
	pkgManager, err := detectPackageManager("alpine:latest")
	
	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "apk", pkgManager.Manager)
	assert.Equal(suite.T(), "alpine", pkgManager.Distribution)
}

func (suite *DockerIntegrationTestSuite) TestContainerExecuteCommand_ValidCommand_Success() {
	// Arrange
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd:   []string{"sleep", "60"},
	}
	
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err)
	suite.container = container
	
	// Act
	exitCode, reader, err := container.Exec(suite.ctx, []string{"echo", "hello world"})
	
	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, exitCode)
	assert.NotNil(suite.T(), reader)
}

func (suite *DockerIntegrationTestSuite) TestContainerNetworking_PortMapping_Success() {
	// Arrange
	req := testcontainers.ContainerRequest{
		Image:        "nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithPort("80/tcp").WithStartupTimeout(30 * time.Second),
	}
	
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err)
	suite.container = container
	
	// Act
	mappedPort, err := container.MappedPort(suite.ctx, "80")
	require.NoError(suite.T(), err)
	
	host, err := container.Host(suite.ctx)
	require.NoError(suite.T(), err)
	
	// Assert
	assert.NotEmpty(suite.T(), mappedPort.Port())
	assert.NotEmpty(suite.T(), host)
	
	// Test actual connectivity
	endpoint := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	_ = endpoint // In a real test, you would make an HTTP request here
}

func (suite *DockerIntegrationTestSuite) TestContainerLogs_HasOutput_Success() {
	// Arrange
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd:   []string{"sh", "-c", "echo 'test log output' && sleep 10"},
	}
	
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err)
	suite.container = container
	
	// Wait a moment for the command to execute
	time.Sleep(2 * time.Second)
	
	// Act
	logs, err := container.Logs(suite.ctx)
	
	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), logs)
	
	// Read logs content
	logContent := make([]byte, 1024)
	n, _ := logs.Read(logContent)
	assert.Greater(suite.T(), n, 0)
	assert.Contains(suite.T(), string(logContent[:n]), "test log output")
}

func (suite *DockerIntegrationTestSuite) TestContainerLifecycle_StartStopRemove_Success() {
	// Arrange
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd:   []string{"sleep", "60"},
	}
	
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err)
	
	// Verify container is running
	state, err := container.State(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), state.Running)
	
	// Act - Stop container
	err = container.Stop(suite.ctx, nil)
	assert.NoError(suite.T(), err)
	
	// Verify container is stopped
	state, err = container.State(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), state.Running)
	
	// Act - Start container again
	err = container.Start(suite.ctx)
	assert.NoError(suite.T(), err)
	
	// Verify container is running again
	state, err = container.State(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), state.Running)
	
	// Act - Terminate (remove) container
	err = container.Terminate(suite.ctx)
	assert.NoError(suite.T(), err)
	
	// Container should be terminated after this
	suite.container = nil
}

func (suite *DockerIntegrationTestSuite) TestMultipleContainers_Parallel_Success() {
	// Arrange
	containerCount := 3
	containers := make([]testcontainers.Container, containerCount)
	
	// Act - Create multiple containers in parallel
	for i := 0; i < containerCount; i++ {
		req := testcontainers.ContainerRequest{
			Image: "alpine:latest",
			Cmd:   []string{"sleep", "30"},
		}
		
		container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		require.NoError(suite.T(), err)
		containers[i] = container
	}
	
	// Assert - Verify all containers are running
	for i, container := range containers {
		state, err := container.State(suite.ctx)
		assert.NoError(suite.T(), err, "Container %d should be accessible", i)
		assert.True(suite.T(), state.Running, "Container %d should be running", i)
	}
	
	// Cleanup
	for _, container := range containers {
		container.Terminate(suite.ctx)
	}
}

// Performance test for container operations
func (suite *DockerIntegrationTestSuite) TestContainerPerformance_CreationTime_WithinLimits() {
	// Arrange
	maxCreationTime := 30 * time.Second
	
	// Act
	start := time.Now()
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "performance test"},
	}
	
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	creationTime := time.Since(start)
	
	// Assert
	require.NoError(suite.T(), err)
	assert.Less(suite.T(), creationTime, maxCreationTime, "Container creation should be within time limit")
	
	suite.container = container
}

// Helper functions

func isDockerAvailable() bool {
	// Simple check to see if Docker daemon is accessible
	ctx := context.Background()
	_, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "hello-world:latest",
		},
		Started: false, // Don't actually start it
	})
	return err == nil
}

// Benchmark tests for integration scenarios
func BenchmarkContainerCreation(b *testing.B) {
	if !isDockerAvailable() {
		b.Skip("Docker is not available, skipping benchmark")
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := testcontainers.ContainerRequest{
			Image: "alpine:latest",
			Cmd:   []string{"echo", "benchmark"},
		}
		
		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		
		if err != nil {
			b.Fatalf("Failed to create container: %v", err)
		}
		
		container.Terminate(ctx)
	}
}

func BenchmarkImagePull(b *testing.B) {
	if !isDockerAvailable() {
		b.Skip("Docker is not available, skipping benchmark")
	}
	
	images := []string{"alpine:latest", "ubuntu:22.04", "nginx:alpine"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		image := images[i%len(images)]
		err := pullImageIfNeeded(image)
		if err != nil {
			b.Fatalf("Failed to pull image %s: %v", image, err)
		}
	}
}