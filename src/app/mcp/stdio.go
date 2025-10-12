package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// RunStdioMode runs Portunix in MCP stdio mode for direct AI assistant integration
func RunStdioMode() {
	// IMPORTANT: No output to stdout except JSON-RPC messages!
	// All logs/info must go to stderr
	
	// Check if running in interactive terminal
	if isInteractiveTerm() && !isMCPClient() {
		// Show brief message to stderr for interactive users
		fmt.Fprintf(os.Stderr, "MCP stdio mode active. Press Ctrl+C to exit. Use 'portunix --help' for help.\n")
	}

	// Create server with default configuration
	server := NewServer(0, "development", "") // Port 0 since we're not using network

	// Start stdio mode - this will handle newline-delimited JSON-RPC
	if err := server.StartStdioNewline(); err != nil {
		fmt.Fprintf(os.Stderr, "Error in MCP stdio mode: %v\n", err)
		os.Exit(1)
	}
}

// isInteractiveTerm checks if the program is running in an interactive terminal
func isInteractiveTerm() bool {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}

	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}

	return true
}

// isMCPClient checks if we're likely being run by an MCP client (AI assistant)
func isMCPClient() bool {
	// Check for common AI assistant environment indicators
	// Claude Code sets specific environment variables
	if os.Getenv("CLAUDE_CODE") != "" {
		return true
	}

	// Cursor IDE sets specific environment
	if os.Getenv("CURSOR_IDE") != "" {
		return true
	}

	// Gemini CLI might set specific indicators
	if os.Getenv("GEMINI_CLI") != "" {
		return true
	}

	// Check if stdin/stdout are pipes (typical for MCP clients)
	stdinStat, _ := os.Stdin.Stat()
	stdoutStat, _ := os.Stdout.Stat()

	stdinIsPipe := (stdinStat.Mode() & os.ModeCharDevice) == 0
	stdoutIsPipe := (stdoutStat.Mode() & os.ModeCharDevice) == 0

	// If both stdin and stdout are pipes, likely an MCP client
	if stdinIsPipe && stdoutIsPipe {
		return true
	}

	// Check parent process name for known AI tools
	ppid := os.Getppid()
	if ppid > 0 {
		// Try to read parent process command
		cmdPath := fmt.Sprintf("/proc/%d/cmdline", ppid)
		if cmdBytes, err := os.ReadFile(cmdPath); err == nil {
			cmd := string(cmdBytes)
			// Check for known AI assistant processes
			if containsAIIndicator(cmd) {
				return true
			}
		}
	}

	return false
}

// containsAIIndicator checks if the command contains known AI assistant indicators
func containsAIIndicator(cmd string) bool {
	indicators := []string{
		"claude",
		"cursor",
		"gemini",
		"copilot",
		"codeium",
		"tabnine",
		"mcp",
	}

	for _, indicator := range indicators {
		if containsIgnoreCase(cmd, indicator) {
			return true
		}
	}

	return false
}

// containsIgnoreCase checks if a string contains another string (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains check
	// Convert both to lowercase for comparison
	lower := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		lower[i] = c
	}

	subLower := make([]byte, len(substr))
	for i := 0; i < len(substr); i++ {
		c := substr[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		subLower[i] = c
	}

	// Search for substring
	for i := 0; i <= len(lower)-len(subLower); i++ {
		match := true
		for j := 0; j < len(subLower); j++ {
			if lower[i+j] != subLower[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}

// StdioHandler handles MCP protocol over stdio with enhanced initialization
type StdioHandler struct {
	server *Server
}

// NewStdioHandler creates a new stdio handler
func NewStdioHandler() *StdioHandler {
	return &StdioHandler{
		server: NewServer(0, "development", ""),
	}
}

// Run starts the stdio handler with proper MCP protocol support
func (h *StdioHandler) Run() error {
	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	// Create JSON decoder/encoder for stdio
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	// Main message loop
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			var request MCPRequest
			if err := decoder.Decode(&request); err != nil {
				// EOF is normal termination
				if err.Error() == "EOF" {
					return nil
				}
				// Log other errors to stderr
				fmt.Fprintf(os.Stderr, "Error decoding request: %v\n", err)
				continue
			}

			// Process the request
			response := h.server.processRequest(request)

			// Send response
			if err := encoder.Encode(response); err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
				continue
			}
		}
	}
}

// DirectStdioMode is the simplified entry point for main.go
func DirectStdioMode() {
	handler := NewStdioHandler()
	if err := handler.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP stdio mode error: %v\n", err)
		os.Exit(1)
	}
}

// StartStdioWithConfig starts stdio mode with custom configuration
func StartStdioWithConfig(config *ServerConfig) error {
	// Load configuration if provided
	permissions := "development"
	if config != nil && config.Permissions != "" {
		permissions = config.Permissions
	}

	server := NewServer(0, permissions, "")
	
	// Use buffered I/O for better performance
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	// Main message loop
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Read line from stdin
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err.Error() == "EOF" {
					return nil
				}
				return fmt.Errorf("error reading stdin: %w", err)
			}

			// Parse request
			var request MCPRequest
			if err := json.Unmarshal(line, &request); err != nil {
				// Skip invalid JSON
				continue
			}

			// Process request
			response := server.processRequest(request)

			// Write response
			responseBytes, err := json.Marshal(response)
			if err != nil {
				continue
			}

			writer.Write(responseBytes)
			writer.WriteByte('\n')
			writer.Flush()
		}
	}
}

// ServerConfig represents MCP server configuration
type ServerConfig struct {
	Permissions string `json:"permissions"`
	LogLevel    string `json:"log_level"`
	Timeout     string `json:"timeout"`
}