package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/test/testframework"
)

func TestIssue049_UniversalVirtCommands(t *testing.T) {
	tf := testframework.NewTestFramework("Issue049_Universal_Virt")
	tf.Start(t, "Test universal virtualization commands implementation")

	success := true
	defer tf.Finish(t, success)

	// Get binary path
	binaryPath, err := filepath.Abs("../../portunix")
	if err != nil {
		tf.Error(t, "Failed to get binary path", err.Error())
		success = false
		return
	}

	tf.Step(t, "Setup test binary")
	tf.Info(t, "Binary path", binaryPath)

	// Verify binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		tf.Error(t, "Binary not found", binaryPath)
		success = false
		return
	}

	tf.Success(t, "Binary found")

	// Run test phases
	if !testUniversalInstallation(t, tf, binaryPath) {
		success = false
	}

	tf.Separator()

	if !testUniversalCommands(t, tf, binaryPath) {
		success = false
	}

	tf.Separator()

	if !testISOManagement(t, tf, binaryPath) {
		success = false
	}

	tf.Separator()

	if !testContainerIsolation(t, tf, binaryPath) {
		success = false
	}
}

func testUniversalInstallation(t *testing.T, tf *testframework.TestFramework, binaryPath string) bool {
	tf.Step(t, "Test universal installation interface")

	// Test that vm install-qemu is removed (should fail)
	tf.Step(t, "Verify vm install-qemu command removed")
	tf.Command(t, binaryPath, []string{"vm", "install-qemu", "--help"})

	cmd := exec.Command(binaryPath, "vm", "install-qemu", "--help")
	output, err := cmd.CombinedOutput()

	if err == nil {
		tf.Error(t, "CRITICAL: vm install-qemu command still exists - should be removed")
		return false
	}

	tf.Success(t, "vm install-qemu command correctly removed")

	// Test install virt dry-run
	tf.Step(t, "Test install virt dry-run")
	tf.Command(t, binaryPath, []string{"install", "virt", "--dry-run"})

	cmd = exec.Command(binaryPath, "install", "virt", "--dry-run")
	output, err = cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "install virt --dry-run failed", err.Error())
		return false
	}

	tf.Output(t, string(output), 300)

	// Should show QEMU on Linux
	if !strings.Contains(string(output), "qemu") && !strings.Contains(string(output), "QEMU") {
		tf.Error(t, "Dry-run should mention QEMU on Linux")
		return false
	}

	tf.Success(t, "install virt dry-run works correctly")

	// Test install qemu dry-run
	tf.Step(t, "Test install qemu dry-run")
	tf.Command(t, binaryPath, []string{"install", "qemu", "--dry-run"})

	cmd = exec.Command(binaryPath, "install", "qemu", "--dry-run")
	output, err = cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "install qemu --dry-run failed", err.Error())
		return false
	}

	tf.Success(t, "install qemu dry-run works correctly")

	return true
}

func testUniversalCommands(t *testing.T, tf *testframework.TestFramework, binaryPath string) bool {
	tf.Step(t, "Test universal virt commands")

	// Test virt help
	tf.Step(t, "Test virt command help")
	tf.Command(t, binaryPath, []string{"virt", "--help"})

	cmd := exec.Command(binaryPath, "virt", "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "virt --help failed", err.Error())
		return false
	}

	tf.Output(t, string(output), 400)

	// Verify essential commands are present
	requiredCommands := []string{"create", "start", "stop", "list", "ssh", "snapshot", "iso"}
	for _, cmd := range requiredCommands {
		if !strings.Contains(string(output), cmd) {
			tf.Error(t, fmt.Sprintf("Missing command in help: %s", cmd))
			return false
		}
	}

	tf.Success(t, "All required commands found in help")

	// Test virt list (should work even without VMs)
	tf.Step(t, "Test virt list command")
	tf.Command(t, binaryPath, []string{"virt", "list"})

	cmd = exec.Command(binaryPath, "virt", "list")
	output, err = cmd.CombinedOutput()

	// This might fail if no backend is installed, which is OK for this test
	tf.Output(t, string(output), 200)

	if err != nil {
		if strings.Contains(string(output), "backend") || strings.Contains(string(output), "virtualization") {
			tf.Info(t, "virt list failed as expected (no backend installed)")
		} else {
			tf.Warning(t, "Unexpected virt list failure", err.Error())
		}
	} else {
		tf.Success(t, "virt list executed successfully")
	}

	// Test invalid VM operations (should show proper errors)
	tf.Step(t, "Test error handling for non-existent VM")
	tf.Command(t, binaryPath, []string{"virt", "start", "nonexistent-vm"})

	cmd = exec.Command(binaryPath, "virt", "start", "nonexistent-vm")
	output, err = cmd.CombinedOutput()

	if err == nil {
		tf.Error(t, "start nonexistent-vm should fail")
		return false
	}

	// Should contain meaningful error message
	if strings.Contains(string(output), "not found") || strings.Contains(string(output), "does not exist") {
		tf.Success(t, "Proper error message for non-existent VM")
	} else {
		tf.Warning(t, "Error message could be more descriptive")
	}

	return true
}

func testISOManagement(t *testing.T, tf *testframework.TestFramework, binaryPath string) bool {
	tf.Step(t, "Test ISO management commands")

	// Test ISO help
	tf.Step(t, "Test virt iso help")
	tf.Command(t, binaryPath, []string{"virt", "iso", "--help"})

	cmd := exec.Command(binaryPath, "virt", "iso", "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "virt iso --help failed", err.Error())
		return false
	}

	tf.Output(t, string(output), 300)

	// Verify ISO commands are present
	isoCommands := []string{"download", "list", "verify"}
	for _, cmd := range isoCommands {
		if !strings.Contains(string(output), cmd) {
			tf.Error(t, fmt.Sprintf("Missing ISO command: %s", cmd))
			return false
		}
	}

	tf.Success(t, "All ISO commands found in help")

	// Test ISO list (should work even if empty)
	tf.Step(t, "Test virt iso list")
	tf.Command(t, binaryPath, []string{"virt", "iso", "list"})

	cmd = exec.Command(binaryPath, "virt", "iso", "list")
	output, err = cmd.CombinedOutput()

	tf.Output(t, string(output), 200)

	if err != nil {
		tf.Warning(t, "ISO list failed", err.Error())
	} else {
		tf.Success(t, "ISO list executed successfully")
	}

	// Test ISO download dry-run
	tf.Step(t, "Test ISO download dry-run")
	tf.Command(t, binaryPath, []string{"virt", "iso", "download", "ubuntu-24.04", "--dry-run"})

	cmd = exec.Command(binaryPath, "virt", "iso", "download", "ubuntu-24.04", "--dry-run")
	output, err = cmd.CombinedOutput()

	tf.Output(t, string(output), 300)

	if err != nil {
		tf.Warning(t, "ISO download dry-run failed", err.Error())
	} else {
		// Should show preview information
		if strings.Contains(string(output), "would") || strings.Contains(string(output), "preview") || strings.Contains(string(output), "download") {
			tf.Success(t, "ISO download dry-run shows preview")
		} else {
			tf.Warning(t, "Dry-run output could be more descriptive")
		}
	}

	return true
}

func testContainerIsolation(t *testing.T, tf *testframework.TestFramework, binaryPath string) bool {
	tf.Step(t, "Test container-based installation isolation")

	// Check if Docker/Podman is available
	tf.Step(t, "Check container runtime availability")

	var containerRuntime string
	for _, runtime := range []string{"docker", "podman"} {
		cmd := exec.Command("which", runtime)
		if cmd.Run() == nil {
			containerRuntime = runtime
			break
		}
	}

	if containerRuntime == "" {
		tf.Warning(t, "No container runtime available, skipping container tests")
		return true
	}

	tf.Info(t, "Using container runtime", containerRuntime)

	// Create test container name
	containerName := fmt.Sprintf("portunix-test-%d", time.Now().Unix())

	tf.Step(t, "Create test container")
	tf.Command(t, containerRuntime, []string{"run", "-d", "--name", containerName, "ubuntu:22.04", "sleep", "3600"})

	cmd := exec.Command(containerRuntime, "run", "-d", "--name", containerName, "ubuntu:22.04", "sleep", "3600")
	output, err := cmd.Output()

	if err != nil {
		tf.Warning(t, "Failed to create test container", err.Error())
		return true // Not critical failure
	}

	containerID := strings.TrimSpace(string(output))
	tf.Success(t, fmt.Sprintf("Container created: %s", containerID[:12]))

	// Cleanup container when done
	defer func() {
		tf.Info(t, "Cleaning up test container")
		stopCmd := exec.Command(containerRuntime, "stop", containerName)
		stopCmd.Run()
		rmCmd := exec.Command(containerRuntime, "rm", containerName)
		rmCmd.Run()
	}()

	// Copy binary to container
	tf.Step(t, "Copy binary to container")
	tf.Command(t, containerRuntime, []string{"cp", binaryPath, containerName + ":/usr/local/bin/portunix"})

	cmd = exec.Command(containerRuntime, "cp", binaryPath, containerName+":/usr/local/bin/portunix")
	if err := cmd.Run(); err != nil {
		tf.Error(t, "Failed to copy binary to container", err.Error())
		return false
	}

	// Make executable
	cmd = exec.Command(containerRuntime, "exec", containerName, "chmod", "+x", "/usr/local/bin/portunix")
	if err := cmd.Run(); err != nil {
		tf.Error(t, "Failed to make binary executable", err.Error())
		return false
	}

	tf.Success(t, "Binary copied to container")

	// Test portunix in container
	tf.Step(t, "Test portunix help in container")
	tf.Command(t, containerRuntime, []string{"exec", containerName, "portunix", "--help"})

	cmd = exec.Command(containerRuntime, "exec", containerName, "portunix", "--help")
	output, err = cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "portunix help failed in container", err.Error())
		return false
	}

	tf.Success(t, "portunix help works in container")

	// Test virt commands in container
	tf.Step(t, "Test virt commands in container")
	tf.Command(t, containerRuntime, []string{"exec", containerName, "portunix", "virt", "--help"})

	cmd = exec.Command(containerRuntime, "exec", containerName, "portunix", "virt", "--help")
	output, err = cmd.CombinedOutput()

	if err != nil {
		tf.Warning(t, "virt help failed in container (expected if no backend)", err.Error())
	} else {
		tf.Success(t, "virt help works in container")
	}

	// Verify host system is not contaminated
	tf.Step(t, "Verify host system is clean")

	// Check that QEMU is not installed on host
	cmd = exec.Command("which", "qemu-system-x86_64")
	if err := cmd.Run(); err == nil {
		tf.Warning(t, "QEMU found on host system - might be pre-existing")
	} else {
		tf.Success(t, "Host system clean - no QEMU found")
	}

	// Check that virtualization cache is clean
	cacheDir := filepath.Join(os.Getenv("HOME"), ".portunix", "cache", "isos")
	if _, err := os.Stat(cacheDir); err == nil {
		tf.Info(t, "ISO cache directory exists (might be from previous tests)")
	} else {
		tf.Success(t, "No ISO cache directory on host")
	}

	return true
}

// Helper function for container tests
func execInContainer(containerRuntime, containerName string, command []string) (string, error) {
	args := append([]string{"exec", containerName}, command...)
	cmd := exec.Command(containerRuntime, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestIssue049_InstallationMatrix(t *testing.T) {
	tf := testframework.NewTestFramework("Issue049_Installation_Matrix")
	tf.Start(t, "Test installation command matrix for virt packages")

	success := true
	defer tf.Finish(t, success)

	binaryPath, err := filepath.Abs("../../portunix")
	if err != nil {
		tf.Error(t, "Failed to get binary path", err.Error())
		success = false
		return
	}

	testCases := []struct {
		name        string
		command     []string
		expectError bool
		description string
	}{
		{
			name:        "INST001",
			command:     []string{"install", "virt", "--dry-run"},
			expectError: false,
			description: "Universal virt installation dry-run",
		},
		{
			name:        "INST002",
			command:     []string{"install", "qemu", "--dry-run"},
			expectError: false,
			description: "Explicit QEMU installation dry-run",
		},
		{
			name:        "INST003",
			command:     []string{"install", "virtualbox", "--dry-run"},
			expectError: false,
			description: "Explicit VirtualBox installation dry-run",
		},
		{
			name:        "INST004",
			command:     []string{"vm", "install-qemu"},
			expectError: true,
			description: "Old vm install-qemu should be removed",
		},
	}

	for _, tc := range testCases {
		tf.Step(t, fmt.Sprintf("Test %s: %s", tc.name, tc.description))
		tf.Command(t, binaryPath, tc.command)

		cmd := exec.Command(binaryPath, tc.command...)
		output, err := cmd.CombinedOutput()

		tf.Output(t, string(output), 200)

		if tc.expectError {
			if err == nil {
				tf.Error(t, fmt.Sprintf("%s should fail but succeeded", tc.name))
				success = false
			} else {
				tf.Success(t, fmt.Sprintf("%s correctly failed", tc.name))
			}
		} else {
			if err != nil {
				tf.Error(t, fmt.Sprintf("%s failed: %v", tc.name, err))
				success = false
			} else {
				tf.Success(t, fmt.Sprintf("%s succeeded", tc.name))
			}
		}

		tf.Separator()
	}
}

func TestIssue049_CommandStructure(t *testing.T) {
	tf := testframework.NewTestFramework("Issue049_Command_Structure")
	tf.Start(t, "Test virt command structure and subcommands")

	success := true
	defer tf.Finish(t, success)

	binaryPath, err := filepath.Abs("../../portunix")
	if err != nil {
		tf.Error(t, "Failed to get binary path", err.Error())
		success = false
		return
	}

	// Test main virt command
	tf.Step(t, "Test main virt command structure")
	tf.Command(t, binaryPath, []string{"virt", "--help"})

	cmd := exec.Command(binaryPath, "virt", "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "virt --help failed", err.Error())
		success = false
		return
	}

	tf.Output(t, string(output), 500)

	// Check for required subcommands
	requiredSubcommands := []string{
		"create", "start", "stop", "restart", "suspend", "resume", "delete",
		"list", "info", "status", "ssh", "snapshot", "iso",
	}

	for _, subcmd := range requiredSubcommands {
		if !strings.Contains(string(output), subcmd) {
			tf.Error(t, fmt.Sprintf("Missing subcommand: %s", subcmd))
			success = false
		}
	}

	if success {
		tf.Success(t, "All required subcommands found")
	}

	// Test subcommand help
	tf.Step(t, "Test subcommand help availability")

	subcommandTests := []string{"create", "ssh", "snapshot", "iso"}
	for _, subcmd := range subcommandTests {
		tf.Step(t, fmt.Sprintf("Test virt %s --help", subcmd))
		tf.Command(t, binaryPath, []string{"virt", subcmd, "--help"})

		cmd := exec.Command(binaryPath, "virt", subcmd, "--help")
		output, err := cmd.CombinedOutput()

		if err != nil {
			tf.Warning(t, fmt.Sprintf("virt %s --help failed", subcmd), err.Error())
		} else {
			tf.Success(t, fmt.Sprintf("virt %s help available", subcmd))
		}

		tf.Output(t, string(output), 150)
	}
}