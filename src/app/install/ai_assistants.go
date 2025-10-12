package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// AIAssistantDetector represents detection configuration for AI assistants
type AIAssistantDetector struct {
	Name             string   `json:"name"`
	DetectionCommand string   `json:"detection_command"`
	DetectionPaths   []string `json:"detection_paths"`
	VersionCommand   string   `json:"version_command"`
	ConfigPaths      []string `json:"config_paths"`
	Category         string   `json:"category"`
}

// AIAssistantStatus represents the status of an AI assistant
type AIAssistantStatus struct {
	Name        string `json:"name"`
	Installed   bool   `json:"installed"`
	Version     string `json:"version,omitempty"`
	Path        string `json:"path,omitempty"`
	ConfigFound bool   `json:"config_found"`
	ConfigPath  string `json:"config_path,omitempty"`
}

// GetAIAssistantDetectors returns detection configurations for all supported AI assistants
func GetAIAssistantDetectors() []AIAssistantDetector {
	detectors := []AIAssistantDetector{
		{
			Name:             "claude-code",
			DetectionCommand: "claude --version",
			DetectionPaths: []string{
				// Windows
				"%LOCALAPPDATA%\\npm\\claude.cmd",
				"%APPDATA%\\npm\\claude.cmd",
				"C:\\Program Files\\nodejs\\node_modules\\@anthropic-ai\\claude-code\\bin\\claude",
				// Linux/macOS
				"/usr/local/bin/claude",
				"/usr/bin/claude",
				"~/.local/bin/claude",
				"~/.npm-global/bin/claude",
			},
			VersionCommand: "claude --version",
			ConfigPaths: []string{
				"~/.config/claude/config.json",
				"~/.claude/config.json",
				"%APPDATA%\\claude\\config.json",
			},
			Category: "AI Assistant",
		},
		{
			Name:             "claude-desktop",
			DetectionCommand: "where claude-desktop",
			DetectionPaths: []string{
				// Windows
				"%LOCALAPPDATA%\\Programs\\Claude\\Claude.exe",
				"%PROGRAMFILES%\\Claude\\Claude.exe",
				// macOS
				"/Applications/Claude.app",
				"/Applications/Claude.app/Contents/MacOS/Claude",
				// Linux
				"~/.local/bin/claude-desktop",
				"/usr/local/bin/claude-desktop",
				"~/.local/share/applications/claude-desktop.AppImage",
			},
			VersionCommand: "",
			ConfigPaths: []string{
				"~/.config/Claude/config.json",
				"%APPDATA%\\Claude\\config.json",
				"~/Library/Application Support/Claude/config.json",
			},
			Category: "AI Assistant",
		},
		{
			Name:             "gemini-cli",
			DetectionCommand: "npm list -g @google/gemini-cli >/dev/null 2>&1",
			DetectionPaths: []string{
				// Windows npm global
				"%LOCALAPPDATA%\\npm\\gemini.cmd",
				"%APPDATA%\\npm\\gemini.cmd",
				"C:\\Program Files\\nodejs\\node_modules\\@google\\gemini-cli\\bin\\gemini",
				// Linux/macOS npm global
				"/usr/local/lib/node_modules/@google/gemini-cli/bin/gemini",
				"~/.npm-global/lib/node_modules/@google/gemini-cli/bin/gemini",
				// NVM paths
				"~/.nvm/versions/node/*/lib/node_modules/@google/gemini-cli/bin/gemini",
			},
			VersionCommand: "npm list -g @google/gemini-cli 2>/dev/null | grep @google/gemini-cli || echo 'Not installed'",
			ConfigPaths: []string{
				"~/.config/gemini/config.json",
				"~/.gemini/config.json",
				"%APPDATA%\\gemini\\config.json",
			},
			Category: "AI Assistant",
		},
	}

	return detectors
}

// DetectAIAssistants detects all installed AI assistants and returns their status
func DetectAIAssistants() []AIAssistantStatus {
	detectors := GetAIAssistantDetectors()
	var statuses []AIAssistantStatus

	for _, detector := range detectors {
		status := DetectSingleAIAssistant(detector)
		statuses = append(statuses, status)
	}

	return statuses
}

// DetectSingleAIAssistant detects a single AI assistant and returns its status
func DetectSingleAIAssistant(detector AIAssistantDetector) AIAssistantStatus {
	status := AIAssistantStatus{
		Name:      detector.Name,
		Installed: false,
	}

	// Try command-based detection first
	if detector.DetectionCommand != "" {
		cmd := createCommand(detector.DetectionCommand)
		if err := cmd.Run(); err == nil {
			status.Installed = true

			// Try to get version if version command is available
			if detector.VersionCommand != "" {
				versionCmd := createCommand(detector.VersionCommand)
				if output, err := versionCmd.Output(); err == nil {
					status.Version = strings.TrimSpace(string(output))
				}
			}
		}
	}

	// Try path-based detection if command failed
	if !status.Installed {
		for _, path := range detector.DetectionPaths {
			expandedPath := expandPath(path)
			if fileExists(expandedPath) {
				status.Installed = true
				status.Path = expandedPath
				break
			}
		}
	}

	// Check for configuration files
	for _, configPath := range detector.ConfigPaths {
		expandedConfigPath := expandPath(configPath)
		if fileExists(expandedConfigPath) {
			status.ConfigFound = true
			status.ConfigPath = expandedConfigPath
			break
		}
	}

	return status
}

// DetectMissingAssistants returns a list of AI assistants that are not installed
func DetectMissingAssistants(requiredAssistants []string) []string {
	allStatuses := DetectAIAssistants()
	statusMap := make(map[string]bool)

	for _, status := range allStatuses {
		statusMap[status.Name] = status.Installed
	}

	var missing []string
	for _, required := range requiredAssistants {
		if !statusMap[required] {
			missing = append(missing, required)
		}
	}

	return missing
}

// PromptInstallMissing prompts user to install missing AI assistants
func PromptInstallMissing(missing []string) error {
	if len(missing) == 0 {
		return nil
	}

	fmt.Printf("❌ Missing AI assistants: %s\n", strings.Join(missing, ", "))
	fmt.Print("Install missing assistants? [Y/n]: ")

	var response string
	fmt.Scanln(&response)

	if response == "" || strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		for _, assistant := range missing {
			fmt.Printf("Installing %s...\n", assistant)
			if err := InstallPackage(assistant, ""); err != nil {
				fmt.Printf("Failed to install %s: %v\n", assistant, err)
				continue
			}
			fmt.Printf("✅ %s installed successfully\n", assistant)
		}
	}

	return nil
}

// PrintAIAssistantStatus prints a formatted status report of all AI assistants
func PrintAIAssistantStatus() {
	statuses := DetectAIAssistants()

	fmt.Println("════════════════════════════════════════")
	fmt.Println("🤖 AI ASSISTANT STATUS")
	fmt.Println("════════════════════════════════════════")

	for _, status := range statuses {
		if status.Installed {
			fmt.Printf("✅ %s", status.Name)
			if status.Version != "" {
				fmt.Printf(" (v%s)", status.Version)
			}
			if status.Path != "" {
				fmt.Printf(" - %s", status.Path)
			}
			fmt.Println()

			if status.ConfigFound {
				fmt.Printf("   🔧 Config: %s\n", status.ConfigPath)
			}
		} else {
			fmt.Printf("❌ %s - Not installed\n", status.Name)
		}
	}

	fmt.Println("════════════════════════════════════════")
}

// GetRecommendedAIAssistants returns recommended AI assistants for the current platform
func GetRecommendedAIAssistants() []string {
	recommendations := []string{"claude-code", "gemini-cli"}

	// Add platform-specific recommendations
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		recommendations = append(recommendations, "claude-desktop")
	}

	return recommendations
}

// Helper functions

// createCommand creates a platform-appropriate command
func createCommand(commandStr string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", commandStr)
	}
	return exec.Command("sh", "-c", commandStr)
}

// expandPath expands environment variables and tilde in paths
func expandPath(path string) string {
	// Expand environment variables
	expanded := os.ExpandEnv(path)

	// Expand tilde on Unix-like systems
	if strings.HasPrefix(expanded, "~/") && runtime.GOOS != "windows" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			expanded = filepath.Join(homeDir, expanded[2:])
		}
	}

	return expanded
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}