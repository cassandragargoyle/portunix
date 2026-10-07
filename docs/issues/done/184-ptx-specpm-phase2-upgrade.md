# Issue #184: PTX-SpecPM Phase 2 — `upgrade` Command + Air-Gapped Workflow

- **Type**: Feature
- **Priority**: Medium
- **Status**: ✅ Implemented
- **Closed**: 2026-05-02 (acceptance protocol: docs/testing/internal/acceptance-184.md)
- **Labels**: enhancement, helper-binary, ptx-specpm, spec-kit-pm, project-management, ai-integration, dispatcher, air-gapped, integration-test
- **Created**: 2026-05-02
- **Architecture Decision Record**: [ADR-041 — PTX-SpecPM Phase 2](../../adr/041-ptx-specpm-phase2-upgrade-and-airgapped.md) (Active, 2026-05-02)
- **Builds on**:
  - [ADR-040 — PTX-SpecPM Helper Architecture](../../adr/040-ptx-specpm-helper-architecture.md) (Active, 2026-05-01)
  - Issue #183 — Phase 1 (✅ Implemented)
- **Phase**: 2 of 5 (per Issue #183 §"Implementation Phases")

## Description

Phase 2 of the `ptx-specpm` helper. Phase 1 (#183) shipped the bootstrap surface
(`init`, `check`, `version`, `integration list`) with kit content fetched from
upstream `spec-kit-pm` at a pinned ref. Phase 2 adds the **kit-refresh path**
(`upgrade` command), the **published air-gapped workflow** (mirror once,
init/upgrade many), and a **container-based integration test** that exercises
the air-gapped invariant (no network traffic).

ADR-040 §"Implementation Phasing Note (revised 2026-05-01)" rejected the
embedded-bundle work item that was originally part of Issue #183's Phase 2; the
remaining scope is what this issue tracks. ADR-041 locks the residual UX
decisions (`upgrade` semantics, drift policy, `.specpm/` overwrite policy,
test fixture).

## Motivation

- **Kit drift is real**: upstream `spec-kit-pm` is in the "early stage" period
  (per its own `README.md`); template / agent-prompt / workflow text will
  evolve. Without `upgrade`, users have to delete `.specpm/` and re-run `init`
  to pick up changes — risky next to live `specs/project/<id>/` content.
- **Air-gapped operators need a documented path**: ADR-040 D2/D3 already
  established `--source <path>` as the supported air-gapped mode, but there is
  no published "how do I actually use this" workflow. CI runners, regulated
  environments, and disconnected demos all need that recipe.
- **Test coverage gap**: Phase 1's hermetic Go test runs on the host. No test
  currently exercises the air-gapped path end-to-end inside a container, which
  is what the Container-Based Testing Policy expects for installation-shaped
  workflows.

## Involved Components

| Component | Role |
| --------- | ---- |
| **`ptx-specpm`** | New `upgrade` Cobra subcommand; `check` gains a drift line; error messages updated when offline-without-cache |
| **`internal/kit`** | `GitProvider.Fetch` already cache-aware; `upgrade` re-uses it; no new provider |
| **`internal/scaffold`** | `Write(...)` extended with carve-out semantics for `kit.json`, `extensions/`, `presets/` (per ADR-041 D3) |
| **`internal/kitjson`** | Manifest write path called on every `upgrade`, even no-op (`written_at` bumps) |
| **`docs/helpers/ptx-specpm.md`** | New "Air-gapped workflow" section (per ADR-041 D4) |
| **`test/integration/`** | New container-based test using `portunix container run-in-container ... --image alpine:latest` |

## Scope

### CLI Surface — `upgrade` (locked by ADR-041 D1)

| Invocation | Behaviour |
| ---------- | --------- |
| `portunix specpm upgrade` | Refresh `.specpm/` from the project's currently-pinned ref (read from `.specpm/kit.json`, fallback `DefaultPinnedRef`). Idempotent on cache hit + unchanged content. |
| `portunix specpm upgrade --ref <X>` | Bump project pin to `<X>`; fetch `<X>`; refresh. Implies `--source git` unless `--source <path>` co-passed. |
| `portunix specpm upgrade --to-default` | Shortcut for `--ref <DefaultPinnedRef>`. |
| `portunix specpm upgrade --source <path>` | Switch project to local-path provider; refresh from `<path>`. Recorded in `kit.json.source`. |

**No `--force` flag** (per ADR-041 D3 — `.specpm/` is the kit cache and is always authoritatively refreshed).

### CLI Surface — `check` (drift line, per ADR-041 D2)

When `.specpm/kit.json` exists and `kit.json.ref != DefaultPinnedRef`, append:

```text
Drift            : binary recommends <DefaultPinnedRef> (this Portunix release)
                   Run 'portunix specpm upgrade --to-default' to follow the recommendation.
```

When refs match, no drift line; no `check` exit-code change.

### Overwrite Policy (locked by ADR-041 D3)

`.specpm/` paths refreshed on `upgrade`:
`templates/`, `agents/`, `workflows/`, `profiles/<bundled>/`, `memory/`, `scripts/`.

`.specpm/` paths preserved on `upgrade`:
`kit.json` (rewritten in place with new `written_at`/`ref`/`source`),
`extensions/` (Phase 3+ user-installed), `presets/` (Phase 3+ user-installed).

`specs/project/<id>/` paths: **never touched**.

### Air-Gapped Workflow (locked by ADR-041 D4)

Add a new section to `docs/helpers/ptx-specpm.md` documenting:

1. Mirror upstream once on a connected host (`git clone --depth 1 --branch <ref> ... && tar czf ...`).
2. Transfer tarball to the air-gapped host; extract to a stable location.
3. Use `init --source <path>` and `upgrade --source <path>` against the mirror.

### Error Hint (locked by ADR-041 D4)

When `--source git` is used and (a) `git` is missing OR (b) cache miss with no
network, the helper exits non-zero with the verbatim hint from ADR-041 D4
(don't paraphrase — the text is part of the locked surface).

### Integration Test (locked by ADR-041 D5)

One new test under `test/integration/`, container-based:

- Runner: `portunix container run-in-container ... --image alpine:latest`
  (NOT raw `docker`/`podman`).
- Network isolation: strongest available (`--network none` if supported by the
  underlying runtime; otherwise rely on no upstream-reachable infrastructure
  inside the container). Test must fail closed if network is silently
  reachable.
- Fixture: in-test kit tree (extract / share `buildKitFixture` pattern from
  `test/integration/ptx_specpm_test.go`).
- Scenario:
  1. `portunix specpm init --here --integration claude --source /tmp/test-kit`
  2. Plant sentinel at `specs/project/<id>/USER.md`.
  3. `portunix specpm upgrade --source /tmp/test-kit` — assert kit refreshed,
     sentinel preserved.
  4. `portunix specpm check` — asserts local-path source.
  5. Assert no `git` binary was invoked (e.g. `git` not on `PATH` inside the
     container).

The Phase 1 host test (`TestIssue183_PtxSpecpm_Phase1`) is **not modified**.

## Acceptance Criteria

- [ ] AC-1: `portunix specpm upgrade` (no flags) re-fetches the project's
      `kit.json.ref` and refreshes `.specpm/`. On cache hit + identical
      content the run is idempotent (only `written_at` changes in `kit.json`);
      no error, no drift hint in the summary, exit 0.
- [ ] AC-2: `portunix specpm upgrade --ref <X>` sets `kit.json.ref = X`,
      fetches X (cache or upstream), refreshes `.specpm/`. `--ref` works for
      tags **and** commit SHAs. `<X>` may not be `main` *unless* the user
      passes `--ref main` explicitly (no floating refs in any default code
      path — already enforced for `init`).
- [ ] AC-3: `portunix specpm upgrade --to-default` is functionally identical
      to `upgrade --ref <DefaultPinnedRef>`; the `<DefaultPinnedRef>` value
      is the constant in `internal/kit/pinned.go`.
- [ ] AC-4: `portunix specpm upgrade --source <path>` switches the project to
      the local-path provider; `kit.json.source = <path>`,
      `kit.json.ref = "(local path)"`; subsequent plain `upgrade` runs reuse
      that source until another `upgrade --source <…>` or `upgrade --ref <X>`
      is issued.
- [ ] AC-5: `upgrade` overwrites every kit-cache file (templates/, agents/,
      workflows/, profiles/<bundled>/, memory/, scripts/) and **preserves**
      `kit.json`, `extensions/`, `presets/`. **`specs/project/<id>/` is never
      touched** — verified by planting a sentinel and asserting it survives.
- [ ] AC-6: `portunix specpm check` adds the drift line **only** when
      `kit.json.ref != DefaultPinnedRef`. The line text matches ADR-041 D2
      verbatim. `check` exit code is unchanged by drift presence.
- [ ] AC-7: `--source git` with missing `git` binary OR cache miss + no
      network exits non-zero with the verbatim error hint from ADR-041 D4.
- [ ] AC-8: New section "Air-gapped workflow" lands in `docs/helpers/ptx-specpm.md`
      with the three-step recipe (mirror / transfer / init+upgrade --source <path>).
- [ ] AC-9: Container-based integration test passes against
      `alpine:latest` via `portunix container run-in-container`. Test asserts:
      (a) `git` is not on `PATH` inside the container, (b) all five scenario
      steps succeed, (c) sentinel under `specs/project/<id>/` survives.
- [ ] AC-10: All Phase 1 helper meta-flags (`--version`, `--description`,
      `--list-commands`, `--help`, `--help-ai`, `--help-expert`) still work
      and `upgrade` appears in `--help`, `--help-expert`, and the `--help-ai`
      JSON `commands[]` list (Issue #163 contract).
- [ ] AC-11: Cross-platform build via `make build` succeeds; `upgrade` is
      reachable both directly (`./ptx-specpm upgrade …`) and via the
      dispatcher (`portunix specpm upgrade …`).
- [ ] AC-12: Hermetic Phase 1 test (`TestIssue183_PtxSpecpm_Phase1`) is
      untouched and continues to pass.

## Non-Goals

- **Embedded bundle / `embed.FS`** — explicitly rejected by ADR-040 D2/D3
  (project-owner directive 2026-05-01). No `internal/kitfs/` package;
  no `make specpm-vendor` target.
- **Auto-bump on `init`** — `init` is bootstrap-only; pin movement is always
  via `upgrade --ref` or `upgrade --to-default`.
- **Three-way merge for `.specpm/` user edits** — by design (ADR-041 D3 and
  trade-off Option A). Customisation lives in profiles / extensions / presets,
  not in hand-edits to the kit cache.
- **`upgrade` history field in `kit.json`** — only most-recent source/ref are
  recorded. Out of scope for Phase 2; called out as a risk mitigation in
  ADR-041 if the audit case ever lands.
- **Multi-project workspace upgrade** — `upgrade` operates on the current
  directory's project. Bulk operations are not in scope.
- **Phase 3 (extensions, presets) and Phase 4 (multi-agent drivers)** —
  separate issues, separate ADRs (TBD).
- **Phase 5 (native workflow subcommands, MCP surface)** — explicitly
  deferred per Issue #183 §"Phase 5 — deferred / spin-off issue".

## Risks

- **Cache poisoning of `~/.cache/portunix/specpm/<ref>/`** — if an attacker
  controls the cache directory, `upgrade` reads tampered kit content.
  Mitigation: cache lives under user-private `~/.cache/`; helper does not
  follow symlinks out of the cache root; pinned ref keeps the on-disk path
  stable so external tampering is detectable by re-cloning. Not a Phase 2
  blocker; tracked as a follow-up if a security audit calls for it.
- **`alpine:latest` ships busybox `tar` / minimal POSIX tools** — the
  in-container fixture must use only POSIX-portable shell. Mitigation: keep
  the fixture builder in Go (TestFramework `Command` calls), not in shell.
- **Container runtime auto-resolution** (`portunix container` chooses
  Podman or Docker) — the `--network none` flag must be supported by the
  resolved runtime. Mitigation: assert availability in test setup; fall back
  to "no `git` on PATH" assertion if `--network none` is unavailable.
- **Drift line wording change** — copywriting the drift hint feels harmless
  but the text is part of the locked ADR surface. Mitigation: any future
  change must update ADR-041 D2 first.

## References

### Internal

- ADR-040 — PTX-SpecPM Helper Architecture (the Phase 1 / global ADR;
  D1/D2/D3/D6 inherited verbatim).
- ADR-041 — PTX-SpecPM Phase 2 (this issue's authoritative architecture).
- Issue #183 — Phase 1 implementation tracker (✅ Implemented).
- `docs/contributing/HELPER-BINARY-DEVELOPMENT.md` — mandatory checklist for
  helper-binary changes.
- `docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md` §"Container-Based
  Testing Policy".
- Issue #163 — `--help-ai` / `--help-expert` flag convention (still applies).

### Upstream

- `https://github.com/CassandraGargoyle/spec-kit-pm` — kit content source.

## Complexity

**Medium** — one new Cobra subcommand reusing the Phase 1 provider/scaffold
plumbing, modest changes to `check`, two new public flags
(`--ref`, `--to-default`, `--source` on `upgrade`), one carve-out tweak to the
ScaffoldWriter, one documentation section, and one container-based integration
test. No new external dependency, no new ADR-grade design surface beyond what
ADR-041 already locks. Implementation is intentionally smaller than Phase 1.
