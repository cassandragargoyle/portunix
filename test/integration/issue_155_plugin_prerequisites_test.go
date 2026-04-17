package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

func TestIssue155_PluginPrerequisitesValidation(t *testing.T) {
	tf := testframework.NewTestFramework("Issue155_PluginPrerequisites")
	tf.Start(t, "Test plugin prerequisites validation command")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	// Find binary
	binaryPath := findBinary(t, tf)
	if binaryPath == "" {
		success = false
		return
	}

	tf.Separator()

	// TC001: Help output includes check subcommand
	tf.Step(t, "TC001: Verify 'plugin check' appears in plugin help")
	tf.Command(t, binaryPath, []string{"plugin", "--help"})

	cmd := exec.Command(binaryPath, "plugin", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Failed to get plugin help", err.Error())
		success = false
		return
	}
	tf.Output(t, string(output), 500)

	if !strings.Contains(string(output), "check") {
		tf.Error(t, "Missing 'check' subcommand in plugin help")
		success = false
		return
	}
	tf.Success(t, "check subcommand found in help")

	tf.Separator()

	// TC002: Check command help
	tf.Step(t, "TC002: Verify 'plugin check --help' works")
	tf.Command(t, binaryPath, []string{"plugin", "check", "--help"})

	cmd = exec.Command(binaryPath, "plugin", "check", "--help")
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "Failed to get check help", err.Error())
		success = false
		return
	}
	tf.Output(t, string(output), 500)

	helpStr := string(output)
	if !strings.Contains(helpStr, "--all") {
		tf.Error(t, "Missing --all flag in check help")
		success = false
		return
	}
	if !strings.Contains(helpStr, "--json") {
		tf.Error(t, "Missing --json flag in check help")
		success = false
		return
	}
	tf.Success(t, "check help displays --all and --json flags")

	tf.Separator()

	// TC003: Check all plugins (human-readable)
	tf.Step(t, "TC003: Run 'plugin check --all' for human-readable output")
	tf.Command(t, binaryPath, []string{"plugin", "check", "--all"})

	cmd = exec.Command(binaryPath, "plugin", "check", "--all")
	output, err = cmd.CombinedOutput()
	outputStr := string(output)
	tf.Output(t, outputStr, 2000)

	// Command may exit with code 1 or 2 for errors/warnings, that's expected
	if err != nil && !strings.Contains(outputStr, "Summary:") && !strings.Contains(outputStr, "No plugins installed") {
		tf.Error(t, "Unexpected error from check --all", err.Error())
		success = false
		return
	}

	// Verify output structure
	if strings.Contains(outputStr, "No plugins installed") {
		tf.Warning(t, "No plugins installed, skipping output validation")
	} else {
		if !strings.Contains(outputStr, "Plugin:") {
			tf.Error(t, "Missing 'Plugin:' header in output")
			success = false
			return
		}
		if !strings.Contains(outputStr, "Runtime:") {
			tf.Error(t, "Missing 'Runtime:' line in output")
			success = false
			return
		}
		if !strings.Contains(outputStr, "Summary:") {
			tf.Error(t, "Missing 'Summary:' line in output")
			success = false
			return
		}
		tf.Success(t, "Human-readable output has correct structure")
	}

	tf.Separator()

	// TC004: Check all plugins (JSON output)
	tf.Step(t, "TC004: Run 'plugin check --all --json' for JSON output")
	tf.Command(t, binaryPath, []string{"plugin", "check", "--all", "--json"})

	cmd = exec.Command(binaryPath, "plugin", "check", "--all", "--json")
	output, err = cmd.CombinedOutput()
	outputStr = string(output)
	tf.Output(t, outputStr, 2000)

	if err != nil && !strings.Contains(outputStr, "results") && !strings.Contains(outputStr, "No plugins installed") {
		tf.Error(t, "Unexpected error from check --all --json", err.Error())
		success = false
		return
	}

	if !strings.Contains(outputStr, "No plugins installed") {
		// Validate JSON structure
		var jsonResult map[string]interface{}
		if err := json.Unmarshal(output, &jsonResult); err != nil {
			tf.Error(t, "Invalid JSON output", err.Error())
			success = false
			return
		}

		if _, ok := jsonResult["results"]; !ok {
			tf.Error(t, "Missing 'results' key in JSON output")
			success = false
			return
		}
		if _, ok := jsonResult["summary"]; !ok {
			tf.Error(t, "Missing 'summary' key in JSON output")
			success = false
			return
		}

		// Validate summary structure
		summary, ok := jsonResult["summary"].(map[string]interface{})
		if !ok {
			tf.Error(t, "Invalid summary structure")
			success = false
			return
		}
		for _, key := range []string{"total", "ok", "warning", "error"} {
			if _, exists := summary[key]; !exists {
				tf.Error(t, fmt.Sprintf("Missing '%s' in summary", key))
				success = false
				return
			}
		}

		tf.Success(t, "JSON output has correct structure with results and summary")
	}

	tf.Separator()

	// TC005: Check single plugin (if any exist)
	tf.Step(t, "TC005: Run 'plugin check <name>' for a single plugin")

	// Get first plugin name from check --all output
	cmd = exec.Command(binaryPath, "plugin", "check", "--all", "--json")
	output, _ = cmd.CombinedOutput()

	var allResult map[string]interface{}
	if err := json.Unmarshal(output, &allResult); err == nil {
		if results, ok := allResult["results"].([]interface{}); ok && len(results) > 0 {
			firstPlugin := results[0].(map[string]interface{})
			pluginName := firstPlugin["plugin"].(string)

			tf.Command(t, binaryPath, []string{"plugin", "check", pluginName})
			cmd = exec.Command(binaryPath, "plugin", "check", pluginName)
			output, err = cmd.CombinedOutput()
			outputStr = string(output)
			tf.Output(t, outputStr, 500)

			if !strings.Contains(outputStr, "Plugin:") || !strings.Contains(outputStr, pluginName) {
				tf.Error(t, "Single plugin check missing plugin name in output")
				success = false
				return
			}
			tf.Success(t, fmt.Sprintf("Single plugin check works for '%s'", pluginName))

			tf.Separator()

			// TC006: Single plugin JSON output
			tf.Step(t, "TC006: Run 'plugin check <name> --json' for single plugin JSON")
			tf.Command(t, binaryPath, []string{"plugin", "check", pluginName, "--json"})

			cmd = exec.Command(binaryPath, "plugin", "check", pluginName, "--json")
			output, err = cmd.CombinedOutput()
			outputStr = string(output)
			tf.Output(t, outputStr, 500)

			var singleResult map[string]interface{}
			if err := json.Unmarshal(output, &singleResult); err != nil {
				tf.Error(t, "Invalid JSON for single plugin check", err.Error())
				success = false
				return
			}

			for _, key := range []string{"plugin", "version", "overall"} {
				if _, exists := singleResult[key]; !exists {
					tf.Error(t, fmt.Sprintf("Missing '%s' in single plugin JSON", key))
					success = false
					return
				}
			}
			tf.Success(t, "Single plugin JSON output has correct structure")
		} else {
			tf.Warning(t, "No plugins available for single plugin test")
		}
	} else {
		tf.Warning(t, "Could not parse plugin list, skipping single plugin test")
	}

	tf.Separator()

	// TC007: Check nonexistent plugin
	tf.Step(t, "TC007: Run 'plugin check nonexistent-plugin' should fail")
	tf.Command(t, binaryPath, []string{"plugin", "check", "nonexistent-plugin-xyz"})

	cmd = exec.Command(binaryPath, "plugin", "check", "nonexistent-plugin-xyz")
	output, err = cmd.CombinedOutput()
	outputStr = string(output)
	tf.Output(t, outputStr, 300)

	if err == nil {
		tf.Error(t, "Expected error for nonexistent plugin, got success")
		success = false
		return
	}
	tf.Success(t, "Nonexistent plugin returns error as expected")

	tf.Separator()

	// TC008: Check without arguments and without --all should fail
	tf.Step(t, "TC008: Run 'plugin check' without args should show error")
	tf.Command(t, binaryPath, []string{"plugin", "check"})

	cmd = exec.Command(binaryPath, "plugin", "check")
	output, err = cmd.CombinedOutput()
	outputStr = string(output)
	tf.Output(t, outputStr, 300)

	if err == nil {
		tf.Error(t, "Expected error when no plugin name and no --all flag")
		success = false
		return
	}
	if !strings.Contains(outputStr, "plugin name required") {
		tf.Error(t, "Expected 'plugin name required' message")
		success = false
		return
	}
	tf.Success(t, "Missing args error message is correct")
}

// findBinary locates the portunix binary
func findBinary(t *testing.T, tf *testframework.TestFramework) string {
	tf.Step(t, "Locate portunix binary")

	// Try relative paths from test directory
	candidates := []string{
		"../../portunix.exe",
		"../../portunix",
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			tf.Success(t, "Binary found", absPath)
			return absPath
		}
	}

	tf.Error(t, "Binary not found in any expected location")
	return ""
}
