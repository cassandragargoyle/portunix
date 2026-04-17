package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue173_ContainerNetworkVolumeInspect verifies the new
// `portunix container {network,volume,inspect}` subcommand trees.
// It exercises AC-1 through AC-6 from docs/issues/internal/173-*.md.
// AC-7 (engine refactor) is covered by existing Odoo acceptance re-run.
func TestIssue173_ContainerNetworkVolumeInspect(t *testing.T) {
	tf := testframework.NewTestFramework("Issue173_Container_Network_Volume_Inspect")
	tf.Start(t, "Verify new network/volume/inspect container subcommands")

	success := true
	defer tf.Finish(t, success)

	binaryPath := "../../portunix"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		binaryPath = "./portunix"
	}
	if _, err := os.Stat(binaryPath); err != nil {
		tf.Error(t, "Binary not found", err.Error())
		success = false
		return
	}
	tf.Success(t, "Binary found at:", binaryPath)

	// Abort early if no container runtime is installed; ptx-container now selects
	// runtime via binary presence so LookPath is the right gate here.
	if _, err := exec.LookPath("podman"); err != nil {
		if _, err2 := exec.LookPath("docker"); err2 != nil {
			tf.Warning(t, "No container runtime installed — skipping functional assertions")
			return
		}
	}

	stamp := time.Now().Format("150405")
	netName := "portunix-173-net-" + stamp
	volName := "portunix-173-vol-" + stamp
	ctrName := "portunix-173-ctr-" + stamp

	// Defer full cleanup regardless of intermediate failures.
	defer func() {
		exec.Command(binaryPath, "container", "rm", "-f", ctrName).Run()
		exec.Command(binaryPath, "container", "volume", "rm", volName).Run()
		exec.Command(binaryPath, "container", "network", "rm", netName).Run()
	}()

	// AC-6: help output exists for each new subcommand tree.
	tf.Step(t, "AC-6: network/volume/inspect --help output")
	for _, cmd := range [][]string{
		{"container", "network", "--help"},
		{"container", "volume", "--help"},
		{"container", "inspect", "--help"},
	} {
		out, err := exec.Command(binaryPath, cmd...).CombinedOutput()
		if err != nil {
			tf.Error(t, "Help command failed", strings.Join(cmd, " "), err.Error())
			success = false
			continue
		}
		if !strings.Contains(string(out), "Usage:") {
			tf.Error(t, "Missing Usage line", strings.Join(cmd, " "), string(out))
			success = false
		}
	}
	tf.Success(t, "All help outputs present")
	tf.Separator()

	// AC-1: network create (initial)
	tf.Step(t, "AC-1: network create")
	out, err := exec.Command(binaryPath, "container", "network", "create", netName).CombinedOutput()
	if err != nil {
		tf.Error(t, "network create failed", err.Error(), string(out))
		success = false
		return
	}
	tf.Success(t, "Network created")

	// AC-1: idempotent — second create is a no-op, not an error.
	tf.Step(t, "AC-1: network create is idempotent")
	out, err = exec.Command(binaryPath, "container", "network", "create", netName).CombinedOutput()
	if err != nil {
		tf.Error(t, "Idempotent re-create returned non-zero", err.Error(), string(out))
		success = false
	} else if !strings.Contains(string(out), "already exists") {
		tf.Error(t, "Expected 'already exists' informational message", string(out))
		success = false
	} else {
		tf.Success(t, "Re-create is a no-op")
	}
	tf.Separator()

	// AC-2: network list surfaces the newly-created network.
	tf.Step(t, "AC-2: network list contains new network")
	out, err = exec.Command(binaryPath, "container", "network", "list").CombinedOutput()
	if err != nil {
		tf.Error(t, "network list failed", err.Error())
		success = false
	} else if !strings.Contains(string(out), netName) {
		tf.Error(t, "network list missing created network", string(out))
		success = false
	} else {
		tf.Success(t, "Network list contains created network")
	}
	tf.Separator()

	// AC-5 (network variant): templated inspect returns a single value.
	tf.Step(t, "AC-5: network inspect -f '{{.Name}}'")
	out, err = exec.Command(binaryPath, "container", "network", "inspect", netName, "-f", "{{.Name}}").CombinedOutput()
	if err != nil {
		tf.Error(t, "network inspect -f failed", err.Error(), string(out))
		success = false
	} else if !strings.Contains(string(out), netName) {
		tf.Error(t, "templated inspect did not return network name", string(out))
		success = false
	} else {
		tf.Success(t, "Templated inspect works")
	}
	tf.Separator()

	// AC-4: volume create + inspect + rm
	tf.Step(t, "AC-4: volume create / inspect / rm")
	if out, err := exec.Command(binaryPath, "container", "volume", "create", volName).CombinedOutput(); err != nil {
		tf.Error(t, "volume create failed", err.Error(), string(out))
		success = false
		return
	}
	if out, err := exec.Command(binaryPath, "container", "volume", "inspect", volName).CombinedOutput(); err != nil {
		tf.Error(t, "volume inspect failed", err.Error(), string(out))
		success = false
	} else if !strings.Contains(string(out), volName) {
		tf.Error(t, "volume inspect output missing name", string(out))
		success = false
	}
	if out, err := exec.Command(binaryPath, "container", "volume", "rm", volName).CombinedOutput(); err != nil {
		tf.Error(t, "volume rm failed", err.Error(), string(out))
		success = false
	}
	// AC-4: rm of absent volume should surface a non-zero exit.
	if err := exec.Command(binaryPath, "container", "volume", "rm", volName).Run(); err == nil {
		tf.Error(t, "volume rm of missing volume must return non-zero exit")
		success = false
	} else {
		tf.Success(t, "volume lifecycle works")
	}
	tf.Separator()

	// AC-5 (container variant): inspect on a real container with -f template.
	tf.Step(t, "AC-5: container inspect")
	if out, err := exec.Command(binaryPath, "container", "run", "-d", "--name", ctrName, "alpine:latest", "sleep", "60").CombinedOutput(); err != nil {
		tf.Warning(t, "Could not create test container — skipping container inspect",
			err.Error(), string(out))
	} else {
		time.Sleep(1 * time.Second)
		out, err := exec.Command(binaryPath, "container", "inspect", ctrName, "-f", "{{.State.Status}}").CombinedOutput()
		if err != nil {
			tf.Error(t, "container inspect -f failed", err.Error(), string(out))
			success = false
		} else if !strings.Contains(string(out), "running") {
			tf.Error(t, "container inspect did not return 'running'", string(out))
			success = false
		} else {
			tf.Success(t, "container inspect works")
		}
	}
	tf.Separator()

	// AC-3: network rm removes the network.
	tf.Step(t, "AC-3: network rm")
	if out, err := exec.Command(binaryPath, "container", "network", "rm", netName).CombinedOutput(); err != nil {
		tf.Error(t, "network rm failed", err.Error(), string(out))
		success = false
	}
	// AC-3: rm of absent network must surface non-zero exit.
	if err := exec.Command(binaryPath, "container", "network", "rm", netName).Run(); err == nil {
		tf.Error(t, "network rm of missing network must return non-zero exit")
		success = false
	} else {
		tf.Success(t, "network rm behaviour correct")
	}
}
