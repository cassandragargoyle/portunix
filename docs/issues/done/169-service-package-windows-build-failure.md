# Issue #169: Service Package Windows Build Failure

**Type**: Bug Fix
**Priority**: High
**Labels**: bug, build, windows, cross-platform, service-orchestration, syscall
**Status**: ✅ Implemented
**Created**: 2026-04-08

## Problem Description

The `src/app/service/` package uses Linux-only syscalls without build constraints,
causing compilation failure on Windows:

```text
# portunix.ai/app/service
src\app\service\orchestrator.go:315:3: unknown field Setpgid in struct literal of type "syscall".SysProcAttr
src\app\service\state.go:80:20: undefined: syscall.Flock
src\app\service\state.go:80:52: undefined: syscall.LOCK_EX
src\app\service\state.go:83:16: undefined: syscall.Flock
src\app\service\state.go:83:48: undefined: syscall.LOCK_UN
```

## Root Cause

Two files in `src/app/service/` use Linux-only syscall functions without
`//go:build` constraints:

1. **`orchestrator.go:315`** — `syscall.SysProcAttr{Setpgid: true}` (process group
   detachment, not available on Windows)
2. **`state.go:80-83`** — `syscall.Flock()` with `LOCK_EX`/`LOCK_UN` (file locking,
   not available on Windows)

Neither file has a `//go:build linux` constraint, so Go attempts to compile them
on all platforms.

## Impact

- `make build` and `go build` fail on Windows
- Blocks all Windows development and local testing
- Introduced by Issue #159 (Service Orchestration Command Group)

## Proposed Solution

Split platform-specific code into separate files using Go build constraints:

### orchestrator.go — Setpgid

Create `proc_unix.go` and `proc_windows.go`:

- **Unix**: `&syscall.SysProcAttr{Setpgid: true}` (detach process group)
- **Windows**: `nil` or `&syscall.SysProcAttr{CreationFlags: 0x00000008}` (DETACHED_PROCESS)

### state.go — Flock

Create `lock_unix.go` and `lock_windows.go`:

- **Unix**: `syscall.Flock()` with `LOCK_EX`/`LOCK_UN`
- **Windows**: `LockFileEx`/`UnlockFileEx` via `golang.org/x/sys/windows` or a no-op fallback

### Alternative

Add `//go:build !windows` to both files and create minimal stub files for Windows
with `//go:build windows` that provide the same interface but with Windows-compatible
implementations or no-ops.

## Acceptance Criteria

- [ ] `make build` succeeds on Windows
- [ ] `make build` succeeds on Linux (no regression)
- [ ] Service orchestration works on Linux
- [ ] Windows build produces functional binary (service features may be limited)

## Affected Files

- `src/app/service/orchestrator.go` — line 315
- `src/app/service/state.go` — lines 80-83

## References

- Related: Issue #159 (Service Orchestration Command Group) — introduced the affected code
