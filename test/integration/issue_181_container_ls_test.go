package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue181_ContainerLsAlias tests that 'ls' and 'ps' subcommands are
// recognized as aliases of 'list' (matching Docker/Podman conventions).
// Issue: Container 'ls' Subcommand Not Recognized (alias for 'list')
// https://github.com/cassandragargoyle/portunix/issues/181
func TestIssue181_ContainerLsAlias(t *testing.T) {
	tf := testframework.NewTestFramework("Issue181_Container_LS_Alias")
	tf.Start(t, "Test that container ls and ps subcommands route to list handler")

	success := true
	defer tf.Finish(t, success)

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

	// TC001: container help advertises aliases
	tf.Step(t, "TC001: 'container --help' advertises ls/ps aliases")
	cmd := exec.Command(binaryPath, "container", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Failed to execute container --help", err.Error())
		success = false
		return
	}
	outputStr := string(output)
	tf.Output(t, outputStr, 800)

	if !strings.Contains(outputStr, "aliases: ls, ps") {
		tf.Error(t, "Help text does not advertise ls/ps aliases")
		success = false
	} else {
		tf.Success(t, "Help text advertises aliases")
	}
	tf.Separator()

	// TC002: Create test container so list output has at least one entry
	tf.Step(t, "TC002: Create test container for listing")
	// Keep the random suffix short — list output truncates long names with
	// an ellipsis, so we match against a prefix that survives truncation.
	containerName := "ptx181-" + time.Now().Format("150405")
	containerNamePrefix := "ptx181-"

	cmd = exec.Command(binaryPath, "container", "run", "-d", "--name", containerName, "ubuntu:22.04", "sleep", "300")
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Failed to create test container", err.Error(), string(output))
		success = false
		return
	}
	tf.Success(t, "Test container created:", containerName)
	tf.Separator()

	defer func() {
		tf.Step(t, "Cleanup: Force remove test container")
		cleanupCmd := exec.Command(binaryPath, "container", "rm", "-f", containerName)
		cleanupCmd.Run()
		tf.Success(t, "Cleanup completed")
	}()

	time.Sleep(2 * time.Second)

	// TC003: 'container ls' is recognized
	tf.Step(t, "TC003: 'container ls' is recognized (no Unknown subcommand)")
	cmd = exec.Command(binaryPath, "container", "ls")
	output, err = cmd.CombinedOutput()
	outputStr = string(output)
	tf.Output(t, outputStr, 500)

	if err != nil {
		tf.Error(t, "container ls failed", err.Error())
		success = false
	} else if strings.Contains(outputStr, "Unknown container subcommand") {
		tf.Error(t, "BUG: ls subcommand not recognized", outputStr)
		success = false
	} else if !strings.Contains(outputStr, containerNamePrefix) {
		tf.Error(t, "Test container missing from ls output", containerName)
		success = false
	} else {
		tf.Success(t, "container ls recognized and lists test container")
	}
	tf.Separator()

	// TC004: 'container ps' is recognized
	tf.Step(t, "TC004: 'container ps' is recognized (no Unknown subcommand)")
	cmd = exec.Command(binaryPath, "container", "ps")
	output, err = cmd.CombinedOutput()
	outputStr = string(output)
	tf.Output(t, outputStr, 500)

	if err != nil {
		tf.Error(t, "container ps failed", err.Error())
		success = false
	} else if strings.Contains(outputStr, "Unknown container subcommand") {
		tf.Error(t, "BUG: ps subcommand not recognized", outputStr)
		success = false
	} else if !strings.Contains(outputStr, containerNamePrefix) {
		tf.Error(t, "Test container missing from ps output", containerName)
		success = false
	} else {
		tf.Success(t, "container ps recognized and lists test container")
	}
	tf.Separator()

	// TC005: 'container list' regression — still works
	tf.Step(t, "TC005: 'container list' still works (no regression)")
	cmd = exec.Command(binaryPath, "container", "list")
	output, err = cmd.CombinedOutput()
	outputStr = string(output)
	tf.Output(t, outputStr, 500)

	if err != nil {
		tf.Error(t, "container list failed", err.Error())
		success = false
	} else if strings.Contains(outputStr, "Unknown container subcommand") {
		tf.Error(t, "BUG: list subcommand broken", outputStr)
		success = false
	} else if !strings.Contains(outputStr, containerNamePrefix) {
		tf.Error(t, "Test container missing from list output", containerName)
		success = false
	} else {
		tf.Success(t, "container list works correctly")
	}
	tf.Separator()

	// TC006: docker/podman dispatched paths share the same fix
	tf.Step(t, "TC006: 'docker ls' is recognized (shared dispatcher)")
	cmd = exec.Command(binaryPath, "docker", "ls")
	output, _ = cmd.CombinedOutput()
	outputStr = string(output)

	if strings.Contains(outputStr, "Unknown docker subcommand") {
		tf.Error(t, "BUG: docker ls not recognized", outputStr)
		success = false
	} else {
		tf.Success(t, "docker ls recognized")
	}
	tf.Separator()

	tf.Step(t, "TC007: 'podman ps' is recognized (shared dispatcher)")
	cmd = exec.Command(binaryPath, "podman", "ps")
	output, _ = cmd.CombinedOutput()
	outputStr = string(output)

	if strings.Contains(outputStr, "Unknown podman subcommand") {
		tf.Error(t, "BUG: podman ps not recognized", outputStr)
		success = false
	} else {
		tf.Success(t, "podman ps recognized")
	}
}
