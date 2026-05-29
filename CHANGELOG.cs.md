# Changelog

Všechny významné změny v Portunixu jsou dokumentovány v tomto souboru.

Formát vychází z [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
a projekt dodržuje [Sémantické verzování](https://semver.org/spec/v2.0.0.html).

## [2.4.0] - 2026-05-20

### Přidáno ve 2.4.0

- **Helper `ptx-python`** (Issue #097) — implementace fáze 3 a fáze 4 samostatného Python helperu.
- **Helper `ptx-wizard`** (Issue #014) — interaktivní YAML-based instalační průvodce přemigrován do samostatného helper binárky s neinteraktivním
  enginem, dispatcher routingem a podmíněnou navigací.
- **MCP wizard — pokročilé funkce** (Issue #034) — skutečný handshake, konfigurace scope/env/timeout, slučování desktop konfigurace a podpora TLS.
- **Strategie správy verzí** (Issue #140, ADR-036) — interní buildy `vX.Y.Z+dev.N` vs. čisté `vX.Y.Z` GitHub releasy; varování v
  `build-with-version.sh` pro čisté stabilní buildy mimo `github` kontext.
- **Univerzální správa kontejnerů** (Issue #032) — `portunix container restart` a cross-runtime objevování mezi Dockerem a Podmanem.
- **Balíčky pro SPICE server, klienta a guest agenta** (Issue #093) — definice balíčků pro display protokol QEMU/KVM na linuxových distribucích.
- **AI prompty pro vox balíčky** (Issue #081) — doplněno pokrytí AI promptů pro zbývající vox balíčky ze seznamu discovery.
- **Aliasy `docker install` / `podman install`** (Issue #019) — směrované přes `ptx-container` pro paritu s nativními package manažery.
- **Příklad `EtlPipeline` v `ptx-trace` SDK** — example třída demonstrující použití SDK pro ETL pipeliny.

### Opraveno ve 2.4.0

- **Override typu variantu v installeru** — aplikováno na úrovni celého registry; explicitní override typu přidán pro dnf/pacman varianty
  (Issue #093), čímž se zabrání chybné detekci package manageru.
- **Regenerace `portunix.syso` v build pipeline** — `portunix.syso` je nyní regenerován při každém buildu s explicitním `asInvoker` manifestem; opravuje
  zastaralé verzovací informace ve výsledné binárce.
- **Error handling MCP wizardu** (Issue #034) — správná propagace chyb pro neznámé AI asistenty a vadné desktop konfigurační soubory.
- **Relativní cesty ve `cs/` skillu** (Issue #140) — opraveny relativní cesty k docs ze souborů `cs/` skillu.

### Změněno ve 2.4.0

- **Úklid legacy kódu v `ptx-installeru`** (Issue #186) — velký refactoring napříč čtyřmi sub-tasky: (186a) extrakce package registry do sdíleného
  package, (186b) `ptx-mcp` nyní instaluje `claude-code` přes subprocess `ptx-installeru`, (186c) extrakce `installconfig` do sdíleného package,
  (186d) migrace admin subcommandů `install apt`/`iso`/`chocolatey` do `ptx-installeru`. Legacy adresář `src/app/install/` odstraněn.
- **Hlavičky licencí a dokumentace** — doplněny MIT licenční hlavičky do několika Go souborů napříč `src/app/` a `src/helpers/`; nová README
  dokumentace pro `cmd` package; zlepšeno formátování v `CONTRIBUTING.md`.
- **Údržba repozitáře** — `.gitignore` přerovnán dle kategorií s globem `ptx-*` helperů; odstraněn zastaralý `TODO.md`; odstraněn legacy prompt `refactor.md`.

## [2.3.0] - 2026-05-11

### Přidáno ve 2.3.0

- **Helper `ptx-database`** (Issue #013, fáze 1) — nový helper binární poskytující správu PostgreSQL a SQLite: install, backup, MCP integrace.
- **Helper `ptx-github`** (Issue #025) — operace pro repozitáře, releasy a autentizaci: clone (čistě Go přes go-git), checkout, tagy, releasy,
  download, info, status; AES-256-GCM šifrovaný token store na `~/.portunix/github/auth.json`; resume + SHA-256 verifikace u stahování assetů.
- **Helper `ptx-specpm`** (Issues #183, #184) — fáze 1 scaffoldingu pro specifikace project-managementu přes spec-kit-pm; fáze 2 přidává příkaz
  `upgrade`, drift line a air-gapped workflow (ADR-040, ADR-041).
- **Správa životního cyklu kontejnerů** (Issue #027) — `portunix container` získává run/stop/start/restart se zárukami cleanup-u.
- **Ověřování Ed25519 podpisu updatů** (Issues #165, #166) — self-signed podepisování releasů pro update channel a ověřování stažených updatů (ADR-042).
- **Vylepšení instalace balíčků** (Issue #079, fáze 1) — alias `--method` pro `--variant`, `--list-variants` / `--list-methods` pro vypsání variant, zlepšená chybová hláška pro neznámé varianty.
- **Sledování URL v package metadatech** (Issue #080, ADR-019) — každý  manifest balíčku nyní nese `installationDocsUrl` a `latestVersionUrl`
  v metadatech; lehký jq guard v lint CI jobu zabraňuje novým manifestům bez těchto polí.
- **Instalace Terraformu** (Issue #065).
- **Instalace Double Commanderu** (Issue #066).
- **Dokumentační engine VitePress a MkDocs** (Issue #154).
- **Bash integrace `ptx-trace`** (Issue #141).
- **`pprof` profiling pro system info** (Issue #118).
- **UX instalace Dockeru na Windows** (Issues #178, #179) — auto-elevation přes UAC, interaktivní výběr prerekvizit když chybí WSL2/Hyper-V, samostatný `portunix install wsl` přes Windows features.

### Opraveno ve 2.3.0

- **Výkon system info** (Issue #099) — zkráceno provedení `portunix system info` z ~1.2 s na <60 ms.
- **Kontrola admin elevace pro Docker Desktop na Windows** (Issue #177) — přidána pre-install kontrola elevace.
- **Chybějící prompt na data-root adresář v `install docker`** (Issue #176) — obnoveno.
- **`ls` / `ps` u kontejneru nebyly rozpoznány** (Issue #181) — nyní přijímány jako aliasy pro `list`.
- **Python wheel pluginy selhávají na Windows** (Issue #182) — ošetřena přípona `.exe` a vynechána kontrola exec bitu v cestě venv.
- **Varování `portunix enable` na nepodporovaném OS** (Issue #155) — testovací binárka je nyní rovněž OS-aware.

### Změněno ve 2.3.0

- **Sdílený kód download/extract pro sandbox / `ptx-installer`** (Issue #127) — refaktorováno do společné cesty.
- **Přepis MCP wizardu** (Issue #034) — dokončení pokročilých funkcí.

## [2.2.5] - 2026-04-20

### Přidáno ve 2.2.5

- **Helper `ptx-proxmox`** (Issue #167) — kompletní správa Proxmox VE přes REST API. Fáze 1: auth profily v `~/.config/portunix/proxmox.json`
  (mode 0600) s API-token a heslovou (ticket-based) autentizací.
  Fáze 2: životní cyklus VM/CT — `list`, `info`, `status`, `start`, `stop`, `shutdown`, `restart`, `delete`, plus `template list`, `create vm`,
  `create ct` a konfigurace `cloud-init`. Fáze 3: `snapshot create|list|revert|delete`, `resolve-ip` a interaktivní `ssh` / `exec` /
  `copy` s IP hosta zjišťovaným přes QEMU guest agent nebo LXC interfaces. Fáze 4: `.ptxbook` soubory mohou cílit na `environment.type: proxmox`
  (integrováno v `ptx-ansible`) s `create_if_missing` pro auto-provisioning. 36 unit testů + 54-scenario acceptance protokol (`docs/testing/internal/acceptance-167*.md`).
- **Helper `ptx-plugin-registry`** (Issue #175) — lokální gRPC démon, který obsluhuje PluginRegistryService, aby hostitelské platformy (Synapse, Pack,
  Agent) mohly objevit, které nainstalované Portunix pluginy je cílují. Unix-socket transport na Linux/macOS, TCP loopback na Windows;
  on-demand životní cyklus (žádná systemd integrace v1). Vystaveno přes `portunix plugin-registry serve …`.

### Změněno ve 2.2.5

- **Hlavičky licencí** — přidán banner `CassandraGargoyle Community Project` / MIT do několika Go zdrojových souborů napříč `src/app/` a `src/helpers/`,
  kterým chyběl. Žádné funkční změny.

## [2.2.4] - 2026-04-18

### Přidáno ve 2.2.4

- **Helper `ptx-ssh`** (Issue #174) — sjednocený SSH klient registrovaný jako `portunix ssh` / `portunix scp`. Nativní Go klient plus opt-in pty wrapper
  mód, neinteraktivní heslová autentizace přes `ptx-credential` (`--credential`, `--credential-fd`, `--credential-file`, `--credential-env`)
  a interaktivní prompty `--ask-user` / `--ask-pass`. Subcommandy: `connect`, `exec`, `copy`, `rsync`, `bootstrap-key`, `trust`, `known-hosts`, `agent`,
  `sshpass-compat`. Hesla jsou maskována typem `SecretString`; `--password` na CLI je odmítáno, `--host-key-check=off` odmítá pod `PORTUNIX_CI=1`. Viz `docs/helpers/ptx-ssh.md`.
- **Workflow pro bump verze** (`.claude/commands/cs/bump-version.md`) — znovupoužitelný skill pro SemVer bump + CHANGELOG + tag ve správném pořadí.

### Opraveno ve 2.2.4

- **`ssh copy` ignoroval ne-default port** ve formátu `user@host:port:/path` — parser zaměnil port za součást cesty a vždy
  vytáčel 22. Opraveno.
- **`ssh exec` nepropagoval remote exit kódy** — dispatcher maskoval `*exec.ExitError` jako exit 1. Nyní `portunix ssh exec … "exit 42"` vrací 42.
- **`known-hosts remove host:port` přehlížel záznamy `[host]:port`** — matcher byl literální; nyní normalizuje bracketové formy na obou stranách.
- **CI test suite odblokována + migrace zastaralých GitHub Actions** (commit `25bd7bf`).

### Změněno ve 2.2.4

- **`LnxExecutePyScriptSsh` přepsán jako shim nad `ptx-ssh`** (Issue #174, AC-5) — odstraňuje hardcoded cestu ke klíči a `ssh.InsecureIgnoreHostKey()`.

## [2.2.3] - 2026-04-17

### Opraveno ve 2.2.3

- **`portunix install-self` instaloval jen 4 z 12 ptx-* helperů** — hardcoded seznam helperů v `src/app/selfinstall/install.go` postrádal
  `ptx-prompting`, `ptx-python`, `ptx-installer`, `ptx-aiops`, `ptx-make`, `ptx-pft`, `ptx-credential` a `ptx-trace`. Uživatelé spouštějící
  `install.sh` z release archivu skončili s neúplnou instalací. Nahrazeno globem `ptx-*`, takže nové helpery jsou pochytány automaticky.
- **`portunix install-self` hlásil hardcoded `v1.5.7`** — `selfinstall.getVersion()` vracel literál z dávno opuštěného TODO.
  Nyní čte `update.Version`, který je nastaven při buildu přes ldflags a propagován z `main.version` při startu.
- **Cesta GitHub repa byla nekonzistentní mezi velkými a malými písmeny** — repozitář byl na GitHubu přejmenován z `Portunix` na `portunix`;
  `github.com/...` URL redirectuje, ale `github.io/Portunix/` ne, takže GitHub Pages byly nasazeny s rozbitými odkazy na assety a reference.
  Normalizováno 143 výskytů napříč 47 soubory na malá písmena. Přidána dvě sensitive-pattern pravidla, aby preflight check selhal, pokud se znovu
  objeví velká písmena.
- **`scripts/upload-release-to-github.py` selhával z Gitea klonu** — `gh release create` nedokázal autodetekovat target repo, protože Gitea
  dev klon záměrně nemá GitHub remote. Pevně nastaveno `--repo cassandragargoyle/portunix`.
- **README install snippety ukazovaly na neexistující názvy archivů** — release archivy se jmenují `portunix_<ver>_<os>_<arch>.tar.gz`, ne
  `portunix_<os>_<arch>.tar.gz`. Snippet aktualizován na verzovaný název, rozbalení do dedikovaného adresáře `portunix-install/` (archivy nemají
  top-level wrapper adresář) a použit přibalený `install.sh` přes `sudo bash install.sh`, aby se nainstalovalo všech 13 binárek.

### Přidáno ve 2.2.3

- **Kořenový `CONTRIBUTING.md`** — GitHub Community Standards rozpoznává `CONTRIBUTING.md` jen v kořeni, `.github/` nebo `docs/`. Existující
  adresář `docs/contributing/` (podadresář + README) nebyl viditelný. Nový kořenový vstupní bod obsahuje quick-start workflow, základní pravidla
  a odkazy do detailní metodologické dokumentace.

### Změněno ve 2.2.3

- **GoReleaser ldflags** — přidáno `-X portunix.ai/app/update.Version={{ .Version }}` a opravena zastaralá
  cesta modulu `portunix.cz/app/version.ProductVersion` na `portunix.ai/app/version.ProductVersion`.
- **`docs/contributing/README-DUAL-SYSTEM.md` + `.claude/commands/cs/deploy-github.md`** — dokumentován nový krok kontroly
  CHANGELOG (KROK 3.5) a `gh api` based lookup stavu GitHubu, který se vyhýbá registraci `github` remote v Gitea repu.

## [2.2.2] - 2026-04-17

### Přidáno ve 2.2.2

- **Podpora instalace Odoo** (Issue #172) — instalovatelné přes `portunix install odoo`; sidecar orchestrace kontejnerů s externí
  PostgreSQL variantou (`container-external-db`) a install flagy `--db-*`
- **Subcommandy pro síť/volume/inspect kontejnerů** (Issue #173) — prvotřídní `portunix container network`, `portunix container volume`
  a `portunix container inspect` sjednocující Docker/Podman inspekční toky
- **Adopce uv Python toolingu** (Issue #171, ADR-039) — `pyproject.toml` + `uv.lock` nahrazují `requirements.txt` / `test/requirements-test.txt`;
  `scripts/setup-venv.*` přepsány na volání `uv sync`; samostatné `scripts/*.py` získávají PEP 723 inline metadata, takže běží přes
  `uv run script.py` na jakémkoli stroji s nainstalovaným uv; balíček `portunix install uv` přidán do ptx-installeru
- **Node.js + markdownlint-cli2 v dev-setupu** (Issue #170) — `scripts/dev-setup.sh` / `dev-setup.ps1` nyní bootstrapují Node.js
  a `markdownlint-cli2`, čímž umožňují `make lint-md` na čerstvých klonech
- **České překlady README** — `README.cs.md` (Gitea) a `README.github.cs.md` (GitHub); dual-README rename v GitHub sync skriptu zobecněn, aby
  zvládal jazykové přípony (`README.github[.<lang>].md` → `README[.<lang>].md`)
- **scripts/README.md** — index všech skriptů s čistým oddělením mezi jednorázovým GitHub onboardingem (`github-00-setup.sh`,
  `github-02-quick-publish.sh`) a opakovanou synchronizací (`github-01-preflight-check.sh`, `github-02-sync-publish.py`)

### Změněno ve 2.2.2

- **Tagline Portunixu** — repozicován jako "unified AI plugin and task platform for development environments" napříč README.md, README.github.md,
  CLAUDE.md, docs-site a release skripty
- **Workflow synchronizace s GitHubem** — `scripts/github-02-sync-publish.py` již neregistruje `github` remote
  v Gitea dev repu; klonuje GitHub přímo přes URL do pracovního adresáře, čímž drží oba repozitáře čistě oddělené
- **Dokumentace command handling** — rozšířena a zpřesněna napříč několika helpery (ptx-aiops, ptx-ansible, ptx-container, ptx-credential, ptx-make,
  ptx-mcp, ptx-pft, ptx-prompting, ptx-python, ptx-trace, ptx-virt)

### Opraveno ve 2.2.2

- **Definice balíčku hugo.json** (Issue #170) — odstraněny koncové čárky, které rozbíjely parsování JSON v striktních parserech
- **Acceptance defekty Odoo** (Issue #172) — vyřešeny install defekty objevené během container-based acceptance testování; adresovány Findings
  #4/#5/#6 v popisech `odoo.json`

### Odstraněno ve 2.2.2

- **`scripts/activate-venv.{sh,ps1,cmd}`** — s `uv run <cmd>` nahrazujícím aktivaci pro skripty, pytest a REPLy, a moderními IDE automaticky
  detekujícími `.venv/` přes `pyproject.toml`/`uv.lock`, již wrappery neměly opodstatnění (ADR-039)
- **Legacy `requirements.txt` a `test/requirements-test.txt`** — nahrazeno `pyproject.toml` + `uv.lock`

## [2.2.1] - 2026-04-08

### Opraveno ve 2.2.1

- **Selhání buildu na Windows** (Issue #169) — service package používal Linux-only syscally (`Setpgid`, `Flock`) bez build constraintů; rozděleno
  do platform-specific souborů s odpovídajícími Windows ekvivalenty (`CREATE_NEW_PROCESS_GROUP`, `LockFileEx`)

## [2.2.0] - 2026-04-08

### Přidáno ve 2.2.0

- **Podpora instalace GNU Make** (Issue #168) — Windows přes ezwinports ZIP, Linux přes nativní package manažery
- **Dev setup skripty** — `scripts/dev-setup.sh` a `scripts/dev-setup.ps1` pro bootstrap vývojového prostředí
- **Správa Proxmox VM/CT** — automatizované provisioning příkazy pro Proxmox virtualizaci

### Opraveno ve 2.2.0

- **Bug v setx PATH** — nahrazen rozbitý `setx PATH "%PATH%;..."` za `environment.PATH_APPEND` v 9 balíčcích (ninja, hugo, hugo-extended,
  actionlint, act, caddy, clang, protoc, tea)
- **Windows install.ps1** — přepsáno jako samostatný installer (Issue #164)

## [2.1.0] - 2026-03-29

### Přidáno ve 2.1.0

- **Sladění schématu Plugin Manifestu** — oprava chybného umístění pole Interfaces (Issue #162)
- **Markdown Style Guide** a `make lint-md` pro konzistentní dokumentaci
- **Flagy `--help-ai` a `--help-expert`** napříč 6 helpery (Issue #163)
- **Validace GitHub Actions workflow** v preflight kontrolách

### Opraveno ve 2.1.0

- Zlepšení formátování a srozumitelnosti dokumentace napříč několika soubory

## [1.10.7] - 2026-03-08

### Přidáno v 1.10.7

- **Docsy template** pro dokumentační kontejnery s auto-instalací závislostí (Issue #153)
- Cross-platform spouštění Python pluginů v dispatcheru

### Opraveno v 1.10.7

- Windows build constraint pro syslog handler
- Přidány hlavičky licencí do několika zdrojových souborů

## [1.10.4] - 2026-02-18

### Přidáno v 1.10.4

- **Integrace vox pluginu s ptx-installerem** — instalace modelů s multi-file downloadem (Issue #149)
- Podpora typu download ve validaci registry

## [1.10.1] - 2026-01-22

### Přidáno v 1.10.1

- **Workflow nasazení na GitHub** s výběrem verze a publikací
- **Nasazení GitHub Pages Hugo sajty** přes GitHub Actions
- Podpora lokálního venv v PTX-Python a generování skriptů
- Nové příkazy pro Python development, AI operace a správu credentials

## [1.9.1] - 2026-01-21

### Přidáno v 1.9.1

- **Podpora project-local venv v PTX-Python** (Issue #138)
- Lokální deployment příkazy a správa pluginů ve výpisu help

### Změněno v 1.9.1

- Inline PowerShell nahrazen Python skripty pro lokální nasazení

## [1.9.0] - 2025-12-27

### Přidáno v 1.9.0

- **Vylepšení PTX-PFT** — verbatim pole, rozšířená pole, case-insensitive kategorie (Issue #117)
- Automatické přiřazení role autora v `pft add`
- Rekurzivní skenování podadresářů pro QFD strukturu

## [1.8.0] - 2025-12-25

### Přidáno v 1.8.0

- **Helper PTX-PFT** — Product Feedback Tool s providery
  Fider/ClearFlask/Eververse
- Helper PTX-Prompting přidán do GoReleaser release konfigurace

### Opraveno v 1.8.0

- Chybějící MCP subcommandy v release buildu (Issue #113)

## [1.7.6] - 2025-12-02

### Přidáno v 1.7.6

- **Helper PTX-Make** — nový helper binární pro cross-platform utility k Makefilům (Issue #102)
  - Souborové operace: `copy`, `mkdir`, `rm`, `exists`
  - Build metadata: `version`, `commit`, `timestamp`
  - Utility: `checksum`, `chmod`, `json`, `env`
  - Integrace s dispatcherem přes `portunix make <command>`
  - `chmod` je no-op na Windows pro přenositelnost

## [1.7.5] - 2025-12-01

### Přidáno v 1.7.5

- **Helper PTX-AIOps** — AI Operations helper pro správu GPU a integraci s Ollama (Issue #101)
  - Monitoring stavu GPU s NVIDIA podporou
  - Správa Ollama kontejnerů
  - Instalace a správa modelů
  - Nasazení Open WebUI

## [1.7.4] - 2025-11-30

### Přidáno v 1.7.4

- Helper binárka PTX-Virt pro správu virtualizace
- Helper PTX-Prompting pro template-based generování promptů

## [1.7.3] - 2025-11-28

### Přidáno v 1.7.3

- Podpora schránky pro interaktivní prompting

## [1.7.2] - 2025-11-25

### Opraveno v 1.7.2

- Embedding verze v build procesu

## [1.7.1] - 2025-11-24

### Opraveno v 1.7.1

- Aktualizace verzí v build skriptu

## [1.7.0] - 2025-11-20

### Přidáno v 1.7.0

- Helper PTX-Installer pro správu balíčků (Issue #100)
- Architektura Package Registry s AI integrací (Issue #082)
- Podpora instalace Hugo (Issue #075)
- Příkaz `container list` (Issue #084)

### Opraveno v 1.7.0

- Problémy s oprávněními při instalaci Hugo (Issue #085)
- Nefunkčnost příkazu container exec (Issue #095)
- Rozpoznání subcommandu container rm (Issue #094)

## [1.6.4] - 2025-11-15

### Přidáno v 1.6.4

- Integrace Ansible Infrastructure as Code (Issue #056)
- Detekce konfliktu VirtualBox/KVM (Issue #088)
- QEMU/KVM adaptér pro virt check (Issue #089)
- Detekce libvirt démona a auto-fix (Issue #090, #091)

### Opraveno v 1.6.4

- Prázdná jména snapshotů virtuálních strojů (Issue #061)
- Rozlišování souborů při instalaci VS Code (Issue #064)

## [1.6.3] - 2025-11-10

### Přidáno v 1.6.3

- Podpora instalace GitHub CLI (Issue #078)
- Podpora Ansible Galaxy Collections (Issue #063)
- Univerzální podpora virtualizace (Issue #049)

## [1.6.0] - 2025-11-01

### Přidáno v 1.6.0

- Víceúrovňový help systém (Issue #050)
- Architektura git-like dispatcheru (Issue #051)
- Detekce capabilit container runtime (Issue #039)
- Podpora instalace Node.js/npm (Issue #041)

### Opraveno v 1.6.0

- Parsování flagů v příkazu container run (Issue #038)
- Nekonzistence v pojmenování cest modulů (Issue #053)
