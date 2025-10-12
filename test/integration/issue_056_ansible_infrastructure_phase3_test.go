package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"portunix.ai/portunix/test/testframework"
)

// TestIssue056Phase3_ConditionalExecution tests conditional execution features
func TestIssue056Phase3_ConditionalExecution(t *testing.T) {
	tf := testframework.NewTestFramework("Issue056_Phase3_ConditionalExecution")
	tf.Start(t, "Test Phase 3 conditional execution with when conditions")

	success := true
	defer tf.Finish(t, success)

	// Setup
	tf.Step(t, "Setup test environment")
	binaryPath := "../../ptx-ansible"

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		tf.Error(t, "Binary not found", binaryPath)
		success = false
		return
	}
	tf.Success(t, "Binary found")

	tf.Separator()

	// TC015: Test conditional package installation
	tf.Step(t, "TC015: Create playbook with conditional packages")

	conditionalPlaybook := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "conditional-test"
  description: "Test conditional execution"

spec:
  variables:
    install_java: true
    install_node: false
    environment: "test"

  portunix:
    packages:
      - name: "python"
        variant: "3.13"
      - name: "java"
        variant: "17"
        when: "install_java"
      - name: "nodejs"
        variant: "20"
        when: "install_node"
      - name: "vscode"
        variant: "stable"
        when: "environment == 'development'"
      - name: "docker"
        variant: "latest"
        when: "os == 'linux'"
`

	tmpDir := "/tmp/ptx-test-phase3"
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	playbookPath := filepath.Join(tmpDir, "conditional.ptxbook")
	if err := os.WriteFile(playbookPath, []byte(conditionalPlaybook), 0644); err != nil {
		tf.Error(t, "Failed to create test playbook", err.Error())
		success = false
		return
	}
	tf.Success(t, "Test playbook created")

	tf.Step(t, "Execute conditional playbook in dry-run mode")
	tf.Command(t, binaryPath, []string{"playbook", "run", playbookPath, "--dry-run"})

	cmd := exec.Command(binaryPath, "playbook", "run", playbookPath, "--dry-run")
	output, err := cmd.CombinedOutput()
	tf.Output(t, string(output), 1000)

	if err != nil {
		tf.Error(t, "Conditional playbook execution failed", err.Error())
		success = false
		return
	}

	outputStr := string(output)

	// Verify conditional logic
	if !strings.Contains(outputStr, "Installing java") && !strings.Contains(outputStr, "Would install: java") {
		tf.Error(t, "Java should be installed (install_java=true)")
		success = false
	} else {
		tf.Success(t, "Java installation condition evaluated correctly")
	}

	if strings.Contains(outputStr, "Installing nodejs") || strings.Contains(outputStr, "Would install: nodejs") {
		tf.Error(t, "Node.js should be skipped (install_node=false)")
		success = false
	} else {
		tf.Success(t, "Node.js installation correctly skipped")
	}

	if strings.Contains(outputStr, "Installing vscode") || strings.Contains(outputStr, "Would install: vscode") {
		tf.Error(t, "VSCode should be skipped (environment != development)")
		success = false
	} else {
		tf.Success(t, "VSCode installation correctly skipped")
	}
}

// TestIssue056Phase3_VariableTemplating tests variable templating features
func TestIssue056Phase3_VariableTemplating(t *testing.T) {
	tf := testframework.NewTestFramework("Issue056_Phase3_VariableTemplating")
	tf.Start(t, "Test Phase 3 variable templating with Jinja2-style templates")

	success := true
	defer tf.Finish(t, success)

	// Setup
	tf.Step(t, "Setup test environment")
	binaryPath := "../../ptx-ansible"

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		tf.Error(t, "Binary not found", binaryPath)
		success = false
		return
	}
	tf.Success(t, "Binary found")

	tf.Separator()

	// TC016: Test variable templating
	tf.Step(t, "TC016: Create playbook with template variables")

	templatePlaybook := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "template-test"
  description: "Test variable templating"

spec:
  variables:
    java_version: "17"
    node_version: "20"
    editor: "vscode"
    environment: "development"

  environment:
    project_name: "test-project"

  portunix:
    packages:
      - name: "java"
        variant: "{{ java_version }}"
        vars:
          custom_flag: "{{ environment }}"
      - name: "{{ editor if environment == 'development' else 'vim' }}"
        variant: "stable"
      - name: "nodejs"
        variant: "{{ node_version }}"
        when: "project_name == 'test-project'"
`

	tmpDir := "/tmp/ptx-test-phase3"
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	playbookPath := filepath.Join(tmpDir, "template.ptxbook")
	if err := os.WriteFile(playbookPath, []byte(templatePlaybook), 0644); err != nil {
		tf.Error(t, "Failed to create test playbook", err.Error())
		success = false
		return
	}
	tf.Success(t, "Template playbook created")

	tf.Step(t, "Execute template playbook in dry-run mode")
	tf.Command(t, binaryPath, []string{"playbook", "run", playbookPath, "--dry-run"})

	cmd := exec.Command(binaryPath, "playbook", "run", playbookPath, "--dry-run")
	output, err := cmd.CombinedOutput()
	tf.Output(t, string(output), 1000)

	if err != nil {
		tf.Error(t, "Template playbook execution failed", err.Error())
		success = false
		return
	}

	outputStr := string(output)

	// Verify template resolution
	if !strings.Contains(outputStr, "variant: 17") && !strings.Contains(outputStr, "Installing java") {
		tf.Error(t, "Java version template not resolved correctly")
		success = false
	} else {
		tf.Success(t, "Java version template resolved correctly")
	}

	if !strings.Contains(outputStr, "Installing vscode") && !strings.Contains(outputStr, "Would install: vscode") {
		tf.Error(t, "Editor conditional template not resolved")
		success = false
	} else {
		tf.Success(t, "Editor conditional template resolved")
	}
}

// TestIssue056Phase3_RollbackMechanism tests rollback features
func TestIssue056Phase3_RollbackMechanism(t *testing.T) {
	tf := testframework.NewTestFramework("Issue056_Phase3_RollbackMechanism")
	tf.Start(t, "Test Phase 3 rollback mechanism and error handling")

	success := true
	defer tf.Finish(t, success)

	// Setup
	tf.Step(t, "Setup test environment")
	binaryPath := "../../ptx-ansible"

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		tf.Error(t, "Binary not found", binaryPath)
		success = false
		return
	}
	tf.Success(t, "Binary found")

	tf.Separator()

	// TC017: Test rollback configuration
	tf.Step(t, "TC017: Create playbook with rollback enabled")

	rollbackPlaybook := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "rollback-test"
  description: "Test rollback functionality"

spec:
  variables:
    enable_logging: true

  rollback:
    enabled: true
    preserve_logs: true
    timeout: "5m"
    on_failure:
      - type: "command"
        command: "echo 'Rollback: Cleaning up test environment'"
        description: "Clean up test files"
      - type: "command"
        command: "echo 'Rollback: Sending notification'"
        description: "Send rollback notification"
        when: "enable_logging"

  portunix:
    packages:
      - name: "python"
        variant: "3.13"
      # This would cause failure in real execution
      - name: "nonexistent-package"
        variant: "latest"
`

	tmpDir := "/tmp/ptx-test-phase3"
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	playbookPath := filepath.Join(tmpDir, "rollback.ptxbook")
	if err := os.WriteFile(playbookPath, []byte(rollbackPlaybook), 0644); err != nil {
		tf.Error(t, "Failed to create test playbook", err.Error())
		success = false
		return
	}
	tf.Success(t, "Rollback playbook created")

	tf.Step(t, "Validate rollback playbook structure")
	tf.Command(t, binaryPath, []string{"playbook", "validate", playbookPath})

	cmd := exec.Command(binaryPath, "playbook", "validate", playbookPath)
	output, err := cmd.CombinedOutput()
	tf.Output(t, string(output), 1000)

	if err != nil {
		tf.Error(t, "Rollback playbook validation failed", err.Error())
		success = false
		return
	}

	outputStr := string(output)

	// Verify rollback parsing
	if !strings.Contains(outputStr, "validation successful") {
		tf.Error(t, "Rollback playbook validation should succeed")
		success = false
	} else {
		tf.Success(t, "Rollback playbook validation successful")
	}

	// Test dry-run with rollback enabled
	tf.Step(t, "Execute rollback playbook in dry-run mode")
	tf.Command(t, binaryPath, []string{"playbook", "run", playbookPath, "--dry-run"})

	cmd2 := exec.Command(binaryPath, "playbook", "run", playbookPath, "--dry-run")
	output2, err2 := cmd2.CombinedOutput()
	tf.Output(t, string(output2), 1000)

	if err2 != nil {
		tf.Error(t, "Rollback dry-run failed", err2.Error())
		success = false
		return
	}

	output2Str := string(output2)

	// Verify rollback protection is mentioned
	if !strings.Contains(output2Str, "Rollback protection") && !strings.Contains(output2Str, "rollback") {
		tf.Warning(t, "Rollback protection indication not found in output")
	} else {
		tf.Success(t, "Rollback protection detected in output")
	}
}

// TestIssue056Phase3_MCPIntegration tests MCP server integration
func TestIssue056Phase3_MCPIntegration(t *testing.T) {
	tf := testframework.NewTestFramework("Issue056_Phase3_MCPIntegration")
	tf.Start(t, "Test Phase 3 MCP server integration for AI-assisted playbook management")

	success := true
	defer tf.Finish(t, success)

	// Setup
	tf.Step(t, "Setup test environment")
	binaryPath := "../../ptx-ansible"

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		tf.Error(t, "Binary not found", binaryPath)
		success = false
		return
	}
	tf.Success(t, "Binary found")

	tf.Separator()

	// TC018: Test MCP manifest export
	tf.Step(t, "TC018: Test MCP tools manifest export")
	tf.Command(t, binaryPath, []string{"mcp", "manifest"})

	cmd := exec.Command(binaryPath, "mcp", "manifest")
	output, err := cmd.CombinedOutput()
	tf.Output(t, string(output), 1000)

	if err != nil {
		tf.Error(t, "MCP manifest export failed", err.Error())
		success = false
		return
	}

	outputStr := string(output)

	if !strings.Contains(outputStr, "Manifest saved") || !strings.Contains(outputStr, "MCP tools manifest") {
		tf.Error(t, "MCP manifest export should succeed")
		success = false
	} else {
		tf.Success(t, "MCP manifest export successful")
	}

	tf.Separator()

	// TC019: Test playbook generation from prompt
	tf.Step(t, "TC019: Test AI-assisted playbook generation")
	tf.Command(t, binaryPath, []string{"mcp", "generate", "Setup a Java development environment", "--name", "test-java-env"})

	cmd2 := exec.Command(binaryPath, "mcp", "generate", "Setup a Java development environment", "--name", "test-java-env")
	output2, err2 := cmd2.CombinedOutput()
	tf.Output(t, string(output2), 1000)

	if err2 != nil {
		tf.Error(t, "MCP playbook generation failed", err2.Error())
		success = false
		return
	}

	output2Str := string(output2)

	if !strings.Contains(output2Str, "Generated playbook") || !strings.Contains(output2Str, "test-java-env") {
		tf.Error(t, "MCP playbook generation should succeed")
		success = false
	} else {
		tf.Success(t, "MCP playbook generation successful")
	}

	tf.Separator()

	// TC020: Test MCP playbook listing
	tf.Step(t, "TC020: Test MCP playbook listing")
	tf.Command(t, binaryPath, []string{"mcp", "list", "generated-playbooks"})

	cmd3 := exec.Command(binaryPath, "mcp", "list", "generated-playbooks")
	output3, _ := cmd3.CombinedOutput()
	tf.Output(t, string(output3), 1000)

	// Note: This might fail if no playbooks exist, which is OK for testing
	output3Str := string(output3)

	if strings.Contains(output3Str, "Found") || strings.Contains(output3Str, "No .ptxbook files found") {
		tf.Success(t, "MCP playbook listing functional")
	} else {
		tf.Warning(t, "MCP playbook listing output unexpected", output3Str)
	}

	tf.Separator()

	// TC021: Test MCP help system
	tf.Step(t, "TC021: Test MCP help system")
	tf.Command(t, binaryPath, []string{"mcp"})

	cmd4 := exec.Command(binaryPath, "mcp")
	output4, err4 := cmd4.CombinedOutput()
	tf.Output(t, string(output4), 500)

	if err4 != nil {
		tf.Error(t, "MCP help command failed", err4.Error())
		success = false
		return
	}

	output4Str := string(output4)

	expectedCommands := []string{"generate", "validate", "list", "manifest"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(output4Str, cmd) {
			tf.Error(t, "MCP help missing command", cmd)
			success = false
		}
	}

	if success {
		tf.Success(t, "MCP help system shows all expected commands")
	}
}

// TestIssue056Phase3_Integration tests overall Phase 3 integration
func TestIssue056Phase3_Integration(t *testing.T) {
	tf := testframework.NewTestFramework("Issue056_Phase3_Integration")
	tf.Start(t, "Test Phase 3 complete integration with all advanced features")

	success := true
	defer tf.Finish(t, success)

	// Setup
	tf.Step(t, "Setup test environment")
	binaryPath := "../../ptx-ansible"

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		tf.Error(t, "Binary not found", binaryPath)
		success = false
		return
	}
	tf.Success(t, "Binary found")

	tf.Separator()

	// TC022: Test comprehensive Phase 3 playbook
	tf.Step(t, "TC022: Create comprehensive Phase 3 playbook")

	comprehensivePlaybook := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "phase3-comprehensive"
  description: "Comprehensive Phase 3 feature demonstration"

spec:
  variables:
    java_version: "17"
    enable_rollback: true
    environment_type: "development"
    user_preference: "vscode"

  environment:
    deploy_stage: "testing"
    use_containers: false

  rollback:
    enabled: true
    preserve_logs: true
    timeout: "10m"
    variables:
      rollback_reason: "Phase 3 test failure"
    on_failure:
      - type: "command"
        command: "echo 'Phase 3 rollback: {{ rollback_reason }}'"
        description: "Log rollback reason"
      - type: "command"
        command: "echo 'Cleaning up {{ environment_type }} environment'"
        description: "Clean up environment-specific files"
        when: "environment_type != 'production'"

  portunix:
    packages:
      - name: "python"
        variant: "3.13"
        vars:
          install_reason: "Base development requirement"
      - name: "java"
        variant: "{{ java_version }}"
        when: "java_version != ''"
        vars:
          jdk_type: "{{ 'full' if environment_type == 'development' else 'minimal' }}"
      - name: "{{ user_preference }}"
        variant: "stable"
        when: "environment_type == 'development'"
        vars:
          editor_config: "default"
      - name: "docker"
        variant: "latest"
        when: "use_containers"
        vars:
          container_runtime: "docker"
`

	tmpDir := "/tmp/ptx-test-phase3"
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	playbookPath := filepath.Join(tmpDir, "comprehensive.ptxbook")
	if err := os.WriteFile(playbookPath, []byte(comprehensivePlaybook), 0644); err != nil {
		tf.Error(t, "Failed to create comprehensive playbook", err.Error())
		success = false
		return
	}
	tf.Success(t, "Comprehensive Phase 3 playbook created")

	// Validate the complex playbook
	tf.Step(t, "Validate comprehensive playbook")
	tf.Command(t, binaryPath, []string{"playbook", "validate", playbookPath})

	cmd := exec.Command(binaryPath, "playbook", "validate", playbookPath)
	output, err := cmd.CombinedOutput()
	tf.Output(t, string(output), 1000)

	if err != nil {
		tf.Error(t, "Comprehensive playbook validation failed", err.Error())
		success = false
		return
	}

	if !strings.Contains(string(output), "validation successful") {
		tf.Error(t, "Comprehensive playbook should validate successfully")
		success = false
	} else {
		tf.Success(t, "Comprehensive playbook validation successful")
	}

	// Execute in dry-run mode
	tf.Step(t, "Execute comprehensive playbook in dry-run")
	tf.Command(t, binaryPath, []string{"playbook", "run", playbookPath, "--dry-run"})

	cmd2 := exec.Command(binaryPath, "playbook", "run", playbookPath, "--dry-run")
	output2, err2 := cmd2.CombinedOutput()
	tf.Output(t, string(output2), 1500)

	if err2 != nil {
		tf.Error(t, "Comprehensive playbook dry-run failed", err2.Error())
		success = false
		return
	}

	output2Str := string(output2)

	// Verify advanced features are working
	featureChecks := map[string]bool{
		"rollback protection": strings.Contains(output2Str, "Rollback") || strings.Contains(output2Str, "rollback"),
		"conditional execution": strings.Contains(output2Str, "Installing java") || strings.Contains(output2Str, "Would install: java"),
		"template variables": strings.Contains(output2Str, "variant: 17") || strings.Contains(output2Str, "java"),
		"conditional skipping": !strings.Contains(output2Str, "docker") && !strings.Contains(output2Str, "Docker"),
	}

	for feature, passed := range featureChecks {
		if passed {
			tf.Success(t, "Feature working", feature)
		} else {
			tf.Warning(t, "Feature not clearly demonstrated", feature)
		}
	}

	if strings.Contains(output2Str, "completed successfully") {
		tf.Success(t, "Comprehensive Phase 3 playbook executed successfully")
	} else {
		tf.Error(t, "Comprehensive playbook execution should complete successfully")
		success = false
	}
}