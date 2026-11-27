# Issue #255: Container bootstrap: deploy portunix with helpers and detect missing CA certificates

**Status**: Open
**Priority**: Medium
**Type**: Enhancement
**Component**: portunix, ptx-container, ptx-installer
**Created**: 2026-11-08
**GitHub**: #255

## Summary

Running portunix inside a fresh container currently requires manual bootstrap
steps. Copying only the `portunix` binary is not enough (helpers are missing)
and minimal images lack CA certificates, which surfaces as a raw `x509` error.

## Problem Description

### Missing helpers

`portunix install …` inside a container fails with
`ptx-installer helper not found` when only `portunix` was copied in. The
message is correct, but there is no supported way to get portunix *with its
helpers* into a container. The existing `run-in-container` command
(`src/helpers/ptx-container/main.go`) itself copies only `./portunix`
(relative path, no helpers) to `/usr/local/bin/portunix`, so the
`portunix install` it runs inside the container hits the same error.

### Missing CA certificates

A clean `ubuntu:24.04` image has no `ca-certificates` package. Every HTTPS
download then fails with
`x509: certificate signed by unknown authority`. The installer only prints the
error; it does not detect the cause or suggest a fix.

## Proposed Solution

- New `ptx-container` subcommand (e.g. `portunix container deploy-portunix
  <container> [--path /usr/local/bin]`) copying `portunix` and all `ptx-*`
  helpers resolved next to the running executable (not the CWD), setting the
  executable bit
- Fix `run-in-container` to use the same deploy logic
- Extend the `helper not found` message with a hint pointing to the new command
- In `ptx-installer`, detect TLS trust errors (`x509.UnknownAuthorityError`)
  on download and print an actionable hint per package manager
  (`apt-get install -y ca-certificates`, `dnf install -y ca-certificates`,
  `apk add ca-certificates`)
- Optionally: the deploy command checks for CA certificates in the container
  and offers to install them

## Acceptance Criteria

1. A single portunix command deploys `portunix` and all helpers into a
   running container; `portunix install --dry-run java` then works inside it
2. `run-in-container` works without a `portunix` binary in the CWD
3. The `helper not found` message references the deploy command
4. A download failing with an unknown-authority TLS error prints a hint naming
   the missing `ca-certificates` package for the detected package manager
5. Works with Podman and Docker (through `portunix container`)

## Testing

- Container test (`ubuntu:24.04`, Podman): deploy, then
  `portunix install --dry-run java` inside the container
- Container test (`ubuntu:24.04` without `ca-certificates`): install a package,
  verify the hint, install `ca-certificates`, retry successfully
- Unit test: TLS error classification and hint selection per package manager
