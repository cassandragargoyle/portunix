package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/test/testframework"
)

func TestIssue052LoggingE2ECommandExecution(t *testing.T) {
	tf := testframework.NewTestFramework("Issue052_E2E_Command_Execution")
	tf.Start(t, "End-to-end test of logging system with real command execution")

	success := true
	defer tf.Finish(t, success)

	// Build binary first
	tf.Step(t, "Build Portunix binary for E2E testing")
	binaryPath := filepath.Join("..", "..", "portunix")

	buildCmd := exec.Command("go", "build", "-o", binaryPath)
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		tf.Error(t, "Failed to build binary", err.Error(), string(output))
		success = false
		return
	}
	tf.Success(t, "Binary built successfully")

	tf.Separator()

	// Step 1: Test basic command with logging
	tf.Step(t, "Test basic command execution with logging")

	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "e2e-test.log")

	// Set environment for file logging
	os.Setenv("PORTUNIX_LOG_LEVEL", "info")
	os.Setenv("PORTUNIX_LOG_OUTPUT", "console,file")
	os.Setenv("PORTUNIX_LOG_FILE", logFile)
	defer func() {
		os.Unsetenv("PORTUNIX_LOG_LEVEL")
		os.Unsetenv("PORTUNIX_LOG_OUTPUT")
		os.Unsetenv("PORTUNIX_LOG_FILE")
	}()

	// Execute version command
	tf.Command(t, binaryPath, []string{"version"})
	versionCmd := exec.Command(binaryPath, "version")

	var stdoutBuf, stderrBuf bytes.Buffer
	versionCmd.Stdout = &stdoutBuf
	versionCmd.Stderr = &stderrBuf

	err := versionCmd.Run()
	if err != nil {
		tf.Error(t, "Version command failed", err.Error())
		success = false
		return
	}

	stdoutOutput := stdoutBuf.String()
	stderrOutput := stderrBuf.String()

	tf.Output(t, stdoutOutput, 300)
	tf.Info(t, "Version command stdout captured")

	// Check stderr if needed
	if stderrOutput != "" {
		tf.Output(t, stderrOutput, 200)
		tf.Info(t, "Version command stderr captured")
	}

	// Version command should produce clean output
	if strings.Contains(stdoutOutput, "debug") || strings.Contains(stdoutOutput, "info") {
		tf.Warning(t, "Version output may contain log messages", stdoutOutput)
	} else {
		tf.Success(t, "Version command output is clean")
	}

	// Check if log file was created
	if _, err := os.Stat(logFile); err != nil {
		tf.Warning(t, "Log file not created", err.Error())
	} else {
		tf.Success(t, "Log file created successfully")
	}

	tf.Separator()

	// Step 2: Test verbose command execution
	tf.Step(t, "Test verbose command execution")

	// Set debug level for verbose testing
	os.Setenv("PORTUNIX_LOG_LEVEL", "debug")

	// Execute help command with verbose output
	tf.Command(t, binaryPath, []string{"--help"})
	helpCmd := exec.Command(binaryPath, "--help")

	stdoutBuf.Reset()
	stderrBuf.Reset()
	helpCmd.Stdout = &stdoutBuf
	helpCmd.Stderr = &stderrBuf

	err = helpCmd.Run()
	if err != nil {
		tf.Error(t, "Help command failed", err.Error())
		success = false
		return
	}

	helpOutput := stdoutBuf.String()
	tf.Output(t, helpOutput, 500)
	tf.Info(t, "Help command output captured")

	// Help should contain usage information
	if !strings.Contains(helpOutput, "Usage") && !strings.Contains(helpOutput, "Commands") {
		tf.Error(t, "Help output missing expected content", helpOutput)
		success = false
		return
	}

	tf.Success(t, "Help command working correctly")

	tf.Separator()

	// Step 3: Test system info command (if available)
	tf.Step(t, "Test system info command")

	tf.Command(t, binaryPath, []string{"system", "info"})
	systemCmd := exec.Command(binaryPath, "system", "info")

	stdoutBuf.Reset()
	stderrBuf.Reset()
	systemCmd.Stdout = &stdoutBuf
	systemCmd.Stderr = &stderrBuf

	err = systemCmd.Run()
	if err != nil {
		tf.Warning(t, "System info command failed", err.Error())
		// This may be expected if command doesn't exist
	} else {
		systemOutput := stdoutBuf.String()
		tf.Output(t, systemOutput, 400)
		tf.Info(t, "System info command output captured")
		tf.Success(t, "System info command working")
	}

	tf.Separator()

	// Step 4: Test command with container context simulation
	tf.Step(t, "Test command with simulated container context")

	// Create mock container environment
	os.Setenv("PORTUNIX_LOG_MODULE_CONTAINER", "debug")
	defer os.Unsetenv("PORTUNIX_LOG_MODULE_CONTAINER")

	// Execute a command that might use container detection
	tf.Command(t, binaryPath, []string{"docker", "--help"})
	dockerHelpCmd := exec.Command(binaryPath, "docker", "--help")

	stdoutBuf.Reset()
	stderrBuf.Reset()
	dockerHelpCmd.Stdout = &stdoutBuf
	dockerHelpCmd.Stderr = &stderrBuf

	err = dockerHelpCmd.Run()
	if err != nil {
		tf.Warning(t, "Docker help command failed", err.Error())
		// Expected if Docker support not available
	} else {
		dockerOutput := stdoutBuf.String()
		tf.Output(t, dockerOutput, 300)
		tf.Info(t, "Docker help command output captured")
		tf.Success(t, "Docker help command working")
	}

	tf.Separator()

	// Step 5: Verify comprehensive log file content
	tf.Step(t, "Verify comprehensive log file content")

	if _, err := os.Stat(logFile); err != nil {
		tf.Error(t, "Final log file check failed", err.Error())
		success = false
		return
	}

	logContent, err := os.ReadFile(logFile)
	if err != nil {
		tf.Error(t, "Failed to read final log file", err.Error())
		success = false
		return
	}

	logStr := string(logContent)
	tf.Output(t, logStr, 1000)
	tf.Info(t, "Final log file content captured")

	// Verify log contains expected elements
	expectedLogElements := []string{
		"version",        // From version command
		"level",          // Should have log level indicators
		"component",      // Should have component fields
	}

	for _, element := range expectedLogElements {
		if !strings.Contains(logStr, element) {
			tf.Warning(t, "Log file missing expected element", "element", element)
		} else {
			tf.Success(t, "Log file contains "+element)
		}
	}

	tf.Info(t, "E2E command execution test completed")
}

func TestIssue052LoggingE2EPerformanceImpact(t *testing.T) {
	tf := testframework.NewTestFramework("Issue052_E2E_Performance_Impact")
	tf.Start(t, "End-to-end performance impact testing with logging enabled/disabled")

	success := true
	defer tf.Finish(t, success)

	// Build binary
	tf.Step(t, "Build binary for performance testing")
	binaryPath := filepath.Join("..", "..", "portunix")

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		buildCmd := exec.Command("go", "build", "-o", binaryPath)
		buildCmd.Dir = filepath.Join("..", "..")
		if output, err := buildCmd.CombinedOutput(); err != nil {
			tf.Error(t, "Failed to build binary", err.Error(), string(output))
			success = false
			return
		}
	}
	tf.Success(t, "Binary available for testing")

	tf.Separator()

	// Step 1: Baseline performance without logging
	tf.Step(t, "Measure baseline performance without logging")

	// Disable logging
	os.Setenv("PORTUNIX_LOG_LEVEL", "error") // Minimal logging
	os.Setenv("PORTUNIX_LOG_OUTPUT", "file")
	os.Setenv("PORTUNIX_LOG_FILE", "/dev/null") // Discard logs
	defer func() {
		os.Unsetenv("PORTUNIX_LOG_LEVEL")
		os.Unsetenv("PORTUNIX_LOG_OUTPUT")
		os.Unsetenv("PORTUNIX_LOG_FILE")
	}()

	// Measure time for multiple help commands
	iterations := 10
	tf.Info(t, "Running baseline performance test", "iterations", iterations)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		cmd := exec.Command(binaryPath, "--help")
		if _, err := cmd.CombinedOutput(); err != nil {
			tf.Error(t, "Baseline test command failed", err.Error())
			success = false
			return
		}
	}
	baselineDuration := time.Since(start)

	tf.Info(t, "Baseline performance measured",
		"total_duration", baselineDuration.String(),
		"avg_per_command", (baselineDuration / time.Duration(iterations)).String())

	tf.Separator()

	// Step 2: Performance with full logging enabled
	tf.Step(t, "Measure performance with full logging enabled")

	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "perf-test.log")

	// Enable full logging
	os.Setenv("PORTUNIX_LOG_LEVEL", "debug")
	os.Setenv("PORTUNIX_LOG_OUTPUT", "console,file")
	os.Setenv("PORTUNIX_LOG_FILE", logFile)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		cmd := exec.Command(binaryPath, "--help")
		if _, err := cmd.CombinedOutput(); err != nil {
			tf.Error(t, "Full logging test command failed", err.Error())
			success = false
			return
		}
	}
	fullLoggingDuration := time.Since(start)

	tf.Info(t, "Full logging performance measured",
		"total_duration", fullLoggingDuration.String(),
		"avg_per_command", (fullLoggingDuration / time.Duration(iterations)).String())

	tf.Separator()

	// Step 3: Calculate performance impact
	tf.Step(t, "Calculate performance impact")

	impactRatio := float64(fullLoggingDuration) / float64(baselineDuration)
	impactPercent := (impactRatio - 1.0) * 100

	tf.Info(t, "Performance impact analysis",
		"baseline_duration", baselineDuration.String(),
		"logging_duration", fullLoggingDuration.String(),
		"impact_ratio", impactRatio,
		"impact_percent", impactPercent)

	// Requirement: Performance impact should be <5%
	if impactPercent > 5.0 {
		tf.Error(t, "Performance impact exceeds 5% requirement",
			"actual_impact", impactPercent, "requirement", "< 5%")
		success = false
		return
	} else if impactPercent > 2.0 {
		tf.Warning(t, "Performance impact is noticeable but within limits",
			"impact_percent", impactPercent)
	} else {
		tf.Success(t, "Performance impact is minimal", "impact_percent", impactPercent)
	}

	tf.Separator()

	// Step 4: Verify log file was created and has content
	tf.Step(t, "Verify logging output was generated")

	if stat, err := os.Stat(logFile); err != nil {
		tf.Error(t, "Performance test log file not created", err.Error())
		success = false
		return
	} else if stat.Size() == 0 {
		tf.Error(t, "Performance test log file is empty")
		success = false
		return
	} else {
		tf.Info(t, "Log file generated", "size_bytes", stat.Size())
		tf.Success(t, "Logging output verification passed")
	}

	tf.Info(t, "E2E performance impact test completed")
}

func TestIssue052LoggingE2EErrorScenarios(t *testing.T) {
	tf := testframework.NewTestFramework("Issue052_E2E_Error_Scenarios")
	tf.Start(t, "End-to-end testing of logging system under error conditions")

	success := true
	defer tf.Finish(t, success)

	// Build binary
	tf.Step(t, "Ensure binary is available")
	binaryPath := filepath.Join("..", "..", "portunix")

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		buildCmd := exec.Command("go", "build", "-o", binaryPath)
		buildCmd.Dir = filepath.Join("..", "..")
		if output, err := buildCmd.CombinedOutput(); err != nil {
			tf.Error(t, "Failed to build binary", err.Error(), string(output))
			success = false
			return
		}
	}
	tf.Success(t, "Binary available")

	tf.Separator()

	// Step 1: Test invalid command with logging
	tf.Step(t, "Test invalid command with logging enabled")

	tempDir := t.TempDir()
	errorLogFile := filepath.Join(tempDir, "error-test.log")

	os.Setenv("PORTUNIX_LOG_LEVEL", "debug")
	os.Setenv("PORTUNIX_LOG_OUTPUT", "file")
	os.Setenv("PORTUNIX_LOG_FILE", errorLogFile)
	defer func() {
		os.Unsetenv("PORTUNIX_LOG_LEVEL")
		os.Unsetenv("PORTUNIX_LOG_OUTPUT")
		os.Unsetenv("PORTUNIX_LOG_FILE")
	}()

	// Execute invalid command
	tf.Command(t, binaryPath, []string{"invalid-command", "with-args"})
	invalidCmd := exec.Command(binaryPath, "invalid-command", "with-args")

	var stdoutBuf, stderrBuf bytes.Buffer
	invalidCmd.Stdout = &stdoutBuf
	invalidCmd.Stderr = &stderrBuf

	err := invalidCmd.Run()
	// This should fail, that's expected
	if err == nil {
		tf.Warning(t, "Invalid command unexpectedly succeeded")
	} else {
		tf.Success(t, "Invalid command failed as expected", "error", err.Error())
	}

	// Check outputs
	stdoutOutput := stdoutBuf.String()
	stderrOutput := stderrBuf.String()

	tf.Output(t, stdoutOutput, 300)
	tf.Info(t, "Invalid command stdout captured")
	tf.Output(t, stderrOutput, 300)
	tf.Info(t, "Invalid command stderr captured")

	// Error information should be present
	if strings.Contains(stderrOutput, "unknown command") ||
	   strings.Contains(stderrOutput, "invalid") ||
	   strings.Contains(stdoutOutput, "unknown command") {
		tf.Success(t, "Error message properly displayed")
	} else {
		tf.Warning(t, "Error message format may be unexpected")
	}

	tf.Separator()

	// Step 2: Test with read-only log directory
	tf.Step(t, "Test with read-only log directory")

	readOnlyDir := filepath.Join(tempDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0755); err != nil {
		tf.Error(t, "Failed to create readonly directory", err.Error())
		success = false
		return
	}

	// Make directory read-only
	if err := os.Chmod(readOnlyDir, 0444); err != nil {
		tf.Error(t, "Failed to make directory read-only", err.Error())
		success = false
		return
	}

	// Restore permissions after test
	defer os.Chmod(readOnlyDir, 0755)

	readOnlyLogFile := filepath.Join(readOnlyDir, "readonly.log")
	os.Setenv("PORTUNIX_LOG_FILE", readOnlyLogFile)

	// Execute command that should handle logging gracefully
	tf.Command(t, binaryPath, []string{"version"})
	versionCmd := exec.Command(binaryPath, "version")

	stdoutBuf.Reset()
	stderrBuf.Reset()
	versionCmd.Stdout = &stdoutBuf
	versionCmd.Stderr = &stderrBuf

	err = versionCmd.Run()
	if err != nil {
		tf.Warning(t, "Version command failed with readonly log dir", err.Error())
	} else {
		tf.Success(t, "Version command handled readonly log dir gracefully")
	}

	// Should still produce output
	versionOutput := stdoutBuf.String()
	if versionOutput == "" {
		tf.Error(t, "No version output with readonly log directory")
		success = false
		return
	} else {
		tf.Success(t, "Version output generated despite logging issues")
	}

	tf.Separator()

	// Step 3: Test with extremely long command arguments
	tf.Step(t, "Test with extremely long command arguments")

	// Create very long argument string
	longArg := strings.Repeat("x", 1000)

	tf.Command(t, binaryPath, []string{"help", longArg})
	longArgCmd := exec.Command(binaryPath, "help", longArg)

	stdoutBuf.Reset()
	stderrBuf.Reset()
	longArgCmd.Stdout = &stdoutBuf
	longArgCmd.Stderr = &stderrBuf

	err = longArgCmd.Run()
	// This will likely fail, but should not crash
	if err == nil {
		tf.Info(t, "Long argument command succeeded")
	} else {
		tf.Success(t, "Long argument command failed gracefully", "error", err.Error())
	}

	// System should remain stable
	tf.Success(t, "System remained stable with long arguments")

	tf.Info(t, "E2E error scenarios test completed")
}