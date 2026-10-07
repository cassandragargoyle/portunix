# Issue #153: Deliver Docker Documentation Environment for KNIFE Project

## Summary

Complete the Docker-based documentation environment feature so that users of the KNIFE project can use Portunix to spin up OS-agnostic Docker containers for Docusaurus, Hugo, and Docsy documentation stacks. The workflow must be simple: clone a GitHub repo, run Portunix, get a working documentation environment with a shared folder.

**Target platforms**: Windows 11, macOS (Intel + Apple Silicon), Linux (Ubuntu, Debian, CentOS/RHEL)

**Engineering status**: PFT VE-V004 / P06 — ~80% implemented, blocked by 4 items
**GitHub Issue**: #22 (Docusaurus multiplatform support)

## Current State

The foundational infrastructure is largely implemented:

- ✅ Playbook templates exist: `docusaurus.ptxbook.tmpl`, `hugo.ptxbook.tmpl`, `docsify.ptxbook.tmpl` (Docsy = Hugo + Docsy theme, covered by Hugo template)
- ✅ Template metadata (`metadata.json`) defines engines and parameters
- ✅ Package definitions: `docusaurus.json`, `hugo.json`, `hugo-extended.json`
- ✅ Quickstart PowerShell script: `quickstart-docusaurus.ps1` (676 lines)
- ✅ Performance optimization: named volumes for node_modules/npm cache (Issue #128)
- ✅ Hugo integration tests and Docusaurus manual tests

However, an external user **cannot use the feature** because:

1. **No GitHub release published** — quickstart script fetches from `releases/latest` which returns 404
2. **Issue #119 incomplete** — `portunix playbook init --template static-docs` does not work
3. **Templates not embedded** — `.ptxbook.tmpl` files are not included in the compiled binary
4. **RBAC blocks standalone usage** — fails with "User not found" without initial setup

## Desired State

```powershell
# Option A: One-liner quickstart (Windows)
irm https://github.com/CassandraGargoyle/Portunix/releases/latest/download/quickstart-docusaurus.ps1 | iex

# Option B: Manual workflow (any OS)
portunix playbook init --template static-docs --engine docusaurus --target container
portunix playbook run my-docs.ptxbook --script create
portunix playbook run my-docs.ptxbook --script dev
# → Browser opens http://localhost:3000 with live-reload Docusaurus site

# Option C: Hugo variant
portunix playbook init --template static-docs --engine hugo --target container
portunix playbook run my-docs.ptxbook --script dev
# → Browser opens http://localhost:1313
```

Users clone their project repo via GitHub client, run one of the above commands, and get a Docker container with the cloned project directory mounted as a shared folder.

## Requirements

### REQ-1: Publish GitHub Release

- Publish v1.10.4 (or next version) to GitHub Releases with binaries for all platforms
- Include `quickstart-docusaurus.ps1` as a release asset
- Include checksums file
- Ensure `releases/latest` API endpoint resolves correctly

### REQ-2: Complete Template System (Issue #119 dependency)

- Implement `portunix playbook template list` — list available templates
- Implement `portunix playbook template show <name>` — display template details
- Implement `portunix playbook init --template <name> [--engine <engine>] [--target <target>]` — generate `.ptxbook` from template
- Embed `.ptxbook.tmpl` files and `metadata.json` into the binary via `go:embed`

### REQ-3: RBAC Standalone Mode

- Disable RBAC by default for standalone/quickstart usage
- Allow `playbook init` and `playbook run` without prior `portunix setup`
- Preserve opt-in RBAC for multi-user environments

### REQ-4: End-to-End Testing

- Automated test: quickstart script downloads, installs, creates project, starts container
- Test on Windows (PowerShell 5.1+) and Linux (bash equivalent if applicable)
- Test all three engines: Docusaurus, Hugo, Docsify
- Validate shared folder mount works correctly

### REQ-5: User Guide

Create a user-facing guide at `docs/guides/quickstart-documentation-site.md` covering:

- **Prerequisites**: Docker Desktop (or Podman), GitHub client, PowerShell (Windows) or bash (Linux/macOS)
- **Quick start**: One-liner command for each OS
- **Step-by-step walkthrough**: Clone repo → init template → select engine → run → open browser
- **Engine comparison**: Table comparing Docusaurus vs Hugo vs Docsy vs Docsify (use case, language, speed, features)
- **Shared folder workflow**: How local edits reflect in the container, live-reload behavior
- **Troubleshooting**: Common issues (Docker not running, port conflicts, permission errors, slow first run)
- **Reference**: Available scripts (create, dev, build, serve) and their purpose

### REQ-6: Docsy Support (Hugo + Docsy Theme)

Per GitHub Issue #22, Roman's "Docsys" refers to **Docsy** — the Google documentation theme for Hugo (https://www.docsy.dev/). Since Hugo is already implemented:

- Create a Docsy-specific variant of the Hugo template (Hugo + Docsy theme pre-installed)
- Add `docsy` as an engine option in `metadata.json` (or as a Hugo sub-variant)
- Update PFT records (VC-V005, VE-V004) — "Docsys" resolved as Docsy

### REQ-7: Direct Install Command

Per GitHub Issue #22, support a simplified direct command alongside the playbook workflow:

- `portunix install docusaurus` — install Docusaurus with automatic Node.js prerequisites
- `portunix install hugo` — install Hugo binary
- These commands complement the playbook approach for users who want quick local setup

### REQ-7a: Automatic Dependency Installation

The install system declares dependencies in package JSON files (e.g., docusaurus depends on nodejs), and `ResolveDependencies()` exists in registry.go, but it is not called automatically during installation. Dependencies must be resolved and installed automatically before the main package:

- `portunix install docusaurus` should auto-install nodejs first if not present
- Dependency resolution must use existing `ResolveDependencies()` function
- User should be informed about dependency installation progress
- Already-installed dependencies should be skipped

## Acceptance Criteria

- [ ] GitHub release exists with binaries and quickstart script
- [ ] `portunix playbook template list` shows static-docs template
- [ ] `portunix playbook init --template static-docs --engine docusaurus` generates valid `.ptxbook`
- [ ] `portunix playbook init --template static-docs --engine hugo` generates valid `.ptxbook`
- [ ] `portunix playbook init --template static-docs --engine docsify` generates valid `.ptxbook`
- [ ] `portunix playbook run <file> --script create` creates documentation project in container
- [ ] `portunix playbook run <file> --script dev` starts dev server with shared folder mount
- [ ] Quickstart script works end-to-end on clean Windows machine
- [ ] RBAC does not block standalone usage
- [ ] Templates embedded in binary (no external file dependency)
- [ ] User guide published at `docs/guides/quickstart-documentation-site.md`
- [ ] Docsy (Hugo + Docsy theme) available as engine option
- [ ] `portunix install docusaurus` works as direct command
- [ ] Tested on Windows 11, macOS (Intel + Apple Silicon), Linux (Ubuntu/Debian)
- [ ] GitHub Issue #22 can be closed or updated

## Priority

**High** — requested by the reporter of GitHub Issue #22.

## Labels

`feature`, `documentation`, `docker`, `knife-project`

## Related

- Issue #119 — PTX-Ansible Standalone Help & Template Examples (blocking dependency)
- Issue #128 — Docusaurus Container Performance Optimization (✅ done)
- Issue #129 — Docusaurus QuickStart Script (✅ done)
- Issue #075 — Hugo Installation Support (✅ done)
- GitHub Issue #22 — Docusaurus multiplatform support
