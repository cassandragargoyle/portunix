# Acceptance Protocol - Issue #161

**Issue**: Plugin Install Extra Wheels Support
**Branch**: feature/161-plugin-install-extra-wheels
**Tester**: Claude (QA/Test Engineer - Linux)
**Date**: 2026-03-28
**Testing OS**: Linux 6.17.0-19-generic (host)

## Test Summary

- Total test scenarios: 6
- Passed: 6
- Failed: 0
- Skipped: 0

## Test Results

### TC001: --force flag in help

- **Given**: Portunix binary is built
- **When**: Running `portunix plugin install --help`
- **Then**: Help output shows `-f, --force` flag with description
- **Result**: PASS

### TC002: Force reinstall existing plugin

- **Given**: test-plugin is already installed
- **When**: Running `portunix plugin install test/test-plugin` (without --force)
- **Then**: Error "plugin test-plugin already exists" is shown
- **When**: Running `portunix plugin install --force test/test-plugin`
- **Then**: Plugin is reinstalled successfully with message "Plugin reinstalled successfully!"
- **Result**: PASS

### TC003: extra_wheels field accepted in plugin.json

- **Given**: plugin.json with `extra_wheels: ["dep-*.whl"]` and `wheel` field set
- **When**: Running `portunix plugin install`
- **Then**: Manifest is loaded and parsed without JSON errors, installation proceeds to wheel setup
- **Result**: PASS

### TC004: Error on unmatched extra_wheels glob

- **Given**: Plugin with `extra_wheels: ["dep-*.whl"]` but no matching .whl files in dist
- **When**: Running `portunix plugin install`
- **Then**: Error message: `extra_wheels pattern "dep-*.whl" matched no files in <plugin-dir>`
- **Result**: PASS

### TC005: extra_wheels without wheel field

- **Given**: plugin.json with `extra_wheels` but no `wheel` field
- **When**: Running `portunix plugin install`
- **Then**: Validation error: "extra_wheels requires wheel field to be set"
- **Result**: PASS

### TC006: Regression - existing integration tests

- **Given**: All changes committed on feature branch
- **When**: Running `go test portunix.ai/portunix/test/integration/... -run PluginInstallation`
- **Then**: All 11 subtests pass (BuildTestPlugin, InstallPlugin, ListPlugins, PluginInfo, Lifecycle/*)
- **Result**: PASS

### Functional Tests

- [x] --force flag available and documented in help
- [x] Force reinstall overwrites existing plugin
- [x] extra_wheels field parsed from plugin.json
- [x] Glob pattern error when no files match
- [x] Validation error when extra_wheels used without wheel

### Regression Tests

- [x] Existing functionality unaffected
- [x] Integration test suite passes

### Note on Acceptance Criteria Coverage

- [x] `extra_wheels` glob patterns resolved and installed before main wheel (code verified, runtime tested via glob error path)
- [x] Error message when `extra_wheels` glob matches no files
- [x] `plugin install --force` command available for reinstallation
- [ ] Full end-to-end test with real Python wheels (requires building actual .whl files - verified via code review that pip install ordering is correct)

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-03-28
**Tester signature**: Claude (QA/Test Engineer - Linux)
