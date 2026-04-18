# Changelog

All notable changes to Portunix will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.2.3] - 2026-04-17

### Fixed in 2.2.3

- **`portunix install-self` installed only 4 of 12 ptx-* helpers** — the hardcoded helpers list in `src/app/selfinstall/install.go` was missing `ptx-prompting`, `ptx-python`, `ptx-installer`, `ptx-aiops`, `ptx-make`, `ptx-pft`, `ptx-credential`, and `ptx-trace`. Users running `install.sh` from a release archive ended up with a partial install. Replaced with a `ptx-*` glob so new helpers are picked up automatically.
- **`portunix install-self` reported a hardcoded `v1.5.7`** — `selfinstall.getVersion()` returned a literal string from a long-abandoned TODO. Now reads `update.Version`, which is set at build time via ldflags and already propagated from `main.version` at startup.
- **GitHub repo path was inconsistent between uppercase and lowercase** — the repository was renamed on GitHub from `Portunix` to `portunix`; `github.com/...` URLs redirect, but `github.io/Portunix/` did not, so GitHub Pages shipped with broken asset and link references. Normalized 143 occurrences across 47 files to lowercase. Added two sensitive-pattern rules so the preflight check fails if the uppercase form reappears.
- **`scripts/upload-release-to-github.py` failed from the Gitea clone** — `gh release create` could not auto-detect the target repo because the Gitea dev clone intentionally has no GitHub remote. Pin `--repo cassandragargoyle/portunix` explicitly.
- **README install snippets pointed at non-existent archive names** — release archives are named `portunix_<ver>_<os>_<arch>.tar.gz`, not `portunix_<os>_<arch>.tar.gz`. Updated the snippet to the versioned filename, extract into a dedicated `portunix-install/` directory (archives have no top-level wrapper folder), and use the bundled `install.sh` via `sudo bash install.sh` so all 13 binaries are installed.

### Added in 2.2.3

- **Root `CONTRIBUTING.md`** — GitHub's Community Standards page recognizes `CONTRIBUTING.md` only in root, `.github/`, or `docs/`. The existing `docs/contributing/` directory (subdirectory + README) was not surfaced. The new root-level entry point carries a quick-start workflow, ground rules, and links into the detailed methodology docs.

### Changed in 2.2.3

- **GoReleaser ldflags** — added `-X portunix.ai/app/update.Version={{ .Version }}` and corrected the stale `portunix.cz/app/version.ProductVersion` module path to `portunix.ai/app/version.ProductVersion`.
- **`docs/contributing/README-DUAL-SYSTEM.md` + `.claude/commands/cs/deploy-github.md`** — document the new CHANGELOG check step (KROK 3.5) and the `gh api` based GitHub state lookup that avoids registering a `github` remote on the Gitea repo.

## [2.2.2] - 2026-04-17

### Added in 2.2.2

- **Odoo installation support** (Issue #172) — installable via `portunix install odoo`; container sidecar orchestration with external PostgreSQL variant (`container-external-db`) and `--db-*` install flags
- **Container network/volume/inspect subcommands** (Issue #173) — first-class `portunix container network`, `portunix container volume`, and `portunix container inspect` commands unifying Docker/Podman inspection flows
- **uv Python tooling adoption** (Issue #171, ADR-039) — `pyproject.toml` + `uv.lock` replace `requirements.txt` / `test/requirements-test.txt`; `scripts/setup-venv.*` rewritten to call `uv sync`; standalone `scripts/*.py` gain PEP 723 inline metadata so they run via `uv run script.py` on any machine with uv installed; `portunix install uv` package added to ptx-installer
- **Node.js + markdownlint-cli2 in dev-setup** (Issue #170) — `scripts/dev-setup.sh` / `dev-setup.ps1` now bootstrap Node.js and `markdownlint-cli2`, enabling `make lint-md` on fresh clones
- **Czech README translations** — `README.cs.md` (Gitea) and `README.github.cs.md` (GitHub); dual-README rename in the GitHub sync script generalized to handle language suffixes (`README.github[.<lang>].md` → `README[.<lang>].md`)
- **scripts/README.md** — index of all scripts with a clean separation between one-time GitHub onboarding (`github-00-setup.sh`, `github-02-quick-publish.sh`) and recurring sync (`github-01-preflight-check.sh`, `github-02-sync-publish.py`)

### Changed in 2.2.2

- **Portunix tagline** — repositioned as "unified AI plugin and task platform for development environments" across README.md, README.github.md, CLAUDE.md, docs-site, and release scripts
- **GitHub sync workflow** — `scripts/github-02-sync-publish.py` no longer registers a `github` remote on the Gitea development repo; it clones GitHub by URL directly into the work directory, keeping the two repositories cleanly separated
- **Command handling documentation** — expanded and clarified across multiple helpers (ptx-aiops, ptx-ansible, ptx-container, ptx-credential, ptx-make, ptx-mcp, ptx-pft, ptx-prompting, ptx-python, ptx-trace, ptx-virt)

### Fixed in 2.2.2

- **hugo.json package definition** (Issue #170) — removed trailing commas that broke JSON parsing on strict parsers
- **Odoo acceptance defects** (Issue #172) — resolved install defects uncovered during container-based acceptance testing; addressed Findings #4/#5/#6 in `odoo.json` descriptions

### Removed in 2.2.2

- **`scripts/activate-venv.{sh,ps1,cmd}`** — with `uv run <cmd>` replacing activation for scripts, pytest, and REPLs, and modern IDEs auto-detecting `.venv/` via `pyproject.toml`/`uv.lock`, the wrappers no longer earn their keep (ADR-039)
- **Legacy `requirements.txt` and `test/requirements-test.txt`** — superseded by `pyproject.toml` + `uv.lock`

## [2.2.1] - 2026-04-08

### Fixed in 2.2.1

- **Windows build failure** (Issue #169) — service package used Linux-only syscalls (`Setpgid`, `Flock`) without build constraints; split into platform-specific files with proper Windows equivalents (`CREATE_NEW_PROCESS_GROUP`, `LockFileEx`)

## [2.2.0] - 2026-04-08

### Added in 2.2.0

- **GNU Make installation support** (Issue #168) — Windows via ezwinports ZIP, Linux via native package managers
- **Dev setup scripts** — `scripts/dev-setup.sh` and `scripts/dev-setup.ps1` for bootstrapping development environment
- **Proxmox VM/CT management** — automated provisioning commands for Proxmox virtualization

### Fixed in 2.2.0

- **setx PATH bug** — replaced broken `setx PATH "%PATH%;..."` with `environment.PATH_APPEND` in 9 packages (ninja, hugo, hugo-extended, actionlint, act, caddy, clang, protoc, tea)
- **Windows install.ps1** — rewritten as standalone installer (Issue #164)

## [2.1.0] - 2026-03-29

### Added in 2.1.0

- **Plugin Manifest schema alignment** — fix Interfaces field misplacement (Issue #162)
- **Markdown Style Guide** and `make lint-md` for consistent documentation
- **Helper --help-ai and --help-expert flags** across 6 helpers (Issue #163)
- **GitHub Actions workflow validation** in preflight checks

### Fixed in 2.1.0

- Documentation formatting and clarity improvements across multiple files

## [1.10.7] - 2026-03-08

### Added in 1.10.7

- **Docsy template** for documentation containers with auto-dependency install (Issue #153)
- Cross-platform Python plugin execution in dispatcher

### Fixed in 1.10.7

- Windows build constraint for syslog handler
- License headers added to multiple source files

## [1.10.4] - 2026-02-18

### Added in 1.10.4

- **Vox plugin ptx-installer integration** — model installation with multi-file download (Issue #149)
- Download type support in registry validation

## [1.10.1] - 2026-01-22

### Added in 1.10.1

- **GitHub deployment workflow** with version selection and publishing
- **GitHub Pages Hugo site** deployment via GitHub Actions
- PTX-Python local venv support and script generation
- New commands for Python development, AI operations, and credential management

## [1.9.1] - 2026-01-21

### Added in 1.9.1

- **PTX-Python project-local venv support** (Issue #138)
- Local deployment commands and plugin management in help output

### Changed in 1.9.1

- Replaced inline PowerShell with Python scripts for local deployment

## [1.9.0] - 2025-12-27

### Added in 1.9.0

- **PTX-PFT enhancements** — verbatim field, extended fields, case-insensitive categories (Issue #117)
- Automatic author role assignment in pft add
- Recursive subdirectory scanning for QFD structure

## [1.8.0] - 2025-12-25

### Added in 1.8.0

- **PTX-PFT Helper** — Product Feedback Tool with Fider/ClearFlask/Eververse providers
- PTX-Prompting helper added to GoReleaser release config

### Fixed in 1.8.0

- MCP subcommands missing in release build (Issue #113)

## [1.7.6] - 2025-12-02

### Added in 1.7.6

- **PTX-Make Helper** - New helper binary for cross-platform Makefile utilities (Issue #102)
  - File operations: `copy`, `mkdir`, `rm`, `exists`
  - Build metadata: `version`, `commit`, `timestamp`
  - Utilities: `checksum`, `chmod`, `json`, `env`
  - Dispatcher integration via `portunix make <command>`
  - `chmod` is no-op on Windows for portability

## [1.7.5] - 2025-12-01

### Added in 1.7.5

- **PTX-AIOps Helper** - AI Operations helper for GPU management and Ollama integration (Issue #101)
  - GPU status monitoring with NVIDIA support
  - Ollama container management
  - Model installation and management
  - Open WebUI deployment

## [1.7.4] - 2025-11-30

### Added in 1.7.4

- PTX-Virt helper binary for virtualization management
- PTX-Prompting helper for template-based prompt generation

## [1.7.3] - 2025-11-28

### Added in 1.7.3

- Clipboard support for interactive prompting

## [1.7.2] - 2025-11-25

### Fixed in 1.7.2

- Version embedding in build process

## [1.7.1] - 2025-11-24

### Fixed in 1.7.1

- Build script version updates

## [1.7.0] - 2025-11-20

### Added in 1.7.0

- PTX-Installer Helper for package management (Issue #100)
- Package Registry Architecture with AI integration (Issue #082)
- Hugo installation support (Issue #075)
- Container list command (Issue #084)

### Fixed in 1.7.0

- Hugo installation permission issues (Issue #085)
- Container exec command malfunction (Issue #095)
- Container rm subcommand recognition (Issue #094)

## [1.6.4] - 2025-11-15

### Added in 1.6.4

- Ansible Infrastructure as Code integration (Issue #056)
- VirtualBox/KVM conflict detection (Issue #088)
- QEMU/KVM adapter for virt check (Issue #089)
- Libvirt daemon detection and auto-fix (Issue #090, #091)

### Fixed in 1.6.4

- Virtual machine snapshot list empty names (Issue #061)
- VS Code installation filename resolution (Issue #064)

## [1.6.3] - 2025-11-10

### Added in 1.6.3

- GitHub CLI installation support (Issue #078)
- Ansible Galaxy Collections support (Issue #063)
- Universal virtualization support (Issue #049)

## [1.6.0] - 2025-11-01

### Added in 1.6.0

- Multi-level help system (Issue #050)
- Git-like dispatcher architecture (Issue #051)
- Container runtime capability detection (Issue #039)
- Node.js/npm installation support (Issue #041)

### Fixed in 1.6.0

- Container run command flag parsing (Issue #038)
- Module path naming inconsistencies (Issue #053)
