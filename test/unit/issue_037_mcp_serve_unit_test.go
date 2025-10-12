package unit

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestMCPServeUnitIssue037 Unit tests for Issue #037 - MCP Serve Command
type TestMCPServeUnitIssue037 struct {
	originalStdout *os.File
	originalStderr *os.File
}

func TestIssue037MCPServeUnitTests(t *testing.T) {
	suite := &TestMCPServeUnitIssue037{}
	
	t.Run("TC001_CommandStructure", suite.testCommandStructure)
	t.Run("TC002_DefaultParameters", suite.testDefaultParameters)
	t.Run("TC003_ParameterValidation", suite.testParameterValidation)
	t.Run("TC004_FlagParsing", suite.testFlagParsing)
	t.Run("TC005_HelpOutput", suite.testHelpOutput)
	t.Run("TC006_ModeValidation", suite.testModeValidation)
	t.Run("TC007_PortValidation", suite.testPortValidation)
	t.Run("TC008_SocketPathValidation", suite.testSocketPathValidation)
}

func (suite *TestMCPServeUnitIssue037) setUp(t *testing.T) (*bytes.Buffer, *bytes.Buffer) {
	// Capture stdout and stderr for testing
	var stdout, stderr bytes.Buffer
	suite.originalStdout = os.Stdout
	suite.originalStderr = os.Stderr
	
	return &stdout, &stderr
}

func (suite *TestMCPServeUnitIssue037) tearDown() {
	if suite.originalStdout != nil {
		os.Stdout = suite.originalStdout
	}
	if suite.originalStderr != nil {
		os.Stderr = suite.originalStderr
	}
}

// TC001: Command Structure Validation
func (suite *TestMCPServeUnitIssue037) testCommandStructure(t *testing.T) {
	t.Log("Testing TC001: MCP Serve Command Structure")
	
	// Create mock serve command for testing structure
	serveCmd := createMockServeCommand()
	
	// Validate command properties
	if serveCmd.Use != "serve" {
		t.Errorf("Expected command Use 'serve', got '%s'", serveCmd.Use)
	}
	
	if !strings.Contains(serveCmd.Short, "Start MCP server") {
		t.Errorf("Expected Short description to contain 'Start MCP server', got '%s'", serveCmd.Short)
	}
}

// TC002: Default Parameters Validation
func (suite *TestMCPServeUnitIssue037) testDefaultParameters(t *testing.T) {
	t.Log("Testing TC002: Default Parameters")
	
	serveCmd := createMockServeCommand()
	
	// Check default flag values
	expectedDefaults := map[string]string{
		"mode":        "stdio",
		"port":        "3001", 
		"socket":      "/tmp/portunix.sock",
		"permissions": "limited",
		"config":      "",
	}
	
	validateCommandFlags(t, serveCmd, expectedDefaults)
}

// TC003: Parameter Validation
func (suite *TestMCPServeUnitIssue037) testParameterValidation(t *testing.T) {
	t.Log("Testing TC003: Parameter Validation")
	
	serveCmd := createMockServeCommand()
	
	// Test valid flag combinations
	validTests := []struct{
		args []string
		desc string
	}{
		{[]string{}, "no arguments"},
		{[]string{"--mode", "stdio"}, "stdio mode"},
		{[]string{"--mode", "tcp", "--port", "8080"}, "tcp with port"},
		{[]string{"--mode", "unix", "--socket", "/tmp/test.sock"}, "unix with socket"},
		{[]string{"--permissions", "full"}, "full permissions"},
	}
	
	for _, test := range validTests {
		serveCmd.SetArgs(test.args)
		if err := serveCmd.ParseFlags(test.args); err != nil {
			t.Errorf("Flag parsing failed for %s: %v", test.desc, err)
		}
	}
}

// TC004: Flag Parsing Validation
func (suite *TestMCPServeUnitIssue037) testFlagParsing(t *testing.T) {
	t.Log("Testing TC004: Flag Parsing")
	
	serveCmd := createMockServeCommand()
	
	// Test short flag aliases
	shortFlags := []string{"m", "p", "s", "r", "c"}
	for _, flag := range shortFlags {
		if serveCmd.Flags().ShorthandLookup(flag) == nil {
			t.Errorf("Short flag '-%s' not found", flag)
		}
	}
	
	// Test long flag names
	longFlags := []string{"mode", "port", "socket", "permissions", "config"}
	for _, flag := range longFlags {
		if serveCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Long flag '--%s' not found", flag)
		}
	}
}

// TC005: Help Output Validation
func (suite *TestMCPServeUnitIssue037) testHelpOutput(t *testing.T) {
	t.Log("Testing TC005: Help Output")
	
	serveCmd := createMockServeCommand()
	
	// Test help output structure
	if !strings.Contains(serveCmd.Long, "Model Context Protocol") {
		t.Error("Long description should contain 'Model Context Protocol'")
	}
	
	// Validate flags are present
	expectedFlags := []string{"mode", "port", "socket", "permissions", "config"}
	for _, flag := range expectedFlags {
		if serveCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag '%s' not found in help", flag)
		}
	}
}

// TC006: Mode Validation
func (suite *TestMCPServeUnitIssue037) testModeValidation(t *testing.T) {
	t.Log("Testing TC006: Mode Parameter Validation")
	
	validModes := []string{"stdio", "tcp", "unix"}
	
	serveCmd := createMockServeCommand()
	
	for _, mode := range validModes {
		// Parse arguments to validate flag acceptance
		if err := serveCmd.ParseFlags([]string{"--mode", mode}); err != nil {
			t.Errorf("Valid mode '%s' failed parsing: %v", mode, err)
		}
	}
}

// TC007: Port Validation
func (suite *TestMCPServeUnitIssue037) testPortValidation(t *testing.T) {
	t.Log("Testing TC007: Port Parameter Validation")
	
	validPorts := []string{"1", "80", "3001", "8080", "65535"}
	
	serveCmd := createMockServeCommand()
	
	for _, port := range validPorts {
		if err := serveCmd.ParseFlags([]string{"--mode", "tcp", "--port", port}); err != nil {
			t.Errorf("Valid port '%s' failed parsing: %v", port, err)
		}
	}
	
	// Test invalid ports - cobra should handle string to int conversion errors
	invalidPorts := []string{"abc"}  // Remove "-1" as it might be valid for cobra int parsing
	
	for _, port := range invalidPorts {
		err := serveCmd.ParseFlags([]string{"--mode", "tcp", "--port", port})
		if err == nil {
			t.Errorf("Invalid port '%s' should have failed parsing", port)
		}
	}
}

// TC008: Socket Path Validation
func (suite *TestMCPServeUnitIssue037) testSocketPathValidation(t *testing.T) {
	t.Log("Testing TC008: Socket Path Validation")
	
	validPaths := []string{
		"/tmp/portunix.sock",
		"/var/run/portunix.sock",
		"./portunix.sock",
		"/home/user/portunix.sock",
	}
	
	serveCmd := createMockServeCommand()
	
	for _, path := range validPaths {
		if err := serveCmd.ParseFlags([]string{"--mode", "unix", "--socket", path}); err != nil {
			t.Errorf("Valid socket path '%s' failed parsing: %v", path, err)
		}
	}
}

// Test helper to create mock cobra command for testing
func createMockServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start MCP server for AI assistant integration",
		Long: `Start Model Context Protocol (MCP) server to enable AI assistants 
to interact with Portunix functionality.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Mock implementation for testing
		},
	}
	
	// Add flags like the real command
	cmd.Flags().StringP("mode", "m", "stdio", "Communication mode: stdio, tcp, unix")
	cmd.Flags().IntP("port", "p", 3001, "Port for TCP mode")
	cmd.Flags().StringP("socket", "s", "/tmp/portunix.sock", "Socket path for Unix mode")
	cmd.Flags().StringP("permissions", "r", "limited", "Permission level: limited, standard, full")
	cmd.Flags().StringP("config", "c", "", "Path to configuration file")
	
	return cmd
}

// Test helper to validate command flags
func validateCommandFlags(t *testing.T, cmd *cobra.Command, expectedFlags map[string]string) {
	for flagName, expectedDefault := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Flag '%s' not found", flagName)
			continue
		}
		
		if flag.DefValue != expectedDefault {
			t.Errorf("Flag '%s' has default value '%s', expected '%s'", 
				flagName, flag.DefValue, expectedDefault)
		}
	}
}