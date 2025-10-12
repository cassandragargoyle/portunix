package integration

import (
	"os/exec"
	"strings"
	"testing"
	"portunix.ai/portunix/test/testframework"
)

func TestIssue059PlaybookHelpCommand(t *testing.T) {
	tf := testframework.NewTestFramework("Issue059_Playbook_Help_Command")
	tf.Start(t, "Test playbook help command functionality (Issue #059)")

	success := true
	defer tf.Finish(t, success)

	// Get binary path
	binaryPath := "../../portunix"

	tf.Step(t, "Test --help flag")
	tf.Command(t, binaryPath, []string{"playbook", "--help"})

	cmd := exec.Command(binaryPath, "playbook", "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "playbook --help command failed", err.Error())
		success = false
		return
	}

	outputStr := string(output)
	tf.Output(t, outputStr, 300)

	// Verify essential content is present
	expectedStrings := []string{
		"portunix playbook - Ansible Infrastructure as Code Management",
		"USAGE:",
		"SUBCOMMANDS:",
		"EXAMPLES:",
		"ENTERPRISE FEATURES:",
		"run         Execute a .ptxbook file",
		"validate    Validate a .ptxbook file",
		"check       Check if ptx-ansible helper is available",
		"list        List available playbooks",
		"init        Generate template playbook",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			tf.Error(t, "Missing expected content in help output",
				"Expected: "+expected,
				"Not found in output")
			success = false
		}
	}

	if success {
		tf.Success(t, "All expected content found in --help output")
	}

	tf.Separator()

	// Test short -h flag
	tf.Step(t, "Test -h flag")
	tf.Command(t, binaryPath, []string{"playbook", "-h"})

	cmd = exec.Command(binaryPath, "playbook", "-h")
	output, err = cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "playbook -h command failed", err.Error())
		success = false
		return
	}

	outputStrShort := string(output)

	// Verify -h produces same output as --help
	if outputStr != outputStrShort {
		tf.Error(t, "-h and --help produce different output")
		success = false
	} else {
		tf.Success(t, "-h and --help produce identical output")
	}

	tf.Separator()

	// Test help subcommand
	tf.Step(t, "Test help subcommand")
	tf.Command(t, binaryPath, []string{"playbook", "help"})

	cmd = exec.Command(binaryPath, "playbook", "help")
	output, err = cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "playbook help command failed", err.Error())
		success = false
		return
	}

	outputStrHelp := string(output)

	// Verify help subcommand produces same output as --help
	if outputStr != outputStrHelp {
		tf.Error(t, "help subcommand and --help produce different output")
		success = false
	} else {
		tf.Success(t, "help subcommand and --help produce identical output")
	}

	tf.Separator()

	// Test that old error behavior is fixed
	tf.Step(t, "Verify old error behavior is fixed")

	// Check that we don't get the old error message
	oldErrorMessage := "Unknown playbook subcommand: --help"

	if strings.Contains(outputStr, oldErrorMessage) {
		tf.Error(t, "Old error behavior still present",
			"Found old error message: "+oldErrorMessage)
		success = false
	} else {
		tf.Success(t, "Old error behavior successfully fixed")
	}

	// Verify no circular reference
	circularReference := "Run 'portunix playbook --help' for available commands"

	if strings.Contains(outputStr, circularReference) {
		tf.Error(t, "Circular reference still present",
			"Found circular reference: "+circularReference)
		success = false
	} else {
		tf.Success(t, "Circular reference successfully eliminated")
	}

	tf.Step(t, "Summary")
	tf.Success(t, "Issue #059 - Playbook Help Command functionality verified")
}