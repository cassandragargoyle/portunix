# PTX-TRACE TypeScript SDK

TypeScript/JavaScript SDK for the Portunix Trace (PTX-TRACE) system - a universal tracing system for software development, optimized for debugging, AI analysis, and monitoring of data pipelines and workflows.

## Requirements

- Node.js 16+
- `portunix` binary in PATH (or specify path)

## Installation

```bash
# From npm (when published)
npm install ptx-trace

# From source
cd sdk/typescript
npm install
npm run build
```

## Quick Start

### Basic Usage

```typescript
import { Session, Severity } from 'ptx-trace';

// Create and start a session
const session = await Session.create("import-customers", {
  piiMasking: true,
  tags: ["production"]
});

try {
  // Trace operations with automatic timing
  await session.withOperation("normalize_phone", async (op) => {
    op.input("phone", "+420 777 123 456");
    const result = await normalizePhone("+420 777 123 456");
    op.output("phone", result);
  });
} finally {
  await session.close();
}
```

### Fluent API

```typescript
import { Session, csvSource } from 'ptx-trace';

const session = await Session.create("etl-pipeline");

try {
  // Fluent builder pattern with execute
  await session.trace("transform_record")
    .input("record", { name: "John", phone: "123" })
    .source(csvSource("customers.csv", 42))
    .tag("transform", "customer")
    .execute(async (ctx) => {
      const result = await transform(record);
      ctx.output("result", result);
    });
} finally {
  await session.close();
}
```

### Manual Operation Control

```typescript
import { Session, Severity } from 'ptx-trace';

const session = await Session.create("validation-pipeline");

try {
  const op = session.startOperation("validate_email");

  try {
    op.input("email", email);
    const valid = await validateEmail(email);
    op.output("valid", valid);
    op.success();
  } catch (e) {
    op.error(String(e), Severity.HIGH);
    throw e;
  } finally {
    await op.end();
  }
} finally {
  await session.close();
}
```

## API Reference

### Session

The `Session` class manages the tracing session lifecycle.

```typescript
// Create and start session
const session = await Session.create("my-session", {
  source: "input.csv",           // Optional source info
  destination: "postgres://...", // Optional destination
  tags: ["production"],          // Optional tags
  sampling: 0.5,                 // Sample 50% of events
  piiMasking: true,              // Mask PII data
  binaryPath: "/usr/bin/portunix" // Custom binary path
});

// Load existing session
const session = await Session.load("ses_2026-01-27_import");

// List sessions
const sessions = await Session.list({ limit: 10, status: "active" });

// Session properties
session.id;        // Session ID
session.name;      // Session name
session.isActive;  // Whether session is active

// End session
await session.close();   // Complete successfully
await session.fail();    // Mark as failed
await session.cancel();  // Mark as cancelled

// Query data
const stats = await session.stats();
const events = await session.events({ limit: 100 });
const errors = await session.events({ level: "error" });
const results = await session.query("operation = 'validate'");
```

### Operation

Operations represent traced units of work.

```typescript
// Start operation manually
const op = session.startOperation("process_record");

try {
  // Set input data
  op.input("field", value);
  op.inputs({ field1: value1, field2: value2 });

  // Set output data
  op.output("result", value);
  op.outputs({ result1: value1, result2: value2 });

  // Add metadata
  op.tag("important", "customer");
  op.context("batch_id", 123);
  op.source(csvSource("data.csv", 42));

  // Record outcome
  op.success();
} catch (e) {
  op.error(String(e), Severity.HIGH);
  throw e;
} finally {
  await op.end();
}

// Or use withOperation for automatic lifecycle
await session.withOperation("process", async (op) => {
  op.input("data", data);
  const result = await process(data);
  op.output("result", result);
  // success() and end() called automatically
});
```

### OperationBuilder (Fluent API)

```typescript
// Full fluent API
await session.trace("transform")
  .withType("etl")                    // Set operation type
  .input("data", inputData)           // Set input
  .inputs({ key1: val1, key2: val2 }) // Multiple inputs
  .source(dbSource("pg://...", "users")) // Set source
  .tag("critical")                    // Add tags
  .withRule("R001", "1.0")           // Set rule info
  .context("env", "production")       // Add context
  .execute(async (ctx) => {           // Execute with tracing
    const result = await transform(inputData);
    ctx.output("result", result);
  });

// Build for manual control
const op = session.trace("transform")
  .input("data", data)
  .build();

try {
  // ... manual operation
  op.success();
} finally {
  await op.end();
}
```

### Source Types

```typescript
import { csvSource, dbSource, fileSource, apiSource } from 'ptx-trace';

// CSV file source
const source = csvSource("data.csv", 42, "email");

// Database source
const source = dbSource("postgres://localhost/db", "users", 100);

// Generic file source
const source = fileSource("config.json", 10);

// API source
const source = apiSource("https://api.example.com/users");
```

### Severity Levels

```typescript
import { Severity } from 'ptx-trace';

op.error("Minor issue", Severity.LOW);
op.error("Validation error", Severity.MEDIUM);
op.error("Data corruption", Severity.HIGH);
op.error("System failure", Severity.CRITICAL);
```

## Complete ETL Example

```typescript
import { Session, csvSource, Severity } from 'ptx-trace';
import * as fs from 'fs';
import * as csv from 'csv-parse/sync';

async function runEtl() {
  const session = await Session.create("customer-import", {
    piiMasking: true,
    tags: ["etl", "daily"]
  });

  try {
    let records: Record<string, string>[];

    // Phase 1: Extract
    await session.withOperation("extract", async (op) => {
      op.source(csvSource("customers.csv"));

      const content = fs.readFileSync("customers.csv", "utf-8");
      records = csv.parse(content, { columns: true });

      op.output("record_count", records.length);
    });

    // Phase 2: Transform each record
    const transformed: Record<string, string>[] = [];

    for (let i = 0; i < records.length; i++) {
      const record = records[i];

      await session.withOperation("transform", async (op) => {
        op.source(csvSource("customers.csv", i + 1));
        op.inputs(record);

        try {
          // Normalize phone
          record.phone = normalizePhone(record.phone);

          // Validate email
          if (!validateEmail(record.email)) {
            op.error("Invalid email", Severity.MEDIUM);
            return;
          }

          op.outputs(record);
          transformed.push(record);

        } catch (e) {
          op.error(String(e), Severity.HIGH);
        }
      });
    }

    // Phase 3: Load
    await session.withOperation("load", async (op) => {
      op.context("destination", "postgres://localhost/customers");

      const inserted = await loadToDatabase(transformed);

      op.output("inserted", inserted);
    });

    // Get statistics
    const stats = await session.stats();
    console.log(`Processed ${stats.stats?.totalEvents} events`);

  } finally {
    await session.close();
  }
}

function normalizePhone(phone: string): string {
  return phone?.replace(/\s+/g, "") ?? "";
}

function validateEmail(email: string): boolean {
  return email?.includes("@") ?? false;
}

async function loadToDatabase(records: Record<string, string>[]): Promise<number> {
  // Implementation
  return records.length;
}

runEtl().catch(console.error);
```

## Error Handling

```typescript
import { Session, Severity, CliError } from 'ptx-trace';

try {
  const session = await Session.create("risky-operation");

  try {
    await session.withOperation("critical_step", async (op) => {
      try {
        const result = await riskyFunction();
        op.output("result", result);
      } catch (e) {
        if (e instanceof ValidationError) {
          op.error(e.message, Severity.MEDIUM);
          // Handle gracefully
        } else {
          op.error(String(e), Severity.CRITICAL);
          throw e;
        }
      }
    });
  } finally {
    await session.close();
  }
} catch (e) {
  if (e instanceof CliError) {
    console.error(`CLI Error: ${e.message}`);
    console.error(`Exit code: ${e.exitCode}`);
  } else {
    throw e;
  }
}
```

## Configuration

Configure the SDK through session options:

```typescript
// Custom binary path
const session = await Session.create("my-session", {
  binaryPath: "/opt/portunix/bin/portunix"
});

// Or via environment variable in code
process.env.PORTUNIX_PATH = "/opt/portunix/bin/portunix";
```

## TypeScript Types

All types are fully exported:

```typescript
import type {
  SessionInfo,
  SessionStats,
  TraceEvent,
  SourceInfo,
  ErrorInfo,
  SessionOptions,
  EventQueryOptions,
} from 'ptx-trace';
```

## Async/Await Pattern

The SDK is fully async and works well with modern JavaScript:

```typescript
// Serial processing
for (const record of records) {
  await session.withOperation("process", async (op) => {
    // ...
  });
}

// Parallel with limit (using p-limit or similar)
import pLimit from 'p-limit';

const limit = pLimit(4);

await Promise.all(
  records.map(record =>
    limit(async () => {
      // Create separate session per parallel task if needed
      await session.withOperation("process", async (op) => {
        // ...
      });
    })
  )
);
```

## License

MIT License - see LICENSE file for details.
