/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// SshpassCompat fetches a credential from ptx-credential and streams it to
// either:
//   - --write-fd N: write to that FD, for anonymous pipes passed into
//     ansible-playbook --ask-pass via bash process substitution;
//   - default: write to a short-lived FIFO in $XDG_RUNTIME_DIR whose path is
//     printed on stdout; the consumer reads from that path exactly once.
//
// The credential is fetched just-in-time and never written to disk.
func SshpassCompat(credentialID, store string, writeFD int) error {
	if credentialID == "" {
		return errors.New("credential id is required")
	}
	secret, err := readFromPtxCredential(credentialID, store)
	if err != nil {
		return err
	}
	defer secret.Zero()

	if writeFD > 0 {
		f := os.NewFile(uintptr(writeFD), "password-fd")
		if f == nil {
			return fmt.Errorf("invalid write fd %d", writeFD)
		}
		defer f.Close()
		if _, err := f.Write(append([]byte(secret.Reveal()), '\n')); err != nil {
			return fmt.Errorf("write to fd %d: %w", writeFD, err)
		}
		return nil
	}

	// Default: FIFO in runtime dir. On Windows FIFOs are not first-class, so
	// we fall back to a short-lived temp file with 0600 perms (less ideal but
	// functional for tooling that expects a path).
	if runtime.GOOS == "windows" {
		return writeTempPasswordFile(secret)
	}
	return writeFIFOPassword(secret)
}

func writeTempPasswordFile(secret SecretString) error {
	dir := os.TempDir()
	f, err := os.CreateTemp(dir, "ptx-ssh-pass-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	if _, err := f.WriteString(secret.Reveal() + "\n"); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Println(name)
	return nil
}

func writeFIFOPassword(secret SecretString) error {
	base := os.Getenv("XDG_RUNTIME_DIR")
	if base == "" {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "ptx-ssh")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	fifo := filepath.Join(dir, fmt.Sprintf("pass-%d", os.Getpid()))
	_ = os.Remove(fifo)
	if err := mkFIFO(fifo, 0o600); err != nil {
		return fmt.Errorf("mkfifo: %w", err)
	}
	// Print path before opening so the consumer can begin reading; open-for-write
	// will block until the consumer opens for read.
	fmt.Println(fifo)
	f, err := os.OpenFile(fifo, os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(fifo)
	if _, err := f.WriteString(secret.Reveal() + "\n"); err != nil {
		return err
	}
	return nil
}
