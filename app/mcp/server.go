package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// Server represents the MCP server instance
type Server struct {
	Port        int
	Permissions string
	Config      string
	upgrader    websocket.Upgrader
	handlers    map[string]MethodHandler
}

// MethodHandler defines the interface for MCP method handlers
type MethodHandler func(params json.RawMessage) (interface{}, error)

// MCPRequest represents an incoming MCP request
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewServer creates a new MCP server instance
func NewServer(port int, permissions, config string) *Server {
	server := &Server{
		Port:        port,
		Permissions: permissions,
		Config:      config,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
		handlers: make(map[string]MethodHandler),
	}

	// Register default handlers
	server.registerHandlers()
	return server
}

// IsRunningFromClaudeCode detects if the process was started by Claude Code
func IsRunningFromClaudeCode() bool {
	// Check if stdin/stdout are pipes (typical for Claude Code)
	if stat, err := os.Stdin.Stat(); err == nil {
		isPipe := (stat.Mode() & os.ModeCharDevice) == 0
		if isPipe {
			fmt.Fprintf(os.Stderr, "DEBUG: Detected pipe mode - likely Claude Code\n")
			return true
		}
	}
	
	// Check for Claude-specific environment variables
	parentCmd := os.Getenv("_")
	if strings.Contains(parentCmd, "claude") {
		fmt.Fprintf(os.Stderr, "DEBUG: Detected Claude in parent command\n")
		return true
	}
	
	// Check if we have no controlling terminal (daemon-like process)
	if _, err := os.Readlink("/proc/self/fd/0"); err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: No controlling terminal detected\n")
		return true
	}
	
	fmt.Fprintf(os.Stderr, "DEBUG: Not running from Claude Code\n")
	return false
}

// StartStdio starts the MCP server in stdio mode for Claude Code integration
func (s *Server) StartStdio() error {
	log.SetOutput(os.Stderr) // Redirect logs to stderr to keep stdout clean
	
	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Fprintf(os.Stderr, "Shutting down MCP server...\n")
		cancel()
	}()

	fmt.Fprintf(os.Stderr, "MCP Server starting in stdio mode\n")
	fmt.Fprintf(os.Stderr, "Permission level: %s\n", s.Permissions)

	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return fmt.Errorf("error reading from stdin: %w", err)
				}
				return nil // EOF
			}
			
			line := scanner.Text()
			if line == "" {
				continue
			}
			
			var request MCPRequest
			if err := json.Unmarshal([]byte(line), &request); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
				continue
			}
			
			response := s.processRequest(request)
			
			responseJSON, err := json.Marshal(response)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling response: %v\n", err)
				continue
			}
			
			fmt.Println(string(responseJSON))
		}
	}
}

// Start starts the MCP server
func (s *Server) Start(daemon bool) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down MCP server...")
		
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		cancel()
	}()

	log.Printf("MCP Server starting on port %d", s.Port)
	log.Printf("Permission level: %s", s.Permissions)
	log.Printf("WebSocket endpoint: ws://localhost:%d/mcp", s.Port)
	log.Printf("Health endpoint: http://localhost:%d/health", s.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}

	<-ctx.Done()
	log.Println("MCP Server stopped")
	return nil
}

// handleWebSocket handles WebSocket connections for MCP
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("New MCP client connected from %s", r.RemoteAddr)

	for {
		var request MCPRequest
		if err := conn.ReadJSON(&request); err != nil {
			log.Printf("Error reading JSON: %v", err)
			break
		}

		response := s.processRequest(request)
		
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("Error writing JSON: %v", err)
			break
		}
	}

	log.Printf("MCP client disconnected from %s", r.RemoteAddr)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":      "healthy",
		"port":        s.Port,
		"permissions": s.Permissions,
		"timestamp":   time.Now().UTC(),
	}
	json.NewEncoder(w).Encode(response)
}

// processRequest processes an incoming MCP request
func (s *Server) processRequest(request MCPRequest) MCPResponse {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
	}

	handler, exists := s.handlers[request.Method]
	if !exists {
		response.Error = &MCPError{
			Code:    -32601,
			Message: "Method not found",
			Data:    request.Method,
		}
		return response
	}

	result, err := handler(request.Params)
	if err != nil {
		response.Error = &MCPError{
			Code:    -32603,
			Message: "Internal error",
			Data:    err.Error(),
		}
		return response
	}

	response.Result = result
	return response
}

// registerHandlers registers all MCP method handlers
func (s *Server) registerHandlers() {
	// Standard MCP protocol methods
	s.handlers["initialize"] = s.handleInitialize
	s.handlers["ping"] = s.handlePing
	s.handlers["tools/list"] = s.handleToolsList
	s.handlers["tools/call"] = s.handleToolsCall
	
	// System information tools
	s.handlers["mcp_get_system_info"] = s.handleGetSystemInfo
	s.handlers["mcp_get_capabilities"] = s.handleGetCapabilities
	s.handlers["mcp_get_environment"] = s.handleGetEnvironment

	// Development environment management
	s.handlers["mcp_detect_project_type"] = s.handleDetectProjectType
	s.handlers["mcp_analyze_dependencies"] = s.handleAnalyzeDependencies
	s.handlers["mcp_suggest_setup"] = s.handleSuggestSetup
	s.handlers["mcp_validate_environment"] = s.handleValidateEnvironment

	// Package management
	s.handlers["mcp_list_available_packages"] = s.handleListAvailablePackages
	s.handlers["mcp_install_package"] = s.handleInstallPackage
	s.handlers["mcp_check_installed"] = s.handleCheckInstalled
	s.handlers["mcp_update_packages"] = s.handleUpdatePackages

	// Container operations
	s.handlers["mcp_list_containers"] = s.handleListContainers
	s.handlers["mcp_manage_container"] = s.handleManageContainer
	s.handlers["mcp_get_container_info"] = s.handleGetContainerInfo
	s.handlers["mcp_create_sandbox"] = s.handleCreateSandbox

	// Security and safety
	s.handlers["mcp_validate_command"] = s.handleValidateCommand
	s.handlers["mcp_get_permissions"] = s.handleGetPermissions
	s.handlers["mcp_audit_log"] = s.handleAuditLog

	// Workflow automation
	s.handlers["mcp_create_project"] = s.handleCreateProject
	s.handlers["mcp_setup_ci_cd"] = s.handleSetupCICD
	s.handlers["mcp_deploy_environment"] = s.handleDeployEnvironment
}