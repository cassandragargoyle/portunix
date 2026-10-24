package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

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
// - Help content is displayed (Usage, Available Commands, Flags, help command)
// - MCP server does NOT start automatically (preventing stdio mode activation)
// - Exit code behavior is correct
//
// Reference: docs/adr/005-revert-default-stdio-mode-for-mcp.md
// Related Issues: Issue #037 (MCP Serve Implementation), Issue #036 (Default stdio Mode)
func TestMainWithoutArguments(t *testing.T) {
	// Build the application first
	buildCmd := exec.Command("go", "build", "-o", "portunix_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("portunix_test") // Cleanup

	// Run the application without arguments
	cmd := exec.Command("./portunix_test")
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
		"Usage:",
		"Available Commands:",
		"Flags:",
		"help", // The help command should be available
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
	// Build the application first
	buildCmd := exec.Command("go", "build", "-o", "portunix_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("portunix_test") // Cleanup

	// Test that mcp serve command exists
	cmd := exec.Command("./portunix_test", "mcp", "serve", "--help")
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
	// Build the application first
	buildCmd := exec.Command("go", "build", "-o", "portunix_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("portunix_test") // Cleanup

	// Test that old mcp serve command no longer exists
	cmd := exec.Command("./portunix_test", "mcp serve", "--help")
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
