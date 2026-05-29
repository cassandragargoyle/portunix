# Acceptance Protocol — Issue #080

**Issue**: Package Metadata URL Tracking Implementation (ADR-019)
**Branch**: `feature/080-package-metadata-url-tracking`
**Commit under test**: `95891b6` (initial pass on `0ba77d2`, scope amended for CI guard `95891b6`)
**PR**: [Gitea PR #1](http://gitea:3000/CassandraGargoyle/portunix/pulls/1)
**Tester**: zdendaku (role: Tester / generic)
**Date**: 2026-05-11 (amended same day for criterion #7)
**Testing OS**:

- Host: Linux 6.17.0-23-generic (Ubuntu 25.10) — host tests sufficient for this metadata-only change
- Container: not required — change is pure data, no installation paths touched

**Build**: `make build` clean from feature branch tip (`95891b6`); both `portunix` and `ptx-installer` helper rebuilt.

## Scope

Implementation of ADR-019 — adds two required URL fields to every package manifest in `src/helpers/ptx-installer/assets/packages/`:

- `installationDocsUrl` — URL to official installation documentation
- `latestVersionUrl` — URL to determine the latest version

Plus a textual update of ADR-019 to reflect the current per-manifest layout (the original ADR referenced the long-removed `assets/install-packages.json`).

**Amendment (same day)** — after the initial PASS, an additional acceptance criterion (#7) was added: a lightweight CI guard ensuring all future manifests carry both URL fields. URL-reachability checks remain explicitly out of CI. Implemented via `scripts/check-package-manifest-urls.sh` and wired into `.github/workflows/test.yml` `lint` job. New test cases TC-11 through TC-18 cover this addition.

**Out of scope** of this issue (potential follow-ups, not blocking acceptance):

- Automated URL-reachability check in CI (intentionally rejected — flaky)
- JSON schema definition / validation in `ptx-installer/registry`
- Generation of an "official sources" doc page

## Test Plan

| Area | Tests |
| ---- | ----- |
| Field presence (acceptance criterion 1) | TC-01 |
| Field non-empty | TC-02 |
| URL format | TC-03 |
| JSON validity | TC-04 |
| Field location (must be in `metadata`) | TC-05 |
| Build regression (criterion 5) | TC-06 |
| Package-loader regression (criterion 6) | TC-07 |
| Installation regression (criterion 5) | TC-08 |
| URL reachability sample (criterion 2) | TC-09 |
| ADR-019 textual update | TC-10 |
| CI guard — happy path (criterion 7) | TC-11 |
| CI guard — missing `installationDocsUrl` | TC-12 |
| CI guard — missing `latestVersionUrl` | TC-13 |
| CI guard — non-http(s) URL rejected | TC-14 |
| CI guard — empty-string URL rejected | TC-15 |
| CI guard — missing directory handled | TC-16 |
| CI wiring — script referenced in workflow | TC-17 |
| CI guard — script executable bit | TC-18 |

## Test Cases (Given/When/Then)

### TC-01 — Both URL fields present in every manifest

- **Given** 63 package manifests under `src/helpers/ptx-installer/assets/packages/`
- **When** counting files where `metadata.installationDocsUrl` and `metadata.latestVersionUrl` exist and are non-null
- **Then** both counts equal 63

**Result**: ✅ PASS — `total=63 has_install=63 has_version=63`

### TC-02 — Both URL fields are non-empty strings

- **Given** the same 63 manifests
- **When** counting files where either URL field is `""` or `null`
- **Then** both counts are 0

**Result**: ✅ PASS — `empty_install=0 empty_version=0`

### TC-03 — URL format is well-formed (http/https)

- **Given** the same 63 manifests
- **When** counting URL values that do NOT start with `http://` or `https://`
- **Then** both counts are 0

**Result**: ✅ PASS — `bad_install=0 bad_version=0`

### TC-04 — All manifests remain valid JSON

- **Given** the same 63 manifests after modification
- **When** running `jq empty` on each file
- **Then** zero parse errors

**Result**: ✅ PASS — `invalid_json=0`

### TC-05 — New fields are in `metadata` block (not root, not `spec`)

- **Given** the same 63 manifests
- **When** verifying both keys exist on `.metadata` object
- **Then** all 63 files pass

**Result**: ✅ PASS — `wrong_location=0`

### TC-06 — Build regression

- **Given** the feature branch tip
- **When** running `make build`
- **Then** main binary + all 18 helper binaries build cleanly

**Result**: ✅ PASS — `All binaries built successfully`

### TC-07 — Package loader regression

- **Given** the rebuilt `portunix` binary
- **When** running `./portunix package list`
- **Then** registry reports `63 packages loaded, 0 errors` and prints all packages

**Result**: ✅ PASS — `packages_loaded=63 errors=0`

### TC-08 — Installation regression (dry-run)

- **Given** the rebuilt `portunix` binary on Linux host
- **When** running `./portunix install nodejs --dry-run`
- **Then** package resolves, variant `apt` selected, no installation attempted

**Result**: ✅ PASS — dry-run output shows correct variant resolution.

### TC-09 — URL reachability (sample)

- **Given** a representative sample of 8 `latestVersionUrl` values across heterogeneous sources (nodejs.org, GitHub releases, python.org, adoptium.net, npmjs.com, docker docs, Microsoft Learn)
- **When** issuing HTTP GET via `curl` with a browser-like User-Agent
- **Then** each URL returns a non-error status (200 expected, 403 acceptable if anti-bot is in front of a real page)

**Result**: ✅ PASS

| URL | HTTP |
| --- | --- |
| https://nodejs.org/dist/latest/ | 200 |
| https://github.com/nektos/act/releases/latest | 200 |
| https://www.python.org/downloads/ | 200 |
| https://adoptium.net/temurin/releases/ | 200 |
| https://www.npmjs.com/package/@anthropic-ai/claude-code | 403 (anti-bot; package confirmed via `registry.npmjs.org` API) |
| https://github.com/astral-sh/uv/releases/latest | 200 |
| https://docs.docker.com/engine/install/ | 200 |
| https://learn.microsoft.com/windows/wsl/install | 200 |

Note: The 403 from npmjs.com is bot protection on the web UI; the package `@anthropic-ai/claude-code` was verified to exist via `https://registry.npmjs.org/@anthropic-ai/claude-code` (returned full metadata incl. `dist-tags.latest`). The URL is correct.

### TC-10 — ADR-019 textual update

- **Given** the change includes an ADR-019 rewrite
- **When** grepping the ADR for old and new path references
- **Then** old path `assets/install-packages.json` appears only once (inside the explicit "Historical note" block) and the new path `src/helpers/ptx-installer/assets/packages` appears multiple times

**Result**: ✅ PASS — `assets/install-packages.json` × 1 (in historical note), `src/helpers/ptx-installer/assets/packages` × 3.

### TC-11 — CI guard happy path

- **Given** all 63 manifests carry both URL fields
- **When** running `./scripts/check-package-manifest-urls.sh`
- **Then** exit code is 0 and message reports `OK: 63 package manifest(s) carry installationDocsUrl + latestVersionUrl`

**Result**: ✅ PASS

### TC-12 — CI guard rejects missing `installationDocsUrl`

- **Given** a temp copy of `nodejs.json` with `metadata.installationDocsUrl` deleted
- **When** running the script against the temp directory
- **Then** exit code is 1 and stderr names the file + field

**Result**: ✅ PASS — exit code 1, message `FAIL nodejs … missing or empty metadata.installationDocsUrl`.

### TC-13 — CI guard rejects missing `latestVersionUrl`

- **Given** a temp copy of `nodejs.json` with `metadata.latestVersionUrl` deleted
- **When** running the script against the temp directory
- **Then** exit code is 1 and stderr names the file + field

**Result**: ✅ PASS — exit code 1, message `FAIL nodejs … missing or empty metadata.latestVersionUrl`.

### TC-14 — CI guard rejects non-http(s) URLs

- **Given** a temp copy of `nodejs.json` with `installationDocsUrl` set to `ftp://example.com/install`
- **When** running the script against the temp directory
- **Then** exit code is 1 and stderr explains the URL is not http(s)

**Result**: ✅ PASS — exit code 1, message `installationDocsUrl is not http(s): ftp://example.com/install`.

### TC-15 — CI guard rejects empty-string URLs

- **Given** a temp copy of `nodejs.json` with `installationDocsUrl` set to `""`
- **When** running the script against the temp directory
- **Then** exit code is 1 (treated as missing)

**Result**: ✅ PASS — exit code 1.

### TC-16 — CI guard handles missing directory gracefully

- **Given** a path that does not exist
- **When** running `./scripts/check-package-manifest-urls.sh /nonexistent/path`
- **Then** exit code is 2 and stderr explains the directory was not found (distinct from the data-failure exit code 1)

**Result**: ✅ PASS — exit code 2.

### TC-17 — CI wiring: lint job references the script

- **Given** `.github/workflows/test.yml`
- **When** grepping for `check-package-manifest-urls.sh`
- **Then** exactly one reference exists (in the `lint` job's `Validate package manifest URL metadata (ADR-019)` step)

**Result**: ✅ PASS — 1 occurrence inside the `lint` job.

### TC-18 — Script is executable

- **Given** the committed script under `scripts/`
- **When** inspecting permissions
- **Then** the user execute bit is set so CI can invoke it directly without `bash` prefix

**Result**: ✅ PASS — mode `-rwxrwxr-x`.

## Test Summary

- Total test scenarios: **18** (10 original + 8 added in amendment)
- Passed: **18**
- Failed: **0**
- Skipped: **0**

## Test Results

### Functional Tests

- [x] Feature works as specified (all 7 acceptance criteria from issue #080 met)
- [x] Acceptance criteria met (including the amended criterion #7: CI guard)
- [x] Edge cases handled — JSON edge cases (missing field, empty string, non-http URL, missing directory) covered by TC-12 through TC-16; Windows-only packages and internal `vox-deps`/`virt`/`ca-certificates` packages all carry sensible URLs

### Regression Tests

- [x] Existing functionality unaffected (build, package list, dry-run all green)
- [x] Cross-platform compatibility verified — change is metadata-only, JSON keys are platform-agnostic and are parsed by Go's `encoding/json` which silently tolerates additional fields. No Linux-specific or Windows-specific code touched.

### Notes / Known issues NOT caused by this change

- `go test ./src/helpers/ptx-installer/engine/...` shows 3 pre-existing failures on Linux host: `TestValidateWindowsPath`, `TestResolveWindowsDataRoot_ExplicitFlag`, `TestResolveWindowsDataRoot_ExplicitFlag_InsufficientSpace`. These test Windows-only code paths (drive-letter parsing) and fail identically on `main` (verified by checking out `main` and re-running). **Not a regression caused by #080.**

## Final Decision

**STATUS**: **PASS** (re-confirmed after scope amendment)

**Approval for merge**: **YES**

**Date**: 2026-05-11 (initial), re-confirmed 2026-05-11 after criterion #7 added
**Tester signature**: zdendaku
