---
title: Brainstorming Methodology
description: Methodology for brainstorming sessions with an AI assistant — JSON verbatim journal, session flow, artifacts, integration with use cases, ADRs, and PFT workflows. Apply when running a brainstorming session via /brainstorm or producing materials in docs/architecture/brainstorming/.
category: methodology
ai_load: scoped
status: active
language: en
created: 2026-03-22
last_updated: 2026-05-16
related:
  - USE-CASE-METHODOLOGY.md
  - MARKDOWN-FRONTMATTER.md
---

# Brainstorming Methodology

## Overview

This document defines how brainstorming sessions are conducted with an AI assistant
(Claude Code) within the CassandraGargoyle / Portunix ecosystem. The goal is to capture
raw ideas, discussions, and decisions in a structured way that feeds into the existing
use case and PFT workflows.

## Session Structure

### 1. Activation

A brainstorming session is started by invoking the `/brainstorm` skill with the topic
or a reference to an existing brainstorming directory.

### 2. Context Loading

When a brainstorming session starts, the assistant MUST load relevant architectural
context before engaging:

- **This document** — `docs/contributing/BRAINSTORMING.md`
- **Use Case Methodology** — `docs/contributing/USE-CASE-METHODOLOGY.md`
- **Architecture Overview** — `docs/architecture/README.md`
- **Component Index** — `docs/architecture/components/README.md`
- **Existing brainstorming materials** in the topic directory (if any)

For architecture-focused brainstorming, the assistant should also load relevant
specifications and ADRs from `docs/architecture/`.

### 3. Verbatim Journal (JSON)

During every brainstorming session, the assistant creates and maintains a **JSON journal
file** that records the raw conversation verbatims. This file preserves the unedited
flow of ideas as they emerge.

#### File Location

```text
docs/architecture/brainstorming/<topic>/journal-<YYYYMMDD>-<seq>.json
```

- `<topic>` — the brainstorming topic directory (e.g., `config-packages`)
- `<YYYYMMDD>` — session date
- `<seq>` — sequential number if multiple sessions occur on the same day (`01`, `02`, ...)

#### JSON Schema

```json
{
  "session": {
    "id": "BS-<TOPIC>-<YYYYMMDD>-<SEQ>",
    "topic": "<topic name>",
    "date": "<YYYY-MM-DD>",
    "participants": ["<name>"],
    "medium": "claude-code",
    "status": "active | completed | paused",
    "related_docs": ["<path to related document>"]
  },
  "verbatims": [
    {
      "seq": 1,
      "timestamp": "<ISO 8601 or HH:MM>",
      "speaker": "<name or role>",
      "type": "idea | question | decision | concern | reference | clarification",
      "text": "<raw verbatim text — NEVER edited after recording>",
      "tags": ["<tag>"],
      "refs": ["<reference to doc, component, or external source>"]
    }
  ],
  "decisions": [
    {
      "seq": 1,
      "summary": "<short decision summary in English>",
      "rationale": "<why this was decided>",
      "verbatim_refs": [3, 5],
      "status": "proposed | accepted | rejected | deferred"
    }
  ],
  "open_questions": [
    {
      "seq": 1,
      "question": "<open question>",
      "verbatim_refs": [2],
      "status": "open | resolved | deferred"
    }
  ],
  "next_steps": [
    "<action item>"
  ]
}
```

#### Verbatim Rules

1. **Immutable** — once a verbatim entry is written, it is NEVER modified
2. **Complete** — every substantive statement from both user and assistant is recorded
3. **Unedited** — preserve original wording, including the user's language (Czech, English, mixed)
4. **Attributed** — each entry identifies the speaker
5. **Typed** — each entry is classified (idea, question, decision, concern, reference, clarification)

### 4. Session Flow

A typical brainstorming session follows this pattern:

1. **Load context** — assistant reads relevant documents and existing materials
2. **Review existing state** — if resuming, review previous journal entries
3. **Explore** — free-form discussion, ideas, questions, challenges
4. **Capture** — assistant continuously records verbatims to the journal
5. **Synthesize** — periodically summarize decisions and open questions
6. **Close** — finalize the journal, update status, list next steps

### 5. Artifacts

A brainstorming topic directory contains:

```text
docs/architecture/brainstorming/<topic>/
├── README.md                              # Topic overview and current status
├── journal-<YYYYMMDD>-<SEQ>.json          # Session verbatim journals
├── <YYYYMMDD>-<source>-<description>.md   # External inputs (ChatGPT logs, notes, etc.)
├── UC-<ID>-<slug>.md                      # Use cases emerging from brainstorming
├── UC-<ID>-<slug>-sequence.puml           # PlantUML sequence diagrams for use cases
├── UC-<ID>-<slug>-use-case.puml           # PlantUML use case diagrams (optional)
└── summary.md                             # Consolidated findings (when topic matures)
```

### 5a. Use Cases and Diagrams

When a use case emerges during brainstorming, create both:

1. **Use case document** (`UC-<ID>-<slug>.md`) — following the template from
   [USE-CASE-METHODOLOGY.md](USE-CASE-METHODOLOGY.md)
2. **PlantUML sequence diagram** (`UC-<ID>-<slug>-sequence.puml`) — visualizing
   the basic flow and alternative/exception flows

The use case document MUST reference its PlantUML diagram(s). Do not embed ASCII
sequence diagrams in the use case document — use PlantUML files instead.

Additional diagram types (use case diagram, activity diagram, component diagram)
are optional and should be created when they add clarity.

### 6. Integration with Other Workflows

Brainstorming outputs feed into:

| Output | Target Workflow | Tool |
| ------ | --------------- | ---- |
| Validated ideas | PFT verbatim | `/add-pft-idea` |
| Formalized requirements | PFT requirement | `/add-pft-requirement` |
| Defined use cases | Use Case Methodology | `/add-pft-usecase` |
| Architecture decisions | ADR | Manual ADR creation in `docs/adr/` |
| Technical specs | Specifications | `docs/architecture/specifications/` |

## Guidelines for the Assistant

### During Brainstorming

- **Be an active participant** — propose alternatives, challenge assumptions, identify gaps
- **Record continuously** — do not wait for the end of the session to record verbatims
- **Preserve language** — record the user's words in their original language
- **Tag effectively** — use tags to make verbatims searchable later
- **Track decisions** — whenever the user makes a decision, record it in the decisions array
- **Surface open questions** — explicitly list unresolved questions

### Architecture Focus

Since `portunix-architecture` is primarily an architecture project, brainstorming
sessions typically involve:

- Component design and interaction patterns
- API contracts and protocol decisions
- Technology selection and trade-offs
- Deployment and operational concerns
- Cross-project integration (CassandraGargoyle ecosystem)

The assistant should bring architectural thinking to the discussion — proposing
variants, identifying trade-offs (latency, cost, complexity), and referencing
relevant patterns and prior art.

### What NOT to Do

- Do not edit or "clean up" recorded verbatims
- Do not skip recording because something seems unimportant
- Do not make architectural decisions without presenting alternatives
- Do not ignore existing materials in the topic directory
