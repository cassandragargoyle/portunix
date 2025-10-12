package integration

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/test/testframework"
)

// Test Requirements:
// Use Portunix container functionality to create containers with mock Docker/Podman binaries
// to test detection logic without actually installing container runtimes.
//
// TODO: This test will be refactored once Portunix has infrastructure with computers/VMs
// where Docker and Podman can be easily installed and uninstalled for real testing.
// Current approach uses mock binaries to simulate different runtime scenarios.

func TestIssue048ContainerDetection(t *testing.T) {
	tf := testframework.NewTestFramework("Issue048_Container_Detection")
	tf.Start(t, "Test enhanced container detection in system info using mock container runtimes")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	binaryPath := "../../portunix"

	// Mock scripts for simulating Docker and Podman
	dockerMockScript := `#!/bin/sh
echo "Docker version 24.0.0, build abcdef"
exit 0`

	podmanMockScript := `#!/bin/sh
echo "podman version 4.5.0"
exit 0`

	// Test scenarios using mock binaries
	testScenarios := []struct {
		name        string
		setupCmds   []string
		expected    map[string]string
	}{
		{
			name: "No container runtime",
			setupCmds: []string{
				// No mock binaries installed
			},
			expected: map[string]string{
				"Docker:":              "not installed",
				"Podman:":              "not installed",
				"Container Available:": "false",
			},
		},
		{
			name: "Docker only",
			setupCmds: []string{
				fmt.Sprintf("echo '%s' > /usr/local/bin/docker && chmod +x /usr/local/bin/docker", dockerMockScript),
			},
			expected: map[string]string{
				"Docker:":              "installed",
				"Podman:":              "not installed",
				"Container Available:": "true",
			},
		},
		{
			name: "Podman only",
			setupCmds: []string{
				fmt.Sprintf("echo '%s' > /usr/local/bin/podman && chmod +x /usr/local/bin/podman", podmanMockScript),
			},
			expected: map[string]string{
				"Docker:":              "not installed",
				"Podman:":              "installed",
				"Container Available:": "true",
			},
		},
		{
			name: "Both Docker and Podman",
			setupCmds: []string{
				fmt.Sprintf("echo '%s' > /usr/local/bin/docker && chmod +x /usr/local/bin/docker", dockerMockScript),
				fmt.Sprintf("echo '%s' > /usr/local/bin/podman && chmod +x /usr/local/bin/podman", podmanMockScript),
			},
			expected: map[string]string{
				"Docker:":              "installed",
				"Podman:":              "installed",
				"Container Available:": "true",
			},
		},
	}

	// Check which container runtime is available on host for creating test containers
	var containerRuntime string
	if _, err := exec.Command("docker", "version").CombinedOutput(); err == nil {
		containerRuntime = "docker"
		tf.Info(t, "Using Docker for creating test containers")
	} else if _, err := exec.Command("podman", "version").CombinedOutput(); err == nil {
		containerRuntime = "podman"
		tf.Info(t, "Using Podman for creating test containers")
	} else {
		tf.Warning(t, "No container runtime on host, testing only host system info")
		// Still test host system info even without containers
		testHostSystemInfo(t, tf, binaryPath, &success)
		return
	}

	for i, scenario := range testScenarios {
		tf.Separator()
		tf.Step(t, fmt.Sprintf("Test scenario %d: %s", i+1, scenario.name))

		// Create container using Portunix run-in-container with empty type
		containerName := fmt.Sprintf("test-issue048-%d", time.Now().Unix())
		tf.Info(t, "Creating test container", fmt.Sprintf("Name: %s", containerName))

		// Use run-in-container with empty type to get minimal setup
		createCmd := exec.Command(binaryPath, containerRuntime, "run-in-container", "empty",
			"--image", "ubuntu:22.04",
			"--name", containerName,
			"--keep-running")

		tf.Command(t, binaryPath, []string{containerRuntime, "run-in-container", "empty",
			"--image", "ubuntu:22.04",
			"--name", containerName,
			"--keep-running"})

		output, err := createCmd.CombinedOutput()
		if err != nil {
			tf.Error(t, "Failed to create container", err.Error())
			tf.Output(t, string(output), 500)
			success = false
			continue
		}
		tf.Success(t, "Container created successfully")

		// Wait for container to be ready
		time.Sleep(3 * time.Second)

		// Clean any existing mock binaries first
		cleanupCmd := fmt.Sprintf("rm -f /usr/local/bin/docker /usr/local/bin/podman")
		execCleanCmd := exec.Command(binaryPath, containerRuntime, "exec", containerName, "sh", "-c", cleanupCmd)
		execCleanCmd.Run()

		// Setup mock binaries according to scenario
		for _, setupCmd := range scenario.setupCmds {
			tf.Info(t, "Setting up mock binary")

			execCmd := exec.Command(binaryPath, containerRuntime, "exec", containerName, "sh", "-c", setupCmd)
			output, err := execCmd.CombinedOutput()
			if err != nil {
				tf.Warning(t, "Failed to setup mock", err.Error(), string(output))
			}
		}

		// Test system info inside container
		tf.Info(t, "Testing system info inside container")
		sysInfoCmd := exec.Command(binaryPath, containerRuntime, "exec", containerName, "portunix", "system", "info")
		tf.Command(t, binaryPath, []string{containerRuntime, "exec", containerName, "portunix", "system", "info"})

		output, err = sysInfoCmd.CombinedOutput()
		if err != nil {
			tf.Error(t, "Failed to get system info", err.Error())
			tf.Output(t, string(output), 500)
			success = false

			// Cleanup before continuing
			cleanupCmd := exec.Command(binaryPath, containerRuntime, "remove", containerName)
			cleanupCmd.Run()
			continue
		}

		outputStr := string(output)
		tf.Output(t, outputStr, 1000)

		// Verify expected values
		tf.Info(t, "Verifying container runtime detection")
		allChecksPass := true
		for key, expectedValue := range scenario.expected {
			// Check for the key and value in output
			if verifyOutput(outputStr, key, expectedValue) {
				tf.Success(t, fmt.Sprintf("✓ %s %s", key, expectedValue))
			} else {
				tf.Error(t, fmt.Sprintf("✗ Expected '%s %s' not found", key, expectedValue))
				allChecksPass = false
				success = false
			}
		}

		if allChecksPass {
			tf.Success(t, fmt.Sprintf("Scenario '%s' passed all checks", scenario.name))
		} else {
			tf.Warning(t, fmt.Sprintf("Scenario '%s' had failures", scenario.name))
		}

		// Cleanup container
		tf.Info(t, "Cleaning up container")
		removeCmd := exec.Command(binaryPath, containerRuntime, "remove", containerName)
		removeCmd.Run()

		// Small delay between scenarios
		time.Sleep(1 * time.Second)
	}

	// Test on host system
	testHostSystemInfo(t, tf, binaryPath, &success)
}

func testHostSystemInfo(t *testing.T, tf *testframework.TestFramework, binaryPath string, success *bool) {
	tf.Separator()
	tf.Step(t, "Test system info on host")

	hostCmd := exec.Command(binaryPath, "system", "info")
	tf.Command(t, binaryPath, []string{"system", "info"})

	output, err := hostCmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Failed to get host system info", err.Error())
		*success = false
		return
	}

	outputStr := string(output)
	tf.Output(t, outputStr, 800)

	// Verify Container Runtimes section exists
	if !strings.Contains(outputStr, "Container Runtimes:") {
		tf.Error(t, "Container Runtimes section not found")
		*success = false
		return
	}
	tf.Success(t, "Container Runtimes section found")

	// Check for Container Available status
	if !strings.Contains(outputStr, "Container Available:") {
		tf.Error(t, "Container Available status not found")
		*success = false
		return
	}
	tf.Success(t, "Container Available status found")

	// Verify logic: if either Docker or Podman is installed, Container Available should be true
	dockerInstalled := strings.Contains(outputStr, "Docker:       installed")
	podmanInstalled := strings.Contains(outputStr, "Podman:       installed")
	containerAvailable := strings.Contains(outputStr, "Container Available: true")

	expectedAvailable := dockerInstalled || podmanInstalled

	if expectedAvailable == containerAvailable {
		tf.Success(t, fmt.Sprintf("Container Available flag logic is correct (Docker=%v, Podman=%v, Available=%v)",
			dockerInstalled, podmanInstalled, containerAvailable))
	} else {
		tf.Error(t, fmt.Sprintf("Container Available logic error: Docker=%v, Podman=%v, Available=%v, Expected=%v",
			dockerInstalled, podmanInstalled, containerAvailable, expectedAvailable))
		*success = false
	}
}

func verifyOutput(output, key, value string) bool {
	// Handle different formatting possibilities
	patterns := []string{
		fmt.Sprintf("%s %s", key, value),           // "Docker: installed"
		fmt.Sprintf("%s%s", key, value),            // "Docker:installed"
		fmt.Sprintf("%s       %s", key, value),     // "Docker:       installed" (aligned)
		fmt.Sprintf("%s      %s", key, value),      // variations in spacing
		fmt.Sprintf("%s     %s", key, value),
		fmt.Sprintf("%s    %s", key, value),
		fmt.Sprintf("%s   %s", key, value),
		fmt.Sprintf("%s  %s", key, value),
		fmt.Sprintf("%s %s", key, value),
	}

	for _, pattern := range patterns {
		if strings.Contains(output, pattern) {
			return true
		}
	}

	return false
}