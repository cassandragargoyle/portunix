---
title: Capability Registry — Validation Tools
description: How to validate Capability Registry YAML nodes and relations against the JSON schemas with ajv-cli.
category: reference
status: active
language: en
---

# Capability Registry — Validation

Schemas live in
[`docs/contributing/capability-registry/schemas/`](../../../contributing/capability-registry/schemas/).
Every schema references `verification.schema.json`; the availability/relation schemas additionally reference `availability.schema.json`.

## Install ajv-cli (optional)

```bash
npm install -g ajv-cli ajv-formats
```

`ajv-formats` is required because the `verification.lastChecked` field uses
`format: date`.

## Validate a node

```bash
SCHEMAS=docs/contributing/capability-registry/schemas
ajv validate --spec=draft7 -c ajv-formats \
  -s $SCHEMAS/capability.schema.json \
  -r $SCHEMAS/verification.schema.json \
  -d docs/architecture/capability-registry/nodes/capabilities/provisioning.package-install.yaml
```

Swap the `-s` schema for the node type you are validating
(`product`, `module`, `capability`, `operation`, `entity`, `interface`, `availability`).

## Validate a relations file

A relations file needs both referenced schemas:

```bash
SCHEMAS=docs/contributing/capability-registry/schemas
ajv validate --spec=draft7 -c ajv-formats \
  -s $SCHEMAS/relation.schema.json \
  -r $SCHEMAS/availability.schema.json \
  -r $SCHEMAS/verification.schema.json \
  -d docs/architecture/capability-registry/relations/provides.yaml
```

## Validate everything

```bash
SCHEMAS=docs/contributing/capability-registry/schemas
NODES=docs/architecture/capability-registry/nodes
ajv validate --spec=draft7 -c ajv-formats -s $SCHEMAS/product.schema.json    -r $SCHEMAS/verification.schema.json -d "$NODES/products/*.yaml"
ajv validate --spec=draft7 -c ajv-formats -s $SCHEMAS/capability.schema.json -r $SCHEMAS/verification.schema.json -d "$NODES/capabilities/*.yaml"
ajv validate --spec=draft7 -c ajv-formats -s $SCHEMAS/operation.schema.json  -r $SCHEMAS/verification.schema.json -d "$NODES/operations/*.yaml"
ajv validate --spec=draft7 -c ajv-formats -s $SCHEMAS/entity.schema.json     -r $SCHEMAS/verification.schema.json -d "$NODES/entities/*.yaml"
ajv validate --spec=draft7 -c ajv-formats -s $SCHEMAS/interface.schema.json  -r $SCHEMAS/verification.schema.json -d "$NODES/interfaces/*.yaml"
```

If `ajv` is not installed, schema validation is skipped — node creation still works; just say so rather than installing it unprompted.
