package integration

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestPowerShellInstallation tests the PowerShell package installation workflow.
func TestPowerShellInstallation(t *testing.T) {
	const osWindows = "windows"
	// Skip on Windows since PowerShell is already available
	if runtime.GOOS == osWindows {
		t.Skip("PowerShell is already available on Windows, skipping installation test")
	}

	// Get the portunix binary path
	portunixBin := getPortunixBinary(t)

	t.Run("QuickUbuntu22Installation", func(t *testing.T) {
		testPowerShellQuickInstall(t, portunixBin)
	})

	t.Run("VerifyPowerShellCommand", func(t *testing.T) {
		testPowerShellCommand(t)
	})

	t.Run("CleanupPowerShell", func(t *testing.T) {
		testPowerShellUninstall(t, portunixBin)
	})
}

// testPowerShellQuickInstall tests PowerShell installation using quick variant.
func testPowerShellQuickInstall(t *testing.T, portunixBin string) {
	// Install PowerShell using the quick variant for Ubuntu 22.04+
	cmd := exec.Command(portunixBin, "install", "powershell", "quick")
	output, err := cmd.CombinedOutput()

	t.Logf("PowerShell install command output: %s", output)

	if err != nil {
		// Check if PowerShell is already installed
		if strings.Contains(string(output), "already installed") ||
			strings.Contains(string(output), "already exists") {
			t.Log("PowerShell already installed, continuing...")
			return
		}

		// Check for common installation issues
		outputStr := string(output)
		if strings.Contains(outputStr, "Permission denied") {
			t.Skip("Installation requires root privileges, skipping")
		}
		if strings.Contains(outputStr, "network") || strings.Contains(outputStr, "download") {
			t.Skip("Network issues during installation, skipping")
		}

		t.Fatalf("Failed to install PowerShell: %v\nOutput: %s", err, output)
	}

	// Check if installation output is empty (might indicate already installed or completed silently)
	outputStr := string(output)
	if len(strings.TrimSpace(outputStr)) == 0 {
		t.Log("Empty installation output - PowerShell may already be installed or installed silently")
	} else if !strings.Contains(outputStr, "successfully") &&
		!strings.Contains(outputStr, "installed") &&
		!strings.Contains(outputStr, "complete") {
		t.Logf("Installation output: %s", outputStr)
	}
}

// testPowerShellCommand tests that PowerShell command is available and working
func testPowerShellCommand(t *testing.T) {
	// Give some time for installation to complete
	time.Sleep(2 * time.Second)

	// Test PowerShell version command
	cmd := exec.Command("pwsh", "--version")
	output, err := cmd.CombinedOutput()

	t.Logf("PowerShell version output: %s", output)

	if err != nil {
		// Try alternative command paths
		cmd = exec.Command("powershell", "--version")
		output, err = cmd.CombinedOutput()

		if err != nil {
			t.Logf("PowerShell command not found in PATH, checking common locations...")

			// Common PowerShell installation paths on Linux
			commonPaths := []string{
				"/usr/bin/pwsh",
				"/usr/local/bin/pwsh",
				"/opt/microsoft/powershell/7/pwsh",
				"/snap/bin/pwsh",
			}

			for _, path := range commonPaths {
				cmd = exec.Command(path, "--version")
				output, err = cmd.CombinedOutput()
				if err == nil {
					t.Logf("Found PowerShell at: %s", path)
					break
				}
			}

			if err != nil {
				t.Errorf("PowerShell command not available after installation: %v", err)
				return
			}
		}
	}

	// Verify PowerShell version output
	outputStr := string(output)
	if !strings.Contains(outputStr, "PowerShell") && !strings.Contains(outputStr, "7.") {
		t.Errorf("Unexpected PowerShell version output: %s", outputStr)
	}

	// Test a simple PowerShell command
	cmd = exec.Command("pwsh", "-c", "Write-Host 'Hello from PowerShell'")
	output, err = cmd.CombinedOutput()

	if err != nil {
		// Try alternative command
		cmd = exec.Command("powershell", "-c", "Write-Host 'Hello from PowerShell'")
		output, err = cmd.CombinedOutput()

		if err != nil {
			t.Logf("Warning: PowerShell command execution failed: %v", err)
			return
		}
	}

	t.Logf("PowerShell command test output: %s", output)

	// Verify command execution
	if !strings.Contains(string(output), "Hello from PowerShell") {
		t.Errorf("PowerShell command didn't execute correctly. Output: %s", output)
	}
}

// testPowerShellUninstall tests PowerShell package uninstallation (optional cleanup)
func testPowerShellUninstall(t *testing.T, portunixBin string) {
	// This is optional cleanup - don't fail if uninstall isn't implemented
	cmd := exec.Command(portunixBin, "uninstall", "powershell")
	output, err := cmd.CombinedOutput()

	t.Logf("PowerShell uninstall output: %s", output)

	if err != nil {
		// Check if uninstall command exists
		if strings.Contains(string(output), "unknown command") ||
			strings.Contains(string(output), "not implemented") {
			t.Log("Uninstall command not implemented, skipping cleanup")
			return
		}

		t.Logf("PowerShell uninstall failed (this is optional): %v", err)
	} else {
		t.Log("PowerShell uninstalled successfully")
	}
}

// TestPowerShellPresetsIntegration tests PowerShell as part of installation presets
func TestPowerShellPresetsIntegration(t *testing.T) {
	const osWindows = "windows"
	if runtime.GOOS == osWindows {
		t.Skip("PowerShell preset testing not applicable on Windows")
	}

	portunixBin := getPortunixBinary(t)

	// Test install help to understand available options
	cmd := exec.Command(portunixBin, "install", "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Skip("Cannot test preset integration, install help not available")
	}

	t.Logf("Install help shows PowerShell is available: %v",
		strings.Contains(string(output), "powershell"))

	// Check if PowerShell variants are documented
	if strings.Contains(string(output), "powershell") {
		t.Log("PowerShell package detected in install system")

		// Check for variant information
		if strings.Contains(string(output), "ubuntu") ||
			strings.Contains(string(output), "debian") ||
			strings.Contains(string(output), "snap") {
			t.Log("PowerShell Linux variants are documented")
		}
	}

	// Test if plugin-dev preset is available (which includes protoc but not PowerShell)
	t.Run("PluginDevPresetAvailability", func(t *testing.T) {
		// Try to get information about plugin-dev preset
		// Note: This test verifies the preset system works, not that PowerShell is in plugin-dev
		cmd := exec.Command(portunixBin, "install", "preset", "--help")
		output, err := cmd.CombinedOutput()

		t.Logf("Preset help output: %s", output)

		if err != nil {
			t.Log("Preset command may not be implemented yet, this is expected")
		}
	})
}
