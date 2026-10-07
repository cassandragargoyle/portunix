# Issue #183: PTX-SpecPM Helper Implementation — `portunix specpm`

- **Type**: Feature
- **Priority**: High
- **Status**: ✅ Implemented (Phase 1)
- **Closed**: 2026-05-02 (Phase 2 tracked under Issue #184)
- **Labels**: enhancement, helper-binary, new-helper, spec-kit-pm, project-management, ai-integration, dispatcher, slash-commands, multi-agent
- **Created**: 2026-05-01
- **Architecture Decision Record**: [ADR-040 — PTX-SpecPM Helper Architecture](../../adr/040-ptx-specpm-helper-architecture.md) (Active, 2026-05-01)
- **Upstream Reference (content)**: [`CassandraGargoyle/spec-kit-pm`](https://github.com/CassandraGargoyle/spec-kit-pm)
  (local checkout: `../spec-kit-pm/`)
- **UX Inspiration (CLI/scaffold pattern)**: [`github/spec-kit`](https://github.com/github/spec-kit)
  — its `specify init` / `.specify/` scaffold / multi-agent integration / slash-command
  bundle pattern is the reference we mirror. **The two specifications differ in domain**
  (see *Domain Distinction* below).

## Description

Implement a new helper binary `ptx-specpm` registered in the Portunix dispatcher as the
top-level command `portunix specpm`. Its primary purpose is to **initialize project-management
specifications** in a working directory using the
[spec-kit-pm](https://github.com/CassandraGargoyle/spec-kit-pm) framework as content, while
adopting the proven UX pattern of [github/spec-kit](https://github.com/github/spec-kit)
(persistent install, `init` / `check` / `version` / `extension` / `preset` / `integration`
subcommands, hidden scaffold directory, agent-glue via slash-command bundles, support for
30+ AI agents).

The helper removes today's manual vendoring step (upstream `README.md` → *Option 3: Manual
setup*) and gives Portunix users a single-command on-ramp into spec-kit-pm with the same
ergonomics they would get from `specify init` for technical specs.

## Domain Distinction — Why Two Specification Tools

| Aspect | `github/spec-kit` (`specify`) | `spec-kit-pm` (`specpm`) — *this issue* |
| ------ | ----------------------------- | --------------------------------------- |
| Specification type | **Technical / code spec** ("Spec-Driven Development") | **Project-management spec** (Charter, roadmap, risks, decisions, KPIs) |
| Primary user | Developer building a feature | PM / lead defining context, scope, governance |
| Workflow | constitution → specify → clarify → plan → analyze → tasks → implement | Initiation gate (Charter) → spec → plan → risks → decisions → review → report |
| Output | Code-level artifacts (`spec.md`, `plan.md`, `tasks.md`, `data-model.md`, `contracts/`) | PM-level artifacts (`charter.md`, `roadmap.md`, `risks.md`, `decisions.md`, `kpi.md`) |
| Scaffold dir | `.specify/` | `.specpm/` (proposed; final name in ADR) |
| Slash-command namespace | `/speckit.*` | `/specpm.*` |
| Lifecycle scope | Per-feature (many per project) | Per-project (one Charter, recurring loop) |
| Relationship | Complementary — a project can use both | `ptx-specpm` is the PM layer; `specify` (or future `ptx-specify`) is the code layer |

**Conclusion**: `ptx-specpm` is **not** a port of `github/spec-kit` — it is a sibling tool
operating one level higher (project governance, not code design), borrowing the well-proven
CLI/scaffold UX pattern.

## Motivation

Today, adopting spec-kit-pm in a project requires the user to manually:

1. Clone or vendor the upstream repository (submodule, shallow clone, or copy of
   `templates/`, `agents/`, `workflows/`).
2. Decide where the artifacts live (`specs/project/<project-id>/…`, `.pm-spec/…`, etc.).
3. Stitch the AI agent prompts into their tool of choice (Claude Code, Copilot, …).
4. Pick the right slash-command bundle once those land upstream.

The upstream `README.md` itself flags this friction: the official `specpm` CLI is on the
roadmap but *not yet shipped*. In the meantime, Portunix users who already have `portunix`
on their PATH and rely on `ptx-*` helpers for everything else (Python, SSH, containers,
installers, …) lack a consistent on-ramp. A first-class `portunix specpm` helper:

- Removes the manual vendoring step for Portunix users.
- Establishes the integration point flagged in the upstream roadmap as
  *"Synapse / Portunix integration"*.
- Provides a stable surface (`portunix specpm …`) that the dispatcher and AI agents
  (`ptx-mcp`, `ptx-prompting`) can target, regardless of how the upstream CLI evolves.
- Aligns with the project's helper-binary architecture (ADR-style consolidation, see
  Issues #051, #073, #097, #100, #174).

## Involved Components

| Component | Role |
| --------- | ---- |
| **`ptx-specpm`** (new) | New helper binary + dispatcher registration |
| **`portunix` (dispatcher)** | Routes `specpm` top-level command to `ptx-specpm` |
| **`spec-kit-pm` upstream** | Source of templates, agent prompts, and workflows (vendored or fetched) |
| **`ptx-installer`** | Possible delivery path for spec-kit-pm asset bundle (TBD in ADR) |
| **`ptx-prompting`** (#073) | Future consumer — render/serve spec-kit-pm agent prompts |
| **`ptx-mcp`** | Future consumer — expose `specpm.*` operations to AI agents |
| **`assets/`** | Possible embedded fallback bundle of spec-kit-pm templates (TBD in ADR) |

## Scope

### Dispatcher Registration

- Register `ptx-specpm` in `src/dispatcher/dispatcher.go` with command `["specpm"]`.
- Implement the standard helper interface flags (per Issue #163):
  `--version`, `--description`, `--list-commands`, `--help`, `--help-ai`, `--help-expert`.
- Verify no conflict with existing top-level commands.

### CLI Surface (mirrors `github/spec-kit` UX pattern)

The CLI surface is modelled directly on `specify` (`github/spec-kit`), substituting
`specpm` content. Commands marked **(P1)** are in scope for Phase 1; **(Pn)** indicates
later phases.

#### Project bootstrap

- **(P1)** `portunix specpm init <PROJECT_NAME>` — bootstrap a new project directory.
- **(P1)** `portunix specpm init .` / `portunix specpm init --here` — initialize in the
  current directory.
- **(P1)** `portunix specpm init . --force` / `--here --force` — merge into a non-empty
  directory (overwrite scaffold files, never user-authored content).
- **(P1)** `--integration <agent>` — target AI agent for slash-command glue
  (`claude`, `copilot`, `cursor`, `gemini`, …). Mirrors `specify init --integration`.
- **(P1)** `--integration-options="--skills"` — emit agent **skills** instead of
  slash-commands where supported (Claude Code, etc.). Mirrors `specify`'s skills mode.
- **(P1)** `--profile <name>` — apply a company/industry profile from `profiles/`
  (e.g., `example-company`).
- **(P1)** `--source git|embedded|<path>` — choose kit origin (see *Kit Source Strategy*
  below). Default behaviour set by ADR.
- **(P1)** `--ignore-agent-tools` — skip detection of installed agent CLIs (analog of
  `specify`'s flag of the same name).
- **(P1)** Idempotent re-runs: refuse to overwrite without `--force`; print a diff/summary.

#### Inspection and tooling

- **(P1)** `portunix specpm check` — verify: which AI-agent CLIs are installed locally,
  which integrations are wired in this project, which profile is active, the bundled
  spec-kit-pm version. Mirrors `specify check`.
- **(P1)** `portunix specpm version` — print helper version **and** bundled spec-kit-pm version.
- **(P2)** `portunix specpm upgrade` — re-vendor the kit at a newer tag; preserve all
  user-authored content under `.specpm/specs/<project-id>/`.

#### Integration registry

- **(P1)** `portunix specpm integration list` — list all supported AI-agent integration
  targets (analog of `specify integration list`).
- **(P4)** Per-integration drivers for at least Claude Code + one other (e.g., Copilot
  or Cursor), establishing the abstraction.

#### Extensions and presets (P3+)

- **(P3)** `portunix specpm extension search` / `extension add <name>` — community-
  contributed add-ons (extra templates, profiles, agent prompts).
- **(P3)** `portunix specpm preset search` / `preset add <name>` — full pre-baked
  project-type bundles (e.g., "consulting engagement", "internal R&D pilot",
  "regulated-industry rollout").

#### Workflow subcommands (P5 — deferred)

The upstream `spec-kit-pm` roadmap defines `/specpm.init`, `/specpm.spec`, `/specpm.plan`,
`/specpm.risks`, `/specpm.decisions`, `/specpm.review`, `/specpm.report`. The user invokes
these **inside the AI agent** after `init` writes the slash-command bundle. Whether
`portunix specpm spec|plan|risks|decisions|review|report` should also exist as native
shell commands (driving the same templates without an AI loop) is an open question for
the ADR — `specify` itself does not provide native equivalents, so the default is "no",
and only the slash-command bundles are scaffolded.

### Scaffolded Project Layout

Modelled on `github/spec-kit`'s `.specify/` (final name to be confirmed in ADR — proposed
`.specpm/`). Content is **PM-domain**, not code-domain.

```text
.specpm/
├── memory/
│   └── governance.md            # Long-lived PM context (charter pointers, decision index)
├── scripts/
│   ├── check-prerequisites.sh   # Used by slash-commands to validate state
│   ├── common.sh
│   ├── new-project.sh           # Helper for /specpm.init
│   └── update-claude-md.sh      # Augments host CLAUDE.md
├── specs/
│   └── <project-id>/
│       ├── project.md           # Charter (output of Initiation gate)
│       ├── spec.md              # Goals, scope, constraints, success criteria
│       ├── plan.md              # Tasks, milestones, owners
│       ├── risks.md             # Risk register
│       ├── decisions.md         # ADR-style decision log
│       ├── kpi.md               # Success metrics
│       └── reports/
│           └── <YYYY-MM-DD>.md  # Stakeholder reports
├── profiles/                    # Vendored from spec-kit-pm `profiles/`
│   └── <profile-name>/
├── extensions/                  # User-installed extensions (P3)
├── presets/                     # User-installed presets (P3)
└── templates/                   # Vendored from spec-kit-pm `templates/`
    ├── project/charter.md
    ├── decisions/
    ├── kpi/
    ├── risks/
    └── roadmap/

CLAUDE.md (root, augmented if --integration claude)
.claude/commands/
├── specpm.init.md
├── specpm.spec.md
├── specpm.plan.md
├── specpm.risks.md
├── specpm.decisions.md
├── specpm.review.md
└── specpm.report.md
```

`specs/<project-id>/` is **per-project, not per-feature** — this is the key structural
difference from `github/spec-kit`'s per-feature `specs/{FEATURE_ID}-{name}/` tree.

### Slash-Command Bundles

Every supported integration receives a bundle of slash-commands under the `/specpm.*`
namespace (mirrors `/speckit.*`):

| Slash command | Purpose |
| ------------- | ------- |
| `/specpm.init` | Run the Initiation gate; produce `project.md` (Charter) |
| `/specpm.spec` | Draft / update the project specification |
| `/specpm.plan` | Generate the actionable plan |
| `/specpm.risks` | Produce the risk register |
| `/specpm.decisions` | Log architecture / approach decisions |
| `/specpm.review` | Reviewer-agent pass over Charter, spec, plan, risks |
| `/specpm.report` | Generate stakeholder progress reports |
| `/specpm.check` | (helper) Surface `portunix specpm check` output to the agent |

For agents that support **skills mode** (`--integration-options="--skills"`), each
slash-command is emitted as a skill manifest instead.

### Multi-Agent Integration Targets

Phase 1 ships one integration (Claude Code). Phase 4 expands to several, with the
target list ultimately driven by what `spec-kit-pm` itself supports and what Portunix
users need. Reasonable initial set, informed by `specify integration list`:

- Claude Code
- GitHub Copilot
- Cursor
- Gemini CLI
- Codex CLI
- *(others added on demand; `integration list` is the source of truth)*

### Kit Source Strategy

The ADR will choose a default and document the trade-off; all three modes are exposed
via `--source`:

| Mode | Behaviour | Trade-off |
| ---- | --------- | --------- |
| `git` | Shallow clone of upstream `spec-kit-pm` at a pinned tag into `.specpm/` | Always fresh; needs network |
| `embedded` | Extract kit bundle baked into the binary via `embed.FS` | Offline / air-gapped; binary-size cost |
| `<path>` | Copy from a local checkout / fork | Developer / forked-profile workflow |

## Implementation Phases

### Phase 1 — Helper Skeleton, Dispatcher Registration, `init` / `check` / `version` / `integration list`

1. **Architect** authors the ADR. Decisions to lock in:
   - Scaffold directory name (`.specpm/` proposed) and final layout
   - Default `--source` mode (`git` vs. `embedded`)
   - Vendoring strategy (shallow clone vs. submodule vs. extracted bundle)
   - Profile-loading precedence (bundled → org → user)
   - Slash-command file format per supported agent
   - Embedded-bundle versioning and update channel
   **Implementation does not start until the ADR lands.**
2. Create `src/helpers/ptx-specpm/` with Cobra CLI scaffold (`main.go`, `cmd/root.go`,
   `version`/`description`/`list-commands` per Issue #163).
3. Follow `docs/contributing/HELPER-BINARY-DEVELOPMENT.md` end-to-end (dispatcher.go,
   Makefile, `.goreleaser.yml`, deploy scripts, release pipeline).
4. Implement `init` (incl. `--here`, `--force`, `--integration`,
   `--integration-options="--skills"`, `--profile`, `--source git|<path>`,
   `--ignore-agent-tools`).
5. Implement `check`, `version`, `integration list`.
6. Ship the **Claude Code** integration only (one driver, end-to-end).
7. Container-based integration tests (per Container-Based Testing Policy — no direct
   `docker`/`podman` calls): `init` in empty dir, `init .` in existing repo,
   idempotent re-run, `--force` merge, slash-command bundle materialisation,
   `check` output.

### Phase 2 — Embedded Bundle, `upgrade`, Air-Gapped Mode

1. Bake a pinned spec-kit-pm bundle into the binary via `embed.FS` (or deliver via
   `ptx-installer` — final channel chosen in the ADR).
2. Implement `--source embedded` and `upgrade` (preserving everything under
   `.specpm/specs/<project-id>/`).
3. Add the air-gapped integration test (network-disabled container fixture).

### Phase 3 — Extensions and Presets

1. Implement `extension search|add|list|remove` (mirror of `specify extension`).
2. Implement `preset search|add|list|remove` (mirror of `specify preset`).
3. Define the extension/preset manifest format (`.specpm/extensions/<name>/manifest.yml`,
   etc.) — locked in the ADR.
4. Integration test: install an extension, verify slash-command additions; install a
   preset, verify it pre-populates the project skeleton.

### Phase 4 — Multi-Agent Drivers

1. Add at least two more drivers from the integration list (e.g., Copilot, Cursor,
   Gemini CLI).
2. Implement skills-mode emission for agents that support it.
3. `integration list` prints true installed-vs-supported state.
4. Per-agent integration tests in container fixtures.

### Phase 5 — Native Workflow Subcommands & MCP Surface (deferred / spin-off issue)

`portunix specpm spec | plan | risks | decisions | review | report` as native
non-AI commands, plus `ptx-mcp` exposure. Tracked as a follow-up issue once the
upstream `spec-kit-pm` workflow templates stabilise.

## Acceptance Criteria

- [ ] AC-1: `portunix specpm init <new-dir>` bootstraps a project directory with the
      `.specpm/` scaffold (memory, scripts, specs, profiles, templates) plus a
      `specs/<project-id>/` placeholder for the Charter.
- [ ] AC-2: `portunix specpm init .` and `portunix specpm init --here` initialize the
      scaffold inside an existing repo without clobbering unrelated files; a second
      run is a no-op unless `--force` is passed; `--force` only overwrites scaffold
      files, never user-authored content under `.specpm/specs/<project-id>/`.
- [ ] AC-3: `portunix specpm init . --integration claude` writes a working Claude Code
      slash-command bundle under `.claude/commands/specpm.*.md` and (where chosen by
      the ADR) augments root `CLAUDE.md`.
- [ ] AC-4: `portunix specpm init . --integration claude --integration-options="--skills"`
      emits Claude Code skills instead of slash-commands.
- [ ] AC-5: `portunix specpm check` reports installed agent CLIs, wired integrations,
      active profile, and bundled kit version. `portunix specpm version` prints helper
      version + bundled spec-kit-pm version. `portunix specpm integration list`
      enumerates all supported targets.
- [ ] AC-6: `--source git` pins to a specific upstream tag (no floating `main` ref);
      `--source embedded` works in a network-disabled container (Phase 2);
      `--source <path>` consumes a local checkout.
- [ ] AC-7: All subcommands expose `--help`, `--help-ai`, `--help-expert` per Issue #163.
- [ ] AC-8: Cross-platform build (Linux, macOS, Windows) succeeds via `make build`;
      helper is delivered by `make deploy-local` and the release pipeline.
- [ ] AC-9: Integration tests pass in container fixtures (Ubuntu, Debian, Alpine) per
      the Container-Based Testing Policy — no direct `docker`/`podman` calls.
- [ ] AC-10: Dispatcher registration does not break any existing top-level command;
      `portunix specpm` is recognised and routed correctly.
- [ ] AC-11: End-to-end smoke test: empty dir → `specpm init --integration claude` →
      `specpm check` → user runs `/specpm.init` in Claude Code → Charter populated →
      `specpm check` reflects populated state.
- [ ] AC-12: Domain distinction is preserved — `specpm init` does **not** scaffold
      per-feature `spec.md`/`plan.md`/`tasks.md` files (those belong to `specify`);
      it scaffolds **per-project** `project.md` (Charter), `risks.md`, `decisions.md`,
      `kpi.md`.

## Non-Goals

- **Re-implementing `github/spec-kit`** — its CLI (`specify`) and content (technical
  code-spec workflow) are out of scope. `ptx-specpm` borrows the *UX pattern only*.
  A separate future `ptx-specify` could wrap `github/spec-kit` itself in the same
  helper-binary style; not part of this issue.
- **Native Go re-implementation of the AI workflow** — `/specpm.spec`, `/specpm.plan`,
  `/specpm.risks`, etc. run inside the host AI agent; the helper only scaffolds the
  bundles. Native shell equivalents are a Phase 5 / spin-off question.
- **Forking spec-kit-pm** — `ptx-specpm` is a consumer/integrator; the upstream
  remains authoritative for templates and agent prompts.
- **MCP surface** — exposing specpm operations to AI agents over MCP is a follow-up
  (`ptx-mcp` collaboration), not in this issue.
- **Server-side or multi-user collaboration features** — local filesystem only.
- **Authoring company-specific profiles** — those live upstream or in user forks.

## Risks

- **Upstream `spec-kit-pm` API churn** — the upstream `specpm` CLI is not yet released;
  its UX may shift. Mitigation: pin the upstream version we vendor; `portunix specpm`
  is the stable surface for Portunix users; reconcile in `upgrade`.
- **Confusion with `github/spec-kit`** — names sound similar, both have `init` /
  `check` / slash-commands. Mitigation: explicit *Domain Distinction* table at the
  top of `docs/helpers/ptx-specpm.md`; `check` output explicitly lists detected
  `specify` installs and clarifies the relationship.
- **Agent-glue drift** — Claude Code / Copilot / Cursor slash-command and skills
  formats evolve. Mitigation: per-agent driver; integration registry; add new
  targets iteratively.
- **Bundle freshness vs. binary size** — embedded bundle grows the helper binary;
  always-git-clone needs network. Mitigation: ADR picks the default; both modes
  exposed via `--source`.
- **Co-installed upstream `specpm`** — once upstream ships its own CLI, users may
  have both on PATH. Mitigation: `portunix specpm check` detects it and prints a
  reconciliation note.
- **Per-project (not per-feature) layout pitfall** — developers used to `specify`'s
  per-feature tree may try to create multiple `<project-id>/` dirs. Mitigation:
  `init` refuses a second project in the same scaffold without `--multi-project`
  (a follow-up flag, not P1).

## References

### Inspiration (UX / CLI / scaffold pattern)

- `https://github.com/github/spec-kit` — `specify` CLI, `.specify/` scaffold,
  `--integration` flag, skills mode, extensions/presets, integration registry,
  `/speckit.*` slash-command bundle pattern. **Mirrored in Phase 1.**

### Content source

- `https://github.com/CassandraGargoyle/spec-kit-pm`
  (local checkout: `../spec-kit-pm/`)
- Upstream README — *Get Started*, *Roadmap*, *Profiles*, *Upstream & Fork Model*

### Portunix internal

- `docs/contributing/HELPER-BINARY-DEVELOPMENT.md` — mandatory checklist for new
  `ptx-*` binaries
- `docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md` — Container-Based Testing Policy
- Issue #051 — Git-like Dispatcher with Python Distribution Architecture
- Issue #073 — PTX-Prompting Helper Implementation (future prompt-template collaborator)
- Issue #097 — PTX-Python Helper Implementation (helper structure precedent)
- Issue #100 — PTX-Installer Helper Implementation (possible kit-delivery channel)
- Issue #163 — `--help-ai` / `--help-expert` flag convention for helpers
- Issue #174 — `ptx-ssh` Helper (recent helper-binary precedent + dispatcher pattern)

## Complexity

**High** — new helper binary with its own Cobra tree, dispatcher registration,
external-asset vendoring strategy (git + embedded), multi-agent driver abstraction
(initial Claude Code, later Copilot/Cursor/Gemini/…), extensions/presets registry,
cross-platform build and release-pipeline integration, ADR on scaffold layout and
slash-command emission format. Phase 1 alone is self-contained and shippable;
Phases 3–4 broaden the surface; Phase 5 is intentionally deferred.
