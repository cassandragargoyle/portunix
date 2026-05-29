/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue188_PluginListVerbose verifies the CLI surface of the description
// truncation / --verbose extension introduced in issue #188.
//
// The algorithm itself (rune-safe truncateString, wrapForTerminal) is covered
// in detail by src/cmd/plugin_list_format_test.go. This test focuses on the
// command-line behavior:
//   - --help advertises the new --verbose semantics.
//   - plugin list executes successfully in compact, verbose, and JSON modes.
//   - --verbose does not break --output json (full description is preserved).
func TestIssue188_PluginListVerbose(t *testing.T) {
	tf := testframework.NewTestFramework("Issue188_PluginListVerbose")
	tf.Start(t, "plugin list: --verbose full description + rune-safe truncation")

	success := true
	defer tf.Finish(t, success)

	binaryPath := locateBinary188(t, tf)
	if binaryPath == "" {
		success = false
		return
	}
	tf.Separator()

	// TC001: --help advertises new --verbose behavior on default listing
	tf.Step(t, "TC001: 'plugin list --help' documents --verbose full-description mode")
	tf.Command(t, binaryPath, []string{"plugin", "list", "--help"})
	cmd := exec.Command(binaryPath, "plugin", "list", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "plugin list --help failed", err.Error())
		success = false
		return
	}
	helpStr := string(output)
	tf.Output(t, helpStr, 1200)

	// Help text should advertise full-description wrapping in the default
	// listing path (not just the platform_payload behavior).
	if !strings.Contains(helpStr, "full description") {
		tf.Error(t, "plugin list --help does not advertise --verbose full-description mode")
		success = false
	} else {
		tf.Success(t, "Help text advertises full-description mode")
	}
	tf.Separator()

	// TC002: plain `plugin list` runs successfully (smoke test, no panic)
	tf.Step(t, "TC002: 'plugin list' runs in compact mode")
	tf.Command(t, binaryPath, []string{"plugin", "list"})
	cmd = exec.Command(binaryPath, "plugin", "list")
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "plugin list failed", err.Error(), string(output))
		success = false
		return
	}
	compactStr := string(output)
	tf.Output(t, compactStr, 1200)

	// If any plugin is installed the table header must appear; if not, the
	// "No plugins installed." sentinel is acceptable.
	hasHeader := strings.Contains(compactStr, "NAME") && strings.Contains(compactStr, "DESCRIPTION")
	hasEmptyMsg := strings.Contains(compactStr, "No plugins installed")
	if !hasHeader && !hasEmptyMsg {
		tf.Error(t, "plugin list output missing both table header and empty-state message")
		success = false
	} else {
		tf.Success(t, "Compact mode produces expected output")
	}
	tf.Separator()

	// TC003: `plugin list --verbose` runs successfully
	tf.Step(t, "TC003: 'plugin list --verbose' runs without error")
	tf.Command(t, binaryPath, []string{"plugin", "list", "--verbose"})
	cmd = exec.Command(binaryPath, "plugin", "list", "--verbose")
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "plugin list --verbose failed", err.Error(), string(output))
		success = false
		return
	}
	verboseStr := string(output)
	tf.Output(t, verboseStr, 1500)

	// Verbose header must not include the DESCRIPTION column — it is printed
	// separately on subsequent indented lines.
	if strings.Contains(verboseStr, "NAME") && strings.Contains(verboseStr, "DESCRIPTION") {
		tf.Error(t, "verbose header should drop the DESCRIPTION column")
		success = false
	} else {
		tf.Success(t, "Verbose mode runs and omits DESCRIPTION column header")
	}
	tf.Separator()

	// TC004: -v short flag is equivalent to --verbose
	tf.Step(t, "TC004: '-v' short flag works like '--verbose'")
	cmd = exec.Command(binaryPath, "plugin", "list", "-v")
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "plugin list -v failed", err.Error(), string(output))
		success = false
		return
	}
	if string(output) != verboseStr {
		// Output may legitimately differ if timing-dependent fields ever sneak
		// in; we only fail if -v dropped the verbose header (i.e. behaved
		// like compact mode).
		shortStr := string(output)
		if strings.Contains(shortStr, "DESCRIPTION") {
			tf.Error(t, "'-v' did not engage verbose mode")
			success = false
		} else {
			tf.Success(t, "'-v' produced verbose-shaped output")
		}
	} else {
		tf.Success(t, "'-v' and '--verbose' produce identical output")
	}
	tf.Separator()

	// TC005: --output json --verbose still emits valid JSON with full description
	tf.Step(t, "TC005: 'plugin list --output json --verbose' is unchanged")
	tf.Command(t, binaryPath, []string{"plugin", "list", "--output", "json", "--verbose"})
	cmd = exec.Command(binaryPath, "plugin", "list", "--output", "json", "--verbose")
	output, err = cmd.CombinedOutput()
	if err != nil {
		tf.Error(t, "plugin list --output json --verbose failed", err.Error(), string(output))
		success = false
		return
	}
	jsonStr := strings.TrimSpace(string(output))
	tf.Output(t, jsonStr, 800)

	// Empty registry -> "No plugins installed." plain text (not JSON). Otherwise
	// the output must parse as JSON. Both branches satisfy AC #5.
	if jsonStr == "" || strings.HasPrefix(jsonStr, "No plugins installed") {
		tf.Success(t, "JSON path runs (no plugins installed)")
	} else {
		var parsed []map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
			tf.Error(t, "JSON output is not valid", err.Error())
			success = false
		} else {
			tf.Success(t, "JSON output parses cleanly")
			// Description must be present (not truncated) in every entry.
			for i, p := range parsed {
				if _, ok := p["description"]; !ok {
					tf.Error(t, "JSON entry missing 'description' field", "index:", string(rune('0'+i)))
					success = false
					break
				}
			}
		}
	}
	tf.Separator()

	// TC006: invalid UTF-8 byte cannot appear in compact output even when
	// descriptions contain multi-byte characters. We can only assert validity
	// of the output we actually got; the algorithm-level UTF-8 safety is
	// covered exhaustively by the unit tests.
	tf.Step(t, "TC006: compact output is valid UTF-8")
	if !utf8.ValidString(compactStr) {
		tf.Error(t, "compact output contains invalid UTF-8 bytes")
		success = false
	} else {
		tf.Success(t, "Compact output is valid UTF-8")
	}
}

// locateBinary188 finds the portunix binary built at the repo root. Mirrors
// findBinary() in issue_155 but lives here so the two tests stay independent.
func locateBinary188(t *testing.T, tf *testframework.TestFramework) string {
	tf.Step(t, "Locate portunix binary")
	candidates := []string{"../../portunix", "../../portunix.exe"}
	if runtime.GOOS == "windows" {
		candidates = []string{"../../portunix.exe", "../../portunix"}
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
	tf.Error(t, "portunix binary not found at ../../portunix[.exe] — run 'make build-main' first")
	return ""
}
