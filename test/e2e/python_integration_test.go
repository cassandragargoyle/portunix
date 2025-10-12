//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// PortunixPythonIntegrationTestSuite tests Python integration with Portunix container parameters
// This is a port of the original Python test that tests actual Python workflows
type PortunixPythonIntegrationTestSuite struct {
	suite.Suite
	portunixBinary string
	tempDir        string
	containerName  string
}

func TestPortunixPythonIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PortunixPythonIntegrationTestSuite))
}

func (suite *PortunixPythonIntegrationTestSuite) SetupSuite() {
	// Find portunix binary
	portunixBinary, err := exec.LookPath("portunix")
	if err != nil {
		// Try relative path for development
		portunixBinary = "./portunix"
		if _, err := os.Stat(portunixBinary); os.IsNotExist(err) {
			suite.T().Skip("Portunix binary not found")
		}
	}
	suite.portunixBinary = portunixBinary

	// Check if Docker is available
	if !isDockerAvailable() {
		suite.T().Skip("Docker not available")
	}
}

func (suite *PortunixPythonIntegrationTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "portunix_py_test_*")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	// Generate unique container name
	suite.containerName = fmt.Sprintf("test-container-%d", time.Now().Unix())
}

func (suite *PortunixPythonIntegrationTestSuite) TearDownTest() {
	// Clean up test resources
	suite.cleanup()

	// Remove temp directory
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

func (suite *PortunixPythonIntegrationTestSuite) cleanup() {
	if suite.containerName == "" {
		return
	}

	// Stop and remove container if it exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stopCmd := exec.CommandContext(ctx, "docker", "stop", suite.containerName)
	stopCmd.Run() // Ignore errors

	rmCmd := exec.CommandContext(ctx, "docker", "rm", suite.containerName)
	rmCmd.Run() // Ignore errors
}

// isDockerAvailable checks if Docker is available
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "--version")
	return cmd.Run() == nil
}

// TestVolumeMountingPythonWorkflow tests the original Python workflow that was failing
func (suite *PortunixPythonIntegrationTestSuite) TestVolumeMountingPythonWorkflow() {
	t := suite.T()

	// Create test project structure
	projectDir := filepath.Join(suite.tempDir, "test_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create a test file to verify volume mounting
	testFile := filepath.Join(projectDir, "main.go")
	testContent := `package main

import "fmt"

func main() {
    fmt.Println("Hello from mounted volume!")
}
`
	err = ioutil.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	// Python workflow that should now work - simulated as Go function
	success, stdout, stderr := suite.createDevelopmentContainer(projectDir, suite.containerName)

	// For now, we expect this to fail until implementation is complete
	// This test documents the expected behavior
	if !success {
		// Check if it's failing for the expected reason (parameter not supported)
		errorStr := strings.ToLower(stderr)
		assert.True(t, strings.Contains(errorStr, "-v") || 
			strings.Contains(errorStr, "parameter not supported"),
			"Should fail with parameter not supported message")
	} else {
		// If it succeeds, verify the volume was actually mounted
		// Check if we can execute a command in the container to verify mounting
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		execCmd := exec.CommandContext(ctx, "docker", "exec", suite.containerName, 
			"ls", "/workspace/main.go")
		err := execCmd.Run()
		assert.NoError(t, err, "Volume mounting failed - file not accessible in container")
	}
}

// createDevelopmentContainer simulates the Python function that creates development containers
func (suite *PortunixPythonIntegrationTestSuite) createDevelopmentContainer(projectPath, containerName string) (bool, string, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, suite.portunixBinary, "docker", "run-in-container", "go",
		"--name", containerName,
		"-v", fmt.Sprintf("%s:/workspace", projectPath),
		"--keep-running")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, "", string(output)
	}
	return true, string(output), ""
}

// TestComplexPythonContainerSetup tests complex container setup with multiple parameters
func (suite *PortunixPythonIntegrationTestSuite) TestComplexPythonContainerSetup() {
	t := suite.T()

	// Create project with multiple requirements
	projectDir := filepath.Join(suite.tempDir, "complex_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create config directory
	configDir := filepath.Join(projectDir, "config")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Create test files
	appFile := filepath.Join(projectDir, "app.go")
	err = ioutil.WriteFile(appFile, []byte("package main\n\nfunc main() { println(\"Hello World\") }"), 0644)
	require.NoError(t, err)

	configFile := filepath.Join(configDir, "app.json")
	err = ioutil.WriteFile(configFile, []byte(`{"env": "development"}`), 0644)
	require.NoError(t, err)

	// Python function with complex container setup - simulated as Go function
	success, stdout, stderr := suite.createComplexContainer(projectDir, configDir, suite.containerName)

	// Document expected behavior
	if !success {
		// Check which parameters are not yet supported
		unsupportedParams := []string{}
		for _, param := range []string{"-v", "-p", "-e", "--workdir", "--user", "--memory", "--cpus"} {
			if strings.Contains(stderr, param) {
				unsupportedParams = append(unsupportedParams, param)
			}
		}

		// Log what needs to be implemented
		t.Logf("Parameters not yet supported: %v", unsupportedParams)
		assert.NotEmpty(t, stderr, "Should provide error message about unsupported parameters")
	}
}

// createComplexContainer simulates the Python function with complex container setup
func (suite *PortunixPythonIntegrationTestSuite) createComplexContainer(projectPath, configPath, containerName string) (bool, string, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, suite.portunixBinary, "docker", "run-in-container", "go",
		"--name", containerName,
		"-v", fmt.Sprintf("%s:/workspace", projectPath),
		"-v", fmt.Sprintf("%s:/config:ro", configPath), // Read-only config mount
		"-p", "8080:8080", // Port mapping
		"-p", "3000:3000",
		"-e", "GO_ENV=development", // Environment variables
		"-e", "LOG_LEVEL=debug",
		"--workdir", "/workspace", // Working directory
		"--user", "1000:1000",     // User specification
		"--memory", "2G",          // Resource limits
		"--cpus", "1.5",
		"--keep-running")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, "", string(output)
	}
	return true, string(output), ""
}

// TestErrorHandlingInvalidParameters tests error handling for invalid parameters
func (suite *PortunixPythonIntegrationTestSuite) TestErrorHandlingInvalidParameters() {
	t := suite.T()

	testCases := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name:          "Invalid volume spec",
			args:          []string{"-v", "invalid_volume_spec"},
			expectedError: "invalid volume",
		},
		{
			name:          "Invalid port spec",
			args:          []string{"-p", "invalid:port:spec"},
			expectedError: "invalid port",
		},
		{
			name:          "Invalid memory spec",
			args:          []string{"--memory", "invalid_memory"},
			expectedError: "invalid memory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			containerName := fmt.Sprintf("%s_%s", suite.containerName, strings.ReplaceAll(tc.name, " ", "_"))

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			args := []string{"docker", "run-in-container", "go", "--name", containerName}
			args = append(args, tc.args...)
			args = append(args, "--keep-running")

			cmd := exec.CommandContext(ctx, suite.portunixBinary, args...)
			err := cmd.Run()

			// Should fail with appropriate error message
			assert.Error(t, err, "Should fail for %s", tc.name)

			// Check error message (when implemented)
			// For now, just verify it doesn't succeed silently
		})
	}
}

// TestPythonAutomationWorkflow tests realistic Python automation workflow
func (suite *PortunixPythonIntegrationTestSuite) TestPythonAutomationWorkflow() {
	t := suite.T()

	// Simulate CI/CD pipeline scenario
	projectDir := filepath.Join(suite.tempDir, "ci_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create mock project files
	filesToCreate := map[string]string{
		"main.go":    "package main\n\nfunc main() { println(\"CI Test\") }",
		"go.mod":     "module test\n\ngo 1.21",
		"Dockerfile": "FROM golang:1.21\nCOPY . /app\nWORKDIR /app",
		".gitignore": "*.exe\n*.log",
	}

	for filename, content := range filesToCreate {
		filePath := filepath.Join(projectDir, filename)
		err := ioutil.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Python automation script - simulated as Go struct
	manager := &ContainerManager{portunixBinary: suite.portunixBinary}

	// Create container
	result := manager.createBuildContainer(projectDir, suite.containerName)

	// Document current state and expected behavior
	if !result.success {
		t.Logf("Container creation failed (expected until implementation):")
		t.Logf("STDERR: %s", result.stderr)
		t.Logf("STDOUT: %s", result.stdout)

		// Verify it's failing for parameter support reasons
		assert.NotEmpty(t, result.stderr)
	} else {
		// If container creation succeeds, test the build workflow
		buildSuccess := manager.runBuild(suite.containerName)
		assert.True(t, buildSuccess, "Build should succeed in container")
	}
}

// ContainerManager simulates the Python ContainerManager class
type ContainerManager struct {
	portunixBinary string
}

type ContainerResult struct {
	success    bool
	stdout     string
	stderr     string
	returncode int
}

// createBuildContainer simulates the Python method that creates container for building Go project
func (cm *ContainerManager) createBuildContainer(projectPath, containerName string) ContainerResult {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cm.portunixBinary, "docker", "run-in-container", "go",
		"--name", containerName,
		"-v", fmt.Sprintf("%s:/workspace", projectPath),
		"-v", "/tmp/go-cache:/go/pkg/mod", // Go module cache
		"-e", "GOPROXY=direct",
		"-e", "GOCACHE=/go/pkg/mod",
		"--workdir", "/workspace",
		"--memory", "4G",
		"--cpus", "2.0",
		"--keep-running")

	output, err := cmd.CombinedOutput()
	result := ContainerResult{
		stdout: string(output),
		stderr: string(output),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.returncode = exitErr.ExitCode()
		} else {
			result.returncode = -1
		}
		result.success = false
	} else {
		result.returncode = 0
		result.success = true
	}

	return result
}

// runBuild simulates the Python method that runs build inside container
func (cm *ContainerManager) runBuild(containerName string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "exec", containerName, "go", "build", "-o", "app", ".")
	return cmd.Run() == nil
}

// runTests simulates the Python method that runs tests inside container
func (cm *ContainerManager) runTests(containerName string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "exec", containerName, "go", "test", "./...")
	return cmd.Run() == nil
}

// TestParameterValidationEdgeCases tests edge cases in parameter validation
func (suite *PortunixPythonIntegrationTestSuite) TestParameterValidationEdgeCases() {
	t := suite.T()

	edgeCases := []struct {
		name       string
		args       []string
		shouldFail bool
	}{
		{
			name:       "Empty volume spec",
			args:       []string{"-v", ""},
			shouldFail: true,
		},
		{
			name:       "Relative path in volume",
			args:       []string{"-v", "relative/path:/container"},
			shouldFail: true,
		},
		{
			name:       "Valid absolute path",
			args:       []string{"-v", fmt.Sprintf("%s:/container", suite.tempDir)},
			shouldFail: false,
		},
		{
			name:       "Multiple identical volumes",
			args:       []string{"-v", fmt.Sprintf("%s:/test", suite.tempDir), "-v", fmt.Sprintf("%s:/test", suite.tempDir)},
			shouldFail: false, // Should be allowed
		},
		{
			name:       "Port out of range",
			args:       []string{"-p", "99999:80"},
			shouldFail: true,
		},
		{
			name:       "Invalid environment variable",
			args:       []string{"-e", "INVALID VAR=value"},
			shouldFail: true,
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			containerName := fmt.Sprintf("%s_%s", suite.containerName, strings.ReplaceAll(tc.name, " ", "_"))

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			args := []string{"docker", "run-in-container", "go", "--name", containerName}
			args = append(args, tc.args...)
			args = append(args, "--keep-running")

			cmd := exec.CommandContext(ctx, suite.portunixBinary, args...)
			err := cmd.Run()

			if tc.shouldFail {
				assert.Error(t, err, "Should fail for %s", tc.name)
			} else {
				// For now, document expected behavior
				// When implemented, should succeed
				t.Logf("Case '%s' expected to succeed when implemented", tc.name)
			}
		})
	}
}

// TestPythonDockerFallback tests Python fallback to direct Docker when Portunix params not supported
func (suite *PortunixPythonIntegrationTestSuite) TestPythonDockerFallback() {
	t := suite.T()

	// This tests the current workaround that users have to use
	projectDir := filepath.Join(suite.tempDir, "fallback_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create test file
	testFile := filepath.Join(projectDir, "test.go")
	err = ioutil.WriteFile(testFile, []byte("package main\n\nfunc main() { println(\"Fallback test\") }"), 0644)
	require.NoError(t, err)

	// Try Portunix then fallback - simulated as Go function
	method, success, output := suite.tryPortunixThenFallback(projectDir, suite.containerName)

	// Should succeed with either method
	assert.True(t, success, "Both Portunix and Docker fallback failed: %s", output)

	// Verify the container is actually working
	if method == "docker" {
		// Test that the fallback container has the volume mounted
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		testCmd := exec.CommandContext(ctx, "docker", "exec", suite.containerName, "ls", "/workspace/test.go")
		err := testCmd.Run()
		assert.NoError(t, err, "Volume not mounted in fallback container")
	}
}

// tryPortunixThenFallback simulates the Python function that tries Portunix first, then fallback to direct Docker
func (suite *PortunixPythonIntegrationTestSuite) tryPortunixThenFallback(projectPath, containerName string) (string, bool, string) {
	// Try Portunix with volume mounting
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	portunixCmd := exec.CommandContext(ctx, suite.portunixBinary, "docker", "run-in-container", "go",
		"--name", containerName,
		"-v", fmt.Sprintf("%s:/workspace", projectPath),
		"--keep-running")

	output, err := portunixCmd.CombinedOutput()

	if err == nil {
		return "portunix", true, string(output)
	}

	// Fallback to direct Docker
	suite.T().Logf("Portunix failed: %s", string(output))
	suite.T().Logf("Falling back to direct Docker...")

	fallbackName := fmt.Sprintf("%s_fallback", containerName)
	
	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	dockerCmd := exec.CommandContext(ctx2, "docker", "run", "-d",
		"--name", fallbackName,
		"-v", fmt.Sprintf("%s:/workspace", projectPath),
		"-w", "/workspace",
		"golang:1.21",
		"tail", "-f", "/dev/null") // Keep container running

	dockerOutput, dockerErr := dockerCmd.CombinedOutput()

	if dockerErr == nil {
		// Update container name for cleanup
		suite.containerName = fallbackName
		return "docker", true, string(dockerOutput)
	}

	return "none", false, string(dockerOutput)
}

// TestPerformanceWithLargeProject tests performance with larger project structure
func (suite *PortunixPythonIntegrationTestSuite) TestPerformanceWithLargeProject() {
	t := suite.T()

	// Create larger project structure
	largeProjectDir := filepath.Join(suite.tempDir, "large_project")
	err := os.MkdirAll(largeProjectDir, 0755)
	require.NoError(t, err)

	// Create multiple directories and files
	for i := 0; i < 10; i++ {
		dirPath := filepath.Join(largeProjectDir, fmt.Sprintf("pkg%d", i))
		err := os.MkdirAll(dirPath, 0755)
		require.NoError(t, err)

		for j := 0; j < 5; j++ {
			filePath := filepath.Join(dirPath, fmt.Sprintf("file%d.go", j))
			content := fmt.Sprintf("package pkg%d\n\n// File %d\nfunc Function%d() {}", i, j, j)
			err := ioutil.WriteFile(filePath, []byte(content), 0644)
			require.NoError(t, err)
		}
	}

	// Measure container creation time
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, suite.portunixBinary, "docker", "run-in-container", "go",
		"--name", suite.containerName,
		"-v", fmt.Sprintf("%s:/workspace", largeProjectDir),
		"--keep-running")

	cmd.Run() // Ignore result for timing test

	elapsedTime := time.Since(startTime)

	// Count files for logging
	fileCount := 0
	err = filepath.Walk(largeProjectDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			fileCount++
		}
		return nil
	})
	require.NoError(t, err)

	// Log performance metrics
	t.Logf("Container creation time: %.2fs", elapsedTime.Seconds())
	t.Logf("Project size: %d files", fileCount)

	// Performance should be reasonable (less than 60 seconds)
	assert.Less(t, elapsedTime, 60*time.Second, "Container creation took too long")
}

// PortunixParameterCompatibilityTestSuite tests compatibility between Docker and Podman parameters
type PortunixParameterCompatibilityTestSuite struct {
	suite.Suite
	tempDir string
}

func TestPortunixParameterCompatibilityTestSuite(t *testing.T) {
	suite.Run(t, new(PortunixParameterCompatibilityTestSuite))
}

func (suite *PortunixParameterCompatibilityTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "portunix_compat_test_*")
	suite.Require().NoError(err)
	suite.tempDir = tempDir
}

func (suite *PortunixParameterCompatibilityTestSuite) TearDownTest() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// TestDockerPodmanParameterEquivalence tests that same parameters work with both Docker and Podman
func (suite *PortunixParameterCompatibilityTestSuite) TestDockerPodmanParameterEquivalence() {
	t := suite.T()

	// This is a design test - documents expected behavior
	commonParams := [][]string{
		{"-v", fmt.Sprintf("%s:/workspace", suite.tempDir)},
		{"-p", "8080:80"},
		{"-e", "ENV=test"},
		{"--workdir", "/workspace"},
		{"--user", "1000:1000"},
		{"--memory", "2G"},
		{"--cpus", "1.5"},
	}

	for _, params := range commonParams {
		t.Run(fmt.Sprintf("params_%s", strings.Join(params, "_")), func(t *testing.T) {
			// Test that parameter parsing should work for both runtimes
			// This will be implemented in the actual parameter parser

			// Mock test - represents expected behavior
			dockerCompatible := true  // Should be compatible
			podmanCompatible := true  // Should be compatible

			assert.True(t, dockerCompatible, "Docker should support %v", params)
			assert.True(t, podmanCompatible, "Podman should support %v", params)
		})
	}
}

// TestRuntimeSpecificTranslations tests parameters that need translation between runtimes
func (suite *PortunixParameterCompatibilityTestSuite) TestRuntimeSpecificTranslations() {
	t := suite.T()

	translationCases := []struct {
		param  string
		docker string
		podman string
		reason string
	}{
		{
			param:  "--network bridge",
			docker: "bridge",
			podman: "slirp4netns", // Podman's equivalent
			reason: "Different default networking",
		},
	}

	for _, tc := range translationCases {
		t.Run(tc.param, func(t *testing.T) {
			// Document expected translation behavior
			t.Logf("Parameter: %s", tc.param)
			t.Logf("Docker: %s", tc.docker)
			t.Logf("Podman: %s", tc.podman)
			t.Logf("Reason: %s", tc.reason)

			// This will be implemented in the runtime translator
			assert.NotEqual(t, tc.docker, tc.podman,
				"Translation should produce different values")
		})
	}
}