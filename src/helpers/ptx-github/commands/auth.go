/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"portunix.ai/portunix/src/helpers/ptx-github/internal/auth"
)

var (
	authFlagToken      string
	authFlagAccount    string
	authFlagDefault    bool
	authFlagSkipVerify bool
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage GitHub authentication",
	}
	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthStatusCmd())
	cmd.AddCommand(newAuthListCmd())
	cmd.AddCommand(newAuthSetDefaultCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store a GitHub personal access token",
		Long: `Store a GitHub personal access token in the encrypted on-disk store.

By default the token is read from a hidden prompt. Pipe a token via stdin
or pass --token to use it non-interactively. Use --account to manage
multiple identities.

Examples:
  portunix github auth login
  portunix github auth login --account work
  echo "$TOKEN" | portunix github auth login --token -
  portunix github auth login --token ghp_xxxx --account ci --default`,
		RunE: runAuthLogin,
	}
	cmd.Flags().StringVar(&authFlagToken, "token", "",
		"token value (use '-' to read from stdin)")
	cmd.Flags().StringVar(&authFlagAccount, "name", "default",
		"account name to store the token under")
	cmd.Flags().BoolVar(&authFlagDefault, "default", false,
		"make this account the default")
	cmd.Flags().BoolVar(&authFlagSkipVerify, "skip-verify", false,
		"do not validate the token against GitHub before saving")
	return cmd
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	store, err := openStore()
	if err != nil {
		return err
	}

	token := authFlagToken
	switch token {
	case "":
		token, err = readPassword("GitHub token: ")
		if err != nil {
			return err
		}
	case "-":
		var buf [4096]byte
		n, err := os.Stdin.Read(buf[:])
		if err != nil && n == 0 {
			return fmt.Errorf("read token from stdin: %w", err)
		}
		token = string(buf[:n])
	}
	token = trimToken(token)
	if token == "" {
		return fmt.Errorf("token is empty")
	}

	username := ""
	if !authFlagSkipVerify {
		// Validate token without persisting it: drop it into the env-var
		// resolver, run /user, then unset.
		prev, hadPrev := os.LookupEnv(auth.EnvTokenVars[0])
		os.Setenv(auth.EnvTokenVars[0], token)
		defer func() {
			if hadPrev {
				os.Setenv(auth.EnvTokenVars[0], prev)
			} else {
				os.Unsetenv(auth.EnvTokenVars[0])
			}
		}()

		uname, scopes, err := auth.NewManager(nil).Validate(context.Background(), "")
		if err != nil {
			return fmt.Errorf("token validation failed: %w (use --skip-verify to override)", err)
		}
		username = uname
		fmt.Fprintf(os.Stderr, "Authenticated as %s (scopes: %s)\n",
			emptyOr(uname, "?"), emptyOr(scopes, "<none>"))
	}

	if err := store.SaveToken(authFlagAccount, token, username, authFlagDefault, flagPassword); err != nil {
		return err
	}
	fmt.Printf("Stored token for account %q at %s\n", authFlagAccount, store.Path())
	return nil
}

func newAuthLogoutCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove a stored GitHub account",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}
			if name == "" {
				_, def, err := store.List()
				if err != nil {
					return err
				}
				name = def
			}
			if name == "" {
				return fmt.Errorf("no account specified and no default set")
			}
			if err := store.Delete(name); err != nil {
				return err
			}
			fmt.Printf("Removed account %q\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "account to remove (defaults to active default)")
	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := authManager()
			if err != nil {
				return err
			}
			resolved, err := mgr.Resolve(flagAccount)
			if err != nil {
				return err
			}

			if flagJSON {
				return writeJSON(map[string]any{
					"source":        resolved.Source.String(),
					"env_var":       resolved.EnvVarName,
					"account":       resolved.AccountName,
					"username":      resolved.Username,
					"authenticated": resolved.Token != "",
				})
			}

			if resolved.Token == "" {
				fmt.Println("Not authenticated. Run 'portunix github auth login' or set GITHUB_TOKEN.")
				return nil
			}
			fmt.Printf("Source:   %s\n", resolved.Source)
			if resolved.EnvVarName != "" {
				fmt.Printf("Env var:  %s\n", resolved.EnvVarName)
			}
			if resolved.AccountName != "" {
				fmt.Printf("Account:  %s\n", resolved.AccountName)
			}

			username, scopes, err := mgr.Validate(context.Background(), flagAccount)
			if err != nil {
				fmt.Printf("User:     <validation failed: %v>\n", err)
				return nil
			}
			fmt.Printf("User:     %s\n", username)
			if scopes != "" {
				fmt.Printf("Scopes:   %s\n", scopes)
			}
			return nil
		},
	}
}

func newAuthListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List stored GitHub accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}
			accounts, def, err := store.List()
			if err != nil {
				return err
			}
			sort.Slice(accounts, func(i, j int) bool {
				return accounts[i].Name < accounts[j].Name
			})

			if flagJSON {
				return writeJSON(map[string]any{
					"default":  def,
					"accounts": accounts,
				})
			}

			if len(accounts) == 0 {
				fmt.Println("No accounts stored.")
				return nil
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "DEFAULT\tNAME\tUSER\tUPDATED\tPW")
			for _, a := range accounts {
				marker := ""
				if a.Name == def {
					marker = "*"
				}
				pw := "-"
				if a.PasswordSet {
					pw = "yes"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					marker, a.Name, emptyOr(a.Username, "-"),
					emptyOr(a.UpdatedAt, "-"), pw)
			}
			return tw.Flush()
		},
	}
}

func newAuthSetDefaultCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-default <account>",
		Short: "Change the default account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}
			if err := store.SetDefault(args[0]); err != nil {
				return err
			}
			fmt.Printf("Default account set to %q\n", args[0])
			return nil
		},
	}
}

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func emptyOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func trimToken(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r' || s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	return s
}
