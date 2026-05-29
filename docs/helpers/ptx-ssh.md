# ptx-ssh — Unified SSH Client Helper

## Overview

`ptx-ssh` consolidates client-side SSH operations under a single top-level
command, `portunix ssh` (plus the `portunix scp` alias). It replaces the
scattered SSH touchpoints in the codebase with a helper that:

- runs SSH natively through `golang.org/x/crypto/ssh` by default, so no
  external `ssh` binary is required;
- supports **non-interactive password authentication** backed by the
  `ptx-credential` store, without the password appearing in `ps`, shell
  history, or `/proc/*/environ`;
- enforces a safe host-key policy (`accept-new` on first use, strict on
  change) with a Portunix-local `known_hosts` at `~/.portunix/ssh/known_hosts`;
- exposes interactive username/password prompts on TTY for ad-hoc use;
- provides a wrapper mode that exec's the system `ssh`/`scp`/`rsync` under a
  pty for tools that only accept passwords via a terminal.

See Issue #174 and ADR-001 for the full motivation.

## Commands

| Command | Purpose |
| ------- | ------- |
| `portunix ssh connect <user@host>` | Interactive SSH shell |
| `portunix ssh exec <user@host> <cmd>` | Run a single remote command |
| `portunix ssh copy <src> <dst>` | SFTP transfer (either side may be `user@host:/path`) |
| `portunix scp <src> <dst>` | Alias for `ssh copy` |
| `portunix ssh rsync <src> <dst>` | Rsync over SSH (requires local `rsync`) |
| `portunix ssh bootstrap-key <user@host>` | Push public key using password auth |
| `portunix ssh trust <user@host>` | Record remote host key |
| `portunix ssh known-hosts list` | List entries in the Portunix store |
| `portunix ssh known-hosts remove <host>` | Remove entries from the store |
| `portunix ssh agent start\|add\|stop` | `ssh-agent` lifecycle |
| `portunix ssh sshpass-compat <credential-id>` | Emit a one-shot password fd/fifo |

## Credential Flags

Exactly one of the non-identity sources may be specified. Identity-based auth
is the default and is preferred.

| Flag | Source | Notes |
| ---- | ------ | ----- |
| `-i, --identity <path>` | Private key file | Preferred default |
| `--credential <id>` | `ptx-credential` store (AES-256-GCM) | Primary recommended password source |
| `--credential-fd <N>` | Open file descriptor | For anonymous pipes in scripts |
| `--credential-file <path>` | First line of a mode-0400 / 0600 file | Permissions enforced on Unix |
| `--credential-env` | `PTX_SSH_PASSWORD` env var | Documented as less secure |
| `--ask-user` | Interactive prompt (TTY) | Username prompt when target lacks `user@` |
| `--ask-pass` | Interactive prompt (TTY) | Password prompt, no echo |
| `--credential-store <name>` | Alternate ptx-credential store | Optional modifier |

**`--password <value>` on the command line is explicitly rejected.** Attempting
to pass a plaintext password through the CLI produces an error pointing at the
recommended alternatives.

## Host-Key Policy

| Flag | Behaviour |
| ---- | --------- |
| `--host-key-check=strict` | Refuse unknown hosts; require prior `trust` or explicit known_hosts entry |
| `--host-key-check=accept-new` (default) | Record on first use; refuse on change |
| `--host-key-check=off` | Allow any key; refuses to run under `PORTUNIX_CI=1` |
| `--use-system-known-hosts` | Also consult `~/.ssh/known_hosts` |
| `--known-hosts-file <path>` | Override the default path |

## Interactive Prompts

- Missing `user@` on the target triggers a username prompt when `--ask-user`
  is set or a TTY is attached.
- When no credential source is supplied and `stdin`/`/dev/tty` is a terminal,
  ptx-ssh prompts for the password automatically (no echo).
- Prompts write to `/dev/tty` (falling back to `stderr` when no controlling
  terminal exists) so piped automation remains clean.
- Under `PORTUNIX_CI=1`, ptx-ssh refuses to prompt and fails fast.

## Examples

```bash
# Key-based interactive session (default identity ~/.ssh/id_ed25519)
portunix ssh connect alice@host.example

# Password from ptx-credential, non-interactive command
portunix ssh exec --credential legacy-admin alice@legacy.example 'uname -a'

# SFTP upload
portunix ssh copy ./build.tar.gz alice@host:/tmp/

# Bootstrap a key on a password-only host
portunix ssh bootstrap-key --credential legacy-admin alice@legacy.example

# One-shot password feed for ansible --ask-pass
ansible-playbook --ask-pass -e "ansible_ssh_pass=$(portunix ssh sshpass-compat legacy-admin --write-fd 3)" …

# Interactive prompts (username + password) on a fresh machine
portunix ssh connect host.example --ask-user --ask-pass
```

## Security Notes

- Passwords are wrapped in a `SecretString` type whose `String`, `GoString`,
  `MarshalJSON`, `MarshalYAML`, and `MarshalText` implementations all return
  `"***"`. Passing a secret to `fmt.Printf`, `log.Print`, `json.Marshal`, or
  similar will produce the mask — not the value.
- Host-key verification cannot be silently disabled: `--host-key-check=off`
  refuses to run whenever `PORTUNIX_CI=1`.
- `--credential-file` requires owner-only permissions (mode ≤ `0600`) on Unix.
- `LnxExecutePyScriptSsh` in `src/app/ssh.go` has been retired and now shells
  out to `ptx-ssh`; the legacy hardcoded-key + `InsecureIgnoreHostKey` code
  path no longer exists.

## Storage

| Path | Purpose |
| ---- | ------- |
| `~/.portunix/ssh/known_hosts` | Portunix-local known_hosts store |
| `~/.ssh/known_hosts` | Optional, consulted only with `--use-system-known-hosts` |

## References

- Issue #174 — ptx-ssh helper
- ADR-001 — ptx-ssh Helper (in `portunix-architecture`)
- ADR-002 — Extract `ptx-credential` storage into shared internal Go SDK
  (in `portunix-architecture`) — follow-up that will let ptx-ssh read the
  credential store in-process instead of shelling out to the helper
