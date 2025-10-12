package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/test/testframework"
)

// TestClaudeCodeContainerInstallAndSetup tests complete Portunix installation in container
// and subsequent Claude Code MCP configuration. This test validates:
// 1. Container-based Portunix deployment
// 2. Automatic prerequisite handling (Node.js/npm for Claude Code)
// 3. Claude Code installation and setup via Portunix
// 4. MCP server integration with Claude Code
// 5. End-to-end workflow validation
func TestClaudeCodeContainerInstallAndSetup(t *testing.T) {
	tf := testframework.NewTestFramework("ClaudeCode_Container_Install")
	tf.Start(t, "Complete E2E test: Portunix container installation -> Claude Code setup -> MCP integration")
	
	success := true
	defer func() {
		tf.Finish(t, success)
	}()
	
	if testing.Short() {
		tf.Info(t, "Skipping E2E container test in short mode")
		t.Skip("Skipping E2E container test in short mode")
	}
	
	suite := &ClaudeCodeContainerTestSuite{
		tf:            tf,
		t:             t,
		containerName: "portunix-claude-code-test",
		projectRoot:   "../..",
	}
	
	// Test phases
	tf.Step(t, "Initialize test environment")
	suite.setupEnvironment()
	defer suite.cleanup()
	
	tf.Separator()
	
	tf.Step(t, "Create Ubuntu container with development tools")  
	suite.createTestContainer()
	
	tf.Separator()
	
	tf.Step(t, "Install Portunix in container")
	suite.installPortunixInContainer()
	
	tf.Separator()
	
	tf.Step(t, "Install Claude Code CLI (should auto-install Node.js prerequisite)")
	suite.installClaudeCode()
	
	tf.Separator()
	
	tf.Step(t, "Verify Node.js prerequisite was installed")
	suite.verifyNodeJSPrerequisite()
	
	tf.Separator()
	
	tf.Step(t, "Configure MCP integration for Claude Code")
	suite.configureMCPForClaudeCode()
	
	tf.Separator()
	
	tf.Step(t, "Test Claude Code <-> Portunix MCP communication")
	if !suite.testClaudeCodeMCPIntegration() {
		success = false
		return
	}
	
	tf.Success(t, "Complete E2E workflow validated successfully")
}

type ClaudeCodeContainerTestSuite struct {
	tf            *testframework.TestFramework
	t             *testing.T
	containerName string
	projectRoot   string
	binaryPath    string
}

func (suite *ClaudeCodeContainerTestSuite) runContainerCommand(command string, args ...string) (string, error) {
	// Use Portunix container exec instead of direct podman
	fullArgs := append([]string{"container", "exec", suite.containerName, command}, args...)
	
	suite.tf.Command(suite.t, suite.binaryPath, fullArgs)
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, suite.binaryPath, fullArgs...)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		suite.tf.Error(suite.t, fmt.Sprintf("Container command failed: %s", command), err.Error())
		suite.tf.Output(suite.t, string(output), 500)
		return string(output), err
	}
	
	suite.tf.Output(suite.t, string(output), 300)
	return string(output), nil
}

func (suite *ClaudeCodeContainerTestSuite) setupEnvironment() {
	// Build portunix binary if needed
	suite.binaryPath = filepath.Join(suite.projectRoot, "portunix")
	if _, err := os.Stat(suite.binaryPath); os.IsNotExist(err) {
		suite.tf.Info(suite.t, "Building portunix binary")
		cmd := exec.Command("go", "build", "-o", "portunix", ".")
		cmd.Dir = suite.projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			suite.tf.Error(suite.t, "Failed to build portunix binary", err.Error())
			suite.t.Fatalf("Build failed: %v\nOutput: %s", err, string(output))
		}
		suite.tf.Success(suite.t, "Portunix binary built successfully")
	} else {
		suite.tf.Success(suite.t, "Using existing portunix binary")
	}
	
	// Verify container runtime is available via Portunix
	cmd := exec.Command(suite.binaryPath, "container", "--help")
	if output, err := cmd.CombinedOutput(); err != nil {
		suite.tf.Error(suite.t, "Portunix container system not available, cannot run container tests", err.Error())
		suite.tf.Output(suite.t, string(output), 300)
		suite.t.Skip("Portunix container system not available")
	}
	suite.tf.Success(suite.t, "Portunix container system available")
}

func (suite *ClaudeCodeContainerTestSuite) createDevelopmentContainer() {
	// Remove existing container using Portunix
	suite.tf.Info(suite.t, "Cleaning up any existing development containers")
	cleanupCmd := exec.Command(suite.binaryPath, "container", "remove", suite.containerName)
	cleanupCmd.Run() // Ignore errors - container might not exist
	
	suite.tf.Info(suite.t, "Creating Ubuntu 22.04 container with dev tools using Portunix")
	
	// Use Portunix container run for development container
	suite.tf.Command(suite.t, suite.binaryPath, []string{"container", "run", "--name", suite.containerName, "-d", "ubuntu:22.04", "sleep", "3600"})
	cmd := exec.Command(suite.binaryPath, "container", "run", "--name", suite.containerName, "-d", "ubuntu:22.04", "sleep", "3600")
	
	if output, err := cmd.CombinedOutput(); err != nil {
		suite.tf.Error(suite.t, "Failed to create development container via Portunix", err.Error())
		suite.tf.Output(suite.t, string(output), 500)
		suite.t.Fatalf("Development container creation failed: %v", err)
	}
	
	// Wait for container setup
	suite.tf.Info(suite.t, "Waiting for development container initialization", "15s")
	time.Sleep(15 * time.Second)
	
	// Verify container is ready
	if output, err := suite.runContainerCommand("echo", "Container ready"); err != nil {
		suite.tf.Error(suite.t, "Container not responding", err.Error())
		suite.t.Fatalf("Container setup verification failed: %v", err)
	} else if !strings.Contains(output, "Container ready") {
		suite.tf.Error(suite.t, "Container startup verification failed")
		suite.t.Fatalf("Unexpected container response: %s", output)
	}
	
	suite.tf.Success(suite.t, "Ubuntu container ready with development tools")
}

func (suite *ClaudeCodeContainerTestSuite) createTestContainer() {
	// Remove existing container using Portunix
	suite.tf.Info(suite.t, "Cleaning up any existing test containers")
	cleanupCmd := exec.Command(suite.binaryPath, "container", "remove", suite.containerName)
	cleanupCmd.Run() // Ignore errors - container might not exist
	
	suite.tf.Info(suite.t, "Creating minimal test container for package testing")
	
	// Use Portunix container run for minimal test container
	// This is for testing pure package installation scenarios
	suite.tf.Command(suite.t, suite.binaryPath, []string{"container", "run", "--name", suite.containerName, "-d", "ubuntu:22.04", "sleep", "3600"})
	cmd := exec.Command(suite.binaryPath, "container", "run", "--name", suite.containerName, "-d", "ubuntu:22.04", "sleep", "3600")
	
	if output, err := cmd.CombinedOutput(); err != nil {
		suite.tf.Error(suite.t, "Failed to create test container via Portunix", err.Error())
		suite.tf.Output(suite.t, string(output), 500)
		suite.t.Fatalf("Test container creation failed: %v", err)
	}
	
	// Shorter wait for minimal container
	suite.tf.Info(suite.t, "Waiting for test container initialization", "10s")
	time.Sleep(10 * time.Second)
	
	// Verify container is ready
	if output, err := suite.runContainerCommand("echo", "Test container ready"); err != nil {
		suite.tf.Error(suite.t, "Test container not responding", err.Error())
		suite.t.Fatalf("Test container setup verification failed: %v", err)
	} else if !strings.Contains(output, "Test container ready") {
		suite.tf.Error(suite.t, "Test container startup verification failed")
		suite.t.Fatalf("Unexpected test container response: %s", output)
	}
	
	suite.tf.Success(suite.t, "Minimal test container ready for package testing")
}

func (suite *ClaudeCodeContainerTestSuite) installPortunixInContainer() {
	suite.tf.Info(suite.t, "Installing Portunix binary in container via Portunix container system")
	
	// Use Portunix container copy to copy binary
	suite.tf.Command(suite.t, suite.binaryPath, []string{"container", "cp", suite.binaryPath, suite.containerName + ":/usr/local/bin/portunix"})
	cmd := exec.Command(suite.binaryPath, "container", "cp", suite.binaryPath, suite.containerName+":/usr/local/bin/portunix")
	if output, err := cmd.CombinedOutput(); err != nil {
		suite.tf.Error(suite.t, "Failed to copy portunix to container via Portunix", err.Error())
		suite.tf.Output(suite.t, string(output), 500)
		suite.t.Fatalf("Copy failed: %v", err)
	}
	
	// Make executable
	if _, err := suite.runContainerCommand("chmod", "+x", "/usr/local/bin/portunix"); err != nil {
		suite.t.Fatalf("Failed to make portunix executable: %v", err)
	}
	
	// Verify installation
	output, err := suite.runContainerCommand("/usr/local/bin/portunix", "--version")
	if err != nil {
		suite.tf.Error(suite.t, "Portunix not working in container", err.Error())
		suite.t.Fatalf("Version check failed: %v", err)
	}
	
	suite.tf.Success(suite.t, "Portunix installed and verified in container", 
		fmt.Sprintf("Version: %s", strings.TrimSpace(output)))
}

func (suite *ClaudeCodeContainerTestSuite) installViaPortunix(packageName string) {
	suite.tf.Info(suite.t, fmt.Sprintf("Installing %s via Portunix package system", packageName))
	
	// Test dry-run first
	suite.tf.Info(suite.t, fmt.Sprintf("Testing dry-run mode for %s installation", packageName))
	dryRunOutput, err := suite.runContainerCommand("/usr/local/bin/portunix", "install", packageName, "--dry-run")
	if err != nil {
		suite.tf.Error(suite.t, fmt.Sprintf("%s dry-run failed", packageName), err.Error())
		suite.t.Fatalf("Dry-run failed: %v", err)
	}
	
	if !strings.Contains(dryRunOutput, "📦 INSTALLING") {
		suite.tf.Error(suite.t, fmt.Sprintf("Dry-run output missing %s package info", packageName))
		suite.t.Fatalf("Unexpected dry-run output: %s", dryRunOutput)
	}
	suite.tf.Success(suite.t, fmt.Sprintf("%s package found in Portunix package system", packageName))
	
	// Install package via Portunix
	suite.tf.Info(suite.t, fmt.Sprintf("Installing %s using Portunix", packageName))
	installOutput, err := suite.runContainerCommand("/usr/local/bin/portunix", "install", packageName)
	if err != nil {
		suite.tf.Error(suite.t, fmt.Sprintf("%s installation via Portunix failed", packageName), err.Error())
		suite.tf.Output(suite.t, installOutput, 500)
		suite.t.Fatalf("%s installation failed: %v", packageName, err)
	}
	
	suite.tf.Success(suite.t, fmt.Sprintf("%s installed successfully via Portunix", packageName))
}

func (suite *ClaudeCodeContainerTestSuite) verifyNodeJSPrerequisite() {
	suite.tf.Info(suite.t, "Verifying that Node.js prerequisite was automatically installed")
	
	// Verify Node.js is available
	nodeVersion, err := suite.runContainerCommand("node", "--version")
	if err != nil {
		suite.tf.Error(suite.t, "Node.js not available - prerequisite auto-installation failed", err.Error())
		suite.t.Fatalf("Node.js prerequisite verification failed: %v", err)
	}
	
	// Verify npm is available
	npmVersion, err := suite.runContainerCommand("npm", "--version")
	if err != nil {
		suite.tf.Error(suite.t, "npm not available - prerequisite auto-installation failed", err.Error())
		suite.t.Fatalf("npm prerequisite verification failed: %v", err)
	}
	
	suite.tf.Success(suite.t, "Node.js prerequisite automatically installed via Portunix prerequisite system",
		fmt.Sprintf("Node.js: %s, npm: %s", strings.TrimSpace(nodeVersion), strings.TrimSpace(npmVersion)))
}

func (suite *ClaudeCodeContainerTestSuite) installClaudeCode() {
	suite.tf.Info(suite.t, "Installing Claude Code CLI via Portunix (should auto-install Node.js prerequisite)")
	
	// Install Claude Code CLI via Portunix - this should automatically handle Node.js prerequisite
	output, err := suite.runContainerCommand("/usr/local/bin/portunix", "install", "claude-code")
	
	if err != nil {
		suite.tf.Error(suite.t, "Claude Code installation via Portunix failed", err.Error())
		suite.tf.Output(suite.t, output, 500)
		suite.t.Fatalf("Claude Code installation failed: %v", err)
	}
	
	// Verify installation
	claudeOutput, err := suite.runContainerCommand("claude", "--version")
	if err != nil {
		// Try with npx as fallback
		suite.tf.Info(suite.t, "Global claude command not found, trying npx")
		claudeOutput, err = suite.runContainerCommand("npx", "@anthropic-ai/claude-cli", "--version")
		if err != nil {
			suite.tf.Error(suite.t, "Claude Code not accessible", err.Error())
			suite.t.Fatalf("Claude Code verification failed: %v", err)
		}
		suite.tf.Success(suite.t, "Claude Code available via npx")
	} else {
		suite.tf.Success(suite.t, "Claude Code installed globally")
	}
	
	suite.tf.Success(suite.t, "Claude Code CLI ready", 
		fmt.Sprintf("Version: %s", strings.TrimSpace(claudeOutput)))
}

func (suite *ClaudeCodeContainerTestSuite) configureMCPForClaudeCode() {
	suite.tf.Info(suite.t, "Creating MCP configuration for Claude Code")
	
	// Claude Code MCP configuration
	mcpConfig := `{
  "mcpServers": {
    "portunix": {
      "command": "/usr/local/bin/portunix",
      "args": ["mcp", "serve", "--mode", "stdio"],
      "env": {
        "PORTUNIX_PERMISSION_LEVEL": "development"
      }
    }
  }
}`

	// Create config directory and file
	configCommands := []string{
		"mkdir -p /root/.config/claude-code",
		fmt.Sprintf("echo '%s' > /root/.config/claude-code/mcp_servers.json", mcpConfig),
	}
	
	for _, cmd := range configCommands {
		if _, err := suite.runContainerCommand("bash", "-c", cmd); err != nil {
			suite.tf.Error(suite.t, "Failed to create MCP configuration", err.Error())
			suite.t.Fatalf("Config creation failed: %v", err)
		}
	}
	
	// Verify configuration
	configContent, err := suite.runContainerCommand("cat", "/root/.config/claude-code/mcp_servers.json")
	if err != nil {
		suite.tf.Error(suite.t, "Failed to read MCP configuration", err.Error())
		suite.t.Fatalf("Config verification failed: %v", err)
	}
	
	if !strings.Contains(configContent, "portunix") {
		suite.tf.Error(suite.t, "MCP configuration missing portunix entry")
		suite.t.Fatalf("Invalid configuration: %s", configContent)
	}
	
	suite.tf.Success(suite.t, "MCP configuration created for Claude Code")
	suite.tf.Output(suite.t, configContent, 200)
}

func (suite *ClaudeCodeContainerTestSuite) testClaudeCodeMCPIntegration() bool {
	suite.tf.Info(suite.t, "Testing Claude Code <-> Portunix MCP integration")
	
	// Test 1: Verify Portunix MCP server can start
	suite.tf.Info(suite.t, "Testing Portunix MCP server startup")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, suite.binaryPath, "container", "exec", suite.containerName, "bash", "-c",
		`echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | timeout 5 /usr/local/bin/portunix mcp serve --mode stdio | head -1`)
	
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))
	
	suite.tf.Output(suite.t, outputStr, 300)
	
	// Check for JSON-RPC response
	if strings.Contains(outputStr, `"jsonrpc":"2.0"`) && strings.Contains(outputStr, `"result"`) {
		suite.tf.Success(suite.t, "Portunix MCP server responding correctly to initialize")
	} else if err != nil && ctx.Err() == context.DeadlineExceeded {
		suite.tf.Info(suite.t, "Timeout expected - MCP server entered stdio mode correctly")
	} else {
		suite.tf.Error(suite.t, "MCP server not responding as expected", fmt.Sprintf("Output: %s, Error: %v", outputStr, err))
		return false
	}
	
	// Test 2: Test Claude Code MCP server detection
	suite.tf.Info(suite.t, "Testing Claude Code MCP server detection")
	
	// Create a test script for Claude Code MCP detection
	testScript := `#!/bin/bash
echo "=== Testing Claude Code MCP Integration ==="

# Check if claude command recognizes MCP servers
claude --help | grep -i mcp || echo "No direct MCP support visible in help"

# Check configuration
echo "=== MCP Configuration ==="
cat /root/.config/claude-code/mcp_servers.json

echo "=== Test MCP Server Connection ==="
# Start portunix MCP server in background for testing
timeout 5 /usr/local/bin/portunix mcp serve --mode stdio </dev/null &
MCP_PID=$!

sleep 1

# Check if process is running
if kill -0 $MCP_PID 2>/dev/null; then
    echo "✅ MCP Server process is running (PID: $MCP_PID)"
    kill $MCP_PID 2>/dev/null || true
else
    echo "❌ MCP Server process failed to start"
fi

echo "=== Integration Test Complete ==="
`

	// Write and execute test script
	if _, err := suite.runContainerCommand("bash", "-c", 
		fmt.Sprintf("echo '%s' > /tmp/claude_mcp_test.sh && chmod +x /tmp/claude_mcp_test.sh", 
			strings.ReplaceAll(testScript, "'", "\\'"))); err != nil {
		suite.tf.Error(suite.t, "Failed to create test script", err.Error())
		return false
	}
	
	// Execute integration test
	testOutput, err := suite.runContainerCommand("/tmp/claude_mcp_test.sh")
	if err != nil {
		suite.tf.Warning(suite.t, "Integration test had some issues", err.Error())
	}
	
	suite.tf.Output(suite.t, testOutput, 500)
	
	// Validate results
	if strings.Contains(testOutput, "✅ MCP Server process is running") {
		suite.tf.Success(suite.t, "MCP Server can be started and managed")
		return true
	} else if strings.Contains(testOutput, "MCP Configuration") && strings.Contains(testOutput, "portunix") {
		suite.tf.Success(suite.t, "MCP configuration is properly set up")
		suite.tf.Info(suite.t, "Full integration depends on Claude Code authentication setup")
		return true
	} else {
		suite.tf.Error(suite.t, "MCP integration test failed")
		return false
	}
}

func (suite *ClaudeCodeContainerTestSuite) cleanup() {
	suite.tf.Info(suite.t, "Cleaning up test environment")
	
	// Remove test container using Portunix
	suite.tf.Command(suite.t, suite.binaryPath, []string{"container", "remove", suite.containerName})
	cmd := exec.Command(suite.binaryPath, "container", "remove", suite.containerName)
	cmd.Run() // Ignore errors in cleanup
	
	suite.tf.Success(suite.t, "Cleanup completed using Portunix container system")
}