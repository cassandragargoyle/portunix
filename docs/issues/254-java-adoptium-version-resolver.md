# Issue #254: Java: resolve latest Temurin release via Adoptium API with pinned fallback

**Status**: Open
**Priority**: Medium
**Type**: Enhancement
**Component**: ptx-installer
**Created**: 2026-11-08
**GitHub**: #254

## Summary

All Java variants in `java.json` are pinned to fixed Temurin builds that are
out of date, and the installer has a single download source per variant.
Resolve the newest release of each feature version via the Adoptium API and
keep the pinned build as a fallback.

## Problem Description

- `portunix install java --variant=21` downloads `21.0.8+9`; the current
  release is `21.0.13+9` (Adoptium API, 2026-11-08). Variants 8, 11 and 17
  are pinned the same way.
- When the GitHub release download fails (e.g. a transient `504 Gateway
  Timeout` from GitHub / Adoptium), there is no alternative source and the
  install fails.

The engine already provides the building blocks: `versionResolver` and
`fallbackVariant` (`src/helpers/ptx-installer/engine/version_resolver.go`),
currently used by `python.org` and `git`.

## Proposed Solution

- Add an `adoptium` version resolver using
  `https://api.adoptium.net/v3/assets/latest/{feature}/hotspot?os={os}&architecture={arch}&image_type=jdk`
  — returns version, download link and SHA-256 checksum
- Variants `8`, `11`, `17`, `21` (Linux and Windows) use the resolver; the
  current pinned builds become `fallbackVariant` targets (updated to the
  latest builds at implementation time)
- Use the checksum returned by the API for verification
- Consider retry with backoff for transient HTTP 5xx download errors (generic
  in the download layer, not Java specific)

## Acceptance Criteria

1. `portunix install java --variant=21 --dry-run` shows the newest 21.x
   release returned by the Adoptium API
2. When the API is unreachable, the installer falls back to the pinned variant
   with a clear message
3. Downloads resolved via the API are verified with the API-provided SHA-256
4. `extractTo` and `update-alternatives` postInstall paths follow the resolved
   version
5. Pinned fallback builds updated to the latest release at implementation time
6. Unit tests for the resolver with a mocked API response

## Testing

- Unit tests: resolver parsing, arch mapping (`x64`, `aarch64`), fallback on
  HTTP error
- Container test (`ubuntu:24.04` with `ca-certificates`):
  `portunix install java --variant=21`, then `java -version`
- Simulated outage (invalid API host) → fallback variant installed
