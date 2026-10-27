# Issue #244: Helpers ptx-database, ptx-github, ptx-plugin-registry, ptx-proxmox missing --help-ai / --help-expert

**Status**: ✅ Implemented
**Priority**: Medium
**Type**: Enhancement
**Component**: ptx-database, ptx-github, ptx-plugin-registry, ptx-proxmox
**Created**: 2026-10-26
**Completed**: 2026-10-26
**GitHub**: #244
**Related**: #163, #50

## Summary

Issue #163 added `--help-ai` and `--help-expert` to the 12 helpers that existed at
that time. Four helpers added later do not implement the flags, so the GitHub
pre-flight check (`scripts/github-01-preflight-check.sh`) fails on every
publication.

## Problem Description

### Current Situation

The pre-flight check runs every `ptx-*` helper with `--help-ai` and
`--help-expert` and reports a failure when the command exits non-zero:

| Helper | --help-ai | --help-expert |
| ------ | --------- | ------------- |
| ptx-database | MISSING | MISSING |
| ptx-github | MISSING | MISSING |
| ptx-plugin-registry | MISSING | MISSING |
| ptx-proxmox | OK | MISSING |

All four helpers use Cobra, which rejects the flags:

```text
$ ./ptx-github --help-ai
Error: unknown flag: --help-ai
```

The other 15 helpers pass the check.

### Reference Implementation

`src/helpers/ptx-ssh/main.go` intercepts `--help-ai` / `--help-expert` before
Cobra parses the arguments; `src/helpers/ptx-proxmox/main.go` already does the
same for `--help-ai` (`showHelpAI`). Follow this pattern.

## Acceptance Criteria

1. `ptx-database`, `ptx-github` and `ptx-plugin-registry` accept `--help-ai` and
   `--help-expert` and exit with code 0
2. `ptx-proxmox` accepts `--help-expert` and exits with code 0
3. `--help-ai` output is machine-readable and lists the helper's commands, in the
   same format as the other helpers
4. `--help-expert` output is detailed (flags, environment variables, examples),
   not a copy of the basic `--help`
5. The pre-flight check reports `All 19 helpers have --help-ai and --help-expert`

## Testing

- Run `bash ./build-with-version.sh v0.0.0+dev.1` and then
  `./ptx-<helper> --help-ai` and `./ptx-<helper> --help-expert` for each of the
  four helpers; check exit code 0 and meaningful output
- Run `./scripts/github-01-preflight-check.sh` and check the helper section
- Regression: `--help` and the regular subcommands of the four helpers still work
