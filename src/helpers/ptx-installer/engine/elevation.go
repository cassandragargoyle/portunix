/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// elevationAction tells a caller how to perform a privileged operation based
// on the current process's privilege state. Shared by the Docker installer and
// the generic package installer so both walk the same on-demand elevation path
// (issue #189).
type elevationAction int

const (
	// elevationDirect: run the operation in-process. Either we already have
	// admin/root, or we're in dry-run and skip the privileged launch entirely.
	elevationDirect elevationAction = iota
	// elevationUAC: re-launch elevated because the current process is not
	// privileged. On Windows this means a UAC consent prompt; on Linux/macOS a
	// re-exec under sudo.
	elevationUAC
)

// decideElevation picks how to run a privileged operation. Pure function so the
// policy is unit-testable without touching the OS.
//
//   - dry-run            → direct (never prompt; preview only)
//   - already elevated   → direct (run in-process)
//   - otherwise          → UAC / sudo re-launch
func decideElevation(dryRun, isAdmin bool) elevationAction {
	if dryRun || isAdmin {
		return elevationDirect
	}
	return elevationUAC
}

// errSudoMissing is returned by reExecElevated on Linux/macOS when no sudo
// binary is on PATH, so the caller can print the exact command for the user to
// run manually instead.
var errSudoMissing = fmt.Errorf("sudo not found on PATH")

// reExecElevated re-launches the *current* executable with the given arguments
// under elevated privileges and waits for it to finish. It is the generic
// counterpart to DockerInstaller.runDockerInstaller: instead of launching a
// foreign installer EXE it re-runs `portunix install …` itself, so the elevated
// child detects IsAdmin()==true and performs the install directly (no second
// prompt, no loop — guaranteed by decideElevation in the child).
//
//   - Windows: elevate via UAC (runWithUAC). Returns errUACDeclined when the
//     user dismisses the consent dialog.
//   - Linux/macOS: re-exec under sudo, streaming stdio so the password prompt
//     and install output stay visible. Returns errSudoMissing when sudo is
//     absent.
//
// args are the arguments to pass to the re-launched executable (typically
// os.Args[1:]). The executable path is resolved via os.Executable so the same
// binary — main or helper — re-runs verbatim.
func reExecElevated(args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine current executable for elevation: %w", err)
	}

	if runtime.GOOS == "windows" {
		return runWithUAC(exe, args)
	}

	// Linux/macOS: re-exec under sudo, preserving the original argv.
	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		return errSudoMissing
	}
	cmd := exec.Command(sudoPath, append([]string{exe}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("elevated re-exec via sudo failed: %w", err)
	}
	return nil
}
