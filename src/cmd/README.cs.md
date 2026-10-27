# cmd

Balíček `cmd` definuje Cobra CLI rozhraní binárky `portunix`: strom příkazů,
parsování flagů a argumentů, vykreslení nápovědy a předání řízení
implementacím v `src/app/`.

Cesta modulu: `portunix.ai/cmd` (viz `go.mod`).

## Role v projektu

`main.go` se nejdřív zeptá dispatcheru. Příkaz, který obsluhuje pomocná
binárka, se do tohoto balíčku vůbec nedostane; vše ostatní běží přes Cobru:

```text
main.go
  ├─ src/dispatcher            helper vedle binárky? → spustí ptx-* a skončí
  └─ cmd.Execute()             (tento balíček — parsování a routing CLI)
       └─ src/app/...          (in-process implementace)
```

Tento balíček jen obsluhuje příkazovou řádku: zpracuje vstup a předá ho dál.
Doménová logika patří do `src/app/` nebo do helperu v `src/helpers/`.

### Příkazy obsluhované helpery

`src/dispatcher/dispatcher.go` předává tyto příkazy nejvyšší úrovně pomocným
binárkám, pokud helper leží vedle `portunix`:

| Příkazy | Helper |
| ------- | ------ |
| `container`, `docker`, `podman` | `ptx-container` |
| `install`, `package` | `ptx-installer` |
| `virt` | `ptx-virt` |
| `proxmox` | `ptx-proxmox` |
| `python` | `ptx-python` |
| `mcp` | `ptx-mcp` |
| `playbook`, `prompt`, `aiops`, `make`, `pft`, `credential`, `trace`, `ssh`, `scp`, `plugin-registry`, `specpm`, `database` / `db`, `github`, `wizard` | `ptx-ansible`, `ptx-prompting`, `ptx-aiops`, `ptx-make`, `ptx-pft`, `ptx-credential`, `ptx-trace`, `ptx-ssh`, `ptx-plugin-registry`, `ptx-specpm`, `ptx-database`, `ptx-github`, `ptx-wizard` |

Cobra příkazy, které tento balíček pro některé z nich stále definuje (např.
`container.go`, `docker*.go`, `podman*.go`, `virt*.go`, `proxmox.go`,
`install.go`, `mcp.go`, `python.go`), jsou **záložní cesta** pro případ, že
helper chybí. Jejich chování držte v souladu s helperem; změna, která se dostane
jen do helperu, nechá záložní cestu pozadu.

## Vstupní body

| Symbol | Soubor | Účel |
| ------ | ------ | ---- |
| `Execute()` | `root.go` | Voláno z `main.go`; spouští Cobra root command. |
| `SetVersion()` | `root.go` | Nastaví build-time verzi na root command. |
| `rootCmd` | `root.go` | Cobra root command (`portunix`). |
| `initPluginDispatcher()` | `plugin_dispatcher.go` | Při startu registruje povolené pluginy jako dynamické podpříkazy. |

## Víceúrovňová nápověda

`root.go` nahrazuje výchozí nápovědu Cobry funkcí `multiLevelHelp`, která se
řídí dvěma persistentními flagy:

- `--help-expert` — rozšířená nápověda se všemi volbami a příklady
  (`GenerateExpertHelp`)
- `--help-ai` — strojově čitelná JSON nápověda pro AI asistenty
  (`GenerateAIHelp`)
- výchozí — stručná nápověda pro člověka (`GenerateBasicHelp`)

Generátory nápovědy a metadata příkazů a parametrů jsou v `help_registry.go`,
popsaném v `help_registry.go.md`.

## Skupiny příkazů

Každý soubor `*.go` registruje jeden příkaz (nebo malou rodinu) ve svém
`init()`. Podpříkazy se připojují ke svému rodiči, příkazy nejvyšší úrovně k
`rootCmd`.

| Oblast | Soubory | Podkladový balíček |
| ------ | ------- | ------------------ |
| Root / nápověda | `root.go`, `help_registry.go` | — |
| Pluginy | `plugin.go`, `plugin_dispatcher.go` | `app/plugins` |
| Instalace / balíčky | `install.go`, `install-self.go`, `winget.go` | `app/install`, `app/selfinstall` |
| Kontejnery (obecné) | `container.go` | `app/container`, `app/docker`, `app/podman` |
| Docker | `docker.go`, `docker_install_alias.go`, `docker_management.go`, `docker_run_in_container.go` | `app/docker` |
| Podman | `podman.go`, `podman_desktop.go`, `podman_management.go`, `podman_run_in_container.go` | `app/podman` |
| Sandbox | `sandbox.go`, `sandbox_default.go`, `sandbox_generate.go`, `sandbox_run_in_sandbox.go`, `sandbox_start.go` | `app/sandbox` |
| Virtualizace | `virt.go`, `virt_exec.go`, `virt_iso.go`, `virt_lifecycle.go`, `virt_snapshot.go`, `virt_ssh.go`, `virt_template.go`, `create_vm.go` | `app/virt` |
| Proxmox | `proxmox.go` | `app/virt` (podstrom Proxmox) |
| Edge / VPS | `edge.go` | `app/edge` |
| MCP / AI | `mcp.go` | `app/mcp` |
| Systém | `system.go`, `cache.go`, `config.go`, `update.go`, `version.go`, `version_check.go`, `service.go` | `app/system`, `app/update`, `app/version`, `app/service` |
| Autentizace | `login.go`, `logout.go` | `app/install`, `app/setup` |
| Ostatní | `create.go`, `datastore.go`, `guid.go`, `python.go`, `registry.go`, `unzip.go` | `app/...` (odpovídající podbalíček) |

## Příkazy pluginů

`initPluginDispatcher` (volaný z `init` v `root.go`) načte
`~/.portunix/plugins/registry.json` a každý nainstalovaný plugin přidá do
`rootCmd` jako podpříkaz přes `createPluginCommand`. Volání pluginu se předá
binárce pluginu. Registrace chyby tiše ignoruje, takže chybějící nebo prázdná
registry jednoduše znamená „žádné podpříkazy pluginů“.

## Testování

Balíček je samostatný modul, testuje se proto podle import cesty, ne podle
adresáře:

```bash
make test-cli                  # go test -v portunix.ai/cmd/...
go test portunix.ai/cmd/...
```

Testy nesmí spouštět skutečné příkazy: test, který zavolá `Execute()` na
podpříkazu, spustí root command s argumenty `go test`, vypíše nápovědu rootu a
čte skutečné `~/.portunix`. Testujte místo toho jednotlivé části — validátor
`Args` příkazu, jeho flagy (`ParseFlags`) a pomocné parsovací funkce, které
používá `Run` — a z unit testu nikdy nespouštějte kontejner, VM ani instalátor.

## Konvence

- Jeden Cobra příkaz na soubor; registrace v `init()`.
- CLI soubory držte tenké: naparsovat flagy, validovat vstup, předat
  `app/...`. Doménová logika sem nepatří.
- Všechny texty pro uživatele (Short/Long popisy, chybové hlášky) jsou
  anglicky.
- Nová funkčnost pro příkaz obsluhovaný helperem patří nejdřív do helperu; pro
  nové helpery viz checklist v `docs/contributing/HELPER-BINARY-DEVELOPMENT.md`.
- Licenční hlavička (`MIT`) je povinná v každém novém souboru — použijte
  stejnou jako v `root.go`.

## Související

- `main.go` — zeptá se dispatcheru, pak volá `cmd.Execute()`.
- `src/dispatcher/dispatcher.go` — mapuje příkazy na helpery `ptx-*` a spouští je.
- `src/app/` — in-process implementace za CLI.
- `src/helpers/` — samostatné pomocné binárky (`ptx-container`,
  `ptx-installer`, `ptx-virt`, …).
- `docs/FEATURES_OVERVIEW.md` — uživatelský přehled příkazů `portunix`.
