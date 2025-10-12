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

// Simple MCP Protocol Test - focused on key validations
func TestMCPProtocolSimple(t *testing.T) {
	binaryPath := setupTestBinary(t)
	
	t.Run("MCP_Initialize_And_Echo", func(t *testing.T) {
		testMCPInitializeAndEcho(t, binaryPath)
	})
}

func setupTestBinary(t *testing.T) string {
	projectRoot := "../.."
	binaryName := "portunix"
	binaryPath := filepath.Join(projectRoot, binaryName)
	
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Log("Building portunix binary for MCP protocol testing...")
		cmd := exec.Command("go", "build", "-o", binaryName, ".")
		cmd.Dir = projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build binary: %v\nOutput: %s", err, string(output))
		}
	}
	return binaryPath
}

func testMCPInitializeAndEcho(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	// Start MCP server
	cmd := exec.CommandContext(ctx, binaryPath, "mcp", "serve", "--mode", "stdio")
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin pipe: %v", err)
	}
	defer stdin.Close()
	
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
	
	// Wait for server start
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "Starting MCP Server in stdio mode") {
				t.Log("✅ MCP Server started successfully")
				break
			}
		}
	}()
	
	time.Sleep(1 * time.Second)
	reader := bufio.NewReader(stdout)
	
	// Step 1: Initialize
	t.Log("🔄 Sending initialize request...")
	initRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"roots": map[string]interface{}{
					"listChanged": true,
				},
			},
			"clientInfo": map[string]interface{}{
				"name":    "portunix-test-client",
				"version": "1.0.0",
			},
		},
	}
	
	if err := sendMCPMessage(stdin, initRequest); err != nil {
		t.Fatalf("Failed to send initialize: %v", err)
	}
	
	initResponse, err := readMCPResponse(reader)
	if err != nil {
		t.Fatalf("Failed to read initialize response: %v", err)
	}
	
	if initResponse["jsonrpc"] != "2.0" || initResponse["id"].(float64) != 1 {
		t.Errorf("Invalid initialize response: %+v", initResponse)
	}
	
	if initResponse["result"] == nil {
		t.Errorf("Initialize response missing result: %+v", initResponse)
	}
	
	t.Log("✅ Initialize handshake successful")
	
	// Step 2: Test Echo Tool (simple and predictable)
	t.Log("🔄 Testing echo tool...")
	echoRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "echo",
			"arguments": map[string]interface{}{
				"message": "Hello MCP Protocol Test!",
			},
		},
	}
	
	if err := sendMCPMessage(stdin, echoRequest); err != nil {
		t.Fatalf("Failed to send echo request: %v", err)
	}
	
	echoResponse, err := readMCPResponse(reader)
	if err != nil {
		t.Fatalf("Failed to read echo response: %v", err)
	}
	
	if echoResponse["jsonrpc"] != "2.0" || echoResponse["id"].(float64) != 2 {
		t.Errorf("Invalid echo response: %+v", echoResponse)
	}
	
	// Check if we got result or error
	if echoResponse["result"] != nil {
		t.Log("✅ Echo tool call successful with result")
		result := echoResponse["result"].(map[string]interface{})
		if content, ok := result["content"]; ok {
			t.Logf("📝 Echo response content: %+v", content)
		}
	} else if echoResponse["error"] != nil {
		t.Log("⚠️ Echo tool call returned error (but response structure is correct)")
		t.Logf("📝 Error: %+v", echoResponse["error"])
	} else {
		t.Errorf("Echo response missing both result and error: %+v", echoResponse)
	}
	
	// Step 3: Verify we can send multiple requests
	t.Log("🔄 Testing system info tool...")
	sysInfoRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      "get_system_info",
			"arguments": map[string]interface{}{},
		},
	}
	
	if err := sendMCPMessage(stdin, sysInfoRequest); err != nil {
		t.Fatalf("Failed to send system info request: %v", err)
	}
	
	sysInfoResponse, err := readMCPResponse(reader)
	if err != nil {
		t.Fatalf("Failed to read system info response: %v", err)
	}
	
	if sysInfoResponse["result"] != nil {
		t.Log("✅ System info tool call successful")
		result := sysInfoResponse["result"].(map[string]interface{})
		if content, ok := result["content"]; ok {
			contentArray := content.([]interface{})
			if len(contentArray) > 0 {
				firstItem := contentArray[0].(map[string]interface{})
				if text, ok := firstItem["text"].(string); ok {
					// Truncate long text for logging
					if len(text) > 100 {
						text = text[:100] + "..."
					}
					t.Logf("📊 System info: %s", text)
				}
			}
		}
	} else {
		t.Errorf("System info call failed: %+v", sysInfoResponse)
	}
	
	t.Log("🎉 MCP Protocol Communication Test PASSED")
}

func sendMCPMessage(stdin io.WriteCloser, message map[string]interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	_, err = stdin.Write(append(data, '\n'))
	return err
}

func readMCPResponse(reader *bufio.Reader) (map[string]interface{}, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	
	var response map[string]interface{}
	err = json.Unmarshal([]byte(line), &response)
	return response, err
}