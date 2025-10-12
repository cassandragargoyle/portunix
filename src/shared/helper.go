package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// HelperInfo contains information about a helper binary
type HelperInfo struct {
	Name        string   `json:"name"`
	Binary      string   `json:"binary"`
	Version     string   `json:"version"`
	Commands    []string `json:"commands"`
	Description string   `json:"description"`
	Available   bool     `json:"available"`
	Path        string   `json:"path"`
}

// HelperDiscovery manages discovery and validation of helper binaries
type HelperDiscovery struct {
	executableDir string
	binSuffix     string
	mainVersion   string
}

// NewHelperDiscovery creates a new helper discovery instance
func NewHelperDiscovery(mainVersion string) *HelperDiscovery {
	execPath, err := os.Executable()
	if err != nil {
		execPath = os.Args[0]
	}

	execDir := filepath.Dir(execPath)

	// Determine binary suffix based on platform
	binSuffix := ""
	if runtime.GOOS == "windows" {
		binSuffix = ".exe"
	}

	return &HelperDiscovery{
		executableDir: execDir,
		binSuffix:     binSuffix,
		mainVersion:   mainVersion,
	}
}

// DiscoverHelpers discovers all available helper binaries in the executable directory
func (hd *HelperDiscovery) DiscoverHelpers() ([]*HelperInfo, error) {
	var helpers []*HelperInfo

	// Read directory contents
	entries, err := os.ReadDir(hd.executableDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read executable directory: %v", err)
	}

	// Look for ptx-* binaries
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, "ptx-") {
			continue
		}

		// Remove extension for comparison
		baseName := strings.TrimSuffix(name, hd.binSuffix)
		if !strings.HasPrefix(baseName, "ptx-") {
			continue
		}

		helperPath := filepath.Join(hd.executableDir, name)
		helperInfo := hd.analyzeHelper(baseName, helperPath)
		helpers = append(helpers, helperInfo)
	}

	return helpers, nil
}

// analyzeHelper analyzes a helper binary and returns its information
func (hd *HelperDiscovery) analyzeHelper(name, path string) *HelperInfo {
	info := &HelperInfo{
		Name:      name,
		Binary:    name + hd.binSuffix,
		Path:      path,
		Available: false,
	}

	// Check if file exists and is executable
	if !hd.isExecutable(path) {
		return info
	}

	info.Available = true

	// Try to get helper information
	hd.getHelperVersion(info)
	hd.getHelperCommands(info)
	hd.getHelperDescription(info)

	return info
}

// isExecutable checks if a file exists and is executable
func (hd *HelperDiscovery) isExecutable(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if it's executable (on Unix systems)
	if runtime.GOOS != "windows" {
		mode := fileInfo.Mode()
		if mode&0111 == 0 {
			return false
		}
	}

	return true
}

// getHelperVersion gets the version of a helper binary
func (hd *HelperDiscovery) getHelperVersion(info *HelperInfo) {
	cmd := exec.Command(info.Path, "--version")
	output, err := cmd.Output()
	if err != nil {
		info.Version = "unknown"
		return
	}

	version := strings.TrimSpace(string(output))
	info.Version = version
}

// getHelperCommands gets the commands supported by a helper binary
func (hd *HelperDiscovery) getHelperCommands(info *HelperInfo) {
	// Try to get commands from helper
	cmd := exec.Command(info.Path, "--list-commands")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: infer commands from helper name
		info.Commands = hd.inferCommandsFromName(info.Name)
		return
	}

	var commands []string
	if err := json.Unmarshal(output, &commands); err != nil {
		// Fallback: treat output as newline-separated list
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line = strings.TrimSpace(line); line != "" {
				commands = append(commands, line)
			}
		}
	}

	info.Commands = commands
}

// getHelperDescription gets the description of a helper binary
func (hd *HelperDiscovery) getHelperDescription(info *HelperInfo) {
	cmd := exec.Command(info.Path, "--description")
	output, err := cmd.Output()
	if err != nil {
		info.Description = hd.getDefaultDescription(info.Name)
		return
	}

	info.Description = strings.TrimSpace(string(output))
}

// inferCommandsFromName infers commands from helper binary name
func (hd *HelperDiscovery) inferCommandsFromName(name string) []string {
	switch name {
	case "ptx-container":
		return []string{"container", "docker", "podman"}
	case "ptx-mcp":
		return []string{"mcp"}
	case "ptx-prompting":
		return []string{"prompt"}
	default:
		// Remove ptx- prefix and use as single command
		cmd := strings.TrimPrefix(name, "ptx-")
		return []string{cmd}
	}
}

// getDefaultDescription returns a default description for known helper types
func (hd *HelperDiscovery) getDefaultDescription(name string) string {
	switch name {
	case "ptx-container":
		return "Unified container management (Docker/Podman)"
	case "ptx-mcp":
		return "Model Context Protocol server"
	case "ptx-prompting":
		return "Template-based prompt generation for AI assistants"
	default:
		return fmt.Sprintf("Portunix helper: %s", name)
	}
}

// ValidateHelperVersion validates that a helper version is compatible with main version
func (hd *HelperDiscovery) ValidateHelperVersion(helperVersion string) error {
	if helperVersion == "unknown" {
		// Skip validation for helpers that don't report version
		return nil
	}

	mainVer, err := ParseVersion(hd.mainVersion)
	if err != nil {
		return fmt.Errorf("invalid main version: %v", err)
	}

	helperVer, err := ParseVersion(helperVersion)
	if err != nil {
		return fmt.Errorf("invalid helper version: %v", err)
	}

	if !mainVer.IsCompatible(helperVer) {
		return fmt.Errorf("version mismatch: main=%s, helper=%s", hd.mainVersion, helperVersion)
	}

	return nil
}