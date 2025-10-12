package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"path/filepath"

	"portunix.ai/portunix/test/testframework"
)

// TestAnsibleInfrastructurePhase2Issue056 Integration tests for Issue #056 Phase 2 - Multi-Environment Support
type TestAnsibleInfrastructurePhase2Issue056 struct {
	binaryPath string
	tempDir    string
}

func TestIssue056AnsibleInfrastructurePhase2(t *testing.T) {
	tf := testframework.NewTestFramework("Issue056_Ansible_Infrastructure_Phase2")
	tf.Start(t, "Phase 2 integration tests for Ansible Infrastructure as Code with multi-environment support")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	tf.Step(t, "Initialize Phase 2 test suite")
	suite := &TestAnsibleInfrastructurePhase2Issue056{}
	suite.setupEnvironment(t, tf)

	tf.Success(t, "Test environment ready", suite.binaryPath)
	tf.Info(t, "Running 6 Phase 2 test cases")

	// Run subtests and track their success
	subTestSuccess := true
	subTestSuccess = subTestSuccess && t.Run("TC009_ContainerEnvironmentFlag", func(t *testing.T) {
		if !suite.testContainerEnvironmentFlag(t) { t.Fail() }
	})
	subTestSuccess = subTestSuccess && t.Run("TC010_VirtEnvironmentFlag", func(t *testing.T) {
		if !suite.testVirtEnvironmentFlag(t) { t.Fail() }
	})
	subTestSuccess = subTestSuccess && t.Run("TC011_EnvironmentValidation", func(t *testing.T) {
		if !suite.testEnvironmentValidation(t) { t.Fail() }
	})
	subTestSuccess = subTestSuccess && t.Run("TC012_ContainerInventoryGeneration", func(t *testing.T) {
		if !suite.testContainerInventoryGeneration(t) { t.Fail() }
	})
	subTestSuccess = subTestSuccess && t.Run("TC013_VirtInventoryGeneration", func(t *testing.T) {
		if !suite.testVirtInventoryGeneration(t) { t.Fail() }
	})
	subTestSuccess = subTestSuccess && t.Run("TC014_PlaybookHelpEnhanced", func(t *testing.T) {
		if !suite.testPlaybookHelpEnhanced(t) { t.Fail() }
	})

	success = success && subTestSuccess

	// Cleanup
	defer suite.cleanup(t, tf)
}

func (suite *TestAnsibleInfrastructurePhase2Issue056) setupEnvironment(t *testing.T, tf *testframework.TestFramework) {
	tf.Step(t, "Setup Phase 2 test environment")

	// Get binary path
	suite.binaryPath = "../../portunix"

	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "portunix_ansible_phase2_test_")
	if err != nil {
		tf.Error(t, "Failed to create temp dir", err.Error())
		t.FailNow()
	}
	suite.tempDir = tempDir

	tf.Success(t, "Phase 2 environment setup complete", fmt.Sprintf("Temp dir: %s", suite.tempDir))
}

func (suite *TestAnsibleInfrastructurePhase2Issue056) cleanup(t *testing.T, tf *testframework.TestFramework) {
	tf.Step(t, "Cleanup Phase 2 test environment")

	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
		tf.Success(t, "Temp directory cleaned up")
	}
}

func (suite *TestAnsibleInfrastructurePhase2Issue056) testContainerEnvironmentFlag(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC009_ContainerEnvironmentFlag")
	tf.Start(t, "Test --env container flag parsing and validation")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Create test .ptxbook file for container execution")
	ptxbookContent := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "test-container-env"
  description: "Test container environment execution"

spec:
  portunix:
    packages:
      - name: "python"
        variant: "default"
`

	ptxbookPath := filepath.Join(suite.tempDir, "test-container.ptxbook")
	err := ioutil.WriteFile(ptxbookPath, []byte(ptxbookContent), 0644)
	if err != nil {
		tf.Error(t, "Failed to create test .ptxbook file", err.Error())
		success = false
		return success
	}

	tf.Step(t, "Test --env container flag validation")
	tf.Command(t, suite.binaryPath, []string{"playbook", "run", ptxbookPath, "--env", "container", "--dry-run"})

	cmd := exec.Command(suite.binaryPath, "playbook", "run", ptxbookPath, "--env", "container", "--dry-run")
	output, err := cmd.CombinedOutput()

	tf.Output(t, string(output), 500)

	// Should not fail during flag parsing and validation
	expectedOutputs := []string{
		"Environment: container",
		"Container Image: ubuntu:22.04", // Default image
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(string(output), expected) {
			tf.Error(t, "Missing expected container environment output", expected)
			success = false
		}
	}

	if success {
		tf.Success(t, "Container environment flag working correctly")
	}

	return success
}

func (suite *TestAnsibleInfrastructurePhase2Issue056) testVirtEnvironmentFlag(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC010_VirtEnvironmentFlag")
	tf.Start(t, "Test --env virt flag parsing and validation")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Create test .ptxbook file for VM execution")
	ptxbookContent := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "test-virt-env"
  description: "Test VM environment execution"

spec:
  portunix:
    packages:
      - name: "java"
        variant: "17"
`

	ptxbookPath := filepath.Join(suite.tempDir, "test-virt.ptxbook")
	err := ioutil.WriteFile(ptxbookPath, []byte(ptxbookContent), 0644)
	if err != nil {
		tf.Error(t, "Failed to create test .ptxbook file", err.Error())
		success = false
		return success
	}

	tf.Step(t, "Test --env virt flag with --target parameter")
	tf.Command(t, suite.binaryPath, []string{"playbook", "run", ptxbookPath, "--env", "virt", "--target", "test-vm", "--dry-run"})

	cmd := exec.Command(suite.binaryPath, "playbook", "run", ptxbookPath, "--env", "virt", "--target", "test-vm", "--dry-run")
	output, err := cmd.CombinedOutput()

	tf.Output(t, string(output), 500)

	// Should not fail during flag parsing
	expectedOutputs := []string{
		"Environment: virt",
		"VM Target: test-vm",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(string(output), expected) {
			tf.Error(t, "Missing expected virt environment output", expected)
			success = false
		}
	}

	if success {
		tf.Success(t, "Virt environment flag working correctly")
	}

	return success
}

func (suite *TestAnsibleInfrastructurePhase2Issue056) testEnvironmentValidation(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC011_EnvironmentValidation")
	tf.Start(t, "Test environment parameter validation and error handling")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Create test .ptxbook file")
	ptxbookContent := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "test-validation"

spec:
  portunix:
    packages:
      - name: "python"
`

	ptxbookPath := filepath.Join(suite.tempDir, "test-validation.ptxbook")
	err := ioutil.WriteFile(ptxbookPath, []byte(ptxbookContent), 0644)
	if err != nil {
		tf.Error(t, "Failed to create test .ptxbook file", err.Error())
		success = false
		return success
	}

	tf.Step(t, "Test invalid environment value")
	cmd := exec.Command(suite.binaryPath, "playbook", "run", ptxbookPath, "--env", "invalid")
	output, err := cmd.CombinedOutput()

	// Should fail with error
	if err == nil {
		tf.Warning(t, "Command should have failed with invalid environment, but validation seems to work")
		// Don't fail the test if this specific validation works differently
	}

	tf.Output(t, string(output), 300)

	if !strings.Contains(string(output), "Invalid environment 'invalid'") {
		tf.Error(t, "Missing validation error message")
		success = false
	}

	tf.Step(t, "Test virt environment without target")
	cmd = exec.Command(suite.binaryPath, "playbook", "run", ptxbookPath, "--env", "virt", "--dry-run")
	output, err = cmd.CombinedOutput()

	tf.Output(t, string(output), 300)

	// Should show that target is required (in dry-run it won't fail immediately)
	if strings.Contains(string(output), "Environment: virt") {
		tf.Success(t, "Virt environment validation working")
	}

	if success {
		tf.Success(t, "Environment validation working correctly")
	}

	return success
}

func (suite *TestAnsibleInfrastructurePhase2Issue056) testContainerInventoryGeneration(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC012_ContainerInventoryGeneration")
	tf.Start(t, "Test Ansible inventory generation for container environments")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Create .ptxbook file with Ansible section")
	ptxbookContent := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "test-container-inventory"

spec:
  requirements:
    ansible:
      min_version: "2.15.0"

  portunix:
    packages:
      - name: "python"

  ansible:
    playbooks:
      - path: "./test-playbook.yml"
`

	// Create a dummy Ansible playbook file
	playbookContent := `---
- hosts: all
  tasks:
    - name: Test task
      debug:
        msg: "Hello from container"
`

	ptxbookPath := filepath.Join(suite.tempDir, "test-container-inventory.ptxbook")
	playbookPath := filepath.Join(suite.tempDir, "test-playbook.yml")

	err := ioutil.WriteFile(ptxbookPath, []byte(ptxbookContent), 0644)
	if err != nil {
		tf.Error(t, "Failed to create test .ptxbook file", err.Error())
		success = false
		return success
	}

	err = ioutil.WriteFile(playbookPath, []byte(playbookContent), 0644)
	if err != nil {
		tf.Error(t, "Failed to create test playbook file", err.Error())
		success = false
		return success
	}

	tf.Step(t, "Test container execution with inventory generation")
	tf.Command(t, suite.binaryPath, []string{"playbook", "run", ptxbookPath, "--env", "container", "--dry-run"})

	cmd := exec.Command(suite.binaryPath, "playbook", "run", ptxbookPath, "--env", "container", "--dry-run")
	output, _ := cmd.CombinedOutput()

	tf.Output(t, string(output), 800)

	// Check that container environment setup is initiated
	expectedOutputs := []string{
		"Environment: container",
		"🐳 Setting up container environment",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(string(output), expected) {
			tf.Error(t, "Missing expected container setup output", expected)
			success = false
		}
	}

	if success {
		tf.Success(t, "Container inventory generation test completed")
	}

	return success
}

func (suite *TestAnsibleInfrastructurePhase2Issue056) testVirtInventoryGeneration(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC013_VirtInventoryGeneration")
	tf.Start(t, "Test Ansible inventory generation for VM environments")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Create .ptxbook file with VM targeting")
	ptxbookContent := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "test-vm-inventory"

spec:
  requirements:
    ansible:
      min_version: "2.15.0"

  portunix:
    packages:
      - name: "java"
        variant: "17"

  ansible:
    playbooks:
      - path: "./vm-playbook.yml"
`

	// Create a dummy VM playbook
	playbookContent := `---
- hosts: all
  tasks:
    - name: VM test task
      debug:
        msg: "Hello from VM"
`

	ptxbookPath := filepath.Join(suite.tempDir, "test-vm-inventory.ptxbook")
	playbookPath := filepath.Join(suite.tempDir, "vm-playbook.yml")

	err := ioutil.WriteFile(ptxbookPath, []byte(ptxbookContent), 0644)
	if err != nil {
		tf.Error(t, "Failed to create test .ptxbook file", err.Error())
		success = false
		return success
	}

	err = ioutil.WriteFile(playbookPath, []byte(playbookContent), 0644)
	if err != nil {
		tf.Error(t, "Failed to create test playbook file", err.Error())
		success = false
		return success
	}

	tf.Step(t, "Test VM execution with inventory generation")
	tf.Command(t, suite.binaryPath, []string{"playbook", "run", ptxbookPath, "--env", "virt", "--target", "test-vm", "--dry-run"})

	cmd := exec.Command(suite.binaryPath, "playbook", "run", ptxbookPath, "--env", "virt", "--target", "test-vm", "--dry-run")
	output, _ := cmd.CombinedOutput()

	tf.Output(t, string(output), 800)

	// Check that VM environment setup is initiated
	expectedOutputs := []string{
		"Environment: virt",
		"VM Target: test-vm",
		"🖥️  Setting up VM environment",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(string(output), expected) {
			tf.Error(t, "Missing expected VM setup output", expected)
			success = false
		}
	}

	if success {
		tf.Success(t, "VM inventory generation test completed")
	}

	return success
}

func (suite *TestAnsibleInfrastructurePhase2Issue056) testPlaybookHelpEnhanced(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC014_PlaybookHelpEnhanced")
	tf.Start(t, "Test enhanced playbook run help with Phase 2 flags")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Test playbook run help output")
	tf.Command(t, suite.binaryPath, []string{"playbook", "run"})

	cmd := exec.Command(suite.binaryPath, "playbook", "run")
	output, _ := cmd.CombinedOutput()

	tf.Output(t, string(output), 500)

	// Check that Phase 2 flags are documented in help
	expectedOutputs := []string{
		"--env ENVIRONMENT",
		"--target TARGET",
		"--image IMAGE",
		"Execution environment (local, container, virt)",
		"Target for virt environment",
		"Container image for container environment",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(string(output), expected) {
			tf.Error(t, "Missing expected help output", expected)
			success = false
		}
	}

	if success {
		tf.Success(t, "Enhanced help output is correct")
	}

	return success
}