# Issue #181: Container `ls` Subcommand Not Recognized (alias for `list`)

## 🎯 Priority

**MEDIUM** — Missing industry-standard alias in the container dispatcher; same
pattern as the already-fixed #094 (`rm`).

## 📋 Status

- **Created**: 2026-04-23
- **Status**: ✅ Implemented
- **Closed**: 2026-05-02
- **Assignee**: -
- **Branch**: feature/issue-181-container-ls-alias (merged)
- **Related**: #094 (container `rm` subcommand not recognized — same root cause)
- **Acceptance Protocol**: [docs/testing/internal/acceptance-181.md](../../testing/internal/acceptance-181.md)

## 📝 Problem Description

### Current Situation

`portunix container ls` is rejected, even though `ls` is registered as an alias
of `list` on the cobra command in `src/cmd/container.go:604`:

```go
var containerListCmd = &cobra.Command{
    Use:     "list",
    Short:   "List containers from all available runtimes",
    Aliases: []string{"ls", "ps"},
    ...
}
```

Actual behavior on `portunix v2.2.5`:

```bash
$ portunix container ls
Unknown container subcommand: ls
Available subcommands: run, run-in-container, exec, list, stop, start, rm,
  logs, cp, info, check, compose, compose-preflight, network, volume, inspect
```

### Expected Behavior

`ls` and `ps` must behave identically to `list`, matching Docker and Podman:

```bash
$ portunix container ls   # lists all containers
$ portunix container ps   # same as `ls`
$ portunix container list # same as `ls`
```

## 🔬 How Docker and Podman Name This Command

Both runtimes expose **the same three names** for the "list containers" command,
with `ls` as the canonical spelling (not `list`):

**Docker** (`docker container ls --help`):

```text
Aliases:
  docker container ls, docker container list, docker container ps, docker ps
```

**Podman** (`podman container ls --help`):

```text
Aliases: ls, list, ps
Also exposed as the top-level `podman ps`.
```

Implications for Portunix:

- Portunix should accept all three subcommand names: `ls`, `list`, `ps`.
- Canonical spelling in help text and docs should be `ls` (matches Docker/Podman
  convention); `list` and `ps` are aliases.
- The top-level shortcut `portunix container ps` is out of scope here but worth
  considering as a follow-up (mirrors `docker ps` / `podman ps`).

## 🎯 Root Cause

The cobra definition in `src/cmd/container.go` is **not reached** for dispatched
subcommands. The parent `portunix` binary routes `container`/`docker`/`podman`
commands to the helper `ptx-container` (see `src/dispatcher/dispatcher.go`),
which re-dispatches via a plain `switch` statement on the subcommand string:

`src/helpers/ptx-container/main.go:126-162`

```go
switch subcommand {
case "run":    handleContainerRun(cmdArgs)
...
case "list":   handleContainerList(cmdArgs)
...
default:
    fmt.Printf("Unknown %s subcommand: %s\n", command, subcommand)
    fmt.Printf("Available subcommands: run, run-in-container, exec, list, ...\n")
}
```

The switch has no cases for `ls` or `ps`, so cobra's `Aliases` field is
effectively ignored for the dispatched path. This is the exact same class of
bug as #094 (which only registered `rm` in the switch, missing `remove`).

## ✅ Acceptance Criteria

1. `portunix container ls` lists containers (identical output to `list`).
2. `portunix container ps` lists containers (identical output to `list`).
3. `portunix container list` continues to work (no regression).
4. `portunix container --help` advertises the aliases, e.g.
   `list (aliases: ls, ps)` or a dedicated line for `ls`.
5. Integration test covers all three spellings against the same container set
   and asserts identical output.
6. The fix is applied to **all three** dispatched top-levels where the helper
   is invoked: `container`, `docker`, `podman` (the switch is shared).

## 🛠️ Suggested Fix

In `src/helpers/ptx-container/main.go`, group the aliases with the canonical:

```go
case "list", "ls", "ps":
    handleContainerList(cmdArgs)
```

Also update the "Available subcommands" fallback message and the help block in
`handleCommand` (around `src/helpers/ptx-container/main.go:92`) to show `ls` as
the primary name with `list, ps` as aliases — consistent with Docker/Podman.

### Broader Fix (optional, recommended as follow-up)

The dispatcher duplicates cobra's alias list by hand, which is what caused both
#094 and #181. Consider one of:

- Have the helper delegate straight to the cobra command tree (let cobra do the
  alias resolution), or
- Introduce a single alias map (`canonical -> []aliases`) used both by the
  switch and the help generator, so a new alias is registered in one place.

This should be captured as a separate follow-up issue rather than expanded in
this one.

## 🧪 Testing Strategy

- Unit test on the dispatcher in `src/helpers/ptx-container/` asserting that
  `ls`, `ps`, and `list` all route to `handleContainerList`.
- Integration test (mirroring `test/integration/issue_094_container_rm_test.go`)
  that runs `portunix container ls` and `portunix container ps` against a live
  runtime and verifies:
  - No `Unknown container subcommand` output.
  - Output matches `portunix container list`.
- Run the new tests in a container per
  `docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md` (container-based
  testing policy).

## 📎 References

- `src/helpers/ptx-container/main.go:126-162` — dispatcher switch (missing aliases)
- `src/cmd/container.go:600-622` — cobra command with `Aliases: {"ls", "ps"}` (not honored on dispatched path)
- `docs/issues/done/094-container-rm-subcommand-not-recognized.md` — same pattern, prior fix
- Docker: `docker container ls` aliases → `list`, `ps`, `docker ps`
- Podman: `podman container ls` aliases → `list`, `ps`, `podman ps`
