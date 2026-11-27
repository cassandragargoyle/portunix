# Acceptance Protocol - Issue #262 (Plugin Templates Implement the Portunix Contract)

**Issue**: Go, Java and Rust plugin templates do not implement the Portunix plugin contract
**Branch**: `fix/262-plugin-templates-grpc-contract` (round 1: commit `e2eb5f6`,
round 2: commit `6a9069c`)
**Tester**: Claude Code in the tester role (generic), on behalf of zdendaku
**Date**: 2026-11-23 (round 1), 2026-11-24 (round 2)
**Testing OS**: Unit tests, guard test and `make build` on the Linux host — Ubuntu
26.04 (amd64, physical). Template builds, template unit tests and the plugin
service lifecycle in a fresh Podman container `ubuntu:24.04` (Ubuntu 24.04.5 LTS,
`r262`) with toolchains installed by `portunix install`: `go` (Go 1.26), `java
--variant=21` + `maven` (Maven 3.8.7, OpenJDK 21), `protoc` (3.21.12), `rust`
(APT variant, cargo 1.75.0) and `rust --variant=stable` (rustup, Rust 1.100.0;
extra toolchains 1.80 and 1.85 via `rustup toolchain install`), `nodejs`
(regression)

## Scope Note

The branch replaces the protos of the Go, Java and Rust templates with the
`PluginService` part of `src/app/plugins/proto/plugin.proto` (package
`portunix.plugins`, language options kept), implements `Initialize`, `GetInfo`,
`Health`, `Execute`, `ListCommands` and `Shutdown` with plugin info and commands
from `plugin.json`, binds the servers to `127.0.0.1`, adds a CLI mode for
`portunix <plugin> <command>`, extends the guard test to all templates and
updates the language guides.

The templates were copied into the container from the read-only mounted
repository and built there; the host development environment was not modified.

Round 2 retests the Rust template after commit `6a9069c`, which fixes the
blocking finding of round 1 (minimum Rust version, documentation only plus
`rust-version` in `Cargo.toml`); see [Round 2](#round-2).

## Test Summary

- Round 1: 47 scenarios (9 unit/build + 5 template build + 24 service lifecycle +
  3 network + 4 edge cases + 2 regression); 46 passed, 1 failed (T3)
- Round 2: 14 scenarios (Rust template retest and documentation check); 14 passed
- Failed: 0 open (T3 resolved as a documented toolchain requirement in round 2)
- Skipped: 0

## Test Results

### Unit Tests and Build (AC1)

| # | Scenario | Result |
| - | -------- | ------ |
| U1 | `go test -count=1 portunix.ai/app/plugins/... portunix.ai/cmd` (full suites) | ✅ PASS |
| U2 | `TestTemplateProtosMatchContract` — go, java, javascript, rust | ✅ PASS |
| U3 | `TestLanguageTemplateManifests` — go, java, rust | ✅ PASS |
| U4 | `TestJavaScriptTemplateManifest` (regression) | ✅ PASS |
| U5 | Injected drift `rpc Health` → `rpc HealthCheck` in each of the go, java, rust protos fails the matching subtest; files restored | ✅ PASS |
| U6 | Injected drift `package portunix.plugins` → `package rust_plugin_template` fails the rust subtest | ✅ PASS |
| U7 | Changed `option go_package` is accepted (language options may differ) | ✅ PASS |
| B1 | `go vet`, `gofmt -l` on `app/plugins` and the Go template | ✅ PASS |
| B2 | `make build` (main binary + helpers) | ✅ PASS |

### Template Build and Unit Tests (AC4)

| # | Scenario | Result |
| - | -------- | ------ |
| T1 | Go: `go install` protoc-gen-go / protoc-gen-go-grpc, `make build`, `go test ./...` (handlers, services) | ✅ PASS |
| T2 | Java: `mvn clean package` — 8 tests run (`PluginBusinessServiceTest`, `PluginServiceImplTest`), shaded JAR built | ✅ PASS |
| T3 | Rust with `portunix install rust` (Ubuntu default resolves to the APT variant, cargo 1.75.0): `cargo build` | ❌ FAIL |
| T4 | Rust with `portunix install rust --variant=stable` (Rust 1.100.0): `cargo build --release`, `cargo test` (8 tests) | ✅ PASS |
| T5 | Rust 1.85 (rustup toolchain) builds the committed `Cargo.lock` | ✅ PASS |

T3 details: cargo 1.75 rejects the committed lock file (`lock file version 4
requires -Znext-lockfile-bump`); without the lock file the build fails as well,
because current `clap` 4.6 requires the Cargo feature `edition2024` (Rust 1.85+).
The minimum Rust version of the template is therefore 1.85, but the Rust guide
states "Rust 1.70 or later (1.65 minimum)" and its Dockerfile uses
`FROM rust:1.78`, neither of which can build the template.

### Service Lifecycle (AC2, AC3)

Each template was installed from its built copy with `portunix plugin install .`.

| # | Step | Expected | Go | Java | Rust |
| - | ---- | -------- | -- | ---- | ---- |
| L1 | `plugin install .` | installed | ✅ | ✅ | ✅ |
| L2 | `plugin enable` | enabled | ✅ | ✅ | ✅ |
| L3 | `plugin start` | started, no `Unimplemented .../Initialize` | ✅ | ✅ | ✅ |
| L4 | `plugin health` | `✅ Healthy`, message `ok`, uptime reported | ✅ | ✅ | ✅ |
| L5 | `portunix <plugin> hello` | `Hello, World!` greeting, exit 0 | ✅ | ✅ | ✅ |
| L6 | `portunix <plugin> hello Portunix` | `Hello, Portunix!` greeting, exit 0 | ✅ | ✅ | ✅ |
| L7 | `plugin stop` | stopped, no `EOF` | ✅ | ✅ | ✅ |
| L8 | Plugin process after stop | none left | ✅ | ✅ | ✅ |

### Network Binding

| # | Scenario | Expected | Result |
| - | -------- | -------- | ------ |
| N1 | LISTEN sockets with all three plugins running (`/proc/net/tcp*`) | loopback only: Go `127.0.0.1:9001`, Rust `127.0.0.1:9003`, Java `::ffff:127.0.0.1:9002` (IPv4 loopback on a dual-stack socket) | ✅ PASS |
| N2 | Three template plugins running concurrently | all `✅ Healthy` | ✅ PASS |
| N3 | LISTEN sockets and plugin processes after stopping all three | none | ✅ PASS |

### Edge Cases

| # | Scenario | Expected | Result |
| - | -------- | -------- | ------ |
| E1 | `plugin validate .` for go, java, rust | accepted | ✅ PASS |
| E2 | `portunix <plugin> nope` for go, java, rust | `Error: unknown command: nope`, exit 1 | ✅ PASS |
| E3 | Go: start → stop → start → health → stop | restart works, healthy | ✅ PASS |
| E4 | Service mode without `plugin.json` in the working directory | non-zero exit with a manifest error (go, rust clean message; java see observations) | ✅ PASS |

### Regression

| # | Scenario | Expected | Result |
| - | -------- | -------- | ------ |
| R1 | JavaScript template (#260): install (npm), enable, start, health, `hello`, stop, no process left | as before | ✅ PASS |
| R2 | `plugin list` with all four templates | all `stopped` | ✅ PASS |

## Round 2

Commit `6a9069c` changes only `rust/template/Cargo.toml`, `rust/getting-started.md`
and `rust/template/README.md`. The Rust template was copied afresh from the
committed state into `/work/rust2` in the same container; the round 1 Rust plugin
was uninstalled first (`plugin uninstall`, confirmed).

| # | Scenario | Expected | Result |
| - | -------- | -------- | ------ |
| R2-1 | `Cargo.toml` declares `rust-version = "1.85"` | present | ✅ PASS |
| R2-2 | APT cargo 1.75 (T3 retest) | fails with `lock file version 4 requires -Znext-lockfile-bump`, the error listed in the guide's Troubleshooting with the fix `portunix install rust --variant=stable` | ✅ PASS |
| R2-3 | Rust 1.80 | fails with `feature edition2024 is required`, also listed in Troubleshooting | ✅ PASS |
| R2-4 | Rust 1.85 (T5 retest): `cargo build --release`, `cargo test` (8 tests) | builds, tests pass | ✅ PASS |
| R2-5 | Rust 1.100 stable (T4 retest): `cargo build --release`, `cargo test` (8 tests) | builds, tests pass | ✅ PASS |
| R2-6 | Lifecycle L1–L8 (install, enable, start, health, `hello`, `hello Portunix`, stop, no process left) | all steps succeed | ✅ PASS |
| R2-7 | LISTEN socket while running | `127.0.0.1:9003` only | ✅ PASS |
| R2-8 | `portunix rust-plugin-template nope` | `Error: unknown command: nope`, exit 1 | ✅ PASS |
| R2-9 | Guide prerequisites | Rust 1.85+, `portunix install rust --variant=stable`, note on the APT package, `portunix install protoc` | ✅ PASS |
| R2-10 | Guide `Cargo.toml` snippet | includes `rust-version = "1.85"` | ✅ PASS |
| R2-11 | Guide Dockerfile | `FROM rust:1.85 AS builder` | ✅ PASS |
| R2-12 | Guide Quick Start | copies the template; no `plugin create --language` | ✅ PASS |
| R2-13 | Template README | Prerequisites section with Rust 1.85+ and `--variant=stable` | ✅ PASS |
| R2-14 | Guard test `TestTemplateProtosMatchContract` / `TestLanguageTemplateManifests` | pass | ✅ PASS |

Note on required change 2: `rust-version` does not produce its own error message
for toolchains older than 1.85 — cargo 1.75 fails on the lock file and cargo 1.80
on the `edition2024` dependencies before the `rust-version` check runs. Both
errors are documented in the guide's Troubleshooting with the fix, and
`rust-version` still records the minimum version for cargo's MSRV-aware
resolver. Accepted.

## Required Changes (round 1, resolved in `6a9069c`)

1. **Rust minimum version** (T3): state Rust 1.85 or later in
   `docs/plugin-development/languages/rust/getting-started.md` (currently "1.70 or
   later (1.65 minimum)") and in the template README, and recommend
   `portunix install rust --variant=stable` — on Ubuntu 24.04 the default
   `portunix install rust` installs the APT cargo 1.75, which cannot build the
   template
2. Set `rust-version = "1.85"` in the template `Cargo.toml`, so an older toolchain
   fails with a clear message instead of a lock file parse error
3. Raise the Dockerfile base image in the Rust guide from `rust:1.78` to a
   version that builds the template (1.85 or later)

## Observations (non-blocking)

- **Java, missing manifest** (E4): service mode without `plugin.json` ends with an
  uncaught `java.nio.file.NoSuchFileException` stack trace; Go and Rust print a
  one-line error. Catching the exception in `Main` would match the other templates
- **Go, plugin name**: `hello` prints `This is my-go-plugin v1.0.0` (name from
  `config.yaml`) while the manifest name is `go-plugin-template`; the two names
  differ in the Java and Rust templates too (`my-java-plugin`, `my-rust-plugin`
  in `config.yaml`), so it is consistent, but may confuse plugin authors
- **Out of scope, existing behavior**: on Ubuntu, `portunix install rust` picks
  the `apt` variant (Rust 1.75, released 2023-12) over the `stable` rustup
  variant. Candidate for a separate issue on the `rust` package default
- **Out of scope, found by the developer**: the JavaScript guide shows
  `portunix plugin create --language=javascript` (no such option) and a
  Dockerfile with `EXPOSE`, which cannot work with a server bound to `127.0.0.1`
- Windows was not tested (no Windows E2E); the templates contain no
  platform-specific code apart from the Unix-only SIGTERM handler in Rust, which
  falls back to Ctrl+C on other platforms

## Final Decision

Round 1 (2026-11-23, `e2eb5f6`): CONDITIONAL — required changes 1–3 above.

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-11-24
**Tester signature**: Claude Code (tester role)
