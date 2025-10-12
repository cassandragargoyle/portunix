package integration

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMCPProtocolCommunication tests actual MCP protocol communication
type TestMCPProtocolCommunication struct {
	binaryPath string
}

// MCPMessage represents a JSON-RPC 2.0 message for MCP protocol
type MCPMessage struct {
	JSONRpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// MCPInitializeParams represents MCP initialize request parameters
type MCPInitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

// ClientInfo represents client information in MCP initialize
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func TestIssue037MCPProtocolCommunication(t *testing.T) {
	suite := &TestMCPProtocolCommunication{}
	suite.setupBinary(t)
	
	t.Run("TC_MCP_01_InitializeHandshake", suite.testMCPInitializeHandshake)
	t.Run("TC_MCP_02_ListToolsRequest", suite.testMCPListToolsRequest)
	t.Run("TC_MCP_03_InvalidMessageHandling", suite.testMCPInvalidMessageHandling)
	t.Run("TC_MCP_04_CallToolRequest", suite.testMCPCallToolRequest)
}

func (suite *TestMCPProtocolCommunication) setupBinary(t *testing.T) {
	// Use existing binary or build if not exists
	projectRoot := "../.."
	binaryName := "portunix"
	suite.binaryPath = filepath.Join(projectRoot, binaryName)
	
	// Check if binary exists, if not build it
	if _, err := os.Stat(suite.binaryPath); os.IsNotExist(err) {
		t.Log("Building portunix binary for MCP protocol testing...")
		cmd := exec.Command("go", "build", "-o", binaryName, ".")
		cmd.Dir = projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build binary: %v\nOutput: %s", err, string(output))
		}
	}
}

// TC_MCP_01: Test MCP Initialize Handshake
func (suite *TestMCPProtocolCommunication) testMCPInitializeHandshake(t *testing.T) {
	t.Log("Testing MCP Initialize Handshake")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Start MCP server in stdio mode
	cmd := exec.CommandContext(ctx, suite.binaryPath, "mcp", "serve", "--mode", "stdio")
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin pipe: %v", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to get stderr pipe: %v", err)
	}
	
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start MCP server: %v", err)
	}
	defer cmd.Process.Kill()
	
	// Wait for server to start
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "Starting MCP Server in stdio mode") {
				t.Log("MCP Server started successfully")
				break
			}
		}
	}()
	
	time.Sleep(1 * time.Second)
	
	// Send MCP initialize request
	initRequest := MCPMessage{
		JSONRpc: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: MCPInitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities: map[string]interface{}{
				"roots": map[string]interface{}{
					"listChanged": true,
				},
			},
			ClientInfo: ClientInfo{
				Name:    "portunix-test-client",
				Version: "1.0.0",
			},
		},
	}
	
	// Send request
	requestBytes, _ := json.Marshal(initRequest)
	requestLine := string(requestBytes) + "\n"
	
	t.Logf("Sending initialize request: %s", requestLine)
	if _, err := stdin.Write([]byte(requestLine)); err != nil {
		t.Fatalf("Failed to send initialize request: %v", err)
	}
	
	// Read response with timeout
	responseChan := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(stdout)
		line, _, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				t.Logf("Failed to read response: %v", err)
			}
			return
		}
		responseChan <- string(line)
	}()
	
	select {
	case response := <-responseChan:
		t.Logf("Received response: %s", response)
		
		// Parse response
		var responseMsg MCPMessage
		if err := json.Unmarshal([]byte(response), &responseMsg); err != nil {
			t.Errorf("Failed to parse response JSON: %v", err)
			return
		}
		
		// Validate response
		if responseMsg.JSONRpc != "2.0" {
			t.Errorf("Expected JSON-RPC 2.0, got %s", responseMsg.JSONRpc)
		}
		
		if responseMsg.ID != 1 {
			t.Errorf("Expected ID 1, got %v", responseMsg.ID)
		}
		
		if responseMsg.Result == nil {
			t.Errorf("Expected result in initialize response, got nil")
		} else {
			t.Log("Initialize handshake successful")
		}
		
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for initialize response")
	}
}

// TC_MCP_02: Test MCP List Tools Request
func (suite *TestMCPProtocolCommunication) testMCPListToolsRequest(t *testing.T) {
	t.Log("Testing MCP List Tools Request")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Start MCP server
	cmd := exec.CommandContext(ctx, suite.binaryPath, "mcp", "serve", "--mode", "stdio")
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin pipe: %v", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}
	
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start MCP server: %v", err)
	}
	defer cmd.Process.Kill()
	
	time.Sleep(1 * time.Second)
	
	// Send initialize first
	suite.sendInitialize(t, stdin, stdout)
	
	// Send list tools request
	listToolsRequest := MCPMessage{
		JSONRpc: "2.0",
		ID:      2,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}
	
	requestBytes, _ := json.Marshal(listToolsRequest)
	requestLine := string(requestBytes) + "\n"
	
	t.Logf("Sending list tools request: %s", requestLine)
	if _, err := stdin.Write([]byte(requestLine)); err != nil {
		t.Fatalf("Failed to send list tools request: %v", err)
	}
	
	// Read response
	reader := bufio.NewReader(stdout)
	response, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read list tools response: %v", err)
	}
	
	t.Logf("Received tools list: %s", string(response))
	
	// Parse response
	var responseMsg MCPMessage
	if err := json.Unmarshal(response, &responseMsg); err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
		return
	}
	
	// Validate response
	if responseMsg.JSONRpc != "2.0" {
		t.Errorf("Expected JSON-RPC 2.0, got %s", responseMsg.JSONRpc)
	}
	
	if responseMsg.ID != 2 {
		t.Errorf("Expected ID 2, got %v", responseMsg.ID)
	}
	
	if responseMsg.Result == nil {
		t.Errorf("Expected tools list result, got nil")
	} else {
		t.Log("List tools request successful")
	}
}

// TC_MCP_03: Test Invalid Message Handling
func (suite *TestMCPProtocolCommunication) testMCPInvalidMessageHandling(t *testing.T) {
	t.Log("Testing MCP Invalid Message Handling")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Start MCP server
	cmd := exec.CommandContext(ctx, suite.binaryPath, "mcp", "serve", "--mode", "stdio")
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin pipe: %v", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}
	
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start MCP server: %v", err)
	}
	defer cmd.Process.Kill()
	
	time.Sleep(1 * time.Second)
	
	// Send invalid JSON
	invalidJSON := "{invalid json}\n"
	
	t.Logf("Sending invalid JSON: %s", invalidJSON)
	if _, err := stdin.Write([]byte(invalidJSON)); err != nil {
		t.Fatalf("Failed to send invalid JSON: %v", err)
	}
	
	// Read error response
	reader := bufio.NewReader(stdout)
	response, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read error response: %v", err)
	}
	
	t.Logf("Received error response: %s", string(response))
	
	// Parse response
	var responseMsg MCPMessage
	if err := json.Unmarshal(response, &responseMsg); err != nil {
		t.Errorf("Failed to parse error response JSON: %v", err)
		return
	}
	
	// Should contain error
	if responseMsg.Error == nil {
		t.Errorf("Expected error response for invalid JSON, got nil")
	} else {
		t.Log("Invalid message properly handled with error response")
	}
}

// TC_MCP_04: Test Call Tool Request
func (suite *TestMCPProtocolCommunication) testMCPCallToolRequest(t *testing.T) {
	t.Log("Testing MCP Call Tool Request")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Start MCP server
	cmd := exec.CommandContext(ctx, suite.binaryPath, "mcp", "serve", "--mode", "stdio")
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin pipe: %v", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}
	
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start MCP server: %v", err)
	}
	defer cmd.Process.Kill()
	
	time.Sleep(1 * time.Second)
	
	// Send initialize first
	suite.sendInitialize(t, stdin, stdout)
	
	// Send call tool request (using correct tool name from list)
	callToolRequest := MCPMessage{
		JSONRpc: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "get_system_info",
			"arguments": map[string]interface{}{},
		},
	}
	
	requestBytes, _ := json.Marshal(callToolRequest)
	requestLine := string(requestBytes) + "\n"
	
	t.Logf("Sending call tool request: %s", requestLine)
	if _, err := stdin.Write([]byte(requestLine)); err != nil {
		t.Fatalf("Failed to send call tool request: %v", err)
	}
	
	// Read response
	reader := bufio.NewReader(stdout)
	response, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read call tool response: %v", err)
	}
	
	t.Logf("Received tool response: %s", string(response))
	
	// Parse response
	var responseMsg MCPMessage
	if err := json.Unmarshal(response, &responseMsg); err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
		return
	}
	
	// Validate response structure
	if responseMsg.JSONRpc != "2.0" {
		t.Errorf("Expected JSON-RPC 2.0, got %s", responseMsg.JSONRpc)
	}
	
	if responseMsg.ID != 3 {
		t.Errorf("Expected ID 3, got %v", responseMsg.ID)
	}
	
	// Either result or error should be present
	if responseMsg.Result == nil && responseMsg.Error == nil {
		t.Errorf("Expected either result or error in call tool response")
	} else {
		t.Log("Call tool request completed (result or error received)")
	}
}

// Helper function to send initialize request
func (suite *TestMCPProtocolCommunication) sendInitialize(t *testing.T, stdin io.Writer, stdout io.Reader) {
	initRequest := MCPMessage{
		JSONRpc: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: MCPInitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities: map[string]interface{}{
				"roots": map[string]interface{}{
					"listChanged": true,
				},
			},
			ClientInfo: ClientInfo{
				Name:    "portunix-test-client",
				Version: "1.0.0",
			},
		},
	}
	
	requestBytes, _ := json.Marshal(initRequest)
	requestLine := string(requestBytes) + "\n"
	
	if _, err := stdin.Write([]byte(requestLine)); err != nil {
		t.Fatalf("Failed to send initialize request: %v", err)
	}
	
	// Read initialize response
	reader := bufio.NewReader(stdout)
	_, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read initialize response: %v", err)
	}
}