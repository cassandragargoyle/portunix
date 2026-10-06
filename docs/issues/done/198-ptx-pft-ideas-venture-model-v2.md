# Issue #198: PTX-PFT `pft ideas` v2 — Venture model (Initiative / Idea / Use-Case / Epic / Team / Product)

| Field | Value |
|-------|-------|
| **ID** | #198 |
| **Title** | PTX-PFT `pft ideas` v2 — Venture model (Initiative / Idea / Use-Case / Epic / Team / Product) |
| **Status** | ✅ Implemented |
| **Priority** | Medium |
| **Type** | Feature |
| **Labels** | enhancement, helper-binary, ptx-pft, opportunity-management, venture, use-case, initiative, epic, team, product, iso-16355, json, graph |
| **Created** | 2026-07-10 |
| **Closed** | 2026-07-11 |
| **Related** | #196 (pft ideas v1 — extended by this issue), ADR-008 (portunix-architecture), #116 (ISO 16355 QFD structure), #191 (pft graph — done), api #015 (v2 contracts), portunix-vscode #098 (Pilot v2 GUI) |

## Summary

Extend the **`pft ideas`** command group (#196) from the single-tier
Discovery/Opportunity model to the **v2 two-spine model** agreed in the
2026-07-10 brainstorming session. `ptx-pft` becomes the authoring/derivation
engine for a **self-contained `.venture`** workspace holding both spines:

- **Discovery/strategy:** Venture → Initiative → Idea (Idea carries coarse
  **`complexity`**, not story points).
- **Delivery:** Team/Project → Epic → Use-Case (Use-Case is product-bound, carries
  **`storyPoints`** via planning-poker, `implements` one or more Ideas).

The container is renamed `.discovery` → **`.venture`**. `ptx-pft` **must provide a
first-class command to create a venture** and to author every entity, validating
every write against the api #015 schemas. It also **computes derived views**
(graphs, SP/complexity rollups); the Pilot UI only renders.

## Motivation

ADR-008 keeps opportunity management dev-first over git-tracked JSON, with `ptx-pft`
as the Go writer/deriver and Pilot as the reader. v1 (#196) put SP on the idea; v2
corrects this — SP belong to the concrete, product-bound use-case, while the idea
keeps a coarse product-agnostic complexity. The tool still supports both
methodologies (ISO 16355 problem-first and the AI Funnel) without forcing one.

## Storage model (v2.1, per api #015 — looser, backlog is first-class)

`teams/` and `backlogs/` are **sibling collections** (a backlog is not nested in a
team). A **Backlog** is first-class with `kind` (`discovery` | `delivery`) and an
*optional* `teamRef`; an **Epic** belongs to a backlog (`backlogRef`), so the team
link is indirect. Records (ideas + use-cases) live in a `records/` pool.

```text
<name>.venture/                      # top-level, self-contained
├── venture.json
├── products/<slug>.product.json
├── teams/<slug>.team.json           # team/project registry
├── initiatives/<slug>.initiative.json   # ideaRefs[] (M:N)
├── records/                         # records pool
│   ├── <slug>.opportunity.json      # Idea: complexity, scores, interest, volunteers
│   ├── <slug>.usecase.json          # storyPoints, productRef, implementsIdeaRefs[]
│   ├── <slug>.estimation.json       # planning-poker — beside its use-case
│   └── <slug>.discussion.json       # generic thread (idea|use-case|initiative|epic)
└── backlogs/                        # SAME LEVEL as teams (first-class)
    ├── <slug>.backlog.json          # kind, optional teamRef, member+relation refs → GraphView (#191)
    └── <slug>.epic.json             # backlogRef, initiativeRefs[], useCaseRefs[]
```

**Linking principle:** every `*Ref` points from the volatile entity (initiative,
epic, use-case, backlog) to the stable asset (idea, product, team); single-owner
refs keep edits local.

## Commands

```bash
# --- container / registry ---
pft ideas new-venture <name>                     # create <name>.venture/ + venture.json (+ empty backlog.json)
pft ideas add-product <slug> --name "..." [--repo <url>]
pft ideas add-team <slug> --name "..." [--kind team|project]

# --- discovery spine ---
pft ideas add-initiative <slug> --title "..." [--approach iso16355|ai-funnel|mixed]
pft ideas add-idea <slug> --title "..." [--origin ai-funnel|iso16355]   # alias: add
pft ideas complexity <idea> --set xs|s|m|l|xl
pft ideas initiative-link <initiative> --idea <idea>                    # append to ideaRefs[]

# --- delivery spine ---
pft ideas add-usecase <slug> --title "..." --product <product> --implements <idea>[,<idea>...]
pft ideas new-backlog <slug> --kind discovery|delivery [--team <team>]  # first-class, sibling of teams
pft ideas add-epic <slug> --backlog <backlog> --title "..." [--initiative <initiative>]
pft ideas epic-link <epic> --usecase <usecase>                         # append to useCaseRefs[]
pft ideas estimate <usecase> --by <who> --sp <n> --why "..."           # append estimation, roll up SP

# --- relations / derived ---
pft ideas link <a> <b> --type inspired-by|duplicate|split-into|merged-with|implements
pft ideas graph [--backlog <slug>|--scope discovery] [--view] [--out <file>.glens.json]
pft ideas rollup [--initiative <i>|--backlog <b>|--epic <e>]           # recompute derived SP/complexity
pft ideas list [--kind idea|usecase|epic|initiative] [--path <venture>]
pft ideas show <slug>
```

## Scope

### In scope

- `pft ideas new-venture` and the full authoring surface above in
  `src/helpers/ptx-pft`, separate from the feedback-tool sync namespace.
- Read/write + schema validation (api #015) for `venture.json`, `product.json`,
  `initiative.json`, `opportunity.json` (with `complexity`), `use-case.json`,
  `team.json`, `epic.json`, `discussion.json`, `estimation.json`.
- **Graph generation** (`pft ideas graph`) for both scopes: the discovery backlog
  (ideas + `implements` edges to use-cases) and per-team delivery backlog (epics +
  use-cases); reuse #191 machinery, emit GraphView `*.glens.json`.
- **Rollups**: use-case `storyPoints` = latest estimation; epic ΣSP; initiative ΣSP
  (via epics) and Σ complexity (via `ideaRefs`).
- Migration helper `.discovery` → `.venture` (rename dir, move estimation from idea
  to any created use-case, `estimate.storyPoints` → drop; keep git history sane).
- Unit tests + a fixture `.venture` aligned with the portunix-vscode example.

### Out of scope (other issues / later)

- The GUI (portunix-vscode #098) — the model is still driven GUI-first there.
- AI chat actions — run through Claude Code from the Pilot side.
- External-system backlog sync (Jira/ADO/Redmine) — future, reuse the existing
  Fider/ClearFlask provider abstraction.
- Global cross-venture product/team catalog; DB/web migration.

## Tasks

- [ ] `new-venture` — create `<name>.venture/` + `venture.json` + empty `backlog.json`
- [ ] `add-product` / `add-team` — registries inside the venture
- [ ] `add-initiative` (+ `initiative-link`) — discovery spine, `ideaRefs[]`
- [ ] `add-idea` / `complexity` — idea CRUD with coarse complexity (no SP)
- [ ] `add-usecase` — product-bound, `implementsIdeaRefs[]`, `productRef`
- [ ] `new-backlog` — first-class backlog (`kind`, optional `teamRef`), sibling of `teams/`
- [ ] `add-epic` (+ `epic-link`) — `backlogRef`, `initiativeRefs[]`, `useCaseRefs[]`
- [ ] `estimate` — append to use-case `*.estimation.json` + roll up `storyPoints`
- [ ] `link` — relations incl. `implements` into a backlog
- [ ] `graph [--backlog|--scope] [--view]` — GraphView emit per backlog (reuse #191)
- [ ] `rollup` — derived SP/complexity for initiatives and epics
- [ ] Schema validation against api #015
- [ ] `.discovery` → `.venture` migration helper
- [ ] Tests + fixture `.venture` (align with portunix-vscode #098 example)
- [ ] README: document `pft ideas` v2 (note #196 → #198 extension)

## Acceptance Criteria

1. `pft ideas new-venture ai-in-HR` creates `ai-in-HR.venture/` with `venture.json`
   and an empty `backlog.json`.
2. `pft ideas add-idea ocr-cv --title "OCR of CVs"` creates
   `ocr-cv.opportunity.json` (with `complexity`, no story points) and a backlog node.
3. `pft ideas add-usecase reco-ocr --product pilot --implements ocr-cv` creates
   `reco-ocr.usecase.json` with `productRef` + `implementsIdeaRefs=["ocr-cv"]`.
4. `pft ideas estimate reco-ocr --by Zdenek --sp 8 --why "..."` appends to
   `reco-ocr.estimation.json` and sets the use-case `storyPoints = 8`.
5. `pft ideas new-backlog hr-delivery --kind delivery --team platform`,
   `add-epic hr-ai --backlog hr-delivery --initiative ai-in-hr` and
   `epic-link hr-ai --usecase reco-ocr` produce a delivery backlog (sibling of
   `teams/`) with an epic referencing the use-case.
6. `pft ideas graph --scope discovery` emits a GraphView `*.glens.json` with an
   `implements` edge; `--view` opens it via the graphlens plugin (as in #191).
7. `pft ideas rollup --initiative ai-in-hr` reports both Σ storyPoints (delivery)
   and Σ complexity (discovery).
8. All written JSON validates against the api #015 schemas.
9. Feedback-tool sync commands remain unaffected (separate namespace).

## References

- ADR-008: Opportunity Management (portunix-architecture
  `docs/adr/ADR-008-opportunity-management.md`) — to be updated for v2
- #196 pft ideas v1 (extended here)
- #116 ISO 16355 QFD structure, #191 pft graph (done)
- api #015 (v2 contracts), portunix-vscode #098 (v2 GUI)
- v2 model: portunix-architecture
  `docs/architecture/brainstorming/opportunity-management/journal-20260710-01.json`
