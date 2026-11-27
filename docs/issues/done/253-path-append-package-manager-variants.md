# Issue #253: PATH_APPEND applied to package-manager variants writes non-existent paths

**Status**: ✅ Implemented
**Closed**: 2026-11-09
**Priority**: Medium
**Type**: Bug
**Component**: ptx-installer
**Created**: 2026-11-08
**GitHub**: #253

## Summary

The platform-level `environment.PATH_APPEND` is applied after every successful
install, regardless of the effective install type. Package-manager variants
(APT, DNF, pacman, ...) therefore get PATH entries that only make sense for the
archive/script layout.

## Problem Description

Reproduction (`ubuntu:24.04`):

```bash
portunix install maven
```

- `selectDefaultVariant` picks the `apt` variant because APT is detected
- APT installs Maven into `/usr/share/maven` (`/usr/bin/mvn` symlink)
- The installer then appends `export PATH="/opt/maven/bin:$PATH"` to
  `~/.profile` — a directory that does not exist

Root cause: `Install` in `src/helpers/ptx-installer/engine/installer.go` reads
`platformSpec.Environment["PATH_APPEND"]` unconditionally. Package manager
installs already place binaries on the system PATH.

Affected packages (Linux platform, `PATH_APPEND` + package-manager variants):

| Package | PATH_APPEND | Package-manager variants |
| ------- | ----------- | ------------------------ |
| maven   | `/opt/maven/bin` | apt |
| nodejs  | `/usr/local/bin` | apt, dnf, pacman |
| rust    | `$HOME/.cargo/bin` | apt, dnf, pacman |

Related observation: the default Maven on Ubuntu is the distribution package
(3.8.7, pulls a JRE as a dependency), while the archive variants provide 3.9.x.

## Acceptance Criteria

1. `PATH_APPEND` is not applied when the effective install type is a package
   manager type (`deb`, `apt`, `dnf`, `yum`, `snap`, `pacman`), unless the
   variant explicitly defines its own environment
2. Optionally: variant-level `environment` overrides the platform-level one
   (schema + registry support, documented)
3. `portunix install maven` on Ubuntu leaves no non-existent directory in the
   user PATH files
4. Archive variants (`tar.gz`, `zip`, `download`, `script`) keep the current
   behaviour
5. Decision documented whether `maven` on Linux should default to the archive
   variant instead of `apt`
6. Unit tests cover the PATH decision per install type

## Implementation Notes

### PATH_APPEND resolution

`effectivePathAppend` in `engine/installer.go` decides the PATH entry per
install:

1. A variant-level `environment.PATH_APPEND` wins (an empty string disables
   PATH wiring for that variant)
2. Otherwise the platform value is skipped for package-manager types
   (`deb`, `apt`, `dnf`, `yum`, `snap`, `pacman`)
3. Otherwise the platform value is used (unchanged behaviour)

The same resolution is used for the post-script verification PATH (issue #203).

Windows package managers (`chocolatey`, `winget`) are intentionally not in the
skip list — some winget installers (e.g. `qemu`) do not put their directory on
PATH, so the platform `PATH_APPEND` is still needed there.

### Variant-level environment

`VariantSpec` accepts an optional `environment` map with the same keys as the
platform-level one; a key present on the variant overrides the platform value:

```json
"apt": {
  "type": "apt",
  "packages": ["example"],
  "environment": { "PATH_APPEND": "/usr/lib/example/bin" }
}
```

### Decision: Maven default variant on Linux

The `apt` variant stays the default on APT systems. `selectDefaultVariant`
gives the detected system package manager priority over `preferred` for all
packages; changing that for one package would need a new global rule. The
archive variants only provide `x64` URLs, so an archive default would break
arm64 hosts. Users who need Maven 3.9.x use `portunix install maven
--variant=latest`.

## Testing

- Unit test: PATH_APPEND applied for `tar.gz`, skipped for `apt` / `dnf` / `pacman`
- Container test (`ubuntu:24.04`): `portunix install maven`, check `~/.profile`
  and `mvn --version` in a new login shell
- Container test: `portunix install maven --variant=latest` still adds
  `/opt/maven/bin`
