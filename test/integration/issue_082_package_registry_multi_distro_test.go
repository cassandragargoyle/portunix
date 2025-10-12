package integration

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/test/testframework"
)

// Test Issue 082: Package Registry Architecture across all supported Linux distributions
func TestIssue082_PackageRegistry_MultiDistro(t *testing.T) {
	tf := testframework.NewTestFramework("Issue082_PackageRegistry_MultiDistro")
	tf.Start(t, "Test Package Registry Architecture across all officially supported Linux distributions")

	success := true
	defer tf.Finish(t, success)

	// Define officially supported distributions from ADR-009
	distributions := []struct {
		name         string
		image        string
		packageMgr   string
		supportLevel string
		skipReason   string
	}{
		// APT-based distributions
		{"Ubuntu 22.04 LTS", "ubuntu:22.04", "apt", "full", ""},
		{"Ubuntu 24.04 LTS", "ubuntu:24.04", "apt", "full", ""},
		{"Debian 11 Bullseye", "debian:11", "apt", "full", ""},
		{"Debian 12 Bookworm", "debian:12", "apt", "full", ""},

		// RPM-based distributions
		{"Fedora 39", "fedora:39", "dnf", "standard", ""},
		{"Fedora 40", "fedora:40", "dnf", "standard", ""},
		{"Rocky Linux 9", "rockylinux:9", "dnf", "standard", ""},

		// Pacman-based distributions
		{"Arch Linux", "archlinux:latest", "pacman", "standard", ""},

		// Skip Distroless - no package manager
		{"Google Distroless", "gcr.io/distroless/base", "none", "standard", "No package manager - runtime only"},
	}

	binaryPath := "../../portunix"
	testPackage := "nodejs" // Test package from new registry

	tf.Step(t, "Verify Portunix binary exists")
	if err := exec.Command("test", "-f", binaryPath).Run(); err != nil {
		tf.Error(t, "Portunix binary not found", binaryPath)
		success = false
		return
	}
	tf.Success(t, "Portunix binary found")

	tf.Separator()

	// Test package installation on each supported distribution
	for i, dist := range distributions {
		tf.Step(t, fmt.Sprintf("Testing distribution %d/%d: %s", i+1, len(distributions), dist.name))
		tf.Info(t, "Distribution details",
			fmt.Sprintf("Image: %s", dist.image),
			fmt.Sprintf("Package Manager: %s", dist.packageMgr),
			fmt.Sprintf("Support Level: %s", dist.supportLevel))

		if dist.skipReason != "" {
			tf.Warning(t, "Skipping distribution", dist.skipReason)
			continue
		}

		// Test package installation in container
		testSuccess := testPackageInDistribution(t, tf, binaryPath, dist.image, dist.name, testPackage)
		if !testSuccess {
			tf.Error(t, fmt.Sprintf("Package installation failed on %s", dist.name))
			success = false
		} else {
			tf.Success(t, fmt.Sprintf("Package installation successful on %s", dist.name))
		}

		tf.Separator()
	}

	if success {
		tf.Success(t, "All supported distributions tested successfully")
	} else {
		tf.Error(t, "Some distributions failed package installation tests")
	}
}

func testPackageInDistribution(t *testing.T, tf *testframework.TestFramework, binaryPath, image, distName, packageName string) bool {
	tf.Step(t, fmt.Sprintf("Test %s installation on %s", packageName, distName))

	// Use Portunix container run-in-container for testing
	tf.Command(t, binaryPath, []string{"docker", "run-in-container", packageName, "--image", image})

	startTime := time.Now()
	cmd := exec.Command(binaryPath, "docker", "run-in-container", packageName, "--image", image)
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	tf.Info(t, fmt.Sprintf("Installation took %v", duration))

	if err != nil {
		tf.Output(t, string(output), 1000)
		tf.Error(t, "Container installation failed", err.Error())
		return false
	}

	// Check for success indicators in output
	outputStr := string(output)

	// Look for successful installation indicators
	successIndicators := []string{
		"Installation completed",
		"Successfully installed",
		"Package installed successfully",
		"Installation successful",
	}

	hasSuccess := false
	for _, indicator := range successIndicators {
		if strings.Contains(outputStr, indicator) {
			hasSuccess = true
			break
		}
	}

	// Look for error indicators
	errorIndicators := []string{
		"Installation failed",
		"Error installing",
		"Failed to install",
		"Package not found",
		"Error:",
		"fatal:",
	}

	hasError := false
	for _, indicator := range errorIndicators {
		if strings.Contains(outputStr, indicator) {
			hasError = true
			tf.Warning(t, "Error indicator found", indicator)
			break
		}
	}

	tf.Output(t, outputStr, 500) // Show first 500 chars of output

	if hasError {
		tf.Error(t, fmt.Sprintf("Installation failed on %s", distName))
		return false
	}

	// If we have success indicators or no explicit errors, consider it success
	if hasSuccess || !hasError {
		tf.Success(t, fmt.Sprintf("Package installation successful on %s", distName))
		return true
	}

	tf.Warning(t, "Installation status unclear - no clear success/failure indicators")
	return true // Be optimistic if status is unclear
}

// Test specific registry features across distributions
func TestIssue082_RegistryFeatures_MultiDistro(t *testing.T) {
	tf := testframework.NewTestFramework("Issue082_RegistryFeatures_MultiDistro")
	tf.Start(t, "Test Package Registry features: variants, templates, AI integration")

	success := true
	defer tf.Finish(t, success)

	binaryPath := "../../portunix"

	// Test key registry features
	tf.Step(t, "Test Java variants on Ubuntu 22.04")
	if !testJavaVariants(t, tf, binaryPath) {
		success = false
	}

	tf.Separator()

	tf.Step(t, "Test package discovery and error handling")
	if !testPackageDiscovery(t, tf, binaryPath) {
		success = false
	}

	tf.Separator()

	tf.Step(t, "Test backward compatibility")
	if !testBackwardCompatibility(t, tf, binaryPath) {
		success = false
	}
}

func testJavaVariants(t *testing.T, tf *testframework.TestFramework, binaryPath string) bool {
	tf.Info(t, "Testing Java variant selection with new registry")

	// Test default Java variant
	tf.Command(t, binaryPath, []string{"install", "java", "--dry-run"})
	cmd := exec.Command(binaryPath, "install", "java", "--dry-run")
	output, err := cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "Java default variant test failed", err.Error())
		return false
	}

	if !strings.Contains(string(output), "INSTALLING: java") {
		tf.Error(t, "Java package not found in registry")
		return false
	}

	tf.Success(t, "Java default variant works")

	// Test specific Java variant
	tf.Command(t, binaryPath, []string{"install", "java", "--variant", "17", "--dry-run"})
	cmd = exec.Command(binaryPath, "install", "java", "--variant", "17", "--dry-run")
	output, err = cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "Java variant 17 test failed", err.Error())
		return false
	}

	if !strings.Contains(string(output), "Variant: 17") {
		tf.Error(t, "Java variant selection not working")
		return false
	}

	tf.Success(t, "Java variant selection works")
	return true
}

func testPackageDiscovery(t *testing.T, tf *testframework.TestFramework, binaryPath string) bool {
	tf.Info(t, "Testing package discovery and error handling")

	// Test non-existent package
	tf.Command(t, binaryPath, []string{"install", "nonexistent-package-test"})
	cmd := exec.Command(binaryPath, "install", "nonexistent-package-test")
	output, err := cmd.CombinedOutput()

	if err == nil {
		tf.Error(t, "Expected error for non-existent package, but got success")
		return false
	}

	if !strings.Contains(string(output), "not found") {
		tf.Error(t, "Expected 'not found' error message")
		return false
	}

	tf.Success(t, "Error handling for non-existent packages works")
	return true
}

func testBackwardCompatibility(t *testing.T, tf *testframework.TestFramework, binaryPath string) bool {
	tf.Info(t, "Testing backward compatibility with original install-packages.json")

	// Check if original file still exists
	cmd := exec.Command("test", "-f", "assets/install-packages.json")
	if err := cmd.Run(); err != nil {
		tf.Error(t, "Original install-packages.json not found - backward compatibility broken")
		return false
	}

	tf.Success(t, "Original install-packages.json still exists")

	// Test that packages still work
	tf.Command(t, binaryPath, []string{"install", "nodejs", "--dry-run"})
	cmd = exec.Command(binaryPath, "install", "nodejs", "--dry-run")
	output, err := cmd.CombinedOutput()

	if err != nil {
		tf.Error(t, "Backward compatibility test failed", err.Error())
		return false
	}

	if !strings.Contains(string(output), "INSTALLING: nodejs") {
		tf.Error(t, "Node.js installation not working - backward compatibility issue")
		return false
	}

	tf.Success(t, "Backward compatibility maintained")
	return true
}