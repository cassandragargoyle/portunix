# Issue #191: PTX-PFT Graph Build & Visualization (via graphlens plugin)

| Field | Value |
|-------|-------|
| **ID** | #191 |
| **Title** | PTX-PFT Graph Build & Visualization (via graphlens plugin) |
| **Status** | ✅ Implemented |
| **Priority** | Medium |
| **Type** | Enhancement |
| **Labels** | enhancement, helper-binary, ptx-pft, qfd, iso-16355, visualization, graph, graphlens, plugin-integration |
| **Created** | 2026-06-28 |
| **Closed** | 2026-07-02 |

## Problem Statement

A PTX-PFT (ISO 16355 / QFD) project is a network of linked Markdown files —
Voices (VoC/VoB/VoE/VoS), their verbatims and needs, technical requirements, and
the `[[wikilink]]` cross-references between them (parent/child needs, sources,
addressed-needs, conflicts, related). Today this structure can only be read
file-by-file; there is no way to **see** the network — which needs are
unconnected, which requirements address many Voices, where conflicts cluster.

We want `ptx-pft` to **build a node-graph from a PFT project** and, when the new
**graphlens** plugin is available, **display it** as an interactive 3D
force-directed graph in the browser.

This is the relocation of the original `build-pft` logic that briefly lived in
the standalone `graphlens` script. The graphlens plugin
(`portunix-plugins`, issue #086 / ADR-022) is now a **domain-agnostic viewer**:
it renders any conforming `nodes` + `links` JSON and knows nothing about PFT.
The PFT-specific knowledge (folder layout, frontmatter fields, ISO 16355
prefixes) belongs here, in `ptx-pft`.

## Goals

1. **Build**: `ptx-pft` produces a graph JSON (the graphlens contract) from a PFT
   project directory.
2. **View**: when the `graphlens` plugin is installed, `ptx-pft` can hand the
   JSON to it for rendering (`portunix graphlens view`), with graceful fallback
   when the plugin is absent.

## Proposed Command

Add a `graph` subcommand to the `pft` dispatcher (`src/helpers/ptx-pft/main.go`,
`handlePFTCommand` switch), GNU/POSIX option order, with `--help`:

```bash
# Build only — write the graph JSON
portunix pft graph --path <project-dir> --out graph.json

# Choose what the central hub nodes represent (generation-time)
portunix pft graph --path <project-dir> --hub tags     # default
portunix pft graph --path <project-dir> --hub author
portunix pft graph --path <project-dir> --hub domains

# Build and view in the browser via the graphlens plugin
portunix pft graph --path <project-dir> --view
portunix pft graph --path <project-dir> --view --port 8080 --no-open
```

| Option | Description |
| ------ | ----------- |
| `--path <dir>` | PFT project root (same resolution as other `pft` commands: `--path` overrides config, see ADR-032) |
| `--out <file>` | Output JSON path (default e.g. `graph.json` in the project dir) |
| `--hub <tags\|author\|domains>` | What the hub nodes represent (default `tags`) |
| `--view` | After building, open the graph in the browser via `portunix graphlens view` |
| `--port <n>` / `--no-open` | Forwarded to `graphlens view` when `--view` is used |

## Graph Data Contract (must match graphlens)

The output is the viewer-agnostic contract defined by the graphlens plugin
(`portunix-plugins` ADR-022): a single JSON document with `meta` + `nodes` +
`links`.

```json
{
  "meta": { "title": "…", "hub": "tags", "voices": 42, "hubs": 7, "related": 5 },
  "nodes": [
    { "id": "tag:hr", "type": "tag", "label": "hr" },
    { "id": "VC-V015", "type": "voice", "voice": "VoC", "label": "VC-V015",
      "title": "…", "description": "…", "meta": { "category": "…", "status": "…" } }
  ],
  "links": [
    { "source": "tag:hr", "target": "VC-V015", "description": "VC-V015 is tagged \"hr\"" },
    { "source": "VC-V015", "target": "VC-S03", "rel": "related", "description": "…" }
  ]
}
```

- Hub nodes use `type: "tag"` (the viewer's generic hub type) regardless of the
  `--hub` dimension, so the unchanged viewer renders them as hubs.
- Voice-to-voice cross-references (from `[[wikilinks]]` / `related` frontmatter)
  are emitted as links with `rel: "related"`, deduplicated as undirected pairs.

## PFT → Graph Mapping (the adapter logic)

Reuse the existing scan/frontmatter machinery (`ScanFeedbackDirectory`,
`getVoiceDir`, the QFD folder conventions from #116) rather than re-walking the
tree from scratch:

- Walk the Voice dirs (`VoC`/`VoB`/`VoE`/`VoS`, both QFD PascalCase and basic
  lowercase variants via `getVoiceDir`), over `verbatims/` and `needs/` (and
  `requirements/`).
- Emit one **node per item** (`id` from frontmatter / filename prefix:
  `VC-V001`, `VC-S01`, `R001`, …), `type: "voice"`, `voice` = VoC/VoB/VoE/VoS,
  `title`, a short `description` (first content paragraph), and selected
  frontmatter as `meta` (category, status, Kano, priority…).
- Emit hub nodes + links per the `--hub` dimension (tags/categories, author,
  domains).
- Emit `rel: "related"` links from `[[wikilink]]` references (parent/child,
  sources, addresses-needs, conflicts) — only between ids that exist as nodes.
- Skip a single malformed-frontmatter file with a logged warning; never abort
  the whole build.

> Implementation note: this is the same shape as the reference Python adapter
> `sources/pft_voices.py` staged at `%TEMP%\portunix-graphlens-src` (Windows:
> `C:\Users\<user>\AppData\Local\Temp\portunix-graphlens-src`), re-expressed in
> Go against the existing `ptx-pft` scanners. ISO 16355 prefixes (V/P/S/T/C, R)
> and the maturity/Kano fields come from #116.

## Plugin Integration (`--view`)

`ptx-pft` must not hard-depend on graphlens. When `--view` is requested:

1. Detect the plugin (e.g. `portunix plugin list` / a capability query, or a
   probe of `portunix graphlens --version`).
2. If present: build to a temp/`--out` file, then invoke
   `portunix graphlens view <file> [--port N] [--no-open]`.
3. If absent: write the JSON and print actionable guidance, e.g.
   `graphlens plugin not installed — run 'portunix plugin install graphlens', then 'portunix graphlens view graph.json'`.
   The build itself still succeeds (exit 0), so the JSON is usable without the plugin.

## Implementation Phases

### Phase 1: Graph builder
- [ ] Add `graph` to `handlePFTCommand`; flag parsing + `--help`
- [ ] Implement the PFT→graph adapter (nodes, hub links, related links) reusing
      existing scanners
- [ ] Emit the graphlens JSON contract; `--out` writing (UTF-8, no BOM)
- [ ] Unit tests on a fixture PFT project (node/link counts, hub dimensions,
      malformed-file skip)

### Phase 2: graphlens view integration
- [ ] Plugin detection + `--view` invocation of `portunix graphlens view`
- [ ] Graceful fallback message when the plugin is absent
- [ ] `--port` / `--no-open` forwarding

### Phase 3: Docs & polish
- [ ] Update `src/helpers/ptx-pft/README.md` (new `graph` command + examples)
- [ ] Add `graph` to `showPFTHelp()` and to `pft info` methodology docs
- [ ] Cross-link with graphlens plugin docs

## Acceptance Criteria

- [ ] `portunix pft graph --path <dir> --out graph.json` writes a valid
      graphlens-contract JSON (nodes + links + meta) from a PFT project
- [ ] `--hub tags|author|domains` changes the hub dimension without changing the
      viewer
- [ ] `[[wikilink]]` cross-references appear as `rel: "related"` links
- [ ] Malformed frontmatter in one file is skipped with a warning, build still succeeds
- [ ] `portunix pft graph --path <dir> --view` opens the graph via the graphlens
      plugin when installed
- [ ] When graphlens is **not** installed, the JSON is still written and a clear
      install hint is printed (no crash)
- [ ] `--help`, README, and `pft info` document the command
- [ ] Unit tests pass on a fixture project

## Related

- ISO 16355 — Quality Function Deployment
- ADR-029 — PTX-PFT Product Feedback Tool Helper (this helper)
- Issue #116 — PTX-PFT ISO 16355 QFD Project Structure (folder layout, prefixes, frontmatter)
- Issue #107 — PTX-PFT initial implementation
- **portunix-plugins #086 / ADR-022** — graphlens plugin (the domain-agnostic
  viewer this command targets; JSON contract source of truth)
- Reference adapter (staged copy): `%TEMP%\portunix-graphlens-src\sources\pft_voices.py`
