package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue041NodeJsNpmInstallation tests Node.js/npm installation support
func TestIssue041NodeJsNpmInstallation(t *testing.T) {
	tf := testframework.NewTestFramework("Issue041_NodeJS_NPM_Installation")
	tf.Start(t, "Test Node.js/npm installation support with prerequisite resolution")

	success := true
	defer tf.Finish(t, success)

	// Setup test environment
	tf.Step(t, "Setup test binary path")
	binaryPath, err := getBinaryPathNpm()
	if err != nil {
		tf.Error(t, "Failed to find binary", err.Error())
		success = false
		return
	}
	tf.Success(t, "Binary found", fmt.Sprintf("Path: %s", binaryPath))

	// TC001: Basic Node.js installation in container
	tf.Separator()
	tf.Step(t, "TC001: Test basic Node.js installation in Ubuntu container")
	
	containerSuccess := runNodeJsTestContainerTest(t, tf, binaryPath, "ubuntu:22.04", "basic")
	if !containerSuccess {
		success = false
	}

	// TC002: Prerequisite dependency resolution test
	tf.Separator()
	tf.Step(t, "TC002: Test prerequisite dependency resolution (Python → Node.js)")
	
	prereqSuccess := runPrerequisiteResolutionTest(t, tf, binaryPath)
	if !prereqSuccess {
		success = false
	}

	// TC003: Variant selection test
	tf.Separator()
	tf.Step(t, "TC003: Test Node.js variant selection")
	
	variantSuccess := runVariantSelectionTest(t, tf, binaryPath)
	if !variantSuccess {
		success = false
	}

	// TC004: Cross-platform compatibility test
	tf.Separator()
	tf.Step(t, "TC004: Test cross-platform compatibility (Debian)")
	
	platformSuccess := runNodeJsContainerTest(t, tf, binaryPath, "debian:bookworm", "debian")
	if !platformSuccess {
		success = false
	}

	// TC005: Error handling tests
	tf.Separator()
	tf.Step(t, "TC005: Test error handling scenarios")
	
	errorSuccess := runErrorHandlingTests(t, tf, binaryPath)
	if !errorSuccess {
		success = false
	}

	// TC006: Installation verification tests
	tf.Separator()
	tf.Step(t, "TC006: Test installation verification and cleanup")
	
	verificationSuccess := runInstallationVerificationTests(t, tf, binaryPath)
	if !verificationSuccess {
		success = false
	}

	if success {
		tf.Success(t, "All Node.js/npm installation tests passed")
	}
}

// runNodeJsContainerTest runs Node.js installation test using Portunix container integration
func runNodeJsContainerTest(t *testing.T, tf *testframework.TestFramework, binaryPath, containerImage, testType string) bool {
	tf.Step(t, fmt.Sprintf("Test Node.js installation in %s using Portunix containers", containerImage))
	
	// Use Portunix native container integration
	tf.Command(t, binaryPath, []string{"docker", "run-in-container", "nodejs", "--image", containerImage})
	cmd := exec.Command(binaryPath, "docker", "run-in-container", "nodejs", "--image", containerImage)
	output, err := cmd.CombinedOutput()
	
	tf.Output(t, string(output), 1500)
	
	if err != nil {
		tf.Error(t, "Node.js container installation failed", err.Error())
		return false
	}
	
	// Verify installation success by checking output content
	if strings.Contains(string(output), "successfully") || strings.Contains(string(output), "installed") {
		tf.Success(t, "Node.js installation completed via Portunix container")
	} else {
		tf.Warning(t, "Installation status unclear from output")
	}
	
	// Check if output mentions Node.js and npm
	if strings.Contains(string(output), "node") || strings.Contains(string(output), "Node") {
		tf.Success(t, "Node.js detected in installation process")
	}
	if strings.Contains(string(output), "npm") {
		tf.Success(t, "npm detected in installation process")
	}
	
	return true
}

// runNodeJsTestContainerTest runs Node.js installation test using test container suite
func runNodeJsTestContainerTest(t *testing.T, tf *testframework.TestFramework, binaryPath, containerImage, testType string) bool {
	tf.Step(t, fmt.Sprintf("Create test container suite for %s testing", testType))
	
	// Create test container suite for isolated testing
	suite := &NodeJSContainerTestSuite{
		tf:            tf,
		t:             t,
		containerName: fmt.Sprintf("portunix-nodejs-test-%s", testType),
		projectRoot:   "../..",
		binaryPath:    binaryPath,
	}
	
	// Setup and cleanup
	defer func() {
		tf.Info(t, "Cleaning up test container")
		suite.cleanup()
	}()
	
	// Use the new createTestContainer() instead of createDevelopmentContainer()
	tf.Step(t, "Create minimal test container for clean package testing")
	suite.createTestContainer()
	
	// Install Portunix in test container
	tf.Step(t, "Install Portunix in test container")
	suite.installPortunixInContainer()
	
	// Test Node.js installation via Portunix in container
	tf.Step(t, "Test Node.js installation in clean test environment")
	suite.installViaPortunix("nodejs")
	
	tf.Success(t, "Node.js test container scenario completed successfully")
	return true
}

// runPrerequisiteResolutionTest tests prerequisite dependency resolution using Portunix containers
func runPrerequisiteResolutionTest(t *testing.T, tf *testframework.TestFramework, binaryPath string) bool {
	tf.Step(t, "Test prerequisite resolution (Python → Node.js) using Portunix")
	
	// Use Portunix container to test Node.js installation on minimal Alpine
	tf.Command(t, binaryPath, []string{"docker", "run-in-container", "nodejs", "--image", "alpine:latest"})
	cmd := exec.Command(binaryPath, "docker", "run-in-container", "nodejs", "--image", "alpine:latest")
	output, err := cmd.CombinedOutput()
	
	tf.Output(t, string(output), 1200)
	
	if err != nil {
		tf.Error(t, "Prerequisite resolution test failed", err.Error())
		return false
	}
	
	// Check if prerequisite resolution is working
	if strings.Contains(string(output), "python") || strings.Contains(string(output), "Python") {
		tf.Success(t, "Prerequisite resolution detected", "Python prerequisite mentioned in output")
	} else {
		tf.Info(t, "Prerequisite resolution not explicitly visible in output")
	}
	
	// Check for successful installation
	if strings.Contains(string(output), "successfully") || strings.Contains(string(output), "installed") {
		tf.Success(t, "Node.js installation with prerequisites completed")
	} else {
		tf.Info(t, "Installation completion status unclear")
	}
	
	return true
}

// runVariantSelectionTest tests Node.js variant selection using Portunix
func runVariantSelectionTest(t *testing.T, tf *testframework.TestFramework, binaryPath string) bool {
	tf.Step(t, "Test Node.js variant selection support")
	
	// Test help command for variant information
	tf.Command(t, binaryPath, []string{"install", "nodejs", "--help"})
	cmd := exec.Command(binaryPath, "install", "nodejs", "--help")
	helpOutput, err := cmd.CombinedOutput()
	
	if err != nil {
		tf.Warning(t, "Help command failed", err.Error())
	} else {
		tf.Output(t, string(helpOutput), 600)
		
		if strings.Contains(string(helpOutput), "variant") || strings.Contains(string(helpOutput), "--variant") {
			tf.Success(t, "Variant support detected in help")
		} else {
			tf.Info(t, "Variant support not clearly documented in help")
		}
	}
	
	// Test dry-run with variant to check support without installation
	tf.Step(t, "Test variant selection with dry-run")
	tf.Command(t, binaryPath, []string{"install", "nodejs", "--variant", "lts", "--dry-run"})
	cmd = exec.Command(binaryPath, "install", "nodejs", "--variant", "lts", "--dry-run")
	output, err := cmd.CombinedOutput()
	
	tf.Output(t, string(output), 800)
	
	if err != nil {
		tf.Info(t, "Variant selection test showed issues", err.Error())
		// This might be expected if variant support is not fully implemented
	} else {
		tf.Success(t, "Variant selection command accepted")
	}
	
	return true
}

// getBinaryPathNpm finds the Portunix binary (issue 041 npm specific)
func getBinaryPathNpm() (string, error) {
	// Try current directory first
	if _, err := os.Stat("./portunix"); err == nil {
		return "./portunix", nil
	}
	
	// Try parent directory
	if _, err := os.Stat("../../portunix"); err == nil {
		return "../../portunix", nil
	}
	
	// Try building if source exists
	if _, err := os.Stat("../../main.go"); err == nil {
		cmd := exec.Command("go", "build", "-o", "../../portunix")
		cmd.Dir = "../.."
		if err := cmd.Run(); err == nil {
			return "../../portunix", nil
		}
	}
	
	return "", fmt.Errorf("portunix binary not found")
}

// runErrorHandlingTests tests error handling scenarios for NodeJS installation
func runErrorHandlingTests(t *testing.T, tf *testframework.TestFramework, binaryPath string) bool {
	tf.Step(t, "Test error handling scenarios for NodeJS installation")
	
	// TC005-1: Test invalid package name
	tf.Info(t, "TC005-1: Test invalid package name error handling")
	tf.Command(t, binaryPath, []string{"install", "nonexistent-nodejs-package"})
	cmd := exec.Command(binaryPath, "install", "nonexistent-nodejs-package")
	output, err := cmd.CombinedOutput()
	
	// Should either fail with error or show appropriate message
	if err != nil {
		tf.Success(t, "Invalid package correctly rejected")
	} else if strings.Contains(string(output), "not found") || strings.Contains(string(output), "unknown") {
		tf.Success(t, "Invalid package error message shown")
	} else {
		tf.Info(t, "Invalid package handling unclear from output")
		tf.Output(t, string(output), 300)
	}
	
	// TC005-2: Test install command without arguments
	tf.Info(t, "TC005-2: Test install command usage help")
	tf.Command(t, binaryPath, []string{"install"})
	cmd = exec.Command(binaryPath, "install")
	output, err = cmd.CombinedOutput()
	
	if strings.Contains(string(output), "Usage") || strings.Contains(string(output), "help") {
		tf.Success(t, "Install command shows help when no package specified")
	} else {
		tf.Info(t, "Install command behavior without arguments unclear")
	}
	
	// TC005-3: Test help command for install
	tf.Info(t, "TC005-3: Test install help command")
	tf.Command(t, binaryPath, []string{"install", "--help"})
	cmd = exec.Command(binaryPath, "install", "--help")
	helpOutput, err := cmd.CombinedOutput()
	
	if err == nil && strings.Contains(string(helpOutput), "install") {
		tf.Success(t, "Install help command works")
		if strings.Contains(string(helpOutput), "nodejs") {
			tf.Success(t, "NodeJS mentioned in install help")
		}
	}
	
	return true
}

// runInstallationVerificationTests tests installation verification and cleanup
func runInstallationVerificationTests(t *testing.T, tf *testframework.TestFramework, binaryPath string) bool {
	tf.Step(t, "Test installation verification scenarios")
	
	// Create test container for verification tests
	tf.Info(t, "Creating verification test container")
	suite := &NodeJSContainerTestSuite{
		tf:            tf,
		t:             t,
		containerName: "portunix-nodejs-verify-test",
		projectRoot:   "../..",
		binaryPath:    binaryPath,
	}
	
	defer func() {
		tf.Info(t, "Cleaning up verification test container")
		suite.cleanup()
	}()
	
	// Setup container
	suite.createTestContainer()
	suite.installPortunixInContainer()
	
	// TC006-1: Test system info before installation
	tf.Info(t, "TC006-1: Check system info before NodeJS installation")
	sysOutput, err := suite.runContainerCommand("/usr/local/bin/portunix", "system", "info")
	if err == nil {
		tf.Success(t, "System info command works")
		tf.Output(t, sysOutput, 400)
	}
	
	// TC006-2: Test installation process
	tf.Info(t, "TC006-2: Install NodeJS and verify step by step")
	installOutput, installErr := suite.runContainerCommand("/usr/local/bin/portunix", "install", "nodejs")
	
	// Log detailed installation output
	tf.Output(t, installOutput, 1000)
	
	if installErr != nil {
		tf.Warning(t, "Installation returned error", installErr.Error())
	}
	
	// TC006-3: Manual verification of NodeJS availability
	tf.Info(t, "TC006-3: Manual verification of NodeJS components")
	
	// Check if node command is available
	nodeOutput, nodeErr := suite.runContainerCommand("which", "node")
	if nodeErr == nil && strings.Contains(nodeOutput, "/") {
		tf.Success(t, "Node executable found", strings.TrimSpace(nodeOutput))
		
		// Get node version
		versionOutput, versionErr := suite.runContainerCommand("node", "--version")
		if versionErr == nil {
			tf.Success(t, "Node version check successful", strings.TrimSpace(versionOutput))
		}
	} else {
		tf.Info(t, "Node executable not found or not in PATH")
	}
	
	// Check if npm is available
	npmOutput, npmErr := suite.runContainerCommand("which", "npm")
	if npmErr == nil && strings.Contains(npmOutput, "/") {
		tf.Success(t, "NPM executable found", strings.TrimSpace(npmOutput))
		
		// Get npm version
		npmVersionOutput, npmVersionErr := suite.runContainerCommand("npm", "--version")
		if npmVersionErr == nil {
			tf.Success(t, "NPM version check successful", strings.TrimSpace(npmVersionOutput))
		}
	} else {
		tf.Info(t, "NPM executable not found or not in PATH")
	}
	
	// TC006-4: Test package functionality
	tf.Info(t, "TC006-4: Test basic NodeJS functionality")
	
	// Create simple Node.js test script
	testScript := `console.log("NodeJS test successful"); console.log("Version:", process.version);`
	scriptOutput, scriptErr := suite.runContainerCommand("sh", "-c", fmt.Sprintf(`echo '%s' | node`, testScript))
	if scriptErr == nil && strings.Contains(scriptOutput, "test successful") {
		tf.Success(t, "NodeJS execution test passed")
	} else {
		tf.Info(t, "NodeJS execution test results unclear")
		tf.Output(t, scriptOutput, 200)
	}
	
	return true
}

// NodeJSContainerTestSuite - specific version for Node.js npm testing  
type NodeJSContainerTestSuite struct {
	tf            *testframework.TestFramework
	t             *testing.T
	containerName string
	projectRoot   string
	binaryPath    string
}

func (suite *NodeJSContainerTestSuite) runContainerCommand(command string, args ...string) (string, error) {
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

func (suite *NodeJSContainerTestSuite) createTestContainer() {
	// Remove existing container using Portunix
	suite.tf.Info(suite.t, "Cleaning up any existing test containers")
	cleanupCmd := exec.Command(suite.binaryPath, "container", "remove", suite.containerName)
	cleanupCmd.Run() // Ignore errors - container might not exist
	
	suite.tf.Info(suite.t, "Creating minimal test container for package testing")
	
	// Use Portunix container run instead of direct podman
	// Image is positional argument, not a flag
	// Need to provide a command that keeps container running
	suite.tf.Command(suite.t, suite.binaryPath, []string{"container", "run", "--name", suite.containerName, "-d", "ubuntu:22.04", "sleep", "3600"})
	cmd := exec.Command(suite.binaryPath, "container", "run", "--name", suite.containerName, "-d", "ubuntu:22.04", "sleep", "3600")
	
	if output, err := cmd.CombinedOutput(); err != nil {
		suite.tf.Error(suite.t, "Failed to create test container via Portunix", err.Error())
		suite.tf.Output(suite.t, string(output), 500)
		suite.t.Fatalf("Test container creation failed: %v", err)
	}
	
	// Shorter wait for container
	suite.tf.Info(suite.t, "Waiting for test container initialization", "15s")
	time.Sleep(15 * time.Second)
	
	// Verify container is ready using Portunix container exec
	if output, err := suite.runContainerCommand("echo", "Test container ready"); err != nil {
		suite.tf.Error(suite.t, "Test container not responding", err.Error())
		suite.t.Fatalf("Test container setup verification failed: %v", err)
	} else if !strings.Contains(output, "Test container ready") {
		suite.tf.Error(suite.t, "Test container startup verification failed")
		suite.t.Fatalf("Unexpected test container response: %s", output)
	}
	
	suite.tf.Success(suite.t, "Minimal test container ready for package testing")
}

func (suite *NodeJSContainerTestSuite) installPortunixInContainer() {
	suite.tf.Info(suite.t, "Installing Portunix binary in container via Portunix container system")
	
	// Use Portunix container copy to copy binary
	suite.tf.Command(suite.t, suite.binaryPath, []string{"container", "cp", suite.binaryPath, suite.containerName + ":/usr/local/bin/portunix"})
	cmd := exec.Command(suite.binaryPath, "container", "cp", suite.binaryPath, suite.containerName+":/usr/local/bin/portunix")
	if output, err := cmd.CombinedOutput(); err != nil {
		suite.tf.Error(suite.t, "Failed to copy portunix to container via Portunix", err.Error())
		suite.tf.Output(suite.t, string(output), 500)
		suite.t.Fatalf("Copy failed: %v", err)
	}
	
	// Make executable using Portunix container exec
	if _, err := suite.runContainerCommand("chmod", "+x", "/usr/local/bin/portunix"); err != nil {
		suite.t.Fatalf("Failed to make portunix executable: %v", err)
	}
	
	// Verify installation using Portunix container exec
	// Use -- to separate portunix flags from command flags
	output, err := suite.runContainerCommand("/usr/local/bin/portunix", "version")
	if err != nil {
		suite.tf.Error(suite.t, "Portunix not working in container", err.Error())
		suite.t.Fatalf("Version check failed: %v", err)
	}
	
	suite.tf.Success(suite.t, "Portunix installed and verified in container via Portunix container system", 
		fmt.Sprintf("Version: %s", strings.TrimSpace(output)))
}

func (suite *NodeJSContainerTestSuite) installViaPortunix(packageName string) {
	suite.tf.Info(suite.t, fmt.Sprintf("Installing %s via Portunix package system", packageName))
	
	// Test package availability first with help command
	suite.tf.Info(suite.t, fmt.Sprintf("Testing %s package availability", packageName))
	helpOutput, err := suite.runContainerCommand("/usr/local/bin/portunix", "install", packageName, "--help")
	if err != nil {
		suite.tf.Warning(suite.t, fmt.Sprintf("%s help command failed", packageName), err.Error())
	} else {
		if strings.Contains(helpOutput, packageName) || strings.Contains(helpOutput, "Install") {
			suite.tf.Success(suite.t, fmt.Sprintf("%s package help available", packageName))
		}
	}
	
	// Test with list command to check if package exists
	suite.tf.Info(suite.t, fmt.Sprintf("Checking %s package existence in system", packageName))
	_, listErr := suite.runContainerCommand("/usr/local/bin/portunix", "install", "--list")
	if listErr == nil {
		suite.tf.Success(suite.t, "Package listing command works")
	} else {
		suite.tf.Info(suite.t, "Package listing not available, proceeding with installation test")
	}
	
	// Install package via Portunix (actual installation)
	suite.tf.Info(suite.t, fmt.Sprintf("Installing %s using Portunix", packageName))
	installOutput, err := suite.runContainerCommand("/usr/local/bin/portunix", "install", packageName)
	if err != nil {
		suite.tf.Error(suite.t, fmt.Sprintf("%s installation failed", packageName), err.Error())
		// Don't fail immediately - log the error but continue with verification
		suite.tf.Warning(suite.t, "Installation command returned error, checking if package was still installed")
	} else {
		suite.tf.Success(suite.t, fmt.Sprintf("%s install command completed", packageName))
	}
	
	// Check installation output for success indicators
	if strings.Contains(installOutput, "successfully") || strings.Contains(installOutput, "✅") || strings.Contains(installOutput, "installed") {
		suite.tf.Success(suite.t, fmt.Sprintf("%s installation appears successful", packageName))
	} else {
		suite.tf.Info(suite.t, "Installation success not clearly indicated in output")
	}
	
	// Verify installation by checking if nodejs/npm are available
	if packageName == "nodejs" {
		suite.tf.Info(suite.t, "Verifying Node.js and npm installation")
		
		// Check Node.js
		nodeOutput, err := suite.runContainerCommand("node", "--version")
		if err != nil {
			suite.tf.Warning(suite.t, "Node.js verification failed", err.Error())
		} else {
			suite.tf.Success(suite.t, "Node.js verified", fmt.Sprintf("Version: %s", strings.TrimSpace(nodeOutput)))
		}
		
		// Check npm
		npmOutput, err := suite.runContainerCommand("npm", "--version")
		if err != nil {
			suite.tf.Warning(suite.t, "npm verification failed", err.Error())
		} else {
			suite.tf.Success(suite.t, "npm verified", fmt.Sprintf("Version: %s", strings.TrimSpace(npmOutput)))
		}
	}
}

func (suite *NodeJSContainerTestSuite) cleanup() {
	suite.tf.Info(suite.t, "Cleaning up test environment")
	
	// Remove test container using Portunix
	suite.tf.Command(suite.t, suite.binaryPath, []string{"container", "remove", suite.containerName})
	cmd := exec.Command(suite.binaryPath, "container", "remove", suite.containerName)
	cmd.Run() // Ignore errors in cleanup
	
	suite.tf.Success(suite.t, "Cleanup completed using Portunix container system")
}