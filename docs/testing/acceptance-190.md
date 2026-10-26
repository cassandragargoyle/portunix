# Acceptance Protocol - Issue #190 (Skill Installation Support — `portunix skills`)

**Issue**: Skill Installation Support — `portunix skills`
**Branch**: `feature/190-skills-install` (commit `a44d2fe`, retest 1 on `8f067d2`, retest 2 on `a80826f`, retest 3 on `bd732c7`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-26
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-38, amd64 (host) — build, static checks, unit tests; Go 1.26.8
- Ubuntu 24.04.3 LTS (`ubuntu:24.04`) container in Podman, started with
  `portunix container run`; `git` 2.43.0, `python3` 3.12.3 installed inside the container —
  manual acceptance scenarios

No software was installed and no host configuration was changed. Skills were installed only
inside the container (project dirs under `/work`, user scope under `/root`). The container
`acc190` was removed after the run.

## Scope Note

This protocol covers **Phase 1** of the issue as agreed for this branch: the `skills` verb in
`ptx-installer`, git and local sources with cache, SKILL.md discovery, universal and `claude`
targets, project and user scope, provenance manifest, trust gate and conflict handling, plus
the `docs/helpers/ptx-installer.md` skills section.

**Independence note**: the code and the unit tests were written by Claude Code in the developer
role in the same session. The 22 manual scenarios below were designed separately in the tester
role.

**Not covered**:

- Phase 2 (Cursor, Gemini CLI, GitHub Copilot, Zed, OpenCode, Amp drivers, interactive agent
  multi-select, extended `sources`) and Phase 3 (shell completion) — not implemented yet
- Interactive TTY paths (numbered skill selection, trust and overwrite prompts) — covered by
  unit tests with a scripted prompter only
- Windows — the helper cross-compiles and vets with `GOOS=windows`, nothing was executed
- A Go integration test in `test/integration/` — scenarios were run with an ad-hoc shell script

## Test Summary

- Initial run (`a44d2fe`): 22 manual + 13 unit tests, all passed
- Retest 1 (`8f067d2`): 22 regression + 16 new manual scenarios, 15 integration cases,
  26 unit tests, all passed (one scenario re-run after a test-input fix, see Retest 1)
- Retest 2 (`a80826f`): 15 new O4 scenarios, 22 + 16 regression scenarios, 16 integration
  cases, 28 unit tests, all passed (three scenarios re-verified after test-design fixes, see
  Retest 2)
- Failed: 0
- Skipped: 0
- Defects found: 0
- Retest 3 (`bd732c7`): 7 new O5 scenarios, 16 trust regression scenarios, 16 integration
  cases, unit tests, all passed (T11/T11b superseded, see Retest 3)
- Observations: 5 (O1–O5 resolved)

## Test Plan

| Level | What | Where |
| ----- | ----- | ----- |
| Static | `make build`, `gofmt`, `go vet` (also `GOOS=windows`) | host |
| Unit | `engine/skills` tests (`go test -race -count=1`), existing `ptx-installer` tests | host |
| Manual | local `--source` fixture, real git source `obra/superpowers`, error paths | `ubuntu:24.04` container |

## Test Cases

| ID | Given / When / Then |
| ----- | ------------------- |
| TC01 | Given the built binaries, when `portunix skills --help` runs, then the help of `ptx-installer skills` is shown |
| TC02 | Given the new verb, when `package info git` and `install git --dry-run` run, then they behave as before |
| TC03 | Given no TTY, when `add` runs without `--skill`/`--all`, then it fails listing the available skills |
| TC04 | Given a new source and no TTY, when `add` runs without `--yes`, then it refuses and writes nothing |
| TC05 | Given `--agent claude --yes`, when a skill is added, then `.agents/skills/` and `.claude/skills/` get the files and manifest; symlinks are skipped |
| TC06 | Given a skill named `../escape`, when discovered, then it is skipped with a warning |
| TC07 | Given an installed skill, when added again without `--yes`, then both targets report `unchanged` and no prompt appears |
| TC08 | Given `--skill gamma`, when it does not exist, then an error lists the available names |
| TC09 | Given installed skills, when `list` and `list --json` run, then the table and valid JSON are printed |
| TC10 | Given the targets, when `sources --json` runs, then the output is clean JSON (no registry log lines) |
| TC11 | Given a changed source and a locally edited claude copy, when `update` runs, then universal is updated, claude is skipped and rc=1 |
| TC12 | Given TC11, when `update --force` runs, then the local edit is overwritten |
| TC13 | Given a hand-placed skill, when `list` and `remove` run, then it is `unmanaged` and removal is refused |
| TC14 | Given a foreign file inside a managed skill, when `remove` runs, then only manifest files are deleted and the dir is kept |
| TC15 | Given `--global --no-universal --agent claude --all`, when added, then skills go to `~/.claude/skills/` only |
| TC16 | Given `https://github.com/obra/superpowers`, when added, then it is cloned and the 40-char commit and normalized `.git` URL are recorded |
| TC17 | Given the cached clone and network blocked by a dead proxy, when `owner/repo` shorthand is added, then it is served from cache |
| TC18 | Given `--ref <sha>`, when added, then the commit is fetched into its own cache dir |
| TC19 | Given a non-existent repository, when added, then a readable clone error with a `--source` hint is shown |
| TC20 | Given `git` removed from `PATH`, when a git source is added, then the error suggests `portunix install git` / `--source` |
| TC21 | Given a read-only `.agents` dir and a non-root user, when added, then `permission denied` is reported and rc≠0 |
| TC22 | Given no managed skill `gamma`, when `update gamma` runs, then an error states it is not managed |

## Test Results

| ID | Result | Evidence |
| ----- | ------ | -------- |
| TC01–TC22 | PASS | `PASS=22 FAIL=0` from the container script |
| TC02 | PASS | Re-verified manually because the scripted grep also matched the echoed command: package metadata printed, `install git --dry-run` completed |
| Unit | PASS | `go test -race -count=1 ./...` in `src/helpers/ptx-installer`: all 4 packages `ok` |
| Static | PASS | `make build`, `gofmt -l` empty, `go vet` (linux, windows) clean |

Selected output (TC17, TC20):

```text
◇  Source: https://github.com/obra/superpowers.git
◇  Repository served from cache
◇  Found 15 skill(s)
◇  Selected 1 skill(s): writing-plans

❌ installing skills from a git URL requires a local 'git' binary, which was not found on PATH.
       Either install it with 'portunix install git', or pass a local checkout path / --source <path>
```

## Observations

### O1 — Write errors are labelled as conflicts

In TC21 the failed write is shown as `skipped (conflict)` followed by the `permission denied`
reason. The exit code and message are correct, only the label is misleading. Recommendation: a
separate `failed` action for write errors. Non-blocking.

### O2 — Interactive paths not exercised end-to-end

Trust prompt, overwrite prompt and numbered skill selection are covered only by unit tests.
Recommendation: one manual TTY run before closing the issue. Non-blocking.

### O3 — Issue remains open after merge

Phase 2 and Phase 3 items, Windows verification and the container integration test in
`test/integration/` are still open. The issue should not be archived as `Implemented` after
this merge.

## Retest 1 — Commit `8f067d2`

Commit `8f067d2` adds the Phase 2 agent drivers (amp, copilot, cursor, gemini, opencode, zed),
detection by registry package or agent config directory, the `skills` CLI aliases,
`XDG_CONFIG_HOME` / `CLAUDE_CONFIG_DIR`, interactive agent selection, directory grouping of
targets, the `failed` action (O1), shell completion (Phase 3) and the Go integration test
`test/integration/issue_190_skills_test.go` (O2, O3).

**Testing OS**: Ubuntu 26.04 host (static checks, unit and integration tests); fresh
`ubuntu:24.04` container `acc190r` in Podman via `portunix container` (manual scenarios),
removed after the run.

### Static Checks and Unit Tests

| Check | Result |
| ----- | ------ |
| `make build` | PASS |
| `gofmt -l` on changed files | PASS (empty) |
| `go vet ./...` (`GOOS=windows`) in `ptx-installer` | PASS |
| `go test -race -count=1 ./...` in `ptx-installer` | PASS, 4 packages |
| Unit tests in `engine/skills` and `ptx-installer` main | PASS, 26 tests |

### Integration Test

`go test -tags integration ./test/integration/ -run TestIssue190 -v`: 15/15 cases PASS in
50.7 s (TC01–TC15, including TTY prompts via `script` and a GitHub clone with cache reuse). The
test container was removed afterwards.

### Regression

The 22 scenarios of the initial run were repeated on `8f067d2`: 22/22 PASS. Two assertions were
adapted to the new output, not to new behaviour: `sources --json` now lists eight targets, and
the universal row of a project carries the label of every agent sharing `.agents/skills`.

### New Scenarios

| ID | Given / When / Then | Result |
| ----- | ------------------- | ------ |
| R01 | Clean system → only universal detected; after creating `~/.gemini`, `~/.config/amp`, `~/.config/zed` → those agents detected | PASS |
| R02 | Detected agents in a project → one write to `.agents/skills`, label `universal,amp,gemini,zed` | PASS |
| R03 | `--global --all-agents` → 7 directories written, universal and zed share `~/.agents/skills` | PASS |
| R04 | `list --global --json` → 7 rows, `targets` array `["universal","zed"]` present | PASS |
| R05 | `remove --global --agent zed` → shared `~/.agents/skills/two` removed, cursor and claude copies kept | PASS |
| R06 | `--agent claude-code,gemini-cli,github-copilot --no-universal` → resolved to claude, gemini, copilot | PASS |
| R07 | `--agent windsurf` → error lists the supported IDs | PASS |
| R08 | `XDG_CONFIG_HOME=/r/xdg CLAUDE_CONFIG_DIR=/r/claude --global` → files under both overrides | PASS |
| R09 | PTY, agent prompt shows detected agents marked `[x]`; `none` → universal only | PASS |
| R10 | PTY, Enter keeps the detected agents; `2` selects claude | PASS (re-run) |
| R11 | PTY, selection `9` → `invalid selection`, nothing written | PASS |
| R12 | PTY with `--yes` → no agent prompt, detected agents used | PASS |
| R13 | Read-only `.agents` as `nobody` → `failed`, no `skipped (conflict)`, rc≠0 (O1) | PASS |
| R14 | `__complete skills remove ""` → installed names; `--agent ""` → universal and agent IDs; `update --global ""` → user-scope names | PASS |
| R15 | `portunix` without `ptx-installer` next to it → clear "requires the ptx-installer helper" error | PASS |
| R16 | `update` after a source change → shared directory updated once, rc=0 | PASS |

R10 failed on the first attempt because the test input did not answer the trust prompt: the
source had been recorded only in another project directory, so it was new for this project.
The behaviour matches ADR-043 D6 (see O4). With the prompt answered, R10 passed.

### Observations of Retest 1

- **O1** resolved — write errors show `failed` (R13, integration TC10)
- **O2** resolved — trust, overwrite, skill and agent prompts exercised on a PTY (R09–R12,
  integration TC11–TC13)
- **O3** partly resolved — Phase 2, Phase 3 and the integration test are done; Windows
  verification is still open, the issue must stay open after the merge
- **O4 — Trust is scoped to the current project and the user scope**: a source installed in
  project A is "new" again in project B. This is consistent with D6 (manifests are the record)
  and safe, but users installing the same source into many projects get the prompt each time.
  Non-blocking; a product decision whether to keep it.

## Retest 2 — Commit `a80826f`

Commit `a80826f` resolves O4: a confirmed (or `--yes`) install records the source in
`<user-config>/portunix/skills-trusted-sources.json` (`PORTUNIX_SKILLS_TRUST` override), the
trust gate consults that record before the manifests, and `portunix skills trust list|remove`
shows and revokes trust. It also adds the `engine/skills` README and integration case TC16.

**Testing OS**: Ubuntu 26.04 host (static checks, unit and integration tests); fresh
`ubuntu:24.04` container `acc190r2` in Podman via `portunix container` (manual scenarios),
removed after the run.

### Static Checks, Unit and Integration Tests

| Check | Result |
| ----- | ------ |
| `make build`, `gofmt -l`, `go vet` (`GOOS=windows`) | PASS |
| `go test -race -count=1 ./...` in `ptx-installer` | PASS, 28 unit tests |
| `TestIssue190_SkillsInContainer` | PASS, 16/16 cases (incl. TC16 for O4) |

### Trust Scenarios

| ID | Given / When / Then | Result |
| ----- | ------------------- | ------ |
| T01 | Clean system → `trust list` says no trusted sources and names the trust file | PASS |
| T02 | `add --yes` → source recorded; file mode 0600, directory 0700 | PASS |
| T03 | Trusted source, another project, no TTY, no `--yes` → installs without the gate | PASS |
| T04 | `trust list --json` → valid JSON with `source` and UTC `trustedAt` | PASS |
| T05 | PTY, prompt declined → source not recorded, nothing written | PASS |
| T06 | PTY, prompt confirmed → source recorded | PASS |
| T07 | Install fails (read-only target, user `nobody`) → source not recorded, rc≠0 | PASS |
| T08 | `trust remove` → fresh project hits the gate again; the project whose manifest records the source still installs | PASS |
| T09 | `trust remove` of an unknown source → error, trust file byte-identical | PASS |
| T10 | Git source added as `https://github.com/obra/superpowers/` → stored as `…/superpowers.git`; revoked with `obra/superpowers` | PASS |
| T11 | Corrupt trust file, fresh project, no `--yes` → `parse <file>` error, rc=1, nothing written; `trust list` reports the parse error | PASS (re-verified) |
| T11b | Corrupt trust file with `--yes` → install proceeds, warning "could not record the trusted source" | PASS (re-verified, see below) |
| T12 | `PORTUNIX_SKILLS_TRUST=/t/custom/trust.json` → record written there, default file untouched | PASS |
| T13 | `trust frobnicate`, `trust remove` without argument → usage errors | PASS |
| T14 | `__complete skills trust ""` → `list`, `remove` | PASS |
| T15 | Same source installed with `--yes` in two projects → one trust entry | PASS |

T11 and T11b were first scripted with wrong expectations: `--yes` skips the gate, so the
corrupt file is only read when recording trust and yields a warning instead of an error. The
follow-up command then ran in a project whose manifest already knew the source. Re-verified in a
fresh project: without `--yes` the gate fails closed with the parse error and writes nothing;
with `--yes` the install proceeds with a warning. Both behaviours are intended.

### Regression

- Initial 22 scenarios: 22/22 PASS
- Retest 1 scenarios R01–R16: 15/16 on the first run. R10 checked a stale label
  `universal,amp,gemini,zed`, but R03 had meanwhile created the config directories of every
  agent, so Enter correctly kept all seven detected agents. R10 also confirms O4: the second
  project no longer asked for trust in `/r/fixB` after the first project confirmed it.
  Re-verified, PASS.

### Observations of Retest 2

- **O4** resolved — trust is recorded user-wide and revocable (T02–T10, TC16)
- **O5 — Corrupt trust file has no recovery hint** (resolved in `bd732c7`, see Retest 3): a broken
  `skills-trusted-sources.json` blocks installs from new sources without `--yes` (fail closed,
  correct) and makes `trust list` fail, but the message does not suggest deleting or fixing the
  file. Non-blocking.

## Retest 3 — Commit `bd732c7`

Commit `bd732c7` resolves O5: when `skills-trusted-sources.json` cannot be parsed, the error
names the file and says to fix the JSON or delete the file to reset the record.

**Testing OS**: Ubuntu 26.04 host (static checks, unit and integration tests); fresh
`ubuntu:24.04` container `acc190r3` in Podman via `portunix container` (manual scenarios),
removed after the run.

### Static Checks, Unit and Integration Tests

| Check | Result |
| ----- | ------ |
| `make build`, `gofmt -l`, `go vet` (`GOOS=windows`) | PASS |
| `go test -race -count=1 ./...` in `ptx-installer` | PASS (the corrupt-file unit test asserts the hint) |
| `TestIssue190_SkillsInContainer` | PASS, 16/16 cases |

### Recovery Scenarios

| ID | Given / When / Then | Result |
| ----- | ------------------- | ------ |
| U01 | Corrupt file → `trust list` names the file, shows the hint, rc=1 | PASS |
| U02 | Corrupt file, new source, no `--yes` → hint, rc=1, nothing written | PASS |
| U03 | Corrupt file, `add --yes` → installed, warning "could not record the trusted source" includes the hint | PASS |
| U04 | Empty (0 byte) file → treated as corrupt, hint shown | PASS |
| U05 | File deleted → `trust list` empty; a new source hits the gate again | PASS |
| U06 | Valid JSON restored → previously trusted source installs without `--yes` | PASS |
| U07 | `{"version": 99}` → "unsupported version 99", no corrupt-file hint | PASS |

### Regression

Trust scenarios T01–T15 from retest 2: 14/16 checks passed on the scripted run. The two
failing checks are T11 and T11b with the scripted expectations that retest 2 already identified
as wrong (`--yes` skips the gate; the follow-up ran in a project whose manifest knew the
source). Their intended behaviour is covered by U01–U03, which passed. No product regression.

## Acceptance Criteria

| Criterion | Status |
| --------- | ------ |
| `add <repo> --skill <name>` installs into `.agents/skills/` and the claude dir | PASS (TC05, TC16) |
| `--all`; no-TTY without `--skill`/`--all` errors | PASS (TC03, TC15); TTY multi-select unit-tested only |
| `--source <path>` works without network | PASS (TC05) |
| Second `add` of the same repo@ref served from cache | PASS (TC17) |
| `--global`/`--user` installs under home; project is default | PASS (TC05, TC15) |
| `list` table + `--json`; `remove` only manifest files; `update` re-resolves | PASS (TC09, TC11, TC14) |
| Trust gate before first write; `--yes` bypasses | PASS (TC04, TC05) |
| Idempotent re-install; content change requires `--force` | PASS (TC07, TC11, TC12) |
| `skills` routed via dispatcher; `install`/`package` unchanged | PASS (TC01, TC02) |
| Default agents from Issue #035 detection | PASS (unit `TestResolveTargets`, `sources` output) |
| Works on Windows and Linux | Linux PASS; Windows not executed (open item on the issue) |
| `git`-missing fails fast with remediation | PASS (TC20) |
| User-friendly errors: invalid repo, unknown skill, missing git, unwritable target | PASS (TC19, TC08, TC20, TC21) |

## Final Decision

**STATUS**: PASS

Phase 1 met its acceptance criteria on Linux in the initial run. Retest 1 confirms Phase 2,
Phase 3 and the fixes for O1 and O2; retest 2 confirms the user-wide trust record for O4;
retest 3 confirms the recovery hint for O5. No regression was found. Windows verification
remains open on the issue (O3) and does not block merging this branch. The issue must not be
archived as `Implemented` until Windows is verified.

**Approval for merge**: YES (commit `bd732c7`)
**Date**: 2026-10-26
**Tester signature**: Claude Code (tester role, Linux)
