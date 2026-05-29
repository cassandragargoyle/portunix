---
title: Capability Registry
description: Source-of-truth registry of software product capabilities as YAML nodes and relations. Generated and maintained via the /capability-registry skill.
category: reference
status: active
language: en
---

# Capability Registry

Machine-readable graph of **what software products do, under what conditions,
through which interfaces, over which entities, and how they depend on each
other**. This directory is the **source of truth** — `nodes/` + `relations/`
are primary; everything else is derivable.

Authoring is driven by the
[`/capability-registry`](../../contributing/capability-registry/INDEX.md) skill, which loads the router + per-type template + schema rather than the full
[methodology](../../contributing/CAPABILITY-REGISTRY-METHODOLOGY.md).

## Layout

```text
capability-registry/
├── _vocabulary/
│   └── domains.yaml      controlled vocabulary of capability/entity domains
├── nodes/
│   ├── products/         product:*
│   ├── modules/          module:*   (only when parent product.composition=modular)
│   ├── capabilities/     capability:*
│   ├── operations/       operation:*
│   ├── entities/         entity:*
│   ├── interfaces/       interface:*
│   └── availability/     availability:*  (standalone nodes only)
├── relations/
│   ├── provides.yaml
│   ├── requires.yaml
│   ├── implements.yaml
│   ├── operates-on.yaml
│   ├── exposed-via.yaml
│   ├── maps-to.yaml
│   └── constrained-by.yaml
└── tools/
    └── README.md         schema validation with ajv-cli
```

## This project

Portunix itself is registered as `product:portunix`. See the skill's "This project (Portunix) as a product" section for the canonical product node,
capability map, CLI interface, and founding relations.
