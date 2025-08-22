package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	return contains(string(output), "portunix")
}

// contains checks if slice contains string (simple string search)
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
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