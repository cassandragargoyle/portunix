# cmd

Package `cmd` defines the Cobra CLI surface of the `portunix` binary. It is the
top-level entrypoint that wires user-facing subcommands to the backing
implementations in `src/app/` and to the helper binaries in `src/helpers/`
(via the dispatcher in `src/dispatcher/`).

Module path: `portunix.ai/cmd` (see `go.mod`).

## Role in the project

```text
main.go
  └─ cmd.Execute()             (this package — CLI parsing & routing)
       ├─ src/app/...          (in-process implementations)
       └─ src/dispatcher       (out-of-process helpers: ptx-*)
```

This package only does CLI plumbing: flag/argument parsing, help rendering,
and forwarding to the right implementation. Business logic belongs in
`src/app/` or in a helper under `src/helpers/`.

## Entry points

| Symbol | File | Purpose |
| ------ | ---- | ------- |
| `Execute()` | `root.go` | Called from `main.go`; runs the Cobra root command. |
| `SetVersion()` | `root.go` | Injects the build-time version into the root command. |
| `rootCmd` | `root.go` | Cobra root command (`portunix`). |
| `initPluginDispatcher()` | `plugin_dispatcher.go` | Registers installed plugins as dynamic subcommands at startup. |

## Multi-level help

`root.go` overrides Cobra's default help with `multiLevelHelp`, which switches
on two persistent flags:

- `--help-expert` — extended help with all options and examples
  (`GenerateExpertHelp`).
- `--help-ai` — machine-readable JSON help for AI assistants
  (`GenerateAIHelp`).
- default — concise human help (`GenerateBasicHelp`).

The help generators and the command/parameter metadata model live in
`help_registry.go`.

## Command groups

Subcommands are grouped by domain. Each `*.go` file registers one command (or
a small family) on `rootCmd` from its `init()`.

| Area | Files | Backing package |
| ---- | ----- | --------------- |
| Root / help | `root.go`, `help_registry.go` | — |
| Plugin dispatch | `plugin.go`, `plugin_dispatcher.go` | `app/plugins` |
| Install / packages | `install.go`, `install_apt.go`, `install_chocolatey.go`, `install_iso.go`, `install-self.go`, `winget.go` | `app/install`, `app/selfinstall` |
| Containers (generic) | `container.go`, `container_*_test.go` | `app/container` |
| Docker | `docker.go`, `docker_install_alias.go`, `docker_management.go`, `docker_run_in_container.go` | `app/docker` |
| Podman | `podman.go`, `podman_desktop.go`, `podman_management.go`, `podman_run_in_container.go` | `app/podman` |
| Sandbox | `sandbox.go`, `sandbox_default.go`, `sandbox_generate.go`, `sandbox_run_in_sandbox.go`, `sandbox_start.go` | `app/sandbox` |
| Virtualization | `virt.go`, `virt_exec.go`, `virt_iso.go`, `virt_lifecycle.go`, `virt_snapshot.go`, `virt_ssh.go`, `virt_template.go`, `create_vm.go` | `app/virt` |
| Proxmox | `proxmox.go` | `app/virt` (Proxmox subtree) |
| Edge / VPS | `edge.go` | `app/edge` |
| MCP / AI | `mcp.go` | `app/mcp` |
| System | `system.go`, `cache.go`, `config.go`, `update.go`, `version.go`, `service.go` | `app/system`, `app/update`, `app/version`, `app/service` |
| Auth | `login.go`, `logout.go` | `app/install`, `app/setup` |
| Misc | `create.go`, `datastore.go`, `guid.go`, `python.go`, `registry.go`, `unzip.go`, `wizard.go` | `app/...` (matching subpackage) |

Files ending in `_test.go` are unit/integration tests for the corresponding
command and run with `go test ./src/cmd/...`.

## Plugin commands

`initPluginDispatcher` (called from `root.go`'s `init`) reads
`~/.portunix/plugins/registry.json`, iterates over installed plugins, and adds
each one to `rootCmd` as a subcommand via `createPluginCommand`. Plugin
invocations are forwarded to the plugin process; the registration is silent on
errors so a missing/empty registry simply means "no plugin subcommands."

## Conventions

- One Cobra command per file; register it on `rootCmd` from `init()`.
- Keep CLI files thin: parse flags, validate input, then delegate to
  `app/...` or a helper. Do not put domain logic here.
- All user-facing strings (Short/Long descriptions, error messages) are in
  English.
- Helper-backed commands go through `src/dispatcher`, not directly through
  `os/exec`, so binary discovery stays in one place.
- License header (`MIT`) is required on every new file — match the header used
  in `root.go`.

## Related

- `main.go` — calls `cmd.Execute()`.
- `src/dispatcher/dispatcher.go` — resolves and invokes `ptx-*` helper binaries.
- `src/app/` — in-process implementations behind the CLI.
- `src/helpers/` — standalone helper binaries (`ptx-container`, `ptx-mcp`,
  `ptx-virt`, …) invoked via the dispatcher.
- `docs/FEATURES_OVERVIEW.md` — user-facing list of `portunix` commands.
