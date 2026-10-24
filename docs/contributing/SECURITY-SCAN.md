# Security Scan

## Overview

The **Security Scan** job in GitHub Actions runs [gosec](https://github.com/securego/gosec)
on every Go module in the repository and fails on any finding. This guide describes the exact
invocation, the rules that are excluded on purpose, and how to annotate an intentional pattern.

## Invocation

CI, `make security` and `scripts/github-01-preflight-check.sh` all run the same script:

```bash
python3 scripts/gosec-scan.py
```

The script is the single source of truth for the gosec command line. For every module it runs:

```bash
gosec -exclude-generated -exclude=G204,G304,G702,G703 ./...
```

| Option | Purpose |
| ------ | ------- |
| `--report-dir DIR` | Keep one JSON report per module (`gosec-<module>.json`), used by CI for the artifact |
| `--summary` | Print counts per rule and list only HIGH findings, used by the preflight check |
| `REPO_DIR` | Scan another checkout instead of the one the script lives in |

Change the excluded rules only in `GOSEC_EXCLUDE` in `scripts/gosec-scan.py`.

## Excluded Rules

Four rules are excluded globally because they describe what Portunix does by design:

| Rule | Meaning | Reason for exclusion |
| ---- | ------- | -------------------- |
| G204 | Subprocess launched with variable | Portunix orchestrates external tools (package managers, container runtimes, SSH, installers) — the command line is always built from variables |
| G304 | File inclusion via variable path | Portunix is a CLI that reads files the user points it at (configs, playbooks, package definitions) |
| G702 | Command injection via taint analysis | Taint-analysis variant of G204 — container, SSH and installer commands are built from the user's arguments |
| G703 | Path traversal via taint analysis | Taint-analysis variant of G304 — paths come from the user's arguments and configuration |

Generated code (`*.pb.go`, ANTLR parser) is skipped with `-exclude-generated`, it cannot be
annotated because it is overwritten on regeneration.

No other rule may be excluded globally without updating this table.

## Suppressing a Single Finding

When a finding is an intentional pattern, annotate the line and state the reason:

```go
return os.MkdirAll(path, 0755) // #nosec G301 -- installed software must be readable by all users
```

Rules for annotations:

- Always name the rule (`#nosec G301`), never use a bare `#nosec`
- Always add a reason after `--`
- **HIGH** severity findings are fixed in code, see the exceptions below
- `//nolint:gosec` is only honoured by golangci-lint, the Security Scan job ignores it

### HIGH Findings That May Be Annotated

Only two kinds of HIGH findings may carry a `#nosec`, everything else is fixed in code:

| Rule | Allowed case | Example |
| ---- | ------------ | ------- |
| G101 | False positive — the string is a name or a documentation placeholder, not a secret | Name of an environment variable, file name, connection string format in help text |
| G402 | TLS verification is disabled on explicit user request or for a local probe | Proxmox profile with `VerifyTLS` off, handshake test of the locally started MCP server |

## File and Directory Permissions

| Data | Mode | Notes |
| ---- | ---- | ----- |
| Files that may hold credentials (configs with tokens, compose files, caches) | `0600` | No annotation needed |
| Log files and log directories | `0600` / `0750` | No annotation needed |
| Installed software, build artifacts, public assets | `0644` / `0755` | `#nosec G301`/`G306` with reason |
| Directories mounted into containers | `0755` | Container users must traverse them, `#nosec G301` with reason |
| VM disks, ISO images and Windows Sandbox folders | `0644` / `0755` | Read by hypervisors and sandbox users, `#nosec G301`/`G306` with reason |
| Scaffolded project sources and output files the user asked for | `0644` / `0755` | Regular source tree permissions, `#nosec G301`/`G306` with reason |
| Shell startup files, desktop entries, Docker daemon configuration | `0644` / `0755` | World-readable by convention, `#nosec G302`/`G306` with reason |
| Executables (installed binaries, generated run scripts) | `0755` / `0700` | `#nosec G302`/`G306` with reason, scripts run only by the owner use `0700` |

## Scope

`gosec ./...` stops at nested module boundaries, so `scripts/gosec-scan.py` discovers every
`go.mod` (the same way as `GO_MODULES` in the `Makefile`, without `docs/` and `docs-site/`) and
scans each module separately. A new helper module is picked up without any configuration.
The root module scan also excludes `docs/` and `docs-site/`, which hold template code, not
product code.

Modules that the root `go.mod` points at through local `replace` directives (`src/app`,
`src/cmd`, `src/app/sandbox`, …) are scanned inside a temporary Go workspace, because several
of them have stub `go.mod` files that do not resolve on their own. A module whose packages
cannot be loaded fails the scan instead of being skipped silently.

gosec analyses the files of the platform it runs on — files with a Windows build constraint
are not covered by the Linux CI job.

## Running Locally

```bash
# With gosec installed
make security

# Scan only, with per-finding output
python3 scripts/gosec-scan.py

# Without gosec — runs it in a golang container through portunix container
./scripts/github-01-preflight-check.sh
```
