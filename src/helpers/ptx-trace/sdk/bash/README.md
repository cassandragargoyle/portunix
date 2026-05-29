# PTX-TRACE Bash SDK

Lightweight shell helpers around the `portunix trace` CLI for use in Bash
scripts, CI pipelines, and ad-hoc operational tooling.

> Česká verze: [`README.cs.md`](README.cs.md)

## Requirements

- Bash 4.0+ (uses `[[ ... ]]` and `local`)
- `portunix` binary on `PATH` (or set `PTX_TRACE_BIN`)

## Installation

The SDK is a single shell script that is meant to be sourced:

```bash
source /path/to/portunix/src/helpers/ptx-trace/sdk/bash/ptx-trace.sh
```

For convenience, copy it next to your scripts or put it on your `PATH`.

## Configuration

| Variable | Description | Default |
| -------- | ----------- | ------- |
| `PTX_TRACE_BIN` | Path or name of the portunix binary | `portunix` |

## Functions

| Function | Wraps |
| -------- | ----- |
| `ptx_trace_start <name> [flags...]` | `portunix trace start` |
| `ptx_trace_end [flags...]` | `portunix trace end` |
| `ptx_trace_event <op> [flags...]` | `portunix trace event` |
| `ptx_trace_pipe <op> [flags...]` | `portunix trace pipe` |
| `ptx_trace_exec <op> [flags...] -- <cmd>` | `portunix trace exec` |
| `ptx_trace_sessions [flags...]` | `portunix trace sessions` |
| `ptx_trace_view [session] [flags...]` | `portunix trace view` |
| `ptx_trace_stats [session] [flags...]` | `portunix trace stats` |
| `ptx_trace_export <kind> [flags...]` | `portunix trace export <kind>` |
| `ptx_trace_session_id` | prints active session id (or non-zero exit) |

All flag arguments are forwarded verbatim — refer to `portunix trace <cmd>
--help` for the full set.

## Quick start

```bash
#!/usr/bin/env bash
set -euo pipefail
source /path/to/sdk/bash/ptx-trace.sh

ptx_trace_start "nightly-import" --tag etl --tag prod

# Pipe a CSV through the trace recorder while streaming to next stage.
cat customers.csv | ptx_trace_pipe "ingest" --format csv > staged.csv

# Run an external tool with full tracing (exit code is propagated).
ptx_trace_exec "load" --capture-stderr -- \
  psql -h db.example.com -d shop -f load.sql

# Record an ad-hoc event derived from shell logic.
rows="$(wc -l < staged.csv)"
ptx_trace_event "summary" --status success --output "rows=${rows}"

ptx_trace_end --summary
```

## Patterns

### Conditional tracing

Wrap the source line in a guard so scripts also work without portunix
installed:

```bash
if command -v portunix >/dev/null 2>&1; then
  source /path/to/sdk/bash/ptx-trace.sh
else
  ptx_trace_start() { :; }
  ptx_trace_end()   { :; }
  ptx_trace_event() { :; }
  ptx_trace_pipe()  { cat; }
  ptx_trace_exec() {
    shift                                                # drop operation
    while [[ $# -gt 0 && "$1" != "--" ]]; do shift; done # drop optional flags
    [[ "${1:-}" == "--" ]] && shift                      # drop the `--`
    "$@"
  }
fi
```

### Ad-hoc sessions

Both `ptx_trace_pipe` and `ptx_trace_exec` create a one-off session
automatically when no active session exists, so you can use them in
standalone scripts without `start`/`end`:

```bash
some-producer | ptx_trace_pipe "filter" --tag oneshot > out.txt
ptx_trace_exec "build" --tag ci -- make build
```

Use `--session-name` to give the auto-created session a stable name.

### Failures and exit codes

`ptx_trace_exec` propagates the wrapped command's exit code, so it is safe
to use under `set -e`:

```bash
set -euo pipefail
ptx_trace_exec "deploy" -- ./deploy.sh prod
# script aborts here if deploy.sh returned non-zero
```

### Capturing stderr on failure

```bash
ptx_trace_exec "migrate" --capture-stderr -- ./run-migrations.sh
```

When the command fails, the last 4 KB of its stderr is attached to the
trace event under `context.stderr_tail`.

## See also

- [PTX-TRACE README](../../README.md) — overview and CLI reference
- [SDK comparison](../README.md) — Python / Java / TypeScript / Go SDKs
- Example: [`../examples/bash/etl-pipeline.sh`](../examples/bash/etl-pipeline.sh)
