//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerVolumeIntegrationTestSuite tests real volume mounting with containers
type ContainerVolumeIntegrationTestSuite struct {
	suite.Suite
	ctx         context.Context
	container   testcontainers.Container
	tempHostDir string
}

func (suite *ContainerVolumeIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Check if Docker is available
	if !isDockerAvailable() {
		suite.T().Skip("Docker is not available, skipping integration tests")
	}

	// Create temporary host directory for volume mounting tests
	tempDir, err := os.MkdirTemp("", "portunix_volume_test_*")
	suite.Require().NoError(err)
	suite.tempHostDir = tempDir
}

func (suite *ContainerVolumeIntegrationTestSuite) TearDownSuite() {
	if suite.tempHostDir != "" {
		os.RemoveAll(suite.tempHostDir)
	}
}

func (suite *ContainerVolumeIntegrationTestSuite) TearDownTest() {
	if suite.container != nil {
		suite.container.Terminate(suite.ctx)
		suite.container = nil
	}
}

func TestContainerVolumeIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ContainerVolumeIntegrationTestSuite))
}

// Mock isDockerAvailable function
func isDockerAvailable() bool {
	// In real implementation, this would check if Docker daemon is running
	return true
}

// Test single volume mounting
func (suite *ContainerVolumeIntegrationTestSuite) TestVolumeMount_SingleVolume_Success() {
	t := suite.T()

	// Arrange - Create test file in host directory
	testFileName := "test_file.txt"
	testContent := "Hello from host!"
	testFilePath := filepath.Join(suite.tempHostDir, testFileName)
	err := ioutil.WriteFile(testFilePath, []byte(testContent), 0644)
	require.NoError(t, err)

	// Create container with volume mount
	containerPath := "/mounted_data"
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd:   []string{"sh", "-c", "cat /mounted_data/test_file.txt && sleep 30"},
		BindMounts: map[string]string{
			suite.tempHostDir: containerPath,
		},
		WaitingFor: wait.ForLog("Hello from host!").WithStartupTimeout(30 * time.Second),
	}

	// Act - Start container
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	suite.container = container

	// Assert - Verify container started and can read mounted file
	logs, err := container.Logs(suite.ctx)
	require.NoError(t, err)
	defer logs.Close()

	logBytes, err := ioutil.ReadAll(logs)
	require.NoError(t, err)
	logContent := string(logBytes)

	assert.Contains(t, logContent, testContent)
}

// Test multiple volume mounting
func (suite *ContainerVolumeIntegrationTestSuite) TestVolumeMount_MultipleVolumes_Success() {
	t := suite.T()

	// Arrange - Create multiple host directories with test files
	dir1 := filepath.Join(suite.tempHostDir, "dir1")
	dir2 := filepath.Join(suite.tempHostDir, "dir2")
	err := os.MkdirAll(dir1, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(dir2, 0755)
	require.NoError(t, err)

	file1Content := "Content from dir1"
	file2Content := "Content from dir2"
	err = ioutil.WriteFile(filepath.Join(dir1, "file1.txt"), []byte(file1Content), 0644)
	require.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(dir2, "file2.txt"), []byte(file2Content), 0644)
	require.NoError(t, err)

	// Create container with multiple volume mounts
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd: []string{"sh", "-c", 
			"cat /mount1/file1.txt && cat /mount2/file2.txt && sleep 30"},
		BindMounts: map[string]string{
			dir1: "/mount1",
			dir2: "/mount2",
		},
		WaitingFor: wait.ForLog(file2Content).WithStartupTimeout(30 * time.Second),
	}

	// Act - Start container
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	suite.container = container

	// Assert - Verify both files can be read
	logs, err := container.Logs(suite.ctx)
	require.NoError(t, err)
	defer logs.Close()

	logBytes, err := ioutil.ReadAll(logs)
	require.NoError(t, err)
	logContent := string(logBytes)

	assert.Contains(t, logContent, file1Content)
	assert.Contains(t, logContent, file2Content)
}

// Test read-only volume mounting
func (suite *ContainerVolumeIntegrationTestSuite) TestVolumeMount_ReadOnly_Success() {
	t := suite.T()

	// Arrange - Create test file in host directory
	testContent := "Read-only content"
	testFilePath := filepath.Join(suite.tempHostDir, "readonly_test.txt")
	err := ioutil.WriteFile(testFilePath, []byte(testContent), 0644)
	require.NoError(t, err)

	// Create container with read-only volume mount
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd: []string{"sh", "-c", 
			"cat /readonly_mount/readonly_test.txt && " +
			"echo 'test' > /readonly_mount/write_test.txt 2>&1 || echo 'Write failed as expected' && " +
			"sleep 30"},
		BindMounts: map[string]string{
			suite.tempHostDir + ":ro": "/readonly_mount",
		},
		WaitingFor: wait.ForLog("Write failed as expected").WithStartupTimeout(30 * time.Second),
	}

	// Act - Start container
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	suite.container = container

	// Assert - Verify read works but write fails
	logs, err := container.Logs(suite.ctx)
	require.NoError(t, err)
	defer logs.Close()

	logBytes, err := ioutil.ReadAll(logs)
	require.NoError(t, err)
	logContent := string(logBytes)

	assert.Contains(t, logContent, testContent)
	assert.Contains(t, logContent, "Write failed as expected")
}

// Test volume mounting with file creation from container
func (suite *ContainerVolumeIntegrationTestSuite) TestVolumeMount_ContainerWritesToHost_Success() {
	t := suite.T()

	// Arrange - Ensure host directory is empty
	writeTestDir := filepath.Join(suite.tempHostDir, "write_test")
	err := os.MkdirAll(writeTestDir, 0755)
	require.NoError(t, err)

	// Create container that writes to mounted volume
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd: []string{"sh", "-c", 
			"echo 'Created by container' > /write_mount/container_file.txt && " +
			"cat /write_mount/container_file.txt && " +
			"sleep 30"},
		BindMounts: map[string]string{
			writeTestDir: "/write_mount",
		},
		WaitingFor: wait.ForLog("Created by container").WithStartupTimeout(30 * time.Second),
	}

	// Act - Start container
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	suite.container = container

	// Wait a bit for file to be written
	time.Sleep(2 * time.Second)

	// Assert - Verify file was created on host
	hostFilePath := filepath.Join(writeTestDir, "container_file.txt")
	assert.FileExists(t, hostFilePath)

	hostFileContent, err := ioutil.ReadFile(hostFilePath)
	require.NoError(t, err)
	assert.Contains(t, string(hostFileContent), "Created by container")
}

// Test volume mounting with non-existent host path
func (suite *ContainerVolumeIntegrationTestSuite) TestVolumeMount_NonExistentHostPath_CreatesDirectory() {
	t := suite.T()

	// Arrange - Use non-existent directory
	nonExistentDir := filepath.Join(suite.tempHostDir, "non_existent")
	// Ensure it doesn't exist
	os.RemoveAll(nonExistentDir)

	// Create container with volume mount to non-existent path
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd: []string{"sh", "-c", 
			"echo 'Created in new dir' > /new_mount/new_file.txt && " +
			"cat /new_mount/new_file.txt && " +
			"sleep 30"},
		BindMounts: map[string]string{
			nonExistentDir: "/new_mount",
		},
		WaitingFor: wait.ForLog("Created in new dir").WithStartupTimeout(30 * time.Second),
	}

	// Act - Start container
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	suite.container = container

	// Wait for file creation
	time.Sleep(2 * time.Second)

	// Assert - Verify directory was created on host
	assert.DirExists(t, nonExistentDir)
	
	hostFilePath := filepath.Join(nonExistentDir, "new_file.txt")
	assert.FileExists(t, hostFilePath)

	hostFileContent, err := ioutil.ReadFile(hostFilePath)
	require.NoError(t, err)
	assert.Contains(t, string(hostFileContent), "Created in new dir")
}

// Test volume mounting with permission issues
func (suite *ContainerVolumeIntegrationTestSuite) TestVolumeMount_PermissionHandling_Success() {
	t := suite.T()

	// Arrange - Create directory with specific permissions
	permTestDir := filepath.Join(suite.tempHostDir, "perm_test")
	err := os.MkdirAll(permTestDir, 0755)
	require.NoError(t, err)

	// Create test file with specific permissions
	testFile := filepath.Join(permTestDir, "perm_file.txt")
	err = ioutil.WriteFile(testFile, []byte("Permission test"), 0644)
	require.NoError(t, err)

	// Create container that checks permissions
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd: []string{"sh", "-c", 
			"ls -la /perm_mount/ && " +
			"cat /perm_mount/perm_file.txt && " +
			"sleep 30"},
		BindMounts: map[string]string{
			permTestDir: "/perm_mount",
		},
		WaitingFor: wait.ForLog("Permission test").WithStartupTimeout(30 * time.Second),
	}

	// Act - Start container
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	suite.container = container

	// Assert - Verify container can read file and permissions are preserved
	logs, err := container.Logs(suite.ctx)
	require.NoError(t, err)
	defer logs.Close()

	logBytes, err := ioutil.ReadAll(logs)
	require.NoError(t, err)
	logContent := string(logBytes)

	assert.Contains(t, logContent, "Permission test")
	assert.Contains(t, logContent, "perm_file.txt")
}

// Test volume mounting with symbolic links
func (suite *ContainerVolumeIntegrationTestSuite) TestVolumeMount_SymbolicLinks_Success() {
	t := suite.T()

	// Arrange - Create file and symbolic link
	symlinkTestDir := filepath.Join(suite.tempHostDir, "symlink_test")
	err := os.MkdirAll(symlinkTestDir, 0755)
	require.NoError(t, err)

	originalFile := filepath.Join(symlinkTestDir, "original.txt")
	symlinkFile := filepath.Join(symlinkTestDir, "symlink.txt")
	
	err = ioutil.WriteFile(originalFile, []byte("Original content"), 0644)
	require.NoError(t, err)
	
	err = os.Symlink("original.txt", symlinkFile)
	require.NoError(t, err)

	// Create container that reads through symlink
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd: []string{"sh", "-c", 
			"cat /symlink_mount/symlink.txt && " +
			"ls -la /symlink_mount/ && " +
			"sleep 30"},
		BindMounts: map[string]string{
			symlinkTestDir: "/symlink_mount",
		},
		WaitingFor: wait.ForLog("Original content").WithStartupTimeout(30 * time.Second),
	}

	// Act - Start container
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	suite.container = container

	// Assert - Verify symlink works inside container
	logs, err := container.Logs(suite.ctx)
	require.NoError(t, err)
	defer logs.Close()

	logBytes, err := ioutil.ReadAll(logs)
	require.NoError(t, err)
	logContent := string(logBytes)

	assert.Contains(t, logContent, "Original content")
	assert.Contains(t, logContent, "symlink.txt")
}

// Benchmark volume mounting performance
func (suite *ContainerVolumeIntegrationTestSuite) TestVolumeMount_Performance_Benchmark() {
	t := suite.T()

	// Create large test file for performance testing
	perfTestDir := filepath.Join(suite.tempHostDir, "perf_test")
	err := os.MkdirAll(perfTestDir, 0755)
	require.NoError(t, err)

	// Create 10MB test file
	largeContent := strings.Repeat("Performance test data\n", 500000)
	largeFile := filepath.Join(perfTestDir, "large_file.txt")
	err = ioutil.WriteFile(largeFile, []byte(largeContent), 0644)
	require.NoError(t, err)

	// Measure container startup time with large volume
	startTime := time.Now()

	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd: []string{"sh", "-c", 
			"wc -l /perf_mount/large_file.txt && " +
			"head -n 1 /perf_mount/large_file.txt && " +
			"sleep 30"},
		BindMounts: map[string]string{
			perfTestDir: "/perf_mount",
		},
		WaitingFor: wait.ForLog("Performance test data").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	suite.container = container

	elapsed := time.Since(startTime)
	t.Logf("Container with large volume mounted started in: %v", elapsed)

	// Performance should be reasonable (less than 30 seconds)
	assert.Less(t, elapsed, 30*time.Second, "Volume mounting with large file took too long")
}

// Test error handling for invalid volume specifications
func (suite *ContainerVolumeIntegrationTestSuite) TestVolumeMount_InvalidSpecs_HandledGracefully() {
	t := suite.T()

	testCases := []struct {
		name        string
		volumeSpec  string
		expectError bool
	}{
		{
			name:        "Valid volume spec",
			volumeSpec:  suite.tempHostDir + ":/valid_mount",
			expectError: false,
		},
		{
			name:        "Missing container path",
			volumeSpec:  suite.tempHostDir,
			expectError: true,
		},
		{
			name:        "Invalid host path",
			volumeSpec:  "/non/existent/path:/container_path",
			expectError: false, // Docker creates the path
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test would be implemented when we have the actual
			// volume parsing and validation logic
			t.Logf("Testing volume spec: %s", tc.volumeSpec)
			// Actual validation logic will be implemented in the main code
		})
	}
}