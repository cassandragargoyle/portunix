# cmd

Balíček `cmd` definuje Cobra CLI rozhraní binárky `portunix`. Je to top-level
vstupní bod, který propojuje uživatelské podpříkazy s implementacemi v
`src/app/` a s pomocnými binárkami v `src/helpers/` (přes dispatcher v
`src/dispatcher/`).

Cesta modulu: `portunix.ai/cmd` (viz `go.mod`).

## Role v projektu

```text
main.go
  └─ cmd.Execute()             (tento balíček — parsování a routing CLI)
       ├─ src/app/...          (in-process implementace)
       └─ src/dispatcher       (out-of-process pomocné binárky: ptx-*)
```

Tento balíček dělá pouze CLI plumbing: parsování flagů a argumentů, vykreslení
nápovědy a předání řízení správné implementaci. Doménová logika patří do
`src/app/` nebo do helperu v `src/helpers/`.

## Vstupní body

| Symbol | Soubor | Účel |
| ------ | ------ | ---- |
| `Execute()` | `root.go` | Voláno z `main.go`; spouští Cobra root command. |
| `SetVersion()` | `root.go` | Nastaví build-time verzi na root command. |
| `rootCmd` | `root.go` | Cobra root command (`portunix`). |
| `initPluginDispatcher()` | `plugin_dispatcher.go` | Při startu registruje nainstalované pluginy jako dynamické podpříkazy. |

## Víceúrovňová nápověda

`root.go` přepisuje výchozí Cobra nápovědu funkcí `multiLevelHelp`, která se
přepíná podle dvou perzistentních flagů:

- `--help-expert` — rozšířená nápověda se všemi volbami a příklady
  (`GenerateExpertHelp`).
- `--help-ai` — strojově čitelná JSON nápověda pro AI asistenty
  (`GenerateAIHelp`).
- výchozí — stručná nápověda pro člověka (`GenerateBasicHelp`).

Generátory nápovědy a metadata model příkazů/parametrů jsou v
`help_registry.go`.

## Skupiny příkazů

Podpříkazy jsou seskupené podle domény. Každý `*.go` soubor registruje jeden
příkaz (nebo malou rodinu) na `rootCmd` ze svého `init()`.

| Oblast | Soubory | Podkladový balíček |
| ------ | ------- | ------------------ |
| Root / nápověda | `root.go`, `help_registry.go` | — |
| Dispatch pluginů | `plugin.go`, `plugin_dispatcher.go` | `app/plugins` |
| Instalace / balíčky | `install.go`, `install_apt.go`, `install_chocolatey.go`, `install_iso.go`, `install-self.go`, `winget.go` | `app/install`, `app/selfinstall` |
| Kontejnery (generické) | `container.go`, `container_*_test.go` | `app/container` |
| Docker | `docker.go`, `docker_install_alias.go`, `docker_management.go`, `docker_run_in_container.go` | `app/docker` |
| Podman | `podman.go`, `podman_desktop.go`, `podman_management.go`, `podman_run_in_container.go` | `app/podman` |
| Sandbox | `sandbox.go`, `sandbox_default.go`, `sandbox_generate.go`, `sandbox_run_in_sandbox.go`, `sandbox_start.go` | `app/sandbox` |
| Virtualizace | `virt.go`, `virt_exec.go`, `virt_iso.go`, `virt_lifecycle.go`, `virt_snapshot.go`, `virt_ssh.go`, `virt_template.go`, `create_vm.go` | `app/virt` |
| Proxmox | `proxmox.go` | `app/virt` (podstrom Proxmox) |
| Edge / VPS | `edge.go` | `app/edge` |
| MCP / AI | `mcp.go` | `app/mcp` |
| Systém | `system.go`, `cache.go`, `config.go`, `update.go`, `version.go`, `service.go` | `app/system`, `app/update`, `app/version`, `app/service` |
| Autentizace | `login.go`, `logout.go` | `app/install`, `app/setup` |
| Ostatní | `create.go`, `datastore.go`, `guid.go`, `python.go`, `registry.go`, `unzip.go`, `wizard.go` | `app/...` (odpovídající podbalíček) |

Soubory s koncovkou `_test.go` jsou unit/integrační testy odpovídajícího
příkazu a spouští se přes `go test ./src/cmd/...`.

## Příkazy z pluginů

`initPluginDispatcher` (volaný z `init` v `root.go`) načítá
`~/.portunix/plugins/registry.json`, iteruje přes nainstalované pluginy a
každý přidává na `rootCmd` jako podpříkaz pomocí `createPluginCommand`. Volání
pluginu se forwarduje do procesu pluginu; registrace je při chybě tichá, takže
chybějící/prázdný registry znamená jen „žádné podpříkazy z pluginů“.

## Konvence

- Jeden Cobra příkaz na soubor; registrace na `rootCmd` z `init()`.
- CLI soubory drž tenké: parsuj flagy, validuj vstup a deleguj do `app/...`
  nebo do helperu. Doménová logika sem nepatří.
- Všechny uživatelské texty (Short/Long popisy, chybové hlášky) jsou v
  angličtině.
- Příkazy obsluhované helperem jdou přes `src/dispatcher`, ne přímo přes
  `os/exec` — vyhledávání binárek zůstává na jednom místě.
- Licenční hlavička (`MIT`) je povinná na každém novém souboru — použij stejné
  záhlaví jako `root.go`.

## Související

- `main.go` — volá `cmd.Execute()`.
- `src/dispatcher/dispatcher.go` — vyhledává a spouští helper binárky `ptx-*`.
- `src/app/` — in-process implementace stojící za CLI.
- `src/helpers/` — samostatné helper binárky (`ptx-container`, `ptx-mcp`,
  `ptx-virt`, …) volané přes dispatcher.
- `docs/FEATURES_OVERVIEW.md` — uživatelský přehled příkazů `portunix`.
