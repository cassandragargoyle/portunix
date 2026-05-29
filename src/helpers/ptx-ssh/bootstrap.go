/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RunBootstrapKey authenticates with a password, appends the specified public
// key to the remote authorized_keys (after checking it is not already present),
// verifies by reading it back, and prints a reminder that password auth
// should now be disabled on the host.
func RunBootstrapKey(opts ClientOptions, publicKeyPath string) error {
	pubPath, err := resolvePublicKey(publicKeyPath)
	if err != nil {
		return err
	}
	pub, err := os.ReadFile(pubPath)
	if err != nil {
		return fmt.Errorf("read public key %s: %w", pubPath, err)
	}
	pubLine := strings.TrimRight(string(pub), "\n\r")
	if pubLine == "" {
		return fmt.Errorf("public key file %s is empty", pubPath)
	}

	// Idempotency: append only when missing.
	check := fmt.Sprintf(
		"mkdir -p ~/.ssh && chmod 700 ~/.ssh && "+
			"touch ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && "+
			"grep -qxF %s ~/.ssh/authorized_keys || echo %s >> ~/.ssh/authorized_keys",
		shellQuote(pubLine), shellQuote(pubLine))
	if _, stderr, code, err := RunExecCapture(opts, check); err != nil || code != 0 {
		return fmt.Errorf("append authorized_keys failed (exit %d): %s: %w", code, strings.TrimSpace(stderr), err)
	}

	// Verify
	verify := fmt.Sprintf("grep -qxF %s ~/.ssh/authorized_keys && echo OK", shellQuote(pubLine))
	stdout, stderr, code, err := RunExecCapture(opts, verify)
	if err != nil || code != 0 {
		return fmt.Errorf("verify authorized_keys failed (exit %d): %s: %w", code, strings.TrimSpace(stderr), err)
	}
	if !strings.Contains(stdout, "OK") {
		return fmt.Errorf("verify: OK marker not found in remote output")
	}

	fmt.Printf("Public key %s installed on %s@%s.\n", pubPath, opts.Auth.Username, opts.Target.Host)
	fmt.Println("Reminder: once key-based login works, disable password authentication on the host:")
	fmt.Println("  sudo sed -i 's/^#*PasswordAuthentication .*/PasswordAuthentication no/' /etc/ssh/sshd_config")
	fmt.Println("  sudo systemctl reload ssh   # or: sudo systemctl reload sshd")
	return nil
}

func resolvePublicKey(explicit string) (string, error) {
	if explicit != "" {
		return expandPath(explicit)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	for _, name := range []string{"id_ed25519.pub", "id_ecdsa.pub", "id_rsa.pub"} {
		p := filepath.Join(home, ".ssh", name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no public key found in ~/.ssh; generate one with 'ssh-keygen -t ed25519' or pass --public-key")
}

// shellQuote returns a single-quoted form safe for POSIX sh.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
