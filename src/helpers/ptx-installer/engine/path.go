/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// AddToUserPath adds binPath to the current user's PATH environment variable.
// Windows: updates HKCU\Environment\Path via PowerShell so the change persists
// across sessions. Linux/macOS: appends a marked export block to the shell's
// rc file. No-op (returns nil) if the path is already present.
func AddToUserPath(binPath string) error {
	if binPath == "" {
		return nil
	}
	if IsInUserPath(binPath) {
		return nil
	}
	switch runtime.GOOS {
	case "windows":
		return addToWindowsUserPath(binPath)
	default:
		return addToUnixUserPath(binPath)
	}
}

// IsInUserPath reports whether binPath is already in the current process's PATH.
// This is a best-effort check (won't see changes made by other processes), but
// it's enough to avoid duplicate entries on repeated installs in one session.
func IsInUserPath(binPath string) bool {
	clean := filepath.Clean(binPath)
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if filepath.Clean(p) == clean {
			return true
		}
	}
	return false
}

func addToWindowsUserPath(binPath string) error {
	// PowerShell handles HKCU\Environment + WM_SETTINGCHANGE broadcast so new
	// shells pick the change up without a logout. We pass the path through
	// $env:PORTUNIX_PATH_ADD to avoid quoting headaches in the script literal.
	script := `
$add = $env:PORTUNIX_PATH_ADD
$existing = [Environment]::GetEnvironmentVariable("Path", "User")
if ([string]::IsNullOrEmpty($existing)) {
    $existing = ""
}
$parts = $existing.Split(';') | Where-Object { $_ -ne "" }
if ($parts -notcontains $add) {
    if ($existing.Length -gt 0 -and -not $existing.EndsWith(';')) {
        $newPath = "$existing;$add"
    } else {
        $newPath = "$existing$add"
    }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "Added to User PATH: $add"
} else {
    Write-Host "Already in User PATH: $add"
}
`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.Env = append(os.Environ(), "PORTUNIX_PATH_ADD="+binPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell SetEnvironmentVariable failed: %w\n%s", err, out)
	}
	if len(out) > 0 {
		fmt.Printf("   %s", out)
	}
	fmt.Println("   ℹ️  Restart your terminal to pick up the new PATH.")
	return nil
}

func addToUnixUserPath(binPath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	shell := os.Getenv("SHELL")
	var candidates []string
	switch {
	case strings.Contains(shell, "zsh"):
		candidates = []string{filepath.Join(home, ".zshrc"), filepath.Join(home, ".zprofile")}
	case strings.Contains(shell, "bash"):
		candidates = []string{filepath.Join(home, ".bashrc"), filepath.Join(home, ".bash_profile"), filepath.Join(home, ".profile")}
	default:
		candidates = []string{filepath.Join(home, ".profile"), filepath.Join(home, ".bashrc")}
	}

	target := ""
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			target = c
			break
		}
	}
	if target == "" {
		target = filepath.Join(home, ".profile")
	}

	content, err := os.ReadFile(target)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(content), binPath) {
		return nil
	}

	f, err := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if len(content) > 0 && content[len(content)-1] != '\n' {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}
	block := fmt.Sprintf("\n# Added by Portunix (ptx-installer)\nexport PATH=\"%s:$PATH\"\n", binPath)
	if _, err := f.WriteString(block); err != nil {
		return err
	}

	fmt.Printf("   Added to %s\n", target)
	fmt.Printf("   ℹ️  Run 'source %s' or restart your shell to pick up the new PATH.\n", target)
	return nil
}
