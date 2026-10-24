# sqlite.go

Documentation generated from code — `src/helpers/ptx-trace/index/sqlite.go`.

## Overview

This file implements the **query index of `ptx-trace`**: a SQLite database that mirrors
trace sessions and events so they can be filtered, grouped and aggregated with SQL.

The NDJSON files written by the `storage` package stay the source of truth. The index is a
derived cache — it can be deleted at any time and is rebuilt from the NDJSON files on demand.

| Property | Value |
| -------- | ----- |
| Package | `index` |
| Database file | `<baseDir>/index/sessions.db` (default base directory `~/.portunix/trace`) |
| Driver | `modernc.org/sqlite` — pure Go, no CGO, registered under the name `sqlite` |
| Directory mode | `0750`, trace data may contain command lines and error messages |

## Lifecycle

`NewIndex(baseDir)` creates the `index` directory, opens the database and runs `initSchema`.
The schema uses `CREATE ... IF NOT EXISTS` only, so opening an existing database is safe and
there is no migration step. When schema creation fails the connection is closed before the
error is returned.

`Close()` closes the database connection. Callers open the index per command and close it
with `defer`.

## Schema

### Table `sessions`

One row per trace session with pre-computed counters.

| Column | Type | Notes |
| ------ | ---- | ----- |
| `id` | TEXT | Primary key |
| `name` | TEXT | Required |
| `status` | TEXT | Required |
| `started_at` | DATETIME | Required |
| `ended_at` | DATETIME | Empty while the session is running |
| `total_events` | INTEGER | From `Session.Stats.TotalEvents` |
| `error_count` | INTEGER | From `Session.Stats.ByStatus["error"]` |
| `warning_count` | INTEGER | From `Session.Stats.ByStatus["warning"]` |
| `success_count` | INTEGER | From `Session.Stats.ByStatus["success"]` |
| `tags` | TEXT | Tags joined with `,` |
| `created_at` | DATETIME | Set on insert |
| `updated_at` | DATETIME | Set on every write |

### Table `events`

One row per trace event. Only the fields needed for querying are stored — inputs, outputs
and context stay in the NDJSON files.

| Column | Type | Notes |
| ------ | ---- | ----- |
| `id` | TEXT | Primary key |
| `session_id` | TEXT | Foreign key to `sessions(id)` |
| `trace_id`, `parent_id` | TEXT | Span hierarchy |
| `timestamp` | DATETIME | Required |
| `duration_us` | INTEGER | Duration in microseconds |
| `operation_type`, `operation_name`, `operation_category` | TEXT | `operation_name` is required |
| `level` | TEXT | Required |
| `status` | TEXT | From `Output.Status`, empty when the event has no output |
| `has_error` | BOOLEAN | True when the event carries an error |
| `error_code`, `error_message`, `error_severity` | TEXT | Empty strings when there is no error |
| `tags` | TEXT | Tags joined with `,` |
| `created_at` | DATETIME | Set on insert |

### Indexes

| Index | Column | Used by |
| ----- | ------ | ------- |
| `idx_events_session` | `events(session_id)` | Every per-session query |
| `idx_events_timestamp` | `events(timestamp)` | Time range filters and ordering |
| `idx_events_operation` | `events(operation_name)` | Operation filter and statistics |
| `idx_events_level` | `events(level)` | Level filter |
| `idx_events_error` | `events(has_error)` | Error queries and grouping |
| `idx_sessions_status` | `sessions(status)` | Session listing by status |
| `idx_sessions_started` | `sessions(started_at)` | Session listing by start time |

## Writing

| Method | Behavior |
| ------ | -------- |
| `IndexSession(session)` | Upsert by `id` — updates name, status, end time, counters and tags |
| `IndexEvent(event)` | Insert by `id`, an existing event is left untouched (`ON CONFLICT DO NOTHING`) |
| `RebuildSessionIndex(sessionID, events, session)` | Replaces all events of one session in a single transaction |

`RebuildSessionIndex` is the write path used in practice. It deletes the events of the
session, inserts the given events through one prepared statement and, when `session` is not
nil, upserts the session row through `indexSessionTx`. Any error rolls the whole transaction
back, so a session is never left half indexed.

Callers rebuild lazily: they call `GetSessionStats` and, when it fails because the session is
not indexed yet, load the NDJSON files and call `RebuildSessionIndex`. This pattern is used by
the CLI and the dashboard server (`rebuildIndexIfNeeded` in `main.go` and `server/api.go`).
The AI export rebuilds through `rebuildIndex` in `export/ai.go` when an index query fails.

## Reading

### QueryEvents

`QueryEvents(sessionID, filter)` returns events as `QueryResult` rows, newest first. An empty
`sessionID` searches all sessions and a nil filter returns everything. All values are passed
as bound parameters, never concatenated into the SQL text.

| `QueryFilter` field | SQL condition |
| ------------------- | ------------- |
| `Operation` | `operation_name = ?` |
| `Level` | `level = ?` |
| `Status` | `status = ?` |
| `HasError` | `has_error = TRUE` (false means no restriction) |
| `ErrorCode` | `error_code = ?` |
| `Since` / `Until` | `timestamp >= ?` / `timestamp <= ?` |
| `Tag` | `tags LIKE '%tag%'` |
| `MinDuration` | `duration_us >= ?` (microseconds) |
| `Limit` | `LIMIT ?` (zero means no limit) |

### GetErrorGroups

`GetErrorGroups(sessionID, limit)` groups error events by `error_code` and `error_message`
and returns `ErrorGroup` values ordered by count, most frequent first. Each group carries the
count, first and last occurrence, severity and the distinct operation names that raised it.

### GetSessionStats

`GetSessionStats(sessionID)` returns `SessionIndexStats`: the counters stored in the
`sessions` row plus, per operation name, the event count and average duration
(`OperationIndexStats`). It returns `sql.ErrNoRows` when the session is not indexed, which is
the signal callers use to trigger a rebuild.

## Limitations

- **Tag filter is a substring match** — tags are stored as one comma-separated string, so the
  filter `db` also matches an event tagged `dbx`.
- **Tags must not contain commas** — a comma inside a tag is indistinguishable from a separator.
- **Events are immutable in the index** — `IndexEvent` ignores an event whose `id` already
  exists, use `RebuildSessionIndex` to refresh changed events.
- **Rebuild does not remove sessions** — deleting a session from storage leaves its rows in
  the index until the database file is removed.
- **Timestamps in error groups are parsed from text** — `GetErrorGroups` accepts the RFC 3339
  and the space-separated SQLite format, any other format yields a zero time.
- **No schema versioning** — a change to an existing column requires deleting
  `sessions.db`, which is safe because the index is rebuilt from the NDJSON files.

---

**Created**: 2026-10-11
**Last Updated**: 2026-10-11
