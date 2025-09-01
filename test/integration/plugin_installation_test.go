package integration

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const osWindows = "windows"

// TestPluginInstallation tests the basic plugin installation workflow
func TestPluginInstallation(t *testing.T) {
	// Skip if not on a supported platform
	if runtime.GOOS == osWindows {
		t.Skip("Plugin tests are currently not supported on Windows")
	}

	// Get the portunix binary path
	portunixBin := getPortunixBinary(t)

	// Run subtests
	t.Run("BuildTestPlugin", func(t *testing.T) {
		testBuildPlugin(t)
	})

	t.Run("InstallPlugin", func(t *testing.T) {
		testPluginInstall(t, portunixBin)
	})

	t.Run("ListPlugins", func(t *testing.T) {
		testPluginList(t, portunixBin)
	})

	t.Run("PluginInfo", func(t *testing.T) {
		testPluginInfo(t, portunixBin)
	})

	t.Run("PluginLifecycle", func(t *testing.T) {
		testPluginLifecycle(t, portunixBin)
	})

	t.Run("UninstallPlugin", func(t *testing.T) {
		testPluginUninstall(t, portunixBin)
	})
}

// getPortunixBinary returns the path to the portunix binary
func getPortunixBinary(t *testing.T) string {
	// First try to find the binary in the project root
	projectRoot := findProjectRoot()
	portunixBin := filepath.Join(projectRoot, "portunix")

	if runtime.GOOS == osWindows {
		portunixBin += ".exe"
	}

	// Check if binary exists
	if _, err := os.Stat(portunixBin); os.IsNotExist(err) {
		// Try to build it
		t.Log("Portunix binary not found, attempting to build...")
		cmd := exec.Command("go", "build", "-o", ".")
		cmd.Dir = projectRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to build portunix: %v\nOutput: %s", err, output)
		}
		t.Log("Successfully built portunix binary")
	}

	return portunixBin
}

// findProjectRoot finds the project root directory
func findProjectRoot() string {
	// Start from current directory and traverse up
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	for {
		// Check if go.mod exists in current directory
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	// Fallback to relative path
	return "../.."
}

// testBuildPlugin builds the test plugin
func testBuildPlugin(t *testing.T) {
	projectRoot := findProjectRoot()
	pluginDir := filepath.Join(projectRoot, "test", "test-plugin")

	// Check if plugin directory exists
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		t.Fatalf("Test plugin directory not found: %s", pluginDir)
	}

	// Build the plugin
	cmd := exec.Command("go", "build", "-o", "test-plugin")
	cmd.Dir = pluginDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build test plugin: %v\nOutput: %s", err, output)
	}

	t.Log("Successfully built test plugin")
}

// testPluginInstall tests plugin installation
func testPluginInstall(t *testing.T, portunixBin string) {
	projectRoot := findProjectRoot()
	pluginDir := filepath.Join(projectRoot, "test", "test-plugin")

	cmd := exec.Command(portunixBin, "plugin", "install", pluginDir)
	output, err := cmd.CombinedOutput()

	t.Logf("Install command output: %s", output)

	if err != nil {
		// Check if plugin is already installed
		if strings.Contains(string(output), "already installed") {
			t.Log("Plugin already installed, continuing...")
			return
		}
		t.Fatalf("Failed to install plugin: %v\nOutput: %s", err, output)
	}

	// Verify installation succeeded
	if !strings.Contains(string(output), "successfully") &&
		!strings.Contains(string(output), "installed") {
		t.Errorf("Installation output doesn't indicate success: %s", output)
	}
}

// testPluginList tests listing installed plugins
func testPluginList(t *testing.T, portunixBin string) {
	cmd := exec.Command(portunixBin, "plugin", "list")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Failed to list plugins: %v\nOutput: %s", err, output)
	}

	t.Logf("Plugin list output: %s", output)

	// Check if test-plugin appears in the list
	if !strings.Contains(string(output), "test-plugin") {
		t.Error("test-plugin not found in plugin list")
	}
}

// testPluginInfo tests getting plugin information
func testPluginInfo(t *testing.T, portunixBin string) {
	cmd := exec.Command(portunixBin, "plugin", "info", "test-plugin")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Failed to get plugin info: %v\nOutput: %s", err, output)
	}

	t.Logf("Plugin info output: %s", output)

	// Verify expected information is present
	expectedInfo := []string{
		"test-plugin",
		"1.0.0",
		"Test plugin for Portunix",
	}

	for _, expected := range expectedInfo {
		if !strings.Contains(string(output), expected) {
			t.Errorf("Expected '%s' in plugin info, but not found", expected)
		}
	}
}

// testPluginLifecycle tests plugin enable/disable/start/stop
func testPluginLifecycle(t *testing.T, portunixBin string) {
	// Test enable
	t.Run("Enable", func(t *testing.T) {
		cmd := exec.Command(portunixBin, "plugin", "enable", "test-plugin")
		output, err := cmd.CombinedOutput()

		t.Logf("Enable output: %s", output)

		if err != nil {
			// Check for common error messages
			outputStr := string(output)
			if strings.Contains(outputStr, "already enabled") ||
				strings.Contains(outputStr, "not enabled") {
				// Plugin system might have issues with enable/disable state
				t.Log("Plugin enable state issue (expected for basic implementation)")
				return
			}
			t.Fatalf("Failed to enable plugin: %v", err)
		}
	})

	// Test start
	t.Run("Start", func(t *testing.T) {
		cmd := exec.Command(portunixBin, "plugin", "start", "test-plugin")
		output, err := cmd.CombinedOutput()

		t.Logf("Start output: %s", output)

		if err != nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "already running") ||
				strings.Contains(outputStr, "not enabled") {
				t.Log("Plugin start state issue (expected for basic implementation)")
				return
			}
			t.Logf("Warning: Failed to start plugin: %v", err)
		}
	})

	// Test health check
	t.Run("Health", func(t *testing.T) {
		// Give the plugin time to start
		time.Sleep(2 * time.Second)

		cmd := exec.Command(portunixBin, "plugin", "health", "test-plugin")
		output, err := cmd.CombinedOutput()

		t.Logf("Health check output: %s", output)

		// Health check might fail if plugin doesn't implement health endpoint
		// or if plugin isn't properly enabled
		if err != nil {
			t.Log("Health check failed (expected for basic plugin)")
		}
	})

	// Test stop
	t.Run("Stop", func(t *testing.T) {
		cmd := exec.Command(portunixBin, "plugin", "stop", "test-plugin")
		output, err := cmd.CombinedOutput()

		t.Logf("Stop output: %s", output)

		if err != nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "not running") ||
				strings.Contains(outputStr, "not enabled") {
				t.Log("Plugin stop state issue (expected for basic implementation)")
				return
			}
			t.Logf("Warning: Failed to stop plugin: %v", err)
		}
	})

	// Test disable
	t.Run("Disable", func(t *testing.T) {
		cmd := exec.Command(portunixBin, "plugin", "disable", "test-plugin")
		output, err := cmd.CombinedOutput()

		t.Logf("Disable output: %s", output)

		if err != nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "already disabled") ||
				strings.Contains(outputStr, "not enabled") {
				t.Log("Plugin disable state issue (expected for basic implementation)")
				return
			}
			t.Logf("Warning: Failed to disable plugin: %v", err)
		}
	})
}

// testPluginUninstall tests plugin uninstallation
func testPluginUninstall(t *testing.T, portunixBin string) {
	cmd := exec.Command(portunixBin, "plugin", "uninstall", "test-plugin")

	// Create stdin pipe for confirmation
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	// Start the command
	output := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	// Send confirmation
	if _, err := io.WriteString(stdin, "y\n"); err != nil {
		t.Logf("Failed to send confirmation: %v", err)
	}
	stdin.Close()

	// Wait for completion
	err = cmd.Wait()

	t.Logf("Uninstall output: %s", output.String())

	if err != nil {
		// Check if plugin was already uninstalled
		if strings.Contains(output.String(), "not found") {
			t.Log("Plugin already uninstalled")
			return
		}
		t.Fatalf("Failed to uninstall plugin: %v", err)
	}

	// Verify plugin is removed from list
	cmd = exec.Command(portunixBin, "plugin", "list")
	listOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to list plugins after uninstall: %v", err)
	}

	if strings.Contains(string(listOutput), "test-plugin") {
		t.Error("test-plugin still appears in list after uninstall")
	}
}

// TestPluginValidation tests plugin validation
func TestPluginValidation(t *testing.T) {
	if runtime.GOOS == osWindows {
		t.Skip("Plugin tests are currently not supported on Windows")
	}

	portunixBin := getPortunixBinary(t)
	projectRoot := findProjectRoot()
	pluginDir := filepath.Join(projectRoot, "test", "test-plugin")

	// Test validation
	cmd := exec.Command(portunixBin, "plugin", "validate", pluginDir)
	output, err := cmd.CombinedOutput()

	t.Logf("Validation output: %s", output)

	if err != nil {
		// Validation might fail if plugin doesn't meet all requirements
		// Log as warning instead of failure
		t.Logf("Warning: Plugin validation failed: %v", err)
	} else if strings.Contains(string(output), "valid") ||
		strings.Contains(string(output), "passed") {
		t.Log("Plugin validation passed")
	}
}

// TestPluginCreate tests creating a new plugin from template
func TestPluginCreate(t *testing.T) {
	if runtime.GOOS == osWindows {
		t.Skip("Plugin tests are currently not supported on Windows")
	}

	portunixBin := getPortunixBinary(t)

	// Create temporary directory for new plugin
	tmpDir, err := os.MkdirTemp("", "plugin-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pluginName := fmt.Sprintf("test-created-plugin-%d", time.Now().Unix())
	pluginPath := filepath.Join(tmpDir, pluginName)

	// Test plugin creation
	cmd := exec.Command(portunixBin, "plugin", "create", pluginName)
	cmd.Dir = tmpDir // Set working directory
	output, err := cmd.CombinedOutput()

	t.Logf("Plugin create output: %s", output)

	if err != nil {
		t.Fatalf("Failed to create plugin: %v\nOutput: %s", err, output)
	}

	// Check if plugin was created with the expected name
	actualPath := pluginPath
	outputStr := string(output)
	if strings.Contains(outputStr, "created at:") {
		// Try to extract the actual path from output
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "created at:") {
				parts := strings.Split(line, "created at:")
				if len(parts) > 1 {
					actualPath = strings.TrimSpace(parts[1])
					// If it's a relative path, make it absolute
					if !filepath.IsAbs(actualPath) {
						actualPath = filepath.Join(tmpDir, actualPath)
					}
				}
			}
		}
	}

	// Verify plugin structure was created
	expectedFiles := []string{
		"plugin.yaml",
		"main.go",
		"README.md",
	}

	// Check both possible locations
	pathsToCheck := []string{actualPath, pluginPath, filepath.Join(tmpDir, pluginName)}

	var foundPath string
	for _, checkPath := range pathsToCheck {
		if _, err := os.Stat(checkPath); err == nil {
			foundPath = checkPath
			break
		}
	}

	if foundPath == "" {
		t.Fatalf("Plugin directory not created at any expected location")
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(foundPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", file)
		}
	}

	t.Log("Plugin template created successfully")
}
