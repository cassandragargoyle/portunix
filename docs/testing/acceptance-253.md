# Acceptance Protocol - Issue #253 (PATH_APPEND Applied to Package-Manager Variants Writes Non-Existent Paths)

**Issue**: PATH_APPEND applied to package-manager variants writes non-existent paths
**Branch**: `fix/253-path-append-package-manager-variants` (commit `2358807`)
**Tester**: Claude Code in the tester role (generic), on behalf of zdendaku
**Date**: 2026-11-08
**Testing OS**:

- Linux host — unit tests, `go vet`, `make build`, nothing installed
- Ubuntu 24.04.5 LTS (container `ubuntu:24.04`, user `root`, started with
  `portunix container run`) — all installation scenarios

## Scope Note

The fix adds `effectivePathAppend` to `ptx-installer`: a variant-level
`environment.PATH_APPEND` wins (an empty value disables PATH wiring), package-manager
types (`deb`, `apt`, `dnf`, `yum`, `snap`, `pacman`) skip the platform value, all
other types keep using it. `VariantSpec` gained an optional `environment` map.

No baseline run from `main` was made: the issue records the faulty
`/opt/maven/bin` entry, and the unit test `TestEffectivePathAppend_Manifests`
covers the decision for the shipped `maven`, `nodejs` and `rust` manifests.

Container preparation: the `ubuntu:24.04` image has no CA certificates, so
`ca-certificates` was installed with `apt-get` before each run (the first
attempt failed with `x509: certificate signed by unknown authority`, one retry
was needed after a network `connection reset by peer`). Neither is related to
the issue.

## Test Summary

- Total test scenarios: 9 (3 unit + 2 host checks + 4 container)
- Passed: 9
- Failed: 0
- Skipped: 0
- Pre-existing findings outside the issue scope: 2 (F1, F2)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | `effectivePathAppend` per install type, variant override, shipped manifests | host |
| Build | `go vet`, `make build` | host |
| E2E | `portunix install` of affected packages, PATH files checked in a new login shell | container |

## Test Cases

### Unit and Build (host)

| ID | Given / When / Then | Result |
| -- | ------------------- | ------ |
| U1 | Given a platform `PATH_APPEND` / when the type is `tar.gz`, `zip`, `download`, `script`, `msi`, `exe`, `chocolatey`, `winget` / then it is applied; for `deb`, `apt`, `dnf`, `yum`, `snap`, `pacman` it is skipped (`TestEffectivePathAppend_PerInstallType`) | ✅ PASS (AC1, AC4, AC6) |
| U2 | Given a variant `environment` / when it defines `PATH_APPEND` / then it wins for `apt` and `tar.gz`, an empty value disables it, a variant without the key falls back to the platform rules (`TestEffectivePathAppend_VariantOverride`) | ✅ PASS (AC2) |
| U3 | Given the shipped registry / then `maven/apt`, `nodejs/apt,dnf,pacman`, `rust/apt,dnf,pacman` get no entry, `maven/latest` keeps `/opt/maven/bin`, `rust/stable` keeps `$HOME/.cargo/bin` (`TestEffectivePathAppend_Manifests`) | ✅ PASS (AC6) |
| H1 | `go test ./...` and `go vet ./...` in `src/helpers/ptx-installer` | ✅ PASS |
| H2 | `make build` — main binary and all helpers | ✅ PASS |

### Container (Ubuntu 24.04)

| ID | Command | Observed | Result |
| -- | ------- | -------- | ------ |
| T1 | `portunix install maven` | Variant `apt` selected. Dependency `java` (variant `8`, `tar.gz`) installed first. The `apt` step prints no `Adding to User PATH`, exit 0. `~/.profile` has no `/opt/maven/bin`, `/opt/maven` does not exist. New login shell: `mvn` → `/usr/bin/mvn`, `Apache Maven 3.8.7` | ✅ PASS (AC1, AC3 — see F1) |
| T2 | `portunix install maven --variant=latest` (fresh container) | `tar.gz` extracted to `/opt/maven`, `Adding to User PATH: /opt/maven/bin`, exit 0. New login shell: `mvn` → `/opt/maven/apache-maven-3.9.9/bin/mvn`, `Apache Maven 3.9.9` | ✅ PASS (AC4 — see F2) |
| T3 | `portunix install nodejs` | Variant `apt`, no `Adding to User PATH`, exit 0. `~/.profile` md5 unchanged before/after. `node --version` → `v18.19.1` | ✅ PASS (AC1) |
| T4 | `portunix install rust --variant apt` | No `Adding to User PATH`, exit 0. `~/.profile` md5 unchanged. `cargo 1.75.0`, `rustc 1.75.0` | ✅ PASS (AC1) |

## Acceptance Criteria

- [x] AC1 — `PATH_APPEND` not applied for package-manager types unless the
  variant defines its own environment (U1, U2, T1, T3, T4)
- [x] AC2 — variant-level `environment` overrides the platform one; documented
  in the issue file (U2)
- [x] AC3 — `portunix install maven` on Ubuntu no longer writes the non-existent
  `/opt/maven/bin` (T1). The `java` dependency still writes a non-existent
  directory, a separate defect (F1)
- [x] AC4 — archive and script variants keep the current behaviour (U1, U3, T2)
- [x] AC5 — decision to keep `apt` as the Maven default documented in the issue file
- [x] AC6 — unit tests cover the PATH decision per install type (U1–U3)

## Findings Outside the Issue Scope

Both findings concern the `tar.gz` path, which #253 does not change; they
exist on `main` as well.

### F1: `java` PATH_APPEND points to the pre-detection directory

`java.json` (Linux) uses `PATH_APPEND: ${extract_to}/bin` and
`JAVA_HOME: ${extract_to}`. `${extract_to}` resolves to the directory before the
archive root is detected (`/opt/java/jdk-8u462b08`), while the JDK lies one
level deeper (`/opt/java/jdk-8u462b08/jdk8u462-b08`). Result:
`export PATH="/opt/java/jdk-8u462b08/bin:$PATH"` in `~/.profile`, a directory
that does not exist. `update-alternatives` registers the correct path, so
`java` still works. Suggested fix: use `${install_path}/bin`. Because `maven`
depends on `java`, this also affects `portunix install maven` (T1, T2).

Also noted: the dependency resolution installs Java **8** (first declared variant)
for Maven, and on the `apt` path APT pulls OpenJDK 21 anyway.

### F2: `maven` archive PATH_APPEND `/opt/maven/bin` does not exist

The Maven archive extracts to `/opt/maven/apache-maven-3.9.9`, so the platform
`PATH_APPEND: /opt/maven/bin` again points to a non-existent directory. `mvn`
works only through `/etc/profile.d/maven.sh` written by `postInstall` (with the
detected root). Suggested fix: `${install_path}/bin`.

Recommendation: track F1 and F2 in a new issue (manifest audit of
`${extract_to}` vs. `${install_path}` in `PATH_APPEND`).

## Coverage

- Unit: every install type of the `Install` switch has a PATH decision assertion;
  the three affected manifests are checked against the shipped registry
- E2E: `apt` on Ubuntu for all three affected packages; `dnf` and `pacman` are
  covered by unit tests only (same code path)

## CI Notes

The new tests run in the existing `go test ./...` of `ptx-installer` and need no
network or container.

## Cleanup

Containers `ptx-test-253` and `ptx-test-253b` removed after the run.

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-11-08
**Tester signature**: Claude Code (tester role) on behalf of zdendaku
