# Acceptance Protocol — Issue #111

**Issue**: PTX-PFT MCP Integration — MCP tools and resources wrapping ptx-pft
**Branch**: `feature/111-pft-mcp-integration`
**Commit under test**: `8516179`
**Tester**: zdendaku (role: Tester / generic)
**Date**: 2026-06-05
**Testing OS**: Windows 11 Pro 10.0.26200 (host)
**Method**: Direct host execution — pure CLI/stdio feature, no installer
side-effects, so container isolation is not required. MCP server driven over
stdio with raw JSON-RPC 2.0 requests piped to `portunix mcp serve`.

## Scope

The change implements Issue #111 (Option A architecture — thin exec wrappers):

- 16 `pft_*` MCP tools (`list`, `show`, `sync`, `link`, `report`, `export`,
  `notify`, `status`, 7× user management, `role_list`) invoking the `ptx-pft`
  binary via exec with a 120 s timeout
- MCP `resources/list` and `resources/read` with the `pft://` scheme
  (`config`, `cache`, `users`, `roles`, `voc/{id}`, `vos/{id}`)
- `initialize` now advertises the `resources` capability
- CLI argument builders separated from execution for unit testability
- Unit tests for all builders, resource URI mapping, and tool coverage

Files changed vs `main`: `src/app/mcp/handlers.go`, `src/app/mcp/server.go`,
`src/app/mcp/pft_handlers.go`, `src/app/mcp/pft_handlers_test.go` (only).

## Test Plan

| Area | Tests |
| ---- | ----- |
| Unit tests | TC-01 |
| Build | TC-02 |
| MCP handshake & discovery | TC-03, TC-04, TC-05 |
| Tool execution (positive) | TC-06 |
| Tool validation errors | TC-07, TC-08, TC-09 |
| Resource read (positive) | TC-10, TC-11 |
| Resource read errors | TC-12 |
| Missing ptx-pft binary | TC-13 |
| Regression (existing tools) | TC-14 |

## Test Cases (Given/When/Then)

### TC-01 — Unit tests pass

- **Given** the `portunix.ai/app/mcp` package on the feature branch
- **When** running `go test . -count=1 -v` in `src/app/mcp`
- **Then** all tests pass — 21 top-level tests including all `buildPft*Args`
  builders (valid/invalid params), `TestPftResourceCommand` (9 URI cases),
  `TestPftToolBuildersCoverAllToolDefinitions` (builder ↔ definition parity),
  and `TestHandlePftToolCallUnknownTool`
- **Result**: ✅ PASS (`ok portunix.ai/app/mcp 1.084s`)

### TC-02 — Build

- **Given** the feature branch
- **When** running `make build` (Go 1.25.0)
- **Then** `portunix.exe` and the deliverables of this issue (`ptx-mcp.exe`,
  `ptx-pft.exe`) build successfully
- **Result**: ✅ PASS for branch deliverables. ⚠️ `make build` as a whole
  stops at `ptx-prompting` — its `go.mod` pins `toolchain go1.24.2` while the
  parent module requires `go >= 1.25.0`. The branch changes no helper files;
  the failure reproduces independently of this change (pre-existing on
  `main`, see Observations).

### TC-03 — `initialize` advertises resources capability

- **Given** a running `portunix mcp serve`
- **When** sending `initialize`
- **Then** the response contains `protocolVersion: 2024-11-05` and
  `capabilities` with both `tools` and `resources`
- **Result**: ✅ PASS

### TC-04 — `tools/list` exposes all 16 pft tools

- **Given** an initialized server
- **When** sending `tools/list`
- **Then** all 16 new tools are present: `pft_list`, `pft_show`, `pft_sync`,
  `pft_link`, `pft_report`, `pft_export`, `pft_notify`, `pft_status`,
  `pft_user_list`, `pft_user_add`, `pft_user_show`, `pft_user_update`,
  `pft_user_role`, `pft_user_link`, `pft_user_remove`, `pft_role_list`
  (total 49 tools; the 4 `pft_config_*` tools are pre-existing from #112)
- **Result**: ✅ PASS

### TC-05 — `resources/list` returns 6 pft:// resources

- **Given** an initialized server
- **When** sending `resources/list`
- **Then** the response lists `pft://config`, `pft://cache`, `pft://users`,
  `pft://roles` (`text/plain`) and `pft://voc/{id}`, `pft://vos/{id}`
  (`text/markdown`)
- **Result**: ✅ PASS

### TC-06 — `tools/call pft_status` executes ptx-pft

- **Given** `ptx-pft.exe` co-located with the server binary
- **When** calling `pft_status` with empty arguments
- **Then** the tool executes `ptx-pft pft status` and returns
  `{"status": "success", "output": "No configuration found. Run 'portunix
  pft configure' first."}` (correct output for an unconfigured workspace)
- **Result**: ✅ PASS

### TC-07 — Missing required parameter is rejected

- **Given** an initialized server
- **When** calling `pft_show` without `id`
- **Then** a JSON-RPC error is returned with detail
  `"id parameter is required"` in the `data` field; ptx-pft is not executed
- **Result**: ✅ PASS

### TC-08 — Invalid enum value is rejected

- **Given** an initialized server
- **When** calling `pft_sync` with `direction: "sideways"`
- **Then** a JSON-RPC error with detail
  `"direction must be one of: pull, push, bidirectional"`
- **Result**: ✅ PASS

### TC-09 — Unknown pft tool is rejected

- **Given** an initialized server
- **When** calling `pft_nonexistent`
- **Then** a JSON-RPC error with detail `"unknown tool: pft_nonexistent"`
- **Result**: ✅ PASS

### TC-10 — `resources/read pft://config`

- **Given** `ptx-pft.exe` available
- **When** reading `pft://config`
- **Then** contents return with `mimeType: text/plain` and the output of
  `ptx-pft pft configure --show` (current configuration text)
- **Result**: ✅ PASS

### TC-11 — `resources/read pft://voc/{id}` routing

- **Given** an unconfigured workspace (no feedback items)
- **When** reading `pft://voc/UC001`
- **Then** the URI maps to `ptx-pft pft show UC001`; ptx-pft reports
  `Item 'UC001' not found: item not found in voc/, vos/, vob/, or voe/
  directories` which is returned as resource text (ptx-pft exits 0)
- **Result**: ✅ PASS (routing verified; positive item read requires a
  configured PFT workspace — see Observations)

### TC-12 — Resource read error paths

- **Given** an initialized server
- **When** reading `pft://unknown`, `pft://voc/` (empty ID), and sending
  `resources/read` without `uri`
- **Then** each returns a JSON-RPC error with a specific detail:
  `"unknown resource URI: pft://unknown"`, `"missing item ID in resource
  URI: pft://voc/"`, `"uri parameter is required"`
- **Result**: ✅ PASS

### TC-13 — Missing ptx-pft binary

- **Given** `ptx-pft.exe` temporarily renamed away
- **When** calling `pft_status` and reading `pft://users`
- **Then** both return the clear error `"ptx-pft is not available: ptx-pft
  binary not found at <path> (ensure ptx-pft is installed alongside
  ptx-mcp)"`; the server keeps running
- **Result**: ✅ PASS (binary restored afterwards)

### TC-14 — Regression of existing tools

- **Given** the modified `tools/call` dispatch (new `pft_` prefix branch in
  the `default` case)
- **When** calling pre-existing tools `echo` and `get_system_info`, and an
  unknown non-pft tool `totally_unknown_tool`
- **Then** `echo` and `get_system_info` return correct results; the unknown
  tool returns `"unknown tool: totally_unknown_tool"` as before
- **Result**: ✅ PASS

## Test Summary

- Total test scenarios: 14
- Passed: 14
- Failed: 0
- Skipped: 0

## Coverage Notes

- Unit: all 16 argument builders incl. negative cases, resource URI mapping,
  builder/definition parity — covered by `pft_handlers_test.go`
- Integration: MCP stdio JSON-RPC round-trips for discovery, execution,
  validation, and binary-missing paths — covered manually (this protocol)
- Not exercised: 120 s command timeout (verified by code review only —
  `exec.CommandContext` + explicit `DeadlineExceeded` check); end-to-end
  with a fully configured PFT workspace + external provider (Fider) sync;
  live Claude Code integration (Issue Phase 6 lists it as a separate item)

## Observations & Recommendations (non-blocking)

1. **Pre-existing build break**: `src/helpers/ptx-prompting/go.mod` pins
   `toolchain go1.24.2`, while the parent module (used via `replace
   ../../..`) requires `go >= 1.25.0`, so `make build` / `make
   build-helpers` aborts before building later helpers (incl. `ptx-pft`).
   Not caused by this branch (0 helper files changed). Recommend a separate
   issue to bump helper `go.mod` toolchains.
2. **Error code semantics**: parameter validation failures surface as
   `-32603 Internal error` with the detail in `data`. `-32602 Invalid
   params` would be semantically cleaner, but the behavior matches the
   existing server convention for all tools — consistent, not blocking.
3. `pft://voc/{id}` for a missing item returns ptx-pft's "not found" text as
   successful resource content (ptx-pft exits 0). Acceptable; AI clients see
   a readable message either way.

## CI Notes

- `go test ./src/app/mcp/...` is sufficient to gate this change in CI
- A future integration smoke test could pipe the JSON-RPC fixtures used in
  TC-03..TC-14 into `portunix mcp serve` (no external services needed; the
  binary-missing path makes it hermetic)

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-06-05
**Tester signature**: zdendaku
