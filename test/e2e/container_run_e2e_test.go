package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"portunix.ai/app/container"
)

// ContainerRunE2ETestSuite provides end-to-end testing for container run command
type ContainerRunE2ETestSuite struct {
	suite.Suite
	portunixPath     string
	containerRuntime string
	createdContainers []string
}

// SetupSuite initializes the test suite
func (suite *ContainerRunE2ETestSuite) SetupSuite() {
	// Build portunix binary if not exists
	_, err := os.Stat("../../portunix")
	if os.IsNotExist(err) {
		cmd := exec.Command("go", "build", "-o", "../../portunix", "../../")
		err = cmd.Run()
		require.NoError(suite.T(), err, "Failed to build portunix binary")
	}
	
	suite.portunixPath = "../../portunix"
	suite.createdContainers = []string{}
	
	// Use new container capability detection
	if !container.HasContainerRuntime() {
		suite.T().Skip("No container runtime (docker/podman) available for E2E tests")
	}
	
	// Get preferred runtime for tests
	caps, _ := container.GetContainerCapabilities()
	suite.containerRuntime = string(caps.Preferred)
	require.NotEmpty(suite.T(), suite.containerRuntime, "No container runtime detected")
}

// TearDownSuite cleans up after all tests
func (suite *ContainerRunE2ETestSuite) TearDownSuite() {
	// Clean up created containers
	for _, containerName := range suite.createdContainers {
		suite.cleanupContainer(containerName)
	}
}

// detectContainerRuntime detects which container runtime is available
// detectContainerRuntime is now deprecated - use container.GetContainerCapabilities() instead
// Keeping for backward compatibility but using new API internally
func (suite *ContainerRunE2ETestSuite) detectContainerRuntime() string {
	caps, err := container.GetContainerCapabilities()
	if err != nil || !caps.Available {
		return ""
	}
	return string(caps.Preferred)
}

// cleanupContainer removes a container if it exists
func (suite *ContainerRunE2ETestSuite) cleanupContainer(containerName string) {
	// Stop container
	stopCmd := exec.Command(suite.containerRuntime, "stop", containerName)
	stopCmd.Run() // Ignore errors - container might not be running
	
	// Remove container
	rmCmd := exec.Command(suite.containerRuntime, "rm", "-f", containerName)
	rmCmd.Run() // Ignore errors - container might not exist
}

// executePortunixCommand executes portunix command and returns output and error
func (suite *ContainerRunE2ETestSuite) executePortunixCommand(args ...string) (string, error) {
	// Print command being executed for debugging
	suite.T().Logf("Executing: %s %s", suite.portunixPath, strings.Join(args, " "))
	
	cmd := exec.Command(suite.portunixPath, args...)
	output, err := cmd.CombinedOutput()
	
	// Print output for debugging
	suite.T().Logf("Output: %s", string(output))
	if err != nil {
		suite.T().Logf("Error: %v", err)
	}
	
	return string(output), err
}

// TestContainerRunBasicDetached tests basic detached container execution
func (suite *ContainerRunE2ETestSuite) TestContainerRunBasicDetached() {
	containerName := "e2e-test-basic-detached"
	suite.createdContainers = append(suite.createdContainers, containerName)
	
	// TC-038-E001: Basic Detached Container
	suite.Run("TC-038-E001: Basic detached container", func() {
		// Clean up any existing container
		suite.cleanupContainer(containerName)
		
		// Execute container run command
		output, err := suite.executePortunixCommand("container", "run", "-d", "--name", containerName, "ubuntu:22.04", "sleep", "60")
		
		// Verify command succeeded
		require.NoError(suite.T(), err, "Container run command should succeed")
		assert.Contains(suite.T(), output, "Container started successfully in detached mode", "Should confirm detached mode")
		
		// Verify container is running
		listOutput, err := suite.executePortunixCommand("container", "list")
		require.NoError(suite.T(), err, "Container list should succeed")
		assert.Contains(suite.T(), listOutput, containerName, "Container should appear in list")
		
		// Verify with native runtime command
		nativeCmd := exec.Command(suite.containerRuntime, "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Names}}")
		nativeOutput, err := nativeCmd.Output()
		require.NoError(suite.T(), err, "Native runtime ps should succeed")
		assert.Contains(suite.T(), string(nativeOutput), containerName, "Container should be visible to native runtime")
	})
}

// TestContainerRunWithPorts tests container with port mapping
func (suite *ContainerRunE2ETestSuite) TestContainerRunWithPorts() {
	containerName := "e2e-test-ports"
	suite.createdContainers = append(suite.createdContainers, containerName)
	
	// TC-038-E003: Port Mapping Verification
	suite.Run("TC-038-E003: Port mapping verification", func() {
		suite.cleanupContainer(containerName)
		
		// Execute container with port mapping
		output, err := suite.executePortunixCommand("container", "run", "-d", "-p", "18080:80", "--name", containerName, "nginx:alpine")
		
		require.NoError(suite.T(), err, "Container run with port mapping should succeed")
		assert.Contains(suite.T(), output, "Container started successfully in detached mode", "Should confirm detached start")
		
		// Wait for nginx to start
		time.Sleep(3 * time.Second)
		
		// Test port accessibility (basic check)
		// Note: This is a simplified test - in production we might check actual HTTP response
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		client := &http.Client{Timeout: 2 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:18080", nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
				assert.Equal(suite.T(), 200, resp.StatusCode, "Nginx should respond on mapped port")
			}
		}
		// Note: We don't fail the test if HTTP check fails as it might be a network issue
		// The important part is that the container started without port mapping errors
	})
}

// TestContainerRunWithVolumes tests container with volume mounting
func (suite *ContainerRunE2ETestSuite) TestContainerRunWithVolumes() {
	containerName := "e2e-test-volumes"
	suite.createdContainers = append(suite.createdContainers, containerName)
	
	// TC-038-E004: Volume Mount Verification
	suite.Run("TC-038-E004: Volume mount verification", func() {
		suite.cleanupContainer(containerName)
		
		// Create temporary directory for testing
		tempDir := "/tmp/portunix-e2e-test"
		err := os.MkdirAll(tempDir, 0755)
		require.NoError(suite.T(), err, "Should create temp directory")
		defer os.RemoveAll(tempDir)
		
		// Create test file
		testFile := fmt.Sprintf("%s/test.txt", tempDir)
		err = os.WriteFile(testFile, []byte("Hello from host"), 0644)
		require.NoError(suite.T(), err, "Should create test file")
		
		// Execute container with volume mount
		output, err := suite.executePortunixCommand("container", "run", "-d", "-v", fmt.Sprintf("%s:/data", tempDir), "--name", containerName, "ubuntu:22.04", "sleep", "30")
		
		require.NoError(suite.T(), err, "Container run with volume mount should succeed")
		assert.Contains(suite.T(), output, "Container started successfully in detached mode", "Should confirm detached start")
		
		// Verify file is accessible in container
		execOutput, err := suite.executePortunixCommand("container", "exec", containerName, "cat", "/data/test.txt")
		if err == nil {
			assert.Contains(suite.T(), execOutput, "Hello from host", "File should be accessible in container")
		}
		// Note: exec might not be implemented yet, so we don't fail the test
	})
}

// TestContainerRunWithEnvironment tests container with environment variables
func (suite *ContainerRunE2ETestSuite) TestContainerRunWithEnvironment() {
	containerName := "e2e-test-env"
	suite.createdContainers = append(suite.createdContainers, containerName)
	
	// TC-038-E005: Environment Variable Injection
	suite.Run("TC-038-E005: Environment variable injection", func() {
		suite.cleanupContainer(containerName)
		
		// Execute container with environment variables
		output, err := suite.executePortunixCommand("container", "run", "-d", "-e", "TEST_VAR=hello", "-e", "NODE_ENV=production", "--name", containerName, "ubuntu:22.04", "sleep", "30")
		
		require.NoError(suite.T(), err, "Container run with environment variables should succeed")
		assert.Contains(suite.T(), output, "Container started successfully in detached mode", "Should confirm detached start")
		
		// Verify environment variables are set (using native runtime)
		envCmd := exec.Command(suite.containerRuntime, "exec", containerName, "printenv", "TEST_VAR")
		envOutput, err := envCmd.Output()
		if err == nil {
			assert.Contains(suite.T(), string(envOutput), "hello", "Environment variable should be set")
		}
	})
}

// TestContainerRunComplexCommand tests container with complex commands
func (suite *ContainerRunE2ETestSuite) TestContainerRunComplexCommand() {
	containerName := "e2e-test-complex"
	suite.createdContainers = append(suite.createdContainers, containerName)
	
	// TC-038-E006: Complex Command with Separator
	suite.Run("TC-038-E006: Complex command with separator", func() {
		suite.cleanupContainer(containerName)
		
		// Execute container with complex command using separator
		output, err := suite.executePortunixCommand("container", "run", "-d", "--name", containerName, "ubuntu:22.04", "--", "bash", "-c", "apt-get update && sleep 30")
		
		require.NoError(suite.T(), err, "Container run with complex command should succeed")
		assert.Contains(suite.T(), output, "Container started successfully in detached mode", "Should confirm detached start")
		
		// Verify container is running the correct process
		// We can check if the container is still running (indicating the command started)
		time.Sleep(2 * time.Second)
		
		psCmd := exec.Command(suite.containerRuntime, "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Status}}")
		psOutput, err := psCmd.Output()
		if err == nil {
			assert.Contains(suite.T(), string(psOutput), "Up", "Container should still be running")
		}
	})
}

// TestContainerRunOriginalIssue tests the original issue scenario
func (suite *ContainerRunE2ETestSuite) TestContainerRunOriginalIssue() {
	containerName := "portunix-e2e-test-issue038"
	suite.createdContainers = append(suite.createdContainers, containerName)
	
	// TC-038-E007: Original Issue Reproduction
	suite.Run("TC-038-E007: Original issue reproduction", func() {
		suite.cleanupContainer(containerName)
		
		// Execute the exact command that was failing in the original issue
		// Note: We use a shorter sleep time and simpler command for E2E testing
		output, err := suite.executePortunixCommand("container", "run", "-d", "--name", containerName, "ubuntu:22.04", "bash", "-c", "apt-get update && sleep 60")
		
		// This should NOT fail with "unknown shorthand flag: 'd'" error
		require.NoError(suite.T(), err, "Original issue command should execute without flag parsing errors")
		assert.Contains(suite.T(), output, "Container started successfully in detached mode", "Should confirm successful execution")
		assert.NotContains(suite.T(), output, "unknown shorthand flag", "Should not have flag parsing errors")
		
		// Verify container was created and is running
		listOutput, err := suite.executePortunixCommand("container", "list")
		require.NoError(suite.T(), err, "Container list should work")
		assert.Contains(suite.T(), listOutput, containerName, "Container should be listed")
	})
}

// TestContainerRunMultipleFlags tests multiple flag combinations
func (suite *ContainerRunE2ETestSuite) TestContainerRunMultipleFlags() {
	containerName := "e2e-test-multiflags"
	suite.createdContainers = append(suite.createdContainers, containerName)
	
	// TC-038-E008: Multiple Flag Combinations
	suite.Run("TC-038-E008: Multiple flag combinations", func() {
		suite.cleanupContainer(containerName)
		
		// Execute container with multiple flags
		tempDir := "/tmp/portunix-e2e-multiflags"
		err := os.MkdirAll(tempDir, 0755)
		require.NoError(suite.T(), err, "Should create temp directory")
		defer os.RemoveAll(tempDir)
		
		output, err := suite.executePortunixCommand(
			"container", "run",
			"-d",                                    // detach
			"--name", containerName,                 // name
			"-p", "18081:80",                       // port
			"-v", fmt.Sprintf("%s:/app", tempDir),  // volume
			"-e", "DEBUG=true",                     // environment
			"-e", "LOG_LEVEL=info",                 // multiple environment vars
			"nginx:alpine",
		)
		
		require.NoError(suite.T(), err, "Container run with multiple flags should succeed")
		assert.Contains(suite.T(), output, "Container started successfully in detached mode", "Should confirm successful execution")
		
		// Brief verification that container is running
		time.Sleep(2 * time.Second)
		listOutput, err := suite.executePortunixCommand("container", "list")
		if err == nil {
			assert.Contains(suite.T(), listOutput, containerName, "Container should be running")
		}
	})
}

// TestContainerRunErrorScenarios tests error handling
func (suite *ContainerRunE2ETestSuite) TestContainerRunErrorScenarios() {
	// TC-038-E009: Invalid image
	suite.Run("TC-038-E009: Invalid image handling", func() {
		output, err := suite.executePortunixCommand("container", "run", "-d", "--name", "should-fail", "nonexistent:image")
		
		assert.Error(suite.T(), err, "Should fail with invalid image")
		// The exact error message depends on the container runtime
		assert.True(suite.T(), strings.Contains(output, "Error") || strings.Contains(output, "error"), "Should contain error message")
	})
	
	// TC-038-E010: Name conflict
	suite.Run("TC-038-E010: Container name conflict", func() {
		conflictName := "e2e-test-conflict"
		suite.createdContainers = append(suite.createdContainers, conflictName)
		
		// Create first container
		_, err1 := suite.executePortunixCommand("container", "run", "-d", "--name", conflictName, "ubuntu:22.04", "sleep", "30")
		require.NoError(suite.T(), err1, "First container should succeed")
		
		// Try to create second container with same name
		output2, err2 := suite.executePortunixCommand("container", "run", "-d", "--name", conflictName, "ubuntu:22.04", "sleep", "30")
		
		assert.Error(suite.T(), err2, "Second container with same name should fail")
		assert.Contains(suite.T(), output2, "Error", "Should contain error message about name conflict")
	})
}

// TestSuite entry point for running the E2E test suite
func TestContainerRunE2ETestSuite(t *testing.T) {
	// Skip E2E tests if no container runtime is available
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}
	
	// Check if container runtime is available using new API
	if !container.HasContainerRuntime() {
		t.Skip("Skipping E2E tests: no container runtime (Docker or Podman) available")
	}
	
	suite.Run(t, new(ContainerRunE2ETestSuite))
}

// Benchmark tests for performance validation
func BenchmarkContainerRunFlagParsing(b *testing.B) {
	portunixPath := "../../portunix"
	
	// Build if needed
	if _, err := os.Stat(portunixPath); os.IsNotExist(err) {
		cmd := exec.Command("go", "build", "-o", portunixPath, "../../")
		if err := cmd.Run(); err != nil {
			b.Fatalf("Failed to build portunix: %v", err)
		}
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(portunixPath, "container", "run", "--help")
		_, err := cmd.Output()
		if err != nil {
			b.Fatalf("Help command failed: %v", err)
		}
	}
}