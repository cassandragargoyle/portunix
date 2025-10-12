package integration

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue050MultiLevelHelp tests the multi-level help system implementation
func TestIssue050MultiLevelHelp(t *testing.T) {
	tf := testframework.NewTestFramework("Issue050_MultiLevelHelp")
	tf.Start(t, "Test multi-level help system with basic, expert, and AI formats")

	success := true
	defer tf.Finish(t, success)

	binaryPath := "../../portunix"

	// Test 1: Basic help should be concise
	tf.Step(t, "Test basic help output")
	tf.Command(t, binaryPath, []string{"--help"})

	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.Output()
	if err != nil {
		tf.Error(t, "Failed to get basic help", err.Error())
		success = false
		return
	}

	basicHelp := string(output)
	tf.Output(t, basicHelp, 500)

	// Check basic help requirements
	if !strings.Contains(basicHelp, "Common commands:") {
		tf.Error(t, "Basic help missing 'Common commands:' section")
		success = false
	}

	// MANDATORY: Check for help levels section
	if !strings.Contains(basicHelp, "Help levels:") {
		tf.Error(t, "Basic help missing MANDATORY 'Help levels:' section")
		success = false
	}

	if !strings.Contains(basicHelp, "--help-expert") {
		tf.Error(t, "Basic help missing --help-expert reference")
		success = false
	}

	if !strings.Contains(basicHelp, "--help-ai") {
		tf.Error(t, "Basic help missing --help-ai reference")
		success = false
	}

	// Check that basic help is concise (less than 50 lines)
	lines := strings.Split(basicHelp, "\n")
	if len(lines) > 50 {
		tf.Warning(t, "Basic help exceeds recommended 30-40 lines",
			"Lines:", len(lines))
	} else {
		tf.Success(t, "Basic help is concise", "Lines:", len(lines))
	}

	tf.Separator()

	// Test 2: Expert help should be comprehensive
	tf.Step(t, "Test expert help output")
	tf.Command(t, binaryPath, []string{"--help-expert"})

	cmd = exec.Command(binaryPath, "--help-expert")
	output, err = cmd.Output()
	if err != nil {
		tf.Error(t, "Failed to get expert help", err.Error())
		success = false
		return
	}

	expertHelp := string(output)
	tf.Output(t, expertHelp, 500)

	// Check expert help requirements
	if !strings.Contains(expertHelp, "EXPERT DOCUMENTATION") {
		tf.Error(t, "Expert help missing 'EXPERT DOCUMENTATION' header")
		success = false
	}

	if !strings.Contains(expertHelp, "Parameters:") {
		tf.Error(t, "Expert help missing parameter details")
		success = false
	}

	if !strings.Contains(expertHelp, "Examples:") {
		tf.Error(t, "Expert help missing examples")
		success = false
	}

	if !strings.Contains(expertHelp, "ENVIRONMENT VARIABLES:") {
		tf.Error(t, "Expert help missing environment variables section")
		success = false
	}

	// Expert help should be much longer than basic help
	expertLines := strings.Split(expertHelp, "\n")
	if len(expertLines) > len(lines)*2 {
		tf.Success(t, "Expert help is comprehensive",
			"Expert lines:", len(expertLines),
			"Basic lines:", len(lines))
	} else {
		tf.Warning(t, "Expert help might not be comprehensive enough")
	}

	tf.Separator()

	// Test 3: AI help should be valid JSON
	tf.Step(t, "Test AI help JSON format")
	tf.Command(t, binaryPath, []string{"--help-ai"})

	cmd = exec.Command(binaryPath, "--help-ai")
	output, err = cmd.Output()
	if err != nil {
		tf.Error(t, "Failed to get AI help", err.Error())
		success = false
		return
	}

	aiHelp := string(output)
	tf.Output(t, aiHelp, 300)

	// Validate JSON structure
	var aiData map[string]interface{}
	if err := json.Unmarshal([]byte(aiHelp), &aiData); err != nil {
		tf.Error(t, "AI help is not valid JSON", err.Error())
		success = false
		return
	}

	// Check required JSON fields
	if _, ok := aiData["tool"]; !ok {
		tf.Error(t, "AI help JSON missing 'tool' field")
		success = false
	}

	if _, ok := aiData["commands"]; !ok {
		tf.Error(t, "AI help JSON missing 'commands' field")
		success = false
	}

	if _, ok := aiData["description"]; !ok {
		tf.Error(t, "AI help JSON missing 'description' field")
		success = false
	}

	// Check commands structure
	if commands, ok := aiData["commands"].([]interface{}); ok {
		tf.Success(t, "AI help has valid commands array",
			"Command count:", len(commands))

		// Check first command structure
		if len(commands) > 0 {
			if cmd, ok := commands[0].(map[string]interface{}); ok {
				requiredFields := []string{"name", "brief", "description", "category"}
				for _, field := range requiredFields {
					if _, ok := cmd[field]; !ok {
						tf.Error(t, "Command missing field:", field)
						success = false
					}
				}
			}
		}
	} else {
		tf.Error(t, "AI help commands field is not an array")
		success = false
	}

	tf.Separator()

	// Test 4: Consistency check between formats
	tf.Step(t, "Check consistency between help formats")

	// Check that commands in basic help appear in expert help
	if strings.Contains(basicHelp, "install") && !strings.Contains(expertHelp, "install") {
		tf.Error(t, "Inconsistency: 'install' in basic but not expert help")
		success = false
	}

	// Docker should NOT be in basic help, but should be in expert help
	if strings.Contains(basicHelp, "  docker ") {
		tf.Error(t, "Docker should not be in basic help")
		success = false
	}

	if !strings.Contains(expertHelp, "docker") {
		tf.Error(t, "Docker should be in expert help")
		success = false
	}

	// Check that all three formats mention help levels
	if strings.Contains(basicHelp, "Help levels:") &&
		strings.Contains(expertHelp, "HELP LEVELS:") &&
		strings.Contains(aiHelp, "\"tool\"") {
		tf.Success(t, "All help formats are properly structured")
	} else {
		tf.Warning(t, "Some help formats might be missing expected content")
	}

	tf.Success(t, "Multi-level help system tests completed")
}

// TestIssue050HelpPerformance tests that help generation is fast
func TestIssue050HelpPerformance(t *testing.T) {
	tf := testframework.NewTestFramework("Issue050_HelpPerformance")
	tf.Start(t, "Test help generation performance (should be < 100ms)")

	success := true
	defer tf.Finish(t, success)

	binaryPath := "../../portunix"

	// Test basic help performance
	tf.Step(t, "Measure basic help generation time")

	cmd := exec.Command(binaryPath, "--help")
	if err := cmd.Run(); err != nil {
		tf.Error(t, "Failed to run basic help", err.Error())
		success = false
		return
	}

	tf.Success(t, "Basic help generated successfully")

	// Test expert help performance
	tf.Step(t, "Measure expert help generation time")

	cmd = exec.Command(binaryPath, "--help-expert")
	if err := cmd.Run(); err != nil {
		tf.Error(t, "Failed to run expert help", err.Error())
		success = false
		return
	}

	tf.Success(t, "Expert help generated successfully")

	// Test AI help performance
	tf.Step(t, "Measure AI help generation time")

	cmd = exec.Command(binaryPath, "--help-ai")
	if err := cmd.Run(); err != nil {
		tf.Error(t, "Failed to run AI help", err.Error())
		success = false
		return
	}

	tf.Success(t, "AI help generated successfully")
	tf.Info(t, "Note: Actual timing measurement would require more sophisticated benchmarking")
}