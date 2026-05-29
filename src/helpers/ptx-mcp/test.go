/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// clientProtocolVersion is the MCP protocol version advertised by the test
// client. Kept in sync with handleInitialize in app/mcp/handlers.go.
const clientProtocolVersion = "2024-11-05"

// defaultHandshakeTimeout is used when no explicit timeout is configured.
const defaultHandshakeTimeout = 10 * time.Second

// mcpInitializeRequest is the JSON-RPC request sent to negotiate the protocol.
type mcpInitializeRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// mcpInitializeResponse is the expected JSON-RPC response shape.
type mcpInitializeResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		ProtocolVersion string                 `json:"protocolVersion"`
		Capabilities    map[string]interface{} `json:"capabilities"`
		ServerInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// MCPHandshakeResult holds the negotiated protocol version and server info.
type MCPHandshakeResult struct {
	ProtocolVersion string
	ServerName      string
	ServerVersion   string
}

// buildInitializeRequest constructs the JSON-RPC initialize request body.
func buildInitializeRequest() ([]byte, error) {
	req := mcpInitializeRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": clientProtocolVersion,
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "portunix-mcp-test",
				"version": "1.0.0",
			},
		},
	}
	return json.Marshal(req)
}

// parseInitializeResponse decodes a single JSON-RPC initialize response.
func parseInitializeResponse(raw []byte) (*MCPHandshakeResult, error) {
	var resp mcpInitializeResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("server error: %s (code %d)", resp.Error.Message, resp.Error.Code)
	}
	if resp.Result.ProtocolVersion == "" {
		return nil, fmt.Errorf("server did not return protocolVersion")
	}
	return &MCPHandshakeResult{
		ProtocolVersion: resp.Result.ProtocolVersion,
		ServerName:      resp.Result.ServerInfo.Name,
		ServerVersion:   resp.Result.ServerInfo.Version,
	}, nil
}

// performStdioHandshake spawns the given command as a stdio MCP server,
// sends a JSON-RPC initialize request, reads exactly one response, then
// terminates the process. Used to verify that a configured stdio server
// speaks the MCP protocol.
func performStdioHandshake(executable string, args []string, env map[string]string, timeout time.Duration) (*MCPHandshakeResult, error) {
	if timeout <= 0 {
		timeout = defaultHandshakeTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, executable, args...)
	if len(env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	// Discard stderr - server diagnostics are not useful for a probe.
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start server: %w", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	reqBytes, err := buildInitializeRequest()
	if err != nil {
		return nil, err
	}
	if _, err := stdin.Write(append(reqBytes, '\n')); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	respCh := make(chan []byte, 1)
	errCh := make(chan error, 1)
	go func() {
		reader := bufio.NewReader(stdout)
		line, err := reader.ReadBytes('\n')
		if err != nil && len(line) == 0 {
			errCh <- err
			return
		}
		respCh <- line
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("handshake timed out after %s", timeout)
	case err := <-errCh:
		return nil, fmt.Errorf("read response: %w", err)
	case raw := <-respCh:
		return parseInitializeResponse(raw)
	}
}

// performNetworkHandshake performs an MCP initialize handshake against a
// running HTTP(S) MCP server.
func performNetworkHandshake(protocol, bindAddress string, port int, timeout time.Duration) (*MCPHandshakeResult, error) {
	if timeout <= 0 {
		timeout = defaultHandshakeTimeout
	}
	host := bindAddress
	if host == "" || host == "0.0.0.0" {
		host = "localhost"
	}
	scheme := "http"
	if protocol == "https" || protocol == "wss" {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s:%d", scheme, host, port)

	reqBytes, err := buildInitializeRequest()
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{}
	if scheme == "https" {
		// Self-signed certificates are common in local development.
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &http.Client{Timeout: timeout, Transport: transport}

	resp, err := client.Post(url, "application/json", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return parseInitializeResponse(buf.Bytes())
}

// handshakeAssistant runs an MCP initialize handshake appropriate for the
// configured assistant. For stdio assistants it spawns the local portunix
// MCP server; for remote assistants it dials the configured port.
//
// The assistant must be present in config.Assistants — passing an unknown
// name returns an error rather than silently falling back to the default
// server type.
func handshakeAssistant(assistant string, config *MCPConfiguration) (*MCPHandshakeResult, error) {
	timeout := defaultHandshakeTimeout
	if config != nil && config.Timeout != "" {
		if d, err := time.ParseDuration(config.Timeout); err == nil {
			timeout = d
		}
	}

	serverType := ""
	if config != nil {
		for _, a := range config.Assistants {
			if a.Name == assistant {
				serverType = a.ServerType
				break
			}
		}
	}
	if serverType == "" {
		return nil, fmt.Errorf("assistant %q is not configured (run: portunix mcp init --assistant %s)", assistant, assistant)
	}

	switch serverType {
	case "stdio":
		portunixPath, err := getPortunixExecutablePath()
		if err != nil {
			return nil, fmt.Errorf("portunix executable not found: %w", err)
		}
		var env map[string]string
		if config != nil {
			env = config.Env
		}
		return performStdioHandshake(portunixPath, []string{"mcp", "serve", "--mode", "stdio"}, env, timeout)
	case "remote":
		if config == nil {
			return nil, fmt.Errorf("remote handshake requires configuration")
		}
		protocol := config.Protocol
		if protocol == "" {
			protocol = "http"
		}
		return performNetworkHandshake(protocol, config.BindAddress, config.Port, timeout)
	default:
		return nil, fmt.Errorf("unsupported server type: %s", serverType)
	}
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test MCP server connection with AI assistants",
	Long: `Test the MCP server connection and verify integration with AI assistants.

This command performs a real JSON-RPC initialize handshake against the
configured MCP server and reports the negotiated protocol version.

Examples:
  portunix mcp test                         # Test all configured assistants
  portunix mcp test --assistant claude-code # Test specific assistant
  portunix mcp test --verbose              # Show detailed test output`,
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

	// Step 1: Load configuration
	fmt.Print("\n1. Loading MCP configuration... ")
	config, err := loadMCPConfiguration()
	if err != nil {
		fmt.Println("❌ NO CONFIG")
		fmt.Println("   Configure with: portunix mcp init")
		return fmt.Errorf("no configuration found")
	}
	fmt.Println("✅ LOADED")

	// Step 2: Determine which assistants to probe
	var targets []string
	if assistant != "" {
		targets = []string{assistant}
	} else {
		for _, a := range config.Assistants {
			targets = append(targets, a.Name)
		}
	}
	if len(targets) == 0 {
		fmt.Println("⚠️  No assistants to test")
		fmt.Println("   Configure assistants with: portunix mcp init")
		return nil
	}

	// Step 3: Run handshake for each target
	fmt.Println("\n2. Performing MCP handshake...")
	var failed []string
	for _, name := range targets {
		fmt.Printf("   %s: ", getAssistantDisplayName(name))
		result, err := handshakeAssistant(name, config)
		if err != nil {
			fmt.Printf("❌ FAILED (%v)\n", err)
			failed = append(failed, name)
			continue
		}
		fmt.Printf("✅ protocol=%s", result.ProtocolVersion)
		if result.ServerName != "" {
			fmt.Printf(", server=%s", result.ServerName)
		}
		if result.ServerVersion != "" {
			fmt.Printf("/%s", result.ServerVersion)
		}
		fmt.Println()
	}

	// Step 4: For Claude Code, also verify CLI registration (best-effort)
	if verbose {
		for _, name := range targets {
			if name == "claude-code" && isClaudeCodeInstalled() {
				fmt.Println("\n3. Verifying Claude Code CLI registration...")
				if err := verifyClaudeCodeRegistration(verbose); err != nil {
					fmt.Printf("   ⚠️  %v\n", err)
				} else {
					fmt.Println("   ✅ Registered in 'claude mcp list'")
				}
			}
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("handshake failed for: %s", strings.Join(failed, ", "))
	}
	return nil
}

// verifyClaudeCodeRegistration checks that portunix appears in `claude mcp list`.
func verifyClaudeCodeRegistration(verbose bool) error {
	claudePath, err := getClaudePath()
	if err != nil {
		return fmt.Errorf("claude executable not found")
	}
	cmd := exec.Command(claudePath, "mcp", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("claude mcp list failed: %w", err)
	}
	if !contains(string(output), "portunix") {
		return fmt.Errorf("portunix not found in Claude Code MCP servers")
	}
	if verbose {
		detail := exec.Command(claudePath, "mcp", "get", "portunix")
		if out, err := detail.Output(); err == nil {
			fmt.Println("   Claude Code registration:")
			fmt.Println("   " + strings.ReplaceAll(string(out), "\n", "\n   "))
		}
	}
	return nil
}

func init() {
	mcpCmd.AddCommand(testCmd)

	testCmd.Flags().String("assistant", "", "Test specific assistant (claude-code, claude-desktop, gemini-cli)")
	testCmd.Flags().BoolP("verbose", "v", false, "Show detailed test output")
}
