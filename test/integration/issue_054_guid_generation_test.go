package integration

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/test/testframework"
)

func TestIssue054GUIDGeneration(t *testing.T) {
	tf := testframework.NewTestFramework("Issue054_GUID_Generation")
	tf.Start(t, "Test GUID generation functionality - random, deterministic, and validation")

	success := true
	defer tf.Finish(t, success)

	// Get binary path
	wd, err := filepath.Abs("../../")
	if err != nil {
		tf.Error(t, "Failed to get working directory", err.Error())
		success = false
		return
	}
	binaryPath := filepath.Join(wd, "portunix")

	tf.Step(t, "Verify Portunix binary exists")
	tf.Info(t, "Binary path", binaryPath)
	if _, err := exec.LookPath(binaryPath); err != nil {
		tf.Error(t, "Portunix binary not found", "Please run 'go build -o .' first")
		success = false
		return
	}
	tf.Success(t, "Binary found and accessible")

	tf.Separator()

	// Test 1: GUID command help (should be hidden from standard help)
	tf.Step(t, "Test GUID command visibility in standard help")
	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		tf.Warning(t, "Help command failed, but continuing", err.Error())
	}

	if strings.Contains(string(output), "guid") {
		tf.Error(t, "GUID command should be hidden from standard help")
		success = false
	} else {
		tf.Success(t, "GUID command correctly hidden from standard help")
	}

	tf.Separator()

	// Test 2: GUID command visible in expert help
	tf.Step(t, "Test GUID command visibility in expert help")
	cmd = exec.Command(binaryPath, "--help-expert")
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Warning(t, "Expert help command failed, but continuing", err.Error())
	}

	if !strings.Contains(string(output), "guid") {
		tf.Error(t, "GUID command should be visible in expert help")
		success = false
	} else {
		tf.Success(t, "GUID command correctly visible in expert help")
	}

	tf.Separator()

	// Test 3: Random GUID generation
	tf.Step(t, "Test random GUID generation")
	tf.Command(t, binaryPath, []string{"guid", "random"})

	cmd = exec.Command(binaryPath, "guid", "random")
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Random GUID generation failed", err.Error())
		success = false
		return
	}

	tf.Output(t, string(output), 200)

	// Validate UUID format
	uuidStr := strings.TrimSpace(string(output))
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(uuidStr) {
		tf.Error(t, "Invalid UUID format", "UUID:", uuidStr)
		success = false
	} else {
		tf.Success(t, "Valid UUID format generated")
	}

	// Check version (should be 4 for random)
	if len(uuidStr) >= 15 && uuidStr[14] != '4' {
		tf.Error(t, "Expected UUID version 4", "Version nibble:", string(uuidStr[14]))
		success = false
	} else {
		tf.Success(t, "Correct UUID version 4")
	}

	tf.Separator()

	// Test 4: Deterministic GUID generation
	tf.Step(t, "Test deterministic GUID generation")
	tf.Command(t, binaryPath, []string{"guid", "from", "test-string-1", "test-string-2"})

	cmd = exec.Command(binaryPath, "guid", "from", "test-string-1", "test-string-2")
	output1, err := cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Deterministic GUID generation failed", err.Error())
		success = false
		return
	}

	tf.Output(t, string(output1), 200)

	// Test deterministic behavior - same inputs should produce same output
	cmd = exec.Command(binaryPath, "guid", "from", "test-string-1", "test-string-2")
	output2, err := cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Second deterministic GUID generation failed", err.Error())
		success = false
		return
	}

	uuid1 := strings.TrimSpace(string(output1))
	uuid2 := strings.TrimSpace(string(output2))

	if uuid1 != uuid2 {
		tf.Error(t, "Deterministic generation not consistent", "First:", uuid1, "Second:", uuid2)
		success = false
	} else {
		tf.Success(t, "Deterministic generation consistent")
	}

	// Validate UUID format
	if !uuidRegex.MatchString(uuid1) {
		tf.Error(t, "Invalid deterministic UUID format", "UUID:", uuid1)
		success = false
	} else {
		tf.Success(t, "Valid deterministic UUID format")
	}

	// Check version (should be 5 for name-based)
	if len(uuid1) >= 15 && uuid1[14] != '5' {
		tf.Error(t, "Expected UUID version 5", "Version nibble:", string(uuid1[14]))
		success = false
	} else {
		tf.Success(t, "Correct UUID version 5")
	}

	tf.Separator()

	// Test 5: Different inputs produce different UUIDs
	tf.Step(t, "Test collision resistance")
	cmd = exec.Command(binaryPath, "guid", "from", "different", "inputs")
	output3, err := cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Third deterministic GUID generation failed", err.Error())
		success = false
		return
	}

	uuid3 := strings.TrimSpace(string(output3))
	if uuid1 == uuid3 {
		tf.Error(t, "Different inputs produced same UUID", "UUID1:", uuid1, "UUID3:", uuid3)
		success = false
	} else {
		tf.Success(t, "Different inputs produce different UUIDs")
	}

	tf.Separator()

	// Test 6: UUID validation - valid UUID
	tf.Step(t, "Test UUID validation - valid UUID")
	tf.Command(t, binaryPath, []string{"guid", "validate", uuid1})

	cmd = exec.Command(binaryPath, "guid", "validate", uuid1)
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "UUID validation failed for valid UUID", err.Error())
		success = false
	} else {
		outputStr := strings.TrimSpace(string(output))
		if outputStr != "Valid UUID" {
			tf.Error(t, "Unexpected validation output", "Expected: 'Valid UUID', Got:", outputStr)
			success = false
		} else {
			tf.Success(t, "Valid UUID correctly identified")
		}
	}

	tf.Separator()

	// Test 7: UUID validation - invalid UUID
	tf.Step(t, "Test UUID validation - invalid UUID")
	tf.Command(t, binaryPath, []string{"guid", "validate", "invalid-uuid"})

	cmd = exec.Command(binaryPath, "guid", "validate", "invalid-uuid")
	output, err = cmd.CombinedOutput()
	// Note: This should exit with code 1, so err is expected
	outputStr := strings.TrimSpace(string(output))
	if outputStr != "Invalid UUID format" {
		tf.Error(t, "Unexpected validation output for invalid UUID", "Expected: 'Invalid UUID format', Got:", outputStr)
		success = false
	} else {
		tf.Success(t, "Invalid UUID correctly identified")
	}

	tf.Separator()

	// Test 8: Error handling - missing arguments
	tf.Step(t, "Test error handling - missing arguments for 'from' command")
	tf.Command(t, binaryPath, []string{"guid", "from", "only-one-string"})

	cmd = exec.Command(binaryPath, "guid", "from", "only-one-string")
	output, err = cmd.CombinedOutput()
	if err == nil {
		tf.Error(t, "Expected error for missing second argument")
		success = false
	} else {
		tf.Success(t, "Correctly rejected insufficient arguments")
	}

	tf.Separator()

	// Test 9: Performance test - random generation
	tf.Step(t, "Test performance - multiple random generations")
	startTime := time.Now()
	for i := 0; i < 10; i++ {
		cmd = exec.Command(binaryPath, "guid", "random")
		_, err = cmd.CombinedOutput()
		if err != nil {
			tf.Error(t, "Performance test failed", "Iteration:", i, "Error:", err.Error())
			success = false
			break
		}
	}
	duration := time.Since(startTime)
	durationMs := int(duration.Milliseconds())
	if durationMs > 1000 { // 1 second for 10 generations should be more than enough
		tf.Warning(t, "Performance slower than expected", "Duration:", durationMs, "ms")
	} else {
		tf.Success(t, "Performance acceptable", "Duration:", durationMs, "ms")
	}

	tf.Separator()

	// Test 10: Edge cases - empty string handling
	tf.Step(t, "Test edge cases - empty string handling")
	tf.Command(t, binaryPath, []string{"guid", "from", "", "non-empty"})

	cmd = exec.Command(binaryPath, "guid", "from", "", "non-empty")
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Failed to handle empty string case", err.Error())
		success = false
	} else {
		uuid := strings.TrimSpace(string(output))
		if !uuidRegex.MatchString(uuid) {
			tf.Error(t, "Invalid UUID generated for empty string case", "UUID:", uuid)
			success = false
		} else {
			tf.Success(t, "Empty string case handled correctly")
		}
	}

	tf.Separator()

	// Test 11: Case sensitivity in validation
	tf.Step(t, "Test case sensitivity in validation")
	upperUUID := strings.ToUpper(uuid1)
	tf.Command(t, binaryPath, []string{"guid", "validate", upperUUID})

	cmd = exec.Command(binaryPath, "guid", "validate", upperUUID)
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Case sensitivity test failed", err.Error())
		success = false
	} else {
		outputStr := strings.TrimSpace(string(output))
		if outputStr != "Valid UUID" {
			tf.Error(t, "Uppercase UUID not accepted", "Output:", outputStr)
			success = false
		} else {
			tf.Success(t, "Case insensitive validation works")
		}
	}

	tf.Success(t, "All GUID generation tests completed")
}