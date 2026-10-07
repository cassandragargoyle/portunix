# Issue #13: PTX-Database Helper Implementation

**Type**: Feature / Helper Binary
**Priority**: High
**Status**: ✅ Implemented (Phase 1)
**Created**: 2025-09-01
**Updated**: 2026-05-09
**Closed**: 2026-05-09 (Phase 1 only — Phases 2-5 tracked separately)
**Acceptance Protocol**: [docs/testing/acceptance-013.md](../../testing/acceptance-013.md) — PASS
**Labels**: helper-binary, database, mcp, installation, backup
**Related Issues**:

- [#051](../051-git-dispatcher-python-distribution-architecture.md) - Git-like Dispatcher with Python Distribution Architecture
- [#100](100-ptx-installer-helper-implementation.md) - PTX-Installer Helper (used for DB engine installation)
- [#033](219-mcp-plugin-development-guide-ai-agents.md) - MCP Server Plugin Development Guide
- [#035](../035-ai-assistant-installation-support.md) - AI Assistant Installation Support

## Summary

Implement `ptx-database` as a helper binary that extends the main `portunix` dispatcher with capabilities to install, configure, and manage popular database systems. The helper provides both CLI subcommands (`portunix database …`) and MCP tools (via `ptx-mcp`) so AI assistants can query database status, structure, and perform basic maintenance.

This is **not a gRPC plugin** — `ptx-database` follows the dispatcher/helper pattern established by `ptx-container`, `ptx-mcp`, `ptx-virt`, `ptx-installer`, `ptx-aiops`, and `ptx-trace` (see [HELPER-BINARY-DEVELOPMENT.md](../../contributing/HELPER-BINARY-DEVELOPMENT.md)).

## Problem Statement

| Problem | Current State | PTX-Database Solution |
| ------- | ------------- | -------------------- |
| Manual DB setup | Each DB engine installed by hand per OS | Unified `portunix database install <engine>` via `ptx-installer` packages |
| Heterogeneous lifecycle | systemd / launchd / Windows services / containers — different UX per engine | Single `database start/stop/status` interface |
| Backup procedures vary | `pg_dump`, `mysqldump`, `mongodump`, … with different flags | Common `database backup/restore` wrapper |
| AI cannot inspect DBs | No MCP tools for live DB queries / metadata | MCP tools (`db_status`, `db_list`, `db_tables`, `db_schema`, …) via `ptx-mcp` |
| Credential leakage risk | Passwords in shell history, plain-text config | Storage via `ptx-credential` (OS keychain) |
| Container vs native confusion | Users unsure when to run DB locally vs in container | Helper auto-selects via flag (`--mode native|container`) |

## Architecture

```text
User Command: portunix database install postgresql --version 16
                    ↓
         ┌──────────────────┐
         │ Main Dispatcher  │   (cmd routing only)
         │   (portunix)     │
         └────────┬─────────┘
                  │ delegates "database" command
                  ↓
         ┌──────────────────┐
         │   ptx-database   │   (this helper)
         │   Helper Binary  │
         └────────┬─────────┘
                  │
   ┌──────────────┼───────────────┬──────────────┐
   ↓              ↓               ↓              ↓
[ptx-installer] [ptx-container] [ptx-credential] [ptx-mcp]
 install pkg     run as ctr      store password   expose tools
```

### Helper Boundaries

`ptx-database` is the **orchestrator**. It does NOT re-implement package installation or container management — it delegates:

- **Installation** → `ptx-installer` reads JSON manifests from `assets/packages/` (one per engine, e.g. `postgresql.json`, `mysql.json`, `mongodb.json`)
- **Container mode** → `ptx-container` handles run/start/stop/exec
- **Credential storage** → `ptx-credential` (OS keychain wrapper)
- **MCP exposure** → tool handlers added to `src/app/mcp/handlers.go` (served by `ptx-mcp`)

What `ptx-database` owns directly:

- Database lifecycle abstractions (engine-agnostic start/stop/status)
- Backup/restore drivers per engine (wraps `pg_dump`, `mysqldump`, `mongodump`, `redis-cli BGSAVE`, …)
- Schema/metadata introspection drivers (read-only DDL/DML for status & schema MCP tools)
- Engine-specific configuration helpers (e.g. PostgreSQL `pg_hba.conf` patching, MySQL `my.cnf`)
- Health checks (engine ping/version negotiation)

## Supported Database Engines

### Phase 1 (MVP)

| Engine | Mode | Versions | Notes |
| ------ | ---- | -------- | ----- |
| PostgreSQL | native + container | 14, 15, 16, 17 | TimescaleDB extension flag |
| MySQL | native + container | 8.0, 8.4 | |
| SQLite | embedded (file) | latest | No service lifecycle, only DDL/backup tools |
| Redis | native + container | 7.0, 7.2 | |

### Phase 2

| Engine | Mode | Versions |
| ------ | ---- | -------- |
| MariaDB | native + container | 10.11, 11.4 |
| MongoDB | native + container | 6.0, 7.0 |
| Elasticsearch | container only | 8.x (existing `assets/packages/elasticsearch.json`) |

### Phase 3 (Time-Series)

| Engine | Mode | Versions |
| ------ | ---- | -------- |
| InfluxDB | container | 2.x |
| TimescaleDB | PostgreSQL extension | (uses pg-16 base) |

## CLI Command Structure

All commands route through the main `portunix` dispatcher to the `ptx-database` helper:

```bash
# Installation (delegates to ptx-installer)
portunix database install <engine> [--version V] [--mode native|container] [--variant V]
portunix database uninstall <engine> [--purge]

# Lifecycle
portunix database start <engine|instance>
portunix database stop <engine|instance>
portunix database restart <engine|instance>
portunix database status [<engine|instance>] [--format table|json]

# Database / user management
portunix database create <name> --engine <postgres|mysql|...> [--owner <user>]
portunix database drop <name> --engine <postgres|mysql|...>
portunix database list [--engine <e>] [--format table|json]
portunix database user create <user> --engine <e> [--password-prompt]
portunix database user grant <user> <db> --engine <e> [--privileges ALL|RO|RW]

# Backup & restore
portunix database backup <name> --engine <e> [--destination <dir>] [--compress]
portunix database restore <file> --engine <e> [--target-db <name>]
portunix database backup list [--engine <e>]

# Introspection (used by MCP tools too)
portunix database tables <db> --engine <e> [--with-sizes] [--format json]
portunix database schema <db> --engine <e> [--format sql|json]
portunix database connections <engine|instance> [--format table|json]
portunix database performance <engine|instance> [--format json]

# Maintenance
portunix database vacuum <db> --engine postgres
portunix database optimize <db> --engine mysql
portunix database health <engine|instance>
```

### Configuration File

`~/.config/portunix/database.yaml` (mode 0600) — managed entries:

```yaml
instances:
  postgres-main:
    engine: postgresql
    version: "16"
    mode: native
    port: 5432
    data_dir: "/var/lib/postgresql/16/main"
    backup_dir: "~/.portunix/backups/postgres-main"
    auto_backup:
      enabled: true
      schedule: "daily"
      retention_days: 14
    credential_ref: "ptx-credential:postgres-main"   # password fetched from keychain

  redis-cache:
    engine: redis
    version: "7.2"
    mode: container
    container:
      name: portunix-redis-cache
      image: redis:7.2-alpine
      port: 6379
      volume: portunix-redis-cache-data
```

## MCP Tool Integration

Tool handlers added to `src/app/mcp/handlers.go`. The `ptx-mcp` helper exposes them via JSON-RPC and shells out to `portunix database …` with `--format json`.

### Status & monitoring

- `db_status` — instance up/down + version
- `db_list` — list databases on an instance
- `db_size` — DB size, growth indicators
- `db_tables` — list tables/collections (with sizes)
- `db_connections` — active connections
- `db_performance` — query stats / cache hit ratio / lag

### Maintenance

- `db_backup_status` — last backup timestamp + status
- `db_backup_create` — initiate backup
- `db_restore` — restore from a named backup
- `db_vacuum` — engine-specific maintenance (VACUUM / OPTIMIZE)

### Schema

- `db_schema` — full schema dump (DDL or JSON)
- `db_table_info` — single-table structure incl. columns, indexes, constraints
- `db_indexes` — list & analyze indexes (unused / missing index hints)
- `db_constraints` — foreign keys, unique, check constraints

Naming follows the convention from `ptx-trace` MCP integration (Issue #141): `db_<action>` prefix.

## Implementation Phases

Implementation MUST follow [HELPER-BINARY-DEVELOPMENT.md](../../contributing/HELPER-BINARY-DEVELOPMENT.md) — all 11 phases of the checklist apply (source code → dispatcher → Makefile → `build-with-version.sh` → `.goreleaser.yml` → install/update → deploy scripts → docs → tests).

### Phase 1: Helper skeleton + PostgreSQL/SQLite (Week 1-2)

- [ ] Create `src/helpers/ptx-database/` (cobra root, `ptx-database` binary)
- [ ] Register helper in `src/dispatcher/dispatcher.go` with command `database` (alias `db`)
- [ ] Update Makefile, `build-with-version.sh`, `.goreleaser.yml`, deploy scripts (full HELPER-BINARY checklist)
- [ ] Add `assets/packages/postgresql.json` if missing (native + container variants)
- [ ] Add `assets/packages/sqlite.json` (CLI tool only)
- [ ] Implement `database install/uninstall` (delegates to `ptx-installer`)
- [ ] Implement `database start/stop/status` for PostgreSQL (systemd + container modes)
- [ ] Implement `database backup/restore` driver for PostgreSQL (`pg_dump`/`pg_restore`)
- [ ] Implement `database list/tables/schema` for PostgreSQL & SQLite
- [ ] Unit tests in `src/helpers/ptx-database/` and integration test in `test/integration/ptx_database_test.go`

### Phase 2: MySQL + Redis + Configuration (Week 3-4)

- [ ] MySQL driver (lifecycle, backup via `mysqldump`, schema introspection)
- [ ] Redis driver (lifecycle, `BGSAVE` snapshots, `INFO` parsing for status)
- [ ] `~/.config/portunix/database.yaml` instance registry
- [ ] Integration with `ptx-credential` for password storage
- [ ] `database health` command + `database performance` JSON output

### Phase 3: MCP Integration (Week 5)

- [ ] Add 13 MCP tool definitions to `src/app/mcp/handlers.go` (`db_status`, `db_list`, `db_size`, `db_tables`, `db_connections`, `db_performance`, `db_backup_status`, `db_backup_create`, `db_restore`, `db_vacuum`, `db_schema`, `db_table_info`, `db_indexes`)
- [ ] Implement handler functions following `executePtxTrace()` pattern
- [ ] Verify with JSON-RPC: `tools/list` + `tools/call`
- [ ] Document tool usage in `docs/contributing/MCP-TOOLS.md` (or equivalent)

### Phase 4: MariaDB + MongoDB + Elasticsearch (Week 6-7)

- [ ] MariaDB driver (largely shares MySQL driver)
- [ ] MongoDB driver (`mongodump`, `mongorestore`, replica-set awareness)
- [ ] Elasticsearch driver (container-only, snapshot/restore via REST API)
- [ ] Reuse existing `assets/packages/elasticsearch.json`

### Phase 5: Time-Series + Polish (Week 8)

- [ ] InfluxDB driver
- [ ] TimescaleDB extension activation flow (post-install hook on PostgreSQL)
- [ ] `database vacuum/optimize` per engine
- [ ] Auto-backup scheduler (writes systemd-timer / Windows scheduled task)
- [ ] Cross-platform validation (Linux, Windows; container-mode covers macOS later)
- [ ] Documentation in `docs/FEATURES_OVERVIEW.md`

## Security

- Credentials stored only via `ptx-credential` (OS keychain) — never written to YAML in plain text. The config file references credentials by ID (`credential_ref`).
- TLS enforced by default for network listeners on engines that support it (PostgreSQL, MySQL, MongoDB, Elasticsearch). Helper generates self-signed certs in `~/.portunix/database/<instance>/tls/` when absent.
- Audit log of state-changing operations (install, drop, restore) appended to `~/.portunix/database/audit.log`.
- `database drop` and `database restore` require `--yes` confirmation flag (or interactive prompt).
- Engine processes never run as root; helper validates ownership of `data_dir` matches the engine service user.

## Testing Requirements

- Unit tests per driver (`postgresql_driver_test.go`, `mysql_driver_test.go`, …) using mocked exec
- Integration tests use **container mode only** (per `ISSUE-DEVELOPMENT-METHODOLOGY.md` — no host installation of test DBs):

```go
tf.Command(t, binaryPath, []string{"database", "install", "postgresql",
    "--version", "16", "--mode", "container"})
tf.Command(t, binaryPath, []string{"database", "start", "postgresql"})
tf.Command(t, binaryPath, []string{"database", "backup", "test_db", "--engine", "postgresql"})
```

- MCP tool integration test: spawn `ptx-mcp serve`, call each `db_*` tool via JSON-RPC, validate response shape
- Cross-platform: Linux primary, Windows for native PostgreSQL & MySQL
- Backup/restore round-trip verification (insert data → backup → drop → restore → verify rows)

## Acceptance Criteria

### Phase 1 (PostgreSQL + SQLite MVP)

- [ ] `portunix database install postgresql --version 16` succeeds in container and on Linux native
- [ ] `portunix database start/stop/status postgresql` controls lifecycle correctly
- [ ] `portunix database backup` produces a restorable dump; `portunix database restore` reproduces data
- [ ] `portunix database list/tables/schema` returns valid JSON when `--format json` is given
- [ ] Helper binary builds and ships through full release pipeline (verified via `make-release.py vX.Y-SNAPSHOT`)

### Phase 2 (MySQL + Redis)

- [ ] At least 4 engines (PostgreSQL, SQLite, MySQL, Redis) installable and manageable
- [ ] Credentials always stored via `ptx-credential` — `database.yaml` contains no plain-text passwords
- [ ] `database health` returns OK for running instances and surfaces actionable error for stopped/misconfigured ones

### Phase 3 (MCP)

- [ ] 13 `db_*` MCP tools listed by `ptx-mcp serve`
- [ ] Each tool returns parseable JSON when invoked through JSON-RPC
- [ ] AI assistant (Claude Code) can call `db_status` / `db_tables` / `db_schema` end-to-end against a live PostgreSQL instance

### Phase 4-5 (Full set)

- [ ] All Phase 1-3 engines + MariaDB, MongoDB, Elasticsearch, InfluxDB supported
- [ ] Auto-backup scheduler creates and rotates backups according to `database.yaml` retention policy
- [ ] Documentation merged into `docs/FEATURES_OVERVIEW.md` and a per-engine guide under `docs/database/`
- [ ] All steps of HELPER-BINARY-DEVELOPMENT.md checklist verified

## Dependencies

- [#051](../051-git-dispatcher-python-distribution-architecture.md) — Dispatcher/helper architecture (foundation)
- [#100](100-ptx-installer-helper-implementation.md) — `ptx-installer` consumes `assets/packages/*.json`
- `ptx-credential` helper — OS keychain access (already shipped)
- `ptx-container` helper — container-mode lifecycle (already shipped)
- `ptx-mcp` helper — MCP tool surface (already shipped)

## Estimated Effort

Large — approximately 8 weeks across 5 phases. Phase 1 (PostgreSQL + SQLite + skeleton) is the smallest shippable slice (~2 weeks).

## Notes

- Helper lives in `src/helpers/ptx-database/` and ships as the `ptx-database` binary — must be added to dispatcher, Makefile, GoReleaser, install/update lists, and deploy scripts (see HELPER-BINARY-DEVELOPMENT.md).
- Engine drivers should share a common `Driver` interface so adding a new engine in later phases is additive only (no changes to core lifecycle commands).
