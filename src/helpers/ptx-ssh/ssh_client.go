/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// ClientOptions bundles per-invocation configuration for the native SSH client.
type ClientOptions struct {
	Target         Target
	Auth           AuthResolution
	KnownHosts     *KnownHostsStore
	ConnectTimeout time.Duration
}

// DialClient opens an SSH connection using the resolved credentials and
// configured host-key policy. Callers must Close the returned client.
func DialClient(opts ClientOptions) (*ssh.Client, error) {
	cfg := &ssh.ClientConfig{
		User:            opts.Auth.Username,
		HostKeyCallback: opts.KnownHosts.Callback(),
		Timeout:         opts.ConnectTimeout,
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 15 * time.Second
	}

	methods, err := buildAuthMethods(opts.Auth)
	if err != nil {
		return nil, err
	}
	cfg.Auth = methods

	addr := net.JoinHostPort(opts.Target.Host, opts.Target.Port)
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	return client, nil
}

func buildAuthMethods(auth AuthResolution) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if auth.KeyPath != "" {
		raw, err := os.ReadFile(auth.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("read identity %s: %w", auth.KeyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(raw)
		if err != nil {
			// Encrypted key path: prompt for passphrase if on TTY.
			if _, ok := err.(*ssh.PassphraseMissingError); ok && IsInteractive() {
				pass, perr := PromptPassword(fmt.Sprintf("Passphrase for %s: ", auth.KeyPath))
				if perr != nil {
					return nil, perr
				}
				signer, err = ssh.ParsePrivateKeyWithPassphrase(raw, []byte(pass.Reveal()))
				if err != nil {
					return nil, fmt.Errorf("parse identity %s: %w", auth.KeyPath, err)
				}
			} else {
				return nil, fmt.Errorf("parse identity %s: %w", auth.KeyPath, err)
			}
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if !auth.Password.IsEmpty() {
		// PasswordCallback allows the library to retry once on auth-required events.
		methods = append(methods, ssh.PasswordCallback(func() (string, error) {
			return auth.Password.Reveal(), nil
		}))
		// keyboard-interactive is a common alternative; reuse the same secret.
		methods = append(methods, ssh.KeyboardInteractive(
			func(name, instruction string, questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range questions {
					answers[i] = auth.Password.Reveal()
				}
				return answers, nil
			}))
	}
	if len(methods) == 0 {
		return nil, errors.New("no auth method resolved")
	}
	return methods, nil
}

// RunConnect runs an interactive shell session with a local pty.
func RunConnect(opts ClientOptions) error {
	client, err := DialClient(opts)
	if err != nil {
		return err
	}
	defer client.Close()
	return runInteractive(client)
}

func runInteractive(client *ssh.Client) error {
	sess, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer sess.Close()

	stdinFd := int(os.Stdin.Fd())
	var restore func()
	if term.IsTerminal(stdinFd) {
		w, h, _ := term.GetSize(stdinFd)
		if w == 0 || h == 0 {
			w, h = 80, 24
		}
		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}
		termType := os.Getenv("TERM")
		if termType == "" {
			termType = "xterm-256color"
		}
		if err := sess.RequestPty(termType, h, w, modes); err != nil {
			return fmt.Errorf("request pty: %w", err)
		}
		oldState, err := term.MakeRaw(stdinFd)
		if err != nil {
			return fmt.Errorf("make raw: %w", err)
		}
		restore = func() { _ = term.Restore(stdinFd, oldState) }
		defer restore()
	}
	sess.Stdin = os.Stdin
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	// Forward SIGINT/SIGTERM to restore the terminal cleanly.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	defer signal.Stop(sigs)
	go func() {
		<-sigs
		if restore != nil {
			restore()
		}
	}()

	if err := sess.Shell(); err != nil {
		return fmt.Errorf("start shell: %w", err)
	}
	if err := sess.Wait(); err != nil {
		var exit *ssh.ExitError
		if errors.As(err, &exit) {
			return fmt.Errorf("remote shell exited with status %d", exit.ExitStatus())
		}
		return err
	}
	return nil
}

// RunExec runs a single non-interactive command and streams stdout/stderr.
// Returns the remote exit status (0 on success).
func RunExec(opts ClientOptions, cmd string) (int, error) {
	client, err := DialClient(opts)
	if err != nil {
		return 1, err
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return 1, fmt.Errorf("new session: %w", err)
	}
	defer sess.Close()

	sess.Stdin = os.Stdin
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	if err := sess.Run(cmd); err != nil {
		return exitStatusFromError(err), nil
	}
	return 0, nil
}

// exitStatusFromError extracts a remote exit code from a session error. Tries
// errors.As / direct type assertion first, then parses common string forms as
// a safety net because some error shapes returned by the ssh package do not
// chain cleanly.
func exitStatusFromError(err error) int {
	if err == nil {
		return 0
	}
	var exit *ssh.ExitError
	if errors.As(err, &exit) {
		return exit.ExitStatus()
	}
	if e, ok := err.(*ssh.ExitError); ok {
		return e.ExitStatus()
	}
	msg := err.Error()
	var code int
	if _, perr := fmt.Sscanf(msg, "Process exited with status %d", &code); perr == nil {
		return code
	}
	if _, perr := fmt.Sscanf(msg, "exit status %d", &code); perr == nil {
		return code
	}
	return 1
}

// RunExecCapture runs a command and returns stdout/stderr (no streaming).
func RunExecCapture(opts ClientOptions, cmd string) (stdout, stderr string, exitCode int, err error) {
	client, derr := DialClient(opts)
	if derr != nil {
		return "", "", 1, derr
	}
	defer client.Close()

	sess, serr := client.NewSession()
	if serr != nil {
		return "", "", 1, fmt.Errorf("new session: %w", serr)
	}
	defer sess.Close()

	var outBuf, errBuf bytes.Buffer
	sess.Stdout = &outBuf
	sess.Stderr = &errBuf

	runErr := sess.Run(cmd)
	if runErr != nil {
		return outBuf.String(), errBuf.String(), exitStatusFromError(runErr), nil
	}
	return outBuf.String(), errBuf.String(), 0, nil
}

// CopyEndpoint describes one side of an scp-style transfer. Remote is true
// when Path is prefixed with user@host:.
type CopyEndpoint struct {
	Remote bool
	Target Target // populated only when Remote
	Path   string
}

// ParseCopyEndpoint recognises "user@host:/path", "user@host:port:/path",
// and "/local/path". When a remote endpoint has no user, defaultUser is used.
func ParseCopyEndpoint(raw string, defaultUser string) (CopyEndpoint, error) {
	ep := CopyEndpoint{}
	// On Windows, a bare path may start with "C:" — guard against treating that as remote.
	if len(raw) >= 2 && raw[1] == ':' && isDriveLetter(raw[0]) {
		ep.Path = raw
		return ep, nil
	}
	colon := strings.Index(raw, ":")
	if colon < 0 {
		ep.Path = raw
		return ep, nil
	}
	spec := raw[:colon]
	rest := raw[colon+1:]

	// Detect user@host:port:/path — when rest starts with digits followed by a
	// second colon, fold the digits into the target spec as the port.
	if sep := strings.IndexAny(rest, ":/"); sep > 0 && rest[sep] == ':' {
		digits := rest[:sep]
		if isAllDigits(digits) {
			spec = spec + ":" + digits
			rest = rest[sep+1:]
		}
	}

	target, err := ParseTarget(spec)
	if err != nil {
		return ep, err
	}
	if target.User == "" {
		target.User = defaultUser
	}
	ep.Remote = true
	ep.Target = target
	ep.Path = rest
	return ep, nil
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isDriveLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

// RunCopy executes an SCP/SFTP-style transfer between local and remote paths.
// Exactly one of src/dst must be remote; both-remote and both-local are rejected.
func RunCopy(opts ClientOptions, src, dst CopyEndpoint) error {
	switch {
	case src.Remote && dst.Remote:
		return errors.New("remote-to-remote copy is not supported; run copy twice")
	case !src.Remote && !dst.Remote:
		return errors.New("at least one endpoint must be remote (use 'cp' for local copies)")
	}

	client, err := DialClient(opts)
	if err != nil {
		return err
	}
	defer client.Close()

	sc, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("sftp: %w", err)
	}
	defer sc.Close()

	if src.Remote {
		return downloadSFTP(sc, src.Path, dst.Path)
	}
	return uploadSFTP(sc, src.Path, dst.Path)
}

func uploadSFTP(sc *sftp.Client, localPath, remotePath string) error {
	lf, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", localPath, err)
	}
	defer lf.Close()

	// If remote is a directory, append the local base name.
	if info, err := sc.Stat(remotePath); err == nil && info.IsDir() {
		remotePath = path(remotePath, filepath.Base(localPath))
	}

	rf, err := sc.Create(remotePath)
	if err != nil {
		return fmt.Errorf("create remote %s: %w", remotePath, err)
	}
	defer rf.Close()

	if _, err := io.Copy(rf, lf); err != nil {
		return fmt.Errorf("copy to remote: %w", err)
	}
	// Preserve local mode where possible.
	if info, err := lf.Stat(); err == nil {
		_ = sc.Chmod(remotePath, info.Mode().Perm())
	}
	return nil
}

func downloadSFTP(sc *sftp.Client, remotePath, localPath string) error {
	rf, err := sc.Open(remotePath)
	if err != nil {
		return fmt.Errorf("open remote %s: %w", remotePath, err)
	}
	defer rf.Close()

	// If local is a directory, append the remote base name.
	if info, err := os.Stat(localPath); err == nil && info.IsDir() {
		localPath = filepath.Join(localPath, filepath.Base(remotePath))
	}

	lf, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("create local %s: %w", localPath, err)
	}
	defer lf.Close()
	if _, err := io.Copy(lf, rf); err != nil {
		return fmt.Errorf("copy from remote: %w", err)
	}
	return nil
}

// path joins remote paths using forward slashes regardless of the local OS.
func path(base, name string) string {
	if strings.HasSuffix(base, "/") {
		return base + name
	}
	return base + "/" + name
}
