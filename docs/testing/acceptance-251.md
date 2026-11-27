# Acceptance Protocol - Issue #251 (plugin list sorting, DESCRIPTION width, --json)

**Issue**: portunix plugin list: sort plugins alphabetically and widen DESCRIPTION column
**Branch**: `feature/251-plugin-list-alphabetical-sort` (commit `c1febd5`; retest 1 on `3701aa3`)
**Tester**: Claude Code in the tester role, on behalf of zdendaku
**Date**: 2026-11-06
**Testing OS**: Ubuntu 26.04.1 LTS, kernel 7.0.0-42, amd64 (host); Go 1.26.8

The change only affects how `portunix plugin list` reads and prints the plugin registry. No
software was installed and no host configuration was changed, so the tests ran on the host.
CLI scenarios used a synthetic registry in a temporary `HOME`, so the real plugin registry
was only read, never modified. The working tree was clean at the end.

## Scope Note

**Independence note**: the change was made by Claude Code in the developer role in the same
session. The scenarios below were designed separately in the tester role.

**Not executed**:

- Run on Windows — the change is platform-neutral Go code (sorting, `fmt` width, flag
  parsing) with no OS-specific branches
- Live gRPC call of `ListPluginsForPlatform` — the server copies the registry result into the
  response in order (`registry_server.go`, reviewed); the registry order is covered by TC02
  and the `portunix.ai/app/plugins/server` tests pass

## Test Summary

- Total test scenarios: 10
- Passed: 10
- Failed: 0
- Observations: O1 (pre-existing, not caused by this change, not blocking)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | sorting of both registry functions, `pluginNameLess`, truncation width, `--json` / `--output` resolution; `-race`, `go vet`, `gofmt` | host |
| Fault injection | sorting and conflict check disabled, tests must fail | host (reverted with `git checkout`) |
| CLI e2e | synthetic registry with mixed-case names, long and multi-byte descriptions, platform declarations | host, temporary `HOME` |
| CLI e2e | empty registry | host, temporary `HOME` |
| Regression | real registry (26 plugins), read only | host |

Synthetic registry: `zeta`, `Beta`, `alpha`, `foo`, `Foo`, `gamma`, `Delta` (inserted in this
order); all except `gamma` declare the `synapse` platform; `alpha` has a description of exactly
45 characters, `Beta` a long Czech description, `zeta` a short one.

## Test Cases

### TC01: unit tests and static checks (AC8, AC13)

- **When** `go test -race` runs on `portunix.ai/app/plugins/manager`, `portunix.ai/app/plugins/server`
  and `portunix.ai/cmd`, followed by `go vet` and `gofmt -l`
- **Then** everything passes and `gofmt` reports no files

Result: **PASS** — all three packages `ok`; new tests `TestListPlugins_SortedAlphabetically`,
`TestListPluginsForPlatform_SortedAlphabetically`, `TestPluginNameLess`,
`TestTruncateString_DescriptionWidth`, `TestResolvePluginListFormat` pass; vet and gofmt clean.

### TC02: the tests detect a regression (fault injection)

- **Given** the sort comparators replaced by `return false`, and the `--json` conflict check
  disabled
- **When** the new tests run
- **Then** they fail

Result: **PASS** — both sorting tests fail with the random map order (e.g.
`[Foo gamma Delta zeta Beta alpha foo]`); `TestResolvePluginListFormat` fails in the
`json flag with output yaml/table` cases. Sources restored with `git checkout`.

### TC03: alphabetical, stable table output (AC1, AC4, AC5)

- **Given** the synthetic registry
- **When** `portunix plugin list` runs five times
- **Then** the output is identical and ordered `alpha, Beta, Delta, Foo, foo, gamma, zeta`

Result: **PASS** — identical MD5 for all five runs; order as expected (case-insensitive,
`Foo` before `foo`).

### TC04: DESCRIPTION column width 45 (AC6, AC7)

- **When** the compact table is printed
- **Then** the header and separator columns are 45 wide, a 45-character description is shown
  in full, longer ones are cut to 45 characters ending with `...`, short ones are padded

Result: **PASS** — `alpha` (exactly 45) shown in full; `Beta`, `Delta`, `Foo`, `foo`, `gamma`
cut to 45 runes with `...`; the Czech description is cut without broken UTF-8; `zeta` shown
as `short`. The width is the single constant `pluginListDescriptionWidth`.

### TC05: `--verbose` keeps the full description (AC6, #188 regression)

- **When** `portunix plugin list -v` runs
- **Then** the DESCRIPTION column is absent and the full description is wrapped below each row

Result: **PASS** — full texts printed, wrapped to the terminal width, in alphabetical order.

### TC06: JSON and YAML order, `--json` equals `--output json` (AC2, AC8)

- **When** `--json`, `-o json` and `-o yaml` run
- **Then** all return the same alphabetical order and `--json` is byte-identical to `-o json`

Result: **PASS** — JSON `['alpha', 'Beta', 'Delta', 'Foo', 'foo', 'gamma', 'zeta']`, YAML in the
same order, `diff` of `--json` and `-o json` empty. `--json -a` returns all 7 plugins.

### TC07: platform query order (AC3, AC8)

- **When** `--platform synapse` runs in table, `--json` and `-o yaml` form, and with
  `--feature a --json`
- **Then** `gamma` is excluded and the remaining plugins are in alphabetical order, stable
  across runs

Result: **PASS** — table and JSON `['alpha', 'Beta', 'Delta', 'Foo', 'foo', 'zeta']`; three YAML
runs identical; `--feature a` returns the same list.

### TC08: `--json` conflicts with a non-JSON `--output` (AC9)

- **When** `--json -o yaml`, `--json -o table`, `--json --platform synapse -o yaml`,
  `--json -o json` and `--json=false -o yaml` run
- **Then** the first three fail with a clear error, the last two succeed

Result: **PASS** — `Error: --json conflicts with --output yaml` / `... table`, exit code 1;
`--json -o json` exit 0; `--json=false -o yaml` prints YAML.

### TC09: empty registry (AC10)

- **Given** an empty temporary `HOME`
- **When** `plugin list`, `--json`, `-o json`, `-o yaml` and `--platform synapse --json` run
- **Then** the table prints `No plugins installed.`, the structured outputs print `[]`

Result: **PASS** — exactly as expected for all five invocations.

### TC10: help and real registry (AC11, regression)

- **When** `portunix plugin list --help` runs, and `portunix plugin list --json` runs three
  times against the real registry (26 plugins)
- **Then** the help documents `--json`, and the output is identical across runs

Result: **PASS** — help lists `--json  Output in JSON format (shorthand for --output json)` and
the long description mentions alphabetical ordering; three identical MD5 hashes.

## Observations

### O1: unknown `--output` values silently fall back to the table (pre-existing)

`portunix plugin list -o xml` prints the table instead of an error, and the format value is
case-sensitive (`--json -o JSON` reports a conflict, `-o JSON` alone prints the table). This
behavior existed before #251 and was outside its original acceptance criteria. Not blocking.

### Resolution of O1

O1 was added to the issue as AC12 (the unit-test criterion moved to AC13) and fixed on the
same branch in `3701aa3`: `resolvePluginListFormat()` lower-cases and trims `--output`,
rejects anything other than `table`, `json`, `yaml`, and only then applies the `--json`
conflict check, so the default listing and the `--platform` query share the validation.

## Retest 1

**Commit**: `3701aa3` (fix for O1, AC12)
**Date**: 2026-11-06
**Testing OS**: same host as above

The retest uses the same synthetic registry and empty `HOME` as the first run, rebuilt with
`make build`.

- Total test scenarios: 5 (R1–R5)
- Passed: 5
- Failed: 0

### R1: unit tests and static checks (AC13)

- **When** `go test -race` runs on `portunix.ai/app/plugins/manager`, `portunix.ai/app/plugins/server`
  and `portunix.ai/cmd`, followed by `go vet` and `gofmt -l`
- **Then** everything passes

Result: **PASS** — all three packages `ok`; `TestResolvePluginListFormat` runs 12 cases
(6 new: uppercase, mixed case, `--json -o JSON`, unknown, `--json` + unknown, empty);
vet and gofmt clean.

### R2: the new tests detect a regression (fault injection)

- **Given** `resolvePluginListFormat()` with lower-casing removed and unknown formats accepted
- **When** `TestResolvePluginListFormat` runs
- **Then** it fails

Result: **PASS** — all 6 new cases fail; source restored with `git checkout`.

### R3: unknown formats are rejected (AC12)

- **When** `-o xml`, `-o csv`, `-o ""`, `--platform synapse -o xml` and `--json -o xml` run
- **Then** each fails with an error listing the allowed formats

Result: **PASS** — exit code 1 with
`Error: invalid --output "xml": allowed formats are table, json, yaml` (and the same for
`csv` and the empty value), in both the default and the `--platform` mode.

### R4: formats are case-insensitive (AC12)

- **When** `-o JSON`, `-o Yaml`, `-o TABLE`, `--json -o JSON`, `--json -o Json`,
  `--platform synapse -o YAML` and `-o " json "` run
- **Then** they succeed and match the lower-case variants

Result: **PASS** — all exit 0; `diff` against `-o json`, `-o yaml` and the default table is
empty; surrounding spaces are ignored.

### R5: regression of TC03–TC10

- **When** the sorting, `--json`, conflict, empty-registry and `--verbose` checks are repeated
- **Then** the results are unchanged

Result: **PASS** — `--json -o yaml` / `--json -o table` still fail with the conflict error;
the table output has the same MD5 as in TC03 and the same order; JSON and the platform query
keep the alphabetical order; the DESCRIPTION column is still 45 wide; the empty registry
prints `No plugins installed.` / `[]` (also for `-o YAML`); the real registry gives the same
`--json` MD5 as in TC10; `-v` prints full descriptions.

## Final Decision

**STATUS**: PASS

All acceptance criteria AC1–AC13 are met. The pre-existing observation O1 was added to the
issue as AC12, fixed in `3701aa3` and passes Retest 1.

**Approval for merge**: YES
**Date**: 2026-11-06
**Tester signature**: Claude Code (tester role)
