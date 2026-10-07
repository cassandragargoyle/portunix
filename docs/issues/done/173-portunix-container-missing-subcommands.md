# Issue #173: Portunix Container: Missing `network`, `volume`, `inspect` Subcommands

- **Type**: Enhancement
- **Priority**: Medium
- **Status**: ✅ Implemented
- **Labels**: enhancement, container, podman, docker, test-methodology, ptx-container
- **Created**: 2026-04-17

## Description

`portunix container` provides a universal wrapper over Podman and Docker for
the operations testers and installers actually perform today (`run`, `exec`,
`list`, `stop`, `start`, `rm`, `logs`, `cp`, `compose`, `info`, `check`). It is,
however, missing three classes of subcommands that are unavoidable when
packaging or validating container-based software:

1. **Network management** — `container network create / list / inspect / rm`
2. **Volume management** — `container volume create / list / inspect / rm / prune`
3. **Container introspection** — `container inspect <name> [--format <tmpl>]`

Without these, any test that needs to set up a shared network, verify
environment variables written into a container, or clean up a test's
persistent data is forced to fall back on raw `podman …` / `docker …`
invocations — which directly violates the Container-Based Testing Policy in
`docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md`.

## Motivation

- The policy is mandatory:
  > All software installation testing MUST use Portunix native container
  > commands instead of direct Docker/Podman calls.

- Issue #172 (Odoo installation) was the first issue that exercised sidecar
  orchestration through ptx-installer. Its re-test cycle had to fall back to
  `podman` repeatedly for setup/verify — documented in Appendix B of
  `docs/testing/acceptance-172.md`. Every future container-service package
  (elasticsearch, minio, future Redis/Kafka/etc.) will hit the same wall.

- The gap also shows up inside the installer engine itself:
  `src/helpers/ptx-installer/engine/container_service.go` had to use
  `exec.LookPath("podman") / exec.LookPath("docker")` plus direct
  `<runtime> network create` invocations because `portunix container network`
  didn't exist. A stable wrapper would let engine code use the single
  abstraction everywhere.

## Scope

### In scope — add three subcommand trees to `portunix container`

```text
container network
    create <name> [--driver bridge] [--subnet CIDR] [--gateway IP]
    list
    inspect <name> [-f '<template>']
    rm <name>...
container volume
    create <name> [--driver local]
    list
    inspect <name> [-f '<template>']
    rm <name>...
    prune [--force]
container inspect <name> [-f '<template>']
```

Each command must:

- Auto-select the underlying runtime (podman first, docker fallback), matching
  the existing `container` subcommand conventions.
- Surface non-zero exit codes from the runtime (unlike the pre-#172
  `runPodmanContainer` / `runDockerContainer` which silently swallowed errors
  until fixed in commit `3c5b9f8`).
- Preserve Docker / Podman flag compatibility where practical — users type the
  same `--format '{{…}}'` template they already know.

### In scope — refactor engine usage

- Replace the direct `exec.Command("podman"|"docker", "network", …)` call in
  `src/helpers/ptx-installer/engine/container_service.go` (function
  `ensureNetwork`) with `portunix container network create` + `inspect`.
- Replace the ad-hoc string-contains `checkContainerExists` with a call to
  the new `portunix container inspect <name>` returning a structured signal.

### Out of scope

- Podman Pods and Docker Compose orchestration beyond what the existing
  `container compose` subcommand already covers.
- Swarm / Kubernetes wrappers.
- A fully generic `container exec --format …` flag — only the minimum needed
  to read env vars and network attachments from `inspect`.

## Acceptance Criteria

- [ ] AC-1: `portunix container network create <name>` creates a bridge network
      named `<name>` (idempotent when the network already exists, with a clear
      informational message — not an error).
- [ ] AC-2: `portunix container network list` shows at least all user-created
      networks from both Docker and Podman runtimes.
- [ ] AC-3: `portunix container network rm <name>` removes the network if no
      container is attached; surfaces the runtime's error otherwise.
- [ ] AC-4: `portunix container volume create <name>` creates a named volume;
      `... rm <name>` removes it. Missing-volume and not-empty errors from the
      runtime propagate with non-zero exit code.
- [ ] AC-5: `portunix container inspect <name>` returns the runtime's JSON
      blob. `inspect <name> -f '{{.NetworkSettings.Networks}}'` returns the
      single templated value, matching podman/docker semantics.
- [ ] AC-6: All subcommands respect the `--help` convention (Long description
      listing supported flags). Help text is updated in the
      `ptx-container` helper's `showHelp` entrypoints, not only in the root
      cobra tree (see Issue #172 Finding #2 for the dispatcher routing pitfall).
- [ ] AC-7: Engine refactor: `ensureNetwork` and `checkContainerExists` in
      `container_service.go` no longer shell out to `podman` / `docker`
      directly; they use the new `portunix container` subcommands.
- [ ] AC-8: Acceptance of Issue #172 can be fully re-run without any direct
      `podman` / `docker` invocation in setup or verify steps.

## Technical Notes

- Dispatcher routing: `container`, `docker`, and `podman` commands go to the
  `ptx-container` helper (`src/dispatcher/dispatcher.go:62`). New subcommands
  need to be implemented there (`src/helpers/ptx-container/main.go`), not in
  `src/cmd/container.go` — changes to the main cobra tree are dead code for
  the normal dispatched path, as demonstrated by Issue #172 Finding #3.
- Rootless Podman can emit a harmless warning during `network rm` (`rootless
  netns: kill network process: permission denied`) — the wrapper should not
  treat this as a failure when the network is in fact removed. A post-check
  via `network inspect` can confirm removal state.
- `docker` does not have a native equivalent of `podman network inspect
  --format` for all fields; the wrapper should document any runtime-specific
  surface differences.

## References

- `docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md` — Container-Based
  Testing Policy (mandatory)
- `docs/testing/acceptance-172.md` — Appendix B (methodology deviation note
  that motivated this issue)
- `src/helpers/ptx-container/main.go` — target file for new subcommands
- `src/helpers/ptx-installer/engine/container_service.go` — refactor target

## Complexity

**Medium** — three subcommand trees, two runtime backends, careful error
propagation (especially given the pre-#172 silent-failure bug), one engine
refactor. No new architectural decisions; the pattern mirrors the existing
`container run` / `container list` implementations.
