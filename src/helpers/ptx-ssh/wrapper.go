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
	"strings"
)

// RunWrapper executes the specified external binary (ssh / scp / rsync /
// ansible-playbook) under a pty, feeds the resolved password at the prompt,
// and forwards stdout/stderr back to the caller. This is only invoked when
// --wrapper is set on the command line.
//
// The pty support is provided by github.com/creack/pty. On Windows this
// currently falls back to plain exec because ConPTY integration is planned
// for a follow-up (per the "Risks" section of Issue #174).
func RunWrapper(binary string, args []string, password SecretString) error {
	if binary == "" {
		return errors.New("wrapper binary is required")
	}
	resolved, err := exec.LookPath(binary)
	if err != nil {
		return fmt.Errorf("locate %s: %w", binary, err)
	}
	return runWrapperPTY(resolved, args, password)
}

// feedPasswordPrompt watches a pty stream for common password prompts and
// writes the secret once. After the secret is written the function simply
// copies all subsequent bytes through.
//
// This helper is exposed at package scope so it can be unit-tested with
// io.Pipe — the real pty plumbing lives in runWrapperPTY below.
func feedPasswordPrompt(ptyFile io.ReadWriter, userStdout io.Writer, password SecretString) error {
	defer password.Zero()
	buf := make([]byte, 4096)
	written := false
	var pending strings.Builder
	for {
		n, err := ptyFile.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			if _, werr := userStdout.Write(chunk); werr != nil {
				return werr
			}
			if !written {
				pending.Write(chunk)
				if looksLikePasswordPrompt(pending.String()) {
					if _, werr := ptyFile.Write([]byte(password.Reveal() + "\n")); werr != nil {
						return fmt.Errorf("write password: %w", werr)
					}
					written = true
					pending.Reset()
				}
				if pending.Len() > 16384 {
					// Keep only the tail to avoid unbounded growth on noisy
					// programs.
					s := pending.String()
					pending.Reset()
					pending.WriteString(s[len(s)-4096:])
				}
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			// "input/output error" is the normal termination signal for a pty.
			if strings.Contains(err.Error(), "input/output error") {
				return nil
			}
			return err
		}
	}
}

func looksLikePasswordPrompt(s string) bool {
	low := strings.ToLower(s)
	needles := []string{"password:", "password for", "passphrase", "sudo password"}
	for _, n := range needles {
		if strings.Contains(low, n) {
			return true
		}
	}
	return false
}

// Rsync invokes the system rsync binary over an SSH tunnel established through
// this helper. The ssh-invocation wiring is passed via -e so rsync delegates
// authentication to the wrapped ssh.
func Rsync(src, dst string, opts ClientOptions, extraRsyncArgs []string) error {
	rsyncBin, err := exec.LookPath("rsync")
	if err != nil {
		return fmt.Errorf("rsync binary not found: %w", err)
	}
	// Build an ssh command that carries the target port and the known_hosts
	// file. Identity handling is left to the underlying ssh for simplicity.
	sshParts := []string{"ssh", "-p", opts.Target.Port, "-o", "StrictHostKeyChecking=accept-new"}
	if opts.KnownHosts != nil && opts.KnownHosts.LocalPath != "" {
		sshParts = append(sshParts, "-o", "UserKnownHostsFile="+opts.KnownHosts.LocalPath)
	}
	if opts.Auth.KeyPath != "" {
		sshParts = append(sshParts, "-i", opts.Auth.KeyPath)
	}
	args := []string{"-e", strings.Join(sshParts, " ")}
	args = append(args, extraRsyncArgs...)
	args = append(args, src, dst)
	cmd := exec.Command(rsyncBin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
