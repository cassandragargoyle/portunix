# Issue #141: PTX-TRACE Helper Implementation

**Type**: Feature / Helper Binary
**Priority**: High
**Status**: ✅ Implemented
**Created**: 2026-01-27
**Closed**: 2026-05-06
**Labels**: helper-binary, data-transformation, tracing, etl, debugging, ai-integration, mcp
**ADR**: ADR-037

## Summary

Implement PTX-TRACE as a universal tracing system for software development. The helper is designed as an extensible platform supporting multiple tracing domains to facilitate debugging, AI analysis, and development workflow monitoring.

**First implementation domain: Structured Data Transformation** - capturing, storage, and visualization of data transformations during ETL/ELT processes, data pipelines, and data preprocessing. Future domains include API tracing, build/CI processes, and AI/ML workflows. The system replaces unstructured text logs with intelligent tracing optimized for debugging, AI analysis, and real-time monitoring.

Future domains (AI/ML workflows, API tracing, Build/CI processes) will be added as domain-specific adapters.

## Problem Statement

| Problem | Current State | PTX-TRACE Solution |
| ------- | ------------- | ----------------- |
| Unreadable logs | Thousands of lines without structure | Hierarchical JSON records with context |
| Hard to find errors | Grep through megabytes | Indexed search, filters, drill-down |
| Missing context | "Error on line 1542" | Full context: input, output, rule, time |
| AI sharing | Copy-paste fragments | Export optimized for Claude/GPT |
| Visualization | None or manual | Real-time dashboard, timeline, graphs |
| Reproducibility | Cannot repeat problem | Replay mode with original data |

## Architecture

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                           USER SOFTWARE                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │ Python App  │  │   Go App    │  │  Node.js    │  │   Bash      │     │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘     │
│         │                │                │                │            │
│         └────────────────┴────────────────┴────────────────┘            │
│                                   │                                     │
│                          ┌────────▼────────┐                            │
│                          │    SDK/Logger   │                            │
│                          │  (trace.Start)  │                            │
│                          └────────┬────────┘                            │
└───────────────────────────────────┼─────────────────────────────────────┘
                                    │
                    ┌───────────────▼───────────────┐
                    │         PTX-TRACE             │
                    │  ┌─────────────────────────┐  │
                    │  │      Collector          │  │
                    │  │  (NDJSON Writer/gRPC)   │  │
                    │  └───────────┬─────────────┘  │
                    │              │                │
                    │  ┌───────────▼─────────────┐  │
                    │  │    Storage Engine       │  │
                    │  │  (Files/SQLite/Index)   │  │
                    │  └───────────┬─────────────┘  │
                    │              │                │
                    │  ┌───────────▼─────────────┐  │
                    │  │    Query Engine         │  │
                    │  │  (Filter/Aggregate)     │  │
                    │  └───────────┬─────────────┘  │
                    └──────────────┼───────────────┘
                                   │
           ┌───────────────────────┼───────────────────────┐
           │                       │                       │
    ┌──────▼──────┐        ┌───────▼───────┐        ┌──────▼──────┐
    │  CLI View   │        │ Web Dashboard │        │   Export    │
    │  (Terminal) │        │   (React)     │        │  (AI/DB)    │
    └─────────────┘        └───────────────┘        └─────────────┘
```

## Data Model

### Trace Event (Core Entity)

```json
{
  "_v": 1,
  "id": "evt_01HQ3X5K7M9N2P4R6S8T0V2W4Y",
  "trace_id": "trc_01HQ3X5K7M9N2P4R6S8T0V2W4X",
  "parent_id": null,
  "session_id": "ses_2026-01-27_import-customers",

  "timestamp": "2026-01-27T14:32:15.123456Z",
  "duration_us": 2340,

  "operation": {
    "type": "transform",
    "name": "normalize_phone",
    "category": "normalization",
    "version": "1.2.0"
  },

  "input": {
    "fields": {"phone": "+420 777-123-456"},
    "source": {
      "type": "csv",
      "file": "customers.csv",
      "row": 1542,
      "column": "phone_number"
    },
    "checksum": "sha256:a1b2c3..."
  },

  "output": {
    "fields": {"phone": "420777123456"},
    "status": "success",
    "checksum": "sha256:d4e5f6..."
  },

  "context": {
    "rule_id": "rule_czech_phone_v2",
    "rule_name": "Czech Phone Format",
    "confidence": 0.98
  },

  "tags": ["customer", "contact", "pii"],
  "level": "info",

  "performance": {
    "cpu_us": 1200,
    "memory_bytes": 4096,
    "allocations": 3
  },

  "metadata": {
    "hostname": "worker-01",
    "pid": 12345
  }
}
```

### Session

```json
{
  "id": "ses_2026-01-27_import-customers",
  "name": "Customer Import Q1 2026",
  "started_at": "2026-01-27T14:30:00Z",
  "ended_at": "2026-01-27T14:45:32Z",
  "status": "completed",

  "source": {
    "type": "csv",
    "files": ["customers_part1.csv", "customers_part2.csv"],
    "total_rows": 45000
  },

  "destination": {
    "type": "postgresql",
    "table": "customers",
    "inserted": 44850,
    "updated": 100,
    "skipped": 50
  },

  "stats": {
    "total_events": 180000,
    "by_status": {"success": 179500, "warning": 450, "error": 50},
    "by_operation": {
      "normalize_phone": {"count": 45000, "avg_us": 234},
      "validate_email": {"count": 45000, "avg_us": 156}
    }
  },

  "config": {
    "sampling_rate": 1.0,
    "pii_masking": true,
    "retention_days": 30
  }
}
```

### Error Event Extension

```json
{
  "level": "error",
  "error": {
    "code": "E_VALIDATION_FAILED",
    "message": "Invalid email format",
    "category": "validation",
    "severity": "medium",
    "details": {
      "expected": "valid email with domain",
      "actual": "john.doe@",
      "rule": "RFC5322"
    },
    "suggestion": "Check for missing domain part after @"
  },
  "recovery": {
    "attempted": true,
    "strategy": "fallback_to_null",
    "success": true
  }
}
```

## Storage Layout

```text
~/.portunix/trace/
├── sessions/
│   └── ses_2026-01-27_import-customers/
│       ├── manifest.json           # Session metadata
│       ├── config.json             # Runtime configuration
│       ├── chunks/
│       │   ├── 000001.ndjson       # Events 1-10000
│       │   ├── 000002.ndjson       # Events 10001-20000
│       │   └── 000003.ndjson.zst   # Compressed older chunks
│       ├── index/
│       │   ├── operations.idx      # Operation index
│       │   ├── errors.idx          # Error index
│       │   ├── timeline.idx        # Time index
│       │   └── tags.idx            # Tag index
│       ├── summary.json            # Aggregated statistics
│       └── lineage.json            # Data lineage graph
│
├── index/
│   ├── sessions.db                 # SQLite: all session metadata
│   └── global_tags.idx             # Global tag index
│
├── exports/
│   ├── ai-context-latest.md        # Latest AI export
│   └── reports/
│
├── config/
│   ├── default.yaml                # Default configuration
│   ├── rules/
│   │   ├── pii-masking.yaml
│   │   └── sampling.yaml
│   └── alerts/
│       └── thresholds.yaml
│
└── cache/
    └── schema-inference/
```

## SDK API

### Go SDK

```go
// Initialize session
session, err := trace.NewSession("import-customers",
    trace.WithSource("customers.csv", trace.SourceCSV),
    trace.WithDestination("postgresql://...", trace.DestPostgres),
    trace.WithPIIMasking(true),
    trace.WithSampling(1.0),
    trace.WithTags("production", "q1-2026"),
)
defer session.Close()

// Basic tracing
op := session.Start("normalize_phone")
op.Input("phone", rawPhone)
op.Source(trace.CSVSource{File: "data.csv", Row: rowNum, Column: "phone"})

result, err := normalizePhone(rawPhone)
if err != nil {
    op.Error(err, trace.SeverityMedium)
    op.Recovery("fallback_to_null", true)
} else {
    op.Output("phone", result)
    op.Success()
}
op.End()

// Fluent API
session.Trace("validate_email").
    Input("email", email).
    Source(trace.CSVSource{Row: rowNum}).
    Tag("validation", "critical").
    WithRule("RFC5322", "1.0").
    Execute(func(ctx trace.Context) error {
        result, err := validateEmail(email)
        if err != nil { return err }
        ctx.Output("valid", result.Valid)
        return nil
    })

// Hierarchical tracing
parent := session.Start("process_record")
parent.Input("record", record)

parent.Child("validate").Execute(func(ctx trace.Context) error {
    return validate(record)
})
parent.Child("transform").Execute(func(ctx trace.Context) error {
    return transform(record)
})

parent.Success()
parent.End()

// Batch operations
batch := session.Batch("bulk_insert", 1000)
for _, record := range records {
    batch.Add(func(ctx trace.Context) error {
        ctx.Input("id", record.ID)
        return nil
    })
}
batch.Flush()
```

### Python SDK

```python
from ptx_trace import Session, Source, Severity

with Session(
    name="import-customers",
    source=Source.csv("customers.csv"),
    pii_masking=True,
    tags=["production"]
) as session:

    # Context manager
    with session.trace("normalize_phone") as op:
        op.input(phone=raw_phone)
        op.source(file="data.csv", row=row_num)

        try:
            result = normalize_phone(raw_phone)
            op.output(phone=result)
            op.success()
        except ValidationError as e:
            op.error(e, severity=Severity.MEDIUM)
            op.recovery("fallback_to_null", success=True)

    # Decorator
    @session.traced("validate_email", tags=["validation"])
    def validate_email(email: str) -> dict:
        return {"valid": True, "normalized": email.lower()}
```

### Java SDK

```java
import ai.portunix.trace.*;

// Try-with-resources pattern
try (Session session = Trace.newSession("import-customers")
        .withSource("customers.csv", SourceType.CSV)
        .withDestination("postgresql://...", DestType.POSTGRES)
        .withPIIMasking(true)
        .withTags("production", "q1-2026")
        .build()) {

    // Basic tracing with try-with-resources
    try (Operation op = session.start("normalize_phone")) {
        op.input("phone", rawPhone);
        op.source(CsvSource.of("data.csv", rowNum, "phone"));

        String result = normalizePhone(rawPhone);
        op.output("phone", result);
        op.success();
    } catch (ValidationException e) {
        op.error(e, Severity.MEDIUM);
        op.recovery("fallback_to_null", true);
    }

    // Fluent API with lambda
    session.trace("validate_email")
        .input("email", email)
        .source(CsvSource.of(rowNum))
        .tag("validation", "critical")
        .withRule("RFC5322", "1.0")
        .execute(ctx -> {
            boolean valid = validateEmail(email);
            ctx.output("valid", valid);
        });

    // Annotation-based tracing
    @Traced(operation = "process_record", tags = {"etl"})
    public Result processRecord(Record record) {
        // Method automatically traced
        return transform(record);
    }
}
```

### Bash/CLI Integration

```bash
# Inline tracing
portunix trace start "backup-database" --tag production

portunix trace event "dump_tables" \
  --input "tables=users,orders,products" \
  --status success \
  --duration 45000

portunix trace end --status completed

# Pipe-based
cat data.csv | portunix trace pipe "process_csv" --format csv | process.sh

# Wrapper
portunix trace exec "pg_dump" -- pg_dump -h localhost mydb > dump.sql
```

## CLI Commands

### Session Management

```bash
portunix trace start <name> [flags]
  --source <file|url>           # Source data
  --destination <connection>    # Target system
  --tag <tag>                   # Add tag (repeatable)
  --sampling <0.0-1.0>          # Sampling rate
  --pii-mask                    # Mask PII data

portunix trace end [flags]
  --status <completed|failed|cancelled>
  --summary                     # Show summary

portunix trace sessions [flags]
  --limit <n>                   # Result count
  --status <status>             # Filter by status
  --tag <tag>                   # Filter by tag
  --since <duration>            # Since when (24h, 7d)
  --format <table|json|csv>     # Output format

portunix trace session <session-id> [flags]
  --stats                       # Show statistics
  --errors                      # Show errors only
  --timeline                    # Timeline view
```

### Viewing & Querying

```bash
portunix trace view [session-id] [flags]
  --operation <name>            # Filter by operation
  --status <success|error|warning>
  --level <debug|info|warn|error>
  --tag <tag>                   # Filter by tag
  --since/--until <timestamp>   # Time range
  --limit <n>                   # Result count
  --follow                      # Live tail
  --format <pretty|json|table>  # Output format

portunix trace browse [session-id]   # Interactive TUI

portunix trace query "SELECT * FROM events WHERE operation.name = 'validate_email' AND status = 'error' LIMIT 10"

portunix trace stats [session-id] [flags]
  --group-by <field>            # Group by field
  --metric <count|avg|sum|p50|p95|p99>
  --format <table|json|chart>

portunix trace errors [session-id] [flags]
  --limit <n>
  --group                       # Group similar errors
  --with-context                # Include surrounding events
```

### Export

```bash
portunix trace export ai [session-id] [flags]
  --focus <errors|slow|all>     # Export focus
  --max-tokens <n>              # Token limit
  --include-samples             # Include sample data
  --output <file>               # Output file

portunix trace export db [session-id] [flags]
  --connection <string>         # Connection string
  --table <name>                # Target table
  --mode <insert|upsert>        # Insert mode

portunix trace export elastic [session-id] [flags]
  --url <elasticsearch-url>
  --index <name>

portunix trace export file [session-id] [flags]
  --format <json|csv|parquet>
  --output <file>
  --compress                    # Compress output

portunix trace export lineage [session-id] [flags]
  --format <mermaid|dot|json>
  --output <file>
```

### Replay & Debug

```bash
portunix trace replay <session-id> [flags]
  --speed <0.5|1|2|10>          # Playback speed
  --pause-on-error              # Pause on error
  --interactive                 # Interactive mode

portunix trace debug <event-id>      # Full context for event

portunix trace reproduce <event-id> [flags]
  --output <script>             # Generate reproduction script
```

### Dashboard

```bash
portunix trace serve [flags]
  --port <port>                 # Port (default: 3000)
  --host <host>                 # Host (default: localhost)
  --auth                        # Require authentication
  --readonly                    # Read-only mode

portunix trace stream [session-id] [flags]
  --port <port>                 # WebSocket port
  --filter <query>              # Event filter
```

### Maintenance

```bash
portunix trace storage [flags]
  stats                         # Storage statistics
  compact <session-id>          # Compress session
  prune --older-than <duration> # Delete old sessions
  export-archive <session-id>   # Archive session

portunix trace index [flags]
  rebuild <session-id>          # Rebuild indexes
  optimize                      # Optimize indexes
```

## Web Dashboard

### Main Views

1. **Dashboard Overview**: Sessions count, events, errors, avg time, trends
2. **Session Detail**: Filters, timeline, event list with drill-down
3. **Event Detail**: Input/output diff, transformation details, context
4. **Error Analysis**: Grouped errors, patterns, AI analysis
5. **Data Lineage**: Visual graph of data transformations

## Advanced Features

### PII Masking

```yaml
# config/rules/pii-masking.yaml
pii_masking:
  enabled: true

  patterns:
    email:
      pattern: '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'
      mask: "***@***.***"
    phone:
      pattern: '\+?[0-9]{9,15}'
      mask: "***-***-****"
      preserve_country_code: true
    credit_card:
      pattern: '\b[0-9]{13,19}\b'
      mask: "****-****-****-{last4}"
      preserve_last_digits: 4

  fields:
    - name: "password"
      action: "redact"
    - name: "api_key"
      action: "hash"
```

### Intelligent Sampling

```yaml
# config/rules/sampling.yaml
sampling:
  default_rate: 0.1  # 10% in production

  rules:
    - condition: "status == 'error'"
      rate: 1.0  # Always log errors
    - condition: "duration_us > 10000"
      rate: 1.0  # Always log slow operations
    - condition: "'debug' in tags"
      rate: 1.0

  adaptive:
    enabled: true
    target_events_per_second: 1000
```

### Alerting

```yaml
# config/alerts/thresholds.yaml
alerts:
  channels:
    slack:
      webhook_url: "${SLACK_WEBHOOK_URL}"

  rules:
    - name: "High Error Rate"
      condition: "error_rate > 0.05"
      window: "5m"
      severity: "critical"
      channels: ["slack"]
```

## AI Export Format

```markdown
# Data Processing Session Analysis

## Quick Summary
- **Session**: import-customers (2026-01-27)
- **Status**: Completed with errors
- **Processed**: 45,000 records in 4m 32s
- **Success Rate**: 99.95% (23 errors)

## Error Breakdown

### 1. Email Validation Failures (15 errors)
**Pattern**: Truncated email addresses
**Sample**: `john.doe@` (missing domain)
**Suggestion**: Check source CSV column width

## Recommended Actions
1. Fix source CSV export to prevent truncation
2. Add PO Box handling rule
```

## Implementation Phases

### Phase 1: Core (Week 1-2)

- [x] Define JSON schema for Trace Event and Session
- [x] Implement NDJSON writer with chunking
- [x] Implement Go SDK basic API (NewSession, Start, End, Input, Output, Error, Success)
- [x] Implement CLI: `portunix trace start`, `portunix trace end`, `portunix trace event`
- [x] Implement CLI: `portunix trace view`, `portunix trace stats`
- [x] Basic file-based storage

### Phase 2: Query & Index (Week 3-4)

- [x] SQLite-based session index
- [x] Implement operation/error/timeline indexes
- [x] Query engine with SQL-like syntax
- [x] CLI: `portunix trace query`
- [x] Filters and aggregations
- [x] CLI: `portunix trace errors`

### Phase 3: Dashboard (Week 5-6)

- [x] React dashboard scaffold
- [x] Session list view
- [x] Session detail view with timeline
- [x] Event browser with drill-down
- [x] Real-time updates via WebSocket
- [x] Error analysis view

### Phase 4: Advanced (Week 7-8)

- [x] PII masking implementation
- [x] Intelligent sampling
- [x] AI export format
- [x] Database export (PostgreSQL/MySQL)
- [x] CLI: `portunix trace export ai`
- [x] CLI: `portunix trace export db`

### Phase 4b: Fulltext Search Integration ✅

- [x] Fulltext search export (via portunix fulltext plugin)
- [x] CLI: `portunix trace export fulltext`
- [x] Integration with Lucene/Elasticsearch backends from fulltext plugin

**Dependencies (resolved):**

- Issue #142: Elasticsearch container installation ✅
- Issue #143: Fulltext plugin gRPC server mode ✅

### Phase 4c: Alerting System

- [x] Alert rules configuration (YAML-based)
- [x] Alert conditions (error_rate, slow_operations, thresholds)
- [x] Alert channels (webhook, file, stdout, slack)
- [x] Real-time alert evaluation during session
- [x] CLI: `portunix trace alerts` (rules, history, stats, test, clear)
- [x] Alert history and management
- [x] SDK integration (`WithAlerting()`, `WithAlertConfig()`, `WithAlertManager()`)
- [x] CLI flags: `--alerts`, `--alerts-config`, `--show-alerts`

### Phase 5: SDKs & Polish (Week 9-10)

- [x] Python SDK with decorator support
- [x] Java SDK with try-with-resources and annotations
- [x] TypeScript SDK
- [x] Bash integration improvements
- [x] Documentation
- [x] Examples and tutorials
- [x] MCP integration for AI assistants

## Configuration

```yaml
# ~/.portunix/trace/config/default.yaml
ptx_trace:
  version: 1

  storage:
    path: "~/.portunix/trace"
    chunk_size: 10000
    compression: "zstd"
    compression_after_hours: 24

  retention:
    default_days: 30
    error_sessions_days: 90
    max_storage_gb: 10

  performance:
    buffer_size: 1000
    flush_interval_ms: 1000
    async_write: true

  defaults:
    sampling_rate: 1.0
    pii_masking: false
    include_stack_traces: true
    max_field_size_kb: 64

  dashboard:
    port: 3000
    auto_refresh_seconds: 5
```

## Acceptance Criteria

### Phase 1

- [x] Go SDK allows creating sessions and tracing operations
- [x] Events stored in NDJSON format with automatic chunking
- [x] CLI can start/end sessions and view events
- [x] Basic statistics available via `portunix trace stats`

### Phase 2

- [x] Sessions indexed in SQLite for fast listing
- [x] SQL-like queries work on events
- [x] Filtering by operation, status, time works
- [x] Error grouping and context available

### Phase 3

- [x] Web dashboard accessible via `portunix trace serve`
- [x] Session list with search and filters
- [x] Event timeline with drill-down
- [x] Real-time updates when session active

### Phase 4

- [x] PII patterns masked in stored events
- [x] Sampling reduces storage for high-volume ops
- [x] AI export generates Claude-friendly markdown
- [x] Database export works for PostgreSQL and MySQL

### Phase 4b ✅

- [x] Trace events can be exported to fulltext search engine
- [x] Integration with portunix fulltext plugin (Lucene/Elasticsearch)
- [x] Elasticsearch available via container (Issue #142)
- [x] Fulltext plugin runs as gRPC server (Issue #143)

### Phase 4c

- [x] Alert rules configurable via YAML
- [x] Alerts fire on error_rate, slow_operations thresholds
- [x] Alert channels work (webhook, file, stdout, slack)
- [x] Alert history queryable
- [x] Real-time alert evaluation during session
- [x] SDK options for enabling alerting (`WithAlerting()`)
- [x] CLI `--alerts` flag for `trace start`

### Phase 5

- [x] Python SDK works with `with` statement and decorators
- [x] Java SDK works with try-with-resources and annotations
- [x] TypeScript SDK works with async/await
- [x] Complete documentation with examples
- [x] MCP tools available for AI assistants

## Related Items

- **ADR-037**: PTX-TRACE Architecture Decision
- **HELPER-BINARY-DEVELOPMENT.md**: Mandatory checklist for helper binary implementation
- **Issue #007**: Plugin System with gRPC Architecture
- **Issue #132**: Text Extractor Plugin (similar pattern)

## Notes

- Helper will be developed in main `portunix` repository as `ptx-trace` binary
- Follow `docs/contributing/HELPER-BINARY-DEVELOPMENT.md` checklist for implementation
- Integrates with main Portunix via dispatcher pattern (like ptx-container, ptx-mcp, ptx-virt)
- Dashboard is optional component, CLI works standalone
- Build integrated into Makefile and GoReleaser configuration
