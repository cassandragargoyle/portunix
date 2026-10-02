/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package integration

import (
	"os/exec"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

// Issue #193: `aiops ollama container status` must detect a pre-existing
// container named `ollama` (external, not managed by Portunix) as a fallback
// when the managed `portunix-ollama` container does not exist.

const (
	issue193ManagedName  = "portunix-ollama"
	issue193ExternalName = "ollama"
	// Lightweight image; status detection only inspects name/state/image and
	// does not require a real Ollama runtime.
	issue193TestImage = "alpine:latest"
)

// detectRuntimeIssue193 returns the available container runtime, mirroring the
// runtime detection order used by ptx-aiops (docker first, then podman).
func detectRuntimeIssue193() string {
	for _, rt := range []string{"docker", "podman"} {
		if _, err := exec.LookPath(rt); err != nil {
			continue
		}
		if err := exec.Command(rt, "info").Run(); err == nil {
			return rt
		}
	}
	return ""
}

// containerExistsIssue193 reports whether a container with the given name exists.
func containerExistsIssue193(runtime, name string) bool {
	err := exec.Command(runtime, "inspect", name, "--format", "{{.State.Running}}").Run()
	return err == nil
}

// containerRunningIssue193 reports whether the named container is running.
func containerRunningIssue193(runtime, name string) bool {
	out, err := exec.Command(runtime, "inspect", name, "--format", "{{.State.Running}}").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

// createTestContainerIssue193 starts a detached, long-running container with the
// given name so that the status command can inspect it.
func createTestContainerIssue193(runtime, name string) error {
	// Remove any leftover from a previous run of this test.
	exec.Command(runtime, "rm", "-f", name).Run()
	cmd := exec.Command(runtime, "run", "-d", "--name", name, issue193TestImage, "sleep", "3600")
	return cmd.Run()
}

// removeTestContainerIssue193 force-removes a container created by the test.
func removeTestContainerIssue193(runtime, name string) {
	exec.Command(runtime, "rm", "-f", name).Run()
}

// TestIssue193_OllamaStatusExternalContainer verifies the three detection states:
// external `ollama` only, managed `portunix-ollama` only, and neither.
func TestIssue193_OllamaStatusExternalContainer(t *testing.T) {
	tf := testframework.NewTestFramework("Issue193_Ollama_Status_External")
	tf.Start(t, "Detect pre-existing external 'ollama' container in aiops status")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	binaryPath, ok := tf.VerifyPortunixBinary(t)
	if !ok {
		success = false
		return
	}

	runtime := detectRuntimeIssue193()
	if runtime == "" {
		tf.Warning(t, "No container runtime (docker/podman) available — skipping")
		t.Skip("No container runtime available")
		return
	}
	tf.Info(t, "Using container runtime: "+runtime)

	// Safety guard: never touch pre-existing real containers with these names.
	if containerExistsIssue193(runtime, issue193ExternalName) ||
		containerExistsIssue193(runtime, issue193ManagedName) {
		tf.Warning(t, "A container named 'ollama' or 'portunix-ollama' already exists — skipping to avoid clobbering it")
		t.Skip("Pre-existing ollama/portunix-ollama container present")
		return
	}

	// Ensure a clean slate on exit.
	defer removeTestContainerIssue193(runtime, issue193ExternalName)
	defer removeTestContainerIssue193(runtime, issue193ManagedName)

	statusOutput := func() string {
		cmd := exec.Command(binaryPath, "aiops", "ollama", "container", "status")
		out, _ := cmd.CombinedOutput()
		return string(out)
	}

	// TC003 (neither present): unchanged "Not found" behaviour.
	tf.Separator()
	tf.Step(t, "TC003: neither container present -> Not found")
	out := statusOutput()
	tf.Output(t, out, 600)
	if strings.Contains(out, "Not found") &&
		strings.Contains(out, "portunix aiops ollama container create") {
		tf.Success(t, "Reports 'Not found' with create hint")
	} else {
		tf.Error(t, "Expected 'Not found' + create hint when no container exists")
		success = false
	}

	// TC001 (external only): report external ollama, marked not managed.
	tf.Separator()
	tf.Step(t, "TC001: external 'ollama' present -> reported as external")
	if err := createTestContainerIssue193(runtime, issue193ExternalName); err != nil {
		tf.Error(t, "Failed to create external 'ollama' test container", err.Error())
		success = false
		return
	}
	out = statusOutput()
	tf.Output(t, out, 600)
	if strings.Contains(out, issue193ExternalName) &&
		strings.Contains(out, "external, not managed by Portunix") {
		tf.Success(t, "External 'ollama' reported and marked as external")
	} else {
		tf.Error(t, "Expected external 'ollama' to be reported with 'external, not managed by Portunix'")
		success = false
	}
	if strings.Contains(out, "🟢 Running") {
		tf.Success(t, "External container running state detected")
	} else {
		tf.Warning(t, "Did not detect Running state for external container")
	}
	// AC #6: running external container footer offers valid lifecycle commands
	// (which now fall back to the external container) plus --container for models.
	if strings.Contains(out, "portunix aiops ollama container stop") &&
		strings.Contains(out, "--container "+issue193ExternalName) {
		tf.Success(t, "External running footer offers valid lifecycle + model commands")
	} else {
		tf.Error(t, "External running footer must suggest working lifecycle and --container model commands")
		success = false
	}

	// AC #5: `stop` falls back to the external container.
	tf.Separator()
	tf.Step(t, "TC001b: 'aiops ollama container stop' falls back to external")
	stopCmd := exec.Command(binaryPath, "aiops", "ollama", "container", "stop")
	stopOut, _ := stopCmd.CombinedOutput()
	tf.Output(t, string(stopOut), 400)
	if strings.Contains(string(stopOut), "not managed by Portunix") &&
		strings.Contains(string(stopOut), "stopped") &&
		!containerRunningIssue193(runtime, issue193ExternalName) {
		tf.Success(t, "External container stopped via aiops with external note")
	} else {
		tf.Error(t, "Expected aiops stop to act on external container with note")
		success = false
	}

	// AC #6: stopped external footer suggests the working start command.
	out = statusOutput()
	tf.Output(t, out, 600)
	if strings.Contains(out, "🔴 Stopped") &&
		strings.Contains(out, "portunix aiops ollama container start") {
		tf.Success(t, "External stopped footer suggests working start command")
	} else {
		tf.Error(t, "External stopped footer must suggest 'portunix aiops ollama container start'")
		success = false
	}

	// AC #5: `start` falls back to the external container.
	tf.Separator()
	tf.Step(t, "TC001c: 'aiops ollama container start' falls back to external")
	startCmd := exec.Command(binaryPath, "aiops", "ollama", "container", "start")
	startOut, _ := startCmd.CombinedOutput()
	tf.Output(t, string(startOut), 400)
	if strings.Contains(string(startOut), "not managed by Portunix") &&
		containerRunningIssue193(runtime, issue193ExternalName) {
		tf.Success(t, "External container started via aiops with external note")
	} else {
		tf.Error(t, "Expected aiops start to act on external container with note")
		success = false
	}

	// AC #5: `remove` of an external container requires confirmation; declining
	// preserves the container.
	tf.Separator()
	tf.Step(t, "TC001d: 'aiops ollama container remove' external requires confirmation")
	rmCmd := exec.Command(binaryPath, "aiops", "ollama", "container", "remove")
	rmCmd.Stdin = strings.NewReader("n\n")
	rmOut, _ := rmCmd.CombinedOutput()
	tf.Output(t, string(rmOut), 400)
	if strings.Contains(string(rmOut), "external") &&
		strings.Contains(string(rmOut), "Cancelled") &&
		containerExistsIssue193(runtime, issue193ExternalName) {
		tf.Success(t, "Declining confirmation preserved the external container")
	} else {
		tf.Error(t, "Expected remove to prompt for confirmation and preserve container on decline")
		success = false
	}

	// TC002 (both present): managed takes precedence.
	tf.Separator()
	tf.Step(t, "TC002: both present -> managed 'portunix-ollama' takes precedence")
	if err := createTestContainerIssue193(runtime, issue193ManagedName); err != nil {
		tf.Error(t, "Failed to create managed 'portunix-ollama' test container", err.Error())
		success = false
		return
	}
	out = statusOutput()
	tf.Output(t, out, 600)
	if strings.Contains(out, issue193ManagedName) &&
		!strings.Contains(out, "external, not managed by Portunix") {
		tf.Success(t, "Managed 'portunix-ollama' reported without external marker")
	} else {
		tf.Error(t, "Expected managed 'portunix-ollama' to take precedence without external marker")
		success = false
	}

	// TC002b (managed only): remove external, managed still reported normally.
	tf.Separator()
	tf.Step(t, "TC002b: managed only -> unchanged managed output")
	removeTestContainerIssue193(runtime, issue193ExternalName)
	out = statusOutput()
	tf.Output(t, out, 600)
	if strings.Contains(out, issue193ManagedName) &&
		!strings.Contains(out, "external, not managed by Portunix") {
		tf.Success(t, "Managed container reported normally with no external marker")
	} else {
		tf.Error(t, "Expected unchanged managed output when only 'portunix-ollama' exists")
		success = false
	}
}
