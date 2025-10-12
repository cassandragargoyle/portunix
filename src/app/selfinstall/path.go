package selfinstall

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// AddToSystemPath adds a directory to the system PATH
func AddToSystemPath(binPath string) error {
	switch runtime.GOOS {
	case "windows":
		return addToWindowsPath(binPath)
	case "darwin":
		return addToMacPath(binPath)
	default: // linux and others
		return addToLinuxPath(binPath)
	}
}

// IsInPath checks if a directory is in the system PATH
func IsInPath(binPath string) bool {
	pathEnv := os.Getenv("PATH")
	paths := filepath.SplitList(pathEnv)

	// Clean the input path
	cleanPath := filepath.Clean(binPath)

	for _, p := range paths {
		if filepath.Clean(p) == cleanPath {
			return true
		}
	}

	return false
}

// addToWindowsPath adds a directory to Windows PATH
func addToWindowsPath(binPath string) error {
	// Try to add to user PATH using PowerShell
	script := fmt.Sprintf(`
$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*%s*") {
    [Environment]::SetEnvironmentVariable("Path", $path + ";%s", "User")
    Write-Host "Path added successfully"
} else {
    Write-Host "Path already exists"
}
`, binPath, binPath)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add to PATH: %w\nOutput: %s", err, output)
	}

	fmt.Println("Note: You may need to restart your terminal for PATH changes to take effect")

	return nil
}

// addToMacPath adds a directory to macOS PATH
func addToMacPath(binPath string) error {
	return addToUnixPath(binPath)
}

// addToLinuxPath adds a directory to Linux PATH
func addToLinuxPath(binPath string) error {
	return addToUnixPath(binPath)
}

// addToUnixPath adds a directory to Unix-like system PATH
func addToUnixPath(binPath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Determine which shell config file to use
	shell := os.Getenv("SHELL")
	var configFiles []string

	if strings.Contains(shell, "zsh") {
		configFiles = []string{
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".zprofile"),
		}
	} else if strings.Contains(shell, "bash") {
		configFiles = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".bash_profile"),
			filepath.Join(home, ".profile"),
		}
	} else {
		// Default to common files
		configFiles = []string{
			filepath.Join(home, ".profile"),
			filepath.Join(home, ".bashrc"),
		}
	}

	// Find the first existing config file
	var configFile string
	for _, cf := range configFiles {
		if _, err := os.Stat(cf); err == nil {
			configFile = cf
			break
		}
	}

	// If no config file exists, create .profile
	if configFile == "" {
		configFile = filepath.Join(home, ".profile")
	}

	// Read existing content
	content, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if PATH export already exists
	exportLine := fmt.Sprintf("export PATH=\"%s:$PATH\"", binPath)
	if strings.Contains(string(content), binPath) {
		return nil // Already in PATH
	}

	// Append PATH export
	file, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add a newline if file doesn't end with one
	if len(content) > 0 && content[len(content)-1] != '\n' {
		if _, err := file.WriteString("\n"); err != nil {
			return err
		}
	}

	// Add PATH export with comment
	pathAddition := fmt.Sprintf("\n# Added by Portunix installer\n%s\n", exportLine)
	if _, err := file.WriteString(pathAddition); err != nil {
		return err
	}

	fmt.Printf("Added to %s\n", configFile)
	fmt.Println("Note: Run 'source " + configFile + "' or restart your terminal for changes to take effect")

	return nil
}

// RemoveFromSystemPath removes a directory from the system PATH
func RemoveFromSystemPath(binPath string) error {
	switch runtime.GOOS {
	case "windows":
		return removeFromWindowsPath(binPath)
	default:
		return removeFromUnixPath(binPath)
	}
}

// removeFromWindowsPath removes a directory from Windows PATH
func removeFromWindowsPath(binPath string) error {
	script := fmt.Sprintf(`
$path = [Environment]::GetEnvironmentVariable("Path", "User")
$newPath = ($path.Split(';') | Where-Object { $_ -ne "%s" }) -join ';'
[Environment]::SetEnvironmentVariable("Path", $newPath, "User")
Write-Host "Path removed successfully"
`, binPath)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove from PATH: %w\nOutput: %s", err, output)
	}

	return nil
}

// removeFromUnixPath removes a directory from Unix PATH
func removeFromUnixPath(binPath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// List of potential config files
	configFiles := []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".profile"),
	}

	for _, configFile := range configFiles {
		content, err := os.ReadFile(configFile)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		// Remove lines containing the binPath
		lines := strings.Split(string(content), "\n")
		var newLines []string
		skipNext := false

		for _, line := range lines {
			if strings.Contains(line, "Added by Portunix installer") {
				skipNext = true
				continue
			}
			if skipNext && strings.Contains(line, binPath) {
				skipNext = false
				continue
			}
			newLines = append(newLines, line)
		}

		// Write back if changes were made
		newContent := strings.Join(newLines, "\n")
		if newContent != string(content) {
			if err := os.WriteFile(configFile, []byte(newContent), 0644); err != nil {
				return err
			}
			fmt.Printf("Removed from %s\n", configFile)
		}
	}

	return nil
}
