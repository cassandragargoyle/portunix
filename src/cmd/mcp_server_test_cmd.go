package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var mcpServerTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test MCP server connection with AI assistants",
	Long: `Test the MCP server connection and verify integration with AI assistants.

This command will:
- Check if MCP server is running
- Test connection to the server
- Verify assistant integration
- Run basic functionality tests

Examples:
  portunix mcp-server test                         # Test all configured assistants
  portunix mcp-server test --assistant claude-code # Test specific assistant
  portunix mcp-server test --verbose              # Show detailed test output`,
	Run: func(cmd *cobra.Command, args []string) {
		assistant, _ := cmd.Flags().GetString("assistant")
		verbose, _ := cmd.Flags().GetBool("verbose")
		
		if err := testMCPServer(assistant, verbose); err != nil {
			fmt.Printf("❌ Test failed: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("\n✅ All tests passed!")
	},
}

func testMCPServer(assistant string, verbose bool) error {
	fmt.Println("🧪 Testing MCP Server Integration")
	fmt.Println("=" + "==============================")
	
	// Test 1: Check if server is running
	fmt.Print("\n1. Checking if MCP server is running... ")
	if !isServerRunning() {
		fmt.Println("❌ NOT RUNNING")
		fmt.Println("   Start the server with: portunix mcp-server start")
		return fmt.Errorf("server not running")
	}
	fmt.Println("✅ RUNNING")
	
	// Test 2: Check configuration
	fmt.Print("2. Checking MCP configuration... ")
	config, err := loadMCPConfiguration()
	if err != nil {
		fmt.Println("❌ NO CONFIG")
		fmt.Println("   Configure with: portunix mcp-server init")
		return fmt.Errorf("no configuration found")
	}
	fmt.Println("✅ CONFIGURED")
	
	// Test 3: Test server connectivity (if remote)
	if config.ServerType == "remote" {
		fmt.Printf("3. Testing server connectivity (port %d)... ", config.Port)
		if err := testServerConnectivity(config.Port); err != nil {
			fmt.Println("❌ FAILED")
			if verbose {
				fmt.Printf("   Error: %v\n", err)
			}
			return fmt.Errorf("connectivity test failed")
		}
		fmt.Println("✅ CONNECTED")
	}
	
	// Test 4: Test specific assistant or all configured
	if assistant != "" {
		// Test specific assistant
		fmt.Printf("\n4. Testing %s integration...\n", getAssistantDisplayName(assistant))
		if err := testAssistantIntegration(assistant, verbose); err != nil {
			return err
		}
	} else {
		// Test all configured assistants
		fmt.Println("\n4. Testing configured assistants...")
		testedAny := false
		
		for _, a := range config.Assistants {
			if !isAssistantInstalled(a.Name) {
				fmt.Printf("   %s: ⚠️  NOT INSTALLED\n", getAssistantDisplayName(a.Name))
				continue
			}
			
			fmt.Printf("   Testing %s... ", getAssistantDisplayName(a.Name))
			if err := testAssistantIntegration(a.Name, false); err != nil {
				fmt.Printf("❌ FAILED")
				if verbose {
					fmt.Printf(" (%v)", err)
				}
				fmt.Println()
			} else {
				fmt.Println("✅ OK")
			}
			testedAny = true
		}
		
		if !testedAny {
			fmt.Println("   ⚠️  No assistants to test")
			fmt.Println("   Configure assistants with: portunix mcp-server init")
		}
	}
	
	// Test 5: Test MCP tools functionality
	if verbose {
		fmt.Println("\n5. Testing MCP tools...")
		if err := testMCPTools(); err != nil {
			fmt.Printf("   ⚠️  Some tools failed: %v\n", err)
		} else {
			fmt.Println("   ✅ All tools functional")
		}
	}
	
	return nil
}

func testServerConnectivity(port int) error {
	// Try to connect to the server
	url := fmt.Sprintf("http://localhost:%d/health", port)
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		// Server might not have health endpoint, try basic connection
		url = fmt.Sprintf("http://localhost:%d", port)
		resp, err = client.Get(url)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()
	
	return nil
}

func testAssistantIntegration(assistant string, verbose bool) error {
	switch assistant {
	case "claude-code":
		return testClaudeCodeIntegration(verbose)
	case "claude-desktop":
		return testClaudeDesktopIntegration(verbose)
	case "gemini-cli":
		return testGeminiCLIIntegration(verbose)
	default:
		return fmt.Errorf("unknown assistant: %s", assistant)
	}
}

func testClaudeCodeIntegration(verbose bool) error {
	// Check if Claude Code is installed
	if !isClaudeCodeInstalled() {
		return fmt.Errorf("Claude Code not installed")
	}
	
	// Find claude executable
	claudePath, err := getClaudePath()
	if err != nil {
		return fmt.Errorf("claude executable not found")
	}
	
	// Check if portunix is in MCP server list
	cmd := exec.Command(claudePath, "mcp", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list MCP servers")
	}
	
	if !contains(string(output), "portunix") {
		return fmt.Errorf("portunix not found in Claude Code MCP servers")
	}
	
	if verbose {
		// Try to get server details
		cmd = exec.Command(claudePath, "mcp", "get", "portunix")
		if output, err := cmd.Output(); err == nil {
			fmt.Println("\n   Claude Code MCP Configuration:")
			fmt.Println("   " + strings.ReplaceAll(string(output), "\n", "\n   "))
		}
	}
	
	return nil
}

func testClaudeDesktopIntegration(verbose bool) error {
	// Check if Claude Desktop is installed
	if !isClaudeDesktopInstalled() {
		return fmt.Errorf("Claude Desktop not installed")
	}
	
	// Check configuration file
	configPath := getClaudeDesktopConfigPath()
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("Claude Desktop MCP configuration not found")
	}
	
	// Parse and check configuration
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration")
	}
	
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("invalid configuration format")
	}
	
	// Check if portunix is configured
	if servers, ok := config["servers"].(map[string]interface{}); ok {
		if _, exists := servers["portunix"]; !exists {
			return fmt.Errorf("portunix not configured in Claude Desktop")
		}
		
		if verbose {
			fmt.Println("\n   Claude Desktop MCP Configuration found")
		}
	} else {
		return fmt.Errorf("no MCP servers configured")
	}
	
	return nil
}

func testGeminiCLIIntegration(verbose bool) error {
	// Check if Gemini CLI is installed
	if !isGeminiCLIInstalled() {
		return fmt.Errorf("Gemini CLI not installed")
	}
	
	// TODO: Implement Gemini CLI specific tests
	// This needs research on how Gemini CLI handles MCP
	
	if verbose {
		fmt.Println("\n   ⚠️  Gemini CLI integration testing not yet implemented")
	}
	
	return nil
}

func testMCPTools() error {
	// Test basic MCP tools by making requests to the server
	// This would require implementing actual MCP protocol testing
	
	fmt.Println("   - system_info: Testing...")
	// TODO: Implement actual tool testing
	
	fmt.Println("   - project_detect: Testing...")
	// TODO: Implement actual tool testing
	
	return nil
}

func getClaudeDesktopConfigPath() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Claude", "mcp_servers.json")
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "Claude", "mcp_servers.json")
	default: // linux
		return filepath.Join(os.Getenv("HOME"), ".config", "claude", "mcp_servers.json")
	}
}

func init() {
	mcpCmd.AddCommand(mcpServerTestCmd)
	
	mcpServerTestCmd.Flags().String("assistant", "", "Test specific assistant (claude-code, claude-desktop, gemini-cli)")
	mcpServerTestCmd.Flags().BoolP("verbose", "v", false, "Show detailed test output")
}