/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// defaultProfileName is used when --profile is not supplied during login and
// the config currently has no profiles.
const defaultProfileName = "default"

// newAuthCmd builds the `proxmox auth` subtree. Subcommands: login, status,
// logout, list.
func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Proxmox VE connection profiles",
		Long: `Configure, inspect, and remove Proxmox VE connection profiles.

Credentials are stored in the user config directory (proxmox.json, mode 0600)
with support for multiple named profiles. Token authentication is recommended
for automation; password authentication acquires a short-lived ticket and does
not persist the password.`,
	}
	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthStatusCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthListCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var (
		host        string
		port        int
		user        string
		tokenID     string
		tokenSecret string
		profileName string
		insecure    bool
		noVerify    bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save a Proxmox VE connection profile",
		Long: `Store connection details for a Proxmox VE endpoint.

Examples:
  # API token (recommended for automation)
  portunix proxmox auth login --host pve.example.com \
      --token-id user@pam!automation --token-secret <secret>

  # Username + password (ticket acquired on every session; password not stored)
  portunix proxmox auth login --host pve.example.com --user root@pam`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthLogin(host, port, user, tokenID, tokenSecret, profileName, insecure || noVerify)
		},
	}
	f := cmd.Flags()
	f.StringVar(&host, "host", "", "Proxmox VE hostname or IP (required)")
	f.IntVar(&port, "port", 8006, "API port")
	f.StringVar(&user, "user", "", "Username in user@realm form (for password auth)")
	f.StringVar(&tokenID, "token-id", "", "API token ID in user@realm!tokenname form")
	f.StringVar(&tokenSecret, "token-secret", "", "API token secret")
	f.StringVar(&profileName, "profile", "", "Profile name (defaults to 'default' or host)")
	f.BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (use for self-signed certs)")
	// --no-verify is a friendlier alias for --insecure.
	f.BoolVar(&noVerify, "no-verify", false, "Alias for --insecure")
	_ = cmd.MarkFlagRequired("host")
	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	var profileName string
	var skipValidation bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show active Proxmox profile and validate credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthStatus(profileName, skipValidation)
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "Profile to inspect (defaults to current)")
	cmd.Flags().BoolVar(&skipValidation, "no-validate", false, "Only print stored configuration, skip /version call")
	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	var profileName string
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove a stored Proxmox profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthLogout(profileName)
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "Profile to remove (defaults to current)")
	return cmd
}

func newAuthListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List stored Proxmox profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthList()
		},
	}
}

func runAuthLogin(host string, port int, user, tokenID, tokenSecret, profileName string, insecure bool) error {
	if host == "" {
		return errors.New("--host is required")
	}

	var authType AuthType
	switch {
	case tokenID != "" && tokenSecret != "":
		authType = AuthTypeToken
	case tokenID != "" || tokenSecret != "":
		return errors.New("both --token-id and --token-secret must be provided together")
	case user != "":
		authType = AuthTypePassword
	default:
		return errors.New("provide either --token-id/--token-secret or --user")
	}

	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	if profileName == "" {
		if len(cfg.Profiles) == 0 {
			profileName = defaultProfileName
		} else {
			profileName = host
		}
	}

	profile := &Profile{
		Host:        host,
		Port:        port,
		User:        user,
		AuthType:    authType,
		TokenID:     tokenID,
		TokenSecret: tokenSecret,
		VerifyTLS:   !insecure,
	}

	// Validate credentials before persisting, so a typo doesn't leave a broken
	// profile lingering in the config file.
	client, err := NewClient(profile)
	if err != nil {
		return err
	}
	if authType == AuthTypePassword {
		password, err := promptPassword(fmt.Sprintf("Password for %s: ", user))
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
		if err := client.Login(password); err != nil {
			return fmt.Errorf("validate credentials: %w", err)
		}
	}
	info, err := client.Version()
	if err != nil {
		return fmt.Errorf("validate connection: %w", err)
	}

	cfg.Profiles[profileName] = profile
	cfg.CurrentProfile = profileName
	path, err := saveConfig(cfg)
	if err != nil {
		return err
	}
	fmt.Printf("Saved profile %q to %s\n", profileName, path)
	fmt.Printf("Proxmox VE %s (%s)\n", info.Version, info.Release)
	if !profile.VerifyTLS {
		fmt.Fprintln(os.Stderr, "Warning: TLS certificate verification is disabled for this profile.")
	}
	if authType == AuthTypeToken {
		fmt.Fprintln(os.Stderr, "Note: token secret is stored in plain JSON (file mode 0600). Revoke via Proxmox UI if leaked.")
	}
	return nil
}

func runAuthStatus(profileName string, skipValidation bool) error {
	cfg, path, err := loadConfig()
	if err != nil {
		return err
	}
	name, profile, err := resolveProfile(cfg, profileName)
	if err != nil {
		return err
	}
	fmt.Printf("Config file:    %s\n", path)
	fmt.Printf("Profile:        %s\n", name)
	fmt.Printf("Host:           %s:%d\n", profile.Host, profile.Port)
	fmt.Printf("Auth type:      %s\n", profile.AuthType)
	if profile.AuthType == AuthTypeToken {
		fmt.Printf("Token ID:       %s\n", profile.TokenID)
		fmt.Printf("Token secret:   %s\n", maskSecret(profile.TokenSecret))
	} else {
		fmt.Printf("User:           %s\n", profile.User)
	}
	fmt.Printf("TLS verify:     %t\n", profile.VerifyTLS)

	if skipValidation {
		return nil
	}
	if profile.AuthType == AuthTypePassword {
		fmt.Println("Connection:     not validated (password profiles require interactive login)")
		return nil
	}
	client, err := NewClient(profile)
	if err != nil {
		return err
	}
	info, err := client.Version()
	if err != nil {
		fmt.Printf("Connection:     FAILED — %v\n", err)
		return err
	}
	fmt.Printf("Connection:     ok (Proxmox VE %s / %s)\n", info.Version, info.Release)
	return nil
}

func runAuthLogout(profileName string) error {
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	name, _, err := resolveProfile(cfg, profileName)
	if err != nil {
		return err
	}
	delete(cfg.Profiles, name)
	if cfg.CurrentProfile == name {
		cfg.CurrentProfile = ""
		// If exactly one profile is left, promote it to current so subsequent
		// commands keep working without --profile.
		if len(cfg.Profiles) == 1 {
			for n := range cfg.Profiles {
				cfg.CurrentProfile = n
			}
		}
	}
	path, err := saveConfig(cfg)
	if err != nil {
		return err
	}
	fmt.Printf("Removed profile %q (%s)\n", name, path)
	return nil
}

func runAuthList() error {
	cfg, path, err := loadConfig()
	if err != nil {
		return err
	}
	if len(cfg.Profiles) == 0 {
		fmt.Printf("No profiles configured in %s\n", path)
		return nil
	}
	names := make([]string, 0, len(cfg.Profiles))
	for n := range cfg.Profiles {
		names = append(names, n)
	}
	sort.Strings(names)
	fmt.Printf("%-20s %-30s %-10s %s\n", "PROFILE", "HOST", "AUTH", "CURRENT")
	fmt.Printf("%-20s %-30s %-10s %s\n", "-------", "----", "----", "-------")
	for _, n := range names {
		p := cfg.Profiles[n]
		mark := ""
		if n == cfg.CurrentProfile {
			mark = "*"
		}
		fmt.Printf("%-20s %-30s %-10s %s\n",
			n, fmt.Sprintf("%s:%d", p.Host, p.Port), string(p.AuthType), mark)
	}
	return nil
}

// promptPassword reads a password from stdin without echoing. If stdin is not
// a terminal (e.g., piped input), it reads the line directly.
func promptPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	defer fmt.Fprintln(os.Stderr)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	var line string
	if _, err := fmt.Fscanln(os.Stdin, &line); err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// maskSecret redacts all but the final 4 chars of a secret for display.
func maskSecret(s string) string {
	if s == "" {
		return "(empty)"
	}
	if len(s) <= 4 {
		return strings.Repeat("*", len(s))
	}
	return strings.Repeat("*", len(s)-4) + s[len(s)-4:]
}
