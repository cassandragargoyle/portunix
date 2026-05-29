/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// errUACDeclined is returned by runWithUAC when the user dismisses the UAC
// consent dialog. Callers should translate it into an actionable user message.
var errUACDeclined = errors.New("UAC elevation was declined by the user")

// runWithUAC launches an executable elevated via UAC and waits for it to
// finish. Implemented through PowerShell's `Start-Process -Verb RunAs -Wait
// -PassThru` because Go's stdlib has no direct ShellExecuteEx wrapper.
//
// Limitations:
//   - stdout/stderr of the elevated child are NOT streamed back to the
//     original console. Callers must rely on the child's own UI.
//   - Returns errUACDeclined when the user clicks No on the UAC prompt
//     (PowerShell surfaces this as exit code 1223 = ERROR_CANCELLED).
//   - Windows-only at runtime.
func runWithUAC(executablePath string, args []string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("UAC elevation is Windows-only (current platform: %s)", runtime.GOOS)
	}

	script := buildUACScript(executablePath, args)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		return nil
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return fmt.Errorf("failed to launch elevated process: %w", err)
	}
	if exitErr.ExitCode() == 1223 {
		return errUACDeclined
	}
	return fmt.Errorf("elevated process exited with code %d", exitErr.ExitCode())
}

// buildUACScript renders the PowerShell one-liner that invokes Start-Process
// with the RunAs verb and forwards the child's exit code. Pulled out for unit
// testing — argument quoting is the only non-trivial part.
func buildUACScript(executablePath string, args []string) string {
	quotedExe := psSingleQuote(executablePath)
	argsClause := ""
	if len(args) > 0 {
		quoted := make([]string, len(args))
		for i, a := range args {
			quoted[i] = psSingleQuote(a)
		}
		argsClause = " -ArgumentList @(" + strings.Join(quoted, ",") + ")"
	}

	// On UAC decline PowerShell raises an InvalidOperationException whose
	// message contains "canceled by the user". Map it to ERROR_CANCELLED so
	// the Go side can recognize the case via exit code.
	return fmt.Sprintf(
		`try { $p = Start-Process -FilePath %s%s -Verb RunAs -Wait -PassThru -ErrorAction Stop; exit $p.ExitCode } `+
			`catch { if ($_.Exception.Message -match 'canceled by the user|operation was canceled') { exit 1223 } `+
			`Write-Error $_.Exception.Message; exit 1 }`,
		quotedExe, argsClause)
}

// psSingleQuote wraps a string in PowerShell single quotes, escaping embedded
// single quotes by doubling them (the documented PowerShell escape rule).
func psSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
