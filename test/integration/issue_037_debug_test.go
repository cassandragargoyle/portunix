package integration

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// DebugTestMCPServe provides detailed debug output for MCP serve testing
func TestDebugMCPServe(t *testing.T) {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🔍 DEBUG: Issue #037 MCP Serve Detailed Test")
	fmt.Println(strings.Repeat("=", 80))
	
	// Setup
	projectRoot := "../.."
	binaryName := "portunix"
	binaryPath := "./portunix"
	
	fmt.Printf("🔧 Binary path: %s\n", binaryPath)
	
	// Check binary exists
	if stat, err := os.Stat(binaryPath); os.IsNotExist(err) {
		fmt.Println("❌ Binary not found, building...")
		cmd := exec.Command("go", "build", "-o", binaryName, ".")
		cmd.Dir = projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("❌ Build failed: %v\n", err)
			fmt.Printf("Build output:\n%s\n", string(output))
			t.Fatalf("Failed to build binary")
		}
		fmt.Println("✅ Binary built successfully")
	} else {
		fmt.Printf("✅ Binary found (size: %d bytes, modified: %v)\n", 
			stat.Size(), stat.ModTime().Format(time.RFC3339))
	}
	
	// Test 1: Basic help
	fmt.Println("\n" + strings.Repeat("─", 40))
	fmt.Println("TEST 1: Basic Help Display")
	fmt.Println(strings.Repeat("─", 40))
	
	fmt.Printf("Executing: %s\n", binaryPath)
	cmd := exec.Command(binaryPath)
	
	// Set working directory
	cmd.Dir = projectRoot
	
	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	
	fmt.Printf("Exit error: %v\n", err)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Exit code: %d\n", exitError.ExitCode())
		}
	}
	
	outputStr := string(output)
	fmt.Printf("Output length: %d characters\n", len(outputStr))
	fmt.Printf("Output preview (first 300 chars):\n%s\n", 
		outputStr[:min(300, len(outputStr))])
	
	if strings.Contains(outputStr, "Usage:") {
		fmt.Println("✅ Contains 'Usage:'")
	} else {
		fmt.Println("❌ Missing 'Usage:'")
	}
	
	if strings.Contains(outputStr, "Available Commands:") {
		fmt.Println("✅ Contains 'Available Commands:'")
	} else {
		fmt.Println("❌ Missing 'Available Commands:'")
	}
	
	// Test 2: MCP command structure
	fmt.Println("\n" + strings.Repeat("─", 40))
	fmt.Println("TEST 2: MCP Command Structure")
	fmt.Println(strings.Repeat("─", 40))
	
	fmt.Printf("Executing: %s mcp --help\n", binaryPath)
	cmd2 := exec.Command(binaryPath, "mcp", "--help")
	cmd2.Dir = projectRoot
	
	output2, err2 := cmd2.CombinedOutput()
	fmt.Printf("Exit error: %v\n", err2)
	
	outputStr2 := string(output2)
	fmt.Printf("MCP help length: %d characters\n", len(outputStr2))
	fmt.Printf("MCP help content:\n%s\n", outputStr2)
	
	if strings.Contains(outputStr2, "serve") {
		fmt.Println("✅ 'serve' command found in MCP help")
	} else {
		fmt.Println("❌ 'serve' command NOT found in MCP help")
	}
	
	// Test 3: MCP serve help
	fmt.Println("\n" + strings.Repeat("─", 40))
	fmt.Println("TEST 3: MCP Serve Help")
	fmt.Println(strings.Repeat("─", 40))
	
	fmt.Printf("Executing: %s mcp serve --help\n", binaryPath)
	cmd3 := exec.Command(binaryPath, "mcp", "serve", "--help")
	cmd3.Dir = projectRoot
	
	output3, err3 := cmd3.CombinedOutput()
	fmt.Printf("Exit error: %v\n", err3)
	
	outputStr3 := string(output3)
	fmt.Printf("MCP serve help length: %d characters\n", len(outputStr3))
	fmt.Printf("MCP serve help content:\n%s\n", outputStr3)
	
	expectedFlags := []string{"--mode", "--port", "--socket", "--permissions"}
	for _, flag := range expectedFlags {
		if strings.Contains(outputStr3, flag) {
			fmt.Printf("✅ Flag '%s' found\n", flag)
		} else {
			fmt.Printf("❌ Flag '%s' NOT found\n", flag)
		}
	}
	
	// Test 4: Quick MCP serve test (with timeout)
	fmt.Println("\n" + strings.Repeat("─", 40))
	fmt.Println("TEST 4: Quick MCP Serve Test")
	fmt.Println(strings.Repeat("─", 40))
	
	fmt.Printf("Executing: timeout 2s %s mcp serve --mode stdio\n", binaryPath)
	cmd4 := exec.Command("timeout", "2s", binaryPath, "mcp", "serve", "--mode", "stdio")
	cmd4.Dir = projectRoot
	
	output4, err4 := cmd4.CombinedOutput()
	fmt.Printf("Exit error: %v\n", err4)
	
	outputStr4 := string(output4)
	fmt.Printf("MCP serve output length: %d characters\n", len(outputStr4))
	fmt.Printf("MCP serve output:\n%s\n", outputStr4)
	
	if strings.Contains(outputStr4, "Starting MCP Server") {
		fmt.Println("✅ MCP server started successfully")
	} else {
		fmt.Println("❌ MCP server did NOT start")
	}
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🎯 DEBUG TEST COMPLETED")
	fmt.Println(strings.Repeat("=", 80))
}

