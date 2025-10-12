package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue094_ContainerRmSubcommand tests container rm subcommand functionality
// Issue: Container 'rm' Subcommand Not Recognized
// https://github.com/cassandragargoyle/Portunix/issues/094
func TestIssue094_ContainerRmSubcommand(t *testing.T) {
	tf := testframework.NewTestFramework("Issue094_Container_RM_Subcommand")
	tf.Start(t, "Test container rm subcommand implementation and recognition")

	success := true
	defer tf.Finish(t, success)

	// Get binary path
	binaryPath := "../../portunix"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		binaryPath = "./portunix"
	}

	tf.Step(t, "Verify binary exists")
	if _, err := os.Stat(binaryPath); err != nil {
		tf.Error(t, "Binary not found", err.Error())
		success = false
		return
	}
	tf.Success(t, "Binary found at:", binaryPath)
	tf.Separator()

	// TC001: Test rm help command
	tf.Step(t, "TC001: Test container rm --help displays correct help")
	cmd := exec.Command(binaryPath, "container", "rm", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Failed to execute rm --help", err.Error())
		success = false
		return
	}

	outputStr := string(output)
	tf.Output(t, outputStr, 500)

	if !strings.Contains(outputStr, "REMOVE CONTAINER") {
		tf.Error(t, "Help text missing expected content")
		success = false
	} else {
		tf.Success(t, "Help text displays correctly")
	}
	tf.Separator()

	// TC002: Create test container
	tf.Step(t, "TC002: Create test container for rm testing")
	containerName := "test-issue-094-rm-" + time.Now().Format("150405")

	cmd = exec.Command(binaryPath, "container", "run", "-d", "--name", containerName, "ubuntu:22.04", "sleep", "300")
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Failed to create test container", err.Error(), string(output))
		success = false
		return
	}
	tf.Success(t, "Test container created:", containerName)
	tf.Separator()

	// Ensure cleanup
	defer func() {
		tf.Step(t, "Cleanup: Force remove test container")
		cleanupCmd := exec.Command(binaryPath, "container", "rm", "-f", containerName)
		cleanupCmd.Run()
		tf.Success(t, "Cleanup completed")
	}()

	// TC003: Test rm command recognition (not "remove")
	tf.Step(t, "TC003: Verify 'rm' subcommand is recognized (not showing 'Unknown subcommand: remove')")
	cmd = exec.Command(binaryPath, "container", "rm", containerName)
	output, err = cmd.CombinedOutput()

	outputStr = string(output)
	tf.Output(t, outputStr, 500)

	// Should fail because container is running (without force flag)
	// But should NOT say "Unknown container subcommand: remove"
	if strings.Contains(outputStr, "Unknown container subcommand") {
		tf.Error(t, "BUG: rm subcommand not recognized", outputStr)
		success = false
	} else if strings.Contains(outputStr, "running or paused containers cannot be removed without force") {
		tf.Success(t, "rm subcommand recognized correctly (expected error for running container)")
	} else {
		tf.Warning(t, "Unexpected output - verify behavior", outputStr)
	}
	tf.Separator()

	// TC004: Test rm with force flag
	tf.Step(t, "TC004: Test container rm with --force flag")
	cmd = exec.Command(binaryPath, "container", "rm", "--force", containerName)
	output, err = cmd.CombinedOutput()

	outputStr = string(output)
	tf.Output(t, outputStr, 500)

	if err != nil {
		tf.Error(t, "Failed to remove container with --force", err.Error())
		success = false
	} else if strings.Contains(outputStr, "removed successfully") {
		tf.Success(t, "Container removed successfully with --force flag")
	} else {
		tf.Warning(t, "Unexpected output", outputStr)
	}
	tf.Separator()

	// TC005: Verify container is removed
	tf.Step(t, "TC005: Verify container was removed from list")
	cmd = exec.Command(binaryPath, "container", "list")
	output, err = cmd.CombinedOutput()

	outputStr = string(output)
	if strings.Contains(outputStr, containerName) {
		tf.Error(t, "Container still present after removal", containerName)
		success = false
	} else {
		tf.Success(t, "Container successfully removed from list")
	}
	tf.Separator()

	// TC006: Test multiple container removal
	tf.Step(t, "TC006: Test removing multiple containers at once")

	// Create two test containers
	container1 := "test-multi-rm-1-" + time.Now().Format("150405")
	container2 := "test-multi-rm-2-" + time.Now().Format("150405")

	cmd = exec.Command(binaryPath, "container", "run", "-d", "--name", container1, "ubuntu:22.04", "sleep", "300")
	cmd.Run()

	cmd = exec.Command(binaryPath, "container", "run", "-d", "--name", container2, "ubuntu:22.04", "sleep", "300")
	cmd.Run()

	// Ensure cleanup
	defer func() {
		cleanupCmd := exec.Command(binaryPath, "container", "rm", "-f", container1, container2)
		cleanupCmd.Run()
	}()

	time.Sleep(2 * time.Second) // Wait for containers to start

	// Remove both at once
	cmd = exec.Command(binaryPath, "container", "rm", "-f", container1, container2)
	output, err = cmd.CombinedOutput()

	outputStr = string(output)
	tf.Output(t, outputStr, 500)

	if strings.Contains(outputStr, "removed successfully") {
		tf.Success(t, "Multiple containers removed successfully")
	} else {
		tf.Warning(t, "Multiple container removal might have issues", outputStr)
	}
	tf.Separator()

	// TC007: Test rm with short force flag (-f)
	tf.Step(t, "TC007: Test container rm with short -f flag")
	containerShort := "test-short-flag-" + time.Now().Format("150405")

	cmd = exec.Command(binaryPath, "container", "run", "-d", "--name", containerShort, "ubuntu:22.04", "sleep", "300")
	cmd.Run()

	defer func() {
		cleanupCmd := exec.Command(binaryPath, "container", "rm", "-f", containerShort)
		cleanupCmd.Run()
	}()

	time.Sleep(2 * time.Second)

	cmd = exec.Command(binaryPath, "container", "rm", "-f", containerShort)
	output, err = cmd.CombinedOutput()

	outputStr = string(output)
	if strings.Contains(outputStr, "removed successfully") {
		tf.Success(t, "Short -f flag works correctly")
	} else {
		tf.Error(t, "Short -f flag failed", outputStr)
		success = false
	}
}
