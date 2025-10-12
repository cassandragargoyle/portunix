package dispatcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"portunix.ai/portunix/src/shared"
)

// HelperConfig defines configuration for helper binaries
type HelperConfig struct {
	Commands []string // Commands that this helper handles
	Binary   string   // Binary name (e.g., "ptx-container")
	Required bool     // Whether this helper is required for operation
}

// Dispatcher manages the Git-like dispatch architecture
type Dispatcher struct {
	version     string
	execDir     string
	helpers     map[string]*HelperConfig
	binSuffix   string
	discovery   *shared.HelperDiscovery
}

// NewDispatcher creates a new dispatcher instance
func NewDispatcher(version string) *Dispatcher {
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

	d := &Dispatcher{
		version:   version,
		execDir:   execDir,
		helpers:   make(map[string]*HelperConfig),
		binSuffix: binSuffix,
		discovery: shared.NewHelperDiscovery(version),
	}

	// Register known helper binaries (Phase 2 preparation)
	d.registerHelpers()

	return d
}

// registerHelpers registers known helper binary configurations
func (d *Dispatcher) registerHelpers() {
	// Phase 2: Container system extraction
	d.helpers["ptx-container"] = &HelperConfig{
		Commands: []string{"container", "docker", "podman"},
		Binary:   "ptx-container",
		Required: false,
	}

	// Phase 2: MCP server extraction
	d.helpers["ptx-mcp"] = &HelperConfig{
		Commands: []string{"mcp"},
		Binary:   "ptx-mcp",
		Required: false,
	}

	// Phase 1: Ansible Infrastructure as Code
	d.helpers["ptx-ansible"] = &HelperConfig{
		Commands: []string{"playbook"},
		Binary:   "ptx-ansible",
		Required: false,
	}

	// Issue #073: PTX-Prompting Helper for template-based prompt generation
	d.helpers["ptx-prompting"] = &HelperConfig{
		Commands: []string{"prompt"},
		Binary:   "ptx-prompting",
		Required: false,
	}
}

// ShouldDispatch checks if a command should be dispatched to a helper binary
func (d *Dispatcher) ShouldDispatch(args []string) (string, bool) {
	if len(args) == 0 {
		return "", false
	}

	command := args[0]

	// Check if any helper handles this command
	for _, config := range d.helpers {
		for _, cmd := range config.Commands {
			if cmd == command {
				helperPath := filepath.Join(d.execDir, config.Binary+d.binSuffix)
				if d.helperExists(helperPath) {
					return helperPath, true
				}
			}
		}
	}

	return "", false
}

// helperExists checks if a helper binary exists and is executable
func (d *Dispatcher) helperExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if it's executable (on Unix systems)
	if runtime.GOOS != "windows" {
		mode := info.Mode()
		if mode&0111 == 0 {
			return false
		}
	}

	return true
}

// Dispatch executes a helper binary with the given arguments
func (d *Dispatcher) Dispatch(helperPath string, args []string) error {
	// Validate helper version compatibility
	if err := d.validateHelperVersion(helperPath); err != nil {
		return fmt.Errorf("helper version validation failed: %v", err)
	}

	// Execute helper binary
	cmd := exec.Command(helperPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// validateHelperVersion validates that helper binary version is compatible
func (d *Dispatcher) validateHelperVersion(helperPath string) error {
	// Try to get helper version
	cmd := exec.Command(helperPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		// If helper doesn't support --version, skip validation for now
		return nil
	}

	helperVersionOutput := strings.TrimSpace(string(output))

	// Extract version from output like "ptx-container version dev"
	// Split by space and take the last part
	parts := strings.Fields(helperVersionOutput)
	if len(parts) == 0 {
		return fmt.Errorf("no version output from helper")
	}

	helperVersion := parts[len(parts)-1]

	// Use shared version validation logic
	return d.discovery.ValidateHelperVersion(helperVersion)
}

// ListHelpers returns information about available helper binaries
func (d *Dispatcher) ListHelpers() map[string]bool {
	result := make(map[string]bool)

	for helperName, config := range d.helpers {
		helperPath := filepath.Join(d.execDir, config.Binary+d.binSuffix)
		result[helperName] = d.helperExists(helperPath)
	}

	return result
}

// DiscoverHelpers discovers all available helper binaries using the discovery mechanism
func (d *Dispatcher) DiscoverHelpers() ([]*shared.HelperInfo, error) {
	return d.discovery.DiscoverHelpers()
}

// GetExecutableDir returns the directory containing the main executable
func (d *Dispatcher) GetExecutableDir() string {
	return d.execDir
}

// GetVersion returns the dispatcher version
func (d *Dispatcher) GetVersion() string {
	return d.version
}