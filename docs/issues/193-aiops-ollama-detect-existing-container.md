# Issue #193: `portunix aiops ollama container status` — Detect Pre-existing `ollama` Container

## 🎯 Priority

**MEDIUM** — UX fix for `portunix aiops ollama container` commands. The status
command only ever inspects a container named `portunix-ollama`. When the user
already has a container simply named `ollama` (the default name used by
`ollama/ollama` and most tutorials/tools), Portunix reports `Not found` and
offers to create a duplicate, which is confusing and can lead to a second,
redundant Ollama container occupying port 11434.

## 📋 Status

- **Created**: 2026-07-06
- **Status**: 📋 Open
- **Assignee**: -
- **Branch**: feature/issue-193-aiops-ollama-detect-existing-container
- **Related**:
  - `docs/commands/core/aiops/` — public documentation of the aiops commands

## 📝 Problem Description

### Current Situation

`getOllamaContainerStatus()` inspects a single, hard-coded container name
`portunix-ollama`:

`src/helpers/ptx-aiops/ollama.go:20, 64-107`

```go
const OllamaContainerName = "portunix-ollama"

func getOllamaContainerStatus() OllamaContainerStatus {
    status := OllamaContainerStatus{Name: OllamaContainerName}
    runtime := detectContainerRuntime()
    if runtime == "" {
        return status
    }
    // Check if container exists
    cmd := exec.Command(runtime, "inspect", OllamaContainerName, "--format", "{{.State.Running}}")
    output, err := cmd.Output()
    if err != nil {
        // Container doesn't exist
        return status
    }
    ...
}
```

If no `portunix-ollama` container exists, `handleOllamaContainerStatus()`
prints:

`src/helpers/ptx-aiops/ollama.go:244-259`

```text
Ollama Container Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Container 'portunix-ollama': Not found

Create container with:
  portunix aiops ollama container create
```

The problem: a container named just `ollama` (very common — it is the name in
the official Ollama Docker instructions and what most people already run) is
**invisible** to Portunix. The user is told nothing exists and is nudged to
create `portunix-ollama`, ending up with two Ollama containers both trying to
bind port 11434 (the second `create` then fails on the port clash, or the user
has two competing runtimes).

The same hard-coded name is used by every container subcommand:

- `handleOllamaContainerCreate` — `ollama.go:109`
- `handleOllamaContainerStatus` — `ollama.go:244`
- `handleOllamaContainerStart` — `ollama.go:310`
- `handleOllamaContainerStop` — `ollama.go:351`
- `handleOllamaContainerRemove` — `ollama.go:377`

### Expected Behavior

1. `portunix aiops ollama container status` also looks for a container named
   `ollama` (in addition to the managed `portunix-ollama`).
2. When a `portunix-ollama` container exists, it is used as before (no change).
3. When `portunix-ollama` does **not** exist but a container named `ollama`
   does, `status` reports **that** container (name, running state, image, port,
   API availability) instead of printing a blank `Not found`. It should make
   clear the container is not the Portunix-managed one, e.g.:

   ```text
   Ollama Container Status
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   Container:   ollama  (external, not managed by Portunix)
   Status:      🟢 Running
   Image:       ollama/ollama:latest
   API Port:    11434

   API Status:  ✓ Available at http://localhost:11434
   ```

4. When neither container exists, behaviour is unchanged (`Not found` + create
   hint).
5. **`start` / `stop` / `remove` fall back to the external container.** When no
   managed `portunix-ollama` exists but an external `ollama` does, these
   lifecycle commands operate on the external container (same precedence as
   `status`). Each prints a `Note: operating on external container 'ollama'
   (not managed by Portunix)` line. `remove` additionally requires an explicit
   `[y/N]` confirmation for an external container (Portunix did not create it),
   and skips the managed data-directory note in that case. `create` is
   unaffected — it always targets the managed `portunix-ollama`.
6. **The `status` footer offers only valid follow-up commands.** For an external
   container the footer marks it as unmanaged and suggests the working lifecycle
   commands; model commands (which still default to `portunix-ollama`) are shown
   with the `--container` flag:

   ```text
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   This container is not managed by Portunix (managed name: portunix-ollama).
   Start container with:
     portunix aiops ollama container start
   ```

   and, when the external container is running:

   ```text
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   This container is not managed by Portunix (managed name: portunix-ollama).
   Commands:
     portunix aiops model list --container ollama - List installed models
     portunix aiops ollama container stop     - Stop container
   ```

## 🎯 Root Cause

- The container name is a single hard-coded constant (`OllamaContainerName =
  "portunix-ollama"`); there is no fallback lookup for the conventional
  `ollama` name.
- `getOllamaContainerStatus()` inspects exactly one name and returns
  `Exists=false` on the first miss.

## ✅ Acceptance Criteria

1. With a container named `ollama` present (and no `portunix-ollama`),
   `portunix aiops ollama container status` reports the `ollama` container's
   real state (name, running/stopped, image, port) and clearly marks it as
   external / not Portunix-managed.
2. With a `portunix-ollama` container present, output is unchanged from today
   (the managed container always takes precedence).
3. With neither container present, output is unchanged (`Not found` + create
   hint).
4. If both `portunix-ollama` and `ollama` exist, `portunix-ollama` wins and is
   the one reported (documented precedence, no ambiguity).
5. `start` / `stop` / `remove` operate on the managed `portunix-ollama` when it
   exists, otherwise fall back to an external `ollama` (same precedence as
   `status`), each printing a "operating on external container" note. `remove`
   of an external container requires an explicit `[y/N]` confirmation and skips
   the managed data-directory note. `create` is unchanged — it always targets
   the managed `portunix-ollama`.
6. For an external `ollama` container, the `status` footer offers only valid
   follow-up commands: `portunix aiops ollama container start` (stopped) or
   `portunix aiops ollama container stop` + `portunix aiops model list
   --container ollama` (running). No footer command dead-ends on a missing
   `portunix-ollama`.
7. Behaviour verified with both Docker and Podman runtimes.
8. Integration test in `test/integration/` (using `testframework`) covers the
   three states: external `ollama` only, managed `portunix-ollama` only, and
   neither.

## 🛠️ Suggested Fix

Introduce a small ordered lookup list and resolve the effective container name
inside `getOllamaContainerStatus()`:

```go
// Names checked in precedence order: managed first, then the conventional
// external name used by ollama/ollama tutorials.
var ollamaContainerCandidates = []string{OllamaContainerName, "ollama"}

func containerExists(runtime, name string) bool {
    err := exec.Command(runtime, "inspect", name, "--format", "{{.State.Running}}").Run()
    return err == nil
}

func getOllamaContainerStatus() OllamaContainerStatus {
    runtime := detectContainerRuntime()
    if runtime == "" {
        return OllamaContainerStatus{Name: OllamaContainerName}
    }
    for _, name := range ollamaContainerCandidates {
        if containerExists(runtime, name) {
            return inspectOllamaContainer(runtime, name) // existing detail logic, parametrised by name
        }
    }
    return OllamaContainerStatus{Name: OllamaContainerName}
}
```

- Add a `Managed bool` (or `External bool`) field to `OllamaContainerStatus`
  so `handleOllamaContainerStatus()` can annotate the external case with
  `(external, not managed by Portunix)`.
- Refactor the existing inspect block (`ollama.go:82-106`) into
  `inspectOllamaContainer(runtime, name)` taking the resolved name instead of
  the constant.

- Lifecycle handlers (`start` / `stop` / `remove`) resolve the target via
  `getOllamaContainerStatus()` and act on `status.Name` instead of the constant,
  so they follow the same managed→external precedence. `create` keeps using a
  managed-only lookup (`getManagedOllamaContainerStatus()`).

### Ownership handling (external containers)

`start` / `stop` / `remove` may act on the external `ollama` container as a
fallback, but with guardrails for the ownership questions this raises:

- Each operation prints an explicit "operating on external container" note.
- `remove` requires `[y/N]` confirmation before deleting a container Portunix
  did not create, and does **not** print the managed data-directory note (the
  `getOllamaDataDir()` assumption only holds for `portunix-ollama`).
- `create` never adopts an external container — it always targets the managed
  `portunix-ollama`.

## 🧪 Testing Strategy

- Integration test (`test/integration/issue_193_aiops_ollama_status_test.go`)
  using `testframework`, run in a container per the container-based testing
  policy in `docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md`:
  - Create a container named `ollama` (via `portunix container run`) and assert
    `aiops ollama container status` reports it as external and Running.
  - Create `portunix-ollama` and assert it is reported (managed) and takes
    precedence when both exist.
  - With neither present, assert the `Not found` + create hint output.
- Verify on both Docker and Podman runtimes where available.

## 📎 References

- `src/helpers/ptx-aiops/ollama.go:20` — `OllamaContainerName` constant
- `src/helpers/ptx-aiops/ollama.go:63-107` — `getOllamaContainerStatus`
- `src/helpers/ptx-aiops/ollama.go:244-294` — `handleOllamaContainerStatus`
- `src/helpers/ptx-aiops/ollama.go:310-377` — start/stop/remove handlers (same
  hard-coded name)
- `test/e2e/issue_101_aiops_ollama_e2e_test.go` — existing aiops ollama E2E test
- `docs/commands/core/aiops/` — public command documentation
