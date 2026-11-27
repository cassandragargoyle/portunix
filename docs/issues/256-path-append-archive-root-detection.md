# Issue #256: PATH_APPEND of archive packages points above the detected package root

**Status**: Open
**Priority**: Medium
**Type**: Bug
**Component**: ptx-installer
**Created**: 2026-11-09
**GitHub**: #256
**Related**: #253, #187

## Summary

Archive installs (`tar.gz`, `zip`) of some packages write a PATH entry to a
directory that does not exist. The manifests reference `${extract_to}` or a
fixed `extractTo` path in `environment.PATH_APPEND`, while `installArchive`
detects the real package root one level deeper (single top-level directory in
the archive). Found as findings F1 and F2 during acceptance testing of #253
(`docs/testing/acceptance-253.md`).

## Problem Description

`installArchive` in `src/helpers/ptx-installer/engine/installer.go` extracts to
`extractTo`, then `archive.FindExtractedRoot` switches to the single top-level
directory when it contains `bin/` or `lib/`. The returned `installResult` carries:

- `installPath` — the detected package root
- `extractTo` — the directory before root detection (kept for packages that
  reference the archive's own top-level name, e.g. maven on Windows)

Packages whose `PATH_APPEND` points at `${extract_to}/bin` (or a fixed
`extractTo` + `/bin`) therefore get a non-existent PATH entry.

Reproduction (`ubuntu:24.04` container):

```bash
portunix install maven --variant=latest
grep PATH ~/.profile
# export PATH="/opt/java/jdk-8u462b08/bin:$PATH"   <- does not exist
# export PATH="/opt/maven/bin:$PATH"               <- does not exist
ls /opt/java/jdk-8u462b08     # jdk8u462-b08/
ls /opt/maven                 # apache-maven-3.9.9/
```

`java` and `mvn` still work by accident: `java` through `update-alternatives`
(postInstall), `mvn` through `/etc/profile.d/maven.sh` (postInstall with the
detected root).

### Affected manifests

| Package | Platform | extractTo | PATH_APPEND | Actual root | Verified |
| ------- | -------- | --------- | ----------- | ----------- | -------- |
| java (8, 11, 17, 21) | linux | `/opt/java/jdk-<ver>` | `${extract_to}/bin` | `/opt/java/jdk-<ver>/jdk<ver>` | yes (8) |
| maven (3.9.9, latest) | linux | `/opt/maven` | `/opt/maven/bin` | `/opt/maven/apache-maven-3.9.9` | yes |
| graalvm (17, 21, 23) | linux, windows | `/opt/graalvm/graalvm-jdk-<ver>` | `${extract_to}/bin` | probably one level deeper | no |

Not affected: `maven` on Windows (`${extract_to}/apache-maven-3.9.9/bin`
references the top-level directory explicitly), `go` on Linux (`/usr/local`
has more than one entry, no root switch).

`JAVA_HOME: ${extract_to}` in `java.json` has the same mistake, but the engine
does not apply environment keys other than `PATH_APPEND` today.

### Related observation

The `maven` dependency on `java` installs Java 8 (first declared variant).

## Acceptance Criteria

1. `java`, `maven` and `graalvm` archive variants add the existing `bin`
   directory of the installed package to the user PATH (`${install_path}/bin`
   or equivalent)
2. All manifests are audited: no `PATH_APPEND` resolves to a non-existent
   directory after an archive install
3. Optional: the engine warns (non-fatal) when the resolved `PATH_APPEND`
   directory does not exist after installation
4. A registry/unit test guards the rule (e.g. archive variants with an
   `extractTo` must not use `${extract_to}/bin` unless the archive has no
   top-level directory)
5. Decision documented which Java version the `maven` dependency should
   install (variant default for dependencies)

## Testing

- Container test (`ubuntu:24.04`): `portunix install java`, `portunix install
  maven --variant=latest`, `portunix install graalvm` — every PATH entry added
  to `~/.profile` exists, `java -version` / `mvn --version` work in a new login
  shell without relying on `/etc/profile.d` or `update-alternatives`
- Windows Sandbox: `portunix install graalvm` — the user PATH entry exists
- Unit test for the manifest rule from AC4
