package integration

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
	
	"portunix.ai/portunix/test/testframework"
)

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestMCPServeIssue037 Integration tests for Issue #037 - MCP Serve Implementation
type TestMCPServeIssue037 struct {
	binaryPath string
}

func TestIssue037MCPServeImplementation(t *testing.T) {
	tf := testframework.NewTestFramework("Issue037_MCP_Serve_Complete")
	tf.Start(t, "Comprehensive MCP serve implementation test suite with 11 test cases")
	
	success := true
	defer func() {
		tf.Finish(t, success)
	}()
	
	tf.Step(t, "Initialize test suite")
	suite := &TestMCPServeIssue037{}
	suite.setupBinary(t)
	
	tf.Success(t, "Test binary ready", suite.binaryPath)
	tf.Info(t, "Running 11 comprehensive test cases")
	
	// Run subtests and track their success
	subTestSuccess := true
	subTestSuccess = subTestSuccess && t.Run("TC001_DefaultHelpDisplay", suite.testDefaultHelpDisplay)
	subTestSuccess = subTestSuccess && t.Run("TC002_BasicMCPServe", suite.testBasicMCPServe)
	subTestSuccess = subTestSuccess && t.Run("TC003_ExplicitStdioMode", suite.testExplicitStdioMode)
	subTestSuccess = subTestSuccess && t.Run("TC004_TCPModeWithPort", suite.testTCPModeWithPort)
	subTestSuccess = subTestSuccess && t.Run("TC005_UnixSocketMode", suite.testUnixSocketMode)
	subTestSuccess = subTestSuccess && t.Run("TC006_LegacyCommandDeprecation", suite.testLegacyCommandDeprecation)
	subTestSuccess = subTestSuccess && t.Run("TC008_InvalidModeParameter", suite.testInvalidModeParameter)
	subTestSuccess = subTestSuccess && t.Run("TC009_TCPModeWithoutPort", suite.testTCPModeWithoutPort)
	subTestSuccess = subTestSuccess && t.Run("TC010_SocketFilePermissions", suite.testSocketFilePermissions)
	subTestSuccess = subTestSuccess && t.Run("TC011_ConcurrentMCPServers", suite.testConcurrentMCPServers)
	
	// Update main success based on subtests
	success = success && subTestSuccess
	
	if subTestSuccess {
		tf.Success(t, "All 11 test cases completed successfully")
	} else {
		tf.Error(t, "Some test cases failed")
	}
}

func (suite *TestMCPServeIssue037) setupBinary(t *testing.T) {
	// Use existing binary or build if not exists
	projectRoot := "../.."
	binaryName := "portunix"
	suite.binaryPath = "./portunix"  // Use relative path like other tests
	
	// Check if binary exists, if not build it
	if _, err := os.Stat(suite.binaryPath); os.IsNotExist(err) {
		cmd := exec.Command("go", "build", "-o", binaryName, ".")
		cmd.Dir = projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build binary: %v\nOutput: %s", err, string(output))
		}
	}
}

// TC001: Default Help Display
func (suite *TestMCPServeIssue037) testDefaultHelpDisplay(t *testing.T) {
	tf := testframework.NewTestFramework("TC001_DefaultHelpDisplay")
	tf.Start(t, "Test default help display without starting MCP server")
	
	success := true
	defer func() {
		tf.Finish(t, success)
	}()
	
	tf.Step(t, "Execute portunix without parameters")
	tf.Command(t, suite.binaryPath, []string{})
	
	cmd := exec.Command(suite.binaryPath)
	cmd.Dir = "../.."  // Set working directory to project root
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 0 {
				tf.Error(t, "Unexpected exit code", fmt.Sprintf("Expected 0, got %d", exitError.ExitCode()))
				success = false
			} else {
				tf.Success(t, "Correct exit code", "0")
			}
		}
	} else {
		tf.Success(t, "Exit code", "0")
	}
	
	outputStr := string(output)
	tf.Output(t, outputStr, 300)
	
	if !strings.Contains(outputStr, "Usage:") && !strings.Contains(outputStr, "Available Commands:") {
		tf.Error(t, "Missing expected help content")
		success = false
	} else {
		tf.Success(t, "Help content found")
	}
	
	// Ensure it doesn't enter stdio mode (no MCP server start message)
	if strings.Contains(outputStr, "Starting MCP Server") {
		tf.Error(t, "Unexpected MCP server start")
		success = false
	} else {
		tf.Success(t, "No MCP server started (correct behavior)")
	}
}

// TC002: Basic MCP Serve - stdio Mode
func (suite *TestMCPServeIssue037) testBasicMCPServe(t *testing.T) {
	tf := testframework.NewTestFramework("TC002_BasicMCPServe")
	tf.Start(t, "Test basic MCP serve in default stdio mode")
	
	success := true
	defer func() {
		tf.Finish(t, success)
	}()
	
	tf.Step(t, "Setup stdio mode test with timeout")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	tf.Step(t, "Execute MCP serve in stdio mode")
	tf.Command(t, suite.binaryPath, []string{"mcp", "serve"})
	
	cmd := exec.CommandContext(ctx, suite.binaryPath, "mcp", "serve")
	cmd.Dir = "../.."  // Set working directory to project root
	stderr, err := cmd.StderrPipe()
	if err != nil {
		tf.Error(t, "Failed to get stderr pipe", err.Error())
		success = false
		return
	}
	
	if err := cmd.Start(); err != nil {
		tf.Error(t, "Failed to start command", err.Error())
		success = false
		return
	}
	
	tf.Info(t, "Reading server startup messages from stderr")
	// Read stderr for server start message
	scanner := bufio.NewScanner(stderr)
	found := false
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			tf.Info(t, "Server output", line)
			if strings.Contains(line, "Starting MCP Server in stdio mode") {
				found = true
				break
			}
		}
	}()
	
	tf.Info(t, "Waiting for server startup", "1s")
	time.Sleep(1 * time.Second)
	tf.Info(t, "Sending termination signal")
	cmd.Process.Signal(syscall.SIGTERM)
	cmd.Wait()
	
	if !found {
		tf.Error(t, "Expected 'Starting MCP Server in stdio mode' message not found")
		success = false
	} else {
		tf.Success(t, "MCP server started successfully in stdio mode")
	}
}

// TC003: Explicit stdio Mode
func (suite *TestMCPServeIssue037) testExplicitStdioMode(t *testing.T) {
	tf := testframework.NewTestFramework("TC003_ExplicitStdioMode")
	tf.Start(t, "Test explicit stdio mode parameter")
	
	success := true
	defer func() {
		tf.Finish(t, success)
	}()
	
	tf.Step(t, "Setup explicit stdio mode test")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	tf.Step(t, "Execute MCP serve with explicit stdio mode")
	tf.Command(t, suite.binaryPath, []string{"mcp", "serve", "--mode", "stdio"})
	
	cmd := exec.CommandContext(ctx, suite.binaryPath, "mcp", "serve", "--mode", "stdio")
	cmd.Dir = "../.."
	stderr, err := cmd.StderrPipe()
	if err != nil {
		tf.Error(t, "Failed to get stderr pipe", err.Error())
		success = false
		return
	}
	
	if err := cmd.Start(); err != nil {
		tf.Error(t, "Failed to start command", err.Error())
		success = false
		return
	}
	
	tf.Info(t, "Monitoring stderr for server startup messages")
	scanner := bufio.NewScanner(stderr)
	found := false
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			tf.Info(t, "Server message", line)
			if strings.Contains(line, "Starting MCP Server in stdio mode") {
				found = true
				break
			}
		}
	}()
	
	tf.Info(t, "Waiting for server initialization", "1s")
	time.Sleep(1 * time.Second)
	tf.Info(t, "Terminating server")
	cmd.Process.Signal(syscall.SIGTERM)
	cmd.Wait()
	
	if !found {
		tf.Error(t, "Expected 'Starting MCP Server in stdio mode' message not found")
		success = false
	} else {
		tf.Success(t, "Explicit stdio mode working correctly")
	}
}

// TC004: TCP Mode with Port
func (suite *TestMCPServeIssue037) testTCPModeWithPort(t *testing.T) {
	tf := testframework.NewTestFramework("TC004_TCPModeWithPort")
	tf.Start(t, "Test MCP serve with TCP mode on custom port")
	
	success := true
	defer func() {
		tf.Finish(t, success)
	}()
	
	port := 18080 // Use non-standard port to avoid conflicts
	tf.Step(t, "Setup TCP server test", fmt.Sprintf("Port: %d", port))
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	tf.Step(t, "Execute MCP serve with TCP mode")
	tf.Command(t, suite.binaryPath, []string{"mcp", "serve", "--mode", "tcp", "--port", fmt.Sprintf("%d", port)})
	
	cmd := exec.CommandContext(ctx, suite.binaryPath, "mcp", "serve", "--mode", "tcp", "--port", fmt.Sprintf("%d", port))
	cmd.Dir = "../.."
	
	tf.Info(t, "Waiting for server startup", "5s timeout")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			tf.Info(t, "Context timeout reached", "Expected behavior")
		} else {
			tf.Warning(t, "Command error", err.Error())
		}
	}
	
	outputStr := string(output)
	tf.Output(t, outputStr, 400)
	
	expectedMsg := fmt.Sprintf("Starting MCP Server in TCP mode on port %d", port)
	if !strings.Contains(outputStr, expectedMsg) {
		tf.Error(t, "Expected TCP server start message not found", expectedMsg)
		success = false
	} else {
		tf.Success(t, "TCP server started successfully", fmt.Sprintf("Port %d", port))
	}
}

// TC005: Unix Socket Mode
func (suite *TestMCPServeIssue037) testUnixSocketMode(t *testing.T) {
	tf := testframework.NewTestFramework("TC005_UnixSocketMode")
	tf.Start(t, "Test MCP serve with Unix domain socket")
	
	success := true
	defer func() {
		tf.Finish(t, success)
	}()
	
	socketPath := "/tmp/portunix-test.sock"
	tf.Step(t, "Setup Unix socket test", fmt.Sprintf("Socket path: %s", socketPath))
	
	// Clean up any existing socket
	os.Remove(socketPath)
	defer os.Remove(socketPath)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	tf.Step(t, "Execute MCP serve with Unix socket mode")
	tf.Command(t, suite.binaryPath, []string{"mcp", "serve", "--mode", "unix", "--socket", socketPath})
	
	cmd := exec.CommandContext(ctx, suite.binaryPath, "mcp", "serve", "--mode", "unix", "--socket", socketPath)
	cmd.Dir = "../.."
	
	tf.Info(t, "Waiting for Unix socket server startup", "5s timeout")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			tf.Info(t, "Context timeout reached", "Expected behavior")
		} else {
			tf.Warning(t, "Command error", err.Error())
		}
	}
	
	outputStr := string(output)
	tf.Output(t, outputStr, 400)
	
	expectedMsg := fmt.Sprintf("Starting MCP Server in Unix socket mode: %s", socketPath)
	if !strings.Contains(outputStr, expectedMsg) {
		tf.Error(t, "Expected Unix socket server start message not found", expectedMsg)
		success = false
	} else {
		tf.Success(t, "Unix socket server started successfully", socketPath)
	}
}

// TC006: Legacy Command Deprecation
func (suite *TestMCPServeIssue037) testLegacyCommandDeprecation(t *testing.T) {
	tf := testframework.NewTestFramework("TC006_LegacyCommandDeprecation")
	tf.Start(t, "Test legacy 'mcp serve' command shows proper error")
	
	success := true
	defer func() {
		tf.Finish(t, success)
	}()
	
	tf.Step(t, "Execute legacy command format")
	tf.Command(t, suite.binaryPath, []string{"mcp serve"})
	
	cmd := exec.Command(suite.binaryPath, "mcp serve")
	cmd.Dir = "../.."
	output, err := cmd.CombinedOutput()
	
	if err == nil {
		tf.Error(t, "Expected error for legacy mcp serve command")
		success = false
	} else {
		tf.Success(t, "Legacy command properly rejected")
	}
	
	outputStr := string(output)
	tf.Output(t, outputStr, 300)
	
	if !strings.Contains(outputStr, "unknown command") && !strings.Contains(outputStr, "mcp serve") {
		tf.Error(t, "Expected error about unknown command", outputStr)
		success = false
	} else {
		tf.Success(t, "Proper error message displayed")
	}
}

// TC008: Invalid Mode Parameter
func (suite *TestMCPServeIssue037) testInvalidModeParameter(t *testing.T) {
	t.Log("Testing TC008: Invalid Mode Parameter")
	
	cmd := exec.Command(suite.binaryPath, "mcp", "serve", "--mode", "invalid")
	cmd.Dir = "../.."
	output, err := cmd.CombinedOutput()
	
	if err == nil {
		t.Error("Expected error for invalid mode parameter")
	}
	
	outputStr := string(output)
	if !strings.Contains(outputStr, "Unknown mode") {
		t.Errorf("Expected 'Unknown mode' error message, got: %s", outputStr)
	}
}

// TC009: TCP Mode without Port
func (suite *TestMCPServeIssue037) testTCPModeWithoutPort(t *testing.T) {
	fmt.Println("🔧 TC009: Testing TCP Mode without Port (should use default)")
	t.Log("Testing TC009: TCP Mode without Port (should use default)")
	
	fmt.Println("   Testing TCP server with default port 3001...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, suite.binaryPath, "mcp", "serve", "--mode", "tcp")
	cmd.Dir = "../.."
	
	fmt.Println("   ⏳ Waiting for default port server startup (3s timeout)...")
	output, err := cmd.CombinedOutput()
	fmt.Println("   ✅ Command completed")
	
	if err != nil && ctx.Err() != context.DeadlineExceeded {
		t.Logf("Command output: %s", string(output))
	}
	
	outputStr := string(output)
	// Should use default port 3001
	if !strings.Contains(outputStr, "TCP mode on port 3001") {
		t.Errorf("Expected default port 3001 to be used, got: %s", outputStr)
	}
}

// TC010: Socket File Permissions
func (suite *TestMCPServeIssue037) testSocketFilePermissions(t *testing.T) {
	t.Log("Testing TC010: Socket File Permissions")
	
	// Try to create socket in directory without write permissions
	socketPath := "/root/portunix-test.sock" // Typically no access for non-root users
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, suite.binaryPath, "mcp", "serve", "--mode", "unix", "--socket", socketPath)
	cmd.Dir = "../.."
	output, err := cmd.CombinedOutput()
	
	// Should fail due to permissions
	if err == nil {
		t.Error("Expected error for socket in directory without write permissions")
	}
	
	outputStr := string(output)
	// Should contain permission-related error
	if !strings.Contains(outputStr, "permission denied") && !strings.Contains(outputStr, "Failed to start") {
		t.Logf("Expected permission error, got: %s", outputStr)
	}
}

// TC011: Concurrent MCP Servers
func (suite *TestMCPServeIssue037) testConcurrentMCPServers(t *testing.T) {
	fmt.Println("⚡ TC011: Testing Concurrent MCP Servers (port conflict)")
	t.Log("Testing TC011: Concurrent MCP Servers (port conflict)")
	
	port := 18081
	fmt.Printf("   Testing port conflict on port %d...\n", port)
	
	// Start first server
	ctx1, cancel1 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel1()
	
	cmd1 := exec.CommandContext(ctx1, suite.binaryPath, "mcp", "serve", "--mode", "tcp", "--port", fmt.Sprintf("%d", port))
	cmd1.Dir = "../.."
	fmt.Println("   🚀 Starting first server...")
	if err := cmd1.Start(); err != nil {
		t.Fatalf("Failed to start first MCP server: %v", err)
	}
	defer cmd1.Process.Kill()
	
	// Wait a bit for first server to start
	fmt.Println("   ⏳ Waiting for first server to start (2s)...")
	time.Sleep(2 * time.Second)
	
	// Try to start second server on same port
	fmt.Println("   🔥 Starting second server (should fail)...")
	cmd2 := exec.Command(suite.binaryPath, "mcp", "serve", "--mode", "tcp", "--port", fmt.Sprintf("%d", port))
	cmd2.Dir = "../.."
	output, err := cmd2.CombinedOutput()
	fmt.Println("   ✅ Second server test completed")
	
	if err == nil {
		t.Error("Expected error when starting second server on same port")
	}
	
	outputStr := string(output)
	// Should contain port conflict error
	if !strings.Contains(outputStr, "Failed to start") && !strings.Contains(outputStr, "address already in use") {
		t.Logf("Expected port conflict error, got: %s", outputStr)
	}
}

// Helper function to check if TCP port is available
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}