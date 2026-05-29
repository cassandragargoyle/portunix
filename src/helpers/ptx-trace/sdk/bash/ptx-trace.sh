#!/usr/bin/env bash
#
# This file is part of CassandraGargoyle Community Project
# Licensed under the MIT License - see LICENSE file for details
#
# PTX-TRACE Bash SDK
#
# Source this file in your shell scripts to get ergonomic helpers around
# the `portunix trace` CLI:
#
#     source /path/to/sdk/bash/ptx-trace.sh
#     ptx_trace_start "import-customers" --tag etl --tag daily
#     ptx_trace_event "validate" --status success --input "rows=42"
#     cat data.csv | ptx_trace_pipe "ingest" --format csv | next-stage
#     ptx_trace_exec "backup" -- pg_dump -h localhost mydb
#     ptx_trace_end --summary
#
# Configuration:
#   PTX_TRACE_BIN  Override path to portunix binary (default: portunix on PATH)

set -o pipefail

# Resolve the portunix binary once.
_ptx_trace_bin() {
  if [[ -n "${PTX_TRACE_BIN:-}" ]]; then
    printf '%s\n' "$PTX_TRACE_BIN"
  else
    printf '%s\n' "portunix"
  fi
}

# Internal: invoke `portunix trace <subcommand>` forwarding all extra args.
_ptx_trace_run() {
  local bin
  bin="$(_ptx_trace_bin)"
  if ! command -v "$bin" >/dev/null 2>&1 && [[ ! -x "$bin" ]]; then
    printf 'ptx-trace: portunix binary not found (set PTX_TRACE_BIN)\n' >&2
    return 127
  fi
  "$bin" trace "$@"
}

# ptx_trace_start <name> [flags...]
#   Start a new trace session. Flags forwarded as-is to `portunix trace start`.
#   Common flags: --source <file>, --tag <tag> (repeatable), --pii-mask, --alerts.
ptx_trace_start() {
  if [[ $# -lt 1 ]]; then
    printf 'usage: ptx_trace_start <name> [flags...]\n' >&2
    return 2
  fi
  _ptx_trace_run start "$@"
}

# ptx_trace_end [flags...]
#   End the currently active trace session.
#   Common flags: --status completed|failed|cancelled, --summary.
ptx_trace_end() {
  _ptx_trace_run end "$@"
}

# ptx_trace_event <operation> [flags...]
#   Record a single trace event in the active session.
#   Common flags: --input "k=v,k2=v2", --output "k=v", --status success|error,
#                 --error "msg", --tag <tag>.
ptx_trace_event() {
  if [[ $# -lt 1 ]]; then
    printf 'usage: ptx_trace_event <operation> [flags...]\n' >&2
    return 2
  fi
  _ptx_trace_run event "$@"
}

# ptx_trace_pipe <operation> [flags...]
#   Read stdin, pass it through to stdout, and record bytes/lines/duration.
#   Common flags: --format csv|json|..., --tag <tag>, --binary, --session-name <name>.
ptx_trace_pipe() {
  if [[ $# -lt 1 ]]; then
    printf 'usage: ptx_trace_pipe <operation> [flags...]\n' >&2
    return 2
  fi
  _ptx_trace_run pipe "$@"
}

# ptx_trace_exec <operation> [flags...] -- <command> [args...]
#   Run a command, forward stdin/stdout/stderr, and record exit code + duration.
#   Common flags: --tag <tag>, --capture-stderr, --session-name <name>.
ptx_trace_exec() {
  if [[ $# -lt 1 ]]; then
    printf 'usage: ptx_trace_exec <operation> [flags...] -- <cmd> [args...]\n' >&2
    return 2
  fi
  _ptx_trace_run exec "$@"
}

# ptx_trace_sessions [flags...]
#   List trace sessions. Common flags: --limit N, --status active, --format json.
ptx_trace_sessions() {
  _ptx_trace_run sessions "$@"
}

# ptx_trace_view [session-id] [flags...]
#   Show events from a session (defaults to the active one).
ptx_trace_view() {
  _ptx_trace_run view "$@"
}

# ptx_trace_stats [session-id] [flags...]
#   Show aggregated statistics for a session.
ptx_trace_stats() {
  _ptx_trace_run stats "$@"
}

# ptx_trace_export <subcommand> [flags...]
#   Forward to `portunix trace export <subcommand>` (ai|file|db|fulltext).
ptx_trace_export() {
  if [[ $# -lt 1 ]]; then
    printf 'usage: ptx_trace_export <ai|file|db|fulltext> [flags...]\n' >&2
    return 2
  fi
  _ptx_trace_run export "$@"
}

# ptx_trace_session_id
#   Print the ID of the currently active session, if any. Returns non-zero
#   when no active session exists.
ptx_trace_session_id() {
  local out
  out="$(_ptx_trace_run sessions --status active --format json 2>/dev/null)" || return 1
  # Extract first session id without requiring jq; exit non-zero on no match.
  printf '%s\n' "$out" | awk '
    BEGIN { found = 0 }
    /"id":/ {
      match($0, /"id"[[:space:]]*:[[:space:]]*"[^"]+"/);
      if (RSTART > 0) {
        s = substr($0, RSTART, RLENGTH);
        sub(/.*"id"[[:space:]]*:[[:space:]]*"/, "", s);
        sub(/".*/, "", s);
        print s;
        found = 1;
        exit;
      }
    }
    END { exit (found ? 0 : 1) }'
}
