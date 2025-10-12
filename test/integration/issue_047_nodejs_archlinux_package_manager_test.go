package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue047NodeJSPackageManagerDetection tests Node.js package manager detection logic
func TestIssue047NodeJSPackageManagerDetection(t *testing.T) {
	tf := testframework.NewTestFramework("Issue047_NodeJS_PackageManager_Detection")
	tf.Start(t, "Test Node.js package manager detection and configuration validation")

	success := true
	defer tf.Finish(t, success)

	// Verify Portunix binary exists and get path
	binaryPath := tf.MustVerifyPortunixBinary(t)

	tf.Separator()

	// Test 1: Basic Node.js Installation Detection (Current Platform)
	tf.Step(t, "Test Node.js package manager detection on current platform")
	
	tf.Command(t, binaryPath, []string{"install", "nodejs", "--dry-run"})
	
	cmd := exec.Command(binaryPath, "install", "nodejs", "--dry-run")
	output, err := cmd.CombinedOutput()
	
	tf.Output(t, string(output), 100)
	
	if err != nil {
		tf.Error(t, "Node.js detection command failed", err.Error())
		success = false
		return
	}

	outputStr := string(output)
	
	// Check that some package manager is detected
	if !strings.Contains(outputStr, "Installation type:") {
		tf.Error(t, "No installation type detected in output")
		success = false
		return
	}
	tf.Success(t, "Package manager detection working - installation type found")

	// Check for correct packages
	if !strings.Contains(outputStr, "nodejs") {
		tf.Error(t, "nodejs package not found in output")
		success = false
		return
	}
	tf.Success(t, "nodejs package correctly identified")

	tf.Separator()

	// Test 2: Manual Pacman Variant Override
	tf.Step(t, "Test manual pacman variant selection")
	
	tf.Command(t, binaryPath, []string{"install", "nodejs", "--variant", "pacman", "--dry-run"})
	
	pacmanCmd := exec.Command(binaryPath, "install", "nodejs", "--variant", "pacman", "--dry-run")
	pacmanOutput, pacmanErr := pacmanCmd.CombinedOutput()
	
	tf.Output(t, string(pacmanOutput), 100)
	
	if pacmanErr != nil {
		tf.Error(t, "Manual pacman variant test failed", pacmanErr.Error())
		success = false
		return
	}

	pacmanOutputStr := string(pacmanOutput)
	if !strings.Contains(pacmanOutputStr, "Installation type: pacman") {
		tf.Error(t, "Manual pacman variant override failed")
		tf.Info(t, "Expected: Installation type: pacman")
		success = false
		return
	}
	tf.Success(t, "Manual pacman variant override works correctly")

	// Verify pacman packages
	if !strings.Contains(pacmanOutputStr, "nodejs, npm") {
		tf.Error(t, "Pacman variant doesn't show correct packages")
		success = false
		return
	}
	tf.Success(t, "Pacman variant shows correct packages (nodejs, npm)")

	tf.Separator()

	// Test 3: Manual APT Variant Override 
	tf.Step(t, "Test manual apt variant selection for regression")
	
	tf.Command(t, binaryPath, []string{"install", "nodejs", "--variant", "apt", "--dry-run"})
	
	aptCmd := exec.Command(binaryPath, "install", "nodejs", "--variant", "apt", "--dry-run")
	aptOutput, aptErr := aptCmd.CombinedOutput()
	
	tf.Output(t, string(aptOutput), 100)
	
	if aptErr != nil {
		tf.Error(t, "Manual apt variant test failed", aptErr.Error())
		success = false
		return
	}

	aptOutputStr := string(aptOutput)
	if !strings.Contains(aptOutputStr, "Installation type: apt") {
		tf.Error(t, "Manual apt variant override failed")
		success = false
		return
	}
	tf.Success(t, "Manual apt variant override works correctly (no regression)")

	tf.Separator()

	// Test 4: Verify Available Variants
	tf.Step(t, "Test that both apt and pacman variants are available")
	
	// Test that pacman variant exists (should not fail)
	pacmanTestCmd := exec.Command(binaryPath, "install", "nodejs", "--variant", "pacman", "--dry-run")
	_, pacmanTestErr := pacmanTestCmd.CombinedOutput()
	
	if pacmanTestErr != nil {
		tf.Error(t, "Pacman variant not available", pacmanTestErr.Error())
		success = false
	} else {
		tf.Success(t, "Pacman variant is available")
	}
	
	// Test that apt variant exists (should not fail)
	aptTestCmd := exec.Command(binaryPath, "install", "nodejs", "--variant", "apt", "--dry-run")
	_, aptTestErr := aptTestCmd.CombinedOutput()
	
	if aptTestErr != nil {
		tf.Error(t, "APT variant not available", aptTestErr.Error())
		success = false
	} else {
		tf.Success(t, "APT variant is available")
	}

	tf.Success(t, "All Issue #047 package manager detection tests completed")
}

// TestIssue047ConfigurationValidation validates the Node.js configuration includes pacman variant
func TestIssue047ConfigurationValidation(t *testing.T) {
	tf := testframework.NewTestFramework("Issue047_Configuration_Validation")
	tf.Start(t, "Validate Node.js package configuration includes pacman variant")

	success := true
	defer tf.Finish(t, success)

	tf.Step(t, "Check install-packages.json contains pacman variant for nodejs")

	// Read the install-packages.json file
	configPath, err := filepath.Abs("../../assets/install-packages.json")
	if err != nil {
		tf.Error(t, "Failed to get config path", err.Error())
		success = false
		return
	}

	tf.Info(t, "Configuration file:", configPath)

	configData, err := os.ReadFile(configPath)
	if err != nil {
		tf.Error(t, "Failed to read configuration file", err.Error())
		success = false
		return
	}

	configContent := string(configData)

	// Check for nodejs package definition
	if !strings.Contains(configContent, `"nodejs"`) {
		tf.Error(t, "nodejs package not found in configuration")
		success = false
		return
	}
	tf.Success(t, "nodejs package found in configuration")

	// Check for pacman variant
	if !strings.Contains(configContent, `"pacman"`) {
		tf.Error(t, "pacman variant not found in nodejs configuration")
		success = false
		return
	}
	tf.Success(t, "pacman variant found in nodejs configuration")

	// Check for pacman packages
	if !strings.Contains(configContent, `"packages": ["nodejs", "npm"]`) {
		tf.Error(t, "Expected pacman packages not found")
		success = false
		return
	}
	tf.Success(t, "Correct pacman packages (nodejs, npm) found in configuration")

	// Check for post-install commands (symbolic links) - more flexible search
	if !strings.Contains(configContent, `ln -sf /usr/bin/node`) {
		tf.Error(t, "Expected post-install command for node not found")
		success = false
		return
	}
	tf.Success(t, "Post-install command for node symbolic link found")

	if !strings.Contains(configContent, `ln -sf /usr/bin/npm`) {
		tf.Error(t, "Expected post-install command for npm not found")
		success = false
		return
	}
	tf.Success(t, "Post-install command for npm symbolic link found")

	// Check for pacman type
	if !strings.Contains(configContent, `"type": "pacman"`) {
		tf.Error(t, "pacman type not found in configuration")
		success = false
		return
	}
	tf.Success(t, "Pacman type correctly specified in configuration")

	tf.Success(t, "Configuration validation completed - all pacman variant components present")
}

// TestIssue047DistributionDetectionLogic tests the distribution detection logic
func TestIssue047DistributionDetectionLogic(t *testing.T) {
	tf := testframework.NewTestFramework("Issue047_Distribution_Detection")
	tf.Start(t, "Test that distribution detection logic is implemented")

	success := true
	defer tf.Finish(t, success)

	// Verify Portunix binary exists and get path
	binaryPath := tf.MustVerifyPortunixBinary(t)

	tf.Separator()
	
	// This test verifies that the detection logic exists by testing variant selection
	tf.Step(t, "Verify variant selection logic works for different package managers")

	// Test cases for different package manager variants
	testCases := []struct {
		variant     string
		expectedPM  string
		description string
	}{
		{
			variant:     "pacman",
			expectedPM:  "pacman",
			description: "Pacman variant should use pacman package manager",
		},
		{
			variant:     "apt", 
			expectedPM:  "apt",
			description: "APT variant should use apt package manager",
		},
	}

	for i, testCase := range testCases {
		tf.Step(t, testCase.description)
		tf.Info(t, "Testing variant:", testCase.variant)
		tf.Info(t, "Expected package manager:", testCase.expectedPM)

		tf.Command(t, binaryPath, []string{"install", "nodejs", "--variant", testCase.variant, "--dry-run"})

		cmd := exec.Command(binaryPath, "install", "nodejs", "--variant", testCase.variant, "--dry-run")
		output, err := cmd.CombinedOutput()

		// Show only relevant lines to keep output clean
		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")
		var relevantLines []string
		for _, line := range lines {
			if strings.Contains(line, "Installation type:") || 
			   strings.Contains(line, "Packages:") ||
			   strings.Contains(line, "ERROR") ||
			   strings.Contains(line, "Failed") {
				relevantLines = append(relevantLines, line)
			}
		}
		
		if len(relevantLines) > 0 {
			tf.Output(t, strings.Join(relevantLines, "\n"), 200)
		}

		if err != nil {
			tf.Error(t, testCase.variant+" variant test failed", err.Error())
			success = false
			continue
		}

		// Check for expected package manager
		expectedText := "Installation type: " + testCase.expectedPM
		if strings.Contains(outputStr, expectedText) {
			tf.Success(t, testCase.variant+" variant correctly uses "+testCase.expectedPM+" package manager")
		} else {
			tf.Error(t, testCase.variant+" variant doesn't use expected package manager")
			tf.Info(t, "Expected:", expectedText)
			success = false
		}

		// Add separator between test cases (except last one)
		if i < len(testCases)-1 {
			tf.Separator()
		}
	}

	tf.Success(t, "Distribution detection logic testing completed")
}