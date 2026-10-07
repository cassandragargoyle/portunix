# Issue #118: System Info pprof Profiling Implementation

**Type**: Enhancement
**Priority**: High
**Status**: ✅ Implemented
**Created**: 2025-12-31
**Completed**: 2026-05-09
**Branch**: feature/118-system-info-pprof-profiling
**Related ADR**: ADR-030

## Summary

Implement pprof profiling support for `portunix system info` command to measure and track performance improvements from native Windows module (Issue #117).

## Background

Issue #117 implements native Windows API calls to replace slow external commands (wmic, powershell). To validate the performance improvement, we need:
1. Baseline measurements before native implementation
2. Profiling tools to analyze performance
3. Benchmark tests for regression detection

## Requirements

### 1. Profiling Flags

Add flags to system info command:
- `--time` - Display execution time
- `--cpuprofile=FILE` - Write CPU profile to file
- `--memprofile=FILE` - Write memory profile to file
- `--trace=FILE` - Write execution trace to file

### 2. Benchmark Tests

Create benchmark tests in `src/app/system/system_test.go`:
- `BenchmarkGetSystemInfo` - Overall system info performance
- `BenchmarkWindowsNative` - Native Windows API performance
- `BenchmarkWindowsExternal` - External command performance (for comparison)
- `BenchmarkAdminCheck` - Admin privilege check
- `BenchmarkVMDetection` - VM detection performance

### 3. Baseline Documentation

Document baseline metrics. Linux baseline measured 2026-05-09 on
AMD Ryzen 9 9950X (`go test -bench=. -benchmem -benchtime=3s`):

| Operation | Linux (current) | Windows external | Windows native | Notes |
|-----------|-----------------|------------------|----------------|-------|
| `GetSystemInfo` | 1386 ms | ~3000 ms | TBD (issue #120) | Aggregate |
| `checkCapabilities` | 1822 ms | TBD | TBD | Primary bottleneck |
| `DetectCertificateBundle` | 680 ms | TBD | TBD | Secondary bottleneck |
| `checkVirtualizationCapabilities` | 30 ms | TBD | TBD | LookPath ×N |
| `GetPodmanVersion` | 9 ms | n/a | n/a | exec overhead |
| Admin check | 33 ns | ~500 ms | TBD | `os.Geteuid` on Unix |

Per-stage table also recorded in ADR-030. Windows numbers will be filled
once issue #120 (native module) lands.

### 4. CI Integration (Future)

Prepare infrastructure for CI benchmark comparison.

## Implementation Steps

1. [x] Add profiling flags to `cmd/system.go`
2. [x] Create `src/app/system/system_test.go` with benchmarks
3. [x] Measure baseline (Linux)
4. [ ] Measure performance with native Windows module — handed off to issue #120
5. [x] Document results in ADR-030
6. [x] Add `--time` flag for quick performance check

## Acceptance Criteria

- [x] `portunix system info --time` shows execution duration
- [x] `portunix system info --cpuprofile=cpu.prof` generates valid profile
- [x] Benchmark tests pass: `go test -bench=. ./src/app/system/...`
- [x] Performance comparison documented (external vs native) — partial
      (Linux baseline complete; Windows native vs external waits on issue #120)
- [ ] Native Windows module shows measurable improvement — moved to issue #120

## Dependencies

- Must be implemented BEFORE merging Issue #117 to main
- Requires Go 1.21+ for pprof improvements

## Files to Modify

- `cmd/system.go` - Add profiling flags
- `src/app/system/system_test.go` - New file with benchmarks
- `docs/adr/030-system-info-performance-profiling.md` - Update with results

## Testing

```bash
# Run benchmarks
go test -bench=BenchmarkGetSystemInfo -benchmem ./src/app/system/...

# Generate CPU profile
./portunix system info --cpuprofile=cpu.prof
go tool pprof cpu.prof

# Quick timing
./portunix system info --time
```

## Notes

- Profiling should work on both Windows and Linux
- External command fallback still available for comparison testing
- Results will validate Issue #117 native implementation benefits
