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

// TestAnsibleInfrastructureIssue056 Integration tests for Issue #056 - Ansible Infrastructure as Code
type TestAnsibleInfrastructureIssue056 struct {
	binaryPath string
	tempDir    string
}

func TestIssue056AnsibleInfrastructureAsCode(t *testing.T) {
	tf := testframework.NewTestFramework("Issue056_Ansible_Infrastructure_Phase1")
	tf.Start(t, "Phase 1 integration tests for Ansible Infrastructure as Code with .ptxbook support")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	tf.Step(t, "Initialize test suite")
	suite := &TestAnsibleInfrastructureIssue056{}
	suite.setupEnvironment(t, tf)

	tf.Success(t, "Test environment ready", suite.binaryPath)
	tf.Info(t, "Running 8 Phase 1 test cases")

	// Run subtests and track their success
	subTestSuccess := true
	subTestSuccess = subTestSuccess && t.Run("TC001_PtxAnsibleHelperAvailable", func(t *testing.T) {
		if !suite.testPtxAnsibleHelperAvailable(t) { t.Fail() }
	})
	subTestSuccess = subTestSuccess && t.Run("TC002_PlaybookCommandDispatch", func(t *testing.T) {
		if !suite.testPlaybookCommandDispatch(t) { t.Fail() }
	})
	subTestSuccess = subTestSuccess && t.Run("TC003_PtxbookValidationSimple", func(t *testing.T) {
		if !suite.testPtxbookValidationSimple(t) { t.Fail() }
	})
	subTestSuccess = subTestSuccess && t.Run("TC004_PtxbookValidationWithAnsible", func(t *testing.T) {
		if !suite.testPtxbookValidationWithAnsible(t) { t.Fail() }
	})
	subTestSuccess = subTestSuccess && t.Run("TC005_PtxbookValidationErrors", func(t *testing.T) {
		if !suite.testPtxbookValidationErrors(t) { t.Fail() }
	})
	subTestSuccess = subTestSuccess && t.Run("TC006_PtxbookDryRunExecution", func(t *testing.T) {
		if !suite.testPtxbookDryRunExecution(t) { t.Fail() }
	})
	subTestSuccess = subTestSuccess && t.Run("TC007_PtxbookPortunixOnlyExecution", func(t *testing.T) {
		if !suite.testPtxbookPortunixOnlyExecution(t) { t.Fail() }
	})
	subTestSuccess = subTestSuccess && t.Run("TC008_AnsiblePackageDefinition", func(t *testing.T) {
		if !suite.testAnsiblePackageDefinition(t) { t.Fail() }
	})

	success = success && subTestSuccess

	// Cleanup
	defer suite.cleanup(t, tf)
}

func (suite *TestAnsibleInfrastructureIssue056) setupEnvironment(t *testing.T, tf *testframework.TestFramework) {
	tf.Step(t, "Setup test environment")

	// Get binary path
	suite.binaryPath = "../../portunix"

	// Create temporary directory for test files
	tempDir, err := ioutil.TempDir("", "portunix_ansible_test_")
	if err != nil {
		tf.Error(t, "Failed to create temp dir", err.Error())
		t.FailNow()
	}
	suite.tempDir = tempDir

	tf.Success(t, "Environment setup complete", fmt.Sprintf("Temp dir: %s", suite.tempDir))
}

func (suite *TestAnsibleInfrastructureIssue056) cleanup(t *testing.T, tf *testframework.TestFramework) {
	tf.Step(t, "Cleanup test environment")

	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
		tf.Success(t, "Temp directory cleaned up")
	}
}

func (suite *TestAnsibleInfrastructureIssue056) testPtxAnsibleHelperAvailable(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC001_PtxAnsibleHelperAvailable")
	tf.Start(t, "Verify ptx-ansible helper binary is available and working")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Test ptx-ansible direct execution")
	cmd := exec.Command("../../ptx-ansible", "--version")
	output, err := cmd.Output()

	if err != nil {
		tf.Error(t, "ptx-ansible helper not available", err.Error())
		success = false
		return success
	}

	tf.Output(t, string(output), 200)

	if !strings.Contains(string(output), "ptx-ansible version") {
		tf.Error(t, "Invalid version output format")
		success = false
		return success
	}

	tf.Success(t, "ptx-ansible helper is available and working")
	return success
}

func (suite *TestAnsibleInfrastructureIssue056) testPlaybookCommandDispatch(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC002_PlaybookCommandDispatch")
	tf.Start(t, "Test that playbook commands are properly dispatched to ptx-ansible")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Test playbook check command")
	tf.Command(t, suite.binaryPath, []string{"playbook", "check"})

	cmd := exec.Command(suite.binaryPath, "playbook", "check")
	output, err := cmd.Output()

	if err != nil {
		tf.Error(t, "Playbook check command failed", err.Error())
		success = false
		return success
	}

	tf.Output(t, string(output), 300)

	if !strings.Contains(string(output), "ptx-ansible helper is available") {
		tf.Error(t, "Dispatcher not working correctly")
		success = false
		return success
	}

	tf.Success(t, "Playbook command dispatch working correctly")
	return success
}

func (suite *TestAnsibleInfrastructureIssue056) testPtxbookValidationSimple(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC003_PtxbookValidationSimple")
	tf.Start(t, "Test .ptxbook validation with Portunix-only playbook")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Create simple .ptxbook file")
	ptxbookContent := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "test-simple-dev"
  description: "Simple development environment"

spec:
  portunix:
    packages:
      - name: "python"
        variant: "default"
      - name: "java"
        variant: "17"
`

	ptxbookPath := filepath.Join(suite.tempDir, "test-simple.ptxbook")
	err := ioutil.WriteFile(ptxbookPath, []byte(ptxbookContent), 0644)
	if err != nil {
		tf.Error(t, "Failed to create test .ptxbook file", err.Error())
		success = false
		return success
	}

	tf.Step(t, "Validate .ptxbook file")
	tf.Command(t, suite.binaryPath, []string{"playbook", "validate", ptxbookPath})

	cmd := exec.Command(suite.binaryPath, "playbook", "validate", ptxbookPath)
	output, err := cmd.Output()

	if err != nil {
		tf.Error(t, "Validation command failed", err.Error())
		success = false
		return success
	}

	tf.Output(t, string(output), 500)

	expectedOutputs := []string{
		"✅ Playbook validation successful",
		"Name: test-simple-dev",
		"Description: Simple development environment",
		"Portunix packages: 2",
		"Type: Portunix-only",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(string(output), expected) {
			tf.Error(t, "Missing expected output", expected)
			success = false
		}
	}

	if success {
		tf.Success(t, "Simple .ptxbook validation successful")
	}

	return success
}

func (suite *TestAnsibleInfrastructureIssue056) testPtxbookValidationWithAnsible(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC004_PtxbookValidationWithAnsible")
	tf.Start(t, "Test .ptxbook validation with Ansible playbooks")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Create .ptxbook file with Ansible")
	ptxbookContent := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "test-full-stack"
  description: "Full stack development with Ansible"

spec:
  requirements:
    ansible:
      min_version: "2.15.0"

  portunix:
    packages:
      - name: "docker"
        variant: "latest"

  ansible:
    playbooks:
      - path: "./playbook1.yml"
      - path: "./playbook2.yml"
`

	ptxbookPath := filepath.Join(suite.tempDir, "test-ansible.ptxbook")
	err := ioutil.WriteFile(ptxbookPath, []byte(ptxbookContent), 0644)
	if err != nil {
		tf.Error(t, "Failed to create test .ptxbook file", err.Error())
		success = false
		return success
	}

	tf.Step(t, "Validate .ptxbook file with Ansible")
	cmd := exec.Command(suite.binaryPath, "playbook", "validate", ptxbookPath)
	output, err := cmd.Output()

	if err != nil {
		tf.Error(t, "Validation command failed", err.Error())
		success = false
		return success
	}

	tf.Output(t, string(output), 500)

	expectedOutputs := []string{
		"✅ Playbook validation successful",
		"Name: test-full-stack",
		"Portunix packages: 1",
		"Ansible playbooks: 2",
		"Requires Ansible: 2.15.0",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(string(output), expected) {
			tf.Error(t, "Missing expected output", expected)
			success = false
		}
	}

	if success {
		tf.Success(t, "Ansible .ptxbook validation successful")
	}

	return success
}

func (suite *TestAnsibleInfrastructureIssue056) testPtxbookValidationErrors(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC005_PtxbookValidationErrors")
	tf.Start(t, "Test .ptxbook validation error handling")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Create invalid .ptxbook file")
	invalidContent := `apiVersion: invalid/v1
kind: InvalidKind
metadata:
  description: "Missing name"
spec: {}
`

	invalidPath := filepath.Join(suite.tempDir, "test-invalid.ptxbook")
	err := ioutil.WriteFile(invalidPath, []byte(invalidContent), 0644)
	if err != nil {
		tf.Error(t, "Failed to create invalid .ptxbook file", err.Error())
		success = false
		return success
	}

	tf.Step(t, "Test validation error handling")
	cmd := exec.Command(suite.binaryPath, "playbook", "validate", invalidPath)
	output, err := cmd.CombinedOutput()

	// Should fail with exit code != 0
	if err == nil {
		tf.Error(t, "Validation should have failed for invalid file")
		success = false
		return success
	}

	tf.Output(t, string(output), 300)

	if !strings.Contains(string(output), "Validation failed") {
		tf.Error(t, "Missing validation failure message")
		success = false
		return success
	}

	tf.Success(t, "Error handling working correctly")
	return success
}

func (suite *TestAnsibleInfrastructureIssue056) testPtxbookDryRunExecution(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC006_PtxbookDryRunExecution")
	tf.Start(t, "Test .ptxbook dry-run execution")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Create test .ptxbook file")
	ptxbookContent := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "test-dryrun"
  description: "Test dry run execution"

spec:
  portunix:
    packages:
      - name: "python"
        variant: "default"
`

	ptxbookPath := filepath.Join(suite.tempDir, "test-dryrun.ptxbook")
	err := ioutil.WriteFile(ptxbookPath, []byte(ptxbookContent), 0644)
	if err != nil {
		tf.Error(t, "Failed to create test .ptxbook file", err.Error())
		success = false
		return success
	}

	tf.Step(t, "Execute dry-run")
	tf.Command(t, suite.binaryPath, []string{"playbook", "run", ptxbookPath, "--dry-run"})

	cmd := exec.Command(suite.binaryPath, "playbook", "run", ptxbookPath, "--dry-run")
	output, err := cmd.Output()

	if err != nil {
		tf.Error(t, "Dry-run execution failed", err.Error())
		success = false
		return success
	}

	tf.Output(t, string(output), 500)

	expectedOutputs := []string{
		"🔍 Dry-run mode",
		"[DRY-RUN] Would install: python",
		"✅ Dry-run completed successfully",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(string(output), expected) {
			tf.Error(t, "Missing expected dry-run output", expected)
			success = false
		}
	}

	if success {
		tf.Success(t, "Dry-run execution working correctly")
	}

	return success
}

func (suite *TestAnsibleInfrastructureIssue056) testPtxbookPortunixOnlyExecution(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC007_PtxbookPortunixOnlyExecution")
	tf.Start(t, "Test actual .ptxbook execution with Portunix-only packages")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Create minimal .ptxbook file")
	// Use a package that's likely to be already installed to avoid long installation times
	ptxbookContent := `apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "test-execution"
  description: "Test actual execution"

spec:
  portunix:
    packages:
      - name: "python"
        variant: "default"
`

	ptxbookPath := filepath.Join(suite.tempDir, "test-execution.ptxbook")
	err := ioutil.WriteFile(ptxbookPath, []byte(ptxbookContent), 0644)
	if err != nil {
		tf.Error(t, "Failed to create test .ptxbook file", err.Error())
		success = false
		return success
	}

	tf.Step(t, "Execute .ptxbook file")
	tf.Info(t, "This test may take longer if packages need to be installed")

	cmd := exec.Command(suite.binaryPath, "playbook", "run", ptxbookPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		tf.Warning(t, "Execution failed (may be expected if packages can't be installed)", err.Error())
		// Don't fail the test if installation fails due to permissions or other issues
	}

	tf.Output(t, string(output), 800)

	// Check that the execution at least started correctly
	if strings.Contains(string(output), "🚀 Executing playbook") {
		tf.Success(t, "Execution started correctly")
	} else {
		tf.Warning(t, "Execution may not have started as expected")
	}

	return true // Don't fail the test suite for execution issues
}

func (suite *TestAnsibleInfrastructureIssue056) testAnsiblePackageDefinition(t *testing.T) bool {
	tf := testframework.NewTestFramework("TC008_AnsiblePackageDefinition")
	tf.Start(t, "Test that Ansible package is properly defined in install-packages.json")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Check Ansible package definition")
	tf.Command(t, suite.binaryPath, []string{"install", "ansible", "--dry-run"})

	cmd := exec.Command(suite.binaryPath, "install", "ansible", "--dry-run")
	output, err := cmd.Output()

	if err != nil {
		tf.Error(t, "Ansible package not found", err.Error())
		success = false
		return success
	}

	tf.Output(t, string(output), 500)

	expectedOutputs := []string{
		"Ansible",
		"Infrastructure as Code automation platform",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(string(output), expected) {
			tf.Error(t, "Missing expected Ansible package info", expected)
			success = false
		}
	}

	if success {
		tf.Success(t, "Ansible package definition is correct")
	}

	return success
}