# cmd

Package `cmd` defines the Cobra CLI surface of the `portunix` binary: command
tree, flag and argument parsing, help rendering, and forwarding to the backing
implementations in `src/app/`.

Module path: `portunix.ai/cmd` (see `go.mod`).

## Role in the project

`main.go` asks the dispatcher first. A command that a helper binary handles
never reaches this package; everything else runs through Cobra:

```text
main.go
  ├─ src/dispatcher            helper next to the binary? → run ptx-* and exit
  └─ cmd.Execute()             (this package — CLI parsing & routing)
       └─ src/app/...          (in-process implementations)
```

This package only does CLI plumbing. Business logic belongs in `src/app/` or in
a helper under `src/helpers/`.

### Commands owned by helpers

`src/dispatcher/dispatcher.go` routes these top-level commands to helper
binaries when the helper exists next to `portunix`:

| Commands | Helper |
| -------- | ------ |
| `container`, `docker`, `podman` | `ptx-container` |
| `install`, `package`, `skills` | `ptx-installer` |
| `virt` | `ptx-virt` |
| `proxmox` | `ptx-proxmox` |
| `python` | `ptx-python` |
| `mcp` | `ptx-mcp` |
| `playbook`, `prompt`, `aiops`, `make`, `pft`, `credential`, `trace`, `ssh`, `scp`, `plugin-registry`, `specpm`, `database` / `db`, `github`, `wizard` | `ptx-ansible`, `ptx-prompting`, `ptx-aiops`, `ptx-make`, `ptx-pft`, `ptx-credential`, `ptx-trace`, `ptx-ssh`, `ptx-plugin-registry`, `ptx-specpm`, `ptx-database`, `ptx-github`, `ptx-wizard` |

The Cobra commands that this package still defines for some of them (for example
`container.go`, `docker*.go`, `podman*.go`, `virt*.go`, `proxmox.go`,
`install.go`, `mcp.go`, `python.go`) are the **fallback** used when the helper
is missing. Keep their behaviour in line with the helper; a change that only
lands in the helper leaves the fallback behind.

`skills.go` is different: it has no fallback implementation. Its commands
exist only so that `portunix completion` can complete the `skills` subcommands,
flags, agent IDs and installed skill names, and they print an error when
`ptx-installer` is missing. Keep its flags and `skillsAgentIDs` in line with
`src/helpers/ptx-installer/cmd_skills.go` and `engine/skills/agents.go`.

## Entry points

| Symbol | File | Purpose |
| ------ | ---- | ------- |
| `Execute()` | `root.go` | Called from `main.go`; runs the Cobra root command. |
| `SetVersion()` | `root.go` | Injects the build-time version into the root command. |
| `rootCmd` | `root.go` | Cobra root command (`portunix`). |
| `initPluginDispatcher()` | `plugin_dispatcher.go` | Registers enabled plugins as dynamic subcommands at startup. |

## Multi-level help

`root.go` replaces Cobra's default help with `multiLevelHelp`, which switches
on two persistent flags:

- `--help-expert` — extended help with all options and examples
  (`GenerateExpertHelp`)
- `--help-ai` — machine-readable JSON help for AI assistants
  (`GenerateAIHelp`)
- default — concise human help (`GenerateBasicHelp`)

The help generators and the command/parameter metadata live in
`help_registry.go`, documented in `help_registry.go.md`.

## Command groups

Each `*.go` file registers one command (or a small family) from its `init()`.
Subcommands attach to their parent command, top-level commands to `rootCmd`.

| Area | Files | Backing package |
| ---- | ----- | --------------- |
| Root / help | `root.go`, `help_registry.go` | — |
| Plugins | `plugin.go`, `plugin_dispatcher.go` | `app/plugins` |
| Install / packages | `install.go`, `install-self.go`, `winget.go` | `app/install`, `app/selfinstall` |
| Containers (generic) | `container.go` | `app/container`, `app/docker`, `app/podman` |
| Docker | `docker.go`, `docker_install_alias.go`, `docker_management.go`, `docker_run_in_container.go` | `app/docker` |
| Podman | `podman.go`, `podman_desktop.go`, `podman_management.go`, `podman_run_in_container.go` | `app/podman` |
| Sandbox | `sandbox.go`, `sandbox_default.go`, `sandbox_generate.go`, `sandbox_run_in_sandbox.go`, `sandbox_start.go` | `app/sandbox` |
| Virtualization | `virt.go`, `virt_exec.go`, `virt_iso.go`, `virt_lifecycle.go`, `virt_snapshot.go`, `virt_ssh.go`, `virt_template.go`, `create_vm.go` | `app/virt` |
| Proxmox | `proxmox.go` | `app/virt` (Proxmox subtree) |
| Edge / VPS | `edge.go` | `app/edge` |
| MCP / AI | `mcp.go` | `app/mcp` |
| System | `system.go`, `cache.go`, `config.go`, `update.go`, `version.go`, `version_check.go`, `service.go` | `app/system`, `app/update`, `app/version`, `app/service` |
| Auth | `login.go`, `logout.go` | `app/install`, `app/setup` |
| Misc | `create.go`, `datastore.go`, `guid.go`, `python.go`, `registry.go`, `unzip.go` | `app/...` (matching subpackage) |

## Plugin commands

`initPluginDispatcher` (called from `root.go`'s `init`) reads
`~/.portunix/plugins/registry.json` and adds every installed plugin to `rootCmd`
as a subcommand via `createPluginCommand`. Plugin invocations are forwarded to
the plugin binary. Registration is silent on errors, so a missing or empty
registry simply means "no plugin subcommands".

## Testing

The package is a separate module, so test it by import path, not by directory:

```bash
make test-cli                  # go test -v portunix.ai/cmd/...
go test portunix.ai/cmd/...
```

Tests must not execute real commands: a test that calls `Execute()` on a
subcommand runs the root command with the arguments of `go test`, prints the
root help and reads the real `~/.portunix`. Test the pieces instead — the
command's `Args` validator, its flags (`ParseFlags`) and the parsing helpers
the `Run` function uses — and never start a container, VM or installer from a
unit test.

## Conventions

- One Cobra command per file; register it from `init()`.
- Keep CLI files thin: parse flags, validate input, then delegate to `app/...`.
  Do not put domain logic here.
- All user-facing strings (Short/Long descriptions, error messages) are in
  English.
- New functionality for a helper-owned command goes into the helper first; see
  the checklist in `docs/contributing/HELPER-BINARY-DEVELOPMENT.md` for new
  helpers.
- License header (`MIT`) is required on every new file — match the header used
  in `root.go`.

## Related

- `main.go` — consults the dispatcher, then calls `cmd.Execute()`.
- `src/dispatcher/dispatcher.go` — maps commands to `ptx-*` helpers and runs them.
- `src/app/` — in-process implementations behind the CLI.
- `src/helpers/` — standalone helper binaries (`ptx-container`, `ptx-installer`,
  `ptx-virt`, …).
- `docs/FEATURES_OVERVIEW.md` — user-facing list of `portunix` commands.
