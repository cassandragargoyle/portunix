/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"portunix.ai/portunix/src/helpers/ptx-github/internal/auth"
	"portunix.ai/portunix/src/helpers/ptx-github/internal/ghclient"
)

// openStore opens the encrypted token store, prompting for a password if
// --password is set.
func openStore() (*auth.Store, error) {
	password := ""
	if flagPassword {
		var err error
		password, err = readPassword("Password: ")
		if err != nil {
			return nil, err
		}
	}
	return auth.NewStore(flagStorePath, password)
}

// authManager builds an auth.Manager backed by the configured store.
func authManager() (*auth.Manager, error) {
	store, err := openStore()
	if err != nil {
		return nil, err
	}
	return auth.NewManager(store), nil
}

// newClient returns a ghclient.Client honouring the persistent flags.
func newClient(ctx context.Context) (*ghclient.Client, error) {
	mgr, err := authManager()
	if err != nil {
		return nil, err
	}
	return ghclient.New(ctx, mgr, flagAccount)
}

// readPassword prompts on stderr without echo and returns the entered text.
func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	bytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(bytes), nil
}

// readLine reads a single line of input from stdin (with prompt on stderr).
func readLine(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	var line string
	if _, err := fmt.Fscanln(os.Stdin, &line); err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// parseOwnerRepo splits an "owner/repo" spec, rejecting full URLs.
// Use parseOwnerRepoFlexible to accept URLs as well.
func parseOwnerRepo(spec string) (owner, repo string, err error) {
	parts := strings.SplitN(spec, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expected owner/repo, got %q", spec)
	}
	return parts[0], strings.TrimSuffix(parts[1], ".git"), nil
}

// parseOwnerRepoFlexible accepts owner/repo or any github URL.
func parseOwnerRepoFlexible(spec string) (owner, repo string, err error) {
	spec = strings.TrimSpace(spec)
	if strings.Contains(spec, "github.com") {
		i := strings.Index(spec, "github.com")
		spec = strings.TrimLeft(spec[i+len("github.com"):], "/:")
	}
	spec = strings.TrimSuffix(spec, ".git")
	return parseOwnerRepo(spec)
}
