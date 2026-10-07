# Issue #116: PTX-PFT ISO 16355 QFD Project Structure

| Field | Value |
|-------|-------|
| **ID** | #116 |
| **Title** | PTX-PFT ISO 16355 QFD Project Structure |
| **Status** | 📋 Open |
| **Priority** | High |
| **Type** | Enhancement |
| **Labels** | enhancement, helper-binary, ptx-pft, iso-16355, qfd, requirements-management |
| **Created** | 2025-12-26 |

## Problem Statement

Current PTX-PFT project creation generates basic Voice directories (VoC, VoB, VoE, VoS) but lacks the structured subdirectories and documentation required for proper ISO 16355 QFD (Quality Function Deployment) methodology implementation. Users need a complete requirements management system with proper file templates, naming conventions, and workflow documentation.

## Requirements

### 1. Enhanced Voice Directory Structure

Each Voice directory shall contain standardized subdirectories:

```
VoX/
├── README.md           # Voice description, purpose, sources
├── verbatims/          # Raw statements (literal quotes)
│   └── .gitkeep
└── needs/              # Structured needs (Primary, Secondary, Tertiary)
    └── .gitkeep
```

**Exception for VoE** (Voice of Engineering):
```
VoE/
├── README.md
├── verbatims/
│   └── .gitkeep
├── needs/
│   └── .gitkeep
└── constraints/        # Technical constraints and dependencies
    └── .gitkeep
```

### 2. New Directory Creation

#### 2.1 Directory `requirements/`

```
requirements/
├── README.md           # Description, workflow, Voice linkages
└── .gitkeep
```

Purpose: Evidence of technical requirements derived from all Voices' needs.

#### 2.2 Directory `matrices/`

```
matrices/
├── README.md           # Matrix descriptions, reading and update guidance
└── .gitkeep
```

Purpose: House of Quality matrices, prioritization tables, relationship mapping.

### 3. Naming Conventions

#### File Prefixes

| Voice | Verbatim | Need |
|-------|----------|------|
| VoC | VC-V001 | VC-P01, VC-S01, VC-T01 |
| VoB | VB-V001 | VB-P01, VB-S01, VB-T01 |
| VoE | VE-V001 | VE-C01 (constraints) |
| VoS | VS-V001 | VS-P01, VS-S01, VS-T01 |
| Requirements | – | R001, R002, ... |

#### Need Levels (ISO 16355)

| Level | Prefix | Description |
|-------|--------|-------------|
| Primary | P | Strategic need (highest abstraction) |
| Secondary | S | Tactical need |
| Tertiary | T | Specific, measurable need |

#### Maturity Statuses

| Status | Meaning |
|--------|---------|
| Captured | Raw input recorded |
| Structured | Placed in hierarchy |
| Analyzed | Context and motivation understood |
| Prioritized | Importance determined (AHP, Kano) |
| Translated | Converted to technical requirement |
| Deployed | Handed off to implementation |

#### Kano Categories

| Category | Description |
|----------|-------------|
| Must-be | Basic requirement (non-fulfillment = dissatisfaction) |
| Performance | Linear relationship between fulfillment and satisfaction |
| Delighter | Surprise (unexpected value) |

### 4. README File Requirements

Each directory requires a README.md with:
- Purpose description
- Sources (where input comes from)
- Structure explanation
- File templates (copy-paste ready)

### 5. Workflow Documentation

Main README.md must document the QFD workflow:

```
Verbatim → Reworded Need → Customer Requirement → Quality Characteristic → Technical Requirement
```

1. **Acquisition** – Capture verbatim into `VoX/verbatims/`
2. **Structuring** – Create need in `VoX/needs/` linked to verbatim
3. **Analysis** – Add context, Kano category
4. **Prioritization** – AHP score, update `matrices/`
5. **Translation** – Create technical requirement in `requirements/`

## Proposed Architecture

### Component Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                 PTX-PFT Project Structure                        │
└─────────────────────────────────────────────────────────────────┘

project/
├── README.md                    # Project overview, conventions, workflow
├── VoC/                         # Voice of Customer
│   ├── README.md                # VoC purpose, sources, templates
│   ├── verbatims/
│   │   └── .gitkeep
│   └── needs/
│       └── .gitkeep
├── VoB/                         # Voice of Business
│   ├── README.md
│   ├── verbatims/
│   │   └── .gitkeep
│   └── needs/
│       └── .gitkeep
├── VoE/                         # Voice of Engineering
│   ├── README.md
│   ├── verbatims/
│   │   └── .gitkeep
│   ├── needs/
│   │   └── .gitkeep
│   └── constraints/             # VoE-specific
│       └── .gitkeep
├── VoS/                         # Voice of Stakeholder
│   ├── README.md
│   ├── verbatims/
│   │   └── .gitkeep
│   └── needs/
│       └── .gitkeep
├── requirements/                # Technical requirements
│   ├── README.md
│   └── .gitkeep
└── matrices/                    # QFD matrices
    ├── README.md
    └── .gitkeep
```

### Implementation Options

#### Option A: Extend Existing `project create` Command

Modify current project creation to include full ISO 16355 structure by default.

**Pros:**
- Single command for complete setup
- Consistent behavior

**Cons:**
- Breaking change for existing workflows
- May be too verbose for simple projects

#### Option B: New `project create --template` Flag (Default: qfd)

Add template parameter to project creation with `qfd` as default.

```bash
./portunix pft project create "Product X"                   # uses --template qfd (default)
./portunix pft project create "Product X" --template qfd    # explicit QFD template
./portunix pft project create "Product X" --template basic  # minimal structure (legacy)
```

**Pros:**
- QFD structure by default (primary use case)
- User can opt-out with `--template basic`

**Cons:**
- Breaking change for existing scripts expecting old structure

#### Option C: Separate `project init-qfd` Command

New command specifically for ISO 16355 initialization.

```bash
./portunix pft project create "Product X"      # basic structure
./portunix pft project init-qfd                # adds QFD structure to existing project
```

**Pros:**
- Can be applied to existing projects
- Clear separation of concerns

**Cons:**
- Two commands needed for full setup

### Recommended Approach

**Option B** - `--template` flag with `qfd` as default when template is not specified.

Rationale:
- Full ISO 16355 structure is the primary use case
- Simple projects can opt-out with `--template basic`
- No extra typing required for standard QFD workflow

## Implementation Phases

### Phase 1: Core Structure Generation
- [ ] Implement subdirectory creation (verbatims/, needs/, constraints/)
- [ ] Generate .gitkeep files for empty directories
- [ ] Add `--template` flag to project create command

### Phase 2: README Generation
- [ ] Create README templates for each Voice directory
- [ ] Create main project README with conventions and workflow
- [ ] Create requirements/ and matrices/ README files

### Phase 3: File Templates
- [ ] Embed verbatim templates in README files
- [ ] Embed need templates in README files
- [ ] Embed requirement templates
- [ ] Embed matrix templates

### Phase 4: Validation & Tooling
- [ ] Add `project validate` command to check structure
- [ ] Add `project status` to show coverage statistics

### Phase 5: MCP Integration
- [ ] Add `ptx-pft info` command returning documentation for all implemented methodologies
- [ ] Integrate into `ptx-mcp` as tool `pft_info` providing PTX-PFT usage guide
- [ ] Info content returned by ptx-pft binary (single source of truth)
- [ ] Methodologies documented: `basic`, `qfd` (ISO 16355)

## Template Specifications

### Main README.md Template

```markdown
# [Project Name]

## Overview

This project uses QFD (Quality Function Deployment) methodology per ISO 16355
for systematic requirements management from voice of customer collection
through technical specification.

## Project Structure

| Directory | Purpose |
|-----------|---------|
| VoC/ | Voice of Customer – customer needs |
| VoB/ | Voice of Business – business goals and constraints |
| VoE/ | Voice of Engineering – technical capabilities and constraints |
| VoS/ | Voice of Stakeholder – regulations, partners, internal stakeholders |
| requirements/ | Technical requirements (intersection of all Voices) |
| matrices/ | QFD matrices and prioritization |

## Naming Conventions

[... full conventions table ...]

## Workflow

[... workflow documentation ...]

## Tools

- **Editor:** MarkText / Obsidian / VS Code with Markdown preview
- **Version Control:** Git
- **Wikilinks:** `[[VC-V001]]` for inter-file navigation
```

### Voice README Templates

Each Voice directory README includes:
- Purpose section
- Sources section
- Structure section
- Verbatim template
- Need template (or Constraint template for VoE)

### Requirements README Template

Includes:
- Purpose section
- Workflow section
- Requirement template
- Rules (linkage requirements, priority derivation)

### Matrices README Template

Includes:
- Purpose section
- Matrix types description
- Priority matrix template
- Conflict matrix template
- House of Quality template

## Acceptance Criteria

- [ ] All Voice directories contain `README.md`, `verbatims/`, `needs/`
- [ ] VoE additionally contains `constraints/`
- [ ] Directory `requirements/` exists with `README.md`
- [ ] Directory `matrices/` exists with `README.md`
- [ ] Main `README.md` contains structure overview, conventions, workflow
- [ ] All templates are usable (copy-paste ready)
- [ ] Wikilinks use consistent format `[[name]]`
- [ ] `--template` flag works with `qfd` (default) and `basic` options
- [ ] Project creation without `--template` flag uses QFD structure by default
- [ ] Templates stored as embedded files in `helpers/ptx-pft/templates/qfd/`
- [ ] Templates easily editable without code changes (separate .tmpl files)
- [ ] `ptx-pft info` returns documentation for all methodologies (basic, qfd)
- [ ] `ptx-mcp` tool `pft_info` calls ptx-pft and returns methodology guide

## Technical Notes

- **Empty directories:** Use `.gitkeep` for version control
- **Encoding:** UTF-8
- **Line endings:** Unix (LF)
- **Obsidian compatibility:** Wikilinks in `[[name]]` format without extension

## Embedded Templates Architecture

Templates are stored as **embedded files** (not hardcoded in Go code) for easy modification without recompilation.

### Template Storage Location

```
helpers/ptx-pft/
├── templates/
│   └── qfd/
│       ├── README.md.tmpl              # Main project README
│       ├── VoC-README.md.tmpl          # Voice of Customer README
│       ├── VoB-README.md.tmpl          # Voice of Business README
│       ├── VoE-README.md.tmpl          # Voice of Engineering README
│       ├── VoS-README.md.tmpl          # Voice of Stakeholder README
│       ├── requirements-README.md.tmpl # Requirements README
│       └── matrices-README.md.tmpl     # Matrices README
└── embed.go                            # Go embed directive
```

### Go Embed Integration

```go
//go:embed templates/*
var templateFS embed.FS
```

### Template Variable Substitution

Templates use Go text/template syntax for variable substitution:
- `{{.ProjectName}}` - Project name from command argument

---

## Complete Template Files

### Template: README.md.tmpl (Main Project)

```markdown
# {{.ProjectName}}

## Overview

This project uses QFD (Quality Function Deployment) methodology per ISO 16355
for systematic requirements management from voice of customer collection
through technical specification.

## Project Structure

| Directory | Purpose |
|-----------|---------|
| VoC/ | Voice of Customer – customer needs |
| VoB/ | Voice of Business – business goals and constraints |
| VoE/ | Voice of Engineering – technical capabilities and constraints |
| VoS/ | Voice of Stakeholder – regulations, partners, internal stakeholders |
| requirements/ | Technical requirements (intersection of all Voices) |
| matrices/ | QFD matrices and prioritization |

## Naming Conventions

### File Prefixes

| Voice | Verbatim | Need |
|-------|----------|------|
| VoC | VC-V001 | VC-P01, VC-S01, VC-T01 |
| VoB | VB-V001 | VB-P01, VB-S01, VB-T01 |
| VoE | VE-V001 | VE-C01 (constraints) |
| VoS | VS-V001 | VS-P01, VS-S01, VS-T01 |
| Requirements | – | R001, R002, ... |

### Need Levels (ISO 16355)

| Level | Prefix | Description |
|-------|--------|-------------|
| Primary | P | Strategic need (highest abstraction) |
| Secondary | S | Tactical need |
| Tertiary | T | Specific, measurable need |

### Maturity Statuses

| Status | Meaning |
|--------|---------|
| Captured | Raw input recorded |
| Structured | Placed in hierarchy |
| Analyzed | Context and motivation understood |
| Prioritized | Importance determined (AHP, Kano) |
| Translated | Converted to technical requirement |
| Deployed | Handed off to implementation |

### Kano Categories

| Category | Description |
|----------|-------------|
| Must-be | Basic requirement (non-fulfillment = dissatisfaction) |
| Performance | Linear relationship between fulfillment and satisfaction |
| Delighter | Surprise (unexpected value) |

## Workflow

```
Verbatim → Reworded Need → Customer Requirement → Quality Characteristic → Technical Requirement
```

1. **Acquisition** – Capture verbatim into `VoX/verbatims/`
2. **Structuring** – Create need in `VoX/needs/` linked to verbatim
3. **Analysis** – Add context, Kano category
4. **Prioritization** – AHP score, update `matrices/`
5. **Translation** – Create technical requirement in `requirements/`

## Tools

- **Editor:** MarkText / Obsidian / VS Code with Markdown preview
- **Version Control:** Git
- **Wikilinks:** `[[VC-V001]]` for inter-file navigation
```

---

### Template: VoC-README.md.tmpl

```markdown
# Voice of Customer (VoC)

## Purpose

Evidence of customer voice – direct and indirect inputs from end users
and product customers.

## VoC Sources

- Customer interviews
- Customer support (tickets, emails)
- User testing
- Feedback tools (Fider)
- Reviews and ratings
- Competitive analysis (what customers praise/criticize)

## Structure

### verbatims/
Raw, literal customer statements. Do not edit, preserve authentic language.

### needs/
Structured needs derived from verbatims:
- **P** (Primary) – strategic needs
- **S** (Secondary) – tactical needs
- **T** (Tertiary) – specific, measurable needs

## Template: Verbatim

File: `verbatims/VC-V001-name.md`

```
# VC-V001 – [Brief Description]

## Metadata
- **Source:** [Name, role, segment]
- **Date:** YYYY-MM-DD
- **Channel:** [Interview / Support / Fider / ...]
- **Segment:** [Enterprise / SMB / Consumer / ...]

## Verbatim
> "[Literal customer quote]"

## Context
[Circumstances in which the statement was made]

## Derived Needs
- [[VC-S01-name]]
- [[VC-T01-name]]
```

## Template: Need

File: `needs/VC-S01-name.md`

```
# VC-S01 – [Need Name]

## Metadata
- **Level:** [Primary / Secondary / Tertiary]
- **Parent:** [[VC-P01-name]] or –
- **Kano:** [Must-be / Performance / Delighter]
- **AHP Priority:** [0.00 - 1.00]
- **Status:** [Captured / Structured / Analyzed / Prioritized / Translated]

## Description
[Reworded need in customer language]

## Sources (Verbatims)
- [[VC-V001-name]]

## Children
- [[VC-T01-name]]

## Technical Requirements
- [[R001-name]]

## Conflicts
- [[VB-P01-name]] – [conflict description]
```
```

---

### Template: VoB-README.md.tmpl

```markdown
# Voice of Business (VoB)

## Purpose

Evidence of business goals, strategic priorities, and constraints from
the organization's perspective.

## VoB Sources

- Strategic documents
- OKR / KPI
- Financial plans and budgets
- Management interviews
- Competitive analysis
- Market data

## Structure

### verbatims/
Management statements, strategic assignments, business requirements.

### needs/
Structured business needs:
- **P** (Primary) – strategic goals
- **S** (Secondary) – tactical goals
- **T** (Tertiary) – specific metrics

## Template: Verbatim

File: `verbatims/VB-V001-name.md`

```
# VB-V001 – [Brief Description]

## Metadata
- **Source:** [Name, role]
- **Date:** YYYY-MM-DD
- **Context:** [Strategic meeting / OKR review / ...]

## Verbatim
> "[Literal quote]"

## Derived Needs
- [[VB-P01-name]]
```

## Template: Need

File: `needs/VB-P01-name.md`

```
# VB-P01 – [Need Name]

## Metadata
- **Level:** [Primary / Secondary / Tertiary]
- **Parent:** [[VB-P01-name]] or –
- **Priority:** [High / Medium / Low]
- **Status:** [Captured / Structured / Analyzed / Prioritized / Translated]

## Description
[Business need]

## Success Metrics
- [KPI 1]
- [KPI 2]

## Sources
- [[VB-V001-name]]

## Conflicts with VoC
- [[VC-P01-name]] – [conflict description]
```
```

---

### Template: VoE-README.md.tmpl

```markdown
# Voice of Engineering (VoE)

## Purpose

Evidence of technical capabilities, constraints, dependencies, and best practices
from the development team's perspective.

## VoE Sources

- Architecture documentation
- Technical debt (backlog)
- External system dependencies
- Capacity constraints
- Security requirements
- Standards and best practices

## Structure

### verbatims/
Technical inputs, developer opinions, architectural decisions.

### needs/
Technical needs and infrastructure requirements.

### constraints/
Technical constraints and dependencies affecting possible solutions.

## Template: Verbatim

File: `verbatims/VE-V001-name.md`

```
# VE-V001 – [Brief Description]

## Metadata
- **Source:** [Name, role]
- **Date:** YYYY-MM-DD
- **Context:** [Tech review / Architectural decision / ...]

## Verbatim
> "[Literal quote]"

## Derived Constraints
- [[VE-C01-name]]
```

## Template: Constraint

File: `constraints/VE-C01-name.md`

```
# VE-C01 – [Constraint Name]

## Metadata
- **Type:** [Technology / Capacity / Security / Dependency]
- **Severity:** [Blocker / Major / Minor]
- **Status:** [Active / Resolved / Accepted]

## Description
[Technical constraint description]

## Impact
[How the constraint affects possible solutions]

## Workaround
[Possible workaround, if any]

## Sources
- [[VE-V001-name]]

## Affected Requirements
- [[R001-name]]
```
```

---

### Template: VoS-README.md.tmpl

```markdown
# Voice of Stakeholder (VoS)

## Purpose

Evidence of requirements and constraints from external and internal stakeholders
beyond direct customers and business.

## VoS Sources

- Regulatory requirements (GDPR, HIPAA, ...)
- Partner agreements and SLAs
- Internal stakeholders (HR, Legal, Finance)
- Industry standards
- Certifications

## Structure

### verbatims/
Stakeholder statements, regulatory texts, contractual requirements.

### needs/
Structured stakeholder requirements.

## Template: Verbatim

File: `verbatims/VS-V001-name.md`

```
# VS-V001 – [Brief Description]

## Metadata
- **Source:** [Stakeholder, organization]
- **Date:** YYYY-MM-DD
- **Type:** [Regulation / Partner / Internal / ...]

## Verbatim
> "[Literal quote or document reference]"

## Reference
[Link to full regulation/contract text]

## Derived Needs
- [[VS-P01-name]]
```

## Template: Need

File: `needs/VS-P01-name.md`

```
# VS-P01 – [Need Name]

## Metadata
- **Level:** [Primary / Secondary / Tertiary]
- **Parent:** [[VS-P01-name]] or –
- **Type:** [Compliance / SLA / Internal]
- **Obligation:** [Mandatory / Recommended / Optional]
- **Status:** [Captured / Structured / Analyzed / Prioritized / Translated]

## Description
[Stakeholder requirement]

## Deadline / Validity
[Date by which it must be fulfilled]

## Non-compliance Consequences
[Consequences of non-fulfillment]

## Sources
- [[VS-V001-name]]

## Technical Requirements
- [[R001-name]]
```
```

---

### Template: requirements-README.md.tmpl

```markdown
# Technical Requirements

## Purpose

Evidence of technical requirements derived from all Voices' needs. Each requirement
must be linked to at least one need.

## Workflow

```
VoC/VoB/VoE/VoS Needs → Translation → Technical Requirement → Implementation
```

## Template: Requirement

File: `R001-name.md`

```
# R001 – [Requirement Name]

## Metadata
- **Status:** [Draft / Ready / In Progress / Done / Cancelled]
- **Priority:** [Critical / High / Medium / Low]
- **Effort:** [XS / S / M / L / XL]
- **Sprint/Release:** [reference]

## Description
[Technical requirement description]

## Addresses Needs
- [[VC-S01-name]] (VoC)
- [[VB-P01-name]] (VoB)

## Constraints
- [[VE-C01-name]]

## Acceptance Criteria
- [ ] [Criterion 1]
- [ ] [Criterion 2]
- [ ] [Criterion 3]

## Technical Notes
[Solution design, architecture, dependencies]

## Dependencies
- **Blocks:** [[R002-name]]
- **Depends on:** [[R003-name]]
```

## Rules

1. Every requirement must link to at least one Need
2. Requirement without linkage = candidate for removal
3. Priority is derived from AHP scores of linked Needs
4. "Ready" status requires defined acceptance criteria
```

---

### Template: matrices-README.md.tmpl

```markdown
# QFD Matrices and Prioritization

## Purpose

Visualization of relationships between Voices, needs, and technical requirements.
Support for prioritization and decision-making.

## Matrix Types

### House of Quality (HoQ)
Mapping customer needs to technical characteristics.

### Priority Matrix
Aggregated priorities across all Voices.

### Conflict Matrix
Identification of conflicts between different Voices.

## Files

### priority-matrix.md

```
# Priority Matrix

| Requirement | VoC | VoB | VoE | VoS | Total | Rank |
|-------------|-----|-----|-----|-----|-------|------|
| [[R001]] | 0.8 | 0.6 | 0.9 | 0.5 | 0.70 | 1 |
| [[R002]] | 0.5 | 0.9 | 0.7 | 0.3 | 0.60 | 2 |

## Methodology
- Voice weights: VoC 40%, VoB 30%, VoE 20%, VoS 10%
- Total = weighted average
```

### conflict-matrix.md

```
# Conflict Matrix

| Need A | Need B | Conflict Type | Resolution |
|--------|--------|---------------|------------|
| [[VC-P01]] | [[VB-P02]] | Resource | VoC prioritization |
| [[VC-S03]] | [[VE-C01]] | Technical | Workaround |
```

### hoq-voc-requirements.md

```
# House of Quality: VoC → Requirements

|  | R001 | R002 | R003 | Priority |
|--|------|------|------|----------|
| VC-P01 | ●●● | ● | – | 0.85 |
| VC-S01 | ●● | ●●● | ● | 0.72 |
| VC-T01 | ● | – | ●●● | 0.65 |

## Legend
- ●●● Strong relationship (9)
- ●● Medium relationship (3)
- ● Weak relationship (1)
- – No relationship (0)
```
```

---

## References

- ISO 16355 - Quality Function Deployment
- Existing PTX-PFT implementation: `helpers/ptx-pft/`
- Related issue: #107 (PTX-PFT initial implementation)
- Related issue: #112 (PTX-PFT Category Management)
