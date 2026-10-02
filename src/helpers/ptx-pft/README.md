# PTX-PFT - Product Feedback Tool Helper

Helper binary for managing integration with external Product Feedback Tools (Fider.io, Canny, ProductBoard, etc.).

## Features

- **Provider abstraction** - easy to add new feedback tool providers
- **Bidirectional sync** - synchronize between local markdown documents and external feedback systems
- **Container deployment** - deploy feedback tools via `portunix container compose`
- **Configuration management** - JSON-based configuration (`.pft-config.json`)

## Standards Compliance

This tool aims to be compliant with **ISO 16355** (Quality Function Deployment - QFD). ISO 16355 provides a framework for translating customer
requirements into product specifications through systematic collection and analysis of customer feedback.

Key QFD concepts supported:

- **Voice of Customer (VoC)** - collecting and organizing customer feedback
- **Requirements management** - linking feedback to product documentation
- **Prioritization** - status and priority tracking for feedback items
- **Traceability** - bidirectional sync between external feedback and local documentation

## Quick Start

### Option 1: Full Example (Recommended for first try)

```bash
# Single command - creates demo with sample use cases
./portunix pft example

# Or specify custom path
./portunix pft example --path /tmp/pft-demo

# Open http://localhost:3000 in browser
```

This will:

1. Create demo directory with 3 sample use cases
2. Configure ptx-pft automatically
3. Deploy feedback tool container
4. Push sample use cases to feedback tool
5. Open browser with running instance

### Option 2: Manual Setup

```bash
# 1. Configure product
./portunix pft configure --name "My Product" --path /path/to/docs

# 2. Deploy feedback tool (requires Docker or Podman)
./portunix pft deploy

# 3. Check status
./portunix pft status

# 4. Open in browser
#    http://localhost:3000

# 5. Cleanup when done
./portunix pft destroy           # keep data
./portunix pft destroy --volumes # remove everything
```

## Commands

| Command | Description |
| ------- | ----------- |
| `pft example` | Full demo: configure + deploy + sample data |
| `pft configure` | Interactive configuration wizard |
| `pft configure --show` | Show current configuration |
| `pft deploy` | Deploy feedback tool to container |
| `pft status` | Check feedback tool status |
| `pft destroy` | Remove feedback tool instance |
| `pft sync` | Bidirectional sync (Phase 4) |
| `pft list` | List feedback items (Phase 3) |
| `pft graph` | Build a node-graph (graphlens JSON) from the project |
| `pft graph --view` | Build and open the graph via the graphlens plugin |
| `pft ideas` | Opportunity Management v2 — author a `.venture` workspace |

## Graph Visualization

`pft graph` turns a PFT project into a node-graph following the graphlens JSON
contract (`meta` + `nodes` + `links`). Voices become `voice` nodes; hub nodes
(tags, author, or domains) group them; `[[wikilink]]` / `related` frontmatter
cross-references become `rel: "related"` links.

```bash
# Build only — write the graph JSON
portunix pft graph --path <project-dir> --out graph.json

# Choose what the central hub nodes represent (default: tags)
portunix pft graph --path <project-dir> --hub tags
portunix pft graph --path <project-dir> --hub author
portunix pft graph --path <project-dir> --hub domains

# Build and view in the browser via the graphlens plugin
portunix pft graph --path <project-dir> --view
portunix pft graph --path <project-dir> --view --port 8080 --no-open
```

| Option | Description |
| ------ | ----------- |
| `--path <dir>` | PFT project root (overrides config) |
| `--out <file>` | Output JSON path (default `graph.json` in the project dir) |
| `--hub <tags\|author\|domains>` | What the hub nodes represent (default `tags`) |
| `--view` | After building, open the graph via `portunix graphlens view` |
| `--port <n>` / `--no-open` | Forwarded to `graphlens view` when `--view` is used |

The graph JSON is always written first, so it remains usable even when the
graphlens plugin is not installed — the command then prints an install hint
(`portunix plugin install graphlens`) instead of failing. See
[portunix-plugins #086 / ADR-022](../../../docs/adr/) for the viewer and the
JSON contract.

## Opportunity Management v2 — `pft ideas` (Venture model)

`pft ideas` authors a self-contained, git-tracked **`.venture`** workspace and
computes derived views (graphs, story-point / complexity rollups). It is
functionally separate from the feedback-tool sync above and validates every
write against the api #015 contract schemas. Extends #196; implemented directly
at v2 per [#198](../../../docs/issues/internal/done/198-ptx-pft-ideas-venture-model-v2.md).

The model has two spines:

- **Discovery / strategy** — Venture → Initiative → **Idea**. An idea is
  product-agnostic and carries a coarse t-shirt **`complexity`** (`xs|s|m|l|xl`),
  never story points.
- **Delivery** — Team/Project → Epic → **Use-Case**. A use-case is bound to one
  **product**, `implements` one or more ideas, and is where real **`storyPoints`**
  live (planning-poker `estimate`).

Entities are referenced by **slug** (the slug is stored verbatim as the id).

### Storage layout (`<name>.venture/`, v2.1)

```text
<name>.venture/
├── venture.json
├── products/<slug>.product.json
├── teams/<slug>.team.json
├── initiatives/<slug>.initiative.json   # ideaRefs[]
├── records/                             # ideas + use-cases pool
│   ├── <slug>.opportunity.json          # Idea: complexity (no story points)
│   ├── <slug>.usecase.json              # storyPoints, productRef, implementsIdeaRefs[]
│   └── <slug>.estimation.json           # planning-poker history
└── backlogs/                            # sibling of teams/, first-class
    ├── <slug>.backlog.json              # kind (discovery|delivery), teamRef?, members
    ├── <slug>.epic.json                 # backlogRef, initiativeRefs[], useCaseRefs[]
    └── <slug>.glens.json                # derived GraphView projection
```

### Example

```bash
# Discovery spine
portunix pft ideas new-venture ai-in-HR
portunix pft ideas add-idea ocr-cv --title "OCR of CVs" --path ai-in-HR.venture
portunix pft ideas complexity ocr-cv --set m --path ai-in-HR.venture
portunix pft ideas add-initiative ai-in-hr --title "AI in HR" --path ai-in-HR.venture
portunix pft ideas initiative-link ai-in-hr --idea ocr-cv --path ai-in-HR.venture

# Delivery spine
portunix pft ideas add-product pilot --name "Pilot" --path ai-in-HR.venture
portunix pft ideas add-usecase reco-ocr --product pilot --implements ocr-cv --path ai-in-HR.venture
portunix pft ideas estimate reco-ocr --by Zdenek --sp 8 --why "…" --path ai-in-HR.venture
portunix pft ideas add-team platform --path ai-in-HR.venture
portunix pft ideas new-backlog hr-delivery --kind delivery --team platform --path ai-in-HR.venture
portunix pft ideas add-epic hr-ai --backlog hr-delivery --initiative ai-in-hr --path ai-in-HR.venture
portunix pft ideas epic-link hr-ai --usecase reco-ocr --path ai-in-HR.venture

# Derived views
portunix pft ideas graph --scope discovery --path ai-in-HR.venture      # ideas + implements edges
portunix pft ideas graph --backlog hr-delivery --view --path ai-in-HR.venture
portunix pft ideas rollup --initiative ai-in-hr --path ai-in-HR.venture  # Σ SP + Σ complexity
portunix pft ideas list --path ai-in-HR.venture
```

Run `portunix pft ideas --help` for the full command reference. Migrate a v1
`.discovery` directory with `portunix pft ideas migrate <name>.discovery`.

## Configuration

Configuration is stored in `.pft-config.json`:

```json
{
  "name": "My Product",
  "path": "/path/to/docs",
  "provider": "fider",
  "endpoint": "http://localhost:3000",
  "api_token": ""
}
```

### Cross-Platform Path Resolution

The `path` field accepts three forms (see [ADR-032](../../../docs/adr/032-pft-cross-platform-config-path-resolution.md)):

| Form | Example | Resolves to |
| ---- | ------- | ----------- |
| Empty | `"path": ""` | Directory containing `.pft-config.json` (recommended for git-tracked configs) |
| Relative | `"path": "docs/pft"` | Joined to the config file directory |
| Absolute | `"path": "/home/user/project"` | Used as-is (works only on the OS it was written on) |

When the `--path` flag is supplied to a command, it always overrides the stored
`path` field — this is what makes a config committed on Linux usable on Windows
and vice versa:

```bash
# Linux author committed an absolute Linux path; on Windows, override it:
portunix pft list --path "E:\Dev\project\docs\pft-project"
```

### Migrating an Existing Config to Cross-Platform

To convert an existing absolute `path` into an empty or relative form so the
config travels across operating systems, run:

```bash
portunix pft configure --fix-paths --path /current/project/dir
```

The command sets `path` to empty when it equals the config directory, or to a
relative path when the project lives elsewhere under the same root. Paths that
cannot be expressed relatively (different drive, too far up the tree) are
cleared to empty so callers must use `--path` explicitly.

## Architecture

```text
ptx-pft/
├── main.go       # CLI entry point (Cobra)
├── config.go     # Configuration management
├── provider.go   # FeedbackProvider interface
├── deploy.go     # Container deployment logic
└── README.md     # This file
```

## Provider Interface

New providers implement `FeedbackProvider` interface:

```go
type FeedbackProvider interface {
    Name() string
    Connect(config ProviderConfig) error
    List() ([]FeedbackItem, error)
    Get(id string) (*FeedbackItem, error)
    Create(item FeedbackItem) (*FeedbackItem, error)
    Update(item FeedbackItem) error
    Delete(id string) error
    Close() error
}

```text
## Feedback tools comparation

| Aspect | Fider | ClearFlask |
| ------ | ----- | ---------- |
| Backend | Go (simple) | Java + Tomcat |
| Database | PostgreSQL | MySQL + DynamoDB + Elasticsearch |
| RAM | ~200MB | ~2GB+ |
| Complexity | Low | High |
| Multi-project | No | Yes |
| Roadmap | Basic | Advanced |

## Dependencies

- `portunix container compose` - for container operations
- `assets/packages/fider.json` - package definition

## Implementation Status

- [x] Phase 1: Helper foundation (CLI, config, provider interface)
- [x] Phase 2: Container deployment (deploy, status, destroy)
- [ ] Phase 3: Fider.io API client
- [ ] Phase 4: Synchronization engine
- [ ] Phase 5: Reporting and export
