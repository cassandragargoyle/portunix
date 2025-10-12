package testframework

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestFramework provides verbose logging and standardized test execution
type TestFramework struct {
	verbose    bool
	testName   string
	startTime  time.Time
	stepCount  int
}

// NewTestFramework creates a new test framework instance
func NewTestFramework(testName string) *TestFramework {
	// Use Go's built-in verbose detection
	verbose := testing.Verbose()
	
	return &TestFramework{
		verbose:   verbose,
		testName:  testName,
		startTime: time.Now(),
		stepCount: 0,
	}
}

// Start begins a test with optional verbose header
func (tf *TestFramework) Start(t *testing.T, description string) {
	tf.stepCount = 0
	tf.startTime = time.Now()
	
	if tf.verbose {
		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Printf("🚀 STARTING: %s\n", tf.testName)
		fmt.Printf("Description: %s\n", description)
		fmt.Printf("Time: %s\n", tf.startTime.Format(time.RFC3339))
		fmt.Printf("%s\n\n", strings.Repeat("=", 80))
	}
	
	t.Logf("Starting %s: %s", tf.testName, description)
}

// Step logs a test step with optional verbose details
func (tf *TestFramework) Step(t *testing.T, stepDescription string, details ...interface{}) {
	tf.stepCount++
	
	if tf.verbose {
		fmt.Printf("📋 STEP %d: %s\n", tf.stepCount, stepDescription)
		if len(details) > 0 {
			for _, detail := range details {
				fmt.Printf("   %v\n", detail)
			}
		}
	}
	
	t.Logf("Step %d: %s", tf.stepCount, stepDescription)
}

// Success logs a successful step
func (tf *TestFramework) Success(t *testing.T, message string, details ...interface{}) {
	if tf.verbose {
		fmt.Printf("   ✅ %s\n", message)
		for _, detail := range details {
			fmt.Printf("      %v\n", detail)
		}
	}
	t.Logf("✅ %s", message)
}

// Error logs an error with details
func (tf *TestFramework) Error(t *testing.T, message string, details ...interface{}) {
	if tf.verbose {
		fmt.Printf("   ❌ %s\n", message)
		for _, detail := range details {
			fmt.Printf("      %v\n", detail)
		}
	}
	t.Errorf("❌ %s", message)
}

// Warning logs a warning
func (tf *TestFramework) Warning(t *testing.T, message string, details ...interface{}) {
	if tf.verbose {
		fmt.Printf("   ⚠️  %s\n", message)
		for _, detail := range details {
			fmt.Printf("      %v\n", detail)
		}
	}
	t.Logf("⚠️ %s", message)
}

// Info logs informational message
func (tf *TestFramework) Info(t *testing.T, message string, details ...interface{}) {
	if tf.verbose {
		fmt.Printf("   %s\n", message)
		for _, detail := range details {
			fmt.Printf("      %v\n", detail)
		}
	} else {
		t.Logf(" %s", message)
	}
}

// Command logs command execution
func (tf *TestFramework) Command(t *testing.T, command string, args []string) {
	fullCmd := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	
	if tf.verbose {
		fmt.Printf("   🔧 Executing: %s\n", fullCmd)
	}
	t.Logf("Executing: %s", fullCmd)
}

// Output logs command output
func (tf *TestFramework) Output(t *testing.T, output string, maxLength int) {
	if maxLength == 0 {
		maxLength = 500
	}
	
	if tf.verbose {
		outputLen := len(output)
		fmt.Printf("   📄 Output (%d chars):\n", outputLen)
		
		if outputLen > maxLength {
			// Show first part and last part for better context
			if maxLength > 100 {
				// Show beginning and end
				firstPart := maxLength * 2 / 3
				lastPart := maxLength / 3
				fmt.Printf("      %s\n", output[:firstPart])
				fmt.Printf("      ... [truncated %d chars] ...\n", outputLen-maxLength)
				fmt.Printf("      %s\n", output[outputLen-lastPart:])
			} else {
				// Just show the end for very short limits
				fmt.Printf("      ...%s\n", output[outputLen-maxLength:])
			}
		} else {
			fmt.Printf("      %s\n", output)
		}
	} else {
		// Non-verbose mode - show last part of output for context
		outputLen := len(output)
		if outputLen > 200 {
			t.Logf("Output (%d chars): ...%s", outputLen, output[outputLen-200:])
		} else {
			t.Logf("Output (%d chars): %s", outputLen, output)
		}
	}
}

// Finish completes the test with summary
func (tf *TestFramework) Finish(t *testing.T, success bool) {
	duration := time.Since(tf.startTime)
	
	if tf.verbose {
		fmt.Printf("\n%s\n", strings.Repeat("-", 80))
		if success {
			fmt.Printf("🎉 COMPLETED: %s\n", tf.testName)
		} else {
			fmt.Printf("💥 FAILED: %s\n", tf.testName)
		}
		fmt.Printf("Duration: %v\n", duration)
		fmt.Printf("Steps: %d\n", tf.stepCount)
		fmt.Printf("%s\n\n", strings.Repeat("-", 80))
	}
	
	if success {
		t.Logf("✅ %s completed successfully in %v (%d steps)", tf.testName, duration, tf.stepCount)
	} else {
		t.Errorf("❌ %s failed after %v (%d steps)", tf.testName, duration, tf.stepCount)
	}
}

// IsVerbose returns whether verbose mode is enabled
func (tf *TestFramework) IsVerbose() bool {
	return tf.verbose
}

// Separator prints a visual separator in verbose mode
func (tf *TestFramework) Separator() {
	if tf.verbose {
		fmt.Println(strings.Repeat("─", 60))
	}
}

// VerifyBinary checks if Portunix binary exists and returns path or fails test
func (tf *TestFramework) VerifyBinary(t *testing.T, relativePath string) (string, bool) {
	tf.Step(t, "Verify Portunix binary exists")
	
	binaryPath, err := filepath.Abs(relativePath)
	if err != nil {
		tf.Error(t, "Failed to get binary path", err.Error())
		return "", false
	}
	
	tf.Info(t, "Binary path:", binaryPath)
	
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		tf.Error(t, "Portunix binary not found", binaryPath)
		return "", false
	}
	
	// Get file info for additional details
	if fileInfo, err := os.Stat(binaryPath); err == nil {
		tf.Success(t, "Portunix binary found")
		tf.Info(t, fmt.Sprintf("Size: %d bytes", fileInfo.Size()))
		tf.Info(t, fmt.Sprintf("Modified: %s", fileInfo.ModTime().Format("2006-01-02 15:04:05")))
	} else {
		tf.Success(t, "Portunix binary found")
	}
	
	return binaryPath, true
}

// VerifyPortunixBinary checks if Portunix binary exists using standard path and fails test if not found
func (tf *TestFramework) VerifyPortunixBinary(t *testing.T) (string, bool) {
	// Try standard locations for Portunix binary
	standardPaths := []string{
		"../../portunix",  // Most common for integration tests
		"./portunix",      // Current directory  
		"../portunix",     // Parent directory
	}
	
	tf.Step(t, "Verify Portunix binary exists")
	
	for _, relativePath := range standardPaths {
		binaryPath, err := filepath.Abs(relativePath)
		if err != nil {
			continue
		}
		
		if _, err := os.Stat(binaryPath); err == nil {
			// Found the binary
			tf.Info(t, "Binary path:", binaryPath)
			
			// Get file info for additional details
			if fileInfo, err := os.Stat(binaryPath); err == nil {
				tf.Success(t, "Portunix binary found")
				tf.Info(t, fmt.Sprintf("Size: %d bytes", fileInfo.Size()))
				tf.Info(t, fmt.Sprintf("Modified: %s", fileInfo.ModTime().Format("2006-01-02 15:04:05")))
			} else {
				tf.Success(t, "Portunix binary found")
			}
			
			return binaryPath, true
		}
	}
	
	// Binary not found in any standard location
	tf.Error(t, "Portunix binary not found in standard locations")
	tf.Info(t, "Searched paths:", strings.Join(standardPaths, ", "))
	tf.Info(t, "Make sure you have built the binary with: go build -o .")
	
	return "", false
}

// MustVerifyPortunixBinary verifies Portunix binary exists and returns path, or fails test immediately
func (tf *TestFramework) MustVerifyPortunixBinary(t *testing.T) string {
	binaryPath, ok := tf.VerifyPortunixBinary(t)
	if !ok {
		t.FailNow() // Immediately fail test if binary not found
	}
	return binaryPath
}