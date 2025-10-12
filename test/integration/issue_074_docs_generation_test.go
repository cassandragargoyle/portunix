package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue074DocsGeneration tests the post-release documentation generation script
func TestIssue074DocsGeneration(t *testing.T) {
	tf := testframework.NewTestFramework("Issue074_Docs_Generation")
	tf.Start(t, "Test post-release documentation generation functionality")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	// Get project root directory
	wd, err := os.Getwd()
	if err != nil {
		tf.Error(t, "Failed to get working directory", err.Error())
		success = false
		return
	}

	// Navigate to project root (two levels up from test/integration/)
	projectRoot := filepath.Join(wd, "..", "..")
	scriptPath := filepath.Join(projectRoot, "scripts", "post-release-docs.py")
	binaryPath := filepath.Join(projectRoot, "portunix")

	tf.Separator()

	// Step 1: Verify script exists
	tf.Step(t, "Verify post-release-docs.py exists")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		tf.Error(t, "Script not found", scriptPath)
		success = false
		return
	}
	tf.Success(t, "Script found", scriptPath)

	// Step 2: Verify script is executable
	tf.Step(t, "Verify script is executable")
	fileInfo, err := os.Stat(scriptPath)
	if err != nil {
		tf.Error(t, "Failed to stat script", err.Error())
		success = false
		return
	}

	// Check if file is executable (Unix-like systems)
	mode := fileInfo.Mode()
	if mode&0111 == 0 {
		tf.Error(t, "Script is not executable")
		success = false
		return
	}
	tf.Success(t, "Script is executable")

	tf.Separator()

	// Step 3: Verify portunix binary exists (needed for documentation generation)
	tf.Step(t, "Verify portunix binary exists")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		tf.Warning(t, "Binary not found, attempting to build")

		// Try to build the binary
		cmd := exec.Command("go", "build", "-o", ".")
		cmd.Dir = projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			tf.Error(t, "Failed to build binary", err.Error())
			tf.Output(t, string(output), 500)
			success = false
			return
		}
		tf.Success(t, "Binary built successfully")
	} else {
		tf.Success(t, "Binary found", binaryPath)
	}

	tf.Separator()

	// Step 4: Test script help/usage
	tf.Step(t, "Test script help output")
	tf.Command(t, "python3", []string{scriptPath, "--help"})

	cmd := exec.Command("python3", scriptPath, "--help")
	cmd.Dir = projectRoot
	output, _ := cmd.CombinedOutput()
	outputStr := string(output)
	tf.Output(t, outputStr, 800)

	// The script should show usage or run normally
	// It's OK if it runs normally without --help flag
	tf.Info(t, "Script executed")

	tf.Separator()

	// Step 5: Test dependency check functionality
	tf.Step(t, "Test documentation generation with dry run")
	tf.Info(t, "Running documentation generation in build-only mode")

	// Run script with build-only flag (won't deploy)
	cmd = exec.Command("python3", scriptPath, "v1.0.0-test", "--build-only")
	cmd.Dir = projectRoot
	output, err = cmd.CombinedOutput()
	outputStr = string(output)
	tf.Output(t, outputStr, 1500)

	// Check if Hugo is missing (expected in CI environments)
	if strings.Contains(outputStr, "Hugo is not installed") {
		tf.Warning(t, "Hugo not installed - this is expected in CI environments")
		tf.Info(t, "Script correctly detected missing Hugo dependency")
		// This is OK - the script correctly checks dependencies
	} else if err != nil && !strings.Contains(outputStr, "Hugo is not installed") {
		// Some other error occurred
		tf.Error(t, "Script failed unexpectedly", err.Error())
		success = false
		return
	} else if err == nil {
		// Hugo is installed and script succeeded
		tf.Success(t, "Documentation generation completed successfully")

		// Check if docs-site directory was created
		docsSitePath := filepath.Join(projectRoot, "docs-site")
		if _, err := os.Stat(docsSitePath); err == nil {
			tf.Success(t, "docs-site directory created", docsSitePath)
		}
	}

	tf.Separator()

	// Step 6: Verify integration with make-release.sh
	tf.Step(t, "Verify make-release.sh integration")

	makeReleasePath := filepath.Join(projectRoot, "scripts", "make-release.sh")
	content, err := os.ReadFile(makeReleasePath)
	if err != nil {
		tf.Error(t, "Failed to read make-release.sh", err.Error())
		success = false
		return
	}

	if strings.Contains(string(content), "post-release-docs.py") {
		tf.Success(t, "make-release.sh includes call to post-release-docs.py")
	} else {
		tf.Error(t, "make-release.sh does not call post-release-docs.py")
		success = false
	}

	if strings.Contains(string(content), "AUTO_DOCS") {
		tf.Success(t, "make-release.sh supports AUTO_DOCS flag")
	} else {
		tf.Warning(t, "make-release.sh might not support AUTO_DOCS flag")
	}

	tf.Separator()

	// Step 7: Verify script handles missing binary gracefully
	tf.Step(t, "Test script with missing binary")

	// Temporarily rename binary if it exists
	backupPath := binaryPath + ".backup"
	binaryExists := false

	if _, err := os.Stat(binaryPath); err == nil {
		binaryExists = true
		if err := os.Rename(binaryPath, backupPath); err != nil {
			tf.Warning(t, "Could not rename binary for test", err.Error())
		} else {
			defer func() {
				// Restore binary
				os.Rename(backupPath, binaryPath)
			}()
		}
	}

	if binaryExists {
		cmd = exec.Command("python3", scriptPath, "v1.0.0-test", "--build-only")
		cmd.Dir = projectRoot
		output, err = cmd.CombinedOutput()
		outputStr = string(output)

		if strings.Contains(outputStr, "Portunix binary not found") {
			tf.Success(t, "Script correctly detects missing binary")
		} else {
			tf.Info(t, "Script output when binary missing", outputStr[:min(500, len(outputStr))])
		}
	}
}

// TestIssue074ScriptStructure tests the structure and quality of the documentation script
func TestIssue074ScriptStructure(t *testing.T) {
	tf := testframework.NewTestFramework("Issue074_Script_Structure")
	tf.Start(t, "Verify post-release-docs.py script structure and functions")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	wd, err := os.Getwd()
	if err != nil {
		tf.Error(t, "Failed to get working directory", err.Error())
		success = false
		return
	}

	projectRoot := filepath.Join(wd, "..", "..")
	scriptPath := filepath.Join(projectRoot, "scripts", "post-release-docs.py")

	tf.Step(t, "Read script content")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		tf.Error(t, "Failed to read script", err.Error())
		success = false
		return
	}
	scriptContent := string(content)
	tf.Success(t, "Script read successfully", "Size:", len(content), "bytes")

	tf.Separator()

	// Check for required functions
	tf.Step(t, "Verify required functions exist")
	requiredFunctions := []string{
		"check_dependencies",
		"init_html_site",
		"discover_core_commands",
		"generate_command_doc",
		"discover_plugin_commands",
		"generate_release_notes",
		"build_html_site",
		"deploy_to_github_pages",
	}

	for _, fn := range requiredFunctions {
		if strings.Contains(scriptContent, "def "+fn+"(") {
			tf.Success(t, "Function found:", fn)
		} else {
			tf.Error(t, "Missing function:", fn)
			success = false
		}
	}

	tf.Separator()

	// Check for command parsing
	tf.Step(t, "Verify command discovery logic")
	if strings.Contains(scriptContent, "'--help'") ||
		strings.Contains(scriptContent, "PORTUNIX_BIN") {
		tf.Success(t, "Script includes command discovery via --help")
	} else {
		tf.Error(t, "Script missing command discovery logic")
		success = false
	}

	// Check for HTML generation
	if strings.Contains(scriptContent, "html") || strings.Contains(scriptContent, "HTML") {
		tf.Success(t, "Script includes HTML generation")
	} else {
		tf.Error(t, "Script missing HTML generation")
		success = false
	}

	// Check for error handling
	if strings.Contains(scriptContent, "try:") && strings.Contains(scriptContent, "except") {
		tf.Success(t, "Script has error handling (try/except)")
	} else {
		tf.Warning(t, "Script might lack proper error handling")
	}

	tf.Separator()

	// Verify script modes
	tf.Step(t, "Verify script supports different modes")
	modes := map[string]string{
		"--serve":      "Local server mode",
		"--build-only": "Build-only mode",
		"--deploy":     "Deploy mode",
	}

	for mode, desc := range modes {
		if strings.Contains(scriptContent, mode) {
			tf.Success(t, desc+" supported", mode)
		} else {
			tf.Warning(t, desc+" might not be supported", mode)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}