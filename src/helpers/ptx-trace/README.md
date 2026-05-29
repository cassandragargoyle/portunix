# ptx-trace

Universal tracing helper for Portunix. Captures, stores, and visualizes
operations during data processing, ETL/ELT pipelines, API calls, build
processes, and any software workflow that benefits from structured
debugging and AI-assisted analysis.

`ptx-trace` is a helper binary invoked by the main `portunix` dispatcher
when running `portunix trace ...` commands. It is also embedded as a Go
package and exposed through external SDKs (Python, Java, TypeScript, Bash).

- Issue: [#141](../../../docs/issues/internal/141-ptx-trace-helper-implementation.md)
- ADR: [037 — PTX-TRACE Architecture](../../../docs/adr/)
- Implementation guide: [`HELPER-BINARY-DEVELOPMENT.md`](../../../docs/contributing/HELPER-BINARY-DEVELOPMENT.md)

## What it does

- **Structured events** — every operation becomes a JSON record with
  input, output, duration, status, tags, and full context.
- **Sessions** — group related events with metadata (source, destination,
  timing, aggregated stats).
- **Storage** — NDJSON chunks per session, plus a SQLite index for fast
  listing and queries.
- **Querying** — SQL-like filters across operations, status, time range,
  tags; grouped error analysis.
- **Exports** — AI-optimized markdown for Claude/GPT, JSON/CSV files,
  PostgreSQL/MySQL, fulltext search (Elasticsearch via the fulltext plugin).
- **Real-time alerting** — error-rate, slow-operation, and threshold
  rules, with webhook/Slack/file/stdout channels.
- **Web dashboard** — React UI served by `portunix trace serve` for
  timeline, drill-down, and live updates.

## Architecture

```text
USER SOFTWARE (Go / Python / Java / TS / Bash)
            │
       SDK / CLI
            │
            ▼
     ┌────────────────────────────────┐
     │           ptx-trace            │
     │  Collector → Storage → Query   │
     │  Sampling │ PII mask │ Alerts  │
     └─────────────┬──────────────────┘
                   │
        ┌──────────┼──────────┐
        ▼          ▼          ▼
     CLI view   Web UI     Exports
                          (AI / DB / Fulltext)
```

Storage layout under `~/.portunix/traces/`:

```text
sessions/<session-id>/
  manifest.json          session metadata
  chunks/000001.ndjson   events (chunked, compressed after 24h)
  index/                 per-session indexes
  summary.json           aggregated stats
index/sessions.db        global SQLite index
exports/                 generated AI / file exports
config/                  default + custom rules (PII, sampling, alerts)
```

## Quick start

```bash
# Start a session
portunix trace start "import-customers" --source data.csv --tag etl

# Record events from your script
portunix trace event "validate_email" --input "email=a@b.c" --status success

# Pipe-based tracing in shell pipelines (records bytes/lines/duration)
cat data.csv | portunix trace pipe "ingest" --format csv | next-stage

# Wrap an external command (records duration + exit code, propagates code)
portunix trace exec "backup" --tag prod -- pg_dump -h localhost mydb

# End the session and inspect
portunix trace end --summary
portunix trace sessions
portunix trace view --status error
```

## CLI commands

| Group | Command | Purpose |
| ----- | ------- | ------- |
| **Session** | `trace start <name>` | Start a new session |
| | `trace end` | End the active session |
| | `trace sessions` | List sessions |
| **Recording** | `trace event <op>` | Record one event in the active session |
| | `trace pipe <op>` | Trace a stdin/stdout pipe (bytes, lines, duration) |
| | `trace exec <op> -- <cmd>` | Wrap an external command (duration, exit code) |
| **Analysis** | `trace view [session]` | View events with filters |
| | `trace stats [session]` | Aggregated statistics |
| | `trace query <sql>` | SQL-like filters across events |
| | `trace errors [session]` | Grouped error analysis |
| **Export** | `trace export ai` | Markdown optimized for AI/LLM |
| | `trace export file` | JSON / CSV |
| | `trace export db` | PostgreSQL / MySQL |
| | `trace export fulltext` | Elasticsearch via fulltext plugin |
| **Alerts** | `trace alerts rules` | Manage alert rules |
| | `trace alerts history` | Alert history |
| | `trace alerts test` | Dry-run rules against a session |
| **Index** | `trace index rebuild` | Rebuild SQLite index |
| **Server** | `trace serve` | Start web dashboard |

Use `portunix trace <command> --help` for full flag documentation, or
`portunix trace --help-expert` for a one-page reference.

## SDKs

External SDKs communicate with `ptx-trace` through the CLI subprocess so
all language bindings share the same validated logic.

| Language | Status | Entry point |
| -------- | ------ | ----------- |
| **Go** | Stable (built-in) | `sdk/sdk.go` |
| **Python** | Stable | [`sdk/python/`](sdk/python/) |
| **Java** | Stable | [`sdk/java/`](sdk/java/) |
| **TypeScript** | Stable | [`sdk/typescript/`](sdk/typescript/) |
| **Bash** | Stable | [`sdk/bash/ptx-trace.sh`](sdk/bash/ptx-trace.sh) |

See the [SDK comparison](sdk/README.md) for side-by-side API examples and
feature matrix.

## Source layout

| Directory | Contents |
| --------- | -------- |
| `main.go` | CLI entry point and command wiring |
| `sdk/` | Go SDK + external SDKs (Python / Java / TS / Bash) |
| `models/` | Trace event and session data models |
| `storage/` | NDJSON writer, chunking, session persistence |
| `index/` | SQLite-backed session and operation index |
| `export/` | AI / file / DB / fulltext exporters |
| `pii/` | PII detection and masking |
| `sampling/` | Static and adaptive sampling |
| `alerts/` | Alert rules, evaluation, channels (webhook/slack/file/stdout) |
| `server/` | Web dashboard (REST + WebSocket) |
| `proto/` | Protocol buffer definitions |

## Configuration

Default config: `~/.portunix/traces/config/default.yaml`. See the issue
specification for all keys (storage, retention, performance, defaults,
dashboard).

## Building

The helper is built by the main Portunix Makefile:

```bash
make build           # main binary + all helpers
make build-helpers   # only helper binaries (including ptx-trace)
```

The helper is registered with the dispatcher and routed automatically
when the user runs `portunix trace ...`.

## Testing

```bash
go test ./src/helpers/ptx-trace/... -v
```

Integration tests for tracing flows live in `test/integration/`.

## See also

- [Issue #141](../../../docs/issues/internal/141-ptx-trace-helper-implementation.md) — full specification
- [SDK README](sdk/README.md) — multi-language SDK overview
- [Bash SDK README](sdk/bash/README.md) — shell helper reference
- [Helper binary checklist](../../../docs/contributing/HELPER-BINARY-DEVELOPMENT.md)

## License

MIT — see the [LICENSE](../../../LICENSE) file at the repository root.
