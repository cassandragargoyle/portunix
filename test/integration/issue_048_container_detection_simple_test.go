package integration

import (
	"os/exec"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue048SimpleDetection tests container runtime detection on the host system only.
// Container-based scenarios are documented but not executed due to current limitations.
//
// TODO: Full container-based testing will be implemented once Portunix has VM infrastructure
// (e.g., ProxMox VMs) where Docker and Podman can be properly installed and tested.

func TestIssue048SimpleDetection(t *testing.T) {
	tf := testframework.NewTestFramework("Issue048_Simple_Detection")
	tf.Start(t, "Test enhanced container runtime detection in system info command")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	binaryPath := "../../portunix"

	// Test 1: Verify system info command runs successfully
	tf.Step(t, "Test system info command execution")

	cmd := exec.Command(binaryPath, "system", "info")
	tf.Command(t, binaryPath, []string{"system", "info"})

	output, err := cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Failed to execute system info", err.Error())
		success = false
		return
	}

	outputStr := string(output)
	tf.Output(t, outputStr, 800)
	tf.Success(t, "System info command executed successfully")

	tf.Separator()

	// Test 2: Verify Container Runtimes section exists
	tf.Step(t, "Verify Container Runtimes section")

	if !strings.Contains(outputStr, "Container Runtimes:") {
		tf.Error(t, "Container Runtimes section not found in output")
		success = false
		return
	}
	tf.Success(t, "✓ Container Runtimes section found")

	// Test 3: Verify Docker status line
	tf.Step(t, "Verify Docker detection")

	if strings.Contains(outputStr, "Docker:       installed") {
		tf.Success(t, "✓ Docker detected as installed")
	} else if strings.Contains(outputStr, "Docker:       not installed") {
		tf.Success(t, "✓ Docker detected as not installed")
	} else {
		tf.Error(t, "Docker status line not found or malformed")
		success = false
	}

	// Test 4: Verify Podman status line
	tf.Step(t, "Verify Podman detection")

	if strings.Contains(outputStr, "Podman:       installed") {
		tf.Success(t, "✓ Podman detected as installed")
	} else if strings.Contains(outputStr, "Podman:       not installed") {
		tf.Success(t, "✓ Podman detected as not installed")
	} else {
		tf.Error(t, "Podman status line not found or malformed")
		success = false
	}

	// Test 5: Verify Container Available status
	tf.Step(t, "Verify Container Available flag")

	if !strings.Contains(outputStr, "Container Available:") {
		tf.Error(t, "Container Available status not found")
		success = false
		return
	}
	tf.Success(t, "✓ Container Available status found")

	// Test 6: Verify Container Available logic
	tf.Step(t, "Verify Container Available logic correctness")

	dockerInstalled := strings.Contains(outputStr, "Docker:       installed")
	podmanInstalled := strings.Contains(outputStr, "Podman:       installed")
	containerAvailable := strings.Contains(outputStr, "Container Available: true")

	expectedAvailable := dockerInstalled || podmanInstalled

	if expectedAvailable && containerAvailable {
		tf.Success(t, "✓ Container Available: true (correct - at least one runtime installed)")
		tf.Info(t, "Detected runtimes",
			"Docker: "+boolToStatus(dockerInstalled),
			"Podman: "+boolToStatus(podmanInstalled))
	} else if !expectedAvailable && !containerAvailable {
		tf.Success(t, "✓ Container Available: false (correct - no runtime installed)")
		tf.Info(t, "No container runtimes detected")
	} else {
		tf.Error(t, "Container Available logic mismatch",
			"Docker: "+boolToStatus(dockerInstalled),
			"Podman: "+boolToStatus(podmanInstalled),
			"Container Available: "+boolToStatus(containerAvailable),
			"Expected Available: "+boolToStatus(expectedAvailable))
		success = false
	}

	tf.Separator()

	// Test 7: Document expected behavior for different scenarios
	tf.Step(t, "Document expected behavior for different scenarios")

	scenarios := []struct {
		name     string
		docker   string
		podman   string
		expected string
	}{
		{
			name:     "No container runtime",
			docker:   "not installed",
			podman:   "not installed",
			expected: "false",
		},
		{
			name:     "Docker only",
			docker:   "installed",
			podman:   "not installed",
			expected: "true",
		},
		{
			name:     "Podman only",
			docker:   "not installed",
			podman:   "installed",
			expected: "true",
		},
		{
			name:     "Both Docker and Podman",
			docker:   "installed",
			podman:   "installed",
			expected: "true",
		},
	}

	tf.Info(t, "Expected behavior matrix:")
	for _, s := range scenarios {
		tf.Info(t, s.name,
			"Docker: "+s.docker,
			"Podman: "+s.podman,
			"Container Available: "+s.expected)
	}

	tf.Success(t, "All expected scenarios documented")

	// Test 8: Verify actual container runtime functionality (if available)
	tf.Separator()
	tf.Step(t, "Verify actual container runtime availability")

	// Check if Docker is actually available
	dockerCmd := exec.Command("docker", "version")
	dockerOutput, dockerErr := dockerCmd.CombinedOutput()
	dockerActuallyAvailable := dockerErr == nil && strings.Contains(string(dockerOutput), "Docker")

	// Check if Podman is actually available
	podmanCmd := exec.Command("podman", "version")
	podmanOutput, podmanErr := podmanCmd.CombinedOutput()
	podmanActuallyAvailable := podmanErr == nil && strings.Contains(string(podmanOutput), "podman")

	// Compare with what Portunix detected
	if dockerActuallyAvailable == dockerInstalled {
		tf.Success(t, "✓ Docker detection matches actual availability")
	} else {
		tf.Warning(t, "Docker detection mismatch",
			"Portunix detected: "+boolToStatus(dockerInstalled),
			"Actually available: "+boolToStatus(dockerActuallyAvailable))
	}

	if podmanActuallyAvailable == podmanInstalled {
		tf.Success(t, "✓ Podman detection matches actual availability")
	} else {
		tf.Warning(t, "Podman detection mismatch",
			"Portunix detected: "+boolToStatus(podmanInstalled),
			"Actually available: "+boolToStatus(podmanActuallyAvailable))
	}
}

func boolToStatus(b bool) string {
	if b {
		return "installed"
	}
	return "not installed"
}