package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// cachedClaudePath caches the claude path to avoid inconsistency
var cachedClaudePath string

// getClaudePath finds the Claude Code executable path
func getClaudePath() (string, error) {
	// Return cached path if available
	if cachedClaudePath != "" {
		if info, err := os.Stat(cachedClaudePath); err == nil && !info.IsDir() {
			return cachedClaudePath, nil
		}
		// Cache is stale, clear it
		cachedClaudePath = ""
	}

	// Try to find claude in PATH first
	if path, err := exec.LookPath("claude"); err == nil {
		// Verify the file actually exists and is executable
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			cachedClaudePath = path
			return path, nil
		}
	}

	// Common installation paths with environment expansion
	commonPaths := []string{
		"/usr/local/bin/claude",
		"/usr/bin/claude",
		"/opt/homebrew/bin/claude",
		os.ExpandEnv("$HOME/.local/bin/claude"),
		os.ExpandEnv("$HOME/.nvm/versions/node/*/bin/claude"),
	}

	for _, pathPattern := range commonPaths {
		if matches, err := filepath.Glob(pathPattern); err == nil {
			for _, path := range matches {
				if info, err := os.Stat(path); err == nil && !info.IsDir() {
					cachedClaudePath = path
					return path, nil
				}
			}
		}
	}

	// Try specific known path based on current environment
	if home := os.Getenv("HOME"); home != "" {
		// Look for any node version in .nvm
		nvmPattern := filepath.Join(home, ".nvm/versions/node/*/bin/claude")
		if matches, err := filepath.Glob(nvmPattern); err == nil && len(matches) > 0 {
			// Use the first match (most recent)
			for _, path := range matches {
				if info, err := os.Stat(path); err == nil && !info.IsDir() {
					cachedClaudePath = path
					return path, nil
				}
			}
		}

		// Fallback to specific version
		nvmPath := filepath.Join(home, ".nvm/versions/node/v22.17.1/bin/claude")
		if info, err := os.Stat(nvmPath); err == nil && !info.IsDir() {
			cachedClaudePath = nvmPath
			return nvmPath, nil
		}
	}

	return "", fmt.Errorf("claude not found in PATH or common locations")
}

// isClaudeCodeInstalled checks if Claude Code is installed
func isClaudeCodeInstalled() bool {
	claudePath, err := getClaudePath()
	return err == nil && claudePath != ""
}

// isMCPAlreadyConfigured checks if Portunix MCP is already configured
func isMCPAlreadyConfigured() bool {
	claudePath, err := getClaudePath()
	if err != nil {
		return false
	}

	cmd := exec.Command(claudePath, "mcp", "list")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Simple check - look for "portunix" in the output
	return strings.Contains(string(output), "portunix")
}

// removeMCPServerFromClaudeCode removes Portunix MCP server from Claude Code
func removeMCPServerFromClaudeCode() error {
	// Find claude executable
	claudePath, err := getClaudePath()
	if err != nil {
		return fmt.Errorf("claude executable not found: %w", err)
	}

	// Use claude mcp remove command
	cmd := exec.Command(claudePath, "mcp", "remove", "portunix")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Handle Claude Code module issues
		if strings.Contains(string(output), "Cannot find module") {
			return nil // Don't fail, just silently skip for module issues
		}
		return fmt.Errorf("claude mcp remove failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

// getCurrentExecutablePath returns the path to the current executable
func getCurrentExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(execPath)
}

// getMCPPidFile returns the path to the MCP PID file
func getMCPPidFile() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.TempDir()
	}
	return filepath.Join(home, ".portunix", "mcp-server.pid")
}

// isClaudeDesktopInstalled checks if Claude Desktop is installed
func isClaudeDesktopInstalled() bool {
	// Check common Claude Desktop installation paths
	paths := []string{
		"/Applications/Claude.app",
		os.ExpandEnv("$HOME/Applications/Claude.app"),
		os.ExpandEnv("$LOCALAPPDATA/Programs/Claude"),
		os.ExpandEnv("$HOME/.local/share/applications/claude.desktop"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// isGeminiCLIInstalled checks if Gemini CLI is installed
func isGeminiCLIInstalled() bool {
	_, err := exec.LookPath("gemini")
	return err == nil
}

// isPortAvailable checks if a port is available for binding
func isPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

// findAvailablePorts finds available ports starting from 3001
func findAvailablePorts(count int) []int {
	var availablePorts []int
	startPort := 3001

	for port := startPort; port < startPort+1000 && len(availablePorts) < count; port++ {
		if isPortAvailable(port) {
			availablePorts = append(availablePorts, port)
		}
	}

	return availablePorts
}

// MCPConfiguration represents MCP server configuration
type MCPConfiguration struct {
	ServerType      string            `json:"server_type"`
	Port            int               `json:"port,omitempty"`
	Protocol        string            `json:"protocol,omitempty"`
	SecurityProfile string            `json:"security_profile"`
	Assistants      []AssistantConfig `json:"assistants"`
}

// AssistantConfig represents configuration for an AI assistant
type AssistantConfig struct {
	Name       string `json:"name"`
	ServerType string `json:"server_type"`
	Configured bool   `json:"configured"`
}

// getMCPConfigFile returns path to MCP configuration file
func getMCPConfigFile() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.TempDir()
	}
	configDir := filepath.Join(home, ".portunix")
	os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "mcp-server.json")
}

// isMCPConfigurationExists checks if MCP configuration file exists
func isMCPConfigurationExists() bool {
	_, err := os.Stat(getMCPConfigFile())
	return err == nil
}

// loadMCPConfiguration loads MCP configuration from file
func loadMCPConfiguration() (*MCPConfiguration, error) {
	configFile := getMCPConfigFile()

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config MCPConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults if empty
	if config.ServerType == "" {
		config.ServerType = "stdio"
	}
	if config.SecurityProfile == "" {
		config.SecurityProfile = "development"
	}
	if config.Port == 0 && config.ServerType == "remote" {
		config.Port = 3001
	}

	return &config, nil
}

// saveMCPConfiguration saves MCP configuration to file
func saveMCPConfiguration(config *MCPConfiguration) error {
	configFile := getMCPConfigFile()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

// removeMCPConfiguration removes MCP configuration file
func removeMCPConfiguration() error {
	configFile := getMCPConfigFile()
	return os.Remove(configFile)
}

// getAssistantDisplayName returns display name for assistant
func getAssistantDisplayName(assistant string) string {
	switch assistant {
	case "claude-code":
		return "Claude Code (CLI)"
	case "claude-desktop":
		return "Claude Desktop"
	case "gemini-cli":
		return "Gemini CLI"
	default:
		return assistant
	}
}

// isAssistantInstalled checks if assistant is installed
func isAssistantInstalled(assistant string) bool {
	switch assistant {
	case "claude-code":
		return isClaudeCodeInstalled()
	case "claude-desktop":
		return isClaudeDesktopInstalled()
	case "gemini-cli":
		return isGeminiCLIInstalled()
	default:
		return false
	}
}

// detectInstalledAssistants returns list of installed AI assistants
func detectInstalledAssistants() []string {
	var assistants []string

	if isClaudeCodeInstalled() {
		assistants = append(assistants, "claude-code")
	}

	if isClaudeDesktopInstalled() {
		assistants = append(assistants, "claude-desktop")
	}

	if isGeminiCLIInstalled() {
		assistants = append(assistants, "gemini-cli")
	}

	return assistants
}

// getDefaultServerType returns default server type for assistant
func getDefaultServerType(assistant string) string {
	switch assistant {
	case "claude-code":
		return "stdio"
	case "claude-desktop":
		return "remote"
	case "gemini-cli":
		return "stdio"
	default:
		return "stdio"
	}
}

// getDefaultSecurityProfile returns default security profile for assistant
func getDefaultSecurityProfile(assistant string) string {
	switch assistant {
	case "claude-code", "gemini-cli":
		return "development"
	case "claude-desktop":
		return "standard"
	default:
		return "development"
	}
}

// getClaudeDesktopConfigPath returns Claude Desktop configuration path
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

// isServerRunning checks if MCP server is running
func isServerRunning() bool {
	pidFile := getMCPPidFile()
	if _, err := os.Stat(pidFile); err != nil {
		return false
	}

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	pid := 0
	fmt.Sscanf(string(data), "%d", &pid)
	if pid == 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, sending signal 0 checks if process exists
	if runtime.GOOS != "windows" {
		err = process.Signal(os.Signal(nil))
		return err == nil
	}

	return true
}

// savePID saves process ID to file
func savePID(pidFile string, pid int) error {
	return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}

// formatDuration formats duration for display
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Suppress unused import warnings - these are used conditionally
var (
	_ = bufio.NewReader
	_ = syscall.SIGTERM
)

// getPortunixExecutablePath finds the main portunix binary path
func getPortunixExecutablePath() (string, error) {
	// ptx-mcp is in the same directory as portunix
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	execDir := filepath.Dir(execPath)

	// Determine binary name based on OS
	portunixName := "portunix"
	if runtime.GOOS == "windows" {
		portunixName = "portunix.exe"
	}

	portunixPath := filepath.Join(execDir, portunixName)
	if _, err := os.Stat(portunixPath); err != nil {
		// Try to find in PATH
		if path, err := exec.LookPath("portunix"); err == nil {
			return path, nil
		}
		return "", fmt.Errorf("portunix not found at %s or in PATH", portunixPath)
	}

	return portunixPath, nil
}
