# Acceptance Protocol - Issue #198 (`pft ideas` v2 — Venture model)

**Issue**: PTX-PFT `pft ideas` v2 — Venture model (Initiative / Idea / Use-Case /
Epic / Team / Product)
**Branch**: `feature/198-pft-ideas-venture-model` (commit `7ce0012`)
**Tester**: zdendaku (QA/Test Engineer — generic)
**Date**: 2026-07-10
**Testing OS**: Ubuntu Linux 6.17.0-40-generic (host) — Go 1.25.0 toolchain

## Scope Note

Issue #196 (`pft ideas` v1, `.discovery` model) was never implemented in code,
so #198 implements the full `pft ideas` command group directly at v2 (`.venture`
two-spine model). This protocol covers the whole authoring surface, the derived
views (graph, rollup), schema validation, and the v1→v2 migration helper.
Out of scope per the issue: the Pilot GUI (portunix-vscode #098), AI chat
actions, and external backlog sync — not tested here.

## Test Summary

- Total test scenarios: 27 (11 unit/integration + 9 acceptance criteria + 7 failure injection)
- Passed: 27
- Failed: 0
- Conditional: 0

## Test Environment

- Helper `ptx-pft` built cleanly (`go build`, exit 0); full `make build` exit 0
  (main binary + all 19 helpers)
- `gofmt -l` clean, `go vet ./...` clean
- Execution via the main dispatcher: `./portunix pft ideas …` (real routing path)
- Schema validation via the api #015 contract validator (ajv 2020, Draft 2020-12)
  against `/home/zdenek/DEV/CassandraGargoyle/api/contract/schemas/`
- No containers required: #198 is a pure Go authoring feature (writes git-tracked
  JSON), not a software-installation flow, so the container-testing policy does
  not apply.

## Test Results

### Unit / Integration (`go test ./...`)

- [x] All `ptx-pft` tests pass (0.008s), including 11 new `TestIdeas_*` cases
- [x] Pre-existing feedback-tool tests still pass (no regression)

| Test | Verifies |
| ---- | -------- |
| `TestIdeas_NewVenture` | venture.json + empty discovery backlog |
| `TestIdeas_IdeaHasComplexityNotStoryPoints` | idea carries complexity, never storyPoints; backlog node added |
| `TestIdeas_UseCaseImplementsIdea` | productRef + implementsIdeaRefs |
| `TestIdeas_EstimateRollsUp` | estimation history + storyPoints rollup |
| `TestIdeas_DeliveryBacklogEpic` | epic backlogRef/useCaseRefs, delivery backlog + teamRef |
| `TestIdeas_DiscoveryGraphHasImplementsEdge` | discovery GraphView implements edge |
| `TestIdeas_BacklogGraphMatchesShape` | contains + implements edges |
| `TestIdeas_RollupInitiative` | Σ storyPoints (8) + complexity breakdown (m:1) |
| `TestIdeas_AllWrittenJSONValidates` | every entity passes `Validate()` |
| `TestIdeas_Migrate` | `.discovery`→`.venture`, legacy estimate dropped |
| `TestIdeas_ValidateRejectsBadEnum` | invalid enum / missing productRef rejected |

### Functional — Acceptance Criteria (e2e via `./portunix pft ideas`)

Given/When/Then style; each maps to an issue acceptance criterion.

- [x] **AC#1** — `new-venture ai-in-HR` → `ai-in-HR.venture/` with `venture.json`
  and an empty `backlogs/ai-in-HR.backlog.json` (kind `discovery`).
- [x] **AC#2** — `add-idea ocr-cv --title "OCR of CVs"` → `ocr-cv.opportunity.json`
  carries `complexity` (null until set), **no `storyPoints`**, and is registered
  as a node in the discovery backlog.
- [x] **AC#3** — `add-usecase reco-ocr --product pilot --implements ocr-cv` →
  `reco-ocr.usecase.json` with `productRef=pilot` and `implementsIdeaRefs=["ocr-cv"]`.
- [x] **AC#4** — `estimate reco-ocr --by Zdenek --sp 8 --why "…"` → appends
  `reco-ocr.estimation.json` and sets use-case `storyPoints = 8`.
- [x] **AC#5** — `new-backlog hr-delivery --kind delivery --team platform` +
  `add-epic hr-ai --backlog hr-delivery --initiative ai-in-hr` +
  `epic-link hr-ai --usecase reco-ocr` → delivery backlog (sibling of `teams/`)
  with epic referencing the use-case (`backlogRef`, `useCaseRefs`).
- [x] **AC#6** — `graph --scope discovery` → `discovery.glens.json` with an
  `implements` edge. `--view` launches the graphlens viewer (interactive; the
  plugin is installed on this host — see note below).
- [x] **AC#7** — `rollup --initiative ai-in-hr` → reports `Σ storyPoints
  (delivery): 8` and `Σ complexity (discovery): m:1 [weighted 3]`.
- [x] **AC#8** — all written JSON validates against the api #015 schemas
  (12/12 files across every entity type + `*.glens.json`; migrated venture 2/2).
- [x] **AC#9** — feedback-tool sync namespace unaffected: `pft sync/pull/push/
  configure` unchanged, pre-existing tests green.

### Regression

- [x] `make build` succeeds (main + 19 helpers)
- [x] Existing `ptx-pft` feedback-tool commands and tests unaffected
- [x] Dispatcher routes `portunix pft ideas …` to the helper correctly

### Failure Injection

- [x] **FI1** unknown `--product` → clear error, no file written
- [x] **FI2** `--implements` unknown idea → clear error, no file written
- [x] **FI3** invalid `complexity` enum → rejected with allowed-values message
- [x] **FI4** invalid backlog `--kind` → rejected with usage hint
- [x] **FI5** command outside a venture → actionable "no .venture found" message
- [x] **FI6** duplicate `link … implements` → idempotent (uniqueItems honored)
- [x] **FI7** failed writes leave no partial/corrupt files behind

## Observations / Recommendations (non-blocking)

1. **Exit codes** — failing `pft ideas` commands print an error but exit `0`.
   This matches the existing base `pft` command convention (e.g. an unknown
   subcommand also exits `0`), so it is **not a regression introduced by #198**.
   Recommendation (whole-helper, future): return non-zero on error for better
   CI/scripting ergonomics.
2. **`ideas graph --view` flag parity** — unlike the base `pft graph`, the new
   `ideas graph --view` does not expose `--no-open` / `--port`; it always opens
   the viewer and blocks (expected interactive behavior, but not automatable).
   Recommendation (future): add `--no-open` / `--port` for parity and headless CI.

Both are minor enhancements, out of scope for this issue's acceptance.

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-07-10
**Tester signature**: zdendaku (QA/Test Engineer — generic)
