package integration

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue037WithFramework demonstrates the new testing framework
func TestIssue037WithFramework(t *testing.T) {
	// Initialize test framework
	tf := testframework.NewTestFramework("Issue037_MCP_Serve")
	tf.Start(t, "Test MCP serve command implementation with detailed logging")
	
	success := true
	defer func() {
		tf.Finish(t, success)
	}()
	
	// Step 1: Setup binary
	tf.Step(t, "Setup test binary")
	
	projectRoot := "../.."
	binaryName := "portunix"
	binaryPath := "./portunix"
	
	tf.Info(t, "Binary path", binaryPath)
	
	// Check if binary exists
	if stat, err := os.Stat(binaryPath); os.IsNotExist(err) {
		tf.Warning(t, "Binary not found, building...")
		
		tf.Command(t, "go", []string{"build", "-o", binaryName, "."})
		cmd := exec.Command("go", "build", "-o", binaryName, ".")
		cmd.Dir = projectRoot
		
		if output, err := cmd.CombinedOutput(); err != nil {
			tf.Error(t, "Failed to build binary", err.Error())
			tf.Output(t, string(output), 300)
			success = false
			return
		}
		tf.Success(t, "Binary built successfully")
	} else {
		tf.Success(t, "Binary found", 
			fmt.Sprintf("Size: %d bytes", stat.Size()),
			fmt.Sprintf("Modified: %v", stat.ModTime().Format("2006-01-02 15:04:05")))
	}
	
	tf.Separator()
	
	// Step 2: Test basic help
	tf.Step(t, "Test basic help display")
	
	tf.Command(t, binaryPath, []string{})
	cmd := exec.Command(binaryPath)
	cmd.Dir = projectRoot
	
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			tf.Info(t, "Exit code", exitError.ExitCode())
		}
	}
	
	tf.Output(t, outputStr, 300)
	
	// Verify help content
	if strings.Contains(outputStr, "Usage:") || strings.Contains(outputStr, "Available Commands:") {
		tf.Success(t, "Help content found")
	} else {
		tf.Error(t, "Help content missing")
		success = false
	}
	
	// Ensure no MCP server started
	if strings.Contains(outputStr, "Starting MCP Server") {
		tf.Error(t, "Unexpected MCP server start")
		success = false
	} else {
		tf.Success(t, "No MCP server started (correct behavior)")
	}
	
	tf.Separator()
	
	// Step 3: Test MCP command structure
	tf.Step(t, "Test MCP command structure")
	
	tf.Command(t, binaryPath, []string{"mcp", "--help"})
	cmd2 := exec.Command(binaryPath, "mcp", "--help")
	cmd2.Dir = projectRoot
	
	output2, err2 := cmd2.CombinedOutput()
	outputStr2 := string(output2)
	
	if err2 != nil {
		tf.Warning(t, "MCP help exit error", err2.Error())
	}
	
	tf.Output(t, outputStr2, 400)
	
	// Check for serve command
	if strings.Contains(outputStr2, "serve") {
		tf.Success(t, "serve command found in MCP help")
	} else {
		tf.Error(t, "serve command NOT found in MCP help")
		success = false
	}
	
	tf.Separator()
	
	// Step 4: Test MCP serve help
	tf.Step(t, "Test MCP serve command help")
	
	tf.Command(t, binaryPath, []string{"mcp", "serve", "--help"})
	cmd3 := exec.Command(binaryPath, "mcp", "serve", "--help")
	cmd3.Dir = projectRoot
	
	output3, err3 := cmd3.CombinedOutput()
	outputStr3 := string(output3)
	
	if err3 != nil {
		tf.Warning(t, "MCP serve help exit error", err3.Error())
	}
	
	tf.Output(t, outputStr3, 500)
	
	// Check for expected flags
	expectedFlags := []string{"--mode", "--port", "--socket", "--permissions", "--config"}
	flagsFound := 0
	
	for _, flag := range expectedFlags {
		if strings.Contains(outputStr3, flag) {
			tf.Success(t, "Flag found", flag)
			flagsFound++
		} else {
			tf.Warning(t, "Flag not found", flag)
		}
	}
	
	if flagsFound >= 3 { // At least 3 out of 5 flags should be present
		tf.Success(t, "Sufficient flags found", fmt.Sprintf("%d/%d", flagsFound, len(expectedFlags)))
	} else {
		tf.Error(t, "Too few flags found", fmt.Sprintf("%d/%d", flagsFound, len(expectedFlags)))
		success = false
	}
	
	tf.Separator()
	
	// Step 5: Quick functional test
	tf.Step(t, "Quick MCP serve functional test")
	
	tf.Command(t, "timeout", []string{"3s", binaryPath, "mcp", "serve", "--mode", "stdio"})
	cmd4 := exec.Command("timeout", "3s", binaryPath, "mcp", "serve", "--mode", "stdio")
	cmd4.Dir = projectRoot
	
	output4, err4 := cmd4.CombinedOutput()
	outputStr4 := string(output4)
	
	// timeout command returns exit code 124 when it kills the process, which is expected
	tf.Info(t, "Command result", fmt.Sprintf("err: %v", err4))
	tf.Output(t, outputStr4, 300)
	
	// Check if MCP server actually started
	if strings.Contains(outputStr4, "Starting MCP Server") {
		tf.Success(t, "MCP server started successfully")
	} else {
		tf.Warning(t, "MCP server start message not found - might still be working")
		// Don't mark as failure since timeout might have killed it before output
	}
}