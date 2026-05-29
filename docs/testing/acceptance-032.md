# Acceptance Protocol — Issue #032

**Issue**: Universal Container Management Commands
**Branch**: `feature/032-universal-container-mgmt`
**Commit under test**: `61c56dd`
**Tester**: claude (role: Tester / generic)
**Date**: 2026-05-16
**Testing OS**:

- Host: Linux 6.17.0-23-generic
- Shell: bash
- Runtime: Podman 5.4.2 (Docker not installed on test host)

**Build**: `make build` clean from feature branch tip (`61c56dd`); main `portunix` binary and all 19 helper binaries rebuilt successfully.

## Scope

Issue #032 requested a complete set of universal container management commands routed through `portunix container <subcommand>`. Audit of the codebase showed that most commands (`stop`, `start`, `rm`, `logs`, `list`, `inspect`, `info`, `cp`, `exec`) were already implemented but had three gaps:

| Gap | Status before PR | Resolved by |
| --- | ---------------- | ----------- |
| 1. `restart` subcommand entirely missing | ❌ Open | **This PR** |
| 2. `remove` alias (issue text) not wired in helper switch — only `rm` accepted | ❌ Open | **This PR** |
| 3. Cross-runtime discovery broken: stop/start/rm/logs always tried Podman first, never falling through to Docker when both runtimes installed and container only in Docker | ❌ Open | **This PR** |

**Out of scope** (intentionally):

- `--all` flag for `list` — mentioned only in issue example, not in acceptance criteria
- Batch operations for `stop`/`start`/`logs` — issue lists as "advanced", `rm` already supports multiple names
- Repair of pre-existing test failures in `TestContainerCommandErrorMessages`, `TestContainerExecCommand`, `TestContainerRunCommand_*` (broken before this branch — separate concern)

## Test Plan

| TC | Area | Description |
| -- | ---- | ----------- |
| TC-01 | Help listing | `portunix container` advertises `restart` and `rm (alias: remove)` |
| TC-02 | `restart --help` | Help banner shows correct usage and 🔄 RESTART CONTAINER section |
| TC-03 | restart missing arg | "Container name required" + usage |
| TC-04 | restart unknown container | "container '<name>' not found in any available runtime" |
| TC-05 | stop unknown container (regression) | Same "not found" error — no raw podman/docker passthrough |
| TC-06 | restart running container | `StartedAt` timestamp changes; ✅ success message |
| TC-07 | stop + start lifecycle | exit → running transition correct |
| TC-08 | `rm` primary | Container disappears from `podman ps -a` |
| TC-09 | `remove` alias with `--force` | Removes running container (alias wires to `handleContainerRm`) |
| TC-10 | logs | Streams container stdout |
| TC-11 | list | Container appears in "🦭 Podman Containers" section |
| TC-12 | Unit tests in scope | 10 `TestContainer*` test sets / 19 sub-tests pass |
| BUILD | `make build` regression | All binaries rebuild cleanly |

## Test Results

### Functional (manual E2E with Podman)

- [x] **TC-01** `portunix container` lists `restart` and `rm (alias: remove)`
- [x] **TC-02** `portunix container restart --help` shows expected banner
- [x] **TC-03** `portunix container restart` → `❌ Error: Container name required`
- [x] **TC-04** `portunix container restart nonexistent-tc04` → `❌ Error: container 'nonexistent-tc04' not found in any available runtime`
- [x] **TC-05** `portunix container stop nonexistent-tc05` → identical "not found" message (regression fix verified)
- [x] **TC-06** Created `ptx-test-032` with podman; `restart` changed `StartedAt` from `2026-05-16 08:18:45` to `2026-05-16 08:19:08`
- [x] **TC-07** `stop` → state=`exited`, `start` → state=`running`
- [x] **TC-08** `rm` removed the container; `podman inspect` reports "no such object"
- [x] **TC-09** `remove` (alias for `rm`) with `--force` removed running container
- [x] **TC-10** `logs` displayed tick counter from container stdout
- [x] **TC-11** `list` showed `ptx-test-032` in Podman section

### Unit Tests

- [x] **TC-12** `go test -vet=off portunix.ai/cmd/... -run "TestContainerCommandHelp|TestContainerCommandRegistration|TestContainerCommandFlags|TestContainerStopCommand|TestContainerStartCommand|TestContainerRemoveCommand|TestContainerLogsCommand|TestContainerListCommand|TestContainerListCommandAliases|TestContainerCommandExamples"` → `ok` (all pass)
- [x] **Bonus regression fix:** This PR's test patch (corrected expected `Use` pattern from `"remove <container-name>"` → `"rm <container-name>"`) made `TestContainerCommandHelp` and `TestContainerCommandRegistration` start passing — they were broken on parent commit `e625472`

### Build

- [x] **BUILD** `make build` produces `portunix` + 19 helpers without errors

### Pre-existing failures (not introduced by this PR — verified on parent `e625472`)

- ❌ `TestContainerCommandErrorMessages` — asserts legacy phrases ("configured container runtime", "Docker or Podman") that don't appear in actual command descriptions
- ❌ `TestContainerExecCommand`, `TestContainerRunCommand_ArgumentParsing/_ErrorHandling/_FlagValidation/_OriginalIssueRepro` — flag-parsing test mismatches in `container run` cobra plumbing

Recommend follow-up issue to repair these — they are blocking `go test ./...` from being green but are unrelated to issue #032 scope.

### Out of scope / not tested on this host

- [ ] **Cross-runtime discovery with both Docker AND Podman installed** — Docker is not available on the test host. Code path was reviewed:
  - `findContainerRuntime()` in `src/helpers/ptx-container/main.go` calls `containerExistsIn("podman", name)` then `containerExistsIn("docker", name)` (each running `<runtime> inspect <name>`); returns the first match.
  - Suggested verification on a system with both runtimes: create `c-test` only via `docker run -d --name c-test alpine sleep 60`, then run `portunix container stop c-test`. Expected: success. Before this PR: would fail because it always tried `podman stop c-test`.

## Final Decision

**STATUS**: PASS (CONDITIONAL)

**Conditions for unconditional approval**:

- Verify cross-runtime discovery on a system with both Docker and Podman installed (scenario described above). This is the central regression-fix aspect of issue #032 and could not be exercised on the Podman-only test host.

**Approval for merge**: YES (conditional on above)
**Date**: 2026-05-16
**Tester signature**: claude (role=tester)
