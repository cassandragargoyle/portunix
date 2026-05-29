/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var version = "dev"

// pendingExitCode carries the remote exit status from `ssh exec` out of Cobra's
// RunE so we can propagate it with a real os.Exit after rootCmd.Execute
// returns. Setting it inside RunE would bypass deferred cleanup.
var pendingExitCode int

// Global flags (persistent on rootCmd)
var (
	flagIdentity        string
	flagCredential      string
	flagCredentialFD    int
	flagCredentialFile  string
	flagCredentialEnv   bool
	flagCredentialStore string
	flagAskUser         bool
	flagAskPass         bool
	flagHostKeyCheck    string
	flagUseSystemHosts  bool
	flagKnownHostsFile  string
	flagWrapper         bool
	flagTimeoutSeconds  int
)

func collectCredFlags() CredFlags {
	return CredFlags{
		Identity:        flagIdentity,
		Credential:      flagCredential,
		CredentialFD:    flagCredentialFD,
		CredentialFile:  flagCredentialFile,
		CredentialEnv:   flagCredentialEnv,
		CredentialStore: flagCredentialStore,
		AskUser:         flagAskUser,
		AskPass:         flagAskPass,
	}
}

// buildOptions turns CLI flags into a ClientOptions ready for DialClient.
// It is shared by every command that opens a connection.
func buildOptions(targetStr string) (ClientOptions, error) {
	target, err := ParseTarget(targetStr)
	if err != nil {
		return ClientOptions{}, err
	}
	flags := collectCredFlags()
	if err := flags.Validate(); err != nil {
		return ClientOptions{}, err
	}
	policy, err := ParseHostKeyPolicy(flagHostKeyCheck)
	if err != nil {
		return ClientOptions{}, err
	}
	sysPath := ""
	if flagUseSystemHosts {
		sysPath, _ = DefaultSystemKnownHosts()
	}
	kh, err := NewKnownHostsStore(flagKnownHostsFile, sysPath, policy)
	if err != nil {
		return ClientOptions{}, err
	}
	auth, err := ResolveAuth(&target, flags)
	if err != nil {
		return ClientOptions{}, err
	}
	opts := ClientOptions{Target: target, Auth: auth, KnownHosts: kh}
	return opts, nil
}

var rootCmd = &cobra.Command{
	Use:     "ptx-ssh",
	Short:   "Portunix unified SSH client helper",
	Version: version,
	Long: `ptx-ssh is the Portunix helper binary for SSH operations, registered with the
dispatcher as 'portunix ssh' and 'portunix scp'. It provides a native Go SSH
client with first-class support for non-interactive password authentication
backed by the ptx-credential store.

This binary is typically invoked by the main portunix dispatcher and should
not be used directly.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// connect
var connectCmd = &cobra.Command{
	Use:   "connect <user@host>",
	Short: "Open an interactive SSH session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := buildOptions(args[0])
		if err != nil {
			return err
		}
		defer opts.Auth.Password.Zero()
		if flagWrapper {
			return RunWrapper("ssh", wrapperSSHArgs(opts, nil), opts.Auth.Password)
		}
		return RunConnect(opts)
	},
}

// exec
var execCmd = &cobra.Command{
	Use:   "exec <user@host> <command...>",
	Short: "Run a single remote command",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := buildOptions(args[0])
		if err != nil {
			return err
		}
		defer opts.Auth.Password.Zero()
		remoteCmd := strings.Join(args[1:], " ")
		if flagWrapper {
			return RunWrapper("ssh", wrapperSSHArgs(opts, []string{remoteCmd}), opts.Auth.Password)
		}
		code, err := RunExec(opts, remoteCmd)
		if err != nil {
			return err
		}
		pendingExitCode = code
		return nil
	},
}

// copy
var copyCmd = &cobra.Command{
	Use:   "copy <src> <dst>",
	Short: "Transfer files via SFTP (user@host:/path on either side)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse endpoints first so we can build options from the remote one.
		defaultUser := os.Getenv("USER")
		if defaultUser == "" {
			defaultUser = os.Getenv("USERNAME")
		}
		src, err := ParseCopyEndpoint(args[0], defaultUser)
		if err != nil {
			return fmt.Errorf("source: %w", err)
		}
		dst, err := ParseCopyEndpoint(args[1], defaultUser)
		if err != nil {
			return fmt.Errorf("destination: %w", err)
		}
		if src.Remote && dst.Remote {
			return errors.New("remote-to-remote copy is not supported; run copy twice")
		}
		if !src.Remote && !dst.Remote {
			return errors.New("at least one endpoint must be remote (use 'cp' for local copies)")
		}
		var remote Target
		if src.Remote {
			remote = src.Target
		} else {
			remote = dst.Target
		}
		opts, err := buildOptions(formatTarget(remote))
		if err != nil {
			return err
		}
		defer opts.Auth.Password.Zero()
		return RunCopy(opts, src, dst)
	},
}

// rsync
var rsyncCmd = &cobra.Command{
	Use:                "rsync <src> <dst> [-- rsync-args...]",
	Short:              "Run rsync over SSH (requires local rsync binary)",
	DisableFlagParsing: false,
	Args:               cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Split optional "-- <rsync-args>" segment.
		var extra []string
		if i := indexOf(args, "--"); i >= 0 {
			extra = args[i+1:]
			args = args[:i]
		}
		if len(args) != 2 {
			return errors.New("rsync requires exactly <src> and <dst>")
		}
		defaultUser := os.Getenv("USER")
		if defaultUser == "" {
			defaultUser = os.Getenv("USERNAME")
		}
		src, err := ParseCopyEndpoint(args[0], defaultUser)
		if err != nil {
			return err
		}
		dst, err := ParseCopyEndpoint(args[1], defaultUser)
		if err != nil {
			return err
		}
		var remote Target
		switch {
		case src.Remote && !dst.Remote:
			remote = src.Target
		case dst.Remote && !src.Remote:
			remote = dst.Target
		default:
			return errors.New("rsync over ssh requires exactly one remote endpoint")
		}
		opts, err := buildOptions(formatTarget(remote))
		if err != nil {
			return err
		}
		defer opts.Auth.Password.Zero()
		return Rsync(args[0], args[1], opts, extra)
	},
}

// bootstrap-key
var bootstrapKeyCmd = &cobra.Command{
	Use:   "bootstrap-key <user@host>",
	Short: "Push a public key to authorized_keys using password auth",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := buildOptions(args[0])
		if err != nil {
			return err
		}
		defer opts.Auth.Password.Zero()
		pub, _ := cmd.Flags().GetString("public-key")
		return RunBootstrapKey(opts, pub)
	},
}

// trust
var trustCmd = &cobra.Command{
	Use:   "trust <user@host>",
	Short: "Accept a remote host key without authentication",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, err := ParseTarget(args[0])
		if err != nil {
			return err
		}
		// Use accept-new policy with a no-op dial that just records the key.
		kh, err := NewKnownHostsStore(flagKnownHostsFile, "", PolicyAcceptNew)
		if err != nil {
			return err
		}
		return trustKey(target, kh)
	},
}

// known-hosts
var knownHostsCmd = &cobra.Command{
	Use:   "known-hosts",
	Short: "Manage the Portunix-local known_hosts store",
}

var knownHostsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List known_hosts entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		kh, err := NewKnownHostsStore(flagKnownHostsFile, "", PolicyAcceptNew)
		if err != nil {
			return err
		}
		lines, err := kh.List()
		if err != nil {
			return err
		}
		for _, ln := range lines {
			fmt.Println(ln)
		}
		return nil
	},
}

var knownHostsRemoveCmd = &cobra.Command{
	Use:   "remove <host>",
	Short: "Remove known_hosts entries matching <host>",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kh, err := NewKnownHostsStore(flagKnownHostsFile, "", PolicyAcceptNew)
		if err != nil {
			return err
		}
		n, err := kh.Remove(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Removed %d entries for %s\n", n, args[0])
		return nil
	},
}

// agent
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "ssh-agent lifecycle management",
}

var agentStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start ssh-agent and print shell env",
	RunE:  func(cmd *cobra.Command, args []string) error { return AgentStart() },
}

var agentAddCmd = &cobra.Command{
	Use:   "add [identity]",
	Short: "Add an identity to the ssh-agent",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := ""
		if len(args) == 1 {
			id = args[0]
		}
		return AgentAdd(id)
	},
}

var agentStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the active ssh-agent",
	RunE:  func(cmd *cobra.Command, args []string) error { return AgentStop() },
}

// sshpass-compat
var sshpassCompatCmd = &cobra.Command{
	Use:   "sshpass-compat <credential-id>",
	Short: "Emit a one-shot password fd/fifo for ansible --ask-pass",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fd, _ := cmd.Flags().GetInt("write-fd")
		store, _ := cmd.Flags().GetString("store")
		return SshpassCompat(args[0], store, fd)
	},
}

func init() {
	p := rootCmd.PersistentFlags()
	p.StringVarP(&flagIdentity, "identity", "i", "", "Private key file (preferred default)")
	p.StringVar(&flagCredential, "credential", "", "ptx-credential entry id")
	p.IntVar(&flagCredentialFD, "credential-fd", 0, "Read password from open file descriptor")
	p.StringVar(&flagCredentialFile, "credential-file", "", "Read password from first line of mode-0400 file")
	p.BoolVar(&flagCredentialEnv, "credential-env", false, "Read password from PTX_SSH_PASSWORD env var")
	p.StringVar(&flagCredentialStore, "credential-store", "", "ptx-credential store name (optional)")
	p.BoolVar(&flagAskUser, "ask-user", false, "Prompt for username on TTY")
	p.BoolVar(&flagAskPass, "ask-pass", false, "Prompt for password on TTY (no echo)")
	p.StringVar(&flagHostKeyCheck, "host-key-check", "accept-new", "Host-key policy: strict|accept-new|off")
	p.BoolVar(&flagUseSystemHosts, "use-system-known-hosts", false, "Also consult ~/.ssh/known_hosts")
	p.StringVar(&flagKnownHostsFile, "known-hosts-file", "", "Override known_hosts path (default: ~/.portunix/ssh/known_hosts)")
	p.BoolVar(&flagWrapper, "wrapper", false, "Delegate to the system ssh/scp/rsync under a pty")
	p.IntVar(&flagTimeoutSeconds, "connect-timeout", 15, "TCP connect timeout (seconds)")

	rootCmd.SetVersionTemplate("ptx-ssh version {{.Version}}\n")

	bootstrapKeyCmd.Flags().String("public-key", "", "Public key path (default: ~/.ssh/id_ed25519.pub)")

	sshpassCompatCmd.Flags().Int("write-fd", 0, "Write password to this open fd instead of a fifo")
	sshpassCompatCmd.Flags().String("store", "", "ptx-credential store name")

	knownHostsCmd.AddCommand(knownHostsListCmd, knownHostsRemoveCmd)
	agentCmd.AddCommand(agentStartCmd, agentAddCmd, agentStopCmd)

	rootCmd.AddCommand(connectCmd, execCmd, copyCmd, rsyncCmd, bootstrapKeyCmd,
		trustCmd, knownHostsCmd, agentCmd, sshpassCompatCmd)
}

// dispatcherMetaFlags handles --version / --description / --list-commands and
// --help-ai / --help-expert before Cobra sees them, matching the convention in
// the other helpers.
func dispatcherMetaFlags(args []string) (handled bool) {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "--description":
		fmt.Println("Portunix unified SSH client and non-interactive password auth")
		return true
	case "--list-commands":
		fmt.Println("ssh")
		fmt.Println("scp")
		return true
	case "--help-ai":
		showHelpAI()
		return true
	case "--help-expert":
		showHelpExpert()
		return true
	}
	return false
}

func main() {
	// Scan for forbidden --password before anything else so the value never
	// reaches downstream loggers.
	if err := RejectPlaintextPasswordFlag(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	// Handle dispatcher meta-flags before stripping the dispatched command.
	for _, arg := range os.Args[1:] {
		if arg == "--help-ai" || arg == "--help-expert" {
			if dispatcherMetaFlags([]string{arg}) {
				return
			}
		}
	}

	// When invoked via the dispatcher as "portunix ssh ..." or "portunix scp ...",
	// the first positional arg is the command name ("ssh" or "scp"). Strip it
	// so Cobra sees our subcommand tree directly. "scp" is an alias for "copy".
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "ssh":
			os.Args = append(os.Args[:1], os.Args[2:]...)
		case "scp":
			os.Args = append([]string{os.Args[0], "copy"}, os.Args[2:]...)
		}
	}

	if len(os.Args) > 1 && dispatcherMetaFlags(os.Args[1:]) {
		return
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "ptx-ssh:", err)
		os.Exit(1)
	}
	if pendingExitCode != 0 {
		os.Exit(pendingExitCode)
	}
}

func indexOf(s []string, v string) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}

func formatTarget(t Target) string {
	if t.User == "" {
		return fmt.Sprintf("%s:%s", t.Host, t.Port)
	}
	return fmt.Sprintf("%s@%s:%s", t.User, t.Host, t.Port)
}

// wrapperSSHArgs composes the argv for a system ssh invocation from our
// parsed target and credential resolution. Passwords are injected by the pty
// helper, not on the command line.
func wrapperSSHArgs(opts ClientOptions, remoteCmd []string) []string {
	args := []string{"-p", opts.Target.Port,
		"-o", "StrictHostKeyChecking=accept-new"}
	if opts.KnownHosts != nil && opts.KnownHosts.LocalPath != "" {
		args = append(args, "-o", "UserKnownHostsFile="+opts.KnownHosts.LocalPath)
	}
	if opts.Auth.KeyPath != "" {
		args = append(args, "-i", opts.Auth.KeyPath)
	}
	args = append(args, fmt.Sprintf("%s@%s", opts.Auth.Username, opts.Target.Host))
	args = append(args, remoteCmd...)
	return args
}

// trustKey opens a throwaway TCP connection to capture the remote host key
// and persists it through the known_hosts store without authenticating.
func trustKey(target Target, kh *KnownHostsStore) error {
	// Reuse DialClient with an empty-auth config — the connection will fail on
	// auth, but the HostKeyCallback already fired by then.
	return errors.New("trust: not yet implemented; connect once and accept-new policy will record the key")
}
