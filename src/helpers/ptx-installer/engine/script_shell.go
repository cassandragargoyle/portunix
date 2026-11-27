/*
 *  SPDX-License-Identifier: MIT
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

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

// powerShellPrelude makes an inline PowerShell script fail on the first
// error instead of carrying on, keeps downloads fast and independent of
// the Internet Explorer engine on Windows PowerShell 5.1
const powerShellPrelude = "$ErrorActionPreference = 'Stop'\r\n" +
	"$ProgressPreference = 'SilentlyContinue'\r\n" +
	"$PSDefaultParameterValues['Invoke-WebRequest:UseBasicParsing'] = $true\r\n"

// powerShellEpilogue passes the exit code of the last native command
// (e.g. npm) through, as -File otherwise exits 0 after a failed native call
const powerShellEpilogue = "\r\nif ($LASTEXITCODE) { exit $LASTEXITCODE }\r\n"

// scriptCommand builds the command that runs an inline install script with
// the interpreter named by shell (issue #203). The returned cleanup removes
// any temporary file and must be called after the command finished
func scriptCommand(shell, script string) (*exec.Cmd, func(), error) {
	noop := func() {}

	switch shell {
	case registry.ScriptShellDefault:
		return shellCommand(script), noop, nil
	case registry.ScriptShellCmd:
		if runtime.GOOS != "windows" {
			return nil, noop, fmt.Errorf("scriptShell %q is only available on Windows", shell)
		}
		return shellCommand(script), noop, nil
	case registry.ScriptShellSh:
		return exec.Command("sh", "-c", script), noop, nil
	case registry.ScriptShellPowerShell:
		return powerShellCommand(script)
	default:
		return nil, noop, fmt.Errorf("unsupported scriptShell %q", shell)
	}
}

// powerShellCommand writes script to a temporary .ps1 file and runs it with
// -File, so the script text is never re-quoted on a command line (issue #201)
func powerShellCommand(script string) (*exec.Cmd, func(), error) {
	tempDir, err := os.MkdirTemp("", "ptx-installer-ps-*")
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to create temp directory: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(tempDir) }

	// UTF-8 BOM: Windows PowerShell 5.1 reads a file without it as ANSI
	content := string([]byte{0xEF, 0xBB, 0xBF}) + powerShellPrelude + script + powerShellEpilogue
	scriptPath := filepath.Join(tempDir, "install.ps1")
	if err := os.WriteFile(scriptPath, []byte(content), 0o600); err != nil {
		cleanup()
		return nil, func() {}, fmt.Errorf("failed to write temp script: %w", err)
	}

	cmd := exec.Command(powerShellExecutable(),
		"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	return cmd, cleanup, nil
}

// powerShellExecutable returns Windows PowerShell on Windows (always present)
// and PowerShell 7 elsewhere
func powerShellExecutable() string {
	if runtime.GOOS == "windows" {
		return "powershell.exe"
	}
	return "pwsh"
}
