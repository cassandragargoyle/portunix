//go:build !windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

func runWrapperPTY(binary string, args []string, password SecretString) error {
	cmd := exec.Command(binary, args...)
	cmd.Env = os.Environ()
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("pty.Start(%s): %w", binary, err)
	}
	defer ptmx.Close()

	// Propagate window resize from the controlling terminal to the child pty.
	sigwinch := make(chan os.Signal, 1)
	signal.Notify(sigwinch, syscall.SIGWINCH)
	defer signal.Stop(sigwinch)
	go func() {
		for range sigwinch {
			_ = pty.InheritSize(os.Stdin, ptmx)
		}
	}()
	_ = pty.InheritSize(os.Stdin, ptmx)

	// Put the local terminal in raw mode so we faithfully forward keystrokes.
	fd := int(os.Stdin.Fd())
	var restore func()
	if term.IsTerminal(fd) {
		old, err := term.MakeRaw(fd)
		if err == nil {
			restore = func() { _ = term.Restore(fd, old) }
			defer restore()
		}
	}

	// Forward user input to the child.
	go io.Copy(ptmx, os.Stdin)

	// Watch for the password prompt and stream output.
	if err := feedPasswordPrompt(ptmx, os.Stdout, password); err != nil {
		// Report non-EOF errors but still wait on the child so exit status
		// propagates correctly.
		fmt.Fprintln(os.Stderr, "ptx-ssh wrapper:", err)
	}

	if err := cmd.Wait(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			if status, ok := exit.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() != 0 {
					return fmt.Errorf("%s exited with status %d", binary, status.ExitStatus())
				}
			}
			return exit
		}
		return err
	}
	return nil
}
