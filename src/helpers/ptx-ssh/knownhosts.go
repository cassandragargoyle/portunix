/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// HostKeyPolicy controls behaviour when the remote key is unknown or has
// changed.
type HostKeyPolicy int

const (
	PolicyStrict    HostKeyPolicy = iota // reject unknown hosts
	PolicyAcceptNew                      // default: accept unknown, reject changed
	PolicyOff                            // allow anything (refuses to run under PORTUNIX_CI=1)
)

// ParseHostKeyPolicy maps the CLI flag values to the typed enum.
func ParseHostKeyPolicy(s string) (HostKeyPolicy, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "accept-new":
		return PolicyAcceptNew, nil
	case "strict":
		return PolicyStrict, nil
	case "off", "no", "insecure":
		return PolicyOff, nil
	default:
		return PolicyAcceptNew, fmt.Errorf("unknown host-key policy %q (expected: strict, accept-new, off)", s)
	}
}

// KnownHostsStore holds paths to the Portunix-local and optional system-wide
// known_hosts files plus the host-key policy.
type KnownHostsStore struct {
	LocalPath      string
	SystemPath     string // empty if --use-system-known-hosts was not set
	Policy         HostKeyPolicy
	mu             sync.Mutex
	localCallback  ssh.HostKeyCallback
	systemCallback ssh.HostKeyCallback
}

// DefaultLocalKnownHosts returns ~/.portunix/ssh/known_hosts.
func DefaultLocalKnownHosts() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".portunix", "ssh")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "known_hosts"), nil
}

// DefaultSystemKnownHosts returns ~/.ssh/known_hosts.
func DefaultSystemKnownHosts() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh", "known_hosts"), nil
}

// NewKnownHostsStore builds a store and ensures the local file exists.
func NewKnownHostsStore(local, system string, policy HostKeyPolicy) (*KnownHostsStore, error) {
	if policy == PolicyOff && os.Getenv(envCIFlag) == "1" {
		return nil, errors.New("host-key-check=off is refused under PORTUNIX_CI=1")
	}
	if local == "" {
		var err error
		local, err = DefaultLocalKnownHosts()
		if err != nil {
			return nil, err
		}
	}
	if err := ensureFile(local); err != nil {
		return nil, fmt.Errorf("create known_hosts: %w", err)
	}
	kh := &KnownHostsStore{LocalPath: local, SystemPath: system, Policy: policy}
	if err := kh.reload(); err != nil {
		return nil, err
	}
	return kh, nil
}

func (k *KnownHostsStore) reload() error {
	k.mu.Lock()
	defer k.mu.Unlock()
	cb, err := knownhosts.New(k.LocalPath)
	if err != nil {
		return err
	}
	k.localCallback = cb
	if k.SystemPath != "" {
		if _, err := os.Stat(k.SystemPath); err == nil {
			sys, err := knownhosts.New(k.SystemPath)
			if err == nil {
				k.systemCallback = sys
			}
		}
	}
	return nil
}

// Callback returns an ssh.HostKeyCallback implementing the configured policy.
func (k *KnownHostsStore) Callback() ssh.HostKeyCallback {
	if k.Policy == PolicyOff {
		return k.offCallback
	}
	return k.policyCallback
}

func (k *KnownHostsStore) offCallback(host string, remote net.Addr, key ssh.PublicKey) error {
	fmt.Fprintf(os.Stderr,
		"WARNING: host-key-check=off — not verifying host key for %s (%s). This is insecure.\n",
		host, remote)
	return nil
}

func (k *KnownHostsStore) policyCallback(host string, remote net.Addr, key ssh.PublicKey) error {
	k.mu.Lock()
	local := k.localCallback
	sys := k.systemCallback
	k.mu.Unlock()

	if local != nil {
		if err := local(host, remote, key); err == nil {
			return nil
		} else if !isKeyNotFound(err) {
			return fmt.Errorf("host key mismatch for %s: %w", host, err)
		}
	}
	if sys != nil {
		if err := sys(host, remote, key); err == nil {
			return nil
		} else if !isKeyNotFound(err) {
			return fmt.Errorf("host key mismatch for %s (system known_hosts): %w", host, err)
		}
	}

	// Unknown host
	switch k.Policy {
	case PolicyStrict:
		return fmt.Errorf("host %s key is not in known_hosts; refusing under host-key-check=strict", host)
	case PolicyAcceptNew:
		if err := k.Add(host, key); err != nil {
			return fmt.Errorf("record host key for %s: %w", host, err)
		}
		fmt.Fprintf(os.Stderr, "ptx-ssh: recorded new host key for %s (%s)\n", host, key.Type())
		return nil
	default:
		return nil
	}
}

// Add appends a host key to the local known_hosts file.
func (k *KnownHostsStore) Add(host string, key ssh.PublicKey) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	line := knownhosts.Line([]string{host}, key) + "\n"
	f, err := os.OpenFile(k.LocalPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(line); err != nil {
		return err
	}
	cb, err := knownhosts.New(k.LocalPath)
	if err != nil {
		return err
	}
	k.localCallback = cb
	return nil
}

// List returns every line of the local known_hosts file.
func (k *KnownHostsStore) List() ([]string, error) {
	data, err := os.ReadFile(k.LocalPath)
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, ln := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(ln)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		lines = append(lines, trimmed)
	}
	return lines, nil
}

// Remove deletes every entry matching host (exact match or as a hashed host).
func (k *KnownHostsStore) Remove(host string) (int, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	data, err := os.ReadFile(k.LocalPath)
	if err != nil {
		return 0, err
	}
	removed := 0
	var out []string
	for _, ln := range strings.Split(string(data), "\n") {
		if matchesHost(ln, host) {
			removed++
			continue
		}
		out = append(out, ln)
	}
	if err := os.WriteFile(k.LocalPath, []byte(strings.Join(out, "\n")), 0o600); err != nil {
		return removed, err
	}
	cb, err := knownhosts.New(k.LocalPath)
	if err != nil {
		return removed, err
	}
	k.localCallback = cb
	return removed, nil
}

func matchesHost(line, host string) bool {
	fields := strings.Fields(line)
	if len(fields) < 1 {
		return false
	}
	for _, h := range strings.Split(fields[0], ",") {
		if hostPatternMatches(h, host) {
			return true
		}
	}
	return false
}

// hostPatternMatches compares a single known_hosts pattern against a user
// query. It handles three forms of input/storage:
//   - plain "host" — exact match only
//   - "host:port" — matches stored "[host]:port" or plain "host"
//   - "[host]:port" — literal match against the same form
//
// and the same three forms of stored pattern.
func hostPatternMatches(pattern, host string) bool {
	if pattern == host {
		return true
	}
	// Normalise bracketed form [host]:port → "host:port" and bare "host"
	if patHost, patPort, ok := stripBrackets(pattern); ok {
		if patHost+":"+patPort == host {
			return true
		}
		if patHost == host {
			return true
		}
	}
	// Normalise user input the same way so "[host]:port" query matches a
	// plain "host:port" stored entry (uncommon but symmetric).
	if qHost, qPort, ok := stripBrackets(host); ok {
		if qHost+":"+qPort == pattern {
			return true
		}
		if qHost == pattern {
			return true
		}
	}
	return false
}

// stripBrackets parses "[host]:port" and returns (host, port, true). Returns
// ("", "", false) if the input is not in the bracketed form.
func stripBrackets(s string) (host, port string, ok bool) {
	if !strings.HasPrefix(s, "[") {
		return "", "", false
	}
	end := strings.Index(s, "]")
	if end <= 1 {
		return "", "", false
	}
	host = s[1:end]
	rest := s[end+1:]
	if strings.HasPrefix(rest, ":") {
		port = rest[1:]
	}
	return host, port, true
}

func isKeyNotFound(err error) bool {
	var ke *knownhosts.KeyError
	if !errors.As(err, &ke) {
		return false
	}
	return len(ke.Want) == 0
}

func ensureFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	return f.Close()
}
