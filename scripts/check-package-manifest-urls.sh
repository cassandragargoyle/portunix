#!/usr/bin/env bash
# Verify every package manifest in src/helpers/ptx-installer/assets/packages/
# carries non-empty installationDocsUrl and latestVersionUrl fields in metadata
# (ADR-019, issue #080). Designed for CI lint stage and local pre-commit use.

set -euo pipefail

PKG_DIR="${1:-src/helpers/ptx-installer/assets/packages}"

if [[ ! -d "$PKG_DIR" ]]; then
  echo "error: package directory not found: $PKG_DIR" >&2
  exit 2
fi

command -v jq >/dev/null 2>&1 || {
  echo "error: jq is required (apt-get install jq / brew install jq)" >&2
  exit 2
}

fail=0
total=0
for f in "$PKG_DIR"/*.json; do
  total=$((total + 1))
  name=$(jq -r '.metadata.name // empty' "$f" 2>/dev/null || true)
  if [[ -z "$name" ]]; then
    echo "FAIL  $f  — missing or invalid .metadata.name" >&2
    fail=$((fail + 1))
    continue
  fi
  install_url=$(jq -r '.metadata.installationDocsUrl // empty' "$f")
  version_url=$(jq -r '.metadata.latestVersionUrl // empty' "$f")
  if [[ -z "$install_url" ]]; then
    echo "FAIL  $name ($f)  — missing or empty metadata.installationDocsUrl" >&2
    fail=$((fail + 1))
  elif [[ ! "$install_url" =~ ^https?:// ]]; then
    echo "FAIL  $name ($f)  — installationDocsUrl is not http(s): $install_url" >&2
    fail=$((fail + 1))
  fi
  if [[ -z "$version_url" ]]; then
    echo "FAIL  $name ($f)  — missing or empty metadata.latestVersionUrl" >&2
    fail=$((fail + 1))
  elif [[ ! "$version_url" =~ ^https?:// ]]; then
    echo "FAIL  $name ($f)  — latestVersionUrl is not http(s): $version_url" >&2
    fail=$((fail + 1))
  fi
done

if (( fail > 0 )); then
  echo "" >&2
  echo "$fail issue(s) across $total manifest(s). See ADR-019 / issue #080." >&2
  exit 1
fi

echo "OK: $total package manifest(s) carry installationDocsUrl + latestVersionUrl"
