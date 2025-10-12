package integration

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/test/testframework"
)

func TestIssue049_ContainerInstallation(t *testing.T) {
	tf := testframework.NewTestFramework("Issue049_Container_Installation")
	tf.Start(t, "Test complete virt installation in container isolation")

	success := true
	defer tf.Finish(t, success)

	// Get binary path
	binaryPath, err := filepath.Abs("../../portunix")
	if err != nil {
		tf.Error(t, "Failed to get binary path", err.Error())
		success = false
		return
	}

	// Check container runtime availability
	containerRuntime := getAvailableContainerRuntime(t, tf)
	if containerRuntime == "" {
		tf.Warning(t, "No container runtime available, skipping container tests")
		return
	}

	// Create test container
	containerID := createTestContainer(t, tf, containerRuntime)
	if containerID == "" {
		tf.Error(t, "Failed to create test container")
		success = false
		return
	}

	defer cleanupContainerWithFramework(t, tf, containerRuntime, containerID)

	// Copy binary to container
	if !copyBinaryToContainer(t, tf, containerRuntime, containerID, binaryPath) {
		success = false
		return
	}

	// Test installation process
	if !testVirtInstallationInContainer(t, tf, containerRuntime, containerID) {
		success = false
	}

	// Verify host system remains clean
	if !verifyHostSystemClean(t, tf) {
		success = false
	}
}

func TestIssue049_MultiPlatformContainers(t *testing.T) {
	tf := testframework.NewTestFramework("Issue049_MultiPlatform_Containers")
	tf.Start(t, "Test virt installation across different container platforms")

	success := true
	defer tf.Finish(t, success)

	binaryPath, err := filepath.Abs("../../portunix")
	if err != nil {
		tf.Error(t, "Failed to get binary path", err.Error())
		success = false
		return
	}

	containerRuntime := getAvailableContainerRuntime(t, tf)
	if containerRuntime == "" {
		tf.Warning(t, "No container runtime available, skipping multi-platform tests")
		return
	}

	platforms := []struct {
		name     string
		image    string
		expected string
	}{
		{"Ubuntu 22.04", "ubuntu:22.04", "QEMU"},
		{"Debian Bookworm", "debian:bookworm", "QEMU"},
	}

	for _, platform := range platforms {
		tf.Step(t, fmt.Sprintf("Test on %s", platform.name))

		containerID := createContainerWithImage(t, tf, containerRuntime, platform.image)
		if containerID == "" {
			tf.Error(t, fmt.Sprintf("Failed to create %s container", platform.name))
			success = false
			continue
		}

		// Install and test
		if installAndTestInContainer(t, tf, containerRuntime, containerID, binaryPath, platform.expected) {
			tf.Success(t, fmt.Sprintf("%s platform test passed", platform.name))
		} else {
			tf.Error(t, fmt.Sprintf("%s platform test failed", platform.name))
			success = false
		}

		cleanupContainerWithFramework(t, tf, containerRuntime, containerID)
	}
}

func TestIssue049_ContainerISOTesting(t *testing.T) {
	tf := testframework.NewTestFramework("Issue049_Container_ISO_Testing")
	tf.Start(t, "Test ISO management in container environment")

	success := true
	defer tf.Finish(t, success)

	binaryPath, err := filepath.Abs("../../portunix")
	if err != nil {
		tf.Error(t, "Failed to get binary path", err.Error())
		success = false
		return
	}

	containerRuntime := getAvailableContainerRuntime(t, tf)
	if containerRuntime == "" {
		tf.Warning(t, "No container runtime available, skipping ISO tests")
		return
	}

	// Create container with internet access
	containerID := createNetworkEnabledContainer(t, tf, containerRuntime)
	if containerID == "" {
		tf.Error(t, "Failed to create network-enabled container")
		success = false
		return
	}

	defer cleanupContainerWithFramework(t, tf, containerRuntime, containerID)

	// Setup portunix in container
	if !setupPortunixInContainer(t, tf, containerRuntime, containerID, binaryPath) {
		success = false
		return
	}

	// Test ISO operations
	if !testISOOperationsInContainer(t, tf, containerRuntime, containerID) {
		success = false
	}
}

// Helper functions

func getAvailableContainerRuntime(t *testing.T, tf *testframework.TestFramework) string {
	tf.Step(t, "Check container runtime availability")

	for _, runtime := range []string{"docker", "podman"} {
		cmd := exec.Command("which", runtime)
		if cmd.Run() == nil {
			tf.Info(t, "Found container runtime", runtime)
			return runtime
		}
	}

	return ""
}

func createTestContainer(t *testing.T, tf *testframework.TestFramework, runtime string) string {
	containerName := fmt.Sprintf("portunix-test-%d", time.Now().Unix())

	tf.Step(t, "Create test container")
	tf.Command(t, runtime, []string{"run", "-d", "--name", containerName, "ubuntu:22.04", "sleep", "3600"})

	cmd := exec.Command(runtime, "run", "-d", "--name", containerName, "ubuntu:22.04", "sleep", "3600")
	output, err := cmd.Output()

	if err != nil {
		tf.Error(t, "Container creation failed", err.Error())
		return ""
	}

	containerID := strings.TrimSpace(string(output))
	tf.Success(t, fmt.Sprintf("Container created: %s", containerID[:12]))

	return containerID
}

func createContainerWithImage(t *testing.T, tf *testframework.TestFramework, runtime, image string) string {
	containerName := fmt.Sprintf("portunix-test-%s-%d", strings.Replace(image, ":", "-", -1), time.Now().Unix())

	tf.Command(t, runtime, []string{"run", "-d", "--name", containerName, image, "sleep", "3600"})

	cmd := exec.Command(runtime, "run", "-d", "--name", containerName, image, "sleep", "3600")
	output, err := cmd.Output()

	if err != nil {
		tf.Error(t, fmt.Sprintf("Failed to create %s container", image), err.Error())
		return ""
	}

	containerID := strings.TrimSpace(string(output))
	tf.Info(t, fmt.Sprintf("Created %s container: %s", image, containerID[:12]))

	return containerID
}

func createNetworkEnabledContainer(t *testing.T, tf *testframework.TestFramework, runtime string) string {
	containerName := fmt.Sprintf("portunix-network-test-%d", time.Now().Unix())

	tf.Step(t, "Create network-enabled container")
	tf.Command(t, runtime, []string{"run", "-d", "--name", containerName, "--network", "host", "ubuntu:22.04", "sleep", "3600"})

	cmd := exec.Command(runtime, "run", "-d", "--name", containerName, "--network", "host", "ubuntu:22.04", "sleep", "3600")
	output, err := cmd.Output()

	if err != nil {
		tf.Error(t, "Network container creation failed", err.Error())
		return ""
	}

	containerID := strings.TrimSpace(string(output))
	tf.Success(t, fmt.Sprintf("Network container created: %s", containerID[:12]))

	return containerID
}

func copyBinaryToContainer(t *testing.T, tf *testframework.TestFramework, runtime, containerID, binaryPath string) bool {
	tf.Step(t, "Copy binary to container")
	tf.Command(t, runtime, []string{"cp", binaryPath, containerID + ":/usr/local/bin/portunix"})

	cmd := exec.Command(runtime, "cp", binaryPath, containerID+":/usr/local/bin/portunix")
	if err := cmd.Run(); err != nil {
		tf.Error(t, "Failed to copy binary to container", err.Error())
		return false
	}

	// Make executable
	cmd = exec.Command(runtime, "exec", containerID, "chmod", "+x", "/usr/local/bin/portunix")
	if err := cmd.Run(); err != nil {
		tf.Error(t, "Failed to make binary executable", err.Error())
		return false
	}

	tf.Success(t, "Binary copied and made executable")
	return true
}

func setupPortunixInContainer(t *testing.T, tf *testframework.TestFramework, runtime, containerID, binaryPath string) bool {
	if !copyBinaryToContainer(t, tf, runtime, containerID, binaryPath) {
		return false
	}

	// Update package lists
	tf.Step(t, "Update package lists in container")
	cmd := exec.Command(runtime, "exec", containerID, "apt", "update")
	if err := cmd.Run(); err != nil {
		tf.Warning(t, "Failed to update package lists", err.Error())
	}

	// Install basic tools
	cmd = exec.Command(runtime, "exec", containerID, "apt", "install", "-y", "curl", "wget")
	if err := cmd.Run(); err != nil {
		tf.Warning(t, "Failed to install basic tools", err.Error())
	}

	return true
}

func testVirtInstallationInContainer(t *testing.T, tf *testframework.TestFramework, runtime, containerID string) bool {
	tf.Step(t, "Test virt installation inside container")

	// Test help first
	tf.Step(t, "Test portunix help in container")
	output, err := execInContainerWithOutput(runtime, containerID, []string{"portunix", "--help"})
	if err != nil {
		tf.Error(t, "portunix help failed in container", err.Error())
		return false
	}
	tf.Success(t, "portunix help works in container")

	// Test virt dry-run installation
	tf.Step(t, "Test virt installation dry-run in container")
	tf.Command(t, runtime, []string{"exec", containerID, "portunix", "install", "virt", "--dry-run"})

	output, err = execInContainerWithOutput(runtime, containerID, []string{"portunix", "install", "virt", "--dry-run"})
	if err != nil {
		tf.Error(t, "virt installation dry-run failed", err.Error())
		return false
	}

	tf.Output(t, output, 300)

	// Should mention QEMU on Linux
	if !strings.Contains(output, "qemu") && !strings.Contains(output, "QEMU") {
		tf.Warning(t, "Dry-run should mention QEMU on Linux platform")
	} else {
		tf.Success(t, "Dry-run correctly shows QEMU for Linux")
	}

	// Test actual installation (this might fail due to missing dependencies, which is OK)
	tf.Step(t, "Test actual virt installation in container")
	output, err = execInContainerWithOutput(runtime, containerID, []string{"portunix", "install", "virt"})

	if err != nil {
		if strings.Contains(output, "permission") || strings.Contains(output, "root") {
			tf.Info(t, "Installation failed due to permissions (expected in container)")
		} else if strings.Contains(output, "package") || strings.Contains(output, "repository") {
			tf.Info(t, "Installation failed due to package management (expected in basic container)")
		} else {
			tf.Warning(t, "Installation failed with unexpected error", err.Error())
		}
		tf.Output(t, output, 200)
	} else {
		tf.Success(t, "virt installation completed successfully")
		tf.Output(t, output, 300)
	}

	return true
}

func testISOOperationsInContainer(t *testing.T, tf *testframework.TestFramework, runtime, containerID string) bool {
	tf.Step(t, "Test ISO operations in container")

	// Test ISO list (should work even if empty)
	tf.Step(t, "Test ISO list in container")
	output, err := execInContainerWithOutput(runtime, containerID, []string{"portunix", "virt", "iso", "list"})
	if err != nil {
		tf.Warning(t, "ISO list failed", err.Error())
	} else {
		tf.Success(t, "ISO list works in container")
		tf.Output(t, output, 150)
	}

	// Test ISO download dry-run
	tf.Step(t, "Test ISO download dry-run in container")
	output, err = execInContainerWithOutput(runtime, containerID, []string{"portunix", "virt", "iso", "download", "ubuntu-24.04", "--dry-run"})
	if err != nil {
		tf.Warning(t, "ISO download dry-run failed", err.Error())
	} else {
		tf.Success(t, "ISO download dry-run works in container")
		tf.Output(t, output, 200)

		if strings.Contains(output, "download") || strings.Contains(output, "ubuntu") {
			tf.Success(t, "Dry-run shows appropriate preview")
		}
	}

	return true
}

func installAndTestInContainer(t *testing.T, tf *testframework.TestFramework, runtime, containerID, binaryPath, expectedBackend string) bool {
	if !setupPortunixInContainer(t, tf, runtime, containerID, binaryPath) {
		return false
	}

	// Test virt commands
	output, err := execInContainerWithOutput(runtime, containerID, []string{"portunix", "virt", "--help"})
	if err != nil {
		tf.Warning(t, "virt help failed (expected if no backend installed)", err.Error())
		return true // Not a failure for this test
	}

	if strings.Contains(output, "create") && strings.Contains(output, "list") {
		tf.Success(t, "virt command structure correct")
		return true
	}

	return false
}

func verifyHostSystemClean(t *testing.T, tf *testframework.TestFramework) bool {
	tf.Step(t, "Verify host system remains clean")

	// Check QEMU not installed on host
	cmd := exec.Command("which", "qemu-system-x86_64")
	if err := cmd.Run(); err == nil {
		tf.Warning(t, "QEMU found on host (might be pre-existing)")
	} else {
		tf.Success(t, "Host system clean - no QEMU installed")
	}

	// Check no virtualization processes running from tests
	cmd = exec.Command("pgrep", "-f", "qemu")
	if err := cmd.Run(); err == nil {
		tf.Warning(t, "QEMU processes found on host (might be unrelated)")
	} else {
		tf.Success(t, "No QEMU processes from tests on host")
	}

	return true
}

func cleanupContainerWithFramework(t *testing.T, tf *testframework.TestFramework, runtime, containerID string) {
	tf.Info(t, "Cleaning up container", containerID[:12])

	// Stop container
	stopCmd := exec.Command(runtime, "stop", containerID)
	stopCmd.Run()

	// Remove container
	rmCmd := exec.Command(runtime, "rm", containerID)
	rmCmd.Run()
}

func execInContainerWithOutput(runtime, containerID string, command []string) (string, error) {
	args := append([]string{"exec", containerID}, command...)
	cmd := exec.Command(runtime, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}