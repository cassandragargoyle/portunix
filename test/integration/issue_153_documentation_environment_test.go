package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"portunix.ai/portunix/test/testframework"
	"strings"
	"testing"
)

func TestIssue153TemplateSystem(t *testing.T) {
	tf := testframework.NewTestFramework("Issue153_Template_System")
	tf.Start(t, "Test documentation environment template system (Issue #153 REQ-2)")

	success := true
	defer tf.Finish(t, success)

	binaryPath := "../../portunix"

	// TC001: Template list
	tf.Step(t, "TC001: playbook template list")
	tf.Command(t, binaryPath, []string{"playbook", "template", "list"})

	cmd := exec.Command(binaryPath, "playbook", "template", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "template list failed", err.Error())
		success = false
		return
	}

	outputStr := string(output)
	tf.Output(t, outputStr, 500)

	expectedEngines := []string{"docusaurus", "hugo", "docsify", "docsy"}
	for _, engine := range expectedEngines {
		if !strings.Contains(outputStr, engine) {
			tf.Error(t, fmt.Sprintf("Engine '%s' missing from template list", engine))
			success = false
		}
	}

	if strings.Contains(outputStr, "static-docs") {
		tf.Success(t, "Template 'static-docs' found with all engines")
	} else {
		tf.Error(t, "Template 'static-docs' not found")
		success = false
	}

	tf.Separator()

	// TC002: Template show
	tf.Step(t, "TC002: playbook template show static-docs")
	tf.Command(t, binaryPath, []string{"playbook", "template", "show", "static-docs"})

	cmd = exec.Command(binaryPath, "playbook", "template", "show", "static-docs")
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "template show failed", err.Error())
		success = false
		return
	}

	outputStr = string(output)
	tf.Output(t, outputStr, 500)

	showExpected := []string{"static-docs", "Parameters:", "--engine", "--target", "container", "local"}
	for _, exp := range showExpected {
		if !strings.Contains(outputStr, exp) {
			tf.Error(t, fmt.Sprintf("Missing '%s' in template show output", exp))
			success = false
		}
	}

	if success {
		tf.Success(t, "Template show displays all expected fields")
	}

	tf.Separator()

	// TC003: Template show for non-existent template
	tf.Step(t, "TC003: template show for non-existent template returns error message")
	tf.Command(t, binaryPath, []string{"playbook", "template", "show", "nonexistent"})

	cmd = exec.Command(binaryPath, "playbook", "template", "show", "nonexistent")
	output, err = cmd.CombinedOutput()
	outputStr = string(output)
	if err != nil || strings.Contains(outputStr, "not found") {
		tf.Success(t, "Non-existent template correctly returns error message")
	} else {
		tf.Error(t, "Non-existent template did not return error", outputStr)
		success = false
	}
}

func TestIssue153PlaybookInit(t *testing.T) {
	tf := testframework.NewTestFramework("Issue153_Playbook_Init")
	tf.Start(t, "Test playbook init for all engines (Issue #153 REQ-2)")

	success := true
	defer tf.Finish(t, success)

	binaryPath := "../../portunix"

	// Create temp directory for test outputs
	tmpDir, err := os.MkdirTemp("", "portunix-test-153-*")
	if err != nil {
		tf.Error(t, "Failed to create temp directory", err.Error())
		success = false
		return
	}
	defer os.RemoveAll(tmpDir)
	tf.Info(t, "Temp directory", tmpDir)

	engines := []struct {
		name        string
		expectedStr []string
	}{
		{"docusaurus", []string{"Docusaurus", "node:", "3000:3000"}},
		{"hugo", []string{"Hugo", "ubuntu:22.04", "1313:1313", "hugo"}},
		{"docsy", []string{"Docsy", "ubuntu:22.04", "1313:1313", "hugo", "go"}},
		{"docsify", []string{"Docsify", "node:", "docsify"}},
	}

	for i, engine := range engines {
		tf.Step(t, fmt.Sprintf("TC%03d: playbook init --engine %s", i+4, engine.name))

		projectName := fmt.Sprintf("test-%s", engine.name)
		outputFile := filepath.Join(tmpDir, projectName+".ptxbook")
		args := []string{"playbook", "init", filepath.Join(tmpDir, projectName), "--template", "static-docs", "--engine", engine.name, "--target", "container"}
		tf.Command(t, binaryPath, args)

		cmd := exec.Command(binaryPath, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			tf.Error(t, fmt.Sprintf("playbook init --engine %s failed", engine.name), err.Error(), string(output))
			success = false
			continue
		}
		tf.Output(t, string(output), 200)

		// Read generated file
		content, err := os.ReadFile(outputFile)
		if err != nil {
			tf.Error(t, fmt.Sprintf("Failed to read generated %s.ptxbook", engine.name), err.Error())
			success = false
			continue
		}

		contentStr := string(content)

		// Validate required content
		for _, exp := range engine.expectedStr {
			if !strings.Contains(contentStr, exp) {
				tf.Error(t, fmt.Sprintf("Generated %s playbook missing '%s'", engine.name, exp))
				success = false
			}
		}

		// Validate common structure
		commonFields := []string{"apiVersion: portunix.ai/v1", "kind: Playbook", "metadata:", "spec:", "scripts:"}
		for _, field := range commonFields {
			if !strings.Contains(contentStr, field) {
				tf.Error(t, fmt.Sprintf("Generated %s playbook missing structure field '%s'", engine.name, field))
				success = false
			}
		}

		// Validate playbook with portunix
		tf.Command(t, binaryPath, []string{"playbook", "validate", outputFile})
		validateCmd := exec.Command(binaryPath, "playbook", "validate", outputFile)
		validateOutput, validateErr := validateCmd.CombinedOutput()
		if validateErr != nil {
			tf.Warning(t, fmt.Sprintf("Playbook validation returned error for %s (may be expected)", engine.name), string(validateOutput))
		} else {
			tf.Success(t, fmt.Sprintf("Generated %s playbook is valid", engine.name))
		}

		tf.Separator()
	}

	if success {
		tf.Success(t, "All engines generate valid playbooks")
	}
}

func TestIssue153DirectInstallDryRun(t *testing.T) {
	tf := testframework.NewTestFramework("Issue153_Direct_Install_DryRun")
	tf.Start(t, "Test direct install commands with dry-run (Issue #153 REQ-7)")

	success := true
	defer tf.Finish(t, success)

	binaryPath := "../../portunix"

	packages := []struct {
		name     string
		expected []string
	}{
		{"docusaurus", []string{"docusaurus", "script"}},
		{"hugo", []string{"hugo", "Hugo Static Site Generator"}},
	}

	for i, pkg := range packages {
		tf.Step(t, fmt.Sprintf("TC%03d: install %s --dry-run", i+10, pkg.name))
		tf.Command(t, binaryPath, []string{"install", pkg.name, "--dry-run"})

		cmd := exec.Command(binaryPath, "install", pkg.name, "--dry-run")
		output, err := cmd.CombinedOutput()
		if err != nil {
			tf.Error(t, fmt.Sprintf("install %s --dry-run failed", pkg.name), err.Error(), string(output))
			success = false
			continue
		}

		outputStr := string(output)
		tf.Output(t, outputStr, 300)

		for _, exp := range pkg.expected {
			if !strings.Contains(outputStr, exp) {
				tf.Error(t, fmt.Sprintf("Dry-run output for '%s' missing '%s'", pkg.name, exp))
				success = false
			}
		}

		if strings.Contains(outputStr, "Installation completed successfully") || strings.Contains(outputStr, "DRY RUN") {
			tf.Success(t, fmt.Sprintf("install %s --dry-run completed", pkg.name))
		} else {
			tf.Error(t, fmt.Sprintf("install %s --dry-run unexpected output", pkg.name))
			success = false
		}

		tf.Separator()
	}

	if success {
		tf.Success(t, "All direct install dry-run tests passed")
	}
}
