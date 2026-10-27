package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// buildTestBinary builds the main binary into the repository root (next to
// the helper binaries the dispatcher looks for) and returns its path. The
// .exe suffix is required on Windows, otherwise exec cannot find the file
func buildTestBinary(t *testing.T) string {
	t.Helper()
	name := "portunix_test"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path, err := filepath.Abs(name)
	if err != nil {
		t.Fatalf("Failed to resolve test binary path: %v", err)
	}
	if out, err := exec.Command("go", "build", "-o", path, ".").CombinedOutput(); err != nil {
		t.Fatalf("Failed to build application: %v\n%s", err, out)
	}
	t.Cleanup(func() { _ = os.Remove(path) })
	return path
}

// requireHelper skips the test when a helper binary is not built next to the
// main binary (e.g. plain go test without make build)
func requireHelper(t *testing.T, helper string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		helper += ".exe"
	}
	if _, err := os.Stat(helper); err != nil {
		t.Skipf("Helper %s not built (run make build first)", helper)
	}
}

// Issue #186 final cleanup (2026-05-19): TestInstallJavaRunDry and
// TestProcessArgumentsInstallJava were removed together with the
// legacy app/install package they exercised. Java installation is now
// owned by ptx-installer; integration tests live next to that helper.

// Issue #231: TestProcessArgumentsUnzip and TestUnzip were removed together
// with the legacy unzip argument parser. The unzip command is covered by
// src/cmd/unzip_test.go and the extraction by src/app/archive_test.go.

// TestMainWithoutArguments verifies that Portunix displays help content when executed without parameters.
// This test ensures compliance with ADR-005 (Revert Default stdio Mode for MCP - Return to Help Display)
// which specifies that CLI should follow standard conventions by showing help rather than entering MCP stdio mode.
//
// The test validates:
// - Help content is displayed (custom help from src/cmd/help_registry.go)
// - MCP server does NOT start automatically (preventing stdio mode activation)
// - Exit code behavior is correct
//
// Reference: docs/adr/005-revert-default-stdio-mode-for-mcp.md
// Related Issues: Issue #037 (MCP Serve Implementation), Issue #036 (Default stdio Mode),
// Issue #225 (expectations updated from cobra help to the custom help layout)
func TestMainWithoutArguments(t *testing.T) {
	binary := buildTestBinary(t)

	// Run the application without arguments
	cmd := exec.Command(binary)
	output, err := cmd.CombinedOutput()

	// The application should exit successfully and show help
	if err != nil {
		// Check if it's just a non-zero exit code (which is expected for help)
		if exitError, ok := err.(*exec.ExitError); ok {
			// Help command typically exits with code 0 or 1, both are acceptable
			if exitError.ExitCode() > 1 {
				t.Fatalf("Application exited with unexpected code: %v, output: %s", exitError.ExitCode(), string(output))
			}
		} else {
			t.Fatalf("Failed to run application: %v", err)
		}
	}

	outputStr := string(output)

	// Verify that help content is displayed
	expectedHelpContent := []string{
		"Usage: portunix [command] [options]",
		"Common commands:",
		"Help levels:",
		"--help-expert",
	}

	for _, expected := range expectedHelpContent {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected help output to contain '%s', but it didn't. Output:\n%s", expected, outputStr)
		}
	}

	// Verify it does NOT enter MCP stdio mode
	if strings.Contains(outputStr, "MCP stdio mode active") {
		t.Error("Application should not enter MCP stdio mode when run without arguments")
	}

	if strings.Contains(outputStr, "MCP Server starting") {
		t.Error("Application should not start MCP server when run without arguments")
	}
}

func TestMCPServeCommand(t *testing.T) {
	// mcp is served by the ptx-mcp helper through the dispatcher
	requireHelper(t, "ptx-mcp")
	binary := buildTestBinary(t)

	// Test that mcp serve command exists
	cmd := exec.Command(binary, "mcp", "serve", "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Help command typically exits with code 0, but some cobra apps exit with 1
			if exitError.ExitCode() > 1 {
				t.Fatalf("mcp serve --help exited with unexpected code: %v, output: %s", exitError.ExitCode(), string(output))
			}
		} else {
			t.Fatalf("Failed to run mcp serve --help: %v", err)
		}
	}

	outputStr := string(output)

	// Verify that mcp serve help content is displayed
	expectedContent := []string{
		"serve",
		"Start Model Context Protocol",
		"Communication Modes:",
		"stdio",
		"tcp",
		"unix",
		"--mode",
		"--port",
		"--socket",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected 'mcp serve --help' output to contain '%s', but it didn't. Output:\n%s", expected, outputStr)
		}
	}
}

func TestMCPServerCommandRemoved(t *testing.T) {
	binary := buildTestBinary(t)

	// Test that old mcp serve command no longer exists
	cmd := exec.Command(binary, "mcp serve", "--help")
	output, err := cmd.CombinedOutput()

	// This should fail since the command no longer exists
	if err == nil {
		t.Error("mcp serve command should not exist anymore, but it does")
	}

	outputStr := string(output)

	// Should contain error about unknown command
	if !strings.Contains(outputStr, "unknown command") && !strings.Contains(outputStr, "Unknown command") {
		t.Errorf("Expected error about unknown command for 'mcp serve', got: %s", outputStr)
	}
}
