# Issue #174: Implement `ptx-ssh` Helper — Unified SSH Client and Non-Interactive Password Auth

- **Type**: Feature
- **Priority**: High
- **Status**: ✅ Implemented
- **Labels**: feature, helper, security, ssh, new-helper, portunix-core, dispatcher
- **Created**: 2026-04-18
- **Source**: `portunix-architecture/docs/architecture/components/portunix/issues/001-ptx-ssh-helper.md`
- **Architecture Decision Records**:
  - ADR-001 — ptx-ssh Helper (in `portunix-architecture` repository)
  - ADR-002 — Extract `ptx-credential` Storage into Shared Internal Go SDK (in `portunix-architecture` repository)

## Description

Implement a new helper binary `ptx-ssh` in the `portunix` repository as specified in ADR-001 from the
`portunix-architecture` repository. The helper introduces a first-class SSH client for Portunix,
registered in the dispatcher as the top-level command `portunix ssh`, with **non-interactive password
authentication** as a core capability (backed by the existing `ptx-credential` store).

## Motivation

Portunix currently has SSH functionality scattered across unrelated places and no unified top-level
`portunix ssh` command. The existing touchpoints are:

- `portunix virt ssh <vm>` — VM-specific sub-command of `ptx-virt`
- `portunix docker ssh <container>` — container-specific sub-command of `ptx-container`
- `src/app/ssh.go` (`LnxExecutePyScriptSsh`) — internal helper with hardcoded key path and
  `ssh.InsecureIgnoreHostKey()`, explicitly marked "not for production"
- `Install-PortableOpenSSH.ps1` — server-side Sandbox provisioning, not a client
- `portunix install openssh` — installer only, no runtime SSH operations

Users who need to run an SSH command against an arbitrary host today either bypass Portunix entirely
or install a third-party password-feeder utility to handle legacy hosts that require password
authentication.

ADR-001 establishes the architectural decision to consolidate client-side SSH into a dedicated helper
with secure defaults and native support for password-based flows via `ptx-credential`.

## Involved Components

| Component | Role |
| --------- | ---- |
| **`ptx-ssh`** (new) | New helper binary + dispatcher registration |
| **`ptx-credential`** | Consumer — password lookup via Go SDK (see ADR-002) |
| **`ptx-ansible`** | Downstream consumer — integration tests for `--ask-pass` workflows |
| **`ptx-virt`** | Co-existence — `virt ssh` sub-command must continue to work |
| **`ptx-container`** | Co-existence — `docker ssh` / `container ssh` sub-commands must continue to work |
| **`src/app/ssh.go`** | Legacy code — `LnxExecutePyScriptSsh` to be refactored as a shim |
| **`src/dispatcher/dispatcher.go`** | Register new helper with commands `["ssh", "scp"]` |

## Scope

### Dispatcher Registration

- Register `ptx-ssh` in `src/dispatcher/dispatcher.go` with commands `["ssh", "scp"]`
- Verify no conflict with existing top-level commands; confirm `virt ssh` and `docker ssh` sub-commands continue to work
- Add `--version`, `--description`, `--list-commands`, `--help-ai`, `--help-expert` to comply with the common helper interface (see Issue #163 for help-flag convention)

### Core SSH Client (Native Mode, `golang.org/x/crypto/ssh`)

- `portunix ssh connect <user@host>` — interactive session (allocates local pty for terminal experience)
- `portunix ssh exec <user@host> "<cmd>"` — run single command, capture stdout / stderr / exit code
- `portunix ssh copy <src> <dst>` — SCP/SFTP transfer, supports `user@host:/path` notation in either direction
- `portunix ssh rsync <src> <dst>` — rsync-over-SSH wrapper (requires local `rsync` binary)
- All subcommands accept the common credential flags (see below)

### Non-Interactive Password Authentication (First-Class)

- **Native mode** (default): authenticate using `golang.org/x/crypto/ssh` `ssh.Password(...)` auth
  method. No pty, no external ssh binary; the password never appears on the command line or in
  `/proc/*/environ`.
- **Wrapper mode** (opt-in, `--wrapper` flag): allocate a pty using `github.com/creack/pty` and exec
  the system `ssh`/`scp`/`rsync`/`ansible-playbook`, feeding the password at the prompt. Required
  for tools that refuse to accept passwords through any other channel.

### Credential Sources (Mutually Exclusive Flags)

| Flag | Source | Notes |
| ---- | ------ | ----- |
| `--identity, -i <path>` | Private key file | Standard key auth, preferred default |
| `--credential <id>` | `ptx-credential` store (AES-256-GCM) | Primary recommended password source |
| `--credential-fd <N>` | Open file descriptor | Anonymous-pipe usage by scripts |
| `--credential-file <path>` | First line of mode-0400 file | File permissions enforced |
| `--credential-env` | `PTX_SSH_PASSWORD` env var | Documented as less secure |
| `--ask-user` | Interactive prompt (TTY) | Prompts for username (when not provided in target) |
| `--ask-pass` | Interactive prompt (TTY) | Prompts for password without echo; also enabled automatically when running on a TTY with no other credential source supplied |

**MUST NOT** provide a `--password <value>` CLI flag. Attempting to pass plaintext on the command
line must produce an error with a link to the recommended alternatives.

### Interactive Credential Prompting

- When target is given without a username (e.g., `portunix ssh connect host.example`) and
  `--ask-user` is set (or the process runs on a TTY and no username is resolvable), prompt:
  `Username: ` (echoed)
- When no credential source is supplied and the process runs on a TTY, prompt:
  `Password for user@host: ` (no echo, via `golang.org/x/term`)
- Prompts **must** be written to `/dev/tty` (falling back to stderr only if no controlling terminal),
  never to stdout, so piped automation is not polluted
- Non-TTY invocations (`PORTUNIX_CI=1` or stdin/stderr not a terminal) **must** fail fast with an
  explicit error when a password would otherwise be prompted — no silent prompt-deadlock
- Interactive prompts must honor the same `SecretString` masking guarantees as other credential
  sources (no echo, cleared after use)

### Key Bootstrap

- `portunix ssh bootstrap-key <user@host> --credential <id>` — authenticate with the password,
  append the specified public key (default: `~/.ssh/id_ed25519.pub`) to the remote
  `~/.ssh/authorized_keys`, verify by reading it back, and print a reminder that password auth
  should now be disabled on the host

### Host-Key Management

- Default host-key policy: `accept-new` (record on first use, reject on change)
- Portunix-local `known_hosts` store at `~/.portunix/ssh/known_hosts`
- `--use-system-known-hosts` flag to additionally consult `~/.ssh/known_hosts`
- `--host-key-check=strict|accept-new|off` flag
- When `off`: emit a warning on every call; refuse to run if `PORTUNIX_CI=1` is set
- `portunix ssh trust <user@host>` — explicit one-time key acceptance command
- `portunix ssh known-hosts list | remove` — management subcommands

### ssh-agent Lifecycle

- `portunix ssh agent start` — start an ssh-agent and print the exported environment
- `portunix ssh agent add <path>` — add identity (prompts for passphrase without echo)
- `portunix ssh agent stop` — stop the agent

### Ansible / Legacy Tool Compatibility

- `portunix ssh sshpass-compat <credential-id>` — emit a short-lived file-descriptor path that a
  consumer (e.g., `ansible-playbook --ask-pass`) can read from; password fetched from
  `ptx-credential` just-in-time and never persisted
- Verified end-to-end against `ansible-playbook` in an integration test

### Replacement of Legacy Code

- Deprecate `LnxExecutePyScriptSsh` in `src/app/ssh.go`
- Rewrite it as a thin shim that invokes the new `ptx-ssh` APIs
- Remove hardcoded key path and `ssh.InsecureIgnoreHostKey()` usage
- Plan removal of the shim in a later release (document in `CHANGELOG.md`)

### Security

- Wrap the credential in a `SecretString` type whose `String()` / `GoString()` / JSON marshaler
  return a masked value (`"***"`). Add a `go vet` / `golangci-lint` custom check that forbids
  passing `SecretString` to `fmt.Printf`-family calls directly.
- Zero password buffers after use where Go's memory model allows.
- Unit tests must verify that:
  - `SecretString.String()` does not leak the value
  - JSON/YAML marshaling of `SecretString` produces the mask
  - Error messages that embed the credential structure do not include the password
- No logging of passwords at any verbosity level.
- All new SSH sessions default to `StrictHostKeyChecking=accept-new` equivalent behaviour.

## Implementation Phases

### Phase 1: Helper Skeleton and Dispatcher Registration

1. Create `src/helpers/ptx-ssh/` with Cobra CLI scaffold (`main.go`, `cmd/root.go`,
   version/description/list-commands)
2. Follow the full checklist in `docs/contributing/HELPER-BINARY-DEVELOPMENT.md` — 11 phases,
   30+ steps covering dispatcher.go, Makefile, .goreleaser.yml, deploy scripts, installation system
3. Register in `src/dispatcher/dispatcher.go` (`Commands: []string{"ssh", "scp"}`)
4. Smoke test: `portunix ssh --version` prints a sensible version string

### Phase 2: Native SSH Client (Key-Based First)

1. Add `connect`, `exec`, `copy` subcommands using `golang.org/x/crypto/ssh` + `github.com/pkg/sftp`
2. Implement Portunix-local `known_hosts` store with `accept-new` policy
3. Add `trust` and `known-hosts` management subcommands
4. Integration tests against an in-test OpenSSH server (use `portunix container run` fixture, per
   the project Container-Based Testing Policy — no direct `docker` / `podman` calls)

### Phase 3: Credential Integration (Password Auth)

1. **Prerequisite — ADR-002**: Extract `ptx-credential` storage into `internal/credential/sdk`.
   If this phase grows beyond a single PR, split it into a dedicated implementation issue as
   suggested in ADR-002's Follow-Up section.
2. Wire `--credential`, `--credential-fd`, `--credential-file`, `--credential-env` flags into auth
   method resolution
3. Implement `SecretString` type with masking and a custom linter check (per ADR-002 — the type is
   defined in the SDK and reused by `ptx-ssh`, not redefined here)
4. Integration tests: password auth against an OpenSSH container with `PasswordAuthentication yes`

### Phase 4: Wrapper Mode and Ansible Compatibility

1. Add `github.com/creack/pty` dependency
2. Implement `--wrapper` mode that execs the system `ssh`/`scp`/`rsync` and feeds the password at
   prompt
3. Implement `portunix ssh sshpass-compat <id>` fd-based password emitter
4. Integration test: `ansible-playbook --ask-pass` against the container fixture

### Phase 5: Key Bootstrap and Agent

1. Implement `bootstrap-key` subcommand
2. Implement `agent start|add|stop` subcommands (wrap `ssh-agent`)
3. Integration test: full password → key migration flow

### Phase 6: Legacy Code Retirement

1. Refactor `src/app/ssh.go` `LnxExecutePyScriptSsh` into a shim over `ptx-ssh`
2. Remove hardcoded paths and `InsecureIgnoreHostKey()`
3. Add deprecation notice to `CHANGELOG.md`

### Phase 7: Documentation

1. Update helper index — add `ptx-ssh` to helpers docs
2. Add `docs/helpers/ptx-ssh.md` with full command reference and examples
3. Update any existing sshpass references with a note pointing to `ptx-ssh` as the in-ecosystem
   replacement

## Acceptance Criteria

- [ ] AC-1: `portunix ssh connect user@host` establishes an interactive SSH session using a private
      key
- [ ] AC-2: `portunix ssh exec --credential legacy-admin user@host "uname -a"` runs a command using
      a password stored in `ptx-credential`, without the password appearing in `ps`, environment,
      or command line
- [ ] AC-3: `portunix virt ssh` and `portunix docker ssh` (and `portunix container ssh`) continue
      to work unchanged
- [ ] AC-4: `ansible-playbook --ask-pass` succeeds against a password-only test host using
      `portunix ssh sshpass-compat`
- [ ] AC-5: `LnxExecutePyScriptSsh` no longer contains hardcoded paths or `InsecureIgnoreHostKey()`
      usage
- [ ] AC-6: Passing `--password <plaintext>` on the command line is rejected with an error linking
      to the recommended credential-source alternatives
- [ ] AC-7: Host-key verification cannot be silently disabled when `PORTUNIX_CI=1` is set
- [ ] AC-8: All subcommands expose `--help`, `--help-ai`, `--help-expert` per Issue #163
- [ ] AC-9: Cross-platform build (Linux, macOS, Windows) succeeds via `make build`; helper binary
      is delivered by `make deploy-local` and the release pipeline
- [ ] AC-10: All new and migrated code passes the existing test suite plus the new integration
      tests (exercised in container fixtures, per the Container-Based Testing Policy)

## Non-Goals

- Server-side SSH provisioning (remains in `Install-PortableOpenSSH.ps1` and `ptx-installer`)
- OpenSSH CA / certificate-based authentication (deferred to a later ADR)
- MCP exposure of SSH primitives (deferred; requires its own approval-loop design)
- Tunneling / port-forwarding UX (may be added later; not in this issue's scope)

## Risks

- **Windows pty differences** — ConPTY requires Windows 10 1809+. Mitigation: document the
  requirement; fall back to native mode on unsupported builds.
- **Ansible version drift** — the fd-based password flow depends on Ansible implementation.
  Mitigation: pin test matrix to supported Ansible versions.
- **`ptx-credential` SDK API stabilisation** — extracting the storage into a shared internal
  package may require iteration. Mitigation: per ADR-002, the package sits in `internal/` with no
  stability promise until three consumers have shipped against it.
- **Dispatcher routing pitfalls** — Issue #172 (Finding #3) showed that edits to the root Cobra
  tree in `src/cmd/` are dead code for dispatched commands. All subcommand definitions and help
  text must live in `src/helpers/ptx-ssh/`.

## References

- Source issue: `portunix-architecture/docs/architecture/components/portunix/issues/001-ptx-ssh-helper.md`
- ADR-001 — ptx-ssh Helper (in `portunix-architecture` repository)
- ADR-002 — Extract `ptx-credential` Storage into Shared Internal Go SDK
  (in `portunix-architecture` repository)
- `docs/contributing/HELPER-BINARY-DEVELOPMENT.md` — mandatory checklist for new `ptx-*` binaries
- `docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md` — Container-Based Testing Policy
- Issue #163 — `--help-ai` / `--help-expert` flag convention for helpers
- Issue #056 — `ptx-ansible` helper (downstream consumer of `--ask-pass` flows)
- Existing insecure code to retire: `src/app/ssh.go` (`LnxExecutePyScriptSsh`)

## Complexity

**High** — new helper binary with its own Cobra tree, native Go SSH client plus wrapper-mode pty
exec, multi-flag credential resolution, SDK extraction from `ptx-credential` (prerequisite),
cross-platform (including Windows ConPTY), integration with Ansible, retirement of legacy
insecure code. Split across 7 implementation phases; Phase 3 (credential SDK) may warrant its
own dedicated sub-issue.
