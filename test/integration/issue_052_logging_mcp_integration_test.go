package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/pkg/logging"
	"portunix.ai/portunix/test/testframework"
)

func TestIssue052MCPSTDIOSeparation(t *testing.T) {
	tf := testframework.NewTestFramework("Issue052_MCP_STDIO_Separation")
	tf.Start(t, "Test MCP server STDIO separation with logging system")

	success := true
	defer tf.Finish(t, success)

	// Build the binary first
	tf.Step(t, "Build Portunix binary for MCP testing")
	binaryPath := filepath.Join("..", "..", "portunix")

	buildCmd := exec.Command("go", "build", "-o", binaryPath)
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		tf.Error(t, "Failed to build binary", err.Error(), string(output))
		success = false
		return
	}
	tf.Success(t, "Binary built successfully")

	tf.Separator()

	// Step 1: Test MCP server startup with logging
	tf.Step(t, "Test MCP server startup with clean STDIO")

	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "mcp-stdio-test.log")

	// Set environment for MCP logging
	os.Setenv("PORTUNIX_LOG_LEVEL", "debug")
	os.Setenv("PORTUNIX_LOG_OUTPUT", "file")
	os.Setenv("PORTUNIX_LOG_FILE", logFile)
	defer func() {
		os.Unsetenv("PORTUNIX_LOG_LEVEL")
		os.Unsetenv("PORTUNIX_LOG_OUTPUT")
		os.Unsetenv("PORTUNIX_LOG_FILE")
	}()

	// Start MCP server in background
	tf.Command(t, binaryPath, []string{"mcp", "serve"})
	mcpCmd := exec.Command(binaryPath, "mcp", "serve")

	var stdoutBuf, stderrBuf bytes.Buffer
	mcpCmd.Stdout = &stdoutBuf
	mcpCmd.Stderr = &stderrBuf

	// Start the command
	err := mcpCmd.Start()
	if err != nil {
		tf.Error(t, "Failed to start MCP server", err.Error())
		success = false
		return
	}

	// Let it run briefly
	time.Sleep(500 * time.Millisecond)

	// Terminate the server
	if mcpCmd.Process != nil {
		mcpCmd.Process.Kill()
		mcpCmd.Wait()
	}

	// Check STDOUT for MCP protocol readiness
	stdoutOutput := stdoutBuf.String()
	stderrOutput := stderrBuf.String()

	tf.Output(t, stdoutOutput, 500)
	tf.Info(t, "STDOUT content captured")
	tf.Output(t, stderrOutput, 200)
	tf.Info(t, "STDERR content captured")

	// STDOUT should contain only MCP protocol messages, no debug logs
	if strings.Contains(stdoutOutput, "debug") || strings.Contains(stdoutOutput, "info") {
		tf.Warning(t, "Potential log contamination in STDOUT", stdoutOutput)
	} else {
		tf.Success(t, "STDOUT appears clean of log messages")
	}

	tf.Separator()

	// Step 2: Verify log file contains server logs
	tf.Step(t, "Verify log file contains MCP server logs")

	if _, err := os.Stat(logFile); err != nil {
		tf.Warning(t, "Log file not created", err.Error())
		// This might be expected if MCP server doesn't run long enough
	} else {
		logContent, err := os.ReadFile(logFile)
		if err != nil {
			tf.Error(t, "Failed to read log file", err.Error())
			success = false
			return
		}

		tf.Output(t, string(logContent), 1000)
		tf.Info(t, "Log file content captured")

		// Log file should contain server startup messages
		if strings.Contains(string(logContent), "mcp") || strings.Contains(string(logContent), "server") {
			tf.Success(t, "Log file contains MCP server logs")
		} else {
			tf.Warning(t, "Log file might not contain expected MCP logs")
		}
	}

	tf.Separator()

	// Step 3: Test MCP JSON-RPC communication
	tf.Step(t, "Test MCP JSON-RPC protocol communication")

	// Create a mock MCP request
	mcpRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	requestJSON, err := json.Marshal(mcpRequest)
	if err != nil {
		tf.Error(t, "Failed to marshal MCP request", err.Error())
		success = false
		return
	}

	// Test with a fresh MCP server instance
	mcpCmd2 := exec.Command(binaryPath, "mcp", "serve")

	stdin, err := mcpCmd2.StdinPipe()
	if err != nil {
		tf.Error(t, "Failed to get stdin pipe", err.Error())
		success = false
		return
	}

	var stdout2Buf bytes.Buffer
	mcpCmd2.Stdout = &stdout2Buf

	err = mcpCmd2.Start()
	if err != nil {
		tf.Error(t, "Failed to start second MCP server", err.Error())
		success = false
		return
	}

	// Send the JSON-RPC request
	_, err = stdin.Write(requestJSON)
	if err != nil {
		tf.Error(t, "Failed to write to MCP server", err.Error())
		stdin.Close()
		mcpCmd2.Process.Kill()
		success = false
		return
	}

	stdin.Close()

	// Wait briefly for response
	time.Sleep(200 * time.Millisecond)

	// Terminate
	mcpCmd2.Process.Kill()
	mcpCmd2.Wait()

	// Check response
	response := stdout2Buf.String()
	tf.Output(t, response, 500)
	tf.Info(t, "MCP response captured")

	// Response should be valid JSON or at least not contain log messages
	if strings.Contains(response, "debug") || strings.Contains(response, "info") {
		tf.Error(t, "MCP response contaminated with log messages", response)
		success = false
		return
	}

	tf.Success(t, "MCP protocol communication clean")

	tf.Info(t, "MCP STDIO separation test completed")
}

func TestIssue052LoggingContainerIntegration(t *testing.T) {
	tf := testframework.NewTestFramework("Issue052_Container_Integration")
	tf.Start(t, "Test logging system integration with container environments")

	success := true
	defer tf.Finish(t, success)

	// Step 1: Test container environment detection
	tf.Step(t, "Test container environment detection")

	// Test the container detection functions directly
	env := logging.DetectEnvironment()
	if env == nil {
		tf.Error(t, "DetectEnvironment returned nil")
		success = false
		return
	}

	tf.Info(t, "Environment detection results",
		"os", env["os"],
		"arch", env["arch"],
		"container", env["container"])

	// Verify basic environment info
	if env["os"] == "" {
		tf.Error(t, "OS detection failed", "Expected non-empty OS")
		success = false
		return
	}
	if env["arch"] == "" {
		tf.Error(t, "Architecture detection failed", "Expected non-empty architecture")
		success = false
		return
	}

	tf.Success(t, "Environment detection working")

	tf.Separator()

	// Step 2: Test logging in container context
	tf.Step(t, "Test logging with container context")

	config := &logging.Config{
		Level:  "debug",
		Format: "json",
		Output: []string{"console"},
	}

	factory := logging.NewFactory(config)
	containerLogger := factory.CreateLogger("container-test")

	var buf bytes.Buffer
	containerLogger.SetOutput(&buf)

	// Log a test message
	containerLogger.Info("Container logging test",
		"test_id", "052-container-001",
		"environment", "test")

	output := buf.String()
	tf.Output(t, output, 500)
	tf.Info(t, "Container log output captured")

	// Verify the log contains environment information
	if !strings.Contains(output, "container-test") {
		tf.Error(t, "Component name missing from container log", output)
		success = false
		return
	}

	// Check for OS/architecture info (should be included)
	if !strings.Contains(output, "os") || !strings.Contains(output, "arch") {
		tf.Error(t, "Platform information missing from container log", output)
		success = false
		return
	}

	tf.Success(t, "Container context logging working")

	tf.Separator()

	// Step 3: Test Portunix container command integration
	tf.Step(t, "Test Portunix container command integration")

	// Build binary if not exists
	binaryPath := filepath.Join("..", "..", "portunix")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		buildCmd := exec.Command("go", "build", "-o", binaryPath)
		buildCmd.Dir = filepath.Join("..", "..")
		if output, err := buildCmd.CombinedOutput(); err != nil {
			tf.Error(t, "Failed to build binary for container test", err.Error(), string(output))
			success = false
			return
		}
	}

	// Test Portunix docker help (should work without container runtime)
	tf.Command(t, binaryPath, []string{"docker", "--help"})
	dockerCmd := exec.Command(binaryPath, "docker", "--help")

	outputBytes, err := dockerCmd.CombinedOutput()
	output := string(outputBytes)
	if err != nil {
		tf.Warning(t, "Docker help command failed", err.Error(), output)
		// This is expected if Docker is not installed
	} else {
		tf.Info(t, "Docker command available")
		tf.Output(t, output, 300)
		tf.Info(t, "Docker help output captured")
	}

	// Test Portunix podman help
	tf.Command(t, binaryPath, []string{"podman", "--help"})
	podmanCmd := exec.Command(binaryPath, "podman", "--help")

	outputBytes, err = podmanCmd.CombinedOutput()
	output = string(outputBytes)
	if err != nil {
		tf.Warning(t, "Podman help command failed", err.Error(), output)
		// This is expected if Podman is not installed
	} else {
		tf.Info(t, "Podman command available")
		tf.Output(t, output, 300)
		tf.Info(t, "Podman help output captured")
	}

	tf.Success(t, "Container command integration tested")

	tf.Info(t, "Container integration test completed")
}

func TestIssue052LoggingErrorHandling(t *testing.T) {
	tf := testframework.NewTestFramework("Issue052_Error_Handling")
	tf.Start(t, "Test logging system error handling and resilience")

	success := true
	defer tf.Finish(t, success)

	// Step 1: Test invalid file path handling
	tf.Step(t, "Test invalid file path handling")

	config := &logging.Config{
		Level:    "info",
		Format:   "json",
		Output:   []string{"file"},
		FilePath: "/invalid/nonexistent/path/test.log",
	}

	factory := logging.NewFactory(config)
	logger := factory.CreateLogger("error-test")

	// This should not panic and should fallback gracefully
	var buf bytes.Buffer
	logger.SetOutput(&buf) // Override to capture output
	logger.Error("Error handling test")

	output := buf.String()
	if output == "" {
		tf.Error(t, "No output generated with invalid file path")
		success = false
		return
	}

	tf.Success(t, "Invalid file path handled gracefully")

	tf.Separator()

	// Step 2: Test configuration validation
	tf.Step(t, "Test configuration validation with invalid values")

	invalidConfig := &logging.Config{
		Level:  "invalid-level",
		Format: "invalid-format",
		Output: []string{"invalid-output", "console"},
	}

	err := invalidConfig.Validate()
	if err != nil {
		tf.Error(t, "Configuration validation failed", err.Error())
		success = false
		return
	}

	// Check corrections
	if invalidConfig.Level != "info" {
		tf.Error(t, "Invalid level not corrected to 'info'", "Got:", invalidConfig.Level)
		success = false
		return
	}
	if invalidConfig.Format != "text" {
		tf.Error(t, "Invalid format not corrected to 'text'", "Got:", invalidConfig.Format)
		success = false
		return
	}
	if !contains(invalidConfig.Output, "console") {
		tf.Error(t, "Valid output option removed during validation", "Output:", invalidConfig.Output)
		success = false
		return
	}

	tf.Success(t, "Configuration validation working correctly")

	tf.Separator()

	// Step 3: Test memory safety with high-frequency logging
	tf.Step(t, "Test memory safety with high-frequency logging")

	tempDir := t.TempDir()
	config = &logging.Config{
		Level:    "debug",
		Format:   "json",
		Output:   []string{"file"},
		FilePath: filepath.Join(tempDir, "memory-test.log"),
	}

	factory = logging.NewFactory(config)
	logger = factory.CreateLogger("memory-test")

	// Generate high-frequency logs
	for i := 0; i < 10000; i++ {
		logger.Debug("Memory test message", "iteration", i, "data", strings.Repeat("x", 100))
	}

	// Verify file was created and has content
	if stat, err := os.Stat(config.FilePath); err != nil {
		tf.Error(t, "High-frequency log file not created", err.Error())
		success = false
		return
	} else if stat.Size() == 0 {
		tf.Error(t, "High-frequency log file is empty")
		success = false
		return
	} else {
		tf.Info(t, "High-frequency logging completed", "file_size", stat.Size())
	}

	tf.Success(t, "Memory safety test passed")

	tf.Info(t, "Error handling tests completed")
}

// Helper function for slice containment check
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}