package sandbox

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"portunix.ai/app"
)

// SandboxConfig defines the configuration for the .wsb file
type SandboxConfig struct {
	EnableNetworking bool
	EnableClipboard  bool
	EnablePrinter    bool
	EnableMicrophone bool
	EnableGPU        bool
	MappedFolders    []MappedFolder
	LogonCommand     string
	EnableSSH        bool
	SSHPort          string
	SandboxArgs      []string // Arguments to pass to portunix in sandbox
}

// MappedFolder defines a folder to be mapped into the sandbox
type MappedFolder struct {
	HostPath    string
	SandboxPath string
	ReadOnly    bool
}

const wsbTemplate = `
<Configuration>
  <Networking>{{if .SandboxConfig.EnableNetworking}}Enable{{else}}Disable{{end}}</Networking>
  <Clipboard>{{if .SandboxConfig.EnableClipboard}}Enable{{else}}Disable{{end}}</Clipboard>
  <Printer>{{if .SandboxConfig.EnablePrinter}}Enable{{else}}Disable{{end}}</Printer>
  <Microphone>{{if .SandboxConfig.EnableMicrophone}}Enable{{else}}Disable{{end}}</Microphone>
  <GPU>{{if .SandboxConfig.EnableGPU}}Enable{{else}}Disable{{end}}</GPU>
  {{if .SandboxConfig.MappedFolders}}
  <MappedFolders>
    {{range .SandboxConfig.MappedFolders}}
    <MappedFolder>
      <HostFolder>{{.HostPath}}</HostFolder>
      <SandboxFolder>{{.SandboxPath}}</SandboxFolder>
      <ReadOnly>{{.ReadOnly}}</ReadOnly>
    </MappedFolder>
    {{end}}
  </MappedFolders>
  {{end}}
  <LogonCommand>
    <Command>explorer.exe c:\portunix</Command>
    <Command>{{.GeneratedLogonCommand}}</Command>
  </LogonCommand>
</Configuration>`

// GenerateWsbFile generates a .wsb file based on the provided configuration
// and returns the path to the generated file and the batch script path.
func GenerateWsbFile(config SandboxConfig) (string, string, error) {
	// Create the temporary directory if it doesn't exist
	tempDir := ".tmp"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		err = os.MkdirAll(tempDir, 0755)
		if err != nil {
			return "", "", fmt.Errorf("failed to create temporary directory: %w", err)
		}
	}

	// Prepare the logon command based on SSH settings
	generatedLogonCommand := config.LogonCommand
	batchScriptPath := ""

	if config.EnableSSH {
		// Create batch script file
		timestamp := time.Now().Format("20060102_150405")
		batchFileName := filepath.Join(tempDir, fmt.Sprintf("setup_%s.bat", timestamp))
		logFilePath := "C:\\Users\\WDAGUtilityAccount\\Desktop\\sandbox_setup.log"

		// Extract the embedded PowerShell script
		_, err := ExtractInstallScript(tempDir)
		if err != nil {
			return "", "", fmt.Errorf("failed to extract PowerShell script: %w", err)
		}

		// Extract the PowerShell system detection module
		err = ExtractPortunixSystemScript(tempDir)
		if err != nil {
			return "", "", fmt.Errorf("failed to extract PowerShell system module: %w", err)
		}

		batchContent := []string{
			"@echo off",
			"title Portunix Sandbox Setup",
			"color 0A",
			"",
			fmt.Sprintf("echo [%%TIME%%] Starting Portunix sandbox setup... > %s", logFilePath),
			"echo [%TIME%] Starting Portunix sandbox setup...",
			"echo ================================================",
			"",
			fmt.Sprintf("echo [%%TIME%%] Installing Win32-OpenSSH using custom script... >> %s", logFilePath),
			"echo [%TIME%] Installing Win32-OpenSSH using custom script...",
			"",
			"rem Install Win32-OpenSSH using our custom PowerShell script",
			fmt.Sprintf("powershell.exe -ExecutionPolicy Bypass -File \"C:\\Portunix\\Install-PortableOpenSSH.ps1\" -OpenSSHPath \"C:\\Portunix\\openssh\\OpenSSH-Win64\" -Verbose >> %s 2>&1", logFilePath),
			"if errorlevel 1 (",
			fmt.Sprintf("    echo [%%TIME%%] ERROR: Failed to install OpenSSH with custom script >> %s", logFilePath),
			"    echo [%TIME%] ERROR: Failed to install OpenSSH with custom script",
			") else (",
			fmt.Sprintf("    echo [%%TIME%%] OpenSSH installed successfully with custom script >> %s", logFilePath),
			"    echo [%TIME%] OpenSSH installed successfully with custom script",
			")",
			"",
			fmt.Sprintf("echo [%%TIME%%] SSH setup complete >> %s", logFilePath),
			"echo [%TIME%] SSH setup complete",
			"echo ================================================",
			"",
			fmt.Sprintf("echo [%%TIME%%] Adding C:\\Portunix to system PATH... >> %s", logFilePath),
			"echo [%TIME%] Adding C:\\Portunix to system PATH...",
			"",
			"rem Add C:\\Portunix to system PATH using PowerShell",
			fmt.Sprintf("powershell.exe -NoProfile -ExecutionPolicy Bypass -Command \"$currentPath = [Environment]::GetEnvironmentVariable('PATH', 'Machine'); if ($currentPath -notlike '*C:\\Portunix*') { [Environment]::SetEnvironmentVariable('PATH', $currentPath + ';C:\\Portunix', 'Machine'); Write-Host 'PATH updated successfully'; } else { Write-Host 'C:\\Portunix already in PATH'; }\" >> %s 2>&1", logFilePath),
			"if errorlevel 1 (",
			fmt.Sprintf("    echo [%%TIME%%] ERROR: Failed to update PATH >> %s", logFilePath),
			"    echo [%TIME%] ERROR: Failed to update PATH",
			") else (",
			fmt.Sprintf("    echo [%%TIME%%] PATH update completed >> %s", logFilePath),
			"    echo [%TIME%] PATH update completed",
			")",
			"",
			fmt.Sprintf("echo [%%TIME%%] Refreshing environment variables... >> %s", logFilePath),
			"echo [%TIME%] Refreshing environment variables...",
			"",
			"rem Refresh environment variables to make PATH changes available",
			fmt.Sprintf("powershell.exe -NoProfile -ExecutionPolicy Bypass -Command \"$env:PATH = [Environment]::GetEnvironmentVariable('PATH','Machine') + ';' + [Environment]::GetEnvironmentVariable('PATH','User')\" >> %s 2>&1", logFilePath),
			fmt.Sprintf("echo [%%TIME%%] Environment variables refreshed >> %s", logFilePath),
			"echo [%TIME%] Environment variables refreshed",
			"echo ================================================",
			"",
			fmt.Sprintf("echo [%%TIME%%] Setting up PowerShell system detection module... >> %s", logFilePath),
			"echo [%TIME%] Setting up PowerShell system detection module...",
			"",
			"rem Load PortunixSystem PowerShell functions for easy system detection",
			"powershell.exe -NoProfile -ExecutionPolicy Bypass -Command \"if (Test-Path 'C:\\Portunix\\PortunixSystem.ps1') { . 'C:\\Portunix\\PortunixSystem.ps1'; Write-Host 'PortunixSystem functions loaded successfully' } else { Write-Host 'PortunixSystem.ps1 not found' }\"",
			fmt.Sprintf("echo [%%TIME%%] PowerShell module setup completed >> %s", logFilePath),
			"echo [%TIME%] PowerShell module setup completed",
			"echo ================================================",
		}

		// Add the main command if provided via SandboxArgs
		if len(config.SandboxArgs) > 0 {
			mainCommand := fmt.Sprintf("C:\\\\Portunix\\\\portunix.exe %s", strings.Join(config.SandboxArgs, " "))
			batchContent = append(batchContent, []string{
				"",
				fmt.Sprintf("echo [%%TIME%%] Executing main command: %s >> %s", mainCommand, logFilePath),
				"echo [%TIME%] Executing main command: " + mainCommand,
				fmt.Sprintf("%s >> %s 2>&1", mainCommand, logFilePath),
				"if errorlevel 1 (",
				fmt.Sprintf("    echo [%%TIME%%] ERROR: Command failed with exit code %%errorlevel%% >> %s", logFilePath),
				"    echo [%TIME%] ERROR: Command failed with exit code %errorlevel%",
				"    echo.",
				"    echo FAILED! Check log file for details.",
				") else (",
				fmt.Sprintf("    echo [%%TIME%%] Command completed successfully >> %s", logFilePath),
				"    echo [%TIME%] Command completed successfully",
				"    echo.",
				"    echo SUCCESS! Command completed.",
				")",
			}...)
		} else if generatedLogonCommand != "" {
			// Fallback to LogonCommand if SandboxArgs is empty
			batchContent = append(batchContent, []string{
				"",
				fmt.Sprintf("echo [%%TIME%%] Executing main command: %s >> %s", generatedLogonCommand, logFilePath),
				"echo [%TIME%] Executing main command: " + generatedLogonCommand,
				fmt.Sprintf("%s >> %s 2>&1", generatedLogonCommand, logFilePath),
				"if errorlevel 1 (",
				fmt.Sprintf("    echo [%%TIME%%] ERROR: Command failed with exit code %%errorlevel%% >> %s", logFilePath),
				"    echo [%TIME%] ERROR: Command failed with exit code %errorlevel%",
				"    echo.",
				"    echo FAILED! Check log file for details.",
				") else (",
				fmt.Sprintf("    echo [%%TIME%%] Command completed successfully >> %s", logFilePath),
				"    echo [%TIME%] Command completed successfully",
				"    echo.",
				"    echo SUCCESS! Command completed.",
				")",
			}...)
		}

		// Add final status
		batchContent = append(batchContent, []string{
			"",
			"echo ================================================",
			fmt.Sprintf("echo [%%TIME%%] All tasks completed >> %s", logFilePath),
			"echo [%TIME%] All tasks completed",
			fmt.Sprintf("echo Log file: %s", logFilePath),
			"echo.",
			"echo Available tools:",
			"echo   - Portunix: C:\\Portunix\\portunix.exe",
			"echo   - Notepad++: C:\\Portunix\\notepadplusplus\\notepad++.exe",
			"echo   - OpenSSH: C:\\Portunix\\openssh\\OpenSSH-Win64\\",
			"echo   - Re-run setup: C:\\Portunix\\setup.bat",
			"echo.",
			"echo Window will close in 30 seconds or press any key...",
			"timeout /t 30",
		}...)

		// Write batch file
		err = os.WriteFile(batchFileName, []byte(strings.Join(batchContent, "\r\n")), 0644)
		if err != nil {
			return "", "", fmt.Errorf("failed to create batch file: %w", err)
		}

		batchScriptPath = batchFileName
		generatedLogonCommand = "powershell -NoProfile -ExecutionPolicy Bypass -Command \"Start-Sleep -Seconds 2; Start-Process cmd -ArgumentList '/k','C:\\Portunix\\setup.bat'\""
	}

	// XML escape the generated logon command for WSB file
	var xmlEscapedLogonCommand string
	if strings.Contains(generatedLogonCommand, "&") || strings.Contains(generatedLogonCommand, "<") || strings.Contains(generatedLogonCommand, ">") {
		// Only escape XML special characters, not quotes or double ampersands
		xmlEscapedLogonCommand = strings.ReplaceAll(generatedLogonCommand, "&", "&amp;")
		xmlEscapedLogonCommand = strings.ReplaceAll(xmlEscapedLogonCommand, "<", "&lt;")
		xmlEscapedLogonCommand = strings.ReplaceAll(xmlEscapedLogonCommand, ">", "&gt;")
		// But fix double escaping of &&
		xmlEscapedLogonCommand = strings.ReplaceAll(xmlEscapedLogonCommand, "&amp;&amp;", "&&")
	} else {
		xmlEscapedLogonCommand = generatedLogonCommand
	}

	// Create a temporary struct to pass to the template, including the generated logon command
	tmplConfig := struct {
		SandboxConfig
		GeneratedLogonCommand string
	}{
		config,
		xmlEscapedLogonCommand,
	}

	// Create a unique .wsb file name
	wsbTimestamp := time.Now().Format("20060102_150405")
	wsbFileName := filepath.Join(tempDir, fmt.Sprintf("sandbox_config_%s.wsb", wsbTimestamp))

	// Create the .wsb file
	file, err := os.Create(wsbFileName)
	if err != nil {
		return "", "", fmt.Errorf("failed to create .wsb file: %w", err)
	}
	defer file.Close()

	// Parse and execute the template
	tmpl, err := template.New("wsb").Parse(wsbTemplate)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse .wsb template: %w", err)
	}

	err = tmpl.Execute(file, tmplConfig)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute .wsb template: %w", err)
	}

	return wsbFileName, batchScriptPath, nil
}

// StartSandbox starts the Windows Sandbox using the specified .wsb file.
func StartSandbox(wsbFilePath string) error {
	cmd := exec.Command("WindowsSandbox.exe", wsbFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("Starting Windows Sandbox with %s...\n", wsbFilePath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to start Windows Sandbox: %w", err)
	}

	fmt.Println("Windows Sandbox started successfully.")
	return nil
}

// WaitForSSH attempts to connect to the SSH port until successful or timeout.
func WaitForSSH(port string, timeout time.Duration) error {
	fmt.Printf("Waiting for SSH setup to complete...\n")

	startTime := time.Now()
	var lastIP string

	for time.Since(startTime) < timeout {
		// Try to get IP from ssh_info.txt file in sandbox
		ip := getSandboxIPFromInfo()
		if ip != "" && ip != lastIP {
			lastIP = ip
			fmt.Printf("Detected sandbox IP: %s\n", ip)
		}

		// Try both localhost (port forwarding) and detected IP
		targets := []string{"localhost"}
		if ip != "" && ip != "localhost" {
			targets = append(targets, ip)
		}

		for _, host := range targets {
			addr := net.JoinHostPort(host, port)
			conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
			if err == nil {
				conn.Close()
				fmt.Printf("✓ SSH is ready on %s\n", addr)
				return nil
			}
		}

		sleepDuration := 5 * time.Second
		if sleepDuration > timeout-time.Since(startTime) {
			sleepDuration = timeout - time.Since(startTime)
		}
		time.Sleep(sleepDuration) // Wait before retrying
	}

	if lastIP != "" {
		return fmt.Errorf("SSH did not become ready within %s (tried localhost and %s)", timeout, lastIP)
	}
	return fmt.Errorf("SSH did not become ready within %s (tried localhost)", timeout)
}

// getSandboxIPFromInfo tries to read IP from ssh_info.txt in the mapped folder
func getSandboxIPFromInfo() string {
	// Find the most recent portunix_* directory in .tmp
	tmpDir := ".tmp"
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return ""
	}

	var newestDir string
	var newestTime int64

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "portunix_") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().Unix() > newestTime {
				newestTime = info.ModTime().Unix()
				newestDir = entry.Name()
			}
		}
	}

	if newestDir == "" {
		return ""
	}

	// Try to read ssh_info.txt from the found directory
	sshInfoPath := filepath.Join(tmpDir, newestDir, ".tmp", "ssh_info.txt")

	// Check if file exists
	if _, err := os.Stat(sshInfoPath); os.IsNotExist(err) {
		return ""
	}

	// Read the file
	content, err := os.ReadFile(sshInfoPath)
	if err != nil {
		return ""
	}

	// Parse IP address from content
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "IP Address:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ip := strings.TrimSpace(parts[1])
				// Validate IP format
				if net.ParseIP(ip) != nil {
					return ip
				}
			}
		}
	}

	return ""
}

// EnsureNotepadPlusPlus ensures Notepad++ is available in .cache and copies it to the temp directory
func EnsureNotepadPlusPlus(tempDir string) error {
	cacheDir := ".cache"
	notepadCacheDir := filepath.Join(cacheDir, "notepadplusplus")
	notepadExecutable := filepath.Join(notepadCacheDir, "notepad++.exe")

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(notepadCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Check if Notepad++ is already cached
	if _, err := os.Stat(notepadExecutable); os.IsNotExist(err) {
		fmt.Println("Downloading Notepad++ portable...")
		if err := downloadNotepadPlusPlus(notepadCacheDir); err != nil {
			return fmt.Errorf("failed to download Notepad++: %w", err)
		}
	} else {
		fmt.Println("Notepad++ found in cache")
	}

	// Copy Notepad++ to temp directory
	notepadTempDir := filepath.Join(tempDir, "notepadplusplus")
	if err := os.MkdirAll(notepadTempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp notepad directory: %w", err)
	}

	fmt.Println("Copying Notepad++ to temp directory...")
	if err := copyDirectory(notepadCacheDir, notepadTempDir); err != nil {
		return fmt.Errorf("failed to copy Notepad++: %w", err)
	}

	return nil
}

// EnsureWin32OpenSSH ensures Win32-OpenSSH is available in .cache and copies it to the temp directory
func EnsureWin32OpenSSH(tempDir string) error {
	cacheDir := ".cache"
	sshCacheDir := filepath.Join(cacheDir, "openssh")
	sshExecutable := filepath.Join(sshCacheDir, "OpenSSH-Win64", "sshd.exe")

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(sshCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create SSH cache directory: %w", err)
	}

	// Check if Win32-OpenSSH is already cached
	if _, err := os.Stat(sshExecutable); os.IsNotExist(err) {
		fmt.Println("Downloading Win32-OpenSSH...")
		if err := downloadWin32OpenSSH(sshCacheDir); err != nil {
			return fmt.Errorf("failed to download Win32-OpenSSH: %w", err)
		}
	} else {
		fmt.Println("Win32-OpenSSH found in cache")
	}

	// Copy Win32-OpenSSH to temp directory
	sshTempDir := filepath.Join(tempDir, "openssh")
	if err := os.MkdirAll(sshTempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp SSH directory: %w", err)
	}

	fmt.Println("Copying Win32-OpenSSH to temp directory...")
	if err := copyDirectory(sshCacheDir, sshTempDir); err != nil {
		return fmt.Errorf("failed to copy Win32-OpenSSH: %w", err)
	}

	return nil
}

// downloadWin32OpenSSH downloads and extracts Win32-OpenSSH from GitHub
func downloadWin32OpenSSH(destDir string) error {
	// Win32-OpenSSH latest release URL
	url := "https://github.com/PowerShell/Win32-OpenSSH/releases/latest/download/OpenSSH-Win64.zip"

	// Download the zip file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create temporary file for download
	tempFile, err := os.CreateTemp("", "openssh_*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy download to temp file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save download: %w", err)
	}

	// Extract zip file
	if err := extractZip(tempFile.Name(), destDir); err != nil {
		return fmt.Errorf("failed to extract zip: %w", err)
	}

	return nil
}

// downloadNotepadPlusPlus downloads and extracts Notepad++ portable version
func downloadNotepadPlusPlus(destDir string) error {
	// Notepad++ portable download URL (latest version)
	url := "https://github.com/notepad-plus-plus/notepad-plus-plus/releases/download/v8.6.2/npp.8.6.2.portable.x64.zip"

	// Download the zip file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create temporary file for download
	tempFile, err := os.CreateTemp("", "notepad_*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy download to temp file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save download: %w", err)
	}

	// Extract zip file
	if err := extractZip(tempFile.Name(), destDir); err != nil {
		return fmt.Errorf("failed to extract zip: %w", err)
	}

	return nil
}

// extractZip extracts a zip file to destination directory
func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Make destination directory
	os.MkdirAll(dest, 0755)

	// Extract files
	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.FileInfo().Mode())
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// Extract file
		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// copyDirectory recursively copies a directory
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// CheckSandboxRunning checks if Windows Sandbox is already running
func CheckSandboxRunning() (bool, error) {
	// Debug mode - can be enabled for troubleshooting
	debug := false

	// Try multiple approaches to detect Windows Sandbox

	// Method 1: Check for WindowsSandboxServer.exe
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq WindowsSandboxServer.exe", "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if debug {
		fmt.Printf("Debug - Method 1 output: %q, err: %v\n", string(output), err)
	}
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && strings.Contains(strings.ToLower(line), "windowssandboxserver.exe") {
				if debug {
					fmt.Printf("Debug - Found sandbox via Method 1: %s\n", line)
				}
				return true, nil
			}
		}
	}

	// Method 2: Check for any sandbox-related processes
	cmd2 := exec.Command("tasklist")
	output2, err2 := cmd2.Output()
	if debug && err2 != nil {
		fmt.Printf("Debug - Method 2 error: %v\n", err2)
	}
	if err2 == nil {
		outputStr := strings.ToLower(string(output2))
		sandboxProcesses := []string{"windowssandboxserver.exe", "windowssandbox.exe", "sandbox.exe", "wsreset.exe"}

		for _, process := range sandboxProcesses {
			if strings.Contains(outputStr, process) {
				if debug {
					fmt.Printf("Debug - Found sandbox via Method 2: %s\n", process)
				}
				return true, nil
			}
		}
	}

	// Method 3: Check for Windows Sandbox service (this might not be the right service name)
	cmd3 := exec.Command("sc", "query", "WindowsSandbox")
	output3, err3 := cmd3.Output()
	if debug {
		fmt.Printf("Debug - Method 3 output: %q, err: %v\n", string(output3), err3)
	}
	if err3 == nil {
		outputStr := strings.ToLower(string(output3))
		if strings.Contains(outputStr, "running") {
			if debug {
				fmt.Printf("Debug - Found sandbox via Method 3 service\n")
			}
			return true, nil
		}
	}

	if debug {
		fmt.Println("Debug - No sandbox detected by any method")
	}
	return false, nil
}

// PromptToCloseSandbox asks user if they want to close existing sandbox
func PromptToCloseSandbox() (bool, error) {
	fmt.Println("\n⚠️  Windows Sandbox is already running!")
	fmt.Println("Only one sandbox instance can run at a time.")
	fmt.Println()
	fmt.Print("Close existing sandbox and continue? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// EnsurePythonEmbeddable checks if --embeddable flag is present and caches Python ZIP
func EnsurePythonEmbeddable(args []string, tempDir string) error {
	// Check if --embeddable flag is present
	var hasEmbeddable bool
	var pythonVersion string = "3.13.6" // Default version

	for i, arg := range args {
		if arg == "python" {
			// Look for --embeddable flag after "python"
			for j := i + 1; j < len(args); j++ {
				if args[j] == "--embeddable" || args[j] == "-embeddable" {
					hasEmbeddable = true
					break
				}
			}
			break
		}
	}

	if !hasEmbeddable {
		return nil // No embeddable Python needed
	}

	fmt.Println("Preparing Python embeddable for sandbox...")

	// Cache directory for Python embeddable
	cacheDir := ".cache"
	pythonCacheDir := filepath.Join(cacheDir, "python-embeddable")
	pythonZipFile := fmt.Sprintf("python-%s-embed-amd64.zip", pythonVersion)
	pythonZipPath := filepath.Join(pythonCacheDir, pythonZipFile)

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(pythonCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create Python cache directory: %w", err)
	}

	// Check if Python ZIP is already cached
	if _, err := os.Stat(pythonZipPath); os.IsNotExist(err) {
		fmt.Println("Downloading Python embeddable to cache...")
		downloadURL := fmt.Sprintf("https://www.python.org/ftp/python/%s/%s", pythonVersion, pythonZipFile)
		if err := app.DownloadFile(downloadURL, pythonZipPath); err != nil {
			return fmt.Errorf("failed to download Python embeddable: %w", err)
		}
		fmt.Printf("Python embeddable cached: %s\n", pythonZipPath)
	} else {
		fmt.Println("Python embeddable found in cache")
	}

	// Copy Python ZIP to temp directory (same structure as cache)
	pythonTempDir := filepath.Join(tempDir, "python-embeddable")
	if err := os.MkdirAll(pythonTempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp Python directory: %w", err)
	}

	fmt.Println("Copying Python embeddable to sandbox...")
	destZipPath := filepath.Join(pythonTempDir, pythonZipFile)
	if err := copyFile(pythonZipPath, destZipPath); err != nil {
		return fmt.Errorf("failed to copy Python embeddable: %w", err)
	}

	fmt.Printf("Python embeddable will be available in sandbox at: C:\\Portunix\\python-embeddable\\%s\n", pythonZipFile)

	return nil
}

// EnsureVSCodeScripts extracts VSCode scripts to temp directory
func EnsureVSCodeScripts(tempDir string) error {
	fmt.Println("Preparing VSCode scripts for sandbox...")

	if err := ExtractVSCodeScripts(tempDir); err != nil {
		return fmt.Errorf("failed to extract VSCode scripts: %w", err)
	}

	fmt.Printf("VSCode scripts extracted to: %s\n", tempDir)
	return nil
}

// UseVSCodeWSBFile generates and starts sandbox with custom VSCode WSB file
func UseVSCodeWSBFile(tempDir string) error {
	fmt.Println("Using custom VSCode WSB configuration...")

	// Get absolute path for mapping
	absDir, err := filepath.Abs(tempDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Generate VSCode sandbox configuration
	config := SandboxConfig{
		EnableNetworking: true,
		EnableClipboard:  true,
		EnablePrinter:    false,
		EnableMicrophone: false,
		EnableGPU:        false,
		MappedFolders: []MappedFolder{
			{
				HostPath:    absDir,
				SandboxPath: "C:\\Portunix",
				ReadOnly:    false,
			},
		},
		LogonCommand: "C:\\Portunix\\VSCodeInstall.cmd",
	}

	// Generate WSB file using the same system as Python
	wsbFilePath, _, err := GenerateWsbFile(config)
	if err != nil {
		return fmt.Errorf("failed to generate VSCode WSB file: %w", err)
	}

	fmt.Printf("Generated VSCode WSB file: %s\n", wsbFilePath)

	// Start the sandbox directly with VSCode WSB file
	err = StartSandbox(wsbFilePath)
	if err != nil {
		return fmt.Errorf("failed to start VSCode sandbox: %w", err)
	}

	fmt.Println("\nVSCode Sandbox started!")
	fmt.Println("- VSCode will be downloaded and installed automatically")
	fmt.Println("- Check the sandbox window for installation progress")
	fmt.Printf("- VSCode install script: C:\\Portunix\\VSCodeInstall.cmd\n")

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// CloseSandbox attempts to close all running Windows Sandbox instances
func CloseSandbox() error {
	fmt.Println("Closing existing Windows Sandbox...")

	// Use taskkill to close WindowsSandboxServer.exe
	cmd := exec.Command("taskkill", "/F", "/IM", "WindowsSandboxServer.exe")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to close sandbox: %w", err)
	}

	// Wait a moment for cleanup
	time.Sleep(2 * time.Second)

	// Verify it's closed
	running, err := CheckSandboxRunning()
	if err != nil {
		return fmt.Errorf("failed to verify sandbox closure: %w", err)
	}

	if running {
		return fmt.Errorf("sandbox is still running after close attempt")
	}

	fmt.Println("✓ Windows Sandbox closed successfully")
	return nil
}

// PreDownloadWinget downloads winget and its dependencies from GitHub to host cache for sandbox use
func PreDownloadWinget(sandboxDir string) error {
	fmt.Println("📦 Pre-downloading winget and dependencies for sandbox...")
	fmt.Println("┌─────────────────────────────────────────────────────────")

	// Use persistent cache directory in current working directory
	cacheDir := ".cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Also create cache in sandbox folder for copying
	sandboxCacheDir := filepath.Join(sandboxDir, ".cache")
	if err := os.MkdirAll(sandboxCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create sandbox cache directory: %w", err)
	}

	// Using compatible VCLibs version (14.0.33321.0) with older winget (v1.6.3482)
	fmt.Println("├─ Using VCLibs 14.0.33321.0 with compatible winget v1.6.3482")

	// Check if winget is already cached
	targetFile := filepath.Join(cacheDir, "Microsoft.DesktopAppInstaller.msixbundle")
	if _, err := os.Stat(targetFile); err == nil {
		fmt.Printf("├─ Microsoft.DesktopAppInstaller.msixbundle ✓ (cached)\n")
	} else {
		// Download older compatible winget version instead of latest
		fmt.Println("├─ Downloading compatible winget version (v1.6.3482)...")
		downloadURL := "https://github.com/microsoft/winget-cli/releases/download/v1.6.3482/Microsoft.DesktopAppInstaller_8wekyb3d8bbwe.msixbundle"

		// Download the file
		if err := downloadFileFromURL(downloadURL, targetFile); err != nil {
			return fmt.Errorf("failed to download winget: %w", err)
		}
	}

	// Copy winget and dependencies to sandbox directory
	fmt.Println("├─ Copying winget and dependencies to sandbox...")

	// Copy winget
	if err := copyFile(targetFile, filepath.Join(sandboxCacheDir, "Microsoft.DesktopAppInstaller.msixbundle")); err != nil {
		return fmt.Errorf("failed to copy winget to sandbox: %w", err)
	}

	// Copy VCLibs if exists in cache
	vcLibsFile := filepath.Join(cacheDir, "Microsoft.VCLibs.140.00.UWPDesktop_latest_x64.appx")
	if _, err := os.Stat(vcLibsFile); err == nil {
		if err := copyFile(vcLibsFile, filepath.Join(sandboxCacheDir, "Microsoft.VCLibs.140.00.UWPDesktop_latest_x64.appx")); err != nil {
			fmt.Printf("├─ Warning: Failed to copy VCLibs: %v\n", err)
		} else {
			fmt.Printf("├─ VCLibs copied to sandbox\n")
		}
	} else {
		fmt.Printf("├─ VCLibs not found in cache\n")
	}

	// Copy UI.Xaml if exists in cache
	xamlFile := filepath.Join(cacheDir, "Microsoft.UI.Xaml.2.8_latest_x64.appx")
	if _, err := os.Stat(xamlFile); err == nil {
		if err := copyFile(xamlFile, filepath.Join(sandboxCacheDir, "Microsoft.UI.Xaml.2.8_latest_x64.appx")); err != nil {
			fmt.Printf("├─ Warning: Failed to copy UI.Xaml: %v\n", err)
		} else {
			fmt.Printf("├─ UI.Xaml copied to sandbox\n")
		}
	} else {
		fmt.Printf("├─ UI.Xaml not found in cache\n")
	}

	fmt.Println("└─────────────────────────────────────────────────────────")
	fmt.Printf("✅ Winget and dependencies cached successfully!\n")
	return nil
}

// downloadWingetDependencies downloads required dependencies for winget
func downloadWingetDependencies(cacheDir string) error {
	dependencies := []struct {
		name string
		url  string
	}{
		{
			name: "Microsoft.VCLibs.140.00.UWPDesktop_latest_x64.appx",
			url:  "https://download.microsoft.com/download/9/3/F/93FCF1E7-E6A4-478B-96E7-D4B285925B00/vc_redist.x64.exe",
		},
		{
			name: "Microsoft.UI.Xaml.2.8_latest_x64.appx",
			url:  "https://github.com/microsoft/microsoft-ui-xaml/releases/download/v2.8.6/Microsoft.UI.Xaml.2.8.x64.appx",
		},
	}

	for _, dep := range dependencies {
		targetFile := filepath.Join(cacheDir, dep.name)

		// Skip if already exists
		if _, err := os.Stat(targetFile); err == nil {
			fmt.Printf("├─ %s ✓ (cached)\n", dep.name)
			continue
		}

		// Clean up any old versions of this specific dependency before downloading new one
		pattern := strings.Split(dep.name, "_")[0] + "*.appx" // e.g., "Microsoft.VCLibs*.appx"
		if matches, err := filepath.Glob(filepath.Join(cacheDir, pattern)); err == nil {
			for _, oldFile := range matches {
				if oldFile != targetFile { // Don't delete the target file
					os.Remove(oldFile)
					fmt.Printf("├─ Removed old version: %s\n", filepath.Base(oldFile))
				}
			}
		}

		if err := downloadFileFromURL(dep.url, targetFile); err != nil {
			return fmt.Errorf("failed to download %s: %w", dep.name, err)
		}
	}

	return nil
}

// downloadFileFromURL downloads a file from URL and saves it with progress indicator
func downloadFileFromURL(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get file size for progress calculation
	size := resp.ContentLength

	if size > 0 {
		// Download with progress indicator
		return downloadWithProgress(resp.Body, out, size, filepath)
	} else {
		// Fallback for unknown size - just show spinner
		return downloadWithSpinner(resp.Body, out, filepath)
	}
}

// downloadWithProgress shows a progress bar during download
func downloadWithProgress(src io.Reader, dst io.Writer, total int64, filename string) error {
	var downloaded int64
	buf := make([]byte, 32*1024) // 32KB buffer

	fmt.Printf("├─ %s ", filepath.Base(filename))

	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				fmt.Println()
				return writeErr
			}
			downloaded += int64(n)

			// Calculate and display progress
			progress := float64(downloaded) / float64(total) * 100
			fmt.Printf("\r├─ %s [%.1f%%] ", filepath.Base(filename), progress)
		}

		if err == io.EOF {
			fmt.Printf("\r├─ %s [100.0%%] ✓\n", filepath.Base(filename))
			return nil
		}
		if err != nil {
			fmt.Println()
			return err
		}
	}
}

// downloadWithSpinner shows a spinning indicator for unknown size downloads
func downloadWithSpinner(src io.Reader, dst io.Writer, filename string) error {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIndex := 0

	buf := make([]byte, 32*1024)

	// Start spinner in goroutine
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				fmt.Printf("\r├─ %s %s ", filepath.Base(filename), spinner[spinnerIndex])
				spinnerIndex = (spinnerIndex + 1) % len(spinner)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Download
	_, err := io.CopyBuffer(dst, src, buf)

	// Stop spinner
	done <- true

	if err != nil {
		fmt.Printf("\r├─ %s ✗\n", filepath.Base(filename))
		return err
	}

	fmt.Printf("\r├─ %s ✓\n", filepath.Base(filename))
	return nil
}

// copyWingetCacheToSandbox copies winget cache files from persistent cache to sandbox
func copyWingetCacheToSandbox(srcCacheDir, dstCacheDir string) error {
	// Copy all winget-related files (use pattern matching for flexibility)
	patterns := []string{
		"Microsoft.VCLibs*.appx",
		"Microsoft.UI.Xaml*.appx",
		"Microsoft.DesktopAppInstaller.msixbundle",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(srcCacheDir, pattern))
		if err != nil {
			continue
		}

		for _, srcFile := range matches {
			filename := filepath.Base(srcFile)
			dstFile := filepath.Join(dstCacheDir, filename)

			// Copy file
			if err := copyFile(srcFile, dstFile); err != nil {
				return fmt.Errorf("failed to copy %s: %w", filename, err)
			}
		}
	}

	return nil
}

// exportHostVCLibs exports the working VCLibs from host Windows system for sandbox use
func exportHostVCLibs(cacheDir string) error {
	fmt.Println("├─ Exporting VCLibs from host system...")

	// Use PowerShell to find and export the installed VCLibs packages
	psScript := `
$ErrorActionPreference = "Stop"
try {
    # Find all installed VCLibs packages
    $vcLibsPackages = Get-AppxPackage | Where-Object { $_.Name -like "*Microsoft.VCLibs*" -and $_.Name -like "*UWPDesktop*" }
    
    if ($vcLibsPackages.Count -eq 0) {
        Write-Host "No VCLibs UWPDesktop packages found on host system"
        exit 1
    }
    
    # Get the latest version (sort by version)
    $latestPackage = $vcLibsPackages | Sort-Object Version -Descending | Select-Object -First 1
    Write-Host "Found VCLibs package: $($latestPackage.Name) v$($latestPackage.Version)"
    
    # Export the package to cache directory
    $exportPath = "` + cacheDir + `\\" + $latestPackage.Name + "_" + $latestPackage.Version + "_host.appx"
    Export-AppxPackage -Package $latestPackage -OutputPath $exportPath
    
    if (Test-Path $exportPath) {
        Write-Host "Successfully exported VCLibs to: $exportPath"
        Write-Host "Package size: $([math]::Round((Get-Item $exportPath).Length / 1MB, 2)) MB"
        exit 0
    } else {
        Write-Host "Export failed - file not created"
        exit 1
    }
} catch {
    Write-Host "Error exporting VCLibs: $($_.Exception.Message)"
    exit 1
}`

	// Execute PowerShell script
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("PowerShell execution failed: %w, output: %s", err, string(output))
	}

	// Parse output to find exported file
	outputStr := string(output)
	if !strings.Contains(outputStr, "Successfully exported VCLibs") {
		return fmt.Errorf("VCLibs export failed: %s", outputStr)
	}

	fmt.Printf("├─ %s", outputStr)
	return nil
}
