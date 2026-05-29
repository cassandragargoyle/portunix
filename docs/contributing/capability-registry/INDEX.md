---
title: Capability Registry — Router
description: Machine-readable router for authoring individual Capability Registry files. Carries the per-type template + schema + target path so the full methodology need not be loaded. Load CAPABILITY-REGISTRY-METHODOLOGY.md only for onboarding or convention changes.
category: reference
ai_load: router
status: active
language: en
---

# Capability Registry — Router (INDEX)

Use this router to author a **single** registry file. It is ~50 lines and replaces
the ~300-line methodology for routine record creation. Load the full
[`CAPABILITY-REGISTRY-METHODOLOGY.md`](../CAPABILITY-REGISTRY-METHODOLOGY.md) only for
onboarding or when changing the conventions / ID grammar / node & relation types.

All paths below are relative to this directory (`docs/contributing/capability-registry/`)
for templates/schemas, and to `docs/architecture/capability-registry/` for output.

## Per-type table

| Type | Template | Schema | Output path |
| ---- | -------- | ------ | ----------- |
| `product` | `templates/product.template.yaml` | `schemas/product.schema.json` | `nodes/products/<slug>.yaml` |
| `module` | `templates/module.template.yaml` | `schemas/module.schema.json` | `nodes/modules/<product>.<name>.yaml` |
| `capability` | `templates/capability.template.yaml` | `schemas/capability.schema.json` | `nodes/capabilities/<domain>.<name>.yaml` |
| `operation` | `templates/operation.template.yaml` | `schemas/operation.schema.json` | `nodes/operations/<product>.<entity>.<action>.yaml` |
| `entity` | `templates/entity.template.yaml` | `schemas/entity.schema.json` | `nodes/entities/<domain>.<name>.yaml` |
| `interface` | `templates/interface.template.yaml` | `schemas/interface.schema.json` | `nodes/interfaces/<product>.<name>.yaml` |
| `availability` | `templates/availability.template.yaml` | `schemas/availability.schema.json` | `nodes/availability/<product>.<capability-ns>.yaml` |
| `relation` | `templates/relation.template.yaml` | `schemas/relation.schema.json` | append to `relations/<type>.yaml` |

**Always also load** the shared `schemas/verification.schema.json` (referenced by every
schema above). For a `provides` relation with an availability block, also load
`schemas/availability.schema.json`.

`relation` target file by kind: `provides` → `relations/provides.yaml`,
`requires` → `relations/requires.yaml`, `implements` → `relations/implements.yaml`,
`operates_on` → `relations/operates-on.yaml`, `exposed_via` → `relations/exposed-via.yaml`,
`maps_to` → `relations/maps-to.yaml`, `constrained_by` → `relations/constrained-by.yaml`.

## Rule files (load when in doubt)

| Question | File |
| -------- | ---- |
| What should the ID look like? Which regex applies? | [`id-grammar.md`](id-grammar.md) |
| Availability inline on a relation, or a standalone node? | [`decision-availability.md`](decision-availability.md) |

## Controlled vocabulary

Before creating `capability:<domain>.<name>` or `entity:<domain>.<name>`, verify
`<domain>` exists in
[`../../architecture/capability-registry/_vocabulary/domains.yaml`](../../architecture/capability-registry/_vocabulary/domains.yaml).
A new domain is added in the same PR (slug, name, description).

## Validation

See [`../../architecture/capability-registry/tools/README.md`](../../architecture/capability-registry/tools/README.md).
