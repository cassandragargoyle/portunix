/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package unit

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// projectRootAtStartup is resolved once at package init so these tests stay
// correct even if another test in the same package changed the process cwd
// via os.Chdir (e.g. container_runtime_test.go, config_test.go). Relative
// cmd.Dir would otherwise intermittently resolve into a temp dir and cause
// `./portunix` to fail to start or behave unexpectedly.
var projectRootAtStartup = func() string {
	wd, _ := os.Getwd()
	root, _ := filepath.Abs(filepath.Join(wd, "..", ".."))
	return root
}()

// portunixBinary returns the absolute path to the portunix binary built at
// project root. Tests should use this instead of a relative "./portunix".
func portunixBinary() string {
	exe := "portunix"
	if filepath.Separator == '\\' {
		exe += ".exe"
	}
	return filepath.Join(projectRootAtStartup, exe)
}

func TestInstallCommandHelpFlag(t *testing.T) {
	// Test that --help flag works correctly
	cmd := exec.Command(portunixBinary(), "install", "--help")
	cmd.Dir = projectRootAtStartup

	output, err := cmd.CombinedOutput()

	// Help command should succeed (exit code 0)
	if err != nil {
		// Check if this is just because binary doesn't exist
		if strings.Contains(err.Error(), "no such file") {
			t.Skip("Portunix binary not found, skipping integration test")
			return
		}
		t.Errorf("Expected --help to succeed, got error: %v", err)
	}

	outputStr := string(output)

	// Should contain help text, not error about package
	if strings.Contains(outputStr, "package '--help' not found") {
		t.Error("Install command incorrectly interprets --help as package name")
	}

	// Should contain usage information
	if !strings.Contains(outputStr, "Install") && !strings.Contains(outputStr, "install") {
		t.Error("Help output should contain install information")
	}
}

func TestInstallCommandShortHelpFlag(t *testing.T) {
	// Test that -h flag works correctly
	cmd := exec.Command(portunixBinary(), "install", "-h")
	cmd.Dir = projectRootAtStartup

	output, err := cmd.CombinedOutput()

	// Help command should succeed
	if err != nil {
		// Check if this is just because binary doesn't exist
		if strings.Contains(err.Error(), "no such file") {
			t.Skip("Portunix binary not found, skipping integration test")
			return
		}
		t.Errorf("Expected -h to succeed, got error: %v", err)
	}

	outputStr := string(output)

	// Should contain help text, not error about package
	if strings.Contains(outputStr, "package '-h' not found") {
		t.Error("Install command incorrectly interprets -h as package name")
	}
}

func TestInstallCommandInvalidPackage(t *testing.T) {
	// Test that invalid package names are handled properly
	cmd := exec.Command(portunixBinary(), "install", "nonexistent-package-xyz")
	cmd.Dir = projectRootAtStartup

	output, err := cmd.CombinedOutput()

	// Should fail with proper error message
	if err == nil {
		t.Error("Expected installation of invalid package to fail")
	}

	outputStr := string(output)

	// Should contain appropriate error message
	if !strings.Contains(outputStr, "not found") && !strings.Contains(outputStr, "unknown") {
		t.Error("Should provide clear error message for unknown package")
	}
}

func TestInstallCommandCACertificates(t *testing.T) {
	// Test that ca-certificates package is recognized
	// Note: This test only checks package recognition, not actual installation

	// Skip if not running as integration test
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test - set INTEGRATION_TEST=1 to run")
	}

	cmd := exec.Command(portunixBinary(), "install", "ca-certificates", "--dry-run")
	cmd.Dir = projectRootAtStartup

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Package should be recognized (not "not found")
	if strings.Contains(outputStr, "package 'ca-certificates' not found") {
		t.Error("ca-certificates package should be available in install configuration")
	}

	// If dry-run not supported, at least shouldn't immediately fail with "not found"
	if err != nil && strings.Contains(outputStr, "not found") {
		t.Error("ca-certificates package should be recognized")
	}
}

func TestInstallCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldError bool
		errorCheck  func(string) bool
	}{
		{
			name:        "No arguments",
			args:        []string{"install"},
			shouldError: false, // Should show help or list packages
			errorCheck:  func(output string) bool { return strings.Contains(output, "Usage") },
		},
		{
			name:        "Help flag long",
			args:        []string{"install", "--help"},
			shouldError: false,
			errorCheck:  func(output string) bool { return strings.Contains(output, "Install") },
		},
		{
			name:        "Help flag short",
			args:        []string{"install", "-h"},
			shouldError: false,
			errorCheck:  func(output string) bool { return strings.Contains(output, "Install") },
		},
		{
			name:        "Invalid flag as package",
			args:        []string{"install", "--invalid-flag"},
			shouldError: true,
			errorCheck:  func(output string) bool { return !strings.Contains(output, "package '--invalid-flag' not found") },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if os.Getenv("INTEGRATION_TEST") != "1" {
				t.Skip("Skipping integration test - set INTEGRATION_TEST=1 to run")
			}

			cmd := exec.Command(portunixBinary())
			cmd.Args = append(cmd.Args, test.args...)
			cmd.Dir = projectRootAtStartup

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if test.shouldError && err == nil {
				t.Errorf("Expected command to fail, but it succeeded. Output: %s", outputStr)
			} else if !test.shouldError && err != nil {
				// Check if it's just binary not found
				if !strings.Contains(err.Error(), "no such file") {
					t.Errorf("Expected command to succeed, but it failed: %v. Output: %s", err, outputStr)
				}
			}

			if test.errorCheck != nil && !test.errorCheck(outputStr) {
				t.Errorf("Output validation failed for %s. Output: %s", test.name, outputStr)
			}
		})
	}
}
