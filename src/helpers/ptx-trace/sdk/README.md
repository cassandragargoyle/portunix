# PTX-TRACE SDKs

This directory contains official SDKs for the PTX-TRACE system - a universal tracing system for software development, optimized for debugging, AI analysis, and monitoring of data pipelines and workflows.

## Available SDKs

| Language | Directory | Status | Min Version |
| -------- | --------- | ------ | ----------- |
| **Python** | [`python/`](python/) | Stable | Python 3.8+ |
| **Java** | [`java/`](java/) | Stable | Java 21+ |
| **TypeScript** | [`typescript/`](typescript/) | Stable | Node.js 16+ |
| **Go** | Built-in | Stable | Go 1.21+ |

## Architecture

All external SDKs communicate with the PTX-TRACE system via CLI subprocess calls:

```text
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Python SDK    │     │    Java SDK     │     │ TypeScript SDK  │
│                 │     │                 │     │                 │
│  Session        │     │  Session        │     │  Session        │
│  Operation      │     │  Operation      │     │  Operation      │
│  OperationBuilder     │  OperationBuilder     │  OperationBuilder
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         │  subprocess/exec      │  ProcessBuilder       │  spawn
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                     portunix trace CLI                          │
│                                                                 │
│  start | end | event | sessions | view | stats | query | ...    │
└─────────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────────┐
│                     PTX-TRACE Storage                           │
│                                                                 │
│  NDJSON files | SQLite index | Session metadata                 │
└─────────────────────────────────────────────────────────────────┘
```

### Why CLI-based Communication?

1. **Consistency**: All SDKs use the same validated CLI commands
2. **Simplicity**: No need for native bindings or complex serialization
3. **Portability**: Works anywhere `portunix` binary is available
4. **Maintainability**: Single source of truth for business logic
5. **Security**: PII masking and validation happen in Go code

## Quick Comparison

### Session Creation

**Python:**

```python
from ptx_trace import Session

with Session("my-session", pii_masking=True) as session:
    # operations...
```

**Java:**

```java
try (Session session = Trace.newSession("my-session")
        .withPIIMasking(true)
        .build()) {
    // operations...
}
```

**TypeScript:**

```typescript
import { Session } from 'ptx-trace';

const session = await Session.create("my-session", {
  piiMasking: true
});
try {
    // operations...
} finally {
    await session.close();
}
```

### Traced Operations

**Python:**

```python
with session.operation("normalize_phone") as op:
    op.input("phone", raw_phone)
    result = normalize(raw_phone)
    op.output("phone", result)
    op.success()
```

**Java:**

```java
try (Operation op = session.start("normalize_phone")) {
    op.input("phone", rawPhone);
    String result = normalize(rawPhone);
    op.output("phone", result);
    op.success();
}
```

**TypeScript:**

```typescript
await session.withOperation("normalize_phone", async (op) => {
    op.input("phone", rawPhone);
    const result = await normalize(rawPhone);
    op.output("phone", result);
});
```

### Fluent API

**Python:**

```python
session.trace("validate_email") \
    .input("email", email) \
    .tag("validation") \
    .execute(lambda ctx: ctx.output("valid", validate(email)))
```

**Java:**

```java
session.trace("validate_email")
    .input("email", email)
    .tag("validation")
    .execute(ctx -> ctx.output("valid", validate(email)));
```

**TypeScript:**

```typescript
await session.trace("validate_email")
    .input("email", email)
    .tag("validation")
    .execute(async (ctx) => {
        ctx.output("valid", await validate(email));
    });
```

## Installation

### Python

```bash
cd sdk/python
pip install -e .
```

### Java

```bash
cd sdk/java
mvn clean package
```

Add to your `pom.xml`:

```xml
<dependency>
    <groupId>ai.portunix</groupId>
    <artifactId>ptx-trace</artifactId>
    <version>1.0.0</version>
</dependency>
```

### TypeScript

```bash
cd sdk/typescript
npm install
npm run build
```

## Prerequisites

All SDKs require the `portunix` binary to be available:

1. **In PATH**: `portunix` command available globally
2. **Custom path**: Specify path in SDK configuration

```python
# Python
Session("my-session", binary_path="/opt/portunix/bin/portunix")

# Java
Trace.setDefaultBinaryPath("/opt/portunix/bin/portunix");

# TypeScript
Session.create("my-session", { binaryPath: "/opt/portunix/bin/portunix" });
```

## Examples

See the [`examples/`](examples/) directory for complete working examples:

- [`examples/python/etl_pipeline.py`](examples/python/etl_pipeline.py) - Python ETL example
- [`examples/java/EtlPipeline.java`](examples/java/EtlPipeline.java) - Java ETL example
- [`examples/typescript/etl-pipeline.ts`](examples/typescript/etl-pipeline.ts) - TypeScript ETL example

## CLI Commands Used

| SDK Method | CLI Command |
| ---------- | ----------- |
| `Session.create()` | `portunix trace start <name>` |
| `Session.close()` | `portunix trace end` |
| `Operation.end()` | `portunix trace event <operation>` |
| `Session.list()` | `portunix trace sessions --format json` |
| `Session.stats()` | `portunix trace stats --format json` |
| `Session.events()` | `portunix trace view --format json` |
| `Session.query()` | `portunix trace query --format json` |

## Features by SDK

| Feature | Python | Java | TypeScript | Go |
| ------- | ------ | ---- | ---------- | --- |
| Session management | ✅ | ✅ | ✅ | ✅ |
| Context manager / try-with-resources | ✅ | ✅ | N/A | ✅ |
| Fluent API | ✅ | ✅ | ✅ | ✅ |
| Decorator / Annotation | ✅ | ⚠️* | N/A | N/A |
| Async/await | N/A | N/A | ✅ | ✅ |
| Source types | ✅ | ✅ | ✅ | ✅ |
| Error handling | ✅ | ✅ | ✅ | ✅ |
| PII masking | ✅ | ✅ | ✅ | ✅ |
| Sampling | ✅ | ✅ | ✅ | ✅ |

*Java annotation requires AspectJ or similar AOP framework

## Thread Safety

**Important**: Session objects are NOT thread-safe. For concurrent operations:

- Create separate sessions per thread/task
- Or use external synchronization

## Documentation

- [Python SDK README](python/README.md)
- [Java SDK README](java/README.md)
- [TypeScript SDK README](typescript/README.md)
- [Go SDK](../src/helpers/ptx-trace/sdk/) (built into ptx-trace helper)

## Support

For issues and feature requests:

- GitHub Issues: https://github.com/cassandragargoyle/portunix/issues

## License

MIT License - see LICENSE file for details.
