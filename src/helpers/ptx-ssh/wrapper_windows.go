//go:build windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// runWrapperPTY on Windows falls back to a direct exec. ConPTY support is
// tracked in the "Risks" section of Issue #174 — when implemented, replace
// this with a proper pseudo-console attach so password prompts on
// Windows-native ssh clients can be fed non-interactively.
func runWrapperPTY(binary string, args []string, password SecretString) error {
	cmd := exec.Command(binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Best-effort: pass the password through SSHPASS for wrappers that honour
	// it. Otherwise the user will be prompted interactively.
	if !password.IsEmpty() {
		env := os.Environ()
		env = append(env, "SSHPASS="+password.Reveal())
		cmd.Env = env
		defer password.Zero()
	}
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return fmt.Errorf("%s exited with status %d", binary, exit.ExitCode())
		}
		return err
	}
	return nil
}
