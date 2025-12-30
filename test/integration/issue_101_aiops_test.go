package integration

import (
	"os/exec"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue101_AIOpsHelp tests basic help commands for aiops
func TestIssue101_AIOpsHelp(t *testing.T) {
	tf := testframework.NewTestFramework("Issue101_AIOps_Help")
	tf.Start(t, "Test PTX-AIOps help commands and basic functionality")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	// Step 1: Verify binary
	binaryPath, ok := tf.VerifyPortunixBinary(t)
	if !ok {
		success = false
		return
	}

	tf.Separator()

	// Step 2: Test main aiops help
	tf.Step(t, "Test aiops --help command")
	tf.Command(t, binaryPath, []string{"aiops", "--help"})

	cmd := exec.Command(binaryPath, "aiops", "--help")
	output, err := cmd.CombinedOutput()
	tf.Output(t, string(output), 1000)

	if err != nil {
		tf.Error(t, "aiops --help failed", err.Error())
		success = false
		return
	}

	// Verify expected sections in help
	expectedSections := []string{
		"GPU Operations",
		"Ollama Container Operations",
		"Model Operations",
		"Open WebUI Operations",
		"Stack Operations",
	}

	for _, section := range expectedSections {
		if strings.Contains(string(output), section) {
			tf.Success(t, "Found section: "+section)
		} else {
			tf.Error(t, "Missing section: "+section)
			success = false
		}
	}

	tf.Separator()

	// Step 3: Test gpu subcommand help
	tf.Step(t, "Test aiops gpu --help command")
	tf.Command(t, binaryPath, []string{"aiops", "gpu", "--help"})

	cmd = exec.Command(binaryPath, "aiops", "gpu", "--help")
	output, err = cmd.CombinedOutput()
	tf.Output(t, string(output), 500)

	if err != nil {
		tf.Warning(t, "aiops gpu --help returned error (may be expected)", err.Error())
	}

	if strings.Contains(string(output), "gpu status") || strings.Contains(string(output), "GPU") {
		tf.Success(t, "GPU help contains expected content")
	}

	tf.Separator()

	// Step 4: Test model subcommand help
	tf.Step(t, "Test aiops model --help command")
	tf.Command(t, binaryPath, []string{"aiops", "model", "--help"})

	cmd = exec.Command(binaryPath, "aiops", "model", "--help")
	output, err = cmd.CombinedOutput()
	tf.Output(t, string(output), 500)

	expectedModelCommands := []string{"list", "install", "remove", "info", "run"}
	for _, cmdName := range expectedModelCommands {
		if strings.Contains(string(output), cmdName) {
			tf.Success(t, "Found model command: "+cmdName)
		} else {
			tf.Warning(t, "Missing model command: "+cmdName)
		}
	}
}

// TestIssue101_GPUStatus tests GPU status command (works with or without GPU)
func TestIssue101_GPUStatus(t *testing.T) {
	tf := testframework.NewTestFramework("Issue101_GPU_Status")
	tf.Start(t, "Test GPU status command (graceful degradation without GPU)")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	binaryPath, ok := tf.VerifyPortunixBinary(t)
	if !ok {
		success = false
		return
	}

	tf.Separator()

	// Test gpu status - should work even without GPU
	tf.Step(t, "Test aiops gpu status command")
	tf.Command(t, binaryPath, []string{"aiops", "gpu", "status"})

	cmd := exec.Command(binaryPath, "aiops", "gpu", "status")
	output, err := cmd.CombinedOutput()
	tf.Output(t, string(output), 1000)

	// Command should complete (exit code 0) even without GPU
	if err != nil {
		tf.Warning(t, "gpu status command failed (may be expected without GPU)", err.Error())
	}

	// Check for expected output patterns
	if strings.Contains(string(output), "NVIDIA GPU Status") ||
		strings.Contains(string(output), "nvidia-smi not found") ||
		strings.Contains(string(output), "No NVIDIA GPUs") {
		tf.Success(t, "GPU status provides informative output")
	} else {
		tf.Warning(t, "Unexpected GPU status output format")
	}

	tf.Separator()

	// Test gpu check
	tf.Step(t, "Test aiops gpu check command")
	tf.Command(t, binaryPath, []string{"aiops", "gpu", "check"})

	cmd = exec.Command(binaryPath, "aiops", "gpu", "check")
	output, err = cmd.CombinedOutput()
	tf.Output(t, string(output), 1000)

	if strings.Contains(string(output), "GPU Container Readiness Check") ||
		strings.Contains(string(output), "NVIDIA GPU Detection") {
		tf.Success(t, "GPU check provides readiness information")
	}
}

// TestIssue101_ModelListAvailable tests model list --available command
func TestIssue101_ModelListAvailable(t *testing.T) {
	tf := testframework.NewTestFramework("Issue101_Model_List_Available")
	tf.Start(t, "Test model list --available command (embedded registry)")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	binaryPath, ok := tf.VerifyPortunixBinary(t)
	if !ok {
		success = false
		return
	}

	tf.Separator()

	tf.Step(t, "Test aiops model list --available command")
	tf.Command(t, binaryPath, []string{"aiops", "model", "list", "--available"})

	cmd := exec.Command(binaryPath, "aiops", "model", "list", "--available")
	output, err := cmd.CombinedOutput()
	tf.Output(t, string(output), 2000)

	if err != nil {
		tf.Error(t, "model list --available failed", err.Error())
		success = false
		return
	}

	// Verify expected models in registry
	expectedModels := []string{
		"llama3.2",
		"mistral",
		"codellama",
		"phi3",
		"gemma2",
		"qwen2.5",
	}

	foundModels := 0
	for _, model := range expectedModels {
		if strings.Contains(string(output), model) {
			tf.Success(t, "Found model in registry: "+model)
			foundModels++
		} else {
			tf.Warning(t, "Model not found in registry: "+model)
		}
	}

	if foundModels < 3 {
		tf.Error(t, "Too few models found in registry", "Expected at least 3")
		success = false
	}

	// Verify rating system
	if strings.Contains(string(output), "★") || strings.Contains(string(output), "☆") {
		tf.Success(t, "Rating system present in output")
	} else {
		tf.Warning(t, "Rating system not visible in output")
	}

	// Verify table structure
	if strings.Contains(string(output), "NAME") && strings.Contains(string(output), "DESCRIPTION") {
		tf.Success(t, "Table headers present")
	}
}

// TestIssue101_ModelInfo tests model info command
func TestIssue101_ModelInfo(t *testing.T) {
	tf := testframework.NewTestFramework("Issue101_Model_Info")
	tf.Start(t, "Test model info command for various models")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	binaryPath, ok := tf.VerifyPortunixBinary(t)
	if !ok {
		success = false
		return
	}

	tf.Separator()

	// Test info for known models
	modelsToTest := []struct {
		name     string
		expected []string
	}{
		{"llama3.2", []string{"Llama 3.2", "Meta", "1b", "3b"}},
		{"phi3", []string{"Phi-3", "Microsoft", "mini", "medium"}},
		{"codellama", []string{"Code Llama", "Meta", "code generation"}},
		{"gemma2", []string{"Gemma 2", "Google", "2b", "9b"}},
	}

	for _, model := range modelsToTest {
		tf.Step(t, "Test model info for: "+model.name)
		tf.Command(t, binaryPath, []string{"aiops", "model", "info", model.name})

		cmd := exec.Command(binaryPath, "aiops", "model", "info", model.name)
		output, err := cmd.CombinedOutput()
		tf.Output(t, string(output), 1000)

		if err != nil {
			tf.Warning(t, "model info command failed for "+model.name, err.Error())
			continue
		}

		foundExpected := 0
		for _, exp := range model.expected {
			if strings.Contains(strings.ToLower(string(output)), strings.ToLower(exp)) {
				foundExpected++
			}
		}

		if foundExpected >= len(model.expected)/2 {
			tf.Success(t, "Model info complete for: "+model.name)
		} else {
			tf.Warning(t, "Model info incomplete for: "+model.name)
		}

		tf.Separator()
	}

	// Test info for unknown model
	tf.Step(t, "Test model info for unknown model")
	tf.Command(t, binaryPath, []string{"aiops", "model", "info", "unknown-model-xyz"})

	unknownCmd := exec.Command(binaryPath, "aiops", "model", "info", "unknown-model-xyz")
	unknownOutput, _ := unknownCmd.CombinedOutput()
	tf.Output(t, string(unknownOutput), 500)

	if strings.Contains(string(unknownOutput), "not available") ||
		strings.Contains(string(unknownOutput), "ollama.ai") ||
		strings.Contains(string(unknownOutput), "unknown-model-xyz") {
		tf.Success(t, "Unknown model handled gracefully")
	}
}

// TestIssue101_ContainerFlag tests --container flag validation
func TestIssue101_ContainerFlag(t *testing.T) {
	tf := testframework.NewTestFramework("Issue101_Container_Flag")
	tf.Start(t, "Test --container flag for model commands")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	binaryPath, ok := tf.VerifyPortunixBinary(t)
	if !ok {
		success = false
		return
	}

	tf.Separator()

	// Test model list with non-existent container
	tf.Step(t, "Test model list with non-existent container")
	tf.Command(t, binaryPath, []string{"aiops", "model", "list", "--container", "nonexistent-container-xyz"})

	listCmd := exec.Command(binaryPath, "aiops", "model", "list", "--container", "nonexistent-container-xyz")
	listOutput, _ := listCmd.CombinedOutput()
	tf.Output(t, string(listOutput), 500)

	// Should fail gracefully with informative message
	if strings.Contains(string(listOutput), "not found") ||
		strings.Contains(string(listOutput), "Container") {
		tf.Success(t, "Non-existent container handled gracefully")
	} else {
		tf.Warning(t, "Expected container not found message")
	}

	tf.Separator()

	// Test model install with non-existent container
	tf.Step(t, "Test model install with non-existent container")
	tf.Command(t, binaryPath, []string{"aiops", "model", "install", "llama3.2", "--container", "nonexistent-container-xyz"})

	installCmd := exec.Command(binaryPath, "aiops", "model", "install", "llama3.2", "--container", "nonexistent-container-xyz")
	installOutput, _ := installCmd.CombinedOutput()
	tf.Output(t, string(installOutput), 500)

	if strings.Contains(string(installOutput), "not found") ||
		strings.Contains(string(installOutput), "Container") {
		tf.Success(t, "Model install with bad container handled gracefully")
	}
}

// TestIssue101_StackHelp tests stack command help
func TestIssue101_StackHelp(t *testing.T) {
	tf := testframework.NewTestFramework("Issue101_Stack_Help")
	tf.Start(t, "Test stack command help and status")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	binaryPath, ok := tf.VerifyPortunixBinary(t)
	if !ok {
		success = false
		return
	}

	tf.Separator()

	// Test stack status (should work without containers)
	tf.Step(t, "Test aiops stack status command")
	tf.Command(t, binaryPath, []string{"aiops", "stack", "status"})

	cmd := exec.Command(binaryPath, "aiops", "stack", "status")
	output, err := cmd.CombinedOutput()
	tf.Output(t, string(output), 1000)

	if err != nil {
		tf.Warning(t, "stack status command failed", err.Error())
	}

	if strings.Contains(string(output), "AI Stack Status") ||
		strings.Contains(string(output), "Ollama") ||
		strings.Contains(string(output), "WebUI") {
		tf.Success(t, "Stack status provides informative output")
	}
}
