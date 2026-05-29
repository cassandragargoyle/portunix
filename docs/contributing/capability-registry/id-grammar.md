---
title: Capability Registry — ID Grammar
description: Canonical ID grammar and regexes for Capability Registry node and relation identifiers.
category: reference
ai_load: scoped
status: active
language: en
---

# Capability Registry — ID Grammar

## Slug

A **slug** is the atomic building block of every ID level:

```text
slug   = [a-z][a-z0-9-]*
```

- lowercase ASCII only — no uppercase, no diacritics, no underscores
- a hyphen separates **words within one level** (`lead-management`)
- a dot separates **hierarchical levels** (`crm.lead-management`)

## Per-type identifiers

| Type | Pattern | Regex |
| ---- | ------- | ----- |
| product | `product:<slug>` | `^product:[a-z][a-z0-9-]*$` |
| module | `module:<product>.<name>` | `^module:[a-z][a-z0-9-]*\.[a-z][a-z0-9-]*$` |
| capability | `capability:<domain>.<name>` (max 4 levels) | `^capability:[a-z][a-z0-9-]*(\.[a-z][a-z0-9-]*){1,3}$` |
| operation | `operation:<product>.<entity>.<action>` | `^operation:[a-z][a-z0-9-]*\.[a-z][a-z0-9-]*\.[a-z][a-z0-9-]*$` |
| entity | `entity:<domain>.<name>` | `^entity:[a-z][a-z0-9-]*\.[a-z][a-z0-9-]*$` |
| interface | `interface:<product>.<name>` | `^interface:[a-z][a-z0-9-]*\.[a-z][a-z0-9-]*$` |
| availability | `availability:<product>.<capability-ns>` | `^availability:[a-z][a-z0-9-]*\.[a-z][a-z0-9-]*(\.[a-z][a-z0-9-]*){0,2}$` |

## Rules that are easy to break

1. A **capability** is named by meaning, not by product:
   `capability:provisioning.package-install`, **not** `capability:portunix.install`.
2. An **operation** is always product-specific:
   `operation:portunix.package.install`.
3. The `<domain>` in a capability/entity ID must exist in
   `_vocabulary/domains.yaml`.
4. No underscores anywhere in IDs (the relation `type` field uses underscores
   like `operates_on`, but those are not IDs).
