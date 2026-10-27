//go:build integration
// +build integration

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

// Issue #242: `aiops ollama container status` must report the host port
// actually published for container port 11434/tcp instead of assuming 11434.
// Reuses the runtime helpers from the #193 test.

const issue242HostPort = "21434"

// TestIssue242_OllamaStatusPublishedPort verifies the reported API port for a
// non-default port mapping and for a container without a published port
func TestIssue242_OllamaStatusPublishedPort(t *testing.T) {
	tf := testframework.NewTestFramework("Issue242_Ollama_Status_Port")
	tf.Start(t, "Report the published host port in aiops ollama container status")

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

	// Safety guard: never touch pre-existing real containers with these names
	if containerExistsIssue193(runtime, issue193ExternalName) ||
		containerExistsIssue193(runtime, issue193ManagedName) {
		tf.Warning(t, "A container named 'ollama' or 'portunix-ollama' already exists — skipping to avoid clobbering it")
		t.Skip("Pre-existing ollama/portunix-ollama container present")
		return
	}
	defer removeTestContainerIssue193(runtime, issue193ExternalName)

	runContainer := func(extraArgs ...string) error {
		removeTestContainerIssue193(runtime, issue193ExternalName)
		args := append([]string{"run", "-d", "--name", issue193ExternalName}, extraArgs...)
		args = append(args, issue193TestImage, "sleep", "3600")
		return exec.Command(runtime, args...).Run()
	}

	statusOutput := func() string {
		out, _ := exec.Command(binaryPath, "aiops", "ollama", "container", "status").CombinedOutput()
		return string(out)
	}

	// TC001: -p 21434:11434 -> API Port 21434
	tf.Separator()
	tf.Step(t, "TC001: external 'ollama' with -p "+issue242HostPort+":11434 -> API Port "+issue242HostPort)
	if err := runContainer("-p", issue242HostPort+":11434"); err != nil {
		tf.Error(t, "Failed to create test container", err.Error())
		success = false
		return
	}
	out := statusOutput()
	tf.Output(t, out, 800)
	if strings.Contains(out, "API Port:    "+issue242HostPort) {
		tf.Success(t, "Published host port reported")
	} else {
		tf.Error(t, "Expected 'API Port:    "+issue242HostPort+"'")
		success = false
	}
	if strings.Contains(out, "localhost:11434") {
		tf.Error(t, "API check must not use localhost:11434 for a non-default mapping")
		success = false
	}

	// TC002: no published port -> "not published", no API check
	tf.Separator()
	tf.Step(t, "TC002: external 'ollama' without published port -> API Port: not published")
	if err := runContainer(); err != nil {
		tf.Error(t, "Failed to create test container", err.Error())
		success = false
		return
	}
	out = statusOutput()
	tf.Output(t, out, 800)
	if strings.Contains(out, "API Port:    not published") &&
		strings.Contains(out, "not published to the host") {
		tf.Success(t, "Missing port binding reported, API check skipped")
	} else {
		tf.Error(t, "Expected 'API Port:    not published' and skipped API check")
		success = false
	}

	// TC003: no GPU requested -> GPU: Disabled
	tf.Separator()
	tf.Step(t, "TC003: container without GPU -> GPU: Disabled")
	if strings.Contains(out, "GPU:         Disabled") {
		tf.Success(t, "GPU reported as Disabled")
	} else {
		tf.Error(t, "Expected 'GPU:         Disabled'")
		success = false
	}
}
