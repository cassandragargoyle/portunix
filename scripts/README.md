# Portunix Scripts

Utility scripts for building, releasing, publishing to GitHub, and local
development setup. This document explains what each script does so the
right tool is used for the right job.

## GitHub Publishing Scripts

There are **two distinct workflows** for publishing to GitHub. Do not
mix them up.

### Workflow A — Initial project onboarding (one-time)

Used only when a repository is **first** being placed on GitHub, or the
local clone does not yet know about the GitHub remote. These scripts
intentionally register a `github` remote on the local Gitea repository
and push directly from it.

| Script | Purpose |
| ------ | ------- |
| `github-00-setup.sh` | Adds `github` remote pointing to the public GitHub repo. Run once per fresh clone that will be used for direct publishing. |
| `github-02-quick-publish.sh` | Creates a staging clone, strips private files, squashes history into a single commit, and force-pushes to `github/main`. Suitable for the very first push or when you intentionally want a clean squashed history. |

**When to use:** first-time onboarding, full reset of the public repo,
or emergency hotfix where code review is deliberately skipped.

**Side effect:** the local Gitea repo will have a `github` remote
afterwards. This is expected for this workflow.

### Workflow B — Recurring synchronization (every normal deploy)

Used for all **routine** syncs from Gitea to GitHub. Never touches the
local repo's remotes; works entirely in a separate clone directory
(`../portunix-github-sync/`) and produces a feature branch for code
review on GitHub.

| Script | Purpose |
| ------ | ------- |
| `github-01-preflight-check.sh` | Pre-flight: private files, sensitive patterns, binaries, size checks. Run before every sync to catch leaks. |
| `github-02-sync-publish.py` | Main publish script. Clones GitHub by URL into `../portunix-github-sync/`, syncs files from the Gitea working tree, applies dual-README rename, creates a feature branch, and pushes it. PR is then opened on GitHub. |

**When to use:** the `/cs:deploy-github` workflow, scheduled releases,
any normal change publication.

**Important:** this path must never register the `github` remote on
the Gitea repo. If a future edit introduces such logic, remove it.

### Supporting files

| File | Purpose |
| ---- | ------- |
| `github-private-files.json` | Shared list of paths excluded from GitHub publication plus sensitive-pattern rules. Loaded by both the preflight check and `sync-publish.py`. |
| `portunix-cleanup-public.ps1` | Historical Windows cleanup script that informed the current private-file list. Kept for reference. |

### Dual README rename

Both workflows produce a repo where `README.github[.<lang>].md` is
renamed to `README[.<lang>].md` so the public mirror shows the
GitHub-specific README. `sync-publish.py` handles all language
variants automatically (e.g. `README.github.cs.md` → `README.cs.md`).

## Release Scripts

| Script | Purpose |
| ------ | ------- |
| `make-release.py` | Main release entry point. Takes a version (e.g. `v1.10.7`), builds all platforms via GoReleaser, generates release notes and checksums. Do not replace with custom build scripts. |
| `make-release-only-win.ps1` | Windows-only release helper. |
| `create-platform-archives.py` | Creates per-platform binary archives (ADR-031). Called by `make-release.py`. |
| `generate-checksums.sh` | Produces SHA256 checksums for dist/ artifacts. |
| `upload-release-to-gitea.py` | Uploads `dist/` artifacts to a Gitea release via the Gitea API. |
| `upload-release-to-github.py` | Uploads `dist/` artifacts to a GitHub release. |
| `post-release-docs.py` | Post-release documentation steps (changelog sync, docs-site updates). |

`build-with-version.sh` in the repo root embeds the version into
binaries and `portunix.rc`; it is part of the release pipeline and must
not be deleted.

## Local Deployment

| Script | Purpose |
| ------ | ------- |
| `install.sh` / `install.ps1` / `install.bat` | User-facing installers used by documentation and CI. |
| `deploy-local.py` | Copies freshly built binaries over an existing local install (auto-detects install path). Does not rebuild. |
| `undeploy-local.py` | Removes a local install. |
| `test-powershell-installation.sh` | Smoke test for the PowerShell installer. |

## Python / venv Tooling

Python-based scripts are managed via [uv](https://docs.astral.sh/uv/)
(ADR-039). The manifest is `pyproject.toml` + `uv.lock` at the repo
root — do not reintroduce `requirements.txt`.

| Script | Purpose |
| ------ | ------- |
| `setup-venv.sh` / `.cmd` / `.ps1` | Provision `.venv` via `uv sync`. Pass `--with-tests` to include test deps. |
| `dev-setup.sh` / `dev-setup.ps1` | Broader developer environment bootstrap. |
| `file-server.py` | Standalone file server (PEP 723 script, run via `uv run`). |

There is no `activate-venv` wrapper. Use `uv run <cmd>` (for scripts, pytest,
REPL) — it resolves `.venv/` automatically with no activation needed. Modern
IDEs detect `.venv/` via `pyproject.toml` / `uv.lock`.

## Documentation / Other

| Script | Purpose |
| ------ | ------- |
| `docs-serve.sh` / `.cmd` / `.ps1` | Launch the docs-site locally. |
| `publish-docs-to-github.py` | Publish built docs-site to the GitHub Pages branch. |
| `aur-prepare.sh` / `aur-publish.sh` / `aur-test-install.sh` | Arch User Repository packaging. |
| `setup.sh` | One-shot project bootstrap. |

---

## Quick Reference — "Which script do I want?"

- **First time pushing this project to GitHub?** → `github-00-setup.sh` then `github-02-quick-publish.sh`.
- **Normal deploy to GitHub?** → `github-01-preflight-check.sh`, then `uv run scripts/github-02-sync-publish.py`.
- **Cutting a release?** → `python3 scripts/make-release.py vX.Y.Z`.
- **Updating local install after a rebuild?** → `python3 scripts/deploy-local.py` (or `make deploy-local`).
- **Setting up a fresh dev environment?** → `./scripts/dev-setup.sh` (Linux) or `scripts\dev-setup.ps1` (Windows).
