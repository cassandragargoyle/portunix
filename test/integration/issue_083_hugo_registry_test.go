package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

func TestIssue083HugoRegistryInstallation(t *testing.T) {
	tf := testframework.NewTestFramework("Issue083_HugoRegistry")
	tf.Start(t, "Test Hugo installation from new Package Registry")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	// Step 1: Determine binary path
	tf.Step(t, "Setup test binary")
	binaryPath := "../../portunix"
	if _, err := os.Stat(binaryPath); err != nil {
		tf.Error(t, "Binary not found at expected path", binaryPath)
		success = false
		return
	}
	tf.Success(t, "Binary found", binaryPath)

	// Test Case 1: Dry-run installation
	tf.Separator()
	tf.Step(t, "TC01: Test dry-run installation of Hugo")
	tf.Command(t, binaryPath, []string{"install", "hugo", "--dry-run"})

	cmd := exec.Command(binaryPath, "install", "hugo", "--dry-run")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	tf.Output(t, outputStr, 500)

	if err != nil {
		tf.Error(t, "Dry-run failed", err.Error())
		success = false
		return
	}

	// Verify expected output
	expectedStrings := []string{
		"INSTALLING: hugo",
		"Fast and flexible static site generator",
		"DRY-RUN MODE",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			tf.Error(t, "Missing expected output", expected)
			success = false
		} else {
			tf.Success(t, "Found expected string", expected)
		}
	}

	// Test Case 2: Container-based installation test
	tf.Separator()
	tf.Step(t, "TC02: Test Hugo installation in container")
	tf.Info(t, "This test should be run with container support")

	// Check if Docker or Podman is available
	dockerAvailable := checkDockerAvailable()
	podmanAvailable := checkPodmanAvailable()

	if !dockerAvailable && !podmanAvailable {
		tf.Warning(t, "Skipping container test - no container runtime available")
		tf.Info(t, "Install Docker or Podman to enable container tests")
	} else {
		containerRuntime := "docker"
		if !dockerAvailable && podmanAvailable {
			containerRuntime = "podman"
		}
		tf.Info(t, "Using container runtime", containerRuntime)

		// Test installation in container
		tf.Command(t, binaryPath, []string{containerRuntime, "run-in-container", "hugo", "--image", "ubuntu:22.04", "--dry-run"})

		cmd := exec.Command(binaryPath, containerRuntime, "run-in-container", "hugo", "--image", "ubuntu:22.04", "--dry-run")
		output, err := cmd.CombinedOutput()

		tf.Output(t, string(output), 500)

		if err == nil {
			tf.Success(t, "Container dry-run completed successfully")
		} else {
			// Container test is optional, don't fail the whole test
			tf.Warning(t, "Container dry-run failed (optional)", err.Error())
		}
	}

	// Test Case 3: Verify registry loading
	tf.Separator()
	tf.Step(t, "TC03: Verify Hugo exists in registry")

	// Check that Hugo package is recognized (will fail if not in registry)
	cmd = exec.Command(binaryPath, "install", "hugo", "--dry-run")
	err = cmd.Run()

	if err != nil {
		tf.Error(t, "Hugo not found in registry", err.Error())
		success = false
	} else {
		tf.Success(t, "Hugo found in registry and can be installed")
	}

	// Test Case 4: Test variant resolution
	tf.Separator()
	tf.Step(t, "TC04: Test platform-specific variant resolution")

	// The registry should pick appropriate variant for current platform
	cmd = exec.Command(binaryPath, "install", "hugo", "--dry-run")
	output, _ = cmd.CombinedOutput()
	outputStr = string(output)

	tf.Output(t, outputStr, 300)

	// Check that a variant was selected
	if strings.Contains(outputStr, "Variant:") {
		tf.Success(t, "Variant was automatically selected")
	} else {
		tf.Error(t, "No variant information found in output")
		success = false
	}

	// Test Case 5: Test error handling for non-existent package
	tf.Separator()
	tf.Step(t, "TC05: Test error handling for non-existent package")

	cmd = exec.Command(binaryPath, "install", "nonexistent-package-xyz", "--dry-run")
	output, err = cmd.CombinedOutput()

	tf.Output(t, string(output), 300)

	if err != nil {
		tf.Success(t, "Correctly failed for non-existent package")
	} else {
		tf.Error(t, "Should have failed for non-existent package")
		success = false
	}
}

// Helper function to check if Docker is available
func checkDockerAvailable() bool {
	cmd := exec.Command("docker", "--version")
	return cmd.Run() == nil
}

// Helper function to check if Podman is available
func checkPodmanAvailable() bool {
	cmd := exec.Command("podman", "--version")
	return cmd.Run() == nil
}