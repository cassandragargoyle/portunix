# Issue #196: PTX-PFT `pft ideas` — Opportunity Management (Discovery / Backlog / Opportunity) — Superseded by #198

> **Superseded by [#198](198-ptx-pft-ideas-venture-model-v2.md)** (2026-07-10).
> The single-tier Discovery/Opportunity model (`.discovery`) defined here was
> never implemented; it is replaced by the v2 two-spine Venture model
> (`.venture`, Initiative / Idea / Use-Case / Epic / Team / Product) which is
> implemented directly at v2. This issue is kept for historical context only.

| Field | Value |
|-------|-------|
| **ID** | #196 |
| **Title** | PTX-PFT `pft ideas` — Opportunity Management (Discovery / Backlog / Opportunity) |
| **Status** | ❌ Closed (Superseded by #198) |
| **Priority** | Medium |
| **Type** | Feature |
| **Labels** | enhancement, helper-binary, ptx-pft, iso-16355, qfd, opportunity-management, ai-funnel, json, graph |
| **Created** | 2026-07-08 |
| **Related** | ADR-008 (portunix-architecture), #116 (PTX-PFT ISO 16355 QFD structure), #191 (PTX-PFT graph visualization — done), api #014 (contracts), portunix-vscode #092 (Pilot GUI) |
| **Extended by** | #198 (`pft ideas` v2 — Venture model: Initiative / Use-Case / Epic / Team / Backlog / Product; `.discovery` → `.venture`) |

## Summary

Add a new **`pft ideas`** command group to the `ptx-pft` helper that authors a
directory-based, git-tracked **Opportunity Management** model per ADR-008:
a **Discovery** (a `.discovery` directory) owns one **backlog** of
**opportunities**, and generates a node-graph of the backlog for the Pilot
viewer. This is functionally separate from the existing feedback-tool sync
(Fider / ClearFlask) but shares the ISO 16355 / QFD lineage (#116).

## Motivation

ADR-008 chose a dev-first path: drive opportunity management from the VS Code
extension over JSON files, keep the JSON SaaS-ready, and reuse the existing
`pft graph` mechanism (#191) for the network view. `ptx-pft` is the home because
it already declares ISO 16355 / QFD compliance. The tool must support **both**
methodologies without forcing one: ISO 16355 (problem-first, AI competes with
other solution variants) and the AI Funnel (AI-first).

## Storage model (per ADR-008)

```text
<workspace>/ai.ideas/
└── <name>.discovery/
    ├── discovery.json          # name, description, approach (iso16355|ai-funnel|mixed), status
    ├── backlog.json            # list + relations (nodes+links) → reuses GraphView (*.glens.json)
    ├── <slug>.opportunity.json # one per opportunity (card data)
    ├── <slug>.discussion.json  # AI-assisted chat thread
    └── <slug>.estimation.json  # planning-poker estimation history
```

- `opportunity.json` carries `origin` (`ai-funnel` | `iso16355`), `status`
  (`new → ready → evaluated → poc → done | rejected`), `scores` (1–5),
  `estimate.storyPoints` (rollup = latest estimation), `interest.count/voters`.
- `estimation.json` keeps the full history: `{ estimator, argumentation,
  storyPoints, timestamp }`; the rollup mirrors the latest into
  `opportunity.json`.
- Schemas are owned by the api contract repo (api #014); ptx-pft validates
  writes against them.

## Commands

```bash
pft ideas new-discovery <name>             # create <name>.discovery/ + discovery.json + empty backlog.json
pft ideas add <slug> --title "..." [--origin ai-funnel|iso16355]
pft ideas list [--path <discovery>]        # opportunities with status + score
pft ideas show <slug>
pft ideas estimate <slug> --by <who> --sp <n> --why "..."   # append estimation, update rollup
pft ideas link <a> <b> --type inspired-by|duplicate|split-into|merged-with
pft ideas graph [--view] [--out <file>.glens.json]          # nodes+links for GraphCanvas/GraphLens3D
```

## Scope

### In scope

- `pft ideas` Cobra subcommand group in `src/helpers/ptx-pft`.
- Read/write of `discovery.json`, `backlog.json`, `*.opportunity.json`,
  `*.discussion.json`, `*.estimation.json` with schema validation (api #014).
- `pft ideas graph` — build the backlog `*.glens.json` (reuse #191 graph
  machinery, retargeted at a `.discovery`; nodes = opportunities, links =
  relations with `rel` mapping).
- Estimation append + rollup (latest → `opportunity.json.estimate.storyPoints`).
- Unit tests for the file model and graph generation; a fixture `.discovery`.

### Out of scope (other issues / later)

- The GUI (portunix-vscode #092) — implementation starts there GUI-first.
- AI chat actions (`/improve`, `/score`, …) — run through Claude Code from the
  Pilot side, not from this CLI.
- Variant B (global backlog, M:N Discovery) — ADR-008 open question.
- Migration of JSON to a database / web app.

## Tasks

- [ ] `pft ideas` command group scaffolding (Cobra), separate from `pft sync`
- [ ] `new-discovery` — create dir + `discovery.json` + empty `backlog.json`
- [ ] `add` / `list` / `show` — opportunity CRUD over `*.opportunity.json`
- [ ] `estimate` — append to `*.estimation.json` + rollup to opportunity
- [ ] `link` — write relations into `backlog.json`
- [ ] `graph [--view]` — emit `*.glens.json` (reuse #191), open via graphlens
- [ ] Schema validation against api #014 contracts
- [ ] Tests + fixture `.discovery` (align with portunix-vscode example)
- [ ] README: document `pft ideas` (separate section from feedback sync)

## Acceptance Criteria

1. `pft ideas new-discovery ai-in-HR` creates `ai-in-HR.discovery/` with
   `discovery.json` + empty `backlog.json`.
2. `pft ideas add ocr-cv --title "OCR of CVs"` creates
   `ocr-cv.opportunity.json` and registers a node in `backlog.json`.
3. `pft ideas estimate ocr-cv --by Zdenek --sp 8 --why "..."` appends to
   `ocr-cv.estimation.json` and sets `estimate.storyPoints = 8` on the opportunity.
4. `pft ideas link a b --type inspired-by` adds a link with the correct `rel`.
5. `pft ideas graph` produces a `*.glens.json` that validates against the
   GraphView contract; `--view` opens it via the graphlens plugin (as in #191).
6. All written JSON validates against the api #014 schemas.
7. Feedback-tool sync commands are unaffected (separate namespace).

## References

- ADR-008: Opportunity Management (portunix-architecture
  `docs/adr/ADR-008-opportunity-management.md`)
- #116 PTX-PFT ISO 16355 QFD Project Structure
- #191 PTX-PFT Graph Build & Visualization (done)
- Design mock: portunix-architecture
  `docs/architecture/brainstorming/opportunity-management/opportunity-card-mockup.svg`
