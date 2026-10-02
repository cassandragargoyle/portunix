#!/usr/bin/env bash
#
# This file is part of CassandraGargoyle Community Project
# Licensed under the MIT License - see LICENSE file for details
#
# PTX-TRACE Bash SDK Example - ETL Pipeline
#
# This example demonstrates a complete ETL pipeline with tracing.
# It shows how to:
#   - Create and end a session
#   - Trace stdin/stdout flows with `ptx_trace_pipe`
#   - Wrap external commands with `ptx_trace_exec`
#   - Record ad-hoc events with `ptx_trace_event`
#
# Usage:
#   ./etl-pipeline.sh
#   PTX_TRACE_BIN=/path/to/portunix ./etl-pipeline.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_PATH="${SCRIPT_DIR}/../../bash/ptx-trace.sh"

if [[ ! -f "$SDK_PATH" ]]; then
  printf 'error: SDK not found at %s\n' "$SDK_PATH" >&2
  exit 1
fi

# shellcheck source=../../bash/ptx-trace.sh
source "$SDK_PATH"

WORKDIR="$(mktemp -d -t ptx-trace-etl.XXXXXX)"
trap 'rm -rf "$WORKDIR"' EXIT

# Sample data (mimics a customer export).
cat > "$WORKDIR/customers.csv" <<'CSV'
id,name,email,phone
1,John Doe,john@example.com,+420 777 123 456
2,Jane Smith,jane@test,+1 555 234 5678
3,Bob Wilson,bob@company.org,invalid
4,Alice Brown,alice@example.com,+44 20 1234 5678
CSV

# ============================================
# Start a tracing session for the whole pipeline
# ============================================
ptx_trace_start "customer-import-bash" \
  --source "$WORKDIR/customers.csv" \
  --tag etl --tag bash --tag demo

# ============================================
# PHASE 1: EXTRACT
# Record what the source looked like.
# ============================================
rows="$(($(wc -l < "$WORKDIR/customers.csv") - 1))" # minus header
ptx_trace_event "extract" \
  --status success \
  --input "file=$WORKDIR/customers.csv" \
  --output "rows=$rows"

# ============================================
# PHASE 2: TRANSFORM
# Stream the CSV through `ptx_trace_pipe` to capture bytes/lines, then
# normalize emails to lowercase via awk.
# ============================================
ptx_trace_pipe "normalize" --format csv --tag transform \
  < "$WORKDIR/customers.csv" \
  | awk -F',' 'NR==1 {print; next} {OFS=","; $3=tolower($3); print}' \
  > "$WORKDIR/normalized.csv"

# ============================================
# PHASE 3: VALIDATE
# Run an external validator (here: grep) and trace it. Failure of the
# wrapped command is recorded but does not abort the script (we accept
# "no matches" as a valid outcome).
# ============================================
set +e
ptx_trace_exec "validate-emails" --tag validate --capture-stderr -- \
  grep -E '@[a-z0-9.-]+\.[a-z]{2,}' "$WORKDIR/normalized.csv" \
  > "$WORKDIR/valid.csv"
validate_rc=$?
set -e

ptx_trace_event "validate-summary" \
  --status success \
  --output "valid_rows=$(wc -l < "$WORKDIR/valid.csv"),exit_code=$validate_rc"

# ============================================
# PHASE 4: LOAD
# Pretend to load the data with a wrapped command.
# ============================================
ptx_trace_exec "load-to-db" --tag load -- \
  bash -c 'cat "$1" >/dev/null && echo "loaded $(wc -l < "$1") rows"' _ \
  "$WORKDIR/valid.csv"

# ============================================
# Finish
# ============================================
ptx_trace_end --status completed --summary

printf '\nPipeline finished. Inspect with: portunix trace sessions\n'
