/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// CredFlags collects every credential-related flag exposed by the CLI. Exactly
// one of the non-identity sources may be set; combinations must be rejected by
// Validate().
type CredFlags struct {
	Identity        string // --identity / -i <path>
	Credential      string // --credential <id>
	CredentialFD    int    // --credential-fd <N> (0 means unset)
	CredentialFile  string // --credential-file <path>
	CredentialEnv   bool   // --credential-env
	AskUser         bool   // --ask-user
	AskPass         bool   // --ask-pass
	CredentialStore string // --credential-store <name> (optional)
}

const (
	envPassword = "PTX_SSH_PASSWORD"
	envCIFlag   = "PORTUNIX_CI"
)

// AuthResolution is the outcome of credential resolution for a connection.
type AuthResolution struct {
	Username string
	Password SecretString
	KeyPath  string
}

// CredErr is a sentinel for credential-related failures that deserve a short,
// user-visible message with links to alternatives.
type CredErr struct{ Msg string }

func (e *CredErr) Error() string { return e.Msg }

// newCredErr formats a credential error with a consistent footer pointing the
// user to the recommended alternatives.
func newCredErr(msg string) *CredErr {
	return &CredErr{Msg: msg + "\n\nAllowed credential sources:\n" +
		"  --identity <path>        Private key file (preferred)\n" +
		"  --credential <id>        Password from ptx-credential store\n" +
		"  --credential-fd <N>      Open file descriptor (for scripts)\n" +
		"  --credential-file <path> First line of mode-0400 file\n" +
		"  --credential-env         PTX_SSH_PASSWORD env var (less secure)\n" +
		"  --ask-user / --ask-pass  Interactive prompt on TTY"}
}

// RejectPlaintextPasswordFlag scans raw argv for the forbidden --password flag
// and returns a credential error pointing to safe alternatives. This runs
// before Cobra parses flags so the plaintext never reaches structured logging.
func RejectPlaintextPasswordFlag(argv []string) error {
	for _, a := range argv {
		if a == "--password" || strings.HasPrefix(a, "--password=") {
			return newCredErr("--password on the command line is not supported. " +
				"The value would be visible in 'ps', /proc/*/environ, shell history, and audit logs.")
		}
	}
	return nil
}

// Validate enforces mutual exclusivity between credential sources.
func (c *CredFlags) Validate() error {
	n := 0
	if c.Credential != "" {
		n++
	}
	if c.CredentialFD > 0 {
		n++
	}
	if c.CredentialFile != "" {
		n++
	}
	if c.CredentialEnv {
		n++
	}
	if c.AskPass {
		n++
	}
	if n > 1 {
		return newCredErr("multiple credential sources specified; choose exactly one of --credential, --credential-fd, --credential-file, --credential-env, --ask-pass")
	}
	return nil
}

// Target parses a user@host[:port] target string. If user is missing and the
// caller allows interactive resolution, ResolveUsername can fill it in.
type Target struct {
	User string
	Host string
	Port string
}

// ParseTarget splits user@host[:port]. Accepts bare host without user.
func ParseTarget(raw string) (Target, error) {
	if raw == "" {
		return Target{}, errors.New("empty ssh target")
	}
	var t Target
	if at := strings.LastIndex(raw, "@"); at >= 0 {
		t.User = raw[:at]
		raw = raw[at+1:]
	}
	if raw == "" {
		return Target{}, errors.New("ssh target missing host")
	}
	// IPv6 address in brackets: [::1]:22
	if strings.HasPrefix(raw, "[") {
		end := strings.Index(raw, "]")
		if end < 0 {
			return Target{}, errors.New("unterminated IPv6 target")
		}
		t.Host = raw[1:end]
		rest := raw[end+1:]
		if strings.HasPrefix(rest, ":") {
			t.Port = rest[1:]
		}
		return t, nil
	}
	if colon := strings.LastIndex(raw, ":"); colon >= 0 && !strings.Contains(raw, "::") {
		t.Host = raw[:colon]
		t.Port = raw[colon+1:]
	} else {
		t.Host = raw
	}
	if t.Port == "" {
		t.Port = "22"
	}
	return t, nil
}

// ResolveAuth picks an authentication method given the flags. Password sources
// take precedence over identity when explicitly set; otherwise identity-first.
// If no source is provided and we are on a TTY, interactive prompts kick in.
func ResolveAuth(target *Target, flags CredFlags) (AuthResolution, error) {
	res := AuthResolution{}

	// Resolve username
	if target.User == "" {
		u, err := ResolveUsername(flags)
		if err != nil {
			return res, err
		}
		target.User = u
	}
	res.Username = target.User

	// Preferred: identity-based auth
	if flags.Identity != "" {
		abs, err := expandPath(flags.Identity)
		if err != nil {
			return res, fmt.Errorf("identity path: %w", err)
		}
		if _, err := os.Stat(abs); err != nil {
			return res, fmt.Errorf("identity file not readable: %w", err)
		}
		res.KeyPath = abs
		return res, nil
	}

	// Any explicit password source wins next
	switch {
	case flags.Credential != "":
		pwd, err := readFromPtxCredential(flags.Credential, flags.CredentialStore)
		if err != nil {
			return res, err
		}
		res.Password = pwd
	case flags.CredentialFD > 0:
		pwd, err := readFromFD(flags.CredentialFD)
		if err != nil {
			return res, err
		}
		res.Password = pwd
	case flags.CredentialFile != "":
		pwd, err := readFromFile(flags.CredentialFile)
		if err != nil {
			return res, err
		}
		res.Password = pwd
	case flags.CredentialEnv:
		v := os.Getenv(envPassword)
		if v == "" {
			return res, newCredErr(fmt.Sprintf("env var %s is empty", envPassword))
		}
		res.Password = NewSecret(v)
	case flags.AskPass:
		pwd, err := PromptPassword(fmt.Sprintf("Password for %s@%s: ", res.Username, target.Host))
		if err != nil {
			return res, err
		}
		res.Password = pwd
	default:
		// Fall back to default identity, then to interactive prompt if on TTY.
		if key, ok := defaultIdentity(); ok {
			res.KeyPath = key
			return res, nil
		}
		if IsInteractive() {
			pwd, err := PromptPassword(fmt.Sprintf("Password for %s@%s: ", res.Username, target.Host))
			if err != nil {
				return res, err
			}
			res.Password = pwd
		} else {
			return res, newCredErr("no credential source supplied and no default identity found")
		}
	}
	return res, nil
}

// ResolveUsername asks the user interactively when --ask-user is set or the
// session is on a TTY. Returns error on CI / non-TTY when no username present.
func ResolveUsername(flags CredFlags) (string, error) {
	if flags.AskUser || IsInteractive() {
		return PromptLine("Username: ")
	}
	if u := os.Getenv("USER"); u != "" {
		return u, nil
	}
	if u := os.Getenv("USERNAME"); u != "" { // Windows
		return u, nil
	}
	return "", newCredErr("no username in target and not running on a TTY; specify user@host or use --ask-user")
}

// IsInteractive reports whether prompts can be rendered. Returns false if
// PORTUNIX_CI=1 or if neither stdin nor /dev/tty is a terminal.
func IsInteractive() bool {
	if os.Getenv(envCIFlag) == "1" {
		return false
	}
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return true
	}
	// /dev/tty as a last resort on Unix-like systems.
	if runtime.GOOS != "windows" {
		f, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0)
		if err == nil {
			_ = f.Close()
			return true
		}
	}
	return false
}

// openTTYPair returns a pair of file handles for prompt output and input.
// Preference: /dev/tty on Unix; stdin + stderr otherwise. Stdout is never used
// so piped automation stays clean. On Windows we use stdin + stderr.
func openTTYPair() (in *os.File, out *os.File, cleanup func(), err error) {
	if runtime.GOOS != "windows" {
		tty, terr := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if terr == nil {
			return tty, tty, func() { _ = tty.Close() }, nil
		}
	}
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return os.Stdin, os.Stderr, func() {}, nil
	}
	return nil, nil, nil, errors.New("no controlling terminal for interactive prompt")
}

// PromptPassword reads a password without echo from the controlling terminal.
// Fails fast when no TTY is available or PORTUNIX_CI=1 is set.
func PromptPassword(prompt string) (SecretString, error) {
	if os.Getenv(envCIFlag) == "1" {
		return SecretString{}, newCredErr("cannot prompt for password under PORTUNIX_CI=1")
	}
	in, out, cleanup, err := openTTYPair()
	if err != nil {
		return SecretString{}, err
	}
	defer cleanup()

	fmt.Fprint(out, prompt)
	pwd, err := term.ReadPassword(int(in.Fd()))
	fmt.Fprintln(out)
	if err != nil {
		return SecretString{}, fmt.Errorf("read password: %w", err)
	}
	return NewSecretBytes(pwd), nil
}

// PromptLine reads one line of input (echoed) from the controlling terminal.
func PromptLine(prompt string) (string, error) {
	if os.Getenv(envCIFlag) == "1" {
		return "", newCredErr("cannot prompt for input under PORTUNIX_CI=1")
	}
	in, out, cleanup, err := openTTYPair()
	if err != nil {
		return "", err
	}
	defer cleanup()

	fmt.Fprint(out, prompt)
	rd := bufio.NewReader(in)
	line, err := rd.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func readFromFD(fd int) (SecretString, error) {
	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd-%d", fd))
	if f == nil {
		return SecretString{}, newCredErr(fmt.Sprintf("invalid fd: %d", fd))
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return SecretString{}, fmt.Errorf("read fd %d: %w", fd, err)
	}
	// Strip trailing newline only; callers embed the password verbatim.
	data = trimNewline(data)
	if len(data) == 0 {
		return SecretString{}, newCredErr(fmt.Sprintf("fd %d produced empty password", fd))
	}
	return NewSecretBytes(data), nil
}

func readFromFile(path string) (SecretString, error) {
	abs, err := expandPath(path)
	if err != nil {
		return SecretString{}, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return SecretString{}, fmt.Errorf("credential file: %w", err)
	}
	// Enforce owner-only permissions on Unix; Windows ACLs are not checked here.
	if runtime.GOOS != "windows" {
		mode := info.Mode().Perm()
		if mode&0o077 != 0 {
			return SecretString{}, newCredErr(fmt.Sprintf(
				"credential file %s has permissions %04o; must be 0400 or 0600 (no group/other access)",
				abs, mode))
		}
	}
	f, err := os.Open(abs)
	if err != nil {
		return SecretString{}, err
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	line, err := rd.ReadBytes('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return SecretString{}, err
	}
	line = trimNewline(line)
	if len(line) == 0 {
		return SecretString{}, newCredErr(fmt.Sprintf("credential file %s is empty", abs))
	}
	return NewSecretBytes(line), nil
}

// readFromPtxCredential shells out to the ptx-credential helper that lives in
// the same directory as this binary. ADR-002 will later replace this with a
// direct SDK call.
func readFromPtxCredential(id, store string) (SecretString, error) {
	bin, err := locatePtxCredential()
	if err != nil {
		return SecretString{}, err
	}
	args := []string{"get", id, "--quiet"}
	if store != "" {
		args = append(args, "--store", store)
	}
	cmd := exec.Command(bin, args...)
	cmd.Stderr = os.Stderr // surface "credential not found" etc. to the user
	out, err := cmd.Output()
	if err != nil {
		return SecretString{}, fmt.Errorf("ptx-credential get %s: %w", id, err)
	}
	// --quiet prints the value without a trailing newline; guard anyway.
	out = trimNewline(out)
	if len(out) == 0 {
		return SecretString{}, newCredErr(fmt.Sprintf("credential %q is empty", id))
	}
	return NewSecretBytes(out), nil
}

func locatePtxCredential() (string, error) {
	name := "ptx-credential"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), name)
		if _, serr := os.Stat(candidate); serr == nil {
			return candidate, nil
		}
	}
	if p, perr := exec.LookPath(name); perr == nil {
		return p, nil
	}
	return "", errors.New("ptx-credential helper not found next to ptx-ssh or in PATH")
}

func defaultIdentity() (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	for _, name := range []string{"id_ed25519", "id_ecdsa", "id_rsa"} {
		p := filepath.Join(home, ".ssh", name)
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}
	return "", false
}

func expandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") || p == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			return home, nil
		}
		return filepath.Join(home, p[2:]), nil
	}
	return filepath.Abs(p)
}

func trimNewline(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r') {
		b = b[:len(b)-1]
	}
	return b
}
