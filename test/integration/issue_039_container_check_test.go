package integration

import (
	"os/exec"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue039ContainerCheck tests the container check command implementation.
// This test verifies the unified container runtime capability detection system.
//
// Issue #039: Container Runtime Capability Detection
// Implements:
//   - Centralized container runtime detection
//   - CLI command: portunix container check
//   - Feature detection (compose, buildx, etc.)
//   - Version detection for Docker and Podman

func TestIssue039ContainerCheck(t *testing.T) {
	tf := testframework.NewTestFramework("Issue039_Container_Check")
	tf.Start(t, "Test container runtime capability detection command")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	binaryPath := "../../portunix"

	// Test 1: Verify container check command exists
	tf.Step(t, "Test container check command execution")

	cmd := exec.Command(binaryPath, "container", "check")
	tf.Command(t, binaryPath, []string{"container", "check"})

	output, err := cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Failed to execute container check command", err.Error())
		success = false
		return
	}

	outputStr := string(output)
	tf.Output(t, outputStr, 1000)
	tf.Success(t, "Container check command executed successfully")

	tf.Separator()

	// Test 2: Verify output contains runtime status section
	tf.Step(t, "Verify Container Runtime Status section")

	if !strings.Contains(outputStr, "Container Runtime Status:") {
		tf.Error(t, "Container Runtime Status section not found in output")
		success = false
		return
	}
	tf.Success(t, "✓ Container Runtime Status section found")

	// Test 3: Verify Docker status line exists
	tf.Step(t, "Verify Docker detection in output")

	dockerDetected := false
	if strings.Contains(outputStr, "Docker: ✓ Available") {
		tf.Success(t, "✓ Docker detected as available")
		dockerDetected = true
	} else if strings.Contains(outputStr, "Docker: ✗ Not available") ||
	          strings.Contains(outputStr, "Docker: ✗ Not detected") {
		tf.Success(t, "✓ Docker detected as not available")
	} else {
		tf.Error(t, "Docker status line not found or malformed")
		success = false
	}

	// Test 4: Verify Podman status line exists
	tf.Step(t, "Verify Podman detection in output")

	podmanDetected := false
	if strings.Contains(outputStr, "Podman: ✓ Available") {
		tf.Success(t, "✓ Podman detected as available")
		podmanDetected = true
	} else if strings.Contains(outputStr, "Podman: ✗ Not available") ||
	          strings.Contains(outputStr, "Podman: ✗ Not detected") {
		tf.Success(t, "✓ Podman detected as not available")
	} else {
		tf.Error(t, "Podman status line not found or malformed")
		success = false
	}

	tf.Separator()

	// Test 5: Verify version detection if runtime is available
	tf.Step(t, "Verify version detection for available runtimes")

	if dockerDetected {
		// Should contain version information like "version 24.0.5"
		if strings.Contains(outputStr, "version") {
			tf.Success(t, "✓ Docker version information found")
		} else {
			tf.Warning(t, "Docker version information not found in output")
		}
	}

	if podmanDetected {
		// Should contain version information
		if strings.Contains(outputStr, "version") {
			tf.Success(t, "✓ Podman version information found")
		} else {
			tf.Warning(t, "Podman version information not found in output")
		}
	}

	if !dockerDetected && !podmanDetected {
		tf.Info(t, "No container runtime detected - version test skipped")

		// Should show installation suggestion
		if strings.Contains(outputStr, "portunix install docker") ||
		   strings.Contains(outputStr, "portunix install podman") {
			tf.Success(t, "✓ Installation suggestion provided")
		}
	}

	tf.Separator()

	// Test 6: Verify preferred runtime selection
	tf.Step(t, "Verify preferred runtime selection logic")

	if dockerDetected || podmanDetected {
		if strings.Contains(outputStr, "Preferred:") {
			tf.Success(t, "✓ Preferred runtime selection displayed")

			// Extract preferred runtime
			if dockerDetected && strings.Contains(outputStr, "Preferred: docker") {
				tf.Success(t, "✓ Docker selected as preferred (expected when available)")
			} else if podmanDetected && strings.Contains(outputStr, "Preferred: podman") {
				tf.Success(t, "✓ Podman selected as preferred")
			}
		} else {
			tf.Warning(t, "Preferred runtime not displayed")
		}
	}

	tf.Separator()

	// Test 7: Verify capabilities section (if runtime available)
	tf.Step(t, "Verify capabilities detection")

	if dockerDetected || podmanDetected {
		if strings.Contains(outputStr, "Capabilities:") {
			tf.Success(t, "✓ Capabilities section found")

			// Check for expected capabilities
			capabilities := []string{
				"Compose support",
				"Volume mounting",
				"Network creation",
				"Runtime active",
			}

			foundCapabilities := 0
			for _, cap := range capabilities {
				if strings.Contains(outputStr, cap) {
					foundCapabilities++
				}
			}

			if foundCapabilities > 0 {
				tf.Success(t, "✓ Capabilities detected",
					"Found "+string(rune(foundCapabilities+'0'))+" capabilities")
			} else {
				tf.Info(t, "No specific capabilities listed (may be normal)")
			}
		} else {
			tf.Info(t, "Capabilities section not found (may be normal if no features detected)")
		}
	}

	tf.Separator()

	// Test 8: Test --refresh flag
	tf.Step(t, "Test --refresh flag functionality")

	refreshCmd := exec.Command(binaryPath, "container", "check", "--refresh")
	tf.Command(t, binaryPath, []string{"container", "check", "--refresh"})

	refreshOutput, refreshErr := refreshCmd.CombinedOutput()
	if refreshErr != nil {
		tf.Error(t, "Failed to execute container check --refresh", refreshErr.Error())
		success = false
	} else {
		refreshStr := string(refreshOutput)
		tf.Output(t, refreshStr, 500)

		// Output should be similar to normal check
		if strings.Contains(refreshStr, "Container Runtime Status:") {
			tf.Success(t, "✓ Refresh flag works correctly")
		} else {
			tf.Error(t, "Refresh flag output malformed")
			success = false
		}
	}

	tf.Separator()

	// Test 9: Verify actual runtime detection accuracy
	tf.Step(t, "Verify detection accuracy against actual system state")

	// Check if Docker is actually available
	dockerCmd := exec.Command("docker", "version")
	dockerActuallyAvailable := dockerCmd.Run() == nil

	// Check if Podman is actually available
	podmanCmd := exec.Command("podman", "version")
	podmanActuallyAvailable := podmanCmd.Run() == nil

	// Compare with what Portunix detected
	if dockerActuallyAvailable == dockerDetected {
		tf.Success(t, "✓ Docker detection matches actual system state")
	} else {
		tf.Error(t, "Docker detection mismatch",
			"Portunix detected: "+boolToAvailable(dockerDetected),
			"Actually available: "+boolToAvailable(dockerActuallyAvailable))
		success = false
	}

	if podmanActuallyAvailable == podmanDetected {
		tf.Success(t, "✓ Podman detection matches actual system state")
	} else {
		tf.Error(t, "Podman detection mismatch",
			"Portunix detected: "+boolToAvailable(podmanDetected),
			"Actually available: "+boolToAvailable(podmanActuallyAvailable))
		success = false
	}

	tf.Separator()

	// Test 10: Verify help text for container check
	tf.Step(t, "Verify help text for container check command")

	helpCmd := exec.Command(binaryPath, "container", "check", "--help")
	tf.Command(t, binaryPath, []string{"container", "check", "--help"})

	helpOutput, helpErr := helpCmd.CombinedOutput()
	if helpErr != nil {
		tf.Warning(t, "Failed to get help text", helpErr.Error())
	} else {
		helpStr := string(helpOutput)
		tf.Output(t, helpStr, 500)

		if strings.Contains(helpStr, "Check container runtime capabilities") ||
		   strings.Contains(helpStr, "container runtime") {
			tf.Success(t, "✓ Help text available and descriptive")
		} else {
			tf.Warning(t, "Help text may need improvement")
		}
	}
}

func boolToAvailable(b bool) string {
	if b {
		return "available"
	}
	return "not available"
}
