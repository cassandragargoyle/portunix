# Security Scan

## Overview

The **Security Scan** job in GitHub Actions runs two scanners on every Go module in the
repository and fails on any finding:

- [gosec](https://github.com/securego/gosec) — static analysis of the Portunix source code
- [govulncheck](https://go.dev/doc/security/vuln/) — known vulnerabilities in dependencies and
  the standard library, see [Dependency Scan](#dependency-scan)

This guide describes the exact invocation, the rules that are excluded on purpose, and how to
annotate an intentional pattern.

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
| Per-user state (plugin `registry.json`, its lock, plugin `plugin.log`) | `0600` / `0750` | Written through a temporary file and rename where readers run concurrently; a symlinked file is updated at its target |
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

## Dependency Scan

`scripts/govulncheck-scan.py` runs `govulncheck ./...` in every module, discovered and joined
into a temporary Go workspace the same way as for gosec. It queries the Go vulnerability
database (`vuln.go.dev`), needs no account and no credentials.

A module fails when govulncheck reports a vulnerability whose vulnerable function the code
calls. Vulnerabilities in imported packages or required modules that the code does not call
are not reported. Each finding lists the OSV ID, the module, the found and the fixed version:

```text
FAIL  src/helpers/ptx-ssh  - 1 called vulnerability(ies)
        GO-2026-5013  golang.org/x/crypto v0.47.0 -> v0.52.0  ...
```

| Option | Purpose |
| ------ | ------- |
| `--report-dir DIR` | Keep the raw govulncheck JSON output per module (`govulncheck-<module>.json`), used by CI for the artifact |
| `REPO_DIR` | Scan another checkout instead of the one the script lives in |

### Fixing a Finding

- **Module dependency** — raise the module to the fixed version in every `go.mod` that requires
  it, then run `go mod tidy` in that module and in the helper modules that point at it through
  `replace` (otherwise `make build` stops with `updates to go.mod needed`)
- **Standard library** (`stdlib`) — the finding belongs to the Go toolchain the scan runs with,
  upgrade the local Go installation. The script prints the minimum Go version that fixes all
  stdlib findings. The CI job uses the latest stable Go, the release is built with the local Go,
  so the local toolchain must be up to date before a release

To inspect a finding in detail, including the call stack, run govulncheck in the module:

```bash
cd src/helpers/ptx-ssh && govulncheck -show traces ./...
```

## Running Locally

```bash
# Install the scanners
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

# Both scans
make security

# One scan only, with per-finding output
python3 scripts/gosec-scan.py
python3 scripts/govulncheck-scan.py

# Without gosec — runs it in a golang container through portunix container
./scripts/github-01-preflight-check.sh
```
