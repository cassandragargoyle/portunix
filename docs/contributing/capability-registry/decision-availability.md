---
title: Capability Registry — Availability Placement Decision
description: Decision rule for storing availability inline on a provides relation versus as a standalone availability node.
category: reference
ai_load: scoped
status: active
language: en
---

# Availability — inline vs. standalone node

A capability is **not a boolean**: it is available under conditions (license,
version, edition, region, roles, …). Those conditions live in an `availability`
block. Where the block lives is the decision below.

## Default: inline on the `provides` relation

Store the `availability` block **inline** on the `provides` edge in
`relations/provides.yaml`. This is correct for the overwhelming majority of cases:

```yaml
- from: product:portunix
  type: provides
  to: capability:provisioning.package-install
  availability:
    status: active
    license: { required: false }
    version: { min: "1.0", max: null }
  verification: { status: documented, lastChecked: "2026-05-26" }
```

## Promote to a standalone node only when ALL apply

Create `nodes/availability/<product>.<capability-ns>.yaml` and a
`constrained_by` relation **only** when the availability description is itself a
first-class, reusable object — i.e. when at least one of these holds:

1. **Shared across many edges** — the same availability profile constrains
   several `provides` relations and you want one source of truth.
2. **Large / complex** — the block carries many dimensions (feature flags,
   per-region limits, deployment matrix) that would bloat the relation file.
3. **Independently versioned / referenced** — other documents or tools need to
   link to the availability object by its own ID.

If none apply, keep it **inline**. Do not create standalone nodes pre-emptively.

## Migration inline → standalone

1. Move the block into `nodes/availability/<product>.<capability-ns>.yaml`, add
   `id:` and `appliesTo:`.
2. Replace the inline block on the relation with a `constrained_by` edge in
   `relations/constrained-by.yaml` pointing at the new node.
3. Keep `verification` on both the node and the relation.
