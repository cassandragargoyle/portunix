package integration

import (
	"os/exec"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue079MethodVersionOverride tests the method and version override functionality
func TestIssue079MethodVersionOverride(t *testing.T) {
	tf := testframework.NewTestFramework("Issue079_Method_Version_Override")
	tf.Start(t, "Test method and version override functionality for install command")

	success := true
	defer tf.Finish(t, success)

	// Get binary path
	binaryPath := "../../portunix"
	tf.Step(t, "Setup test binary")
	tf.Info(t, "Binary path:", binaryPath)

	// Test 1: --list-methods functionality
	tf.Step(t, "Test --list-methods functionality")
	cmd := exec.Command(binaryPath, "install", "hugo", "--list-methods")
	output, err := cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "Command failed", err.Error())
		success = false
		return
	}

	outputStr := string(output)
	tf.Output(t, outputStr, 500)

	// Verify output contains expected information
	if !strings.Contains(outputStr, "Available installation methods for 'hugo'") {
		tf.Error(t, "Missing expected header in list-methods output")
		success = false
		return
	}

	if !strings.Contains(outputStr, "extended") || !strings.Contains(outputStr, "apt") {
		tf.Error(t, "Missing expected methods in output")
		success = false
		return
	}

	tf.Success(t, "--list-methods test passed")
	tf.Separator()

	// Test 2: Method override with dry-run
	tf.Step(t, "Test method override with --method=apt")
	cmd = exec.Command(binaryPath, "install", "hugo", "--method=apt", "--dry-run")
	output, err = cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "Method override command failed", err.Error())
		success = false
		return
	}

	outputStr = string(output)
	tf.Output(t, outputStr, 500)

	// Verify method override worked
	if !strings.Contains(outputStr, "Variant: apt") {
		tf.Error(t, "Method override did not work - expected apt variant")
		success = false
		return
	}

	if !strings.Contains(outputStr, "Installation type: apt") {
		tf.Error(t, "Method override did not set correct installation type")
		success = false
		return
	}

	tf.Success(t, "Method override test passed")
	tf.Separator()

	// Test 3: Version override with latest
	tf.Step(t, "Test version override with --version=latest")
	cmd = exec.Command(binaryPath, "install", "hugo", "--version=latest", "--dry-run")
	output, err = cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "Version override command failed", err.Error())
		success = false
		return
	}

	outputStr = string(output)
	tf.Output(t, outputStr, 500)

	// Verify version override selected a different method (tar.gz for latest)
	if !strings.Contains(outputStr, "DRY-RUN MODE") {
		tf.Error(t, "Version override did not trigger dry-run mode correctly")
		success = false
		return
	}

	// Should select a direct download method for latest
	if !strings.Contains(outputStr, "tar.gz") && !strings.Contains(outputStr, "deb") {
		tf.Error(t, "Version override did not select appropriate method for latest version")
		success = false
		return
	}

	tf.Success(t, "Version override test passed")
	tf.Separator()

	// Test 4: Invalid method handling
	tf.Step(t, "Test invalid method error handling")
	cmd = exec.Command(binaryPath, "install", "hugo", "--method=nonexistent", "--dry-run")
	output, err = cmd.CombinedOutput()

	// This should fail with error
	if err == nil {
		tf.Error(t, "Invalid method should have failed but didn't")
		success = false
		return
	}

	outputStr = string(output)
	tf.Output(t, outputStr, 300)

	// Verify error message is appropriate
	if !strings.Contains(outputStr, "not found") || !strings.Contains(outputStr, "nonexistent") {
		tf.Error(t, "Error message for invalid method is not descriptive enough")
		success = false
		return
	}

	tf.Success(t, "Invalid method error handling test passed")
	tf.Separator()

	// Test 5: Combination of method and dry-run
	tf.Step(t, "Test method override with snap")
	cmd = exec.Command(binaryPath, "install", "hugo", "--method=snap", "--dry-run")
	output, err = cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "Snap method override failed", err.Error())
		success = false
		return
	}

	outputStr = string(output)
	tf.Output(t, outputStr, 500)

	// Verify snap method was selected
	if !strings.Contains(outputStr, "Variant: snap") {
		tf.Error(t, "Snap method override did not work")
		success = false
		return
	}

	if !strings.Contains(outputStr, "Installation type: snap") {
		tf.Error(t, "Snap method override did not set correct installation type")
		success = false
		return
	}

	tf.Success(t, "Snap method override test passed")

	tf.Success(t, "All Issue 079 tests completed successfully")
}

// TestIssue079BackwardCompatibility tests that existing functionality still works
func TestIssue079BackwardCompatibility(t *testing.T) {
	tf := testframework.NewTestFramework("Issue079_Backward_Compatibility")
	tf.Start(t, "Test backward compatibility with existing variant system")

	success := true
	defer tf.Finish(t, success)

	binaryPath := "../../portunix"

	// Test 1: Legacy variant syntax still works
	tf.Step(t, "Test legacy variant syntax")
	cmd := exec.Command(binaryPath, "install", "hugo", "extended", "--dry-run")
	output, err := cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "Legacy variant syntax failed", err.Error())
		success = false
		return
	}

	outputStr := string(output)
	tf.Output(t, outputStr, 500)

	// Should still work with positional variant
	if !strings.Contains(outputStr, "Variant: extended") {
		tf.Error(t, "Legacy variant syntax not working")
		success = false
		return
	}

	tf.Success(t, "Legacy variant syntax test passed")
	tf.Separator()

	// Test 2: --variant flag still works
	tf.Step(t, "Test --variant flag compatibility")
	cmd = exec.Command(binaryPath, "install", "hugo", "--variant=standard", "--dry-run")
	output, err = cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "--variant flag failed", err.Error())
		success = false
		return
	}

	outputStr = string(output)
	tf.Output(t, outputStr, 500)

	// Should work with --variant flag
	if !strings.Contains(outputStr, "Variant: standard") {
		tf.Error(t, "--variant flag not working")
		success = false
		return
	}

	tf.Success(t, "--variant flag test passed")

	tf.Success(t, "All backward compatibility tests passed")
}