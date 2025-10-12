package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
)

// Embedded scripts will be set from main package
var InstallOpenSSHScript string
var VSCodeInstallScript string
var PortunixSystemPSScript string

// ExtractInstallScript extracts the embedded PowerShell script to a temporary location
func ExtractInstallScript(tempDir string) (string, error) {
	if InstallOpenSSHScript == "" {
		return "", fmt.Errorf("PowerShell script not embedded")
	}

	scriptPath := filepath.Join(tempDir, "Install-PortableOpenSSH.ps1")

	err := os.WriteFile(scriptPath, []byte(InstallOpenSSHScript), 0644)
	if err != nil {
		return "", err
	}

	return scriptPath, nil
}

// ExtractVSCodeScripts extracts the embedded VSCode scripts to a temporary location
func ExtractVSCodeScripts(tempDir string) error {
	if VSCodeInstallScript == "" {
		return fmt.Errorf("VSCode install script not embedded")
	}

	scriptPath := filepath.Join(tempDir, "VSCodeInstall.cmd")

	// Modify the script to include mkdir c:\temp
	modifiedScript := "REM Create temp directory first\nmkdir c:\\temp\n\n" + VSCodeInstallScript

	err := os.WriteFile(scriptPath, []byte(modifiedScript), 0644)
	if err != nil {
		return err
	}

	return nil
}

// ExtractPortunixSystemScript extracts the PowerShell system detection module
func ExtractPortunixSystemScript(tempDir string) error {
	if PortunixSystemPSScript == "" {
		return fmt.Errorf("PortunixSystem PowerShell script not embedded")
	}

	scriptPath := filepath.Join(tempDir, "PortunixSystem.ps1")

	err := os.WriteFile(scriptPath, []byte(PortunixSystemPSScript), 0644)
	if err != nil {
		return err
	}

	return nil
}
